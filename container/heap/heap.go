// Package heap contains the implementation of a generic heap derived from the
// standard container/heap package. Additionally, this package provides some
// convenince implementations that operate using go slices, simplifying the
// interface requirements.
package heap

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import "sort"

// The Interface type describes the requirements
// for a type using the routines in this package.
// Any type that implements it may be used as a
// min-heap with the following invariants (established after
// [Init] has been called or if the data is empty or sorted):
//
//	!h.Less(j, i) for 0 <= i < h.Len() and 2*i+1 <= j <= 2*i+2 and j < h.Len()
//
// Note that [Push] and [Pop] in this interface are for package heap's
// implementation to call. To add and remove things from the heap,
// use [heap.Push] and [heap.Pop].
type Interface[T any] interface {
	sort.Interface
	Push(T)
	Pop() T
}

// Init establishes the heap invariants required by the other routines in this package.
// Init is idempotent with respect to the heap invariants
// and may be called whenever the heap invariants may have been invalidated.
// The complexity is O(n) where n = h.Len().
func Init[T any, H Interface[T]](h H) {
	n := h.Len()
	for i := n/2 - 1; i >= 0; i-- {
		down(h, i, n)
	}
}

// Push pushes the element x onto the heap.
// The complexity is O(log n) where n = h.Len().
func Push[T any, H Interface[T]](h H, value T) {
	h.Push(value)
	up(h, h.Len()-1)
}

// Pop removes and returns the minimum element (according to Less) from the heap.
// The complexity is O(log n) where n = h.Len().
// Pop is equivalent to [Remove](h, 0).
func Pop[T any, H Interface[T]](h H) T {
	n := h.Len() - 1
	h.Swap(0, n)
	down(h, 0, n)
	return h.Pop()
}

// Remove removes and returns the element at index i from the heap.
// The complexity is O(log n) where n = h.Len().
func Remove[T any, H Interface[T]](h H, i int) T {
	n := h.Len() - 1
	if n != i {
		h.Swap(i, n)
		if !down(h, i, n) {
			up(h, i)
		}
	}
	return h.Pop()
}

// Fix re-establishes the heap ordering after the element at index i has changed its value.
// Changing the value of the element at index i and then calling Fix is equivalent to,
// but less expensive than, calling [Remove](h, i) followed by a Push of the new value.
// The complexity is O(log n) where n = h.Len().
func Fix[T any, H Interface[T]](h H, i int) {
	if !down(h, i, h.Len()) {
		up(h, i)
	}
}

func up[T any, H Interface[T]](h H, child int) {
	for {
		parent := (child - 1) / 2
		if parent == child || !h.Less(child, parent) {
			break
		}
		h.Swap(parent, child)
		child = parent
	}
}

func down[T any, H Interface[T]](h H, parent, n int) bool {
	start := parent
	for {
		left := 2*parent + 1
		if left >= n || left < 0 {
			break
		}
		child := left
		if right := left + 1; right < n && h.Less(right, left) {
			child = right
		}
		if !h.Less(child, parent) {
			break
		}
		h.Swap(parent, child)
		parent = child
	}
	return parent > start
}

// Heap provides a convenience for implementing the [heap.Interface] as a slice
// with entries of type T.
type Heap[T Entry[T]] []T

// Entry defines the interface requirement for each entry in a [heap.Heap].
type Entry[T any] interface {
	// Less reports whether the receiver is less than the passed-in value.
	//
	// This must hold the same transitivity of the Less method in [sort.Interface].
	Less(T) bool
}

// Len returns the number of items in the heap. This implements the Len method
// of [heap.Interface].
func (h Heap[T]) Len() int { return len(h) }

// Less implements [heap.Interface].
func (h Heap[T]) Less(i, j int) bool { return h[i].Less(h[j]) }

// Swap implements [heap.Interface]. This method should not be called directly.
func (h Heap[T]) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

// Push implements [heap.Interface]. This method should not be called directly,
// use [heap.Push] instead.
func (h *Heap[T]) Push(x T) { *h = append(*h, x) }

// Pop implements [heap.Interface]. This method should not be called directly,w
// use [heap.Pop] instead.
func (h *Heap[T]) Pop() T {
	old := *h
	n := len(old) - 1
	value := old[n]
	*h = old[:n]
	return value
}

func pop[T any](h *[]T) T {
	old := *h
	n := len(old) - 1
	value := old[n]
	*h = old[:n]
	return value
}

const NoIndex = -1

// IndexHeap provides a convenience for implementing the [heap.Interface] as a
// slice with entries of type T that also track the index position within each
// entry. By storing the index, entries can easily be updated with [heap.Fix]
// and removed using [heap.Remove].
type IndexHeap[T IndexEntry[T]] []T

// IndexEntry defines the interface requirement for each entry in a [heap.IndexHeap].
type IndexEntry[T any] interface {
	Entry[T]
	// SetIndex updates the entrie's index position within the heap slice.
	SetIndex(int)
}

// Len returns the number of items in the heap. This implements the Len method
// of [heap.Interface].
func (h IndexHeap[T]) Len() int { return len(h) }

// Less implements [heap.Interface].
func (h IndexHeap[T]) Less(i, j int) bool { return h[i].Less(h[j]) }

// Swap implements [heap.Interface]. This method should not be called directly.
func (h IndexHeap[T]) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].SetIndex(i)
	h[j].SetIndex(j)
}

// Push implements [heap.Interface]. This method should not be called directly,
// use [heap.Push] instead.
func (h *IndexHeap[T]) Push(x T) {
	x.SetIndex(h.Len())
	*h = append(*h, x)
}

// Pop implements [heap.Interface]. This method should not be called directly,w
// use [heap.Pop] instead.
func (h *IndexHeap[T]) Pop() T {
	old := *h
	n := len(old) - 1
	value := old[n]
	*h = old[:n]
	value.SetIndex(NoIndex)
	return value
}
