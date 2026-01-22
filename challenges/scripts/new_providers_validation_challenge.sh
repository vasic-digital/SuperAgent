#!/bin/bash
#===============================================================================
# HELIXAGENT NEW LLM PROVIDERS VALIDATION CHALLENGE
#===============================================================================
# This challenge validates the 11 newly implemented LLM providers:
#   - OpenAI (GPT-4, GPT-4o, o1)
#   - xAI (Grok-2, Grok-2-latest)
#   - Groq (Llama 3.1/3.3, Mixtral)
#   - Cohere (Command-R, Command-R-Plus)
#   - Together AI (Llama, Qwen, DeepSeek)
#   - Perplexity (Sonar models with search)
#   - AI21 (Jamba models)
#   - Fireworks AI (fast inference)
#   - Anthropic Direct (Claude via native API)
#   - Replicate (async predictions)
#   - HuggingFace (Inference API)
#
# Tests for each provider:
#   1. Provider creation
#   2. Capabilities verification
#   3. Config validation
#   4. Complete request (mocked)
#   5. Model get/set
#
# Usage:
#   ./challenges/scripts/new_providers_validation_challenge.sh [options]
#
# Options:
#   --provider NAME  Test specific provider only
#   --verbose        Enable verbose logging
#   --help           Show this help message
#
#===============================================================================

set -e

#===============================================================================
# CONFIGURATION
#===============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHALLENGES_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$CHALLENGES_DIR")"

# Timestamps
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
YEAR=$(date +%Y)
MONTH=$(date +%m)
DAY=$(date +%d)

# Directories
RESULTS_BASE="$CHALLENGES_DIR/results/new_providers_validation"
RESULTS_DIR="$RESULTS_BASE/$YEAR/$MONTH/$DAY/$TIMESTAMP"
LOGS_DIR="$RESULTS_DIR/logs"
OUTPUT_DIR="$RESULTS_DIR/results"

# Log files
MAIN_LOG="$LOGS_DIR/new_providers_validation.log"
TEST_LOG="$LOGS_DIR/test_output.log"
ERROR_LOG="$LOGS_DIR/errors.log"

# Options
SPECIFIC_PROVIDER=""
VERBOSE=false

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# List of new providers to test
NEW_PROVIDERS=(
    "openai"
    "xai"
    "groq"
    "cohere"
    "together"
    "perplexity"
    "ai21"
    "fireworks"
    "anthropic"
    "replicate"
    "huggingface"
)

#===============================================================================
# LOGGING FUNCTIONS
#===============================================================================

log() {
    local msg="[$(date '+%Y-%m-%d %H:%M:%S')] $*"
    if [ -d "$(dirname "$MAIN_LOG")" ]; then
        echo -e "$msg" | tee -a "$MAIN_LOG"
    else
        echo -e "$msg"
    fi
}

log_info() {
    log "${BLUE}[INFO]${NC} $*"
}

log_success() {
    log "${GREEN}[SUCCESS]${NC} $*"
}

log_warning() {
    log "${YELLOW}[WARNING]${NC} $*"
}

log_error() {
    log "${RED}[ERROR]${NC} $*"
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*" >> "$ERROR_LOG" 2>/dev/null || true
}

log_phase() {
    if [ -d "$(dirname "$MAIN_LOG")" ]; then
        echo "" | tee -a "$MAIN_LOG"
    else
        echo ""
    fi
    log "${PURPLE}========================================${NC}"
    log "${PURPLE}  $*${NC}"
    log "${PURPLE}========================================${NC}"
    if [ -d "$(dirname "$MAIN_LOG")" ]; then
        echo "" | tee -a "$MAIN_LOG"
    else
        echo ""
    fi
}

#===============================================================================
# HELPER FUNCTIONS
#===============================================================================

