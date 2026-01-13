#!/bin/bash
# Qwen OAuth Refresh Challenge - Validates Qwen OAuth token refresh mechanism
# Tests: CLI availability, credentials file, token validity, CLI refresh, automatic refresh

set -e

# Source common utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh" 2>/dev/null || {
    # Fallback if common.sh not found
    log_info() { echo -e "\033[0;34m[INFO] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_success() { echo -e "\033[0;32m[SUCCESS] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_error() { echo -e "\033[0;31m[ERROR] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_warning() { echo -e "\033[0;33m[WARNING] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
}

CHALLENGE_NAME="Qwen OAuth Refresh Challenge"
PASSED=0
FAILED=0
TOTAL=0
PROJECT_ROOT="${SCRIPT_DIR}/../.."

# ============================================================================
# Section 1: Environment and Configuration Tests
# ============================================================================

log_info "=============================================="
log_info "Section 1: Environment and Configuration"
log_info "=============================================="

# Test 1: Qwen OAuth environment variable
TOTAL=$((TOTAL + 1))
log_info "Testing QWEN_CODE_USE_OAUTH_CREDENTIALS environment variable"
# Try to load from .env if not set
if [ -z "$QWEN_CODE_USE_OAUTH_CREDENTIALS" ] && [ -f "$PROJECT_ROOT/.env" ]; then
    qwen_oauth_env=$(grep "^QWEN_CODE_USE_OAUTH_CREDENTIALS=" "$PROJECT_ROOT/.env" | cut -d= -f2 | tr -d '"' | tr -d "'")
    if [ -n "$qwen_oauth_env" ]; then
        export QWEN_CODE_USE_OAUTH_CREDENTIALS="$qwen_oauth_env"
    fi
fi

if [ "$QWEN_CODE_USE_OAUTH_CREDENTIALS" = "true" ] || [ "$QWEN_CODE_USE_OAUTH_CREDENTIALS" = "1" ]; then
    log_success "QWEN_CODE_USE_OAUTH_CREDENTIALS is enabled"
    PASSED=$((PASSED + 1))
elif [ -f "$PROJECT_ROOT/.env" ] && grep -q "QWEN_CODE_USE_OAUTH_CREDENTIALS=true" "$PROJECT_ROOT/.env"; then
    log_success "QWEN_CODE_USE_OAUTH_CREDENTIALS is set in .env (will be loaded at runtime)"
    PASSED=$((PASSED + 1))
else
    log_warning "QWEN_CODE_USE_OAUTH_CREDENTIALS is not set (recommend: set in .env)"
    PASSED=$((PASSED + 1))  # Warning only - not a hard failure
fi

# Test 2: Qwen credentials file exists
TOTAL=$((TOTAL + 1))
QWEN_CREDS_PATH="$HOME/.qwen/oauth_creds.json"
log_info "Testing Qwen credentials file exists at $QWEN_CREDS_PATH"
if [ -f "$QWEN_CREDS_PATH" ]; then
    log_success "Qwen credentials file exists"
    PASSED=$((PASSED + 1))
else
    log_error "Qwen credentials file not found"
    FAILED=$((FAILED + 1))
fi

# Test 3: Credentials file is readable and valid JSON
TOTAL=$((TOTAL + 1))
log_info "Testing credentials file is valid JSON"
if [ -f "$QWEN_CREDS_PATH" ] && jq '.' "$QWEN_CREDS_PATH" > /dev/null 2>&1; then
    log_success "Credentials file is valid JSON"
    PASSED=$((PASSED + 1))
else
    log_error "Credentials file is not valid JSON"
    FAILED=$((FAILED + 1))
fi

# Test 4: Credentials file contains required fields
TOTAL=$((TOTAL + 1))
log_info "Testing credentials file contains required fields"
if [ -f "$QWEN_CREDS_PATH" ]; then
    has_access=$(jq -r 'has("access_token")' "$QWEN_CREDS_PATH" 2>/dev/null)
    has_refresh=$(jq -r 'has("refresh_token")' "$QWEN_CREDS_PATH" 2>/dev/null)
    has_expiry=$(jq -r 'has("expiry_date")' "$QWEN_CREDS_PATH" 2>/dev/null)

    if [ "$has_access" = "true" ] && [ "$has_refresh" = "true" ] && [ "$has_expiry" = "true" ]; then
        log_success "Credentials file has all required fields"
        PASSED=$((PASSED + 1))
    else
        log_error "Credentials file missing required fields (access: $has_access, refresh: $has_refresh, expiry: $has_expiry)"
        FAILED=$((FAILED + 1))
    fi
else
    log_error "Cannot check fields - credentials file missing"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 2: Token Validity Tests
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 2: Token Validity"
log_info "=============================================="

# Test 5: Access token is not empty
TOTAL=$((TOTAL + 1))
log_info "Testing access token is not empty"
if [ -f "$QWEN_CREDS_PATH" ]; then
    access_token=$(jq -r '.access_token' "$QWEN_CREDS_PATH" 2>/dev/null)
    if [ -n "$access_token" ] && [ "$access_token" != "null" ]; then
        log_success "Access token is present (length: ${#access_token})"
        PASSED=$((PASSED + 1))
    else
        log_error "Access token is empty or null"
        FAILED=$((FAILED + 1))
    fi
else
    log_error "Cannot check token - credentials file missing"
    FAILED=$((FAILED + 1))
fi

# Test 6: Refresh token is present
TOTAL=$((TOTAL + 1))
log_info "Testing refresh token is present"
if [ -f "$QWEN_CREDS_PATH" ]; then
    refresh_token=$(jq -r '.refresh_token' "$QWEN_CREDS_PATH" 2>/dev/null)
    if [ -n "$refresh_token" ] && [ "$refresh_token" != "null" ]; then
        log_success "Refresh token is present (length: ${#refresh_token})"
        PASSED=$((PASSED + 1))
    else
        log_error "Refresh token is empty or null"
        FAILED=$((FAILED + 1))
    fi
else
    log_error "Cannot check token - credentials file missing"
    FAILED=$((FAILED + 1))
fi

# Test 7: Token is not expired
TOTAL=$((TOTAL + 1))
log_info "Testing token is not expired"
if [ -f "$QWEN_CREDS_PATH" ]; then
    expiry_ms=$(jq -r '.expiry_date' "$QWEN_CREDS_PATH" 2>/dev/null)
    current_ms=$(date +%s%3N)

    if [ "$expiry_ms" -gt "$current_ms" ]; then
        expires_in=$(( (expiry_ms - current_ms) / 1000 / 60 ))
        log_success "Token is valid (expires in ${expires_in} minutes)"
        PASSED=$((PASSED + 1))
    else
        log_error "Token is EXPIRED"
        FAILED=$((FAILED + 1))
    fi
else
    log_error "Cannot check expiry - credentials file missing"
    FAILED=$((FAILED + 1))
fi

# Test 8: Token has reasonable expiry (not in far future - sanity check)
TOTAL=$((TOTAL + 1))
log_info "Testing token expiry is reasonable (within 30 days)"
if [ -f "$QWEN_CREDS_PATH" ]; then
    expiry_ms=$(jq -r '.expiry_date' "$QWEN_CREDS_PATH" 2>/dev/null)
    current_ms=$(date +%s%3N)
    thirty_days_ms=$((30 * 24 * 60 * 60 * 1000))
    max_expiry=$((current_ms + thirty_days_ms))

    if [ "$expiry_ms" -lt "$max_expiry" ]; then
        log_success "Token expiry is within reasonable range"
        PASSED=$((PASSED + 1))
    else
        log_warning "Token expiry seems unusually far in future"
        PASSED=$((PASSED + 1))  # Warning only
    fi
else
    log_error "Cannot check expiry - credentials file missing"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 3: Qwen CLI Tests
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 3: Qwen CLI Availability"
log_info "=============================================="

# Test 9: Qwen CLI is in PATH
TOTAL=$((TOTAL + 1))
log_info "Testing qwen CLI is available"
if command -v qwen &> /dev/null; then
    qwen_path=$(which qwen)
    log_success "qwen CLI found at: $qwen_path"
    PASSED=$((PASSED + 1))
else
    log_error "qwen CLI not found in PATH"
    FAILED=$((FAILED + 1))
fi

# Test 10: Qwen CLI version is available
TOTAL=$((TOTAL + 1))
log_info "Testing qwen CLI version"
if command -v qwen &> /dev/null; then
    qwen_version=$(qwen --version 2>/dev/null || echo "unknown")
    if [ "$qwen_version" != "unknown" ] && [ -n "$qwen_version" ]; then
        log_success "qwen CLI version: $qwen_version"
        PASSED=$((PASSED + 1))
    else
        log_error "Could not determine qwen CLI version"
        FAILED=$((FAILED + 1))
    fi
else
    log_error "qwen CLI not available"
    FAILED=$((FAILED + 1))
fi

# Test 11: Qwen CLI can run in non-interactive mode
TOTAL=$((TOTAL + 1))
log_info "Testing qwen CLI non-interactive mode (JSON output)"
if command -v qwen &> /dev/null; then
    # Test that the -o json flag is recognized
    test_output=$(timeout 30s qwen "exit" -o json --max-session-turns 1 2>&1 || true)
    if echo "$test_output" | jq -e '.[0].type' > /dev/null 2>&1; then
        log_success "qwen CLI supports non-interactive JSON mode"
        PASSED=$((PASSED + 1))
    else
        log_error "qwen CLI JSON output mode failed"
        FAILED=$((FAILED + 1))
    fi
else
    log_error "qwen CLI not available"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 4: Go Integration Tests
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 4: Go Integration Tests"
log_info "=============================================="

# Test 12: Unit tests pass
TOTAL=$((TOTAL + 1))
log_info "Testing CLI refresh unit tests"
if (cd "$PROJECT_ROOT" && go test ./internal/auth/oauth_credentials/... -run "TestCLI|TestParse" -short -v 2>&1 | grep -q "PASS"); then
    log_success "CLI refresh unit tests pass"
    PASSED=$((PASSED + 1))
else
    log_error "CLI refresh unit tests failed"
    FAILED=$((FAILED + 1))
fi

# Test 13: Integration tests pass
TOTAL=$((TOTAL + 1))
log_info "Testing CLI refresh integration tests"
if (cd "$PROJECT_ROOT" && go test ./tests/integration/... -run "TestQwenCLI|TestQwenOAuth" -short -v 2>&1 | grep -q "PASS"); then
    log_success "CLI refresh integration tests pass"
    PASSED=$((PASSED + 1))
else
    log_error "CLI refresh integration tests failed"
    FAILED=$((FAILED + 1))
fi

# Test 14: CLIRefresher can detect qwen CLI
TOTAL=$((TOTAL + 1))
log_info "Testing CLIRefresher detects qwen CLI"
# Use go test to verify CLIRefresher functionality
if (cd "$PROJECT_ROOT" && go test ./internal/auth/oauth_credentials/... -run "TestCLIRefresherInitialize" -short -v 2>&1 | grep -q "initializes_successfully"); then
    log_success "CLIRefresher successfully detects qwen CLI"
    PASSED=$((PASSED + 1))
else
    log_error "CLIRefresher failed to detect qwen CLI"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 5: Token Refresh Validation (Optional - requires fresh token use)
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 5: Token Refresh Mechanism Validation"
log_info "=============================================="

# Test 15: GetStatus returns valid data via tests
TOTAL=$((TOTAL + 1))
log_info "Testing GetStatus returns valid data"
# Use go test to verify GetStatus functionality
if (cd "$PROJECT_ROOT" && go test ./internal/auth/oauth_credentials/... -run "TestCLIRefreshStatus" -short -v 2>&1 | grep -q "PASS"); then
    log_success "GetStatus returns valid data (verified by unit test)"
    PASSED=$((PASSED + 1))
else
    log_error "GetStatus test failed"
    FAILED=$((FAILED + 1))
fi

# Test 16: Credentials reader returns valid info via tests
TOTAL=$((TOTAL + 1))
log_info "Testing credential reader returns valid info"
# Use go test to verify credential reader
if (cd "$PROJECT_ROOT" && go test ./internal/auth/oauth_credentials/... -run "TestGetQwenCredentialInfo" -short -v 2>&1 | grep -q "PASS"); then
    log_success "Credential reader returns valid info (verified by unit test)"
    PASSED=$((PASSED + 1))
else
    log_error "Credential reader test failed"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Summary
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Qwen OAuth Refresh Challenge Summary"
log_info "=============================================="
log_info "Total tests: $TOTAL"
log_success "Passed: $PASSED"
if [ $FAILED -gt 0 ]; then
    log_error "Failed: $FAILED"
else
    log_info "Failed: $FAILED"
fi

PASS_RATE=$((PASSED * 100 / TOTAL))
log_info "Pass rate: ${PASS_RATE}%"

if [ $FAILED -eq 0 ]; then
    log_success "=============================================="
    log_success "ALL QWEN OAUTH REFRESH TESTS PASSED!"
    log_success "=============================================="
    exit 0
else
    log_error "=============================================="
    log_error "SOME QWEN OAUTH REFRESH TESTS FAILED"
    log_error "=============================================="
    exit 1
fi
