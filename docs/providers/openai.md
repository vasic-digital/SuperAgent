# OpenAI Provider

## Overview

OpenAI is the creator of the GPT series of large language models, including GPT-4o, GPT-4 Turbo, and the o-series reasoning models. HelixAgent integrates with OpenAI's Chat Completions API to provide access to their full range of models with support for chat, streaming, tool calling, vision, and JSON mode.

## Authentication

OpenAI uses Bearer token authentication via the `Authorization` header. An optional `OpenAI-Organization` header can be set for organization-scoped requests.

| Header | Format | Required |
|--------|--------|----------|
| `Authorization` | `Bearer <api_key>` | Yes |
| `OpenAI-Organization` | `<org_id>` | No |

### Environment Variable

```bash
OPENAI_API_KEY=sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

## Configuration

Add the following to your `.env` file or environment:

```bash
# Required
OPENAI_API_KEY=sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

# Optional - Override default settings
OPENAI_BASE_URL=https://api.openai.com/v1/chat/completions
OPENAI_MODEL=gpt-4o
```

### Configuration Options

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `OPENAI_API_KEY` | Yes | - | Your OpenAI API key |
| `OPENAI_BASE_URL` | No | `https://api.openai.com/v1/chat/completions` | API endpoint URL |
| `OPENAI_MODEL` | No | `gpt-4o` | Default model to use |

## Supported Models

Models are discovered dynamically via the OpenAI `/v1/models` API endpoint, with the following fallback list:

- `gpt-4o` (default) - Latest multimodal flagship model
- `gpt-4o-mini` - Smaller, cost-effective variant
- `gpt-4-turbo` - GPT-4 Turbo with vision
- `gpt-4` - Original GPT-4
- `gpt-3.5-turbo` - Fast and cost-effective
- `o1` - Reasoning model
- `o1-mini` - Smaller reasoning model
- `o3` - Advanced reasoning model
- `o3-mini` - Smaller advanced reasoning model

## Capabilities

- Chat completion: Yes
- Streaming: Yes
- Tool/function calling: Yes
- Vision: Yes
- Embeddings: Yes (via separate endpoint)
- JSON mode: Yes
- Reasoning: Yes
- Code completion: Yes
- Code analysis: Yes

### Model Limits

| Parameter | Value |
|-----------|-------|
| Max tokens (context) | 128,000 |
| Max input length | 128,000 |
| Max output length | 16,384 |
| Max concurrent requests | 100 |
| HTTP client timeout | 120 seconds |

## API Endpoint

- **Chat Completions**: `https://api.openai.com/v1/chat/completions`
- **Models list**: `https://api.openai.com/v1/models`

## Rate Limits

Rate limits vary by model and subscription tier:

| Tier | RPM | TPM |
|------|-----|-----|
| Free | 3 | 40,000 |
| Tier 1 | 500 | 200,000 |
| Tier 2 | 5,000 | 2,000,000 |
| Tier 3+ | 10,000+ | Custom |

HelixAgent automatically implements retry with exponential backoff (initial delay 1s, max delay 30s, multiplier 2.0) for rate limit (429) and server error (5xx) responses.

## Known Limitations

- Vision support depends on the specific model selected (GPT-4o, GPT-4 Turbo).
- The o-series reasoning models do not support streaming in the same way as GPT models.
- Token output limits vary by model (e.g., `gpt-4o` supports up to 16,384 output tokens, while `gpt-4` supports 8,192).
- Organization-scoped requests require a valid `OpenAI-Organization` header.

## Example Usage

### Go Code Example

```go
package main

import (
    "context"
    "fmt"
    "os"

    "dev.helix.agent/internal/llm/providers/openai"
    "dev.helix.agent/internal/models"
)

func main() {
    provider := openai.NewProvider(
        os.Getenv("OPENAI_API_KEY"),
        "", // Use default base URL
        "", // Use default model (gpt-4o)
    )

    req := &models.LLMRequest{
        ID:     "request-1",
        Prompt: "You are a helpful assistant.",
        Messages: []models.Message{
            {Role: "user", Content: "Explain the difference between goroutines and threads."},
        },
        ModelParams: models.ModelParams{
            MaxTokens:   4096,
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

### Tool Calling Example

```go
req := &models.LLMRequest{
    ID:     "request-1",
    Prompt: "You are a helpful assistant with access to tools.",
    Messages: []models.Message{
        {Role: "user", Content: "What is the weather in San Francisco?"},
    },
    Tools: []models.Tool{
        {
            Type: "function",
            Function: models.ToolFunction{
                Name:        "get_weather",
                Description: "Get the current weather for a location",
                Parameters: map[string]interface{}{
                    "type": "object",
                    "properties": map[string]interface{}{
                        "location": map[string]interface{}{
                            "type":        "string",
                            "description": "City name",
                        },
                    },
                    "required": []string{"location"},
                },
            },
        },
    },
    ModelParams: models.ModelParams{
        MaxTokens:   1024,
        Temperature: 0.3,
    },
}
```

### Custom Retry Configuration

```go
retryConfig := openai.RetryConfig{
    MaxRetries:   5,
    InitialDelay: 2 * time.Second,
    MaxDelay:     60 * time.Second,
    Multiplier:   2.0,
}

provider := openai.NewProviderWithRetry(
    apiKey, "", "", retryConfig,
)
```

## Troubleshooting

### Authentication Error (401)

Verify your API key is correct and has not been revoked. Ensure the `Authorization: Bearer` header is properly formatted.

### Rate Limit Error (429)

HelixAgent automatically retries with exponential backoff. If persistent, consider upgrading your OpenAI plan or reducing request frequency.

### Content Filter

If responses are blocked by the content filter, the confidence score is reduced by 0.3. Rephrase prompts to avoid triggering content moderation.

## Additional Resources

- [OpenAI API Documentation](https://platform.openai.com/docs/api-reference)
- [OpenAI Platform](https://platform.openai.com)
- [Model Overview](https://platform.openai.com/docs/models)
