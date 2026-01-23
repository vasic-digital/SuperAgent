# Agent Reference

Complete reference for all 48 supported CLI agents.

## Original 18 Agents

### OpenCode

| Property | Value |
|----------|-------|
| **Type** | `opencode` |
| **Config File** | `opencode.json` |
| **Config Directory** | `~/.config/opencode/` |
| **Language** | Go |
| **Tier** | 1 (Full Plugin Support) |
| **MCP Support** | Yes (stdio/SSE) |
| **Description** | OpenCode.ai CLI - AI-powered coding assistant |

**Key Features:**
- Native Go implementation
- MCP server support (stdio and SSE transports)
- LSP integration
- Multi-provider support

**Example Config:**
```json
{
  "provider": {
    "options": {
      "baseURL": "http://localhost:7061/v1",
      "models": [{"id": "helixagent-debate"}]
    }
  },
  "mcp": {
    "helixagent": {"type": "sse", "url": "http://localhost:7061/v1/mcp"}
  }
}
```

---

### Crush

| Property | Value |
|----------|-------|
| **Type** | `crush` |
| **Config File** | `crush.json` |
| **Config Directory** | `~/.config/crush/` |
| **Language** | Go |
| **Tier** | 1 (Custom Generator) |
| **Description** | Charm Land Crush CLI - AI coding assistant from Charm |

**Key Features:**
- Built by Charm (makers of Bubble Tea TUI framework)
- Beautiful terminal UI
- Multi-provider support

---

### HelixCode

| Property | Value |
|----------|-------|
| **Type** | `helixcode` |
| **Config File** | `helixcode.json` |
| **Config Directory** | `~/.config/helixcode/` |
| **Language** | Go |
| **Tier** | 1 (Native) |
| **Description** | HelixCode CLI - Native CLI for HelixAgent AI Debate Ensemble |

**Key Features:**
- Native HelixAgent integration
- Full AI Debate support
- All protocol support (MCP, ACP, LSP)

---

### Kiro

| Property | Value |
|----------|-------|
| **Type** | `kiro` |
| **Config File** | `kiro.json` |
| **Config Directory** | `~/.config/kiro/` |
| **Tier** | 2 |
| **Description** | Kiro - AI coding assistant |

---

### Aider

| Property | Value |
|----------|-------|
| **Type** | `aider` |
| **Config File** | `.aider.conf.yml` |
| **Config Directory** | `~/` (home directory) |
| **Language** | Python |
| **Tier** | 2 |
| **Description** | Aider CLI - AI pair programming in the terminal |

**Key Features:**
- Git-aware code editing
- Multiple edit formats (diff, whole, etc.)
- litellm provider support

**Example Config:**
```yaml
model: helixagent-debate
openai-api-base: http://localhost:7061/v1
auto-commits: true
edit-format: diff
stream: true
```

---

### Claude Code

| Property | Value |
|----------|-------|
| **Type** | `claude-code` |
| **Config File** | `settings.json` |
| **Config Directory** | `~/.claude/` |
| **Language** | TypeScript |
| **Tier** | 1 (Full Plugin Support) |
| **Description** | Claude Code - Anthropic's CLI for Claude |

**Key Features:**
- Official Anthropic CLI
- Plugin marketplace
- 2 lifecycle hooks (SessionStart, SessionEnd)
- Full permissions system

**Plugin Integration:**
```
~/.claude/plugins/helixagent-integration/
├── .claude-plugin/
│   └── plugin.json
├── hooks/
│   ├── session_start.js
│   └── session_end.js
└── lib/
    └── transport.js
```

---

### Cline

| Property | Value |
|----------|-------|
| **Type** | `cline` |
| **Config File** | `cline.json` |
| **Config Directory** | `~/.config/cline/` |
| **Language** | TypeScript |
| **Tier** | 1 (Full Plugin Support) |
| **Description** | Cline - AI coding assistant CLI |

**Key Features:**
- 8 lifecycle hooks
- MCP manager
- VS Code extension

**Lifecycle Hooks:**
1. `TaskStart` - Initialize HelixAgent session
2. `TaskResume` - Restore session
3. `TaskCancel` - Cleanup
4. `TaskComplete` - Final render
5. `UserPromptSubmit` - Intercept for TOON encoding
6. `PreToolUse` - Transform helix_* tool calls
7. `PostToolUse` - Render debate results
8. `PreCompact` - Save context

---

### Codename Goose

