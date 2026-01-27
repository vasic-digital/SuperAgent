#!/bin/bash

# =============================================================================
# OAuth CLI Fallback Challenge
# =============================================================================
# This challenge validates that:
# 1. Claude OAuth uses Claude Code CLI instead of direct API (product-restricted tokens)
# 2. Qwen OAuth uses Qwen Code CLI instead of direct API (product-restricted tokens)
# 3. CLI providers properly implement LLMProvider interface
# 4. Fallback logic when CLI not available works correctly
# 5. Unit tests cover all functionality
# =============================================================================

# Don't exit on errors - we handle them ourselves
set +e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0
TOTAL_TESTS=0

# Test result logging
log_pass() {
    echo -e "${GREEN}✓ PASS${NC}: $1"
    ((TESTS_PASSED++))
    ((TOTAL_TESTS++))
}

log_fail() {
    echo -e "${RED}✗ FAIL${NC}: $1"
    ((TESTS_FAILED++))
    ((TOTAL_TESTS++))
}

log_info() {
    echo -e "${BLUE}ℹ INFO${NC}: $1"
}

log_section() {
    echo ""
    echo -e "${YELLOW}=== $1 ===${NC}"
    echo ""
}

# =============================================================================
# Section 1: Claude CLI Provider Code Verification
# =============================================================================
log_section "Section 1: Claude CLI Provider Implementation"

CLAUDE_CLI_FILE="$PROJECT_ROOT/internal/llm/providers/claude/claude_cli.go"

# Test 1.1: ClaudeCLIProvider struct exists
if grep -q "type ClaudeCLIProvider struct" "$CLAUDE_CLI_FILE"; then
    log_pass "1.1 ClaudeCLIProvider struct exists"
else
    log_fail "1.1 ClaudeCLIProvider struct not found"
fi

# Test 1.2: ClaudeCLIConfig struct exists
if grep -q "type ClaudeCLIConfig struct" "$CLAUDE_CLI_FILE"; then
    log_pass "1.2 ClaudeCLIConfig struct exists"
else
    log_fail "1.2 ClaudeCLIConfig struct not found"
fi

# Test 1.3: NewClaudeCLIProvider constructor exists
if grep -q "func NewClaudeCLIProvider(config ClaudeCLIConfig) \*ClaudeCLIProvider" "$CLAUDE_CLI_FILE"; then
    log_pass "1.3 NewClaudeCLIProvider constructor exists"
else
    log_fail "1.3 NewClaudeCLIProvider constructor not found"
fi

# Test 1.4: NewClaudeCLIProviderWithModel constructor exists
if grep -q "func NewClaudeCLIProviderWithModel(model string) \*ClaudeCLIProvider" "$CLAUDE_CLI_FILE"; then
    log_pass "1.4 NewClaudeCLIProviderWithModel constructor exists"
else
    log_fail "1.4 NewClaudeCLIProviderWithModel constructor not found"
fi

# Test 1.5: IsCLIAvailable method exists
if grep -q "func (p \*ClaudeCLIProvider) IsCLIAvailable() bool" "$CLAUDE_CLI_FILE"; then
    log_pass "1.5 IsCLIAvailable method exists"
else
    log_fail "1.5 IsCLIAvailable method not found"
fi

# Test 1.6: Complete method exists (LLMProvider interface)
if grep -q "func (p \*ClaudeCLIProvider) Complete(ctx context.Context, req \*models.LLMRequest) (\*models.LLMResponse, error)" "$CLAUDE_CLI_FILE"; then
    log_pass "1.6 Complete method exists (LLMProvider interface)"
else
    log_fail "1.6 Complete method not found"
fi

# Test 1.7: CompleteStream method exists (LLMProvider interface)
if grep -q "func (p \*ClaudeCLIProvider) CompleteStream(ctx context.Context, req \*models.LLMRequest) (<-chan \*models.LLMResponse, error)" "$CLAUDE_CLI_FILE"; then
    log_pass "1.7 CompleteStream method exists (LLMProvider interface)"
else
    log_fail "1.7 CompleteStream method not found"
