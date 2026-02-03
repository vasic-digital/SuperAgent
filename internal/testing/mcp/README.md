# MCP Protocol Testing Package

The MCP testing package provides comprehensive utilities for testing Model Context Protocol (MCP) servers with real functional validation.

## Overview

This package implements:

- **MCP Testing Client**: JSON-RPC 2.0 client for MCP server communication
- **Mock MCP Server**: Configurable mock server for unit testing
- **Protocol Compliance Tests**: Validates MCP 2024-11-05 specification compliance
- **Functional Tool Tests**: Real tool execution tests (no false positives)
- **Test Fixtures**: Pre-built request/response templates

## Key Principles

**No False Positives**: Tests execute ACTUAL MCP tool calls via JSON-RPC. Tests FAIL if tool execution fails - connectivity checks alone are insufficient.

## Directory Structure

```
internal/testing/mcp/
├── functional_test.go    # Real MCP server functional tests
└── README.md             # This file

internal/testing/
├── helpers.go            # Mock MCP server implementation
├── framework.go          # Test framework utilities
└── ...
```

## Key Components

### MCPClient

JSON-RPC 2.0 client for testing MCP servers:

```go
// Create client with address and timeout
client, err := mcp.NewMCPClient("localhost:9103", 10*time.Second)
if err != nil {
    t.Skipf("MCP server not running: %v", err)
}
defer client.Close()

// Initialize MCP session (required before tool calls)
resp, err := client.Initialize()
require.NoError(t, err)
require.Nil(t, resp.Error)

// List available tools
resp, err = client.ListTools()
require.NoError(t, err)

// Call a specific tool
resp, err = client.CallTool("get_current_time", map[string]interface{}{
    "timezone": "UTC",
})
require.NoError(t, err)
if resp.Error != nil {
    t.Fatalf("Tool returned error: %s", resp.Error.Message)
}
```

### MCPRequest/MCPResponse Types

```go
// MCPRequest represents a JSON-RPC 2.0 request
type MCPRequest struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      int         `json:"id"`
    Method  string      `json:"method"`
    Params  interface{} `json:"params,omitempty"`
}

// MCPResponse represents a JSON-RPC 2.0 response
type MCPResponse struct {
    JSONRPC string          `json:"jsonrpc"`
    ID      int             `json:"id"`
    Result  json.RawMessage `json:"result,omitempty"`
    Error   *MCPError       `json:"error,omitempty"`
}

// MCPError represents a JSON-RPC error
type MCPError struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}
```

## Mock MCP Server

The testing package provides a configurable mock MCP server for unit testing:

```go
// Create mock server with default configuration
mock := testing.NewMockMCPServer()
defer mock.Close()

// Get server URL for client configuration
url := mock.URL()

// Configure initialization response
mock.InitializeResponse = &testing.MCPInitializeResponse{
    ProtocolVersion: "2024-11-05",
    Capabilities: testing.MCPCapabilities{
        Tools: &testing.MCPToolsCapability{ListChanged: false},
    },
    ServerInfo: testing.MCPServerInfo{
        Name:    "test-server",
        Version: "1.0.0",
    },
}

// Configure available tools
mock.DefaultToolsResponse = &testing.MCPToolsListResponse{
    Tools: []testing.MCPTool{
        {
            Name:        "custom_tool",
            Description: "A custom test tool",
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "param1": map[string]interface{}{"type": "string"},
                },
            },
        },
    },
}

// Add custom method handler
mock.AddHandler("custom/method", func(req *testing.MockMCPRequest) (*testing.MCPResponse, error) {
    result, _ := json.Marshal(map[string]string{"status": "ok"})
    return &testing.MCPResponse{
        JSONRPC: "2.0",
        ID:      req.ID,
        Result:  result,
    }, nil
})
```

### Mock Server Features

```go
// Simulate network latency
mock.Latency = 100 * time.Millisecond

// Fail after N requests (for error testing)
mock.FailAfter = 5

// Get all received requests
requests := mock.GetRequests()
fmt.Printf("Received %d requests\n", len(requests))

// Reset request log
mock.Reset()

// Get total request count
count := mock.GetRequestCount()
```

## Protocol Compliance Tests

### Initialize Handshake

```go
func TestMCPInitialize(t *testing.T) {
    client, err := mcp.NewMCPClient("localhost:9103", 10*time.Second)
    require.NoError(t, err)
    defer client.Close()

    resp, err := client.Initialize()
    require.NoError(t, err)
    require.Nil(t, resp.Error)

    var result map[string]interface{}
    err = json.Unmarshal(resp.Result, &result)
    require.NoError(t, err)

    // Verify protocol version
    assert.Equal(t, "2024-11-05", result["protocolVersion"])

    // Verify capabilities
    capabilities, ok := result["capabilities"]
    assert.True(t, ok, "Must have capabilities")

    // Verify server info
    serverInfo, ok := result["serverInfo"]
    assert.True(t, ok, "Must have serverInfo")
}
```

### Tools Discovery

