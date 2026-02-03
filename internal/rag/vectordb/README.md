# RAG Vector DB Abstraction

This package documents the vector database abstraction layer used by the RAG system for storing and retrieving document embeddings.

## Overview

The RAG vector database abstraction provides a unified interface for multiple vector database backends, enabling semantic search with configurable caching and optimization strategies. The implementation supports Chroma, Qdrant, and Weaviate with automatic collection management.

## Architecture

```
                    +------------------+
                    |   RAG Pipeline   |
                    +--------+---------+
                             |
              +--------------+--------------+
              |                             |
              v                             v
     +--------+--------+          +--------+--------+
     | EmbeddingRegistry|         | Vector Store    |
     | (Encode Text)    |         | Abstraction     |
     +-----------------+          +--------+--------+
                                           |
              +----------------------------+----------------------------+
              |                            |                            |
              v                            v                            v
     +--------+--------+         +--------+--------+          +--------+--------+
     | ChromaAdapter   |         | QdrantAdapter   |          | WeaviateAdapter |
     | (Chroma DB)     |         | (Qdrant DB)     |          | (Weaviate DB)   |
     +-----------------+         +-----------------+          +-----------------+
```

## Supported Vector Databases

### Chroma

Lightweight, in-memory or persistent vector database.

```go
config := &servers.ChromaAdapterConfig{
    Host:       "localhost",
    Port:       8000,
    Collection: "rag_documents",
}

adapter := servers.NewChromaAdapter(*config)
err := adapter.Connect(ctx)
```

### Qdrant

High-performance vector similarity search engine.

```go
config := &servers.QdrantAdapterConfig{
    Host:       "localhost",
    Port:       6333,
    Collection: "rag_documents",
    APIKey:     "", // Optional
}

adapter := servers.NewQdrantAdapter(*config)
err := adapter.Connect(ctx)
```

### Weaviate

Vector search engine with GraphQL API.

```go
config := &servers.WeaviateAdapterConfig{
    Host:   "localhost",
    Port:   8080,
    Scheme: "http",
    Class:  "RAGDocument",
}

adapter := servers.NewWeaviateAdapter(*config)
err := adapter.Connect(ctx)
```

## Vector Store Interface

### Pipeline Configuration

```go
type PipelineConfig struct {
    VectorDBType   VectorDBType                   `json:"vector_db_type"`
    CollectionName string                         `json:"collection_name"`
    EmbeddingModel string                         `json:"embedding_model"`
    ChunkingConfig ChunkingConfig                 `json:"chunking_config"`
    ChromaConfig   *servers.ChromaAdapterConfig   `json:"chroma_config"`
    QdrantConfig   *servers.QdrantAdapterConfig   `json:"qdrant_config"`
    WeaviateConfig *servers.WeaviateAdapterConfig `json:"weaviate_config"`
    EnableCache    bool                           `json:"enable_cache"`
    CacheTTL       time.Duration                  `json:"cache_ttl"`
}
```

### Vector Database Types

```go
const (
    VectorDBChroma   VectorDBType = "chroma"
    VectorDBQdrant   VectorDBType = "qdrant"
    VectorDBWeaviate VectorDBType = "weaviate"
)
```

## Search Optimization

### Index Configuration

The pipeline automatically creates collections with appropriate settings:

**Qdrant**:
```go
// Create collection with cosine similarity
err = adapter.CreateCollection(ctx, collectionName, uint64(dim), "Cosine")
```

**Weaviate**:
```go
class := &servers.WeaviateClass{
    Class:       collectionName,
    Description: "RAG collection",
    Properties: []servers.WeaviateProperty{
        {Name: "content", DataType: []string{"text"}},
        {Name: "doc_id", DataType: []string{"text"}},
        {Name: "chunk_id", DataType: []string{"text"}},
        {Name: "start_idx", DataType: []string{"int"}},
        {Name: "end_idx", DataType: []string{"int"}},
    },
}
err = adapter.CreateClass(ctx, class)
```

