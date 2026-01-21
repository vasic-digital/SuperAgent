# HelixAgent Implementation Plan

**Version**: 1.0
**Created**: January 21, 2026
**Duration**: 14 Weeks
**Goal**: 100% Completion - Zero Broken/Disabled/Undocumented Items

---

## Quick Reference: Commands by Phase

```bash
# Phase 1: Critical Coverage
make test-coverage              # Check current coverage
go test -v -coverprofile=coverage.out ./internal/router/...
go tool cover -html=coverage.out -o coverage.html

# Phase 4: Enable All Tests
make test-infra-start           # Start test infrastructure
make test-with-infra            # Run all tests with infra
make test-integration           # Integration tests
make test-e2e                   # End-to-end tests
make test-security              # Security tests
make test-stress                # Stress tests

# Phase 8: Final Verification
make test-all                   # All tests
make lint                       # Code quality
make security-scan              # Security scan
```

---

## Phase 1: Critical Test Coverage (Week 1-2)

### Week 1: Days 1-3 - Router Package

**Target**: `internal/router` from 53.6% to 80%

**Step 1.1: Analyze Uncovered Code**
```bash
cd /run/media/milosvasic/DATA4TB/Projects/HelixAgent
go test -coverprofile=/tmp/router_coverage.out ./internal/router/...
go tool cover -func=/tmp/router_coverage.out | grep -v "100.0%"
```

**Step 1.2: Create Test File Structure**
```
internal/router/
├── router_test.go          # Existing tests
├── router_handlers_test.go # New: Handler tests
├── router_middleware_test.go # New: Middleware tests
└── router_errors_test.go   # New: Error handling tests
```

**Step 1.3: Write Tests for Uncovered Paths**

Create `internal/router/router_handlers_test.go`:
```go
package router

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestRouter_HandleChatCompletion(t *testing.T) {
    tests := []struct {
        name           string
        requestBody    string
        expectedStatus int
        expectedError  string
    }{
        {
            name:           "valid request",
            requestBody:    `{"model": "test", "messages": [{"role": "user", "content": "hello"}]}`,
            expectedStatus: http.StatusOK,
        },
        {
            name:           "invalid JSON",
            requestBody:    `{invalid}`,
            expectedStatus: http.StatusBadRequest,
            expectedError:  "invalid JSON",
        },
        {
            name:           "missing model",
            requestBody:    `{"messages": [{"role": "user", "content": "hello"}]}`,
            expectedStatus: http.StatusBadRequest,
            expectedError:  "model required",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            router := setupTestRouter(t)
            req := httptest.NewRequest("POST", "/v1/chat/completions", strings.NewReader(tt.requestBody))
            req.Header.Set("Content-Type", "application/json")
            w := httptest.NewRecorder()

            router.ServeHTTP(w, req)

            assert.Equal(t, tt.expectedStatus, w.Code)
            if tt.expectedError != "" {
                assert.Contains(t, w.Body.String(), tt.expectedError)
            }
        })
    }
}
```

**Step 1.4: Verify Coverage**
```bash
go test -coverprofile=/tmp/router_coverage.out ./internal/router/...
go tool cover -func=/tmp/router_coverage.out
# Target: >= 80%
```

### Week 1: Days 4-5 - Background Package

**Target**: `internal/background` from 35% to 80%

**Step 1.5: Create Test Structure**
```
internal/background/
├── executor_test.go      # Task execution tests
├── queue_test.go         # Queue operations tests
├── worker_test.go        # Worker pool tests
├── monitor_test.go       # Resource monitoring tests
└── stuck_detector_test.go # Stuck detection tests
```

**Step 1.6: Write Comprehensive Tests**

Create `internal/background/executor_test.go`:
```go
package background

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestTaskExecutor_Execute(t *testing.T) {
    executor := NewTaskExecutor(DefaultConfig())

    t.Run("successful execution", func(t *testing.T) {
        task := &Task{
            ID:   "test-1",
            Type: TaskTypeCommand,
            Payload: map[string]interface{}{
                "command": "echo hello",
            },
        }

        result, err := executor.Execute(context.Background(), task)
        require.NoError(t, err)
        assert.Equal(t, TaskStatusCompleted, result.Status)
    })

    t.Run("timeout handling", func(t *testing.T) {
        ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
        defer cancel()

        task := &Task{
            ID:   "test-2",
            Type: TaskTypeCommand,
            Payload: map[string]interface{}{
                "command": "sleep 10",
            },
        }

        result, err := executor.Execute(ctx, task)
        require.Error(t, err)
        assert.Equal(t, TaskStatusFailed, result.Status)
    })

    t.Run("panic recovery", func(t *testing.T) {
        task := &Task{
            ID:   "test-3",
            Type: TaskTypeCustom,
            Handler: func(ctx context.Context) error {
                panic("test panic")
            },
        }

        result, err := executor.Execute(context.Background(), task)
        require.Error(t, err)
        assert.Contains(t, err.Error(), "panic recovered")
    })
}
```

