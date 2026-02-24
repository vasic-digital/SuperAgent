package services

import (
	"context"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	runtime.GOMAXPROCS(2)
}

// newProtocolCacheManagerTestLogger creates a silent logger for cache manager tests.
func newProtocolCacheManagerTestLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	return logger
}

// newTestProtocolCacheManager creates a ProtocolCacheManager with a real
// ProtocolCache for testing. The repo field is nil since database tests
// are not the focus here.
func newTestProtocolCacheManager(t *testing.T) (*ProtocolCacheManager, *ProtocolCache) {
	t.Helper()
	logger := newProtocolCacheManagerTestLogger()
	cache := NewProtocolCache(100, 5*time.Minute, logger)
	mgr := NewProtocolCacheManager(nil, cache, logger)
	return mgr, cache
}

// =============================================================================
// NewProtocolCacheManager Tests
// =============================================================================

func TestNewProtocolCacheManager(t *testing.T) {
	logger := newProtocolCacheManagerTestLogger()
	cache := NewProtocolCache(100, 5*time.Minute, logger)
	defer cache.Stop()

	mgr := NewProtocolCacheManager(nil, cache, logger)

	require.NotNil(t, mgr)
	assert.Nil(t, mgr.repo)
	assert.Equal(t, cache, mgr.cache)
	assert.Equal(t, logger, mgr.log)
}

func TestNewProtocolCacheManager_AllFieldsSet(t *testing.T) {
	logger := newProtocolCacheManagerTestLogger()
	cache := NewProtocolCache(50, 1*time.Minute, logger)
	defer cache.Stop()

	mgr := NewProtocolCacheManager(nil, cache, logger)

	assert.NotNil(t, mgr.cache)
	assert.NotNil(t, mgr.log)
}

// =============================================================================
// Set Tests
// =============================================================================

func TestProtocolCacheManager_Set_Simple(t *testing.T) {
	mgr, cache := newTestProtocolCacheManager(t)
	defer cache.Stop()

	ctx := context.Background()

	err := mgr.Set(ctx, "mcp", "server-list", []string{"server1", "server2"}, 5*time.Minute)
	require.NoError(t, err)

	// Verify via underlying cache
	data, found, err := cache.Get(ctx, "protocol_cache_mcp_server-list")
	require.NoError(t, err)
	assert.True(t, found)
	assert.NotNil(t, data)
}

func TestProtocolCacheManager_Set_DifferentProtocols(t *testing.T) {
	mgr, cache := newTestProtocolCacheManager(t)
	defer cache.Stop()

	ctx := context.Background()

	protocols := []struct {
		protocol string
		key      string
		data     interface{}
	}{
		{"mcp", "tools", map[string]interface{}{"tool1": true}},
		{"lsp", "diagnostics", []string{"error1"}},
		{"acp", "config", "config-data"},
		{"embedding", "vectors", []float64{0.1, 0.2, 0.3}},
	}

	for _, p := range protocols {
		t.Run(p.protocol, func(t *testing.T) {
			err := mgr.Set(ctx, p.protocol, p.key, p.data, 5*time.Minute)
			require.NoError(t, err)
		})
	}
}

func TestProtocolCacheManager_Set_WithTTL(t *testing.T) {
	mgr, cache := newTestProtocolCacheManager(t)
	defer cache.Stop()

	ctx := context.Background()

	err := mgr.Set(ctx, "mcp", "expiring-key", "value", 1*time.Millisecond)
	require.NoError(t, err)

	// Verify it exists immediately
	_, found, _ := cache.Get(ctx, "protocol_cache_mcp_expiring-key")
	assert.True(t, found)

	// Wait for TTL
	time.Sleep(5 * time.Millisecond)

	// Should be expired
	_, found, _ = cache.Get(ctx, "protocol_cache_mcp_expiring-key")
	assert.False(t, found)
}

func TestProtocolCacheManager_Set_OverwriteExisting(t *testing.T) {
	mgr, cache := newTestProtocolCacheManager(t)
	defer cache.Stop()

	ctx := context.Background()

	err := mgr.Set(ctx, "mcp", "key1", "value1", 5*time.Minute)
	require.NoError(t, err)

	err = mgr.Set(ctx, "mcp", "key1", "value2", 5*time.Minute)
	require.NoError(t, err)

	data, found, err := cache.Get(ctx, "protocol_cache_mcp_key1")
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "value2", data)
}

