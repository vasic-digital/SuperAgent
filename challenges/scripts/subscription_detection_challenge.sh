#!/bin/bash
# HelixAgent Challenge: Provider Subscription Detection System
# Tests: ~20 tests across 5 sections
# Validates: Subscription types, provider access registry, rate limit parsing,
#            subscription detector, and functional validation

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHALLENGES_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$CHALLENGES_DIR")"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PASSED=0
FAILED=0
TOTAL=0

pass() {
    PASSED=$((PASSED + 1))
    TOTAL=$((TOTAL + 1))
    echo -e "  ${GREEN}[PASS]${NC} $1"
}

fail() {
    FAILED=$((FAILED + 1))
    TOTAL=$((TOTAL + 1))
    echo -e "  ${RED}[FAIL]${NC} $1"
}

section() {
    echo -e "\n${BLUE}=== $1 ===${NC}"
}

#===============================================================================
# Section 1: Type System (4 tests)
#===============================================================================
section "Section 1: Type System"

# Test 1.1: SubscriptionType constants exist
if grep -q 'SubTypeFree.*SubscriptionType.*=.*"free"' "$PROJECT_ROOT/internal/verifier/subscription_types.go" && \
   grep -q 'SubTypeFreeCredits.*=.*"free_credits"' "$PROJECT_ROOT/internal/verifier/subscription_types.go" && \
   grep -q 'SubTypeFreeTier.*=.*"free_tier"' "$PROJECT_ROOT/internal/verifier/subscription_types.go" && \
   grep -q 'SubTypePayAsYouGo.*=.*"pay_as_you_go"' "$PROJECT_ROOT/internal/verifier/subscription_types.go" && \
   grep -q 'SubTypeMonthly.*=.*"monthly"' "$PROJECT_ROOT/internal/verifier/subscription_types.go" && \
   grep -q 'SubTypeEnterprise.*=.*"enterprise"' "$PROJECT_ROOT/internal/verifier/subscription_types.go"; then
    pass "SubscriptionType constants defined (6 types)"
else
    fail "Missing SubscriptionType constants"
fi

# Test 1.2: AuthMechanism struct
if grep -q 'type AuthMechanism struct' "$PROJECT_ROOT/internal/verifier/subscription_types.go" && \
   grep -q 'HeaderName' "$PROJECT_ROOT/internal/verifier/subscription_types.go" && \
   grep -q 'HeaderPrefix' "$PROJECT_ROOT/internal/verifier/subscription_types.go" && \
   grep -q 'ExtraHeaders' "$PROJECT_ROOT/internal/verifier/subscription_types.go" && \
   grep -q 'NoAuth' "$PROJECT_ROOT/internal/verifier/subscription_types.go"; then
    pass "AuthMechanism struct with required fields"
else
    fail "AuthMechanism struct incomplete"
fi

# Test 1.3: SubscriptionInfo struct
if grep -q 'type SubscriptionInfo struct' "$PROJECT_ROOT/internal/verifier/subscription_types.go" && \
   grep -q 'AvailableTiers' "$PROJECT_ROOT/internal/verifier/subscription_types.go" && \
   grep -q 'DetectionSource' "$PROJECT_ROOT/internal/verifier/subscription_types.go" && \
   grep -q 'CreditsRemaining' "$PROJECT_ROOT/internal/verifier/subscription_types.go" && \
   grep -q 'RateLimits' "$PROJECT_ROOT/internal/verifier/subscription_types.go"; then
    pass "SubscriptionInfo struct with required fields"
else
    fail "SubscriptionInfo struct incomplete"
fi

# Test 1.4: RateLimitInfo struct
if grep -q 'type RateLimitInfo struct' "$PROJECT_ROOT/internal/verifier/subscription_types.go" && \
   grep -q 'RequestsLimit' "$PROJECT_ROOT/internal/verifier/subscription_types.go" && \
   grep -q 'TokensLimit' "$PROJECT_ROOT/internal/verifier/subscription_types.go" && \
   grep -q 'DailyLimit' "$PROJECT_ROOT/internal/verifier/subscription_types.go"; then
    pass "RateLimitInfo struct with required fields"
else
    fail "RateLimitInfo struct incomplete"
fi

#===============================================================================
# Section 2: Provider Access Registry (4 tests)
#===============================================================================
section "Section 2: Provider Access Registry"

# Test 2.1: Registry file exists
if [ -f "$PROJECT_ROOT/internal/verifier/provider_access.go" ]; then
    pass "provider_access.go exists"
else
    fail "provider_access.go missing"
