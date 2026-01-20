// Package integration provides comprehensive integration tests for HelixAgent.
// This file contains verification tests for all LLM providers and Cognee integration.
package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"dev.helix.agent/internal/config"
	"dev.helix.agent/internal/llm/providers/cerebras"
	"dev.helix.agent/internal/llm/providers/claude"
	"dev.helix.agent/internal/llm/providers/deepseek"
	"dev.helix.agent/internal/llm/providers/gemini"
	"dev.helix.agent/internal/llm/providers/mistral"
	"dev.helix.agent/internal/llm/providers/ollama"
	"dev.helix.agent/internal/llm/providers/openrouter"
	"dev.helix.agent/internal/llm/providers/qwen"
	"dev.helix.agent/internal/llm/providers/zai"
	"dev.helix.agent/internal/llm/providers/zen"
	"dev.helix.agent/internal/models"
	"dev.helix.agent/internal/services"
	"dev.helix.agent/internal/verifier"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLLMProviderVerification_AllProviders tests all LLM providers
func TestLLMProviderVerification_AllProviders(t *testing.T) {
	if testing.Short() {
		t.Log("Short mode - skipping full LLM verification test (acceptable)")
		return
	}

	ctx := context.Background()

	t.Run("ClaudeProvider", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.NotEmpty(t, r.Header.Get("x-api-key"))
			assert.Equal(t, "2023-06-01", r.Header.Get("anthropic-version"))

			body, _ := io.ReadAll(r.Body)
			var req claude.ClaudeRequest
			err := json.Unmarshal(body, &req)
			require.NoError(t, err)

			assert.NotEmpty(t, req.Model)
			assert.NotEmpty(t, req.Messages)

			resp := claude.ClaudeResponse{
				ID:   "msg-001",
				Type: "message",
				Role: "assistant",
				Content: []claude.ClaudeContent{
					{Type: "text", Text: "Hello! I'm Claude, an AI assistant."},
				},
				Model:      req.Model,
				StopReason: strPtr("end_turn"),
				Usage: claude.ClaudeUsage{
					InputTokens:  10,
					OutputTokens: 15,
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		provider := claude.NewClaudeProvider("test-api-key", server.URL, "claude-3-sonnet")

		caps := provider.GetCapabilities()
		assert.True(t, caps.SupportsStreaming)
		assert.True(t, caps.SupportsTools)
		assert.True(t, caps.SupportsVision)

		req := &models.LLMRequest{
			ID:     "test-1",
			Prompt: "Hello",
			Messages: []models.Message{
				{Role: "user", Content: "Hello"},
			},
			ModelParams: models.ModelParameters{
				Model:       "claude-3-sonnet",
				MaxTokens:   100,
				Temperature: 0.7,
			},
		}

		resp, err := provider.Complete(ctx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.Content)
		assert.Equal(t, "claude", resp.ProviderID)
		assert.Greater(t, resp.Confidence, 0.0)
	})

	t.Run("DeepSeekProvider", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)

			resp := map[string]interface{}{
				"id":      "chatcmpl-001",
				"object":  "chat.completion",
				"created": time.Now().Unix(),
				"model":   "deepseek-chat",
				"choices": []map[string]interface{}{
					{
						"index": 0,
						"message": map[string]interface{}{
							"role":    "assistant",
							"content": "Hello! I'm DeepSeek.",
						},
						"finish_reason": "stop",
					},
				},
				"usage": map[string]interface{}{
					"prompt_tokens":     10,
					"completion_tokens": 12,
					"total_tokens":      22,
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		provider := deepseek.NewDeepSeekProvider("test-api-key", server.URL, "deepseek-chat")

		caps := provider.GetCapabilities()
		assert.True(t, caps.SupportsStreaming)

		req := &models.LLMRequest{
			ID: "test-2",
			Messages: []models.Message{
				{Role: "user", Content: "Hello"},
			},
			ModelParams: models.ModelParameters{
				Model:       "deepseek-chat",
				MaxTokens:   100,
				Temperature: 0.7,
			},
		}

		resp, err := provider.Complete(ctx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.Content)
	})

	t.Run("GeminiProvider", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := map[string]interface{}{
				"candidates": []map[string]interface{}{
					{
						"content": map[string]interface{}{
							"parts": []map[string]interface{}{
								{"text": "Hello! I'm Gemini."},
							},
							"role": "model",
						},
						"finishReason": "STOP",
					},
				},
				"usageMetadata": map[string]interface{}{
					"promptTokenCount":     10,
					"candidatesTokenCount": 8,
					"totalTokenCount":      18,
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		provider := gemini.NewGeminiProvider("test-api-key", server.URL, "gemini-pro")

		caps := provider.GetCapabilities()
		assert.True(t, caps.SupportsVision)

		req := &models.LLMRequest{
			ID: "test-3",
			Messages: []models.Message{
				{Role: "user", Content: "Hello"},
			},
			ModelParams: models.ModelParameters{
				Model:       "gemini-pro",
				MaxTokens:   100,
				Temperature: 0.7,
			},
		}

		resp, err := provider.Complete(ctx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.Content)
	})

	t.Run("MistralProvider", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := map[string]interface{}{
				"id":      "chatcmpl-001",
				"object":  "chat.completion",
				"created": time.Now().Unix(),
				"model":   "mistral-large",
				"choices": []map[string]interface{}{
					{
						"index": 0,
						"message": map[string]interface{}{
							"role":    "assistant",
							"content": "Hello! I'm Mistral.",
						},
						"finish_reason": "stop",
					},
				},
				"usage": map[string]interface{}{
					"prompt_tokens":     10,
					"completion_tokens": 10,
					"total_tokens":      20,
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		provider := mistral.NewMistralProvider("test-api-key", server.URL, "mistral-large")

		req := &models.LLMRequest{
			ID: "test-4",
			Messages: []models.Message{
				{Role: "user", Content: "Hello"},
			},
			ModelParams: models.ModelParameters{
				Model:       "mistral-large",
				MaxTokens:   100,
				Temperature: 0.7,
			},
		}

		resp, err := provider.Complete(ctx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.Content)
	})

	t.Run("QwenProvider", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := map[string]interface{}{
				"id":      "chatcmpl-001",
				"object":  "chat.completion",
				"created": time.Now().Unix(),
				"model":   "qwen-max",
				"choices": []map[string]interface{}{
					{
						"index": 0,
						"message": map[string]interface{}{
							"role":    "assistant",
							"content": "Hello! I'm Qwen.",
						},
						"finish_reason": "stop",
					},
				},
				"usage": map[string]interface{}{
					"prompt_tokens":     10,
					"completion_tokens": 8,
					"total_tokens":      18,
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		provider := qwen.NewQwenProvider("test-api-key", server.URL, "qwen-max")

		req := &models.LLMRequest{
			ID: "test-5",
			Messages: []models.Message{
				{Role: "user", Content: "Hello"},
			},
			ModelParams: models.ModelParameters{
				Model:       "qwen-max",
				MaxTokens:   100,
				Temperature: 0.7,
			},
		}

		resp, err := provider.Complete(ctx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.Content)
	})

	t.Run("OpenRouterProvider", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := map[string]interface{}{
				"id":      "gen-001",
				"object":  "chat.completion",
				"created": time.Now().Unix(),
				"model":   "openai/gpt-4",
				"choices": []map[string]interface{}{
					{
						"index": 0,
						"message": map[string]interface{}{
							"role":    "assistant",
							"content": "Hello! I'm via OpenRouter.",
						},
						"finish_reason": "stop",
					},
				},
				"usage": map[string]interface{}{
					"prompt_tokens":     10,
					"completion_tokens": 11,
					"total_tokens":      21,
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		provider := openrouter.NewSimpleOpenRouterProviderWithBaseURL("test-api-key", server.URL)

		req := &models.LLMRequest{
			ID: "test-6",
			Messages: []models.Message{
				{Role: "user", Content: "Hello"},
			},
			ModelParams: models.ModelParameters{
				Model:       "openai/gpt-4",
				MaxTokens:   100,
				Temperature: 0.7,
			},
		}

		resp, err := provider.Complete(ctx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.Content)
	})

	t.Run("ZAIProvider", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := map[string]interface{}{
				"id":      "chatcmpl-001",
				"object":  "chat.completion",
				"created": time.Now().Unix(),
				"model":   "zai-turbo",
				"choices": []map[string]interface{}{
					{
						"index": 0,
						"message": map[string]interface{}{
							"role":    "assistant",
							"content": "Hello! I'm ZAI.",
						},
						"finish_reason": "stop",
					},
				},
				"usage": map[string]interface{}{
					"prompt_tokens":     10,
					"completion_tokens": 7,
					"total_tokens":      17,
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		provider := zai.NewZAIProvider("test-api-key", server.URL, "zai-turbo")

		req := &models.LLMRequest{
			ID: "test-7",
			Messages: []models.Message{
				{Role: "user", Content: "Hello"},
			},
			ModelParams: models.ModelParameters{
				Model:       "zai-turbo",
				MaxTokens:   100,
				Temperature: 0.7,
			},
		}

		resp, err := provider.Complete(ctx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.Content)
	})

	t.Run("CerebrasProvider", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := map[string]interface{}{
				"id":      "chatcmpl-001",
				"object":  "chat.completion",
				"created": time.Now().Unix(),
				"model":   "llama-3.3-70b",
				"choices": []map[string]interface{}{
					{
						"index": 0,
						"message": map[string]interface{}{
							"role":    "assistant",
							"content": "Hello! I'm powered by Cerebras.",
						},
						"finish_reason": "stop",
					},
				},
				"usage": map[string]interface{}{
					"prompt_tokens":     10,
					"completion_tokens": 9,
					"total_tokens":      19,
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		provider := cerebras.NewCerebrasProvider("test-api-key", server.URL, "llama-3.3-70b")

		caps := provider.GetCapabilities()
		assert.True(t, caps.SupportsStreaming)

		req := &models.LLMRequest{
			ID: "test-8",
			Messages: []models.Message{
				{Role: "user", Content: "Hello"},
			},
			ModelParams: models.ModelParameters{
				Model:       "llama-3.3-70b",
				MaxTokens:   100,
				Temperature: 0.7,
			},
		}

		resp, err := provider.Complete(ctx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.Content)
	})

	t.Run("ZenProvider", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.Header.Get("X-Device-Id"))

			resp := map[string]interface{}{
				"id":      "chatcmpl-001",
				"object":  "chat.completion",
				"created": time.Now().Unix(),
				"model":   "opencode/grok-code",
				"choices": []map[string]interface{}{
					{
						"index": 0,
						"message": map[string]interface{}{
							"role":    "assistant",
							"content": "Hello! I'm Zen (OpenCode).",
						},
						"finish_reason": "stop",
					},
				},
				"usage": map[string]interface{}{
					"prompt_tokens":     10,
					"completion_tokens": 9,
					"total_tokens":      19,
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		provider := zen.NewZenProvider("", server.URL, "opencode/grok-code")

		caps := provider.GetCapabilities()
		assert.True(t, caps.SupportsStreaming)

		req := &models.LLMRequest{
			ID: "test-9",
			Messages: []models.Message{
				{Role: "user", Content: "Hello"},
			},
			ModelParams: models.ModelParameters{
				Model:       "opencode/grok-code",
				MaxTokens:   100,
				Temperature: 0.7,
			},
		}

		resp, err := provider.Complete(ctx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.Content)
	})

	t.Run("OllamaProvider", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Path, "/api/tags") {
				resp := map[string]interface{}{
					"models": []map[string]interface{}{
						{"name": "llama3.2"},
						{"name": "mistral"},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
				return
			}

			resp := map[string]interface{}{
				"model":      "llama3.2",
				"created_at": time.Now().Format(time.RFC3339),
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": "Hello! I'm running on Ollama.",
				},
				"done":                 true,
				"done_reason":          "stop",
				"total_duration":       1000000000,
				"load_duration":        100000000,
				"prompt_eval_count":    10,
				"prompt_eval_duration": 200000000,
				"eval_count":           12,
				"eval_duration":        700000000,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		provider := ollama.NewOllamaProvider(server.URL, "llama3.2")

		caps := provider.GetCapabilities()
		assert.True(t, caps.SupportsStreaming)

		req := &models.LLMRequest{
			ID: "test-10",
			Messages: []models.Message{
				{Role: "user", Content: "Hello"},
			},
			ModelParams: models.ModelParameters{
				Model:       "llama3.2",
				MaxTokens:   100,
				Temperature: 0.7,
			},
		}

		resp, err := provider.Complete(ctx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.Content)
	})
}

// TestCogneeIntegrationVerification tests the Cognee service integration
func TestCogneeIntegrationVerification(t *testing.T) {
	if testing.Short() {
		t.Log("Short mode - skipping Cognee integration test (acceptable)")
		return
	}

	ctx := context.Background()
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	t.Run("CogneeServiceConfiguration", func(t *testing.T) {
		cfg := &services.CogneeServiceConfig{
			Enabled:                true,
			BaseURL:                "http://localhost:8000",
			Timeout:                30 * time.Second,
			AuthEmail:              "test@example.com",
			AuthPassword:           "testpass123",
			AutoCognify:            true,
			EnhancePrompts:         true,
			StoreResponses:         true,
			MaxContextSize:         4096,
			RelevanceThreshold:     0.7,
			TemporalAwareness:      true,
			EnableFeedbackLoop:     true,
			EnableGraphReasoning:   true,
			EnableCodeIntelligence: true,
			DefaultSearchLimit:     10,
			DefaultDataset:         "test-dataset",
			SearchTypes:            []string{"CHUNKS", "GRAPH_COMPLETION"},
			CombineSearchResults:   true,
			CacheEnabled:           true,
			CacheTTL:               5 * time.Minute,
			MaxConcurrency:         10,
			BatchSize:              50,
			AsyncProcessing:        true,
		}

		service := services.NewCogneeServiceWithConfig(cfg, logger)
		require.NotNil(t, service)

		config := service.GetConfig()
		assert.True(t, config.Enabled)
		assert.True(t, config.EnableGraphReasoning)
		assert.True(t, config.EnableCodeIntelligence)
		assert.Equal(t, 0.7, config.RelevanceThreshold)
	})

	t.Run("CogneeServiceWithMockAPI", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.URL.Path == "/":
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Hello, World, I am alive!"))

			case r.URL.Path == "/api/v1/auth/login":
				resp := map[string]interface{}{
					"access_token": "test-token-12345",
					"token_type":   "bearer",
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)

			case r.URL.Path == "/api/v1/memify":
				resp := map[string]interface{}{
					"id":          "mem-001",
					"vector_id":   "vec-001",
					"graph_nodes": []string{"node-1", "node-2"},
					"status":      "success",
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)

			case r.URL.Path == "/api/v1/search":
				resp := map[string]interface{}{
					"results": []map[string]interface{}{
						{
							"id":        "result-1",
							"content":   "Relevant search result about LLMs",
							"relevance": 0.85,
						},
						{
							"id":        "result-2",
							"content":   "Another relevant result about AI",
							"relevance": 0.78,
						},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)

			case r.URL.Path == "/api/v1/cognify":
				resp := map[string]interface{}{
					"status":  "success",
					"message": "Cognify operation completed",
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)

			case r.URL.Path == "/api/v1/datasets":
				if r.Method == "GET" {
					resp := map[string]interface{}{
						"datasets": []map[string]interface{}{
							{"name": "default", "description": "Default dataset"},
						},
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(resp)
				} else {
					w.WriteHeader(http.StatusCreated)
					resp := map[string]interface{}{"status": "created"}
					json.NewEncoder(w).Encode(resp)
				}

			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		cfg := &services.CogneeServiceConfig{
			Enabled:              true,
			BaseURL:              server.URL,
			Timeout:              10 * time.Second,
			AuthEmail:            "test@example.com",
			AuthPassword:         "testpass",
			EnhancePrompts:       true,
			StoreResponses:       true,
			MaxContextSize:       4096,
			RelevanceThreshold:   0.5,
			EnableGraphReasoning: true,
			DefaultSearchLimit:   10,
			DefaultDataset:       "default",
			SearchTypes:          []string{"CHUNKS"},
			CombineSearchResults: true,
		}

		service := services.NewCogneeServiceWithConfig(cfg, logger)
		service.SetReady(true)

		isHealthy := service.IsHealthy(ctx)
		assert.True(t, isHealthy)

		memory, err := service.AddMemory(ctx, "Test content about AI and LLMs", "default", "text", map[string]interface{}{
			"source": "test",
		})
		require.NoError(t, err)
		assert.NotNil(t, memory)
		assert.Equal(t, "mem-001", memory.ID)

		searchResult, err := service.SearchMemory(ctx, "LLM testing", "default", 10)
		require.NoError(t, err)
		assert.NotNil(t, searchResult)
		assert.Greater(t, searchResult.TotalResults, 0)

		err = service.Cognify(ctx, []string{"default"})
		require.NoError(t, err)

		stats := service.GetStats()
		assert.Greater(t, stats.TotalMemoriesStored, int64(0))
		assert.Greater(t, stats.TotalSearches, int64(0))
	})

	t.Run("CogneeEnhancedContext", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/":
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			case "/api/v1/auth/login":
				resp := map[string]interface{}{
					"access_token": "test-token",
					"token_type":   "bearer",
				}
				json.NewEncoder(w).Encode(resp)
			case "/api/v1/search":
				resp := map[string]interface{}{
					"results": []map[string]interface{}{
						{
							"id":        "ctx-1",
							"content":   "Previous conversation about Go programming",
							"relevance": 0.9,
						},
					},
				}
				json.NewEncoder(w).Encode(resp)
			default:
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"})
			}
		}))
		defer server.Close()

		cfg := &services.CogneeServiceConfig{
			Enabled:              true,
			BaseURL:              server.URL,
			Timeout:              10 * time.Second,
			EnhancePrompts:       true,
			MaxContextSize:       4096,
			RelevanceThreshold:   0.5,
			EnableGraphReasoning: true,
			DefaultSearchLimit:   5,
			DefaultDataset:       "default",
			SearchTypes:          []string{"CHUNKS"},
			CombineSearchResults: true,
		}

		service := services.NewCogneeServiceWithConfig(cfg, logger)
		service.SetReady(true)

		req := &models.LLMRequest{
			ID:     "test-req-1",
			Prompt: "How do I create a goroutine in Go?",
			Messages: []models.Message{
				{Role: "user", Content: "How do I create a goroutine in Go?"},
			},
		}

		enhanced, err := service.EnhanceRequest(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, enhanced)
		assert.NotEmpty(t, enhanced.OriginalPrompt)
	})
}

// TestStartupVerifierPipeline tests the complete startup verification pipeline
func TestStartupVerifierPipeline(t *testing.T) {
	if testing.Short() {
		t.Log("Short mode - skipping startup verifier test (acceptable)")
		return
	}

	_ = context.Background()
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	t.Run("StartupVerifierConfiguration", func(t *testing.T) {
		cfg := verifier.DefaultStartupConfig()
		require.NotNil(t, cfg)

		assert.Equal(t, 5, cfg.PositionCount)
		assert.Equal(t, 15, cfg.DebateTeamSize)
		assert.Equal(t, 2, cfg.FallbacksPerPosition)
		assert.True(t, cfg.ParallelVerification)
		assert.True(t, cfg.EnableFreeProviders)
	})

	t.Run("StartupVerifierCreation", func(t *testing.T) {
		cfg := verifier.DefaultStartupConfig()
		sv := verifier.NewStartupVerifier(cfg, logger)
		require.NotNil(t, sv)
		assert.False(t, sv.IsInitialized())
	})

	t.Run("ProviderDiscoveryMock", func(t *testing.T) {
		origDeepSeekKey := os.Getenv("DEEPSEEK_API_KEY")
		defer os.Setenv("DEEPSEEK_API_KEY", origDeepSeekKey)

		os.Setenv("DEEPSEEK_API_KEY", "test-deepseek-key-12345")

		cfg := verifier.DefaultStartupConfig()
		cfg.ParallelVerification = false
		cfg.VerificationTimeout = 2 * time.Second

		sv := verifier.NewStartupVerifier(cfg, logger)
		require.NotNil(t, sv)

		providers := sv.GetRankedProviders()
		assert.Empty(t, providers)
	})

	t.Run("DebateTeamSelection", func(t *testing.T) {
		cfg := verifier.DefaultStartupConfig()
		cfg.MinScore = 5.0
		cfg.PositionCount = 5
		cfg.FallbacksPerPosition = 2

		sv := verifier.NewStartupVerifier(cfg, logger)
		require.NotNil(t, sv)

		team := sv.GetDebateTeam()
		assert.Nil(t, team)
	})
}

// TestLLMProviderConcurrency tests concurrent provider access
func TestLLMProviderConcurrency(t *testing.T) {
	if testing.Short() {
		t.Log("Short mode - skipping concurrency test (acceptable)")
		return
	}

	ctx := context.Background()

	var requestCount int
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestCount++
		count := requestCount
		mu.Unlock()

		time.Sleep(10 * time.Millisecond)

		resp := claude.ClaudeResponse{
			ID:   fmt.Sprintf("msg-%d", count),
			Type: "message",
			Role: "assistant",
			Content: []claude.ClaudeContent{
				{Type: "text", Text: fmt.Sprintf("Response %d", count)},
			},
			Model:      "claude-3-sonnet",
			StopReason: strPtr("end_turn"),
			Usage:      claude.ClaudeUsage{InputTokens: 10, OutputTokens: 5},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := claude.NewClaudeProvider("test-key", server.URL, "claude-3-sonnet")

	t.Run("ConcurrentRequests", func(t *testing.T) {
		var wg sync.WaitGroup
		numRequests := 20
		results := make(chan *models.LLMResponse, numRequests)
		errors := make(chan error, numRequests)

		for i := 0; i < numRequests; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()

				req := &models.LLMRequest{
					ID: fmt.Sprintf("concurrent-%d", idx),
					Messages: []models.Message{
						{Role: "user", Content: fmt.Sprintf("Request %d", idx)},
					},
					ModelParams: models.ModelParameters{
						Model:       "claude-3-sonnet",
						MaxTokens:   100,
						Temperature: 0.7,
					},
				}

				resp, err := provider.Complete(ctx, req)
				if err != nil {
					errors <- err
					return
				}
				results <- resp
			}(i)
		}

		wg.Wait()
		close(results)
		close(errors)

		successCount := 0
		for range results {
			successCount++
		}

		errorCount := 0
		for err := range errors {
			t.Logf("Concurrent request error: %v", err)
			errorCount++
		}

		assert.Equal(t, numRequests, successCount+errorCount)
		assert.Equal(t, numRequests, successCount, "All concurrent requests should succeed")
	})
}

// TestProviderToolCalling tests tool/function calling capabilities
func TestProviderToolCalling(t *testing.T) {
	if testing.Short() {
		t.Log("Short mode - skipping tool calling test (acceptable)")
		return
	}

	ctx := context.Background()

	t.Run("ClaudeToolCalling", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			var req claude.ClaudeRequest
			json.Unmarshal(body, &req)

			assert.NotEmpty(t, req.Tools)
			assert.Equal(t, "get_weather", req.Tools[0].Name)

			resp := claude.ClaudeResponse{
				ID:   "msg-tool",
				Type: "message",
				Role: "assistant",
				Content: []claude.ClaudeContent{
					{
						Type: "tool_use",
						ID:   "tool-call-1",
						Name: "get_weather",
						Input: map[string]interface{}{
							"location": "San Francisco",
						},
					},
				},
				Model:      "claude-3-sonnet",
				StopReason: strPtr("tool_use"),
				Usage:      claude.ClaudeUsage{InputTokens: 20, OutputTokens: 15},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		provider := claude.NewClaudeProvider("test-key", server.URL, "claude-3-sonnet")

		req := &models.LLMRequest{
			ID: "test-tool-call",
			Messages: []models.Message{
				{Role: "user", Content: "What's the weather in San Francisco?"},
			},
			ModelParams: models.ModelParameters{
				Model:       "claude-3-sonnet",
				MaxTokens:   100,
				Temperature: 0.7,
			},
			Tools: []models.Tool{
				{
					Type: "function",
					Function: models.ToolFunction{
						Name:        "get_weather",
						Description: "Get current weather for a location",
						Parameters: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"location": map[string]interface{}{
									"type":        "string",
									"description": "City name",
								},
							},
							"required": []string{"location"},
						},
					},
				},
			},
			ToolChoice: "auto",
		}

		resp, err := provider.Complete(ctx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.ToolCalls)
		assert.Equal(t, "get_weather", resp.ToolCalls[0].Function.Name)
	})
}

