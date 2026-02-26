package services

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/models"
)

func TestDefaultDebateOptimizationConfig(t *testing.T) {
	config := DefaultDebateOptimizationConfig()

	assert.True(t, config.EnableParallelExecution)
	assert.True(t, config.EnableResponseCaching)
	assert.True(t, config.EnableEarlyTermination)
	assert.True(t, config.EnableStreaming)
	assert.True(t, config.EnableSmartFallback)
	assert.Equal(t, 3, config.MaxParallelRequests)
	assert.Equal(t, 5*time.Minute, config.CacheTTL)
	assert.Equal(t, 0.95, config.EarlyTerminationThreshold)
	assert.Equal(t, 60*time.Second, config.RequestTimeout)
	assert.Equal(t, 30*time.Second, config.FallbackTimeout)
}

func TestNewDebatePerformanceOptimizer(t *testing.T) {
	logger := logrus.New()
	registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)
	config := DefaultDebateOptimizationConfig()

	optimizer := NewDebatePerformanceOptimizer(config, registry, logger)

	assert.NotNil(t, optimizer)
	assert.NotNil(t, optimizer.semaphore)
	assert.NotNil(t, optimizer.stats)
	assert.Equal(t, 3, cap(optimizer.semaphore))
}

func TestNewDebatePerformanceOptimizer_NilLogger(t *testing.T) {
	registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)
	config := DefaultDebateOptimizationConfig()

	optimizer := NewDebatePerformanceOptimizer(config, registry, nil)

	assert.NotNil(t, optimizer)
}

func TestDebatePerformanceOptimizer_GetStats(t *testing.T) {
	logger := logrus.New()
	registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)
	config := DefaultDebateOptimizationConfig()
	optimizer := NewDebatePerformanceOptimizer(config, registry, logger)

	stats := optimizer.GetStats()

	assert.NotNil(t, stats)
	assert.Equal(t, int64(0), stats.TotalRequests)
	assert.Equal(t, int64(0), stats.CacheHits)
	assert.Equal(t, int64(0), stats.CacheMisses)
}

func TestDebatePerformanceOptimizer_ClearCache(t *testing.T) {
	logger := logrus.New()
	registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)
	config := DefaultDebateOptimizationConfig()
	optimizer := NewDebatePerformanceOptimizer(config, registry, logger)

	optimizer.cache.Store("test-key", &CachedResponse{
		Response:  &models.LLMResponse{Content: "test"},
		Timestamp: time.Now(),
		Model:     "test-model",
		Provider:  "test-provider",
	})

	optimizer.ClearCache()

	_, exists := optimizer.cache.Load("test-key")
	assert.False(t, exists)
}

func TestDebatePerformanceOptimizer_CacheResponse(t *testing.T) {
	logger := logrus.New()
	registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)
	config := DefaultDebateOptimizationConfig()
	optimizer := NewDebatePerformanceOptimizer(config, registry, logger)

	response := &models.LLMResponse{Content: "test response"}
	optimizer.cacheResponse("test prompt", "test-model", "test-provider", response)

	cached := optimizer.getCachedResponse("test prompt", "test-model")
	require.NotNil(t, cached)
	assert.Equal(t, "test response", cached.Response.Content)
	assert.Equal(t, "test-model", cached.Model)
	assert.Equal(t, "test-provider", cached.Provider)
}

func TestDebatePerformanceOptimizer_CacheExpiry(t *testing.T) {
	logger := logrus.New()
	registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)
	config := DebateOptimizationConfig{
		EnableResponseCaching: true,
		CacheTTL:              100 * time.Millisecond,
	}
	optimizer := NewDebatePerformanceOptimizer(config, registry, logger)

	response := &models.LLMResponse{Content: "test response"}
	optimizer.cacheResponse("test prompt", "test-model", "test-provider", response)

	cached := optimizer.getCachedResponse("test prompt", "test-model")
	require.NotNil(t, cached)

	time.Sleep(150 * time.Millisecond)

	cached = optimizer.getCachedResponse("test prompt", "test-model")
	assert.Nil(t, cached)
}

