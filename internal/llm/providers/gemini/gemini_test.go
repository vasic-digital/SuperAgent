package gemini

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"dev.helix.agent/internal/models"
)

func TestNewGeminiProvider(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		baseURL string
		model   string
		want    *GeminiProvider
	}{
		{
			name:    "all parameters provided",
			apiKey:  "test-api-key-123",
			baseURL: "https://custom.example.com/v1beta/models/%s:generateContent",
			model:   "gemini-ultra",
			want: &GeminiProvider{
				apiKey:  "test-api-key-123",
				baseURL: "https://custom.example.com/v1beta/models/%s:generateContent",
				model:   "gemini-ultra",
				httpClient: &http.Client{
					Timeout: 60 * time.Second,
				},
			},
		},
		{
			name:    "default parameters",
			apiKey:  "test-key",
			baseURL: "",
			model:   "",
			want: &GeminiProvider{
				apiKey:  "test-key",
				baseURL: "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent",
				model:   "gemini-2.0-flash",
				httpClient: &http.Client{
					Timeout: 120 * time.Second,
				},
			},
		},
		{
			name:    "empty api key",
			apiKey:  "",
			baseURL: "https://api.example.com/v1beta/models/%s:generateContent",
			model:   "gemini-pro",
			want: &GeminiProvider{
				apiKey:  "",
				baseURL: "https://api.example.com/v1beta/models/%s:generateContent",
				model:   "gemini-pro",
				httpClient: &http.Client{
					Timeout: 60 * time.Second,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewGeminiProvider(tt.apiKey, tt.baseURL, tt.model)
			assert.Equal(t, tt.want.apiKey, got.apiKey)
			assert.Equal(t, tt.want.baseURL, got.baseURL)
			assert.Equal(t, tt.want.model, got.model)
			assert.NotNil(t, got.httpClient)
			assert.Equal(t, 120*time.Second, got.httpClient.Timeout)
		})
	}
}

