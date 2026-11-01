package pokemon

import (
	"sync"
	"time"
)

const (
	DefaultCacheTTL      = time.Hour
	DefaultMaxCacheItems = 500
)

type cacheItem struct {
	pokemon   PokemonResponse
	expiresAt time.Time
}

type Cache struct {
	mu       sync.RWMutex
	items    map[int]cacheItem
	ttl      time.Duration
	maxItems int
}

func NewCache() *Cache {
	return &Cache{
		items:    make(map[int]cacheItem),
		ttl:      DefaultCacheTTL,
		maxItems: DefaultMaxCacheItems,
	}
}

func NewCacheWithConfig(ttl time.Duration, maxItems int) *Cache {
	return &Cache{
		items:    make(map[int]cacheItem),
		ttl:      ttl,
		maxItems: maxItems,
	}
}

func (c *Cache) Get(id int) (PokemonResponse, bool) {
	c.mu.RLock()
	item, exists := c.items[id]
	c.mu.RUnlock()

	if !exists {
		return PokemonResponse{}, false
	}

	if time.Now().After(item.expiresAt) {
		c.mu.Lock()
		// Check again because another request may have refreshed it.
		if current, ok := c.items[id]; ok &&
			current.expiresAt.Equal(item.expiresAt) {
			delete(c.items, id)
		}
		c.mu.Unlock()

		return PokemonResponse{}, false
	}

	return item.pokemon, true
}

func (c *Cache) Set(id int, pokemon PokemonResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove expired entries before inserting.
	now := time.Now()
	for key, item := range c.items {
		if now.After(item.expiresAt) {
			delete(c.items, key)
		}
	}

	// Simple eviction: remove an arbitrary entry when full.
	if _, exists := c.items[id]; !exists &&
		len(c.items) >= c.maxItems {
		for key := range c.items {
			delete(c.items, key)
			break
		}
	}

	c.items[id] = cacheItem{
		pokemon:   pokemon,
		expiresAt: now.Add(c.ttl),
	}
}

func (c *Cache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()

	for id, item := range c.items {
		if now.After(item.expiresAt) {
			delete(c.items, id)
		}
	}
}

func (c *Cache) StartCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)

	go func() {
		for range ticker.C {
			c.cleanup()
		}
	}()
}
