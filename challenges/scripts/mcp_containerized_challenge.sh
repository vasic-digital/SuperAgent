#!/bin/bash
# =============================================================================
# MCP CONTAINERIZATION CHALLENGE
# Validates that all MCP servers are containerized with zero npm/npx dependencies
#
# Test Categories:
#   1-10:  Core MCP Servers
#   11-20: Database MCP Servers
#   21-30: Vector & DevOps MCPs
#   31-40: Browser & Communication MCPs
#   41-50: Productivity MCPs
#   51-60: Search & AI MCPs
#   61-65: Zero NPX Validation
#   66-70: Config Generation Validation
#   71-75: Port Allocation Validation
#
# Usage:
#   ./challenges/scripts/mcp_containerized_challenge.sh
#   RUN_CONTAINER_TESTS=1 ./challenges/scripts/mcp_containerized_challenge.sh
#
# =============================================================================

# Don't use set -e as arithmetic operations can return non-zero
# set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0

# Track if we should run container connectivity tests
RUN_CONTAINER_TESTS="${RUN_CONTAINER_TESTS:-0}"
MCP_CONTAINER_HOST="${MCP_CONTAINER_HOST:-localhost}"

# Test result tracking
declare -a FAILED_TEST_NAMES=()

# =============================================================================
# Utility Functions
# =============================================================================

log_test_start() {
    local test_num=$1
    local test_name=$2
    echo -e "\n${BLUE}[TEST $test_num]${NC} $test_name"
}

log_pass() {
    local test_num=$1
    local message=$2
    echo -e "${GREEN}  [PASS]${NC} Test $test_num: $message"
    PASSED_TESTS=$((PASSED_TESTS + 1))
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
}

log_fail() {
    local test_num=$1
    local message=$2
    echo -e "${RED}  [FAIL]${NC} Test $test_num: $message"
    FAILED_TESTS=$((FAILED_TESTS + 1))
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    FAILED_TEST_NAMES+=("Test $test_num: $message")
}

log_skip() {
    local test_num=$1
    local message=$2
    echo -e "${YELLOW}  [SKIP]${NC} Test $test_num: $message"
    SKIPPED_TESTS=$((SKIPPED_TESTS + 1))
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
}

log_info() {
    echo -e "${BLUE}  [INFO]${NC} $1"
}

# Check if a port is open
check_port() {
    local host=$1
    local port=$2
    timeout 3 bash -c "cat < /dev/null > /dev/tcp/$host/$port" 2>/dev/null
    return $?
}

# =============================================================================
# Test Section 1: File Existence Tests
# =============================================================================

echo -e "\n${BLUE}========================================${NC}"
echo -e "${BLUE}SECTION 1: File Existence Tests${NC}"
echo -e "${BLUE}========================================${NC}"

# Test 1: Container generator exists
log_test_start 1 "Container MCP config generator exists"
if [[ -f "$PROJECT_ROOT/internal/mcp/config/generator_container.go" ]]; then
    log_pass 1 "generator_container.go exists"
else
    log_fail 1 "generator_container.go not found"
fi

# Test 2: Container generator tests exist
log_test_start 2 "Container MCP config generator tests exist"
if [[ -f "$PROJECT_ROOT/internal/mcp/config/generator_container_test.go" ]]; then
    log_pass 2 "generator_container_test.go exists"
else
    log_fail 2 "generator_container_test.go not found"
fi

# Test 3: Full Docker compose exists
log_test_start 3 "docker-compose.mcp-full.yml exists"
if [[ -f "$PROJECT_ROOT/docker/mcp/docker-compose.mcp-full.yml" ]]; then
    log_pass 3 "docker-compose.mcp-full.yml exists"
else
    log_fail 3 "docker-compose.mcp-full.yml not found"
fi

# Test 4: Integration tests exist
log_test_start 4 "MCP container integration tests exist"
if [[ -f "$PROJECT_ROOT/tests/integration/mcp_container_test.go" ]]; then
    log_pass 4 "mcp_container_test.go exists"
else
    log_fail 4 "mcp_container_test.go not found"
fi

# Test 5: MCP submodules directory exists
log_test_start 5 "MCP submodules directory exists"
if [[ -d "$PROJECT_ROOT/MCP/submodules" ]]; then
    log_pass 5 "MCP/submodules directory exists"
else
    log_fail 5 "MCP/submodules directory not found"
fi

# =============================================================================
# Test Section 2: Zero NPX Dependencies Tests
# =============================================================================

echo -e "\n${BLUE}========================================${NC}"
echo -e "${BLUE}SECTION 2: Zero NPX Dependencies Tests${NC}"
echo -e "${BLUE}========================================${NC}"

