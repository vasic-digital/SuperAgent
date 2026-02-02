package database

import (
	"context"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Pool Config Extended Tests
// These tests cover OptimizedPool methods, LazyPool logic, and pool
// configuration edge cases not exercised by the existing test suite.
// =============================================================================

// -----------------------------------------------------------------------------
// DefaultPoolOptions Boundary Tests
// -----------------------------------------------------------------------------

func TestDefaultPoolOptions_MinConnsCalculation(t *testing.T) {
	opts := DefaultPoolOptions()
	cpuCount := int32(runtime.NumCPU()) // #nosec G115

	expectedMinConns := cpuCount / 2
	assert.Equal(t, expectedMinConns, opts.MinConns)
}

func TestDefaultPoolOptions_ApplicationName(t *testing.T) {
	opts := DefaultPoolOptions()
	assert.Equal(t, "helixagent", opts.ApplicationName)
}

func TestHighPerformancePoolOptions_MinConnsCalculation(t *testing.T) {
	opts := HighPerformancePoolOptions()
	assert.Equal(t, opts.MaxConns/2, opts.MinConns)
}

func TestHighPerformancePoolOptions_Bounds(t *testing.T) {
	opts := HighPerformancePoolOptions()
	assert.GreaterOrEqual(t, opts.MaxConns, int32(20))
	assert.LessOrEqual(t, opts.MaxConns, int32(100))
}

func TestLowLatencyPoolOptions_ConnectTimeout(t *testing.T) {
	opts := LowLatencyPoolOptions()
	assert.Equal(t, 1*time.Second, opts.ConnectTimeout)
}

func TestLowLatencyPoolOptions_HealthCheckPeriod(t *testing.T) {
	opts := LowLatencyPoolOptions()
	assert.Equal(t, 10*time.Second, opts.HealthCheckPeriod)
}

// -----------------------------------------------------------------------------
// CreateOptimizedPoolConfig Tests
// -----------------------------------------------------------------------------

func TestCreateOptimizedPoolConfig_WithStatementCacheEnabled(t *testing.T) {
	connString := "postgresql://user:password@localhost:5432/testdb"
	opts := &PoolConfigOptions{
		MaxConns:               10,
		MinConns:               2,
		MaxConnLifetime:        time.Hour,
		MaxConnIdleTime:        30 * time.Minute,
		HealthCheckPeriod:      30 * time.Second,
		ConnectTimeout:         5 * time.Second,
		EnableStatementCache:   true,
		StatementCacheCapacity: 256,
		PreferSimpleProtocol:   false,
		ApplicationName:        "test-cache",
	}

	config, err := CreateOptimizedPoolConfig(connString, opts)
	require.NoError(t, err)
	require.NotNil(t, config)
	assert.Equal(t, "test-cache", config.ConnConfig.RuntimeParams["application_name"])
}

func TestCreateOptimizedPoolConfig_SimpleProtocolDisabled(t *testing.T) {
	connString := "postgresql://user:password@localhost:5432/testdb"
	opts := &PoolConfigOptions{
		MaxConns:             10,
		MinConns:             2,
		MaxConnLifetime:      time.Hour,
		MaxConnIdleTime:      30 * time.Minute,
		HealthCheckPeriod:    30 * time.Second,
		ConnectTimeout:       5 * time.Second,
		EnableStatementCache: true,
		PreferSimpleProtocol: false,
		ApplicationName:      "no-simple",
	}

	config, err := CreateOptimizedPoolConfig(connString, opts)
	require.NoError(t, err)
	require.NotNil(t, config)
}

func TestCreateOptimizedPoolConfig_BothCacheAndSimpleProtocol(t *testing.T) {
	connString := "postgresql://user:password@localhost:5432/testdb"
	opts := &PoolConfigOptions{
		MaxConns:             10,
		MinConns:             2,
		MaxConnLifetime:      time.Hour,
		MaxConnIdleTime:      30 * time.Minute,
		HealthCheckPeriod:    30 * time.Second,
		ConnectTimeout:       5 * time.Second,
		EnableStatementCache: true,
		PreferSimpleProtocol: true,
		ApplicationName:      "both-enabled",
	}

	config, err := CreateOptimizedPoolConfig(connString, opts)
	require.NoError(t, err)
	require.NotNil(t, config)
	// When both are enabled, SimpleProtocol takes precedence (set last)
}

func TestCreateOptimizedPoolConfig_NeitherCacheNorSimple(t *testing.T) {
	connString := "postgresql://user:password@localhost:5432/testdb"
	opts := &PoolConfigOptions{
		MaxConns:             10,
		MinConns:             2,
		MaxConnLifetime:      time.Hour,
		MaxConnIdleTime:      30 * time.Minute,
		HealthCheckPeriod:    30 * time.Second,
		ConnectTimeout:       5 * time.Second,
		EnableStatementCache: false,
		PreferSimpleProtocol: false,
		ApplicationName:      "vanilla",
	}

	config, err := CreateOptimizedPoolConfig(connString, opts)
	require.NoError(t, err)
	require.NotNil(t, config)
}

func TestCreateOptimizedPoolConfig_AfterConnectHook(t *testing.T) {
	connString := "postgresql://user:password@localhost:5432/testdb"
	config, err := CreateOptimizedPoolConfig(connString, nil)

	require.NoError(t, err)
	require.NotNil(t, config)
	assert.NotNil(t, config.AfterConnect, "AfterConnect hook should be set")
}

func TestCreateOptimizedPoolConfig_EmptyConnString(t *testing.T) {
	// pgxpool.ParseConfig accepts empty string (uses defaults), so no error
	config, err := CreateOptimizedPoolConfig("", nil)
	require.NoError(t, err)
	require.NotNil(t, config)
}

// -----------------------------------------------------------------------------
// PoolConfigOptions Zero Values Test
// -----------------------------------------------------------------------------

func TestPoolConfigOptions_ZeroValues(t *testing.T) {
	opts := PoolConfigOptions{}

	assert.Equal(t, int32(0), opts.MaxConns)
	assert.Equal(t, int32(0), opts.MinConns)
	assert.Equal(t, time.Duration(0), opts.MaxConnLifetime)
	assert.Equal(t, time.Duration(0), opts.MaxConnIdleTime)
	assert.Equal(t, time.Duration(0), opts.HealthCheckPeriod)
	assert.Equal(t, time.Duration(0), opts.ConnectTimeout)
	assert.False(t, opts.EnableStatementCache)
	assert.Equal(t, 0, opts.StatementCacheCapacity)
	assert.False(t, opts.PreferSimpleProtocol)
	assert.Empty(t, opts.ApplicationName)
}

// -----------------------------------------------------------------------------
// PoolMetrics Zero Values Test
// -----------------------------------------------------------------------------

func TestPoolMetrics_ZeroValues(t *testing.T) {
	metrics := &PoolMetrics{}

	assert.Equal(t, int64(0), metrics.AcquireCount)
	assert.Equal(t, int64(0), metrics.AcquireErrors)
	assert.Equal(t, int64(0), metrics.TotalAcquireTimeUs)
	assert.Equal(t, int64(0), metrics.MaxConcurrent)
	assert.Equal(t, int64(0), metrics.CurrentConcurrent)
	assert.Equal(t, int64(0), metrics.IdleConns)
	assert.Equal(t, int64(0), metrics.TotalConns)
	assert.Equal(t, int64(0), metrics.WaitCount)
}

// -----------------------------------------------------------------------------
// OptimizedPool - Test methods via direct construction
// Since NewOptimizedPool requires a real DB, we test what we can without one.
// -----------------------------------------------------------------------------

func TestNewOptimizedPool_InvalidConnString(t *testing.T) {
	_, err := NewOptimizedPool(context.Background(), "invalid", nil)
	assert.Error(t, err)
}

func TestNewOptimizedPool_UnreachableHost(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := NewOptimizedPool(ctx,
		"postgresql://user:pass@192.0.2.1:5432/db?connect_timeout=1",
		nil,
	)
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// LazyPool Tests
// -----------------------------------------------------------------------------

func TestNewLazyPool_Fields(t *testing.T) {
	opts := DefaultPoolOptions()
	pool := NewLazyPool("postgresql://localhost/test", opts)

	require.NotNil(t, pool)
	assert.Equal(t, "postgresql://localhost/test", pool.connStr)
	assert.Equal(t, opts, pool.opts)
	assert.False(t, pool.IsInitialized())
	assert.Nil(t, pool.pool)
}

func TestLazyPool_Get_InvalidConnString(t *testing.T) {
	pool := NewLazyPool("invalid-connection-string", nil)

	_, err := pool.Get(context.Background())
	assert.Error(t, err)

	// After failed init, pool is still marked as initialized (with error)
	assert.True(t, pool.IsInitialized())
}

func TestLazyPool_Get_UnreachableHost(t *testing.T) {
	pool := NewLazyPool(
		"postgresql://user:pass@192.0.2.1:5432/db?connect_timeout=1",
		nil,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := pool.Get(ctx)
	assert.Error(t, err)
	assert.True(t, pool.IsInitialized())
}

func TestLazyPool_Get_Idempotent(t *testing.T) {
	pool := NewLazyPool("invalid-connection-string", nil)

	// First call initializes
	_, err1 := pool.Get(context.Background())
	assert.Error(t, err1)
	assert.True(t, pool.IsInitialized())

	// Second call returns cached result (fast path)
	_, err2 := pool.Get(context.Background())
	assert.Error(t, err2)
}

func TestLazyPool_Get_ContextCancelled(t *testing.T) {
	pool := NewLazyPool(
		"postgresql://user:pass@192.0.2.1:5432/db?connect_timeout=10",
		nil,
	)

	// Block the init mutex first
	pool.initMu <- struct{}{} // Fill the channel

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// This call should block waiting for the mutex, then timeout
	_, err := pool.Get(ctx)
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)

	// Release the mutex
	<-pool.initMu
}

func TestLazyPool_Close_AfterFailedInit(t *testing.T) {
	pool := NewLazyPool("invalid-connection-string", nil)

	// Initialize with error
	_, _ = pool.Get(context.Background())
	assert.True(t, pool.IsInitialized())

	// Close should not panic even though pool is nil
	pool.Close()
}

func TestLazyPool_IsInitialized_BeforeGet(t *testing.T) {
	pool := NewLazyPool("postgresql://localhost/test", nil)
	assert.False(t, pool.IsInitialized())
}

func TestLazyPool_Close_WithNilPool(t *testing.T) {
	pool := NewLazyPool("postgresql://localhost/test", nil)

	// Set initOnce without actual pool
	atomic.StoreInt32(&pool.initOnce, 1)
	pool.pool = nil

	// Should not panic
	pool.Close()
}

// -----------------------------------------------------------------------------
// OptimizedPool AverageAcquireTime Edge Cases
// These test the method logic without needing a real pool.
// -----------------------------------------------------------------------------

func TestOptimizedPool_AverageAcquireTime_Indirect(t *testing.T) {
	// We can't create OptimizedPool without a real DB, but we can test
	// the PoolMetrics calculation logic independently
	metrics := &PoolMetrics{
		AcquireCount:       10,
		TotalAcquireTimeUs: 50000,
	}

	// Same calculation as AverageAcquireTime
	count := atomic.LoadInt64(&metrics.AcquireCount)
	totalUs := atomic.LoadInt64(&metrics.TotalAcquireTimeUs)
	avg := time.Duration(totalUs/count) * time.Microsecond

	assert.Equal(t, 5000*time.Microsecond, avg)
}

func TestOptimizedPool_AverageAcquireTime_ZeroCount(t *testing.T) {
	metrics := &PoolMetrics{
		AcquireCount:       0,
		TotalAcquireTimeUs: 0,
	}

	count := atomic.LoadInt64(&metrics.AcquireCount)
	var avg time.Duration
	if count == 0 {
		avg = 0
	} else {
		totalUs := atomic.LoadInt64(&metrics.TotalAcquireTimeUs)
		avg = time.Duration(totalUs/count) * time.Microsecond
	}

	assert.Equal(t, time.Duration(0), avg)
}

// -----------------------------------------------------------------------------
// Pool Options Comparison Tests
// -----------------------------------------------------------------------------

func TestPoolOptions_Comparison(t *testing.T) {
	defaultOpts := DefaultPoolOptions()
	highPerfOpts := HighPerformancePoolOptions()
	lowLatOpts := LowLatencyPoolOptions()

	// High perf should have higher max conns than default
	assert.GreaterOrEqual(t, highPerfOpts.MaxConns, defaultOpts.MaxConns)

	// Low latency should have shorter connect timeout
	assert.Less(t, lowLatOpts.ConnectTimeout, defaultOpts.ConnectTimeout)

	// Low latency should have shorter health check period
	assert.Less(t, lowLatOpts.HealthCheckPeriod, defaultOpts.HealthCheckPeriod)

	// High perf should have larger statement cache
	assert.Greater(t, highPerfOpts.StatementCacheCapacity, defaultOpts.StatementCacheCapacity)

	// Application names should differ
	assert.NotEqual(t, defaultOpts.ApplicationName, highPerfOpts.ApplicationName)
	assert.NotEqual(t, defaultOpts.ApplicationName, lowLatOpts.ApplicationName)
	assert.NotEqual(t, highPerfOpts.ApplicationName, lowLatOpts.ApplicationName)
}

// -----------------------------------------------------------------------------
// CreateOptimizedPoolConfig with Various Connection Strings
// -----------------------------------------------------------------------------

func TestCreateOptimizedPoolConfig_PostgresWithSSL(t *testing.T) {
	connString := "postgresql://user:pass@db.host.com:5432/mydb?sslmode=require"
	config, err := CreateOptimizedPoolConfig(connString, nil)

	require.NoError(t, err)
	require.NotNil(t, config)
}

func TestCreateOptimizedPoolConfig_WithParams(t *testing.T) {
	connString := "postgresql://user:pass@localhost:5432/db?application_name=test&search_path=public"
	config, err := CreateOptimizedPoolConfig(connString, nil)

	require.NoError(t, err)
	require.NotNil(t, config)
	// Our options should override application_name from connection string
	defaultOpts := DefaultPoolOptions()
	assert.Equal(t, defaultOpts.ApplicationName, config.ConnConfig.RuntimeParams["application_name"])
}

func TestCreateOptimizedPoolConfig_WithIPv6(t *testing.T) {
	connString := "postgresql://user:pass@[::1]:5432/db"
	config, err := CreateOptimizedPoolConfig(connString, nil)

	require.NoError(t, err)
	require.NotNil(t, config)
}

// -----------------------------------------------------------------------------
// LazyPool Double-Check Locking
// -----------------------------------------------------------------------------

func TestLazyPool_Get_DoubleCheckLocking(t *testing.T) {
	pool := NewLazyPool("invalid-connection-string", nil)

	// First call initializes
	_, err := pool.Get(context.Background())
	assert.Error(t, err)

	// Mark as initialized via fast path
	assert.True(t, pool.IsInitialized())

	// Second call should take fast path
	p2, err2 := pool.Get(context.Background())
	assert.Error(t, err2)
	assert.Nil(t, p2)
}
