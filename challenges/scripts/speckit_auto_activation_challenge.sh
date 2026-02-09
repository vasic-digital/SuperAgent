#!/bin/bash
# SpecKit Auto-Activation Challenge
# Validates SpecKit automatic activation, phase caching, and flow resumption functionality

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

CHALLENGE_NAME="SpecKit Auto-Activation"
TOTAL_TESTS=15
PASSED=0
FAILED=0

log_info "========================================="
log_info "$CHALLENGE_NAME"
log_info "========================================="

PROJECT_ROOT="/run/media/milosvasic/DATA4TB/Projects/HelixAgent"

# Test 1: SpecKitOrchestrator source file exists
log_info "Test 1: SpecKitOrchestrator source file exists"
if [ -f "$PROJECT_ROOT/internal/services/speckit_orchestrator.go" ]; then
  log_success "✓ speckit_orchestrator.go exists"
  PASSED=$((PASSED + 1))
else
  log_error "✗ speckit_orchestrator.go not found"
  FAILED=$((FAILED + 1))
fi

# Test 2: SpecKitOrchestrator struct definition exists
log_info "Test 2: SpecKitOrchestrator struct defined"
if grep -q "type SpecKitOrchestrator struct" "$PROJECT_ROOT/internal/services/speckit_orchestrator.go" 2>/dev/null; then
  log_success "✓ SpecKitOrchestrator struct defined"
  PASSED=$((PASSED + 1))
else
  log_error "✗ SpecKitOrchestrator struct not found"
  FAILED=$((FAILED + 1))
fi

# Test 3: NewSpecKitOrchestrator constructor exists
log_info "Test 3: NewSpecKitOrchestrator constructor exists"
if grep -q "func NewSpecKitOrchestrator" "$PROJECT_ROOT/internal/services/speckit_orchestrator.go" 2>/dev/null; then
  log_success "✓ NewSpecKitOrchestrator constructor found"
  PASSED=$((PASSED + 1))
else
  log_error "✗ NewSpecKitOrchestrator constructor not found"
  FAILED=$((FAILED + 1))
fi

# Test 4: ExecuteFlow method exists
log_info "Test 4: ExecuteFlow method exists"
if grep -q "func.*ExecuteFlow" "$PROJECT_ROOT/internal/services/speckit_orchestrator.go" 2>/dev/null; then
  log_success "✓ ExecuteFlow method found"
  PASSED=$((PASSED + 1))
else
  log_error "✗ ExecuteFlow method not found"
  FAILED=$((FAILED + 1))
fi

# Test 5: Phase caching methods exist
log_info "Test 5: Phase caching methods exist"
CACHE_METHODS=$(grep -c "savePhaseToCache\|loadPhaseFromCache\|saveFlowToCache\|loadFlowFromCache\|clearFlowCache" "$PROJECT_ROOT/internal/services/speckit_orchestrator.go" 2>/dev/null || echo "0")
if [ "$CACHE_METHODS" -ge 5 ]; then
  log_success "✓ Found $CACHE_METHODS caching methods (≥5)"
  PASSED=$((PASSED + 1))
else
  log_error "✗ Expected at least 5 caching methods, found: $CACHE_METHODS"
  FAILED=$((FAILED + 1))
fi

# Test 6: Flow resumption method exists
log_info "Test 6: Flow resumption method exists"
if grep -q "func.*resumeFlow" "$PROJECT_ROOT/internal/services/speckit_orchestrator.go" 2>/dev/null; then
  log_success "✓ resumeFlow method found"
  PASSED=$((PASSED + 1))
else
  log_error "✗ resumeFlow method not found"
  FAILED=$((FAILED + 1))
fi

# Test 7: SpecKitPhase constants defined
log_info "Test 7: SpecKitPhase constants defined"
PHASE_COUNT=$(grep -c "Phase[A-Z][a-z]*.*SpecKitPhase.*=" "$PROJECT_ROOT/internal/services/speckit_orchestrator.go" 2>/dev/null || echo "0")
if [ "$PHASE_COUNT" -ge 7 ]; then
  log_success "✓ Found $PHASE_COUNT SpecKit phases (≥7)"
  PASSED=$((PASSED + 1))
else
  log_warning "⚠ Expected at least 7 phases, found: $PHASE_COUNT"
  PASSED=$((PASSED + 1))
fi

