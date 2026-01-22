# User Manual 07: Protocol Integration Guide

## Overview

HelixAgent supports multiple communication protocols for integrating with development tools, language servers, and AI agents. This guide covers the MCP, LSP, and ACP protocols.

## Supported Protocols

| Protocol | Purpose | Endpoint |
|----------|---------|----------|
| MCP | Model Context Protocol | `/v1/mcp` |
| LSP | Language Server Protocol | `/v1/lsp` |
| ACP | Agent Communication Protocol | `/v1/acp` |

## Model Context Protocol (MCP)

MCP enables AI models to interact with external tools and resources through a standardized interface.

### Available MCP Tools

```bash
# List available MCP tools
curl http://localhost:8080/v1/mcp/tools
```

Response:
```json
{
    "tools": [
        {
            "name": "read_file",
            "description": "Read contents of a file",
            "parameters": {
                "path": "string (required)"
            }
        },
        {
            "name": "write_file",
            "description": "Write contents to a file",
            "parameters": {
                "path": "string (required)",
                "content": "string (required)"
            }
        },
        {
            "name": "search_files",
            "description": "Search for files matching a pattern",
            "parameters": {
                "pattern": "string (required)",
                "directory": "string (optional)"
            }
        }
    ]
}
```

### Executing MCP Tools

```bash
# Execute a tool
curl -X POST http://localhost:8080/v1/mcp/tools/read_file \
  -H "Content-Type: application/json" \
  -d '{"path": "/path/to/file.txt"}'
```

Response:
```json
{
    "result": {
        "content": "File contents here...",
        "metadata": {
            "size": 1234,
            "modified": "2024-01-15T10:30:00Z"
        }
    }
}
```

### MCP Resources

Access external resources through MCP:

```bash
# List available resources
curl http://localhost:8080/v1/mcp/resources

# Get specific resource
curl http://localhost:8080/v1/mcp/resources/project-files
```

### MCP Configuration

Configure MCP in your environment:

```yaml
# config/mcp.yaml
mcp:
  enabled: true
  tools:
    - read_file
    - write_file
    - search_files
    - execute_command
  resources:
    - project-files
    - documentation
  security:
    sandbox: true
    allowed_paths:
      - /home/user/projects
    denied_commands:
      - rm -rf
```

## Language Server Protocol (LSP)

LSP integration provides IDE-like features for code analysis and completion.

### LSP Capabilities

- Code completion
- Hover information
- Go to definition
- Find references
- Diagnostics
- Code actions
- Formatting

### Code Completion

```bash
curl -X POST http://localhost:8080/v1/lsp/completions \
  -H "Content-Type: application/json" \
  -d '{
    "document": {
      "uri": "file:///path/to/main.go",
      "content": "package main\n\nfunc main() {\n\tfmt.Pr"
    },
    "position": {
      "line": 3,
      "character": 9
    }
  }'
```

Response:
```json
{
    "completions": [
        {
            "label": "Print",
            "kind": "Function",
            "detail": "func Print(a ...any) (n int, err error)",
            "documentation": "Print formats using default formats..."
        },
        {
            "label": "Printf",
            "kind": "Function",
            "detail": "func Printf(format string, a ...any) (n int, err error)"
        },
        {
            "label": "Println",
            "kind": "Function",
            "detail": "func Println(a ...any) (n int, err error)"
        }
    ]
}
```

### Hover Information

```bash
curl -X POST http://localhost:8080/v1/lsp/hover \
  -H "Content-Type: application/json" \
  -d '{
    "document": {
      "uri": "file:///path/to/main.go"
    },
    "position": {
      "line": 10,
      "character": 5
    }
  }'
```

### Go to Definition

```bash
curl -X POST http://localhost:8080/v1/lsp/definition \
  -H "Content-Type: application/json" \
  -d '{
    "document": {
      "uri": "file:///path/to/main.go"
    },
    "position": {
      "line": 15,
      "character": 12
    }
  }'
```

Response:
```json
{
    "location": {
        "uri": "file:///path/to/utils.go",
        "range": {
            "start": {"line": 25, "character": 0},
            "end": {"line": 25, "character": 15}
        }
    }
}
```

### Supported Language Servers

| Language | Server | Command |
|----------|--------|---------|
| Go | gopls | `gopls serve` |
| Python | pylsp | `pylsp` |
| TypeScript | tsserver | `typescript-language-server --stdio` |
| Rust | rust-analyzer | `rust-analyzer` |
| Java | jdtls | `jdtls` |

### LSP Configuration

