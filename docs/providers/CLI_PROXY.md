# CLI Proxy Mechanism for OAuth/Free Providers

This document describes the CLI proxy mechanism used by HelixAgent for providers where direct API access is restricted.

## Overview

HelixAgent supports three types of provider authentication:

1. **API Key** - Direct API access using bearer tokens
2. **OAuth** - OAuth2 tokens from CLI tools (product-restricted)
3. **Free/Anonymous** - No authentication required

For OAuth and free providers, direct API calls often fail due to:
- **Claude OAuth**: Tokens are product-restricted to Claude Code only
- **Qwen OAuth**: Tokens are for Qwen Portal, not DashScope API
- **Zen Free**: Anonymous API may be unreliable

**Solution**: CLI Proxy routes requests through the CLI tools themselves, which have proper authentication context.

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  HelixAgent     │    │  CLI Provider   │    │   CLI Tool      │
│  Provider       │───>│  (Proxy)        │───>│  (claude/qwen/  │
│  Registry       │    │                 │    │   opencode)     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                     │                      │
         │  1. Check OAuth     │  2. Execute CLI      │
         │     enabled         │     command          │
         │                     │                      │
         │  3. Check CLI       │  4. Parse output     │
         │     available       │     (JSON/text)      │
         v                     v                      v
    Direct API           CLI Proxy              Response
    (if API key)       (if OAuth/free)          Returned
```

## Supported Providers

| Provider | Method | Command/Protocol | JSON Output | Trigger Condition |
|----------|--------|------------------|-------------|-------------------|
| **Claude** | CLI Proxy | `claude -p "prompt"` | No | `CLAUDE_CODE_USE_OAUTH_CREDENTIALS=true` + no API key |
| **Qwen** | ACP (preferred) | `qwen --acp` | JSON-RPC | `QWEN_CODE_USE_OAUTH_CREDENTIALS=true` + no API key |
| **Qwen** | CLI Proxy (fallback) | `qwen -p "prompt"` | No | ACP not available |
| **Zen (OpenCode)** | HTTP Server (preferred) | `opencode serve --port 4096` REST API | Yes | No `OPENCODE_API_KEY` (free mode) |
| **Zen (OpenCode)** | CLI Proxy (fallback) | `opencode -p "prompt" -f json` | Yes | HTTP server not available |

## Qwen ACP (Agent Communication Protocol)

Qwen Code supports ACP - a JSON-RPC 2.0 protocol over stdin/stdout. This is preferred over CLI because:

- **Session Management**: Maintains conversation history across requests
- **Streaming Support**: Real-time response streaming via notifications
- **Lower Latency**: Single process for multiple requests
- **Better Authentication**: Handles OAuth refresh automatically

### ACP Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  HelixAgent     │    │  QwenACP        │    │  qwen --acp     │
│  Provider       │───>│  Provider       │───>│  (long-running) │
│  Registry       │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │                       │
                              │  stdin: JSON-RPC      │
                              │  stdout: JSON-RPC     │
                              └───────────────────────┘
```

### ACP Methods Used

| Method | Description |
|--------|-------------|
| `initialize` | Initialize ACP connection |
| `session/new` | Create a new conversation session |
| `session/prompt` | Send prompt and get response |
| `session/update` | Streaming updates (notifications) |

### Key File

`internal/llm/providers/qwen/qwen_acp.go` - Qwen ACP provider implementation

## Environment Variables

```bash
# Enable Claude OAuth CLI proxy
CLAUDE_CODE_USE_OAUTH_CREDENTIALS=true

# Enable Qwen OAuth CLI proxy
QWEN_CODE_USE_OAUTH_CREDENTIALS=true

# Zen/OpenCode API key (optional - CLI proxy for free mode if not set)
OPENCODE_API_KEY=your-api-key
```

## Provider Priority

For each provider, HelixAgent follows this priority:

1. **API Key** (if set) → Direct API access
2. **OAuth + CLI available** → CLI proxy
3. **OAuth + no CLI** → Provider not registered (warning logged)
4. **Free + CLI available** → CLI proxy
5. **Free + no CLI** → Anonymous API (fallback)

## Key Files

### CLI Providers

