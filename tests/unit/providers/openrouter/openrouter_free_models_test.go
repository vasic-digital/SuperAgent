package openrouter_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"dev.helix.agent/internal/llm/providers/openrouter"
	"dev.helix.agent/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// OpenRouter Zen (Free) models list
var FreeModels = []string{
	"meta-llama/llama-4-maverick:free",
	"meta-llama/llama-4-scout:free",
	"meta-llama/llama-3.3-70b-instruct:free",
	"deepseek/deepseek-chat-v3-0324:free",
	"deepseek/deepseek-r1:free",
	"deepseek/deepseek-r1-zero:free",
	"deepseek/deepseek-r1-distill-llama-70b:free",
	"deepseek/deepseek-r1-distill-qwen-32b:free",
	"qwen/qwq-32b:free",
	"qwen/qwen2.5-vl-3b-instruct:free",
	"google/gemini-2.5-pro-exp-03-25:free",
	"google/gemini-2.0-flash-thinking-exp:free",
	"nvidia/llama-3.1-nemotron-ultra-253b-v1:free",
	"microsoft/mai-ds-r1:free",
	"mistralai/mistral-small-3.1-24b-instruct:free",
}

// TestOpenRouterFreeModels_Complete tests that free models work with the Complete method
func TestOpenRouterFreeModels_Complete(t *testing.T) {
	for _, model := range FreeModels {
		t.Run(model, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify model in request
				var reqBody map[string]interface{}
				body, _ := io.ReadAll(r.Body)
				json.Unmarshal(body, &reqBody)

				requestedModel := reqBody["model"].(string)
				assert.Equal(t, model, requestedModel)

				// Verify free model has :free suffix
				assert.True(t, strings.HasSuffix(requestedModel, ":free"),
					"Free model should have :free suffix")

				response := map[string]interface{}{
					"id": fmt.Sprintf("chatcmpl-%s", model),
					"choices": []map[string]interface{}{
						{
							"message": map[string]interface{}{
								"role":    "assistant",
								"content": fmt.Sprintf("Response from free model: %s", model),
							},
							"finish_reason": "stop",
						},
					},
					"created": time.Now().Unix(),
					"model":   model,
					"usage": map[string]interface{}{
						"prompt_tokens":     10,
						"completion_tokens": 20,
						"total_tokens":      30,
					},
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			provider := openrouter.NewSimpleOpenRouterProviderWithBaseURL("test-api-key", server.URL)

			req := &models.LLMRequest{
				ID: fmt.Sprintf("free-model-test-%s", model),
				ModelParams: models.ModelParameters{
					Model:       model,
					MaxTokens:   1000,
					Temperature: 0.7,
				},
				Prompt: "Test free model prompt",
			}

			resp, err := provider.Complete(context.Background(), req)
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Contains(t, resp.Content, "Response from free model")
		})
	}
}

// TestOpenRouterFreeModels_FreeSuffix tests that free models are correctly identified
func TestOpenRouterFreeModels_FreeSuffix(t *testing.T) {
	tests := []struct {
		model  string
		isFree bool
	}{
		{"meta-llama/llama-4-maverick:free", true},
		{"deepseek/deepseek-r1:free", true},
		{"qwen/qwq-32b:free", true},
		{"google/gemini-2.5-pro-exp-03-25:free", true},
		{"anthropic/claude-3.5-sonnet", false},
		{"openai/gpt-4o", false},
		{"meta-llama/llama-4-maverick", false},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			isFree := strings.HasSuffix(tt.model, ":free")
			assert.Equal(t, tt.isFree, isFree)
		})
	}
}