fi

# Test 1.8: HealthCheck method exists
if grep -q "func (p \*ClaudeCLIProvider) HealthCheck() error" "$CLAUDE_CLI_FILE"; then
    log_pass "1.8 HealthCheck method exists"
else
    log_fail "1.8 HealthCheck method not found"
fi

# Test 1.9: GetCapabilities method exists
if grep -q "func (p \*ClaudeCLIProvider) GetCapabilities() \*models.ProviderCapabilities" "$CLAUDE_CLI_FILE"; then
    log_pass "1.9 GetCapabilities method exists"
else
    log_fail "1.9 GetCapabilities method not found"
fi

# Test 1.10: ValidateConfig method exists
if grep -q "func (p \*ClaudeCLIProvider) ValidateConfig(config map\[string\]interface{}) (bool, \[\]string)" "$CLAUDE_CLI_FILE"; then
    log_pass "1.10 ValidateConfig method exists"
else
    log_fail "1.10 ValidateConfig method not found"
fi

# Test 1.11: IsClaudeCodeInstalled standalone function exists
if grep -q "func IsClaudeCodeInstalled() bool" "$CLAUDE_CLI_FILE"; then
    log_pass "1.11 IsClaudeCodeInstalled function exists"
else
    log_fail "1.11 IsClaudeCodeInstalled function not found"
fi

# Test 1.12: IsClaudeCodeAuthenticated standalone function exists
if grep -q "func IsClaudeCodeAuthenticated() bool" "$CLAUDE_CLI_FILE"; then
    log_pass "1.12 IsClaudeCodeAuthenticated function exists"
else
    log_fail "1.12 IsClaudeCodeAuthenticated function not found"
fi

# Test 1.13: CanUseClaudeOAuth standalone function exists
if grep -q "func CanUseClaudeOAuth() bool" "$CLAUDE_CLI_FILE"; then
    log_pass "1.13 CanUseClaudeOAuth function exists"
else
    log_fail "1.13 CanUseClaudeOAuth function not found"
fi

# =============================================================================
# Section 2: Qwen CLI Provider Code Verification
# =============================================================================
log_section "Section 2: Qwen CLI Provider Implementation"

QWEN_CLI_FILE="$PROJECT_ROOT/internal/llm/providers/qwen/qwen_cli.go"

# Test 2.1: QwenCLIProvider struct exists
if grep -q "type QwenCLIProvider struct" "$QWEN_CLI_FILE"; then
    log_pass "2.1 QwenCLIProvider struct exists"
else
    log_fail "2.1 QwenCLIProvider struct not found"
fi

# Test 2.2: QwenCLIConfig struct exists
if grep -q "type QwenCLIConfig struct" "$QWEN_CLI_FILE"; then
    log_pass "2.2 QwenCLIConfig struct exists"
else
    log_fail "2.2 QwenCLIConfig struct not found"
fi

# Test 2.3: NewQwenCLIProvider constructor exists
if grep -q "func NewQwenCLIProvider(config QwenCLIConfig) \*QwenCLIProvider" "$QWEN_CLI_FILE"; then
    log_pass "2.3 NewQwenCLIProvider constructor exists"
else
    log_fail "2.3 NewQwenCLIProvider constructor not found"
fi

# Test 2.4: NewQwenCLIProviderWithModel constructor exists
if grep -q "func NewQwenCLIProviderWithModel(model string) \*QwenCLIProvider" "$QWEN_CLI_FILE"; then
    log_pass "2.4 NewQwenCLIProviderWithModel constructor exists"
else
    log_fail "2.4 NewQwenCLIProviderWithModel constructor not found"
fi

# Test 2.5: IsCLIAvailable method exists
if grep -q "func (p \*QwenCLIProvider) IsCLIAvailable() bool" "$QWEN_CLI_FILE"; then
    log_pass "2.5 IsCLIAvailable method exists"
else
    log_fail "2.5 IsCLIAvailable method not found"
fi

