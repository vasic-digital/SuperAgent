#!/bin/bash
# HelixAgent Challenge - BigData Comprehensive Validation
# Validates all BigData components: source files, compilation, tests, stress tests, coverage

set -e

# Source challenge framework
source "$(dirname "${BASH_SOURCE[0]}")/challenge_framework.sh"

# Initialize challenge
init_challenge "bigdata-comprehensive" "BigData Comprehensive Validation"
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
# SECTION 1: BIGDATA SOURCE FILES
# ============================================================================
log_info "=============================================="
log_info "SECTION 1: BigData Source Files"
log_info "=============================================="

BIGDATA_FILES=(
    "config_converter.go"
    "datalake.go"
    "handler.go"
    "integration.go"
    "spark_processor.go"
    "debate_integration.go"
    "memory_integration.go"
    "entity_integration.go"
    "analytics_integration.go"
    "debate_wrapper.go"
)

for src_file in "${BIGDATA_FILES[@]}"; do
    run_test "BigData source: $src_file exists" \
        "[[ -f '$PROJECT_ROOT/internal/bigdata/$src_file' ]]"
done

# ============================================================================
# SECTION 2: BIGDATA TEST FILES
# ============================================================================
log_info "=============================================="
log_info "SECTION 2: BigData Test Files"
log_info "=============================================="

BIGDATA_TEST_FILES=(
    "config_converter_test.go"
    "datalake_test.go"
    "handler_test.go"
    "integration_test.go"
    "spark_processor_test.go"
    "debate_integration_test.go"
    "memory_integration_test.go"
    "entity_integration_test.go"
    "analytics_integration_test.go"
    "debate_wrapper_test.go"
)

for test_file in "${BIGDATA_TEST_FILES[@]}"; do
    run_test "BigData test: $test_file exists" \
        "[[ -f '$PROJECT_ROOT/internal/bigdata/$test_file' ]]"
done

# ============================================================================
# SECTION 3: BIGDATA COMPILATION
# ============================================================================
log_info "=============================================="
log_info "SECTION 3: BigData Package Compilation"
log_info "=============================================="

BUILD_START=$(date +%s%N)
run_test "BigData package compiles successfully" \
    "cd '$PROJECT_ROOT' && go build ./internal/bigdata/"
BUILD_END=$(date +%s%N)
BUILD_DURATION_MS=$(( (BUILD_END - BUILD_START) / 1000000 ))
record_metric "bigdata_build_time_ms" "$BUILD_DURATION_MS"

# ============================================================================
# SECTION 4: BIGDATA UNIT TESTS
# ============================================================================
log_info "=============================================="
log_info "SECTION 4: BigData Unit Tests"
log_info "=============================================="

TEST_START=$(date +%s%N)
run_test "BigData tests pass" \
    "cd '$PROJECT_ROOT' && go test -v -count=1 ./internal/bigdata/... -timeout 120s"
TEST_END=$(date +%s%N)
TEST_DURATION_MS=$(( (TEST_END - TEST_START) / 1000000 ))
record_metric "bigdata_test_time_ms" "$TEST_DURATION_MS"

# ============================================================================
# SECTION 5: BIGDATA STRESS TESTS
# ============================================================================
log_info "=============================================="
log_info "SECTION 5: BigData Stress Tests"
log_info "=============================================="

STRESS_START=$(date +%s%N)
run_test "BigData stress tests pass" \
    "cd '$PROJECT_ROOT' && go test -v -count=1 -run TestBigdata ./tests/stress/... -timeout 180s"
STRESS_END=$(date +%s%N)
STRESS_DURATION_MS=$(( (STRESS_END - STRESS_START) / 1000000 ))
record_metric "bigdata_stress_test_time_ms" "$STRESS_DURATION_MS"

# ============================================================================
# SECTION 6: BIGDATA TEST COVERAGE
# ============================================================================
log_info "=============================================="
log_info "SECTION 6: BigData Test Coverage"
log_info "=============================================="

COVERAGE_FILE="$OUTPUT_DIR/logs/bigdata_coverage.out"
cd "$PROJECT_ROOT"
if go test -coverprofile="$COVERAGE_FILE" ./internal/bigdata/... -timeout 120s >> "$LOG_FILE" 2>&1; then
    COVERAGE=$(go tool cover -func="$COVERAGE_FILE" 2>/dev/null | grep total | awk '{print $3}' | tr -d '%')
    record_metric "bigdata_coverage_percent" "$COVERAGE"
    log_info "BigData test coverage: ${COVERAGE}%"

    # Check coverage >= 80%
    COVERAGE_INT=${COVERAGE%%.*}
    if [[ "$COVERAGE_INT" -ge 80 ]]; then
        log_success "PASS: BigData coverage >= 80% (${COVERAGE}%)"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        record_assertion "coverage" "bigdata_coverage_80" "true" "Coverage: ${COVERAGE}%"
    else
        log_error "FAIL: BigData coverage < 80% (${COVERAGE}%)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        record_assertion "coverage" "bigdata_coverage_80" "false" "Coverage: ${COVERAGE}% (expected >= 80%)"
    fi
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
else
    log_error "FAIL: Could not generate coverage report"
    TESTS_FAILED=$((TESTS_FAILED + 1))
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    record_assertion "coverage" "bigdata_coverage_80" "false" "Coverage report generation failed"
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
    log_success "ALL BIGDATA TESTS PASSED!"
    log_success "=============================================="
    finalize_challenge "PASSED"
else
    log_error "=============================================="
    log_error "SOME TESTS FAILED"
    log_error "=============================================="
    finalize_challenge "FAILED"
fi
