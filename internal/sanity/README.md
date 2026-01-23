# Sanity Package

The sanity package provides boot-time sanity checks for HelixAgent, verifying system readiness before starting the main service.

## Overview

This package implements a comprehensive boot-time verification system that checks all critical dependencies and configurations before HelixAgent starts serving requests. It helps prevent startup failures and provides detailed diagnostic information.

## Key Components

### BootChecker

```go
type BootChecker struct {
    config     *BootCheckConfig
    httpClient *http.Client
    results    []CheckResult
}
```

The main checker that orchestrates all sanity checks.

### CheckResult

```go
type CheckResult struct {
    Name      string        `json:"name"`
    Category  string        `json:"category"`
    Status    CheckStatus   `json:"status"`
    Message   string        `json:"message,omitempty"`
    Details   string        `json:"details,omitempty"`
    Duration  time.Duration `json:"duration"`
    Critical  bool          `json:"critical"`
    Timestamp time.Time     `json:"timestamp"`
}
```

### Check Statuses

- `PASS` - Check completed successfully
- `FAIL` - Check failed (may be critical)
- `WARN` - Check passed with warnings
- `SKIP` - Check was skipped

## Checks Performed

### Infrastructure Checks
- PostgreSQL connectivity and migrations
- Redis connectivity and operations
- HelixAgent API availability

### External Service Checks
- Cognee service connectivity
- LLM provider availability
- MCP server connectivity

### Configuration Checks
- Required environment variables
- Valid configuration values
- Port availability

## Configuration

```go
type BootCheckConfig struct {
    HelixAgentHost     string
    HelixAgentPort     int
    PostgresHost       string
    PostgresPort       int
    RedisHost          string
    RedisPort          int
    CogneeHost         string
    CogneePort         int
    Timeout            time.Duration
    SkipExternalChecks bool
}
```

## Usage

```go
import "dev.helix.agent/internal/sanity"

config := sanity.DefaultConfig()
checker := sanity.NewBootChecker(config)

report := checker.RunAllChecks(context.Background())
if !report.ReadyToStart {
    log.Fatalf("Boot checks failed: %d critical failures", report.FailedChecks)
}
```

## Boot Check Report

```go
type BootCheckReport struct {
    Timestamp       time.Time     `json:"timestamp"`
    Duration        time.Duration `json:"duration"`
    TotalChecks     int           `json:"total_checks"`
    PassedChecks    int           `json:"passed_checks"`
    FailedChecks    int           `json:"failed_checks"`
    WarningChecks   int           `json:"warning_checks"`
    SkippedChecks   int           `json:"skipped_checks"`
    CriticalFailure bool          `json:"critical_failure"`
    Results         []CheckResult `json:"results"`
    ReadyToStart    bool          `json:"ready_to_start"`
}
```

## Testing

```bash
go test -v ./internal/sanity/...
```

## Related Packages

- `internal/config` - Configuration loading
- `internal/database` - Database connectivity
- `internal/cache` - Redis cache layer