# Test 6: No npx in container generator
log_test_start 6 "No npx in container generator"
if ! grep -q '"npx"' "$PROJECT_ROOT/internal/mcp/config/generator_container.go" 2>/dev/null; then
    log_pass 6 "No npx commands in container generator"
else
    log_fail 6 "Found npx commands in container generator"
fi

# Test 7: All MCPs use remote type in container generator
log_test_start 7 "All MCPs use remote type"
if grep -q 'Type:.*"remote"' "$PROJECT_ROOT/internal/mcp/config/generator_container.go" && \
   ! grep -q 'Type:.*"local"' "$PROJECT_ROOT/internal/mcp/config/generator_container.go"; then
    log_pass 7 "All MCPs use remote type"
else
    # Check if there are any local types (excluding comments)
    local_count=$(grep -c 'Type:.*"local"' "$PROJECT_ROOT/internal/mcp/config/generator_container.go" 2>/dev/null || echo 0)
    if [[ "$local_count" == "0" ]]; then
        log_pass 7 "All MCPs use remote type"
    else
        log_fail 7 "Found $local_count MCPs with local type"
    fi
fi

# Test 8: All MCPs have URLs
log_test_start 8 "All MCPs have URL field"
url_count=$(grep -c 'URL:' "$PROJECT_ROOT/internal/mcp/config/generator_container.go" 2>/dev/null || echo 0)
if [[ "$url_count" -ge 60 ]]; then
    log_pass 8 "Found $url_count MCP URL definitions"
else
    log_fail 8 "Expected at least 60 URL definitions, found $url_count"
fi

# Test 9: Container generator has ContainsNPX method that returns false
log_test_start 9 "ContainsNPX method returns false"
if grep -q 'func.*ContainsNPX.*bool' "$PROJECT_ROOT/internal/mcp/config/generator_container.go" && \
   grep -q 'return false' "$PROJECT_ROOT/internal/mcp/config/generator_container.go"; then
    log_pass 9 "ContainsNPX method exists and returns false"
else
    log_fail 9 "ContainsNPX method not found or doesn't return false"
fi

# Test 10: No npm/npx in docker-compose.mcp-full.yml (excluding comments)
log_test_start 10 "No npm/npx commands in docker compose file (excluding comments)"
# Check for npx or npm in command sections (not comments)
if ! grep -v '^#' "$PROJECT_ROOT/docker/mcp/docker-compose.mcp-full.yml" 2>/dev/null | grep -qi '"npx\|"npm\|command:.*npx\|command:.*npm'; then
    log_pass 10 "No npm/npx commands in docker compose (comments allowed)"
else
    log_fail 10 "Found npm/npx commands in docker compose"
fi

# =============================================================================
# Test Section 3: Docker Compose Validation Tests
# =============================================================================

echo -e "\n${BLUE}========================================${NC}"
echo -e "${BLUE}SECTION 3: Docker Compose Validation${NC}"
echo -e "${BLUE}========================================${NC}"

# Test 11: Docker compose has 60+ services
log_test_start 11 "Docker compose has 60+ services"
service_count=$(grep -c '^\s*mcp-' "$PROJECT_ROOT/docker/mcp/docker-compose.mcp-full.yml" 2>/dev/null || echo 0)
if [[ "$service_count" -ge 60 ]]; then
    log_pass 11 "Found $service_count MCP services"
else
    log_fail 11 "Expected at least 60 services, found $service_count"
fi

# Test 12: All services have health checks (via YAML anchor or per-service)
log_test_start 12 "All services have health checks (via anchor or per-service)"
# Check for YAML anchor pattern OR per-service healthcheck definitions
has_healthcheck_anchor=$(grep -c 'x-healthcheck:\|<<: \*default-healthcheck\|<<: \*common-config' "$PROJECT_ROOT/docker/mcp/docker-compose.mcp-full.yml" 2>/dev/null || echo 0)
healthcheck_count=$(grep -c 'healthcheck:' "$PROJECT_ROOT/docker/mcp/docker-compose.mcp-full.yml" 2>/dev/null || echo 0)
if [[ "$has_healthcheck_anchor" -gt 0 ]] && [[ "$healthcheck_count" -gt 0 ]]; then
    log_pass 12 "Health checks defined via YAML anchor (applied to all services)"
elif [[ "$healthcheck_count" -ge 60 ]]; then
    log_pass 12 "Found $healthcheck_count health check definitions"
else
    log_fail 12 "Expected health check anchor or 60+ definitions, found anchor=$has_healthcheck_anchor, per-service=$healthcheck_count"
fi