### Search Parameters

**Qdrant Search**:
```go
qdrantOpts := &qdrant.SearchOptions{
    Limit:          opts.TopK,
    ScoreThreshold: float32(opts.MinScore),
    WithPayload:    true,
    WithVectors:    false,
}
results, err := client.Search(ctx, collection, vector, qdrantOpts)
```

**Chroma Query**:
```go
result, err := adapter.Query(ctx, collectionName, [][]float32{queryEmbedding}, topK, filter)
```

**Weaviate Vector Search**:
```go
results, err := adapter.VectorSearch(ctx, className, queryEmbedding, topK, minCertainty, filter)
```

### Filtering

All backends support metadata filtering:

```go
filter := map[string]interface{}{
    "category": "documentation",
    "version":  "v2",
}

results, err := pipeline.SearchWithFilter(ctx, query, topK, filter)
```

## Caching Strategy

### Embedding Cache

The pipeline can cache embeddings to reduce API calls:

```go
config := PipelineConfig{
    EnableCache: true,
    CacheTTL:    time.Hour,
}
```

### Query Result Cache

For frequently executed queries, results can be cached at the application level using Redis or in-memory caching.

### Qdrant-Specific Caching

Qdrant supports internal caching of vectors:

```go
type QdrantEnhancedConfig struct {
    DenseWeight         float64      `json:"dense_weight"`
    SparseWeight        float64      `json:"sparse_weight"`
    UseDebateEvaluation bool         `json:"use_debate_evaluation"`
    DebateTopK          int          `json:"debate_top_k"`
    FusionMethod        FusionMethod `json:"fusion_method"`
    RRFK                float64      `json:"rrf_k"`
}
```

## Integration Patterns

### Document Store Pattern

```go
store := NewQdrantDocumentStore(client, collection, embedder, logger)

// Ensure collection exists
err := store.EnsureCollection(ctx, vectorSize)

// Add documents
err = store.AddDocuments(ctx, []*Document{
    {ID: "doc1", Content: "...", Metadata: map[string]interface{}{"source": "api"}},
    {ID: "doc2", Content: "...", Metadata: map[string]interface{}{"source": "web"}},
})

// Retrieve document
doc, err := store.GetDocument(ctx, "doc1")

// Delete document
err = store.DeleteDocument(ctx, "doc1")
```

### Dense Retriever Pattern

```go
retriever := NewQdrantDenseRetriever(client, collection, embedder, logger)

results, err := retriever.Retrieve(ctx, query, &SearchOptions{
    TopK:            10,
    MinScore:        0.5,
    IncludeMetadata: true,
})
```

### Enhanced Retriever Pattern

Combines Qdrant with BM25 for hybrid search:

```go
retriever := NewQdrantEnhancedRetriever(
    denseRetriever,
    reranker,
    &QdrantEnhancedConfig{
        DenseWeight:  0.6,
        SparseWeight: 0.4,
        FusionMethod: FusionRRF,
        RRFK:         60.0,
    },
    logger,
)

// Index documents (adds to both dense and sparse indices)
err := retriever.Index(ctx, docs)

// Hybrid search
results, err := retriever.Retrieve(ctx, query, opts)
```

## Data Types

### Pipeline Document

