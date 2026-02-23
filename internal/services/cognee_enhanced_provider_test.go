package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"dev.helix.agent/internal/llm"
	"dev.helix.agent/internal/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CogneeMockProvider implements llm.LLMProvider for Cognee testing
type CogneeMockProvider struct {
	completeFunc       func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error)
	completeStreamFunc func(ctx context.Context, req *models.LLMRequest) (<-chan *models.LLMResponse, error)
	healthCheckFunc    func() error
	capabilities       *models.ProviderCapabilities
}

func (m *CogneeMockProvider) Complete(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
	if m.completeFunc != nil {
		return m.completeFunc(ctx, req)
	}
	return &models.LLMResponse{
		ID:           "test-response",
		Content:      "Mock response",
		ProviderName: "mock",
		TokensUsed:   100,
		ResponseTime: 500, // milliseconds as int64
		FinishReason: "stop",
	}, nil
}

func (m *CogneeMockProvider) CompleteStream(ctx context.Context, req *models.LLMRequest) (<-chan *models.LLMResponse, error) {
	if m.completeStreamFunc != nil {
		return m.completeStreamFunc(ctx, req)
	}

	ch := make(chan *models.LLMResponse, 3)
	go func() {
		defer close(ch)
		ch <- &models.LLMResponse{Content: "Hello "}
		ch <- &models.LLMResponse{Content: "World"}
		ch <- &models.LLMResponse{Content: "!", FinishReason: "stop"}
	}()
	return ch, nil
}

func (m *CogneeMockProvider) HealthCheck() error {
	if m.healthCheckFunc != nil {
		return m.healthCheckFunc()
	}
	return nil
}

func (m *CogneeMockProvider) GetCapabilities() *models.ProviderCapabilities {
	if m.capabilities != nil {
		return m.capabilities
	}
	return &models.ProviderCapabilities{
		SupportedModels:   []string{"gpt-4", "gpt-3.5-turbo"},
		SupportedFeatures: []string{"chat", "completion"},
		SupportsStreaming: true,
		Metadata:          make(map[string]string),
	}
}

func (m *CogneeMockProvider) ValidateConfig(config map[string]interface{}) (bool, []string) {
	return true, nil
}

// =====================================================
// COGNEE ENHANCED PROVIDER TESTS
// =====================================================

func TestNewCogneeEnhancedProvider(t *testing.T) {
	logger := newTestLogger()
	mockProvider := &CogneeMockProvider{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"})
	}))
	defer server.Close()

	cogneeConfig := &CogneeServiceConfig{
		Enabled: true,
		BaseURL: server.URL,
	}
	cogneeService := NewCogneeServiceWithConfig(cogneeConfig, logger)

	t.Run("creates enhanced provider", func(t *testing.T) {
		enhanced := NewCogneeEnhancedProvider("test-provider", mockProvider, cogneeService, logger)

		require.NotNil(t, enhanced)
		assert.Equal(t, "test-provider", enhanced.name)
		assert.NotNil(t, enhanced.config)
		assert.NotNil(t, enhanced.stats)
		assert.Equal(t, mockProvider, enhanced.provider)
	})

	t.Run("creates with nil logger", func(t *testing.T) {
		enhanced := NewCogneeEnhancedProvider("test", mockProvider, cogneeService, nil)
		assert.NotNil(t, enhanced.logger)
	})
}

func TestNewCogneeEnhancedProviderWithConfig(t *testing.T) {
	logger := newTestLogger()
	mockProvider := &CogneeMockProvider{}

	config := &CogneeProviderConfig{
		EnhanceBeforeRequest: true,
		StoreAfterResponse:   true,
		MaxContextInjection:  4096,
		RelevanceThreshold:   0.8,
		DefaultDataset:       "custom",
	}

	enhanced := NewCogneeEnhancedProviderWithConfig(
		"test-provider",
		mockProvider,
		nil,
		config,
		logger,
	)

	require.NotNil(t, enhanced)
	assert.True(t, enhanced.config.EnhanceBeforeRequest)
	assert.Equal(t, 4096, enhanced.config.MaxContextInjection)
	assert.Equal(t, 0.8, enhanced.config.RelevanceThreshold)
	assert.Equal(t, "custom", enhanced.config.DefaultDataset)
}