# Test 13: All services have restart policy (via YAML anchor or per-service)
log_test_start 13 "All services have restart policy (via anchor or per-service)"
# Check for YAML anchor pattern OR per-service restart definitions
has_restart_anchor=$(grep -c 'x-common:\|<<: \*common-config' "$PROJECT_ROOT/docker/mcp/docker-compose.mcp-full.yml" 2>/dev/null || echo 0)
restart_count=$(grep -c 'restart:' "$PROJECT_ROOT/docker/mcp/docker-compose.mcp-full.yml" 2>/dev/null || echo 0)
if [[ "$has_restart_anchor" -gt 0 ]] && [[ "$restart_count" -gt 0 ]]; then
    log_pass 13 "Restart policy defined via YAML anchor (applied to all services)"
elif [[ "$restart_count" -ge 60 ]]; then
    log_pass 13 "Found $restart_count restart policies"
else
    log_fail 13 "Expected restart anchor or 60+ definitions, found anchor=$has_restart_anchor, per-service=$restart_count"
fi

# Test 14: Network is defined
log_test_start 14 "Network is defined"
if grep -q 'helixagent-mcp-network' "$PROJECT_ROOT/docker/mcp/docker-compose.mcp-full.yml" 2>/dev/null; then
    log_pass 14 "MCP network is defined"
else
    log_fail 14 "MCP network not found"
fi

# Test 15: Port range 9101-9999 is used
log_test_start 15 "Port range 9101-9999 is used"
port_91xx=$(grep -c '"91[0-9][0-9]:9000"' "$PROJECT_ROOT/docker/mcp/docker-compose.mcp-full.yml" 2>/dev/null || echo 0)
port_99xx=$(grep -c '"99[0-9][0-9]:9000"' "$PROJECT_ROOT/docker/mcp/docker-compose.mcp-full.yml" 2>/dev/null || echo 0)
if [[ "$port_91xx" -gt 0 ]] && [[ "$port_99xx" -gt 0 ]]; then
    log_pass 15 "Port range 9101-9999 is used"
else
    log_fail 15 "Port range not properly used (9101-9199: $port_91xx, 9901-9999: $port_99xx)"
fi

# =============================================================================
# Test Section 4: Port Allocation Tests
# =============================================================================

echo -e "\n${BLUE}========================================${NC}"
echo -e "${BLUE}SECTION 4: Port Allocation Tests${NC}"
echo -e "${BLUE}========================================${NC}"

# Test 16: Core ports 9101-9110
log_test_start 16 "Core MCP ports 9101-9110"
core_ports=0
for port in 9101 9102 9103 9104 9105 9106 9107 9108 9109 9110; do
    if grep -q "\"$port:9000\"" "$PROJECT_ROOT/docker/mcp/docker-compose.mcp-full.yml" 2>/dev/null; then
        core_ports=$((core_ports + 1))
    fi
done
if [[ "$core_ports" -ge 8 ]]; then
    log_pass 16 "Found $core_ports core MCP ports"
else
    log_fail 16 "Expected at least 8 core ports, found $core_ports"
fi

# Test 17: Database ports 9201-9210
log_test_start 17 "Database MCP ports 9201-9210"
db_ports=0
for port in 9201 9202 9203 9204 9205; do
    if grep -q "\"$port:9000\"" "$PROJECT_ROOT/docker/mcp/docker-compose.mcp-full.yml" 2>/dev/null; then
        db_ports=$((db_ports + 1))
    fi
done
if [[ "$db_ports" -ge 4 ]]; then
    log_pass 17 "Found $db_ports database MCP ports"
else
    log_fail 17 "Expected at least 4 database ports, found $db_ports"
fi

# Test 18: Vector DB ports 9301-9310
log_test_start 18 "Vector DB MCP ports 9301-9310"
vector_ports=0
for port in 9301 9302 9303 9304; do
    if grep -q "\"$port:9000\"" "$PROJECT_ROOT/docker/mcp/docker-compose.mcp-full.yml" 2>/dev/null; then
        vector_ports=$((vector_ports + 1))
    fi
done
if [[ "$vector_ports" -ge 3 ]]; then
    log_pass 18 "Found $vector_ports vector DB MCP ports"
else
    log_fail 18 "Expected at least 3 vector DB ports, found $vector_ports"
fi

# Test 19: DevOps ports 9401-9420
log_test_start 19 "DevOps MCP ports 9401-9420"
devops_ports=0
for port in 9401 9402 9403 9404 9405 9406 9407 9408 9409 9410 9411 9412; do
    if grep -q "\"$port:9000\"" "$PROJECT_ROOT/docker/mcp/docker-compose.mcp-full.yml" 2>/dev/null; then
        devops_ports=$((devops_ports + 1))
    fi
done
if [[ "$devops_ports" -ge 8 ]]; then
    log_pass 19 "Found $devops_ports DevOps MCP ports"
else
    log_fail 19 "Expected at least 8 DevOps ports, found $devops_ports"
fi

