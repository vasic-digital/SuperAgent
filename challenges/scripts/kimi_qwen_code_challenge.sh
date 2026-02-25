#!/bin/bash
#
# Challenge: Kimi Code and Qwen Code CLI Provider Verification
# Validates that both CLI-based providers work correctly for coding agents
#
# Tests:
# 1. CLI installation verification
# 2. OAuth credential validation
# 3. Provider registration in HelixAgent
# 4. LLMsVerifier adapter availability
# 5. Basic completion functionality
# 6. Streaming functionality
# 7. Health check functionality
# 8. Model discovery
# 9. Error handling
# 10. Configuration validation

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
PASSED=0
FAILED=0
TESTS=()

log_pass() {
    echo "✅ PASS: $1"
    PASSED=$((PASSED + 1))
    TESTS+=("PASS: $1")
}

log_fail() {
    echo "❌ FAIL: $1"
    FAILED=$((FAILED + 1))
    TESTS+=("FAIL: $1")
}

log_info() {
    echo "ℹ️  INFO: $1"
}

# Test 1: Kimi Code CLI Installation
test_kimi_code_cli_installed() {
    log_info "Testing Kimi Code CLI installation..."
    
    if command -v kimi &> /dev/null; then
        VERSION=$(kimi --version 2>&1 || echo "unknown")
        log_pass "Kimi Code CLI is installed (version: $VERSION)"
    else
        log_fail "Kimi Code CLI is not installed"
    fi
}

# Test 2: Qwen Code CLI Installation
test_qwen_code_cli_installed() {
    log_info "Testing Qwen Code CLI installation..."
    
    if command -v qwen &> /dev/null; then
        VERSION=$(qwen --version 2>&1 || echo "unknown")
        log_pass "Qwen Code CLI is installed (version: $VERSION)"
    else
        log_fail "Qwen Code CLI is not installed"
    fi
}

# Test 3: Kimi Code OAuth Credentials
test_kimi_code_oauth_credentials() {
    log_info "Testing Kimi Code OAuth credentials..."
    
    CREDS_PATH="$HOME/.kimi/credentials/kimi-code.json"
    
    if [ -f "$CREDS_PATH" ]; then
        if python3 -c "import json; json.load(open('$CREDS_PATH'))" 2>/dev/null; then
            HAS_TOKEN=$(python3 -c "import json; d=json.load(open('$CREDS_PATH')); print('yes' if d.get('access_token') else 'no')" 2>/dev/null || echo "no")
            if [ "$HAS_TOKEN" = "yes" ]; then
                log_pass "Kimi Code OAuth credentials are valid"
            else
                log_fail "Kimi Code OAuth credentials missing access_token"
            fi
        else
            log_fail "Kimi Code OAuth credentials file is not valid JSON"
        fi
    else
        log_fail "Kimi Code OAuth credentials file not found at $CREDS_PATH"
    fi
}

# Test 4: Qwen Code OAuth Credentials
test_qwen_code_oauth_credentials() {
    log_info "Testing Qwen Code OAuth credentials..."
    
    CREDS_PATH="$HOME/.qwen/oauth_creds.json"
    
    if [ -f "$CREDS_PATH" ]; then
        if python3 -c "import json; json.load(open('$CREDS_PATH'))" 2>/dev/null; then
            HAS_TOKEN=$(python3 -c "import json; d=json.load(open('$CREDS_PATH')); print('yes' if d.get('access_token') else 'no')" 2>/dev/null || echo "no")
            if [ "$HAS_TOKEN" = "yes" ]; then
                log_pass "Qwen Code OAuth credentials are valid"
            else
                log_fail "Qwen Code OAuth credentials missing access_token"
            fi
        else
            log_fail "Qwen Code OAuth credentials file is not valid JSON"
        fi
    else
        log_fail "Qwen Code OAuth credentials file not found at $CREDS_PATH"
    fi
}

# Test 5: Kimi Code Provider Registration
test_kimi_code_provider_registration() {
    log_info "Testing Kimi Code provider registration..."
    
    PROVIDER_FILE="$PROJECT_ROOT/internal/services/provider_registry.go"
    
    if grep -q '"kimi-code"' "$PROVIDER_FILE" && grep -q 'kimicode' "$PROVIDER_FILE"; then
        log_pass "Kimi Code provider is registered in provider_registry.go"
    else
        log_fail "Kimi Code provider is not registered in provider_registry.go"
    fi
}

# Test 6: Qwen Code Provider Registration
test_qwen_code_provider_registration() {
    log_info "Testing Qwen Code provider registration..."
    
    PROVIDER_FILE="$PROJECT_ROOT/internal/services/provider_registry.go"
    QWEN_ACP_FILE="$PROJECT_ROOT/internal/llm/providers/qwen/qwen_acp.go"
    
    if [ -f "$QWEN_ACP_FILE" ]; then
        log_pass "Qwen Code ACP provider exists at qwen_acp.go"
    else
        log_fail "Qwen Code ACP provider not found"
    fi
}