func TestCogneeEnhancedProvider_Complete(t *testing.T) {
	logger := newTestLogger()

	t.Run("completes with enhancement", func(t *testing.T) {
		// Mock Cognee server
		cogneeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			if r.URL.Path == "/health" {
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"results": []interface{}{
					map[string]interface{}{"content": "relevant context"},
				},
			})
		}))
		defer cogneeServer.Close()

		cogneeConfig := &CogneeServiceConfig{
			Enabled:            true,
			BaseURL:            cogneeServer.URL,
			EnhancePrompts:     true,
			SearchTypes:        []string{"VECTOR"},
			RelevanceThreshold: 0.5,
		}
		cogneeService := NewCogneeServiceWithConfig(cogneeConfig, logger)
		cogneeService.isReady = true

		mockProvider := &CogneeMockProvider{
			completeFunc: func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
				return &models.LLMResponse{
					ID:           "resp-1",
					Content:      "Response to: " + req.Prompt,
					ProviderName: "mock",
				}, nil
			},
		}

		providerConfig := &CogneeProviderConfig{
			EnhanceBeforeRequest: true,
			StoreAfterResponse:   false, // Disable to avoid background goroutines
			EnhancementTimeout:   5 * time.Second,
			RelevanceThreshold:   0.5,
		}

		enhanced := NewCogneeEnhancedProviderWithConfig(
			"test",
			mockProvider,
			cogneeService,
			providerConfig,
			logger,
		)

		ctx := context.Background()
		req := &models.LLMRequest{
			Prompt: "What is AI?",
		}

		resp, err := enhanced.Complete(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "mock", resp.ProviderName)
		assert.True(t, resp.Metadata["cognee_enhanced"].(bool))
	})

	t.Run("completes without cognee service", func(t *testing.T) {
		mockProvider := &CogneeMockProvider{}
		enhanced := NewCogneeEnhancedProvider("test", mockProvider, nil, logger)

		ctx := context.Background()
		req := &models.LLMRequest{Prompt: "Hello"}

		resp, err := enhanced.Complete(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
	})
}

func TestCogneeEnhancedProvider_CompleteStream(t *testing.T) {
	logger := newTestLogger()

	t.Run("streams with enhancement", func(t *testing.T) {
		cogneeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"results": []interface{}{}})
		}))
		defer cogneeServer.Close()

		cogneeConfig := &CogneeServiceConfig{
			Enabled:        true,
			BaseURL:        cogneeServer.URL,
			EnhancePrompts: true,
			SearchTypes:    []string{"VECTOR"},
		}
		cogneeService := NewCogneeServiceWithConfig(cogneeConfig, logger)
		cogneeService.isReady = true

		mockProvider := &CogneeMockProvider{}

		providerConfig := &CogneeProviderConfig{
			EnhanceStreamingPrompt: true,
			StoreAfterResponse:     false,
			StreamingBufferSize:    10,
		}

		enhanced := NewCogneeEnhancedProviderWithConfig(
			"test",
			mockProvider,
			cogneeService,
			providerConfig,
			logger,
		)

		ctx := context.Background()
		req := &models.LLMRequest{Prompt: "Hello"}

		stream, err := enhanced.CompleteStream(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, stream)

		// Collect stream responses
		var content string
		for resp := range stream {
			content += resp.Content
		}

		assert.Equal(t, "Hello World!", content)
	})
}

