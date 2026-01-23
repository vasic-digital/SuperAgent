# Testing Package

The testing package provides a comprehensive test framework for HelixAgent, supporting multiple test types and execution strategies.

## Overview

This package implements the TestBankFramework, a flexible test orchestration system that manages unit tests, integration tests, end-to-end tests, stress tests, and security tests across the entire HelixAgent codebase.

## Key Components

### TestBankFramework

```go
type TestBankFramework struct {
    suites  map[TestType]*TestSuite
    results map[TestType][]TestResult
}
```

Central framework managing all test execution and reporting.

### Test Types

```go
const (
    UnitTest        TestType = "unit"
    IntegrationTest TestType = "integration"
    E2ETest         TestType = "e2e"
    StressTest      TestType = "stress"
    SecurityTest    TestType = "security"
    StandaloneTest  TestType = "standalone"
)
```

### TestSuite

```go
type TestSuite struct {
    Name   string
    Type   TestType
    Tests  []TestCase
    Config TestConfig
}
```

### TestCase

```go
type TestCase struct {
    Name        string
    Description string
    Command     string
    Args        []string
    Timeout     time.Duration
    Expected    TestResult
}
```

## Features

- **Parallel Execution**: Run tests concurrently for faster feedback
- **Coverage Tracking**: Collect and report code coverage metrics
- **Multiple Test Types**: Support for unit, integration, E2E, stress, and security tests
- **Configurable Timeouts**: Per-test and suite-level timeout configuration
- **Result Aggregation**: Unified reporting across all test types

## Configuration

```go
type TestConfig struct {
    Parallel     bool
    Coverage     bool
    Verbose      bool
    Timeout      time.Duration
    CoverageFile string
}
```

## Usage

```go
import "dev.helix.agent/internal/testing"

framework := testing.NewTestBankFramework()

// Register test suites
suite := &testing.TestSuite{
    Name: "API Tests",
    Type: testing.IntegrationTest,
    Tests: []testing.TestCase{
        {Name: "TestHealthEndpoint", Command: "go", Args: []string{"test", "-v", "-run", "TestHealth"}},
    },
}
framework.RegisterSuite(suite)

// Run all tests
results := framework.RunAll(context.Background())
```

## Test Commands

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests
make test-integration

# Run with coverage
make test-coverage

# Run stress tests
make test-stress

# Run security tests
make test-security
```

## Subpackages

### internal/testing/llm

DeepEval-style LLM testing framework with RAGAS metrics for evaluating LLM responses:
- Faithfulness scoring
- Answer relevancy
- Context precision/recall
- Hallucination detection

## Testing

```bash
go test -v ./internal/testing/...
```

## Related Packages

- `tests/` - Actual test implementations
- `internal/testing/llm` - LLM-specific testing
