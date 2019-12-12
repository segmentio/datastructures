package list

import (
	"math/rand"
	"sync"
	"testing"
	"time"
)

type Int struct {
	Node
	Value int
}

func TestPushFront(t *testing.T) {
	list := new(List)

	for i := 0; i < 10; i++ {
		list.PushFront(&Int{Value: i})
	}

	assertList(t, list, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0)
}

func TestPushBack(t *testing.T) {
	list := new(List)

	for i := 0; i < 10; i++ {
		list.PushBack(&Int{Value: i})
	}

	assertList(t, list, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9)
}

func TestMoveToFront(t *testing.T) {
	list := new(List)
	elem := (*Int)(nil)

	for i := 0; i < 10; i++ {
		e := &Int{Value: i}
		list.PushBack(e)
		if i == 4 {
			elem = e
		}
	}

	list.MoveToFront(list.Front()) // no-op
	assertList(t, list, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9)

	list.MoveToFront(elem)
	assertList(t, list, 4, 0, 1, 2, 3, 5, 6, 7, 8, 9)

	list.MoveToFront(list.Back())
	assertList(t, list, 9, 4, 0, 1, 2, 3, 5, 6, 7, 8)
}

func TestMoveToBack(t *testing.T) {
	list := new(List)
	elem := (*Int)(nil)

	for i := 0; i < 10; i++ {
		e := &Int{Value: i}
		list.PushBack(e)
		if i == 4 {
			elem = e
		}
	}

	list.MoveToBack(list.Front())
	assertList(t, list, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0)

	list.MoveToBack(elem)
	assertList(t, list, 1, 2, 3, 5, 6, 7, 8, 9, 0, 4)

	list.MoveToBack(list.Back()) // no-op
	assertList(t, list, 1, 2, 3, 5, 6, 7, 8, 9, 0, 4)
}

func TestRemoveFront(t *testing.T) {
	list := new(List)
	values := [10]int{}

	for i := range values {
		values[i] = i
		list.PushBack(&Int{Value: i})
	}

	for i, v := range values {
		assertInt(t, list.RemoveFront(), v)
		assertList(t, list, values[i+1:]...)
	}

	assertList(t, list)
}

func TestRemoveBack(t *testing.T) {
	list := new(List)
	values := [10]int{}

	for i := range values {
		values[i] = i
		list.PushBack(&Int{Value: i})
	}

	for i := range values {
		j := len(values) - (i + 1)
		assertInt(t, list.RemoveBack(), values[j])
		assertList(t, list, values[:j]...)
	}

	assertList(t, list)
}

func TestRemove(t *testing.T) {
	list := new(List)
	elem := (*Int)(nil)

	for i := 0; i < 10; i++ {
		e := &Int{Value: i}
		list.PushBack(e)
		if i == 4 {
			elem = e
		}
	}

	list.Remove(list.Front())
	assertList(t, list, 1, 2, 3, 4, 5, 6, 7, 8, 9)

	list.Remove(elem)
	assertList(t, list, 1, 2, 3, 5, 6, 7, 8, 9)

	list.Remove(list.Back())
	assertList(t, list, 1, 2, 3, 5, 6, 7, 8)
}

func assertInt(t *testing.T, found interface{}, expected int) {
	t.Helper()

	if i := found.(*Int); i.Value != expected {
		t.Errorf("value mismatch, expected %d but found %d", expected, i.Value)
	}
}

func assertList(t *testing.T, l *List, v ...int) {
	t.Helper()

	if len(v) == 0 {
		if front := l.Front(); front != nil {
			t.Errorf("front of list mismatch, expected <nil> but found %+v", front)
		}
		if back := l.Back(); back != nil {
			t.Errorf("back of list mismatch, expected <nil> but found %+v", back)
		}
	} else {
		if front, _ := l.Front().(*Int); front == nil {
			t.Errorf("front of list mismatch, expected %d but found <nil>", v[0])
		} else if front.Value != v[0] {
			t.Errorf("front of list mismatch, expected %d but found %d", v[0], front.Value)
		}

		if back, _ := l.Back().(*Int); back == nil {
			t.Errorf("back of list mismatch, expected %d but found <nil>", v[len(v)-1])
		} else if back.Value != v[len(v)-1] {
			t.Errorf("back of list mismatch, expected %d but found %d", v[len(v)-1], back.Value)
		}
	}

	for i, x := 0, l.Front(); x != nil; i, x = i+1, l.Next(x) {
		if i >= len(v) {
			t.Errorf("[forward] list contains too many elements, expected %d but found %d", len(v), i+1)
			break
		}
		if x.(*Int).Value != v[i] {
			t.Errorf("[forward] list element at index %d mismatch, expected %d but found %d", i, v[i], x.(*Int).Value)
			break
		}
	}

	for i, x := len(v)-1, l.Back(); x != nil; i, x = i-1, l.Prev(x) {
		if i < 0 {
			t.Errorf("[backward] list contains too many elements, expected %d but found %d", len(v), len(v)-(i+1))
			break
		}
		if x.(*Int).Value != v[i] {
			t.Errorf("[backward] list element at index %d mismatch, expected %d but found %d", i, v[i], x.(*Int).Value)
			break
		}
	}

	if n := l.Len(); n != len(v) {
		t.Errorf("list length mismatch, expected %d but found %d", len(v), n)
	}
}

func BenchmarkMove(b *testing.B) {
	values := make([]Int, 1000)
	for i := range values {
		values[i].Value = i
	}

	list := new(List)
	for i := range values {
		list.PushBack(&values[i])
	}

	mutex := sync.Mutex{}
	b.RunParallel(func(pb *testing.PB) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		n := len(values)

		for pb.Next() {
			i := r.Intn(n)

			mutex.Lock()
			if (i % 2) == 0 {
				list.MoveToFront(&values[i])
			} else {
				list.MoveToBack(&values[i])
			}
			mutex.Unlock()
		}
	})

	seen := make(map[int]int)
	for x := list.Front(); x != nil; x = list.Next(x) {
		seen[x.(*Int).Value]++
	}

	for value, count := range seen {
		if count > 1 {
			b.Errorf("%d occurrences of %d found in the list", count, value)
			break
		}
	}

	if len(seen) != len(values) {
		b.Errorf("expected %d values but found %d", len(values), len(seen))
	}
}
