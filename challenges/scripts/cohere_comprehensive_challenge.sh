#!/bin/bash
# Cohere Comprehensive Verification Challenge
# VALIDATES: Cohere provider structure, model coverage, features,
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

COHERE_DIR="$PROJECT_ROOT/internal/llm/providers/cohere"
VERIFIER_DIR="$PROJECT_ROOT/internal/verifier"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Cohere Comprehensive Challenge${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "  Validates: Provider structure, models, features,"
echo -e "  integration, test execution, LLMsVerifier integration"

#===============================================================================
# Group 1: Provider Structure (8 tests)
#===============================================================================
section "Group 1: Provider Structure (8 tests)"

# Test 1.1: Provider struct exists (NewProvider or similar in cohere.go)
if grep -q 'func NewProvider' "$COHERE_DIR/cohere.go" 2>/dev/null; then
    pass "NewProvider constructor exists in cohere.go"
else
    fail "NewProvider constructor NOT found in cohere.go"
fi

# Test 1.2: Complete method exists
if grep -q 'func (p \*Provider) Complete' "$COHERE_DIR/cohere.go" 2>/dev/null; then
    pass "Complete method exists on Provider"
else
    fail "Complete method NOT found on Provider"
fi

# Test 1.3: CompleteStream method exists
if grep -q 'func (p \*Provider) CompleteStream' "$COHERE_DIR/cohere.go" 2>/dev/null; then
    pass "CompleteStream method exists on Provider"
else
    fail "CompleteStream method NOT found on Provider"
fi

# Test 1.4: HealthCheck method exists
if grep -q 'func (p \*Provider) HealthCheck' "$COHERE_DIR/cohere.go" 2>/dev/null; then
    pass "HealthCheck method exists on Provider"
else
    fail "HealthCheck method NOT found on Provider"
fi

# Test 1.5: GetCapabilities method exists
if grep -q 'func (p \*Provider) GetCapabilities' "$COHERE_DIR/cohere.go" 2>/dev/null; then
    pass "GetCapabilities method exists on Provider"
else
    fail "GetCapabilities method NOT found on Provider"
fi

# Test 1.6: ValidateConfig method exists
if grep -q 'func (p \*Provider) ValidateConfig' "$COHERE_DIR/cohere.go" 2>/dev/null; then
    pass "ValidateConfig method exists on Provider"
else
    fail "ValidateConfig method NOT found on Provider"
fi

# Test 1.7: Registered in provider_registry.go
if grep -q 'case "cohere":' "$PROJECT_ROOT/internal/services/provider_registry.go" 2>/dev/null; then
    pass "Cohere registered in provider_registry.go"
else
    fail "Cohere NOT registered in provider_registry.go"
fi

# Test 1.8: Uses v2 API endpoint (api.cohere.com/v2 in code)
if grep -q 'api\.cohere\.com/v2' "$COHERE_DIR/cohere.go" 2>/dev/null; then
    pass "Uses v2 API endpoint (api.cohere.com/v2)"
else
    fail "v2 API endpoint (api.cohere.com/v2) NOT found in cohere.go"
fi

#===============================================================================
# Group 2: Model Coverage (8 tests)
#===============================================================================
section "Group 2: Model Coverage (8 tests)"

# Models are checked across both provider code and LLMsVerifier fallback
FALLBACK="$PROJECT_ROOT/LLMsVerifier/llm-verifier/providers/fallback_models.go"

# Test 2.1: command-a-03-2025 in models
if grep -rq 'command-a-03-2025' "$COHERE_DIR/" 2>/dev/null; then
    pass "command-a-03-2025 in provider models"
else
    fail "command-a-03-2025 NOT found in provider models"
fi

# Test 2.2: command-a-reasoning in models
if grep -rq 'command-a-reasoning' "$COHERE_DIR/" 2>/dev/null; then
    pass "command-a-reasoning in provider models"
else
    fail "command-a-reasoning NOT found in provider models"
fi

# Test 2.3: command-a-vision in models
if grep -rq 'command-a-vision' "$COHERE_DIR/" 2>/dev/null; then
    pass "command-a-vision in provider models"
else
    fail "command-a-vision NOT found in provider models"
fi

# Test 2.4: command-r7b-12-2024 in models
if grep -rq 'command-r7b-12-2024' "$COHERE_DIR/" 2>/dev/null; then
    pass "command-r7b-12-2024 in provider models"
else
    fail "command-r7b-12-2024 NOT found in provider models"
fi

# Test 2.5: command-r-plus in models
if grep -rq 'command-r-plus' "$COHERE_DIR/" 2>/dev/null; then
    pass "command-r-plus in provider models"
else
    fail "command-r-plus NOT found in provider models"
fi

# Test 2.6: command-r-08-2024 in models
if grep -rq 'command-r-08-2024' "$COHERE_DIR/" 2>/dev/null; then
    pass "command-r-08-2024 in provider models"
else
    fail "command-r-08-2024 NOT found in provider models"
fi

# Test 2.7: c4ai-aya-expanse-32b in models
if grep -rq 'c4ai-aya-expanse-32b' "$COHERE_DIR/" 2>/dev/null; then
    pass "c4ai-aya-expanse-32b in provider models"
else
    fail "c4ai-aya-expanse-32b NOT found in provider models"
fi

# Test 2.8: c4ai-aya-vision-32b in models
if grep -rq 'c4ai-aya-vision-32b' "$COHERE_DIR/" 2>/dev/null; then
    pass "c4ai-aya-vision-32b in provider models"
else
    fail "c4ai-aya-vision-32b NOT found in provider models"
fi

#===============================================================================
# Group 3: Features (6 tests)
#===============================================================================
section "Group 3: Features (6 tests)"

# Test 3.1: Tool calling support (tools field in code)
if grep -q 'Tools' "$COHERE_DIR/cohere.go" 2>/dev/null && grep -q '"tools"' "$COHERE_DIR/cohere.go" 2>/dev/null; then
    pass "Tool calling support (tools field in code)"
else
    fail "Tool calling support (tools field) NOT found in cohere.go"
fi

# Test 3.2: Streaming support (SSE/stream in code)
if grep -q 'Stream' "$COHERE_DIR/cohere.go" 2>/dev/null && grep -q 'data: ' "$COHERE_DIR/cohere.go" 2>/dev/null; then
    pass "Streaming support (SSE data: prefix parsing)"
else
    fail "Streaming support (SSE data: prefix) NOT found in cohere.go"
fi

# Test 3.3: RAG/documents support (documents in code)
if grep -q 'Documents' "$COHERE_DIR/cohere.go" 2>/dev/null || grep -q 'documents' "$COHERE_DIR/cohere.go" 2>/dev/null; then
    pass "RAG/documents support in code"
else
    fail "RAG/documents support NOT found in cohere.go"
fi

# Test 3.4: Citation support (citation in code)
if grep -qi 'citation' "$COHERE_DIR/cohere.go" 2>/dev/null; then
    pass "Citation support in code"
else
    fail "Citation support NOT found in cohere.go"
fi

# Test 3.5: Safety mode support (safety_mode in code)
if grep -q 'safety_mode' "$COHERE_DIR/cohere.go" 2>/dev/null || grep -q 'SafetyMode' "$COHERE_DIR/cohere.go" 2>/dev/null; then
    pass "Safety mode support in code"
else
    fail "Safety mode support NOT found in cohere.go"
fi

# Test 3.6: Bearer token authentication
if grep -q 'Bearer' "$COHERE_DIR/cohere.go" 2>/dev/null; then
    pass "Bearer token authentication"
else
    fail "Bearer token authentication NOT found in cohere.go"
fi

#===============================================================================
# Group 4: Integration (5 tests)
#===============================================================================
section "Group 4: Integration (5 tests)"

# Test 4.1: COHERE_API_KEY in .env.example
if grep -q 'COHERE_API_KEY' "$PROJECT_ROOT/.env.example" 2>/dev/null; then
    pass "COHERE_API_KEY in .env.example"
else
    fail "COHERE_API_KEY NOT found in .env.example"
fi

# Test 4.2: Unit tests exist (cohere_test.go)
if [ -f "$COHERE_DIR/cohere_test.go" ]; then
    pass "Unit tests exist (cohere_test.go)"
else
    fail "Unit tests NOT found (cohere_test.go)"
fi

# Test 4.3: go vet passes on cohere package
echo -e "  ${YELLOW}Running go vet on cohere package...${NC}"
if (cd "$PROJECT_ROOT" && nice -n 19 ionice -c 3 go vet ./internal/llm/providers/cohere/ 2>&1); then
    pass "go vet passes on cohere package"
else
    fail "go vet FAILED on cohere package"
fi

# Test 4.4: Cohere in verifier provider_types.go
if grep -q '"cohere"' "$VERIFIER_DIR/provider_types.go" 2>/dev/null; then
    pass "Cohere in verifier provider_types.go"
else
    fail "Cohere NOT found in verifier provider_types.go"
fi

# Test 4.5: Cohere in discovery (ParseCohereModelsResponse)
if grep -q 'Cohere' "$PROJECT_ROOT/internal/llm/discovery/discovery.go" 2>/dev/null; then
    pass "Cohere in discovery.go (ParseCohereModelsResponse)"
else
    fail "Cohere NOT found in discovery.go"
fi

#===============================================================================
# Group 5: Test Execution (4 tests)
#===============================================================================
section "Group 5: Test Execution (4 tests)"

# Test 5.1: go test -short passes on cohere package
echo -e "  ${YELLOW}Running go test -short on cohere package...${NC}"
TEST_OUTPUT=$(cd "$PROJECT_ROOT" && nice -n 19 ionice -c 3 go test -v -short -count=1 -p 1 -timeout 120s ./internal/llm/providers/cohere/ 2>&1) || true
PASS_COUNT=$(echo "$TEST_OUTPUT" | grep -c "^--- PASS:" || echo "0")
PASS_COUNT=${PASS_COUNT//[^0-9]/}
PASS_COUNT=${PASS_COUNT:-0}
FAIL_COUNT=$(echo "$TEST_OUTPUT" | grep -c "^--- FAIL:" || echo "0")
FAIL_COUNT=${FAIL_COUNT//[^0-9]/}
FAIL_COUNT=${FAIL_COUNT:-0}
if echo "$TEST_OUTPUT" | grep -q "^ok" && [ "$FAIL_COUNT" -eq 0 ]; then
    pass "go test -short passes on cohere package ($PASS_COUNT tests passed)"
else
    fail "go test -short FAILED on cohere package ($PASS_COUNT passed, $FAIL_COUNT failed)"
fi

# Test 5.2: go vet clean
echo -e "  ${YELLOW}Running go vet verification...${NC}"
VET_OUTPUT=$(cd "$PROJECT_ROOT" && nice -n 19 ionice -c 3 go vet ./internal/llm/providers/cohere/ 2>&1) || true
if [ -z "$VET_OUTPUT" ]; then
    pass "go vet produces clean output (no warnings)"
else
    fail "go vet produced warnings: $VET_OUTPUT"
fi

# Test 5.3: cohere package builds successfully
echo -e "  ${YELLOW}Building cohere package...${NC}"
if (cd "$PROJECT_ROOT" && nice -n 19 ionice -c 3 go build ./internal/llm/providers/cohere/ 2>&1); then
    pass "cohere package builds successfully"
else
    fail "cohere package build FAILED"
fi

# Test 5.4: helixagent binary builds successfully
echo -e "  ${YELLOW}Building helixagent binary...${NC}"
if (cd "$PROJECT_ROOT" && nice -n 19 ionice -c 3 go build ./cmd/helixagent/ 2>&1); then
    pass "helixagent binary builds successfully"
else
    fail "helixagent binary build FAILED"
fi

#===============================================================================
# Group 6: LLMsVerifier Integration (3 tests)
#===============================================================================
section "Group 6: LLMsVerifier Integration (3 tests)"

# Test 6.1: Cohere in LLMsVerifier fallback models
VERIFIER_FALLBACK="$PROJECT_ROOT/LLMsVerifier/llm-verifier/providers/fallback_models.go"
if grep -q '"cohere"' "$VERIFIER_FALLBACK" 2>/dev/null; then
    pass "Cohere in LLMsVerifier fallback models"
else
    fail "Cohere NOT found in LLMsVerifier fallback models"
fi

# Test 6.2: At least 4 Cohere models in fallback
# Count model entries within the cohere section of fallback_models.go
CO_MODEL_COUNT=$(awk '/"cohere":/{found=1} found && /ProviderID.*cohere/{count++} found && /^\t\},/{exit} END{print count+0}' "$VERIFIER_FALLBACK" 2>/dev/null)
CO_MODEL_COUNT=${CO_MODEL_COUNT//[^0-9]/}
CO_MODEL_COUNT=${CO_MODEL_COUNT:-0}
if [ "$CO_MODEL_COUNT" -ge 4 ]; then
    pass "At least 4 Cohere models in LLMsVerifier fallback (found $CO_MODEL_COUNT)"
else
    # Fallback counting method
    CO_MODEL_COUNT2=$(grep -A 30 '"cohere":' "$VERIFIER_FALLBACK" 2>/dev/null | grep -c 'ProviderID.*cohere' || echo "0")
    CO_MODEL_COUNT2=${CO_MODEL_COUNT2//[^0-9]/}
    CO_MODEL_COUNT2=${CO_MODEL_COUNT2:-0}
    if [ "$CO_MODEL_COUNT2" -ge 4 ]; then
        pass "At least 4 Cohere models in LLMsVerifier fallback (found $CO_MODEL_COUNT2)"
    else
        fail "Less than 4 Cohere models in LLMsVerifier fallback (found $CO_MODEL_COUNT2)"
    fi
fi

# Test 6.3: Cohere provider config exists in LLMsVerifier (cohere.go in providers)
VERIFIER_COHERE="$PROJECT_ROOT/LLMsVerifier/llm-verifier/providers/cohere.go"
if [ -f "$VERIFIER_COHERE" ]; then
    pass "Cohere provider config exists in LLMsVerifier (cohere.go)"
else
    fail "Cohere provider config NOT found in LLMsVerifier (cohere.go)"
fi

#===============================================================================
# Summary
#===============================================================================
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Cohere Comprehensive Challenge${NC}"
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
