#!/bin/bash
# Groq Comprehensive Verification Challenge
# VALIDATES: Groq provider structure, model coverage, features,
#            integration, test execution, and LLMsVerifier integration
# Total: ~36 tests across 6 groups

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

GROQ_DIR="$PROJECT_ROOT/internal/llm/providers/groq"
VERIFIER_DIR="$PROJECT_ROOT/internal/verifier"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Groq Comprehensive Challenge${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "  Validates: Provider structure, models, features,"
echo -e "  integration, test execution, LLMsVerifier integration"

#===============================================================================
# Group 1: Provider Structure (8 tests)
#===============================================================================
section "Group 1: Provider Structure (8 tests)"

# Test 1.1: Provider struct exists in groq.go
if grep -q 'type Provider struct' "$GROQ_DIR/groq.go" 2>/dev/null; then
    pass "Provider struct exists in groq.go"
else
    fail "Provider struct NOT found in groq.go"
fi

# Test 1.2: Complete method exists
if grep -q 'func (p \*Provider) Complete' "$GROQ_DIR/groq.go" 2>/dev/null; then
    pass "Complete method exists on Provider"
else
    fail "Complete method NOT found on Provider"
fi

# Test 1.3: CompleteStream method exists
if grep -q 'func (p \*Provider) CompleteStream' "$GROQ_DIR/groq.go" 2>/dev/null; then
    pass "CompleteStream method exists on Provider"
else
    fail "CompleteStream method NOT found on Provider"
fi

# Test 1.4: HealthCheck method exists
if grep -q 'func (p \*Provider) HealthCheck' "$GROQ_DIR/groq.go" 2>/dev/null; then
    pass "HealthCheck method exists on Provider"
else
    fail "HealthCheck method NOT found on Provider"
fi

# Test 1.5: GetCapabilities method exists
if grep -q 'func (p \*Provider) GetCapabilities' "$GROQ_DIR/groq.go" 2>/dev/null; then
    pass "GetCapabilities method exists on Provider"
else
    fail "GetCapabilities method NOT found on Provider"
fi

# Test 1.6: ValidateConfig method exists
if grep -q 'func (p \*Provider) ValidateConfig' "$GROQ_DIR/groq.go" 2>/dev/null; then
    pass "ValidateConfig method exists on Provider"
else
    fail "ValidateConfig method NOT found on Provider"
fi

# Test 1.7: Registered in provider_registry.go
if grep -q 'case "groq":' "$PROJECT_ROOT/internal/services/provider_registry.go" 2>/dev/null; then
    pass "Groq registered in provider_registry.go"
else
    fail "Groq NOT registered in provider_registry.go"
fi

# Test 1.8: Provider name is "groq"
if grep -q 'return "groq"' "$GROQ_DIR/groq.go" 2>/dev/null; then
    pass "Provider name is 'groq'"
else
    fail "Provider name 'groq' NOT found in groq.go"
fi

#===============================================================================
# Group 2: Model Coverage (8 tests)
#===============================================================================
section "Group 2: Model Coverage (8 tests)"

# Test 2.1: llama-4-maverick in known models
if grep -rq 'llama-4-maverick' "$GROQ_DIR/" 2>/dev/null; then
    pass "llama-4-maverick in known models"
else
    fail "llama-4-maverick NOT found in known models"
fi

# Test 2.2: llama-4-scout in known models
if grep -rq 'llama-4-scout' "$GROQ_DIR/" 2>/dev/null; then
    pass "llama-4-scout in known models"
else
    fail "llama-4-scout NOT found in known models"
fi

# Test 2.3: llama-3.3-70b-versatile in known models
if grep -rq 'llama-3\.3-70b-versatile' "$GROQ_DIR/" 2>/dev/null; then
    pass "llama-3.3-70b-versatile in known models"
else
    fail "llama-3.3-70b-versatile NOT found in known models"
fi

# Test 2.4: llama-3.1-70b-versatile in known models
if grep -rq 'llama-3\.1-70b-versatile' "$GROQ_DIR/" 2>/dev/null; then
    pass "llama-3.1-70b-versatile in known models"
else
    fail "llama-3.1-70b-versatile NOT found in known models"
fi

# Test 2.5: llama-3.1-8b-instant in known models
if grep -rq 'llama-3\.1-8b-instant' "$GROQ_DIR/" 2>/dev/null; then
    pass "llama-3.1-8b-instant in known models"
else
    fail "llama-3.1-8b-instant NOT found in known models"
fi

# Test 2.6: deepseek-r1-distill-llama-70b in known models
if grep -rq 'deepseek-r1-distill-llama-70b' "$GROQ_DIR/" 2>/dev/null; then
    pass "deepseek-r1-distill-llama-70b in known models"
else
    fail "deepseek-r1-distill-llama-70b NOT found in known models"
fi

# Test 2.7: mixtral-8x7b-32768 in known models
if grep -rq 'mixtral-8x7b-32768' "$GROQ_DIR/" 2>/dev/null; then
    pass "mixtral-8x7b-32768 in known models"
else
    fail "mixtral-8x7b-32768 NOT found in known models"
fi

# Test 2.8: gemma2-9b-it in known models
if grep -rq 'gemma2-9b-it' "$GROQ_DIR/" 2>/dev/null; then
    pass "gemma2-9b-it in known models"
else
    fail "gemma2-9b-it NOT found in known models"
fi

#===============================================================================
# Group 3: Features (6 tests)
#===============================================================================
section "Group 3: Features (6 tests)"

# Test 3.1: Tool/function calling support (tools field in code)
if grep -q 'Tools.*\[\]Tool' "$GROQ_DIR/groq.go" 2>/dev/null; then
    pass "Tool/function calling support (Tools field in request struct)"
else
    fail "Tool/function calling support NOT found in groq.go"
fi

# Test 3.2: Streaming SSE support (text/event-stream in code)
if grep -rq 'text/event-stream' "$GROQ_DIR/" 2>/dev/null; then
    pass "Streaming SSE support (text/event-stream in code)"
else
    # Also check for SSE data: prefix parsing which indicates SSE support
    if grep -q 'data: ' "$GROQ_DIR/groq.go" 2>/dev/null; then
        pass "Streaming SSE support (data: prefix parsing in code)"
    else
        fail "Streaming SSE support NOT found in groq code"
    fi
fi

# Test 3.3: Retry with exponential backoff
if grep -q 'calculateBackoff' "$GROQ_DIR/groq.go" 2>/dev/null; then
    pass "Retry with exponential backoff (calculateBackoff)"
else
    fail "Retry with exponential backoff NOT found in groq.go"
fi

# Test 3.4: Confidence calculation exists
if grep -q 'calculateConfidence' "$GROQ_DIR/groq.go" 2>/dev/null; then
    pass "Confidence calculation exists (calculateConfidence)"
else
    fail "Confidence calculation NOT found in groq.go"
fi

# Test 3.5: API key validation (gsk_ prefix check)
if grep -q 'gsk_' "$GROQ_DIR/groq.go" 2>/dev/null; then
    pass "API key validation (gsk_ prefix check)"
else
    fail "API key validation (gsk_ prefix) NOT found in groq.go"
fi

# Test 3.6: Groq-specific metadata extraction (prompt_time, completion_time)
if grep -q 'prompt_time' "$GROQ_DIR/groq.go" 2>/dev/null && grep -q 'completion_time' "$GROQ_DIR/groq.go" 2>/dev/null; then
    pass "Groq-specific metadata extraction (prompt_time, completion_time)"
else
    fail "Groq-specific metadata extraction NOT found in groq.go"
fi

#===============================================================================
# Group 4: Integration (6 tests)
#===============================================================================
section "Group 4: Integration (6 tests)"

# Test 4.1: GROQ_API_KEY in .env.example
if grep -q 'GROQ_API_KEY' "$PROJECT_ROOT/.env.example" 2>/dev/null; then
    pass "GROQ_API_KEY in .env.example"
else
    fail "GROQ_API_KEY NOT found in .env.example"
fi

# Test 4.2: Unit tests exist (groq_test.go)
if [ -f "$GROQ_DIR/groq_test.go" ]; then
    pass "Unit tests exist (groq_test.go)"
else
    fail "Unit tests NOT found (groq_test.go)"
fi

# Test 4.3: Benchmark tests exist
if grep -q 'func Benchmark' "$GROQ_DIR/groq_test.go" 2>/dev/null; then
    pass "Benchmark tests exist in groq_test.go"
else
    fail "Benchmark tests NOT found in groq_test.go"
fi

# Test 4.4: go vet passes on groq package
echo -e "  ${YELLOW}Running go vet on groq package...${NC}"
if (cd "$PROJECT_ROOT" && nice -n 19 ionice -c 3 go vet ./internal/llm/providers/groq/ 2>&1); then
    pass "go vet passes on groq package"
else
    fail "go vet FAILED on groq package"
fi

# Test 4.5: Groq in verifier provider_types.go
if grep -q '"groq"' "$VERIFIER_DIR/provider_types.go" 2>/dev/null; then
    pass "Groq in verifier provider_types.go"
else
    fail "Groq NOT found in verifier provider_types.go"
fi

# Test 4.6: Groq in debate_team_config.go
if grep -q 'groq' "$PROJECT_ROOT/internal/services/debate_team_config.go" 2>/dev/null; then
    pass "Groq in debate_team_config.go"
else
    fail "Groq NOT found in debate_team_config.go"
fi

#===============================================================================
# Group 5: Test Execution (4 tests)
#===============================================================================
section "Group 5: Test Execution (4 tests)"

# Test 5.1: go test -short passes on groq package
echo -e "  ${YELLOW}Running go test -short on groq package...${NC}"
TEST_OUTPUT=$(cd "$PROJECT_ROOT" && nice -n 19 ionice -c 3 go test -v -short -count=1 -p 1 -timeout 120s ./internal/llm/providers/groq/ 2>&1) || true
PASS_COUNT=$(echo "$TEST_OUTPUT" | grep -c "^--- PASS:" || echo "0")
PASS_COUNT=${PASS_COUNT//[^0-9]/}
PASS_COUNT=${PASS_COUNT:-0}
FAIL_COUNT=$(echo "$TEST_OUTPUT" | grep -c "^--- FAIL:" || echo "0")
FAIL_COUNT=${FAIL_COUNT//[^0-9]/}
FAIL_COUNT=${FAIL_COUNT:-0}
if echo "$TEST_OUTPUT" | grep -q "^ok" && [ "$FAIL_COUNT" -eq 0 ]; then
    pass "go test -short passes on groq package ($PASS_COUNT tests passed)"
else
    fail "go test -short FAILED on groq package ($PASS_COUNT passed, $FAIL_COUNT failed)"
fi

# Test 5.2: go vet clean
echo -e "  ${YELLOW}Running go vet verification...${NC}"
VET_OUTPUT=$(cd "$PROJECT_ROOT" && nice -n 19 ionice -c 3 go vet ./internal/llm/providers/groq/ 2>&1) || true
if [ -z "$VET_OUTPUT" ]; then
    pass "go vet produces clean output (no warnings)"
else
    fail "go vet produced warnings: $VET_OUTPUT"
fi

# Test 5.3: Groq package builds successfully
echo -e "  ${YELLOW}Building groq package...${NC}"
if (cd "$PROJECT_ROOT" && nice -n 19 ionice -c 3 go build ./internal/llm/providers/groq/ 2>&1); then
    pass "Groq package builds successfully"
else
    fail "Groq package build FAILED"
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

# Test 6.1: Groq in LLMsVerifier fallback models
VERIFIER_FALLBACK="$PROJECT_ROOT/LLMsVerifier/llm-verifier/providers/fallback_models.go"
if grep -q '"groq"' "$VERIFIER_FALLBACK" 2>/dev/null; then
    pass "Groq in LLMsVerifier fallback models"
else
    fail "Groq NOT found in LLMsVerifier fallback models"
fi

# Test 6.2: Groq provider config exists
VERIFIER_CONFIG="$PROJECT_ROOT/LLMsVerifier/llm-verifier/providers/config.go"
if grep -q '"groq"' "$VERIFIER_CONFIG" 2>/dev/null; then
    pass "Groq provider config exists in LLMsVerifier"
else
    fail "Groq provider config NOT found in LLMsVerifier"
fi

# Test 6.3: At least 5 Groq models in fallback
GROQ_MODEL_COUNT=$(grep -A 50 '"groq":' "$VERIFIER_FALLBACK" 2>/dev/null | grep -c 'ProviderID.*groq' || echo "0")
GROQ_MODEL_COUNT=${GROQ_MODEL_COUNT//[^0-9]/}
GROQ_MODEL_COUNT=${GROQ_MODEL_COUNT:-0}
if [ "$GROQ_MODEL_COUNT" -ge 5 ]; then
    pass "At least 5 Groq models in LLMsVerifier fallback (found $GROQ_MODEL_COUNT)"
else
    fail "Less than 5 Groq models in LLMsVerifier fallback (found $GROQ_MODEL_COUNT)"
fi

# Test 6.4: Groq scoring entries exist
if grep -q 'groq' "$VERIFIER_DIR/scoring.go" 2>/dev/null; then
    pass "Groq scoring entries exist in verifier scoring.go"
else
    fail "Groq scoring entries NOT found in verifier scoring.go"
fi

#===============================================================================
# Summary
#===============================================================================
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Groq Comprehensive Challenge${NC}"
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
