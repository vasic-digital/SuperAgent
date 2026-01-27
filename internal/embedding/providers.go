// Package embedding provides additional embedding model implementations.
// Implements Cohere, Voyage AI, Jina, Google, and AWS Bedrock embedding providers.
package embedding

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Additional ModelType constants for new providers
const (
	ModelTypeCohere  ModelType = "cohere"
	ModelTypeVoyage  ModelType = "voyage"
	ModelTypeJina    ModelType = "jina"
	ModelTypeGoogle  ModelType = "google"
	ModelTypeBedrock ModelType = "bedrock"
)

// =============================================================================
// Cohere Embedding Model
// =============================================================================

// CohereEmbedding implements Cohere embedding models.
type CohereEmbedding struct {
	config     EmbeddingConfig
	httpClient *http.Client
	dimension  int
	cache      *EmbeddingCache
}

// CohereEmbedRequest represents a Cohere embed API request.
type CohereEmbedRequest struct {
	Texts          []string `json:"texts"`
	Model          string   `json:"model"`
	InputType      string   `json:"input_type"`
	EmbeddingTypes []string `json:"embedding_types,omitempty"`
	Truncate       string   `json:"truncate,omitempty"`
}

// CohereEmbedResponse represents a Cohere embed API response.
type CohereEmbedResponse struct {
	ID            string                 `json:"id"`
	Embeddings    [][]float64            `json:"embeddings"`
	EmbeddingsObj *CohereEmbeddingsObj   `json:"embeddings_by_type,omitempty"`
	Texts         []string               `json:"texts"`
	Meta          map[string]interface{} `json:"meta"`
	ResponseType  string                 `json:"response_type,omitempty"`
}

// CohereEmbeddingsObj represents typed embeddings.
type CohereEmbeddingsObj struct {
	Float   [][]float64 `json:"float,omitempty"`
	Int8    [][]int8    `json:"int8,omitempty"`
	Uint8   [][]uint8   `json:"uint8,omitempty"`
	Binary  [][]int     `json:"binary,omitempty"`
	Ubinary [][]int     `json:"ubinary,omitempty"`
}

// NewCohereEmbedding creates a new Cohere embedding model.
func NewCohereEmbedding(config EmbeddingConfig) *CohereEmbedding {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.cohere.com/v2"
	}

	// Set dimensions based on model
	dimension := 1024
	switch config.ModelName {
	case "embed-english-v3.0", "embed-multilingual-v3.0":
		dimension = 1024
	case "embed-english-light-v3.0", "embed-multilingual-light-v3.0":
		dimension = 384
	case "embed-english-v2.0":
		dimension = 4096
	case "embed-multilingual-v2.0":
		dimension = 768
	}

	model := &CohereEmbedding{
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
func (m *CohereEmbedding) Name() string {
	return fmt.Sprintf("cohere/%s", m.config.ModelName)
}

// Dimension returns the embedding dimension.
func (m *CohereEmbedding) Dimension() int {
	return m.dimension
}

// Embed generates an embedding for the given text.
func (m *CohereEmbedding) Embed(ctx context.Context, text string) ([]float64, error) {
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
func (m *CohereEmbedding) EmbedBatch(ctx context.Context, texts []string) ([][]float64, error) {
	reqBody := CohereEmbedRequest{
		Texts:     texts,
		Model:     m.config.ModelName,
		InputType: "search_document",
		Truncate:  "END",
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/embed", m.config.BaseURL),
		bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.config.APIKey))
	req.Header.Set("X-Client-Name", "helix-agent")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Cohere API error: %s - %s", resp.Status, string(respBody))
	}

	var result CohereEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Handle different response formats
	if result.Embeddings != nil {
		return result.Embeddings, nil
	}
	if result.EmbeddingsObj != nil && result.EmbeddingsObj.Float != nil {
		return result.EmbeddingsObj.Float, nil
	}

	return nil, fmt.Errorf("no embeddings in response")
}

// Close closes the model connection.
func (m *CohereEmbedding) Close() error {
	return nil
}

// =============================================================================
// Voyage AI Embedding Model
// =============================================================================

// VoyageEmbedding implements Voyage AI embedding models.
type VoyageEmbedding struct {
	config     EmbeddingConfig
	httpClient *http.Client
	dimension  int
	cache      *EmbeddingCache
}