| Property | Value |
|----------|-------|
| **Type** | `codename-goose` |
| **Config File** | `goose.yaml` |
| **Config Directory** | `~/.config/goose/` |
| **Language** | TypeScript/Rust |
| **Tier** | 2 |
| **Description** | Codename Goose - Block's AI coding agent |

---

### DeepSeek CLI

| Property | Value |
|----------|-------|
| **Type** | `deepseek-cli` |
| **Config File** | `deepseek.json` |
| **Config Directory** | `~/.config/deepseek/` |
| **Tier** | 3 |
| **Description** | DeepSeek CLI - DeepSeek AI coding assistant |

---

### Forge

| Property | Value |
|----------|-------|
| **Type** | `forge` |
| **Config File** | `forge.yaml` |
| **Config Directory** | `~/.config/forge/` |
| **Language** | TypeScript |
| **Tier** | 2 |
| **Description** | Forge - AI-powered project scaffolding |

---

### Gemini CLI

| Property | Value |
|----------|-------|
| **Type** | `gemini-cli` |
| **Config File** | `gemini.json` |
| **Config Directory** | `~/.config/gemini/` |
| **Tier** | 2 |
| **Description** | Gemini CLI - Google's AI coding assistant |

---

### GPT Engineer

| Property | Value |
|----------|-------|
| **Type** | `gpt-engineer` |
| **Config File** | `gpt-engineer.toml` |
| **Config Directory** | `~/.config/gpt-engineer/` |
| **Language** | Python |
| **Tier** | 2 |
| **Description** | GPT Engineer - AI code generation from prompts |

---

### KiloCode

| Property | Value |
|----------|-------|
| **Type** | `kilocode` |
| **Config File** | `kilocode-settings.json` |
| **Config Directory** | `~/.config/kilocode/` |
| **Language** | TypeScript |
| **Tier** | 1 (Full Plugin Support) |
| **Description** | KiloCode VS Code extension - AI-powered code completion |

**Key Features:**
- Multi-platform monorepo (CLI, VS Code, JetBrains)
- MCP support
- Services architecture

---

### Mistral Code

| Property | Value |
|----------|-------|
| **Type** | `mistral-code` |
| **Config File** | `mistral.json` |
| **Config Directory** | `~/.config/mistral/` |
| **Tier** | 3 |
| **Description** | Mistral Code - Mistral AI coding assistant |

---

### Ollama Code

| Property | Value |
|----------|-------|
| **Type** | `ollama-code` |
| **Config File** | `ollama.json` |
| **Config Directory** | `~/.config/ollama/` |
| **Tier** | 3 |
| **Status** | **DEPRECATED** |
| **Description** | Ollama Code - Local LLM coding assistant |

> **Note:** Ollama is deprecated in HelixAgent (score: 5.0). Use only as fallback.

---

### Plandex

| Property | Value |
|----------|-------|
| **Type** | `plandex` |
| **Config File** | `plandex.json` |
| **Config Directory** | `~/.plandex/` |
| **Tier** | 3 |
| **Description** | Plandex - AI-powered development planning |

---

### Qwen Code

| Property | Value |
|----------|-------|
| **Type** | `qwen-code` |
| **Config File** | `qwen-code.json` |
| **Config Directory** | `~/.config/qwen-code/` |
| **Tier** | 3 |
| **Description** | Qwen Code - Alibaba's AI coding assistant CLI |

---

### Amazon Q

| Property | Value |
|----------|-------|
| **Type** | `amazon-q` |
| **Config File** | `amazon-q.json` |
| **Config Directory** | `~/.aws/amazon-q/` |
| **Tier** | 2 |
| **Description** | Amazon Q - AWS AI coding assistant |

---

## Extended 30 Agents

### Agent-Deck

| Property | Value |
|----------|-------|
| **Type** | `agent-deck` |
| **Config File** | `agent-deck.json` |
| **Config Directory** | `~/.config/agent-deck/` |
| **Tier** | 3 |
| **Description** | Agent-Deck - Multi-agent orchestration platform |

---

### Bridle

| Property | Value |
|----------|-------|
| **Type** | `bridle` |
| **Config File** | `bridle.yaml` |
| **Config Directory** | `~/.config/bridle/` |
| **Tier** | 3 |
| **Description** | Bridle - Constrained AI agent framework |

---

### Cheshire Cat

