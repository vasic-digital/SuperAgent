#!/bin/bash
# HelixAgent Challenge: CLI Agent Configuration Generation System
# Tests: ~60 tests across 9 sections
# Validates: Generator infrastructure, MCP servers, plugins, extensions, skills,
#            custom generator completeness, generic agent completeness, test coverage,
#            and functional validation via go test

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHALLENGES_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$CHALLENGES_DIR")"

CLIAGENTS_DIR="$PROJECT_ROOT/LLMsVerifier/llm-verifier/pkg/cliagents"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PASSED=0
FAILED=0
TOTAL=0

pass() {
    PASSED=$((PASSED + 1))
    TOTAL=$((TOTAL + 1))
    echo -e "  ${GREEN}[PASS]${NC} $1"
}

fail() {
    FAILED=$((FAILED + 1))
    TOTAL=$((TOTAL + 1))
    echo -e "  ${RED}[FAIL]${NC} $1"
}

section() {
    echo -e "\n${BLUE}=== $1 ===${NC}"
}

echo -e "${BLUE}============================================================${NC}"
echo -e "${BLUE} HelixAgent Challenge: CLI Agent Config Generation System${NC}"
echo -e "${BLUE}============================================================${NC}"
echo -e "CLI Agents Directory: $CLIAGENTS_DIR"

#===============================================================================
# Section 1: Generator Infrastructure (5 tests)
#===============================================================================
section "Section 1: Generator Infrastructure"

# Test 1.1: generator.go exists
if [ -f "$CLIAGENTS_DIR/generator.go" ]; then
    pass "generator.go exists"
else
    fail "generator.go does not exist"
fi

# Test 1.2: additional_agents.go exists
if [ -f "$CLIAGENTS_DIR/additional_agents.go" ]; then
    pass "additional_agents.go exists"
else
    fail "additional_agents.go does not exist"
fi

# Test 1.3: formatters_config.go exists
if [ -f "$CLIAGENTS_DIR/formatters_config.go" ]; then
    pass "formatters_config.go exists"
else
    fail "formatters_config.go does not exist"
fi

# Test 1.4: 48 AgentType constants exist in generator.go
AGENT_TYPE_COUNT=$(grep -c 'Agent[A-Z][A-Za-z]*\s*AgentType\s*=' "$CLIAGENTS_DIR/generator.go" 2>/dev/null || echo "0")
if [ "$AGENT_TYPE_COUNT" -ge 48 ]; then
    pass "48 AgentType constants exist in generator.go (found: $AGENT_TYPE_COUNT)"
else
    fail "Expected 48 AgentType constants in generator.go, found: $AGENT_TYPE_COUNT"
fi

# Test 1.5: SupportedAgents has 48 entries
SUPPORTED_ENTRIES=$(sed -n '/var SupportedAgents/,/^}/p' "$CLIAGENTS_DIR/generator.go" | grep -c '"' 2>/dev/null || echo "0")
if [ "$SUPPORTED_ENTRIES" -ge 48 ]; then
    pass "SupportedAgents has 48 entries (found: $SUPPORTED_ENTRIES)"
else
    fail "Expected SupportedAgents to have 48 entries, found: $SUPPORTED_ENTRIES"
fi

#===============================================================================
# Section 2: Default MCP Servers (6 tests)
#===============================================================================
section "Section 2: Default MCP Servers"

# Test 2.1: DefaultMCPServers function exists
if grep -q 'func DefaultMCPServers()' "$CLIAGENTS_DIR/generator.go"; then
    pass "DefaultMCPServers function exists"
else
    fail "DefaultMCPServers function does not exist"
fi

# Test 2.2: DefaultMCPServersForHost function exists
if grep -q 'func DefaultMCPServersForHost(' "$CLIAGENTS_DIR/generator.go"; then
    pass "DefaultMCPServersForHost function exists"
else
    fail "DefaultMCPServersForHost function does not exist"
fi

