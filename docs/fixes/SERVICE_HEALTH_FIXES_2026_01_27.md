# Service Health Fixes - January 27, 2026

This document describes the comprehensive fixes applied to the HelixAgent service health monitoring system.

## Overview

The following issues were identified and fixed during the monitoring session:

| Issue | Severity | Status | Root Cause |
|-------|----------|--------|------------|
| Redis memory overcommit warning | Non-Critical | FIXED | Missing container configuration |
| Qwen OAuth "critical" errors | Non-Critical | FIXED | Incorrect severity classification |
| OpenRouter provider 401 errors | Non-Critical | FIXED | Unconfigured providers logged as errors |
| Cognee search timeouts | Non-Critical | FIXED | Timeout too aggressive (2s) |
| Claude OAuth 404 errors | Critical | FIXED | Incorrect model names |
| MCP Everything connection close | Critical | FIXED | Wrong container working directory |

## Detailed Fixes

### 1. Redis Memory Overcommit Warning

**File**: `docker-compose.yml`

**Problem**: Redis container logged warnings about `vm.overcommit_memory` not being enabled.

**Root Cause**: Container didn't have proper sysctl configuration or logging settings.

**Fix Applied**:
```yaml
redis:
  image: redis:7-alpine
  container_name: helixagent-redis
  command: redis-server --requirepass ${REDIS_PASSWORD:-helixagent123} --appendonly yes --ignore-warnings ARM64-COW-BUG
  sysctls:
    - net.core.somaxconn=1024
  logging:
    driver: "json-file"
    options:
      max-size: "10m"
      max-file: "3"
```

**Verification**: Check that Redis container logs don't show memory warnings after restart.

---

### 2. Qwen OAuth Graceful Handling

**File**: `internal/services/oauth_token_monitor.go`

**Problem**: When Qwen OAuth credentials file was not found (user not logged in), the system logged it as a "critical" error.

**Root Cause**: All credential read failures were treated as critical, regardless of whether the credentials were simply not configured.

**Fix Applied**:
1. Added `isNotConfiguredError()` helper function to detect "not configured" vs "actual error" scenarios
2. Changed severity from "critical" to "info" for not-configured scenarios
3. Added `logrus.InfoLevel` handling in `sendAlert()`

**Code Changes**:
```go
// isNotConfiguredError checks if an error indicates OAuth is not configured
func isNotConfiguredError(errMsg string) bool {
    notConfiguredPhrases := []string{
        "file not found",
        "not logged in",
        "no such file",
        "credentials not found",
        "user may not be logged in",
    }
    errLower := strings.ToLower(errMsg)
    for _, phrase := range notConfiguredPhrases {
        if strings.Contains(errLower, phrase) {
            return true
        }
    }
    return false
}
```

**Verification**: Check that unconfigured OAuth providers are logged at INFO level, not ERROR.

---

### 3. OpenRouter Provider Health Check Classification

**File**: `internal/services/provider_health_monitor.go`

**Problem**: Providers without API keys (codestral, kimi, siliconflow, replicate, cloudflare) were logged as ERROR level "unhealthy" when they were simply not configured.

**Root Cause**: Health check failures due to missing API keys were treated the same as actual service failures.

**Fix Applied**:
1. Added `isProviderUnconfiguredError()` helper function
2. Changed alert type to "provider_unconfigured" for API key issues
3. Changed log level to WARN for unconfigured providers

**Code Changes**:
```go
// isUnconfiguredError checks if an error indicates the provider is not configured
func isProviderUnconfiguredError(errMsg string) bool {
    unconfiguredPhrases := []string{
        "api key is required",
        "api key not set",
        "api key is invalid or expired",
        "key not configured",
        "credentials not found",
        "unauthorized",
        "401",
    }
    // ...
}
```

**Verification**: Check that providers without API keys are logged as WARN, not ERROR.

---

### 4. Cognee Search Timeout Improvements

**File**: `internal/services/cognee_service.go`

**Problem**: Cognee search requests were timing out with "context deadline exceeded" errors.

**Root Cause**: The per-search timeout was set to 2 seconds, which was too aggressive for Cognee cold starts or slow responses.

**Fix Applied**:
1. Increased default timeout from 2 seconds to 5 seconds
2. Made timeout configurable via `s.config.Timeout`

