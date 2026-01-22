// Package embedding provides tests for the additional embedding provider implementations.
package embedding

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

// =============================================================================
// Cohere Embedding Tests
// =============================================================================

func TestNewCohereEmbedding(t *testing.T) {
	config := DefaultEmbeddingConfigExtended(ModelTypeCohere)
	config.APIKey = "test-api-key"

	model := NewCohereEmbedding(config)

	assert.NotNil(t, model)
	assert.Equal(t, "cohere/embed-english-v3.0", model.Name())
	assert.Equal(t, 1024, model.Dimension())
}

func TestCohereEmbedding_Dimensions(t *testing.T) {
	testCases := []struct {
		modelName         string
		expectedDimension int
	}{
		{"embed-english-v3.0", 1024},
		{"embed-multilingual-v3.0", 1024},
		{"embed-english-light-v3.0", 384},
		{"embed-multilingual-light-v3.0", 384},
		{"embed-english-v2.0", 4096},
		{"embed-multilingual-v2.0", 768},
	}

	for _, tc := range testCases {
		t.Run(tc.modelName, func(t *testing.T) {
			config := DefaultEmbeddingConfigExtended(ModelTypeCohere)
			config.ModelName = tc.modelName
			model := NewCohereEmbedding(config)
			assert.Equal(t, tc.expectedDimension, model.Dimension())
		})
	}
}

func TestCohereEmbedding_Embed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/embed", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Contains(t, r.Header.Get("Authorization"), "Bearer ")

		// Parse request
		var req CohereEmbedRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "embed-english-v3.0", req.Model)
		assert.Equal(t, "search_document", req.InputType)

		response := CohereEmbedResponse{
			ID:         "test-id",
			Embeddings: [][]float64{make([]float64, 1024)},
			Texts:      req.Texts,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeCohere,
		ModelName:    "embed-english-v3.0",
		APIKey:       "test-key",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewCohereEmbedding(config)
	ctx := context.Background()

	embedding, err := model.Embed(ctx, "test text")

	assert.NoError(t, err)
	assert.Len(t, embedding, 1024)
}

func TestCohereEmbedding_EmbedBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := CohereEmbedResponse{
			Embeddings: [][]float64{
				make([]float64, 1024),
				make([]float64, 1024),
				make([]float64, 1024),
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeCohere,
		ModelName:    "embed-english-v3.0",
		APIKey:       "test-key",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewCohereEmbedding(config)
	ctx := context.Background()

	embeddings, err := model.EmbedBatch(ctx, []string{"text1", "text2", "text3"})

	assert.NoError(t, err)
	assert.Len(t, embeddings, 3)
	for _, emb := range embeddings {
		assert.Len(t, emb, 1024)
	}
}

func TestCohereEmbedding_EmbeddingsObjResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return embeddings in embeddings_by_type format
		response := map[string]interface{}{
			"id": "test-id",
			"embeddings_by_type": map[string]interface{}{
				"float": [][]float64{make([]float64, 1024)},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeCohere,
		ModelName:    "embed-english-v3.0",
		APIKey:       "test-key",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewCohereEmbedding(config)
	ctx := context.Background()

	embedding, err := model.Embed(ctx, "test text")

	assert.NoError(t, err)
	assert.Len(t, embedding, 1024)
}

func TestCohereEmbedding_Cache(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		response := CohereEmbedResponse{
			Embeddings: [][]float64{make([]float64, 1024)},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeCohere,
		ModelName:    "embed-english-v3.0",
		APIKey:       "test-key",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: true,
		CacheSize:    100,
	}

	model := NewCohereEmbedding(config)
	ctx := context.Background()

	// First call
	_, err := model.Embed(ctx, "test text")
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)

	// Second call - should use cache
	_, err = model.Embed(ctx, "test text")
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)
}

func TestCohereEmbedding_ErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message": "invalid api key"}`))
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeCohere,
		ModelName:    "embed-english-v3.0",
		APIKey:       "invalid-key",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewCohereEmbedding(config)
	ctx := context.Background()

	_, err := model.Embed(ctx, "test text")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}

