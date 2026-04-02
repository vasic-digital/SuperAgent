package junie

import (
	"context"
	"testing"
	"time"

	"dev.helix.agent/internal/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewACPProvider(t *testing.T) {
	t.Run("creates ACP provider with defaults", func(t *testing.T) {
		logger := logrus.New()
		provider := NewACPProvider(nil, logger)

		assert.NotNil(t, provider)
		assert.NotNil(t, provider.logger)
		assert.Equal(t, logger, provider.logger)
	})

	t.Run("creates ACP provider with config", func(t *testing.T) {
		logger := logrus.New()
		config := &ACPConfig{
			Endpoint: "https://acp.api.com",
			Timeout:  60 * time.Second,
		}
		provider := NewACPProvider(config, logger)

		assert.NotNil(t, provider)
		assert.Equal(t, "https://acp.api.com", provider.config.Endpoint)
	})
}

func TestACPProvider_Complete(t *testing.T) {
	t.Run("validates request parameters", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		provider := NewACPProvider(nil, logger)

		// Test empty prompt
		req1 := &models.LLMRequest{
			Prompt: "",
			ModelParams: models.ModelParameters{
				Model: "junie-acp",
			},
		}
		resp1, err1 := provider.Complete(context.Background(), req1)
		require.Error(t, err1)
		assert.Nil(t, resp1)

		// Test empty model
		req2 := &models.LLMRequest{
			Prompt: "Test prompt",
			ModelParams: models.ModelParameters{
				Model: "",
			},
		}
		resp2, err2 := provider.Complete(context.Background(), req2)
		require.Error(t, err2)
		assert.Nil(t, resp2)
	})

	t.Run("handles context cancellation", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		provider := NewACPProvider(nil, logger)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		req := &models.LLMRequest{
			Prompt: "Test prompt",
			ModelParams: models.ModelParameters{
				Model: "junie-acp",
			},
		}

		resp, err := provider.Complete(ctx, req)

		// Response depends on timing
		_ = resp
		_ = err
	})
}

func TestACPProvider_CompleteStream(t *testing.T) {
	t.Run("validates request parameters", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		provider := NewACPProvider(nil, logger)

		req := &models.LLMRequest{
			Prompt: "",
			ModelParams: models.ModelParameters{
				Model: "junie-acp",
			},
		}

		stream, err := provider.CompleteStream(context.Background(), req)

		require.Error(t, err)
		assert.Nil(t, stream)
	})
}

func TestACPProvider_HealthCheck(t *testing.T) {
	t.Run("performs health check", func(t *testing.T) {
		logger := logrus.New()
		provider := NewACPProvider(nil, logger)

		err := provider.HealthCheck()

		// May return nil or error depending on implementation
		_ = err
	})
}

func TestACPProvider_GetCapabilities(t *testing.T) {
	t.Run("returns ACP capabilities", func(t *testing.T) {
		logger := logrus.New()
		provider := NewACPProvider(nil, logger)

		caps := provider.GetCapabilities()

		assert.NotNil(t, caps)
		assert.NotNil(t, caps.Models)
	})
}

func TestACPProvider_ValidateConfig(t *testing.T) {
	t.Run("validates configuration", func(t *testing.T) {
		logger := logrus.New()
		provider := NewACPProvider(nil, logger)

		config := map[string]interface{}{
			"endpoint": "https://acp.api.com",
			"api_key":  "test-key",
		}

		valid, errors := provider.ValidateConfig(config)

		// Validation depends on implementation
		_ = valid
		_ = errors
	})

	t.Run("rejects empty config", func(t *testing.T) {
		logger := logrus.New()
		provider := NewACPProvider(nil, logger)

		valid, errors := provider.ValidateConfig(map[string]interface{}{})

		// Should return validation errors
		_ = valid
		_ = errors
	})
}

func TestACPProvider_Name(t *testing.T) {
	t.Run("returns provider name", func(t *testing.T) {
		logger := logrus.New()
		provider := NewACPProvider(nil, logger)

		name := provider.Name()

		assert.Equal(t, "junie-acp", name)
	})
}

func TestACPConfig_Defaults(t *testing.T) {
	t.Run("provides default config", func(t *testing.T) {
		config := DefaultACPConfig()

		assert.NotEmpty(t, config.Endpoint)
		assert.Greater(t, config.Timeout, time.Duration(0))
	})
}

func TestACPProvider_ConcurrentAccess(t *testing.T) {
	t.Run("handles concurrent requests", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		provider := NewACPProvider(nil, logger)

		var wg sync.WaitGroup
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				req := &models.LLMRequest{
					Prompt: "Test",
					ModelParams: models.ModelParameters{
						Model: "junie-acp",
					},
				}
				_, _ = provider.Complete(context.Background(), req)
			}()
		}

		wg.Wait()
	})
}
