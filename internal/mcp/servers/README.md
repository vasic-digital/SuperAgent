# MCP Servers Package

This package provides MCP server implementations and adapters for various backend services.

## Overview

The MCP servers package implements standardized MCP server interfaces for database connections, vector stores, and other backend services.

## Available Servers

### Vector Stores

| Server | File | Purpose |
|--------|------|---------|
| Chroma | `chroma_adapter.go` | ChromaDB vector store |
| Qdrant | `qdrant_adapter.go` | Qdrant vector database |
| Weaviate | `weaviate_adapter.go` | Weaviate vector search |

### Databases

| Server | File | Purpose |
|--------|------|---------|
| PostgreSQL | `postgres_adapter.go` | PostgreSQL operations |
| SQLite | `sqlite_adapter.go` | SQLite database |
| Redis | `redis_adapter.go` | Redis cache/store |

### Services

| Server | File | Purpose |
|--------|------|---------|
| Fetch | `fetch_adapter.go` | HTTP fetch operations |
| Git | `git_adapter.go` | Git repository operations |
| GitHub | `github_adapter.go` | GitHub API integration |
| Memory | `memory_adapter.go` | In-memory data store |
| Replicate | `replicate_adapter.go` | Replicate ML models |
| SVG Maker | `svgmaker_adapter.go` | SVG generation |

### Image Generation

| Server | File | Purpose |
|--------|------|---------|
| Stable Diffusion | `stablediffusion_adapter.go` | Image generation |

### Unified Manager (`unified_manager.go`)

Manages multiple MCP servers with unified interface:

```go
manager := servers.NewUnifiedManager(config)
manager.Register("chroma", servers.NewChromaAdapter(chromaConfig))
manager.Register("qdrant", servers.NewQdrantAdapter(qdrantConfig))

// Use server
result, err := manager.Execute(ctx, "chroma", "add_documents", params)
```

## Architecture

```
┌─────────────────────────────────────────────┐
│              Unified Manager                 │
│  ┌─────────────────────────────────────┐   │
│  │         Server Registry             │   │
│  │  ┌──────┐ ┌──────┐ ┌──────┐        │   │
│  │  │Chroma│ │Qdrant│ │GitHub│ ...    │   │
│  │  └──────┘ └──────┘ └──────┘        │   │
│  └─────────────────────────────────────┘   │
│                    │                        │
│                    ▼                        │
│  ┌─────────────────────────────────────┐   │
│  │     MCP Protocol Implementation     │   │
│  │   initialize │ tools │ resources    │   │
│  └─────────────────────────────────────┘   │
└─────────────────────────────────────────────┘
```

## Usage

### Vector Store Example (Qdrant)

```go
import "dev.helix.agent/internal/mcp/servers"

// Create Qdrant adapter
qdrant := servers.NewQdrantAdapter(servers.QdrantConfig{
    Host:       "localhost",
    Port:       6333,
    Collection: "documents",
})

// Add documents
err := qdrant.AddDocuments(ctx, []servers.Document{
    {ID: "1", Content: "Hello world", Metadata: map[string]interface{}{"source": "test"}},
})

// Search
results, err := qdrant.Search(ctx, "hello", 10)
```

### Database Example (PostgreSQL)

```go
postgres := servers.NewPostgresAdapter(servers.PostgresConfig{
    ConnectionString: "postgres://user:pass@localhost/db",
})

// Execute query
result, err := postgres.Query(ctx, "SELECT * FROM users WHERE id = $1", userID)
```

## Configuration

```yaml
mcp:
  servers:
    qdrant:
      enabled: true
      host: "localhost"
      port: 6333
      collection: "documents"

    chroma:
      enabled: true
      host: "localhost"
      port: 8000

    github:
      enabled: true
      token: "${GITHUB_TOKEN}"
      repos:
        - "owner/repo1"
        - "owner/repo2"

    memory:
      enabled: true
      max_items: 10000
```

## Testing

```bash
go test -v ./internal/mcp/servers/...
```

## Files

- `unified_manager.go` - Unified server management
- `chroma_adapter.go` - ChromaDB adapter
- `chroma_adapter_test.go` - ChromaDB tests
- `fetch_adapter.go` - HTTP fetch adapter
- `figma_adapter.go` - Figma adapter
- `git_adapter.go` - Git operations
- `github_adapter.go` - GitHub API adapter
- `memory_adapter.go` - In-memory store
- `memory_adapter_test.go` - Memory adapter tests
- `miro_adapter.go` - Miro adapter
- `miro_adapter_test.go` - Miro tests
- `postgres_adapter.go` - PostgreSQL adapter
- `qdrant_adapter.go` - Qdrant adapter
- `qdrant_adapter_test.go` - Qdrant tests
- `redis_adapter.go` - Redis adapter
- `replicate_adapter.go` - Replicate adapter
- `sqlite_adapter.go` - SQLite adapter
- `stablediffusion_adapter.go` - Stable Diffusion
- `stablediffusion_adapter_test.go` - SD tests
- `svgmaker_adapter.go` - SVG generation
- `weaviate_adapter.go` - Weaviate adapter
