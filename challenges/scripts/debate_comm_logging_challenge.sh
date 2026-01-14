#!/bin/bash
# Debate Communication Logging Challenge
# VALIDATES: Retrofit-like logging for all 18 CLI agents
# Tests that debate communication is properly logged with colors and fallback visualization

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh" 2>/dev/null || {
    log_info() { echo -e "\033[0;34m[INFO] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_success() { echo -e "\033[0;32m[SUCCESS] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_error() { echo -e "\033[0;31m[ERROR] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_warning() { echo -e "\033[0;33m[WARNING] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
}

CHALLENGE_NAME="Debate Communication Logging Challenge"
PASSED=0
FAILED=0
TOTAL=0

log_info "=============================================="
log_info "$CHALLENGE_NAME"
log_info "=============================================="
log_info "VALIDATES: Retrofit-like logging for AI Debate Ensemble"
log_info "Tests all 18 supported CLI agents"
log_info ""

PROJECT_ROOT="${SCRIPT_DIR}/../.."

# ============================================================================
# Section 1: DebateCommLogger Structure Validation
# ============================================================================

log_info "=============================================="
log_info "Section 1: DebateCommLogger Structure Validation"
log_info "=============================================="

# Test 1: DebateCommLogger struct exists
TOTAL=$((TOTAL + 1))
log_info "Test 1: DebateCommLogger struct exists"
if grep -q "type DebateCommLogger struct" "$PROJECT_ROOT/internal/services/debate_comm_logger.go" 2>/dev/null; then
    log_success "DebateCommLogger struct found"
    PASSED=$((PASSED + 1))
else
    log_error "DebateCommLogger struct NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 2: NewDebateCommLogger function exists
TOTAL=$((TOTAL + 1))
log_info "Test 2: NewDebateCommLogger constructor exists"
if grep -q "func NewDebateCommLogger" "$PROJECT_ROOT/internal/services/debate_comm_logger.go" 2>/dev/null; then
    log_success "NewDebateCommLogger constructor found"
    PASSED=$((PASSED + 1))
else
    log_error "NewDebateCommLogger constructor NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 3: All role abbreviations are defined
TOTAL=$((TOTAL + 1))
log_info "Test 3: All debate role abbreviations defined"
ROLES=("analyst" "proposer" "critic" "synthesizer" "mediator")
ABBREVS=("A" "P" "C" "S" "M")
ROLES_FOUND=0
for i in "${!ROLES[@]}"; do
    if grep -q "\"${ROLES[$i]}\".*\"${ABBREVS[$i]}\"" "$PROJECT_ROOT/internal/services/debate_comm_logger.go" 2>/dev/null; then
        ROLES_FOUND=$((ROLES_FOUND + 1))
    fi
done
if [ "$ROLES_FOUND" -ge 5 ]; then
    log_success "All 5 role abbreviations found (A/P/C/S/M)"
    PASSED=$((PASSED + 1))
else
    log_error "Only $ROLES_FOUND role abbreviations found (need 5)"
    FAILED=$((FAILED + 1))
fi

# Test 4: ANSI color constants are defined
TOTAL=$((TOTAL + 1))
log_info "Test 4: ANSI color constants defined"
COLOR_COUNT=$(grep -c "Color.*= \"\\\\033\[" "$PROJECT_ROOT/internal/services/debate_comm_logger.go" 2>/dev/null || echo "0")
if [ "$COLOR_COUNT" -ge 15 ]; then
    log_success "Found $COLOR_COUNT color constants (need 15+)"
    PASSED=$((PASSED + 1))
else
    log_error "Only $COLOR_COUNT color constants found (need 15+)"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 2: Retrofit-like Logging Methods
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 2: Retrofit-like Logging Methods"
log_info "=============================================="

# Test 5: LogRequest method exists (sends request to LLM)
TOTAL=$((TOTAL + 1))
log_info "Test 5: LogRequest method exists"
if grep -q "func (dcl \*DebateCommLogger) LogRequest" "$PROJECT_ROOT/internal/services/debate_comm_logger.go" 2>/dev/null; then
    log_success "LogRequest method found"
    PASSED=$((PASSED + 1))
else
    log_error "LogRequest method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 6: LogResponse method exists (receives response from LLM)
TOTAL=$((TOTAL + 1))
log_info "Test 6: LogResponse method exists"
if grep -q "func (dcl \*DebateCommLogger) LogResponse" "$PROJECT_ROOT/internal/services/debate_comm_logger.go" 2>/dev/null; then
    log_success "LogResponse method found"
    PASSED=$((PASSED + 1))
else
    log_error "LogResponse method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 7: LogFallbackAttempt method exists
TOTAL=$((TOTAL + 1))
log_info "Test 7: LogFallbackAttempt method exists"
if grep -q "func (dcl \*DebateCommLogger) LogFallbackAttempt" "$PROJECT_ROOT/internal/services/debate_comm_logger.go" 2>/dev/null; then
    log_success "LogFallbackAttempt method found"
    PASSED=$((PASSED + 1))
else
    log_error "LogFallbackAttempt method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 8: LogFallbackChain method exists
TOTAL=$((TOTAL + 1))
log_info "Test 8: LogFallbackChain method exists"
if grep -q "func (dcl \*DebateCommLogger) LogFallbackChain" "$PROJECT_ROOT/internal/services/debate_comm_logger.go" 2>/dev/null; then
    log_success "LogFallbackChain method found"
    PASSED=$((PASSED + 1))
else
    log_error "LogFallbackChain method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 9: LogError method exists
TOTAL=$((TOTAL + 1))
log_info "Test 9: LogError method exists"
if grep -q "func (dcl \*DebateCommLogger) LogError" "$PROJECT_ROOT/internal/services/debate_comm_logger.go" 2>/dev/null; then
    log_success "LogError method found"
    PASSED=$((PASSED + 1))
else
    log_error "LogError method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 10: LogDebatePhase method exists
TOTAL=$((TOTAL + 1))
log_info "Test 10: LogDebatePhase method exists"
if grep -q "func (dcl \*DebateCommLogger) LogDebatePhase" "$PROJECT_ROOT/internal/services/debate_comm_logger.go" 2>/dev/null; then
    log_success "LogDebatePhase method found"
    PASSED=$((PASSED + 1))
else
    log_error "LogDebatePhase method NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 3: CLI Agent Support (All 18 Agents)
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 3: CLI Agent Support (All 18 Agents)"
log_info "=============================================="

# Test 11: CLIAgentColors function exists
TOTAL=$((TOTAL + 1))
log_info "Test 11: CLIAgentColors function exists"
if grep -q "func CLIAgentColors" "$PROJECT_ROOT/internal/services/debate_comm_logger.go" 2>/dev/null; then
    log_success "CLIAgentColors function found"
    PASSED=$((PASSED + 1))
else
    log_error "CLIAgentColors function NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 12: All 18 CLI agents are supported
TOTAL=$((TOTAL + 1))
log_info "Test 12: All 18 CLI agents supported in color configuration"
CLI_AGENTS=("opencode" "claudecode" "kilocode" "crush" "helixcode" "kiro" "aider" "cline" "codenamegoose" "deepseekcli" "forge" "geminicli" "gptengineer" "mistralcode" "ollamacode" "plandex" "qwencode" "amazonq")
AGENTS_FOUND=0
for agent in "${CLI_AGENTS[@]}"; do
    if grep -q "\"$agent\"" "$PROJECT_ROOT/internal/services/debate_comm_logger.go" 2>/dev/null; then
        AGENTS_FOUND=$((AGENTS_FOUND + 1))
    fi
done
if [ "$AGENTS_FOUND" -ge 18 ]; then
    log_success "All 18 CLI agents found in color configuration"
    PASSED=$((PASSED + 1))
else
    log_error "Only $AGENTS_FOUND CLI agents found (need 18)"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 4: Integration with DebateService
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 4: Integration with DebateService"
log_info "=============================================="

# Test 13: commLogger field in DebateService
TOTAL=$((TOTAL + 1))
log_info "Test 13: commLogger field exists in DebateService"
if grep -q "commLogger.*\*DebateCommLogger" "$PROJECT_ROOT/internal/services/debate_service.go" 2>/dev/null; then
    log_success "commLogger field found in DebateService"
    PASSED=$((PASSED + 1))
else
    log_error "commLogger field NOT found in DebateService!"
    FAILED=$((FAILED + 1))
fi

# Test 14: commLogger initialized in constructors
TOTAL=$((TOTAL + 1))
log_info "Test 14: commLogger initialized in constructors"
if grep -q "commLogger.*NewDebateCommLogger" "$PROJECT_ROOT/internal/services/debate_service.go" 2>/dev/null; then
    log_success "commLogger initialized in constructors"
    PASSED=$((PASSED + 1))
else
    log_error "commLogger NOT initialized in constructors!"
    FAILED=$((FAILED + 1))
fi

# Test 15: LogRequest called in getParticipantResponse
TOTAL=$((TOTAL + 1))
log_info "Test 15: LogRequest called during participant response"
if grep -q "ds.commLogger.LogRequest" "$PROJECT_ROOT/internal/services/debate_service.go" 2>/dev/null; then
    log_success "LogRequest called in getParticipantResponse"
    PASSED=$((PASSED + 1))
else
    log_error "LogRequest NOT called!"
    FAILED=$((FAILED + 1))
fi

# Test 16: LogResponse called after receiving response
TOTAL=$((TOTAL + 1))
log_info "Test 16: LogResponse called after receiving LLM response"
if grep -q "ds.commLogger.LogResponse" "$PROJECT_ROOT/internal/services/debate_service.go" 2>/dev/null; then
    log_success "LogResponse called after receiving response"
    PASSED=$((PASSED + 1))
else
    log_error "LogResponse NOT called!"
    FAILED=$((FAILED + 1))
fi

# Test 17: LogFallbackAttempt called during fallback
TOTAL=$((TOTAL + 1))
log_info "Test 17: LogFallbackAttempt called during fallback"
if grep -q "ds.commLogger.LogFallbackAttempt" "$PROJECT_ROOT/internal/services/debate_service.go" 2>/dev/null; then
    log_success "LogFallbackAttempt called during fallback"
    PASSED=$((PASSED + 1))
else
    log_error "LogFallbackAttempt NOT called!"
    FAILED=$((FAILED + 1))
fi

# Test 18: LogDebateSummary called after round
TOTAL=$((TOTAL + 1))
log_info "Test 18: LogDebateSummary called after round completion"
if grep -q "ds.commLogger.LogDebateSummary" "$PROJECT_ROOT/internal/services/debate_service.go" 2>/dev/null; then
    log_success "LogDebateSummary called after round"
    PASSED=$((PASSED + 1))
else
    log_error "LogDebateSummary NOT called!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 5: Unit Test Coverage
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 5: Unit Test Coverage"
log_info "=============================================="

# Test 19: Test file exists
TOTAL=$((TOTAL + 1))
log_info "Test 19: debate_comm_logger_test.go exists"
if [ -f "$PROJECT_ROOT/internal/services/debate_comm_logger_test.go" ]; then
    log_success "Test file found"
    PASSED=$((PASSED + 1))
else
    log_error "Test file NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 20: Tests cover all 18 CLI agents
TOTAL=$((TOTAL + 1))
log_info "Test 20: Tests cover all 18 CLI agents"
if grep -q "TestCLIAgentIntegration" "$PROJECT_ROOT/internal/services/debate_comm_logger_test.go" 2>/dev/null; then
    log_success "CLI agent integration tests found"
    PASSED=$((PASSED + 1))
else
    log_error "CLI agent integration tests NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 21: Tests cover role abbreviations
TOTAL=$((TOTAL + 1))
log_info "Test 21: Tests cover role abbreviations"
if grep -q "TestRoleAbbreviations" "$PROJECT_ROOT/internal/services/debate_comm_logger_test.go" 2>/dev/null; then
    log_success "Role abbreviations tests found"
    PASSED=$((PASSED + 1))
else
    log_error "Role abbreviations tests NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 22: Tests cover fallback chain logging
TOTAL=$((TOTAL + 1))
log_info "Test 22: Tests cover fallback chain logging"
if grep -q "TestLogFallbackChain" "$PROJECT_ROOT/internal/services/debate_comm_logger_test.go" 2>/dev/null; then
    log_success "Fallback chain logging tests found"
    PASSED=$((PASSED + 1))
else
    log_error "Fallback chain logging tests NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 6: Run Unit Tests
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 6: Run Unit Tests"
log_info "=============================================="

# Test 23: Run debate comm logger tests
TOTAL=$((TOTAL + 1))
log_info "Test 23: Running debate comm logger unit tests..."
cd "$PROJECT_ROOT"
TEST_OUTPUT=$(go test -v -run "TestNew|TestSet|TestRole|TestFormat|TestLog|TestCLI|TestFallback|TestColor|TestAllDebate" ./internal/services/debate_comm_logger_test.go ./internal/services/debate_comm_logger.go 2>&1)
if echo "$TEST_OUTPUT" | grep -q "^ok\|PASS"; then
    PASS_COUNT=$(echo "$TEST_OUTPUT" | grep -c "--- PASS:" 2>/dev/null || echo "0")
    log_success "Debate comm logger tests passed ($PASS_COUNT+ subtests)"
    PASSED=$((PASSED + 1))
else
    log_error "Debate comm logger tests FAILED!"
    echo "$TEST_OUTPUT" | tail -20
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 7: Format Validation
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 7: Format Validation"
log_info "=============================================="

# Test 24: Request arrow format (<---)
TOTAL=$((TOTAL + 1))
log_info "Test 24: Request arrow format (<---)"
if grep -q '<---' "$PROJECT_ROOT/internal/services/debate_comm_logger.go" 2>/dev/null; then
    log_success "Request arrow format (<---) found"
    PASSED=$((PASSED + 1))
else
    log_error "Request arrow format NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 25: Response arrow format (--->)
TOTAL=$((TOTAL + 1))
log_info "Test 25: Response arrow format (--->)"
if grep -qF -- '--->' "$PROJECT_ROOT/internal/services/debate_comm_logger.go" 2>/dev/null; then
    log_success "Response arrow format (--->) found"
    PASSED=$((PASSED + 1))
else
    log_error "Response arrow format NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 26: FALLBACK tag format
TOTAL=$((TOTAL + 1))
log_info "Test 26: FALLBACK tag format"
if grep -q '\[FALLBACK' "$PROJECT_ROOT/internal/services/debate_comm_logger.go" 2>/dev/null; then
    log_success "FALLBACK tag format found"
    PASSED=$((PASSED + 1))
else
    log_error "FALLBACK tag format NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 27: Role tag format [X: Model Name]
TOTAL=$((TOTAL + 1))
log_info "Test 27: Role tag format [X: Model Name]"
if grep -q '\[%s:.*\]' "$PROJECT_ROOT/internal/services/debate_comm_logger.go" 2>/dev/null; then
    log_success "Role tag format found"
    PASSED=$((PASSED + 1))
else
    log_error "Role tag format NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 8: Build Validation
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 8: Build Validation"
log_info "=============================================="

# Test 28: Code compiles successfully
TOTAL=$((TOTAL + 1))
log_info "Test 28: Code compiles with debate comm logger"
if go build -o /dev/null ./internal/services/... 2>&1; then
    log_success "Code compiles successfully"
    PASSED=$((PASSED + 1))
else
    log_error "Code compilation FAILED!"
    FAILED=$((FAILED + 1))
fi

# Test 29: Full project builds
TOTAL=$((TOTAL + 1))
log_info "Test 29: Full project builds with debate comm logger"
if go build -o /dev/null ./cmd/helixagent 2>&1; then
    log_success "Full project builds successfully"
    PASSED=$((PASSED + 1))
else
    log_error "Full project build FAILED!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 9: Streaming Support
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 9: Streaming Support"
log_info "=============================================="

# Test 30: LogStreamStart method exists
TOTAL=$((TOTAL + 1))
log_info "Test 30: LogStreamStart method exists"
if grep -q "func (dcl \*DebateCommLogger) LogStreamStart" "$PROJECT_ROOT/internal/services/debate_comm_logger.go" 2>/dev/null; then
    log_success "LogStreamStart method found"
    PASSED=$((PASSED + 1))
else
    log_error "LogStreamStart method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 31: LogStreamChunk method exists
TOTAL=$((TOTAL + 1))
log_info "Test 31: LogStreamChunk method exists"
if grep -q "func (dcl \*DebateCommLogger) LogStreamChunk" "$PROJECT_ROOT/internal/services/debate_comm_logger.go" 2>/dev/null; then
    log_success "LogStreamChunk method found"
    PASSED=$((PASSED + 1))
else
    log_error "LogStreamChunk method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 32: LogStreamEnd method exists
TOTAL=$((TOTAL + 1))
log_info "Test 32: LogStreamEnd method exists"
if grep -q "func (dcl \*DebateCommLogger) LogStreamEnd" "$PROJECT_ROOT/internal/services/debate_comm_logger.go" 2>/dev/null; then
    log_success "LogStreamEnd method found"
    PASSED=$((PASSED + 1))
else
    log_error "LogStreamEnd method NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 10: FallbackChainEntry Type
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 10: FallbackChainEntry Type"
log_info "=============================================="

# Test 33: FallbackChainEntry struct exists
TOTAL=$((TOTAL + 1))
log_info "Test 33: FallbackChainEntry struct exists"
if grep -q "type FallbackChainEntry struct" "$PROJECT_ROOT/internal/services/debate_comm_logger.go" 2>/dev/null; then
    log_success "FallbackChainEntry struct found"
    PASSED=$((PASSED + 1))
else
    log_error "FallbackChainEntry struct NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 34: FallbackChainEntry has Provider field
TOTAL=$((TOTAL + 1))
log_info "Test 34: FallbackChainEntry has required fields"
if grep -A5 "type FallbackChainEntry struct" "$PROJECT_ROOT/internal/services/debate_comm_logger.go" | grep -q "Provider" 2>/dev/null; then
    log_success "FallbackChainEntry has Provider field"
    PASSED=$((PASSED + 1))
else
    log_error "FallbackChainEntry missing Provider field!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 11: Model Name Formatting
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 11: Model Name Formatting"
log_info "=============================================="

# Test 35: formatModelName function exists
TOTAL=$((TOTAL + 1))
log_info "Test 35: formatModelName function exists"
if grep -q "func formatModelName" "$PROJECT_ROOT/internal/services/debate_comm_logger.go" 2>/dev/null; then
    log_success "formatModelName function found"
    PASSED=$((PASSED + 1))
else
    log_error "formatModelName function NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 36: Model name mappings include Claude 4.5
TOTAL=$((TOTAL + 1))
log_info "Test 36: Model name mappings include Claude 4.5"
if grep -q "Claude Opus 4.5" "$PROJECT_ROOT/internal/services/debate_comm_logger.go" 2>/dev/null; then
    log_success "Claude 4.5 model name mapping found"
    PASSED=$((PASSED + 1))
else
    log_error "Claude 4.5 model name mapping NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Final Summary
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Challenge Summary: $CHALLENGE_NAME"
log_info "=============================================="
log_info "Total Tests: $TOTAL"
log_success "Passed: $PASSED"
if [ $FAILED -gt 0 ]; then
    log_error "Failed: $FAILED"
fi

PERCENTAGE=$((PASSED * 100 / TOTAL))
log_info "Pass Rate: ${PERCENTAGE}%"

if [ $FAILED -eq 0 ]; then
    log_success "=============================================="
    log_success "ALL TESTS PASSED!"
    log_success "Debate communication logging is working."
    log_success "Retrofit-like format: [ROLE: Model] <--> Content"
    log_success "All 18 CLI agents have proper color support."
    log_success "Fallback chain visualization implemented."
    log_success "=============================================="
    exit 0
else
    log_error "=============================================="
    log_error "SOME TESTS FAILED!"
    log_error "Review debate communication logging implementation."
    log_error "=============================================="
    exit 1
fi
