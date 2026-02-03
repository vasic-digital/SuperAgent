# RAG Retrieval Strategies

This package documents the retrieval strategies implemented in the parent `rag` package for Retrieval-Augmented Generation pipelines.

## Overview

The RAG retrieval system provides multiple strategies for document retrieval, combining semantic search with keyword matching to deliver highly relevant results. The implementation supports multi-stage retrieval pipelines with configurable ranking and scoring mechanisms.

## Architecture

```
                         +------------------+
                         |   Query Input    |
                         +--------+---------+
                                  |
                                  v
                    +-------------+-------------+
                    |    Query Expansion        |
                    | (Synonyms, LLM, Hyponyms) |
                    +-------------+-------------+
                                  |
           +----------------------+----------------------+
           |                                             |
           v                                             v
+----------+-----------+                    +------------+-----------+
|   Dense Retriever    |                    |   Sparse Retriever     |
| (Semantic/Embedding) |                    |   (BM25/Keyword)       |
+----------+-----------+                    +------------+-----------+
           |                                             |
           +----------------------+----------------------+
                                  |
                                  v
                    +-------------+-------------+
                    |      Result Fusion        |
                    |  (RRF/Weighted/Max)       |
                    +-------------+-------------+
                                  |
                                  v
                    +-------------+-------------+
                    |        Reranker           |
                    |  (Cross-Encoder/Cohere)   |
                    +-------------+-------------+
                                  |
                                  v
                    +-------------+-------------+
                    | Contextual Compression    |
                    +-------------+-------------+
                                  |
                                  v
                         +-------+--------+
                         | Final Results  |
                         +----------------+
```

## Retrieval Strategies

### Dense Retrieval (Semantic Search)

Dense retrieval uses embedding vectors to find semantically similar documents.

```go
type DenseRetriever interface {
    Retriever
    Embed(ctx context.Context, texts []string) ([][]float32, error)
}
```

**Implementation**: `QdrantDenseRetriever` in `qdrant_retriever.go`

```go
retriever := NewQdrantDenseRetriever(client, collection, embedder, logger)
results, err := retriever.Retrieve(ctx, query, &SearchOptions{
    TopK:     10,
    MinScore: 0.5,
})
```

### Sparse Retrieval (Keyword Search)

Sparse retrieval uses BM25 (Best Matching 25) algorithm for keyword-based search.

```go
type SparseRetriever interface {
    Retriever
    GetTermFrequencies(ctx context.Context, docID string) (map[string]float64, error)
}
```

**Implementation**: `EnhancedBM25Index` in `qdrant_enhanced.go`

```go
index := NewEnhancedBM25Index()
index.AddDocument(id, content)
results := index.Search(query, topK)
```

BM25 Parameters:
- `k1`: 1.2 (term frequency saturation)
- `b`: 0.75 (document length normalization)

### Hybrid Retrieval

Combines dense and sparse retrieval for best results.

```go
retriever := NewHybridRetriever(
    denseRetriever,
    sparseRetriever,
    reranker,
    &HybridConfig{
        Alpha:                 0.5,  // 0=sparse only, 1=dense only
        FusionMethod:          FusionRRF,
        RRFK:                  60,
        EnableReranking:       true,
        RerankTopK:            50,
        PreRetrieveMultiplier: 3,
    },
    logger,
)

results, err := retriever.Retrieve(ctx, query, opts)
```

## Multi-Stage Retrieval Pipeline

### Stage 1: Query Expansion

Expands the original query using synonyms and related terms.

```go
config := QueryExpansionConfig{
    MaxExpansions:      5,
    EnableSynonyms:     true,
    EnableHyponyms:     false,
    EnableHypernyms:    false,
    EnableLLMExpansion: false,
    SynonymWeight:      0.8,
}

expansions := advancedRAG.ExpandQuery(ctx, query)
// Returns: [{Query: "original", Weight: 1.0}, {Query: "expanded", Weight: 0.8}]
```

### Stage 2: Parallel Retrieval

Dense and sparse retrieval run concurrently.

```go
var wg sync.WaitGroup
var denseResults, sparseResults []*SearchResult

wg.Add(2)
go func() {
    defer wg.Done()
    denseResults, _ = denseRetriever.Retrieve(ctx, query, opts)
}()
go func() {
    defer wg.Done()
    sparseResults, _ = sparseRetriever.Retrieve(ctx, query, opts)
}()
wg.Wait()
```

### Stage 3: Result Fusion

Three fusion methods available:

#### Reciprocal Rank Fusion (RRF)
```
RRF(d) = sum(1 / (k + rank(d)))
```
Best for combining results from different scoring scales.

#### Weighted Fusion
```
Score = alpha * normalize(dense_score) + (1-alpha) * normalize(sparse_score)
```
Allows explicit control over dense vs. sparse weighting.

#### Max Fusion
```
Score = max(dense_score, sparse_score)
```
Simple approach when one method is expected to dominate.

### Stage 4: Reranking

Cross-encoder models re-score results for better relevance.

```go
reranker := NewCrossEncoderReranker(&RerankerConfig{
    Model:        "BAAI/bge-reranker-v2-m3",
    Endpoint:     "http://localhost:8080/rerank",
    Timeout:      30 * time.Second,
    BatchSize:    32,
    ReturnScores: true,
}, logger)

reranked, err := reranker.Rerank(ctx, query, results, topK)
```

**Supported Rerankers**:
- `CrossEncoderReranker`: Generic cross-encoder API
- `CohereReranker`: Cohere's rerank API

### Stage 5: Contextual Compression

Reduces context size while preserving relevant information.

