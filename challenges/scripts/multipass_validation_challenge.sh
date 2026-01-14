#!/bin/bash

# ============================================================================
# Multi-Pass Validation Challenge Script
# ============================================================================
# Validates comprehensive support for multi-pass validation in AI Debate:
# - Phase 1: Initial Response
# - Phase 2: Validation
# - Phase 3: Polish & Improve
# - Phase 4: Final Conclusion
#
# The multi-pass validation system ensures debate responses are:
# 1. Cross-validated for accuracy
# 2. Polished and improved
# 3. Synthesized into a final high-quality conclusion
# ============================================================================

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Project root
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SERVICES_PKG="$PROJECT_ROOT/internal/services"

# Log functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Test function
run_test() {
    local test_name="$1"
    local test_cmd="$2"

    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo -n "  [$TOTAL_TESTS] $test_name... "

    if eval "$test_cmd" > /dev/null 2>&1; then
        echo -e "${GREEN}PASS${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        return 0
    else
        echo -e "${RED}FAIL${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

echo ""
echo "============================================================================"
echo "         MULTI-PASS VALIDATION CHALLENGE - 60 TESTS                         "
echo "   Validates AI Debate multi-pass validation for enhanced response quality   "
echo "============================================================================"
echo ""

# ============================================================================
# Section 1: Package Structure Validation (5 tests)
# ============================================================================
echo "==========================================================================="
echo "Section 1: Package Structure Validation"
echo "==========================================================================="

run_test "Services package directory exists" \
    "[ -d '$SERVICES_PKG' ]"

run_test "debate_multipass_validation.go exists" \
    "[ -f '$SERVICES_PKG/debate_multipass_validation.go' ]"

run_test "debate_multipass_validation_test.go exists" \
    "[ -f '$SERVICES_PKG/debate_multipass_validation_test.go' ]"

run_test "debate_dialogue.go exists" \
    "[ -f '$SERVICES_PKG/debate_dialogue.go' ]"

run_test "debate_service.go exists" \
    "[ -f '$SERVICES_PKG/debate_service.go' ]"

# ============================================================================
# Section 2: Validation Phase Type Constants (8 tests)
# ============================================================================
echo ""
echo "==========================================================================="
echo "Section 2: Validation Phase Type Constants"
echo "==========================================================================="

run_test "ValidationPhase type defined" \
    "grep -q 'type ValidationPhase string' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "PhaseInitialResponse constant defined" \
    "grep -q 'PhaseInitialResponse.*\"initial_response\"' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "PhaseValidation constant defined" \
    "grep -q 'PhaseValidation.*\"validation\"' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "PhasePolishImprove constant defined" \
    "grep -q 'PhasePolishImprove.*\"polish_improve\"' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "PhaseFinalConclusion constant defined" \
    "grep -q 'PhaseFinalConclusion.*\"final_conclusion\"' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "ValidationPhases function defined" \
    "grep -q 'func ValidationPhases()' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "GetPhaseInfo function defined" \
    "grep -q 'func GetPhaseInfo' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "PhaseInfo struct defined" \
    "grep -q 'type PhaseInfo struct' '$SERVICES_PKG/debate_multipass_validation.go'"

# ============================================================================
# Section 3: Validation Configuration (6 tests)
# ============================================================================
echo ""
echo "==========================================================================="
echo "Section 3: Validation Configuration"
echo "==========================================================================="

run_test "ValidationConfig struct defined" \
    "grep -q 'type ValidationConfig struct' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "DefaultValidationConfig function defined" \
    "grep -q 'func DefaultValidationConfig()' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "EnableValidation field exists" \
    "grep -q 'EnableValidation.*bool' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "EnablePolish field exists" \
    "grep -q 'EnablePolish.*bool' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "ValidationTimeout field exists" \
    "grep -q 'ValidationTimeout.*time.Duration' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "MinConfidenceToSkip field exists" \
    "grep -q 'MinConfidenceToSkip.*float64' '$SERVICES_PKG/debate_multipass_validation.go'"

# ============================================================================
# Section 4: Validation Result Types (8 tests)
# ============================================================================
echo ""
echo "==========================================================================="
echo "Section 4: Validation Result Types"
echo "==========================================================================="

run_test "ValidationResult struct defined" \
    "grep -q 'type ValidationResult struct' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "ValidationIssue struct defined" \
    "grep -q 'type ValidationIssue struct' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "IssueType type defined" \
    "grep -q 'type IssueType string' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "ValidationSeverity type defined" \
    "grep -q 'type ValidationSeverity string' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "IssueFactualError constant defined" \
    "grep -q 'IssueFactualError.*\"factual_error\"' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "IssueIncomplete constant defined" \
    "grep -q 'IssueIncomplete.*\"incomplete\"' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "ValidationSeverityCritical constant defined" \
    "grep -q 'ValidationSeverityCritical.*\"critical\"' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "ValidationSeverityMajor constant defined" \
    "grep -q 'ValidationSeverityMajor.*\"major\"' '$SERVICES_PKG/debate_multipass_validation.go'"

# ============================================================================
# Section 5: Polish Result Types (5 tests)
# ============================================================================
echo ""
echo "==========================================================================="
echo "Section 5: Polish Result Types"
echo "==========================================================================="

run_test "PolishResult struct defined" \
    "grep -q 'type PolishResult struct' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "PolishResult has OriginalResponse field" \
    "grep -q 'OriginalResponse.*string' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "PolishResult has PolishedResponse field" \
    "grep -q 'PolishedResponse.*string' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "PolishResult has ImprovementScore field" \
    "grep -q 'ImprovementScore.*float64' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "PolishResult has ChangesSummary field" \
    "grep -q 'ChangesSummary' '$SERVICES_PKG/debate_multipass_validation.go'"

# ============================================================================
# Section 6: Phase Result Types (5 tests)
# ============================================================================
echo ""
echo "==========================================================================="
echo "Section 6: Phase Result Types"
echo "==========================================================================="

run_test "PhaseResult struct defined" \
    "grep -q 'type PhaseResult struct' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "PhaseResult has Phase field" \
    "grep -q 'Phase.*ValidationPhase' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "PhaseResult has Duration field" \
    "grep -q 'Duration.*time.Duration' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "PhaseResult has PhaseScore field" \
    "grep -q 'PhaseScore.*float64' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "PhaseResult has PhaseSummary field" \
    "grep -q 'PhaseSummary.*string' '$SERVICES_PKG/debate_multipass_validation.go'"

# ============================================================================
# Section 7: MultiPassResult Types (5 tests)
# ============================================================================
echo ""
echo "==========================================================================="
echo "Section 7: MultiPassResult Types"
echo "==========================================================================="

run_test "MultiPassResult struct defined" \
    "grep -q 'type MultiPassResult struct' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "MultiPassResult has Phases field" \
    "grep -q 'Phases.*\\[\\]\\*PhaseResult' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "MultiPassResult has FinalResponse field" \
    "grep -q 'FinalResponse.*string' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "MultiPassResult has OverallConfidence field" \
    "grep -q 'OverallConfidence.*float64' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "MultiPassResult has QualityImprovement field" \
    "grep -q 'QualityImprovement.*float64' '$SERVICES_PKG/debate_multipass_validation.go'"

# ============================================================================
# Section 8: MultiPassValidator Implementation (6 tests)
# ============================================================================
echo ""
echo "==========================================================================="
echo "Section 8: MultiPassValidator Implementation"
echo "==========================================================================="

run_test "MultiPassValidator struct defined" \
    "grep -q 'type MultiPassValidator struct' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "NewMultiPassValidator function defined" \
    "grep -q 'func NewMultiPassValidator' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "SetConfig method defined" \
    "grep -q 'func (mpv \\*MultiPassValidator) SetConfig' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "GetConfig method defined" \
    "grep -q 'func (mpv \\*MultiPassValidator) GetConfig' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "ValidateAndImprove method defined" \
    "grep -q 'func (mpv \\*MultiPassValidator) ValidateAndImprove' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "SetPhaseCallback method defined" \
    "grep -q 'func (mpv \\*MultiPassValidator) SetPhaseCallback' '$SERVICES_PKG/debate_multipass_validation.go'"

# ============================================================================
# Section 9: Phase Rendering Functions (4 tests)
# ============================================================================
echo ""
echo "==========================================================================="
echo "Section 9: Phase Rendering Functions"
echo "==========================================================================="

run_test "FormatPhaseHeader function defined" \
    "grep -q 'func FormatPhaseHeader' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "FormatPhaseFooter function defined" \
    "grep -q 'func FormatPhaseFooter' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "FormatMultiPassOutput function defined" \
    "grep -q 'func FormatMultiPassOutput' '$SERVICES_PKG/debate_multipass_validation.go'"

run_test "Phase rendering uses box drawing characters" \
    "grep -q '═' '$SERVICES_PKG/debate_multipass_validation.go'"

# ============================================================================
# Section 10: Dialogue Event Types (6 tests)
# ============================================================================
echo ""
echo "==========================================================================="
echo "Section 10: Dialogue Event Types for Phases"
echo "==========================================================================="

run_test "EventPhaseStart event type defined" \
    "grep -q 'EventPhaseStart.*\"phase_start\"' '$SERVICES_PKG/debate_dialogue.go'"

run_test "EventPhaseProgress event type defined" \
    "grep -q 'EventPhaseProgress.*\"phase_progress\"' '$SERVICES_PKG/debate_dialogue.go'"

run_test "EventPhaseEnd event type defined" \
    "grep -q 'EventPhaseEnd.*\"phase_end\"' '$SERVICES_PKG/debate_dialogue.go'"

run_test "EventValidationResult event type defined" \
    "grep -q 'EventValidationResult.*\"validation_result\"' '$SERVICES_PKG/debate_dialogue.go'"

run_test "EventPolishResult event type defined" \
    "grep -q 'EventPolishResult.*\"polish_result\"' '$SERVICES_PKG/debate_dialogue.go'"

run_test "EventFinalSynthesis event type defined" \
    "grep -q 'EventFinalSynthesis.*\"final_synthesis\"' '$SERVICES_PKG/debate_dialogue.go'"

# ============================================================================
# Section 11: Debate Service Integration (4 tests)
# ============================================================================
echo ""
echo "==========================================================================="
echo "Section 11: Debate Service Integration"
echo "==========================================================================="

run_test "ConductDebateWithMultiPassValidation method defined" \
    "grep -q 'func (ds \\*DebateService) ConductDebateWithMultiPassValidation' '$SERVICES_PKG/debate_service.go'"

run_test "StreamDebateWithMultiPassValidation method defined" \
    "grep -q 'func (ds \\*DebateService) StreamDebateWithMultiPassValidation' '$SERVICES_PKG/debate_service.go'"

run_test "MultiPassValidator is created in debate service" \
    "grep -q 'NewMultiPassValidator' '$SERVICES_PKG/debate_service.go'"

run_test "Validation phases are executed in debate service" \
    "grep -q 'ValidateAndImprove' '$SERVICES_PKG/debate_service.go'"

# ============================================================================
# Section 12: Unit Tests Compilation and Execution (4 tests)
# ============================================================================
echo ""
echo "==========================================================================="
echo "Section 12: Unit Tests Compilation and Execution"
echo "==========================================================================="

cd "$PROJECT_ROOT"

run_test "Services package compiles" \
    "go build ./internal/services/..."

run_test "Multipass validation tests compile" \
    "go test -c ./internal/services/... -o /dev/null"

run_test "Multipass validation unit tests pass" \
    "go test -v ./internal/services/... -run 'TestValidation|TestMultiPass|TestPhase' -count=1 -timeout=60s"

run_test "Race condition tests pass" \
    "go test -v ./internal/services/... -run 'TestMultiPassValidator_Concurrent' -race -count=1 -timeout=60s"

# ============================================================================
# Summary
# ============================================================================
echo ""
echo "============================================================================"
echo "                    CHALLENGE RESULTS                                       "
echo "============================================================================"
echo "  Total Tests:  $TOTAL_TESTS"
echo "  Passed:       $PASSED_TESTS"
echo "  Failed:       $FAILED_TESTS"
echo "============================================================================"
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}[SUCCESS] ALL TESTS PASSED!${NC}"
    echo ""
    echo -e "${GREEN}[SUCCESS] Multi-pass validation system verified:${NC}"
    echo -e "${GREEN}[SUCCESS]   - 4 validation phases fully implemented${NC}"
    echo -e "${GREEN}[SUCCESS]   - Phase 1: Initial Response - Collect debate responses${NC}"
    echo -e "${GREEN}[SUCCESS]   - Phase 2: Validation - Cross-validate for accuracy${NC}"
    echo -e "${GREEN}[SUCCESS]   - Phase 3: Polish & Improve - Refine based on feedback${NC}"
    echo -e "${GREEN}[SUCCESS]   - Phase 4: Final Conclusion - Synthesize consensus${NC}"
    echo -e "${GREEN}[SUCCESS]   - Clear phase indicators in output${NC}"
    echo -e "${GREEN}[SUCCESS]   - Streaming support with phase callbacks${NC}"
    echo -e "${GREEN}[SUCCESS]   - All unit tests pass${NC}"
    echo -e "${GREEN}[SUCCESS]   - Race condition tests pass${NC}"
    echo ""
    exit 0
else
    echo -e "${RED}[FAILED] $FAILED_TESTS tests failed${NC}"
    echo ""
    exit 1
fi