### Week 2: Days 1-3 - Notifications Package

**Target**: `internal/notifications` from 42% to 80%

**Step 1.7: Test SSE, WebSocket, Webhooks**

Create `internal/notifications/sse_test.go`:
```go
package notifications

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestSSENotifier_Send(t *testing.T) {
    notifier := NewSSENotifier()

    t.Run("single subscriber", func(t *testing.T) {
        w := httptest.NewRecorder()
        ctx, cancel := context.WithCancel(context.Background())

        // Start subscriber in goroutine
        go func() {
            notifier.Subscribe(ctx, w, "channel-1")
        }()

        // Wait for subscription
        time.Sleep(100 * time.Millisecond)

        // Send notification
        err := notifier.Send("channel-1", &Notification{
            Type:    "test",
            Payload: map[string]interface{}{"message": "hello"},
        })
        require.NoError(t, err)

        // Cancel and verify
        cancel()
        time.Sleep(100 * time.Millisecond)

        assert.Contains(t, w.Body.String(), "event: test")
        assert.Contains(t, w.Body.String(), `"message":"hello"`)
    })
}
```

### Week 2: Days 4-5 - Plugins & LLM Providers

**Target**: `internal/plugins` from 38% to 80%, `zen` from 40% to 80%, `cerebras` from 45% to 80%

**Step 1.8: Plugin System Tests**

Create `internal/plugins/loader_test.go`:
```go
package plugins

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestPluginLoader_Load(t *testing.T) {
    loader := NewPluginLoader(DefaultConfig())

    t.Run("load valid plugin", func(t *testing.T) {
        plugin, err := loader.Load(context.Background(), "testdata/valid_plugin.so")
        require.NoError(t, err)
        assert.NotNil(t, plugin)
        assert.Equal(t, "valid_plugin", plugin.Name())
    })

    t.Run("load invalid plugin", func(t *testing.T) {
        _, err := loader.Load(context.Background(), "testdata/invalid_plugin.so")
        require.Error(t, err)
    })

    t.Run("hot reload", func(t *testing.T) {
        plugin, err := loader.Load(context.Background(), "testdata/reload_plugin.so")
        require.NoError(t, err)

        // Simulate file change
        err = loader.Reload(context.Background(), plugin.Name())
        require.NoError(t, err)
    })
}
```

**Step 1.9: Provider Tests with Mocks**

Create `internal/llm/providers/zen/zen_mock_test.go`:
```go
package zen

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestZenProvider_Complete(t *testing.T) {
    // Mock server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(`{
            "choices": [{"message": {"content": "Hello!"}}],
            "usage": {"total_tokens": 10}
        }`))
    }))
    defer server.Close()

    provider := NewProvider(Config{
        BaseURL: server.URL,
    })

    t.Run("successful completion", func(t *testing.T) {
        response, err := provider.Complete(context.Background(), &CompletionRequest{
            Model:   "zen-1",
            Prompt:  "Hello",
        })
        require.NoError(t, err)
        assert.Equal(t, "Hello!", response.Content)
    })

    t.Run("error handling", func(t *testing.T) {
        badServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.WriteHeader(http.StatusInternalServerError)
        }))
        defer badServer.Close()

        badProvider := NewProvider(Config{
            BaseURL: badServer.URL,
        })

        _, err := badProvider.Complete(context.Background(), &CompletionRequest{})
        require.Error(t, err)
    })
}
```

---

## Phase 2: High Priority Coverage (Week 3-4)

### Week 3: Handlers, Middleware, Cache

**Step 2.1: Handler Tests**

Create comprehensive tests for all HTTP handlers:

