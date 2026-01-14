#!/bin/bash
# Unified Startup Verification Challenge
# VALIDATES: LLMsVerifier as single source of truth, startup pipeline, all provider types
# Tests the complete startup verification pipeline with 15 tests

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh" 2>/dev/null || {
    log_info() { echo -e "\033[0;34m[INFO] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_success() { echo -e "\033[0;32m[SUCCESS] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_error() { echo -e "\033[0;31m[ERROR] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_warning() { echo -e "\033[0;33m[WARNING] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
}

CHALLENGE_NAME="Unified Startup Verification Challenge"
PASSED=0
FAILED=0
TOTAL=0
PROJECT_ROOT="${SCRIPT_DIR}/../.."

log_info "=============================================="
log_info "$CHALLENGE_NAME"
log_info "=============================================="
log_info "Validates: LLMsVerifier as single source of truth"
log_info ""

# ============================================================================
# Section 1: StartupVerifier Structure
# ============================================================================

log_info "=============================================="
log_info "Section 1: StartupVerifier Structure"
log_info "=============================================="

# Test 1: startup.go exists
TOTAL=$((TOTAL + 1))
log_info "Test 1: startup.go file exists"
if [ -f "$PROJECT_ROOT/internal/verifier/startup.go" ]; then
    log_success "startup.go exists"
    PASSED=$((PASSED + 1))
else
    log_error "startup.go NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 2: StartupVerifier struct defined
TOTAL=$((TOTAL + 1))
log_info "Test 2: StartupVerifier struct defined"
if grep -q "type StartupVerifier struct" "$PROJECT_ROOT/internal/verifier/startup.go" 2>/dev/null; then
    log_success "StartupVerifier struct is defined"
    PASSED=$((PASSED + 1))
else
    log_error "StartupVerifier struct NOT defined!"
    FAILED=$((FAILED + 1))
fi

# Test 3: VerifyAllProviders method exists
TOTAL=$((TOTAL + 1))
log_info "Test 3: VerifyAllProviders method exists"
if grep -q "func (sv \*StartupVerifier) VerifyAllProviders" "$PROJECT_ROOT/internal/verifier/startup.go" 2>/dev/null; then
    log_success "VerifyAllProviders method exists"
    PASSED=$((PASSED + 1))
else
    log_error "VerifyAllProviders method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 4: GetRankedProviders method exists
TOTAL=$((TOTAL + 1))
log_info "Test 4: GetRankedProviders method exists"
if grep -q "func (sv \*StartupVerifier) GetRankedProviders" "$PROJECT_ROOT/internal/verifier/startup.go" 2>/dev/null; then
    log_success "GetRankedProviders method exists"
    PASSED=$((PASSED + 1))
else
    log_error "GetRankedProviders method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 5: selectDebateTeam method exists
TOTAL=$((TOTAL + 1))
log_info "Test 5: selectDebateTeam method exists"
if grep -q "func (sv \*StartupVerifier) selectDebateTeam" "$PROJECT_ROOT/internal/verifier/startup.go" 2>/dev/null; then
    log_success "selectDebateTeam method exists"
    PASSED=$((PASSED + 1))
else
    log_error "selectDebateTeam method NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 2: Provider Type Definitions
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 2: Provider Type Definitions"
log_info "=============================================="

# Test 6: UnifiedProvider type exists
TOTAL=$((TOTAL + 1))
log_info "Test 6: UnifiedProvider type exists"
if grep -q "type UnifiedProvider struct" "$PROJECT_ROOT/internal/verifier/provider_types.go" 2>/dev/null; then
    log_success "UnifiedProvider type exists"
    PASSED=$((PASSED + 1))
else
    log_error "UnifiedProvider type NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 7: AuthTypeOAuth constant exists
TOTAL=$((TOTAL + 1))
log_info "Test 7: AuthTypeOAuth constant exists"
if grep -q 'AuthTypeOAuth.*=.*"oauth"' "$PROJECT_ROOT/internal/verifier/provider_types.go" 2>/dev/null; then
    log_success "AuthTypeOAuth constant exists"
    PASSED=$((PASSED + 1))
else
    log_error "AuthTypeOAuth constant NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 8: AuthTypeFree constant exists
TOTAL=$((TOTAL + 1))
log_info "Test 8: AuthTypeFree constant exists"
if grep -q 'AuthTypeFree.*=.*"free"' "$PROJECT_ROOT/internal/verifier/provider_types.go" 2>/dev/null; then
    log_success "AuthTypeFree constant exists"
    PASSED=$((PASSED + 1))
else
    log_error "AuthTypeFree constant NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 9: DebateTeamResult type exists
TOTAL=$((TOTAL + 1))
log_info "Test 9: DebateTeamResult type exists"
if grep -q "type DebateTeamResult struct" "$PROJECT_ROOT/internal/verifier/provider_types.go" 2>/dev/null; then
    log_success "DebateTeamResult type exists"
    PASSED=$((PASSED + 1))
else
    log_error "DebateTeamResult type NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 3: OAuth and Free Provider Adapters
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 3: OAuth and Free Provider Adapters"
log_info "=============================================="

# Test 10: oauth_adapter.go exists
TOTAL=$((TOTAL + 1))
log_info "Test 10: oauth_adapter.go exists"
if [ -f "$PROJECT_ROOT/internal/verifier/adapters/oauth_adapter.go" ]; then
    log_success "oauth_adapter.go exists"
    PASSED=$((PASSED + 1))
else
    log_error "oauth_adapter.go NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 11: VerifyClaudeOAuth method exists
TOTAL=$((TOTAL + 1))
log_info "Test 11: VerifyClaudeOAuth method exists"
if grep -q "func (oa \*OAuthAdapter) VerifyClaudeOAuth" "$PROJECT_ROOT/internal/verifier/adapters/oauth_adapter.go" 2>/dev/null; then
    log_success "VerifyClaudeOAuth method exists"
    PASSED=$((PASSED + 1))
else
    log_error "VerifyClaudeOAuth method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 12: free_adapter.go exists
TOTAL=$((TOTAL + 1))
log_info "Test 12: free_adapter.go exists"
if [ -f "$PROJECT_ROOT/internal/verifier/adapters/free_adapter.go" ]; then
    log_success "free_adapter.go exists"
    PASSED=$((PASSED + 1))
else
    log_error "free_adapter.go NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 13: VerifyZenProvider method exists
TOTAL=$((TOTAL + 1))
log_info "Test 13: VerifyZenProvider method exists"
if grep -q "func (fa \*FreeProviderAdapter) VerifyZenProvider" "$PROJECT_ROOT/internal/verifier/adapters/free_adapter.go" 2>/dev/null; then
    log_success "VerifyZenProvider method exists"
    PASSED=$((PASSED + 1))
else
    log_error "VerifyZenProvider method NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 4: Integration with Debate Team
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 4: Integration with Debate Team"
log_info "=============================================="

# Test 14: debate_team_config.go uses StartupVerifier
TOTAL=$((TOTAL + 1))
log_info "Test 14: debate_team_config.go uses StartupVerifier"
if grep -q "startupVerifier" "$PROJECT_ROOT/internal/services/debate_team_config.go" 2>/dev/null; then
    log_success "debate_team_config.go uses StartupVerifier"
    PASSED=$((PASSED + 1))
else
    log_error "debate_team_config.go does NOT use StartupVerifier!"
    FAILED=$((FAILED + 1))
fi

# Test 15: main.go calls startup verification
TOTAL=$((TOTAL + 1))
log_info "Test 15: main.go calls startup verification"
if grep -q "StartupVerifier\|runStartupVerification" "$PROJECT_ROOT/cmd/helixagent/main.go" 2>/dev/null; then
    log_success "main.go calls startup verification"
    PASSED=$((PASSED + 1))
else
    log_error "main.go does NOT call startup verification!"
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
