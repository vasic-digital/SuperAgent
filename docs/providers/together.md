# Together AI Provider

## Overview

Together AI is a cloud platform for running and fine-tuning open-source AI models. HelixAgent integrates with Together AI's OpenAI-compatible Chat Completions API, providing access to a broad catalog of open-source models from Meta, Mistral, DeepSeek, Google, Qwen, NVIDIA, and Databricks with support for chat, streaming, tool calling, vision, and JSON mode.

## Authentication

Together AI uses Bearer token authentication via the `Authorization` header.

| Header | Format | Required |
|--------|--------|----------|
| `Authorization` | `Bearer <api_key>` | Yes |

### Environment Variable

```bash
TOGETHER_API_KEY=your-api-key-here
```

## Configuration

Add the following to your `.env` file or environment:

```bash
# Required
TOGETHER_API_KEY=your-api-key-here

# Optional - Override default settings
TOGETHER_BASE_URL=https://api.together.xyz/v1/chat/completions
TOGETHER_MODEL=meta-llama/Llama-3.3-70B-Instruct-Turbo
```

### Configuration Options

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `TOGETHER_API_KEY` | Yes | - | Your Together AI API key |
| `TOGETHER_BASE_URL` | No | `https://api.together.xyz/v1/chat/completions` | API endpoint URL |
| `TOGETHER_MODEL` | No | `meta-llama/Llama-3.3-70B-Instruct-Turbo` | Default model to use |

## Supported Models

Models are discovered dynamically via the Together AI `/v1/models` API endpoint, with the following fallback list:

### Meta Llama Models
- `meta-llama/Llama-3.3-70B-Instruct-Turbo` (default)
- `meta-llama/Llama-3.2-90B-Vision-Instruct-Turbo`
- `meta-llama/Llama-3.2-11B-Vision-Instruct-Turbo`
- `meta-llama/Meta-Llama-3.1-405B-Instruct-Turbo`
- `meta-llama/Meta-Llama-3.1-70B-Instruct-Turbo`
- `meta-llama/Meta-Llama-3.1-8B-Instruct-Turbo`

### Qwen Models
- `Qwen/Qwen2.5-72B-Instruct-Turbo`
- `Qwen/Qwen2.5-7B-Instruct-Turbo`
- `Qwen/QwQ-32B-Preview`

### Mistral Models
- `mistralai/Mistral-Small-24B-Instruct-2501`
- `mistralai/Mixtral-8x22B-Instruct-v0.1`
- `mistralai/Mixtral-8x7B-Instruct-v0.1`

### DeepSeek Models
- `deepseek-ai/DeepSeek-R1`
- `deepseek-ai/DeepSeek-R1-Distill-Llama-70B`
- `deepseek-ai/DeepSeek-V3`

### Google Models
- `google/gemma-2-27b-it`
- `google/gemma-2-9b-it`

### NVIDIA Models
- `nvidia/Llama-3.1-Nemotron-70B-Instruct-HF`

### Databricks Models
- `databricks/dbrx-instruct`

## Capabilities

- Chat completion: Yes
- Streaming: Yes
- Tool/function calling: Yes
- Vision: Yes (via Llama 3.2 Vision models)
- Embeddings: Yes (via separate endpoint)
- JSON mode: Yes
- Code completion: Yes
- Reasoning: Yes
- Code analysis: Yes

### Model Limits

| Parameter | Value |
|-----------|-------|
| Max tokens (context) | 131,072 |
| Max input length | 131,072 |
| Max output length | 4,096 |
| Max concurrent requests | 100 |
| HTTP client timeout | 120 seconds |

## API Endpoint

- **Chat Completions**: `https://api.together.xyz/v1/chat/completions`
- **Models list**: `https://api.together.xyz/v1/models`

Together AI uses an OpenAI-compatible API format, which means requests and responses follow the same structure as OpenAI's Chat Completions API.

## Rate Limits

Rate limits depend on your Together AI plan:

| Plan | RPM | TPM |
|------|-----|-----|
| Free | 60 | 60,000 |
| Starter | 600 | 600,000 |
| Pro | 6,000 | 6,000,000 |
| Enterprise | Custom | Custom |

HelixAgent automatically implements retry with exponential backoff (initial delay 1s, max delay 30s, multiplier 2.0) for rate limit (429) and server error (5xx) responses.

## Known Limitations

- Vision support depends on the specific model selected (e.g., Llama 3.2 Vision models).
- Not all models support tool/function calling; availability is model-dependent.
- Context window sizes vary by model; the 131,072 limit is the maximum across all models.
- The `repetition_penalty` parameter is available but not mapped from the standard request parameters by default.
- Response format (`json_object` mode) requires model support and may not work with all models.

## Example Usage

### Go Code Example

```go
package main

import (
    "context"
    "fmt"
    "os"

    "dev.helix.agent/internal/llm/providers/together"
    "dev.helix.agent/internal/models"
)

func main() {
    provider := together.NewProvider(
        os.Getenv("TOGETHER_API_KEY"),
        "", // Use default base URL
        "", // Use default model (Llama-3.3-70B-Instruct-Turbo)
    )

    req := &models.LLMRequest{
        ID:     "request-1",
        Prompt: "You are a helpful assistant.",
        Messages: []models.Message{
            {Role: "user", Content: "Explain the difference between concurrency and parallelism."},
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
        {Role: "user", Content: "What is the current time in Tokyo?"},
    },
    Tools: []models.Tool{
        {
            Type: "function",
            Function: models.ToolFunction{
                Name:        "get_current_time",
                Description: "Get the current time for a timezone",
                Parameters: map[string]interface{}{
                    "type": "object",
                    "properties": map[string]interface{}{
                        "timezone": map[string]interface{}{
                            "type":        "string",
                            "description": "IANA timezone name",
                        },
                    },
                    "required": []string{"timezone"},
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

### Using a DeepSeek Model

```go
provider := together.NewProvider(
    os.Getenv("TOGETHER_API_KEY"),
    "",
    "deepseek-ai/DeepSeek-R1",
)
```

### Custom Retry Configuration

```go
retryConfig := together.RetryConfig{
    MaxRetries:   5,
    InitialDelay: 2 * time.Second,
    MaxDelay:     60 * time.Second,
    Multiplier:   2.0,
}

provider := together.NewProviderWithRetry(
    apiKey, "", "", retryConfig,
)
```

## Troubleshooting

### Authentication Error (401)

Verify your Together AI API key is correct. Keys can be generated at [api.together.xyz/settings/api-keys](https://api.together.xyz/settings/api-keys).

### Rate Limit Error (429)

HelixAgent automatically retries with exponential backoff. Consider upgrading your plan for higher rate limits.

### Model Not Available

Not all models listed in the catalog are always available. Check model status on the Together AI platform dashboard.

### Content Filter

If responses are blocked by a content filter, the confidence score is reduced by 0.3. Rephrase prompts to avoid triggering content moderation.

## Additional Resources

- [Together AI API Documentation](https://docs.together.ai)
- [Together AI Platform](https://api.together.xyz)
- [Together AI Model Library](https://api.together.xyz/models)
