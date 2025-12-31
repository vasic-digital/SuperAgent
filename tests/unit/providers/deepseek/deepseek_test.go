package deepseek_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/superagent/superagent/internal/llm/providers/deepseek"
)

func TestNewDeepSeekProvider(t *testing.T) {
	t.Run("valid configuration", func(t *testing.T) {
		provider := deepseek.NewDeepSeekProvider(
			"test-api-key",
			"https://api.deepseek.com",
			"deepseek-chat",
		)
		require.NotNil(t, provider)
	})

	t.Run("with default base URL", func(t *testing.T) {
		provider := deepseek.NewDeepSeekProvider(
			"test-api-key",
			"",
			"deepseek-chat",
		)
		require.NotNil(t, provider)
	})

	t.Run("with default model", func(t *testing.T) {
		provider := deepseek.NewDeepSeekProvider(
			"test-api-key",
			"https://api.deepseek.com",
			"",
		)
		require.NotNil(t, provider)
	})
}

func TestDeepSeekProvider_GetCapabilities(t *testing.T) {
	provider := deepseek.NewDeepSeekProvider(
		"test-api-key",
		"https://api.deepseek.com",
		"deepseek-chat",
	)
	require.NotNil(t, provider)

	capabilities := provider.GetCapabilities()

	assert.NotNil(t, capabilities)
	assert.True(t, capabilities.SupportsStreaming)
	assert.Greater(t, capabilities.Limits.MaxTokens, 0)
	assert.True(t, capabilities.SupportsFunctionCalling)
	assert.False(t, capabilities.SupportsVision) // DeepSeek does not support vision
	assert.NotEmpty(t, capabilities.SupportedModels)
	assert.Contains(t, capabilities.SupportedModels, "deepseek-chat")
	assert.Contains(t, capabilities.SupportedModels, "deepseek-coder")

	// Check supported features
	assert.Contains(t, capabilities.SupportedFeatures, "text_completion")
	assert.Contains(t, capabilities.SupportedFeatures, "chat")
	assert.Contains(t, capabilities.SupportedFeatures, "function_calling")
	assert.Contains(t, capabilities.SupportedFeatures, "streaming")
}

func TestDeepSeekProvider_ValidateConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		provider := deepseek.NewDeepSeekProvider(
			"test-api-key",
			"https://api.deepseek.com",
			"deepseek-chat",
		)
		require.NotNil(t, provider)

		valid, errors := provider.ValidateConfig(nil)
		assert.True(t, valid)
		assert.Empty(t, errors)
	})

	t.Run("missing api key", func(t *testing.T) {
		provider := deepseek.NewDeepSeekProvider(
			"",
			"https://api.deepseek.com",
			"deepseek-chat",
		)
		require.NotNil(t, provider)

		valid, errors := provider.ValidateConfig(nil)
		assert.False(t, valid)
		assert.NotEmpty(t, errors)
	})

	t.Run("missing model uses default", func(t *testing.T) {
		provider := deepseek.NewDeepSeekProvider(
			"test-api-key",
			"https://api.deepseek.com",
			"",
		)
		require.NotNil(t, provider)

		valid, errors := provider.ValidateConfig(nil)
		assert.True(t, valid) // Default model is used
		assert.Empty(t, errors)
	})
}

func TestDeepSeekProvider_WithRetry(t *testing.T) {
	retryConfig := deepseek.RetryConfig{
		MaxRetries:   5,
		InitialDelay: 100,
		MaxDelay:     1000,
		Multiplier:   2.0,
	}
	provider := deepseek.NewDeepSeekProviderWithRetry("test-api-key", "", "deepseek-chat", retryConfig)
	require.NotNil(t, provider)
}

// Integration tests that require external API are skipped
func TestDeepSeekProvider_Complete(t *testing.T) {
	t.Skip("Skipping integration test - requires valid DeepSeek API endpoint")
}

func TestDeepSeekProvider_CompleteStream(t *testing.T) {
	t.Skip("Skipping integration test - requires valid DeepSeek API endpoint")
}

func TestDeepSeekProvider_HealthCheck(t *testing.T) {
	t.Skip("Skipping integration test - requires valid DeepSeek API endpoint")
}
