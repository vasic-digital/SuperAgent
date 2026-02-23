package testutils

import (
	"testing"
	"time"

	"dev.helix.agent/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTestProviderNames(t *testing.T) {
	assert.NotEmpty(t, TestProviderNames)
	assert.Contains(t, TestProviderNames, "claude")
	assert.Contains(t, TestProviderNames, "deepseek")
}

func TestTestProviderCapabilities(t *testing.T) {
	caps := TestProviderCapabilities
	require.NotNil(t, caps)

	assert.True(t, caps.SupportsStreaming)
	assert.True(t, caps.SupportsFunctionCalling)
	assert.False(t, caps.SupportsVision)
	assert.Equal(t, 4096, caps.Limits.MaxTokens)
}

func TestTestUser(t *testing.T) {
	user := TestUser
	require.NotNil(t, user)
	assert.Equal(t, "test-user-1", user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "user", user.Role)
}

func TestTestAdminUser(t *testing.T) {
	admin := TestAdminUser
	require.NotNil(t, admin)
	assert.Equal(t, "test-admin-1", admin.ID)
	assert.Equal(t, "admin@example.com", admin.Email)
	assert.Equal(t, "admin", admin.Role)
}

func TestNewTestCompletionRequest(t *testing.T) {
	req := NewTestCompletionRequest()
	require.NotNil(t, req)
	assert.Equal(t, "Test prompt for completion", req["prompt"])
	assert.Equal(t, "test-model", req["model"])
	assert.Equal(t, 100, req["max_tokens"])
}

func TestNewTestChatCompletionRequest(t *testing.T) {
	req := NewTestChatCompletionRequest()
	require.NotNil(t, req)
	assert.Equal(t, "test-model", req["model"])

	messages, ok := req["messages"].([]map[string]string)
	require.True(t, ok)
	assert.Len(t, messages, 2)
	assert.Equal(t, "system", messages[0]["role"])
	assert.Equal(t, "user", messages[1]["role"])
}

func TestNewTestEnsembleRequest(t *testing.T) {
	req := NewTestEnsembleRequest()
	require.NotNil(t, req)
	assert.Equal(t, "Test ensemble prompt", req["prompt"])

	ensembleConfig, ok := req["ensemble_config"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "confidence_weighted", ensembleConfig["strategy"])
	assert.Equal(t, 2, ensembleConfig["min_providers"])
}

func TestNewTestLLMRequest(t *testing.T) {
	req := NewTestLLMRequest("test prompt")
	require.NotNil(t, req)
	assert.Contains(t, req.ID, "test-request-")
	assert.Equal(t, "test prompt", req.Prompt)
	assert.Equal(t, "pending", req.Status)
	assert.Equal(t, "completion", req.RequestType)
}

func TestNewTestLLMResponse(t *testing.T) {
	resp := NewTestLLMResponse("req-123", "test content")
	require.NotNil(t, resp)
	assert.Equal(t, "response-req-123", resp.ID)
	assert.Equal(t, "req-123", resp.RequestID)
	assert.Equal(t, "test content", resp.Content)
	assert.Equal(t, 0.9, resp.Confidence)
	assert.True(t, resp.Selected)
}

func TestNewTestRegistrationRequest(t *testing.T) {
	req := NewTestRegistrationRequest()
	require.NotNil(t, req)
	assert.Equal(t, "newuser@example.com", req["email"])
	assert.Equal(t, "SecurePassword123!", req["password"])
}

func TestNewTestLoginRequest(t *testing.T) {
	req := NewTestLoginRequest()
	require.NotNil(t, req)
	assert.Equal(t, "test@example.com", req["email"])
	assert.Equal(t, "testpassword", req["password"])
}

func TestTokenConstants(t *testing.T) {
	assert.NotEmpty(t, TestValidToken)
	assert.NotEmpty(t, TestExpiredToken)
	assert.NotEmpty(t, TestInvalidToken)
	assert.NotEmpty(t, TestAPIKey)
	assert.NotEqual(t, TestValidToken, TestExpiredToken)
}

func TestNewTestEmbeddingRequest(t *testing.T) {
	req := NewTestEmbeddingRequest()
	require.NotNil(t, req)
	assert.Equal(t, "Test text for embedding generation", req["input"])
	assert.Equal(t, "test-embedding-model", req["model"])
}

func TestNewTestEmbeddingResponse(t *testing.T) {
	emb := NewTestEmbeddingResponse()
	require.NotNil(t, emb)
	assert.Len(t, emb, 8)
}

