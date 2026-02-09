#!/bin/bash
# Chat Error Handling Challenge
# Tests error responses and validation for invalid requests

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/challenge_framework.sh"

CHALLENGE_PORT="${HELIXAGENT_PORT:-7061}"
BASE_URL="http://localhost:$CHALLENGE_PORT"

# Initialize challenge
init_challenge "chat_error_handling" "Chat Error Handling Challenge"
load_env

log_info "Testing error handling and validation..."

# Test 1: Missing required fields
test_missing_required_fields() {
    log_info "Test 1: Missing required fields (no messages)"

    local request='{"model": "helixagent-debate"}'

    local response=$(curl -s -w "\n%{http_code}" "$BASE_URL/v1/chat/completions" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer ${HELIXAGENT_API_KEY:-test}" \
        -d "$request" --max-time 30 2>/dev/null || true)

    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | head -n -1)

    # Should return 400 Bad Request
    if [[ "$http_code" == "400" ]]; then
        record_assertion "missing_fields" "http_status" "true" "Returns 400 for missing messages"

        # Should have error message
        if echo "$body" | jq -e '.error' >/dev/null 2>&1; then
            record_assertion "missing_fields" "error_message" "true" "Error message present"
        else
            record_assertion "missing_fields" "error_message" "false" "No error message"
        fi
    else
        record_assertion "missing_fields" "http_status" "false" "Expected 400, got $http_code"
    fi
}

# Test 2: Invalid JSON
test_invalid_json() {
    log_info "Test 2: Invalid JSON payload"

    local request='{"model": "helixagent-debate", "messages": [invalid json'

    local response=$(curl -s -w "\n%{http_code}" "$BASE_URL/v1/chat/completions" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer ${HELIXAGENT_API_KEY:-test}" \
        -d "$request" --max-time 30 2>/dev/null || true)

    local http_code=$(echo "$response" | tail -n1)

    if [[ "$http_code" == "400" ]]; then
        record_assertion "invalid_json" "rejected" "true" "Invalid JSON rejected (400)"
    else
        record_assertion "invalid_json" "rejected" "false" "Expected 400, got $http_code"
    fi
}

# Test 3: Invalid model name
test_invalid_model() {
    log_info "Test 3: Invalid model name"

    local request='{
        "model": "nonexistent-model-12345",
        "messages": [{"role": "user", "content": "Hello"}]
    }'

    local response=$(curl -s -w "\n%{http_code}" "$BASE_URL/v1/chat/completions" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer ${HELIXAGENT_API_KEY:-test}" \
        -d "$request" --max-time 30 2>/dev/null || true)

    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | head -n -1)

    # Should return 400 or 404
    if [[ "$http_code" == "400" ]] || [[ "$http_code" == "404" ]]; then
        record_assertion "invalid_model" "rejected" "true" "Invalid model rejected ($http_code)"

        # Check for error details
        if echo "$body" | grep -qiE "(model|not found|invalid)"; then
            record_assertion "invalid_model" "error_details" "true" "Error details mention model"
        else
            record_assertion "invalid_model" "error_details" "false" "No model-specific error"
        fi
    else
        record_assertion "invalid_model" "rejected" "false" "Expected 400/404, got $http_code"
    fi
}

# Test 4: Invalid role in message
test_invalid_role() {
    log_info "Test 4: Invalid role in messages"

    local request='{
        "model": "helixagent-debate",
        "messages": [{"role": "invalid-role", "content": "Hello"}]
    }'

    local response=$(curl -s -w "\n%{http_code}" "$BASE_URL/v1/chat/completions" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer ${HELIXAGENT_API_KEY:-test}" \
        -d "$request" --max-time 30 2>/dev/null || true)

    local http_code=$(echo "$response" | tail -n1)

    if [[ "$http_code" == "400" ]]; then
        record_assertion "invalid_role" "rejected" "true" "Invalid role rejected (400)"
    else
        # May accept and process anyway
        record_assertion "invalid_role" "rejected" "false" "Invalid role not rejected ($http_code)"
    fi
}

