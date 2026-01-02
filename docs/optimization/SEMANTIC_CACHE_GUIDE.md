# Semantic Cache Guide

The semantic cache provides intelligent response caching based on vector similarity, inspired by GPTCache.

## Overview

Unlike traditional key-based caching, semantic caching uses embedding vectors to find similar queries and return cached responses. This enables cache hits even when queries are semantically similar but textually different.

## Features

- **Vector Similarity Search**: Cosine and Euclidean distance metrics
- **Configurable Threshold**: Control similarity matching strictness
- **Multiple Eviction Policies**: LRU, TTL, and relevance-based
- **Metadata Support**: Store additional context with cache entries
- **Statistics Tracking**: Monitor hit rates and performance

## Configuration

```yaml
optimization:
  semantic_cache:
    enabled: true
    similarity_threshold: 0.85    # 0.0-1.0, higher = stricter matching
    max_entries: 10000            # Maximum cache entries
    ttl: "24h"                    # Time-to-live for entries
    embedding_model: "text-embedding-3-small"
    eviction_policy: "lru_with_relevance"
```

## Basic Usage

```go
import (
    "github.com/superagent/superagent/internal/optimization/gptcache"
)

// Create cache
cache := gptcache.NewSemanticCache(
    gptcache.WithSimilarityThreshold(0.85),
    gptcache.WithMaxEntries(10000),
    gptcache.WithTTL(24 * time.Hour),
)

// Store a response
embedding := getEmbedding("What is machine learning?")
cache.Set(ctx, "What is machine learning?", "Machine learning is...", embedding, nil)

// Query with a similar prompt
queryEmbedding := getEmbedding("Explain machine learning")
hit, err := cache.Get(ctx, queryEmbedding)
if hit != nil && hit.Entry != nil {
    // Cache hit! Use hit.Entry.Response
    fmt.Println("Cached response:", hit.Entry.Response)
}
```

## Similarity Threshold

The similarity threshold controls how strict the matching is:

| Threshold | Behavior |
|-----------|----------|
| 0.95+ | Very strict - nearly identical queries only |
| 0.85-0.95 | Balanced - semantically similar queries |
| 0.70-0.85 | Loose - broadly related queries |
| <0.70 | Very loose - may produce false positives |

Recommended: **0.85** for general use cases.

## Eviction Policies

### LRU (Least Recently Used)

```go
cache := gptcache.NewSemanticCache(
    gptcache.WithEvictionPolicy("lru"),
)
```

Evicts entries that haven't been accessed recently.

### TTL (Time-To-Live)

```go
cache := gptcache.NewSemanticCache(
    gptcache.WithTTL(24 * time.Hour),
)
```

Evicts entries after a specified duration.

### Relevance-Based

```go
cache := gptcache.NewSemanticCache(
    gptcache.WithEvictionPolicy("lru_with_relevance"),
)
```

Combines LRU with relevance scoring - keeps high-value entries longer.

## Advanced Features

### Metadata

Store additional context with cache entries:

```go
metadata := map[string]interface{}{
    "model": "gpt-4",
    "temperature": 0.7,
    "user_id": "user123",
}
cache.Set(ctx, query, response, embedding, metadata)
```

### Cache Statistics

```go
stats := cache.Stats(ctx)
fmt.Printf("Entries: %d\n", stats.TotalEntries)
fmt.Printf("Hit Rate: %.2f%%\n", stats.HitRate * 100)
fmt.Printf("Avg Similarity: %.3f\n", stats.AverageSimilarity)
```

### Invalidation

```go
// Remove by query hash
cache.Remove(ctx, queryHash)

// Invalidate by criteria
cache.Invalidate(ctx, gptcache.InvalidationCriteria{
    OlderThan: time.Now().Add(-7 * 24 * time.Hour),
    Metadata: map[string]interface{}{"model": "gpt-3.5"},
})

// Clear all
cache.Clear(ctx)
```

### Top-K Similar

Find multiple similar entries:

```go
hits, err := cache.GetTopK(ctx, embedding, 5)
for _, hit := range hits {
    fmt.Printf("Similarity: %.3f - %s\n", hit.Similarity, hit.Entry.Response[:50])
}
```

## Integration with OptimizationService

The semantic cache is automatically used when enabled:

```go
svc, _ := optimization.NewService(config)

// OptimizeRequest checks cache first
optimized, _ := svc.OptimizeRequest(ctx, prompt, embedding)
if optimized.CacheHit {
    return optimized.CachedResponse // Use cached response
}

// Process with LLM...
response := callLLM(prompt)

// OptimizeResponse caches the response
svc.OptimizeResponse(ctx, response, embedding, prompt, nil)
```

## Similarity Functions

### Cosine Similarity

Default metric, range [-1, 1], normalized to [0, 1]:

```go
import "github.com/superagent/superagent/internal/optimization/gptcache"

similarity := gptcache.CosineSimilarity(vec1, vec2)
```

### Euclidean Distance

Converts distance to similarity score:

```go
distance := gptcache.EuclideanDistance(vec1, vec2)
// Lower distance = more similar
```

## Best Practices

1. **Choose Appropriate Threshold**: Start with 0.85 and adjust based on false positive/negative rates

2. **Use Consistent Embeddings**: Always use the same embedding model for storage and retrieval

3. **Monitor Hit Rates**: Track cache effectiveness with statistics

4. **Set Reasonable TTL**: Balance freshness vs cache efficiency

5. **Consider Metadata**: Store model/temperature for consistent retrieval

## Performance Characteristics

| Operation | Complexity | Notes |
|-----------|------------|-------|
| Get | O(n) | Linear scan of all entries |
| Set | O(1) | Constant time insertion |
| GetTopK | O(n log k) | Partial sort for top-k |
| Eviction | O(1) | Amortized with LRU |

For very large caches (100k+ entries), consider using an external vector database.
