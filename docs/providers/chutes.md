# Chutes Provider

## Overview

Chutes AI is an inference platform that provides access to popular open-source models through an OpenAI-compatible API. It hosts models from DeepSeek, Meta (Llama), Qwen, and Mistral, offering a unified gateway to multiple model families. Chutes is particularly useful for accessing high-capability open models like DeepSeek-V3 and DeepSeek-R1 at competitive inference costs.

**Important**: The Chutes inference API endpoint is `llm.chutes.ai`, NOT `api.chutes.ai`. This is a common source of confusion.

## Authentication

Chutes uses the standard Bearer token authentication.

| Header | Value |
|--------|-------|
| `Authorization` | `Bearer <your_api_key>` |

### Step 1: Create a Chutes Account

1. Visit [chutes.ai](https://chutes.ai)
2. Sign up for an account
3. Navigate to API settings

### Step 2: Generate an API Key

1. Go to your Chutes dashboard
2. Generate a new API key
3. Copy the key immediately (prefix: `cpk_`)

### Step 3: Store Your API Key

```bash
export CHUTES_API_KEY=cpk_xxxxxxxxxxxxxxxxxxxxxxxx
```

## Configuration

Add the following to your `.env` file or environment:

```bash
# Required
CHUTES_API_KEY=cpk_xxxxxxxxxxxxxxxxxxxxxxxx

# Optional - Override default settings
CHUTES_BASE_URL=https://llm.chutes.ai/v1/chat/completions
CHUTES_MODEL=deepseek-ai/DeepSeek-V3
```

### Configuration Options

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `CHUTES_API_KEY` | Yes | - | Your Chutes API key |
| `CHUTES_BASE_URL` | No | `https://llm.chutes.ai/v1/chat/completions` | API endpoint URL |
| `CHUTES_MODEL` | No | `deepseek-ai/DeepSeek-V3` | Default model to use |

## Supported Models

Models are discovered dynamically via 3-tier discovery (API at `https://llm.chutes.ai/v1/models`, models.dev, fallback). Fallback models include:

- `deepseek-ai/DeepSeek-V3` - DeepSeek V3 (default)
- `deepseek-ai/DeepSeek-R1` - DeepSeek R1 reasoning model
- `meta-llama/Llama-4-Maverick-17B-128E-Instruct-FP8` - Llama 4 Maverick
- `meta-llama/Llama-4-Scout-17B-16E-Instruct` - Llama 4 Scout
- `Qwen/Qwen3-235B-A22B` - Qwen 3 235B
- `Qwen/Qwen3-32B` - Qwen 3 32B
- `mistralai/Devstral-Small-2505` - Mistral Devstral Small

Model IDs use the `org/model-name` format (e.g., `deepseek-ai/DeepSeek-V3`).

## Capabilities

- Chat completion: Yes
- Streaming: Yes
- Tool/function calling: Yes
- Vision: No
- Embeddings: No
- Code completion: Yes
- Code analysis: Yes
- Reasoning: Yes

## API Endpoint

| Endpoint | URL |
|----------|-----|
| Chat completions | `https://llm.chutes.ai/v1/chat/completions` |
| Models list | `https://llm.chutes.ai/v1/models` |

## Rate Limits

Rate limits depend on your Chutes subscription tier. Check your Chutes dashboard for current limits.

### Model Context Limits (as configured in HelixAgent)

| Limit | Value |
|-------|-------|
| Max tokens (context window) | 131,072 |
| Max input length | 131,072 |
| Max output length | 4,096 |
| Max concurrent requests | 50 |

## Known Limitations

- **Inference endpoint URL**: The API uses `llm.chutes.ai`, not `api.chutes.ai`. Using the wrong hostname will result in connection failures.
- **Model ID format**: Models use the `org/model-name` format (e.g., `deepseek-ai/DeepSeek-V3`), not simple model names.
- **No vision support**: Image inputs are not supported.
- **No embeddings**: Chutes does not provide an embeddings API through this provider.
- **Health check uses models endpoint**: The health check queries `https://llm.chutes.ai/v1/models` rather than performing a completion, so it does not consume tokens.

## Example Usage

### Go Code Example

```go
package main

import (
    "context"
    "fmt"
    "os"

    "dev.helix.agent/internal/llm/providers/chutes"
    "dev.helix.agent/internal/models"
)

func main() {
    provider := chutes.NewProvider(
        os.Getenv("CHUTES_API_KEY"),
        "", // Use default base URL (llm.chutes.ai)
        "", // Use default model (deepseek-ai/DeepSeek-V3)
    )

    req := &models.LLMRequest{
        ID:     "request-1",
        Prompt: "You are a helpful AI assistant.",
        Messages: []models.Message{
            {Role: "user", Content: "Explain how a B-tree index works in databases."},
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
}
```

### Using DeepSeek-R1 Reasoning Model

```go
req := &models.LLMRequest{
    ID:     "request-1",
    Prompt: "Think step by step.",
    Messages: []models.Message{
        {Role: "user", Content: "Prove that the square root of 2 is irrational."},
    },
    ModelParams: models.ModelParameters{
        Model:       "deepseek-ai/DeepSeek-R1",
        MaxTokens:   4096,
        Temperature: 0.3,
    },
}
```

### Tool Calling Example

```go
req := &models.LLMRequest{
    ID:     "request-1",
    Prompt: "You are a helpful assistant.",
    Messages: []models.Message{
        {Role: "user", Content: "What files are in the current directory?"},
    },
    Tools: []models.Tool{
        {
            Type: "function",
            Function: models.ToolFunction{
                Name:        "list_files",
                Description: "List files in a directory",
                Parameters: map[string]interface{}{
                    "type": "object",
                    "properties": map[string]interface{}{
                        "path": map[string]interface{}{
                            "type":        "string",
                            "description": "Directory path",
                        },
                    },
                    "required": []string{"path"},
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
retryConfig := chutes.RetryConfig{
    MaxRetries:   5,
    InitialDelay: 2 * time.Second,
    MaxDelay:     60 * time.Second,
    Multiplier:   2.0,
}

provider := chutes.NewProviderWithRetry(
    apiKey, "", "", retryConfig,
)
```

## Additional Resources

- [Chutes AI](https://chutes.ai)
- [Chutes API Documentation](https://docs.chutes.ai)
