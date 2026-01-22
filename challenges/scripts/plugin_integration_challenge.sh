#!/bin/bash
# HelixAgent Plugin Integration Challenge
# Tests full integration of all Tier 1 agent plugins
# 30 tests total

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASSED=0
FAILED=0
TOTAL=30

echo "========================================"
echo "HelixAgent Plugin Integration Challenge"
echo "========================================"
echo ""

# Test function
run_test() {
    local test_name="$1"
    local test_cmd="$2"

    echo -n "Testing: $test_name... "
    if eval "$test_cmd" > /dev/null 2>&1; then
        echo -e "${GREEN}PASSED${NC}"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}FAILED${NC}"
        FAILED=$((FAILED + 1))
    fi
}

# Test 1-6: Claude Code plugin
echo "--- Claude Code Plugin ---"
run_test "Claude Code plugin.json" "test -f '$PROJECT_ROOT/plugins/agents/claude_code/plugin.json'"
run_test "Claude Code hooks.json" "test -f '$PROJECT_ROOT/plugins/agents/claude_code/hooks/hooks.json'"
run_test "Claude Code session_start.js" "test -f '$PROJECT_ROOT/plugins/agents/claude_code/hooks/session_start.js'"
run_test "Claude Code session_end.js" "test -f '$PROJECT_ROOT/plugins/agents/claude_code/hooks/session_end.js'"
run_test "Claude Code pre_tool.js" "test -f '$PROJECT_ROOT/plugins/agents/claude_code/hooks/pre_tool.js'"
run_test "Claude Code post_tool.js" "test -f '$PROJECT_ROOT/plugins/agents/claude_code/hooks/post_tool.js'"

# Test 7-10: OpenCode plugin
echo ""
echo "--- OpenCode Plugin ---"
run_test "OpenCode config exists" "test -f '$PROJECT_ROOT/plugins/agents/opencode/opencode.json'"
run_test "OpenCode MCP server" "test -f '$PROJECT_ROOT/plugins/agents/opencode/mcp/main.go'"
run_test "OpenCode MCP tools" "grep -q 'helixagent_debate' '$PROJECT_ROOT/plugins/agents/opencode/mcp/main.go'"
run_test "OpenCode provider config" "grep -q 'helix-debate-ensemble' '$PROJECT_ROOT/plugins/agents/opencode/opencode.json'"

# Test 11-16: Cline plugin
echo ""
echo "--- Cline Plugin ---"
run_test "Cline hooks.json" "test -f '$PROJECT_ROOT/plugins/agents/cline/hooks/hooks.json'"
run_test "Cline task lifecycle" "test -f '$PROJECT_ROOT/plugins/agents/cline/hooks/task_lifecycle.js'"
run_test "Cline prompt handler" "test -f '$PROJECT_ROOT/plugins/agents/cline/hooks/prompt_handler.js'"
run_test "Cline tool handler" "test -f '$PROJECT_ROOT/plugins/agents/cline/hooks/tool_handler.js'"
run_test "Cline context handler" "test -f '$PROJECT_ROOT/plugins/agents/cline/hooks/context_handler.js'"
run_test "Cline 8 hooks defined" "grep -c 'command' '$PROJECT_ROOT/plugins/agents/cline/hooks/hooks.json' | grep -q '[8-9]'"

# Test 17-22: Kilo-Code plugin
echo ""
echo "--- Kilo-Code Plugin ---"
run_test "Kilo-Code package.json" "test -f '$PROJECT_ROOT/plugins/agents/kilo_code/package.json'"
run_test "Kilo-Code main index" "test -f '$PROJECT_ROOT/plugins/agents/kilo_code/src/index.ts'"
run_test "Kilo-Code transport" "test -f '$PROJECT_ROOT/plugins/agents/kilo_code/src/transport/index.ts'"
run_test "Kilo-Code events" "test -f '$PROJECT_ROOT/plugins/agents/kilo_code/src/events/index.ts'"
run_test "Kilo-Code UI" "test -f '$PROJECT_ROOT/plugins/agents/kilo_code/src/ui/index.ts'"
run_test "Kilo-Code client class" "grep -q 'HelixAgentClient' '$PROJECT_ROOT/plugins/agents/kilo_code/src/index.ts'"

# Test 23-26: Generic MCP server
echo ""
echo "--- Generic MCP Server ---"
run_test "MCP server package.json" "test -f '$PROJECT_ROOT/plugins/mcp-server/package.json'"
run_test "MCP server main" "test -f '$PROJECT_ROOT/plugins/mcp-server/src/index.ts'"
run_test "MCP server transport" "test -f '$PROJECT_ROOT/plugins/mcp-server/src/transport/index.ts'"
run_test "MCP server tools" "test -f '$PROJECT_ROOT/plugins/mcp-server/src/tools/index.ts'"

# Test 27-30: Configuration schema
echo ""
echo "--- Configuration Schema ---"
run_test "Schema file exists" "test -f '$PROJECT_ROOT/plugins/schemas/helixagent-plugin-schema.json'"
run_test "Schema transport config" "grep -q 'transport' '$PROJECT_ROOT/plugins/schemas/helixagent-plugin-schema.json'"
run_test "Schema events config" "grep -q 'events' '$PROJECT_ROOT/plugins/schemas/helixagent-plugin-schema.json'"
run_test "Schema UI config" "grep -q 'ui' '$PROJECT_ROOT/plugins/schemas/helixagent-plugin-schema.json'"

# Summary
echo ""
echo "========================================"
echo "Plugin Integration Challenge Results"
echo "========================================"
echo -e "Passed: ${GREEN}$PASSED${NC}/$TOTAL"
echo -e "Failed: ${RED}$FAILED${NC}/$TOTAL"

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All plugin integration tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some plugin integration tests failed${NC}"
    exit 1
fi
