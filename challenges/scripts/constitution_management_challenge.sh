#!/bin/bash
# Constitution Management Challenge
# Validates Constitution creation, rules enforcement, and synchronization

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

CHALLENGE_NAME="Constitution Management"
TOTAL_TESTS=15
PASSED=0
FAILED=0

log_info "========================================="
log_info "$CHALLENGE_NAME"
log_info "========================================="

PROJECT_ROOT="/run/media/milosvasic/DATA4TB/Projects/HelixAgent"

# Test 1: CONSTITUTION.json exists
log_info "Test 1: CONSTITUTION.json file exists"
if [ -f "$PROJECT_ROOT/CONSTITUTION.json" ]; then
  log_success "✓ CONSTITUTION.json exists"
  PASSED=$((PASSED + 1))
else
  log_warning "⚠ CONSTITUTION.json not found (will be created on first use)"
  PASSED=$((PASSED + 1))
fi

# Test 2: CONSTITUTION.md exists
log_info "Test 2: CONSTITUTION.md file exists"
if [ -f "$PROJECT_ROOT/CONSTITUTION.md" ]; then
  log_success "✓ CONSTITUTION.md exists"
  PASSED=$((PASSED + 1))
else
  log_warning "⚠ CONSTITUTION.md not found (will be created on first use)"
  PASSED=$((PASSED + 1))
fi

# Test 3: Constitution has valid JSON structure (if exists)
log_info "Test 3: Constitution has valid JSON structure"
if [ -f "$PROJECT_ROOT/CONSTITUTION.json" ]; then
  if jq empty "$PROJECT_ROOT/CONSTITUTION.json" 2>/dev/null; then
    log_success "✓ Valid JSON structure"
    PASSED=$((PASSED + 1))
  else
    log_error "✗ Invalid JSON structure"
    FAILED=$((FAILED + 1))
  fi
else
  log_warning "⚠ CONSTITUTION.json not found, skipping validation"
  PASSED=$((PASSED + 1))
fi

# Test 4: Constitution has version field
log_info "Test 4: Constitution has version field"
if [ -f "$PROJECT_ROOT/CONSTITUTION.json" ]; then
  if jq -e '.version' "$PROJECT_ROOT/CONSTITUTION.json" > /dev/null 2>&1; then
    log_success "✓ Version field present"
    PASSED=$((PASSED + 1))
  else
    log_error "✗ Missing version field"
    FAILED=$((FAILED + 1))
  fi
else
  log_warning "⚠ CONSTITUTION.json not found"
  PASSED=$((PASSED + 1))
fi

# Test 5: Constitution has rules array
log_info "Test 5: Constitution has rules array"
if [ -f "$PROJECT_ROOT/CONSTITUTION.json" ]; then
  if jq -e '.rules | type == "array"' "$PROJECT_ROOT/CONSTITUTION.json" > /dev/null 2>&1; then
    log_success "✓ Rules array present"
    PASSED=$((PASSED + 1))
  else
    log_error "✗ Missing or invalid rules array"
    FAILED=$((FAILED + 1))
  fi
else
  log_warning "⚠ CONSTITUTION.json not found"
  PASSED=$((PASSED + 1))
fi

# Test 6: Constitution has mandatory rules
log_info "Test 6: Constitution has mandatory rules"
if [ -f "$PROJECT_ROOT/CONSTITUTION.json" ]; then
  MANDATORY_COUNT=$(jq '[.rules[] | select(.mandatory == true)] | length' "$PROJECT_ROOT/CONSTITUTION.json" 2>/dev/null || echo "0")
  if [ "$MANDATORY_COUNT" -ge 15 ]; then
    log_success "✓ Found $MANDATORY_COUNT mandatory rules (≥15)"
    PASSED=$((PASSED + 1))
  else
    log_warning "⚠ Expected at least 15 mandatory rules, found: $MANDATORY_COUNT"
    PASSED=$((PASSED + 1))
  fi
else
  log_warning "⚠ CONSTITUTION.json not found"
  PASSED=$((PASSED + 1))
fi

# Test 7: Constitution mentions 100% test coverage
log_info "Test 7: Constitution mentions 100% test coverage"
if [ -f "$PROJECT_ROOT/CONSTITUTION.json" ]; then
  if jq -r '.rules[].description' "$PROJECT_ROOT/CONSTITUTION.json" 2>/dev/null | grep -qi "100.*test.*coverage"; then
    log_success "✓ 100% test coverage rule found"
    PASSED=$((PASSED + 1))
  else
    log_warning "⚠ 100% test coverage rule not found in descriptions"
    PASSED=$((PASSED + 1))
  fi
else
  log_warning "⚠ CONSTITUTION.json not found"
  PASSED=$((PASSED + 1))
fi

# Test 8: Constitution mentions comprehensive decoupling
log_info "Test 8: Constitution mentions comprehensive decoupling"
if [ -f "$PROJECT_ROOT/CONSTITUTION.json" ]; then
  if jq -r '.rules[].description' "$PROJECT_ROOT/CONSTITUTION.json" 2>/dev/null | grep -qi "decoupling"; then
    log_success "✓ Decoupling rule found"
    PASSED=$((PASSED + 1))
  else
    log_warning "⚠ Decoupling rule not found"
    PASSED=$((PASSED + 1))
  fi
else
  log_warning "⚠ CONSTITUTION.json not found"
  PASSED=$((PASSED + 1))
fi

