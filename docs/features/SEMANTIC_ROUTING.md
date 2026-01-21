# Semantic Routing System

## Overview

The Semantic Routing System in HelixAgent provides intelligent query routing based on embedding similarity. It analyzes incoming queries and routes them to the most appropriate handler, LLM provider, or pipeline based on semantic understanding rather than keyword matching.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Semantic Router                          │
│  ┌─────────────────┐  ┌─────────────────────────────────┐ │
│  │ Query Embedder  │  │ Route Registry                  │ │
│  │  (vectorize)    │  │  ├─ Route definitions           │ │
│  └────────┬────────┘  │  ├─ Reference embeddings        │ │
│           │           │  └─ Similarity thresholds       │ │
│           ▼           └─────────────┬───────────────────┘ │
│  ┌─────────────────┐               │                      │
│  │ Similarity      │◄──────────────┘                      │
│  │ Calculator      │                                      │
│  └────────┬────────┘                                      │
│           │                                                │
│           ▼                                                │
│  ┌─────────────────┐                                      │
│  │ Route Selector  │───▶ Best matching route              │
│  └─────────────────┘                                      │
└─────────────────────────────────────────────────────────────┘
```

## Components

### 1. Router (`internal/routing/semantic/router.go`)

The main router component that manages routes and performs matching.

```go
import "dev.helix.agent/internal/routing/semantic"

// Create router with embedding provider
router := semantic.NewRouter(&semantic.Config{
    EmbeddingProvider:   embeddingProvider,
    SimilarityThreshold: 0.7,
    TopK:                3,
})

// Add routes
router.AddRoute(&semantic.Route{
    Name:        "code-review",
    Description: "Code review and analysis requests",
    Utterances: []string{
        "review this code",
        "check for bugs",
        "analyze the code quality",
        "find issues in my code",
    },
    Handler: codeReviewHandler,
})

// Route a query
result, err := router.Route(ctx, "please review my function for bugs")
// result.Route.Name == "code-review"
// result.Confidence == 0.92
```

### 2. Route Definition

Routes are defined with sample utterances that represent the intent.

```go
type Route struct {
    Name        string                 // Unique route identifier
    Description string                 // Human-readable description
    Utterances  []string               // Sample utterances for this route
    Handler     RouteHandler           // Handler function
    Metadata    map[string]interface{} // Custom metadata
    Priority    int                    // Priority for tie-breaking
}