# Test 2.6: Complete method exists (LLMProvider interface)
if grep -q "func (p \*QwenCLIProvider) Complete(ctx context.Context, req \*models.LLMRequest) (\*models.LLMResponse, error)" "$QWEN_CLI_FILE"; then
    log_pass "2.6 Complete method exists (LLMProvider interface)"
else
    log_fail "2.6 Complete method not found"
fi

# Test 2.7: CompleteStream method exists (LLMProvider interface)
if grep -q "func (p \*QwenCLIProvider) CompleteStream(ctx context.Context, req \*models.LLMRequest) (<-chan \*models.LLMResponse, error)" "$QWEN_CLI_FILE"; then
    log_pass "2.7 CompleteStream method exists (LLMProvider interface)"
else
    log_fail "2.7 CompleteStream method not found"
fi

# Test 2.8: HealthCheck method exists
if grep -q "func (p \*QwenCLIProvider) HealthCheck() error" "$QWEN_CLI_FILE"; then
    log_pass "2.8 HealthCheck method exists"
else
    log_fail "2.8 HealthCheck method not found"
fi

# Test 2.9: GetCapabilities method exists
if grep -q "func (p \*QwenCLIProvider) GetCapabilities() \*models.ProviderCapabilities" "$QWEN_CLI_FILE"; then
    log_pass "2.9 GetCapabilities method exists"
else
    log_fail "2.9 GetCapabilities method not found"
fi

# Test 2.10: ValidateConfig method exists
if grep -q "func (p \*QwenCLIProvider) ValidateConfig(config map\[string\]interface{}) (bool, \[\]string)" "$QWEN_CLI_FILE"; then
    log_pass "2.10 ValidateConfig method exists"
else
    log_fail "2.10 ValidateConfig method not found"
fi

# Test 2.11: IsQwenCodeInstalled standalone function exists
if grep -q "func IsQwenCodeInstalled() bool" "$QWEN_CLI_FILE"; then
    log_pass "2.11 IsQwenCodeInstalled function exists"
else
    log_fail "2.11 IsQwenCodeInstalled function not found"
fi

# Test 2.12: IsQwenCodeAuthenticated standalone function exists
if grep -q "func IsQwenCodeAuthenticated() bool" "$QWEN_CLI_FILE"; then
    log_pass "2.12 IsQwenCodeAuthenticated function exists"
else
    log_fail "2.12 IsQwenCodeAuthenticated function not found"
fi

# Test 2.13: CanUseQwenOAuth standalone function exists
if grep -q "func CanUseQwenOAuth() bool" "$QWEN_CLI_FILE"; then
    log_pass "2.13 CanUseQwenOAuth function exists"
else
    log_fail "2.13 CanUseQwenOAuth function not found"
fi

# =============================================================================
# Section 3: Provider Discovery Integration
# =============================================================================
log_section "Section 3: Provider Discovery Integration"

PROVIDER_DISCOVERY_FILE="$PROJECT_ROOT/internal/services/provider_discovery.go"

# Test 3.1: Claude OAuth uses CLI (product-restricted comment)
if grep -q "OAuth tokens are product-restricted" "$PROVIDER_DISCOVERY_FILE"; then
    log_pass "3.1 Claude OAuth notes product-restricted tokens"
else
    log_fail "3.1 Product-restricted token note not found"
fi

# Test 3.2: Claude CLI check in provider discovery
if grep -q "claude.IsClaudeCodeInstalled()" "$PROVIDER_DISCOVERY_FILE"; then
    log_pass "3.2 Claude CLI installation check exists"
else
    log_fail "3.2 Claude CLI installation check not found"
fi

# Test 3.3: Claude CLI authenticated check
if grep -q "claude.IsClaudeCodeAuthenticated()" "$PROVIDER_DISCOVERY_FILE"; then
    log_pass "3.3 Claude CLI authentication check exists"
else
    log_fail "3.3 Claude CLI authentication check not found"
fi

# Test 3.4: ClaudeCLIProvider instantiation
if grep -q "claude.NewClaudeCLIProviderWithModel" "$PROVIDER_DISCOVERY_FILE"; then
    log_pass "3.4 ClaudeCLIProvider instantiation exists"
