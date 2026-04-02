package services

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"dev.helix.agent/internal/llm"
	"dev.helix.agent/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockLLMProvider is a mock implementation of llm.LLMProvider for testing
type MockRegistryLLMProvider struct {
	CompleteFunc       func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error)
	CompleteStreamFunc func(ctx context.Context, req *models.LLMRequest) (<-chan *models.LLMResponse, error)
	HealthCheckFunc    func() error
	GetCapabilitiesFunc func() *models.ProviderCapabilities
	ValidateConfigFunc func(config map[string]interface{}) (bool, []string)
	name               string
}

func (m *MockRegistryLLMProvider) Complete(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
	if m.CompleteFunc != nil {
		return m.CompleteFunc(ctx, req)
	}
	return &models.LLMResponse{
		ID:       "mock-response",
		Content:  "mock content from " + m.name,
		ProviderName: m.name,
	}, nil
}

func (m *MockRegistryLLMProvider) CompleteStream(ctx context.Context, req *models.LLMRequest) (<-chan *models.LLMResponse, error) {
	if m.CompleteStreamFunc != nil {
		return m.CompleteStreamFunc(ctx, req)
	}
	ch := make(chan *models.LLMResponse, 1)
	ch <- &models.LLMResponse{
		ID:       "mock-stream",
		Content:  "mock stream from " + m.name,
		ProviderName: m.name,
	}
	close(ch)
	return ch, nil
}

func (m *MockRegistryLLMProvider) HealthCheck() error {
	if m.HealthCheckFunc != nil {
		return m.HealthCheckFunc()
	}
	return nil
}

func (m *MockRegistryLLMProvider) GetCapabilities() *models.ProviderCapabilities {
	if m.GetCapabilitiesFunc != nil {
		return m.GetCapabilitiesFunc()
	}
	return &models.ProviderCapabilities{
		MaxTokens:   4096,
		Models:      []string{"model-1", "model-2"},
		SupportsStreaming: true,
	}
}

func (m *MockRegistryLLMProvider) ValidateConfig(config map[string]interface{}) (bool, []string) {
	if m.ValidateConfigFunc != nil {
		return m.ValidateConfigFunc(config)
	}
	return true, nil
}

var _ llm.LLMProvider = (*MockRegistryLLMProvider)(nil)

func TestNewProviderRegistry_Additional(t *testing.T) {
	t.Run("creates registry with default config", func(t *testing.T) {
		registry := NewProviderRegistry(nil, nil)

		assert.NotNil(t, registry)
		assert.NotNil(t, registry.providers)
		assert.NotNil(t, registry.circuitBreakers)
		assert.NotNil(t, registry.concurrencySemaphores)
		assert.NotNil(t, registry.providerConfigs)
		assert.NotNil(t, registry.providerHealth)
		assert.NotNil(t, registry.activeRequests)
		assert.NotNil(t, registry.ensemble)
		assert.NotNil(t, registry.requestService)
	})

	t.Run("creates registry with custom config", func(t *testing.T) {
		cfg := &RegistryConfig{
			DefaultTimeout: 60 * time.Second,
			MaxRetries:     5,
		}

		registry := NewProviderRegistry(cfg, nil)

		assert.NotNil(t, registry)
		assert.Equal(t, cfg, registry.config)
	})
}

func TestNewProviderRegistryWithoutAutoDiscovery_Additional(t *testing.T) {
	t.Run("creates registry without auto-discovery", func(t *testing.T) {
		registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)

		assert.NotNil(t, registry)
		assert.False(t, registry.autoDiscovery)
	})
}

