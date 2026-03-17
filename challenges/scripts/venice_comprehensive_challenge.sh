#!/bin/bash
# Venice AI Comprehensive Verification Challenge
# VALIDATES: Venice provider structure, model coverage, features,
#            integration, test execution, and LLMsVerifier integration
# Total: ~34 tests across 6 groups

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

VENICE_DIR="$PROJECT_ROOT/internal/llm/providers/venice"
VERIFIER_DIR="$PROJECT_ROOT/internal/verifier"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Venice AI Comprehensive Challenge${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "  Validates: Provider structure, models, features,"
echo -e "  integration, test execution, LLMsVerifier integration"

#===============================================================================
# Group 1: Provider Structure (8 tests)
#===============================================================================
section "Group 1: Provider Structure (8 tests)"

# Test 1.1: Provider struct exists in venice.go
if grep -q 'type Provider struct' "$VENICE_DIR/venice.go" 2>/dev/null; then
    pass "Provider struct exists in venice.go"
else
    fail "Provider struct NOT found in venice.go"
fi

# Test 1.2: Complete method exists
if grep -q 'func (p \*Provider) Complete' "$VENICE_DIR/venice.go" 2>/dev/null; then
    pass "Complete method exists on Provider"
else
    fail "Complete method NOT found on Provider"
fi

# Test 1.3: CompleteStream method exists
if grep -q 'func (p \*Provider) CompleteStream' "$VENICE_DIR/venice.go" 2>/dev/null; then
    pass "CompleteStream method exists on Provider"
else
    fail "CompleteStream method NOT found on Provider"
fi

# Test 1.4: HealthCheck method exists
if grep -q 'func (p \*Provider) HealthCheck' "$VENICE_DIR/venice.go" 2>/dev/null; then
    pass "HealthCheck method exists on Provider"
else
    fail "HealthCheck method NOT found on Provider"
fi

# Test 1.5: GetCapabilities method exists
if grep -q 'func (p \*Provider) GetCapabilities' "$VENICE_DIR/venice.go" 2>/dev/null; then
    pass "GetCapabilities method exists on Provider"
else
    fail "GetCapabilities method NOT found on Provider"
fi

# Test 1.6: ValidateConfig method exists
if grep -q 'func (p \*Provider) ValidateConfig' "$VENICE_DIR/venice.go" 2>/dev/null; then
    pass "ValidateConfig method exists on Provider"
else
    fail "ValidateConfig method NOT found on Provider"
fi

# Test 1.7: Registered in provider_registry.go
if grep -q 'case "venice":' "$PROJECT_ROOT/internal/services/provider_registry.go" 2>/dev/null; then
    pass "Venice registered in provider_registry.go"
else
    fail "Venice NOT registered in provider_registry.go"
fi

# Test 1.8: Provider name is "venice"
if grep -q 'return "venice"' "$VENICE_DIR/venice.go" 2>/dev/null; then
    pass "Provider name is 'venice'"
else
    fail "Provider name 'venice' NOT found in venice.go"
fi

#===============================================================================
# Group 2: Model Coverage (7 tests)
#===============================================================================
section "Group 2: Model Coverage (7 tests)"

# Test 2.1: llama-3.3-70b in known models
if grep -rq 'llama-3\.3-70b' "$VENICE_DIR/" 2>/dev/null; then
    pass "llama-3.3-70b in known models"
else
    fail "llama-3.3-70b NOT found in known models"
fi

# Test 2.2: llama-3.1-405b in known models
if grep -rq 'llama-3\.1-405b' "$VENICE_DIR/" 2>/dev/null; then
    pass "llama-3.1-405b in known models"
else
    fail "llama-3.1-405b NOT found in known models"
fi

# Test 2.3: deepseek-r1-671b in known models
if grep -rq 'deepseek-r1-671b' "$VENICE_DIR/" 2>/dev/null; then
    pass "deepseek-r1-671b in known models"
else
    fail "deepseek-r1-671b NOT found in known models"
fi

# Test 2.4: venice-uncensored in known models
if grep -rq 'venice-uncensored' "$VENICE_DIR/" 2>/dev/null; then
    pass "venice-uncensored in known models"
else
    fail "venice-uncensored NOT found in known models"
fi

# Test 2.5: qwen3-vl-235b-a22b in known models
if grep -rq 'qwen3-vl-235b-a22b' "$VENICE_DIR/" 2>/dev/null; then
    pass "qwen3-vl-235b-a22b in known models"
else
    fail "qwen3-vl-235b-a22b NOT found in known models"
fi

# Test 2.6: qwen-2.5-vl in known models
if grep -rq 'qwen-2\.5-vl' "$VENICE_DIR/" 2>/dev/null; then
    pass "qwen-2.5-vl in known models"
else
    fail "qwen-2.5-vl NOT found in known models"
fi

# Test 2.7: zai-org-glm-4.7 in known models
if grep -rq 'zai-org-glm-4\.7' "$VENICE_DIR/" 2>/dev/null; then
    pass "zai-org-glm-4.7 in known models"
else
    fail "zai-org-glm-4.7 NOT found in known models"
fi

#===============================================================================
# Group 3: Features (6 tests)
#===============================================================================
section "Group 3: Features (6 tests)"

# Test 3.1: Tool/function calling support
if grep -q 'Tools.*\[\]Tool' "$VENICE_DIR/venice.go" 2>/dev/null; then
    pass "Tool/function calling support (Tools field in request struct)"
else
    fail "Tool/function calling support NOT found in venice.go"
fi

# Test 3.2: Streaming SSE support (text/event-stream in code)
if grep -rq 'text/event-stream' "$VENICE_DIR/" 2>/dev/null; then
    pass "Streaming SSE support (text/event-stream in code)"
else
    if grep -q 'data: ' "$VENICE_DIR/venice.go" 2>/dev/null; then
        pass "Streaming SSE support (data: prefix parsing in code)"
    else
        fail "Streaming SSE support NOT found in venice code"
    fi
fi

# Test 3.3: Retry with exponential backoff
if grep -q 'calculateBackoff' "$VENICE_DIR/venice.go" 2>/dev/null; then
    pass "Retry with exponential backoff (calculateBackoff)"
else
    fail "Retry with exponential backoff NOT found in venice.go"
fi

# Test 3.4: Confidence calculation exists
if grep -q 'calculateConfidence' "$VENICE_DIR/venice.go" 2>/dev/null; then
    pass "Confidence calculation exists (calculateConfidence)"
else
    fail "Confidence calculation NOT found in venice.go"
fi

# Test 3.5: Base URL constant defined
if grep -q 'VeniceAPIURL' "$VENICE_DIR/venice.go" 2>/dev/null; then
    pass "VeniceAPIURL constant defined"
else
    fail "VeniceAPIURL constant NOT found in venice.go"
fi

# Test 3.6: Constructor exists
if grep -q 'func NewProvider' "$VENICE_DIR/venice.go" 2>/dev/null; then
    pass "NewProvider constructor exists"
else
    fail "NewProvider constructor NOT found in venice.go"
fi

#===============================================================================
# Group 4: Integration (5 tests)
#===============================================================================
section "Group 4: Integration (5 tests)"

# Test 4.1: VENICE_API_KEY in .env.example
if grep -q 'VENICE_API_KEY' "$PROJECT_ROOT/.env.example" 2>/dev/null; then
    pass "VENICE_API_KEY in .env.example"
else
    fail "VENICE_API_KEY NOT found in .env.example"
fi

# Test 4.2: Unit tests exist (venice_test.go)
if [ -f "$VENICE_DIR/venice_test.go" ]; then
    pass "Unit tests exist (venice_test.go)"
else
    fail "Unit tests NOT found (venice_test.go)"
fi

# Test 4.3: Benchmark tests exist
if grep -q 'func Benchmark' "$VENICE_DIR/venice_test.go" 2>/dev/null; then
    pass "Benchmark tests exist in venice_test.go"
else
    fail "Benchmark tests NOT found in venice_test.go"
fi

# Test 4.4: Venice in verifier provider_types.go
if grep -q '"venice"' "$VERIFIER_DIR/provider_types.go" 2>/dev/null; then
    pass "Venice in verifier provider_types.go"
else
    fail "Venice NOT found in verifier provider_types.go"
fi

# Test 4.5: Venice in provider_discovery.go
if grep -q 'venice' "$PROJECT_ROOT/internal/services/provider_discovery.go" 2>/dev/null; then
    pass "Venice in provider_discovery.go"
else
    fail "Venice NOT found in provider_discovery.go"
fi

#===============================================================================
# Group 5: Test Execution (4 tests)
#===============================================================================
section "Group 5: Test Execution (4 tests)"

# Test 5.1: go test -short passes on venice package
echo -e "  ${YELLOW}Running go test -short on venice package...${NC}"
TEST_OUTPUT=$(cd "$PROJECT_ROOT" && nice -n 19 ionice -c 3 go test -v -short -count=1 -p 1 -timeout 120s ./internal/llm/providers/venice/ 2>&1) || true
PASS_COUNT=$(echo "$TEST_OUTPUT" | grep -c "^--- PASS:" || echo "0")
PASS_COUNT=${PASS_COUNT//[^0-9]/}
PASS_COUNT=${PASS_COUNT:-0}
FAIL_COUNT=$(echo "$TEST_OUTPUT" | grep -c "^--- FAIL:" || echo "0")
FAIL_COUNT=${FAIL_COUNT//[^0-9]/}
FAIL_COUNT=${FAIL_COUNT:-0}
if echo "$TEST_OUTPUT" | grep -q "^ok" && [ "$FAIL_COUNT" -eq 0 ]; then
    pass "go test -short passes on venice package ($PASS_COUNT tests passed)"
else
    fail "go test -short FAILED on venice package ($PASS_COUNT passed, $FAIL_COUNT failed)"
fi

# Test 5.2: go vet clean
echo -e "  ${YELLOW}Running go vet verification...${NC}"
VET_OUTPUT=$(cd "$PROJECT_ROOT" && nice -n 19 ionice -c 3 go vet ./internal/llm/providers/venice/ 2>&1) || true
if [ -z "$VET_OUTPUT" ]; then
    pass "go vet produces clean output (no warnings)"
else
    fail "go vet produced warnings: $VET_OUTPUT"
fi

# Test 5.3: Venice package builds successfully
echo -e "  ${YELLOW}Building venice package...${NC}"
if (cd "$PROJECT_ROOT" && nice -n 19 ionice -c 3 go build ./internal/llm/providers/venice/ 2>&1); then
    pass "Venice package builds successfully"
else
    fail "Venice package build FAILED"
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

# Test 6.1: Venice in LLMsVerifier fallback models
VERIFIER_FALLBACK="$PROJECT_ROOT/LLMsVerifier/llm-verifier/providers/fallback_models.go"
if grep -q '"venice"' "$VERIFIER_FALLBACK" 2>/dev/null; then
    pass "Venice in LLMsVerifier fallback models"
else
    fail "Venice NOT found in LLMsVerifier fallback models"
fi

# Test 6.2: At least 5 Venice models in fallback
VENICE_MODEL_COUNT=$(grep -A 50 '"venice":' "$VERIFIER_FALLBACK" 2>/dev/null | grep -c 'ProviderID.*venice' || echo "0")
VENICE_MODEL_COUNT=${VENICE_MODEL_COUNT//[^0-9]/}
VENICE_MODEL_COUNT=${VENICE_MODEL_COUNT:-0}
if [ "$VENICE_MODEL_COUNT" -ge 5 ]; then
    pass "At least 5 Venice models in LLMsVerifier fallback (found $VENICE_MODEL_COUNT)"
else
    fail "Less than 5 Venice models in LLMsVerifier fallback (found $VENICE_MODEL_COUNT)"
fi

# Test 6.3: Venice in verifier discovery.go
if grep -q 'venice' "$VERIFIER_DIR/discovery.go" 2>/dev/null; then
    pass "Venice in verifier discovery.go"
else
    fail "Venice NOT found in verifier discovery.go"
fi

# Test 6.4: Venice in verifier health.go
if grep -q 'venice' "$VERIFIER_DIR/health.go" 2>/dev/null; then
    pass "Venice in verifier health.go"
else
    fail "Venice NOT found in verifier health.go"
fi

#===============================================================================
# Summary
#===============================================================================
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Venice AI Comprehensive Challenge${NC}"
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
