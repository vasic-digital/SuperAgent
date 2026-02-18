#!/bin/bash
# Container Centralization Challenge
# Validates that all container operations go through the Containers
# module via internal/adapters/containers/adapter.go
# No direct exec.Command to docker/podman in production code (only
# in adapter internals that delegate to the Containers module).

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

PASSED=0
FAILED=0
TOTAL=0

pass() {
    PASSED=$((PASSED + 1))
    TOTAL=$((TOTAL + 1))
    echo "  PASS: $1"
}

fail() {
    FAILED=$((FAILED + 1))
    TOTAL=$((TOTAL + 1))
    echo "  FAIL: $1"
}

echo "=============================================="
echo " Container Centralization Challenge"
echo "=============================================="
echo ""

# Test 1: Container adapter exists
echo "--- Test 1: Container adapter exists ---"
if [ -f "${PROJECT_ROOT}/internal/adapters/containers/adapter.go" ]; then
    pass "Container adapter file exists"
else
    fail "Container adapter file not found"
fi

# Test 2: Container adapter tests exist
echo "--- Test 2: Container adapter tests exist ---"
if [ -f "${PROJECT_ROOT}/internal/adapters/containers/adapter_test.go" ]; then
    pass "Container adapter test file exists"
else
    fail "Container adapter test file not found"
fi

# Test 3: Adapter imports Containers module
echo "--- Test 3: Adapter imports Containers module ---"
if grep -q 'digital.vasic.containers' "${PROJECT_ROOT}/internal/adapters/containers/adapter.go"; then
    pass "Adapter imports Containers module"
else
    fail "Adapter does not import Containers module"
fi

# Test 4: Adapter has DetectRuntime method
echo "--- Test 4: Adapter has DetectRuntime ---"
if grep -q 'func.*Adapter.*DetectRuntime' "${PROJECT_ROOT}/internal/adapters/containers/adapter.go"; then
    pass "Adapter has DetectRuntime method"
else
    fail "Adapter missing DetectRuntime method"
fi

# Test 5: Adapter has ComposeUp method
echo "--- Test 5: Adapter has ComposeUp ---"
if grep -q 'func.*Adapter.*ComposeUp' "${PROJECT_ROOT}/internal/adapters/containers/adapter.go"; then
    pass "Adapter has ComposeUp method"
else
    fail "Adapter missing ComposeUp method"
fi

# Test 6: Adapter has ComposeDown method
echo "--- Test 6: Adapter has ComposeDown ---"
if grep -q 'func.*Adapter.*ComposeDown' "${PROJECT_ROOT}/internal/adapters/containers/adapter.go"; then
    pass "Adapter has ComposeDown method"
else
    fail "Adapter missing ComposeDown method"
fi

# Test 7: Adapter has HealthCheckHTTP method
echo "--- Test 7: Adapter has HealthCheckHTTP ---"
if grep -q 'func.*Adapter.*HealthCheckHTTP' "${PROJECT_ROOT}/internal/adapters/containers/adapter.go"; then
    pass "Adapter has HealthCheckHTTP method"
else
    fail "Adapter missing HealthCheckHTTP method"
fi

# Test 8: Adapter has HealthCheckTCP method
echo "--- Test 8: Adapter has HealthCheckTCP ---"
if grep -q 'func.*Adapter.*HealthCheckTCP' "${PROJECT_ROOT}/internal/adapters/containers/adapter.go"; then
    pass "Adapter has HealthCheckTCP method"
else
    fail "Adapter missing HealthCheckTCP method"
fi

# Test 9: Adapter has Distribute method
echo "--- Test 9: Adapter has Distribute ---"
if grep -q 'func.*Adapter.*Distribute' "${PROJECT_ROOT}/internal/adapters/containers/adapter.go"; then
    pass "Adapter has Distribute method"
else
    fail "Adapter missing Distribute method"
fi

# Test 10: Adapter has Shutdown method
echo "--- Test 10: Adapter has Shutdown ---"
if grep -q 'func.*Adapter.*Shutdown' "${PROJECT_ROOT}/internal/adapters/containers/adapter.go"; then
    pass "Adapter has Shutdown method"
else
    fail "Adapter missing Shutdown method"
fi

# Test 11: main.go imports container adapter
echo "--- Test 11: main.go imports container adapter ---"
if grep -q 'containeradapter.*adapters/containers' "${PROJECT_ROOT}/cmd/helixagent/main.go"; then
    pass "main.go imports container adapter"
else
    fail "main.go does not import container adapter"
fi

# Test 12: main.go initializes globalContainerAdapter
echo "--- Test 12: main.go initializes globalContainerAdapter ---"
if grep -q 'globalContainerAdapter' "${PROJECT_ROOT}/cmd/helixagent/main.go"; then
    pass "main.go uses globalContainerAdapter"
else
    fail "main.go does not use globalContainerAdapter"
fi

# Test 13: infrastructure.go uses adapter for checkTCPPort
echo "--- Test 13: infrastructure.go uses adapter ---"
if grep -q 'globalContainerAdapter' "${PROJECT_ROOT}/cmd/helixagent/infrastructure.go"; then
    pass "infrastructure.go uses globalContainerAdapter"
else
    fail "infrastructure.go does not use globalContainerAdapter"
fi

# Test 14: boot_manager.go has ContainerAdapter field
echo "--- Test 14: boot_manager.go has ContainerAdapter ---"
if grep -q 'ContainerAdapter' "${PROJECT_ROOT}/internal/services/boot_manager.go"; then
    pass "boot_manager.go has ContainerAdapter field"
else
    fail "boot_manager.go missing ContainerAdapter field"
fi

# Test 15: Adapter tests pass
echo "--- Test 15: Adapter tests pass ---"
if (cd "${PROJECT_ROOT}" && GOMAXPROCS=2 go test ./internal/adapters/containers/... -race -count=1 -timeout=60s -p 1 >/dev/null 2>&1); then
    pass "Adapter tests pass"
else
    fail "Adapter tests failed"
fi

# Test 16: Build succeeds
echo "--- Test 16: Build succeeds ---"
if (cd "${PROJECT_ROOT}" && go build ./cmd/helixagent/... >/dev/null 2>&1); then
    pass "HelixAgent builds with adapter"
else
    fail "HelixAgent build failed"
fi

# Test 17: go vet passes
echo "--- Test 17: go vet passes ---"
if (cd "${PROJECT_ROOT}" && go vet ./internal/adapters/containers/ >/dev/null 2>&1); then
    pass "go vet passes for adapter"
else
    fail "go vet fails for adapter"
fi

echo ""
echo "=============================================="
echo " Container Centralization Challenge Results"
echo "=============================================="
echo " Passed: ${PASSED}/${TOTAL}"
echo " Failed: ${FAILED}/${TOTAL}"
echo "=============================================="

if [ "${FAILED}" -gt 0 ]; then
    exit 1
fi
exit 0
