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
echo "  Benchmark Regression Challenge"
echo "=========================================="
echo ""

# --------------------------------------------------------------------------
# Test 1: benchmark_test.go exists in tests/performance/
# --------------------------------------------------------------------------
if [ -f "$PROJECT_ROOT/tests/performance/benchmark_test.go" ]; then
    record_result "benchmark_test.go exists in tests/performance/" "PASS"
else
    record_result "benchmark_test.go exists in tests/performance/" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 2: ensemble_benchmark_test.go exists
# --------------------------------------------------------------------------
if [ -f "$PROJECT_ROOT/tests/performance/ensemble_benchmark_test.go" ]; then
    record_result "ensemble_benchmark_test.go exists" "PASS"
else
    record_result "ensemble_benchmark_test.go exists" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 3: lazy_loading_benchmark_test.go exists
# --------------------------------------------------------------------------
if [ -f "$PROJECT_ROOT/tests/performance/lazy_loading_benchmark_test.go" ]; then
    record_result "lazy_loading_benchmark_test.go exists" "PASS"
else
    record_result "lazy_loading_benchmark_test.go exists" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 4: semaphore_benchmark_test.go exists
# --------------------------------------------------------------------------
if [ -f "$PROJECT_ROOT/tests/performance/semaphore_benchmark_test.go" ]; then
    record_result "semaphore_benchmark_test.go exists" "PASS"
else
    record_result "semaphore_benchmark_test.go exists" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 5: At least 10 Benchmark functions defined across performance tests
# --------------------------------------------------------------------------
BENCH_FUNC_COUNT=$(grep -r "^func Benchmark" "$PROJECT_ROOT/tests/performance/" \
    --include="*.go" | wc -l)
if [ "$BENCH_FUNC_COUNT" -ge 10 ]; then
    record_result "At least 10 Benchmark functions defined (found: $BENCH_FUNC_COUNT)" "PASS"
else
    record_result "At least 10 Benchmark functions defined (found: $BENCH_FUNC_COUNT)" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 6: Benchmarks compile with -tags performance
# --------------------------------------------------------------------------
cd "$PROJECT_ROOT"
if GOMAXPROCS=2 go test -tags performance -run='^$' -count=1 ./tests/performance/ 2>/dev/null; then
    record_result "Performance benchmarks compile with -tags performance" "PASS"
else
    record_result "Performance benchmarks compile with -tags performance" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 7: Performance build tag declared in benchmark_test.go
# --------------------------------------------------------------------------
if grep -q "//go:build performance\|// +build performance" \
    "$PROJECT_ROOT/tests/performance/benchmark_test.go" 2>/dev/null; then
    record_result "performance build tag declared in benchmark_test.go" "PASS"
else
    record_result "performance build tag declared in benchmark_test.go" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 8: docs/performance/ directory exists
# --------------------------------------------------------------------------
if [ -d "$PROJECT_ROOT/docs/performance" ]; then
    record_result "docs/performance/ directory exists" "PASS"
else
    record_result "docs/performance/ directory exists" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 9: docs/performance/BENCHMARKS.md exists
# --------------------------------------------------------------------------
if [ -f "$PROJECT_ROOT/docs/performance/BENCHMARKS.md" ]; then
    record_result "docs/performance/BENCHMARKS.md exists" "PASS"
else
    record_result "docs/performance/BENCHMARKS.md exists" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 10: Resource limits enforced — GOMAXPROCS=2 used in performance tests
# --------------------------------------------------------------------------
GOMAXPROCS_COUNT=$(grep -r "GOMAXPROCS\|runtime\.GOMAXPROCS" \
    "$PROJECT_ROOT/tests/performance/" --include="*.go" | wc -l)
if [ "$GOMAXPROCS_COUNT" -ge 1 ]; then
    record_result "GOMAXPROCS resource limit enforced in performance tests (found: $GOMAXPROCS_COUNT refs)" "PASS"
else
    record_result "GOMAXPROCS resource limit enforced in performance tests (found: $GOMAXPROCS_COUNT refs)" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 11: debate_benchmark_test.go exists
# --------------------------------------------------------------------------
if [ -f "$PROJECT_ROOT/tests/performance/debate_benchmark_test.go" ]; then
    record_result "debate_benchmark_test.go exists" "PASS"
else
    record_result "debate_benchmark_test.go exists" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 12: Benchmark functions cover cache operations (BenchmarkCache_*)
# --------------------------------------------------------------------------
if grep -q "func BenchmarkCache_" "$PROJECT_ROOT/tests/performance/benchmark_test.go" 2>/dev/null; then
    record_result "Cache benchmark functions defined (BenchmarkCache_*)" "PASS"
else
    record_result "Cache benchmark functions defined (BenchmarkCache_*)" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 13: Benchmark functions cover ensemble operations
# --------------------------------------------------------------------------
if grep -q "func BenchmarkEnsemble_" "$PROJECT_ROOT/tests/performance/ensemble_benchmark_test.go" 2>/dev/null; then
    record_result "Ensemble benchmark functions defined (BenchmarkEnsemble_*)" "PASS"
else
    record_result "Ensemble benchmark functions defined (BenchmarkEnsemble_*)" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 14: messaging benchmark tests exist
# --------------------------------------------------------------------------
if [ -f "$PROJECT_ROOT/tests/performance/messaging/benchmark_test.go" ]; then
    record_result "Messaging benchmark test file exists" "PASS"
else
    record_result "Messaging benchmark test file exists" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 15: pprof memory profiling test exists (regression detection)
# --------------------------------------------------------------------------
if [ -f "$PROJECT_ROOT/tests/performance/pprof_leak_detection_test.go" ]; then
    record_result "pprof memory leak detection test exists" "PASS"
else
    record_result "pprof memory leak detection test exists" "FAIL"
fi

echo ""
echo "=========================================="
echo "  Results: $PASSED/$TOTAL passed, $FAILED failed"
echo "=========================================="

[ $FAILED -eq 0 ] && exit 0 || exit 1
