# xAI Provider

## Overview

xAI is the AI company behind the Grok family of models. The xAI provider in HelixAgent uses their OpenAI-compatible API to access Grok models, which offer very large context windows (up to 2M tokens), vision capabilities, web/X search integration, code execution, and reasoning. xAI supports multi-region deployment with both US and EU endpoints, and its API keys follow the `xai-` prefix convention.

## Authentication

xAI uses the standard Bearer token authentication.

| Header | Value |
|--------|-------|
| `Authorization` | `Bearer <your_api_key>` |

API keys must start with the `xai-` prefix.

### Step 1: Create an xAI Account

1. Visit [console.x.ai](https://console.x.ai)
2. Sign up for an account
3. Complete the verification process

### Step 2: Generate an API Key

1. Navigate to **API Keys** in your xAI console
2. Create a new API key (prefix: `xai-`)
3. Copy the key immediately

### Step 3: Store Your API Key

```bash
export XAI_API_KEY=xai-xxxxxxxxxxxxxxxxxxxxxxxx
```

## Configuration

Add the following to your `.env` file or environment:

```bash
# Required
XAI_API_KEY=xai-xxxxxxxxxxxxxxxxxxxxxxxx

# Optional - Override default settings
XAI_BASE_URL=https://api.x.ai/v1
XAI_MODEL=grok-3-beta
XAI_REGION=us-east-1
```

### Configuration Options

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `XAI_API_KEY` | Yes | - | Your xAI API key (must start with `xai-`) |
| `XAI_BASE_URL` | No | `https://api.x.ai/v1` | API base URL |
| `XAI_MODEL` | No | `grok-3-beta` | Default model to use |
| `XAI_REGION` | No | `us-east-1` | API region (`us-east-1` or `eu-west-1`) |

### Regional Endpoints

| Region | Base URL |
|--------|----------|
| US (default) | `https://api.x.ai/v1` |
| EU | `https://api.x.ai/eu-west-1/v1` |

## Supported Models

Models are discovered dynamically via 3-tier discovery (API at `https://api.x.ai/v1/models`, models.dev, fallback). Fallback models include:

**Grok 4 Models**
- `grok-4` - Grok 4
- `grok-4-fast` - Grok 4 Fast

**Grok 3 Models**
- `grok-3` - Grok 3
- `grok-3-beta` - Grok 3 Beta (default)
- `grok-3-mini` - Grok 3 Mini
- `grok-3-mini-beta` - Grok 3 Mini Beta
- `grok-3-fast` - Grok 3 Fast
- `grok-3-fast-beta` - Grok 3 Fast Beta

**Grok 2 Models**
- `grok-2` - Grok 2
- `grok-2-1212` - Grok 2 (December 2024)
- `grok-2-vision` - Grok 2 Vision
- `grok-2-vision-1212` - Grok 2 Vision (December 2024)

**Legacy Models**
- `grok-vision-beta` - Grok Vision Beta

## Capabilities

- Chat completion: Yes
- Streaming: Yes (with `stream_options.include_usage` support)
- Tool/function calling: Yes
- Vision: Yes (via Grok vision models)
- Embeddings: No (not via this provider)
- JSON mode: Yes (`response_format` with `type: "json_object"`)
- Web search: Yes (via `web_search` feature)
- X (Twitter) search: Yes (via `x_search` feature)
- Code execution: Yes
- Reasoning: Yes
- Code completion: Yes
- Code analysis: Yes

## API Endpoint

| Endpoint | URL |
|----------|-----|
| Chat completions (US) | `https://api.x.ai/v1/chat/completions` |
| Chat completions (EU) | `https://api.x.ai/eu-west-1/v1/chat/completions` |
| Models list | `https://api.x.ai/v1/models` |

## Rate Limits

Rate limits depend on your xAI subscription tier. Check your xAI console for current limits.

### Model Context Limits (as configured in HelixAgent)

| Limit | Value |
|-------|-------|
| Max tokens (context window) | 2,000,000 |
| Max input length | 2,000,000 |
| Max output length | 131,072 |
| Max concurrent requests | 100 |

## Known Limitations

- **API key format validation**: The provider validates that API keys start with `xai-`. Keys without this prefix will be rejected during config validation.
- **Region affects base URL**: When switching regions, the base URL changes. Use `SetRegion()` to update both simultaneously, or use `NewProviderWithRegion()` at creation time.
- **Chat completions path appended**: The provider appends `/chat/completions` to the base URL when making requests. The base URL should be `https://api.x.ai/v1`, not `https://api.x.ai/v1/chat/completions`.
- **Health check path**: The health check appends `/models` to the base URL, querying the models list endpoint.
- **Region metadata**: The current region is included in response metadata under the `region` key.

## Example Usage

### Go Code Example

```go
package main

import (
    "context"
    "fmt"
    "os"

    "dev.helix.agent/internal/llm/providers/xai"
    "dev.helix.agent/internal/models"
)

func main() {
    provider := xai.NewProvider(
        os.Getenv("XAI_API_KEY"),
        "", // Use default base URL (US region)
        "", // Use default model (grok-3-beta)
    )

    req := &models.LLMRequest{
        ID:     "request-1",
        Prompt: "You are a helpful AI assistant.",
        Messages: []models.Message{
            {Role: "user", Content: "Explain the CAP theorem and its implications for distributed systems."},
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
    fmt.Printf("Region: %s\n", resp.Metadata["region"])
}
```

### Using EU Region

```go
provider := xai.NewProviderWithRegion(
    os.Getenv("XAI_API_KEY"),
    "",          // Use default model
    "eu-west-1", // EU region
)
```

### Switching Region at Runtime

```go
provider := xai.NewProvider(apiKey, "", "")
provider.SetRegion("eu-west-1") // Updates both region and base URL
```

### Using a Vision Model

```go
provider := xai.NewProvider(
    apiKey, "",
    "grok-2-vision",
)

req := &models.LLMRequest{
    ID:     "request-1",
    Prompt: "You are an image analysis assistant.",
    Messages: []models.Message{
        {Role: "user", Content: "Describe what you see in this image: [image data]"},
    },
    ModelParams: models.ModelParameters{
        MaxTokens:   1024,
        Temperature: 0.5,
    },
}
```

### Tool Calling Example

```go
req := &models.LLMRequest{
    ID:     "request-1",
    Prompt: "You are a helpful assistant with access to tools.",
    Messages: []models.Message{
        {Role: "user", Content: "Search for recent news about AI regulation."},
    },
    Tools: []models.Tool{
        {
            Type: "function",
            Function: models.ToolFunction{
                Name:        "web_search",
                Description: "Search the web for information",
                Parameters: map[string]interface{}{
                    "type": "object",
                    "properties": map[string]interface{}{
                        "query": map[string]interface{}{
                            "type":        "string",
                            "description": "Search query",
                        },
                    },
                    "required": []string{"query"},
                },
            },
        },
    },
    ModelParams: models.ModelParameters{
        MaxTokens: 2048,
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
retryConfig := xai.RetryConfig{
    MaxRetries:   5,
    InitialDelay: 2 * time.Second,
    MaxDelay:     60 * time.Second,
    Multiplier:   2.0,
}

provider := xai.NewProviderWithRetry(
    apiKey, "", "", "us-east-1", retryConfig,
)
```

## Additional Resources

- [xAI Console](https://console.x.ai)
- [xAI API Documentation](https://docs.x.ai)
- [Grok Models Overview](https://docs.x.ai/docs/models)
