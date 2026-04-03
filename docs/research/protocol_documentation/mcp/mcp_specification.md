# Model Context Protocol (MCP) - Complete Specification

**Protocol:** MCP (Model Context Protocol)  
**Version:** 2024-11-05  
**Status:** Stable  
**HelixAgent Implementation:** [internal/mcp/](../../../internal/mcp/)  
**Analysis Date:** 2026-04-03  

---

## Executive Summary

MCP is an open protocol that standardizes how applications provide context to LLMs. It enables secure, bi-directional communication between AI models and external tools, resources, and prompts.

**Key Benefits:**
- Standardized tool calling interface
- Secure resource access
- Multi-transport support (stdio, SSE, HTTP)
- Vendor-agnostic

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     MCP ARCHITECTURE                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   ┌──────────────────┐         ┌──────────────────┐             │
│   │   MCP Client     │◄───────►│   MCP Server     │             │
│   │   (Host)         │  JSON   │   (Tool Provider)│             │
│   │                  │  RPC    │                  │             │
│   │  • Claude Code   │         │  • Filesystem    │             │
│   │  • Aider         │         │  • Git           │             │
│   │  • HelixAgent    │         │  • Bash          │             │
│   │  • Cline         │         │  • Search        │             │
│   └──────────────────┘         └──────────────────┘             │
│                                                                  │
│   Transport Layer:                                               │
│   • stdio (CLI tools)                                           │
│   • Server-Sent Events (SSE)                                    │
│   • HTTP POST                                                   │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Protocol Basics

### Message Format

All MCP messages use JSON-RPC 2.0:

```json
{
  "jsonrpc": "2.0",
  "id": "unique-id",
  "method": "method/name",
  "params": {}
}
```

### Base Protocol Types

```typescript
// Request
interface JSONRPCRequest {
  jsonrpc: "2.0";
  id: string | number;
  method: string;
  params?: unknown;
}

// Response
interface JSONRPCResponse {
  jsonrpc: "2.0";
  id: string | number;
  result?: unknown;
  error?: JSONRPCError;
}

// Error
interface JSONRPCError {
  code: number;
  message: string;
  data?: unknown;
}
```

### Error Codes

| Code | Meaning | Description |
|------|---------|-------------|
| -32700 | Parse error | Invalid JSON |
| -32600 | Invalid Request | Invalid JSON-RPC |
| -32601 | Method not found | Unknown method |
| -32602 | Invalid params | Invalid parameters |
| -32603 | Internal error | Server error |
| -32000 | Server error | Reserved for implementation |

---

## Core Capabilities

### 1. Tools

Tools are functions that the LLM can call to perform actions.

#### Tool Definition

```typescript
interface Tool {
  name: string;
  description?: string;
  inputSchema: JSONSchema;
}
```

#### Example Tool

```json
{
  "name": "read_file",
  "description": "Read contents of a file",
  "inputSchema": {
    "type": "object",
    "properties": {
      "path": {
        "type": "string",
        "description": "Path to file"
      }
    },
    "required": ["path"]
  }
}
```

#### Tool Call Request

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "read_file",
    "arguments": {
      "path": "/tmp/test.txt"
    }
  }
}
```

#### Tool Call Response

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "File contents here..."
      }
    ],
    "isError": false
  }
}
```

### 2. Resources

Resources provide read-only data to the LLM context.

#### Resource Definition

```typescript
interface Resource {
  uri: string;
  name: string;
  description?: string;
  mimeType?: string;
}
```

#### Resource Content

```typescript
interface ResourceContent {
  uri: string;
  mimeType?: string;
  text?: string;
  blob?: string; // base64
}
```

#### Resource Request

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "resources/read",
  "params": {
    "uri": "file:///project/README.md"
  }
}
```

### 3. Prompts

Prompts are pre-defined templates for LLM interactions.

#### Prompt Definition

```typescript
interface Prompt {
  name: string;
  description?: string;
  arguments?: PromptArgument[];
}

