# Embeddings Protocol - Complete Specification

**Protocol:** Embeddings API  
**Version:** 1.0  
**Status:** Stable  
**HelixAgent Implementation:** [internal/embeddings/](../../../internal/embeddings/)  
**Analysis Date:** 2026-04-03  

---

## Executive Summary

The Embeddings Protocol standardizes text-to-vector conversion for semantic search, similarity comparison, and clustering. HelixAgent supports multiple embedding providers and vector databases with a unified API.

**Key Capabilities:**
- Multi-provider embedding generation
- Batch processing
- Vector similarity search
- Caching and optimization
- Multiple vector database backends

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                  EMBEDDINGS ARCHITECTURE                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   Input Text → Embedding Provider → Vector (1536+ dimensions)   │
│                                                                  │
│   ┌──────────────────────────────────────────────────────────┐  │
│   │                   EMBEDDING PROVIDERS                     │  │
│   │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐ │  │
│   │  │ OpenAI   │  │ Google   │  │ Cohere   │  │ Local    │ │  │
│   │  │ text-emb │  │ Gecko    │  │ Embed    │  │ Models   │ │  │
│   │  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘ │  │
│   │       └─────────────┴─────────────┴─────────────┘       │  │
│   │                        │                                 │  │
│   │                   Normalized Vectors                      │  │
│   └─────────────────────────┬────────────────────────────────┘  │
│                             │                                    │
│   ┌─────────────────────────┴────────────────────────────────┐  │
│   │                   VECTOR DATABASES                        │  │
│   │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐ │  │
│   │  │ChromaDB  │  │ Qdrant   │  │Pinecone  │  │Weaviate  │ │  │
│   │  │          │  │          │  │          │  │          │ │  │
│   │  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘ │  │
│   │       └─────────────┴─────────────┴─────────────┘       │  │
│   │                      Similarity Search                   │  │
│   └──────────────────────────────────────────────────────────┘  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Protocol Basics

### Embedding Request Format

```json
{
  "model": "text-embedding-3-large",
  "input": [
    "The quick brown fox",
    "jumps over the lazy dog"
  ],
  "encoding_format": "float",
  "dimensions": 3072
}
```

### Embedding Response Format

```json
{
  "object": "list",
  "data": [
    {
      "object": "embedding",
      "embedding": [0.0023, -0.0091, 0.0156, ...],  // 3072 floats
      "index": 0
    },
    {
      "object": "embedding",
      "embedding": [0.0012, -0.0078, 0.0123, ...],
      "index": 1
    }
  ],
  "model": "text-embedding-3-large",
  "usage": {
    "prompt_tokens": 12,
    "total_tokens": 12
  }
}
```

### Vector Properties

| Property | Description | Typical Values |
|----------|-------------|----------------|
| **Dimensions** | Vector size | 384, 768, 1536, 3072 |
| **Normalization** | L2 norm | Usually 1.0 |
| **Data Type** | Storage format | float32, float64, int8 |
| **Similarity Metric** | Distance measure | cosine, euclidean, dot |

---

## Supported Models

### OpenAI Embeddings

| Model | Dimensions | Max Tokens | Cost ($/1M) | Best For |
|-------|-----------|------------|-------------|----------|
| text-embedding-3-large | 3072 | 8191 | $0.13 | Best quality |
| text-embedding-3-small | 1536 | 8191 | $0.02 | Balanced |
| text-embedding-ada-002 | 1536 | 8191 | $0.10 | Legacy |

### Google Embeddings

| Model | Dimensions | Max Tokens | Best For |
|-------|-----------|------------|----------|
| text-embedding-004 | 768 | 2048 | Google Cloud |
| embedding-001 | 768 | 1024 | General |
| gecko-multilingual | 768 | 1024 | Multilingual |

### Cohere Embeddings

| Model | Dimensions | Max Tokens | Best For |
|-------|-----------|------------|----------|
| embed-english-v3 | 1024 | 512 | English |
| embed-multilingual-v3 | 1024 | 512 | Multilingual |

### Local Models

