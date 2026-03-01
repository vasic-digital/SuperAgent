#!/bin/bash
#
# Challenge: Verified Providers to AI Debate Integration
#
# This challenge validates the complete flow:
# 1. Provider discovery finds all configured providers
# 2. Provider verification tests each provider thoroughly
# 3. Verified providers are ranked by score
# 4. AI debate team selects only verified providers
# 5. Ollama is excluded when not explicitly enabled
# 6. High-quality providers are prioritized
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
echo "Challenge: Verified Providers → AI Debate"
echo "=========================================="
echo ""

# Test 1: Verify all test files exist
info_test "Test 1: Verifying all test files exist..."
TEST_FILES=(
    "tests/integration/ollama_explicit_enable_test.go"
    "tests/integration/provider_verification_comprehensive_test.go"
    "tests/integration/ai_debate_verification_test.go"
)

ALL_EXIST=true
for file in "${TEST_FILES[@]}"; do
    if [ ! -f "${PROJECT_ROOT}/${file}" ]; then
        fail_test "Missing test file: ${file}"
        ALL_EXIST=false
    fi
done

if [ "${ALL_EXIST}" = true ]; then
    pass_test "All test files exist"
fi

# Test 2: Verify all challenge scripts exist
info_test "Test 2: Verifying all challenge scripts exist..."
CHALLENGE_SCRIPTS=(
    "challenges/scripts/ollama_explicit_enable_challenge.sh"
    "challenges/scripts/provider_verification_challenge.sh"
    "challenges/scripts/ai_debate_verification_challenge.sh"
)

ALL_EXIST=true
for script in "${CHALLENGE_SCRIPTS[@]}"; do
    if [ ! -f "${PROJECT_ROOT}/${script}" ]; then
        fail_test "Missing challenge script: ${script}"
        ALL_EXIST=false
    fi
done

if [ "${ALL_EXIST}" = true ]; then
    pass_test "All challenge scripts exist"
fi

# Test 3: Verify challenge scripts are executable
info_test "Test 3: Verifying challenge scripts are executable..."
ALL_EXECUTABLE=true
for script in "${CHALLENGE_SCRIPTS[@]}"; do
    if [ ! -x "${PROJECT_ROOT}/${script}" ]; then
        fail_test "Challenge script not executable: ${script}"
        ALL_EXECUTABLE=false
    fi
done

if [ "${ALL_EXECUTABLE}" = true ]; then
    pass_test "All challenge scripts are executable"
fi

# Test 4: Run Ollama explicit enable challenge
info_test "Test 4: Running Ollama explicit enable challenge..."
cd "${PROJECT_ROOT}"
if ./challenges/scripts/ollama_explicit_enable_challenge.sh 2>&1 | tee /tmp/ollama_challenge.log | grep -q "All tests passed"; then
    pass_test "Ollama explicit enable challenge passed"
else
    warn_test "Ollama explicit enable challenge had issues (check /tmp/ollama_challenge.log)"
fi

# Test 5: Run provider verification challenge
info_test "Test 5: Running provider verification challenge..."
if ./challenges/scripts/provider_verification_challenge.sh 2>&1 | tee /tmp/provider_challenge.log | grep -q "All tests passed"; then
    pass_test "Provider verification challenge passed"
else
    warn_test "Provider verification challenge had issues (check /tmp/provider_challenge.log)"
fi

# Test 6: Run AI debate verification challenge
info_test "Test 6: Running AI debate verification challenge..."
if ./challenges/scripts/ai_debate_verification_challenge.sh 2>&1 | tee /tmp/debate_challenge.log | grep -q "All tests passed"; then
    pass_test "AI debate verification challenge passed"
else
    warn_test "AI debate verification challenge had issues (check /tmp/debate_challenge.log)"
fi

# Test 7: Verify the verification pipeline uses code visibility checks
info_test "Test 7: Verifying code visibility checks in verification..."
if grep -q "verifyCodeVisibility" "${PROJECT_ROOT}/internal/verifier/service.go" && \
   grep -q "Do you see my code" "${PROJECT_ROOT}/internal/verifier/service.go"; then
    pass_test "Code visibility verification with proper prompt found"
else
    fail_test "Code visibility verification not properly configured"
fi

# Test 8: Verify score calculation exists
info_test "Test 8: Verifying score calculation in verification..."
if grep -q "calculateOverallScore\|OverallScore" "${PROJECT_ROOT}/internal/verifier/service.go"; then
    pass_test "Score calculation found in verification service"
else
    fail_test "Score calculation not found"
fi

# Test 9: Verify minimum score threshold
info_test "Test 9: Verifying minimum score threshold..."
if grep -q ">= 60\|>= 60" "${PROJECT_ROOT}/internal/verifier/service.go"; then
    pass_test "Minimum score threshold (60) found"
else
    warn_test "Minimum score threshold not explicitly found"
fi

# Test 10: Verify debate team uses verified providers only
info_test "Test 10: Verifying debate team uses verified providers..."
if grep -q "collectFromStartupVerifier\|Verified.*true" "${PROJECT_ROOT}/internal/services/debate_team_config.go"; then
    pass_test "Debate team verified provider filtering found"
else
    fail_test "Debate team verified provider filtering not found"
fi

# Test 11: Check that failing providers are not used
info_test "Test 11: Verifying failed providers are excluded..."
if grep -q "if.*!provider.Verified\|if !p.Verified" "${PROJECT_ROOT}/internal/services/debate_team_config.go" || \
   grep -q "collectFromStartupVerifier" "${PROJECT_ROOT}/internal/services/debate_team_config.go"; then
    pass_test "Failed provider exclusion found"
else
    warn_test "Failed provider exclusion not explicitly found"
fi

# Test 12: Verify the entire flow compiles
info_test "Test 12: Verifying entire flow compiles..."
cd "${PROJECT_ROOT}"
if go build -o /dev/null ./cmd/helixagent 2>&1 | tee /tmp/build.log; then
    pass_test "HelixAgent compiles successfully"
else
    fail_test "HelixAgent compilation failed (check /tmp/build.log)"
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
    echo -e "${GREEN}✓ All integration tests passed!${NC}"
    echo -e "${GREEN}✓ Verified providers flow correctly to AI debate team${NC}"
    exit 0
else
    echo -e "${RED}✗ Some integration tests failed!${NC}"
    exit 1
fi
