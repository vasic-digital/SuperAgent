# Zen Provider Setup Guide (OpenCode)

## Overview

OpenCode Zen is a gateway service that provides access to multiple LLM models, including several **free models** that can be used without an API key. This makes it an excellent entry point for developers who want to experiment with AI without upfront costs. HelixAgent integrates with Zen's API to provide access to both free and premium models.

### Supported Models

#### Free Models (No API Key Required)

- `grok-code` - Fast code-focused model (default)
- `big-pickle` - General-purpose model
- `glm-4.7-free` - GLM 4.7 free tier
- `gpt-5-nano` - Lightweight GPT variant

#### Premium Models (API Key Required)

Access to additional models requires an OpenCode API key.

### Key Features

- **Free tier access** - Use free models without API key
- **Anonymous mode** - Device-ID based authentication for free models
- Text completion and chat
- Streaming responses
- Function calling
- Tool use support
- Code completion and analysis
- Large context windows (up to 200K tokens)

## Anonymous Access (Free Models)

One of Zen's unique features is **anonymous access** to free models. No API key is required - just create the provider and start using it.

### Quick Start with Free Models

```go
package main

import (
    "context"
    "fmt"

    "dev.helix.agent/internal/llm/providers/zen"
    "dev.helix.agent/internal/models"
)

func main() {
    // Create anonymous provider - no API key needed!
    provider := zen.NewZenProviderAnonymous("grok-code")

    req := &models.LLMRequest{
        ID:     "request-1",
        Prompt: "You are a helpful coding assistant.",
        Messages: []models.Message{
            {Role: "user", Content: "Write a hello world function in Python."},
        },
        ModelParams: models.ModelParams{
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

## API Key Setup (Premium Models)

For access to premium models, you'll need an OpenCode API key.

### Step 1: Create an OpenCode Account

1. Visit [opencode.ai](https://opencode.ai)
2. Sign up for an account
3. Complete the verification process

### Step 2: Generate an API Key

1. Navigate to your account settings
2. Go to the **API Keys** section
3. Click **Create new API key**
4. Copy the API key immediately - store it securely

### Step 3: Store Your API Key Securely

```bash
# Add to your environment or .env file
export OPENCODE_API_KEY=your_api_key_here
```

## Environment Variable Configuration

Add the following to your `.env` file or environment:

```bash
# Optional for free models, required for premium models
OPENCODE_API_KEY=xxxxxxxxxxxxxxxxxxxxxxxx

# Optional - Override default settings
ZEN_BASE_URL=https://opencode.ai/zen/v1/chat/completions
ZEN_MODEL=grok-code
```

### Configuration Options

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `OPENCODE_API_KEY` | No* | - | Your OpenCode API key (*required for premium models) |
| `ZEN_BASE_URL` | No | `https://opencode.ai/zen/v1/chat/completions` | API endpoint URL |
| `ZEN_MODEL` | No | `grok-code` | Default model to use |

## Basic Usage Examples

### Anonymous Mode (Free Models)

```go
// Create anonymous provider for free models
provider := zen.NewZenProviderAnonymous("grok-code")

// Check if running in anonymous mode
if provider.IsAnonymousMode() {
    fmt.Println("Running in anonymous mode with free model")
}
```

### Authenticated Mode (Premium Models)

```go
// Create authenticated provider
provider := zen.NewZenProvider(
    os.Getenv("OPENCODE_API_KEY"),
    "", // Use default base URL
    "", // Use default model
)
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

### Listing Available Models

```go
// Get all available models
allModels, err := provider.GetAvailableModels(ctx)
if err != nil {
    fmt.Printf("Error: %v\n", err)
    return
}

for _, model := range allModels {
    fmt.Printf("Model: %s (Context: %d tokens)\n", model.ID, model.ContextWindow)
}

// Get only free models
freeModels, err := provider.GetFreeModels(ctx)
if err != nil {
    fmt.Printf("Error: %v\n", err)
    return
}

