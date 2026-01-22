# DeepSeek Provider

This package implements the LLM provider interface for DeepSeek AI models.

## Overview

The DeepSeek provider enables HelixAgent to communicate with DeepSeek's API, supporting their latest models including DeepSeek-V3 and DeepSeek-R1.

## Supported Models

| Model | Context | Description |
|-------|---------|-------------|
| deepseek-chat | 64K | General conversation model |
| deepseek-coder | 64K | Code-specialized model |
| deepseek-reasoner | 64K | Reasoning-focused model (R1) |

## Authentication

```bash
export DEEPSEEK_API_KEY="sk-..."
```

## Configuration

```yaml
providers:
  deepseek:
    enabled: true
    api_key: "${DEEPSEEK_API_KEY}"
    base_url: "https://api.deepseek.com"
    default_model: "deepseek-chat"
    max_retries: 3
    timeout_seconds: 60
```

## Features

- **Tool Calling**: Full function calling support
- **Streaming**: Real-time response streaming
- **Cost Effective**: Competitive pricing
- **OpenAI Compatible**: Uses OpenAI-compatible API format

## Usage

```go
import "dev.helix.agent/internal/llm/providers/deepseek"

provider := deepseek.NewDeepSeekProvider(config)
response, err := provider.Complete(ctx, request)
```

## Rate Limits

DeepSeek has generous rate limits with pay-as-you-go pricing.

## Error Handling

Standard error handling with retries for:
- Network errors
- Rate limits
- Server errors

## Testing

```bash
go test -v ./internal/llm/providers/deepseek/...
```

## Files

- `deepseek.go` - Main provider implementation
- `deepseek_test.go` - Unit tests