# Test 5: Negative max_tokens
test_negative_max_tokens() {
    log_info "Test 5: Negative max_tokens"

    local request='{
        "model": "helixagent-debate",
        "messages": [{"role": "user", "content": "Hello"}],
        "max_tokens": -10
    }'

    local response=$(curl -s -w "\n%{http_code}" "$BASE_URL/v1/chat/completions" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer ${HELIXAGENT_API_KEY:-test}" \
        -d "$request" --max-time 30 2>/dev/null || true)

    local http_code=$(echo "$response" | tail -n1)

    if [[ "$http_code" == "400" ]]; then
        record_assertion "negative_tokens" "rejected" "true" "Negative max_tokens rejected"
    else
        record_assertion "negative_tokens" "rejected" "false" "Negative max_tokens not validated ($http_code)"
    fi
}

# Test 6: Invalid temperature range
test_invalid_temperature() {
    log_info "Test 6: Invalid temperature (> 2.0)"

    local request='{
        "model": "helixagent-debate",
        "messages": [{"role": "user", "content": "Hello"}],
        "temperature": 5.0
    }'

    local response=$(curl -s -w "\n%{http_code}" "$BASE_URL/v1/chat/completions" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer ${HELIXAGENT_API_KEY:-test}" \
        -d "$request" --max-time 30 2>/dev/null || true)

    local http_code=$(echo "$response" | tail -n1)

    # May accept (clamp to valid range) or reject
    if [[ "$http_code" == "400" ]]; then
        record_assertion "invalid_temp" "validation" "true" "Out-of-range temperature rejected"
    elif [[ "$http_code" == "200" ]]; then
        record_assertion "invalid_temp" "validation" "true" "Out-of-range temperature clamped/accepted"
    else
        record_assertion "invalid_temp" "validation" "false" "Unexpected status: $http_code"
    fi
}

# Test 7: Empty messages array
test_empty_messages() {
    log_info "Test 7: Empty messages array"

    local request='{
        "model": "helixagent-debate",
        "messages": []
    }'

    local response=$(curl -s -w "\n%{http_code}" "$BASE_URL/v1/chat/completions" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer ${HELIXAGENT_API_KEY:-test}" \
        -d "$request" --max-time 30 2>/dev/null || true)

    local http_code=$(echo "$response" | tail -n1)

    if [[ "$http_code" == "400" ]]; then
        record_assertion "empty_messages" "rejected" "true" "Empty messages array rejected"
    else
        record_assertion "empty_messages" "rejected" "false" "Expected 400, got $http_code"
    fi
}

# Test 8: Missing Authorization header
test_missing_auth() {
    log_info "Test 8: Missing Authorization header"

    local request='{
        "model": "helixagent-debate",
        "messages": [{"role": "user", "content": "Hello"}]
    }'

    local response=$(curl -s -w "\n%{http_code}" "$BASE_URL/v1/chat/completions" \
        -H "Content-Type: application/json" \
        -d "$request" --max-time 30 2>/dev/null || true)

    local http_code=$(echo "$response" | tail -n1)

    # May require auth (401) or allow anonymous (200)
    if [[ "$http_code" == "401" ]]; then
        record_assertion "missing_auth" "enforced" "true" "Authentication enforced (401)"
    elif [[ "$http_code" == "200" ]]; then
        record_assertion "missing_auth" "enforced" "false" "Anonymous access allowed (200)"
    else
        record_assertion "missing_auth" "enforced" "false" "Unexpected status: $http_code"
    fi
}

