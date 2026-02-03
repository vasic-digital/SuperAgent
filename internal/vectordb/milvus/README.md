# Milvus Vector Database Client

This package provides a production-ready Go client for [Milvus](https://milvus.io/), a cloud-native vector database designed for scalable similarity search. The client supports both standalone and cluster deployments, with comprehensive collection management and hybrid search capabilities.

## Overview

The Milvus client integrates with HelixAgent's vector storage layer to provide:

- **Scalable Vector Storage**: Support for billions of vectors with horizontal scaling
- **Hybrid Search**: Combine vector similarity with scalar filtering
- **Collection Management**: Full schema definition and index management
- **Memory Management**: Load/release collections for memory optimization
- **Zilliz Cloud Support**: Compatible with Zilliz Cloud managed service

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    HelixAgent Application                        │
├─────────────────────────────────────────────────────────────────┤
│                     Milvus Client                                │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │   Config    │  │   Client    │  │    Schema Builder       │  │
│  │  Validator  │  │   Manager   │  │    & Index Config       │  │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘  │
├─────────────────────────────────────────────────────────────────┤
│                     REST API Transport                           │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
                  ┌───────────────────────┐
                  │    Milvus Server      │
                  │   (Standalone/Cluster)│
                  │      Port 19530       │
                  └───────────────────────┘
```

## Cluster vs Standalone Setup

### Standalone Deployment

For development and small-scale production:

```yaml
# docker-compose.yml
services:
  milvus:
    image: milvusdb/milvus:latest
    ports:
      - "19530:19530"
    volumes:
      - ./milvus-data:/var/lib/milvus
    environment:
      ETCD_ENDPOINTS: etcd:2379
      MINIO_ADDRESS: minio:9000
```

```go
config := &milvus.Config{
    Host:   "localhost",
    Port:   19530,
    DBName: "default",
}
```

### Cluster Deployment

For high availability and scalability:

```yaml
# Kubernetes deployment with Helm
helm install milvus milvus/milvus \
  --set cluster.enabled=true \
  --set rootCoord.replicas=2 \
  --set queryCoord.replicas=2 \
  --set dataCoord.replicas=2
```

```go
config := &milvus.Config{
    Host:     "milvus-proxy.milvus.svc.cluster.local",
    Port:     19530,
    Username: os.Getenv("MILVUS_USERNAME"),
    Password: os.Getenv("MILVUS_PASSWORD"),
    DBName:   "production",
    Secure:   true,
}
```

### Zilliz Cloud (Managed)

For fully managed cloud deployment:

```go
config := &milvus.Config{
    Host:   "your-instance.zillizcloud.com",
    Port:   443,
    Token:  os.Getenv("ZILLIZ_TOKEN"),
    Secure: true,
    DBName: "default",
}
```

## Configuration

### Configuration Options

```go
type Config struct {
    Host     string        // Milvus server hostname
    Port     int           // Server port (default: 19530)
    Username string        // Username for authentication
    Password string        // Password for authentication
    DBName   string        // Database name (default: "default")
    Secure   bool          // Use HTTPS/TLS
    Token    string        // Bearer token (for Zilliz Cloud)
    Timeout  time.Duration // Request timeout (default: 30s)
}
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `MILVUS_HOST` | Server hostname | `localhost` |
| `MILVUS_PORT` | Server port | `19530` |
| `MILVUS_USERNAME` | Username | (none) |
| `MILVUS_PASSWORD` | Password | (none) |
| `MILVUS_DB_NAME` | Database name | `default` |
| `ZILLIZ_TOKEN` | Zilliz Cloud token | (none) |

## Collection Management

### Create Collection with Schema

```go
// Quick setup (auto-generates schema)
req := &milvus.CreateCollectionRequest{
    CollectionName: "documents",
    Dimension:      1536,
    MetricType:     milvus.MetricTypeCosine,
    PrimaryField:   "id",
    VectorField:    "embedding",
    IDType:         milvus.DataTypeVarChar,
}
if err := client.CreateCollection(ctx, req); err != nil {
    log.Fatal(err)
}

// Full schema definition
schema := milvus.CollectionSchema{
    CollectionName: "documents",
    Description:    "Document embeddings with metadata",
    Fields: []milvus.FieldSchema{
        {
            FieldName:    "id",
            DataType:     milvus.DataTypeVarChar,
            IsPrimaryKey: true,
            Params:       map[string]interface{}{"max_length": 128},
        },
        {
            FieldName: "embedding",
            DataType:  milvus.DataTypeFloatVector,
            Params:    map[string]interface{}{"dim": 1536},
        },
        {
            FieldName:   "title",
            DataType:    milvus.DataTypeVarChar,
            Params:      map[string]interface{}{"max_length": 512},
        },
        {
            FieldName:   "category",
            DataType:    milvus.DataTypeVarChar,
            IsPartition: true,
            Params:      map[string]interface{}{"max_length": 64},
        },
        {
            FieldName: "metadata",
            DataType:  milvus.DataTypeJSON,
        },
    },
}

req := &milvus.CreateCollectionRequest{
    CollectionName: "documents",
    Schema:         schema,
}
if err := client.CreateCollection(ctx, req); err != nil {
    log.Fatal(err)
}
```

### List and Describe Collections

```go
// List all collections
collections, err := client.ListCollections(ctx)
for _, name := range collections {
    fmt.Println(name)
}

// Get collection details
info, err := client.DescribeCollection(ctx, "documents")
fmt.Printf("Name: %s, Fields: %d, Shards: %d\n",
    info.CollectionName, len(info.Fields), info.ShardsNum)
```

### Drop Collection

```go
if err := client.DropCollection(ctx, "documents"); err != nil {
    log.Fatal(err)
}
```

## Index Management

### Index Types

| Index Type | Constant | Best For |
|------------|----------|----------|
| IVF_FLAT | `IndexTypeIVFFlat` | Balanced accuracy/speed |
| IVF_SQ8 | `IndexTypeIVFSQ8` | Memory-efficient |
| IVF_PQ | `IndexTypeIVFPQ` | Large-scale datasets |
| HNSW | `IndexTypeHNSW` | High recall requirements |
| AUTOINDEX | `IndexTypeAutoIndex` | Automatic optimization |

### Create Index

```go
// Create HNSW index for high accuracy
req := &milvus.CreateIndexRequest{
    CollectionName: "documents",
    FieldName:      "embedding",
    IndexName:      "embedding_hnsw_idx",
    IndexType:      milvus.IndexTypeHNSW,
    MetricType:     milvus.MetricTypeCosine,
    Params: map[string]interface{}{
        "M":              16,   // Max connections per node
        "efConstruction": 256,  // Build-time search width
    },
}
if err := client.CreateIndex(ctx, req); err != nil {
    log.Fatal(err)
}

// Create IVF_FLAT index for balanced performance
req := &milvus.CreateIndexRequest{
    CollectionName: "documents",
    FieldName:      "embedding",
    IndexType:      milvus.IndexTypeIVFFlat,
    MetricType:     milvus.MetricTypeL2,
    Params: map[string]interface{}{
        "nlist": 1024,  // Number of cluster units
    },
}
```

## Data Operations

### Insert Data

```go
// Insert records
data := []map[string]interface{}{
    {
        "id":        "doc-1",
        "embedding": []float32{0.1, 0.2, 0.3, ...},
        "title":     "Introduction to AI",
        "category":  "tech",
        "metadata":  map[string]interface{}{"author": "John"},
    },
    {
        "id":        "doc-2",
        "embedding": []float32{0.4, 0.5, 0.6, ...},
        "title":     "Machine Learning Basics",
        "category":  "tech",
        "metadata":  map[string]interface{}{"author": "Jane"},
    },
}

resp, err := client.Insert(ctx, "documents", data)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Inserted %d records\n", resp.InsertCount)
```

### Get by IDs

```go
ids := []string{"doc-1", "doc-2"}
entities, err := client.Get(ctx, "documents", ids, []string{"id", "title", "category"})
for _, entity := range entities {
    fmt.Printf("ID: %v, Title: %v\n", entity["id"], entity["title"])
}
```

### Query with Filter

```go
req := &milvus.QueryRequest{
    CollectionName: "documents",
    Filter:         "category == 'tech' and metadata['author'] == 'John'",
    OutputFields:   []string{"id", "title", "category"},
    Limit:          100,
}

results, err := client.Query(ctx, req)
for _, row := range results {
    fmt.Printf("%v\n", row)
}
```

### Delete Data

```go
// Delete by IDs
if err := client.Delete(ctx, "documents", "", []string{"doc-1", "doc-2"}); err != nil {
    log.Fatal(err)
}

// Delete by filter
if err := client.Delete(ctx, "documents", "category == 'outdated'", nil); err != nil {
    log.Fatal(err)
}
```

## Hybrid Search Capabilities

### Vector Similarity Search

```go
req := &milvus.SearchRequest{
    CollectionName: "documents",
    Data:           [][]float32{queryVector},
    AnnsField:      "embedding",
    Limit:          10,
    OutputFields:   []string{"id", "title", "category"},
    SearchParams: map[string]interface{}{
        "ef": 64,  // HNSW search width
    },
}

resp, err := client.Search(ctx, req)
for _, results := range resp.Results {
    for _, result := range results {
        fmt.Printf("ID: %v, Distance: %f\n", result.ID, result.Distance)
    }
}
```

### Hybrid Search (Vector + Scalar Filter)

```go
req := &milvus.SearchRequest{
    CollectionName: "documents",
    Data:           [][]float32{queryVector},
    AnnsField:      "embedding",
    Limit:          10,
    Filter:         "category == 'tech' and metadata['year'] >= 2023",
    OutputFields:   []string{"id", "title", "category", "metadata"},
}

resp, err := client.Search(ctx, req)
```

### Multi-Vector Search

```go
// Search with multiple query vectors
queryVectors := [][]float32{
    queryVector1,
    queryVector2,
    queryVector3,
}

req := &milvus.SearchRequest{
    CollectionName: "documents",
    Data:           queryVectors,
    AnnsField:      "embedding",
    Limit:          5,
}

resp, err := client.Search(ctx, req)
// resp.Results[0] = results for queryVector1
// resp.Results[1] = results for queryVector2
// etc.
```

## Memory Management

### Load Collection

Collections must be loaded into memory before searching:

```go
// Load collection
if err := client.LoadCollection(ctx, "documents"); err != nil {
    log.Fatal(err)
}

// Check load state
state, err := client.GetLoadState(ctx, "documents")
fmt.Printf("Load state: %s\n", state) // "Loaded", "Loading", "NotLoad"

// Wait for loading to complete
for {
    state, _ := client.GetLoadState(ctx, "documents")
    if state == "Loaded" {
        break
    }
    time.Sleep(time.Second)
}
```

### Release Collection

Free memory when collection is not needed:

```go
if err := client.ReleaseCollection(ctx, "documents"); err != nil {
    log.Fatal(err)
}
```

## Metric Types

| Metric | Constant | Description |
|--------|----------|-------------|
| L2 | `MetricTypeL2` | Euclidean distance |
| IP | `MetricTypeIP` | Inner product |
| COSINE | `MetricTypeCosine` | Cosine similarity |

```go
// Use cosine for text embeddings (normalized)
req := &milvus.CreateIndexRequest{
    MetricType: milvus.MetricTypeCosine,
}

// Use L2 for image features
req := &milvus.CreateIndexRequest{
    MetricType: milvus.MetricTypeL2,
}
```

## Data Types

| Type | Constant | Description |
|------|----------|-------------|
| Int64 | `DataTypeInt64` | 64-bit integer |
| VarChar | `DataTypeVarChar` | Variable-length string |
| Float | `DataTypeFloat` | 32-bit float |
| Double | `DataTypeDouble` | 64-bit float |
| Bool | `DataTypeBool` | Boolean |
| JSON | `DataTypeJSON` | JSON document |
| FloatVector | `DataTypeFloatVector` | Float32 vector |
| BinaryVector | `DataTypeBinaryVector` | Binary vector |

## Error Handling

```go
resp, err := client.Search(ctx, req)
if err != nil {
    if strings.Contains(err.Error(), "not connected") {
        // Reconnect
        client.Connect(ctx)
    } else if strings.Contains(err.Error(), "API error") {
        // Handle Milvus API error
    }
}
```

## Performance Tips

### Index Tuning

```go
// For high recall (>95%)
req := &milvus.CreateIndexRequest{
    IndexType: milvus.IndexTypeHNSW,
    Params: map[string]interface{}{
        "M":              32,
        "efConstruction": 512,
    },
}

// For fast search with good recall
req := &milvus.SearchRequest{
    SearchParams: map[string]interface{}{
        "ef": 128,  // Higher = better recall, slower
    },
}
```

### Batch Operations

```go
// Batch insert (recommended: 1000-10000 records per batch)
batchSize := 5000
for i := 0; i < len(allData); i += batchSize {
    end := min(i+batchSize, len(allData))
    batch := allData[i:end]
    client.Insert(ctx, "documents", batch)
}
```

### Partition Strategy

```go
// Use partitions for better query performance
schema := milvus.CollectionSchema{
    Fields: []milvus.FieldSchema{
        {
            FieldName:   "category",
            DataType:    milvus.DataTypeVarChar,
            IsPartition: true,  // Partition by category
        },
    },
}
```

## Testing

```bash
# Unit tests
go test -v ./internal/vectordb/milvus/...

# Integration tests (requires running Milvus)
MILVUS_HOST=localhost MILVUS_PORT=19530 \
go test -v ./internal/vectordb/milvus/... -run Integration

# Comprehensive tests
go test -v ./internal/vectordb/milvus/... -run Comprehensive
```

## Related Files

- `client.go` - Main client implementation
- `client_test.go` - Unit tests
- `client_comprehensive_test.go` - Comprehensive tests