```go
// internal/handlers/debate_handler_test.go
func TestDebateHandler_Create(t *testing.T) {
    handler := setupTestHandler(t)

    tests := []struct {
        name           string
        body           string
        expectedStatus int
    }{
        {"valid debate", `{"topic": "Test", "participants": ["claude", "deepseek"]}`, 201},
        {"invalid JSON", `{invalid}`, 400},
        {"missing topic", `{"participants": ["claude"]}`, 400},
        {"empty participants", `{"topic": "Test", "participants": []}`, 400},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

**Step 2.2: Middleware Tests**

```go
// internal/middleware/ratelimit_test.go
func TestRateLimitMiddleware(t *testing.T) {
    middleware := NewRateLimitMiddleware(Config{
        RequestsPerSecond: 10,
        Burst:             5,
    })

    t.Run("allows requests under limit", func(t *testing.T) {
        for i := 0; i < 5; i++ {
            req := httptest.NewRequest("GET", "/", nil)
            w := httptest.NewRecorder()
            middleware.Handle(w, req)
            assert.Equal(t, http.StatusOK, w.Code)
        }
    })

    t.Run("blocks requests over limit", func(t *testing.T) {
        // Exhaust burst
        for i := 0; i < 10; i++ {
            req := httptest.NewRequest("GET", "/", nil)
            w := httptest.NewRecorder()
            middleware.Handle(w, req)
        }

        // This should be rate limited
        req := httptest.NewRequest("GET", "/", nil)
        w := httptest.NewRecorder()
        middleware.Handle(w, req)
        assert.Equal(t, http.StatusTooManyRequests, w.Code)
    })
}
```

### Week 4: Database, Verifier, Services

**Step 2.3: Database Repository Tests**

```go
// internal/database/repository_test.go
//go:build integration

