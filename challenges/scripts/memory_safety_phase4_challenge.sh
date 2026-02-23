#!/bin/bash
# Phase 4: Memory Safety & Concurrency Challenge
# Validates race condition fixes and thread-safety improvements

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

echo "=========================================="
echo "Phase 4: Memory Safety & Concurrency Challenge"
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
    echo "Command: $test_cmd"
    
    if eval "$test_cmd" > /tmp/phase4_test_output.txt 2>&1; then
        echo "✓ PASSED"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo "✗ FAILED"
        cat /tmp/phase4_test_output.txt
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

run_race_test() {
    local test_name="$1"
    local test_pkg="$2"
    local test_run="$3"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo ""
    echo "Race Test $TOTAL_TESTS: $test_name"
    
    if GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -race -run "$test_run" "$test_pkg" > /tmp/phase4_race_output.txt 2>&1; then
        if grep -q "WARNING: DATA RACE" /tmp/phase4_race_output.txt; then
            echo "✗ FAILED - Race detected"
            cat /tmp/phase4_race_output.txt
            FAILED_TESTS=$((FAILED_TESTS + 1))
        else
            echo "✓ PASSED - No race detected"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        fi
    else
        echo "✗ FAILED - Test execution failed"
        cat /tmp/phase4_race_output.txt
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

cd "$PROJECT_ROOT"

echo ""
echo "=== Section 1: Messaging Adapter Race Tests ==="

run_race_test "Messaging adapter publish batch" "./internal/adapters/messaging" "TestInMemoryBrokerAdapter_PublishBatch"

echo ""
echo "=== Section 2: Messaging Hub Race Tests ==="

run_race_test "Messaging hub concurrent subscribe" "./internal/messaging" "TestMessagingHub_ConcurrentSubscribe"

echo ""
echo "=== Section 3: Verifier Adapter Tests ==="

run_test "Verifier adapter model display names" "go test -v -run TestGetModelDisplayName ./internal/verifier/adapters"

echo ""
echo "=== Section 4: Core Package Race Tests ==="

run_race_test "Adapters auth race" "./internal/adapters/auth" ".*"
run_race_test "Adapters containers race" "./internal/adapters/containers" ".*"
run_race_test "Adapters formatters race" "./internal/adapters/formatters" ".*"

echo ""
echo "=== Section 5: Build Verification ==="

run_test "Project builds successfully" "go build ./cmd/... ./internal/..."

echo ""
echo "=== Section 6: Memory Safety Pattern Checks ==="

TOTAL_TESTS=$((TOTAL_TESTS + 1))
echo ""
echo "Pattern Check $TOTAL_TESTS: Atomic FallbackUsages usage"
if grep -q "FallbackUsages.Add(1)" "$PROJECT_ROOT/internal/messaging/hub.go"; then
    echo "✓ PASSED - FallbackUsages uses atomic.Add(1)"
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    echo "✗ FAILED - FallbackUsages not using atomic operations"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi

TOTAL_TESTS=$((TOTAL_TESTS + 1))
echo ""
echo "Pattern Check $TOTAL_TESTS: No FallbackUsages++ remains"
if ! grep -q "FallbackUsages++" "$PROJECT_ROOT/internal/messaging/hub.go"; then
    echo "✓ PASSED - No unsafe FallbackUsages++ found"
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    echo "✗ FAILED - Unsafe FallbackUsages++ still present"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi

TOTAL_TESTS=$((TOTAL_TESTS + 1))
echo ""
echo "Pattern Check $TOTAL_TESTS: Atomic import in hub.go"
if grep -q '"sync/atomic"' "$PROJECT_ROOT/internal/messaging/hub.go"; then
    echo "✓ PASSED - sync/atomic imported in hub.go"
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    echo "✗ FAILED - sync/atomic not imported in hub.go"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi

TOTAL_TESTS=$((TOTAL_TESTS + 1))
echo ""
echo "Pattern Check $TOTAL_TESTS: Atomic import in messaging adapter test"
if grep -q '"sync/atomic"' "$PROJECT_ROOT/internal/adapters/messaging/inmemory_adapter_test.go"; then
    echo "✓ PASSED - sync/atomic imported in test"
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    echo "✗ FAILED - sync/atomic not imported in test"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi

echo ""
echo "=========================================="
echo "Phase 4 Memory Safety Challenge Summary"
echo "=========================================="
echo "Total Tests: $TOTAL_TESTS"
echo "Passed: $PASSED_TESTS"
echo "Failed: $FAILED_TESTS"
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    echo "✓ All Phase 4 memory safety tests PASSED"
    exit 0
else
    echo "✗ Some Phase 4 memory safety tests FAILED"
    exit 1
fi
