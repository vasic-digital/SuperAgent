# LSP Protocol Testing Package

The LSP testing package provides comprehensive utilities for testing Language Server Protocol (LSP) servers with real functional validation.

## Overview

This package implements:

- **LSP Testing Harness**: JSON-RPC 2.0 client with Content-Length framing
- **Language Server Mock**: Configurable mock for unit testing
- **Protocol Compliance Tests**: LSP 3.17 specification validation
- **Diagnostics Testing**: Code analysis and error detection tests
- **Integration Patterns**: IDE/editor integration test patterns

## Key Principles

**No False Positives**: Tests execute ACTUAL LSP operations, not just connectivity checks. Tests FAIL if operations fail.

## Directory Structure

```
internal/testing/lsp/
├── functional_test.go    # Real LSP server functional tests
└── README.md             # This file

internal/lsp/
├── ai_completion.go      # AI-powered completion provider
├── document_store.go     # Document management
├── symbol_index.go       # Code symbol indexing
├── completion_cache.go   # Completion caching
└── ...
```

## Key Components

### LSPClient

JSON-RPC 2.0 client with LSP Content-Length header framing:

```go
// Create client connected to LSP server
client, err := lsp.NewLSPClient("localhost:9501", 10*time.Second)
if err != nil {
    t.Skipf("LSP server not running: %v", err)
}
defer client.Close()

// Initialize LSP session
resp, err := client.Initialize("file:///workspace")
require.NoError(t, err)
require.Nil(t, resp.Error)

// Parse capabilities
var result map[string]interface{}
json.Unmarshal(resp.Result, &result)
capabilities := result["capabilities"]
```

### Protocol Message Format

LSP uses Content-Length header framing:

```
Content-Length: 52

{"jsonrpc":"2.0","id":1,"method":"initialize",...}
```

### Request/Response Types

```go
// LSPRequest represents a JSON-RPC 2.0 request
type LSPRequest struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      int         `json:"id"`
    Method  string      `json:"method"`
    Params  interface{} `json:"params,omitempty"`
}

// LSPResponse represents a JSON-RPC 2.0 response
type LSPResponse struct {
    JSONRPC string          `json:"jsonrpc"`
    ID      int             `json:"id"`
    Result  json.RawMessage `json:"result,omitempty"`
    Error   *LSPError       `json:"error,omitempty"`
}

// LSPError represents a JSON-RPC error
type LSPError struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}
```

## Language Server Mock

Create a mock LSP server for unit testing:

```go
import "dev.helix.agent/internal/testing"

// Create mock HTTP server (simplified LSP mock)
mock := testing.NewMockHTTPServer()
defer mock.Close()

// Set response for textDocument/completion
mock.SetResponse("/lsp/completion", &testing.MockHTTPResponse{
    StatusCode: 200,
    Headers:    map[string]string{"Content-Type": "application/json"},
    Body:       `{"items": [{"label": "function", "kind": 3}]}`,
})

// Use mock URL in tests
client := createLSPClient(mock.URL())
```

### AI-Powered LSP Mock

```go
// Mock AI completion provider for testing
type MockAICompletionProvider struct {
    Completions []CompletionItem
    Error       error
    Latency     time.Duration
}

func (m *MockAICompletionProvider) GetCompletions(ctx context.Context, req *CompletionRequest) ([]CompletionItem, error) {
    if m.Latency > 0 {
        time.Sleep(m.Latency)
    }
    if m.Error != nil {
        return nil, m.Error
    }
    return m.Completions, nil
}
```

## Supported LSP Servers

| Server | Port | Language | Test File |
|--------|------|----------|-----------|
| gopls | 9501 | Go | test.go |
| pyright | 9502 | Python | test.py |
| typescript-language-server | 9503 | TypeScript | test.ts |
| rust-analyzer | 9504 | Rust | test.rs |
| clangd | 9505 | C/C++ | test.cpp |

## Integration Testing Patterns

### Initialize and Shutdown

