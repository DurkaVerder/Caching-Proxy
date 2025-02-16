package cache

import (
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type CachedResponse struct {
	Status  int
	Headers http.Header
	Body    []byte
}

type Item struct {
	Value      CachedResponse
	Created    time.Time
	Expiration int64
}

type Cache struct {
	sync.RWMutex
	defaultExpiration time.Duration
	cleanupInterval   time.Duration
	maxSize           int64
	items             map[string]Item
}

func NewCache(defaultExpiration, cleanupInterval time.Duration, maxSize int64) *Cache {
	cache := &Cache{
		defaultExpiration: defaultExpiration,
		cleanupInterval:   cleanupInterval,
		items:             make(map[string]Item),
		maxSize:           maxSize,
	}
	if cleanupInterval > 0 {
		cache.runCleaner()
	}

	return cache
}

func (c *Cache) Set(key string, value CachedResponse, duration time.Duration) {
	var expiration int64

	if duration == 0 {
		duration = c.defaultExpiration
	}

	if duration > 0 {
		expiration = time.Now().Add(duration).UnixNano()
	}

	c.Lock()
	if len(c.items) >= int(c.maxSize) {
		c.deleteRandomItem()
	}

	c.items[key] = Item{
		Value:      value,
		Created:    time.Now(),
		Expiration: expiration,
	}
	c.Unlock()
}

func (c *Cache) Get(key string) (CachedResponse, bool) {
	c.RLock()
	item, found := c.items[key]
	c.RUnlock()

	if !found {
		return CachedResponse{}, false
	}

	if item.Expiration > 0 {
		if time.Now().UnixNano() > item.Expiration {
			return CachedResponse{}, false
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

func (c *Cache) deleteRandomItem() {
	if len(c.items) == 0 {
		return
	}

	keys := make([]string, 0, len(c.items))
	for k := range c.items {
		keys = append(keys, k)
	}

	randomKey := keys[rand.Intn(len(keys))]
	delete(c.items, randomKey)
}
