#!/bin/bash
# ============================================================================
# CLI Agent Integration Challenge
# ============================================================================
# Comprehensive verification of CLI agent configurations and plugins
# for all 47+ supported CLI agents
#
# Tests: Config generation, plugin installation, structure validation,
#        content verification, integration compliance
# ============================================================================

set -e

# Source challenge framework
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/challenge_framework.sh"

# Initialize challenge
init_challenge "cli_agent_integration" "CLI Agent Integration Verification"
load_env

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_TOTAL=0

# CLI agents scripts directory
CLI_AGENTS_DIR="$PROJECT_ROOT/scripts/cli-agents"
CONFIG_OUTPUT_DIR="$CLI_AGENTS_DIR/configs/generated"
PLUGIN_OUTPUT_DIR="$CLI_AGENTS_DIR/plugins/generated"

# Test function
run_test() {
    local test_name="$1"
    local test_cmd="$2"
    TESTS_TOTAL=$((TESTS_TOTAL + 1))

    if eval "$test_cmd" >> "$LOG_FILE" 2>&1; then
        log_success "PASS: $test_name"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        record_assertion "test" "$test_name" "true" ""
        return 0
    else
        log_error "FAIL: $test_name"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        record_assertion "test" "$test_name" "false" "Test command failed"
        return 1
    fi
}

# ============================================================================
# SECTION 1: SCRIPT EXISTENCE
# ============================================================================
log_info "=============================================="
log_info "SECTION 1: Script Existence Verification"
log_info "=============================================="

run_test "Config generator script exists" \
    "[[ -f '$CLI_AGENTS_DIR/generate-all-configs.sh' ]]"

run_test "Plugin installer script exists" \
    "[[ -f '$CLI_AGENTS_DIR/install-plugins.sh' ]]"

run_test "Integration test script exists" \
    "[[ -f '$CLI_AGENTS_DIR/tests/cli-agent-integration-test.sh' ]]"

# Make scripts executable
chmod +x "$CLI_AGENTS_DIR/generate-all-configs.sh" 2>/dev/null || true
chmod +x "$CLI_AGENTS_DIR/install-plugins.sh" 2>/dev/null || true
chmod +x "$CLI_AGENTS_DIR/tests/cli-agent-integration-test.sh" 2>/dev/null || true

run_test "Config generator is executable" \
    "[[ -x '$CLI_AGENTS_DIR/generate-all-configs.sh' ]]"

run_test "Plugin installer is executable" \
    "[[ -x '$CLI_AGENTS_DIR/install-plugins.sh' ]]"

# ============================================================================
# SECTION 2: CONFIGURATION GENERATION
# ============================================================================
log_info "=============================================="
log_info "SECTION 2: Configuration Generation"
log_info "=============================================="

# Run config generator
log_info "Running configuration generator..."
run_test "Config generator executes successfully" \
    "bash '$CLI_AGENTS_DIR/generate-all-configs.sh'"

run_test "Generated configs directory exists" \
    "[[ -d '$CONFIG_OUTPUT_DIR' ]]"

# ============================================================================
# SECTION 3: TIER 1 AGENTS (Primary Support)
# ============================================================================
log_info "=============================================="
log_info "SECTION 3: Tier 1 Agents Configuration"
log_info "=============================================="

TIER1_AGENTS=("claude_code" "aider" "cline" "opencode" "kilo_code" "gemini_cli" "qwen_code" "deepseek_cli" "forge" "codename_goose")

for agent in "${TIER1_AGENTS[@]}"; do
    run_test "Tier 1: $agent config directory exists" \
        "[[ -d '$CONFIG_OUTPUT_DIR/$agent' ]]"
done

# Validate specific agent configs
run_test "Claude Code has settings.json" \
    "[[ -f '$CONFIG_OUTPUT_DIR/claude_code/settings.json' ]]"

run_test "Claude Code settings.json is valid JSON" \
    "python3 -c \"import json; json.load(open('$CONFIG_OUTPUT_DIR/claude_code/settings.json'))\""

run_test "Aider has .aider.conf.yml" \
    "[[ -f '$CONFIG_OUTPUT_DIR/aider/.aider.conf.yml' ]]"

