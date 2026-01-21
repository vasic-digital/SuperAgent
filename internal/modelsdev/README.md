# Modelsdev Package

The modelsdev package provides integration with the **Models.dev** API for retrieving and caching LLM model metadata in HelixAgent.

## Overview

This package implements a caching service that fetches model information from Models.dev, a comprehensive directory of LLM models. It provides efficient access to model capabilities, pricing, context windows, and other metadata to support intelligent model selection.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Models.dev Service                        │
│                                                              │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                      Cache Layer                         ││
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐ ││
│  │  │   Models    │  │  Providers  │  │   Capabilities  │ ││
│  │  │   Cache     │  │    Cache    │  │     Index       │ ││
│  │  └─────────────┘  └─────────────┘  └─────────────────┘ ││
│  └─────────────────────────────────────────────────────────┘│
│                           │                                  │
│  ┌────────────────────────▼─────────────────────────────┐  │
│  │                   API Client                           │  │
│  │  HTTP Client | Rate Limiting | Retry Logic            │  │
│  └──────────────────────────────────────────────────────┘  │
│                           │                                  │
│  ┌────────────────────────▼─────────────────────────────┐  │
│  │                Background Refresh                      │  │
│  │  Periodic Updates | Cache Invalidation | Metrics      │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## Key Types

### Service

The main Models.dev service.

```go
type Service struct {
    client      *http.Client
    baseURL     string
    cache       *ModelCache
    refresher   *BackgroundRefresher
    config      ServiceConfig
    metrics     *ServiceMetrics
}
```

### Model

Represents an LLM model with its metadata.

```go
type Model struct {
    ID              string            // Model identifier
    Name            string            // Display name
    Provider        string            // Provider (anthropic, openai, etc.)
    Version         string            // Model version
    Description     string            // Model description
    ContextWindow   int               // Max context tokens
    MaxOutputTokens int               // Max output tokens
    Capabilities    []string          // Model capabilities
    Pricing         *ModelPricing     // Pricing information
    ReleaseDate     time.Time         // When released
    Deprecated      bool              // Is deprecated
    Metadata        map[string]interface{}
}
```

### ModelPricing

Pricing information for a model.

```go
type ModelPricing struct {
    InputPerMillion  float64 // Cost per million input tokens
    OutputPerMillion float64 // Cost per million output tokens
    Currency         string  // Currency code (USD)
}
```

### Provider

Represents an LLM provider.

```go
type Provider struct {
    ID          string   // Provider identifier
    Name        string   // Display name
    Website     string   // Provider website
    APIBase     string   // API base URL
    Models      []string // Available model IDs
    AuthType    string   // Authentication type
    RateLimits  *RateLimits
}

type RateLimits struct {
    RequestsPerMinute int
    TokensPerMinute   int
}
```

### ModelCache

In-memory cache for model data.

```go
type ModelCache struct {
    models    map[string]*Model
    providers map[string]*Provider
    byCapability map[string][]string // capability -> model IDs
    mu        sync.RWMutex
    ttl       time.Duration
    lastRefresh time.Time
}
```

## Configuration

```go
type ServiceConfig struct {
    // API settings
    BaseURL         string        // Models.dev API URL
    APIKey          string        // Optional API key

    // Cache settings
    CacheTTL        time.Duration // Cache time-to-live
    RefreshInterval time.Duration // Background refresh interval
    MaxCacheSize    int           // Maximum cached models

    // HTTP settings
    RequestTimeout  time.Duration // API request timeout
    MaxRetries      int           // Max retry attempts

    // Features
    EnableBackground bool         // Enable background refresh
    EnableMetrics    bool         // Enable metrics collection
}
```

## Usage Examples

### Initialize Service

```go
import "dev.helix.agent/internal/modelsdev"

// Create service with default config
service := modelsdev.NewService(modelsdev.ServiceConfig{
    BaseURL:         "https://models.dev/api/v1",
    CacheTTL:        1 * time.Hour,
    RefreshInterval: 30 * time.Minute,
    EnableBackground: true,
})

// Start background refresh
service.Start()
defer service.Stop()
```

### Get Model Information

```go
// Get a specific model
model, err := service.GetModel(ctx, "claude-3-opus-20240229")
if err != nil {
    return err
}

fmt.Printf("Model: %s\n", model.Name)
fmt.Printf("Context: %d tokens\n", model.ContextWindow)
fmt.Printf("Price: $%.2f/M input, $%.2f/M output\n",
    model.Pricing.InputPerMillion,
    model.Pricing.OutputPerMillion)
```

### List Models

```go
// List all models
models, err := service.ListModels(ctx)
if err != nil {
    return err
}

for _, model := range models {
    fmt.Printf("- %s (%s)\n", model.Name, model.Provider)
}
```

### Get Provider Information

