#!/bin/bash

# Challenge: Comprehensive 11-Role Debate System Verification
# This challenge verifies that the 11-role comprehensive debate system works correctly

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

TESTS_PASSED=0
TESTS_FAILED=0

echo -e "${BLUE}============================================${NC}"
echo -e "${BLUE}11-Role Comprehensive Debate System Challenge${NC}"
echo -e "${BLUE}============================================${NC}"
echo ""

# Test 1: Verify all 11 roles are defined
test_all_roles_defined() {
    echo -e "${YELLOW}Test 1: Verify all 11 roles are defined${NC}"
    
    local roles=(
        "RoleArchitect"
        "RoleGenerator"
        "RoleCritic"
        "RoleRefactoring"
        "RoleTester"
        "RoleValidator"
        "RoleSecurity"
        "RolePerformance"
        "RoleModerator"
        "RoleRedTeam"
        "RoleBlueTeam"
    )
    
    local roles_file="$PROJECT_ROOT/internal/debate/comprehensive/types.go"
    local all_found=true
    
    for role in "${roles[@]}"; do
        if grep -q "$role" "$roles_file"; then
            echo -e "  ${GREEN}✓${NC} Found role: $role"
        else
            echo -e "  ${RED}✗${NC} Missing role: $role"
            all_found=false
        fi
    done
    
    if [ "$all_found" = true ]; then
        echo -e "${GREEN}PASS: All 11 roles are defined${NC}"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}FAIL: Not all roles are defined${NC}"
        ((TESTS_FAILED++))
    fi
}

# Test 2: Verify all role prompts exist
test_all_role_prompts_exist() {
    echo -e "${YELLOW}Test 2: Verify all role prompts exist${NC}"
    
    local prompts=(
        "Architect"
        "Generator"
        "Critic"
        "Refactoring"
        "Tester"
        "Validator"
        "Security"
        "Performance"
        "Moderator"
        "RedTeam"
        "BlueTeam"
    )
    
    local utils_file="$PROJECT_ROOT/internal/debate/comprehensive/utils.go"
    local all_found=true
    
    for prompt in "${prompts[@]}"; do
        if grep -q "func (rp RolePrompts) $prompt()" "$utils_file"; then
            echo -e "  ${GREEN}✓${NC} Found prompt: $prompt"
        else
            echo -e "  ${RED}✗${NC} Missing prompt: $prompt"
            all_found=false
        fi
    done
    
    if [ "$all_found" = true ]; then
        echo -e "${GREEN}PASS: All 11 role prompts exist${NC}"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}FAIL: Not all role prompts exist${NC}"
        ((TESTS_FAILED++))
    fi
}

# Test 3: Verify comprehensive role prompt function is dynamic
test_dynamic_role_prompts() {
    echo -e "${YELLOW}Test 3: Verify comprehensive role prompt function uses dynamic RolePrompts${NC}"
    
    local handler_file="$PROJECT_ROOT/internal/handlers/openai_compatible.go"
    
    if grep -q "rp := comprehensive.RolePrompts{}" "$handler_file" && \
       grep -q "return rp.Architect()" "$handler_file" && \
       grep -q "return rp.RedTeam()" "$handler_file" && \
       grep -q "return rp.BlueTeam()" "$handler_file"; then
        echo -e "${GREEN}PASS: getComprehensiveRolePrompt uses dynamic RolePrompts${NC}"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}FAIL: getComprehensiveRolePrompt is not using dynamic RolePrompts${NC}"
        ((TESTS_FAILED++))
    fi
}

# Test 4: Verify footer shows 11 perspectives
test_footer_11_perspectives() {
    echo -e "${YELLOW}Test 4: Verify footer shows 11 AI perspectives${NC}"
    
    local handler_file="$PROJECT_ROOT/internal/handlers/openai_compatible.go"
    
    if grep -q "11 AI perspectives" "$handler_file"; then
        echo -e "${GREEN}PASS: Footer shows 11 AI perspectives${NC}"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}FAIL: Footer does not show 11 AI perspectives${NC}"
        ((TESTS_FAILED++))
    fi
}

# Test 5: Verify test updated for 11 perspectives
test_test_updated() {
    echo -e "${YELLOW}Test 5: Verify test updated for 11 perspectives${NC}"
    
    local test_file="$PROJECT_ROOT/internal/handlers/openai_compatible_test.go"
    
    if grep -q "11 AI perspectives" "$test_file"; then
        echo -e "${GREEN}PASS: Test updated for 11 AI perspectives${NC}"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}FAIL: Test not updated for 11 AI perspectives${NC}"
        ((TESTS_FAILED++))
    fi
}

