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
echo "  Lazy Loading Comprehensive Challenge"
echo "=========================================="
echo ""

# --------------------------------------------------------------------------
# Test 1: sync.Once used in LLM provider files (>= 5 occurrences)
# --------------------------------------------------------------------------
SYNC_ONCE_LLM=$(grep -r "sync\.Once" "$PROJECT_ROOT/internal/llm/" --include="*.go" \
    | grep -v "_test\.go" | wc -l)
if [ "$SYNC_ONCE_LLM" -ge 5 ]; then
    record_result "sync.Once usage in internal/llm/ >= 5 (found: $SYNC_ONCE_LLM)" "PASS"
else
    record_result "sync.Once usage in internal/llm/ >= 5 (found: $SYNC_ONCE_LLM)" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 2: LazyProvider implementation exists
# --------------------------------------------------------------------------
if [ -f "$PROJECT_ROOT/internal/llm/lazy_provider.go" ]; then
    record_result "LazyProvider implementation exists (internal/llm/lazy_provider.go)" "PASS"
else
    record_result "LazyProvider implementation exists (internal/llm/lazy_provider.go)" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 3: LazyProvider has tests
# --------------------------------------------------------------------------
if [ -f "$PROJECT_ROOT/internal/llm/lazy_provider_test.go" ]; then
    record_result "LazyProvider tests exist (lazy_provider_test.go)" "PASS"
else
    record_result "LazyProvider tests exist (lazy_provider_test.go)" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 4: LazyProviderRegistry type defined in lazy_provider.go
# --------------------------------------------------------------------------
if grep -q "LazyProviderRegistry" "$PROJECT_ROOT/internal/llm/lazy_provider.go" 2>/dev/null; then
    record_result "LazyProviderRegistry defined in lazy_provider.go" "PASS"
else
    record_result "LazyProviderRegistry defined in lazy_provider.go" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 5: At least 10 files use sync.Once across the entire codebase
# --------------------------------------------------------------------------
SYNC_ONCE_TOTAL=$(grep -rl "sync\.Once" "$PROJECT_ROOT/internal/" --include="*.go" \
    | grep -v "_test\.go" | wc -l | tr -d '[:space:]')
if [ "$SYNC_ONCE_TOTAL" -ge 10 ]; then
    record_result "At least 10 files use sync.Once (found: $SYNC_ONCE_TOTAL)" "PASS"
else
    record_result "At least 10 files use sync.Once (found: $SYNC_ONCE_TOTAL)" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 6: sync.Once used in LLM provider CLI implementations
# --------------------------------------------------------------------------
CLI_ONCE_COUNT=$(grep -r "sync\.Once" "$PROJECT_ROOT/internal/llm/providers/" --include="*.go" \
    | grep -v "_test\.go" | wc -l)
if [ "$CLI_ONCE_COUNT" -ge 5 ]; then
    record_result "sync.Once usage in LLM providers >= 5 (found: $CLI_ONCE_COUNT)" "PASS"
else
    record_result "sync.Once usage in LLM providers >= 5 (found: $CLI_ONCE_COUNT)" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 7: No heavy init() functions — init() count <= 5 in internal/ production code
# --------------------------------------------------------------------------
INIT_COUNT=$(grep -rn "^func init()" "$PROJECT_ROOT/internal/" --include="*.go" \
    | grep -v "_test\.go" | wc -l)
if [ "$INIT_COUNT" -le 5 ]; then
    record_result "init() count <= 5 in internal/ production code (found: $INIT_COUNT)" "PASS"
else
    record_result "init() count <= 5 in internal/ production code (found: $INIT_COUNT)" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 8: metricsOnce pattern used in services (lazy metric registration)
# --------------------------------------------------------------------------
METRICS_ONCE=$(grep -r "metricsOnce\|MetricsOnce" "$PROJECT_ROOT/internal/services/" \
    --include="*.go" | grep -v "_test\.go" | wc -l)
if [ "$METRICS_ONCE" -ge 3 ]; then
    record_result "metricsOnce lazy registration in services >= 3 (found: $METRICS_ONCE)" "PASS"
else
    record_result "metricsOnce lazy registration in services >= 3 (found: $METRICS_ONCE)" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 9: Weighted semaphore used for concurrency limiting
# --------------------------------------------------------------------------
if grep -q "semaphore\.NewWeighted\|golang\.org/x/sync/semaphore" \
    "$PROJECT_ROOT/internal/llm/ensemble.go" 2>/dev/null; then
    record_result "Weighted semaphore used in ensemble.go" "PASS"
else
    record_result "Weighted semaphore used in ensemble.go" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 10: Lazy loading benchmark test exists
# --------------------------------------------------------------------------
if [ -f "$PROJECT_ROOT/tests/performance/lazy_loading_benchmark_test.go" ]; then
    record_result "Lazy loading benchmark test file exists" "PASS"
else
    record_result "Lazy loading benchmark test file exists" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 11: Semaphore benchmark test exists
# --------------------------------------------------------------------------
if [ -f "$PROJECT_ROOT/tests/performance/semaphore_benchmark_test.go" ]; then
    record_result "Semaphore benchmark test file exists" "PASS"
else
    record_result "Semaphore benchmark test file exists" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 12: Build passes cleanly (main helixagent binary)
# --------------------------------------------------------------------------
cd "$PROJECT_ROOT"
if GOMAXPROCS=2 go build ./cmd/helixagent/ 2>/dev/null; then
    record_result "Main helixagent binary builds cleanly" "PASS"
else
    record_result "Main helixagent binary builds cleanly" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 13: provider_registry.go uses sync.Once per-provider lazy init
# --------------------------------------------------------------------------
if grep -q "sync\.Once\|initOnce" "$PROJECT_ROOT/internal/services/provider_registry.go" 2>/dev/null; then
    record_result "provider_registry.go uses per-provider sync.Once lazy init" "PASS"
else
    record_result "provider_registry.go uses per-provider sync.Once lazy init" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 14: Debate performance optimizer uses semaphore (non-blocking)
# --------------------------------------------------------------------------
if grep -q "semaphore" "$PROJECT_ROOT/internal/services/debate_performance_optimizer.go" 2>/dev/null; then
    record_result "Debate performance optimizer uses semaphore" "PASS"
else
    record_result "Debate performance optimizer uses semaphore" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 15: cliCheckOnce or modelsDiscoveryOnce used in at least one provider
# --------------------------------------------------------------------------
CLI_CHECK_ONCE=$(grep -r "cliCheckOnce\|modelsDiscoveryOnce\|startOnce\|initOnce" \
    "$PROJECT_ROOT/internal/llm/providers/" --include="*.go" | grep -v "_test\.go" | wc -l)
if [ "$CLI_CHECK_ONCE" -ge 3 ]; then
    record_result "Provider-specific Once fields used >= 3 (found: $CLI_CHECK_ONCE)" "PASS"
else
    record_result "Provider-specific Once fields used >= 3 (found: $CLI_CHECK_ONCE)" "FAIL"
fi

echo ""
echo "=========================================="
echo "  Results: $PASSED/$TOTAL passed, $FAILED failed"
echo "=========================================="

[ $FAILED -eq 0 ] && exit 0 || exit 1