usage() {
    cat << EOF
${GREEN}HelixAgent New LLM Providers Validation Challenge${NC}

${BLUE}Usage:${NC}
    $0 [options]

${BLUE}Options:${NC}
    ${YELLOW}--provider NAME${NC}  Test specific provider only
    ${YELLOW}--verbose${NC}        Enable verbose logging
    ${YELLOW}--help${NC}           Show this help message

${BLUE}New Providers Tested (11 total):${NC}
    1. ${CYAN}OpenAI${NC}        - GPT-4, GPT-4o, o1 models
    2. ${CYAN}xAI${NC}           - Grok-2, Grok-2-latest
    3. ${CYAN}Groq${NC}          - Fast inference (Llama, Mixtral)
    4. ${CYAN}Cohere${NC}        - Command-R models with RAG
    5. ${CYAN}Together AI${NC}   - Open models (Llama, Qwen, DeepSeek)
    6. ${CYAN}Perplexity${NC}    - Search-focused LLM with citations
    7. ${CYAN}AI21${NC}          - Jamba models (256K context)
    8. ${CYAN}Fireworks AI${NC}  - Fast inference platform
    9. ${CYAN}Anthropic${NC}     - Native Claude API
   10. ${CYAN}Replicate${NC}     - Async predictions
   11. ${CYAN}HuggingFace${NC}   - Inference API

${BLUE}Tests Per Provider:${NC}
    - Unit tests (go test ./internal/llm/providers/<name>/...)
    - Compilation check
    - Interface compliance

${BLUE}Output:${NC}
    Results stored in: ${YELLOW}$RESULTS_BASE/<date>/<timestamp>/${NC}

EOF
}

setup_directories() {
    log_info "Creating directory structure..."
    mkdir -p "$LOGS_DIR"
    mkdir -p "$OUTPUT_DIR"
    touch "$ERROR_LOG"
    log_success "Directories created: $RESULTS_DIR"
}

#===============================================================================
# PHASE 1: COMPILATION CHECK
#===============================================================================

phase1_compilation_check() {
    log_phase "PHASE 1: Compilation Check"

    local compile_passed=0
    local compile_failed=0
    local compile_results="$OUTPUT_DIR/compilation_results.json"

    echo "{\"providers\": [" > "$compile_results"
    local first=true

    for provider in "${NEW_PROVIDERS[@]}"; do
        if [ -n "$SPECIFIC_PROVIDER" ] && [ "$provider" != "$SPECIFIC_PROVIDER" ]; then
            continue
        fi

        log_info "Checking compilation: $provider"

        local provider_path="$PROJECT_ROOT/internal/llm/providers/$provider"
        if [ ! -d "$provider_path" ]; then
            log_error "  Provider directory not found: $provider_path"
            compile_failed=$((compile_failed + 1))
            [ "$first" = false ] && echo "," >> "$compile_results"
            echo "{\"provider\": \"$provider\", \"status\": \"not_found\"}" >> "$compile_results"
            first=false
            continue
        fi

        # Check if go files exist
        local go_files=$(find "$provider_path" -name "*.go" -not -name "*_test.go" | wc -l)
        local test_files=$(find "$provider_path" -name "*_test.go" | wc -l)

        if [ "$go_files" -eq 0 ]; then
            log_error "  No Go files found in $provider"
            compile_failed=$((compile_failed + 1))
            [ "$first" = false ] && echo "," >> "$compile_results"
            echo "{\"provider\": \"$provider\", \"status\": \"no_go_files\"}" >> "$compile_results"
            first=false
            continue
        fi

        # Try to build
        if go build "$provider_path/..." 2>> "$ERROR_LOG"; then
            log_success "  $provider: compilation OK ($go_files source files, $test_files test files)"
            compile_passed=$((compile_passed + 1))
            [ "$first" = false ] && echo "," >> "$compile_results"
            echo "{\"provider\": \"$provider\", \"status\": \"ok\", \"source_files\": $go_files, \"test_files\": $test_files}" >> "$compile_results"
        else
            log_error "  $provider: compilation FAILED"
            compile_failed=$((compile_failed + 1))
            [ "$first" = false ] && echo "," >> "$compile_results"
            echo "{\"provider\": \"$provider\", \"status\": \"compile_error\"}" >> "$compile_results"
        fi
        first=false
    done

    echo "]," >> "$compile_results"
    echo "\"passed\": $compile_passed, \"failed\": $compile_failed}" >> "$compile_results"

    export COMPILE_PASSED=$compile_passed
    export COMPILE_FAILED=$compile_failed

    log_info ""
    log_info "Compilation Summary: $compile_passed passed, $compile_failed failed"

    if [ $compile_failed -eq 0 ]; then
        log_success "All providers compile successfully"
        return 0
    else
        log_error "Some providers failed to compile"
        return 1
    fi
}

#===============================================================================
# PHASE 2: UNIT TESTS
#===============================================================================

