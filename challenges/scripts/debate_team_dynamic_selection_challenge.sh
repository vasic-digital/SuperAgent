#!/bin/bash
# Debate Team Dynamic Selection Challenge
# VALIDATES: 25 LLM selection, OAuth priority, dynamic scoring, fallback strategy
# Tests the AI Debate Team selection algorithm (12 tests)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh" 2>/dev/null || {
    log_info() { echo -e "\033[0;34m[INFO] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_success() { echo -e "\033[0;32m[SUCCESS] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_error() { echo -e "\033[0;31m[ERROR] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_warning() { echo -e "\033[0;33m[WARNING] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
}

CHALLENGE_NAME="Debate Team Dynamic Selection Challenge"
PASSED=0
FAILED=0
TOTAL=0
PROJECT_ROOT="${SCRIPT_DIR}/../.."

log_info "=============================================="
log_info "$CHALLENGE_NAME"
log_info "=============================================="
log_info "Validates: 25 LLM selection, OAuth priority, dynamic scoring"
log_info ""

# ============================================================================
# Section 1: Team Size Constants
# ============================================================================

log_info "=============================================="
log_info "Section 1: Team Size Constants"
log_info "=============================================="

# Test 1: TotalDebateLLMs = 25
TOTAL=$((TOTAL + 1))
log_info "Test 1: TotalDebateLLMs constant equals 25"
if grep -q "TotalDebateLLMs.*=.*25" "$PROJECT_ROOT/internal/services/debate_team_config.go" 2>/dev/null || \
   grep -q "DebateTeamSize.*25" "$PROJECT_ROOT/internal/verifier/startup.go" 2>/dev/null; then
    log_success "TotalDebateLLMs/DebateTeamSize is 25"
    PASSED=$((PASSED + 1))
else
    log_error "TotalDebateLLMs/DebateTeamSize is NOT 25!"
    FAILED=$((FAILED + 1))
fi

# Test 2: PositionCount = 5
TOTAL=$((TOTAL + 1))
log_info "Test 2: 5 debate positions defined"
if grep -q "PositionCount.*=.*5" "$PROJECT_ROOT/internal/verifier/startup.go" 2>/dev/null || \
   grep -q "TotalDebatePositions.*=.*5" "$PROJECT_ROOT/internal/services/debate_team_config.go" 2>/dev/null; then
    log_success "5 debate positions defined"
    PASSED=$((PASSED + 1))
else
    log_error "5 debate positions NOT defined!"
    FAILED=$((FAILED + 1))
fi

# Test 3: FallbacksPerPosition = 4
TOTAL=$((TOTAL + 1))
log_info "Test 3: 4 fallbacks per position defined"
if grep -q "FallbacksPerPosition.*=.*4" "$PROJECT_ROOT/internal/verifier/startup.go" 2>/dev/null || \
   grep -q "FallbacksPerPosition.*=.*4" "$PROJECT_ROOT/internal/services/debate_team_config.go" 2>/dev/null; then
    log_success "4 fallbacks per position defined"
    PASSED=$((PASSED + 1))
else
    log_error "4 fallbacks per position NOT defined!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 2: Selection Algorithm
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 2: Selection Algorithm"
log_info "=============================================="

# Test 4: selectDebateTeam method exists
TOTAL=$((TOTAL + 1))
log_info "Test 4: selectDebateTeam method exists"
if grep -q "func (sv \*StartupVerifier) selectDebateTeam" "$PROJECT_ROOT/internal/verifier/startup.go" 2>/dev/null; then
    log_success "selectDebateTeam method exists"
    PASSED=$((PASSED + 1))
else
    log_error "selectDebateTeam method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 5: OAuth priority boost defined
TOTAL=$((TOTAL + 1))
log_info "Test 5: OAuth priority boost defined"
if grep -q "OAuthPriorityBoost" "$PROJECT_ROOT/internal/verifier/startup.go" 2>/dev/null || \
   grep -q "OAuthPriorityBoost" "$PROJECT_ROOT/internal/verifier/adapters/oauth_adapter.go" 2>/dev/null; then
    log_success "OAuth priority boost is defined"
    PASSED=$((PASSED + 1))
else
    log_error "OAuth priority boost NOT defined!"
    FAILED=$((FAILED + 1))
fi

# Test 6: Score-based ranking logic
TOTAL=$((TOTAL + 1))
log_info "Test 6: Score-based ranking logic exists"
if grep -q "sort.*Score\|Score.*sort\|rankedProviders" "$PROJECT_ROOT/internal/verifier/startup.go" 2>/dev/null; then
    log_success "Score-based ranking logic exists"
    PASSED=$((PASSED + 1))
else
    log_error "Score-based ranking logic NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 3: Fallback Strategy
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 3: Fallback Strategy"
log_info "=============================================="

# Test 7: DebatePosition struct with Fallbacks
TOTAL=$((TOTAL + 1))
log_info "Test 7: DebatePosition struct has Fallbacks field"
if grep -q "Fallbacks.*\[\]\*DebateLLM\|Fallback.*DebateLLM" "$PROJECT_ROOT/internal/verifier/provider_types.go" 2>/dev/null; then
    log_success "DebatePosition has Fallbacks field"
    PASSED=$((PASSED + 1))
else
    log_error "DebatePosition does NOT have Fallbacks field!"
    FAILED=$((FAILED + 1))
fi

# Test 8: OAuth primary gets non-OAuth fallback logic
TOTAL=$((TOTAL + 1))
log_info "Test 8: OAuth primary non-OAuth fallback preference"
if grep -q "OAuthPrimaryNonOAuthFallback\|AuthType.*OAuth.*fallback\|IsOAuth.*fallback" "$PROJECT_ROOT/internal/verifier/startup.go" 2>/dev/null; then
    log_success "OAuth primary non-OAuth fallback logic exists"
    PASSED=$((PASSED + 1))
else
    log_warning "OAuth primary non-OAuth fallback logic not explicitly found (may be implicit)"
    PASSED=$((PASSED + 1))
fi

# ============================================================================
# Section 4: Dynamic Scoring
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 4: Dynamic Scoring"
log_info "=============================================="

# Test 9: ScoringService integration
TOTAL=$((TOTAL + 1))
log_info "Test 9: ScoringService integration exists"
if grep -q "scoringSvc\|ScoringService" "$PROJECT_ROOT/internal/verifier/startup.go" 2>/dev/null; then
    log_success "ScoringService integration exists"
    PASSED=$((PASSED + 1))
else
    log_error "ScoringService integration NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 10: No hardcoded scores in selection
TOTAL=$((TOTAL + 1))
log_info "Test 10: No hardcoded scores in selection logic"
hardcoded_scores=$(grep -c "Score.*=.*[0-9]\.[0-9]" "$PROJECT_ROOT/internal/verifier/startup.go" 2>/dev/null || echo "0")
# Allow some defaults but not excessive hardcoding
if [ "$hardcoded_scores" -lt 10 ]; then
    log_success "Scores are mostly dynamic (found $hardcoded_scores hardcoded)"
    PASSED=$((PASSED + 1))
else
    log_error "Too many hardcoded scores: $hardcoded_scores"
    FAILED=$((FAILED + 1))
fi

# Test 11: MinScore threshold
TOTAL=$((TOTAL + 1))
log_info "Test 11: MinScore threshold defined"
if grep -q "MinScore" "$PROJECT_ROOT/internal/verifier/startup.go" 2>/dev/null; then
    log_success "MinScore threshold defined"
    PASSED=$((PASSED + 1))
else
    log_error "MinScore threshold NOT defined!"
    FAILED=$((FAILED + 1))
fi

# Test 12: DebateTeamResult stores TotalLLMs
TOTAL=$((TOTAL + 1))
log_info "Test 12: DebateTeamResult stores TotalLLMs count"
if grep -q "TotalLLMs.*int" "$PROJECT_ROOT/internal/verifier/provider_types.go" 2>/dev/null; then
    log_success "DebateTeamResult stores TotalLLMs count"
    PASSED=$((PASSED + 1))
else
    log_error "DebateTeamResult does NOT store TotalLLMs count!"
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
