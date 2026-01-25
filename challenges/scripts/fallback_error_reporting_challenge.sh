#!/bin/bash
# ==============================================================================
# FALLBACK ERROR REPORTING CHALLENGE
# ==============================================================================
# This challenge validates that fallback events include detailed error
# information in responses, with proper visual indicators for CLI agent plugins.
#
# Key Requirements:
# 1. Error cause must be included when fallback occurs
# 2. Error categories must be properly detected
# 3. Visual indicators (icons) must match error categories
# 4. Format-aware output (ANSI/Markdown/Plain) must work correctly
# 5. Fallback chain visualization must show all attempts and errors
#
# Run: ./challenges/scripts/fallback_error_reporting_challenge.sh
# ==============================================================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

PASS_COUNT=0
FAIL_COUNT=0
TOTAL_TESTS=0

# Helper function to run a test
run_test() {
    local test_name="$1"
    local test_cmd="$2"
    local expected="$3"

    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    echo -n "  [$TOTAL_TESTS] $test_name... "

    result=$(eval "$test_cmd" 2>&1) || true

    if echo "$result" | grep -q "$expected"; then
        echo -e "${GREEN}PASS${NC}"
        PASS_COUNT=$((PASS_COUNT + 1))
        return 0
    else
        echo -e "${RED}FAIL${NC}"
        echo "    Expected to find: $expected"
        echo "    Got: ${result:0:200}..."
        FAIL_COUNT=$((FAIL_COUNT + 1))
        return 1
    fi
}

# Helper function to run a negative test (should NOT contain)
run_negative_test() {
    local test_name="$1"
    local test_cmd="$2"
    local not_expected="$3"

    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    echo -n "  [$TOTAL_TESTS] $test_name... "

    result=$(eval "$test_cmd" 2>&1) || true

    if ! echo "$result" | grep -q "$not_expected"; then
        echo -e "${GREEN}PASS${NC}"
        PASS_COUNT=$((PASS_COUNT + 1))
        return 0
    else
        echo -e "${RED}FAIL${NC}"
        echo "    Should NOT contain: $not_expected"
        FAIL_COUNT=$((FAIL_COUNT + 1))
        return 1
    fi
}

echo -e "\n${CYAN}============================================================${NC}"
echo -e "${CYAN}   FALLBACK ERROR REPORTING CHALLENGE${NC}"
echo -e "${CYAN}============================================================${NC}"
echo ""

# ==============================================================================
# SECTION 1: Error Category Detection Tests
# ==============================================================================
echo -e "\n${YELLOW}Section 1: Error Category Detection${NC}"
echo "Testing that error messages are correctly categorized..."

# These tests run Go unit tests directly
run_test "Rate limit error detection" \
    "go test -v -run 'TestCategorizeErrorString/Rate_limit_error' ./internal/handlers/... 2>&1" \
    "PASS"

run_test "Timeout error detection" \
    "go test -v -run 'TestCategorizeErrorString/Timeout_error' ./internal/handlers/... 2>&1" \
    "PASS"

run_test "Auth error detection" \
    "go test -v -run 'TestCategorizeErrorString/Auth_error_401' ./internal/handlers/... 2>&1" \
    "PASS"

run_test "Connection error detection" \
    "go test -v -run 'TestCategorizeErrorString/Connection_refused' ./internal/handlers/... 2>&1" \
    "PASS"

run_test "Service unavailable detection" \
    "go test -v -run 'TestCategorizeErrorString/Service_unavailable' ./internal/handlers/... 2>&1" \
    "PASS"

run_test "Quota exceeded detection" \
    "go test -v -run 'TestCategorizeErrorString/Quota_exceeded' ./internal/handlers/... 2>&1" \
    "PASS"

# ==============================================================================
# SECTION 2: Visual Indicator Tests
# ==============================================================================
echo -e "\n${YELLOW}Section 2: Visual Indicators (Icons)${NC}"
echo "Testing that error categories map to correct visual icons..."

run_test "Rate limit icon" \
    "go test -v -run 'TestGetCategoryIcon/rate_limit' ./internal/handlers/... 2>&1" \
    "PASS"

run_test "Timeout icon" \
    "go test -v -run 'TestGetCategoryIcon/timeout' ./internal/handlers/... 2>&1" \
    "PASS"

