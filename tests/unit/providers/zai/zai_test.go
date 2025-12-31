package zai_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/superagent/superagent/internal/llm/providers/zai"
)

func TestNewZaiProvider(t *testing.T) {
	t.Run("valid configuration", func(t *testing.T) {
		provider := zai.NewZAIProvider(
			"test-api-key",
			"https://api.zai.com",
			"zai-pro",
		)

		require.NotNil(t, provider)
	})

	t.Run("with default base URL", func(t *testing.T) {
		provider := zai.NewZAIProvider(
			"test-api-key",
			"",
			"zai-pro",
		)
		require.NotNil(t, provider)
	})

	t.Run("with default model", func(t *testing.T) {
		provider := zai.NewZAIProvider(
			"test-api-key",
			"https://api.zai.com",
			"",
		)
		require.NotNil(t, provider)
	})
}

func TestZaiProvider_GetCapabilities(t *testing.T) {
	provider := zai.NewZAIProvider(
		"test-api-key",
		"https://api.zai.com",
		"zai-pro",
	)
	require.NotNil(t, provider)

	capabilities := provider.GetCapabilities()

	assert.NotNil(t, capabilities)
	// ZAI does NOT support streaming (explicitly documented)
	assert.False(t, capabilities.SupportsStreaming)
	// ZAI does NOT support function calling
	assert.False(t, capabilities.SupportsFunctionCalling)
	// ZAI does NOT support vision
	assert.False(t, capabilities.SupportsVision)
	// Check limits
	assert.Greater(t, capabilities.Limits.MaxTokens, 0)
	// Check supported models
	assert.NotEmpty(t, capabilities.SupportedModels)
	assert.Contains(t, capabilities.SupportedModels, "z-ai-base")
	assert.Contains(t, capabilities.SupportedModels, "z-ai-pro")
	// Check supported features
	assert.Contains(t, capabilities.SupportedFeatures, "text_completion")
	assert.Contains(t, capabilities.SupportedFeatures, "chat")
}

func TestZaiProvider_ValidateConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		provider := zai.NewZAIProvider(
			"test-api-key",
			"https://api.zai.com",
			"zai-pro",
		)
		require.NotNil(t, provider)

		valid, errors := provider.ValidateConfig(nil)
		assert.True(t, valid)
		assert.Empty(t, errors)
	})

	t.Run("missing api key", func(t *testing.T) {
		provider := zai.NewZAIProvider(
			"",
			"https://api.zai.com",
			"zai-pro",
		)
		require.NotNil(t, provider)

		valid, errors := provider.ValidateConfig(nil)
		assert.False(t, valid)
		assert.NotEmpty(t, errors)
	})

	t.Run("missing model uses default", func(t *testing.T) {
		// Empty model gets filled with default
		provider := zai.NewZAIProvider(
			"test-api-key",
			"https://api.zai.com",
			"",
		)
		require.NotNil(t, provider)

		valid, errors := provider.ValidateConfig(nil)
		assert.True(t, valid) // Default model is used
		assert.Empty(t, errors)
	})
}

func TestZaiProvider_WithRetry(t *testing.T) {
	retryConfig := zai.RetryConfig{
		MaxRetries:   5,
		InitialDelay: 100,
		MaxDelay:     1000,
		Multiplier:   2.0,
	}
	provider := zai.NewZAIProviderWithRetry("test-api-key", "", "zai-pro", retryConfig)
	require.NotNil(t, provider)
}

func TestZaiProvider_CompleteStream_NotSupported(t *testing.T) {
	provider := zai.NewZAIProvider(
		"test-api-key",
		"https://api.zai.com",
		"zai-pro",
	)
	require.NotNil(t, provider)

	// Streaming is not supported - should return error
	ch, err := provider.CompleteStream(nil, nil)
	assert.Error(t, err)
	assert.Nil(t, ch)
	assert.Contains(t, err.Error(), "streaming not yet implemented")
}

// Integration tests that require external API are skipped
func TestZaiProvider_Complete(t *testing.T) {
	t.Skip("Skipping integration test - requires valid ZAI API endpoint")
}

func TestZaiProvider_HealthCheck(t *testing.T) {
	t.Skip("Skipping integration test - requires valid ZAI API endpoint")
}