func TestDebatePerformanceOptimizer_ShouldTerminateEarly_Disabled(t *testing.T) {
	logger := logrus.New()
	registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)
	config := DebateOptimizationConfig{
		EnableEarlyTermination:    false,
		EarlyTerminationThreshold: 0.95,
	}
	optimizer := NewDebatePerformanceOptimizer(config, registry, logger)

	responses := map[DebateTeamPosition]string{
		PositionAnalyst:  "This is a consensus response with agreement",
		PositionProposer: "This is a consensus response with agreement",
		PositionCritic:   "This is a consensus response with agreement",
	}

	result := optimizer.ShouldTerminateEarly(responses)
	assert.False(t, result)
}

func TestDebatePerformanceOptimizer_ShouldTerminateEarly_InsufficientResponses(t *testing.T) {
	logger := logrus.New()
	registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)
	config := DefaultDebateOptimizationConfig()
	optimizer := NewDebatePerformanceOptimizer(config, registry, logger)

	responses := map[DebateTeamPosition]string{
		PositionAnalyst:  "Response 1",
		PositionProposer: "Response 2",
	}

	result := optimizer.ShouldTerminateEarly(responses)
	assert.False(t, result)
}

func TestDebatePerformanceOptimizer_ShouldTerminateEarly_HighConsensus(t *testing.T) {
	logger := logrus.New()
	registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)
	config := DebateOptimizationConfig{
		EnableEarlyTermination:    true,
		EarlyTerminationThreshold: 0.5,
	}
	optimizer := NewDebatePerformanceOptimizer(config, registry, logger)

	consensusText := "The quick brown fox jumps over the lazy dog repeatedly for testing consensus"
	responses := map[DebateTeamPosition]string{
		PositionAnalyst:  consensusText,
		PositionProposer: consensusText,
		PositionCritic:   consensusText,
	}

	result := optimizer.ShouldTerminateEarly(responses)
	assert.True(t, result)
	assert.Equal(t, int64(1), optimizer.GetStats().EarlyTerminations)
}

func TestDebatePerformanceOptimizer_ShouldTerminateEarly_LowConsensus(t *testing.T) {
	logger := logrus.New()
	registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)
	config := DebateOptimizationConfig{
		EnableEarlyTermination:    true,
		EarlyTerminationThreshold: 0.95,
	}
	optimizer := NewDebatePerformanceOptimizer(config, registry, logger)

	responses := map[DebateTeamPosition]string{
		PositionAnalyst:  "Completely different response about apples",
		PositionProposer: "Another unique response about bananas",
		PositionCritic:   "Yet another distinct response about cherries",
	}

	result := optimizer.ShouldTerminateEarly(responses)
	assert.False(t, result)
}

func TestDebatePerformanceOptimizer_ExecuteParallel(t *testing.T) {
	logger := logrus.New()
	registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)
	config := DefaultDebateOptimizationConfig()
	optimizer := NewDebatePerformanceOptimizer(config, registry, logger)

	members := []*DebateTeamMember{
		{
			Position:     PositionAnalyst,
			ProviderName: "test-provider",
			ModelName:    "test-model-1",
			Provider:     &mockProviderForOptimizer{response: "response 1"},
		},
		{
			Position:     PositionProposer,
			ProviderName: "test-provider",
			ModelName:    "test-model-2",
			Provider:     &mockProviderForOptimizer{response: "response 2"},
		},
	}

	results := optimizer.ExecuteParallel(context.Background(), members, "test prompt")

	assert.Len(t, results, 2)
	assert.Equal(t, int64(2), optimizer.GetStats().ParallelRequests)
}