// TestEndToEndLLMWorkflow tests a complete end-to-end workflow
func TestEndToEndLLMWorkflow(t *testing.T) {
	if testing.Short() {
		t.Log("Short mode - skipping E2E workflow test (acceptable)")
		return
	}

	ctx := context.Background()
	logger := logrus.New()

	t.Run("MultiProviderWorkflow", func(t *testing.T) {
		claudeServer := createMockClaudeServer(t)
		defer claudeServer.Close()

		deepseekServer := createMockDeepSeekServer(t)
		defer deepseekServer.Close()

		geminiServer := createMockGeminiServer(t)
		defer geminiServer.Close()

		claudeProvider := claude.NewClaudeProvider("test-key", claudeServer.URL, "claude-3-sonnet")
		deepseekProvider := deepseek.NewDeepSeekProvider("test-key", deepseekServer.URL, "deepseek-chat")
		geminiProvider := gemini.NewGeminiProvider("test-key", geminiServer.URL, "gemini-pro")

		req := &models.LLMRequest{
			ID: "e2e-test",
			Messages: []models.Message{
				{Role: "user", Content: "Explain the concept of recursion in programming."},
			},
			ModelParams: models.ModelParameters{
				MaxTokens:   200,
				Temperature: 0.7,
			},
		}

		var responses []*models.LLMResponse
		var mu sync.Mutex
		var wg sync.WaitGroup

		providers := []struct {
			name     string
			provider interface {
				Complete(context.Context, *models.LLMRequest) (*models.LLMResponse, error)
			}
		}{
			{"claude", claudeProvider},
			{"deepseek", deepseekProvider},
			{"gemini", geminiProvider},
		}

		for _, p := range providers {
			wg.Add(1)
			go func(name string, provider interface {
				Complete(context.Context, *models.LLMRequest) (*models.LLMResponse, error)
			}) {
				defer wg.Done()

				resp, err := provider.Complete(ctx, req)
				if err != nil {
					logger.WithError(err).WithField("provider", name).Error("Provider failed")
					return
				}

				mu.Lock()
				responses = append(responses, resp)
				mu.Unlock()

				logger.WithFields(logrus.Fields{
					"provider":   name,
					"content":    resp.Content[:llmMinInt(50, len(resp.Content))] + "...",
					"tokens":     resp.TokensUsed,
					"confidence": resp.Confidence,
				}).Info("Provider responded")
			}(p.name, p.provider)
		}

		wg.Wait()

		assert.Equal(t, 3, len(responses), "All three providers should respond")

		for _, resp := range responses {
			assert.NotEmpty(t, resp.Content)
			assert.Greater(t, resp.Confidence, 0.0)
		}
	})
}

func createMockClaudeServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := claude.ClaudeResponse{
			ID:   "msg-claude",
			Type: "message",
			Role: "assistant",
			Content: []claude.ClaudeContent{
				{Type: "text", Text: "Claude explanation of recursion: A function that calls itself."},
			},
			Model:      "claude-3-sonnet",
			StopReason: strPtr("end_turn"),
			Usage:      claude.ClaudeUsage{InputTokens: 20, OutputTokens: 30},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
}

func createMockDeepSeekServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"id":      "chatcmpl-deepseek",
			"object":  "chat.completion",
			"created": time.Now().Unix(),
			"model":   "deepseek-chat",
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "DeepSeek explanation of recursion: Self-referential problem solving.",
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]interface{}{
				"prompt_tokens":     20,
				"completion_tokens": 25,
				"total_tokens":      45,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
}

func createMockGeminiServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"candidates": []map[string]interface{}{
				{
					"content": map[string]interface{}{
						"parts": []map[string]interface{}{
							{"text": "Gemini explanation of recursion: Breaking problems into smaller subproblems."},
						},
						"role": "model",
					},
					"finishReason": "STOP",
				},
			},
			"usageMetadata": map[string]interface{}{
				"promptTokenCount":     20,
				"candidatesTokenCount": 28,
				"totalTokenCount":      48,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
}

func strPtr(s string) *string {
	return &s
}

func llmMinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestCogneeServiceFromConfig tests Cognee service creation from config
func TestCogneeServiceFromConfig(t *testing.T) {
	if testing.Short() {
		t.Log("Short mode - skipping Cognee config test (acceptable)")
		return
	}

	logger := logrus.New()

	t.Run("CreateFromConfig", func(t *testing.T) {
		cfg := &config.Config{
			Cognee: config.CogneeConfig{
				Enabled:     true,
				BaseURL:     "http://localhost:8000",
				APIKey:      "",
				Timeout:     30 * time.Second,
				AutoCognify: true,
			},
		}

		service := services.NewCogneeService(cfg, logger)
		require.NotNil(t, service)

		serviceConfig := service.GetConfig()
		assert.True(t, serviceConfig.Enabled)
		assert.Equal(t, "http://localhost:8000", serviceConfig.BaseURL)
		assert.True(t, serviceConfig.AutoCognify)
	})
}
