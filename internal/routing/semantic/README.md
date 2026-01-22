# Semantic Router Package

This package provides semantic routing for intelligent query routing based on intent classification and embedding similarity.

## Overview

The Semantic Router uses vector embeddings to match user queries to predefined routes, enabling intelligent model selection and request handling based on semantic meaning rather than exact keyword matching.

## Features

- **Embedding-Based Routing**: Match queries by semantic similarity
- **Model Tier Selection**: Route to appropriate model tiers (simple, standard, complex)
- **Caching**: Cache routing decisions and embeddings
- **Multiple Routes**: Support for many route definitions

## Components

### Router (`router.go`)

Main semantic routing implementation:

```go
config := &semantic.RouterConfig{
    Threshold:       0.7,
    DefaultRoute:    "general",
    EnableCache:     true,
    CacheTTL:        time.Hour,
}

router := semantic.NewRouter(encoder, config, logger)
```

### Cache (`cache.go`)

Caching for routing decisions and embeddings:

```go
cache := semantic.NewSemanticCache(&semantic.CacheConfig{
    MaxEntries: 10000,
    TTL:        time.Hour,
})
```

## Data Types

### Route

```go
type Route struct {
    Name        string        // Route identifier
    Description string        // Human-readable description
    Utterances  []string      // Example phrases for matching
    Handler     RouteHandler  // Function to handle matches
    Metadata    map[string]interface{}
    ModelTier   ModelTier     // simple, standard, complex
    Embedding   []float32     // Computed route embedding
    Score       float64       // Match score
}
```

### ModelTier

```go
const (
    ModelTierSimple   ModelTier = "simple"   // Fast, cheap models
    ModelTierStandard ModelTier = "standard" // Balanced models
    ModelTierComplex  ModelTier = "complex"  // Powerful models
)
```

### RouteResult

```go
type RouteResult struct {
    Content  string        // Handler result
    Model    string        // Model used
    Metadata map[string]interface{}
    CacheKey string        // Cache key used
    Latency  time.Duration // Processing time
}
```

## Usage

### Basic Routing

```go
import "dev.helix.agent/internal/routing/semantic"

// Create router with encoder
router := semantic.NewRouter(embeddingEncoder, nil, logger)

// Define routes
router.AddRoute(&semantic.Route{
    Name:        "greeting",
    Description: "Handle greetings",
    Utterances:  []string{"hello", "hi", "hey there", "good morning"},
    ModelTier:   semantic.ModelTierSimple,
    Handler: func(ctx context.Context, query string) (*semantic.RouteResult, error) {
        return &semantic.RouteResult{Content: "Hello! How can I help?"}, nil
    },
})

router.AddRoute(&semantic.Route{
    Name:        "code_review",
    Description: "Code review and analysis",
    Utterances:  []string{"review this code", "check my function", "analyze this snippet"},
    ModelTier:   semantic.ModelTierComplex,
    Handler:     codeReviewHandler,
})

// Route a query
result, route, err := router.Route(ctx, "can you look at my code?")
// route.Name: "code_review" (semantic match)
```

### With Threshold

```go
config := &semantic.RouterConfig{
    Threshold: 0.8, // Require 80% similarity
}

router := semantic.NewRouter(encoder, config, logger)
result, route, err := router.Route(ctx, query)

if route == nil {
    // No route matched above threshold
    // Fall back to default handling
}
```

### Model Tier Routing

```go
// Route determines model tier automatically
result, route, _ := router.Route(ctx, "write me a poem")
switch route.ModelTier {
case semantic.ModelTierSimple:
    // Use Claude Haiku or GPT-3.5
case semantic.ModelTierStandard:
    // Use Claude Sonnet or GPT-4
case semantic.ModelTierComplex:
    // Use Claude Opus or GPT-4 Turbo
}
```

### With Caching

```go
config := &semantic.RouterConfig{
    EnableCache: true,
    CacheTTL:    30 * time.Minute,
}

router := semantic.NewRouter(encoder, config, logger)

// First call computes embedding and routes
result1, _, _ := router.Route(ctx, "hello world")

// Subsequent similar queries use cached routing
result2, _, _ := router.Route(ctx, "hello there") // Cache hit
```

## Encoder Interface

```go
type Encoder interface {
    // Encode generates embeddings for texts
    Encode(ctx context.Context, texts []string) ([][]float32, error)
}
```

Implement with your embedding provider (OpenAI, Sentence Transformers, etc.).

## Testing

```bash
go test -v ./internal/routing/semantic/...
```

## Files

- `router.go` - Main semantic router implementation
- `cache.go` - Routing cache implementation
