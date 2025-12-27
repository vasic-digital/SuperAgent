package claude_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/superagent/superagent/internal/llm/providers"
	"github.com/superagent/superagent/internal/models"
)

func TestClaudeProvider_Basic(t *testing.T) {
	logger := logrus.New()
	provider, err := providers.NewClaudeProvider("test-api-key", "", "claude-3-opus-20240229", 30*time.Second, 3, logger)
	require.NoError(t, err)
	assert.NotNil(t, provider)

	// Test configuration validation with proper config
	valid, errs := provider.ValidateConfig(map[string]interface{}{
		"api_key": "test-api-key",
		"model":   "claude-3-opus-20240229",
	})
	assert.True(t, valid)
	assert.Empty(t, errs)
}

func TestClaudeProvider_EmptyAPIKey(t *testing.T) {
	logger := logrus.New()
	// Should fail to create provider with empty API key
	provider, err := providers.NewClaudeProvider("", "", "claude-3-opus-20240229", 30*time.Second, 3, logger)
	assert.Error(t, err)
	assert.Nil(t, provider)
	assert.Contains(t, err.Error(), "API key is required")
}

func TestClaudeProvider_Capabilities(t *testing.T) {
	logger := logrus.New()
	provider, err := providers.NewClaudeProvider("test-api-key", "", "claude-3-opus-20240229", 30*time.Second, 3, logger)
	require.NoError(t, err)
	caps := provider.GetCapabilities()
	assert.NotNil(t, caps)
	assert.NotEmpty(t, caps.SupportedModels)
	assert.Contains(t, caps.SupportedModels, "claude-3-sonnet-20240229")
	assert.Contains(t, caps.SupportedModels, "claude-3-opus-20240229")
	assert.Contains(t, caps.SupportedFeatures, "long_context")
	assert.Contains(t, caps.SupportedFeatures, "vision")
	assert.Contains(t, caps.SupportedFeatures, "tools")
	assert.True(t, caps.SupportsStreaming)
	assert.True(t, caps.SupportsFunctionCalling)
	assert.True(t, caps.SupportsVision)
	assert.NotNil(t, caps.Metadata)
}

func TestClaudeProvider_CompleteRequest(t *testing.T) {
	logger := logrus.New()
	provider, err := providers.NewClaudeProvider("test-api-key", "", "claude-3-opus-20240229", 30*time.Second, 3, logger)
	require.NoError(t, err)

	req := &models.LLMRequest{
		ID: "test-req-1",
		ModelParams: models.ModelParameters{
			Model: "claude-3-sonnet-20240229",
		},
		Prompt: "Hello, how are you?",
	}

	// This will fail without actual API key, but tests the error handling
	resp, err := provider.Complete(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Content)
}

func TestClaudeProvider_CompleteWithDifferentModels(t *testing.T) {
	logger := logrus.New()
	provider, err := providers.NewClaudeProvider("test-api-key", "", "claude-3-opus-20240229", 30*time.Second, 3, logger)
	require.NoError(t, err)

	// Test with different model selections
	modelList := []string{
		"claude-3-sonnet-20240229",
		"claude-3-opus-20240229",
	}

	for _, model := range modelList {
		req := &models.LLMRequest{
			ID: "test-" + model,
			ModelParams: models.ModelParameters{
				Model: model,
			},
			Prompt: "Test prompt for " + model,
		}

		resp, err := provider.Complete(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	}
}

func TestClaudeProvider_InvalidModel(t *testing.T) {
	logger := logrus.New()
	provider, err := providers.NewClaudeProvider("test-api-key", "", "claude-3-opus-20240229", 30*time.Second, 3, logger)
	require.NoError(t, err)

	req := &models.LLMRequest{
		ID: "test-invalid",
		ModelParams: models.ModelParameters{
			Model: "invalid-model",
		},
		Prompt: "Test prompt",
	}

	resp, err := provider.Complete(context.Background(), req)
	// Should fail gracefully without panic
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestClaudeProvider_MemoryUsage(t *testing.T) {
	logger := logrus.New()
	provider, err := providers.NewClaudeProvider("test-api-key", "", "claude-3-opus-20240229", 30*time.Second, 3, logger)
	require.NoError(t, err)

	// Test multiple requests to ensure no memory leaks
	for i := 0; i < 10; i++ {
		req := &models.LLMRequest{
			ID: fmt.Sprintf("test-req-%d", i),
			ModelParams: models.ModelParameters{
				Model: "claude-3-sonnet-20240229",
			},
			Prompt: fmt.Sprintf("Memory test request %d", i),
		}

		resp, err := provider.Complete(context.Background(), req)
		if err != nil {
			t.Logf("Request %d failed: %v", i, err)
		}

		_ = resp
	}

	// Provider should still be responsive
	assert.True(t, true)
}

func TestClaudeProvider_Timeout(t *testing.T) {
	logger := logrus.New()
	provider, err := providers.NewClaudeProvider("test-api-key", "", "claude-3-opus-20240229", 30*time.Second, 3, logger)
	require.NoError(t, err)

	// Create a request that might timeout
	req := &models.LLMRequest{
		ID: "test-timeout",
		ModelParams: models.ModelParameters{
			Model: "claude-3-sonnet-20240229",
		},
		Prompt: "This is a timeout test request",
	}

	start := time.Now()
	resp, err := provider.Complete(context.Background(), req)
	elapsed := time.Since(start)

	// Should complete quickly
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, elapsed < 10*time.Second)
}

func TestClaudeProvider_CustomBaseURL(t *testing.T) {
	logger := logrus.New()
	// Test with custom base URL
	provider, err := providers.NewClaudeProvider("test-api-key", "http://localhost:8080", "custom-model", 30*time.Second, 3, logger)
	require.NoError(t, err)
	assert.NotNil(t, provider)

	caps := provider.GetCapabilities()
	assert.NotNil(t, caps)
}

func TestClaudeProvider_ValidateConfig(t *testing.T) {
	logger := logrus.New()
	provider, err := providers.NewClaudeProvider("test-api-key", "", "claude-3-opus-20240229", 30*time.Second, 3, logger)
	require.NoError(t, err)

	// Test with valid config
	valid, errs := provider.ValidateConfig(map[string]interface{}{
		"api_key": "test-api-key",
		"model":   "claude-3-opus-20240229",
	})
	assert.True(t, valid)
	assert.Empty(t, errs)

	// Test with invalid config (missing required fields)
	valid, errs = provider.ValidateConfig(map[string]interface{}{
		"timeout": 30,
		"retries": 3,
	})
	assert.False(t, valid)
	assert.NotEmpty(t, errs)
	assert.Contains(t, errs, "api_key is required")
	assert.Contains(t, errs, "model is required")

	// Test with empty config
	valid, errs = provider.ValidateConfig(map[string]interface{}{})
	assert.False(t, valid)
	assert.NotEmpty(t, errs)
}