func TestCogneeEnhancedProvider_HealthCheck(t *testing.T) {
	logger := newTestLogger()

	t.Run("healthy when provider healthy", func(t *testing.T) {
		mockProvider := &CogneeMockProvider{
			healthCheckFunc: func() error { return nil },
		}

		enhanced := NewCogneeEnhancedProvider("test", mockProvider, nil, logger)
		err := enhanced.HealthCheck()

		assert.NoError(t, err)
	})

	t.Run("unhealthy when provider fails", func(t *testing.T) {
		mockProvider := &CogneeMockProvider{
			healthCheckFunc: func() error { return assert.AnError },
		}

		enhanced := NewCogneeEnhancedProvider("test", mockProvider, nil, logger)
		err := enhanced.HealthCheck()

		assert.Error(t, err)
	})

	t.Run("checks cognee service health when available", func(t *testing.T) {
		cogneeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "healthy"})
		}))
		defer cogneeServer.Close()

		cogneeConfig := &CogneeServiceConfig{
			Enabled: true,
			BaseURL: cogneeServer.URL,
		}
		cogneeService := NewCogneeServiceWithConfig(cogneeConfig, logger)
		cogneeService.isReady = true

		mockProvider := &CogneeMockProvider{
			healthCheckFunc: func() error { return nil },
		}

		enhanced := NewCogneeEnhancedProvider("test", mockProvider, cogneeService, logger)
		err := enhanced.HealthCheck()

		assert.NoError(t, err)
	})

	t.Run("logs warning when cognee service unhealthy", func(t *testing.T) {
		// Server that returns error for health check
		cogneeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer cogneeServer.Close()

		cogneeConfig := &CogneeServiceConfig{
			Enabled: true,
			BaseURL: cogneeServer.URL,
		}
		cogneeService := NewCogneeServiceWithConfig(cogneeConfig, logger)
		cogneeService.isReady = false // Force unhealthy

		mockProvider := &CogneeMockProvider{
			healthCheckFunc: func() error { return nil },
		}

		enhanced := NewCogneeEnhancedProvider("test", mockProvider, cogneeService, logger)
		err := enhanced.HealthCheck()

		// Should still return no error, just log warning
		assert.NoError(t, err)
	})
}

func TestCogneeEnhancedProvider_GetCapabilities(t *testing.T) {
	logger := newTestLogger()
	mockProvider := &CogneeMockProvider{
		capabilities: &models.ProviderCapabilities{
			SupportedModels:   []string{"gpt-4"},
			SupportedFeatures: []string{"chat"},
			SupportsStreaming: true,
			Metadata:          make(map[string]string),
		},
	}

	enhanced := NewCogneeEnhancedProvider("test", mockProvider, nil, logger)
	caps := enhanced.GetCapabilities()

	require.NotNil(t, caps)
	assert.Contains(t, caps.SupportedModels, "gpt-4")
	assert.Equal(t, "true", caps.Metadata["cognee_enhanced"])
	assert.Contains(t, caps.SupportedFeatures, "cognee_memory")
	assert.Contains(t, caps.SupportedFeatures, "knowledge_graph")
}

func TestCogneeEnhancedProvider_ValidateConfig(t *testing.T) {
	logger := newTestLogger()
	mockProvider := &CogneeMockProvider{}

	enhanced := NewCogneeEnhancedProvider("test", mockProvider, nil, logger)

	valid, errors := enhanced.ValidateConfig(map[string]interface{}{"key": "value"})

	assert.True(t, valid)
	assert.Empty(t, errors)
}

func TestCogneeEnhancedProvider_GetStats(t *testing.T) {
	logger := newTestLogger()
	mockProvider := &CogneeMockProvider{}

	enhanced := NewCogneeEnhancedProvider("test", mockProvider, nil, logger)

	// Update stats
	enhanced.stats.mu.Lock()
	enhanced.stats.TotalRequests = 100
	enhanced.stats.EnhancedRequests = 80
	enhanced.stats.StoredResponses = 75
	enhanced.stats.mu.Unlock()

	stats := enhanced.GetStats()

	assert.Equal(t, int64(100), stats.TotalRequests)
	assert.Equal(t, int64(80), stats.EnhancedRequests)
	assert.Equal(t, int64(75), stats.StoredResponses)
}

func TestCogneeEnhancedProvider_GetConfig(t *testing.T) {
	logger := newTestLogger()
	mockProvider := &CogneeMockProvider{}

	config := &CogneeProviderConfig{
		EnhanceBeforeRequest: true,
		DefaultDataset:       "custom",
	}

	enhanced := NewCogneeEnhancedProviderWithConfig("test", mockProvider, nil, config, logger)
	cfg := enhanced.GetConfig()

	assert.True(t, cfg.EnhanceBeforeRequest)
	assert.Equal(t, "custom", cfg.DefaultDataset)
}

