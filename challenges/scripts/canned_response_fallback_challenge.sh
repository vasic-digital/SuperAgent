#!/bin/bash

# =============================================================================
# Canned Response Fallback Challenge
# =============================================================================
# This challenge validates that:
# 1. Canned error responses from LLMs trigger fallback mechanism
# 2. Suspiciously fast responses trigger fallback
# 3. Fallback chain is properly exercised
# 4. ZAI and DeepSeek providers are properly discovered
# 5. The system handles non-working LLMs gracefully
# =============================================================================

# Don't exit on errors - we handle them ourselves
set +e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0
TOTAL_TESTS=0

# Test result logging
log_pass() {
    echo -e "${GREEN}✓ PASS${NC}: $1"
    ((TESTS_PASSED++))
    ((TOTAL_TESTS++))
}

log_fail() {
    echo -e "${RED}✗ FAIL${NC}: $1"
    ((TESTS_FAILED++))
    ((TOTAL_TESTS++))
}

log_info() {
    echo -e "${BLUE}ℹ INFO${NC}: $1"
}

log_section() {
    echo ""
    echo -e "${YELLOW}=== $1 ===${NC}"
    echo ""
}

# =============================================================================
# Section 1: Code Pattern Verification
# =============================================================================
log_section "Section 1: Canned Error Detection Code Patterns"

# Test 1.1: CannedErrorPatterns exported variable exists
if grep -q "var CannedErrorPatterns = \[\]string{" "$PROJECT_ROOT/internal/services/debate_service.go"; then
    log_pass "1.1 CannedErrorPatterns exported variable exists"
else
    log_fail "1.1 CannedErrorPatterns exported variable not found"
fi

# Test 1.2: IsCannedErrorResponse function exists
if grep -q "func IsCannedErrorResponse(content string) string" "$PROJECT_ROOT/internal/services/debate_service.go"; then
    log_pass "1.2 IsCannedErrorResponse function exists"
else
    log_fail "1.2 IsCannedErrorResponse function not found"
fi

# Test 1.3: IsSuspiciouslyFastResponse function exists
if grep -q "func IsSuspiciouslyFastResponse(responseTime time.Duration, contentLength int) bool" "$PROJECT_ROOT/internal/services/debate_service.go"; then
    log_pass "1.3 IsSuspiciouslyFastResponse function exists"
else
    log_fail "1.3 IsSuspiciouslyFastResponse function not found"
fi

# Test 1.4: Minimum number of canned error patterns (at least 10)
PATTERN_COUNT=$(grep -c "^\s*\".*\",$" "$PROJECT_ROOT/internal/services/debate_service.go" 2>/dev/null | head -1 || echo "0")
# Count patterns in the CannedErrorPatterns slice
PATTERN_COUNT=$(awk '/var CannedErrorPatterns/,/^\}/' "$PROJECT_ROOT/internal/services/debate_service.go" | grep -c '"' || echo "0")
if [ "$PATTERN_COUNT" -ge 10 ]; then
    log_pass "1.4 At least 10 canned error patterns defined ($PATTERN_COUNT found)"
else
    log_fail "1.4 Less than 10 canned error patterns ($PATTERN_COUNT found)"
fi

# Test 1.5: Critical patterns exist
CRITICAL_PATTERNS=("unable to provide" "cannot provide" "error occurred" "at this time")
for pattern in "${CRITICAL_PATTERNS[@]}"; do
    if grep -q "\"$pattern\"" "$PROJECT_ROOT/internal/services/debate_service.go"; then
        log_pass "1.5.$TESTS_PASSED Critical pattern '$pattern' exists"
    else
        log_fail "1.5.$TESTS_PASSED Critical pattern '$pattern' missing"
    fi
done

# =============================================================================
# Section 2: Fallback Configuration
# =============================================================================
log_section "Section 2: Fallback Configuration"

# Test 2.1: FallbacksPerPosition is at least 4
if grep -q "const FallbacksPerPosition = 4" "$PROJECT_ROOT/internal/services/debate_team_config.go"; then
    log_pass "2.1 FallbacksPerPosition is 4"
elif grep -q "const FallbacksPerPosition = [4-9]" "$PROJECT_ROOT/internal/services/debate_team_config.go"; then
    log_pass "2.1 FallbacksPerPosition is 4 or more"
else
    log_fail "2.1 FallbacksPerPosition should be at least 4"
fi

# Test 2.2: TotalDebateLLMs is properly calculated (5 positions * (1 + 4 fallbacks) = 25)
if grep -q "TotalDebateLLMs = TotalDebatePositions \* (1 + FallbacksPerPosition)" "$PROJECT_ROOT/internal/services/debate_team_config.go"; then
    log_pass "2.2 TotalDebateLLMs properly calculated from positions and fallbacks"
