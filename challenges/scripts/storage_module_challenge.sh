#!/bin/bash
# Storage Module Challenge
# Validates the Storage module: code structure, compilation, tests, and core functionality.

set -e

# Source framework
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/challenge_framework.sh"

init_challenge "storage_module" "Storage Module"
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

run_test "Storage directory exists" \
    "test -d '$PROJECT_ROOT/Storage'"

run_test "Storage go.mod exists" \
    "test -f '$PROJECT_ROOT/Storage/go.mod'"

run_test "Storage module name correct" \
    "grep -q 'module digital.vasic.storage' '$PROJECT_ROOT/Storage/go.mod'"

run_test "Main go.mod has storage require" \
    "grep -q 'digital.vasic.storage' '$PROJECT_ROOT/go.mod'"

run_test "Main go.mod has storage replace directive" \
    "grep -q 'replace digital.vasic.storage => ./Storage' '$PROJECT_ROOT/go.mod'"

# ============================================================================
# Section 2: Documentation
# ============================================================================
log_info "Section 2: Documentation"

run_test "README.md exists" \
    "test -f '$PROJECT_ROOT/Storage/README.md'"

run_test "CLAUDE.md exists" \
    "test -f '$PROJECT_ROOT/Storage/CLAUDE.md'"

run_test "AGENTS.md exists" \
    "test -f '$PROJECT_ROOT/Storage/AGENTS.md'"

run_test "docs/ directory exists" \
    "test -d '$PROJECT_ROOT/Storage/docs'"

# ============================================================================
# Section 3: Package Structure
# ============================================================================
log_info "Section 3: Package Structure"

run_test "pkg/local package exists" \
    "test -d '$PROJECT_ROOT/Storage/pkg/local'"

run_test "pkg/object package exists" \
    "test -d '$PROJECT_ROOT/Storage/pkg/object'"

run_test "pkg/provider package exists" \
    "test -d '$PROJECT_ROOT/Storage/pkg/provider'"

run_test "pkg/s3 package exists" \
    "test -d '$PROJECT_ROOT/Storage/pkg/s3'"

# ============================================================================
# Section 4: Compilation
# ============================================================================
log_info "Section 4: Compilation"

run_test "Storage compiles" \
    "cd '$PROJECT_ROOT/Storage' && go build ./..."

run_test "Storage passes go vet" \
    "cd '$PROJECT_ROOT/Storage' && go vet ./..."

# ============================================================================
# Section 5: Unit Tests
# ============================================================================
log_info "Section 5: Unit Tests"

run_test "Storage unit tests pass" \
    "cd '$PROJECT_ROOT/Storage' && GOMAXPROCS=2 go test -short -count=1 -p 1 ./..."

# ============================================================================
# Section 6: Test Type Spectrum
# ============================================================================
log_info "Section 6: Test Type Spectrum"

run_test "Integration tests exist" \
    "ls '$PROJECT_ROOT/Storage/tests/integration/'*_test.go 2>/dev/null | head -1"

run_test "E2E tests exist" \
    "ls '$PROJECT_ROOT/Storage/tests/e2e/'*_test.go 2>/dev/null | head -1"

run_test "Security tests exist" \
    "ls '$PROJECT_ROOT/Storage/tests/security/'*_test.go 2>/dev/null | head -1"

run_test "Stress tests exist" \
    "ls '$PROJECT_ROOT/Storage/tests/stress/'*_test.go 2>/dev/null | head -1"

run_test "Benchmark tests exist" \
    "ls '$PROJECT_ROOT/Storage/tests/benchmark/'*_test.go 2>/dev/null | head -1"

# ============================================================================
# Section 7: Adapter Integration
# ============================================================================
log_info "Section 7: Adapter Integration"

run_test "Storage adapter directory exists in main project" \
    "test -d '$PROJECT_ROOT/internal/adapters/storage'"

run_test "Storage MinIO adapter exists" \
    "test -f '$PROJECT_ROOT/internal/adapters/storage/minio/adapter.go'"

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
