package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync/atomic"
	"time"
)

// MCPCacheConfig holds configuration for MCP tool result caching
type MCPCacheConfig struct {
	// DefaultTTL for cached results
	DefaultTTL time.Duration
	// TTLByTool allows different TTLs per tool
	TTLByTool map[string]time.Duration
	// NeverCacheTools lists tools that should never be cached
	NeverCacheTools []string
	// MaxResultSize to cache (bytes)
	MaxResultSize int
}

// DefaultMCPCacheConfig returns sensible defaults
func DefaultMCPCacheConfig() *MCPCacheConfig {
	return &MCPCacheConfig{
		DefaultTTL: 5 * time.Minute,
		TTLByTool: map[string]time.Duration{
			// Filesystem tools - files can change
			"filesystem.read_file": 5 * time.Minute,
			"filesystem.list_dir":  2 * time.Minute,
			"filesystem.search":    5 * time.Minute,
			// GitHub tools - repos change less frequently
			"github.get_repo":          1 * time.Hour,
			"github.list_repos":        30 * time.Minute,
			"github.get_file_contents": 15 * time.Minute,
			"github.search_repos":      30 * time.Minute,
			// Fetch tools - web content changes
			"fetch.fetch": 10 * time.Minute,
			// SQLite tools - data can change
			"sqlite.query": 2 * time.Minute,
			// Puppeteer - screenshots are point-in-time
			"puppeteer.screenshot": 1 * time.Minute,
		},
		NeverCacheTools: []string{
			// Memory tools should never be cached
			"memory.store",
			"memory.retrieve",
			"memory.delete",
			// Write operations should never be cached
			"filesystem.write_file",
			"filesystem.create_dir",
			"filesystem.delete",
			"github.create_file",
			"github.update_file",
			"github.create_issue",
			"github.create_pull_request",
			"sqlite.execute",
		},
		MaxResultSize: 512 * 1024, // 512KB
	}
}

// MCPServerCache caches MCP tool execution results
type MCPServerCache struct {
	cache   *TieredCache
	tagInv  *TagBasedInvalidation
	config  *MCPCacheConfig
	metrics *MCPCacheMetrics
}

// MCPCacheMetrics tracks MCP cache statistics
type MCPCacheMetrics struct {
	Hits              int64
	Misses            int64
	Sets              int64
	Invalidations     int64
	SkippedNeverCache int64
	SkippedLarge      int64
	ByServer          map[string]*mcpServerStats
	ByTool            map[string]*mcpToolStats
}

type mcpServerStats struct {
	Hits   int64
	Misses int64
	Sets   int64
}

type mcpToolStats struct {
	Hits   int64
	Misses int64
	Sets   int64
}

// NewMCPServerCache creates a new MCP server cache
func NewMCPServerCache(cache *TieredCache, config *MCPCacheConfig) *MCPServerCache {
	if config == nil {
		config = DefaultMCPCacheConfig()
	}

	return &MCPServerCache{
		cache:  cache,
		tagInv: NewTagBasedInvalidation(),
		config: config,
		metrics: &MCPCacheMetrics{
			ByServer: make(map[string]*mcpServerStats),
			ByTool:   make(map[string]*mcpToolStats),
		},
	}
}

// CacheKey generates a deterministic cache key for tool execution
func (c *MCPServerCache) CacheKey(server, tool string, args interface{}) string {
	keyData := struct {
		Server string      `json:"server"`
		Tool   string      `json:"tool"`
		Args   interface{} `json:"args"`
	}{
		Server: server,
		Tool:   tool,
		Args:   args,
	}

	data, _ := json.Marshal(keyData)
	hash := sha256.Sum256(data)
	return "mcp:" + server + ":" + tool + ":" + hex.EncodeToString(hash[:12])
}

