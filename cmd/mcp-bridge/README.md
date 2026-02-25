# MCP Bridge

The MCP Bridge is a standalone binary that wraps stdio-based MCP (Model Context Protocol) servers and exposes them over HTTP with Server-Sent Events (SSE) support.

## Purpose

MCP servers typically communicate over stdin/stdout, which limits their use in distributed systems. The MCP Bridge solves this by:

1. **Protocol Translation**: Converting stdio MCP to HTTP/SSE
2. **Network Access**: Making MCP servers accessible over the network
3. **Multi-Client Support**: Allowing multiple clients to connect simultaneously
4. **Connection Management**: Handling reconnection and error recovery

## Installation

```bash
# Build from source
make build

# Or using go install
go install dev.helix.agent/cmd/mcp-bridge@latest
```

## Usage

### Basic Usage

```bash
# Start the bridge with default settings
./mcp-bridge

# With custom port
./mcp-bridge --port 8080

# With specific MCP server
./mcp-bridge --mcp-server /path/to/mcp-server
```

### Configuration

| Flag | Default | Description |
|------|---------|-------------|
| `--port` | 9100 | HTTP server port |
| `--mcp-server` | - | Path to MCP server binary |
| `--log-level` | info | Logging level |
| `--timeout` | 30s | Request timeout |

## API Endpoints

### POST /mcp

Execute an MCP tool call.

**Request:**
```json
{
  "method": "tools/call",
  "params": {
    "name": "tool_name",
    "arguments": {}
  }
}
```

**Response:**
```json
{
  "result": {},
  "error": null
}
```

### GET /sse

Connect to SSE event stream.

## Architecture

```
┌─────────────┐     HTTP/SSE     ┌─────────────┐
│   Client    │ ◄──────────────► │  MCP Bridge │
└─────────────┘                  └──────┬──────┘
                                        │
                                   stdin/stdout
                                        │
                                 ┌──────▼──────┐
                                 │ MCP Server  │
                                 └─────────────┘
```

## Integration with HelixAgent

The MCP Bridge is automatically started by HelixAgent when MCP services are configured. See `internal/mcp/bridge/` for the implementation details.

## Troubleshooting

### Bridge won't start
- Check if the port is already in use
- Verify MCP server path is correct
- Check logs with `--log-level debug`

### Connection drops
- Increase timeout with `--timeout`
- Check network stability
- Verify MCP server isn't crashing
