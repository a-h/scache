package scache

import (
	"testing"
	"time"
)

func TestCachePutAndGet(t *testing.T) {
	c := New()
	input := "the thing to cache"
	c.Put("key_1", input)
	output, ok := c.Get("key_1")
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
	c.PutCacheItem("key_1", NewCacheItem("item", expiry, time.Millisecond*10))
	_, ok := c.Get("key_1")
	if !ok {
		t.Fatal("haven't removed expired items yet")
	}
	c.RemoveExpired()
	_, ok = c.Get("key_1")
	if ok {
		t.Fatal("expired items should have been removed")
	}
}

func TestCacheItemsCanBeOverwritten(t *testing.T) {
	c := New()
	c.Put("key_1", "v1")
	c.Put("key_1", "v2")
	ci, ok := c.Get("key_1")
	if !ok {
		t.Error("expected to be able to get the value back")
	}
	if ci.(string) != "v2" {
		t.Errorf("expected 'v2', got '%v'", ci)
	}
}

func TestCacheItemsCanBeRemoved(t *testing.T) {
	c := New()
	c.Put("key_1", "item")
	c.Put("key_2", "item")
	c.Put("key_3", "item")
	c.RemoveByIDs("key_1", "key_2")
	if _, ok := c.Get("key_1"); ok {
		t.Errorf("expected %v to have been removed, but it wasn't", "key_1")
	}
	if _, ok := c.Get("key_2"); ok {
		t.Errorf("expected %v to have been removed, but it wasn't", "key_2")
	}
	if _, ok := c.Get("key_3"); !ok {
		t.Errorf("expected %v to still be accessible, but it wasn't", "key_3")
	}
	c.Remove("key_3")
	if _, ok := c.Get("key_3"); ok {
		t.Errorf("expected %v to have been removed, but it wasn't", "key_3")
	}
}
