#!/bin/bash
# HelixAgent Agent Configuration Generator
# Generates plugin configurations for all 48 CLI agents

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
OUTPUT_DIR="$PROJECT_ROOT/configs"

# Default HelixAgent endpoint
ENDPOINT="${HELIXAGENT_ENDPOINT:-https://localhost:7061}"

# Colors
GREEN='\033[0;32m'
CYAN='\033[0;36m'
NC='\033[0m'

# Agent definitions (48 agents total)
declare -A TIER1_AGENTS=(
    ["claude_code"]="Claude Code|TypeScript|hooks"
    ["opencode"]="OpenCode|Go|mcp"
    ["cline"]="Cline|TypeScript|hooks"
    ["kilo_code"]="Kilo-Code|TypeScript|package"
)

declare -A TIER2_AGENTS=(
    ["aider"]="Aider|Python|config"
    ["codename_goose"]="Codename Goose|TypeScript/Rust|mcp"
    ["forge"]="Forge|TypeScript|templates"
    ["amazon_q"]="Amazon Q|TypeScript|config"
    ["kiro"]="Kiro|TypeScript|config"
    ["gpt_engineer"]="GPT Engineer|Python|config"
    ["gemini_cli"]="Gemini CLI|TypeScript|config"
    ["deepseek_cli"]="DeepSeek CLI|TypeScript|config"
)

declare -A TIER3_AGENTS=(
    ["mistral_code"]="Mistral Code|TypeScript|mcp"
    ["ollama_code"]="Ollama Code|TypeScript|mcp"
    ["plandex"]="Plandex|Go|mcp"
    ["qwen_code"]="Qwen Code|TypeScript|mcp"
    ["crush"]="Crush|TypeScript|mcp"
    ["helix_code"]="HelixCode|Go|mcp"
    ["cursor"]="Cursor|TypeScript|mcp"
    ["windsurf"]="Windsurf|TypeScript|mcp"
    ["zed"]="Zed|Rust|mcp"
    ["continue"]="Continue|TypeScript|mcp"
    ["tabby"]="Tabby|Rust|mcp"
    ["sourcegraph_cody"]="Sourcegraph Cody|TypeScript|mcp"
    ["github_copilot"]="GitHub Copilot|TypeScript|mcp"
    ["tabnine"]="Tabnine|TypeScript|mcp"
    ["codeium"]="Codeium|TypeScript|mcp"
    ["replit_ghostwriter"]="Replit Ghostwriter|TypeScript|mcp"
    ["aws_codewhisperer"]="AWS CodeWhisperer|TypeScript|mcp"
    ["jetbrains_ai"]="JetBrains AI|Kotlin|mcp"
    ["intellicode"]="IntelliCode|TypeScript|mcp"
    ["codegpt"]="CodeGPT|TypeScript|mcp"
    ["aicommit"]="AI Commit|TypeScript|mcp"
    ["grit"]="Grit|TypeScript|mcp"
    ["mentat"]="Mentat|Python|mcp"
    ["smol_developer"]="Smol Developer|Python|mcp"
    ["gpt_pilot"]="GPT Pilot|Python|mcp"
    ["auto_gpt"]="Auto-GPT|Python|mcp"
    ["agent_gpt"]="AgentGPT|TypeScript|mcp"
    ["superagi"]="SuperAGI|Python|mcp"
    ["babyagi"]="BabyAGI|Python|mcp"
    ["langchain_agents"]="LangChain Agents|Python|mcp"
    ["autogen"]="AutoGen|Python|mcp"
    ["crewai"]="CrewAI|Python|mcp"
    ["taskweaver"]="TaskWeaver|Python|mcp"
    ["devika"]="Devika|Python|mcp"
    ["open_interpreter"]="Open Interpreter|Python|mcp"
    ["shell_gpt"]="Shell GPT|Python|mcp"
)

# Create output directory
mkdir -p "$OUTPUT_DIR"

echo "=============================================="
echo "HelixAgent Agent Configuration Generator"
echo "=============================================="
echo ""
echo "Endpoint: $ENDPOINT"
echo "Output: $OUTPUT_DIR"
echo ""

# Generate MCP configuration for an agent
generate_mcp_config() {
    local agent_id="$1"
    local agent_name="$2"
    local language="$3"

    cat > "$OUTPUT_DIR/${agent_id}_mcp.json" << EOF
{
  "\$schema": "https://schema.helixagent.ai/plugin/v1.json",
  "name": "helixagent-${agent_id}",
  "description": "HelixAgent integration for ${agent_name}",
  "endpoint": "${ENDPOINT}",
  "transport": {
    "preferHTTP3": true,
    "enableTOON": true,
    "enableBrotli": true,
    "timeout": 30000
  },
  "events": {
    "transport": "sse",
    "subscribeToDebates": true,
    "subscribeToTasks": true,
    "reconnectInterval": 5000
  },
  "ui": {
    "renderStyle": "theater",
    "progressStyle": "unicode",
    "colorScheme": "256"
  },
  "debate": {
    "showPhaseIndicators": true,
    "showConfidenceScores": true,
    "enableMultiPassValidation": true
  },
  "mcpServer": {
    "command": "npx",
    "args": ["@helixagent/mcp-server", "--endpoint", "${ENDPOINT}"],
    "transport": "stdio"
  }
}
EOF
    echo -e "  ${GREEN}✓${NC} Generated: ${agent_id}_mcp.json"
}

