package array

import (
	"sync"
	"unsafe"
)

// Pool is a memory pool for reusing block allocations. The pool can be
// shared across any BucketArray of the same type.
type Pool[T any] struct{ p sync.Pool }

// get returns a slice out of the pool. If the pool is nil or empty, a new
// slice is allocated.
func (p *Pool[T]) get() []T {
	if p != nil {
		if b, ok := p.p.Get().(*T); ok {
			return unsafe.Slice(b, BlockLen[T]())
		}
	}
	return make([]T, BlockLen[T]())
}

// put returns a slice back into the pool.
func (p *Pool[T]) put(s []T) {
	p.p.Put(unsafe.SliceData(s))
}