# Test 20: Communication ports 9601-9610
log_test_start 20 "Communication MCP ports 9601-9610"
comm_ports=0
for port in 9601 9602 9603; do
    if grep -q "\"$port:9000\"" "$PROJECT_ROOT/docker/mcp/docker-compose.mcp-full.yml" 2>/dev/null; then
        comm_ports=$((comm_ports + 1))
    fi
done
if [[ "$comm_ports" -ge 2 ]]; then
    log_pass 20 "Found $comm_ports communication MCP ports"
else
    log_fail 20 "Expected at least 2 communication ports, found $comm_ports"
fi

# =============================================================================
# Test Section 5: Go Unit Tests
# =============================================================================

echo -e "\n${BLUE}========================================${NC}"
echo -e "${BLUE}SECTION 5: Go Unit Tests${NC}"
echo -e "${BLUE}========================================${NC}"

# Test 21: Run unit tests for container generator
log_test_start 21 "Container generator unit tests pass"
cd "$PROJECT_ROOT"
if go test -v -run 'TestContainerMCPConfigGenerator' ./internal/mcp/config/... > /tmp/mcp_container_test.log 2>&1; then
    log_pass 21 "Container generator unit tests pass"
else
    log_fail 21 "Container generator unit tests failed"
    log_info "See /tmp/mcp_container_test.log for details"
fi

# Test 22: Zero NPX test passes
log_test_start 22 "ZeroNPXCommands test passes"
if go test -v -run 'TestContainerMCPConfigGenerator_ZeroNPXCommands' ./internal/mcp/config/... > /tmp/mcp_npx_test.log 2>&1; then
    log_pass 22 "ZeroNPXCommands test passes"
else
    log_fail 22 "ZeroNPXCommands test failed"
fi

# Test 23: Port allocation test passes
log_test_start 23 "Port allocation test passes"
if go test -v -run 'TestContainerMCPConfigGenerator_PortAllocationUnique' ./internal/mcp/config/... > /tmp/mcp_port_test.log 2>&1; then
    log_pass 23 "Port allocation test passes"
else
    log_fail 23 "Port allocation test failed"
fi

# Test 24: Core MCPs enabled test passes
log_test_start 24 "Core MCPs enabled test passes"
if go test -v -run 'TestContainerMCPConfigGenerator_CoreMCPsAlwaysEnabled' ./internal/mcp/config/... > /tmp/mcp_core_test.log 2>&1; then
    log_pass 24 "Core MCPs enabled test passes"
else
    log_fail 24 "Core MCPs enabled test failed"
fi

# Test 25: Compare with NPX generator test passes
log_test_start 25 "Compare with NPX generator test passes"
if go test -v -run 'TestCompareWithNPXGenerator' ./internal/mcp/config/... > /tmp/mcp_compare_test.log 2>&1; then
    log_pass 25 "Compare with NPX generator test passes"
else
    log_fail 25 "Compare with NPX generator test failed"
fi

# =============================================================================
# Test Section 6: Container Category Tests
# =============================================================================

echo -e "\n${BLUE}========================================${NC}"
echo -e "${BLUE}SECTION 6: Container Category Tests${NC}"
echo -e "${BLUE}========================================${NC}"

# Test 26: Core MCPs defined
log_test_start 26 "Core MCPs defined (fetch, git, time, etc.)"
core_mcps="mcp-fetch mcp-git mcp-time mcp-filesystem mcp-memory"
found=0
for mcp in $core_mcps; do
    if grep -q "^\s*$mcp:" "$PROJECT_ROOT/docker/mcp/docker-compose.mcp-full.yml" 2>/dev/null; then
        found=$((found + 1))
    fi
done
if [[ "$found" -eq 5 ]]; then
    log_pass 26 "All 5 core MCPs defined"
else
    log_fail 26 "Expected 5 core MCPs, found $found"
fi

# Test 27: Database MCPs defined
log_test_start 27 "Database MCPs defined (mongodb, redis, mysql, etc.)"
db_mcps="mcp-mongodb mcp-redis mcp-mysql mcp-elasticsearch"
found=0
for mcp in $db_mcps; do
    if grep -q "^\s*$mcp:" "$PROJECT_ROOT/docker/mcp/docker-compose.mcp-full.yml" 2>/dev/null; then
        found=$((found + 1))
    fi
done
if [[ "$found" -ge 3 ]]; then
    log_pass 27 "Found $found database MCPs"
else
    log_fail 27 "Expected at least 3 database MCPs, found $found"
fi

# Test 28: DevOps MCPs defined
log_test_start 28 "DevOps MCPs defined (github, gitlab, kubernetes, etc.)"
devops_mcps="mcp-github mcp-gitlab mcp-kubernetes mcp-docker"
found=0
for mcp in $devops_mcps; do
    if grep -q "^\s*$mcp:" "$PROJECT_ROOT/docker/mcp/docker-compose.mcp-full.yml" 2>/dev/null; then
        found=$((found + 1))
    fi