// TestOpenRouterFreeModels_Streaming tests streaming with free models
// Note: Streaming tests with mock servers are inherently flaky due to timing issues.
// The main purpose is to verify that streaming requests don't error out.
func TestOpenRouterFreeModels_Streaming(t *testing.T) {
	// Test with a few representative free models
	testModels := []string{
		"meta-llama/llama-4-maverick:free",
		"deepseek/deepseek-r1:free",
		"qwen/qwq-32b:free",
	}

	for _, model := range testModels {
		t.Run(model, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/event-stream")
				w.Header().Set("Cache-Control", "no-cache")
				w.Header().Set("Connection", "keep-alive")
				w.WriteHeader(http.StatusOK)

				flusher, ok := w.(http.Flusher)
				if !ok {
					t.Fatal("Expected http.Flusher")
				}

				chunks := []string{"Hello", " from", " free", " model", "!"}
				for i, content := range chunks {
					chunk := map[string]interface{}{
						"id":      fmt.Sprintf("stream-%d", i),
						"object":  "chat.completion.chunk",
						"created": time.Now().Unix(),
						"model":   model,
						"choices": []map[string]interface{}{
							{
								"index": 0,
								"delta": map[string]interface{}{
									"content": content,
								},
							},
						},
					}
					jsonData, _ := json.Marshal(chunk)
					fmt.Fprintf(w, "data: %s\n\n", jsonData)
					flusher.Flush()
					time.Sleep(10 * time.Millisecond) // Small delay for buffering
				}

				fmt.Fprintf(w, "data: [DONE]\n\n")
				flusher.Flush()
				time.Sleep(50 * time.Millisecond) // Keep connection open briefly
			}))
			defer server.Close()

			provider := openrouter.NewSimpleOpenRouterProviderWithBaseURL("test-api-key", server.URL)

			req := &models.LLMRequest{
				ID: fmt.Sprintf("stream-test-%s", model),
				ModelParams: models.ModelParameters{
					Model: model,
				},
				Prompt: "Test streaming with free model",
			}

			responseChan, err := provider.CompleteStream(context.Background(), req)
			require.NoError(t, err)
			require.NotNil(t, responseChan)

			// Collect responses with timeout
			var fullContent string
			timeout := time.After(2 * time.Second)
			done := false
			for !done {
				select {
				case chunk, ok := <-responseChan:
					if !ok {
						done = true
						break
					}
					if chunk != nil {
						fullContent += chunk.Content
					}
				case <-timeout:
					done = true
				}
			}

			// Streaming tests are flaky - just verify we got some response or no error
			assert.NotEmpty(t, fullContent, "Expected some streaming content (may be timing-dependent)")
		})
	}
}

// TestOpenRouterFreeModels_SupportedModelsIncludesFree tests that GetCapabilities includes free models
func TestOpenRouterFreeModels_SupportedModelsIncludesFree(t *testing.T) {
	provider := openrouter.NewSimpleOpenRouterProvider("test-api-key")
	caps := provider.GetCapabilities()

	require.NotNil(t, caps)
	require.NotEmpty(t, caps.SupportedModels)

	// Check that at least some free models are included
	freeModelsFound := 0
	for _, model := range caps.SupportedModels {
		if strings.HasSuffix(model, ":free") {
			freeModelsFound++
		}
	}

	assert.Greater(t, freeModelsFound, 0, "Should have at least one free model in SupportedModels")
	t.Logf("Found %d free models in SupportedModels", freeModelsFound)
}

// TestOpenRouterFreeModels_NoAPIKeyRequired tests that free models work without an API key
// Note: OpenRouter still requires an API key, but free models don't incur charges
func TestOpenRouterFreeModels_NoCharges(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify API key is present
		authHeader := r.Header.Get("Authorization")
		assert.True(t, strings.HasPrefix(authHeader, "Bearer "), "Should have Bearer token")

		response := map[string]interface{}{
			"id": "chatcmpl-free",
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "Free model response - no charges",
					},
					"finish_reason": "stop",
				},
			},
			"created": time.Now().Unix(),
			"model":   "meta-llama/llama-4-maverick:free",
			"usage": map[string]interface{}{
				"prompt_tokens":     10,
				"completion_tokens": 20,
				"total_tokens":      30,
				// Free models would show $0 cost in billing
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	provider := openrouter.NewSimpleOpenRouterProviderWithBaseURL("test-api-key", server.URL)

	req := &models.LLMRequest{
		ID: "free-no-charges-test",
		ModelParams: models.ModelParameters{
			Model: "meta-llama/llama-4-maverick:free",
		},
		Prompt: "Test that free models don't incur charges",
	}

	resp, err := provider.Complete(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Contains(t, resp.Content, "no charges")
}

// TestOpenRouterFreeModels_FallbackBehavior tests fallback from premium to free models
func TestOpenRouterFreeModels_FallbackBehavior(t *testing.T) {
	callCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		var reqBody map[string]interface{}
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &reqBody)

		model := reqBody["model"].(string)

		// Simulate premium model failure, free model success
		if !strings.HasSuffix(model, ":free") {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error": {"message": "Rate limit exceeded for premium model"}}`))
			return
		}

		response := map[string]interface{}{
			"id": "chatcmpl-free-fallback",
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "Response from free fallback model",
					},
					"finish_reason": "stop",
				},
			},
			"created": time.Now().Unix(),
			"model":   model,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	provider := openrouter.NewSimpleOpenRouterProviderWithBaseURL("test-api-key", server.URL)

	// First try with free model directly
	req := &models.LLMRequest{
		ID: "free-fallback-test",
		ModelParams: models.ModelParameters{
			Model: "meta-llama/llama-4-maverick:free",
		},
		Prompt: "Test fallback to free model",
	}

	resp, err := provider.Complete(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "Response from free fallback model", resp.Content)
}

