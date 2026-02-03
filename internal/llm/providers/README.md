# LLM Provider Abstraction Layer

This package provides a unified interface for integrating multiple Large Language Model providers into HelixAgent. It handles provider registration, lifecycle management, health monitoring, and circuit breaking.

## Overview

The provider abstraction layer enables HelixAgent to:
- Support 20+ LLM providers through a common interface
- Dynamically select providers based on verification scores
- Automatically handle fallbacks when providers fail
- Support both API key and OAuth authentication
- Run ensemble queries across multiple providers

## Architecture

```
                          +------------------------+
                          |   ProviderRegistry     |
                          | (Provider Management)  |
                          +----------+-------------+
                                     |
           +-------------------------+-------------------------+
           |                         |                         |
           v                         v                         v
+----------+---------+   +-----------+-----------+   +--------+----------+
| CircuitBreaker     |   | ConcurrencySemaphore  |   | HealthMonitor     |
| (Fault Tolerance)  |   | (Rate Limiting)       |   | (Health Checks)   |
+----------+---------+   +-----------+-----------+   +--------+----------+
           |                         |                         |
           +-------------------------+-------------------------+
                                     |
                                     v
                          +----------+-----------+
                          |    LLMProvider       |
                          |    Interface         |
                          +----------+-----------+
                                     |
     +-------------------------------+-------------------------------+
     |           |           |           |           |               |
     v           v           v           v           v               v
+----+----+ +----+----+ +----+----+ +----+----+ +----+----+    +-----+-----+
| Claude  | | DeepSeek| | Gemini  | |  Qwen   | | Mistral |...| Ollama    |
+---------+ +---------+ +---------+ +---------+ +---------+    +-----------+
```

## LLMProvider Interface

All providers implement this core interface:

```go
type LLMProvider interface {
    // Complete sends a request and returns a response
    Complete(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error)

    // CompleteStream returns a channel of streaming responses
    CompleteStream(ctx context.Context, req *models.LLMRequest) (<-chan *models.LLMResponse, error)

    // HealthCheck verifies provider connectivity
    HealthCheck() error

    // GetCapabilities returns provider feature support
    GetCapabilities() *models.ProviderCapabilities

    // ValidateConfig validates provider configuration
    ValidateConfig(config map[string]interface{}) (bool, []string)
}
```

## Provider Registry

The `ProviderRegistry` manages all provider instances:

```go
import "dev.helix.agent/internal/services"

registry := services.NewProviderRegistry(config, memoryService)

// Register a provider
err := registry.RegisterProvider("claude", claudeProvider)

// Get a provider
provider, err := registry.GetProvider("claude")

// List all providers
providers := registry.ListProviders()

// Get providers ordered by LLMsVerifier score
orderedProviders := registry.ListProvidersOrderedByScore()
```

## Provider Lifecycle

### 1. Registration

```go
// Direct registration
err := registry.RegisterProvider("my-provider", provider)

// Configuration-based registration
config := services.ProviderConfig{
    Name:    "claude",
    Type:    "claude",
    Enabled: true,
    APIKey:  os.Getenv("CLAUDE_API_KEY"),
    Models: []services.ModelConfig{{
        ID:      "claude-3-opus",
        Enabled: true,
    }},
}
err := registry.RegisterProviderFromConfig(config)
```

### 2. Auto-Discovery

Providers are automatically discovered from environment variables:

```go
// Auto-discovery reads these env vars:
// ANTHROPIC_API_KEY, CLAUDE_API_KEY
// DEEPSEEK_API_KEY
// GEMINI_API_KEY
// QWEN_API_KEY
// OPENROUTER_API_KEY
// MISTRAL_API_KEY
// ZAI_API_KEY
// OPENCODE_API_KEY (for Zen)
// CEREBRAS_API_KEY
// ... and more

registry := services.NewProviderRegistry(config, memory)
// Providers with valid API keys are automatically registered
```

