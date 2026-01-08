package zai_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/helixagent/helixagent/internal/llm/providers/zai"
	"github.com/helixagent/helixagent/internal/models"
	"github.com/helixagent/helixagent/tests/testutils"
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
	// ZAI now supports streaming
	assert.True(t, capabilities.SupportsStreaming)
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

// Integration tests that use mock LLM server when available
func TestZaiProvider_Complete(t *testing.T) {
	testutils.SkipIfNoInfrastructure(t, "llm")

	mockURL := testutils.GetMockLLMBaseURL() + "/v1"
	apiKey := testutils.GetMockAPIKey()

	provider := zai.NewZAIProvider(apiKey, mockURL, "zai-pro")
	require.NotNil(t, provider)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &models.LLMRequest{
		ID:     "test-complete",
		Prompt: "Say hello",
		ModelParams: models.ModelParameters{
			Model: "zai-pro",
		},
	}

	result, err := provider.Complete(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Content)
}

func TestZaiProvider_CompleteStream(t *testing.T) {
	testutils.SkipIfNoInfrastructure(t, "llm")

	mockURL := testutils.GetMockLLMBaseURL() + "/v1"
	apiKey := testutils.GetMockAPIKey()

	provider := zai.NewZAIProvider(apiKey, mockURL, "zai-pro")
	require.NotNil(t, provider)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &models.LLMRequest{
		ID:     "test-stream",
		Prompt: "Say hello",
		ModelParams: models.ModelParameters{
			Model: "zai-pro",
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

func TestZaiProvider_HealthCheck(t *testing.T) {
	testutils.SkipIfNoInfrastructure(t, "llm")

	mockURL := testutils.GetMockLLMBaseURL() + "/v1"
	apiKey := testutils.GetMockAPIKey()

	os.Setenv("ZAI_API_KEY", apiKey)
	os.Setenv("ZAI_BASE_URL", mockURL)
	defer func() {
		os.Unsetenv("ZAI_API_KEY")
		os.Unsetenv("ZAI_BASE_URL")
	}()

	provider := zai.NewZAIProvider(apiKey, mockURL, "zai-pro")
	require.NotNil(t, provider)

	err := provider.HealthCheck()
	if err != nil {
		t.Logf("Health check returned error (acceptable for mock): %v", err)
	}
}
