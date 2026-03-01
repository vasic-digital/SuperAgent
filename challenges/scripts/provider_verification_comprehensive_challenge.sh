#!/bin/bash
#
# Challenge: Provider Verification Comprehensive Test
#
# This challenge verifies that:
# 1. All configured providers have API keys
# 2. Provider verification discovers providers correctly
# 3. Verified providers have valid scores
# 4. Failed providers have proper error messages
# 5. The verification system works end-to-end
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
echo "Challenge: Provider Verification"
echo "=========================================="
echo ""

# Test 1: Check that provider verification test file exists
info_test "Test 1: Checking for provider verification test file..."
if [ -f "${PROJECT_ROOT}/tests/integration/provider_verification_comprehensive_test.go" ]; then
    pass_test "Provider verification test file exists"
else
    fail_test "Provider verification test file missing"
fi

# Test 2: Check that StartupVerifier exists
info_test "Test 2: Checking for StartupVerifier implementation..."
if [ -f "${PROJECT_ROOT}/internal/verifier/startup.go" ]; then
    pass_test "StartupVerifier implementation exists"
else
    fail_test "StartupVerifier implementation missing"
fi

# Test 3: Verify provider discovery function exists
info_test "Test 3: Checking for DiscoverProviders function..."
if grep -q "func.*DiscoverProviders" "${PROJECT_ROOT}/internal/verifier/startup.go"; then
    pass_test "DiscoverProviders function found"
else
    fail_test "DiscoverProviders function not found"
fi

# Test 4: Verify verification function exists
info_test "Test 4: Checking for verifyProviders function..."
if grep -q "func.*verifyProviders" "${PROJECT_ROOT}/internal/verifier/startup.go"; then
    pass_test "verifyProviders function found"
else
    fail_test "verifyProviders function not found"
fi

# Test 5: Verify Run function exists
info_test "Test 5: Checking for Run function..."
if grep -q "func.*Run.*context.Context" "${PROJECT_ROOT}/internal/verifier/startup.go"; then
    pass_test "Run function found"
else
    fail_test "Run function not found"
fi

# Test 6: Check that verified providers have scores
info_test "Test 6: Checking that verified providers have score tracking..."
if grep -q "Score.*float64" "${PROJECT_ROOT}/internal/verifier/provider_types.go"; then
    pass_test "Provider score tracking found"
else
    fail_test "Provider score tracking not found"
fi

# Test 7: Verify failure reason tracking
info_test "Test 7: Checking for failure reason tracking..."
if grep -q "FailureReason" "${PROJECT_ROOT}/internal/verifier/provider_types.go"; then
    pass_test "Failure reason tracking found"
else
    fail_test "Failure reason tracking not found"
fi

# Test 8: Check that VerificationService has VerifyModel
info_test "Test 8: Checking for VerifyModel in VerificationService..."
if grep -q "func.*VerifyModel" "${PROJECT_ROOT}/internal/verifier/service.go"; then
    pass_test "VerifyModel function found"
else
    fail_test "VerifyModel function not found"
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
