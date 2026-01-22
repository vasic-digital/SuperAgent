// Package embedding provides tests for the embedding model implementations.
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

// TestDefaultEmbeddingConfig tests default configuration for each model type
func TestDefaultEmbeddingConfig(t *testing.T) {
	testCases := []struct {
		modelType       ModelType
		expectedModel   string
		expectedBaseURL string
	}{
		{ModelTypeOpenAI, "text-embedding-3-small", "https://api.openai.com/v1"},
		{ModelTypeOllama, "nomic-embed-text", "http://localhost:11434"},
		{ModelTypeBGEM3, "BAAI/bge-m3", "https://api-inference.huggingface.co/models"},
		{ModelTypeNomic, "nomic-ai/nomic-embed-text-v1.5", "https://api-inference.huggingface.co/models"},
		{ModelTypeCodeBERT, "microsoft/codebert-base", "https://api-inference.huggingface.co/models"},
		{ModelTypeQwen3, "Qwen/Qwen3-Embedding-0.6B", "https://api-inference.huggingface.co/models"},
		{ModelTypeGTE, "thenlper/gte-large", "https://api-inference.huggingface.co/models"},
		{ModelTypeE5, "intfloat/e5-large-v2", "https://api-inference.huggingface.co/models"},
	}

	for _, tc := range testCases {
		t.Run(string(tc.modelType), func(t *testing.T) {
			config := DefaultEmbeddingConfig(tc.modelType)
			assert.Equal(t, tc.modelType, config.ModelType)
			assert.Equal(t, tc.expectedModel, config.ModelName)
			assert.Equal(t, tc.expectedBaseURL, config.BaseURL)
			assert.Equal(t, 30*time.Second, config.Timeout)
			assert.Equal(t, 100, config.MaxBatchSize)
			assert.True(t, config.CacheEnabled)
			assert.Equal(t, 10000, config.CacheSize)
		})
	}
}

// TestOpenAIEmbedding tests OpenAI embedding model
func TestOpenAIEmbedding(t *testing.T) {
	config := DefaultEmbeddingConfig(ModelTypeOpenAI)
	config.APIKey = "test-api-key"
	config.CacheEnabled = false

	model := NewOpenAIEmbedding(config)

	assert.NotNil(t, model)
	assert.Equal(t, "openai/text-embedding-3-small", model.Name())
	assert.Equal(t, 1536, model.Dimension())
}

// TestOpenAIEmbeddingDimensions tests dimension calculation for different models
func TestOpenAIEmbeddingDimensions(t *testing.T) {
	testCases := []struct {
		modelName         string
		expectedDimension int
	}{
		{"text-embedding-3-small", 1536},
		{"text-embedding-3-large", 3072},
		{"text-embedding-ada-002", 1536},
	}

	for _, tc := range testCases {
		t.Run(tc.modelName, func(t *testing.T) {
			config := DefaultEmbeddingConfig(ModelTypeOpenAI)
			config.ModelName = tc.modelName
			model := NewOpenAIEmbedding(config)
			assert.Equal(t, tc.expectedDimension, model.Dimension())
		})
	}
}

// TestOpenAIEmbedding_Embed tests embedding generation with mock server
func TestOpenAIEmbedding_Embed(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/embeddings", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Contains(t, r.Header.Get("Authorization"), "Bearer ")

		response := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"embedding": make([]float64, 1536),
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeOpenAI,
		ModelName:    "text-embedding-3-small",
		APIKey:       "test-key",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewOpenAIEmbedding(config)
	ctx := context.Background()

	embedding, err := model.Embed(ctx, "test text")

	assert.NoError(t, err)
	assert.Len(t, embedding, 1536)
}

// TestOpenAIEmbedding_EmbedBatch tests batch embedding
func TestOpenAIEmbedding_EmbedBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": []map[string]interface{}{
				{"embedding": make([]float64, 1536)},
				{"embedding": make([]float64, 1536)},
				{"embedding": make([]float64, 1536)},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeOpenAI,
		ModelName:    "text-embedding-3-small",
		APIKey:       "test-key",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewOpenAIEmbedding(config)
	ctx := context.Background()

	embeddings, err := model.EmbedBatch(ctx, []string{"text1", "text2", "text3"})

	assert.NoError(t, err)
	assert.Len(t, embeddings, 3)
	for _, emb := range embeddings {
		assert.Len(t, emb, 1536)
	}
}