func TestCohereEmbedding_Close(t *testing.T) {
	config := DefaultEmbeddingConfigExtended(ModelTypeCohere)
	model := NewCohereEmbedding(config)

	err := model.Close()
	assert.NoError(t, err)
}

// =============================================================================
// Voyage Embedding Tests
// =============================================================================

func TestNewVoyageEmbedding(t *testing.T) {
	config := DefaultEmbeddingConfigExtended(ModelTypeVoyage)
	config.APIKey = "test-api-key"

	model := NewVoyageEmbedding(config)

	assert.NotNil(t, model)
	assert.Equal(t, "voyage/voyage-3", model.Name())
	assert.Equal(t, 1024, model.Dimension())
}

func TestVoyageEmbedding_Dimensions(t *testing.T) {
	testCases := []struct {
		modelName         string
		expectedDimension int
	}{
		{"voyage-3", 1024},
		{"voyage-3-lite", 512},
		{"voyage-code-3", 1024},
		{"voyage-finance-2", 1024},
		{"voyage-law-2", 1024},
		{"voyage-large-2", 1536},
		{"voyage-large-2-instruct", 1536},
		{"voyage-2", 1024},
	}

	for _, tc := range testCases {
		t.Run(tc.modelName, func(t *testing.T) {
			config := DefaultEmbeddingConfigExtended(ModelTypeVoyage)
			config.ModelName = tc.modelName
			model := NewVoyageEmbedding(config)
			assert.Equal(t, tc.expectedDimension, model.Dimension())
		})
	}
}

func TestVoyageEmbedding_Embed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/embeddings", r.URL.Path)
		assert.Contains(t, r.Header.Get("Authorization"), "Bearer ")

		// Parse request
		var req VoyageEmbedRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "voyage-3", req.Model)
		assert.Equal(t, "document", req.InputType)

		response := VoyageEmbedResponse{
			Object: "list",
			Data: []struct {
				Object    string    `json:"object"`
				Embedding []float64 `json:"embedding"`
				Index     int       `json:"index"`
			}{
				{Object: "embedding", Embedding: make([]float64, 1024), Index: 0},
			},
			Model: "voyage-3",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeVoyage,
		ModelName:    "voyage-3",
		APIKey:       "test-key",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewVoyageEmbedding(config)
	ctx := context.Background()

	embedding, err := model.Embed(ctx, "test text")

	assert.NoError(t, err)
	assert.Len(t, embedding, 1024)
}

func TestVoyageEmbedding_EmbedBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := VoyageEmbedResponse{
			Object: "list",
			Data: []struct {
				Object    string    `json:"object"`
				Embedding []float64 `json:"embedding"`
				Index     int       `json:"index"`
			}{
				{Object: "embedding", Embedding: make([]float64, 1024), Index: 0},
				{Object: "embedding", Embedding: make([]float64, 1024), Index: 1},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeVoyage,
		ModelName:    "voyage-3",
		APIKey:       "test-key",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewVoyageEmbedding(config)
	ctx := context.Background()

	embeddings, err := model.EmbedBatch(ctx, []string{"text1", "text2"})

	assert.NoError(t, err)
	assert.Len(t, embeddings, 2)
}

func TestVoyageEmbedding_Cache(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		response := VoyageEmbedResponse{
			Data: []struct {
				Object    string    `json:"object"`
				Embedding []float64 `json:"embedding"`
				Index     int       `json:"index"`
			}{
				{Embedding: make([]float64, 1024), Index: 0},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeVoyage,
		ModelName:    "voyage-3",
		APIKey:       "test-key",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: true,
		CacheSize:    100,
	}

	model := NewVoyageEmbedding(config)
	ctx := context.Background()

	_, err := model.Embed(ctx, "test text")
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)

	_, err = model.Embed(ctx, "test text")
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)
}

