#!/bin/bash
# Free Provider Fallback Challenge
# VALIDATES: Zen provider, OpenRouter free models, free tier handling
# Tests free provider verification and fallback (8 tests)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh" 2>/dev/null || {
    log_info() { echo -e "\033[0;34m[INFO] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_success() { echo -e "\033[0;32m[SUCCESS] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_error() { echo -e "\033[0;31m[ERROR] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_warning() { echo -e "\033[0;33m[WARNING] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
}

CHALLENGE_NAME="Free Provider Fallback Challenge"
PASSED=0
FAILED=0
TOTAL=0
PROJECT_ROOT="${SCRIPT_DIR}/../.."

log_info "=============================================="
log_info "$CHALLENGE_NAME"
log_info "=============================================="
log_info "Validates: Zen provider, OpenRouter free models, free tier handling"
log_info ""

# ============================================================================
# Section 1: Free Provider Adapter Structure
# ============================================================================

log_info "=============================================="
log_info "Section 1: Free Provider Adapter Structure"
log_info "=============================================="

# Test 1: free_adapter.go exists
TOTAL=$((TOTAL + 1))
log_info "Test 1: free_adapter.go file exists"
if [ -f "$PROJECT_ROOT/internal/verifier/adapters/free_adapter.go" ]; then
    log_success "free_adapter.go exists"
    PASSED=$((PASSED + 1))
else
    log_error "free_adapter.go NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 2: FreeProviderAdapter struct defined
TOTAL=$((TOTAL + 1))
log_info "Test 2: FreeProviderAdapter struct defined"
if grep -q "type FreeProviderAdapter struct" "$PROJECT_ROOT/internal/verifier/adapters/free_adapter.go" 2>/dev/null; then
    log_success "FreeProviderAdapter struct is defined"
    PASSED=$((PASSED + 1))
else
    log_error "FreeProviderAdapter struct NOT defined!"
    FAILED=$((FAILED + 1))
fi

# Test 3: VerifyZenProvider method exists
TOTAL=$((TOTAL + 1))
log_info "Test 3: VerifyZenProvider method exists"
if grep -q "func (fa \*FreeProviderAdapter) VerifyZenProvider" "$PROJECT_ROOT/internal/verifier/adapters/free_adapter.go" 2>/dev/null; then
    log_success "VerifyZenProvider method exists"
    PASSED=$((PASSED + 1))
else
    log_error "VerifyZenProvider method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 4: VerifyOpenRouterFreeModels method exists
TOTAL=$((TOTAL + 1))
log_info "Test 4: VerifyOpenRouterFreeModels method exists"
if grep -q "func (fa \*FreeProviderAdapter) VerifyOpenRouterFreeModels" "$PROJECT_ROOT/internal/verifier/adapters/free_adapter.go" 2>/dev/null; then
    log_success "VerifyOpenRouterFreeModels method exists"
    PASSED=$((PASSED + 1))
else
    log_error "VerifyOpenRouterFreeModels method NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 2: Free Provider Scoring
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 2: Free Provider Scoring"
log_info "=============================================="

# Test 5: Base score in range 6.0-7.0
TOTAL=$((TOTAL + 1))
log_info "Test 5: Base score for free providers (6.0-7.0)"
if grep -q "BaseScore.*6\." "$PROJECT_ROOT/internal/verifier/adapters/free_adapter.go" 2>/dev/null; then
    log_success "BaseScore is in 6.0 range for free providers"
    PASSED=$((PASSED + 1))
else
    log_error "BaseScore not properly configured for free providers!"
    FAILED=$((FAILED + 1))
fi

# Test 6: MaxScore limit defined
TOTAL=$((TOTAL + 1))
log_info "Test 6: MaxScore limit defined for free providers"
if grep -q "MaxScore.*7\." "$PROJECT_ROOT/internal/verifier/adapters/free_adapter.go" 2>/dev/null; then
    log_success "MaxScore limit is ~7.0 for free providers"
    PASSED=$((PASSED + 1))
else
    log_error "MaxScore limit NOT properly configured!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 3: AuthTypeFree Integration
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 3: AuthTypeFree Integration"
log_info "=============================================="

# Test 7: AuthTypeFree constant used
TOTAL=$((TOTAL + 1))
log_info "Test 7: AuthTypeFree used in free adapter"
if grep -q "AuthTypeFree\|verifier.AuthTypeFree" "$PROJECT_ROOT/internal/verifier/adapters/free_adapter.go" 2>/dev/null; then
    log_success "AuthTypeFree is used in free adapter"
    PASSED=$((PASSED + 1))
else
    log_error "AuthTypeFree NOT used in free adapter!"
    FAILED=$((FAILED + 1))
fi

# Test 8: Free provider models include Zen models
TOTAL=$((TOTAL + 1))
log_info "Test 8: Zen free models referenced"
if grep -q "zen.FreeModels\|zen.Model\|opencode/\|FreeProviderZen" "$PROJECT_ROOT/internal/verifier/adapters/free_adapter.go" 2>/dev/null; then
    log_success "Zen free models are referenced"
    PASSED=$((PASSED + 1))
else
    log_error "Zen free models NOT referenced!"
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