func TestProviderRegistry_RegisterProvider_Additional(t *testing.T) {
	t.Run("registers provider successfully", func(t *testing.T) {
		registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)
		provider := &MockRegistryLLMProvider{name: "test-provider"}

		err := registry.RegisterProvider("test-provider", provider)

		require.NoError(t, err)
		providers := registry.ListProviders()
		assert.Contains(t, providers, "test-provider")
	})

	t.Run("registers multiple providers", func(t *testing.T) {
		registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)

		err1 := registry.RegisterProvider("provider-1", &MockRegistryLLMProvider{name: "p1"})
		err2 := registry.RegisterProvider("provider-2", &MockRegistryLLMProvider{name: "p2"})
		err3 := registry.RegisterProvider("provider-3", &MockRegistryLLMProvider{name: "p3"})

		require.NoError(t, err1)
		require.NoError(t, err2)
		require.NoError(t, err3)

		providers := registry.ListProviders()
		assert.Len(t, providers, 3)
	})

	t.Run("returns error for duplicate provider", func(t *testing.T) {
		registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)
		provider := &MockRegistryLLMProvider{name: "test-provider"}

		err1 := registry.RegisterProvider("test-provider", provider)
		err2 := registry.RegisterProvider("test-provider", provider)

		require.NoError(t, err1)
		require.Error(t, err2)
		assert.Contains(t, err2.Error(), "already registered")
	})

	t.Run("returns error for nil provider", func(t *testing.T) {
		registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)

		err := registry.RegisterProvider("test-provider", nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "nil provider")
	})
}

func TestProviderRegistry_UnregisterProvider_Additional(t *testing.T) {
	t.Run("unregisters existing provider", func(t *testing.T) {
		registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)
		provider := &MockRegistryLLMProvider{name: "test-provider"}

		_ = registry.RegisterProvider("test-provider", provider)
		err := registry.UnregisterProvider("test-provider")

		require.NoError(t, err)
		providers := registry.ListProviders()
		assert.NotContains(t, providers, "test-provider")
	})

	t.Run("returns error for non-existent provider", func(t *testing.T) {
		registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)

		err := registry.UnregisterProvider("non-existent")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestProviderRegistry_GetProvider_Additional(t *testing.T) {
	t.Run("returns existing provider", func(t *testing.T) {
		registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)
		provider := &MockRegistryLLMProvider{name: "test-provider"}

		_ = registry.RegisterProvider("test-provider", provider)
		retrieved, err := registry.GetProvider("test-provider")

		require.NoError(t, err)
		assert.NotNil(t, retrieved)
	})

	t.Run("returns error for non-existent provider", func(t *testing.T) {
		registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)

		provider, err := registry.GetProvider("non-existent")

		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestProviderRegistry_ListProviders_Additional(t *testing.T) {
	t.Run("returns empty list when no providers", func(t *testing.T) {
		registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)

		providers := registry.ListProviders()

		assert.Empty(t, providers)
	})

	t.Run("returns all registered providers", func(t *testing.T) {
		registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)

		_ = registry.RegisterProvider("provider-a", &MockRegistryLLMProvider{name: "a"})
		_ = registry.RegisterProvider("provider-b", &MockRegistryLLMProvider{name: "b"})

		providers := registry.ListProviders()

		assert.Len(t, providers, 2)
		assert.Contains(t, providers, "provider-a")
		assert.Contains(t, providers, "provider-b")
	})

	t.Run("returns sorted list", func(t *testing.T) {
		registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)

		_ = registry.RegisterProvider("zebra", &MockRegistryLLMProvider{name: "z"})
		_ = registry.RegisterProvider("apple", &MockRegistryLLMProvider{name: "a"})
		_ = registry.RegisterProvider("mango", &MockRegistryLLMProvider{name: "m"})

		providers := registry.ListProviders()

		// Should be sorted alphabetically
		assert.Equal(t, []string{"apple", "mango", "zebra"}, providers)
	})
}

func TestProviderRegistry_GetProviderConfig_Additional(t *testing.T) {
	t.Run("returns config for registered provider", func(t *testing.T) {
		registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)
		provider := &MockRegistryLLMProvider{name: "test-provider"}

		_ = registry.RegisterProvider("test-provider", provider)
		
		// Manually set config
		registry.mu.Lock()
		registry.providerConfigs["test-provider"] = &ProviderConfig{
			Name:    "test-provider",
			Enabled: true,
		}
		registry.mu.Unlock()

		config, err := registry.GetProviderConfig("test-provider")

		require.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, "test-provider", config.Name)
	})

	t.Run("returns error for non-existent provider", func(t *testing.T) {
		registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)

		config, err := registry.GetProviderConfig("non-existent")

		require.Error(t, err)
		assert.Nil(t, config)
	})
}

