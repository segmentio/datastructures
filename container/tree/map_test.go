package tree

import (
	"fmt"
	"sort"
	"testing"
	"testing/quick"

	"github.com/segmentio/datastructures/v2/compare"
)

func TestMap(t *testing.T) {
	tests := []struct {
		scenario string
		function func(*testing.T, *Map[int32, int64])
	}{
		{
			scenario: "an empty map has a length of zero",
			function: testMapEmpty,
		},

		{
			scenario: "entries inserted in the tree are found when looking up their keys",
			function: testMapInsertAndLookup,
		},

		{
			scenario: "inserting the same keys multiple times replaces the previous values",
			function: testMapInsertAndReplace,
		},

		{
			scenario: "entries deleted from the tree are not found when looking up their keys",
			function: testMapInsertAndDelete,
		},

		{
			scenario: "deleting entries that do not exist does not modify the map",
			function: testMapDeleteNotExist,
		},

		{
			scenario: "ranging over entries produces map keys ordered by the comparison function",
			function: testMapRange,
		},

		{
			scenario: "searching for an existing key returns the assiocated value",
			function: testMapSearchExist,
		},

		{
			scenario: "searching for a non-existing key returns the value associated to the highest key that is lower or equal",
			function: testMapSearchNotExist,
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			m := NewMap[int32, int64](compare.Function[int32])
			test.function(t, m)
			m.checkInvariants()
		})
	}
}

func testMapEmpty(t *testing.T, m *Map[int32, int64]) {
	if n := m.Len(); n != 0 {
		t.Errorf("wrong number of map entries: got=%d want=0", n)
	}
}

func testMapInsertAndLookup(t *testing.T, m *Map[int32, int64]) {
	f := func(keys map[int32]int64) bool {
		m.Init(compare.Function[int32])

		for k, v := range keys {
			previous, replaced := m.Insert(k, v)
			if replaced {
				t.Errorf("replaced key=%d with value=%d which did not exist in the map", k, previous)
				return false
			}
		}

		if n := m.Len(); n != len(keys) {
			t.Errorf("wrong number of entries in map: got=%d want=%d", n, len(keys))
			return false
		}

		for k, v := range keys {
			value, found := m.Lookup(k)
			if !found {
				t.Errorf("key not found in map: %d", k)
				return false
			} else if value != v {
				t.Errorf("wrong value returned for key=%d: got=%d want=%d", k, value, v)
				return false
			}
		}

		return true
	}
	quick.Check(f, nil)
}

func testMapInsertAndReplace(t *testing.T, m *Map[int32, int64]) {
	f := func(keys map[int32]int64) bool {
		m.Init(compare.Function[int32])

		for k, v := range keys {
			previous, replaced := m.Insert(k, v)
			if replaced {
				t.Errorf("replaced key=%d with value=%d which did not exist in the map", k, previous)
				return false
			}
		}

		for k, v := range keys {
			previous, replaced := m.Insert(k, v+1)
			if !replaced {
				t.Errorf("value was not replaced for key=%d", k)
				return false
			}
			if previous != v {
				t.Errorf("wrong previous value returned when replacing key=%d: got=%d want=%d", k, previous, v)
				return false
			}
		}

		if n := m.Len(); n != len(keys) {
			t.Errorf("wrong number of entries in map: got=%d want=%d", n, len(keys))
			return false
		}

		for k, v := range keys {
			value, found := m.Lookup(k)
			if !found {
				t.Errorf("key not found in map: %d", k)
				return false
			} else if value != (v + 1) {
				t.Errorf("wrong value returned for key=%d: got=%d want=%d", k, value, v+1)
				return false
			}
		}

		return true
	}
	quick.Check(f, nil)
}

