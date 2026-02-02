#!/bin/bash
# HelixAgent Challenge - Memory System Validation
# Validates memory system: CRDT, distributed manager, event sourcing, store, tests, coverage

set -e

# Source challenge framework
source "$(dirname "${BASH_SOURCE[0]}")/challenge_framework.sh"

# Initialize challenge
init_challenge "memory-system" "Memory System Validation"
load_env

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_TOTAL=0

# Test function
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
# SECTION 1: MEMORY SYSTEM SOURCE FILES
# ============================================================================
log_info "=============================================="
log_info "SECTION 1: Memory System Source Files"
log_info "=============================================="

MEMORY_FILES=(
    "crdt.go"
    "distributed_manager.go"
    "event_sourcing.go"
    "store_memory.go"
    "manager.go"
    "types.go"
    "doc.go"
)

for src_file in "${MEMORY_FILES[@]}"; do
    run_test "Memory source: $src_file exists" \
        "[[ -f '$PROJECT_ROOT/internal/memory/$src_file' ]]"
done

# ============================================================================
# SECTION 2: MEMORY SYSTEM TEST FILES
# ============================================================================
log_info "=============================================="
log_info "SECTION 2: Memory System Test Files"
log_info "=============================================="

MEMORY_TEST_FILES=(
    "crdt_test.go"
    "distributed_manager_test.go"
    "event_sourcing_test.go"
    "memory_test.go"
)

for test_file in "${MEMORY_TEST_FILES[@]}"; do
    run_test "Memory test: $test_file exists" \
        "[[ -f '$PROJECT_ROOT/internal/memory/$test_file' ]]"
done

# ============================================================================
# SECTION 3: MEMORY PACKAGE COMPILATION
# ============================================================================
log_info "=============================================="
log_info "SECTION 3: Memory Package Compilation"
log_info "=============================================="

BUILD_START=$(date +%s%N)
run_test "Memory package compiles successfully" \
    "cd '$PROJECT_ROOT' && go build ./internal/memory/"
BUILD_END=$(date +%s%N)
BUILD_DURATION_MS=$(( (BUILD_END - BUILD_START) / 1000000 ))
record_metric "memory_build_time_ms" "$BUILD_DURATION_MS"

# ============================================================================
# SECTION 4: MEMORY UNIT TESTS
# ============================================================================
log_info "=============================================="
log_info "SECTION 4: Memory Unit Tests"
log_info "=============================================="

TEST_START=$(date +%s%N)
run_test "Memory tests pass" \
    "cd '$PROJECT_ROOT' && go test -v -count=1 ./internal/memory/... -timeout 120s"
TEST_END=$(date +%s%N)
TEST_DURATION_MS=$(( (TEST_END - TEST_START) / 1000000 ))
record_metric "memory_test_time_ms" "$TEST_DURATION_MS"

# ============================================================================
# SECTION 5: MEMORY STRESS TESTS
# ============================================================================
log_info "=============================================="
log_info "SECTION 5: Memory Stress Tests"
log_info "=============================================="

STRESS_START=$(date +%s%N)
run_test "Memory stress tests pass" \
    "cd '$PROJECT_ROOT' && go test -v -count=1 -run TestMemory ./tests/stress/... -timeout 180s"
STRESS_END=$(date +%s%N)
STRESS_DURATION_MS=$(( (STRESS_END - STRESS_START) / 1000000 ))
record_metric "memory_stress_test_time_ms" "$STRESS_DURATION_MS"

# ============================================================================
# SECTION 6: MEMORY TEST COVERAGE
# ============================================================================
log_info "=============================================="
log_info "SECTION 6: Memory Test Coverage"
log_info "=============================================="

COVERAGE_FILE="$OUTPUT_DIR/logs/memory_coverage.out"
cd "$PROJECT_ROOT"
if go test -coverprofile="$COVERAGE_FILE" ./internal/memory/... -timeout 120s >> "$LOG_FILE" 2>&1; then
    COVERAGE=$(go tool cover -func="$COVERAGE_FILE" 2>/dev/null | grep total | awk '{print $3}' | tr -d '%')
    record_metric "memory_coverage_percent" "$COVERAGE"
    log_info "Memory test coverage: ${COVERAGE}%"

    # Check coverage >= 80%
    COVERAGE_INT=${COVERAGE%%.*}
    if [[ "$COVERAGE_INT" -ge 80 ]]; then
        log_success "PASS: Memory coverage >= 80% (${COVERAGE}%)"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        record_assertion "coverage" "memory_coverage_80" "true" "Coverage: ${COVERAGE}%"
    else
        log_error "FAIL: Memory coverage < 80% (${COVERAGE}%)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        record_assertion "coverage" "memory_coverage_80" "false" "Coverage: ${COVERAGE}% (expected >= 80%)"
    fi
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
else
    log_error "FAIL: Could not generate coverage report"
    TESTS_FAILED=$((TESTS_FAILED + 1))
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    record_assertion "coverage" "memory_coverage_80" "false" "Coverage report generation failed"
fi

# ============================================================================
# SUMMARY
# ============================================================================
log_info "=============================================="
log_info "CHALLENGE SUMMARY"
log_info "=============================================="

log_info "Total tests: $TESTS_TOTAL"
log_info "Passed: $TESTS_PASSED"
log_info "Failed: $TESTS_FAILED"

record_metric "total_tests" "$TESTS_TOTAL"
record_metric "passed_tests" "$TESTS_PASSED"
record_metric "failed_tests" "$TESTS_FAILED"

if [[ $TESTS_FAILED -eq 0 ]]; then
    log_success "=============================================="
    log_success "ALL MEMORY SYSTEM TESTS PASSED!"
    log_success "=============================================="
    finalize_challenge "PASSED"
else
    log_error "=============================================="
    log_error "SOME TESTS FAILED"
    log_error "=============================================="
    finalize_challenge "FAILED"
fi
