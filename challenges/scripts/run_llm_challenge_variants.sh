#!/bin/bash
#===============================================================================
# HELIXAGENT LLM CHALLENGE VARIANTS RUNNER
#===============================================================================
# Runs LLM challenges in BOTH OAuth and non-OAuth modes to ensure
# full compatibility with different authentication methods.
#
# Usage:
#   ./run_llm_challenge_variants.sh <challenge_name>
#   ./run_llm_challenge_variants.sh ai_debate_formation
#   ./run_llm_challenge_variants.sh content_generation_challenge
#   ./run_llm_challenge_variants.sh --all  # Run all LLM challenges in both modes
#
# OAuth Mode:
#   - Uses Claude Code OAuth credentials (sk-ant-oat01-*)
#   - CLAUDE_CODE_USE_OAUTH_CREDENTIALS=true
#   - Primary provider: claude-oauth
#
# Non-OAuth Mode:
#   - Uses API key-based providers (DeepSeek, Mistral, Gemini, etc.)
#   - CLAUDE_CODE_USE_OAUTH_CREDENTIALS=false
#   - Primary providers: deepseek, mistral, gemini (based on verification scores)
#
#===============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHALLENGES_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$CHALLENGES_DIR")"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# Load environment
if [ -f "$PROJECT_ROOT/.env" ]; then
    set -a
    source "$PROJECT_ROOT/.env"
    set +a
fi

# LLM Challenges that need both OAuth and non-OAuth testing
LLM_CHALLENGES=(
    "ai_debate_formation"
    "ai_debate_workflow"
    "content_generation_challenge"
    "cli_agents_challenge"
    "opencode"
    "opencode_init"
    "single_provider_challenge"
    "ensemble_voting"
    "embeddings_service"
    "streaming_responses"
)

# Results tracking
OAUTH_PASSED=0
OAUTH_FAILED=0
NONOAUTH_PASSED=0
NONOAUTH_FAILED=0
RESULTS_FILE=""

#===============================================================================
# FUNCTIONS
#===============================================================================

print_header() {
    echo -e "${PURPLE}============================================================${NC}"
    echo -e "${PURPLE}  $1${NC}"
    echo -e "${PURPLE}============================================================${NC}"
}

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

is_llm_challenge() {
    local challenge="$1"
    for llm_challenge in "${LLM_CHALLENGES[@]}"; do
        if [[ "$challenge" == "$llm_challenge" ]]; then
            return 0
        fi
    done
    return 1
}

run_challenge_oauth_mode() {
    local challenge="$1"
    print_header "Running $challenge in OAuth Mode"

    # Set OAuth environment
    export CLAUDE_CODE_USE_OAUTH_CREDENTIALS=true
    export LLM_CHALLENGE_MODE="oauth"
    export PREFER_OAUTH_PROVIDERS=true

    print_info "OAuth mode enabled"
    print_info "CLAUDE_CODE_USE_OAUTH_CREDENTIALS=true"

    # Run the challenge
    local result=0
    if "$SCRIPT_DIR/run_challenges.sh" "$challenge" 2>&1; then
        print_success "Challenge $challenge PASSED in OAuth mode"
        ((OAUTH_PASSED++))
    else
        print_error "Challenge $challenge FAILED in OAuth mode"
        ((OAUTH_FAILED++))
        result=1
    fi

    return $result
}

run_challenge_nonoauth_mode() {
    local challenge="$1"
    print_header "Running $challenge in Non-OAuth Mode"

    # Set non-OAuth environment (disable OAuth providers)
    export CLAUDE_CODE_USE_OAUTH_CREDENTIALS=false
    export LLM_CHALLENGE_MODE="api_key"
    export PREFER_OAUTH_PROVIDERS=false

    # Temporarily unset Claude OAuth credentials to force API key mode
    local saved_claude_oauth_token="${CLAUDE_OAUTH_TOKEN:-}"
    unset CLAUDE_OAUTH_TOKEN

    print_info "Non-OAuth mode enabled"
    print_info "CLAUDE_CODE_USE_OAUTH_CREDENTIALS=false"
    print_info "Using API key providers: DeepSeek, Mistral, Gemini, etc."

    # Run the challenge
    local result=0
    if "$SCRIPT_DIR/run_challenges.sh" "$challenge" 2>&1; then
        print_success "Challenge $challenge PASSED in Non-OAuth mode"
        ((NONOAUTH_PASSED++))
    else
        print_error "Challenge $challenge FAILED in Non-OAuth mode"
        ((NONOAUTH_FAILED++))
        result=1
    fi

    # Restore Claude OAuth token if it was set
    if [ -n "$saved_claude_oauth_token" ]; then
        export CLAUDE_OAUTH_TOKEN="$saved_claude_oauth_token"
    fi

    return $result
}

