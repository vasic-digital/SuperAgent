package modelsdev

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_ListModels(t *testing.T) {
	t.Run("successful list with no options", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/models", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)

			response := ModelsListResponse{
				Models: []ModelInfo{
					{ID: "gpt-4", Name: "GPT-4", Provider: "openai"},
					{ID: "claude-3", Name: "Claude 3", Provider: "anthropic"},
				},
				Total: 2,
				Page:  1,
				Limit: 20,
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewClient(&ClientConfig{BaseURL: server.URL})
		result, err := client.ListModels(context.Background(), nil)

		require.NoError(t, err)
		assert.Equal(t, 2, result.Total)
		assert.Len(t, result.Models, 2)
	})

	t.Run("list with all options", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/models", r.URL.Path)
			assert.Equal(t, "openai", r.URL.Query().Get("provider"))
			assert.Equal(t, "gpt", r.URL.Query().Get("search"))
			assert.Equal(t, "chat", r.URL.Query().Get("type"))
			assert.Equal(t, "vision", r.URL.Query().Get("capability"))
			assert.Equal(t, "2", r.URL.Query().Get("page"))
			assert.Equal(t, "10", r.URL.Query().Get("limit"))

			response := ModelsListResponse{
				Models: []ModelInfo{{ID: "gpt-4-vision", Name: "GPT-4 Vision", Provider: "openai"}},
				Total:  1,
				Page:   2,
				Limit:  10,
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewClient(&ClientConfig{BaseURL: server.URL})
		opts := &ListModelsOptions{
			Provider:   "openai",
			Search:     "gpt",
			ModelType:  "chat",
			Capability: "vision",
			Page:       2,
			Limit:      10,
		}
		result, err := client.ListModels(context.Background(), opts)

		require.NoError(t, err)
		assert.Equal(t, 1, result.Total)
	})

	t.Run("error response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(APIError{
				Type:    "server_error",
				Message: "Internal server error",
				Code:    500,
			})
		}))
		defer server.Close()

		client := NewClient(&ClientConfig{BaseURL: server.URL})
		_, err := client.ListModels(context.Background(), nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list models")
	})
}

func TestClient_GetModel(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/models/gpt-4", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)

			response := ModelDetailsResponse{
				Model: ModelInfo{
					ID:            "gpt-4",
					Name:          "GPT-4",
					Provider:      "openai",
					ContextWindow: 8192,
					MaxTokens:     4096,
					Capabilities: ModelCapabilities{
						Vision:          true,
						FunctionCalling: true,
						Streaming:       true,
					},
				},
				Benchmarks: []ModelBenchmark{
					{Name: "MMLU", Type: "knowledge", Score: 86.4, Rank: 1},
				},
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewClient(&ClientConfig{BaseURL: server.URL})
		result, err := client.GetModel(context.Background(), "gpt-4")

		require.NoError(t, err)
		assert.Equal(t, "gpt-4", result.Model.ID)
		assert.Equal(t, "GPT-4", result.Model.Name)
		assert.True(t, result.Model.Capabilities.Vision)
		assert.Len(t, result.Benchmarks, 1)
	})

	t.Run("empty model ID", func(t *testing.T) {
		client := NewClient(&ClientConfig{BaseURL: "http://localhost"})
		_, err := client.GetModel(context.Background(), "")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "model ID is required")
	})

	t.Run("model not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(APIError{
				Type:    "not_found",
				Message: "Model not found",
				Code:    404,
			})
		}))
		defer server.Close()

		client := NewClient(&ClientConfig{BaseURL: server.URL})
		_, err := client.GetModel(context.Background(), "nonexistent")

		assert.Error(t, err)
	})

	t.Run("url escaping", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check that special characters are properly escaped
			assert.Equal(t, "/models/model%2Fwith%2Fslashes", r.URL.EscapedPath())
			json.NewEncoder(w).Encode(ModelDetailsResponse{})
		}))
		defer server.Close()

		client := NewClient(&ClientConfig{BaseURL: server.URL})
		_, _ = client.GetModel(context.Background(), "model/with/slashes")
	})
}

func TestClient_SearchModels(t *testing.T) {
	t.Run("search with query", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "vision model", r.URL.Query().Get("search"))

			response := ModelsListResponse{
				Models: []ModelInfo{
					{ID: "gpt-4-vision", Name: "GPT-4 Vision"},
				},
				Total: 1,
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewClient(&ClientConfig{BaseURL: server.URL})
		result, err := client.SearchModels(context.Background(), "vision model", nil)

		require.NoError(t, err)
		assert.Equal(t, 1, result.Total)
	})

	t.Run("search with additional options", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "gpt", r.URL.Query().Get("search"))
			assert.Equal(t, "openai", r.URL.Query().Get("provider"))

			response := ModelsListResponse{Models: []ModelInfo{}, Total: 0}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewClient(&ClientConfig{BaseURL: server.URL})
		opts := &ListModelsOptions{Provider: "openai"}
		_, err := client.SearchModels(context.Background(), "gpt", opts)

		require.NoError(t, err)
	})
}

