#!/bin/bash
# ============================================================================
# Fallback Visualization Challenge
# ============================================================================
# Validates the fallback chain visualization system including:
# - Fallback chain indicators with timing
# - User-friendly fallback reasons
# - Darker timing colors
# - Multiple chained fallbacks support
# ============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Source common utilities
source "$SCRIPT_DIR/common.sh" 2>/dev/null || true

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

pass_test() {
    PASSED_TESTS=$((PASSED_TESTS + 1))
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo -e "${GREEN}✓ PASS:${NC} $1"
}

fail_test() {
    FAILED_TESTS=$((FAILED_TESTS + 1))
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo -e "${RED}✗ FAIL:${NC} $1"
}

echo "============================================================================"
echo "FALLBACK VISUALIZATION CHALLENGE"
echo "============================================================================"
echo ""
echo "Testing the fallback chain visualization system..."
echo ""

cd "$PROJECT_ROOT"

# Test 1: Run debate visualization unit tests
echo -e "${BLUE}[Test 1] Running debate visualization unit tests...${NC}"
if go test -run "TestFormat|TestFallback|TestTiming|TestComplex" ./internal/handlers/... > /dev/null 2>&1; then
    pass_test "Debate visualization unit tests pass"
else
    fail_test "Debate visualization unit tests failed"
fi

# Test 2: FormatFallbackChainIndicator exists
echo -e "${BLUE}[Test 2] Verifying FormatFallbackChainIndicator function...${NC}"
if grep -q "func FormatFallbackChainIndicator" internal/handlers/debate_visualization.go; then
    pass_test "FormatFallbackChainIndicator function exists"
else
    fail_test "FormatFallbackChainIndicator function not found"
fi

# Test 3: FormatFallbackChainWithContent exists
echo -e "${BLUE}[Test 3] Verifying FormatFallbackChainWithContent function...${NC}"
if grep -q "func FormatFallbackChainWithContent" internal/handlers/debate_visualization.go; then
    pass_test "FormatFallbackChainWithContent function exists"
else
    fail_test "FormatFallbackChainWithContent function not found"
fi

# Test 4: formatFallbackReason helper exists
echo -e "${BLUE}[Test 4] Verifying formatFallbackReason helper...${NC}"
if grep -q "func formatFallbackReason" internal/handlers/debate_visualization.go; then
    pass_test "formatFallbackReason helper exists"
else
    fail_test "formatFallbackReason helper not found"
fi

# Test 5: FallbackAttempt struct exists
echo -e "${BLUE}[Test 5] Verifying FallbackAttempt struct...${NC}"
if grep -q "type FallbackAttempt struct" internal/handlers/debate_visualization.go; then
    pass_test "FallbackAttempt struct exists"
else
    fail_test "FallbackAttempt struct not found"
fi

# Test 6: Darker timing color
echo -e "${BLUE}[Test 6] Verifying darker timing colors...${NC}"
if grep -q "ANSIDim+ANSIBrightBlack" internal/handlers/debate_visualization.go; then
    pass_test "Darker timing colors implemented"
else
    fail_test "Darker timing colors not found"
fi

# Test 7: Rate limit reason formatting
echo -e "${BLUE}[Test 7] Verifying rate limit reason formatting...${NC}"
if grep -q 'Rate limit reached' internal/handlers/debate_visualization.go; then
    pass_test "Rate limit reason formatting implemented"
else
    fail_test "Rate limit reason formatting not found"
fi

# Test 8: Timeout reason formatting
echo -e "${BLUE}[Test 8] Verifying timeout reason formatting...${NC}"
if grep -q 'return "Timeout"' internal/handlers/debate_visualization.go; then
    pass_test "Timeout reason formatting implemented"
else
    fail_test "Timeout reason formatting not found"
fi

# Test 9: Connection error reason formatting
echo -e "${BLUE}[Test 9] Verifying connection error reason formatting...${NC}"
if grep -q 'Connection error' internal/handlers/debate_visualization.go; then
    pass_test "Connection error reason formatting implemented"
else
    fail_test "Connection error reason formatting not found"
fi

# Test 10: Service unavailable reason formatting
echo -e "${BLUE}[Test 10] Verifying service unavailable reason formatting...${NC}"
if grep -q 'Service unavailable' internal/handlers/debate_visualization.go; then
    pass_test "Service unavailable reason formatting implemented"
else
    fail_test "Service unavailable reason formatting not found"
fi

