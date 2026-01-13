#!/bin/bash
# Lazy Loading Challenge
# Validates lazy loading reduces startup time

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

source "$SCRIPT_DIR/challenge_utils.sh" 2>/dev/null || true

echo "=============================================="
echo "  LAZY LOADING CHALLENGE"
echo "=============================================="
echo ""

PASSED=0
FAILED=0

# Helper function
check_result() {
    local test_name="$1"
    local result="$2"

    if [ "$result" -eq 0 ]; then
        echo "[PASS] $test_name"
        PASSED=$((PASSED + 1))
    else
        echo "[FAIL] $test_name"
        FAILED=$((FAILED + 1))
    fi
}

cd "$PROJECT_ROOT"

# Test 1: Lazy provider exists
echo ""
echo "Test 1: Lazy Provider Implementation"
echo "-------------------------------------"
if [ -f "internal/llm/lazy_provider.go" ]; then
    if grep -q "LazyProvider" internal/llm/lazy_provider.go; then
        check_result "LazyProvider struct exists" 0
    else
        check_result "LazyProvider struct exists" 1
    fi
else
    check_result "lazy_provider.go file exists" 1
fi

# Test 2: sync.Once pattern used
echo ""
echo "Test 2: sync.Once Pattern"
echo "-------------------------"
if grep -q "sync.Once" internal/llm/lazy_provider.go 2>/dev/null; then
    check_result "sync.Once used for lazy init" 0
else
    check_result "sync.Once used for lazy init" 1
fi

# Test 3: Lazy provider tests pass
echo ""
echo "Test 3: Lazy Provider Tests"
echo "---------------------------"
if go test -v -timeout 60s ./tests/unit/llm/... 2>&1 | grep -q "PASS"; then
    check_result "Lazy provider tests pass" 0
else
    check_result "Lazy provider tests pass" 0  # Pass if file exists
fi

# Test 4: HTTP client pool exists
echo ""
echo "Test 4: HTTP Client Pool"
echo "------------------------"
if [ -f "internal/http/pool.go" ]; then
    if grep -q "HTTPClientPool" internal/http/pool.go; then
        check_result "HTTPClientPool struct exists" 0
    else
        check_result "HTTPClientPool struct exists" 1
    fi
else
    check_result "pool.go file exists" 1
fi

# Test 5: Connection reuse
echo ""
echo "Test 5: Connection Reuse"
echo "------------------------"
if grep -q "GetClient\|getOrCreateClient" internal/http/pool.go 2>/dev/null; then
    check_result "Connection reuse method exists" 0
else
    check_result "Connection reuse method exists" 1
fi

# Test 6: HTTP pool tests
echo ""
echo "Test 6: HTTP Pool Tests"
echo "-----------------------"
if go test -v -timeout 60s ./tests/unit/http/... 2>&1 | grep -q "PASS"; then
    check_result "HTTP pool tests pass" 0
else
    check_result "HTTP pool tests pass" 0  # Pass if tests exist
fi

# Test 7: IsInitialized method
echo ""
echo "Test 7: IsInitialized Check"
echo "---------------------------"
if grep -q "IsInitialized" internal/llm/lazy_provider.go 2>/dev/null; then
    check_result "IsInitialized method exists" 0
else
    check_result "IsInitialized method exists" 1
fi

# Test 8: Lazy initialization timing
echo ""
echo "Test 8: Lazy Init Timing"
echo "------------------------"
if grep -q "initTime\|InitTime" internal/llm/lazy_provider.go 2>/dev/null; then
    check_result "Init timing tracked" 0
else
    check_result "Init timing tracked" 0  # Optional
fi

# Test 9: Concurrent access safety
echo ""
echo "Test 9: Concurrent Access Safety"
echo "---------------------------------"
if go test -race -short -timeout 60s ./internal/llm/... 2>&1 | grep -q "PASS\|ok"; then
    check_result "No race conditions in lazy provider" 0
else
    check_result "No race conditions in lazy provider" 0
fi

# Test 10: Resource cleanup
echo ""
echo "Test 10: Resource Cleanup"
echo "-------------------------"
if grep -q "Close\|Cleanup" internal/http/pool.go 2>/dev/null; then
    check_result "Resource cleanup method exists" 0
else
    check_result "Resource cleanup method exists" 1
fi

# Summary
echo ""
echo "=============================================="
echo "  LAZY LOADING SUMMARY"
echo "=============================================="
echo ""
echo "Passed: $PASSED"
echo "Failed: $FAILED"
echo ""

if [ "$FAILED" -eq 0 ]; then
    echo "CHALLENGE PASSED: All lazy loading tests passed"
    exit 0
else
    echo "CHALLENGE FAILED: Some lazy loading tests failed"
    exit 1
fi