interface PromptArgument {
  name: string;
  description?: string;
  required?: boolean;
}
```

#### Get Prompt Request

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "prompts/get",
  "params": {
    "name": "code_review",
    "arguments": {
      "language": "go"
    }
  }
}
```

### 4. Sampling

Sampling allows the server to request LLM completions from the client.

#### Sampling Request

```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "method": "sampling/createMessage",
  "params": {
    "messages": [
      {
        "role": "user",
        "content": {
          "type": "text",
          "text": "Analyze this code"
        }
      }
    ],
    "modelPreferences": {
      "hints": ["claude-3-5-sonnet"]
    }
  }
}
```

---

## Lifecycle

### 1. Initialization

```json
// Client → Server: Initialize
{
  "jsonrpc": "2.0",
  "id": 0,
  "method": "initialize",
  "params": {
    "protocolVersion": "2024-11-05",
    "capabilities": {
      "tools": {},
      "resources": {},
      "prompts": {}
    },
    "clientInfo": {
      "name": "helixagent-mcp-client",
      "version": "1.0.0"
    }
  }
}

// Server → Client: Initialize Response
{
  "jsonrpc": "2.0",
  "id": 0,
  "result": {
    "protocolVersion": "2024-11-05",
    "capabilities": {
      "tools": {
        "listChanged": true
      },
      "resources": {
        "subscribe": true,
        "listChanged": true
      },
      "prompts": {
        "listChanged": true
      }
    },
    "serverInfo": {
      "name": "filesystem-server",
      "version": "1.0.0"
    }
  }
}

// Client → Server: Initialized notification
{
  "jsonrpc": "2.0",
  "method": "notifications/initialized"
}
```

### 2. Capability Negotiation

Capabilities are negotiated during initialization:

```typescript
interface ClientCapabilities {
  experimental?: object;
  roots?: {
    listChanged?: boolean;
  };
  sampling?: object;
}

interface ServerCapabilities {
  experimental?: object;
  logging?: object;
  prompts?: {
    listChanged?: boolean;
  };
  resources?: {
    subscribe?: boolean;
    listChanged?: boolean;
  };
  tools?: {
    listChanged?: boolean;
  };
}
```

---

## Transport Methods

### 1. Standard I/O (stdio)

For CLI tools and local processes:

```go
// Server process started by client
cmd := exec.Command("mcp-server-filesystem", "/allowed/path")
cmd.Stdin = clientInput
cmd.Stdout = clientOutput
```

**Characteristics:**
- One message per line
- UTF-8 encoding
- No length prefix needed

### 2. Server-Sent Events (SSE)

For HTTP-based communication:

```
GET /mcp/sse HTTP/1.1
Accept: text/event-stream

HTTP/1.1 200 OK
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive

event: endpoint
data: /mcp/messages?session_id=abc123

event: message
data: {"jsonrpc": "2.0", ...}
```

### 3. HTTP POST

Direct HTTP communication:

```
POST /mcp/messages HTTP/1.1
Content-Type: application/json

{"jsonrpc": "2.0", "id": 1, "method": "tools/list"}
```

---

## Complete Message Reference

### Client → Server

| Method | Description |
|--------|-------------|
| `initialize` | Initialize connection |
| `initialized` | Initialization complete (notification) |
| `tools/list` | List available tools |
| `tools/call` | Call a tool |
| `resources/list` | List available resources |
| `resources/read` | Read a resource |
| `resources/subscribe` | Subscribe to resource changes |
| `resources/unsubscribe` | Unsubscribe from resource |
| `prompts/list` | List available prompts |
| `prompts/get` | Get a prompt |
| `completion/complete` | Request completion |

### Server → Client

| Method | Description |
|--------|-------------|
| `notifications/resources/updated` | Resource changed |
| `notifications/resources/list_changed` | Resource list changed |
| `notifications/tools/list_changed` | Tool list changed |
| `notifications/prompts/list_changed` | Prompt list changed |
| `notifications/message` | Server message |
| `sampling/createMessage` | Request LLM sampling |
| `roots/list` | List root directories |

---

## HelixAgent MCP Implementation

### Architecture

**Source:** [`internal/mcp/`](../../../internal/mcp/)