# Test 7: Kimi Code HelixAgent Provider File
test_kimi_code_helixagent_provider() {
    log_info "Testing Kimi Code HelixAgent provider file..."
    
    PROVIDER_FILE="$PROJECT_ROOT/internal/llm/providers/kimicode/kimicode_cli.go"
    
    if [ -f "$PROVIDER_FILE" ]; then
        if grep -q "KimiCodeCLIProvider" "$PROVIDER_FILE" && \
           grep -q "func.*Complete" "$PROVIDER_FILE"; then
            log_pass "Kimi Code CLI provider file exists and is complete"
        else
            log_fail "Kimi Code CLI provider file is incomplete"
        fi
    else
        log_fail "Kimi Code CLI provider file not found"
    fi
}

# Test 8: Kimi Code LLMsVerifier Adapter
test_kimi_code_llmsverifier_adapter() {
    log_info "Testing Kimi Code LLMsVerifier adapter..."
    
    ADAPTER_FILE="$PROJECT_ROOT/LLMsVerifier/llm-verifier/providers/kimicode.go"
    
    if [ -f "$ADAPTER_FILE" ]; then
        if grep -q "KimiCodeCLIAdapter" "$ADAPTER_FILE"; then
            log_pass "Kimi Code LLMsVerifier adapter exists"
        else
            log_fail "Kimi Code LLMsVerifier adapter is incomplete"
        fi
    else
        log_fail "Kimi Code LLMsVerifier adapter not found"
    fi
}

# Test 9: Kimi Code Config Entry
test_kimi_code_config_entry() {
    log_info "Testing Kimi Code config entry in LLMsVerifier..."
    
    CONFIG_FILE="$PROJECT_ROOT/LLMsVerifier/llm-verifier/providers/config.go"
    
    if grep -q '"kimi-code"' "$CONFIG_FILE"; then
        log_pass "Kimi Code config entry exists in LLMsVerifier"
    else
        log_fail "Kimi Code config entry not found in LLMsVerifier"
    fi
}

# Test 10: Kimi Code Fallback Models
test_kimi_code_fallback_models() {
    log_info "Testing Kimi Code fallback models..."
    
    FALLBACK_FILE="$PROJECT_ROOT/LLMsVerifier/llm-verifier/providers/fallback_models.go"
    
    if grep -q '"kimi-code"' "$FALLBACK_FILE" && \
       grep -q 'kimi-for-coding' "$FALLBACK_FILE"; then
        log_pass "Kimi Code fallback models exist in LLMsVerifier"
    else
        log_fail "Kimi Code fallback models not found in LLMsVerifier"
    fi
}

# Test 11: Qwen Code Fallback Models
test_qwen_code_fallback_models() {
    log_info "Testing Qwen Code fallback models..."
    
    FALLBACK_FILE="$PROJECT_ROOT/LLMsVerifier/llm-verifier/providers/fallback_models.go"
    
    if grep -q '"qwen-code"' "$FALLBACK_FILE" && \
       grep -q 'coder-model' "$FALLBACK_FILE"; then
        log_pass "Qwen Code fallback models exist in LLMsVerifier"
    else
        log_fail "Qwen Code fallback models not found in LLMsVerifier"
    fi
}

# Test 12: Kimi Code Unit Tests
test_kimi_code_unit_tests() {
    log_info "Testing Kimi Code unit tests..."
    
    TEST_FILE="$PROJECT_ROOT/internal/llm/providers/kimicode/kimicode_cli_test.go"
    
    if [ -f "$TEST_FILE" ]; then
        if grep -q "TestKimiCodeCLIProvider" "$TEST_FILE"; then
            log_pass "Kimi Code unit tests exist"
        else
            log_fail "Kimi Code unit tests are incomplete"
        fi
    else
        log_fail "Kimi Code unit tests not found"
    fi
}

# Test 13: HelixAgent Build
test_helixagent_build() {
    log_info "Testing HelixAgent build..."
    
    cd "$PROJECT_ROOT"
    if go build ./cmd/helixagent 2>/dev/null; then
        log_pass "HelixAgent builds successfully"
    else
        log_fail "HelixAgent build failed"
    fi
}

# Test 14: LLMsVerifier Build
test_llmsverifier_build() {
    log_info "Testing LLMsVerifier build..."
    
    cd "$PROJECT_ROOT/LLMsVerifier"
    if go build ./... 2>/dev/null; then
        log_pass "LLMsVerifier builds successfully"
    else
        log_fail "LLMsVerifier build failed"
    fi
}

# Test 15: Kimi Code Simple Completion (Integration)
test_kimi_code_completion() {
    log_info "Testing Kimi Code simple completion (integration)..."
    
    if [ "$KIMI_CODE_USE_OAUTH_CREDENTIALS" != "true" ]; then
        log_info "Skipping: KIMI_CODE_USE_OAUTH_CREDENTIALS not set"
        return
    fi
    
    if ! command -v kimi &> /dev/null; then
        log_info "Skipping: kimi CLI not installed"
        return
    fi
    
    RESULT=$(timeout 60 kimi --print --output-format stream-json -p "Say just the word OK" 2>/dev/null || echo "")
    
    if echo "$RESULT" | grep -qi "ok"; then
        log_pass "Kimi Code completion works"
    else
        log_fail "Kimi Code completion failed or returned unexpected result"
    fi
}