### 3. Verification

```go
// Verify a single provider
result := registry.VerifyProvider(ctx, "claude")
// Returns: {Status, Verified, Score, ResponseTime, Error}

// Verify all providers
results := registry.VerifyAllProviders(ctx)

// Get healthy providers
healthy := registry.GetHealthyProviders()
```

### 4. Unregistration

```go
// Graceful removal (waits for active requests)
err := registry.RemoveProvider("claude", false)

// Force removal (immediate)
err := registry.RemoveProvider("claude", true)
```

## How to Implement a New Provider

### Step 1: Create Provider Package

```go
// internal/llm/providers/newprovider/newprovider.go
package newprovider

import (
    "context"
    "dev.helix.agent/internal/llm"
    "dev.helix.agent/internal/models"
)

type NewProvider struct {
    apiKey  string
    baseURL string
    model   string
}

func NewNewProvider(apiKey, baseURL, model string) *NewProvider {
    return &NewProvider{
        apiKey:  apiKey,
        baseURL: baseURL,
        model:   model,
    }
}
```

### Step 2: Implement LLMProvider Interface

```go
func (p *NewProvider) Complete(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
    // 1. Build API request
    // 2. Send request to provider API
    // 3. Parse response
    // 4. Return models.LLMResponse
}

func (p *NewProvider) CompleteStream(ctx context.Context, req *models.LLMRequest) (<-chan *models.LLMResponse, error) {
    // 1. Build streaming request
    // 2. Return channel that yields responses
}

func (p *NewProvider) HealthCheck() error {
    // Test connectivity (e.g., list models)
    return nil
}

func (p *NewProvider) GetCapabilities() *models.ProviderCapabilities {
    return &models.ProviderCapabilities{
        SupportsStreaming:    true,
        SupportsTools:        true,
        SupportsVision:       false,
        MaxTokens:            4096,
        SupportedModalities:  []string{"text"},
        SupportedLanguages:   []string{"en"},
    }
}

func (p *NewProvider) ValidateConfig(config map[string]interface{}) (bool, []string) {
    var errors []string
    if _, ok := config["api_key"]; !ok {
        errors = append(errors, "api_key is required")
    }
    return len(errors) == 0, errors
}
```

### Step 3: Register in Provider Registry

Add to `internal/services/provider_registry.go`:

```go
import "dev.helix.agent/internal/llm/providers/newprovider"

// In createProviderFromConfig:
case "newprovider":
    if cfg.Enabled && cfg.APIKey != "" {
        return newprovider.NewNewProvider(cfg.APIKey, baseURL, model), nil
    }
    return nil, fmt.Errorf("NewProvider not available: API key missing")
```

### Step 4: Add Environment Variable Mapping

Add to `internal/services/provider_discovery.go`:

```go
var providerMappings = []ProviderMapping{
    // ... existing mappings
    {
        EnvKey:       "NEWPROVIDER_API_KEY",
        ProviderType: "newprovider",
        ProviderName: "newprovider",
        CreateFunc: func(apiKey, baseURL, model string) llm.LLMProvider {
            return newprovider.NewNewProvider(apiKey, baseURL, model)
        },
        DefaultModel: "newprovider-default",
    },
}
```

### Step 5: Add Tests

```go
// internal/llm/providers/newprovider/newprovider_test.go
package newprovider

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestNewProvider_Complete(t *testing.T) {
    provider := NewNewProvider("test-key", "", "model")

    req := &models.LLMRequest{
        Prompt: "Hello",
    }

    resp, err := provider.Complete(context.Background(), req)
    assert.NoError(t, err)
    assert.NotEmpty(t, resp.Content)
}
```

## Supported Providers (20)

### API Key Providers