func testMapInsertAndDelete(t *testing.T, m *Map[int32, int64]) {
	f := func(keys map[int32]int64) bool {
		m.Init(compare.Function[int32])

		for k, v := range keys {
			previous, replaced := m.Insert(k, v)
			if replaced {
				t.Errorf("replaced key=%d with value=%d which did not exist in the map", k, previous)
				return false
			}
		}

		numKeys := len(keys)
		for k, v := range keys {
			if (v % 2) == 0 {
				numKeys--
				value, deleted := m.Delete(k)
				if !deleted {
					t.Errorf("value not deleted for key=%d value=%d", k, v)
					return false
				}
				if value != v {
					t.Errorf("wrong value deleted for key=%d: got=%d want=%d", k, value, v)
					return false
				}
			}
		}

		if n := m.Len(); n != numKeys {
			t.Errorf("wrong number of entries in map: got=%d want=%d", n, numKeys)
			return false
		}

		for k, v := range keys {
			value, found := m.Lookup(k)
			expected := v%2 != 0
			if found != expected {
				t.Errorf("key not found in map: %d", k)
				return false
			} else if expected && value != v {
				t.Errorf("wrong value returned for key=%d: got=%d want=%d", k, value, v)
				return false
			}
		}

		// Re-insert all the deleted keys and expect the find all afterwards.
		for k, v := range keys {
			if (v % 2) == 0 {
				m.Insert(k, v)
			}
		}

		for k, v := range keys {
			value, found := m.Lookup(k)
			if !found {
				t.Errorf("key not found in map: %d", k)
				return false
			} else if value != v {
				t.Errorf("wrong value returned for key=%d: got=%d want=%d", k, value, v)
				return false
			}
		}

		return true
	}
	quick.Check(f, nil)
}

func testMapDeleteNotExist(t *testing.T, m *Map[int32, int64]) {
	f := func(keys map[int32]int64) bool {
		m.Init(compare.Function[int32])

		deleteKeys := map[int32]struct{}{
			0: {},
			1: {},
			2: {},
			3: {},
		}

		numKeys := 0
		for k, v := range keys {
			if _, skip := deleteKeys[k]; !skip {
				numKeys++
				previous, replaced := m.Insert(k, v)
				if replaced {
					t.Errorf("replaced key=%d with value=%d which did not exist in the map", k, previous)
					return false
				}
			}
		}

		for k := range deleteKeys {
			v, deleted := m.Delete(k)
			if deleted {
				t.Errorf("successfully deleted entry which did not exist in the map: key=%d value=%d", k, v)
				return false
			}
		}

		if n := m.Len(); n != numKeys {
			t.Errorf("wrong number of entries in map: got=%d want=%d", n, numKeys)
			return false
		}

		for k, v := range keys {
			value, found := m.Lookup(k)
			if !found {
				t.Errorf("key not found in map: %d", k)
				return false
			} else if value != v {
				t.Errorf("wrong value returned for key=%d: got=%d want=%d", k, value, v)
				return false
			}
		}

		return true
	}
	quick.Check(f, nil)
}

func testMapRange(t *testing.T, m *Map[int32, int64]) {
	f := func(keys map[int32]int64) bool {
		m.Init(compare.Function[int32])

		for k, v := range keys {
			previous, replaced := m.Insert(k, v)
			if replaced {
				t.Errorf("replaced key=%d with value=%d which did not exist in the map", k, previous)
				return false
			}
		}

		if n := m.Len(); n != len(keys) {
			t.Errorf("wrong number of entries in map: got=%d want=%d", n, len(keys))
			return false
		}

		type entry struct {
			k int32
			v int64
		}

		entries := make([]entry, 0, len(keys))
		for k, v := range keys {
			entries = append(entries, entry{k: k, v: v})
		}
		sort.Slice(entries, func(i, j int) bool { return entries[i].k < entries[j].k })

		i := 0
		m.Range(func(k int32, v int64) bool {
			if k != entries[i].k {
				t.Errorf("wrong key for entry at index %d: got=%d want=%d", i, k, entries[i].k)
				return false
			}
			if v != entries[i].v {
				t.Errorf("wrong value for entry at index %d: got=%d want=%d", i, v, entries[i].v)
				return false
			}
			i++
			return true
		})

		if i < len(keys) {
			t.Errorf("ranging over keys did not expose all entries: got=%d want=%d", i, len(keys))
		} else if i > len(keys) {
			t.Errorf("ranging over keys exposed too many entries: got=%d want=%d", i, len(keys))
		}
		return true
	}
	quick.Check(f, nil)
}

