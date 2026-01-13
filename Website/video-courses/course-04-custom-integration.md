# Course 4: Custom Integration

**Level**: Advanced Developer
**Duration**: 45 minutes
**Prerequisites**: Courses 1-3, Go programming knowledge

## Overview

Learn how to extend HelixAgent with custom plugins, providers, and integrations.

---

## Module 1: Plugin Development (15 minutes)

### 1.1 Plugin Architecture
- Plugin system design principles
- Hot-reload mechanism
- Plugin lifecycle management
- Discovery and registration

### 1.2 Implementing Plugin Interface

```go
package myplugin

import (
    "context"
    "github.com/helixagent/helixagent/internal/plugins"
)

type MyPlugin struct {
    config *Config
    logger *logrus.Logger
}

func (p *MyPlugin) Name() string {
    return "my-plugin"
}

func (p *MyPlugin) Version() string {
    return "1.0.0"
}

func (p *MyPlugin) Init(ctx context.Context) error {
    p.logger.Info("Initializing my-plugin")
    return nil
}

func (p *MyPlugin) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    // Plugin logic here
    return result, nil
}

func (p *MyPlugin) Shutdown(ctx context.Context) error {
    p.logger.Info("Shutting down my-plugin")
    return nil
}

func (p *MyPlugin) Health() plugins.HealthStatus {
    return plugins.HealthStatus{
        Healthy: true,
        Message: "Plugin is running",
    }
}
```

### 1.3 Plugin Configuration

```yaml
# plugins/my-plugin/config.yaml
name: my-plugin
version: 1.0.0
enabled: true
config:
  setting1: value1
  setting2: value2
dependencies:
  - core-services
```

### 1.4 Testing Plugins

```go
func TestMyPlugin(t *testing.T) {
    plugin := &MyPlugin{
        config: &Config{},
        logger: logrus.New(),
    }

    err := plugin.Init(context.Background())
    assert.NoError(t, err)

    result, err := plugin.Execute(context.Background(), "test input")
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

### Hands-On Exercise
Create a simple logging plugin that tracks all API requests.

---

## Module 2: Custom Provider Integration (15 minutes)

### 2.1 Provider Interface

```go
package providers

type LLMProvider interface {
    // Complete sends a completion request
    Complete(ctx context.Context, req *Request) (*Response, error)

    // CompleteStream sends a streaming completion request
    CompleteStream(ctx context.Context, req *Request) (<-chan *StreamChunk, error)

    // HealthCheck verifies provider is accessible
    HealthCheck(ctx context.Context) error

    // GetCapabilities returns provider capabilities
    GetCapabilities() *Capabilities

    // ValidateConfig validates provider configuration
    ValidateConfig() error
}
```

### 2.2 Implementing Custom Provider

```go
package mycustomprovider

import (
    "context"
    "github.com/helixagent/helixagent/internal/llm"
)

type MyCustomProvider struct {
    apiKey   string
    baseURL  string
    client   *http.Client
}

func NewMyCustomProvider(config *Config) (*MyCustomProvider, error) {
    return &MyCustomProvider{
        apiKey:  config.APIKey,
        baseURL: config.BaseURL,
        client:  &http.Client{Timeout: 30 * time.Second},
    }, nil
}

func (p *MyCustomProvider) Complete(ctx context.Context, req *llm.Request) (*llm.Response, error) {
    // Make API request to your custom provider
    payload := map[string]interface{}{
        "prompt":     req.Messages,
        "max_tokens": req.MaxTokens,
        "temperature": req.Temperature,
    }

    resp, err := p.makeRequest(ctx, "/v1/complete", payload)
    if err != nil {
        return nil, err
    }

    return &llm.Response{
        Content:     resp.Text,
        FinishReason: "stop",
        Usage: &llm.Usage{
            PromptTokens:     resp.Usage.Input,
            CompletionTokens: resp.Usage.Output,
        },
    }, nil
}

func (p *MyCustomProvider) CompleteStream(ctx context.Context, req *llm.Request) (<-chan *llm.StreamChunk, error) {
    chunks := make(chan *llm.StreamChunk)

    go func() {
        defer close(chunks)
        // Stream implementation
    }()

    return chunks, nil
}

func (p *MyCustomProvider) HealthCheck(ctx context.Context) error {
    resp, err := p.client.Get(p.baseURL + "/health")
    if err != nil {
        return err
    }
    if resp.StatusCode != 200 {
        return fmt.Errorf("provider unhealthy: %d", resp.StatusCode)
    }
    return nil
}