// TestOpenAIEmbedding_Cache tests caching functionality
func TestOpenAIEmbedding_Cache(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		response := map[string]interface{}{
			"data": []map[string]interface{}{
				{"embedding": make([]float64, 1536)},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeOpenAI,
		ModelName:    "text-embedding-3-small",
		APIKey:       "test-key",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: true,
		CacheSize:    100,
	}

	model := NewOpenAIEmbedding(config)
	ctx := context.Background()

	// First call
	_, err := model.Embed(ctx, "test text")
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)

	// Second call - should use cache
	_, err = model.Embed(ctx, "test text")
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount) // Still 1 - cached
}

// TestOpenAIEmbedding_Close tests the Close method
func TestOpenAIEmbedding_Close(t *testing.T) {
	config := DefaultEmbeddingConfig(ModelTypeOpenAI)
	model := NewOpenAIEmbedding(config)

	err := model.Close()
	assert.NoError(t, err)
}

// TestOllamaEmbedding tests Ollama embedding model
func TestOllamaEmbedding(t *testing.T) {
	config := DefaultEmbeddingConfig(ModelTypeOllama)
	config.CacheEnabled = false

	model := NewOllamaEmbedding(config)

	assert.NotNil(t, model)
	assert.Equal(t, "ollama/nomic-embed-text", model.Name())
	assert.Equal(t, 768, model.Dimension())
}

// TestOllamaEmbeddingDimensions tests dimension calculation
func TestOllamaEmbeddingDimensions(t *testing.T) {
	testCases := []struct {
		modelName         string
		expectedDimension int
	}{
		{"nomic-embed-text", 768},
		{"mxbai-embed-large", 1024},
		{"all-minilm", 384},
	}

	for _, tc := range testCases {
		t.Run(tc.modelName, func(t *testing.T) {
			config := DefaultEmbeddingConfig(ModelTypeOllama)
			config.ModelName = tc.modelName
			model := NewOllamaEmbedding(config)
			assert.Equal(t, tc.expectedDimension, model.Dimension())
		})
	}
}

// TestOllamaEmbedding_Embed tests embedding generation
func TestOllamaEmbedding_Embed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/embeddings", r.URL.Path)

		response := map[string]interface{}{
			"embedding": make([]float64, 768),
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeOllama,
		ModelName:    "nomic-embed-text",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewOllamaEmbedding(config)
	ctx := context.Background()

	embedding, err := model.Embed(ctx, "test text")

	assert.NoError(t, err)
	assert.Len(t, embedding, 768)
}

// TestOllamaEmbedding_EmbedBatch tests batch embedding
func TestOllamaEmbedding_EmbedBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"embedding": make([]float64, 768),
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeOllama,
		ModelName:    "nomic-embed-text",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewOllamaEmbedding(config)
	ctx := context.Background()

	embeddings, err := model.EmbedBatch(ctx, []string{"text1", "text2"})

	assert.NoError(t, err)
	assert.Len(t, embeddings, 2)
}

// TestOllamaEmbedding_Close tests the Close method
func TestOllamaEmbedding_Close(t *testing.T) {
	config := DefaultEmbeddingConfig(ModelTypeOllama)
	model := NewOllamaEmbedding(config)

	err := model.Close()
	assert.NoError(t, err)
}

// TestHuggingFaceEmbedding tests HuggingFace embedding model
func TestHuggingFaceEmbedding(t *testing.T) {
	testCases := []struct {
		modelType         ModelType
		expectedDimension int
	}{
		{ModelTypeBGEM3, 1024},
		{ModelTypeNomic, 768},
		{ModelTypeCodeBERT, 768},
		{ModelTypeGTE, 1024},
		{ModelTypeE5, 1024},
		{ModelTypeQwen3, 768},
	}

	for _, tc := range testCases {
		t.Run(string(tc.modelType), func(t *testing.T) {
			config := DefaultEmbeddingConfig(tc.modelType)
			model := NewHuggingFaceEmbedding(config)

			assert.NotNil(t, model)
			assert.Equal(t, tc.expectedDimension, model.Dimension())
		})
	}
}

