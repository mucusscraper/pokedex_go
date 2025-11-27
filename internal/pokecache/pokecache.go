package pokecache

import (
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

type Cache struct {
	holdsMap   map[string]cacheEntry
	holdsMutex *sync.Mutex
	interval   time.Duration
}

func NewCache(interval time.Duration) *Cache {
	holdsMap := map[string]cacheEntry{}
	holdsMutex := &sync.Mutex{}
	nc := &Cache{
		holdsMap:   holdsMap,
		holdsMutex: holdsMutex,
		interval:   interval,
	}
	go nc.reapLoop()
	return nc
}

func (c *Cache) Add(key string, val []byte) {
	c.holdsMutex.Lock()
	defer c.holdsMutex.Unlock()
	c.holdsMap[key] = cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.holdsMutex.Lock()
	defer c.holdsMutex.Unlock()
	if val, ok := c.holdsMap[key]; ok {
		return val.val, true
	}
	return nil, false
}
func (c *Cache) reapLoop() {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	for _ = range ticker.C {
		c.holdsMutex.Lock()
		for key, c_e := range c.holdsMap {
			if time.Now().Sub(c_e.createdAt) > c.interval {
				delete(c.holdsMap, key)
			}
		}
		c.holdsMutex.Unlock()
	}
}
