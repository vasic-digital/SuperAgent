# MCP-to-HelixAgent Bridge

This package provides an HTTP/SSE bridge that wraps stdio-based MCP (Model Context Protocol) servers and exposes them via HTTP endpoints. The bridge enables web clients and remote services to communicate with MCP servers using Server-Sent Events (SSE) for streaming responses.

## Overview

The MCP bridge solves the integration challenge between:
- **Stdio-based MCP servers**: Traditional MCP servers that communicate via stdin/stdout
- **HTTP clients**: Web applications, REST APIs, and remote services

Key features:
- **Protocol Translation**: Converts HTTP requests to MCP stdin and MCP stdout to SSE events
- **Connection Management**: Handles multiple concurrent SSE clients
- **Health Monitoring**: Process monitoring and health check endpoints
- **Metrics Collection**: Request counts, latency, and error tracking

## Bridge Architecture

```
┌──────────────────────────────────────────────────────────────────────────┐
│                           HTTP Clients                                    │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐                      │
│  │ Browser │  │ CLI App │  │ API Svc │  │ Worker  │                      │
│  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘                      │
│       │            │            │            │                            │
│       └────────────┴─────┬──────┴────────────┘                            │
│                          │                                                │
│                     HTTP/SSE                                              │
│                          │                                                │
├──────────────────────────┼────────────────────────────────────────────────┤
│                          ▼                                                │
│  ┌─────────────────────────────────────────────────────────────────────┐ │
│  │                      SSE Bridge Server                               │ │
│  │  ┌──────────────┐  ┌──────────────┐  ┌───────────────────────────┐  │ │
│  │  │ /sse         │  │ /message     │  │ /health                   │  │ │
│  │  │ SSE Handler  │  │ POST Handler │  │ Health Check              │  │ │
│  │  └──────┬───────┘  └──────┬───────┘  └───────────────────────────┘  │ │
│  │         │                 │                                          │ │
│  │         │    ┌────────────┴─────────────┐                           │ │
│  │         │    │     Message Router       │                           │ │
│  │         │    │  (pending requests map)  │                           │ │
│  │         │    └────────────┬─────────────┘                           │ │
│  │         │                 │                                          │ │
│  │  ┌──────┴─────────────────┴───────┐                                 │ │
│  │  │        Process Manager          │                                 │ │
│  │  │   (stdin writer, stdout reader) │                                 │ │
│  │  └──────────────┬──────────────────┘                                 │ │
│  └─────────────────┼────────────────────────────────────────────────────┘ │
│                    │                                                      │
│               stdin/stdout                                                │
│                    │                                                      │
│                    ▼                                                      │
│  ┌─────────────────────────────────────────────────────────────────────┐ │
│  │                     MCP Server Process                               │ │
│  │  npx @modelcontextprotocol/server-filesystem /path/to/allowed       │ │
│  └─────────────────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────────────────┘
```

## Protocol Translation

### HTTP to MCP

```
Client                          Bridge                      MCP Server
  │                               │                             │
  │  POST /message               │                             │
  │  {"jsonrpc":"2.0",           │                             │
  │   "id":1,                    │                             │
  │   "method":"tools/list"}     │                             │
  │ ─────────────────────────────>│                             │
  │                               │  stdin:                    │
  │                               │  {"jsonrpc":"2.0",...}\n   │
  │                               │────────────────────────────>│
  │                               │                             │
  │                               │  stdout:                   │
  │                               │  {"jsonrpc":"2.0",         │
  │                               │   "id":1,                  │
  │                               │   "result":{...}}\n        │
  │                               │<────────────────────────────│
  │  200 OK                      │                             │
  │  {"jsonrpc":"2.0",           │                             │
  │   "id":1,                    │                             │
  │   "result":{...}}            │                             │
  │<─────────────────────────────│                             │
```

### SSE Streaming

```
Client                          Bridge                      MCP Server
  │                               │                             │
  │  GET /sse                    │                             │
  │ ─────────────────────────────>│                             │
  │  event: connected            │                             │
  │  data: {"clientId":"..."}    │                             │
  │<─────────────────────────────│                             │
  │                               │                             │
  │  event: endpoint             │                             │
  │  data: http://host/message   │                             │
  │<─────────────────────────────│                             │
  │                               │                             │
  │         (POST /message from another client)                │
  │                               │                             │
  │  event: message              │  (response from MCP)        │
  │  data: {"jsonrpc":"2.0",...} │<────────────────────────────│
  │<─────────────────────────────│                             │
```

## Usage

### Basic Bridge Setup

```go
import (
    "context"
    "dev.helix.agent/internal/mcp/bridge"
)

// Create bridge configuration
config := bridge.SSEBridgeConfig{
    Command:     []string{"npx", "-y", "@modelcontextprotocol/server-filesystem", "/home"},
    Address:     ":8080",
    ReadTimeout: 30 * time.Second,
    WriteTimeout: 30 * time.Second,
}

// Create and start bridge
b, err := bridge.NewSSEBridge(config)
if err != nil {
    log.Fatal(err)
}

if err := b.Start(); err != nil {
    log.Fatal(err)
}

// Graceful shutdown
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
b.Shutdown(ctx)
```

### Environment Variable Configuration

```go
// Using environment variables
func Main() {
    config := bridge.DefaultConfig()

    if portStr := os.Getenv("PORT"); portStr != "" {
        if port, err := strconv.Atoi(portStr); err == nil {
            config.Port = port
        }
    }

    config.MCPCommand = os.Getenv("MCP_COMMAND")
    // Example: MCP_COMMAND="npx @modelcontextprotocol/server-fetch"

    b := bridge.New(config)
    b.Start(context.Background())
}
```

### Docker Deployment

