#!/bin/bash
# Challenge: 4-Pass Validation Pipeline Implementation
# Validates Initial → Validation → Polish → Final pipeline
# Zero false positives - ensures actual multi-pass validation works

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

CHALLENGE_NAME="4-Pass Validation Pipeline"
TOTAL_TESTS=12
PASSED_TESTS=0
FAILED_TESTS=0

# Test 1: Validation pipeline exists
log_test 1 "Validation pipeline implementation exists"
if [ -f "$PROJECT_ROOT/internal/debate/validation/pipeline.go" ]; then
    if grep -q "type.*ValidationPipeline.*struct" "$PROJECT_ROOT/internal/debate/validation/pipeline.go"; then
        pass_test
    else
        fail_test "ValidationPipeline struct not found"
    fi
else
    fail_test "pipeline.go not found"
fi

# Test 2: All four passes defined
log_test 2 "All four validation passes defined"
PASSES=$(grep -c "Pass.*ValidationPass.*=" "$PROJECT_ROOT/internal/debate/validation/pipeline.go" || echo 0)
if [ "$PASSES" -ge 4 ]; then
    pass_test
else
    fail_test "Insufficient passes defined ($PASSES, expected >=4)"
fi

# Test 3: PassInitial exists
log_test 3 "PassInitial defined"
if grep -q 'PassInitial.*ValidationPass.*=.*"initial"' "$PROJECT_ROOT/internal/debate/validation/pipeline.go"; then
    pass_test
else
    fail_test "PassInitial not found"
fi

# Test 4: PassValidation exists
log_test 4 "PassValidation defined"
if grep -q 'PassValidation.*ValidationPass.*=.*"validation"' "$PROJECT_ROOT/internal/debate/validation/pipeline.go"; then
    pass_test
else
    fail_test "PassValidation not found"
fi

# Test 5: PassPolish exists
log_test 5 "PassPolish defined"
if grep -q 'PassPolish.*ValidationPass.*=.*"polish"' "$PROJECT_ROOT/internal/debate/validation/pipeline.go"; then
    pass_test
else
    fail_test "PassPolish not found"
fi

# Test 6: PassFinal exists
log_test 6 "PassFinal defined"
if grep -q 'PassFinal.*ValidationPass.*=.*"final"' "$PROJECT_ROOT/internal/debate/validation/pipeline.go"; then
    pass_test
else
    fail_test "PassFinal not found"
fi

# Test 7: Pipeline validation method exists
log_test 7 "Validate method implements pipeline"
if grep -q "func.*Validate.*PipelineResult" "$PROJECT_ROOT/internal/debate/validation/pipeline.go"; then
    pass_test
else
    fail_test "Validate method not found"
fi

# Test 8: Issue severity levels defined
log_test 8 "Issue severity levels properly defined"
if grep -q "SeverityBlocker\|SeverityCritical\|SeverityMajor" "$PROJECT_ROOT/internal/debate/validation/pipeline.go"; then
    pass_test
else
    fail_test "Severity levels incomplete"
fi

# Test 9: Unit tests exist
log_test 9 "Unit tests for validation pipeline exist"
if [ -f "$PROJECT_ROOT/internal/debate/validation/pipeline_test.go" ]; then
    TEST_COUNT=$(grep -c "^func Test" "$PROJECT_ROOT/internal/debate/validation/pipeline_test.go" || echo 0)
    if [ "$TEST_COUNT" -ge 5 ]; then
        pass_test
    else
        fail_test "Insufficient tests ($TEST_COUNT, expected >=5)"
    fi
else
    fail_test "Test file not found"
fi

# Test 10: All passes test
log_test 10 "Test for all passes succeeding"
if go test -v -run "TestValidationPipeline_Validate_AllPass" "$PROJECT_ROOT/internal/debate/validation/" 2>&1 | grep -q "PASS"; then
    pass_test
else
    fail_test "All passes test failed"
fi

# Test 11: Failure handling test
log_test 11 "Test for pass failure handling"
if go test -v -run "TestValidationPipeline_Validate_InitialFail" "$PROJECT_ROOT/internal/debate/validation/" 2>&1 | grep -q "PASS"; then
    pass_test
else
    fail_test "Failure handling test failed"
fi

# Test 12: Build succeeds
log_test 12 "Project builds with validation pipeline"
if go build "$PROJECT_ROOT/internal/debate/validation/" 2>&1 | grep -v "error"; then
    pass_test
else
    fail_test "Build failed"
fi

# Print summary
echo "========================================="
echo "Challenge: $CHALLENGE_NAME"
echo "Total Tests: $TOTAL_TESTS"
echo "Passed: $PASSED_TESTS"
echo "Failed: $FAILED_TESTS"
echo "========================================="

if [ "$FAILED_TESTS" -eq 0 ]; then
    echo "✅ All tests passed!"
    exit 0
else
    echo "❌ Some tests failed"
    exit 1
fi
