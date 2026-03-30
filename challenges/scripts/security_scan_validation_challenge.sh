#!/bin/bash
# HelixAgent Challenge - Security Scan Validation
# Validates Phase 4 security hardening: strconv.Itoa fixes, SSRF annotations,
# gosec config, sonar-project.properties, and absence of integer-to-rune casts.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

if [ -f "$SCRIPT_DIR/challenge_framework.sh" ]; then
    source "$SCRIPT_DIR/challenge_framework.sh"
    init_challenge "security-scan-validation" "Security Scan Validation"
    load_env
    FRAMEWORK_LOADED=true
else
    FRAMEWORK_LOADED=false
fi

PASSED=0
FAILED=0

record_result() {
    local name="$1" status="$2"
    if [ "$FRAMEWORK_LOADED" = true ]; then
        if [ "$status" = "PASS" ]; then
            record_assertion "test" "$name" "true" "$name"
        else
            record_assertion "test" "$name" "false" "$name"
        fi
    fi
    if [ "$status" = "PASS" ]; then
        PASSED=$((PASSED + 1))
        echo -e "\033[0;32m[PASS]\033[0m $name"
    else
        FAILED=$((FAILED + 1))
        echo -e "\033[0;31m[FAIL]\033[0m $name"
    fi
}

echo "=== Security Scan Validation Challenge ==="
echo ""

# --- Integer Conversion Fix (broker.go) ---
echo "--- Broker Port Conversion Fix ---"

# Test 1: broker.go uses strconv.Itoa for port (file exists)
if [ -f "$PROJECT_ROOT/internal/messaging/broker.go" ]; then
    record_result "Messaging broker.go source file exists" "PASS"
else
    record_result "Messaging broker.go source file exists" "FAIL"
fi

# Test 2: broker.go uses strconv.Itoa for port conversion (not string(rune()))
if grep -q "strconv\.Itoa" "$PROJECT_ROOT/internal/messaging/broker.go"; then
    record_result "broker.go uses strconv.Itoa for port conversion" "PASS"
else
    record_result "broker.go uses strconv.Itoa for port conversion" "FAIL"
fi

echo ""
echo "--- CLI Types Attempt Number Fix ---"

# Test 3: notifications/cli/types.go exists
if [ -f "$PROJECT_ROOT/internal/notifications/cli/types.go" ]; then
    record_result "CLI types.go source file exists" "PASS"
else
    record_result "CLI types.go source file exists" "FAIL"
fi

# Test 4: cli/types.go uses strconv.Itoa for attempt number
if grep -q "strconv\.Itoa" "$PROJECT_ROOT/internal/notifications/cli/types.go"; then
    record_result "cli/types.go uses strconv.Itoa for attempt number" "PASS"
else
    record_result "cli/types.go uses strconv.Itoa for attempt number" "FAIL"
fi

echo ""
echo "--- SSRF Suppression Annotations ---"

# Test 5: webhook_dispatcher.go has #nosec annotation for URL usage
if grep -q "#nosec" "$PROJECT_ROOT/internal/notifications/webhook_dispatcher.go"; then
    record_result "webhook_dispatcher.go has #nosec SSRF annotation" "PASS"
else
    record_result "webhook_dispatcher.go has #nosec SSRF annotation" "FAIL"
fi

# Test 6: formatter service base.go has #nosec annotation for URL usage
if grep -q "#nosec" "$PROJECT_ROOT/internal/formatters/providers/service/base.go"; then
    record_result "formatter service base.go has #nosec SSRF annotation" "PASS"
else
    record_result "formatter service base.go has #nosec SSRF annotation" "FAIL"
fi

# Test 7: SSRF annotation in webhook_dispatcher.go explains admin-configured URL
if grep -q "admin-configured\|admin configured" \
    "$PROJECT_ROOT/internal/notifications/webhook_dispatcher.go"; then
    record_result "webhook_dispatcher.go SSRF annotation explains rationale" "PASS"
else
    record_result "webhook_dispatcher.go SSRF annotation explains rationale" "FAIL"
fi

echo ""
echo "--- Security Tooling Configuration ---"

# Test 8: gosec configuration file exists
if [ -f "$PROJECT_ROOT/.gosec.yml" ]; then
    record_result "gosec configuration file (.gosec.yml) exists" "PASS"
else
    record_result "gosec configuration file (.gosec.yml) exists" "FAIL"
fi

# Test 9: sonar-project.properties exists at project root
if [ -f "$PROJECT_ROOT/sonar-project.properties" ]; then
    record_result "sonar-project.properties exists at project root" "PASS"
else
    record_result "sonar-project.properties exists at project root" "FAIL"
fi

echo ""
echo "--- No Dangerous Rune Casts ---"

# Test 10: No string(rune(c.Port)) or string(rune(integer)) patterns remain
RUNE_CAST_COUNT=$(grep -rn "string(rune(" "$PROJECT_ROOT/internal/" \
    --include="*.go" 2>/dev/null | grep -v "_test.go" | wc -l)
if [ "$RUNE_CAST_COUNT" -eq 0 ]; then
    record_result "No string(rune(integer)) patterns remain in production code" "PASS"
else
    record_result "No string(rune(integer)) patterns remain in production code (found $RUNE_CAST_COUNT)" "FAIL"
fi

echo ""
echo "=== Results ==="
TOTAL=$((PASSED + FAILED))
echo "Passed: $PASSED/$TOTAL"
echo "Failed: $FAILED/$TOTAL"

if [ "$FRAMEWORK_LOADED" = true ]; then
    finalize_challenge "$PASSED" "$TOTAL"
fi

if [ "$FAILED" -gt 0 ]; then
    exit 1
fi