| Model | Dimensions | Framework | Best For |
|-------|-----------|-----------|----------|
| all-MiniLM-L6-v2 | 384 | SentenceTransformers | Speed |
| all-mpnet-base-v2 | 768 | SentenceTransformers | Quality |
| bge-large-en | 1024 | FlagEmbedding | RAG |

---

## Similarity Search

### Cosine Similarity

```python
def cosine_similarity(a: list[float], b: list[float]) -> float:
    """Calculate cosine similarity between two vectors"""
    dot_product = sum(x * y for x, y in zip(a, b))
    norm_a = sum(x * x for x in a) ** 0.5
    norm_b = sum(x * x for x in b) ** 0.5
    return dot_product / (norm_a * norm_b)

# Interpretation:
# 1.0 = identical
# 0.0 = orthogonal
# -1.0 = opposite
```

### Similarity Thresholds

| Threshold | Interpretation | Use Case |
|-----------|---------------|----------|
| 0.95+ | Nearly identical | Duplicate detection |
| 0.85-0.95 | Very similar | Semantic search |
| 0.70-0.85 | Related | Recommendation |
| 0.50-0.70 | Somewhat related | Broad matching |
| < 0.50 | Not related | Filter out |

---

## HelixAgent Embeddings Implementation

### Architecture

**Source:** [`internal/embeddings/`](../../../internal/embeddings/)

```
internal/embeddings/
├── embeddings.go           # Core embedding interface
├── providers/
│   ├── openai.go          # OpenAI provider
│   ├── google.go          # Google provider
│   ├── cohere.go          # Cohere provider
│   └── local.go           # Local models
├── cache/
│   ├── redis.go           # Redis cache
│   └── memory.go          # In-memory cache
├── batch.go               # Batch processing
└── normalize.go           # Vector normalization
```

### Core Interface

**Source:** [`internal/embeddings/embeddings.go`](../../../internal/embeddings/embeddings.go)

```go
package embeddings

// EmbeddingProvider interface for all providers
// Source: internal/embeddings/embeddings.go#L15-45

type EmbeddingProvider interface {
    // Create embeddings for texts
    CreateEmbeddings(ctx context.Context, texts []string) ([]Embedding, error)
    
    // Get model info
    GetModelInfo() ModelInfo
    
    // Close provider connection
    Close() error
}

// Embedding represents a text embedding
type Embedding struct {
    Vector     []float32
    Text       string
    Model      string
    Dimensions int
    Normalized bool
}

// ModelInfo describes embedding model
type ModelInfo struct {
    Name       string
    Dimensions int
    MaxTokens  int
    Provider   string
}
```

### OpenAI Provider Implementation

**Source:** [`internal/embeddings/providers/openai.go`](../../../internal/embeddings/providers/openai.go)

```go
package providers

// OpenAIEmbeddingProvider implements OpenAI embeddings
// Source: internal/embeddings/providers/openai.go#L1-120

type OpenAIEmbeddingProvider struct {
    client *openai.Client
    model  string
    config *Config
}

// CreateEmbeddings generates embeddings via OpenAI API
// Source: internal/embeddings/providers/openai.go#L45-95
func (p *OpenAIEmbeddingProvider) CreateEmbeddings(ctx context.Context, texts []string) ([]embeddings.Embedding, error) {
    // Batch requests for efficiency
    const batchSize = 100
    var allEmbeddings []embeddings.Embedding
    
    for i := 0; i < len(texts); i += batchSize {
        end := i + batchSize
        if end > len(texts) {
            end = len(texts)
        }
        batch := texts[i:end]
        
        resp, err := p.client.Embeddings.New(ctx, openai.EmbeddingNewParams{
            Model: p.model,
            Input: batch,
        })
        if err != nil {
            return nil, err
        }
        
        for j, data := range resp.Data {
            allEmbeddings = append(allEmbeddings, embeddings.Embedding{
                Vector:     convertToFloat32(data.Embedding),
                Text:       batch[j],
                Model:      p.model,
                Dimensions: len(data.Embedding),
                Normalized: true,
            })
        }
    }
    
    return allEmbeddings, nil
}

// GetModelInfo returns model information
func (p *OpenAIEmbeddingProvider) GetModelInfo() embeddings.ModelInfo {
    dimensions := 1536
    if p.model == "text-embedding-3-large" {
        dimensions = 3072
    }
    
    return embeddings.ModelInfo{
        Name:       p.model,
        Dimensions: dimensions,
        MaxTokens:  8191,
        Provider:   "openai",
    }
}
```

