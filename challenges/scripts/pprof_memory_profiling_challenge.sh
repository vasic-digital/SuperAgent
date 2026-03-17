#!/usr/bin/env bash
set -euo pipefail

# pprof Memory Profiling Challenge
# Validates that pprof debugging endpoints are properly wired, memory
# profiling tests exist, and goroutine monitoring is in place.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

setup_challenge "pprof_memory_profiling" "$@"

PASSED=0
FAILED=0
TOTAL=0
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

print_header "pprof Memory Profiling Challenge"
echo "Validates pprof endpoints, memory leak detection, and goroutine monitoring"
echo ""

# Test 1: pprof import exists in router
if grep -q "net/http/pprof" "$PROJECT_ROOT/internal/router/router.go"; then
    record_result "pprof imported in router" "PASS"
else
    record_result "pprof imported in router" "FAIL"
fi

# Test 2: pprof endpoints registered
for endpoint in "/debug/pprof/" "/debug/pprof/heap" "/debug/pprof/goroutine" "/debug/pprof/profile"; do
    if grep -q "$endpoint" "$PROJECT_ROOT/internal/router/router.go"; then
        record_result "pprof endpoint $endpoint registered" "PASS"
    else
        record_result "pprof endpoint $endpoint registered" "FAIL"
    fi
done

# Test 3: ENABLE_PPROF environment variable check exists
if grep -q "ENABLE_PPROF" "$PROJECT_ROOT/internal/router/router.go"; then
    record_result "ENABLE_PPROF env var check exists" "PASS"
else
    record_result "ENABLE_PPROF env var check exists" "FAIL"
fi

# Test 4: Goroutine monitoring test exists
if [ -f "$PROJECT_ROOT/tests/monitoring/resource_monitoring_test.go" ]; then
    record_result "Goroutine monitoring test file exists" "PASS"
else
    record_result "Goroutine monitoring test file exists" "FAIL"
fi

# Test 5: Memory profiling test exists
if [ -f "$PROJECT_ROOT/tests/performance/pprof_leak_detection_test.go" ]; then
    record_result "pprof leak detection test file exists" "PASS"
else
    record_result "pprof leak detection test file exists" "FAIL"
fi

# Test 6: Goroutine leak detection in pprof test
if grep -q "TestMemoryLeak_GoroutineCount" "$PROJECT_ROOT/tests/performance/pprof_leak_detection_test.go"; then
    record_result "Goroutine leak test function exists" "PASS"
else
    record_result "Goroutine leak test function exists" "FAIL"
fi

# Test 7: Heap growth detection in pprof test
if grep -q "TestMemoryLeak_HeapGrowth" "$PROJECT_ROOT/tests/performance/pprof_leak_detection_test.go"; then
    record_result "Heap growth test function exists" "PASS"
else
    record_result "Heap growth test function exists" "FAIL"
fi

# Test 8: Channel draining test exists
if grep -q "TestMemoryLeak_ChannelDraining" "$PROJECT_ROOT/tests/performance/pprof_leak_detection_test.go"; then
    record_result "Channel draining leak test exists" "PASS"
else
    record_result "Channel draining leak test exists" "FAIL"
fi

# Test 9: runtime.MemStats usage for profiling
if grep -q "runtime.MemStats" "$PROJECT_ROOT/tests/performance/pprof_leak_detection_test.go"; then
    record_result "runtime.MemStats used for memory profiling" "PASS"
else
    record_result "runtime.MemStats used for memory profiling" "FAIL"
fi

# Test 10: runtime.NumGoroutine usage for goroutine tracking
if grep -q "runtime.NumGoroutine" "$PROJECT_ROOT/tests/performance/pprof_leak_detection_test.go"; then
    record_result "runtime.NumGoroutine used for goroutine tracking" "PASS"
else
    record_result "runtime.NumGoroutine used for goroutine tracking" "FAIL"
fi

# Test 11: pprof test compiles
cd "$PROJECT_ROOT"
if GOMAXPROCS=2 go test -run=XXX_NOMATCH ./tests/performance/ 2>/dev/null; then
    record_result "pprof leak detection tests compile" "PASS"
else
    record_result "pprof leak detection tests compile" "FAIL"
fi

# Test 12: pprof mutex endpoint registered
if grep -q "/debug/pprof/mutex" "$PROJECT_ROOT/internal/router/router.go"; then
    record_result "pprof mutex endpoint registered" "PASS"
else
    record_result "pprof mutex endpoint registered" "FAIL"
fi

echo ""
print_summary "pprof Memory Profiling Challenge" "$PASSED" "$FAILED"
[ "$FAILED" -eq 0 ] && exit 0 || exit 1
