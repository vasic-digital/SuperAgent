# Mistral Provider Setup Guide

## Overview

Mistral AI is a French AI company that develops open and commercial large language models known for their efficiency and strong performance. HelixAgent integrates with Mistral's API to provide access to their full range of models, from the efficient Mistral Small to the powerful Mistral Large, including specialized code models.

### Supported Models

- `mistral-large-latest` - Most capable Mistral model (default)
- `mistral-medium` - Balanced performance and cost
- `mistral-small-latest` - Efficient and fast
- `open-mistral-7b` - Open-weight 7B model
- `open-mixtral-8x7b` - Mixture of Experts 8x7B
- `open-mixtral-8x22b` - Large Mixture of Experts 8x22B
- `codestral-latest` - Specialized for code generation

### Key Features

- Text completion and chat
- Function calling
- Streaming responses
- Tool use support
- Reasoning capabilities
- Code completion and analysis
- Refactoring support
- Safe prompt filtering

## API Key Setup

### Step 1: Create a Mistral Account

1. Visit [console.mistral.ai](https://console.mistral.ai)
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
export MISTRAL_API_KEY=your_api_key_here
```

## Environment Variable Configuration

Add the following to your `.env` file or environment:

```bash
# Required
MISTRAL_API_KEY=xxxxxxxxxxxxxxxxxxxxxxxx

# Optional - Override default settings
MISTRAL_BASE_URL=https://api.mistral.ai/v1/chat/completions
MISTRAL_MODEL=mistral-large-latest
```

### Configuration Options

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `MISTRAL_API_KEY` | Yes | - | Your Mistral API key |
| `MISTRAL_BASE_URL` | No | `https://api.mistral.ai/v1/chat/completions` | API endpoint URL |
| `MISTRAL_MODEL` | No | `mistral-large-latest` | Default model to use |

## Basic Usage Example

### Go Code Example

```go
package main

import (
    "context"
    "fmt"
    "os"

    "dev.helix.agent/internal/llm/providers/mistral"
    "dev.helix.agent/internal/models"
)

func main() {
    // Create provider
    provider := mistral.NewMistralProvider(
        os.Getenv("MISTRAL_API_KEY"),
        "", // Use default base URL
        "", // Use default model
    )

    // Create request
    req := &models.LLMRequest{
        ID:     "request-1",
        Prompt: "You are a helpful AI assistant.",
        Messages: []models.Message{
            {Role: "user", Content: "Explain the difference between REST and GraphQL APIs."},
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

### Function Calling Example

```go
req := &models.LLMRequest{
    ID:     "request-1",
    Prompt: "You are a helpful assistant with access to tools.",
    Messages: []models.Message{
        {Role: "user", Content: "What's the weather in Paris?"},
    },
    Tools: []models.Tool{
        {
            Type: "function",
            Function: models.ToolFunction{
                Name:        "get_weather",
                Description: "Get the current weather for a location",
                Parameters: map[string]interface{}{
                    "type": "object",
                    "properties": map[string]interface{}{
                        "location": map[string]interface{}{
                            "type":        "string",
                            "description": "The city and country",
                        },
                    },
                    "required": []string{"location"},
                },
            },
        },
    },
    ToolChoice: "auto",
    ModelParams: models.ModelParams{
        MaxTokens:   1024,
        Temperature: 0.7,
    },
}

resp, err := provider.Complete(ctx, req)
if err != nil {
    fmt.Printf("Error: %v\n", err)
    return
}

// Check for tool calls
if len(resp.ToolCalls) > 0 {
    for _, tc := range resp.ToolCalls {
        fmt.Printf("Tool call: %s(%s)\n", tc.Function.Name, tc.Function.Arguments)
    }
}
```

### Code-Focused Request (Codestral)

```go
// For code generation, use the Codestral model
provider := mistral.NewMistralProvider(
    os.Getenv("MISTRAL_API_KEY"),
    "",
    "codestral-latest",
)

req := &models.LLMRequest{
    ID:     "request-1",
    Prompt: "You are an expert programmer.",
    Messages: []models.Message{
        {Role: "user", Content: "Write a Python async function for concurrent HTTP requests with rate limiting."},
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
| Free Tier | 5 | 5,000 |
| Standard | 120 | 120,000 |
| Enterprise | Custom | Custom |

### Model Context Limits

| Model | Max Input Tokens | Max Output Tokens |
|-------|-----------------|-------------------|
| mistral-large-latest | 32,768 | 32,768 |
| mistral-medium | 32,768 | 32,768 |
| mistral-small-latest | 32,768 | 32,768 |
| codestral-latest | 32,768 | 32,768 |
| open-mixtral-8x22b | 65,536 | 32,768 |

### Best Practices for Rate Limits

1. **Use exponential backoff** - HelixAgent automatically implements retry with backoff
2. **Monitor token usage** - Track consumption through response metadata
3. **Choose the right model** - Use smaller models for simple tasks
4. **Choose appropriate temperature** - Use lower values (0.1-0.3) for code generation

## Model Selection Guide

| Use Case | Recommended Model | Reason |
|----------|-------------------|--------|
| Complex reasoning | `mistral-large-latest` | Most capable, best for nuanced tasks |
| Code generation | `codestral-latest` | Specialized for programming |
| General chat | `mistral-medium` | Good balance of performance and cost |
| High-volume tasks | `mistral-small-latest` | Fast and cost-effective |
| Open-source needs | `open-mixtral-8x22b` | Best open model |

## Troubleshooting

### Common Errors

#### Authentication Error (401)

```
Mistral API error: 401 - {"message": "Unauthorized"}
```

**Solution:**
- Verify your API key is correct
- Check that the key hasn't expired or been revoked
- Ensure the `Authorization: Bearer` header is properly formatted

#### Rate Limit Error (429)

```
Mistral API error: 429 - {"message": "Rate limit exceeded"}
```

**Solution:**
- Wait for the rate limit window to reset
- HelixAgent automatically retries with exponential backoff
- Consider upgrading your plan for higher limits

#### Model Not Found (404)

```
Mistral API error: 404 - {"message": "Model not found"}
```

**Solution:**
- Verify the model name is correct
- Check Mistral documentation for currently available models
- Note: Some models require specific API access

#### Content Filter (400)

```
Mistral API error: 400 - {"message": "Content filtered"}
```

**Solution:**
- Review your prompt for potentially problematic content
- Rephrase the request to be more specific and technical
- Note: Safe prompt mode can be disabled with `SafePrompt: false`

### Health Check

HelixAgent provides a health check endpoint for Mistral:

```go
err := provider.HealthCheck()
if err != nil {
    fmt.Printf("Mistral provider unhealthy: %v\n", err)
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

1. Check network connectivity to `api.mistral.ai`
2. Verify DNS resolution is working
3. Check if the service is accessible from your region
4. Consider increasing timeout (default is 120 seconds)

```go
// Custom retry configuration
retryConfig := mistral.RetryConfig{
    MaxRetries:   5,
    InitialDelay: 2 * time.Second,
    MaxDelay:     60 * time.Second,
    Multiplier:   2.0,
}

provider := mistral.NewMistralProviderWithRetry(
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
4. **Choose the right model** - Use Codestral for code, Large for complex reasoning
5. **Include examples** - Show expected input/output formats

Example improved prompt:

```go
req := &models.LLMRequest{
    ID:     "request-1",
    Prompt: `You are an expert software architect. Write clean, production-ready code.
Follow these guidelines:
- Use proper error handling
- Add meaningful comments
- Follow language-specific best practices
- Include type annotations
- Consider edge cases`,
    Messages: []models.Message{
        {Role: "user", Content: "Design a microservice for handling user authentication with JWT."},
    },
    ModelParams: models.ModelParams{
        MaxTokens:   4096,
        Temperature: 0.2,
    },
}
```

## Unique Features

### Safe Prompt Mode

Mistral offers a safe prompt mode that adds safety guardrails:

```go
// Safe prompt is disabled by default in HelixAgent
// To enable, modify the request in your custom implementation
```

### Mixture of Experts (MoE)

Mistral's Mixtral models use a Mixture of Experts architecture:

- **Efficient inference** - Only activates relevant experts per token
- **High capacity** - 8x7B and 8x22B parameter variants
- **Open weights** - Available for self-hosting

## Additional Resources

- [Mistral AI Console](https://console.mistral.ai)
- [Mistral API Documentation](https://docs.mistral.ai)
- [Mistral Models Overview](https://docs.mistral.ai/getting-started/models/)
- [La Plateforme Pricing](https://mistral.ai/technology/#pricing)
