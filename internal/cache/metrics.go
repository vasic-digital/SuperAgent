package cache

import (
	"sync/atomic"
	"time"
)

// CacheMetrics provides unified cache performance metrics
type CacheMetrics struct {
	// Global metrics
	TotalHits          int64
	TotalMisses        int64
	TotalSets          int64
	TotalDeletes       int64
	TotalInvalidations int64

	// Performance metrics
	TotalGetLatencyUs int64
	TotalSetLatencyUs int64
	GetCount          int64
	SetCount          int64

	// Memory metrics
	L1Size    int64
	L1MaxSize int64

	// Error metrics
	GetErrors int64
	SetErrors int64

	// Tiered cache specific
	L1Hits   int64
	L1Misses int64
	L2Hits   int64
	L2Misses int64

	// Compression metrics
	CompressionSaved int64

	// By-type breakdown
	ProviderHits   int64
	ProviderMisses int64
	MCPHits        int64
	MCPMisses      int64
	SessionHits    int64
	SessionMisses  int64
}

// CacheMetricsCollector aggregates metrics from multiple cache components
type CacheMetricsCollector struct {
	tieredCache   *TieredCache
	providerCache *ProviderCache
	mcpCache      *MCPServerCache
	expiration    *ExpirationManager

	aggregated *CacheMetrics
	lastUpdate time.Time
}

// NewCacheMetricsCollector creates a new metrics collector
func NewCacheMetricsCollector(
	tieredCache *TieredCache,
	providerCache *ProviderCache,
	mcpCache *MCPServerCache,
	expiration *ExpirationManager,
) *CacheMetricsCollector {
	return &CacheMetricsCollector{
		tieredCache:   tieredCache,
		providerCache: providerCache,
		mcpCache:      mcpCache,
		expiration:    expiration,
		aggregated:    &CacheMetrics{},
	}
}

// Collect gathers metrics from all cache components
func (c *CacheMetricsCollector) Collect() *CacheMetrics {
	metrics := &CacheMetrics{}

	// Collect tiered cache metrics
	if c.tieredCache != nil {
		tm := c.tieredCache.Metrics()
		metrics.L1Hits = tm.L1Hits
		metrics.L1Misses = tm.L1Misses
		metrics.L2Hits = tm.L2Hits
		metrics.L2Misses = tm.L2Misses
		metrics.L1Size = tm.L1Size
		metrics.CompressionSaved = tm.CompressionSaved
		metrics.TotalInvalidations = tm.Invalidations

		metrics.TotalHits = tm.L1Hits + tm.L2Hits
		metrics.TotalMisses = tm.L1Misses + tm.L2Misses
	}

	// Collect provider cache metrics
	if c.providerCache != nil {
		pm := c.providerCache.Metrics()
		metrics.ProviderHits = pm.Hits
		metrics.ProviderMisses = pm.Misses
		metrics.TotalSets += pm.Sets
	}

	// Collect MCP cache metrics
	if c.mcpCache != nil {
		mm := c.mcpCache.Metrics()
		metrics.MCPHits = mm.Hits
		metrics.MCPMisses = mm.Misses
		metrics.TotalSets += mm.Sets
	}

	c.aggregated = metrics
	c.lastUpdate = time.Now()

	return metrics
}

// HitRate returns the overall cache hit rate as a percentage
func (c *CacheMetricsCollector) HitRate() float64 {
	metrics := c.Collect()
	total := metrics.TotalHits + metrics.TotalMisses
	if total == 0 {
		return 0
	}
	return float64(metrics.TotalHits) / float64(total) * 100
}

// L1HitRate returns the L1 (memory) cache hit rate
func (c *CacheMetricsCollector) L1HitRate() float64 {
	metrics := c.Collect()
	total := metrics.L1Hits + metrics.L1Misses
	if total == 0 {
		return 0
	}
	return float64(metrics.L1Hits) / float64(total) * 100
}

// L2HitRate returns the L2 (Redis) cache hit rate
func (c *CacheMetricsCollector) L2HitRate() float64 {
	metrics := c.Collect()
	// L2 hits are when L1 missed but L2 hit
	total := metrics.L2Hits + metrics.L2Misses
	if total == 0 {
		return 0
	}
	return float64(metrics.L2Hits) / float64(total) * 100
}

// AverageGetLatency returns the average get latency
func (c *CacheMetricsCollector) AverageGetLatency() time.Duration {
	if c.aggregated.GetCount == 0 {
		return 0
	}
	return time.Duration(c.aggregated.TotalGetLatencyUs/c.aggregated.GetCount) * time.Microsecond
}

// AverageSetLatency returns the average set latency
func (c *CacheMetricsCollector) AverageSetLatency() time.Duration {
	if c.aggregated.SetCount == 0 {
		return 0
	}
	return time.Duration(c.aggregated.TotalSetLatencyUs/c.aggregated.SetCount) * time.Microsecond
}