func (p *MyCustomProvider) GetCapabilities() *llm.Capabilities {
    return &llm.Capabilities{
        Streaming:     true,
        MaxTokens:     4096,
        Vision:        false,
        FunctionCalls: true,
    }
}
```

### 2.3 Registering Custom Provider

```go
// In provider_registry.go
func (r *ProviderRegistry) RegisterCustomProvider(name string, factory ProviderFactory) {
    r.factories[name] = factory
}

// Usage
registry.RegisterCustomProvider("my-provider", func(cfg *Config) (LLMProvider, error) {
    return NewMyCustomProvider(cfg)
})
```

### Hands-On Exercise
Create a wrapper for a new LLM API endpoint.

---

## Module 3: Advanced API Usage (15 minutes)

### 3.1 Custom Endpoints

```go
package handlers

func RegisterCustomEndpoints(router *gin.Engine) {
    custom := router.Group("/v1/custom")
    {
        custom.POST("/analyze", handleAnalyze)
        custom.POST("/summarize", handleSummarize)
        custom.GET("/metrics", handleMetrics)
    }
}

func handleAnalyze(c *gin.Context) {
    var req AnalyzeRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    result, err := analyzeService.Analyze(c.Request.Context(), req)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, result)
}
```

### 3.2 Custom Middleware

```go
package middleware

func RequestLogger() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path

        c.Next()

        latency := time.Since(start)
        status := c.Writer.Status()

        log.Printf("[%d] %s %s - %v", status, c.Request.Method, path, latency)
    }
}

func RateLimiter(rps int) gin.HandlerFunc {
    limiter := rate.NewLimiter(rate.Limit(rps), rps)

    return func(c *gin.Context) {
        if !limiter.Allow() {
            c.JSON(429, gin.H{"error": "rate limit exceeded"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

### 3.3 Request Processing

```go
package processing

type RequestProcessor struct {
    providers *ProviderRegistry
    cache     *Cache
    metrics   *Metrics
}

func (p *RequestProcessor) Process(ctx context.Context, req *Request) (*Response, error) {
    // Check cache
    if cached, ok := p.cache.Get(req.CacheKey()); ok {
        p.metrics.CacheHit()
        return cached.(*Response), nil
    }

    // Select provider
    provider, err := p.providers.SelectBest(ctx)
    if err != nil {
        return nil, err
    }

    // Process request
    resp, err := provider.Complete(ctx, req)
    if err != nil {
        return nil, err
    }

    // Cache result
    p.cache.Set(req.CacheKey(), resp, time.Hour)

    return resp, nil
}
```

### 3.4 Response Customization

```go
package transformers

type ResponseTransformer interface {
    Transform(response *Response) *Response
}

type JSONResponseTransformer struct{}

func (t *JSONResponseTransformer) Transform(resp *Response) *Response {
    // Extract JSON from response
    jsonContent := extractJSON(resp.Content)
    resp.Content = jsonContent
    resp.Format = "json"
    return resp
}

type SummaryTransformer struct {
    maxLength int
}

func (t *SummaryTransformer) Transform(resp *Response) *Response {
    if len(resp.Content) > t.maxLength {
        resp.Content = resp.Content[:t.maxLength] + "..."
    }
    return resp
}
```

### Hands-On Exercise
Create a custom endpoint that combines multiple LLM responses.

---

## Lab Exercise

### Building a Complete Custom Integration

1. **Create Custom Plugin**
   - Implement sentiment analysis plugin
   - Add configuration for thresholds
   - Integrate with debate system

2. **Add Custom Provider**
   - Wrap a new LLM API
   - Implement health checks
   - Add to provider registry

3. **Create Custom Endpoint**
   - Build `/v1/analyze/sentiment` endpoint
   - Add input validation
   - Implement caching

### Deliverables
- Working plugin code
- Provider implementation
- API endpoint
- Unit tests

---

## Best Practices

### Plugin Development
- Use dependency injection
- Implement proper error handling
- Add comprehensive logging
- Write unit tests

### Provider Integration
- Implement retry logic
- Handle rate limiting
- Add timeout management
- Monitor health metrics

### API Development
- Validate all inputs
- Use consistent error formats
- Implement proper caching
- Add request tracing

---

## Resources

- [Plugin API Reference](/docs/api/plugins)
- [Provider Interface Guide](/docs/guides/providers)
- [Middleware Documentation](/docs/api/middleware)
- [Example Plugins](https://github.com/helixagent/example-plugins)

---

## Next Steps

- Explore advanced optimization features
- Join the developer community
- Contribute plugins to the ecosystem
- Get certified as HelixAgent Developer

---

*Course Version: 1.0.0*
*Last Updated: January 2026*