| Property | Value |
|----------|-------|
| **Type** | `cheshire-cat` |
| **Config File** | `cheshire-cat.json` |
| **Config Directory** | `~/.config/cheshire-cat/` |
| **Tier** | 3 |
| **Description** | Cheshire Cat AI - Customizable AI assistant framework |

---

### Claude Plugins

| Property | Value |
|----------|-------|
| **Type** | `claude-plugins` |
| **Config File** | `plugins.json` |
| **Config Directory** | `~/.claude/plugins/` |
| **Tier** | 3 |
| **Description** | Claude Code Plugins - Extensions for Claude Code |

---

### Claude Squad

| Property | Value |
|----------|-------|
| **Type** | `claude-squad` |
| **Config File** | `claude-squad.yaml` |
| **Config Directory** | `~/.config/claude-squad/` |
| **Tier** | 3 |
| **Description** | Claude Squad - Multi-agent Claude orchestration |

---

### Codai

| Property | Value |
|----------|-------|
| **Type** | `codai` |
| **Config File** | `codai.json` |
| **Config Directory** | `~/.config/codai/` |
| **Tier** | 3 |
| **Description** | Codai - AI code assistant CLI |

---

### Codex

| Property | Value |
|----------|-------|
| **Type** | `codex` |
| **Config File** | `codex.json` |
| **Config Directory** | `~/.config/codex/` |
| **Tier** | 3 |
| **MCP Support** | Yes |
| **Description** | Codex - OpenAI Codex-powered CLI |

---

### Codex Skills

| Property | Value |
|----------|-------|
| **Type** | `codex-skills` |
| **Config File** | `codex-skills.json` |
| **Config Directory** | `~/.config/codex-skills/` |
| **Tier** | 3 |
| **Description** | Codex Skills - Custom skill definitions for Codex |

---

### Conduit

| Property | Value |
|----------|-------|
| **Type** | `conduit` |
| **Config File** | `conduit.json` |
| **Config Directory** | `~/.config/conduit/` |
| **Tier** | 3 |
| **Description** | Conduit - AI data pipeline assistant |

---

### Continue

| Property | Value |
|----------|-------|
| **Type** | `continue` |
| **Config File** | `config.json` |
| **Config Directory** | `~/.continue/` |
| **Language** | TypeScript |
| **Tier** | 2 |
| **Description** | Continue.dev - Open-source AI code assistant |

**Key Features:**
- VS Code and JetBrains extensions
- Context providers
- Slash commands
- Tab autocomplete

---

### Emdash

| Property | Value |
|----------|-------|
| **Type** | `emdash` |
| **Config File** | `emdash.json` |
| **Config Directory** | `~/.config/emdash/` |
| **Tier** | 3 |
| **Description** | Emdash - AI-powered text editing CLI |

---

### FauxPilot

| Property | Value |
|----------|-------|
| **Type** | `fauxpilot` |
| **Config File** | `fauxpilot.yaml` |
| **Config Directory** | `~/.config/fauxpilot/` |
| **Tier** | 3 |
| **Description** | FauxPilot - Self-hosted Copilot alternative |

---

### Get Shit Done

| Property | Value |
|----------|-------|
| **Type** | `get-shit-done` |
| **Config File** | `gsd.json` |
| **Config Directory** | `~/.config/gsd/` |
| **Tier** | 3 |
| **Description** | Get Shit Done - Task-focused AI assistant |

---

### GitHub Copilot CLI

| Property | Value |
|----------|-------|
| **Type** | `github-copilot-cli` |
| **Config File** | `copilot-cli.json` |
| **Config Directory** | `~/.config/github-copilot-cli/` |
| **Tier** | 3 |
| **Description** | GitHub Copilot CLI - Terminal command suggestions |

---

### GitHub Spec Kit

| Property | Value |
|----------|-------|
| **Type** | `github-spec-kit` |
| **Config File** | `spec-kit.json` |
| **Config Directory** | `~/.config/github-spec-kit/` |
| **Tier** | 3 |
| **Description** | GitHub Spec Kit - AI specification generator |

---

### GitMCP

| Property | Value |
|----------|-------|
| **Type** | `git-mcp` |
| **Config File** | `gitmcp.json` |
| **Config Directory** | `~/.config/gitmcp/` |
| **Tier** | 3 |
| **MCP Support** | Yes (Native) |
| **Description** | GitMCP - Git-based MCP server management |

---

### GPTME

