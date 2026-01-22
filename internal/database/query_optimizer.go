package database

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// QueryOptimizer provides optimized query execution with prepared statements,
// query caching, and batch operations
type QueryOptimizer struct {
	pool          *pgxpool.Pool
	preparedStmts map[string]*pgxpool.Conn
	stmtMu        sync.RWMutex
	queryCache    *QueryCache
	metrics       *QueryMetrics
	config        *OptimizerConfig
}

// OptimizerConfig holds configuration for the query optimizer
type OptimizerConfig struct {
	// Maximum prepared statements to cache
	MaxPreparedStmts int
	// Query cache TTL
	CacheTTL time.Duration
	// Enable query caching
	EnableCache bool
	// Batch size for bulk operations
	DefaultBatchSize int
	// Query timeout
	QueryTimeout time.Duration
}

// DefaultOptimizerConfig returns sensible defaults
func DefaultOptimizerConfig() *OptimizerConfig {
	return &OptimizerConfig{
		MaxPreparedStmts: 100,
		CacheTTL:         5 * time.Minute,
		EnableCache:      true,
		DefaultBatchSize: 1000,
		QueryTimeout:     30 * time.Second,
	}
}

// QueryMetrics tracks query performance statistics
type QueryMetrics struct {
	TotalQueries      int64
	CacheHits         int64
	CacheMisses       int64
	TotalLatencyUs    int64
	SlowQueries       int64 // queries > 100ms
	PreparedStmtHits  int64
	BulkInsertRows    int64
	BulkInsertBatches int64
}

// QueryCache provides simple query result caching
type QueryCache struct {
	cache   map[string]*cacheEntry
	mu      sync.RWMutex
	ttl     time.Duration
	maxSize int
}

type cacheEntry struct {
	result    interface{}
	expiresAt time.Time
}

// NewQueryCache creates a new query cache
func NewQueryCache(ttl time.Duration, maxSize int) *QueryCache {
	qc := &QueryCache{
		cache:   make(map[string]*cacheEntry),
		ttl:     ttl,
		maxSize: maxSize,
	}
	go qc.cleanupLoop()
	return qc
}

func (c *QueryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.cache[key]
	if !exists {
		return nil, false
	}
	if time.Now().After(entry.expiresAt) {
		return nil, false
	}
	return entry.result, true
}

func (c *QueryCache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict oldest if at capacity
	if len(c.cache) >= c.maxSize {
		var oldestKey string
		var oldestTime time.Time
		for k, v := range c.cache {
			if oldestKey == "" || v.expiresAt.Before(oldestTime) {
				oldestKey = k
				oldestTime = v.expiresAt
			}
		}
		if oldestKey != "" {
			delete(c.cache, oldestKey)
		}
	}

	c.cache[key] = &cacheEntry{
		result:    value,
		expiresAt: time.Now().Add(c.ttl),
	}
}

func (c *QueryCache) Invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, key)
}

func (c *QueryCache) InvalidatePrefix(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key := range c.cache {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(c.cache, key)
		}
	}
}

func (c *QueryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[string]*cacheEntry)
}

func (c *QueryCache) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.cache {
			if now.After(entry.expiresAt) {
				delete(c.cache, key)
			}
		}
		c.mu.Unlock()
	}
}

// NewQueryOptimizer creates a new query optimizer
func NewQueryOptimizer(pool *pgxpool.Pool, config *OptimizerConfig) *QueryOptimizer {
	if config == nil {
		config = DefaultOptimizerConfig()
	}

	var cache *QueryCache
	if config.EnableCache {
		cache = NewQueryCache(config.CacheTTL, config.MaxPreparedStmts*10)
	}

	return &QueryOptimizer{
		pool:          pool,
		preparedStmts: make(map[string]*pgxpool.Conn),
		queryCache:    cache,
		metrics:       &QueryMetrics{},
		config:        config,
	}
}

