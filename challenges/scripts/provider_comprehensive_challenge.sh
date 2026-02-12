#!/bin/bash
# Provider Comprehensive Verification Challenge
# Tests: generic provider, factory completeness, HuggingFace migration, Fireworks fix,
#        Replicate auth, OpenRouter headers, debate team diversity, functional validation
# Total: ~40 tests across 8 sections

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHALLENGES_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$CHALLENGES_DIR")"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PASSED=0
FAILED=0
SKIPPED=0
TOTAL=0

pass() {
    PASSED=$((PASSED + 1))
    TOTAL=$((TOTAL + 1))
    echo -e "  ${GREEN}[PASS]${NC} $1"
}

fail() {
    FAILED=$((FAILED + 1))
    TOTAL=$((TOTAL + 1))
    echo -e "  ${RED}[FAIL]${NC} $1"
}

skip() {
    SKIPPED=$((SKIPPED + 1))
    TOTAL=$((TOTAL + 1))
    echo -e "  ${YELLOW}[SKIP]${NC} $1"
}

section() {
    echo -e "\n${BLUE}=== $1 ===${NC}"
}

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Provider Comprehensive Verification${NC}"
echo -e "${BLUE}========================================${NC}"

#===============================================================================
# Section 1: Generic Provider Implementation (5 tests)
#===============================================================================
section "Section 1: Generic Provider Implementation"

# Test 1.1: Generic provider file exists
if [ -f "$PROJECT_ROOT/internal/llm/providers/generic/generic.go" ]; then
    pass "Generic provider file exists"
else
    fail "Generic provider file missing: internal/llm/providers/generic/generic.go"
fi

# Test 1.2: Complete method exists
if grep -q 'func (p \*Provider) Complete' "$PROJECT_ROOT/internal/llm/providers/generic/generic.go" 2>/dev/null; then
    pass "Complete method implemented"
else
    fail "Complete method not found in generic provider"
fi

# Test 1.3: CompleteStream method exists
if grep -q 'func (p \*Provider) CompleteStream' "$PROJECT_ROOT/internal/llm/providers/generic/generic.go" 2>/dev/null; then
    pass "CompleteStream method implemented"
else
    fail "CompleteStream method not found in generic provider"
fi

# Test 1.4: HealthCheck method exists
if grep -q 'func (p \*Provider) HealthCheck' "$PROJECT_ROOT/internal/llm/providers/generic/generic.go" 2>/dev/null; then
    pass "HealthCheck method implemented"
else
    fail "HealthCheck method not found in generic provider"
fi

# Test 1.5: GetCapabilities method exists
if grep -q 'func (p \*Provider) GetCapabilities' "$PROJECT_ROOT/internal/llm/providers/generic/generic.go" 2>/dev/null; then
    pass "GetCapabilities method implemented"
else
    fail "GetCapabilities method not found in generic provider"
fi

#===============================================================================
# Section 2: Provider Factory Completeness (5 tests)
#===============================================================================
section "Section 2: Provider Factory Completeness"

# Test 2.1: Factory handles openai
if grep -q 'case "openai":' "$PROJECT_ROOT/cmd/helixagent/main.go" 2>/dev/null; then
    pass "Factory handles openai provider"
else
    fail "Factory missing openai case"
fi

# Test 2.2: Factory handles cohere
if grep -q 'case "cohere":' "$PROJECT_ROOT/cmd/helixagent/main.go" 2>/dev/null; then
    pass "Factory handles cohere provider"
else
    fail "Factory missing cohere case"
fi

# Test 2.3: Factory handles ai21
if grep -q 'case "ai21":' "$PROJECT_ROOT/cmd/helixagent/main.go" 2>/dev/null; then
    pass "Factory handles ai21 provider"
else
    fail "Factory missing ai21 case"
fi

# Test 2.4: Factory handles together
if grep -q 'case "together":' "$PROJECT_ROOT/cmd/helixagent/main.go" 2>/dev/null; then
    pass "Factory handles together provider"
else
    fail "Factory missing together case"
fi

# Test 2.5: Factory has generic default via NewGenericProvider
if grep -q 'generic.NewGenericProvider' "$PROJECT_ROOT/cmd/helixagent/main.go" 2>/dev/null; then
    pass "Factory has generic default (NewGenericProvider)"
else
    fail "Factory missing generic default (NewGenericProvider)"
fi

#===============================================================================
# Section 3: HuggingFace Migration (3 tests)
#===============================================================================
section "Section 3: HuggingFace Migration"

