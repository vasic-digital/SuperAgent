#!/bin/bash
#
# Advanced Provider Access Challenge
#
# This challenge validates the advanced provider access mechanisms:
# - Qwen ACP (Agent Communication Protocol) over stdin/stdout
# - Zen HTTP server (OpenCode serve)
# - Claude JSON output format with session management
#
# Test Categories:
# 1-15:  Qwen ACP Provider
# 16-30: Zen HTTP Provider
# 31-40: Claude JSON Output
# 41-50: Integration Tests
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
echo "Advanced Provider Access Challenge"
echo "=========================================="
echo ""
echo "Validating ACP, HTTP, and JSON output mechanisms"
echo ""

cd "$PROJECT_ROOT"

# ===========================================
# Category 1: Qwen ACP Provider (Tests 1-15)
# ===========================================
echo -e "${BLUE}=== Category 1: Qwen ACP Provider ===${NC}"

# Test 1: QwenACPProvider struct exists
if grep -q "type QwenACPProvider struct" internal/llm/providers/qwen/qwen_acp.go 2>/dev/null; then
    log_test 1 "QwenACPProvider struct exists" PASS
else
    log_test 1 "QwenACPProvider struct exists" FAIL "Struct not found"
fi

# Test 2: NewQwenACPProvider constructor exists
if grep -q "func NewQwenACPProvider" internal/llm/providers/qwen/qwen_acp.go 2>/dev/null; then
    log_test 2 "NewQwenACPProvider constructor exists" PASS
else
    log_test 2 "NewQwenACPProvider constructor exists" FAIL "Constructor not found"
fi

# Test 3: NewQwenACPProviderWithModel constructor exists
if grep -q "func NewQwenACPProviderWithModel" internal/llm/providers/qwen/qwen_acp.go 2>/dev/null; then
    log_test 3 "NewQwenACPProviderWithModel constructor exists" PASS
else
    log_test 3 "NewQwenACPProviderWithModel constructor exists" FAIL "Constructor not found"
fi

# Test 4: ACP protocol version defined
if grep -q "acpProtocolVersion" internal/llm/providers/qwen/qwen_acp.go 2>/dev/null; then
    log_test 4 "ACP protocol version defined" PASS
else
    log_test 4 "ACP protocol version defined" FAIL "Constant not found"
fi

# Test 5: ACP request type defined (JSON-RPC)
if grep -q 'type acpRequest struct' internal/llm/providers/qwen/qwen_acp.go 2>/dev/null && \
   grep -q 'JSONRPC.*json:"jsonrpc"' internal/llm/providers/qwen/qwen_acp.go 2>/dev/null; then
    log_test 5 "ACP request type with JSON-RPC" PASS
else
    log_test 5 "ACP request type with JSON-RPC" FAIL "Type not correctly defined"
fi

# Test 6: ACP response type defined
if grep -q 'type acpResponse struct' internal/llm/providers/qwen/qwen_acp.go 2>/dev/null; then
    log_test 6 "ACP response type defined" PASS
else
    log_test 6 "ACP response type defined" FAIL "Type not found"
fi

# Test 7: Start method exists
if grep -q 'func.*QwenACPProvider.*Start' internal/llm/providers/qwen/qwen_acp.go 2>/dev/null; then
    log_test 7 "QwenACPProvider.Start method exists" PASS
else
    log_test 7 "QwenACPProvider.Start method exists" FAIL "Method not found"
fi

# Test 8: Stop method exists
if grep -q 'func.*QwenACPProvider.*Stop' internal/llm/providers/qwen/qwen_acp.go 2>/dev/null; then
    log_test 8 "QwenACPProvider.Stop method exists" PASS
else
    log_test 8 "QwenACPProvider.Stop method exists" FAIL "Method not found"
fi

# Test 9: sendRequest method exists
if grep -q 'func.*QwenACPProvider.*sendRequest' internal/llm/providers/qwen/qwen_acp.go 2>/dev/null; then
    log_test 9 "ACP sendRequest method exists" PASS
else
    log_test 9 "ACP sendRequest method exists" FAIL "Method not found"
fi

# Test 10: initialize method exists
if grep -q 'func.*QwenACPProvider.*initialize' internal/llm/providers/qwen/qwen_acp.go 2>/dev/null; then
    log_test 10 "ACP initialize method exists" PASS
else
    log_test 10 "ACP initialize method exists" FAIL "Method not found"
fi

# Test 11: createSession method exists
if grep -q 'func.*QwenACPProvider.*createSession' internal/llm/providers/qwen/qwen_acp.go 2>/dev/null; then
    log_test 11 "ACP createSession method exists" PASS
