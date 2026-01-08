# LlamaIndex + Cognee Integration Guide

LlamaIndex provides advanced document retrieval capabilities that integrate with Cognee's knowledge graph.

## Overview

This integration follows a **Cognee-primary** architecture:

- **Cognee**: Primary source for document indexing, embeddings, and knowledge graph
- **LlamaIndex**: Advanced retrieval features (HyDE, query fusion, reranking)

## Architecture

```
┌────────────────────────────────────────────────┐
│              LlamaIndex Service                 │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐     │
│  │   HyDE   │  │ Reranker │  │  Query   │     │
│  │ Expander │  │          │  │  Fusion  │     │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘     │
│       │             │             │            │
│       └─────────────┼─────────────┘            │
│                     │                          │
│                     ▼                          │
│            ┌────────────────┐                  │
│            │ Cognee Adapter │                  │
│            └───────┬────────┘                  │
└────────────────────┼───────────────────────────┘
                     │
                     ▼
         ┌───────────────────┐
         │      Cognee       │
         │  (Index + Graph)  │
         └───────────────────┘
```

## Docker Setup

```bash
# Start LlamaIndex service
docker-compose --profile optimization up -d llamaindex-server

# Verify
curl http://localhost:8012/health
```

## Configuration

```yaml
optimization:
  llamaindex:
    enabled: true
    endpoint: "http://localhost:8012"
    timeout: "120s"
    use_cognee_index: true  # Query Cognee's index
```

## Basic Usage

### Client Initialization

```go
import "dev.helix.agent/internal/optimization/llamaindex"

config := &llamaindex.ClientConfig{
    BaseURL: "http://localhost:8012",
    Timeout: 120 * time.Second,
}

client := llamaindex.NewClient(config)
```

### Simple Query

```go
response, err := client.Query(ctx, &llamaindex.QueryRequest{
    Query:     "What is machine learning?",
    TopK:      5,
    UseCognee: true,
})

for _, source := range response.Sources {
    fmt.Printf("Score: %.3f\nContent: %s\n\n", source.Score, source.Content)
}
```

## Advanced Retrieval

### HyDE (Hypothetical Document Embeddings)

HyDE generates a hypothetical answer, then searches for similar documents:

```go
response, err := client.QueryWithHyDE(ctx, &llamaindex.QueryRequest{
    Query: "How does photosynthesis work?",
    TopK:  5,
})

// response.HyDEExpansion contains the hypothetical document
fmt.Println("HyDE expansion:", response.HyDEExpansion)
```

### HyDE Expansion Only

Get just the hypothetical document:

```go
expansion, err := client.HyDEExpand(ctx, "Explain quantum entanglement")
fmt.Println("Hypothetical document:", expansion)
```

### Query Decomposition

Break complex queries into simpler sub-queries:

```go
subqueries, err := client.DecomposeQuery(ctx,
    "Compare the economic policies of the US and EU in 2023")

for _, sq := range subqueries {
    fmt.Println("Sub-query:", sq)
}
```

### Query Fusion

Combine results from multiple query variations:

```go
response, err := client.QueryFusion(ctx, &llamaindex.QueryFusionRequest{
    Query:       "machine learning applications",
    NumVariants: 3,
    TopK:        10,
    UseCognee:   true,
})
```

### Reranking

Rerank results for better relevance:

```go
docs := []string{
    "Machine learning is a subset of AI.",
    "Deep learning uses neural networks.",
    "The weather is nice today.",
}

reranked, err := client.Rerank(ctx, &llamaindex.RerankRequest{
    Query:     "What is machine learning?",
    Documents: docs,
    TopK:      2,
})

for _, doc := range reranked {
    fmt.Printf("Score: %.3f - %s\n", doc.Score, doc.Content)
}
```

## Integration with OptimizationService

```go
config := optimization.DefaultConfig()
config.LlamaIndex.Enabled = true
config.LlamaIndex.UseCogneeIndex = true

svc, err := optimization.NewService(config)

// Query documents
response, err := svc.QueryDocuments(ctx, "What is machine learning?", nil)

for _, source := range response.Sources {
    fmt.Println(source.Content)
}
```

### In Request Optimization

The optimization service automatically retrieves context:

```go
optimized, err := svc.OptimizeRequest(ctx, prompt, embedding)

// Retrieved context is added to the prompt
if len(optimized.RetrievedContext) > 0 {
    fmt.Println("Context retrieved from", len(optimized.RetrievedContext), "sources")
    // optimized.OptimizedPrompt contains the augmented prompt
}
```

## Cognee Sync

### How It Works

1. Documents are indexed in Cognee (single source of truth)
2. LlamaIndex queries Cognee's vector store
3. Results are enhanced with LlamaIndex's retrieval features
4. Knowledge graph relationships from Cognee are preserved

### Configuration

```yaml
optimization:
  llamaindex:
    use_cognee_index: true  # Enable Cognee sync
```

### Benefits

- **No Duplicate Indexing**: Documents indexed once in Cognee
- **Graph Relationships**: Access to Cognee's knowledge graph
- **Advanced Retrieval**: LlamaIndex's HyDE, fusion, reranking
- **Unified Management**: Single place for document management

## Query Options

```go
type QueryRequest struct {
    Query       string   // The search query
    TopK        int      // Number of results (default: 5)
    UseCognee   bool     // Use Cognee's index
    UseHyDE     bool     // Enable HyDE expansion
    Rerank      bool     // Apply reranking
    Filters     map[string]interface{} // Metadata filters
    MinScore    float64  // Minimum relevance score
}
```

## Response Structure

```go
type QueryResponse struct {
    Sources       []Source // Retrieved documents
    Query         string   // Original query
    HyDEExpansion string   // HyDE hypothetical document
}

type Source struct {
    Content  string                 // Document content
    Score    float64                // Relevance score
    Metadata map[string]interface{} // Document metadata
    NodeID   string                 // Cognee node ID
}
```

## Best Practices

1. **Enable Cognee Sync**: Set `use_cognee_index: true` to avoid duplicate indexing

2. **Use HyDE for Conceptual Queries**: HyDE works best for questions seeking explanations

3. **Apply Reranking**: Use reranking for large result sets to improve relevance

4. **Adjust TopK**: Start with 5-10 results, increase if needed

5. **Filter by Metadata**: Use filters to narrow down results by source, date, etc.

## Performance Tuning

| Feature | Latency Impact | When to Use |
|---------|---------------|-------------|
| Basic Query | Low | Simple lookups |
| HyDE | Medium | Conceptual queries |
| Reranking | Medium | Large result sets |
| Query Fusion | High | Complex queries |
| Decomposition | High | Multi-part questions |

## Troubleshooting

### No Results

```go
// Check if Cognee index is populated
response, err := client.Query(ctx, &llamaindex.QueryRequest{
    Query:   "test",
    TopK:    1,
    MinScore: 0.0, // Accept any score
})
```

### Low Relevance Scores

```go
// Try HyDE for better matching
response, err := client.QueryWithHyDE(ctx, request)

// Or use query fusion
response, err := client.QueryFusion(ctx, fusionRequest)
```

### Service Unavailable

```go
if !client.IsAvailable(ctx) {
    // Fall back to direct Cognee query
    log.Warn("LlamaIndex unavailable, using Cognee directly")
}
```

## API Endpoints

| Endpoint | Method | Purpose |
|----------|--------|---------|
| /health | GET | Health check |
| /query | POST | Document query |
| /hyde | POST | HyDE expansion |
| /decompose | POST | Query decomposition |
| /rerank | POST | Result reranking |
| /query_fusion | POST | Multi-query fusion |
