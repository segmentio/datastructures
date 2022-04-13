package pagecache_test

import (
	"bytes"
	"io"
	"math/rand"
	"sync"
	"testing"
	"testing/iotest"
	"time"

	"github.com/segmentio/datastructures/v2/pagecache"
)

func TestPageCache(t *testing.T) {
	const size = 2e6 // ~2MB
	r := rand.New(rand.NewSource(3))
	b := new(bytes.Buffer)
	b.Grow(size)

	_, err := io.CopyN(b, r, size)
	if err != nil {
		t.Fatal(err)
	}

	cache := pagecache.New(
		pagecache.PageSize(512),
		pagecache.PageCount(1024),
	)

	wg := sync.WaitGroup{}
	data := b.Bytes()

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cachedFile := cache.NewFile(1, bytes.NewReader(data), size)

			if err := iotest.TestReader(io.NewSectionReader(cachedFile, 0, size), data); err != nil {
				t.Error(err)
			}
		}()
	}

	wg.Wait()
}

func BenchmarkPageCacheNoEvictions(b *testing.B) {
	const size = 2e6 // ~2MB
	prng := rand.New(rand.NewSource(3))
	data := new(bytes.Buffer)
	data.Grow(size)

	_, err := io.CopyN(data, prng, size)
	if err != nil {
		b.Fatal(err)
	}

	// 4 MiB cache, no evictions
	cache := pagecache.New(
		pagecache.PageSize(4096),
		pagecache.PageCount(1024),
	)

	file := cache.NewFile(1, bytes.NewReader(data.Bytes()), size)

	b.RunParallel(func(pb *testing.PB) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		b := make([]byte, 1000)
		n := len(b) / 2

		for pb.Next() {
			offset := r.Int63n(size)
			length := r.Intn(n) + n
			file.ReadAt(b[:length], offset)
		}
	})

	stats := cache.Stats()
	b.Logf("hit rate: %.2f%%", 100*stats.HitRate())
}

func BenchmarkPageCacheWithEvictions(b *testing.B) {
	const size = 2e6 // ~2MB
	prng := rand.New(rand.NewSource(3))
	data := new(bytes.Buffer)
	data.Grow(size)

	_, err := io.CopyN(data, prng, size)
	if err != nil {
		b.Fatal(err)
	}

	// <2 MiB cache, some evictions will occur
	cache := pagecache.New(
		pagecache.PageSize(4096),
		pagecache.PageCount(100),
	)

	file := cache.NewFile(1, bytes.NewReader(data.Bytes()), size)

	b.RunParallel(func(pb *testing.PB) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		b := make([]byte, 1000)
		n := len(b) / 2

		for pb.Next() {
			offset := r.Int63n(size)
			length := r.Intn(n) + n
			file.ReadAt(b[:length], offset)
		}
	})

	stats := cache.Stats()
	b.Logf("hit rate: %.2f%%", 100*stats.HitRate())
}