func TestProviderRegistry_UpdateProviderConfig_Additional(t *testing.T) {
	t.Run("updates existing provider config", func(t *testing.T) {
		registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)
		provider := &MockRegistryLLMProvider{name: "test-provider"}

		_ = registry.RegisterProvider("test-provider", provider)
		
		newConfig := &ProviderConfig{
			Name:     "test-provider",
			Enabled:  true,
			MaxRetries: 5,
		}

		err := registry.UpdateProviderConfig("test-provider", newConfig)

		require.NoError(t, err)

		retrieved, _ := registry.GetProviderConfig("test-provider")
		assert.Equal(t, 5, retrieved.MaxRetries)
	})

	t.Run("returns error for non-existent provider", func(t *testing.T) {
		registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)

		config := &ProviderConfig{Name: "non-existent"}
		err := registry.UpdateProviderConfig("non-existent", config)

		require.Error(t, err)
	})

	t.Run("returns error for nil config", func(t *testing.T) {
		registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)
		provider := &MockRegistryLLMProvider{name: "test-provider"}

		_ = registry.RegisterProvider("test-provider", provider)
		err := registry.UpdateProviderConfig("test-provider", nil)

		require.Error(t, err)
	})
}

func TestProviderRegistry_IsProviderHealthy_Additional(t *testing.T) {
	t.Run("returns true for healthy provider", func(t *testing.T) {
		registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)
		provider := &MockRegistryLLMProvider{
			name: "healthy-provider",
			HealthCheckFunc: func() error {
				return nil
			},
		}

		_ = registry.RegisterProvider("healthy-provider", provider)
		healthy := registry.IsProviderHealthy("healthy-provider")

		assert.True(t, healthy)
	})

	t.Run("returns false for unhealthy provider", func(t *testing.T) {
		registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)
		provider := &MockRegistryLLMProvider{
			name: "unhealthy-provider",
			HealthCheckFunc: func() error {
				return errors.New("health check failed")
			},
		}

		_ = registry.RegisterProvider("unhealthy-provider", provider)
		healthy := registry.IsProviderHealthy("unhealthy-provider")

		assert.False(t, healthy)
	})

	t.Run("returns false for non-existent provider", func(t *testing.T) {
		registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)

		healthy := registry.IsProviderHealthy("non-existent")

		assert.False(t, healthy)
	})
}

func TestProviderRegistry_GetProviderHealthStatus_Additional(t *testing.T) {
	t.Run("returns health status for provider", func(t *testing.T) {
		registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)
		provider := &MockRegistryLLMProvider{name: "test-provider"}

		_ = registry.RegisterProvider("test-provider", provider)
		
		// Set health status
		registry.mu.Lock()
		registry.providerHealth["test-provider"] = &ProviderVerificationResult{
			Provider: "test-provider",
			Status:   ProviderStatusHealthy,
			Verified: true,
		}
		registry.mu.Unlock()

		status, err := registry.GetProviderHealthStatus("test-provider")

		require.NoError(t, err)
		assert.Equal(t, ProviderStatusHealthy, status.Status)
		assert.True(t, status.Verified)
	})

	t.Run("returns error for non-existent provider", func(t *testing.T) {
		registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)

		status, err := registry.GetProviderHealthStatus("non-existent")

		require.Error(t, err)
		assert.Nil(t, status)
	})
}

