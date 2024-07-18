package heap

import (
	"math/rand"
	"testing"
)

type int2 int

func (i int2) Less(j int2) bool { return i < j }

type item struct {
	name  string
	value int
	index int
}

func (i *item) Less(j *item) bool { return i.value < j.value }
func (i *item) SetIndex(v int)    { i.index = v }

func newItem(name string, value int) *item {
	return &item{name: name, value: value, index: NoIndex}
}

func verify(t *testing.T, h Heap[int2], i int) {
	t.Helper()
	n := h.Len()
	j1 := 2*i + 1
	j2 := 2*i + 2
	if j1 < n {
		if h.Less(j1, i) {
			t.Errorf("heap invariant invalidated [%d] = %d > [%d] = %d", i, h[i], j1, h[j1])
			return
		}
		verify(t, h, j1)
	}
	if j2 < n {
		if h.Less(j2, i) {
			t.Errorf("heap invariant invalidated [%d] = %d > [%d] = %d", i, h[i], j1, h[j2])
			return
		}
		verify(t, h, j2)
	}
}

func TestInit0(t *testing.T) {
	h := Heap[int2]{}
	for i := 20; i > 0; i-- {
		h.Push(0) // all elements are the same
	}
	Init(&h)
	verify(t, h, 0)

	for i := 1; h.Len() > 0; i++ {
		x := Pop(&h)
		verify(t, h, 0)
		if x != 0 {
			t.Errorf("%d.th pop got %d; want %d", i, x, 0)
		}
	}
}

func TestInit1(t *testing.T) {
	h := Heap[int2]{}
	for i := 20; i > 0; i-- {
		h.Push(int2(i)) // all elements are different
	}
	Init(&h)
	verify(t, h, 0)

	for i := 1; h.Len() > 0; i++ {
		x := Pop(&h)
		verify(t, h, 0)
		if int(x) != i {
			t.Errorf("%d.th pop got %d; want %d", i, x, i)
		}
	}
}

func Test(t *testing.T) {
	h := Heap[int2]{}
	verify(t, h, 0)

	for i := 20; i > 10; i-- {
		h.Push(int2(i))
	}
	Init(&h)
	verify(t, h, 0)

	for i := 10; i > 0; i-- {
		Push(&h, int2(i))
		verify(t, h, 0)
	}

	for i := 1; h.Len() > 0; i++ {
		x := Pop(&h)
		if i < 20 {
			Push(&h, int2(20+i))
		}
		verify(t, h, 0)
		if int(x) != i {
			t.Errorf("%d.th pop got %d; want %d", i, x, i)
		}
	}
}

func TestRemove0(t *testing.T) {
	h := Heap[int2]{}
	for i := 0; i < 10; i++ {
		h.Push(int2(i))
	}
	verify(t, h, 0)

	for h.Len() > 0 {
		i := h.Len() - 1
		x := Remove(&h, i)
		if int(x) != i {
			t.Errorf("Remove(%d) got %d; want %d", i, x, i)
		}
		verify(t, h, 0)
	}
}

func TestRemove1(t *testing.T) {
	h := Heap[int2]{}
	for i := 0; i < 10; i++ {
		h.Push(int2(i))
	}
	verify(t, h, 0)

	for i := 0; h.Len() > 0; i++ {
		x := Remove(&h, 0)
		if int(x) != i {
			t.Errorf("Remove(0) got %d; want %d", x, i)
		}
		verify(t, h, 0)
	}
}

func TestRemove2(t *testing.T) {
	N := 10

	h := Heap[int2]{}
	for i := 0; i < N; i++ {
		h.Push(int2(i))
	}
	verify(t, h, 0)

	m := make(map[int2]bool)
	for h.Len() > 0 {
		m[Remove(&h, (h.Len()-1)/2)] = true
		verify(t, h, 0)
	}

	if len(m) != N {
		t.Errorf("len(m) = %d; want %d", len(m), N)
	}
	for i := 0; i < len(m); i++ {
		if !m[int2(i)] {
			t.Errorf("m[%d] doesn't exist", i)
		}
	}
}

func TestFix(t *testing.T) {
	h := Heap[int2]{}
	verify(t, h, 0)

	for i := 200; i > 0; i -= 10 {
		Push(&h, int2(i))
	}
	verify(t, h, 0)

	if h[0] != 10 {
		t.Fatalf("Expected head to be 10, was %d", h[0])
	}
	h[0] = 210
	Fix(&h, 0)
	verify(t, h, 0)

	for i := 100; i > 0; i-- {
		elem := rand.Intn(h.Len())
		if i&1 == 0 {
			h[elem] *= 2
		} else {
			h[elem] /= 2
		}
		Fix(&h, elem)
		verify(t, h, 0)
	}
}

func TestIndexHeap(t *testing.T) {
	var h IndexHeap[*item]
	a := newItem("a", 2)
	b := newItem("b", 1)
	c := newItem("c", 3)

	Push(&h, a)
	Push(&h, b)
	Push(&h, c)

	for i, item := range []*item{b, a, c} {
		if h[i].name != item.name {
			t.Errorf("expected index %d to be %q, got %q", i, item.name, h[i].name)
		}

		if h[i].index != i {
			t.Errorf("exected %q to have index %d, got %d", item.index, i, h[i].index)
		}
	}
}
