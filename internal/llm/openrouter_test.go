package llm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/superagent/superagent/internal/models"
)

func TestNewOpenRouterProvider(t *testing.T) {
	provider := NewOpenRouterProvider("test-key")
	require.NotNil(t, provider)
}

func TestOpenRouterProvider_GetCapabilities(t *testing.T) {
	provider := NewOpenRouterProvider("test-key")
	require.NotNil(t, provider)

	caps := provider.GetCapabilities()
	require.NotNil(t, caps)

	// Check some expected models
	assert.Contains(t, caps.SupportedModels, "x-ai/grok-4")
	assert.Contains(t, caps.SupportedModels, "google/gemini-2.0-flash-exp")
	assert.Contains(t, caps.SupportedModels, "anthropic/claude-3.5-sonnet")
	assert.Contains(t, caps.SupportedModels, "openai/gpt-4o")
	assert.Contains(t, caps.SupportedModels, "meta-llama/llama-3.1-405b-instruct")

	// Check features
	assert.Contains(t, caps.SupportedFeatures, "text-generation")
	assert.Contains(t, caps.SupportedFeatures, "code-generation")
	assert.Contains(t, caps.SupportedFeatures, "reasoning")
	assert.Contains(t, caps.SupportedFeatures, "function-calling")
	assert.Contains(t, caps.SupportedFeatures, "multi-turn")

	// Check request types
	assert.Contains(t, caps.SupportedRequestTypes, "chat")
	assert.Contains(t, caps.SupportedRequestTypes, "completion")

	// Check capabilities
	assert.True(t, caps.SupportsStreaming)
	assert.True(t, caps.SupportsFunctionCalling)
	assert.False(t, caps.SupportsVision)

	// Check metadata
	assert.Equal(t, "openrouter", caps.Metadata["provider"])
	assert.Equal(t, "100+ models available", caps.Metadata["models"])
}

func TestOpenRouterProvider_ValidateConfig(t *testing.T) {
	provider := NewOpenRouterProvider("test-key")
	require.NotNil(t, provider)

	// Valid config
	valid, errors := provider.ValidateConfig(map[string]interface{}{
		"api_key": "test-key",
	})
	assert.True(t, valid)
	assert.Empty(t, errors)

	// Valid config with optional base_url (should generate warning)
	valid, warnings := provider.ValidateConfig(map[string]interface{}{
		"api_key":  "test-key",
		"base_url": "https://custom.openrouter.ai",
	})
	assert.True(t, valid)
	assert.NotEmpty(t, warnings)
	assert.Contains(t, warnings[0], "Custom base URL may not be supported")

	// Invalid config - missing API key
	valid, errors = provider.ValidateConfig(map[string]interface{}{})
	assert.False(t, valid)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0], "API key is required")

	// Invalid config - empty API key
	valid, errors = provider.ValidateConfig(map[string]interface{}{
		"api_key": "",
	})
	assert.False(t, valid)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0], "API key is required")
}

func TestOpenRouterProvider_Complete(t *testing.T) {
	provider := NewOpenRouterProvider("test-key")
	require.NotNil(t, provider)

	// Note: Complete method delegates to underlying provider
	// We can't test the actual completion without mocking
	// But we can verify the method exists and doesn't panic
	req := &models.LLMRequest{
		ID: "test-request",
	}
	// This will likely fail due to network/authentication, but that's expected
	// We're just testing that the method exists
	provider.Complete(req)
}

func TestOpenRouterProvider_HealthCheck(t *testing.T) {
	provider := NewOpenRouterProvider("test-key")
	require.NotNil(t, provider)

	// HealthCheck delegates to underlying provider
	// We can't test the actual health check without mocking
	// But we can verify the method exists
	provider.HealthCheck()
}