# Test 11: Quota exceeded reason formatting
echo -e "${BLUE}[Test 11] Verifying quota exceeded reason formatting...${NC}"
if grep -q 'Quota exceeded' internal/handlers/debate_visualization.go; then
    pass_test "Quota exceeded reason formatting implemented"
else
    fail_test "Quota exceeded reason formatting not found"
fi

# Test 12: Service overloaded reason formatting
echo -e "${BLUE}[Test 12] Verifying service overloaded reason formatting...${NC}"
if grep -q 'Service overloaded' internal/handlers/debate_visualization.go; then
    pass_test "Service overloaded reason formatting implemented"
else
    fail_test "Service overloaded reason formatting not found"
fi

# Test 13: FormatFallbackReason tests
echo -e "${BLUE}[Test 13] Running FormatFallbackReason tests...${NC}"
if go test -run "TestFormatFallbackReason" ./internal/handlers/... > /dev/null 2>&1; then
    pass_test "FormatFallbackReason tests pass"
else
    fail_test "FormatFallbackReason tests failed"
fi

# Test 14: Timing color tests
echo -e "${BLUE}[Test 14] Running timing color tests...${NC}"
if go test -run "TestTimingColorIsDarker" ./internal/handlers/... > /dev/null 2>&1; then
    pass_test "Timing color tests pass"
else
    fail_test "Timing color tests failed"
fi

# Test 15: Complex fallback scenario tests
echo -e "${BLUE}[Test 15] Running complex fallback scenario tests...${NC}"
if go test -run "TestComplexFallbackScenarios" ./internal/handlers/... > /dev/null 2>&1; then
    pass_test "Complex fallback scenario tests pass"
else
    fail_test "Complex fallback scenario tests failed"
fi

# Test 16: Fallback chain indicator format
echo -e "${BLUE}[Test 16] Verifying fallback chain indicator format...${NC}"
if grep -q '<---' internal/handlers/debate_visualization.go && grep -q 'Fallback' internal/handlers/debate_visualization.go; then
    pass_test "Fallback chain uses proper format with <--- indicator"
else
    fail_test "Fallback chain format incorrect"
fi

# Test 17: FormatFallbackChainIndicator tests
echo -e "${BLUE}[Test 17] Running FormatFallbackChainIndicator tests...${NC}"
if go test -run "TestFormatFallbackChainIndicator" ./internal/handlers/... > /dev/null 2>&1; then
    pass_test "FormatFallbackChainIndicator tests pass"
else
    fail_test "FormatFallbackChainIndicator tests failed"
fi

# Test 18: FormatFallbackChainWithContent tests
echo -e "${BLUE}[Test 18] Running FormatFallbackChainWithContent tests...${NC}"
if go test -run "TestFormatFallbackChainWithContent" ./internal/handlers/... > /dev/null 2>&1; then
    pass_test "FormatFallbackChainWithContent tests pass"
else
    fail_test "FormatFallbackChainWithContent tests failed"
fi

# Test 19: Error message truncation
echo -e "${BLUE}[Test 19] Verifying error message truncation...${NC}"
if grep -q 'len(errorMsg) > 30' internal/handlers/debate_visualization.go; then
    pass_test "Error message truncation implemented"
else
    fail_test "Error message truncation not found"
fi

# Test 20: FallbackAttempt struct fields
echo -e "${BLUE}[Test 20] Verifying FallbackAttempt struct fields...${NC}"
if grep -q 'Provider' internal/handlers/debate_visualization.go && \
   grep -q 'Success' internal/handlers/debate_visualization.go && \
   grep -q 'Duration' internal/handlers/debate_visualization.go && \
   grep -q 'AttemptNum' internal/handlers/debate_visualization.go; then
    pass_test "FallbackAttempt struct has all required fields"
else
    fail_test "FallbackAttempt struct missing fields"
fi

# Summary
echo ""
echo "============================================================================"
echo "CHALLENGE SUMMARY"
echo "============================================================================"
echo ""
echo -e "Total Tests:  ${TOTAL_TESTS}"
echo -e "Passed:       ${GREEN}${PASSED_TESTS}${NC}"
echo -e "Failed:       ${RED}${FAILED_TESTS}${NC}"
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}============================================================================${NC}"
    echo -e "${GREEN}  ALL TESTS PASSED - Fallback Visualization Challenge Complete!${NC}"
    echo -e "${GREEN}============================================================================${NC}"
    exit 0
else
    echo -e "${RED}============================================================================${NC}"
    echo -e "${RED}  CHALLENGE FAILED - ${FAILED_TESTS} test(s) failed${NC}"
    echo -e "${RED}============================================================================${NC}"
    exit 1
fi
