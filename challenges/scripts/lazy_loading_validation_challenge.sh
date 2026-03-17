#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'

PASSED=0
FAILED=0
TOTAL=0

record_result() {
    TOTAL=$((TOTAL + 1))
    if [ "$2" = "PASS" ]; then
        PASSED=$((PASSED + 1))
        echo -e "${GREEN}[PASS]${NC} $1"
    else
        FAILED=$((FAILED + 1))
        echo -e "${RED}[FAIL]${NC} $1"
    fi
}

echo "=========================================="
echo "  Lazy Loading Validation Challenge"
echo "=========================================="
echo ""

# --------------------------------------------------------------------------
# Test 1: LazyProvider implementation exists
# --------------------------------------------------------------------------
if [ -f "$PROJECT_ROOT/internal/llm/lazy_provider.go" ]; then
    record_result "LazyProvider implementation exists" "PASS"
else
    record_result "LazyProvider implementation exists" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 2: LazyProvider has tests
# --------------------------------------------------------------------------
if [ -f "$PROJECT_ROOT/internal/llm/lazy_provider_test.go" ]; then
    record_result "LazyProvider tests exist" "PASS"
else
    record_result "LazyProvider tests exist" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 3: sync.Once usage in codebase (>= 10 instances in production code)
# --------------------------------------------------------------------------
SYNC_ONCE_COUNT=$(grep -r "sync\.Once" "$PROJECT_ROOT/internal/" --include="*.go" \
    | grep -v "_test\.go" | grep -v vendor | wc -l)
if [ "$SYNC_ONCE_COUNT" -ge 10 ]; then
    record_result "sync.Once usage >= 10 instances (found: $SYNC_ONCE_COUNT)" "PASS"
else
    record_result "sync.Once usage >= 10 instances (found: $SYNC_ONCE_COUNT)" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 4: Weighted semaphore usage in ensemble
# --------------------------------------------------------------------------
if grep -q "semaphore\.NewWeighted" "$PROJECT_ROOT/internal/llm/ensemble.go" 2>/dev/null; then
    record_result "Weighted semaphore in ensemble.go" "PASS"
else
    record_result "Weighted semaphore in ensemble.go" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 5: Semaphore import in ensemble
# --------------------------------------------------------------------------
if grep -q "golang.org/x/sync/semaphore" "$PROJECT_ROOT/internal/llm/ensemble.go" 2>/dev/null; then
    record_result "Semaphore import in ensemble.go" "PASS"
else
    record_result "Semaphore import in ensemble.go" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 6: Performance benchmark files exist (>= 3 benchmark files)
# --------------------------------------------------------------------------
BENCH_COUNT=$(find "$PROJECT_ROOT/tests/performance/" -name "*benchmark*" -o -name "*bench*" 2>/dev/null | wc -l)
if [ "$BENCH_COUNT" -ge 3 ]; then
    record_result "Performance benchmarks >= 3 files (found: $BENCH_COUNT)" "PASS"
else
    record_result "Performance benchmarks >= 3 files (found: $BENCH_COUNT)" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 7: Lazy loading benchmark file
# --------------------------------------------------------------------------
if [ -f "$PROJECT_ROOT/tests/performance/lazy_loading_benchmark_test.go" ]; then
    record_result "Lazy loading benchmark file exists" "PASS"
else
    record_result "Lazy loading benchmark file exists" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 8: Semaphore benchmark file
# --------------------------------------------------------------------------
if [ -f "$PROJECT_ROOT/tests/performance/semaphore_benchmark_test.go" ]; then
    record_result "Semaphore benchmark file exists" "PASS"
else
    record_result "Semaphore benchmark file exists" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 9: Ensemble benchmark file
# --------------------------------------------------------------------------
if [ -f "$PROJECT_ROOT/tests/performance/ensemble_benchmark_test.go" ]; then
    record_result "Ensemble benchmark file exists" "PASS"
else
    record_result "Ensemble benchmark file exists" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 10: No heavy init() functions in internal/ (audit)
# Only lightweight or registry inits should remain
# --------------------------------------------------------------------------
INIT_COUNT=$(grep -rn "func init()" "$PROJECT_ROOT/internal/" --include="*.go" \
    | grep -v "_test\.go" | grep -v vendor | wc -l)
if [ "$INIT_COUNT" -le 5 ]; then
    record_result "init() count <= 5 in internal/ production code (found: $INIT_COUNT)" "PASS"
else
    record_result "init() count <= 5 in internal/ production code (found: $INIT_COUNT)" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 11: LazyProviderRegistry exists
# --------------------------------------------------------------------------
if grep -q "LazyProviderRegistry" "$PROJECT_ROOT/internal/llm/lazy_provider.go" 2>/dev/null; then
    record_result "LazyProviderRegistry exists in lazy_provider.go" "PASS"
else
    record_result "LazyProviderRegistry exists in lazy_provider.go" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 12: Ensemble context cancellation test exists
# --------------------------------------------------------------------------
if grep -q "ContextCancellation" "$PROJECT_ROOT/internal/llm/ensemble_test.go" 2>/dev/null; then
    record_result "Ensemble context cancellation test exists" "PASS"
else
    record_result "Ensemble context cancellation test exists" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 13: testutil/infra.go has no init() function (converted to lazy)
# --------------------------------------------------------------------------
if grep -q "^func init()" "$PROJECT_ROOT/internal/testutil/infra.go" 2>/dev/null; then
    record_result "testutil/infra.go init() removed (lazy init)" "FAIL"
else
    record_result "testutil/infra.go init() removed (lazy init)" "PASS"
fi

# --------------------------------------------------------------------------
# Test 14: testutil/infra.go uses getInfraResults() accessor
# --------------------------------------------------------------------------
if grep -q "getInfraResults" "$PROJECT_ROOT/internal/testutil/infra.go" 2>/dev/null; then
    record_result "testutil/infra.go uses lazy getInfraResults()" "PASS"
else
    record_result "testutil/infra.go uses lazy getInfraResults()" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 15: Debate performance optimizer has semaphore
# --------------------------------------------------------------------------
if grep -q "semaphore" "$PROJECT_ROOT/internal/services/debate_performance_optimizer.go" 2>/dev/null; then
    record_result "Debate performance optimizer has semaphore" "PASS"
else
    record_result "Debate performance optimizer has semaphore" "FAIL"
fi

echo ""
echo "=========================================="
echo "  Results: $PASSED/$TOTAL passed, $FAILED failed"
echo "=========================================="

[ $FAILED -eq 0 ] && exit 0 || exit 1
