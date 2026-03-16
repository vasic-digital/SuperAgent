#!/usr/bin/env bash
set -euo pipefail

# Race Condition Challenge
# Validates that critical packages pass race condition detection tests.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

setup_challenge "race_condition" "$@"

PASSED=0
FAILED=0
TOTAL=0
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

print_header "Race Condition Challenge"
echo "Validates zero race conditions in critical packages using go test -race"
echo ""

# Critical packages to test for race conditions
CRITICAL_PACKAGES=(
    "internal/cache"
    "internal/background"
    "internal/adapters/background"
    "internal/models"
    "internal/utils"
    "internal/knowledge"
    "internal/planning"
    "internal/observability"
    "internal/llmops"
    "internal/selfimprove"
    "internal/optimization"
)

# =============================================================================
# Test each critical package for race conditions
# =============================================================================

for pkg in "${CRITICAL_PACKAGES[@]}"; do
    pkg_dir="$PROJECT_ROOT/$pkg"

    # Skip if directory doesn't exist
    if [ ! -d "$pkg_dir" ]; then
        continue
    fi

    # Skip if no Go test files
    if ! ls "$pkg_dir"/*_test.go > /dev/null 2>&1; then
        TOTAL=$((TOTAL + 1))
        test_start "Race detection: $pkg"
        test_pass
        echo "  (no test files - skipped)"
        PASSED=$((PASSED + 1))
        continue
    fi

    TOTAL=$((TOTAL + 1))
    test_start "Race detection: $pkg"

    race_output=$(GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -race -count=1 -short -timeout=5m \
        "$PROJECT_ROOT/$pkg/..." 2>&1 || true)

    if echo "$race_output" | grep -q "DATA RACE"; then
        test_fail "Race condition detected"
        FAILED=$((FAILED + 1))

        # Extract race details for debugging
        echo "    Race details:"
        echo "$race_output" | grep -A 3 "DATA RACE" | head -10 | sed 's/^/    /'
    elif echo "$race_output" | grep -q "FAIL"; then
        # Check if it's a test failure vs race failure
        if echo "$race_output" | grep -q "DATA RACE"; then
            test_fail "Race condition detected"
            FAILED=$((FAILED + 1))
        else
            test_fail "Test failure (not race-related)"
            FAILED=$((FAILED + 1))
        fi
    elif echo "$race_output" | grep -q "no test files"; then
        test_pass
        echo "  (no test files - skipped)"
        PASSED=$((PASSED + 1))
    else
        test_pass
        PASSED=$((PASSED + 1))
    fi
done

# =============================================================================
# Test: Verify -race flag compiles critical packages
# =============================================================================

TOTAL=$((TOTAL + 1))
test_start "Race-instrumented build compiles for models package"
if GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -race -c -o /dev/null \
    "$PROJECT_ROOT/internal/models/..." > /dev/null 2>&1; then
    test_pass
    PASSED=$((PASSED + 1))
else
    test_fail "Failed to compile with race detector"
    FAILED=$((FAILED + 1))
fi

# =============================================================================
# Test: Verify concurrent-safe types use sync primitives
# =============================================================================

TOTAL=$((TOTAL + 1))
test_start "Concurrent structures use sync primitives"
sync_issues=0

# Check that files with goroutines also have sync imports
for dir in cache background services; do
    pkg_dir="$PROJECT_ROOT/internal/$dir"
    [ -d "$pkg_dir" ] || continue

    for f in "$pkg_dir"/*.go; do
        [ -f "$f" ] || continue
        [[ "$f" == *_test.go ]] && continue

        # If file has goroutines or channels, it should use sync
        if grep -qE '(go func|make\(chan|<-chan|chan<-)' "$f" 2>/dev/null; then
            if ! grep -qE '(sync\.|context\.)' "$f" 2>/dev/null; then
                sync_issues=$((sync_issues + 1))
            fi
        fi
    done
done

if [ "$sync_issues" -eq 0 ]; then
    test_pass
    PASSED=$((PASSED + 1))
else
    test_fail "Found $sync_issues files with goroutines but no sync primitives"
    FAILED=$((FAILED + 1))
fi

# =============================================================================
# Summary
# =============================================================================

print_summary "Race Condition Challenge" "$PASSED" "$FAILED"

finalize_challenge "$FAILED"
exit "$FAILED"
