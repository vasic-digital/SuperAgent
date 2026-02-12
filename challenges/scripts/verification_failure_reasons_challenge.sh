#!/bin/bash
# Verification Failure Reasons Challenge
# Validates that provider verification failures include actionable reasons
# throughout the pipeline: struct fields, helper functions, API response, and tests.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

CHALLENGE_NAME="Verification Failure Reasons"
TOTAL_TESTS=15
PASSED=0
FAILED=0

log_info "========================================="
log_info "$CHALLENGE_NAME"
log_info "========================================="

PROJECT_ROOT="/run/media/milosvasic/DATA4TB/Projects/HelixAgent"

#===============================================================================
# Section 1: Code Structure (5 tests)
#===============================================================================
log_info "--- Section 1: Code Structure ---"

# Test 1: ProviderTestDetail struct exists
log_info "Test 1: ProviderTestDetail struct defined"
if grep -q "type ProviderTestDetail struct" "$PROJECT_ROOT/internal/verifier/provider_types.go" 2>/dev/null; then
  log_success "ProviderTestDetail struct defined"
  PASSED=$((PASSED + 1))
else
  log_error "ProviderTestDetail struct not found"
  FAILED=$((FAILED + 1))
fi

# Test 2: FailureReason field exists in UnifiedProvider
log_info "Test 2: FailureReason field in UnifiedProvider"
if grep -q 'FailureReason.*string.*json:"failure_reason' "$PROJECT_ROOT/internal/verifier/provider_types.go" 2>/dev/null; then
  log_success "FailureReason field found"
  PASSED=$((PASSED + 1))
else
  log_error "FailureReason field not found"
  FAILED=$((FAILED + 1))
fi

# Test 3: FailureCategory field exists in UnifiedProvider
log_info "Test 3: FailureCategory field in UnifiedProvider"
if grep -q 'FailureCategory.*string.*json:"failure_category' "$PROJECT_ROOT/internal/verifier/provider_types.go" 2>/dev/null; then
  log_success "FailureCategory field found"
  PASSED=$((PASSED + 1))
else
  log_error "FailureCategory field not found"
  FAILED=$((FAILED + 1))
fi

# Test 4: buildFailureReason function exists
log_info "Test 4: buildFailureReason function defined"
if grep -q "func buildFailureReason" "$PROJECT_ROOT/internal/verifier/startup.go" 2>/dev/null; then
  log_success "buildFailureReason function found"
  PASSED=$((PASSED + 1))
else
  log_error "buildFailureReason function not found"
  FAILED=$((FAILED + 1))
fi

# Test 5: categorizeFailure function exists
log_info "Test 5: categorizeFailure function defined"
if grep -q "func categorizeFailure" "$PROJECT_ROOT/internal/verifier/startup.go" 2>/dev/null; then
  log_success "categorizeFailure function found"
  PASSED=$((PASSED + 1))
else
  log_error "categorizeFailure function not found"
  FAILED=$((FAILED + 1))
fi

#===============================================================================
# Section 2: API Response (3 tests)
#===============================================================================
log_info "--- Section 2: API Response Integration ---"

# Test 6: failure_reason exposed in API endpoint
log_info "Test 6: failure_reason in API endpoint"
if grep -q '"failure_reason"' "$PROJECT_ROOT/cmd/helixagent/main.go" 2>/dev/null; then
  log_success "failure_reason in API endpoint"
  PASSED=$((PASSED + 1))
else
  log_error "failure_reason not in API endpoint"
  FAILED=$((FAILED + 1))
fi

# Test 7: failure_category exposed in API endpoint
log_info "Test 7: failure_category in API endpoint"
if grep -q '"failure_category"' "$PROJECT_ROOT/cmd/helixagent/main.go" 2>/dev/null; then
  log_success "failure_category in API endpoint"
  PASSED=$((PASSED + 1))
else
  log_error "failure_category not in API endpoint"
  FAILED=$((FAILED + 1))
fi

# Test 8: test_details exposed in API endpoint
log_info "Test 8: test_details in API endpoint"
if grep -q '"test_details"' "$PROJECT_ROOT/cmd/helixagent/main.go" 2>/dev/null; then
  log_success "test_details in API endpoint"
  PASSED=$((PASSED + 1))
