package pokemon

import (
	"sync"
	"time"
)

const (
	DefaultCacheTTL        = time.Hour
	DefaultMaxCacheItems   = 500
	DefaultCleanupInterval = 5 * time.Minute
)

type Cache[V any] struct {
	mu       sync.RWMutex
	items    map[string]cacheItem[V]
	ttl      time.Duration
	maxItems int

	stop     chan struct{}
	stopOnce sync.Once
}

type cacheItem[V any] struct {
	value     V
	expiresAt time.Time
}

func NewCache[V any]() *Cache[V] {
	return NewCacheWithConfig[V](
		DefaultCacheTTL,
		DefaultMaxCacheItems,
		DefaultCleanupInterval,
	)
}

func NewCacheWithConfig[V any](
	ttl time.Duration,
	maxItems int,
	cleanupInterval time.Duration,
) *Cache[V] {
	if ttl <= 0 {
		ttl = DefaultCacheTTL
	}

	if maxItems <= 0 {
		maxItems = DefaultMaxCacheItems
	}

	if cleanupInterval <= 0 {
		cleanupInterval = DefaultCleanupInterval
	}

	cache := &Cache[V]{
		items:    make(map[string]cacheItem[V]),
		ttl:      ttl,
		maxItems: maxItems,
		stop:     make(chan struct{}),
	}

	cache.StartCleanup(cleanupInterval)

	return cache
}

func (c *Cache[V]) Get(key string) (V, bool) {
	c.mu.RLock()
	item, exists := c.items[key]
	c.mu.RUnlock()

	if !exists {
		var zero V
		return zero, false
	}

	if time.Now().After(item.expiresAt) {
		c.mu.Lock()

		// Check again because another request may have refreshed it.
		if current, ok := c.items[key]; ok &&
			current.expiresAt.Equal(item.expiresAt) {
			delete(c.items, key)
		}

		c.mu.Unlock()

		var zero V
		return zero, false
	}

	return item.value, true
}

func (c *Cache[V]) Set(key string, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()

	// Remove expired entries before inserting.
	for existingKey, item := range c.items {
		if now.After(item.expiresAt) {
			delete(c.items, existingKey)
		}
	}

	// Simple eviction: remove an arbitrary entry when full.
	if _, exists := c.items[key]; !exists &&
		len(c.items) >= c.maxItems {
		for existingKey := range c.items {
			delete(c.items, existingKey)
			break
		}
	}

	c.items[key] = cacheItem[V]{
		value:     value,
		expiresAt: now.Add(c.ttl),
	}
}

func (c *Cache[V]) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

func (c *Cache[V]) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()

	for key, item := range c.items {
		if now.After(item.expiresAt) {
			delete(c.items, key)
		}
	}
}

func (c *Cache[V]) StartCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.cleanup()

			case <-c.stop:
				return
			}
		}
	}()
}

func (c *Cache[V]) Close() {
	c.stopOnce.Do(func() {
		close(c.stop)
	})
}