// GetToolResult retrieves a cached tool execution result
func (c *MCPServerCache) GetToolResult(ctx context.Context, server, tool string, args interface{}) (interface{}, bool) {
	// Check if tool should never be cached
	if c.isNeverCache(tool) {
		atomic.AddInt64(&c.metrics.SkippedNeverCache, 1)
		return nil, false
	}

	key := c.CacheKey(server, tool, args)

	var result interface{}
	found, err := c.cache.Get(ctx, key, &result)
	if err != nil || !found {
		atomic.AddInt64(&c.metrics.Misses, 1)
		c.trackServerMiss(server)
		c.trackToolMiss(tool)
		return nil, false
	}

	atomic.AddInt64(&c.metrics.Hits, 1)
	c.trackServerHit(server)
	c.trackToolHit(tool)
	return result, true
}

// SetToolResult caches a tool execution result
func (c *MCPServerCache) SetToolResult(ctx context.Context, server, tool string, args, result interface{}) error {
	// Check if tool should never be cached
	if c.isNeverCache(tool) {
		atomic.AddInt64(&c.metrics.SkippedNeverCache, 1)
		return nil
	}

	// Check result size
	resultData, _ := json.Marshal(result)
	if len(resultData) > c.config.MaxResultSize {
		atomic.AddInt64(&c.metrics.SkippedLarge, 1)
		return nil
	}

	key := c.CacheKey(server, tool, args)
	ttl := c.getTTL(tool)

	// Add tags for invalidation
	tags := []string{
		"mcp:" + server,
		"mcp-tool:" + tool,
		"mcp-result",
	}

	if err := c.cache.Set(ctx, key, result, ttl, tags...); err != nil {
		return err
	}

	// Track in tag invalidation
	c.tagInv.AddTag(key, tags...)

	atomic.AddInt64(&c.metrics.Sets, 1)
	c.trackServerSet(server)
	c.trackToolSet(tool)
	return nil
}

// InvalidateServer removes all cached results for a server
func (c *MCPServerCache) InvalidateServer(ctx context.Context, server string) (int, error) {
	keys := c.tagInv.InvalidateByTag("mcp:" + server)

	for _, key := range keys {
		c.cache.Delete(ctx, key)
		c.tagInv.RemoveKey(key)
	}

	// Also invalidate by prefix
	count, err := c.cache.InvalidatePrefix(ctx, "mcp:"+server+":")
	if err != nil {
		return len(keys), err
	}

	atomic.AddInt64(&c.metrics.Invalidations, int64(len(keys)+count))
	return len(keys) + count, nil
}

// InvalidateTool removes all cached results for a tool
func (c *MCPServerCache) InvalidateTool(ctx context.Context, tool string) (int, error) {
	keys := c.tagInv.InvalidateByTag("mcp-tool:" + tool)

	for _, key := range keys {
		c.cache.Delete(ctx, key)
		c.tagInv.RemoveKey(key)
	}

	atomic.AddInt64(&c.metrics.Invalidations, int64(len(keys)))
	return len(keys), nil
}

// InvalidateAll removes all cached MCP results
func (c *MCPServerCache) InvalidateAll(ctx context.Context) (int, error) {
	keys := c.tagInv.InvalidateByTag("mcp-result")

	for _, key := range keys {
		c.cache.Delete(ctx, key)
		c.tagInv.RemoveKey(key)
	}

	// Also invalidate by prefix
	count, err := c.cache.InvalidatePrefix(ctx, "mcp:")
	if err != nil {
		return len(keys), err
	}

	atomic.AddInt64(&c.metrics.Invalidations, int64(len(keys)+count))
	return len(keys) + count, nil
}

func (c *MCPServerCache) isNeverCache(tool string) bool {
	for _, t := range c.config.NeverCacheTools {
		if t == tool {
			return true
		}
	}
	return false
}

func (c *MCPServerCache) getTTL(tool string) time.Duration {
	if ttl, ok := c.config.TTLByTool[tool]; ok {
		return ttl
	}
	return c.config.DefaultTTL
}

func (c *MCPServerCache) trackServerHit(server string) {
	if c.metrics.ByServer[server] == nil {
		c.metrics.ByServer[server] = &mcpServerStats{}
	}
	atomic.AddInt64(&c.metrics.ByServer[server].Hits, 1)
}

