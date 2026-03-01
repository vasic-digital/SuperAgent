#!/bin/bash

# Challenge: Verification Report Generation
# This challenge verifies that the provider verification report is properly generated

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
echo -e "${BLUE}Verification Report Challenge${NC}"
echo -e "${BLUE}============================================${NC}"
echo ""

# Test 1: Verify report generator file exists
test_report_generator_exists() {
    echo -e "${YELLOW}Test 1: Verify report generator file exists${NC}"
    
    local report_file="$PROJECT_ROOT/internal/services/verification_report.go"
    
    if [ -f "$report_file" ]; then
        echo -e "  ${GREEN}✓${NC} Report generator file exists: verification_report.go"
        ((TESTS_PASSED++))
    else
        echo -e "  ${RED}✗${NC} Report generator file missing"
        ((TESTS_FAILED++))
    fi
}

# Test 2: Verify report generator test file exists
test_report_generator_test_exists() {
    echo -e "${YELLOW}Test 2: Verify report generator test file exists${NC}"
    
    local test_file="$PROJECT_ROOT/internal/services/verification_report_test.go"
    
    if [ -f "$test_file" ]; then
        echo -e "  ${GREEN}✓${NC} Test file exists: verification_report_test.go"
        ((TESTS_PASSED++))
    else
        echo -e "  ${RED}✗${NC} Test file missing"
        ((TESTS_FAILED++))
    fi
}

# Test 3: Verify ReportGenerator has required methods
test_report_generator_methods() {
    echo -e "${YELLOW}Test 3: Verify ReportGenerator has required methods${NC}"
    
    local report_file="$PROJECT_ROOT/internal/services/verification_report.go"
    local all_found=true
    
    local methods=(
        "GenerateReport"
        "GetReportPath"
        "GetReportURL"
        "formatReport"
        "saveReport"
    )
    
    for method in "${methods[@]}"; do
        if grep -q "func.*$method" "$report_file"; then
            echo -e "  ${GREEN}✓${NC} Method found: $method"
        else
            echo -e "  ${RED}✗${NC} Method missing: $method"
            all_found=false
        fi
    done
    
    if [ "$all_found" = true ]; then
        ((TESTS_PASSED++))
    else
        ((TESTS_FAILED++))
    fi
}

# Test 4: Verify handler has report generator field
test_handler_has_generator() {
    echo -e "${YELLOW}Test 4: Verify handler has report generator field${NC}"
    
    local handler_file="$PROJECT_ROOT/internal/handlers/openai_compatible.go"
    
    if grep -q "reportGenerator.*VerificationReportGenerator" "$handler_file"; then
        echo -e "  ${GREEN}✓${NC} Handler has reportGenerator field"
        ((TESTS_PASSED++))
    else
        echo -e "  ${RED}✗${NC} Handler missing reportGenerator field"
        ((TESTS_FAILED++))
    fi
}

# Test 5: Verify handler has SetReportGenerator method
test_handler_setter() {
    echo -e "${YELLOW}Test 5: Verify handler has SetReportGenerator method${NC}"
    
    local handler_file="$PROJECT_ROOT/internal/handlers/openai_compatible.go"
    
    if grep -q "func.*SetReportGenerator" "$handler_file"; then
        echo -e "  ${GREEN}✓${NC} Handler has SetReportGenerator method"
        ((TESTS_PASSED++))
    else
        echo -e "  ${RED}✗${NC} Handler missing SetReportGenerator method"
        ((TESTS_FAILED++))
    fi
}

# Test 6: Verify debate introduction includes report link
test_debate_intro_has_link() {
    echo -e "${YELLOW}Test 6: Verify debate introduction includes report link${NC}"
    
    local handler_file="$PROJECT_ROOT/internal/handlers/openai_compatible.go"
    
    if grep -q "Provider Verification Report" "$handler_file"; then
        echo -e "  ${GREEN}✓${NC} Debate introduction includes report link"
        ((TESTS_PASSED++))
    else
        echo -e "  ${RED}✗${NC} Debate introduction missing report link"
        ((TESTS_FAILED++))
    fi
}

# Test 7: Verify positions have team information
test_positions_have_teams() {
    echo -e "${YELLOW}Test 7: Verify positions have team information${NC}"
    
    local handler_file="$PROJECT_ROOT/internal/handlers/openai_compatible.go"
    
    if grep -q "teamName.*string" "$handler_file" && grep -q "teamEmoji.*string" "$handler_file"; then
        echo -e "  ${GREEN}✓${NC} Positions have team information"
        ((TESTS_PASSED++))
    else
        echo -e "  ${RED}✗${NC} Positions missing team information"
        ((TESTS_FAILED++))
    fi
}

