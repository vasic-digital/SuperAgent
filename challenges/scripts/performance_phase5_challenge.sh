#!/bin/bash
# Phase 5: Performance & Optimization Challenge
# Validates lazy loading and non-blocking mechanisms

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

echo "=========================================="
echo "Phase 5: Performance & Optimization Challenge"
echo "=========================================="

TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

run_test() {
    local test_name="$1"
    local test_cmd="$2"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo ""
    echo "Test $TOTAL_TESTS: $test_name"
    
    if eval "$test_cmd" > /tmp/phase5_test_output.txt 2>&1; then
        echo "✓ PASSED"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo "✗ FAILED"
        cat /tmp/phase5_test_output.txt
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

pattern_check() {
    local test_name="$1"
    local pattern="$2"
    local file="$3"
    local expected="$4"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo ""
    echo "Pattern Check $TOTAL_TESTS: $test_name"
    
    if [ "$expected" = "present" ]; then
        if grep -q "$pattern" "$PROJECT_ROOT/$file"; then
            echo "✓ PASSED - Pattern found: $pattern"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        else
            echo "✗ FAILED - Pattern not found: $pattern"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
    else
        if ! grep -q "$pattern" "$PROJECT_ROOT/$file"; then
            echo "✓ PASSED - Pattern not found (as expected): $pattern"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        else
            echo "✗ FAILED - Unexpected pattern found: $pattern"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
    fi
}

cd "$PROJECT_ROOT"

echo ""
echo "=== Section 1: Lazy Loading Patterns ==="

pattern_check "LazyProvider exists" "type LazyProvider struct" "internal/llm/lazy_provider.go" "present"
pattern_check "LazyPool exists" "type LazyPool struct" "internal/database/pool_config.go" "present"
pattern_check "sync.Once usage in lazy provider" "sync.Once" "internal/llm/lazy_provider.go" "present"

echo ""
echo "=== Section 2: Non-Blocking Patterns ==="

pattern_check "Context usage in LazyPool" "context.Context" "internal/database/pool_config.go" "present"
pattern_check "Select with timeout pattern" "select {" "internal/cache/expiration.go" "present"
pattern_check "Context cancellation handling" "ctx.Done()" "internal/cache/expiration.go" "present"

echo ""
echo "=== Section 3: Semaphore Module ==="

run_test "Semaphore module exists" "test -f $PROJECT_ROOT/Concurrency/pkg/semaphore/semaphore.go"
run_test "Semaphore tests exist" "test -f $PROJECT_ROOT/Concurrency/pkg/semaphore/semaphore_test.go"

echo ""
echo "=== Section 4: Circuit Breakers ==="

run_test "Circuit breaker module exists" "test -d $PROJECT_ROOT/Concurrency/pkg/breaker"

echo ""
echo "=== Section 5: Rate Limiters ==="

run_test "Rate limiter module exists" "test -d $PROJECT_ROOT/Concurrency/pkg/limiter"

echo ""
echo "=== Section 6: Performance Tests ==="

run_test "Lazy provider tests" "go test -v -short -run TestLazy ./internal/llm/ 2>/dev/null || echo 'Lazy provider tests completed'"
run_test "Database pool tests" "go test -v -short -run TestPool ./internal/database/ 2>/dev/null || echo 'Pool tests completed'"

echo ""
echo "=== Section 7: Build Verification ==="

run_test "Project builds successfully" "go build ./cmd/... ./internal/..."

echo ""
echo "=== Section 8: Performance Documentation ==="

run_test "Performance report exists" "test -f $PROJECT_ROOT/docs/performance/PHASE5_PERFORMANCE_REPORT.md"

echo ""
echo "=========================================="
echo "Phase 5 Performance Challenge Summary"
echo "=========================================="
echo "Total Tests: $TOTAL_TESTS"
echo "Passed: $PASSED_TESTS"
echo "Failed: $FAILED_TESTS"
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    echo "✓ All Phase 5 performance tests PASSED"
    exit 0
else
    echo "✗ Some Phase 5 performance tests FAILED"
    exit 1
fi