func TestCogneeEnhancedProvider_GetUnderlyingProvider(t *testing.T) {
	logger := newTestLogger()
	mockProvider := &CogneeMockProvider{}

	enhanced := NewCogneeEnhancedProvider("test", mockProvider, nil, logger)
	underlying := enhanced.GetUnderlyingProvider()

	assert.Equal(t, mockProvider, underlying)
}

func TestCogneeEnhancedProvider_GetName(t *testing.T) {
	logger := newTestLogger()
	mockProvider := &CogneeMockProvider{}

	enhanced := NewCogneeEnhancedProvider("my-provider", mockProvider, nil, logger)

	assert.Equal(t, "my-provider", enhanced.GetName())
}

func TestCogneeEnhancedProvider_SetConfig(t *testing.T) {
	logger := newTestLogger()
	mockProvider := &CogneeMockProvider{}

	enhanced := NewCogneeEnhancedProvider("test", mockProvider, nil, logger)

	newConfig := &CogneeProviderConfig{
		EnhanceBeforeRequest: false,
		DefaultDataset:       "new-dataset",
	}

	enhanced.SetConfig(newConfig)

	assert.False(t, enhanced.config.EnhanceBeforeRequest)
	assert.Equal(t, "new-dataset", enhanced.config.DefaultDataset)
}

func TestGetDefaultCogneeProviderConfig(t *testing.T) {
	config := getDefaultCogneeProviderConfig()

	assert.True(t, config.EnhanceBeforeRequest)
	assert.True(t, config.StoreAfterResponse)
	assert.True(t, config.AutoCognifyResponses)
	assert.True(t, config.EnableGraphReasoning)
	assert.True(t, config.EnableCodeIntelligence)
	assert.Equal(t, 2048, config.MaxContextInjection)
	assert.Equal(t, 0.7, config.RelevanceThreshold)
	assert.Equal(t, "default", config.DefaultDataset)
	assert.True(t, config.CacheEnhancements)
	assert.Equal(t, 30*time.Minute, config.CacheTTL)
}

// Compile-time check that CogneeMockProvider implements llm.LLMProvider
var _ llm.LLMProvider = (*CogneeMockProvider)(nil)

func TestCogneeEnhancedProvider_enhanceMessages(t *testing.T) {
	logger := newTestLogger()
	mockProvider := &CogneeMockProvider{}
	enhanced := NewCogneeEnhancedProvider("test", mockProvider, nil, logger)

	t.Run("returns original messages when empty", func(t *testing.T) {
		messages := []models.Message{}
		context := &EnhancedContext{}

		result := enhanced.enhanceMessages(messages, context)

		assert.Empty(t, result)
	})

	t.Run("returns original when no relevant memories", func(t *testing.T) {
		messages := []models.Message{{Role: "user", Content: "Hello"}}
		context := &EnhancedContext{
			RelevantMemories: []MemoryEntry{},
		}

		result := enhanced.enhanceMessages(messages, context)

		assert.Len(t, result, 1)
		assert.Equal(t, "Hello", result[0].Content)
	})

	t.Run("adds system message with memories", func(t *testing.T) {
		messages := []models.Message{{Role: "user", Content: "Hello"}}
		context := &EnhancedContext{
			RelevantMemories: []MemoryEntry{
				{Content: "Memory 1"},
				{Content: "Memory 2"},
			},
		}

		result := enhanced.enhanceMessages(messages, context)

		assert.Len(t, result, 2)
		assert.Equal(t, "system", result[0].Role)
		assert.Contains(t, result[0].Content, "Memory 1")
		assert.Contains(t, result[0].Content, "Memory 2")
	})

	t.Run("appends to existing system message", func(t *testing.T) {
		messages := []models.Message{
			{Role: "system", Content: "You are helpful"},
			{Role: "user", Content: "Hello"},
		}
		context := &EnhancedContext{
			RelevantMemories: []MemoryEntry{
				{Content: "Memory 1"},
			},
		}

		result := enhanced.enhanceMessages(messages, context)

		assert.Len(t, result, 2)
		assert.Equal(t, "system", result[0].Role)
		assert.Contains(t, result[0].Content, "You are helpful")
		assert.Contains(t, result[0].Content, "Memory 1")
	})

	t.Run("includes graph insights", func(t *testing.T) {
		messages := []models.Message{{Role: "user", Content: "Hello"}}
		context := &EnhancedContext{
			RelevantMemories: []MemoryEntry{{Content: "Memory"}},
			GraphInsights: []map[string]interface{}{
				{"text": "Insight 1"},
				{"text": "Insight 2"},
			},
		}

		result := enhanced.enhanceMessages(messages, context)

		assert.Len(t, result, 2)
		assert.Contains(t, result[0].Content, "Knowledge Graph Insights")
		assert.Contains(t, result[0].Content, "Insight 1")
	})
}

