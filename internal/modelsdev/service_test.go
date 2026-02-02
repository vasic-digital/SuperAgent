package modelsdev

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createMockServer creates a test server that returns mock Models.dev API responses
func createMockServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/models":
			response := ModelsListResponse{
				Models: []ModelInfo{
					{
						ID:            "claude-3-sonnet",
						Name:          "Claude 3 Sonnet",
						Provider:      "anthropic",
						DisplayName:   "Claude 3 Sonnet",
						Description:   "A powerful AI model",
						ContextWindow: 200000,
						MaxTokens:     4096,
						Pricing: &ModelPricing{
							InputPrice:  0.003,
							OutputPrice: 0.015,
							Currency:    "USD",
							Unit:        "tokens",
						},
						Capabilities: ModelCapabilities{
							Vision:          true,
							FunctionCalling: true,
							Streaming:       true,
						},
					},
					{
						ID:       "gpt-4",
						Name:     "GPT-4",
						Provider: "openai",
						Capabilities: ModelCapabilities{
							Vision:          true,
							FunctionCalling: true,
							Streaming:       true,
							Reasoning:       true,
						},
					},
				},
				Total: 2,
				Page:  1,
				Limit: 100,
			}
			_ = json.NewEncoder(w).Encode(response)

		case "/models/claude-3-sonnet":
			response := ModelDetailsResponse{
				Model: ModelInfo{
					ID:            "claude-3-sonnet",
					Name:          "Claude 3 Sonnet",
					Provider:      "anthropic",
					DisplayName:   "Claude 3 Sonnet",
					Description:   "A powerful AI model",
					ContextWindow: 200000,
					MaxTokens:     4096,
					Pricing: &ModelPricing{
						InputPrice:  0.003,
						OutputPrice: 0.015,
						Currency:    "USD",
						Unit:        "tokens",
					},
					Capabilities: ModelCapabilities{
						Vision:          true,
						FunctionCalling: true,
						Streaming:       true,
					},
				},
			}
			_ = json.NewEncoder(w).Encode(response)

		case "/models/nonexistent":
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(APIError{
				Type:    "not_found",
				Message: "Model not found",
				Code:    404,
			})

		case "/providers":
			response := ProvidersListResponse{
				Providers: []ProviderInfo{
					{
						ID:          "anthropic",
						Name:        "Anthropic",
						DisplayName: "Anthropic",
						Description: "AI safety company",
						ModelsCount: 10,
						Website:     "https://anthropic.com",
					},
					{
						ID:          "openai",
						Name:        "OpenAI",
						DisplayName: "OpenAI",
						Description: "AI research lab",
						ModelsCount: 15,
						Website:     "https://openai.com",
					},
				},
				Total: 2,
			}
			_ = json.NewEncoder(w).Encode(response)

		case "/providers/anthropic":
			response := ProviderInfo{
				ID:          "anthropic",
				Name:        "Anthropic",
				DisplayName: "Anthropic",
				Description: "AI safety company",
				ModelsCount: 10,
				Website:     "https://anthropic.com",
			}
			_ = json.NewEncoder(w).Encode(response)

		case "/providers/nonexistent":
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(APIError{
				Type:    "not_found",
				Message: "Provider not found",
				Code:    404,
			})

		default:
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(APIError{
				Type:    "not_found",
				Message: "Endpoint not found",
				Code:    404,
			})
		}
	}))
}

func TestNewService(t *testing.T) {
	config := &ServiceConfig{
		Client: ClientConfig{
			BaseURL: "https://api.models.dev",
			Timeout: 30 * time.Second,
		},
		Cache: CacheConfig{
			ModelTTL:        1 * time.Hour,
			ProviderTTL:     2 * time.Hour,
			MaxModels:       1000,
			MaxProviders:    50,
			CleanupInterval: 10 * time.Minute,
		},
		RefreshOnStart: false,
		AutoRefresh:    false,
	}

	log := logrus.New()
	log.SetLevel(logrus.ErrorLevel)

	service := NewService(config, log)
	require.NotNil(t, service)
	defer func() { _ = service.Stop() }()

	assert.NotNil(t, service.client)
	assert.NotNil(t, service.cache)
}

func TestNewService_DefaultConfig(t *testing.T) {
	service := NewService(nil, nil)
	require.NotNil(t, service)
	defer func() { _ = service.Stop() }()

	// Verify defaults are applied
	assert.NotNil(t, service.client)
	assert.NotNil(t, service.cache)
	assert.NotNil(t, service.log)
}