// VoyageEmbedRequest represents a Voyage embed API request.
type VoyageEmbedRequest struct {
	Input           []string `json:"input"`
	Model           string   `json:"model"`
	InputType       string   `json:"input_type,omitempty"`
	Truncation      bool     `json:"truncation,omitempty"`
	OutputDimension int      `json:"output_dimension,omitempty"`
}

// VoyageEmbedResponse represents a Voyage embed API response.
type VoyageEmbedResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

// NewVoyageEmbedding creates a new Voyage AI embedding model.
func NewVoyageEmbedding(config EmbeddingConfig) *VoyageEmbedding {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.voyageai.com/v1"
	}

	// Set dimensions based on model
	dimension := 1024
	switch config.ModelName {
	case "voyage-3":
		dimension = 1024
	case "voyage-3-lite":
		dimension = 512
	case "voyage-code-3":
		dimension = 1024
	case "voyage-finance-2":
		dimension = 1024
	case "voyage-law-2":
		dimension = 1024
	case "voyage-large-2", "voyage-large-2-instruct":
		dimension = 1536
	case "voyage-2":
		dimension = 1024
	}

	model := &VoyageEmbedding{
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
func (m *VoyageEmbedding) Name() string {
	return fmt.Sprintf("voyage/%s", m.config.ModelName)
}

// Dimension returns the embedding dimension.
func (m *VoyageEmbedding) Dimension() int {
	return m.dimension
}

// Embed generates an embedding for the given text.
func (m *VoyageEmbedding) Embed(ctx context.Context, text string) ([]float64, error) {
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
func (m *VoyageEmbedding) EmbedBatch(ctx context.Context, texts []string) ([][]float64, error) {
	reqBody := VoyageEmbedRequest{
		Input:      texts,
		Model:      m.config.ModelName,
		InputType:  "document",
		Truncation: true,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/embeddings", m.config.BaseURL),
		bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.config.APIKey))

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Voyage API error: %s - %s", resp.Status, string(respBody))
	}

	var result VoyageEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	embeddings := make([][]float64, len(result.Data))
	for _, item := range result.Data {
		embeddings[item.Index] = item.Embedding
	}

	return embeddings, nil
}

// Close closes the model connection.
func (m *VoyageEmbedding) Close() error {
	return nil
}

// =============================================================================
// Jina AI Embedding Model
// =============================================================================

// JinaEmbedding implements Jina AI embedding models.
type JinaEmbedding struct {
	config     EmbeddingConfig
	httpClient *http.Client
	dimension  int
	cache      *EmbeddingCache
}

// JinaEmbedRequest represents a Jina embed API request.
type JinaEmbedRequest struct {
	Input          []string `json:"input"`
	Model          string   `json:"model"`
	EncodingFormat string   `json:"encoding_format,omitempty"`
	Task           string   `json:"task,omitempty"`
	Dimensions     int      `json:"dimensions,omitempty"`
	Late_chunking  bool     `json:"late_chunking,omitempty"`
}

