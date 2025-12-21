package cache

import (
	"sync"
	"time"
)

// Cache interface defines the behavior for caching mechanisms.
type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration)
	Delete(key string)
	Flush()
}

// item represents a cached value with an expiration.
type item struct {
	value      interface{}
	expiration int64
}

func (i item) isExpired() bool {
	if i.expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > i.expiration
}

// MemoryCache is an in-memory implementation of the Cache interface.
type MemoryCache struct {
	items sync.Map
}

// NewMemoryCache creates a new in-memory cache.
func NewMemoryCache() *MemoryCache {
	c := &MemoryCache{}
	go c.cleanupLoop()
	return c
}

// Get retrieves a value from the cache.
func (c *MemoryCache) Get(key string) (interface{}, bool) {
	val, ok := c.items.Load(key)
	if !ok {
		return nil, false
	}
	it := val.(item)
	if it.isExpired() {
		c.items.Delete(key)
		return nil, false
	}
	return it.value, true
}

// Set stores a value in the cache with a TTL.
func (c *MemoryCache) Set(key string, value interface{}, ttl time.Duration) {
	var exp int64
	if ttl > 0 {
		exp = time.Now().Add(ttl).UnixNano()
	}

	c.items.Store(key, item{
		value:      value,
		expiration: exp,
	})
}

// Delete removes a value from the cache.
func (c *MemoryCache) Delete(key string) {
	c.items.Delete(key)
}

// Flush clears the cache.
func (c *MemoryCache) Flush() {
	c.items = sync.Map{}
}

func (c *MemoryCache) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		c.items.Range(func(key, value interface{}) bool {
			it := value.(item)
			if it.isExpired() {
				c.items.Delete(key)
			}
			return true
		})
	}
}
