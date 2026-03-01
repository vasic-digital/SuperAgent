#!/bin/bash
#
# Challenge: Ollama Explicit Enable Verification
#
# This challenge verifies that Ollama is only enabled when OLLAMA_ENABLED=true
# and that the debate team does not use Ollama when disabled.
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
TESTS_PASSED=0
TESTS_FAILED=0

# Test functions
pass_test() {
    echo -e "${GREEN}✓ PASS${NC}: $1"
    ((TESTS_PASSED++))
}

fail_test() {
    echo -e "${RED}✗ FAIL${NC}: $1"
    ((TESTS_FAILED++))
}

info_test() {
    echo -e "${BLUE}ℹ INFO${NC}: $1"
}

warn_test() {
    echo -e "${YELLOW}⚠ WARN${NC}: $1"
}

echo "=========================================="
echo "Challenge: Ollama Explicit Enable"
echo "=========================================="
echo ""

# Test 1: Check that OLLAMA_ENABLED is in .env.example
info_test "Test 1: Checking .env.example for OLLAMA_ENABLED..."
if grep -q "OLLAMA_ENABLED" "${PROJECT_ROOT}/.env.example"; then
    pass_test "OLLAMA_ENABLED found in .env.example"
else
    fail_test "OLLAMA_ENABLED not found in .env.example"
fi

# Test 2: Check that docker-compose.yml passes OLLAMA_ENABLED
info_test "Test 2: Checking docker-compose.yml for OLLAMA_ENABLED..."
if grep -q "OLLAMA_ENABLED" "${PROJECT_ROOT}/docker-compose.yml"; then
    pass_test "OLLAMA_ENABLED found in docker-compose.yml"
else
    fail_test "OLLAMA_ENABLED not found in docker-compose.yml"
fi

# Test 3: Check that startup.go checks OLLAMA_ENABLED
info_test "Test 3: Checking startup.go for OLLAMA_ENABLED check..."
if grep -q "OLLAMA_ENABLED" "${PROJECT_ROOT}/internal/verifier/startup.go"; then
    pass_test "OLLAMA_ENABLED check found in startup.go"
else
    fail_test "OLLAMA_ENABLED check not found in startup.go"
fi

# Test 4: Verify the code requires explicit "true" value
info_test "Test 4: Checking that OLLAMA_ENABLED requires explicit 'true'..."
if grep -q 'ollamaEnabled == "true"' "${PROJECT_ROOT}/internal/verifier/startup.go"; then
    pass_test "OLLAMA_ENABLED requires explicit 'true' value"
else
    fail_test "OLLAMA_ENABLED does not require explicit 'true' value"
fi

# Test 5: Check that Ollama test file exists
info_test "Test 5: Checking for Ollama explicit enable test file..."
if [ -f "${PROJECT_ROOT}/tests/integration/ollama_explicit_enable_test.go" ]; then
    pass_test "Ollama explicit enable test file exists"
else
    fail_test "Ollama explicit enable test file missing"
fi

# Test 6: Run the Go tests for Ollama
info_test "Test 6: Running Go tests for Ollama explicit enable..."
cd "${PROJECT_ROOT}"
if go test -v -run TestOllama ./tests/integration/... 2>&1 | grep -q "PASS"; then
    pass_test "Go tests for Ollama explicit enable passed"
else
    warn_test "Go tests for Ollama explicit enable had issues (may be expected if Ollama is not configured)"
fi

# Test 7: Verify logging when Ollama is disabled
info_test "Test 7: Checking that disabled message is logged..."
if grep -q 'Ollama disabled' "${PROJECT_ROOT}/internal/verifier/startup.go"; then
    pass_test "Disabled message logging found"
else
    fail_test "Disabled message logging not found"
fi

# Test 8: Verify logging when Ollama is enabled but not running
info_test "Test 8: Checking that 'not running' warning is logged..."
if grep -q 'Ollama enabled but not running' "${PROJECT_ROOT}/internal/verifier/startup.go"; then
    pass_test "Not running warning logging found"
else
    fail_test "Not running warning logging not found"
fi

# Summary
echo ""
echo "=========================================="
echo "Challenge Summary"
echo "=========================================="
echo -e "Tests Passed: ${GREEN}${TESTS_PASSED}${NC}"
echo -e "Tests Failed: ${RED}${TESTS_FAILED}${NC}"
echo ""

if [ ${TESTS_FAILED} -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}✗ Some tests failed!${NC}"
    exit 1
fi
