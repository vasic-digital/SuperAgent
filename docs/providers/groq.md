# Groq Provider

## Overview

Groq is an AI infrastructure company known for its custom LPU (Language Processing Unit) hardware that delivers extremely fast inference speeds. HelixAgent integrates with Groq's OpenAI-compatible API, providing access to a range of open-source models with ultra-low latency, including Llama, Mixtral, Gemma, and Qwen models, as well as Whisper for audio transcription.

## Authentication

Groq uses Bearer token authentication via the `Authorization` header. Groq API keys use the `gsk_` prefix.

| Header | Format | Required |
|--------|--------|----------|
| `Authorization` | `Bearer <api_key>` | Yes |

### Environment Variable

```bash
GROQ_API_KEY=gsk_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

## Configuration

Add the following to your `.env` file or environment:

```bash
# Required
GROQ_API_KEY=gsk_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

# Optional - Override default settings
GROQ_BASE_URL=https://api.groq.com/openai/v1/chat/completions
GROQ_MODEL=llama-3.3-70b-versatile
```

### Configuration Options

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `GROQ_API_KEY` | Yes | - | Your Groq API key (starts with `gsk_`) |
| `GROQ_BASE_URL` | No | `https://api.groq.com/openai/v1/chat/completions` | API endpoint URL |
| `GROQ_MODEL` | No | `llama-3.3-70b-versatile` | Default model to use |

## Supported Models

Models are discovered dynamically via the Groq `/openai/v1/models` API endpoint, with the following fallback list:

### Llama Models
- `llama-3.3-70b-versatile` (default)
- `llama-3.3-70b-specdec` - Speculative decoding variant
- `llama-3.2-90b-vision-preview` - Vision-capable
- `llama-3.2-11b-vision-preview` - Vision-capable
- `llama-3.2-3b-preview`
- `llama-3.2-1b-preview`
- `llama-3.1-70b-versatile`
- `llama-3.1-8b-instant`
- `llama3-70b-8192`
- `llama3-8b-8192`

### Llama 4 Models
- `llama-4-scout-17b-16e-instruct`
- `llama-4-maverick-17b-128e-instruct`

### Mixtral Models
- `mixtral-8x7b-32768`

### Gemma Models
- `gemma-7b-it`
- `gemma2-9b-it`

### Qwen Models
- `qwen-qwq-32b`
- `qwen-2.5-coder-32b`
- `qwen-2.5-32b`

### Whisper Models (Audio Transcription)
- `whisper-large-v3`
- `whisper-large-v3-turbo`
- `distil-whisper-large-v3-en`

## Capabilities

- Chat completion: Yes
- Streaming: Yes
- Tool/function calling: Yes
- Vision: Yes (via Llama 3.2 Vision models)
- Audio transcription: Yes (via Whisper models)
- JSON mode: Yes
- Code completion: Yes
- Fast inference: Yes (LPU-accelerated)
- Reasoning: Yes
- Code analysis: Yes

### Model Limits

| Parameter | Value |
|-----------|-------|
| Max tokens (context) | 131,072 |
| Max input length | 131,072 |
| Max output length | 32,768 |
| Max concurrent requests | 100 |
| HTTP client timeout | 60 seconds |

The HTTP client timeout is set to 60 seconds, shorter than most providers, because Groq's LPU hardware delivers responses significantly faster than GPU-based providers.

## API Endpoint

- **Chat Completions**: `https://api.groq.com/openai/v1/chat/completions`
- **Models list**: `https://api.groq.com/openai/v1/models`
- **Audio Transcription**: `https://api.groq.com/openai/v1/audio/transcriptions`

Groq uses an OpenAI-compatible API format under the `/openai/v1/` path prefix.

## Rate Limits

Rate limits depend on your Groq plan and the specific model:

| Plan | RPM | TPM |
|------|-----|-----|
| Free | 30 | 15,000 |
| Developer | 30 | 15,000 |
| Team | 300 | 150,000 |
| Enterprise | Custom | Custom |

Note: Rate limits vary by model. Smaller models generally have higher rate limits.

