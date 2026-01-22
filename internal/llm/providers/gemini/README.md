# Gemini Provider

This package implements the LLM provider interface for Google's Gemini models.

## Overview

The Gemini provider enables HelixAgent to communicate with Google's Generative AI API, supporting Gemini Pro, Gemini Flash, and experimental models.

## Supported Models

| Model | Context | Description |
|-------|---------|-------------|
| gemini-2.0-flash-exp | 1M | Latest experimental model |
| gemini-1.5-pro | 2M | Most capable production model |
| gemini-1.5-flash | 1M | Fast and efficient |
| gemini-1.5-flash-8b | 1M | Lightweight variant |

## Authentication

```bash
export GEMINI_API_KEY="AIza..."
```

## Configuration

```yaml
providers:
  gemini:
    enabled: true
    api_key: "${GEMINI_API_KEY}"
    base_url: "https://generativelanguage.googleapis.com"
    default_model: "gemini-1.5-flash"
    max_retries: 3
    timeout_seconds: 60
```

## Features

- **Tool Calling**: Full function calling support
- **Vision**: Multi-modal image and video support
- **Long Context**: Up to 2M tokens
- **Streaming**: Real-time response streaming
- **Grounding**: Google Search grounding capability

## Usage

```go
import "dev.helix.agent/internal/llm/providers/gemini"

provider := gemini.NewGeminiProvider(config)
response, err := provider.Complete(ctx, request)
```

## Rate Limits

| Model | Free RPM | Paid RPM |
|-------|----------|----------|
| gemini-1.5-flash | 15 | 1000 |
| gemini-1.5-pro | 2 | 360 |

## Error Handling

Implements retry logic for:
- RESOURCE_EXHAUSTED (rate limits)
- UNAVAILABLE (server errors)
- DEADLINE_EXCEEDED (timeouts)

## Testing

```bash
go test -v ./internal/llm/providers/gemini/...
```

## Files

- `gemini.go` - Main provider implementation
- `gemini_test.go` - Unit tests
