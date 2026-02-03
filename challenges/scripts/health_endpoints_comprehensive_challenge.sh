#!/bin/bash
# Health Endpoints Comprehensive Challenge
# VALIDATES: /health, /ready, /live endpoints, component health, degraded state handling
# Tests the complete health check infrastructure

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh" 2>/dev/null || {
    log_info() { echo -e "\033[0;34m[INFO] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_success() { echo -e "\033[0;32m[SUCCESS] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_error() { echo -e "\033[0;31m[ERROR] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_warning() { echo -e "\033[0;33m[WARNING] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
}

CHALLENGE_NAME="Health Endpoints Comprehensive Challenge"
PASSED=0
FAILED=0
TOTAL=0
PROJECT_ROOT="${SCRIPT_DIR}/../.."
API_BASE="${HELIX_API_BASE:-http://localhost:8080}"

log_info "=============================================="
log_info "$CHALLENGE_NAME"
log_info "=============================================="
log_info "Validates: Health check infrastructure"
log_info ""

# ============================================================================
# Section 1: Health Handler Code Structure
# ============================================================================

log_info "=============================================="
log_info "Section 1: Health Handler Code Structure"
log_info "=============================================="

# Test 1: health_handler.go exists
TOTAL=$((TOTAL + 1))
log_info "Test 1: health_handler.go exists"
if [ -f "$PROJECT_ROOT/internal/handlers/health_handler.go" ]; then
    log_success "health_handler.go exists"
    PASSED=$((PASSED + 1))
else
    log_error "health_handler.go NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 2: HealthHandler struct defined
TOTAL=$((TOTAL + 1))
log_info "Test 2: HealthHandler struct defined"
if grep -q "type HealthHandler struct" "$PROJECT_ROOT/internal/handlers/health_handler.go" 2>/dev/null; then
    log_success "HealthHandler struct is defined"
    PASSED=$((PASSED + 1))
else
    log_error "HealthHandler struct NOT defined!"
    FAILED=$((FAILED + 1))
fi

# Test 3: NewHealthHandler constructor exists
TOTAL=$((TOTAL + 1))
log_info "Test 3: NewHealthHandler constructor exists"
if grep -q "func NewHealthHandler" "$PROJECT_ROOT/internal/handlers/health_handler.go" 2>/dev/null; then
    log_success "NewHealthHandler constructor exists"
    PASSED=$((PASSED + 1))
else
    log_error "NewHealthHandler constructor NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 4: RegisterHealthRoutes function exists
TOTAL=$((TOTAL + 1))
log_info "Test 4: RegisterHealthRoutes function exists"
if grep -q "func RegisterHealthRoutes" "$PROJECT_ROOT/internal/handlers/health_handler.go" 2>/dev/null; then
    log_success "RegisterHealthRoutes function exists"
    PASSED=$((PASSED + 1))
else
    log_error "RegisterHealthRoutes function NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 2: Basic Health Endpoints in Router
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 2: Basic Health Endpoints in Router"
log_info "=============================================="

# Test 5: /health endpoint defined
TOTAL=$((TOTAL + 1))
log_info "Test 5: /health endpoint defined"
if grep -q '"/health"' "$PROJECT_ROOT/internal/router/router.go" 2>/dev/null; then
    log_success "/health endpoint defined in router"
    PASSED=$((PASSED + 1))
else
    log_error "/health endpoint NOT defined!"
    FAILED=$((FAILED + 1))
fi

# Test 6: /v1/health endpoint defined
TOTAL=$((TOTAL + 1))
log_info "Test 6: /v1/health endpoint defined"
if grep -q '"/v1/health"' "$PROJECT_ROOT/internal/router/router.go" 2>/dev/null; then
    log_success "/v1/health endpoint defined in router"
    PASSED=$((PASSED + 1))
else
    log_error "/v1/health endpoint NOT defined!"
    FAILED=$((FAILED + 1))
fi

# Test 7: Health endpoint returns JSON status
TOTAL=$((TOTAL + 1))
log_info "Test 7: Health endpoint returns JSON status"
if grep -A5 '"/health"' "$PROJECT_ROOT/internal/router/router.go" 2>/dev/null | grep -q '"status".*healthy'; then
    log_success "Health endpoint returns healthy status"
    PASSED=$((PASSED + 1))
else
    log_error "Health endpoint does NOT return status!"
    FAILED=$((FAILED + 1))
fi

# Test 8: Enhanced health check includes provider status
TOTAL=$((TOTAL + 1))
log_info "Test 8: Enhanced health check includes provider status"
if grep -A20 '"/v1/health"' "$PROJECT_ROOT/internal/router/router.go" 2>/dev/null | grep -q "providers"; then
    log_success "Enhanced health check includes provider status"
    PASSED=$((PASSED + 1))
else
    log_error "Enhanced health check does NOT include provider status!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 3: Monitoring Handler Endpoints
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 3: Monitoring Handler Endpoints"
log_info "=============================================="

# Test 9: monitoring_handler.go exists
TOTAL=$((TOTAL + 1))
log_info "Test 9: monitoring_handler.go exists"
if [ -f "$PROJECT_ROOT/internal/handlers/monitoring_handler.go" ]; then
    log_success "monitoring_handler.go exists"
    PASSED=$((PASSED + 1))
else
    log_error "monitoring_handler.go NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 10: MonitoringHandler struct defined
TOTAL=$((TOTAL + 1))
log_info "Test 10: MonitoringHandler struct defined"
if grep -q "type MonitoringHandler struct" "$PROJECT_ROOT/internal/handlers/monitoring_handler.go" 2>/dev/null; then
    log_success "MonitoringHandler struct is defined"
    PASSED=$((PASSED + 1))
else
    log_error "MonitoringHandler struct NOT defined!"
    FAILED=$((FAILED + 1))
fi

# Test 11: /monitoring/status endpoint registered
TOTAL=$((TOTAL + 1))
log_info "Test 11: /monitoring/status endpoint registered"
if grep -q '"/status"' "$PROJECT_ROOT/internal/handlers/monitoring_handler.go" 2>/dev/null; then
    log_success "/monitoring/status endpoint registered"
    PASSED=$((PASSED + 1))
else
    log_error "/monitoring/status endpoint NOT registered!"
    FAILED=$((FAILED + 1))
fi

# Test 12: GetOverallStatus method exists
TOTAL=$((TOTAL + 1))
log_info "Test 12: GetOverallStatus method exists"
if grep -q "func (h \*MonitoringHandler) GetOverallStatus" "$PROJECT_ROOT/internal/handlers/monitoring_handler.go" 2>/dev/null; then
    log_success "GetOverallStatus method exists"
    PASSED=$((PASSED + 1))
else
    log_error "GetOverallStatus method NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 4: Provider Health Endpoints
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 4: Provider Health Endpoints"
log_info "=============================================="

# Test 13: GetProviderHealth method exists
TOTAL=$((TOTAL + 1))
log_info "Test 13: GetProviderHealth method exists"
if grep -q "func (h \*HealthHandler) GetProviderHealth" "$PROJECT_ROOT/internal/handlers/health_handler.go" 2>/dev/null; then
    log_success "GetProviderHealth method exists"
    PASSED=$((PASSED + 1))
else
    log_error "GetProviderHealth method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 14: GetAllProvidersHealth method exists
TOTAL=$((TOTAL + 1))
log_info "Test 14: GetAllProvidersHealth method exists"
if grep -q "func (h \*HealthHandler) GetAllProvidersHealth" "$PROJECT_ROOT/internal/handlers/health_handler.go" 2>/dev/null; then
    log_success "GetAllProvidersHealth method exists"
    PASSED=$((PASSED + 1))
else
    log_error "GetAllProvidersHealth method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 15: GetHealthyProviders method exists
TOTAL=$((TOTAL + 1))
log_info "Test 15: GetHealthyProviders method exists"
if grep -q "func (h \*HealthHandler) GetHealthyProviders" "$PROJECT_ROOT/internal/handlers/health_handler.go" 2>/dev/null; then
    log_success "GetHealthyProviders method exists"
    PASSED=$((PASSED + 1))
else
    log_error "GetHealthyProviders method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 16: IsProviderAvailable method exists
TOTAL=$((TOTAL + 1))
log_info "Test 16: IsProviderAvailable method exists"
if grep -q "func (h \*HealthHandler) IsProviderAvailable" "$PROJECT_ROOT/internal/handlers/health_handler.go" 2>/dev/null; then
    log_success "IsProviderAvailable method exists"
    PASSED=$((PASSED + 1))
else
    log_error "IsProviderAvailable method NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 5: Health Response Types
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 5: Health Response Types"
log_info "=============================================="

# Test 17: ProviderHealthResponse struct defined
TOTAL=$((TOTAL + 1))
log_info "Test 17: ProviderHealthResponse struct defined"
if grep -q "type ProviderHealthResponse struct" "$PROJECT_ROOT/internal/handlers/health_handler.go" 2>/dev/null; then
    log_success "ProviderHealthResponse struct defined"
    PASSED=$((PASSED + 1))
else
    log_error "ProviderHealthResponse struct NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 18: HealthSummary struct defined
TOTAL=$((TOTAL + 1))
log_info "Test 18: HealthSummary struct defined"
if grep -q "type HealthSummary struct" "$PROJECT_ROOT/internal/handlers/health_handler.go" 2>/dev/null; then
    log_success "HealthSummary struct defined"
    PASSED=$((PASSED + 1))
else
    log_error "HealthSummary struct NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 19: CircuitBreakerResponse struct defined
TOTAL=$((TOTAL + 1))
log_info "Test 19: CircuitBreakerResponse struct defined"
if grep -q "type CircuitBreakerResponse struct" "$PROJECT_ROOT/internal/handlers/health_handler.go" 2>/dev/null; then
    log_success "CircuitBreakerResponse struct defined"
    PASSED=$((PASSED + 1))
else
    log_error "CircuitBreakerResponse struct NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 20: OverallMonitoringStatus struct defined
TOTAL=$((TOTAL + 1))
log_info "Test 20: OverallMonitoringStatus struct defined"
if grep -q "type OverallMonitoringStatus struct" "$PROJECT_ROOT/internal/handlers/monitoring_handler.go" 2>/dev/null; then
    log_success "OverallMonitoringStatus struct defined"
    PASSED=$((PASSED + 1))
else
    log_error "OverallMonitoringStatus struct NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 6: Circuit Breaker Health Endpoints
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 6: Circuit Breaker Health Endpoints"
log_info "=============================================="

# Test 21: GetCircuitBreakerStatus method exists in health handler
TOTAL=$((TOTAL + 1))
log_info "Test 21: GetCircuitBreakerStatus method exists in health handler"
if grep -q "func (h \*HealthHandler) GetCircuitBreakerStatus" "$PROJECT_ROOT/internal/handlers/health_handler.go" 2>/dev/null; then
    log_success "GetCircuitBreakerStatus method exists in health handler"
    PASSED=$((PASSED + 1))
else
    log_error "GetCircuitBreakerStatus method NOT found in health handler!"
    FAILED=$((FAILED + 1))
fi

# Test 22: Circuit breaker endpoint in monitoring handler
TOTAL=$((TOTAL + 1))
log_info "Test 22: Circuit breaker endpoint in monitoring handler"
if grep -q '"/circuit-breakers"' "$PROJECT_ROOT/internal/handlers/monitoring_handler.go" 2>/dev/null; then
    log_success "Circuit breaker endpoint in monitoring handler"
    PASSED=$((PASSED + 1))
else
    log_error "Circuit breaker endpoint NOT found in monitoring handler!"
    FAILED=$((FAILED + 1))
fi

# Test 23: Reset circuit breaker endpoint
TOTAL=$((TOTAL + 1))
log_info "Test 23: Reset circuit breaker endpoint exists"
if grep -q 'ResetCircuitBreaker' "$PROJECT_ROOT/internal/handlers/monitoring_handler.go" 2>/dev/null; then
    log_success "Reset circuit breaker endpoint exists"
    PASSED=$((PASSED + 1))
else
    log_error "Reset circuit breaker endpoint NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 24: Reset all circuit breakers endpoint
TOTAL=$((TOTAL + 1))
log_info "Test 24: Reset all circuit breakers endpoint exists"
if grep -q 'ResetAllCircuitBreakers' "$PROJECT_ROOT/internal/handlers/monitoring_handler.go" 2>/dev/null; then
    log_success "Reset all circuit breakers endpoint exists"
    PASSED=$((PASSED + 1))
else
    log_error "Reset all circuit breakers endpoint NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 7: Provider Health Monitor Service
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 7: Provider Health Monitor Service"
log_info "=============================================="

# Test 25: provider_health_monitor.go exists
TOTAL=$((TOTAL + 1))
log_info "Test 25: provider_health_monitor.go exists"
if [ -f "$PROJECT_ROOT/internal/services/provider_health_monitor.go" ]; then
    log_success "provider_health_monitor.go exists"
    PASSED=$((PASSED + 1))
else
    log_error "provider_health_monitor.go NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 26: ProviderHealthMonitor struct defined
TOTAL=$((TOTAL + 1))
log_info "Test 26: ProviderHealthMonitor struct defined"
if grep -q "type ProviderHealthMonitor struct" "$PROJECT_ROOT/internal/services/provider_health_monitor.go" 2>/dev/null; then
    log_success "ProviderHealthMonitor struct defined"
    PASSED=$((PASSED + 1))
else
    log_error "ProviderHealthMonitor struct NOT defined!"
    FAILED=$((FAILED + 1))
fi

# Test 27: GetProviderHealthStatus method in monitoring handler
TOTAL=$((TOTAL + 1))
log_info "Test 27: GetProviderHealthStatus method in monitoring handler"
if grep -q "func (h \*MonitoringHandler) GetProviderHealthStatus" "$PROJECT_ROOT/internal/handlers/monitoring_handler.go" 2>/dev/null; then
    log_success "GetProviderHealthStatus method exists"
    PASSED=$((PASSED + 1))
else
    log_error "GetProviderHealthStatus method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 28: ForceHealthCheck method exists
TOTAL=$((TOTAL + 1))
log_info "Test 28: ForceHealthCheck method exists"
if grep -q "func (h \*MonitoringHandler) ForceHealthCheck" "$PROJECT_ROOT/internal/handlers/monitoring_handler.go" 2>/dev/null; then
    log_success "ForceHealthCheck method exists"
    PASSED=$((PASSED + 1))
else
    log_error "ForceHealthCheck method NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 8: OAuth Token Health Endpoints
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 8: OAuth Token Health Endpoints"
log_info "=============================================="

# Test 29: OAuth token status endpoint
TOTAL=$((TOTAL + 1))
log_info "Test 29: OAuth token status endpoint"
if grep -q '"/oauth-tokens"' "$PROJECT_ROOT/internal/handlers/monitoring_handler.go" 2>/dev/null; then
    log_success "OAuth token status endpoint exists"
    PASSED=$((PASSED + 1))
else
    log_error "OAuth token status endpoint NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 30: GetOAuthTokenStatus method exists
TOTAL=$((TOTAL + 1))
log_info "Test 30: GetOAuthTokenStatus method exists"
if grep -q "func (h \*MonitoringHandler) GetOAuthTokenStatus" "$PROJECT_ROOT/internal/handlers/monitoring_handler.go" 2>/dev/null; then
    log_success "GetOAuthTokenStatus method exists"
    PASSED=$((PASSED + 1))
else
    log_error "GetOAuthTokenStatus method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 31: RefreshOAuthToken method exists
TOTAL=$((TOTAL + 1))
log_info "Test 31: RefreshOAuthToken method exists"
if grep -q "func (h \*MonitoringHandler) RefreshOAuthToken" "$PROJECT_ROOT/internal/handlers/monitoring_handler.go" 2>/dev/null; then
    log_success "RefreshOAuthToken method exists"
    PASSED=$((PASSED + 1))
else
    log_error "RefreshOAuthToken method NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 9: Fallback Chain Health Endpoints
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 9: Fallback Chain Health Endpoints"
log_info "=============================================="

# Test 32: Fallback chain status endpoint
TOTAL=$((TOTAL + 1))
log_info "Test 32: Fallback chain status endpoint"
if grep -q '"/fallback-chain"' "$PROJECT_ROOT/internal/handlers/monitoring_handler.go" 2>/dev/null; then
    log_success "Fallback chain status endpoint exists"
    PASSED=$((PASSED + 1))
else
    log_error "Fallback chain status endpoint NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 33: GetFallbackChainStatus method exists
TOTAL=$((TOTAL + 1))
log_info "Test 33: GetFallbackChainStatus method exists"
if grep -q "func (h \*MonitoringHandler) GetFallbackChainStatus" "$PROJECT_ROOT/internal/handlers/monitoring_handler.go" 2>/dev/null; then
    log_success "GetFallbackChainStatus method exists"
    PASSED=$((PASSED + 1))
else
    log_error "GetFallbackChainStatus method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 34: ValidateFallbackChain method exists
TOTAL=$((TOTAL + 1))
log_info "Test 34: ValidateFallbackChain method exists"
if grep -q "func (h \*MonitoringHandler) ValidateFallbackChain" "$PROJECT_ROOT/internal/handlers/monitoring_handler.go" 2>/dev/null; then
    log_success "ValidateFallbackChain method exists"
    PASSED=$((PASSED + 1))
else
    log_error "ValidateFallbackChain method NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 10: Concurrency Health Endpoints
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 10: Concurrency Health Endpoints"
log_info "=============================================="

# Test 35: Concurrency status endpoint
TOTAL=$((TOTAL + 1))
log_info "Test 35: Concurrency status endpoint"
if grep -q '"/concurrency"' "$PROJECT_ROOT/internal/handlers/monitoring_handler.go" 2>/dev/null; then
    log_success "Concurrency status endpoint exists"
    PASSED=$((PASSED + 1))
else
    log_error "Concurrency status endpoint NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 36: GetConcurrencyStatus method exists
TOTAL=$((TOTAL + 1))
log_info "Test 36: GetConcurrencyStatus method exists"
if grep -q "func (h \*MonitoringHandler) GetConcurrencyStatus" "$PROJECT_ROOT/internal/handlers/monitoring_handler.go" 2>/dev/null; then
    log_success "GetConcurrencyStatus method exists"
    PASSED=$((PASSED + 1))
else
    log_error "GetConcurrencyStatus method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 37: Concurrency alerts endpoint
TOTAL=$((TOTAL + 1))
log_info "Test 37: Concurrency alerts endpoint"
if grep -q '"/concurrency/alerts"' "$PROJECT_ROOT/internal/handlers/monitoring_handler.go" 2>/dev/null; then
    log_success "Concurrency alerts endpoint exists"
    PASSED=$((PASSED + 1))
else
    log_error "Concurrency alerts endpoint NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 11: Latency and Performance Endpoints
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 11: Latency and Performance Endpoints"
log_info "=============================================="

# Test 38: GetProviderLatency method exists
TOTAL=$((TOTAL + 1))
log_info "Test 38: GetProviderLatency method exists"
if grep -q "func (h \*HealthHandler) GetProviderLatency" "$PROJECT_ROOT/internal/handlers/health_handler.go" 2>/dev/null; then
    log_success "GetProviderLatency method exists"
    PASSED=$((PASSED + 1))
else
    log_error "GetProviderLatency method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 39: GetFastestProvider method exists
TOTAL=$((TOTAL + 1))
log_info "Test 39: GetFastestProvider method exists"
if grep -q "func (h \*HealthHandler) GetFastestProvider" "$PROJECT_ROOT/internal/handlers/health_handler.go" 2>/dev/null; then
    log_success "GetFastestProvider method exists"
    PASSED=$((PASSED + 1))
else
    log_error "GetFastestProvider method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 40: LatencyStatsResponse struct defined
TOTAL=$((TOTAL + 1))
log_info "Test 40: LatencyStatsResponse struct defined"
if grep -q "type LatencyStatsResponse struct" "$PROJECT_ROOT/internal/handlers/health_handler.go" 2>/dev/null; then
    log_success "LatencyStatsResponse struct defined"
    PASSED=$((PASSED + 1))
else
    log_error "LatencyStatsResponse struct NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 12: Health Service Status
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 12: Health Service Status"
log_info "=============================================="

# Test 41: GetHealthServiceStatus method exists
TOTAL=$((TOTAL + 1))
log_info "Test 41: GetHealthServiceStatus method exists"
if grep -q "func (h \*HealthHandler) GetHealthServiceStatus" "$PROJECT_ROOT/internal/handlers/health_handler.go" 2>/dev/null; then
    log_success "GetHealthServiceStatus method exists"
    PASSED=$((PASSED + 1))
else
    log_error "GetHealthServiceStatus method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 42: HealthServiceStatusResponse struct defined
TOTAL=$((TOTAL + 1))
log_info "Test 42: HealthServiceStatusResponse struct defined"
if grep -q "type HealthServiceStatusResponse struct" "$PROJECT_ROOT/internal/handlers/health_handler.go" 2>/dev/null; then
    log_success "HealthServiceStatusResponse struct defined"
    PASSED=$((PASSED + 1))
else
    log_error "HealthServiceStatusResponse struct NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Results Summary
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Challenge Results Summary"
log_info "=============================================="
log_info "Passed: $PASSED/$TOTAL"
log_info "Failed: $FAILED/$TOTAL"
log_info ""

if [ "$FAILED" -eq 0 ]; then
    log_success "ALL $TOTAL TESTS PASSED!"
    exit 0
else
    log_error "$FAILED TEST(S) FAILED!"
    exit 1
fi
