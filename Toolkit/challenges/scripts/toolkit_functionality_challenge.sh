#!/usr/bin/env bash
# toolkit_functionality_challenge.sh - Validates Toolkit module core functionality and structure
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MODULE_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"
MODULE_NAME="Toolkit"

PASS=0
FAIL=0
TOTAL=0

pass() { PASS=$((PASS+1)); TOTAL=$((TOTAL+1)); echo "  PASS: $1"; }
fail() { FAIL=$((FAIL+1)); TOTAL=$((TOTAL+1)); echo "  FAIL: $1"; }

echo "=== ${MODULE_NAME} Functionality Challenge ==="
echo ""

# Test 1: Core package structure exists
echo "Test: Core package structure exists"
dirs_ok=true
for dir in pkg/toolkit Commons Providers; do
    if [ ! -d "${MODULE_DIR}/${dir}" ]; then
        fail "Missing directory: ${dir}"
        dirs_ok=false
    fi
done
if [ "$dirs_ok" = true ]; then
    pass "Core directories present (pkg/toolkit, Commons, Providers)"
fi

# Test 2: Commons utility packages exist
echo "Test: Commons utility packages exist"
commons_ok=true
for pkg in auth config discovery errors http ratelimit response testing; do
    if [ ! -d "${MODULE_DIR}/Commons/${pkg}" ]; then
        fail "Missing Commons package: ${pkg}"
        commons_ok=false
    fi
done
if [ "$commons_ok" = true ]; then
    pass "All Commons packages present (auth, config, discovery, errors, http, ratelimit, response, testing)"
fi

# Test 3: Provider implementations exist
echo "Test: Provider implementations exist"
provider_count=$(find "${MODULE_DIR}/Providers" -mindepth 1 -maxdepth 1 -type d | wc -l)
if [ "$provider_count" -ge 1 ]; then
    pass "Found ${provider_count} provider implementation(s)"
else
    fail "No provider implementations found"
fi

# Test 4: Agent types exist
echo "Test: Agent types exist"
if [ -d "${MODULE_DIR}/pkg/toolkit/agents" ] && grep -rq "Agent\|agent" "${MODULE_DIR}/pkg/toolkit/agents/"; then
    pass "Agent types found in pkg/toolkit/agents"
else
    fail "No agent types found"
fi

# Test 5: Discovery service exists
echo "Test: Discovery service exists"
if grep -rq "type DiscoveryService struct\|DiscoveryService\|BaseDiscovery" "${MODULE_DIR}/Commons/discovery/"; then
    pass "Discovery service found in Commons/discovery"
else
    fail "Discovery service not found"
fi

# Test 6: HTTP client utilities exist
echo "Test: HTTP client utilities exist"
if grep -rq "type Client struct\|Client\|Request\|Response" "${MODULE_DIR}/Commons/http/"; then
    pass "HTTP client utilities found in Commons/http"
else
    fail "No HTTP client utilities found"
fi

# Test 7: Rate limiter support exists
echo "Test: Rate limiter support exists"
if grep -rq "TokenBucket\|RateLimit\|Limiter" "${MODULE_DIR}/Commons/ratelimit/"; then
    pass "Rate limiter support found in Commons/ratelimit"
else
    fail "No rate limiter support found"
fi

# Test 8: Auth utilities exist
echo "Test: Auth utilities exist"
if grep -rq "Auth\|Token\|Key\|Bearer" "${MODULE_DIR}/Commons/auth/"; then
    pass "Auth utilities found in Commons/auth"
else
    fail "No auth utilities found"
fi

# Test 9: Error handling utilities exist
echo "Test: Error handling utilities exist"
if grep -rq "Error\|error\|Wrap" "${MODULE_DIR}/Commons/errors/"; then
    pass "Error handling utilities found in Commons/errors"
else
    fail "No error handling utilities found"
fi

# Test 10: Config utilities exist
echo "Test: Config utilities exist"
if grep -rq "Config\|config\|Load\|Parse" "${MODULE_DIR}/Commons/config/"; then
    pass "Config utilities found in Commons/config"
else
    fail "No config utilities found"
fi

# Test 11: Interfaces defined in toolkit package
echo "Test: Core interfaces defined"
if grep -rq "interface" "${MODULE_DIR}/pkg/toolkit/interfaces.go" 2>/dev/null; then
    pass "Core interfaces found in pkg/toolkit/interfaces.go"
else
    fail "No core interfaces found in pkg/toolkit/interfaces.go"
fi

# Test 12: Testing utilities exist
echo "Test: Testing utilities exist"
if grep -rq "Mock\|mock\|Test\|Helper" "${MODULE_DIR}/Commons/testing/"; then
    pass "Testing utilities found in Commons/testing"
else
    fail "No testing utilities found"
fi

echo ""
echo "=== Results: ${PASS}/${TOTAL} passed, ${FAIL} failed ==="
[ "${FAIL}" -eq 0 ] && exit 0 || exit 1