// =============================================================================
// Get Tests
// =============================================================================

func TestProtocolCacheManager_Get_Existing(t *testing.T) {
	mgr, cache := newTestProtocolCacheManager(t)
	defer cache.Stop()

	ctx := context.Background()

	// Set via manager
	err := mgr.Set(ctx, "mcp", "test-key", "test-value", 5*time.Minute)
	require.NoError(t, err)

	// Get via manager
	data, found, err := mgr.Get(ctx, "mcp", "test-key")
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "test-value", data)
}

func TestProtocolCacheManager_Get_NonExistent(t *testing.T) {
	mgr, cache := newTestProtocolCacheManager(t)
	defer cache.Stop()

	ctx := context.Background()

	data, found, err := mgr.Get(ctx, "mcp", "nonexistent")
	require.NoError(t, err)
	assert.False(t, found)
	assert.Nil(t, data)
}

func TestProtocolCacheManager_Get_WrongProtocol(t *testing.T) {
	mgr, cache := newTestProtocolCacheManager(t)
	defer cache.Stop()

	ctx := context.Background()

	err := mgr.Set(ctx, "mcp", "key1", "value1", 5*time.Minute)
	require.NoError(t, err)

	// Same key but different protocol
	data, found, err := mgr.Get(ctx, "lsp", "key1")
	require.NoError(t, err)
	assert.False(t, found)
	assert.Nil(t, data)
}

func TestProtocolCacheManager_Get_AfterExpiry(t *testing.T) {
	mgr, cache := newTestProtocolCacheManager(t)
	defer cache.Stop()

	ctx := context.Background()

	err := mgr.Set(ctx, "mcp", "short-ttl", "value", 1*time.Millisecond)
	require.NoError(t, err)

	time.Sleep(5 * time.Millisecond)

	data, found, err := mgr.Get(ctx, "mcp", "short-ttl")
	require.NoError(t, err)
	assert.False(t, found)
	assert.Nil(t, data)
}

func TestProtocolCacheManager_Get_MapData(t *testing.T) {
	mgr, cache := newTestProtocolCacheManager(t)
	defer cache.Stop()

	ctx := context.Background()

	mapData := map[string]interface{}{
		"tools":  []string{"tool1", "tool2"},
		"count":  42,
		"active": true,
	}

	err := mgr.Set(ctx, "mcp", "complex-data", mapData, 5*time.Minute)
	require.NoError(t, err)

	data, found, err := mgr.Get(ctx, "mcp", "complex-data")
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, mapData, data)
}

// =============================================================================
// Delete Tests
// =============================================================================

func TestProtocolCacheManager_Delete_Existing(t *testing.T) {
	mgr, cache := newTestProtocolCacheManager(t)
	defer cache.Stop()

	ctx := context.Background()

	err := mgr.Set(ctx, "mcp", "to-delete", "value", 5*time.Minute)
	require.NoError(t, err)

	err = mgr.Delete(ctx, "mcp", "to-delete")
	require.NoError(t, err)

	_, found, err := mgr.Get(ctx, "mcp", "to-delete")
	require.NoError(t, err)
	assert.False(t, found)
}

func TestProtocolCacheManager_Delete_NonExistent(t *testing.T) {
	mgr, cache := newTestProtocolCacheManager(t)
	defer cache.Stop()

	ctx := context.Background()

	// Should not error on deleting non-existent key
	err := mgr.Delete(ctx, "mcp", "nonexistent")
	require.NoError(t, err)
}

func TestProtocolCacheManager_Delete_OnlyAffectsTargetKey(t *testing.T) {
	mgr, cache := newTestProtocolCacheManager(t)
	defer cache.Stop()

	ctx := context.Background()

	err := mgr.Set(ctx, "mcp", "key1", "value1", 5*time.Minute)
	require.NoError(t, err)
	err = mgr.Set(ctx, "mcp", "key2", "value2", 5*time.Minute)
	require.NoError(t, err)

	err = mgr.Delete(ctx, "mcp", "key1")
	require.NoError(t, err)

	// key2 should still exist
	_, found, err := mgr.Get(ctx, "mcp", "key2")
	require.NoError(t, err)
	assert.True(t, found)
}

// =============================================================================
// CleanupExpired Tests
// =============================================================================

func TestProtocolCacheManager_CleanupExpired(t *testing.T) {
	mgr, cache := newTestProtocolCacheManager(t)
	defer cache.Stop()

	ctx := context.Background()

	// Should not error even with empty cache
	err := mgr.CleanupExpired(ctx)
	require.NoError(t, err)
}