type RouteHandler func(ctx context.Context, query string, meta RouteMetadata) (*RouteResult, error)
```

### 3. Similarity Calculation

The router uses cosine similarity between query embeddings and route embeddings.

```go
// Cosine similarity calculation
func cosineSimilarity(a, b []float64) float64 {
    var dot, normA, normB float64
    for i := range a {
        dot += a[i] * b[i]
        normA += a[i] * a[i]
        normB += b[i] * b[i]
    }
    return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
```

## Route Types

### Intent-Based Routes

Route based on user intent detection:

```go
routes := []*semantic.Route{
    {
        Name: "generate-code",
        Utterances: []string{
            "write code for",
            "generate a function",
            "create a class",
            "implement this feature",
        },
        Handler: codeGenerationHandler,
    },
    {
        Name: "explain-code",
        Utterances: []string{
            "explain this code",
            "what does this do",
            "help me understand",
            "walk me through",
        },
        Handler: codeExplanationHandler,
    },
    {
        Name: "debug-code",
        Utterances: []string{
            "fix this bug",
            "why is this failing",
            "debug this error",
            "find the issue",
        },
        Handler: debugHandler,
    },
}
```

### Provider-Based Routes

Route to specific LLM providers based on task type:

```go
routes := []*semantic.Route{
    {
        Name: "complex-reasoning",
        Utterances: []string{
            "solve this complex problem",
            "analyze this deeply",
            "think through this carefully",
        },
        Metadata: map[string]interface{}{
            "preferred_provider": "claude",
            "model": "claude-opus-4-5-20251101",
        },
    },
    {
        Name: "fast-completion",
        Utterances: []string{
            "quick question",
            "simple task",
            "just complete this",
        },
        Metadata: map[string]interface{}{
            "preferred_provider": "deepseek",
            "model": "deepseek-chat",
        },
    },
}
```

### Domain-Specific Routes

Route to specialized handlers based on domain:

```go
routes := []*semantic.Route{
    {
        Name: "database-queries",
        Utterances: []string{
            "SQL query",
            "database schema",
            "optimize this query",
            "PostgreSQL help",
        },
        Handler: databaseHandler,
    },
    {
        Name: "kubernetes",
        Utterances: []string{
            "deploy to k8s",
            "kubernetes config",
            "helm chart",
            "pod configuration",
        },
        Handler: kubernetesHandler,
    },
}
```

## Configuration

```go
type Config struct {
    // Embedding provider for vectorization
    EmbeddingProvider EmbeddingProvider

    // Minimum similarity score to consider a match
    SimilarityThreshold float64 // Default: 0.7

    // Number of top matches to consider
    TopK int // Default: 3

    // Whether to use weighted averaging for multiple utterances
    UseWeightedAverage bool // Default: true

    // Cache embeddings for performance
    CacheEmbeddings bool // Default: true

    // Fallback route when no match found
    FallbackRoute *Route
}
```

## Usage Examples

### Basic Routing

```go
// Initialize router
router := semantic.NewRouter(config)

// Add routes
router.AddRoute(codeReviewRoute)
router.AddRoute(debugRoute)
router.AddRoute(generateRoute)

// Route query
result, err := router.Route(ctx, "can you check my code for potential issues?")
if err != nil {
    return err
}

fmt.Printf("Matched route: %s (confidence: %.2f)\n",
    result.Route.Name, result.Confidence)

// Execute handler
response, err := result.Route.Handler(ctx, query, result.Metadata)
```

### With Fallback

```go
router := semantic.NewRouter(&semantic.Config{
    SimilarityThreshold: 0.8,
    FallbackRoute: &semantic.Route{
        Name:    "general",
        Handler: generalHandler,
    },
})

// If no route matches above 0.8, use fallback
result, _ := router.Route(ctx, "random question")
// result.Route.Name might be "general" if no good match
```

### Multi-Route Matching

```go
// Get top K matches
results, err := router.RouteMultiple(ctx, query, 3)
for _, result := range results {
    fmt.Printf("%s: %.2f\n", result.Route.Name, result.Confidence)
}
```

## Integration with HelixAgent

### Handler Integration

```go
// In internal/handlers/router_integration.go
func (h *Handler) routeRequest(ctx context.Context, req *Request) (*Response, error) {
    // Route the query
    routeResult, err := h.semanticRouter.Route(ctx, req.Query)
    if err != nil {
        return nil, fmt.Errorf("routing failed: %w", err)
    }

    // Execute matched handler
    return routeResult.Route.Handler(ctx, req.Query, routeResult.Metadata)
}
```

### Provider Selection

```go
// Route to best provider based on query type
func (s *Service) selectProvider(ctx context.Context, query string) (string, error) {
    result, err := s.router.Route(ctx, query)
    if err != nil {
        return s.defaultProvider, nil
    }

    if provider, ok := result.Route.Metadata["preferred_provider"].(string); ok {
        return provider, nil
    }

    return s.defaultProvider, nil
}
```

## Testing

```bash
# Run semantic routing tests
go test -v ./internal/routing/semantic/...

# Run with coverage
go test -cover ./internal/routing/semantic/...

# Run benchmarks
go test -bench=. ./internal/routing/semantic/...
```

### Test Examples

```go
func TestRouter_RouteMatching(t *testing.T) {
    router := semantic.NewRouter(testConfig)

    router.AddRoute(&semantic.Route{
        Name: "greeting",
        Utterances: []string{"hello", "hi there", "good morning"},
    })

    result, err := router.Route(ctx, "hey, good morning!")
    require.NoError(t, err)
    assert.Equal(t, "greeting", result.Route.Name)
    assert.Greater(t, result.Confidence, 0.7)
}
```

## Performance Considerations

1. **Embedding Caching**: Route embeddings are computed once and cached
2. **Batch Processing**: Multiple queries can be routed in batch for efficiency
3. **Similarity Indexing**: For large route sets, consider using approximate nearest neighbor (ANN) search

## Key Files

| File | Description |
|------|-------------|
| `internal/routing/semantic/router.go` | Main router implementation |
| `internal/routing/semantic/route.go` | Route definition and types |
| `internal/routing/semantic/similarity.go` | Similarity calculation |
| `internal/routing/semantic/config.go` | Configuration options |
| `internal/routing/semantic/semantic_test.go` | Comprehensive tests |

## See Also

- [Intent Classification](./INTENT_CLASSIFICATION.md)
- [Provider Registry](../api/providers.md)
- [Handler Architecture](../architecture/handlers.md)