run_test "Auth icon" \
    "go test -v -run 'TestGetCategoryIcon/auth' ./internal/handlers/... 2>&1" \
    "PASS"

run_test "Connection icon" \
    "go test -v -run 'TestGetCategoryIcon/connection' ./internal/handlers/... 2>&1" \
    "PASS"

run_test "Unavailable icon" \
    "go test -v -run 'TestGetCategoryIcon/unavailable' ./internal/handlers/... 2>&1" \
    "PASS"

# ==============================================================================
# SECTION 3: Markdown Format Tests
# ==============================================================================
echo -e "\n${YELLOW}Section 3: Markdown Format Output${NC}"
echo "Testing that Markdown output is clean and informative..."

run_test "Fallback triggered contains error info" \
    "go test -v -run 'TestFormatFallbackTriggeredMarkdown/Contains_all_required_information' ./internal/handlers/... 2>&1" \
    "PASS"

run_test "Fallback triggered contains no ANSI" \
    "go test -v -run 'TestFormatFallbackTriggeredMarkdown/Contains_no_ANSI_codes' ./internal/handlers/... 2>&1" \
    "PASS"

run_test "Fallback success format" \
    "go test -v -run 'TestFormatFallbackSuccessMarkdown/Shows_success_message' ./internal/handlers/... 2>&1" \
    "PASS"

run_test "Fallback failed format with error" \
    "go test -v -run 'TestFormatFallbackFailedMarkdown/Shows_failure_with_error' ./internal/handlers/... 2>&1" \
    "PASS"

run_test "Fallback exhausted format" \
    "go test -v -run 'TestFormatFallbackExhaustedMarkdown/Shows_exhausted_message' ./internal/handlers/... 2>&1" \
    "PASS"

# ==============================================================================
# SECTION 4: Format-Aware Output Tests
# ==============================================================================
echo -e "\n${YELLOW}Section 4: Format-Aware Output${NC}"
echo "Testing that output adapts to different output formats..."

run_test "ANSI format contains ANSI codes" \
    "go test -v -run 'TestFormatFallbackWithErrorForFormat/ANSI_format' ./internal/handlers/... 2>&1" \
    "PASS"

run_test "Markdown format is clean" \
    "go test -v -run 'TestFormatFallbackWithErrorForFormat/Markdown_format' ./internal/handlers/... 2>&1" \
    "PASS"

run_test "Plain format has no formatting" \
    "go test -v -run 'TestFormatFallbackWithErrorForFormat/Plain_format' ./internal/handlers/... 2>&1" \
    "PASS"

# ==============================================================================
# SECTION 5: Fallback Chain Visualization Tests
# ==============================================================================
echo -e "\n${YELLOW}Section 5: Fallback Chain Visualization${NC}"
echo "Testing that complete fallback chains are properly visualized..."

run_test "Chain shows all attempts" \
    "go test -v -run 'TestFormatFallbackChainMarkdown/Shows_complete_chain' ./internal/handlers/... 2>&1" \
    "PASS"

run_test "Empty chain handling" \
    "go test -v -run 'TestFormatFallbackChainMarkdown/Empty_chain_returns_empty_string' ./internal/handlers/... 2>&1" \
    "PASS"

run_test "Chain with errors format - ANSI" \
    "go test -v -run 'TestFormatFallbackChainWithErrorsForFormat/ANSI_format' ./internal/handlers/... 2>&1" \
    "PASS"

run_test "Chain with errors format - Markdown" \
    "go test -v -run 'TestFormatFallbackChainWithErrorsForFormat/Markdown_format' ./internal/handlers/... 2>&1" \
    "PASS"

run_test "Chain with errors format - Plain" \
    "go test -v -run 'TestFormatFallbackChainWithErrorsForFormat/Plain_format' ./internal/handlers/... 2>&1" \
    "PASS"

# ==============================================================================
# SECTION 6: Edge Cases
# ==============================================================================
echo -e "\n${YELLOW}Section 6: Edge Cases${NC}"
echo "Testing edge cases and error handling..."

run_test "Multiple keywords - first match wins" \
    "go test -v -run 'TestErrorCategoryEdgeCases/Multiple_keywords' ./internal/handlers/... 2>&1" \
    "PASS"