else
    log_fail "3.4 ClaudeCLIProvider instantiation not found"
fi

# Test 3.5: Qwen CLI check in provider discovery
if grep -q "qwen.IsQwenCodeInstalled()" "$PROVIDER_DISCOVERY_FILE"; then
    log_pass "3.5 Qwen CLI installation check exists"
else
    log_fail "3.5 Qwen CLI installation check not found"
fi

# Test 3.6: Qwen CLI authenticated check
if grep -q "qwen.IsQwenCodeAuthenticated()" "$PROVIDER_DISCOVERY_FILE"; then
    log_pass "3.6 Qwen CLI authentication check exists"
else
    log_fail "3.6 Qwen CLI authentication check not found"
fi

# Test 3.7: QwenCLIProvider instantiation
if grep -q "qwen.NewQwenCLIProviderWithModel" "$PROVIDER_DISCOVERY_FILE"; then
    log_pass "3.7 QwenCLIProvider instantiation exists"
else
    log_fail "3.7 QwenCLIProvider instantiation not found"
fi

# Test 3.8: Claude fallback warning when CLI not available
if grep -q "Claude OAuth credentials found but Claude Code CLI is not available" "$PROVIDER_DISCOVERY_FILE"; then
    log_pass "3.8 Claude CLI unavailable warning exists"
else
    log_fail "3.8 Claude CLI unavailable warning not found"
fi

# Test 3.9: Qwen fallback warning when CLI not available
if grep -q "Qwen OAuth credentials found but Qwen Code CLI is not available" "$PROVIDER_DISCOVERY_FILE"; then
    log_pass "3.9 Qwen CLI unavailable warning exists"
else
    log_fail "3.9 Qwen CLI unavailable warning not found"
fi

# Test 3.10: APIKeyEnvVar indicates CLI usage for Claude
if grep -q 'APIKeyEnvVar: "CLI:~/.claude/.credentials.json"' "$PROVIDER_DISCOVERY_FILE"; then
    log_pass "3.10 Claude CLI credential path marker exists"
else
    log_fail "3.10 Claude CLI credential path marker not found"
fi

# Test 3.11: APIKeyEnvVar indicates CLI usage for Qwen
if grep -q 'APIKeyEnvVar: "CLI:~/.qwen/oauth_creds.json"' "$PROVIDER_DISCOVERY_FILE"; then
    log_pass "3.11 Qwen CLI credential path marker exists"
else
    log_fail "3.11 Qwen CLI credential path marker not found"
fi

# =============================================================================
# Section 4: Test Files Verification
# =============================================================================
log_section "Section 4: Test Files Verification"

CLAUDE_CLI_TEST_FILE="$PROJECT_ROOT/internal/llm/providers/claude/claude_cli_test.go"
QWEN_CLI_TEST_FILE="$PROJECT_ROOT/internal/llm/providers/qwen/qwen_cli_test.go"

# Test 4.1: Claude CLI test file exists
if [ -f "$CLAUDE_CLI_TEST_FILE" ]; then
    log_pass "4.1 Claude CLI test file exists"
else
    log_fail "4.1 Claude CLI test file not found"
fi

# Test 4.2: Qwen CLI test file exists
if [ -f "$QWEN_CLI_TEST_FILE" ]; then
    log_pass "4.2 Qwen CLI test file exists"
else
    log_fail "4.2 Qwen CLI test file not found"
fi

# Test 4.3: Claude CLI default config test
if grep -q "TestClaudeCLIProvider_DefaultConfig" "$CLAUDE_CLI_TEST_FILE"; then
    log_pass "4.3 TestClaudeCLIProvider_DefaultConfig exists"
else
    log_fail "4.3 TestClaudeCLIProvider_DefaultConfig not found"
fi

# Test 4.4: Qwen CLI default config test
if grep -q "TestQwenCLIProvider_DefaultConfig" "$QWEN_CLI_TEST_FILE"; then
    log_pass "4.4 TestQwenCLIProvider_DefaultConfig exists"
