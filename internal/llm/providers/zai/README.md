# ZAI Provider

This package implements the LLM provider interface for ZAI (Z-AI) models.

## Overview

The ZAI provider enables HelixAgent to communicate with ZAI's API, supporting their AI models for various tasks.

## Supported Models

| Model | Context | Description |
|-------|---------|-------------|
| zai-default | 32K | Default general model |

## Authentication

```bash
export ZAI_API_KEY="..."
```

## Configuration

```yaml
providers:
  zai:
    enabled: true
    api_key: "${ZAI_API_KEY}"
    base_url: "https://api.zai.ai/v1"
    default_model: "zai-default"
    max_retries: 3
    timeout_seconds: 60
```

## Features

- **OpenAI Compatible**: Uses OpenAI-compatible API format
- **Streaming**: Real-time response streaming

## Usage

```go
import "dev.helix.agent/internal/llm/providers/zai"

provider := zai.NewZAIProvider(config)
response, err := provider.Complete(ctx, request)
```

## Rate Limits

Rate limits depend on your subscription tier.

## Error Handling

Standard error handling with retries for:
- Rate limits
- Server errors
- Network timeouts

## Testing

```bash
go test -v ./internal/llm/providers/zai/...
```

## Files

- `zai.go` - Main provider implementation
- `zai_test.go` - Unit tests
