# Cohere Provider

## Overview

Cohere is an enterprise AI company specializing in language models for search, classification, and generation. HelixAgent integrates with Cohere's v2 Chat API, providing access to the Command R family of models with support for RAG (Retrieval-Augmented Generation), tool calling, reranking, and citation generation.

## Authentication

Cohere uses Bearer token authentication via the `Authorization` header.

| Header | Format | Required |
|--------|--------|----------|
| `Authorization` | `Bearer <api_key>` | Yes |
| `Accept` | `application/json` | Yes |

### Environment Variable

```bash
COHERE_API_KEY=your-api-key-here
```

## Configuration

Add the following to your `.env` file or environment:

```bash
# Required
COHERE_API_KEY=your-api-key-here

# Optional - Override default settings
COHERE_BASE_URL=https://api.cohere.com/v2/chat
COHERE_MODEL=command-r-plus
```

### Configuration Options

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `COHERE_API_KEY` | Yes | - | Your Cohere API key |
| `COHERE_BASE_URL` | No | `https://api.cohere.com/v2/chat` | API endpoint URL |
| `COHERE_MODEL` | No | `command-r-plus` | Default model to use |

## Supported Models

Models are discovered dynamically via the Cohere `/v1/models` API endpoint (using a custom response parser), with the following fallback list:

- `command-r-plus` (default) - Most capable model for complex tasks
- `command-r-plus-08-2024` - August 2024 snapshot
- `command-r` - Fast and capable for most tasks
- `command-r-08-2024` - August 2024 snapshot
- `command` - Previous generation command model
- `command-light` - Lightweight variant
- `command-nightly` - Nightly development build
- `command-light-nightly` - Lightweight nightly build
- `c4ai-aya-expanse-8b` - Multilingual model (8B)
- `c4ai-aya-expanse-32b` - Multilingual model (32B)

## Capabilities

- Chat completion: Yes
- Streaming: Yes
- Tool/function calling: Yes
- Vision: No
- Embeddings: Yes (via separate endpoint)
- RAG (documents + citations): Yes
- Reranking: Yes
- Classification: Yes
- Summarization: Yes
- JSON mode: Yes
- Reasoning: Yes
- Code completion: Yes
- Code analysis: Yes

### Model Limits

| Parameter | Value |
|-----------|-------|
| Max tokens (context) | 128,000 |
| Max input length | 128,000 |
| Max output length | 4,096 |
| Max concurrent requests | 100 |
| HTTP client timeout | 120 seconds |

## API Endpoint

- **Chat (v2)**: `https://api.cohere.com/v2/chat`
- **Models list**: `https://api.cohere.com/v1/models`

The Cohere provider uses the v2 Chat API. Note: the base URL uses `api.cohere.com` (not `api.cohere.ai`), and the chat endpoint is at `/v2/chat` (not `/v1`).

## Rate Limits

Rate limits depend on your Cohere plan:

| Plan | RPM | TPM |
|------|-----|-----|
| Trial | 20 | 10,000 |
| Production | 10,000 | 10,000,000 |
| Enterprise | Custom | Custom |

HelixAgent automatically implements retry with exponential backoff (initial delay 1s, max delay 30s, multiplier 2.0) for rate limit (429) and server error (5xx) responses.

## Known Limitations

- Vision is not supported by any Cohere model.
- The v2 API uses a different message structure than OpenAI: response content is returned as an array of `ContentPart` objects (each with `type` and `text`), not a single string.
- System prompts are sent via the `preamble` field rather than a system message in the messages array.
- Streaming uses Cohere-specific event types (`content-delta`, `message-end`) rather than OpenAI-style SSE.
- Cohere finish reasons differ from OpenAI: `COMPLETE` (instead of `stop`), `MAX_TOKENS` (instead of `length`), `ERROR`.
- Tool call responses use a `parameters` field (with a map) in addition to the standard `function.arguments` string format.

## Example Usage

### Go Code Example

```go
package main

import (
    "context"
    "fmt"
    "os"

    "dev.helix.agent/internal/llm/providers/cohere"
    "dev.helix.agent/internal/models"
)

func main() {
    provider := cohere.NewProvider(
        os.Getenv("COHERE_API_KEY"),
        "", // Use default base URL
        "", // Use default model (command-r-plus)
    )

    req := &models.LLMRequest{
        ID:     "request-1",
        Prompt: "You are a helpful research assistant.",
        Messages: []models.Message{
            {Role: "user", Content: "Summarize the key differences between TCP and UDP."},
        },
        ModelParams: models.ModelParams{
            MaxTokens:   4096,
            Temperature: 0.5,
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

### Custom Retry Configuration

```go
retryConfig := cohere.RetryConfig{
    MaxRetries:   5,
    InitialDelay: 2 * time.Second,
    MaxDelay:     60 * time.Second,
    Multiplier:   2.0,
}

provider := cohere.NewProviderWithRetry(
    apiKey, "", "", retryConfig,
)
```

## Troubleshooting

### Authentication Error (401)

Verify your API key is correct. Cohere API keys can be generated at [dashboard.cohere.com](https://dashboard.cohere.com).

### Rate Limit Error (429)

HelixAgent automatically retries with exponential backoff. For production workloads, upgrade from the trial plan.

### URL Mismatch

Ensure you are using `api.cohere.com` (not `api.cohere.ai`) and the `/v2/chat` endpoint (not `/v1`).

## Additional Resources

- [Cohere API Documentation](https://docs.cohere.com)
- [Cohere Dashboard](https://dashboard.cohere.com)
- [Cohere Model Cards](https://docs.cohere.com/docs/models)
