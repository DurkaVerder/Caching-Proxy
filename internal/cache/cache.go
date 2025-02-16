package cache

import (
	"sync"
	"time"
)

type Item struct {
	Value      any
	Created    time.Time
	Expiration int64
}

type Cache struct {
	sync.RWMutex
	defaultExpiration time.Duration
	cleanupInterval   time.Duration
	items             map[string]Item
}

func NewCache(defaultExpiration, cleanupInterval time.Duration) *Cache {
	cache := &Cache{
		defaultExpiration: defaultExpiration,
		cleanupInterval:   cleanupInterval,
		items:             make(map[string]Item),
	}
	if cleanupInterval > 0 {
		cache.runCleaner()
	}

	return cache
}

func (c *Cache) Set(key string, value any, duration time.Duration) {
	var expiration int64

	if duration == 0 {
		duration = c.defaultExpiration
	}

	if duration > 0 {
		expiration = time.Now().Add(duration).UnixNano()
	}

	c.Lock()
	c.items[key] = Item{
		Value:      value,
		Created:    time.Now(),
		Expiration: expiration,
	}
	c.Unlock()
}

func (c *Cache) Get(key string) (any, bool) {
	c.RLock()
	item, found := c.items[key]
	c.RUnlock()

	if !found {
		return nil, false
	}

	if item.Expiration > 0 {
		if time.Now().UnixNano() > item.Expiration {
			return nil, false
		}
	}

	return item.Value, true
}

func (c *Cache) Delete(key string) {
	c.Lock()
	delete(c.items, key)
	c.Unlock()
}

func (c *Cache) runCleaner() {
	go c.cleaner()
}

func (c *Cache) cleaner() {
	for {
		<-time.After(c.cleanupInterval)

		if c.items == nil {
			return
		}

		if keys := c.expiredKeys(); len(keys) != 0 {
			c.clearItems(keys)
		}
	}
}

func (c *Cache) expiredKeys() (keys []string) {
	c.RLock()

	for k, i := range c.items {
		if i.Expiration > 0 && time.Now().UnixNano() > i.Expiration {
			keys = append(keys, k)
		}
	}
	c.RUnlock()
	return
}

func (c *Cache) clearItems(keys []string) {
	c.Lock()

	for _, k := range keys {
		delete(c.items, k)
	}

	c.Unlock()
}



func (c *Cache) FlushAll() {
	c.Lock()
	c.items = make(map[string]Item)
	c.Unlock()
}
