package scache

import (
	"math/rand"
	"sync"
	"time"
)

// A DataID is a string which uniquely identifies a piece of data.
type DataID struct {
	Source string
	ID     string
}

// NewDataID creates a new DataID, e.g. source = "mydb.mytable.id", id=12345
func NewDataID(source, id string) DataID {
	return DataID{
		Source: source,
		ID:     id,
	}
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
func (c *Cache) Put(id DataID, item interface{}) {
	expiryTime := c.Expiration(c.Now)
	c.PutCacheItem(id, NewCacheItem(item, expiryTime))
}

// PutCacheItem puts a cache item into memory.
func (c *Cache) PutCacheItem(id DataID, item CacheItem) {
	c.Data.Store(id, item)
}

// Get some data from the cache.
func (c *Cache) Get(id DataID) (value interface{}, ok bool) {
	var ci CacheItem
	ci, ok = c.GetItem(id)
	if ok {
		value = ci.Item
	}
	return
}

// GetItem gets the item from the cache.
func (c *Cache) GetItem(id DataID) (item CacheItem, ok bool) {
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
		if item.Expiry.Before(c.Now()) {
			c.Data.Delete(k)
		}
		return true
	}
	c.Data.Range(remover)
}

// RemoveByIDs removes values from the cache by their ID.
func (c *Cache) RemoveByIDs(ids ...DataID) {
	for _, id := range ids {
		c.Remove(id)
	}
}
