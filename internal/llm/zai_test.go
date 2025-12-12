package llm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/superagent/superagent/internal/models"
)

func TestNewZaiProvider(t *testing.T) {
	provider := NewZaiProvider("test-key", "https://test.com", "zephyr")
	require.NotNil(t, provider)
}

func TestZaiProvider_GetCapabilities(t *testing.T) {
	provider := NewZaiProvider("test-key", "", "")
	require.NotNil(t, provider)

	caps := provider.GetCapabilities()
	require.NotNil(t, caps)

	assert.Contains(t, caps.SupportedModels, "zephyr")
	assert.Contains(t, caps.SupportedModels, "mistral")
	assert.Contains(t, caps.SupportedModels, "llama")
	assert.Contains(t, caps.SupportedFeatures, "streaming")
	assert.Contains(t, caps.SupportedFeatures, "function_calling")
	assert.Contains(t, caps.SupportedRequestTypes, "text_completion")
	assert.Contains(t, caps.SupportedRequestTypes, "chat")
	assert.True(t, caps.SupportsStreaming)
	assert.True(t, caps.SupportsFunctionCalling)
	assert.False(t, caps.SupportsVision)
	assert.True(t, caps.SupportsTools)
	assert.True(t, caps.SupportsSearch)
	assert.True(t, caps.SupportsReasoning)
	assert.True(t, caps.SupportsCodeCompletion)
	assert.True(t, caps.SupportsCodeAnalysis)
	assert.True(t, caps.SupportsRefactoring)
	assert.Equal(t, 4096, caps.Limits.MaxTokens)
	assert.Equal(t, 4096, caps.Limits.MaxInputLength)
	assert.Equal(t, 2048, caps.Limits.MaxOutputLength)
	assert.Equal(t, 10, caps.Limits.MaxConcurrentRequests)
	assert.Equal(t, "zai", caps.Metadata["provider"])
	assert.Equal(t, "1.0", caps.Metadata["version"])
}

func TestZaiProvider_ValidateConfig(t *testing.T) {
	provider := NewZaiProvider("test-key", "", "")
	require.NotNil(t, provider)

	valid, errors := provider.ValidateConfig(map[string]interface{}{
		"api_key": "test-key",
		"model":   "zephyr",
	})
	assert.True(t, valid)
	assert.Empty(t, errors)
}

func TestZaiProvider_HealthCheck_Error(t *testing.T) {
	// Create provider with nil internal provider to trigger error
	provider := &ZaiProvider{}
	err := provider.HealthCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "zai provider not initialized")
}

func TestZaiProvider_Complete_Error_NilProvider(t *testing.T) {
	// Create provider with nil internal provider to trigger error
	provider := &ZaiProvider{}
	req := &models.LLMRequest{
		ID: "test-request",
	}
	resp, err := provider.Complete(req)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "zai provider not initialized")
}

func TestZaiProvider_Complete_Error_NilRequest(t *testing.T) {
	provider := NewZaiProvider("test-key", "", "")
	require.NotNil(t, provider)

	resp, err := provider.Complete(nil)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "request cannot be nil")
}
