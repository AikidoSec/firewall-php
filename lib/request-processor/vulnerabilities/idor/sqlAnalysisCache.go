package idor

import (
	"container/list"
	"sync"
)

const maxCacheEntries = 1000

type cacheEntry struct {
	key     string
	results []SqlQueryResult
}

type analyzeCache struct {
	mu    sync.Mutex
	items map[string]*list.Element
	order *list.List
}

var sqlAnalysisCache = &analyzeCache{
	items: make(map[string]*list.Element),
	order: list.New(),
}

func (c *analyzeCache) get(key string) ([]SqlQueryResult, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	el, ok := c.items[key]
	if !ok {
		return nil, false
	}
	c.order.MoveToFront(el)
	return el.Value.(*cacheEntry).results, true
}

func (c *analyzeCache) set(key string, results []SqlQueryResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if el, ok := c.items[key]; ok {
		el.Value.(*cacheEntry).results = results
		c.order.MoveToFront(el)
		return
	}

	el := c.order.PushFront(&cacheEntry{key: key, results: results})
	c.items[key] = el
	if c.order.Len() > maxCacheEntries {
		oldest := c.order.Back()
		if oldest != nil {
			c.order.Remove(oldest)
			delete(c.items, oldest.Value.(*cacheEntry).key)
		}
	}
}