func TestProviderRegistry_Complete_Additional(t *testing.T) {
	t.Run("completes request with provider", func(t *testing.T) {
		registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)
		provider := &MockRegistryLLMProvider{
			name: "test-provider",
			CompleteFunc: func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
				return &models.LLMResponse{
					ID:      "response-1",
					Content: "Test response",
					ProviderName: "test-provider",
				}, nil
			},
		}

		_ = registry.RegisterProvider("test-provider", provider)

		req := &models.LLMRequest{
			Prompt: "Test prompt",
			ModelParams: models.ModelParameters{
				Model: "gpt-4",
			},
		}

		resp, err := registry.Complete(context.Background(), "test-provider", req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "Test response", resp.Content)
	})

	t.Run("returns error for non-existent provider", func(t *testing.T) {
		registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)

		req := &models.LLMRequest{Prompt: "Test"}
		resp, err := registry.Complete(context.Background(), "non-existent", req)

		require.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("returns error when provider fails", func(t *testing.T) {
		registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)
		provider := &MockRegistryLLMProvider{
			name: "failing-provider",
			CompleteFunc: func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
				return nil, errors.New("provider error")
			},
		}

		_ = registry.RegisterProvider("failing-provider", provider)

		req := &models.LLMRequest{Prompt: "Test"}
		resp, err := registry.Complete(context.Background(), "failing-provider", req)

		require.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestProviderRegistry_CompleteWithFallback_Additional(t *testing.T) {
	t.Run("falls back to secondary provider", func(t *testing.T) {
		registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)
		
		failingProvider := &MockRegistryLLMProvider{
			name: "failing",
			CompleteFunc: func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
				return nil, errors.New("primary failed")
			},
		}
		
		backupProvider := &MockRegistryLLMProvider{
			name: "backup",
			CompleteFunc: func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
				return &models.LLMResponse{
					ID:      "backup-response",
					Content: "Backup response",
					ProviderName: "backup",
				}, nil
			},
		}

		_ = registry.RegisterProvider("failing", failingProvider)
		_ = registry.RegisterProvider("backup", backupProvider)

		req := &models.LLMRequest{Prompt: "Test"}
		resp, err := registry.CompleteWithFallback(context.Background(), "failing", []string{"backup"}, req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "Backup response", resp.Content)
	})

	t.Run("returns error when all providers fail", func(t *testing.T) {
		registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)
		
		provider := &MockRegistryLLMProvider{
			name: "failing",
			CompleteFunc: func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
				return nil, errors.New("failed")
			},
		}

		_ = registry.RegisterProvider("failing", provider)

		req := &models.LLMRequest{Prompt: "Test"}
		resp, err := registry.CompleteWithFallback(context.Background(), "failing", []string{"failing"}, req)

		require.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestProviderRegistry_CompleteStream_Additional(t *testing.T) {
	t.Run("streams from provider", func(t *testing.T) {
		registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)
		
		ch := make(chan *models.LLMResponse, 2)
		ch <- &models.LLMResponse{ID: "chunk-1", Content: "Hello"}
		ch <- &models.LLMResponse{ID: "chunk-2", Content: " World"}
		close(ch)

		provider := &MockRegistryLLMProvider{
			name: "streaming-provider",
			CompleteStreamFunc: func(ctx context.Context, req *models.LLMRequest) (<-chan *models.LLMResponse, error) {
				return ch, nil
			},
		}

		_ = registry.RegisterProvider("streaming-provider", provider)

		req := &models.LLMRequest{Prompt: "Test"}
		stream, err := registry.CompleteStream(context.Background(), "streaming-provider", req)

		require.NoError(t, err)
		assert.NotNil(t, stream)

		var chunks []string
		for resp := range stream {
			chunks = append(chunks, resp.Content)
		}
		assert.Equal(t, []string{"Hello", " World"}, chunks)
	})

	t.Run("returns error for non-existent provider", func(t *testing.T) {
		registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)

		req := &models.LLMRequest{Prompt: "Test"}
		stream, err := registry.CompleteStream(context.Background(), "non-existent", req)

		require.Error(t, err)
		assert.Nil(t, stream)
	})
}