### Multi-Provider Support

**Source:** [`internal/embeddings/providers/multi.go`](../../../internal/embeddings/providers/multi.go)

```go
package providers

// MultiProvider routes embeddings to best provider
// Source: internal/embeddings/providers/multi.go#L1-100

type MultiProvider struct {
    providers map[string]embeddings.EmbeddingProvider
    strategy  RoutingStrategy
}

// CreateEmbeddings selects optimal provider
func (p *MultiProvider) CreateEmbeddings(ctx context.Context, texts []string) ([]embeddings.Embedding, error) {
    // Select provider based on:
    // - Language detection
    // - Token count
    // - Cost optimization
    // - Availability
    
    provider := p.strategy.Select(texts)
    return provider.CreateEmbeddings(ctx, texts)
}
```

---

## Vector Database Integration

### ChromaDB

**Source:** [`internal/vectordb/chromadb/`](../../../internal/vectordb/chromadb/)

```go
// ChromaDB client
// Source: internal/vectordb/chromadb/client.go

type ChromaClient struct {
    client *chromago.Client
}

// Add vectors to collection
func (c *ChromaClient) Add(ctx context.Context, collection string, vectors []Vector) error {
    // Implementation
}

// Search similar vectors
func (c *ChromaClient) Search(ctx context.Context, collection string, query []float32, n int) ([]SearchResult, error) {
    // Implementation
}
```

### Qdrant

**Source:** [`internal/vectordb/qdrant/`](../../../internal/vectordb/qdrant/)

```go
// Qdrant client
// Source: internal/vectordb/qdrant/client.go

type QdrantClient struct {
    client *qdrant.Client
}

// Search with filters
func (c *QdrantClient) Search(ctx context.Context, collection string, query []float32, filter Filter, n int) ([]SearchResult, error) {
    resp, err := c.client.Query(ctx, &qdrant.QueryPoints{
        CollectionName: collection,
        Vector:        query,
        Filter:        convertFilter(filter),
        Limit:         uint64(n),
    })
    return convertResults(resp), nil
}
```

---

## API Endpoints

### REST API

```
POST /v1/embeddings           # Create embeddings
POST /v1/embeddings/similar  # Find similar texts
GET  /v1/embeddings/models   # List available models
POST /v1/embeddings/batch    # Batch embedding
```

### Create Embeddings

```bash
curl -X POST http://localhost:7061/v1/embeddings \
  -H "Content-Type: application/json" \
  -d '{
    "model": "text-embedding-3-large",
    "input": ["Hello world", "Goodbye world"],
    "encoding_format": "float"
  }'
```

**Response:**
```json
{
  "object": "list",
  "data": [
    {
      "object": "embedding",
      "embedding": [0.0023, -0.0091, ...],
      "index": 0
    }
  ],
  "model": "text-embedding-3-large",
  "usage": {
    "prompt_tokens": 4,
    "total_tokens": 4
  }
}
```

### Similarity Search

```bash
curl -X POST http://localhost:7061/v1/embeddings/similar \
  -H "Content-Type: application/json" \
  -d '{
    "query": "machine learning",
    "collection": "documents",
    "top_k": 5,
    "threshold": 0.7
  }'
```

**Response:**
```json
{
  "results": [
    {
      "text": "Introduction to ML algorithms",
      "similarity": 0.92,
      "metadata": {"source": "ml-book.pdf"}
    },
    {
      "text": "Deep learning fundamentals",
      "similarity": 0.85,
      "metadata": {"source": "dl-guide.pdf"}
    }
  ]
}
```

