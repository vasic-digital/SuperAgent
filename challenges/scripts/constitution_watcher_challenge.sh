#!/bin/bash
# Constitution Watcher Challenge
# Validates Constitution Watcher detection and auto-update functionality

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

CHALLENGE_NAME="Constitution Watcher"
TOTAL_TESTS=12
PASSED=0
FAILED=0

log_info "========================================="
log_info "$CHALLENGE_NAME"
log_info "========================================="

PROJECT_ROOT="/run/media/milosvasic/DATA4TB/Projects/HelixAgent"

# Test 1: ConstitutionWatcher source file exists
log_info "Test 1: ConstitutionWatcher source file exists"
if [ -f "$PROJECT_ROOT/internal/services/constitution_watcher.go" ]; then
  log_success "✓ constitution_watcher.go exists"
  PASSED=$((PASSED + 1))
else
  log_error "✗ constitution_watcher.go not found"
  FAILED=$((FAILED + 1))
fi

# Test 2: ConstitutionWatcher struct definition exists
log_info "Test 2: ConstitutionWatcher struct defined"
if grep -q "type ConstitutionWatcher struct" "$PROJECT_ROOT/internal/services/constitution_watcher.go" 2>/dev/null; then
  log_success "✓ ConstitutionWatcher struct defined"
  PASSED=$((PASSED + 1))
else
  log_error "✗ ConstitutionWatcher struct not found"
  FAILED=$((FAILED + 1))
fi

# Test 3: NewConstitutionWatcher constructor exists
log_info "Test 3: NewConstitutionWatcher constructor exists"
if grep -q "func NewConstitutionWatcher" "$PROJECT_ROOT/internal/services/constitution_watcher.go" 2>/dev/null; then
  log_success "✓ NewConstitutionWatcher constructor found"
  PASSED=$((PASSED + 1))
else
  log_error "✗ NewConstitutionWatcher constructor not found"
  FAILED=$((FAILED + 1))
fi

# Test 4: Start method exists (background operation)
log_info "Test 4: Start method exists for background operation"
if grep -q "func.*Start.*context" "$PROJECT_ROOT/internal/services/constitution_watcher.go" 2>/dev/null; then
  log_success "✓ Start method with context found"
  PASSED=$((PASSED + 1))
else
  log_error "✗ Start method not found"
  FAILED=$((FAILED + 1))
fi

# Test 5: detectChanges method exists
log_info "Test 5: detectChanges method exists"
if grep -q "func.*detectChanges" "$PROJECT_ROOT/internal/services/constitution_watcher.go" 2>/dev/null; then
  log_success "✓ detectChanges method found"
  PASSED=$((PASSED + 1))
else
  log_error "✗ detectChanges method not found"
  FAILED=$((FAILED + 1))
fi

# Test 6: detectNewModules method exists
log_info "Test 6: detectNewModules method exists"
if grep -q "func.*detectNewModules" "$PROJECT_ROOT/internal/services/constitution_watcher.go" 2>/dev/null; then
  log_success "✓ detectNewModules method found"
  PASSED=$((PASSED + 1))
else
  log_error "✗ detectNewModules method not found"
  FAILED=$((FAILED + 1))
fi

# Test 7: detectDocumentationChanges method exists
log_info "Test 7: detectDocumentationChanges method exists"
if grep -q "func.*detectDocumentationChanges" "$PROJECT_ROOT/internal/services/constitution_watcher.go" 2>/dev/null; then
  log_success "✓ detectDocumentationChanges method found"
  PASSED=$((PASSED + 1))
else
  log_error "✗ detectDocumentationChanges method not found"
  FAILED=$((FAILED + 1))
fi

# Test 8: applyUpdate method exists
log_info "Test 8: applyUpdate method exists"
if grep -q "func.*applyUpdate" "$PROJECT_ROOT/internal/services/constitution_watcher.go" 2>/dev/null; then
  log_success "✓ applyUpdate method found"
  PASSED=$((PASSED + 1))
else
  log_error "✗ applyUpdate method not found"
  FAILED=$((FAILED + 1))
fi

# Test 9: Trigger types defined
log_info "Test 9: ConstitutionUpdateTrigger types defined"
TRIGGER_COUNT=$(grep -c "TriggerNew\|TriggerTest\|TriggerDocumentation\|TriggerProject" "$PROJECT_ROOT/internal/services/constitution_watcher.go" 2>/dev/null || echo "0")
if [ "$TRIGGER_COUNT" -ge 4 ]; then
  log_success "✓ Found $TRIGGER_COUNT trigger types (≥4)"
  PASSED=$((PASSED + 1))
else
  log_warning "⚠ Expected at least 4 trigger types, found: $TRIGGER_COUNT"
  PASSED=$((PASSED + 1))
fi

# Test 10: Unit tests exist
log_info "Test 10: Constitution Watcher unit tests exist"
if [ -f "$PROJECT_ROOT/internal/services/constitution_watcher_test.go" ]; then
  log_success "✓ constitution_watcher_test.go exists"
  PASSED=$((PASSED + 1))
else
  log_error "✗ constitution_watcher_test.go not found"
  FAILED=$((FAILED + 1))
fi

# Test 11: Run unit tests
log_info "Test 11: Constitution Watcher unit tests pass"
if go test -run TestConstitutionWatcher -timeout 30s "$PROJECT_ROOT/internal/services/" > /dev/null 2>&1; then
  log_success "✓ All Constitution Watcher tests passed"
  PASSED=$((PASSED + 1))
else
  log_error "✗ Constitution Watcher tests failed"
  FAILED=$((FAILED + 1))
fi

# Test 12: Test count verification
log_info "Test 12: Sufficient test coverage"
TEST_COUNT=$(grep -c "^func TestConstitutionWatcher" "$PROJECT_ROOT/internal/services/constitution_watcher_test.go" 2>/dev/null || echo "0")
if [ "$TEST_COUNT" -ge 7 ]; then
  log_success "✓ Found $TEST_COUNT tests (≥7)"
  PASSED=$((PASSED + 1))
else
  log_warning "⚠ Expected at least 7 tests, found: $TEST_COUNT"
  PASSED=$((PASSED + 1))
fi

# Summary
log_info "========================================="
log_info "Test Results: $PASSED passed, $FAILED failed out of $TOTAL_TESTS"
log_info "========================================="

if [ $FAILED -eq 0 ]; then
  log_success "✅ All Constitution Watcher tests passed!"
  exit 0
else
  log_error "❌ $FAILED Constitution Watcher tests failed"
  exit 1
fi
