# CLI Agents Overview

## What are CLI Agents?

CLI agents are command-line AI coding assistants that help developers write, review, and debug code. HelixAgent provides a unified backend for 48 different CLI agents, allowing them all to leverage the AI Debate Ensemble for superior code assistance.

## The HelixAgent Advantage

Instead of each CLI agent connecting to a single LLM provider, HelixAgent routes requests through its **AI Debate Ensemble** system:

```
┌─────────────────────────────────────────────────────────────┐
│                      CLI Agents                              │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐           │
│  │OpenCode │ │ Claude  │ │  Cline  │ │  Codex  │  ... x48  │
│  │         │ │  Code   │ │         │ │         │           │
│  └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘           │
│       │           │           │           │                 │
│       └───────────┴───────────┴───────────┘                 │
│                       │                                      │
│              HelixAgent Gateway                              │
│       ┌───────────────┴───────────────┐                     │
│       │    AI Debate Ensemble         │                     │
│       │  ┌─────┐ ┌─────┐ ┌─────┐     │                     │
│       │  │Claude│ │Gemini│ │DeepSk│    │                     │
│       │  └─────┘ └─────┘ └─────┘     │                     │
│       │  25 LLMs debating for best   │                     │
│       │  consensus response          │                     │
│       └───────────────────────────────┘                     │
└─────────────────────────────────────────────────────────────┘
```

## Key Features

### 1. Unified Configuration Generation

Generate configurations for any of the 48 supported CLI agents:

```bash
./bin/helixagent --generate-agent-config=opencode
./bin/helixagent --generate-agent-config=claude-code
./bin/helixagent --generate-agent-config=codex
```

### 2. Configuration Validation

Validate existing configurations against LLMsVerifier schemas:

```bash
./bin/helixagent --validate-agent-config=opencode:~/.config/opencode/opencode.json
```

### 3. Batch Generation

Generate all 48 configurations at once:

```bash
./bin/helixagent --generate-all-agents --all-agents-output-dir=~/agent-configs/
```

### 4. Plugin System

Tier 1 agents (Claude Code, OpenCode, Cline, Kilo-Code) have full plugin support with:
- HTTP/3 + QUIC transport
- TOON protocol (40-70% token savings)
- Brotli compression
- Real-time event streaming
- AI Debate visualization

### 5. MCP Integration

All agents can connect via MCP (Model Context Protocol):
- 6 HelixAgent MCP endpoints (mcp, acp, lsp, embeddings, vision, cognee)
- 6 standard MCP servers (filesystem, github, memory, fetch, puppeteer, sqlite)

## Agent Tiers

### Tier 1: Full Plugin Support (4 agents)

These agents have rich extension mechanisms allowing deep HelixAgent integration:

| Agent | Language | Extension Mechanism |
|-------|----------|---------------------|
| Claude Code | TypeScript | Plugin marketplace, 2 hooks |
| OpenCode | Go | MCP servers (stdio/SSE), LSP |
| Cline | TypeScript | 8 lifecycle hooks, MCP manager |
| Kilo-Code | TypeScript | MCP, services, monorepo |

### Tier 2: Moderate Extensibility (8 agents)

These agents support configuration-based customization and MCP:

- Aider (Python - coder classes, litellm)
- Codename Goose (TypeScript/Rust - MCP transport)
- Forge (TypeScript - templates)
- Amazon Q, Kiro, GPT Engineer, Gemini CLI, DeepSeek CLI

### Tier 3: Configuration Only (36 agents)

These agents use the generic MCP server approach:

- All 30 new agents (Codex, OpenHands, TaskWeaver, etc.)
- Remaining original agents (Mistral Code, Ollama Code, Plandex, etc.)

## Architecture Overview

```
HelixAgent CLI Agent Support
├── Configuration Generation (LLMsVerifier)
│   ├── pkg/cliagents/generator.go      # Unified generator
│   ├── pkg/cliagents/additional_agents.go  # 44 generic generators
│   ├── pkg/cliagents/opencode.go       # OpenCode custom generator
│   ├── pkg/cliagents/crush.go          # Crush custom generator
│   ├── pkg/cliagents/kilocode.go       # KiloCode custom generator
│   └── pkg/cliagents/helixcode.go      # HelixCode custom generator
│
├── Agent Registry (HelixAgent)
│   └── internal/agents/registry.go     # 48 agent definitions
│
├── CLI Interface
│   └── cmd/helixagent/main.go          # CLI flags and handlers
│
├── Plugin Libraries
│   ├── packages/transport/             # HTTP/3 + TOON + Brotli
│   ├── packages/events/                # SSE, WebSocket clients
│   └── packages/ui/                    # Debate renderer, progress bars
│
└── Plugins
    ├── agents/claude_code/             # Claude Code plugin
    ├── agents/opencode/                # OpenCode MCP server
    ├── agents/cline/                   # Cline extension
    ├── agents/kilo_code/               # Kilo-Code packages
    └── mcp-server/                     # Generic MCP for Tier 2-3
```

## Getting Started

1. **Build HelixAgent**:
   ```bash
   make build
   ```

2. **List available agents**:
   ```bash
   ./bin/helixagent --list-agents
   ```

3. **Generate configuration for your agent**:
   ```bash
   ./bin/helixagent --generate-agent-config=<your-agent>
   ```

4. **Install the configuration**:
   Copy the generated config to your agent's config directory.

5. **Start HelixAgent**:
   ```bash
   ./bin/helixagent
   ```

6. **Use your CLI agent**:
   Your agent now routes through HelixAgent's AI Debate Ensemble!

## Next Steps

- [Quick Start Guide](./02-quick-start.md) - Detailed setup instructions
- [Agent Reference](./03-agent-reference.md) - All 48 agents with details
- [Plugin Architecture](./05-plugin-architecture.md) - How plugins work