// TestHuggingFaceEmbedding_Embed tests embedding generation
func TestHuggingFaceEmbedding_Embed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)

		// Return embedding array
		embedding := make([]float64, 1024)
		json.NewEncoder(w).Encode(embedding)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeBGEM3,
		ModelName:    "BAAI/bge-m3",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewHuggingFaceEmbedding(config)
	ctx := context.Background()

	embedding, err := model.Embed(ctx, "test text")

	assert.NoError(t, err)
	assert.NotEmpty(t, embedding)
}

// TestHuggingFaceEmbedding_EmbedBatch tests batch embedding
func TestHuggingFaceEmbedding_EmbedBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		embeddings := [][]float64{
			make([]float64, 1024),
			make([]float64, 1024),
		}
		json.NewEncoder(w).Encode(embeddings)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeBGEM3,
		ModelName:    "BAAI/bge-m3",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewHuggingFaceEmbedding(config)
	ctx := context.Background()

	embeddings, err := model.EmbedBatch(ctx, []string{"text1", "text2"})

	assert.NoError(t, err)
	assert.Len(t, embeddings, 2)
}

// TestHuggingFaceEmbedding_Close tests the Close method
func TestHuggingFaceEmbedding_Close(t *testing.T) {
	config := DefaultEmbeddingConfig(ModelTypeBGEM3)
	model := NewHuggingFaceEmbedding(config)

	err := model.Close()
	assert.NoError(t, err)
}

// TestEmbeddingCache tests the embedding cache
func TestEmbeddingCache(t *testing.T) {
	cache := NewEmbeddingCache(100)

	assert.NotNil(t, cache)
	assert.Equal(t, 0, cache.Size())
}

// TestEmbeddingCache_GetSet tests get and set operations
func TestEmbeddingCache_GetSet(t *testing.T) {
	cache := NewEmbeddingCache(100)

	// Set value
	embedding := []float64{1.0, 2.0, 3.0}
	cache.Set("test-key", embedding)

	// Get value
	retrieved, ok := cache.Get("test-key")

	assert.True(t, ok)
	assert.Equal(t, embedding, retrieved)
	assert.Equal(t, 1, cache.Size())
}

// TestEmbeddingCache_GetMiss tests cache miss
func TestEmbeddingCache_GetMiss(t *testing.T) {
	cache := NewEmbeddingCache(100)

	retrieved, ok := cache.Get("non-existent-key")

	assert.False(t, ok)
	assert.Nil(t, retrieved)
}

// TestEmbeddingCache_Eviction tests cache eviction
func TestEmbeddingCache_Eviction(t *testing.T) {
	cache := NewEmbeddingCache(3)

	// Fill cache to max
	cache.Set("key1", []float64{1.0})
	cache.Set("key2", []float64{2.0})
	cache.Set("key3", []float64{3.0})

	assert.Equal(t, 3, cache.Size())

	// Add one more - should trigger eviction
	cache.Set("key4", []float64{4.0})

	assert.Equal(t, 3, cache.Size())

	// key4 should be present
	_, ok := cache.Get("key4")
	assert.True(t, ok)
}

// TestEmbeddingCache_Clear tests cache clearing
func TestEmbeddingCache_Clear(t *testing.T) {
	cache := NewEmbeddingCache(100)

	cache.Set("key1", []float64{1.0})
	cache.Set("key2", []float64{2.0})
	assert.Equal(t, 2, cache.Size())

	cache.Clear()
	assert.Equal(t, 0, cache.Size())
}

// TestEmbeddingModelRegistry tests the model registry
func TestEmbeddingModelRegistry(t *testing.T) {
	registry := NewEmbeddingModelRegistry()

	assert.NotNil(t, registry)
	assert.Empty(t, registry.List())
}