# Test 6: Verify role prompt tests exist
test_role_prompt_tests() {
    echo -e "${YELLOW}Test 6: Verify role prompt tests exist${NC}"
    
    local test_file="$PROJECT_ROOT/internal/debate/comprehensive/utils_test.go"
    
    if grep -q "TestRolePrompts_RedTeam" "$test_file" && \
       grep -q "TestRolePrompts_BlueTeam" "$test_file"; then
        echo -e "${GREEN}PASS: Tests for all role prompts exist${NC}"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}FAIL: Missing tests for RedTeam or BlueTeam${NC}"
        ((TESTS_FAILED++))
    fi
}

# Test 7: Verify position mapping includes all 11 roles
test_position_mapping() {
    echo -e "${YELLOW}Test 7: Verify position mapping includes all 11 roles${NC}"
    
    local handler_file="$PROJECT_ROOT/internal/handlers/openai_compatible.go"
    local roles=(
        "Architect"
        "Moderator"
        "Generator"
        "Blue Team"
        "Critic"
        "Tester"
        "Validator"
        "Security"
        "Performance"
        "Red Team"
        "Refactoring"
    )
    
    local all_found=true
    for role in "${roles[@]}"; do
        if grep -q "\"$role\"" "$handler_file"; then
            echo -e "  ${GREEN}✓${NC} Found position mapping: $role"
        else
            echo -e "  ${RED}✗${NC} Missing position mapping: $role"
            all_found=false
        fi
    done
    
    if [ "$all_found" = true ]; then
        echo -e "${GREEN}PASS: All 11 roles are mapped in positions${NC}"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}FAIL: Not all roles are mapped in positions${NC}"
        ((TESTS_FAILED++))
    fi
}

# Test 8: Verify LLMsVerifier timeout retry logic
test_llmsverifier_timeout_retry() {
    echo -e "${YELLOW}Test 8: Verify LLMsVerifier timeout retry logic${NC}"
    
    local verification_file="$PROJECT_ROOT/LLMsVerifier/llm-verifier/verification/code_verification.go"
    
    if grep -q "maxRetries = 3" "$verification_file" && \
       grep -q "Consistent timeout" "$verification_file"; then
        echo -e "${GREEN}PASS: LLMsVerifier has timeout retry logic${NC}"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}FAIL: LLMsVerifier missing timeout retry logic${NC}"
        ((TESTS_FAILED++))
    fi
}

# Test 9: Verify LLMsVerifier timeout penalty
test_llmsverifier_timeout_penalty() {
    echo -e "${YELLOW}Test 9: Verify LLMsVerifier timeout penalty${NC}"
    
    local verification_file="$PROJECT_ROOT/LLMsVerifier/llm-verifier/verification/code_verification.go"
    
    if grep -q "timeoutPenalty" "$verification_file" && \
       grep -q "calculateAverageResponseTime" "$verification_file"; then
        echo -e "${GREEN}PASS: LLMsVerifier has timeout penalty logic${NC}"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}FAIL: LLMsVerifier missing timeout penalty logic${NC}"
        ((TESTS_FAILED++))
    fi
}

# Test 10: Verify comprehensive role constants
test_role_constants() {
    echo -e "${YELLOW}Test 10: Verify comprehensive role constants${NC}"
    
    local types_file="$PROJECT_ROOT/internal/debate/comprehensive/types.go"
    
    local expected_constants=(
        "RoleArchitect"
        "RoleGenerator"
        "RoleCritic"
        "RoleRefactoring"
        "RoleTester"
        "RoleValidator"
        "RoleSecurity"
        "RolePerformance"
        "RoleModerator"
        "RoleRedTeam"
        "RoleBlueTeam"
    )
    
    local all_found=true
    for const in "${expected_constants[@]}"; do
        if grep -q "$const Role = " "$types_file"; then
            echo -e "  ${GREEN}✓${NC} Found constant: $const"
        else
            echo -e "  ${RED}✗${NC} Missing constant: $const"
            all_found=false
        fi
    done
    
    if [ "$all_found" = true ]; then
        echo -e "${GREEN}PASS: All role constants are defined${NC}"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}FAIL: Not all role constants are defined${NC}"
        ((TESTS_FAILED++))
    fi
}

# Run all tests
test_all_roles_defined
echo ""
test_all_role_prompts_exist
echo ""
test_dynamic_role_prompts
echo ""
test_footer_11_perspectives
echo ""
test_test_updated
echo ""
test_role_prompt_tests
echo ""
test_position_mapping
echo ""
test_llmsverifier_timeout_retry
echo ""
test_llmsverifier_timeout_penalty
echo ""
test_role_constants
echo ""

# Summary
echo -e "${BLUE}============================================${NC}"
echo -e "${BLUE}Challenge Summary${NC}"
echo -e "${BLUE}============================================${NC}"
echo -e "Tests Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests Failed: ${RED}$TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed! 11-role system is properly implemented.${NC}"
    exit 0
else
    echo -e "${RED}✗ Some tests failed. Please review the implementation.${NC}"
    exit 1
fi
