# User Manual 22: Custom Provider Guide

## Overview
Adding custom LLM providers to HelixAgent.

## Implementation Steps

### 1. Create Provider File
```go
package myprovider

type Provider struct {
    apiKey string
    client *http.Client
}

func New(apiKey string) *Provider {
    return &Provider{
        apiKey: apiKey,
        client: &http.Client{Timeout: 30 * time.Second},
    }
}
```

### 2. Implement Interface
```go
func (p *Provider) Complete(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
    // Implementation
}
```

### 3. Register Provider
```go
registry.Register("myprovider", func(config map[string]string) (LLMProvider, error) {
    return New(config["api_key"]), nil
})
```

## Testing
```bash
go test ./internal/llm/providers/myprovider/...
```
