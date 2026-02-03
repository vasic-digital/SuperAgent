# Qdrant Vector Database Client

This package provides a production-ready Go client for [Qdrant](https://qdrant.tech/), a high-performance vector similarity search engine. The client supports both HTTP and gRPC protocols, connection pooling, and comprehensive CRUD operations for vector data.

## Overview

The Qdrant client integrates with HelixAgent's vector storage layer to provide:

- **High-Performance Vector Search**: Optimized similarity search with multiple distance metrics
- **Collection Management**: Full lifecycle management of vector collections
- **Batch Operations**: Efficient bulk upsert and search capabilities
- **Flexible Filtering**: Rich metadata filtering for precise queries
- **Production Ready**: Connection pooling, health checks, and graceful degradation

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    HelixAgent Application                        │
├─────────────────────────────────────────────────────────────────┤
│                     Qdrant Client                                │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │   Config    │  │   Client    │  │    Search Options       │  │
│  │  Validator  │  │   Manager   │  │    & Filtering          │  │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘  │
├─────────────────────────────────────────────────────────────────┤
│                     HTTP/gRPC Transport                          │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
                    ┌─────────────────┐
                    │  Qdrant Server  │
                    │   (Port 6333)   │
                    └─────────────────┘
```

## Configuration

### Configuration Options

```go
type Config struct {
    // Connection settings
    Host     string        // Qdrant server hostname (default: "localhost")
    HTTPPort int           // HTTP API port (default: 6333)
    GRPCPort int           // gRPC port (default: 6334)
    APIKey   string        // API key for authentication
    UseGRPC  bool          // Use gRPC instead of HTTP

    // Connection options
    Timeout    time.Duration // Request timeout (default: 30s)
    MaxRetries int           // Maximum retry attempts (default: 3)
    RetryDelay time.Duration // Delay between retries (default: 1s)

    // Search defaults
    DefaultLimit   int     // Default search result limit (default: 10)
    ScoreThreshold float32 // Minimum score threshold (default: 0.0)
    WithPayload    bool    // Include payload in results (default: true)
    WithVectors    bool    // Include vectors in results (default: false)
}
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `QDRANT_HOST` | Qdrant server hostname | `localhost` |
| `QDRANT_HTTP_PORT` | HTTP API port | `6333` |
| `QDRANT_GRPC_PORT` | gRPC port | `6334` |
| `QDRANT_API_KEY` | API key for authentication | (none) |
| `QDRANT_TIMEOUT` | Request timeout | `30s` |

### Example Configuration

```go
config := &qdrant.Config{
    Host:           "localhost",
    HTTPPort:       6333,
    GRPCPort:       6334,
    APIKey:         os.Getenv("QDRANT_API_KEY"),
    UseGRPC:        false,
    Timeout:        30 * time.Second,
    MaxRetries:     3,
    RetryDelay:     1 * time.Second,
    DefaultLimit:   10,
    ScoreThreshold: 0.0,
    WithPayload:    true,
    WithVectors:    false,
}
```

## CRUD Operations

### Connect and Health Check

```go
// Create client
client, err := qdrant.NewClient(config, logger)
if err != nil {
    log.Fatal(err)
}

// Connect to Qdrant
ctx := context.Background()
if err := client.Connect(ctx); err != nil {
    log.Fatal(err)
}
defer client.Close()

// Health check
if err := client.HealthCheck(ctx); err != nil {
    log.Printf("Qdrant unhealthy: %v", err)
}
```

### Collection Management

```go
// Create collection with configuration
collConfig := qdrant.DefaultCollectionConfig("my_collection", 1536).
    WithDistance(qdrant.DistanceCosine).
    WithOnDiskPayload().
    WithShards(2).
    WithReplication(3)

if err := client.CreateCollection(ctx, collConfig); err != nil {
    log.Fatal(err)
}

// Check if collection exists
exists, err := client.CollectionExists(ctx, "my_collection")

// List all collections
collections, err := client.ListCollections(ctx)

// Get collection info
info, err := client.GetCollectionInfo(ctx, "my_collection")
fmt.Printf("Collection status: %s, Points: %d\n", info.Status, info.PointsCount)

// Delete collection
if err := client.DeleteCollection(ctx, "my_collection"); err != nil {
    log.Fatal(err)
}
```

### Upsert Operations

```go
// Single point upsert
points := []qdrant.Point{
    {
        ID:     "doc-1",
        Vector: []float32{0.1, 0.2, 0.3, ...}, // 1536-dimensional vector
        Payload: map[string]interface{}{
            "title":    "Document Title",
            "content":  "Document content...",
            "category": "tech",
            "timestamp": time.Now().Unix(),
        },
    },
}

if err := client.UpsertPoints(ctx, "my_collection", points); err != nil {
    log.Fatal(err)
}

// Batch upsert with auto-generated IDs
batchPoints := make([]qdrant.Point, 1000)
for i := range batchPoints {
    batchPoints[i] = qdrant.Point{
        // ID will be auto-generated as UUID
        Vector:  generateEmbedding(documents[i]),
        Payload: map[string]interface{}{"doc_index": i},
    }
}
if err := client.UpsertPoints(ctx, "my_collection", batchPoints); err != nil {
    log.Fatal(err)
}
```

### Search Operations

```go
// Basic vector search
queryVector := []float32{0.15, 0.25, 0.35, ...}
opts := qdrant.DefaultSearchOptions().
    WithLimit(10).
    WithScoreThreshold(0.7)

results, err := client.Search(ctx, "my_collection", queryVector, opts)
for _, result := range results {
    fmt.Printf("ID: %s, Score: %f, Payload: %v\n",
        result.ID, result.Score, result.Payload)
}

// Search with metadata filter
filterOpts := qdrant.DefaultSearchOptions().
    WithLimit(20).
    WithFilter(map[string]interface{}{
        "must": []map[string]interface{}{
            {"key": "category", "match": map[string]interface{}{"value": "tech"}},
        },
    })

filteredResults, err := client.Search(ctx, "my_collection", queryVector, filterOpts)

// Batch search (multiple queries in one request)
queryVectors := [][]float32{
    queryVector1,
    queryVector2,
    queryVector3,
}
batchResults, err := client.SearchBatch(ctx, "my_collection", queryVectors, opts)
```

### Get Operations

```go
// Get single point by ID
point, err := client.GetPoint(ctx, "my_collection", "doc-1")

// Get multiple points by IDs
ids := []string{"doc-1", "doc-2", "doc-3"}
points, err := client.GetPoints(ctx, "my_collection", ids)

// Scroll through all points with pagination
var offset *string
for {
    points, nextOffset, err := client.Scroll(ctx, "my_collection", 100, offset, nil)
    if err != nil {
        log.Fatal(err)
    }

    for _, p := range points {
        processPoint(p)
    }

    if nextOffset == nil {
        break
    }
    offset = nextOffset
}

// Count points
count, err := client.CountPoints(ctx, "my_collection", nil)

// Count with filter
filterCount, err := client.CountPoints(ctx, "my_collection", map[string]interface{}{
    "must": []map[string]interface{}{
        {"key": "category", "match": map[string]interface{}{"value": "tech"}},
    },
})
```

### Delete Operations

```go
// Delete by IDs
ids := []string{"doc-1", "doc-2", "doc-3"}
if err := client.DeletePoints(ctx, "my_collection", ids); err != nil {
    log.Fatal(err)
}
```

## Distance Metrics

The client supports four distance metrics:

| Metric | Constant | Description |
|--------|----------|-------------|
| Cosine | `DistanceCosine` | Cosine similarity (normalized dot product) |
| Euclidean | `DistanceEuclid` | L2 distance (Euclidean) |
| Dot Product | `DistanceDot` | Inner product (unnormalized) |
| Manhattan | `DistanceManhattan` | L1 distance (city block) |

```go
// Use cosine similarity for text embeddings
config := qdrant.DefaultCollectionConfig("embeddings", 1536).
    WithDistance(qdrant.DistanceCosine)

// Use Euclidean for image feature vectors
config := qdrant.DefaultCollectionConfig("images", 512).
    WithDistance(qdrant.DistanceEuclid)
```

## Performance Tuning

### Collection Configuration

```go
config := qdrant.DefaultCollectionConfig("high_performance", 1536).
    WithDistance(qdrant.DistanceCosine).
    WithOnDiskPayload().                    // Store payload on disk for large datasets
    WithIndexingThreshold(50000).           // Delay indexing until 50k points
    WithShards(4).                          // Distribute across 4 shards
    WithReplication(2)                      // Replicate for high availability
```

### Search Optimization

```go
// Enable approximate search with score threshold
opts := qdrant.DefaultSearchOptions().
    WithLimit(10).
    WithScoreThreshold(0.5).  // Filter low-quality results early
    WithOffset(0)             // Use offset for pagination
```

### Batch Operations

For high-throughput scenarios, use batch operations:

```go
// Batch upsert: Process in chunks of 100-1000 points
chunkSize := 500
for i := 0; i < len(allPoints); i += chunkSize {
    end := min(i+chunkSize, len(allPoints))
    chunk := allPoints[i:end]

    if err := client.UpsertPoints(ctx, "collection", chunk); err != nil {
        log.Printf("Failed to upsert chunk %d: %v", i/chunkSize, err)
    }
}

// Batch search: Multiple queries in single request
queryVectors := make([][]float32, len(queries))
for i, q := range queries {
    queryVectors[i] = embedQuery(q)
}
results, err := client.SearchBatch(ctx, "collection", queryVectors, opts)
```

### Connection Tuning

```go
config := &qdrant.Config{
    Host:       "qdrant.example.com",
    HTTPPort:   6333,
    Timeout:    60 * time.Second,  // Increase for large batch operations
    MaxRetries: 5,
    RetryDelay: 2 * time.Second,
}
```

## Snapshots and Backup

```go
// Create a snapshot for backup
snapshotName, err := client.CreateSnapshot(ctx, "my_collection")
fmt.Printf("Created snapshot: %s\n", snapshotName)

// Snapshots are stored in Qdrant's snapshot directory
// Use the Qdrant REST API to list and download snapshots
```

## Monitoring

```go
// Get telemetry metrics
metrics, err := client.GetMetrics(ctx)
if err != nil {
    log.Printf("Failed to get metrics: %v", err)
} else {
    fmt.Printf("Metrics: %v\n", metrics)
}

// Wait for collection to be ready
if err := client.WaitForCollection(ctx, "my_collection", 30*time.Second); err != nil {
    log.Printf("Collection not ready: %v", err)
}
```

## Error Handling

The client wraps all errors with context:

```go
client, err := qdrant.NewClient(config, logger)
if err != nil {
    // Config validation error
    log.Fatalf("Invalid config: %v", err)
}

if err := client.Connect(ctx); err != nil {
    // Connection error
    log.Fatalf("Failed to connect: %v", err)
}

results, err := client.Search(ctx, "collection", vector, opts)
if err != nil {
    // Check for specific error types
    if strings.Contains(err.Error(), "not connected") {
        // Handle disconnection
    } else if strings.Contains(err.Error(), "request failed") {
        // Handle API error
    }
}
```

## Thread Safety

The client is thread-safe and uses `sync.RWMutex` for concurrent access:

```go
// Safe for concurrent use
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(idx int) {
        defer wg.Done()
        results, err := client.Search(ctx, "collection", vectors[idx], opts)
        // Process results...
    }(i)
}
wg.Wait()
```

## Integration with HelixAgent

The Qdrant client integrates with HelixAgent's RAG and embedding systems:

```go
// Used by internal/rag for hybrid retrieval
// Used by internal/embedding for vector storage
// Used by internal/memory for persistent memory storage

// Example: RAG integration
import (
    "dev.helix.agent/internal/vectordb/qdrant"
    "dev.helix.agent/internal/embedding"
)

// Store document embeddings
embeddings, err := embeddingProvider.Embed(ctx, documents)
points := make([]qdrant.Point, len(documents))
for i, doc := range documents {
    points[i] = qdrant.Point{
        ID:     doc.ID,
        Vector: embeddings[i],
        Payload: map[string]interface{}{
            "content": doc.Content,
            "source":  doc.Source,
        },
    }
}
client.UpsertPoints(ctx, "knowledge_base", points)

// Query for relevant documents
queryEmbed, _ := embeddingProvider.Embed(ctx, []string{query})
results, _ := client.Search(ctx, "knowledge_base", queryEmbed[0], opts)
```

## Testing

Run tests with:

```bash
# Unit tests
go test -v ./internal/vectordb/qdrant/...

# With infrastructure (requires running Qdrant)
QDRANT_HOST=localhost QDRANT_HTTP_PORT=6333 \
go test -v ./internal/vectordb/qdrant/... -run Integration

# Benchmarks
go test -bench=. ./internal/vectordb/qdrant/...
```

## Related Files

- `client.go` - Main client implementation
- `config.go` - Configuration types and validation
- `client_test.go` - Unit tests
- `client_mock_test.go` - Mock-based tests
- `config_test.go` - Configuration tests
- `config_comprehensive_test.go` - Comprehensive config tests
- `qdrant_bench_test.go` - Benchmark tests
