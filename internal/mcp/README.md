# MCP (Model Context Protocol) Package

The MCP package provides comprehensive Model Context Protocol server management, connection pooling, and adapter interfaces for HelixAgent.

## Overview

This package implements:

- **Connection Pool**: Manages connections to MCP servers with automatic reconnection
- **Server Registry**: Tracks and manages available MCP servers
- **Pre-installer**: Handles npm package installation for MCP servers
- **Adapters**: Provides interfaces for 45+ MCP server integrations

## Directory Structure

```
internal/mcp/
├── adapters/           # MCP server adapters
│   ├── registry.go     # Adapter registry and interfaces
│   ├── brave_search.go # Brave Search integration
│   ├── figma.go        # Figma design tool
│   ├── miro.go         # Miro collaboration
│   └── ...             # 40+ additional adapters
├── servers/            # Server-specific implementations
│   ├── chroma_adapter.go
│   ├── qdrant_adapter.go
│   ├── postgres_adapter.go
│   └── ...
├── connection_pool.go  # Connection pool management
├── server_registry.go  # Server configuration and discovery
├── preinstaller.go     # NPM package pre-installation
├── extended_packages.go # Extended MCP package definitions
└── types.go            # Shared types
```

## Key Components

### Connection Pool

Manages persistent connections to MCP servers:

```go
pool := mcp.NewConnectionPool(config, logger)

// Get a connection
conn, err := pool.GetConnection(ctx, "server-name")
if err != nil {
    return err
}
defer pool.ReleaseConnection(conn)

// Use the connection
result, err := conn.CallTool(ctx, "tool_name", args)
```

### Server Registry

Manages MCP server configurations:

```go
registry := mcp.NewServerRegistry(logger)

// Register a server
registry.RegisterServer(mcp.ServerConfig{
    Name: "my-server",
    Command: "npx",
    Args: []string{"@modelcontextprotocol/server-filesystem"},
})

// Get server info
server, ok := registry.GetServer("my-server")
```

### Adapters

45+ MCP server adapters organized by category:

| Category | Adapters |
|----------|----------|
| **Database** | PostgreSQL, SQLite, MongoDB, Redis |
| **Storage** | AWS S3, Google Drive |
| **Version Control** | GitHub, GitLab |
| **Search** | Brave Search, Exa |
| **Design** | Figma, Miro, SVGMaker |
| **AI/ML** | Stable Diffusion, Replicate |
| **Infrastructure** | Docker, Kubernetes |
| **Analytics** | Sentry, Datadog |

### Using Adapters

```go
// Create adapter
config := adapters.BraveSearchConfig{
    APIKey: os.Getenv("BRAVE_API_KEY"),
}
adapter := adapters.NewBraveSearchAdapter(config)

// List available tools
tools := adapter.ListTools()

// Call a tool
result, err := adapter.CallTool(ctx, "brave_web_search", map[string]interface{}{
    "query": "golang programming",
    "count": 10,
})
```

## MCP Package Categories

| Category | Description |
|----------|-------------|
| `CategoryCore` | Official Anthropic servers |
| `CategoryDatabase` | Database operations |
| `CategoryStorage` | File and cloud storage |
| `CategoryVersionControl` | Git providers |
| `CategorySearch` | Web search engines |
| `CategoryDesign` | Design tools |
| `CategoryAI` | AI/ML services |
| `CategoryInfrastructure` | DevOps tools |

## Configuration

Environment variables:

```bash
# Server configuration
MCP_INSTALL_DIR=/path/to/mcp-servers
MCP_CONCURRENT_INSTALLS=4
MCP_INSTALL_TIMEOUT=300s

# Adapter API keys
BRAVE_API_KEY=your-key
FIGMA_ACCESS_TOKEN=your-token
GITHUB_TOKEN=your-token
```

## Testing

```bash
# Run all MCP tests
go test -v ./internal/mcp/...

# Run adapter tests
go test -v ./internal/mcp/adapters/...

# Run server tests
go test -v ./internal/mcp/servers/...
```

## Adding New Adapters

1. Create adapter file in `adapters/`:
```go
type MyAdapter struct {
    config MyConfig
    client *http.Client
}

func (a *MyAdapter) GetServerInfo() ServerInfo { ... }
func (a *MyAdapter) ListTools() []ToolDefinition { ... }
func (a *MyAdapter) CallTool(ctx context.Context, name string, args map[string]interface{}) (*ToolResult, error) { ... }
```

2. Add to registry:
```go
var AvailableAdapters = []AdapterMetadata{
    // ...
    {Name: "my-adapter", Category: CategoryUtility, ...},
}
```

3. Create tests in `adapters/my_adapter_test.go`

## Dependencies

- `github.com/sirupsen/logrus` - Logging
- `net/http` - HTTP client for API calls
- Standard library for process management

## See Also

- [Model Context Protocol Specification](https://modelcontextprotocol.io/)
- [MCP Server Examples](https://github.com/modelcontextprotocol/servers)
