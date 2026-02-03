# Integration Test Framework Package

The integration package provides comprehensive integration testing utilities for HelixAgent, enabling end-to-end validation of MCP servers, LLM providers, and AI debate systems.

## Overview

This package implements:

- **MCP-LLM Integration Testing**: Tests MCP servers with all 10 LLM providers
- **AI Debate Integration**: Validates debate system with MCP tool context
- **Docker Compose Test Environment**: Containerized infrastructure for reproducible tests
- **Database Migration Testing**: PostgreSQL schema validation and migration tests
- **Multi-Provider Verification**: Tests provider discovery, completion, and tool calling

## Directory Structure

```
internal/testing/integration/
├── mcp_llm_provider_test.go     # LLM provider integration tests
├── mcp_debate_integration_test.go # AI debate system integration tests
└── README.md                     # This file
```

## Docker Compose Test Environment

### Starting the Test Environment

```bash
# Start full test infrastructure
docker-compose -f docker-compose.test.yml up -d

# Wait for health checks
make test-infra-start

# Run integration tests
make test-integration
```

### Test Infrastructure Services

| Service | Port | Purpose |
|---------|------|---------|
| `mock-llm` | 18081 | OpenAI/Claude/Ollama compatible mock API |
| `oauth-mock` | 18091 | OAuth 2.0 mock server for testing flows |
| `postgres` | 15432 | PostgreSQL test database |
| `redis` | 16379 | Redis cache for testing |
| `ollama` | 11434 | Local LLM for free testing |
| `helixagent` | 8080 | HelixAgent application |
| `prometheus` | 9090 | Metrics collection |
| `grafana` | 3000 | Dashboard visualization |

### Environment Configuration

```bash
# Database connection
DB_HOST=localhost
DB_PORT=15432
DB_USER=helixagent
DB_PASSWORD=helixagent123
DB_NAME=helixagent_db

# Redis connection
REDIS_HOST=localhost
REDIS_PORT=16379
REDIS_PASSWORD=helixagent123

# Mock LLM
MOCK_LLM_PORT=18081
```

## Key Components

### LLMClient

HTTP client for testing LLM provider endpoints:

```go
client := integration.NewLLMClient("http://localhost:8080")

// List available providers
providers, err := client.ListProviders()

// Send completion request
resp, err := client.Complete(&integration.CompletionRequest{
    Provider: "deepseek",
    Model:    "deepseek-chat",
    Messages: []integration.Message{
        {Role: "user", Content: "Hello"},
    },
    Temperature: 0.1,
    MaxTokens:   100,
})
```

### MCPClient

JSON-RPC client for testing MCP servers directly:

```go
client, err := integration.NewMCPClient(9103, 10*time.Second)
defer client.Close()

// Initialize MCP session
err = client.Initialize()

// List available tools
tools, err := client.ListTools()

// Call a tool
resp, err := client.CallTool("get_current_time", map[string]interface{}{
    "timezone": "UTC",
})
```

### DebateClient

HTTP client for testing AI debate endpoints:

```go
client := integration.NewDebateClient("http://localhost:8080")

// Create debate with MCP context
resp, err := client.CreateDebate(&integration.DebateRequest{
    Topic:      "Discuss the importance of testing",
    Mode:       "consensus",
    MCPContext: mcpContext,
})

fmt.Printf("Consensus: %s (confidence: %.2f)\n", resp.Consensus, resp.Confidence)
```

## Supported LLM Providers

| Provider | Model | Auth Type |
|----------|-------|-----------|
| Claude | claude-3-5-sonnet-20241022 | OAuth |
| DeepSeek | deepseek-chat | API Key |
| Gemini | gemini-2.0-flash-exp | API Key |
| Mistral | mistral-large-latest | API Key |
| OpenRouter | openrouter/auto | API Key |
| Qwen | qwen-turbo | OAuth |
| ZAI | zai-chat | API Key |
| Zen | gpt-4o-mini | Free |
| Cerebras | llama3.1-8b | API Key |
| Ollama | llama3.1:8b | Local |

## Test Types

### Provider Discovery Tests

```go
func TestLLMProviderDiscovery(t *testing.T) {
    client := NewLLMClient("http://localhost:8080")

    providers, err := client.ListProviders()
    require.NoError(t, err)
    assert.NotEmpty(t, providers)
}
```

### Completion Tests

```go
func TestLLMProviderCompletion(t *testing.T) {
    for _, provider := range SupportedLLMProviders {
        t.Run(provider.Name, func(t *testing.T) {
            // Test each provider with simple completion
        })
    }
}
```

### MCP Context Integration

