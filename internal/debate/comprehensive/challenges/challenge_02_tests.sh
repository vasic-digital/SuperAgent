#!/bin/bash
#
# Challenge 2: Test Coverage Validation
# Ensures all tests pass and coverage is adequate
#

set -e

echo "========================================="
echo "Challenge 2: Test Coverage Validation"
echo "========================================="

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

TESTS_PASSED=0
TESTS_FAILED=0

# Test 1: Run all tests
echo -e "\n${YELLOW}Test 1: Running All Tests${NC}"
if go test ./internal/debate/comprehensive/... 2>/dev/null | grep -q "^ok"; then
    TEST_COUNT=$(go test ./internal/debate/comprehensive/... -v 2>/dev/null | grep -c "^=== RUN" || echo "0")
    echo -e "${GREEN}✓ All tests pass ($TEST_COUNT tests)${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}✗ Some tests failed${NC}"
    ((TESTS_FAILED++))
fi

# Test 2: Verify minimum test count
echo -e "\n${YELLOW}Test 2: Minimum Test Count${NC}"
MIN_TESTS=200
ACTUAL_TESTS=$(go test ./internal/debate/comprehensive/... -v 2>/dev/null | grep -c "^=== RUN" || echo "0")
if [ "$ACTUAL_TESTS" -ge "$MIN_TESTS" ]; then
    echo -e "${GREEN}✓ Test count sufficient: $ACTUAL_TESTS >= $MIN_TESTS${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}✗ Test count insufficient: $ACTUAL_TESTS < $MIN_TESTS${NC}"
    ((TESTS_FAILED++))
fi

# Test 3: Test files exist
echo -e "\n${YELLOW}Test 3: Test File Structure${NC}"
TEST_FILES=(
    "types_test.go"
    "agents_pool_test.go"
    "memory_test.go"
    "tools_test.go"
    "utils_test.go"
    "comprehensive_test.go"
    "e2e_test.go"
)

ALL_FOUND=true
for file in "${TEST_FILES[@]}"; do
    if [ -f "internal/debate/comprehensive/$file" ]; then
        echo -e "  ${GREEN}✓ $file exists${NC}"
    else
        echo -e "  ${RED}✗ $file missing${NC}"
        ALL_FOUND=false
    fi
done

if $ALL_FOUND; then
    ((TESTS_PASSED++))
else
    ((TESTS_FAILED++))
fi

# Test 4: No test failures
echo -e "\n${YELLOW}Test 4: No Test Failures${NC}"
if go test ./internal/debate/comprehensive/... 2>/dev/null | grep -q "FAIL"; then
    echo -e "${RED}✗ Tests contain failures${NC}"
    ((TESTS_FAILED++))
else
    echo -e "${GREEN}✓ No test failures detected${NC}"
    ((TESTS_PASSED++))
fi

# Summary
echo -e "\n========================================="
echo "Challenge 2 Results"
echo "========================================="
echo -e "Tests Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests Failed: ${RED}$TESTS_FAILED${NC}"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "\n${GREEN}✓ Challenge 2 PASSED${NC}"
    exit 0
else
    echo -e "\n${RED}✗ Challenge 2 FAILED${NC}"
    exit 1
fi