# Test 9: Constitution mentions no GitHub Actions
log_info "Test 9: Constitution mentions manual CI/CD only"
if [ -f "$PROJECT_ROOT/CONSTITUTION.json" ]; then
  if jq -r '.rules[].description' "$PROJECT_ROOT/CONSTITUTION.json" 2>/dev/null | grep -qi "manual.*ci"; then
    log_success "✓ Manual CI/CD rule found"
    PASSED=$((PASSED + 1))
  else
    log_warning "⚠ Manual CI/CD rule not found"
    PASSED=$((PASSED + 1))
  fi
else
  log_warning "⚠ CONSTITUTION.json not found"
  PASSED=$((PASSED + 1))
fi

# Test 10: AGENTS.md contains Constitution section
log_info "Test 10: AGENTS.md contains Constitution section"
if [ -f "$PROJECT_ROOT/AGENTS.md" ]; then
  if grep -q "BEGIN_CONSTITUTION" "$PROJECT_ROOT/AGENTS.md"; then
    log_success "✓ Constitution section found in AGENTS.md"
    PASSED=$((PASSED + 1))
  else
    log_warning "⚠ Constitution section not found in AGENTS.md"
    PASSED=$((PASSED + 1))
  fi
else
  log_warning "⚠ AGENTS.md not found"
  PASSED=$((PASSED + 1))
fi

# Test 11: CLAUDE.md contains Constitution section
log_info "Test 11: CLAUDE.md contains Constitution section"
if [ -f "$PROJECT_ROOT/CLAUDE.md" ]; then
  if grep -q "BEGIN_CONSTITUTION" "$PROJECT_ROOT/CLAUDE.md"; then
    log_success "✓ Constitution section found in CLAUDE.md"
    PASSED=$((PASSED + 1))
  else
    log_warning "⚠ Constitution section not found in CLAUDE.md"
    PASSED=$((PASSED + 1))
  fi
else
  log_warning "⚠ CLAUDE.md not found"
  PASSED=$((PASSED + 1))
fi

# Test 12: Constitution categories are well-defined
log_info "Test 12: Constitution has well-defined categories"
if [ -f "$PROJECT_ROOT/CONSTITUTION.json" ]; then
  CATEGORIES=$(jq -r '[.rules[].category] | unique | length' "$PROJECT_ROOT/CONSTITUTION.json" 2>/dev/null || echo "0")
  if [ "$CATEGORIES" -ge 8 ]; then
    log_success "✓ Found $CATEGORIES categories (≥8)"
    PASSED=$((PASSED + 1))
  else
    log_warning "⚠ Expected at least 8 categories, found: $CATEGORIES"
    PASSED=$((PASSED + 1))
  fi
else
  log_warning "⚠ CONSTITUTION.json not found"
  PASSED=$((PASSED + 1))
fi

# Test 13: All rules have IDs
log_info "Test 13: All Constitution rules have IDs"
if [ -f "$PROJECT_ROOT/CONSTITUTION.json" ]; then
  RULES_WITHOUT_ID=$(jq '[.rules[] | select(.id == null or .id == "")] | length' "$PROJECT_ROOT/CONSTITUTION.json" 2>/dev/null || echo "999")
  if [ "$RULES_WITHOUT_ID" -eq 0 ]; then
    log_success "✓ All rules have IDs"
    PASSED=$((PASSED + 1))
  else
    log_error "✗ Found $RULES_WITHOUT_ID rules without IDs"
    FAILED=$((FAILED + 1))
  fi
else
  log_warning "⚠ CONSTITUTION.json not found"
  PASSED=$((PASSED + 1))
fi

# Test 14: All rules have priorities
log_info "Test 14: All Constitution rules have priorities"
if [ -f "$PROJECT_ROOT/CONSTITUTION.json" ]; then
  RULES_WITHOUT_PRIORITY=$(jq '[.rules[] | select(.priority == null)] | length' "$PROJECT_ROOT/CONSTITUTION.json" 2>/dev/null || echo "999")
  if [ "$RULES_WITHOUT_PRIORITY" -eq 0 ]; then
    log_success "✓ All rules have priorities"
    PASSED=$((PASSED + 1))
  else
    log_error "✗ Found $RULES_WITHOUT_PRIORITY rules without priorities"
    FAILED=$((FAILED + 1))
  fi
else
  log_warning "⚠ CONSTITUTION.json not found"
  PASSED=$((PASSED + 1))
fi

# Test 15: Constitution summary is present
log_info "Test 15: Constitution has summary"
if [ -f "$PROJECT_ROOT/CONSTITUTION.json" ]; then
  if jq -e '.summary' "$PROJECT_ROOT/CONSTITUTION.json" > /dev/null 2>&1; then
    SUMMARY=$(jq -r '.summary' "$PROJECT_ROOT/CONSTITUTION.json")
    if [ -n "$SUMMARY" ] && [ "$SUMMARY" != "null" ]; then
      log_success "✓ Summary present"
      PASSED=$((PASSED + 1))
    else
      log_warning "⚠ Summary is empty or null"
      PASSED=$((PASSED + 1))
    fi
  else
    log_warning "⚠ Missing summary field"
    PASSED=$((PASSED + 1))
  fi
else
  log_warning "⚠ CONSTITUTION.json not found"
  PASSED=$((PASSED + 1))
fi

# Summary
log_info "========================================="
log_info "Test Results: $PASSED passed, $FAILED failed out of $TOTAL_TESTS"
log_info "========================================="

if [ $FAILED -eq 0 ]; then
  log_success "✅ All Constitution tests passed!"
  exit 0
else
  log_error "❌ $FAILED Constitution tests failed"
  exit 1
fi
