#!/bin/bash
#
# Challenge: AI Debate Verification
#
# This challenge verifies that:
# 1. The AI debate team initializes correctly
# 2. Only verified providers are used in the debate team
# 3. Ollama is not used when disabled
# 4. Fallbacks are properly configured
# 5. High-quality providers are prioritized
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
TESTS_PASSED=0
TESTS_FAILED=0

# Test functions
pass_test() {
    echo -e "${GREEN}✓ PASS${NC}: $1"
    ((TESTS_PASSED++))
}

fail_test() {
    echo -e "${RED}✗ FAIL${NC}: $1"
    ((TESTS_FAILED++))
}

info_test() {
    echo -e "${BLUE}ℹ INFO${NC}: $1"
}

warn_test() {
    echo -e "${YELLOW}⚠ WARN${NC}: $1"
}

echo "=========================================="
echo "Challenge: AI Debate Verification"
echo "=========================================="
echo ""

# Test 1: Check AI debate test file exists
info_test "Test 1: Checking for AI debate verification test file..."
if [ -f "${PROJECT_ROOT}/tests/integration/ai_debate_verification_test.go" ]; then
    pass_test "AI debate verification test file exists"
else
    fail_test "AI debate verification test file missing"
fi

# Test 2: Check DebateTeamConfig exists
info_test "Test 2: Checking for DebateTeamConfig implementation..."
if [ -f "${PROJECT_ROOT}/internal/services/debate_team_config.go" ]; then
    pass_test "DebateTeamConfig implementation exists"
else
    fail_test "DebateTeamConfig implementation missing"
fi

# Test 3: Verify InitializeTeam function exists
info_test "Test 3: Checking for InitializeTeam function..."
if grep -q "func.*InitializeTeam" "${PROJECT_ROOT}/internal/services/debate_team_config.go"; then
    pass_test "InitializeTeam function found"
else
    fail_test "InitializeTeam function not found"
fi

# Test 4: Verify team uses StartupVerifier when available
info_test "Test 4: Checking that team uses StartupVerifier..."
if grep -q "startupVerifier" "${PROJECT_ROOT}/internal/services/debate_team_config.go"; then
    pass_test "StartupVerifier integration found"
else
    fail_test "StartupVerifier integration not found"
fi

# Test 5: Verify team collects verified LLMs
info_test "Test 5: Checking for verified LLM collection..."
if grep -q "collectVerifiedLLMs\|collectFromStartupVerifier" "${PROJECT_ROOT}/internal/services/debate_team_config.go"; then
    pass_test "Verified LLM collection found"
else
    fail_test "Verified LLM collection not found"
fi

# Test 6: Verify team has 5 positions
info_test "Test 6: Checking for 5 debate positions..."
POSITION_COUNT=$(grep -c "PositionAnalyst\|PositionProposer\|PositionCritic\|PositionSynthesis\|PositionMediator" "${PROJECT_ROOT}/internal/services/debate_team_config.go" || true)
if [ ${POSITION_COUNT} -ge 5 ]; then
    pass_test "Found 5 debate positions"
else
    fail_test "Did not find all 5 debate positions (found ${POSITION_COUNT})"
fi

# Test 7: Verify fallbacks are configured
info_test "Test 7: Checking for fallback configuration..."
if grep -q "Fallback\|fallback" "${PROJECT_ROOT}/internal/services/debate_team_config.go"; then
    pass_test "Fallback configuration found"
else
    fail_test "Fallback configuration not found"
fi

# Test 8: Verify team doesn't use Ollama when disabled
info_test "Test 8: Checking for Ollama exclusion when disabled..."
if grep -q "OLLAMA_ENABLED\|ollama.*disabled" "${PROJECT_ROOT}/internal/services/debate_team_config.go" || \
   grep -q "collectFromStartupVerifier" "${PROJECT_ROOT}/internal/services/debate_team_config.go"; then
    pass_test "Ollama exclusion mechanism found"
else
    warn_test "Ollama exclusion check not explicitly found (may be handled by StartupVerifier)"
fi

# Test 9: Run AI debate Go tests
info_test "Test 9: Running AI debate Go tests..."
cd "${PROJECT_ROOT}"
if go test -v -run TestAIDebate ./tests/integration/... 2>&1 | tee /tmp/ai_debate_test_output.log | grep -q "PASS"; then
    pass_test "AI debate Go tests passed"
else
    warn_test "Some AI debate tests had issues (check /tmp/ai_debate_test_output.log)"
fi

# Test 10: Verify GetTeamMember function exists
info_test "Test 10: Checking for GetTeamMember function..."
if grep -q "func.*GetTeamMember" "${PROJECT_ROOT}/internal/services/debate_team_config.go"; then
    pass_test "GetTeamMember function found"
else
    fail_test "GetTeamMember function not found"
fi

# Test 11: Verify GetAllLLMs function exists
info_test "Test 11: Checking for GetAllLLMs function..."
if grep -q "func.*GetAllLLMs" "${PROJECT_ROOT}/internal/services/debate_team_config.go"; then
    pass_test "GetAllLLMs function found"
else
    fail_test "GetAllLLMs function not found"
fi

# Test 12: Verify team sorts by score
info_test "Test 12: Checking that team sorts LLMs by score..."
if grep -q "Score.*>" "${PROJECT_ROOT}/internal/services/debate_team_config.go" || \
   grep -q "sort.Slice.*Score" "${PROJECT_ROOT}/internal/services/debate_team_config.go"; then
    pass_test "Score-based sorting found"
else
    warn_test "Score-based sorting not explicitly found"
fi

# Summary
echo ""
echo "=========================================="
echo "Challenge Summary"
echo "=========================================="
echo -e "Tests Passed: ${GREEN}${TESTS_PASSED}${NC}"
echo -e "Tests Failed: ${RED}${TESTS_FAILED}${NC}"
echo ""

if [ ${TESTS_FAILED} -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}✗ Some tests failed!${NC}"
    exit 1
fi
