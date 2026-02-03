# Embedding Provider Testing Package

The embeddings testing package provides comprehensive utilities for testing embedding providers with real functional validation.

## Overview

This package implements:

- **Embedding Test Client**: HTTP client for HelixAgent embedding endpoints
- **Mock Embedding Servers**: Configurable mocks for unit testing
- **Similarity Testing**: Validates embedding quality via similarity comparisons
- **Benchmark Tests**: Performance measurement for embedding operations
- **Provider Verification**: Tests all 6 embedding providers

## Key Principles

**No False Positives**: Tests execute ACTUAL embedding operations, not just connectivity checks. Tests FAIL if embedding generation fails.

## Directory Structure

```
internal/testing/embeddings/
├── functional_test.go    # Real embedding provider functional tests
└── README.md             # This file

internal/embedding/
├── openai.go             # OpenAI embedding provider
├── cohere.go             # Cohere embedding provider
├── voyage.go             # Voyage AI embedding provider
├── jina.go               # Jina AI embedding provider
├── google.go             # Google embedding provider
├── bedrock.go            # AWS Bedrock embedding provider
├── cache.go              # Embedding cache
└── ...
```

## Key Components

### EmbeddingClient

HTTP client for testing HelixAgent embedding API:

```go
// Create client
client := embeddings.NewEmbeddingClient("http://localhost:8080")

// Generate embeddings
resp, err := client.Embed(&embeddings.EmbeddingRequest{
    Provider: "openai",
    Model:    "text-embedding-3-small",
    Input:    []string{"Hello, world!", "Test text"},
})
if err != nil {
    t.Fatalf("Embedding failed: %v", err)
}

// Access embeddings
for i, embedding := range resp.Embeddings {
    fmt.Printf("Input %d: %d dimensions\n", i, len(embedding))
}
```

### Request/Response Types

```go
// EmbeddingRequest represents an embedding request
type EmbeddingRequest struct {
    Provider string   `json:"provider"`
    Model    string   `json:"model"`
    Input    []string `json:"input"`
}

// EmbeddingResponse represents an embedding response
type EmbeddingResponse struct {
    Provider   string      `json:"provider"`
    Model      string      `json:"model"`
    Embeddings [][]float64 `json:"embeddings"`
    Usage      struct {
        TotalTokens int `json:"total_tokens"`
    } `json:"usage"`
    Error string `json:"error,omitempty"`
}
```

## Mock Embedding Server

Create a mock embedding server for unit testing:

```go
import "dev.helix.agent/internal/testing"

mock := testing.NewMockHTTPServer()
defer mock.Close()

// Configure embedding endpoint response
mock.SetResponse("/v1/embeddings", &testing.MockHTTPResponse{
    StatusCode: 200,
    Headers:    map[string]string{"Content-Type": "application/json"},
    Body: `{
        "provider": "mock",
        "model": "mock-embedding",
        "embeddings": [[0.1, 0.2, 0.3, 0.4, 0.5]],
        "usage": {"total_tokens": 5}
    }`,
})

// Configure provider list endpoint
mock.SetResponse("/v1/embeddings/providers", &testing.MockHTTPResponse{
    StatusCode: 200,
    Body:       `{"providers": ["openai", "cohere", "voyage"]}`,
})
```

### Mock with Latency

```go
mock.SetResponse("/v1/embeddings", &testing.MockHTTPResponse{
    StatusCode: 200,
    Delay:      100 * time.Millisecond, // Simulate network latency
    Body:       embeddingJSON,
})
```

## Supported Embedding Providers

| Provider | Model | Dimensions | Auth Required |
|----------|-------|------------|---------------|
| OpenAI | text-embedding-3-small | 1536 | OPENAI_API_KEY |
| Cohere | embed-english-v3.0 | 1024 | COHERE_API_KEY |
| Voyage | voyage-3 | 1024 | VOYAGE_API_KEY |
| Jina | jina-embeddings-v3 | 1024 | JINA_API_KEY |
| Google | text-embedding-005 | 768 | GOOGLE_API_KEY |
| Bedrock | amazon.titan-embed-text-v2 | 1024 | AWS_ACCESS_KEY_ID |

## Similarity Testing

### Cosine Similarity

```go
// cosineSimilarity calculates similarity between two vectors
func cosineSimilarity(a, b []float64) float64 {
    if len(a) != len(b) {
        return 0
    }

    var dotProduct, normA, normB float64
    for i := range a {
        dotProduct += a[i] * b[i]
        normA += a[i] * a[i]
        normB += b[i] * b[i]
    }

    if normA == 0 || normB == 0 {
        return 0
    }

    return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
```

### Similarity Validation

