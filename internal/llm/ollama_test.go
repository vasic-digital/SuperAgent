package llm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/superagent/superagent/internal/models"
)

func TestNewOllamaProvider(t *testing.T) {
	provider := NewOllamaProvider("http://localhost:11434", "llama2")
	require.NotNil(t, provider)
}

func TestOllamaProvider_GetCapabilities(t *testing.T) {
	provider := NewOllamaProvider("http://localhost:11434", "llama2")
	require.NotNil(t, provider)

	caps := provider.GetCapabilities()
	require.NotNil(t, caps)

	// Ollama capabilities come from the underlying provider
	// We can't assert specific values since they're dynamic
	assert.NotNil(t, caps.SupportedModels)
	assert.NotNil(t, caps.SupportedFeatures)
	assert.NotNil(t, caps.SupportedRequestTypes)
}

func TestOllamaProvider_GetCapabilities_NilProvider(t *testing.T) {
	// Create provider with nil internal provider
	provider := &OllamaProvider{}
	caps := provider.GetCapabilities()
	require.NotNil(t, caps)

	// Should return empty capabilities when provider is nil
	assert.Empty(t, caps.SupportedModels)
	assert.Empty(t, caps.SupportedFeatures)
	assert.Empty(t, caps.SupportedRequestTypes)
	assert.False(t, caps.SupportsStreaming)
	assert.False(t, caps.SupportsFunctionCalling)
	assert.False(t, caps.SupportsVision)
}

func TestOllamaProvider_ValidateConfig(t *testing.T) {
	provider := NewOllamaProvider("http://localhost:11434", "llama2")
	require.NotNil(t, provider)

	valid, errors := provider.ValidateConfig(map[string]interface{}{
		"base_url": "http://localhost:11434",
		"model":    "llama2",
	})
	assert.True(t, valid)
	assert.Empty(t, errors)
}

func TestOllamaProvider_ValidateConfig_NilProvider(t *testing.T) {
	// Create provider with nil internal provider
	provider := &OllamaProvider{}
	valid, errors := provider.ValidateConfig(map[string]interface{}{
		"base_url": "http://localhost:11434",
		"model":    "llama2",
	})
	assert.False(t, valid)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0], "Ollama provider not initialized")
}

func TestOllamaProvider_HealthCheck_Error(t *testing.T) {
	// Create provider with nil internal provider to trigger error
	provider := &OllamaProvider{}
	err := provider.HealthCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Ollama provider not initialized")
}

func TestOllamaProvider_Complete_Error(t *testing.T) {
	// Create provider with nil internal provider to trigger error
	provider := &OllamaProvider{}
	req := &models.LLMRequest{
		ID: "test-request",
	}
	resp, err := provider.Complete(req)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Ollama provider not initialized")
}
