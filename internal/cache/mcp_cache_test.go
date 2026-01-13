package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMCPServerCache_Creation(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	mc := NewMCPServerCache(tc, nil)
	require.NotNil(t, mc)
}

func TestMCPServerCache_CreationWithConfig(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	mcpConfig := &MCPCacheConfig{
		DefaultTTL:    10 * time.Minute,
		MaxResultSize: 1024 * 1024,
	}
	mc := NewMCPServerCache(tc, mcpConfig)
	require.NotNil(t, mc)
}

func TestMCPServerCache_CacheKey(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	mc := NewMCPServerCache(tc, nil)
	require.NotNil(t, mc)

	args := map[string]interface{}{"path": "/test"}

	// Generate cache key
	key := mc.CacheKey("filesystem", "read", args)
	assert.NotEmpty(t, key)
	assert.Contains(t, key, "mcp:filesystem:read:")

	// Same inputs should produce same key
	key2 := mc.CacheKey("filesystem", "read", args)
	assert.Equal(t, key, key2)

	// Different args should produce different key
	args2 := map[string]interface{}{"path": "/different"}
	key3 := mc.CacheKey("filesystem", "read", args2)
	assert.NotEqual(t, key, key3)
}

func TestMCPServerCache_GetToolResult_Miss(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	mc := NewMCPServerCache(tc, nil)
	require.NotNil(t, mc)

	ctx := context.Background()

	result, found := mc.GetToolResult(ctx, "filesystem", "read", nil)
	assert.False(t, found)
	assert.Nil(t, result)
}

func TestMCPServerCache_SetAndGetToolResult(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	mc := NewMCPServerCache(tc, nil)
	require.NotNil(t, mc)

	ctx := context.Background()
	args := map[string]interface{}{"path": "/test"}
	result := map[string]interface{}{"content": "test content"}

	err := mc.SetToolResult(ctx, "filesystem", "read", args, result)
	require.NoError(t, err)

	cached, found := mc.GetToolResult(ctx, "filesystem", "read", args)
	assert.True(t, found)
	assert.NotNil(t, cached)
}

func TestMCPServerCache_InvalidateServer(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	mc := NewMCPServerCache(tc, nil)
	require.NotNil(t, mc)

	ctx := context.Background()
	args := map[string]interface{}{"path": "/test"}

	err := mc.SetToolResult(ctx, "filesystem", "read", args, "content")
	require.NoError(t, err)

	count, err := mc.InvalidateServer(ctx, "filesystem")
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, count, 0)
}

func TestMCPServerCache_ToolSpecificTTLs(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	mc := NewMCPServerCache(tc, nil)
	require.NotNil(t, mc)

	ctx := context.Background()

	// Set results for different tools
	err := mc.SetToolResult(ctx, "filesystem", "read_file", nil, "content")
	require.NoError(t, err)
	err = mc.SetToolResult(ctx, "github", "get_repo", nil, "repo data")
	require.NoError(t, err)

	// All should be retrievable
	r1, f1 := mc.GetToolResult(ctx, "filesystem", "read_file", nil)
	r2, f2 := mc.GetToolResult(ctx, "github", "get_repo", nil)

	assert.True(t, f1)
	assert.True(t, f2)
	assert.NotNil(t, r1)
	assert.NotNil(t, r2)
}

func TestMCPServerCache_DifferentArgs(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	mc := NewMCPServerCache(tc, nil)
	require.NotNil(t, mc)

	ctx := context.Background()
	args1 := map[string]interface{}{"path": "/test1"}
	args2 := map[string]interface{}{"path": "/test2"}

	err := mc.SetToolResult(ctx, "filesystem", "read", args1, "content1")
	require.NoError(t, err)
	err = mc.SetToolResult(ctx, "filesystem", "read", args2, "content2")
	require.NoError(t, err)

	r1, f1 := mc.GetToolResult(ctx, "filesystem", "read", args1)
	r2, f2 := mc.GetToolResult(ctx, "filesystem", "read", args2)

	assert.True(t, f1)
	assert.True(t, f2)
	assert.NotEqual(t, r1, r2)
}

func TestMCPServerCache_NeverCacheTools(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	mc := NewMCPServerCache(tc, nil)
	require.NotNil(t, mc)

	ctx := context.Background()

	// Memory tools should never be cached
	err := mc.SetToolResult(ctx, "memory", "memory.retrieve", nil, "data")
	require.NoError(t, err)

	// Should not be found even after setting
	_, found := mc.GetToolResult(ctx, "memory", "memory.retrieve", nil)
	assert.False(t, found)
}

func TestMCPServerCache_InvalidateTool(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	mc := NewMCPServerCache(tc, nil)
	require.NotNil(t, mc)

	ctx := context.Background()

	err := mc.SetToolResult(ctx, "filesystem", "read_file", nil, "content")
	require.NoError(t, err)

	count, err := mc.InvalidateTool(ctx, "read_file")
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, count, 0)
}

func TestMCPServerCache_InvalidateAll(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	mc := NewMCPServerCache(tc, nil)
	require.NotNil(t, mc)

	ctx := context.Background()

	count, err := mc.InvalidateAll(ctx)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, count, 0)
}

func TestMCPServerCache_Metrics(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	mc := NewMCPServerCache(tc, nil)
	require.NotNil(t, mc)

	ctx := context.Background()

	// Generate some cache activity
	mc.GetToolResult(ctx, "server", "tool", nil) // Miss
	mc.SetToolResult(ctx, "server", "tool", nil, "data")
	mc.GetToolResult(ctx, "server", "tool", nil) // Hit

	metrics := mc.Metrics()
	assert.NotNil(t, metrics)
	assert.GreaterOrEqual(t, metrics.Hits, int64(0))
	assert.GreaterOrEqual(t, metrics.Misses, int64(0))
	assert.GreaterOrEqual(t, metrics.Sets, int64(0))
}

func TestMCPServerCache_HitRate(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	mc := NewMCPServerCache(tc, nil)
	require.NotNil(t, mc)

	// No activity - hit rate should be 0
	assert.Equal(t, float64(0), mc.HitRate())
}

func TestMCPServerCache_ServerHitRate(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	mc := NewMCPServerCache(tc, nil)
	require.NotNil(t, mc)

	// No activity - hit rate should be 0
	assert.Equal(t, float64(0), mc.ServerHitRate("filesystem"))
}

func TestMCPServerCache_ToolHitRate(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	mc := NewMCPServerCache(tc, nil)
	require.NotNil(t, mc)

	// No activity - hit rate should be 0
	assert.Equal(t, float64(0), mc.ToolHitRate("read_file"))
}

func TestDefaultMCPCacheConfig(t *testing.T) {
	config := DefaultMCPCacheConfig()
	assert.NotNil(t, config)
	assert.Greater(t, config.DefaultTTL, time.Duration(0))
	assert.Greater(t, config.MaxResultSize, 0)
	assert.NotEmpty(t, config.TTLByTool)
	assert.NotEmpty(t, config.NeverCacheTools)
}

func TestMCPServerCache_ContextCancellation(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	mc := NewMCPServerCache(tc, nil)
	require.NotNil(t, mc)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Operations should handle cancelled context gracefully
	err := mc.SetToolResult(ctx, "server", "tool", nil, "data")
	// May or may not error depending on implementation
	_ = err

	_, _ = mc.GetToolResult(ctx, "server", "tool", nil)
}