func TestProtocolCacheManager_CleanupExpired_WithEntries(t *testing.T) {
	mgr, cache := newTestProtocolCacheManager(t)
	defer cache.Stop()

	ctx := context.Background()

	// Set some entries
	err := mgr.Set(ctx, "mcp", "key1", "value1", 5*time.Minute)
	require.NoError(t, err)

	// CleanupExpired invalidates by protocol tags
	err = mgr.CleanupExpired(ctx)
	require.NoError(t, err)
}

// =============================================================================
// GetCacheStats Tests
// =============================================================================

func TestProtocolCacheManager_GetCacheStats_EmptyCache(t *testing.T) {
	mgr, cache := newTestProtocolCacheManager(t)
	defer cache.Stop()

	ctx := context.Background()

	stats, err := mgr.GetCacheStats(ctx)
	require.NoError(t, err)
	require.NotNil(t, stats)

	assert.Equal(t, "protocol", stats["cacheType"])
	assert.NotNil(t, stats["timestamp"])
	assert.Equal(t, 0, stats["totalEntries"])
}

func TestProtocolCacheManager_GetCacheStats_WithEntries(t *testing.T) {
	mgr, cache := newTestProtocolCacheManager(t)
	defer cache.Stop()

	ctx := context.Background()

	// Add entries
	_ = mgr.Set(ctx, "mcp", "key1", "value1", 5*time.Minute)
	_ = mgr.Set(ctx, "lsp", "key2", "value2", 5*time.Minute)

	// Access to generate hits
	_, _, _ = mgr.Get(ctx, "mcp", "key1")

	stats, err := mgr.GetCacheStats(ctx)
	require.NoError(t, err)

	assert.Equal(t, 2, stats["totalEntries"])
	assert.True(t, stats["totalSize"].(int) > 0)
}

func TestProtocolCacheManager_GetCacheStats_ContainsAllFields(t *testing.T) {
	mgr, cache := newTestProtocolCacheManager(t)
	defer cache.Stop()

	ctx := context.Background()

	stats, err := mgr.GetCacheStats(ctx)
	require.NoError(t, err)

	expectedFields := []string{
		"cacheType", "timestamp", "totalEntries", "totalSize",
		"hitRate", "missRate", "totalHits", "totalMisses", "evictions",
	}

	for _, field := range expectedFields {
		assert.Contains(t, stats, field, "Stats should contain field: %s", field)
	}
}

// =============================================================================
// InvalidateByPattern Tests
// =============================================================================

func TestProtocolCacheManager_InvalidateByPattern(t *testing.T) {
	mgr, cache := newTestProtocolCacheManager(t)
	defer cache.Stop()

	ctx := context.Background()

	// Set entries with pattern-matchable keys
	_ = mgr.Set(ctx, "mcp", "server:1", "value1", 5*time.Minute)
	_ = mgr.Set(ctx, "mcp", "server:2", "value2", 5*time.Minute)
	_ = mgr.Set(ctx, "mcp", "tools:1", "value3", 5*time.Minute)

	// Invalidate by pattern
	err := mgr.InvalidateByPattern(ctx, "mcp", "server:*")
	require.NoError(t, err)

	// server entries should be gone (they use protocol_cache_mcp_ prefix internally)
	_, found, _ := mgr.Get(ctx, "mcp", "server:1")
	assert.False(t, found)
	_, found, _ = mgr.Get(ctx, "mcp", "server:2")
	assert.False(t, found)
}

func TestProtocolCacheManager_InvalidateByPattern_NoMatch(t *testing.T) {
	mgr, cache := newTestProtocolCacheManager(t)
	defer cache.Stop()

	ctx := context.Background()

	_ = mgr.Set(ctx, "mcp", "key1", "value1", 5*time.Minute)

	// Pattern that does not match
	err := mgr.InvalidateByPattern(ctx, "mcp", "nomatch:*")
	require.NoError(t, err)

	// Original entry should still exist
	_, found, _ := mgr.Get(ctx, "mcp", "key1")
	assert.True(t, found)
}

// =============================================================================
// SetWithInvalidation Tests
// =============================================================================

