# Protocol Discovery Service

**Location:** `docker/protocol-discovery/`  
**Language:** Go  
**Purpose:** Central registry for MCP, LSP, ACP, and Embedding servers with semantic tool search

## Overview

The Protocol Discovery Service provides a centralized registry and semantic search capability for all protocol servers (MCP, LSP, ACP, Embedding) used by HelixAgent. It implements **MCP Tool Search technology** to enable intelligent discovery and matching of tools based on natural language queries.

## Features

- **Central Registry**: Maintains a registry of all available protocol servers
- **Semantic Tool Search**: Find tools using natural language descriptions
- **Multi-Protocol Support**: MCP, LSP, ACP, and Embedding servers
- **Health Monitoring**: Tracks server health and availability
- **Dynamic Updates**: Runtime registration and deregistration of servers
- **REST API**: HTTP interface for queries and management

## Architecture

### Core Components

```
┌─────────────────────────────────────────────┐
│     Protocol Discovery Service              │
├─────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────────────┐  │
│  │   Registry  │  │   Search Engine     │  │
│  │  (in-mem)   │  │  (semantic + BM25)  │  │
│  └─────────────┘  └─────────────────────┘  │
│  ┌─────────────┐  ┌─────────────────────┐  │
│  │Health Monitor│  │  HTTP API Server   │  │
│  │             │  │    (port 8765)     │  │
│  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────┘
```

### Protocol Support

| Protocol | Description | Port Range |
|----------|-------------|------------|
| MCP | Model Context Protocol | 9101-9999 |
| LSP | Language Server Protocol | Configurable |
| ACP | Agent Communication Protocol | Configurable |
| Embedding | Vector embedding services | Configurable |

## Usage

### Running the Service

```bash
# Build the container
docker build -t helixagent/protocol-discovery docker/protocol-discovery/

# Run the service
docker run -d \
  -p 8765:8765 \
  -e REGISTRY_REFRESH_INTERVAL=30s \
  -e MAX_SERVERS=1000 \
  helixagent/protocol-discovery
```

### Docker Compose

```yaml
version: '3.8'
services:
  protocol-discovery:
    build: docker/protocol-discovery/
    ports:
      - "8765:8765"
    environment:
      - PORT=8765
      - MAX_SERVERS=1000
      - HEALTH_CHECK_INTERVAL=30s
    networks:
      - helixagent-network
```

## API Reference

### Search for Tools

```bash
curl "http://localhost:8765/api/v1/search?q=read+file+from+filesystem" \
  -H "Content-Type: application/json"
```

Response:
```json
{
  "results": [
    {
      "tool": {
        "name": "filesystem_read_file",
        "description": "Read the contents of a file",
        "category": "filesystem"
      },
      "score": 0.95,
      "match_type": "semantic"
    }
  ],
  "total": 1,
  "query_time_ms": 12
}
```

### Register a Server

```bash
curl -X POST http://localhost:8765/api/v1/servers \
  -H "Content-Type: application/json" \
  -d '{
    "id": "mcp-filesystem-01",
    "protocol": "mcp",
    "host": "mcp-filesystem",
    "port": 9101,
    "tools": [
      {
        "name": "read_file",
        "description": "Read file contents",
        "category": "filesystem"
      }
    ]
  }'
```

### List All Servers

```bash
curl http://localhost:8765/api/v1/servers
```

### Health Check

```bash
curl http://localhost:8765/health
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8765` | HTTP server port |
| `MAX_SERVERS` | `1000` | Maximum registered servers |
| `HEALTH_CHECK_INTERVAL` | `30s` | Health check frequency |
| `SEARCH_MAX_RESULTS` | `10` | Maximum search results |
| `SEARCH_MIN_SCORE` | `0.3` | Minimum relevance score |
| `METRICS_ENABLED` | `true` | Enable Prometheus metrics |

### Search Configuration

The search engine combines multiple scoring methods:

- **Semantic Similarity** (40%): Cosine similarity of word embeddings
- **BM25** (35%): Classic text retrieval scoring
- **Exact Match** (15%): Exact word matches
- **Alias Match** (10%): Tool alias matching

## Tool Schema

Tools are defined using a schema compatible with HelixAgent's internal tool system:

```go
type ToolSchema struct {
    Name           string           // Unique tool identifier
    Description    string           // Natural language description
    RequiredFields []string         // Required parameter fields
    OptionalFields []string         // Optional parameter fields
    Aliases        []string         // Alternative names
    Category       string           // Tool category (e.g., "filesystem")
    Parameters     map[string]Param // Parameter definitions
}
```

## Integration

### With HelixAgent

HelixAgent queries the Protocol Discovery Service to:

1. **Discover available tools** for a given task
2. **Route tool calls** to appropriate servers
3. **Monitor server health** and failover
4. **Cache tool schemas** for performance

### Example: Tool Discovery Flow

```go
// HelixAgent queries for appropriate tool
resp, _ := http.Get("http://protocol-discovery:8765/api/v1/search?q=read+file")

// Parse results
var results SearchResponse
json.NewDecoder(resp.Body).Decode(&results)

// Select best matching tool
bestTool := results.Results[0].Tool

// Execute tool call via appropriate MCP server
result, _ := callMCPTool(bestTool.Name, params)
```

## Development

### Building

```bash
cd docker/protocol-discovery/
go build -o protocol-discovery .
```

### Testing

```bash
go test ./...
```

### Running Locally

```bash
go run main.go
```

The service will start on port 8765 with default configuration.

## Monitoring

### Metrics Endpoint

Prometheus metrics available at `/metrics`:

- `protocol_discovery_servers_total` - Number of registered servers
- `protocol_discovery_search_requests_total` - Search request count
- `protocol_discovery_search_latency_seconds` - Search latency
- `protocol_discovery_health_check_failures_total` - Health check failures

### Health Endpoint

Health status at `/health`:

```json
{
  "status": "healthy",
  "servers": {
    "total": 15,
    "healthy": 14,
    "unhealthy": 1
  },
  "uptime_seconds": 3600
}
```

## Troubleshooting

### High Memory Usage

If memory usage is high:
- Reduce `MAX_SERVERS` limit
- Decrease search result cache size
- Enable server pruning for inactive servers

### Slow Search Performance

If searches are slow:
- Increase `SEARCH_MIN_SCORE` to filter early
- Reduce `SEARCH_MAX_RESULTS`
- Check CPU usage (search is CPU-intensive)

### Servers Not Appearing

If registered servers don't appear:
- Verify server registration HTTP POST succeeded
- Check server health check endpoint
- Ensure server ID is unique

## Security Considerations

- **No Authentication**: Service is designed for internal use
- **Network Isolation**: Should run on internal Docker network
- **Input Validation**: All inputs are validated and sanitized
- **Rate Limiting**: Consider adding rate limiting for production

## Contributing

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for guidelines.

## License

MIT License