# Test 2.3: At least 6 HelixAgent remote endpoints
if grep -q '"helixagent-mcp"' "$CLIAGENTS_DIR/generator.go" && \
   grep -q '"helixagent-acp"' "$CLIAGENTS_DIR/generator.go" && \
   grep -q '"helixagent-lsp"' "$CLIAGENTS_DIR/generator.go" && \
   grep -q '"helixagent-embeddings"' "$CLIAGENTS_DIR/generator.go" && \
   grep -q '"helixagent-vision"' "$CLIAGENTS_DIR/generator.go" && \
   grep -q '"helixagent-cognee"' "$CLIAGENTS_DIR/generator.go"; then
    pass "At least 6 HelixAgent remote endpoints (mcp, acp, lsp, embeddings, vision, cognee)"
else
    fail "Missing HelixAgent remote endpoints"
fi

# Test 2.4: At least 3 extended HelixAgent endpoints
if grep -q '"helixagent-rag"' "$CLIAGENTS_DIR/generator.go" && \
   grep -q '"helixagent-formatters"' "$CLIAGENTS_DIR/generator.go" && \
   grep -q '"helixagent-monitoring"' "$CLIAGENTS_DIR/generator.go"; then
    pass "At least 3 extended HelixAgent endpoints (rag, formatters, monitoring)"
else
    fail "Missing extended HelixAgent endpoints"
fi

# Test 2.5: At least 6 local npx MCP servers
if grep -q '"filesystem"' "$CLIAGENTS_DIR/generator.go" && \
   grep -q '"memory"' "$CLIAGENTS_DIR/generator.go" && \
   grep -q '"sequential-thinking"' "$CLIAGENTS_DIR/generator.go" && \
   grep -q '"everything"' "$CLIAGENTS_DIR/generator.go" && \
   grep -q '"puppeteer"' "$CLIAGENTS_DIR/generator.go" && \
   grep -q '"sqlite"' "$CLIAGENTS_DIR/generator.go"; then
    pass "At least 6 local npx MCP servers (filesystem, memory, sequential-thinking, everything, puppeteer, sqlite)"
else
    fail "Missing local npx MCP servers"
fi

# Test 2.6: At least 3 free remote MCP servers
if grep -q '"context7"' "$CLIAGENTS_DIR/generator.go" && \
   grep -q '"deepwiki"' "$CLIAGENTS_DIR/generator.go" && \
   grep -q '"cloudflare-docs"' "$CLIAGENTS_DIR/generator.go"; then
    pass "At least 3 free remote MCP servers (context7, deepwiki, cloudflare-docs)"
else
    fail "Missing free remote MCP servers"
fi

#===============================================================================
# Section 3: Plugins System (4 tests)
#===============================================================================
section "Section 3: Plugins System"

# Test 3.1: DefaultPlugins function exists
if grep -q 'func DefaultPlugins()' "$CLIAGENTS_DIR/generator.go"; then
    pass "DefaultPlugins function exists"
else
    fail "DefaultPlugins function does not exist"
fi

# Test 3.2: At least 10 plugins defined
PLUGIN_COUNT=$(sed -n '/func DefaultPlugins/,/^}/p' "$CLIAGENTS_DIR/generator.go" | grep -c '"helixagent-' 2>/dev/null || echo "0")
if [ "$PLUGIN_COUNT" -ge 10 ]; then
    pass "At least 10 plugins defined (found: $PLUGIN_COUNT)"
else
    fail "Expected at least 10 plugins, found: $PLUGIN_COUNT"
fi

# Test 3.3: Key plugins present
if grep -q '"helixagent-mcp"' "$CLIAGENTS_DIR/generator.go" && \
   grep -q '"helixagent-lsp"' "$CLIAGENTS_DIR/generator.go" && \
   grep -q '"helixagent-acp"' "$CLIAGENTS_DIR/generator.go" && \
   grep -q '"helixagent-embeddings"' "$CLIAGENTS_DIR/generator.go" && \
   grep -q '"helixagent-rag"' "$CLIAGENTS_DIR/generator.go"; then
    pass "Key plugins present (helixagent-mcp, helixagent-lsp, helixagent-acp, helixagent-embeddings, helixagent-rag)"
