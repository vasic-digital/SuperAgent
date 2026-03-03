#!/bin/bash
# Concurrency Module Challenge
# Validates the Concurrency module: code structure, compilation, tests, and core functionality.

set -e

# Source framework
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/challenge_framework.sh"

init_challenge "concurrency_module" "Concurrency Module"
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

run_test "Concurrency directory exists" \
    "test -d '$PROJECT_ROOT/Concurrency'"

run_test "Concurrency go.mod exists" \
    "test -f '$PROJECT_ROOT/Concurrency/go.mod'"

run_test "Concurrency module name correct" \
    "grep -q 'module digital.vasic.concurrency' '$PROJECT_ROOT/Concurrency/go.mod'"

run_test "Main go.mod has concurrency require" \
    "grep -q 'digital.vasic.concurrency' '$PROJECT_ROOT/go.mod'"

run_test "Main go.mod has concurrency replace directive" \
    "grep -q 'replace digital.vasic.concurrency => ./Concurrency' '$PROJECT_ROOT/go.mod'"

# ============================================================================
# Section 2: Documentation
# ============================================================================
log_info "Section 2: Documentation"

run_test "README.md exists" \
    "test -f '$PROJECT_ROOT/Concurrency/README.md'"

run_test "CLAUDE.md exists" \
    "test -f '$PROJECT_ROOT/Concurrency/CLAUDE.md'"

run_test "AGENTS.md exists" \
    "test -f '$PROJECT_ROOT/Concurrency/AGENTS.md'"

run_test "docs/ directory exists" \
    "test -d '$PROJECT_ROOT/Concurrency/docs'"

# ============================================================================
# Section 3: Package Structure
# ============================================================================
log_info "Section 3: Package Structure"

run_test "pkg/pool package exists" \
    "test -d '$PROJECT_ROOT/Concurrency/pkg/pool'"

run_test "pkg/queue package exists" \
    "test -d '$PROJECT_ROOT/Concurrency/pkg/queue'"

run_test "pkg/limiter package exists" \
    "test -d '$PROJECT_ROOT/Concurrency/pkg/limiter'"

run_test "pkg/breaker package exists" \
    "test -d '$PROJECT_ROOT/Concurrency/pkg/breaker'"

run_test "pkg/semaphore package exists" \
    "test -d '$PROJECT_ROOT/Concurrency/pkg/semaphore'"

run_test "pkg/monitor package exists" \
    "test -d '$PROJECT_ROOT/Concurrency/pkg/monitor'"

# ============================================================================
# Section 4: Compilation
# ============================================================================
log_info "Section 4: Compilation"

run_test "Concurrency compiles" \
    "cd '$PROJECT_ROOT/Concurrency' && go build ./..."

run_test "Concurrency passes go vet" \
    "cd '$PROJECT_ROOT/Concurrency' && go vet ./..."

# ============================================================================
# Section 5: Unit Tests
# ============================================================================
log_info "Section 5: Unit Tests"

run_test "Concurrency unit tests pass" \
    "cd '$PROJECT_ROOT/Concurrency' && GOMAXPROCS=2 go test -short -count=1 -p 1 ./..."

# ============================================================================
# Section 6: Test Type Spectrum
# ============================================================================
log_info "Section 6: Test Type Spectrum"

run_test "Integration tests exist" \
    "ls '$PROJECT_ROOT/Concurrency/tests/integration/'*_test.go 2>/dev/null | head -1"

run_test "E2E tests exist" \
    "ls '$PROJECT_ROOT/Concurrency/tests/e2e/'*_test.go 2>/dev/null | head -1"

run_test "Security tests exist" \
    "ls '$PROJECT_ROOT/Concurrency/tests/security/'*_test.go 2>/dev/null | head -1"

run_test "Stress tests exist" \
    "ls '$PROJECT_ROOT/Concurrency/tests/stress/'*_test.go 2>/dev/null | head -1"

run_test "Benchmark tests exist" \
    "ls '$PROJECT_ROOT/Concurrency/tests/benchmark/'*_test.go 2>/dev/null | head -1"

# ============================================================================
# Section 7: Adapter Integration
# ============================================================================
log_info "Section 7: Adapter Integration"

run_test "Concurrency used in main project internal packages" \
    "grep -r 'digital.vasic.concurrency' '$PROJECT_ROOT/internal/' 2>/dev/null | head -1"

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