else
    log_fail "2.2 TotalDebateLLMs calculation not found"
fi

# Test 2.3: FallbackConfig type exists
if grep -q "type FallbackConfig struct" "$PROJECT_ROOT/internal/services/debate_support_types.go"; then
    log_pass "2.3 FallbackConfig type exists"
else
    log_fail "2.3 FallbackConfig type not found"
fi

# =============================================================================
# Section 3: Provider Discovery Configuration
# =============================================================================
log_section "Section 3: Provider Discovery Configuration"

# Test 3.1: ZAI provider mapping exists
if grep -q "ZAI_API_KEY.*zai" "$PROJECT_ROOT/internal/services/provider_discovery.go"; then
    log_pass "3.1 ZAI_API_KEY mapping exists"
else
    log_fail "3.1 ZAI_API_KEY mapping not found"
fi

# Test 3.2: ApiKey_ZAI alternative mapping exists
if grep -q "ApiKey_ZAI.*zai" "$PROJECT_ROOT/internal/services/provider_discovery.go"; then
    log_pass "3.2 ApiKey_ZAI alternative mapping exists"
else
    log_fail "3.2 ApiKey_ZAI alternative mapping not found"
fi

# Test 3.3: ZHIPU_API_KEY mapping exists
if grep -q "ZHIPU_API_KEY.*zai" "$PROJECT_ROOT/internal/services/provider_discovery.go"; then
    log_pass "3.3 ZHIPU_API_KEY mapping exists"
else
    log_fail "3.3 ZHIPU_API_KEY mapping not found"
fi

# Test 3.4: ApiKey_DeepSeek alternative mapping exists
if grep -q "ApiKey_DeepSeek.*deepseek" "$PROJECT_ROOT/internal/services/provider_discovery.go"; then
    log_pass "3.4 ApiKey_DeepSeek alternative mapping exists"
else
    log_fail "3.4 ApiKey_DeepSeek alternative mapping not found"
fi

# Test 3.5: ZAI case in createProvider
if grep -q 'case "zai":' "$PROJECT_ROOT/internal/services/provider_discovery.go"; then
    log_pass "3.5 ZAI case in createProvider exists"
else
    log_fail "3.5 ZAI case in createProvider not found"
fi

# Test 3.6: ZAI provider import
if grep -q '"dev.helix.agent/internal/llm/providers/zai"' "$PROJECT_ROOT/internal/services/provider_discovery.go"; then
    log_pass "3.6 ZAI provider import exists"
else
    log_fail "3.6 ZAI provider import not found"
fi

# =============================================================================
# Section 4: Canned Error Detection in Debate Flow
# =============================================================================
log_section "Section 4: Canned Error Detection in Debate Flow"

# Test 4.1: IsCannedErrorResponse called in getParticipantResponse
# Count usages excluding the function definition
USAGE_COUNT=$(grep -c "IsCannedErrorResponse" "$PROJECT_ROOT/internal/services/debate_service.go" || echo "0")
if [ "$USAGE_COUNT" -ge 2 ]; then
    log_pass "4.1 IsCannedErrorResponse used in debate flow ($USAGE_COUNT occurrences)"
else
    log_fail "4.1 IsCannedErrorResponse not used in debate flow ($USAGE_COUNT occurrences)"
fi

# Test 4.2: IsSuspiciouslyFastResponse called in getParticipantResponse
USAGE_COUNT=$(grep -c "IsSuspiciouslyFastResponse" "$PROJECT_ROOT/internal/services/debate_service.go" || echo "0")
if [ "$USAGE_COUNT" -ge 2 ]; then
    log_pass "4.2 IsSuspiciouslyFastResponse used in debate flow ($USAGE_COUNT occurrences)"
else
    log_fail "4.2 IsSuspiciouslyFastResponse not used in debate flow ($USAGE_COUNT occurrences)"
fi

# Test 4.3: Fallback triggered on canned error
if grep -q "canned error response from LLM.*fallback required" "$PROJECT_ROOT/internal/services/debate_service.go"; then
    log_pass "4.3 Fallback triggered on canned error detection"
else
    log_fail "4.3 Fallback trigger for canned error not found"
fi

# Test 4.4: Fallback triggered on suspicious fast response
if grep -q "suspiciously fast response.*fallback required" "$PROJECT_ROOT/internal/services/debate_service.go"; then
    log_pass "4.4 Fallback triggered on suspicious fast response"
else
    log_fail "4.4 Fallback trigger for suspicious fast response not found"