func TestProtocolCacheManager_SetWithInvalidation(t *testing.T) {
	mgr, cache := newTestProtocolCacheManager(t)
	defer cache.Stop()

	ctx := context.Background()

	// Set initial entries that will be invalidated
	_ = mgr.Set(ctx, "mcp", "old-key", "old-value", 5*time.Minute)

	// Set new entry and invalidate old
	err := mgr.SetWithInvalidation(ctx, "mcp", "new-key", "new-value", 5*time.Minute, "old-key")
	require.NoError(t, err)

	// New entry should exist
	data, found, _ := mgr.Get(ctx, "mcp", "new-key")
	assert.True(t, found)
	assert.Equal(t, "new-value", data)

	// Old entry should be deleted
	_, found, _ = mgr.Get(ctx, "mcp", "old-key")
	assert.False(t, found)
}

func TestProtocolCacheManager_SetWithInvalidation_MultiplePatterns(t *testing.T) {
	mgr, cache := newTestProtocolCacheManager(t)
	defer cache.Stop()

	ctx := context.Background()

	_ = mgr.Set(ctx, "mcp", "pattern1", "val1", 5*time.Minute)
	_ = mgr.Set(ctx, "mcp", "pattern2", "val2", 5*time.Minute)

	err := mgr.SetWithInvalidation(ctx, "mcp", "new", "new-val", 5*time.Minute, "pattern1,pattern2")
	require.NoError(t, err)

	// Both old entries should be deleted
	_, found, _ := mgr.Get(ctx, "mcp", "pattern1")
	assert.False(t, found)
	_, found, _ = mgr.Get(ctx, "mcp", "pattern2")
	assert.False(t, found)

	// New entry should exist
	_, found, _ = mgr.Get(ctx, "mcp", "new")
	assert.True(t, found)
}

// =============================================================================
// WarmupCache Tests
// =============================================================================

func TestProtocolCacheManager_WarmupCache(t *testing.T) {
	mgr, cache := newTestProtocolCacheManager(t)
	defer cache.Stop()

	ctx := context.Background()

	err := mgr.WarmupCache(ctx)
	require.NoError(t, err)

	// Verify that warmup data was populated
	// The warmup creates protocol_cache_<proto>_registry keys
	protocols := []string{"mcp", "lsp", "acp", "embedding"}
	for _, proto := range protocols {
		key := "protocol_cache_" + proto + "_registry"
		data, found, err := cache.Get(ctx, key)
		require.NoError(t, err)
		assert.True(t, found, "Warmup should have created entry for %s", proto)
		assert.NotNil(t, data)
	}
}

// =============================================================================
// GetProtocolsWithCache Tests
// =============================================================================

func TestProtocolCacheManager_GetProtocolsWithCache_EmptyCache(t *testing.T) {
	mgr, cache := newTestProtocolCacheManager(t)
	defer cache.Stop()

	ctx := context.Background()

	result, err := mgr.GetProtocolsWithCache(ctx)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result)
}

func TestProtocolCacheManager_GetProtocolsWithCache_AfterWarmup(t *testing.T) {
	mgr, cache := newTestProtocolCacheManager(t)
	defer cache.Stop()

	ctx := context.Background()

	err := mgr.WarmupCache(ctx)
	require.NoError(t, err)

	result, err := mgr.GetProtocolsWithCache(ctx)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// After warmup, protocols should have registry entries
	for _, proto := range []string{"mcp", "lsp", "acp", "embedding"} {
		if protoData, ok := result[proto]; ok {
			assert.NotNil(t, protoData)
			assert.True(t, protoData["initialized"].(bool))
		}
	}
}

// =============================================================================
// MonitorCacheHealth Tests
// =============================================================================

func TestProtocolCacheManager_MonitorCacheHealth_EmptyCache(t *testing.T) {
	mgr, cache := newTestProtocolCacheManager(t)
	defer cache.Stop()

	ctx := context.Background()

	err := mgr.MonitorCacheHealth(ctx)
	require.NoError(t, err)
}

func TestProtocolCacheManager_MonitorCacheHealth_WithEntries(t *testing.T) {
	mgr, cache := newTestProtocolCacheManager(t)
	defer cache.Stop()

	ctx := context.Background()

	// Add entries and access them to generate stats
	_ = mgr.Set(ctx, "mcp", "key1", "value1", 5*time.Minute)
	_, _, _ = mgr.Get(ctx, "mcp", "key1")
	_, _, _ = mgr.Get(ctx, "mcp", "key1")

	err := mgr.MonitorCacheHealth(ctx)
	require.NoError(t, err)
}