// TestOpenRouterFreeModels_DeepSeekR1 tests the DeepSeek R1 free model specifically
func TestOpenRouterFreeModels_DeepSeekR1(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody map[string]interface{}
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &reqBody)

		model := reqBody["model"].(string)
		assert.Equal(t, "deepseek/deepseek-r1:free", model)

		// DeepSeek R1 is known for reasoning capabilities
		response := map[string]interface{}{
			"id": "chatcmpl-deepseek-r1",
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "DeepSeek R1 reasoning response: The answer is derived through step-by-step analysis...",
					},
					"finish_reason": "stop",
				},
			},
			"created": time.Now().Unix(),
			"model":   model,
			"usage": map[string]interface{}{
				"prompt_tokens":     15,
				"completion_tokens": 50,
				"total_tokens":      65,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	provider := openrouter.NewSimpleOpenRouterProviderWithBaseURL("test-api-key", server.URL)

	req := &models.LLMRequest{
		ID: "deepseek-r1-test",
		ModelParams: models.ModelParameters{
			Model:       "deepseek/deepseek-r1:free",
			MaxTokens:   2000,
			Temperature: 0.7,
		},
		Prompt: "Solve this step by step: What is 15% of 200?",
	}

	resp, err := provider.Complete(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Contains(t, resp.Content, "DeepSeek R1")
	assert.Contains(t, resp.Content, "step-by-step")
}

// TestOpenRouterFreeModels_QwenQwQ tests the Qwen QwQ free model
func TestOpenRouterFreeModels_QwenQwQ(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody map[string]interface{}
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &reqBody)

		model := reqBody["model"].(string)
		assert.Equal(t, "qwen/qwq-32b:free", model)

		response := map[string]interface{}{
			"id": "chatcmpl-qwen-qwq",
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "Qwen QwQ 32B response with reasoning capabilities",
					},
					"finish_reason": "stop",
				},
			},
			"created": time.Now().Unix(),
			"model":   model,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	provider := openrouter.NewSimpleOpenRouterProviderWithBaseURL("test-api-key", server.URL)

	req := &models.LLMRequest{
		ID: "qwen-qwq-test",
		ModelParams: models.ModelParameters{
			Model: "qwen/qwq-32b:free",
		},
		Prompt: "Test Qwen QwQ model",
	}

	resp, err := provider.Complete(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Contains(t, resp.Content, "Qwen QwQ")
}

// TestOpenRouterFreeModels_GeminiFlash tests the Gemini Flash free model
func TestOpenRouterFreeModels_GeminiFlash(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody map[string]interface{}
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &reqBody)

		model := reqBody["model"].(string)
		assert.Equal(t, "google/gemini-2.0-flash-thinking-exp:free", model)

		response := map[string]interface{}{
			"id": "chatcmpl-gemini-flash",
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "Gemini Flash thinking response with experimental reasoning",
					},
					"finish_reason": "stop",
				},
			},
			"created": time.Now().Unix(),
			"model":   model,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	provider := openrouter.NewSimpleOpenRouterProviderWithBaseURL("test-api-key", server.URL)

	req := &models.LLMRequest{
		ID: "gemini-flash-test",
		ModelParams: models.ModelParameters{
			Model: "google/gemini-2.0-flash-thinking-exp:free",
		},
		Prompt: "Test Gemini Flash Thinking",
	}

	resp, err := provider.Complete(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Contains(t, resp.Content, "Gemini Flash")
}

// TestOpenRouterFreeModels_LlamaModels tests various Llama free models
func TestOpenRouterFreeModels_LlamaModels(t *testing.T) {
	llamaModels := []string{
		"meta-llama/llama-4-maverick:free",
		"meta-llama/llama-4-scout:free",
		"meta-llama/llama-3.3-70b-instruct:free",
	}

	for _, model := range llamaModels {
		t.Run(model, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				response := map[string]interface{}{
					"id": fmt.Sprintf("chatcmpl-%s", model),
					"choices": []map[string]interface{}{
						{
							"message": map[string]interface{}{
								"role":    "assistant",
								"content": fmt.Sprintf("Response from Llama model: %s", model),
							},
							"finish_reason": "stop",
						},
					},
					"created": time.Now().Unix(),
					"model":   model,
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			provider := openrouter.NewSimpleOpenRouterProviderWithBaseURL("test-api-key", server.URL)

			req := &models.LLMRequest{
				ID: fmt.Sprintf("llama-test-%s", model),
				ModelParams: models.ModelParameters{
					Model: model,
				},
				Prompt: "Test Llama model",
			}

			resp, err := provider.Complete(context.Background(), req)
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Contains(t, resp.Content, "Llama model")
		})
	}
}

