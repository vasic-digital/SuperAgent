package rag

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/sirupsen/logrus"
)

// HybridRetriever combines dense and sparse retrieval with fusion
type HybridRetriever struct {
	denseRetriever  DenseRetriever
	sparseRetriever SparseRetriever
	reranker        Reranker
	config          *HybridConfig
	logger          *logrus.Logger
	mu              sync.RWMutex
}

// HybridConfig configures hybrid retrieval
type HybridConfig struct {
	// Alpha controls the balance: 0=sparse only, 1=dense only
	Alpha float64 `json:"alpha"`
	// FusionMethod determines how to combine results
	FusionMethod FusionMethod `json:"fusion_method"`
	// K for RRF (Reciprocal Rank Fusion)
	RRFK int `json:"rrf_k"`
	// EnableReranking enables cross-encoder reranking
	EnableReranking bool `json:"enable_reranking"`
	// RerankTopK limits results before reranking
	RerankTopK int `json:"rerank_top_k"`
	// PreRetrieveMultiplier retrieves N*topK before fusion
	PreRetrieveMultiplier int `json:"pre_retrieve_multiplier"`
}

// FusionMethod defines how to combine results
type FusionMethod string

const (
	FusionRRF      FusionMethod = "rrf"      // Reciprocal Rank Fusion
	FusionWeighted FusionMethod = "weighted" // Weighted score combination
	FusionMax      FusionMethod = "max"      // Max score wins
)

// DefaultHybridConfig returns default configuration
func DefaultHybridConfig() *HybridConfig {
	return &HybridConfig{
		Alpha:                 0.5,
		FusionMethod:          FusionRRF,
		RRFK:                  60,
		EnableReranking:       true,
		RerankTopK:            50,
		PreRetrieveMultiplier: 3,
	}
}

// NewHybridRetriever creates a new hybrid retriever
func NewHybridRetriever(
	denseRetriever DenseRetriever,
	sparseRetriever SparseRetriever,
	reranker Reranker,
	config *HybridConfig,
	logger *logrus.Logger,
) *HybridRetriever {
	if config == nil {
		config = DefaultHybridConfig()
	}
	if logger == nil {
		logger = logrus.New()
	}

	return &HybridRetriever{
		denseRetriever:  denseRetriever,
		sparseRetriever: sparseRetriever,
		reranker:        reranker,
		config:          config,
		logger:          logger,
	}
}

// Retrieve performs hybrid search combining dense and sparse results
func (h *HybridRetriever) Retrieve(ctx context.Context, query string, opts *SearchOptions) ([]*SearchResult, error) {
	if opts == nil {
		opts = DefaultSearchOptions()
	}

	// Determine pre-retrieval count
	preRetrieveK := opts.TopK * h.config.PreRetrieveMultiplier

	// Run dense and sparse retrieval in parallel
	var wg sync.WaitGroup
	var denseResults, sparseResults []*SearchResult
	var denseErr, sparseErr error

	wg.Add(2)

	// Dense retrieval (semantic)
	go func() {
		defer wg.Done()
		denseOpts := *opts
		denseOpts.TopK = preRetrieveK
		denseResults, denseErr = h.denseRetriever.Retrieve(ctx, query, &denseOpts)
	}()

	// Sparse retrieval (keyword)
	go func() {
		defer wg.Done()
		sparseOpts := *opts
		sparseOpts.TopK = preRetrieveK
		sparseResults, sparseErr = h.sparseRetriever.Retrieve(ctx, query, &sparseOpts)
	}()

	wg.Wait()

	// Handle errors
	if denseErr != nil && sparseErr != nil {
		return nil, fmt.Errorf("both retrievers failed: dense=%v, sparse=%v", denseErr, sparseErr)
	}

	// Mark match types
	for _, r := range denseResults {
		r.MatchType = MatchTypeDense
	}
	for _, r := range sparseResults {
		r.MatchType = MatchTypeSparse
	}

	// Fuse results
	var fusedResults []*SearchResult
	switch h.config.FusionMethod {
	case FusionRRF:
		fusedResults = h.reciprocalRankFusion(denseResults, sparseResults)
	case FusionWeighted:
		fusedResults = h.weightedFusion(denseResults, sparseResults, h.config.Alpha)
	case FusionMax:
		fusedResults = h.maxFusion(denseResults, sparseResults)
	default:
		fusedResults = h.reciprocalRankFusion(denseResults, sparseResults)
	}

	// Mark as hybrid match
	for _, r := range fusedResults {
		r.MatchType = MatchTypeHybrid
	}

	// Rerank if enabled
	if h.config.EnableReranking && h.reranker != nil && len(fusedResults) > 0 {
		rerankK := h.config.RerankTopK
		if rerankK > len(fusedResults) {
			rerankK = len(fusedResults)
		}

		reranked, err := h.reranker.Rerank(ctx, query, fusedResults[:rerankK], opts.TopK)
		if err != nil {
			h.logger.WithError(err).Warn("Reranking failed, using fused results")
		} else {
			fusedResults = reranked
		}
	}

	// Limit to topK
	if len(fusedResults) > opts.TopK {
		fusedResults = fusedResults[:opts.TopK]
	}

	// Filter by minimum score
	if opts.MinScore > 0 {
		filtered := make([]*SearchResult, 0, len(fusedResults))
		for _, r := range fusedResults {
			if r.Score >= opts.MinScore {
				filtered = append(filtered, r)
			}
		}
		fusedResults = filtered
	}

	h.logger.WithFields(logrus.Fields{
		"query":         query[:min(50, len(query))],
		"dense_count":   len(denseResults),
		"sparse_count":  len(sparseResults),
		"fused_count":   len(fusedResults),
		"fusion_method": h.config.FusionMethod,
	}).Debug("Hybrid search completed")

	return fusedResults, nil
}

