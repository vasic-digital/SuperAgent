package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"dev.helix.agent/internal/models"
	"dev.helix.agent/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock services for testing
type mockProviderRegistry struct {
	providers []string
}

func (m *mockProviderRegistry) ListProviders() []string {
	return m.providers
}

func setupCompletionTestRouter(handler *CompletionHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.POST("/v1/completions", handler.HandleCompletion)
	r.POST("/v1/completions/stream", handler.HandleCompletionStream)
	r.POST("/v1/ensemble/completions", handler.HandleEnsembleCompletion)
	r.GET("/v1/models", handler.GetModels)
	r.GET("/v1/providers", handler.GetProviders)

	return r
}

func TestNewCompletionHandler(t *testing.T) {
	t.Run("creates handler with dependencies", func(t *testing.T) {
		logger := logrus.New()
		registry := &services.ProviderRegistry{}
		ensembleService := &services.EnsembleService{}
		config := &models.HandlerConfig{}

		handler := NewCompletionHandler(registry, ensembleService, config, logger)

		assert.NotNil(t, handler)
		assert.Equal(t, registry, handler.providerRegistry)
		assert.Equal(t, ensembleService, handler.ensembleService)
		assert.Equal(t, config, handler.config)
		assert.Equal(t, logger, handler.logger)
	})
}

