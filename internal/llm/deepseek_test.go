package llm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/superagent/superagent/internal/models"
)

func TestNewDeepSeekProvider(t *testing.T) {
	provider := NewDeepSeekProvider("test-key", "https://test.com", "deepseek-test")
	require.NotNil(t, provider)
}

func TestDeepSeekProvider_GetCapabilities(t *testing.T) {
	provider := NewDeepSeekProvider("test-key", "", "")
	require.NotNil(t, provider)

	caps := provider.GetCapabilities()
	require.NotNil(t, caps)

	assert.Contains(t, caps.SupportedModels, "deepseek-coder")
	assert.Contains(t, caps.SupportedModels, "deepseek-chat")
	assert.Contains(t, caps.SupportedFeatures, "streaming")
	assert.Contains(t, caps.SupportedFeatures, "coding")
	assert.Contains(t, caps.SupportedFeatures, "reasoning")
	assert.Contains(t, caps.SupportedRequestTypes, "code_generation")
	assert.Contains(t, caps.SupportedRequestTypes, "text_completion")
	assert.True(t, caps.SupportsStreaming)
	assert.False(t, caps.SupportsFunctionCalling)
	assert.False(t, caps.SupportsVision)
	assert.True(t, caps.SupportsTools)
	assert.True(t, caps.SupportsReasoning)
	assert.True(t, caps.SupportsCodeCompletion)
	assert.True(t, caps.SupportsCodeAnalysis)
	assert.True(t, caps.SupportsRefactoring)
}

func TestDeepSeekProvider_ValidateConfig(t *testing.T) {
	provider := NewDeepSeekProvider("test-key", "", "")
	require.NotNil(t, provider)

	valid, errors := provider.ValidateConfig(map[string]interface{}{
		"api_key": "test-key",
		"model":   "deepseek-chat",
	})
	assert.True(t, valid)
	assert.Empty(t, errors)
}

func TestDeepSeekProvider_HealthCheck_Error(t *testing.T) {
	// Create provider with nil internal provider to trigger error
	provider := &DeepSeekProvider{}
	err := provider.HealthCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DeepSeek provider not initialized")
}

func TestDeepSeekProvider_Complete_Error(t *testing.T) {
	// Create provider with nil internal provider to trigger error
	provider := &DeepSeekProvider{}
	req := &models.LLMRequest{
		ID: "test-request",
	}
	resp, err := provider.Complete(req)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DeepSeek provider not initialized")
}
