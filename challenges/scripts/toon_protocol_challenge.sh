#!/bin/bash
# toon_protocol_challenge.sh - TOON Protocol Challenge
# Tests Token-Optimized Object Notation implementation for HelixAgent

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
source "$SCRIPT_DIR/common.sh"

CHALLENGE_NAME="TOON Protocol Challenge"
PASSED=0
FAILED=0
TOTAL=0

log_test() {
    local test_name="$1"
    local status="$2"
    TOTAL=$((TOTAL + 1))
    if [ "$status" = "PASS" ]; then
        PASSED=$((PASSED + 1))
        echo -e "  \e[32m✓\e[0m $test_name"
    else
        FAILED=$((FAILED + 1))
        echo -e "  \e[31m✗\e[0m $test_name"
    fi
}

echo "=============================================="
echo "  $CHALLENGE_NAME"
echo "=============================================="
echo ""

cd "$PROJECT_ROOT"

# Test 1: TOON package structure
echo "[1] TOON Package Structure"
if [ -f "internal/toon/encoder.go" ]; then
    log_test "encoder.go exists" "PASS"
else
    log_test "encoder.go exists" "FAIL"
fi

if [ -f "internal/toon/transport.go" ]; then
    log_test "transport.go exists" "PASS"
else
    log_test "transport.go exists" "FAIL"
fi

if [ -f "internal/toon/encoder_test.go" ]; then
    log_test "encoder_test.go exists" "PASS"
else
    log_test "encoder_test.go exists" "FAIL"
fi

if [ -f "internal/toon/transport_test.go" ]; then
    log_test "transport_test.go exists" "PASS"
else
    log_test "transport_test.go exists" "FAIL"
fi

# Test 2: Encoder implementation
echo ""
echo "[2] Encoder Implementation"
if grep -q "type Encoder struct" internal/toon/encoder.go 2>/dev/null; then
    log_test "Encoder struct defined" "PASS"
else
    log_test "Encoder struct defined" "FAIL"
fi

if grep -q "func NewEncoder" internal/toon/encoder.go 2>/dev/null; then
    log_test "NewEncoder function" "PASS"
else
    log_test "NewEncoder function" "FAIL"
fi

if grep -q "func.*Encoder.*Encode" internal/toon/encoder.go 2>/dev/null; then
    log_test "Encode method" "PASS"
else
    log_test "Encode method" "FAIL"
fi

if grep -q "func.*Encoder.*EncodeToString" internal/toon/encoder.go 2>/dev/null; then
    log_test "EncodeToString method" "PASS"
else
    log_test "EncodeToString method" "FAIL"
fi

# Test 3: Decoder implementation
echo ""
echo "[3] Decoder Implementation"
if grep -q "type Decoder struct" internal/toon/encoder.go 2>/dev/null; then
    log_test "Decoder struct defined" "PASS"
else
    log_test "Decoder struct defined" "FAIL"
fi

if grep -q "func NewDecoder" internal/toon/encoder.go 2>/dev/null; then
    log_test "NewDecoder function" "PASS"
else
    log_test "NewDecoder function" "FAIL"
fi

if grep -q "func.*Decoder.*Decode" internal/toon/encoder.go 2>/dev/null; then
    log_test "Decode method" "PASS"
else
    log_test "Decode method" "FAIL"
fi

# Test 4: Compression levels
echo ""
echo "[4] Compression Levels"
if grep -q "CompressionNone" internal/toon/encoder.go 2>/dev/null; then
    log_test "CompressionNone level" "PASS"
else
    log_test "CompressionNone level" "FAIL"
fi

if grep -q "CompressionMinimal" internal/toon/encoder.go 2>/dev/null; then
    log_test "CompressionMinimal level" "PASS"
else
    log_test "CompressionMinimal level" "FAIL"
fi

if grep -q "CompressionStandard" internal/toon/encoder.go 2>/dev/null; then
    log_test "CompressionStandard level" "PASS"
else
    log_test "CompressionStandard level" "FAIL"
fi

if grep -q "CompressionAggressive" internal/toon/encoder.go 2>/dev/null; then
    log_test "CompressionAggressive level" "PASS"
else
    log_test "CompressionAggressive level" "FAIL"
fi

# Test 5: Key mapping
echo ""
echo "[5] Key Mapping"
if grep -q "DefaultKeyMapping" internal/toon/encoder.go 2>/dev/null; then
    log_test "DefaultKeyMapping function" "PASS"
else
    log_test "DefaultKeyMapping function" "FAIL"
fi

if grep -q '"id".*"i"' internal/toon/encoder.go 2>/dev/null; then
    log_test "id -> i mapping" "PASS"
else
    log_test "id -> i mapping" "FAIL"
fi

if grep -q '"name".*"n"' internal/toon/encoder.go 2>/dev/null; then
    log_test "name -> n mapping" "PASS"
else
    log_test "name -> n mapping" "FAIL"
fi

if grep -q '"status".*"s"' internal/toon/encoder.go 2>/dev/null; then
    log_test "status -> s mapping" "PASS"
