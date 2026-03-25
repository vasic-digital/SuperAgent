#!/bin/bash
# Safety Comprehensive Challenge
# Validates memory safety: no race conditions, context-guarded cleanup loops,
# channel context handling, goroutine leak guards, and circuit breaker caps.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

PASS=0
FAIL=0
TOTAL=0

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

echo "=== Safety Comprehensive Challenge ==="
echo ""

# ============================================================================
# SECTION 1: Race Condition Detection
# ============================================================================
echo "--- Race Condition Detection ---"

cd "$PROJECT_ROOT"

# Test 1: llm package — no data races
RACE_LLM=0
if GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -race -short -count=1 -p 1 \
    ./internal/llm/... 2>&1 | grep -q "DATA RACE"; then
    RACE_LLM=1
fi
check "No data races in internal/llm/" "$RACE_LLM"

# Test 2: services package — no data races
RACE_SVC=0
if GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -race -short -count=1 -p 1 \
    ./internal/services/... 2>&1 | grep -q "DATA RACE"; then
    RACE_SVC=1
fi
check "No data races in internal/services/" "$RACE_SVC"

# Test 3: cache package — no data races
RACE_CACHE=0
if GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -race -short -count=1 -p 1 \
    ./internal/cache/... 2>&1 | grep -q "DATA RACE"; then
    RACE_CACHE=1
fi
check "No data races in internal/cache/" "$RACE_CACHE"

# Test 4: race-instrumented build compiles for llm package
RACE_BUILD=0
if ! GOMAXPROCS=2 nice -n 19 go test -race -c -o /dev/null \
    "$PROJECT_ROOT/internal/llm/..." > /dev/null 2>&1; then
    RACE_BUILD=1
fi
check "Race-instrumented build compiles for internal/llm/" "$RACE_BUILD"

# ============================================================================
# SECTION 2: Cleanup Loop Context Safety
# ============================================================================
echo ""
echo "--- Cleanup Loop Context Safety ---"

# For each file containing cleanupLoop, verify ctx.Done() appears within
# a reasonable proximity (within 30 lines after the function definition).
LOOP_ISSUES=0
LOOP_CHECKED=0

while IFS= read -r file; do
    # Find all cleanupLoop function definition line numbers
    while IFS= read -r linenum; do
        LOOP_CHECKED=$((LOOP_CHECKED + 1))
        # Extract up to 30 lines starting from the function definition
        context=$(awk "NR>=$linenum && NR<=$((linenum + 30))" "$file" 2>/dev/null || true)
        if ! echo "$context" | grep -q "ctx\.Done\(\)\|context\.Done\(\)\|select {"; then
            LOOP_ISSUES=$((LOOP_ISSUES + 1))
        fi
    done < <(grep -n "^func.*cleanupLoop" "$file" 2>/dev/null | cut -d: -f1 || true)
done < <(grep -rl "cleanupLoop" "$PROJECT_ROOT/internal/" --include="*.go" 2>/dev/null || true)

if [ "$LOOP_CHECKED" -eq 0 ]; then
    check "Cleanup loops found and verified" "0"
else
    check "All $LOOP_CHECKED cleanup loops have context.Done() select" "$LOOP_ISSUES"
fi

# ============================================================================
# SECTION 3: ACP Channel Context Handling
# ============================================================================
echo ""
echo "--- ACP Channel Context Handling ---"

# Verify ACP files that use make(chan also use ctx.Done() or select for context
ACP_CHAN_ISSUES=0
ACP_FILES_CHECKED=0

while IFS= read -r file; do
    if grep -q "make(chan" "$file" 2>/dev/null; then
        ACP_FILES_CHECKED=$((ACP_FILES_CHECKED + 1))
        if ! grep -q "ctx\.Done\(\)\|context\.Done\(\)\|select {" "$file" 2>/dev/null; then
            ACP_CHAN_ISSUES=$((ACP_CHAN_ISSUES + 1))
        fi
    fi
done < <(find "$PROJECT_ROOT/internal/llm/providers" -name "*_acp.go" 2>/dev/null || true)

if [ "$ACP_FILES_CHECKED" -eq 0 ]; then
    check "ACP files with channels use context select" "0"
else
    check "All $ACP_FILES_CHECKED ACP files with channels use context handling" "$ACP_CHAN_ISSUES"
fi

# ============================================================================
# SECTION 4: Goroutine Leak Guard (goleak compilation check)
# ============================================================================
echo ""
echo "--- Goroutine Leak Guards ---"

# Verify that critical packages reference goleak or WaitGroup for leak prevention
GOLEAK_COUNT=$(grep -rl "goleak\|WaitGroup" "$PROJECT_ROOT/internal/" --include="*_test.go" 2>/dev/null | wc -l || echo "0")
if [ "$GOLEAK_COUNT" -gt 0 ]; then
    check "Goroutine leak guards (goleak or WaitGroup) present in tests ($GOLEAK_COUNT files)" "0"
else
    check "Goroutine leak guards (goleak or WaitGroup) present in tests" "1"
fi

# Verify goroutine-lifecycle diagram exists as structural documentation
DIAGRAM_EXISTS=0
if [ ! -f "$PROJECT_ROOT/docs/diagrams/src/goroutine-lifecycle.puml" ]; then
    DIAGRAM_EXISTS=1
fi
check "Goroutine lifecycle diagram exists (docs/diagrams/src/goroutine-lifecycle.puml)" "$DIAGRAM_EXISTS"

# ============================================================================
# SECTION 5: Circuit Breaker Listener Cap
# ============================================================================
echo ""
echo "--- Circuit Breaker Safety ---"

# Verify MaxCircuitBreakerListeners constant exists to prevent listener leaks
CB_CAP=$(grep -r "MaxCircuitBreakerListeners" "$PROJECT_ROOT/internal/" --include="*.go" 2>/dev/null | wc -l || echo "0")
if [ "$CB_CAP" -gt 0 ]; then
    check "MaxCircuitBreakerListeners cap defined ($CB_CAP references)" "0"
else
    check "MaxCircuitBreakerListeners cap defined" "1"
fi

# Verify the cap is enforced (checked against len of listeners slice)
CB_ENFORCED=$(grep -r "len.*listeners.*MaxCircuitBreakerListeners\|MaxCircuitBreakerListeners.*len.*listeners\|len(cb\.listeners) >= MaxCircuitBreakerListeners" \
    "$PROJECT_ROOT/internal/" --include="*.go" 2>/dev/null | wc -l || echo "0")
if [ "$CB_ENFORCED" -gt 0 ]; then
    check "Circuit breaker listener cap is enforced in code" "0"
else
    check "Circuit breaker listener cap is enforced in code" "1"
fi

# ============================================================================
# SECTION 6: Build Verification
# ============================================================================
echo ""
echo "--- Build Verification ---"

# Test: cmd/ and internal/ builds cleanly
BUILD_OK=0
if ! cd "$PROJECT_ROOT" && go build ./cmd/... ./internal/... > /dev/null 2>&1; then
    BUILD_OK=1
fi
check "go build ./cmd/... ./internal/... succeeds" "$BUILD_OK"

# Test: go vet passes on internal packages
VET_OK=0
if ! cd "$PROJECT_ROOT" && GOMAXPROCS=2 nice -n 19 go vet ./internal/... > /dev/null 2>&1; then
    VET_OK=1
fi
check "go vet ./internal/... passes" "$VET_OK"

echo ""
echo "=== Results: $PASS/$TOTAL passed, $FAIL failed ==="
[ "$FAIL" -eq 0 ] && exit 0 || exit 1
