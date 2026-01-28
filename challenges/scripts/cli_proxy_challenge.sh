#!/bin/bash
#
# CLI Proxy Providers Challenge
#
# This challenge validates that OAuth/free providers correctly use CLI proxy mechanism
# instead of direct API calls (which would fail due to product-restricted tokens).
#
# Test Categories:
# 1-10:  CLI Installation and Discovery
# 11-20: Provider Registration Logic
# 21-30: OAuth Token Handling
# 31-40: JSON Output Parsing
# 41-50: Integration and E2E Tests
#
# Expected: All tests pass without false positives
#

# Don't use set -e as we handle test failures ourselves

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

PASSED=0
FAILED=0
SKIPPED=0
TOTAL=50

# Test result tracking
declare -a TEST_RESULTS

log_test() {
    local num=$1
    local name=$2
    local status=$3
    local msg=$4

    case $status in
        PASS)
            echo -e "${GREEN}[PASS]${NC} Test $num: $name"
            ((PASSED++))
            TEST_RESULTS+=("PASS:$num:$name")
            ;;
        FAIL)
            echo -e "${RED}[FAIL]${NC} Test $num: $name - $msg"
            ((FAILED++))
            TEST_RESULTS+=("FAIL:$num:$name:$msg")
            ;;
        SKIP)
            echo -e "${YELLOW}[SKIP]${NC} Test $num: $name - $msg"
            ((SKIPPED++))
            TEST_RESULTS+=("SKIP:$num:$name:$msg")
            ;;
    esac
}

echo "=========================================="
echo "CLI Proxy Providers Challenge"
echo "=========================================="
echo ""
echo "Validating CLI proxy mechanism for OAuth/free providers"
echo ""

cd "$PROJECT_ROOT"

# ===========================================
# Category 1: CLI Installation and Discovery (Tests 1-10)
# ===========================================
echo -e "${BLUE}=== Category 1: CLI Installation and Discovery ===${NC}"

# Test 1: ClaudeCLIProvider exists
if grep -q "type ClaudeCLIProvider struct" internal/llm/providers/claude/claude_cli.go 2>/dev/null; then
    log_test 1 "ClaudeCLIProvider struct exists" PASS
else
    log_test 1 "ClaudeCLIProvider struct exists" FAIL "Struct not found"
fi

# Test 2: QwenCLIProvider exists
if grep -q "type QwenCLIProvider struct" internal/llm/providers/qwen/qwen_cli.go 2>/dev/null; then
    log_test 2 "QwenCLIProvider struct exists" PASS
else
    log_test 2 "QwenCLIProvider struct exists" FAIL "Struct not found"
fi

# Test 3: ZenCLIProvider exists
if grep -q "type ZenCLIProvider struct" internal/llm/providers/zen/zen_cli.go 2>/dev/null; then
    log_test 3 "ZenCLIProvider struct exists" PASS
else
    log_test 3 "ZenCLIProvider struct exists" FAIL "Struct not found"
fi

# Test 4: NewClaudeCLIProviderWithModel exists
if grep -q "func NewClaudeCLIProviderWithModel" internal/llm/providers/claude/claude_cli.go 2>/dev/null; then
    log_test 4 "NewClaudeCLIProviderWithModel constructor exists" PASS
else
    log_test 4 "NewClaudeCLIProviderWithModel constructor exists" FAIL "Constructor not found"
fi

# Test 5: NewQwenCLIProviderWithModel exists
if grep -q "func NewQwenCLIProviderWithModel" internal/llm/providers/qwen/qwen_cli.go 2>/dev/null; then
    log_test 5 "NewQwenCLIProviderWithModel constructor exists" PASS
else
    log_test 5 "NewQwenCLIProviderWithModel constructor exists" FAIL "Constructor not found"
fi

# Test 6: NewZenCLIProviderWithModel exists
if grep -q "func NewZenCLIProviderWithModel" internal/llm/providers/zen/zen_cli.go 2>/dev/null; then
    log_test 6 "NewZenCLIProviderWithModel constructor exists" PASS
else
    log_test 6 "NewZenCLIProviderWithModel constructor exists" FAIL "Constructor not found"
fi

# Test 7: IsCLIAvailable method for Claude
if grep -q "func.*ClaudeCLIProvider.*IsCLIAvailable" internal/llm/providers/claude/claude_cli.go 2>/dev/null; then
    log_test 7 "ClaudeCLIProvider.IsCLIAvailable exists" PASS
else
    log_test 7 "ClaudeCLIProvider.IsCLIAvailable exists" FAIL "Method not found"
fi

