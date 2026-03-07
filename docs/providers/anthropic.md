# Anthropic Provider

## Overview

Anthropic is an AI safety company that builds the Claude family of large language models. The Anthropic provider in HelixAgent uses Anthropic's native Messages API (not the OpenAI-compatible wrapper), providing direct access to Claude's full capabilities including extended thinking, computer use, vision, and tool calling. Anthropic uses a custom API format with content blocks rather than the OpenAI-compatible chat completions format.

## Authentication

Anthropic uses a custom authentication header rather than the standard `Authorization: Bearer` pattern.

| Header | Value |
|--------|-------|
| `x-api-key` | Your Anthropic API key |
| `anthropic-version` | `2023-06-01` |

### Step 1: Create an Anthropic Account

1. Visit [console.anthropic.com](https://console.anthropic.com)
2. Sign up for an account
3. Complete the verification process

### Step 2: Generate an API Key

1. Navigate to **API Keys** in your console
2. Click **Create Key**
3. Copy the API key immediately (prefix: `sk-ant-`)

### Step 3: Store Your API Key

```bash
export ANTHROPIC_API_KEY=sk-ant-xxxxxxxxxxxxxxxxxxxxxxxx
```

## Configuration

Add the following to your `.env` file or environment:

```bash
# Required
ANTHROPIC_API_KEY=sk-ant-xxxxxxxxxxxxxxxxxxxxxxxx

# Optional - Override default settings
ANTHROPIC_BASE_URL=https://api.anthropic.com/v1/messages
ANTHROPIC_MODEL=claude-sonnet-4-20250514
```

### Configuration Options

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `ANTHROPIC_API_KEY` | Yes | - | Your Anthropic API key |
| `ANTHROPIC_BASE_URL` | No | `https://api.anthropic.com/v1/messages` | API endpoint URL |
| `ANTHROPIC_MODEL` | No | `claude-sonnet-4-20250514` | Default model to use |

## Supported Models

Models are discovered dynamically via 3-tier discovery (API, models.dev, fallback). Fallback models include:

- `claude-opus-4-6` - Claude 4.6 Opus (latest, February 2026)
- `claude-opus-4-5-20251101` - Claude 4.5 Opus (November 2025)
- `claude-sonnet-4-5-20250929` - Claude 4.5 Sonnet (September 2025)
- `claude-haiku-4-5-20251001` - Claude 4.5 Haiku (October 2025)
- `claude-sonnet-4-20250514` - Claude 4 Sonnet (default)
- `claude-opus-4-20250514` - Claude 4 Opus
- `claude-haiku-4-20250514` - Claude 4 Haiku
- `claude-3-5-sonnet-20241022` - Claude 3.5 Sonnet
- `claude-3-5-haiku-20241022` - Claude 3.5 Haiku
- `claude-3-opus-20240229` - Claude 3 Opus
- `claude-3-sonnet-20240229` - Claude 3 Sonnet
- `claude-3-haiku-20240307` - Claude 3 Haiku

## Capabilities

- Chat completion: Yes
- Streaming: Yes
- Tool/function calling: Yes (native content block format with `input_schema`)
- Vision: Yes
- Embeddings: No (Anthropic does not offer an embeddings API)
- Extended thinking: Yes
- Computer use: Yes
- System prompts: Yes (sent as top-level `system` field, not as a message)
- Code completion: Yes
- Code analysis: Yes
- Reasoning: Yes

## API Endpoint

| Endpoint | URL |
|----------|-----|
| Messages (chat) | `https://api.anthropic.com/v1/messages` |
| Models list | `https://api.anthropic.com/v1/models` |

## Rate Limits

| Tier | Requests/Minute | Input Tokens/Minute | Output Tokens/Minute |
|------|-----------------|---------------------|----------------------|
| Free | 5 | 20,000 | 4,000 |
| Build (Tier 1) | 50 | 40,000 | 8,000 |
| Build (Tier 2) | 1,000 | 80,000 | 16,000 |
| Scale (Tier 3) | 2,000 | 200,000 | 40,000 |
| Scale (Tier 4) | 4,000 | 400,000 | 80,000 |

Limits vary by model. Check [Anthropic's rate limits documentation](https://docs.anthropic.com/en/api/rate-limits) for current values.

### Model Context Limits (as configured in HelixAgent)

| Limit | Value |
|-------|-------|
| Max tokens (context window) | 200,000 |
| Max input length | 200,000 |
| Max output length | 8,192 |
| Max concurrent requests | 50 |

## Known Limitations

- **Non-OpenAI API format**: Anthropic uses its own Messages API format with content blocks. System messages are not sent as regular messages but as a top-level `system` field.
- **API key format**: Keys must start with `sk-ant-`.
- **No embeddings**: Anthropic does not provide an embeddings API. Use a separate embedding provider (OpenAI, Cohere, Voyage, etc.).
- **HTTP client timeout**: Set to 300 seconds (5 minutes) due to potentially long responses from large models.
- **Health check**: Performs a minimal completion request rather than a lightweight endpoint check, which consumes tokens.

## Example Usage

### Go Code Example

```go
package main

import (
    "context"
    "fmt"
    "os"

    "dev.helix.agent/internal/llm/providers/anthropic"
    "dev.helix.agent/internal/models"
)

func main() {
    provider := anthropic.NewProvider(
        os.Getenv("ANTHROPIC_API_KEY"),
        "", // Use default base URL
        "", // Use default model (claude-sonnet-4-20250514)
    )

    req := &models.LLMRequest{
        ID:     "request-1",
        Prompt: "You are a helpful AI assistant.",
        Messages: []models.Message{
            {Role: "user", Content: "Explain the difference between goroutines and threads."},
        },
        ModelParams: models.ModelParameters{
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
    ModelParams: models.ModelParameters{
        MaxTokens: 1024,
    },
}
```

### Custom Retry Configuration

```go
retryConfig := anthropic.RetryConfig{
    MaxRetries:   5,
    InitialDelay: 2 * time.Second,
    MaxDelay:     60 * time.Second,
    Multiplier:   2.0,
}

provider := anthropic.NewProviderWithRetry(
    apiKey, "", "", retryConfig,
)
```

## Additional Resources

- [Anthropic API Documentation](https://docs.anthropic.com/en/api)
- [Anthropic Console](https://console.anthropic.com)
- [Claude Model Overview](https://docs.anthropic.com/en/docs/about-claude/models)