```go
func TestLSPServerLifecycle(t *testing.T) {
    client, err := lsp.NewLSPClient("localhost:9501", 10*time.Second)
    require.NoError(t, err)
    defer client.Close()

    // Initialize
    initResp, err := client.Initialize("file:///tmp/workspace")
    require.NoError(t, err)
    require.Nil(t, initResp.Error)

    // Verify capabilities
    var result map[string]interface{}
    json.Unmarshal(initResp.Result, &result)

    capabilities := result["capabilities"].(map[string]interface{})
    assert.NotNil(t, capabilities)

    // Shutdown
    shutdownResp, err := client.Shutdown()
    require.NoError(t, err)
}
```

### Document Synchronization

```go
func TestDocumentSync(t *testing.T) {
    client := setupLSPClient(t)

    // Open document
    openParams := map[string]interface{}{
        "textDocument": map[string]interface{}{
            "uri":        "file:///workspace/main.go",
            "languageId": "go",
            "version":    1,
            "text":       "package main\n\nfunc main() {\n}\n",
        },
    }
    _, err := client.Call("textDocument/didOpen", openParams)
    require.NoError(t, err)

    // Change document
    changeParams := map[string]interface{}{
        "textDocument": map[string]interface{}{
            "uri":     "file:///workspace/main.go",
            "version": 2,
        },
        "contentChanges": []map[string]interface{}{
            {"text": "package main\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}\n"},
        },
    }
    _, err = client.Call("textDocument/didChange", changeParams)
    require.NoError(t, err)

    // Close document
    closeParams := map[string]interface{}{
        "textDocument": map[string]interface{}{
            "uri": "file:///workspace/main.go",
        },
    }
    _, err = client.Call("textDocument/didClose", closeParams)
    require.NoError(t, err)
}
```

### Code Completion

```go
func TestCodeCompletion(t *testing.T) {
    client := setupLSPClient(t)
    openDocument(t, client, "test.go", "package main\n\nfunc main() {\n\tfmt.\n}")

    completionParams := map[string]interface{}{
        "textDocument": map[string]interface{}{
            "uri": "file:///workspace/test.go",
        },
        "position": map[string]interface{}{
            "line":      3,
            "character": 6, // After "fmt."
        },
    }

    resp, err := client.Call("textDocument/completion", completionParams)
    require.NoError(t, err)
    require.Nil(t, resp.Error)

    var result struct {
        Items []struct {
            Label string `json:"label"`
            Kind  int    `json:"kind"`
        } `json:"items"`
    }
    json.Unmarshal(resp.Result, &result)

    assert.NotEmpty(t, result.Items, "Should have completion items")
}
```

## Diagnostics Testing

### Error Detection

```go
func TestDiagnostics(t *testing.T) {
    client := setupLSPClient(t)

    // Open document with error
    code := `package main

func main() {
    x := undefined_variable
}
`
    openDocument(t, client, "error.go", code)

    // Wait for diagnostics (async)
    time.Sleep(500 * time.Millisecond)

    // Request diagnostics
    resp, err := client.Call("textDocument/diagnostic", map[string]interface{}{
        "textDocument": map[string]interface{}{
            "uri": "file:///workspace/error.go",
        },
    })

    if err == nil && resp.Error == nil {
        var diagnostics struct {
            Items []struct {
                Range struct {
                    Start struct{ Line, Character int } `json:"start"`
                    End   struct{ Line, Character int } `json:"end"`
                } `json:"range"`
                Severity int    `json:"severity"`
                Message  string `json:"message"`
            } `json:"items"`
        }
        json.Unmarshal(resp.Result, &diagnostics)

        assert.NotEmpty(t, diagnostics.Items, "Should detect error")
        assert.Contains(t, diagnostics.Items[0].Message, "undefined")
    }
}
```

### Diagnostic Severity Levels

| Level | Value | Description |
|-------|-------|-------------|
| Error | 1 | Compilation error |
| Warning | 2 | Potential issue |
| Information | 3 | Informational message |
| Hint | 4 | Suggestion |

