#!/bin/bash
# HelixAgent Challenge: Remote Services Configuration
# Tests: ~25 - Remote service config, env var overrides, health check behavior
# Usage: ./challenges/scripts/remote_services_challenge.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

PASSED=0
FAILED=0
TOTAL=0

pass() { PASSED=$((PASSED+1)); TOTAL=$((TOTAL+1)); echo -e "${GREEN}[PASS]${NC} $1"; }
fail() { FAILED=$((FAILED+1)); TOTAL=$((TOTAL+1)); echo -e "${RED}[FAIL]${NC} $1"; }
info() { echo -e "${BLUE}[INFO]${NC} $1"; }

check_contains() {
    local desc="$1"
    local file="$2"
    local pattern="$3"
    if grep -q "$pattern" "$file" 2>/dev/null; then
        pass "$desc"
    else
        fail "$desc - Pattern '$pattern' not found in $file"
    fi
}

check_file() {
    local desc="$1"
    local file="$2"
    if [ -f "$file" ]; then
        pass "$desc"
    else
        fail "$desc - File not found: $file"
    fi
}

echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║        REMOTE SERVICES CONFIGURATION CHALLENGE (~25 tests)     ║"
echo "╚══════════════════════════════════════════════════════════════════╝"
echo ""

# ============================================================================
# SECTION 1: Remote Configuration Support
# ============================================================================
info "Section 1: Remote Configuration Support"

check_contains "ServiceEndpoint has Remote field" "$PROJECT_ROOT/internal/config/config.go" 'Remote.*bool.*yaml:"remote"'
check_contains "ServiceEndpoint has URL field" "$PROJECT_ROOT/internal/config/config.go" 'URL.*string.*yaml:"url"'
check_contains "URL takes precedence in ResolvedURL" "$PROJECT_ROOT/internal/config/config.go" 'if e.URL != ""'
check_contains "Remote services example config exists" "$PROJECT_ROOT/configs/remote-services-example.yaml" "remote: true"

# ============================================================================
# SECTION 2: Environment Variable Overrides
# ============================================================================
info "Section 2: Environment Variable Overrides"

check_contains "SVC_POSTGRESQL env prefix" "$PROJECT_ROOT/internal/config/config.go" 'SVC_POSTGRESQL'
check_contains "SVC_REDIS env prefix" "$PROJECT_ROOT/internal/config/config.go" 'SVC_REDIS'
check_contains "SVC_COGNEE env prefix" "$PROJECT_ROOT/internal/config/config.go" 'SVC_COGNEE'
check_contains "SVC_CHROMADB env prefix" "$PROJECT_ROOT/internal/config/config.go" 'SVC_CHROMADB'
check_contains "SVC_PROMETHEUS env prefix" "$PROJECT_ROOT/internal/config/config.go" 'SVC_PROMETHEUS'
check_contains "SVC_GRAFANA env prefix" "$PROJECT_ROOT/internal/config/config.go" 'SVC_GRAFANA'
check_contains "SVC_QDRANT env prefix" "$PROJECT_ROOT/internal/config/config.go" 'SVC_QDRANT'
check_contains "SVC_WEAVIATE env prefix" "$PROJECT_ROOT/internal/config/config.go" 'SVC_WEAVIATE'
check_contains "SVC_LANGCHAIN env prefix" "$PROJECT_ROOT/internal/config/config.go" 'SVC_LANGCHAIN'
check_contains "SVC_LLAMAINDEX env prefix" "$PROJECT_ROOT/internal/config/config.go" 'SVC_LLAMAINDEX'
check_contains "_HOST env suffix support" "$PROJECT_ROOT/internal/config/config.go" '_HOST"'
check_contains "_PORT env suffix support" "$PROJECT_ROOT/internal/config/config.go" '_PORT"'
check_contains "_REMOTE env suffix support" "$PROJECT_ROOT/internal/config/config.go" '_REMOTE"'
check_contains "_ENABLED env suffix support" "$PROJECT_ROOT/internal/config/config.go" '_ENABLED"'
check_contains "_URL env suffix support" "$PROJECT_ROOT/internal/config/config.go" '_URL"'

# ============================================================================
# SECTION 3: Remote Service Boot Behavior
# ============================================================================
info "Section 3: Remote Service Boot Behavior"

check_contains "Boot manager skips compose for remote" "$PROJECT_ROOT/internal/services/boot_manager.go" "Remote"
check_contains "Boot manager marks remote status" "$PROJECT_ROOT/internal/services/boot_manager.go" '"remote"'
check_contains "Health check still runs for remote" "$PROJECT_ROOT/internal/services/boot_manager.go" "CheckWithRetry"

# ============================================================================
# SECTION 4: Remote Example Config
# ============================================================================
info "Section 4: Remote Example Config"

check_file "Remote example YAML exists" "$PROJECT_ROOT/configs/remote-services-example.yaml"
check_contains "PostgreSQL remote example" "$PROJECT_ROOT/configs/remote-services-example.yaml" "postgresql:"
check_contains "Redis remote example" "$PROJECT_ROOT/configs/remote-services-example.yaml" "redis:"
check_contains "Cognee remote example" "$PROJECT_ROOT/configs/remote-services-example.yaml" "cognee:"
check_contains "Example shows remote host" "$PROJECT_ROOT/configs/remote-services-example.yaml" "production.example.com"

# ============================================================================
# SECTION 5: Unit Tests Cover Remote Scenarios
# ============================================================================
info "Section 5: Unit Tests Cover Remote Scenarios"

check_contains "Test for remote flag" "$PROJECT_ROOT/internal/config/services_test.go" "TestRemoteFlag"
check_contains "Test for env overrides" "$PROJECT_ROOT/internal/config/services_test.go" "TestEnvironmentOverrides"
check_contains "Test remote skips compose" "$PROJECT_ROOT/internal/services/boot_manager_test.go" "RemoteSkipsCompose"

# ============================================================================
# SUMMARY
# ============================================================================
echo ""
echo "════════════════════════════════════════════════════════════════════"
echo -e "  Results: ${GREEN}${PASSED} passed${NC} / ${RED}${FAILED} failed${NC} / ${TOTAL} total"
if [ "$FAILED" -eq 0 ]; then
    echo -e "  Status: ${GREEN}ALL TESTS PASSED${NC}"
else
    echo -e "  Status: ${RED}SOME TESTS FAILED${NC}"
fi
echo "════════════════════════════════════════════════════════════════════"

exit $FAILED
