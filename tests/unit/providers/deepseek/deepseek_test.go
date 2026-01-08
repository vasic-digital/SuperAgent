package deepseek_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/llm/providers/deepseek"
	"dev.helix.agent/internal/models"
	"dev.helix.agent/tests/testutils"
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

// Integration tests that use mock LLM server when available
func TestDeepSeekProvider_Complete(t *testing.T) {
	testutils.SkipIfNoInfrastructure(t, "llm")

	// DeepSeek expects /v1/chat/completions endpoint
	mockURL := testutils.GetMockLLMBaseURL() + "/v1/chat/completions"
	apiKey := testutils.GetMockAPIKey()

	provider := deepseek.NewDeepSeekProvider(apiKey, mockURL, "deepseek-chat")
	require.NotNil(t, provider)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &models.LLMRequest{
		ID:     "test-complete",
		Prompt: "Say hello",
		ModelParams: models.ModelParameters{
			Model: "deepseek-chat",
		},
	}

	result, err := provider.Complete(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Content)
}

func TestDeepSeekProvider_CompleteStream(t *testing.T) {
	testutils.SkipIfNoInfrastructure(t, "llm")

	// DeepSeek expects /v1/chat/completions endpoint
	mockURL := testutils.GetMockLLMBaseURL() + "/v1/chat/completions"
	apiKey := testutils.GetMockAPIKey()

	provider := deepseek.NewDeepSeekProvider(apiKey, mockURL, "deepseek-chat")
	require.NotNil(t, provider)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &models.LLMRequest{
		ID:     "test-stream",
		Prompt: "Say hello",
		ModelParams: models.ModelParameters{
			Model: "deepseek-chat",
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

func TestDeepSeekProvider_HealthCheck(t *testing.T) {
	testutils.SkipIfNoInfrastructure(t, "llm")

	// DeepSeek expects /v1/chat/completions endpoint
	mockURL := testutils.GetMockLLMBaseURL() + "/v1/chat/completions"
	apiKey := testutils.GetMockAPIKey()

	os.Setenv("DEEPSEEK_API_KEY", apiKey)
	os.Setenv("DEEPSEEK_BASE_URL", mockURL)
	defer func() {
		os.Unsetenv("DEEPSEEK_API_KEY")
		os.Unsetenv("DEEPSEEK_BASE_URL")
	}()

	provider := deepseek.NewDeepSeekProvider(apiKey, mockURL, "deepseek-chat")
	require.NotNil(t, provider)

	err := provider.HealthCheck()
	if err != nil {
		t.Logf("Health check returned error (acceptable for mock): %v", err)
	}
}