// GetActiveProviders returns healthy, enabled providers optimized for routing
func (o *QueryOptimizer) GetActiveProviders(ctx context.Context) ([]ActiveProvider, error) {
	cacheKey := "active_providers"

	// Check cache first
	if o.queryCache != nil {
		if cached, ok := o.queryCache.Get(cacheKey); ok {
			atomic.AddInt64(&o.metrics.CacheHits, 1)
			return cached.([]ActiveProvider), nil
		}
		atomic.AddInt64(&o.metrics.CacheMisses, 1)
	}

	start := time.Now()
	defer func() {
		latency := time.Since(start).Microseconds()
		atomic.AddInt64(&o.metrics.TotalLatencyUs, latency)
		atomic.AddInt64(&o.metrics.TotalQueries, 1)
		if latency > 100000 {
			atomic.AddInt64(&o.metrics.SlowQueries, 1)
		}
	}()

	ctx, cancel := context.WithTimeout(ctx, o.config.QueryTimeout)
	defer cancel()

	// Optimized query using the performance index
	const query = `
		SELECT id, name, type, weight, health_status, response_time, config
		FROM llm_providers
		WHERE enabled = TRUE AND health_status = 'healthy'
		ORDER BY weight DESC, response_time ASC
		LIMIT 20
	`

	rows, err := o.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query active providers: %w", err)
	}
	defer rows.Close()

	var providers []ActiveProvider
	for rows.Next() {
		var p ActiveProvider
		if err := rows.Scan(&p.ID, &p.Name, &p.Type, &p.Weight, &p.HealthStatus, &p.ResponseTime, &p.Config); err != nil {
			return nil, fmt.Errorf("scan provider: %w", err)
		}
		providers = append(providers, p)
	}

	// Cache result
	if o.queryCache != nil && len(providers) > 0 {
		o.queryCache.Set(cacheKey, providers)
	}

	return providers, nil
}

// ActiveProvider represents a provider ready for routing
type ActiveProvider struct {
	ID           string
	Name         string
	Type         string
	Weight       float64
	HealthStatus string
	ResponseTime int64
	Config       []byte
}

// GetProviderPerformance returns aggregated provider performance from materialized view
func (o *QueryOptimizer) GetProviderPerformance(ctx context.Context) ([]ProviderPerformance, error) {
	cacheKey := "provider_performance"

	if o.queryCache != nil {
		if cached, ok := o.queryCache.Get(cacheKey); ok {
			atomic.AddInt64(&o.metrics.CacheHits, 1)
			return cached.([]ProviderPerformance), nil
		}
		atomic.AddInt64(&o.metrics.CacheMisses, 1)
	}

	start := time.Now()
	defer func() {
		atomic.AddInt64(&o.metrics.TotalLatencyUs, time.Since(start).Microseconds())
		atomic.AddInt64(&o.metrics.TotalQueries, 1)
	}()

	ctx, cancel := context.WithTimeout(ctx, o.config.QueryTimeout)
	defer cancel()

	// Query from materialized view for fast response
	const query = `
		SELECT provider_name, provider_type, health_status,
		       total_requests_24h, avg_response_time_ms,
		       p95_response_time_ms, avg_confidence, success_rate
		FROM mv_provider_performance
		ORDER BY success_rate DESC, avg_response_time_ms ASC
	`

	rows, err := o.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query provider performance: %w", err)
	}
	defer rows.Close()

	var results []ProviderPerformance
	for rows.Next() {
		var p ProviderPerformance
		if err := rows.Scan(&p.ProviderName, &p.ProviderType, &p.HealthStatus,
			&p.TotalRequests24h, &p.AvgResponseTimeMs,
			&p.P95ResponseTimeMs, &p.AvgConfidence, &p.SuccessRate); err != nil {
			return nil, fmt.Errorf("scan performance: %w", err)
		}
		results = append(results, p)
	}

	if o.queryCache != nil && len(results) > 0 {
		o.queryCache.Set(cacheKey, results)
	}

	return results, nil
}

// ProviderPerformance represents aggregated provider metrics
type ProviderPerformance struct {
	ProviderName      string
	ProviderType      string
	HealthStatus      string
	TotalRequests24h  int64
	AvgResponseTimeMs int
	P95ResponseTimeMs int
	AvgConfidence     float64
	SuccessRate       float64
}

