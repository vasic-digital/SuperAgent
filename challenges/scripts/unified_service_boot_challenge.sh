#!/bin/bash
# HelixAgent Challenge: Unified Service Boot
# Tests: ~40 - Services config, boot manager, health checking, shutdown
# Usage: ./challenges/scripts/unified_service_boot_challenge.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PASSED=0
FAILED=0
TOTAL=0

pass() { PASSED=$((PASSED+1)); TOTAL=$((TOTAL+1)); echo -e "${GREEN}[PASS]${NC} $1"; }
fail() { FAILED=$((FAILED+1)); TOTAL=$((TOTAL+1)); echo -e "${RED}[FAIL]${NC} $1"; }
info() { echo -e "${BLUE}[INFO]${NC} $1"; }

check() {
    local desc="$1"
    shift
    if "$@" >/dev/null 2>&1; then
        pass "$desc"
    else
        fail "$desc"
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

echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║          UNIFIED SERVICE BOOT CHALLENGE (~40 tests)            ║"
echo "╚══════════════════════════════════════════════════════════════════╝"
echo ""

# ============================================================================
# SECTION 1: Services Config Structure
# ============================================================================
info "Section 1: Services Config Structure"

check_file "Config file exists" "$PROJECT_ROOT/internal/config/config.go"
check_contains "ServiceEndpoint type defined" "$PROJECT_ROOT/internal/config/config.go" "type ServiceEndpoint struct"
check_contains "ServicesConfig type defined" "$PROJECT_ROOT/internal/config/config.go" "type ServicesConfig struct"
check_contains "Config has Services field" "$PROJECT_ROOT/internal/config/config.go" "Services.*ServicesConfig"
check_contains "ServiceEndpoint has Host field" "$PROJECT_ROOT/internal/config/config.go" 'Host.*string.*yaml:"host"'
check_contains "ServiceEndpoint has Port field" "$PROJECT_ROOT/internal/config/config.go" 'Port.*string.*yaml:"port"'
check_contains "ServiceEndpoint has Enabled field" "$PROJECT_ROOT/internal/config/config.go" 'Enabled.*bool.*yaml:"enabled"'
check_contains "ServiceEndpoint has Required field" "$PROJECT_ROOT/internal/config/config.go" 'Required.*bool.*yaml:"required"'
check_contains "ServiceEndpoint has Remote field" "$PROJECT_ROOT/internal/config/config.go" 'Remote.*bool.*yaml:"remote"'
check_contains "ServiceEndpoint has HealthType field" "$PROJECT_ROOT/internal/config/config.go" 'HealthType.*string.*yaml:"health_type"'
check_contains "ServiceEndpoint has HealthPath field" "$PROJECT_ROOT/internal/config/config.go" 'HealthPath.*string.*yaml:"health_path"'
check_contains "ServiceEndpoint has Timeout field" "$PROJECT_ROOT/internal/config/config.go" 'Timeout.*time.Duration.*yaml:"timeout"'
check_contains "ServiceEndpoint has RetryCount field" "$PROJECT_ROOT/internal/config/config.go" 'RetryCount.*int.*yaml:"retry_count"'
check_contains "ServiceEndpoint has ComposeFile field" "$PROJECT_ROOT/internal/config/config.go" 'ComposeFile.*string.*yaml:"compose_file"'
check_contains "ServiceEndpoint has ServiceName field" "$PROJECT_ROOT/internal/config/config.go" 'ServiceName.*string.*yaml:"service_name"'

# ============================================================================
# SECTION 2: ServicesConfig Has All Services
# ============================================================================
info "Section 2: ServicesConfig Has All Services"

check_contains "PostgreSQL in ServicesConfig" "$PROJECT_ROOT/internal/config/config.go" 'PostgreSQL.*ServiceEndpoint.*yaml:"postgresql"'
check_contains "Redis in ServicesConfig" "$PROJECT_ROOT/internal/config/config.go" 'Redis.*ServiceEndpoint.*yaml:"redis"'
check_contains "Cognee in ServicesConfig" "$PROJECT_ROOT/internal/config/config.go" 'Cognee.*ServiceEndpoint.*yaml:"cognee"'
check_contains "ChromaDB in ServicesConfig" "$PROJECT_ROOT/internal/config/config.go" 'ChromaDB.*ServiceEndpoint.*yaml:"chromadb"'
check_contains "Prometheus in ServicesConfig" "$PROJECT_ROOT/internal/config/config.go" 'Prometheus.*ServiceEndpoint.*yaml:"prometheus"'
check_contains "Grafana in ServicesConfig" "$PROJECT_ROOT/internal/config/config.go" 'Grafana.*ServiceEndpoint.*yaml:"grafana"'
check_contains "Qdrant in ServicesConfig" "$PROJECT_ROOT/internal/config/config.go" 'Qdrant.*ServiceEndpoint.*yaml:"qdrant"'
check_contains "Weaviate in ServicesConfig" "$PROJECT_ROOT/internal/config/config.go" 'Weaviate.*ServiceEndpoint.*yaml:"weaviate"'
check_contains "MCPServers in ServicesConfig" "$PROJECT_ROOT/internal/config/config.go" 'MCPServers.*map\[string\]ServiceEndpoint'

# ============================================================================
# SECTION 3: Default Config and Helper Functions
# ============================================================================
info "Section 3: Default Config and Helper Functions"

check_contains "DefaultServicesConfig function exists" "$PROJECT_ROOT/internal/config/config.go" "func DefaultServicesConfig"
check_contains "LoadServicesFromEnv function exists" "$PROJECT_ROOT/internal/config/config.go" "func LoadServicesFromEnv"
check_contains "AllEndpoints method exists" "$PROJECT_ROOT/internal/config/config.go" "func.*ServicesConfig.*AllEndpoints"
check_contains "RequiredEndpoints method exists" "$PROJECT_ROOT/internal/config/config.go" "func.*ServicesConfig.*RequiredEndpoints"
check_contains "ResolvedURL method exists" "$PROJECT_ROOT/internal/config/config.go" "func.*ServiceEndpoint.*ResolvedURL"

# ============================================================================
# SECTION 4: Boot Manager
# ============================================================================
info "Section 4: Boot Manager"

check_file "Boot manager file exists" "$PROJECT_ROOT/internal/services/boot_manager.go"
check_contains "BootManager struct defined" "$PROJECT_ROOT/internal/services/boot_manager.go" "type BootManager struct"
check_contains "BootResult struct defined" "$PROJECT_ROOT/internal/services/boot_manager.go" "type BootResult struct"
check_contains "NewBootManager function" "$PROJECT_ROOT/internal/services/boot_manager.go" "func NewBootManager"
check_contains "BootAll method" "$PROJECT_ROOT/internal/services/boot_manager.go" "func.*BootManager.*BootAll"
check_contains "ShutdownAll method" "$PROJECT_ROOT/internal/services/boot_manager.go" "func.*BootManager.*ShutdownAll"
check_contains "HealthCheckAll method" "$PROJECT_ROOT/internal/services/boot_manager.go" "func.*BootManager.*HealthCheckAll"

# ============================================================================
# SECTION 5: Health Checker
# ============================================================================
info "Section 5: Health Checker"

check_file "Health checker file exists" "$PROJECT_ROOT/internal/services/health_checker.go"
check_contains "HealthChecker struct defined" "$PROJECT_ROOT/internal/services/health_checker.go" "type ServiceHealthChecker struct"
check_contains "Check method" "$PROJECT_ROOT/internal/services/health_checker.go" "func.*HealthChecker.*Check"
check_contains "CheckWithRetry method" "$PROJECT_ROOT/internal/services/health_checker.go" "func.*HealthChecker.*CheckWithRetry"

# ============================================================================
# SECTION 6: Main.go Integration
# ============================================================================
info "Section 6: Main.go Integration"

check_contains "BootManager used in main" "$PROJECT_ROOT/cmd/helixagent/main.go" "services.NewBootManager"
check_contains "BootAll called" "$PROJECT_ROOT/cmd/helixagent/main.go" "bootMgr.BootAll"
check_contains "ShutdownAll in shutdown" "$PROJECT_ROOT/cmd/helixagent/main.go" "bootMgr.ShutdownAll"

# ============================================================================
# SECTION 7: YAML Configuration
# ============================================================================
info "Section 7: YAML Configuration"

check_contains "Development YAML has services section" "$PROJECT_ROOT/configs/development.yaml" "^services:"
check_contains "Development YAML has postgresql" "$PROJECT_ROOT/configs/development.yaml" "postgresql:"
check_contains "Development YAML has redis" "$PROJECT_ROOT/configs/development.yaml" "redis:"
check_contains "Development YAML has cognee" "$PROJECT_ROOT/configs/development.yaml" "cognee:"
check_contains "Development YAML has chromadb" "$PROJECT_ROOT/configs/development.yaml" "chromadb:"
check_file "Remote services example exists" "$PROJECT_ROOT/configs/remote-services-example.yaml"
check_contains "Remote example has remote: true" "$PROJECT_ROOT/configs/remote-services-example.yaml" "remote: true"

# ============================================================================
# SECTION 8: Unit Tests
# ============================================================================
info "Section 8: Unit Tests"

check_file "Services config tests exist" "$PROJECT_ROOT/internal/config/services_test.go"
check_file "Boot manager tests exist" "$PROJECT_ROOT/internal/services/boot_manager_test.go"
check_file "Health checker tests exist" "$PROJECT_ROOT/internal/services/health_checker_test.go"

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