func TestRepository_CRUD(t *testing.T) {
    db := setupTestDB(t)
    defer teardownTestDB(t, db)

    repo := NewRepository(db)

    t.Run("create and retrieve", func(t *testing.T) {
        entity := &Entity{Name: "test"}
        err := repo.Create(context.Background(), entity)
        require.NoError(t, err)
        assert.NotEmpty(t, entity.ID)

        retrieved, err := repo.GetByID(context.Background(), entity.ID)
        require.NoError(t, err)
        assert.Equal(t, entity.Name, retrieved.Name)
    })

    t.Run("update", func(t *testing.T) {
        // ...
    })

    t.Run("delete", func(t *testing.T) {
        // ...
    })
}
```

---

## Phase 3: Medium Priority Coverage (Week 5-6)

### Complete Coverage for Remaining Packages

**Step 3.1: Create Test Template**

For each package with 70-80% coverage:

```go
package <packagename>

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// Test all public functions
func Test<FunctionName>(t *testing.T) {
    // Setup
    sut := NewSystemUnderTest()

    // Test cases
    tests := []struct {
        name     string
        input    interface{}
        expected interface{}
        wantErr  bool
    }{
        // Happy path
        {"valid input", validInput, expectedOutput, false},
        // Edge cases
        {"empty input", "", nil, true},
        {"nil input", nil, nil, true},
        // Error cases
        {"invalid input", invalidInput, nil, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := sut.Function(context.Background(), tt.input)
            if tt.wantErr {
                require.Error(t, err)
                return
            }
            require.NoError(t, err)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

---

## Phase 4: Fix All Skipped Tests (Week 7-8)

### Week 7: Create Mock Infrastructure

**Step 4.1: Create Mock LLM Provider**

```go
// internal/testing/mocks/llm_provider.go
package mocks

import (
    "context"

    "dev.helix.agent/internal/llm"
)

type MockLLMProvider struct {
    CompleteFunc func(ctx context.Context, req *llm.CompletionRequest) (*llm.CompletionResponse, error)
    responses    map[string]string
}

func NewMockLLMProvider() *MockLLMProvider {
    return &MockLLMProvider{
        responses: map[string]string{
            "default": "Mock response",
        },
    }
}

func (m *MockLLMProvider) Complete(ctx context.Context, req *llm.CompletionRequest) (*llm.CompletionResponse, error) {
    if m.CompleteFunc != nil {
        return m.CompleteFunc(ctx, req)
    }
    return &llm.CompletionResponse{
        Content: m.responses["default"],
    }, nil
}
```

**Step 4.2: Create Test Infrastructure Docker Compose**

Create `docker-compose.test.yml`:
```yaml
version: '3.8'

services:
  postgres-test:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: helixagent
      POSTGRES_PASSWORD: helixagent123
      POSTGRES_DB: helixagent_test
    ports:
      - "15432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U helixagent"]
      interval: 5s
      timeout: 5s
      retries: 5

  redis-test:
    image: redis:7-alpine
    command: redis-server --requirepass helixagent123
    ports:
      - "16379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "-a", "helixagent123", "ping"]
      interval: 5s
      timeout: 5s
      retries: 5

  mock-llm:
    build:
      context: .
      dockerfile: tests/mock-llm/Dockerfile
    ports:
      - "18080:8080"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 5s
      timeout: 5s
      retries: 5
```

**Step 4.3: Update Makefile**

```makefile
# Test targets
.PHONY: test-infra-start test-infra-stop test-with-infra

test-infra-start:
	docker-compose -f docker-compose.test.yml up -d
	@echo "Waiting for services to be healthy..."
	@sleep 10

test-infra-stop:
	docker-compose -f docker-compose.test.yml down

test-with-infra: test-infra-start
	DB_HOST=localhost DB_PORT=15432 DB_USER=helixagent DB_PASSWORD=helixagent123 DB_NAME=helixagent_test \
	REDIS_HOST=localhost REDIS_PORT=16379 REDIS_PASSWORD=helixagent123 \
	go test -v -tags=integration,e2e ./...
	$(MAKE) test-infra-stop
```

### Week 8: Enable All Skipped Tests

**Step 4.4: Fix Flaky Tests**

1. Increase timeouts where appropriate
2. Use `t.Parallel()` carefully
3. Add retry mechanisms for network tests
4. Use deterministic test data

**Step 4.5: Convert Skipped Tests to Conditional**

```go
// Before
func TestAPIIntegration(t *testing.T) {
    t.Skip("Requires API key")
}

// After
func TestAPIIntegration(t *testing.T) {
    if os.Getenv("LLM_API_KEY") == "" {
        t.Skip("Skipping: LLM_API_KEY not set")
    }
    // Test implementation
}
```

---

## Phase 5: Documentation Completion (Week 9-10)

### Week 9: Internal Package READMEs

**Step 5.1: README Template**

Create `docs/templates/INTERNAL_README_TEMPLATE.md`:
```markdown
# Package: {package_name}

## Overview

{Brief description of what this package does and its role in the system}

## Architecture

```
{package_name}/
├── {main_file}.go       # {description}
├── {secondary_file}.go  # {description}
└── {test_file}_test.go  # Unit tests
```

## Key Types

### {TypeName}

{Description}

```go
type TypeName struct {
    Field1 Type1 // Description
    Field2 Type2 // Description
}
```

## Usage

### Basic Usage

```go
import "dev.helix.agent/internal/{package_name}"

// Example usage
instance := {package_name}.New(config)
result, err := instance.DoSomething(ctx, input)
```

### Advanced Usage

```go
// Advanced example
```

## Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| Option1 | string | "" | Description |
| Option2 | int | 0 | Description |

## Testing

```bash
# Run unit tests
go test -v ./internal/{package_name}/...

# Run with coverage
go test -coverprofile=coverage.out ./internal/{package_name}/...
```

## Dependencies

### Internal Dependencies
- `internal/models` - Data models
- `internal/config` - Configuration

### External Dependencies
- `github.com/...` - Description

## See Also

- [Related Documentation](../docs/...)
- [API Reference](../api/...)
```

**Step 5.2: Create All 9 READMEs**

Execute for each package:
```bash
# List of packages needing READMEs
packages=(
    "background"
    "notifications"
    "plugins"
    "optimization"
    "structured"
    "benchmark"
    "testing"
    "verifier"
    "routing"
)

for pkg in "${packages[@]}"; do
    # Create README using template
    cp docs/templates/INTERNAL_README_TEMPLATE.md internal/$pkg/README.md
    # Update placeholders
    sed -i "s/{package_name}/$pkg/g" internal/$pkg/README.md
done
```

### Week 10: Root Documentation

**Step 5.3: Create CONTRIBUTING.md**

```markdown
# Contributing to HelixAgent

## Code of Conduct

We are committed to providing a welcoming and inspiring community for all.

## Getting Started

### Prerequisites

- Go 1.24+
- Docker and Docker Compose
- Make

### Development Setup

```bash
# Clone the repository
git clone https://github.com/your-org/HelixAgent.git
cd HelixAgent

# Install dependencies
make deps

# Install development tools
make install-deps

# Run tests
make test
```

## Development Workflow

### 1. Create a Branch

```bash
git checkout -b feature/your-feature-name
```

### 2. Make Changes

- Write code following our style guide
- Add tests for new functionality
- Update documentation

### 3. Run Tests

```bash
make test          # Unit tests
make lint          # Code quality
make test-all      # Full test suite
```

### 4. Submit Pull Request

- Ensure all tests pass
- Ensure documentation is updated
- Request review from maintainers

## Code Style

### Go Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Use meaningful variable names
- Add comments for exported functions

### Testing Requirements

- All new code must have tests
- Maintain 80%+ coverage
- Use table-driven tests
- Test error cases

## Documentation

- Update README for new features
- Add inline documentation for complex code
- Update CHANGELOG for notable changes
```

**Step 5.4: Create DEVELOPER_GUIDE.md**

```markdown
# HelixAgent Developer Guide

## Architecture Overview

### System Components

```
┌─────────────────────────────────────────────────────────┐
│                      API Gateway                         │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐   │
│  │  REST   │  │  gRPC   │  │   MCP   │  │   LSP   │   │
│  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘   │
└───────┼────────────┼────────────┼────────────┼─────────┘
        │            │            │            │
┌───────▼────────────▼────────────▼────────────▼─────────┐
│                    Service Layer                        │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐   │
│  │ Debate  │  │Ensemble │  │ Intent  │  │ Context │   │
│  │ Service │  │ Service │  │Classifier│ │ Manager │   │
│  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘   │
└───────┼────────────┼────────────┼────────────┼─────────┘
        │            │            │            │
┌───────▼────────────▼────────────▼────────────▼─────────┐
│                   Provider Layer                        │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐   │
│  │ Claude  │  │DeepSeek │  │ Gemini  │  │   ...   │   │
│  └─────────┘  └─────────┘  └─────────┘  └─────────┘   │
└─────────────────────────────────────────────────────────┘
```

## Getting Started

### 1. Understanding the Codebase

```bash
# Key directories
internal/           # Core application code
  llm/              # LLM provider implementations
  services/         # Business logic
  handlers/         # HTTP handlers
  middleware/       # Request middleware
cmd/                # Entry points
tests/              # Test suites
docs/               # Documentation
```

### 2. Running Locally

```bash
# Start infrastructure
docker-compose up -d postgres redis

# Run the application
make run-dev

# Access API
curl http://localhost:8080/health
```

### 3. Adding a New Provider

1. Create provider package:
   ```bash
   mkdir -p internal/llm/providers/newprovider
   ```

2. Implement interface:
   ```go
   // internal/llm/providers/newprovider/provider.go
   package newprovider

   type Provider struct {
       config Config
   }

   func (p *Provider) Complete(ctx context.Context, req *llm.CompletionRequest) (*llm.CompletionResponse, error) {
       // Implementation
   }
   ```

3. Register in provider registry:
   ```go
   // internal/services/provider_registry.go
   registry.Register("newprovider", newprovider.New)
   ```

4. Add tests:
   ```go
   // internal/llm/providers/newprovider/provider_test.go
   func TestProvider_Complete(t *testing.T) {
       // Tests
   }
   ```

## Debugging

### Common Issues

1. **Connection refused**
   - Check if services are running: `docker-compose ps`
   - Verify environment variables

2. **Test failures**
   - Ensure test infrastructure: `make test-infra-start`
   - Check logs: `docker-compose logs`

3. **Build errors**
   - Update dependencies: `go mod tidy`
   - Check Go version: `go version`

## Best Practices

1. **Error Handling**
   - Always wrap errors with context
   - Use structured logging
   - Return meaningful error messages

2. **Testing**
   - Write tests first (TDD)
   - Use table-driven tests
   - Mock external dependencies

3. **Documentation**
   - Document public APIs
   - Update README for changes
   - Add inline comments for complex logic
```

---

## Phase 6: Video Course Completion (Week 11-12)

### Week 11: Create Missing Labs

**Step 6.1: Create LAB_04_MCP_INTEGRATION.md**

```markdown
# Lab 4: MCP Protocol Integration

## Learning Objectives

By the end of this lab, you will be able to:
- Understand the Model Context Protocol (MCP)
- Configure MCP servers in HelixAgent
- Register and use MCP tools
- Test MCP integrations

## Prerequisites

- Completed Labs 1-3
- HelixAgent running locally
- Docker installed
- Basic understanding of protocols

## Duration

Estimated time: 2.5 hours

---

## Part 1: Understanding MCP (30 minutes)

### What is MCP?

The Model Context Protocol (MCP) is a standardized protocol for connecting AI models with external tools and data sources.

### MCP Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Client    │────▶│ MCP Server  │────▶│   Tools     │
│ (HelixAgent)│     │             │     │ (External)  │
└─────────────┘     └─────────────┘     └─────────────┘
```

### Key Concepts

1. **Tools**: Functions that can be called by the AI
2. **Resources**: Data sources accessible to the AI
3. **Prompts**: Pre-defined prompt templates

## Part 2: Setting Up MCP Server (30 minutes)

### Exercise 2.1: Start MCP Server

```bash
# Start the built-in MCP server
cd /run/media/milosvasic/DATA4TB/Projects/HelixAgent
./bin/helixagent mcp-server --port 8081
```

### Exercise 2.2: Configure Client

```yaml
# config/mcp.yaml
mcp:
  servers:
    - name: "helixagent-mcp"
      url: "http://localhost:8081"
      tools:
        - read_file
        - write_file
        - execute_command
```

## Part 3: Registering Tools (45 minutes)

### Exercise 3.1: Define a Custom Tool

```go
// Create a new tool definition
tool := &mcp.Tool{
    Name:        "search_codebase",
    Description: "Search the codebase for patterns",
    Parameters: mcp.Parameters{
        Type: "object",
        Properties: map[string]mcp.Property{
            "pattern": {
                Type:        "string",
                Description: "Search pattern (regex)",
            },
            "path": {
                Type:        "string",
                Description: "Directory to search",
                Default:     ".",
            },
        },
        Required: []string{"pattern"},
    },
}
```

### Exercise 3.2: Implement Tool Handler

```go
func handleSearchCodebase(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    pattern := params["pattern"].(string)
    path := params["path"].(string)

    // Implementation
    results, err := searchFiles(path, pattern)
    if err != nil {
        return nil, err
    }

    return results, nil
}
```

### Exercise 3.3: Register Tool

```go
server.RegisterTool(tool, handleSearchCodebase)
```

## Part 4: Testing Integration (45 minutes)

### Exercise 4.1: Test Tool Execution

```bash
# Use curl to test the tool
curl -X POST http://localhost:8081/tools/execute \
  -H "Content-Type: application/json" \
  -d '{
    "name": "search_codebase",
    "parameters": {
      "pattern": "func Test",
      "path": "./internal"
    }
  }'
```

### Exercise 4.2: Integration Test

```go
func TestMCPIntegration(t *testing.T) {
    client := mcp.NewClient("http://localhost:8081")

    result, err := client.ExecuteTool(ctx, "search_codebase", map[string]interface{}{
        "pattern": "func Test",
    })
    require.NoError(t, err)
    assert.NotEmpty(t, result)
}
```

## Verification Checklist

- [ ] MCP server starts successfully
- [ ] Custom tool is registered
- [ ] Tool execution returns expected results
- [ ] Integration tests pass

## Summary

In this lab, you learned:
- MCP protocol fundamentals
- How to set up an MCP server
- How to register and implement tools
- How to test MCP integrations

## Next Steps

- Lab 5: Production Deployment
- Explore advanced MCP features
- Build custom tool suites
```

**Step 6.2: Create LAB_05_PRODUCTION_DEPLOYMENT.md**

```markdown
# Lab 5: Production Deployment

## Learning Objectives

By the end of this lab, you will be able to:
- Configure HelixAgent for production
- Deploy to Kubernetes
- Set up monitoring and alerting
- Implement security best practices

## Prerequisites

- Completed Labs 1-4
- Kubernetes cluster access
- kubectl configured
- Helm 3.x installed

## Duration

Estimated time: 2.5 hours

---

## Part 1: Production Configuration (30 minutes)

### Exercise 1.1: Create Production Config

```yaml
# configs/production.yaml
server:
  port: 8080
  gin_mode: release
  read_timeout: 30s
  write_timeout: 30s

database:
  host: ${DB_HOST}
  port: ${DB_PORT}
  user: ${DB_USER}
  password: ${DB_PASSWORD}
  database: ${DB_NAME}
  max_connections: 100
  ssl_mode: require

redis:
  host: ${REDIS_HOST}
  port: ${REDIS_PORT}
  password: ${REDIS_PASSWORD}
  tls: true

security:
  jwt_secret: ${JWT_SECRET}
  api_key_rotation: 90d
  rate_limit: 1000

monitoring:
  enabled: true
  prometheus_port: 9090
  trace_sampling: 0.1
```

### Exercise 1.2: Environment Variables

```bash
# .env.production
DB_HOST=postgres.production.svc.cluster.local
DB_PORT=5432
DB_USER=helixagent
DB_PASSWORD=<secure-password>
DB_NAME=helixagent

REDIS_HOST=redis.production.svc.cluster.local
REDIS_PORT=6379
REDIS_PASSWORD=<secure-password>

JWT_SECRET=<256-bit-secret>
```

## Part 2: Kubernetes Deployment (45 minutes)

### Exercise 2.1: Create Helm Chart

```yaml
# helm/helixagent/values.yaml
replicaCount: 3

image:
  repository: helixagent
  tag: latest
  pullPolicy: IfNotPresent

resources:
  requests:
    memory: "512Mi"
    cpu: "250m"
  limits:
    memory: "2Gi"
    cpu: "1000m"

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilization: 70

ingress:
  enabled: true
  className: nginx
  hosts:
    - host: api.helixagent.example.com
      paths:
        - path: /
          pathType: Prefix
```

### Exercise 2.2: Deploy

```bash
# Create namespace
kubectl create namespace helixagent

# Create secrets
kubectl create secret generic helixagent-secrets \
  --from-env-file=.env.production \
  -n helixagent

# Deploy with Helm
helm install helixagent ./helm/helixagent \
  -n helixagent \
  -f configs/production.yaml
```

## Part 3: Monitoring Setup (45 minutes)

### Exercise 3.1: Prometheus Configuration

```yaml
# monitoring/prometheus-config.yaml
scrape_configs:
  - job_name: 'helixagent'
    kubernetes_sd_configs:
      - role: pod
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_label_app]
        regex: helixagent
        action: keep
```

### Exercise 3.2: Grafana Dashboard

```json
{
  "dashboard": {
    "title": "HelixAgent Metrics",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total{app=\"helixagent\"}[5m])"
          }
        ]
      },
      {
        "title": "Latency P99",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.99, rate(http_request_duration_seconds_bucket{app=\"helixagent\"}[5m]))"
          }
        ]
      }
    ]
  }
}
```

## Part 4: Security Hardening (30 minutes)

### Exercise 4.1: Network Policies

```yaml
# network-policy.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: helixagent-network-policy
spec:
  podSelector:
    matchLabels:
      app: helixagent
  ingress:
    - from:
        - podSelector:
            matchLabels:
              app: ingress-nginx
      ports:
        - port: 8080
  egress:
    - to:
        - podSelector:
            matchLabels:
              app: postgres
      ports:
        - port: 5432
    - to:
        - podSelector:
            matchLabels:
              app: redis
      ports:
        - port: 6379
```

### Exercise 4.2: Pod Security

```yaml
# pod-security.yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  fsGroup: 1000
  seccompProfile:
    type: RuntimeDefault

containers:
  - name: helixagent
    securityContext:
      allowPrivilegeEscalation: false
      readOnlyRootFilesystem: true
      capabilities:
        drop:
          - ALL
```

## Verification Checklist

- [ ] Production config validated
- [ ] Kubernetes deployment successful
- [ ] Monitoring dashboard working
- [ ] Security policies applied
- [ ] Load test passed

## Summary

In this lab, you learned:
- Production configuration best practices
- Kubernetes deployment with Helm
- Monitoring setup with Prometheus/Grafana
- Security hardening techniques

## Next Steps

- Configure CI/CD pipeline
- Set up disaster recovery
- Implement blue-green deployments
```

### Week 12: Create Missing Quizzes and Complete Slides

**Step 6.3: Create QUIZ_MODULE_7_9.md**

```markdown
# Quiz: Modules 7-9 (Advanced Features)

## Instructions

- 30 questions total
- 60 minutes to complete
- Passing score: 80%

---

## Section 1: Observability (Questions 1-10)

### Q1. What is OpenTelemetry?

A) A logging framework
B) A distributed tracing standard
C) An observability framework combining traces, metrics, and logs
D) A monitoring dashboard

