# pgvector - PostgreSQL Vector Extension Client

This package provides a production-ready Go client for [pgvector](https://github.com/pgvector/pgvector), the open-source vector similarity search extension for PostgreSQL. It enables efficient vector storage and similarity search within your existing PostgreSQL infrastructure.

## Overview

The pgvector client integrates with HelixAgent's vector storage layer to provide:

- **PostgreSQL Native**: Use vectors alongside relational data in a single database
- **SQL-Based Operations**: Full SQL support for vector operations
- **Multiple Index Types**: IVFFlat and HNSW indexes for different use cases
- **Connection Pooling**: Efficient connection management with pgxpool
- **Transactional Consistency**: ACID compliance for vector operations

## Prerequisites

### PostgreSQL Setup

1. Install PostgreSQL 11+ (PostgreSQL 15+ recommended)
2. Install pgvector extension:

```sql
-- Create the extension (requires superuser)
CREATE EXTENSION IF NOT EXISTS vector;
```

### Docker Setup

```yaml
# docker-compose.yml
services:
  postgres:
    image: pgvector/pgvector:pg15
    ports:
      - "5432:5432"
    environment:
      POSTGRES_USER: helixagent
      POSTGRES_PASSWORD: helixagent123
      POSTGRES_DB: helixagent_db
    volumes:
      - pgdata:/var/lib/postgresql/data
```

### Installation Verification

```sql
-- Verify extension is installed
SELECT * FROM pg_extension WHERE extname = 'vector';

-- Check version
SELECT extversion FROM pg_extension WHERE extname = 'vector';
```

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    HelixAgent Application                        │
├─────────────────────────────────────────────────────────────────┤
│                    pgvector Client                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │   Config    │  │  Connection │  │    SQL Query Builder    │  │
│  │  Validator  │  │    Pool     │  │    & Vector Ops         │  │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘  │
├─────────────────────────────────────────────────────────────────┤
│                     pgx/v5 Driver                                │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
                    ┌─────────────────┐
                    │   PostgreSQL    │
                    │ + pgvector ext  │
                    └─────────────────┘
```

## Configuration

### Configuration Options

```go
type Config struct {
    Host            string        // PostgreSQL host (default: "localhost")
    Port            int           // PostgreSQL port (default: 5432)
    User            string        // Database user (default: "postgres")
    Password        string        // User password
    Database        string        // Database name (default: "postgres")
    SSLMode         string        // SSL mode (default: "disable")
    MaxConns        int32         // Max pool connections (default: 10)
    MinConns        int32         // Min pool connections (default: 2)
    MaxConnLifetime time.Duration // Max connection lifetime (default: 1h)
    MaxConnIdleTime time.Duration // Max idle time (default: 30m)
    ConnectTimeout  time.Duration // Connection timeout (default: 30s)
}
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | User password | (none) |
| `DB_NAME` | Database name | `postgres` |
| `DB_SSLMODE` | SSL mode | `disable` |

### Example Configuration

```go
config := &pgvector.Config{
    Host:            "localhost",
    Port:            5432,
    User:            "helixagent",
    Password:        os.Getenv("DB_PASSWORD"),
    Database:        "helixagent_db",
    SSLMode:         "require",
    MaxConns:        20,
    MinConns:        5,
    MaxConnLifetime: time.Hour,
    MaxConnIdleTime: 30 * time.Minute,
    ConnectTimeout:  30 * time.Second,
}
```

## SQL Vector Operations

### Vector Type

pgvector introduces the `vector` type to PostgreSQL:

```sql
-- Create a vector column
CREATE TABLE documents (
    id TEXT PRIMARY KEY,
    embedding vector(1536),  -- 1536-dimensional vector
    content TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Insert vectors
INSERT INTO documents (id, embedding, content)
VALUES ('doc-1', '[0.1, 0.2, 0.3, ...]', 'Document content');

-- Query vectors
SELECT * FROM documents WHERE id = 'doc-1';
```

### Distance Operators

| Operator | Function | Description |
|----------|----------|-------------|
| `<->` | L2 distance | Euclidean distance |
| `<#>` | Negative inner product | Inner product (negate for similarity) |
| `<=>` | Cosine distance | Cosine distance (1 - similarity) |

```sql
-- L2 distance (lower is more similar)
SELECT id, embedding <-> '[0.1, 0.2, ...]' AS distance
FROM documents ORDER BY distance LIMIT 10;

-- Cosine distance
SELECT id, embedding <=> '[0.1, 0.2, ...]' AS distance
FROM documents ORDER BY distance LIMIT 10;

-- Inner product (negate for similarity ranking)
SELECT id, embedding <#> '[0.1, 0.2, ...]' AS negative_ip
FROM documents ORDER BY negative_ip LIMIT 10;
```

## Index Types

### IVFFlat Index

Inverted file with flat storage - good balance of speed and accuracy:

```go
req := &pgvector.CreateIndexRequest{
    TableName:    "documents",
    IndexName:    "documents_embedding_ivfflat_idx",
    VectorColumn: "embedding",
    IndexType:    pgvector.IndexTypeIVFFlat,
    Metric:       pgvector.DistanceCosine,
    Lists:        100,  // Number of inverted lists
}
if err := client.CreateIndex(ctx, req); err != nil {
    log.Fatal(err)
}
```

**SQL equivalent:**
```sql
CREATE INDEX documents_embedding_ivfflat_idx
ON documents USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100);
```

**Tuning tips:**
- `lists` = rows / 1000 for up to 1M rows
- `lists` = sqrt(rows) for over 1M rows

### HNSW Index

Hierarchical Navigable Small World - better recall, more memory:

```go
req := &pgvector.CreateIndexRequest{
    TableName:      "documents",
    IndexName:      "documents_embedding_hnsw_idx",
    VectorColumn:   "embedding",
    IndexType:      pgvector.IndexTypeHNSW,
    Metric:         pgvector.DistanceCosine,
    M:              16,   // Max connections per node
    EfConstruction: 64,   // Build-time search width
}
if err := client.CreateIndex(ctx, req); err != nil {
    log.Fatal(err)
}
```

**SQL equivalent:**
```sql
CREATE INDEX documents_embedding_hnsw_idx
ON documents USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);
```

**Tuning tips:**
- `m`: Higher = better recall, more memory (default: 16)
- `ef_construction`: Higher = better index quality, slower build (default: 64)

## Client Operations

### Connect

```go
client, err := pgvector.NewClient(config, logger)
if err != nil {
    log.Fatal(err)
}

if err := client.Connect(ctx); err != nil {
    log.Fatal(err)
}
defer client.Close()

// Health check
if err := client.HealthCheck(ctx); err != nil {
    log.Printf("Database unhealthy: %v", err)
}
```

### Table Management

```go
// Create table with schema
schema := &pgvector.TableSchema{
    TableName:    "documents",
    IDColumn:     "id",
    VectorColumn: "embedding",
    Dimension:    1536,
    MetadataColumns: []pgvector.ColumnDef{
        {Name: "title", Type: "TEXT", Nullable: true},
        {Name: "category", Type: "VARCHAR(64)", Nullable: false},
        {Name: "metadata", Type: "JSONB", Nullable: true},
    },
}
if err := client.CreateTable(ctx, schema); err != nil {
    log.Fatal(err)
}

// Check if table exists
exists, err := client.TableExists(ctx, "documents")

// Drop table
if err := client.DropTable(ctx, "documents"); err != nil {
    log.Fatal(err)
}
```

### Upsert (Insert/Update)

```go
vectors := []pgvector.Vector{
    {
        ID:     "doc-1",
        Vector: []float32{0.1, 0.2, 0.3, ...},
        Metadata: map[string]interface{}{
            "title":    "Introduction to AI",
            "category": "tech",
        },
    },
    {
        ID:     "doc-2",
        Vector: []float32{0.4, 0.5, 0.6, ...},
        Metadata: map[string]interface{}{
            "title":    "Machine Learning Basics",
            "category": "tech",
        },
    },
}

req := &pgvector.UpsertRequest{
    TableName:    "documents",
    IDColumn:     "id",
    VectorColumn: "embedding",
    Vectors:      vectors,
}

count, err := client.Upsert(ctx, req)
fmt.Printf("Upserted %d vectors\n", count)
```

### Vector Search

```go
// Basic similarity search
req := &pgvector.SearchRequest{
    TableName:    "documents",
    VectorColumn: "embedding",
    IDColumn:     "id",
    QueryVector:  queryVector,
    Limit:        10,
    Metric:       pgvector.DistanceCosine,
    OutputColumns: []string{"title", "category"},
}

results, err := client.Search(ctx, req)
for _, r := range results {
    fmt.Printf("ID: %s, Distance: %f, Title: %v\n",
        r.ID, r.Distance, r.Metadata["title"])
}

// Search with SQL filter
req := &pgvector.SearchRequest{
    TableName:     "documents",
    VectorColumn:  "embedding",
    IDColumn:      "id",
    QueryVector:   queryVector,
    Limit:         10,
    Metric:        pgvector.DistanceCosine,
    Filter:        "category = 'tech' AND created_at > '2024-01-01'",
    OutputColumns: []string{"title", "category", "created_at"},
}
```

### Get by IDs

```go
ids := []string{"doc-1", "doc-2", "doc-3"}
outputCols := []string{"id", "title", "category", "embedding"}

rows, err := client.Get(ctx, "documents", "id", ids, outputCols)
for _, row := range rows {
    fmt.Printf("ID: %v, Title: %v\n", row["id"], row["title"])
}
```

### Delete

```go
// Delete by IDs
req := &pgvector.DeleteRequest{
    TableName: "documents",
    IDColumn:  "id",
    IDs:       []string{"doc-1", "doc-2"},
}
count, err := client.Delete(ctx, req)

// Delete by filter
req := &pgvector.DeleteRequest{
    TableName: "documents",
    IDColumn:  "id",
    Filter:    "category = 'outdated'",
}
count, err := client.Delete(ctx, req)
```

### Count

```go
// Count all
count, err := client.Count(ctx, "documents", "")

// Count with filter
count, err := client.Count(ctx, "documents", "category = 'tech'")
```

## Integration with Existing PostgreSQL

### Using with Existing Tables

```go
// Access the underlying pool for custom queries
pool := client.GetPool()

// Execute custom SQL
rows, err := pool.Query(ctx, `
    SELECT d.id, d.title, d.embedding <=> $1::vector AS distance
    FROM documents d
    JOIN categories c ON d.category_id = c.id
    WHERE c.name = $2
    ORDER BY distance
    LIMIT $3
`, vectorToString(queryVector), "tech", 10)
```

### Transactions

```go
pool := client.GetPool()

tx, err := pool.Begin(ctx)
if err != nil {
    log.Fatal(err)
}
defer tx.Rollback(ctx)

// Insert document
_, err = tx.Exec(ctx, `
    INSERT INTO documents (id, embedding, title)
    VALUES ($1, $2::vector, $3)
`, id, vectorStr, title)

// Update related table
_, err = tx.Exec(ctx, `
    UPDATE document_stats SET count = count + 1
`)

if err := tx.Commit(ctx); err != nil {
    log.Fatal(err)
}
```

### Batch Processing

```go
pool := client.GetPool()

// Use COPY for bulk inserts
batch := &pgx.Batch{}
for _, vec := range vectors {
    batch.Queue(`
        INSERT INTO documents (id, embedding, title)
        VALUES ($1, $2::vector, $3)
        ON CONFLICT (id) DO UPDATE SET
            embedding = EXCLUDED.embedding,
            title = EXCLUDED.title,
            updated_at = NOW()
    `, vec.ID, vectorToString(vec.Vector), vec.Metadata["title"])
}

results := pool.SendBatch(ctx, batch)
defer results.Close()

for range vectors {
    if _, err := results.Exec(); err != nil {
        log.Printf("Batch error: %v", err)
    }
}
```

## Performance Optimization

### Index Selection

| Use Case | Index | Parameters |
|----------|-------|------------|
| < 100K vectors | IVFFlat | lists = 100 |
| 100K - 1M vectors | IVFFlat | lists = sqrt(n) |
| High recall needed | HNSW | m = 16, ef_construction = 64 |
| Memory constrained | IVFFlat | Lower lists value |

### Search Optimization

```sql
-- Set probes for IVFFlat (higher = better recall, slower)
SET ivfflat.probes = 10;

-- Set ef_search for HNSW (higher = better recall, slower)
SET hnsw.ef_search = 100;
```

### Connection Pool Tuning

```go
config := &pgvector.Config{
    MaxConns:        50,   // Handle concurrent requests
    MinConns:        10,   // Keep connections warm
    MaxConnLifetime: time.Hour,
    MaxConnIdleTime: 5 * time.Minute,
}
```

### Partial Indexes

```sql
-- Index only active documents
CREATE INDEX documents_active_embedding_idx
ON documents USING hnsw (embedding vector_cosine_ops)
WHERE status = 'active';
```

## Distance Metrics

| Metric | Go Constant | SQL Operator | Operator Class |
|--------|-------------|--------------|----------------|
| L2 (Euclidean) | `DistanceL2` | `<->` | `vector_l2_ops` |
| Inner Product | `DistanceIP` | `<#>` | `vector_ip_ops` |
| Cosine | `DistanceCosine` | `<=>` | `vector_cosine_ops` |

```go
// L2 distance for image features
req.Metric = pgvector.DistanceL2

// Cosine for normalized text embeddings
req.Metric = pgvector.DistanceCosine

// Inner product for unnormalized embeddings
req.Metric = pgvector.DistanceIP
```

## Error Handling

```go
count, err := client.Upsert(ctx, req)
if err != nil {
    if strings.Contains(err.Error(), "not connected") {
        // Reconnect
        client.Connect(ctx)
    } else if strings.Contains(err.Error(), "duplicate key") {
        // Handle conflict
    } else if strings.Contains(err.Error(), "dimension mismatch") {
        // Vector dimension doesn't match column definition
    }
}
```

## Testing

```bash
# Unit tests
go test -v ./internal/vectordb/pgvector/...

# Integration tests (requires PostgreSQL with pgvector)
DB_HOST=localhost DB_PORT=5432 DB_USER=postgres DB_PASSWORD=password \
go test -v ./internal/vectordb/pgvector/... -run Integration

# Comprehensive tests
go test -v ./internal/vectordb/pgvector/... -run Comprehensive
```

## SQL Reference

### Create Extension
```sql
CREATE EXTENSION IF NOT EXISTS vector;
```

### Create Table
```sql
CREATE TABLE documents (
    id TEXT PRIMARY KEY,
    embedding vector(1536),
    title TEXT,
    category VARCHAR(64) NOT NULL,
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

### Create Indexes
```sql
-- HNSW index (recommended)
CREATE INDEX documents_embedding_hnsw_idx
ON documents USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);

-- IVFFlat index (alternative)
CREATE INDEX documents_embedding_ivfflat_idx
ON documents USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100);
```

### Similarity Search
```sql
SELECT id, title, embedding <=> $1::vector AS distance
FROM documents
WHERE category = 'tech'
ORDER BY distance
LIMIT 10;
```

## Related Files

- `client.go` - Main client implementation
- `client_test.go` - Unit tests
- `client_comprehensive_test.go` - Comprehensive tests