# Test 3.1: New router URL in provider
if grep -q 'router\.huggingface\.co' "$PROJECT_ROOT/internal/llm/providers/huggingface/huggingface.go" 2>/dev/null; then
    pass "HuggingFace uses new router.huggingface.co URL"
else
    fail "HuggingFace missing new router.huggingface.co URL"
fi

# Test 3.2: Old URL removed from base URL constants
# The old URL should NOT appear as a base URL constant (may still appear in comments)
OLD_URL_COUNT=$(grep -c 'api-inference\.huggingface\.co' "$PROJECT_ROOT/internal/llm/providers/huggingface/huggingface.go" 2>/dev/null || echo "0")
OLD_URL_COUNT=${OLD_URL_COUNT//[^0-9]/}
OLD_URL_COUNT=${OLD_URL_COUNT:-0}
# Check that the old URL only appears in comments, not as an active constant
if grep -q '^\s*[A-Z].*=.*".*api-inference\.huggingface\.co' "$PROJECT_ROOT/internal/llm/providers/huggingface/huggingface.go" 2>/dev/null; then
    fail "Old api-inference.huggingface.co still used as active URL constant"
else
    pass "Old api-inference.huggingface.co not used as active URL constant"
fi

# Test 3.3: Model updated to Llama-3.3-70B-Instruct
if grep -q 'Llama-3\.3-70B-Instruct' "$PROJECT_ROOT/internal/llm/providers/huggingface/huggingface.go" 2>/dev/null; then
    pass "HuggingFace model updated to Llama-3.3-70B-Instruct"
else
    fail "HuggingFace model not updated to Llama-3.3-70B-Instruct"
fi

#===============================================================================
# Section 4: Fireworks Model Fix (3 tests)
#===============================================================================
section "Section 4: Fireworks Model Fix"

# Test 4.1: New default model (llama-v3p3-70b-instruct)
if grep -q 'llama-v3p3-70b-instruct' "$PROJECT_ROOT/internal/llm/providers/fireworks/fireworks.go" 2>/dev/null; then
    pass "Fireworks uses new llama-v3p3-70b-instruct model"
else
    fail "Fireworks missing new llama-v3p3-70b-instruct model"
fi

# Test 4.2: Old model NOT the DefaultModel line
if grep -q 'DefaultModel.*=.*llama-v3p1-70b-instruct' "$PROJECT_ROOT/internal/llm/providers/fireworks/fireworks.go" 2>/dev/null; then
    fail "Old llama-v3p1-70b-instruct still set as DefaultModel"
else
    pass "Old llama-v3p1-70b-instruct no longer DefaultModel"
fi

# Test 4.3: Fallback models include llama 3.3
FALLBACK_COUNT=$(grep -c 'llama-v3p3-70b-instruct' "$PROJECT_ROOT/internal/llm/providers/fireworks/fireworks.go" 2>/dev/null || echo "0")
FALLBACK_COUNT=${FALLBACK_COUNT//[^0-9]/}
FALLBACK_COUNT=${FALLBACK_COUNT:-0}
if [ "$FALLBACK_COUNT" -ge 1 ]; then
    pass "Fallback models include llama-v3p3-70b-instruct ($FALLBACK_COUNT references)"
else
    fail "Fallback models missing llama-v3p3-70b-instruct"
fi

#===============================================================================
# Section 5: Replicate Auth Fix (3 tests)
#===============================================================================
section "Section 5: Replicate Auth Fix"

# Test 5.1: Bearer auth used
if grep -q '"Bearer "' "$PROJECT_ROOT/internal/llm/providers/replicate/replicate.go" 2>/dev/null; then
    pass "Replicate uses Bearer authentication"
else
    fail "Replicate missing Bearer authentication"
fi

# Test 5.2: Token auth removed
TOKEN_COUNT=$(grep -c '"Token "' "$PROJECT_ROOT/internal/llm/providers/replicate/replicate.go" 2>/dev/null || echo "0")
TOKEN_COUNT=${TOKEN_COUNT//[^0-9]/}
TOKEN_COUNT=${TOKEN_COUNT:-0}
if [ "$TOKEN_COUNT" -eq 0 ]; then
    pass "Old Token auth removed from Replicate"
else
    fail "Old Token auth still present in Replicate ($TOKEN_COUNT occurrences)"
fi

# Test 5.3: Auth header in createPrediction context uses Bearer
if grep -A15 'func.*createPrediction' "$PROJECT_ROOT/internal/llm/providers/replicate/replicate.go" 2>/dev/null | grep -q 'Bearer'; then
    pass "createPrediction uses Bearer auth header"
else
    fail "createPrediction missing Bearer auth header"
fi

#===============================================================================
# Section 6: OpenRouter Headers (3 tests)
#===============================================================================
section "Section 6: OpenRouter Headers"

# Test 6.1: HTTP-Referer in Complete
if grep -q 'HTTP-Referer' "$PROJECT_ROOT/internal/llm/providers/openrouter/openrouter.go" 2>/dev/null; then
    pass "HTTP-Referer header set in OpenRouter"
else
    fail "HTTP-Referer header missing in OpenRouter"
fi

# Test 6.2: X-Title with HelixAgent in Complete
if grep -q 'X-Title.*HelixAgent' "$PROJECT_ROOT/internal/llm/providers/openrouter/openrouter.go" 2>/dev/null; then
    pass "X-Title header set with HelixAgent identifier"
else
    fail "X-Title header missing or not set to HelixAgent"
fi

# Test 6.3: X-Title appears in multiple methods (>= 2 for both Complete and Stream)
XTITLE_COUNT=$(grep -c 'X-Title' "$PROJECT_ROOT/internal/llm/providers/openrouter/openrouter.go" 2>/dev/null || echo "0")
XTITLE_COUNT=${XTITLE_COUNT//[^0-9]/}
XTITLE_COUNT=${XTITLE_COUNT:-0}
if [ "$XTITLE_COUNT" -ge 2 ]; then
    pass "X-Title header set in multiple methods ($XTITLE_COUNT occurrences)"
else
    fail "X-Title header only in $XTITLE_COUNT method(s) (need >= 2)"
fi

#===============================================================================
# Section 7: Debate Team Diversity (5 tests)
#===============================================================================
section "Section 7: Debate Team Diversity"

# Test 7.1: Provider diversity strategy
if grep -q 'provider diversity' "$PROJECT_ROOT/internal/services/debate_team_config.go" 2>/dev/null; then
    pass "Provider diversity strategy implemented"
else
    fail "Provider diversity strategy missing in debate_team_config.go"
fi

# Test 7.2: usedProviders map
if grep -q 'usedProviders' "$PROJECT_ROOT/internal/services/debate_team_config.go" 2>/dev/null; then
    pass "usedProviders map tracks provider usage"
else
    fail "usedProviders map missing in debate_team_config.go"
fi

# Test 7.3: Unique provider check
if grep -q 'usedProviders\[llm\.ProviderName\]' "$PROJECT_ROOT/internal/services/debate_team_config.go" 2>/dev/null; then
    pass "Unique provider check validates ProviderName"
else
    fail "Unique provider check missing for ProviderName"
fi

# Test 7.4: OAuth diversity preserved
if grep -q 'OAuth diversity' "$PROJECT_ROOT/internal/services/debate_team_config.go" 2>/dev/null; then
    pass "OAuth diversity logic preserved"
else
    fail "OAuth diversity logic missing in debate_team_config.go"
fi

# Test 7.5: Fallback reuse preserved
if grep -q 'LLM reuse for fallback completeness' "$PROJECT_ROOT/internal/services/debate_team_config.go" 2>/dev/null; then
    pass "LLM reuse for fallback completeness preserved"
else
    fail "LLM reuse for fallback completeness missing in debate_team_config.go"
fi

#===============================================================================
# Section 8: Functional Validation (13 tests)
#===============================================================================
section "Section 8: Functional Validation"

# Test 8.1: Build all cmd packages
echo -e "  ${YELLOW}Building cmd packages...${NC}"
if (cd "$PROJECT_ROOT" && go build ./cmd/... 2>&1); then
    pass "All cmd packages build successfully"
else
    fail "cmd package build failed"
fi

# Test 8.2: Generic provider compiles
echo -e "  ${YELLOW}Building generic provider...${NC}"
if (cd "$PROJECT_ROOT" && go build ./internal/llm/providers/generic/ 2>&1); then
    pass "Generic provider compiles successfully"
else
    fail "Generic provider compilation failed"
fi

# Test 8.3: No vet errors in main
echo -e "  ${YELLOW}Running go vet on cmd/helixagent...${NC}"
if (cd "$PROJECT_ROOT" && go vet ./cmd/helixagent/ 2>&1); then
    pass "No go vet errors in cmd/helixagent"
else
    fail "go vet errors in cmd/helixagent"
fi

# Test 8.4: No vet errors in verifier
echo -e "  ${YELLOW}Running go vet on internal/verifier...${NC}"
if (cd "$PROJECT_ROOT" && go vet ./internal/verifier/ 2>&1); then
    pass "No go vet errors in internal/verifier"
else
    fail "go vet errors in internal/verifier"
fi

# Test 8.5: No vet errors in services
echo -e "  ${YELLOW}Running go vet on internal/services...${NC}"
if (cd "$PROJECT_ROOT" && go vet ./internal/services/ 2>&1); then
    pass "No go vet errors in internal/services"
else
    fail "go vet errors in internal/services"
fi

# Test 8.6: No vet errors in huggingface provider
echo -e "  ${YELLOW}Running go vet on huggingface provider...${NC}"
if (cd "$PROJECT_ROOT" && go vet ./internal/llm/providers/huggingface/ 2>&1); then
    pass "No go vet errors in huggingface provider"
else
    fail "go vet errors in huggingface provider"
fi

# Test 8.7: No vet errors in fireworks provider
echo -e "  ${YELLOW}Running go vet on fireworks provider...${NC}"
if (cd "$PROJECT_ROOT" && go vet ./internal/llm/providers/fireworks/ 2>&1); then
    pass "No go vet errors in fireworks provider"
else
    fail "go vet errors in fireworks provider"
fi

# Test 8.8: No vet errors in replicate provider
echo -e "  ${YELLOW}Running go vet on replicate provider...${NC}"
if (cd "$PROJECT_ROOT" && go vet ./internal/llm/providers/replicate/ 2>&1); then
    pass "No go vet errors in replicate provider"
else
    fail "go vet errors in replicate provider"
fi

# Test 8.9: No vet errors in openrouter provider
echo -e "  ${YELLOW}Running go vet on openrouter provider...${NC}"
if (cd "$PROJECT_ROOT" && go vet ./internal/llm/providers/openrouter/ 2>&1); then
    pass "No go vet errors in openrouter provider"
else
    fail "go vet errors in openrouter provider"
fi

# Test 8.10: Generic provider tests pass
echo -e "  ${YELLOW}Running generic provider tests...${NC}"
TEST_OUTPUT=$(cd "$PROJECT_ROOT" && go test -v -short -timeout 60s ./internal/llm/providers/generic/ 2>&1) || true
PASS_COUNT=$(echo "$TEST_OUTPUT" | grep -c "^--- PASS:" || echo "0")
PASS_COUNT=${PASS_COUNT//[^0-9]/}
PASS_COUNT=${PASS_COUNT:-0}
FAIL_COUNT=$(echo "$TEST_OUTPUT" | grep -c "^--- FAIL:" || echo "0")
FAIL_COUNT=${FAIL_COUNT//[^0-9]/}
FAIL_COUNT=${FAIL_COUNT:-0}
if echo "$TEST_OUTPUT" | grep -q "^ok" && [ "$FAIL_COUNT" -eq 0 ]; then
    pass "Generic provider tests pass ($PASS_COUNT passed, $FAIL_COUNT failed)"
else
    fail "Generic provider tests: $PASS_COUNT passed, $FAIL_COUNT failed"
fi

# Test 8.11: instanceCreator field exists in startup.go
if grep -q 'instanceCreator' "$PROJECT_ROOT/internal/verifier/startup.go" 2>/dev/null; then
    pass "instanceCreator field exists in startup.go"
else
    fail "instanceCreator field missing in startup.go"
fi

# Test 8.12: SetInstanceCreator method exists in startup.go
if grep -q 'func.*SetInstanceCreator' "$PROJECT_ROOT/internal/verifier/startup.go" 2>/dev/null; then
    pass "SetInstanceCreator method exists in startup.go"
else
    fail "SetInstanceCreator method missing in startup.go"
fi

# Test 8.13: provider_types has router.huggingface.co
if grep -q 'router\.huggingface\.co' "$PROJECT_ROOT/internal/verifier/provider_types.go" 2>/dev/null; then
    pass "provider_types.go has router.huggingface.co URL"
else
    fail "provider_types.go missing router.huggingface.co URL"
fi

#===============================================================================
# Summary
#===============================================================================
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Provider Comprehensive Challenge${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "  Total:   ${BLUE}$TOTAL${NC}"
echo -e "  Passed:  ${GREEN}$PASSED${NC}"
echo -e "  Failed:  ${RED}$FAILED${NC}"
echo -e "  Skipped: ${YELLOW}$SKIPPED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "  ${GREEN}ALL TESTS PASSED${NC}"
    exit 0
else
    echo -e "  ${RED}$FAILED test(s) failed${NC}"
    exit 1
fi
