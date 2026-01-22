package semantic

import (
	"sync"
	"time"
)

// SemanticCache provides caching for semantic routing
type SemanticCache struct {
	cache   map[string]*cacheEntry
	encoder Encoder
	ttl     time.Duration
	mu      sync.RWMutex
}

type cacheEntry struct {
	route     *Route
	embedding []float32
	timestamp time.Time
}

// NewSemanticCache creates a new semantic cache
func NewSemanticCache(ttl time.Duration, encoder Encoder) *SemanticCache {
	c := &SemanticCache{
		cache:   make(map[string]*cacheEntry),
		encoder: encoder,
		ttl:     ttl,
	}

	// Start cleanup goroutine
	go c.cleanup()

	return c
}

// Get retrieves a cached route for a query
func (c *SemanticCache) Get(query string) *Route {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Direct match
	if entry, exists := c.cache[query]; exists {
		if time.Since(entry.timestamp) < c.ttl {
			return entry.route
		}
	}

	return nil
}

// GetSemantic retrieves a cached route using semantic similarity
func (c *SemanticCache) GetSemantic(queryEmbedding []float32, threshold float64) *Route {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var bestRoute *Route
	bestScore := threshold

	for _, entry := range c.cache {
		if time.Since(entry.timestamp) >= c.ttl {
			continue
		}

		score := cosineSimilarity(queryEmbedding, entry.embedding)
		if score > bestScore {
			bestScore = score
			bestRoute = entry.route
		}
	}

	return bestRoute
}

// Set caches a route for a query
func (c *SemanticCache) Set(query string, route *Route) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry := &cacheEntry{
		route:     route,
		timestamp: time.Now(),
	}

	c.cache[query] = entry
}

// SetWithEmbedding caches a route with its embedding
func (c *SemanticCache) SetWithEmbedding(query string, route *Route, embedding []float32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[query] = &cacheEntry{
		route:     route,
		embedding: embedding,
		timestamp: time.Now(),
	}
}

// Clear clears all cached entries
func (c *SemanticCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[string]*cacheEntry)
}

// Size returns the number of cached entries
func (c *SemanticCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cache)
}

// cleanup periodically removes expired entries
func (c *SemanticCache) cleanup() {
	ticker := time.NewTicker(c.ttl / 2)
	defer ticker.Stop()

	for range ticker.C {
		c.removeExpired()
	}
}

func (c *SemanticCache) removeExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.cache {
		if now.Sub(entry.timestamp) >= c.ttl {
			delete(c.cache, key)
		}
	}
}

// Stats returns cache statistics
type CacheStats struct {
	Size        int           `json:"size"`
	OldestEntry *time.Time    `json:"oldest_entry,omitempty"`
	NewestEntry *time.Time    `json:"newest_entry,omitempty"`
	TTL         time.Duration `json:"ttl"`
}

// GetStats returns cache statistics
func (c *SemanticCache) GetStats() *CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := &CacheStats{
		Size: len(c.cache),
		TTL:  c.ttl,
	}

	for _, entry := range c.cache {
		if stats.OldestEntry == nil || entry.timestamp.Before(*stats.OldestEntry) {
			t := entry.timestamp
			stats.OldestEntry = &t
		}
		if stats.NewestEntry == nil || entry.timestamp.After(*stats.NewestEntry) {
			t := entry.timestamp
			stats.NewestEntry = &t
		}
	}

	return stats
}
