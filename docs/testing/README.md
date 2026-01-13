# HelixAgent Testing Documentation

This documentation provides a comprehensive guide to testing the HelixAgent project.

## Table of Contents

1. [Overview](#overview)
2. [Test Structure](#test-structure)
3. [Running Tests](#running-tests)
4. [Test Categories](#test-categories)
5. [Writing Tests](#writing-tests)
6. [Coverage Goals](#coverage-goals)
7. [Best Practices](#best-practices)

## Overview

HelixAgent maintains a comprehensive test suite covering:
- Unit tests for core business logic
- Integration tests for service interactions
- E2E tests for full workflow validation
- Security tests for authentication and authorization
- Stress tests for load and performance
- Chaos tests for resilience

## Test Structure

```
tests/
├── unit/                 # Unit tests organized by package
│   ├── background/       # Worker pool and task queue tests
│   ├── notifications/    # Notification hub and polling tests
│   ├── providers/        # LLM provider tests
│   └── ...
├── integration/          # Integration tests
├── e2e/                  # End-to-end tests
├── security/             # Security tests
├── stress/               # Stress and load tests
└── challenge/            # Chaos/challenge tests

internal/                 # Package-level tests
├── cache/               # Cache layer tests
├── database/            # Database repository tests
├── handlers/            # HTTP handler tests
├── services/            # Business logic tests
└── ...
```

## Running Tests

### Basic Commands

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run with coverage report
make test-coverage

# Run integration tests
make test-integration

# Run e2e tests
make test-e2e

# Run security tests
make test-security

# Run stress tests
make test-stress

# Run chaos/challenge tests
make test-chaos

# Run benchmarks
make test-bench

# Run race detector
make test-race
```

### Running Specific Tests

```bash
# Run a single test
go test -v -run TestFunctionName ./path/to/package

# Run tests in a specific package
go test -v ./internal/cache/...

# Run tests with short flag (skip long-running tests)
go test -short ./...

# Run tests with verbose output
go test -v ./...
```

### Test Infrastructure

```bash
# Start test infrastructure (PostgreSQL, Redis, Mock LLM)
make test-infra-start

# Stop test infrastructure
make test-infra-stop

# Clean up test infrastructure
make test-infra-clean

# Run tests with infrastructure
make test-with-infra
```

## Test Categories

### Unit Tests

Unit tests focus on individual functions and methods in isolation.

**Location:** `tests/unit/` and `internal/*/`

**Examples:**
- `internal/cache/tiered_cache_test.go` - Cache layer tests
- `internal/database/memory_test.go` - In-memory database tests
- `tests/unit/background/worker_pool_test.go` - Worker pool tests

### Integration Tests

Integration tests verify service interactions.

**Location:** `tests/integration/`

**Examples:**
- Provider registry integration
- MCP server connectivity
- Database operations

### E2E Tests

End-to-end tests validate complete workflows.

**Location:** `tests/e2e/`

**Examples:**
- Full API request flow
- MCP SSE streaming
- Startup and shutdown

### Security Tests

Security tests verify authentication, authorization, and input validation.

**Location:** `tests/security/`

**Examples:**
- JWT token validation
- Rate limiting
- SQL injection prevention

### Stress Tests

Stress tests validate performance under load.

**Location:** `tests/stress/`

**Examples:**
- Concurrent request handling
- Memory usage under load
- Connection pool exhaustion

### Chaos Tests

Chaos tests validate resilience to failures.

**Location:** `tests/challenge/`

**Examples:**
- Provider failover
- Network partition recovery
- Database reconnection

## Writing Tests

### Test File Naming

- Test files must end with `_test.go`
- Place tests in the same package or `_test` package
- Name test functions with `Test` prefix

### Test Structure

```go
func TestFunctionName(t *testing.T) {
    // Arrange
    input := setupTestData()

    // Act
    result, err := functionToTest(input)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

### Table-Driven Tests

```go
func TestFunctionWithMultipleCases(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {"valid input", "hello", "HELLO", false},
        {"empty input", "", "", true},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            result, err := functionToTest(tc.input)
            if tc.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tc.expected, result)
            }
        })
    }
}
```

### Mocking

Use interfaces for dependency injection and mock implementations in tests.

```go
type mockExecutor struct {
    ExecuteFunc   func(ctx context.Context, task *models.BackgroundTask) error
    ExecuteCalled int
}

func (m *mockExecutor) Execute(ctx context.Context, task *models.BackgroundTask) error {
    m.ExecuteCalled++
    if m.ExecuteFunc != nil {
        return m.ExecuteFunc(ctx, task)
    }
    return nil
}
```

### Handling Prometheus Metrics

When testing code that registers Prometheus metrics, use TestMain to replace the default registry:

```go
func TestMain(m *testing.M) {
    prometheus.DefaultRegisterer = prometheus.NewRegistry()
    os.Exit(m.Run())
}
```

## Coverage Goals

| Package | Target | Current |
|---------|--------|---------|
| internal/cache | 90%+ | 90%+ |
| internal/database | 80%+ | 80%+ |
| internal/services | 70%+ | 67% |
| internal/handlers | 60%+ | 55% |
| tests/unit/* | 100% | 100% |

### Generating Coverage Report

```bash
# Generate HTML coverage report
make test-coverage

# View coverage in browser
open coverage.html

# Check specific package coverage
go test -coverprofile=coverage.out ./internal/cache/...
go tool cover -html=coverage.out
```

## Best Practices

### 1. Test Isolation

Each test should be independent and not rely on the state from other tests.

```go
func TestIsolated(t *testing.T) {
    // Create fresh instances for each test
    cache := NewTestCache()
    defer cache.Close()
    // ... test logic
}
```

### 2. Use TestMain for Setup/Teardown

```go
func TestMain(m *testing.M) {
    // Setup
    setupTestInfrastructure()

    code := m.Run()

    // Teardown
    teardownTestInfrastructure()

    os.Exit(code)
}
```

### 3. Use testify for Assertions

```go
import (
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestWithTestify(t *testing.T) {
    // require stops execution on failure
    require.NotNil(t, result)

    // assert continues on failure
    assert.Equal(t, expected, result)
}
```

### 4. Test Error Paths

```go
func TestErrorHandling(t *testing.T) {
    _, err := functionWithError(invalidInput)
    require.Error(t, err)
    assert.Contains(t, err.Error(), "expected error message")
}
```

### 5. Use Short Flag for Long Tests

```go
func TestLongRunning(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping long-running test in short mode")
    }
    // ... long-running test
}
```

### 6. Clean Up Resources

```go
func TestWithCleanup(t *testing.T) {
    resource := createResource()
    t.Cleanup(func() {
        resource.Close()
    })
    // ... test logic
}
```

## Continuous Integration

Tests are automatically run on:
- Pull request creation
- Push to main branch
- Scheduled nightly builds

### CI Test Matrix

| Stage | Command | Timeout |
|-------|---------|---------|
| Unit | `make test-unit` | 5m |
| Integration | `make test-integration` | 10m |
| E2E | `make test-e2e` | 15m |
| Security | `make test-security` | 5m |
| Race Detection | `make test-race` | 10m |

## Troubleshooting

### Common Issues

1. **Prometheus Metric Registration Panic**
   - Use TestMain to reset the registry
   - Use shared pools for tests that create metrics

2. **Database Connection Failures**
   - Start test infrastructure: `make test-infra-start`
   - Use memory mode for unit tests

3. **Flaky Tests**
   - Add proper waits with timeouts
   - Use eventually assertions for async operations

4. **Resource Leaks**
   - Always use defer for cleanup
   - Run with `-race` flag to detect data races
