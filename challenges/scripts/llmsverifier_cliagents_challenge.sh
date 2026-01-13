#!/bin/bash
# LLMsVerifier CLI Agents Challenge
# Tests the unified CLI agent configuration generation and validation via LLMsVerifier

# Don't use set -e as we track test failures manually

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
OUTPUT_DIR="${HOME}/Downloads/helixagent-cli-configs-test"
GENERATOR_DIR="$PROJECT_ROOT/challenges/codebase/go_files/unified_cli_generator"
LLMSVERIFIER_DIR="$PROJECT_ROOT/LLMsVerifier/llm-verifier"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

TESTS_PASSED=0
TESTS_FAILED=0

print_header() {
    echo ""
    echo -e "${BLUE}======================================================================${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}======================================================================${NC}"
    echo ""
}

print_test() {
    echo -e "${YELLOW}[TEST]${NC} $1"
}

pass_test() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((TESTS_PASSED++))
}

fail_test() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((TESTS_FAILED++))
}

# Test 1: LLMsVerifier cliagents package compiles
test_cliagents_package_compiles() {
    print_test "Testing LLMsVerifier cliagents package compilation..."

    cd "$LLMSVERIFIER_DIR"
    if go build ./pkg/cliagents/... 2>&1; then
        pass_test "LLMsVerifier cliagents package compiles successfully"
    else
        fail_test "LLMsVerifier cliagents package compilation failed"
        return 1
    fi
}

# Test 2: LLMsVerifier cliagents tests pass
test_cliagents_tests_pass() {
    print_test "Running LLMsVerifier cliagents package tests..."

    cd "$LLMSVERIFIER_DIR"
    local test_output
    test_output=$(go test -v ./pkg/cliagents/... 2>&1) || true
    echo "$test_output" | tail -20

    if echo "$test_output" | grep -q "^PASS"; then
        pass_test "All LLMsVerifier cliagents tests pass"
    else
        fail_test "LLMsVerifier cliagents tests failed"
        return 1
    fi
}

# Test 3: Unified generator builds
test_unified_generator_builds() {
    print_test "Testing unified CLI generator build..."

    cd "$GENERATOR_DIR"
    if go build -o unified_cli_generator . 2>&1; then
        pass_test "Unified CLI generator builds successfully"
    else
        fail_test "Unified CLI generator build failed"
        return 1
    fi
}

# Test 4: Generator lists 16 agents
test_generator_lists_16_agents() {
    print_test "Verifying generator supports 16 CLI agents..."

    cd "$GENERATOR_DIR"
    AGENT_COUNT=$(./unified_cli_generator -list 2>&1 | grep -c "Config:")

    if [ "$AGENT_COUNT" -eq 16 ]; then
        pass_test "Generator supports exactly 16 CLI agents"
    else
        fail_test "Expected 16 agents, got $AGENT_COUNT"
        return 1
    fi
}

# Test 5: Generate all configurations
test_generate_all_configs() {
    print_test "Generating all CLI agent configurations..."

    rm -rf "$OUTPUT_DIR"
    mkdir -p "$OUTPUT_DIR"

    cd "$GENERATOR_DIR"
    if ./unified_cli_generator -agent all -output-dir "$OUTPUT_DIR" 2>&1; then
        pass_test "All configurations generated successfully"
    else
        fail_test "Configuration generation failed"
        return 1
    fi
}

# Test 6: Verify config files created
test_config_files_created() {
    print_test "Verifying configuration files were created..."

    FILE_COUNT=$(find "$OUTPUT_DIR" -name "*.json" -o -name "*.yml" -o -name "*.xml" -o -name "*.lua" 2>/dev/null | wc -l)

    if [ "$FILE_COUNT" -ge 12 ]; then
        pass_test "Created $FILE_COUNT configuration files"
    else
        fail_test "Expected at least 12 config files, got $FILE_COUNT"
        return 1
    fi
}

# Test 7: Validate OpenCode config structure
test_opencode_config_structure() {
    print_test "Validating OpenCode configuration structure..."

    OPENCODE_CONFIG=$(find "$OUTPUT_DIR" -name "*opencode*" | head -1)

    if [ -z "$OPENCODE_CONFIG" ]; then
        fail_test "OpenCode configuration not found"
        return 1
    fi

    # Check for required fields using jq or python
    if command -v jq &> /dev/null; then
        if jq -e '.provider' "$OPENCODE_CONFIG" > /dev/null 2>&1; then
            pass_test "OpenCode config has valid structure (provider field present)"
        else
            fail_test "OpenCode config missing provider field"
            return 1
        fi
    else
        # Fallback to grep
        if grep -q '"provider"' "$OPENCODE_CONFIG"; then
            pass_test "OpenCode config has valid structure"
        else
            fail_test "OpenCode config missing provider field"
            return 1
        fi
    fi
}

