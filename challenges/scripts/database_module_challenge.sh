#!/bin/bash
# Database Module Challenge
# Validates the Database module: code structure, compilation, tests, and core functionality.

set -euo pipefail

# Source framework
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/challenge_framework.sh"

init_challenge "database_module" "Database Module"
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

run_test "Database directory exists" \
    "test -d '$PROJECT_ROOT/Database'"

run_test "Database go.mod exists" \
    "test -f '$PROJECT_ROOT/Database/go.mod'"

run_test "Database module name correct" \
    "grep -q 'module digital.vasic.database' '$PROJECT_ROOT/Database/go.mod'"

run_test "Main go.mod has database require" \
    "grep -q 'digital.vasic.database' '$PROJECT_ROOT/go.mod'"

run_test "Main go.mod has database replace directive" \
    "grep -q 'replace digital.vasic.database => ./Database' '$PROJECT_ROOT/go.mod'"

# ============================================================================
# Section 2: Documentation
# ============================================================================
log_info "Section 2: Documentation"

run_test "README.md exists" \
    "test -f '$PROJECT_ROOT/Database/README.md'"

run_test "CLAUDE.md exists" \
    "test -f '$PROJECT_ROOT/Database/CLAUDE.md'"

run_test "AGENTS.md exists" \
    "test -f '$PROJECT_ROOT/Database/AGENTS.md'"

run_test "docs/ directory exists" \
    "test -d '$PROJECT_ROOT/Database/docs'"

# ============================================================================
# Section 3: Package Structure
# ============================================================================
log_info "Section 3: Package Structure"

run_test "pkg/database package exists" \
    "test -d '$PROJECT_ROOT/Database/pkg/database'"

run_test "pkg/migration package exists" \
    "test -d '$PROJECT_ROOT/Database/pkg/migration'"

run_test "pkg/pool package exists" \
    "test -d '$PROJECT_ROOT/Database/pkg/pool'"

run_test "pkg/postgres package exists" \
    "test -d '$PROJECT_ROOT/Database/pkg/postgres'"

run_test "pkg/query package exists" \
    "test -d '$PROJECT_ROOT/Database/pkg/query'"

run_test "pkg/repository package exists" \
    "test -d '$PROJECT_ROOT/Database/pkg/repository'"

run_test "pkg/sqlite package exists" \
    "test -d '$PROJECT_ROOT/Database/pkg/sqlite'"

# ============================================================================
# Section 4: Compilation
# ============================================================================
log_info "Section 4: Compilation"

run_test "Database compiles" \
    "cd '$PROJECT_ROOT/Database' && go build ./..."

run_test "Database passes go vet" \
    "cd '$PROJECT_ROOT/Database' && go vet ./..."

# ============================================================================
# Section 5: Unit Tests
# ============================================================================
log_info "Section 5: Unit Tests"

run_test "Database unit tests pass" \
    "cd '$PROJECT_ROOT/Database' && GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -short -count=1 -p 1 ./..."

# ============================================================================
# Section 6: Test Type Spectrum
# ============================================================================
log_info "Section 6: Test Type Spectrum"

run_test "Integration tests exist" \
    "ls '$PROJECT_ROOT/Database/tests/integration/'*_test.go 2>/dev/null | head -1"

run_test "E2E tests exist" \
    "ls '$PROJECT_ROOT/Database/tests/e2e/'*_test.go 2>/dev/null | head -1"

run_test "Security tests exist" \
    "ls '$PROJECT_ROOT/Database/tests/security/'*_test.go 2>/dev/null | head -1"

run_test "Stress tests exist" \
    "ls '$PROJECT_ROOT/Database/tests/stress/'*_test.go 2>/dev/null | head -1"

run_test "Benchmark tests exist" \
    "ls '$PROJECT_ROOT/Database/tests/benchmark/'*_test.go 2>/dev/null | head -1"

# ============================================================================
# Section 7: Adapter Integration
# ============================================================================
log_info "Section 7: Adapter Integration"

run_test "Database adapter exists in main project" \
    "test -f '$PROJECT_ROOT/internal/adapters/database/adapter.go'"

run_test "Database adapter test exists" \
    "test -f '$PROJECT_ROOT/internal/adapters/database/adapter_test.go'"

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
