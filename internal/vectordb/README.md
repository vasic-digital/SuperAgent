# VectorDB Package

The vectordb package provides vector database integrations for HelixAgent.

## Overview

Supports multiple vector databases:
- Chroma
- Qdrant
- Weaviate
- Pinecone

## Key Components

```go
db := vectordb.NewChromaClient(config)

// Store vectors
err := db.Upsert(ctx, documents)

// Search
results, err := db.Search(ctx, queryVector, topK)
```

## See Also

- `internal/embedding/` - Vector embeddings
- `internal/rag/` - RAG pipeline