| Provider | Package | Env Variable | Default Model |
|----------|---------|--------------|---------------|
| Claude | `claude` | `ANTHROPIC_API_KEY` | `claude-3-sonnet-20240229` |
| DeepSeek | `deepseek` | `DEEPSEEK_API_KEY` | `deepseek-coder` |
| Gemini | `gemini` | `GEMINI_API_KEY` | `gemini-pro` |
| Mistral | `mistral` | `MISTRAL_API_KEY` | `mistral-large-latest` |
| OpenRouter | `openrouter` | `OPENROUTER_API_KEY` | `anthropic/claude-3` |
| ZAI | `zai` | `ZAI_API_KEY` | `zai-default` |
| Cerebras | `cerebras` | `CEREBRAS_API_KEY` | `cerebras-gpt` |
| OpenAI | `openai` | `OPENAI_API_KEY` | `gpt-4` |
| Cohere | `cohere` | `COHERE_API_KEY` | `command` |
| Fireworks | `fireworks` | `FIREWORKS_API_KEY` | `accounts/fireworks/models/llama-v2-7b` |
| Groq | `groq` | `GROQ_API_KEY` | `mixtral-8x7b-32768` |
| HuggingFace | `huggingface` | `HUGGINGFACE_API_KEY` | `meta-llama/Llama-2-7b-chat-hf` |
| Perplexity | `perplexity` | `PERPLEXITY_API_KEY` | `pplx-7b-online` |
| Replicate | `replicate` | `REPLICATE_API_KEY` | `meta/llama-2-70b-chat` |
| Together | `together` | `TOGETHER_API_KEY` | `mistralai/Mixtral-8x7B-Instruct-v0.1` |
| XAI | `xai` | `XAI_API_KEY` | `grok-1` |
| AI21 | `ai21` | `AI21_API_KEY` | `j2-ultra` |
| Anthropic | `anthropic` | `ANTHROPIC_API_KEY` | `claude-3-sonnet-20240229` |

### OAuth Providers

| Provider | Package | OAuth Method | Fallback |
|----------|---------|--------------|----------|
| Claude | `claude` | `claude auth login` | CLI proxy |
| Qwen | `qwen` | ACP Protocol | CLI proxy |

### Free/Local Providers

| Provider | Package | Configuration |
|----------|---------|---------------|
| Zen | `zen` | `OPENCODE_API_KEY` (optional) |
| Ollama | `ollama` | `OLLAMA_BASE_URL` |

## Provider Features

### Tool Calling Support

```go
caps := provider.GetCapabilities()
if caps.SupportsTools {
    req := &models.LLMRequest{
        Tools: []models.Tool{
            {
                Name:        "search",
                Description: "Search the web",
                Parameters:  schema,
            },
        },
    }
}
```

### Streaming Support

```go
if caps.SupportsStreaming {
    stream, err := provider.CompleteStream(ctx, req)
    for chunk := range stream {
        fmt.Print(chunk.Content)
    }
}
```

### Vision Support

```go
if caps.SupportsVision {
    req := &models.LLMRequest{
        Messages: []models.Message{
            {
                Role:    "user",
                Content: "What's in this image?",
                Images:  []string{base64Image},
            },
        },
    }
}
```

## Configuration

### Provider Config

```go
type ProviderConfig struct {
    Name                  string            `json:"name"`
    Type                  string            `json:"type"`
    Enabled               bool              `json:"enabled"`
    APIKey                string            `json:"api_key"`
    BaseURL               string            `json:"base_url"`
    Models                []ModelConfig     `json:"models"`
    Timeout               time.Duration     `json:"timeout"`
    MaxRetries            int               `json:"max_retries"`
    MaxConcurrentRequests int               `json:"max_concurrent_requests"`
    HealthCheckURL        string            `json:"health_check_url"`
    Weight                float64           `json:"weight"`
    Tags                  []string          `json:"tags"`
    Capabilities          map[string]string `json:"capabilities"`
}
```

### Environment Variables