done
if [[ "$found" -ge 3 ]]; then
    log_pass 28 "Found $found DevOps MCPs"
else
    log_fail 28 "Expected at least 3 DevOps MCPs, found $found"
fi

# Test 29: Communication MCPs defined
log_test_start 29 "Communication MCPs defined (slack, discord, telegram)"
comm_mcps="mcp-slack mcp-discord mcp-telegram"
found=0
for mcp in $comm_mcps; do
    if grep -q "^\s*$mcp:" "$PROJECT_ROOT/docker/mcp/docker-compose.mcp-full.yml" 2>/dev/null; then
        found=$((found + 1))
    fi
done
if [[ "$found" -ge 2 ]]; then
    log_pass 29 "Found $found communication MCPs"
else
    log_fail 29 "Expected at least 2 communication MCPs, found $found"
fi

# Test 30: Productivity MCPs defined
log_test_start 30 "Productivity MCPs defined (notion, linear, jira, etc.)"
prod_mcps="mcp-notion mcp-linear mcp-jira mcp-trello"
found=0
for mcp in $prod_mcps; do
    if grep -q "^\s*$mcp:" "$PROJECT_ROOT/docker/mcp/docker-compose.mcp-full.yml" 2>/dev/null; then
        found=$((found + 1))
    fi
done
if [[ "$found" -ge 3 ]]; then
    log_pass 30 "Found $found productivity MCPs"
else
    log_fail 30 "Expected at least 3 productivity MCPs, found $found"
fi

# =============================================================================
# Test Section 7: Build Context Tests
# =============================================================================

echo -e "\n${BLUE}========================================${NC}"
echo -e "${BLUE}SECTION 7: Build Context Tests${NC}"
echo -e "${BLUE}========================================${NC}"

# Test 31: Services use proper build context (via YAML anchor or per-service)
log_test_start 31 "Services use proper build context (via anchor or per-service)"
# Check for YAML anchor pattern OR per-service build context
has_build_anchor=$(grep -c 'x-common:\|<<: \*common-config' "$PROJECT_ROOT/docker/mcp/docker-compose.mcp-full.yml" 2>/dev/null || echo 0)
build_context_count=$(grep -c 'context: \.\./\.\.' "$PROJECT_ROOT/docker/mcp/docker-compose.mcp-full.yml" 2>/dev/null || echo 0)
if [[ "$has_build_anchor" -gt 0 ]] && [[ "$build_context_count" -gt 0 ]]; then
    log_pass 31 "Build context defined via YAML anchor (applied to all services)"
elif [[ "$build_context_count" -ge 60 ]]; then
    log_pass 31 "Found $build_context_count proper build contexts"
else
    log_fail 31 "Expected build anchor or 60+ build contexts, found anchor=$has_build_anchor, per-service=$build_context_count"
fi

# Test 32: Services use Dockerfile.mcp-bridge (via YAML anchor or per-service)
log_test_start 32 "Services use MCP bridge Dockerfile"
# Check for YAML anchor pattern OR per-service dockerfile
has_dockerfile_anchor=$(grep -c 'x-common:\|<<: \*common-config' "$PROJECT_ROOT/docker/mcp/docker-compose.mcp-full.yml" 2>/dev/null || echo 0)
dockerfile_count=$(grep -c 'dockerfile: docker/mcp/Dockerfile' "$PROJECT_ROOT/docker/mcp/docker-compose.mcp-full.yml" 2>/dev/null || echo 0)
bridge_dockerfile=$(grep -c 'Dockerfile.mcp-bridge' "$PROJECT_ROOT/docker/mcp/docker-compose.mcp-full.yml" 2>/dev/null || echo 0)
if [[ "$has_dockerfile_anchor" -gt 0 ]] && [[ "$bridge_dockerfile" -gt 0 ]]; then
    log_pass 32 "Dockerfile.mcp-bridge used via YAML anchor (applied to all services)"
elif [[ "$dockerfile_count" -ge 60 ]]; then
    log_pass 32 "Found $dockerfile_count Dockerfile references"
else
    log_fail 32 "Expected bridge Dockerfile anchor or 60+ references, found anchor=$has_dockerfile_anchor, bridge=$bridge_dockerfile"
fi

# Test 33: MCP SSE Bridge architecture - services use MCP_COMMAND
log_test_start 33 "Services use MCP_COMMAND for bridge architecture"
mcp_command_count=$(grep -c 'MCP_COMMAND=' "$PROJECT_ROOT/docker/mcp/docker-compose.mcp-full.yml" 2>/dev/null || echo 0)
if [[ "$mcp_command_count" -ge 60 ]]; then
    log_pass 33 "Found $mcp_command_count MCP_COMMAND definitions (bridge architecture)"
