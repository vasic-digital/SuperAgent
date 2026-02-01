// Package embedding provides embedding model implementations.
// Implements multiple embedding models including OpenAI, Ollama, BGE-M3, Nomic, and CodeBERT.
package embedding

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// EmbeddingModel defines the interface for embedding models.
type EmbeddingModel interface {
	// Name returns the model name
	Name() string
	// Dimension returns the embedding dimension
	Dimension() int
	// Embed generates an embedding for the given text
	Embed(ctx context.Context, text string) ([]float64, error)
	// EmbedBatch generates embeddings for multiple texts
	EmbedBatch(ctx context.Context, texts []string) ([][]float64, error)
	// Close closes the model connection
	Close() error
}

// ModelType represents the type of embedding model.
type ModelType string

const (
	ModelTypeOpenAI   ModelType = "openai"
	ModelTypeOllama   ModelType = "ollama"
	ModelTypeBGEM3    ModelType = "bge-m3"
	ModelTypeNomic    ModelType = "nomic"
	ModelTypeCodeBERT ModelType = "codebert"
	ModelTypeQwen3    ModelType = "qwen3"
	ModelTypeGTE      ModelType = "gte"
	ModelTypeE5       ModelType = "e5"
)

// EmbeddingConfig configures embedding models.
type EmbeddingConfig struct {
	ModelType    ModelType     `json:"model_type"`
	ModelName    string        `json:"model_name"`
	APIKey       string        `json:"api_key,omitempty"`
	BaseURL      string        `json:"base_url,omitempty"`
	Timeout      time.Duration `json:"timeout"`
	MaxBatchSize int           `json:"max_batch_size"`
	CacheEnabled bool          `json:"cache_enabled"`
	CacheSize    int           `json:"cache_size"`
}

// DefaultEmbeddingConfig returns default configuration.
func DefaultEmbeddingConfig(modelType ModelType) EmbeddingConfig {
	config := EmbeddingConfig{
		ModelType:    modelType,
		Timeout:      30 * time.Second,
		MaxBatchSize: 100,
		CacheEnabled: true,
		CacheSize:    10000,
	}

	switch modelType {
	case ModelTypeOpenAI:
		config.ModelName = "text-embedding-3-small"
		config.BaseURL = "https://api.openai.com/v1"
	case ModelTypeOllama:
		config.ModelName = "nomic-embed-text"
		config.BaseURL = "http://localhost:11434"
	case ModelTypeBGEM3:
		config.ModelName = "BAAI/bge-m3"
		config.BaseURL = "https://api-inference.huggingface.co/models"
	case ModelTypeNomic:
		config.ModelName = "nomic-ai/nomic-embed-text-v1.5"
		config.BaseURL = "https://api-inference.huggingface.co/models"
	case ModelTypeCodeBERT:
		config.ModelName = "microsoft/codebert-base"
		config.BaseURL = "https://api-inference.huggingface.co/models"
	case ModelTypeQwen3:
		config.ModelName = "Qwen/Qwen3-Embedding-0.6B"
		config.BaseURL = "https://api-inference.huggingface.co/models"
	case ModelTypeGTE:
		config.ModelName = "thenlper/gte-large"
		config.BaseURL = "https://api-inference.huggingface.co/models"
	case ModelTypeE5:
		config.ModelName = "intfloat/e5-large-v2"
		config.BaseURL = "https://api-inference.huggingface.co/models"
	}

	return config
}

// OpenAIEmbedding implements OpenAI embedding models.
type OpenAIEmbedding struct {
	config     EmbeddingConfig
	httpClient *http.Client
	dimension  int
	cache      *EmbeddingCache
}