func TestGeminiProvider_Complete_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1beta/models/gemini-pro:generateContent", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "test-api-key", r.Header.Get("x-goog-api-key"))
		assert.Equal(t, "HelixAgent/1.0", r.Header.Get("User-Agent"))

		var reqBody GeminiRequest
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		err = json.Unmarshal(body, &reqBody)
		require.NoError(t, err)

		assert.Len(t, reqBody.Contents, 1)
		assert.Equal(t, "user", reqBody.Contents[0].Role)
		assert.Len(t, reqBody.Contents[0].Parts, 1)
		assert.Equal(t, "Hello, how are you?", reqBody.Contents[0].Parts[0].Text)
		assert.Equal(t, 0.7, reqBody.GenerationConfig.Temperature)
		assert.Equal(t, 1000, reqBody.GenerationConfig.MaxOutputTokens)
		assert.Len(t, reqBody.SafetySettings, 4)

		response := GeminiResponse{
			Candidates: []GeminiCandidate{
				{
					Content: GeminiContent{
						Parts: []GeminiPart{
							{Text: "I'm doing well, thank you for asking!"},
						},
						Role: "model",
					},
					FinishReason: "STOP",
					Index:        0,
				},
			},
			UsageMetadata: &GeminiUsageMetadata{
				PromptTokenCount:     10,
				CandidatesTokenCount: 20,
				TotalTokenCount:      30,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	provider := NewGeminiProvider("test-api-key", server.URL+"/v1beta/models/%s:generateContent", "gemini-pro")
	provider.httpClient = server.Client()

	req := &models.LLMRequest{
		ID: "test-req-123",
		ModelParams: models.ModelParameters{
			Model:       "gemini-pro",
			MaxTokens:   1000,
			Temperature: 0.7,
		},
		Prompt: "Hello, how are you?",
	}

	resp, err := provider.Complete(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, "gemini-test-req-123", resp.ID)
	assert.Equal(t, "test-req-123", resp.RequestID)
	assert.Equal(t, "gemini", resp.ProviderID)
	assert.Equal(t, "Gemini", resp.ProviderName)
	assert.Equal(t, "I'm doing well, thank you for asking!", resp.Content)
	assert.Greater(t, resp.Confidence, 0.8) // Should be high confidence for successful response
	assert.Equal(t, 30, resp.TokensUsed)
	assert.Equal(t, "STOP", resp.FinishReason)
	assert.Equal(t, "gemini-pro", resp.Metadata["model"])
}

func TestGeminiProvider_Complete_WithMessages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody GeminiRequest
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		err = json.Unmarshal(body, &reqBody)
		require.NoError(t, err)

		assert.Len(t, reqBody.Contents, 3)
		assert.Equal(t, "user", reqBody.Contents[0].Role)
		assert.Equal(t, "You are a helpful assistant", reqBody.Contents[0].Parts[0].Text)
		assert.Equal(t, "user", reqBody.Contents[1].Role)
		assert.Equal(t, "Hello", reqBody.Contents[1].Parts[0].Text)
		assert.Equal(t, "model", reqBody.Contents[2].Role)
		assert.Equal(t, "Hi there!", reqBody.Contents[2].Parts[0].Text)

		response := GeminiResponse{
			Candidates: []GeminiCandidate{
				{
					Content: GeminiContent{
						Parts: []GeminiPart{
							{Text: "How can I help you today?"},
						},
						Role: "model",
					},
					FinishReason: "STOP",
					Index:        0,
				},
			},
			UsageMetadata: &GeminiUsageMetadata{
				PromptTokenCount:     15,
				CandidatesTokenCount: 25,
				TotalTokenCount:      40,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	provider := NewGeminiProvider("test-api-key", server.URL+"/v1beta/models/%s:generateContent", "gemini-pro")
	provider.httpClient = server.Client()

	req := &models.LLMRequest{
		ID: "test-req-messages",
		ModelParams: models.ModelParameters{
			Model: "gemini-pro",
		},
		Prompt: "You are a helpful assistant",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
			{Role: "assistant", Content: "Hi there!"},
		},
	}

	resp, err := provider.Complete(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "How can I help you today?", resp.Content)
	assert.Equal(t, 40, resp.TokensUsed)
}

func TestGeminiProvider_Complete_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": {"code": 401, "message": "API key not valid"}}`))
	}))
	defer server.Close()

	provider := NewGeminiProvider("invalid-api-key", server.URL+"/v1beta/models/%s:generateContent", "gemini-pro")
	provider.httpClient = server.Client()

	req := &models.LLMRequest{
		ID: "test-req-456",
		ModelParams: models.ModelParameters{
			Model: "gemini-pro",
		},
		Prompt: "Test prompt",
	}

	resp, err := provider.Complete(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Gemini API error: 401")
}

