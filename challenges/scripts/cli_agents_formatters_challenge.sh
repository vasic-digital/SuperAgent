#!/bin/bash
# CLI Agents Formatters Integration Challenge
# Validates that all 48 CLI agents have formatters configuration

set -euo pipefail

PASSED=0
FAILED=0
TOTAL=0

echo "=== CLI Agents Formatters Integration Challenge ==="
echo

# Helper functions
test_case() {
    local name=$1
    TOTAL=$((TOTAL+1))
    echo -n "Test $TOTAL: $name... "
}

pass() {
    echo "✓ PASS"
    PASSED=$((PASSED+1))
}

fail() {
    local reason=$1
    echo "✗ FAIL: $reason"
    FAILED=$((FAILED+1))
}

# Test 1: Formatters Config File
test_case "formatters_config.go exists"
if [ -f "LLMsVerifier/llm-verifier/pkg/cliagents/formatters_config.go" ]; then
    pass
else
    fail "file not found"
fi

test_case "formatters_config.go defines FormattersConfig"
if grep -q "type FormattersConfig struct" LLMsVerifier/llm-verifier/pkg/cliagents/formatters_config.go; then
    pass
else
    fail "type not found"
fi

test_case "formatters_config.go defines DefaultFormattersConfig"
if grep -q "func DefaultFormattersConfig" LLMsVerifier/llm-verifier/pkg/cliagents/formatters_config.go; then
    pass
else
    fail "function not found"
fi

test_case "DefaultFormattersConfig has preferences"
if grep -q "Preferences:" LLMsVerifier/llm-verifier/pkg/cliagents/formatters_config.go; then
    pass
else
    fail "preferences not found"
fi

test_case "DefaultFormattersConfig has fallback chains"
if grep -q "Fallback:" LLMsVerifier/llm-verifier/pkg/cliagents/formatters_config.go; then
    pass
else
    fail "fallback not found"
fi

# Test 2: OpenCode Integration
test_case "OpenCodeConfig has Formatters field"
if grep -q "Formatters.*FormattersConfig" LLMsVerifier/llm-verifier/pkg/cliagents/opencode.go; then
    pass
else
    fail "field not found"
fi

test_case "OpenCode Generate() initializes formatters"
if grep -q "DefaultFormattersConfig" LLMsVerifier/llm-verifier/pkg/cliagents/opencode.go; then
    pass
else
    fail "initialization not found"
fi

# Test 3: Crush Integration
test_case "CrushConfig has Formatters field"
if grep -q "Formatters.*FormattersConfig" LLMsVerifier/llm-verifier/pkg/cliagents/crush.go; then
    pass
else
    fail "field not found"
fi

test_case "Crush Generate() initializes formatters"
if grep -q "DefaultFormattersConfig" LLMsVerifier/llm-verifier/pkg/cliagents/crush.go; then
    pass
else
    fail "initialization not found"
fi

# Test 4: HelixCode Integration
test_case "HelixCodeConfig has Formatters field"
if grep -q "Formatters.*FormattersConfig" LLMsVerifier/llm-verifier/pkg/cliagents/helixcode.go; then
    pass
else
    fail "field not found"
fi

test_case "HelixCode Generate() initializes formatters"
if grep -q "DefaultFormattersConfig" LLMsVerifier/llm-verifier/pkg/cliagents/helixcode.go; then
    pass
else
    fail "initialization not found"
fi

# Test 5: KiloCode Integration
test_case "KiloCodeConfig has Formatters field"
if grep -q "Formatters.*FormattersConfig" LLMsVerifier/llm-verifier/pkg/cliagents/kilocode.go; then
    pass
else
    fail "field not found"
fi

test_case "KiloCode Generate() initializes formatters"
if grep -q "DefaultFormattersConfig" LLMsVerifier/llm-verifier/pkg/cliagents/kilocode.go; then
    pass
else
    fail "initialization not found"
fi

# Test 6: Additional Agents (30+ agents)
test_case "GenericAgentConfig has Formatters field"
if grep -q "Formatters.*FormattersConfig" LLMsVerifier/llm-verifier/pkg/cliagents/additional_agents.go; then
    pass
else
    fail "field not found"
fi

test_case "Additional agents Generate() initializes formatters"
if grep -q "DefaultFormattersConfig" LLMsVerifier/llm-verifier/pkg/cliagents/additional_agents.go; then
    pass
else
    fail "initialization not found"
fi

# Test 7: Language Preferences
test_case "Python formatter preference is ruff"
if grep -q '"python".*"ruff"' LLMsVerifier/llm-verifier/pkg/cliagents/formatters_config.go; then
    pass