else
    fail "Missing key plugins"
fi

# Test 3.4: Plugins are referenced from all config structs
PLUGIN_REFS=0
# OpenCode uses Plugin (singular)
if grep -q 'Plugin.*\[\]string' "$CLIAGENTS_DIR/opencode.go"; then
    PLUGIN_REFS=$((PLUGIN_REFS + 1))
fi
# Crush uses Plugins
if grep -q 'Plugins.*\[\]string' "$CLIAGENTS_DIR/crush.go"; then
    PLUGIN_REFS=$((PLUGIN_REFS + 1))
fi
# HelixCode uses Plugins
if grep -q 'Plugins.*\[\]string' "$CLIAGENTS_DIR/helixcode.go"; then
    PLUGIN_REFS=$((PLUGIN_REFS + 1))
fi
# KiloCode uses Plugins
if grep -q 'Plugins.*\[\]string' "$CLIAGENTS_DIR/kilocode.go"; then
    PLUGIN_REFS=$((PLUGIN_REFS + 1))
fi
# GenericAgentConfig uses Plugins
if grep -q 'Plugins.*\[\]string' "$CLIAGENTS_DIR/additional_agents.go"; then
    PLUGIN_REFS=$((PLUGIN_REFS + 1))
fi
if [ "$PLUGIN_REFS" -ge 5 ]; then
    pass "Plugins referenced from all config structs (OpenCode, Crush, HelixCode, KiloCode, GenericAgent)"
else
    fail "Plugins not referenced from all config structs (found: $PLUGIN_REFS/5)"
fi

#===============================================================================
# Section 4: Extensions System (6 tests)
#===============================================================================
section "Section 4: Extensions System"

# Test 4.1: HelixAgentExtensions struct exists
if grep -q 'type HelixAgentExtensions struct' "$CLIAGENTS_DIR/generator.go"; then
    pass "HelixAgentExtensions struct exists"
else
    fail "HelixAgentExtensions struct does not exist"
fi

# Test 4.2: LSPConfig struct exists with Enabled and Endpoint fields
if grep -q 'type LSPConfig struct' "$CLIAGENTS_DIR/generator.go" && \
   grep -A5 'type LSPConfig struct' "$CLIAGENTS_DIR/generator.go" | grep -q 'Enabled' && \
   grep -A5 'type LSPConfig struct' "$CLIAGENTS_DIR/generator.go" | grep -q 'Endpoint'; then
    pass "LSPConfig struct exists with Enabled and Endpoint fields"
else
    fail "LSPConfig struct missing or incomplete"
fi

# Test 4.3: ACPConfig struct exists with Enabled and Endpoint fields
if grep -q 'type ACPConfig struct' "$CLIAGENTS_DIR/generator.go" && \
   grep -A5 'type ACPConfig struct' "$CLIAGENTS_DIR/generator.go" | grep -q 'Enabled' && \
   grep -A5 'type ACPConfig struct' "$CLIAGENTS_DIR/generator.go" | grep -q 'Endpoint'; then
    pass "ACPConfig struct exists with Enabled and Endpoint fields"
else
    fail "ACPConfig struct missing or incomplete"
fi

# Test 4.4: EmbeddingsConfig struct exists with Enabled and Endpoint fields
if grep -q 'type EmbeddingsConfig struct' "$CLIAGENTS_DIR/generator.go" && \
   grep -A5 'type EmbeddingsConfig struct' "$CLIAGENTS_DIR/generator.go" | grep -q 'Enabled' && \
   grep -A5 'type EmbeddingsConfig struct' "$CLIAGENTS_DIR/generator.go" | grep -q 'Endpoint'; then
    pass "EmbeddingsConfig struct exists with Enabled and Endpoint fields"
else
    fail "EmbeddingsConfig struct missing or incomplete"
fi

