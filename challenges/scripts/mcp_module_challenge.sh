#!/bin/bash
# MCP Module Challenge
# Validates the MCP module: code structure, compilation, tests, and core functionality.

set -euo pipefail

# Source framework
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/challenge_framework.sh"

init_challenge "mcp_module" "MCP Module"
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

run_test "MCP_Module directory exists" \
    "test -d '$PROJECT_ROOT/MCP_Module'"

run_test "MCP_Module go.mod exists" \
    "test -f '$PROJECT_ROOT/MCP_Module/go.mod'"

run_test "MCP module name correct" \
    "grep -q 'module digital.vasic.mcp' '$PROJECT_ROOT/MCP_Module/go.mod'"

run_test "Main go.mod has mcp require" \
    "grep -q 'digital.vasic.mcp' '$PROJECT_ROOT/go.mod'"

run_test "Main go.mod has mcp replace directive" \
    "grep -q 'replace digital.vasic.mcp => ./MCP_Module' '$PROJECT_ROOT/go.mod'"

# ============================================================================
# Section 2: Documentation
# ============================================================================
log_info "Section 2: Documentation"

run_test "README.md exists" \
    "test -f '$PROJECT_ROOT/MCP_Module/README.md'"

run_test "CLAUDE.md exists" \
    "test -f '$PROJECT_ROOT/MCP_Module/CLAUDE.md'"

run_test "AGENTS.md exists" \
    "test -f '$PROJECT_ROOT/MCP_Module/AGENTS.md'"

run_test "docs/ directory exists" \
    "test -d '$PROJECT_ROOT/MCP_Module/docs'"

# ============================================================================
# Section 3: Package Structure
# ============================================================================
log_info "Section 3: Package Structure"

run_test "pkg/adapter package exists" \
    "test -d '$PROJECT_ROOT/MCP_Module/pkg/adapter'"

run_test "pkg/client package exists" \
    "test -d '$PROJECT_ROOT/MCP_Module/pkg/client'"

run_test "pkg/config package exists" \
    "test -d '$PROJECT_ROOT/MCP_Module/pkg/config'"

run_test "pkg/protocol package exists" \
    "test -d '$PROJECT_ROOT/MCP_Module/pkg/protocol'"

run_test "pkg/registry package exists" \
    "test -d '$PROJECT_ROOT/MCP_Module/pkg/registry'"

run_test "pkg/server package exists" \
    "test -d '$PROJECT_ROOT/MCP_Module/pkg/server'"

# ============================================================================
# Section 4: Compilation
# ============================================================================
log_info "Section 4: Compilation"

run_test "MCP Module compiles" \
    "cd '$PROJECT_ROOT/MCP_Module' && go build ./..."

run_test "MCP Module passes go vet" \
    "cd '$PROJECT_ROOT/MCP_Module' && go vet ./..."

# ============================================================================
# Section 5: Unit Tests
# ============================================================================
log_info "Section 5: Unit Tests"

run_test "MCP Module unit tests pass" \
    "cd '$PROJECT_ROOT/MCP_Module' && GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -short -count=1 -p 1 ./..."

# ============================================================================
# Section 6: Test Type Spectrum
# ============================================================================
log_info "Section 6: Test Type Spectrum"

run_test "Integration tests exist" \
    "ls '$PROJECT_ROOT/MCP_Module/tests/integration/'*_test.go 2>/dev/null | head -1"

run_test "E2E tests exist" \
    "ls '$PROJECT_ROOT/MCP_Module/tests/e2e/'*_test.go 2>/dev/null | head -1"

run_test "Security tests exist" \
    "ls '$PROJECT_ROOT/MCP_Module/tests/security/'*_test.go 2>/dev/null | head -1"

run_test "Stress tests exist" \
    "ls '$PROJECT_ROOT/MCP_Module/tests/stress/'*_test.go 2>/dev/null | head -1"

run_test "Benchmark tests exist" \
    "ls '$PROJECT_ROOT/MCP_Module/tests/benchmark/'*_test.go 2>/dev/null | head -1"

# ============================================================================
# Section 7: Adapter Integration
# ============================================================================
log_info "Section 7: Adapter Integration"

run_test "MCP adapter exists in main project" \
    "test -f '$PROJECT_ROOT/internal/adapters/mcp/mcp.go'"

run_test "MCP adapter test exists" \
    "test -f '$PROJECT_ROOT/internal/adapters/mcp/mcp_test.go'"

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