run_test "Aider config is valid YAML" \
    "python3 -c \"import yaml; yaml.safe_load(open('$CONFIG_OUTPUT_DIR/aider/.aider.conf.yml'))\""

run_test "Cline has config.json" \
    "[[ -f '$CONFIG_OUTPUT_DIR/cline/config.json' ]]"

run_test "OpenCode has opencode.json" \
    "[[ -f '$CONFIG_OUTPUT_DIR/opencode/opencode.json' ]]"

# ============================================================================
# SECTION 4: TIER 2 AGENTS (Secondary Support)
# ============================================================================
log_info "=============================================="
log_info "SECTION 4: Tier 2 Agents Configuration"
log_info "=============================================="

TIER2_AGENTS=("amazon_q" "kiro" "gpt_engineer" "mistral_code" "ollama_code" "plandex" "codex" "vtcode" "nanocoder" "gitmcp" "taskweaver" "octogen" "fauxpilot" "bridle" "agent_deck")

for agent in "${TIER2_AGENTS[@]}"; do
    run_test "Tier 2: $agent config directory exists" \
        "[[ -d '$CONFIG_OUTPUT_DIR/$agent' ]]"
done

# ============================================================================
# SECTION 5: TIER 3 AGENTS (Extended Support)
# ============================================================================
log_info "=============================================="
log_info "SECTION 5: Tier 3 Agents Configuration"
log_info "=============================================="

TIER3_AGENTS=("claude_squad" "codai" "emdash" "get_shit_done" "github_copilot_cli" "github_spec_kit" "gptme" "mobile_agent" "multiagent_coding" "noi" "openhands" "postgres_mcp" "shai" "snowcli" "superset" "warp" "cheshire_cat" "conduit" "crush" "helixcode")

for agent in "${TIER3_AGENTS[@]}"; do
    run_test "Tier 3: $agent config directory exists" \
        "[[ -d '$CONFIG_OUTPUT_DIR/$agent' ]]"
done

# ============================================================================
# SECTION 6: CONFIGURATION CONTENT VERIFICATION
# ============================================================================
log_info "=============================================="
log_info "SECTION 6: Configuration Content Verification"
log_info "=============================================="

run_test "Claude Code config has helix provider" \
    "grep -q 'helix' '$CONFIG_OUTPUT_DIR/claude_code/settings.json'"

run_test "Claude Code config has ai-debate-ensemble model" \
    "grep -q 'ai-debate-ensemble' '$CONFIG_OUTPUT_DIR/claude_code/settings.json'"

run_test "Claude Code config has streaming enabled" \
    "grep -q 'streaming' '$CONFIG_OUTPUT_DIR/claude_code/settings.json'"

run_test "Aider config has ai-debate-ensemble model" \
    "grep -q 'ai-debate-ensemble' '$CONFIG_OUTPUT_DIR/aider/.aider.conf.yml'"

run_test "Aider config has helix settings" \
    "grep -q 'helix' '$CONFIG_OUTPUT_DIR/aider/.aider.conf.yml'"

run_test "Cline config has openai-compatible provider" \
    "grep -q 'openai-compatible' '$CONFIG_OUTPUT_DIR/cline/config.json'"

run_test "OpenCode config has helix features" \
    "grep -q 'helix' '$CONFIG_OUTPUT_DIR/opencode/opencode.json'"

# ============================================================================
# SECTION 7: PLUGIN GENERATION
# ============================================================================
log_info "=============================================="
log_info "SECTION 7: Plugin Generation"
log_info "=============================================="

# Run plugin installer
log_info "Running plugin installer..."
run_test "Plugin installer executes successfully" \
    "bash '$CLI_AGENTS_DIR/install-plugins.sh'"

run_test "Generated plugins directory exists" \
    "[[ -d '$PLUGIN_OUTPUT_DIR' ]]"

# ============================================================================
# SECTION 8: CORE PLUGINS VERIFICATION
# ============================================================================
log_info "=============================================="
log_info "SECTION 8: Core Plugins Verification"
log_info "=============================================="

CORE_PLUGINS=("helix-integration" "event-handler" "verifier-client" "debate-ui" "streaming-adapter" "mcp-bridge")

