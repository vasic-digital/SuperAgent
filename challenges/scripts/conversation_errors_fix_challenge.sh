#!/bin/bash
#
# Conversation Errors Fix Challenge
# Validates fixes for issues found in conversation errors:
# 1. NVIDIA duplicate prefix bug (nvidia/nvidia/model)
# 2. Claude OAuth CLI empty stderr bug
#
# Usage: ./conversation_errors_fix_challenge.sh
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_SKIPPED=0

# Helper functions
pass() {
    echo -e "${GREEN}✓ PASS${NC}: $1"
    ((TESTS_PASSED++))
}

fail() {
    echo -e "${RED}✗ FAIL${NC}: $1"
    echo "  Reason: $2"
    ((TESTS_FAILED++))
}

skip() {
    echo -e "${YELLOW}○ SKIP${NC}: $1"
    echo "  Reason: $2"
    ((TESTS_SKIPPED++))
}

section() {
    echo ""
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
}

# ============================================================================
# Test 1: Model Reference Formatting (formatModelRef)
# ============================================================================

test_format_model_ref() {
    section "Test 1: Model Reference Formatting (formatModelRef)"
    
    cd "$PROJECT_ROOT"
    
    # Test 1.1: Standard case - no prefix duplication
    echo -e "\n${YELLOW}Test 1.1: Standard case - no prefix duplication${NC}"
    if go test -v -run "TestFormatModelRef/Standard_case" ./internal/handlers/... 2>&1 | grep -q "PASS"; then
        pass "Standard model format (anthropic/claude-3-opus)"
    else
        fail "Standard model format" "Test did not pass"
    fi
    
    # Test 1.2: NVIDIA model with org prefix
    echo -e "\n${YELLOW}Test 1.2: NVIDIA model with org prefix in model ID${NC}"
    if go test -v -run "TestFormatModelRef/NVIDIA_model_with_org_prefix" ./internal/handlers/... 2>&1 | grep -q "PASS"; then
        pass "NVIDIA model format (nvidia/llama-3.1-nemotron-70b-instruct)"
    else
        fail "NVIDIA model format" "Test did not pass"
    fi
    
    # Test 1.3: Meta model with org prefix
    echo -e "\n${YELLOW}Test 1.3: Meta model with org prefix${NC}"
    if go test -v -run "TestFormatModelRef/Meta_model_with_org_prefix" ./internal/handlers/... 2>&1 | grep -q "PASS"; then
        pass "Meta model format (nvidia/meta/llama-3.1-405b-instruct)"
    else
        fail "Meta model format" "Test did not pass"
    fi
    
    # Test 1.4: HuggingFace model with org prefix
    echo -e "\n${YELLOW}Test 1.4: HuggingFace model with org prefix${NC}"
    if go test -v -run "TestFormatModelRef/HuggingFace_model_with_org_prefix" ./internal/handlers/... 2>&1 | grep -q "PASS"; then
        pass "HuggingFace model format (huggingface/meta-llama/Llama-3.3-70B-Instruct)"
    else
        fail "HuggingFace model format" "Test did not pass"
    fi
    
    # Test 1.5: Case insensitive prefix matching
    echo -e "\n${YELLOW}Test 1.5: Case insensitive prefix matching${NC}"
    if go test -v -run "TestFormatModelRef/Case_insensitive_prefix_matching" ./internal/handlers/... 2>&1 | grep -q "PASS"; then
        pass "Case insensitive prefix matching"
    else
        fail "Case insensitive prefix matching" "Test did not pass"
    fi
}

# ============================================================================
# Test 2: Fallback Triggered Markdown (No Double Prefix)
# ============================================================================

test_fallback_triggered_markdown() {
    section "Test 2: Fallback Triggered Markdown (No Double Prefix)"
    
    cd "$PROJECT_ROOT"
    
    # Test 2.1: NVIDIA model does not double prefix
    echo -e "\n${YELLOW}Test 2.1: NVIDIA model with org prefix does not double prefix${NC}"
    if go test -v -run "TestFormatFallbackTriggeredMarkdown_NoDoublePrefix/NVIDIA_model_with_org_prefix_does_not_double_prefix" ./internal/handlers/... 2>&1 | grep -q "PASS"; then
        pass "NVIDIA fallback display does not double prefix"
    else
        fail "NVIDIA fallback display" "Test did not pass"
    fi
    
    # Test 2.2: Standard model displays correctly
    echo -e "\n${YELLOW}Test 2.2: Standard model displays correctly${NC}"
    if go test -v -run "TestFormatFallbackTriggeredMarkdown_NoDoublePrefix/Standard_model_displays_correctly" ./internal/handlers/... 2>&1 | grep -q "PASS"; then
        pass "Standard model fallback display"
    else
        fail "Standard model fallback display" "Test did not pass"
    fi
}

# ============================================================================
# Test 3: Claude CLI Session Detection
# ============================================================================