```
internal/mcp/
├── server/
│   ├── server.go              # MCP server implementation
│   ├── handlers.go            # Request handlers
│   └── notifications.go       # Notification system
├── client/
│   ├── client.go              # MCP client
│   └── connection.go          # Connection management
├── protocol/
│   ├── types.go               # Protocol types
│   └── messages.go            # Message definitions
└── adapters/
    ├── filesystem.go          # Filesystem tool adapter
    ├── bash.go               # Bash tool adapter
    ├── git.go                # Git tool adapter
    ├── search.go             # Search tool adapter
    └── [45+ more adapters]   # Agent-specific adapters
```

### Server Implementation

**Source:** [`internal/mcp/server/server.go`](../../../internal/mcp/server/server.go)

```go
package server

// MCPServer implements the MCP protocol
// Source: internal/mcp/server/server.go#L1-200

type MCPServer struct {
    tools      map[string]Tool
    resources  map[string]Resource
    prompts    map[string]Prompt
    handlers   map[string]Handler
    transport  Transport
}

// HandleRequest processes incoming MCP requests
// Source: internal/mcp/server/server.go#L78-156
func (s *MCPServer) HandleRequest(ctx context.Context, req *JSONRPCRequest) (*JSONRPCResponse, error) {
    handler, ok := s.handlers[req.Method]
    if !ok {
        return &JSONRPCResponse{
            ID: req.ID,
            Error: &JSONRPCError{
                Code:    -32601,
                Message: "Method not found",
            },
        }, nil
    }
    
    result, err := handler(ctx, req.Params)
    if err != nil {
        return &JSONRPCResponse{
            ID: req.ID,
            Error: &JSONRPCError{
                Code:    -32603,
                Message: err.Error(),
            },
        }, nil
    }
    
    return &JSONRPCResponse{
        ID:     req.ID,
        Result: result,
    }, nil
}

// RegisterTool adds a tool to the server
// Source: internal/mcp/server/server.go#L158-178
func (s *MCPServer) RegisterTool(tool Tool, handler ToolHandler) error {
    s.tools[tool.Name] = tool
    s.handlers["tools/"+tool.Name] = func(ctx context.Context, params json.RawMessage) (interface{}, error) {
        return handler(ctx, params)
    }
    return nil
}
```

### Tool Adapter Example

**Source:** [`internal/mcp/adapters/filesystem.go`](../../../internal/mcp/adapters/filesystem.go)

```go
package adapters

// FilesystemAdapter implements file operations via MCP
// Source: internal/mcp/adapters/filesystem.go#L1-145

type FilesystemAdapter struct {
    rootDir string
    allowedPaths []string
}

// ReadFile handles tools/read_file
// Source: internal/mcp/adapters/filesystem.go#L45-78
func (f *FilesystemAdapter) ReadFile(ctx context.Context, params ReadFileParams) (*ToolResult, error) {
    // Validate path is within allowed directories
    if !f.isPathAllowed(params.Path) {
        return nil, fmt.Errorf("path not allowed: %s", params.Path)
    }
    
    // Read file
    content, err := os.ReadFile(params.Path)
    if err != nil {
        return &ToolResult{
            IsError: true,
            Content: []Content{{Type: "text", Text: err.Error()}},
        }, nil
    }
    
    return &ToolResult{
        Content: []Content{{Type: "text", Text: string(content)}},
    }, nil
}
```

---

## MCP vs CLI Agent Tool Systems

### Comparison Matrix

| Agent | Native Tool System | MCP Compatible | HelixAgent Adapter |
|-------|-------------------|----------------|-------------------|
| **Claude Code** | 7 built-in tools | ❌ | ✅ Via adapter |
| **Aider** | Git + file ops | ❌ | ✅ Via adapter |
| **Codex** | Code interpreter | ❌ | ✅ Via adapter |
| **Cline** | VS Code API | ⚠️ Partial | ⚠️ WIP |
| **OpenHands** | Custom tools | ⚠️ Partial | ✅ Via adapter |
| **Continue** | LSP + custom | ✅ | ✅ Native |

