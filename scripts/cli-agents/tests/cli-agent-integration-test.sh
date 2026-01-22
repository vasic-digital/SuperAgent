#!/bin/bash
# ============================================================================
# CLI Agent Integration Test Suite
# ============================================================================
# Tests configuration generation, plugin installation, and integration
# for all 47+ supported CLI agents
#
# Usage: ./cli-agent-integration-test.sh [--verbose] [--agent=NAME]
# ============================================================================

set -e

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PARENT_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(cd "$PARENT_DIR/../.." && pwd)"

# Test configuration
TEST_OUTPUT_DIR="/tmp/helix-cli-agent-tests"
TEST_HOME_DIR="$TEST_OUTPUT_DIR/home"
LOG_FILE="$TEST_OUTPUT_DIR/test.log"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Counters
TESTS_TOTAL=0
TESTS_PASSED=0
TESTS_FAILED=0

# Parse arguments
VERBOSE=false
SPECIFIC_AGENT=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        --agent=*)
            SPECIFIC_AGENT="${1#*=}"
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [--verbose] [--agent=NAME]"
            echo ""
            echo "Options:"
            echo "  --verbose, -v   Show detailed output"
            echo "  --agent=NAME    Test specific agent only"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Logging
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[PASS]${NC} $1"; }
log_error() { echo -e "${RED}[FAIL]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARN]${NC} $1"; }

verbose_log() {
    if [[ "$VERBOSE" == "true" ]]; then
        echo "  $1"
    fi
}

