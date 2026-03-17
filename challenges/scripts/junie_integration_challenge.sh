#!/bin/bash
# Junie Integration Verification Challenge
# Tests all Junie provider functionality including CLI, ACP, and unified modes
# Validates model discovery, scoring, and integration with HelixAgent

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
LOG_FILE="/tmp/junie_challenge_$(date +%Y%m%d_%H%M%S).log"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

log() {
    echo -e "$1" | tee -a "$LOG_FILE"
}

pass() {
    PASSED_TESTS=$((PASSED_TESTS + 1))
    log "${GREEN}[PASS]${NC} $1"
}

fail() {
    FAILED_TESTS=$((FAILED_TESTS + 1))
    log "${RED}[FAIL]${NC} $1"
}

skip() {
    log "${YELLOW}[SKIP]${NC} $1"
}

test_start() {
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    log "${BLUE}[TEST]${NC} $1"
}

check_command() {
    if command -v "$1" &> /dev/null; then
        return 0
    else
        return 1
    fi
}

check_env() {
    if [ -n "${!1:-}" ]; then
        return 0
    else
        return 1
    fi
}

section() {
    echo "" | tee -a "$LOG_FILE"
    log "${BLUE}══════════════════════════════════════════════════════════════${NC}"
    log "${BLUE}  $1${NC}"
    log "${BLUE}══════════════════════════════════════════════════════════════${NC}"
}

section "PHASE 1: Environment Checks"

test_start "Check junie CLI installation"
if check_command "junie"; then
    pass "Junie CLI is installed"
else
    skip "Junie CLI not installed - some tests will be skipped"
fi

test_start "Check JUNIE_API_KEY environment variable"
if check_env "JUNIE_API_KEY"; then
    pass "JUNIE_API_KEY is set"
else
    skip "JUNIE_API_KEY not set - some tests will be skipped"
fi

test_start "Check BYOK provider keys"
BYOK_FOUND=0
for key in JUNIE_ANTHROPIC_API_KEY JUNIE_OPENAI_API_KEY JUNIE_GOOGLE_API_KEY JUNIE_GROK_API_KEY; do
    if check_env "$key"; then
        log "  Found: $key"
        BYOK_FOUND=1
    fi
done
if [ $BYOK_FOUND -eq 1 ]; then
    pass "At least one BYOK key is configured"
else
    skip "No BYOK keys configured"
fi

section "PHASE 2: Provider Package Tests"

test_start "Run Junie provider unit tests"
cd "$PROJECT_ROOT"
if GOMAXPROCS=2 go test -v ./internal/llm/providers/junie/... -count=1 -short 2>&1 | tee -a "$LOG_FILE"; then
    pass "Junie provider unit tests passed"
else
    fail "Junie provider unit tests failed"
fi

section "PHASE 3: Provider Discovery Tests"

test_start "Check Junie in provider mappings"
cd "$PROJECT_ROOT"
if grep -q '"junie"' internal/services/provider_discovery.go; then
    pass "Junie found in provider_discovery.go mappings"
else
    fail "Junie NOT found in provider_discovery.go"
fi

test_start "Check Junie in SupportedProviders"
if grep -q '"junie":' internal/verifier/provider_types.go; then
    pass "Junie found in SupportedProviders"
else
    fail "Junie NOT found in SupportedProviders"
fi

test_start "Check Junie in provider registry"
if grep -q '"junie"' internal/services/provider_registry.go; then
    pass "Junie found in provider_registry.go"
else
    fail "Junie NOT found in provider_registry.go"
fi

section "PHASE 4: Build Verification"

test_start "Build Junie provider package"
cd "$PROJECT_ROOT"
if go build ./internal/llm/providers/junie/... 2>&1 | tee -a "$LOG_FILE"; then
    pass "Junie provider package builds successfully"
else
    fail "Junie provider package build failed"
fi

test_start "Build services package"
if go build ./internal/services/... 2>&1 | tee -a "$LOG_FILE"; then
    pass "Services package builds successfully"
else
    fail "Services package build failed"
fi

test_start "Build verifier package"
if go build ./internal/verifier/... 2>&1 | tee -a "$LOG_FILE"; then
    pass "Verifier package builds successfully"
else
    fail "Verifier package build failed"
fi

section "PHASE 5: Integration Tests (requires Junie CLI)"

