# RAG (Retrieval Augmented Generation) System

## Overview

The RAG System in HelixAgent provides advanced retrieval-augmented generation capabilities, combining dense (embedding-based) and sparse (keyword-based) retrieval methods with intelligent reranking for optimal context retrieval.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        RAG Pipeline                              │
│                                                                  │
│  ┌──────────┐    ┌──────────────────┐    ┌─────────────────┐  │
│  │  Query   │───▶│  Hybrid Retriever │───▶│    Reranker     │  │
│  └──────────┘    │  ├─ Dense (Vector)│    │  (Cross-Encoder)│  │
│                  │  └─ Sparse (BM25) │    └────────┬────────┘  │
│                  └──────────────────┘              │            │
│                                                     ▼            │
│  ┌──────────┐    ┌──────────────────┐    ┌─────────────────┐  │
│  │ Response │◀───│  LLM Generation  │◀───│ Context Builder │  │
│  └──────────┘    └──────────────────┘    └─────────────────┘  │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    Vector Store (Qdrant)                  │  │
│  │  ├─ Document embeddings                                   │  │
│  │  ├─ Metadata filtering                                    │  │
│  │  └─ Hybrid search support                                 │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## Components

### 1. Hybrid Retriever (`internal/rag/retriever.go`)

Combines dense and sparse retrieval for better results.

```go
import "dev.helix.agent/internal/rag"

// Create hybrid retriever
retriever := rag.NewHybridRetriever(&rag.HybridConfig{
    DenseWeight:    0.7,  // Weight for vector similarity
    SparseWeight:   0.3,  // Weight for BM25 scoring
    TopK:           20,   // Initial retrieval count
    FinalK:         5,    // Final reranked count
    VectorStore:    qdrantClient,
    BM25Index:      bm25Index,
})

// Retrieve relevant documents
docs, err := retriever.Retrieve(ctx, "How do I configure rate limiting?")
```

### 2. Document Processor (`internal/rag/processor.go`)

Handles document chunking and embedding generation.

```go
processor := rag.NewDocumentProcessor(&rag.ProcessorConfig{
    ChunkSize:     512,    // Tokens per chunk
    ChunkOverlap:  50,     // Overlap between chunks
    Embedder:      embeddingProvider,
    Tokenizer:     tokenizer,
})

// Process a document
chunks, err := processor.Process(ctx, document)
for _, chunk := range chunks {
    fmt.Printf("Chunk %d: %d tokens\n", chunk.Index, chunk.TokenCount)
}
```

### 3. Reranker (`internal/rag/reranker.go`)

Reorders retrieved documents by relevance using cross-encoder models.

```go
reranker := rag.NewReranker(&rag.RerankerConfig{
    Model:        "cross-encoder/ms-marco-MiniLM-L-12-v2",
    MaxLength:    512,
    BatchSize:    32,
})

// Rerank documents
rerankedDocs, err := reranker.Rerank(ctx, query, retrievedDocs, 5)
```

### 4. Context Builder (`internal/rag/context.go`)

Assembles retrieved context for LLM generation.

```go
builder := rag.NewContextBuilder(&rag.ContextConfig{
    MaxTokens:        4000,
    IncludeMetadata:  true,
    CitationStyle:    "numbered",
})

// Build context from documents
context, citations := builder.Build(ctx, docs, query)
```

## Retrieval Methods

### Dense Retrieval (Vector Search)

Uses embedding similarity for semantic matching:

```go
// Dense-only retrieval
docs, err := retriever.DenseRetrieve(ctx, query, &rag.DenseOptions{
    TopK:           10,
    ScoreThreshold: 0.7,
    Filter: map[string]interface{}{
        "source": "documentation",
    },
})
```

### Sparse Retrieval (BM25)

Uses keyword matching for exact term matching:

```go
// Sparse-only retrieval
docs, err := retriever.SparseRetrieve(ctx, query, &rag.SparseOptions{
    TopK:        10,
    B:           0.75,  // BM25 b parameter
    K1:          1.2,   // BM25 k1 parameter
})
```

### Hybrid Retrieval

Combines both methods using Reciprocal Rank Fusion (RRF):

```go
// Hybrid retrieval with RRF
docs, err := retriever.HybridRetrieve(ctx, query, &rag.HybridOptions{
    DenseTopK:   20,
    SparseTopK:  20,
    FinalTopK:   10,
    RRFConstant: 60,  // RRF k parameter
})
```

## Document Ingestion

### Single Document

```go
// Ingest a single document
err := pipeline.Ingest(ctx, &rag.Document{
    ID:       "doc-001",
    Content:  documentContent,
    Metadata: map[string]interface{}{
        "source":    "api-docs",
        "version":   "1.0",
        "timestamp": time.Now(),
    },
})
```

### Batch Ingestion