```go
// Get provider details
provider, err := service.GetProvider(ctx, "anthropic")
if err != nil {
    return err
}

fmt.Printf("Provider: %s\n", provider.Name)
fmt.Printf("Models: %v\n", provider.Models)
```

### Search Models

```go
// Search by capability
models, err := service.SearchModels(ctx, modelsdev.SearchQuery{
    Capabilities: []string{"vision", "tool_use"},
    MinContext:   100000,
    Provider:     "anthropic",
})
if err != nil {
    return err
}
```

### Get Models by Capability

```go
// Find all models with specific capabilities
visionModels, err := service.GetModelsByCapability(ctx, "vision")
if err != nil {
    return err
}

for _, model := range visionModels {
    fmt.Printf("%s supports vision\n", model.Name)
}
```

### Compare Models

```go
// Compare multiple models
comparison := service.CompareModels(ctx, []string{
    "claude-3-opus-20240229",
    "gpt-4-turbo",
    "gemini-1.5-pro",
})

for _, model := range comparison {
    fmt.Printf("%s: context=%d, input=$%.2f/M\n",
        model.Name,
        model.ContextWindow,
        model.Pricing.InputPerMillion)
}
```

### Cache Management

```go
// Force cache refresh
err := service.RefreshCache(ctx)
if err != nil {
    log.Printf("Cache refresh failed: %v", err)
}

// Invalidate specific model
service.InvalidateModel("claude-3-opus-20240229")

// Clear entire cache
service.ClearCache()

// Get cache stats
stats := service.GetCacheStats()
fmt.Printf("Cached: %d models, Last refresh: %v\n",
    stats.ModelCount, stats.LastRefresh)
```

## Integration with HelixAgent

The service is used for model selection and verification:

```go
// In provider registry
func (r *ProviderRegistry) GetModelInfo(modelID string) (*ModelInfo, error) {
    model, err := r.modelsdevService.GetModel(ctx, modelID)
    if err != nil {
        return nil, err
    }

    return &ModelInfo{
        ContextWindow: model.ContextWindow,
        MaxTokens:     model.MaxOutputTokens,
        Capabilities:  model.Capabilities,
        Pricing:       model.Pricing,
    }, nil
}

// In verification scoring
func (v *Verifier) GetCostEffectiveness(modelID string) float64 {
    model, _ := v.modelsdevService.GetModel(ctx, modelID)
    if model == nil || model.Pricing == nil {
        return 0.5 // Default score
    }
    // Score based on price-performance ratio
    return calculateCostEffectiveness(model)
}
```

## Background Refresh

```go
// Background refresh runs automatically
service := modelsdev.NewService(modelsdev.ServiceConfig{
    RefreshInterval:  30 * time.Minute,
    EnableBackground: true,
})

// Or manually trigger
service.RefreshCache(ctx)

// Monitor refresh status
service.OnRefresh(func(stats RefreshStats) {
    log.Printf("Refreshed %d models in %v",
        stats.ModelsUpdated, stats.Duration)
})
```

## Error Handling

```go
model, err := service.GetModel(ctx, "unknown-model")
if err != nil {
    switch {
    case errors.Is(err, modelsdev.ErrModelNotFound):
        // Model doesn't exist
    case errors.Is(err, modelsdev.ErrAPIError):
        // API call failed
    case errors.Is(err, modelsdev.ErrCacheExpired):
        // Cache expired and refresh failed
    default:
        // Other error
    }
}
```

## Testing

```bash
go test -v ./internal/modelsdev/...
```

### Mocking

```go
func TestWithMockService(t *testing.T) {
    service := modelsdev.NewMockService()
    service.AddModel(&modelsdev.Model{
        ID:            "test-model",
        Name:          "Test Model",
        ContextWindow: 100000,
    })

    model, err := service.GetModel(ctx, "test-model")
    require.NoError(t, err)
    assert.Equal(t, "Test Model", model.Name)
}
```

## Metrics

```go
metrics := service.GetMetrics()
fmt.Printf("API calls: %d\n", metrics.APICallCount)
fmt.Printf("Cache hits: %d (%.1f%%)\n",
    metrics.CacheHits,
    float64(metrics.CacheHits)/float64(metrics.TotalRequests)*100)
fmt.Printf("Average latency: %v\n", metrics.AverageLatency)
```

## Performance

| Operation | Typical Latency | Notes |
|-----------|----------------|-------|
| Cache Hit | < 1ms | In-memory lookup |
| API Call | 100-500ms | Network latency |
| Full Refresh | 2-5s | All models |

## Best Practices

1. **Use caching**: Don't disable cache for production
2. **Set appropriate TTL**: Balance freshness vs API calls
3. **Handle errors gracefully**: Fallback to cached data on API failure
4. **Monitor metrics**: Track cache hit rates
5. **Pre-warm cache**: Load on startup for better cold-start performance
