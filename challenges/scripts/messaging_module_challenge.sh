#!/bin/bash
# Messaging Module Challenge
# Validates the Messaging module: code structure, compilation, tests, and core functionality.

set -euo pipefail

# Source framework
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/challenge_framework.sh"

init_challenge "messaging_module" "Messaging Module"
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

run_test "Messaging directory exists" \
    "test -d '$PROJECT_ROOT/Messaging'"

run_test "Messaging go.mod exists" \
    "test -f '$PROJECT_ROOT/Messaging/go.mod'"

run_test "Messaging module name correct" \
    "grep -q 'module digital.vasic.messaging' '$PROJECT_ROOT/Messaging/go.mod'"

run_test "Main go.mod has messaging require" \
    "grep -q 'digital.vasic.messaging' '$PROJECT_ROOT/go.mod'"

run_test "Main go.mod has messaging replace directive" \
    "grep -q 'replace digital.vasic.messaging => ./Messaging' '$PROJECT_ROOT/go.mod'"

# ============================================================================
# Section 2: Documentation
# ============================================================================
log_info "Section 2: Documentation"

run_test "README.md exists" \
    "test -f '$PROJECT_ROOT/Messaging/README.md'"

run_test "CLAUDE.md exists" \
    "test -f '$PROJECT_ROOT/Messaging/CLAUDE.md'"

run_test "AGENTS.md exists" \
    "test -f '$PROJECT_ROOT/Messaging/AGENTS.md'"

run_test "docs/ directory exists" \
    "test -d '$PROJECT_ROOT/Messaging/docs'"

# ============================================================================
# Section 3: Package Structure
# ============================================================================
log_info "Section 3: Package Structure"

run_test "pkg/broker package exists" \
    "test -d '$PROJECT_ROOT/Messaging/pkg/broker'"

run_test "pkg/consumer package exists" \
    "test -d '$PROJECT_ROOT/Messaging/pkg/consumer'"

run_test "pkg/kafka package exists" \
    "test -d '$PROJECT_ROOT/Messaging/pkg/kafka'"

run_test "pkg/producer package exists" \
    "test -d '$PROJECT_ROOT/Messaging/pkg/producer'"

run_test "pkg/rabbitmq package exists" \
    "test -d '$PROJECT_ROOT/Messaging/pkg/rabbitmq'"

# ============================================================================
# Section 4: Compilation
# ============================================================================
log_info "Section 4: Compilation"

run_test "Messaging compiles" \
    "cd '$PROJECT_ROOT/Messaging' && go build ./..."

run_test "Messaging passes go vet" \
    "cd '$PROJECT_ROOT/Messaging' && go vet ./..."

# ============================================================================
# Section 5: Unit Tests
# ============================================================================
log_info "Section 5: Unit Tests"

run_test "Messaging unit tests pass" \
    "cd '$PROJECT_ROOT/Messaging' && GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -short -count=1 -p 1 ./..."

# ============================================================================
# Section 6: Test Type Spectrum
# ============================================================================
log_info "Section 6: Test Type Spectrum"

run_test "Integration tests exist" \
    "ls '$PROJECT_ROOT/Messaging/tests/integration/'*_test.go 2>/dev/null | head -1"

run_test "E2E tests exist" \
    "ls '$PROJECT_ROOT/Messaging/tests/e2e/'*_test.go 2>/dev/null | head -1"

run_test "Security tests exist" \
    "ls '$PROJECT_ROOT/Messaging/tests/security/'*_test.go 2>/dev/null | head -1"

run_test "Stress tests exist" \
    "ls '$PROJECT_ROOT/Messaging/tests/stress/'*_test.go 2>/dev/null | head -1"

run_test "Benchmark tests exist" \
    "ls '$PROJECT_ROOT/Messaging/tests/benchmark/'*_test.go 2>/dev/null | head -1"

# ============================================================================
# Section 7: Adapter Integration
# ============================================================================
log_info "Section 7: Adapter Integration"

run_test "Messaging adapter exists in main project" \
    "test -f '$PROJECT_ROOT/internal/adapters/messaging/adapter.go'"

run_test "Messaging adapter test exists" \
    "test -f '$PROJECT_ROOT/internal/adapters/messaging/adapter_test.go'"

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
