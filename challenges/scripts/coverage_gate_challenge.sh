#!/usr/bin/env bash
set -euo pipefail

# Coverage Gate Challenge
# Validates that code coverage tooling is operational, coverage targets exist
# in the Makefile, and coverage profiles can be generated and parsed.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

setup_challenge "coverage_gate" "$@"

PASSED=0
FAILED=0
TOTAL=0
MIN_COVERAGE=70
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

record_result() {
    TOTAL=$((TOTAL + 1))
    test_start "$1"
    if [ "$2" = "PASS" ]; then
        PASSED=$((PASSED + 1))
        test_pass
    else
        FAILED=$((FAILED + 1))
        test_fail "$1"
    fi
}

print_header "Coverage Gate Challenge"
echo "Minimum threshold: ${MIN_COVERAGE}%"
echo ""

# Test 1: Go coverage tooling available
# go tool cover -help exits with code 2, so check output for usage text
COVER_OUTPUT=$(go tool cover -help 2>&1 || true)
if echo "$COVER_OUTPUT" | grep -q "coverage profile"; then
    record_result "Go coverage tooling available" "PASS"
else
    record_result "Go coverage tooling available" "FAIL"
fi

# Test 2: Makefile has test-coverage target
if grep -q "test-coverage" "$PROJECT_ROOT/Makefile"; then
    record_result "Makefile has test-coverage target" "PASS"
else
    record_result "Makefile has test-coverage target" "FAIL"
fi

# Test 3: Makefile has test-coverage-100 target
if grep -q "test-coverage-100" "$PROJECT_ROOT/Makefile"; then
    record_result "Makefile has test-coverage-100 target" "PASS"
else
    record_result "Makefile has test-coverage-100 target" "FAIL"
fi

# Test 4: Coverage profile can be generated for internal/models
cd "$PROJECT_ROOT"
COVERAGE_FILE="/tmp/helixagent_coverage_gate_$$.out"
if GOMAXPROCS=2 nice -n 19 go test -short -count=1 -coverprofile="$COVERAGE_FILE" ./internal/models/... 2>/dev/null; then
    record_result "Coverage profile generated for internal/models" "PASS"

    # Test 5: Parse coverage percentage
    COVERAGE=$(go tool cover -func="$COVERAGE_FILE" 2>/dev/null | tail -1 | awk '{print $NF}' | tr -d '%')
    if [ -n "$COVERAGE" ]; then
        COVERAGE_INT=${COVERAGE%.*}
        if [ "$COVERAGE_INT" -ge "$MIN_COVERAGE" ]; then
            record_result "internal/models coverage >= ${MIN_COVERAGE}% (actual: ${COVERAGE}%)" "PASS"
        else
            record_result "internal/models coverage >= ${MIN_COVERAGE}% (actual: ${COVERAGE}%)" "FAIL"
        fi
    else
        record_result "internal/models coverage parsing" "FAIL"
    fi

    # Test 6: HTML report generation works
    HTML_FILE="/tmp/helixagent_coverage_gate_$$.html"
    if go tool cover -html="$COVERAGE_FILE" -o "$HTML_FILE" 2>/dev/null; then
        record_result "HTML coverage report generation" "PASS"
        rm -f "$HTML_FILE"
    else
        record_result "HTML coverage report generation" "FAIL"
    fi
else
    record_result "Coverage profile generation for internal/models" "FAIL"
    # Skip dependent tests
    record_result "internal/models coverage parsing (skipped - no profile)" "FAIL"
    record_result "HTML coverage report generation (skipped - no profile)" "FAIL"
fi
rm -f "$COVERAGE_FILE"

# Test 7: Fuzz tests directory exists
if [ -d "$PROJECT_ROOT/tests/fuzz" ]; then
    record_result "Fuzz tests directory exists" "PASS"
else
    record_result "Fuzz tests directory exists" "FAIL"
fi

# Test 8: Fuzz tests compile
if GOMAXPROCS=2 go test -run=XXX_NOMATCH ./tests/fuzz/ 2>/dev/null; then
    record_result "Fuzz tests compile successfully" "PASS"
else
    record_result "Fuzz tests compile successfully" "FAIL"
fi

# Test 9: Stress tests directory has test files
STRESS_COUNT=$(find "$PROJECT_ROOT/tests/stress" -name "*_test.go" -type f 2>/dev/null | wc -l)
if [ "$STRESS_COUNT" -gt 5 ]; then
    record_result "Stress tests directory has $STRESS_COUNT test files (>5)" "PASS"
else
    record_result "Stress tests directory has $STRESS_COUNT test files (need >5)" "FAIL"
fi

# Test 10: Chaos tests directory has test files
CHAOS_COUNT=$(find "$PROJECT_ROOT/tests/chaos" -name "*_test.go" -type f 2>/dev/null | wc -l)
if [ "$CHAOS_COUNT" -gt 3 ]; then
    record_result "Chaos tests directory has $CHAOS_COUNT test files (>3)" "PASS"
else
    record_result "Chaos tests directory has $CHAOS_COUNT test files (need >3)" "FAIL"
fi

# Test 11: Performance tests have pprof leak detection
if grep -rq "TestMemoryLeak" "$PROJECT_ROOT/tests/performance/" 2>/dev/null; then
    record_result "Performance tests include memory leak detection" "PASS"
else
    record_result "Performance tests include memory leak detection" "FAIL"
fi

# Test 12: Test resource limits enforced (GOMAXPROCS=2 pattern)
GOMAXPROCS_COUNT=$(grep -r "runtime.GOMAXPROCS(2)" "$PROJECT_ROOT/tests/" 2>/dev/null | wc -l)
if [ "$GOMAXPROCS_COUNT" -gt 5 ]; then
    record_result "GOMAXPROCS(2) resource limit pattern found ($GOMAXPROCS_COUNT occurrences)" "PASS"
else
    record_result "GOMAXPROCS(2) resource limit pattern found ($GOMAXPROCS_COUNT, need >5)" "FAIL"
fi

echo ""
print_summary "Coverage Gate Challenge" "$PASSED" "$FAILED"
[ "$FAILED" -eq 0 ] && exit 0 || exit 1