func TestVoyageEmbedding_ErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error": "rate limit exceeded"}`))
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeVoyage,
		ModelName:    "voyage-3",
		APIKey:       "test-key",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewVoyageEmbedding(config)
	ctx := context.Background()

	_, err := model.Embed(ctx, "test text")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "429")
}

func TestVoyageEmbedding_Close(t *testing.T) {
	config := DefaultEmbeddingConfigExtended(ModelTypeVoyage)
	model := NewVoyageEmbedding(config)

	err := model.Close()
	assert.NoError(t, err)
}

// =============================================================================
// Jina Embedding Tests
// =============================================================================

func TestNewJinaEmbedding(t *testing.T) {
	config := DefaultEmbeddingConfigExtended(ModelTypeJina)
	config.APIKey = "test-api-key"

	model := NewJinaEmbedding(config)

	assert.NotNil(t, model)
	assert.Equal(t, "jina/jina-embeddings-v3", model.Name())
	assert.Equal(t, 1024, model.Dimension())
}

func TestJinaEmbedding_Dimensions(t *testing.T) {
	testCases := []struct {
		modelName         string
		expectedDimension int
	}{
		{"jina-embeddings-v3", 1024},
		{"jina-embeddings-v2-base-en", 768},
		{"jina-embeddings-v2-small-en", 512},
		{"jina-embeddings-v2-base-de", 768},
		{"jina-clip-v1", 768},
		{"jina-colbert-v2", 128},
	}

	for _, tc := range testCases {
		t.Run(tc.modelName, func(t *testing.T) {
			config := DefaultEmbeddingConfigExtended(ModelTypeJina)
			config.ModelName = tc.modelName
			model := NewJinaEmbedding(config)
			assert.Equal(t, tc.expectedDimension, model.Dimension())
		})
	}
}

func TestJinaEmbedding_Embed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/embeddings", r.URL.Path)
		assert.Contains(t, r.Header.Get("Authorization"), "Bearer ")

		var req JinaEmbedRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "jina-embeddings-v3", req.Model)
		assert.Equal(t, "retrieval.document", req.Task)

		response := JinaEmbedResponse{
			Model: "jina-embeddings-v3",
			Data: []struct {
				Object    string    `json:"object"`
				Index     int       `json:"index"`
				Embedding []float64 `json:"embedding"`
			}{
				{Object: "embedding", Index: 0, Embedding: make([]float64, 1024)},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeJina,
		ModelName:    "jina-embeddings-v3",
		APIKey:       "test-key",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewJinaEmbedding(config)
	ctx := context.Background()

	embedding, err := model.Embed(ctx, "test text")

	assert.NoError(t, err)
	assert.Len(t, embedding, 1024)
}

func TestJinaEmbedding_EmbedBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := JinaEmbedResponse{
			Data: []struct {
				Object    string    `json:"object"`
				Index     int       `json:"index"`
				Embedding []float64 `json:"embedding"`
			}{
				{Index: 0, Embedding: make([]float64, 1024)},
				{Index: 1, Embedding: make([]float64, 1024)},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeJina,
		ModelName:    "jina-embeddings-v3",
		APIKey:       "test-key",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewJinaEmbedding(config)
	ctx := context.Background()

	embeddings, err := model.EmbedBatch(ctx, []string{"text1", "text2"})

	assert.NoError(t, err)
	assert.Len(t, embeddings, 2)
}

func TestJinaEmbedding_Cache(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		response := JinaEmbedResponse{
			Data: []struct {
				Object    string    `json:"object"`
				Index     int       `json:"index"`
				Embedding []float64 `json:"embedding"`
			}{
				{Index: 0, Embedding: make([]float64, 1024)},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeJina,
		ModelName:    "jina-embeddings-v3",
		APIKey:       "test-key",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: true,
		CacheSize:    100,
	}

	model := NewJinaEmbedding(config)
	ctx := context.Background()

	_, err := model.Embed(ctx, "test text")
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)

	_, err = model.Embed(ctx, "test text")
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)
}

func TestJinaEmbedding_ErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error": "invalid token"}`))
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeJina,
		ModelName:    "jina-embeddings-v3",
		APIKey:       "invalid-key",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewJinaEmbedding(config)
	ctx := context.Background()

	_, err := model.Embed(ctx, "test text")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "403")
}