else
    log_test "status -> s mapping" "FAIL"
fi

# Test 6: Value abbreviations
echo ""
echo "[6] Value Abbreviations"
if grep -q 'case "healthy"' internal/toon/encoder.go 2>/dev/null && grep -q 'return "H"' internal/toon/encoder.go 2>/dev/null; then
    log_test "healthy -> H abbreviation" "PASS"
else
    log_test "healthy -> H abbreviation" "FAIL"
fi

if grep -q 'case "pending"' internal/toon/encoder.go 2>/dev/null && grep -q 'return "P"' internal/toon/encoder.go 2>/dev/null; then
    log_test "pending -> P abbreviation" "PASS"
else
    log_test "pending -> P abbreviation" "FAIL"
fi

if grep -q 'case "completed"' internal/toon/encoder.go 2>/dev/null && grep -q 'return "C"' internal/toon/encoder.go 2>/dev/null; then
    log_test "completed -> C abbreviation" "PASS"
else
    log_test "completed -> C abbreviation" "FAIL"
fi

# Test 7: Transport layer
echo ""
echo "[7] Transport Layer"
if grep -q "type Transport struct" internal/toon/transport.go 2>/dev/null; then
    log_test "Transport struct defined" "PASS"
else
    log_test "Transport struct defined" "FAIL"
fi

if grep -q "func NewTransport" internal/toon/transport.go 2>/dev/null; then
    log_test "NewTransport function" "PASS"
else
    log_test "NewTransport function" "FAIL"
fi

if grep -q "func.*Transport.*Do" internal/toon/transport.go 2>/dev/null; then
    log_test "Do method" "PASS"
else
    log_test "Do method" "FAIL"
fi

if grep -q "func.*Transport.*Get" internal/toon/transport.go 2>/dev/null; then
    log_test "Get method" "PASS"
else
    log_test "Get method" "FAIL"
fi

if grep -q "func.*Transport.*Post" internal/toon/transport.go 2>/dev/null; then
    log_test "Post method" "PASS"
else
    log_test "Post method" "FAIL"
fi

# Test 8: Middleware
echo ""
echo "[8] Middleware"
if grep -q "type Middleware struct" internal/toon/transport.go 2>/dev/null; then
    log_test "Middleware struct defined" "PASS"
else
    log_test "Middleware struct defined" "FAIL"
fi

if grep -q "func NewMiddleware" internal/toon/transport.go 2>/dev/null; then
    log_test "NewMiddleware function" "PASS"
else
    log_test "NewMiddleware function" "FAIL"
fi

if grep -q "func.*Middleware.*Handler" internal/toon/transport.go 2>/dev/null; then
    log_test "Handler method" "PASS"
else
    log_test "Handler method" "FAIL"
fi

# Test 9: Metrics
echo ""
echo "[9] Metrics"
if grep -q "TransportMetrics" internal/toon/transport.go 2>/dev/null; then
    log_test "TransportMetrics struct" "PASS"
else
    log_test "TransportMetrics struct" "FAIL"
fi

if grep -q "BytesSaved" internal/toon/transport.go 2>/dev/null; then
    log_test "BytesSaved metric" "PASS"
else
    log_test "BytesSaved metric" "FAIL"
fi

if grep -q "TokensSaved" internal/toon/transport.go 2>/dev/null; then
    log_test "TokensSaved metric" "PASS"
else
    log_test "TokensSaved metric" "FAIL"
fi

if grep -q "CompressionRatio" internal/toon/transport.go 2>/dev/null; then
    log_test "CompressionRatio metric" "PASS"
else
    log_test "CompressionRatio metric" "FAIL"
fi

# Test 10: Content type
echo ""
echo "[10] Content Type"
if grep -q "application/toon+json" internal/toon/transport.go 2>/dev/null; then
    log_test "TOON content type" "PASS"
else
    log_test "TOON content type" "FAIL"
fi

# Test 11: Unit tests
echo ""
echo "[11] Unit Tests"
if go test -v ./internal/toon/... -count=1 2>&1 | grep -q "PASS"; then
    log_test "TOON unit tests pass" "PASS"
else
    log_test "TOON unit tests pass" "FAIL"
fi

# Test 12: Token counting
echo ""
echo "[12] Token Utilities"
if grep -q "TokenCount" internal/toon/encoder.go 2>/dev/null; then
    log_test "TokenCount function" "PASS"
else
    log_test "TokenCount function" "FAIL"
fi

if grep -q "CompressionRatio" internal/toon/encoder.go 2>/dev/null; then
    log_test "CompressionRatio method" "PASS"
else
    log_test "CompressionRatio method" "FAIL"
fi

echo ""
echo "=============================================="
echo "  Results: $PASSED/$TOTAL tests passed"
echo "=============================================="

if [ $FAILED -gt 0 ]; then
    echo -e "\e[31m$FAILED test(s) failed\e[0m"
    exit 1
else
    echo -e "\e[32mAll tests passed!\e[0m"
    exit 0
fi