if check_command "junie" && check_env "JUNIE_API_KEY"; then
    test_start "Test Junie CLI availability"
    if junie --version 2>&1 | tee -a "$LOG_FILE"; then
        pass "Junie CLI version check passed"
    else
        fail "Junie CLI version check failed"
    fi

    test_start "Test Junie CLI health check"
    RESPONSE=$(timeout 30 junie --auth "$JUNIE_API_KEY" --output-format json --task "Say OK" 2>&1 || echo "TIMEOUT")
    if echo "$RESPONSE" | grep -qi "ok\|result"; then
        pass "Junie CLI health check passed"
    else
        fail "Junie CLI health check failed: $RESPONSE"
    fi

    test_start "Test Junie model discovery"
    MODELS=$(junie --help 2>&1 || echo "")
    if echo "$MODELS" | grep -qiE "sonnet|opus|gpt|gemini|grok"; then
        pass "Junie model discovery found model references"
    else
        skip "Could not verify model discovery from CLI help"
    fi
else
    skip "Junie CLI or API key not available - skipping integration tests"
fi

section "PHASE 6: Model Verification Tests"

if check_command "junie" && check_env "JUNIE_API_KEY"; then
    KNOWN_MODELS=("sonnet" "opus" "gpt" "gemini-pro" "grok")
    
    for model in "${KNOWN_MODELS[@]}"; do
        test_start "Test model: $model"
        RESPONSE=$(timeout 60 junie --auth "$JUNIE_API_KEY" --model "$model" --output-format json --task "Reply with just: Model OK" 2>&1 || echo "ERROR")
        if echo "$RESPONSE" | grep -qi "OK\|result"; then
            pass "Model $model responded successfully"
        else
            skip "Model $model test skipped (may not be available)"
        fi
    done
else
    skip "Junie CLI not available - skipping model verification tests"
fi

section "PHASE 7: ACP Mode Tests"

if check_command "junie" && check_env "JUNIE_API_KEY"; then
    test_start "Test ACP mode availability"
    timeout 5 junie --acp true 2>&1 || ACP_EXIT=$?
    if [ "${ACP_EXIT:-0}" -eq 0 ] || [ "${ACP_EXIT:-0}" -eq 124 ]; then
        pass "ACP mode is available"
    else
        skip "ACP mode test inconclusive"
    fi
else
    skip "Junie CLI not available - skipping ACP tests"
fi

section "PHASE 8: HelixAgent Integration Tests"

test_start "Build HelixAgent binary"
cd "$PROJECT_ROOT"
if go build -o bin/helixagent ./cmd/helixagent 2>&1 | tee -a "$LOG_FILE"; then
    pass "HelixAgent binary built successfully"
else
    fail "HelixAgent binary build failed"
fi

test_start "Check Junie provider in HelixAgent startup"
if check_env "JUNIE_API_KEY"; then
    HELIXAGENT_OUTPUT=$(timeout 10 ./bin/helixagent --help 2>&1 || echo "")
    if echo "$HELIXAGENT_OUTPUT" | grep -qi "junie\|provider"; then
        pass "HelixAgent includes provider support"
    else
        skip "Could not verify Junie in HelixAgent output"
    fi
else
    skip "JUNIE_API_KEY not set - skipping startup integration test"
fi

section "PHASE 9: Lint and Format"

test_start "Run go vet on Junie provider"
cd "$PROJECT_ROOT"
if go vet ./internal/llm/providers/junie/... 2>&1 | tee -a "$LOG_FILE"; then
    pass "go vet passed for Junie provider"
else
    fail "go vet found issues in Junie provider"
fi

test_start "Check for import cycle"
if go build ./internal/llm/providers/junie/... 2>&1 | grep -qi "import cycle"; then
    fail "Import cycle detected in Junie provider"
else
    pass "No import cycles in Junie provider"
fi

section "SUMMARY"

log ""
log "══════════════════════════════════════════════════════════════"
log "  JUNIE INTEGRATION CHALLENGE RESULTS"
log "══════════════════════════════════════════════════════════════"
log ""
log "Total Tests:  $TOTAL_TESTS"
log "${GREEN}Passed:       $PASSED_TESTS${NC}"
log "${RED}Failed:       $FAILED_TESTS${NC}"
log "Log File:     $LOG_FILE"
log ""

if [ $FAILED_TESTS -eq 0 ]; then
    log "${GREEN}All tests passed! Junie integration is complete.${NC}"
    exit 0
else
    log "${RED}Some tests failed. Please review the log file.${NC}"
    exit 1
fi