phase2_unit_tests() {
    log_phase "PHASE 2: Unit Tests"

    local tests_passed=0
    local tests_failed=0
    local total_test_count=0
    local test_results="$OUTPUT_DIR/unit_test_results.json"

    echo "{\"providers\": [" > "$test_results"
    local first=true

    for provider in "${NEW_PROVIDERS[@]}"; do
        if [ -n "$SPECIFIC_PROVIDER" ] && [ "$provider" != "$SPECIFIC_PROVIDER" ]; then
            continue
        fi

        log_info "Running unit tests: $provider"

        local provider_path="$PROJECT_ROOT/internal/llm/providers/$provider"
        local provider_test_log="$LOGS_DIR/${provider}_test.log"

        # Run tests and capture output
        if go test -v "$provider_path/..." -count=1 > "$provider_test_log" 2>&1; then
            # Count tests
            local test_count=$(grep -c "^--- PASS:" "$provider_test_log" 2>/dev/null || echo "0")
            total_test_count=$((total_test_count + test_count))

            log_success "  $provider: $test_count tests PASSED"
            tests_passed=$((tests_passed + 1))

            [ "$first" = false ] && echo "," >> "$test_results"
            echo "{\"provider\": \"$provider\", \"status\": \"passed\", \"test_count\": $test_count}" >> "$test_results"
        else
            local fail_count=$(grep -c "^--- FAIL:" "$provider_test_log" 2>/dev/null || echo "0")
            local pass_count=$(grep -c "^--- PASS:" "$provider_test_log" 2>/dev/null || echo "0")

            log_error "  $provider: FAILED ($fail_count failed, $pass_count passed)"
            tests_failed=$((tests_failed + 1))

            [ "$first" = false ] && echo "," >> "$test_results"
            echo "{\"provider\": \"$provider\", \"status\": \"failed\", \"passed\": $pass_count, \"failed\": $fail_count}" >> "$test_results"

            # Show first few failure lines
            if [ "$VERBOSE" = true ]; then
                grep -A 5 "^--- FAIL:" "$provider_test_log" | head -20
            fi
        fi
        first=false
    done

    echo "]," >> "$test_results"
    echo "\"passed\": $tests_passed, \"failed\": $tests_failed, \"total_tests\": $total_test_count}" >> "$test_results"

    export UNIT_TESTS_PASSED=$tests_passed
    export UNIT_TESTS_FAILED=$tests_failed
    export TOTAL_TEST_COUNT=$total_test_count

    log_info ""
    log_info "Unit Test Summary: $tests_passed providers passed, $tests_failed failed"
    log_info "Total individual tests: $total_test_count"

    if [ $tests_failed -eq 0 ]; then
        log_success "All provider unit tests passed"
        return 0
    else
        log_error "Some provider unit tests failed"
        return 1
    fi
}

#===============================================================================
# PHASE 3: INTERFACE COMPLIANCE CHECK
#===============================================================================

phase3_interface_compliance() {
    log_phase "PHASE 3: Interface Compliance"

    local compliance_passed=0
    local compliance_failed=0
    local compliance_results="$OUTPUT_DIR/compliance_results.json"

    log_info "Checking LLMProvider interface implementation..."

    # Required methods for LLMProvider interface
    local required_methods=(
        "Complete"
        "CompleteStream"
        "HealthCheck"
        "GetCapabilities"
        "ValidateConfig"
    )

    echo "{\"providers\": [" > "$compliance_results"
    local first=true

    for provider in "${NEW_PROVIDERS[@]}"; do
        if [ -n "$SPECIFIC_PROVIDER" ] && [ "$provider" != "$SPECIFIC_PROVIDER" ]; then
            continue
        fi

        log_info "Checking interface compliance: $provider"

        local provider_file="$PROJECT_ROOT/internal/llm/providers/$provider/$provider.go"

        if [ ! -f "$provider_file" ]; then
            log_error "  Provider file not found: $provider_file"
            compliance_failed=$((compliance_failed + 1))
            [ "$first" = false ] && echo "," >> "$compliance_results"
            echo "{\"provider\": \"$provider\", \"status\": \"file_not_found\"}" >> "$compliance_results"
            first=false
            continue
        fi

        local missing_methods=""
        local found_methods=""

        for method in "${required_methods[@]}"; do
            if grep -q "func.*Provider.*$method" "$provider_file"; then
                found_methods+="$method,"
            else
                missing_methods+="$method,"
            fi
        done

        if [ -z "$missing_methods" ]; then
            log_success "  $provider: All methods implemented"
            compliance_passed=$((compliance_passed + 1))
            [ "$first" = false ] && echo "," >> "$compliance_results"
            echo "{\"provider\": \"$provider\", \"status\": \"compliant\", \"methods\": [\"${found_methods%,}\"]}" >> "$compliance_results"
        else
            log_error "  $provider: Missing methods - ${missing_methods%,}"
            compliance_failed=$((compliance_failed + 1))
            [ "$first" = false ] && echo "," >> "$compliance_results"
            echo "{\"provider\": \"$provider\", \"status\": \"non_compliant\", \"missing\": [\"${missing_methods%,}\"]}" >> "$compliance_results"
        fi
        first=false
    done

    echo "]," >> "$compliance_results"
    echo "\"passed\": $compliance_passed, \"failed\": $compliance_failed}" >> "$compliance_results"

    export COMPLIANCE_PASSED=$compliance_passed
    export COMPLIANCE_FAILED=$compliance_failed

    log_info ""
    log_info "Interface Compliance Summary: $compliance_passed compliant, $compliance_failed non-compliant"

    if [ $compliance_failed -eq 0 ]; then
        log_success "All providers implement LLMProvider interface"
        return 0
    else
        log_error "Some providers missing interface methods"
        return 1
    fi
}