// =============================================================================
// ProtocolCacheEntry Tests
// =============================================================================

func TestProtocolCacheEntry_Fields(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(5 * time.Minute)

	entry := ProtocolCacheEntry{
		Key:       "test-key",
		Data:      map[string]interface{}{"value": 42},
		ExpiresAt: &expiresAt,
		Protocol:  "mcp",
		Timestamp: now,
	}

	assert.Equal(t, "test-key", entry.Key)
	assert.NotNil(t, entry.Data)
	assert.NotNil(t, entry.ExpiresAt)
	assert.Equal(t, "mcp", entry.Protocol)
	assert.Equal(t, now, entry.Timestamp)
}

func TestProtocolCacheEntry_NilExpiry(t *testing.T) {
	entry := ProtocolCacheEntry{
		Key:       "no-expiry",
		Data:      "value",
		ExpiresAt: nil,
		Protocol:  "lsp",
		Timestamp: time.Now(),
	}

	assert.Nil(t, entry.ExpiresAt)
}

func TestProtocolCacheEntry_Protocols(t *testing.T) {
	protocols := []string{"mcp", "lsp", "acp", "embedding"}

	for _, proto := range protocols {
		t.Run(proto, func(t *testing.T) {
			entry := ProtocolCacheEntry{
				Key:      "key",
				Protocol: proto,
			}
			assert.Equal(t, proto, entry.Protocol)
		})
	}
}

// =============================================================================
// Concurrent Access Tests
// =============================================================================

func TestProtocolCacheManager_ConcurrentSetGet(t *testing.T) {
	mgr, cache := newTestProtocolCacheManager(t)
	defer cache.Stop()

	ctx := context.Background()
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			key := string(rune('A' + idx))
			_ = mgr.Set(ctx, "mcp", key, idx, 5*time.Minute)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			key := string(rune('A' + idx))
			_, _, _ = mgr.Get(ctx, "mcp", key)
		}(i)
	}

	wg.Wait()
}

func TestProtocolCacheManager_ConcurrentSetDelete(t *testing.T) {
	mgr, cache := newTestProtocolCacheManager(t)
	defer cache.Stop()

	ctx := context.Background()
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(2)
		go func(idx int) {
			defer wg.Done()
			key := string(rune('A' + idx))
			_ = mgr.Set(ctx, "mcp", key, idx, 5*time.Minute)
		}(i)
		go func(idx int) {
			defer wg.Done()
			key := string(rune('A' + idx))
			_ = mgr.Delete(ctx, "mcp", key)
		}(i)
	}

	wg.Wait()
}

func TestProtocolCacheManager_ConcurrentStats(t *testing.T) {
	mgr, cache := newTestProtocolCacheManager(t)
	defer cache.Stop()

	ctx := context.Background()
	var wg sync.WaitGroup

	// Write entries while reading stats
	for i := 0; i < 10; i++ {
		wg.Add(2)
		go func(idx int) {
			defer wg.Done()
			key := string(rune('A' + idx))
			_ = mgr.Set(ctx, "mcp", key, idx, 5*time.Minute)
		}(i)
		go func() {
			defer wg.Done()
			_, _ = mgr.GetCacheStats(ctx)
		}()
	}

	wg.Wait()
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkProtocolCacheManager_Set(b *testing.B) {
	logger := newProtocolCacheManagerTestLogger()
	cache := NewProtocolCache(10000, 5*time.Minute, logger)
	defer cache.Stop()
	mgr := NewProtocolCacheManager(nil, cache, logger)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mgr.Set(ctx, "mcp", "key", "value", 5*time.Minute)
	}
}

func BenchmarkProtocolCacheManager_Get(b *testing.B) {
	logger := newProtocolCacheManagerTestLogger()
	cache := NewProtocolCache(10000, 5*time.Minute, logger)
	defer cache.Stop()
	mgr := NewProtocolCacheManager(nil, cache, logger)
	ctx := context.Background()
	_ = mgr.Set(ctx, "mcp", "key", "value", 5*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = mgr.Get(ctx, "mcp", "key")
	}
}

func BenchmarkProtocolCacheManager_GetCacheStats(b *testing.B) {
	logger := newProtocolCacheManagerTestLogger()
	cache := NewProtocolCache(10000, 5*time.Minute, logger)
	defer cache.Stop()
	mgr := NewProtocolCacheManager(nil, cache, logger)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = mgr.GetCacheStats(ctx)
	}
}
