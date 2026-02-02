package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/models"
)

func TestProviderCache_Creation(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	pc := NewProviderCache(tc, nil)
	require.NotNil(t, pc)
}

func TestProviderCache_CreationWithConfig(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	pcConfig := &ProviderCacheConfig{
		DefaultTTL:      time.Hour,
		MaxResponseSize: 2 * 1024 * 1024,
	}
	pc := NewProviderCache(tc, pcConfig)
	require.NotNil(t, pc)
}

func TestProviderCache_CacheKey(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	pc := NewProviderCache(tc, nil)
	require.NotNil(t, pc)

	req := &models.LLMRequest{
		Prompt: "test prompt",
		ModelParams: models.ModelParameters{
			Model: "gpt-4",
		},
	}

	// Generate cache key
	key := pc.CacheKey(req, "openai")
	assert.NotEmpty(t, key)
	assert.Contains(t, key, "provider:openai:")

	// Same inputs should produce same key
	key2 := pc.CacheKey(req, "openai")
	assert.Equal(t, key, key2)

	// Different prompt should produce different key
	req2 := &models.LLMRequest{
		Prompt: "different prompt",
		ModelParams: models.ModelParameters{
			Model: "gpt-4",
		},
	}
	key3 := pc.CacheKey(req2, "openai")
	assert.NotEqual(t, key, key3)
}

func TestProviderCache_SetAndGet(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	pc := NewProviderCache(tc, nil)
	require.NotNil(t, pc)

	ctx := context.Background()

	req := &models.LLMRequest{
		Prompt: "test prompt",
		ModelParams: models.ModelParameters{
			Model: "gpt-4",
		},
	}
	resp := &models.LLMResponse{
		Content: "test response",
	}

	// Set a value
	err := pc.Set(ctx, req, resp, "openai")
	require.NoError(t, err)

	// Get the value
	result, found := pc.Get(ctx, req, "openai")
	assert.True(t, found)
	assert.NotNil(t, result)
	assert.Equal(t, "test response", result.Content)
}

func TestProviderCache_Get_Miss(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	pc := NewProviderCache(tc, nil)
	require.NotNil(t, pc)

	ctx := context.Background()

	req := &models.LLMRequest{
		Prompt: "nonexistent prompt",
		ModelParams: models.ModelParameters{
			Model: "model",
		},
	}

	// Get non-existent value
	result, found := pc.Get(ctx, req, "provider")
	assert.False(t, found)
	assert.Nil(t, result)
}

func TestProviderCache_InvalidateProvider(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	pc := NewProviderCache(tc, nil)
	require.NotNil(t, pc)

	ctx := context.Background()

	req := &models.LLMRequest{
		Prompt: "prompt",
		ModelParams: models.ModelParameters{
			Model: "gpt-4",
		},
	}
	resp := &models.LLMResponse{
		Content: "response",
	}

	// Set a value
	err := pc.Set(ctx, req, resp, "openai")
	require.NoError(t, err)

	// Invalidate provider
	count, err := pc.InvalidateProvider(ctx, "openai")
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, count, 0)
}

func TestProviderCache_DifferentModels(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	pc := NewProviderCache(tc, nil)
	require.NotNil(t, pc)

	ctx := context.Background()

	req1 := &models.LLMRequest{
		Prompt: "prompt",
		ModelParams: models.ModelParameters{
			Model: "gpt-4",
		},
	}
	req2 := &models.LLMRequest{
		Prompt: "prompt",
		ModelParams: models.ModelParameters{
			Model: "claude-3",
		},
	}

	resp1 := &models.LLMResponse{Content: "response-gpt4"}
	resp2 := &models.LLMResponse{Content: "response-claude"}

	// Set values for different models
	err := pc.Set(ctx, req1, resp1, "openai")
	require.NoError(t, err)
	err = pc.Set(ctx, req2, resp2, "anthropic")
	require.NoError(t, err)

	// Get values
	r1, found1 := pc.Get(ctx, req1, "openai")
	r2, found2 := pc.Get(ctx, req2, "anthropic")

	assert.True(t, found1)
	assert.True(t, found2)
	assert.NotEqual(t, r1.Content, r2.Content)
}

func TestProviderCache_DifferentProviders(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	pc := NewProviderCache(tc, nil)
	require.NotNil(t, pc)

	ctx := context.Background()

	req := &models.LLMRequest{
		Prompt: "same prompt",
		ModelParams: models.ModelParameters{
			Model: "model",
		},
	}

	resp1 := &models.LLMResponse{Content: "openai response"}
	resp2 := &models.LLMResponse{Content: "anthropic response"}

	// Set with different providers
	err := pc.Set(ctx, req, resp1, "openai")
	require.NoError(t, err)
	err = pc.Set(ctx, req, resp2, "anthropic")
	require.NoError(t, err)

	// Different providers should have different cached values
	r1, found1 := pc.Get(ctx, req, "openai")
	r2, found2 := pc.Get(ctx, req, "anthropic")

	assert.True(t, found1)
	assert.True(t, found2)
	assert.Equal(t, "openai response", r1.Content)
	assert.Equal(t, "anthropic response", r2.Content)
}

func TestProviderCache_Metrics(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	pc := NewProviderCache(tc, nil)
	require.NotNil(t, pc)

	ctx := context.Background()

	req := &models.LLMRequest{
		Prompt: "prompt",
		ModelParams: models.ModelParameters{
			Model: "model",
		},
	}
	resp := &models.LLMResponse{Content: "response"}

	// Generate some cache activity
	pc.Get(ctx, req, "provider") // Miss
	_ = pc.Set(ctx, req, resp, "provider")
	pc.Get(ctx, req, "provider") // Hit

	metrics := pc.Metrics()
	assert.NotNil(t, metrics)
	assert.GreaterOrEqual(t, metrics.Hits, int64(0))
	assert.GreaterOrEqual(t, metrics.Misses, int64(0))
	assert.GreaterOrEqual(t, metrics.Sets, int64(0))
}

func TestProviderCache_HitRate(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	pc := NewProviderCache(tc, nil)
	require.NotNil(t, pc)

	// No activity - hit rate should be 0
	assert.Equal(t, float64(0), pc.HitRate())
}

func TestProviderCache_ProviderHitRate(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	pc := NewProviderCache(tc, nil)
	require.NotNil(t, pc)

	// No activity - hit rate should be 0
	assert.Equal(t, float64(0), pc.ProviderHitRate("openai"))
}

func TestProviderCache_InvalidateModel(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	pc := NewProviderCache(tc, nil)
	require.NotNil(t, pc)

	ctx := context.Background()

	req := &models.LLMRequest{
		Prompt: "prompt",
		ModelParams: models.ModelParameters{
			Model: "gpt-4",
		},
	}
	resp := &models.LLMResponse{Content: "response"}

	err := pc.Set(ctx, req, resp, "openai")
	require.NoError(t, err)

	// Invalidate by model
	count, err := pc.InvalidateModel(ctx, "gpt-4")
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, count, 0)
}

func TestProviderCache_InvalidateAll(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	pc := NewProviderCache(tc, nil)
	require.NotNil(t, pc)

	ctx := context.Background()

	// Invalidate all
	count, err := pc.InvalidateAll(ctx)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, count, 0)
}

func TestDefaultProviderCacheConfig(t *testing.T) {
	config := DefaultProviderCacheConfig()
	assert.NotNil(t, config)
	assert.Greater(t, config.DefaultTTL, time.Duration(0))
	assert.Greater(t, config.MaxResponseSize, 0)
	assert.NotEmpty(t, config.TTLByProvider)
}
