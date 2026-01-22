// Package rag provides enhanced Qdrant retriever with hybrid search and debate evaluation.
package rag

import (
	"context"
	"math"
	"sort"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

// QdrantEnhancedRetriever combines dense Qdrant retrieval with BM25 sparse retrieval
// and optional AI debate-based relevance evaluation.
type QdrantEnhancedRetriever struct {
	denseRetriever  Retriever
	sparseIndex     *EnhancedBM25Index
	reranker        Reranker
	debateEvaluator QdrantDebateEvaluator
	config          *QdrantEnhancedConfig
	logger          *logrus.Logger
	mu              sync.RWMutex
}

// QdrantDebateEvaluator uses AI debate to evaluate document relevance
type QdrantDebateEvaluator interface {
	EvaluateRelevance(ctx context.Context, query, document string) (float64, error)
}

// QdrantEnhancedConfig configuration for enhanced retriever
type QdrantEnhancedConfig struct {
	DenseWeight         float64      `json:"dense_weight"`
	SparseWeight        float64      `json:"sparse_weight"`
	UseDebateEvaluation bool         `json:"use_debate_evaluation"`
	DebateTopK          int          `json:"debate_top_k"`
	FusionMethod        FusionMethod `json:"fusion_method"`
	RRFK                float64      `json:"rrf_k"`
}

// DefaultQdrantEnhancedConfig returns default configuration
func DefaultQdrantEnhancedConfig() *QdrantEnhancedConfig {
	return &QdrantEnhancedConfig{
		DenseWeight:         0.6,
		SparseWeight:        0.4,
		UseDebateEvaluation: false,
		DebateTopK:          5,
		FusionMethod:        FusionRRF,
		RRFK:                60.0,
	}
}

// NewQdrantEnhancedRetriever creates a new enhanced Qdrant retriever
func NewQdrantEnhancedRetriever(
	denseRetriever Retriever,
	reranker Reranker,
	config *QdrantEnhancedConfig,
	logger *logrus.Logger,
) *QdrantEnhancedRetriever {
	if config == nil {
		config = DefaultQdrantEnhancedConfig()
	}
	if logger == nil {
		logger = logrus.New()
	}

	return &QdrantEnhancedRetriever{
		denseRetriever: denseRetriever,
		sparseIndex:    NewEnhancedBM25Index(),
		reranker:       reranker,
		config:         config,
		logger:         logger,
	}
}

// SetDebateEvaluator sets the debate evaluator for AI-based relevance
func (r *QdrantEnhancedRetriever) SetDebateEvaluator(evaluator QdrantDebateEvaluator) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.debateEvaluator = evaluator
	r.config.UseDebateEvaluation = evaluator != nil
}

// Retrieve implements Retriever interface with hybrid search
func (r *QdrantEnhancedRetriever) Retrieve(ctx context.Context, query string, opts *SearchOptions) ([]*SearchResult, error) {
	if opts == nil {
		opts = DefaultSearchOptions()
	}

	// Expand retrieval to get more candidates
	expandedOpts := &SearchOptions{
		TopK:            opts.TopK * 3,
		MinScore:        opts.MinScore,
		Filter:          opts.Filter,
		EnableReranking: false, // We'll rerank later
		HybridAlpha:     opts.HybridAlpha,
		IncludeMetadata: opts.IncludeMetadata,
		Namespace:       opts.Namespace,
	}

	// Dense retrieval
	denseResults, err := r.denseRetriever.Retrieve(ctx, query, expandedOpts)
	if err != nil {
		r.logger.WithError(err).Warn("Dense retrieval failed, using sparse only")
		denseResults = []*SearchResult{}
	}

	// Sparse retrieval using BM25
	sparseResults := r.sparseIndex.Search(query, expandedOpts.TopK)

	// Fuse results
	fusedResults := r.fuseResults(denseResults, sparseResults)

	// Limit to requested TopK
	if len(fusedResults) > opts.TopK {
		fusedResults = fusedResults[:opts.TopK]
	}

	// Rerank if enabled and reranker available
	if opts.EnableReranking && r.reranker != nil {
		reranked, err := r.reranker.Rerank(ctx, query, fusedResults, opts.TopK)
		if err != nil {
			r.logger.WithError(err).Warn("Reranking failed, using fused results")
		} else {
			fusedResults = reranked
		}
	}

	// Use debate-based evaluation if enabled
	r.mu.RLock()
	useDebate := r.config.UseDebateEvaluation && r.debateEvaluator != nil
	debateEval := r.debateEvaluator
	debateTopK := r.config.DebateTopK
	r.mu.RUnlock()

	if useDebate && len(fusedResults) > 0 {
		fusedResults = r.evaluateWithDebate(ctx, query, fusedResults, debateEval, debateTopK)
	}

	r.logger.WithFields(logrus.Fields{
		"query":        truncateText(query, 50),
		"dense_count":  len(denseResults),
		"sparse_count": len(sparseResults),
		"fused_count":  len(fusedResults),
	}).Debug("Hybrid retrieval completed")

	return fusedResults, nil
}

