package database

import (
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultPoolOptions(t *testing.T) {
	opts := DefaultPoolOptions()

	require.NotNil(t, opts)
	assert.Greater(t, opts.MaxConns, int32(0))
	assert.GreaterOrEqual(t, opts.MaxConns, int32(10))
	assert.LessOrEqual(t, opts.MaxConns, int32(50))
	assert.GreaterOrEqual(t, opts.MinConns, int32(0))
	assert.Equal(t, time.Hour, opts.MaxConnLifetime)
	assert.Equal(t, 30*time.Minute, opts.MaxConnIdleTime)
	assert.Equal(t, 30*time.Second, opts.HealthCheckPeriod)
	assert.Equal(t, 5*time.Second, opts.ConnectTimeout)
	assert.True(t, opts.EnableStatementCache)
	assert.Equal(t, 512, opts.StatementCacheCapacity)
	assert.True(t, opts.PreferSimpleProtocol)
	assert.Equal(t, "helixagent", opts.ApplicationName)
}

func TestDefaultPoolOptions_MaxConns_Bounds(t *testing.T) {
	opts := DefaultPoolOptions()

	cpuCount := int32(runtime.NumCPU())
	expectedMax := cpuCount*2 + 1

	// Check bounds are applied
	if expectedMax < 10 {
		assert.GreaterOrEqual(t, opts.MaxConns, int32(10))
	}
	if expectedMax > 50 {
		assert.LessOrEqual(t, opts.MaxConns, int32(50))
	}
}

func TestHighPerformancePoolOptions(t *testing.T) {
	opts := HighPerformancePoolOptions()

	require.NotNil(t, opts)
	assert.GreaterOrEqual(t, opts.MaxConns, int32(20))
	assert.LessOrEqual(t, opts.MaxConns, int32(100))
	assert.Equal(t, opts.MaxConns/2, opts.MinConns)
	assert.Equal(t, 30*time.Minute, opts.MaxConnLifetime)
	assert.Equal(t, 10*time.Minute, opts.MaxConnIdleTime)
	assert.Equal(t, 15*time.Second, opts.HealthCheckPeriod)
	assert.Equal(t, 3*time.Second, opts.ConnectTimeout)
	assert.True(t, opts.EnableStatementCache)
	assert.Equal(t, 1024, opts.StatementCacheCapacity)
	assert.True(t, opts.PreferSimpleProtocol)
	assert.Equal(t, "helixagent-high-perf", opts.ApplicationName)
}

func TestLowLatencyPoolOptions(t *testing.T) {
	opts := LowLatencyPoolOptions()

	require.NotNil(t, opts)
	cpuCount := int32(runtime.NumCPU())
	assert.Equal(t, cpuCount*2, opts.MaxConns)
	assert.Equal(t, cpuCount, opts.MinConns)
	assert.Equal(t, 15*time.Minute, opts.MaxConnLifetime)
	assert.Equal(t, 5*time.Minute, opts.MaxConnIdleTime)
	assert.Equal(t, 10*time.Second, opts.HealthCheckPeriod)
	assert.Equal(t, 1*time.Second, opts.ConnectTimeout)
	assert.True(t, opts.EnableStatementCache)
	assert.Equal(t, 256, opts.StatementCacheCapacity)
	assert.True(t, opts.PreferSimpleProtocol)
	assert.Equal(t, "helixagent-low-latency", opts.ApplicationName)
}

func TestCreateOptimizedPoolConfig_InvalidConnString(t *testing.T) {
	_, err := CreateOptimizedPoolConfig("invalid connection string", nil)
	assert.Error(t, err)
}

func TestCreateOptimizedPoolConfig_ValidConnString(t *testing.T) {
	connString := "postgresql://user:password@localhost:5432/testdb"
	config, err := CreateOptimizedPoolConfig(connString, nil)

	require.NoError(t, err)
	require.NotNil(t, config)

	// Should use default options
	defaultOpts := DefaultPoolOptions()
	assert.Equal(t, defaultOpts.MaxConns, config.MaxConns)
	assert.Equal(t, defaultOpts.MinConns, config.MinConns)
	assert.Equal(t, defaultOpts.MaxConnLifetime, config.MaxConnLifetime)
	assert.Equal(t, defaultOpts.MaxConnIdleTime, config.MaxConnIdleTime)
	assert.Equal(t, defaultOpts.HealthCheckPeriod, config.HealthCheckPeriod)
}

func TestCreateOptimizedPoolConfig_WithCustomOptions(t *testing.T) {
	connString := "postgresql://user:password@localhost:5432/testdb"
	customOpts := &PoolConfigOptions{
		MaxConns:               25,
		MinConns:               5,
		MaxConnLifetime:        2 * time.Hour,
		MaxConnIdleTime:        45 * time.Minute,
		HealthCheckPeriod:      1 * time.Minute,
		ConnectTimeout:         10 * time.Second,
		EnableStatementCache:   false,
		StatementCacheCapacity: 100,
		PreferSimpleProtocol:   false,
		ApplicationName:        "custom-app",
	}

	config, err := CreateOptimizedPoolConfig(connString, customOpts)

	require.NoError(t, err)
	require.NotNil(t, config)
	assert.Equal(t, int32(25), config.MaxConns)
	assert.Equal(t, int32(5), config.MinConns)
	assert.Equal(t, 2*time.Hour, config.MaxConnLifetime)
	assert.Equal(t, 45*time.Minute, config.MaxConnIdleTime)
	assert.Equal(t, 1*time.Minute, config.HealthCheckPeriod)
	assert.Equal(t, 10*time.Second, config.ConnConfig.ConnectTimeout)
	assert.Equal(t, "custom-app", config.ConnConfig.RuntimeParams["application_name"])
}

func TestPoolConfigOptions_Fields(t *testing.T) {
	opts := PoolConfigOptions{
		MaxConns:               50,
		MinConns:               10,
		MaxConnLifetime:        time.Hour,
		MaxConnIdleTime:        30 * time.Minute,
		HealthCheckPeriod:      30 * time.Second,
		ConnectTimeout:         5 * time.Second,
		EnableStatementCache:   true,
		StatementCacheCapacity: 512,
		PreferSimpleProtocol:   true,
		ApplicationName:        "test-app",
	}

	assert.Equal(t, int32(50), opts.MaxConns)
	assert.Equal(t, int32(10), opts.MinConns)
	assert.Equal(t, time.Hour, opts.MaxConnLifetime)
	assert.Equal(t, 30*time.Minute, opts.MaxConnIdleTime)
	assert.Equal(t, 30*time.Second, opts.HealthCheckPeriod)
	assert.Equal(t, 5*time.Second, opts.ConnectTimeout)
	assert.True(t, opts.EnableStatementCache)
	assert.Equal(t, 512, opts.StatementCacheCapacity)
	assert.True(t, opts.PreferSimpleProtocol)
	assert.Equal(t, "test-app", opts.ApplicationName)
}

func TestPoolMetrics_Fields(t *testing.T) {
	metrics := &PoolMetrics{
		AcquireCount:       100,
		AcquireErrors:      5,
		TotalAcquireTimeUs: 50000,
		MaxConcurrent:      10,
		CurrentConcurrent:  3,
		IdleConns:          7,
		TotalConns:         10,
		WaitCount:          50,
	}

	assert.Equal(t, int64(100), metrics.AcquireCount)
	assert.Equal(t, int64(5), metrics.AcquireErrors)
	assert.Equal(t, int64(50000), metrics.TotalAcquireTimeUs)
	assert.Equal(t, int64(10), metrics.MaxConcurrent)
	assert.Equal(t, int64(3), metrics.CurrentConcurrent)
	assert.Equal(t, int64(7), metrics.IdleConns)
	assert.Equal(t, int64(10), metrics.TotalConns)
	assert.Equal(t, int64(50), metrics.WaitCount)
}

func TestNewLazyPool(t *testing.T) {
	connString := "postgresql://user:password@localhost:5432/testdb"
	opts := DefaultPoolOptions()

	pool := NewLazyPool(connString, opts)

	require.NotNil(t, pool)
	assert.Equal(t, connString, pool.connStr)
	assert.Equal(t, opts, pool.opts)
	assert.False(t, pool.IsInitialized())
}

func TestLazyPool_IsInitialized_NotInitialized(t *testing.T) {
	pool := NewLazyPool("postgresql://localhost/test", nil)

	assert.False(t, pool.IsInitialized())
}

func TestLazyPool_Close_NotInitialized(t *testing.T) {
	pool := NewLazyPool("postgresql://localhost/test", nil)

	// Should not panic even when not initialized
	pool.Close()
}