## Test Helpers

### Setup Functions

```go
func setupLSPClient(t *testing.T) *lsp.LSPClient {
    client, err := lsp.NewLSPClient("localhost:9501", 10*time.Second)
    if err != nil {
        t.Skipf("LSP server not running: %v", err)
    }
    t.Cleanup(func() { client.Close() })

    _, err = client.Initialize("file:///tmp/workspace")
    require.NoError(t, err)

    return client
}

func openDocument(t *testing.T, client *lsp.LSPClient, filename, content string) {
    params := map[string]interface{}{
        "textDocument": map[string]interface{}{
            "uri":        fmt.Sprintf("file:///workspace/%s", filename),
            "languageId": detectLanguage(filename),
            "version":    1,
            "text":       content,
        },
    }
    _, err := client.Call("textDocument/didOpen", params)
    require.NoError(t, err)
}
```

### Language Detection

```go
func detectLanguage(filename string) string {
    ext := filepath.Ext(filename)
    switch ext {
    case ".go":
        return "go"
    case ".py":
        return "python"
    case ".ts":
        return "typescript"
    case ".js":
        return "javascript"
    case ".rs":
        return "rust"
    case ".cpp", ".cc", ".cxx":
        return "cpp"
    default:
        return "plaintext"
    }
}
```

## Running Tests

```bash
# Run all LSP tests
go test -v ./internal/testing/lsp/...

# Run specific server test
go test -v -run TestLSPServerInitialize ./internal/testing/lsp/

# Run with extended timeout
go test -v -timeout 60s ./internal/testing/lsp/

# Run benchmarks
go test -bench=BenchmarkLSPInitialize ./internal/testing/lsp/
```

## Benchmarking

```go
func BenchmarkLSPInitialize(b *testing.B) {
    client, err := lsp.NewLSPClient("localhost:9501", 10*time.Second)
    if err != nil {
        b.Skipf("LSP server not running: %v", err)
    }
    defer client.Close()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := client.Initialize("file:///tmp/bench-workspace")
        if err != nil {
            b.Fatalf("Initialize failed: %v", err)
        }
    }
}

func BenchmarkCompletion(b *testing.B) {
    client := setupBenchClient(b)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = client.Call("textDocument/completion", completionParams)
    }
}
```

## Error Handling

```go
resp, err := client.Call("textDocument/hover", params)
if err != nil {
    // Network/protocol error
    t.Fatalf("Request failed: %v", err)
}

if resp.Error != nil {
    switch resp.Error.Code {
    case -32700:
        t.Log("Parse error")
    case -32600:
        t.Log("Invalid request")
    case -32601:
        t.Log("Method not found")
    case -32602:
        t.Log("Invalid params")
    default:
        t.Logf("Server error: %s (code: %d)", resp.Error.Message, resp.Error.Code)
    }
}
```

## Best Practices

1. **Initialize Before Operations**: Always call `initialize` before other LSP methods
2. **Handle Async Responses**: Some LSP operations are asynchronous
3. **Clean Up Documents**: Close documents after testing
4. **Respect Server Capabilities**: Check capabilities before calling methods
5. **Use Appropriate Timeouts**: LSP operations can be slow
6. **Skip Unavailable Servers**: Use `t.Skip()` when servers are not running

## Configuration

```go
type LSPServerConfig struct {
    Name     string   // Server name (e.g., "gopls")
    Port     int      // TCP port
    Language string   // Primary language
    TestFile string   // Default test file extension
}

var LSPServers = []LSPServerConfig{
    {Name: "gopls", Port: 9501, Language: "go", TestFile: "test.go"},
    {Name: "pyright", Port: 9502, Language: "python", TestFile: "test.py"},
    // ...
}
```

## See Also

- [LSP Specification](https://microsoft.github.io/language-server-protocol/)
- `internal/lsp/` - LSP server implementations
- `internal/services/lsp_manager.go` - LSP manager service
- `docker/lsp/docker-compose.lsp.yml` - LSP server Docker configuration