func TestNewTestCogneeAddRequest(t *testing.T) {
	req := NewTestCogneeAddRequest()
	require.NotNil(t, req)
	assert.Equal(t, "test-dataset", req["dataset_name"])
	assert.Equal(t, "Test content for Cognee processing", req["content"])
}

func TestNewTestCogneeSearchRequest(t *testing.T) {
	req := NewTestCogneeSearchRequest()
	require.NotNil(t, req)
	assert.Equal(t, "test-dataset", req["dataset_name"])
	assert.Equal(t, "test query", req["query"])
	assert.Equal(t, 10, req["limit"])
}

func TestNewTestMCPToolCall(t *testing.T) {
	req := NewTestMCPToolCall()
	require.NotNil(t, req)
	assert.Equal(t, "test_tool", req["name"])

	args, ok := req["arguments"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "value1", args["param1"])
	assert.Equal(t, 42, args["param2"])
}

func TestNewTestMCPCapabilities(t *testing.T) {
	caps := NewTestMCPCapabilities()
	require.NotNil(t, caps)
	assert.Equal(t, "1.0", caps["version"])

	providers, ok := caps["providers"].([]string)
	require.True(t, ok)
	assert.Contains(t, providers, "claude")
}

func TestNewTestUserSession(t *testing.T) {
	session := NewTestUserSession()
	require.NotNil(t, session)
	assert.Equal(t, "test-session-1", session.ID)
	assert.Equal(t, "test-user-1", session.UserID)
	assert.Equal(t, "active", session.Status)
	assert.True(t, session.ExpiresAt.After(time.Now()))
}

func TestNewTestHealthResponse(t *testing.T) {
	resp := NewTestHealthResponse()
	require.NotNil(t, resp)
	assert.Equal(t, "healthy", resp["status"])

	providers, ok := resp["providers"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, 5, providers["total"])
	assert.Equal(t, 4, providers["healthy"])
	assert.Equal(t, 1, providers["unhealthy"])
}

func TestNewTestProviderHealth(t *testing.T) {
	health := NewTestProviderHealth()
	require.NotNil(t, len(health) > 0)
	assert.Len(t, health, 5)
	assert.Nil(t, health["claude"])
}

func TestNewTestEnsembleConfig(t *testing.T) {
	config := NewTestEnsembleConfig()
	require.NotNil(t, config)
	assert.Equal(t, "confidence_weighted", config.Strategy)
	assert.Equal(t, 2, config.MinProviders)
	assert.Equal(t, 0.8, config.ConfidenceThreshold)
	assert.True(t, config.FallbackToBest)
}

func TestNewTestMessages(t *testing.T) {
	messages := NewTestMessages("Hello world")
	require.Len(t, messages, 2)
	assert.Equal(t, "system", messages[0].Role)
	assert.Equal(t, "You are a helpful assistant.", messages[0].Content)
	assert.Equal(t, "user", messages[1].Role)
	assert.Equal(t, "Hello world", messages[1].Content)
}

func TestNewTestLLMProvider(t *testing.T) {
	provider := NewTestLLMProvider("test-provider")
	require.NotNil(t, provider)
	assert.Equal(t, "provider-test-provider", provider.ID)
	assert.Equal(t, "test-provider", provider.Name)
	assert.Equal(t, "api", provider.Type)
	assert.True(t, provider.Enabled)
	assert.Equal(t, "healthy", provider.HealthStatus)
}

func BenchmarkNewTestLLMRequest(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewTestLLMRequest("test prompt")
	}
}

func BenchmarkNewTestMessages(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewTestMessages("test message")
	}
}

func TestTestProviderCapabilitiesNil(t *testing.T) {
	var caps *models.ProviderCapabilities
	assert.Nil(t, caps)
}

func TestMultipleLLMRequests(t *testing.T) {
	reqs := make([]*models.LLMRequest, 10)
	for i := 0; i < 10; i++ {
		reqs[i] = NewTestLLMRequest("prompt")
	}

	assert.Len(t, reqs, 10)
	for _, req := range reqs {
		assert.NotEmpty(t, req.ID)
		assert.Contains(t, req.ID, "test-request-")
	}
}

func TestTestMessagesContent(t *testing.T) {
	testCases := []string{
		"simple",
		"with special chars !@#$%",
		"with\nnewlines",
		"with\ttabs",
		"unicode 你好世界",
	}

	for _, tc := range testCases {
		messages := NewTestMessages(tc)
		assert.Equal(t, tc, messages[1].Content)
	}
}