// Index implements Retriever interface
func (r *QdrantEnhancedRetriever) Index(ctx context.Context, docs []*Document) error {
	// Index in dense retriever
	if err := r.denseRetriever.Index(ctx, docs); err != nil {
		return err
	}

	// Index in sparse BM25 index
	for _, doc := range docs {
		r.sparseIndex.AddDocument(doc.ID, doc.Content)
	}

	r.logger.WithField("count", len(docs)).Debug("Documents indexed for hybrid search")
	return nil
}

// Delete implements Retriever interface
func (r *QdrantEnhancedRetriever) Delete(ctx context.Context, ids []string) error {
	// Delete from dense retriever
	if err := r.denseRetriever.Delete(ctx, ids); err != nil {
		return err
	}

	// Delete from sparse index
	for _, id := range ids {
		r.sparseIndex.RemoveDocument(id)
	}

	return nil
}

func (r *QdrantEnhancedRetriever) fuseResults(denseResults, sparseResults []*SearchResult) []*SearchResult {
	switch r.config.FusionMethod {
	case FusionRRF:
		return r.rrfFusion(denseResults, sparseResults)
	case FusionWeighted:
		return r.weightedFusion(denseResults, sparseResults)
	default:
		return r.rrfFusion(denseResults, sparseResults)
	}
}

func (r *QdrantEnhancedRetriever) rrfFusion(denseResults, sparseResults []*SearchResult) []*SearchResult {
	k := r.config.RRFK
	scores := make(map[string]float64)
	docs := make(map[string]*SearchResult)

	// Score dense results
	for rank, result := range denseResults {
		if result.Document == nil {
			continue
		}
		id := result.Document.ID
		scores[id] += r.config.DenseWeight / (k + float64(rank+1))
		if _, exists := docs[id]; !exists {
			docs[id] = result
		}
	}

	// Score sparse results
	for rank, result := range sparseResults {
		if result.Document == nil {
			continue
		}
		id := result.Document.ID
		scores[id] += r.config.SparseWeight / (k + float64(rank+1))
		if _, exists := docs[id]; !exists {
			docs[id] = result
		}
	}

	// Create fused results
	var fused []*SearchResult
	for id, score := range scores {
		if doc, ok := docs[id]; ok {
			fused = append(fused, &SearchResult{
				Document:  doc.Document,
				Score:     score,
				MatchType: MatchTypeHybrid,
			})
		}
	}

	sort.Slice(fused, func(i, j int) bool {
		return fused[i].Score > fused[j].Score
	})

	return fused
}

func (r *QdrantEnhancedRetriever) weightedFusion(denseResults, sparseResults []*SearchResult) []*SearchResult {
	scores := make(map[string]float64)
	docs := make(map[string]*SearchResult)

	// Normalize and weight dense scores
	maxDense := 0.0
	for _, result := range denseResults {
		if result.Score > maxDense {
			maxDense = result.Score
		}
	}
	if maxDense > 0 {
		for _, result := range denseResults {
			if result.Document == nil {
				continue
			}
			id := result.Document.ID
			scores[id] += (result.Score / maxDense) * r.config.DenseWeight
			if _, exists := docs[id]; !exists {
				docs[id] = result
			}
		}
	}

	// Normalize and weight sparse scores
	maxSparse := 0.0
	for _, result := range sparseResults {
		if result.Score > maxSparse {
			maxSparse = result.Score
		}
	}
	if maxSparse > 0 {
		for _, result := range sparseResults {
			if result.Document == nil {
				continue
			}
			id := result.Document.ID
			scores[id] += (result.Score / maxSparse) * r.config.SparseWeight
			if _, exists := docs[id]; !exists {
				docs[id] = result
			}
		}
	}

	var fused []*SearchResult
	for id, score := range scores {
		if doc, ok := docs[id]; ok {
			fused = append(fused, &SearchResult{
				Document:  doc.Document,
				Score:     score,
				MatchType: MatchTypeHybrid,
			})
		}
	}

	sort.Slice(fused, func(i, j int) bool {
		return fused[i].Score > fused[j].Score
	})

	return fused
}

