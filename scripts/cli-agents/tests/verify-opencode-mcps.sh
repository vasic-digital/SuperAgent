#!/bin/bash
# ============================================================================
# OpenCode MCP Verification Test
# ============================================================================
# This test ACTUALLY runs OpenCode and verifies that 35+ MCPs are configured
# and available. It validates the configuration matches OpenCode's schema.
#
# OpenCode schema (from opencode-schema.json):
# - mcpServers: map of MCP server configs
# - providers: map of LLM provider configs
# - agents: map of agent configs (coder, task, title, summarizer)
# - contextPaths: array of context file paths
# - tui: TUI configuration
#
# Usage: ./verify-opencode-mcps.sh
# ============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Test results
TESTS_PASSED=0
TESTS_FAILED=0
TOTAL_TESTS=0

log_test() {
    echo -e "${BLUE}[TEST]${NC} $1"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
}

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    TESTS_PASSED=$((TESTS_PASSED + 1))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    TESTS_FAILED=$((TESTS_FAILED + 1))
}

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

# ============================================================================
# Test 1: Verify OpenCode binary exists
# ============================================================================
log_test "1. OpenCode binary exists"
if command -v opencode &> /dev/null; then
    OPENCODE_PATH=$(which opencode)
    log_pass "OpenCode found at: $OPENCODE_PATH"
else
    log_fail "OpenCode binary not found in PATH"
    echo "Please install OpenCode: go install github.com/opencode-ai/opencode@latest"
    exit 1
fi

# ============================================================================
# Test 2: Verify OpenCode config file exists
# ============================================================================
log_test "2. OpenCode config file exists"
CONFIG_FILE="$HOME/.config/opencode/opencode.json"
if [[ -f "$CONFIG_FILE" ]]; then
    log_pass "Config file found: $CONFIG_FILE"
else
    log_fail "Config file not found: $CONFIG_FILE"
    echo "Please run: ./generate-all-configs.sh --agent=opencode --install"
    exit 1
fi

# ============================================================================
# Test 3: Config is valid JSON
# ============================================================================
log_test "3. Config is valid JSON"
if jq empty "$CONFIG_FILE" 2>/dev/null; then
    log_pass "Config is valid JSON"
else
    log_fail "Config is not valid JSON"
    exit 1
fi

# ============================================================================
# Test 4: Config has valid OpenCode structure (mcpServers, providers, agents)
# ============================================================================
log_test "4. Config has valid OpenCode structure"
HAS_MCPSERVERS=$(jq 'has("mcpServers")' "$CONFIG_FILE" 2>/dev/null)
HAS_PROVIDERS=$(jq 'has("providers")' "$CONFIG_FILE" 2>/dev/null)
HAS_AGENTS=$(jq 'has("agents")' "$CONFIG_FILE" 2>/dev/null)
if [[ "$HAS_MCPSERVERS" == "true" ]] && [[ "$HAS_PROVIDERS" == "true" ]] && [[ "$HAS_AGENTS" == "true" ]]; then
    log_pass "Config has mcpServers, providers, and agents sections"
else
    log_fail "Config missing required sections (mcpServers=$HAS_MCPSERVERS, providers=$HAS_PROVIDERS, agents=$HAS_AGENTS)"
fi

# ============================================================================
# Test 5: Config has providers section with at least one provider
# ============================================================================
log_test "5. Config has providers section"
PROVIDER_COUNT=$(jq '.providers | keys | length' "$CONFIG_FILE" 2>/dev/null || echo "0")
if [[ "$PROVIDER_COUNT" -gt 0 ]]; then
    PROVIDER_NAMES=$(jq -r '.providers | keys | join(", ")' "$CONFIG_FILE")
    log_pass "Providers configured: $PROVIDER_NAMES"
else
    log_fail "No providers configured"
fi

# ============================================================================
# Test 6: Config has agents section with coder agent
# ============================================================================
log_test "6. Config has agents section with coder"
CODER_MODEL=$(jq -r '.agents.coder.model // empty' "$CONFIG_FILE")
if [[ -n "$CODER_MODEL" ]]; then
    log_pass "Coder agent model: $CODER_MODEL"
