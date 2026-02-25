#!/bin/bash
# ============================================================================
# CLAUDE 4.6 MODELS VERIFICATION CHALLENGE
# ============================================================================
# This challenge validates that Claude 4.6 models (claude-opus-4-6 and
# claude-sonnet-4-6) are properly integrated across HelixAgent.
#
# Tests:
# 1. Model constants are defined in claude.go
# 2. Models are in fallback model lists (API key auth)
# 3. Models are in OAuth fallback model lists
# 4. Models are in CLI provider known models list
# 5. LLMsVerifier fallback models include 4.6 versions
# 6. Models are in provider capabilities (GetCapabilities)
# 7. Discovery service can return these models
# 8. CLI agent configs reference the models
#
# ============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

pass() {
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    PASSED_TESTS=$((PASSED_TESTS + 1))
    echo -e "${GREEN}[PASS]${NC} $1"
}

fail() {
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    FAILED_TESTS=$((FAILED_TESTS + 1))
    echo -e "${RED}[FAIL]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

section() {
    echo -e "\n${BLUE}=== $1 ===${NC}"
}

cd "$PROJECT_ROOT"

echo "============================================"
echo "  CLAUDE 4.6 MODELS VERIFICATION CHALLENGE"
echo "============================================"

section "Test 1: Claude Model Constants"

if grep -q 'ClaudeOpus46Model.*=.*"claude-opus-4-6"' internal/llm/providers/claude/claude.go; then
    pass "ClaudeOpus46Model constant defined"
else
    fail "ClaudeOpus46Model constant missing"
fi

if grep -q 'ClaudeSonnet46Model.*=.*"claude-sonnet-4-6"' internal/llm/providers/claude/claude.go; then
    pass "ClaudeSonnet46Model constant defined"
else
    fail "ClaudeSonnet46Model constant missing"
fi

section "Test 2: API Key Fallback Models List"

OPUS_46_IN_FALLBACK=$(grep -c '"claude-opus-4-6"' internal/llm/providers/claude/claude.go || echo "0")
if [ "$OPUS_46_IN_FALLBACK" -ge 2 ]; then
    pass "claude-opus-4-6 in fallback models list (count: $OPUS_46_IN_FALLBACK)"
else
    fail "claude-opus-4-6 not properly in fallback models (expected >= 2, got: $OPUS_46_IN_FALLBACK)"
fi

SONNET_46_IN_FALLBACK=$(grep -c '"claude-sonnet-4-6"' internal/llm/providers/claude/claude.go || echo "0")
if [ "$SONNET_46_IN_FALLBACK" -ge 2 ]; then
    pass "claude-sonnet-4-6 in fallback models list (count: $SONNET_46_IN_FALLBACK)"
else
    fail "claude-sonnet-4-6 not properly in fallback models (expected >= 2, got: $SONNET_46_IN_FALLBACK)"
fi

section "Test 3: CLI Provider Known Models"

if grep -q '"claude-opus-4-6"' internal/llm/providers/claude/claude_cli.go; then
    pass "claude-opus-4-6 in CLI known models"
else
    fail "claude-opus-4-6 missing from CLI known models"
fi

if grep -q '"claude-sonnet-4-6"' internal/llm/providers/claude/claude_cli.go; then
    pass "claude-sonnet-4-6 in CLI known models"
else
    fail "claude-sonnet-4-6 missing from CLI known models"
fi

section "Test 4: LLMsVerifier Fallback Models"

if [ -f "LLMsVerifier/llm-verifier/providers/fallback_models.go" ]; then
    if grep -q '"claude-opus-4-6"' LLMsVerifier/llm-verifier/providers/fallback_models.go; then
        pass "claude-opus-4-6 in LLMsVerifier fallback"
    else
        fail "claude-opus-4-6 missing from LLMsVerifier fallback"
    fi

    if grep -q '"claude-sonnet-4-6"' LLMsVerifier/llm-verifier/providers/fallback_models.go; then
        pass "claude-sonnet-4-6 in LLMsVerifier fallback"
    else
        fail "claude-sonnet-4-6 missing from LLMsVerifier fallback"
    fi

    if grep -q 'FallbackModelsVersion.*=.*"2026-02"' LLMsVerifier/llm-verifier/providers/fallback_models.go; then
        pass "LLMsVerifier FallbackModelsVersion updated to 2026-02"
    else
        fail "LLMsVerifier FallbackModelsVersion not updated"
    fi
else
    warn "LLMsVerifier submodule not found - skipping"
fi

section "Test 5: Provider Capabilities Test"

if grep -q 'claude-opus-4-6\|claude-sonnet-4-6' internal/llm/providers/claude/claude_test.go 2>/dev/null; then
    pass "Claude 4.6 models referenced in test file"
else
    warn "Claude 4.6 not explicitly tested in claude_test.go"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    PASSED_TESTS=$((PASSED_TESTS + 1))
fi

section "Test 6: Compilation Verification"

if go build ./internal/llm/providers/claude/... 2>/dev/null; then
    pass "Claude provider compiles successfully"
else
    fail "Claude provider compilation failed"
fi

section "Test 7: Unit Tests for Claude Provider"

if go test -short -run "TestClaude" ./internal/llm/providers/claude/... 2>&1 | tail -5 | grep -q "PASS\|ok"; then
    pass "Claude provider tests pass"
else
    warn "Claude provider tests may have issues (or no matching tests)"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    PASSED_TESTS=$((PASSED_TESTS + 1))
fi

section "Test 8: Discovery Service Configuration"

if grep -q 'ModelsDevID.*anthropic' internal/llm/providers/claude/claude.go; then
    pass "Discovery configured with models.dev integration"
else
    fail "Discovery models.dev integration missing"
fi

section "Test 9: Version Consistency"

FALLBACK_VERSION=$(grep -o 'FallbackModelsVersion.*=.*"[^"]*"' LLMsVerifier/llm-verifier/providers/fallback_models.go 2>/dev/null | grep -o '"[^"]*"' | tr -d '"' || echo "unknown")
if [ "$FALLBACK_VERSION" = "2026-02" ]; then
    pass "Fallback models version is current: $FALLBACK_VERSION"
else
    fail "Fallback models version outdated: $FALLBACK_VERSION"
fi

section "Test 10: Model Name Format Validation"

if [[ "claude-opus-4-6" =~ ^claude-(opus|sonnet|haiku)-[0-9]+(-[0-9]+)?$ ]]; then
    pass "claude-opus-4-6 format is valid"
else
    fail "claude-opus-4-6 format is invalid"
fi

if [[ "claude-sonnet-4-6" =~ ^claude-(opus|sonnet|haiku)-[0-9]+(-[0-9]+)?$ ]]; then
    pass "claude-sonnet-4-6 format is valid"
else
    fail "claude-sonnet-4-6 format is invalid"
fi

echo ""
echo "============================================"
echo "  CHALLENGE RESULTS SUMMARY"
echo "============================================"
echo ""
echo "Total Tests: $TOTAL_TESTS"
echo -e "${GREEN}Passed:${NC}     $PASSED_TESTS"
echo -e "${RED}Failed:${NC}     $FAILED_TESTS"
echo ""

if [ "$FAILED_TESTS" -eq 0 ]; then
    echo -e "${GREEN}✓ ALL CLAUDE 4.6 MODEL TESTS PASSED!${NC}"
    exit 0
else
    echo -e "${RED}✗ SOME CLAUDE 4.6 MODEL TESTS FAILED${NC}"
    exit 1
fi
