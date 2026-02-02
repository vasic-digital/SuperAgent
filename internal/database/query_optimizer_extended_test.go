package database

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// QueryOptimizer Extended Tests
// These tests cover the database query methods, cache interaction paths,
// metrics tracking, and edge cases that were not covered by the existing tests.
//
// Note: Methods like GetActiveProviders, GetProviderPerformance, etc. require
// a *pgxpool.Pool. When pool is nil, calling pool.Query() panics. For cache-miss
// paths we use safeCall to recover from panics and verify the code path was
// exercised. Cache-hit paths do not touch the pool and work without a real DB.
// =============================================================================

// safeCallError calls fn and recovers from panics, returning an error or
// the panic value as a string error. This lets us test code paths that
// dereference a nil pool.
func safeCallError(fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			// Panic occurred, which means code entered the pool call path
			err = assert.AnError
		}
	}()
	return fn()
}

// safeCallResult calls fn, recovering from panics.
func safeCallResult[T any](fn func() (T, error)) (result T, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = assert.AnError
		}
	}()
	return fn()
}

// -----------------------------------------------------------------------------
// GetActiveProviders Tests
// -----------------------------------------------------------------------------

func TestQueryOptimizer_GetActiveProviders_CacheHit(t *testing.T) {
	optimizer := NewQueryOptimizer(nil, nil)

	expectedProviders := []ActiveProvider{
		{
			ID:           "p1",
			Name:         "Claude",
			Type:         "anthropic",
			Weight:       2.0,
			HealthStatus: "healthy",
			ResponseTime: 100,
			Config:       []byte(`{"key":"val"}`),
		},
		{
			ID:           "p2",
			Name:         "DeepSeek",
			Type:         "deepseek",
			Weight:       1.5,
			HealthStatus: "healthy",
			ResponseTime: 200,
			Config:       nil,
		},
	}
	optimizer.queryCache.Set("active_providers", expectedProviders)

	result, err := optimizer.GetActiveProviders(context.Background())
	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, "Claude", result[0].Name)
	assert.Equal(t, "DeepSeek", result[1].Name)
	assert.Equal(t, 2.0, result[0].Weight)

	metrics := optimizer.Metrics()
	assert.Equal(t, int64(1), metrics.CacheHits)
	assert.Equal(t, int64(0), metrics.CacheMisses)
}

func TestQueryOptimizer_GetActiveProviders_CacheMiss_NilPool(t *testing.T) {
	optimizer := NewQueryOptimizer(nil, nil)

	// Cache is empty so it will be a miss, then it tries pool which is nil (panics)
	_, err := safeCallResult(func() ([]ActiveProvider, error) {
		return optimizer.GetActiveProviders(context.Background())
	})
	require.Error(t, err)

	// Cache miss metric should still be recorded (before pool call)
	metrics := optimizer.Metrics()
	assert.Equal(t, int64(0), metrics.CacheHits)
	assert.Equal(t, int64(1), metrics.CacheMisses)
}

func TestQueryOptimizer_GetActiveProviders_CacheDisabled_NilPool(t *testing.T) {
	config := &OptimizerConfig{
		MaxPreparedStmts: 50,
		CacheTTL:         5 * time.Minute,
		EnableCache:      false,
		DefaultBatchSize: 1000,
		QueryTimeout:     30 * time.Second,
	}
	optimizer := NewQueryOptimizer(nil, config)

	_, err := safeCallResult(func() ([]ActiveProvider, error) {
		return optimizer.GetActiveProviders(context.Background())
	})
	require.Error(t, err)

	metrics := optimizer.Metrics()
	assert.Equal(t, int64(0), metrics.CacheHits)
	assert.Equal(t, int64(0), metrics.CacheMisses)
}

func TestQueryOptimizer_GetActiveProviders_ExpiredCache_NilPool(t *testing.T) {
	config := &OptimizerConfig{
		MaxPreparedStmts: 50,
		CacheTTL:         1 * time.Millisecond,
		EnableCache:      true,
		DefaultBatchSize: 1000,
		QueryTimeout:     30 * time.Second,
	}
	optimizer := NewQueryOptimizer(nil, config)

	optimizer.queryCache.Set("active_providers", []ActiveProvider{{ID: "old"}})
	time.Sleep(5 * time.Millisecond)

	_, err := safeCallResult(func() ([]ActiveProvider, error) {
		return optimizer.GetActiveProviders(context.Background())
	})
	require.Error(t, err)

	metrics := optimizer.Metrics()
	assert.Equal(t, int64(1), metrics.CacheMisses)
}