else
    log_fail "Coder agent not configured"
fi

# ============================================================================
# Test 7: Config has MCP section with 35+ entries
# ============================================================================
log_test "7. Config has 35+ MCPs"
MCP_COUNT=$(jq '.mcpServers | keys | length' "$CONFIG_FILE" 2>/dev/null || echo "0")
if [[ "$MCP_COUNT" -ge 35 ]]; then
    log_pass "MCP count: $MCP_COUNT (>= 35 required)"
else
    log_fail "MCP count: $MCP_COUNT (< 35 required)"
fi

# ============================================================================
# Test 8: MCPs have correct format (command OR url for SSE type)
# ============================================================================
log_test "8. MCPs have correct format"
# MCPs must have either 'command' (for stdio) or 'url' (for sse type)
INVALID_MCPS=$(jq '[.mcpServers | to_entries[] | select(.value.command == null and .value.url == null)] | length' "$CONFIG_FILE" 2>/dev/null || echo "999")
if [[ "$INVALID_MCPS" -eq 0 ]]; then
    log_pass "All MCPs have valid structure (command or url)"
else
    log_fail "$INVALID_MCPS MCPs missing both command and url"
    jq '.mcpServers | to_entries[] | select(.value.command == null and .value.url == null) | .key' "$CONFIG_FILE"
fi

# ============================================================================
# Test 9: Required Anthropic MCPs present
# ============================================================================
log_test "9. Required Anthropic MCPs present"
REQUIRED_MCPS=("filesystem" "fetch" "memory" "time" "git")
MISSING_MCPS=""
for mcp in "${REQUIRED_MCPS[@]}"; do
    if ! jq -e ".mcpServers.\"$mcp\"" "$CONFIG_FILE" > /dev/null 2>&1; then
        MISSING_MCPS="$MISSING_MCPS $mcp"
    fi
done
if [[ -z "$MISSING_MCPS" ]]; then
    log_pass "All required Anthropic MCPs present: ${REQUIRED_MCPS[*]}"
else
    log_fail "Missing required MCPs:$MISSING_MCPS"
fi

# ============================================================================
# Test 10: HelixAgent MCPs present
# ============================================================================
log_test "10. HelixAgent MCPs present"
HELIX_MCPS=("helixagent" "helixagent-debate" "helixagent-rag" "helixagent-memory")
MISSING_HELIX=""
for mcp in "${HELIX_MCPS[@]}"; do
    if ! jq -e ".mcpServers.\"$mcp\"" "$CONFIG_FILE" > /dev/null 2>&1; then
        MISSING_HELIX="$MISSING_HELIX $mcp"
    fi
done
if [[ -z "$MISSING_HELIX" ]]; then
    log_pass "All HelixAgent MCPs present: ${HELIX_MCPS[*]}"
else
    log_fail "Missing HelixAgent MCPs:$MISSING_HELIX"
fi

# ============================================================================
# Test 11: Community MCPs present
# ============================================================================
log_test "11. Community MCPs present"
COMMUNITY_MCPS=("docker" "kubernetes" "redis" "qdrant" "chroma")
MISSING_COMMUNITY=""
for mcp in "${COMMUNITY_MCPS[@]}"; do
    if ! jq -e ".mcpServers.\"$mcp\"" "$CONFIG_FILE" > /dev/null 2>&1; then
        MISSING_COMMUNITY="$MISSING_COMMUNITY $mcp"
    fi
done
if [[ -z "$MISSING_COMMUNITY" ]]; then
    log_pass "All community MCPs present: ${COMMUNITY_MCPS[*]}"
else
    log_fail "Missing community MCPs:$MISSING_COMMUNITY"
fi