HelixAgent automatically implements retry with exponential backoff (initial delay 500ms, max delay 30s, multiplier 2.0) for rate limit (429) and server error (5xx) responses. The initial delay is shorter (500ms vs. 1s for other providers) because Groq's fast inference means requests complete quickly.

## Known Limitations

- Vision support is limited to specific models (Llama 3.2 Vision variants marked as "preview").
- Audio transcription uses a separate endpoint and is not integrated through the standard chat completion flow.
- Some models are marked as "preview" and may have reduced availability or different behavior.
- The `specdec` variant uses speculative decoding for faster inference but may have slightly different output characteristics.
- Rate limits on the free tier are relatively low compared to other providers.
- API key validation checks for the `gsk_` prefix during config validation.

### Groq-Specific Metadata

Groq responses include additional timing metadata not available from other providers:

- `prompt_time` - Time spent processing the prompt
- `completion_time` - Time spent generating the completion
- `total_time` - Total inference time
- `queue_time` - Time spent in the queue

These metrics are available in the response metadata and are useful for monitoring Groq's inference performance.

## Example Usage

### Go Code Example

```go
package main

import (
    "context"
    "fmt"
    "os"

    "dev.helix.agent/internal/llm/providers/groq"
    "dev.helix.agent/internal/models"
)

func main() {
    provider := groq.NewProvider(
        os.Getenv("GROQ_API_KEY"),
        "", // Use default base URL
        "", // Use default model (llama-3.3-70b-versatile)
    )

    req := &models.LLMRequest{
        ID:     "request-1",
        Prompt: "You are a helpful assistant.",
        Messages: []models.Message{
            {Role: "user", Content: "What are the SOLID principles in software engineering?"},
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
    fmt.Printf("Tokens used: %d\n", resp.TokensUsed)
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
        {Role: "user", Content: "Calculate the factorial of 10."},
    },
    Tools: []models.Tool{
        {
            Type: "function",
            Function: models.ToolFunction{
                Name:        "calculate_factorial",
                Description: "Calculate the factorial of a number",
                Parameters: map[string]interface{}{
                    "type": "object",
                    "properties": map[string]interface{}{
                        "n": map[string]interface{}{
                            "type":        "integer",
                            "description": "The number to calculate factorial for",
                        },
                    },
                    "required": []string{"n"},
                },
            },
        },
    },
    ModelParams: models.ModelParams{
        MaxTokens:   1024,
        Temperature: 0.0,
    },
}
```

### Accessing Groq-Specific Metrics

```go
resp, err := provider.Complete(ctx, req)
if err != nil {
    // handle error
}

if promptTime, ok := resp.Metadata["prompt_time"]; ok {
    fmt.Printf("Prompt processing time: %v\n", promptTime)
}
if completionTime, ok := resp.Metadata["completion_time"]; ok {
    fmt.Printf("Completion time: %v\n", completionTime)
}
```

### Custom Retry Configuration

```go
retryConfig := groq.RetryConfig{
    MaxRetries:   5,
    InitialDelay: 1 * time.Second,
    MaxDelay:     60 * time.Second,
    Multiplier:   2.0,
}

provider := groq.NewProviderWithRetry(
    apiKey, "", "", retryConfig,
)
```

## Troubleshooting

### Authentication Error (401)

Verify your Groq API key is correct and starts with `gsk_`. Keys can be generated at [console.groq.com/keys](https://console.groq.com/keys).

### Invalid API Key Format

The provider validates that the API key starts with `gsk_`. If you receive a validation error, ensure your key has the correct prefix.

### Rate Limit Error (429)

HelixAgent automatically retries with exponential backoff. Groq has relatively strict free-tier rate limits. Consider upgrading your plan or using smaller models which often have higher rate limits.

### Model Not Found

Verify the model name matches Groq's naming convention (e.g., `llama-3.3-70b-versatile`, not `meta-llama/Llama-3.3-70B-Instruct`). Groq uses simplified model names without the organization prefix.

## Additional Resources

- [Groq API Documentation](https://console.groq.com/docs)
- [Groq Console](https://console.groq.com)
- [Groq Supported Models](https://console.groq.com/docs/models)
