#!/bin/bash
# HelixAgent Plugin Transport Challenge
# Tests HTTP/3, TOON protocol, and Brotli compression for CLI agent plugins
# 25 tests total

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASSED=0
FAILED=0
TOTAL=25

echo "========================================"
echo "HelixAgent Plugin Transport Challenge"
echo "========================================"
echo ""

# Test function
run_test() {
    local test_name="$1"
    local test_cmd="$2"

    echo -n "Testing: $test_name... "
    if eval "$test_cmd" > /dev/null 2>&1; then
        echo -e "${GREEN}PASSED${NC}"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}FAILED${NC}"
        FAILED=$((FAILED + 1))
    fi
}

# Test 1-5: Transport library structure
echo "--- Transport Library Structure ---"
run_test "Go transport exists" "test -f '$PROJECT_ROOT/plugins/packages/transport/go/transport.go'"
run_test "Go TOON encoder exists" "test -f '$PROJECT_ROOT/plugins/packages/transport/go/toon.go'"
run_test "TypeScript transport exists" "test -f '$PROJECT_ROOT/plugins/packages/transport/typescript/transport.ts'"
run_test "Transport exports Protocol type" "grep -q 'export type Protocol' '$PROJECT_ROOT/plugins/packages/transport/typescript/transport.ts'"
run_test "Transport exports ContentType" "grep -q 'export type ContentType' '$PROJECT_ROOT/plugins/packages/transport/typescript/transport.ts'"

# Test 6-10: HTTP/3 support
echo ""
echo "--- HTTP/3 Support ---"
run_test "Go supports HTTP/3" "grep -q 'http3' '$PROJECT_ROOT/plugins/packages/transport/go/transport.go'"
run_test "TS protocol negotiation" "grep -q 'negotiateProtocol' '$PROJECT_ROOT/plugins/packages/transport/typescript/transport.ts'"
run_test "Protocol fallback chain" "grep -q 'h2.*http/1.1' '$PROJECT_ROOT/plugins/packages/transport/typescript/transport.ts'"
run_test "Health endpoint check" "grep -q '/health' '$PROJECT_ROOT/plugins/packages/transport/typescript/transport.ts'"
run_test "Connection timeout" "grep -q 'timeout' '$PROJECT_ROOT/plugins/packages/transport/typescript/transport.ts'"

# Test 11-15: TOON protocol
echo ""
echo "--- TOON Protocol ---"
run_test "TOON abbreviations defined" "grep -q 'TOON_ABBREVIATIONS' '$PROJECT_ROOT/plugins/packages/transport/typescript/transport.ts'"
run_test "Content abbreviation" "grep -q \"'content': 'c'\" '$PROJECT_ROOT/plugins/packages/transport/typescript/transport.ts'"
run_test "Role abbreviation" "grep -q \"'role': 'r'\" '$PROJECT_ROOT/plugins/packages/transport/typescript/transport.ts'"
run_test "Messages abbreviation" "grep -q \"'messages': 'm'\" '$PROJECT_ROOT/plugins/packages/transport/typescript/transport.ts'"
run_test "TOON encoder function" "grep -q 'encodeTOON' '$PROJECT_ROOT/plugins/packages/transport/typescript/transport.ts'"

# Test 16-20: Brotli compression
echo ""
echo "--- Brotli Compression ---"
run_test "Brotli support flag" "grep -q 'enableBrotli' '$PROJECT_ROOT/plugins/packages/transport/typescript/transport.ts'"
run_test "Accept-Encoding header" "grep -q 'Accept-Encoding' '$PROJECT_ROOT/plugins/packages/transport/typescript/transport.ts'"
run_test "Compression fallback" "grep -q 'gzip' '$PROJECT_ROOT/plugins/packages/transport/typescript/transport.ts'"
run_test "Content-Type header" "grep -q 'Content-Type' '$PROJECT_ROOT/plugins/packages/transport/typescript/transport.ts'"
run_test "TOON content type" "grep -q 'application/toon+json' '$PROJECT_ROOT/plugins/packages/transport/typescript/transport.ts'"

# Test 21-25: Go transport implementation
echo ""
echo "--- Go Transport Implementation ---"
run_test "Go HelixTransport interface" "grep -q 'type HelixTransport interface' '$PROJECT_ROOT/plugins/packages/transport/go/transport.go'"
run_test "Go Connect method" "grep -q 'Connect(' '$PROJECT_ROOT/plugins/packages/transport/go/transport.go'"
run_test "Go Do method" "grep -q 'Do(' '$PROJECT_ROOT/plugins/packages/transport/go/transport.go'"
run_test "Go Stream method" "grep -q 'Stream(' '$PROJECT_ROOT/plugins/packages/transport/go/transport.go'"
run_test "Go TOON encoding" "grep -q 'EncodeTOON' '$PROJECT_ROOT/plugins/packages/transport/go/toon.go'"

# Summary
echo ""
echo "========================================"
echo "Transport Challenge Results"
echo "========================================"
echo -e "Passed: ${GREEN}$PASSED${NC}/$TOTAL"
echo -e "Failed: ${RED}$FAILED${NC}/$TOTAL"

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All transport tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some transport tests failed${NC}"
    exit 1
fi
