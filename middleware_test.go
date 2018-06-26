package scache

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/a-h/scache/cache"
	"github.com/a-h/scache/data"
)

type valueInCache struct {
	A string
	B string
}

func TestGetAndPut(t *testing.T) {
	// Arrange.
	r := httptest.NewRequest("GET", "/", nil)

	c := cache.New()
	vic := valueInCache{
		A: "A",
		B: "B",
	}
	dataKey := data.NewID("db.table.id", "12345")
	// Add value to cache.

	//c.Put(dataKey.String(), vic)
	ccc := cacheContextContent{
		Cache: c,
	}
	ctx := context.WithValue(r.Context(), cacheContextKey, &ccc)
	r = r.WithContext(ctx)

	// Act: put the vaue into the cache.
	added := Add(r, dataKey, vic)
	if !added {
		t.Fatal("Failed to add item to cache.")
	}

	// Act: get the value from the cache.
	var vfc valueInCache
	ok := Get(r, dataKey, &vfc)

	// Assert.
	if !ok {
		t.Fatal("Expected to be able to get the value from the cache, but didn't.")
	}
	if vfc.A != vic.A {
		t.Errorf("Expected the value from the cache to equal the value we put in the cache, but got A == '%v'", vic.A)
	}
	if vfc.B != vic.B {
		t.Errorf("Expected the value from the cache to equal the value we put in the cache, but got B == '%v'", vic.B)
	}
}