func TestJinaEmbedding_Close(t *testing.T) {
	config := DefaultEmbeddingConfigExtended(ModelTypeJina)
	model := NewJinaEmbedding(config)

	err := model.Close()
	assert.NoError(t, err)
}

// =============================================================================
// Google Embedding Tests
// =============================================================================

func TestNewGoogleEmbedding(t *testing.T) {
	config := DefaultEmbeddingConfigExtended(ModelTypeGoogle)
	config.APIKey = "test-api-key"

	model := NewGoogleEmbedding(config, "test-project", "us-central1")

	assert.NotNil(t, model)
	assert.Equal(t, "google/text-embedding-005", model.Name())
	assert.Equal(t, 768, model.Dimension())
}

func TestGoogleEmbedding_Dimensions(t *testing.T) {
	testCases := []struct {
		modelName         string
		expectedDimension int
	}{
		{"text-embedding-005", 768},
		{"text-multilingual-embedding-002", 768},
		{"textembedding-gecko@003", 768},
		{"text-embedding-004", 768},
	}

	for _, tc := range testCases {
		t.Run(tc.modelName, func(t *testing.T) {
			config := DefaultEmbeddingConfigExtended(ModelTypeGoogle)
			config.ModelName = tc.modelName
			model := NewGoogleEmbedding(config, "test-project", "us-central1")
			assert.Equal(t, tc.expectedDimension, model.Dimension())
		})
	}
}

func TestGoogleEmbedding_Embed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "predict")
		assert.Contains(t, r.Header.Get("Authorization"), "Bearer ")

		response := GoogleEmbedResponse{
			Predictions: []struct {
				Embeddings struct {
					Values     []float64 `json:"values"`
					Statistics struct {
						TokenCount int `json:"token_count"`
					} `json:"statistics"`
				} `json:"embeddings"`
			}{
				{Embeddings: struct {
					Values     []float64 `json:"values"`
					Statistics struct {
						TokenCount int `json:"token_count"`
					} `json:"statistics"`
				}{Values: make([]float64, 768)}},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeGoogle,
		ModelName:    "text-embedding-005",
		APIKey:       "test-key",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewGoogleEmbedding(config, "test-project", "us-central1")
	ctx := context.Background()

	embedding, err := model.Embed(ctx, "test text")

	assert.NoError(t, err)
	assert.Len(t, embedding, 768)
}

func TestGoogleEmbedding_EmbedBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := GoogleEmbedResponse{
			Predictions: []struct {
				Embeddings struct {
					Values     []float64 `json:"values"`
					Statistics struct {
						TokenCount int `json:"token_count"`
					} `json:"statistics"`
				} `json:"embeddings"`
			}{
				{Embeddings: struct {
					Values     []float64 `json:"values"`
					Statistics struct {
						TokenCount int `json:"token_count"`
					} `json:"statistics"`
				}{Values: make([]float64, 768)}},
				{Embeddings: struct {
					Values     []float64 `json:"values"`
					Statistics struct {
						TokenCount int `json:"token_count"`
					} `json:"statistics"`
				}{Values: make([]float64, 768)}},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeGoogle,
		ModelName:    "text-embedding-005",
		APIKey:       "test-key",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewGoogleEmbedding(config, "test-project", "us-central1")
	ctx := context.Background()

	embeddings, err := model.EmbedBatch(ctx, []string{"text1", "text2"})

	assert.NoError(t, err)
	assert.Len(t, embeddings, 2)
}

func TestGoogleEmbedding_Cache(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		response := GoogleEmbedResponse{
			Predictions: []struct {
				Embeddings struct {
					Values     []float64 `json:"values"`
					Statistics struct {
						TokenCount int `json:"token_count"`
					} `json:"statistics"`
				} `json:"embeddings"`
			}{
				{Embeddings: struct {
					Values     []float64 `json:"values"`
					Statistics struct {
						TokenCount int `json:"token_count"`
					} `json:"statistics"`
				}{Values: make([]float64, 768)}},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeGoogle,
		ModelName:    "text-embedding-005",
		APIKey:       "test-key",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: true,
		CacheSize:    100,
	}

	model := NewGoogleEmbedding(config, "test-project", "us-central1")
	ctx := context.Background()

	_, err := model.Embed(ctx, "test text")
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)

	_, err = model.Embed(ctx, "test text")
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)
}

func TestGoogleEmbedding_ErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": {"message": "invalid credentials"}}`))
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeGoogle,
		ModelName:    "text-embedding-005",
		APIKey:       "invalid-key",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewGoogleEmbedding(config, "test-project", "us-central1")
	ctx := context.Background()

	_, err := model.Embed(ctx, "test text")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}