```go
// Ingest multiple documents
docs := []*rag.Document{doc1, doc2, doc3}
results, err := pipeline.IngestBatch(ctx, docs, &rag.IngestOptions{
    BatchSize:   100,
    Parallel:    4,
    OnProgress:  progressCallback,
})
```

### File Ingestion

```go
// Ingest from files
err := pipeline.IngestFile(ctx, "/path/to/document.md", &rag.FileOptions{
    Parser:    "markdown",
    ChunkBy:   "heading",
})

// Ingest directory
err := pipeline.IngestDirectory(ctx, "/path/to/docs", &rag.DirectoryOptions{
    Recursive:  true,
    Extensions: []string{".md", ".txt", ".go"},
    Ignore:     []string{"*_test.go", "vendor/**"},
})
```

## Query Pipeline

### Complete RAG Query

```go
pipeline := rag.NewPipeline(&rag.PipelineConfig{
    Retriever:      retriever,
    Reranker:       reranker,
    ContextBuilder: builder,
    LLMProvider:    llmProvider,
})

// Execute RAG query
response, err := pipeline.Query(ctx, &rag.Query{
    Text:           "How do I implement rate limiting?",
    TopK:           5,
    IncludeSources: true,
})

fmt.Println(response.Answer)
for _, source := range response.Sources {
    fmt.Printf("- %s (score: %.2f)\n", source.Title, source.Score)
}
```

### With Streaming

```go
stream, err := pipeline.QueryStream(ctx, query)
for chunk := range stream.Chunks() {
    fmt.Print(chunk.Text)
}
fmt.Println("\n\nSources:", stream.Sources())
```

## Configuration

```go
type Config struct {
    // Vector store configuration
    VectorStore struct {
        Type       string // "qdrant", "pinecone", "weaviate"
        URL        string
        Collection string
        Dimension  int    // Embedding dimension (e.g., 1536)
    }

    // Embedding configuration
    Embedding struct {
        Provider string // "openai", "cohere", "local"
        Model    string // e.g., "text-embedding-3-small"
    }

    // Chunking configuration
    Chunking struct {
        Strategy    string // "fixed", "semantic", "recursive"
        ChunkSize   int    // Target chunk size in tokens
        ChunkOverlap int   // Overlap between chunks
    }

    // Retrieval configuration
    Retrieval struct {
        Method     string  // "dense", "sparse", "hybrid"
        DenseWeight float64
        SparseWeight float64
        TopK       int
    }

    // Reranking configuration
    Reranking struct {
        Enabled   bool
        Model     string
        TopK      int
    }
}
```

## Evaluation Metrics

The RAG system supports RAGAS-style evaluation:

```go
evaluator := rag.NewEvaluator(&rag.EvaluatorConfig{
    Metrics: []string{
        "faithfulness",     // Answer grounded in context
        "answer_relevancy", // Answer relevant to question
        "context_precision",// Retrieved context precision
        "context_recall",   // Retrieved context recall
    },
})

results, err := evaluator.Evaluate(ctx, &rag.EvalDataset{
    Questions: questions,
    GroundTruth: groundTruth,
})

fmt.Printf("Faithfulness: %.2f\n", results.Faithfulness)
fmt.Printf("Answer Relevancy: %.2f\n", results.AnswerRelevancy)
```

## Integration with Qdrant

### Collection Setup

```go
// Create collection for RAG
err := qdrantClient.CreateCollection(ctx, &qdrant.CollectionConfig{
    Name:       "documents",
    VectorSize: 1536,
    Distance:   qdrant.DistanceCosine,
    OnDiskPayload: true,
})
```

### Search with Filters

```go
results, err := qdrantClient.Search(ctx, "documents", query, &qdrant.SearchOptions{
    TopK: 10,
    Filter: &qdrant.Filter{
        Must: []qdrant.Condition{
            {Key: "source", Match: qdrant.MatchValue{Value: "docs"}},
            {Key: "version", Range: qdrant.RangeCondition{Gte: 2.0}},
        },
    },
    WithPayload: true,
    WithVectors: false,
})
```

## Testing

```bash
# Run RAG tests
go test -v ./internal/rag/...

# Run with coverage
go test -cover ./internal/rag/...

# Run integration tests (requires Qdrant)
go test -v -tags=integration ./internal/rag/...
```

## Key Files

| File | Description |
|------|-------------|
| `internal/rag/retriever.go` | Hybrid retrieval implementation |
| `internal/rag/processor.go` | Document processing and chunking |
| `internal/rag/reranker.go` | Cross-encoder reranking |
| `internal/rag/context.go` | Context building for LLM |
| `internal/rag/pipeline.go` | Complete RAG pipeline |
| `internal/rag/eval.go` | Evaluation metrics |
| `internal/vectordb/qdrant/` | Qdrant client integration |

## See Also

- [Vector Database Integration](./VECTOR_DATABASE.md)
- [Memory System](./MEMORY_SYSTEM.md)
- [LLM Testing Framework](./LLM_TESTING.md)
