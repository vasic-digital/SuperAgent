# RAG (Retrieval Augmented Generation) Package

The RAG package provides Retrieval Augmented Generation capabilities for HelixAgent.

## Overview

Implements:
- Document retrieval from vector databases
- Context augmentation for LLM queries
- Hybrid search (semantic + keyword)
- Re-ranking for relevance

## Key Components

```go
rag := NewRAGPipeline(config, vectorDB, embedder)

// Query with context augmentation
response, err := rag.Query(ctx, &RAGRequest{
    Query:     "How do I implement error handling?",
    TopK:      5,
    Threshold: 0.7,
})
```

## See Also

- `internal/vectordb/` - Vector database integration
- `internal/embedding/` - Text embeddings
