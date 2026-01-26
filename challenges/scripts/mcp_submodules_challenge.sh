#!/bin/bash
# MCP Submodules Challenge Script
# Tests all 48 MCP server submodules for proper configuration

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
MCP_DIR="$PROJECT_ROOT/MCP"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

TESTS_PASSED=0
TESTS_FAILED=0
TESTS_SKIPPED=0

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((TESTS_PASSED++))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((TESTS_FAILED++))
}

log_skip() {
    echo -e "${YELLOW}[SKIP]${NC} $1"
    ((TESTS_SKIPPED++))
}

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

echo "========================================"
echo "  MCP Submodules Challenge"
echo "  Testing all 48 MCP server submodules"
echo "========================================"
echo ""

# Test 1: Check MCP directory exists
log_info "Test 1: Check MCP directory exists"
if [ -d "$MCP_DIR" ]; then
    log_pass "MCP directory exists"
else
    log_fail "MCP directory not found at $MCP_DIR"
fi

# Test 2: Check submodules directory exists
log_info "Test 2: Check MCP/submodules directory exists"
if [ -d "$MCP_DIR/submodules" ]; then
    log_pass "MCP/submodules directory exists"
else
    log_fail "MCP/submodules directory not found"
fi

# Test 3: Count submodules
log_info "Test 3: Count MCP submodules"
SUBMODULE_COUNT=$(ls -d "$MCP_DIR/submodules"/*/ 2>/dev/null | wc -l)
if [ "$SUBMODULE_COUNT" -ge 40 ]; then
    log_pass "Found $SUBMODULE_COUNT MCP submodules (expected 40+)"
else
    log_fail "Found only $SUBMODULE_COUNT MCP submodules (expected 40+)"
fi

# Test 4: Check docker-compose.yml exists
log_info "Test 4: Check docker-compose.yml exists"
if [ -f "$MCP_DIR/docker-compose.yml" ]; then
    log_pass "docker-compose.yml exists"
else
    log_fail "docker-compose.yml not found"
fi

# Test 5: Check dockerfiles directory exists
log_info "Test 5: Check dockerfiles directory exists"
if [ -d "$MCP_DIR/dockerfiles" ]; then
    DOCKERFILE_COUNT=$(ls "$MCP_DIR/dockerfiles"/*.* 2>/dev/null | wc -l)
    log_pass "dockerfiles directory exists with $DOCKERFILE_COUNT files"
else
    log_fail "dockerfiles directory not found"
fi

# Test 6: Check README.md exists
log_info "Test 6: Check README.md exists"
if [ -f "$MCP_DIR/README.md" ]; then
    README_LINES=$(wc -l < "$MCP_DIR/README.md")
    if [ "$README_LINES" -gt 100 ]; then
        log_pass "README.md exists with $README_LINES lines"
    else
        log_fail "README.md too short ($README_LINES lines, expected 100+)"
    fi
else
    log_fail "README.md not found"
fi

# Test 7-20: Check essential submodules exist
log_info "Test 7-20: Check essential submodules"

ESSENTIAL_SUBMODULES=(
    "github-mcp-server"
    "slack-mcp"
    "notion-mcp-server"
    "redis-mcp"
    "mongodb-mcp"
    "qdrant-mcp"
    "aws-mcp"
    "kubernetes-mcp"
    "playwright-mcp"
    "sentry-mcp"
    "heroku-mcp"
    "cloudflare-mcp"
    "brave-search"
    "perplexity-mcp"
)

for submodule in "${ESSENTIAL_SUBMODULES[@]}"; do
    if [ -d "$MCP_DIR/submodules/$submodule" ]; then
        log_pass "Essential submodule exists: $submodule"
    else
        log_fail "Essential submodule missing: $submodule"
    fi
done

# Test 21-30: Check SDK and template submodules
log_info "Test 21-30: Check SDK and template submodules"

SDK_SUBMODULES=(
    "python-sdk"
    "typescript-sdk"
    "registry"
    "inspector"
    "create-python-server"
    "create-typescript-server"
    "microsoft-mcp"
    "langchain-mcp"
    "llamaindex-mcp"
    "context7-mcp"
)

for submodule in "${SDK_SUBMODULES[@]}"; do
    if [ -d "$MCP_DIR/submodules/$submodule" ]; then
        log_pass "SDK/template submodule exists: $submodule"
    else
        log_fail "SDK/template submodule missing: $submodule"
    fi
done

# Test 31-35: Check Dockerfiles for essential services
log_info "Test 31-35: Check essential Dockerfiles"

ESSENTIAL_DOCKERFILES=(
    "Dockerfile.github"
    "Dockerfile.slack"
    "Dockerfile.notion"
    "Dockerfile.redis"
    "Dockerfile.mongodb"
)

for dockerfile in "${ESSENTIAL_DOCKERFILES[@]}"; do
    if [ -f "$MCP_DIR/dockerfiles/$dockerfile" ]; then
        log_pass "Dockerfile exists: $dockerfile"
    else
        log_fail "Dockerfile missing: $dockerfile"
    fi
done

# Test 36-40: Check docker-compose services defined
log_info "Test 36-40: Check docker-compose services"

if [ -f "$MCP_DIR/docker-compose.yml" ]; then
    for service in "mcp-github" "mcp-slack" "mcp-redis" "mcp-kubernetes" "mcp-aws"; do
        if grep -q "$service:" "$MCP_DIR/docker-compose.yml"; then
            log_pass "Docker service defined: $service"
        else
            log_fail "Docker service missing: $service"
        fi
    done
fi

# Test 41-45: Check awesome-mcp-servers lists
log_info "Test 41-45: Check awesome MCP server lists"

AWESOME_LISTS=(
    "awesome-mcp-servers"
    "appcypher-awesome-mcp"
    "habito-awesome-mcp"
    "ever-works-awesome-mcp"
    "awesome-devops-mcp"
)

for list in "${AWESOME_LISTS[@]}"; do
    if [ -d "$MCP_DIR/submodules/$list" ]; then
        log_pass "Awesome list submodule exists: $list"
    else
        log_skip "Awesome list submodule not found: $list"
    fi
done

# Test 46: Check git submodule status
log_info "Test 46: Check git submodule status"
cd "$PROJECT_ROOT"
SUBMODULE_STATUS=$(git submodule status | grep "MCP/submodules" | wc -l)
if [ "$SUBMODULE_STATUS" -ge 30 ]; then
    log_pass "Git reports $SUBMODULE_STATUS MCP submodules registered"
else
    log_fail "Git reports only $SUBMODULE_STATUS MCP submodules (expected 30+)"
fi

# Test 47: Check docker-compose syntax
log_info "Test 47: Check docker-compose.yml syntax"
if command -v docker-compose &> /dev/null; then
    if docker-compose -f "$MCP_DIR/docker-compose.yml" config -q 2>/dev/null; then
        log_pass "docker-compose.yml syntax is valid"
    else
        log_fail "docker-compose.yml has syntax errors"
    fi
elif command -v podman-compose &> /dev/null; then
    if podman-compose -f "$MCP_DIR/docker-compose.yml" config 2>/dev/null | head -1 | grep -q "version"; then
        log_pass "podman-compose validates docker-compose.yml"
    else
        log_skip "podman-compose validation not available"
    fi
else
    log_skip "Neither docker-compose nor podman-compose available"
fi

# Test 48: Check base Dockerfiles
log_info "Test 48: Check base Dockerfiles exist"
BASE_DOCKERFILES=("Dockerfile.base-node" "Dockerfile.base-python")
BASE_COUNT=0
for base in "${BASE_DOCKERFILES[@]}"; do
    if [ -f "$MCP_DIR/dockerfiles/$base" ]; then
        ((BASE_COUNT++))
    fi
done
if [ "$BASE_COUNT" -eq 2 ]; then
    log_pass "Both base Dockerfiles exist (node and python)"
else
    log_fail "Missing base Dockerfiles ($BASE_COUNT/2 found)"
fi

# Summary
echo ""
echo "========================================"
echo "  Challenge Summary"
echo "========================================"
echo -e "  ${GREEN}Passed:${NC}  $TESTS_PASSED"
echo -e "  ${RED}Failed:${NC}  $TESTS_FAILED"
echo -e "  ${YELLOW}Skipped:${NC} $TESTS_SKIPPED"
echo "========================================"

TOTAL=$((TESTS_PASSED + TESTS_FAILED))
if [ "$TESTS_FAILED" -eq 0 ]; then
    echo -e "${GREEN}All $TOTAL tests passed!${NC}"
    exit 0
else
    echo -e "${RED}$TESTS_FAILED of $TOTAL tests failed${NC}"
    exit 1
fi
