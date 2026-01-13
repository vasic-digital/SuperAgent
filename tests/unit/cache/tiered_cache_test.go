package cache

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/cache"
)

func TestTieredCache_BasicOperations(t *testing.T) {
	config := &cache.TieredCacheConfig{
		L1MaxSize: 100,
		L1TTL:     time.Minute,
		EnableL1:  true,
		EnableL2:  false, // Test L1 only
	}
	tc := cache.NewTieredCache(nil, config)
	defer tc.Close()

	ctx := context.Background()

	// Set a value
	err := tc.Set(ctx, "key1", map[string]string{"hello": "world"}, time.Minute)
	require.NoError(t, err)

	// Get the value
	var result map[string]string
	found, err := tc.Get(ctx, "key1", &result)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "world", result["hello"])

	// Get non-existent key
	var missing map[string]string
	found, err = tc.Get(ctx, "nonexistent", &missing)
	require.NoError(t, err)
	assert.False(t, found)
}

func TestTieredCache_Expiration(t *testing.T) {
	config := &cache.TieredCacheConfig{
		L1MaxSize:         100,
		L1TTL:             50 * time.Millisecond,
		L1CleanupInterval: 10 * time.Millisecond,
		EnableL1:          true,
		EnableL2:          false,
	}
	tc := cache.NewTieredCache(nil, config)
	defer tc.Close()

	ctx := context.Background()

	err := tc.Set(ctx, "expiring", "value", 50*time.Millisecond)
	require.NoError(t, err)

	// Should exist initially
	var result string
	found, _ := tc.Get(ctx, "expiring", &result)
	assert.True(t, found)

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should be expired
	found, _ = tc.Get(ctx, "expiring", &result)
	assert.False(t, found)
}

func TestTieredCache_Delete(t *testing.T) {
	config := &cache.TieredCacheConfig{
		L1MaxSize: 100,
		L1TTL:     time.Minute,
		EnableL1:  true,
		EnableL2:  false,
	}
	tc := cache.NewTieredCache(nil, config)
	defer tc.Close()

	ctx := context.Background()

	tc.Set(ctx, "to_delete", "value", time.Minute)

	err := tc.Delete(ctx, "to_delete")
	require.NoError(t, err)

	var result string
	found, _ := tc.Get(ctx, "to_delete", &result)
	assert.False(t, found)
}

func TestTieredCache_Tags(t *testing.T) {
	config := &cache.TieredCacheConfig{
		L1MaxSize: 100,
		L1TTL:     time.Minute,
		EnableL1:  true,
		EnableL2:  false,
	}
	tc := cache.NewTieredCache(nil, config)
	defer tc.Close()

	ctx := context.Background()

	// Set values with tags
	tc.Set(ctx, "user:1:profile", "profile1", time.Minute, "user:1")
	tc.Set(ctx, "user:1:settings", "settings1", time.Minute, "user:1")
	tc.Set(ctx, "user:2:profile", "profile2", time.Minute, "user:2")

	// Invalidate by tag
	count, err := tc.InvalidateByTag(ctx, "user:1")
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	// User 1 data should be gone
	var result string
	found, _ := tc.Get(ctx, "user:1:profile", &result)
	assert.False(t, found)

	// User 2 data should still exist
	found, _ = tc.Get(ctx, "user:2:profile", &result)
	assert.True(t, found)
}

func TestTieredCache_InvalidatePrefix(t *testing.T) {
	config := &cache.TieredCacheConfig{
		L1MaxSize: 100,
		L1TTL:     time.Minute,
		EnableL1:  true,
		EnableL2:  false,
	}
	tc := cache.NewTieredCache(nil, config)
	defer tc.Close()

	ctx := context.Background()

	tc.Set(ctx, "provider:claude:response1", "r1", time.Minute)
	tc.Set(ctx, "provider:claude:response2", "r2", time.Minute)
	tc.Set(ctx, "provider:gemini:response1", "r3", time.Minute)

	count, err := tc.InvalidatePrefix(ctx, "provider:claude:")
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	var result string
	found, _ := tc.Get(ctx, "provider:claude:response1", &result)
	assert.False(t, found)

	found, _ = tc.Get(ctx, "provider:gemini:response1", &result)
	assert.True(t, found)
}

