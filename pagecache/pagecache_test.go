package pagecache_test

import (
	"bytes"
	"io"
	"math/rand"
	"sync"
	"testing"
	"testing/iotest"

	"github.com/segmentio/datastructures/v2/pagecache"
)

func TestPageCache(t *testing.T) {
	const size = 2e6 // ~2MB
	r := rand.New(rand.NewSource(3))
	b := new(bytes.Buffer)
	b.Grow(size)

	_, err := io.CopyN(b, r, size)
	if err != nil {
		t.Fatal(err)
	}

	cache := pagecache.New(
		pagecache.PageSize(512),
		pagecache.PageCount(1024),
	)

	wg := sync.WaitGroup{}
	data := b.Bytes()

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cachedFile := cache.NewFile(1, bytes.NewReader(data), size)

			if err := iotest.TestReader(io.NewSectionReader(cachedFile, 0, size), data); err != nil {
				t.Error(err)
			}
		}()
	}

	wg.Wait()
}