// NewOpenAIEmbedding creates a new OpenAI embedding model.
func NewOpenAIEmbedding(config EmbeddingConfig) *OpenAIEmbedding {
	dimension := 1536
	switch config.ModelName {
	case "text-embedding-3-small":
		dimension = 1536
	case "text-embedding-3-large":
		dimension = 3072
	case "text-embedding-ada-002":
		dimension = 1536
	}

	model := &OpenAIEmbedding{
		config:    config,
		dimension: dimension,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}

	if config.CacheEnabled {
		model.cache = NewEmbeddingCache(config.CacheSize)
	}

	return model
}

// Name returns the model name.
func (m *OpenAIEmbedding) Name() string {
	return fmt.Sprintf("openai/%s", m.config.ModelName)
}

// Dimension returns the embedding dimension.
func (m *OpenAIEmbedding) Dimension() int {
	return m.dimension
}

// Embed generates an embedding for the given text.
func (m *OpenAIEmbedding) Embed(ctx context.Context, text string) ([]float64, error) {
	if m.cache != nil {
		if cached, ok := m.cache.Get(text); ok {
			return cached, nil
		}
	}

	embeddings, err := m.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}

	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}

	if m.cache != nil {
		m.cache.Set(text, embeddings[0])
	}

	return embeddings[0], nil
}

// EmbedBatch generates embeddings for multiple texts.
func (m *OpenAIEmbedding) EmbedBatch(ctx context.Context, texts []string) ([][]float64, error) {
	reqBody := map[string]interface{}{
		"input": texts,
		"model": m.config.ModelName,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/embeddings", m.config.BaseURL),
		strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.config.APIKey))

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API error: %s - %s", resp.Status, string(respBody))
	}

	var result struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	embeddings := make([][]float64, len(result.Data))
	for i, item := range result.Data {
		embeddings[i] = item.Embedding
	}

	return embeddings, nil
}

// Close closes the model connection.
func (m *OpenAIEmbedding) Close() error {
	return nil
}

// OllamaEmbedding implements Ollama embedding models.
type OllamaEmbedding struct {
	config     EmbeddingConfig
	httpClient *http.Client
	dimension  int
	cache      *EmbeddingCache
}

// NewOllamaEmbedding creates a new Ollama embedding model.
func NewOllamaEmbedding(config EmbeddingConfig) *OllamaEmbedding {
	dimension := 768
	switch config.ModelName {
	case "nomic-embed-text":
		dimension = 768
	case "mxbai-embed-large":
		dimension = 1024
	case "all-minilm":
		dimension = 384
	}

	model := &OllamaEmbedding{
		config:    config,
		dimension: dimension,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}

	if config.CacheEnabled {
		model.cache = NewEmbeddingCache(config.CacheSize)
	}

	return model
}

// Name returns the model name.
func (m *OllamaEmbedding) Name() string {
	return fmt.Sprintf("ollama/%s", m.config.ModelName)
}

// Dimension returns the embedding dimension.
func (m *OllamaEmbedding) Dimension() int {
	return m.dimension
}