**Answer: C**

### Q2. Which component is responsible for collecting traces in HelixAgent?

A) Prometheus
B) Jaeger
C) OpenTelemetry Collector
D) Grafana

**Answer: C**

### Q3. What is the purpose of trace sampling?

A) To reduce storage costs
B) To improve query performance
C) To reduce overhead while maintaining visibility
D) All of the above

**Answer: D**

[... Continue with Q4-Q10 ...]

---

## Section 2: RAG System (Questions 11-20)

### Q11. What does RAG stand for?

A) Retrieval Augmented Generation
B) Random Access Generation
C) Rapid Answer Generator
D) Recursive Answer Graph

**Answer: A**

### Q12. What is hybrid retrieval?

A) Using only dense vectors
B) Combining dense and sparse retrieval methods
C) Using only keyword search
D) Combining SQL and NoSQL

**Answer: B**

[... Continue with Q13-Q20 ...]

---

## Section 3: Memory Management (Questions 21-30)

### Q21. What is Mem0-style memory?

A) In-memory caching
B) Persistent memory with entity graphs
C) Database storage
D) File system storage

**Answer: B**

### Q22. What are entity graphs used for?

A) Visualizing data
B) Representing relationships between entities
C) Database optimization
D) API documentation

**Answer: B**

[... Continue with Q23-Q30 ...]

