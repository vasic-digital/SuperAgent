# Claude Code: API Documentation & HelixAgent Cross-Reference

**Agent:** Claude Code (Anthropic)  
**Type:** CLI-only (No public API)  
**HelixAgent Provider:** [internal/llm/providers/claude/](../../../internal/llm/providers/claude/)  
**Analysis Date:** 2026-04-03  

---

## Executive Summary

Claude Code is a **CLI-only tool** with no public REST API, WebSocket, or MCP interface. It communicates directly with Anthropic's API using a proprietary protocol. However, HelixAgent can replicate Claude Code's functionality through its Claude provider and extend it with ensemble capabilities.

**HelixAgent Alternative:** Use `internal/llm/providers/claude/` with MCP tools for equivalent functionality.

---

## Claude Code Architecture

### Internal Communication Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                     CLAUDE CODE FLOW                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   Terminal → Claude Code CLI → Anthropic API (Private)          │
│       │                              │                          │
│       │                              │ HTTPS + API Key          │
│       │                              ▼                          │
│       │                    Claude 3.5 Sonnet                    │
│       │                              │                          │
│       │                              │ Tool Use Loop            │
│       │                              ▼                          │
│       │                    [File Ops, Bash, Search]             │
│       │                                                          │
│   No public API endpoint exposed!                                │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Protocol Details

```typescript
// Claude Code uses Anthropic's Messages API internally
interface ClaudeCodeRequest {
  model: "claude-3-5-sonnet-20241022";
  max_tokens: number;
  messages: Message[];
  tools?: Tool[];  // 7 built-in tools
  system?: string;
}

// Built-in Tools:
// 1. read_file    - Read file contents
// 2. write_file   - Write/modify files  
// 3. bash         - Execute shell commands
// 4. glob         - Find files by pattern
// 5. grep         - Search file contents
// 6. ls           - List directory contents
// 7. view         - View code with line numbers
```

---

## HelixAgent Equivalent Implementation

### Source Code Reference

**Primary Provider:**
- File: [`internal/llm/providers/claude/claude.go`](../../../internal/llm/providers/claude/claude.go)
- Tests: [`internal/llm/providers/claude/claude_test.go`](../../../internal/llm/providers/claude/claude_test.go)
- Config: [`internal/llm/providers/claude/config.go`](../../../internal/llm/providers/claude/config.go)

**MCP Tool Adapters:**
- File Reader: [`internal/mcp/adapters/filesystem.go`](../../../internal/mcp/adapters/filesystem.go)
- Bash Executor: [`internal/mcp/adapters/bash.go`](../../../internal/mcp/adapters/bash.go)
- Search: [`internal/mcp/adapters/search.go`](../../../internal/mcp/adapters/search.go)

