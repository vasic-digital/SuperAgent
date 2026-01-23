// Package cache provides caching layer for HelixAgent.
//
// This package implements a multi-tier caching system with Redis and in-memory
// caching for improved performance and reduced API calls.
//
// # Cache Architecture
//
// Two-tier caching system:
//
//  1. L1 Cache: In-memory (fast, limited size)
//  2. L2 Cache: Redis (larger capacity, shared across instances)
//
// # Cache Interface
//
// Common interface for all cache implementations:
//
//	type CacheInterface interface {
//	    Get(ctx context.Context, key string) ([]byte, error)
//	    Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
//	    Delete(ctx context.Context, key string) error
//	    Exists(ctx context.Context, key string) (bool, error)
//	    Clear(ctx context.Context) error
//	}
//
// # Redis Cache
//
// Primary distributed cache:
//
//	redisCache := cache.NewRedisCache(&RedisConfig{
//	    Host:     "localhost",
//	    Port:     6379,
//	    Password: "secret",
//	    DB:       0,
//	    PoolSize: 10,
//	})
//
//	// Set with TTL
//	err := redisCache.Set(ctx, "key", data, 5*time.Minute)
//
//	// Get
//	data, err := redisCache.Get(ctx, "key")
//
// # In-Memory Cache
//
// Fast local cache:
//
//	memCache := cache.NewMemoryCache(&MemoryConfig{
//	    MaxSize:      1000,
//	    DefaultTTL:   5 * time.Minute,
//	    CleanupInterval: 1 * time.Minute,
//	})
//
// # Tiered Cache
//
// Combines L1 and L2 caches:
//
//	tieredCache := cache.NewTieredCache(memCache, redisCache)
//
//	// Checks L1 first, then L2
//	data, err := tieredCache.Get(ctx, "key")
//
//	// Writes to both L1 and L2
//	err := tieredCache.Set(ctx, "key", data, ttl)
//
// # Cache Service
//
// High-level caching service:
//
//	service := cache.NewCacheService(tieredCache, logger)
//
//	// Cache LLM responses
//	response, err := service.GetOrSet(ctx, cacheKey, func() (interface{}, error) {
//	    return llmProvider.Complete(ctx, request)
//	}, 10*time.Minute)
//
// # Cache Keys
//
// Consistent key generation:
//
//	// For LLM completions
//	key := cache.CompletionKey(provider, model, prompt)
//
//	// For embeddings
//	key := cache.EmbeddingKey(provider, model, text)
//
//	// For debate results
//	key := cache.DebateKey(topic, participants)
//
// # Cache Invalidation
//
// Invalidation strategies:
//
//	// TTL-based (automatic)
//	cache.Set(ctx, key, data, 5*time.Minute)
//
//	// Manual invalidation
//	cache.Delete(ctx, key)
//
//	// Pattern-based (Redis)
//	cache.DeletePattern(ctx, "completion:claude:*")
//
//	// Clear all
//	cache.Clear(ctx)
//
// # Graceful Degradation
//
// Cache failures don't break the system:
//
//	// CacheService handles failures gracefully
//	response, err := service.GetOrSet(ctx, key, fetchFunc, ttl)
//	// If cache is down, fetchFunc is called directly
//
// # Configuration
//
//	config := &cache.Config{
//	    Enabled:     true,
//	    RedisURL:    "redis://localhost:6379",
//	    MaxMemory:   "100mb",
//	    DefaultTTL:  5 * time.Minute,
//	    EnableL1:    true,
//	    L1MaxSize:   1000,
//	}
//
// # Environment Variables
//
//	REDIS_HOST     - Redis host (default: localhost)
//	REDIS_PORT     - Redis port (default: 6379)
//	REDIS_PASSWORD - Redis password
//	REDIS_DB       - Redis database number
//
// # Key Files
//
//   - redis.go: Redis cache implementation
//   - memory.go: In-memory cache
//   - tiered_cache.go: Tiered cache (L1+L2)
//   - cache_service.go: High-level cache service
//   - keys.go: Key generation utilities
//
// # Example: Caching LLM Responses
//
//	service := cache.NewCacheService(tieredCache, logger)
//
//	cacheKey := cache.CompletionKey("claude", "claude-3", prompt)
//
//	response, err := service.GetOrSet(ctx, cacheKey, func() (interface{}, error) {
//	    return claudeProvider.Complete(ctx, &CompletionRequest{
//	        Prompt: prompt,
//	    })
//	}, 30*time.Minute)
package cache