fmt.Printf("Available free models: %d\n", len(freeModels))
```

## Rate Limits and Quotas

### Free Tier Limits

| Limit Type | Value |
|------------|-------|
| Requests/Day | 100 |
| Tokens/Request | 4,096 |
| Concurrent Requests | 1 |

### Premium Tier Limits

| Plan | Requests/Minute | Tokens/Minute |
|------|-----------------|---------------|
| Standard | 60 | 100,000 |
| Pro | 300 | 500,000 |
| Enterprise | Custom | Custom |

### Model Context Limits

| Model | Max Input Tokens | Max Output Tokens |
|-------|-----------------|-------------------|
| grok-code | 200,000 | 16,384 |
| big-pickle | 200,000 | 16,384 |
| glm-4.7-free | 128,000 | 16,384 |
| gpt-5-nano | 128,000 | 16,384 |

### Best Practices for Rate Limits

1. **Use exponential backoff** - HelixAgent automatically implements retry with backoff
2. **Monitor token usage** - Track consumption through response metadata
3. **Start with free models** - Test your implementation before upgrading
4. **Batch requests thoughtfully** - Combine related operations when possible

## Model Selection Guide

| Use Case | Recommended Model | Reason |
|----------|-------------------|--------|
| Code generation | `grok-code` | Optimized for programming tasks |
| General tasks | `big-pickle` | Well-rounded performance |
| Chinese language | `glm-4.7-free` | Based on GLM model family |
| Low latency | `gpt-5-nano` | Lightweight and fast |

## Troubleshooting

### Common Errors

#### Anonymous Access Denied

```
Zen API error: 401 - {"error": {"message": "Anonymous access denied for this model"}}
```

**Solution:**
- Verify you're using a free model (`grok-code`, `big-pickle`, `glm-4.7-free`, `gpt-5-nano`)
- For premium models, provide an OPENCODE_API_KEY
- Check the model name doesn't have the "opencode/" prefix (it's stripped automatically)

#### Rate Limit Error (429)

```
Zen API error: 429 - {"error": {"message": "Rate limit exceeded"}}
```

**Solution:**
- Wait for the rate limit window to reset
- HelixAgent automatically retries with exponential backoff
- Consider upgrading to a premium plan for higher limits
- Free tier has strict daily limits

#### Model Not Found (404)

```
Zen API error: 404 - {"error": {"message": "Model not found"}}
```

**Solution:**
- Verify the model name is correct (use names without "opencode/" prefix)
- Check available models using `GetAvailableModels()`
- Some models may be temporarily unavailable

#### No Choices Returned

```
Zen API returned no choices
```

**Solution:**
- Check if your prompt is valid and not empty
- Verify the model is available and functioning
- Review your request parameters

### Health Check

HelixAgent provides a health check endpoint for Zen:

```go
err := provider.HealthCheck()
if err != nil {
    fmt.Printf("Zen provider unhealthy: %v\n", err)
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

1. Check network connectivity to `opencode.ai`
2. Verify DNS resolution is working
3. Check if the service is accessible from your region
4. Consider increasing timeout (default is 120 seconds)

```go
// Custom retry configuration
retryConfig := zen.RetryConfig{
    MaxRetries:   5,
    InitialDelay: 2 * time.Second,
    MaxDelay:     60 * time.Second,
    Multiplier:   2.0,
}

provider := zen.NewZenProviderWithRetry(
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
4. **Try different models** - Each free model has different strengths
5. **Include examples** - Show expected input/output formats

Example improved prompt:

```go
req := &models.LLMRequest{
    ID:     "request-1",
    Prompt: `You are an expert programmer. Write clean, well-documented code.
Follow these guidelines:
- Use proper error handling
- Add meaningful comments
- Follow language-specific best practices
- Consider edge cases`,
    Messages: []models.Message{
        {Role: "user", Content: "Create a function to validate email addresses in JavaScript."},
    },
    ModelParams: models.ModelParams{
        MaxTokens:   2048,
        Temperature: 0.2,
    },
}
```

## Unique Features

### Anonymous Mode

Zen's anonymous mode is unique among LLM providers:

```go
// Check available free models
freeModels := zen.FreeModels()
// Returns: ["big-pickle", "grok-code", "glm-4.7-free", "gpt-5-nano"]

// Check if a model allows anonymous access
if zen.IsAnonymousAccessAllowed("grok-code") {
    provider := zen.NewZenProviderAnonymous("grok-code")
}
```

### Device ID Authentication

In anonymous mode, Zen uses a device ID for tracking:

```go
// Device ID is automatically generated
provider := zen.NewZenProviderAnonymous("grok-code")
// Internal header: X-Device-ID: helix-xxxxxxxxxxxx
```

### Model ID Normalization

Zen automatically handles different model ID formats:

```go
// All of these work:
provider := zen.NewZenProvider(apiKey, "", "grok-code")
provider := zen.NewZenProvider(apiKey, "", "opencode/grok-code")
provider := zen.NewZenProvider(apiKey, "", "opencode-grok-code")
// They all normalize to "grok-code"
```

## Integration with HelixAgent

Zen is integrated into HelixAgent's unified verification pipeline as a **free provider**:

- Score range: 6.0 - 7.0 (free tier baseline)
- Used as fallback when premium providers are unavailable
- Ideal for development and testing without API costs

### Fallback Strategy

When premium providers fail, Zen free models provide reliable fallback:

```
Primary: Claude/DeepSeek/Gemini
    ↓ (on failure)
Fallback: Zen free models (grok-code, big-pickle)
```

## Additional Resources

- [OpenCode Platform](https://opencode.ai)
- [Zen API Documentation](https://opencode.ai/docs)
- [Available Models List](https://opencode.ai/zen/v1/models)