---

## Caching and Optimization

### Redis Cache

**Source:** [`internal/embeddings/cache/redis.go`](../../../internal/embeddings/cache/redis.go)

```go
// RedisEmbeddingCache
// Source: internal/embeddings/cache/redis.go

type RedisEmbeddingCache struct {
    client *redis.Client
    ttl    time.Duration
}

// Get retrieves cached embedding
func (c *RedisEmbeddingCache) Get(ctx context.Context, text string, model string) (*embeddings.Embedding, error) {
    key := c.generateKey(text, model)
    data, err := c.client.Get(ctx, key).Bytes()
    if err == redis.Nil {
        return nil, nil  // Cache miss
    }
    if err != nil {
        return nil, err
    }
    return c.deserialize(data), nil
}

// Set stores embedding in cache
func (c *RedisEmbeddingCache) Set(ctx context.Context, text string, model string, emb *embeddings.Embedding) error {
    key := c.generateKey(text, model)
    data := c.serialize(emb)
    return c.client.Set(ctx, key, data, c.ttl).Err()
}
```

### Semantic Caching

```go
// SemanticCache finds similar cached embeddings
// Source: internal/cache/semantic.go

type SemanticCache struct {
    store      map[string]CachedEmbedding
    threshold  float64
}

// Get finds semantically similar cached result
func (c *SemanticCache) Get(ctx context.Context, query string) (*embeddings.Embedding, error) {
    queryHash := hashText(query)
    
    // Check exact match first
    if cached, ok := c.store[queryHash]; ok {
        return cached.Embedding, nil
    }
    
    // Check semantic similarity
    for _, cached := range c.store {
        similarity := cosineSimilarity(cached.Hash, queryHash)
        if similarity > c.threshold {
            return cached.Embedding, nil
        }
    }
    
    return nil, nil  // Cache miss
}
```

---

## CLI Agent Embeddings Usage

### Which Agents Use Embeddings?

| Agent | Embedding Usage | Provider | HelixAgent Compatible |
|-------|----------------|----------|----------------------|
| **Continue** | Code search | Local/OpenAI | ✅ |
| **Kiro** | Project memory | Local | ✅ |
| **Aider** | Repo mapping | Local | ✅ |
| **OpenHands** | Knowledge base | OpenAI | ✅ |
| **Cline** | Context retrieval | OpenAI | ✅ |
| **Claude Code** | No native | N/A | N/A |
| **Codex** | No native | N/A | N/A |

---

## Source Code Reference

### Embeddings Core Files

| Component | Source File | Lines | Description |
|-----------|-------------|-------|-------------|
| Interface | `internal/embeddings/embeddings.go` | 150 | Core types |
| OpenAI Provider | `internal/embeddings/providers/openai.go` | 120 | OpenAI client |
| Google Provider | `internal/embeddings/providers/google.go` | 110 | Google client |
| Local Provider | `internal/embeddings/providers/local.go` | 95 | Local models |
| Cache | `internal/embeddings/cache/redis.go` | 80 | Redis cache |
| Batch | `internal/embeddings/batch.go` | 90 | Batch processing |
| Normalize | `internal/embeddings/normalize.go` | 60 | Vector norm |
| ChromaDB | `internal/vectordb/chromadb/client.go` | 140 | ChromaDB client |
| Qdrant | `internal/vectordb/qdrant/client.go` | 130 | Qdrant client |
| Handler | `internal/handlers/embeddings.go` | 67 | HTTP handler |

---

## Conclusion

HelixAgent's Embeddings Protocol provides:

- ✅ Multi-provider support (OpenAI, Google, Cohere, Local)
- ✅ Vector database integration (ChromaDB, Qdrant, Pinecone)
- ✅ Caching for cost optimization
- ✅ Batch processing for efficiency
- ✅ Similarity search with filtering

**Recommendation:** Use embeddings for RAG, semantic search, and content clustering.

---

*Specification Version: 1.0*  
*Last Updated: 2026-04-03*  
*HelixAgent Commit: aa960946*
