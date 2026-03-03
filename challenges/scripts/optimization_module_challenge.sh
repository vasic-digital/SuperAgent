#!/bin/bash
# Optimization Module Challenge
# Validates the Optimization module: code structure, compilation, tests, and core functionality.

set -e

# Source framework
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/challenge_framework.sh"

init_challenge "optimization_module" "Optimization Module"
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

run_test "Optimization directory exists" \
    "test -d '$PROJECT_ROOT/Optimization'"

run_test "Optimization go.mod exists" \
    "test -f '$PROJECT_ROOT/Optimization/go.mod'"

run_test "Optimization module name correct" \
    "grep -q 'module digital.vasic.optimization' '$PROJECT_ROOT/Optimization/go.mod'"

run_test "Main go.mod has optimization require" \
    "grep -q 'digital.vasic.optimization' '$PROJECT_ROOT/go.mod'"

run_test "Main go.mod has optimization replace directive" \
    "grep -q 'replace digital.vasic.optimization => ./Optimization' '$PROJECT_ROOT/go.mod'"

# ============================================================================
# Section 2: Documentation
# ============================================================================
log_info "Section 2: Documentation"

run_test "README.md exists" \
    "test -f '$PROJECT_ROOT/Optimization/README.md'"

run_test "CLAUDE.md exists" \
    "test -f '$PROJECT_ROOT/Optimization/CLAUDE.md'"

run_test "AGENTS.md exists" \
    "test -f '$PROJECT_ROOT/Optimization/AGENTS.md'"

run_test "docs/ directory exists" \
    "test -d '$PROJECT_ROOT/Optimization/docs'"

# ============================================================================
# Section 3: Package Structure
# ============================================================================
log_info "Section 3: Package Structure"

run_test "pkg/gptcache package exists" \
    "test -d '$PROJECT_ROOT/Optimization/pkg/gptcache'"

run_test "pkg/outlines package exists" \
    "test -d '$PROJECT_ROOT/Optimization/pkg/outlines'"

run_test "pkg/streaming package exists" \
    "test -d '$PROJECT_ROOT/Optimization/pkg/streaming'"

run_test "pkg/sglang package exists" \
    "test -d '$PROJECT_ROOT/Optimization/pkg/sglang'"

run_test "pkg/prompt package exists" \
    "test -d '$PROJECT_ROOT/Optimization/pkg/prompt'"

run_test "pkg/adapter package exists" \
    "test -d '$PROJECT_ROOT/Optimization/pkg/adapter'"

# ============================================================================
# Section 4: Compilation
# ============================================================================
log_info "Section 4: Compilation"

run_test "Optimization compiles" \
    "cd '$PROJECT_ROOT/Optimization' && go build ./..."

run_test "Optimization passes go vet" \
    "cd '$PROJECT_ROOT/Optimization' && go vet ./..."

# ============================================================================
# Section 5: Unit Tests
# ============================================================================
log_info "Section 5: Unit Tests"

run_test "Optimization unit tests pass" \
    "cd '$PROJECT_ROOT/Optimization' && GOMAXPROCS=2 go test -short -count=1 -p 1 ./..."

# ============================================================================
# Section 6: Test Type Spectrum
# ============================================================================
log_info "Section 6: Test Type Spectrum"

run_test "Integration tests exist" \
    "ls '$PROJECT_ROOT/Optimization/tests/integration/'*_test.go 2>/dev/null | head -1"

run_test "E2E tests exist" \
    "ls '$PROJECT_ROOT/Optimization/tests/e2e/'*_test.go 2>/dev/null | head -1"

run_test "Security tests exist" \
    "ls '$PROJECT_ROOT/Optimization/tests/security/'*_test.go 2>/dev/null | head -1"

run_test "Stress tests exist" \
    "ls '$PROJECT_ROOT/Optimization/tests/stress/'*_test.go 2>/dev/null | head -1"

run_test "Benchmark tests exist" \
    "ls '$PROJECT_ROOT/Optimization/tests/benchmark/'*_test.go 2>/dev/null | head -1"

# ============================================================================
# Section 7: Adapter Integration
# ============================================================================
log_info "Section 7: Adapter Integration"

run_test "Optimization adapter exists in main project" \
    "test -f '$PROJECT_ROOT/internal/adapters/optimization/adapter.go'"

run_test "Optimization adapter test exists" \
    "test -f '$PROJECT_ROOT/internal/adapters/optimization/adapter_test.go'"

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
