# Mistral Provider

This package implements the LLM provider interface for Mistral AI models.

## Overview

The Mistral provider enables HelixAgent to communicate with Mistral's API, supporting their range of models from Mistral 7B to Mistral Large.

## Supported Models

| Model | Context | Description |
|-------|---------|-------------|
| mistral-large-latest | 128K | Most capable model |
| mistral-medium-latest | 32K | Balanced performance |
| mistral-small-latest | 32K | Cost-effective option |
| codestral-latest | 32K | Code-specialized model |
| mistral-nemo | 128K | Efficient 12B model |

## Authentication

```bash
export MISTRAL_API_KEY="..."
```

## Configuration

```yaml
providers:
  mistral:
    enabled: true
    api_key: "${MISTRAL_API_KEY}"
    base_url: "https://api.mistral.ai"
    default_model: "mistral-small-latest"
    max_retries: 3
    timeout_seconds: 60
```

## Features

- **Tool Calling**: Full function calling support
- **Streaming**: Real-time response streaming
- **JSON Mode**: Structured output generation
- **Fine-tuning**: Custom model training support

## Usage

```go
import "dev.helix.agent/internal/llm/providers/mistral"

provider := mistral.NewMistralProvider(config)
response, err := provider.Complete(ctx, request)
```

## Rate Limits

Rate limits vary by subscription tier and model.

## Error Handling

Standard error handling with retries for:
- Rate limit errors (429)
- Server errors (5xx)
- Network timeouts

## Testing

```bash
go test -v ./internal/llm/providers/mistral/...
```

## Files

- `mistral.go` - Main provider implementation
- `mistral_test.go` - Unit tests
