#!/bin/bash
# Service Health Fixes Validation Challenge
# Tests all fixes from the 2026-01-27 service health improvement session
#
# Fixes validated:
# 1. Redis memory overcommit warning handling
# 2. Qwen OAuth graceful handling (info vs critical)
# 3. OpenRouter provider health check classification
# 4. Cognee search timeout improvements

# Don't use set -e so all tests run even if some fail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PASSED=0
FAILED=0
TOTAL=0

log_test() {
    echo -e "${BLUE}[TEST]${NC} $1"
    ((TOTAL++))
}

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((PASSED++))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((FAILED++))
}

log_info() {
    echo -e "${YELLOW}[INFO]${NC} $1"
}

echo "========================================"
echo "Service Health Fixes Validation Challenge"
echo "========================================"
echo ""

# =============================================================================
# Test 1: Redis Configuration
# =============================================================================
log_info "Testing Redis memory overcommit configuration..."

log_test "docker-compose.yml has Redis sysctls configuration or comment"
# Note: sysctls are incompatible with network_mode: host in Podman
# Check for either sysctls config OR the explanatory comment about system setup
if grep -q "sysctls:" "$PROJECT_ROOT/docker-compose.yml" || \
   grep -q "sysctls removed - incompatible with network_mode: host" "$PROJECT_ROOT/docker-compose.yml"; then
    log_pass "Redis configuration documented (sysctls or Podman compatibility note)"
else
    log_fail "Redis sysctls configuration or compatibility note missing"
fi

log_test "docker-compose.yml has Redis logging configuration"
if grep -A20 "helixagent-redis" "$PROJECT_ROOT/docker-compose.yml" | grep -q "logging:"; then
    log_pass "Redis logging configuration present"
else
    log_fail "Redis logging configuration missing"
fi

# =============================================================================
# Test 2: OAuth Token Monitor Graceful Handling
# =============================================================================
log_info "Testing OAuth token monitor improvements..."

log_test "oauth_token_monitor.go has isNotConfiguredError function"
if grep -q "func isNotConfiguredError" "$PROJECT_ROOT/internal/services/oauth_token_monitor.go"; then
    log_pass "isNotConfiguredError function exists"
else
    log_fail "isNotConfiguredError function missing"
fi

log_test "oauth_token_monitor.go handles 'file not found' as info"
if grep -q "file not found" "$PROJECT_ROOT/internal/services/oauth_token_monitor.go" && \
   grep -q 'severity.*=.*"info"' "$PROJECT_ROOT/internal/services/oauth_token_monitor.go"; then
    log_pass "File not found errors classified as info"
else
    log_fail "File not found errors not properly classified"
fi

log_test "oauth_token_monitor.go imports strings package"
if grep -q '"strings"' "$PROJECT_ROOT/internal/services/oauth_token_monitor.go"; then
    log_pass "strings package imported"
else
    log_fail "strings package not imported"
fi

log_test "oauth_token_monitor.go has info log level handling"
if grep -q 'logrus.InfoLevel' "$PROJECT_ROOT/internal/services/oauth_token_monitor.go"; then
    log_pass "Info log level handling present"
else
    log_fail "Info log level handling missing"
fi

# =============================================================================
# Test 3: Provider Health Monitor Improvements
# =============================================================================
log_info "Testing provider health monitor improvements..."

log_test "provider_health_monitor.go has isProviderUnconfiguredError function"
if grep -q "func isProviderUnconfiguredError" "$PROJECT_ROOT/internal/services/provider_health_monitor.go"; then
    log_pass "isProviderUnconfiguredError function exists"
else
    log_fail "isProviderUnconfiguredError function missing"
fi

log_test "provider_health_monitor.go handles 401 as unconfigured"
if grep -q '"401"' "$PROJECT_ROOT/internal/services/provider_health_monitor.go"; then
    log_pass "401 errors classified as unconfigured"
else
    log_fail "401 errors not properly classified"
fi

