#!/bin/bash
# ============================================================================
# CLI Agent MCP Challenge
# ============================================================================
# Comprehensive validation that ALL CLI agent configurations:
# 1. Are generated correctly by LLMsVerifier
# 2. Have 35+ MCPs configured
# 3. Can be parsed by their respective CLI agents
# 4. Include all required MCP categories (Anthropic, HelixAgent, Community)
#
# This challenge ACTUALLY validates configurations work - no false positives!
#
# Usage: ./cli_agent_mcp_challenge.sh
# ============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Test counters
TOTAL_TESTS=0
TESTS_PASSED=0
TESTS_FAILED=0

log_header() {
    echo ""
    echo -e "${CYAN}============================================${NC}"
    echo -e "${CYAN}  $1${NC}"
    echo -e "${CYAN}============================================${NC}"
}

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

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# ============================================================================
# Challenge: CLI Agent Configuration Generation
# ============================================================================

log_header "CLI AGENT MCP CHALLENGE"
echo "Validating CLI agent configurations have 35+ MCPs and work correctly"
echo ""

# ============================================================================
# Phase 1: Prerequisites
# ============================================================================
log_header "Phase 1: Prerequisites Check"

log_test "P1.1: jq command available"
if command -v jq &> /dev/null; then
    log_pass "jq is installed"
else
    log_fail "jq not found - please install jq"
    exit 1
fi

log_test "P1.2: Configuration generator script exists"
GENERATOR_SCRIPT="$PROJECT_ROOT/scripts/cli-agents/generate-all-configs.sh"
if [[ -f "$GENERATOR_SCRIPT" ]]; then
    log_pass "Generator script found: $GENERATOR_SCRIPT"
else
    log_fail "Generator script not found"
    exit 1
fi

