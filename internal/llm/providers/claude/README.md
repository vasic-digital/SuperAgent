# Claude Provider

This package implements the LLM provider interface for Anthropic's Claude models.

## Overview

The Claude provider enables HelixAgent to communicate with Anthropic's Claude API, supporting all Claude 3 and Claude 4 model variants including Opus, Sonnet, and Haiku.

## Supported Models

| Model | Context | Description |
|-------|---------|-------------|
| claude-opus-4-5-20251101 | 200K | Most capable model |
| claude-sonnet-4-20250514 | 200K | Balanced performance |
| claude-3-opus-20240229 | 200K | Previous generation Opus |
| claude-3-sonnet-20240229 | 200K | Previous generation Sonnet |
| claude-3-haiku-20240307 | 200K | Fast and efficient |

## Authentication

### API Key (Recommended)
```bash
export CLAUDE_API_KEY="sk-ant-..."
```

### OAuth (Limited)
OAuth credentials from `claude auth login` are restricted to Claude Code CLI only and cannot be used for general API calls.

## Configuration

```yaml
providers:
  claude:
    enabled: true
    api_key: "${CLAUDE_API_KEY}"
    base_url: "https://api.anthropic.com"
    default_model: "claude-sonnet-4-20250514"
    max_retries: 3
    timeout_seconds: 120
```

## Features

- **Tool Calling**: Full support for function/tool calling
- **Vision**: Image analysis with Claude 3+ models
- **Streaming**: Real-time response streaming
- **System Prompts**: Configurable system messages

## Usage

```go
import "dev.helix.agent/internal/llm/providers/claude"

provider := claude.NewClaudeProvider(config)
response, err := provider.Complete(ctx, request)
```

## Rate Limits

| Tier | RPM | TPM |
|------|-----|-----|
| Free | 5 | 25K |
| Build | 50 | 100K |
| Scale | 1000 | 400K |

## Error Handling

The provider implements circuit breaker patterns for:
- Rate limit errors (429)
- Server errors (500, 502, 503)
- Timeout errors

## Testing

```bash
go test -v ./internal/llm/providers/claude/...
```

## Files

- `claude.go` - Main provider implementation
- `claude_test.go` - Unit tests
