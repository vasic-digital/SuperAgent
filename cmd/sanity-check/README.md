# Sanity Check Tool

This package contains a sanity check utility for validating HelixAgent installations.

## Overview

The sanity check tool performs comprehensive validation of:
- Configuration files
- Environment variables
- Database connectivity
- Redis connectivity
- LLM provider availability
- API endpoint functionality

## Files

- `main.go` - Sanity check entry point
- `validation.go` - Validation logic

## Checks Performed

### Configuration Validation

- Required environment variables present
- Configuration file syntax valid
- Required fields populated
- Values within expected ranges

### Infrastructure Checks

- PostgreSQL connectivity and schema
- Redis connectivity and permissions
- File system permissions
- Network connectivity

### Provider Checks

- API key validity
- Provider endpoint accessibility
- Model availability
- Rate limit status

### Endpoint Checks

- Health endpoints responding
- API authentication working
- Basic completion request succeeds

## Usage

### Build and Run

```bash
# Build
go build -o bin/sanity-check ./cmd/sanity-check

# Run all checks
./bin/sanity-check

# Run specific checks
./bin/sanity-check --check config
./bin/sanity-check --check database
./bin/sanity-check --check providers
./bin/sanity-check --check endpoints
```

### Command Line Flags

```bash
./bin/sanity-check --help

Flags:
  --check string    Specific check to run (config|database|providers|endpoints|all)
  --config string   Configuration file path
  --verbose         Show detailed output
  --json            Output results as JSON
  --fail-fast       Stop on first failure
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `CONFIG_PATH` | Config file location | `./configs/development.yaml` |
| `SANITY_TIMEOUT` | Check timeout | `30s` |

## Output

### Normal Output

```
HelixAgent Sanity Check
=======================

Configuration:
  ✓ Environment variables loaded
  ✓ Configuration file valid
  ✓ Required fields present

Infrastructure:
  ✓ PostgreSQL connection successful
  ✓ Redis connection successful
  ✓ File permissions OK

Providers:
  ✓ Claude API accessible
  ✓ DeepSeek API accessible
  ✓ Gemini API accessible
  ✗ Mistral API key invalid

Endpoints:
  ✓ Health endpoint responding
  ✓ Models endpoint responding
  ✓ Completions endpoint working

Summary: 11/12 checks passed
```

### JSON Output

```bash
./bin/sanity-check --json

{
    "timestamp": "2026-01-22T10:30:00Z",
    "checks": {
        "config": {"status": "pass", "duration_ms": 5},
        "database": {"status": "pass", "duration_ms": 120},
        "providers": {"status": "partial", "passed": 3, "failed": 1},
        "endpoints": {"status": "pass", "duration_ms": 250}
    },
    "summary": {
        "total": 12,
        "passed": 11,
        "failed": 1,
        "status": "partial"
    }
}
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All checks passed |
| 1 | Some checks failed |
| 2 | Critical checks failed |
| 3 | Configuration error |

## Integration

### Pre-deployment Check

```bash
#!/bin/bash
# deploy.sh

# Run sanity check before deployment
if ! ./bin/sanity-check --fail-fast; then
    echo "Sanity check failed, aborting deployment"
    exit 1
fi

# Proceed with deployment
kubectl apply -f k8s/
```

### CI Pipeline

```yaml
# .github/workflows/ci.yml
- name: Run Sanity Check
  run: |
    make build-sanity-check
    ./bin/sanity-check --check config --check database
```

## Testing

```bash
go test -v ./cmd/sanity-check/...
```
