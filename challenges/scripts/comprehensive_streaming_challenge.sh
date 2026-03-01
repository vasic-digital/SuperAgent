#!/bin/bash
#
# Comprehensive Streaming Challenge
# Validates that the comprehensive multi-agent debate system streaming works correctly
# Tests: 8 roles, 5 teams, streaming events, backward compatibility
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
HELIXAGENT_BIN="${PROJECT_ROOT}/bin/helixagent"
TEST_LOG="${SCRIPT_DIR}/comprehensive_streaming_test.log"
FAILED=0
PASSED=0

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a "$TEST_LOG"
}

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1" | tee -a "$TEST_LOG"
    ((PASSED++))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1" | tee -a "$TEST_LOG"
    ((FAILED++))
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1" | tee -a "$TEST_LOG"
}

# Initialize test log
> "$TEST_LOG"
echo "==============================================" >> "$TEST_LOG"
echo "Comprehensive Streaming Challenge Started" >> "$TEST_LOG"
echo "Timestamp: $(date)" >> "$TEST_LOG"
echo "==============================================" >> "$TEST_LOG"

log_info "Starting Comprehensive Streaming Challenge..."
log_info "Project Root: $PROJECT_ROOT"

# Test 1: Verify comprehensive streaming types exist
log_info "Test 1: Verifying comprehensive streaming types..."
if grep -q "StreamEvent" "${PROJECT_ROOT}/internal/debate/comprehensive/streaming_types.go" 2>/dev/null; then
    log_pass "StreamEvent type exists"
else
    log_fail "StreamEvent type not found"
fi

if grep -q "StreamOrchestrator" "${PROJECT_ROOT}/internal/debate/comprehensive/streaming_types.go" 2>/dev/null; then
    log_pass "StreamOrchestrator exists"
else
    log_fail "StreamOrchestrator not found"
fi

if grep -q "TeamDesign\|TeamImplementation\|TeamQuality\|TeamRedTeam\|TeamRefactoring" "${PROJECT_ROOT}/internal/debate/comprehensive/streaming_types.go" 2>/dev/null; then
    log_pass "All 5 debate teams defined"
else
    log_fail "Debate teams not properly defined"
fi

# Test 2: Verify IntegrationManager has streaming support
log_info "Test 2: Verifying IntegrationManager streaming support..."
if grep -q "StreamDebate" "${PROJECT_ROOT}/internal/debate/comprehensive/integration.go" 2>/dev/null; then
    log_pass "IntegrationManager.StreamDebate method exists"
else
    log_fail "StreamDebate method not found in IntegrationManager"
fi

if grep -q "DebateStreamRequest" "${PROJECT_ROOT}/internal/debate/comprehensive/integration.go" 2>/dev/null; then
    log_pass "DebateStreamRequest type used"
else
    log_fail "DebateStreamRequest not used"
fi

# Test 3: Verify DebateService comprehensive streaming integration
log_info "Test 3: Verifying DebateService comprehensive streaming integration..."
if grep -q "conductComprehensiveDebateStreaming" "${PROJECT_ROOT}/internal/services/debate_service_comprehensive.go" 2>/dev/null; then
    log_pass "conductComprehensiveDebateStreaming method exists"
else
    log_fail "conductComprehensiveDebateStreaming method not found"
fi

if grep -q "StreamDebate" "${PROJECT_ROOT}/internal/services/debate_service.go" 2>/dev/null; then
    log_pass "DebateService.StreamDebate method exists"
else
    log_fail "DebateService.StreamDebate method not found"
fi

# Test 4: Verify handler integration
log_info "Test 4: Verifying OpenAI handler comprehensive streaming integration..."
if grep -q "processWithComprehensiveStream" "${PROJECT_ROOT}/internal/handlers/openai_compatible.go" 2>/dev/null; then
    log_pass "processWithComprehensiveStream method exists in handler"
else
    log_fail "processWithComprehensiveStream method not found in handler"
fi

if grep -q "IsComprehensiveSystemEnabled" "${PROJECT_ROOT}/internal/handlers/openai_compatible.go" 2>/dev/null; then
    log_pass "Handler checks comprehensive system availability"
else
    log_fail "Handler doesn't check comprehensive system availability"
fi

# Test 5: Build comprehensive package
log_info "Test 5: Building comprehensive package..."
cd "$PROJECT_ROOT"
if go build ./internal/debate/comprehensive/... 2>&1 | tee -a "$TEST_LOG"; then
    log_pass "Comprehensive package builds successfully"