```go
func TestMCPContextWithLLMProvider(t *testing.T) {
    // Collect MCP tool results
    mcpContext := collectMCPContext()

    // Test LLM providers can use MCP context
    for _, provider := range SupportedLLMProviders {
        // Verify provider can interpret MCP tool results
    }
}
```

### Tool Calling Tests

```go
func TestLLMToolCalling(t *testing.T) {
    tools := []Tool{
        {Type: "function", Function: FunctionDef{
            Name:        "get_current_time",
            Description: "Get the current time",
            Parameters:  timeParams,
        }},
    }

    // Test each provider with tool definitions
}
```

## Database Migration Testing

### Schema Validation

```go
func TestDatabaseMigrations(t *testing.T) {
    db := setupTestDatabase(t)
    defer db.Close()

    // Run migrations
    err := runMigrations(db, "./sql/schema/")
    require.NoError(t, err)

    // Verify tables exist
    tables := []string{"users", "sessions", "providers", "debates"}
    for _, table := range tables {
        assertTableExists(t, db, table)
    }
}
```

### Migration Rollback Testing

```go
func TestMigrationRollback(t *testing.T) {
    db := setupTestDatabase(t)

    // Apply migration
    applyMigration(db, "001_initial_schema.sql")

    // Rollback migration
    rollbackMigration(db, "001_initial_schema.sql")

    // Verify clean state
}
```

## Test Patterns and Best Practices

### 1. Skip Unavailable Services

```go
client, err := NewMCPClient(port, timeout)
if err != nil {
    t.Skipf("Service not running: %v", err)
    return
}
defer client.Close()
```

### 2. Use Table-Driven Tests

```go
testCases := []struct {
    name     string
    provider string
    model    string
    expected string
}{
    {"DeepSeek", "deepseek", "deepseek-chat", "hello"},
    {"Gemini", "gemini", "gemini-2.0-flash-exp", "hello"},
}

for _, tc := range testCases {
    t.Run(tc.name, func(t *testing.T) {
        // Test logic
    })
}
```

### 3. Clean Up Resources

```go
func TestWithCleanup(t *testing.T) {
    client := setupClient(t)
    t.Cleanup(func() {
        client.Close()
    })

    // Test logic
}
```

### 4. Use Contexts with Timeouts

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

resp, err := client.CompleteWithContext(ctx, req)
```

### 5. Parallel Test Execution

```go
func TestProviders(t *testing.T) {
    t.Parallel() // Enable parallel execution

    for _, provider := range providers {
        provider := provider // Capture range variable
        t.Run(provider.Name, func(t *testing.T) {
            t.Parallel() // Run subtests in parallel
            // Test logic
        })
    }
}
```

## Running Integration Tests

```bash
# Run all integration tests
make test-integration

# Run with infrastructure
make test-with-infra

# Run specific test
go test -v -run TestMCPDebateIntegration ./internal/testing/integration/

# Run with verbose output
go test -v -count=1 ./internal/testing/integration/...

# Run with race detection
go test -race ./internal/testing/integration/...
```

## Benchmarking

```bash
# Run benchmarks
go test -bench=. ./internal/testing/integration/

# Benchmark with memory profiling
go test -bench=. -benchmem ./internal/testing/integration/

# Benchmark specific function
go test -bench=BenchmarkLLMCompletion ./internal/testing/integration/
```

### Example Benchmark

```go
func BenchmarkLLMCompletion(b *testing.B) {
    client := NewLLMClient("http://localhost:8080")
    req := &CompletionRequest{
        Model:    "auto",
        Messages: []Message{{Role: "user", Content: "Say 'test'."}},
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := client.Complete(req)
        if err != nil {
            b.Skipf("LLM not available: %v", err)
        }
    }
}
```

## Troubleshooting

### Common Issues

1. **Connection Refused**: Ensure test infrastructure is running
   ```bash
   docker-compose -f docker-compose.test.yml up -d
   ```

2. **Provider Not Configured**: Set required API keys
   ```bash
   export DEEPSEEK_API_KEY=your-key
   ```

3. **Timeout Errors**: Increase test timeouts
   ```go
   client, err := NewMCPClient(port, 30*time.Second)
   ```

4. **Database Connection**: Verify PostgreSQL is healthy
   ```bash
   docker-compose -f docker-compose.test.yml ps postgres
   ```

## See Also

- `docker-compose.test.yml` - Test infrastructure configuration
- `internal/testing/mcp/` - MCP protocol testing utilities
- `internal/testing/llm/` - LLM testing framework
- `tests/integration/` - Additional integration test suites