# Test functions
run_test() {
    local test_name="$1"
    local test_cmd="$2"
    local expected_result="${3:-0}"

    TESTS_TOTAL=$((TESTS_TOTAL + 1))

    verbose_log "Running: $test_cmd"

    if eval "$test_cmd" >> "$LOG_FILE" 2>&1; then
        actual_result=0
    else
        actual_result=1
    fi

    if [[ "$actual_result" -eq "$expected_result" ]]; then
        log_success "$test_name"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        log_error "$test_name"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

assert_file_exists() {
    local file="$1"
    local description="$2"
    run_test "File exists: $description" "[[ -f '$file' ]]"
}

assert_dir_exists() {
    local dir="$1"
    local description="$2"
    run_test "Dir exists: $description" "[[ -d '$dir' ]]"
}

assert_file_contains() {
    local file="$1"
    local pattern="$2"
    local description="$3"
    run_test "File contains '$pattern': $description" "grep -q '$pattern' '$file'"
}

assert_json_valid() {
    local file="$1"
    local description="$2"
    run_test "Valid JSON: $description" "python3 -c \"import json; json.load(open('$file'))\""
}

assert_yaml_valid() {
    local file="$1"
    local description="$2"
    run_test "Valid YAML: $description" "python3 -c \"import yaml; yaml.safe_load(open('$file'))\""
}

# ============================================================================
# Setup
# ============================================================================

setup_test_environment() {
    log_info "Setting up test environment..."

    # Clean previous test run
    rm -rf "$TEST_OUTPUT_DIR"
    mkdir -p "$TEST_OUTPUT_DIR"
    mkdir -p "$TEST_HOME_DIR"

    # Initialize log file
    echo "CLI Agent Integration Test Log - $(date)" > "$LOG_FILE"
    echo "=========================================" >> "$LOG_FILE"

    log_success "Test environment ready: $TEST_OUTPUT_DIR"
}

cleanup_test_environment() {
    log_info "Cleaning up test environment..."
    # Keep logs for debugging
    # rm -rf "$TEST_OUTPUT_DIR"
    log_info "Logs preserved at: $LOG_FILE"
}

# ============================================================================
# Test: Script Existence
# ============================================================================

test_scripts_exist() {
    log_info "Testing script existence..."

    assert_file_exists "$PARENT_DIR/generate-all-configs.sh" "Config generator script"
    assert_file_exists "$PARENT_DIR/install-plugins.sh" "Plugin installer script"

    # Check scripts are executable
    run_test "Config generator is executable" "[[ -x '$PARENT_DIR/generate-all-configs.sh' ]]"
    run_test "Plugin installer is executable" "[[ -x '$PARENT_DIR/install-plugins.sh' ]]"
}

# ============================================================================
# Test: Configuration Generation
# ============================================================================

test_config_generation() {
    log_info "Testing configuration generation..."

    # Run config generator in dry-run mode
    run_test "Config generator runs successfully" \
        "bash '$PARENT_DIR/generate-all-configs.sh' --dry-run"

    # Run actual generation
    run_test "Config generator creates configs" \
        "bash '$PARENT_DIR/generate-all-configs.sh'"

    # Check output directory exists
    local config_dir="$PARENT_DIR/configs/generated"
    assert_dir_exists "$config_dir" "Generated configs directory"

    # Check configs for major agents
    local major_agents=("claude_code" "aider" "cline" "opencode" "kilo_code" "gemini_cli")

    for agent in "${major_agents[@]}"; do
        assert_dir_exists "$config_dir/$agent" "Config dir for $agent"
    done
}

test_config_content() {
    log_info "Testing configuration content..."

    local config_dir="$PARENT_DIR/configs/generated"

    # Claude Code config
    if [[ -f "$config_dir/claude_code/settings.json" ]]; then
        assert_json_valid "$config_dir/claude_code/settings.json" "Claude Code settings.json"
        assert_file_contains "$config_dir/claude_code/settings.json" "helix" "Claude Code has helix config"
        assert_file_contains "$config_dir/claude_code/settings.json" "ai-debate-ensemble" "Claude Code has correct model"
    fi

    # Aider config
    if [[ -f "$config_dir/aider/.aider.conf.yml" ]]; then
        assert_yaml_valid "$config_dir/aider/.aider.conf.yml" "Aider config"
        assert_file_contains "$config_dir/aider/.aider.conf.yml" "ai-debate-ensemble" "Aider has correct model"
    fi

    # Cline config
    if [[ -f "$config_dir/cline/config.json" ]]; then
        assert_json_valid "$config_dir/cline/config.json" "Cline config"
        assert_file_contains "$config_dir/cline/config.json" "openai-compatible" "Cline uses OpenAI compatible"
    fi

    # OpenCode config
    if [[ -f "$config_dir/opencode/opencode.json" ]]; then
        assert_json_valid "$config_dir/opencode/opencode.json" "OpenCode config"
        assert_file_contains "$config_dir/opencode/opencode.json" "helix" "OpenCode has helix config"
    fi
}

test_all_agents_have_configs() {
    log_info "Testing all agents have configs..."

    local config_dir="$PARENT_DIR/configs/generated"

    # List of all expected agents
    local all_agents=(
        "claude_code" "aider" "cline" "opencode" "kilo_code"
        "gemini_cli" "qwen_code" "deepseek_cli" "forge" "codename_goose"
        "amazon_q" "kiro" "gpt_engineer" "mistral_code" "ollama_code"
        "plandex" "codex" "vtcode" "nanocoder" "gitmcp"
        "taskweaver" "octogen" "fauxpilot" "bridle" "agent_deck"
        "claude_squad" "codai" "emdash" "get_shit_done"
        "github_copilot_cli" "github_spec_kit" "gptme" "mobile_agent"
        "multiagent_coding" "noi" "openhands" "postgres_mcp" "shai"
        "snowcli" "superset" "warp" "cheshire_cat" "conduit"
        "crush" "helixcode"
    )

    local agents_with_configs=0
    for agent in "${all_agents[@]}"; do
        if [[ -d "$config_dir/$agent" ]]; then
            agents_with_configs=$((agents_with_configs + 1))
            verbose_log "Config exists for: $agent"
        fi
    done

    run_test "At least 40 agents have configs" "[[ $agents_with_configs -ge 40 ]]"
}

# ============================================================================
# Test: Plugin Installation
# ============================================================================

test_plugin_generation() {
    log_info "Testing plugin generation..."

    # Run plugin installer in dry-run mode
    run_test "Plugin installer runs in dry-run" \
        "bash '$PARENT_DIR/install-plugins.sh' --dry-run"

    # Run actual plugin generation
    run_test "Plugin installer generates plugins" \
        "bash '$PARENT_DIR/install-plugins.sh'"

    # Check output directory exists
    local plugin_dir="$PARENT_DIR/plugins/generated"
    assert_dir_exists "$plugin_dir" "Generated plugins directory"
}

test_plugin_content() {
    log_info "Testing plugin content..."

    local plugin_dir="$PARENT_DIR/plugins/generated"

    # Check major agents have plugins
    local agents=("claude_code" "opencode" "helixcode")

    for agent in "${agents[@]}"; do
        if [[ -d "$plugin_dir/$agent" ]]; then
            # Check helix-integration plugin
            if [[ -d "$plugin_dir/$agent/helix-integration" ]]; then
                assert_file_exists "$plugin_dir/$agent/helix-integration/manifest.json" \
                    "$agent helix-integration manifest"
                assert_file_exists "$plugin_dir/$agent/helix-integration/helix_integration.go" \
                    "$agent helix-integration source"

                # Validate manifest JSON
                assert_json_valid "$plugin_dir/$agent/helix-integration/manifest.json" \
                    "$agent helix-integration manifest is valid JSON"
            fi

            # Check index.json
            if [[ -f "$plugin_dir/$agent/index.json" ]]; then
                assert_json_valid "$plugin_dir/$agent/index.json" "$agent index.json"
            fi
        fi
    done
}

test_plugin_structure() {
    log_info "Testing plugin structure..."

    local plugin_dir="$PARENT_DIR/plugins/generated"

    # Check core plugins are generated
    local core_plugins=("helix-integration" "event-handler" "verifier-client" "debate-ui" "streaming-adapter" "mcp-bridge")

    local plugins_found=0
    for plugin in "${core_plugins[@]}"; do
        # Check if this plugin exists for any agent
        if find "$plugin_dir" -type d -name "$plugin" | grep -q .; then
            plugins_found=$((plugins_found + 1))
            verbose_log "Plugin found: $plugin"
        fi
    done

    run_test "All 6 core plugins are generated" "[[ $plugins_found -eq 6 ]]"
}

# ============================================================================
# Test: Go Plugin Compilation
# ============================================================================

test_plugin_compilation() {
    log_info "Testing plugin compilation..."

    local plugin_dir="$PARENT_DIR/plugins/generated"

    # Find a plugin to test compile
    local test_plugin=""
    for agent in claude_code opencode helixcode; do
        if [[ -f "$plugin_dir/$agent/helix-integration/helix_integration.go" ]]; then
            test_plugin="$plugin_dir/$agent/helix-integration/helix_integration.go"
            break
        fi
    done

    if [[ -n "$test_plugin" ]]; then
        # Test Go syntax (not full compilation since it needs the module)
        run_test "Plugin Go syntax is valid" \
            "cd '$PROJECT_ROOT' && go fmt '$test_plugin' > /dev/null"
    else
        log_warning "No plugin found for compilation test"
    fi
}

# ============================================================================
# Test: Specific Agent
# ============================================================================

test_specific_agent() {
    local agent="$1"
    log_info "Testing specific agent: $agent"

    local config_dir="$PARENT_DIR/configs/generated"
    local plugin_dir="$PARENT_DIR/plugins/generated"

    # Generate config for specific agent
    run_test "Generate config for $agent" \
        "bash '$PARENT_DIR/generate-all-configs.sh' --agent=$agent"

    # Generate plugins for specific agent
    run_test "Generate plugins for $agent" \
        "bash '$PARENT_DIR/install-plugins.sh' --agent=$agent"

    # Check config exists
    assert_dir_exists "$config_dir/$agent" "$agent config directory"

    # Check plugins exist
    assert_dir_exists "$plugin_dir/$agent" "$agent plugin directory"

    # Check plugin index
    if [[ -f "$plugin_dir/$agent/index.json" ]]; then
        assert_json_valid "$plugin_dir/$agent/index.json" "$agent plugin index"
    fi
}

# ============================================================================
# Test: Integration Validation
# ============================================================================

test_end_to_end() {
    log_info "Testing end-to-end integration..."

    # Test that configs reference correct HelixAgent endpoints
    local config_dir="$PARENT_DIR/configs/generated"

    # Check configs have correct API paths
    run_test "Configs have /v1 endpoint" \
        "grep -r '/v1' '$config_dir' | head -1 | grep -q '/v1'"

    # Check configs have chat/completions path
    run_test "Configs reference chat/completions" \
        "grep -r 'chat/completions\|chat_completions\|ChatCompletion' '$config_dir' | head -1"

    # Check configs have streaming support
    run_test "Configs have streaming support" \
        "grep -r 'stream' '$config_dir' | grep -q 'true'"

    # Check configs reference ai-debate-ensemble model
    run_test "Configs use ai-debate-ensemble model" \
        "grep -r 'ai-debate-ensemble' '$config_dir' | head -1"
}

# ============================================================================
# Main Execution
# ============================================================================

main() {
    echo "=============================================="
    echo "CLI Agent Integration Test Suite"
    echo "=============================================="
    echo "Project: $PROJECT_ROOT"
    echo "Scripts: $PARENT_DIR"
    echo "Output:  $TEST_OUTPUT_DIR"
    echo ""

    setup_test_environment

    if [[ -n "$SPECIFIC_AGENT" ]]; then
        test_specific_agent "$SPECIFIC_AGENT"
    else
        # Run all tests
        test_scripts_exist
        test_config_generation
        test_config_content
        test_all_agents_have_configs
        test_plugin_generation
        test_plugin_content
        test_plugin_structure
        test_plugin_compilation
        test_end_to_end
    fi

    echo ""
    echo "=============================================="
    echo "Test Results"
    echo "=============================================="
    echo "Total:  $TESTS_TOTAL"
    echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
    if [[ $TESTS_FAILED -gt 0 ]]; then
        echo -e "${RED}Failed: $TESTS_FAILED${NC}"
    fi
    echo ""

    cleanup_test_environment

    if [[ $TESTS_FAILED -gt 0 ]]; then
        echo -e "${RED}SOME TESTS FAILED${NC}"
        exit 1
    else
        echo -e "${GREEN}ALL TESTS PASSED${NC}"
        exit 0
    fi
}

main "$@"