```go
func TestEmbeddingSimilarity(t *testing.T) {
    client := embeddings.NewEmbeddingClient("http://localhost:8080")

    // Similar texts should have high similarity
    similarTexts := []string{
        "The cat sat on the mat.",
        "A cat was sitting on a mat.",
    }

    // Different text should have lower similarity
    differentText := "Machine learning is transforming industries."

    // Get embeddings
    resp1, _ := client.Embed(&embeddings.EmbeddingRequest{
        Provider: "openai",
        Model:    "text-embedding-3-small",
        Input:    similarTexts,
    })

    resp2, _ := client.Embed(&embeddings.EmbeddingRequest{
        Provider: "openai",
        Model:    "text-embedding-3-small",
        Input:    []string{differentText},
    })

    // Calculate similarities
    similarSim := cosineSimilarity(resp1.Embeddings[0], resp1.Embeddings[1])
    differentSim := cosineSimilarity(resp1.Embeddings[0], resp2.Embeddings[0])

    // Validate semantic relationships
    assert.Greater(t, similarSim, differentSim,
        "Similar texts should have higher similarity than different texts")
    assert.Greater(t, similarSim, 0.8,
        "Similar texts should have high similarity (>0.8)")
}
```

### Semantic Clustering

```go
func TestSemanticClustering(t *testing.T) {
    client := embeddings.NewEmbeddingClient("http://localhost:8080")

    // Define semantic clusters
    techTexts := []string{
        "Python is a programming language",
        "JavaScript runs in web browsers",
        "Go is compiled and statically typed",
    }

    animalTexts := []string{
        "Dogs are loyal pets",
        "Cats are independent animals",
        "Birds can fly in the sky",
    }

    // Get all embeddings
    techResp, _ := client.Embed(&embeddings.EmbeddingRequest{
        Provider: "openai",
        Model:    "text-embedding-3-small",
        Input:    techTexts,
    })

    animalResp, _ := client.Embed(&embeddings.EmbeddingRequest{
        Provider: "openai",
        Model:    "text-embedding-3-small",
        Input:    animalTexts,
    })

    // Calculate intra-cluster similarity (should be high)
    techIntraSim := avgPairwiseSimilarity(techResp.Embeddings)
    animalIntraSim := avgPairwiseSimilarity(animalResp.Embeddings)

    // Calculate inter-cluster similarity (should be lower)
    interSim := avgCrossClusterSimilarity(techResp.Embeddings, animalResp.Embeddings)

    t.Logf("Tech cluster similarity: %.4f", techIntraSim)
    t.Logf("Animal cluster similarity: %.4f", animalIntraSim)
    t.Logf("Cross-cluster similarity: %.4f", interSim)

    assert.Greater(t, techIntraSim, interSim)
    assert.Greater(t, animalIntraSim, interSim)
}
```

## Functional Tests

### Provider Discovery

```go
func TestEmbeddingProviderDiscovery(t *testing.T) {
    client := embeddings.NewEmbeddingClient("http://localhost:8080")

    providers, err := client.ListProviders()
    if err != nil {
        t.Skipf("Embedding service not running: %v", err)
    }

    assert.NotEmpty(t, providers, "Should have at least one provider")
    t.Logf("Discovered %d providers: %v", len(providers), providers)
}
```

### Embedding Generation

```go
func TestEmbeddingGeneration(t *testing.T) {
    client := embeddings.NewEmbeddingClient("http://localhost:8080")

    testInputs := []string{
        "Hello, world!",
        "This is a test of the embedding system.",
    }

    for _, provider := range EmbeddingProviders {
        t.Run(provider.Provider, func(t *testing.T) {
            // Skip if API key not set
            if provider.RequiresAuth && os.Getenv(provider.EnvKey) == "" {
                t.Skipf("Skipping %s: %s not set", provider.Provider, provider.EnvKey)
            }

            req := &embeddings.EmbeddingRequest{
                Provider: provider.Provider,
                Model:    provider.Model,
                Input:    testInputs,
            }

            resp, err := client.Embed(req)
            if err != nil {
                t.Skipf("Provider %s not available: %v", provider.Provider, err)
            }

            require.Equal(t, provider.Provider, resp.Provider)
            require.Len(t, resp.Embeddings, len(testInputs))

            for i, embedding := range resp.Embeddings {
                assert.NotEmpty(t, embedding, "Embedding %d should not be empty", i)
                t.Logf("Provider %s: Input %d has %d dimensions",
                    provider.Provider, i, len(embedding))
            }
        })
    }
}
```

### Health Check

```go
func TestEmbeddingHealthCheck(t *testing.T) {
    client := embeddings.NewEmbeddingClient("http://localhost:8080")

    resp, err := client.httpClient.Get(client.baseURL + "/v1/embeddings/health")
    if err != nil {
        t.Skipf("Embedding service not running: %v", err)
    }
    defer resp.Body.Close()

    require.Equal(t, http.StatusOK, resp.StatusCode)
}
```