# Test 8: Verify team headers are emitted
test_team_headers_emitted() {
    echo -e "${YELLOW}Test 8: Verify team headers are emitted during debate${NC}"
    
    local handler_file="$PROJECT_ROOT/internal/handlers/openai_compatible.go"
    
    if grep -q "teamHeader.*teamName.*teamEmoji" "$handler_file"; then
        echo -e "  ${GREEN}✓${NC} Team headers are emitted"
        ((TESTS_PASSED++))
    else
        echo -e "  ${RED}✗${NC} Team headers not emitted"
        ((TESTS_FAILED++))
    fi
}

# Test 9: Verify report includes model rankings
test_report_has_rankings() {
    echo -e "${YELLOW}Test 9: Verify report includes model rankings${NC}"
    
    local report_file="$PROJECT_ROOT/internal/services/verification_report.go"
    
    if grep -q "Model Rankings" "$report_file" && grep -q "sort.Slice" "$report_file"; then
        echo -e "  ${GREEN}✓${NC} Report includes sorted model rankings"
        ((TESTS_PASSED++))
    else
        echo -e "  ${RED}✗${NC} Report missing model rankings"
        ((TESTS_FAILED++))
    fi
}

# Test 10: Verify report includes error reasons
test_report_has_error_reasons() {
    echo -e "${YELLOW}Test 10: Verify report includes error reasons for failed models${NC}"
    
    local report_file="$PROJECT_ROOT/internal/services/verification_report.go"
    
    if grep -q "inferErrorReason" "$report_file" && grep -q "ErrorReason" "$report_file"; then
        echo -e "  ${GREEN}✓${NC} Report includes error reasons"
        ((TESTS_PASSED++))
    else
        echo -e "  ${RED}✗${NC} Report missing error reasons"
        ((TESTS_FAILED++))
    fi
}

# Test 11: Verify report includes detailed breakdown
test_report_has_details() {
    echo -e "${YELLOW}Test 11: Verify report includes detailed breakdown${NC}"
    
    local report_file="$PROJECT_ROOT/internal/services/verification_report.go"
    
    if grep -q "Detailed Breakdown" "$report_file" && grep -q "Healthy Models" "$report_file" && grep -q "Failed Models" "$report_file"; then
        echo -e "  ${GREEN}✓${NC} Report includes detailed breakdown sections"
        ((TESTS_PASSED++))
    else
        echo -e "  ${RED}✗${NC} Report missing detailed breakdown"
        ((TESTS_FAILED++))
    fi
}

# Test 12: Verify unit tests cover key scenarios
test_unit_tests_coverage() {
    echo -e "${YELLOW}Test 12: Verify unit tests cover key scenarios${NC}"
    
    local test_file="$PROJECT_ROOT/internal/services/verification_report_test.go"
    local all_found=true
    
    local tests=(
        "TestVerificationReportGenerator_GenerateReport"
        "TestVerificationReportGenerator_FormatReport"
        "TestVerificationReportGenerator_InferErrorReason"
        "TestVerificationReportGenerator_SortedByScore"
    )
    
    for test in "${tests[@]}"; do
        if grep -q "func $test" "$test_file"; then
            echo -e "  ${GREEN}✓${NC} Test found: $test"
        else
            echo -e "  ${RED}✗${NC} Test missing: $test"
            all_found=false
        fi
    done
    
    if [ "$all_found" = true ]; then
        ((TESTS_PASSED++))
    else
        ((TESTS_FAILED++))
    fi
}

# Run all tests
test_report_generator_exists
echo ""
test_report_generator_test_exists
echo ""
test_report_generator_methods
echo ""
test_handler_has_generator
echo ""
test_handler_setter
echo ""
test_debate_intro_has_link
echo ""
test_positions_have_teams
echo ""
test_team_headers_emitted
echo ""
test_report_has_rankings
echo ""
test_report_has_error_reasons
echo ""
test_report_has_details
echo ""
test_unit_tests_coverage
echo ""

# Summary
echo -e "${BLUE}============================================${NC}"
echo -e "${BLUE}Challenge Summary${NC}"
echo -e "${BLUE}============================================${NC}"
echo -e "Tests Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests Failed: ${RED}$TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed! Verification report system is properly implemented.${NC}"
    exit 0
else
    echo -e "${RED}✗ Some tests failed. Please review the implementation.${NC}"
    exit 1
fi