test_claude_cli_session_detection() {
    section "Test 3: Claude CLI Session Detection"
    
    cd "$PROJECT_ROOT"
    
    # Test 3.1: Session detection function works
    echo -e "\n${YELLOW}Test 3.1: Session detection function works${NC}"
    if go test -v -run "TestIsInsideClaudeCodeSession" ./internal/llm/providers/claude/... 2>&1 | grep -q "PASS"; then
        pass "IsInsideClaudeCodeSession function works"
    else
        fail "Session detection function" "Test did not pass"
    fi
    
    # Test 3.2: Complete blocks inside session
    echo -e "\n${YELLOW}Test 3.2: Complete blocks inside session${NC}"
    if go test -v -run "TestClaudeCLIProvider_Complete_InsideSession" ./internal/llm/providers/claude/... 2>&1 | grep -q "PASS"; then
        pass "Complete blocks when inside Claude Code session"
    else
        fail "Complete session blocking" "Test did not pass"
    fi
    
    # Test 3.3: CompleteStream blocks inside session
    echo -e "\n${YELLOW}Test 3.3: CompleteStream blocks inside session${NC}"
    if go test -v -run "TestClaudeCLIProvider_CompleteStream_InsideSession" ./internal/llm/providers/claude/... 2>&1 | grep -q "PASS"; then
        pass "CompleteStream blocks when inside Claude Code session"
    else
        fail "CompleteStream session blocking" "Test did not pass"
    fi
}

# ============================================================================
# Test 4: All Fallback Formatting Tests
# ============================================================================

test_all_fallback_formatting() {
    section "Test 4: All Fallback Formatting Tests"
    
    cd "$PROJECT_ROOT"
    
    # Test 4.1: All fallback formatting tests pass
    echo -e "\n${YELLOW}Test 4.1: All fallback formatting tests pass${NC}"
    if go test -v -run "TestFormatFallback" ./internal/handlers/... 2>&1 | grep -q "PASS"; then
        pass "All fallback formatting tests pass"
    else
        fail "Fallback formatting tests" "Some tests failed"
    fi
}

# ============================================================================
# Test 5: Code Compiles
# ============================================================================

test_code_compiles() {
    section "Test 5: Code Compiles"
    
    cd "$PROJECT_ROOT"
    
    # Test 5.1: Handlers package compiles
    echo -e "\n${YELLOW}Test 5.1: Handlers package compiles${NC}"
    if go build ./internal/handlers/... 2>&1; then
        pass "Handlers package compiles"
    else
        fail "Handlers package compilation" "Build failed"
    fi
    
    # Test 5.2: Claude provider package compiles
    echo -e "\n${YELLOW}Test 5.2: Claude provider package compiles${NC}"
    if go build ./internal/llm/providers/claude/... 2>&1; then
        pass "Claude provider package compiles"
    else
        fail "Claude provider package compilation" "Build failed"
    fi
    
    # Test 5.3: Formatters package compiles
    echo -e "\n${YELLOW}Test 5.3: Formatters package compiles${NC}"
    if go build ./internal/formatters/... 2>&1; then
        pass "Formatters package compiles"
    else
        fail "Formatters package compilation" "Build failed"
    fi
}

# ============================================================================
# Test 6: No Regressions in Existing Tests
# ============================================================================

test_no_regressions() {
    section "Test 6: No Regressions in Existing Tests"
    
    cd "$PROJECT_ROOT"
    
    # Test 6.1: All handler tests pass
    echo -e "\n${YELLOW}Test 6.1: All handler tests pass${NC}"
    if go test -short ./internal/handlers/... 2>&1 | grep -q "ok"; then
        pass "Handler tests pass"
    else
        fail "Handler tests" "Some tests failed"
    fi
    
    # Test 6.2: Claude provider tests pass
    echo -e "\n${YELLOW}Test 6.2: Claude provider tests pass${NC}"
    if go test -short ./internal/llm/providers/claude/... 2>&1 | grep -q "ok"; then
        pass "Claude provider tests pass"
    else
        fail "Claude provider tests" "Some tests failed"
    fi
}

# ============================================================================
# Main
# ============================================================================

main() {
    echo -e "${BLUE}"
    echo "╔═══════════════════════════════════════════════════════════════╗"
    echo "║       Conversation Errors Fix Challenge                       ║"
    echo "║       Validates fixes for NVIDIA prefix and Claude CLI bugs   ║"
    echo "╚═══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    
    test_format_model_ref
    test_fallback_triggered_markdown
    test_claude_cli_session_detection
    test_all_fallback_formatting
    test_code_compiles
    test_no_regressions
    
    # Summary
    section "Summary"
    echo -e "${GREEN}Passed:${NC} $TESTS_PASSED"
    echo -e "${RED}Failed:${NC} $TESTS_FAILED"
    echo -e "${YELLOW}Skipped:${NC} $TESTS_SKIPPED"
    
    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "\n${GREEN}All tests passed!${NC}"
        exit 0
    else
        echo -e "\n${RED}Some tests failed!${NC}"
        exit 1
    fi
}

main "$@"
