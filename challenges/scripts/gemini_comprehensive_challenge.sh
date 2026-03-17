#!/bin/bash
# Gemini Comprehensive Verification Challenge
# VALIDATES: Gemini provider structure, model coverage, power features,
#            integration, test execution, and LLMsVerifier integration
# Total: ~40 tests across 6 groups

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

GEMINI_DIR="$PROJECT_ROOT/internal/llm/providers/gemini"
VERIFIER_DIR="$PROJECT_ROOT/internal/verifier"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Gemini Comprehensive Challenge${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "  Validates: Provider structure, models, power features,"
echo -e "  integration, test execution, LLMsVerifier integration"

#===============================================================================
# Group 1: Provider Structure (10 tests)
#===============================================================================
section "Group 1: Provider Structure (10 tests)"

# Test 1.1: GeminiUnifiedProvider struct exists in gemini.go
if grep -q 'type GeminiUnifiedProvider struct' "$GEMINI_DIR/gemini.go" 2>/dev/null; then
    pass "GeminiUnifiedProvider struct exists in gemini.go"
else
    fail "GeminiUnifiedProvider struct NOT found in gemini.go"
fi

# Test 1.2: GeminiAPIProvider struct exists in gemini_api.go
if grep -q 'type GeminiAPIProvider struct' "$GEMINI_DIR/gemini_api.go" 2>/dev/null; then
    pass "GeminiAPIProvider struct exists in gemini_api.go"
else
    fail "GeminiAPIProvider struct NOT found in gemini_api.go"
fi

# Test 1.3: GeminiCLIProvider struct exists in gemini_cli.go
if grep -q 'type GeminiCLIProvider struct' "$GEMINI_DIR/gemini_cli.go" 2>/dev/null; then
    pass "GeminiCLIProvider struct exists in gemini_cli.go"
else
    fail "GeminiCLIProvider struct NOT found in gemini_cli.go"
fi

# Test 1.4: GeminiACPProvider struct exists in gemini_acp.go
if grep -q 'type GeminiACPProvider struct' "$GEMINI_DIR/gemini_acp.go" 2>/dev/null; then
    pass "GeminiACPProvider struct exists in gemini_acp.go"
else
    fail "GeminiACPProvider struct NOT found in gemini_acp.go"
fi

# Test 1.5: Complete method exists on unified provider
if grep -q 'func (p \*GeminiUnifiedProvider) Complete' "$GEMINI_DIR/gemini.go" 2>/dev/null; then
    pass "Complete method exists on GeminiUnifiedProvider"
else
    fail "Complete method NOT found on GeminiUnifiedProvider"
fi

# Test 1.6: CompleteStream method exists on unified provider
if grep -q 'func (p \*GeminiUnifiedProvider) CompleteStream' "$GEMINI_DIR/gemini.go" 2>/dev/null; then
    pass "CompleteStream method exists on GeminiUnifiedProvider"
else
    fail "CompleteStream method NOT found on GeminiUnifiedProvider"
fi

# Test 1.7: HealthCheck method exists on unified provider
if grep -q 'func (p \*GeminiUnifiedProvider) HealthCheck' "$GEMINI_DIR/gemini.go" 2>/dev/null; then
    pass "HealthCheck method exists on GeminiUnifiedProvider"
else
    fail "HealthCheck method NOT found on GeminiUnifiedProvider"
fi

# Test 1.8: GetCapabilities method exists on unified provider
if grep -q 'func (p \*GeminiUnifiedProvider) GetCapabilities' "$GEMINI_DIR/gemini.go" 2>/dev/null; then
    pass "GetCapabilities method exists on GeminiUnifiedProvider"
else
    fail "GetCapabilities method NOT found on GeminiUnifiedProvider"
fi

# Test 1.9: ValidateConfig method exists on unified provider
if grep -q 'func (p \*GeminiUnifiedProvider) ValidateConfig' "$GEMINI_DIR/gemini.go" 2>/dev/null; then
    pass "ValidateConfig method exists on GeminiUnifiedProvider"
else
    fail "ValidateConfig method NOT found on GeminiUnifiedProvider"
fi

# Test 1.10: Registered in provider_registry.go
if grep -q 'case "gemini":' "$PROJECT_ROOT/internal/services/provider_registry.go" 2>/dev/null; then
    pass "Gemini registered in provider_registry.go"
else
    fail "Gemini NOT registered in provider_registry.go"
fi

#===============================================================================
# Group 2: Model Coverage (8 tests)
#===============================================================================
section "Group 2: Model Coverage (8 tests)"

# Test 2.1: gemini-3.1-pro-preview in known models
if grep -rq 'gemini-3\.1-pro-preview' "$GEMINI_DIR/" 2>/dev/null; then
    pass "gemini-3.1-pro-preview in known models"
else
    fail "gemini-3.1-pro-preview NOT found in known models"
fi

# Test 2.2: gemini-3-pro-preview in known models
if grep -rq 'gemini-3-pro-preview' "$GEMINI_DIR/" 2>/dev/null; then
    pass "gemini-3-pro-preview in known models"
else
    fail "gemini-3-pro-preview NOT found in known models"
fi

# Test 2.3: gemini-3-flash-preview in known models
if grep -rq 'gemini-3-flash-preview' "$GEMINI_DIR/" 2>/dev/null; then
    pass "gemini-3-flash-preview in known models"
else
    fail "gemini-3-flash-preview NOT found in known models"
fi

# Test 2.4: gemini-2.5-pro in known models
if grep -rq 'gemini-2\.5-pro' "$GEMINI_DIR/" 2>/dev/null; then
    pass "gemini-2.5-pro in known models"
else
    fail "gemini-2.5-pro NOT found in known models"
fi

# Test 2.5: gemini-2.5-flash in known models
if grep -rq 'gemini-2\.5-flash' "$GEMINI_DIR/" 2>/dev/null; then
    pass "gemini-2.5-flash in known models"
else
    fail "gemini-2.5-flash NOT found in known models"
fi

# Test 2.6: gemini-2.5-flash-lite in known models
if grep -rq 'gemini-2\.5-flash-lite' "$GEMINI_DIR/" 2>/dev/null; then
    pass "gemini-2.5-flash-lite in known models"
else
    fail "gemini-2.5-flash-lite NOT found in known models"
fi

# Test 2.7: gemini-2.0-flash in known models
if grep -rq 'gemini-2\.0-flash' "$GEMINI_DIR/" 2>/dev/null; then
    pass "gemini-2.0-flash in known models"
else
    fail "gemini-2.0-flash NOT found in known models"
fi

# Test 2.8: gemini-embedding-001 in known models
if grep -rq 'gemini-embedding-001' "$GEMINI_DIR/" 2>/dev/null; then
    pass "gemini-embedding-001 in known models"
else
    fail "gemini-embedding-001 NOT found in known models"
fi

#===============================================================================
# Group 3: Power Features (6 tests)
#===============================================================================
section "Group 3: Power Features (6 tests)"

# Test 3.1: Extended thinking support (ThinkingConfig/thinkingBudget)
if grep -q 'ThinkingConfig\|thinkingBudget' "$GEMINI_DIR/gemini_api.go" 2>/dev/null; then
    pass "Extended thinking support (ThinkingConfig/thinkingBudget)"
else
    fail "Extended thinking support NOT found in gemini_api.go"
fi

# Test 3.2: Google Search grounding (GoogleSearch/googleSearch)
if grep -q 'GoogleSearch\|googleSearch' "$GEMINI_DIR/gemini_api.go" 2>/dev/null; then
    pass "Google Search grounding support"
else
    fail "Google Search grounding NOT found in gemini_api.go"
fi

# Test 3.3: Session management (sessionID in CLI provider)
if grep -q 'sessionID' "$GEMINI_DIR/gemini_cli.go" 2>/dev/null; then
    pass "Session management (sessionID in CLI provider)"
else
    fail "Session management (sessionID) NOT found in gemini_cli.go"
fi

# Test 3.4: ACP protocol support (experimental-acp)
if grep -q 'experimental-acp' "$GEMINI_DIR/gemini_acp.go" 2>/dev/null; then
    pass "ACP protocol support (experimental-acp)"
else
    fail "ACP protocol support (experimental-acp) NOT found in gemini_acp.go"
fi

# Test 3.5: Model discovery (DiscoverModels function exists)
if grep -q 'func.*DiscoverModels' "$GEMINI_DIR/gemini_cli.go" 2>/dev/null; then
    pass "Model discovery (DiscoverModels function exists)"
else
    fail "Model discovery (DiscoverModels) NOT found in gemini_cli.go"
fi

# Test 3.6: Streaming support (stream-json in CLI code)
if grep -q 'stream-json' "$GEMINI_DIR/gemini_cli.go" 2>/dev/null; then
    pass "Streaming support (stream-json in CLI code)"
else
    fail "Streaming support (stream-json) NOT found in gemini_cli.go"
fi

#===============================================================================
# Group 4: Integration (6 tests)
#===============================================================================
section "Group 4: Integration (6 tests)"

# Test 4.1: GEMINI_API_KEY in .env.example
if grep -q 'GEMINI_API_KEY' "$PROJECT_ROOT/.env.example" 2>/dev/null; then
    pass "GEMINI_API_KEY in .env.example"
else
    fail "GEMINI_API_KEY NOT found in .env.example"
fi

# Test 4.2: Unit tests exist (gemini_test.go)
if [ -f "$GEMINI_DIR/gemini_test.go" ]; then
    pass "Unit tests exist (gemini_test.go)"
else
    fail "Unit tests NOT found (gemini_test.go)"
fi

# Test 4.3: CLI provider tests exist
if grep -q 'TestGeminiCLIProvider' "$GEMINI_DIR/gemini_test.go" 2>/dev/null; then
    pass "CLI provider tests exist (TestGeminiCLIProvider)"
else
    fail "CLI provider tests NOT found (TestGeminiCLIProvider)"
fi

# Test 4.4: ACP provider tests exist
if grep -q 'TestGeminiACPProvider' "$GEMINI_DIR/gemini_test.go" 2>/dev/null; then
    pass "ACP provider tests exist (TestGeminiACPProvider)"
else
    fail "ACP provider tests NOT found (TestGeminiACPProvider)"
fi

# Test 4.5: Benchmark tests exist
if grep -q 'func Benchmark' "$GEMINI_DIR/gemini_test.go" 2>/dev/null; then
    pass "Benchmark tests exist in gemini_test.go"
else
    fail "Benchmark tests NOT found in gemini_test.go"
fi

# Test 4.6: go vet passes on gemini package
echo -e "  ${YELLOW}Running go vet on gemini package...${NC}"
if (cd "$PROJECT_ROOT" && nice -n 19 ionice -c 3 go vet ./internal/llm/providers/gemini/ 2>&1); then
    pass "go vet passes on gemini package"
else
    fail "go vet FAILED on gemini package"
fi

#===============================================================================
# Group 5: Test Execution (5 tests)
#===============================================================================
section "Group 5: Test Execution (5 tests)"

# Test 5.1: go test -short passes on gemini package
echo -e "  ${YELLOW}Running go test -short on gemini package...${NC}"
TEST_OUTPUT=$(cd "$PROJECT_ROOT" && nice -n 19 ionice -c 3 go test -v -short -count=1 -p 1 -timeout 120s ./internal/llm/providers/gemini/ 2>&1) || true
PASS_COUNT=$(echo "$TEST_OUTPUT" | grep -c "^--- PASS:" || echo "0")
PASS_COUNT=${PASS_COUNT//[^0-9]/}
PASS_COUNT=${PASS_COUNT:-0}
FAIL_COUNT=$(echo "$TEST_OUTPUT" | grep -c "^--- FAIL:" || echo "0")
FAIL_COUNT=${FAIL_COUNT//[^0-9]/}
FAIL_COUNT=${FAIL_COUNT:-0}
if echo "$TEST_OUTPUT" | grep -q "^ok" && [ "$FAIL_COUNT" -eq 0 ]; then
    pass "go test -short passes on gemini package ($PASS_COUNT tests passed)"
else
    fail "go test -short FAILED on gemini package ($PASS_COUNT passed, $FAIL_COUNT failed)"
fi

# Test 5.2: go vet clean on gemini package (separate from 4.6 for Group 5 completeness)
echo -e "  ${YELLOW}Running go vet verification...${NC}"
VET_OUTPUT=$(cd "$PROJECT_ROOT" && nice -n 19 ionice -c 3 go vet ./internal/llm/providers/gemini/ 2>&1) || true
if [ -z "$VET_OUTPUT" ]; then
    pass "go vet produces clean output (no warnings)"
else
    fail "go vet produced warnings: $VET_OUTPUT"
fi

# Test 5.3: Gemini package builds successfully
echo -e "  ${YELLOW}Building gemini package...${NC}"
if (cd "$PROJECT_ROOT" && nice -n 19 ionice -c 3 go build ./internal/llm/providers/gemini/ 2>&1); then
    pass "Gemini package builds successfully"
else
    fail "Gemini package build FAILED"
fi

# Test 5.4: helixagent binary builds successfully
echo -e "  ${YELLOW}Building helixagent binary...${NC}"
if (cd "$PROJECT_ROOT" && nice -n 19 ionice -c 3 go build ./cmd/helixagent/ 2>&1); then
    pass "helixagent binary builds successfully"
else
    fail "helixagent binary build FAILED"
fi

# Test 5.5: Test function count >= 40
TEST_FUNC_COUNT=$(grep -c '^func Test' "$GEMINI_DIR/gemini_test.go" 2>/dev/null || echo "0")
TEST_FUNC_COUNT=${TEST_FUNC_COUNT//[^0-9]/}
TEST_FUNC_COUNT=${TEST_FUNC_COUNT:-0}
BENCH_FUNC_COUNT=$(grep -c '^func Benchmark' "$GEMINI_DIR/gemini_test.go" 2>/dev/null || echo "0")
BENCH_FUNC_COUNT=${BENCH_FUNC_COUNT//[^0-9]/}
BENCH_FUNC_COUNT=${BENCH_FUNC_COUNT:-0}
TOTAL_FUNCS=$((TEST_FUNC_COUNT + BENCH_FUNC_COUNT))
if [ "$TOTAL_FUNCS" -ge 40 ]; then
    pass "Test function count >= 40 (found $TOTAL_FUNCS: $TEST_FUNC_COUNT tests + $BENCH_FUNC_COUNT benchmarks)"
else
    fail "Test function count < 40 (found $TOTAL_FUNCS: $TEST_FUNC_COUNT tests + $BENCH_FUNC_COUNT benchmarks)"
fi

#===============================================================================
# Group 6: LLMsVerifier Integration (5 tests)
#===============================================================================
section "Group 6: LLMsVerifier Integration (5 tests)"

# Test 6.1: Gemini in LLMsVerifier provider_types fallback models
if grep -q 'gemini' "$VERIFIER_DIR/provider_types.go" 2>/dev/null; then
    pass "Gemini in LLMsVerifier provider_types.go"
else
    fail "Gemini NOT found in LLMsVerifier provider_types.go"
fi

# Test 6.2: Gemini provider in verifier provider_access config
if grep -q '"gemini"' "$VERIFIER_DIR/provider_access.go" 2>/dev/null; then
    pass "Gemini provider in verifier provider_access.go"
else
    fail "Gemini NOT found in verifier provider_access.go"
fi

# Test 6.3: At least 5 Gemini models across provider codebase
GEMINI_MODEL_COUNT=$(grep -ro 'gemini-[0-9][a-z0-9.\-]*' "$GEMINI_DIR/gemini.go" "$GEMINI_DIR/gemini_api.go" "$GEMINI_DIR/gemini_cli.go" "$GEMINI_DIR/gemini_acp.go" 2>/dev/null | sort -u | wc -l)
GEMINI_MODEL_COUNT=${GEMINI_MODEL_COUNT//[^0-9]/}
GEMINI_MODEL_COUNT=${GEMINI_MODEL_COUNT:-0}
if [ "$GEMINI_MODEL_COUNT" -ge 5 ]; then
    pass "At least 5 unique Gemini models in provider code (found $GEMINI_MODEL_COUNT)"
else
    fail "Less than 5 unique Gemini models in provider code (found $GEMINI_MODEL_COUNT)"
fi

# Test 6.4: Provider ID consistency ("gemini" not "google" as primary)
if grep -q '"gemini"' "$PROJECT_ROOT/internal/services/provider_registry.go" 2>/dev/null; then
    # Also verify the provider_types uses "gemini" as key
    if grep -q '"gemini":' "$VERIFIER_DIR/provider_types.go" 2>/dev/null; then
        pass "Provider ID consistency: 'gemini' used as primary ID (not 'google')"
    else
        fail "Provider ID inconsistency: 'gemini' key not in provider_types.go"
    fi
else
    fail "Provider ID 'gemini' NOT found in provider_registry.go"
fi

# Test 6.5: Gemini API endpoint correct (generativelanguage.googleapis.com)
if grep -q 'generativelanguage\.googleapis\.com' "$GEMINI_DIR/gemini.go" 2>/dev/null; then
    pass "Gemini API endpoint correct (generativelanguage.googleapis.com)"
else
    fail "Gemini API endpoint NOT using generativelanguage.googleapis.com"
fi

#===============================================================================
# Summary
#===============================================================================
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Gemini Comprehensive Challenge${NC}"
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