func TestTieredCache_Metrics(t *testing.T) {
	config := &cache.TieredCacheConfig{
		L1MaxSize: 100,
		L1TTL:     time.Minute,
		EnableL1:  true,
		EnableL2:  false,
	}
	tc := cache.NewTieredCache(nil, config)
	defer tc.Close()

	ctx := context.Background()

	// Generate some hits and misses
	tc.Set(ctx, "key", "value", time.Minute)

	var result string
	for i := 0; i < 5; i++ {
		tc.Get(ctx, "key", &result)      // hit
		tc.Get(ctx, "missing", &result)   // miss
	}

	metrics := tc.Metrics()
	assert.Equal(t, int64(5), metrics.L1Hits)
	assert.Equal(t, int64(5), metrics.L1Misses)
}

func TestTieredCache_HitRate(t *testing.T) {
	config := &cache.TieredCacheConfig{
		L1MaxSize: 100,
		L1TTL:     time.Minute,
		EnableL1:  true,
		EnableL2:  false,
	}
	tc := cache.NewTieredCache(nil, config)
	defer tc.Close()

	ctx := context.Background()

	tc.Set(ctx, "key", "value", time.Minute)

	var result string
	for i := 0; i < 80; i++ {
		tc.Get(ctx, "key", &result)
	}
	for i := 0; i < 20; i++ {
		tc.Get(ctx, "missing", &result)
	}

	hitRate := tc.HitRate()
	assert.InDelta(t, 80.0, hitRate, 1.0)
}

func TestTieredCache_Eviction(t *testing.T) {
	config := &cache.TieredCacheConfig{
		L1MaxSize: 5,
		L1TTL:     time.Minute,
		EnableL1:  true,
		EnableL2:  false,
	}
	tc := cache.NewTieredCache(nil, config)
	defer tc.Close()

	ctx := context.Background()

	// Fill the cache beyond capacity
	for i := 0; i < 10; i++ {
		tc.Set(ctx, string(rune('a'+i)), i, time.Minute)
	}

	metrics := tc.Metrics()
	assert.True(t, metrics.L1Evictions > 0)
	assert.LessOrEqual(t, metrics.L1Size, int64(5))
}

func TestTieredCache_ConcurrentAccess(t *testing.T) {
	config := &cache.TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
		EnableL2:  false,
	}
	tc := cache.NewTieredCache(nil, config)
	defer tc.Close()

	ctx := context.Background()
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			key := string(rune('a' + idx%26))
			tc.Set(ctx, key, idx, time.Minute)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			key := string(rune('a' + idx%26))
			var result int
			tc.Get(ctx, key, &result)
		}(i)
	}

	wg.Wait()
}

func TestProviderCache_BasicOperations(t *testing.T) {
	config := &cache.TieredCacheConfig{
		L1MaxSize: 100,
		L1TTL:     time.Minute,
		EnableL1:  true,
		EnableL2:  false,
	}
	tc := cache.NewTieredCache(nil, config)
	defer tc.Close()

	pc := cache.NewProviderCache(tc, nil)

	// Test cache key generation
	// Note: This test would need actual LLMRequest models
	// For now, just verify the cache exists
	assert.NotNil(t, pc)

	metrics := pc.Metrics()
	assert.NotNil(t, metrics)
	assert.Equal(t, int64(0), metrics.Hits)
	assert.Equal(t, int64(0), metrics.Misses)
}

func TestMCPServerCache_BasicOperations(t *testing.T) {
	config := &cache.TieredCacheConfig{
		L1MaxSize: 100,
		L1TTL:     time.Minute,
		EnableL1:  true,
		EnableL2:  false,
	}
	tc := cache.NewTieredCache(nil, config)
	defer tc.Close()

	mc := cache.NewMCPServerCache(tc, nil)
	ctx := context.Background()

	// Set tool result
	err := mc.SetToolResult(ctx, "filesystem", "filesystem.read_file", map[string]string{"path": "/test"}, "file content")
	require.NoError(t, err)

	// Get tool result
	result, found := mc.GetToolResult(ctx, "filesystem", "filesystem.read_file", map[string]string{"path": "/test"})
	assert.True(t, found)
	assert.Equal(t, "file content", result)

	// Different args should miss
	result, found = mc.GetToolResult(ctx, "filesystem", "filesystem.read_file", map[string]string{"path": "/other"})
	assert.False(t, found)
}