func (r *QdrantEnhancedRetriever) evaluateWithDebate(
	ctx context.Context,
	query string,
	results []*SearchResult,
	evaluator QdrantDebateEvaluator,
	maxEval int,
) []*SearchResult {
	if maxEval > len(results) {
		maxEval = len(results)
	}

	for i := 0; i < maxEval; i++ {
		result := results[i]
		if result.Document == nil {
			continue
		}

		relevance, err := evaluator.EvaluateRelevance(ctx, query, result.Document.Content)
		if err != nil {
			r.logger.WithError(err).Warn("Debate evaluation failed")
			continue
		}

		// Combine original score with debate relevance
		result.Score = result.Score*0.6 + relevance*0.4
		result.RerankedScore = relevance
	}

	// Re-sort after debate evaluation
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

// EnhancedBM25Index provides BM25 sparse retrieval for hybrid search
type EnhancedBM25Index struct {
	documents  map[string]string
	termFreqs  map[string]map[string]int
	docFreqs   map[string]int
	docLengths map[string]int
	avgDocLen  float64
	totalDocs  int
	k1         float64
	b          float64
	mu         sync.RWMutex
}

// NewEnhancedBM25Index creates a new BM25 index for enhanced retrieval
func NewEnhancedBM25Index() *EnhancedBM25Index {
	return &EnhancedBM25Index{
		documents:  make(map[string]string),
		termFreqs:  make(map[string]map[string]int),
		docFreqs:   make(map[string]int),
		docLengths: make(map[string]int),
		k1:         1.2,
		b:          0.75,
	}
}

// AddDocument adds a document to the BM25 index
func (idx *EnhancedBM25Index) AddDocument(id, content string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	terms := enhancedTokenize(content)

	idx.documents[id] = content
	idx.termFreqs[id] = make(map[string]int)
	idx.docLengths[id] = len(terms)

	termsSeen := make(map[string]bool)
	for _, term := range terms {
		idx.termFreqs[id][term]++
		if !termsSeen[term] {
			idx.docFreqs[term]++
			termsSeen[term] = true
		}
	}

	idx.totalDocs++
	idx.recalculateAvgDocLen()
}

// RemoveDocument removes a document from the index
func (idx *EnhancedBM25Index) RemoveDocument(id string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	if _, exists := idx.documents[id]; !exists {
		return
	}

	// Decrement doc frequencies
	for term := range idx.termFreqs[id] {
		idx.docFreqs[term]--
		if idx.docFreqs[term] <= 0 {
			delete(idx.docFreqs, term)
		}
	}

	delete(idx.documents, id)
	delete(idx.termFreqs, id)
	delete(idx.docLengths, id)
	idx.totalDocs--
	idx.recalculateAvgDocLen()
}

// Search performs BM25 search
func (idx *EnhancedBM25Index) Search(query string, topK int) []*SearchResult {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	queryTerms := enhancedTokenize(query)
	scores := make(map[string]float64)

	for _, term := range queryTerms {
		df, exists := idx.docFreqs[term]
		if !exists {
			continue
		}

		idf := idx.calculateIDF(df)

		for docID, tf := range idx.termFreqs {
			termFreq, ok := tf[term]
			if !ok {
				continue
			}

			docLen := float64(idx.docLengths[docID])
			tfScore := idx.calculateTF(float64(termFreq), docLen)
			scores[docID] += idf * tfScore
		}
	}

	// Convert to results
	var results []*SearchResult
	for docID, score := range scores {
		results = append(results, &SearchResult{
			Document: &Document{
				ID:      docID,
				Content: idx.documents[docID],
			},
			Score:     score,
			MatchType: MatchTypeSparse,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > topK {
		results = results[:topK]
	}

	return results
}

func (idx *EnhancedBM25Index) calculateIDF(df int) float64 {
	n := float64(idx.totalDocs)
	return math.Log((n-float64(df)+0.5)/(float64(df)+0.5) + 1)
}

func (idx *EnhancedBM25Index) calculateTF(tf, docLen float64) float64 {
	return (tf * (idx.k1 + 1)) / (tf + idx.k1*(1-idx.b+idx.b*(docLen/idx.avgDocLen)))
}

func (idx *EnhancedBM25Index) recalculateAvgDocLen() {
	total := 0
	for _, length := range idx.docLengths {
		total += length
	}
	if idx.totalDocs > 0 {
		idx.avgDocLen = float64(total) / float64(idx.totalDocs)
	}
}

// enhancedTokenize tokenizes text for BM25 (renamed to avoid conflict with existing tokenize)
func enhancedTokenize(text string) []string {
	text = strings.ToLower(text)
	words := strings.Fields(text)

	var tokens []string
	for _, word := range words {
		cleaned := strings.Trim(word, ".,!?;:\"'()[]{}#$%&*+-/<>=@\\^_`|~")
		if len(cleaned) > 0 {
			tokens = append(tokens, cleaned)
		}
	}

	return tokens
}

// truncateText truncates text to a maximum length
func truncateText(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
