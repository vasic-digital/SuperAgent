# Toolkit Library Documentation

This directory contains documentation for the HelixAgent Toolkit library, which provides shared components and provider implementations for building AI-powered applications.

## Overview

The Toolkit is a Go library located at `/Toolkit/` that provides:
- Shared utility components (Commons)
- LLM provider implementations (Providers)
- Configuration management
- HTTP client utilities with retry and rate limiting
- Model discovery services

## Documentation Index

| Document | Description |
|----------|-------------|
| [Agents Development Guide](./AGENTS.md) | Development guidelines and coding standards |
| [SiliconFlow Agents](./SILICONFLOW_AGENTS.md) | SiliconFlow provider development guide |
| [Chutes Implementation Summary](./CHUTES_IMPLEMENTATION_SUMMARY.md) | Chutes provider implementation details |
| [Comprehensive Implementation Plan](./COMPREHENSIVE_IMPLEMENTATION_PLAN.md) | Full implementation roadmap |
| [Migration Summary](./MIGRATION_SUMMARY.md) | Provider migration documentation |
| [Refactoring Summary](./REFACTORING_SUMMARY.md) | Project refactoring details |
| [Relocation Summary](./RELOCATION_SUMMARY.md) | Module relocation documentation |
| [Final Refactoring Summary](./FINAL_REFACTORING_SUMMARY.md) | Final refactoring outcomes |
| [Commit Summary](./COMMIT_SUMMARY.md) | Recent commit history |

## Core Components

### Commons (`Toolkit/Commons/`)

Shared codebase used by all provider implementations:

```
Toolkit/Commons/
├── http/           # HTTP client with retry and rate limiting
├── config/         # Configuration management utilities
├── auth/           # Authentication helpers
├── discovery/      # Model discovery interfaces
├── errors/         # Error handling utilities
├── ratelimit/      # Rate limiting functionality
├── response/       # Response handling utilities
└── testing/        # Testing utilities and mocks
```

#### HTTP Client

```go
import "github.com/helixagent/toolkit/Commons/http"

client := http.NewClient(&http.Config{
    Timeout:       30 * time.Second,
    MaxRetries:    3,
    RetryWait:     time.Second,
    RateLimitRPS:  10,
})

resp, err := client.Post(ctx, url, body)
```

#### Configuration Management

```go
import "github.com/helixagent/toolkit/Commons/config"

cfg := config.NewBuilder().
    WithAPIKey(os.Getenv("API_KEY")).
    WithBaseURL("https://api.example.com").
    WithTimeout(30 * time.Second).
    Build()
```

### Providers (`Toolkit/Providers/`)

Individual LLM provider implementations:

```
Toolkit/Providers/
├── SiliconFlow/    # SiliconFlow provider
│   ├── siliconflow.go
│   ├── builder.go
│   ├── client.go
│   ├── discovery.go
│   └── siliconflow_test.go
└── Chutes/         # Chutes provider
    ├── chutes.go
    ├── builder.go
    ├── client.go
    ├── discovery.go
    └── chutes_test.go
```

## Extension Development

### Creating a New Provider

1. **Create provider directory**:
   ```bash
   mkdir -p Toolkit/Providers/YourProvider
   ```

2. **Implement the provider interface**:
   ```go
   // yourprovider.go
   package yourprovider

   import "github.com/helixagent/toolkit/pkg/toolkit"

   type YourProvider struct {
       client *Client
       config *Config
   }

   func (p *YourProvider) Name() string {
       return "yourprovider"
   }

   func (p *YourProvider) Chat(ctx context.Context, req *toolkit.ChatRequest) (*toolkit.ChatResponse, error) {
       // Implementation
   }

   func (p *YourProvider) Embed(ctx context.Context, req *toolkit.EmbedRequest) (*toolkit.EmbedResponse, error) {
       // Implementation
   }

   func (p *YourProvider) DiscoverModels(ctx context.Context) ([]toolkit.Model, error) {
       // Implementation
   }

   func (p *YourProvider) ValidateConfig() error {
       // Implementation
   }
   ```