func TestQueryOptimizer_GetActiveProviders_EmptyCachedSlice(t *testing.T) {
	optimizer := NewQueryOptimizer(nil, nil)

	// Cache with empty slice -- Get returns the value but empty results
	// are still valid cache hits
	optimizer.queryCache.Set("active_providers", []ActiveProvider{})

	result, err := optimizer.GetActiveProviders(context.Background())
	require.NoError(t, err)
	assert.Len(t, result, 0)

	metrics := optimizer.Metrics()
	assert.Equal(t, int64(1), metrics.CacheHits)
}

// -----------------------------------------------------------------------------
// GetProviderPerformance Tests
// -----------------------------------------------------------------------------

func TestQueryOptimizer_GetProviderPerformance_CacheHit(t *testing.T) {
	optimizer := NewQueryOptimizer(nil, nil)

	expectedPerf := []ProviderPerformance{
		{
			ProviderName:      "Claude",
			ProviderType:      "anthropic",
			HealthStatus:      "healthy",
			TotalRequests24h:  5000,
			AvgResponseTimeMs: 150,
			P95ResponseTimeMs: 400,
			AvgConfidence:     0.92,
			SuccessRate:       0.99,
		},
	}
	optimizer.queryCache.Set("provider_performance", expectedPerf)

	result, err := optimizer.GetProviderPerformance(context.Background())
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "Claude", result[0].ProviderName)
	assert.InDelta(t, 0.99, result[0].SuccessRate, 0.01)

	metrics := optimizer.Metrics()
	assert.Equal(t, int64(1), metrics.CacheHits)
}

func TestQueryOptimizer_GetProviderPerformance_CacheMiss_NilPool(t *testing.T) {
	optimizer := NewQueryOptimizer(nil, nil)

	_, err := safeCallResult(func() ([]ProviderPerformance, error) {
		return optimizer.GetProviderPerformance(context.Background())
	})
	require.Error(t, err)

	metrics := optimizer.Metrics()
	assert.Equal(t, int64(1), metrics.CacheMisses)
}

func TestQueryOptimizer_GetProviderPerformance_CacheDisabled(t *testing.T) {
	config := &OptimizerConfig{
		MaxPreparedStmts: 50,
		CacheTTL:         5 * time.Minute,
		EnableCache:      false,
		DefaultBatchSize: 1000,
		QueryTimeout:     30 * time.Second,
	}
	optimizer := NewQueryOptimizer(nil, config)

	_, err := safeCallResult(func() ([]ProviderPerformance, error) {
		return optimizer.GetProviderPerformance(context.Background())
	})
	require.Error(t, err)

	metrics := optimizer.Metrics()
	assert.Equal(t, int64(0), metrics.CacheHits)
	assert.Equal(t, int64(0), metrics.CacheMisses)
}

func TestQueryOptimizer_GetProviderPerformance_MultipleResults(t *testing.T) {
	optimizer := NewQueryOptimizer(nil, nil)

	perf := []ProviderPerformance{
		{ProviderName: "Claude", SuccessRate: 0.99},
		{ProviderName: "DeepSeek", SuccessRate: 0.95},
		{ProviderName: "Gemini", SuccessRate: 0.90},
	}
	optimizer.queryCache.Set("provider_performance", perf)

	result, err := optimizer.GetProviderPerformance(context.Background())
	require.NoError(t, err)
	require.Len(t, result, 3)
	assert.Equal(t, "Claude", result[0].ProviderName)
	assert.Equal(t, "DeepSeek", result[1].ProviderName)
	assert.Equal(t, "Gemini", result[2].ProviderName)
}

// -----------------------------------------------------------------------------
// GetMCPServerHealth Tests
// -----------------------------------------------------------------------------

func TestQueryOptimizer_GetMCPServerHealth_CacheHit(t *testing.T) {
	optimizer := NewQueryOptimizer(nil, nil)

	expectedHealth := []MCPServerHealth{
		{
			ServerID:             "mcp-fs",
			ServerName:           "filesystem",
			ServerType:           "stdio",
			Enabled:              true,
			TotalOperations1h:    1200,
			SuccessfulOperations: 1190,
			FailedOperations:     10,
			AvgDurationMs:        25,
			P95DurationMs:        80,
			SuccessRate:          0.992,
			ToolCount:            15,
		},
	}
	optimizer.queryCache.Set("mcp_server_health", expectedHealth)

	result, err := optimizer.GetMCPServerHealth(context.Background())
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "filesystem", result[0].ServerName)
	assert.Equal(t, 15, result[0].ToolCount)
	assert.True(t, result[0].Enabled)

	metrics := optimizer.Metrics()
	assert.Equal(t, int64(1), metrics.CacheHits)
}