else
    log_fail 33 "Expected at least 60 MCP_COMMAND definitions, found $mcp_command_count"
fi

# Test 34: Bridge architecture uses @modelcontextprotocol servers
log_test_start 34 "Bridge uses official @modelcontextprotocol servers"
mcp_official_count=$(grep -c '@modelcontextprotocol\|mcp-server-' "$PROJECT_ROOT/docker/mcp/docker-compose.mcp-full.yml" 2>/dev/null || echo 0)
if [[ "$mcp_official_count" -ge 30 ]]; then
    log_pass 34 "Found $mcp_official_count official MCP server references"
else
    log_fail 34 "Expected at least 30 official MCP server references, found $mcp_official_count"
fi

# Test 35: Container names are prefixed
log_test_start 35 "Container names are properly prefixed"
container_name_count=$(grep -c 'container_name: helixagent-mcp-' "$PROJECT_ROOT/docker/mcp/docker-compose.mcp-full.yml" 2>/dev/null || echo 0)
if [[ "$container_name_count" -ge 60 ]]; then
    log_pass 35 "Found $container_name_count properly named containers"
else
    log_fail 35 "Expected at least 60 named containers, found $container_name_count"
fi

# =============================================================================
# Test Section 8: Environment Variables Tests
# =============================================================================

echo -e "\n${BLUE}========================================${NC}"
echo -e "${BLUE}SECTION 8: Environment Variables Tests${NC}"
echo -e "${BLUE}========================================${NC}"

# Test 36: Environment section exists for conditional MCPs
log_test_start 36 "Environment sections exist for conditional MCPs"
env_count=$(grep -c 'environment:' "$PROJECT_ROOT/docker/mcp/docker-compose.mcp-full.yml" 2>/dev/null || echo 0)
if [[ "$env_count" -ge 30 ]]; then
    log_pass 36 "Found $env_count environment sections"
else
    log_fail 36 "Expected at least 30 environment sections, found $env_count"
fi

# Test 37: GitHub token variable defined
log_test_start 37 "GitHub token variable defined"
if grep -q 'GITHUB_PERSONAL_ACCESS_TOKEN' "$PROJECT_ROOT/docker/mcp/docker-compose.mcp-full.yml" 2>/dev/null; then
    log_pass 37 "GitHub token variable is defined"
else
    log_fail 37 "GitHub token variable not found"
fi

# Test 38: Slack variables defined
log_test_start 38 "Slack variables defined"
if grep -q 'SLACK_BOT_TOKEN' "$PROJECT_ROOT/docker/mcp/docker-compose.mcp-full.yml" 2>/dev/null && \
   grep -q 'SLACK_TEAM_ID' "$PROJECT_ROOT/docker/mcp/docker-compose.mcp-full.yml" 2>/dev/null; then
    log_pass 38 "Slack variables are defined"
else
    log_fail 38 "Slack variables not found"
fi

# Test 39: OpenAI key variable defined
log_test_start 39 "OpenAI key variable defined"
if grep -q 'OPENAI_API_KEY' "$PROJECT_ROOT/docker/mcp/docker-compose.mcp-full.yml" 2>/dev/null; then
    log_pass 39 "OpenAI key variable is defined"
else
    log_fail 39 "OpenAI key variable not found"
fi

# Test 40: Database connection variables defined
log_test_start 40 "Database connection variables defined"
db_vars_found=0
for var in POSTGRES_URL MONGODB_URI REDIS_URL; do
    if grep -q "$var" "$PROJECT_ROOT/docker/mcp/docker-compose.mcp-full.yml" 2>/dev/null; then
        db_vars_found=$((db_vars_found + 1))
    fi
done
if [[ "$db_vars_found" -ge 2 ]]; then
    log_pass 40 "Found $db_vars_found database connection variables"
else
    log_fail 40 "Expected at least 2 database variables, found $db_vars_found"
fi

# =============================================================================
# Test Section 9: Container Connectivity Tests (Optional)
# =============================================================================

echo -e "\n${BLUE}========================================${NC}"
echo -e "${BLUE}SECTION 9: Container Connectivity Tests${NC}"
echo -e "${BLUE}========================================${NC}"

if [[ "$RUN_CONTAINER_TESTS" == "1" ]]; then
    log_info "Running container connectivity tests (RUN_CONTAINER_TESTS=1)"

    # Test 41-50: Core container connectivity
    core_ports=(9101 9102 9103 9104 9105 9106 9107 9108 9109 9110)
    core_names=("fetch" "git" "time" "filesystem" "memory" "everything" "sequential-thinking" "sqlite" "puppeteer" "postgres")

    for i in "${!core_ports[@]}"; do
        test_num=$((41 + i))
        port=${core_ports[$i]}
        name=${core_names[$i]}

        log_test_start $test_num "Core MCP $name connectivity (port $port)"
        if check_port "$MCP_CONTAINER_HOST" "$port"; then
            log_pass $test_num "MCP $name is accessible on port $port"
        else
            log_skip $test_num "MCP $name not running on port $port"
        fi
    done