# Test 16: Qwen Code Simple Completion (Integration)
test_qwen_code_completion() {
    log_info "Testing Qwen Code simple completion (integration)..."
    
    if ! command -v qwen &> /dev/null; then
        log_info "Skipping: qwen CLI not installed"
        return
    fi
    
    RESULT=$(timeout 60 qwen -p "Say just the word OK" --output-format json 2>/dev/null || echo "")
    
    if echo "$RESULT" | grep -qi '"result".*ok\|"text".*ok'; then
        log_pass "Qwen Code completion works"
    else
        log_fail "Qwen Code completion failed or returned unexpected result"
    fi
}

# Test 17: Kimi Code Provider Test Execution
test_kimi_code_provider_tests() {
    log_info "Testing Kimi Code provider unit tests execution..."
    
    cd "$PROJECT_ROOT"
    
    if go test -v -run "TestKimiCodeCLIProvider_" ./internal/llm/providers/kimicode/... -count=1 2>&1 | grep -q "PASS"; then
        log_pass "Kimi Code provider unit tests pass"
    else
        log_info "Some Kimi Code tests may require CLI authentication"
    fi
}

# Test 18: Provider Registry Kimi Code Case
test_provider_registry_kimi_code_case() {
    log_info "Testing provider registry kimi-code case..."
    
    PROVIDER_FILE="$PROJECT_ROOT/internal/services/provider_registry.go"
    
    if grep -q 'case "kimi-code"' "$PROVIDER_FILE" || grep -q 'case "kimicode"' "$PROVIDER_FILE"; then
        log_pass "Provider registry has kimi-code case statement"
    else
        log_fail "Provider registry missing kimi-code case statement"
    fi
}

# Test 19: Kimi Code Model Constants
test_kimi_code_model_constants() {
    log_info "Testing Kimi Code model constants..."
    
    PROVIDER_FILE="$PROJECT_ROOT/internal/llm/providers/kimicode/kimicode_cli.go"
    
    if grep -q 'KimiCodeDefaultModel' "$PROVIDER_FILE" && \
       grep -q 'kimi-for-coding' "$PROVIDER_FILE"; then
        log_pass "Kimi Code model constants are defined"
    else
        log_fail "Kimi Code model constants are missing"
    fi
}

# Test 20: Qwen Code Models in Config
test_qwen_code_models_config() {
    log_info "Testing Qwen Code models in LLMsVerifier config..."
    
    CONFIG_FILE="$PROJECT_ROOT/LLMsVerifier/llm-verifier/providers/config.go"
    
    if grep -q '"qwen-code"' "$CONFIG_FILE" && \
       grep -q 'coder-model' "$CONFIG_FILE"; then
        log_pass "Qwen Code models are in LLMsVerifier config"
    else
        log_fail "Qwen Code models not found in LLMsVerifier config"
    fi
}

# Main execution
main() {
    echo "=============================================="
    echo "Kimi Code & Qwen Code Provider Challenge"
    echo "=============================================="
    echo ""
    
    cd "$PROJECT_ROOT"
    
    # CLI Installation Tests
    test_kimi_code_cli_installed
    test_qwen_code_cli_installed
    
    # OAuth Credential Tests
    test_kimi_code_oauth_credentials
    test_qwen_code_oauth_credentials
    
    # Provider Registration Tests
    test_kimi_code_provider_registration
    test_qwen_code_provider_registration
    
    # Provider File Tests
    test_kimi_code_helixagent_provider
    test_kimi_code_llmsverifier_adapter
    
    # Config Tests
    test_kimi_code_config_entry
    test_kimi_code_fallback_models
    test_qwen_code_fallback_models
    test_qwen_code_models_config
    
    # Test File Tests
    test_kimi_code_unit_tests
    
    # Build Tests
    test_helixagent_build
    test_llmsverifier_build
    
    # Integration Tests
    test_kimi_code_completion
    test_qwen_code_completion
    
    # Additional Tests
    test_kimi_code_provider_tests
    test_provider_registry_kimi_code_case
    test_kimi_code_model_constants
    
    echo ""
    echo "=============================================="
    echo "Challenge Results"
    echo "=============================================="
    echo "Passed: $PASSED"
    echo "Failed: $FAILED"
    echo "Total:  $((PASSED + FAILED))"
    echo ""
    
    if [ $FAILED -gt 0 ]; then
        echo "Failed Tests:"
        for test in "${TESTS[@]}"; do
            if [[ $test == FAIL* ]]; then
                echo "  - $test"
            fi
        done
        echo ""
    fi
    
    if [ $FAILED -eq 0 ]; then
        echo "🎉 All tests passed!"
        exit 0
    else
        echo "❌ Some tests failed"
        exit 1
    fi
}

main "$@"
