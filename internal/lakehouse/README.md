# Lakehouse Package

The lakehouse package provides integration with data lakehouse technologies, specifically Apache Iceberg for large-scale analytics and data management.

## Overview

This package implements a client for Apache Iceberg REST Catalog, enabling HelixAgent to store, manage, and query large datasets using the open table format. It supports table operations, schema evolution, and partition management.

## Key Components

### Client

```go
type Client struct {
    config     *Config
    httpClient *http.Client
    logger     *logrus.Logger
    connected  bool
}
```

The main client for interacting with Iceberg REST Catalog.

### Configuration

```go
type Config struct {
    CatalogURI   string        `json:"catalog_uri"`
    Warehouse    string        `json:"warehouse"`
    Namespace    string        `json:"namespace"`
    Timeout      time.Duration `json:"timeout"`
    MaxRetries   int           `json:"max_retries"`
    AuthType     string        `json:"auth_type"` // "none", "bearer", "oauth2"
    AuthToken    string        `json:"auth_token,omitempty"`
}
```

## Features

- **Table Management**: Create, alter, drop tables
- **Schema Evolution**: Add, remove, rename columns
- **Partition Management**: Partition pruning and management
- **Snapshot Operations**: Time travel queries
- **Metadata Operations**: Table properties and statistics

## Subpackages

### internal/lakehouse/iceberg

Apache Iceberg REST Catalog client implementation:
- `client.go` - Main client implementation
- `config.go` - Configuration management

## Table Operations

```go
// Create a table
table, err := client.CreateTable(ctx, &CreateTableRequest{
    Name: "events",
    Schema: Schema{
        Fields: []Field{
            {Name: "id", Type: "long", Required: true},
            {Name: "timestamp", Type: "timestamp", Required: true},
            {Name: "data", Type: "string"},
        },
    },
})

// List tables
tables, err := client.ListTables(ctx, namespace)

// Get table metadata
metadata, err := client.GetTable(ctx, namespace, tableName)
```

## Usage

```go
import "dev.helix.agent/internal/lakehouse/iceberg"

config := &iceberg.Config{
    CatalogURI: "http://localhost:8181",
    Warehouse:  "s3://my-warehouse",
    Namespace:  "default",
}

client, err := iceberg.NewClient(config, logger)
if err != nil {
    log.Fatal(err)
}

err = client.Connect(ctx)
if err != nil {
    log.Fatal(err)
}
defer client.Close()
```

## Testing

```bash
go test -v ./internal/lakehouse/...
```

## Related Packages

- `internal/storage` - Object storage (MinIO/S3)
- `internal/database` - Relational database
