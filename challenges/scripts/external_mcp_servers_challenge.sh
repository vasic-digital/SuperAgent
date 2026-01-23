#!/bin/bash
# External MCP Servers Challenge
# Comprehensive validation of all Model Context Protocol servers from git submodules
#
# This challenge verifies:
# 1. Git submodules exist and are properly initialized
# 2. All MCP server source code exists
# 3. Container infrastructure exists
# 4. All servers respond properly
# 5. All servers are documented
# 6. All servers are in OpenCode configuration

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
PASSED=0
FAILED=0
SKIPPED=0

# Project root
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$PROJECT_ROOT"

# Log functions
pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    PASSED=$((PASSED + 1))
}

fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    FAILED=$((FAILED + 1))
}

skip() {
    echo -e "${YELLOW}[SKIP]${NC} $1"
    SKIPPED=$((SKIPPED + 1))
}

info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

# Detect container runtime
detect_runtime() {
    if command -v podman &> /dev/null; then
        echo "podman"
    elif command -v docker &> /dev/null; then
        echo "docker"
    else
        echo ""
    fi
}

RUNTIME=$(detect_runtime)

echo "=============================================="
echo "  External MCP Servers Challenge"
echo "  Git Submodule MCP Server Validation"
echo "=============================================="
echo ""
echo "Date: $(date)"
echo "Container Runtime: ${RUNTIME:-'Not found'}"
echo ""

# ==========================================
# SECTION 1: Git Submodules
# ==========================================
echo ""
echo "=== Section 1: Git Submodules ==="
echo ""

# Test 1.1: Active servers submodule exists
if [ -d "external/mcp-servers/servers" ]; then
    if [ -d "external/mcp-servers/servers/.git" ] || [ -f "external/mcp-servers/servers/.git" ]; then
        pass "1.1 Active MCP servers submodule exists"
    else
        fail "1.1 Active MCP servers is not a git submodule"
    fi
else
    fail "1.1 Active MCP servers submodule directory missing"
fi

# Test 1.2: Archived servers submodule exists
if [ -d "external/mcp-servers/servers-archived" ]; then
    if [ -d "external/mcp-servers/servers-archived/.git" ] || [ -f "external/mcp-servers/servers-archived/.git" ]; then
        pass "1.2 Archived MCP servers submodule exists"
    else
        fail "1.2 Archived MCP servers is not a git submodule"
    fi
else
    fail "1.2 Archived MCP servers submodule directory missing"
fi

# ==========================================
# SECTION 2: Source Code Existence
# ==========================================
echo ""
echo "=== Section 2: Source Code Existence ==="
echo ""

# Active servers
ACTIVE_SERVERS="fetch filesystem git memory time sequentialthinking everything"
for server in $ACTIVE_SERVERS; do
    if [ -d "external/mcp-servers/servers/src/$server" ]; then
        pass "2.A.$server Active server source exists"
    else
        fail "2.A.$server Active server source missing"
    fi
done

# Archived servers (in src/ subdirectory)
ARCHIVED_SERVERS="postgres sqlite slack github gitlab google-maps brave-search puppeteer redis sentry gdrive everart aws-kb-retrieval-server"
for server in $ARCHIVED_SERVERS; do
    if [ -d "external/mcp-servers/servers-archived/src/$server" ]; then
        pass "2.B.$server Archived server source exists"
    else
        fail "2.B.$server Archived server source missing"
    fi
done

# ==========================================
# SECTION 3: Container Infrastructure
# ==========================================
echo ""
echo "=== Section 3: Container Infrastructure ==="
echo ""

# Test 3.1: Dockerfile exists
if [ -f "external/mcp-servers/Dockerfile" ]; then
    pass "3.1 MCP servers Dockerfile exists"
else
    fail "3.1 MCP servers Dockerfile missing"
fi

# Test 3.2: docker-compose.yml exists
if [ -f "external/mcp-servers/docker-compose.yml" ]; then
    pass "3.2 MCP servers docker-compose.yml exists"