// CompressionRatio returns bytes saved by compression
func (c *CacheMetricsCollector) CompressionSavings() int64 {
	return c.aggregated.CompressionSaved
}

// Summary returns a summary of cache performance
func (c *CacheMetricsCollector) Summary() *CacheSummary {
	metrics := c.Collect()

	return &CacheSummary{
		HitRate:            c.HitRate(),
		L1HitRate:          c.L1HitRate(),
		L2HitRate:          c.L2HitRate(),
		TotalHits:          metrics.TotalHits,
		TotalMisses:        metrics.TotalMisses,
		TotalInvalidations: metrics.TotalInvalidations,
		L1Size:             metrics.L1Size,
		CompressionSaved:   metrics.CompressionSaved,
		ProviderHitRate:    c.providerHitRate(),
		MCPHitRate:         c.mcpHitRate(),
		LastUpdate:         c.lastUpdate,
	}
}

func (c *CacheMetricsCollector) providerHitRate() float64 {
	total := c.aggregated.ProviderHits + c.aggregated.ProviderMisses
	if total == 0 {
		return 0
	}
	return float64(c.aggregated.ProviderHits) / float64(total) * 100
}

func (c *CacheMetricsCollector) mcpHitRate() float64 {
	total := c.aggregated.MCPHits + c.aggregated.MCPMisses
	if total == 0 {
		return 0
	}
	return float64(c.aggregated.MCPHits) / float64(total) * 100
}

// CacheSummary provides a high-level summary of cache performance
type CacheSummary struct {
	HitRate            float64   `json:"hit_rate"`
	L1HitRate          float64   `json:"l1_hit_rate"`
	L2HitRate          float64   `json:"l2_hit_rate"`
	TotalHits          int64     `json:"total_hits"`
	TotalMisses        int64     `json:"total_misses"`
	TotalInvalidations int64     `json:"total_invalidations"`
	L1Size             int64     `json:"l1_size"`
	CompressionSaved   int64     `json:"compression_saved_bytes"`
	ProviderHitRate    float64   `json:"provider_hit_rate"`
	MCPHitRate         float64   `json:"mcp_hit_rate"`
	LastUpdate         time.Time `json:"last_update"`
}

// IncrementingMetrics provides a thread-safe way to track incrementing metrics
type IncrementingMetrics struct {
	hits          int64
	misses        int64
	sets          int64
	deletes       int64
	invalidations int64
	errors        int64
}

// NewIncrementingMetrics creates a new incrementing metrics tracker
func NewIncrementingMetrics() *IncrementingMetrics {
	return &IncrementingMetrics{}
}

// Hit records a cache hit
func (m *IncrementingMetrics) Hit() {
	atomic.AddInt64(&m.hits, 1)
}

// Miss records a cache miss
func (m *IncrementingMetrics) Miss() {
	atomic.AddInt64(&m.misses, 1)
}

// Set records a cache set
func (m *IncrementingMetrics) Set() {
	atomic.AddInt64(&m.sets, 1)
}

// Delete records a cache delete
func (m *IncrementingMetrics) Delete() {
	atomic.AddInt64(&m.deletes, 1)
}

// Invalidate records a cache invalidation
func (m *IncrementingMetrics) Invalidate() {
	atomic.AddInt64(&m.invalidations, 1)
}

// Error records a cache error
func (m *IncrementingMetrics) Error() {
	atomic.AddInt64(&m.errors, 1)
}

// Hits returns the total hits
func (m *IncrementingMetrics) Hits() int64 {
	return atomic.LoadInt64(&m.hits)
}

// Misses returns the total misses
func (m *IncrementingMetrics) Misses() int64 {
	return atomic.LoadInt64(&m.misses)
}

// Sets returns the total sets
func (m *IncrementingMetrics) Sets() int64 {
	return atomic.LoadInt64(&m.sets)
}

// Deletes returns the total deletes
func (m *IncrementingMetrics) Deletes() int64 {
	return atomic.LoadInt64(&m.deletes)
}

// Invalidations returns the total invalidations
func (m *IncrementingMetrics) Invalidations() int64 {
	return atomic.LoadInt64(&m.invalidations)
}

// Errors returns the total errors
func (m *IncrementingMetrics) Errors() int64 {
	return atomic.LoadInt64(&m.errors)
}

// HitRate returns the hit rate as a percentage
func (m *IncrementingMetrics) HitRate() float64 {
	hits := atomic.LoadInt64(&m.hits)
	misses := atomic.LoadInt64(&m.misses)
	total := hits + misses
	if total == 0 {
		return 0
	}
	return float64(hits) / float64(total) * 100
}

// Reset resets all metrics to zero
func (m *IncrementingMetrics) Reset() {
	atomic.StoreInt64(&m.hits, 0)
	atomic.StoreInt64(&m.misses, 0)
	atomic.StoreInt64(&m.sets, 0)
	atomic.StoreInt64(&m.deletes, 0)
	atomic.StoreInt64(&m.invalidations, 0)
	atomic.StoreInt64(&m.errors, 0)
}
