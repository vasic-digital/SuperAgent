#!/bin/bash
# HelixAgent Challenge - Security Scanning Validation
# Validates security scanning tools: .snyk, SBOM generation, gosec, go vet, Makefile targets

set -e

# Source challenge framework
source "$(dirname "${BASH_SOURCE[0]}")/challenge_framework.sh"

# Initialize challenge
init_challenge "security-scanning" "Security Scanning Validation"
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
# SECTION 1: SECURITY CONFIGURATION FILES
# ============================================================================
log_info "=============================================="
log_info "SECTION 1: Security Configuration Files"
log_info "=============================================="

run_test "Snyk configuration (.snyk) exists" \
    "[[ -f '$PROJECT_ROOT/.snyk' ]]"

run_test "SBOM generation script exists" \
    "[[ -f '$PROJECT_ROOT/scripts/generate-sbom.sh' ]]"

run_test "SBOM generation script is executable" \
    "[[ -x '$PROJECT_ROOT/scripts/generate-sbom.sh' ]]"

# ============================================================================
# SECTION 2: SECURITY TOOLS AVAILABILITY
# ============================================================================
log_info "=============================================="
log_info "SECTION 2: Security Tools Availability"
log_info "=============================================="

run_test "gosec is available" \
    "command -v gosec >/dev/null 2>&1 || [[ -f '$PROJECT_ROOT/bin/gosec' ]] || go tool -n gosec >/dev/null 2>&1"

run_test "go vet is available" \
    "go vet --help >/dev/null 2>&1"

# ============================================================================
# SECTION 3: STATIC ANALYSIS
# ============================================================================
log_info "=============================================="
log_info "SECTION 3: Static Analysis (go vet)"
log_info "=============================================="

VET_START=$(date +%s%N)
run_test "go vet passes on all packages" \
    "cd '$PROJECT_ROOT' && go vet ./..."
VET_END=$(date +%s%N)
VET_DURATION_MS=$(( (VET_END - VET_START) / 1000000 ))
record_metric "go_vet_time_ms" "$VET_DURATION_MS"

# ============================================================================
# SECTION 4: MAKEFILE SECURITY TARGETS
# ============================================================================
log_info "=============================================="
log_info "SECTION 4: Makefile Security Targets"
log_info "=============================================="

run_test "Makefile has security-scan target" \
    "grep -qE '^security-scan:' '$PROJECT_ROOT/Makefile'"

run_test "Makefile has sbom target" \
    "grep -qE '^sbom:' '$PROJECT_ROOT/Makefile'"

# ============================================================================
# SECTION 5: SECURITY TEST FILES
# ============================================================================
log_info "=============================================="
log_info "SECTION 5: Security Test Infrastructure"
log_info "=============================================="

run_test "Security test directory exists" \
    "[[ -d '$PROJECT_ROOT/tests/security' ]]"

run_test "Security tests exist" \
    "ls '$PROJECT_ROOT'/tests/security/*_test.go >/dev/null 2>&1"

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
    log_success "ALL SECURITY SCANNING TESTS PASSED!"
    log_success "=============================================="
    finalize_challenge "PASSED"
else
    log_error "=============================================="
    log_error "SOME TESTS FAILED"
    log_error "=============================================="
    finalize_challenge "FAILED"
fi