```yaml
# config/lsp.yaml
lsp:
  enabled: true
  servers:
    go:
      command: gopls
      args: ["serve"]
      root_markers: ["go.mod"]
    python:
      command: pylsp
      root_markers: ["pyproject.toml", "setup.py"]
  timeout: 30s
  max_completions: 50
```

## Agent Communication Protocol (ACP)

ACP enables communication between AI agents for collaborative tasks.

### Agent Registration

```bash
# Register an agent
curl -X POST http://localhost:8080/v1/acp/agents \
  -H "Content-Type: application/json" \
  -d '{
    "id": "code-reviewer",
    "name": "Code Review Agent",
    "capabilities": ["code_review", "security_audit"],
    "endpoint": "http://localhost:8081/agent"
  }'
```

### Sending Messages to Agents

```bash
# Send message to agent
curl -X POST http://localhost:8080/v1/acp/messages \
  -H "Content-Type: application/json" \
  -d '{
    "to": "code-reviewer",
    "type": "request",
    "content": {
      "action": "review_code",
      "file": "/path/to/main.go"
    }
  }'
```

Response:
```json
{
    "message_id": "msg-123",
    "status": "delivered",
    "response": {
        "findings": [
            {
                "severity": "warning",
                "line": 42,
                "message": "Consider error handling here"
            }
        ]
    }
}
```

### Agent Discovery

```bash
# List available agents
curl http://localhost:8080/v1/acp/agents

# Get agent capabilities
curl http://localhost:8080/v1/acp/agents/code-reviewer/capabilities
```

### ACP Configuration

```yaml
# config/acp.yaml
acp:
  enabled: true
  discovery:
    enabled: true
    interval: 60s
  agents:
    - id: code-reviewer
      name: Code Review Agent
      endpoint: http://localhost:8081
    - id: test-runner
      name: Test Runner Agent
      endpoint: http://localhost:8082
  security:
    require_auth: true
    allowed_agents:
      - code-reviewer
      - test-runner
```

## Protocol Integration with AI Debate

HelixAgent's protocols integrate seamlessly with the AI debate system:

```bash
# Start debate with protocol access
curl -X POST http://localhost:8080/v1/debates \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "Review this code for security issues",
    "context": {
      "mcp_tools": ["read_file", "search_files"],
      "lsp_enabled": true,
      "file": "/path/to/main.go"
    },
    "participants": ["claude", "gemini", "deepseek"]
  }'
```

## Real-time Protocol Events (SSE)

Subscribe to protocol events:

```bash
# Subscribe to MCP events
curl http://localhost:8080/v1/events/mcp

# Subscribe to LSP events
curl http://localhost:8080/v1/events/lsp

# Subscribe to ACP events
curl http://localhost:8080/v1/events/acp
```

Event format:
```
event: tool_execution
data: {"tool":"read_file","status":"completed","duration_ms":45}

event: completion_request
data: {"document":"main.go","completions_returned":15}

event: agent_message
data: {"from":"code-reviewer","type":"response"}
```

## Security Considerations

### MCP Security

- Sandbox tool execution
- Whitelist allowed paths
- Deny dangerous commands
- Rate limit tool calls

### LSP Security

- Validate document URIs
- Sandbox code execution
- Limit response sizes
- Timeout long operations

### ACP Security

- Authenticate agents
- Encrypt messages
- Validate payloads
- Audit message logs

## Troubleshooting

### MCP Issues

| Issue | Solution |
|-------|----------|
| Tool not found | Check tool is enabled in config |
| Permission denied | Verify path is in allowed_paths |
| Timeout | Increase timeout or check tool execution |

### LSP Issues

| Issue | Solution |
|-------|----------|
| No completions | Verify language server is running |
| Slow responses | Check server resource usage |
| Definition not found | Ensure project is indexed |

### ACP Issues

| Issue | Solution |
|-------|----------|
| Agent not found | Check agent registration |
| Message timeout | Verify agent endpoint |
| Auth failed | Check agent credentials |

## Advanced MCP Server Configuration

### Custom MCP Server Setup

Create and register custom MCP servers:

```yaml
# configs/mcp-servers.yaml
mcp_servers:
  custom_server:
    name: "Custom Data Server"
    type: "stdio"
    command: "/path/to/custom-server"
    args: ["--mode", "production"]
    env:
      API_KEY: "${CUSTOM_API_KEY}"
    capabilities:
      tools: true
      resources: true
      prompts: true
    timeout: 30s

  database_server:
    name: "Database Tools"
    type: "http"
    endpoint: "http://localhost:9000/mcp"
    auth:
      type: "bearer"
      token: "${DB_MCP_TOKEN}"
```

### MCP Server Registry