func testMapSearchExist(t *testing.T, m *Map[int32, int64]) {
	f := func(keys map[int32]int64) bool {
		m.Init(compare.Function[int32])

		for k, v := range keys {
			previous, replaced := m.Insert(k, v)
			if replaced {
				t.Errorf("replaced key=%d with value=%d which did not exist in the map", k, previous)
				return false
			}
		}

		if n := m.Len(); n != len(keys) {
			t.Errorf("wrong number of entries in map: got=%d want=%d", n, len(keys))
			return false
		}

		for k, v := range keys {
			key, value, found := m.Search(k)
			if !found {
				t.Errorf("key not found in map: %d", k)
				return false
			} else if key != k {
				t.Errorf("wrong key returned: got=%d want=%d", key, k)
				return false
			} else if value != v {
				t.Errorf("wrong value returned for key=%d: got=%d want=%d", k, value, v)
				return false
			}
		}

		return true
	}
	quick.Check(f, nil)
}

func testMapSearchNotExist(t *testing.T, m *Map[int32, int64]) {
	f := func(keys map[int32]int64) bool {
		m.Init(compare.Function[int32])

		limit := len(keys) / 2
		exist := make(map[int32]int64, limit)
		dontExist := make(map[int32]int64, limit)

		for k, v := range keys {
			if len(exist) < limit {
				exist[k] = v
			} else {
				dontExist[k] = v
			}
		}

		for k, v := range exist {
			previous, replaced := m.Insert(k, v)
			if replaced {
				t.Errorf("replaced key=%d with value=%d which did not exist in the map", k, previous)
				return false
			}
		}

		if n := m.Len(); n != len(exist) {
			t.Errorf("wrong number of entries in map: got=%d want=%d", n, len(exist))
			return false
		}

		search := func(k int32) (int32, int64, bool) {
			if len(exist) == 0 {
				return 0, 0, false
			}
			key, value, found := int32(0), int64(0), false
			for existKey, existValue := range exist {
				if existKey <= k && (!found || existKey > key) {
					key, value, found = existKey, existValue, true
				}
			}
			return key, value, found
		}

		for k := range dontExist {
			key, value, found := m.Search(k)
			expectKey, expectValue, expectFound := search(k)
			if found != expectFound {
				t.Errorf("key search mismatch: key=%d got=%t want=%t", k, found, expectFound)
				return false
			} else if key != expectKey {
				t.Errorf("wrong key returned: got=%d want=%d", key, expectKey)
				return false
			} else if value != expectValue {
				t.Errorf("wrong value returned for key=%d: got=%d want=%d", k, value, expectValue)
				return false
			}
		}

		return true
	}
	quick.Check(f, nil)
}

func (m *Map[K, V]) checkInvariants() {
	if m.root.color != B {
		panic("root must be black")
	}
	ys := make([]int, 0)
	xs := &ys
	m.check(m.root, 0, xs)
	i := 1
	for i < len(*xs) {
		if (*xs)[i-1] != (*xs)[i] {
			fmt.Println(xs)
			panic("black height not same for all the leaves")
		}
		i++
	}
}

func (m *Map[K, V]) check(n *node[K, V], bh int, xs *[]int) {
	if n == &m.leaf {
		*xs = append(*xs, bh)
		return
	}
	if n.color == R {
		if !colors(n, n.a, n.b, R, B, B) {
			m.preorder(m.root, "")
			fmt.Println(n, n.a, n.b)
			panic("red node without both children black")
		}
	}
	if n.color == B {
		bh += 1
	}
	m.check(n.a, bh, xs)
	m.check(n.b, bh, xs)
}

func (m *Map[K, V]) preorder(n *node[K, V], tab string) {
	if n != &m.leaf {
		fmt.Println(tab, n.key, "=>", n.value, n.color)
		m.preorder(n.a, ":"+tab)
		m.preorder(n.b, ":"+tab)
	}
}

func BenchmarkInsert(b *testing.B) {
	const N = 1024
	m := NewMap[int, int](compare.Function[int])

	for i := 0; i < b.N; i++ {
		m.Insert(i%N, i)
	}
}

func BenchmarkLookup(b *testing.B) {
	const N = 1024
	m := NewMap[int, int](compare.Function[int])

	for i := 0; i < N; i++ {
		m.Insert(i, i)
	}

	for i := 0; i < b.N; i++ {
		m.Lookup(i % N)
	}
}
