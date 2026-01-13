# Testing Quick Start Guide

Get started with testing HelixAgent in 5 minutes.

## Prerequisites

- Go 1.24+
- Make
- Docker (for integration tests)

## Quick Commands

```bash
# Run all unit tests
make test-unit

# Run tests with coverage
make test-coverage

# Run a specific test
go test -v -run TestName ./path/to/package
```

## Common Test Scenarios

### 1. Running Unit Tests

```bash
# All unit tests
go test ./... -short

# Specific package
go test ./internal/cache/...
go test ./internal/database/...
go test ./tests/unit/background/...
```

### 2. Running Integration Tests

```bash
# Start test infrastructure
make test-infra-start

# Run integration tests
make test-integration

# Stop infrastructure when done
make test-infra-stop
```

### 3. Running E2E Tests

```bash
# Start HelixAgent
make run-dev

# In another terminal
make test-e2e
```

### 4. Checking Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View in terminal
go tool cover -func=coverage.out

# View in browser
go tool cover -html=coverage.out
```

## Writing Your First Test

Create a file named `myfunction_test.go`:

```go
package mypackage

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestMyFunction(t *testing.T) {
    result := MyFunction("input")
    assert.Equal(t, "expected", result)
}
```

Run it:

```bash
go test -v -run TestMyFunction ./path/to/package
```

## Test Packages

| Package | Tests | Purpose |
|---------|-------|---------|
| `internal/cache/` | 54 | Cache layer |
| `internal/database/` | 702 | Database operations |
| `tests/unit/background/` | 28 | Worker pool |
| `tests/unit/notifications/` | 36 | Notifications |

## Next Steps

- Read the [full testing documentation](README.md)
- Review [advanced testing patterns](ADVANCED.md)
- Check the [test coverage report](../../coverage.html)
