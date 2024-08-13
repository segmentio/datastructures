package queue

import (
	"testing"

	"github.com/segmentio/datastructures/v2/container/array"
)

var pool array.Pool[int]

func TestQueue(t *testing.T) {
	var q Queue[int]
	count := (array.BlockLen[int]() * 5) - 10
	t.Logf("using %d entries with block size %d", count, array.BlockLen[int]())

	for i := 0; i < count; i++ {
		if q.Len() != i {
			t.Fatalf("incorrect count: expect=%d, got=%d", i, q.Len())
		}
		q.PushWithPool(i+1, &pool)
	}

	if q.Len() != count {
		t.Fatalf("incorrect count: expect=%d, got=%d", count, q.Len())
	}

	c := q.values.Cap()

	v, ok := q.PopWithPool(&pool)
	if !ok {
		t.Fatal("expected pop to remove value")
	}
	if v != 1 {
		t.Fatalf("incorrected popped value: expect=%d, got=%d", 1, v)
	}
	if q.Len() != count-1 {
		t.Errorf("incorrect count: expect=%d, got=%d", count-1, q.Len())
	}
	if q.values.Cap() != c {
		t.Errorf("capacity should not have changed: expect=%d, got=%d", c, q.values.Cap())
	}

	at := 2
	for q.Len() > 10 {
		v, ok = q.PopWithPool(&pool)
		if !ok {
			t.Error("expected pop to remove value")
		}
		if v != at {
			t.Fatalf("incorrect popped value: expect=%d, got=%d", at, v)
		}
		at += 1
	}

	if q.values.Cap() != array.BlockLen[int]() {
		t.Errorf("capactiy should be one block: expect=%d, got=%d", array.BlockLen[int](), q.values.Cap())
	}

	for q.Len() > 0 {
		before := q.Len()
		v, ok = q.PopWithPool(&pool)
		t.Logf("len = %d -> %d", before, q.Len())
		if !ok {
			t.Error("expected pop to remove value")
		}
		if v != at {
			t.Fatalf("incorrected popped value: expect=%d, got=%d", at, v)
		}
		at += 1
	}

	if q.Len() != 0 {
		t.Fatalf("incorrect count: expect=%d, got=%d", 0, q.Len())
	}

	if q.values.Cap() != array.BlockLen[int]() {
		t.Errorf("capactiy should be one block: expect=%d, got=%d", array.BlockLen[int](), q.values.Cap())
	}
}