```bash
# List registered MCP servers
curl http://localhost:7061/v1/mcp/servers

# Register new server
curl -X POST http://localhost:7061/v1/mcp/servers \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "github-server",
    "type": "npm",
    "package": "@modelcontextprotocol/server-github",
    "config": {
      "token": "${GITHUB_TOKEN}"
    }
  }'

# Check server health
curl http://localhost:7061/v1/mcp/servers/github-server/health
```

### Built-in MCP Adapters (40+)

| Adapter | Purpose | Configuration |
|---------|---------|---------------|
| `aws_s3` | S3 file operations | AWS credentials |
| `brave_search` | Web search | Brave API key |
| `datadog` | Monitoring | Datadog API key |
| `docker` | Container management | Docker socket |
| `figma` | Design files | Figma token |
| `gitlab` | Git operations | GitLab token |
| `google_drive` | File storage | OAuth credentials |
| `kubernetes` | K8s management | Kubeconfig |
| `mongodb` | Database operations | Connection string |
| `notion` | Note management | Notion token |
| `puppeteer` | Browser automation | None required |
| `sentry` | Error tracking | Sentry DSN |
| `slack` | Messaging | Slack token |

## Custom Tool Development

### Creating Custom MCP Tools

```go
// internal/mcp/custom/my_tool.go
package custom

import (
    "context"
    "dev.helix.agent/internal/mcp"
)

type MyCustomTool struct {
    config MyToolConfig
}

func (t *MyCustomTool) Name() string {
    return "my_custom_tool"
}

func (t *MyCustomTool) Description() string {
    return "Performs custom operations"
}

func (t *MyCustomTool) Schema() mcp.ToolSchema {
    return mcp.ToolSchema{
        Type: "object",
        Properties: map[string]mcp.PropertySchema{
            "input": {
                Type:        "string",
                Description: "Input data",
                Required:    true,
            },
            "options": {
                Type:        "object",
                Description: "Optional configuration",
                Required:    false,
            },
        },
    }
}

func (t *MyCustomTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    input := params["input"].(string)

    // Perform custom operation
    result := t.processInput(input)

    return map[string]interface{}{
        "result": result,
        "status": "success",
    }, nil
}

// Register the tool
func init() {
    mcp.RegisterTool(&MyCustomTool{})
}
```

### Tool Registration API

```bash
# Register tool via API
curl -X POST http://localhost:7061/v1/mcp/tools/register \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{
    "name": "custom_analyzer",
    "description": "Analyzes data with custom logic",
    "handler": "http://localhost:9001/analyze",
    "schema": {
      "type": "object",
      "properties": {
        "data": {"type": "string", "required": true},
        "format": {"type": "string", "enum": ["json", "xml"]}
      }
    }
  }'
```

## WebSocket Protocol Integration

### WebSocket Connection

```javascript
// Client-side WebSocket connection
const ws = new WebSocket('wss://localhost:7061/v1/ws');

ws.onopen = () => {
    // Subscribe to protocol events
    ws.send(JSON.stringify({
        type: 'subscribe',
        channels: ['mcp', 'lsp', 'acp', 'debates']
    }));
};

ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    switch (message.type) {
        case 'mcp_tool_result':
            handleToolResult(message.data);
            break;
        case 'lsp_completion':
            handleCompletion(message.data);
            break;
        case 'debate_update':
            handleDebateUpdate(message.data);
            break;
    }
};
```

### WebSocket Protocol Messages

```json
// Tool execution request
{
    "type": "mcp_execute",
    "request_id": "req-123",
    "tool": "read_file",
    "params": {
        "path": "/path/to/file.txt"
    }
}

// Tool execution response
{
    "type": "mcp_result",
    "request_id": "req-123",
    "result": {
        "content": "File contents...",
        "metadata": {"size": 1024}
    }
}

// LSP completion request
{
    "type": "lsp_complete",
    "request_id": "req-456",
    "document": "file:///main.go",
    "position": {"line": 10, "character": 15}
}

// Debate subscription
{
    "type": "subscribe_debate",
    "debate_id": "debate-789"
}
```

### WebSocket Configuration

```yaml
websocket:
  enabled: true
  path: "/v1/ws"
  ping_interval: 30s
  pong_timeout: 10s
  max_message_size: "1MB"
  compression: true
  auth:
    required: true
    methods: ["bearer", "api_key"]
  rate_limit:
    messages_per_second: 100
    burst: 200
```

## Real-Time Streaming Patterns

### SSE Streaming