```dockerfile
FROM node:20-slim

# Install MCP server
RUN npm install -g @modelcontextprotocol/server-filesystem

# Copy bridge binary
COPY mcp-bridge /usr/local/bin/

# Configure
ENV PORT=8080
ENV MCP_COMMAND="mcp-server-filesystem /data"

EXPOSE 8080
CMD ["mcp-bridge"]
```

## Extension Development

### Creating a Custom Bridge Extension

```go
import (
    "dev.helix.agent/internal/mcp/bridge"
)

// Custom initialization handler
type CustomBridge struct {
    *bridge.SSEBridge
    customConfig map[string]string
}

func NewCustomBridge(config bridge.SSEBridgeConfig) (*CustomBridge, error) {
    base, err := bridge.NewSSEBridge(config)
    if err != nil {
        return nil, err
    }

    return &CustomBridge{
        SSEBridge:    base,
        customConfig: make(map[string]string),
    }, nil
}

// Add middleware
func (c *CustomBridge) WithAuth(authToken string) *CustomBridge {
    // Add authentication middleware
    return c
}
```

### Programmatic API

```go
// Send a request and wait for response
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

resp, err := bridge.SendRequest(ctx, "tools/list", nil)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Tools: %v\n", resp.Result)

// Send notification (no response expected)
err = bridge.SendNotification("notifications/progress", map[string]interface{}{
    "progressToken": "abc123",
    "progress":      50,
    "total":         100,
})
```

### Embedding in Existing Server

```go
// Use bridge handler in existing HTTP server
bridge, _ := NewSSEBridge(config)
bridge.Start()

// Get the handler
handler := bridge.Handler()

// Mount at custom path
mux := http.NewServeMux()
mux.Handle("/mcp/", http.StripPrefix("/mcp", handler))
mux.HandleFunc("/api/other", otherHandler)

http.ListenAndServe(":8080", mux)
```

## Endpoints

### `GET /sse` - SSE Connection

Establishes a Server-Sent Events connection:

```bash
curl -N http://localhost:8080/sse
```

Events:
- `connected` - Initial connection with client ID
- `endpoint` - Message endpoint URL
- `message` - MCP response messages
- `shutdown` - Bridge shutdown notification

### `POST /message` - Send MCP Message

Send JSON-RPC messages to the MCP server:

```bash
curl -X POST http://localhost:8080/message \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}'
```

### `GET /health` - Health Check

Check bridge and MCP process health:

```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "running",
  "healthy": true,
  "processReady": true,
  "processPid": 12345,
  "uptime": "1h30m45s",
  "metrics": {
    "totalRequests": 1000,
    "successfulRequests": 995,
    "failedRequests": 5,
    "activeSSEConnections": 3,
    "bytesSent": 1048576,
    "bytesReceived": 524288
  }
}
```

## Configuration Options

```go
type SSEBridgeConfig struct {
    // Command to run the MCP server
    Command []string

    // Additional environment variables
    Environment map[string]string

    // HTTP server address (default: ":8080")
    Address string

    // Request timeouts
    ReadTimeout  time.Duration  // default: 30s
    WriteTimeout time.Duration  // default: 30s
    IdleTimeout  time.Duration  // default: 120s

    // Shutdown timeout
    ShutdownTimeout time.Duration  // default: 30s

    // Max request body size (default: 10MB)
    MaxRequestSize int64

    // SSE heartbeat interval (default: 30s)
    SSEHeartbeatInterval time.Duration

    // Logger instance
    Logger *logrus.Logger

    // Process exit callback
    OnProcessExit func(error)

    // Working directory for MCP process
    WorkingDirectory string
}
```

## Error Handling

### JSON-RPC Error Codes

| Code | Constant | Description |
|------|----------|-------------|
| -32700 | `JSONRPCParseError` | Invalid JSON |
| -32600 | `JSONRPCInvalidRequest` | Invalid request |
| -32601 | `JSONRPCMethodNotFound` | Method not found |
| -32602 | `JSONRPCInvalidParams` | Invalid parameters |
| -32603 | `JSONRPCInternalError` | Internal error |
| -32000 | `JSONRPCServerError` | Server error |
| -32001 | `JSONRPCProcessNotReady` | MCP process not ready |
| -32002 | `JSONRPCProcessClosed` | MCP process closed |
| -32003 | `JSONRPCTimeout` | Request timeout |

### Error Responses

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32001,
    "message": "MCP process not ready"
  }
}
```

## Monitoring

### Metrics Access

```go
metrics := bridge.Metrics()
fmt.Printf("Total requests: %d\n", metrics.TotalRequests)
fmt.Printf("Success rate: %.2f%%\n",
    float64(metrics.SuccessfulRequests)/float64(metrics.TotalRequests)*100)
fmt.Printf("Active SSE clients: %d\n", metrics.ActiveSSEConnections)
```

### Health Checks

```go
if bridge.IsHealthy() {
    // Bridge is running and MCP process is ready
}

state := bridge.State()
// StateIdle, StateStarting, StateRunning, StateStopping, StateStopped, StateError
```

## Testing

```bash
# Unit tests
go test -v ./internal/mcp/bridge/...

# Integration tests
go test -v ./internal/mcp/bridge/... -run Integration

# Comprehensive tests
go test -v ./internal/mcp/bridge/... -run Comprehensive

# SSE-specific tests
go test -v ./internal/mcp/bridge/... -run SSE
```

## Related Files

- `bridge.go` - Basic bridge implementation
- `sse_bridge.go` - Full SSE bridge with metrics
- `main.go` - Standalone entry point
- `bridge_test.go` - Unit tests
- `sse_bridge_test.go` - SSE-specific tests
- `sse_bridge_comprehensive_test.go` - Comprehensive tests
