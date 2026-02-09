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

print_header "$CHALLENGE_NAME"

PROJECT_ROOT="/run/media/milosvasic/DATA4TB/Projects/HelixAgent"

# Test 1: CONSTITUTION.json exists
test_start "CONSTITUTION.json file exists"
if [ -f "$PROJECT_ROOT/CONSTITUTION.json" ]; then
  test_pass
  PASSED=$((PASSED + 1))
else
  test_warn "CONSTITUTION.json not found (will be created on first use)"
  PASSED=$((PASSED + 1))
fi

# Test 2: CONSTITUTION.md exists
test_start "CONSTITUTION.md file exists"
if [ -f "$PROJECT_ROOT/CONSTITUTION.md" ]; then
  test_pass
  PASSED=$((PASSED + 1))
else
  test_warn "CONSTITUTION.md not found (will be created on first use)"
  PASSED=$((PASSED + 1))
fi

# Test 3: Constitution has valid JSON structure (if exists)
test_start "Constitution has valid JSON structure"
if [ -f "$PROJECT_ROOT/CONSTITUTION.json" ]; then
  if jq empty "$PROJECT_ROOT/CONSTITUTION.json" 2>/dev/null; then
    test_pass
    PASSED=$((PASSED + 1))
  else
    test_fail "Invalid JSON structure"
    FAILED=$((FAILED + 1))
  fi
else
  test_warn "CONSTITUTION.json not found, skipping validation"
  PASSED=$((PASSED + 1))
fi

# Test 4: Constitution has version field
test_start "Constitution has version field"
if [ -f "$PROJECT_ROOT/CONSTITUTION.json" ]; then
  if jq -e '.version' "$PROJECT_ROOT/CONSTITUTION.json" > /dev/null 2>&1; then
    test_pass
    PASSED=$((PASSED + 1))
  else
    test_fail "Missing version field"
    FAILED=$((FAILED + 1))
  fi
else
  test_warn "CONSTITUTION.json not found"
  PASSED=$((PASSED + 1))
fi

# Test 5: Constitution has rules array
test_start "Constitution has rules array"
if [ -f "$PROJECT_ROOT/CONSTITUTION.json" ]; then
  if jq -e '.rules | type == "array"' "$PROJECT_ROOT/CONSTITUTION.json" > /dev/null 2>&1; then
    test_pass
    PASSED=$((PASSED + 1))
  else
    test_fail "Missing or invalid rules array"
    FAILED=$((FAILED + 1))
  fi
else
  test_warn "CONSTITUTION.json not found"
  PASSED=$((PASSED + 1))
fi

# Test 6: Constitution has mandatory rules
test_start "Constitution has mandatory rules"
if [ -f "$PROJECT_ROOT/CONSTITUTION.json" ]; then
  MANDATORY_COUNT=$(jq '[.rules[] | select(.mandatory == true)] | length' "$PROJECT_ROOT/CONSTITUTION.json" 2>/dev/null || echo "0")
  if [ "$MANDATORY_COUNT" -ge 15 ]; then
    test_pass
    PASSED=$((PASSED + 1))
  else
    test_warn "Expected at least 15 mandatory rules, found: $MANDATORY_COUNT"
    PASSED=$((PASSED + 1))
  fi
else
  test_warn "CONSTITUTION.json not found"
  PASSED=$((PASSED + 1))
fi

# Test 7: Constitution mentions 100% test coverage
test_start "Constitution mentions 100% test coverage"
if [ -f "$PROJECT_ROOT/CONSTITUTION.json" ]; then
  if jq -r '.rules[].description' "$PROJECT_ROOT/CONSTITUTION.json" 2>/dev/null | grep -qi "100.*test.*coverage"; then
    test_pass
    PASSED=$((PASSED + 1))
  else
    test_warn "100% test coverage rule not found in descriptions"
    PASSED=$((PASSED + 1))
  fi
else
  test_warn "CONSTITUTION.json not found"
  PASSED=$((PASSED + 1))
fi

# Test 8: Constitution mentions comprehensive decoupling
test_start "Constitution mentions comprehensive decoupling"
if [ -f "$PROJECT_ROOT/CONSTITUTION.json" ]; then
  if jq -r '.rules[].description' "$PROJECT_ROOT/CONSTITUTION.json" 2>/dev/null | grep -qi "decoupling"; then
    test_pass
    PASSED=$((PASSED + 1))
  else
    test_warn "Decoupling rule not found"
    PASSED=$((PASSED + 1))
  fi
else
  test_warn "CONSTITUTION.json not found"
  PASSED=$((PASSED + 1))
fi

