package deepseek

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/superagent/superagent/internal/models"
)

func TestNewDeepSeekProvider(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		baseURL  string
		model    string
		expected *DeepSeekProvider
	}{
		{
			name:    "default values",
			apiKey:  "test-key",
			baseURL: "",
			model:   "",
			expected: &DeepSeekProvider{
				apiKey:  "test-key",
				baseURL: DeepSeekAPIURL,
				model:   DeepSeekModel,
			},
		},
		{
			name:    "custom values",
			apiKey:  "test-key",
			baseURL: "https://custom.example.com",
			model:   "deepseek-chat",
			expected: &DeepSeekProvider{
				apiKey:  "test-key",
				baseURL: "https://custom.example.com",
				model:   "deepseek-chat",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewDeepSeekProvider(tt.apiKey, tt.baseURL, tt.model)
			require.NotNil(t, provider)
			assert.Equal(t, tt.expected.apiKey, provider.apiKey)
			assert.Equal(t, tt.expected.baseURL, provider.baseURL)
			assert.Equal(t, tt.expected.model, provider.model)
			assert.NotNil(t, provider.httpClient)
			assert.Equal(t, 60*time.Second, provider.httpClient.Timeout)
		})
	}
}

func TestDeepSeekProvider_HealthCheck(t *testing.T) {
	t.Skip("HealthCheck method may make real API calls, skipping for unit tests")
}

func TestDeepSeekProvider_HealthCheck_Error(t *testing.T) {
	t.Skip("HealthCheck method may make real API calls, skipping for unit tests")
}

func TestDeepSeekProvider_Complete_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/chat/completions", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "chat_123",
			"object": "chat.completion",
			"created": 1677858242,
			"model": "deepseek-coder",
			"choices": [{
				"index": 0,
				"message": {"role": "assistant", "content": "Hello from DeepSeek!"},
				"finish_reason": "stop"
			}],
			"usage": {"prompt_tokens": 10, "completion_tokens": 5, "total_tokens": 15}
		}`))
	}))
	defer server.Close()

	provider := NewDeepSeekProvider("test-key", server.URL+"/v1/chat/completions", "deepseek-coder")
	req := &models.LLMRequest{
		ID: "test-request",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	resp, err := provider.Complete(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "chat_123", resp.ID)
	assert.Equal(t, "Hello from DeepSeek!", resp.Content)
	assert.Equal(t, "stop", resp.FinishReason)
	assert.Greater(t, resp.Confidence, 0.0)
	assert.Less(t, resp.Confidence, 1.0)
}

func TestDeepSeekProvider_Complete_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": {"message": "Invalid request", "type": "invalid_request_error"}}`))
	}))
	defer server.Close()

	provider := NewDeepSeekProvider("test-key", server.URL+"/v1/chat/completions", "deepseek-coder")
	req := &models.LLMRequest{
		ID: "test-request",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	resp, err := provider.Complete(context.Background(), req)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DeepSeek API error: 400")
}

func TestDeepSeekProvider_ConvertRequest(t *testing.T) {
	provider := NewDeepSeekProvider("test-key", "", "")
	req := &models.LLMRequest{
		ID: "test-request",
		Messages: []models.Message{
			{Role: "user", Content: "Hello, DeepSeek!"},
			{Role: "assistant", Content: "Hi there!"},
			{Role: "user", Content: "How are you?"},
		},
		ModelParams: models.ModelParameters{
			Model:         "deepseek-coder",
			MaxTokens:     100,
			Temperature:   0.7,
			TopP:          0.9,
			StopSequences: []string{"\n", "STOP"},
		},
	}

	deepseekReq := provider.convertRequest(req)
	assert.Equal(t, "deepseek-coder", deepseekReq.Model)
	assert.Equal(t, 100, deepseekReq.MaxTokens)
	assert.Equal(t, 0.7, deepseekReq.Temperature)
	assert.Equal(t, 0.9, deepseekReq.TopP)
	assert.Equal(t, []string{"\n", "STOP"}, deepseekReq.Stop)
	assert.Len(t, deepseekReq.Messages, 3)
	assert.Equal(t, "user", deepseekReq.Messages[0].Role)
	assert.Equal(t, "Hello, DeepSeek!", deepseekReq.Messages[0].Content)
	assert.Equal(t, "assistant", deepseekReq.Messages[1].Role)
	assert.Equal(t, "Hi there!", deepseekReq.Messages[1].Content)
	assert.Equal(t, "user", deepseekReq.Messages[2].Role)
	assert.Equal(t, "How are you?", deepseekReq.Messages[2].Content)
}

func TestDeepSeekProvider_CalculateConfidence(t *testing.T) {
	provider := NewDeepSeekProvider("test-key", "", "")

	tests := []struct {
		name         string
		content      string
		finishReason string
		expectedMin  float64
		expectedMax  float64
	}{
		{
			name:         "good response with stop",
			content:      "This is a comprehensive and well-formed response.",
			finishReason: "stop",
			expectedMin:  0.8,
			expectedMax:  1.0,
		},
		{
			name:         "short response with length",
			content:      "Short",
			finishReason: "length",
			expectedMin:  0.7,
			expectedMax:  0.9,
		},
		{
			name:         "content filter",
			content:      "Filtered content",
			finishReason: "content_filter",
			expectedMin:  0.3,
			expectedMax:  0.6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confidence := provider.calculateConfidence(tt.content, tt.finishReason)
			assert.GreaterOrEqual(t, confidence, tt.expectedMin)
			assert.LessOrEqual(t, confidence, tt.expectedMax)
		})
	}
}