func TestCompletionHandler_HandleCompletion(t *testing.T) {
	t.Run("returns error for invalid JSON", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		handler := NewCompletionHandler(nil, nil, nil, logger)
		router := setupCompletionTestRouter(handler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/completions", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns error for empty prompt", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		handler := NewCompletionHandler(nil, nil, nil, logger)
		router := setupCompletionTestRouter(handler)

		body := map[string]interface{}{
			"model": "gpt-4",
			// Missing prompt
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/completions", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns error when provider registry not available", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		handler := NewCompletionHandler(nil, nil, nil, logger)
		router := setupCompletionTestRouter(handler)

		body := map[string]interface{}{
			"model":  "gpt-4",
			"prompt": "Hello world",
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/completions", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}

func TestCompletionHandler_HandleCompletionStream(t *testing.T) {
	t.Run("returns error for invalid JSON", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		handler := NewCompletionHandler(nil, nil, nil, logger)
		router := setupCompletionTestRouter(handler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/completions/stream", bytes.NewBuffer([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns error for empty prompt", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		handler := NewCompletionHandler(nil, nil, nil, logger)
		router := setupCompletionTestRouter(handler)

		body := map[string]interface{}{
			"model": "gpt-4",
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/completions/stream", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestCompletionHandler_HandleEnsembleCompletion(t *testing.T) {
	t.Run("returns error for invalid JSON", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		handler := NewCompletionHandler(nil, nil, nil, logger)
		router := setupCompletionTestRouter(handler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/ensemble/completions", bytes.NewBuffer([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns error for empty prompt", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		handler := NewCompletionHandler(nil, nil, nil, logger)
		router := setupCompletionTestRouter(handler)

		body := map[string]interface{}{
			"model": "gpt-4",
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/ensemble/completions", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns error when ensemble service not available", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		handler := NewCompletionHandler(nil, nil, nil, logger)
		router := setupCompletionTestRouter(handler)

		body := map[string]interface{}{
			"model":  "gpt-4",
			"prompt": "Test prompt",
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/ensemble/completions", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}

func TestCompletionHandler_GetModels(t *testing.T) {
	t.Run("returns models list", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		handler := NewCompletionHandler(nil, nil, nil, logger)
		router := setupCompletionTestRouter(handler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/models", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotNil(t, response["models"])
		assert.NotNil(t, response["object"])
	})
}

func TestCompletionHandler_GetProviders(t *testing.T) {
	t.Run("returns providers list", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		handler := NewCompletionHandler(nil, nil, nil, logger)
		router := setupCompletionTestRouter(handler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/providers", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotNil(t, response["providers"])
		assert.NotNil(t, response["object"])
	})
}

func TestCompletionHandler_buildLLMRequest(t *testing.T) {
	t.Run("builds request from completion request", func(t *testing.T) {
		logger := logrus.New()
		handler := NewCompletionHandler(nil, nil, nil, logger)

		req := &CompletionRequest{
			Model:       "gpt-4",
			Prompt:      "Test prompt",
			MaxTokens:   100,
			Temperature: 0.7,
			TopP:        0.9,
			N:           1,
			Stream:      false,
			Stop:        []string{"\n"},
		}

		llmReq := handler.buildLLMRequest(req, "test-session")

		assert.Equal(t, "gpt-4", llmReq.ModelParams.Model)
		assert.Equal(t, "Test prompt", llmReq.Prompt)
		assert.Equal(t, 100, llmReq.ModelParams.MaxTokens)
		assert.Equal(t, 0.7, llmReq.ModelParams.Temperature)
		assert.Equal(t, 0.9, llmReq.ModelParams.TopP)
		assert.Equal(t, "test-session", llmReq.SessionID)
	})

	t.Run("builds request from chat completion request", func(t *testing.T) {
		logger := logrus.New()
		handler := NewCompletionHandler(nil, nil, nil, logger)

		req := &ChatCompletionRequest{
			Model: "gpt-4",
			Messages: []Message{
				{Role: "system", Content: "You are helpful"},
				{Role: "user", Content: "Hello"},
			},
			MaxTokens:   150,
			Temperature: 0.5,
		}

		llmReq := handler.buildLLMRequestFromChat(req, "test-session")

		assert.Equal(t, "gpt-4", llmReq.ModelParams.Model)
		assert.Equal(t, 2, len(llmReq.Messages))
		assert.Equal(t, 150, llmReq.ModelParams.MaxTokens)
	})

	t.Run("uses default values when not specified", func(t *testing.T) {
		logger := logrus.New()
		handler := NewCompletionHandler(nil, nil, nil, logger)

		req := &CompletionRequest{
			Model:  "gpt-4",
			Prompt: "Test",
		}

		llmReq := handler.buildLLMRequest(req, "session")

		// Should use sensible defaults
		assert.Greater(t, llmReq.ModelParams.MaxTokens, 0)
	})
}

func TestCompletionHandler_formatCompletionResponse(t *testing.T) {
	t.Run("formats response correctly", func(t *testing.T) {
		logger := logrus.New()
		handler := NewCompletionHandler(nil, nil, nil, logger)

		llmResp := &models.LLMResponse{
			ID:       "resp-123",
			Content:  "Generated text",
			TokensUsed: 10,
			Model:    "gpt-4",
		}

		response := handler.formatCompletionResponse(llmResp, "test-model")

		assert.Equal(t, "resp-123", response.ID)
		assert.Equal(t, "text_completion", response.Object)
		assert.Equal(t, "test-model", response.Model)
		assert.Equal(t, 1, len(response.Choices))
		assert.Equal(t, "Generated text", response.Choices[0].Text)
	})

	t.Run("handles empty response", func(t *testing.T) {
		logger := logrus.New()
		handler := NewCompletionHandler(nil, nil, nil, logger)

		llmResp := &models.LLMResponse{
			ID:      "resp-456",
			Content: "",
			Model:   "gpt-4",
		}

		response := handler.formatCompletionResponse(llmResp, "gpt-4")

		assert.Equal(t, 1, len(response.Choices))
		assert.Equal(t, "", response.Choices[0].Text)
	})
}

func TestCompletionHandler_formatChatCompletionResponse(t *testing.T) {
	t.Run("formats chat response correctly", func(t *testing.T) {
		logger := logrus.New()
		handler := NewCompletionHandler(nil, nil, nil, logger)

		llmResp := &models.LLMResponse{
			ID:       "chat-resp-123",
			Content:  "Chat response",
			TokensUsed: 15,
			Model:    "gpt-4",
		}

		response := handler.formatChatCompletionResponse(llmResp, "gpt-4")

		assert.Equal(t, "chat-resp-123", response.ID)
		assert.Equal(t, "chat.completion", response.Object)
		assert.Equal(t, 1, len(response.Choices))
		assert.Equal(t, "assistant", response.Choices[0].Message.Role)
		assert.Equal(t, "Chat response", response.Choices[0].Message.Content)
	})
}

func TestCompletionHandler_formatStreamingResponse(t *testing.T) {
	t.Run("formats streaming chunk correctly", func(t *testing.T) {
		logger := logrus.New()
		handler := NewCompletionHandler(nil, nil, nil, logger)

		llmResp := &models.LLMResponse{
			ID:      "stream-123",
			Content: "chunk of text",
			Model:   "gpt-4",
		}

		chunk := handler.formatStreamingResponse(llmResp, "gpt-4", 0)

		assert.Equal(t, "stream-123", chunk.ID)
		assert.Equal(t, "text_completion", chunk.Object)
		assert.Equal(t, 1, len(chunk.Choices))
		assert.Equal(t, "chunk of text", chunk.Choices[0].Text)
	})
}

func TestCompletionHandler_formatChatStreamingResponse(t *testing.T) {
	t.Run("formats chat streaming chunk correctly", func(t *testing.T) {
		logger := logrus.New()
		handler := NewCompletionHandler(nil, nil, nil, logger)

		llmResp := &models.LLMResponse{
			ID:      "chat-stream-123",
			Content: "chat chunk",
			Model:   "gpt-4",
		}

		chunk := handler.formatChatStreamingResponse(llmResp, "gpt-4", 0)

		assert.Equal(t, "chat-stream-123", chunk.ID)
		assert.Equal(t, "chat.completion.chunk", chunk.Object)
		assert.Equal(t, 1, len(chunk.Choices))
		assert.Equal(t, "chat chunk", chunk.Choices[0].Delta.Content)
	})
}

func TestCompletionHandler_selectProvider(t *testing.T) {
	t.Run("returns error when no providers available", func(t *testing.T) {
		logger := logrus.New()
		handler := NewCompletionHandler(nil, nil, nil, logger)

		provider, err := handler.selectProvider("gpt-4")

		assert.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "no providers available")
	})
}

func TestCompletionHandler_selectFallbackProvider(t *testing.T) {
	t.Run("returns error when no fallback available", func(t *testing.T) {
		logger := logrus.New()
		handler := NewCompletionHandler(nil, nil, nil, logger)

		provider, err := handler.selectFallbackProvider("gpt-4", []string{"failed-provider"})

		assert.Error(t, err)
		assert.Nil(t, provider)
	})
}

func TestCompletionHandler_handleProviderError(t *testing.T) {
	t.Run("categorizes different error types", func(t *testing.T) {
		logger := logrus.New()
		handler := NewCompletionHandler(nil, nil, nil, logger)

		tests := []struct {
			name     string
			err      error
			expected string
		}{
			{
				name:     "timeout error",
				err:      errors.New("request timeout"),
				expected: "timeout",
			},
			{
				name:     "rate limit error",
				err:      errors.New("rate limit exceeded"),
				expected: "rate_limit",
			},
			{
				name:     "auth error",
				err:      errors.New("authentication failed"),
				expected: "auth",
			},
			{
				name:     "network error",
				err:      errors.New("connection refused"),
				expected: "network",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				category := handler.handleProviderError(tt.err, "test-provider")
				assert.Equal(t, tt.expected, category)
			})
		}
	})
}

func TestCompletionHandler_enrichResponseWithMetadata(t *testing.T) {
	t.Run("adds metadata to response", func(t *testing.T) {
		logger := logrus.New()
		handler := NewCompletionHandler(nil, nil, nil, logger)

		response := &CompletionResponse{
			ID:     "test-id",
			Model:  "gpt-4",
			Object: "text_completion",
		}

		llmResp := &models.LLMResponse{
			ProviderName:   "openai",
			Confidence:     0.95,
			SelectionScore: 0.92,
			Selected:       true,
		}

		enriched := handler.enrichResponseWithMetadata(response, llmResp)

		assert.NotNil(t, enriched.Metadata)
		assert.Equal(t, "openai", enriched.Metadata["provider"])
		assert.Equal(t, 0.95, enriched.Metadata["confidence"])
		assert.Equal(t, 0.92, enriched.Metadata["selection_score"])
		assert.Equal(t, true, enriched.Metadata["selected"])
	})
}

func TestCompletionHandler_ConcurrentAccess(t *testing.T) {
	t.Run("handles concurrent requests", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		handler := NewCompletionHandler(nil, nil, nil, logger)
		router := setupCompletionTestRouter(handler)

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/v1/models", nil)
				router.ServeHTTP(w, req)
				assert.Equal(t, http.StatusOK, w.Code)
			}()
		}

		wg.Wait()
	})
}

// Helper function tests

func TestValidateCompletionRequest(t *testing.T) {
	t.Run("validates prompt presence", func(t *testing.T) {
		req := &CompletionRequest{
			Model: "gpt-4",
		}
		err := validateCompletionRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "prompt")
	})

	t.Run("validates model presence", func(t *testing.T) {
		req := &CompletionRequest{
			Prompt: "Test",
		}
		err := validateCompletionRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "model")
	})

	t.Run("accepts valid request", func(t *testing.T) {
		req := &CompletionRequest{
			Model:  "gpt-4",
			Prompt: "Test prompt",
		}
		err := validateCompletionRequest(req)
		assert.NoError(t, err)
	})
}

func TestValidateChatCompletionRequest(t *testing.T) {
	t.Run("validates messages presence", func(t *testing.T) {
		req := &ChatCompletionRequest{
			Model: "gpt-4",
		}
		err := validateChatCompletionRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "messages")
	})

	t.Run("validates model presence", func(t *testing.T) {
		req := &ChatCompletionRequest{
			Messages: []Message{{Role: "user", Content: "Hello"}},
		}
		err := validateChatCompletionRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "model")
	})

	t.Run("accepts valid request", func(t *testing.T) {
		req := &ChatCompletionRequest{
			Model:    "gpt-4",
			Messages: []Message{{Role: "user", Content: "Hello"}},
		}
		err := validateChatCompletionRequest(req)
		assert.NoError(t, err)
	})
}