```go
config := ContextualCompressionConfig{
    MaxContextLength:         4096,
    CompressionRatio:         0.5,
    EnableSentenceExtraction: true,
    EnableSummarization:      false,
    PreserveKeyPhrases:       true,
}

compressed, err := advancedRAG.CompressContext(ctx, query, results)
// Returns: {Content: "...", KeyPhrases: [...], CompressionRatio: 0.5}
```

## Ranking and Scoring

### Score Types

| Type | Range | Description |
|------|-------|-------------|
| Dense Score | 0-1 | Cosine similarity from embeddings |
| Sparse Score | 0+ | BM25 score (unbounded) |
| Fused Score | 0+ | Combined score after fusion |
| Reranked Score | 0-1 | Cross-encoder relevance |

### Match Types

```go
const (
    MatchTypeDense  MatchType = "dense"   // Semantic match
    MatchTypeSparse MatchType = "sparse"  // Keyword match
    MatchTypeHybrid MatchType = "hybrid"  // Combined match
)
```

### Relevance Scoring

The `calculateRelevanceScore` function computes term overlap with frequency bonuses:

```go
score = sum(1.0 + log1p(freq-1) * 0.1)  // for exact matches
score += 0.5 + log1p(freq-1) * 0.05     // for partial matches
normalized = score / (maxScore * 1.5)
```

## Configuration Options

### Search Options

```go
type SearchOptions struct {
    TopK            int                    `json:"top_k"`
    MinScore        float64                `json:"min_score"`
    Filter          map[string]interface{} `json:"filter"`
    EnableReranking bool                   `json:"enable_reranking"`
    HybridAlpha     float64                `json:"hybrid_alpha"`
    IncludeMetadata bool                   `json:"include_metadata"`
    Namespace       string                 `json:"namespace"`
}
```

### Hybrid Search Config

```go
type HybridSearchConfig struct {
    VectorWeight     float64 `json:"vector_weight"`     // Default: 0.7
    KeywordWeight    float64 `json:"keyword_weight"`    // Default: 0.3
    MinKeywordScore  float64 `json:"min_keyword_score"` // Default: 0.1
    EnableFuzzyMatch bool    `json:"enable_fuzzy_match"`// Default: true
    FuzzyThreshold   float64 `json:"fuzzy_threshold"`   // Default: 0.8
}
```

### ReRanker Config

```go
type ReRankerConfig struct {
    Model              string  `json:"model"`
    TopK               int     `json:"top_k"`               // Default: 100
    BatchSize          int     `json:"batch_size"`          // Default: 32
    ScoreThreshold     float64 `json:"score_threshold"`     // Default: 0.5
    EnableCrossEncoder bool    `json:"enable_cross_encoder"`// Default: true
}
```

## Advanced Retrieval Patterns

### Multi-Hop Search

Follows document relationships for deep exploration.

```go
config := &MultiHopConfig{
    MaxHops:           3,
    MinRelevanceScore: 0.3,
    MaxResultsPerHop:  10,
    DecayFactor:       0.8,
    EnableBacklinks:   true,
}

results, err := advancedRAG.MultiHopSearch(ctx, query, topK, config)
```

### Iterative Search

Refines queries across multiple iterations.

```go
config := &IterativeRetrievalConfig{
    MaxIterations:         5,
    ConvergenceThreshold:  0.05,
    ResultsPerIteration:   20,
    FeedbackWeight:        0.3,
    EnableQueryRefinement: true,
}

results, metrics, err := advancedRAG.IterativeSearch(ctx, query, topK, config)
```

### Hierarchical Retrieval

Retrieves documents with parent-child relationships.

```go
retriever := NewHierarchicalRetriever(baseRetriever)
results, err := retriever.RetrieveWithChildren(ctx, parentID, opts)
```

### Temporal Retrieval

Time-aware retrieval with recency weighting.

```go
retriever := NewTemporalRetriever(baseRetriever, 0.001) // decay factor
results, err := retriever.RetrieveByDateRange(ctx, startTime, endTime, opts)
```

## Usage Examples

### Basic Hybrid Search

```go
pipeline := NewPipeline(PipelineConfig{
    VectorDBType:   VectorDBQdrant,
    CollectionName: "documents",
    EmbeddingModel: "text-embedding-3-small",
}, embeddingRegistry)

err := pipeline.Initialize(ctx)
results, err := pipeline.Search(ctx, "How to handle errors?", 10)
```

### Advanced RAG with All Stages

```go
advancedRAG := NewAdvancedRAG(DefaultAdvancedRAGConfig(), pipeline)
err := advancedRAG.Initialize(ctx)

// Hybrid search with reranking
hybridResults, err := advancedRAG.HybridSearch(ctx, query, 20)

// Re-rank for precision
reranked, err := advancedRAG.ReRank(ctx, query, pipelineResults)

// Compress for LLM context
compressed, err := advancedRAG.CompressContext(ctx, query, reranked)
```

## Performance Considerations

- **Pre-retrieval Multiplier**: Retrieve 3x TopK before fusion for better coverage
- **Batch Reranking**: Process in batches of 32 for optimal throughput
- **Parallel Retrieval**: Dense and sparse run concurrently
- **Early Stopping**: Iterative search stops when improvement < threshold

## Testing

```bash
# Run all RAG tests
go test -v ./internal/rag/...

# Run retrieval-specific tests
go test -v -run TestHybrid ./internal/rag/
go test -v -run TestReRank ./internal/rag/
```

## Related Packages

- `internal/rag/` - Main RAG pipeline
- `internal/vectordb/` - Vector database clients
- `internal/embedding/` - Embedding providers
- `internal/mcp/servers/` - Vector DB MCP adapters