# Test 4.5: RAGConfig struct exists with Enabled and Endpoint fields
if grep -q 'type RAGConfig struct' "$CLIAGENTS_DIR/generator.go" && \
   grep -A5 'type RAGConfig struct' "$CLIAGENTS_DIR/generator.go" | grep -q 'Enabled' && \
   grep -A5 'type RAGConfig struct' "$CLIAGENTS_DIR/generator.go" | grep -q 'Endpoint'; then
    pass "RAGConfig struct exists with Enabled and Endpoint fields"
else
    fail "RAGConfig struct missing or incomplete"
fi

# Test 4.6: SkillConfig struct exists with Name, Description, Endpoint, Enabled fields
if grep -q 'type SkillConfig struct' "$CLIAGENTS_DIR/generator.go" && \
   grep -A10 'type SkillConfig struct' "$CLIAGENTS_DIR/generator.go" | grep -q 'Name' && \
   grep -A10 'type SkillConfig struct' "$CLIAGENTS_DIR/generator.go" | grep -q 'Description' && \
   grep -A10 'type SkillConfig struct' "$CLIAGENTS_DIR/generator.go" | grep -q 'Endpoint' && \
   grep -A10 'type SkillConfig struct' "$CLIAGENTS_DIR/generator.go" | grep -q 'Enabled'; then
    pass "SkillConfig struct exists with Name, Description, Endpoint, Enabled fields"
else
    fail "SkillConfig struct missing or incomplete"
fi

#===============================================================================
# Section 5: Skills System (3 tests)
#===============================================================================
section "Section 5: Skills System"

# Test 5.1: DefaultSkills function exists
if grep -q 'func DefaultSkills(' "$CLIAGENTS_DIR/generator.go"; then
    pass "DefaultSkills function exists"
else
    fail "DefaultSkills function does not exist"
fi

# Test 5.2: At least 8 skills defined
SKILL_COUNT=$(sed -n '/func DefaultSkills/,/^}/p' "$CLIAGENTS_DIR/generator.go" | grep -c 'Name:' 2>/dev/null || echo "0")
if [ "$SKILL_COUNT" -ge 8 ]; then
    pass "At least 8 skills defined (found: $SKILL_COUNT)"
else
    fail "Expected at least 8 skills, found: $SKILL_COUNT"
fi

# Test 5.3: Key skills present
SKILLS_BLOCK=$(sed -n '/func DefaultSkills/,/^}/p' "$CLIAGENTS_DIR/generator.go")
SKILLS_FOUND=0
for SKILL in "code-review" "code-format" "semantic-search" "vision-analysis" "memory-recall" "rag-retrieval" "lsp-diagnostics" "agent-communication"; do
    if echo "$SKILLS_BLOCK" | grep -q "\"$SKILL\""; then
        SKILLS_FOUND=$((SKILLS_FOUND + 1))
    fi
done
if [ "$SKILLS_FOUND" -ge 8 ]; then
    pass "Key skills present (code-review, code-format, semantic-search, vision-analysis, memory-recall, rag-retrieval, lsp-diagnostics, agent-communication)"
else
    fail "Missing key skills ($SKILLS_FOUND/8 found)"
fi

#===============================================================================
# Section 6: Custom Generator Completeness (16 tests - 4 per custom agent)
#===============================================================================
section "Section 6: Custom Generator Completeness"

# --- OpenCode (4 tests) ---

# Test 6.1: OpenCodeConfig struct has Plugin/Plugins field
if grep -q 'Plugin.*\[\]string' "$CLIAGENTS_DIR/opencode.go"; then
    pass "OpenCodeConfig struct has Plugin field"
else
    fail "OpenCodeConfig struct missing Plugin field"
fi

# Test 6.2: OpenCodeConfig struct has Extensions field
if grep -q 'Extensions.*\*HelixAgentExtensions' "$CLIAGENTS_DIR/opencode.go"; then
    pass "OpenCodeConfig struct has Extensions field"
else
    fail "OpenCodeConfig struct missing Extensions field"
fi

