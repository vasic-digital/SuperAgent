# OpenRouter Provider

## Overview

OpenRouter is a universal API gateway providing access to 300+ AI models from multiple providers through a single OpenAI-compatible API. It offers unified billing, automatic model routing, and access to both premium and free-tier models, making it especially useful for HelixAgent's ensemble debate system.

## Authentication

| Header | Value |
|--------|-------|
| `Authorization` | `Bearer <OPENROUTER_API_KEY>` |

### Obtaining an API Key

1. Visit [openrouter.ai](https://openrouter.ai) and sign in
2. Navigate to **API Keys** in your dashboard
3. Click **Create Key**, name it (e.g., "HelixAgent")
4. Copy the key immediately (prefix: `sk-or-v1-`)

```bash
export OPENROUTER_API_KEY=sk-or-v1-xxxxxxxxxxxxxxxxxxxxxxxx
```

## API Base URL

```
https://openrouter.ai/api/v1
```

## Endpoint

| Endpoint | URL | Method |
|----------|-----|--------|
| Chat completions | `/chat/completions` | POST |
| Models list | `/models` | GET |

OpenRouter exposes a single OpenAI-compatible chat completions endpoint. All model access goes through `/chat/completions`.

## Model Naming Convention

Models use the `provider/model-name` format:

```
anthropic/claude-3.5-sonnet
openai/gpt-4o
google/gemini-pro
meta-llama/llama-3.1-405b-instruct
deepseek/deepseek-chat
```

### Free Models

Append `:free` to a model ID for free-tier access (rate-limited, no credits required):

```
meta-llama/llama-4-maverick:free
deepseek/deepseek-r1:free
google/gemini-2.5-pro-exp-03-25:free
qwen/qwq-32b:free
mistralai/mistral-7b-instruct:free
```

## Configuration

Add to your `.env` file:

```bash
# Required
OPENROUTER_API_KEY=sk-or-v1-xxxxxxxxxxxxxxxxxxxxxxxx

# Optional
OPENROUTER_BASE_URL=https://openrouter.ai/api/v1
OPENROUTER_DEFAULT_MODEL=anthropic/claude-3.5-sonnet
```

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `OPENROUTER_API_KEY` | Yes | - | Your OpenRouter API key |
| `OPENROUTER_BASE_URL` | No | `https://openrouter.ai/api/v1` | API base URL |
| `OPENROUTER_DEFAULT_MODEL` | No | - | Default model to use |

## Key Models by Provider

| Provider | Model ID | Notes |
|----------|----------|-------|
| Anthropic | `anthropic/claude-3.5-sonnet` | Premium |
| OpenAI | `openai/gpt-4o` | Premium |
| Google | `google/gemini-pro` | Premium |
| Meta | `meta-llama/llama-3.1-405b-instruct` | Premium |
| Meta | `meta-llama/llama-4-maverick:free` | Free |
| Meta | `meta-llama/llama-4-scout:free` | Free |
| DeepSeek | `deepseek/deepseek-chat` | Premium |
| DeepSeek | `deepseek/deepseek-r1:free` | Free |
| DeepSeek | `deepseek/deepseek-chat-v3-0324:free` | Free |
| Mistral | `mistralai/mistral-large` | Premium |
| Google | `google/gemini-2.5-pro-exp-03-25:free` | Free |
| Google | `google/gemma-3-27b-it:free` | Free |
| Qwen | `qwen/qwq-32b:free` | Free |
| NVIDIA | `nvidia/llama-3.1-nemotron-ultra-253b-v1:free` | Free |
| Microsoft | `microsoft/phi-3-medium-128k-instruct:free` | Free |

## App Attribution Headers

OpenRouter requires (or strongly recommends) attribution headers:

```go
httpReq.Header.Set("HTTP-Referer", "helixagent")
httpReq.Header.Set("X-Title", "HelixAgent")
```

These headers identify your application to OpenRouter and may be required for certain models. HelixAgent sets these automatically.

## Streaming

OpenRouter uses standard SSE (Server-Sent Events) for streaming:

- Each chunk is a `data: {json}\n\n` line
- Processing comments (lines starting with `:`) may appear mid-stream for keep-alive
- Errors can occur mid-stream as SSE events with error payloads
- Stream terminates with `data: [DONE]`

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

## Tool/Function Calling

OpenRouter passes tool definitions through to the underlying model. Not all models support tools. The request format is OpenAI-compatible:

```go
req := &models.LLMRequest{
    ID: "request-1",
    Messages: []models.Message{
        {Role: "user", Content: "What's the weather in London?"},
    },
    Tools: []models.Tool{
        {
            Type: "function",
            Function: models.ToolFunction{
                Name:        "get_weather",
                Description: "Get current weather",
                Parameters: map[string]interface{}{
                    "type": "object",
                    "properties": map[string]interface{}{
                        "city": map[string]interface{}{
                            "type": "string",
                        },
                    },
                    "required": []string{"city"},
                },
            },
        },
    },
    ModelParams: models.ModelParams{
        Model:     "openai/gpt-4o",
        MaxTokens: 1024,
    },
}
```

## Rate Limits

Rate limits vary by model and account tier:

| Tier | Requests/Minute | Concurrent Requests |
|------|-----------------|---------------------|
| Free | 20 | 5 |
| Standard | 200 | 50 |
| Pro | 1000 | 100 |

Free-tier models (`:free` suffix) have lower limits. Check the [OpenRouter Models page](https://openrouter.ai/models) for per-model specifics.

## Cost Tracking

OpenRouter charges per-token based on the underlying model's pricing. Costs vary significantly:

- Free models: $0
- Open-source models: fractions of a cent per 1K tokens
- Premium models (GPT-4o, Claude): standard provider pricing with OpenRouter markup

Monitor spending in the OpenRouter dashboard. HelixAgent extracts token usage from response metadata.

## HelixAgent Integration

OpenRouter is a key provider in the HelixAgent ecosystem:

- **Provider type**: API Key (Bearer token)
- **Provider ID**: `openrouter`
- **Discovery**: 3-tier (OpenRouter `/v1/models` API, models.dev, hardcoded fallback with 30+ models)
- **Health check**: Queries `/v1/models` endpoint
- **Free model discovery**: Automatically discovers `:free` models for zero-cost debate participants
- **Debate team selection**: Both premium and free models scored by LLMsVerifier
- **Auto-routing**: HelixAgent can route through OpenRouter as fallback when direct provider access fails
- **Retry**: Exponential backoff with jitter (default: 3 retries, 1s initial, 30s max)
- **Max tokens cap**: 16,384 (safe maximum across most OpenRouter models)
- **HTTP timeout**: 60 seconds

## Example Usage

### Go Code Example

```go
package main

import (
    "context"
    "fmt"
    "os"

    "dev.helix.agent/internal/llm/providers/openrouter"
    "dev.helix.agent/internal/models"
)

func main() {
    provider := openrouter.NewSimpleOpenRouterProvider(
        os.Getenv("OPENROUTER_API_KEY"),
    )

    req := &models.LLMRequest{
        ID: "request-1",
        Messages: []models.Message{
            {Role: "user", Content: "What is the capital of France?"},
        },
        ModelParams: models.ModelParams{
            Model:       "anthropic/claude-3.5-sonnet",
            MaxTokens:   1024,
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

### Using a Free Model

```go
req := &models.LLMRequest{
    ID: "request-1",
    Messages: []models.Message{
        {Role: "user", Content: "Explain quicksort."},
    },
    ModelParams: models.ModelParams{
        Model:       "deepseek/deepseek-r1:free",
        MaxTokens:   2048,
        Temperature: 0.3,
    },
}
```

## Troubleshooting

### Authentication Error (401)

Verify your API key starts with `sk-or-v1-` and has not been revoked.

### Insufficient Credits (402)

Add credits in the OpenRouter dashboard. Free-tier models (`:free`) do not require credits.

### Rate Limit Error (429)

HelixAgent retries automatically with exponential backoff. Upgrade your tier or switch to a less loaded model.

### Model Not Found (404)

Verify the model ID uses the correct `provider/model-name` format. Browse available models at [openrouter.ai/models](https://openrouter.ai/models).

### Provider Unavailable (503)

The underlying provider is temporarily down. Try an alternative model from a different provider. HelixAgent handles this via its fallback chain.

### Debug Logging

```bash
export GIN_MODE=debug
export LOG_LEVEL=debug
```

## Additional Resources

- [OpenRouter Website](https://openrouter.ai)
- [OpenRouter API Documentation](https://openrouter.ai/docs)
- [Model Browser](https://openrouter.ai/models)
- [Pricing](https://openrouter.ai/docs#pricing)
- Provider source: `internal/llm/providers/openrouter/openrouter.go`