fi

# Test 2.2: Registry covers 22+ providers
PROVIDER_COUNT=$(grep -c '"[a-z0-9]*":.*{' "$PROJECT_ROOT/internal/verifier/provider_access.go" || echo "0")
PROVIDER_COUNT=${PROVIDER_COUNT//[^0-9]/}
PROVIDER_COUNT=${PROVIDER_COUNT:-0}
if [ "$PROVIDER_COUNT" -ge 22 ]; then
    pass "Registry has $PROVIDER_COUNT providers (>= 22)"
else
    fail "Registry has only $PROVIDER_COUNT providers (need >= 22)"
fi

# Test 2.3: Each provider has AuthMechanisms
if grep -q 'AuthMechanisms:' "$PROJECT_ROOT/internal/verifier/provider_access.go" && \
   grep -q 'PrimaryAuth:' "$PROJECT_ROOT/internal/verifier/provider_access.go"; then
    pass "Providers have AuthMechanisms and PrimaryAuth"
else
    fail "Missing AuthMechanisms or PrimaryAuth in registry"
fi

# Test 2.4: GetProviderAccessConfig function exists
if grep -q 'func GetProviderAccessConfig' "$PROJECT_ROOT/internal/verifier/provider_access.go"; then
    pass "GetProviderAccessConfig function exists"
else
    fail "GetProviderAccessConfig function missing"
fi

#===============================================================================
# Section 3: Rate Limit Parsing (3 tests)
#===============================================================================
section "Section 3: Rate Limit Parsing"

# Test 3.1: ParseRateLimitHeaders function
if grep -q 'func ParseRateLimitHeaders' "$PROJECT_ROOT/internal/verifier/rate_limit_headers.go"; then
    pass "ParseRateLimitHeaders function exists"
else
    fail "ParseRateLimitHeaders function missing"
fi

# Test 3.2: RateLimitHeaderMap has entries
HEADER_MAP_ENTRIES=$(grep -c '"[a-z]*":.*{' "$PROJECT_ROOT/internal/verifier/rate_limit_headers.go" || echo "0")
HEADER_MAP_ENTRIES=${HEADER_MAP_ENTRIES//[^0-9]/}
HEADER_MAP_ENTRIES=${HEADER_MAP_ENTRIES:-0}
if [ "$HEADER_MAP_ENTRIES" -ge 5 ]; then
    pass "RateLimitHeaderMap has $HEADER_MAP_ENTRIES entries (>= 5)"
else
    fail "RateLimitHeaderMap has only $HEADER_MAP_ENTRIES entries"
fi

# Test 3.3: Test file exists with tests
if [ -f "$PROJECT_ROOT/internal/verifier/rate_limit_headers_test.go" ]; then
    TEST_COUNT=$(grep -c 'func Test' "$PROJECT_ROOT/internal/verifier/rate_limit_headers_test.go" || echo "0")
    TEST_COUNT=${TEST_COUNT//[^0-9]/}
    TEST_COUNT=${TEST_COUNT:-0}
    if [ "$TEST_COUNT" -ge 5 ]; then
        pass "rate_limit_headers_test.go has $TEST_COUNT tests"
    else
        fail "rate_limit_headers_test.go has only $TEST_COUNT tests"
    fi
else
    fail "rate_limit_headers_test.go missing"
fi

#===============================================================================
# Section 4: Subscription Detector (4 tests)
#===============================================================================
section "Section 4: Subscription Detector"

# Test 4.1: SubscriptionDetector struct
if grep -q 'type SubscriptionDetector struct' "$PROJECT_ROOT/internal/verifier/subscription_detector.go"; then
    pass "SubscriptionDetector struct exists"
else
    fail "SubscriptionDetector struct missing"
fi

# Test 4.2: DetectSubscription method (3-tier)
if grep -q 'func.*SubscriptionDetector.*DetectSubscription' "$PROJECT_ROOT/internal/verifier/subscription_detector.go"; then
    pass "DetectSubscription method exists"
else
    fail "DetectSubscription method missing"
fi

# Test 4.3: Tier 1 API detectors exist
if grep -q 'detectOpenRouterSubscription' "$PROJECT_ROOT/internal/verifier/subscription_detector.go" && \
   grep -q 'detectCohereSubscription' "$PROJECT_ROOT/internal/verifier/subscription_detector.go"; then
    pass "Tier 1 API detectors exist (OpenRouter, Cohere)"
else
    fail "Missing Tier 1 API detectors"
fi

# Test 4.4: Test file exists with tests
if [ -f "$PROJECT_ROOT/internal/verifier/subscription_detector_test.go" ]; then
    TEST_COUNT=$(grep -c 'func Test' "$PROJECT_ROOT/internal/verifier/subscription_detector_test.go" || echo "0")
    TEST_COUNT=${TEST_COUNT//[^0-9]/}
    TEST_COUNT=${TEST_COUNT:-0}
    if [ "$TEST_COUNT" -ge 8 ]; then
        pass "subscription_detector_test.go has $TEST_COUNT tests"
    else
        fail "subscription_detector_test.go has only $TEST_COUNT tests"
    fi
else
    fail "subscription_detector_test.go missing"
fi

#===============================================================================
# Section 5: Functional Validation (5 tests)
#===============================================================================
section "Section 5: Functional Validation"

# Test 5.1: Build succeeds
echo -e "  ${YELLOW}Building verifier package...${NC}"
if (cd "$PROJECT_ROOT" && go build ./internal/verifier/ 2>&1); then
    pass "Verifier package builds successfully"
else
    fail "Verifier package build failed"
fi

# Test 5.2: All subscription tests pass
echo -e "  ${YELLOW}Running subscription tests...${NC}"
TEST_OUTPUT=$(cd "$PROJECT_ROOT" && go test -v -short -timeout 60s -run "TestSubscription|TestProviderAccess|TestRateLimit|TestParseRateLimit|TestParseHeader|TestRateLimitHeader|TestNewSubscription|TestDetect|TestInferSub|TestGetProvidersWith|TestAuth" ./internal/verifier/ 2>&1) || true
PASS_COUNT=$(echo "$TEST_OUTPUT" | grep -c "^--- PASS:" || echo "0")
PASS_COUNT=${PASS_COUNT//[^0-9]/}
PASS_COUNT=${PASS_COUNT:-0}
FAIL_COUNT=$(echo "$TEST_OUTPUT" | grep -c "^--- FAIL:" || echo "0")
FAIL_COUNT=${FAIL_COUNT//[^0-9]/}
FAIL_COUNT=${FAIL_COUNT:-0}
if [ "$PASS_COUNT" -ge 30 ] && [ "$FAIL_COUNT" -eq 0 ]; then
    pass "All subscription tests pass ($PASS_COUNT passed, $FAIL_COUNT failed)"
else
    fail "Subscription tests: $PASS_COUNT passed, $FAIL_COUNT failed"
fi

# Test 5.3: All verifier tests pass (regression check)
echo -e "  ${YELLOW}Running all verifier tests...${NC}"
VERIFIER_OUTPUT=$(cd "$PROJECT_ROOT" && go test -short -timeout 60s ./internal/verifier/ 2>&1) || true
if echo "$VERIFIER_OUTPUT" | grep -q "^ok"; then
    pass "All verifier tests pass (regression check)"
else
    fail "Some verifier tests failed"
fi

# Test 5.4: No vet errors
echo -e "  ${YELLOW}Running go vet...${NC}"
if (cd "$PROJECT_ROOT" && go vet ./internal/verifier/ 2>&1); then
    pass "No go vet errors"
else
    fail "go vet found errors"
fi

# Test 5.5: Sufficient test count
TOTAL_TESTS=$(cd "$PROJECT_ROOT" && go test -v -short -timeout 60s -run "TestSubscription|TestProviderAccess|TestRateLimit|TestParseRateLimit|TestParseHeader|TestRateLimitHeader|TestNewSubscription|TestDetect|TestInferSub|TestGetProvidersWith|TestAuth" ./internal/verifier/ 2>&1 | grep -c "^--- PASS:" || echo "0")
TOTAL_TESTS=${TOTAL_TESTS//[^0-9]/}
TOTAL_TESTS=${TOTAL_TESTS:-0}
if [ "$TOTAL_TESTS" -ge 30 ]; then
    pass "Sufficient test count: $TOTAL_TESTS tests (>= 30)"
else
    fail "Insufficient tests: $TOTAL_TESTS (need >= 30)"
fi

#===============================================================================
# Summary
#===============================================================================
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Subscription Detection Challenge${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "  Passed: ${GREEN}$PASSED${NC} / $TOTAL"
echo -e "  Failed: ${RED}$FAILED${NC} / $TOTAL"

if [ $FAILED -eq 0 ]; then
    echo -e "  ${GREEN}ALL TESTS PASSED${NC}"
    exit 0
else
    echo -e "  ${RED}$FAILED test(s) failed${NC}"
    exit 1
fi
