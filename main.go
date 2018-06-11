package scache

import (
	"sync"
	"time"
)

// A DataID is a string which uniquely identifies a piece of data.
type DataID string

// NewDataID creates a new DataID, e.g. source = "mydb.mytable.id", id=12345
func NewDataID(source, id string) DataID {
	return DataID(source + "=" + id)
}

// NewCacheItem creates a new cache item.
func NewCacheItem(item interface{}, expiry time.Time) CacheItem {
	return CacheItem{
		Item:   item,
		Expiry: expiry,
	}
}

// CacheItem is the information stored in the cache.
type CacheItem struct {
	// Item is the data item itself.
	Item interface{}
	// Expiry is the time when the item will expire.
	Expiry time.Time
}

// New creates a new Cache.
func New() *Cache {
	return &Cache{
		Now: time.Now,
	}
}

// Cache is a concurrent cache for storing data.
type Cache struct {
	Data sync.Map
	Now  func() time.Time
}

// Put some data into the cache.
func (c *Cache) Put(id DataID, item interface{}, expireAfter time.Duration) {
	c.Data.Store(id, NewCacheItem(item, c.Now().Add(expireAfter)))
}

// Get some data from the cache.
func (c *Cache) Get(id DataID) (item CacheItem, ok bool) {
	d, ok := c.Data.Load(id)
	if !ok {
		return
	}
	item, ok = d.(CacheItem)
	return
}

// Remove an item from the cache.
func (c *Cache) Remove(id DataID) {
	c.Data.Delete(id)
}

// RemoveExpired removes expired values from the cache.
func (c *Cache) RemoveExpired() {
	remover := func(k, v interface{}) bool {
		item := v.(CacheItem)
		if item.Expiry.After(c.Now()) {
			c.Data.Delete(k)
		}
		return true
	}
	c.Data.Range(remover)
}

// RemoveByIDs removes values from the cache by their ID.
func (c *Cache) RemoveByIDs(ids []DataID) {
	for _, id := range ids {
		c.Remove(id)
	}
}
