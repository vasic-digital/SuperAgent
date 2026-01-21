# Cerebras Provider Setup Guide

## Overview

Cerebras is a pioneering AI compute company that specializes in building large-scale AI accelerators. Their Wafer-Scale Engine (WSE) enables ultra-fast inference speeds, making Cerebras ideal for latency-sensitive AI applications. HelixAgent integrates with Cerebras's API to provide access to their hosted Llama models with industry-leading inference performance.

### Supported Models

- `llama-3.3-70b` - Latest Llama 3.3 70B parameter model (default)
- `llama-3.1-8b` - Efficient Llama 3.1 8B parameter model
- `llama-3.1-70b` - Llama 3.1 70B parameter model

### Key Features

- Ultra-fast inference (100+ tokens/second)
- Text completion and chat
- Streaming responses
- Reasoning capabilities
- Code completion and analysis
- Refactoring support
- Low latency responses

## API Key Setup

### Step 1: Create a Cerebras Account

1. Visit [cloud.cerebras.ai](https://cloud.cerebras.ai)
2. Sign up for an account
3. Complete the verification process

### Step 2: Generate an API Key

1. Navigate to **API Keys** in your dashboard
2. Click **Create new API key**
3. Name your key and set permissions
4. Copy the API key immediately - store it securely

### Step 3: Store Your API Key Securely

```bash
# Add to your environment or .env file
export CEREBRAS_API_KEY=your_api_key_here
```

## Environment Variable Configuration

Add the following to your `.env` file or environment:

```bash
# Required
CEREBRAS_API_KEY=csk-xxxxxxxxxxxxxxxxxxxxxxxx

# Optional - Override default settings
CEREBRAS_BASE_URL=https://api.cerebras.ai/v1/chat/completions
CEREBRAS_MODEL=llama-3.3-70b
```

### Configuration Options

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `CEREBRAS_API_KEY` | Yes | - | Your Cerebras API key |
| `CEREBRAS_BASE_URL` | No | `https://api.cerebras.ai/v1/chat/completions` | API endpoint URL |
| `CEREBRAS_MODEL` | No | `llama-3.3-70b` | Default model to use |

## Basic Usage Example

### Go Code Example

```go
package main

import (
    "context"
    "fmt"
    "os"

    "dev.helix.agent/internal/llm/providers/cerebras"
    "dev.helix.agent/internal/models"
)

func main() {
    // Create provider
    provider := cerebras.NewCerebrasProvider(
        os.Getenv("CEREBRAS_API_KEY"),
        "", // Use default base URL
        "", // Use default model
    )

    // Create request
    req := &models.LLMRequest{
        ID:     "request-1",
        Prompt: "You are a helpful AI assistant.",
        Messages: []models.Message{
            {Role: "user", Content: "Explain quantum computing in simple terms."},
        },
        ModelParams: models.ModelParams{
            MaxTokens:   1024,
            Temperature: 0.7,
        },
    }

    // Make completion request
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
// Enable streaming
streamChan, err := provider.CompleteStream(ctx, req)
if err != nil {
    fmt.Printf("Error: %v\n", err)
    return
}

for chunk := range streamChan {
    fmt.Print(chunk.Content)
}
```

### Code-Focused Request

```go
req := &models.LLMRequest{
    ID:     "request-1",
    Prompt: "You are an expert software engineer.",
    Messages: []models.Message{
        {Role: "user", Content: "Write a concurrent-safe map implementation in Go."},
    },
    ModelParams: models.ModelParams{
        MaxTokens:   2048,
        Temperature: 0.2, // Lower temperature for more deterministic code
    },
}
```

## Rate Limits and Quotas

### Default Rate Limits

| Plan | Requests/Minute | Tokens/Minute |
|------|-----------------|---------------|
| Free Tier | 30 | 30,000 |
| Standard | 300 | 300,000 |
| Enterprise | Custom | Custom |

### Model Context Limits

| Model | Max Input Tokens | Max Output Tokens |
|-------|-----------------|-------------------|
| llama-3.3-70b | 8,192 | 8,192 |
| llama-3.1-70b | 8,192 | 8,192 |
| llama-3.1-8b | 8,192 | 8,192 |

### Best Practices for Rate Limits

1. **Use exponential backoff** - HelixAgent automatically implements retry with backoff
2. **Monitor token usage** - Track consumption through response metadata
3. **Leverage ultra-fast inference** - Cerebras is ideal for latency-sensitive applications
4. **Choose appropriate temperature** - Use lower values (0.1-0.3) for code generation

## Troubleshooting

### Common Errors

#### Authentication Error (401)

```
Cerebras API error: 401 - {"error": {"message": "Invalid API key"}}
```

**Solution:**
- Verify your API key is correct
- Check that the key hasn't expired or been revoked
- Ensure the `Authorization: Bearer` header is properly formatted

#### Rate Limit Error (429)

```
Cerebras API error: 429 - {"error": {"message": "Rate limit exceeded"}}
```

**Solution:**
- Wait for the rate limit window to reset
- HelixAgent automatically retries with exponential backoff
- Consider upgrading your plan for higher limits

#### Model Not Found (404)

```
Cerebras API error: 404 - {"error": {"message": "Model not found"}}
```

**Solution:**
- Verify the model name is correct (`llama-3.3-70b`, `llama-3.1-70b`, `llama-3.1-8b`)
- Check Cerebras documentation for currently available models

#### No Choices Returned

```
Cerebras API returned no choices
```

**Solution:**
- Check if your prompt is valid and not empty
- Verify the model is available and functioning
- Review your request parameters

### Health Check

HelixAgent provides a health check endpoint for Cerebras:

```go
err := provider.HealthCheck()
if err != nil {
    fmt.Printf("Cerebras provider unhealthy: %v\n", err)
}
```

### Debug Logging

Enable debug logging to troubleshoot issues:

```bash
export GIN_MODE=debug
export LOG_LEVEL=debug
```

### Connection Issues

If experiencing connection timeouts:

1. Check network connectivity to `api.cerebras.ai`
2. Verify DNS resolution is working
3. Check if the service is accessible from your region
4. Consider increasing timeout (default is 120 seconds)

```go
// Custom retry configuration
retryConfig := cerebras.RetryConfig{
    MaxRetries:   5,
    InitialDelay: 2 * time.Second,
    MaxDelay:     60 * time.Second,
    Multiplier:   2.0,
}

provider := cerebras.NewCerebrasProviderWithRetry(
    apiKey,
    baseURL,
    model,
    retryConfig,
)
```

### Response Quality Issues

If output quality is poor:

1. **Lower temperature** - Use 0.1-0.3 for deterministic code
2. **Be specific** - Provide detailed requirements and context
3. **Use system prompts** - Set clear expectations for the assistant
4. **Include examples** - Show expected input/output formats

Example improved prompt:

```go
req := &models.LLMRequest{
    ID:     "request-1",
    Prompt: `You are an expert software engineer. Write clean, well-documented code.
Follow these guidelines:
- Use proper error handling
- Add meaningful comments
- Follow idiomatic patterns
- Include type definitions where appropriate`,
    Messages: []models.Message{
        {Role: "user", Content: "Create a REST API endpoint handler for user authentication."},
    },
    ModelParams: models.ModelParams{
        MaxTokens:   2048,
        Temperature: 0.2,
    },
}
```

## Performance Advantages

Cerebras offers unique performance benefits:

- **Ultra-Fast Inference**: 100+ tokens per second generation speed
- **Low Latency**: Responses start in milliseconds
- **Consistent Performance**: Wafer-Scale Engine provides predictable response times
- **Ideal for**: Real-time chat applications, interactive coding assistants, time-sensitive tasks

## Additional Resources

- [Cerebras Cloud Platform](https://cloud.cerebras.ai)
- [Cerebras API Documentation](https://cloud.cerebras.ai/docs)
- [Cerebras Inference Overview](https://www.cerebras.net/inference)
