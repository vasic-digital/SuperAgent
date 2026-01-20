package rag

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
	"dev.helix.agent/internal/vectordb/qdrant"
)

// EnhancedQdrantRetriever combines dense retrieval with BM25 sparse retrieval
type EnhancedQdrantRetriever struct {
	denseRetriever  *QdrantDenseRetriever
	sparseIndex     *BM25Index
	reranker        Reranker
	fusionMethod    FusionMethod
	fusionWeights   *FusionWeights
	logger          *logrus.Logger
	debateEvaluator DebateEvaluator
}

// DebateEvaluator interface for debate-based relevance evaluation
type DebateEvaluator interface {
	EvaluateRelevance(ctx context.Context, query, document string) (float64, error)
}

// NewEnhancedQdrantRetriever creates a new enhanced Qdrant retriever
func NewEnhancedQdrantRetriever(
	client *qdrant.Client,
	collection string,
	embedder Embedder,
	reranker Reranker,
	logger *logrus.Logger,
) *EnhancedQdrantRetriever {
	if logger == nil {
		logger = logrus.New()
	}

	return &EnhancedQdrantRetriever{
		denseRetriever: NewQdrantDenseRetriever(client, collection, embedder, logger),
		sparseIndex:    NewBM25Index(),
		reranker:       reranker,
		fusionMethod:   FusionRRF,
		fusionWeights: &FusionWeights{
			DenseWeight:  0.6,
			SparseWeight: 0.4,
		},
		logger: logger,
	}
}

// SetFusionMethod sets the fusion method for hybrid retrieval
func (r *EnhancedQdrantRetriever) SetFusionMethod(method FusionMethod) {
	r.fusionMethod = method
}

// SetFusionWeights sets custom weights for weighted fusion
func (r *EnhancedQdrantRetriever) SetFusionWeights(weights *FusionWeights) {
	r.fusionWeights = weights
}

// SetDebateEvaluator sets the debate evaluator for AI-based relevance
func (r *EnhancedQdrantRetriever) SetDebateEvaluator(evaluator DebateEvaluator) {
	r.debateEvaluator = evaluator
}

// IndexDocuments indexes documents for sparse retrieval
func (r *EnhancedQdrantRetriever) IndexDocuments(ctx context.Context, docs []*Document) error {
	for _, doc := range docs {
		r.sparseIndex.AddDocument(doc.ID, doc.Content)
	}
	r.logger.WithField("count", len(docs)).Debug("Documents indexed for sparse retrieval")
	return nil
}

// HybridRetrieve performs hybrid dense + sparse retrieval
func (r *EnhancedQdrantRetriever) HybridRetrieve(ctx context.Context, query string, opts *SearchOptions) ([]*SearchResult, error) {
	if opts == nil {
		opts = &SearchOptions{TopK: 10}
	}

	// Get more results than needed for fusion
	expandedOpts := &SearchOptions{
		TopK:     opts.TopK * 3,
		MinScore: opts.MinScore,
		Filters:  opts.Filters,
	}

	// Dense retrieval from Qdrant
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

	// Rerank if reranker is available
	if r.reranker != nil {
		reranked, err := r.reranker.Rerank(ctx, query, fusedResults, &RerankOptions{TopK: opts.TopK})
		if err != nil {
			r.logger.WithError(err).Warn("Reranking failed, using fused results")
		} else {
			fusedResults = reranked
		}
	}

	// Use debate-based evaluation if available for top results
	if r.debateEvaluator != nil && len(fusedResults) > 0 {
		fusedResults = r.evaluateWithDebate(ctx, query, fusedResults)
	}

	r.logger.WithFields(logrus.Fields{
		"query":        truncate(query, 50),
		"dense_count":  len(denseResults),
		"sparse_count": len(sparseResults),
		"fused_count":  len(fusedResults),
	}).Debug("Hybrid retrieval completed")

	return fusedResults, nil
}

// Retrieve implements DenseRetriever interface (delegates to hybrid)
func (r *EnhancedQdrantRetriever) Retrieve(ctx context.Context, query string, opts *SearchOptions) ([]*SearchResult, error) {
	return r.HybridRetrieve(ctx, query, opts)
}

// GetName returns the retriever name
func (r *EnhancedQdrantRetriever) GetName() string {
	return "enhanced_qdrant_hybrid"
}

func (r *EnhancedQdrantRetriever) fuseResults(denseResults, sparseResults []*SearchResult) []*SearchResult {
	switch r.fusionMethod {
	case FusionRRF:
		return r.reciprocalRankFusion(denseResults, sparseResults)
	case FusionWeighted:
		return r.weightedFusion(denseResults, sparseResults)
	default:
		return r.reciprocalRankFusion(denseResults, sparseResults)
	}
}

