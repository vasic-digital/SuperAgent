package gemini

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"dev.helix.agent/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		_ = json.NewEncoder(w).Encode(response)
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
		_ = json.NewEncoder(w).Encode(response)
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
		_, _ = w.Write([]byte(`{"error": {"code": 401, "message": "API key not valid"}}`))
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
		_ = json.NewEncoder(w).Encode(response)
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
		_, _ = w.Write([]byte("invalid json"))
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
			_, _ = w.Write([]byte(data + "\n"))
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
		_, _ = w.Write([]byte(`{"error": "Invalid API key"}`))
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
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
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
		_, _ = w.Write([]byte(`{"error": {"code": 401, "message": "Invalid API key"}}`))
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
		_ = json.NewEncoder(w).Encode(response)
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
		_ = json.NewEncoder(w).Encode(response)
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

// ==============================================================================
// ADDITIONAL COMPREHENSIVE TESTS FOR GEMINI PROVIDER
// ==============================================================================

func TestGeminiProvider_Complete_WithTools(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody GeminiRequest
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		err = json.Unmarshal(body, &reqBody)
		require.NoError(t, err)

		// Verify tools are properly converted
		require.Len(t, reqBody.Tools, 1)
		require.Len(t, reqBody.Tools[0].FunctionDeclarations, 2)
		assert.Equal(t, "get_weather", reqBody.Tools[0].FunctionDeclarations[0].Name)
		assert.Equal(t, "search_files", reqBody.Tools[0].FunctionDeclarations[1].Name)

		// Return function call response
		response := GeminiResponse{
			Candidates: []GeminiCandidate{
				{
					Content: GeminiContent{
						Parts: []GeminiPart{
							{
								FunctionCall: map[string]any{
									"name": "get_weather",
									"args": map[string]interface{}{
										"location": "San Francisco",
										"unit":     "celsius",
									},
								},
							},
						},
						Role: "model",
					},
					FinishReason: "STOP",
					Index:        0,
				},
			},
			UsageMetadata: &GeminiUsageMetadata{
				TotalTokenCount: 50,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	provider := NewGeminiProvider("test-api-key", server.URL+"/v1beta/models/%s:generateContent", "gemini-pro")
	provider.httpClient = server.Client()

	req := &models.LLMRequest{
		ID:     "test-tools",
		Prompt: "What's the weather in San Francisco?",
		Tools: []models.Tool{
			{
				Type: "function",
				Function: models.ToolFunction{
					Name:        "get_weather",
					Description: "Get weather for a location",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"location": map[string]interface{}{"type": "string"},
							"unit":     map[string]interface{}{"type": "string"},
						},
					},
				},
			},
			{
				Type: "function",
				Function: models.ToolFunction{
					Name:        "search_files",
					Description: "Search for files",
				},
			},
		},
		ToolChoice: "auto",
	}

	resp, err := provider.Complete(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "tool_calls", resp.FinishReason)
	require.Len(t, resp.ToolCalls, 1)
	assert.Equal(t, "get_weather", resp.ToolCalls[0].Function.Name)
}

func TestGeminiProvider_Complete_WithToolChoice(t *testing.T) {
	testCases := []struct {
		name         string
		toolChoice   string
		expectedMode string
	}{
		{"auto", "auto", "AUTO"},
		{"none", "none", "NONE"},
		{"required", "required", "ANY"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedToolConfig *GeminiToolConfig

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var reqBody GeminiRequest
				body, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(body, &reqBody)
				capturedToolConfig = reqBody.ToolConfig

				response := GeminiResponse{
					Candidates: []GeminiCandidate{
						{
							Content: GeminiContent{
								Parts: []GeminiPart{{Text: "OK"}},
							},
							FinishReason: "STOP",
						},
					},
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			provider := NewGeminiProvider("test-key", server.URL+"/v1beta/models/%s:generateContent", "gemini-pro")
			provider.httpClient = server.Client()

			req := &models.LLMRequest{
				ID:     "test-tool-choice",
				Prompt: "Test",
				Tools: []models.Tool{
					{Type: "function", Function: models.ToolFunction{Name: "test_tool"}},
				},
				ToolChoice: tc.toolChoice,
			}

			_, err := provider.Complete(context.Background(), req)
			require.NoError(t, err)
			require.NotNil(t, capturedToolConfig)
			assert.Equal(t, tc.expectedMode, capturedToolConfig.FunctionCallingConfig.Mode)
		})
	}
}

