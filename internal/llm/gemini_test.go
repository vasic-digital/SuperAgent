package llm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/superagent/superagent/internal/models"
)

func TestNewGeminiProvider(t *testing.T) {
	provider := NewGeminiProvider("test-key", "https://test.com", "gemini-test")
	require.NotNil(t, provider)
}

func TestGeminiProvider_GetCapabilities(t *testing.T) {
	provider := NewGeminiProvider("test-key", "", "")
	require.NotNil(t, provider)

	caps := provider.GetCapabilities()
	require.NotNil(t, caps)

	assert.Contains(t, caps.SupportedModels, "gemini-pro")
	assert.Contains(t, caps.SupportedModels, "gemini-pro-vision")
	assert.Contains(t, caps.SupportedFeatures, "streaming")
	assert.Contains(t, caps.SupportedFeatures, "vision")
	assert.Contains(t, caps.SupportedRequestTypes, "text_completion")
	assert.Contains(t, caps.SupportedRequestTypes, "chat")
	assert.Contains(t, caps.SupportedRequestTypes, "multimodal")
	assert.True(t, caps.SupportsStreaming)
	assert.False(t, caps.SupportsFunctionCalling)
	assert.True(t, caps.SupportsVision)
}

func TestGeminiProvider_ValidateConfig(t *testing.T) {
	provider := NewGeminiProvider("test-key", "", "")
	require.NotNil(t, provider)

	valid, errors := provider.ValidateConfig(map[string]interface{}{
		"api_key": "test-key",
		"model":   "gemini-pro",
	})
	assert.True(t, valid)
	assert.Empty(t, errors)
}

func TestGeminiProvider_HealthCheck_Error(t *testing.T) {
	// Create provider with nil internal provider to trigger error
	provider := &GeminiProvider{}
	err := provider.HealthCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Gemini provider not initialized")
}

func TestGeminiProvider_Complete_Error(t *testing.T) {
	// Create provider with nil internal provider to trigger error
	provider := &GeminiProvider{}
	req := &models.LLMRequest{
		ID: "test-request",
	}
	resp, err := provider.Complete(req)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Gemini provider not initialized")
}
