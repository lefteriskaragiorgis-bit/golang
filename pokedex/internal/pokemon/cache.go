package pokemon

import "sync"

type Cache struct {
	mu    sync.RWMutex
	items map[int]PokemonResponse
}

func NewCache() *Cache {
	return &Cache{
		items: make(map[int]PokemonResponse),
	}
}

func (c *Cache) Get(id int) (PokemonResponse, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	pokemon, exists := c.items[id]
	return pokemon, exists
}

func (c *Cache) Set(id int, pokemon PokemonResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[id] = pokemon
}
