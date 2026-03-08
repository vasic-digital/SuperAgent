#!/usr/bin/env bash
# toolkit_unit_challenge.sh - Validates Toolkit module unit tests
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MODULE_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"
MODULE_NAME="Toolkit"

PASS=0
FAIL=0
TOTAL=0

pass() { PASS=$((PASS+1)); TOTAL=$((TOTAL+1)); echo "  PASS: $1"; }
fail() { FAIL=$((FAIL+1)); TOTAL=$((TOTAL+1)); echo "  FAIL: $1"; }

echo "=== ${MODULE_NAME} Unit Test Challenge ==="
echo ""

# Test 1: Test files exist
echo "Test: Test files exist"
test_count=$(find "${MODULE_DIR}" -name "*_test.go" | wc -l)
if [ "${test_count}" -gt 0 ]; then
    pass "Found ${test_count} test files"
else
    fail "No test files found"
fi

# Test 2: Tests exist across key directories
echo "Test: Test coverage across directories"
dirs_with_tests=0
for sub_dir in pkg/toolkit Commons Providers tests; do
    if [ -d "${MODULE_DIR}/${sub_dir}" ]; then
        sub_tests=$(find "${MODULE_DIR}/${sub_dir}" -name "*_test.go" | wc -l)
        if [ "$sub_tests" -gt 0 ]; then
            dirs_with_tests=$((dirs_with_tests + 1))
        fi
    fi
done
if [ "$dirs_with_tests" -ge 2 ]; then
    pass "At least 2 directories have tests (found ${dirs_with_tests})"
else
    fail "Only ${dirs_with_tests} directories have tests (expected at least 2)"
fi

# Test 3: Unit tests pass
echo "Test: Unit tests pass"
if (cd "${MODULE_DIR}" && GOMAXPROCS=2 nice -n 19 go test -short -count=1 -p 1 ./... 2>&1); then
    pass "Unit tests pass"
else
    fail "Unit tests failed"
fi

# Test 4: No race conditions (short mode)
echo "Test: Race detector clean"
if (cd "${MODULE_DIR}" && GOMAXPROCS=2 nice -n 19 go test -short -race -count=1 -p 1 ./... 2>&1); then
    pass "No race conditions detected"
else
    fail "Race conditions detected"
fi

echo ""
echo "=== Results: ${PASS}/${TOTAL} passed, ${FAIL} failed ==="
[ "${FAIL}" -eq 0 ] && exit 0 || exit 1
