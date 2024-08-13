package array

import "unsafe"

// BlockSize is the memory allocation size in bytes for each block slot. This
// is converted into a slice capacity using the BlockLen function.
const BlockSize = 64 * 1024

// BlockLen calculates the entry count required to fill a slice of BlockSize
// bytes. Because the type is generic, we can't use const. However, as of go
// 1.20, the compiler will turn this into a constant expression. For example,
// BlockLen[int]() turns into 8192. On x86_64, this can be seen in Next:
//
//	CALL    runtime.makeslice(SB)
//	MOVL    $8192, CX
//	MOVL    $8192, DI
//
// ... and fast integer division/modulus in Ptr:
//
//	MOVQ    BX, SI
//	SARQ    $13, BX
//	ANDQ    $-8192, SI
func BlockLen[T any]() int {
	var v T
	return BlockSize / int(unsafe.Sizeof(v))
}

// BlockCap rounds n up to the nearest BlockLen.
func BlockCap[T any](n int) int {
	return (n + BlockLen[T]() - 1) / BlockLen[T]() * BlockLen[T]()
}

type block[T any] struct {
	slots []T
	count int
}

// reset clears the block and returns the slot slice for possible reuse.
func (b *block[T]) reset() []T {
	slots := b.slots
	b.slots = nil
	b.count = 0
	return slots
}

// next increments the count and returns a pointer to the new location. This
// does NOT validate that count is availble in the underlying slice.
func (b *block[T]) next() (p *T, index int) {
	index = b.count
	p = &b.slots[index]
	b.count = index + 1
	return
}

func (b *block[T]) at(i int) *T     { return &b.slots[i] }
func (b *block[T]) len() int        { return b.count }
func (b *block[T]) available() bool { return b.count < cap(b.slots) }

// pop zeroes out the last entries and decrements the count. We zero the entry
// so that it allows any pointers to be nil'ed and availble for GC. This also
// ensures that Next always returns zeroed memory.
func (b *block[T]) pop() T {
	i := b.count - 1
	b.count = i

	var v T
	v, b.slots[i] = b.slots[i], v
	return v
}
