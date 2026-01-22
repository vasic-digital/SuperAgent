# Qwen Provider

This package implements the LLM provider interface for Alibaba's Qwen (Tongyi Qianwen) models.

## Overview

The Qwen provider enables HelixAgent to communicate with Alibaba's DashScope API, supporting Qwen's latest models including Qwen-Max and Qwen-Plus.

## Supported Models

| Model | Context | Description |
|-------|---------|-------------|
| qwen-max | 32K | Most capable model |
| qwen-max-latest | 32K | Latest version |
| qwen-plus | 128K | Balanced performance |
| qwen-plus-latest | 128K | Latest version |
| qwen-turbo | 128K | Fast and efficient |
| qwen-turbo-latest | 128K | Latest version |
| qwen2.5-coder-32b | 128K | Code-specialized |

## Authentication

### API Key (Recommended)
```bash
export QWEN_API_KEY="sk-..."
```

### OAuth (Limited)
OAuth credentials from Qwen CLI login are for the Qwen Portal only and cannot be used for DashScope API calls. You need a separate DashScope API key.

## Configuration

```yaml
providers:
  qwen:
    enabled: true
    api_key: "${QWEN_API_KEY}"
    base_url: "https://dashscope.aliyuncs.com/compatible-mode/v1"
    default_model: "qwen-plus"
    max_retries: 3
    timeout_seconds: 60
```

## Features

- **Tool Calling**: Full function calling support
- **Streaming**: Real-time response streaming
- **Long Context**: Up to 128K tokens
- **Multi-language**: Strong multilingual support

## Usage

```go
import "dev.helix.agent/internal/llm/providers/qwen"

provider := qwen.NewQwenProvider(config)
response, err := provider.Complete(ctx, request)
```

## Rate Limits

Rate limits depend on your DashScope subscription tier.

## Error Handling

Standard error handling with retries for:
- Rate limits
- Server errors
- Network timeouts

## Testing

```bash
go test -v ./internal/llm/providers/qwen/...
```

## Files

- `qwen.go` - Main provider implementation
- `qwen_test.go` - Unit tests
