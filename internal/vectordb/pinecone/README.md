# Pinecone Vector Database Client

This package provides a production-ready Go client for [Pinecone](https://www.pinecone.io/), a fully managed vector database service optimized for machine learning applications. The client supports both serverless and pod-based deployments with comprehensive namespace and metadata filtering capabilities.

## Overview

The Pinecone client integrates with HelixAgent's vector storage layer to provide:

- **Fully Managed Service**: No infrastructure to maintain
- **Namespace Isolation**: Logical data separation within indexes
- **Rich Metadata Filtering**: Filter vectors by metadata attributes
- **High Availability**: Built-in replication and failover
- **Serverless Option**: Pay-per-use pricing model

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    HelixAgent Application                        │
├─────────────────────────────────────────────────────────────────┤
│                    Pinecone Client                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │   Config    │  │   Client    │  │    Query Builder &      │  │
│  │  Validator  │  │   Manager   │  │    Metadata Filtering   │  │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘  │
├─────────────────────────────────────────────────────────────────┤
│                     HTTP REST Transport                          │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
                    ┌─────────────────┐
                    │  Pinecone Cloud │
                    │  (Index Host)   │
                    └─────────────────┘
```

## Cloud Configuration and API Keys

### Getting API Keys

1. Sign up at [Pinecone Console](https://app.pinecone.io/)
2. Create an API key in the console
3. Create an index (serverless or pod-based)
4. Get the index host URL from the index details

### Configuration Options

```go
type Config struct {
    APIKey      string        // Pinecone API key (required)
    Environment string        // Environment (e.g., "us-west1-gcp")
    ProjectID   string        // Project ID (optional for serverless)
    IndexHost   string        // Full index host URL (required)
    Timeout     time.Duration // Request timeout (default: 30s)
}
```

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `PINECONE_API_KEY` | API key for authentication | Yes |
| `PINECONE_INDEX_HOST` | Index host URL | Yes |
| `PINECONE_ENVIRONMENT` | Environment name | No |
| `PINECONE_PROJECT_ID` | Project ID | No |

### Example Configuration

```go
config := &pinecone.Config{
    APIKey:    os.Getenv("PINECONE_API_KEY"),
    IndexHost: "https://my-index-abc123.svc.us-west1-gcp.pinecone.io",
    Timeout:   30 * time.Second,
}

client, err := pinecone.NewClient(config, logger)
if err != nil {
    log.Fatal(err)
}

if err := client.Connect(ctx); err != nil {
    log.Fatal(err)
}
defer client.Close()
```

## Serverless vs Pod-Based

### Serverless Indexes

Best for variable workloads with automatic scaling:

```go
// Create via Pinecone Console or API:
// - Choose "Serverless" option
// - Select cloud provider and region
// - No pod configuration needed

config := &pinecone.Config{
    APIKey:    os.Getenv("PINECONE_API_KEY"),
    IndexHost: "https://my-serverless-index.svc.aped-1234-a1.pinecone.io",
}
```

**Characteristics:**
- Pay per read/write unit
- Auto-scales to zero
- Lower cost for sporadic workloads
- Slightly higher latency

### Pod-Based Indexes

Best for consistent, high-performance workloads:

```go
// Create via Pinecone Console or API:
// - Choose pod type (s1, p1, p2)
// - Select number of replicas
// - Configure shards for large datasets

config := &pinecone.Config{
    APIKey:      os.Getenv("PINECONE_API_KEY"),
    Environment: "us-west1-gcp",
    ProjectID:   "abc123",
    IndexHost:   "https://my-pod-index-abc123.svc.us-west1-gcp.pinecone.io",
}
```

**Pod Types:**
| Type | Use Case | Performance |
|------|----------|-------------|
| s1 | Starter | Low latency, small data |
| p1 | Production | Balanced performance |
| p2 | High Performance | Lowest latency, highest throughput |

## Namespace and Index Management

### Namespaces

Namespaces provide logical isolation within an index:

```go
// Upsert to specific namespace
vectors := []pinecone.Vector{
    {
        ID:     "doc-1",
        Values: embedding,
        Metadata: map[string]interface{}{
            "title": "Document 1",
        },
    },
}

resp, err := client.Upsert(ctx, vectors, "tenant-123")

// Query specific namespace
queryReq := &pinecone.QueryRequest{
    Vector:          queryEmbedding,
    TopK:            10,
    Namespace:       "tenant-123",
    IncludeMetadata: true,
}

// List namespaces
namespaces, err := client.ListNamespaces(ctx)
for _, ns := range namespaces {
    fmt.Printf("Namespace: %s\n", ns)
}
```

### Index Statistics

```go
// Get index stats (optional filter)
stats, err := client.DescribeIndexStats(ctx, nil)
fmt.Printf("Dimension: %d, Total Vectors: %d\n",
    stats.Dimension, stats.TotalVectorCount)

// Stats per namespace
for ns, nsStats := range stats.Namespaces {
    fmt.Printf("Namespace %s: %d vectors\n", ns, nsStats.VectorCount)
}

// Stats with metadata filter
filter := map[string]interface{}{
    "category": map[string]interface{}{"$eq": "tech"},
}
filteredStats, err := client.DescribeIndexStats(ctx, filter)
```

## Metadata Filtering

### Filter Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `$eq` | Equals | `{"status": {"$eq": "active"}}` |
| `$ne` | Not equals | `{"status": {"$ne": "deleted"}}` |
| `$gt` | Greater than | `{"score": {"$gt": 0.5}}` |
| `$gte` | Greater than or equal | `{"score": {"$gte": 0.5}}` |
| `$lt` | Less than | `{"score": {"$lt": 0.5}}` |
| `$lte` | Less than or equal | `{"score": {"$lte": 0.5}}` |
| `$in` | In array | `{"category": {"$in": ["a", "b"]}}` |
| `$nin` | Not in array | `{"category": {"$nin": ["c", "d"]}}` |

### Combining Filters

```go
// AND filter (implicit)
filter := map[string]interface{}{
    "category": map[string]interface{}{"$eq": "tech"},
    "year":     map[string]interface{}{"$gte": 2023},
}

// OR filter (explicit)
filter := map[string]interface{}{
    "$or": []map[string]interface{}{
        {"category": map[string]interface{}{"$eq": "tech"}},
        {"category": map[string]interface{}{"$eq": "science"}},
    },
}

// Complex nested filter
filter := map[string]interface{}{
    "$and": []map[string]interface{}{
        {"status": map[string]interface{}{"$eq": "published"}},
        {
            "$or": []map[string]interface{}{
                {"category": map[string]interface{}{"$eq": "tech"}},
                {"tags": map[string]interface{}{"$in": []string{"ml", "ai"}}},
            },
        },
    },
}
```

### Query with Filter

```go
queryReq := &pinecone.QueryRequest{
    Vector:          queryEmbedding,
    TopK:            10,
    Namespace:       "production",
    IncludeMetadata: true,
    IncludeValues:   false,
    Filter: map[string]interface{}{
        "category": map[string]interface{}{"$eq": "tech"},
        "score":    map[string]interface{}{"$gte": 0.8},
    },
}

resp, err := client.Query(ctx, queryReq)
for _, match := range resp.Matches {
    fmt.Printf("ID: %s, Score: %f, Metadata: %v\n",
        match.ID, match.Score, match.Metadata)
}
```

### Delete with Filter

```go
deleteReq := &pinecone.DeleteRequest{
    Namespace: "production",
    Filter: map[string]interface{}{
        "status": map[string]interface{}{"$eq": "archived"},
    },
}

if err := client.Delete(ctx, deleteReq); err != nil {
    log.Fatal(err)
}
```

## CRUD Operations

### Upsert Vectors

```go
// Single namespace upsert
vectors := []pinecone.Vector{
    {
        ID:     "doc-1",
        Values: []float32{0.1, 0.2, 0.3, ...},
        Metadata: map[string]interface{}{
            "title":    "Introduction to AI",
            "category": "tech",
            "year":     2024,
            "tags":     []string{"ai", "ml"},
        },
    },
    {
        ID:     "doc-2",
        Values: []float32{0.4, 0.5, 0.6, ...},
        Metadata: map[string]interface{}{
            "title":    "Machine Learning Basics",
            "category": "tech",
            "year":     2024,
        },
    },
}

resp, err := client.Upsert(ctx, vectors, "production")
fmt.Printf("Upserted: %d vectors\n", resp.UpsertedCount)

// Batch upsert (process in chunks of 100)
batchSize := 100
for i := 0; i < len(allVectors); i += batchSize {
    end := min(i+batchSize, len(allVectors))
    batch := allVectors[i:end]
    _, err := client.Upsert(ctx, batch, "production")
    if err != nil {
        log.Printf("Batch %d failed: %v", i/batchSize, err)
    }
}
```

### Query Vectors

```go
// Query by vector
queryReq := &pinecone.QueryRequest{
    Vector:          queryEmbedding,
    TopK:            10,
    Namespace:       "production",
    IncludeMetadata: true,
    IncludeValues:   true,  // Include vector values in response
}

resp, err := client.Query(ctx, queryReq)
for _, match := range resp.Matches {
    fmt.Printf("ID: %s, Score: %f\n", match.ID, match.Score)
    if match.Metadata != nil {
        fmt.Printf("  Title: %v\n", match.Metadata["title"])
    }
}

// Query by ID (find similar to existing vector)
queryReq := &pinecone.QueryRequest{
    ID:              "doc-1",  // Use existing vector as query
    TopK:            10,
    Namespace:       "production",
    IncludeMetadata: true,
}
```

### Fetch Vectors

```go
ids := []string{"doc-1", "doc-2", "doc-3"}
resp, err := client.Fetch(ctx, ids, "production")

for id, vector := range resp.Vectors {
    fmt.Printf("ID: %s, Values: %v\n", id, vector.Values[:5])
}
```

### Update Vector Metadata

```go
updateReq := &pinecone.UpdateRequest{
    ID:        "doc-1",
    Namespace: "production",
    SetMetadata: map[string]interface{}{
        "status":     "reviewed",
        "updated_at": time.Now().Unix(),
    },
}

if err := client.Update(ctx, updateReq); err != nil {
    log.Fatal(err)
}

// Update vector values
updateReq := &pinecone.UpdateRequest{
    ID:        "doc-1",
    Namespace: "production",
    Values:    newEmbedding,  // New vector values
}
```

### Delete Vectors

```go
// Delete by IDs
deleteReq := &pinecone.DeleteRequest{
    IDs:       []string{"doc-1", "doc-2"},
    Namespace: "production",
}
if err := client.Delete(ctx, deleteReq); err != nil {
    log.Fatal(err)
}

// Delete all in namespace
deleteReq := &pinecone.DeleteRequest{
    DeleteAll: true,
    Namespace: "test-namespace",
}

// Delete by filter
deleteReq := &pinecone.DeleteRequest{
    Namespace: "production",
    Filter: map[string]interface{}{
        "category": map[string]interface{}{"$eq": "deprecated"},
    },
}
```

## Best Practices

### Metadata Design

```go
// Good: Indexed fields for filtering
metadata := map[string]interface{}{
    "category":   "tech",           // Filterable
    "year":       2024,             // Filterable (numeric)
    "tags":       []string{"ai"},   // Filterable (array)
    "score":      0.95,             // Filterable (numeric)
}

// Avoid: Large text in metadata (use external storage)
// metadata["full_content"] = largeTextContent  // Bad
```

### Batch Operations

```go
// Recommended batch size: 100-1000 vectors
batchSize := 100

// Parallel batch upsert
var wg sync.WaitGroup
semaphore := make(chan struct{}, 5) // Max 5 concurrent batches

for i := 0; i < len(vectors); i += batchSize {
    wg.Add(1)
    semaphore <- struct{}{}

    go func(start int) {
        defer wg.Done()
        defer func() { <-semaphore }()

        end := min(start+batchSize, len(vectors))
        batch := vectors[start:end]
        client.Upsert(ctx, batch, namespace)
    }(i)
}
wg.Wait()
```

### Error Handling

```go
resp, err := client.Query(ctx, queryReq)
if err != nil {
    if strings.Contains(err.Error(), "not connected") {
        // Reconnect
        client.Connect(ctx)
    } else if strings.Contains(err.Error(), "401") {
        // Invalid API key
        log.Fatal("Invalid Pinecone API key")
    } else if strings.Contains(err.Error(), "429") {
        // Rate limited - implement backoff
        time.Sleep(time.Second)
        // Retry...
    }
}
```

### Namespace Strategy

```go
// Multi-tenant isolation
namespace := fmt.Sprintf("tenant-%s", tenantID)
client.Upsert(ctx, vectors, namespace)

// Environment separation
devNamespace := "dev"
prodNamespace := "prod"

// Versioned data
namespace := fmt.Sprintf("v%d", schemaVersion)
```

## Health Check

```go
// Health check via describe_index_stats
if err := client.HealthCheck(ctx); err != nil {
    log.Printf("Pinecone unhealthy: %v", err)
}

// Check connection status
if !client.IsConnected() {
    client.Connect(ctx)
}
```

## Testing

```bash
# Unit tests
go test -v ./internal/vectordb/pinecone/...

# Integration tests (requires Pinecone API key)
PINECONE_API_KEY=your-key PINECONE_INDEX_HOST=your-host \
go test -v ./internal/vectordb/pinecone/... -run Integration

# Comprehensive tests
go test -v ./internal/vectordb/pinecone/... -run Comprehensive
```

## Pricing Considerations

### Serverless
- **Read units**: Per query operation
- **Write units**: Per upsert operation
- **Storage**: Per GB stored

### Pod-Based
- **Pod hours**: Per pod per hour
- **Replicas**: Multiply by replica count
- **Storage**: Included in pod pricing

### Cost Optimization

```go
// Reduce query costs
queryReq := &pinecone.QueryRequest{
    TopK:            5,      // Limit results
    IncludeValues:   false,  // Don't fetch vectors
    IncludeMetadata: true,   // Only fetch metadata
}

// Efficient batch operations
// Fewer large batches > many small requests
```

## Related Files

- `client.go` - Main client implementation
- `client_test.go` - Unit tests
- `client_comprehensive_test.go` - Comprehensive tests