// JinaEmbedResponse represents a Jina embed API response.
type JinaEmbedResponse struct {
	Model  string `json:"model"`
	Object string `json:"object"`
	Usage  struct {
		TotalTokens  int `json:"total_tokens"`
		PromptTokens int `json:"prompt_tokens"`
	} `json:"usage"`
	Data []struct {
		Object    string    `json:"object"`
		Index     int       `json:"index"`
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
}

// NewJinaEmbedding creates a new Jina AI embedding model.
func NewJinaEmbedding(config EmbeddingConfig) *JinaEmbedding {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.jina.ai/v1"
	}

	// Set dimensions based on model
	dimension := 1024
	switch config.ModelName {
	case "jina-embeddings-v3":
		dimension = 1024
	case "jina-embeddings-v2-base-en":
		dimension = 768
	case "jina-embeddings-v2-small-en":
		dimension = 512
	case "jina-embeddings-v2-base-de":
		dimension = 768
	case "jina-embeddings-v2-base-es":
		dimension = 768
	case "jina-embeddings-v2-base-zh":
		dimension = 768
	case "jina-clip-v1":
		dimension = 768
	case "jina-colbert-v2":
		dimension = 128
	case "jina-reranker-v2-base-multilingual":
		dimension = 768
	}

	model := &JinaEmbedding{
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
func (m *JinaEmbedding) Name() string {
	return fmt.Sprintf("jina/%s", m.config.ModelName)
}

// Dimension returns the embedding dimension.
func (m *JinaEmbedding) Dimension() int {
	return m.dimension
}

// Embed generates an embedding for the given text.
func (m *JinaEmbedding) Embed(ctx context.Context, text string) ([]float64, error) {
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
func (m *JinaEmbedding) EmbedBatch(ctx context.Context, texts []string) ([][]float64, error) {
	reqBody := JinaEmbedRequest{
		Input:          texts,
		Model:          m.config.ModelName,
		EncodingFormat: "float",
		Task:           "retrieval.document",
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/embeddings", m.config.BaseURL),
		bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.config.APIKey))

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Jina API error: %s - %s", resp.Status, string(respBody))
	}

	var result JinaEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	embeddings := make([][]float64, len(result.Data))
	for _, item := range result.Data {
		embeddings[item.Index] = item.Embedding
	}

	return embeddings, nil
}

// Close closes the model connection.
func (m *JinaEmbedding) Close() error {
	return nil
}

// =============================================================================
// Google Vertex AI Embedding Model
// =============================================================================

// GoogleEmbedding implements Google Vertex AI embedding models.
type GoogleEmbedding struct {
	config     EmbeddingConfig
	httpClient *http.Client
	dimension  int
	cache      *EmbeddingCache
	projectID  string
	location   string
}

// GoogleEmbedRequest represents a Google embedding API request.
type GoogleEmbedRequest struct {
	Instances []GoogleEmbedInstance `json:"instances"`
}

// GoogleEmbedInstance represents a single embedding input.
type GoogleEmbedInstance struct {
	Content  string `json:"content"`
	TaskType string `json:"task_type,omitempty"`
}

// GoogleEmbedResponse represents a Google embedding API response.
type GoogleEmbedResponse struct {
	Predictions []struct {
		Embeddings struct {
			Values     []float64 `json:"values"`
			Statistics struct {
				TokenCount int `json:"token_count"`
			} `json:"statistics"`
		} `json:"embeddings"`
	} `json:"predictions"`
	DeployedModelID string `json:"deployedModelId,omitempty"`
}

// GoogleEmbeddingConfig extends EmbeddingConfig for Google-specific settings.
type GoogleEmbeddingConfig struct {
	EmbeddingConfig
	ProjectID string `json:"project_id"`
	Location  string `json:"location"`
}

// NewGoogleEmbedding creates a new Google Vertex AI embedding model.
func NewGoogleEmbedding(config EmbeddingConfig, projectID, location string) *GoogleEmbedding {
	if location == "" {
		location = "us-central1"
	}
	if config.BaseURL == "" {
		config.BaseURL = fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1", location)
	}

	// Set dimensions based on model
	dimension := 768
	switch config.ModelName {
	case "text-embedding-005", "textembedding-gecko@003":
		dimension = 768
	case "text-multilingual-embedding-002":
		dimension = 768
	case "text-embedding-004":
		dimension = 768
	case "textembedding-gecko-multilingual@001":
		dimension = 768
	}

	model := &GoogleEmbedding{
		config:    config,
		dimension: dimension,
		projectID: projectID,
		location:  location,
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
func (m *GoogleEmbedding) Name() string {
	return fmt.Sprintf("google/%s", m.config.ModelName)
}

// Dimension returns the embedding dimension.
func (m *GoogleEmbedding) Dimension() int {
	return m.dimension
}

// Embed generates an embedding for the given text.
func (m *GoogleEmbedding) Embed(ctx context.Context, text string) ([]float64, error) {
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
func (m *GoogleEmbedding) EmbedBatch(ctx context.Context, texts []string) ([][]float64, error) {
	instances := make([]GoogleEmbedInstance, len(texts))
	for i, text := range texts {
		instances[i] = GoogleEmbedInstance{
			Content:  text,
			TaskType: "RETRIEVAL_DOCUMENT",
		}
	}

	reqBody := GoogleEmbedRequest{
		Instances: instances,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/projects/%s/locations/%s/publishers/google/models/%s:predict",
		m.config.BaseURL, m.projectID, m.location, m.config.ModelName)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.config.APIKey))

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Google API error: %s - %s", resp.Status, string(respBody))
	}

	var result GoogleEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	embeddings := make([][]float64, len(result.Predictions))
	for i, pred := range result.Predictions {
		embeddings[i] = pred.Embeddings.Values
	}

	return embeddings, nil
}

// Close closes the model connection.
func (m *GoogleEmbedding) Close() error {
	return nil
}

// =============================================================================
// AWS Bedrock Embedding Model
// =============================================================================

// BedrockEmbedding implements AWS Bedrock embedding models.
type BedrockEmbedding struct {
	config      EmbeddingConfig
	httpClient  *http.Client
	dimension   int
	cache       *EmbeddingCache
	region      string
	accessKeyID string
	secretKey   string
}

// BedrockTitanRequest represents an AWS Titan embedding request.
type BedrockTitanRequest struct {
	InputText string `json:"inputText"`
}

// BedrockTitanResponse represents an AWS Titan embedding response.
type BedrockTitanResponse struct {
	Embedding      []float64 `json:"embedding"`
	InputTextToken int       `json:"inputTextTokenCount"`
}

// BedrockCohereRequest represents a Bedrock Cohere embedding request.
type BedrockCohereRequest struct {
	Texts     []string `json:"texts"`
	InputType string   `json:"input_type"`
}

// BedrockCohereResponse represents a Bedrock Cohere embedding response.
type BedrockCohereResponse struct {
	Embeddings [][]float64 `json:"embeddings"`
}

// NewBedrockEmbedding creates a new AWS Bedrock embedding model.
func NewBedrockEmbedding(config EmbeddingConfig, region, accessKeyID, secretKey string) *BedrockEmbedding {
	if region == "" {
		region = "us-east-1"
	}
	if config.BaseURL == "" {
		config.BaseURL = fmt.Sprintf("https://bedrock-runtime.%s.amazonaws.com", region)
	}

	// Set dimensions based on model
	dimension := 1536
	switch config.ModelName {
	case "amazon.titan-embed-text-v1":
		dimension = 1536
	case "amazon.titan-embed-text-v2:0":
		dimension = 1024
	case "amazon.titan-embed-image-v1":
		dimension = 1024
	case "cohere.embed-english-v3":
		dimension = 1024
	case "cohere.embed-multilingual-v3":
		dimension = 1024
	}

	model := &BedrockEmbedding{
		config:      config,
		dimension:   dimension,
		region:      region,
		accessKeyID: accessKeyID,
		secretKey:   secretKey,
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
func (m *BedrockEmbedding) Name() string {
	return fmt.Sprintf("bedrock/%s", m.config.ModelName)
}

// Dimension returns the embedding dimension.
func (m *BedrockEmbedding) Dimension() int {
	return m.dimension
}

// Embed generates an embedding for the given text.
func (m *BedrockEmbedding) Embed(ctx context.Context, text string) ([]float64, error) {
	if m.cache != nil {
		if cached, ok := m.cache.Get(text); ok {
			return cached, nil
		}
	}

	var embedding []float64
	var err error

	if strings.HasPrefix(m.config.ModelName, "amazon.titan") {
		embedding, err = m.embedTitan(ctx, text)
	} else if strings.HasPrefix(m.config.ModelName, "cohere.") {
		embeddings, err := m.embedCohere(ctx, []string{text})
		if err != nil {
			return nil, err
		}
		if len(embeddings) > 0 {
			embedding = embeddings[0]
		}
	} else {
		return nil, fmt.Errorf("unsupported model: %s", m.config.ModelName)
	}

	if err != nil {
		return nil, err
	}

	if m.cache != nil {
		m.cache.Set(text, embedding)
	}

	return embedding, nil
}

// EmbedBatch generates embeddings for multiple texts.
func (m *BedrockEmbedding) EmbedBatch(ctx context.Context, texts []string) ([][]float64, error) {
	if strings.HasPrefix(m.config.ModelName, "cohere.") {
		return m.embedCohere(ctx, texts)
	}

	// Titan models don't support batch, so we call individually
	embeddings := make([][]float64, len(texts))
	for i, text := range texts {
		emb, err := m.Embed(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("failed to embed text %d: %w", i, err)
		}
		embeddings[i] = emb
	}
	return embeddings, nil
}

// embedTitan generates an embedding using Titan model.
func (m *BedrockEmbedding) embedTitan(ctx context.Context, text string) ([]float64, error) {
	reqBody := BedrockTitanRequest{
		InputText: text,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/model/%s/invoke", m.config.BaseURL, m.config.ModelName)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Sign the request with AWS SigV4
	if err := m.signRequest(req, body); err != nil {
		return nil, fmt.Errorf("failed to sign request: %w", err)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Bedrock API error: %s - %s", resp.Status, string(respBody))
	}

	var result BedrockTitanResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result.Embedding, nil
}

// embedCohere generates embeddings using Cohere model on Bedrock.
func (m *BedrockEmbedding) embedCohere(ctx context.Context, texts []string) ([][]float64, error) {
	reqBody := BedrockCohereRequest{
		Texts:     texts,
		InputType: "search_document",
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/model/%s/invoke", m.config.BaseURL, m.config.ModelName)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Sign the request with AWS SigV4
	if err := m.signRequest(req, body); err != nil {
		return nil, fmt.Errorf("failed to sign request: %w", err)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Bedrock API error: %s - %s", resp.Status, string(respBody))
	}

	var result BedrockCohereResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result.Embeddings, nil
}

// signRequest signs an HTTP request with AWS SigV4.
func (m *BedrockEmbedding) signRequest(req *http.Request, body []byte) error {
	// Get current time
	t := time.Now().UTC()
	amzDate := t.Format("20060102T150405Z")
	dateStamp := t.Format("20060102")

	// Create canonical request
	hashedPayload := sha256Hash(body)

	canonicalHeaders := fmt.Sprintf("content-type:%s\nhost:%s\nx-amz-date:%s\n",
		req.Header.Get("Content-Type"), req.URL.Host, amzDate)
	signedHeaders := "content-type;host;x-amz-date"

	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		req.Method,
		req.URL.Path,
		req.URL.RawQuery,
		canonicalHeaders,
		signedHeaders,
		hashedPayload)

	// Create string to sign
	credentialScope := fmt.Sprintf("%s/%s/bedrock/aws4_request", dateStamp, m.region)
	stringToSign := fmt.Sprintf("AWS4-HMAC-SHA256\n%s\n%s\n%s",
		amzDate, credentialScope, sha256Hash([]byte(canonicalRequest)))

	// Calculate signature
	kDate := hmacSHA256([]byte("AWS4"+m.secretKey), dateStamp)
	kRegion := hmacSHA256(kDate, m.region)
	kService := hmacSHA256(kRegion, "bedrock")
	kSigning := hmacSHA256(kService, "aws4_request")
	signature := hex.EncodeToString(hmacSHA256(kSigning, stringToSign))

	// Create authorization header
	authHeader := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		m.accessKeyID, credentialScope, signedHeaders, signature)

	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("Authorization", authHeader)

	return nil
}

// Close closes the model connection.
func (m *BedrockEmbedding) Close() error {
	return nil
}

// =============================================================================
// Helper Functions
// =============================================================================

// sha256Hash computes SHA256 hash of data.
func sha256Hash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// hmacSHA256 computes HMAC-SHA256.
func hmacSHA256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
}

// =============================================================================
// Extended Factory Functions
// =============================================================================

// init registers the new model types in DefaultEmbeddingConfig.
func init() {
	// This ensures the new model types are recognized
}

// DefaultEmbeddingConfigExtended returns default configuration for extended models.
func DefaultEmbeddingConfigExtended(modelType ModelType) EmbeddingConfig {
	config := EmbeddingConfig{
		ModelType:    modelType,
		Timeout:      30 * time.Second,
		MaxBatchSize: 100,
		CacheEnabled: true,
		CacheSize:    10000,
	}

	switch modelType {
	case ModelTypeCohere:
		config.ModelName = "embed-english-v3.0"
		config.BaseURL = "https://api.cohere.com/v2"
	case ModelTypeVoyage:
		config.ModelName = "voyage-3"
		config.BaseURL = "https://api.voyageai.com/v1"
	case ModelTypeJina:
		config.ModelName = "jina-embeddings-v3"
		config.BaseURL = "https://api.jina.ai/v1"
	case ModelTypeGoogle:
		config.ModelName = "text-embedding-005"
		// BaseURL set dynamically based on location
	case ModelTypeBedrock:
		config.ModelName = "amazon.titan-embed-text-v2:0"
		// BaseURL set dynamically based on region
	default:
		// Fall back to the original DefaultEmbeddingConfig
		return DefaultEmbeddingConfig(modelType)
	}

	return config
}

// CreateModelExtended creates an embedding model from config (extended version).
func CreateModelExtended(config EmbeddingConfig) (EmbeddingModel, error) {
	switch config.ModelType {
	case ModelTypeCohere:
		return NewCohereEmbedding(config), nil
	case ModelTypeVoyage:
		return NewVoyageEmbedding(config), nil
	case ModelTypeJina:
		return NewJinaEmbedding(config), nil
	case ModelTypeGoogle:
		return NewGoogleEmbedding(config, "", ""), nil
	case ModelTypeBedrock:
		return NewBedrockEmbedding(config, "", "", ""), nil
	default:
		return CreateModel(config)
	}
}

// =============================================================================
// Extended AvailableModels
// =============================================================================

// AvailableModelsExtended lists all available embedding models including new providers.
var AvailableModelsExtended = []struct {
	Type        ModelType
	Name        string
	Dimension   int
	Description string
}{
	// Cohere models
	{ModelTypeCohere, "embed-english-v3.0", 1024, "Cohere Embed v3 English"},
	{ModelTypeCohere, "embed-multilingual-v3.0", 1024, "Cohere Embed v3 Multilingual"},
	{ModelTypeCohere, "embed-english-light-v3.0", 384, "Cohere Embed v3 English Light"},
	{ModelTypeCohere, "embed-multilingual-light-v3.0", 384, "Cohere Embed v3 Multilingual Light"},

	// Voyage models
	{ModelTypeVoyage, "voyage-3", 1024, "Voyage AI v3 general embedding"},
	{ModelTypeVoyage, "voyage-3-lite", 512, "Voyage AI v3 lite embedding"},
	{ModelTypeVoyage, "voyage-code-3", 1024, "Voyage AI v3 code embedding"},
	{ModelTypeVoyage, "voyage-finance-2", 1024, "Voyage AI finance embedding"},
	{ModelTypeVoyage, "voyage-law-2", 1024, "Voyage AI law embedding"},
	{ModelTypeVoyage, "voyage-large-2", 1536, "Voyage AI large embedding"},

	// Jina models
	{ModelTypeJina, "jina-embeddings-v3", 1024, "Jina Embeddings v3"},
	{ModelTypeJina, "jina-embeddings-v2-base-en", 768, "Jina Embeddings v2 Base English"},
	{ModelTypeJina, "jina-embeddings-v2-small-en", 512, "Jina Embeddings v2 Small English"},
	{ModelTypeJina, "jina-clip-v1", 768, "Jina CLIP v1 multimodal"},
	{ModelTypeJina, "jina-colbert-v2", 128, "Jina ColBERT v2"},

	// Google models
	{ModelTypeGoogle, "text-embedding-005", 768, "Google Text Embedding 005"},
	{ModelTypeGoogle, "text-multilingual-embedding-002", 768, "Google Multilingual Embedding"},
	{ModelTypeGoogle, "textembedding-gecko@003", 768, "Google Gecko Embedding"},

	// AWS Bedrock models
	{ModelTypeBedrock, "amazon.titan-embed-text-v1", 1536, "AWS Titan Embed Text v1"},
	{ModelTypeBedrock, "amazon.titan-embed-text-v2:0", 1024, "AWS Titan Embed Text v2"},
	{ModelTypeBedrock, "cohere.embed-english-v3", 1024, "Cohere Embed English v3 on Bedrock"},
	{ModelTypeBedrock, "cohere.embed-multilingual-v3", 1024, "Cohere Embed Multilingual v3 on Bedrock"},
}

// GetModelInfoExtended returns information about all available models including new providers.
func GetModelInfoExtended() []map[string]interface{} {
	// Combine original and extended models
	info := GetModelInfo()

	for _, model := range AvailableModelsExtended {
		info = append(info, map[string]interface{}{
			"type":        string(model.Type),
			"name":        model.Name,
			"dimension":   model.Dimension,
			"description": model.Description,
		})
	}
	return info
}