func TestService_GetModel(t *testing.T) {
	server := createMockServer()
	defer server.Close()

	config := &ServiceConfig{
		Client: ClientConfig{
			BaseURL: server.URL,
			Timeout: 5 * time.Second,
		},
		Cache: CacheConfig{
			ModelTTL:        1 * time.Hour,
			CleanupInterval: 1 * time.Hour,
			MaxModels:       100,
		},
		RefreshOnStart: false,
		AutoRefresh:    false,
	}

	log := logrus.New()
	log.SetLevel(logrus.ErrorLevel)

	service := NewService(config, log)
	defer func() { _ = service.Stop() }()

	ctx := context.Background()

	// Get model
	model, err := service.GetModel(ctx, "claude-3-sonnet")
	require.NoError(t, err)
	require.NotNil(t, model)
	assert.Equal(t, "claude-3-sonnet", model.ID)
	assert.Equal(t, "Claude 3 Sonnet", model.Name)
	assert.Equal(t, "anthropic", model.Provider)

	// Verify caching - second call should use cache
	model2, err := service.GetModel(ctx, "claude-3-sonnet")
	require.NoError(t, err)
	assert.Equal(t, model.ID, model2.ID)
}

func TestService_GetModel_NotFound(t *testing.T) {
	server := createMockServer()
	defer server.Close()

	config := &ServiceConfig{
		Client: ClientConfig{
			BaseURL: server.URL,
			Timeout: 5 * time.Second,
		},
		Cache: CacheConfig{
			ModelTTL:        1 * time.Hour,
			CleanupInterval: 1 * time.Hour,
			MaxModels:       100,
		},
		RefreshOnStart: false,
		AutoRefresh:    false,
	}

	log := logrus.New()
	log.SetLevel(logrus.ErrorLevel)

	service := NewService(config, log)
	defer func() { _ = service.Stop() }()

	ctx := context.Background()

	_, err := service.GetModel(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestService_GetModel_EmptyID(t *testing.T) {
	service := NewService(nil, nil)
	defer func() { _ = service.Stop() }()

	ctx := context.Background()

	_, err := service.GetModel(ctx, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "model ID is required")
}

func TestService_GetModelWithCache(t *testing.T) {
	server := createMockServer()
	defer server.Close()

	config := &ServiceConfig{
		Client: ClientConfig{
			BaseURL: server.URL,
			Timeout: 5 * time.Second,
		},
		Cache: CacheConfig{
			ModelTTL:        1 * time.Hour,
			CleanupInterval: 1 * time.Hour,
			MaxModels:       100,
		},
		RefreshOnStart: false,
		AutoRefresh:    false,
	}

	log := logrus.New()
	log.SetLevel(logrus.ErrorLevel)

	service := NewService(config, log)
	defer func() { _ = service.Stop() }()

	ctx := context.Background()

	// Get with cache enabled (default)
	model, err := service.GetModelWithCache(ctx, "claude-3-sonnet", true)
	require.NoError(t, err)
	assert.NotNil(t, model)

	// Get bypassing cache
	model2, err := service.GetModelWithCache(ctx, "claude-3-sonnet", false)
	require.NoError(t, err)
	assert.NotNil(t, model2)
}

func TestService_ListModels(t *testing.T) {
	server := createMockServer()
	defer server.Close()

	config := &ServiceConfig{
		Client: ClientConfig{
			BaseURL: server.URL,
			Timeout: 5 * time.Second,
		},
		Cache: CacheConfig{
			ModelTTL:        1 * time.Hour,
			CleanupInterval: 1 * time.Hour,
			MaxModels:       100,
		},
		RefreshOnStart: false,
		AutoRefresh:    false,
	}

	log := logrus.New()
	log.SetLevel(logrus.ErrorLevel)

	service := NewService(config, log)
	defer func() { _ = service.Stop() }()

	ctx := context.Background()

	models, total, err := service.ListModels(ctx, nil)
	require.NoError(t, err)
	assert.Len(t, models, 2)
	assert.Equal(t, 2, total)
}

func TestService_ListModels_WithFilters(t *testing.T) {
	server := createMockServer()
	defer server.Close()

	config := &ServiceConfig{
		Client: ClientConfig{
			BaseURL: server.URL,
			Timeout: 5 * time.Second,
		},
		Cache: CacheConfig{
			ModelTTL:        1 * time.Hour,
			CleanupInterval: 1 * time.Hour,
			MaxModels:       100,
		},
		RefreshOnStart: false,
		AutoRefresh:    false,
	}

	log := logrus.New()
	log.SetLevel(logrus.ErrorLevel)

	service := NewService(config, log)
	defer func() { _ = service.Stop() }()

	ctx := context.Background()

	filters := &ModelFilters{
		Provider: "anthropic",
		Page:     1,
		Limit:    10,
	}

	models, _, err := service.ListModels(ctx, filters)
	require.NoError(t, err)
	assert.NotEmpty(t, models)
}

func TestService_GetProvider(t *testing.T) {
	server := createMockServer()
	defer server.Close()

	config := &ServiceConfig{
		Client: ClientConfig{
			BaseURL: server.URL,
			Timeout: 5 * time.Second,
		},
		Cache: CacheConfig{
			ProviderTTL:     2 * time.Hour,
			CleanupInterval: 1 * time.Hour,
			MaxProviders:    50,
		},
		RefreshOnStart: false,
		AutoRefresh:    false,
	}

	log := logrus.New()
	log.SetLevel(logrus.ErrorLevel)

	service := NewService(config, log)
	defer func() { _ = service.Stop() }()

	ctx := context.Background()

	provider, err := service.GetProvider(ctx, "anthropic")
	require.NoError(t, err)
	require.NotNil(t, provider)
	assert.Equal(t, "anthropic", provider.ID)
	assert.Equal(t, "Anthropic", provider.Name)
}

func TestService_GetProvider_NotFound(t *testing.T) {
	server := createMockServer()
	defer server.Close()

	config := &ServiceConfig{
		Client: ClientConfig{
			BaseURL: server.URL,
			Timeout: 5 * time.Second,
		},
		RefreshOnStart: false,
		AutoRefresh:    false,
	}

	log := logrus.New()
	log.SetLevel(logrus.ErrorLevel)

	service := NewService(config, log)
	defer func() { _ = service.Stop() }()

	ctx := context.Background()

	_, err := service.GetProvider(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestService_GetProvider_EmptyID(t *testing.T) {
	service := NewService(nil, nil)
	defer func() { _ = service.Stop() }()

	ctx := context.Background()

	_, err := service.GetProvider(ctx, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "provider ID is required")
}

func TestService_ListProviders(t *testing.T) {
	server := createMockServer()
	defer server.Close()

	config := &ServiceConfig{
		Client: ClientConfig{
			BaseURL: server.URL,
			Timeout: 5 * time.Second,
		},
		Cache: CacheConfig{
			ProviderTTL:     2 * time.Hour,
			CleanupInterval: 1 * time.Hour,
			MaxProviders:    50,
		},
		RefreshOnStart: false,
		AutoRefresh:    false,
	}

	log := logrus.New()
	log.SetLevel(logrus.ErrorLevel)

	service := NewService(config, log)
	defer func() { _ = service.Stop() }()

	ctx := context.Background()

	providers, err := service.ListProviders(ctx)
	require.NoError(t, err)
	assert.Len(t, providers, 2)
}

func TestService_SearchModels(t *testing.T) {
	server := createMockServer()
	defer server.Close()

	config := &ServiceConfig{
		Client: ClientConfig{
			BaseURL: server.URL,
			Timeout: 5 * time.Second,
		},
		RefreshOnStart: false,
		AutoRefresh:    false,
	}

	log := logrus.New()
	log.SetLevel(logrus.ErrorLevel)

	service := NewService(config, log)
	defer func() { _ = service.Stop() }()

	ctx := context.Background()

	models, total, err := service.SearchModels(ctx, "claude", nil)
	require.NoError(t, err)
	assert.NotEmpty(t, models)
	assert.Greater(t, total, 0)
}

func TestService_RefreshCache(t *testing.T) {
	server := createMockServer()
	defer server.Close()

	config := &ServiceConfig{
		Client: ClientConfig{
			BaseURL: server.URL,
			Timeout: 5 * time.Second,
		},
		Cache: CacheConfig{
			ModelTTL:        1 * time.Hour,
			ProviderTTL:     2 * time.Hour,
			CleanupInterval: 1 * time.Hour,
			MaxModels:       100,
			MaxProviders:    50,
		},
		RefreshOnStart: false,
		AutoRefresh:    false,
	}

	log := logrus.New()
	log.SetLevel(logrus.ErrorLevel)

	service := NewService(config, log)
	defer func() { _ = service.Stop() }()

	ctx := context.Background()

	result := service.RefreshCache(ctx)
	assert.True(t, result.Success)
	assert.Greater(t, result.ModelsRefreshed, 0)
	assert.Greater(t, result.ProvidersRefreshed, 0)
	assert.Empty(t, result.Errors)
}

func TestService_InvalidateCache(t *testing.T) {
	server := createMockServer()
	defer server.Close()

	config := &ServiceConfig{
		Client: ClientConfig{
			BaseURL: server.URL,
			Timeout: 5 * time.Second,
		},
		Cache: CacheConfig{
			ModelTTL:        1 * time.Hour,
			CleanupInterval: 1 * time.Hour,
			MaxModels:       100,
		},
		RefreshOnStart: false,
		AutoRefresh:    false,
	}

	log := logrus.New()
	log.SetLevel(logrus.ErrorLevel)

	service := NewService(config, log)
	defer func() { _ = service.Stop() }()

	ctx := context.Background()

	// Cache a model
	_, err := service.GetModel(ctx, "claude-3-sonnet")
	require.NoError(t, err)

	// Invalidate
	service.InvalidateCache(ctx, "claude-3-sonnet", "")

	// Cache should be invalidated (stats check)
	stats := service.CacheStats()
	assert.Equal(t, 0, stats.ModelCount)
}

func TestService_InvalidateAll(t *testing.T) {
	server := createMockServer()
	defer server.Close()

	config := &ServiceConfig{
		Client: ClientConfig{
			BaseURL: server.URL,
			Timeout: 5 * time.Second,
		},
		Cache: CacheConfig{
			ModelTTL:        1 * time.Hour,
			ProviderTTL:     2 * time.Hour,
			CleanupInterval: 1 * time.Hour,
			MaxModels:       100,
			MaxProviders:    50,
		},
		RefreshOnStart: false,
		AutoRefresh:    false,
	}

	log := logrus.New()
	log.SetLevel(logrus.ErrorLevel)

	service := NewService(config, log)
	defer func() { _ = service.Stop() }()

	ctx := context.Background()

	// Populate cache
	service.RefreshCache(ctx)

	stats := service.CacheStats()
	assert.Greater(t, stats.ModelCount, 0)

	// Invalidate all
	service.InvalidateAll(ctx)

	stats = service.CacheStats()
	assert.Equal(t, 0, stats.ModelCount)
	assert.Equal(t, 0, stats.ProviderCount)
}

func TestService_CacheStats(t *testing.T) {
	service := NewService(nil, nil)
	defer func() { _ = service.Stop() }()

	stats := service.CacheStats()
	assert.Equal(t, 0, stats.ModelCount)
	assert.Equal(t, 0, stats.ProviderCount)
	assert.Equal(t, int64(0), stats.TotalHits)
	assert.Equal(t, int64(0), stats.TotalMisses)
}

func TestService_StartStop(t *testing.T) {
	config := &ServiceConfig{
		RefreshOnStart: false,
		AutoRefresh:    false,
	}

	service := NewService(config, nil)

	ctx := context.Background()

	// Start service
	err := service.Start(ctx)
	assert.NoError(t, err)

	// Start again (should be no-op)
	err = service.Start(ctx)
	assert.NoError(t, err)

	// Stop service
	err = service.Stop()
	assert.NoError(t, err)

	// Stop again (should be no-op)
	err = service.Stop()
	assert.NoError(t, err)
}

func TestService_GetModelCapabilities(t *testing.T) {
	server := createMockServer()
	defer server.Close()

	config := &ServiceConfig{
		Client: ClientConfig{
			BaseURL: server.URL,
			Timeout: 5 * time.Second,
		},
		RefreshOnStart: false,
		AutoRefresh:    false,
	}

	log := logrus.New()
	log.SetLevel(logrus.ErrorLevel)

	service := NewService(config, log)
	defer func() { _ = service.Stop() }()

	ctx := context.Background()

	capabilities, err := service.GetModelCapabilities(ctx, "claude-3-sonnet")
	require.NoError(t, err)
	require.NotNil(t, capabilities)
	assert.True(t, capabilities.Vision)
	assert.True(t, capabilities.FunctionCalling)
	assert.True(t, capabilities.Streaming)
}

func TestService_GetModelPricing(t *testing.T) {
	server := createMockServer()
	defer server.Close()

	config := &ServiceConfig{
		Client: ClientConfig{
			BaseURL: server.URL,
			Timeout: 5 * time.Second,
		},
		RefreshOnStart: false,
		AutoRefresh:    false,
	}

	log := logrus.New()
	log.SetLevel(logrus.ErrorLevel)

	service := NewService(config, log)
	defer func() { _ = service.Stop() }()

	ctx := context.Background()

	pricing, err := service.GetModelPricing(ctx, "claude-3-sonnet")
	require.NoError(t, err)
	require.NotNil(t, pricing)
	assert.Equal(t, 0.003, pricing.InputCost)
	assert.Equal(t, 0.015, pricing.OutputCost)
	assert.Equal(t, "USD", pricing.Currency)
}

func TestService_NewServiceWithClient(t *testing.T) {
	client := NewClient(&ClientConfig{
		BaseURL: "https://api.models.dev",
		Timeout: 30 * time.Second,
	})

	cache := NewCache(&CacheConfig{
		ModelTTL:        1 * time.Hour,
		CleanupInterval: 1 * time.Hour,
		MaxModels:       100,
	})

	config := &ServiceConfig{
		RefreshOnStart: false,
		AutoRefresh:    false,
	}

	service := NewServiceWithClient(client, cache, config, nil)
	require.NotNil(t, service)
	defer func() { _ = service.Stop() }()

	assert.Equal(t, client, service.client)
}

func TestService_ConvertModelInfoToModel(t *testing.T) {
	service := NewService(nil, nil)
	defer func() { _ = service.Stop() }()

	info := &ModelInfo{
		ID:            "test-model",
		Name:          "Test Model",
		Provider:      "test-provider",
		DisplayName:   "Test Model Display",
		Description:   "A test model",
		ContextWindow: 100000,
		MaxTokens:     4096,
		Pricing: &ModelPricing{
			InputPrice:  0.001,
			OutputPrice: 0.002,
			Currency:    "USD",
			Unit:        "tokens",
		},
		Capabilities: ModelCapabilities{
			Vision:          true,
			FunctionCalling: true,
			Streaming:       true,
		},
		Performance: &ModelPerformance{
			BenchmarkScore:   95.5,
			PopularityScore:  1000,
			ReliabilityScore: 0.99,
		},
		Tags:       []string{"test", "model"},
		Categories: []string{"text"},
		Family:     "test-family",
		Version:    "1.0",
	}

	model := service.convertModelInfoToModel(info)
	require.NotNil(t, model)
	assert.Equal(t, "test-model", model.ID)
	assert.Equal(t, "Test Model", model.Name)
	assert.Equal(t, "test-provider", model.Provider)
	assert.Equal(t, 100000, model.ContextWindow)
	assert.NotNil(t, model.Pricing)
	assert.Equal(t, 0.001, model.Pricing.InputCost)
	assert.NotNil(t, model.Capabilities)
	assert.True(t, model.Capabilities.Vision)
	assert.NotNil(t, model.Performance)
	assert.Equal(t, 95.5, model.Performance.BenchmarkScore)
}

func TestService_ConvertProviderInfoToProvider(t *testing.T) {
	service := NewService(nil, nil)
	defer func() { _ = service.Stop() }()

	info := &ProviderInfo{
		ID:          "test-provider",
		Name:        "Test Provider",
		DisplayName: "Test Provider Display",
		Description: "A test provider",
		ModelsCount: 10,
		Website:     "https://test.com",
		APIDocsURL:  "https://test.com/docs",
		Features:    []string{"streaming", "function_calling"},
	}

	provider := service.convertProviderInfoToProvider(info)
	require.NotNil(t, provider)
	assert.Equal(t, "test-provider", provider.ID)
	assert.Equal(t, "Test Provider", provider.Name)
	assert.Equal(t, "Test Provider Display", provider.DisplayName)
	assert.Equal(t, 10, provider.ModelsCount)
	assert.Equal(t, "https://test.com", provider.Website)
}

func TestDefaultConfigs(t *testing.T) {
	// Test DefaultClientConfig
	clientConfig := DefaultClientConfig()
	assert.NotEmpty(t, clientConfig.BaseURL)
	assert.Greater(t, clientConfig.Timeout, time.Duration(0))
	assert.NotEmpty(t, clientConfig.UserAgent)

	// Test DefaultCacheConfig
	cacheConfig := DefaultCacheConfig()
	assert.Greater(t, cacheConfig.ModelTTL, time.Duration(0))
	assert.Greater(t, cacheConfig.ProviderTTL, time.Duration(0))
	assert.Greater(t, cacheConfig.MaxModels, 0)
	assert.Greater(t, cacheConfig.MaxProviders, 0)

	// Test DefaultServiceConfig
	serviceConfig := DefaultServiceConfig()
	assert.NotEmpty(t, serviceConfig.Client.BaseURL)
	assert.Greater(t, serviceConfig.Cache.ModelTTL, time.Duration(0))
}
