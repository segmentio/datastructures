package queue

import "github.com/segmentio/datastructures/v2/container/array"

// Queue implements a FIFO queue backed by an array.Array for efficient
// block-based allocations.
type Queue[T any] struct {
	values array.Array[T]
	index  int
}

// Len gets the number of values in the queue.
func (q *Queue[T]) Len() int {
	return q.values.Len() - q.index
}

// Empty reports whether if the queue contains no entries.
func (q *Queue[T]) Empty() bool {
	return q.values.Len() == q.index
}

// Get obtains the value of an entry at a specific offset. Like a Go slice
// this results in a panic if the value is outside the array bounds.
func (q *Queue[T]) Get(i int) T {
	return q.values.Get(i + q.index)
}

// Ptr obtains the pointer to a value for an entry at a specific offset. The
// returned pointer is guaranteed not to change until the index is popped from
// the array. Like a Go slice this results in a panic if the value is outside
// the array bounds.
func (q *Queue[T]) Ptr(i int) *T {
	return q.values.Ptr(i + q.index)
}

// Push adds a new value to the end of the queue.
func (q *Queue[T]) Push(v T) {
	q.PushWithPool(v, nil)
}

// PushWithPool adds a new value to the end of the queue. If a new block is
// required by the backing array.Array, it will be allocated from the pool.
func (q *Queue[T]) PushWithPool(v T, pool *array.Pool[T]) {
	q.values.PushWithPool(v, pool)
}

// Pop removes the oldest value from the queue in FIFO order if any are available.
func (q *Queue[T]) Pop() (v T, ok bool) {
	return q.PopWithPool(nil)
}

// PopWithPool removes the oldest value from the queue in FIFO order if any are
// available. If the removed value was the only item remaining in the block of
// the backing array.Array, that block will be returned to the pool.
func (q *Queue[T]) PopWithPool(pool *array.Pool[T]) (v T, ok bool) {
	if i, n := q.index, q.values.Len(); i < n {
		ptr := q.values.Ptr(i)
		v, ok, i = *ptr, true, i+1

		var zero T
		*ptr = zero // zero out as we go to avoid possible downstream memory references

		if i == array.BlockLen[T]() {
			q.values.ShiftBlocksWithPool(1, pool)
			i = 0
		}

		q.index = i
	}
	return
}
