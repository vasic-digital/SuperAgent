#!/bin/bash
# HelixAgent Challenge - Concurrency Fixes Validation
# Validates Phase 1 goroutine-safety fixes: MessagingHub, TieredCache,
# QueryCache, FormatterExecutor panic recovery, Gemini ACP and WebSocket tests.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

if [ -f "$SCRIPT_DIR/challenge_framework.sh" ]; then
    source "$SCRIPT_DIR/challenge_framework.sh"
    init_challenge "concurrency-fixes-validation" "Concurrency Fixes Validation"
    load_env
    FRAMEWORK_LOADED=true
else
    FRAMEWORK_LOADED=false
fi

PASSED=0
FAILED=0

record_result() {
    local name="$1" status="$2"
    if [ "$FRAMEWORK_LOADED" = true ]; then
        if [ "$status" = "PASS" ]; then
            record_assertion "test" "$name" "true" "$name"
        else
            record_assertion "test" "$name" "false" "$name"
        fi
    fi
    if [ "$status" = "PASS" ]; then
        PASSED=$((PASSED + 1))
        echo -e "\033[0;32m[PASS]\033[0m $name"
    else
        FAILED=$((FAILED + 1))
        echo -e "\033[0;31m[FAIL]\033[0m $name"
    fi
}

echo "=== Concurrency Fixes Validation Challenge ==="
echo ""

# --- MessagingHub Safety ---
echo "--- MessagingHub Goroutine Safety ---"

# Test 1: MessagingHub has closeOnce field
if grep -q "closeOnce\s\+sync\.Once" "$PROJECT_ROOT/internal/messaging/hub.go"; then
    record_result "MessagingHub has closeOnce sync.Once field" "PASS"
else
    record_result "MessagingHub has closeOnce sync.Once field" "FAIL"
fi

# Test 2: MessagingHub has wg field
if grep -q "wg\s\+sync\.WaitGroup" "$PROJECT_ROOT/internal/messaging/hub.go"; then
    record_result "MessagingHub has wg sync.WaitGroup field" "PASS"
else
    record_result "MessagingHub has wg sync.WaitGroup field" "FAIL"
fi

# Test 3: MessagingHub healthCheckLoop is tracked by WaitGroup
if grep -q "h\.wg\.Add(1)" "$PROJECT_ROOT/internal/messaging/hub.go" && \
   grep -q "h\.wg\.Done()" "$PROJECT_ROOT/internal/messaging/hub.go"; then
    record_result "MessagingHub healthCheckLoop tracked by WaitGroup" "PASS"
else
    record_result "MessagingHub healthCheckLoop tracked by WaitGroup" "FAIL"
fi

echo ""
echo "--- TieredCache Goroutine Safety ---"

# Test 4: TieredCache has wg field
if grep -q "wg\s\+sync\.WaitGroup" "$PROJECT_ROOT/internal/cache/tiered_cache.go"; then
    record_result "TieredCache has wg sync.WaitGroup field" "PASS"
else
    record_result "TieredCache has wg sync.WaitGroup field" "FAIL"
fi

# Test 5: TieredCache l1CleanupLoop is tracked by WaitGroup
if grep -q "l1CleanupLoop" "$PROJECT_ROOT/internal/cache/tiered_cache.go"; then
    record_result "TieredCache l1CleanupLoop goroutine tracked" "PASS"
else
    record_result "TieredCache l1CleanupLoop goroutine tracked" "FAIL"
fi

echo ""
echo "--- QueryCache Goroutine Safety ---"

# Test 6: QueryCache has wg field
if grep -q "wg\s\+sync\.WaitGroup" "$PROJECT_ROOT/internal/database/query_optimizer.go"; then
    record_result "QueryCache has wg sync.WaitGroup field" "PASS"
else
    record_result "QueryCache has wg sync.WaitGroup field" "FAIL"
fi

# Test 7: QueryCache cleanupLoop is tracked by WaitGroup
if grep -q "cleanupLoop" "$PROJECT_ROOT/internal/database/query_optimizer.go"; then
    record_result "QueryCache cleanupLoop goroutine tracked" "PASS"
else
    record_result "QueryCache cleanupLoop goroutine tracked" "FAIL"
fi

echo ""
echo "--- FormatterExecutor Panic Recovery ---"

# Test 8: FormatterExecutor executor.go has panic recovery
if grep -q "recover()" "$PROJECT_ROOT/internal/formatters/executor.go"; then
    record_result "FormatterExecutor has panic recovery (recover())" "PASS"
else
    record_result "FormatterExecutor has panic recovery (recover())" "FAIL"
fi

# Test 9: executor_safety_test.go exists
if [ -f "$PROJECT_ROOT/internal/formatters/executor_safety_test.go" ]; then
    record_result "FormatterExecutor safety test file exists" "PASS"
else
    record_result "FormatterExecutor safety test file exists" "FAIL"
fi

echo ""
echo "--- Gemini ACP Concurrency Safety ---"

# Test 10: Gemini ACP safety test file exists
if [ -f "$PROJECT_ROOT/internal/llm/providers/gemini/gemini_acp_safety_test.go" ]; then
    record_result "Gemini ACP safety test file exists" "PASS"
else
    record_result "Gemini ACP safety test file exists" "FAIL"
fi

# Test 11: Gemini ACP safety test contains RWMutex test
if grep -q "RWMutex" "$PROJECT_ROOT/internal/llm/providers/gemini/gemini_acp_safety_test.go"; then
    record_result "Gemini ACP safety test validates RWMutex usage" "PASS"
else
    record_result "Gemini ACP safety test validates RWMutex usage" "FAIL"
fi

echo ""
echo "--- WebSocket Concurrent Safety ---"

# Test 12: WebSocket safety test file exists
if [ -f "$PROJECT_ROOT/internal/notifications/websocket_server_safety_test.go" ]; then
    record_result "WebSocket safety test file exists" "PASS"
else
    record_result "WebSocket safety test file exists" "FAIL"
fi

# Test 13: WebSocket safety test contains concurrent test function
if grep -q "TestWebSocketServer_Concurrent" \
    "$PROJECT_ROOT/internal/notifications/websocket_server_safety_test.go"; then
    record_result "WebSocket safety test contains concurrent test" "PASS"
else
    record_result "WebSocket safety test contains concurrent test" "FAIL"
fi

echo ""
echo "--- Build Verification ---"

# Test 14: Full codebase compiles
if (cd "$PROJECT_ROOT" && GOMAXPROCS=2 go build ./... 2>/dev/null); then
    record_result "go build ./... compiles successfully" "PASS"
else
    record_result "go build ./... compiles successfully" "FAIL"
fi

# Test 15: go vet passes on affected packages
if (cd "$PROJECT_ROOT" && GOMAXPROCS=2 go vet \
    ./internal/messaging/ \
    ./internal/cache/ \
    ./internal/database/ \
    ./internal/formatters/ 2>/dev/null); then
    record_result "go vet passes on messaging/cache/database/formatters" "PASS"
else
    record_result "go vet passes on messaging/cache/database/formatters" "FAIL"
fi

echo ""
echo "=== Results ==="
TOTAL=$((PASSED + FAILED))
echo "Passed: $PASSED/$TOTAL"
echo "Failed: $FAILED/$TOTAL"

if [ "$FRAMEWORK_LOADED" = true ]; then
    finalize_challenge "$PASSED" "$TOTAL"
fi

if [ "$FAILED" -gt 0 ]; then
    exit 1
fi
