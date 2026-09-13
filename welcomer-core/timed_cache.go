package welcomer

import (
	"sync"
	"time"
)

type TTLLRUCache[K comparable, V any] struct {
	mu          sync.RWMutex
	capacity    int
	maxDuration time.Duration
	cache       map[K]cacheItem[V]
	order       []K
}

type cacheItem[V any] struct {
	value      V
	lastAccess time.Time
}

func NewTimedCache[K comparable, V any](capacity int, maxDuration time.Duration) *TTLLRUCache[K, V] {
	return &TTLLRUCache[K, V]{
		mu:          sync.RWMutex{},
		capacity:    capacity,
		maxDuration: maxDuration,
		cache:       make(map[K]cacheItem[V], capacity),
		order:       make([]K, 0, capacity),
	}
}

func (c *TTLLRUCache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	value, ok := c.cache[key]
	if !ok {
		var zero V

		return zero, false
	}

	if time.Since(value.lastAccess) > c.maxDuration {
		delete(c.cache, key)

		var zero V

		return zero, false
	}

	c.moveToFront(key)

	return value.value, true
}

func (c *TTLLRUCache[K, V]) Put(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.cache[key]; ok {
		c.cache[key] = cacheItem[V]{value: value, lastAccess: time.Now()}
		c.moveToFront(key)

		return
	}

	if len(c.cache) >= c.capacity {
		oldestKey := c.order[len(c.order)-1]
		delete(c.cache, oldestKey)
		c.order = c.order[:len(c.order)-1]
	}

	c.cache[key] = cacheItem[V]{value: value, lastAccess: time.Now()}
	c.order = append([]K{key}, c.order...)
}

func (c *TTLLRUCache[K, V]) moveToFront(key K) {
	for i, k := range c.order {
		if k == key {
			c.order = append([]K{key}, append(c.order[:i], c.order[i+1:]...)...)

			return
		}
	}
}