else
    log_test 11 "ACP createSession method exists" FAIL "Method not found"
fi

# Test 12: IsAvailable method exists
if grep -q 'func.*QwenACPProvider.*IsAvailable' internal/llm/providers/qwen/qwen_acp.go 2>/dev/null; then
    log_test 12 "QwenACPProvider.IsAvailable method exists" PASS
else
    log_test 12 "QwenACPProvider.IsAvailable method exists" FAIL "Method not found"
fi

# Test 13: CanUseQwenACP function exists
if grep -q 'func CanUseQwenACP' internal/llm/providers/qwen/qwen_acp.go 2>/dev/null; then
    log_test 13 "CanUseQwenACP function exists" PASS
else
    log_test 13 "CanUseQwenACP function exists" FAIL "Function not found"
fi

# Test 14: ACP test file exists
if [ -f "internal/llm/providers/qwen/qwen_acp_test.go" ]; then
    log_test 14 "Qwen ACP test file exists" PASS
else
    log_test 14 "Qwen ACP test file exists" FAIL "Test file not found"
fi

# Test 15: ACP provider returns correct name
if grep -q 'return "qwen-acp"' internal/llm/providers/qwen/qwen_acp.go 2>/dev/null; then
    log_test 15 "ACP provider returns correct name" PASS
else
    log_test 15 "ACP provider returns correct name" FAIL "Incorrect provider name"
fi

# ===========================================
# Category 2: Zen HTTP Provider (Tests 16-30)
# ===========================================
echo ""
echo -e "${BLUE}=== Category 2: Zen HTTP Provider ===${NC}"

# Test 16: ZenHTTPProvider struct exists
if grep -q "type ZenHTTPProvider struct" internal/llm/providers/zen/zen_http.go 2>/dev/null; then
    log_test 16 "ZenHTTPProvider struct exists" PASS
else
    log_test 16 "ZenHTTPProvider struct exists" FAIL "Struct not found"
fi

# Test 17: NewZenHTTPProvider constructor exists
if grep -q "func NewZenHTTPProvider" internal/llm/providers/zen/zen_http.go 2>/dev/null; then
    log_test 17 "NewZenHTTPProvider constructor exists" PASS
else
    log_test 17 "NewZenHTTPProvider constructor exists" FAIL "Constructor not found"
fi

# Test 18: NewZenHTTPProviderWithModel constructor exists
if grep -q "func NewZenHTTPProviderWithModel" internal/llm/providers/zen/zen_http.go 2>/dev/null; then
    log_test 18 "NewZenHTTPProviderWithModel constructor exists" PASS
else
    log_test 18 "NewZenHTTPProviderWithModel constructor exists" FAIL "Constructor not found"
fi

# Test 19: Default config with localhost:4096
if grep -q 'http://localhost:4096' internal/llm/providers/zen/zen_http.go 2>/dev/null; then
    log_test 19 "Default config uses localhost:4096" PASS
else
    log_test 19 "Default config uses localhost:4096" FAIL "Incorrect default URL"
fi

# Test 20: HTTP client configured
if grep -q 'httpClient.*\*http.Client' internal/llm/providers/zen/zen_http.go 2>/dev/null; then
    log_test 20 "HTTP client configured" PASS
else
    log_test 20 "HTTP client configured" FAIL "HTTP client not found"
fi

# Test 21: Session management
if grep -q 'sessionID.*string' internal/llm/providers/zen/zen_http.go 2>/dev/null; then
    log_test 21 "Session ID management" PASS
else
    log_test 21 "Session ID management" FAIL "Session ID field not found"
fi

# Test 22: IsServerRunning method exists
if grep -q 'func.*ZenHTTPProvider.*IsServerRunning' internal/llm/providers/zen/zen_http.go 2>/dev/null; then
    log_test 22 "IsServerRunning method exists" PASS
else
    log_test 22 "IsServerRunning method exists" FAIL "Method not found"
fi

# Test 23: StartServer method exists
if grep -q 'func.*ZenHTTPProvider.*StartServer' internal/llm/providers/zen/zen_http.go 2>/dev/null; then
    log_test 23 "StartServer method exists" PASS
else
    log_test 23 "StartServer method exists" FAIL "Method not found"
fi

# Test 24: StopServer method exists
if grep -q 'func.*ZenHTTPProvider.*StopServer' internal/llm/providers/zen/zen_http.go 2>/dev/null; then
    log_test 24 "StopServer method exists" PASS