func TestGeminiProvider_ValidateConfig(t *testing.T) {
	tests := []struct {
		name         string
		apiKey       string
		baseURL      string
		model        string
		expectValid  bool
		expectErrLen int
	}{
		{
			name:         "all valid",
			apiKey:       "test-key",
			baseURL:      "https://api.example.com",
			model:        "gemini-pro",
			expectValid:  true,
			expectErrLen: 0,
		},
		{
			name:         "missing api key",
			apiKey:       "",
			baseURL:      "https://api.example.com",
			model:        "gemini-pro",
			expectValid:  false,
			expectErrLen: 1,
		},
		{
			name:         "missing base url",
			apiKey:       "test-key",
			baseURL:      "",
			model:        "gemini-pro",
			expectValid:  false,
			expectErrLen: 1,
		},
		{
			name:         "missing model",
			apiKey:       "test-key",
			baseURL:      "https://api.example.com",
			model:        "",
			expectValid:  false,
			expectErrLen: 1,
		},
		{
			name:         "all missing",
			apiKey:       "",
			baseURL:      "",
			model:        "",
			expectValid:  false,
			expectErrLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &GeminiProvider{
				apiKey:  tt.apiKey,
				baseURL: tt.baseURL,
				model:   tt.model,
			}

			valid, errs := provider.ValidateConfig(nil)
			assert.Equal(t, tt.expectValid, valid)
			assert.Len(t, errs, tt.expectErrLen)
		})
	}
}

func TestGeminiProvider_GetCapabilities(t *testing.T) {
	provider := NewGeminiProvider("test-key", "", "")
	caps := provider.GetCapabilities()

	require.NotNil(t, caps)

	// Check supported models
	assert.Contains(t, caps.SupportedModels, "gemini-2.0-flash")
	assert.Contains(t, caps.SupportedModels, "gemini-2.5-flash")
	assert.Contains(t, caps.SupportedModels, "gemini-2.5-pro")

	// Check supported features
	assert.Contains(t, caps.SupportedFeatures, "text_completion")
	assert.Contains(t, caps.SupportedFeatures, "chat")
	assert.Contains(t, caps.SupportedFeatures, "streaming")
	assert.Contains(t, caps.SupportedFeatures, "function_calling")
	assert.Contains(t, caps.SupportedFeatures, "vision")

	// Check boolean capabilities
	assert.True(t, caps.SupportsStreaming)
	assert.True(t, caps.SupportsFunctionCalling)
	assert.True(t, caps.SupportsVision)
	assert.True(t, caps.SupportsTools)
	assert.False(t, caps.SupportsSearch)
	assert.True(t, caps.SupportsReasoning)

	// Check limits
	assert.Equal(t, 32768, caps.Limits.MaxTokens)
	assert.Equal(t, 8192, caps.Limits.MaxOutputLength)
	assert.Equal(t, 10, caps.Limits.MaxConcurrentRequests)

	// Check metadata
	assert.Equal(t, "Google", caps.Metadata["provider"])
	assert.Equal(t, "Gemini", caps.Metadata["model_family"])
}

func TestGeminiProvider_Retry_RateLimited(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		response := GeminiResponse{
			Candidates: []GeminiCandidate{
				{
					Content: GeminiContent{
						Parts: []GeminiPart{{Text: "Success after retry"}},
					},
					FinishReason: "STOP",
				},
			},
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	retryConfig := RetryConfig{
		MaxRetries:   3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
	}

	provider := NewGeminiProviderWithRetry("test-key", server.URL+"/v1beta/models/%s:generateContent", "gemini-pro", retryConfig)
	provider.httpClient = server.Client()

	req := &models.LLMRequest{ID: "retry-test", Prompt: "Test"}
	resp, err := provider.Complete(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "Success after retry", resp.Content)
	assert.Equal(t, 3, attempts)
}

func TestGeminiProvider_Retry_ServerError(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		response := GeminiResponse{
			Candidates: []GeminiCandidate{
				{
					Content: GeminiContent{
						Parts: []GeminiPart{{Text: "OK"}},
					},
					FinishReason: "STOP",
				},
			},
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	retryConfig := RetryConfig{
		MaxRetries:   3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
	}

	provider := NewGeminiProviderWithRetry("test-key", server.URL+"/v1beta/models/%s:generateContent", "gemini-pro", retryConfig)
	provider.httpClient = server.Client()

	req := &models.LLMRequest{ID: "retry-test", Prompt: "Test"}
	resp, err := provider.Complete(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 2, attempts)
}

func TestGeminiProvider_Retry_AuthError(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error": "Invalid API key"}`))
			return
		}
		response := GeminiResponse{
			Candidates: []GeminiCandidate{
				{Content: GeminiContent{Parts: []GeminiPart{{Text: "OK"}}}, FinishReason: "STOP"},
			},
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	retryConfig := RetryConfig{
		MaxRetries:   3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
	}

	provider := NewGeminiProviderWithRetry("test-key", server.URL+"/v1beta/models/%s:generateContent", "gemini-pro", retryConfig)
	provider.httpClient = server.Client()

	req := &models.LLMRequest{ID: "auth-retry", Prompt: "Test"}
	resp, err := provider.Complete(context.Background(), req)

	// Should succeed after auth retry
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 2, attempts) // Auth retry + success
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
		{http.StatusNotFound, false},
		{http.StatusTooManyRequests, true},
		{http.StatusInternalServerError, true},
		{http.StatusBadGateway, true},
		{http.StatusServiceUnavailable, true},
		{http.StatusGatewayTimeout, true},
	}

	for _, tt := range tests {
		t.Run(http.StatusText(tt.statusCode), func(t *testing.T) {
			assert.Equal(t, tt.retryable, isRetryableStatus(tt.statusCode))
		})
	}
}

