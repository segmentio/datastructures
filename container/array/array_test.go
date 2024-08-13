package array

import (
	"testing"
	"unsafe"
)

const testLen = 100000

func TestCapacity(t *testing.T) {
	var a Array[int]
	for i := 0; i < testLen; i++ {
		a.Push(i)
	}

	if a.Len() != testLen {
		t.Errorf("incorrect length: got=%d, expect=%d", a.Len(), testLen)
	}

	if expectCap := BlockCap[int](testLen); a.Cap() != expectCap {
		t.Errorf("incorrect capacity: got=%d, expect=%d", a.Cap(), expectCap)
	}

	for i := 0; i < testLen; i++ {
		if a.Get(i) != i {
			t.Errorf("incorrect value at %d: got=%d, expect=%d", i, a.Get(i), i)
		}
	}

	for i := 0; i < testLen; i++ {
		expect := testLen - i - 1
		if v := a.Pop(); v != expect {
			t.Errorf("incorrect popped value: got=%d, expect=%d", v, expect)
		}
	}

	if a.Len() != 0 {
		t.Errorf("incorrect length: got=%d, expect=0", a.Len())
	}
	if a.Cap() != 0 {
		t.Errorf("incorrect capacity: got=%d, expect=%d", a.Cap(), 0)
	}

	a.Push(testLen + 1)

	if a.Len() != 1 {
		t.Errorf("incorrect length: got=%d, expect=1", a.Len())
	}
	if expect := BlockLen[int](); a.Cap() != expect {
		t.Errorf("incorrect capacity: got=%d, expect=%d", a.Cap(), expect)
	}
}

func TestPointerStability(t *testing.T) {
	var a Array[int]
	var ptrs [100]*int

	for i := range ptrs {
		a.Push(i)
		if a.Len() != i+1 {
			t.Errorf("incorrect length: got=%d, expect=%d", a.Len(), i+1)
		}
		if a.Get(i) != i {
			t.Errorf("incorrect value at %d: got=%d, expect=%d", i, a.Get(i), i)
		}
		ptrs[i] = a.Ptr(i)
	}

	for i := len(ptrs); i < 100000; i++ {
		a.Push(i)
		if a.Len() != i+1 {
			t.Errorf("incorrect length: got=%d, expect=%d", a.Len(), i+1)
		}
		if a.Get(i) != i {
			t.Errorf("incorrect value at %d: got=%d, expect=%d", i, a.Get(i), i)
		}
	}

	for i, ptr := range ptrs {
		if ptr != a.Ptr(i) {
			t.Errorf("incorrect ptr at %d: got=%p, expect=%p", i, a.Ptr(i), ptr)
		}
		if *ptr != i {
			t.Errorf("incorrect value at for pointer %d: got=%d, expect=%d", i, *ptr, i)
		}
	}

	for a.Len() > len(ptrs) {
		a.Pop()
	}

	for i, ptr := range ptrs {
		if ptr != a.Ptr(i) {
			t.Errorf("incorrect ptr at %d: got=%p, expect=%p", i, a.Ptr(i), ptr)
		}
		if *ptr != i {
			t.Errorf("incorrect value at for pointer %d: got=%d, expect=%d", i, *ptr, i)
		}
	}
}

func TestPushPtr(t *testing.T) {
	var a Array[int]
	var val int

	for i := 0; i < testLen; i++ {
		val = i * 3
		a.PushPtr(&val)
	}

	for i := 0; i < testLen; i++ {
		expect := i * 3
		actual := a.Get(i)

		if expect != actual {
			t.Errorf("incorrect value at %d: got=%d, expect=%d", i, actual, expect)
		}
	}
}

func TestPool(t *testing.T) {
	var pool Pool[int]
	var a Array[int]

	for i := 0; i < testLen; i++ {
		ptr, _ := a.NextWithPool(&pool)
		*ptr = i
	}

	addrs := map[uintptr]struct{}{}
	for i := range a.blocks {
		addr := uintptr(unsafe.Pointer(&a.blocks[i].slots[0]))
		addrs[addr] = struct{}{}
	}

	for i := 0; i < testLen/2; i++ {
		a.PopWithPool(&pool)
	}

	for i := 0; i < testLen/2; i++ {
		ptr, _ := a.NextWithPool(&pool)
		if *ptr != 0 {
			t.Fatalf("incorrect zero value at %d: got=%d, expect=0", i, *ptr)
		}
		*ptr = i
	}

	// We can't test if all pointers are reused, because sometimes the pool does
	// not keep the value.
	for i := range a.blocks {
		addr := uintptr(unsafe.Pointer(&a.blocks[i].slots[0]))
		if _, ok := addrs[addr]; ok {
			return
		}
	}

	t.Error("no pointers were reused")
}

func TestEmpty(t *testing.T) {
	var a Array[int]

	if !a.Empty() {
		t.Error("should be empty")
	}

	for i := 0; i < testLen; i++ {
		a.Push(i)
		if a.Empty() {
			t.Error("should not be empty")
		}
	}

	for i := 0; i < testLen; i++ {
		a.Pop()
	}

	if !a.Empty() {
		t.Error("should be empty")
	}
}

func TestPanicEmpty(t *testing.T) {
	defer func() {
		expect := "runtime error: index out of range [2] with length 0"
		actual := recover()
		if expect != actual {
			t.Errorf("incorrect panic: got=%v, expect=%v", actual, expect)
		}
	}()

	var a Array[int]
	invalid := a.Get(2)
	t.Errorf("should not be reachable: got=%d", invalid)
}

func TestPanicOver(t *testing.T) {
	defer func() {
		expect := "runtime error: index out of range [2] with length 2"
		actual := recover()
		if expect != actual {
			t.Errorf("incorrect panic: got=%v, expect=%v", actual, expect)
		}
	}()

	var a Array[int]
	a.Push(123)
	a.Push(456)

	invalid := a.Get(2)
	t.Errorf("should not be reachable: got=%d", invalid)
}