func TestDebatePerformanceOptimizer_ExecuteWithOptimization_CacheHit(t *testing.T) {
	logger := logrus.New()
	registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)
	config := DefaultDebateOptimizationConfig()
	optimizer := NewDebatePerformanceOptimizer(config, registry, logger)

	cachedResponse := &models.LLMResponse{Content: "cached response"}
	optimizer.cacheResponse("test prompt", "test-model", "test-provider", cachedResponse)

	member := &DebateTeamMember{
		Position:     PositionAnalyst,
		ProviderName: "test-provider",
		ModelName:    "test-model",
	}

	response, err := optimizer.ExecuteWithOptimization(
		context.Background(),
		member,
		"test prompt",
		nil,
	)

	require.NoError(t, err)
	assert.Equal(t, "cached response", response.Content)
	assert.Equal(t, int64(1), optimizer.GetStats().CacheHits)
}

func TestDebatePerformanceOptimizer_ExecuteWithOptimization_CacheMiss(t *testing.T) {
	logger := logrus.New()
	registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)
	config := DefaultDebateOptimizationConfig()
	optimizer := NewDebatePerformanceOptimizer(config, registry, logger)

	member := &DebateTeamMember{
		Position:     PositionAnalyst,
		ProviderName: "test-provider",
		ModelName:    "test-model",
		Provider:     &mockProviderForOptimizer{response: "fresh response"},
	}

	response, err := optimizer.ExecuteWithOptimization(
		context.Background(),
		member,
		"unique prompt",
		nil,
	)

	require.NoError(t, err)
	assert.Equal(t, "fresh response", response.Content)
	assert.Equal(t, int64(1), optimizer.GetStats().CacheMisses)
	assert.Equal(t, int64(1), optimizer.GetStats().TotalRequests)
}

func TestDebatePerformanceOptimizer_ExecuteWithSmartFallback_PrimarySuccess(t *testing.T) {
	logger := logrus.New()
	registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)
	config := DebateOptimizationConfig{
		EnableSmartFallback:   true,
		EnableResponseCaching: false,
		RequestTimeout:        5 * time.Second,
	}
	optimizer := NewDebatePerformanceOptimizer(config, registry, logger)

	member := &DebateTeamMember{
		Position:     PositionAnalyst,
		ProviderName: "test-provider",
		ModelName:    "test-model",
		Provider:     &mockProviderForOptimizer{response: "primary response"},
		Fallbacks: []*DebateTeamMember{
			{
				ProviderName: "fallback-provider",
				ModelName:    "fallback-model",
				Provider:     &mockProviderForOptimizer{response: "fallback response"},
				IsActive:     true,
			},
		},
	}

	response, err := optimizer.ExecuteWithOptimization(
		context.Background(),
		member,
		"test prompt",
		nil,
	)

	require.NoError(t, err)
	assert.Equal(t, "primary response", response.Content)
	assert.Equal(t, int64(0), optimizer.GetStats().FallbacksTriggered)
}

func TestDebatePerformanceOptimizer_ExecuteWithSmartFallback_FallbackUsed(t *testing.T) {
	logger := logrus.New()
	registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)
	config := DebateOptimizationConfig{
		EnableSmartFallback:   true,
		EnableResponseCaching: false,
		RequestTimeout:        5 * time.Second,
		FallbackTimeout:       5 * time.Second,
	}
	optimizer := NewDebatePerformanceOptimizer(config, registry, logger)

	member := &DebateTeamMember{
		Position:     PositionAnalyst,
		ProviderName: "test-provider",
		ModelName:    "test-model",
		Provider:     &mockProviderForOptimizer{err: assert.AnError},
		Fallbacks: []*DebateTeamMember{
			{
				ProviderName: "fallback-provider",
				ModelName:    "fallback-model",
				Provider:     &mockProviderForOptimizer{response: "fallback response"},
				IsActive:     true,
			},
		},
	}

	response, err := optimizer.ExecuteWithOptimization(
		context.Background(),
		member,
		"test prompt",
		nil,
	)

	require.NoError(t, err)
	assert.Equal(t, "fallback response", response.Content)
	assert.Equal(t, int64(1), optimizer.GetStats().FallbacksTriggered)
}