func TestGoogleEmbedding_DefaultLocation(t *testing.T) {
	config := DefaultEmbeddingConfigExtended(ModelTypeGoogle)
	model := NewGoogleEmbedding(config, "test-project", "")

	// Default location should be us-central1
	assert.NotNil(t, model)
}

func TestGoogleEmbedding_Close(t *testing.T) {
	config := DefaultEmbeddingConfigExtended(ModelTypeGoogle)
	model := NewGoogleEmbedding(config, "test-project", "us-central1")

	err := model.Close()
	assert.NoError(t, err)
}

// =============================================================================
// Bedrock Embedding Tests
// =============================================================================

func TestNewBedrockEmbedding(t *testing.T) {
	config := DefaultEmbeddingConfigExtended(ModelTypeBedrock)

	model := NewBedrockEmbedding(config, "us-east-1", "test-access-key", "test-secret-key")

	assert.NotNil(t, model)
	assert.Equal(t, "bedrock/amazon.titan-embed-text-v2:0", model.Name())
	assert.Equal(t, 1024, model.Dimension())
}

func TestBedrockEmbedding_Dimensions(t *testing.T) {
	testCases := []struct {
		modelName         string
		expectedDimension int
	}{
		{"amazon.titan-embed-text-v1", 1536},
		{"amazon.titan-embed-text-v2:0", 1024},
		{"amazon.titan-embed-image-v1", 1024},
		{"cohere.embed-english-v3", 1024},
		{"cohere.embed-multilingual-v3", 1024},
	}

	for _, tc := range testCases {
		t.Run(tc.modelName, func(t *testing.T) {
			config := DefaultEmbeddingConfigExtended(ModelTypeBedrock)
			config.ModelName = tc.modelName
			model := NewBedrockEmbedding(config, "us-east-1", "", "")
			assert.Equal(t, tc.expectedDimension, model.Dimension())
		})
	}
}

func TestBedrockEmbedding_TitanEmbed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "invoke")

		response := BedrockTitanResponse{
			Embedding:      make([]float64, 1024),
			InputTextToken: 5,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeBedrock,
		ModelName:    "amazon.titan-embed-text-v2:0",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewBedrockEmbedding(config, "us-east-1", "test-key-id", "test-secret")
	ctx := context.Background()

	embedding, err := model.Embed(ctx, "test text")

	assert.NoError(t, err)
	assert.Len(t, embedding, 1024)
}

func TestBedrockEmbedding_CohereEmbed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := BedrockCohereResponse{
			Embeddings: [][]float64{make([]float64, 1024)},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeBedrock,
		ModelName:    "cohere.embed-english-v3",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewBedrockEmbedding(config, "us-east-1", "test-key-id", "test-secret")
	ctx := context.Background()

	embedding, err := model.Embed(ctx, "test text")

	assert.NoError(t, err)
	assert.Len(t, embedding, 1024)
}

