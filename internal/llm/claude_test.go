package llm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/superagent/superagent/internal/models"
)

func TestNewClaudeProvider(t *testing.T) {
	provider := NewClaudeProvider("test-key", "https://test.com", "claude-3-test")
	require.NotNil(t, provider)
}

func TestClaudeProvider_GetCapabilities(t *testing.T) {
	provider := NewClaudeProvider("test-key", "", "")
	require.NotNil(t, provider)

	caps := provider.GetCapabilities()
	require.NotNil(t, caps)

	assert.Contains(t, caps.SupportedModels, "claude-3-sonnet-20240229")
	assert.Contains(t, caps.SupportedModels, "claude-3-opus-20240229")
	assert.Contains(t, caps.SupportedFeatures, "streaming")
	assert.Contains(t, caps.SupportedFeatures, "function_calling")
	assert.Contains(t, caps.SupportedRequestTypes, "text_completion")
	assert.Contains(t, caps.SupportedRequestTypes, "chat")
	assert.True(t, caps.SupportsStreaming)
	assert.True(t, caps.SupportsFunctionCalling)
	assert.False(t, caps.SupportsVision)
}

func TestClaudeProvider_ValidateConfig(t *testing.T) {
	provider := NewClaudeProvider("test-key", "", "")
	require.NotNil(t, provider)

	valid, errors := provider.ValidateConfig(map[string]interface{}{
		"api_key": "test-key",
		"model":   "claude-3-sonnet-20240229",
	})
	assert.True(t, valid)
	assert.Empty(t, errors)
}

func TestClaudeProvider_HealthCheck_Error(t *testing.T) {
	// Create provider with nil internal provider to trigger error
	provider := &ClaudeProvider{}
	err := provider.HealthCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Claude provider not initialized")
}

func TestClaudeProvider_Complete_Error(t *testing.T) {
	// Create provider with nil internal provider to trigger error
	provider := &ClaudeProvider{}
	req := &models.LLMRequest{
		ID: "test-request",
	}
	resp, err := provider.Complete(req)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Claude provider not initialized")
}
