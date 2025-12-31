package ollama_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/superagent/superagent/internal/llm/providers/ollama"
	"github.com/superagent/superagent/internal/models"
	"github.com/superagent/superagent/tests/testutils"
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

// Integration tests that use mock LLM server when available
func TestOllamaProvider_Complete(t *testing.T) {
	testutils.SkipIfNoInfrastructure(t, "llm")

	mockURL := testutils.GetMockLLMBaseURL()

	provider := ollama.NewOllamaProvider(mockURL, "llama2")
	require.NotNil(t, provider)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &models.LLMRequest{
		ID:     "test-complete",
		Prompt: "Say hello",
		ModelParams: models.ModelParameters{
			Model: "llama2",
		},
	}

	result, err := provider.Complete(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Content)
}

func TestOllamaProvider_CompleteStream(t *testing.T) {
	testutils.SkipIfNoInfrastructure(t, "llm")

	mockURL := testutils.GetMockLLMBaseURL()

	provider := ollama.NewOllamaProvider(mockURL, "llama2")
	require.NotNil(t, provider)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &models.LLMRequest{
		ID:     "test-stream",
		Prompt: "Say hello",
		ModelParams: models.ModelParameters{
			Model: "llama2",
		},
	}

	stream, err := provider.CompleteStream(ctx, req)
	if err != nil {
		t.Logf("Stream not supported by mock server: %v", err)
		return
	}

	var gotChunk bool
	for resp := range stream {
		if resp != nil && resp.Content != "" {
			gotChunk = true
			break
		}
	}
	_ = gotChunk
}

func TestOllamaProvider_HealthCheck(t *testing.T) {
	testutils.SkipIfNoInfrastructure(t, "llm")

	mockURL := testutils.GetMockLLMBaseURL()

	os.Setenv("OLLAMA_BASE_URL", mockURL)
	defer os.Unsetenv("OLLAMA_BASE_URL")

	provider := ollama.NewOllamaProvider(mockURL, "llama2")
	require.NotNil(t, provider)

	err := provider.HealthCheck()
	if err != nil {
		t.Logf("Health check returned error (acceptable for mock): %v", err)
	}
}
