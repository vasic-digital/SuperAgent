# User Manual 22: Custom Provider Guide

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [The LLMProvider Interface](#the-llmprovider-interface)
4. [Step 1: Create the Provider Package](#step-1-create-the-provider-package)
5. [Step 2: Implement the LLMProvider Interface](#step-2-implement-the-llmprovider-interface)
6. [Step 3: Add Streaming Support](#step-3-add-streaming-support)
7. [Step 4: Implement Health Checks](#step-4-implement-health-checks)
8. [Step 5: Define Capabilities](#step-5-define-capabilities)
9. [Step 6: Add Configuration Validation](#step-6-add-configuration-validation)
10. [Step 7: Register the Provider](#step-7-register-the-provider)
11. [Step 8: Add Environment Variables](#step-8-add-environment-variables)
12. [Step 9: Add Tool Support (Optional)](#step-9-add-tool-support-optional)
13. [Step 10: Integrate with Model Discovery](#step-10-integrate-with-model-discovery)
14. [Provider Architecture](#provider-architecture)
15. [Using the Generic Provider](#using-the-generic-provider)
16. [Testing Your Provider](#testing-your-provider)
17. [Troubleshooting](#troubleshooting)
18. [Related Resources](#related-resources)

## Overview

HelixAgent supports 22+ dedicated LLM providers and a generic OpenAI-compatible provider for additional services. This guide covers implementing a new dedicated provider from scratch, registering it with the provider registry, adding it to the verification pipeline, and testing it end to end.

If your target service exposes an OpenAI-compatible API, you may not need a dedicated provider at all. See the [Using the Generic Provider](#using-the-generic-provider) section for that path.

## Prerequisites

- Go 1.24+ development environment
- Access to the target LLM provider API documentation
- An API key for the target provider
- Familiarity with the HelixAgent codebase structure

## The LLMProvider Interface

Every provider must implement the `LLMProvider` interface defined in the core models:

```go
type LLMProvider interface {
    // Complete sends a chat completion request and returns a response.
    Complete(ctx context.Context, req *LLMRequest) (*LLMResponse, error)

    // CompleteStream sends a streaming chat completion request.
    CompleteStream(ctx context.Context, req *LLMRequest) (<-chan StreamChunk, error)

    // HealthCheck verifies the provider is reachable and functional.
    HealthCheck(ctx context.Context) error

    // GetCapabilities returns the provider's feature support matrix.
    GetCapabilities() ProviderCapabilities

    // ValidateConfig checks that the provider configuration is valid.
    ValidateConfig() error
}
```

## Step 1: Create the Provider Package

Create a new directory under `internal/llm/providers/`:

```bash
mkdir -p internal/llm/providers/myprovider
```

Create the main provider file:

```go
// internal/llm/providers/myprovider/myprovider.go
package myprovider

import (
    "context"
    "fmt"
    "net/http"
    "time"

    "dev.helix.agent/internal/models"
)

const (
    defaultBaseURL = "https://api.myprovider.com/v1"
    defaultTimeout = 30 * time.Second
    providerName   = "myprovider"
)

// Provider implements the LLMProvider interface for MyProvider.
type Provider struct {
    apiKey  string
    baseURL string
    client  *http.Client
    model   string
}

// New creates a new MyProvider instance.
func New(apiKey string) *Provider {
    return &Provider{
        apiKey:  apiKey,
        baseURL: defaultBaseURL,
        client:  &http.Client{Timeout: defaultTimeout},
        model:   "myprovider-default",
    }
}

// NewWithConfig creates a provider with custom configuration.
func NewWithConfig(config Config) *Provider {
    baseURL := config.BaseURL
    if baseURL == "" {
        baseURL = defaultBaseURL
    }
    timeout := config.Timeout
    if timeout == 0 {
        timeout = defaultTimeout
    }
    return &Provider{
        apiKey:  config.APIKey,
        baseURL: baseURL,
        client:  &http.Client{Timeout: timeout},
        model:   config.Model,
    }
}

// Config holds provider configuration.
type Config struct {
    APIKey  string
    BaseURL string
    Model   string
    Timeout time.Duration
}
```

## Step 2: Implement the LLMProvider Interface

### Complete Method

```go
func (p *Provider) Complete(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
    if err := p.ValidateConfig(); err != nil {
        return nil, fmt.Errorf("config validation failed: %w", err)
    }

    // Build the API request body
    apiReq := buildAPIRequest(req, p.model)

    body, err := json.Marshal(apiReq)
    if err != nil {
        return nil, fmt.Errorf("marshal request: %w", err)
    }

    httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
        p.baseURL+"/chat/completions", bytes.NewReader(body))
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }

    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

    resp, err := p.client.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("provider request failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        bodyBytes, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("provider returned %d: %s", resp.StatusCode, string(bodyBytes))
    }

    var apiResp APIResponse
    if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
        return nil, fmt.Errorf("decode response: %w", err)
    }

    return convertToLLMResponse(&apiResp), nil
}
```

### Helper Types

```go
type APIRequest struct {
    Model    string       `json:"model"`
    Messages []APIMessage `json:"messages"`
    MaxTokens int        `json:"max_tokens,omitempty"`
}

type APIMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type APIResponse struct {
    ID      string     `json:"id"`
    Choices []Choice   `json:"choices"`
    Usage   UsageStats `json:"usage"`
}

type Choice struct {
    Index   int        `json:"index"`
    Message APIMessage `json:"message"`
}

type UsageStats struct {
    PromptTokens     int `json:"prompt_tokens"`
    CompletionTokens int `json:"completion_tokens"`
    TotalTokens      int `json:"total_tokens"`
}
```

## Step 3: Add Streaming Support

```go
func (p *Provider) CompleteStream(ctx context.Context, req *models.LLMRequest) (<-chan models.StreamChunk, error) {
    if err := p.ValidateConfig(); err != nil {
        return nil, fmt.Errorf("config validation failed: %w", err)
    }

    apiReq := buildAPIRequest(req, p.model)
    apiReq.Stream = true

    body, err := json.Marshal(apiReq)
    if err != nil {
        return nil, fmt.Errorf("marshal request: %w", err)
    }

    httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
        p.baseURL+"/chat/completions", bytes.NewReader(body))
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }

    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
    httpReq.Header.Set("Accept", "text/event-stream")

    resp, err := p.client.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("stream request failed: %w", err)
    }

    ch := make(chan models.StreamChunk, 100)
    go func() {
        defer close(ch)
        defer resp.Body.Close()

        scanner := bufio.NewScanner(resp.Body)
        for scanner.Scan() {
            line := scanner.Text()
            if !strings.HasPrefix(line, "data: ") {
                continue
            }
            data := strings.TrimPrefix(line, "data: ")
            if data == "[DONE]" {
                return
            }

            var chunk StreamResponse
            if err := json.Unmarshal([]byte(data), &chunk); err != nil {
                continue
            }

            ch <- models.StreamChunk{
                Content: chunk.Choices[0].Delta.Content,
                Done:    false,
            }
        }
    }()

    return ch, nil
}
```

## Step 4: Implement Health Checks

```go
func (p *Provider) HealthCheck(ctx context.Context) error {
    httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet,
        p.baseURL+"/models", nil)
    if err != nil {
        return fmt.Errorf("create health check request: %w", err)
    }

    httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

    resp, err := p.client.Do(httpReq)
    if err != nil {
        return fmt.Errorf("health check failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("health check returned status %d", resp.StatusCode)
    }
    return nil
}
```

## Step 5: Define Capabilities

```go
func (p *Provider) GetCapabilities() models.ProviderCapabilities {
    return models.ProviderCapabilities{
        SupportsStreaming:   true,
        SupportsTools:      false, // Set true if provider supports function calling
        SupportsFunctions:  false,
        SupportsVision:     false,
        SupportsEmbeddings: false,
        MaxContextLength:   128000,
        Models: []string{
            "myprovider-small",
            "myprovider-large",
        },
    }
}
```

## Step 6: Add Configuration Validation

```go
func (p *Provider) ValidateConfig() error {
    if p.apiKey == "" {
        return fmt.Errorf("MYPROVIDER_API_KEY is required")
    }
    if p.baseURL == "" {
        return fmt.Errorf("base URL is required")
    }
    return nil
}
```

## Step 7: Register the Provider

Add your provider to the provider registry in `internal/services/provider_registry.go`:

```go
func (r *ProviderRegistry) registerDefaultProviders() {
    // ... existing providers ...

    r.Register("myprovider", func(config map[string]string) (LLMProvider, error) {
        apiKey := config["api_key"]
        if apiKey == "" {
            return nil, fmt.Errorf("myprovider: API key is required")
        }
        return myprovider.New(apiKey), nil
    })
}
```

## Step 8: Add Environment Variables

Add to `.env.example`:

```bash
# MyProvider
MYPROVIDER_API_KEY=your_api_key_here
MYPROVIDER_BASE_URL=https://api.myprovider.com/v1  # optional
```

Update the configuration loader in `internal/config/config.go` to read the new variables.

## Step 9: Add Tool Support (Optional)

If your provider supports function calling / tool use:

```go
func (p *Provider) GetCapabilities() models.ProviderCapabilities {
    return models.ProviderCapabilities{
        SupportsTools: true,
        // ...
    }
}

func buildAPIRequest(req *models.LLMRequest, model string) *APIRequest {
    apiReq := &APIRequest{
        Model:    model,
        Messages: convertMessages(req.Messages),
    }

    if len(req.Tools) > 0 {
        apiReq.Tools = convertTools(req.Tools)
        apiReq.ToolChoice = "auto"
    }

    return apiReq
}
```

Ensure tool schemas follow the HelixAgent convention (all parameters use **snake_case**).

## Step 10: Integrate with Model Discovery

Add your provider to the 3-tier model discovery system in `internal/llm/discovery/`:

```go
// Tier 3: Hardcoded fallback models
func getHardcodedModels(provider string) []string {
    switch provider {
    // ... existing providers ...
    case "myprovider":
        return []string{"myprovider-small", "myprovider-large"}
    }
    return nil
}
```

For Tier 1 (API discovery), implement a response parser if the provider does not follow the OpenAI `/v1/models` format.

## Provider Architecture

```
internal/llm/providers/
+-- myprovider/
|   +-- myprovider.go         # Main provider implementation
|   +-- myprovider_test.go    # Unit tests
|   +-- types.go              # API request/response types
|   +-- helpers.go            # Conversion helpers
+-- claude/                   # Reference implementation (OAuth/CLI)
+-- deepseek/                 # Reference implementation (API key)
+-- openai/                   # Reference implementation (full features)
+-- generic/                  # Generic OpenAI-compatible provider

internal/services/
+-- provider_registry.go      # Provider registration

internal/llm/discovery/       # 3-tier model discovery

internal/verifier/            # Startup verification pipeline
+-- startup.go
+-- provider_types.go
```

## Using the Generic Provider

If the target service exposes an OpenAI-compatible API (same request/response format, same endpoints), you can use the generic provider without writing any Go code:

```bash
# In .env, add:
MYPROVIDER_API_KEY=your_key
MYPROVIDER_BASE_URL=https://api.myprovider.com/v1
```

The generic provider in `internal/llm/providers/generic/` handles OpenAI-compatible services for 17+ additional providers (NVIDIA, SambaNova, Hyperbolic, Novita, SiliconFlow, Kimi, Upstage, and others).

## Testing Your Provider

### Unit Tests

```go
// internal/llm/providers/myprovider/myprovider_test.go
package myprovider

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestProvider_Complete_Success(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(`{
            "id": "resp-1",
            "choices": [{"index": 0, "message": {"role": "assistant", "content": "Hello"}}],
            "usage": {"prompt_tokens": 5, "completion_tokens": 1, "total_tokens": 6}
        }`))
    }))
    defer server.Close()

    p := &Provider{
        apiKey:  "test-key",
        baseURL: server.URL,
        client:  &http.Client{},
        model:   "test-model",
    }

    resp, err := p.Complete(context.Background(), &models.LLMRequest{
        Messages: []models.Message{{Role: "user", Content: "Hi"}},
    })

    require.NoError(t, err)
    assert.Equal(t, "Hello", resp.Content)
}

func TestProvider_ValidateConfig_MissingKey(t *testing.T) {
    p := &Provider{apiKey: "", baseURL: defaultBaseURL}
    err := p.ValidateConfig()
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "API_KEY is required")
}

func TestProvider_HealthCheck_Success(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, "/models", r.URL.Path)
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"data": []}`))
    }))
    defer server.Close()

    p := &Provider{apiKey: "test-key", baseURL: server.URL, client: &http.Client{}}
    err := p.HealthCheck(context.Background())
    assert.NoError(t, err)
}
```

### Integration Test

```bash
# With real API key
MYPROVIDER_API_KEY=real_key go test -v -run TestProvider ./internal/llm/providers/myprovider/...
```

### Challenge

Create `challenges/scripts/myprovider_challenge.sh` to validate end-to-end functionality.

## Troubleshooting

### Provider Returns Empty Responses

**Symptom:** `Complete` returns nil content.

**Solutions:**
1. Check the API response format matches your `APIResponse` struct
2. Log the raw response body before parsing to inspect actual format
3. Verify the model name is valid for the provider
4. Check if the provider requires specific request headers

### Provider Fails Verification Pipeline

**Symptom:** Provider scores below 5.0 during startup.

**Solutions:**
1. Verify the health check endpoint returns 200
2. Ensure response times are reasonable (< 5s for verification)
3. Check API key validity and rate limits
4. Review `internal/verifier/startup.go` for scoring criteria

### Streaming Produces Garbled Output

**Symptom:** Stream chunks are malformed or incomplete.

**Solutions:**
1. Verify the SSE format matches `data: {json}\n\n`
2. Check for providers that use non-standard streaming formats
3. Ensure the scanner has a large enough buffer for long responses
4. Handle the `[DONE]` sentinel correctly

## Related Resources

- [User Manual 20: Testing Strategies](20-testing-strategies.md) -- Testing the provider
- [User Manual 23: Observability Setup](23-observability-setup.md) -- Monitoring provider health
- Existing providers: `internal/llm/providers/`
- Generic provider: `internal/llm/providers/generic/`
- Provider registry: `internal/services/provider_registry.go`
- Model discovery: `internal/llm/discovery/`
- Verification pipeline: `internal/verifier/`