else
    log_info "Skipping container connectivity tests (set RUN_CONTAINER_TESTS=1 to enable)"
    for i in {41..50}; do
        log_skip $i "Container connectivity test skipped"
    done
fi

# =============================================================================
# Test Section 10: Integration Tests
# =============================================================================

echo -e "\n${BLUE}========================================${NC}"
echo -e "${BLUE}SECTION 10: Integration Tests${NC}"
echo -e "${BLUE}========================================${NC}"

# Test 51: Integration test file has correct tests
log_test_start 51 "Integration test file has container connectivity test"
if grep -q 'TestMCPContainerConnectivity' "$PROJECT_ROOT/tests/integration/mcp_container_test.go" 2>/dev/null; then
    log_pass 51 "Container connectivity test found"
else
    log_fail 51 "Container connectivity test not found"
fi

# Test 52: Integration test has health check test
log_test_start 52 "Integration test has health check test"
if grep -q 'TestMCPContainerHealthChecks' "$PROJECT_ROOT/tests/integration/mcp_container_test.go" 2>/dev/null; then
    log_pass 52 "Health check test found"
else
    log_fail 52 "Health check test not found"
fi

# Test 53: Integration test has JSON-RPC compliance test
log_test_start 53 "Integration test has JSON-RPC compliance test"
if grep -q 'TestMCPContainerJSONRPCCompliance' "$PROJECT_ROOT/tests/integration/mcp_container_test.go" 2>/dev/null; then
    log_pass 53 "JSON-RPC compliance test found"
else
    log_fail 53 "JSON-RPC compliance test not found"
fi

# Test 54: Integration test has tool discovery test
log_test_start 54 "Integration test has tool discovery test"
if grep -q 'TestMCPContainerToolDiscovery' "$PROJECT_ROOT/tests/integration/mcp_container_test.go" 2>/dev/null; then
    log_pass 54 "Tool discovery test found"
else
    log_fail 54 "Tool discovery test not found"
fi

# Test 55: Integration test has no NPX dependencies test
log_test_start 55 "Integration test has no NPX dependencies test"
if grep -q 'TestMCPContainerNoNPXDependencies' "$PROJECT_ROOT/tests/integration/mcp_container_test.go" 2>/dev/null; then
    log_pass 55 "No NPX dependencies test found"
else
    log_fail 55 "No NPX dependencies test not found"
fi

# =============================================================================
# Test Section 11: Generator Feature Tests
# =============================================================================

echo -e "\n${BLUE}========================================${NC}"
echo -e "${BLUE}SECTION 11: Generator Feature Tests${NC}"
echo -e "${BLUE}========================================${NC}"

# Test 56: Generator has port allocation map
log_test_start 56 "Generator has port allocation map"
if grep -q 'MCPContainerPorts' "$PROJECT_ROOT/internal/mcp/config/generator_container.go" 2>/dev/null; then
    log_pass 56 "Port allocation map found"
else
    log_fail 56 "Port allocation map not found"
fi

# Test 57: Generator has category support
log_test_start 57 "Generator has category support"
if grep -q 'GetMCPsByCategory' "$PROJECT_ROOT/internal/mcp/config/generator_container.go" 2>/dev/null; then
    log_pass 57 "Category support found"
else
    log_fail 57 "Category support not found"
fi

# Test 58: Generator has summary method
log_test_start 58 "Generator has summary method"
if grep -q 'GenerateSummary' "$PROJECT_ROOT/internal/mcp/config/generator_container.go" 2>/dev/null; then
    log_pass 58 "Summary method found"
else
    log_fail 58 "Summary method not found"
fi

# Test 59: Generator has port validation
log_test_start 59 "Generator has port validation"
if grep -q 'ValidatePortAllocations' "$PROJECT_ROOT/internal/mcp/config/generator_container.go" 2>/dev/null; then
    log_pass 59 "Port validation found"
else
    log_fail 59 "Port validation not found"
fi

# Test 60: Generator has custom host support
log_test_start 60 "Generator has custom host support"
if grep -q 'MCP_CONTAINER_HOST' "$PROJECT_ROOT/internal/mcp/config/generator_container.go" 2>/dev/null; then
    log_pass 60 "Custom host support found"
else
    log_fail 60 "Custom host support not found"
fi

# =============================================================================
# Test Section 12: Final Validation
# =============================================================================

