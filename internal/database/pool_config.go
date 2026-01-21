package database

import (
	"context"
	"fmt"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PoolConfigOptions provides configurable pool settings
type PoolConfigOptions struct {
	// Maximum number of connections in the pool
	MaxConns int32
	// Minimum number of connections to maintain
	MinConns int32
	// Maximum lifetime of a connection
	MaxConnLifetime time.Duration
	// Maximum idle time for a connection
	MaxConnIdleTime time.Duration
	// Health check period
	HealthCheckPeriod time.Duration
	// Connection timeout
	ConnectTimeout time.Duration
	// Enable prepared statement caching
	EnableStatementCache bool
	// Statement cache capacity
	StatementCacheCapacity int
	// Use simple protocol (faster for simple queries)
	PreferSimpleProtocol bool
	// Application name for connection identification
	ApplicationName string
}

// DefaultPoolOptions returns optimized default pool options
func DefaultPoolOptions() *PoolConfigOptions {
	cpuCount := int32(runtime.NumCPU()) // #nosec G115 - CPU count fits in int32
	// Rule of thumb: (2 * CPU cores) + effective spindle count (1 for SSD)
	maxConns := cpuCount*2 + 1
	if maxConns < 10 {
		maxConns = 10
	}
	if maxConns > 50 {
		maxConns = 50
	}

	return &PoolConfigOptions{
		MaxConns:               maxConns,
		MinConns:               cpuCount / 2,
		MaxConnLifetime:        time.Hour,
		MaxConnIdleTime:        30 * time.Minute,
		HealthCheckPeriod:      30 * time.Second,
		ConnectTimeout:         5 * time.Second,
		EnableStatementCache:   true,
		StatementCacheCapacity: 512,
		PreferSimpleProtocol:   true,
		ApplicationName:        "helixagent",
	}
}

// HighPerformancePoolOptions returns options optimized for high throughput
func HighPerformancePoolOptions() *PoolConfigOptions {
	cpuCount := int32(runtime.NumCPU()) // #nosec G115 - CPU count fits in int32
	maxConns := cpuCount * 4
	if maxConns < 20 {
		maxConns = 20
	}
	if maxConns > 100 {
		maxConns = 100
	}

	return &PoolConfigOptions{
		MaxConns:               maxConns,
		MinConns:               maxConns / 2,
		MaxConnLifetime:        30 * time.Minute,
		MaxConnIdleTime:        10 * time.Minute,
		HealthCheckPeriod:      15 * time.Second,
		ConnectTimeout:         3 * time.Second,
		EnableStatementCache:   true,
		StatementCacheCapacity: 1024,
		PreferSimpleProtocol:   true,
		ApplicationName:        "helixagent-high-perf",
	}
}

// LowLatencyPoolOptions returns options optimized for low latency
func LowLatencyPoolOptions() *PoolConfigOptions {
	cpuCount := int32(runtime.NumCPU()) // #nosec G115 - CPU count fits in int32

	return &PoolConfigOptions{
		MaxConns:               cpuCount * 2,
		MinConns:               cpuCount,
		MaxConnLifetime:        15 * time.Minute,
		MaxConnIdleTime:        5 * time.Minute,
		HealthCheckPeriod:      10 * time.Second,
		ConnectTimeout:         1 * time.Second,
		EnableStatementCache:   true,
		StatementCacheCapacity: 256,
		PreferSimpleProtocol:   true,
		ApplicationName:        "helixagent-low-latency",
	}
}

// CreateOptimizedPoolConfig creates a pgxpool.Config with optimized settings
func CreateOptimizedPoolConfig(connString string, opts *PoolConfigOptions) (*pgxpool.Config, error) {
	if opts == nil {
		opts = DefaultPoolOptions()
	}

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("parse connection string: %w", err)
	}

	// Pool size settings
	config.MaxConns = opts.MaxConns
	config.MinConns = opts.MinConns
	config.MaxConnLifetime = opts.MaxConnLifetime
	config.MaxConnIdleTime = opts.MaxConnIdleTime
	config.HealthCheckPeriod = opts.HealthCheckPeriod

	// Connection settings
	config.ConnConfig.ConnectTimeout = opts.ConnectTimeout

	// Runtime parameters
	config.ConnConfig.RuntimeParams["application_name"] = opts.ApplicationName

	// Statement cache configuration
	if opts.EnableStatementCache {
		config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheStatement
	}

	// Simple protocol for faster simple queries
	if opts.PreferSimpleProtocol {
		config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	}

	// Configure after connect hook for additional setup
	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		// Set session-level optimizations
		_, err := conn.Exec(ctx, "SET synchronous_commit = off")
		if err != nil {
			return fmt.Errorf("set synchronous_commit: %w", err)
		}
		return nil
	}

	return config, nil
}

// OptimizedPool wraps pgxpool.Pool with additional metrics and features
type OptimizedPool struct {
	pool    *pgxpool.Pool
	metrics *PoolMetrics
	opts    *PoolConfigOptions
}

// PoolMetrics tracks connection pool statistics
type PoolMetrics struct {
	AcquireCount        int64
	AcquireErrors       int64
	TotalAcquireTimeUs  int64
	MaxConcurrent       int64
	CurrentConcurrent   int64
	IdleConns           int64
	TotalConns          int64
	WaitCount           int64
}

// NewOptimizedPool creates an optimized connection pool
func NewOptimizedPool(ctx context.Context, connString string, opts *PoolConfigOptions) (*OptimizedPool, error) {
	config, err := CreateOptimizedPoolConfig(connString, opts)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &OptimizedPool{
		pool:    pool,
		metrics: &PoolMetrics{},
		opts:    opts,
	}, nil
}