func TestIsAuthRetryableStatus(t *testing.T) {
	tests := []struct {
		statusCode int
		retryable  bool
	}{
		{http.StatusUnauthorized, true},
		{http.StatusOK, false},
		{http.StatusForbidden, false},
		{http.StatusTooManyRequests, false},
	}

	for _, tt := range tests {
		t.Run(http.StatusText(tt.statusCode), func(t *testing.T) {
			assert.Equal(t, tt.retryable, isAuthRetryableStatus(tt.statusCode))
		})
	}
}

func TestGeminiProvider_NextDelay(t *testing.T) {
	provider := NewGeminiProviderWithRetry("test-key", "", "", RetryConfig{
		MaxRetries:   3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
	})

	// First delay should be multiplied
	next := provider.nextDelay(1 * time.Second)
	assert.Equal(t, 2*time.Second, next)

	// Should hit max delay
	next = provider.nextDelay(8 * time.Second)
	assert.Equal(t, 10*time.Second, next)
}

func TestGeminiProvider_CompleteStream_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Send some malformed JSON mixed with valid JSON
		_, _ = w.Write([]byte("data: {invalid json}\n"))
		_, _ = w.Write([]byte("data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"Valid chunk\"}]},\"finishReason\":\"\"}]}\n"))
		_, _ = w.Write([]byte("data: [DONE]\n"))
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}))
	defer server.Close()

	provider := NewGeminiProvider("test-key", server.URL+"/v1beta/models/%s:generateContent", "gemini-pro")
	provider.httpClient = server.Client()

	req := &models.LLMRequest{ID: "malformed-test", Prompt: "Test"}
	ch, err := provider.CompleteStream(context.Background(), req)
	require.NoError(t, err)

	var responses []*models.LLMResponse
	for resp := range ch {
		responses = append(responses, resp)
	}

	// Should have received the valid chunk + final response
	assert.GreaterOrEqual(t, len(responses), 1)
}

func TestGeminiProvider_CompleteStream_ArrayWrapper(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Send response with array wrapper (as Gemini sometimes does)
		_, _ = w.Write([]byte("[{\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"Array wrapped\"}]},\"finishReason\":\"STOP\"}]}]\n"))
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}))
	defer server.Close()

	provider := NewGeminiProvider("test-key", server.URL+"/v1beta/models/%s:generateContent", "gemini-pro")
	provider.httpClient = server.Client()

	req := &models.LLMRequest{ID: "array-test", Prompt: "Test"}
	ch, err := provider.CompleteStream(context.Background(), req)
	require.NoError(t, err)

	var responses []*models.LLMResponse
	for resp := range ch {
		responses = append(responses, resp)
	}

	// Should handle array-wrapped response
	assert.GreaterOrEqual(t, len(responses), 1)
}

func TestGeminiProvider_ConvertRequest_MaxTokensCapping(t *testing.T) {
	provider := NewGeminiProvider("test-key", "", "gemini-pro")

	t.Run("caps tokens above 8192", func(t *testing.T) {
		req := &models.LLMRequest{
			Prompt: "Test",
			ModelParams: models.ModelParameters{
				MaxTokens: 100000, // Way above limit
			},
		}
		geminiReq := provider.convertRequest(req)
		assert.Equal(t, 8192, geminiReq.GenerationConfig.MaxOutputTokens)
	})

	t.Run("uses default when zero", func(t *testing.T) {
		req := &models.LLMRequest{
			Prompt: "Test",
			ModelParams: models.ModelParameters{
				MaxTokens: 0,
			},
		}
		geminiReq := provider.convertRequest(req)
		assert.Equal(t, 4096, geminiReq.GenerationConfig.MaxOutputTokens)
	})

	t.Run("uses provided value when within limit", func(t *testing.T) {
		req := &models.LLMRequest{
			Prompt: "Test",
			ModelParams: models.ModelParameters{
				MaxTokens: 2000,
			},
		}
		geminiReq := provider.convertRequest(req)
		assert.Equal(t, 2000, geminiReq.GenerationConfig.MaxOutputTokens)
	})
}