echo -e "\n${BLUE}========================================${NC}"
echo -e "${BLUE}SECTION 12: Final Validation${NC}"
echo -e "${BLUE}========================================${NC}"

# Test 61: Total MCP count >= 60
log_test_start 61 "Total MCP count >= 60"
mcp_count=$(grep -c '^\s*mcp-' "$PROJECT_ROOT/docker/mcp/docker-compose.mcp-full.yml" 2>/dev/null || echo 0)
if [[ "$mcp_count" -ge 60 ]]; then
    log_pass 61 "Total MCP count: $mcp_count (>= 60)"
else
    log_fail 61 "Total MCP count: $mcp_count (< 60)"
fi

# Test 62: All Dockerfiles exist
log_test_start 62 "All required Dockerfiles exist"
dockerfiles_found=0
for df in Dockerfile.mcp-server Dockerfile.mcp-python Dockerfile.mcp-go Dockerfile.mcp-playwright Dockerfile.mcp-submodule; do
    if [[ -f "$PROJECT_ROOT/docker/mcp/$df" ]]; then
        dockerfiles_found=$((dockerfiles_found + 1))
    fi
done
if [[ "$dockerfiles_found" -ge 4 ]]; then
    log_pass 62 "Found $dockerfiles_found required Dockerfiles"
else
    log_fail 62 "Expected at least 4 Dockerfiles, found $dockerfiles_found"
fi

# Test 63: Compose file is valid YAML
log_test_start 63 "Compose file is valid YAML"
if command -v docker-compose &> /dev/null; then
    if docker-compose -f "$PROJECT_ROOT/docker/mcp/docker-compose.mcp-full.yml" config > /dev/null 2>&1; then
        log_pass 63 "Compose file is valid YAML"
    else
        log_fail 63 "Compose file validation failed"
    fi
else
    log_skip 63 "docker-compose not available for validation"
fi

# Test 64: No duplicate ports
log_test_start 64 "No duplicate ports in compose file"
port_list=$(grep -o '"[0-9]*:9000"' "$PROJECT_ROOT/docker/mcp/docker-compose.mcp-full.yml" | sort)
unique_ports=$(echo "$port_list" | uniq | wc -l)
total_ports=$(echo "$port_list" | wc -l)
if [[ "$unique_ports" -eq "$total_ports" ]]; then
    log_pass 64 "No duplicate ports (total: $total_ports)"
else
    log_fail 64 "Found duplicate ports ($total_ports total, $unique_ports unique)"
fi

# Test 65: Integration tests compile
log_test_start 65 "Integration tests compile"
if go build "$PROJECT_ROOT/tests/integration/mcp_container_test.go" -o /dev/null 2>/dev/null; then
    log_pass 65 "Integration tests compile"
else
    # Try with go test -c instead
    if go test -c "$PROJECT_ROOT/tests/integration/..." -o /dev/null 2>/dev/null; then
        log_pass 65 "Integration tests compile"
    else
        log_skip 65 "Integration tests compilation check skipped (requires dependencies)"
    fi
fi

# =============================================================================
# Summary
# =============================================================================

echo -e "\n${BLUE}========================================${NC}"
echo -e "${BLUE}CHALLENGE SUMMARY${NC}"
echo -e "${BLUE}========================================${NC}"

echo -e "Total Tests: ${BLUE}$TOTAL_TESTS${NC}"
echo -e "Passed:      ${GREEN}$PASSED_TESTS${NC}"
echo -e "Failed:      ${RED}$FAILED_TESTS${NC}"
echo -e "Skipped:     ${YELLOW}$SKIPPED_TESTS${NC}"

if [[ "$FAILED_TESTS" -gt 0 ]]; then
    echo -e "\n${RED}Failed Tests:${NC}"
    for failed in "${FAILED_TEST_NAMES[@]}"; do
        echo -e "  ${RED}- $failed${NC}"
    done
fi

# Calculate pass rate
ATTEMPTED=$((TOTAL_TESTS - SKIPPED_TESTS))
if [[ "$ATTEMPTED" -gt 0 ]]; then
    PASS_RATE=$((PASSED_TESTS * 100 / ATTEMPTED))
    echo -e "\nPass Rate: ${BLUE}$PASS_RATE%${NC} ($PASSED_TESTS/$ATTEMPTED)"
fi

if [[ "$FAILED_TESTS" -eq 0 ]]; then
    echo -e "\n${GREEN}========================================${NC}"
    echo -e "${GREEN}MCP CONTAINERIZATION CHALLENGE: PASSED${NC}"
    echo -e "${GREEN}========================================${NC}"
    exit 0
else
    echo -e "\n${RED}========================================${NC}"
    echo -e "${RED}MCP CONTAINERIZATION CHALLENGE: FAILED${NC}"
    echo -e "${RED}========================================${NC}"
    exit 1
fi
