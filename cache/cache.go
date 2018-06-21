package cache

import (
	"math/rand"
	"sync"
	"time"
)

// NewCacheItem creates a new cache item.
func NewCacheItem(item interface{}, expiry time.Time, saved time.Duration) CacheItem {
	return CacheItem{
		Item:   item,
		Expiry: expiry,
		Saved:  saved,
	}
}

// CacheItem is the information stored in the cache.
type CacheItem struct {
	// Item is the data item itself.
	Item interface{}
	// Expiry is the time when the item will expire.
	Expiry time.Time
	// Saved is the amount of time saved by getting this item from the cache.
	Saved time.Duration
}

// DefaultExpiration is the default expiration function, it caches for at least one hour,
// randomising expiry within a further 30 minute range.
var DefaultExpiration = func(now func() time.Time) time.Time {
	return now().Add(time.Hour * 1).Add(time.Duration(rand.Intn(30)) * time.Minute)
}

// New creates a new Cache.
func New() *Cache {
	return &Cache{
		Now:        time.Now,
		Expiration: DefaultExpiration,
	}
}

// Cache is a concurrent cache for storing data.
type Cache struct {
	Data       sync.Map
	Now        func() time.Time
	Expiration func(now func() time.Time) time.Time
}

// Put some data into the cache.
func (c *Cache) Put(key string, item interface{}) {
	c.PutWithDuration(key, item, time.Duration(0))
}

// PutWithDuration puts some data into the cache, including the duration.
func (c *Cache) PutWithDuration(key string, item interface{}, saved time.Duration) {
	expiryTime := c.Expiration(c.Now)
	c.PutCacheItem(key, NewCacheItem(item, expiryTime, saved))
}

// PutCacheItem puts a cache item into memory.
func (c *Cache) PutCacheItem(key string, item CacheItem) {
	c.Data.Store(key, item)
}

// Get some data from the cache.
func (c *Cache) Get(key string) (value interface{}, ok bool) {
	var ci CacheItem
	ci, ok = c.GetItem(key)
	if ok {
		value = ci.Item
	}
	return
}

// GetWithDuration gets data from the cache, including how much time was saved by getting it from the cache.
func (c *Cache) GetWithDuration(key string) (value interface{}, saved time.Duration, ok bool) {
	var ci CacheItem
	ci, ok = c.GetItem(key)
	if ok {
		value = ci.Item
		saved = ci.Saved
	}
	return
}

// GetItem gets the item from the cache.
func (c *Cache) GetItem(key string) (item CacheItem, ok bool) {
	d, ok := c.Data.Load(key)
	if !ok {
		return
	}
	item, ok = d.(CacheItem)
	return
}

// Remove an item from the cache.
func (c *Cache) Remove(key string) {
	c.Data.Delete(key)
}

// RemoveExpired removes expired values from the cache.
func (c *Cache) RemoveExpired() {
	remover := func(k, v interface{}) bool {
		item := v.(CacheItem)
		if item.Expiry.Before(c.Now()) {
			c.Data.Delete(k)
		}
		return true
	}
	c.Data.Range(remover)
}

// RemoveMany removes values from the cache by their ID.
func (c *Cache) RemoveMany(keys ...string) {
	for _, key := range keys {
		c.Remove(key)
	}
}

// Count the number of items in the cache.
func (c *Cache) Count() (count int) {
	counter := func(k, v interface{}) bool {
		count++
		return true
	}
	c.Data.Range(counter)
	return
}