```go
type PipelineDocument struct {
    ID        string                 `json:"id"`
    Content   string                 `json:"content"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
    Source    string                 `json:"source,omitempty"`
    Chunks    []PipelineChunk        `json:"chunks,omitempty"`
    CreatedAt time.Time              `json:"created_at"`
    UpdatedAt time.Time              `json:"updated_at"`
}
```

### Pipeline Chunk

```go
type PipelineChunk struct {
    ID        string                 `json:"id"`
    Content   string                 `json:"content"`
    Embedding []float32              `json:"embedding,omitempty"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
    StartIdx  int                    `json:"start_idx"`
    EndIdx    int                    `json:"end_idx"`
    DocID     string                 `json:"doc_id"`
}
```

### Search Result

```go
type PipelineSearchResult struct {
    Chunk    PipelineChunk          `json:"chunk"`
    Score    float32                `json:"score"`
    Distance float32                `json:"distance"`
    Metadata map[string]interface{} `json:"metadata,omitempty"`
}
```

## Usage Examples

### Initialize Pipeline

```go
config := PipelineConfig{
    VectorDBType:   VectorDBQdrant,
    CollectionName: "my_documents",
    EmbeddingModel: "text-embedding-3-small",
    ChunkingConfig: ChunkingConfig{
        ChunkSize:    512,
        ChunkOverlap: 50,
        Separator:    "\n\n",
    },
    QdrantConfig: &servers.QdrantAdapterConfig{
        Host: "localhost",
        Port: 6333,
    },
}

pipeline := NewPipeline(config, embeddingRegistry)
err := pipeline.Initialize(ctx)
defer pipeline.Close()
```

### Ingest Documents

```go
doc := &PipelineDocument{
    ID:      "doc-001",
    Content: "This is the document content...",
    Metadata: map[string]interface{}{
        "source":   "upload",
        "category": "technical",
    },
}

err := pipeline.IngestDocument(ctx, doc)
```

### Search Documents

```go
results, err := pipeline.Search(ctx, "How to configure authentication?", 10)

for _, result := range results {
    fmt.Printf("Score: %.4f | Content: %s\n", result.Score, result.Chunk.Content[:100])
}
```

### Get Pipeline Stats

```go
stats, err := pipeline.GetStats(ctx)
// Returns: vector_db_type, collection_name, document_count, etc.
```

## Chunking Configuration

The pipeline automatically chunks documents:

```go
type ChunkingConfig struct {
    ChunkSize    int    `json:"chunk_size"`    // Default: 512
    ChunkOverlap int    `json:"chunk_overlap"` // Default: 50
    Separator    string `json:"separator"`     // Default: "\n\n"
}
```

Chunking algorithm:
1. Split by separator
2. Accumulate parts until chunk size exceeded
3. Create chunk with overlap from previous
4. Continue until all content processed

## Health Checks

```go
// Check pipeline health
err := pipeline.Health(ctx)
if err != nil {
    log.Printf("Pipeline unhealthy: %v", err)
}

// Check specific adapter
switch config.VectorDBType {
case VectorDBQdrant:
    err = adapter.Health(ctx)
case VectorDBChroma:
    err = adapter.Health(ctx)
case VectorDBWeaviate:
    err = adapter.Health(ctx)
}
```

## Performance Tuning

### Batch Operations

```go
// Batch ingest for better performance
docs := []*PipelineDocument{doc1, doc2, doc3}
err := pipeline.IngestDocuments(ctx, docs)
```

### Connection Pooling

Vector database clients maintain connection pools internally:
- Qdrant: gRPC connection with keepalive
- Chroma: HTTP client with connection reuse
- Weaviate: GraphQL client with connection pooling

### Index Optimization

- Use appropriate vector dimensions (1536 for OpenAI, 768 for smaller models)
- Enable HNSW index for Qdrant (default)
- Configure ef_construction and m parameters for recall vs. speed tradeoff

## Testing

```bash
# Run vector DB integration tests
go test -v ./internal/rag/... -run TestQdrant
go test -v ./internal/rag/... -run TestPipeline

# With infrastructure
make test-infra-start
go test -v ./internal/rag/... -tags=integration
make test-infra-stop
```

## Related Packages

- `internal/rag/` - Main RAG pipeline
- `internal/vectordb/` - Low-level vector DB clients
- `internal/mcp/servers/` - MCP adapters for vector DBs
- `internal/embedding/` - Embedding providers