// Acquire acquires a connection from the pool with metrics
func (p *OptimizedPool) Acquire(ctx context.Context) (*pgxpool.Conn, error) {
	start := time.Now()
	atomic.AddInt64(&p.metrics.WaitCount, 1)

	conn, err := p.pool.Acquire(ctx)

	atomic.AddInt64(&p.metrics.TotalAcquireTimeUs, time.Since(start).Microseconds())
	atomic.AddInt64(&p.metrics.AcquireCount, 1)

	if err != nil {
		atomic.AddInt64(&p.metrics.AcquireErrors, 1)
		return nil, err
	}

	current := atomic.AddInt64(&p.metrics.CurrentConcurrent, 1)
	for {
		max := atomic.LoadInt64(&p.metrics.MaxConcurrent)
		if current <= max || atomic.CompareAndSwapInt64(&p.metrics.MaxConcurrent, max, current) {
			break
		}
	}

	return conn, nil
}

// Release releases a connection back to the pool
func (p *OptimizedPool) Release(conn *pgxpool.Conn) {
	atomic.AddInt64(&p.metrics.CurrentConcurrent, -1)
	conn.Release()
}

// Query executes a query with automatic connection management
func (p *OptimizedPool) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return p.pool.Query(ctx, sql, args...)
}

// QueryRow executes a query that returns a single row
func (p *OptimizedPool) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return p.pool.QueryRow(ctx, sql, args...)
}

// Exec executes a query that doesn't return rows
func (p *OptimizedPool) Exec(ctx context.Context, sql string, args ...interface{}) (int64, error) {
	tag, err := p.pool.Exec(ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// SendBatch sends a batch of queries
func (p *OptimizedPool) SendBatch(ctx context.Context, batch *pgx.Batch) pgx.BatchResults {
	return p.pool.SendBatch(ctx, batch)
}

// BeginTx starts a transaction
func (p *OptimizedPool) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	return p.pool.BeginTx(ctx, txOptions)
}

// CopyFrom performs a COPY FROM operation
func (p *OptimizedPool) CopyFrom(ctx context.Context, tableName pgx.Identifier, columns []string, rows pgx.CopyFromSource) (int64, error) {
	return p.pool.CopyFrom(ctx, tableName, columns, rows)
}

// Stat returns pool statistics
func (p *OptimizedPool) Stat() *pgxpool.Stat {
	return p.pool.Stat()
}

// Metrics returns pool metrics
func (p *OptimizedPool) Metrics() *PoolMetrics {
	stat := p.pool.Stat()

	return &PoolMetrics{
		AcquireCount:       atomic.LoadInt64(&p.metrics.AcquireCount),
		AcquireErrors:      atomic.LoadInt64(&p.metrics.AcquireErrors),
		TotalAcquireTimeUs: atomic.LoadInt64(&p.metrics.TotalAcquireTimeUs),
		MaxConcurrent:      atomic.LoadInt64(&p.metrics.MaxConcurrent),
		CurrentConcurrent:  atomic.LoadInt64(&p.metrics.CurrentConcurrent),
		IdleConns:          int64(stat.IdleConns()),
		TotalConns:         int64(stat.TotalConns()),
		WaitCount:          atomic.LoadInt64(&p.metrics.WaitCount),
	}
}

// AverageAcquireTime returns the average time to acquire a connection
func (p *OptimizedPool) AverageAcquireTime() time.Duration {
	count := atomic.LoadInt64(&p.metrics.AcquireCount)
	if count == 0 {
		return 0
	}
	totalUs := atomic.LoadInt64(&p.metrics.TotalAcquireTimeUs)
	return time.Duration(totalUs/count) * time.Microsecond
}

// HealthCheck performs a health check on the pool
func (p *OptimizedPool) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	return p.pool.Ping(ctx)
}

// Close closes the pool
func (p *OptimizedPool) Close() {
	p.pool.Close()
}

// Pool returns the underlying pgxpool.Pool
func (p *OptimizedPool) Pool() *pgxpool.Pool {
	return p.pool
}

// LazyPool provides lazy initialization of database connection pool
type LazyPool struct {
	pool      *OptimizedPool
	connStr   string
	opts      *PoolConfigOptions
	initErr   error
	initOnce  int32
	initMu    chan struct{}
}

// NewLazyPool creates a new lazy-initialized pool
func NewLazyPool(connString string, opts *PoolConfigOptions) *LazyPool {
	return &LazyPool{
		connStr:  connString,
		opts:     opts,
		initMu:   make(chan struct{}, 1),
	}
}

// Get returns the pool, initializing on first call
func (p *LazyPool) Get(ctx context.Context) (*OptimizedPool, error) {
	// Fast path: already initialized
	if atomic.LoadInt32(&p.initOnce) == 1 {
		return p.pool, p.initErr
	}

	// Slow path: try to initialize
	select {
	case p.initMu <- struct{}{}:
		defer func() { <-p.initMu }()

		// Double check after acquiring lock
		if atomic.LoadInt32(&p.initOnce) == 1 {
			return p.pool, p.initErr
		}

		// Initialize
		pool, err := NewOptimizedPool(ctx, p.connStr, p.opts)
		p.pool = pool
		p.initErr = err
		atomic.StoreInt32(&p.initOnce, 1)

		return p.pool, p.initErr

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// IsInitialized returns whether the pool has been initialized
func (p *LazyPool) IsInitialized() bool {
	return atomic.LoadInt32(&p.initOnce) == 1
}

// Close closes the pool if initialized
func (p *LazyPool) Close() {
	if atomic.LoadInt32(&p.initOnce) == 1 && p.pool != nil {
		p.pool.Close()
	}
}
