package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"github.com/sirupsen/logrus"
)

// CrossEncoderReranker uses a cross-encoder model for reranking
type CrossEncoderReranker struct {
	config     *RerankerConfig
	httpClient *http.Client
	logger     *logrus.Logger
}

// RerankerConfig configures the reranker
type RerankerConfig struct {
	// Model name for the cross-encoder
	Model string `json:"model"`
	// Endpoint for the reranker API
	Endpoint string `json:"endpoint"`
	// APIKey for authentication
	APIKey string `json:"api_key"`
	// Timeout for requests
	Timeout time.Duration `json:"timeout"`
	// BatchSize for batching requests
	BatchSize int `json:"batch_size"`
	// ReturnScores whether to return scores
	ReturnScores bool `json:"return_scores"`
}

// DefaultRerankerConfig returns default configuration
func DefaultRerankerConfig() *RerankerConfig {
	return &RerankerConfig{
		Model:        "BAAI/bge-reranker-v2-m3",
		Timeout:      30 * time.Second,
		BatchSize:    32,
		ReturnScores: true,
	}
}

// NewCrossEncoderReranker creates a new cross-encoder reranker
func NewCrossEncoderReranker(config *RerankerConfig, logger *logrus.Logger) *CrossEncoderReranker {
	if config == nil {
		config = DefaultRerankerConfig()
	}
	if logger == nil {
		logger = logrus.New()
	}

	return &CrossEncoderReranker{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		logger: logger,
	}
}

// Rerank reorders results using cross-encoder scoring
func (r *CrossEncoderReranker) Rerank(ctx context.Context, query string, results []*SearchResult, topK int) ([]*SearchResult, error) {
	if len(results) == 0 {
		return results, nil
	}

	// If no API endpoint, use a simple fallback
	if r.config.Endpoint == "" {
		return r.fallbackRerank(query, results, topK)
	}

	// Prepare pairs for cross-encoder
	pairs := make([][2]string, len(results))
	for i, result := range results {
		pairs[i] = [2]string{query, result.Document.Content}
	}

	// Score in batches
	scores := make([]float64, len(results))
	for i := 0; i < len(pairs); i += r.config.BatchSize {
		end := i + r.config.BatchSize
		if end > len(pairs) {
			end = len(pairs)
		}

		batchScores, err := r.scoreBatch(ctx, pairs[i:end])
		if err != nil {
			r.logger.WithError(err).Warn("Batch scoring failed, using original scores")
			// Fall back to original scores
			for j, result := range results[i:end] {
				scores[i+j] = result.Score
			}
		} else {
			copy(scores[i:end], batchScores)
		}
	}

	// Update scores and sort
	for i, result := range results {
		result.RerankedScore = scores[i]
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].RerankedScore > results[j].RerankedScore
	})

	// Limit to topK
	if len(results) > topK {
		results = results[:topK]
	}

	return results, nil
}

// scoreBatch scores a batch of query-document pairs
func (r *CrossEncoderReranker) scoreBatch(ctx context.Context, pairs [][2]string) ([]float64, error) {
	reqBody := map[string]interface{}{
		"model": r.config.Model,
		"pairs": pairs,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.config.Endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if r.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+r.config.APIKey)
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("reranker returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Scores []float64 `json:"scores"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Scores, nil
}

// fallbackRerank uses a simple heuristic when no API is available
func (r *CrossEncoderReranker) fallbackRerank(query string, results []*SearchResult, topK int) ([]*SearchResult, error) {
	// Simple keyword overlap scoring as fallback
	queryWords := tokenizeToFrequencyMap(query)

	for _, result := range results {
		docWords := tokenizeToFrequencyMap(result.Document.Content)
		overlap := computeOverlap(queryWords, docWords)
		// Combine original score with overlap
		result.RerankedScore = result.Score*0.7 + overlap*0.3
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].RerankedScore > results[j].RerankedScore
	})

	if len(results) > topK {
		results = results[:topK]
	}

	return results, nil
}

// CohereReranker uses Cohere's reranker API
type CohereReranker struct {
	apiKey     string
	model      string
	httpClient *http.Client
	logger     *logrus.Logger
}

// NewCohereReranker creates a Cohere reranker
func NewCohereReranker(apiKey, model string, logger *logrus.Logger) *CohereReranker {
	if model == "" {
		model = "rerank-english-v3.0"
	}
	if logger == nil {
		logger = logrus.New()
	}

	return &CohereReranker{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// Rerank uses Cohere's reranker
func (r *CohereReranker) Rerank(ctx context.Context, query string, results []*SearchResult, topK int) ([]*SearchResult, error) {
	if len(results) == 0 {
		return results, nil
	}

	// Prepare documents
	documents := make([]string, len(results))
	for i, result := range results {
		documents[i] = result.Document.Content
	}

	reqBody := map[string]interface{}{
		"model":           r.model,
		"query":           query,
		"documents":       documents,
		"top_n":           topK,
		"return_documents": false,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.cohere.ai/v1/rerank", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+r.apiKey)

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Cohere returned status %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Results []struct {
			Index          int     `json:"index"`
			RelevanceScore float64 `json:"relevance_score"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Reorder results based on Cohere ranking
	reranked := make([]*SearchResult, len(response.Results))
	for i, r := range response.Results {
		result := results[r.Index]
		result.RerankedScore = r.RelevanceScore
		reranked[i] = result
	}

	return reranked, nil
}

// Helper functions

func tokenizeToFrequencyMap(text string) map[string]int {
	words := make(map[string]int)
	word := ""
	for _, r := range text {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
			word += string(r)
		} else if word != "" {
			words[word]++
			word = ""
		}
	}
	if word != "" {
		words[word]++
	}
	return words
}

func computeOverlap(query, doc map[string]int) float64 {
	if len(query) == 0 || len(doc) == 0 {
		return 0
	}

	overlap := 0
	for word := range query {
		if _, exists := doc[word]; exists {
			overlap++
		}
	}

	return float64(overlap) / float64(len(query))
}
