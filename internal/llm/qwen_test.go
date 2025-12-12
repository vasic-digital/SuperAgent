package llm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/superagent/superagent/internal/models"
)

func TestNewQwenProvider(t *testing.T) {
	provider := NewQwenProvider("test-key", "https://test.com", "qwen-turbo")
	require.NotNil(t, provider)
}

func TestQwenProvider_GetCapabilities(t *testing.T) {
	provider := NewQwenProvider("test-key", "", "")
	require.NotNil(t, provider)

	caps := provider.GetCapabilities()
	require.NotNil(t, caps)

	assert.Contains(t, caps.SupportedModels, "qwen-turbo")
	assert.Contains(t, caps.SupportedModels, "qwen-plus")
	assert.Contains(t, caps.SupportedModels, "qwen-max")
	assert.Contains(t, caps.SupportedFeatures, "streaming")
	assert.Contains(t, caps.SupportedFeatures, "function_calling")
	assert.Contains(t, caps.SupportedFeatures, "vision")
	assert.Contains(t, caps.SupportedRequestTypes, "text_completion")
	assert.Contains(t, caps.SupportedRequestTypes, "chat")
	assert.Contains(t, caps.SupportedRequestTypes, "image_analysis")
	assert.True(t, caps.SupportsStreaming)
	assert.True(t, caps.SupportsFunctionCalling)
	assert.True(t, caps.SupportsVision)
	assert.True(t, caps.SupportsTools)
	assert.True(t, caps.SupportsReasoning)
	assert.True(t, caps.SupportsCodeCompletion)
	assert.True(t, caps.SupportsCodeAnalysis)
	assert.True(t, caps.SupportsRefactoring)
	assert.Equal(t, 8192, caps.Limits.MaxTokens)
	assert.Equal(t, 8192, caps.Limits.MaxInputLength)
	assert.Equal(t, 4096, caps.Limits.MaxOutputLength)
	assert.Equal(t, 5, caps.Limits.MaxConcurrentRequests)
	assert.Equal(t, "qwen", caps.Metadata["provider"])
	assert.Equal(t, "1.0", caps.Metadata["version"])
}

func TestQwenProvider_ValidateConfig(t *testing.T) {
	provider := NewQwenProvider("test-key", "", "")
	require.NotNil(t, provider)

	valid, errors := provider.ValidateConfig(map[string]interface{}{
		"api_key": "test-key",
		"model":   "qwen-turbo",
	})
	assert.True(t, valid)
	assert.Empty(t, errors)
}

func TestQwenProvider_HealthCheck_Error(t *testing.T) {
	// Create provider with nil internal provider to trigger error
	provider := &QwenProvider{}
	err := provider.HealthCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "qwen provider not initialized")
}

func TestQwenProvider_Complete_Error_NilProvider(t *testing.T) {
	// Create provider with nil internal provider to trigger error
	provider := &QwenProvider{}
	req := &models.LLMRequest{
		ID: "test-request",
	}
	resp, err := provider.Complete(req)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "qwen provider not initialized")
}

func TestQwenProvider_Complete_Error_NilRequest(t *testing.T) {
	provider := NewQwenProvider("test-key", "", "")
	require.NotNil(t, provider)

	resp, err := provider.Complete(nil)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "request cannot be nil")
}
