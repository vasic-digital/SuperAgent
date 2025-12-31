package qwen_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/superagent/superagent/internal/llm/providers/qwen"
)

func TestNewQwenProvider(t *testing.T) {
	t.Run("valid configuration", func(t *testing.T) {
		provider := qwen.NewQwenProvider(
			"test-api-key",
			"https://dashscope.aliyuncs.com",
			"qwen-turbo",
		)
		require.NotNil(t, provider)
	})

	t.Run("with default base URL", func(t *testing.T) {
		provider := qwen.NewQwenProvider(
			"test-api-key",
			"",
			"qwen-turbo",
		)
		require.NotNil(t, provider)
	})

	t.Run("with default model", func(t *testing.T) {
		provider := qwen.NewQwenProvider(
			"test-api-key",
			"https://dashscope.aliyuncs.com",
			"",
		)
		require.NotNil(t, provider)
	})
}

func TestQwenProvider_GetCapabilities(t *testing.T) {
	provider := qwen.NewQwenProvider(
		"test-api-key",
		"https://dashscope.aliyuncs.com",
		"qwen-turbo",
	)
	require.NotNil(t, provider)

	capabilities := provider.GetCapabilities()

	assert.NotNil(t, capabilities)
	assert.True(t, capabilities.SupportsStreaming)
	assert.Greater(t, capabilities.Limits.MaxTokens, 0)
	assert.True(t, capabilities.SupportsFunctionCalling)
	// Qwen does NOT support vision
	assert.False(t, capabilities.SupportsVision)
	assert.NotEmpty(t, capabilities.SupportedModels)
	assert.Contains(t, capabilities.SupportedModels, "qwen-turbo")
	assert.Contains(t, capabilities.SupportedModels, "qwen-plus")
	assert.Contains(t, capabilities.SupportedModels, "qwen-max")

	// Check supported features
	assert.Contains(t, capabilities.SupportedFeatures, "text_completion")
	assert.Contains(t, capabilities.SupportedFeatures, "chat")
	assert.Contains(t, capabilities.SupportedFeatures, "function_calling")
}

func TestQwenProvider_ValidateConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		provider := qwen.NewQwenProvider(
			"test-api-key",
			"https://dashscope.aliyuncs.com",
			"qwen-turbo",
		)
		require.NotNil(t, provider)

		valid, errors := provider.ValidateConfig(nil)
		assert.True(t, valid)
		assert.Empty(t, errors)
	})

	t.Run("missing api key", func(t *testing.T) {
		provider := qwen.NewQwenProvider(
			"",
			"https://dashscope.aliyuncs.com",
			"qwen-turbo",
		)
		require.NotNil(t, provider)

		valid, errors := provider.ValidateConfig(nil)
		assert.False(t, valid)
		assert.NotEmpty(t, errors)
	})

	t.Run("missing model uses default", func(t *testing.T) {
		provider := qwen.NewQwenProvider(
			"test-api-key",
			"https://dashscope.aliyuncs.com",
			"",
		)
		require.NotNil(t, provider)

		valid, errors := provider.ValidateConfig(nil)
		assert.True(t, valid) // Default model is used
		assert.Empty(t, errors)
	})
}

func TestQwenProvider_WithRetry(t *testing.T) {
	retryConfig := qwen.RetryConfig{
		MaxRetries:   5,
		InitialDelay: 100,
		MaxDelay:     1000,
		Multiplier:   2.0,
	}
	provider := qwen.NewQwenProviderWithRetry("test-api-key", "", "qwen-turbo", retryConfig)
	require.NotNil(t, provider)
}

// Integration tests that require external API are skipped
func TestQwenProvider_Complete(t *testing.T) {
	t.Skip("Skipping integration test - requires valid Qwen API endpoint")
}

func TestQwenProvider_CompleteStream(t *testing.T) {
	t.Skip("Skipping integration test - requires valid Qwen API endpoint")
}

func TestQwenProvider_HealthCheck(t *testing.T) {
	t.Skip("Skipping integration test - requires valid Qwen API endpoint")
}
