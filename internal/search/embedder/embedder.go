// Package embedder provides embedding generation capabilities
package embedder

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"dev.helix.agent/internal/search/types"
)

// OpenAIEmbedder uses OpenAI's embedding API
type OpenAIEmbedder struct {
	apiKey     string
	model      string
	dimensions int
	client     *http.Client
	cache      map[string][]float32
}

// NewOpenAIEmbedder creates a new OpenAI embedder
func NewOpenAIEmbedder(apiKey, model string) *OpenAIEmbedder {
	if model == "" {
		model = "text-embedding-3-small"
	}

	dims := 1536
	if model == "text-embedding-3-large" {
		dims = 3072
	}

	return &OpenAIEmbedder{
		apiKey:     apiKey,
		model:      model,
		dimensions: dims,
		client:     &http.Client{Timeout: 30 * time.Second},
		cache:      make(map[string][]float32),
	}
}

// Embed generates embeddings for texts
func (e *OpenAIEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	results := make([][]float32, len(texts))

	// Check cache first
	var uncachedTexts []string
	var uncachedIndices []int

	for i, text := range texts {
		hash := e.hashText(text)
		if cached, ok := e.cache[hash]; ok {
			results[i] = cached
		} else {
			uncachedTexts = append(uncachedTexts, text)
			uncachedIndices = append(uncachedIndices, i)
		}
	}

	// Fetch uncached embeddings
	if len(uncachedTexts) > 0 {
		embeddings, err := e.fetchEmbeddings(ctx, uncachedTexts)
		if err != nil {
			return nil, err
		}

		for i, embedding := range embeddings {
			idx := uncachedIndices[i]
			results[idx] = embedding
			hash := e.hashText(uncachedTexts[i])
			e.cache[hash] = embedding
		}
	}

	return results, nil
}

// EmbedQuery generates embedding for a query
func (e *OpenAIEmbedder) EmbedQuery(ctx context.Context, query string) ([]float32, error) {
	embeddings, err := e.Embed(ctx, []string{query})
	if err != nil {
		return nil, err
	}
	return embeddings[0], nil
}

// Dimensions returns the embedding dimension
func (e *OpenAIEmbedder) Dimensions() int {
	return e.dimensions
}

func (e *OpenAIEmbedder) hashText(text string) string {
	hash := sha256.Sum256([]byte(text))
	return hex.EncodeToString(hash[:])
}

func (e *OpenAIEmbedder) fetchEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	// Prepare request
	reqBody := map[string]interface{}{
		"model": e.model,
		"input": texts,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST",
		"https://api.openai.com/v1/embeddings",
		strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+e.apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// Parse response
	var result struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
			Index     int       `json:"index"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	embeddings := make([][]float32, len(result.Data))
	for _, item := range result.Data {
		embeddings[item.Index] = item.Embedding
	}

	return embeddings, nil
}

// LocalEmbedder generates deterministic embeddings locally without external API calls.
// Useful for testing and development scenarios where API access is not available.
type LocalEmbedder struct {
	dimensions int
}

// NewLocalEmbedder creates a local embedder with deterministic output
func NewLocalEmbedder(dimensions int) *LocalEmbedder {
	return &LocalEmbedder{dimensions: dimensions}
}

// Embed generates deterministic embeddings based on text hash
func (e *LocalEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	results := make([][]float32, len(texts))
	for i, text := range texts {
		results[i] = e.generateEmbedding(text)
	}
	return results, nil
}

// EmbedQuery generates deterministic query embedding
func (e *LocalEmbedder) EmbedQuery(ctx context.Context, query string) ([]float32, error) {
	return e.generateEmbedding(query), nil
}

// Dimensions returns the embedding dimension
func (e *LocalEmbedder) Dimensions() int {
	return e.dimensions
}

// generateEmbedding creates a deterministic embedding from text using hash-based values
func (e *LocalEmbedder) generateEmbedding(text string) []float32 {
	embedding := make([]float32, e.dimensions)
	hash := sha256.Sum256([]byte(text))

	// Use hash bytes to generate deterministic but varied values
	for i := 0; i < e.dimensions; i++ {
		hashByte := hash[i%len(hash)]
		// Normalize to range [-1, 1]
		embedding[i] = (float32(hashByte) / 128.0) - 1.0
	}

	return embedding
}

// Ensure interfaces are implemented
var _ types.Embedder = (*OpenAIEmbedder)(nil)
var _ types.Embedder = (*LocalEmbedder)(nil)
