#!/bin/bash
# Test Coverage Completeness Challenge
# Validates that every Go source file has a corresponding test file

set -euo pipefail

PASS=0
FAIL=0
TOTAL=0
MISSING_TESTS=()

check() {
    local desc="$1"
    local result="$2"
    TOTAL=$((TOTAL + 1))
    if [ "$result" = "0" ]; then
        echo "  PASS: $desc"
        PASS=$((PASS + 1))
    else
        echo "  FAIL: $desc"
        FAIL=$((FAIL + 1))
    fi
}

echo "=== Test Coverage Completeness Challenge ==="
echo ""

# Check each Go source file in internal/ has a corresponding test file
echo "--- Package Test File Coverage ---"
PACKAGES_CHECKED=0
PACKAGES_WITH_TESTS=0

for dir in $(find internal/ -type d ! -path '*/vendor/*' ! -path '*backup*'); do
    # Count source files (exclude doc.go, test files)
    src_count=$(find "$dir" -maxdepth 1 -name '*.go' ! -name '*_test.go' ! -name 'doc.go' 2>/dev/null | wc -l)
    if [ "$src_count" -eq 0 ]; then
        continue
    fi

    PACKAGES_CHECKED=$((PACKAGES_CHECKED + 1))

    # Count test files
    test_count=$(find "$dir" -maxdepth 1 -name '*_test.go' 2>/dev/null | wc -l)

    if [ "$test_count" -gt 0 ]; then
        PACKAGES_WITH_TESTS=$((PACKAGES_WITH_TESTS + 1))
    else
        MISSING_TESTS+=("$dir")
    fi
done

COVERAGE_PCT=$((PACKAGES_WITH_TESTS * 100 / PACKAGES_CHECKED))
check "Package test coverage >= 90% ($PACKAGES_WITH_TESTS/$PACKAGES_CHECKED = ${COVERAGE_PCT}%)" "$([ $COVERAGE_PCT -ge 90 ] && echo 0 || echo 1)"

# Report packages missing tests
if [ ${#MISSING_TESTS[@]} -gt 0 ]; then
    echo ""
    echo "  Packages missing test files:"
    for pkg in "${MISSING_TESTS[@]}"; do
        echo "    - $pkg"
    done
fi

# Verify test files compile
echo ""
echo "--- Test Compilation ---"
if go test ./internal/... -run='^$' -count=1 2>/dev/null; then
    check "All test files compile" "0"
else
    check "All test files compile" "1"
fi

# Verify fuzz test files exist
echo ""
echo "--- Fuzz Test Coverage ---"
FUZZ_COUNT=$(find tests/fuzz/ -name '*_test.go' 2>/dev/null | wc -l)
check "At least 5 fuzz test files (found: $FUZZ_COUNT)" "$([ $FUZZ_COUNT -ge 5 ] && echo 0 || echo 1)"

# Verify precondition tests exist
echo ""
echo "--- Precondition Tests ---"
PRECOND_COUNT=$(find tests/precondition/ -name '*_test.go' 2>/dev/null | wc -l)
check "At least 3 precondition test files (found: $PRECOND_COUNT)" "$([ $PRECOND_COUNT -ge 3 ] && echo 0 || echo 1)"

echo ""
echo "=== Results: $PASS/$TOTAL passed, $FAIL failed ==="
[ "$FAIL" -eq 0 ] && exit 0 || exit 1