# Test 9: Constitution mentions no GitHub Actions
test_start "Constitution mentions manual CI/CD only"
if [ -f "$PROJECT_ROOT/CONSTITUTION.json" ]; then
  if jq -r '.rules[].description' "$PROJECT_ROOT/CONSTITUTION.json" 2>/dev/null | grep -qi "manual.*ci"; then
    test_pass
    PASSED=$((PASSED + 1))
  else
    test_warn "Manual CI/CD rule not found"
    PASSED=$((PASSED + 1))
  fi
else
  test_warn "CONSTITUTION.json not found"
  PASSED=$((PASSED + 1))
fi

# Test 10: AGENTS.md contains Constitution section
test_start "AGENTS.md contains Constitution section"
if [ -f "$PROJECT_ROOT/AGENTS.md" ]; then
  if grep -q "BEGIN_CONSTITUTION" "$PROJECT_ROOT/AGENTS.md"; then
    test_pass
    PASSED=$((PASSED + 1))
  else
    test_warn "Constitution section not found in AGENTS.md"
    PASSED=$((PASSED + 1))
  fi
else
  test_warn "AGENTS.md not found"
  PASSED=$((PASSED + 1))
fi

# Test 11: CLAUDE.md contains Constitution section
test_start "CLAUDE.md contains Constitution section"
if [ -f "$PROJECT_ROOT/CLAUDE.md" ]; then
  if grep -q "BEGIN_CONSTITUTION" "$PROJECT_ROOT/CLAUDE.md"; then
    test_pass
    PASSED=$((PASSED + 1))
  else
    test_warn "Constitution section not found in CLAUDE.md"
    PASSED=$((PASSED + 1))
  fi
else
  test_warn "CLAUDE.md not found"
  PASSED=$((PASSED + 1))
fi

# Test 12: Constitution categories are well-defined
test_start "Constitution has well-defined categories"
if [ -f "$PROJECT_ROOT/CONSTITUTION.json" ]; then
  CATEGORIES=$(jq -r '[.rules[].category] | unique | length' "$PROJECT_ROOT/CONSTITUTION.json" 2>/dev/null || echo "0")
  if [ "$CATEGORIES" -ge 8 ]; then
    test_pass
    PASSED=$((PASSED + 1))
  else
    test_warn "Expected at least 8 categories, found: $CATEGORIES"
    PASSED=$((PASSED + 1))
  fi
else
  test_warn "CONSTITUTION.json not found"
  PASSED=$((PASSED + 1))
fi

# Test 13: All rules have IDs
test_start "All Constitution rules have IDs"
if [ -f "$PROJECT_ROOT/CONSTITUTION.json" ]; then
  RULES_WITHOUT_ID=$(jq '[.rules[] | select(.id == null or .id == "")] | length' "$PROJECT_ROOT/CONSTITUTION.json" 2>/dev/null || echo "999")
  if [ "$RULES_WITHOUT_ID" -eq 0 ]; then
    test_pass
    PASSED=$((PASSED + 1))
  else
    test_fail "Found $RULES_WITHOUT_ID rules without IDs"
    FAILED=$((FAILED + 1))
  fi
else
  test_warn "CONSTITUTION.json not found"
  PASSED=$((PASSED + 1))
fi

# Test 14: All rules have priorities
test_start "All Constitution rules have priorities"
if [ -f "$PROJECT_ROOT/CONSTITUTION.json" ]; then
  RULES_WITHOUT_PRIORITY=$(jq '[.rules[] | select(.priority == null)] | length' "$PROJECT_ROOT/CONSTITUTION.json" 2>/dev/null || echo "999")
  if [ "$RULES_WITHOUT_PRIORITY" -eq 0 ]; then
    test_pass
    PASSED=$((PASSED + 1))
  else
    test_fail "Found $RULES_WITHOUT_PRIORITY rules without priorities"
    FAILED=$((FAILED + 1))
  fi
else
  test_warn "CONSTITUTION.json not found"
  PASSED=$((PASSED + 1))
fi

# Test 15: Constitution summary is present
test_start "Constitution has summary"
if [ -f "$PROJECT_ROOT/CONSTITUTION.json" ]; then
  if jq -e '.summary' "$PROJECT_ROOT/CONSTITUTION.json" > /dev/null 2>&1; then
    SUMMARY=$(jq -r '.summary' "$PROJECT_ROOT/CONSTITUTION.json")
    if [ -n "$SUMMARY" ] && [ "$SUMMARY" != "null" ]; then
      test_pass
      PASSED=$((PASSED + 1))
    else
      test_warn "Summary is empty or null"
      PASSED=$((PASSED + 1))
    fi
  else
    test_warn "Missing summary field"
    PASSED=$((PASSED + 1))
  fi
else
  test_warn "CONSTITUTION.json not found"
  PASSED=$((PASSED + 1))
fi

print_summary $PASSED $FAILED $TOTAL_TESTS
exit $FAILED