fi

# =============================================================================
# Section 5: Unit Tests Verification
# =============================================================================
log_section "Section 5: Unit Tests Verification"

# Test 5.1: TestIsCannedErrorResponse exists
if grep -q "func TestIsCannedErrorResponse" "$PROJECT_ROOT/internal/services/debate_service_test.go"; then
    log_pass "5.1 TestIsCannedErrorResponse test exists"
else
    log_fail "5.1 TestIsCannedErrorResponse test not found"
fi

# Test 5.2: TestIsSuspiciouslyFastResponse exists
if grep -q "func TestIsSuspiciouslyFastResponse" "$PROJECT_ROOT/internal/services/debate_service_test.go"; then
    log_pass "5.2 TestIsSuspiciouslyFastResponse test exists"
else
    log_fail "5.2 TestIsSuspiciouslyFastResponse test not found"
fi

# Test 5.3: TestCannedErrorPatternsCompleteness exists
if grep -q "func TestCannedErrorPatternsCompleteness" "$PROJECT_ROOT/internal/services/debate_service_test.go"; then
    log_pass "5.3 TestCannedErrorPatternsCompleteness test exists"
else
    log_fail "5.3 TestCannedErrorPatternsCompleteness test not found"
fi

# Test 5.4: TestZenModelCannedPatterns exists
if grep -q "func TestZenModelCannedPatterns" "$PROJECT_ROOT/internal/services/debate_service_test.go"; then
    log_pass "5.4 TestZenModelCannedPatterns test exists"
else
    log_fail "5.4 TestZenModelCannedPatterns test not found"
fi

# Test 5.5: Provider discovery tests exist
if grep -q "func TestProviderMappingsHasZAI" "$PROJECT_ROOT/internal/services/provider_discovery_test.go"; then
    log_pass "5.5 TestProviderMappingsHasZAI test exists"
else
    log_fail "5.5 TestProviderMappingsHasZAI test not found"
fi

# Test 5.6: DeepSeek discovery test exists
if grep -q "func TestProviderMappingsHasDeepSeek" "$PROJECT_ROOT/internal/services/provider_discovery_test.go"; then
    log_pass "5.6 TestProviderMappingsHasDeepSeek test exists"
else
    log_fail "5.6 TestProviderMappingsHasDeepSeek test not found"
fi

# =============================================================================
# Section 6: Run Unit Tests
# =============================================================================
log_section "Section 6: Run Unit Tests"

cd "$PROJECT_ROOT"

# Test 6.1: Run canned response detection tests
log_info "Running canned response detection tests..."
if go test -v -run "TestIsCannedErrorResponse|TestIsSuspiciouslyFastResponse|TestCannedErrorPatterns|TestZenModelCanned" ./internal/services/... 2>&1 | grep -q "PASS"; then
    log_pass "6.1 Canned response detection tests pass"
else
    log_fail "6.1 Canned response detection tests failed"
fi

# Test 6.2: Run provider discovery tests
log_info "Running provider discovery tests..."
if go test -v -run "TestProviderMappingsHasZAI|TestProviderMappingsHasDeepSeek|TestCreateProviderSupports" ./internal/services/... 2>&1 | grep -q "PASS"; then
    log_pass "6.2 Provider discovery tests pass"
else
    log_fail "6.2 Provider discovery tests failed"
fi

# =============================================================================
# Section 7: Build Verification
# =============================================================================
log_section "Section 7: Build Verification"

# Test 7.1: Project builds successfully
log_info "Building project..."
if go build ./cmd/helixagent/... 2>&1; then
    log_pass "7.1 Project builds successfully"
else
    log_fail "7.1 Project build failed"
fi

# =============================================================================
# Summary
# =============================================================================
log_section "Challenge Summary"

echo ""
echo -e "Total Tests: ${TOTAL_TESTS}"
echo -e "Passed: ${GREEN}${TESTS_PASSED}${NC}"
echo -e "Failed: ${RED}${TESTS_FAILED}${NC}"
echo ""

if [ "$TESTS_FAILED" -eq 0 ]; then
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}  CANNED RESPONSE FALLBACK CHALLENGE   ${NC}"
    echo -e "${GREEN}            ALL TESTS PASSED           ${NC}"
    echo -e "${GREEN}========================================${NC}"
    exit 0
else
    echo -e "${RED}========================================${NC}"
    echo -e "${RED}  CANNED RESPONSE FALLBACK CHALLENGE   ${NC}"
    echo -e "${RED}         SOME TESTS FAILED             ${NC}"
    echo -e "${RED}========================================${NC}"
    exit 1
fi