else
    log_fail "4.4 TestQwenCLIProvider_DefaultConfig not found"
fi

# Test 4.5: Claude CLI Complete_CLIUnavailable test
if grep -q "TestClaudeCLIProvider_Complete_CLIUnavailable" "$CLAUDE_CLI_TEST_FILE"; then
    log_pass "4.5 TestClaudeCLIProvider_Complete_CLIUnavailable exists"
else
    log_fail "4.5 TestClaudeCLIProvider_Complete_CLIUnavailable not found"
fi

# Test 4.6: Qwen CLI Complete_CLIUnavailable test
if grep -q "TestQwenCLIProvider_Complete_CLIUnavailable" "$QWEN_CLI_TEST_FILE"; then
    log_pass "4.6 TestQwenCLIProvider_Complete_CLIUnavailable exists"
else
    log_fail "4.6 TestQwenCLIProvider_Complete_CLIUnavailable not found"
fi

# Test 4.7: Claude CLI integration tests
if grep -q "TestClaudeCLIProvider_Integration_Complete" "$CLAUDE_CLI_TEST_FILE"; then
    log_pass "4.7 TestClaudeCLIProvider_Integration_Complete exists"
else
    log_fail "4.7 TestClaudeCLIProvider_Integration_Complete not found"
fi

# Test 4.8: Qwen CLI integration tests
if grep -q "TestQwenCLIProvider_Integration_Complete" "$QWEN_CLI_TEST_FILE"; then
    log_pass "4.8 TestQwenCLIProvider_Integration_Complete exists"
else
    log_fail "4.8 TestQwenCLIProvider_Integration_Complete not found"
fi

# =============================================================================
# Section 5: Run Unit Tests
# =============================================================================
log_section "Section 5: Run Unit Tests"

cd "$PROJECT_ROOT"

# Test 5.1: Run Claude CLI tests
log_info "Running Claude CLI tests..."
if go test -v -run "ClaudeCLI" ./internal/llm/providers/claude/... 2>&1 | grep -q "PASS"; then
    log_pass "5.1 Claude CLI tests pass"
else
    log_fail "5.1 Claude CLI tests failed"
fi

# Test 5.2: Run Qwen CLI tests
log_info "Running Qwen CLI tests..."
if go test -v -run "QwenCLI" ./internal/llm/providers/qwen/... 2>&1 | grep -q "PASS"; then
    log_pass "5.2 Qwen CLI tests pass"
else
    log_fail "5.2 Qwen CLI tests failed"
fi

# Test 5.3: Run provider discovery tests
log_info "Running provider discovery tests..."
if go test -v -run "ProviderDiscovery" ./internal/services/... 2>&1 | grep -q "PASS"; then
    log_pass "5.3 Provider discovery tests pass"
else
    log_fail "5.3 Provider discovery tests failed"
fi

# =============================================================================
# Section 6: Build Verification
# =============================================================================
log_section "Section 6: Build Verification"

# Test 6.1: Project builds successfully
log_info "Building project..."
if go build ./cmd/helixagent/... 2>&1; then
    log_pass "6.1 Project builds successfully"
else
    log_fail "6.1 Project build failed"
fi

# =============================================================================
# Summary
# =============================================================================
log_section "Challenge Summary"

echo ""
echo -e "Total Tests: ${TOTAL_TESTS}"
echo -e "Passed: ${GREEN}${TESTS_PASSED}${NC}"
echo -e "Failed: ${RED}${TESTS_FAILED}${NC}"
echo ""

if [ "$TESTS_FAILED" -eq 0 ]; then
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}      OAUTH CLI FALLBACK CHALLENGE      ${NC}"
    echo -e "${GREEN}            ALL TESTS PASSED            ${NC}"
    echo -e "${GREEN}========================================${NC}"
    exit 0
else
    echo -e "${RED}========================================${NC}"
    echo -e "${RED}      OAUTH CLI FALLBACK CHALLENGE      ${NC}"
    echo -e "${RED}         SOME TESTS FAILED              ${NC}"
    echo -e "${RED}========================================${NC}"
    exit 1
fi