else
  log_error "test_details not in API endpoint"
  FAILED=$((FAILED + 1))
fi

#===============================================================================
# Section 3: Unit Tests (4 tests)
#===============================================================================
log_info "--- Section 3: Unit Tests ---"

# Test 9: Test file exists
log_info "Test 9: startup_failure_reasons_test.go exists"
if [ -f "$PROJECT_ROOT/internal/verifier/startup_failure_reasons_test.go" ]; then
  log_success "Test file exists"
  PASSED=$((PASSED + 1))
else
  log_error "Test file not found"
  FAILED=$((FAILED + 1))
fi

# Test 10: Sufficient test count
log_info "Test 10: Sufficient test functions"
TEST_COUNT=$(grep -c "^func Test" "$PROJECT_ROOT/internal/verifier/startup_failure_reasons_test.go" 2>/dev/null || echo "0")
TEST_COUNT=${TEST_COUNT//[^0-9]/}
TEST_COUNT=${TEST_COUNT:-0}
if [ "$TEST_COUNT" -ge 8 ]; then
  log_success "Found $TEST_COUNT test functions (>=8)"
  PASSED=$((PASSED + 1))
else
  log_error "Expected >=8 test functions, found: $TEST_COUNT"
  FAILED=$((FAILED + 1))
fi

# Test 11: Benchmark exists
log_info "Test 11: Benchmark function exists"
if grep -q "^func Benchmark" "$PROJECT_ROOT/internal/verifier/startup_failure_reasons_test.go" 2>/dev/null; then
  log_success "Benchmark function found"
  PASSED=$((PASSED + 1))
else
  log_error "Benchmark function not found"
  FAILED=$((FAILED + 1))
fi

# Test 12: Run failure reason tests
log_info "Test 12: Failure reason tests pass"
if go test -run "TestBuildFailureReason|TestCategorizeFailure|TestMapTestDetails|TestProviderTestDetail|TestUnifiedProvider_Failure|TestTruncate|TestPopulate" -timeout 30s "$PROJECT_ROOT/internal/verifier/" > /dev/null 2>&1; then
  log_success "All failure reason tests passed"
  PASSED=$((PASSED + 1))
else
  log_error "Failure reason tests failed"
  FAILED=$((FAILED + 1))
fi

#===============================================================================
# Section 4: Functional Validation (3 tests)
#===============================================================================
log_info "--- Section 4: Functional Validation ---"

# Test 13: Project builds successfully
log_info "Test 13: Project builds"
if go build "$PROJECT_ROOT/..." > /dev/null 2>&1; then
  log_success "Project builds successfully"
  PASSED=$((PASSED + 1))
else
  log_error "Project build failed"
  FAILED=$((FAILED + 1))
fi

# Test 14: All verifier tests pass
log_info "Test 14: All verifier tests pass"
if go test -short -timeout 60s "$PROJECT_ROOT/internal/verifier/" > /dev/null 2>&1; then
  log_success "All verifier tests pass"
  PASSED=$((PASSED + 1))
else
  log_error "Some verifier tests failed"
  FAILED=$((FAILED + 1))
fi

# Test 15: Failure category constants are defined
log_info "Test 15: Failure category constants defined"
CATEGORY_COUNT=$(grep -c 'FailureCategory[A-Z]' "$PROJECT_ROOT/internal/verifier/provider_types.go" 2>/dev/null || echo "0")
CATEGORY_COUNT=${CATEGORY_COUNT//[^0-9]/}
CATEGORY_COUNT=${CATEGORY_COUNT:-0}
if [ "$CATEGORY_COUNT" -ge 7 ]; then
  log_success "Found $CATEGORY_COUNT failure category constants (>=7)"
  PASSED=$((PASSED + 1))
else
  log_error "Expected >=7 failure category constants, found: $CATEGORY_COUNT"
  FAILED=$((FAILED + 1))
fi

# Summary
log_info "========================================="
log_info "Test Results: $PASSED passed, $FAILED failed out of $TOTAL_TESTS"
log_info "========================================="

if [ $FAILED -eq 0 ]; then
  log_success "All Verification Failure Reasons tests passed!"
  exit 0
else
  log_error "$FAILED Verification Failure Reasons tests failed"
  exit 1
fi