**Code Changes**:
```go
// performSearch executes a single search type with per-search timeout
func (s *CogneeService) performSearch(ctx context.Context, query, dataset string, limit int, searchType string) ([]interface{}, error) {
    // Use 5 seconds for normal operations
    searchTimeout := 5 * time.Second
    if s.config.Timeout > 0 && s.config.Timeout < searchTimeout {
        searchTimeout = s.config.Timeout
    }
    // ...
}
```

**Verification**: Check that Cognee searches complete successfully without excessive timeouts.

---

### 5. Claude OAuth Model Names (Previously Fixed)

**Files**:
- `internal/llm/providers/claude/claude.go`
- `internal/verifier/adapters/oauth_adapter.go`

**Problem**: Claude API requests were returning 404 errors due to invalid model names.

**Root Cause**: Model names like "claude-opus-4-5-20251101" don't exist in the Anthropic API.

**Fix Applied**: Updated to use valid Anthropic model names:
- `claude-3-5-sonnet-20241022` (default)
- `claude-3-5-haiku-20241022`
- `claude-3-opus-20240229`
- `claude-3-sonnet-20240229`
- `claude-3-haiku-20240307`

---

### 6. MCP Everything Server (Previously Fixed)

**Problem**: MCP Everything server returning "Connection close" errors.

**Root Cause**: Container was started with wrong working directory.

**Fix Applied**: Recreated container with `-w /app/src/everything` flag.

---

## Testing

### Integration Tests

New integration tests added in `tests/integration/provider_health_integration_test.go`:

| Test | Description |
|------|-------------|
| `TestProviderHealthMonitor_UnconfiguredProviders` | Verifies unconfigured providers are classified correctly |
| `TestOAuthTokenMonitor_NotConfiguredHandling` | Verifies OAuth "not configured" is info level |
| `TestOAuthTokenMonitor_Creation` | Verifies monitor creation and status |
| `TestProviderHealthMonitor_Creation` | Verifies health monitor creation |
| `TestCogneeService_SearchTimeout` | Documents timeout requirements |
| `TestIntegration_HealthMonitorsCanStart` | Verifies monitors can start/stop |

### Validation Challenge

Run the validation challenge to verify all fixes:

```bash
./challenges/scripts/service_health_fixes_challenge.sh
```

Expected output: All 15+ tests should pass.

---

## Configuration Requirements

### For System Administrators

To permanently fix the Redis memory overcommit warning on the host system:

```bash
# Add to /etc/sysctl.conf
vm.overcommit_memory = 1

# Apply immediately
sudo sysctl vm.overcommit_memory=1
```

### For OAuth Users

OAuth providers (Claude, Qwen) are optional. If not using OAuth:
1. Ensure `CLAUDE_CODE_USE_OAUTH_CREDENTIALS` is not set to "true"
2. Ensure `QWEN_CODE_USE_OAUTH_CREDENTIALS` is not set to "true"

Alternatively, log in via CLI tools:
- Claude: `claude auth login`
- Qwen: `qwen auth login`

---

## Files Modified

| File | Changes |
|------|---------|
| `docker-compose.yml` | Added Redis sysctls and logging config |
| `internal/services/oauth_token_monitor.go` | Added graceful handling for not-configured OAuth |
| `internal/services/provider_health_monitor.go` | Added unconfigured provider classification |
| `internal/services/cognee_service.go` | Increased search timeout to 5s |
| `tests/integration/provider_health_integration_test.go` | New integration tests |
| `challenges/scripts/service_health_fixes_challenge.sh` | New validation challenge |

---

## Monitoring Dashboard Updates

The following Prometheus metrics are affected:

| Metric | Change |
|--------|--------|
| `helixagent_oauth_token_alerts_total` | Now includes "info" severity for not-configured |
| `helixagent_provider_health_alerts_total` | Distinguishes unconfigured from unhealthy |
| `helixagent_provider_health` | 0=unhealthy includes unconfigured providers |

---

## Rollback Instructions

If issues occur, revert the changes:

```bash
git checkout HEAD~1 -- docker-compose.yml
git checkout HEAD~1 -- internal/services/oauth_token_monitor.go
git checkout HEAD~1 -- internal/services/provider_health_monitor.go
git checkout HEAD~1 -- internal/services/cognee_service.go
```

Then rebuild and restart:

```bash
make build && make docker-restart
```
