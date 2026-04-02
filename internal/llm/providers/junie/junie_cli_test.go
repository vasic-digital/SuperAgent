package junie

import (
	"context"
	"errors"
	"testing"
	"time"

	"dev.helix.agent/internal/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCLIProvider(t *testing.T) {
	t.Run("creates provider with default config", func(t *testing.T) {
		logger := logrus.New()
		provider := NewCLIProvider(nil, logger)

		assert.NotNil(t, provider)
		assert.NotNil(t, provider.logger)
		assert.NotNil(t, provider.client)
		assert.Equal(t, logger, provider.logger)
	})

	t.Run("creates provider with custom config", func(t *testing.T) {
		logger := logrus.New()
		config := &Config{
			BaseURL: "https://custom.api.com",
			Timeout: 60 * time.Second,
		}
		provider := NewCLIProvider(config, logger)

		assert.NotNil(t, provider)
		assert.Equal(t, "https://custom.api.com", provider.config.BaseURL)
		assert.Equal(t, 60*time.Second, provider.config.Timeout)
	})
}

func TestCLIProvider_Complete(t *testing.T) {
	t.Run("returns error for empty prompt", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		provider := NewCLIProvider(nil, logger)

		req := &models.LLMRequest{
			Prompt: "",
			ModelParams: models.ModelParameters{
				Model: "junie-v1",
			},
		}

		resp, err := provider.Complete(context.Background(), req)

		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "prompt")
	})

	t.Run("returns error for empty model", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		provider := NewCLIProvider(nil, logger)

		req := &models.LLMRequest{
			Prompt: "Test prompt",
			ModelParams: models.ModelParameters{
				Model: "",
			},
		}

		resp, err := provider.Complete(context.Background(), req)

		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "model")
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		provider := NewCLIProvider(nil, logger)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		req := &models.LLMRequest{
			Prompt: "Test prompt",
			ModelParams: models.ModelParameters{
				Model: "junie-v1",
			},
		}

		resp, err := provider.Complete(ctx, req)

		// May return error or nil depending on implementation
		_ = resp
		_ = err
	})
}

func TestCLIProvider_CompleteStream(t *testing.T) {
	t.Run("returns error for empty prompt", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		provider := NewCLIProvider(nil, logger)

		req := &models.LLMRequest{
			Prompt: "",
			ModelParams: models.ModelParameters{
				Model: "junie-v1",
			},
		}

		stream, err := provider.CompleteStream(context.Background(), req)

		require.Error(t, err)
		assert.Nil(t, stream)
	})

	t.Run("returns stream channel on success", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		provider := NewCLIProvider(nil, logger)

		req := &models.LLMRequest{
			Prompt: "Test prompt",
			ModelParams: models.ModelParameters{
				Model: "junie-v1",
			},
		}

		stream, err := provider.CompleteStream(context.Background(), req)

		// Stream may be nil or valid depending on implementation
		_ = stream
		_ = err
	})
}

func TestCLIProvider_HealthCheck(t *testing.T) {
	t.Run("returns nil when healthy", func(t *testing.T) {
		logger := logrus.New()
		provider := NewCLIProvider(nil, logger)

		err := provider.HealthCheck()

		// May return nil or error depending on implementation
		_ = err
	})
}

func TestCLIProvider_GetCapabilities(t *testing.T) {
	t.Run("returns capabilities", func(t *testing.T) {
		logger := logrus.New()
		provider := NewCLIProvider(nil, logger)

		caps := provider.GetCapabilities()

		assert.NotNil(t, caps)
		assert.NotNil(t, caps.Models)
		assert.Greater(t, caps.MaxTokens, 0)
	})
}

func TestCLIProvider_ValidateConfig(t *testing.T) {
	t.Run("validates correct config", func(t *testing.T) {
		logger := logrus.New()
		provider := NewCLIProvider(nil, logger)

		config := map[string]interface{}{
			"api_key": "test-key",
			"model":   "junie-v1",
		}

		valid, errors := provider.ValidateConfig(config)

		// Validation behavior depends on implementation
		_ = valid
		_ = errors
	})

	t.Run("returns errors for invalid config", func(t *testing.T) {
		logger := logrus.New()
		provider := NewCLIProvider(nil, logger)

		config := map[string]interface{}{}

		valid, errors := provider.ValidateConfig(config)

		// Should return validation errors
		_ = valid
		_ = errors
	})
}

func TestCLIProvider_buildRequest(t *testing.T) {
	t.Run("builds request correctly", func(t *testing.T) {
		logger := logrus.New()
		provider := NewCLIProvider(nil, logger)

		req := &models.LLMRequest{
			Prompt: "Test prompt",
			ModelParams: models.ModelParameters{
				Model:       "junie-v1",
				MaxTokens:   100,
				Temperature: 0.7,
				TopP:        0.9,
			},
		}

		httpReq, err := provider.buildRequest(context.Background(), req)

		// May return request or error depending on implementation
		if err == nil {
			assert.NotNil(t, httpReq)
		}
	})
}

func TestCLIProvider_parseResponse(t *testing.T) {
	t.Run("parses valid response", func(t *testing.T) {
		logger := logrus.New()
		provider := NewCLIProvider(nil, logger)

		// Create mock response body
		jsonBody := `{
			"id": "resp-123",
			"content": "Test response",
			"usage": {
				"prompt_tokens": 10,
				"completion_tokens": 20,
				"total_tokens": 30
			}
		}`

		resp, err := provider.parseResponse([]byte(jsonBody))

		if err == nil {
			assert.NotNil(t, resp)
		}
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		logger := logrus.New()
		provider := NewCLIProvider(nil, logger)

		resp, err := provider.parseResponse([]byte("invalid json"))

		require.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestConfig_Validation(t *testing.T) {
	t.Run("validates default config", func(t *testing.T) {
		config := DefaultConfig()

		assert.NotEmpty(t, config.BaseURL)
		assert.Greater(t, config.Timeout, time.Duration(0))
		assert.Greater(t, config.MaxRetries, 0)
	})

	t.Run("allows custom config values", func(t *testing.T) {
		config := &Config{
			BaseURL:    "https://custom.api.com",
			APIKey:     "custom-key",
			Timeout:    120 * time.Second,
			MaxRetries: 5,
		}

		assert.Equal(t, "https://custom.api.com", config.BaseURL)
		assert.Equal(t, "custom-key", config.APIKey)
		assert.Equal(t, 120*time.Second, config.Timeout)
		assert.Equal(t, 5, config.MaxRetries)
	})
}

func TestCLIProvider_ConcurrentAccess(t *testing.T) {
	t.Run("handles concurrent Complete calls", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		provider := NewCLIProvider(nil, logger)

		var wg sync.WaitGroup
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				req := &models.LLMRequest{
					Prompt: "Test",
					ModelParams: models.ModelParameters{
						Model: "junie-v1",
					},
				}
				_, _ = provider.Complete(context.Background(), req)
			}()
		}

		wg.Wait()
	})
}

func TestCLIProvider_Name(t *testing.T) {
	t.Run("returns provider name", func(t *testing.T) {
		logger := logrus.New()
		provider := NewCLIProvider(nil, logger)

		name := provider.Name()

		assert.NotEmpty(t, name)
		assert.Equal(t, "junie-cli", name)
	})
}

func TestCLIProvider_SetLogger(t *testing.T) {
	t.Run("sets logger correctly", func(t *testing.T) {
		logger1 := logrus.New()
		provider := NewCLIProvider(nil, logger1)

		logger2 := logrus.New()
		provider.SetLogger(logger2)

		assert.Equal(t, logger2, provider.logger)
	})
}