# Test 9: Malformed message structure
test_malformed_message() {
    log_info "Test 9: Malformed message structure"

    local request='{
        "model": "helixagent-debate",
        "messages": [{"role": "user"}]
    }'

    local response=$(curl -s -w "\n%{http_code}" "$BASE_URL/v1/chat/completions" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer ${HELIXAGENT_API_KEY:-test}" \
        -d "$request" --max-time 30 2>/dev/null || true)

    local http_code=$(echo "$response" | tail -n1)

    if [[ "$http_code" == "400" ]]; then
        record_assertion "malformed_msg" "rejected" "true" "Message without content rejected"
    else
        record_assertion "malformed_msg" "rejected" "false" "Expected 400, got $http_code"
    fi
}

# Test 10: Very large max_tokens
test_excessive_max_tokens() {
    log_info "Test 10: Excessive max_tokens value"

    local request='{
        "model": "helixagent-debate",
        "messages": [{"role": "user", "content": "Hello"}],
        "max_tokens": 1000000
    }'

    local response=$(curl -s -w "\n%{http_code}" "$BASE_URL/v1/chat/completions" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer ${HELIXAGENT_API_KEY:-test}" \
        -d "$request" --max-time 30 2>/dev/null || true)

    local http_code=$(echo "$response" | tail -n1)

    # Should reject (400) or clamp to valid range (200)
    if [[ "$http_code" == "400" ]]; then
        record_assertion "excessive_tokens" "handled" "true" "Excessive max_tokens rejected"
    elif [[ "$http_code" == "200" ]]; then
        record_assertion "excessive_tokens" "handled" "true" "Excessive max_tokens clamped"
    else
        record_assertion "excessive_tokens" "handled" "false" "Unexpected status: $http_code"
    fi
}

# Test 11: Error response format validation
test_error_format() {
    log_info "Test 11: Error response format validation"

    local request='{"model": "helixagent-debate"}'

    local response=$(curl -s -w "\n%{http_code}" "$BASE_URL/v1/chat/completions" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer ${HELIXAGENT_API_KEY:-test}" \
        -d "$request" --max-time 30 2>/dev/null || true)

    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | head -n -1)

    if [[ "$http_code" == "400" ]]; then
        # Check OpenAI error format
        local has_error=$(echo "$body" | jq -e '.error' >/dev/null 2>&1 && echo "true" || echo "false")
        local has_message=$(echo "$body" | jq -e '.error.message' >/dev/null 2>&1 && echo "true" || echo "false")
        local has_type=$(echo "$body" | jq -e '.error.type' >/dev/null 2>&1 && echo "true" || echo "false")

        if [[ "$has_error" == "true" ]]; then
            record_assertion "error_format" "has_error_object" "true" "Error object present"
        else
            record_assertion "error_format" "has_error_object" "false" "No error object"
        fi

        if [[ "$has_message" == "true" ]]; then
            record_assertion "error_format" "has_message" "true" "Error message present"
        else
            record_assertion "error_format" "has_message" "false" "No error message"
        fi
    else
        record_assertion "error_format" "http_status" "false" "Expected 400 for format check"
    fi
}

# Main execution
main() {
    log_info "Starting Chat Error Handling Challenge..."

    # Check if server is running
    if ! curl -s "$BASE_URL/health" > /dev/null 2>&1; then
        log_warning "HelixAgent not running, attempting to start..."
        start_helixagent "$CHALLENGE_PORT" || {
            log_error "Failed to start HelixAgent"
            finalize_challenge "FAILED"
            exit 1
        }
    fi

    # Run tests
    test_missing_required_fields
    test_invalid_json
    test_invalid_model
    test_invalid_role
    test_negative_max_tokens
    test_invalid_temperature
    test_empty_messages
    test_missing_auth
    test_malformed_message
    test_excessive_max_tokens
    test_error_format

    # Calculate results
    local failed_count=$(grep -c "|FAILED|" "$OUTPUT_DIR/logs/assertions.log" 2>/dev/null || echo "0")
    failed_count=${failed_count:-0}

    if [[ "$failed_count" -eq 0 ]]; then
        finalize_challenge "PASSED"
    else
        finalize_challenge "FAILED"
    fi
}

main "$@"