// reciprocalRankFusion implements RRF algorithm
// RRF(d) = Σ 1/(k + rank(d))
func (h *HybridRetriever) reciprocalRankFusion(denseResults, sparseResults []*SearchResult) []*SearchResult {
	k := float64(h.config.RRFK)
	scoreMap := make(map[string]float64)
	docMap := make(map[string]*SearchResult)

	// Add dense results
	for i, r := range denseResults {
		id := r.Document.ID
		scoreMap[id] += 1.0 / (k + float64(i+1))
		if _, exists := docMap[id]; !exists {
			docMap[id] = r
		}
	}

	// Add sparse results
	for i, r := range sparseResults {
		id := r.Document.ID
		scoreMap[id] += 1.0 / (k + float64(i+1))
		if _, exists := docMap[id]; !exists {
			docMap[id] = r
		}
	}

	// Create result list
	results := make([]*SearchResult, 0, len(scoreMap))
	for id, score := range scoreMap {
		result := docMap[id]
		result.Score = score
		results = append(results, result)
	}

	// Sort by fused score
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

// weightedFusion combines scores with alpha weighting
func (h *HybridRetriever) weightedFusion(denseResults, sparseResults []*SearchResult, alpha float64) []*SearchResult {
	scoreMap := make(map[string]float64)
	docMap := make(map[string]*SearchResult)

	// Normalize and add dense scores
	maxDense := 0.0
	for _, r := range denseResults {
		if r.Score > maxDense {
			maxDense = r.Score
		}
	}
	for _, r := range denseResults {
		id := r.Document.ID
		normalizedScore := 0.0
		if maxDense > 0 {
			normalizedScore = r.Score / maxDense
		}
		scoreMap[id] += alpha * normalizedScore
		if _, exists := docMap[id]; !exists {
			docMap[id] = r
		}
	}

	// Normalize and add sparse scores
	maxSparse := 0.0
	for _, r := range sparseResults {
		if r.Score > maxSparse {
			maxSparse = r.Score
		}
	}
	for _, r := range sparseResults {
		id := r.Document.ID
		normalizedScore := 0.0
		if maxSparse > 0 {
			normalizedScore = r.Score / maxSparse
		}
		scoreMap[id] += (1 - alpha) * normalizedScore
		if _, exists := docMap[id]; !exists {
			docMap[id] = r
		}
	}

	// Create result list
	results := make([]*SearchResult, 0, len(scoreMap))
	for id, score := range scoreMap {
		result := docMap[id]
		result.Score = score
		results = append(results, result)
	}

	// Sort by fused score
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

// maxFusion takes the maximum score for each document
func (h *HybridRetriever) maxFusion(denseResults, sparseResults []*SearchResult) []*SearchResult {
	scoreMap := make(map[string]float64)
	docMap := make(map[string]*SearchResult)

	for _, r := range denseResults {
		id := r.Document.ID
		if r.Score > scoreMap[id] {
			scoreMap[id] = r.Score
			docMap[id] = r
		}
	}

	for _, r := range sparseResults {
		id := r.Document.ID
		if r.Score > scoreMap[id] {
			scoreMap[id] = r.Score
			docMap[id] = r
		}
	}

	results := make([]*SearchResult, 0, len(scoreMap))
	for id := range scoreMap {
		results = append(results, docMap[id])
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

// Index adds documents to both retrievers
func (h *HybridRetriever) Index(ctx context.Context, docs []*Document) error {
	var wg sync.WaitGroup
	var denseErr, sparseErr error

	wg.Add(2)

	go func() {
		defer wg.Done()
		denseErr = h.denseRetriever.Index(ctx, docs)
	}()

	go func() {
		defer wg.Done()
		sparseErr = h.sparseRetriever.Index(ctx, docs)
	}()

	wg.Wait()

	if denseErr != nil {
		return fmt.Errorf("dense indexing failed: %w", denseErr)
	}
	if sparseErr != nil {
		return fmt.Errorf("sparse indexing failed: %w", sparseErr)
	}

	return nil
}

// Delete removes documents from both retrievers
func (h *HybridRetriever) Delete(ctx context.Context, ids []string) error {
	var wg sync.WaitGroup
	var denseErr, sparseErr error

	wg.Add(2)

	go func() {
		defer wg.Done()
		denseErr = h.denseRetriever.Delete(ctx, ids)
	}()

	go func() {
		defer wg.Done()
		sparseErr = h.sparseRetriever.Delete(ctx, ids)
	}()

	wg.Wait()

	if denseErr != nil {
		return fmt.Errorf("dense deletion failed: %w", denseErr)
	}
	if sparseErr != nil {
		return fmt.Errorf("sparse deletion failed: %w", sparseErr)
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