func (c *MCPServerCache) trackServerMiss(server string) {
	if c.metrics.ByServer[server] == nil {
		c.metrics.ByServer[server] = &mcpServerStats{}
	}
	atomic.AddInt64(&c.metrics.ByServer[server].Misses, 1)
}

func (c *MCPServerCache) trackServerSet(server string) {
	if c.metrics.ByServer[server] == nil {
		c.metrics.ByServer[server] = &mcpServerStats{}
	}
	atomic.AddInt64(&c.metrics.ByServer[server].Sets, 1)
}

func (c *MCPServerCache) trackToolHit(tool string) {
	if c.metrics.ByTool[tool] == nil {
		c.metrics.ByTool[tool] = &mcpToolStats{}
	}
	atomic.AddInt64(&c.metrics.ByTool[tool].Hits, 1)
}

func (c *MCPServerCache) trackToolMiss(tool string) {
	if c.metrics.ByTool[tool] == nil {
		c.metrics.ByTool[tool] = &mcpToolStats{}
	}
	atomic.AddInt64(&c.metrics.ByTool[tool].Misses, 1)
}

func (c *MCPServerCache) trackToolSet(tool string) {
	if c.metrics.ByTool[tool] == nil {
		c.metrics.ByTool[tool] = &mcpToolStats{}
	}
	atomic.AddInt64(&c.metrics.ByTool[tool].Sets, 1)
}

// Metrics returns current metrics
func (c *MCPServerCache) Metrics() *MCPCacheMetrics {
	metrics := &MCPCacheMetrics{
		Hits:              atomic.LoadInt64(&c.metrics.Hits),
		Misses:            atomic.LoadInt64(&c.metrics.Misses),
		Sets:              atomic.LoadInt64(&c.metrics.Sets),
		Invalidations:     atomic.LoadInt64(&c.metrics.Invalidations),
		SkippedNeverCache: atomic.LoadInt64(&c.metrics.SkippedNeverCache),
		SkippedLarge:      atomic.LoadInt64(&c.metrics.SkippedLarge),
		ByServer:          make(map[string]*mcpServerStats),
		ByTool:            make(map[string]*mcpToolStats),
	}

	for server, stats := range c.metrics.ByServer {
		metrics.ByServer[server] = &mcpServerStats{
			Hits:   atomic.LoadInt64(&stats.Hits),
			Misses: atomic.LoadInt64(&stats.Misses),
			Sets:   atomic.LoadInt64(&stats.Sets),
		}
	}

	for tool, stats := range c.metrics.ByTool {
		metrics.ByTool[tool] = &mcpToolStats{
			Hits:   atomic.LoadInt64(&stats.Hits),
			Misses: atomic.LoadInt64(&stats.Misses),
			Sets:   atomic.LoadInt64(&stats.Sets),
		}
	}

	return metrics
}

// HitRate returns the overall hit rate
func (c *MCPServerCache) HitRate() float64 {
	hits := atomic.LoadInt64(&c.metrics.Hits)
	misses := atomic.LoadInt64(&c.metrics.Misses)
	total := hits + misses
	if total == 0 {
		return 0
	}
	return float64(hits) / float64(total) * 100
}

// ServerHitRate returns the hit rate for a specific server
func (c *MCPServerCache) ServerHitRate(server string) float64 {
	stats := c.metrics.ByServer[server]
	if stats == nil {
		return 0
	}

	hits := atomic.LoadInt64(&stats.Hits)
	misses := atomic.LoadInt64(&stats.Misses)
	total := hits + misses
	if total == 0 {
		return 0
	}
	return float64(hits) / float64(total) * 100
}

// ToolHitRate returns the hit rate for a specific tool
func (c *MCPServerCache) ToolHitRate(tool string) float64 {
	stats := c.metrics.ByTool[tool]
	if stats == nil {
		return 0
	}

	hits := atomic.LoadInt64(&stats.Hits)
	misses := atomic.LoadInt64(&stats.Misses)
	total := hits + misses
	if total == 0 {
		return 0
	}
	return float64(hits) / float64(total) * 100
}
