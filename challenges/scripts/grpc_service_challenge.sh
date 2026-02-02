#!/bin/bash
# HelixAgent Challenge - gRPC Service Validation
# Validates gRPC server build, protobuf definitions, tests, and health methods

set -e

# Source challenge framework
source "$(dirname "${BASH_SOURCE[0]}")/challenge_framework.sh"

# Initialize challenge
init_challenge "grpc-service" "gRPC Service Validation"
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
# SECTION 1: gRPC SERVER SOURCE FILES
# ============================================================================
log_info "=============================================="
log_info "SECTION 1: gRPC Server Source Files"
log_info "=============================================="

run_test "gRPC server main.go exists" \
    "[[ -f '$PROJECT_ROOT/cmd/grpc-server/main.go' ]]"

run_test "gRPC server test file exists" \
    "[[ -f '$PROJECT_ROOT/cmd/grpc-server/main_test.go' ]]"

# ============================================================================
# SECTION 2: PROTOBUF DEFINITIONS
# ============================================================================
log_info "=============================================="
log_info "SECTION 2: Protobuf Definitions"
log_info "=============================================="

run_test "Protobuf generated code exists (pb.go)" \
    "ls '$PROJECT_ROOT'/pkg/api/*.pb.go >/dev/null 2>&1"

run_test "gRPC generated code exists (grpc.pb.go)" \
    "ls '$PROJECT_ROOT'/pkg/api/*_grpc.pb.go >/dev/null 2>&1"

# ============================================================================
# SECTION 3: gRPC SERVER BUILD
# ============================================================================
log_info "=============================================="
log_info "SECTION 3: gRPC Server Build"
log_info "=============================================="

log_info "Building gRPC server binary (to /dev/null)..."
BUILD_START=$(date +%s%N)
run_test "gRPC server builds successfully" \
    "cd '$PROJECT_ROOT' && go build -o /dev/null ./cmd/grpc-server/"
BUILD_END=$(date +%s%N)
BUILD_DURATION_MS=$(( (BUILD_END - BUILD_START) / 1000000 ))
record_metric "grpc_build_time_ms" "$BUILD_DURATION_MS"

# ============================================================================
# SECTION 4: gRPC SERVER TESTS
# ============================================================================
log_info "=============================================="
log_info "SECTION 4: gRPC Server Tests"
log_info "=============================================="

log_info "Running gRPC server tests..."
TEST_START=$(date +%s%N)
run_test "gRPC server tests pass" \
    "cd '$PROJECT_ROOT' && go test -v -count=1 ./cmd/grpc-server/... -timeout 120s"
TEST_END=$(date +%s%N)
TEST_DURATION_MS=$(( (TEST_END - TEST_START) / 1000000 ))
record_metric "grpc_test_time_ms" "$TEST_DURATION_MS"

# ============================================================================
# SECTION 5: gRPC SERVICE IMPLEMENTATION
# ============================================================================
log_info "=============================================="
log_info "SECTION 5: gRPC Service Implementation"
log_info "=============================================="

run_test "LLMFacadeServer struct defined" \
    "grep -q 'type LLMFacadeServer struct' '$PROJECT_ROOT/cmd/grpc-server/main.go'"

run_test "gRPC server registers service (RegisterLLMFacade)" \
    "grep -qE 'Register.*Server|RegisterLLMFacade' '$PROJECT_ROOT/cmd/grpc-server/main.go'"

run_test "gRPC server listens on network port" \
    "grep -qE 'net.Listen|grpc.NewServer' '$PROJECT_ROOT/cmd/grpc-server/main.go'"

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
    log_success "ALL gRPC SERVICE TESTS PASSED!"
    log_success "=============================================="
    finalize_challenge "PASSED"
else
    log_error "=============================================="
    log_error "SOME TESTS FAILED"
    log_error "=============================================="
    finalize_challenge "FAILED"
fi
