#!/bin/bash
# =============================================================================
# All 48 CLI Agents E2E Challenge
# =============================================================================
# This script validates that HelixAgent can:
# 1. List all 48 CLI agents
# 2. Generate valid configurations for all 48 agents
# 3. Validate the generated configurations
# 4. Use the unified generator from LLMsVerifier
#
# Tests: 48 generate + 48 validate + 5 meta = 101 tests
# =============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
BINARY="$PROJECT_ROOT/bin/helixagent"
OUTPUT_DIR="/tmp/helix-agents-challenge-$(date +%s)"
PASSED=0
FAILED=0

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((PASSED++))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((FAILED++))
}

log_info() {
    echo -e "${YELLOW}[INFO]${NC} $1"
}

cleanup() {
    rm -rf "$OUTPUT_DIR" 2>/dev/null || true
}
trap cleanup EXIT

# =============================================================================
# Build Binary
# =============================================================================

log_info "Building HelixAgent binary..."
cd "$PROJECT_ROOT"
if go build -o bin/helixagent ./cmd/helixagent/...; then
    log_pass "Build succeeded"
else
    log_fail "Build failed"
    exit 1
fi

# Create output directory
mkdir -p "$OUTPUT_DIR"

# =============================================================================
# Meta Tests
# =============================================================================

echo ""
echo "=== Meta Tests ==="

# Test 1: List agents command exists
log_info "Testing --list-agents command..."
if "$BINARY" --list-agents > "$OUTPUT_DIR/list-agents.txt" 2>&1; then
    log_pass "--list-agents command works"
else
    log_fail "--list-agents command failed"
fi

# Test 2: List shows 48 agents
if grep -q "48 total" "$OUTPUT_DIR/list-agents.txt"; then
    log_pass "--list-agents shows 48 total"
else
    log_fail "--list-agents does not show 48 total"
fi

# Test 3: Generate all agents command
log_info "Testing --generate-all-agents command..."
if "$BINARY" --generate-all-agents --all-agents-output-dir="$OUTPUT_DIR/all-agents" > "$OUTPUT_DIR/generate-all.txt" 2>&1; then
    log_pass "--generate-all-agents command works"
else
    log_fail "--generate-all-agents command failed"
fi

# Test 4: All 48 succeeded
if grep -q "48 succeeded, 0 failed" "$OUTPUT_DIR/generate-all.txt"; then
    log_pass "All 48 agents generated successfully"
else
    log_fail "Not all agents generated successfully"
fi

# Test 5: Output directory has files
FILE_COUNT=$(ls -1 "$OUTPUT_DIR/all-agents" 2>/dev/null | wc -l)
if [ "$FILE_COUNT" -ge 40 ]; then
    log_pass "Output directory has $FILE_COUNT config files"
else
    log_fail "Output directory only has $FILE_COUNT files (expected 40+)"
fi

# =============================================================================
# Per-Agent Generation Tests (48 tests)
# =============================================================================

echo ""
echo "=== Per-Agent Generation Tests ==="

# All 48 agent names
AGENTS=(
    # Original 18
    "opencode" "crush" "helixcode" "kiro" "aider" "claude-code" "cline"
    "codename-goose" "deepseek-cli" "forge" "gemini-cli" "gpt-engineer"
    "kilocode" "mistral-code" "ollama-code" "plandex" "qwen-code" "amazon-q"
    # New 30
    "agent-deck" "bridle" "cheshire-cat" "claude-plugins" "claude-squad"
    "codai" "codex" "codex-skills" "conduit" "continue" "emdash"
    "fauxpilot" "get-shit-done" "github-copilot-cli" "github-spec-kit"
    "git-mcp" "gptme" "mobile-agent" "multiagent-coding" "nanocoder"
    "noi" "octogen" "openhands" "postgres-mcp" "shai" "snow-cli"
    "task-weaver" "ui-ux-pro-max" "vtcode" "warp"
)

for agent in "${AGENTS[@]}"; do
    CONFIG_FILE="$OUTPUT_DIR/individual/$agent.json"
    mkdir -p "$(dirname "$CONFIG_FILE")"

    if "$BINARY" --generate-agent-config="$agent" --agent-config-output="$CONFIG_FILE" > /dev/null 2>&1; then
        if [ -f "$CONFIG_FILE" ] && [ -s "$CONFIG_FILE" ]; then
            # Check JSON is valid
            if python3 -c "import json; json.load(open('$CONFIG_FILE'))" 2>/dev/null; then
                log_pass "Generate $agent"
            else
                log_fail "Generate $agent (invalid JSON)"
            fi
        else
            log_fail "Generate $agent (empty file)"
        fi
    else
        log_fail "Generate $agent (command failed)"
    fi
done

# =============================================================================
# Per-Agent Validation Tests (48 tests)
# =============================================================================

echo ""
echo "=== Per-Agent Validation Tests ==="

for agent in "${AGENTS[@]}"; do
    CONFIG_FILE="$OUTPUT_DIR/individual/$agent.json"

    if [ -f "$CONFIG_FILE" ]; then
        if "$BINARY" --validate-agent-config="$agent:$CONFIG_FILE" > /dev/null 2>&1; then
            log_pass "Validate $agent"
        else
            log_fail "Validate $agent"
        fi
    else
        log_fail "Validate $agent (no config file)"
    fi
done

# =============================================================================
# Summary
# =============================================================================

echo ""
echo "=============================================="
echo "All 48 CLI Agents E2E Challenge Results"
echo "=============================================="
echo -e "Passed: ${GREEN}$PASSED${NC}"
echo -e "Failed: ${RED}$FAILED${NC}"
echo "Total:  $((PASSED + FAILED))"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}SUCCESS: All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}FAILURE: $FAILED tests failed${NC}"
    exit 1
fi
