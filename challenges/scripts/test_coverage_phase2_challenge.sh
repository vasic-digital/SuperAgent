#!/bin/bash
# Test Coverage Challenge - Phase 2
# Validates that test coverage improvements are in place
#
# Run: ./challenges/scripts/test_coverage_phase2_challenge.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_TOTAL=0

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_test() {
    echo -e "${BLUE}[TEST]${NC} $1"
    ((TESTS_TOTAL++))
}

pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((TESTS_PASSED++))
}

fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((TESTS_FAILED++))
}

section() {
    echo ""
    echo -e "${BLUE}══════════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}══════════════════════════════════════════════════════════════${NC}"
}

# Test utilities
test_testutils() {
    section "Test Utilities Coverage"
    
    log_test "fixtures_test.go exists"
    if [ -f "$PROJECT_ROOT/tests/testutils/fixtures_test.go" ]; then
        pass "fixtures_test.go exists"
    else
        fail "fixtures_test.go not found"
    fi
    
    log_test "test_helpers_test.go exists"
    if [ -f "$PROJECT_ROOT/tests/testutils/test_helpers_test.go" ]; then
        pass "test_helpers_test.go exists"
    else
        fail "test_helpers_test.go not found"
    fi
    
    log_test "Test utilities tests pass"
    if go test -short "$PROJECT_ROOT/tests/testutils/..." 2>&1 | grep -q "PASS"; then
        pass "Test utilities tests pass"
    else
        fail "Test utilities tests failed"
    fi
    
    log_test "Test count in fixtures_test.go"
    TEST_COUNT=$(grep -c "func Test" "$PROJECT_ROOT/tests/testutils/fixtures_test.go" 2>/dev/null || echo "0")
    if [ "$TEST_COUNT" -ge 30 ]; then
        pass "fixtures_test.go has $TEST_COUNT tests (>= 30)"
    else
        fail "fixtures_test.go has only $TEST_COUNT tests (expected >= 30)"
    fi
    
    log_test "Test count in test_helpers_test.go"
    TEST_COUNT=$(grep -c "func Test" "$PROJECT_ROOT/tests/testutils/test_helpers_test.go" 2>/dev/null || echo "0")
    if [ "$TEST_COUNT" -ge 30 ]; then
        pass "test_helpers_test.go has $TEST_COUNT tests (>= 30)"
    else
        fail "test_helpers_test.go has only $TEST_COUNT tests (expected >= 30)"
    fi
}

# Formatters package
test_formatters() {
    section "Formatters Package Coverage"
    
    log_test "Native formatters tests pass"
    if go test -short "$PROJECT_ROOT/internal/formatters/providers/native/..." 2>&1 | grep -q "PASS"; then
        pass "Native formatters tests pass"
    else
        fail "Native formatters tests failed"
    fi
    
    log_test "Service formatters tests pass"
    if go test -short "$PROJECT_ROOT/internal/formatters/providers/service/..." 2>&1 | grep -q "PASS"; then
        pass "Service formatters tests pass"
    else
        fail "Service formatters tests failed"
    fi
    
    log_test "All formatters tests pass"
    if go test -short "$PROJECT_ROOT/internal/formatters/..." 2>&1 | grep -q "PASS"; then
        pass "All formatters tests pass"
    else
        fail "Formatters tests failed"
    fi
    
    log_test "Native formatter test count"
    TEST_COUNT=$(grep -c "func Test" "$PROJECT_ROOT/internal/formatters/providers/native/native_test.go" 2>/dev/null || echo "0")
    if [ "$TEST_COUNT" -ge 15 ]; then
        pass "Native formatters have $TEST_COUNT tests (>= 15)"
    else
        fail "Native formatters have only $TEST_COUNT tests (expected >= 15)"
    fi
    
    log_test "Service formatter test count"
    TEST_COUNT=$(grep -c "func Test" "$PROJECT_ROOT/internal/formatters/providers/service/service_test.go" 2>/dev/null || echo "0")
    if [ "$TEST_COUNT" -ge 5 ]; then
        pass "Service formatters have $TEST_COUNT tests (>= 5)"
    else
        fail "Service formatters have only $TEST_COUNT tests (expected >= 5)"
    fi
}

# Optimization package
test_optimization() {
    section "Optimization Package Coverage"
    
    log_test "Optimization tests pass"
    if go test -short "$PROJECT_ROOT/internal/optimization/..." 2>&1 | grep -q "PASS"; then
        pass "Optimization tests pass"
    else
        fail "Optimization tests failed"
    fi
    
    log_test "Streaming optimization tests pass"
    if go test -short "$PROJECT_ROOT/internal/optimization/streaming/..." 2>&1 | grep -q "PASS"; then
        pass "Streaming optimization tests pass"
    else
        fail "Streaming optimization tests failed"
    fi
}

# Debate testing
test_debate_testing() {
    section "Debate Testing Coverage"
    
    log_test "Debate testing package compiles"
    if go build "$PROJECT_ROOT/internal/debate/..." 2>/dev/null; then
        pass "Debate testing package compiles"
    else
        fail "Debate testing package compilation failed"
    fi
    
    log_test "Debate tests pass"
    if go test -short "$PROJECT_ROOT/internal/debate/..." 2>&1 | grep -q "PASS"; then
        pass "Debate tests pass"
    else
        fail "Debate tests failed"
    fi
}

# Compilation check
test_compilation() {
    section "Compilation Check"
    
    log_test "All packages compile"
    if go build ./... 2>/dev/null; then
        pass "All packages compile successfully"
    else
        fail "Compilation failed"
    fi
    
    log_test "Services package compiles"
    if go build "$PROJECT_ROOT/internal/services/..." 2>/dev/null; then
        pass "Services package compiles"
    else
        fail "Services package compilation failed"
    fi
}

# Main
main() {
    echo ""
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║           Test Coverage Challenge - Phase 2                   ║"
    echo "║                   Coverage Validation                         ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    
    cd "$PROJECT_ROOT"
    
    test_testutils
    test_formatters
    test_optimization
    test_debate_testing
    test_compilation
    
    echo ""
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                      SUMMARY                                 ║"
    echo "╠══════════════════════════════════════════════════════════════╣"
    printf "║  Total:  %-3d  Passed: %-3d  Failed: %-3d                   ║\n" "$TESTS_TOTAL" "$TESTS_PASSED" "$TESTS_FAILED"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo ""
    
    if [ "$TESTS_FAILED" -eq 0 ]; then
        echo -e "${GREEN}✓ All tests passed!${NC}"
        exit 0
    else
        echo -e "${RED}✗ Some tests failed. Please review and fix.${NC}"
        exit 1
    fi
}

main "$@"
