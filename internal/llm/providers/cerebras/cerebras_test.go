package cerebras

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/models"
)

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 1*time.Second, config.InitialDelay)
	assert.Equal(t, 30*time.Second, config.MaxDelay)
	assert.Equal(t, 2.0, config.Multiplier)
}

func TestNewCerebrasProvider(t *testing.T) {
	provider := NewCerebrasProvider("test-key", "", "")

	assert.NotNil(t, provider)
	assert.Equal(t, "test-key", provider.apiKey)
	assert.Equal(t, CerebrasAPIURL, provider.baseURL)
	assert.Equal(t, CerebrasModel, provider.model)
}

func TestNewCerebrasProvider_CustomValues(t *testing.T) {
	provider := NewCerebrasProvider("api-key", "https://custom.api.com", "custom-model")

	assert.Equal(t, "api-key", provider.apiKey)
	assert.Equal(t, "https://custom.api.com", provider.baseURL)
	assert.Equal(t, "custom-model", provider.model)
}

func TestNewCerebrasProviderWithRetry(t *testing.T) {
	retryConfig := RetryConfig{
		MaxRetries:   5,
		InitialDelay: 500 * time.Millisecond,
		MaxDelay:     60 * time.Second,
		Multiplier:   3.0,
	}

	provider := NewCerebrasProviderWithRetry("key", "", "", retryConfig)

	assert.Equal(t, 5, provider.retryConfig.MaxRetries)
	assert.Equal(t, 500*time.Millisecond, provider.retryConfig.InitialDelay)
	assert.Equal(t, 60*time.Second, provider.retryConfig.MaxDelay)
	assert.Equal(t, 3.0, provider.retryConfig.Multiplier)
}

func TestCerebrasProvider_ConvertRequest(t *testing.T) {
	provider := NewCerebrasProvider("key", "", "")

	req := &models.LLMRequest{
		ID:     "req-1",
		Prompt: "System prompt",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
			{Role: "assistant", Content: "Hi"},
		},
		ModelParams: models.ModelParameters{
			Temperature: 0.7,
			MaxTokens:   1000,
			TopP:        0.9,
		},
	}

	cerebrasReq := provider.convertRequest(req)

	assert.Equal(t, CerebrasModel, cerebrasReq.Model)
	assert.Len(t, cerebrasReq.Messages, 3) // System + 2 messages
	assert.Equal(t, "system", cerebrasReq.Messages[0].Role)
	assert.Equal(t, "System prompt", cerebrasReq.Messages[0].Content)
	assert.Equal(t, 0.7, cerebrasReq.Temperature)
	assert.Equal(t, 1000, cerebrasReq.MaxTokens)
	assert.Equal(t, 0.9, cerebrasReq.TopP)
	assert.False(t, cerebrasReq.Stream)
}

func TestCerebrasProvider_ConvertRequest_MaxTokensLimit(t *testing.T) {
	provider := NewCerebrasProvider("key", "", "")

	// Test exceeding max limit
	req := &models.LLMRequest{
		Messages: []models.Message{{Role: "user", Content: "test"}},
		ModelParams: models.ModelParameters{
			MaxTokens: 100000, // Exceeds 8192 limit
		},
	}

	cerebrasReq := provider.convertRequest(req)
	assert.Equal(t, 8192, cerebrasReq.MaxTokens)

	// Test default when 0
	req.ModelParams.MaxTokens = 0
	cerebrasReq = provider.convertRequest(req)
	assert.Equal(t, 4096, cerebrasReq.MaxTokens)
}

func TestCerebrasProvider_CalculateConfidence(t *testing.T) {
	provider := NewCerebrasProvider("key", "", "")

	tests := []struct {
		content      string
		finishReason string
		minConfidence float64
		maxConfidence float64
	}{
		{"short", "stop", 0.85, 0.95},
		{"short", "length", 0.65, 0.75},
		{string(make([]byte, 150)), "stop", 0.90, 1.0},
		{string(make([]byte, 600)), "stop", 0.95, 1.0},
		{"", "stop", 0.85, 0.95},
	}

	for _, tc := range tests {
		confidence := provider.calculateConfidence(tc.content, tc.finishReason)
		assert.GreaterOrEqual(t, confidence, tc.minConfidence, "Content len=%d, finishReason=%s", len(tc.content), tc.finishReason)
		assert.LessOrEqual(t, confidence, tc.maxConfidence, "Content len=%d, finishReason=%s", len(tc.content), tc.finishReason)
	}
}