# Test 6.3: OpenCode Generate function sets Plugins
if grep -q 'openCodeConfig\.Plugin\s*=' "$CLIAGENTS_DIR/opencode.go" || \
   grep -q 'DefaultPlugins' "$CLIAGENTS_DIR/opencode.go"; then
    pass "OpenCode Generate function sets Plugins"
else
    fail "OpenCode Generate function does not set Plugins"
fi

# Test 6.4: OpenCode Generate function sets Extensions
if grep -q 'DefaultHelixAgentExtensions' "$CLIAGENTS_DIR/opencode.go"; then
    pass "OpenCode Generate function sets Extensions (DefaultHelixAgentExtensions)"
else
    fail "OpenCode Generate function does not set Extensions"
fi

# --- Crush (4 tests) ---

# Test 6.5: CrushConfig struct has Plugins field
if grep -q 'Plugins.*\[\]string' "$CLIAGENTS_DIR/crush.go"; then
    pass "CrushConfig struct has Plugins field"
else
    fail "CrushConfig struct missing Plugins field"
fi

# Test 6.6: CrushConfig struct has Extensions field
if grep -q 'Extensions.*\*HelixAgentExtensions' "$CLIAGENTS_DIR/crush.go"; then
    pass "CrushConfig struct has Extensions field"
else
    fail "CrushConfig struct missing Extensions field"
fi

# Test 6.7: Crush Generate function sets Plugins
if grep -q 'DefaultPlugins' "$CLIAGENTS_DIR/crush.go"; then
    pass "Crush Generate function sets Plugins (DefaultPlugins)"
else
    fail "Crush Generate function does not set Plugins via DefaultPlugins"
fi

# Test 6.8: Crush Generate function sets Extensions
if grep -q 'DefaultHelixAgentExtensions' "$CLIAGENTS_DIR/crush.go"; then
    pass "Crush Generate function sets Extensions (DefaultHelixAgentExtensions)"
else
    fail "Crush Generate function does not set Extensions"
fi

# --- HelixCode (4 tests) ---

# Test 6.9: HelixCodeConfig struct has Plugins field
if grep -q 'Plugins.*\[\]string' "$CLIAGENTS_DIR/helixcode.go"; then
    pass "HelixCodeConfig struct has Plugins field"
else
    fail "HelixCodeConfig struct missing Plugins field"
fi

# Test 6.10: HelixCodeConfig struct has Extensions field
if grep -q 'Extensions.*\*HelixAgentExtensions' "$CLIAGENTS_DIR/helixcode.go"; then
    pass "HelixCodeConfig struct has Extensions field"
else
    fail "HelixCodeConfig struct missing Extensions field"
fi

# Test 6.11: HelixCode Generate function sets Plugins
if grep -q 'DefaultPlugins' "$CLIAGENTS_DIR/helixcode.go"; then
    pass "HelixCode Generate function sets Plugins (DefaultPlugins)"
else
    fail "HelixCode Generate function does not set Plugins via DefaultPlugins"
fi

# Test 6.12: HelixCode Generate function sets Extensions
if grep -q 'DefaultHelixAgentExtensions' "$CLIAGENTS_DIR/helixcode.go"; then
    pass "HelixCode Generate function sets Extensions (DefaultHelixAgentExtensions)"
else
    fail "HelixCode Generate function does not set Extensions"
fi

# --- KiloCode (4 tests) ---

# Test 6.13: KiloCodeConfig struct has Plugins field
if grep -q 'Plugins.*\[\]string' "$CLIAGENTS_DIR/kilocode.go"; then
    pass "KiloCodeConfig struct has Plugins field"
else
    fail "KiloCodeConfig struct missing Plugins field"
fi

# Test 6.14: KiloCodeConfig struct has Extensions field
if grep -q 'Extensions.*\*HelixAgentExtensions' "$CLIAGENTS_DIR/kilocode.go"; then
    pass "KiloCodeConfig struct has Extensions field"
else
    fail "KiloCodeConfig struct missing Extensions field"
fi