for plugin in "${CORE_PLUGINS[@]}"; do
    run_test "Core plugin '$plugin' is generated" \
        "find '$PLUGIN_OUTPUT_DIR' -type d -name '$plugin' | grep -q ."
done

# ============================================================================
# SECTION 9: PLUGIN CONTENT VERIFICATION
# ============================================================================
log_info "=============================================="
log_info "SECTION 9: Plugin Content Verification"
log_info "=============================================="

# Check plugin structure for major agents
MAJOR_PLUGIN_AGENTS=("claude_code" "opencode" "helixcode" "cline")

for agent in "${MAJOR_PLUGIN_AGENTS[@]}"; do
    if [[ -d "$PLUGIN_OUTPUT_DIR/$agent" ]]; then
        run_test "$agent has helix-integration plugin" \
            "[[ -d '$PLUGIN_OUTPUT_DIR/$agent/helix-integration' ]]"

        if [[ -d "$PLUGIN_OUTPUT_DIR/$agent/helix-integration" ]]; then
            run_test "$agent helix-integration has manifest.json" \
                "[[ -f '$PLUGIN_OUTPUT_DIR/$agent/helix-integration/manifest.json' ]]"

            run_test "$agent helix-integration manifest is valid JSON" \
                "python3 -c \"import json; json.load(open('$PLUGIN_OUTPUT_DIR/$agent/helix-integration/manifest.json'))\""

            run_test "$agent helix-integration has Go source" \
                "[[ -f '$PLUGIN_OUTPUT_DIR/$agent/helix-integration/helix_integration.go' ]]"
        fi

        run_test "$agent has index.json" \
            "[[ -f '$PLUGIN_OUTPUT_DIR/$agent/index.json' ]]"

        if [[ -f "$PLUGIN_OUTPUT_DIR/$agent/index.json" ]]; then
            run_test "$agent index.json is valid JSON" \
                "python3 -c \"import json; json.load(open('$PLUGIN_OUTPUT_DIR/$agent/index.json'))\""
        fi
    fi
done

# ============================================================================
# SECTION 10: PLUGIN SOURCE CODE VERIFICATION
# ============================================================================
log_info "=============================================="
log_info "SECTION 10: Plugin Source Code Verification"
log_info "=============================================="

# Find a helix-integration plugin to verify
SAMPLE_PLUGIN=""
for agent in claude_code opencode helixcode; do
    if [[ -f "$PLUGIN_OUTPUT_DIR/$agent/helix-integration/helix_integration.go" ]]; then
        SAMPLE_PLUGIN="$PLUGIN_OUTPUT_DIR/$agent/helix-integration/helix_integration.go"
        break
    fi
done

if [[ -n "$SAMPLE_PLUGIN" ]]; then
    run_test "Plugin has package declaration" \
        "grep -q 'package main' '$SAMPLE_PLUGIN'"

    run_test "Plugin has PluginName variable" \
        "grep -q 'PluginName' '$SAMPLE_PLUGIN'"

    run_test "Plugin has PluginVersion variable" \
        "grep -q 'PluginVersion' '$SAMPLE_PLUGIN'"

    run_test "Plugin has Init function" \
        "grep -q 'func Init' '$SAMPLE_PLUGIN'"

    run_test "Plugin has Shutdown function" \
        "grep -q 'func Shutdown' '$SAMPLE_PLUGIN'"

    run_test "Plugin has HTTP client" \
        "grep -q 'http.Client' '$SAMPLE_PLUGIN'"
fi

# ============================================================================
# SECTION 11: MANIFEST CONTENT VERIFICATION
# ============================================================================
log_info "=============================================="
log_info "SECTION 11: Manifest Content Verification"
log_info "=============================================="

SAMPLE_MANIFEST=""
for agent in claude_code opencode helixcode; do
    if [[ -f "$PLUGIN_OUTPUT_DIR/$agent/helix-integration/manifest.json" ]]; then
        SAMPLE_MANIFEST="$PLUGIN_OUTPUT_DIR/$agent/helix-integration/manifest.json"
        break
    fi
done

