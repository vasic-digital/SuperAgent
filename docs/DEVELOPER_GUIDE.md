# HelixAgent Developer Guide

This guide provides comprehensive information for developers working on HelixAgent.

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Development Environment](#development-environment)
3. [Core Components](#core-components)
4. [Adding New Features](#adding-new-features)
5. [Testing Strategy](#testing-strategy)
6. [Debugging](#debugging)
7. [Performance Optimization](#performance-optimization)
8. [Best Practices](#best-practices)

---

## Architecture Overview

### System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        API Gateway                               │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌───────┐ │
│  │  REST   │  │  gRPC   │  │   MCP   │  │   LSP   │  │  ACP  │ │
│  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘  └───┬───┘ │
└───────┼────────────┼────────────┼────────────┼───────────┼─────┘
        │            │            │            │           │
┌───────▼────────────▼────────────▼────────────▼───────────▼─────┐
│                       Service Layer                             │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐       │
│  │  Debate  │  │ Ensemble │  │  Intent  │  │ Context  │       │
│  │ Service  │  │ Service  │  │Classifier│  │ Manager  │       │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘       │
└───────┼─────────────┼─────────────┼─────────────┼──────────────┘
        │             │             │             │
┌───────▼─────────────▼─────────────▼─────────────▼──────────────┐
│                      Provider Layer                             │
│  ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐       │
│  │ Claude │ │DeepSeek│ │ Gemini │ │Mistral │ │  ...   │       │
│  └────────┘ └────────┘ └────────┘ └────────┘ └────────┘       │
└─────────────────────────────────────────────────────────────────┘
        │             │             │             │
┌───────▼─────────────▼─────────────▼─────────────▼──────────────┐
│                    Infrastructure Layer                         │
│  ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐       │
│  │  DB    │ │ Redis  │ │ Qdrant │ │ Kafka  │ │Prometheus│     │
│  └────────┘ └────────┘ └────────┘ └────────┘ └────────┘       │
└─────────────────────────────────────────────────────────────────┘
```

### Request Flow

1. Request arrives at API Gateway
2. Authentication/Authorization middleware validates
3. Request routed to appropriate handler
4. Service layer processes business logic
5. Provider layer communicates with LLMs
6. Response aggregated and returned

---

## Development Environment

### Required Tools

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.24+ | Primary language |
| Docker | 24+ | Containerization |
| Make | Any | Build automation |
| golangci-lint | 1.55+ | Code linting |
| govulncheck | Latest | Security scanning |

### Environment Setup

```bash
# 1. Clone repository
git clone https://github.com/your-org/HelixAgent.git
cd HelixAgent

# 2. Install Go dependencies
go mod download

# 3. Install dev tools
make install-deps

# 4. Copy environment file
cp .env.example .env

# 5. Configure API keys in .env
# DEEPSEEK_API_KEY=your-key
# GEMINI_API_KEY=your-key
# etc.

# 6. Verify setup
make test
make build
```

### IDE Configuration

#### VS Code

Recommended extensions:
- Go (by Google)
- GitLens
- Error Lens

Settings (`settings.json`):
```json
{
    "go.useLanguageServer": true,
    "go.lintTool": "golangci-lint",
    "go.lintOnSave": "workspace",
    "go.testOnSave": true,
    "editor.formatOnSave": true
}
```

#### GoLand

1. Enable "Go Modules" integration
2. Set "File Watchers" for `gofmt`
3. Configure golangci-lint as external tool

---

## Core Components

### LLM Provider System

Location: `internal/llm/providers/`

Each provider implements the `LLMProvider` interface:

```go
type LLMProvider interface {
    Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
    CompleteStream(ctx context.Context, req *CompletionRequest) (<-chan *StreamChunk, error)
    HealthCheck(ctx context.Context) error
    GetCapabilities() *ProviderCapabilities
    ValidateConfig() error
}
```

### AI Debate System

Location: `internal/services/debate_service.go`

The debate system orchestrates multi-LLM consensus:

1. **Topic Introduction**: Present debate topic
2. **Position Gathering**: Each LLM provides position
3. **Cross-Examination**: LLMs critique each other
4. **Synthesis**: Generate consensus response

### Ensemble Voting

Location: `internal/llm/ensemble.go`

Voting strategies:
- **Majority Vote**: Most common response wins
- **Weighted Vote**: Provider scores influence votes
- **Confidence Weighted**: Response confidence affects weight

### Intent Classification

Location: `internal/services/llm_intent_classifier.go`

Uses LLM-based semantic understanding (not hardcoded patterns):

```go
type IntentClassifier interface {
    Classify(ctx context.Context, message string) (*IntentResult, error)
}
```

---

## Adding New Features

### Adding a New LLM Provider

1. **Create provider directory**:
```bash
mkdir -p internal/llm/providers/newprovider
```

2. **Implement provider**:
```go
// internal/llm/providers/newprovider/provider.go
package newprovider

type Provider struct {
    config  *Config
    client  *http.Client
    logger  *logrus.Logger
}

func NewProvider(config *Config, logger *logrus.Logger) *Provider {
    return &Provider{
        config: config,
        client: &http.Client{Timeout: config.Timeout},
        logger: logger,
    }
}

func (p *Provider) Complete(ctx context.Context, req *llm.CompletionRequest) (*llm.CompletionResponse, error) {
    // Implementation
}

// Implement remaining interface methods...
```

3. **Add tests**:
```go
// internal/llm/providers/newprovider/provider_test.go
func TestProvider_Complete(t *testing.T) {
    // Test implementation
}
```

4. **Register provider**:
```go
// internal/services/provider_registry.go
registry.Register("newprovider", func(config *Config) LLMProvider {
    return newprovider.NewProvider(config, logger)
})
```

5. **Update documentation**:
- Add to `CLAUDE.md`
- Add to provider documentation
- Update `.env.example`

### Adding a New API Endpoint

1. **Create handler**:
```go
// internal/handlers/new_handler.go
func (h *Handler) HandleNewEndpoint(c *gin.Context) {
    // Implementation
}
```

2. **Add route**:
```go
// internal/router/router.go
router.POST("/v1/new-endpoint", handler.HandleNewEndpoint)
```

3. **Add tests**:
```go
// internal/handlers/new_handler_test.go
func TestHandler_HandleNewEndpoint(t *testing.T) {
    // Test with httptest
}
```

### Adding a New Tool

1. **Define tool schema**:
```go
// internal/tools/schema.go
var NewTool = &Tool{
    Name:        "new_tool",
    Description: "Description of the tool",
    Parameters: Parameters{
        Type: "object",
        Properties: map[string]Property{
            "param1": {Type: "string", Description: "..."},
        },
        Required: []string{"param1"},
    },
}
```

2. **Implement handler**:
```go
// internal/tools/handler.go
func (h *ToolHandler) HandleNewTool(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    // Implementation
}
```

---

## Testing Strategy

### Test Pyramid

```
        /\
       /  \      E2E Tests (few)
      /----\
     /      \    Integration Tests (moderate)
    /--------\
   /          \  Unit Tests (many)
  /------------\
```

### Unit Tests

```go
func TestFunction(t *testing.T) {
    // Arrange
    sut := NewSystemUnderTest()
    input := "test input"
    expected := "expected output"

    // Act
    result, err := sut.Function(input)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

### Integration Tests

```go
//go:build integration

func TestDatabaseIntegration(t *testing.T) {
    db := setupTestDB(t)
    defer teardownTestDB(t, db)

    // Test with real database
}
```

### Mocking

Use interfaces for mockability:

```go
// Define interface
type LLMClient interface {
    Complete(ctx context.Context, req *Request) (*Response, error)
}

// Mock implementation
type MockLLMClient struct {
    mock.Mock
}

func (m *MockLLMClient) Complete(ctx context.Context, req *Request) (*Response, error) {
    args := m.Called(ctx, req)
    return args.Get(0).(*Response), args.Error(1)
}
```

---

## Debugging

### Logging

```go
import "github.com/sirupsen/logrus"

logger.WithFields(logrus.Fields{
    "provider": "claude",
    "request_id": requestID,
}).Info("Processing request")
```

### Tracing

```go
import "dev.helix.agent/internal/observability"

ctx, span := tracer.Start(ctx, "operation-name")
defer span.End()

span.SetAttributes(attribute.String("key", "value"))
```

### Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| Context deadline exceeded | Timeout too short | Increase timeout |
| Connection refused | Service not running | Start dependencies |
| Rate limited | Too many requests | Add backoff/retry |
| JSON unmarshal error | Schema mismatch | Check API response format |

### Debug Mode

```bash
# Run with debug logging
GIN_MODE=debug LOG_LEVEL=debug make run-dev

# Run specific test with verbose output
go test -v -run TestName ./path/to/package
```

---

## Performance Optimization

### Caching

```go
// Use Redis for distributed caching
cache := redis.NewCache(config)
result, err := cache.GetOrSet(ctx, key, func() (interface{}, error) {
    return expensiveOperation()
}, ttl)
```

### Connection Pooling

```go
// Database connection pool
db, err := pgx.NewPool(ctx, &pgx.PoolConfig{
    MaxConns: 100,
    MinConns: 10,
})
```

### Parallel Processing

```go
// Use errgroup for parallel operations
g, ctx := errgroup.WithContext(ctx)
for _, provider := range providers {
    p := provider
    g.Go(func() error {
        return p.Process(ctx)
    })
}
err := g.Wait()
```

---

## Best Practices

### Error Handling

```go
// Wrap errors with context
if err != nil {
    return nil, fmt.Errorf("failed to process request: %w", err)
}

// Use custom error types for specific cases
type ValidationError struct {
    Field   string
    Message string
}
```

### Context Usage

```go
// Always pass context
func Process(ctx context.Context, data *Data) error {
    // Check for cancellation
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    
    // Continue processing
}
```

### Resource Cleanup

```go
// Use defer for cleanup
func ProcessFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close()
    
    // Process file
}
```

### Configuration

```go
// Use environment variables with defaults
func LoadConfig() *Config {
    return &Config{
        Port:    getEnvOrDefault("PORT", "8080"),
        Timeout: getDurationOrDefault("TIMEOUT", 30*time.Second),
    }
}
```

---

## Quick Reference

### Common Commands

```bash
# Build
make build

# Test
make test
make test-coverage

# Quality
make lint
make fmt

# Run
make run-dev

# Infrastructure
make test-infra-start
make test-infra-stop
```

### Package Documentation

```bash
# Generate godoc
godoc -http=:6060
# Visit http://localhost:6060/pkg/dev.helix.agent/
```

### API Documentation

Swagger UI available at `/swagger/index.html` when running in development mode.

---

## Resources

- [Go Documentation](https://golang.org/doc/)
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [HelixAgent API Reference](./api/)