3. **Implement configuration builder**:
   ```go
   // builder.go
   package yourprovider

   type ConfigBuilder struct {
       apiKey  string
       baseURL string
       timeout time.Duration
   }

   func NewConfigBuilder() *ConfigBuilder {
       return &ConfigBuilder{
           baseURL: "https://api.yourprovider.com",
           timeout: 30 * time.Second,
       }
   }

   func (b *ConfigBuilder) WithAPIKey(key string) *ConfigBuilder {
       b.apiKey = key
       return b
   }

   func (b *ConfigBuilder) Build() (*Config, error) {
       if b.apiKey == "" {
           return nil, errors.New("API key required")
       }
       return &Config{
           APIKey:  b.apiKey,
           BaseURL: b.baseURL,
           Timeout: b.timeout,
       }, nil
   }
   ```

4. **Implement HTTP client**:
   ```go
   // client.go
   package yourprovider

   type Client struct {
       httpClient *http.Client
       baseURL    string
       apiKey     string
   }

   func NewClient(config *Config) *Client {
       return &Client{
           httpClient: http.NewClient(config.Timeout),
           baseURL:    config.BaseURL,
           apiKey:     config.APIKey,
       }
   }

   func (c *Client) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
       // Make API request
   }
   ```

5. **Implement model discovery**:
   ```go
   // discovery.go
   package yourprovider

   func (p *YourProvider) DiscoverModels(ctx context.Context) ([]toolkit.Model, error) {
       // Query provider API for available models
       // Return normalized model list
   }
   ```

6. **Add auto-registration**:
   ```go
   func init() {
       toolkit.RegisterProvider("yourprovider", NewYourProvider)
   }
   ```

7. **Write tests**:
   ```go
   // yourprovider_test.go
   package yourprovider

   func TestYourProvider_Chat(t *testing.T) {
       // Test implementation
   }

   func TestYourProvider_Embed(t *testing.T) {
       // Test implementation
   }
   ```

### Provider Interface

All providers must implement the toolkit.Provider interface:

```go
type Provider interface {
    Name() string
    Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
    Embed(ctx context.Context, req *EmbedRequest) (*EmbedResponse, error)
    Rerank(ctx context.Context, req *RerankRequest) (*RerankResponse, error)
    DiscoverModels(ctx context.Context) ([]Model, error)
    ValidateConfig() error
}
```

## Integration Patterns

### Using Providers in HelixAgent

```go
import (
    "github.com/helixagent/toolkit/Providers/SiliconFlow"
    "github.com/helixagent/toolkit/Providers/Chutes"
)

// Initialize providers
sfProvider := siliconflow.NewProvider(sfConfig)
chutesProvider := chutes.NewProvider(chutesConfig)

// Register with HelixAgent
registry.Register(sfProvider)
registry.Register(chutesProvider)
```

### Multi-Provider Configuration

```go
// Load multiple providers from config
providers := []toolkit.Provider{}
for name, cfg := range providerConfigs {
    provider, err := toolkit.CreateProvider(name, cfg)
    if err != nil {
        log.Printf("Failed to create %s: %v", name, err)
        continue
    }
    providers = append(providers, provider)
}
```

## Build and Test

```bash
# Build toolkit
cd Toolkit
go build ./...

# Run tests
go test ./...

# Run specific provider tests
go test -v ./Providers/SiliconFlow/...
go test -v ./Providers/Chutes/...

# Run with coverage
go test -cover ./...
```

## Code Style Guidelines

- **Go version**: 1.23+
- **Formatting**: Use `gofmt` for consistent formatting
- **Imports**: Group imports (stdlib, third-party, internal)
- **Naming**: CamelCase for exports, camelCase for private
- **Error handling**: Always wrap errors with context
- **Documentation**: Document all exported functions

## Related Documentation

- [Main HelixAgent Documentation](../README.md)
- [Provider Development Guide](../development/PROVIDERS.md)
- [API Reference](../api/README.md)