func TestCogneeEnhancedProvider_storeResponse(t *testing.T) {
	logger := newTestLogger()

	t.Run("stores successfully", func(t *testing.T) {
		cogneeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"})
		}))
		defer cogneeServer.Close()

		cogneeConfig := &CogneeServiceConfig{
			Enabled:        true,
			StoreResponses: true, // Must be true for ProcessResponse to call AddMemory
			BaseURL:        cogneeServer.URL,
		}
		cogneeService := NewCogneeServiceWithConfig(cogneeConfig, logger)
		cogneeService.isReady = true

		mockProvider := &CogneeMockProvider{}
		providerConfig := &CogneeProviderConfig{
			StoreAfterResponse: true,
		}
		enhanced := NewCogneeEnhancedProviderWithConfig("test", mockProvider, cogneeService, providerConfig, logger)

		ctx := context.Background()
		req := &models.LLMRequest{Prompt: "Hello"}
		resp := &models.LLMResponse{Content: "Response"}

		// This runs in background normally, but we can call directly
		enhanced.storeResponse(ctx, req, resp)

		stats := enhanced.GetStats()
		assert.Equal(t, int64(1), stats.StoredResponses)
	})

	t.Run("handles storage error", func(t *testing.T) {
		cogneeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer cogneeServer.Close()

		cogneeConfig := &CogneeServiceConfig{
			Enabled:        true,
			StoreResponses: true, // Must be true for ProcessResponse to call AddMemory
			BaseURL:        cogneeServer.URL,
		}
		cogneeService := NewCogneeServiceWithConfig(cogneeConfig, logger)
		cogneeService.isReady = true

		mockProvider := &CogneeMockProvider{}
		providerConfig := &CogneeProviderConfig{
			StoreAfterResponse: true,
		}
		enhanced := NewCogneeEnhancedProviderWithConfig("test", mockProvider, cogneeService, providerConfig, logger)

		ctx := context.Background()
		req := &models.LLMRequest{Prompt: "Hello"}
		resp := &models.LLMResponse{Content: "Response"}

		enhanced.storeResponse(ctx, req, resp)

		stats := enhanced.GetStats()
		assert.Equal(t, int64(1), stats.StorageErrors)
	})
}

