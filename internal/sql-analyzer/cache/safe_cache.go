package cache

import (
	"sync"

	lru "github.com/hashicorp/golang-lru/v2"
)

// SafeLRUCache is a thread-safe wrapper around an LRU cache.
type SafeLRUCache[K comparable, V any] struct {
	lru  *lru.Cache[K, V]
	lock sync.Mutex
}

// NewSafeLRUCache creates a new thread-safe LRU cache.
func NewSafeLRUCache[K comparable, V any](size int) (*SafeLRUCache[K, V], error) {
	l, err := lru.New[K, V](size)
	if err != nil {
		return nil, err
	}
	return &SafeLRUCache[K, V]{
		lru: l,
	}, nil
}

// Add adds a value to the cache.
func (c *SafeLRUCache[K, V]) Add(key K, value V) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.lru.Add(key, value)
}

// Remove removes a value from the cache.
func (c *SafeLRUCache[K, V]) Remove(key K) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.lru.Remove(key)
}

// Contains checks if a key is in the cache.
func (c *SafeLRUCache[K, V]) Contains(key K) bool {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.lru.Contains(key)
}

// Get retrieves a value from the cache.
func (c *SafeLRUCache[K, V]) Get(key K) (V, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.lru.Get(key)
}

// Len returns the number of items in the cache.
func (c *SafeLRUCache[K, V]) Len() int {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.lru.Len()
}
