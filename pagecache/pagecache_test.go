package pagecache_test

import (
	"bytes"
	"io"
	"math/rand"
	"os"
	"testing"
	"testing/iotest"

	"github.com/segmentio/datastructures/v2/pagecache"
)

func TestPageCache(t *testing.T) {
	const size = 20e6 // ~20MB
	r := rand.New(rand.NewSource(3))
	b := new(bytes.Buffer)
	b.Grow(size)

	_, err := io.CopyN(b, r, size)
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.CreateTemp("", "cache.*")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if _, err := f.Write(b.Bytes()); err != nil {
		t.Fatal(err)
	}

	cache := pagecache.New(
		pagecache.PageSize(8192),
		pagecache.PageCount(1024),
	)

	cachedFile := cache.NewFile(1, f, size)

	if err := iotest.TestReader(io.NewSectionReader(cachedFile, 0, size), b.Bytes()); err != nil {
		t.Error(err)
	}
}