run_challenge_variants() {
    local challenge="$1"
    local overall_result=0

    print_header "Running $challenge - BOTH VARIANTS"
    echo ""

    # Run OAuth variant
    print_info "Variant 1/2: OAuth Mode"
    if ! run_challenge_oauth_mode "$challenge"; then
        overall_result=1
    fi
    echo ""

    # Run non-OAuth variant
    print_info "Variant 2/2: Non-OAuth (API Key) Mode"
    if ! run_challenge_nonoauth_mode "$challenge"; then
        overall_result=1
    fi
    echo ""

    return $overall_result
}

run_all_llm_challenges() {
    print_header "Running ALL LLM Challenges in Both Variants"
    echo ""
    print_info "LLM Challenges to test: ${#LLM_CHALLENGES[@]}"
    print_info "Total runs: $((${#LLM_CHALLENGES[@]} * 2))"
    echo ""

    local failed_challenges=()

    for challenge in "${LLM_CHALLENGES[@]}"; do
        echo ""
        echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${CYAN}  Challenge: $challenge${NC}"
        echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

        if ! run_challenge_variants "$challenge"; then
            failed_challenges+=("$challenge")
        fi
    done

    # Print summary
    print_summary "$failed_challenges"
}

print_summary() {
    local failed_challenges=("$@")

    echo ""
    echo -e "${PURPLE}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${PURPLE}║            LLM CHALLENGE VARIANTS - FINAL SUMMARY          ║${NC}"
    echo -e "${PURPLE}╠════════════════════════════════════════════════════════════╣${NC}"

    echo -e "${PURPLE}║${NC}  ${GREEN}OAuth Mode:${NC}"
    echo -e "${PURPLE}║${NC}    Passed: $OAUTH_PASSED"
    echo -e "${PURPLE}║${NC}    Failed: $OAUTH_FAILED"
    echo -e "${PURPLE}║${NC}"
    echo -e "${PURPLE}║${NC}  ${BLUE}Non-OAuth Mode (API Keys):${NC}"
    echo -e "${PURPLE}║${NC}    Passed: $NONOAUTH_PASSED"
    echo -e "${PURPLE}║${NC}    Failed: $NONOAUTH_FAILED"
    echo -e "${PURPLE}║${NC}"

    local total_passed=$((OAUTH_PASSED + NONOAUTH_PASSED))
    local total_failed=$((OAUTH_FAILED + NONOAUTH_FAILED))
    local total=$((total_passed + total_failed))

    echo -e "${PURPLE}║${NC}  ${YELLOW}Total:${NC}"
    echo -e "${PURPLE}║${NC}    Passed: $total_passed / $total"
    echo -e "${PURPLE}║${NC}    Failed: $total_failed / $total"

    if [ ${#failed_challenges[@]} -gt 0 ]; then
        echo -e "${PURPLE}║${NC}"
        echo -e "${PURPLE}║${NC}  ${RED}Failed challenges:${NC}"
        for failed in "${failed_challenges[@]}"; do
            echo -e "${PURPLE}║${NC}    - $failed"
        done
    fi

    echo -e "${PURPLE}╚════════════════════════════════════════════════════════════╝${NC}"

    if [ $total_failed -eq 0 ]; then
        echo ""
        print_success "All LLM challenge variants PASSED!"
        return 0
    else
        echo ""
        print_error "Some LLM challenge variants FAILED"
        return 1
    fi
}

show_help() {
    echo "LLM Challenge Variants Runner"
    echo ""
    echo "Usage:"
    echo "  $0 <challenge_name>     Run specific challenge in both OAuth and non-OAuth modes"
    echo "  $0 --all                Run all LLM challenges in both modes"
    echo "  $0 --list               List all LLM challenges"
    echo "  $0 --help               Show this help"
    echo ""
    echo "LLM Challenges:"
    for challenge in "${LLM_CHALLENGES[@]}"; do
        echo "  - $challenge"
    done
    echo ""
    echo "Modes:"
    echo "  OAuth Mode:     Uses Claude Code OAuth credentials"
    echo "  Non-OAuth Mode: Uses API key-based providers (DeepSeek, Mistral, etc.)"
}

#===============================================================================
# MAIN
#===============================================================================

if [ $# -eq 0 ]; then
    show_help
    exit 1
fi

case "$1" in
    --help|-h)
        show_help
        exit 0
        ;;
    --all)
        run_all_llm_challenges
        ;;
    --list)
        echo "LLM Challenges:"
        for challenge in "${LLM_CHALLENGES[@]}"; do
            echo "  - $challenge"
        done
        ;;
    *)
        if is_llm_challenge "$1"; then
            run_challenge_variants "$1"
        else
            print_warning "$1 is not in the LLM challenges list"
            print_info "Running challenge anyway in both modes..."
            run_challenge_variants "$1"
        fi
        ;;
esac
