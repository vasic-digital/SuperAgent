# CLI Agents Documentation

HelixAgent supports **47 CLI agents** with unified configuration generation, validation, and plugin integration.

## Documentation Status

All 47 CLI agents now have documentation:

| Agent | Documentation | Status |
|-------|--------------|--------|
| **Claude Code** | [View Docs](../cli_agents/claude-code/README.md) | ✅ Comprehensive (8 files, 3330 lines) |
| **OpenAI Codex** | [View Docs](../cli_agents/codex/README.md) | ✅ Comprehensive (7 files, 944 lines) |
| **Aider** | [View Docs](../cli_agents/aider/README.md) | ✅ Comprehensive (5 files, 1740 lines) |
| **OpenHands** | [View Docs](../cli_agents/openhands/README.md) | ✅ Comprehensive (5 files, 2014 lines) |
| **Gemini CLI** | [View Docs](../cli_agents/gemini-cli/README.md) | ✅ Comprehensive (5 files, ~2000 lines) |
| **Amazon Q** | [View Docs](../cli_agents/amazon-q/README.md) | ✅ Comprehensive (5 files, ~1800 lines) |
| **Agent Deck** | [View Docs](../cli_agents/agent-deck/README.md) | ✅ README |
| **Forge** | [View Docs](../cli_agents/forge/README.md) | ✅ README |
| **GPTMe** | [View Docs](../cli_agents/gptme/README.md) | ✅ README |
| **Kilo-Code** | [View Docs](../cli_agents/kilo-code/README.md) | ✅ README |
| **NanoCoder** | [View Docs](../cli_agents/nanocoder/README.md) | ✅ README |
| **TaskWeaver** | [View Docs](../cli_agents/taskweaver/README.md) | ✅ README |
| **Codename Goose** | [View Docs](../cli_agents/codename-goose/README.md) | ✅ README |
| **DeepSeek CLI** | [View Docs](../cli_agents/deepseek-cli/README.md) | ✅ README |
| **Spec Kit** | [View Docs](../cli_agents/spec-kit/README.md) | ✅ README |
| **Cline** | [View Docs](../cli_agents/cline/README.md) | ✅ README |
| **GPT-Engineer** | [View Docs](../cli_agents/gpt-engineer/README.md) | ✅ README |
| **OpenCode CLI** | [View Docs](../cli_agents/opencode-cli/README.md) | ✅ README |
| **Ollama Code** | [View Docs](../cli_agents/ollama-code/README.md) | ✅ README |
| **Fauxpilot** | [View Docs](../cli_agents/fauxpilot/README.md) | ✅ README |
| **Qwen Code** | [View Docs](../cli_agents/qwen-code/README.md) | ✅ README |
| **Mistral Code** | [View Docs](../cli_agents/mistral-code/README.md) | ✅ README |
| **Shai** | [View Docs](../cli_agents/shai/README.md) | ✅ README |
| **Snow CLI** | [View Docs](../cli_agents/snow-cli/README.md) | ✅ README |
| **Junie** | [View Docs](../cli_agents/junie/README.md) | ✅ README |
| **Kiro CLI** | [View Docs](../cli_agents/kiro-cli/README.md) | ✅ README |
| **Octogen** | [View Docs](../cli_agents/octogen/README.md) | ✅ README |
| **Noi** | [View Docs](../cli_agents/noi/README.md) | ✅ README |
| **Plandex** | [View Docs](../cli_agents/plandex/README.md) | ✅ README |
| **Postgres MCP** | [View Docs](../cli_agents/postgres-mcp/README.md) | ✅ README |
| **Git MCP** | [View Docs](../cli_agents/git-mcp/README.md) | ✅ README |
| **Codai** | [View Docs](../cli_agents/codai/README.md) | ✅ README |
| **MultiAgent Coding** | [View Docs](../cli_agents/multiagent-coding/README.md) | ✅ README |
| **Mobile Agent** | [View Docs](../cli_agents/mobile-agent/README.md) | ✅ README |
| **Superset** | [View Docs](../cli_agents/superset/README.md) | ✅ README |
| **Get Shit Done** | [View Docs](../cli_agents/get-shit-done/README.md) | ✅ README |
| **Bridle** | [View Docs](../cli_agents/bridle/README.md) | ✅ README |
| **Claude Plugins** | [View Docs](../cli_agents/claude-plugins/README.md) | ✅ README |
| **Claude Squad** | [View Docs](../cli_agents/claude-squad/README.md) | ✅ README |
| **Codex Skills** | [View Docs](../cli_agents/codex-skills/README.md) | ✅ README |
| **Conduit** | [View Docs](../cli_agents/conduit/README.md) | ✅ README |
| **Copilot CLI** | [View Docs](../cli_agents/copilot-cli/README.md) | ✅ README |
| **Crush** | [View Docs](../cli_agents/crush/README.md) | ✅ README |
| **UI UX Pro Max** | [View Docs](../cli_agents/ui-ux-pro-max/README.md) | ✅ README |
| **VTCode** | [View Docs](../cli_agents/vtcode/README.md) | ✅ README |
| **Warp** | [View Docs](../cli_agents/warp/README.md) | ✅ README |
| **Claude Code Source** | [View Docs](../cli_agents/claude-code-source/README.md) | ✅ README |

---

## Quick Links

### Generate Configurations

```bash
# List all 47 agents
./bin/helixagent --list-agents

# Generate for specific agent
./bin/helixagent --generate-agent-config=codex --agent-config-output=codex.json

# Generate all 47 at once
./bin/helixagent --generate-all-agents --all-agents-output-dir=~/agent-configs/
```

### Validate Configurations

```bash
./bin/helixagent --validate-agent-config=opencode:~/.config/opencode/opencode.json
```

---

## Agent Categories

| Category | Count | Examples |
|----------|-------|----------|
| Tier 1 (Major) | 8 | Claude Code, Codex, Aider, OpenHands, Gemini CLI, Amazon Q, GPT-Engineer, Cline |
| Tier 2 (Important) | 15 | Forge, GPTMe, Kilo-Code, NanoCoder, TaskWeaver, Qwen Code, Mistral Code, etc. |
| Tier 3 (Complete Set) | 24 | Remaining specialized agents |
| **Total** | **47** | |

---

## Support Matrix

| Feature | Tier 1 | Tier 2 | Tier 3 |
|---------|--------|--------|--------|
| Full Plugin Support | Yes | Partial | Config Only |
| HTTP/3 Transport | Yes | Yes | Via MCP |
| TOON Protocol | Yes | Yes | Via MCP |
| Event Streaming | Yes | Yes | Limited |
| UI Extensions | Yes | Limited | No |
