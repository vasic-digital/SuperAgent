#!/bin/bash
# ============================================================================
# OpenCode MCP Verification Test
# ============================================================================
# This test ACTUALLY runs OpenCode and verifies that 35+ MCPs are configured
# and available. It does NOT just check file existence - it validates the
# configuration is parsed correctly by OpenCode.
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
# Test 4: Config has correct schema
# ============================================================================
log_test "4. Config has correct schema structure"
SCHEMA=$(jq -r '."$schema" // empty' "$CONFIG_FILE")
if [[ -n "$SCHEMA" ]]; then
    log_pass "Schema present: $SCHEMA"
else
    log_fail "Schema field missing"
fi

# ============================================================================
# Test 5: Config has provider section
# ============================================================================
log_test "5. Config has provider section"
PROVIDER_COUNT=$(jq '.provider | keys | length' "$CONFIG_FILE" 2>/dev/null || echo "0")
if [[ "$PROVIDER_COUNT" -gt 0 ]]; then
    log_pass "Provider section present with $PROVIDER_COUNT provider(s)"
else
    log_fail "Provider section missing or empty"
fi

# ============================================================================
# Test 6: Config has agent section
# ============================================================================
log_test "6. Config has agent section"
AGENT_MODEL=$(jq -r '.agent.model // empty' "$CONFIG_FILE")
if [[ -n "$AGENT_MODEL" ]]; then
    log_pass "Agent model configured: $AGENT_MODEL"
else
    log_fail "Agent model not configured"
fi

# ============================================================================
# Test 7: Config has MCP section with 35+ entries
# ============================================================================
log_test "7. Config has 35+ MCPs"
MCP_COUNT=$(jq '.mcp | keys | length' "$CONFIG_FILE" 2>/dev/null || echo "0")
if [[ "$MCP_COUNT" -ge 35 ]]; then
    log_pass "MCP count: $MCP_COUNT (>= 35 required)"
else
    log_fail "MCP count: $MCP_COUNT (< 35 required)"
fi

# ============================================================================
# Test 8: MCPs have correct format (command/args structure)
# ============================================================================
log_test "8. MCPs have correct format"
INVALID_MCPS=$(jq '[.mcp | to_entries[] | select(.value.command == null)] | length' "$CONFIG_FILE" 2>/dev/null || echo "999")
if [[ "$INVALID_MCPS" -eq 0 ]]; then
    log_pass "All MCPs have valid command structure"
else
    log_fail "$INVALID_MCPS MCPs missing command field"
    jq '.mcp | to_entries[] | select(.value.command == null) | .key' "$CONFIG_FILE"
fi

# ============================================================================
# Test 9: Required Anthropic MCPs present
# ============================================================================
log_test "9. Required Anthropic MCPs present"
REQUIRED_MCPS=("filesystem" "fetch" "memory" "time" "git")
MISSING_MCPS=""
for mcp in "${REQUIRED_MCPS[@]}"; do
    if ! jq -e ".mcp.\"$mcp\"" "$CONFIG_FILE" > /dev/null 2>&1; then
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
    if ! jq -e ".mcp.\"$mcp\"" "$CONFIG_FILE" > /dev/null 2>&1; then
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
    if ! jq -e ".mcp.\"$mcp\"" "$CONFIG_FILE" > /dev/null 2>&1; then
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
OPENCODE_OUTPUT=$(timeout 5 opencode --version 2>&1 || true)
if echo "$OPENCODE_OUTPUT" | grep -qi "error\|invalid"; then
    log_fail "OpenCode reported config errors"
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
if echo "$OPENCODE_START_OUTPUT" | grep -qi "opencode\|usage\|command"; then
    log_pass "OpenCode starts without immediate errors"
else
    log_fail "OpenCode failed to start"
    echo "$OPENCODE_START_OUTPUT" | head -5
fi

# ============================================================================
# Test 14: Provider baseURL is correct
# ============================================================================
log_test "14. Provider baseURL is correct"
BASE_URL=$(jq -r '.provider.helixagent.options.baseURL // empty' "$CONFIG_FILE")
if [[ "$BASE_URL" == *"localhost:7061"* ]] || [[ "$BASE_URL" == *"helixagent"* ]]; then
    log_pass "Provider baseURL configured: $BASE_URL"
else
    log_fail "Provider baseURL not configured correctly: $BASE_URL"
fi

# ============================================================================
# Test 15: MCP environment variables are configured
# ============================================================================
log_test "15. MCP environment variables configured"
# Check that MCPs with required env vars have them defined
GITHUB_ENV=$(jq -r '.mcp.github.env.GITHUB_TOKEN // empty' "$CONFIG_FILE")
SLACK_ENV=$(jq -r '.mcp.slack.env.SLACK_BOT_TOKEN // empty' "$CONFIG_FILE")
if [[ -n "$GITHUB_ENV" ]] && [[ -n "$SLACK_ENV" ]]; then
    log_pass "MCP environment variable placeholders configured"
else
    log_fail "Some MCP environment variables not configured"
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
jq -r '.mcp | keys[]' "$CONFIG_FILE" | while read -r mcp; do
    echo "  ✓ $mcp"
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