---

## Answer Key

| Q | A | Q | A | Q | A |
|---|---|---|---|---|---|
| 1 | C | 11 | A | 21 | B |
| 2 | C | 12 | B | 22 | B |
| 3 | D | 13 | - | 23 | - |
| ... | ... | ... | ... | ... | ... |
```

---

## Phase 7: Website Content Update (Week 13)

### Step 7.1: Update Features Section

Update `Website/features.html` with new features:

- AI Debate Orchestrator Framework
- Semantic Routing System
- RAG Hybrid Retrieval
- Memory Management
- LLM Testing Framework
- Security Red Team Framework

### Step 7.2: Update Documentation Links

Ensure all documentation links point to correct files:

```bash
# Find all documentation links
grep -r "docs/" Website/ --include="*.html" --include="*.md"

# Verify links are valid
for link in $(grep -roh "docs/[^\"']*" Website/); do
    if [ ! -f "$link" ]; then
        echo "Broken link: $link"
    fi
done
```

### Step 7.3: Update Video Courses Section

Add new labs and quizzes to the courses listing.

---

## Phase 8: Final Verification (Week 14)

### Step 8.1: Run Complete Test Suite

```bash
# Start infrastructure
make test-infra-start

# Run all test types
make test-unit
make test-integration
make test-e2e
make test-security
make test-bench
make test-chaos