# Generate Python config (for Aider-style agents)
generate_python_config() {
    local agent_id="$1"
    local agent_name="$2"

    cat > "$OUTPUT_DIR/${agent_id}_config.py" << EOF
"""
HelixAgent configuration for ${agent_name}

Add to your ${agent_name} configuration or import this module.
"""

HELIXAGENT_CONFIG = {
    "endpoint": "${ENDPOINT}",
    "transport": {
        "prefer_http3": True,
        "enable_toon": True,
        "enable_brotli": True,
        "timeout": 30,
    },
    "events": {
        "transport": "sse",
        "subscribe_to_debates": True,
        "subscribe_to_tasks": True,
    },
    "debate": {
        "show_phase_indicators": True,
        "show_confidence_scores": True,
        "enable_multi_pass_validation": True,
    },
}

# LiteLLM-compatible provider configuration
LITELLM_HELIXAGENT = {
    "model": "openai/helix-debate-ensemble",
    "api_base": "${ENDPOINT}/v1",
    "api_key": "helixagent",  # Use actual key if required
}
EOF
    echo -e "  ${GREEN}✓${NC} Generated: ${agent_id}_config.py"
}

# Generate generic config
generate_generic_config() {
    local agent_id="$1"
    local agent_name="$2"

    cat > "$OUTPUT_DIR/${agent_id}_config.json" << EOF
{
  "\$schema": "https://schema.helixagent.ai/plugin/v1.json",
  "name": "helixagent-${agent_id}",
  "description": "HelixAgent integration for ${agent_name}",
  "endpoint": "${ENDPOINT}",
  "transport": {
    "preferHTTP3": true,
    "enableTOON": true,
    "enableBrotli": true,
    "timeout": 30000
  },
  "events": {
    "transport": "sse",
    "subscribeToDebates": true,
    "subscribeToTasks": true
  },
  "ui": {
    "renderStyle": "theater",
    "progressStyle": "unicode"
  },
  "debate": {
    "showPhaseIndicators": true,
    "showConfidenceScores": true
  },
  "provider": {
    "type": "openai-compatible",
    "baseUrl": "${ENDPOINT}/v1",
    "model": "helix-debate-ensemble"
  }
}
EOF
    echo -e "  ${GREEN}✓${NC} Generated: ${agent_id}_config.json"
}

# Generate Tier 1 configs (already have full plugins)
echo -e "${CYAN}--- Tier 1 Agents (Full Plugin Support) ---${NC}"
for agent_id in "${!TIER1_AGENTS[@]}"; do
    IFS='|' read -r agent_name language config_type <<< "${TIER1_AGENTS[$agent_id]}"
    echo "  $agent_name - Using dedicated plugin in plugins/agents/$agent_id/"
done
echo ""

# Generate Tier 2 configs
echo -e "${CYAN}--- Tier 2 Agents (MCP + Config) ---${NC}"
for agent_id in "${!TIER2_AGENTS[@]}"; do
    IFS='|' read -r agent_name language config_type <<< "${TIER2_AGENTS[$agent_id]}"

    if [[ "$language" == *"Python"* ]]; then
        generate_python_config "$agent_id" "$agent_name"
    else
        generate_mcp_config "$agent_id" "$agent_name" "$language"
    fi
done
echo ""

# Generate Tier 3 configs
echo -e "${CYAN}--- Tier 3 Agents (Generic MCP Server) ---${NC}"
for agent_id in "${!TIER3_AGENTS[@]}"; do
    IFS='|' read -r agent_name language config_type <<< "${TIER3_AGENTS[$agent_id]}"

    if [[ "$language" == *"Python"* ]]; then
        generate_python_config "$agent_id" "$agent_name"
    else
        generate_mcp_config "$agent_id" "$agent_name" "$language"
    fi
done
echo ""

# Count generated files
GENERATED=$(ls -1 "$OUTPUT_DIR"/*.json "$OUTPUT_DIR"/*.py 2>/dev/null | wc -l)

echo "=============================================="
echo "Configuration Generation Complete"
echo "=============================================="
echo ""
echo "Generated $GENERATED configuration files"
echo "Output directory: $OUTPUT_DIR"
echo ""
echo "To use with an agent:"
echo "  1. Copy the appropriate config to your agent's config directory"
echo "  2. For MCP-based configs, ensure @helixagent/mcp-server is installed:"
echo "     npm install -g @helixagent/mcp-server"
echo "  3. Start HelixAgent server at $ENDPOINT"
echo ""
