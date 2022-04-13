// Package pagecache provides the implementation of a caching layer for files
// implementing the io.ReaderAt interface.
package pagecache

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/maphash"
	"io"
	"math/bits"
	"sync"

	"github.com/segmentio/datastructures/v2/cache"
)

const (
	// DefaultPageSize is the default page size used when creating a Cache
	// instance.
	DefaultPageSize = 4096

	// DefaultPageCount is the default page count used when creating a Cache
	// instance.
	DefaultPageCount = 16384
)

const (
	// The number of buckets in a Cache instance.
	//
	// At this time, numBuckets is a constant value, though it may be
	// interesting to make it configurable in the future. Having this value be
	// a constant and a power of two allows the compiler to optimize modulo
	// operations using bit masks, which are instructions that tend to be orders
	// of magnitude faster. If we make this value configurable, we might want to
	// ensure that we retain the same performance characteristics, which would
	// require us to only allow powers of two as bucket counts, and implement
	// the bitwise optimizations in the code.
	numBuckets = 64
)

var (
	// ErrNoPages is returned when memory pressure is too high and it is not
	// possible to read files through the page cache.
	ErrNoPages = errors.New("there are no free pages left in the cache")
)

// Config carries the configuration for the page cache.
type Config struct {
	PageSize  int64
	PageCount int64
}

// DefaultConfig constructs a new Config instance initialized with the default
// configuration.
func DefaultConfig() *Config {
	return &Config{
		PageSize:  DefaultPageSize,
		PageCount: DefaultPageCount,
	}
}

// Apply applies the list of options passed as arguments to c.
func (c *Config) Apply(options ...Option) {
	for _, opt := range options {
		opt.Configure(c)
	}
}

// Option is an interface implemented by options allowing configuration of new
// Cache instances.
type Option interface {
	Configure(*Config)
}

type option func(*Config)

func (opt option) Configure(config *Config) { opt(config) }

// PageSize is a cache configuration option setting the size of individual pages
// in a Cache instance.
//
// If it is not a power of two, the size will be adjusted to the nearest one.
//
// Default: 4 KiB
func PageSize(size int64) Option {
	return option(func(config *Config) { config.PageSize = size })
}

// PageCount is a configuration option setting the number of pages in a Cache
// instance.
//
// Default: 16384
func PageCount(count int64) Option {
	return option(func(config *Config) { config.PageCount = count })
}

// Cache instances implement the page caching layer of files.
type Cache struct {
	hashseed maphash.Seed
	shift    uint
	// The cache is divided into buckets, each bucket holding a section of the
	// total page count. Each bucket can synchronize cache access and evict
	// outdated pages independently. Having multiple buckets helps scale cache
	// access when running in multi-threaded programs where a single cache mutex
	// could quickly become a bottleneck in the cache.
	buckets [numBuckets]bucket
}

// New constructs a new Cache instance, using the list of options passed as
// arguments to configure the cache.
func New(options ...Option) *Cache {
	config := DefaultConfig()
	config.Apply(options...)
	return NewWithConfig(config)
}

// NewWithConfig is like New but uses a Config instance to pass the cache
// configuration instead of a list of options.
func NewWithConfig(config *Config) *Cache {
	pageSize := config.PageSize
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}

	pageCount := config.PageCount
	if pageCount <= 0 {
		pageCount = 1
	}
	if (pageCount % numBuckets) != 0 {
		pageCount = ((pageCount / numBuckets) + 1) * numBuckets
	}

	shift := uint(bits.Len64(uint64(pageSize - 1)))
	pageSize = int64(1) << shift
	cacheSize := pageSize * pageCount
	bucketSize := cacheSize / numBuckets
	// TODO: should we make the allocator configurable?
	data := make([]byte, cacheSize)

	c := &Cache{
		hashseed: maphash.MakeSeed(),
		shift:    shift,
	}

	for i := range c.buckets {
		b := &c.buckets[i]
		b.init(data[int64(i)*bucketSize:int64(i+1)*bucketSize], pageSize)
	}

	return c
}

// NewFile constructs a wrapper around the file of the given size. id is a
// unique identifier intended to uniquely represent the file within the cache.
// If multiple io.ReaderAt interfaces point at the same underlying file, they
// could share the same id to reference the same pages in the cache.
func (c *Cache) NewFile(id uint32, file io.ReaderAt, size int64) io.ReaderAt {
	return &cachedFile{
		cache: c,
		id:    id,
		file:  file,
		size:  size,
	}
}

