# CLI Agents Documentation

HelixAgent supports **47 CLI agents** with unified configuration generation, validation, and plugin integration.

## Documentation Status

All 47 CLI agents now have documentation:

| Agent | Documentation | Status |
|-------|--------------|--------|
| **Claude Code** | [View Docs](./claude-code/README.md) | ✅ Comprehensive (8 files, 3330 lines) |
| **OpenAI Codex** | [View Docs](./codex/README.md) | ✅ Comprehensive (7 files, 944 lines) |
| **Aider** | [View Docs](./aider/README.md) | ✅ Comprehensive (5 files, 1740 lines) |
| **OpenHands** | [View Docs](./openhands/README.md) | ✅ Comprehensive (5 files, 2014 lines) |
| **Gemini CLI** | [View Docs](./gemini-cli/README.md) | ✅ Comprehensive (5 files, ~2000 lines) |
| **Amazon Q** | [View Docs](./amazon-q/README.md) | ✅ Comprehensive (5 files, ~1800 lines) |
| **Agent Deck** | [View Docs](./agent-deck/README.md) | ✅ README |
| **Forge** | [View Docs](./forge/README.md) | ✅ README |
| **GPTMe** | [View Docs](./gptme/README.md) | ✅ README |
| **Kilo-Code** | [View Docs](./kilo-code/README.md) | ✅ README |
| **NanoCoder** | [View Docs](./nanocoder/README.md) | ✅ README |
| **TaskWeaver** | [View Docs](./taskweaver/README.md) | ✅ README |
| **Codename Goose** | [View Docs](./codename-goose/README.md) | ✅ README |
| **DeepSeek CLI** | [View Docs](./deepseek-cli/README.md) | ✅ README |
| **Spec Kit** | [View Docs](./spec-kit/README.md) | ✅ README |
| **Cline** | [View Docs](./cline/README.md) | ✅ README |
| **GPT-Engineer** | [View Docs](./gpt-engineer/README.md) | ✅ README |
| **OpenCode CLI** | [View Docs](./opencode-cli/README.md) | ✅ README |
| **Ollama Code** | [View Docs](./ollama-code/README.md) | ✅ README |
| **Fauxpilot** | [View Docs](./fauxpilot/README.md) | ✅ README |
| **Qwen Code** | [View Docs](./qwen-code/README.md) | ✅ README |
| **Mistral Code** | [View Docs](./mistral-code/README.md) | ✅ README |
| **Shai** | [View Docs](./shai/README.md) | ✅ README |
| **Snow CLI** | [View Docs](./snow-cli/README.md) | ✅ README |
| **Junie** | [View Docs](./junie/README.md) | ✅ README |
| **Kiro CLI** | [View Docs](./kiro-cli/README.md) | ✅ README |
| **Octogen** | [View Docs](./octogen/README.md) | ✅ README |
| **Noi** | [View Docs](./noi/README.md) | ✅ README |
| **Plandex** | [View Docs](./plandex/README.md) | ✅ README |
| **Postgres MCP** | [View Docs](./postgres-mcp/README.md) | ✅ README |
| **Git MCP** | [View Docs](./git-mcp/README.md) | ✅ README |
| **Codai** | [View Docs](./codai/README.md) | ✅ README |
| **MultiAgent Coding** | [View Docs](./multiagent-coding/README.md) | ✅ README |
| **Mobile Agent** | [View Docs](./mobile-agent/README.md) | ✅ README |
| **Superset** | [View Docs](./superset/README.md) | ✅ README |
| **Get Shit Done** | [View Docs](./get-shit-done/README.md) | ✅ README |
| **Bridle** | [View Docs](./bridle/README.md) | ✅ README |
| **Claude Plugins** | [View Docs](./claude-plugins/README.md) | ✅ README |
| **Claude Squad** | [View Docs](./claude-squad/README.md) | ✅ README |
| **Codex Skills** | [View Docs](./codex-skills/README.md) | ✅ README |
| **Conduit** | [View Docs](./conduit/README.md) | ✅ README |
| **Copilot CLI** | [View Docs](./copilot-cli/README.md) | ✅ README |
| **Crush** | [View Docs](./crush/README.md) | ✅ README |
| **UI UX Pro Max** | [View Docs](./ui-ux-pro-max/README.md) | ✅ README |
| **VTCode** | [View Docs](./vtcode/README.md) | ✅ README |
| **Warp** | [View Docs](./warp/README.md) | ✅ README |
| **Claude Code Source** | [View Docs](./claude-code-source/README.md) | ✅ README |

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