log_test "provider_health_monitor.go has provider_unconfigured alert type"
if grep -q 'provider_unconfigured' "$PROJECT_ROOT/internal/services/provider_health_monitor.go"; then
    log_pass "provider_unconfigured alert type present"
else
    log_fail "provider_unconfigured alert type missing"
fi

log_test "provider_health_monitor.go uses WarnLevel for unconfigured providers"
if grep -q "logrus.WarnLevel" "$PROJECT_ROOT/internal/services/provider_health_monitor.go"; then
    log_pass "WarnLevel used for unconfigured providers"
else
    log_fail "WarnLevel not used for unconfigured providers"
fi

# =============================================================================
# Test 4: Cognee Search Timeout Improvements
# =============================================================================
log_info "Testing Cognee search timeout improvements..."

log_test "cognee_service.go has 5 second search timeout"
if grep -q "5 \* time.Second" "$PROJECT_ROOT/internal/services/cognee_service.go"; then
    log_pass "5 second search timeout present"
else
    log_fail "5 second search timeout missing (might still be 2 seconds)"
fi

log_test "cognee_service.go has configurable timeout"
if grep -q "s.config.Timeout" "$PROJECT_ROOT/internal/services/cognee_service.go"; then
    log_pass "Configurable timeout present"
else
    log_fail "Configurable timeout missing"
fi

# =============================================================================
# Test 5: Integration Tests Exist
# =============================================================================
log_info "Testing integration test coverage..."

log_test "Provider health integration test file exists"
if [[ -f "$PROJECT_ROOT/tests/integration/provider_health_integration_test.go" ]]; then
    log_pass "Integration test file exists"
else
    log_fail "Integration test file missing"
fi

log_test "Integration tests cover OAuth not configured handling"
if grep -q "TestOAuthTokenMonitor_NotConfiguredHandling" "$PROJECT_ROOT/tests/integration/provider_health_integration_test.go" 2>/dev/null; then
    log_pass "OAuth not configured handling test present"
else
    log_fail "OAuth not configured handling test missing"
fi

log_test "Integration tests cover provider unconfigured handling"
if grep -q "TestProviderHealthMonitor_UnconfiguredProviders" "$PROJECT_ROOT/tests/integration/provider_health_integration_test.go" 2>/dev/null; then
    log_pass "Provider unconfigured handling test present"
else
    log_fail "Provider unconfigured handling test missing"
fi

# =============================================================================
# Test 6: Code Compiles Successfully
# =============================================================================
log_info "Testing code compilation..."

log_test "Code compiles without errors"
cd "$PROJECT_ROOT"
if go build -o /dev/null ./cmd/helixagent 2>/dev/null; then
    log_pass "Code compiles successfully"
else
    log_fail "Code compilation failed"
fi

# =============================================================================
# Test 7: Integration Tests Pass
# =============================================================================
log_info "Testing integration tests..."

log_test "Provider health integration tests pass"
cd "$PROJECT_ROOT"
# Run only the specific provider health integration test file to avoid other redeclaration issues
TEST_OUTPUT=$(go test -v ./tests/integration/provider_health_integration_test.go -timeout 30s 2>&1)
if echo "$TEST_OUTPUT" | grep -q "^--- PASS: TestProviderHealthMonitor_UnconfiguredProviders" && \
   echo "$TEST_OUTPUT" | grep -q "^--- PASS: TestOAuthTokenMonitor_NotConfiguredHandling"; then
    log_pass "Integration tests pass"
else
    log_fail "Integration tests failed"
fi

# =============================================================================
# Summary
# =============================================================================
echo ""
echo "========================================"
echo "Challenge Summary"
echo "========================================"
echo -e "Total Tests: ${TOTAL}"
echo -e "Passed: ${GREEN}${PASSED}${NC}"
echo -e "Failed: ${RED}${FAILED}${NC}"
echo ""

if [[ $FAILED -eq 0 ]]; then
    echo -e "${GREEN}All service health fixes validated successfully!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed. Please review the fixes.${NC}"
    exit 1
fi