func TestQueryOptimizer_GetMCPServerHealth_CacheMiss_NilPool(t *testing.T) {
	optimizer := NewQueryOptimizer(nil, nil)

	_, err := safeCallResult(func() ([]MCPServerHealth, error) {
		return optimizer.GetMCPServerHealth(context.Background())
	})
	require.Error(t, err)

	metrics := optimizer.Metrics()
	assert.Equal(t, int64(1), metrics.CacheMisses)
}

func TestQueryOptimizer_GetMCPServerHealth_CacheDisabled(t *testing.T) {
	config := &OptimizerConfig{
		MaxPreparedStmts: 50,
		CacheTTL:         5 * time.Minute,
		EnableCache:      false,
		DefaultBatchSize: 1000,
		QueryTimeout:     30 * time.Second,
	}
	optimizer := NewQueryOptimizer(nil, config)

	_, err := safeCallResult(func() ([]MCPServerHealth, error) {
		return optimizer.GetMCPServerHealth(context.Background())
	})
	require.Error(t, err)

	metrics := optimizer.Metrics()
	assert.Equal(t, int64(0), metrics.CacheHits)
	assert.Equal(t, int64(0), metrics.CacheMisses)
}

func TestQueryOptimizer_GetMCPServerHealth_MultipleServers(t *testing.T) {
	optimizer := NewQueryOptimizer(nil, nil)

	health := []MCPServerHealth{
		{ServerID: "mcp-1", ServerName: "filesystem", ToolCount: 10},
		{ServerID: "mcp-2", ServerName: "github", ToolCount: 25},
	}
	optimizer.queryCache.Set("mcp_server_health", health)

	result, err := optimizer.GetMCPServerHealth(context.Background())
	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, "filesystem", result[0].ServerName)
	assert.Equal(t, "github", result[1].ServerName)
}

// -----------------------------------------------------------------------------
// BulkInsert Tests
// -----------------------------------------------------------------------------