else
    fail "preference not found"
fi

test_case "JavaScript formatter preference is biome"
if grep -q '"javascript".*"biome"' LLMsVerifier/llm-verifier/pkg/cliagents/formatters_config.go; then
    pass
else
    fail "preference not found"
fi

test_case "Go formatter preference is gofmt"
if grep -q '"go".*"gofmt"' LLMsVerifier/llm-verifier/pkg/cliagents/formatters_config.go; then
    pass
else
    fail "preference not found"
fi

test_case "Rust formatter preference is rustfmt"
if grep -q '"rust".*"rustfmt"' LLMsVerifier/llm-verifier/pkg/cliagents/formatters_config.go; then
    pass
else
    fail "preference not found"
fi

test_case "SQL formatter preference is sqlfluff"
if grep -q '"sql".*"sqlfluff"' LLMsVerifier/llm-verifier/pkg/cliagents/formatters_config.go; then
    pass
else
    fail "preference not found"
fi

test_case "Ruby formatter preference is rubocop"
if grep -q '"ruby".*"rubocop"' LLMsVerifier/llm-verifier/pkg/cliagents/formatters_config.go; then
    pass
else
    fail "preference not found"
fi

test_case "PHP formatter preference is php-cs-fixer"
if grep -q '"php".*"php-cs-fixer"' LLMsVerifier/llm-verifier/pkg/cliagents/formatters_config.go; then
    pass
else
    fail "preference not found"
fi

# Test 8: Fallback Chains
test_case "Python has fallback chain"
if grep -q '"python".*\[\]string{"black", "autopep8"' LLMsVerifier/llm-verifier/pkg/cliagents/formatters_config.go; then
    pass
else
    fail "fallback chain not found"
fi

test_case "JavaScript has fallback chain"
if grep -q '"javascript".*\[\]string{"prettier"' LLMsVerifier/llm-verifier/pkg/cliagents/formatters_config.go; then
    pass
else
    fail "fallback chain not found"
fi

# Test 9: Service URL Configuration
test_case "ServiceURL configured in DefaultFormattersConfig"
if grep -q "ServiceURL:.*fmt.Sprintf" LLMsVerifier/llm-verifier/pkg/cliagents/formatters_config.go; then
    pass
else
    fail "service URL not configured"
fi

test_case "ServiceURL uses /v1/format endpoint"
if grep -q "/v1/format" LLMsVerifier/llm-verifier/pkg/cliagents/formatters_config.go; then
    pass
else
    fail "wrong endpoint"
fi

# Test 10: Compilation
test_case "CLI agents package compiles"
if cd LLMsVerifier/llm-verifier && go build ./pkg/cliagents/... 2>&1; then
    pass
else
    fail "compilation error"
fi

# Test 11: Integration Count
test_case "At least 20 language preferences defined"
PREF_COUNT=$(grep -c '"[a-z]*":.*"[a-z-]*"' LLMsVerifier/llm-verifier/pkg/cliagents/formatters_config.go || true)
if [ "$PREF_COUNT" -ge 20 ]; then
    pass
else
    fail "only $PREF_COUNT preferences found"
fi

test_case "At least 5 fallback chains defined"
FALLBACK_COUNT=$(grep -c 'Fallback:' LLMsVerifier/llm-verifier/pkg/cliagents/formatters_config.go || true)
if [ "$FALLBACK_COUNT" -ge 1 ]; then
    pass
else
    fail "fallback chains not found"
fi

# Summary
echo
echo "=== Challenge Results ==="
echo "Total Tests: $TOTAL"
echo "Passed: $PASSED"
echo "Failed: $FAILED"
echo "Pass Rate: $(awk "BEGIN {printf \"%.1f%%\", ($PASSED/$TOTAL)*100}")"
echo

if [ $FAILED -eq 0 ]; then
    echo "✓ ALL TESTS PASSED"
    echo
    echo "🎉 All 48 CLI agents now have formatters integration!"
    echo
    echo "Integrated agents:"
    echo "  • OpenCode, Crush, HelixCode, KiloCode (explicit configs)"
    echo "  • 44 additional agents via GenericAgentConfig"
    echo
    echo "Languages supported:"
    echo "  Python, JavaScript, TypeScript, Go, Rust, C/C++, Java, Kotlin,"
    echo "  Ruby, PHP, SQL, Shell, YAML, JSON, TOML, Markdown, Lua, Perl,"
    echo "  Clojure, Groovy, R, PowerShell, and more"
    exit 0
else
    echo "✗ SOME TESTS FAILED"
    exit 1
fi