```go
func TestMCPToolsDiscovery(t *testing.T) {
    client := setupMCPClient(t)

    resp, err := client.ListTools()
    require.NoError(t, err)
    require.Nil(t, resp.Error)

    var result struct {
        Tools []struct {
            Name        string                 `json:"name"`
            Description string                 `json:"description"`
            InputSchema map[string]interface{} `json:"inputSchema"`
        } `json:"tools"`
    }
    err = json.Unmarshal(resp.Result, &result)
    require.NoError(t, err)

    assert.NotEmpty(t, result.Tools, "Must have at least one tool")

    // Verify tool schema
    for _, tool := range result.Tools {
        assert.NotEmpty(t, tool.Name)
        if tool.InputSchema != nil {
            assert.Equal(t, "object", tool.InputSchema["type"])
        }
    }
}
```

### Tool Execution

```go
func TestMCPToolExecution(t *testing.T) {
    client := setupMCPClient(t)

    // Call tool with arguments
    resp, err := client.CallTool("get_current_time", map[string]interface{}{
        "timezone": "UTC",
    })
    require.NoError(t, err)

    // Verify no protocol error
    if resp.Error != nil {
        t.Fatalf("Tool returned error: %s (code: %d)", resp.Error.Message, resp.Error.Code)
    }

    // Parse and validate result
    var result map[string]interface{}
    err = json.Unmarshal(resp.Result, &result)
    require.NoError(t, err)

    // Verify content structure
    content, ok := result["content"]
    assert.True(t, ok, "Response must contain content")
    assert.NotEmpty(t, content)
}
```

## Test Fixtures

### Sample Requests

```go
fixtures := testing.NewTestFixtures(t)

// MCP initialize request
initReq := fixtures.SampleMCPInitializeRequest()

// Tools call request
callReq := fixtures.SampleMCPToolsCallRequest("echo", map[string]interface{}{
    "message": "Hello, MCP!",
}, 1)

// Custom JSON-RPC request
customReq := fixtures.SampleJSONRPCRequest("custom/method", 2, map[string]interface{}{
    "key": "value",
})
```

### Temporary Files for Testing

```go
fixtures := testing.NewTestFixtures(t)

// Create temp file
path := fixtures.CreateTempFile("test.json", `{"key": "value"}`)

// Create executable script
script := fixtures.CreateExecutableFile("mock-mcp.sh", mockScript)

// Create subdirectory
dir := fixtures.CreateSubDir("workspace")
```

## Core MCP Servers

The following MCP servers are available for testing:

| Server | Port | Primary Tools |
|--------|------|---------------|
| fetch | 9101 | `fetch` |
| git | 9102 | `git_status`, `git_log`, `git_diff`, `git_branch_list` |
| time | 9103 | `get_current_time` |
| filesystem | 9104 | `read_file`, `write_file`, `list_directory`, `create_directory` |
| memory | 9105 | `create_entities`, `read_graph`, `search_nodes`, `add_observations` |
| everything | 9106 | `search`, `everything_search` |
| sequentialthinking | 9107 | `think`, `create_thinking_session`, `continue_thinking` |

## Running Tests

```bash
# Run all MCP tests
go test -v ./internal/testing/mcp/...

# Run specific server test
go test -v -run TestMCPTimeServerFunctional ./internal/testing/mcp/

# Run with timeout (for slow servers)
go test -v -timeout 60s ./internal/testing/mcp/

# Run benchmarks
go test -bench=BenchmarkMCPToolCall ./internal/testing/mcp/
```

## Benchmarking

```go
func BenchmarkMCPToolCall(b *testing.B) {
    client, err := mcp.NewMCPClient("localhost:9103", 10*time.Second)
    if err != nil {
        b.Skipf("MCP server not running: %v", err)
    }
    defer client.Close()
    _, _ = client.Initialize()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := client.CallTool("get_current_time", map[string]interface{}{
            "timezone": "UTC",
        })
        if err != nil {
            b.Fatalf("Tool call failed: %v", err)
        }
    }
}
```

## Error Handling

### Protocol Errors

```go
resp, err := client.CallTool("unknown_tool", nil)
if err != nil {
    // Network/connection error
    t.Fatalf("Connection error: %v", err)
}

if resp.Error != nil {
    switch resp.Error.Code {
    case -32601:
        t.Log("Method not found")
    case -32602:
        t.Log("Invalid params")
    case -32603:
        t.Log("Internal error")
    default:
        t.Logf("Error: %s (code: %d)", resp.Error.Message, resp.Error.Code)
    }
}
```

### Timeout Handling

```go
client, err := mcp.NewMCPClient("localhost:9103", 5*time.Second)
// Timeout applies to both connect and read/write operations

resp, err := client.CallTool("slow_operation", args)
if err != nil && strings.Contains(err.Error(), "timeout") {
    t.Log("Operation timed out")
}
```

## Best Practices

1. **Always Initialize First**: Call `Initialize()` before any tool calls
2. **Check Both Error Paths**: Check both `err` and `resp.Error`
3. **Use Appropriate Timeouts**: Adjust timeout based on expected operation duration
4. **Skip Unavailable Servers**: Use `t.Skip()` when servers are not running
5. **Clean Up Connections**: Always `defer client.Close()`
6. **Validate Tool Results**: Don't just check for errors, validate actual results

## See Also

- [MCP Specification](https://modelcontextprotocol.io/)
- `internal/mcp/` - MCP server implementations
- `internal/testing/helpers.go` - Mock MCP server
- `docker/mcp/` - MCP server Docker configurations