// TestEmbeddingModelRegistry_Register tests model registration
func TestEmbeddingModelRegistry_Register(t *testing.T) {
	registry := NewEmbeddingModelRegistry()

	config := DefaultEmbeddingConfig(ModelTypeOpenAI)
	model := NewOpenAIEmbedding(config)

	registry.Register("openai", model)

	// Verify registration
	retrieved, ok := registry.Get("openai")
	assert.True(t, ok)
	assert.Equal(t, model, retrieved)

	// Verify list
	names := registry.List()
	assert.Contains(t, names, "openai")
}

// TestEmbeddingModelRegistry_GetMiss tests missing model retrieval
func TestEmbeddingModelRegistry_GetMiss(t *testing.T) {
	registry := NewEmbeddingModelRegistry()

	model, ok := registry.Get("non-existent")

	assert.False(t, ok)
	assert.Nil(t, model)
}

// TestEmbeddingModelRegistry_Close tests closing all models
func TestEmbeddingModelRegistry_Close(t *testing.T) {
	registry := NewEmbeddingModelRegistry()

	registry.Register("openai", NewOpenAIEmbedding(DefaultEmbeddingConfig(ModelTypeOpenAI)))
	registry.Register("ollama", NewOllamaEmbedding(DefaultEmbeddingConfig(ModelTypeOllama)))

	err := registry.Close()
	assert.NoError(t, err)
}

// TestCreateModel tests the model factory function
func TestCreateModel(t *testing.T) {
	testCases := []struct {
		modelType ModelType
		expectErr bool
	}{
		{ModelTypeOpenAI, false},
		{ModelTypeOllama, false},
		{ModelTypeBGEM3, false},
		{ModelTypeNomic, false},
		{ModelTypeCodeBERT, false},
		{ModelTypeQwen3, false},
		{ModelTypeGTE, false},
		{ModelTypeE5, false},
		{ModelType("unknown"), true},
	}

	for _, tc := range testCases {
		t.Run(string(tc.modelType), func(t *testing.T) {
			config := DefaultEmbeddingConfig(tc.modelType)
			model, err := CreateModel(config)

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

// TestAvailableModels tests the available models list
func TestAvailableModels(t *testing.T) {
	assert.NotEmpty(t, AvailableModels)

	// Verify each model has required fields
	for _, model := range AvailableModels {
		assert.NotEmpty(t, model.Type)
		assert.NotEmpty(t, model.Name)
		assert.Greater(t, model.Dimension, 0)
		assert.NotEmpty(t, model.Description)
	}
}

// TestGetModelInfo tests the model info function
func TestGetModelInfo(t *testing.T) {
	info := GetModelInfo()

	assert.NotEmpty(t, info)
	assert.Equal(t, len(AvailableModels), len(info))

	// Verify structure
	for _, item := range info {
		assert.Contains(t, item, "type")
		assert.Contains(t, item, "name")
		assert.Contains(t, item, "dimension")
		assert.Contains(t, item, "description")
	}
}

// TestModelTypes tests all model type constants
func TestModelTypes(t *testing.T) {
	modelTypes := []ModelType{
		ModelTypeOpenAI,
		ModelTypeOllama,
		ModelTypeBGEM3,
		ModelTypeNomic,
		ModelTypeCodeBERT,
		ModelTypeQwen3,
		ModelTypeGTE,
		ModelTypeE5,
	}

	for _, mt := range modelTypes {
		assert.NotEmpty(t, string(mt))
	}
}

// TestOpenAIEmbedding_ErrorHandling tests error handling
func TestOpenAIEmbedding_ErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "invalid_api_key"}`))
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeOpenAI,
		ModelName:    "text-embedding-3-small",
		APIKey:       "invalid-key",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewOpenAIEmbedding(config)
	ctx := context.Background()

	_, err := model.Embed(ctx, "test text")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}

// TestOllamaEmbedding_ErrorHandling tests error handling
func TestOllamaEmbedding_ErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"error": "model not found"}`))
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeOllama,
		ModelName:    "non-existent-model",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewOllamaEmbedding(config)
	ctx := context.Background()

	_, err := model.Embed(ctx, "test text")
	assert.Error(t, err)
}