# Test 6.15: KiloCode Generate function sets Plugins
if grep -q 'DefaultPlugins' "$CLIAGENTS_DIR/kilocode.go"; then
    pass "KiloCode Generate function sets Plugins (DefaultPlugins)"
else
    fail "KiloCode Generate function does not set Plugins via DefaultPlugins"
fi

# Test 6.16: KiloCode Generate function sets Extensions
if grep -q 'DefaultHelixAgentExtensions' "$CLIAGENTS_DIR/kilocode.go"; then
    pass "KiloCode Generate function sets Extensions (DefaultHelixAgentExtensions)"
else
    fail "KiloCode Generate function does not set Extensions"
fi

#===============================================================================
# Section 7: Generic Agent Completeness (6 tests)
#===============================================================================
section "Section 7: Generic Agent Completeness"

# Test 7.1: GenericAgentConfig struct has Plugins field
if grep -q 'type GenericAgentConfig struct' "$CLIAGENTS_DIR/additional_agents.go" && \
   grep -A20 'type GenericAgentConfig struct' "$CLIAGENTS_DIR/additional_agents.go" | grep -q 'Plugins.*\[\]string'; then
    pass "GenericAgentConfig struct has Plugins field"
else
    fail "GenericAgentConfig struct missing Plugins field"
fi

# Test 7.2: GenericAgentConfig struct has Extensions field
if grep -A20 'type GenericAgentConfig struct' "$CLIAGENTS_DIR/additional_agents.go" | grep -q 'Extensions.*\*HelixAgentExtensions'; then
    pass "GenericAgentConfig struct has Extensions field"
else
    fail "GenericAgentConfig struct missing Extensions field"
fi

# Test 7.3: Generic Generate() sets Plugins
if grep -q 'DefaultPlugins' "$CLIAGENTS_DIR/additional_agents.go"; then
    pass "Generic Generate() sets Plugins via DefaultPlugins"
else
    fail "Generic Generate() does not set Plugins via DefaultPlugins"
fi

# Test 7.4: Generic Generate() sets Extensions
if grep -q 'DefaultHelixAgentExtensions' "$CLIAGENTS_DIR/additional_agents.go"; then
    pass "Generic Generate() sets Extensions via DefaultHelixAgentExtensions"
else
    fail "Generic Generate() does not set Extensions via DefaultHelixAgentExtensions"
fi

# Test 7.5: API key uses os.Getenv
if grep -q 'os\.Getenv' "$CLIAGENTS_DIR/additional_agents.go"; then
    pass "API key uses os.Getenv for environment variable lookup"
else
    fail "API key does not use os.Getenv"
fi

# Test 7.6: Model capabilities include "rag"
if grep -q '"rag"' "$CLIAGENTS_DIR/additional_agents.go"; then
    pass "Model capabilities include 'rag'"
else
    fail "Model capabilities missing 'rag'"
fi

#===============================================================================
# Section 8: Test Coverage (6 tests)
#===============================================================================
section "Section 8: Test Coverage"

# Test 8.1: generator_test.go exists
if [ -f "$CLIAGENTS_DIR/generator_test.go" ]; then
    pass "generator_test.go exists"
else
    fail "generator_test.go does not exist"
fi

# Test 8.2: TestAllAgentsCompleteness function exists
if grep -q 'func TestAllAgentsCompleteness' "$CLIAGENTS_DIR/generator_test.go"; then
    pass "TestAllAgentsCompleteness function exists"
else
    fail "TestAllAgentsCompleteness function does not exist"
fi

# Test 8.3: TestDefaultMCPServers function exists
if grep -q 'func TestDefaultMCPServers' "$CLIAGENTS_DIR/generator_test.go"; then
    pass "TestDefaultMCPServers function exists"
else
    fail "TestDefaultMCPServers function does not exist"
fi

# Test 8.4: TestDefaultPlugins function exists
if grep -q 'func TestDefaultPlugins' "$CLIAGENTS_DIR/generator_test.go"; then
    pass "TestDefaultPlugins function exists"
else
    fail "TestDefaultPlugins function does not exist"
fi