func TestDebatePerformanceOptimizer_ExecuteWithSmartFallback_AllFail(t *testing.T) {
	logger := logrus.New()
	registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)
	config := DebateOptimizationConfig{
		EnableSmartFallback:   true,
		EnableResponseCaching: false,
		RequestTimeout:        5 * time.Second,
		FallbackTimeout:       5 * time.Second,
	}
	optimizer := NewDebatePerformanceOptimizer(config, registry, logger)

	member := &DebateTeamMember{
		Position:     PositionAnalyst,
		ProviderName: "test-provider",
		ModelName:    "test-model",
		Provider:     &mockProviderForOptimizer{err: assert.AnError},
		Fallbacks: []*DebateTeamMember{
			{
				ProviderName: "fallback-provider",
				ModelName:    "fallback-model",
				Provider:     &mockProviderForOptimizer{err: assert.AnError},
				IsActive:     true,
			},
		},
	}

	response, err := optimizer.ExecuteWithOptimization(
		context.Background(),
		member,
		"test prompt",
		nil,
	)

	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple words",
			input:    "hello world",
			expected: []string{"hello", "world"},
		},
		{
			name:     "with punctuation",
			input:    "Hello, World!",
			expected: []string{"hello", "world"},
		},
		{
			name:     "with numbers",
			input:    "test123 and more",
			expected: []string{"test123", "and", "more"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "only punctuation",
			input:    "!@#$%",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tokenize(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHashString(t *testing.T) {
	hash1 := hashString("test string")
	hash2 := hashString("test string")
	hash3 := hashString("different string")

	assert.Equal(t, hash1, hash2, "same input should produce same hash")
	assert.NotEqual(t, hash1, hash3, "different input should produce different hash")
	assert.Len(t, hash1, 8, "hash should be 8 characters")
}

func TestDebatePerformanceOptimizer_ConcurrentAccess(t *testing.T) {
	logger := logrus.New()
	registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)
	config := DefaultDebateOptimizationConfig()
	optimizer := NewDebatePerformanceOptimizer(config, registry, logger)

	var wg sync.WaitGroup
	numOps := 100

	for i := 0; i < numOps; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			prompt := fmt.Sprintf("prompt-%d", idx%10)
			optimizer.cacheResponse(prompt, "model", "provider", &models.LLMResponse{Content: "response"})
			_ = optimizer.getCachedResponse(prompt, "model")
			_ = optimizer.GetStats()
		}(i)
	}

	wg.Wait()

	stats := optimizer.GetStats()
	assert.NotNil(t, stats, "Should be able to get stats after concurrent access")
}

func TestDebatePerformanceOptimizer_GenerateCacheKey(t *testing.T) {
	logger := logrus.New()
	registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)
	config := DefaultDebateOptimizationConfig()
	optimizer := NewDebatePerformanceOptimizer(config, registry, logger)

	key1 := optimizer.generateCacheKey("prompt1", "model1")
	key2 := optimizer.generateCacheKey("prompt1", "model2")
	key3 := optimizer.generateCacheKey("prompt2", "model1")

	assert.NotEqual(t, key1, key2, "different models should have different keys")
	assert.NotEqual(t, key1, key3, "different prompts should have different keys")

	assert.Contains(t, key1, "model1:")
}

type mockProviderForOptimizer struct {
	response string
	err      error
}

func (m *mockProviderForOptimizer) Complete(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &models.LLMResponse{Content: m.response}, nil
}

func (m *mockProviderForOptimizer) CompleteStream(ctx context.Context, req *models.LLMRequest) (<-chan *models.LLMResponse, error) {
	return nil, nil
}

func (m *mockProviderForOptimizer) HealthCheck() error {
	return nil
}

func (m *mockProviderForOptimizer) GetCapabilities() *models.ProviderCapabilities {
	return &models.ProviderCapabilities{}
}

func (m *mockProviderForOptimizer) ValidateConfig(config map[string]interface{}) (bool, []string) {
	return true, nil
}