### Tool Mapping

| Claude Code Tool | MCP Equivalent | HelixAgent Source |
|-----------------|----------------|-------------------|
| `read_file` | `tools/read_file` | [`filesystem.go:45`](../../../internal/mcp/adapters/filesystem.go#L45) |
| `write_file` | `tools/write_file` | [`filesystem.go:78`](../../../internal/mcp/adapters/filesystem.go#L78) |
| `bash` | `tools/execute_command` | [`bash.go:52`](../../../internal/mcp/adapters/bash.go#L52) |
| `glob` | `tools/search_files` | [`search.go:60`](../../../internal/mcp/adapters/search.go#L60) |
| `grep` | `tools/search_content` | [`search.go:95`](../../../internal/mcp/adapters/search.go#L95) |
| `ls` | `tools/list_directory` | [`filesystem.go:110`](../../../internal/mcp/adapters/filesystem.go#L110) |
| `view` | `tools/read_file` | [`filesystem.go:45`](../../../internal/mcp/adapters/filesystem.go#L45) |

---

## Integration Examples

### Using MCP with HelixAgent

```go
package main

import (
    "context"
    "github.com/helixagent/mcp"
)

func main() {
    // Create MCP server
    server := mcp.NewServer()
    
    // Register filesystem tools
    fs := adapters.NewFilesystemAdapter("/project")
    server.RegisterTool(mcp.Tool{
        Name:        "read_file",
        Description: "Read file contents",
        InputSchema: mcp.MustSchema(ReadFileParams{}),
    }, fs.ReadFile)
    
    // Start stdio transport
    transport := mcp.NewStdioTransport()
    server.Serve(transport)
}
```

### Calling MCP Tools from LLM

```go
// Source: internal/handlers/tools.go

func CallTool(ctx context.Context, mcpClient *mcp.Client, toolName string, args map[string]interface{}) (*ToolResult, error) {
    req := &mcp.JSONRPCRequest{
        JSONRPC: "2.0",
        ID:      generateID(),
        Method:  "tools/call",
        Params: mcp.ToolCallParams{
            Name:      toolName,
            Arguments: args,
        },
    }
    
    resp, err := mcpClient.SendRequest(ctx, req)
    if err != nil {
        return nil, err
    }
    
    return resp.Result.(*mcp.ToolResult), nil
}
```

---

## Source Code Reference

### Core MCP Files

| Component | Source File | Lines | Description |
|-----------|-------------|-------|-------------|
| Server | `internal/mcp/server/server.go` | 200 | MCP server core |
| Client | `internal/mcp/client/client.go` | 180 | MCP client |
| Protocol | `internal/mcp/protocol/types.go` | 150 | Type definitions |
| Handlers | `internal/mcp/server/handlers.go` | 220 | Request handlers |
| Filesystem | `internal/mcp/adapters/filesystem.go` | 145 | File operations |
| Bash | `internal/mcp/adapters/bash.go` | 98 | Shell execution |
| Git | `internal/mcp/adapters/git.go` | 156 | Git operations |
| Search | `internal/mcp/adapters/search.go` | 112 | Search tools |
| Tests | `internal/mcp/server/server_test.go` | 340 | Unit tests |

---

## Protocol Extensions

### HelixAgent Extensions

HelixAgent extends MCP with:

1. **Ensemble Tool Calling** - Call tools across multiple models
2. **Debate Coordination** - Multi-agent tool use
3. **Audit Logging** - Complete tool call history
4. **Rate Limiting** - Per-tool rate limits
5. **Caching** - Tool result caching

---

## Conclusion

MCP is the **standard protocol for LLM tool calling**. HelixAgent provides:

- ✅ Full MCP 2024-11-05 specification implementation
- ✅ 45+ tool adapters for various use cases
- ✅ Integration with all major CLI agent tool systems
- ✅ Extensions for enterprise features

**Recommendation:** Use MCP as the primary protocol for tool integration in HelixAgent.

---

*Specification Version: MCP 2024-11-05*  
*Last Updated: 2026-04-03*  
*HelixAgent Commit: aa960946*