func (c *Cache) bucketOf(key region) *bucket {
	b := [8]byte{}
	binary.LittleEndian.PutUint32(b[:4], key.object)
	binary.LittleEndian.PutUint32(b[4:], key.offset)
	// This hashing strategy ensures that we will not see hotspots from cache
	// access, pages are spread evenly across buckets, independently of their
	// position or files that they belong to. Those properties must be retained
	// if the hashing algorithm is changed.
	h := maphash.Hash{}
	h.SetSeed(c.hashseed)
	h.Write(b[:])
	return &c.buckets[h.Sum64()%numBuckets]
}

// Stats is a structure carrying statistics collected on cache access.
//
// All counters are absolute values accumulated since a cache instance was
// created.
type Stats struct {
	Lookups   int64 // reads from the cache
	Hits      int64 // page reads that were found in the cache
	Inserts   int64 // pages inserted in the cache
	Evictions int64 // pages evicted from the cache
	Allocs    int64 // number of free pages allocated by the cache
	Frees     int64 // number of allocated pages returned to the free pool
}

// HitRate returns the hit rate of cache lookups, as a floating point value
// between 0 and 1 (inclusive).
func (s *Stats) HitRate() float64 {
	return float64(s.Hits) / float64(s.Lookups)
}

// Stats returns the current values of cache statistics.
func (c *Cache) Stats() (stats Stats) {
	for i := range c.buckets {
		b := &c.buckets[i]
		s := b.stats()
		stats.Lookups += s.lookups
		stats.Hits += s.hits
		stats.Inserts += s.inserts
		stats.Evictions += s.evictions
		stats.Allocs += s.allocs
		stats.Frees += s.frees
	}
	return stats
}

type cachedFile struct {
	cache *Cache
	id    uint32
	file  io.ReaderAt
	size  int64
}

func (f *cachedFile) ReadAt(b []byte, off int64) (n int, err error) {
	if off < 0 {
		return 0, fmt.Errorf("offset out of range: %d/%d", off, f.size)
	}
	if off >= f.size {
		return 0, io.EOF
	}
	if limit := f.size - off; limit < int64(len(b)) {
		b = b[:limit]
	}
	if len(b) == 0 {
		return 0, nil
	}

	cache := f.cache
	shift := cache.shift
	pageSize := int64(1) << shift

	for {
		key := region{
			object: f.id,
			offset: uint32(off >> shift),
		}

		pageOffset := int64(key.offset) << shift
		readOffset := off - pageOffset

		if bucket := cache.bucketOf(key); !bucket.read(b[n:], key, shift, readOffset) {
			page, data, ok := bucket.get(shift)
			if !ok {
				return n, ErrNoPages
			}

			rn, err := f.file.ReadAt(data, pageOffset)
			if rn < len(data) && !errors.Is(err, io.EOF) {
				if err == nil {
					err = io.ErrNoProgress
				}
				return n, err
			}

			copy(b[n:], data[readOffset:rn])
			bucket.put(key, page, shift)
		}

		readBytes := pageSize - readOffset
		if n += int(readBytes); n >= len(b) {
			return len(b), nil
		}
		if off += readBytes; off >= f.size {
			return n, io.EOF
		}
	}
}

type region struct {
	object uint32
	offset uint32
}

type page struct {
	offset uint32
}

type bucket struct {
	mutex sync.Mutex
	cache cache.LRU[region, page]
	freed []page
	pages []byte
	bucketStats
}

type bucketStats struct {
	lookups   int64
	hits      int64
	inserts   int64
	evictions int64
	allocs    int64
	frees     int64
}

func (b *bucket) init(data []byte, pageSize int64) {
	b.pages = data
	b.freed = make([]page, int64(len(data))/pageSize)
	for i := range b.freed {
		b.freed[i].offset = uint32(i)
	}
}

func (b *bucket) bytes(page page, shift uint) []byte {
	offset := int64(page.offset) << shift
	length := int64(1) << shift
	return b.pages[offset : offset+length]
}

func (b *bucket) read(data []byte, key region, shift uint, off int64) bool {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	page, ok := b.cache.Lookup(key)
	if ok {
		b.hits++
		copy(data, b.bytes(page, shift)[off:])
	}
	b.lookups++
	return ok
}

func (b *bucket) get(shift uint) (page, []byte, bool) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if i := len(b.freed) - 1; i >= 0 {
		page := b.freed[i]
		b.freed = b.freed[:i]
		b.allocs++
		return page, b.bytes(page, shift), true
	}

	_, page, evicted := b.cache.Evict()
	if evicted {
		b.evictions++
		return page, b.bytes(page, shift), true
	}

	return page, nil, false
}

func (b *bucket) put(key region, page page, shift uint) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	page, replaced := b.cache.Insert(key, page)
	if replaced {
		b.freed = append(b.freed, page)
		b.frees++
	}

	b.inserts++
}

func (b *bucket) stats() (stats bucketStats) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.bucketStats
}
