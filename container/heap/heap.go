package heap

type Index int

const NoIndex Index = -1

func (i Index) IsSet() bool {
	return i >= 0
}

type Heap[T Entry[T]] []T

type Entry[T any] interface {
	Less(T) bool
	Index() Index
	SetIndex(i Index)
}

func (h Heap[T]) Len() int {
	return len(h)
}

func (h *Heap[T]) Push(value T) {
	value.SetIndex(Index(h.Len()))
	*h = append(*h, value)
	h.up(h.Len() - 1)
}

func (h *Heap[T]) Pop() T {
	n := h.Len() - 1
	h.swap(0, n)
	h.down(0, n)
	return h.pop(n)
}

func (h *Heap[T]) Remove(value T) T {
	i := int(value.Index())
	n := h.Len() - 1
	if n != i {
		h.swap(i, n)
		if !h.down(i, n) {
			h.up(i)
		}
	}
	return h.pop(n)
}

func (h *Heap[T]) Fix(i int) {
	if !h.down(i, h.Len()) {
		h.up(i)
	}
}

func (h *Heap[T]) pop(n int) T {
	old := *h
	g := old[n]
	*h = old[:n]
	g.SetIndex(NoIndex)
	return g
}

func (h Heap[T]) swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].SetIndex(Index(i))
	h[j].SetIndex(Index(j))
}

func (h Heap[T]) up(child int) {
	for {
		parent := (child - 1) / 2
		if parent == child || !h[child].Less(h[parent]) {
			break
		}
		h.swap(parent, child)
		child = parent
	}
}

func (h Heap[T]) down(parent, n int) bool {
	start := parent
	for {
		left := 2*parent + 1
		if left >= n || left < 0 {
			break
		}
		child := left
		if right := left + 1; right < n && h[right].Less(h[left]) {
			child = right
		}
		if !h[child].Less(h[parent]) {
			break
		}
		h.swap(parent, child)
		parent = child
	}
	return parent > start
}
