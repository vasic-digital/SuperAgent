package gemini_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/superagent/superagent/internal/llm/providers/gemini"
)

func TestNewGeminiProvider(t *testing.T) {
	t.Run("valid configuration", func(t *testing.T) {
		provider := gemini.NewGeminiProvider(
			"test-api-key",
			"https://generativelanguage.googleapis.com",
			"gemini-pro",
		)
		require.NotNil(t, provider)
	})

	t.Run("with default base URL", func(t *testing.T) {
		provider := gemini.NewGeminiProvider(
			"test-api-key",
			"",
			"gemini-pro",
		)
		require.NotNil(t, provider)
	})

	t.Run("with default model", func(t *testing.T) {
		provider := gemini.NewGeminiProvider(
			"test-api-key",
			"https://generativelanguage.googleapis.com",
			"",
		)
		require.NotNil(t, provider)
	})
}

func TestGeminiProvider_GetCapabilities(t *testing.T) {
	provider := gemini.NewGeminiProvider(
		"test-api-key",
		"https://generativelanguage.googleapis.com",
		"gemini-pro",
	)
	require.NotNil(t, provider)

	capabilities := provider.GetCapabilities()

	assert.NotNil(t, capabilities)
	assert.True(t, capabilities.SupportsStreaming)
	assert.Greater(t, capabilities.Limits.MaxTokens, 0)
	assert.True(t, capabilities.SupportsFunctionCalling)
	assert.True(t, capabilities.SupportsVision)
	assert.NotEmpty(t, capabilities.SupportedModels)
	assert.Contains(t, capabilities.SupportedModels, "gemini-pro")
	assert.Contains(t, capabilities.SupportedModels, "gemini-pro-vision")
	assert.Contains(t, capabilities.SupportedModels, "gemini-1.5-pro")

	// Check supported features
	assert.Contains(t, capabilities.SupportedFeatures, "text_completion")
	assert.Contains(t, capabilities.SupportedFeatures, "chat")
	assert.Contains(t, capabilities.SupportedFeatures, "function_calling")
	assert.Contains(t, capabilities.SupportedFeatures, "streaming")
	assert.Contains(t, capabilities.SupportedFeatures, "vision")
}

func TestGeminiProvider_ValidateConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		provider := gemini.NewGeminiProvider(
			"test-api-key",
			"https://generativelanguage.googleapis.com",
			"gemini-pro",
		)
		require.NotNil(t, provider)

		valid, errors := provider.ValidateConfig(nil)
		assert.True(t, valid)
		assert.Empty(t, errors)
	})

	t.Run("missing api key", func(t *testing.T) {
		provider := gemini.NewGeminiProvider(
			"",
			"https://generativelanguage.googleapis.com",
			"gemini-pro",
		)
		require.NotNil(t, provider)

		valid, errors := provider.ValidateConfig(nil)
		assert.False(t, valid)
		assert.NotEmpty(t, errors)
	})

	t.Run("missing model uses default", func(t *testing.T) {
		provider := gemini.NewGeminiProvider(
			"test-api-key",
			"https://generativelanguage.googleapis.com",
			"",
		)
		require.NotNil(t, provider)

		valid, errors := provider.ValidateConfig(nil)
		assert.True(t, valid) // Default model is used
		assert.Empty(t, errors)
	})
}

func TestGeminiProvider_WithRetry(t *testing.T) {
	retryConfig := gemini.RetryConfig{
		MaxRetries:   5,
		InitialDelay: 100,
		MaxDelay:     1000,
		Multiplier:   2.0,
	}
	provider := gemini.NewGeminiProviderWithRetry("test-api-key", "", "gemini-pro", retryConfig)
	require.NotNil(t, provider)
}

// Integration tests that require external API are skipped
func TestGeminiProvider_Complete(t *testing.T) {
	t.Skip("Skipping integration test - requires valid Gemini API endpoint")
}

func TestGeminiProvider_CompleteStream(t *testing.T) {
	t.Skip("Skipping integration test - requires valid Gemini API endpoint")
}

func TestGeminiProvider_HealthCheck(t *testing.T) {
	t.Skip("Skipping integration test - requires valid Gemini API endpoint")
}