// Embed generates an embedding for the given text.
func (m *OllamaEmbedding) Embed(ctx context.Context, text string) ([]float64, error) {
	if m.cache != nil {
		if cached, ok := m.cache.Get(text); ok {
			return cached, nil
		}
	}

	reqBody := map[string]interface{}{
		"model":  m.config.ModelName,
		"prompt": text,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/api/embeddings", m.config.BaseURL),
		strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama API error: %s - %s", resp.Status, string(respBody))
	}

	var result struct {
		Embedding []float64 `json:"embedding"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if m.cache != nil {
		m.cache.Set(text, result.Embedding)
	}

	return result.Embedding, nil
}

// EmbedBatch generates embeddings for multiple texts.
func (m *OllamaEmbedding) EmbedBatch(ctx context.Context, texts []string) ([][]float64, error) {
	embeddings := make([][]float64, len(texts))
	for i, text := range texts {
		emb, err := m.Embed(ctx, text)
		if err != nil {
			return nil, err
		}
		embeddings[i] = emb
	}
	return embeddings, nil
}

// Close closes the model connection.
func (m *OllamaEmbedding) Close() error {
	return nil
}

// HuggingFaceEmbedding implements HuggingFace embedding models.
type HuggingFaceEmbedding struct {
	config     EmbeddingConfig
	httpClient *http.Client
	dimension  int
	cache      *EmbeddingCache
}

// NewHuggingFaceEmbedding creates a new HuggingFace embedding model.
func NewHuggingFaceEmbedding(config EmbeddingConfig) *HuggingFaceEmbedding {
	dimension := 768
	switch {
	case strings.Contains(config.ModelName, "bge-m3"):
		dimension = 1024
	case strings.Contains(config.ModelName, "nomic-embed"):
		dimension = 768
	case strings.Contains(config.ModelName, "codebert"):
		dimension = 768
	case strings.Contains(config.ModelName, "gte-large"):
		dimension = 1024
	case strings.Contains(config.ModelName, "e5-large"):
		dimension = 1024
	case strings.Contains(config.ModelName, "Qwen3"):
		dimension = 768
	}

	model := &HuggingFaceEmbedding{
		config:    config,
		dimension: dimension,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}

	if config.CacheEnabled {
		model.cache = NewEmbeddingCache(config.CacheSize)
	}

	return model
}

// Name returns the model name.
func (m *HuggingFaceEmbedding) Name() string {
	return fmt.Sprintf("huggingface/%s", m.config.ModelName)
}

// Dimension returns the embedding dimension.
func (m *HuggingFaceEmbedding) Dimension() int {
	return m.dimension
}

// Embed generates an embedding for the given text.
func (m *HuggingFaceEmbedding) Embed(ctx context.Context, text string) ([]float64, error) {
	if m.cache != nil {
		if cached, ok := m.cache.Get(text); ok {
			return cached, nil
		}
	}

	reqBody := map[string]interface{}{
		"inputs": text,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/%s", m.config.BaseURL, m.config.ModelName),
		strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if m.config.APIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.config.APIKey))
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HuggingFace API error: %s - %s", resp.Status, string(respBody))
	}

	var embedding []float64
	if err := json.NewDecoder(resp.Body).Decode(&embedding); err != nil {
		// Try nested format
		var nested [][]float64
		respBody, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(respBody, &nested); err == nil && len(nested) > 0 {
			embedding = nested[0]
		} else {
			return nil, err
		}
	}

	if m.cache != nil {
		m.cache.Set(text, embedding)
	}

	return embedding, nil
}

// EmbedBatch generates embeddings for multiple texts.
func (m *HuggingFaceEmbedding) EmbedBatch(ctx context.Context, texts []string) ([][]float64, error) {
	reqBody := map[string]interface{}{
		"inputs": texts,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/%s", m.config.BaseURL, m.config.ModelName),
		strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if m.config.APIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.config.APIKey))
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HuggingFace API error: %s - %s", resp.Status, string(respBody))
	}

	var embeddings [][]float64
	if err := json.NewDecoder(resp.Body).Decode(&embeddings); err != nil {
		return nil, err
	}

	return embeddings, nil
}

// Close closes the model connection.
func (m *HuggingFaceEmbedding) Close() error {
	return nil
}

// EmbeddingCache caches embeddings.
type EmbeddingCache struct {
	cache   map[string][]float64
	maxSize int
	mu      sync.RWMutex
}

// NewEmbeddingCache creates a new embedding cache.
func NewEmbeddingCache(maxSize int) *EmbeddingCache {
	return &EmbeddingCache{
		cache:   make(map[string][]float64),
		maxSize: maxSize,
	}
}

// Get retrieves an embedding from the cache.
func (c *EmbeddingCache) Get(key string) ([]float64, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	emb, ok := c.cache[key]
	return emb, ok
}

// Set stores an embedding in the cache.
func (c *EmbeddingCache) Set(key string, embedding []float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.cache) >= c.maxSize {
		// Evict oldest entry (simple approach)
		for k := range c.cache {
			delete(c.cache, k)
			break
		}
	}

	c.cache[key] = embedding
}

// Size returns the current cache size.
func (c *EmbeddingCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cache)
}

// Clear clears the cache.
func (c *EmbeddingCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[string][]float64)
}

// EmbeddingModelRegistry manages embedding models.
type EmbeddingModelRegistry struct {
	models map[string]EmbeddingModel
	mu     sync.RWMutex
}

// NewEmbeddingModelRegistry creates a new registry.
func NewEmbeddingModelRegistry() *EmbeddingModelRegistry {
	return &EmbeddingModelRegistry{
		models: make(map[string]EmbeddingModel),
	}
}

// Register registers a model.
func (r *EmbeddingModelRegistry) Register(name string, model EmbeddingModel) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.models[name] = model
}

// Get retrieves a model.
func (r *EmbeddingModelRegistry) Get(name string) (EmbeddingModel, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	model, ok := r.models[name]
	return model, ok
}

// List returns all registered model names.
func (r *EmbeddingModelRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.models))
	for name := range r.models {
		names = append(names, name)
	}
	return names
}

// Close closes all models.
func (r *EmbeddingModelRegistry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, model := range r.models {
		if err := model.Close(); err != nil {
			return err
		}
	}
	return nil
}

// CreateModel creates an embedding model from config.
func CreateModel(config EmbeddingConfig) (EmbeddingModel, error) {
	switch config.ModelType {
	case ModelTypeOpenAI:
		return NewOpenAIEmbedding(config), nil
	case ModelTypeOllama:
		return NewOllamaEmbedding(config), nil
	case ModelTypeBGEM3, ModelTypeNomic, ModelTypeCodeBERT, ModelTypeQwen3, ModelTypeGTE, ModelTypeE5:
		return NewHuggingFaceEmbedding(config), nil
	default:
		return nil, fmt.Errorf("unknown model type: %s", config.ModelType)
	}
}

// AvailableModels lists all available embedding models.
var AvailableModels = []struct {
	Type        ModelType
	Name        string
	Dimension   int
	Description string
}{
	{ModelTypeOpenAI, "text-embedding-3-small", 1536, "OpenAI's compact embedding model"},
	{ModelTypeOpenAI, "text-embedding-3-large", 3072, "OpenAI's large embedding model"},
	{ModelTypeOpenAI, "text-embedding-ada-002", 1536, "OpenAI's Ada embedding model"},
	{ModelTypeOllama, "nomic-embed-text", 768, "Nomic's text embedding via Ollama"},
	{ModelTypeOllama, "mxbai-embed-large", 1024, "MixedBread AI large embedding via Ollama"},
	{ModelTypeOllama, "all-minilm", 384, "MiniLM embedding via Ollama"},
	{ModelTypeBGEM3, "BAAI/bge-m3", 1024, "BGE-M3 multilingual embedding"},
	{ModelTypeNomic, "nomic-ai/nomic-embed-text-v1.5", 768, "Nomic Embed Text v1.5"},
	{ModelTypeCodeBERT, "microsoft/codebert-base", 768, "Microsoft CodeBERT for code"},
	{ModelTypeQwen3, "Qwen/Qwen3-Embedding-0.6B", 768, "Qwen3 Embedding model"},
	{ModelTypeGTE, "thenlper/gte-large", 1024, "GTE Large embedding"},
	{ModelTypeE5, "intfloat/e5-large-v2", 1024, "E5 Large v2 embedding"},
}

// GetModelInfo returns information about available models.
func GetModelInfo() []map[string]interface{} {
	var info []map[string]interface{}
	for _, model := range AvailableModels {
		info = append(info, map[string]interface{}{
			"type":        string(model.Type),
			"name":        model.Name,
			"dimension":   model.Dimension,
			"description": model.Description,
		})
	}
	return info
}
