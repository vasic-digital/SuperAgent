#!/usr/bin/env bash
set -euo pipefail

# Resource Limits Challenge
# Validates that all Makefile test targets enforce resource limits
# per Constitution Rule 15 (30-40% host resources).

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

setup_challenge "resource_limits" "$@"

PASSED=0
FAILED=0
TOTAL=0
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
MAKEFILE="$PROJECT_ROOT/Makefile"

print_header "Resource Limits Challenge"
echo "Validates all Makefile test targets enforce resource limits"
echo ""

# =============================================================================
# Test 1: RESOURCE_PREFIX variable is defined
# =============================================================================

test_start "RESOURCE_PREFIX variable defined in Makefile"
TOTAL=$((TOTAL + 1))
if grep -q '^RESOURCE_PREFIX' "$MAKEFILE"; then
    test_pass
    PASSED=$((PASSED + 1))
else
    test_fail "RESOURCE_PREFIX not defined in Makefile"
    FAILED=$((FAILED + 1))
fi

# =============================================================================
# Test 2: GO_TEST_FLAGS variable is defined
# =============================================================================

test_start "GO_TEST_FLAGS variable defined in Makefile"
TOTAL=$((TOTAL + 1))
if grep -q '^GO_TEST_FLAGS' "$MAKEFILE"; then
    test_pass
    PASSED=$((PASSED + 1))
else
    test_fail "GO_TEST_FLAGS not defined in Makefile"
    FAILED=$((FAILED + 1))
fi

# =============================================================================
# Test 3: GOMAXPROCS is exported
# =============================================================================

test_start "GOMAXPROCS exported in Makefile"
TOTAL=$((TOTAL + 1))
if grep -q '^export GOMAXPROCS' "$MAKEFILE"; then
    test_pass
    PASSED=$((PASSED + 1))
else
    test_fail "GOMAXPROCS not exported in Makefile"
    FAILED=$((FAILED + 1))
fi

# =============================================================================
# Test 4: RESOURCE_PREFIX includes nice
# =============================================================================

test_start "RESOURCE_PREFIX includes nice -n 19"
TOTAL=$((TOTAL + 1))
if grep '^RESOURCE_PREFIX' "$MAKEFILE" | grep -q 'nice -n 19'; then
    test_pass
    PASSED=$((PASSED + 1))
else
    test_fail "RESOURCE_PREFIX missing nice -n 19"
    FAILED=$((FAILED + 1))
fi

# =============================================================================
# Test 5: RESOURCE_PREFIX includes ionice
# =============================================================================

test_start "RESOURCE_PREFIX includes ionice -c 3"
TOTAL=$((TOTAL + 1))
if grep '^RESOURCE_PREFIX' "$MAKEFILE" | grep -q 'ionice -c 3'; then
    test_pass
    PASSED=$((PASSED + 1))
else
    test_fail "RESOURCE_PREFIX missing ionice -c 3"
    FAILED=$((FAILED + 1))
fi

# =============================================================================
# Test 6: GO_TEST_FLAGS includes -p 1
# =============================================================================

test_start "GO_TEST_FLAGS includes -p 1"
TOTAL=$((TOTAL + 1))
if grep '^GO_TEST_FLAGS' "$MAKEFILE" | grep -q '\-p 1'; then
    test_pass
    PASSED=$((PASSED + 1))
else
    test_fail "GO_TEST_FLAGS missing -p 1"
    FAILED=$((FAILED + 1))
fi

# =============================================================================
# Test 7-N: Each go test invocation uses RESOURCE_PREFIX
# =============================================================================

# Find all lines with 'go test' that are NOT inside docker exec commands
# and check they use RESOURCE_PREFIX
TARGETS_WITH_GO_TEST=()
while IFS= read -r line; do
    # Skip docker exec lines
    if echo "$line" | grep -q 'docker.*exec'; then
        continue
    fi
    # Skip comment lines
    if echo "$line" | grep -q '^\s*#'; then
        continue
    fi
    TARGETS_WITH_GO_TEST+=("$line")
done < <(grep -n 'go test' "$MAKEFILE")

for entry in "${TARGETS_WITH_GO_TEST[@]}"; do
    line_num=$(echo "$entry" | cut -d: -f1)
    line_content=$(echo "$entry" | cut -d: -f2-)
    # Trim leading whitespace for display
    display=$(echo "$line_content" | sed 's/^[[:space:]]*//' | head -c 80)

    test_start "Resource prefix on line $line_num: $display"
    TOTAL=$((TOTAL + 1))

    if echo "$line_content" | grep -q 'RESOURCE_PREFIX'; then
        test_pass
        PASSED=$((PASSED + 1))
    else
        test_fail "Line $line_num missing RESOURCE_PREFIX: $display"
        FAILED=$((FAILED + 1))
    fi
done

# =============================================================================
# Test: make -n test-unit shows resource limits
# =============================================================================

test_start "make -n test-unit includes nice/ionice"
TOTAL=$((TOTAL + 1))
MAKE_OUTPUT=$(make -n test-unit -C "$PROJECT_ROOT" 2>&1 || true)
if echo "$MAKE_OUTPUT" | grep -q 'nice'; then
    test_pass
    PASSED=$((PASSED + 1))
else
    test_fail "make -n test-unit does not show nice in output"
    FAILED=$((FAILED + 1))
fi

# =============================================================================
# Summary
# =============================================================================

echo ""
echo "=================================================="
echo "Resource Limits Challenge Results"
echo "=================================================="
echo -e "Total:  $TOTAL"
echo -e "Passed: ${GREEN}$PASSED${NC}"
echo -e "Failed: ${RED}$FAILED${NC}"
echo ""

if [ "$FAILED" -gt 0 ]; then
    echo -e "${RED}CHALLENGE FAILED${NC}"
    exit 1
else
    echo -e "${GREEN}CHALLENGE PASSED${NC}"
    exit 0
fi
