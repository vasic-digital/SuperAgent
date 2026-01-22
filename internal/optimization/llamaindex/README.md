# LlamaIndex Package

This package provides an HTTP client for the LlamaIndex service, enabling document indexing, retrieval, and RAG (Retrieval-Augmented Generation) workflows.

## Overview

LlamaIndex provides powerful document indexing and retrieval capabilities, supporting various index types, query transformations, and integration with Cognee for knowledge graph queries.

## Features

- **Document Indexing**: Index documents for retrieval
- **Semantic Search**: Vector-based document retrieval
- **Query Transformation**: Automatic query rewriting
- **Reranking**: Result reranking for relevance
- **Cognee Integration**: Knowledge graph queries

## Components

### Client (`client.go`)

HTTP client for LlamaIndex service:

```go
config := &llamaindex.ClientConfig{
    BaseURL: "http://localhost:8012",
    Timeout: 120 * time.Second,
}

client := llamaindex.NewClient(config)
```

## Data Types

### ClientConfig

```go
type ClientConfig struct {
    BaseURL string        // LlamaIndex server URL
    Timeout time.Duration // Request timeout
}
```

### QueryRequest

```go
type QueryRequest struct {
    Query          string                 // Search query
    TopK           int                    // Number of results
    UseCognee      bool                   // Use Cognee knowledge graph
    Rerank         bool                   // Enable reranking
    QueryTransform *string                // Query transformation type
    Filters        map[string]interface{} // Metadata filters
}
```

### Source

```go
type Source struct {
    Content  string                 // Document content
    Score    float64                // Relevance score
    Metadata map[string]interface{} // Document metadata
}
```

## Usage

### Basic Document Query

```go
import "dev.helix.agent/internal/optimization/llamaindex"

client := llamaindex.NewClient(nil)

results, err := client.Query(ctx, &llamaindex.QueryRequest{
    Query: "How do I configure authentication?",
    TopK:  5,
})

for _, source := range results.Sources {
    fmt.Printf("Score: %.2f - %s\n", source.Score, source.Content[:100])
}
```

### With Reranking

```go
results, _ := client.Query(ctx, &llamaindex.QueryRequest{
    Query:  "What is the deployment process?",
    TopK:   10,
    Rerank: true, // Rerank results for better relevance
})
```

### Query Transformation

```go
// Automatically expand/rewrite query for better retrieval
transform := "hyde" // Hypothetical Document Embeddings
results, _ := client.Query(ctx, &llamaindex.QueryRequest{
    Query:          "authentication setup",
    TopK:           5,
    QueryTransform: &transform,
})
```

### With Cognee Knowledge Graph

```go
results, _ := client.Query(ctx, &llamaindex.QueryRequest{
    Query:     "What entities are related to user authentication?",
    TopK:      5,
    UseCognee: true, // Query knowledge graph
})
```

### Metadata Filtering

```go
results, _ := client.Query(ctx, &llamaindex.QueryRequest{
    Query: "API endpoints",
    TopK:  10,
    Filters: map[string]interface{}{
        "document_type": "api_docs",
        "version":       "v2",
    },
})
```

### Document Indexing

```go
err := client.Index(ctx, &llamaindex.IndexRequest{
    Documents: []llamaindex.Document{
        {
            ID:      "doc-1",
            Content: "This is the document content...",
            Metadata: map[string]interface{}{
                "title":  "Getting Started",
                "author": "HelixAgent Team",
            },
        },
    },
    IndexName: "documentation",
})
```

### RAG Pipeline

```go
// Query documents and generate response
response, _ := client.RAGQuery(ctx, &llamaindex.RAGRequest{
    Query:     "How do I set up the database?",
    TopK:      3,
    SystemPrompt: "Answer based on the provided context. Be concise.",
})

fmt.Println("Answer:", response.Answer)
fmt.Println("Sources:", response.Sources)
```

## Query Transformation Types

| Type | Description |
|------|-------------|
| `hyde` | Hypothetical Document Embeddings |
| `multi_query` | Generate multiple query variations |
| `step_back` | Ask broader questions first |
| `decompose` | Break into sub-questions |

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `LLAMAINDEX_BASE_URL` | LlamaIndex server URL | `http://localhost:8012` |
| `LLAMAINDEX_TIMEOUT` | Request timeout | `120s` |

### Server Setup

```bash
# Start LlamaIndex service
python -m llama_index.serve --port 8012
```

## Testing

```bash
go test -v ./internal/optimization/llamaindex/...
```

## Files

- `client.go` - HTTP client implementation
