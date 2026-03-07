# HuggingFace Provider

## Overview

HuggingFace is the largest open-source AI model hub, hosting thousands of models from Meta, Google, Microsoft, Mistral, and the broader community. HelixAgent integrates with HuggingFace's Inference Router API, supporting both the OpenAI-compatible chat completions endpoint (Pro mode) and the legacy Inference API for direct model access.

## Authentication

HuggingFace uses Bearer token authentication via the `Authorization` header.

| Header | Format | Required |
|--------|--------|----------|
| `Authorization` | `Bearer <api_key>` | Yes |

### Environment Variable

```bash
HUGGINGFACE_API_KEY=hf_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

## Configuration

Add the following to your `.env` file or environment:

```bash
# Required
HUGGINGFACE_API_KEY=hf_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

# Optional - Override default settings
HUGGINGFACE_BASE_URL=https://router.huggingface.co/v1/chat/completions
HUGGINGFACE_MODEL=meta-llama/Llama-3.3-70B-Instruct
```

### Configuration Options

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `HUGGINGFACE_API_KEY` | Yes | - | Your HuggingFace API token |
| `HUGGINGFACE_BASE_URL` | No | `https://router.huggingface.co/v1/chat/completions` | API endpoint URL |
| `HUGGINGFACE_MODEL` | No | `meta-llama/Llama-3.3-70B-Instruct` | Default model to use |

### API Modes

The provider automatically selects the API mode based on the base URL:

- **Pro mode** (default): Uses the OpenAI-compatible chat completions endpoint at `router.huggingface.co/v1/chat/completions`. Supports streaming natively.
- **Inference mode**: Uses the legacy inference endpoint at `router.huggingface.co/hf-inference/models/`. Streaming falls back to polling. Activated when the base URL does not contain `chat/completions`.

## Supported Models

Models are discovered via the models.dev catalog, with the following fallback list:

### Meta Llama Models
- `meta-llama/Llama-3.3-70B-Instruct` (default)
- `meta-llama/Llama-3.2-3B-Instruct`
- `meta-llama/Llama-3.1-8B-Instruct`
- `meta-llama/Llama-3.1-70B-Instruct`

### Mistral Models
- `mistralai/Mistral-7B-Instruct-v0.3`
- `mistralai/Mixtral-8x7B-Instruct-v0.1`

### Google Models
- `google/gemma-2-9b-it`
- `google/gemma-2-27b-it`

### Microsoft Models
- `microsoft/Phi-3-mini-4k-instruct`
- `microsoft/Phi-3-medium-4k-instruct`

### Other Models
- `Qwen/Qwen2.5-72B-Instruct`
- `bigcode/starcoder2-15b`

## Capabilities

- Chat completion: Yes
- Streaming: Yes (native in Pro mode; polling fallback in Inference mode)
- Tool/function calling: No
- Vision: Yes (model-dependent)
- Embeddings: Yes (via separate models)
- Text generation: Yes
- Code completion: Yes
- Classification: Yes
- Translation: Yes
- Reasoning: Yes
- Code analysis: Yes

### Model Limits

| Parameter | Value |
|-----------|-------|
| Max tokens (context) | 8,192 |
| Max input length | 8,192 |
| Max output length | 4,096 |
| Max concurrent requests | 50 |
| HTTP client timeout | 120 seconds |

Note: Actual context limits vary significantly by model. The Llama 3 models support up to 128K tokens. The limits above represent the provider-level defaults.

## API Endpoint

- **Chat Completions (Pro)**: `https://router.huggingface.co/v1/chat/completions`
- **Inference (Legacy)**: `https://router.huggingface.co/hf-inference/models/<model_id>`
- **Health Check**: `https://huggingface.co/api/models/<model_id>`

Note: The previous endpoint `api-inference.huggingface.co` returns 410 Gone as of November 2025 and is no longer used.

## Rate Limits

Rate limits depend on your HuggingFace subscription:

| Plan | RPM | Notes |
|------|-----|-------|
| Free | 30 | Shared infrastructure, possible cold starts |
| Pro | 1,000 | Priority access, faster cold starts |
| Enterprise | Custom | Dedicated infrastructure |

HelixAgent automatically implements retry with exponential backoff (initial delay 1s, max delay 30s, multiplier 2.0) for rate limit (429), server error (5xx), and service unavailable (503) responses. The 503 status is specifically handled because HuggingFace returns it during model cold starts.

## Known Limitations

- Tool/function calling is not supported.
- Model availability depends on the HuggingFace hosting infrastructure; some models may have cold start delays.
- The legacy Inference API does not support streaming -- the provider falls back to a single polling response.
- Max output tokens default to 1,024 in Pro mode and 512 in Inference mode (lower than other providers).
- The Inference API option `wait_for_model: true` is enabled by default to handle cold starts gracefully.
- Context window sizes are model-dependent and can range from 4K to 128K tokens.

## Example Usage

### Go Code Example

```go
package main

import (
    "context"
    "fmt"
    "os"

    "dev.helix.agent/internal/llm/providers/huggingface"
    "dev.helix.agent/internal/models"
)

func main() {
    provider := huggingface.NewProvider(
        os.Getenv("HUGGINGFACE_API_KEY"),
        "", // Use default base URL (Pro mode)
        "", // Use default model (Llama-3.3-70B-Instruct)
    )

    req := &models.LLMRequest{
        ID:     "request-1",
        Prompt: "You are a helpful coding assistant.",
        Messages: []models.Message{
            {Role: "user", Content: "Explain how goroutines work in Go."},
        },
        ModelParams: models.ModelParams{
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

### Using a Specific Model

```go
provider := huggingface.NewProvider(
    os.Getenv("HUGGINGFACE_API_KEY"),
    "",
    "google/gemma-2-27b-it",
)
```

### Custom Retry Configuration

```go
retryConfig := huggingface.RetryConfig{
    MaxRetries:   5,
    InitialDelay: 2 * time.Second,
    MaxDelay:     60 * time.Second,
    Multiplier:   2.0,
}

provider := huggingface.NewProviderWithRetry(
    apiKey, "", "", retryConfig,
)
```

## Troubleshooting

### Authentication Error (401)

Verify your HuggingFace token is correct. Tokens can be generated at [huggingface.co/settings/tokens](https://huggingface.co/settings/tokens).

### Model Loading (503)

HuggingFace returns 503 when a model is loading (cold start). The provider automatically retries. Ensure `WaitForModel` is enabled (default behavior).

### Rate Limit Error (429)

HelixAgent automatically retries with exponential backoff. Consider upgrading to HuggingFace Pro for higher rate limits.

### Endpoint Deprecation

If you see 410 Gone errors, ensure you are not using the deprecated `api-inference.huggingface.co` endpoint. The provider defaults to `router.huggingface.co`.

## Additional Resources

- [HuggingFace Inference API Documentation](https://huggingface.co/docs/api-inference)
- [HuggingFace Model Hub](https://huggingface.co/models)
- [HuggingFace Pro Inference](https://huggingface.co/docs/inference-endpoints)
