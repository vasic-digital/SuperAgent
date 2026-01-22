# OpenRouter Provider

This package implements the LLM provider interface for OpenRouter, a unified gateway to multiple LLM providers.

## Overview

The OpenRouter provider enables HelixAgent to access models from multiple providers (OpenAI, Anthropic, Google, Meta, etc.) through a single API, with automatic failover and cost optimization.

## Supported Models

OpenRouter provides access to 100+ models including:

| Model | Provider | Description |
|-------|----------|-------------|
| openai/gpt-4o | OpenAI | GPT-4o latest |
| anthropic/claude-3.5-sonnet | Anthropic | Claude 3.5 Sonnet |
| google/gemini-pro-1.5 | Google | Gemini Pro 1.5 |
| meta-llama/llama-3.1-405b | Meta | Llama 3.1 405B |
| deepseek/deepseek-r1 | DeepSeek | DeepSeek R1 |

### Free Models

Models with `:free` suffix are available without API key charges:
- `google/gemini-2.0-flash-exp:free`
- `meta-llama/llama-3.2-3b-instruct:free`
- `mistralai/mistral-7b-instruct:free`

## Authentication

```bash
export OPENROUTER_API_KEY="sk-or-..."
```

## Configuration

```yaml
providers:
  openrouter:
    enabled: true
    api_key: "${OPENROUTER_API_KEY}"
    base_url: "https://openrouter.ai/api/v1"
    default_model: "google/gemini-2.0-flash-exp:free"
    max_retries: 3
    timeout_seconds: 120
```

## Features

- **Multi-Provider Access**: Single API for all major providers
- **Automatic Failover**: Routes to available providers
- **Cost Tracking**: Real-time cost monitoring
- **Free Tier**: Access to free models
- **Tool Calling**: Support for function calling
- **Streaming**: Real-time response streaming

## Usage

```go
import "dev.helix.agent/internal/llm/providers/openrouter"

provider := openrouter.NewOpenRouterProvider(config)
response, err := provider.Complete(ctx, request)
```

## Rate Limits

Rate limits vary by model and your account tier. Free models have lower limits.

## Error Handling

The provider handles:
- Provider failover (automatically switches providers)
- Rate limits (queues requests)
- Model availability (selects alternatives)

## Testing

```bash
go test -v ./internal/llm/providers/openrouter/...
```

## Files

- `openrouter.go` - Main provider implementation
- `openrouter_test.go` - Unit tests
