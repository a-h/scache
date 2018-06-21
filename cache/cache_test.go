package cache

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
	c.RemoveMany("key_1", "key_2")
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

func TestCacheItemsCanBeCounted(t *testing.T) {
	c := New()
	c.Put("key_1", "item")
	c.Put("key_2", "item")
	c.Put("key_3", "item")
	if c.Count() != 3 {
		t.Errorf("expected to have %d items, but got %d", 3, c.Count())
	}
	c.Remove("key_3")
	c.Remove("key_3")
	if c.Count() != 2 {
		t.Errorf("expected to have %d items, but got %d", 2, c.Count())
	}
}

func TestCacheItemsCanBeRetrievedWithTiming(t *testing.T) {
	c := New()
	c.PutCacheItem("key_1", Item{
		Expiry: time.Date(2000, time.January, 1, 12, 0, 0, 0, time.UTC),
		Saved:  time.Second * 30,
		Value:  "123",
	})
	v, ts, ok := c.GetWithDuration("key_1")
	if !ok {
		t.Fatal("could not get data we just put in")
	}
	if v.(string) != "123" {
		t.Errorf("expected '123' got '%v'", v)
	}
	if ts != time.Second*30 {
		t.Errorf("expected 30 seconds, got %v", ts)
	}
}