func TestGeminiProvider_Complete_NoCandidates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := GeminiResponse{
			Candidates: []GeminiCandidate{},
			UsageMetadata: &GeminiUsageMetadata{
				PromptTokenCount:     5,
				CandidatesTokenCount: 0,
				TotalTokenCount:      5,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	provider := NewGeminiProvider("test-api-key", server.URL+"/v1beta/models/%s:generateContent", "gemini-pro")
	provider.httpClient = server.Client()

	req := &models.LLMRequest{
		ID: "test-req-789",
		ModelParams: models.ModelParameters{
			Model: "gemini-pro",
		},
		Prompt: "Test prompt",
	}

	resp, err := provider.Complete(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "", resp.Content) // Empty content when no candidates
	assert.Equal(t, 5, resp.TokensUsed)
}

func TestGeminiProvider_Complete_NetworkError(t *testing.T) {
	provider := NewGeminiProvider("test-api-key", "https://invalid-url-that-does-not-exist.example.com/v1beta/models/%s:generateContent", "gemini-pro")
	// Create a client that will fail quickly
	provider.httpClient = &http.Client{
		Timeout: 1 * time.Millisecond,
	}

	req := &models.LLMRequest{
		ID: "test-req-network",
		ModelParams: models.ModelParameters{
			Model: "gemini-pro",
		},
		Prompt: "Test prompt",
	}

	resp, err := provider.Complete(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Gemini API call failed")
}

func TestGeminiProvider_Complete_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	provider := NewGeminiProvider("test-api-key", server.URL+"/v1beta/models/%s:generateContent", "gemini-pro")
	provider.httpClient = server.Client()

	req := &models.LLMRequest{
		ID: "test-req-json",
		ModelParams: models.ModelParameters{
			Model: "gemini-pro",
		},
		Prompt: "Test prompt",
	}

	resp, err := provider.Complete(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to parse Gemini response")
}

func TestGeminiProvider_CompleteStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Send streaming response
		streamData := []string{
			`data: {"candidates":[{"content":{"parts":[{"text":"Hello "}],"role":"model"},"finishReason":"","index":0}]}`,
			`data: {"candidates":[{"content":{"parts":[{"text":"world"}],"role":"model"},"finishReason":"","index":0}]}`,
			`data: {"candidates":[{"content":{"parts":[{"text":"!"}],"role":"model"},"finishReason":"STOP","index":0}]}`,
			`data: [DONE]`,
		}

		for _, data := range streamData {
			w.Write([]byte(data + "\n"))
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}))
	defer server.Close()

	provider := NewGeminiProvider("test-api-key", server.URL+"/v1beta/models/%s:generateContent", "gemini-pro")
	provider.httpClient = server.Client()

	req := &models.LLMRequest{
		ID: "test-req-stream",
		ModelParams: models.ModelParameters{
			Model: "gemini-pro",
		},
		Prompt: "Say hello",
	}

	ch, err := provider.CompleteStream(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, ch)

	// Collect responses
	var responses []*models.LLMResponse
	for resp := range ch {
		responses = append(responses, resp)
	}

	// Should have at least 3 chunk responses + final response
	assert.GreaterOrEqual(t, len(responses), 3)
	assert.Equal(t, "gemini", responses[0].ProviderID)
	assert.Equal(t, "Gemini", responses[0].ProviderName)
}

func TestGeminiProvider_CompleteStream_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "Invalid API key"}`))
	}))
	defer server.Close()

	provider := NewGeminiProvider("invalid-api-key", server.URL+"/v1beta/models/%s:streamGenerateContent", "gemini-2.0-flash")
	provider.httpClient = server.Client()

	req := &models.LLMRequest{
		ID: "test-req-stream-error",
		ModelParams: models.ModelParameters{
			Model: "gemini-2.0-flash",
		},
		Prompt: "Test prompt",
	}

	// CompleteStream now returns an error for non-2xx HTTP status codes
	ch, err := provider.CompleteStream(context.Background(), req)
	require.Error(t, err) // Should return error for 401 status
	require.Nil(t, ch)
	assert.Contains(t, err.Error(), "HTTP 401")
}

func TestGeminiProvider_HealthCheck_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"models": []map[string]string{
				{"name": "gemini-pro"},
				{"name": "gemini-ultra"},
			},
		})
	}))
	defer server.Close()

	provider := NewGeminiProvider("test-api-key", server.URL+"/v1beta/models/%s:generateContent", "gemini-pro")
	provider.healthURL = server.URL + "/v1beta/models"
	provider.httpClient = server.Client()

	err := provider.HealthCheck()
	require.NoError(t, err)
}

func TestGeminiProvider_HealthCheck_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": {"code": 401, "message": "Invalid API key"}}`))
	}))
	defer server.Close()

	provider := NewGeminiProvider("invalid-api-key", server.URL+"/v1beta/models/%s:generateContent", "gemini-pro")
	provider.healthURL = server.URL + "/v1beta/models"
	provider.httpClient = server.Client()

	err := provider.HealthCheck()
	assert.Error(t, err)
}