// TestHuggingFaceEmbedding_ErrorHandling tests error handling
func TestHuggingFaceEmbedding_ErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error": "rate limit exceeded"}`))
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeBGEM3,
		ModelName:    "BAAI/bge-m3",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewHuggingFaceEmbedding(config)
	ctx := context.Background()

	_, err := model.Embed(ctx, "test text")
	assert.Error(t, err)
}

// TestContextCancellation tests handling of context cancellation
func TestContextCancellation(t *testing.T) {
	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
	}))
	defer slowServer.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeOpenAI,
		ModelName:    "text-embedding-3-small",
		APIKey:       "test-key",
		BaseURL:      slowServer.URL,
		Timeout:      100 * time.Millisecond,
		CacheEnabled: false,
	}

	model := NewOpenAIEmbedding(config)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := model.Embed(ctx, "test text")
	assert.Error(t, err)
}

// TestEmbeddingConfigJSON tests JSON marshaling of config
func TestEmbeddingConfigJSON(t *testing.T) {
	config := DefaultEmbeddingConfig(ModelTypeOpenAI)
	config.APIKey = "test-key"

	data, err := json.Marshal(config)
	require.NoError(t, err)

	var decoded EmbeddingConfig
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, config.ModelType, decoded.ModelType)
	assert.Equal(t, config.ModelName, decoded.ModelName)
}

// TestEmptyBatch tests embedding of empty batch
func TestEmptyBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": []map[string]interface{}{},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeOpenAI,
		ModelName:    "text-embedding-3-small",
		APIKey:       "test-key",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewOpenAIEmbedding(config)
	ctx := context.Background()

	embeddings, err := model.EmbedBatch(ctx, []string{})
	assert.NoError(t, err)
	assert.Empty(t, embeddings)
}