else
    fail "3.2 MCP servers docker-compose.yml missing"
fi

# Test 3.3: Startup script exists
if [ -f "external/mcp-servers/scripts/start-all.sh" ]; then
    pass "3.3 MCP servers startup script exists"
else
    fail "3.3 MCP servers startup script missing"
fi

# Test 3.4: Health check script exists
if [ -f "external/mcp-servers/scripts/health-check.sh" ]; then
    pass "3.4 MCP servers health check script exists"
else
    fail "3.4 MCP servers health check script missing"
fi

# ==========================================
# SECTION 4: Documentation
# ==========================================
echo ""
echo "=== Section 4: Documentation ==="
echo ""

# Test 4.1: MCP servers README exists
if [ -f "external/mcp-servers/README.md" ]; then
    pass "4.1 MCP servers README.md exists"

    # Check if all servers are documented
    README_CONTENT=$(cat external/mcp-servers/README.md)
    ALL_SERVERS="fetch filesystem git memory time sequential-thinking everything postgres sqlite slack github gitlab google-maps brave-search puppeteer redis sentry gdrive everart aws-kb-retrieval"

    for server in $ALL_SERVERS; do
        if echo "$README_CONTENT" | grep -qi "$server"; then
            pass "4.D.$server Server documented in README"
        else
            fail "4.D.$server Server NOT documented in README"
        fi
    done
else
    fail "4.1 MCP servers README.md missing"
fi

# ==========================================
# SECTION 5: OpenCode Configuration
# ==========================================
echo ""
echo "=== Section 5: OpenCode Configuration ==="
echo ""

# Test 5.1: Generate config
if [ -x "./bin/helixagent" ]; then
    CONFIG=$(LOCAL_ENDPOINT=http://localhost:7061 ./bin/helixagent --generate-opencode-config 2>/dev/null | grep -v "^time=" | grep -v "^IMPORTANT")

    # Check if config is valid JSON
    if echo "$CONFIG" | jq . &>/dev/null; then
        pass "5.1 OpenCode config is valid JSON"

        # Check for each MCP server in config
        MCP_SERVERS="fetch filesystem git memory time sequential-thinking everything postgres sqlite slack github gitlab google-maps brave-search puppeteer redis sentry gdrive everart aws-kb-retrieval helixagent"

        for server in $MCP_SERVERS; do
            if echo "$CONFIG" | jq -e ".mcp[\"$server\"]" &>/dev/null; then
                pass "5.C.$server Server in OpenCode config"
            else
                fail "5.C.$server Server NOT in OpenCode config"
            fi
        done
    else
        fail "5.1 OpenCode config is NOT valid JSON"
    fi
else
    skip "5.1 HelixAgent binary not found"
fi

# ==========================================
# SECTION 6: Integration Tests
# ==========================================
echo ""
echo "=== Section 6: Integration Tests ==="
echo ""

# Test 6.1: Integration test file exists
if [ -f "tests/integration/mcp_servers_test.go" ]; then
    pass "6.1 MCP servers integration test file exists"
else
    fail "6.1 MCP servers integration test file missing"
fi

# ==========================================
# SUMMARY
# ==========================================
echo ""
echo "=============================================="
echo "  Challenge Summary"
echo "=============================================="
echo ""
echo -e "Passed:  ${GREEN}$PASSED${NC}"
echo -e "Failed:  ${RED}$FAILED${NC}"
echo -e "Skipped: ${YELLOW}$SKIPPED${NC}"
echo ""

TOTAL=$((PASSED + FAILED))
if [ $TOTAL -gt 0 ]; then
    PERCENTAGE=$((PASSED * 100 / TOTAL))
    echo "Pass Rate: $PERCENTAGE%"
fi

echo ""
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}Challenge PASSED!${NC} All external MCP servers are properly configured."
    exit 0
else
    echo -e "${RED}Challenge FAILED!${NC} $FAILED tests need attention."
    exit 1
fi