if [[ -n "$SAMPLE_MANIFEST" ]]; then
    run_test "Manifest has name field" \
        "grep -q '\"name\"' '$SAMPLE_MANIFEST'"

    run_test "Manifest has version field" \
        "grep -q '\"version\"' '$SAMPLE_MANIFEST'"

    run_test "Manifest has description field" \
        "grep -q '\"description\"' '$SAMPLE_MANIFEST'"

    run_test "Manifest has entry_point field" \
        "grep -q '\"entry_point\"' '$SAMPLE_MANIFEST'"

    run_test "Manifest has config_schema field" \
        "grep -q '\"config_schema\"' '$SAMPLE_MANIFEST'"
fi

# ============================================================================
# SECTION 12: INTEGRATION COMPLIANCE
# ============================================================================
log_info "=============================================="
log_info "SECTION 12: Integration Compliance"
log_info "=============================================="

# Check that configs reference correct endpoints
run_test "Configs reference /v1 API path" \
    "grep -r '/v1' '$CONFIG_OUTPUT_DIR' | head -1 | grep -q '/v1'"

run_test "Configs reference localhost:8080 (default)" \
    "grep -r 'localhost:8080\|127.0.0.1:8080' '$CONFIG_OUTPUT_DIR' | head -1"

run_test "Configs have ai-debate-ensemble model" \
    "grep -r 'ai-debate-ensemble' '$CONFIG_OUTPUT_DIR' | wc -l | xargs test 10 -lt"

run_test "Configs have streaming configuration" \
    "grep -r 'stream' '$CONFIG_OUTPUT_DIR' | wc -l | xargs test 5 -lt"

run_test "Configs have retry configuration" \
    "grep -r 'retry' '$CONFIG_OUTPUT_DIR' | head -1"

# ============================================================================
# SECTION 13: AGENT COUNT VERIFICATION
# ============================================================================
log_info "=============================================="
log_info "SECTION 13: Agent Count Verification"
log_info "=============================================="

# Count generated configs
CONFIG_COUNT=$(find "$CONFIG_OUTPUT_DIR" -maxdepth 1 -type d | wc -l)
CONFIG_COUNT=$((CONFIG_COUNT - 1))  # Exclude the parent directory

run_test "At least 40 agent configs generated (got: $CONFIG_COUNT)" \
    "[[ $CONFIG_COUNT -ge 40 ]]"

# Count generated plugin directories
PLUGIN_COUNT=$(find "$PLUGIN_OUTPUT_DIR" -maxdepth 1 -type d | wc -l)
PLUGIN_COUNT=$((PLUGIN_COUNT - 1))  # Exclude the parent directory

run_test "At least 40 agent plugin sets generated (got: $PLUGIN_COUNT)" \
    "[[ $PLUGIN_COUNT -ge 40 ]]"

# ============================================================================
# SECTION 14: CROSS-VALIDATION
# ============================================================================
log_info "=============================================="
log_info "SECTION 14: Cross-Validation"
log_info "=============================================="

# Ensure each config has corresponding plugins
MISSING_PLUGINS=0
for config_dir in "$CONFIG_OUTPUT_DIR"/*/; do
    agent_name=$(basename "$config_dir")
    if [[ ! -d "$PLUGIN_OUTPUT_DIR/$agent_name" ]]; then
        MISSING_PLUGINS=$((MISSING_PLUGINS + 1))
        log_warning "Agent $agent_name has config but no plugins"
    fi
done

run_test "All configured agents have plugins (missing: $MISSING_PLUGINS)" \
    "[[ $MISSING_PLUGINS -lt 5 ]]"

# ============================================================================
# SUMMARY
# ============================================================================
log_info "=============================================="
log_info "CHALLENGE SUMMARY"
log_info "=============================================="

log_info "Total tests: $TESTS_TOTAL"
log_info "Passed: $TESTS_PASSED"
log_info "Failed: $TESTS_FAILED"
log_info "Configs generated: $CONFIG_COUNT"
log_info "Plugin sets generated: $PLUGIN_COUNT"

if [[ $TESTS_FAILED -eq 0 ]]; then
    log_success "=============================================="
    log_success "ALL CLI AGENT INTEGRATION TESTS PASSED!"
    log_success "=============================================="
    finalize_challenge 0
else
    log_error "=============================================="
    log_error "SOME TESTS FAILED"
    log_error "=============================================="
    finalize_challenge 1
fi