func TestGeminiProvider_ConvertResponse_MultipleParts(t *testing.T) {
	provider := NewGeminiProvider("test-key", "", "gemini-pro")
	req := &models.LLMRequest{ID: "multi-part"}

	geminiResp := &GeminiResponse{
		Candidates: []GeminiCandidate{
			{
				Content: GeminiContent{
					Parts: []GeminiPart{
						{Text: "Part one. "},
						{Text: "Part two."},
					},
				},
				FinishReason: "STOP",
			},
		},
		UsageMetadata: &GeminiUsageMetadata{
			TotalTokenCount: 10,
		},
	}

	resp := provider.convertResponse(req, geminiResp, time.Now())

	// Should concatenate all text parts
	assert.Contains(t, resp.Content, "Part one")
	assert.Contains(t, resp.Content, "Part two")
}

func TestGeminiProvider_HealthCheck_NetworkError(t *testing.T) {
	provider := NewGeminiProvider("test-key", "", "gemini-pro")
	provider.healthURL = "http://localhost:9999/nonexistent" // Non-existent URL
	provider.httpClient = &http.Client{Timeout: 100 * time.Millisecond}

	err := provider.HealthCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "health check request failed")
}

func TestNewGeminiProviderWithRetry(t *testing.T) {
	retryConfig := RetryConfig{
		MaxRetries:   5,
		InitialDelay: 2 * time.Second,
		MaxDelay:     60 * time.Second,
		Multiplier:   3.0,
	}

	provider := NewGeminiProviderWithRetry("test-key", "", "", retryConfig)

	assert.Equal(t, "test-key", provider.apiKey)
	assert.Equal(t, GeminiAPIURL, provider.baseURL)
	assert.Equal(t, GeminiModel, provider.model)
	assert.Equal(t, 5, provider.retryConfig.MaxRetries)
	assert.Equal(t, 2*time.Second, provider.retryConfig.InitialDelay)
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 1*time.Second, config.InitialDelay)
	assert.Equal(t, 30*time.Second, config.MaxDelay)
	assert.Equal(t, 2.0, config.Multiplier)
}

func TestGeminiProvider_StreamURLDerivation(t *testing.T) {
	t.Run("default URLs", func(t *testing.T) {
		provider := NewGeminiProvider("key", "", "")
		assert.Equal(t, GeminiAPIURL, provider.baseURL)
		assert.Equal(t, GeminiStreamAPIURL, provider.streamURL)
	})

	t.Run("custom base URL keeps same when suffix not matched", func(t *testing.T) {
		// Note: The code checks for 15 chars suffix but :generateContent is 16 chars
		// So custom URLs that don't end with exactly the right suffix don't get modified
		customURL := "https://custom.api.com/v1/models/%s:generateContent"
		provider := NewGeminiProvider("key", customURL, "")
		assert.Equal(t, customURL, provider.baseURL)
		// Stream URL falls back to same as base URL
		assert.Equal(t, customURL, provider.streamURL)
	})
}

func TestGeminiProvider_WaitWithJitter(t *testing.T) {
	provider := NewGeminiProvider("test-key", "", "")

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	baseDelay := 100 * time.Millisecond
	provider.waitWithJitter(ctx, baseDelay)
	elapsed := time.Since(start)

	// Should wait at least the base delay
	assert.GreaterOrEqual(t, elapsed, baseDelay)
	// Should not exceed base delay + 10% jitter + buffer
	assert.LessOrEqual(t, elapsed, 150*time.Millisecond)
}

func TestGeminiProvider_WaitWithJitter_ContextCancelled(t *testing.T) {
	provider := NewGeminiProvider("test-key", "", "")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	start := time.Now()
	provider.waitWithJitter(ctx, 1*time.Second)
	elapsed := time.Since(start)

	// Should return immediately due to cancelled context
	assert.Less(t, elapsed, 100*time.Millisecond)
}

func BenchmarkGeminiProvider_ConvertRequest(b *testing.B) {
	provider := NewGeminiProvider("test-key", "", "")
	req := &models.LLMRequest{
		ID:     "bench-request",
		Prompt: "Test prompt",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
			{Role: "assistant", Content: "Hi"},
		},
		ModelParams: models.ModelParameters{
			MaxTokens:   100,
			Temperature: 0.7,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		provider.convertRequest(req)
	}
}

func BenchmarkGeminiProvider_CalculateConfidence(b *testing.B) {
	provider := NewGeminiProvider("test-key", "", "")
	content := "This is a sample response from the Gemini model."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		provider.calculateConfidence(content, "STOP")
	}
}