func TestEnhanceProviderRegistry(t *testing.T) {
	logger := newTestLogger()

	cogneeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer cogneeServer.Close()

	cogneeConfig := &CogneeServiceConfig{
		Enabled: true,
		BaseURL: cogneeServer.URL,
	}
	cogneeService := NewCogneeServiceWithConfig(cogneeConfig, logger)

	t.Run("enhances all providers in registry", func(t *testing.T) {
		registryConfig := &RegistryConfig{
			DefaultTimeout:       30 * time.Second,
			DisableAutoDiscovery: true,
			MaxRetries:           3,
		}
		registry := NewProviderRegistry(registryConfig, nil)
		_ = registry.RegisterProvider("provider1", &CogneeMockProvider{})
		_ = registry.RegisterProvider("provider2", &CogneeMockProvider{})

		err := EnhanceProviderRegistry(registry, cogneeService, logger)
		require.NoError(t, err)

		// Check that providers are enhanced by verifying their capabilities
		// The registry wraps providers in circuitBreakerProvider, but GetCapabilities
		// should pass through to the underlying CogneeEnhancedProvider
		p1, err := registry.GetProvider("provider1")
		require.NoError(t, err)

		caps := p1.GetCapabilities()
		require.NotNil(t, caps)
		require.NotNil(t, caps.Metadata)
		assert.Equal(t, "true", caps.Metadata["cognee_enhanced"], "Provider should be enhanced with Cognee")
	})

	t.Run("skips already enhanced providers", func(t *testing.T) {
		registryConfig := &RegistryConfig{
			DefaultTimeout:       30 * time.Second,
			MaxRetries:           3,
			DisableAutoDiscovery: true,
		}
		registry := NewProviderRegistry(registryConfig, nil)
		mockProvider := &CogneeMockProvider{}
		enhanced := NewCogneeEnhancedProvider("enhanced", mockProvider, cogneeService, logger)
		_ = registry.RegisterProvider("enhanced", enhanced)

		err := EnhanceProviderRegistry(registry, cogneeService, logger)
		require.NoError(t, err)

		// Should still work and not double-wrap
		p, err := registry.GetProvider("enhanced")
		require.NoError(t, err)
		assert.NotNil(t, p)
	})
}

func TestWrapProvidersWithCognee(t *testing.T) {
	logger := newTestLogger()

	providers := map[string]llm.LLMProvider{
		"provider1": &CogneeMockProvider{},
		"provider2": &CogneeMockProvider{},
	}

	cogneeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer cogneeServer.Close()

	cogneeConfig := &CogneeServiceConfig{
		Enabled: true,
		BaseURL: cogneeServer.URL,
	}
	cogneeService := NewCogneeServiceWithConfig(cogneeConfig, logger)

	wrapped := WrapProvidersWithCognee(providers, cogneeService, logger)

	require.Len(t, wrapped, 2)

	for name, provider := range wrapped {
		enhanced, ok := provider.(*CogneeEnhancedProvider)
		require.True(t, ok, "Provider %s should be enhanced", name)
		assert.NotNil(t, enhanced)
	}
}

// =====================================================
// TESTS FOR ISREADY CHECKS
// =====================================================

func TestCogneeEnhancedProvider_storeResponse_IsReadyCheck(t *testing.T) {
	logger := newTestLogger()

	t.Run("skips storage when cognee is nil", func(t *testing.T) {
		mockProvider := &CogneeMockProvider{}
		enhanced := NewCogneeEnhancedProvider("test", mockProvider, nil, logger)

		ctx := context.Background()
		req := &models.LLMRequest{Prompt: "Hello"}
		resp := &models.LLMResponse{Content: "Response"}

		// Should return early without error
		enhanced.storeResponse(ctx, req, resp)

		// Stats should not change
		stats := enhanced.GetStats()
		assert.Equal(t, int64(0), stats.StoredResponses)
		assert.Equal(t, int64(0), stats.StorageErrors)
	})

	t.Run("skips storage when cognee is not ready", func(t *testing.T) {
		cogneeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer cogneeServer.Close()

		cogneeConfig := &CogneeServiceConfig{
			Enabled:        true,
			StoreResponses: true,
			BaseURL:        cogneeServer.URL,
		}
		cogneeService := NewCogneeServiceWithConfig(cogneeConfig, logger)
		cogneeService.isReady = false // NOT ready

		mockProvider := &CogneeMockProvider{}
		providerConfig := &CogneeProviderConfig{
			StoreAfterResponse: true,
		}
		enhanced := NewCogneeEnhancedProviderWithConfig("test", mockProvider, cogneeService, providerConfig, logger)

		ctx := context.Background()
		req := &models.LLMRequest{Prompt: "Hello"}
		resp := &models.LLMResponse{Content: "Response"}

		enhanced.storeResponse(ctx, req, resp)

		// Stats should not change - storage was skipped
		stats := enhanced.GetStats()
		assert.Equal(t, int64(0), stats.StoredResponses)
		assert.Equal(t, int64(0), stats.StorageErrors)
	})

	t.Run("proceeds when cognee is ready", func(t *testing.T) {
		cogneeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"})
		}))
		defer cogneeServer.Close()

		cogneeConfig := &CogneeServiceConfig{
			Enabled:        true,
			StoreResponses: true,
			BaseURL:        cogneeServer.URL,
		}
		cogneeService := NewCogneeServiceWithConfig(cogneeConfig, logger)
		cogneeService.isReady = true // Ready

		mockProvider := &CogneeMockProvider{}
		providerConfig := &CogneeProviderConfig{
			StoreAfterResponse: true,
		}
		enhanced := NewCogneeEnhancedProviderWithConfig("test", mockProvider, cogneeService, providerConfig, logger)

		ctx := context.Background()
		req := &models.LLMRequest{Prompt: "Hello"}
		resp := &models.LLMResponse{Content: "Response"}

		enhanced.storeResponse(ctx, req, resp)

		// Should have attempted storage
		stats := enhanced.GetStats()
		assert.Equal(t, int64(1), stats.StoredResponses)
	})
}

