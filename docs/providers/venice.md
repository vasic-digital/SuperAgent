# Venice AI Provider

## Overview

Venice AI is a privacy-first AI platform offering uncensored access to open-source models. It provides end-to-end encrypted inference with no data logging, making it suitable for sensitive workloads. Venice hosts a curated selection of high-capability models including Llama, DeepSeek, Qwen, and its own Venice Uncensored model.

## Authentication

| Header | Value |
|--------|-------|
| `Authorization` | `Bearer <VENICE_API_KEY>` |

### Obtaining an API Key

1. Visit [venice.ai](https://venice.ai) and create an account
2. Navigate to API settings in your dashboard
3. Generate a new API key
4. Store it securely

```bash
export VENICE_API_KEY=your_api_key_here
```

## API Base URL

```
https://api.venice.ai/api/v1
```

## Endpoints

| Endpoint | URL | Method |
|----------|-----|--------|
| Chat completions | `/chat/completions` | POST |
| Models list | `/models` | GET |
| Embeddings | `/embeddings` | POST |
| Image generation | `/image/generate` | POST |
| Image edit | `/image/edit` | POST |
| Image upscale | `/image/upscale` | POST |
| Text-to-speech | `/audio/speech` | POST |
| Speech-to-text | `/audio/transcriptions` | POST |

All endpoints are prefixed with `https://api.venice.ai/api/v1`.

## Supported Models

Models are discovered dynamically via 3-tier discovery (API, models.dev, fallback). Fallback models include:

| Model | Description |
|-------|-------------|
| `llama-3.3-70b` | Meta Llama 3.3 70B (default) |
| `zai-org-glm-4.7` | ZAI/Zhipu GLM-4.7 |
| `venice-uncensored` | Venice proprietary uncensored model |
| `qwen3-vl-235b-a22b` | Qwen 3 Vision-Language 235B |
| `qwen-2.5-vl` | Qwen 2.5 Vision-Language |
| `deepseek-r1-671b` | DeepSeek R1 671B reasoning model |
| `llama-3.1-405b` | Meta Llama 3.1 405B |

## Configuration

Add to your `.env` file:

```bash
# Required
VENICE_API_KEY=your_api_key_here

# Optional
VENICE_BASE_URL=https://api.venice.ai/api/v1/chat/completions
VENICE_MODEL=llama-3.3-70b
```

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `VENICE_API_KEY` | Yes | - | Your Venice API key |
| `VENICE_BASE_URL` | No | `https://api.venice.ai/api/v1/chat/completions` | Chat completions endpoint |
| `VENICE_MODEL` | No | `llama-3.3-70b` | Default model |

## Capabilities

- Chat completion: Yes
- Streaming: Yes (SSE)
- Tool/function calling: Yes
- Vision: Yes
- Embeddings: Yes (separate endpoint)
- Reasoning: Yes
- Web search: Yes (Venice-specific)
- Code completion: Yes
- Code analysis: Yes
- Uncensored mode: Yes

## Venice-Specific Features

### Web Search

Enable real-time web search to ground responses in current information:

```go
req.ModelParams.ProviderSpecific = map[string]interface{}{
    "enable_web_search":    "on",   // "on", "off", or "auto"
    "enable_web_citations": true,   // Include source citations
}
```

### Reasoning Effort

Control the depth of reasoning for supported models (e.g., DeepSeek R1):

```go
req.ModelParams.ProviderSpecific = map[string]interface{}{
    "reasoning": "high",  // "low", "medium", "high"
}
```

### Strip Thinking Response

Remove chain-of-thought `<think>` blocks from reasoning model output:

```go
req.ModelParams.ProviderSpecific = map[string]interface{}{
    "strip_thinking_response": true,
}
```

### venice_parameters in Request Body

Venice-specific parameters are sent as a `venice_parameters` object in the API request:

```json
{
  "model": "llama-3.3-70b",
  "messages": [{"role": "user", "content": "..."}],
  "venice_parameters": {
    "enable_web_search": "on",
    "enable_web_citations": true,
    "strip_thinking_response": false
  }
}
```

## Context and Token Limits

| Limit | Value |
|-------|-------|
| Max context window | 131,072 tokens |
| Max input length | 131,072 tokens |
| Max output length | 32,768 tokens |
| Max concurrent requests | 50 |
| Default max tokens | 4,096 |

## Example Usage

### Go Code Example

```go
package main

import (
    "context"
    "fmt"
    "os"

    "dev.helix.agent/internal/llm/providers/venice"
    "dev.helix.agent/internal/models"
)

func main() {
    provider := venice.NewProvider(
        os.Getenv("VENICE_API_KEY"),
        "", // Use default base URL
        "", // Use default model (llama-3.3-70b)
    )

    req := &models.LLMRequest{
        ID:     "request-1",
        Prompt: "You are a helpful AI assistant.",
        Messages: []models.Message{
            {Role: "user", Content: "Explain zero-knowledge proofs."},
        },
        ModelParams: models.ModelParams{
            MaxTokens:   2048,
            Temperature: 0.7,
        },
    }

    ctx := context.Background()
    resp, err := provider.Complete(ctx, req)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    fmt.Printf("Response: %s\n", resp.Content)
}
```

### Streaming Example

```go
streamChan, err := provider.CompleteStream(ctx, req)
if err != nil {
    fmt.Printf("Error: %v\n", err)
    return
}

for chunk := range streamChan {
    fmt.Print(chunk.Content)
}
```

### Web Search with Citations

```go
req := &models.LLMRequest{
    ID: "request-1",
    Messages: []models.Message{
        {Role: "user", Content: "What are the latest developments in quantum computing?"},
    },
    ModelParams: models.ModelParams{
        MaxTokens:   2048,
        Temperature: 0.7,
        ProviderSpecific: map[string]interface{}{
            "enable_web_search":    "on",
            "enable_web_citations": true,
        },
    },
}
```

## HelixAgent Integration

Venice participates in the HelixAgent debate system as a standard API key provider:

- **Provider type**: API Key (Bearer token)
- **Provider ID**: `venice`
- **Discovery**: 3-tier (Venice `/v1/models` API, models.dev, hardcoded fallback)
- **Health check**: Queries `/v1/models` endpoint (no token consumption)
- **Debate team selection**: Eligible based on LLMsVerifier scoring
- **Retry**: Exponential backoff with jitter (default: 3 retries, 1s initial, 30s max)

## Troubleshooting

### Authentication Error (401)

Verify your API key is correct and active. Check that the `Authorization: Bearer` header is set.

### Rate Limit Error (429)

HelixAgent automatically retries with exponential backoff. If persistent, check your Venice account quota.

### Model Not Found (404)

Verify the model ID matches one from the Venice models list. Model names do not use an `org/model` prefix (use `llama-3.3-70b`, not `meta/llama-3.3-70b`).

### Connection Issues

- Verify network connectivity to `api.venice.ai`
- Default HTTP timeout is 120 seconds
- Retryable status codes: 429, 500, 502, 503, 504

### Debug Logging

```bash
export GIN_MODE=debug
export LOG_LEVEL=debug
```

## Additional Resources

- [Venice AI](https://venice.ai)
- [Venice API Documentation](https://docs.venice.ai)
- Provider source: `internal/llm/providers/venice/venice.go`
