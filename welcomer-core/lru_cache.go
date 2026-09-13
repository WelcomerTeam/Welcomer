package welcomer

import "sync"

type LRUCache[K comparable, V any] struct {
	mu       sync.RWMutex
	capacity int
	cache    map[K]V
	order    []K
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		mu:       sync.RWMutex{},
		capacity: capacity,
		cache:    make(map[K]V),
		order:    make([]K, 0, capacity),
	}
}

func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	value, ok := c.cache[key]
	if !ok {
		var zero V

		return zero, false
	}

	c.moveToFront(key)

	return value, true
}

func (c *LRUCache[K, V]) Put(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.cache[key]; ok {
		c.cache[key] = value
		c.moveToFront(key)

		return
	}

	if len(c.cache) >= c.capacity {
		oldestKey := c.order[len(c.order)-1]
		delete(c.cache, oldestKey)
		c.order = c.order[:len(c.order)-1]
	}

	c.cache[key] = value
	c.order = append([]K{key}, c.order...)
}

func (c *LRUCache[K, V]) moveToFront(key K) {
	for i, k := range c.order {
		if k == key {
			c.order = append([]K{key}, append(c.order[:i], c.order[i+1:]...)...)

			return
		}
	}
}
