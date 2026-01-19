package modelsdev

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestModelCapabilities_ToCapabilitiesList(t *testing.T) {
	tests := []struct {
		name         string
		capabilities *ModelCapabilities
		expected     []string
	}{
		{
			name:         "nil capabilities",
			capabilities: nil,
			expected:     nil,
		},
		{
			name:         "no capabilities",
			capabilities: &ModelCapabilities{},
			expected:     []string{},
		},
		{
			name: "all capabilities",
			capabilities: &ModelCapabilities{
				Vision:          true,
				FunctionCalling: true,
				Streaming:       true,
				JSONMode:        true,
				ImageGeneration: true,
				Audio:           true,
				CodeGeneration:  true,
				Reasoning:       true,
				ToolUse:         true,
			},
			expected: []string{"vision", "function_calling", "streaming", "json_mode", "image_generation", "audio", "code_generation", "reasoning", "tool_use"},
		},
		{
			name: "some capabilities",
			capabilities: &ModelCapabilities{
				Vision:    true,
				Streaming: true,
			},
			expected: []string{"vision", "streaming"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.capabilities.ToCapabilitiesList()
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.ElementsMatch(t, tt.expected, result)
			}
		})
	}
}

func TestModelCapabilities_HasCapability(t *testing.T) {
	tests := []struct {
		name         string
		capabilities *ModelCapabilities
		capability   string
		expected     bool
	}{
		{
			name:         "nil capabilities",
			capabilities: nil,
			capability:   "vision",
			expected:     false,
		},
		{
			name:         "has vision",
			capabilities: &ModelCapabilities{Vision: true},
			capability:   "vision",
			expected:     true,
		},
		{
			name:         "no vision",
			capabilities: &ModelCapabilities{Vision: false},
			capability:   "vision",
			expected:     false,
		},
		{
			name:         "has function_calling",
			capabilities: &ModelCapabilities{FunctionCalling: true},
			capability:   "function_calling",
			expected:     true,
		},
		{
			name:         "has streaming",
			capabilities: &ModelCapabilities{Streaming: true},
			capability:   "streaming",
			expected:     true,
		},
		{
			name:         "has json_mode",
			capabilities: &ModelCapabilities{JSONMode: true},
			capability:   "json_mode",
			expected:     true,
		},
		{
			name:         "has image_generation",
			capabilities: &ModelCapabilities{ImageGeneration: true},
			capability:   "image_generation",
			expected:     true,
		},
		{
			name:         "has audio",
			capabilities: &ModelCapabilities{Audio: true},
			capability:   "audio",
			expected:     true,
		},
		{
			name:         "has code_generation",
			capabilities: &ModelCapabilities{CodeGeneration: true},
			capability:   "code_generation",
			expected:     true,
		},
		{
			name:         "has reasoning",
			capabilities: &ModelCapabilities{Reasoning: true},
			capability:   "reasoning",
			expected:     true,
		},
		{
			name:         "has tool_use",
			capabilities: &ModelCapabilities{ToolUse: true},
			capability:   "tool_use",
			expected:     true,
		},
		{
			name:         "unknown capability",
			capabilities: &ModelCapabilities{Vision: true},
			capability:   "unknown",
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.capabilities.HasCapability(tt.capability)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestModel_GetDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		model    Model
		expected string
	}{
		{
			name:     "has display name",
			model:    Model{Name: "test-name", DisplayName: "Test Display Name"},
			expected: "Test Display Name",
		},
		{
			name:     "no display name",
			model:    Model{Name: "test-name"},
			expected: "test-name",
		},
		{
			name:     "empty display name",
			model:    Model{Name: "test-name", DisplayName: ""},
			expected: "test-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.model.GetDisplayName()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProvider_GetDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		provider Provider
		expected string
	}{
		{
			name:     "has display name",
			provider: Provider{Name: "test-name", DisplayName: "Test Display Name"},
			expected: "Test Display Name",
		},
		{
			name:     "no display name",
			provider: Provider{Name: "test-name"},
			expected: "test-name",
		},
		{
			name:     "empty display name",
			provider: Provider{Name: "test-name", DisplayName: ""},
			expected: "test-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.provider.GetDisplayName()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Note: TestCachedModel_IsExpired and TestCachedProvider_IsExpired are in cache_test.go

func TestDefaultCacheConfig(t *testing.T) {
	config := DefaultCacheConfig()

	assert.Equal(t, 1*time.Hour, config.ModelTTL)
	assert.Equal(t, 2*time.Hour, config.ProviderTTL)
	assert.Equal(t, 5000, config.MaxModels)
	assert.Equal(t, 100, config.MaxProviders)
	assert.Equal(t, 10*time.Minute, config.CleanupInterval)
}

func TestDefaultServiceConfig(t *testing.T) {
	config := DefaultServiceConfig()

	assert.NotEmpty(t, config.Client.BaseURL)
	assert.True(t, config.RefreshOnStart)
	assert.True(t, config.AutoRefresh)
	assert.Equal(t, 24*time.Hour, config.RefreshInterval)

	// Check cache defaults
	assert.Equal(t, 1*time.Hour, config.Cache.ModelTTL)
	assert.Equal(t, 2*time.Hour, config.Cache.ProviderTTL)
}

func TestDefaultClientConfig(t *testing.T) {
	config := DefaultClientConfig()

	assert.Equal(t, DefaultBaseURL, config.BaseURL)
	assert.Equal(t, DefaultTimeout, config.Timeout)
	assert.Equal(t, DefaultUserAgent, config.UserAgent)
}

func TestModelFilters(t *testing.T) {
	hasVision := true
	hasStreaming := false

	filters := ModelFilters{
		Provider:     "anthropic",
		Family:       "claude",
		Category:     "chat",
		Tags:         []string{"ai", "llm"},
		MinContext:   1000,
		MaxContext:   200000,
		HasVision:    &hasVision,
		HasStreaming: &hasStreaming,
		HasTools:     nil,
		Page:         1,
		Limit:        50,
		SortBy:       "popularity",
		SortOrder:    "desc",
	}

	assert.Equal(t, "anthropic", filters.Provider)
	assert.Equal(t, "claude", filters.Family)
	assert.True(t, *filters.HasVision)
	assert.False(t, *filters.HasStreaming)
	assert.Nil(t, filters.HasTools)
	assert.Equal(t, 50, filters.Limit)
}

func TestRefreshResult(t *testing.T) {
	result := RefreshResult{
		ModelsRefreshed:    100,
		ProvidersRefreshed: 10,
		Errors:             []string{"error1", "error2"},
		Duration:           5 * time.Second,
		Success:            false,
	}

	assert.Equal(t, 100, result.ModelsRefreshed)
	assert.Equal(t, 10, result.ProvidersRefreshed)
	assert.Len(t, result.Errors, 2)
	assert.Equal(t, 5*time.Second, result.Duration)
	assert.False(t, result.Success)
}

func TestCacheStats(t *testing.T) {
	stats := CacheStats{
		ModelCount:       1000,
		ProviderCount:    50,
		TotalHits:        5000,
		TotalMisses:      500,
		HitRate:          0.909,
		LastRefresh:      time.Now(),
		OldestEntry:      time.Now().Add(-1 * time.Hour),
		MemoryUsageBytes: 1024 * 1024,
	}

	assert.Equal(t, 1000, stats.ModelCount)
	assert.Equal(t, 50, stats.ProviderCount)
	assert.Equal(t, int64(5000), stats.TotalHits)
	assert.Equal(t, int64(500), stats.TotalMisses)
	assert.InDelta(t, 0.909, stats.HitRate, 0.001)
}

func TestPricing(t *testing.T) {
	pricing := Pricing{
		InputCost:       3.0,
		OutputCost:      15.0,
		Currency:        "USD",
		Unit:            "per_million_tokens",
		CachedInputCost: 1.5,
	}

	assert.Equal(t, 3.0, pricing.InputCost)
	assert.Equal(t, 15.0, pricing.OutputCost)
	assert.Equal(t, "USD", pricing.Currency)
	assert.Equal(t, "per_million_tokens", pricing.Unit)
	assert.Equal(t, 1.5, pricing.CachedInputCost)
}