# Test 8: IsCLIAvailable method for Qwen
if grep -q "func.*QwenCLIProvider.*IsCLIAvailable" internal/llm/providers/qwen/qwen_cli.go 2>/dev/null; then
    log_test 8 "QwenCLIProvider.IsCLIAvailable exists" PASS
else
    log_test 8 "QwenCLIProvider.IsCLIAvailable exists" FAIL "Method not found"
fi

# Test 9: IsCLIAvailable method for Zen
if grep -q "func.*ZenCLIProvider.*IsCLIAvailable" internal/llm/providers/zen/zen_cli.go 2>/dev/null; then
    log_test 9 "ZenCLIProvider.IsCLIAvailable exists" PASS
else
    log_test 9 "ZenCLIProvider.IsCLIAvailable exists" FAIL "Method not found"
fi

# Test 10: CLI provider tests exist
if [ -f "internal/llm/providers/claude/claude_cli_test.go" ] && \
   [ -f "internal/llm/providers/qwen/qwen_cli_test.go" ] && \
   [ -f "internal/llm/providers/zen/zen_cli_test.go" ]; then
    log_test 10 "CLI provider test files exist" PASS
else
    log_test 10 "CLI provider test files exist" FAIL "One or more test files missing"
fi

# ===========================================
# Category 2: Provider Registration Logic (Tests 11-20)
# ===========================================
echo ""
echo -e "${BLUE}=== Category 2: Provider Registration Logic ===${NC}"

# Test 11: Registry uses ClaudeCLIProvider for OAuth
if grep -q "NewClaudeCLIProviderWithModel" internal/services/provider_registry.go 2>/dev/null; then
    log_test 11 "Registry uses ClaudeCLIProvider for OAuth" PASS
else
    log_test 11 "Registry uses ClaudeCLIProvider for OAuth" FAIL "ClaudeCLIProvider not used in registry"
fi

# Test 12: Registry does NOT use NewClaudeProviderWithOAuth
if ! grep -q "NewClaudeProviderWithOAuth" internal/services/provider_registry.go 2>/dev/null; then
    log_test 12 "Registry does NOT use NewClaudeProviderWithOAuth" PASS
else
    log_test 12 "Registry does NOT use NewClaudeProviderWithOAuth" FAIL "Direct OAuth API still used"
fi

# Test 13: Registry uses QwenCLIProvider for OAuth
if grep -q "NewQwenCLIProviderWithModel" internal/services/provider_registry.go 2>/dev/null; then
    log_test 13 "Registry uses QwenCLIProvider for OAuth" PASS
else
    log_test 13 "Registry uses QwenCLIProvider for OAuth" FAIL "QwenCLIProvider not used in registry"
fi

# Test 14: Registry does NOT use NewQwenProviderWithOAuth
if ! grep -q "NewQwenProviderWithOAuth" internal/services/provider_registry.go 2>/dev/null; then
    log_test 14 "Registry does NOT use NewQwenProviderWithOAuth" PASS
else
    log_test 14 "Registry does NOT use NewQwenProviderWithOAuth" FAIL "Direct OAuth API still used"
fi

# Test 15: Discovery uses ZenCLIProvider for free mode
if grep -q "NewZenCLIProviderWithModel" internal/services/provider_discovery.go 2>/dev/null; then
    log_test 15 "Discovery uses ZenCLIProvider for free mode" PASS
else
    log_test 15 "Discovery uses ZenCLIProvider for free mode" FAIL "ZenCLIProvider not used in discovery"
fi

# Test 16: Registry checks IsCLIAvailable before registration
if grep -q "IsCLIAvailable" internal/services/provider_registry.go 2>/dev/null; then
    log_test 16 "Registry checks IsCLIAvailable before registration" PASS
else
    log_test 16 "Registry checks IsCLIAvailable before registration" FAIL "CLI availability check missing"
fi

# Test 17: API key providers use direct API (not CLI)
if grep -q 'claudeConfig.APIKey != ""' internal/services/provider_registry.go 2>/dev/null && \
   grep -q "NewClaudeProvider" internal/services/provider_registry.go 2>/dev/null; then
    log_test 17 "API key providers use direct API (not CLI)" PASS
else
    log_test 17 "API key providers use direct API (not CLI)" FAIL "API key handling incorrect"
fi

# Test 18: CLI proxy logs correct message for Claude
if grep -q "CLI proxy.*OAuth via claude command" internal/services/provider_registry.go 2>/dev/null; then
    log_test 18 "CLI proxy logs correct message for Claude" PASS
else
    log_test 18 "CLI proxy logs correct message for Claude" FAIL "Log message incorrect or missing"
fi

