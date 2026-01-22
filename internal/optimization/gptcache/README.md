# GPTCache Package

This package provides semantic similarity-based caching for LLM queries to reduce latency and API costs.

## Overview

GPTCache uses embedding vectors to find semantically similar queries in the cache, returning cached responses for queries that are similar enough to previous ones.

## Components

### Semantic Cache (`semantic_cache.go`)

Core caching implementation with similarity-based lookup:

```go
cache := gptcache.NewSemanticCache(&gptcache.CacheConfig{
    SimilarityThreshold: 0.95,
    MaxEntries:          10000,
    TTL:                 24 * time.Hour,
})

// Store a response
cache.Set(ctx, query, embedding, response)

// Look up similar queries
hit, err := cache.Get(ctx, query, embedding)
if err == nil {
    // Cache hit with similarity score
    fmt.Printf("Hit: %.2f similarity\n", hit.Similarity)
}
```

### Similarity Functions (`similarity.go`)

Vector similarity calculations:
- **Cosine Similarity**: Angle-based similarity (default)
- **Euclidean Distance**: Distance-based similarity
- **Dot Product**: Magnitude-weighted similarity

### Eviction Strategies (`eviction.go`)

Cache eviction policies:
- **LRU**: Least Recently Used
- **LFU**: Least Frequently Used
- **FIFO**: First In First Out
- **TTL**: Time-To-Live based

## Data Types

### CacheEntry

```go
type CacheEntry struct {
    ID          string                 // Unique entry ID
    Query       string                 // Original query text
    QueryHash   string                 // SHA256 hash of query
    Response    string                 // Cached response
    Embedding   []float64              // Query embedding vector
    Metadata    map[string]interface{} // Custom metadata
    CreatedAt   time.Time              // Creation timestamp
    AccessedAt  time.Time              // Last access timestamp
    AccessCount int                    // Access frequency counter
}
```

### CacheStats

```go
type CacheStats struct {
    TotalEntries  int     // Current entries in cache
    Hits          int64   // Total cache hits
    Misses        int64   // Total cache misses
    HitRate       float64 // Hit/Miss ratio
    AvgSimilarity float64 // Average hit similarity
}
```

## Configuration

```go
type CacheConfig struct {
    // SimilarityThreshold is minimum similarity for cache hit (0.0-1.0)
    SimilarityThreshold float64

    // MaxEntries is maximum cache entries
    MaxEntries int

    // TTL is time-to-live for entries
    TTL time.Duration

    // EvictionStrategy is the eviction policy
    EvictionStrategy EvictionType

    // EmbeddingDimension is the expected embedding size
    EmbeddingDimension int
}
```

## Usage

### Basic Usage

```go
import "dev.helix.agent/internal/optimization/gptcache"

// Create cache with configuration
config := &gptcache.CacheConfig{
    SimilarityThreshold: 0.92,
    MaxEntries:          5000,
    TTL:                 12 * time.Hour,
    EvictionStrategy:    gptcache.EvictionLRU,
}

cache := gptcache.NewSemanticCache(config)

// Your embedding provider
embedding, _ := embeddingProvider.Embed(ctx, "What is Go?")

// Try cache lookup
hit, err := cache.Get(ctx, "What is Go?", embedding)
if err == gptcache.ErrCacheMiss {
    // Cache miss - call LLM
    response, _ := llm.Complete(ctx, "What is Go?")
    cache.Set(ctx, "What is Go?", embedding, response)
} else {
    // Cache hit - use cached response
    return hit.Entry.Response
}
```

### With Metadata

```go
cache.SetWithMetadata(ctx, query, embedding, response, map[string]interface{}{
    "model":       "claude-sonnet-4-20250514",
    "tokens_used": 150,
    "provider":    "claude",
})
```

## Performance

- O(n) similarity search for exact matching
- Configurable similarity thresholds
- Memory-efficient with LRU eviction
- Thread-safe operations

## Testing

```bash
go test -v ./internal/optimization/gptcache/...
```

## Files

- `semantic_cache.go` - Core semantic cache implementation
- `similarity.go` - Vector similarity functions
- `eviction.go` - Eviction strategy implementations
- `config.go` - Configuration types
- `types.go` - Data type definitions
