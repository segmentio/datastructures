// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This example demonstrates an integer heap built using the heap interface.
package heap_test

import (
	"fmt"

	"github.com/segmentio/datastructures/v2/container/heap"
)

// Int is a min-heap entry when used in a Heap.
type Int int

func (i Int) Less(v Int) bool { return i < v }

// This example inserts several ints into an IntHeap, checks the minimum,
// and removes them in order of priority.
func Example_IntHeap() {
	h := &heap.Heap[Int]{2, 1, 5}
	heap.Init(h)
	heap.Push(h, 3)
	fmt.Printf("minimum: %d\n", (*h)[0])
	for h.Len() > 0 {
		fmt.Printf("%d ", heap.Pop(h))
	}
	// Output:
	// minimum: 1
	// 1 2 3 5
}
