# Embeddings Package

The embeddings package provides embedding model management and integration for HelixAgent's semantic search, RAG, and similarity-based features.

## Overview

This package contains the embedding model registry and interfaces for working with multiple embedding providers. It supports various embedding models from OpenAI, Ollama, sentence-transformers, and local models.

## Key Components

### EmbeddingModel Interface

```go
type EmbeddingModel interface {
    // Encode generates embeddings for multiple texts.
    Encode(ctx context.Context, texts []string) ([][]float32, error)

    // EncodeSingle generates an embedding for a single text.
    EncodeSingle(ctx context.Context, text string) ([]float32, error)

    // Name returns the model name.
    Name() string

    // Dimensions returns the embedding dimensions.
    Dimensions() int

    // MaxTokens returns the maximum tokens supported.
    MaxTokens() int

    // Provider returns the provider name.
    Provider() string

    // Health checks if the model is healthy.
    Health(ctx context.Context) error

    // Close releases any resources.
    Close() error
}
```

### EmbeddingModelConfig

```go
type EmbeddingModelConfig struct {
    Name         string        `json:"name"`
    Provider     string        `json:"provider"` // "openai", "ollama", "sentence-transformers", "local"
    ModelID      string        `json:"model_id"`
    Dimensions   int           `json:"dimensions"`
    MaxTokens    int           `json:"max_tokens"`
    BatchSize    int           `json:"batch_size"`
    Timeout      time.Duration `json:"timeout"`
    CacheEnabled bool          `json:"cache_enabled"`
    CacheTTL     time.Duration `json:"cache_ttl"`
    BaseURL      string        `json:"base_url,omitempty"`
    APIKey       string        `json:"api_key,omitempty"`
}
```

## Supported Providers

| Provider | Models | Dimensions |
|----------|--------|------------|
| **OpenAI** | text-embedding-3-small, text-embedding-3-large, ada-002 | 512-3072 |
| **Cohere** | embed-english-v3.0, embed-multilingual-v3.0 | 384-4096 |
| **Voyage** | voyage-3, voyage-code-3, voyage-finance-2 | 512-1536 |
| **Jina** | jina-embeddings-v3, jina-clip-v1 | 128-1024 |
| **Google** | text-embedding-005, textembedding-gecko | 768 |
| **AWS Bedrock** | amazon.titan-embed-text-v1/v2 | 1024-1536 |

## Features

- **Multi-provider Support**: Work with various embedding providers
- **Caching**: Optional embedding caching with configurable TTL
- **Batch Processing**: Efficient batch embedding for large document sets
- **Health Checks**: Monitor embedding model availability
- **Similarity Functions**: Cosine, dot product, and Euclidean distance

## Subpackages

### internal/embeddings/models

Contains the embedding model registry and provider-specific implementations.

## Usage

```go
import "dev.helix.agent/internal/embeddings/models"

// Create a registry
registry := models.NewRegistry()

// Register an embedding model
config := &models.EmbeddingModelConfig{
    Name:       "openai-small",
    Provider:   "openai",
    ModelID:    "text-embedding-3-small",
    Dimensions: 1536,
}
model, err := registry.CreateModel(config)

// Generate embeddings
embeddings, err := model.Encode(ctx, []string{"Hello world", "How are you?"})
```

## Testing

```bash
go test -v ./internal/embeddings/...
```

## Related Packages

- `internal/embedding` - Embedding provider clients
- `internal/vectordb` - Vector database storage
- `internal/rag` - Retrieval-augmented generation