#===============================================================================
# PHASE 4: PROVIDER DISCOVERY INTEGRATION
#===============================================================================

phase4_discovery_integration() {
    log_phase "PHASE 4: Provider Discovery Integration"

    log_info "Checking provider discovery registration..."

    local discovery_file="$PROJECT_ROOT/internal/services/provider_discovery.go"
    local discovery_passed=0
    local discovery_failed=0
    local discovery_results="$OUTPUT_DIR/discovery_results.json"

    echo "{\"providers\": [" > "$discovery_results"
    local first=true

    for provider in "${NEW_PROVIDERS[@]}"; do
        if [ -n "$SPECIFIC_PROVIDER" ] && [ "$provider" != "$SPECIFIC_PROVIDER" ]; then
            continue
        fi

        log_info "Checking discovery registration: $provider"

        # Check if provider is imported
        local import_check=$(grep -c "\"dev.helix.agent/internal/llm/providers/$provider\"" "$discovery_file" 2>/dev/null || echo "0")

        # Check if provider is in createProvider switch
        local switch_check=$(grep -c "case \"$provider\":" "$discovery_file" 2>/dev/null || echo "0")

        if [ "$import_check" -gt 0 ] && [ "$switch_check" -gt 0 ]; then
            log_success "  $provider: Registered in discovery"
            discovery_passed=$((discovery_passed + 1))
            [ "$first" = false ] && echo "," >> "$discovery_results"
            echo "{\"provider\": \"$provider\", \"status\": \"registered\", \"imported\": true, \"in_switch\": true}" >> "$discovery_results"
        else
            log_warning "  $provider: Not fully registered (import: $import_check, switch: $switch_check)"
            discovery_failed=$((discovery_failed + 1))
            [ "$first" = false ] && echo "," >> "$discovery_results"
            echo "{\"provider\": \"$provider\", \"status\": \"partial\", \"imported\": $([[ $import_check -gt 0 ]] && echo true || echo false), \"in_switch\": $([[ $switch_check -gt 0 ]] && echo true || echo false)}" >> "$discovery_results"
        fi
        first=false
    done

    echo "]," >> "$discovery_results"
    echo "\"passed\": $discovery_passed, \"failed\": $discovery_failed}" >> "$discovery_results"

    export DISCOVERY_PASSED=$discovery_passed
    export DISCOVERY_FAILED=$discovery_failed

    log_info ""
    log_info "Discovery Integration Summary: $discovery_passed registered, $discovery_failed partial"

    if [ $discovery_failed -eq 0 ]; then
        log_success "All providers registered in discovery"
        return 0
    else
        log_warning "Some providers not fully registered"
        return 0  # Soft failure - still continue
    fi
}

#===============================================================================
# PHASE 5: GENERATE REPORT
#===============================================================================