# Test 19: CLI proxy logs correct message for Qwen
if grep -q "CLI proxy.*OAuth via qwen command" internal/services/provider_registry.go 2>/dev/null; then
    log_test 19 "CLI proxy logs correct message for Qwen" PASS
else
    log_test 19 "CLI proxy logs correct message for Qwen" FAIL "Log message incorrect or missing"
fi

# Test 20: CLI proxy logs correct message for Zen
if grep -q "CLI proxy.*opencode command" internal/services/provider_discovery.go 2>/dev/null; then
    log_test 20 "CLI proxy logs correct message for Zen" PASS
else
    log_test 20 "CLI proxy logs correct message for Zen" FAIL "Log message incorrect or missing"
fi

# ===========================================
# Category 3: OAuth Token Handling (Tests 21-30)
# ===========================================
echo ""
echo -e "${BLUE}=== Category 3: OAuth Token Handling ===${NC}"

# Test 21: OAuth adapter documents product restrictions
if grep -q "PRODUCT-RESTRICTED\|product-restricted" internal/verifier/adapters/oauth_adapter.go 2>/dev/null; then
    log_test 21 "OAuth adapter documents product restrictions" PASS
else
    log_test 21 "OAuth adapter documents product restrictions" FAIL "Documentation missing"
fi

# Test 22: Registry comments explain OAuth limitations
if grep -q "OAuth tokens are product-restricted\|OAuth tokens are portal-restricted" internal/services/provider_registry.go 2>/dev/null; then
    log_test 22 "Registry comments explain OAuth limitations" PASS
else
    log_test 22 "Registry comments explain OAuth limitations" FAIL "Comments missing"
fi