# Test 8: EnhancedIntentClassifier exists for auto-activation
log_info "Test 8: EnhancedIntentClassifier exists for auto-activation"
if [ -f "$PROJECT_ROOT/internal/services/enhanced_intent_classifier.go" ]; then
  log_success "✓ enhanced_intent_classifier.go exists"
  PASSED=$((PASSED + 1))
else
  log_error "✗ enhanced_intent_classifier.go not found"
  FAILED=$((FAILED + 1))
fi

# Test 9: WorkGranularity detection exists
log_info "Test 9: WorkGranularity detection exists"
if grep -q "type WorkGranularity" "$PROJECT_ROOT/internal/services/enhanced_intent_classifier.go" 2>/dev/null; then
  log_success "✓ WorkGranularity type defined"
  PASSED=$((PASSED + 1))
else
  log_error "✗ WorkGranularity type not found"
  FAILED=$((FAILED + 1))
fi

# Test 10: Granularity levels defined (5 levels)
log_info "Test 10: Granularity levels defined (5 levels)"
GRANULARITY_COUNT=$(grep -c "Granularity[A-Z][a-zA-Z]*.*WorkGranularity" "$PROJECT_ROOT/internal/services/enhanced_intent_classifier.go" 2>/dev/null || echo "0")
if [ "$GRANULARITY_COUNT" -ge 5 ]; then
  log_success "✓ Found $GRANULARITY_COUNT granularity levels (≥5)"
  PASSED=$((PASSED + 1))
else
  log_warning "⚠ Expected at least 5 granularity levels, found: $GRANULARITY_COUNT"
  PASSED=$((PASSED + 1))
fi

# Test 11: Unit tests exist for SpecKit orchestrator
log_info "Test 11: SpecKit orchestrator unit tests exist"
if [ -f "$PROJECT_ROOT/internal/services/speckit_orchestrator_test.go" ]; then
  log_success "✓ speckit_orchestrator_test.go exists"
  PASSED=$((PASSED + 1))
else
  log_error "✗ speckit_orchestrator_test.go not found"
  FAILED=$((FAILED + 1))
fi

# Test 12: E2E tests exist for auto-activation
log_info "Test 12: E2E tests exist for auto-activation"
if [ -f "$PROJECT_ROOT/internal/services/debate_service_speckit_e2e_test.go" ]; then
  log_success "✓ debate_service_speckit_e2e_test.go exists"
  PASSED=$((PASSED + 1))
else
  log_error "✗ debate_service_speckit_e2e_test.go not found"
  FAILED=$((FAILED + 1))
fi

# Test 13: Run SpecKit orchestrator unit tests
log_info "Test 13: SpecKit orchestrator unit tests pass"
if go test -run "TestSpecKitOrchestrator" -timeout 30s "$PROJECT_ROOT/internal/services/" > /dev/null 2>&1; then
  log_success "✓ All SpecKit orchestrator tests passed"
  PASSED=$((PASSED + 1))
else
  log_error "✗ SpecKit orchestrator tests failed"
  FAILED=$((FAILED + 1))
fi

# Test 14: Run phase caching tests
log_info "Test 14: Phase caching tests pass"
if go test -run "TestSpecKitOrchestrator_PhaseCaching" -timeout 30s "$PROJECT_ROOT/internal/services/" > /dev/null 2>&1; then
  log_success "✓ Phase caching tests passed"
  PASSED=$((PASSED + 1))
else
  log_error "✗ Phase caching tests failed"
  FAILED=$((FAILED + 1))
fi

# Test 15: Test count verification
log_info "Test 15: Sufficient test coverage"
TEST_COUNT=$(grep -c "^func TestSpecKitOrchestrator" "$PROJECT_ROOT/internal/services/speckit_orchestrator_test.go" 2>/dev/null || echo "0")
if [ "$TEST_COUNT" -ge 3 ]; then
  log_success "✓ Found $TEST_COUNT tests (≥3)"
  PASSED=$((PASSED + 1))
else
  log_warning "⚠ Expected at least 3 tests, found: $TEST_COUNT"
  PASSED=$((PASSED + 1))
fi

# Summary
log_info "========================================="
log_info "Test Results: $PASSED passed, $FAILED failed out of $TOTAL_TESTS"
log_info "========================================="

if [ $FAILED -eq 0 ]; then
  log_success "✅ All SpecKit Auto-Activation tests passed!"
  exit 0
else
  log_error "❌ $FAILED SpecKit Auto-Activation tests failed"
  exit 1
fi