func TestBedrockEmbedding_EmbedBatch_Titan(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		response := BedrockTitanResponse{
			Embedding:      make([]float64, 1024),
			InputTextToken: 5,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeBedrock,
		ModelName:    "amazon.titan-embed-text-v2:0",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewBedrockEmbedding(config, "us-east-1", "test-key-id", "test-secret")
	ctx := context.Background()

	embeddings, err := model.EmbedBatch(ctx, []string{"text1", "text2"})

	assert.NoError(t, err)
	assert.Len(t, embeddings, 2)
	// Titan doesn't support batch, so it should call individually
	assert.Equal(t, 2, callCount)
}

func TestBedrockEmbedding_EmbedBatch_Cohere(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := BedrockCohereResponse{
			Embeddings: [][]float64{
				make([]float64, 1024),
				make([]float64, 1024),
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeBedrock,
		ModelName:    "cohere.embed-english-v3",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewBedrockEmbedding(config, "us-east-1", "test-key-id", "test-secret")
	ctx := context.Background()

	embeddings, err := model.EmbedBatch(ctx, []string{"text1", "text2"})

	assert.NoError(t, err)
	assert.Len(t, embeddings, 2)
}

func TestBedrockEmbedding_Cache(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		response := BedrockTitanResponse{
			Embedding:      make([]float64, 1024),
			InputTextToken: 5,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeBedrock,
		ModelName:    "amazon.titan-embed-text-v2:0",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: true,
		CacheSize:    100,
	}

	model := NewBedrockEmbedding(config, "us-east-1", "test-key-id", "test-secret")
	ctx := context.Background()

	_, err := model.Embed(ctx, "test text")
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)

	_, err = model.Embed(ctx, "test text")
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)
}

func TestBedrockEmbedding_ErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"message": "Access Denied"}`))
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeBedrock,
		ModelName:    "amazon.titan-embed-text-v2:0",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewBedrockEmbedding(config, "us-east-1", "invalid-key", "invalid-secret")
	ctx := context.Background()

	_, err := model.Embed(ctx, "test text")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "403")
}

func TestBedrockEmbedding_UnsupportedModel(t *testing.T) {
	config := EmbeddingConfig{
		ModelType:    ModelTypeBedrock,
		ModelName:    "unsupported-model",
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewBedrockEmbedding(config, "us-east-1", "key", "secret")
	ctx := context.Background()

	_, err := model.Embed(ctx, "test text")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported model")
}

func TestBedrockEmbedding_DefaultRegion(t *testing.T) {
	config := DefaultEmbeddingConfigExtended(ModelTypeBedrock)
	model := NewBedrockEmbedding(config, "", "key", "secret")

	// Default region should be us-east-1
	assert.NotNil(t, model)
}

func TestBedrockEmbedding_Close(t *testing.T) {
	config := DefaultEmbeddingConfigExtended(ModelTypeBedrock)
	model := NewBedrockEmbedding(config, "us-east-1", "key", "secret")

	err := model.Close()
	assert.NoError(t, err)
}

// =============================================================================
// Helper Function Tests
// =============================================================================

func TestSHA256Hash(t *testing.T) {
	hash := sha256Hash([]byte("test"))
	assert.Len(t, hash, 64) // SHA256 produces 64 hex characters
	assert.NotEmpty(t, hash)
}

func TestHmacSHA256(t *testing.T) {
	result := hmacSHA256([]byte("key"), "data")
	assert.NotEmpty(t, result)
	assert.Len(t, result, 32) // SHA256 produces 32 bytes
}

// =============================================================================
// Factory Function Tests
// =============================================================================

func TestDefaultEmbeddingConfigExtended(t *testing.T) {
	testCases := []struct {
		modelType       ModelType
		expectedModel   string
		expectedBaseURL string
	}{
		{ModelTypeCohere, "embed-english-v3.0", "https://api.cohere.com/v2"},
		{ModelTypeVoyage, "voyage-3", "https://api.voyageai.com/v1"},
		{ModelTypeJina, "jina-embeddings-v3", "https://api.jina.ai/v1"},
		{ModelTypeGoogle, "text-embedding-005", ""},
		{ModelTypeBedrock, "amazon.titan-embed-text-v2:0", ""},
	}

	for _, tc := range testCases {
		t.Run(string(tc.modelType), func(t *testing.T) {
			config := DefaultEmbeddingConfigExtended(tc.modelType)
			assert.Equal(t, tc.modelType, config.ModelType)
			assert.Equal(t, tc.expectedModel, config.ModelName)
			assert.Equal(t, 30*time.Second, config.Timeout)
			assert.Equal(t, 100, config.MaxBatchSize)
			assert.True(t, config.CacheEnabled)
		})
	}
}

func TestDefaultEmbeddingConfigExtended_FallbackToOriginal(t *testing.T) {
	config := DefaultEmbeddingConfigExtended(ModelTypeOpenAI)
	assert.Equal(t, ModelTypeOpenAI, config.ModelType)
	assert.Equal(t, "text-embedding-3-small", config.ModelName)
}

func TestCreateModelExtended(t *testing.T) {
	testCases := []struct {
		modelType ModelType
		expectErr bool
	}{
		{ModelTypeCohere, false},
		{ModelTypeVoyage, false},
		{ModelTypeJina, false},
		{ModelTypeGoogle, false},
		{ModelTypeBedrock, false},
		{ModelTypeOpenAI, false}, // Falls back to original
		{ModelType("totally-unknown"), true},
	}

	for _, tc := range testCases {
		t.Run(string(tc.modelType), func(t *testing.T) {
			config := EmbeddingConfig{
				ModelType:    tc.modelType,
				ModelName:    "test-model",
				Timeout:      5 * time.Second,
				CacheEnabled: false,
			}

			model, err := CreateModelExtended(config)

			if tc.expectErr {
				assert.Error(t, err)
				assert.Nil(t, model)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, model)
			}
		})
	}
}

func TestAvailableModelsExtended(t *testing.T) {
	assert.NotEmpty(t, AvailableModelsExtended)

	// Verify each model has required fields
	for _, model := range AvailableModelsExtended {
		assert.NotEmpty(t, model.Type)
		assert.NotEmpty(t, model.Name)
		assert.Greater(t, model.Dimension, 0)
		assert.NotEmpty(t, model.Description)
	}

	// Verify we have models for each new provider
	providerCount := make(map[ModelType]int)
	for _, model := range AvailableModelsExtended {
		providerCount[model.Type]++
	}

	assert.Greater(t, providerCount[ModelTypeCohere], 0, "Should have Cohere models")
	assert.Greater(t, providerCount[ModelTypeVoyage], 0, "Should have Voyage models")
	assert.Greater(t, providerCount[ModelTypeJina], 0, "Should have Jina models")
	assert.Greater(t, providerCount[ModelTypeGoogle], 0, "Should have Google models")
	assert.Greater(t, providerCount[ModelTypeBedrock], 0, "Should have Bedrock models")
}

func TestGetModelInfoExtended(t *testing.T) {
	info := GetModelInfoExtended()

	assert.NotEmpty(t, info)
	// Should include both original and extended models
	assert.Greater(t, len(info), len(AvailableModels))

	// Verify structure
	for _, item := range info {
		assert.Contains(t, item, "type")
		assert.Contains(t, item, "name")
		assert.Contains(t, item, "dimension")
		assert.Contains(t, item, "description")
	}
}

// =============================================================================
// Context Cancellation Tests
// =============================================================================

func TestCohereEmbedding_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeCohere,
		ModelName:    "embed-english-v3.0",
		APIKey:       "test-key",
		BaseURL:      server.URL,
		Timeout:      100 * time.Millisecond,
		CacheEnabled: false,
	}

	model := NewCohereEmbedding(config)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := model.Embed(ctx, "test text")
	assert.Error(t, err)
}

func TestVoyageEmbedding_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeVoyage,
		ModelName:    "voyage-3",
		APIKey:       "test-key",
		BaseURL:      server.URL,
		Timeout:      100 * time.Millisecond,
		CacheEnabled: false,
	}

	model := NewVoyageEmbedding(config)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := model.Embed(ctx, "test text")
	assert.Error(t, err)
}

// =============================================================================
// No Embedding Returned Tests
// =============================================================================

func TestCohereEmbedding_NoEmbeddingReturned(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return response without embeddings
		response := map[string]interface{}{
			"id": "test-id",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeCohere,
		ModelName:    "embed-english-v3.0",
		APIKey:       "test-key",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewCohereEmbedding(config)
	ctx := context.Background()

	_, err := model.Embed(ctx, "test text")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no embedding")
}

func TestVoyageEmbedding_NoEmbeddingReturned(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := VoyageEmbedResponse{
			Data: []struct {
				Object    string    `json:"object"`
				Embedding []float64 `json:"embedding"`
				Index     int       `json:"index"`
			}{},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeVoyage,
		ModelName:    "voyage-3",
		APIKey:       "test-key",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewVoyageEmbedding(config)
	ctx := context.Background()

	_, err := model.Embed(ctx, "test text")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no embedding")
}

func TestJinaEmbedding_NoEmbeddingReturned(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := JinaEmbedResponse{
			Data: []struct {
				Object    string    `json:"object"`
				Index     int       `json:"index"`
				Embedding []float64 `json:"embedding"`
			}{},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeJina,
		ModelName:    "jina-embeddings-v3",
		APIKey:       "test-key",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewJinaEmbedding(config)
	ctx := context.Background()

	_, err := model.Embed(ctx, "test text")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no embedding")
}

func TestGoogleEmbedding_NoEmbeddingReturned(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := GoogleEmbedResponse{
			Predictions: []struct {
				Embeddings struct {
					Values     []float64 `json:"values"`
					Statistics struct {
						TokenCount int `json:"token_count"`
					} `json:"statistics"`
				} `json:"embeddings"`
			}{},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeGoogle,
		ModelName:    "text-embedding-005",
		APIKey:       "test-key",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewGoogleEmbedding(config, "test-project", "us-central1")
	ctx := context.Background()

	_, err := model.Embed(ctx, "test text")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no embedding")
}

// =============================================================================
// Request Body Verification Tests
// =============================================================================

func TestCohereEmbedding_RequestBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CohereEmbedRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, []string{"test text"}, req.Texts)
		assert.Equal(t, "embed-english-v3.0", req.Model)
		assert.Equal(t, "search_document", req.InputType)
		assert.Equal(t, "END", req.Truncate)

		response := CohereEmbedResponse{
			Embeddings: [][]float64{make([]float64, 1024)},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeCohere,
		ModelName:    "embed-english-v3.0",
		APIKey:       "test-key",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewCohereEmbedding(config)
	ctx := context.Background()

	_, err := model.Embed(ctx, "test text")
	assert.NoError(t, err)
}

func TestJinaEmbedding_RequestBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req JinaEmbedRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, []string{"test text"}, req.Input)
		assert.Equal(t, "jina-embeddings-v3", req.Model)
		assert.Equal(t, "float", req.EncodingFormat)
		assert.Equal(t, "retrieval.document", req.Task)

		response := JinaEmbedResponse{
			Data: []struct {
				Object    string    `json:"object"`
				Index     int       `json:"index"`
				Embedding []float64 `json:"embedding"`
			}{
				{Index: 0, Embedding: make([]float64, 1024)},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeJina,
		ModelName:    "jina-embeddings-v3",
		APIKey:       "test-key",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewJinaEmbedding(config)
	ctx := context.Background()

	_, err := model.Embed(ctx, "test text")
	assert.NoError(t, err)
}