func TestGeminiProvider_CalculateConfidence(t *testing.T) {
	provider := NewGeminiProvider("test-key", "", "")

	tests := []struct {
		name         string
		content      string
		finishReason string
		wantMin      float64
		wantMax      float64
	}{
		{
			name:         "STOP finish reason",
			content:      "This is a long response that should increase confidence",
			finishReason: "STOP",
			wantMin:      0.95,
			wantMax:      1.0,
		},
		{
			name:         "MAX_TOKENS finish reason",
			content:      "Short",
			finishReason: "MAX_TOKENS",
			wantMin:      0.75,
			wantMax:      0.85,
		},
		{
			name:         "SAFETY finish reason",
			content:      "Content",
			finishReason: "SAFETY",
			wantMin:      0.55,
			wantMax:      0.65,
		},
		{
			name:         "RECITATION finish reason",
			content:      "Content",
			finishReason: "RECITATION",
			wantMin:      0.64, // Adjusted for floating point precision
			wantMax:      0.75,
		},
		{
			name: "long content",
			content: "This is a very long response that exceeds 500 characters. " +
				"This is a very long response that exceeds 500 characters. " +
				"This is a very long response that exceeds 500 characters. " +
				"This is a very long response that exceeds 500 characters. " +
				"This is a very long response that exceeds 500 characters. " +
				"This is a very long response that exceeds 500 characters.",
			finishReason: "STOP",
			wantMin:      0.98,
			wantMax:      1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confidence := provider.calculateConfidence(tt.content, tt.finishReason)
			assert.GreaterOrEqual(t, confidence, tt.wantMin)
			assert.LessOrEqual(t, confidence, tt.wantMax)
		})
	}
}

func TestGeminiProvider_Complete_ContextTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(100 * time.Millisecond)
		response := GeminiResponse{
			Candidates: []GeminiCandidate{
				{
					Content: GeminiContent{
						Parts: []GeminiPart{
							{Text: "Slow response"},
						},
						Role: "model",
					},
					FinishReason: "STOP",
					Index:        0,
				},
			},
			UsageMetadata: &GeminiUsageMetadata{
				PromptTokenCount:     5,
				CandidatesTokenCount: 10,
				TotalTokenCount:      15,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	provider := NewGeminiProvider("test-api-key", server.URL+"/v1beta/models/%s:generateContent", "gemini-pro")
	provider.httpClient = server.Client()

	req := &models.LLMRequest{
		ID: "test-req-timeout",
		ModelParams: models.ModelParameters{
			Model: "gemini-pro",
		},
		Prompt: "Test prompt",
	}

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	resp, err := provider.Complete(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestGeminiProvider_Complete_WithStopSequences(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody GeminiRequest
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		err = json.Unmarshal(body, &reqBody)
		require.NoError(t, err)

		assert.Equal(t, []string{"STOP", "END"}, reqBody.GenerationConfig.StopSequences)

		response := GeminiResponse{
			Candidates: []GeminiCandidate{
				{
					Content: GeminiContent{
						Parts: []GeminiPart{
							{Text: "Response with stop sequences"},
						},
						Role: "model",
					},
					FinishReason: "STOP",
					Index:        0,
				},
			},
			UsageMetadata: &GeminiUsageMetadata{
				PromptTokenCount:     8,
				CandidatesTokenCount: 12,
				TotalTokenCount:      20,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	provider := NewGeminiProvider("test-api-key", server.URL+"/v1beta/models/%s:generateContent", "gemini-pro")
	provider.httpClient = server.Client()

	req := &models.LLMRequest{
		ID: "test-req-stop",
		ModelParams: models.ModelParameters{
			Model:         "gemini-pro",
			StopSequences: []string{"STOP", "END"},
			Temperature:   0.8,
			TopP:          0.9,
			MaxTokens:     500,
		},
		Prompt: "Generate text but stop when you see STOP or END",
	}

	resp, err := provider.Complete(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "Response with stop sequences", resp.Content)
	assert.Equal(t, 20, resp.TokensUsed)
}
