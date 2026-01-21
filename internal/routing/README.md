# Package: routing

## Overview

The `routing` package provides intelligent request routing capabilities including semantic routing based on embedding similarity for directing queries to the most appropriate handlers or models.

## Architecture

```
routing/
├── semantic/
│   ├── router.go       # Semantic router implementation
│   └── semantic_test.go # Unit tests (96.2% coverage)
```

## Features

- **Semantic Routing**: Route based on embedding similarity
- **Multi-Route Support**: Define multiple routes with examples
- **Threshold-Based**: Configurable similarity thresholds
- **Fallback Handling**: Default routes for unmatched queries

## Key Types

### SemanticRouter

```go
type SemanticRouter struct {
    routes     []*Route
    embedder   Embedder
    threshold  float64
    defaultRoute string
}
```

### Route

```go
type Route struct {
    Name        string
    Description string
    Examples    []string
    Embeddings  [][]float32  // Pre-computed
    Handler     RouteHandler
}
```

## Usage

### Basic Semantic Routing

```go
import "dev.helix.agent/internal/routing/semantic"

// Create router
router := semantic.NewRouter(embedder, &semantic.Config{
    Threshold:    0.75,
    DefaultRoute: "general",
})

// Add routes
router.AddRoute(&semantic.Route{
    Name:        "code",
    Description: "Programming and code-related queries",
    Examples: []string{
        "Write a Python function",
        "Debug this JavaScript code",
        "Explain this algorithm",
    },
})

router.AddRoute(&semantic.Route{
    Name:        "math",
    Description: "Mathematical problems",
    Examples: []string{
        "Solve this equation",
        "Calculate the derivative",
        "What is the integral of x^2?",
    },
})

// Route a query
result, err := router.Route(ctx, "Help me write a sorting algorithm")
// result.Name == "code"
// result.Score == 0.92
```

### With Custom Handlers

```go
router.AddRoute(&semantic.Route{
    Name: "code",
    Handler: func(ctx context.Context, query string) (*Response, error) {
        // Use code-specialized model
        return codeModel.Complete(ctx, query)
    },
})

// Execute route handler
response, err := router.RouteAndExecute(ctx, query)
```

### Multi-Match Routing

```go
// Get top-k matches
matches, err := router.RouteTopK(ctx, query, 3)
for _, match := range matches {
    fmt.Printf("Route: %s, Score: %.2f\n", match.Name, match.Score)
}
```

## Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| Threshold | float64 | 0.7 | Minimum similarity score |
| DefaultRoute | string | "" | Fallback route name |
| TopK | int | 1 | Number of matches to return |
| UseCache | bool | true | Cache route embeddings |

## Testing

```bash
go test -v ./internal/routing/...
go test -cover ./internal/routing/semantic/...  # 96.2% coverage
```

## Dependencies

### Internal
- `internal/embedding` - Text embeddings

### External
- Standard library only

## See Also

- [Semantic Router Paper](https://arxiv.org/abs/2401.02093)
- [Routing API Reference](../../docs/api/routing.md)
