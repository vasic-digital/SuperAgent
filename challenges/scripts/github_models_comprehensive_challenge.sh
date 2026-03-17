#!/bin/bash
# GitHub Models Comprehensive Verification Challenge
# VALIDATES: GitHub Models provider structure, model coverage, features,
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

GITHUB_MODELS_DIR="$PROJECT_ROOT/internal/llm/providers/githubmodels"
VERIFIER_DIR="$PROJECT_ROOT/internal/verifier"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  GitHub Models Comprehensive Challenge${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "  Validates: Provider structure, models, features,"
echo -e "  integration, test execution, LLMsVerifier integration"

#===============================================================================
# Group 1: Provider Structure (8 tests)
#===============================================================================
section "Group 1: Provider Structure (8 tests)"

# Test 1.1: GitHubModelsProvider struct exists in githubmodels.go
if grep -q 'type GitHubModelsProvider struct' "$GITHUB_MODELS_DIR/githubmodels.go" 2>/dev/null; then
    pass "GitHubModelsProvider struct exists in githubmodels.go"
else
    fail "GitHubModelsProvider struct NOT found in githubmodels.go"
fi

# Test 1.2: Complete method exists
if grep -q 'func (p \*GitHubModelsProvider) Complete' "$GITHUB_MODELS_DIR/githubmodels.go" 2>/dev/null; then
    pass "Complete method exists on GitHubModelsProvider"
else
    fail "Complete method NOT found on GitHubModelsProvider"
fi

# Test 1.3: CompleteStream method exists
if grep -q 'func (p \*GitHubModelsProvider) CompleteStream' "$GITHUB_MODELS_DIR/githubmodels.go" 2>/dev/null; then
    pass "CompleteStream method exists on GitHubModelsProvider"
else
    fail "CompleteStream method NOT found on GitHubModelsProvider"
fi

# Test 1.4: HealthCheck method exists
if grep -q 'func (p \*GitHubModelsProvider) HealthCheck' "$GITHUB_MODELS_DIR/githubmodels.go" 2>/dev/null; then
    pass "HealthCheck method exists on GitHubModelsProvider"
else
    fail "HealthCheck method NOT found on GitHubModelsProvider"
fi

# Test 1.5: GetCapabilities method exists
if grep -q 'func (p \*GitHubModelsProvider) GetCapabilities' "$GITHUB_MODELS_DIR/githubmodels.go" 2>/dev/null; then
    pass "GetCapabilities method exists on GitHubModelsProvider"
else
    fail "GetCapabilities method NOT found on GitHubModelsProvider"
fi

# Test 1.6: ValidateConfig method exists
if grep -q 'func (p \*GitHubModelsProvider) ValidateConfig' "$GITHUB_MODELS_DIR/githubmodels.go" 2>/dev/null; then
    pass "ValidateConfig method exists on GitHubModelsProvider"
else
    fail "ValidateConfig method NOT found on GitHubModelsProvider"
fi

# Test 1.7: Registered in provider_registry.go
if grep -q 'case "github-models":' "$PROJECT_ROOT/internal/services/provider_registry.go" 2>/dev/null; then
    pass "github-models registered in provider_registry.go"
else
    fail "github-models NOT registered in provider_registry.go"
fi

# Test 1.8: Uses models.github.ai endpoint
if grep -q 'models.github.ai' "$GITHUB_MODELS_DIR/githubmodels.go" 2>/dev/null; then
    pass "Uses models.github.ai endpoint"
else
    fail "models.github.ai endpoint NOT found in githubmodels.go"
fi

#===============================================================================
# Group 2: Model Coverage (10 tests)
#===============================================================================
section "Group 2: Model Coverage (10 tests)"

# Test 2.1: openai/gpt-5 in models
if grep -rq 'openai/gpt-5' "$GITHUB_MODELS_DIR/" 2>/dev/null; then
    pass "openai/gpt-5 in known models"
else
    fail "openai/gpt-5 NOT found in known models"
fi

# Test 2.2: openai/gpt-4.1 in models
if grep -rq 'openai/gpt-4\.1' "$GITHUB_MODELS_DIR/" 2>/dev/null; then
    pass "openai/gpt-4.1 in known models"
else
    fail "openai/gpt-4.1 NOT found in known models"
fi

# Test 2.3: openai/gpt-4o in models
if grep -rq 'openai/gpt-4o' "$GITHUB_MODELS_DIR/" 2>/dev/null; then
    pass "openai/gpt-4o in known models"
else
    fail "openai/gpt-4o NOT found in known models"
fi

# Test 2.4: DeepSeek/DeepSeek-V3 in models
if grep -rq 'DeepSeek/DeepSeek-V3' "$GITHUB_MODELS_DIR/" 2>/dev/null; then
    pass "DeepSeek/DeepSeek-V3 in known models"
else
    fail "DeepSeek/DeepSeek-V3 NOT found in known models"
fi

# Test 2.5: DeepSeek/DeepSeek-R1 in models
if grep -rq 'DeepSeek/DeepSeek-R1' "$GITHUB_MODELS_DIR/" 2>/dev/null; then
    pass "DeepSeek/DeepSeek-R1 in known models"
else
    fail "DeepSeek/DeepSeek-R1 NOT found in known models"
fi

# Test 2.6: Meta/Llama-4-Scout in models
if grep -rq 'Meta/Llama-4-Scout' "$GITHUB_MODELS_DIR/" 2>/dev/null; then
    pass "Meta/Llama-4-Scout in known models"
else
    fail "Meta/Llama-4-Scout NOT found in known models"
fi

# Test 2.7: Meta/Llama-3.3-70B in models
if grep -rq 'Meta/Llama-3\.3-70B' "$GITHUB_MODELS_DIR/" 2>/dev/null; then
    pass "Meta/Llama-3.3-70B in known models"
else
    fail "Meta/Llama-3.3-70B NOT found in known models"
fi

# Test 2.8: Microsoft/Phi-4 in models
if grep -rq 'Microsoft/Phi-4' "$GITHUB_MODELS_DIR/" 2>/dev/null; then
    pass "Microsoft/Phi-4 in known models"
else
    fail "Microsoft/Phi-4 NOT found in known models"
fi

# Test 2.9: Mistral/Mistral-Large in models
if grep -rq 'Mistral/Mistral-Large' "$GITHUB_MODELS_DIR/" 2>/dev/null; then
    pass "Mistral/Mistral-Large in known models"
else
    fail "Mistral/Mistral-Large NOT found in known models"
fi

# Test 2.10: Cohere/Command-A in models
if grep -rq 'Cohere/Command-A' "$GITHUB_MODELS_DIR/" 2>/dev/null; then
    pass "Cohere/Command-A in known models"
else
    fail "Cohere/Command-A NOT found in known models"
fi

#===============================================================================
# Group 3: Features (6 tests)
#===============================================================================
section "Group 3: Features (6 tests)"

# Test 3.1: OpenAI-compatible API (chat/completions in code)
if grep -q 'chat/completions' "$GITHUB_MODELS_DIR/githubmodels.go" 2>/dev/null; then
    pass "OpenAI-compatible API (chat/completions endpoint)"
else
    fail "OpenAI-compatible API (chat/completions) NOT found in githubmodels.go"
fi

# Test 3.2: Streaming SSE support (text/event-stream or data: prefix parsing)
if grep -rq 'text/event-stream' "$GITHUB_MODELS_DIR/" 2>/dev/null; then
    pass "Streaming SSE support (text/event-stream in code)"
else
    # Also check for SSE data: prefix parsing which indicates SSE support
    if grep -q 'data: ' "$GITHUB_MODELS_DIR/githubmodels.go" 2>/dev/null; then
        pass "Streaming SSE support (data: prefix parsing in code)"
    else
        fail "Streaming SSE support NOT found in githubmodels code"
    fi
fi

# Test 3.3: Bearer token authentication
if grep -q 'Bearer' "$GITHUB_MODELS_DIR/githubmodels.go" 2>/dev/null; then
    pass "Bearer token authentication"
else
    fail "Bearer token authentication NOT found in githubmodels.go"
fi

# Test 3.4: X-GitHub-Api-Version header
if grep -q 'X-GitHub-Api-Version' "$GITHUB_MODELS_DIR/githubmodels.go" 2>/dev/null; then
    pass "X-GitHub-Api-Version header present"
else
    fail "X-GitHub-Api-Version header NOT found in githubmodels.go"
fi

# Test 3.5: Tool/function calling support (Tools field in request struct)
if grep -q 'Tools.*\[\]Tool' "$GITHUB_MODELS_DIR/githubmodels.go" 2>/dev/null; then
    pass "Tool/function calling support (Tools field in request struct)"
else
    fail "Tool/function calling support NOT found in githubmodels.go"
fi

# Test 3.6: Retry with exponential backoff
if grep -q 'calculateBackoff' "$GITHUB_MODELS_DIR/githubmodels.go" 2>/dev/null; then
    pass "Retry with exponential backoff (calculateBackoff)"
else
    fail "Retry with exponential backoff NOT found in githubmodels.go"
fi

#===============================================================================
# Group 4: Integration (5 tests)
#===============================================================================
section "Group 4: Integration (5 tests)"

# Test 4.1: GITHUB_MODELS_API_KEY in .env.example
if grep -q 'GITHUB_MODELS_API_KEY' "$PROJECT_ROOT/.env.example" 2>/dev/null; then
    pass "GITHUB_MODELS_API_KEY in .env.example"
else
    fail "GITHUB_MODELS_API_KEY NOT found in .env.example"
fi

# Test 4.2: Unit tests exist (githubmodels_test.go)
if [ -f "$GITHUB_MODELS_DIR/githubmodels_test.go" ]; then
    pass "Unit tests exist (githubmodels_test.go)"
else
    fail "Unit tests NOT found (githubmodels_test.go)"
fi

# Test 4.3: go vet passes on githubmodels package
echo -e "  ${YELLOW}Running go vet on githubmodels package...${NC}"
if (cd "$PROJECT_ROOT" && nice -n 19 ionice -c 3 go vet ./internal/llm/providers/githubmodels/ 2>&1); then
    pass "go vet passes on githubmodels package"
else
    fail "go vet FAILED on githubmodels package"
fi

# Test 4.4: github-models in verifier provider_types.go
if grep -q '"github-models"' "$VERIFIER_DIR/provider_types.go" 2>/dev/null; then
    pass "github-models in verifier provider_types.go"
else
    fail "github-models NOT found in verifier provider_types.go"
fi

# Test 4.5: github-models in provider discovery.go
if grep -q 'github-models' "$VERIFIER_DIR/discovery.go" 2>/dev/null; then
    pass "github-models in verifier discovery.go"
else
    fail "github-models NOT found in verifier discovery.go"
fi

#===============================================================================
# Group 5: Test Execution (4 tests)
#===============================================================================
section "Group 5: Test Execution (4 tests)"

# Test 5.1: go test -short passes on githubmodels package
echo -e "  ${YELLOW}Running go test -short on githubmodels package...${NC}"
TEST_OUTPUT=$(cd "$PROJECT_ROOT" && nice -n 19 ionice -c 3 go test -v -short -count=1 -p 1 -timeout 120s ./internal/llm/providers/githubmodels/ 2>&1) || true
PASS_COUNT=$(echo "$TEST_OUTPUT" | grep -c "^--- PASS:" || echo "0")
PASS_COUNT=${PASS_COUNT//[^0-9]/}
PASS_COUNT=${PASS_COUNT:-0}
FAIL_COUNT=$(echo "$TEST_OUTPUT" | grep -c "^--- FAIL:" || echo "0")
FAIL_COUNT=${FAIL_COUNT//[^0-9]/}
FAIL_COUNT=${FAIL_COUNT:-0}
if echo "$TEST_OUTPUT" | grep -q "^ok" && [ "$FAIL_COUNT" -eq 0 ]; then
    pass "go test -short passes on githubmodels package ($PASS_COUNT tests passed)"
else
    fail "go test -short FAILED on githubmodels package ($PASS_COUNT passed, $FAIL_COUNT failed)"
fi

# Test 5.2: go vet clean
echo -e "  ${YELLOW}Running go vet verification...${NC}"
VET_OUTPUT=$(cd "$PROJECT_ROOT" && nice -n 19 ionice -c 3 go vet ./internal/llm/providers/githubmodels/ 2>&1) || true
if [ -z "$VET_OUTPUT" ]; then
    pass "go vet produces clean output (no warnings)"
else
    fail "go vet produced warnings: $VET_OUTPUT"
fi

# Test 5.3: githubmodels package builds successfully
echo -e "  ${YELLOW}Building githubmodels package...${NC}"
if (cd "$PROJECT_ROOT" && nice -n 19 ionice -c 3 go build ./internal/llm/providers/githubmodels/ 2>&1); then
    pass "githubmodels package builds successfully"
else
    fail "githubmodels package build FAILED"
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

# Test 6.1: github-models in LLMsVerifier fallback models
VERIFIER_FALLBACK="$PROJECT_ROOT/LLMsVerifier/llm-verifier/providers/fallback_models.go"
if grep -q '"github-models"' "$VERIFIER_FALLBACK" 2>/dev/null; then
    pass "github-models in LLMsVerifier fallback models"
else
    fail "github-models NOT found in LLMsVerifier fallback models"
fi

# Test 6.2: At least 8 github-models entries in LLMsVerifier fallback
GH_MODEL_COUNT=$(grep -A 50 '"github-models":' "$VERIFIER_FALLBACK" 2>/dev/null | grep -c 'ProviderID.*github-models' || echo "0")
GH_MODEL_COUNT=${GH_MODEL_COUNT//[^0-9]/}
GH_MODEL_COUNT=${GH_MODEL_COUNT:-0}
if [ "$GH_MODEL_COUNT" -ge 8 ]; then
    pass "At least 8 github-models entries in LLMsVerifier fallback (found $GH_MODEL_COUNT)"
else
    fail "Less than 8 github-models entries in LLMsVerifier fallback (found $GH_MODEL_COUNT)"
fi

# Test 6.3: github-models scoring entries exist
if grep -q 'github-models' "$VERIFIER_DIR/scoring.go" 2>/dev/null; then
    pass "github-models scoring entries exist in verifier scoring.go"
else
    fail "github-models scoring entries NOT found in verifier scoring.go"
fi

#===============================================================================
# Summary
#===============================================================================
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  GitHub Models Comprehensive Challenge${NC}"
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