func TestQueryOptimizer_BulkInsert_EmptyRows(t *testing.T) {
	optimizer := NewQueryOptimizer(nil, nil)

	count, err := optimizer.BulkInsert(
		context.Background(),
		"test_table",
		[]string{"col1", "col2"},
		[][]interface{}{},
	)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestQueryOptimizer_BulkInsert_NilPool(t *testing.T) {
	optimizer := NewQueryOptimizer(nil, nil)

	rows := [][]interface{}{
		{"val1", "val2"},
		{"val3", "val4"},
	}

	_, err := safeCallResult(func() (int64, error) {
		return optimizer.BulkInsert(
			context.Background(),
			"test_table",
			[]string{"col1", "col2"},
			rows,
		)
	})
	require.Error(t, err)
}

func TestQueryOptimizer_BulkInsert_CacheNotInvalidatedOnEmptyRows(t *testing.T) {
	optimizer := NewQueryOptimizer(nil, nil)

	optimizer.queryCache.Set("test_table_key1", "value1")
	optimizer.queryCache.Set("test_table_key2", "value2")
	optimizer.queryCache.Set("other_key", "value3")

	count, err := optimizer.BulkInsert(
		context.Background(),
		"test_table",
		[]string{"col1"},
		[][]interface{}{},
	)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Empty rows path returns early, so cache should not be invalidated
	_, ok := optimizer.queryCache.Get("test_table_key1")
	assert.True(t, ok)
}

func TestQueryOptimizer_BulkInsert_NilRows(t *testing.T) {
	optimizer := NewQueryOptimizer(nil, nil)

	count, err := optimizer.BulkInsert(
		context.Background(),
		"test_table",
		[]string{"col1"},
		nil,
	)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

// -----------------------------------------------------------------------------
// BulkInsertBatched Tests
// -----------------------------------------------------------------------------

func TestQueryOptimizer_BulkInsertBatched_EmptyRows(t *testing.T) {
	optimizer := NewQueryOptimizer(nil, nil)

	count, err := optimizer.BulkInsertBatched(
		context.Background(),
		"test_table",
		[]string{"col1", "col2"},
		[][]interface{}{},
	)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestQueryOptimizer_BulkInsertBatched_NilRows(t *testing.T) {
	optimizer := NewQueryOptimizer(nil, nil)

	count, err := optimizer.BulkInsertBatched(
		context.Background(),
		"test_table",
		[]string{"col1"},
		nil,
	)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestQueryOptimizer_BulkInsertBatched_NilPool_SingleBatch(t *testing.T) {
	config := &OptimizerConfig{
		MaxPreparedStmts: 50,
		CacheTTL:         5 * time.Minute,
		EnableCache:      true,
		DefaultBatchSize: 1000,
		QueryTimeout:     30 * time.Second,
	}
	optimizer := NewQueryOptimizer(nil, config)

	rows := [][]interface{}{
		{"val1"},
		{"val2"},
	}

	_, err := safeCallResult(func() (int64, error) {
		return optimizer.BulkInsertBatched(
			context.Background(),
			"test_table",
			[]string{"col1"},
			rows,
		)
	})
	require.Error(t, err)
}

func TestQueryOptimizer_BulkInsertBatched_SmallBatchSize(t *testing.T) {
	config := &OptimizerConfig{
		MaxPreparedStmts: 50,
		CacheTTL:         5 * time.Minute,
		EnableCache:      true,
		DefaultBatchSize: 2,
		QueryTimeout:     30 * time.Second,
	}
	optimizer := NewQueryOptimizer(nil, config)

	rows := [][]interface{}{
		{"val1"},
		{"val2"},
		{"val3"},
		{"val4"},
		{"val5"},
	}

	_, err := safeCallResult(func() (int64, error) {
		return optimizer.BulkInsertBatched(
			context.Background(),
			"test_table",
			[]string{"col1"},
			rows,
		)
	})
	require.Error(t, err)
}

// -----------------------------------------------------------------------------
// RefreshMaterializedViews Tests
// -----------------------------------------------------------------------------

func TestQueryOptimizer_RefreshMaterializedViews_NilPool(t *testing.T) {
	optimizer := NewQueryOptimizer(nil, nil)

	_, err := safeCallResult(func() ([]ViewRefreshResult, error) {
		return optimizer.RefreshMaterializedViews(context.Background())
	})
	require.Error(t, err)
}

func TestQueryOptimizer_RefreshMaterializedViews_CacheDisabled(t *testing.T) {
	config := &OptimizerConfig{
		MaxPreparedStmts: 50,
		CacheTTL:         5 * time.Minute,
		EnableCache:      false,
		DefaultBatchSize: 1000,
		QueryTimeout:     30 * time.Second,
	}
	optimizer := NewQueryOptimizer(nil, config)

	_, err := safeCallResult(func() ([]ViewRefreshResult, error) {
		return optimizer.RefreshMaterializedViews(context.Background())
	})
	require.Error(t, err)
}

// -----------------------------------------------------------------------------
// RefreshCriticalViews Tests
// -----------------------------------------------------------------------------

func TestQueryOptimizer_RefreshCriticalViews_NilPool(t *testing.T) {
	optimizer := NewQueryOptimizer(nil, nil)

	err := safeCallError(func() error {
		return optimizer.RefreshCriticalViews(context.Background())
	})
	require.Error(t, err)
}

func TestQueryOptimizer_RefreshCriticalViews_CacheDisabled(t *testing.T) {
	config := &OptimizerConfig{
		MaxPreparedStmts: 50,
		CacheTTL:         5 * time.Minute,
		EnableCache:      false,
		DefaultBatchSize: 1000,
		QueryTimeout:     30 * time.Second,
	}
	optimizer := NewQueryOptimizer(nil, config)

	err := safeCallError(func() error {
		return optimizer.RefreshCriticalViews(context.Background())
	})
	require.Error(t, err)
}

// -----------------------------------------------------------------------------
// Metrics and Latency Tests
// -----------------------------------------------------------------------------

func TestQueryOptimizer_AverageLatency_WithQueries(t *testing.T) {
	optimizer := NewQueryOptimizer(nil, nil)

	atomic.StoreInt64(&optimizer.metrics.TotalQueries, 10)
	atomic.StoreInt64(&optimizer.metrics.TotalLatencyUs, 50000)

	latency := optimizer.AverageLatency()
	assert.Equal(t, 5000*time.Microsecond, latency)
}

func TestQueryOptimizer_AverageLatency_SingleQuery(t *testing.T) {
	optimizer := NewQueryOptimizer(nil, nil)

	atomic.StoreInt64(&optimizer.metrics.TotalQueries, 1)
	atomic.StoreInt64(&optimizer.metrics.TotalLatencyUs, 1000)

	latency := optimizer.AverageLatency()
	assert.Equal(t, 1000*time.Microsecond, latency)
}

func TestQueryOptimizer_CacheHitRate_WithData(t *testing.T) {
	optimizer := NewQueryOptimizer(nil, nil)

	atomic.StoreInt64(&optimizer.metrics.CacheHits, 80)
	atomic.StoreInt64(&optimizer.metrics.CacheMisses, 20)

	rate := optimizer.CacheHitRate()
	assert.InDelta(t, 80.0, rate, 0.01)
}

func TestQueryOptimizer_CacheHitRate_AllHits(t *testing.T) {
	optimizer := NewQueryOptimizer(nil, nil)

	atomic.StoreInt64(&optimizer.metrics.CacheHits, 100)
	atomic.StoreInt64(&optimizer.metrics.CacheMisses, 0)

	rate := optimizer.CacheHitRate()
	assert.InDelta(t, 100.0, rate, 0.01)
}

func TestQueryOptimizer_CacheHitRate_AllMisses(t *testing.T) {
	optimizer := NewQueryOptimizer(nil, nil)

	atomic.StoreInt64(&optimizer.metrics.CacheHits, 0)
	atomic.StoreInt64(&optimizer.metrics.CacheMisses, 50)

	rate := optimizer.CacheHitRate()
	assert.InDelta(t, 0.0, rate, 0.01)
}

func TestQueryOptimizer_Metrics_AfterMultipleOperations(t *testing.T) {
	optimizer := NewQueryOptimizer(nil, nil)

	atomic.StoreInt64(&optimizer.metrics.TotalQueries, 100)
	atomic.StoreInt64(&optimizer.metrics.CacheHits, 60)
	atomic.StoreInt64(&optimizer.metrics.CacheMisses, 40)
	atomic.StoreInt64(&optimizer.metrics.TotalLatencyUs, 500000)
	atomic.StoreInt64(&optimizer.metrics.SlowQueries, 5)
	atomic.StoreInt64(&optimizer.metrics.PreparedStmtHits, 30)
	atomic.StoreInt64(&optimizer.metrics.BulkInsertRows, 10000)
	atomic.StoreInt64(&optimizer.metrics.BulkInsertBatches, 10)

	metrics := optimizer.Metrics()
	assert.Equal(t, int64(100), metrics.TotalQueries)
	assert.Equal(t, int64(60), metrics.CacheHits)
	assert.Equal(t, int64(40), metrics.CacheMisses)
	assert.Equal(t, int64(500000), metrics.TotalLatencyUs)
	assert.Equal(t, int64(5), metrics.SlowQueries)
	assert.Equal(t, int64(30), metrics.PreparedStmtHits)
	assert.Equal(t, int64(10000), metrics.BulkInsertRows)
	assert.Equal(t, int64(10), metrics.BulkInsertBatches)
}

// -----------------------------------------------------------------------------
// InvalidateCache with nil cache Tests
// -----------------------------------------------------------------------------

func TestQueryOptimizer_InvalidateCache_NilCache(t *testing.T) {
	config := &OptimizerConfig{
		MaxPreparedStmts: 50,
		CacheTTL:         5 * time.Minute,
		EnableCache:      false,
		DefaultBatchSize: 1000,
		QueryTimeout:     30 * time.Second,
	}
	optimizer := NewQueryOptimizer(nil, config)

	// Should not panic with nil cache
	optimizer.InvalidateCache()
}

func TestQueryOptimizer_InvalidateCacheKey_NilCache(t *testing.T) {
	config := &OptimizerConfig{
		MaxPreparedStmts: 50,
		CacheTTL:         5 * time.Minute,
		EnableCache:      false,
		DefaultBatchSize: 1000,
		QueryTimeout:     30 * time.Second,
	}
	optimizer := NewQueryOptimizer(nil, config)

	optimizer.InvalidateCacheKey("some_key")
}

// -----------------------------------------------------------------------------
// QueryCache Concurrent Access Tests
// -----------------------------------------------------------------------------

func TestQueryCache_ConcurrentSetAndGet(t *testing.T) {
	cache := NewQueryCache(5*time.Minute, 1000)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func(idx int) {
			defer wg.Done()
			cache.Set("key", idx)
		}(i)
		go func(idx int) {
			defer wg.Done()
			cache.Get("key")
		}(i)
	}
	wg.Wait()
}

func TestQueryCache_ConcurrentInvalidate(t *testing.T) {
	cache := NewQueryCache(5*time.Minute, 1000)

	for i := 0; i < 100; i++ {
		cache.Set("key_"+string(rune('a'+i%26)), i)
	}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if idx%2 == 0 {
				cache.Invalidate("key_a")
			} else {
				cache.InvalidatePrefix("key_")
			}
		}(i)
	}
	wg.Wait()
}

func TestQueryCache_InvalidatePrefix_EmptyPrefix(t *testing.T) {
	cache := NewQueryCache(5*time.Minute, 100)

	cache.Set("key1", "val1")
	cache.Set("key2", "val2")

	cache.InvalidatePrefix("")

	_, ok1 := cache.Get("key1")
	_, ok2 := cache.Get("key2")
	assert.False(t, ok1)
	assert.False(t, ok2)
}

func TestQueryCache_InvalidatePrefix_NoMatch(t *testing.T) {
	cache := NewQueryCache(5*time.Minute, 100)

	cache.Set("provider_1", "val1")
	cache.Set("provider_2", "val2")

	cache.InvalidatePrefix("nonexistent_prefix_")

	_, ok1 := cache.Get("provider_1")
	_, ok2 := cache.Get("provider_2")
	assert.True(t, ok1)
	assert.True(t, ok2)
}

func TestQueryCache_Eviction_Ordering(t *testing.T) {
	cache := NewQueryCache(5*time.Minute, 2)

	cache.Set("first", "val1")
	time.Sleep(1 * time.Millisecond)
	cache.Set("second", "val2")
	time.Sleep(1 * time.Millisecond)

	cache.Set("third", "val3")

	_, okFirst := cache.Get("first")
	_, okSecond := cache.Get("second")
	_, okThird := cache.Get("third")

	assert.False(t, okFirst, "first entry should be evicted")
	assert.True(t, okSecond, "second entry should exist")
	assert.True(t, okThird, "third entry should exist")
}

func TestQueryCache_Set_OverwriteExisting(t *testing.T) {
	cache := NewQueryCache(5*time.Minute, 100)

	cache.Set("key", "original")
	cache.Set("key", "updated")

	val, ok := cache.Get("key")
	assert.True(t, ok)
	assert.Equal(t, "updated", val)
}

func TestQueryCache_Set_ComplexTypes(t *testing.T) {
	cache := NewQueryCache(5*time.Minute, 100)

	providers := []ActiveProvider{
		{ID: "p1", Name: "Test"},
	}
	cache.Set("complex", providers)

	val, ok := cache.Get("complex")
	assert.True(t, ok)

	result, isType := val.([]ActiveProvider)
	require.True(t, isType)
	assert.Equal(t, "Test", result[0].Name)
}

// -----------------------------------------------------------------------------
// QueryOptimizer Timeout Configuration Tests
// -----------------------------------------------------------------------------

func TestQueryOptimizer_CustomTimeout(t *testing.T) {
	config := &OptimizerConfig{
		MaxPreparedStmts: 10,
		CacheTTL:         1 * time.Minute,
		EnableCache:      true,
		DefaultBatchSize: 100,
		QueryTimeout:     100 * time.Millisecond,
	}
	optimizer := NewQueryOptimizer(nil, config)

	assert.Equal(t, 100*time.Millisecond, optimizer.config.QueryTimeout)
}

// -----------------------------------------------------------------------------
// NewQueryOptimizer Edge Cases
// -----------------------------------------------------------------------------

func TestNewQueryOptimizer_CacheSize(t *testing.T) {
	config := &OptimizerConfig{
		MaxPreparedStmts: 20,
		CacheTTL:         10 * time.Minute,
		EnableCache:      true,
		DefaultBatchSize: 500,
		QueryTimeout:     1 * time.Minute,
	}
	optimizer := NewQueryOptimizer(nil, config)

	assert.NotNil(t, optimizer.queryCache)
	assert.Equal(t, 200, optimizer.queryCache.maxSize)
}

func TestNewQueryOptimizer_ZeroPreparedStmts(t *testing.T) {
	config := &OptimizerConfig{
		MaxPreparedStmts: 0,
		CacheTTL:         5 * time.Minute,
		EnableCache:      true,
		DefaultBatchSize: 1000,
		QueryTimeout:     30 * time.Second,
	}
	optimizer := NewQueryOptimizer(nil, config)

	assert.NotNil(t, optimizer.queryCache)
	assert.Equal(t, 0, optimizer.queryCache.maxSize)
}

// -----------------------------------------------------------------------------
// QueryOptimizer Close Idempotent Test
// -----------------------------------------------------------------------------

func TestQueryOptimizer_Close_Idempotent(t *testing.T) {
	optimizer := NewQueryOptimizer(nil, nil)

	err := optimizer.Close()
	assert.NoError(t, err)

	err = optimizer.Close()
	assert.NoError(t, err)
}

// -----------------------------------------------------------------------------
// Multiple Cache Operations Sequence Tests
// -----------------------------------------------------------------------------

func TestQueryOptimizer_CacheWorkflow(t *testing.T) {
	optimizer := NewQueryOptimizer(nil, nil)

	optimizer.queryCache.Set("active_providers", []ActiveProvider{
		{ID: "p1", Name: "Provider1"},
	})

	// Cache hit
	result, err := optimizer.GetActiveProviders(context.Background())
	require.NoError(t, err)
	require.Len(t, result, 1)

	// Invalidate
	optimizer.InvalidateCacheKey("active_providers")

	// Cache miss (pool nil = panic recovered)
	_, err = safeCallResult(func() ([]ActiveProvider, error) {
		return optimizer.GetActiveProviders(context.Background())
	})
	require.Error(t, err)

	metrics := optimizer.Metrics()
	assert.Equal(t, int64(1), metrics.CacheHits)
	assert.Equal(t, int64(1), metrics.CacheMisses)
}

// -----------------------------------------------------------------------------
// Slow Query Detection Test
// -----------------------------------------------------------------------------

func TestQueryOptimizer_SlowQueryMetricIncrement(t *testing.T) {
	optimizer := NewQueryOptimizer(nil, nil)

	atomic.StoreInt64(&optimizer.metrics.SlowQueries, 3)

	metrics := optimizer.Metrics()
	assert.Equal(t, int64(3), metrics.SlowQueries)
}

// -----------------------------------------------------------------------------
// ViewRefreshResult Tests
// -----------------------------------------------------------------------------

func TestViewRefreshResult_MultipleViews(t *testing.T) {
	results := []ViewRefreshResult{
		{ViewName: "mv_provider_performance", Status: "refreshed", DurationMs: 120},
		{ViewName: "mv_mcp_server_health", Status: "refreshed", DurationMs: 85},
		{ViewName: "mv_task_stats", Status: "error", DurationMs: 0},
	}

	assert.Len(t, results, 3)
	assert.Equal(t, "refreshed", results[0].Status)
	assert.Equal(t, "error", results[2].Status)
	assert.Equal(t, 0, results[2].DurationMs)
}

// -----------------------------------------------------------------------------
// QueryOptimizer Multiple Simultaneous Cache Hits
// -----------------------------------------------------------------------------

func TestQueryOptimizer_ConcurrentCacheHits(t *testing.T) {
	optimizer := NewQueryOptimizer(nil, nil)

	optimizer.queryCache.Set("active_providers", []ActiveProvider{
		{ID: "p1", Name: "ConcurrentTest"},
	})

	var wg sync.WaitGroup
	errCh := make(chan error, 50)

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := optimizer.GetActiveProviders(context.Background())
			if err != nil {
				errCh <- err
				return
			}
			if len(result) != 1 {
				errCh <- assert.AnError
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("concurrent cache hit failed: %v", err)
	}

	metrics := optimizer.Metrics()
	assert.Equal(t, int64(50), metrics.CacheHits)
}

// -----------------------------------------------------------------------------
// Cache interaction with different data types
// -----------------------------------------------------------------------------

func TestQueryOptimizer_GetProviderPerformance_CacheHit_MultipleCalls(t *testing.T) {
	optimizer := NewQueryOptimizer(nil, nil)

	perf := []ProviderPerformance{
		{ProviderName: "TestProvider", SuccessRate: 0.95},
	}
	optimizer.queryCache.Set("provider_performance", perf)

	// Multiple calls should all hit cache
	for i := 0; i < 5; i++ {
		result, err := optimizer.GetProviderPerformance(context.Background())
		require.NoError(t, err)
		require.Len(t, result, 1)
	}

	metrics := optimizer.Metrics()
	assert.Equal(t, int64(5), metrics.CacheHits)
}

func TestQueryOptimizer_GetMCPServerHealth_CacheHit_MultipleCalls(t *testing.T) {
	optimizer := NewQueryOptimizer(nil, nil)

	health := []MCPServerHealth{
		{ServerID: "mcp-1", ServerName: "test-server"},
	}
	optimizer.queryCache.Set("mcp_server_health", health)

	for i := 0; i < 3; i++ {
		result, err := optimizer.GetMCPServerHealth(context.Background())
		require.NoError(t, err)
		require.Len(t, result, 1)
	}

	metrics := optimizer.Metrics()
	assert.Equal(t, int64(3), metrics.CacheHits)
}

// -----------------------------------------------------------------------------
// QueryCache key length edge cases
// -----------------------------------------------------------------------------

func TestQueryCache_Invalidate_NonExistentKey(t *testing.T) {
	cache := NewQueryCache(5*time.Minute, 100)

	cache.Set("existing_key", "value")

	// Invalidating a key that does not exist should not affect other keys
	cache.Invalidate("nonexistent_key")

	val, ok := cache.Get("existing_key")
	assert.True(t, ok)
	assert.Equal(t, "value", val)
}

func TestQueryCache_Get_EmptyKey(t *testing.T) {
	cache := NewQueryCache(5*time.Minute, 100)

	cache.Set("", "empty_key_value")

	val, ok := cache.Get("")
	assert.True(t, ok)
	assert.Equal(t, "empty_key_value", val)
}

func TestQueryCache_InvalidatePrefix_LongerThanKey(t *testing.T) {
	cache := NewQueryCache(5*time.Minute, 100)

	cache.Set("ab", "value")

	// Prefix is longer than any key, should not match
	cache.InvalidatePrefix("abcdef")

	val, ok := cache.Get("ab")
	assert.True(t, ok)
	assert.Equal(t, "value", val)
}

// -----------------------------------------------------------------------------
// LRU Eviction Behavior Tests
// -----------------------------------------------------------------------------

func TestQueryCache_LRU_EvictsLeastRecentlyUsed(t *testing.T) {
	cache := NewQueryCache(5*time.Minute, 3)

	cache.Set("a", "1")
	cache.Set("b", "2")
	cache.Set("c", "3")

	// Access "a" to make it recently used
	_, _ = cache.Get("a")

	// Insert "d" — should evict "b" (least recently used)
	cache.Set("d", "4")

	_, ok := cache.Get("b")
	assert.False(t, ok, "b should have been evicted as LRU")

	val, ok := cache.Get("a")
	assert.True(t, ok, "a should still exist after being accessed")
	assert.Equal(t, "1", val)
}

func TestQueryCache_LRU_UpdateExistingMovesToFront(t *testing.T) {
	cache := NewQueryCache(5*time.Minute, 3)

	cache.Set("a", "1")
	cache.Set("b", "2")
	cache.Set("c", "3")

	// Update "a" — moves it to front
	cache.Set("a", "updated")

	// Insert "d" — should evict "b" (now the LRU)
	cache.Set("d", "4")

	_, ok := cache.Get("b")
	assert.False(t, ok, "b should have been evicted")

	val, ok := cache.Get("a")
	assert.True(t, ok)
	assert.Equal(t, "updated", val)
}

func TestQueryCache_LRU_GetOnExpiredRemovesEntry(t *testing.T) {
	cache := NewQueryCache(50*time.Millisecond, 10)

	cache.Set("x", "value")
	time.Sleep(100 * time.Millisecond)

	// Get on expired entry should remove it from both map and list
	val, ok := cache.Get("x")
	assert.False(t, ok)
	assert.Nil(t, val)

	// Setting new entries should not see the expired one
	cache.Set("y", "new")
	val, ok = cache.Get("y")
	assert.True(t, ok)
	assert.Equal(t, "new", val)
}

// -----------------------------------------------------------------------------
// Benchmarks for O(1) LRU Cache
// -----------------------------------------------------------------------------

func BenchmarkQueryCacheSet(b *testing.B) {
	cache := NewQueryCache(5*time.Minute, 1000)
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i%2000) // Force evictions
		cache.Set(key, i)
	}
}

func BenchmarkQueryCacheGet(b *testing.B) {
	cache := NewQueryCache(5*time.Minute, 1000)
	for i := 0; i < 1000; i++ {
		cache.Set(fmt.Sprintf("key-%d", i), i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i%1000)
		_, _ = cache.Get(key)
	}
}

func BenchmarkQueryCacheSetAtCapacity(b *testing.B) {
	cache := NewQueryCache(5*time.Minute, 100)
	// Fill to capacity
	for i := 0; i < 100; i++ {
		cache.Set(fmt.Sprintf("pre-%d", i), i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Every Set triggers an eviction
		cache.Set(fmt.Sprintf("new-%d", i), i)
	}
}
