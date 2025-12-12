package claude

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

func TestNewClaudeProvider(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		baseURL  string
		model    string
		expected *ClaudeProvider
	}{
		{
			name:    "default values",
			apiKey:  "test-key",
			baseURL: "",
			model:   "",
			expected: &ClaudeProvider{
				apiKey:  "test-key",
				baseURL: ClaudeAPIURL,
				model:   ClaudeModel,
			},
		},
		{
			name:    "custom values",
			apiKey:  "test-key",
			baseURL: "https://custom.example.com",
			model:   "claude-3-custom",
			expected: &ClaudeProvider{
				apiKey:  "test-key",
				baseURL: "https://custom.example.com",
				model:   "claude-3-custom",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewClaudeProvider(tt.apiKey, tt.baseURL, tt.model)
			require.NotNil(t, provider)
			assert.Equal(t, tt.expected.apiKey, provider.apiKey)
			assert.Equal(t, tt.expected.baseURL, provider.baseURL)
			assert.Equal(t, tt.expected.model, provider.model)
			assert.NotNil(t, provider.httpClient)
			assert.Equal(t, 60*time.Second, provider.httpClient.Timeout)
		})
	}
}

func TestClaudeProvider_HealthCheck(t *testing.T) {
	t.Skip("HealthCheck method uses hardcoded Anthropic URL, cannot test with mock server")
}

func TestClaudeProvider_HealthCheck_Error(t *testing.T) {
	t.Skip("HealthCheck method uses hardcoded Anthropic URL, cannot test with mock server")
}

func TestClaudeProvider_Complete_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/messages", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "test-key", r.Header.Get("x-api-key"))
		assert.Equal(t, "2023-06-01", r.Header.Get("anthropic-version"))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "msg_123",
			"type": "message",
			"role": "assistant",
			"content": [{"type": "text", "text": "Hello, world!"}],
			"model": "claude-3-sonnet-20240229",
			"stop_reason": "end_turn",
			"usage": {"input_tokens": 10, "output_tokens": 5}
		}`))
	}))
	defer server.Close()

	provider := NewClaudeProvider("test-key", server.URL+"/v1/messages", "claude-3-sonnet-20240229")
	req := &models.LLMRequest{
		ID: "test-request",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	resp, err := provider.Complete(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "msg_123", resp.ID)
	assert.Equal(t, "Hello, world!", resp.Content)
	assert.Equal(t, "end_turn", resp.FinishReason)
	assert.Greater(t, resp.Confidence, 0.0)
	assert.Less(t, resp.Confidence, 1.0)
}

func TestClaudeProvider_Complete_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": {"type": "invalid_request_error", "message": "Invalid request"}}`))
	}))
	defer server.Close()

	provider := NewClaudeProvider("test-key", server.URL, "claude-3-test")
	req := &models.LLMRequest{
		ID: "test-request",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	resp, err := provider.Complete(context.Background(), req)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Claude API error: 400")
}

func TestClaudeProvider_ConvertRequest(t *testing.T) {
	provider := NewClaudeProvider("test-key", "", "")
	req := &models.LLMRequest{
		ID: "test-request",
		Messages: []models.Message{
			{Role: "user", Content: "Hello, Claude!"},
			{Role: "assistant", Content: "Hi there!"},
			{Role: "user", Content: "How are you?"},
		},
		ModelParams: models.ModelParameters{
			Model:         "claude-3-sonnet-20240229",
			MaxTokens:     100,
			Temperature:   0.7,
			TopP:          0.9,
			StopSequences: []string{"\n", "STOP"},
		},
	}

	claudeReq := provider.convertRequest(req)
	assert.Equal(t, "claude-3-sonnet-20240229", claudeReq.Model)
	assert.Equal(t, 100, claudeReq.MaxTokens)
	assert.Equal(t, 0.7, claudeReq.Temperature)
	assert.Equal(t, 0.9, claudeReq.TopP)
	assert.Equal(t, []string{"\n", "STOP"}, claudeReq.StopSequences)
	assert.Len(t, claudeReq.Messages, 3)
	assert.Equal(t, "user", claudeReq.Messages[0].Role)
	assert.Equal(t, "Hello, Claude!", claudeReq.Messages[0].Content)
	assert.Equal(t, "assistant", claudeReq.Messages[1].Role)
	assert.Equal(t, "Hi there!", claudeReq.Messages[1].Content)
	assert.Equal(t, "user", claudeReq.Messages[2].Role)
	assert.Equal(t, "How are you?", claudeReq.Messages[2].Content)
}

func TestClaudeProvider_CalculateConfidence(t *testing.T) {
	provider := NewClaudeProvider("test-key", "", "")

	tests := []struct {
		name         string
		content      string
		finishReason string
		expectedMin  float64
		expectedMax  float64
	}{
		{
			name:         "good response with end_turn",
			content:      "This is a comprehensive and well-formed response.",
			finishReason: "end_turn",
			expectedMin:  0.9,
			expectedMax:  1.0,
		},
		{
			name:         "short response with max_tokens",
			content:      "Short",
			finishReason: "max_tokens",
			expectedMin:  0.8,
			expectedMax:  0.9,
		},
		{
			name:         "empty response",
			content:      "",
			finishReason: "stop_sequence",
			expectedMin:  0.9,
			expectedMax:  1.0,
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