// GetMCPServerHealth returns MCP server health from materialized view
func (o *QueryOptimizer) GetMCPServerHealth(ctx context.Context) ([]MCPServerHealth, error) {
	cacheKey := "mcp_server_health"

	if o.queryCache != nil {
		if cached, ok := o.queryCache.Get(cacheKey); ok {
			atomic.AddInt64(&o.metrics.CacheHits, 1)
			return cached.([]MCPServerHealth), nil
		}
		atomic.AddInt64(&o.metrics.CacheMisses, 1)
	}

	start := time.Now()
	defer func() {
		atomic.AddInt64(&o.metrics.TotalLatencyUs, time.Since(start).Microseconds())
		atomic.AddInt64(&o.metrics.TotalQueries, 1)
	}()

	ctx, cancel := context.WithTimeout(ctx, o.config.QueryTimeout)
	defer cancel()

	const query = `
		SELECT server_id, server_name, server_type, enabled,
		       total_operations_1h, successful_operations, failed_operations,
		       avg_duration_ms, p95_duration_ms, success_rate, tool_count
		FROM mv_mcp_server_health
		WHERE enabled = TRUE
		ORDER BY success_rate DESC
	`

	rows, err := o.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query mcp health: %w", err)
	}
	defer rows.Close()

	var results []MCPServerHealth
	for rows.Next() {
		var h MCPServerHealth
		if err := rows.Scan(&h.ServerID, &h.ServerName, &h.ServerType, &h.Enabled,
			&h.TotalOperations1h, &h.SuccessfulOperations, &h.FailedOperations,
			&h.AvgDurationMs, &h.P95DurationMs, &h.SuccessRate, &h.ToolCount); err != nil {
			return nil, fmt.Errorf("scan mcp health: %w", err)
		}
		results = append(results, h)
	}

	if o.queryCache != nil && len(results) > 0 {
		o.queryCache.Set(cacheKey, results)
	}

	return results, nil
}

// MCPServerHealth represents MCP server health metrics
type MCPServerHealth struct {
	ServerID             string
	ServerName           string
	ServerType           string
	Enabled              bool
	TotalOperations1h    int64
	SuccessfulOperations int64
	FailedOperations     int64
	AvgDurationMs        int
	P95DurationMs        int
	SuccessRate          float64
	ToolCount            int
}

// BulkInsert performs efficient bulk inserts using COPY protocol
func (o *QueryOptimizer) BulkInsert(ctx context.Context, table string, columns []string, rows [][]interface{}) (int64, error) {
	if len(rows) == 0 {
		return 0, nil
	}

	start := time.Now()
	defer func() {
		atomic.AddInt64(&o.metrics.TotalLatencyUs, time.Since(start).Microseconds())
		atomic.AddInt64(&o.metrics.TotalQueries, 1)
	}()

	ctx, cancel := context.WithTimeout(ctx, o.config.QueryTimeout*2)
	defer cancel()

	// Use COPY for maximum throughput
	copyCount, err := o.pool.CopyFrom(
		ctx,
		pgx.Identifier{table},
		columns,
		pgx.CopyFromRows(rows),
	)

	if err != nil {
		return 0, fmt.Errorf("bulk insert to %s: %w", table, err)
	}

	atomic.AddInt64(&o.metrics.BulkInsertRows, copyCount)
	atomic.AddInt64(&o.metrics.BulkInsertBatches, 1)

	// Invalidate related cache entries
	if o.queryCache != nil {
		o.queryCache.InvalidatePrefix(table)
	}

	return copyCount, nil
}

// BulkInsertBatched performs bulk inserts in batches to avoid memory issues
func (o *QueryOptimizer) BulkInsertBatched(ctx context.Context, table string, columns []string, rows [][]interface{}) (int64, error) {
	if len(rows) == 0 {
		return 0, nil
	}

	batchSize := o.config.DefaultBatchSize
	totalInserted := int64(0)

	for i := 0; i < len(rows); i += batchSize {
		end := i + batchSize
		if end > len(rows) {
			end = len(rows)
		}

		batch := rows[i:end]
		inserted, err := o.BulkInsert(ctx, table, columns, batch)
		if err != nil {
			return totalInserted, fmt.Errorf("batch %d: %w", i/batchSize, err)
		}
		totalInserted += inserted
	}

	return totalInserted, nil
}