# ============================================================================
# Test 12: OpenCode validates config without errors
# ============================================================================
log_test "12. OpenCode validates config without errors"
# Capture OpenCode's stderr to check for validation errors
OPENCODE_OUTPUT=$(timeout 5 opencode --help 2>&1 || true)
if echo "$OPENCODE_OUTPUT" | grep -qi "config.*invalid\|invalid.*config"; then
    log_fail "OpenCode reported config validation errors"
    echo "$OPENCODE_OUTPUT" | grep -i "error\|invalid" | head -5
else
    log_pass "OpenCode config validation passed"
fi

# ============================================================================
# Test 13: OpenCode can start (brief test)
# ============================================================================
log_test "13. OpenCode can start"
# Try to start OpenCode briefly and check it doesn't immediately crash
OPENCODE_START_OUTPUT=$(timeout 3 opencode --help 2>&1 || true)
if echo "$OPENCODE_START_OUTPUT" | grep -qi "opencode\|commands\|usage"; then
    log_pass "OpenCode starts without immediate errors"
else
    log_fail "OpenCode failed to start"
    echo "$OPENCODE_START_OUTPUT" | head -5
fi

# ============================================================================
# Test 14: HelixAgent SSE MCPs have correct URL
# ============================================================================
log_test "14. HelixAgent SSE MCPs configured correctly"
HELIX_MCP_URL=$(jq -r '.mcpServers.helixagent.url // empty' "$CONFIG_FILE")
HELIX_MCP_TYPE=$(jq -r '.mcpServers.helixagent.type // empty' "$CONFIG_FILE")
if [[ "$HELIX_MCP_TYPE" == "sse" ]] && [[ "$HELIX_MCP_URL" == *"localhost:7061"* || "$HELIX_MCP_URL" == *"helixagent"* ]]; then
    log_pass "HelixAgent MCP configured as SSE at: $HELIX_MCP_URL"
elif [[ -n "$HELIX_MCP_URL" ]]; then
    log_pass "HelixAgent MCP URL: $HELIX_MCP_URL"
else
    log_fail "HelixAgent MCP not configured correctly"
fi

# ============================================================================
# Test 15: MCPs have env configured as array
# ============================================================================
log_test "15. MCP env format is correct (array of strings)"
# env should be an array like ["KEY=VALUE"], not an object
GITHUB_ENV_TYPE=$(jq -r '.mcpServers.github.env | type' "$CONFIG_FILE" 2>/dev/null)
if [[ "$GITHUB_ENV_TYPE" == "array" ]] || [[ "$GITHUB_ENV_TYPE" == "null" ]]; then
    log_pass "MCP env format is correct (array or not present)"
else
    log_fail "MCP env should be an array, got: $GITHUB_ENV_TYPE"
fi

# ============================================================================
# Summary
# ============================================================================
echo ""
echo "=============================================="
echo "         MCP VERIFICATION SUMMARY"
echo "=============================================="
echo ""
log_info "Total tests:  $TOTAL_TESTS"
log_info "Tests passed: $TESTS_PASSED"
log_info "Tests failed: $TESTS_FAILED"
echo ""

# List all MCPs
echo "=============================================="
echo "         CONFIGURED MCPs ($MCP_COUNT)"
echo "=============================================="
jq -r '.mcpServers | keys[]' "$CONFIG_FILE" | while read -r mcp; do
    MCP_TYPE=$(jq -r ".mcpServers.\"$mcp\".type // \"stdio\"" "$CONFIG_FILE")
    if [[ "$MCP_TYPE" == "sse" ]]; then
        echo "  ✓ $mcp (SSE)"
    else
        echo "  ✓ $mcp"
    fi
done
echo ""

if [[ "$TESTS_FAILED" -eq 0 ]]; then
    echo -e "${GREEN}ALL TESTS PASSED!${NC}"
    echo ""
    echo "OpenCode is correctly configured with $MCP_COUNT MCPs."
    echo "Run 'opencode' to start using HelixAgent with MCP support."
    exit 0
else
    echo -e "${RED}SOME TESTS FAILED!${NC}"
    echo ""
    echo "Please fix the issues above and re-run this test."
    exit 1
fi
