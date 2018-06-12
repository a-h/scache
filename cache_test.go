package scache

import (
	"testing"
	"time"

	"github.com/a-h/scache/data"
)

func TestCachePutAndGet(t *testing.T) {
	c := New()
	input := "the thing to cache"
	c.Put(data.NewID("mydb.table.tableid", "12345"), input)
	output, ok := c.Get(data.NewID("mydb.table.tableid", "12345"))
	if !ok {
		t.Fatal("could not get data we just put in")
	}
	if output != input {
		t.Errorf("expected input and output to be equal, got '%v'", output)
	}
}

func TestCacheItemsCanExpire(t *testing.T) {
	c := New()
	expiry := time.Now().Add(time.Second * -1)
	c.PutCacheItem(data.NewID("mydb.table.tableid", "12345"), NewCacheItem("item", expiry))
	_, ok := c.Get(data.NewID("mydb.table.tableid", "12345"))
	if !ok {
		t.Fatal("haven't removed expired items yet")
	}
	c.RemoveExpired()
	_, ok = c.Get(data.NewID("mydb.table.tableid", "12345"))
	if ok {
		t.Fatal("expired items should have been removed")
	}
}

func TestCacheItemsCanBeOverwritten(t *testing.T) {
	id := data.NewID("mydb.table.tableid", "1")
	c := New()
	c.Put(id, "v1")
	c.Put(id, "v2")
	ci, ok := c.Get(id)
	if !ok {
		t.Error("expected to be able to get the value back")
	}
	if ci.(string) != "v2" {
		t.Errorf("expected 'v2', got '%v'", ci)
	}
}

func TestCacheItemsCanBeRemoved(t *testing.T) {
	c := New()
	id1, id2, id3 := data.NewID("mydb.table.tableid", "1"),
		data.NewID("mydb.table.tableid", "2"),
		data.NewID("mydb.table.tableid", "3")
	c.Put(id1, "item")
	c.Put(id2, "item")
	c.Put(id3, "item")
	c.RemoveByIDs(id1, id2)
	if _, ok := c.Get(id1); ok {
		t.Errorf("expected %v to have been removed, but it wasn't", id1)
	}
	if _, ok := c.Get(id2); ok {
		t.Errorf("expected %v to have been removed, but it wasn't", id2)
	}
	if _, ok := c.Get(id3); !ok {
		t.Errorf("expected %v to still be accessible, but it wasn't", id3)
	}
	c.Remove(id3)
	if _, ok := c.Get(id3); ok {
		t.Errorf("expected %v to have been removed, but it wasn't", id3)
	}
}