| Property | Value |
|----------|-------|
| **Type** | `gptme` |
| **Config File** | `gptme.toml` |
| **Config Directory** | `~/.config/gptme/` |
| **Language** | Python |
| **Tier** | 3 |
| **Description** | GPTME - Personal AI assistant in terminal |

---

### Mobile Agent

| Property | Value |
|----------|-------|
| **Type** | `mobile-agent` |
| **Config File** | `mobile-agent.json` |
| **Config Directory** | `~/.config/mobile-agent/` |
| **Tier** | 3 |
| **Description** | Mobile Agent - AI mobile device automation |

---

### Multiagent Coding

| Property | Value |
|----------|-------|
| **Type** | `multiagent-coding` |
| **Config File** | `multiagent.yaml` |
| **Config Directory** | `~/.config/multiagent-coding/` |
| **Tier** | 3 |
| **Description** | Multiagent Coding - Coordinated multi-agent development |

---

### Nanocoder

| Property | Value |
|----------|-------|
| **Type** | `nanocoder` |
| **Config File** | `nanocoder.json` |
| **Config Directory** | `~/.config/nanocoder/` |
| **Tier** | 3 |
| **Description** | Nanocoder - Lightweight AI code generator |

---

### Noi

| Property | Value |
|----------|-------|
| **Type** | `noi` |
| **Config File** | `noi.json` |
| **Config Directory** | `~/.config/noi/` |
| **Tier** | 3 |
| **Description** | Noi - Cross-platform AI chat interface |

---

### Octogen

| Property | Value |
|----------|-------|
| **Type** | `octogen` |
| **Config File** | `octogen.yaml` |
| **Config Directory** | `~/.config/octogen/` |
| **Tier** | 3 |
| **Description** | Octogen - AI code interpreter and executor |

---

### OpenHands

| Property | Value |
|----------|-------|
| **Type** | `openhands` |
| **Config File** | `openhands.toml` |
| **Config Directory** | `~/.config/openhands/` |
| **Language** | Python |
| **Tier** | 3 |
| **MCP Support** | Yes |
| **Description** | OpenHands - Open-source AI software engineer |

**Key Features:**
- Docker runtime
- Full workspace access
- Tool use support

---

### PostgresMCP

| Property | Value |
|----------|-------|
| **Type** | `postgres-mcp` |
| **Config File** | `postgres-mcp.json` |
| **Config Directory** | `~/.config/postgres-mcp/` |
| **Tier** | 3 |
| **MCP Support** | Yes (Native) |
| **Description** | PostgresMCP - MCP server for PostgreSQL |

---

### Shai

| Property | Value |
|----------|-------|
| **Type** | `shai` |
| **Config File** | `shai.json` |
| **Config Directory** | `~/.config/shai/` |
| **Tier** | 3 |
| **Description** | Shai - Shell AI assistant |

---

### SnowCLI

| Property | Value |
|----------|-------|
| **Type** | `snow-cli` |
| **Config File** | `snowcli.yaml` |
| **Config Directory** | `~/.config/snowcli/` |
| **Tier** | 3 |
| **Description** | SnowCLI - Snowflake AI-assisted CLI |

---

### TaskWeaver

| Property | Value |
|----------|-------|
| **Type** | `task-weaver` |
| **Config File** | `taskweaver.yaml` |
| **Config Directory** | `~/.config/taskweaver/` |
| **Language** | Python |
| **Tier** | 3 |
| **Description** | TaskWeaver - Microsoft's code-first AI agent |

**Key Features:**
- Code interpreter
- Plugin system
- Workspace management

---

### UI/UX Pro Max

| Property | Value |
|----------|-------|
| **Type** | `ui-ux-pro-max` |
| **Config File** | `uiux-pro-max.json` |
| **Config Directory** | `~/.config/uiux-pro-max/` |
| **Tier** | 3 |
| **Description** | UI/UX Pro Max - AI UI/UX design assistant |

---

### VTCode

| Property | Value |
|----------|-------|
| **Type** | `vtcode` |
| **Config File** | `vtcode.json` |
| **Config Directory** | `~/.config/vtcode/` |
| **Tier** | 3 |
| **Description** | VTCode - Visual Terminal Code AI assistant |

---

### Warp

| Property | Value |
|----------|-------|
| **Type** | `warp` |
| **Config File** | `warp.yaml` |
| **Config Directory** | `~/.config/warp/` |
| **Tier** | 3 |
| **Description** | Warp - AI-powered terminal |

**Key Features:**
- Built-in AI assistance
- Workflows
- Block-based terminal