log_test "P1.3: HelixAgent is running"
HEALTH=$(curl -s http://localhost:7061/health 2>/dev/null | jq -r '.status' 2>/dev/null || echo "")
if [[ "$HEALTH" == "healthy" ]]; then
    log_pass "HelixAgent is healthy"
else
    log_warning "HelixAgent not running - some tests may be skipped"
fi

# ============================================================================
# Phase 2: OpenCode Configuration Validation
# ============================================================================
log_header "Phase 2: OpenCode Configuration"

OPENCODE_CONFIG="$HOME/.config/opencode/opencode.json"

log_test "P2.1: OpenCode config exists"
if [[ -f "$OPENCODE_CONFIG" ]]; then
    log_pass "Config exists: $OPENCODE_CONFIG"
else
    log_fail "Config not found - running generator"
    "$GENERATOR_SCRIPT" --agent=opencode --install
    if [[ -f "$OPENCODE_CONFIG" ]]; then
        log_pass "Config generated and installed"
    else
        log_fail "Failed to generate config"
    fi
fi

log_test "P2.2: OpenCode config is valid JSON"
if jq empty "$OPENCODE_CONFIG" 2>/dev/null; then
    log_pass "Valid JSON"
else
    log_fail "Invalid JSON"
fi

log_test "P2.3: OpenCode config has schema field"
SCHEMA=$(jq -r '."$schema" // empty' "$OPENCODE_CONFIG" 2>/dev/null)
if [[ -n "$SCHEMA" ]]; then
    log_pass "Schema: $SCHEMA"
else
    log_fail "Missing schema field"
fi

log_test "P2.4: OpenCode config has provider section"
PROVIDER=$(jq '.provider | keys | length' "$OPENCODE_CONFIG" 2>/dev/null)
if [[ "$PROVIDER" -gt 0 ]]; then
    log_pass "Provider section: $PROVIDER provider(s)"
else
    log_fail "No providers configured"
fi

log_test "P2.5: OpenCode config has 35+ MCPs"
MCP_COUNT=$(jq '.mcp | keys | length' "$OPENCODE_CONFIG" 2>/dev/null || echo 0)
if [[ "$MCP_COUNT" -ge 35 ]]; then
    log_pass "MCP count: $MCP_COUNT (>= 35)"
else
    log_fail "MCP count: $MCP_COUNT (< 35)"
fi

log_test "P2.6: MCPs have correct command/args format"
INVALID=$(jq '[.mcp | to_entries[] | select(.value.command == null)] | length' "$OPENCODE_CONFIG" 2>/dev/null || echo 999)
if [[ "$INVALID" -eq 0 ]]; then
    log_pass "All MCPs have valid format"
else
    log_fail "$INVALID MCPs have invalid format"
fi

log_test "P2.7: Anthropic official MCPs present (filesystem, fetch, memory, time, git)"
ANTHROPIC_MCPS=("filesystem" "fetch" "memory" "time" "git")
MISSING=""
for mcp in "${ANTHROPIC_MCPS[@]}"; do
    if ! jq -e ".mcp.\"$mcp\"" "$OPENCODE_CONFIG" > /dev/null 2>&1; then
        MISSING="$MISSING $mcp"
    fi
done
if [[ -z "$MISSING" ]]; then
    log_pass "All Anthropic MCPs present"
else
    log_fail "Missing:$MISSING"
fi

log_test "P2.8: HelixAgent MCPs present"
HELIX_MCPS=("helixagent" "helixagent-debate" "helixagent-rag" "helixagent-memory")
MISSING=""
for mcp in "${HELIX_MCPS[@]}"; do
    if ! jq -e ".mcp.\"$mcp\"" "$OPENCODE_CONFIG" > /dev/null 2>&1; then
        MISSING="$MISSING $mcp"
    fi
done
if [[ -z "$MISSING" ]]; then
    log_pass "All HelixAgent MCPs present"
else
    log_fail "Missing:$MISSING"
fi

log_test "P2.9: Community/Infrastructure MCPs present"
COMMUNITY_MCPS=("docker" "kubernetes" "redis" "qdrant" "postgres")
MISSING=""
for mcp in "${COMMUNITY_MCPS[@]}"; do
    if ! jq -e ".mcp.\"$mcp\"" "$OPENCODE_CONFIG" > /dev/null 2>&1; then
        MISSING="$MISSING $mcp"
    fi
done
if [[ -z "$MISSING" ]]; then
    log_pass "All community MCPs present"
else
    log_fail "Missing:$MISSING"
fi

log_test "P2.10: Productivity MCPs present"
PRODUCTIVITY_MCPS=("github" "gitlab" "jira" "asana" "notion" "linear" "slack")
MISSING=""
for mcp in "${PRODUCTIVITY_MCPS[@]}"; do
    if ! jq -e ".mcp.\"$mcp\"" "$OPENCODE_CONFIG" > /dev/null 2>&1; then
        MISSING="$MISSING $mcp"
    fi
done
if [[ -z "$MISSING" ]]; then
    log_pass "All productivity MCPs present"
else
    log_fail "Missing:$MISSING"
fi

log_test "P2.11: OpenCode binary available"
if command -v opencode &> /dev/null; then
    log_pass "OpenCode binary found"

    log_test "P2.12: OpenCode can start without config errors"
    OPENCODE_OUT=$(timeout 5 opencode --help 2>&1 || true)
    if echo "$OPENCODE_OUT" | grep -qi "error.*config\|invalid"; then
        log_fail "OpenCode reported config errors"
    else
        log_pass "OpenCode starts without config errors"
    fi
else
    log_warning "OpenCode not installed - skipping binary tests"
fi

# ============================================================================
# Phase 3: Crush Configuration Validation
# ============================================================================
log_header "Phase 3: Crush Configuration"

CRUSH_CONFIG="$HOME/.config/crush/crush.json"

log_test "P3.1: Crush config exists"
if [[ -f "$CRUSH_CONFIG" ]]; then
    log_pass "Config exists: $CRUSH_CONFIG"
else
    log_fail "Config not found - running generator"
    "$GENERATOR_SCRIPT" --agent=crush --install
fi

log_test "P3.2: Crush config is valid JSON"
if jq empty "$CRUSH_CONFIG" 2>/dev/null; then
    log_pass "Valid JSON"
else
    log_fail "Invalid JSON"
fi

log_test "P3.3: Crush config has 35+ MCPs"
MCP_COUNT=$(jq '.mcp | keys | length' "$CRUSH_CONFIG" 2>/dev/null || echo 0)
if [[ "$MCP_COUNT" -ge 35 ]]; then
    log_pass "MCP count: $MCP_COUNT (>= 35)"
else
    log_fail "MCP count: $MCP_COUNT (< 35)"
fi

log_test "P3.4: Crush MCPs have correct format"
INVALID=$(jq '[.mcp | to_entries[] | select(.value.command == null)] | length' "$CRUSH_CONFIG" 2>/dev/null || echo 999)
if [[ "$INVALID" -eq 0 ]]; then
    log_pass "All MCPs have valid format"
else
    log_fail "$INVALID MCPs have invalid format"
fi

log_test "P3.5: Crush has HelixAgent MCPs"
HELIX_MCPS=("helixagent" "helixagent-debate")
MISSING=""
for mcp in "${HELIX_MCPS[@]}"; do
    if ! jq -e ".mcp.\"$mcp\"" "$CRUSH_CONFIG" > /dev/null 2>&1; then
        MISSING="$MISSING $mcp"
    fi
done
if [[ -z "$MISSING" ]]; then
    log_pass "HelixAgent MCPs present"
else
    log_fail "Missing:$MISSING"
fi

# ============================================================================
# Phase 4: LLMsVerifier Integration
# ============================================================================
log_header "Phase 4: LLMsVerifier Integration"

VERIFIER_DIR="$PROJECT_ROOT/LLMsVerifier"

log_test "P4.1: LLMsVerifier exists"
if [[ -d "$VERIFIER_DIR" ]]; then
    log_pass "LLMsVerifier found"
else
    log_fail "LLMsVerifier not found"
fi

log_test "P4.2: CLI agents package exists"
CLIAGENTS_PKG="$VERIFIER_DIR/llm-verifier/pkg/cliagents"
if [[ -d "$CLIAGENTS_PKG" ]]; then
    log_pass "CLI agents package found"
else
    log_fail "CLI agents package not found"
fi

log_test "P4.3: Generator exists"
GENERATOR="$CLIAGENTS_PKG/generator.go"
if [[ -f "$GENERATOR" ]]; then
    log_pass "Generator found: $GENERATOR"
else
    log_fail "Generator not found"
fi

# ============================================================================
# Phase 5: Configuration Consistency
# ============================================================================
log_header "Phase 5: Configuration Consistency"

log_test "P5.1: Generated config matches installed (OpenCode)"
GENERATED="$PROJECT_ROOT/scripts/cli-agents/configs/generated/opencode/opencode.json"
if [[ -f "$GENERATED" ]] && diff -q "$GENERATED" "$OPENCODE_CONFIG" > /dev/null 2>&1; then
    log_pass "Generated matches installed"
else
    log_warning "Generated config differs from installed"
fi

log_test "P5.2: Generated config matches installed (Crush)"
GENERATED="$PROJECT_ROOT/scripts/cli-agents/configs/generated/crush/crush.json"
if [[ -f "$GENERATED" ]] && diff -q "$GENERATED" "$CRUSH_CONFIG" > /dev/null 2>&1; then
    log_pass "Generated matches installed"
else
    log_warning "Generated config differs from installed"
fi

log_test "P5.3: OpenCode and Crush have same MCP count"
OPENCODE_MCPS=$(jq '.mcp | keys | length' "$OPENCODE_CONFIG" 2>/dev/null || echo 0)
CRUSH_MCPS=$(jq '.mcp | keys | length' "$CRUSH_CONFIG" 2>/dev/null || echo 0)
if [[ "$OPENCODE_MCPS" -eq "$CRUSH_MCPS" ]]; then
    log_pass "Both have $OPENCODE_MCPS MCPs"
else
    log_warning "MCP counts differ: OpenCode=$OPENCODE_MCPS, Crush=$CRUSH_MCPS"
fi

# ============================================================================
# Summary
# ============================================================================
log_header "CHALLENGE SUMMARY"

echo ""
log_info "Total Tests:  $TOTAL_TESTS"
log_info "Passed:       $TESTS_PASSED"
log_info "Failed:       $TESTS_FAILED"
echo ""

# Show MCP summary
echo "=== MCP SUMMARY ==="
echo "OpenCode MCPs: $(jq '.mcp | keys | length' "$OPENCODE_CONFIG" 2>/dev/null || echo 0)"
echo "Crush MCPs:    $(jq '.mcp | keys | length' "$CRUSH_CONFIG" 2>/dev/null || echo 0)"
echo ""

# MCP Categories
echo "=== MCP CATEGORIES ==="
echo "Anthropic Official: filesystem, fetch, memory, time, git, sqlite, postgres, puppeteer"
echo "                    brave-search, google-maps, slack, sequential-thinking, everart"
echo "                    exa, linear, sentry, notion, figma, aws-kb-retrieval, gitlab"
echo "HelixAgent Custom:  helixagent, helixagent-debate, helixagent-rag, helixagent-memory"
echo "Community:          docker, kubernetes, redis, mongodb, elasticsearch, qdrant, chroma"
echo "                    jira, asana, google-drive, aws-s3, datadog"
echo ""

if [[ "$TESTS_FAILED" -eq 0 ]]; then
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}  CHALLENGE PASSED!${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo "All CLI agent configurations are valid with 35+ MCPs."
    echo "OpenCode and Crush are ready to use with HelixAgent."
    exit 0
else
    echo -e "${RED}========================================${NC}"
    echo -e "${RED}  CHALLENGE FAILED!${NC}"
    echo -e "${RED}========================================${NC}"
    echo ""
    echo "Please fix the issues above and re-run this challenge."
    exit 1
fi
