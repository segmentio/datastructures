package cache

import "github.com/segmentio/datastructures/v2/container/list"

// LRU is an Interface implementation which caches elements and tracks least
// recently used items as candidates for eviction.
type LRU[K comparable, V any] struct {
	index map[K]*list.Element[entry[K, V]]
	queue list.List[entry[K, V]]
}

type entry[K comparable, V any] struct {
	key   K
	value V
}

func (lru *LRU[K, V]) Len() int {
	return lru.queue.Len()
}

func (lru *LRU[K, V]) Insert(key K, value V) (previous V, replaced bool) {
	if lru.index == nil {
		lru.index = make(map[K]*list.Element[entry[K, V]])
	}
	e, ok := lru.index[key]
	if ok {
		previous, replaced = e.Value.value, true
		lru.queue.Remove(e)
	}
	lru.index[key] = lru.queue.PushFront(entry[K, V]{key: key, value: value})
	return previous, replaced
}

func (lru *LRU[K, V]) Lookup(key K) (value V, found bool) {
	e, ok := lru.index[key]
	if ok {
		lru.queue.MoveToFront(e)
		value, found = e.Value.value, true
	}
	return value, found
}

func (lru *LRU[K, V]) Delete(key K) (value V, deleted bool) {
	e, ok := lru.index[key]
	if ok {
		delete(lru.index, key)
		lru.queue.Remove(e)
		value, deleted = e.Value.value, true
	}
	return value, deleted
}

func (lru *LRU[K, V]) Evict() (key K, value V, evicted bool) {
	if lru.queue.Len() > 0 {
		e := lru.queue.Back()
		lru.queue.Remove(e)
		delete(lru.index, e.Value.key)
		key, value, evicted = e.Value.key, e.Value.value, true
	}
	return key, value, evicted
}

func (lru *LRU[K, V]) Range(f func(K, V) bool) {
	for _, e := range lru.index {
		if !f(e.Value.key, e.Value.value) {
			break
		}
	}
}
