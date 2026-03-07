# Perplexity Provider

## Overview

Perplexity is a search-focused AI platform that combines large language models with real-time web search capabilities. The Perplexity provider in HelixAgent uses their OpenAI-compatible chat completions API, providing access to Sonar models that can search the web and return responses with inline citations. This makes Perplexity particularly useful for queries requiring up-to-date information or factual grounding.

## Authentication

Perplexity uses the standard Bearer token authentication.

| Header | Value |
|--------|-------|
| `Authorization` | `Bearer <your_api_key>` |

### Step 1: Create a Perplexity Account

1. Visit [perplexity.ai](https://www.perplexity.ai)
2. Sign up for an account
3. Navigate to API settings

### Step 2: Generate an API Key

1. Go to [perplexity.ai/settings/api](https://www.perplexity.ai/settings/api)
2. Generate a new API key
3. Copy the key immediately (prefix: `pplx-`)

### Step 3: Store Your API Key

```bash
export PERPLEXITY_API_KEY=pplx-xxxxxxxxxxxxxxxxxxxxxxxx
```

## Configuration

Add the following to your `.env` file or environment:

```bash
# Required
PERPLEXITY_API_KEY=pplx-xxxxxxxxxxxxxxxxxxxxxxxx

# Optional - Override default settings
PERPLEXITY_BASE_URL=https://api.perplexity.ai/chat/completions
PERPLEXITY_MODEL=llama-3.1-sonar-large-128k-online
```

### Configuration Options

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PERPLEXITY_API_KEY` | Yes | - | Your Perplexity API key |
| `PERPLEXITY_BASE_URL` | No | `https://api.perplexity.ai/chat/completions` | API endpoint URL |
| `PERPLEXITY_MODEL` | No | `llama-3.1-sonar-large-128k-online` | Default model to use |

## Supported Models

Models are discovered dynamically via 3-tier discovery (API, models.dev, fallback). Fallback models include:

**Sonar Online Models (with web search)**
- `llama-3.1-sonar-small-128k-online` - Small Sonar with online search
- `llama-3.1-sonar-large-128k-online` - Large Sonar with online search (default)
- `llama-3.1-sonar-huge-128k-online` - Huge Sonar with online search

**Sonar Chat Models (without web search)**
- `llama-3.1-sonar-small-128k-chat` - Small Sonar chat
- `llama-3.1-sonar-large-128k-chat` - Large Sonar chat

**Open Models**
- `llama-3.1-8b-instruct` - Llama 3.1 8B
- `llama-3.1-70b-instruct` - Llama 3.1 70B

## Capabilities

- Chat completion: Yes
- Streaming: Yes
- Tool/function calling: No
- Vision: No
- Embeddings: No
- Online search: Yes (Sonar online models)
- Citations: Yes (returned in response metadata)
- Search domain filtering: Yes
- Search recency filtering: Yes
- Reasoning: Yes
- Code completion: Yes
- Code analysis: Yes

## API Endpoint

| Endpoint | URL |
|----------|-----|
| Chat completions | `https://api.perplexity.ai/chat/completions` |

## Rate Limits

Rate limits depend on your subscription tier. Check your Perplexity dashboard for current limits.

### Model Context Limits (as configured in HelixAgent)

| Limit | Value |
|-------|-------|
| Max tokens (context window) | 128,000 |
| Max input length | 127,000 |
| Max output length | 4,096 |
| Max concurrent requests | 50 |

## Known Limitations

- **No tool/function calling**: Perplexity does not support the OpenAI-compatible tools API.
- **No vision**: Image inputs are not supported.
- **No models endpoint**: Perplexity does not expose a `/v1/models` endpoint, so model discovery relies on models.dev and hardcoded fallbacks.
- **Health check uses completion**: Because there is no lightweight models endpoint, the health check performs a minimal completion request, which consumes tokens.
- **Citations format**: Citations are returned as a list of URLs in the response metadata under the `citations` key. The response text may contain inline references to these citations.

## Provider-Specific Options

Perplexity supports additional search-specific parameters via `ModelParams.ProviderSpecific`:

| Parameter | Type | Description |
|-----------|------|-------------|
| `search_domain_filter` | `[]string` | Restrict search to specific domains |
| `search_recency_filter` | `string` | Filter results by recency (e.g., `"month"`, `"week"`, `"day"`) |
| `return_images` | `bool` | Include images in search results |

## Example Usage

### Go Code Example

```go
package main

import (
    "context"
    "fmt"
    "os"

    "dev.helix.agent/internal/llm/providers/perplexity"
    "dev.helix.agent/internal/models"
)

func main() {
    provider := perplexity.NewProvider(
        os.Getenv("PERPLEXITY_API_KEY"),
        "", // Use default base URL
        "", // Use default model (llama-3.1-sonar-large-128k-online)
    )

    req := &models.LLMRequest{
        ID:     "request-1",
        Prompt: "You are a helpful research assistant.",
        Messages: []models.Message{
            {Role: "user", Content: "What are the latest developments in quantum computing?"},
        },
        ModelParams: models.ModelParameters{
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

    // Access citations from metadata
    if citations, ok := resp.Metadata["citations"]; ok {
        fmt.Printf("Citations: %v\n", citations)
    }
}
```

### Search with Domain Filtering

```go
req := &models.LLMRequest{
    ID:     "request-1",
    Prompt: "You are a technical research assistant.",
    Messages: []models.Message{
        {Role: "user", Content: "Latest Go 1.24 features"},
    },
    ModelParams: models.ModelParameters{
        MaxTokens:   2048,
        Temperature: 0.5,
        ProviderSpecific: map[string]interface{}{
            "search_domain_filter":  []string{"go.dev", "github.com"},
            "search_recency_filter": "month",
        },
    },
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
retryConfig := perplexity.RetryConfig{
    MaxRetries:   5,
    InitialDelay: 2 * time.Second,
    MaxDelay:     60 * time.Second,
    Multiplier:   2.0,
}

provider := perplexity.NewProviderWithRetry(
    apiKey, "", "", retryConfig,
)
```

## Additional Resources

- [Perplexity API Documentation](https://docs.perplexity.ai)
- [Perplexity API Settings](https://www.perplexity.ai/settings/api)
- [Sonar Models Overview](https://docs.perplexity.ai/docs/model-cards)