func TestClient_ListProviders(t *testing.T) {
	t.Run("successful list", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/providers", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)

			response := ProvidersListResponse{
				Providers: []ProviderInfo{
					{ID: "openai", Name: "OpenAI", DisplayName: "OpenAI", ModelsCount: 10},
					{ID: "anthropic", Name: "Anthropic", DisplayName: "Anthropic", ModelsCount: 5},
				},
				Total: 2,
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewClient(&ClientConfig{BaseURL: server.URL})
		result, err := client.ListProviders(context.Background())

		require.NoError(t, err)
		assert.Equal(t, 2, result.Total)
		assert.Len(t, result.Providers, 2)
	})

	t.Run("error response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer server.Close()

		client := NewClient(&ClientConfig{BaseURL: server.URL})
		_, err := client.ListProviders(context.Background())

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list providers")
	})
}

func TestClient_GetProvider(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/providers/openai", r.URL.Path)

			response := ProviderInfo{
				ID:          "openai",
				Name:        "OpenAI",
				DisplayName: "OpenAI",
				Description: "AI research company",
				ModelsCount: 10,
				Website:     "https://openai.com",
				Features:    []string{"chat", "embeddings", "images"},
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewClient(&ClientConfig{BaseURL: server.URL})
		result, err := client.GetProvider(context.Background(), "openai")

		require.NoError(t, err)
		assert.Equal(t, "openai", result.ID)
		assert.Equal(t, "OpenAI", result.DisplayName)
		assert.Len(t, result.Features, 3)
	})

	t.Run("empty provider ID", func(t *testing.T) {
		client := NewClient(&ClientConfig{BaseURL: "http://localhost"})
		_, err := client.GetProvider(context.Background(), "")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "provider ID is required")
	})

	t.Run("provider not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(APIError{
				Type:    "not_found",
				Message: "Provider not found",
				Code:    404,
			})
		}))
		defer server.Close()

		client := NewClient(&ClientConfig{BaseURL: server.URL})
		_, err := client.GetProvider(context.Background(), "nonexistent")

		assert.Error(t, err)
	})
}

func TestClient_ListProviderModels(t *testing.T) {
	t.Run("successful list", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "openai", r.URL.Query().Get("provider"))

			response := ModelsListResponse{
				Models: []ModelInfo{
					{ID: "gpt-4", Provider: "openai"},
					{ID: "gpt-3.5-turbo", Provider: "openai"},
				},
				Total: 2,
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewClient(&ClientConfig{BaseURL: server.URL})
		result, err := client.ListProviderModels(context.Background(), "openai", nil)

		require.NoError(t, err)
		assert.Equal(t, 2, result.Total)
	})

	t.Run("empty provider ID", func(t *testing.T) {
		client := NewClient(&ClientConfig{BaseURL: "http://localhost"})
		_, err := client.ListProviderModels(context.Background(), "", nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "provider ID is required")
	})

	t.Run("with options", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "anthropic", r.URL.Query().Get("provider"))
			assert.Equal(t, "5", r.URL.Query().Get("limit"))

			json.NewEncoder(w).Encode(ModelsListResponse{})
		}))
		defer server.Close()

		client := NewClient(&ClientConfig{BaseURL: server.URL})
		opts := &ListModelsOptions{Limit: 5}
		_, err := client.ListProviderModels(context.Background(), "anthropic", opts)

		require.NoError(t, err)
	})
}

func TestClient_DoRequest(t *testing.T) {
	t.Run("sets correct headers", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "CustomAgent/1.0", r.Header.Get("User-Agent"))
			assert.Equal(t, "application/json", r.Header.Get("Accept"))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		}))
		defer server.Close()

		client := NewClient(&ClientConfig{
			BaseURL:   server.URL,
			APIKey:    "test-api-key",
			UserAgent: "CustomAgent/1.0",
		})
		_, err := client.ListProviders(context.Background())

		require.NoError(t, err)
	})

	t.Run("handles context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClient(&ClientConfig{BaseURL: server.URL})
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := client.ListProviders(ctx)
		assert.Error(t, err)
	})

	t.Run("handles malformed json response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("{invalid json"))
		}))
		defer server.Close()

		client := NewClient(&ClientConfig{BaseURL: server.URL})
		_, err := client.ListProviders(context.Background())

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode response")
	})

	t.Run("handles non-json error response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte("Bad Gateway"))
		}))
		defer server.Close()

		client := NewClient(&ClientConfig{BaseURL: server.URL})
		_, err := client.ListProviders(context.Background())

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API error")
	})
}

func TestClient_DoPost(t *testing.T) {
	t.Run("sends json body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)

			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			assert.Equal(t, "test-value", body["key"])

			json.NewEncoder(w).Encode(map[string]string{"result": "ok"})
		}))
		defer server.Close()

		client := NewClient(&ClientConfig{BaseURL: server.URL})
		var result map[string]string
		err := client.doPost(context.Background(), "/test", map[string]string{"key": "test-value"}, &result)

		require.NoError(t, err)
		assert.Equal(t, "ok", result["result"])
	})
}