// RefreshMaterializedViews triggers refresh of all materialized views
func (o *QueryOptimizer) RefreshMaterializedViews(ctx context.Context) ([]ViewRefreshResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	const query = `SELECT * FROM refresh_all_materialized_views()`

	rows, err := o.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("refresh materialized views: %w", err)
	}
	defer rows.Close()

	var results []ViewRefreshResult
	for rows.Next() {
		var r ViewRefreshResult
		if err := rows.Scan(&r.ViewName, &r.Status, &r.DurationMs); err != nil {
			return nil, fmt.Errorf("scan refresh result: %w", err)
		}
		results = append(results, r)
	}

	// Clear query cache after refresh
	if o.queryCache != nil {
		o.queryCache.Clear()
	}

	return results, nil
}

// ViewRefreshResult represents the result of refreshing a materialized view
type ViewRefreshResult struct {
	ViewName   string
	Status     string
	DurationMs int
}

// RefreshCriticalViews triggers refresh of performance-critical views only
func (o *QueryOptimizer) RefreshCriticalViews(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	_, err := o.pool.Exec(ctx, "SELECT refresh_critical_views()")
	if err != nil {
		return fmt.Errorf("refresh critical views: %w", err)
	}

	// Invalidate related cache entries
	if o.queryCache != nil {
		o.queryCache.InvalidatePrefix("provider_")
		o.queryCache.InvalidatePrefix("mcp_")
		o.queryCache.InvalidatePrefix("task_")
	}

	return nil
}

// Metrics returns current query metrics
func (o *QueryOptimizer) Metrics() *QueryMetrics {
	return &QueryMetrics{
		TotalQueries:      atomic.LoadInt64(&o.metrics.TotalQueries),
		CacheHits:         atomic.LoadInt64(&o.metrics.CacheHits),
		CacheMisses:       atomic.LoadInt64(&o.metrics.CacheMisses),
		TotalLatencyUs:    atomic.LoadInt64(&o.metrics.TotalLatencyUs),
		SlowQueries:       atomic.LoadInt64(&o.metrics.SlowQueries),
		PreparedStmtHits:  atomic.LoadInt64(&o.metrics.PreparedStmtHits),
		BulkInsertRows:    atomic.LoadInt64(&o.metrics.BulkInsertRows),
		BulkInsertBatches: atomic.LoadInt64(&o.metrics.BulkInsertBatches),
	}
}

// AverageLatency returns the average query latency
func (o *QueryOptimizer) AverageLatency() time.Duration {
	total := atomic.LoadInt64(&o.metrics.TotalQueries)
	if total == 0 {
		return 0
	}
	latencyUs := atomic.LoadInt64(&o.metrics.TotalLatencyUs)
	return time.Duration(latencyUs/total) * time.Microsecond
}

// CacheHitRate returns the cache hit rate as a percentage
func (o *QueryOptimizer) CacheHitRate() float64 {
	hits := atomic.LoadInt64(&o.metrics.CacheHits)
	misses := atomic.LoadInt64(&o.metrics.CacheMisses)
	total := hits + misses
	if total == 0 {
		return 0
	}
	return float64(hits) / float64(total) * 100
}

// InvalidateCache invalidates all cache entries
func (o *QueryOptimizer) InvalidateCache() {
	if o.queryCache != nil {
		o.queryCache.Clear()
	}
}

// InvalidateCacheKey invalidates a specific cache key
func (o *QueryOptimizer) InvalidateCacheKey(key string) {
	if o.queryCache != nil {
		o.queryCache.Invalidate(key)
	}
}

// Close closes the query optimizer and releases resources
func (o *QueryOptimizer) Close() error {
	o.stmtMu.Lock()
	defer o.stmtMu.Unlock()

	// Clear prepared statement cache
	o.preparedStmts = make(map[string]*pgxpool.Conn)

	return nil
}