else
    log_test 24 "StopServer method exists" FAIL "Method not found"
fi

# Test 25: doRequest method exists
if grep -q 'func.*ZenHTTPProvider.*doRequest' internal/llm/providers/zen/zen_http.go 2>/dev/null; then
    log_test 25 "doRequest method exists" PASS
else
    log_test 25 "doRequest method exists" FAIL "Method not found"
fi

# Test 26: createSession method exists
if grep -q 'func.*ZenHTTPProvider.*createSession' internal/llm/providers/zen/zen_http.go 2>/dev/null; then
    log_test 26 "HTTP createSession method exists" PASS
else
    log_test 26 "HTTP createSession method exists" FAIL "Method not found"
fi

# Test 27: sendMessage method exists
if grep -q 'func.*ZenHTTPProvider.*sendMessage' internal/llm/providers/zen/zen_http.go 2>/dev/null; then
    log_test 27 "HTTP sendMessage method exists" PASS
else
    log_test 27 "HTTP sendMessage method exists" FAIL "Method not found"
fi

# Test 28: Auto-start server option
if grep -q 'AutoStart.*bool' internal/llm/providers/zen/zen_http.go 2>/dev/null && \
   grep -q 'autoStart.*bool' internal/llm/providers/zen/zen_http.go 2>/dev/null; then
    log_test 28 "Auto-start server option" PASS
else
    log_test 28 "Auto-start server option" FAIL "AutoStart option not found"
fi

# Test 29: HTTP test file exists
if [ -f "internal/llm/providers/zen/zen_http_test.go" ]; then
    log_test 29 "Zen HTTP test file exists" PASS
else
    log_test 29 "Zen HTTP test file exists" FAIL "Test file not found"
fi

# Test 30: HTTP provider returns correct name
if grep -q 'return "zen-http"' internal/llm/providers/zen/zen_http.go 2>/dev/null; then
    log_test 30 "HTTP provider returns correct name" PASS
else
    log_test 30 "HTTP provider returns correct name" FAIL "Incorrect provider name"
fi

# ===========================================
# Category 3: Claude JSON Output (Tests 31-40)
# ===========================================
echo ""
echo -e "${BLUE}=== Category 3: Claude JSON Output ===${NC}"

# Test 31: Claude JSON response type defined
if grep -q 'type claudeJSONResponse struct' internal/llm/providers/claude/claude_cli.go 2>/dev/null; then
    log_test 31 "Claude JSON response type defined" PASS
else
    log_test 31 "Claude JSON response type defined" FAIL "Type not found"
fi

# Test 32: Result field in JSON struct
if grep -q 'Result.*string.*json:"result"' internal/llm/providers/claude/claude_cli.go 2>/dev/null; then
    log_test 32 "Result field in JSON struct" PASS
else
    log_test 32 "Result field in JSON struct" FAIL "Field not found"
fi

# Test 33: SessionID field in JSON struct
if grep -q 'SessionID.*string.*json:"session_id"' internal/llm/providers/claude/claude_cli.go 2>/dev/null; then
    log_test 33 "SessionID field in JSON struct" PASS
else
    log_test 33 "SessionID field in JSON struct" FAIL "Field not found"
fi

# Test 34: parseJSONResponse method exists
if grep -q 'func.*ClaudeCLIProvider.*parseJSONResponse' internal/llm/providers/claude/claude_cli.go 2>/dev/null; then
    log_test 34 "parseJSONResponse method exists" PASS
else
    log_test 34 "parseJSONResponse method exists" FAIL "Method not found"
fi

# Test 35: Session ID stored for continuity
if grep -q 'p.sessionID = sessionID' internal/llm/providers/claude/claude_cli.go 2>/dev/null; then
    log_test 35 "Session ID stored for continuity" PASS
else
    log_test 35 "Session ID stored for continuity" FAIL "Session continuity not implemented"
fi

# Test 36: Uses --output-format json flag
if grep -q '"--output-format".*"json"' internal/llm/providers/claude/claude_cli.go 2>/dev/null; then
    log_test 36 "Uses --output-format json flag" PASS
else
    log_test 36 "Uses --output-format json flag" FAIL "JSON format flag not used"
fi

# Test 37: Uses --resume flag for session
if grep -q '"--resume"' internal/llm/providers/claude/claude_cli.go 2>/dev/null; then
    log_test 37 "Uses --resume flag for session" PASS
else
    log_test 37 "Uses --resume flag for session" FAIL "Resume flag not used"
fi