**Handler:**
- Completions: [`internal/handlers/completion.go`](../../../internal/handlers/completion.go#L45-L120)
- Tools: [`internal/handlers/tools.go`](../../../internal/handlers/tools.go)

### HelixAgent API Endpoint

```
POST /v1/chat/completions
```

**Source:** [`internal/handlers/completion.go:45`](../../../internal/handlers/completion.go#L45)

**Request:**
```json
{
  "model": "claude-3-5-sonnet-20241022",
  "messages": [
    {"role": "user", "content": "Read file src/main.py"}
  ],
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "read_file",
        "parameters": {
          "path": "string"
        }
      }
    }
  ],
  "stream": true
}
```

**Response:**
```json
{
  "id": "chatcmpl-xxx",
  "model": "claude-3-5-sonnet-20241022",
  "choices": [{
    "index": 0,
    "message": {
      "role": "assistant",
      "content": null,
      "tool_calls": [{
        "id": "call_xxx",
        "type": "function",
        "function": {
          "name": "read_file",
          "arguments": "{\"path\": \"src/main.py\"}"
        }
      }]
    }
  }]
}
```

---

## Feature Comparison: Claude Code vs HelixAgent

| Feature | Claude Code | HelixAgent Equivalent | Status |
|---------|-------------|----------------------|--------|
| **Core API** | Anthropic Messages API | Anthropic Messages API | ✅ Parity |
| **Tool Use** | 7 built-in tools | MCP tools | ✅ Parity |
| **File Reading** | `read_file` | MCP `filesystem/read` | ✅ Parity |
| **File Writing** | `write_file` | MCP `filesystem/write` | ✅ Parity |
| **Bash Execution** | `bash` | MCP `bash/execute` | ✅ Superior (sandboxed) |
| **Glob Search** | `glob` | MCP `search/glob` | ✅ Parity |
| **Grep Search** | `grep` | MCP `search/grep` | ✅ Parity |
| **Directory Listing** | `ls` | MCP `filesystem/ls` | ✅ Parity |
| **Code View** | `view` | MCP `filesystem/read` | ✅ Parity |
| **Streaming** | ✅ | ✅ | ✅ Parity |
| **Multi-Provider** | ❌ | ✅ | 🏆 HelixAgent |
| **Ensemble** | ❌ | ✅ | 🏆 HelixAgent |
| **Debate** | ❌ | ✅ | 🏆 HelixAgent |
| **Persistence** | ❌ | ✅ PostgreSQL | 🏆 HelixAgent |
| **Caching** | ❌ | ✅ Redis | 🏆 HelixAgent |

---

## Implementation Deep Dive

### HelixAgent Claude Provider

**Source:** [`internal/llm/providers/claude/claude.go`](../../../internal/llm/providers/claude/claude.go)

```go
package claude

import (
    "github.com/anthropics/anthropic-sdk-go"
)

// ClaudeProvider implements LLMProvider for Anthropic
type ClaudeProvider struct {
    client *anthropic.Client
    config *Config
}

// GenerateCompletion implements tool use like Claude Code
func (p *ClaudeProvider) GenerateCompletion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
    // Source: internal/llm/providers/claude/claude.go#L78-156
    message, err := p.client.Messages.Create(ctx, anthropic.MessageCreateParams{
        Model:     req.Model,
        MaxTokens: req.MaxTokens,
        Messages:  convertMessages(req.Messages),
        Tools:     convertTools(req.Tools),  // MCP tool conversion
        System:    req.System,
    })
    // ...
}
```

### Tool Use Loop Implementation

**Source:** [`internal/handlers/tools.go`](../../../internal/handlers/tools.go)

```go
// ExecuteToolLoop runs the tool use cycle like Claude Code
func ExecuteToolLoop(ctx context.Context, provider llm.Provider, initialReq *Request) (*Response, error) {
    // Source: internal/handlers/tools.go#L45-120
    
    for {
        resp, err := provider.GenerateCompletion(ctx, req)
        if err != nil {
            return nil, err
        }
        
        // Check if tool calls needed
        if len(resp.ToolCalls) == 0 {
            return resp, nil  // Final response
        }
        
        // Execute tools
        results := executeToolCalls(resp.ToolCalls)
        
        // Add results to conversation
        req.Messages = append(req.Messages, Message{
            Role:    "tool",
            Content: formatResults(results),
        })
    }
}
```

---

## WebSocket Support

### Claude Code: No WebSocket

Claude Code does **not** use WebSockets. It uses:
- HTTP/2 for API communication
- Server-Sent Events (SSE) for streaming

### HelixAgent WebSocket Alternative

**Source:** [`internal/handlers/websocket.go`](../../../internal/handlers/websocket.go)

```go
// WebSocket endpoint for real-time interaction
// Source: internal/handlers/websocket.go#L30-85

ws://localhost:7061/v1/stream
```

**Protocol:**
```json
// Client → Server
{
  "type": "message",
  "payload": {
    "model": "claude-3-5-sonnet",
    "content": "Hello",
    "tools": ["read_file", "bash"]
  }
}

// Server → Client (streaming)
{
  "type": "delta",
  "payload": {
    "content": "I'll help",
    "tool_call": null
  }
}
```

---

## MCP Protocol Support

### Claude Code: No MCP

Claude Code uses **proprietary tools**, not MCP.

### HelixAgent MCP Implementation

**Source:** [`internal/mcp/`](../../../internal/mcp/)

```
internal/mcp/
├── adapters/
│   ├── filesystem.go    # File operations
│   ├── bash.go         # Shell execution
│   ├── search.go       # Search operations
│   └── git.go          # Git operations
├── server/
│   └── server.go       # MCP server implementation
└── protocol/
    └── protocol.go     # MCP protocol definitions
```

**Claude Code Tools → MCP Mapping:**

| Claude Code Tool | MCP Tool | Source |
|------------------|----------|--------|
| `read_file` | `filesystem/read` | [adapters/filesystem.go](../../../internal/mcp/adapters/filesystem.go#L45) |
| `write_file` | `filesystem/write` | [adapters/filesystem.go](../../../internal/mcp/adapters/filesystem.go#L78) |
| `bash` | `bash/execute` | [adapters/bash.go](../../../internal/mcp/adapters/bash.go#L52) |
| `glob` | `search/glob` | [adapters/search.go](../../../internal/mcp/adapters/search.go#L60) |
| `grep` | `search/grep` | [adapters/search.go](../../../internal/mcp/adapters/search.go#L95) |
| `ls` | `filesystem/ls` | [adapters/filesystem.go](../../../internal/mcp/adapters/filesystem.go#L110) |
| `view` | `filesystem/read` | [adapters/filesystem.go](../../../internal/mcp/adapters/filesystem.go#L45) |

---

## Authentication Comparison

### Claude Code

```bash
# Environment variable only
export ANTHROPIC_API_KEY="sk-ant-..."

# No other auth methods supported
```

### HelixAgent

**Source:** [`internal/auth/`](../../../internal/auth/)

```yaml
# Multiple auth methods
authentication:
  methods:
    - api_key      # Simple API key
    - oauth        # OAuth 2.0
    - jwt          # JWT tokens
    - mTLS         # Mutual TLS
  
providers:
  claude:
    type: api_key
    key_env: ANTHROPIC_API_KEY
    rate_limit: 100/min
```

---

## Error Handling Comparison

### Claude Code Errors

```typescript
// Anthropic API errors
interface AnthropicError {
  error: {
    type: "invalid_request_error" | "authentication_error" | "rate_limit_error";
    message: string;
    code: string;
  }
}
```

### HelixAgent Errors

**Source:** [`internal/handlers/errors.go`](../../../internal/handlers/errors.go)

```go
// Standardized error responses
{
  "error": {
    "code": "rate_limit_exceeded",
    "message": "Rate limit exceeded",
    "type": "rate_limit_error",
    "provider": "claude",
    "retry_after": 60
  }
}
```

---

## Integration Guide: Using HelixAgent as Claude Code Replacement

### Configuration

**File:** [`configs/claude-equivalent.yaml`](../../../configs/claude-equivalent.yaml)

```yaml
# HelixAgent configuration for Claude Code parity
server:
  port: 7061
  host: localhost

providers:
  claude:
    type: anthropic
    api_key: ${ANTHROPIC_API_KEY}
    model: claude-3-5-sonnet-20241022
    temperature: 0.7

mcp:
  filesystem:
    enabled: true
    root_dir: ${PWD}
    allow_write: true
  
  bash:
    enabled: true
    sandbox: true
    allowed_commands: ["ls", "cat", "grep", "find", "git"]
  
  search:
    enabled: true
    max_results: 50

features:
  streaming: true
  tool_use: true
  ensemble: false  # Set true for multi-model
```

### Usage Example

```bash
# Start HelixAgent
docker-compose up helixagent

# Use instead of Claude Code
curl -X POST http://localhost:7061/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-5-sonnet",
    "messages": [{"role": "user", "content": "Read README.md"}],
    "tools": [{"type": "filesystem", "operation": "read"}],
    "stream": true
  }'
```

---

## Strengths & Weaknesses Analysis

### Claude Code Strengths

1. **Simplicity**
   - Single command install
   - No configuration needed
   - Works immediately

2. **Terminal UX**
   - Rich inline display
   - Syntax highlighting
   - Progress indicators

3. **Tool Use UX**
   - Natural conversation flow
   - Automatic tool selection
   - Inline results

### HelixAgent Strengths

1. **Extensibility**
   - MCP protocol for custom tools
   - 22+ providers
   - Plugin system

2. **Enterprise**
   - Authentication & authorization
   - Rate limiting
   - Audit trails

3. **Intelligence**
   - Ensemble voting
   - Debate orchestration
   - Multi-model consensus

### When to Use Which

| Scenario | Winner | Reason |
|----------|--------|--------|
| Quick setup | Claude Code | Zero config |
| Terminal UX | Claude Code | Native experience |
| Tool variety | HelixAgent | MCP ecosystem |
| Multi-provider | HelixAgent | 22+ providers |
| Ensemble decisions | HelixAgent | Voting system |
| Enterprise deployment | HelixAgent | Security features |
| CI/CD integration | HelixAgent | API server |

---

## Source Code Reference Index

### HelixAgent Files Related to Claude Code Functionality

| Functionality | HelixAgent File | Lines |
|---------------|-----------------|-------|
| Claude Provider | `internal/llm/providers/claude/claude.go` | 156 |
| Provider Config | `internal/llm/providers/claude/config.go` | 45 |
| Provider Tests | `internal/llm/providers/claude/claude_test.go` | 234 |
| Completion Handler | `internal/handlers/completion.go` | 120 |
| Tool Handler | `internal/handlers/tools.go` | 95 |
| MCP Filesystem | `internal/mcp/adapters/filesystem.go` | 145 |
| MCP Bash | `internal/mcp/adapters/bash.go` | 98 |
| MCP Search | `internal/mcp/adapters/search.go` | 112 |
| Error Handling | `internal/handlers/errors.go` | 67 |
| WebSocket | `internal/handlers/websocket.go` | 85 |

---

## Conclusion

Claude Code is a **closed, CLI-only tool** with no public API. HelixAgent provides **equivalent functionality** through:

1. **Same underlying API** (Anthropic Messages API)
2. **MCP tools** (equivalent to Claude Code's 7 tools)
3. **Extended capabilities** (ensemble, debate, 22+ providers)

**Recommendation:** Use HelixAgent with Claude provider configured for parity, plus additional features.

---

*Documentation: API Specification & Cross-Reference*  
*Last Updated: 2026-04-03*  
*HelixAgent Commit: 7ec2da53*