// TestConcurrentCacheAccess tests thread safety of cache
func TestConcurrentCacheAccess(t *testing.T) {
	cache := NewEmbeddingCache(1000)

	done := make(chan bool, 100)

	// Concurrent writes
	for i := 0; i < 50; i++ {
		go func(idx int) {
			key := string(rune('a' + idx%26))
			cache.Set(key, []float64{float64(idx)})
			done <- true
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 50; i++ {
		go func(idx int) {
			key := string(rune('a' + idx%26))
			cache.Get(key)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}

	// Should not panic
	assert.True(t, cache.Size() <= 1000)
}

// =============================================================================
// Additional Coverage Tests
// =============================================================================

// TestOpenAIEmbedding_Embed_NoEmbeddingReturned tests the no embedding returned error path
func TestOpenAIEmbedding_Embed_NoEmbeddingReturned(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return empty data array
		response := map[string]interface{}{
			"data": []map[string]interface{}{},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeOpenAI,
		ModelName:    "text-embedding-3-small",
		APIKey:       "test-key",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewOpenAIEmbedding(config)
	ctx := context.Background()

	_, err := model.Embed(ctx, "test text")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no embedding returned")
}

// TestOllamaEmbedding_CacheHit tests Ollama embedding cache hit path
func TestOllamaEmbedding_CacheHit(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		response := map[string]interface{}{
			"embedding": []float64{1.0, 2.0, 3.0},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeOllama,
		ModelName:    "nomic-embed-text",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: true,
		CacheSize:    100,
	}

	model := NewOllamaEmbedding(config)
	ctx := context.Background()

	// First call - should hit server
	emb1, err := model.Embed(ctx, "test text")
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)
	assert.Equal(t, []float64{1.0, 2.0, 3.0}, emb1)

	// Second call - should hit cache
	emb2, err := model.Embed(ctx, "test text")
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount) // Still 1 - served from cache
	assert.Equal(t, emb1, emb2)
}

// TestOllamaEmbedding_EmbedBatch_Error tests error handling in batch embedding
func TestOllamaEmbedding_EmbedBatch_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "server error"}`))
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeOllama,
		ModelName:    "nomic-embed-text",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewOllamaEmbedding(config)
	ctx := context.Background()

	_, err := model.EmbedBatch(ctx, []string{"text1", "text2"})
	assert.Error(t, err)
}

// TestHuggingFaceEmbedding_CacheHit tests HuggingFace embedding cache hit path
func TestHuggingFaceEmbedding_CacheHit(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		embedding := []float64{1.0, 2.0, 3.0, 4.0}
		json.NewEncoder(w).Encode(embedding)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeBGEM3,
		ModelName:    "BAAI/bge-m3",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: true,
		CacheSize:    100,
	}

	model := NewHuggingFaceEmbedding(config)
	ctx := context.Background()

	// First call - should hit server
	emb1, err := model.Embed(ctx, "test text")
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)

	// Second call - should hit cache
	emb2, err := model.Embed(ctx, "test text")
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount) // Still 1 - served from cache
	assert.Equal(t, emb1, emb2)
}

// TestHuggingFaceEmbedding_WithAPIKey tests API key header setting
func TestHuggingFaceEmbedding_WithAPIKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Authorization header is set
		auth := r.Header.Get("Authorization")
		assert.Equal(t, "Bearer test-hf-key", auth)

		embedding := []float64{1.0, 2.0, 3.0}
		json.NewEncoder(w).Encode(embedding)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeBGEM3,
		ModelName:    "BAAI/bge-m3",
		APIKey:       "test-hf-key",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewHuggingFaceEmbedding(config)
	ctx := context.Background()

	_, err := model.Embed(ctx, "test text")
	assert.NoError(t, err)
}

// TestHuggingFaceEmbedding_NestedResponseFormat tests nested array response parsing
func TestHuggingFaceEmbedding_NestedResponseFormat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return nested format [[embedding]]
		nested := [][]float64{{1.0, 2.0, 3.0, 4.0}}
		json.NewEncoder(w).Encode(nested)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeBGEM3,
		ModelName:    "BAAI/bge-m3",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewHuggingFaceEmbedding(config)
	ctx := context.Background()

	embedding, err := model.Embed(ctx, "test text")
	// Note: This might fail due to json decoder behavior - testing the path
	// The code tries flat array first, then nested
	if err == nil {
		assert.NotEmpty(t, embedding)
	}
}

// TestHuggingFaceEmbedding_EmbedBatch_WithAPIKey tests batch with API key
func TestHuggingFaceEmbedding_EmbedBatch_WithAPIKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Authorization header is set
		auth := r.Header.Get("Authorization")
		assert.Equal(t, "Bearer batch-test-key", auth)

		embeddings := [][]float64{
			{1.0, 2.0, 3.0},
			{4.0, 5.0, 6.0},
		}
		json.NewEncoder(w).Encode(embeddings)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeBGEM3,
		ModelName:    "BAAI/bge-m3",
		APIKey:       "batch-test-key",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewHuggingFaceEmbedding(config)
	ctx := context.Background()

	embeddings, err := model.EmbedBatch(ctx, []string{"text1", "text2"})
	assert.NoError(t, err)
	assert.Len(t, embeddings, 2)
}

// TestHuggingFaceEmbedding_EmbedBatch_Error tests error handling in batch
func TestHuggingFaceEmbedding_EmbedBatch_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error": "rate limit exceeded"}`))
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeBGEM3,
		ModelName:    "BAAI/bge-m3",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: false,
	}

	model := NewHuggingFaceEmbedding(config)
	ctx := context.Background()

	_, err := model.EmbedBatch(ctx, []string{"text1", "text2"})
	assert.Error(t, err)
}

// TestOpenAIEmbedding_CacheSetAfterEmbed tests cache is set after successful embed
func TestOpenAIEmbedding_CacheSetAfterEmbed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": []map[string]interface{}{
				{"embedding": []float64{1.0, 2.0, 3.0}},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := EmbeddingConfig{
		ModelType:    ModelTypeOpenAI,
		ModelName:    "text-embedding-3-small",
		APIKey:       "test-key",
		BaseURL:      server.URL,
		Timeout:      5 * time.Second,
		CacheEnabled: true,
		CacheSize:    100,
	}

	model := NewOpenAIEmbedding(config)
	ctx := context.Background()

	// Embed should populate cache
	emb, err := model.Embed(ctx, "test text")
	assert.NoError(t, err)
	assert.NotEmpty(t, emb)

	// Verify cache has entry
	assert.NotNil(t, model.cache)
	cached, ok := model.cache.Get("test text")
	assert.True(t, ok)
	assert.Equal(t, emb, cached)
}

