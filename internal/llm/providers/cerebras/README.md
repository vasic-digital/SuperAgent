# Cerebras Provider

This package implements the LLM provider interface for Cerebras AI models.

## Overview

The Cerebras provider enables HelixAgent to communicate with Cerebras Cloud API, leveraging their specialized AI accelerator hardware for high-performance inference.

## Supported Models

| Model | Context | Description |
|-------|---------|-------------|
| llama3.1-8b | 8K | Llama 3.1 8B on Cerebras |
| llama3.1-70b | 8K | Llama 3.1 70B on Cerebras |

## Authentication

```bash
export CEREBRAS_API_KEY="csk-..."
```

## Configuration

```yaml
providers:
  cerebras:
    enabled: true
    api_key: "${CEREBRAS_API_KEY}"
    base_url: "https://api.cerebras.ai/v1"
    default_model: "llama3.1-8b"
    max_retries: 3
    timeout_seconds: 60
```

## Features

- **High Performance**: Optimized for Cerebras hardware
- **OpenAI Compatible**: Uses OpenAI-compatible API format
- **Streaming**: Real-time response streaming
- **Low Latency**: Fast inference times

## Usage

```go
import "dev.helix.agent/internal/llm/providers/cerebras"

provider := cerebras.NewCerebrasProvider(config)
response, err := provider.Complete(ctx, request)
```

## Performance

Cerebras specializes in high-throughput inference:
- Extremely fast token generation
- Consistent low latency
- Efficient for batch processing

## Rate Limits

Rate limits depend on your Cerebras subscription tier.

## Error Handling

Standard error handling with retries for:
- Rate limits (429)
- Server errors (5xx)
- Network timeouts

## Testing

```bash
go test -v ./internal/llm/providers/cerebras/...
```

## Files

- `cerebras.go` - Main provider implementation
- `cerebras_test.go` - Unit tests