func (r *EnhancedQdrantRetriever) reciprocalRankFusion(denseResults, sparseResults []*SearchResult) []*SearchResult {
	const k = 60.0 // RRF constant

	scores := make(map[string]float64)
	docs := make(map[string]*SearchResult)

	// Score dense results
	for rank, result := range denseResults {
		id := result.Document.ID
		scores[id] += 1.0 / (k + float64(rank+1))
		if _, exists := docs[id]; !exists {
			docs[id] = result
		}
	}

	// Score sparse results
	for rank, result := range sparseResults {
		id := result.Document.ID
		scores[id] += 1.0 / (k + float64(rank+1))
		if _, exists := docs[id]; !exists {
			docs[id] = result
		}
	}

	// Create fused results
	var fused []*SearchResult
	for id, score := range scores {
		if doc, ok := docs[id]; ok {
			fused = append(fused, &SearchResult{
				Document: doc.Document,
				Score:    score,
				Metadata: doc.Metadata,
			})
		}
	}

	// Sort by score descending
	sort.Slice(fused, func(i, j int) bool {
		return fused[i].Score > fused[j].Score
	})

	return fused
}

func (r *EnhancedQdrantRetriever) weightedFusion(denseResults, sparseResults []*SearchResult) []*SearchResult {
	denseWeight := r.fusionWeights.DenseWeight
	sparseWeight := r.fusionWeights.SparseWeight

	// Normalize weights
	total := denseWeight + sparseWeight
	denseWeight /= total
	sparseWeight /= total

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
			id := result.Document.ID
			scores[id] += (result.Score / maxDense) * denseWeight
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
			id := result.Document.ID
			scores[id] += (result.Score / maxSparse) * sparseWeight
			if _, exists := docs[id]; !exists {
				docs[id] = result
			}
		}
	}

	// Create fused results
	var fused []*SearchResult
	for id, score := range scores {
		if doc, ok := docs[id]; ok {
			fused = append(fused, &SearchResult{
				Document: doc.Document,
				Score:    score,
				Metadata: doc.Metadata,
			})
		}
	}

	sort.Slice(fused, func(i, j int) bool {
		return fused[i].Score > fused[j].Score
	})

	return fused
}

func (r *EnhancedQdrantRetriever) evaluateWithDebate(ctx context.Context, query string, results []*SearchResult) []*SearchResult {
	// Only evaluate top N results with debate (expensive operation)
	maxEval := 5
	if len(results) < maxEval {
		maxEval = len(results)
	}

	for i := 0; i < maxEval; i++ {
		result := results[i]
		relevance, err := r.debateEvaluator.EvaluateRelevance(ctx, query, result.Document.Content)
		if err != nil {
			r.logger.WithError(err).Warn("Debate evaluation failed")
			continue
		}

		// Combine original score with debate relevance
		result.Score = result.Score*0.6 + relevance*0.4
		if result.Metadata == nil {
			result.Metadata = make(map[string]interface{})
		}
		result.Metadata["debate_relevance"] = relevance
	}

	// Re-sort after debate evaluation
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

// BM25Index provides BM25 sparse retrieval
type BM25Index struct {
	documents   map[string]string
	termFreqs   map[string]map[string]int // docID -> term -> freq
	docFreqs    map[string]int            // term -> doc count
	docLengths  map[string]int
	avgDocLen   float64
	totalDocs   int
	k1          float64
	b           float64
}

// NewBM25Index creates a new BM25 index
func NewBM25Index() *BM25Index {
	return &BM25Index{
		documents:  make(map[string]string),
		termFreqs:  make(map[string]map[string]int),
		docFreqs:   make(map[string]int),
		docLengths: make(map[string]int),
		k1:         1.2,
		b:          0.75,
	}
}

// AddDocument adds a document to the index
func (idx *BM25Index) AddDocument(id, content string) {
	terms := tokenize(content)

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

// Search performs BM25 search
func (idx *BM25Index) Search(query string, topK int) []*SearchResult {
	queryTerms := tokenize(query)

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
			Score: score,
		})
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > topK {
		results = results[:topK]
	}

	return results
}

func (idx *BM25Index) calculateIDF(df int) float64 {
	n := float64(idx.totalDocs)
	return logN((n-float64(df)+0.5)/(float64(df)+0.5) + 1)
}

func (idx *BM25Index) calculateTF(tf, docLen float64) float64 {
	return (tf * (idx.k1 + 1)) / (tf + idx.k1*(1-idx.b+idx.b*(docLen/idx.avgDocLen)))
}

func (idx *BM25Index) recalculateAvgDocLen() {
	total := 0
	for _, length := range idx.docLengths {
		total += length
	}
	if idx.totalDocs > 0 {
		idx.avgDocLen = float64(total) / float64(idx.totalDocs)
	}
}

func tokenize(text string) []string {
	// Simple whitespace tokenization with lowercase
	text = strings.ToLower(text)
	words := strings.Fields(text)

	// Remove punctuation
	var tokens []string
	for _, word := range words {
		cleaned := strings.Trim(word, ".,!?;:\"'()[]{}#$%&*+-/<>=@\\^_`|~")
		if len(cleaned) > 0 {
			tokens = append(tokens, cleaned)
		}
	}

	return tokens
}

