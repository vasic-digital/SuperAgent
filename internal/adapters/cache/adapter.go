// Package cache provides an adapter layer between HelixAgent's internal cache
// package and the extracted digital.vasic.cache module.
//
// This adapter bridges the generic cache module with HelixAgent-specific types
// like LLMRequest, LLMResponse, MemorySource, and UserSession from internal/models.
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"digital.vasic.cache/pkg/cache"
	cacheMemory "digital.vasic.cache/pkg/memory"
	cacheRedis "digital.vasic.cache/pkg/redis"

	"dev.helix.agent/internal/config"
)

// RedisClientAdapter wraps the extracted cache module's Redis client
// to provide the interface expected by HelixAgent's internal packages.
type RedisClientAdapter struct {
	client *cacheRedis.Client
}

// NewRedisClientAdapter creates a new Redis client adapter from HelixAgent config
func NewRedisClientAdapter(cfg *config.Config) *RedisClientAdapter {
	if cfg == nil {
		// Return adapter with nil client that will fail on operations
		return &RedisClientAdapter{client: nil}
	}

	redisCfg := &cacheRedis.Config{
		Addr:         cfg.Redis.Host + ":" + cfg.Redis.Port,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     10,
		MinIdleConns: 2,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	return &RedisClientAdapter{
		client: cacheRedis.New(redisCfg),
	}
}

// NewRedisClientAdapterFromClient creates an adapter from an existing cache.Cache
func NewRedisClientAdapterFromClient(client *cacheRedis.Client) *RedisClientAdapter {
	return &RedisClientAdapter{client: client}
}

// Set stores a value with JSON serialization
func (r *RedisClientAdapter) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if r.client == nil {
		return fmt.Errorf("redis client not initialized")
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal value: %w", err)
	}

	return r.client.Set(ctx, key, data, expiration)
}

// Get retrieves and deserializes a value
func (r *RedisClientAdapter) Get(ctx context.Context, key string, dest interface{}) error {
	if r.client == nil {
		return fmt.Errorf("redis client not initialized")
	}

	data, err := r.client.Get(ctx, key)
	if err != nil {
		return err
	}
	if data == nil {
		return fmt.Errorf("key not found")
	}

	return json.Unmarshal(data, dest)
}

// Delete removes a key
func (r *RedisClientAdapter) Delete(ctx context.Context, key string) error {
	if r.client == nil {
		return fmt.Errorf("redis client not initialized")
	}

	return r.client.Delete(ctx, key)
}

// Exists checks if a key exists
func (r *RedisClientAdapter) Exists(ctx context.Context, key string) (bool, error) {
	if r.client == nil {
		return false, fmt.Errorf("redis client not initialized")
	}

	return r.client.Exists(ctx, key)
}

// Ping checks Redis connectivity
func (r *RedisClientAdapter) Ping(ctx context.Context) error {
	if r.client == nil {
		return fmt.Errorf("redis client not initialized")
	}

	return r.client.HealthCheck(ctx)
}

// Close closes the Redis connection
func (r *RedisClientAdapter) Close() error {
	if r.client == nil {
		return nil
	}

	return r.client.Close()
}

// Underlying returns the underlying cache.Cache for advanced operations
func (r *RedisClientAdapter) Underlying() cache.Cache {
	return r.client
}

// UnderlyingRedis returns the underlying Redis client for direct operations
func (r *RedisClientAdapter) UnderlyingRedis() *cacheRedis.Client {
	return r.client
}

// MemoryCacheAdapter wraps the extracted cache module's in-memory cache
type MemoryCacheAdapter struct {
	cache *cacheMemory.Cache
}

// NewMemoryCacheAdapter creates a new in-memory cache adapter
func NewMemoryCacheAdapter(maxEntries int, defaultTTL time.Duration) *MemoryCacheAdapter {
	cfg := &cacheMemory.Config{
		MaxEntries:      maxEntries,
		DefaultTTL:      defaultTTL,
		CleanupInterval: time.Minute,
		EvictionPolicy:  cache.LRU,
	}

	return &MemoryCacheAdapter{
		cache: cacheMemory.New(cfg),
	}
}