# Test 8: Validate config contains MCP servers
test_config_has_mcp_servers() {
    print_test "Verifying configs include MCP servers..."

    OPENCODE_CONFIG=$(find "$OUTPUT_DIR" -name "*opencode*" | head -1)

    if grep -q '"mcp"' "$OPENCODE_CONFIG" 2>/dev/null || grep -q '"mcpServers"' "$OPENCODE_CONFIG" 2>/dev/null; then
        pass_test "Configuration includes MCP server definitions"
    else
        fail_test "Configuration missing MCP server definitions"
        return 1
    fi
}

# Test 9: Validate config file is valid JSON
test_configs_valid_json() {
    print_test "Verifying all JSON configs are valid..."

    INVALID_COUNT=0
    for config in "$OUTPUT_DIR"/*.json; do
        if [ -f "$config" ]; then
            if command -v jq &> /dev/null; then
                if ! jq . "$config" > /dev/null 2>&1; then
                    echo "  Invalid JSON: $config"
                    ((INVALID_COUNT++))
                fi
            else
                if ! python3 -c "import json; json.load(open('$config'))" 2>/dev/null; then
                    echo "  Invalid JSON: $config"
                    ((INVALID_COUNT++))
                fi
            fi
        fi
    done

    if [ "$INVALID_COUNT" -eq 0 ]; then
        pass_test "All JSON configuration files are valid"
    else
        fail_test "$INVALID_COUNT configuration files have invalid JSON"
        return 1
    fi
}

# Test 10: Generate specific agent (OpenCode)
test_generate_specific_agent() {
    print_test "Testing specific agent generation (opencode)..."

    cd "$GENERATOR_DIR"
    rm -rf "$OUTPUT_DIR/single"
    mkdir -p "$OUTPUT_DIR/single"
    if ./unified_cli_generator -agent opencode -output-dir "$OUTPUT_DIR/single" 2>&1; then
        # Check for unique filename pattern: opencode-helixagent.json
        if [ -f "$OUTPUT_DIR/single/opencode-helixagent.json" ]; then
            pass_test "Successfully generated specific OpenCode configuration"
        else
            fail_test "OpenCode configuration file not created (expected opencode-helixagent.json)"
            return 1
        fi
    else
        fail_test "Specific agent generation failed"
        return 1
    fi
}

# Test 11: Verify all primary agents have configs
test_primary_agents_have_configs() {
    print_test "Verifying all primary agents (OpenCode, Crush, KiloCode, HelixCode) have configs..."

    PRIMARY_AGENTS=("opencode" "crush" "kilocode" "helixcode")
    MISSING=0

    for agent in "${PRIMARY_AGENTS[@]}"; do
        if ! find "$OUTPUT_DIR" -iname "*$agent*" | grep -q .; then
            echo "  Missing config for: $agent"
            ((MISSING++))
        fi
    done

    if [ "$MISSING" -eq 0 ]; then
        pass_test "All primary agents have configurations"
    else
        fail_test "$MISSING primary agents missing configurations"
        return 1
    fi
}

# Test 12: Test schema retrieval
test_schema_retrieval() {
    print_test "Testing schema information retrieval..."

    cd "$GENERATOR_DIR"
    SCHEMA_OUTPUT=$(./unified_cli_generator -list 2>&1)

    # Check that schema info contains required fields
    if echo "$SCHEMA_OUTPUT" | grep -q "Config:" && echo "$SCHEMA_OUTPUT" | grep -q "Dir:"; then
        pass_test "Schema information retrieved successfully"
    else
        fail_test "Schema information incomplete"
        return 1
    fi
}

# Run all tests
run_all_tests() {
    print_header "LLMSVERIFIER CLI AGENTS CHALLENGE"
    echo "Testing unified CLI agent configuration generation via LLMsVerifier"
    echo ""

    test_cliagents_package_compiles
    test_cliagents_tests_pass
    test_unified_generator_builds
    test_generator_lists_16_agents
    test_generate_all_configs
    test_config_files_created
    test_opencode_config_structure
    test_config_has_mcp_servers
    test_configs_valid_json
    test_generate_specific_agent
    test_primary_agents_have_configs
    test_schema_retrieval

    print_header "CHALLENGE RESULTS"
    echo -e "Tests Passed: ${GREEN}$TESTS_PASSED${NC}"
    echo -e "Tests Failed: ${RED}$TESTS_FAILED${NC}"
    echo ""

    TOTAL=$((TESTS_PASSED + TESTS_FAILED))
    if [ "$TESTS_FAILED" -eq 0 ]; then
        echo -e "${GREEN}======================================================================${NC}"
        echo -e "${GREEN}  CHALLENGE PASSED - All $TOTAL tests passed!${NC}"
        echo -e "${GREEN}======================================================================${NC}"
        exit 0
    else
        echo -e "${RED}======================================================================${NC}"
        echo -e "${RED}  CHALLENGE FAILED - $TESTS_FAILED of $TOTAL tests failed${NC}"
        echo -e "${RED}======================================================================${NC}"
        exit 1
    fi
}

# Main
run_all_tests