# Generate coverage report
make test-coverage

# Stop infrastructure
make test-infra-stop
```

### Step 8.2: Verify Coverage

```bash
# Check all packages have 80%+ coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep -v "100.0%" | awk '{if ($3+0 < 80) print $1, $3}'

# Should output nothing (all packages >= 80%)
```

### Step 8.3: Verify No Skipped Tests

```bash
# Run all tests and check for skips
go test -v ./... 2>&1 | grep -c "SKIP"
# Should output 0
```

### Step 8.4: Verify Documentation

```bash
# Check all READMEs exist
for dir in internal/*/; do
    if [ ! -f "${dir}README.md" ]; then
        echo "Missing README: ${dir}"
    fi
done

# Check CONTRIBUTING.md
[ -f CONTRIBUTING.md ] && echo "CONTRIBUTING.md exists" || echo "CONTRIBUTING.md missing"

# Check DEVELOPER_GUIDE.md
[ -f docs/DEVELOPER_GUIDE.md ] && echo "DEVELOPER_GUIDE.md exists" || echo "DEVELOPER_GUIDE.md missing"
```

### Step 8.5: Final Sign-off Checklist

```markdown
# Final Sign-off Checklist

## Test Coverage
- [ ] All packages have >= 80% coverage
- [ ] Coverage report generated and reviewed

## Tests
- [ ] All unit tests passing
- [ ] All integration tests passing
- [ ] All e2e tests passing
- [ ] All security tests passing
- [ ] All performance tests passing
- [ ] All chaos tests passing
- [ ] Zero skipped tests

## Documentation
- [ ] All internal READMEs created
- [ ] CONTRIBUTING.md created
- [ ] DEVELOPER_GUIDE.md created
- [ ] All feature documentation complete
- [ ] CLAUDE.md updated

## Video Courses
- [ ] LAB_04 created
- [ ] LAB_05 created
- [ ] QUIZ_MODULE_7_9 created
- [ ] QUIZ_MODULE_10_11 created
- [ ] All module slides complete

## Website
- [ ] Features section updated
- [ ] Documentation links working
- [ ] Video courses section updated

## Sign-off
- [ ] Technical Lead Approval
- [ ] QA Lead Approval
- [ ] Documentation Lead Approval
```

---

## Appendix: Quick Reference

### Test Commands

```bash
# Unit tests
go test -v -short ./internal/...

# Integration tests
go test -v -tags=integration ./tests/integration/...

# E2E tests
go test -v -tags=e2e ./tests/e2e/...

# Security tests
go test -v -tags=security ./tests/security/...

# Performance benchmarks
go test -bench=. ./...

# Chaos tests
go test -v -tags=challenge ./tests/challenge/...

# Coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Documentation Commands

```bash
# Generate godoc
godoc -http=:6060

# Check markdown links
find . -name "*.md" -exec markdown-link-check {} \;

# Generate architecture diagrams
mermaid -i docs/architecture.mmd -o docs/architecture.png
```

### Infrastructure Commands

```bash
# Start test infrastructure
docker-compose -f docker-compose.test.yml up -d

# Check service health
docker-compose -f docker-compose.test.yml ps

# View logs
docker-compose -f docker-compose.test.yml logs -f

# Stop infrastructure
docker-compose -f docker-compose.test.yml down -v
```
