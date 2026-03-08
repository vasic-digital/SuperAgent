# internal/testutil

## Overview

The testutil package provides shared test helper functions for
HelixAgent's test suites. It centralizes infrastructure availability
checks and connection configuration for test dependencies.

## Key Types

### InfraConfig

Holds connection details for the test infrastructure stack:

```go
type InfraConfig struct {
    PostgresHost string  // Default: localhost (env: DB_HOST)
    PostgresPort string  // Default: 15432   (env: DB_PORT)
    RedisHost    string  // Default: localhost (env: REDIS_HOST)
    RedisPort    string  // Default: 16379   (env: REDIS_PORT)
    MockLLMHost  string  // Default: localhost (env: MOCK_LLM_HOST)
    MockLLMPort  string  // Default: 18081   (env: MOCK_LLM_PORT)
    ServerHost   string  // Default: localhost (env: HELIXAGENT_HOST)
    ServerPort   string  // Default: 7061    (env: HELIXAGENT_PORT)
}
```

## Key Functions

| Function | Description |
|----------|-------------|
| `DefaultInfraConfig()` | Returns config from env vars with test-stack defaults |
| `RequirePostgres(t)` | Skips the test if PostgreSQL is unavailable |
| `RequireRedis(t)` | Skips the test if Redis is unavailable |
| `RequireMockLLM(t)` | Skips the test if the Mock LLM server is unavailable |
| `RequireServer(t)` | Skips the test if the HelixAgent server is unavailable |

## Usage

```go
func TestSomethingWithDB(t *testing.T) {
    testutil.RequirePostgres(t) // Skips if DB not running
    cfg := testutil.DefaultInfraConfig()
    // ... use cfg.PostgresHost, cfg.PostgresPort ...
}
```

Infrastructure availability is cached once per test run using
`sync.Once` to avoid redundant connectivity checks.