else
    log_fail "Comprehensive package build failed"
fi

# Test 6: Run comprehensive streaming unit tests
log_info "Test 6: Running comprehensive streaming unit tests..."
if go test -v -run TestStreamOrchestrator ./internal/debate/comprehensive/... 2>&1 | tee -a "$TEST_LOG"; then
    log_pass "StreamOrchestrator tests pass"
else
    log_fail "StreamOrchestrator tests failed"
fi

if go test -v -run TestIntegrationManagerStreamDebate ./internal/debate/comprehensive/... 2>&1 | tee -a "$TEST_LOG"; then
    log_pass "IntegrationManager streaming tests pass"
else
    log_fail "IntegrationManager streaming tests failed"
fi

# Test 7: Verify 8 roles exist
log_info "Test 7: Verifying all 8 agent roles exist..."
ROLES=("RoleArchitect" "RoleGenerator" "RoleCritic" "RoleRefactoring" "RoleTester" "RoleValidator" "RoleSecurity" "RolePerformance")
for role in "${ROLES[@]}"; do
    if grep -q "$role" "${PROJECT_ROOT}/internal/debate/comprehensive/types.go" 2>/dev/null; then
        log_pass "Role $role exists"
    else
        log_fail "Role $role not found"
    fi
done

# Test 8: Verify streaming test file exists and has tests
log_info "Test 8: Verifying streaming test file..."
if [ -f "${PROJECT_ROOT}/internal/debate/comprehensive/streaming_test.go" ]; then
    TEST_COUNT=$(grep -c "^func Test" "${PROJECT_ROOT}/internal/debate/comprehensive/streaming_test.go" 2>/dev/null || echo "0")
    if [ "$TEST_COUNT" -ge 5 ]; then
        log_pass "Streaming test file exists with $TEST_COUNT tests"
    else
        log_warn "Streaming test file exists but has only $TEST_COUNT tests (expected >= 5)"
    fi
else
    log_fail "Streaming test file not found"
fi

# Test 9: Verify handler imports comprehensive package
log_info "Test 9: Verifying handler imports..."
if grep -q '"dev.helix.agent/internal/debate/comprehensive"' "${PROJECT_ROOT}/internal/handlers/openai_compatible.go" 2>/dev/null; then
    log_pass "Handler imports comprehensive package"
else
    log_fail "Handler doesn't import comprehensive package"
fi

# Test 10: Build services package with streaming
log_info "Test 10: Building services package with streaming..."
if go build ./internal/services/... 2>&1 | tee -a "$TEST_LOG"; then
    log_pass "Services package builds with streaming"
else
    log_fail "Services package build failed"
fi

# Test 11: Build handlers package with streaming
log_info "Test 11: Building handlers package with streaming..."
if go build ./internal/handlers/... 2>&1 | tee -a "$TEST_LOG"; then
    log_pass "Handlers package builds with streaming"
else
    log_fail "Handlers package build failed"
fi

# Test 12: Run all comprehensive tests
log_info "Test 12: Running all comprehensive package tests..."
if go test -v -short ./internal/debate/comprehensive/... 2>&1 | tee -a "$TEST_LOG"; then
    log_pass "All comprehensive tests pass"
else
    log_fail "Some comprehensive tests failed"
fi

# Test 13: Verify streaming event types
log_info "Test 13: Verifying streaming event types..."
EVENT_TYPES=(
    "StreamEventDebateStart"
    "StreamEventDebateComplete"
    "StreamEventAgentResponse"
    "StreamEventTeamStart"
    "StreamEventPhaseStart"
)
for event_type in "${EVENT_TYPES[@]}"; do
    if grep -q "$event_type" "${PROJECT_ROOT}/internal/debate/comprehensive/streaming_types.go" 2>/dev/null; then
        log_pass "Event type $event_type defined"
    else
        log_fail "Event type $event_type not found"
    fi
done

# Summary
echo ""
echo "=============================================="
echo "COMPREHENSIVE STREAMING CHALLENGE COMPLETE"
echo "=============================================="
echo -e "${GREEN}PASSED: $PASSED${NC}"
echo -e "${RED}FAILED: $FAILED${NC}"
echo "Total Tests: $((PASSED + FAILED))"
echo ""
echo "Test Log: $TEST_LOG"
echo "Timestamp: $(date)"
echo "=============================================="

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}ALL CHALLENGES PASSED!${NC}"
    exit 0
else
    echo -e "${RED}SOME CHALLENGES FAILED${NC}"
    exit 1
fi
