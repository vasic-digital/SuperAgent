# Fireworks AI Provider

## Overview

Fireworks AI is a fast inference platform specializing in high-throughput, low-latency model serving. It provides access to a wide catalog of open-source models from Llama, Qwen, DeepSeek, Mistral, and its own Firefunction/Firellava models, all through an OpenAI-compatible API. Fireworks is particularly strong in fast inference, vision capabilities, structured output (grammar mode), and function calling.

## Authentication

Fireworks uses the standard Bearer token authentication.

| Header | Value |
|--------|-------|
| `Authorization` | `Bearer <your_api_key>` |

### Step 1: Create a Fireworks Account

1. Visit [fireworks.ai](https://fireworks.ai)
2. Sign up for an account
3. Complete the verification process

### Step 2: Generate an API Key

1. Navigate to **API Keys** in your Fireworks dashboard
2. Create a new API key (prefix: `fw_`)
3. Copy the key immediately

### Step 3: Store Your API Key

```bash
export FIREWORKS_API_KEY=fw_xxxxxxxxxxxxxxxxxxxxxxxx
```

## Configuration

Add the following to your `.env` file or environment:

```bash
# Required
FIREWORKS_API_KEY=fw_xxxxxxxxxxxxxxxxxxxxxxxx

# Optional - Override default settings
FIREWORKS_BASE_URL=https://api.fireworks.ai/inference/v1/chat/completions
FIREWORKS_MODEL=accounts/fireworks/models/llama-v3p3-70b-instruct
```

### Configuration Options

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `FIREWORKS_API_KEY` | Yes | - | Your Fireworks API key |
| `FIREWORKS_BASE_URL` | No | `https://api.fireworks.ai/inference/v1/chat/completions` | API endpoint URL |
| `FIREWORKS_MODEL` | No | `accounts/fireworks/models/llama-v3p3-70b-instruct` | Default model to use |

## Supported Models

Models are discovered dynamically via 3-tier discovery (API at `https://api.fireworks.ai/inference/v1/models`, models.dev, fallback). Fallback models include:

**Llama 3.3 Models**
- `accounts/fireworks/models/llama-v3p3-70b-instruct` - Llama 3.3 70B (default)

**Llama 3.1 Models**
- `accounts/fireworks/models/llama-v3p1-405b-instruct` - Llama 3.1 405B
- `accounts/fireworks/models/llama-v3p1-70b-instruct` - Llama 3.1 70B
- `accounts/fireworks/models/llama-v3p1-8b-instruct` - Llama 3.1 8B

**Llama 3.2 Models (with vision)**
- `accounts/fireworks/models/llama-v3p2-90b-vision-instruct` - Llama 3.2 90B Vision
- `accounts/fireworks/models/llama-v3p2-11b-vision-instruct` - Llama 3.2 11B Vision
- `accounts/fireworks/models/llama-v3p2-3b-instruct` - Llama 3.2 3B
- `accounts/fireworks/models/llama-v3p2-1b-instruct` - Llama 3.2 1B

**Qwen Models**
- `accounts/fireworks/models/qwen2p5-72b-instruct` - Qwen 2.5 72B
- `accounts/fireworks/models/qwen2p5-coder-32b-instruct` - Qwen 2.5 Coder 32B

**DeepSeek Models**
- `accounts/fireworks/models/deepseek-r1` - DeepSeek R1
- `accounts/fireworks/models/deepseek-v3` - DeepSeek V3

**Mistral Models**
- `accounts/fireworks/models/mixtral-8x22b-instruct` - Mixtral 8x22B
- `accounts/fireworks/models/mixtral-8x7b-instruct` - Mixtral 8x7B

**Fireworks Native Models**
- `accounts/fireworks/models/firefunction-v2` - Firefunction V2 (optimized for function calling)
- `accounts/fireworks/models/firellava-13b` - Firellava 13B (vision model)

Model IDs use the `accounts/fireworks/models/<name>` format.

## Capabilities

- Chat completion: Yes
- Streaming: Yes
- Tool/function calling: Yes
- Vision: Yes (via Llama 3.2 vision and Firellava models)
- Embeddings: Yes (separate endpoint)
- JSON mode: Yes (`response_format` with `type: "json_object"` and optional `schema`)
- Grammar mode: Yes (structured output via grammar constraints)
- Code completion: Yes
- Code analysis: Yes
- Reasoning: Yes

## API Endpoint

| Endpoint | URL |
|----------|-----|
| Chat completions | `https://api.fireworks.ai/inference/v1/chat/completions` |
| Models list | `https://api.fireworks.ai/inference/v1/models` |

## Rate Limits

Rate limits depend on your Fireworks subscription tier. Check your Fireworks dashboard for current limits.

### Model Context Limits (as configured in HelixAgent)

| Limit | Value |
|-------|-------|
| Max tokens (context window) | 131,072 |
| Max input length | 131,072 |
| Max output length | 16,384 |
| Max concurrent requests | 100 |

## Known Limitations

- **Model ID format**: Fireworks uses the `accounts/fireworks/models/<name>` prefix for model IDs, not simple model names. Using a bare model name (e.g., `llama-v3p3-70b-instruct`) may result in a 404 error.
- **API path includes `/inference/v1/`**: The Fireworks API uses `/inference/v1/` in its path, not just `/v1/`.
- **Model deprecation**: Some older model IDs (e.g., `llama-v3p1-70b-instruct` without the `accounts/fireworks/models/` prefix) have been deprecated and return 404. Always use the full model ID.
- **Health check uses models endpoint**: The health check queries the models list API, so it does not consume tokens.

## Example Usage

### Go Code Example

```go
package main

import (
    "context"
    "fmt"
    "os"

    "dev.helix.agent/internal/llm/providers/fireworks"
    "dev.helix.agent/internal/models"
)

func main() {
    provider := fireworks.NewProvider(
        os.Getenv("FIREWORKS_API_KEY"),
        "", // Use default base URL
        "", // Use default model (llama-v3p3-70b-instruct)
    )

    req := &models.LLMRequest{
        ID:     "request-1",
        Prompt: "You are an expert software engineer.",
        Messages: []models.Message{
            {Role: "user", Content: "Write a thread-safe LRU cache in Go."},
        },
        ModelParams: models.ModelParameters{
            MaxTokens:   2048,
            Temperature: 0.2,
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

### Using a Specific Model

```go
req := &models.LLMRequest{
    ID:     "request-1",
    Prompt: "You are a coding assistant.",
    Messages: []models.Message{
        {Role: "user", Content: "Implement a merge sort in Python."},
    },
    ModelParams: models.ModelParameters{
        Model:       "accounts/fireworks/models/qwen2p5-coder-32b-instruct",
        MaxTokens:   2048,
        Temperature: 0.1,
    },
}
```

### Tool Calling with Firefunction

```go
provider := fireworks.NewProvider(
    apiKey, "",
    "accounts/fireworks/models/firefunction-v2",
)

req := &models.LLMRequest{
    ID:     "request-1",
    Prompt: "You are a helpful assistant with access to tools.",
    Messages: []models.Message{
        {Role: "user", Content: "What is the weather in Tokyo?"},
    },
    Tools: []models.Tool{
        {
            Type: "function",
            Function: models.ToolFunction{
                Name:        "get_weather",
                Description: "Get current weather for a city",
                Parameters: map[string]interface{}{
                    "type": "object",
                    "properties": map[string]interface{}{
                        "city": map[string]interface{}{
                            "type":        "string",
                            "description": "City name",
                        },
                    },
                    "required": []string{"city"},
                },
            },
        },
    },
    ModelParams: models.ModelParameters{
        MaxTokens: 1024,
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
retryConfig := fireworks.RetryConfig{
    MaxRetries:   5,
    InitialDelay: 2 * time.Second,
    MaxDelay:     60 * time.Second,
    Multiplier:   2.0,
}

provider := fireworks.NewProviderWithRetry(
    apiKey, "", "", retryConfig,
)
```

## Additional Resources

- [Fireworks AI Platform](https://fireworks.ai)
- [Fireworks API Documentation](https://docs.fireworks.ai)
- [Fireworks Model Library](https://fireworks.ai/models)
