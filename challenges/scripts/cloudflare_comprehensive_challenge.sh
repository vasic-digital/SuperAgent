#!/bin/bash
# Cloudflare Workers AI Comprehensive Verification Challenge
# VALIDATES: Cloudflare provider structure, model coverage, features,
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

CLOUDFLARE_DIR="$PROJECT_ROOT/internal/llm/providers/cloudflare"
VERIFIER_DIR="$PROJECT_ROOT/internal/verifier"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Cloudflare Comprehensive Challenge${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "  Validates: Provider structure, models, features,"
echo -e "  integration, test execution, LLMsVerifier integration"

#===============================================================================
# Group 1: Provider Structure (8 tests)
#===============================================================================
section "Group 1: Provider Structure (8 tests)"

# Test 1.1: CloudflareProvider struct exists in cloudflare.go
if grep -q 'type CloudflareProvider struct' "$CLOUDFLARE_DIR/cloudflare.go" 2>/dev/null; then
    pass "CloudflareProvider struct exists in cloudflare.go"
else
    fail "CloudflareProvider struct NOT found in cloudflare.go"
fi

# Test 1.2: Complete method exists
if grep -q 'func (p \*CloudflareProvider) Complete' "$CLOUDFLARE_DIR/cloudflare.go" 2>/dev/null; then
    pass "Complete method exists on CloudflareProvider"
else
    fail "Complete method NOT found on CloudflareProvider"
fi

# Test 1.3: CompleteStream method exists
if grep -q 'func (p \*CloudflareProvider) CompleteStream' "$CLOUDFLARE_DIR/cloudflare.go" 2>/dev/null; then
    pass "CompleteStream method exists on CloudflareProvider"
else
    fail "CompleteStream method NOT found on CloudflareProvider"
fi

# Test 1.4: HealthCheck method exists
if grep -q 'func (p \*CloudflareProvider) HealthCheck' "$CLOUDFLARE_DIR/cloudflare.go" 2>/dev/null; then
    pass "HealthCheck method exists on CloudflareProvider"
else
    fail "HealthCheck method NOT found on CloudflareProvider"
fi

# Test 1.5: GetCapabilities method exists
if grep -q 'func (p \*CloudflareProvider) GetCapabilities' "$CLOUDFLARE_DIR/cloudflare.go" 2>/dev/null; then
    pass "GetCapabilities method exists on CloudflareProvider"
else
    fail "GetCapabilities method NOT found on CloudflareProvider"
fi

# Test 1.6: ValidateConfig method exists
if grep -q 'func (p \*CloudflareProvider) ValidateConfig' "$CLOUDFLARE_DIR/cloudflare.go" 2>/dev/null; then
    pass "ValidateConfig method exists on CloudflareProvider"
else
    fail "ValidateConfig method NOT found on CloudflareProvider"
fi

# Test 1.7: Registered in provider_registry.go
if grep -q 'case "cloudflare":' "$PROJECT_ROOT/internal/services/provider_registry.go" 2>/dev/null; then
    pass "Cloudflare registered in provider_registry.go"
else
    fail "Cloudflare NOT registered in provider_registry.go"
fi

# Test 1.8: AccountID support (CLOUDFLARE_ACCOUNT_ID in code)
if grep -q 'CLOUDFLARE_ACCOUNT_ID' "$CLOUDFLARE_DIR/cloudflare.go" 2>/dev/null; then
    pass "AccountID support (CLOUDFLARE_ACCOUNT_ID in code)"
else
    fail "AccountID support (CLOUDFLARE_ACCOUNT_ID) NOT found in cloudflare.go"
fi

#===============================================================================
# Group 2: Model Coverage (10 tests)
#===============================================================================
section "Group 2: Model Coverage (10 tests)"

# Models are checked across both provider code and LLMsVerifier fallback
FALLBACK="$PROJECT_ROOT/LLMsVerifier/llm-verifier/providers/fallback_models.go"

# Test 2.1: @cf/meta/llama-3.1-8b-instruct in models
if grep -rq '@cf/meta/llama-3\.1-8b-instruct' "$CLOUDFLARE_DIR/" 2>/dev/null; then
    pass "@cf/meta/llama-3.1-8b-instruct in provider models"
else
    fail "@cf/meta/llama-3.1-8b-instruct NOT found in provider models"
fi

# Test 2.2: @cf/meta/llama-3.1-70b-instruct in models
if grep -rq '@cf/meta/llama-3\.1-70b-instruct' "$CLOUDFLARE_DIR/" 2>/dev/null; then
    pass "@cf/meta/llama-3.1-70b-instruct in provider models"
else
    fail "@cf/meta/llama-3.1-70b-instruct NOT found in provider models"
fi

# Test 2.3: @cf/mistral/mistral-7b-instruct in models
if grep -rq '@cf/mistral/mistral-7b-instruct' "$CLOUDFLARE_DIR/" 2>/dev/null; then
    pass "@cf/mistral/mistral-7b-instruct in provider models"
else
    fail "@cf/mistral/mistral-7b-instruct NOT found in provider models"
fi

# Test 2.4: @cf/qwen/qwen3-30b-a3b-fp8 in models
if grep -rq '@cf/qwen/qwen3-30b-a3b-fp8' "$CLOUDFLARE_DIR/" 2>/dev/null; then
    pass "@cf/qwen/qwen3-30b-a3b-fp8 in provider models"
else
    fail "@cf/qwen/qwen3-30b-a3b-fp8 NOT found in provider models"
fi

# Test 2.5: @cf/meta/llama-3.1-8b-instruct in LLMsVerifier fallback
if grep -q '@cf/meta/llama-3\.1-8b-instruct' "$FALLBACK" 2>/dev/null; then
    pass "@cf/meta/llama-3.1-8b-instruct in LLMsVerifier fallback"
else
    fail "@cf/meta/llama-3.1-8b-instruct NOT found in LLMsVerifier fallback"
fi

# Test 2.6: @cf/meta/llama-3.1-70b-instruct in LLMsVerifier fallback
if grep -q '@cf/meta/llama-3\.1-70b-instruct' "$FALLBACK" 2>/dev/null; then
    pass "@cf/meta/llama-3.1-70b-instruct in LLMsVerifier fallback"
else
    fail "@cf/meta/llama-3.1-70b-instruct NOT found in LLMsVerifier fallback"
fi

# Test 2.7: @cf/mistral/mistral-small-3.1-24b-instruct in LLMsVerifier fallback
if grep -q '@cf/mistral/mistral-small-3\.1-24b-instruct' "$FALLBACK" 2>/dev/null; then
    pass "@cf/mistral/mistral-small-3.1-24b-instruct in LLMsVerifier fallback"
else
    fail "@cf/mistral/mistral-small-3.1-24b-instruct NOT found in LLMsVerifier fallback"
fi

# Test 2.8: @cf/qwen/qwq-32b in LLMsVerifier fallback
if grep -q '@cf/qwen/qwq-32b' "$FALLBACK" 2>/dev/null; then
    pass "@cf/qwen/qwq-32b in LLMsVerifier fallback"
else
    fail "@cf/qwen/qwq-32b NOT found in LLMsVerifier fallback"
fi

# Test 2.9: Default model constant defined
if grep -q 'CloudflareModel.*=.*"@cf/' "$CLOUDFLARE_DIR/cloudflare.go" 2>/dev/null; then
    pass "Default CloudflareModel constant defined"
else
    fail "Default CloudflareModel constant NOT defined"
fi

# Test 2.10: CloudflareModelsURL constant for discovery
if grep -q 'CloudflareModelsURL' "$CLOUDFLARE_DIR/cloudflare.go" 2>/dev/null; then
    pass "CloudflareModelsURL constant defined for model discovery"
else
    fail "CloudflareModelsURL constant NOT defined"
fi

#===============================================================================
# Group 3: Features (6 tests)
#===============================================================================
section "Group 3: Features (6 tests)"

# Test 3.1: OpenAI-compatible endpoint (ai/v1/chat/completions in code)
if grep -q 'ai/v1/chat/completions' "$CLOUDFLARE_DIR/cloudflare.go" 2>/dev/null; then
    pass "OpenAI-compatible endpoint (ai/v1/chat/completions in code)"
else
    fail "OpenAI-compatible endpoint (ai/v1/chat/completions) NOT found in cloudflare.go"
fi

# Test 3.2: Streaming support (SSE parsing with data: prefix)
if grep -q 'data: ' "$CLOUDFLARE_DIR/cloudflare.go" 2>/dev/null; then
    pass "Streaming support (SSE data: prefix parsing)"
else
    fail "Streaming support (SSE data: prefix) NOT found in cloudflare.go"
fi

# Test 3.3: Bearer token authentication
if grep -q 'Bearer' "$CLOUDFLARE_DIR/cloudflare.go" 2>/dev/null; then
    pass "Bearer token authentication"
else
    fail "Bearer token authentication NOT found in cloudflare.go"
fi

# Test 3.4: Retry with exponential backoff
if grep -q 'MaxRetries' "$CLOUDFLARE_DIR/cloudflare.go" 2>/dev/null && grep -q 'Multiplier' "$CLOUDFLARE_DIR/cloudflare.go" 2>/dev/null; then
    pass "Retry with exponential backoff (MaxRetries + Multiplier)"
else
    fail "Retry with exponential backoff NOT found in cloudflare.go"
fi

# Test 3.5: Model discovery integration
if grep -q 'discovery\.NewDiscoverer' "$CLOUDFLARE_DIR/cloudflare.go" 2>/dev/null; then
    pass "Model discovery integration (discovery.NewDiscoverer)"
else
    fail "Model discovery integration NOT found in cloudflare.go"
fi

# Test 3.6: Confidence calculation
if grep -q 'confidence' "$CLOUDFLARE_DIR/cloudflare.go" 2>/dev/null; then
    pass "Confidence calculation exists"
else
    fail "Confidence calculation NOT found in cloudflare.go"
fi

#===============================================================================
# Group 4: Integration (5 tests)
#===============================================================================
section "Group 4: Integration (5 tests)"

# Test 4.1: CLOUDFLARE_API_KEY in .env.example
if grep -q 'CLOUDFLARE_API_KEY' "$PROJECT_ROOT/.env.example" 2>/dev/null; then
    pass "CLOUDFLARE_API_KEY in .env.example"
else
    fail "CLOUDFLARE_API_KEY NOT found in .env.example"
fi

# Test 4.2: CLOUDFLARE_ACCOUNT_ID in .env.example
if grep -q 'CLOUDFLARE_ACCOUNT_ID' "$PROJECT_ROOT/.env.example" 2>/dev/null; then
    pass "CLOUDFLARE_ACCOUNT_ID in .env.example"
else
    fail "CLOUDFLARE_ACCOUNT_ID NOT found in .env.example"
fi

# Test 4.3: Unit tests exist (cloudflare_test.go)
if [ -f "$CLOUDFLARE_DIR/cloudflare_test.go" ]; then
    pass "Unit tests exist (cloudflare_test.go)"
else
    fail "Unit tests NOT found (cloudflare_test.go)"
fi

# Test 4.4: go vet passes on cloudflare package
echo -e "  ${YELLOW}Running go vet on cloudflare package...${NC}"
if (cd "$PROJECT_ROOT" && nice -n 19 ionice -c 3 go vet ./internal/llm/providers/cloudflare/ 2>&1); then
    pass "go vet passes on cloudflare package"
else
    fail "go vet FAILED on cloudflare package"
fi

# Test 4.5: Cloudflare in verifier provider_types.go
if grep -q '"cloudflare"' "$VERIFIER_DIR/provider_types.go" 2>/dev/null; then
    pass "Cloudflare in verifier provider_types.go"
else
    fail "Cloudflare NOT found in verifier provider_types.go"
fi

#===============================================================================
# Group 5: Test Execution (4 tests)
#===============================================================================
section "Group 5: Test Execution (4 tests)"

# Test 5.1: go test -short passes on cloudflare package
echo -e "  ${YELLOW}Running go test -short on cloudflare package...${NC}"
TEST_OUTPUT=$(cd "$PROJECT_ROOT" && nice -n 19 ionice -c 3 go test -v -short -count=1 -p 1 -timeout 120s ./internal/llm/providers/cloudflare/ 2>&1) || true
PASS_COUNT=$(echo "$TEST_OUTPUT" | grep -c "^--- PASS:" || echo "0")
PASS_COUNT=${PASS_COUNT//[^0-9]/}
PASS_COUNT=${PASS_COUNT:-0}
FAIL_COUNT=$(echo "$TEST_OUTPUT" | grep -c "^--- FAIL:" || echo "0")
FAIL_COUNT=${FAIL_COUNT//[^0-9]/}
FAIL_COUNT=${FAIL_COUNT:-0}
if echo "$TEST_OUTPUT" | grep -q "^ok" && [ "$FAIL_COUNT" -eq 0 ]; then
    pass "go test -short passes on cloudflare package ($PASS_COUNT tests passed)"
else
    fail "go test -short FAILED on cloudflare package ($PASS_COUNT passed, $FAIL_COUNT failed)"
fi

# Test 5.2: go vet clean
echo -e "  ${YELLOW}Running go vet verification...${NC}"
VET_OUTPUT=$(cd "$PROJECT_ROOT" && nice -n 19 ionice -c 3 go vet ./internal/llm/providers/cloudflare/ 2>&1) || true
if [ -z "$VET_OUTPUT" ]; then
    pass "go vet produces clean output (no warnings)"
else
    fail "go vet produced warnings: $VET_OUTPUT"
fi

# Test 5.3: cloudflare package builds successfully
echo -e "  ${YELLOW}Building cloudflare package...${NC}"
if (cd "$PROJECT_ROOT" && nice -n 19 ionice -c 3 go build ./internal/llm/providers/cloudflare/ 2>&1); then
    pass "cloudflare package builds successfully"
else
    fail "cloudflare package build FAILED"
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

# Test 6.1: Cloudflare in LLMsVerifier fallback models
VERIFIER_FALLBACK="$PROJECT_ROOT/LLMsVerifier/llm-verifier/providers/fallback_models.go"
if grep -q '"cloudflare"' "$VERIFIER_FALLBACK" 2>/dev/null; then
    pass "Cloudflare in LLMsVerifier fallback models"
else
    fail "Cloudflare NOT found in LLMsVerifier fallback models"
fi

# Test 6.2: At least 4 Cloudflare models in fallback
CF_MODEL_COUNT=$(grep -A 20 '"cloudflare":' "$VERIFIER_FALLBACK" 2>/dev/null | grep -c 'ProviderID.*cloudflare' || echo "0")
CF_MODEL_COUNT=${CF_MODEL_COUNT//[^0-9]/}
CF_MODEL_COUNT=${CF_MODEL_COUNT:-0}
if [ "$CF_MODEL_COUNT" -ge 4 ]; then
    pass "At least 4 Cloudflare models in LLMsVerifier fallback (found $CF_MODEL_COUNT)"
else
    fail "Less than 4 Cloudflare models in LLMsVerifier fallback (found $CF_MODEL_COUNT)"
fi

# Test 6.3: Cloudflare provider config exists in LLMsVerifier
VERIFIER_CONFIG="$PROJECT_ROOT/LLMsVerifier/llm-verifier/providers/config.go"
if grep -q '"cloudflare"' "$VERIFIER_CONFIG" 2>/dev/null; then
    pass "Cloudflare provider config exists in LLMsVerifier"
else
    fail "Cloudflare provider config NOT found in LLMsVerifier"
fi

#===============================================================================
# Summary
#===============================================================================
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Cloudflare Comprehensive Challenge${NC}"
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
