// Package cache contains data structures that are useful to build caches.
//
// The types provided by the package are generic building blocks implementing
// caching algorithms. Synchronization in caching strategies is often very
// specific to the application and harder to generalize, so the types provided
// by this package do not make opinionated choices on how synchronization should
// be handled, which makes them unsafe to use concurrently from multiple
// goroutines
package cache

// Interface is the interface implemented by caches.
type Interface[K comparable, V any] interface {
	// Returns the number of items in the cache.
	Len() int

	// Inserts an item in the cache, returning the previous value associated
	// with the cache key.
	Insert(key K, value V) (previous V, replaced bool)

	// Returns the value associated with the given key in the cache.
	Lookup(key K) (value V, found bool)

	// Deletes an item from the cache.
	Delete(key K) (value V, deleted bool)

	// Evicts an item from the cache.
	Evict() (key K, value V, evicted bool)

	// Calls f for each entry in the cache. The order in which entries are
	// presented is unspecified. If f returns false, iteration stops.
	Range(f func(K, V) bool)
}

// Stats contains counters tracking usage of a cache.
type Stats struct {
	Inserts   int64
	Updates   int64
	Deletes   int64
	Lookups   int64
	Hits      int64
	Evictions int64
}

// Cache wraps an underlying caching implementation, adding measures of usage.
//
// By default, a LRU caching strategy is used.
type Cache[K comparable, V any] struct {
	inserts   int64
	updates   int64
	deletes   int64
	lookups   int64
	hits      int64
	evictions int64
	backend   Interface[K, V]
}

func (c *Cache[K, V]) Init(backend Interface[K, V]) {
	c.inserts = 0
	c.updates = 0
	c.deletes = 0
	c.lookups = 0
	c.hits = 0
	c.evictions = 0
	c.backend = backend
}

func (c *Cache[K, V]) Len() int {
	if c.backend != nil {
		return c.backend.Len()
	}
	return 0
}

func (c *Cache[K, V]) Insert(key K, value V) (previous V, replaced bool) {
	if c.backend == nil {
		c.backend = new(LRU[K, V])
	}
	previous, replaced = c.backend.Insert(key, value)
	if replaced {
		c.updates++
	} else {
		c.inserts++
	}
	return previous, replaced
}

func (c *Cache[K, V]) Lookup(key K) (value V, found bool) {
	if c.backend != nil {
		value, found = c.backend.Lookup(key)
		c.lookups++
		if found {
			c.hits++
		}
	}
	return value, found
}

func (c *Cache[K, V]) Delete(key K) (value V, deleted bool) {
	if c.backend != nil {
		value, deleted = c.backend.Delete(key)
		if deleted {
			c.deletes++
		}
	}
	return value, deleted
}

func (c *Cache[K, V]) Evict() (key K, value V, evicted bool) {
	if c.backend != nil {
		key, value, evicted = c.backend.Evict()
		if evicted {
			c.evictions++
		}
	}
	return key, value, evicted
}

func (c *Cache[K, V]) Range(f func(K, V) bool) {
	if c.backend != nil {
		c.backend.Range(f)
	}
}

func (c *Cache[K, V]) Stats() Stats {
	return Stats{
		Inserts:   c.inserts,
		Updates:   c.updates,
		Deletes:   c.deletes,
		Lookups:   c.lookups,
		Hits:      c.hits,
		Evictions: c.evictions,
	}
}