run_test "Case insensitivity" \
    "go test -v -run 'TestErrorCategoryEdgeCases/Case_insensitivity' ./internal/handlers/... 2>&1" \
    "PASS"

# ==============================================================================
# SECTION 7: Integration Test
# ==============================================================================
echo -e "\n${YELLOW}Section 7: Integration Test${NC}"
echo "Testing complete fallback error reporting flow..."

run_test "Full fallback sequence is clean" \
    "go test -v -run 'TestFullFallbackErrorReportingFlow' ./internal/handlers/... 2>&1" \
    "PASS"

# ==============================================================================
# SECTION 8: CLI Notification Types Tests
# ==============================================================================
echo -e "\n${YELLOW}Section 8: CLI Notification Types${NC}"
echo "Testing CLI agent plugin notification types..."

# Check that the FallbackIndicatorContent type exists
run_test "FallbackIndicatorContent type exists" \
    "grep -r 'type FallbackIndicatorContent struct' ./internal/notifications/cli/types.go" \
    "FallbackIndicatorContent"

# Check that fallback icons are defined
run_test "Fallback status icons defined" \
    "grep -r 'IconFallbackTriggered' ./internal/notifications/cli/types.go" \
    "IconFallbackTriggered"

# Check error category icons
run_test "Error category icons defined" \
    "grep -r 'IconErrorRateLimit' ./internal/notifications/cli/types.go" \
    "IconErrorRateLimit"

# ==============================================================================
# SECTION 9: Event Stream Types Tests
# ==============================================================================
echo -e "\n${YELLOW}Section 9: Event Stream Types${NC}"
echo "Testing fallback event types in event stream..."

# Check that fallback event types are defined
run_test "FallbackTriggered event type defined" \
    "grep -r 'EventTypeFallbackTriggered' ./internal/messaging/event_stream.go" \
    "fallback.triggered"

run_test "FallbackEventData type defined" \
    "grep -r 'type FallbackEventData struct' ./internal/messaging/event_stream.go" \
    "FallbackEventData"

run_test "ErrorCategory field in FallbackEventData" \
    "grep -r 'ErrorCategory' ./internal/messaging/event_stream.go" \
    "ErrorCategory"

run_test "FallbackChainEntry type defined" \
    "grep -r 'type FallbackChainEntry struct' ./internal/messaging/event_stream.go" \
    "FallbackChainEntry"

# ==============================================================================
# SECTION 10: Handler Integration Tests
# ==============================================================================
echo -e "\n${YELLOW}Section 10: Handler Integration${NC}"
echo "Testing that handlers use the new fallback error reporting..."

# Check that openai_compatible.go uses format-aware fallback formatting
run_test "Handler uses FormatFallbackWithErrorForFormat" \
    "grep -r 'FormatFallbackWithErrorForFormat' ./internal/handlers/openai_compatible.go" \
    "FormatFallbackWithErrorForFormat"

# Check that handler extracts error from fallback chain
run_test "Handler extracts primary error" \
    "grep -r 'primaryError' ./internal/handlers/openai_compatible.go" \
    "primaryError"

# Check that handler logs detailed fallback info
run_test "Handler logs fallback with error category" \
    "grep -r 'error_category' ./internal/handlers/openai_compatible.go" \
    "error_category"

# ==============================================================================
# RESULTS SUMMARY
# ==============================================================================
echo -e "\n${CYAN}============================================================${NC}"
echo -e "${CYAN}   CHALLENGE RESULTS${NC}"
echo -e "${CYAN}============================================================${NC}"
echo ""
echo -e "Total Tests: $TOTAL_TESTS"
echo -e "Passed: ${GREEN}$PASS_COUNT${NC}"
echo -e "Failed: ${RED}$FAIL_COUNT${NC}"
echo ""

if [ $FAIL_COUNT -eq 0 ]; then
    echo -e "${GREEN}============================================================${NC}"
    echo -e "${GREEN}   ALL TESTS PASSED! CHALLENGE COMPLETED SUCCESSFULLY!${NC}"
    echo -e "${GREEN}============================================================${NC}"
    exit 0
else
    echo -e "${RED}============================================================${NC}"
    echo -e "${RED}   CHALLENGE FAILED - $FAIL_COUNT tests need attention${NC}"
    echo -e "${RED}============================================================${NC}"
    exit 1
fi
