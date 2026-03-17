#!/bin/bash
# OpenRouter Comprehensive Verification Challenge
# VALIDATES: OpenRouter provider structure, model coverage, features,
#            integration, test execution, and LLMsVerifier integration
# Total: ~30 tests across 6 groups

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

OPENROUTER_DIR="$PROJECT_ROOT/internal/llm/providers/openrouter"
VERIFIER_DIR="$PROJECT_ROOT/internal/verifier"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  OpenRouter Comprehensive Challenge${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "  Validates: Provider structure, models, features,"
echo -e "  integration, test execution, LLMsVerifier integration"

#===============================================================================
# Group 1: Provider Structure (6 tests)
#===============================================================================
section "Group 1: Provider Structure (6 tests)"

# Test 1.1: Provider struct exists
if grep -q 'type SimpleOpenRouterProvider struct' "$OPENROUTER_DIR/openrouter.go" 2>/dev/null; then
    pass "SimpleOpenRouterProvider struct exists in openrouter.go"
else
    fail "SimpleOpenRouterProvider struct NOT found in openrouter.go"
fi

# Test 1.2: Complete method exists
if grep -q 'func (p \*SimpleOpenRouterProvider) Complete' "$OPENROUTER_DIR/openrouter.go" 2>/dev/null; then
    pass "Complete method exists on SimpleOpenRouterProvider"
else
    fail "Complete method NOT found on SimpleOpenRouterProvider"
fi

# Test 1.3: CompleteStream method exists
if grep -q 'func (p \*SimpleOpenRouterProvider) CompleteStream' "$OPENROUTER_DIR/openrouter.go" 2>/dev/null; then
    pass "CompleteStream method exists on SimpleOpenRouterProvider"
else
    fail "CompleteStream method NOT found on SimpleOpenRouterProvider"
fi

# Test 1.4: HealthCheck method exists
if grep -q 'func (p \*SimpleOpenRouterProvider) HealthCheck' "$OPENROUTER_DIR/openrouter.go" 2>/dev/null; then
    pass "HealthCheck method exists on SimpleOpenRouterProvider"
else
    fail "HealthCheck method NOT found on SimpleOpenRouterProvider"
fi

# Test 1.5: GetCapabilities method exists
if grep -q 'func (p \*SimpleOpenRouterProvider) GetCapabilities' "$OPENROUTER_DIR/openrouter.go" 2>/dev/null; then
    pass "GetCapabilities method exists on SimpleOpenRouterProvider"
else
    fail "GetCapabilities method NOT found on SimpleOpenRouterProvider"
fi

# Test 1.6: Registered in provider_registry.go
if grep -q 'case "openrouter":' "$PROJECT_ROOT/internal/services/provider_registry.go" 2>/dev/null; then
    pass "OpenRouter registered in provider_registry.go"
else
    fail "OpenRouter NOT registered in provider_registry.go"
fi

#===============================================================================
# Group 2: Updated Model Coverage (8 tests)
#===============================================================================
section "Group 2: Updated Model Coverage (8 tests)"

VERIFIER_FALLBACK="$PROJECT_ROOT/LLMsVerifier/llm-verifier/providers/fallback_models.go"

# Test 2.1: openai/gpt-5.2 in fallback models
if grep -q 'openai/gpt-5\.2' "$VERIFIER_FALLBACK" 2>/dev/null; then
    pass "openai/gpt-5.2 in OpenRouter fallback models"
else
    fail "openai/gpt-5.2 NOT found in OpenRouter fallback models"
fi

# Test 2.2: openai/gpt-4.1 in fallback models
if grep -q 'openai/gpt-4\.1' "$VERIFIER_FALLBACK" 2>/dev/null; then
    pass "openai/gpt-4.1 in OpenRouter fallback models"
else
    fail "openai/gpt-4.1 NOT found in OpenRouter fallback models"
fi

# Test 2.3: anthropic/claude-sonnet-4.6 in fallback models
if grep -q 'anthropic/claude-sonnet-4\.6' "$VERIFIER_FALLBACK" 2>/dev/null; then
    pass "anthropic/claude-sonnet-4.6 in OpenRouter fallback models"
else
    fail "anthropic/claude-sonnet-4.6 NOT found in OpenRouter fallback models"
fi

# Test 2.4: google/gemini-2.5-pro in fallback models
if grep -q 'google/gemini-2\.5-pro' "$VERIFIER_FALLBACK" 2>/dev/null; then
    pass "google/gemini-2.5-pro in OpenRouter fallback models"
else
    fail "google/gemini-2.5-pro NOT found in OpenRouter fallback models"
fi

# Test 2.5: meta-llama/llama-4-scout in fallback models
if grep -q 'meta-llama/llama-4-scout' "$VERIFIER_FALLBACK" 2>/dev/null; then
    pass "meta-llama/llama-4-scout in OpenRouter fallback models"
else
    fail "meta-llama/llama-4-scout NOT found in OpenRouter fallback models"
fi

# Test 2.6: deepseek/deepseek-r1 in fallback models
if grep -q 'deepseek/deepseek-r1' "$VERIFIER_FALLBACK" 2>/dev/null; then
    pass "deepseek/deepseek-r1 in OpenRouter fallback models"
else
    fail "deepseek/deepseek-r1 NOT found in OpenRouter fallback models"
fi

# Test 2.7: mistralai/mistral-large in fallback models
if grep -q 'mistralai/mistral-large' "$VERIFIER_FALLBACK" 2>/dev/null; then
    pass "mistralai/mistral-large in OpenRouter fallback models"
else
    fail "mistralai/mistral-large NOT found in OpenRouter fallback models"
fi

# Test 2.8: qwen/qwen3-235b-a22b in fallback models
if grep -q 'qwen/qwen3-235b-a22b' "$VERIFIER_FALLBACK" 2>/dev/null; then
    pass "qwen/qwen3-235b-a22b in OpenRouter fallback models"
else
    fail "qwen/qwen3-235b-a22b NOT found in OpenRouter fallback models"
fi

#===============================================================================
# Group 3: Features (4 tests)
#===============================================================================
section "Group 3: Features (4 tests)"

# Test 3.1: Streaming SSE support
if grep -rq 'text/event-stream\|data: ' "$OPENROUTER_DIR/" 2>/dev/null; then
    pass "Streaming SSE support in OpenRouter code"
else
    fail "Streaming SSE support NOT found in OpenRouter code"
fi

# Test 3.2: Retry with backoff
if grep -q 'Retry\|retry\|backoff\|Backoff' "$OPENROUTER_DIR/openrouter.go" 2>/dev/null; then
    pass "Retry mechanism in OpenRouter provider"
else
    fail "Retry mechanism NOT found in OpenRouter provider"
fi

# Test 3.3: Base URL constant defined
if grep -q 'defaultBaseURL\|openrouter\.ai/api/v1' "$OPENROUTER_DIR/openrouter.go" 2>/dev/null; then
    pass "Base URL constant defined for OpenRouter"
else
    fail "Base URL constant NOT found for OpenRouter"
fi

# Test 3.4: Constructor exists
if grep -q 'func NewSimpleOpenRouterProvider' "$OPENROUTER_DIR/openrouter.go" 2>/dev/null; then
    pass "NewSimpleOpenRouterProvider constructor exists"
else
    fail "NewSimpleOpenRouterProvider constructor NOT found"
fi

#===============================================================================
# Group 4: Integration (4 tests)
#===============================================================================
section "Group 4: Integration (4 tests)"

# Test 4.1: OPENROUTER_API_KEY in .env.example
if grep -q 'OPENROUTER_API_KEY' "$PROJECT_ROOT/.env.example" 2>/dev/null; then
    pass "OPENROUTER_API_KEY in .env.example"
else
    fail "OPENROUTER_API_KEY NOT found in .env.example"
fi

# Test 4.2: Unit tests exist
if ls "$OPENROUTER_DIR/"*_test.go 1>/dev/null 2>&1; then
    pass "Unit tests exist for OpenRouter"
else
    fail "Unit tests NOT found for OpenRouter"
fi

# Test 4.3: OpenRouter in verifier provider_types.go
if grep -q '"openrouter"' "$VERIFIER_DIR/provider_types.go" 2>/dev/null; then
    pass "OpenRouter in verifier provider_types.go"
else
    fail "OpenRouter NOT found in verifier provider_types.go"
fi

# Test 4.4: Integration test exists
if [ -f "$PROJECT_ROOT/tests/integration/openrouter_integration_test.go" ]; then
    pass "Integration test exists (openrouter_integration_test.go)"
else
    fail "Integration test NOT found (openrouter_integration_test.go)"
fi

#===============================================================================
# Group 5: Test Execution (4 tests)
#===============================================================================
section "Group 5: Test Execution (4 tests)"

# Test 5.1: go vet passes on openrouter package
echo -e "  ${YELLOW}Running go vet on openrouter package...${NC}"
if (cd "$PROJECT_ROOT" && nice -n 19 ionice -c 3 go vet ./internal/llm/providers/openrouter/ 2>&1); then
    pass "go vet passes on openrouter package"
else
    fail "go vet FAILED on openrouter package"
fi

# Test 5.2: OpenRouter package builds successfully
echo -e "  ${YELLOW}Building openrouter package...${NC}"
if (cd "$PROJECT_ROOT" && nice -n 19 ionice -c 3 go build ./internal/llm/providers/openrouter/ 2>&1); then
    pass "OpenRouter package builds successfully"
else
    fail "OpenRouter package build FAILED"
fi

# Test 5.3: go test -short passes on openrouter package
echo -e "  ${YELLOW}Running go test -short on openrouter package...${NC}"
TEST_OUTPUT=$(cd "$PROJECT_ROOT" && nice -n 19 ionice -c 3 go test -v -short -count=1 -p 1 -timeout 120s ./internal/llm/providers/openrouter/ 2>&1) || true
PASS_COUNT=$(echo "$TEST_OUTPUT" | grep -c "^--- PASS:" || echo "0")
PASS_COUNT=${PASS_COUNT//[^0-9]/}
PASS_COUNT=${PASS_COUNT:-0}
FAIL_COUNT=$(echo "$TEST_OUTPUT" | grep -c "^--- FAIL:" || echo "0")
FAIL_COUNT=${FAIL_COUNT//[^0-9]/}
FAIL_COUNT=${FAIL_COUNT:-0}
if echo "$TEST_OUTPUT" | grep -q "^ok" && [ "$FAIL_COUNT" -eq 0 ]; then
    pass "go test -short passes on openrouter package ($PASS_COUNT tests passed)"
else
    fail "go test -short FAILED on openrouter package ($PASS_COUNT passed, $FAIL_COUNT failed)"
fi

# Test 5.4: helixagent binary builds successfully
echo -e "  ${YELLOW}Building helixagent binary...${NC}"
if (cd "$PROJECT_ROOT" && nice -n 19 ionice -c 3 go build ./cmd/helixagent/ 2>&1); then
    pass "helixagent binary builds successfully"
else
    fail "helixagent binary build FAILED"
fi

#===============================================================================
# Group 6: LLMsVerifier Integration (4 tests)
#===============================================================================
section "Group 6: LLMsVerifier Integration (4 tests)"

# Test 6.1: OpenRouter in LLMsVerifier fallback models
if grep -q '"openrouter"' "$VERIFIER_FALLBACK" 2>/dev/null; then
    pass "OpenRouter in LLMsVerifier fallback models"
else
    fail "OpenRouter NOT found in LLMsVerifier fallback models"
fi

# Test 6.2: At least 5 OpenRouter models in fallback
OR_MODEL_COUNT=$(grep -A 50 '"openrouter":' "$VERIFIER_FALLBACK" 2>/dev/null | grep -c 'ProviderID.*openrouter' || echo "0")
OR_MODEL_COUNT=${OR_MODEL_COUNT//[^0-9]/}
OR_MODEL_COUNT=${OR_MODEL_COUNT:-0}
if [ "$OR_MODEL_COUNT" -ge 5 ]; then
    pass "At least 5 OpenRouter models in LLMsVerifier fallback (found $OR_MODEL_COUNT)"
else
    fail "Less than 5 OpenRouter models in LLMsVerifier fallback (found $OR_MODEL_COUNT)"
fi

# Test 6.3: OpenRouter in verifier discovery.go
if grep -q 'openrouter' "$VERIFIER_DIR/discovery.go" 2>/dev/null; then
    pass "OpenRouter in verifier discovery.go"
else
    fail "OpenRouter NOT found in verifier discovery.go"
fi

# Test 6.4: OpenRouter in verifier health.go
if grep -q 'openrouter' "$VERIFIER_DIR/health.go" 2>/dev/null; then
    pass "OpenRouter in verifier health.go"
else
    fail "OpenRouter NOT found in verifier health.go"
fi

#===============================================================================
# Summary
#===============================================================================
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  OpenRouter Comprehensive Challenge${NC}"
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