func TestMCPServerCache_NeverCache(t *testing.T) {
	config := &cache.TieredCacheConfig{
		L1MaxSize: 100,
		L1TTL:     time.Minute,
		EnableL1:  true,
		EnableL2:  false,
	}
	tc := cache.NewTieredCache(nil, config)
	defer tc.Close()

	mc := cache.NewMCPServerCache(tc, nil)
	ctx := context.Background()

	// Memory tools should never be cached
	err := mc.SetToolResult(ctx, "memory", "memory.store", nil, "stored")
	require.NoError(t, err)

	result, found := mc.GetToolResult(ctx, "memory", "memory.store", nil)
	assert.False(t, found)
	assert.Nil(t, result)

	metrics := mc.Metrics()
	assert.True(t, metrics.SkippedNeverCache > 0)
}

func TestExpirationManager_BasicOperation(t *testing.T) {
	config := &cache.TieredCacheConfig{
		L1MaxSize: 100,
		L1TTL:     time.Minute,
		EnableL1:  true,
		EnableL2:  false,
	}
	tc := cache.NewTieredCache(nil, config)
	defer tc.Close()

	em := cache.NewExpirationManager(tc, nil)
	em.Start()
	defer em.Stop()

	// Register a validator
	em.RegisterValidator("provider:*", cache.ProviderHealthValidator(5*time.Minute))

	assert.NotNil(t, em.Metrics())
}

func TestInvalidation_TagBased(t *testing.T) {
	inv := cache.NewTagBasedInvalidation()

	// Add tags
	inv.AddTag("key1", "tag1", "tag2")
	inv.AddTag("key2", "tag1")
	inv.AddTag("key3", "tag2")

	// Get keys by tag
	keys := inv.GetKeys("tag1")
	assert.Len(t, keys, 2)
	assert.Contains(t, keys, "key1")
	assert.Contains(t, keys, "key2")

	// Invalidate by tag
	invalidated := inv.InvalidateByTag("tag1")
	assert.Len(t, invalidated, 2)
}

func TestCacheMetricsCollector(t *testing.T) {
	config := &cache.TieredCacheConfig{
		L1MaxSize: 100,
		L1TTL:     time.Minute,
		EnableL1:  true,
		EnableL2:  false,
	}
	tc := cache.NewTieredCache(nil, config)
	defer tc.Close()

	pc := cache.NewProviderCache(tc, nil)
	mc := cache.NewMCPServerCache(tc, nil)

	collector := cache.NewCacheMetricsCollector(tc, pc, mc, nil)

	metrics := collector.Collect()
	assert.NotNil(t, metrics)

	summary := collector.Summary()
	assert.NotNil(t, summary)
	assert.NotZero(t, summary.LastUpdate)
}

func BenchmarkTieredCache_Get(b *testing.B) {
	config := &cache.TieredCacheConfig{
		L1MaxSize: 10000,
		L1TTL:     time.Minute,
		EnableL1:  true,
		EnableL2:  false,
	}
	tc := cache.NewTieredCache(nil, config)
	defer tc.Close()

	ctx := context.Background()
	tc.Set(ctx, "bench_key", "value", time.Minute)

	var result string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tc.Get(ctx, "bench_key", &result)
	}
}

func BenchmarkTieredCache_Set(b *testing.B) {
	config := &cache.TieredCacheConfig{
		L1MaxSize: 100000,
		L1TTL:     time.Minute,
		EnableL1:  true,
		EnableL2:  false,
	}
	tc := cache.NewTieredCache(nil, config)
	defer tc.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tc.Set(ctx, "key", "value", time.Minute)
	}
}

func BenchmarkTieredCache_Parallel(b *testing.B) {
	config := &cache.TieredCacheConfig{
		L1MaxSize: 100000,
		L1TTL:     time.Minute,
		EnableL1:  true,
		EnableL2:  false,
	}
	tc := cache.NewTieredCache(nil, config)
	defer tc.Close()

	ctx := context.Background()
	tc.Set(ctx, "key", "value", time.Minute)

	var counter int64
	var result string

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n := atomic.AddInt64(&counter, 1)
			if n%2 == 0 {
				tc.Get(ctx, "key", &result)
			} else {
				tc.Set(ctx, "key", "value", time.Minute)
			}
		}
	})
}
