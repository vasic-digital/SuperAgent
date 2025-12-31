package claude_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/superagent/superagent/internal/llm/providers/claude"
)

func TestClaudeProvider_Basic(t *testing.T) {
	provider := claude.NewClaudeProvider("test-api-key", "", "claude-3-opus-20240229")
	require.NotNil(t, provider)
}

func TestClaudeProvider_WithCustomBaseURL(t *testing.T) {
	provider := claude.NewClaudeProvider("test-api-key", "https://custom.api.com", "claude-3-opus-20240229")
	require.NotNil(t, provider)
}

func TestClaudeProvider_WithDefaultModel(t *testing.T) {
	provider := claude.NewClaudeProvider("test-api-key", "", "")
	require.NotNil(t, provider)
}

func TestClaudeProvider_Capabilities(t *testing.T) {
	provider := claude.NewClaudeProvider("test-api-key", "", "claude-3-opus-20240229")
	require.NotNil(t, provider)

	capabilities := provider.GetCapabilities()
	require.NotNil(t, capabilities)

	// Check supported features (actual values from implementation)
	assert.Contains(t, capabilities.SupportedFeatures, "text_completion")
	assert.Contains(t, capabilities.SupportedFeatures, "chat")
	assert.Contains(t, capabilities.SupportedFeatures, "function_calling")
	assert.Contains(t, capabilities.SupportedFeatures, "streaming")

	// Check capability flags
	assert.True(t, capabilities.SupportsStreaming)
	assert.True(t, capabilities.SupportsFunctionCalling)
	assert.True(t, capabilities.SupportsVision)

	// Check limits
	assert.Greater(t, capabilities.Limits.MaxTokens, 0)
	assert.Greater(t, capabilities.Limits.MaxInputLength, 0)
	assert.Greater(t, capabilities.Limits.MaxOutputLength, 0)

	// Check supported models
	assert.NotEmpty(t, capabilities.SupportedModels)
	assert.Contains(t, capabilities.SupportedModels, "claude-3-opus-20240229")
	assert.Contains(t, capabilities.SupportedModels, "claude-3-sonnet-20240229")

	// Check metadata
	assert.NotNil(t, capabilities.Metadata)
}

func TestClaudeProvider_ValidateConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		provider := claude.NewClaudeProvider("test-api-key", "https://api.anthropic.com", "claude-3-opus-20240229")
		require.NotNil(t, provider)

		valid, errors := provider.ValidateConfig(nil)
		assert.True(t, valid)
		assert.Empty(t, errors)
	})

	t.Run("missing api key", func(t *testing.T) {
		provider := claude.NewClaudeProvider("", "https://api.anthropic.com", "claude-3-opus-20240229")
		require.NotNil(t, provider)

		valid, errors := provider.ValidateConfig(nil)
		assert.False(t, valid)
		assert.NotEmpty(t, errors)
	})

	t.Run("missing model uses default", func(t *testing.T) {
		// Provider fills in default model when empty, so this should still be valid
		provider := claude.NewClaudeProvider("test-api-key", "https://api.anthropic.com", "")
		require.NotNil(t, provider)

		valid, errors := provider.ValidateConfig(nil)
		assert.True(t, valid) // Default model is used
		assert.Empty(t, errors)
	})
}

func TestClaudeProvider_WithRetry(t *testing.T) {
	retryConfig := claude.RetryConfig{
		MaxRetries:   5,
		InitialDelay: 100,
		MaxDelay:     1000,
		Multiplier:   2.0,
	}
	provider := claude.NewClaudeProviderWithRetry("test-api-key", "", "claude-3-opus-20240229", retryConfig)
	require.NotNil(t, provider)
}

// Integration tests that require external API are skipped
func TestClaudeProvider_Complete(t *testing.T) {
	t.Skip("Skipping integration test - requires valid Claude API key")
}

func TestClaudeProvider_CompleteStream(t *testing.T) {
	t.Skip("Skipping integration test - requires valid Claude API key")
}

func TestClaudeProvider_HealthCheck(t *testing.T) {
	t.Skip("Skipping integration test - requires valid Claude API key")
}