```bash
# API Keys
export ANTHROPIC_API_KEY="sk-ant-..."
export DEEPSEEK_API_KEY="sk-..."
export GEMINI_API_KEY="..."
export QWEN_API_KEY="sk-..."
export OPENROUTER_API_KEY="sk-or-..."
export MISTRAL_API_KEY="..."
export CEREBRAS_API_KEY="..."

# OAuth settings
export CLAUDE_USE_OAUTH_CREDENTIALS=true
export QWEN_USE_OAUTH_CREDENTIALS=true

# Custom endpoints
export DEEPSEEK_BASE_URL="https://api.deepseek.com"
export OLLAMA_BASE_URL="http://localhost:11434"
```

## Health and Monitoring

### Health Status

```go
const (
    ProviderStatusUnknown     = "unknown"
    ProviderStatusHealthy     = "healthy"
    ProviderStatusRateLimited = "rate_limited"
    ProviderStatusAuthFailed  = "auth_failed"
    ProviderStatusUnhealthy   = "unhealthy"
)
```

### Concurrency Stats

```go
stats, err := registry.GetConcurrencyStats("claude")
// Returns:
// - TotalPermits: 10
// - AcquiredPermits: 3
// - ActiveRequests: 3
// - AvailablePermits: 7
```

### LLMsVerifier Integration

The registry integrates with LLMsVerifier for dynamic provider scoring:

```go
// Providers are ordered by verification score
ordered := registry.ListProvidersOrderedByScore()

// Scores are updated after verification
registry.UpdateProviderScore("claude", "claude-3-opus", 8.5)

// Get score for a provider
score, found := registry.GetScoreAdapter().GetProviderScore("claude")
```

## Error Handling

### Error Categories

- `rate_limit`: Provider rate limit exceeded (429)
- `timeout`: Request timed out
- `auth`: Authentication failed (401, 403)
- `connection`: Network connectivity issues
- `unavailable`: Service unavailable (503)
- `overloaded`: Provider overloaded (529)

### Circuit Breaker

```go
// Circuit breaker protects against cascading failures
cb := registry.GetCircuitBreaker("claude")
// States: Closed -> Open -> Half-Open -> Closed

// Configuration
circuitConfig := CircuitBreakerConfig{
    Enabled:          true,
    FailureThreshold: 5,
    RecoveryTimeout:  60 * time.Second,
    SuccessThreshold: 2,
}
```

## Testing

```bash
# Run all provider tests
go test -v ./internal/llm/providers/...

# Test specific provider
go test -v ./internal/llm/providers/claude/...

# Integration tests (requires API keys)
ANTHROPIC_API_KEY=... go test -v -tags=integration ./internal/llm/providers/claude/...
```

## Files

| Directory/File | Description |
|----------------|-------------|
| `claude/` | Anthropic Claude provider |
| `deepseek/` | DeepSeek provider |
| `gemini/` | Google Gemini provider |
| `mistral/` | Mistral AI provider |
| `openrouter/` | OpenRouter aggregator |
| `qwen/` | Alibaba Qwen/DashScope |
| `zai/` | ZAI provider |
| `zen/` | Zen (OpenCode) provider |
| `cerebras/` | Cerebras provider |
| `ollama/` | Local Ollama models |
| `openai/` | OpenAI provider |
| `anthropic/` | Anthropic (alternative) |
| `cohere/` | Cohere provider |
| `fireworks/` | Fireworks AI |
| `groq/` | Groq provider |
| `huggingface/` | HuggingFace Inference |
| `perplexity/` | Perplexity AI |
| `replicate/` | Replicate provider |
| `together/` | Together AI |
| `xai/` | xAI (Grok) |
| `ai21/` | AI21 Labs |

## Related Packages

- `internal/llm/` - Core LLM package with interface definitions
- `internal/services/provider_registry.go` - Provider registration and management
- `internal/services/provider_discovery.go` - Auto-discovery from environment
- `internal/verifier/` - LLMsVerifier integration
- `internal/models/` - Request/response types
