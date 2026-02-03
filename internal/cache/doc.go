// Package cache provides caching layer for HelixAgent.
//
// This package implements a multi-tier caching system with Redis and in-memory
// caching for improved performance and reduced API calls. It wraps the extracted
// digital.vasic.cache module while providing HelixAgent-specific functionality.
//
// # Architecture
//
// The cache package uses the extracted digital.vasic.cache module for core
// caching operations (Redis client, in-memory cache, eviction policies) while
// adding HelixAgent-specific features:
//
//   - LLM response caching with request-based key generation
//   - Provider health caching
//   - User session caching with per-user invalidation
//   - MCP tool result caching with tool-specific TTLs
//   - Event-driven cache invalidation
//
// # Module Integration
//
// The package uses digital.vasic.cache for:
//   - Core Cache interface (Get, Set, Delete, Exists, Close)
//   - Redis client implementation (pkg/redis)
//   - In-memory cache with LRU/LFU/FIFO eviction (pkg/memory)
//   - Typed cache wrappers (TypedCache[T])
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
//	    Close() error
//	}
//
// # Redis Cache
//
// Primary distributed cache using the extracted module:
//
//	redisClient := cache.NewRedisClient(cfg)
//
//	// Set with TTL
//	err := redisClient.Set(ctx, "key", data, 5*time.Minute)
//
//	// Get
//	err := redisClient.Get(ctx, "key", &dest)
//
// # Tiered Cache
//
// Combines L1 (memory) and L2 (Redis) caches:
//
//	tieredCache := cache.NewTieredCache(redisClient.Client(), config)
//
//	// Checks L1 first, then L2
//	found, err := tieredCache.Get(ctx, "key", &dest)
//
//	// Writes to both L1 and L2
//	err := tieredCache.Set(ctx, "key", value, ttl, tags...)
//
// # Cache Service
//
// High-level caching service with HelixAgent-specific methods:
//
//	service, err := cache.NewCacheService(cfg)
//
//	// Cache LLM responses
//	err := service.SetLLMResponse(ctx, req, resp, ttl)
//	resp, err := service.GetLLMResponse(ctx, req)
//
//	// Cache user sessions
//	err := service.SetUserSession(ctx, session, ttl)
//	session, err := service.GetUserSession(ctx, sessionID)
//
// # Provider Cache
//
// Specialized cache for LLM provider responses:
//
//	providerCache := cache.NewProviderCache(tieredCache, config)
//	resp, found := providerCache.Get(ctx, req, "claude")
//	err := providerCache.Set(ctx, req, resp, "claude")
//
// # MCP Cache
//
// Specialized cache for MCP tool execution results:
//
//	mcpCache := cache.NewMCPServerCache(tieredCache, config)
//	result, found := mcpCache.GetToolResult(ctx, "filesystem", "read_file", args)
//	err := mcpCache.SetToolResult(ctx, "filesystem", "read_file", args, result)
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
//	// Tag-based invalidation
//	count, err := tieredCache.InvalidateByTag(ctx, "provider:claude")
//
//	// Pattern-based invalidation
//	count, err := tieredCache.InvalidatePrefix(ctx, "provider:")
//
//	// User-based invalidation
//	err := service.InvalidateUserCache(ctx, userID)
//
// # Event-Driven Invalidation
//
// Automatic cache invalidation based on system events:
//
//	inv := cache.NewEventDrivenInvalidation(eventBus, tieredCache)
//	inv.Start()
//	// Listens for provider.health.changed, mcp.server.disconnected, etc.
//	// and invalidates relevant cache entries
//
// # Graceful Degradation
//
// Cache failures don't break the system:
//
//	response, err := service.GetLLMResponse(ctx, req)
//	if err != nil {
//	    // Cache miss or disabled - proceed without cache
//	}
//
// # Configuration
//
//	config := &cache.TieredCacheConfig{
//	    L1MaxSize:   10000,
//	    L1TTL:       5 * time.Minute,
//	    L2TTL:       30 * time.Minute,
//	    EnableL1:    true,
//	    EnableL2:    true,
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
//   - redis.go: Redis client wrapper (uses digital.vasic.cache/pkg/redis)
//   - tiered_cache.go: L1+L2 tiered cache
//   - cache_service.go: High-level cache service for HelixAgent
//   - provider_cache.go: LLM provider response cache
//   - mcp_cache.go: MCP tool result cache
//   - invalidation.go: Event-driven invalidation
//   - expiration.go: TTL and validation management
//   - metrics.go: Cache metrics collection
//
// # Adapter Package
//
// The internal/adapters/cache package provides additional adapters for
// integration with the extracted digital.vasic.cache module:
//
//   - RedisClientAdapter: Wraps extracted module for HelixAgent config
//   - MemoryCacheAdapter: In-memory cache with JSON serialization
//   - TypedCacheAdapter[T]: Generic typed cache wrapper
package cache