func TestCerebrasProvider_GetCapabilities(t *testing.T) {
	provider := NewCerebrasProvider("key", "", "")

	caps := provider.GetCapabilities()

	assert.NotNil(t, caps)
	assert.Contains(t, caps.SupportedModels, "llama-3.3-70b")
	assert.Contains(t, caps.SupportedFeatures, "streaming")
	assert.True(t, caps.SupportsStreaming)
	assert.False(t, caps.SupportsFunctionCalling)
	assert.False(t, caps.SupportsVision)
	assert.Equal(t, 8192, caps.Limits.MaxTokens)
}

func TestCerebrasProvider_ValidateConfig(t *testing.T) {
	// Note: NewCerebrasProvider sets default values for empty baseURL and model,
	// so only the apiKey check can fail via the constructor.
	tests := []struct {
		apiKey  string
		baseURL string
		model   string
		valid   bool
		errLen  int
	}{
		{"key", "url", "model", true, 0},
		{"", "url", "model", false, 1},      // Only apiKey error (baseURL and model use defaults if empty)
		{"key", "", "model", true, 0},       // Empty baseURL gets default
		{"key", "url", "", true, 0},         // Empty model gets default
		{"", "", "", false, 1},              // Only apiKey error (others get defaults)
	}

	for _, tc := range tests {
		provider := NewCerebrasProvider(tc.apiKey, tc.baseURL, tc.model)
		valid, errors := provider.ValidateConfig(nil)

		assert.Equal(t, tc.valid, valid, "apiKey=%s, baseURL=%s, model=%s", tc.apiKey, tc.baseURL, tc.model)
		assert.Len(t, errors, tc.errLen, "apiKey=%s, baseURL=%s, model=%s", tc.apiKey, tc.baseURL, tc.model)
	}
}

func TestCerebrasProvider_NextDelay(t *testing.T) {
	provider := NewCerebrasProviderWithRetry("key", "", "", RetryConfig{
		MaxRetries:   3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
	})

	// Test exponential increase
	delay1 := provider.nextDelay(1 * time.Second)
	assert.Equal(t, 2*time.Second, delay1)

	delay2 := provider.nextDelay(2 * time.Second)
	assert.Equal(t, 4*time.Second, delay2)

	// Test max cap
	delay3 := provider.nextDelay(8 * time.Second)
	assert.Equal(t, 10*time.Second, delay3) // Capped at max
}

func TestIsRetryableStatus(t *testing.T) {
	tests := []struct {
		statusCode int
		retryable  bool
	}{
		{http.StatusOK, false},
		{http.StatusBadRequest, false},
		{http.StatusUnauthorized, false},
		{http.StatusForbidden, false},
		{http.StatusTooManyRequests, true},
		{http.StatusInternalServerError, true},
		{http.StatusBadGateway, true},
		{http.StatusServiceUnavailable, true},
		{http.StatusGatewayTimeout, true},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.retryable, isRetryableStatus(tc.statusCode), "Status %d", tc.statusCode)
	}
}

func TestIsAuthRetryableStatus(t *testing.T) {
	assert.True(t, isAuthRetryableStatus(http.StatusUnauthorized))
	assert.False(t, isAuthRetryableStatus(http.StatusForbidden))
	assert.False(t, isAuthRetryableStatus(http.StatusOK))
}

func TestMin(t *testing.T) {
	assert.Equal(t, 1, min(1, 2))
	assert.Equal(t, 1, min(2, 1))
	assert.Equal(t, 0, min(0, 5))
	assert.Equal(t, -1, min(-1, 0))
}

func TestCerebrasRequest_Fields(t *testing.T) {
	req := CerebrasRequest{
		Model:       "llama-3.3-70b",
		Messages:    []CerebrasMessage{{Role: "user", Content: "test"}},
		Temperature: 0.7,
		MaxTokens:   1000,
		TopP:        0.9,
		Stream:      true,
	}

	assert.Equal(t, "llama-3.3-70b", req.Model)
	assert.Len(t, req.Messages, 1)
	assert.Equal(t, 0.7, req.Temperature)
	assert.Equal(t, 1000, req.MaxTokens)
	assert.True(t, req.Stream)
}

func TestCerebrasMessage_Fields(t *testing.T) {
	msg := CerebrasMessage{
		Role:    "assistant",
		Content: "Hello, how can I help?",
	}

	assert.Equal(t, "assistant", msg.Role)
	assert.Equal(t, "Hello, how can I help?", msg.Content)
}