# Test 8.5: TestDefaultSkills function exists
if grep -q 'func TestDefaultSkills' "$CLIAGENTS_DIR/generator_test.go"; then
    pass "TestDefaultSkills function exists"
else
    fail "TestDefaultSkills function does not exist"
fi

# Test 8.6: TestDefaultHelixAgentExtensions function exists
if grep -q 'func TestDefaultHelixAgentExtensions' "$CLIAGENTS_DIR/generator_test.go"; then
    pass "TestDefaultHelixAgentExtensions function exists"
else
    fail "TestDefaultHelixAgentExtensions function does not exist"
fi

#===============================================================================
# Section 9: Functional Validation (8 tests)
#===============================================================================
section "Section 9: Functional Validation"

VERIFIER_ROOT="$PROJECT_ROOT/LLMsVerifier/llm-verifier"

# Test 9.1: TestDefaultMCPServers passes
if cd "$VERIFIER_ROOT" && go test ./pkg/cliagents/ -run TestDefaultMCPServers -v > /dev/null 2>&1; then
    pass "go test TestDefaultMCPServers passes"
else
    fail "go test TestDefaultMCPServers failed"
fi

# Test 9.2: TestDefaultPlugins passes
if cd "$VERIFIER_ROOT" && go test ./pkg/cliagents/ -run TestDefaultPlugins -v > /dev/null 2>&1; then
    pass "go test TestDefaultPlugins passes"
else
    fail "go test TestDefaultPlugins failed"
fi

# Test 9.3: TestDefaultSkills passes
if cd "$VERIFIER_ROOT" && go test ./pkg/cliagents/ -run TestDefaultSkills -v > /dev/null 2>&1; then
    pass "go test TestDefaultSkills passes"
else
    fail "go test TestDefaultSkills failed"
fi

# Test 9.4: TestDefaultHelixAgentExtensions passes
if cd "$VERIFIER_ROOT" && go test ./pkg/cliagents/ -run TestDefaultHelixAgentExtensions -v > /dev/null 2>&1; then
    pass "go test TestDefaultHelixAgentExtensions passes"
else
    fail "go test TestDefaultHelixAgentExtensions failed"
fi

# Test 9.5: TestGenerateOpenCode passes
if cd "$VERIFIER_ROOT" && go test ./pkg/cliagents/ -run TestGenerateOpenCode -v > /dev/null 2>&1; then
    pass "go test TestGenerateOpenCode passes"
else
    fail "go test TestGenerateOpenCode failed"
fi

# Test 9.6: TestGenerateCrush passes
if cd "$VERIFIER_ROOT" && go test ./pkg/cliagents/ -run TestGenerateCrush -v > /dev/null 2>&1; then
    pass "go test TestGenerateCrush passes"
else
    fail "go test TestGenerateCrush failed"
fi

# Test 9.7: TestGenericAgentGenerator passes
if cd "$VERIFIER_ROOT" && go test ./pkg/cliagents/ -run TestGenericAgentGenerator -v > /dev/null 2>&1; then
    pass "go test TestGenericAgentGenerator passes"
else
    fail "go test TestGenericAgentGenerator failed"
fi

# Test 9.8: TestAllAgentsCompleteness passes
if cd "$VERIFIER_ROOT" && go test ./pkg/cliagents/ -run TestAllAgentsCompleteness -v > /dev/null 2>&1; then
    pass "go test TestAllAgentsCompleteness passes"
else
    fail "go test TestAllAgentsCompleteness failed"
fi

#===============================================================================
# Summary
#===============================================================================
echo ""
echo -e "${BLUE}================================${NC}"
echo -e "Results: ${GREEN}$PASSED passed${NC}, ${RED}$FAILED failed${NC}, $TOTAL total"
echo -e "${BLUE}================================${NC}"

if [ $FAILED -gt 0 ]; then
    echo -e "\n${RED}CHALLENGE FAILED${NC}"
    exit 1
fi

echo -e "\n${GREEN}CHALLENGE PASSED${NC}"
exit 0
