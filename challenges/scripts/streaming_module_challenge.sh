#!/bin/bash
# Streaming Module Challenge
# Validates the Streaming module: code structure, compilation, tests, and core functionality.

set -e

# Source framework
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/challenge_framework.sh"

init_challenge "streaming_module" "Streaming Module"
load_env

TESTS_PASSED=0
TESTS_FAILED=0
TESTS_TOTAL=0

run_test() {
    local test_name="$1"
    local test_cmd="$2"
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    if eval "$test_cmd" >> "$LOG_FILE" 2>&1; then
        log_success "PASS: $test_name"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        record_assertion "test" "$test_name" "true" ""
        return 0
    else
        log_error "FAIL: $test_name"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        record_assertion "test" "$test_name" "false" "Test command failed"
        return 1
    fi
}

# ============================================================================
# Section 1: Module Structure
# ============================================================================
log_info "Section 1: Module Structure"

run_test "Streaming directory exists" \
    "test -d '$PROJECT_ROOT/Streaming'"

run_test "Streaming go.mod exists" \
    "test -f '$PROJECT_ROOT/Streaming/go.mod'"

run_test "Streaming module name correct" \
    "grep -q 'module digital.vasic.streaming' '$PROJECT_ROOT/Streaming/go.mod'"

run_test "Main go.mod has streaming require" \
    "grep -q 'digital.vasic.streaming' '$PROJECT_ROOT/go.mod'"

run_test "Main go.mod has streaming replace directive" \
    "grep -q 'replace digital.vasic.streaming => ./Streaming' '$PROJECT_ROOT/go.mod'"

# ============================================================================
# Section 2: Documentation
# ============================================================================
log_info "Section 2: Documentation"

run_test "README.md exists" \
    "test -f '$PROJECT_ROOT/Streaming/README.md'"

run_test "CLAUDE.md exists" \
    "test -f '$PROJECT_ROOT/Streaming/CLAUDE.md'"

run_test "AGENTS.md exists" \
    "test -f '$PROJECT_ROOT/Streaming/AGENTS.md'"

run_test "docs/ directory exists" \
    "test -d '$PROJECT_ROOT/Streaming/docs'"

# ============================================================================
# Section 3: Package Structure
# ============================================================================
log_info "Section 3: Package Structure"

run_test "pkg/sse package exists" \
    "test -d '$PROJECT_ROOT/Streaming/pkg/sse'"

run_test "pkg/websocket package exists" \
    "test -d '$PROJECT_ROOT/Streaming/pkg/websocket'"

run_test "pkg/grpc package exists" \
    "test -d '$PROJECT_ROOT/Streaming/pkg/grpc'"

run_test "pkg/webhook package exists" \
    "test -d '$PROJECT_ROOT/Streaming/pkg/webhook'"

run_test "pkg/http package exists" \
    "test -d '$PROJECT_ROOT/Streaming/pkg/http'"

run_test "pkg/transport package exists" \
    "test -d '$PROJECT_ROOT/Streaming/pkg/transport'"

# ============================================================================
# Section 4: Compilation
# ============================================================================
log_info "Section 4: Compilation"

run_test "Streaming compiles" \
    "cd '$PROJECT_ROOT/Streaming' && go build ./..."

run_test "Streaming passes go vet" \
    "cd '$PROJECT_ROOT/Streaming' && go vet ./..."

# ============================================================================
# Section 5: Unit Tests
# ============================================================================
log_info "Section 5: Unit Tests"

run_test "Streaming unit tests pass" \
    "cd '$PROJECT_ROOT/Streaming' && GOMAXPROCS=2 go test -short -count=1 -p 1 ./..."

# ============================================================================
# Section 6: Test Type Spectrum
# ============================================================================
log_info "Section 6: Test Type Spectrum"

run_test "Integration tests exist" \
    "ls '$PROJECT_ROOT/Streaming/tests/integration/'*_test.go 2>/dev/null | head -1"

run_test "E2E tests exist" \
    "ls '$PROJECT_ROOT/Streaming/tests/e2e/'*_test.go 2>/dev/null | head -1"

run_test "Security tests exist" \
    "ls '$PROJECT_ROOT/Streaming/tests/security/'*_test.go 2>/dev/null | head -1"

run_test "Stress tests exist" \
    "ls '$PROJECT_ROOT/Streaming/tests/stress/'*_test.go 2>/dev/null | head -1"

run_test "Benchmark tests exist" \
    "ls '$PROJECT_ROOT/Streaming/tests/benchmark/'*_test.go 2>/dev/null | head -1"

# ============================================================================
# Section 7: Adapter Integration
# ============================================================================
log_info "Section 7: Adapter Integration"

run_test "Streaming adapter exists in main project" \
    "test -f '$PROJECT_ROOT/internal/adapters/streaming/adapter.go'"

run_test "Streaming adapter test exists" \
    "test -f '$PROJECT_ROOT/internal/adapters/streaming/adapter_test.go'"

# ============================================================================
# Summary
# ============================================================================
echo ""
log_info "=========================================="
log_info "Challenge Results: $CHALLENGE_NAME"
log_info "=========================================="
log_info "Passed: $TESTS_PASSED / $TESTS_TOTAL"
if [ "$TESTS_FAILED" -gt 0 ]; then
    log_error "Failed: $TESTS_FAILED"
fi

if [[ $TESTS_FAILED -eq 0 ]]; then
    finalize_challenge "PASSED"
else
    finalize_challenge "FAILED"
fi
