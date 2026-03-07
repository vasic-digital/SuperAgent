# Replicate Provider

## Overview

Replicate is a platform for running open-source machine learning models in the cloud. HelixAgent integrates with Replicate's Predictions API, providing access to a wide range of open-source LLMs, image generation models, speech-to-text models, and more. Replicate uses an asynchronous prediction model where requests are submitted and then polled for completion.

## Authentication

Replicate uses Bearer token authentication via the `Authorization` header.

| Header | Format | Required |
|--------|--------|----------|
| `Authorization` | `Bearer <api_key>` | Yes |

### Environment Variable

```bash
REPLICATE_API_KEY=r8_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

## Configuration

Add the following to your `.env` file or environment:

```bash
# Required
REPLICATE_API_KEY=r8_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

# Optional - Override default settings
REPLICATE_BASE_URL=https://api.replicate.com/v1/predictions
REPLICATE_MODEL=meta/llama-2-70b-chat
```

### Configuration Options

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `REPLICATE_API_KEY` | Yes | - | Your Replicate API token |
| `REPLICATE_BASE_URL` | No | `https://api.replicate.com/v1/predictions` | API endpoint URL |
| `REPLICATE_MODEL` | No | `meta/llama-2-70b-chat` | Default model to use |

## Supported Models

Models are discovered dynamically via the Replicate `/v1/models` API endpoint (using a custom response parser), with the following fallback list:

### Meta Llama Models
- `meta/llama-2-70b-chat` (default)
- `meta/llama-2-13b-chat`
- `meta/llama-2-7b-chat`
- `meta/meta-llama-3-70b-instruct`
- `meta/meta-llama-3-8b-instruct`
- `meta/meta-llama-3.1-405b-instruct`

### Mistral Models
- `mistralai/mistral-7b-instruct-v0.2`
- `mistralai/mixtral-8x7b-instruct-v0.1`

### Other Models
- `stability-ai/stable-diffusion` - Image generation
- `openai/whisper` - Speech-to-text
- `replicate/all-mpnet-base-v2` - Embeddings

## Capabilities

- Chat completion: Yes
- Streaming: Yes (via polling-based streaming)
- Tool/function calling: No
- Vision: Yes (model-dependent)
- Embeddings: Yes (via embedding models)
- Image generation: Yes (via Stable Diffusion and similar)
- Speech-to-text: Yes (via Whisper)
- Reasoning: Yes
- Code completion: Yes
- Code analysis: Yes

### Model Limits

| Parameter | Value |
|-----------|-------|
| Max tokens | 4,096 |
| Max input length | 4,096 |
| Max output length | 4,096 |
| Max concurrent requests | 50 |
| HTTP client timeout | 300 seconds |

The HTTP client timeout is set to 300 seconds (5 minutes), significantly higher than other providers, to accommodate Replicate's cold start times.

## API Endpoint

- **Predictions**: `https://api.replicate.com/v1/predictions`
- **Models**: `https://api.replicate.com/v1/models`

## Rate Limits

Replicate uses a pay-per-use model with rate limits based on your plan:

| Plan | Concurrent Predictions | Notes |
|------|----------------------|-------|
| Free | 1 | Community GPU access |
| Standard | 10 | Priority GPU access |
| Enterprise | Custom | Dedicated GPUs |

HelixAgent automatically implements retry with exponential backoff (initial delay 1s, max delay 30s, multiplier 2.0) for rate limit (429) and server error (5xx) responses.

## Known Limitations

- **Asynchronous execution model**: Replicate uses a prediction-based workflow. Requests are submitted and return immediately with a prediction ID and status URLs. The provider polls for completion rather than receiving a synchronous response.
- **Cold starts**: Models that are not frequently used may require cold start time (30 seconds to several minutes). The 300-second HTTP timeout accommodates this.
- **No tool/function calling**: Replicate does not support tool calling or function calling.
- **Polling-based streaming**: Unlike providers that use SSE, the Replicate provider implements streaming by polling the prediction URL every 500ms and emitting deltas as new output tokens are detected.
- **Prompt format**: Messages are formatted using the Llama instruction format (`[INST]...[/INST]`) rather than the OpenAI-compatible chat format.
- **Default max tokens is low**: The default max output tokens is 512, lower than most other providers.
- **Output format varies**: The prediction output can be a string or an array of strings, and the provider handles both cases.

## Example Usage

### Go Code Example

```go
package main

import (
    "context"
    "fmt"
    "os"

    "dev.helix.agent/internal/llm/providers/replicate"
    "dev.helix.agent/internal/models"
)

func main() {
    provider := replicate.NewProvider(
        os.Getenv("REPLICATE_API_KEY"),
        "", // Use default base URL
        "", // Use default model (meta/llama-2-70b-chat)
    )

    req := &models.LLMRequest{
        ID:     "request-1",
        Prompt: "You are a helpful assistant.",
        Messages: []models.Message{
            {Role: "user", Content: "What is the capital of France?"},
        },
        ModelParams: models.ModelParams{
            MaxTokens:   512,
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
    if chunk.FinishReason == "" {
        fmt.Print(chunk.Content) // Delta content
    }
}
```

### Using a Specific Model

```go
provider := replicate.NewProvider(
    os.Getenv("REPLICATE_API_KEY"),
    "",
    "meta/meta-llama-3.1-405b-instruct",
)
```

### Custom Retry Configuration

```go
retryConfig := replicate.RetryConfig{
    MaxRetries:   5,
    InitialDelay: 2 * time.Second,
    MaxDelay:     60 * time.Second,
    Multiplier:   2.0,
}

provider := replicate.NewProviderWithRetry(
    apiKey, "", "", retryConfig,
)
```

## Troubleshooting

### Authentication Error (401)

Verify your Replicate API token is correct. Tokens can be found at [replicate.com/account/api-tokens](https://replicate.com/account/api-tokens).

### Cold Start Delays

If predictions take a long time to start, the model is likely undergoing a cold start. The provider polls every 1 second and handles the `starting` and `processing` statuses automatically.

### Prediction Failed

If a prediction returns status `failed`, the error message from Replicate will be included in the error response. Common causes include invalid model versions or unsupported input parameters.

### Context Deadline Exceeded

For very large models (e.g., Llama 3.1 405B), cold starts may exceed the 300-second timeout. Consider using smaller models or dedicated hardware via Replicate's deployment feature.

## Additional Resources

- [Replicate API Documentation](https://replicate.com/docs/reference/http)
- [Replicate Model Explorer](https://replicate.com/explore)
- [Replicate Pricing](https://replicate.com/pricing)