func TestCogneeEnhancedProvider_CompleteStream_IsReadyCheck(t *testing.T) {
	logger := newTestLogger()

	t.Run("does not store when cognee is not ready during stream", func(t *testing.T) {
		cogneeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"results": []interface{}{}})
		}))
		defer cogneeServer.Close()

		cogneeConfig := &CogneeServiceConfig{
			Enabled:        true,
			BaseURL:        cogneeServer.URL,
			EnhancePrompts: true,
			StoreResponses: true,
			SearchTypes:    []string{"VECTOR"},
		}
		cogneeService := NewCogneeServiceWithConfig(cogneeConfig, logger)
		cogneeService.isReady = false // NOT ready

		mockProvider := &CogneeMockProvider{}

		providerConfig := &CogneeProviderConfig{
			EnhanceStreamingPrompt: true,
			StoreAfterResponse:     true,
			StreamingBufferSize:    10,
		}

		enhanced := NewCogneeEnhancedProviderWithConfig(
			"test",
			mockProvider,
			cogneeService,
			providerConfig,
			logger,
		)

		ctx := context.Background()
		req := &models.LLMRequest{Prompt: "Hello"}

		stream, err := enhanced.CompleteStream(ctx, req)
		require.NoError(t, err)

		// Consume stream
		var content string
		for resp := range stream {
			content += resp.Content
		}
		assert.Equal(t, "Hello World!", content)

		// Wait a bit for any async operations
		time.Sleep(100 * time.Millisecond)

		// Should not have stored anything since not ready
		stats := enhanced.GetStats()
		assert.Equal(t, int64(0), stats.StoredResponses)
	})

	t.Run("stores when cognee is ready during stream", func(t *testing.T) {
		storeCalledCh := make(chan struct{}, 1)
		cogneeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			if r.URL.Path == "/api/v1/add" {
				select {
				case storeCalledCh <- struct{}{}:
				default:
				}
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"})
		}))
		defer cogneeServer.Close()

		cogneeConfig := &CogneeServiceConfig{
			Enabled:        true,
			BaseURL:        cogneeServer.URL,
			EnhancePrompts: false,
			StoreResponses: true,
			SearchTypes:    []string{"VECTOR"},
		}
		cogneeService := NewCogneeServiceWithConfig(cogneeConfig, logger)
		cogneeService.isReady = true // Ready

		mockProvider := &CogneeMockProvider{}

		providerConfig := &CogneeProviderConfig{
			EnhanceStreamingPrompt: false,
			StoreAfterResponse:     true,
			StreamingBufferSize:    10,
		}

		enhanced := NewCogneeEnhancedProviderWithConfig(
			"test",
			mockProvider,
			cogneeService,
			providerConfig,
			logger,
		)

		ctx := context.Background()
		req := &models.LLMRequest{Prompt: "Hello"}

		stream, err := enhanced.CompleteStream(ctx, req)
		require.NoError(t, err)

		// Consume stream
		var content string
		for resp := range stream {
			content += resp.Content
		}
		assert.Equal(t, "Hello World!", content)

		// Wait for async storage
		select {
		case <-storeCalledCh:
			// Storage was attempted
		case <-time.After(2 * time.Second):
			// Storage might have been attempted - check stats
		}

		// Give async operation time to complete
		time.Sleep(200 * time.Millisecond)

		// Should have attempted storage
		stats := enhanced.GetStats()
		assert.True(t, stats.StoredResponses > 0 || stats.StorageErrors > 0,
			"Expected storage attempt when cognee is ready")
	})
}