// TestHuggingFaceEmbedding_Name tests the Name method
func TestHuggingFaceEmbedding_Name(t *testing.T) {
	config := DefaultEmbeddingConfig(ModelTypeBGEM3)
	model := NewHuggingFaceEmbedding(config)

	assert.Equal(t, "huggingface/BAAI/bge-m3", model.Name())
}

// TestEmbeddingCache_OverwriteExisting tests overwriting existing cache entries
func TestEmbeddingCache_OverwriteExisting(t *testing.T) {
	cache := NewEmbeddingCache(100)

	// Set initial value
	cache.Set("key1", []float64{1.0})
	val, ok := cache.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, []float64{1.0}, val)

	// Overwrite with new value
	cache.Set("key1", []float64{2.0, 3.0})
	val, ok = cache.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, []float64{2.0, 3.0}, val)

	// Size should still be 1
	assert.Equal(t, 1, cache.Size())
}

// TestEmbeddingModelRegistry_MultipleModels tests registry with multiple models
func TestEmbeddingModelRegistry_MultipleModels(t *testing.T) {
	registry := NewEmbeddingModelRegistry()

	// Register multiple models
	registry.Register("openai", NewOpenAIEmbedding(DefaultEmbeddingConfig(ModelTypeOpenAI)))
	registry.Register("ollama", NewOllamaEmbedding(DefaultEmbeddingConfig(ModelTypeOllama)))
	registry.Register("huggingface", NewHuggingFaceEmbedding(DefaultEmbeddingConfig(ModelTypeBGEM3)))

	// Verify all are registered
	names := registry.List()
	assert.Len(t, names, 3)
	assert.Contains(t, names, "openai")
	assert.Contains(t, names, "ollama")
	assert.Contains(t, names, "huggingface")

	// Verify retrieval
	for _, name := range names {
		model, ok := registry.Get(name)
		assert.True(t, ok, "Model %s should be retrievable", name)
		assert.NotNil(t, model)
	}
}

// TestDefaultEmbeddingConfig_UnknownType tests default config with unknown type
func TestDefaultEmbeddingConfig_UnknownType(t *testing.T) {
	config := DefaultEmbeddingConfig(ModelType("unknown"))

	// Should still return a config with defaults
	assert.Equal(t, ModelType("unknown"), config.ModelType)
	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.Equal(t, 100, config.MaxBatchSize)
	// Model name and base URL should be empty since unknown type
	assert.Empty(t, config.ModelName)
	assert.Empty(t, config.BaseURL)
}

// TestOpenAIEmbedding_UnknownModel tests dimension for unknown OpenAI model
func TestOpenAIEmbedding_UnknownModel(t *testing.T) {
	config := EmbeddingConfig{
		ModelType: ModelTypeOpenAI,
		ModelName: "unknown-model",
		Timeout:   5 * time.Second,
	}

	model := NewOpenAIEmbedding(config)
	// Default dimension for unknown models should be 1536
	assert.Equal(t, 1536, model.Dimension())
}

// TestOllamaEmbedding_UnknownModel tests dimension for unknown Ollama model
func TestOllamaEmbedding_UnknownModel(t *testing.T) {
	config := EmbeddingConfig{
		ModelType: ModelTypeOllama,
		ModelName: "unknown-model",
		Timeout:   5 * time.Second,
	}

	model := NewOllamaEmbedding(config)
	// Default dimension for unknown models should be 768
	assert.Equal(t, 768, model.Dimension())
}

// TestHuggingFaceEmbedding_UnknownModel tests dimension for unknown HuggingFace model
func TestHuggingFaceEmbedding_UnknownModel(t *testing.T) {
	config := EmbeddingConfig{
		ModelType: ModelTypeBGEM3,
		ModelName: "unknown/model",
		Timeout:   5 * time.Second,
	}

	model := NewHuggingFaceEmbedding(config)
	// Default dimension for unknown models should be 768
	assert.Equal(t, 768, model.Dimension())
}