```bash
# Stream debate progress
curl -N http://localhost:7061/v1/debates/{id}/stream \
  -H "Accept: text/event-stream"

# Stream completion responses
curl -N -X POST http://localhost:7061/v1/completions \
  -H "Content-Type: application/json" \
  -H "Accept: text/event-stream" \
  -d '{"model": "claude-sonnet-4-20250514", "messages": [...], "stream": true}'
```

### Streaming Response Format

```
event: message_start
data: {"id":"msg-001","type":"message_start"}

event: content_block_start
data: {"type":"content_block_start","index":0}

event: content_block_delta
data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"Hello"}}

event: content_block_delta
data: {"type":"content_block_delta","delta":{"type":"text_delta","text":" world"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: message_stop
data: {"type":"message_stop"}
```

### Handling Stream Interruptions

```go
// Server-side streaming with reconnection support
func (h *Handler) StreamResponse(w http.ResponseWriter, r *http.Request) {
    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "Streaming not supported", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    w.Header().Set("X-Accel-Buffering", "no")

    // Send keep-alive comments every 15 seconds
    ticker := time.NewTicker(15 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case event := <-eventChan:
            fmt.Fprintf(w, "event: %s\n", event.Type)
            fmt.Fprintf(w, "data: %s\n\n", event.Data)
            flusher.Flush()
        case <-ticker.C:
            fmt.Fprintf(w, ": keep-alive\n\n")
            flusher.Flush()
        case <-r.Context().Done():
            return
        }
    }
}
```

## Protocol Error Handling

### Error Response Format

```json
{
    "error": {
        "code": "TOOL_EXECUTION_FAILED",
        "message": "Failed to execute tool: permission denied",
        "details": {
            "tool": "write_file",
            "path": "/etc/passwd",
            "reason": "Path not in allowed list"
        },
        "request_id": "req-123",
        "timestamp": "2026-01-22T10:30:00Z"
    }
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `TOOL_NOT_FOUND` | 404 | Requested tool doesn't exist |
| `TOOL_EXECUTION_FAILED` | 500 | Tool execution error |
| `PERMISSION_DENIED` | 403 | Access not allowed |
| `VALIDATION_ERROR` | 400 | Invalid parameters |
| `TIMEOUT` | 504 | Operation timed out |
| `RATE_LIMITED` | 429 | Too many requests |
| `SERVER_NOT_AVAILABLE` | 503 | MCP server offline |

### Retry Strategy

```yaml
protocols:
  retry:
    enabled: true
    max_attempts: 3
    initial_delay: 100ms
    max_delay: 5s
    backoff_multiplier: 2
    retryable_errors:
      - TIMEOUT
      - SERVER_NOT_AVAILABLE
    non_retryable_errors:
      - PERMISSION_DENIED
      - VALIDATION_ERROR
```

### Client-Side Error Handling

```javascript
async function executeTool(toolName, params) {
    try {
        const response = await fetch('/v1/mcp/tools/' + toolName, {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(params)
        });

        if (!response.ok) {
            const error = await response.json();
            switch (error.error.code) {
                case 'RATE_LIMITED':
                    // Wait and retry
                    await sleep(error.error.details.retry_after * 1000);
                    return executeTool(toolName, params);
                case 'TIMEOUT':
                    // Retry with longer timeout
                    return executeTool(toolName, {...params, timeout: 60000});
                case 'PERMISSION_DENIED':
                    throw new PermissionError(error.error.message);
                default:
                    throw new ToolError(error.error);
            }
        }

        return response.json();
    } catch (e) {
        if (e instanceof TypeError) {
            // Network error - retry
            await sleep(1000);
            return executeTool(toolName, params);
        }
        throw e;
    }
}
```

## Protocol Monitoring

### Metrics

```yaml
metrics:
  protocols:
    mcp:
      - tool_execution_total
      - tool_execution_duration_seconds
      - tool_errors_total
    lsp:
      - completion_requests_total
      - completion_latency_seconds
      - server_health
    acp:
      - messages_sent_total
      - messages_received_total
      - agent_response_time_seconds
```

### Dashboard Queries

```promql
# MCP tool success rate
sum(rate(mcp_tool_execution_total{status="success"}[5m]))
/ sum(rate(mcp_tool_execution_total[5m]))

# LSP completion latency p99
histogram_quantile(0.99, rate(lsp_completion_latency_seconds_bucket[5m]))

# ACP message throughput
sum(rate(acp_messages_total[1m]))
```

## Next Steps

- [Troubleshooting Guide](08-troubleshooting.md)
- [API Reference](04-api-reference.md)
- [Deployment Guide](05-deployment-guide.md)

---

*Protocol Integration Guide Version: 2.0.0*
*Last Updated: January 2026*
