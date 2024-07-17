package array

// Array works similarly to a Go slice, however the allocation is performed
// in blocks rather than a single sequential allocation. This means that appends
// do not need to fully copy prior entries during reallocation, and entries have
// pointer stability throughout their lifetime.
type Array[T any] struct {
	blocks []block[T]
}

// Push appends a value into the array.
func (a *Array[T]) Push(v T) {
	p, _ := a.Next()
	*p = v
}

// PushPtr appends a value into the array by copying a value from an existing
// memory address.
func (a *Array[T]) PushPtr(v *T) {
	p, _ := a.Next()
	*p = *v
}

// Next obtains a pointer to the next slot, allocating new space if needed.
// The returned pointer will be zero-initialized.
func (a *Array[T]) Next() (ptr *T, index int) {
	return a.NextWithPool(nil)
}

// NextWithPool obtains a pointer to the next slot, allocating new space from the
// optionally provided pool if needed. The returned pointer will be zero-initialized.
func (a *Array[T]) NextWithPool(pool *Pool[T]) (ptr *T, index int) {
	i := len(a.blocks) - 1
	if i >= 0 && a.blocks[i].available() {
		ptr, index = a.blocks[i].next()
	} else {
		a.blocks = append(a.blocks, block[T]{slots: pool.get(), count: 1})
		i++
		ptr = &a.blocks[i].slots[index]
	}
	index += i * BlockLen[T]()
	return
}

// Pop removes the last item from the array.
func (a *Array[T]) Pop() {
	a.PopWithPool(nil)
}

func (a *Array[T]) PopWithPool(pool *Pool[T]) {
	i := len(a.blocks) - 1
	a.blocks[i].pop()

	if a.blocks[i].len() == 0 {
		free := a.blocks[i].reset()
		if pool != nil {
			pool.put(free)
		}
		a.blocks = a.blocks[:i]

		if len(a.blocks) < cap(a.blocks)/4 {
			blocks := make([]block[T], len(a.blocks), cap(a.blocks)/2)
			copy(blocks, a.blocks)
			a.blocks = blocks
		}
	}
}

// Get obtains the value of an entry at a specific offset. Like a Go slice
// this results in a panic if the value is outside the array bounds.
func (a *Array[T]) Get(i int) T {
	return *a.Ptr(i)
}

// Ptr obtains the pointer to a value for an entry at a specific offset. The
// returned pointer is guaranteed not to change until the index is popped from
// the array. Like a Go slice this results in a panic if the value is outside
// the array bounds.
func (a *Array[T]) Ptr(i int) *T {
	size := BlockLen[T]()
	b, o := i/size, i%size
	return a.blocks[b].at(o)
}

// Len returns the number of entries added to the array.
func (a *Array[T]) Len() int {
	n := len(a.blocks)
	if n > 0 {
		n = (BlockLen[T]() * (n - 1)) + a.blocks[n-1].len()
	}
	return n
}

// Cap returns the current capacity of the array without furthur reallocation.
func (a *Array[T]) Cap() int {
	return len(a.blocks) * BlockLen[T]()
}

// Reset removes all entries and releases all blocks into the pool if provided.
func (a *Array[T]) Reset(pool *Pool[T]) {
	if pool != nil {
		for _, b := range a.blocks {
			pool.put(b.slots)
		}
	}
	*a = Array[T]{}
}
