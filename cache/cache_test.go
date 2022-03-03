package cache

import "testing"

func TestCache(t *testing.T) {
	testCache(t, func() Interface[int, int] { return new(Cache[int, int]) })
}

func TestLRU(t *testing.T) {
	testCache(t, func() Interface[int, int] { return new(LRU[int, int]) })
}

func testCache(t *testing.T, newCache func() Interface[int, int]) {
	tests := []struct {
		scenario string
		function func(*testing.T, Interface[int, int])
	}{
		{
			scenario: "a newly created cache contains no entries",
			function: testCacheNewHasNoEntries,
		},

		{
			scenario: "entries inserted in the cache can be found when looking up their keys",
			function: testCacheInsertAndLookup,
		},

		{
			scenario: "entries deleted from the cache are not returned anymore when looking up keys",
			function: testCacheInsertAndDeleteAndLookup,
		},

		{
			scenario: "deleting entries that did not exist is a no-op",
			function: testCacheDeleteNotExist,
		},

		{
			scenario: "cache evictions returns entries that were previously inserted",
			function: testCacheInsertAndEvict,
		},

		{
			scenario: "inserting entries for existing keys replaces the previous values",
			function: testCacheInsertAndReplace,
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			test.function(t, newCache())
		})
	}
}

func testCacheNewHasNoEntries(t *testing.T, cache Interface[int, int]) {
	if n := cache.Len(); n != 0 {
		t.Errorf("wrong number of cache entries: got=%d want=0", n)
	}
}

func testCacheInsertAndLookup(t *testing.T, cache Interface[int, int]) {
	cache.Insert(1, 10)
	cache.Insert(2, 11)
	cache.Insert(3, 12)

	if n := cache.Len(); n != 3 {
		t.Errorf("wrong number of cache entries: got=%d want=3", n)
	}

	assertCacheLookup(t, cache, 1, 10, true)
	assertCacheLookup(t, cache, 2, 11, true)
	assertCacheLookup(t, cache, 3, 12, true)
}

func testCacheInsertAndDeleteAndLookup(t *testing.T, cache Interface[int, int]) {
	cache.Insert(1, 10)
	cache.Insert(2, 11)
	cache.Insert(3, 12)

	if v, deleted := cache.Delete(3); !deleted {
		t.Error("deleting key=3 was not found in the cache")
	} else if v != 12 {
		t.Errorf("deleting key=3 returned the wrong value: got=%v want=12", v)
	}

	assertCacheLookup(t, cache, 1, 10, true)
	assertCacheLookup(t, cache, 2, 11, true)
	assertCacheLookup(t, cache, 3, 0, false)
}

func testCacheDeleteNotExist(t *testing.T, cache Interface[int, int]) {
	if v, deleted := cache.Delete(42); deleted {
		t.Error("cache successfully deleted non-existing key")
	} else if v != 0 {
		t.Errorf("deletion of non-existing key returned non-zero value: %v", v)
	}
}

func testCacheInsertAndEvict(t *testing.T, cache Interface[int, int]) {
	cache.Insert(1, 10)
	cache.Insert(2, 11)
	cache.Insert(3, 12)

	if k, v, evicted := cache.Evict(); !evicted {
		t.Error("non-empty cache failed to evict anything")
	} else {
		switch k {
		case 1:
			if v != 10 {
				t.Errorf("wrong value returned for key=1: got=%v want=10", v)
			}
		case 2:
			if v != 11 {
				t.Errorf("wrong value returned for key=2: got=%v want=11", v)
			}
		case 3:
			if v != 12 {
				t.Errorf("wrong value returned for key=3: got=%v want=12", v)
			}
		}
	}
}

func testCacheInsertAndReplace(t *testing.T, cache Interface[int, int]) {
	cache.Insert(1, 10)

	if v, replaced := cache.Insert(1, 11); !replaced {
		t.Error("inserting existing key did not replace the previous entry")
	} else if v != 10 {
		t.Errorf("wrong replaced value returned: got=%v want=10", v)
	}

	assertCacheLookup(t, cache, 1, 11, true)
}

func assertCacheLookup(t *testing.T, cache Interface[int, int], key, value int, ok bool) {
	t.Helper()
	v, found := cache.Lookup(key)
	if found != ok {
		t.Errorf("wrong result to cache lookup: got=%t want=%t", found, ok)
	}
	if value != v {
		t.Errorf("wrong value returned by cache lookup: got=%v want=%v", value, v)
	}
	keyFoundInRange, valueFoundInRange := false, false
	cache.Range(func(k, v int) bool {
		if k == key {
			keyFoundInRange = true
			valueFoundInRange = v == value
			return false
		}
		return true
	})
	if keyFoundInRange != ok {
		t.Errorf("the key was not found when ranging over cache entries: %v", key)
	}
	if valueFoundInRange != ok {
		t.Errorf("the value was not found when ranging over cache entries: %v", value)
	}
}
