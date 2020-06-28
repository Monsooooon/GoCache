package lru

import (
	"container/list"
)

// Cache is a LRU cache. It is not safe for concurrent access.
type Cache struct {
	maxBytes  int64
	nBytes    int64
	cache     map[string]*list.Element
	ll        *list.List
	OnEvicted func(key string, value Value)
}

type entry struct {
	key   string
	value Value // store anything
}

// Value represents the data to store in this cache.
// It must have Len() since the cache need to limit its max size
type Value interface {
	Len() int64
}

// New creates a new LRU cache with given max # of bytes and OnEvicted callback function
// if maxBytes is non-postive, the cache has no limit on its max bytes
func New(maxBytes int64, OnEvicted func(key string, value Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		cache:     make(map[string]*list.Element),
		ll:        list.New(),
		OnEvicted: OnEvicted,
	}
}

// Get returns the key's value and move its corresponding element to the front of list
func (c *Cache) Get(key string) (Value, bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return nil, false
}

// RemoveOldest removes the oldest entry and updates cache & nBytes
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele == nil {
		return
	}
	kv := ele.Value.(*entry)
	c.nBytes -= int64(len(kv.key)) + kv.value.Len()
	c.ll.Remove(ele)
	delete(c.cache, kv.key)
	if c.OnEvicted != nil {
		c.OnEvicted(kv.key, kv.value)
	}
}

// Put updates key with value if this kv pair exists, otherwise creates a new kv pair
// cache removes oldest entries until overall nBytes is smaller than maxBytes
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		kv.value = value
		c.nBytes += value.Len() - kv.value.Len()
	} else {
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nBytes += int64(len(key)) + value.Len()
	}

	for c.maxBytes > 0 && c.nBytes > c.maxBytes {
		c.RemoveOldest()
	}
}

// Len returns the number of cache entries
func (c *Cache) Len() int {
	return c.ll.Len()
}