func TestCerebrasResponse_Fields(t *testing.T) {
	resp := CerebrasResponse{
		ID:      "resp-123",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   "llama-3.3-70b",
		Choices: []CerebrasChoice{
			{
				Index:        0,
				Message:      CerebrasMessage{Role: "assistant", Content: "Response"},
				FinishReason: "stop",
			},
		},
		Usage: CerebrasUsage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
	}

	assert.Equal(t, "resp-123", resp.ID)
	assert.Len(t, resp.Choices, 1)
	assert.Equal(t, 30, resp.Usage.TotalTokens)
}

func TestCerebrasUsage_Fields(t *testing.T) {
	usage := CerebrasUsage{
		PromptTokens:     100,
		CompletionTokens: 200,
		TotalTokens:      300,
	}

	assert.Equal(t, 100, usage.PromptTokens)
	assert.Equal(t, 200, usage.CompletionTokens)
	assert.Equal(t, 300, usage.TotalTokens)
}

func TestCerebrasProvider_Complete_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Contains(t, r.Header.Get("Authorization"), "Bearer")

		response := `{
			"id": "test-id",
			"object": "chat.completion",
			"created": 1234567890,
			"model": "llama-3.3-70b",
			"choices": [{
				"index": 0,
				"message": {"role": "assistant", "content": "Test response"},
				"finish_reason": "stop"
			}],
			"usage": {"prompt_tokens": 10, "completion_tokens": 20, "total_tokens": 30}
		}`

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	provider := NewCerebrasProvider("test-key", server.URL, "")

	req := &models.LLMRequest{
		ID:       "req-1",
		Messages: []models.Message{{Role: "user", Content: "Hello"}},
	}

	resp, err := provider.Complete(context.Background(), req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test-id", resp.ID)
	assert.Equal(t, "Test response", resp.Content)
	assert.Equal(t, "cerebras", resp.ProviderID)
	assert.Equal(t, 30, resp.TokensUsed)
}

func TestCerebrasProvider_Complete_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{"error": {"message": "Invalid API key", "type": "auth_error", "code": "401"}}`
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(response))
	}))
	defer server.Close()

	provider := NewCerebrasProvider("invalid-key", server.URL, "")

	req := &models.LLMRequest{
		ID:       "req-1",
		Messages: []models.Message{{Role: "user", Content: "Hello"}},
	}

	resp, err := provider.Complete(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Invalid API key")
}

func TestCerebrasProvider_Complete_EmptyChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{"id": "test", "choices": [], "usage": {}}`
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	provider := NewCerebrasProvider("key", server.URL, "")

	req := &models.LLMRequest{
		ID:       "req-1",
		Messages: []models.Message{{Role: "user", Content: "Hello"}},
	}

	resp, err := provider.Complete(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "no choices")
}

func TestCerebrasProvider_ConvertResponse(t *testing.T) {
	provider := NewCerebrasProvider("key", "", "")

	req := &models.LLMRequest{ID: "req-1"}
	cerebrasResp := &CerebrasResponse{
		ID:    "resp-1",
		Model: "llama-3.3-70b",
		Choices: []CerebrasChoice{
			{
				Index:        0,
				Message:      CerebrasMessage{Role: "assistant", Content: "Test content"},
				FinishReason: "stop",
			},
		},
		Usage: CerebrasUsage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
	}

	startTime := time.Now().Add(-100 * time.Millisecond)
	resp := provider.convertResponse(req, cerebrasResp, startTime)

	assert.Equal(t, "resp-1", resp.ID)
	assert.Equal(t, "req-1", resp.RequestID)
	assert.Equal(t, "cerebras", resp.ProviderID)
	assert.Equal(t, "Cerebras", resp.ProviderName)
	assert.Equal(t, "Test content", resp.Content)
	assert.Equal(t, 30, resp.TokensUsed)
	assert.Equal(t, "stop", resp.FinishReason)
	assert.GreaterOrEqual(t, resp.ResponseTime, int64(100))
}

func TestCerebrasStreamResponse_Fields(t *testing.T) {
	finishReason := "stop"
	resp := CerebrasStreamResponse{
		ID:      "stream-1",
		Object:  "chat.completion.chunk",
		Created: 1234567890,
		Model:   "llama-3.3-70b",
		Choices: []CerebrasStreamChoice{
			{
				Index:        0,
				Delta:        CerebrasMessage{Role: "assistant", Content: "chunk"},
				FinishReason: &finishReason,
			},
		},
	}

	assert.Equal(t, "stream-1", resp.ID)
	assert.Len(t, resp.Choices, 1)
	assert.Equal(t, "chunk", resp.Choices[0].Delta.Content)
	assert.NotNil(t, resp.Choices[0].FinishReason)
}