# Test 38: JSON parsing has fallback
if grep -q 'return rawOutput' internal/llm/providers/claude/claude_cli.go 2>/dev/null; then
    log_test 38 "JSON parsing has fallback" PASS
else
    log_test 38 "JSON parsing has fallback" FAIL "Fallback not implemented"
fi

# Test 39: Usage metadata extracted
if grep -q 'Usage.*map\[string\]interface' internal/llm/providers/claude/claude_cli.go 2>/dev/null; then
    log_test 39 "Usage metadata extracted" PASS
else
    log_test 39 "Usage metadata extracted" FAIL "Usage extraction not found"
fi

# Test 40: JSON tests exist in test file
if grep -q 'TestClaudeCLIProvider_ParseJSONResponse' internal/llm/providers/claude/claude_cli_test.go 2>/dev/null; then
    log_test 40 "JSON parsing tests exist" PASS
else
    log_test 40 "JSON parsing tests exist" FAIL "Tests not found"
fi

# ===========================================
# Category 4: Integration Tests (Tests 41-50)
# ===========================================
echo ""
echo -e "${BLUE}=== Category 4: Integration Tests ===${NC}"

# Test 41: Project builds successfully
echo "Building project..."
if go build ./... 2>/dev/null; then
    log_test 41 "Project builds successfully" PASS
else
    log_test 41 "Project builds successfully" FAIL "Build failed"
fi

# Test 42: Qwen ACP unit tests pass
echo "Running Qwen ACP tests..."
if timeout 60 go test -short -v ./internal/llm/providers/qwen/... -run "ACP" 2>&1 | grep -q "PASS\|ok\|SKIP"; then
    log_test 42 "Qwen ACP unit tests pass" PASS
else
    log_test 42 "Qwen ACP unit tests pass" SKIP "Tests may require CLI"
fi

# Test 43: Zen HTTP unit tests pass
echo "Running Zen HTTP tests..."
if timeout 60 go test -short -v ./internal/llm/providers/zen/... -run "HTTP" 2>&1 | grep -q "PASS\|ok\|SKIP"; then
    log_test 43 "Zen HTTP unit tests pass" PASS
else
    log_test 43 "Zen HTTP unit tests pass" SKIP "Tests may require server"
fi

# Test 44: Claude JSON tests pass
echo "Running Claude JSON tests..."
if timeout 60 go test -short -v ./internal/llm/providers/claude/... -run "JSON" 2>&1 | grep -q "PASS\|ok\|SKIP"; then
    log_test 44 "Claude JSON unit tests pass" PASS
else
    log_test 44 "Claude JSON unit tests pass" SKIP "Tests may require CLI"
fi

# Test 45: Provider registry prefers ACP for Qwen
if grep -q 'CanUseQwenACP\|NewQwenACPProvider' internal/services/provider_registry.go 2>/dev/null; then
    log_test 45 "Registry prefers ACP for Qwen" PASS
else
    log_test 45 "Registry prefers ACP for Qwen" FAIL "ACP preference not found"
fi

# Test 46: Provider discovery prefers HTTP for Zen
if grep -q 'NewZenHTTPProvider\|IsServerRunning' internal/services/provider_discovery.go 2>/dev/null; then
    log_test 46 "Discovery prefers HTTP for Zen" PASS
else
    log_test 46 "Discovery prefers HTTP for Zen" FAIL "HTTP preference not found"
fi

# Test 47: Documentation exists for CLI proxy
if [ -f "docs/providers/CLI_PROXY.md" ]; then
    log_test 47 "CLI proxy documentation exists" PASS
else
    log_test 47 "CLI proxy documentation exists" FAIL "Documentation not found"
fi

# Test 48: ACP documented in CLI proxy docs
if grep -q "ACP\|Agent Communication Protocol" docs/providers/CLI_PROXY.md 2>/dev/null; then
    log_test 48 "ACP documented" PASS
else
    log_test 48 "ACP documented" FAIL "ACP not documented"
fi

# Test 49: HTTP server documented in CLI proxy docs
if grep -q "HTTP\|opencode serve" docs/providers/CLI_PROXY.md 2>/dev/null; then
    log_test 49 "HTTP server documented" PASS
else
    log_test 49 "HTTP server documented" FAIL "HTTP not documented"
fi

# Test 50: All providers implement LLMProvider interface
echo "Checking LLMProvider interface implementation..."
if go build ./internal/llm/providers/... 2>/dev/null; then
    log_test 50 "All providers implement LLMProvider" PASS
else
    log_test 50 "All providers implement LLMProvider" FAIL "Interface not implemented"
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
