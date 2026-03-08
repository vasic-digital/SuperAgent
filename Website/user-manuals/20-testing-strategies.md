# User Manual 20: Testing Strategies

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Test Categories](#test-categories)
4. [Test Infrastructure](#test-infrastructure)
5. [Unit Testing](#unit-testing)
6. [Integration Testing](#integration-testing)
7. [End-to-End Testing](#end-to-end-testing)
8. [Security Testing](#security-testing)
9. [Stress Testing](#stress-testing)
10. [Chaos and Challenge Testing](#chaos-and-challenge-testing)
11. [Benchmark Testing](#benchmark-testing)
12. [Table-Driven Tests](#table-driven-tests)
13. [Test Naming Conventions](#test-naming-conventions)
14. [Resource Limits](#resource-limits)
15. [Make Targets Reference](#make-targets-reference)
16. [Test Architecture](#test-architecture)
17. [Troubleshooting](#troubleshooting)
18. [Related Resources](#related-resources)

## Overview

HelixAgent mandates 100% test coverage across all components. Every feature, service, and module must have unit, integration, E2E, security, stress, and benchmark tests. Mocks and stubs are permitted only in unit tests; all other test types must use real data, live services, and actual API calls.

The test suite uses the `testify` assertion library (v1.11.1), table-driven test patterns, and Go's built-in testing infrastructure. Infrastructure containers (PostgreSQL, Redis, Mock LLM) must be running before executing integration, E2E, or challenge tests.

## Prerequisites

- Go 1.24+ with `go test` available
- Docker or Podman for test infrastructure
- Infrastructure containers started via `make test-infra-start`
- `testify` v1.11.1 (included in go.mod)
- `golangci-lint` for pre-test code quality checks

## Test Categories

| Category | Location | Real Services | Mocks Allowed | Coverage Target |
|---|---|---|---|---|
| Unit | `./internal/...` | No | Yes | 100% |
| Integration | `./tests/integration/` | Yes | No | 100% |
| E2E | `./tests/e2e/` | Yes | No | 100% |
| Security | `./tests/security/` | Yes | No | 100% |
| Stress | `./tests/stress/` | Yes | No | N/A |
| Challenge | `./tests/challenge/` | Yes | No | N/A |
| Benchmark | `./internal/...` | Optional | Yes | N/A |
| Precondition | `./tests/precondition/` | Yes | No | N/A |

## Test Infrastructure

### Starting Infrastructure

Infrastructure containers must be running before integration, E2E, security, and challenge tests:

```bash
# Start test infrastructure (auto-detects Docker/Podman)
make test-infra-start

# Alternative for Podman rootless (uses --userns=host)
make test-infra-direct-start

# Verify infrastructure is running
make test-infra-status
```

### Infrastructure Ports

| Service | Port | Credentials |
|---|---|---|
| PostgreSQL | 15432 | helixagent / helixagent123 / helixagent_db |
| Redis | 16379 | helixagent123 |
| Mock LLM | 18081 | N/A |

### Environment Variables for Tests

```bash
export DB_HOST=localhost
export DB_PORT=15432
export DB_USER=helixagent
export DB_PASSWORD=helixagent123
export DB_NAME=helixagent_db
export REDIS_HOST=localhost
export REDIS_PORT=16379
export REDIS_PASSWORD=helixagent123
```

### Stopping Infrastructure

```bash
make test-infra-stop
```

## Unit Testing

Unit tests cover individual functions, methods, and types in isolation. Mocks and stubs are permitted only at this level.

### Basic Unit Test

```go
func TestProviderRegistry_Register(t *testing.T) {
    // Arrange
    registry := NewProviderRegistry()
    provider := &MockProvider{name: "test"}

    // Act
    err := registry.Register("test", provider)

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, 1, registry.Count())
}
```

### Using testify/assert

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestEnsembleVoting_MajorityVote(t *testing.T) {
    responses := []*LLMResponse{
        {Content: "answer A", Confidence: 0.9},
        {Content: "answer A", Confidence: 0.8},
        {Content: "answer B", Confidence: 0.7},
    }

    result, err := MajorityVote(responses)
    require.NoError(t, err) // fail immediately if error
    assert.Equal(t, "answer A", result.Content)
    assert.GreaterOrEqual(t, result.Confidence, 0.8)
}
```

### Running Unit Tests

```bash
# All unit tests (short mode, skips long-running tests)
make test-unit

# Specific package
go test -v -short ./internal/services/...

# Specific test
go test -v -run TestProviderRegistry_Register ./internal/services/...
```

## Integration Testing

Integration tests verify interactions between components using real infrastructure (PostgreSQL, Redis):

```go
func TestDatabaseRepository_SaveDebateSession(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }

    db := setupTestDB(t)
    defer db.Close()

    repo := NewDebateRepository(db)
    session := &DebateSession{
        ID:       uuid.New(),
        Topology: "mesh",
        Status:   "active",
    }

    err := repo.Save(context.Background(), session)
    require.NoError(t, err)

    loaded, err := repo.GetByID(context.Background(), session.ID)
    require.NoError(t, err)
    assert.Equal(t, session.Topology, loaded.Topology)
}
```

```bash
# Run integration tests (infrastructure must be running)
make test-integration
```

## End-to-End Testing

E2E tests exercise the full request path from HTTP handler through to database and back:

```go
func TestChatEndpoint_E2E(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping E2E test")
    }

    // Start a test server
    server := setupTestServer(t)
    defer server.Close()

    // Send a real HTTP request
    payload := `{"model": "helixagent-debate", "messages": [{"role": "user", "content": "Hello"}]}`
    resp, err := http.Post(server.URL+"/v1/chat/completions", "application/json", strings.NewReader(payload))
    require.NoError(t, err)
    defer resp.Body.Close()

    assert.Equal(t, http.StatusOK, resp.StatusCode)

    var result map[string]interface{}
    err = json.NewDecoder(resp.Body).Decode(&result)
    require.NoError(t, err)
    assert.Contains(t, result, "choices")
}
```

```bash
make test-e2e
```

## Security Testing

Security tests validate authentication, authorization, PII detection, input sanitization, and vulnerability scanning:

```go
func TestAuth_InvalidJWT(t *testing.T) {
    server := setupTestServer(t)
    defer server.Close()

    req, _ := http.NewRequest("GET", server.URL+"/v1/models", nil)
    req.Header.Set("Authorization", "Bearer invalid-token")

    resp, err := http.DefaultClient.Do(req)
    require.NoError(t, err)
    assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestPIIDetection_RedactsSSN(t *testing.T) {
    detector := security.NewPIIDetector()
    input := "My SSN is 123-45-6789"

    result := detector.Redact(input)
    assert.NotContains(t, result, "123-45-6789")
    assert.Contains(t, result, "[REDACTED]")
}
```

```bash
make test-security
make security-scan   # gosec static analysis
```

## Stress Testing

Stress tests verify system stability under load. All stress tests must respect resource limits (30-40% of host resources):

```go
func TestStress_ConcurrentChatRequests(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping stress test")
    }

    server := setupTestServer(t)
    defer server.Close()

    const concurrency = 50
    const requestsPerWorker = 20

    var wg sync.WaitGroup
    errors := make(chan error, concurrency*requestsPerWorker)

    for i := 0; i < concurrency; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := 0; j < requestsPerWorker; j++ {
                resp, err := sendChatRequest(server.URL)
                if err != nil {
                    errors <- err
                    continue
                }
                resp.Body.Close()
            }
        }()
    }

    wg.Wait()
    close(errors)

    var errCount int
    for range errors {
        errCount++
    }
    assert.Less(t, errCount, concurrency*requestsPerWorker/10, "error rate should be below 10%")
}
```

```bash
# Run with resource limits
GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -v -p 1 ./tests/stress/...

# Or via make target
make test-stress
```

## Chaos and Challenge Testing

Challenge tests validate real-life scenarios using shell scripts and Go-native challenge tests:

```bash
# Run all challenge tests
./challenges/scripts/run_all_challenges.sh

# Run a specific challenge
./challenges/scripts/debate_orchestrator_challenge.sh

# Go-native challenge tests
go test -v -run TestChallenge ./tests/challenge/...
```

See [User Manual 21: Challenge Development](21-challenge-development.md) for creating new challenges.

## Benchmark Testing

Benchmark tests measure performance characteristics and detect regressions:

```go
func BenchmarkEnsembleVoting_MajorityVote(b *testing.B) {
    responses := generateTestResponses(10)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        MajorityVote(responses)
    }
}

func BenchmarkProviderRegistry_Lookup(b *testing.B) {
    registry := setupRegistryWith100Providers()

    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            registry.Get("provider-50")
        }
    })
}
```

```bash
make test-bench

# Specific benchmark
go test -bench=BenchmarkEnsembleVoting -benchmem ./internal/services/...
```

## Table-Driven Tests

HelixAgent mandates table-driven tests for comprehensive scenario coverage:

```go
func TestCircuitBreaker_StateTransitions(t *testing.T) {
    tests := []struct {
        name           string
        failures       int
        threshold      int
        expectedState  State
    }{
        {
            name:          "stays closed below threshold",
            failures:      2,
            threshold:     5,
            expectedState: StateClosed,
        },
        {
            name:          "opens at threshold",
            failures:      5,
            threshold:     5,
            expectedState: StateOpen,
        },
        {
            name:          "opens above threshold",
            failures:      10,
            threshold:     5,
            expectedState: StateOpen,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cb := NewCircuitBreaker(tt.threshold, time.Minute)
            for i := 0; i < tt.failures; i++ {
                cb.RecordFailure()
            }
            assert.Equal(t, tt.expectedState, cb.State())
        })
    }
}
```

## Test Naming Conventions

Follow the pattern: `Test<Struct>_<Method>_<Scenario>`

```
TestProviderRegistry_Register_Success
TestProviderRegistry_Register_DuplicateProvider
TestProviderRegistry_Get_NotFound
TestCircuitBreaker_Execute_OpensAfterThreshold
TestCircuitBreaker_Execute_HalfOpenAfterTimeout
TestEnsembleService_Vote_MajorityConsensus
TestEnsembleService_Vote_NoConsensus
```

For benchmarks: `Benchmark<Struct>_<Method>`

For E2E: `TestE2E_<Feature>_<Scenario>`

## Resource Limits

All test execution must be limited to 30-40% of host resources to avoid system instability:

```bash
# Manual resource-limited execution
GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -v -p 1 ./...

# Container resource limits
docker run --cpus=2 --memory=4g helixagent-test
```

The `-p 1` flag limits parallel test package execution to 1 at a time. `GOMAXPROCS=2` limits Go scheduler threads. `nice -n 19` gives lowest CPU priority, and `ionice -c 3` gives idle I/O priority.

## Make Targets Reference

| Target | Description |
|---|---|
| `make test` | All tests (auto-detects infrastructure) |
| `make test-unit` | Unit tests (`./internal/... -short`) |
| `make test-integration` | Integration tests (`./tests/integration/`) |
| `make test-e2e` | End-to-end tests (`./tests/e2e/`) |
| `make test-security` | Security tests (`./tests/security/`) |
| `make test-stress` | Stress tests (`./tests/stress/`) |
| `make test-chaos` | Challenge tests (`./tests/challenge/`) |
| `make test-bench` | Benchmark tests |
| `make test-race` | All tests with race detector |
| `make test-coverage` | Coverage with HTML report |
| `make test-with-infra` | All tests with Docker infrastructure |
| `make test-infra-start` | Start PostgreSQL, Redis, Mock LLM |
| `make test-infra-stop` | Stop infrastructure containers |

## Test Architecture

```
tests/
+-- integration/          # Cross-component tests with real services
+-- e2e/                  # Full HTTP request path tests
+-- security/             # Auth, PII, input validation tests
+-- stress/               # Load and concurrency tests
+-- challenge/            # Go-native challenge tests
+-- precondition/         # Container boot verification
    +-- containers_boot_test.go

internal/
+-- services/
|   +-- *_test.go         # Unit tests alongside source
+-- handlers/
|   +-- *_test.go
+-- llm/providers/
    +-- claude/
    |   +-- claude_test.go
    +-- deepseek/
        +-- deepseek_test.go

challenges/scripts/       # Shell-based challenge scripts
+-- run_all_challenges.sh
+-- debate_orchestrator_challenge.sh
+-- ...
```

## Troubleshooting

### Tests Fail with "connection refused"

**Symptom:** Integration or E2E tests fail connecting to PostgreSQL or Redis.

**Solutions:**
1. Start infrastructure: `make test-infra-start`
2. Verify containers are running: `docker ps | grep helixagent`
3. Check environment variables are set (DB_PORT=15432, REDIS_PORT=16379)
4. For Podman rootless: use `make test-infra-direct-start`

### Race Detector Reports False Positives

**Symptom:** `-race` flag reports issues in third-party libraries.

**Solutions:**
1. Verify the race is in HelixAgent code, not a dependency
2. If in HelixAgent code, add proper synchronization (mutex, atomic, channel)
3. Run the specific test in isolation to reproduce consistently

### Tests Pass Locally but Fail in CI

**Symptom:** Tests depend on timing or resource availability.

**Solutions:**
1. Use `time.After` with generous timeouts instead of `time.Sleep`
2. Use `require.Eventually` for async assertions
3. Ensure test infrastructure ports do not conflict
4. Check resource limits are applied consistently

### Coverage Report Shows 0%

**Symptom:** `make test-coverage` generates an HTML report with no data.

**Solutions:**
1. Ensure the correct package paths are specified
2. Use `-coverprofile=coverage.out` explicitly
3. Check that tests are not all skipped (look for `t.Skip` calls)

## Related Resources

- [User Manual 21: Challenge Development](21-challenge-development.md) -- Creating new challenge tests
- [User Manual 19: Concurrency Patterns](19-concurrency-patterns.md) -- Testing concurrent code
- [User Manual 18: Performance Monitoring](18-performance-monitoring.md) -- Benchmark result analysis
- Test infrastructure: `docker/test/docker-compose.test.yml`
- Challenges framework: `Challenges/`
- Go testing documentation: https://pkg.go.dev/testing
