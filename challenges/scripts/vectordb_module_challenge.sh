#!/bin/bash
# VectorDB Module Challenge
# Validates the VectorDB module: code structure, compilation, tests, and core functionality.

set -euo pipefail

# Source framework
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/challenge_framework.sh"

init_challenge "vectordb_module" "VectorDB Module"
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

run_test "VectorDB directory exists" \
    "test -d '$PROJECT_ROOT/VectorDB'"

run_test "VectorDB go.mod exists" \
    "test -f '$PROJECT_ROOT/VectorDB/go.mod'"

run_test "VectorDB module name correct" \
    "grep -q 'module digital.vasic.vectordb' '$PROJECT_ROOT/VectorDB/go.mod'"

run_test "Main go.mod has vectordb require" \
    "grep -q 'digital.vasic.vectordb' '$PROJECT_ROOT/go.mod'"

run_test "Main go.mod has vectordb replace directive" \
    "grep -q 'replace digital.vasic.vectordb => ./VectorDB' '$PROJECT_ROOT/go.mod'"

# ============================================================================
# Section 2: Documentation
# ============================================================================
log_info "Section 2: Documentation"

run_test "README.md exists" \
    "test -f '$PROJECT_ROOT/VectorDB/README.md'"

run_test "CLAUDE.md exists" \
    "test -f '$PROJECT_ROOT/VectorDB/CLAUDE.md'"

run_test "AGENTS.md exists" \
    "test -f '$PROJECT_ROOT/VectorDB/AGENTS.md'"

run_test "docs/ directory exists" \
    "test -d '$PROJECT_ROOT/VectorDB/docs'"

# ============================================================================
# Section 3: Package Structure
# ============================================================================
log_info "Section 3: Package Structure"

run_test "pkg/client package exists" \
    "test -d '$PROJECT_ROOT/VectorDB/pkg/client'"

run_test "pkg/qdrant package exists" \
    "test -d '$PROJECT_ROOT/VectorDB/pkg/qdrant'"

run_test "pkg/pinecone package exists" \
    "test -d '$PROJECT_ROOT/VectorDB/pkg/pinecone'"

run_test "pkg/milvus package exists" \
    "test -d '$PROJECT_ROOT/VectorDB/pkg/milvus'"

run_test "pkg/pgvector package exists" \
    "test -d '$PROJECT_ROOT/VectorDB/pkg/pgvector'"

# ============================================================================
# Section 4: Compilation
# ============================================================================
log_info "Section 4: Compilation"

run_test "VectorDB compiles" \
    "cd '$PROJECT_ROOT/VectorDB' && go build ./..."

run_test "VectorDB passes go vet" \
    "cd '$PROJECT_ROOT/VectorDB' && go vet ./..."

# ============================================================================
# Section 5: Unit Tests
# ============================================================================
log_info "Section 5: Unit Tests"

run_test "VectorDB unit tests pass" \
    "cd '$PROJECT_ROOT/VectorDB' && GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -short -count=1 -p 1 ./..."

# ============================================================================
# Section 6: Test Type Spectrum
# ============================================================================
log_info "Section 6: Test Type Spectrum"

run_test "Integration tests exist" \
    "ls '$PROJECT_ROOT/VectorDB/tests/integration/'*_test.go 2>/dev/null | head -1"

run_test "E2E tests exist" \
    "ls '$PROJECT_ROOT/VectorDB/tests/e2e/'*_test.go 2>/dev/null | head -1"

run_test "Security tests exist" \
    "ls '$PROJECT_ROOT/VectorDB/tests/security/'*_test.go 2>/dev/null | head -1"

run_test "Stress tests exist" \
    "ls '$PROJECT_ROOT/VectorDB/tests/stress/'*_test.go 2>/dev/null | head -1"

run_test "Benchmark tests exist" \
    "ls '$PROJECT_ROOT/VectorDB/tests/benchmark/'*_test.go 2>/dev/null | head -1"

# ============================================================================
# Section 7: Adapter Integration
# ============================================================================
log_info "Section 7: Adapter Integration"

run_test "VectorDB adapter exists in main project" \
    "test -d '$PROJECT_ROOT/internal/adapters/vectordb/qdrant' && test -f '$PROJECT_ROOT/internal/adapters/vectordb/qdrant/adapter.go'"

run_test "VectorDB adapter test exists" \
    "test -f '$PROJECT_ROOT/internal/adapters/vectordb/qdrant/adapter_test.go'"

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
