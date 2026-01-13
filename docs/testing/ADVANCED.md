# Advanced Testing Guide

Advanced testing patterns and techniques for HelixAgent.

## Table of Contents

1. [Parallel Testing](#parallel-testing)
2. [Benchmark Testing](#benchmark-testing)
3. [Fuzz Testing](#fuzz-testing)
4. [Property-Based Testing](#property-based-testing)
5. [Prometheus Metrics Testing](#prometheus-metrics-testing)
6. [Database Testing](#database-testing)
7. [HTTP Handler Testing](#http-handler-testing)
8. [Async Testing](#async-testing)

## Parallel Testing

Run tests in parallel for faster execution:

```go
func TestParallel(t *testing.T) {
    t.Parallel()

    // Test logic here
}

func TestParallelGroup(t *testing.T) {
    tests := []struct {
        name string
        // ...
    }{
        {"test1", ...},
        {"test2", ...},
    }

    for _, tc := range tests {
        tc := tc // capture range variable
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()
            // Test logic
        })
    }
}
```

## Benchmark Testing

Write benchmarks to measure performance:

```go
func BenchmarkFunction(b *testing.B) {
    // Setup
    input := prepareData()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        functionToTest(input)
    }
}

func BenchmarkFunctionWithMemory(b *testing.B) {
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        functionToTest()
    }
}
```

Run benchmarks:

```bash
go test -bench=. -benchmem ./path/to/package
```

## Fuzz Testing

Go 1.18+ native fuzzing:

```go
func FuzzParser(f *testing.F) {
    // Seed corpus
    f.Add("valid input")
    f.Add("edge case")

    f.Fuzz(func(t *testing.T, input string) {
        result, err := Parse(input)
        if err != nil {
            return // Valid to return error
        }
        // Validate result invariants
        if result == nil {
            t.Fatal("nil result on success")
        }
    })
}
```

Run fuzzing:

```bash
go test -fuzz=FuzzParser -fuzztime=30s ./...
```

## Property-Based Testing

Use rapid for property-based testing:

```go
import "pgregory.net/rapid"

func TestPropertyBased(t *testing.T) {
    rapid.Check(t, func(t *rapid.T) {
        // Generate random inputs
        input := rapid.String().Draw(t, "input")

        // Test property
        result := function(input)
        if len(result) > len(input) {
            t.Fatal("result should not be longer than input")
        }
    })
}
```

## Prometheus Metrics Testing

Handle Prometheus metric registration in tests:

```go
package mypackage_test

import (
    "os"
    "testing"

    "github.com/prometheus/client_golang/prometheus"
)

func TestMain(m *testing.M) {
    // Replace default registry to avoid duplicate registration panics
    prometheus.DefaultRegisterer = prometheus.NewRegistry()
    os.Exit(m.Run())
}

func TestWithMetrics(t *testing.T) {
    // Now safe to create components that register metrics
    pool := NewWorkerPool(config)
    // ...
}
```

## Database Testing

### In-Memory Database

Use MemoryDB for unit tests:

```go
func TestWithMemoryDB(t *testing.T) {
    db := database.NewMemoryDB()
    defer db.Close()

    repo := NewRepository(db)
    // Test repository methods
}
```

### Real Database Tests

Use docker-compose for integration tests:

```go
func TestWithRealDB(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping database test in short mode")
    }

    db, err := database.Connect(testConfig)
    require.NoError(t, err)
    defer db.Close()

    // Run tests
}
```

### Transaction Rollback Pattern

```go
func TestWithTransaction(t *testing.T) {
    tx, err := db.Begin()
    require.NoError(t, err)
    defer tx.Rollback() // Always rollback

    // Insert test data
    // Verify results
    // Transaction is rolled back, no cleanup needed
}
```

## HTTP Handler Testing

### Using httptest

```go
func TestHandler(t *testing.T) {
    // Setup
    router := setupRouter()

    // Create request
    req := httptest.NewRequest("GET", "/api/health", nil)
    rec := httptest.NewRecorder()

    // Execute
    router.ServeHTTP(rec, req)

    // Assert
    assert.Equal(t, http.StatusOK, rec.Code)

    var response map[string]interface{}
    err := json.Unmarshal(rec.Body.Bytes(), &response)
    require.NoError(t, err)
    assert.Equal(t, "healthy", response["status"])
}
```

### Testing with Authentication

```go
func TestAuthenticatedHandler(t *testing.T) {
    req := httptest.NewRequest("GET", "/api/protected", nil)
    req.Header.Set("Authorization", "Bearer "+testToken)

    rec := httptest.NewRecorder()
    router.ServeHTTP(rec, req)

    assert.Equal(t, http.StatusOK, rec.Code)
}
```

## Async Testing

### Eventually Pattern

```go
func TestEventually(t *testing.T) {
    // Start async operation
    go startAsyncProcess()

    // Wait for condition
    assert.Eventually(t, func() bool {
        return checkCondition()
    }, 5*time.Second, 100*time.Millisecond)
}
```

### Channel-Based Testing

```go
func TestWithChannels(t *testing.T) {
    done := make(chan struct{})

    go func() {
        result := asyncOperation()
        assert.NotNil(t, result)
        close(done)
    }()

    select {
    case <-done:
        // Success
    case <-time.After(5 * time.Second):
        t.Fatal("timeout waiting for async operation")
    }
}
```

### Context Cancellation

```go
func TestWithCancellation(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    result, err := operationWithContext(ctx)
    if ctx.Err() != nil {
        t.Fatal("operation timed out")
    }
    require.NoError(t, err)
}
```

## Test Helpers

### Creating Test Fixtures

```go
// testdata/fixtures.go
package testdata

func CreateTestUser(t *testing.T) *models.User {
    t.Helper()
    return &models.User{
        ID:       uuid.New().String(),
        Username: "testuser",
        Email:    "test@example.com",
    }
}
```

### Cleanup Helpers

```go
func TestWithCleanup(t *testing.T) {
    resource := createResource()
    t.Cleanup(func() {
        resource.Close()
    })

    // Test logic
}
```

### Test Builders

```go
type TestUserBuilder struct {
    user *models.User
}

func NewTestUserBuilder() *TestUserBuilder {
    return &TestUserBuilder{
        user: &models.User{
            ID:     uuid.New().String(),
            Active: true,
        },
    }
}

func (b *TestUserBuilder) WithEmail(email string) *TestUserBuilder {
    b.user.Email = email
    return b
}

func (b *TestUserBuilder) Build() *models.User {
    return b.user
}

// Usage
user := NewTestUserBuilder().WithEmail("custom@example.com").Build()
```

## Test Environment Configuration

### Environment Variables

```go
func TestMain(m *testing.M) {
    // Set test environment
    os.Setenv("GIN_MODE", "test")
    os.Setenv("LOG_LEVEL", "error")

    code := m.Run()

    // Cleanup
    os.Unsetenv("GIN_MODE")
    os.Exit(code)
}
```

### Test Configuration Files

```go
func loadTestConfig(t *testing.T) *config.Config {
    t.Helper()

    cfg, err := config.LoadFromFile("testdata/config.yaml")
    require.NoError(t, err)
    return cfg
}
```

## Mocking External Services

### HTTP Mock Server

```go
func TestWithMockServer(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"ok"}`))
    }))
    defer server.Close()

    client := NewClient(server.URL)
    result, err := client.Fetch()
    require.NoError(t, err)
    assert.Equal(t, "ok", result.Status)
}
```

### Interface Mocking

```go
type mockProvider struct {
    CompleteFunc func(ctx context.Context, req *LLMRequest) (*LLMResponse, error)
}

func (m *mockProvider) Complete(ctx context.Context, req *LLMRequest) (*LLMResponse, error) {
    if m.CompleteFunc != nil {
        return m.CompleteFunc(ctx, req)
    }
    return &LLMResponse{Content: "mock response"}, nil
}
```

## Running Tests in CI/CD

### GitHub Actions Example

```yaml
- name: Run Tests
  run: |
    go test -v -race -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out

- name: Upload Coverage
  uses: codecov/codecov-action@v3
  with:
    files: ./coverage.out
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "Running tests..."
go test -short ./... || exit 1

echo "Running race detector..."
go test -race -short ./... || exit 1

echo "All tests passed!"
```