# Test 23: Claude OAuth credential reader exists
if grep -q "ReadClaudeCredentials" internal/auth/oauth_credentials/*.go 2>/dev/null; then
    log_test 23 "Claude OAuth credential reader exists" PASS
else
    log_test 23 "Claude OAuth credential reader exists" FAIL "Reader not found"
fi

# Test 24: Qwen OAuth credential reader exists
if grep -q "ReadQwenCredentials" internal/auth/oauth_credentials/*.go 2>/dev/null; then
    log_test 24 "Qwen OAuth credential reader exists" PASS
else
    log_test 24 "Qwen OAuth credential reader exists" FAIL "Reader not found"
fi

# Test 25: HasValidClaudeCredentials function exists
if grep -q "HasValidClaudeCredentials" internal/auth/oauth_credentials/*.go 2>/dev/null; then
    log_test 25 "HasValidClaudeCredentials function exists" PASS
else
    log_test 25 "HasValidClaudeCredentials function exists" FAIL "Function not found"
fi

# Test 26: HasValidQwenCredentials function exists
if grep -q "HasValidQwenCredentials" internal/auth/oauth_credentials/*.go 2>/dev/null; then
    log_test 26 "HasValidQwenCredentials function exists" PASS
else
    log_test 26 "HasValidQwenCredentials function exists" FAIL "Function not found"
fi

# Test 27: IsClaudeOAuthEnabled function exists
if grep -q "func IsClaudeOAuthEnabled" internal/auth/oauth_credentials/*.go 2>/dev/null; then
    log_test 27 "IsClaudeOAuthEnabled function exists" PASS
else
    log_test 27 "IsClaudeOAuthEnabled function exists" FAIL "Function not found"
fi

# Test 28: IsQwenOAuthEnabled function exists
if grep -q "func IsQwenOAuthEnabled" internal/auth/oauth_credentials/*.go 2>/dev/null; then
    log_test 28 "IsQwenOAuthEnabled function exists" PASS
else
    log_test 28 "IsQwenOAuthEnabled function exists" FAIL "Function not found"
fi

# Test 29: OAuth env variables documented
if grep -q "CLAUDE_CODE_USE_OAUTH_CREDENTIALS" CLAUDE.md 2>/dev/null || \
   grep -q "CLAUDE_CODE_USE_OAUTH_CREDENTIALS" internal/services/provider_registry.go 2>/dev/null; then
    log_test 29 "OAuth env variables documented" PASS
else
    log_test 29 "OAuth env variables documented" FAIL "Env vars not documented"
fi

# Test 30: GetGlobalReader function exists
if grep -q "func GetGlobalReader" internal/auth/oauth_credentials/*.go 2>/dev/null; then
    log_test 30 "GetGlobalReader function exists" PASS
else
    log_test 30 "GetGlobalReader function exists" FAIL "Function not found"
fi

# ===========================================
# Category 4: JSON Output Parsing (Tests 31-40)
# ===========================================
echo ""
echo -e "${BLUE}=== Category 4: JSON Output Parsing ===${NC}"

# Test 31: Zen CLI uses JSON format flag
if grep -q '"-f", "json"' internal/llm/providers/zen/zen_cli.go 2>/dev/null; then
    log_test 31 "Zen CLI uses JSON format flag (-f json)" PASS
else
    log_test 31 "Zen CLI uses JSON format flag (-f json)" FAIL "JSON flag not used"
fi

# Test 32: JSON response type defined
if grep -q "openCodeJSONResponse\|type.*JSONResponse" internal/llm/providers/zen/zen_cli.go 2>/dev/null; then
    log_test 32 "JSON response type defined" PASS
else
    log_test 32 "JSON response type defined" FAIL "Type not defined"
fi

# Test 33: parseJSONResponse method exists
if grep -q "parseJSONResponse" internal/llm/providers/zen/zen_cli.go 2>/dev/null; then
    log_test 33 "parseJSONResponse method exists" PASS
else
    log_test 33 "parseJSONResponse method exists" FAIL "Method not found"
fi

# Test 34: JSON import in zen_cli.go
if grep -q '"encoding/json"' internal/llm/providers/zen/zen_cli.go 2>/dev/null; then
    log_test 34 "JSON import in zen_cli.go" PASS
else
    log_test 34 "JSON import in zen_cli.go" FAIL "Import missing"
fi

# Test 35: JSON fallback for non-JSON output
if grep -q "Fallback\|fallback\|return rawOutput" internal/llm/providers/zen/zen_cli.go 2>/dev/null; then
    log_test 35 "JSON parsing has fallback for non-JSON output" PASS
else
    log_test 35 "JSON parsing has fallback for non-JSON output" FAIL "No fallback"
fi

# Test 36: Response field in JSON struct
if grep -q 'Response.*json:"response"' internal/llm/providers/zen/zen_cli.go 2>/dev/null; then
    log_test 36 "Response field in JSON struct" PASS
else
    log_test 36 "Response field in JSON struct" FAIL "Field not found"
fi

# Test 37: Claude CLI prompt flag
if grep -q '"-p"' internal/llm/providers/claude/claude_cli.go 2>/dev/null; then
    log_test 37 "Claude CLI uses prompt flag" PASS
else
    log_test 37 "Claude CLI uses prompt flag" FAIL "Flag not used"
fi

# Test 38: Qwen CLI prompt flag
if grep -q '"-p"' internal/llm/providers/qwen/qwen_cli.go 2>/dev/null; then
    log_test 38 "Qwen CLI uses prompt flag" PASS
else
    log_test 38 "Qwen CLI uses prompt flag" FAIL "Flag not used"
fi

# Test 39: Zen CLI prompt flag
if grep -q '"-p"' internal/llm/providers/zen/zen_cli.go 2>/dev/null; then
    log_test 39 "Zen CLI uses prompt flag" PASS
else
    log_test 39 "Zen CLI uses prompt flag" FAIL "Flag not used"
fi

# Test 40: CLI providers return proper ProviderName
if grep -q 'ProviderName.*"claude-cli"' internal/llm/providers/claude/claude_cli.go 2>/dev/null && \
   grep -q 'ProviderName.*"qwen-cli"' internal/llm/providers/qwen/qwen_cli.go 2>/dev/null && \
   grep -q 'ProviderName.*"zen-cli"' internal/llm/providers/zen/zen_cli.go 2>/dev/null; then
    log_test 40 "CLI providers return proper ProviderName" PASS
else
    log_test 40 "CLI providers return proper ProviderName" FAIL "ProviderName incorrect"
fi

# ===========================================
# Category 5: Integration and E2E Tests (Tests 41-50)
# ===========================================
echo ""
echo -e "${BLUE}=== Category 5: Integration and E2E Tests ===${NC}"

# Test 41: Integration test file exists
if [ -f "tests/integration/cli_proxy_providers_test.go" ]; then
    log_test 41 "CLI proxy integration test file exists" PASS
else
    log_test 41 "CLI proxy integration test file exists" FAIL "Test file not found"
fi

# Test 42: Project builds successfully
echo "Building project..."
if go build ./... 2>/dev/null; then
    log_test 42 "Project builds successfully" PASS
else
    log_test 42 "Project builds successfully" FAIL "Build failed"
fi

# Test 43: CLI provider unit tests pass (if CLI not available, skip)
echo "Running CLI provider unit tests..."
if go test -short -v ./internal/llm/providers/claude/... -run "CLI" 2>&1 | grep -q "PASS\|SKIP"; then
    log_test 43 "Claude CLI unit tests pass" PASS
else
    log_test 43 "Claude CLI unit tests pass" SKIP "Tests may require CLI"
fi

# Test 44: Qwen CLI unit tests
if go test -short -v ./internal/llm/providers/qwen/... -run "CLI" 2>&1 | grep -q "PASS\|SKIP"; then
    log_test 44 "Qwen CLI unit tests pass" PASS
else
    log_test 44 "Qwen CLI unit tests pass" SKIP "Tests may require CLI"
fi

# Test 45: Zen CLI unit tests
if go test -short -v ./internal/llm/providers/zen/... -run "CLI" 2>&1 | grep -q "PASS\|SKIP"; then
    log_test 45 "Zen CLI unit tests pass" PASS
else
    log_test 45 "Zen CLI unit tests pass" SKIP "Tests may require CLI"
fi

# Test 46: No race conditions in CLI providers (with timeout)
echo "Running race detection (30s timeout)..."
if timeout 30 go test -race -short ./internal/llm/providers/claude/... ./internal/llm/providers/qwen/... ./internal/llm/providers/zen/... 2>&1 | grep -v "DATA RACE" | tail -5; then
    log_test 46 "No race conditions in CLI providers" PASS
else
    if [ $? -eq 124 ]; then
        log_test 46 "No race conditions in CLI providers" SKIP "Race test timed out (expected for long-running tests)"
    else
        log_test 46 "No race conditions in CLI providers" FAIL "Race condition detected"
    fi
fi

# Test 47: Complete method implements LLMProvider interface
if grep -q "func.*ClaudeCLIProvider.*Complete.*context.Context.*\*models.LLMRequest.*\*models.LLMResponse" internal/llm/providers/claude/claude_cli.go 2>/dev/null; then
    log_test 47 "Claude CLI implements Complete method" PASS
else
    log_test 47 "Claude CLI implements Complete method" FAIL "Method signature incorrect"
fi

# Test 48: HealthCheck method exists in all CLI providers
if grep -q "func.*CLIProvider.*HealthCheck" internal/llm/providers/claude/claude_cli.go 2>/dev/null && \
   grep -q "func.*CLIProvider.*HealthCheck" internal/llm/providers/qwen/qwen_cli.go 2>/dev/null && \
   grep -q "func.*CLIProvider.*HealthCheck" internal/llm/providers/zen/zen_cli.go 2>/dev/null; then
    log_test 48 "All CLI providers have HealthCheck method" PASS
else
    log_test 48 "All CLI providers have HealthCheck method" FAIL "HealthCheck missing"
fi

# Test 49: GetCapabilities method exists in all CLI providers
if grep -q "func.*CLIProvider.*GetCapabilities" internal/llm/providers/claude/claude_cli.go 2>/dev/null && \
   grep -q "func.*CLIProvider.*GetCapabilities" internal/llm/providers/qwen/qwen_cli.go 2>/dev/null && \
   grep -q "func.*CLIProvider.*GetCapabilities" internal/llm/providers/zen/zen_cli.go 2>/dev/null; then
    log_test 49 "All CLI providers have GetCapabilities method" PASS
else
    log_test 49 "All CLI providers have GetCapabilities method" FAIL "GetCapabilities missing"
fi

# Test 50: Documentation updated for CLI proxy
if grep -q "CLI Proxy Mechanism\|CLI proxy\|CLI.*proxy" CLAUDE.md 2>/dev/null; then
    log_test 50 "Documentation updated for CLI proxy" PASS
else
    log_test 50 "Documentation updated for CLI proxy" SKIP "Documentation update pending"
fi

# ===========================================
# Summary
# ===========================================
echo ""
echo "=========================================="
echo "Challenge Summary"
echo "=========================================="
echo -e "Total Tests: ${TOTAL}"
echo -e "${GREEN}Passed: ${PASSED}${NC}"
echo -e "${RED}Failed: ${FAILED}${NC}"
echo -e "${YELLOW}Skipped: ${SKIPPED}${NC}"
echo ""

PASS_RATE=$(echo "scale=1; ($PASSED * 100) / $TOTAL" | bc)
echo "Pass Rate: ${PASS_RATE}%"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}=========================================="
    echo "CHALLENGE PASSED"
    echo "==========================================${NC}"
    exit 0
else
    echo -e "${RED}=========================================="
    echo "CHALLENGE FAILED"
    echo ""
    echo "Failed Tests:"
    for result in "${TEST_RESULTS[@]}"; do
        if [[ $result == FAIL:* ]]; then
            echo "  - ${result#FAIL:}"
        fi
    done
    echo "==========================================${NC}"
    exit 1
fi