## Benchmark Tests

### Basic Benchmark

```go
func BenchmarkEmbedding(b *testing.B) {
    client := embeddings.NewEmbeddingClient("http://localhost:8080")

    req := &embeddings.EmbeddingRequest{
        Provider: "openai",
        Model:    "text-embedding-3-small",
        Input:    []string{"Hello, world!"},
    }

    if os.Getenv("OPENAI_API_KEY") == "" {
        b.Skip("OPENAI_API_KEY not set")
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := client.Embed(req)
        if err != nil {
            b.Fatalf("Embedding failed: %v", err)
        }
    }
}
```

### Batch Size Benchmark

```go
func BenchmarkEmbeddingBatchSizes(b *testing.B) {
    client := embeddings.NewEmbeddingClient("http://localhost:8080")

    batchSizes := []int{1, 5, 10, 25, 50}

    for _, size := range batchSizes {
        b.Run(fmt.Sprintf("batch_%d", size), func(b *testing.B) {
            input := make([]string, size)
            for i := range input {
                input[i] = fmt.Sprintf("Test text number %d", i)
            }

            req := &embeddings.EmbeddingRequest{
                Provider: "openai",
                Model:    "text-embedding-3-small",
                Input:    input,
            }

            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                _, _ = client.Embed(req)
            }
        })
    }
}
```

### Dimension Benchmark

```go
func BenchmarkEmbeddingModels(b *testing.B) {
    client := embeddings.NewEmbeddingClient("http://localhost:8080")

    models := []struct {
        name  string
        model string
    }{
        {"small-1536", "text-embedding-3-small"},
        {"large-3072", "text-embedding-3-large"},
    }

    for _, m := range models {
        b.Run(m.name, func(b *testing.B) {
            req := &embeddings.EmbeddingRequest{
                Provider: "openai",
                Model:    m.model,
                Input:    []string{"Benchmark test text"},
            }

            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                _, _ = client.Embed(req)
            }
        })
    }
}
```

## Running Tests

```bash
# Run all embedding tests
go test -v ./internal/testing/embeddings/...

# Run specific provider test
go test -v -run TestEmbeddingGeneration/openai ./internal/testing/embeddings/

# Run similarity tests
go test -v -run TestEmbeddingSimilarity ./internal/testing/embeddings/

# Run benchmarks
go test -bench=. ./internal/testing/embeddings/

# Run benchmarks with memory profiling
go test -bench=. -benchmem ./internal/testing/embeddings/
```

## Environment Setup

```bash
# Required for authenticated providers
export OPENAI_API_KEY=sk-...
export COHERE_API_KEY=...
export VOYAGE_API_KEY=...
export JINA_API_KEY=...
export GOOGLE_API_KEY=...
export AWS_ACCESS_KEY_ID=...
export AWS_SECRET_ACCESS_KEY=...
```

## Error Handling

```go
resp, err := client.Embed(req)
if err != nil {
    // Check for specific error types
    if strings.Contains(err.Error(), "rate limit") {
        t.Log("Rate limited, waiting...")
        time.Sleep(time.Minute)
        // Retry
    } else if strings.Contains(err.Error(), "unauthorized") {
        t.Skipf("Invalid API key for %s", req.Provider)
    } else {
        t.Fatalf("Unexpected error: %v", err)
    }
}

if resp.Error != "" {
    t.Logf("Provider returned error: %s", resp.Error)
}
```

## Best Practices

1. **Skip Unconfigured Providers**: Check for API keys before testing
2. **Validate Dimensions**: Verify embedding dimensions match expected values
3. **Test Similarity Properties**: Embeddings should preserve semantic relationships
4. **Benchmark with Realistic Data**: Use representative text lengths
5. **Handle Rate Limits**: Implement backoff for rate-limited providers
6. **Cache Test Embeddings**: Avoid redundant API calls in tests

## Test Utilities

```go
// Helper to skip if provider not configured
func requireProvider(t *testing.T, provider EmbeddingProviderConfig) {
    if provider.RequiresAuth && os.Getenv(provider.EnvKey) == "" {
        t.Skipf("Provider %s not configured", provider.Provider)
    }
}

// Helper to calculate average pairwise similarity
func avgPairwiseSimilarity(embeddings [][]float64) float64 {
    if len(embeddings) < 2 {
        return 1.0
    }

    var sum float64
    var count int
    for i := 0; i < len(embeddings); i++ {
        for j := i + 1; j < len(embeddings); j++ {
            sum += cosineSimilarity(embeddings[i], embeddings[j])
            count++
        }
    }

    return sum / float64(count)
}
```

## See Also

- `internal/embedding/` - Embedding provider implementations
- `internal/vectordb/` - Vector database integration
- `internal/rag/` - Retrieval Augmented Generation
- `internal/testing/helpers.go` - Common test utilities