| File | Description |
|------|-------------|
| `internal/llm/providers/claude/claude_cli.go` | Claude CLI provider implementation |
| `internal/llm/providers/qwen/qwen_cli.go` | Qwen CLI provider implementation |
| `internal/llm/providers/zen/zen_cli.go` | Zen/OpenCode CLI provider with JSON parsing |

### Registration Logic

| File | Description |
|------|-------------|
| `internal/services/provider_registry.go` | Registry logic for Claude/Qwen OAuth |
| `internal/services/provider_discovery.go` | Discovery logic for Zen free mode |

### OAuth Credentials

| File | Description |
|------|-------------|
| `internal/auth/oauth_credentials/reader.go` | OAuth credential reader |
| `internal/auth/oauth_credentials/claude.go` | Claude OAuth handling |
| `internal/auth/oauth_credentials/qwen.go` | Qwen OAuth handling |

## CLI Provider Interface

All CLI providers implement the standard `LLMProvider` interface:

```go
type LLMProvider interface {
    Complete(ctx context.Context, req *LLMRequest) (*LLMResponse, error)
    CompleteStream(ctx context.Context, req *LLMRequest) (<-chan *LLMResponse, error)
    HealthCheck() error
    GetCapabilities() *ProviderCapabilities
    ValidateConfig(config map[string]interface{}) (bool, []string)
}
```

### Additional CLI-Specific Methods

```go
// Check if CLI tool is available
IsCLIAvailable() bool

// Get CLI availability error
GetCLIError() error

// Get current model
GetCurrentModel() string

// Set model to use
SetModel(model string)
```

## JSON Output Parsing (Zen/OpenCode)

OpenCode supports structured JSON output with `-f json` flag:

```bash
opencode -p "prompt" -f json
```

Output format:
```json
{
  "response": "The AI's response content"
}
```

The Zen CLI provider parses this format:

```go
type openCodeJSONResponse struct {
    Response string `json:"response"`
}

func (p *ZenCLIProvider) parseJSONResponse(rawOutput string) string {
    var jsonResp openCodeJSONResponse
    if err := json.Unmarshal([]byte(rawOutput), &jsonResp); err == nil {
        if jsonResp.Response != "" {
            return jsonResp.Response
        }
    }
    // Fallback: return raw output if JSON parsing fails
    return rawOutput
}
```

## Testing

### Unit Tests

```bash
# Run CLI provider unit tests
go test -v ./internal/llm/providers/claude/... -run "CLI"
go test -v ./internal/llm/providers/qwen/... -run "CLI"
go test -v ./internal/llm/providers/zen/... -run "CLI"
```

### Integration Tests

```bash
# Run CLI proxy integration tests
go test -v ./tests/integration/... -run "CLIProxy"
```

### Challenge Validation

```bash
# Run CLI proxy challenge (50 tests)
./challenges/scripts/cli_proxy_challenge.sh
```

## Troubleshooting

### CLI Not Available

If you see "Claude CLI not available" or similar:

1. Ensure the CLI tool is installed and in PATH
2. Ensure you're logged in (`claude auth login`, `qwen auth login`)
3. Check CLI works manually: `claude --version`

### OAuth Tokens Expired

If OAuth tokens expire:

1. Re-authenticate with the CLI tool
2. Claude: `claude auth login`
3. Qwen: `qwen auth login`

### JSON Parsing Errors (Zen)

If Zen CLI output isn't parsing correctly:

1. Ensure `-f json` flag is supported by your OpenCode version
2. Check raw output format matches expected JSON structure
3. The provider falls back to raw output if JSON parsing fails

## Performance Considerations

CLI proxy has higher latency than direct API:

| Method | Latency | Throughput |
|--------|---------|------------|
| Direct API | ~100ms | High |
| CLI Proxy | ~500-2000ms | Lower |

For high-performance scenarios, prefer API key authentication.

## Security

- OAuth tokens stay on local filesystem (never sent over network by HelixAgent)
- CLI tools handle their own authentication
- No credentials are logged or exposed

## Version History

| Version | Changes |
|---------|---------|
| v1.0.0 | Initial CLI proxy implementation |
| v1.1.0 | Added JSON output parsing for Zen |
| v1.2.0 | Added comprehensive tests and challenge |