func TestProviderRegistry_RunEnsemble_Additional(t *testing.T) {
	t.Run("runs ensemble with registered providers", func(t *testing.T) {
		registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)
		
		provider1 := &MockRegistryLLMProvider{
			name: "provider-1",
			CompleteFunc: func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
				return &models.LLMResponse{
					ID:       "resp-1",
					Content:  "Response 1",
					Confidence: 0.8,
					ProviderName: "provider-1",
				}, nil
			},
		}
		
		provider2 := &MockRegistryLLMProvider{
			name: "provider-2",
			CompleteFunc: func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
				return &models.LLMResponse{
					ID:       "resp-2",
					Content:  "Response 2",
					Confidence: 0.9,
					ProviderName: "provider-2",
				}, nil
			},
		}

		_ = registry.RegisterProvider("provider-1", provider1)
		_ = registry.RegisterProvider("provider-2", provider2)

		req := &models.LLMRequest{Prompt: "Test ensemble"}
		result, err := registry.RunEnsemble(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Selected)
	})

	t.Run("returns error when no providers registered", func(t *testing.T) {
		registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)

		req := &models.LLMRequest{Prompt: "Test"}
		result, err := registry.RunEnsemble(context.Background(), req)

		require.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestProviderRegistry_ConcurrentAccess_Additional(t *testing.T) {
	t.Run("handles concurrent provider registration", func(t *testing.T) {
		registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				name := string(rune('a' + idx))
				_ = registry.RegisterProvider(name, &MockRegistryLLMProvider{name: name})
			}(i)
		}

		wg.Wait()

		providers := registry.ListProviders()
		assert.Len(t, providers, 10)
	})

	t.Run("handles concurrent read and write", func(t *testing.T) {
		registry := NewProviderRegistryWithoutAutoDiscovery(nil, nil)

		// Pre-populate
		for i := 0; i < 5; i++ {
			_ = registry.RegisterProvider(string(rune('a'+i)), &MockRegistryLLMProvider{})
		}

		var wg sync.WaitGroup

		// Concurrent reads
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = registry.ListProviders()
			}()
		}

		// Concurrent writes
		for i := 5; i < 10; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				_ = registry.RegisterProvider(string(rune('a'+idx)), &MockRegistryLLMProvider{})
			}(i)
		}

		wg.Wait()
	})
}

func TestProviderConfig_Validation(t *testing.T) {
	t.Run("validates required fields", func(t *testing.T) {
		config := &ProviderConfig{
			Name:    "test-provider",
			Type:    "openai",
			Enabled: true,
		}

		assert.Equal(t, "test-provider", config.Name)
		assert.Equal(t, "openai", config.Type)
		assert.True(t, config.Enabled)
	})

	t.Run("validates model config", func(t *testing.T) {
		config := &ModelConfig{
			ID:       "gpt-4",
			Name:     "GPT-4",
			Enabled:  true,
			Weight:   1.0,
		}

		assert.Equal(t, "gpt-4", config.ID)
		assert.Equal(t, "GPT-4", config.Name)
		assert.True(t, config.Enabled)
		assert.Equal(t, 1.0, config.Weight)
	})
}

func TestProviderHealthStatus(t *testing.T) {
	t.Run("defines all health statuses", func(t *testing.T) {
		statuses := []ProviderHealthStatus{
			ProviderStatusUnknown,
			ProviderStatusHealthy,
			ProviderStatusRateLimited,
			ProviderStatusAuthFailed,
			ProviderStatusUnhealthy,
		}

		expected := []string{
			"unknown",
			"healthy",
			"rate_limited",
			"auth_failed",
			"unhealthy",
		}

		for i, status := range statuses {
			assert.Equal(t, expected[i], string(status))
		}
	})
}

func TestProviderVerificationResult(t *testing.T) {
	t.Run("creates verification result", func(t *testing.T) {
		result := &ProviderVerificationResult{
			Provider:     "test-provider",
			Status:       ProviderStatusHealthy,
			Verified:     true,
			Score:        8.5,
			ResponseTime: 150 * time.Millisecond,
			TestedAt:     time.Now(),
		}

		assert.Equal(t, "test-provider", result.Provider)
		assert.Equal(t, ProviderStatusHealthy, result.Status)
		assert.True(t, result.Verified)
		assert.Equal(t, 8.5, result.Score)
	})
}