func TestCogneeEnhancedProvider_GracefulDegradation(t *testing.T) {
	logger := newTestLogger()

	t.Run("works without cognee service", func(t *testing.T) {
		mockProvider := &CogneeMockProvider{
			completeFunc: func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
				return &models.LLMResponse{
					Content:      "Response without Cognee",
					ProviderName: "mock",
				}, nil
			},
		}

		// Disable enhancement in config when no Cognee service
		config := &CogneeProviderConfig{
			EnhanceBeforeRequest: false,
			StoreAfterResponse:   false,
		}
		enhanced := NewCogneeEnhancedProviderWithConfig("test", mockProvider, nil, config, logger)

		ctx := context.Background()
		req := &models.LLMRequest{Prompt: "Hello"}

		resp, err := enhanced.Complete(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, "Response without Cognee", resp.Content)
		// Metadata reflects config setting, not Cognee availability
		assert.False(t, resp.Metadata["cognee_enhanced"].(bool))
	})

	t.Run("works when cognee becomes unavailable", func(t *testing.T) {
		cogneeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate server going down
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer cogneeServer.Close()

		cogneeConfig := &CogneeServiceConfig{
			Enabled:        true,
			BaseURL:        cogneeServer.URL,
			EnhancePrompts: true,
			StoreResponses: true,
		}
		cogneeService := NewCogneeServiceWithConfig(cogneeConfig, logger)
		cogneeService.isReady = true // Was ready, but server fails

		mockProvider := &CogneeMockProvider{
			completeFunc: func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
				return &models.LLMResponse{
					Content:      "Response despite Cognee failure",
					ProviderName: "mock",
				}, nil
			},
		}

		providerConfig := &CogneeProviderConfig{
			EnhanceBeforeRequest: true,
			StoreAfterResponse:   true,
			EnhancementTimeout:   1 * time.Second,
		}

		enhanced := NewCogneeEnhancedProviderWithConfig("test", mockProvider, cogneeService, providerConfig, logger)

		ctx := context.Background()
		req := &models.LLMRequest{Prompt: "Hello"}

		resp, err := enhanced.Complete(ctx, req)

		// Should still complete despite Cognee failure
		require.NoError(t, err)
		assert.Equal(t, "Response despite Cognee failure", resp.Content)
	})
}

// =====================================================
// BENCHMARK TESTS
// =====================================================

func BenchmarkCogneeEnhancedProvider_Complete(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	mockProvider := &CogneeMockProvider{}
	enhanced := NewCogneeEnhancedProvider("test", mockProvider, nil, logger)

	req := &models.LLMRequest{Prompt: "Test prompt"}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = enhanced.Complete(ctx, req)
	}
}

func BenchmarkCogneeEnhancedProvider_CompleteWithEnhancement(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	cogneeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"results": []interface{}{}})
	}))
	defer cogneeServer.Close()

	cogneeConfig := &CogneeServiceConfig{
		Enabled:        true,
		BaseURL:        cogneeServer.URL,
		EnhancePrompts: true,
		SearchTypes:    []string{"VECTOR"},
	}
	cogneeService := NewCogneeServiceWithConfig(cogneeConfig, logger)
	cogneeService.isReady = true

	mockProvider := &CogneeMockProvider{}

	providerConfig := &CogneeProviderConfig{
		EnhanceBeforeRequest: true,
		StoreAfterResponse:   false,
		EnhancementTimeout:   5 * time.Second,
	}

	enhanced := NewCogneeEnhancedProviderWithConfig("test", mockProvider, cogneeService, providerConfig, logger)

	req := &models.LLMRequest{Prompt: "Test prompt"}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = enhanced.Complete(ctx, req)
	}
}
