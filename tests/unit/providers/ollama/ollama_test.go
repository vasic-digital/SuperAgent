package ollama_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/superagent/superagent/internal/llm/providers/ollama"
)

func TestNewOllamaProvider(t *testing.T) {
	t.Run("valid configuration", func(t *testing.T) {
		provider := ollama.NewOllamaProvider(
			"http://localhost:11434",
			"llama2",
		)
		require.NotNil(t, provider)
	})

	t.Run("with default model", func(t *testing.T) {
		provider := ollama.NewOllamaProvider(
			"http://localhost:11434",
			"",
		)
		require.NotNil(t, provider)
	})
}

func TestOllamaProvider_GetCapabilities(t *testing.T) {
	provider := ollama.NewOllamaProvider(
		"http://localhost:11434",
		"llama2",
	)
	require.NotNil(t, provider)

	capabilities := provider.GetCapabilities()

	assert.NotNil(t, capabilities)
	assert.True(t, capabilities.SupportsStreaming)
	assert.Greater(t, capabilities.Limits.MaxTokens, 0)
	// Ollama does NOT support function calling
	assert.False(t, capabilities.SupportsFunctionCalling)
	assert.False(t, capabilities.SupportsVision)
	assert.NotEmpty(t, capabilities.SupportedModels)
	assert.Contains(t, capabilities.SupportedModels, "llama2")
	assert.Contains(t, capabilities.SupportedModels, "codellama")
	assert.Contains(t, capabilities.SupportedModels, "mistral")

	// Check supported features
	assert.Contains(t, capabilities.SupportedFeatures, "text_completion")
	assert.Contains(t, capabilities.SupportedFeatures, "chat")
	assert.Contains(t, capabilities.SupportedFeatures, "streaming")
}

func TestOllamaProvider_ValidateConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		provider := ollama.NewOllamaProvider(
			"http://localhost:11434",
			"llama2",
		)
		require.NotNil(t, provider)

		valid, errors := provider.ValidateConfig(nil)
		assert.True(t, valid)
		assert.Empty(t, errors)
	})

	t.Run("missing model uses default", func(t *testing.T) {
		provider := ollama.NewOllamaProvider(
			"http://localhost:11434",
			"",
		)
		require.NotNil(t, provider)

		valid, errors := provider.ValidateConfig(nil)
		assert.True(t, valid) // Default model is used
		assert.Empty(t, errors)
	})
}

// Integration tests that require running Ollama are skipped
func TestOllamaProvider_Complete(t *testing.T) {
	t.Skip("Skipping integration test - requires running Ollama server")
}

func TestOllamaProvider_CompleteStream(t *testing.T) {
	t.Skip("Skipping integration test - requires running Ollama server")
}

func TestOllamaProvider_HealthCheck(t *testing.T) {
	t.Skip("Skipping integration test - requires running Ollama server")
}