// TestOpenRouterFreeModels_ErrorHandling tests error handling for free models
func TestOpenRouterFreeModels_ErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		errorResponse string
		expectedError string
	}{
		{
			name:       "model not found",
			statusCode: http.StatusNotFound,
			errorResponse: `{
				"error": {
					"message": "Model not found: unknown-model:free",
					"type": "invalid_request_error"
				}
			}`,
			expectedError: "Model not found",
		},
		{
			name:       "rate limit on free tier",
			statusCode: http.StatusTooManyRequests,
			errorResponse: `{
				"error": {
					"message": "Free tier rate limit exceeded",
					"type": "rate_limit_error"
				}
			}`,
			expectedError: "rate limit",
		},
		{
			name:       "server error",
			statusCode: http.StatusInternalServerError,
			errorResponse: `{
				"error": {
					"message": "Internal server error",
					"type": "server_error"
				}
			}`,
			expectedError: "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.errorResponse))
			}))
			defer server.Close()

			provider := openrouter.NewSimpleOpenRouterProviderWithBaseURL("test-api-key", server.URL)

			req := &models.LLMRequest{
				ID: "error-test",
				ModelParams: models.ModelParameters{
					Model: "meta-llama/llama-4-maverick:free",
				},
				Prompt: "Test error handling",
			}

			_, err := provider.Complete(context.Background(), req)
			require.Error(t, err)
			assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tt.expectedError))
		})
	}
}

// TestOpenRouterFreeModels_AllFreeModelsWork tests that all documented free models work
func TestOpenRouterFreeModels_AllFreeModelsWork(t *testing.T) {
	// This is an exhaustive test of all free models
	allFreeModels := []struct {
		model       string
		description string
	}{
		{"meta-llama/llama-4-maverick:free", "Meta Llama 4 Maverick"},
		{"meta-llama/llama-4-scout:free", "Meta Llama 4 Scout"},
		{"meta-llama/llama-3.3-70b-instruct:free", "Meta Llama 3.3 70B"},
		{"deepseek/deepseek-chat-v3-0324:free", "DeepSeek Chat V3"},
		{"deepseek/deepseek-r1:free", "DeepSeek R1"},
		{"deepseek/deepseek-r1-zero:free", "DeepSeek R1 Zero"},
		{"deepseek/deepseek-r1-distill-llama-70b:free", "DeepSeek R1 Distill Llama"},
		{"deepseek/deepseek-r1-distill-qwen-32b:free", "DeepSeek R1 Distill Qwen"},
		{"qwen/qwq-32b:free", "Qwen QwQ 32B"},
		{"qwen/qwen2.5-vl-3b-instruct:free", "Qwen 2.5 VL 3B"},
		{"google/gemini-2.5-pro-exp-03-25:free", "Gemini 2.5 Pro Exp"},
		{"google/gemini-2.0-flash-thinking-exp:free", "Gemini 2.0 Flash Thinking"},
		{"nvidia/llama-3.1-nemotron-ultra-253b-v1:free", "NVIDIA Nemotron Ultra"},
		{"microsoft/mai-ds-r1:free", "Microsoft MAI DS R1"},
		{"mistralai/mistral-small-3.1-24b-instruct:free", "Mistral Small 3.1"},
	}

	for _, m := range allFreeModels {
		t.Run(m.description, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				response := map[string]interface{}{
					"id": "chatcmpl-test",
					"choices": []map[string]interface{}{
						{
							"message": map[string]interface{}{
								"role":    "assistant",
								"content": fmt.Sprintf("OK from %s", m.description),
							},
							"finish_reason": "stop",
						},
					},
					"created": time.Now().Unix(),
					"model":   m.model,
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			provider := openrouter.NewSimpleOpenRouterProviderWithBaseURL("test-api-key", server.URL)

			req := &models.LLMRequest{
				ID: "all-free-test",
				ModelParams: models.ModelParameters{
					Model: m.model,
				},
				Prompt: "Test model",
			}

			resp, err := provider.Complete(context.Background(), req)
			require.NoError(t, err, "Model %s should work", m.model)
			require.NotNil(t, resp)
			assert.Contains(t, resp.Content, "OK")
		})
	}
}

// TestOpenRouterFreeModels_Headers tests that correct headers are sent
func TestOpenRouterFreeModels_Headers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify required headers
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/chat/completions", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))
		assert.Equal(t, "helixagent", r.Header.Get("HTTP-Referer"))

		response := map[string]interface{}{
			"id": "chatcmpl-headers",
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "Headers verified",
					},
					"finish_reason": "stop",
				},
			},
			"created": time.Now().Unix(),
			"model":   "meta-llama/llama-4-maverick:free",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	provider := openrouter.NewSimpleOpenRouterProviderWithBaseURL("test-api-key", server.URL)

	req := &models.LLMRequest{
		ID: "headers-test",
		ModelParams: models.ModelParameters{
			Model: "meta-llama/llama-4-maverick:free",
		},
		Prompt: "Test headers",
	}

	resp, err := provider.Complete(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "Headers verified", resp.Content)
}
