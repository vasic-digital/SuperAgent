# Embedding Package

The embedding package provides vector embedding services for HelixAgent, enabling semantic search and similarity computation.

## Overview

This package implements:

- **Multiple Embedding Providers**: OpenAI, Ollama, HuggingFace
- **Embedding Cache**: Efficient caching for repeated embeddings
- **Model Registry**: Dynamic model registration and discovery
- **Batch Processing**: Efficient batch embedding operations

## Key Components

### Embedding Providers

#### OpenAI Embeddings

```go
config := &embedding.OpenAIConfig{
    APIKey:     os.Getenv("OPENAI_API_KEY"),
    Model:      "text-embedding-3-small",
    Dimensions: 1536,
}

provider := embedding.NewOpenAIEmbedding(config)

embeddings, err := provider.Embed(ctx, []string{"text to embed"})
```

#### Ollama Embeddings

```go
config := &embedding.OllamaConfig{
    BaseURL: "http://localhost:11434",
    Model:   "nomic-embed-text",
}

provider := embedding.NewOllamaEmbedding(config)
```

#### HuggingFace Embeddings

```go
config := &embedding.HuggingFaceConfig{
    APIKey: os.Getenv("HF_API_KEY"),
    Model:  "sentence-transformers/all-MiniLM-L6-v2",
}

provider := embedding.NewHuggingFaceEmbedding(config)
```

### Embedding Cache

```go
cache := embedding.NewEmbeddingCache(&embedding.CacheConfig{
    MaxSize: 10000,
    TTL:     24 * time.Hour,
})

// Get or compute embedding
emb, err := cache.GetOrCompute(ctx, text, func() ([]float64, error) {
    return provider.Embed(ctx, []string{text})
})
```

### Model Registry

```go
registry := embedding.NewEmbeddingModelRegistry()

// Register model
registry.Register("openai-3-small", &embedding.ModelInfo{
    Provider:   "openai",
    Dimensions: 1536,
    MaxTokens:  8191,
})

// Get model info
info, ok := registry.Get("openai-3-small")
```

## Supported Models

| Provider | Model | Dimensions |
|----------|-------|------------|
| OpenAI | text-embedding-3-small | 1536 |
| OpenAI | text-embedding-3-large | 3072 |
| OpenAI | text-embedding-ada-002 | 1536 |
| Ollama | nomic-embed-text | 768 |
| Ollama | mxbai-embed-large | 1024 |
| HuggingFace | all-MiniLM-L6-v2 | 384 |

## Usage Examples

### Single Embedding

```go
embedding, err := provider.Embed(ctx, []string{"Hello world"})
if err != nil {
    return err
}

vector := embedding[0] // []float64
```

### Batch Embedding

```go
texts := []string{"text 1", "text 2", "text 3"}
embeddings, err := provider.Embed(ctx, texts)

for i, emb := range embeddings {
    fmt.Printf("Text %d: %d dimensions\n", i, len(emb))
}
```

### Similarity Computation

```go
similarity := embedding.CosineSimilarity(vec1, vec2)
fmt.Printf("Similarity: %.4f\n", similarity)

// Find most similar
matches := embedding.FindMostSimilar(query, candidates, 10)
```

## Configuration

```go
type EmbeddingConfig struct {
    Provider    string        // "openai", "ollama", "huggingface"
    Model       string        // Model identifier
    Dimensions  int           // Output dimensions
    MaxBatch    int           // Max texts per batch
    Timeout     time.Duration // Request timeout
    RetryCount  int           // Retry on failure
    EnableCache bool          // Enable embedding cache
}
```

## Error Handling

```go
emb, err := provider.Embed(ctx, texts)
if err != nil {
    switch e := err.(type) {
    case *embedding.RateLimitError:
        time.Sleep(e.RetryAfter)
        // Retry
    case *embedding.TokenLimitError:
        // Split text and retry
    default:
        return err
    }
}
```

## Testing

```bash
# Run all embedding tests
go test -v ./internal/embedding/...

# Test with mock server
go test -v -run TestMock ./internal/embedding/

# Benchmark embedding performance
go test -bench=. ./internal/embedding/
```

## Performance Tips

1. **Use Batch API**: Embed multiple texts in single request
2. **Enable Caching**: Cache frequently used embeddings
3. **Choose Model Size**: Balance quality vs. speed
4. **Async Processing**: Use goroutines for large batches

## See Also

- `internal/vectordb/` - Vector database integration
- `internal/rag/` - Retrieval Augmented Generation
- `internal/debate/` - Lesson Bank semantic search