// Set stores a value with JSON serialization
func (m *MemoryCacheAdapter) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal value: %w", err)
	}

	return m.cache.Set(ctx, key, data, ttl)
}

// Get retrieves and deserializes a value
func (m *MemoryCacheAdapter) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := m.cache.Get(ctx, key)
	if err != nil {
		return err
	}
	if data == nil {
		return fmt.Errorf("key not found")
	}

	return json.Unmarshal(data, dest)
}

// Delete removes a key
func (m *MemoryCacheAdapter) Delete(ctx context.Context, key string) error {
	return m.cache.Delete(ctx, key)
}

// Exists checks if a key exists
func (m *MemoryCacheAdapter) Exists(ctx context.Context, key string) (bool, error) {
	return m.cache.Exists(ctx, key)
}

// Close closes the cache
func (m *MemoryCacheAdapter) Close() error {
	return m.cache.Close()
}

// Underlying returns the underlying cache.Cache
func (m *MemoryCacheAdapter) Underlying() cache.Cache {
	return m.cache
}

// Stats returns cache statistics
func (m *MemoryCacheAdapter) Stats() *cache.Stats {
	return m.cache.Stats()
}

// Len returns the number of entries in the cache
func (m *MemoryCacheAdapter) Len() int {
	return m.cache.Len()
}

// Flush removes all entries
func (m *MemoryCacheAdapter) Flush() {
	m.cache.Flush()
}

// TypedCacheAdapter provides a generic typed cache wrapper
type TypedCacheAdapter[T any] struct {
	inner cache.Cache
}

// NewTypedCacheAdapter creates a typed cache adapter
func NewTypedCacheAdapter[T any](c cache.Cache) *TypedCacheAdapter[T] {
	return &TypedCacheAdapter[T]{inner: c}
}

// Get retrieves a typed value
func (tc *TypedCacheAdapter[T]) Get(ctx context.Context, key string) (T, bool, error) {
	var zero T
	data, err := tc.inner.Get(ctx, key)
	if err != nil {
		return zero, false, fmt.Errorf("typed cache get: %w", err)
	}
	if data == nil {
		return zero, false, nil
	}

	var val T
	if err := json.Unmarshal(data, &val); err != nil {
		return zero, false, fmt.Errorf("typed cache unmarshal: %w", err)
	}

	return val, true, nil
}

// Set stores a typed value
func (tc *TypedCacheAdapter[T]) Set(ctx context.Context, key string, value T, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("typed cache marshal: %w", err)
	}

	return tc.inner.Set(ctx, key, data, ttl)
}

// Delete removes a key
func (tc *TypedCacheAdapter[T]) Delete(ctx context.Context, key string) error {
	return tc.inner.Delete(ctx, key)
}

// Exists checks if a key exists
func (tc *TypedCacheAdapter[T]) Exists(ctx context.Context, key string) (bool, error) {
	return tc.inner.Exists(ctx, key)
}

// CacheConfig re-exports the extracted module's Config
type CacheConfig = cache.Config

// EvictionPolicy re-exports the extracted module's EvictionPolicy
type EvictionPolicy = cache.EvictionPolicy

// Policy constants
const (
	LRU  = cache.LRU
	LFU  = cache.LFU
	FIFO = cache.FIFO
)

// Stats re-exports the extracted module's Stats
type Stats = cache.Stats

// DefaultConfig returns the default cache configuration
func DefaultConfig() *CacheConfig {
	return cache.DefaultConfig()
}

// RedisConfig re-exports the extracted module's Redis Config
type RedisConfig = cacheRedis.Config

// DefaultRedisConfig returns the default Redis configuration
func DefaultRedisConfig() *RedisConfig {
	return cacheRedis.DefaultConfig()
}

// MemoryConfig re-exports the extracted module's Memory Config
type MemoryConfig = cacheMemory.Config

// DefaultMemoryConfig returns the default in-memory cache configuration
func DefaultMemoryConfig() *MemoryConfig {
	return cacheMemory.DefaultConfig()
}