func logN(x float64) float64 {
	// Natural log with protection against log(0)
	if x <= 0 {
		return 0
	}
	// Using approximation for simplicity
	result := 0.0
	for x > 2 {
		x /= 2
		result += 0.693147 // ln(2)
	}
	// Taylor series for ln(x) around x=1
	y := x - 1
	for i := 1; i <= 10; i++ {
		if i%2 == 1 {
			result += pow(y, i) / float64(i)
		} else {
			result -= pow(y, i) / float64(i)
		}
	}
	return result
}

func pow(base float64, exp int) float64 {
	result := 1.0
	for i := 0; i < exp; i++ {
		result *= base
	}
	return result
}

// QdrantRAGPipeline provides a complete RAG pipeline using Qdrant
type QdrantRAGPipeline struct {
	retriever       *EnhancedQdrantRetriever
	documentStore   *QdrantDocumentStore
	contextBuilder  ContextBuilder
	logger          *logrus.Logger
}

// ContextBuilder builds context from retrieved documents
type ContextBuilder interface {
	Build(ctx context.Context, query string, results []*SearchResult) (string, error)
}

// SimpleContextBuilder implements basic context building
type SimpleContextBuilder struct {
	maxTokens   int
	separator   string
	includeMetadata bool
}

// NewSimpleContextBuilder creates a new simple context builder
func NewSimpleContextBuilder(maxTokens int) *SimpleContextBuilder {
	return &SimpleContextBuilder{
		maxTokens:       maxTokens,
		separator:       "\n\n---\n\n",
		includeMetadata: true,
	}
}

// Build implements ContextBuilder
func (b *SimpleContextBuilder) Build(ctx context.Context, query string, results []*SearchResult) (string, error) {
	var parts []string
	tokenCount := 0

	for _, result := range results {
		docContent := result.Document.Content
		docTokens := len(docContent) / 4 // Rough token estimate

		if tokenCount+docTokens > b.maxTokens {
			break
		}

		part := docContent
		if b.includeMetadata {
			if result.Document.Title != "" {
				part = fmt.Sprintf("Title: %s\n\n%s", result.Document.Title, part)
			}
			if result.Document.Source != "" {
				part = fmt.Sprintf("%s\n\nSource: %s", part, result.Document.Source)
			}
		}

		parts = append(parts, part)
		tokenCount += docTokens
	}

	return strings.Join(parts, b.separator), nil
}

// NewQdrantRAGPipeline creates a new RAG pipeline
func NewQdrantRAGPipeline(
	client *qdrant.Client,
	collection string,
	embedder Embedder,
	reranker Reranker,
	logger *logrus.Logger,
) *QdrantRAGPipeline {
	if logger == nil {
		logger = logrus.New()
	}

	return &QdrantRAGPipeline{
		retriever:      NewEnhancedQdrantRetriever(client, collection, embedder, reranker, logger),
		documentStore:  NewQdrantDocumentStore(client, collection, embedder, logger),
		contextBuilder: NewSimpleContextBuilder(4000),
		logger:         logger,
	}
}

// SetContextBuilder sets a custom context builder
func (p *QdrantRAGPipeline) SetContextBuilder(builder ContextBuilder) {
	p.contextBuilder = builder
}

// SetDebateEvaluator sets the debate evaluator for the retriever
func (p *QdrantRAGPipeline) SetDebateEvaluator(evaluator DebateEvaluator) {
	p.retriever.SetDebateEvaluator(evaluator)
}

// AddDocuments adds documents to the pipeline
func (p *QdrantRAGPipeline) AddDocuments(ctx context.Context, docs []*Document) error {
	// Add to vector store
	if err := p.documentStore.AddDocuments(ctx, docs); err != nil {
		return fmt.Errorf("failed to add to vector store: %w", err)
	}

	// Index for sparse retrieval
	if err := p.retriever.IndexDocuments(ctx, docs); err != nil {
		return fmt.Errorf("failed to index documents: %w", err)
	}

	return nil
}

// Query performs retrieval and builds context
func (p *QdrantRAGPipeline) Query(ctx context.Context, query string, opts *SearchOptions) (string, []*SearchResult, error) {
	// Retrieve relevant documents
	results, err := p.retriever.HybridRetrieve(ctx, query, opts)
	if err != nil {
		return "", nil, fmt.Errorf("retrieval failed: %w", err)
	}

	// Build context from results
	context, err := p.contextBuilder.Build(ctx, query, results)
	if err != nil {
		return "", results, fmt.Errorf("context building failed: %w", err)
	}

	return context, results, nil
}

// GetRetriever returns the underlying retriever
func (p *QdrantRAGPipeline) GetRetriever() *EnhancedQdrantRetriever {
	return p.retriever
}

// GetDocumentStore returns the underlying document store
func (p *QdrantRAGPipeline) GetDocumentStore() *QdrantDocumentStore {
	return p.documentStore
}
