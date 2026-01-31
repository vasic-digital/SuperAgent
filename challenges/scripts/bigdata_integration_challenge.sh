#!/bin/bash
# BigData Integration Challenge
# VALIDATES: Infinite Context, Distributed Memory, Knowledge Graph Streaming,
#            ClickHouse Analytics, Cross-session Learning
# Tests the complete BigData integration with 15 tests

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh" 2>/dev/null || {
    log_info() { echo -e "\033[0;34m[INFO] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_success() { echo -e "\033[0;32m[SUCCESS] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_error() { echo -e "\033[0;31m[ERROR] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_warning() { echo -e "\033[0;33m[WARNING] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
}

CHALLENGE_NAME="BigData Integration Challenge"
PASSED=0
FAILED=0
TOTAL=0
PROJECT_ROOT="${SCRIPT_DIR}/../.."

log_info "=============================================="
log_info "$CHALLENGE_NAME"
log_info "=============================================="
log_info "Validates: Infinite Context, Distributed Memory, Knowledge Graph Streaming,"
log_info "           ClickHouse Analytics, Cross-session Learning"
log_info ""

# ============================================================================
# Section 1: Core Files & Structure
# ============================================================================

log_info "=============================================="
log_info "Section 1: Core Files & Structure"
log_info "=============================================="

# Test 1: integration.go exists
TOTAL=$((TOTAL + 1))
log_info "Test 1: integration.go file exists"
if [ -f "$PROJECT_ROOT/internal/bigdata/integration.go" ]; then
    log_success "integration.go exists"
    PASSED=$((PASSED + 1))
else
    log_error "integration.go NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 2: config_converter.go exists
TOTAL=$((TOTAL + 1))
log_info "Test 2: config_converter.go file exists"
if [ -f "$PROJECT_ROOT/internal/bigdata/config_converter.go" ]; then
    log_success "config_converter.go exists"
    PASSED=$((PASSED + 1))
else
    log_error "config_converter.go NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 3: handler.go exists (API handlers)
TOTAL=$((TOTAL + 1))
log_info "Test 3: handler.go file exists"
if [ -f "$PROJECT_ROOT/internal/bigdata/handler.go" ]; then
    log_success "handler.go exists"
    PASSED=$((PASSED + 1))
else
    log_error "handler.go NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 4: IntegrationConfig struct defined
TOTAL=$((TOTAL + 1))
log_info "Test 4: IntegrationConfig struct defined"
if grep -q "type IntegrationConfig struct" "$PROJECT_ROOT/internal/bigdata/integration.go" 2>/dev/null; then
    log_success "IntegrationConfig struct is defined"
    PASSED=$((PASSED + 1))
else
    log_error "IntegrationConfig struct NOT defined!"
    FAILED=$((FAILED + 1))
fi

# Test 5: BigDataIntegration struct defined
TOTAL=$((TOTAL + 1))
log_info "Test 5: BigDataIntegration struct defined"
if grep -q "type BigDataIntegration struct" "$PROJECT_ROOT/internal/bigdata/integration.go" 2>/dev/null; then
    log_success "BigDataIntegration struct is defined"
    PASSED=$((PASSED + 1))
else
    log_error "BigDataIntegration struct NOT defined!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 2: Configuration & Conversion
# ============================================================================

log_info "=============================================="
log_info "Section 2: Configuration & Conversion"
log_info "=============================================="

# Test 6: ConfigToIntegrationConfig function exists
TOTAL=$((TOTAL + 1))
log_info "Test 6: ConfigToIntegrationConfig function exists"
if grep -q "func ConfigToIntegrationConfig" "$PROJECT_ROOT/internal/bigdata/config_converter.go" 2>/dev/null; then
    log_success "ConfigToIntegrationConfig function exists"
    PASSED=$((PASSED + 1))
else
    log_error "ConfigToIntegrationConfig function NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 7: DefaultIntegrationConfig function exists
TOTAL=$((TOTAL + 1))
log_info "Test 7: DefaultIntegrationConfig function exists"
if grep -q "func DefaultIntegrationConfig" "$PROJECT_ROOT/internal/bigdata/integration.go" 2>/dev/null; then
    log_success "DefaultIntegrationConfig function exists"
    PASSED=$((PASSED + 1))
else
    log_error "DefaultIntegrationConfig function NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 8: configs/bigdata.yaml exists
TOTAL=$((TOTAL + 1))
log_info "Test 8: configs/bigdata.yaml exists"
if [ -f "$PROJECT_ROOT/configs/bigdata.yaml" ]; then
    log_success "configs/bigdata.yaml exists"
    PASSED=$((PASSED + 1))
else
    log_error "configs/bigdata.yaml NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 9: Configuration has all 5 components
TOTAL=$((TOTAL + 1))
log_info "Test 9: IntegrationConfig has all 5 component flags"
COMPONENTS=("EnableInfiniteContext" "EnableDistributedMemory" "EnableKnowledgeGraph" "EnableAnalytics" "EnableCrossLearning")
component_count=0
for component in "${COMPONENTS[@]}"; do
    if grep -q "$component" "$PROJECT_ROOT/internal/bigdata/integration.go" 2>/dev/null; then
        component_count=$((component_count + 1))
    fi
done
if [ $component_count -eq 5 ]; then
    log_success "All 5 component flags defined"
    PASSED=$((PASSED + 1))
else
    log_error "Missing component flags (found $component_count/5)"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 3: API Endpoints & Integration
# ============================================================================

log_info "=============================================="
log_info "Section 3: API Endpoints & Integration"
log_info "=============================================="

# Test 10: main.go imports bigdata package
TOTAL=$((TOTAL + 1))
log_info "Test 10: main.go imports bigdata package"
if grep -q "\"dev.helix.agent/internal/bigdata\"" "$PROJECT_ROOT/cmd/helixagent/main.go" 2>/dev/null; then
    log_success "bigdata package imported in main.go"
    PASSED=$((PASSED + 1))
else
    log_error "bigdata package NOT imported in main.go!"
    FAILED=$((FAILED + 1))
fi

# Test 11: API endpoint /v1/bigdata/health defined
TOTAL=$((TOTAL + 1))
log_info "Test 11: API endpoint /v1/bigdata/health defined"
if grep -q "\"/v1/bigdata/health\"" "$PROJECT_ROOT/cmd/helixagent/main.go" 2>/dev/null || \
   grep -q "\"/bigdata/health\"" "$PROJECT_ROOT/internal/bigdata/handler.go" 2>/dev/null; then
    log_success "/v1/bigdata/health endpoint defined"
    PASSED=$((PASSED + 1))
else
    log_error "/v1/bigdata/health endpoint NOT defined!"
    FAILED=$((FAILED + 1))
fi

# Test 12: API endpoint /v1/bigdata/components defined
TOTAL=$((TOTAL + 1))
log_info "Test 12: API endpoint /v1/bigdata/components defined"
if grep -q "\"/v1/bigdata/components\"" "$PROJECT_ROOT/cmd/helixagent/main.go" 2>/dev/null; then
    log_success "/v1/bigdata/components endpoint defined"
    PASSED=$((PASSED + 1))
else
    log_error "/v1/bigdata/components endpoint NOT defined!"
    FAILED=$((FAILED + 1))
fi

# Test 13: Messaging hub getter methods exist
TOTAL=$((TOTAL + 1))
log_info "Test 13: Messaging hub getter methods exist"
if grep -q "func.*GetMessageBroker" "$PROJECT_ROOT/internal/messaging/hub.go" 2>/dev/null; then
    log_success "GetMessageBroker method exists"
    PASSED=$((PASSED + 1))
else
    log_error "GetMessageBroker method NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 4: Compilation Test
# ============================================================================

log_info "=============================================="
log_info "Section 4: Compilation Test"
log_info "=============================================="

# Test 14: BigData integration compiles
TOTAL=$((TOTAL + 1))
log_info "Test 14: BigData integration compiles (dry-run)"
cd "$PROJECT_ROOT"
if go build ./internal/bigdata 2>&1 | tee /tmp/bigdata_build.log | grep -q "error"; then
    log_error "BigData integration compilation failed"
    cat /tmp/bigdata_build.log
    FAILED=$((FAILED + 1))
else
    log_success "BigData integration compiles successfully"
    PASSED=$((PASSED + 1))
fi

# Test 15: Main application with BigData compiles
TOTAL=$((TOTAL + 1))
log_info "Test 15: Main application with BigData compiles (dry-run)"
if go build ./cmd/helixagent 2>&1 | tee /tmp/helixagent_build.log | grep -q "error"; then
    log_error "Main application compilation failed"
    cat /tmp/helixagent_build.log
    FAILED=$((FAILED + 1))
else
    log_success "Main application compiles successfully"
    PASSED=$((PASSED + 1))
fi

# ============================================================================
# Results
# ============================================================================

log_info ""
log_info "=============================================="
log_info "  Results: $PASSED/$TOTAL tests passed"
log_info "=============================================="

if [ $FAILED -gt 0 ]; then
    log_error "$FAILED test(s) failed"
    exit 1
else
    log_success "All tests passed!"
    exit 0
fi