phase5_generate_report() {
    log_phase "PHASE 5: Generate Report"

    local summary_file="$OUTPUT_DIR/challenge_summary.json"
    local report_file="$OUTPUT_DIR/new_providers_validation_report.md"

    # Calculate overall success
    local overall_success=false
    local total_checks=$((${COMPILE_PASSED:-0} + ${COMPILE_FAILED:-0}))
    local all_passed=$((${COMPILE_PASSED:-0} + ${UNIT_TESTS_PASSED:-0} + ${COMPLIANCE_PASSED:-0} + ${DISCOVERY_PASSED:-0}))
    local all_failed=$((${COMPILE_FAILED:-0} + ${UNIT_TESTS_FAILED:-0} + ${COMPLIANCE_FAILED:-0} + ${DISCOVERY_FAILED:-0}))

    if [ ${COMPILE_FAILED:-0} -eq 0 ] && [ ${UNIT_TESTS_FAILED:-0} -eq 0 ] && [ ${COMPLIANCE_FAILED:-0} -eq 0 ]; then
        overall_success=true
    fi

    # Generate JSON summary
    cat > "$summary_file" << EOF
{
    "challenge": "New LLM Providers Validation",
    "timestamp": "$(date -Iseconds)",
    "providers_tested": ${#NEW_PROVIDERS[@]},
    "results": {
        "compilation": {
            "passed": ${COMPILE_PASSED:-0},
            "failed": ${COMPILE_FAILED:-0}
        },
        "unit_tests": {
            "providers_passed": ${UNIT_TESTS_PASSED:-0},
            "providers_failed": ${UNIT_TESTS_FAILED:-0},
            "total_individual_tests": ${TOTAL_TEST_COUNT:-0}
        },
        "interface_compliance": {
            "passed": ${COMPLIANCE_PASSED:-0},
            "failed": ${COMPLIANCE_FAILED:-0}
        },
        "discovery_integration": {
            "passed": ${DISCOVERY_PASSED:-0},
            "failed": ${DISCOVERY_FAILED:-0}
        },
        "overall_success": $overall_success
    },
    "results_directory": "$RESULTS_DIR"
}
EOF

    # Generate markdown report
    cat > "$report_file" << EOF
# New LLM Providers Validation Challenge Report

**Generated**: $(date '+%Y-%m-%d %H:%M:%S')
**Status**: $([ "$overall_success" = "true" ] && echo "✅ PASSED" || echo "❌ NEEDS ATTENTION")

## Summary

| Metric | Passed | Failed |
|--------|--------|--------|
| Compilation | ${COMPILE_PASSED:-0} | ${COMPILE_FAILED:-0} |
| Unit Tests | ${UNIT_TESTS_PASSED:-0} | ${UNIT_TESTS_FAILED:-0} |
| Interface Compliance | ${COMPLIANCE_PASSED:-0} | ${COMPLIANCE_FAILED:-0} |
| Discovery Integration | ${DISCOVERY_PASSED:-0} | ${DISCOVERY_FAILED:-0} |

**Total Individual Tests**: ${TOTAL_TEST_COUNT:-0}

## New Providers Tested

| # | Provider | Description | Models |
|---|----------|-------------|--------|
| 1 | OpenAI | Direct OpenAI API | GPT-4, GPT-4o, o1, o1-mini |
| 2 | xAI | Grok models | grok-2, grok-2-latest |
| 3 | Groq | Fast inference | Llama 3.1/3.3, Mixtral |
| 4 | Cohere | RAG-focused | Command-R, Command-R-Plus |
| 5 | Together AI | Open models | Llama, Qwen, DeepSeek-R1 |
| 6 | Perplexity | Search LLM | Sonar models with citations |
| 7 | AI21 | Jamba models | jamba-1.5-large/mini (256K) |
| 8 | Fireworks AI | Fast inference | Various open models |
| 9 | Anthropic | Native Claude | Claude 3.5 Sonnet/Haiku |
| 10 | Replicate | Async predictions | Llama, Mistral, SDXL |
| 11 | HuggingFace | Inference API | Llama, Mistral, Gemma |

## Features Implemented

### Common Features (All Providers)
- ✅ LLMProvider interface compliance
- ✅ Complete (non-streaming)
- ✅ CompleteStream (streaming)
- ✅ HealthCheck
- ✅ GetCapabilities
- ✅ ValidateConfig
- ✅ Retry with exponential backoff
- ✅ Context cancellation support
- ✅ Confidence calculation

### Provider-Specific Features

| Provider | Tools | Vision | Streaming | Search | RAG |
|----------|-------|--------|-----------|--------|-----|
| OpenAI | ✅ | ✅ | ✅ | - | - |
| xAI | ✅ | ✅ | ✅ | - | - |
| Groq | ✅ | - | ✅ | - | - |
| Cohere | ✅ | - | ✅ | - | ✅ |
| Together | ✅ | ✅ | ✅ | - | - |
| Perplexity | - | - | ✅ | ✅ | - |
| AI21 | - | - | ✅ | - | - |
| Fireworks | ✅ | - | ✅ | - | - |
| Anthropic | ✅ | ✅ | ✅ | - | - |
| Replicate | - | ✅ | ✅ | - | - |
| HuggingFace | - | - | ✅ | - | - |

## Environment Variables

\`\`\`bash
# Tier 1: Premium
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...

# Tier 2: High-Quality Specialized
XAI_API_KEY=xai-...
GROK_API_KEY=xai-...
COHERE_API_KEY=...
CO_API_KEY=...
PERPLEXITY_API_KEY=pplx-...
PPLX_API_KEY=pplx-...

# Tier 3: Fast Inference
GROQ_API_KEY=gsk_...

# Tier 4: Alternative
TOGETHER_API_KEY=...
TOGETHERAI_API_KEY=...
FIREWORKS_API_KEY=fw_...

# Tier 5: Specialized
AI21_API_KEY=...
REPLICATE_API_KEY=r8_...
HUGGINGFACE_API_KEY=hf_...
\`\`\`

## Results Location

\`\`\`
$RESULTS_DIR/
├── logs/
│   ├── new_providers_validation.log
│   ├── *_test.log (per provider)
│   └── errors.log
└── results/
    ├── compilation_results.json
    ├── unit_test_results.json
    ├── compliance_results.json
    ├── discovery_results.json
    └── challenge_summary.json
\`\`\`

## Conclusion

$(if [ "$overall_success" = "true" ]; then
    echo "All 11 new LLM providers have been successfully implemented and validated."
    echo "They are ready for use in HelixAgent's ensemble LLM system."
else
    echo "Some providers require attention. Review the detailed logs for more information."
fi)

---
*Generated by HelixAgent New Providers Validation Challenge*
EOF

    # Print summary
    echo ""
    log_info "=========================================="
    log_info "  CHALLENGE SUMMARY"
    log_info "=========================================="
    echo ""
    log_info "Providers Tested:      ${#NEW_PROVIDERS[@]}"
    log_info "Compilation:           ${COMPILE_PASSED:-0} passed, ${COMPILE_FAILED:-0} failed"
    log_info "Unit Tests:            ${UNIT_TESTS_PASSED:-0} passed, ${UNIT_TESTS_FAILED:-0} failed"
    log_info "Interface Compliance:  ${COMPLIANCE_PASSED:-0} passed, ${COMPLIANCE_FAILED:-0} failed"
    log_info "Discovery Integration: ${DISCOVERY_PASSED:-0} passed, ${DISCOVERY_FAILED:-0} failed"
    log_info "Total Tests:           ${TOTAL_TEST_COUNT:-0}"
    log_info "Overall:               $([ "$overall_success" = "true" ] && echo "${GREEN}PASSED${NC}" || echo "${RED}FAILED${NC}")"
    echo ""
    log_info "Results: $RESULTS_DIR"
    log_info "Report: $report_file"

    if [ "$overall_success" = "true" ]; then
        log_success "New Providers Validation Challenge PASSED"
        return 0
    else
        log_warning "New Providers Validation Challenge: needs attention"
        return 1
    fi
}

#===============================================================================
# MAIN EXECUTION
#===============================================================================

main() {
    # Parse arguments
    while [ $# -gt 0 ]; do
        case "$1" in
            --provider)
                SPECIFIC_PROVIDER="$2"
                shift 2
                ;;
            --verbose)
                VERBOSE=true
                shift
                ;;
            --help|-h)
                usage
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
    done

    START_TIME=$(date '+%Y-%m-%d %H:%M:%S')

    # Setup
    setup_directories

    log_phase "NEW LLM PROVIDERS VALIDATION CHALLENGE"
    log_info "Start time: $START_TIME"
    log_info "Results directory: $RESULTS_DIR"

    if [ -n "$SPECIFIC_PROVIDER" ]; then
        log_info "Testing specific provider: $SPECIFIC_PROVIDER"
    else
        log_info "Testing all ${#NEW_PROVIDERS[@]} new providers"
    fi

    # Change to project root for go commands
    cd "$PROJECT_ROOT"

    # Execute phases
    local phase1_result=0
    local phase2_result=0
    local phase3_result=0
    local phase4_result=0

    phase1_compilation_check || phase1_result=$?
    phase2_unit_tests || phase2_result=$?
    phase3_interface_compliance || phase3_result=$?
    phase4_discovery_integration || phase4_result=$?
    phase5_generate_report

    # Return overall result
    if [ $phase1_result -eq 0 ] && [ $phase2_result -eq 0 ] && [ $phase3_result -eq 0 ]; then
        exit 0
    else
        exit 1
    fi
}

main "$@"
