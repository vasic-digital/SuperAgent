# HelixAgent CLI Agent Plugins

This directory contains plugins and integration tools for 48 CLI agents to connect with HelixAgent's AI Debate Ensemble system.

## Overview

HelixAgent provides plugins for CLI agents with:
- **HTTP/3 (QUIC)** transport with automatic fallback to HTTP/2 and HTTP/1.1
- **TOON Protocol** encoding for 40-70% token savings
- **Brotli Compression** for additional bandwidth reduction
- **Real-time Events** via SSE, WebSocket, or Webhooks
- **Rich UI/UX** with debate visualization and progress bars

## Directory Structure

```
plugins/
├── packages/                    # Shared libraries
│   ├── transport/               # HTTP/3 + TOON + Brotli
│   │   ├── go/                  # Go implementation
│   │   └── typescript/          # TypeScript implementation
│   ├── events/                  # Event subscription
│   │   └── event_client.ts      # SSE/WebSocket clients
│   └── ui/                      # UI/UX rendering
│       └── debate_renderer.ts   # Debate visualization
├── agents/                      # Per-agent plugins
│   ├── claude_code/             # Tier 1 - Full plugin
│   ├── opencode/                # Tier 1 - MCP server (Go)
│   ├── cline/                   # Tier 1 - 8 hooks
│   └── kilo_code/               # Tier 1 - NPM package
├── mcp-server/                  # Generic MCP server (Tier 2-3)
├── schemas/                     # Configuration schemas
├── configs/                     # Generated agent configs
└── tools/                       # Utility scripts
```

## Agent Tiers

### Tier 1 - Full Plugin Support (4 agents)

| Agent | Language | Integration | Directory |
|-------|----------|-------------|-----------|
| **Claude Code** | TypeScript | 4 hooks (SessionStart, SessionEnd, PreToolUse, PostToolUse) | `agents/claude_code/` |
| **OpenCode** | Go | Native MCP server | `agents/opencode/` |
| **Cline** | TypeScript | 8 hooks (full lifecycle) | `agents/cline/` |
| **Kilo-Code** | TypeScript | NPM package (multi-platform) | `agents/kilo_code/` |

### Tier 2 - MCP + Config (8 agents)

Aider, Codename Goose, Forge, Amazon Q, Kiro, GPT Engineer, Gemini CLI, DeepSeek CLI

### Tier 3 - Generic MCP Server (36 agents)

All other CLI agents use the generic MCP server with agent-specific configurations.

## Quick Start

### Using Tier 1 Plugins

**Claude Code:**
```bash
# Copy plugin to Claude Code plugins directory
cp -r plugins/agents/claude_code ~/.claude/plugins/helixagent/
```

**OpenCode:**
```bash
# Add to .opencode.json
cat plugins/agents/opencode/opencode.json >> ~/.opencode.json
```

**Cline:**
```bash
# Copy hooks to Cline rules directory
cp -r plugins/agents/cline/hooks ~/.clinerules/hooks/helixagent/
```

**Kilo-Code:**
```bash
# Install as npm package
cd plugins/agents/kilo_code
npm install && npm run build
npm link
```

### Using Generic MCP Server (Tier 2-3)

```bash
# Install the generic MCP server
cd plugins/mcp-server
npm install && npm run build
npm link

# Run with any MCP-compatible agent
helixagent-mcp --endpoint https://localhost:7061

# Or generate config for a specific agent
./tools/generate_agent_config.sh
```

## Configuration

All plugins use a universal configuration schema (`schemas/helixagent-plugin-schema.json`):

```json
{
  "endpoint": "https://localhost:7061",
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
  }
}
```

## Tools Available

All plugins provide these HelixAgent tools:

| Tool | Description |
|------|-------------|
| `helixagent_debate` | Start AI debate with 15 LLMs (5 positions × 3 each) |
| `helixagent_ensemble` | Single query to AI Debate Ensemble |
| `helixagent_task` | Create background tasks |
| `helixagent_rag` | Hybrid RAG query (dense + sparse) |
| `helixagent_memory` | Mem0-style memory system |
| `helixagent_providers` | Get provider information |

## Render Styles

Debate output supports 5 render styles:

- **theater** - Dramatic presentation with character entrances
- **novel** - Narrative prose style
- **screenplay** - Script format with centered names
- **minimal** - Clean markdown format
- **plain** - No formatting

## Progress Bar Styles

- **ascii** - `[====    ]`
- **unicode** - `┃████░░░░┃`
- **block** - Smooth Unicode blocks
- **dots** - `●●●○○○○○`

## Event Types

### Task Events (14 types)
```
task.created, task.started, task.progress, task.heartbeat,
task.paused, task.resumed, task.completed, task.failed,
task.stuck, task.cancelled, task.retrying, task.deadletter,
task.log, task.resource
```

### Debate Events (8 types)
```
debate.started, debate.round_started, debate.position_submitted,
debate.validation_phase, debate.polish_phase, debate.consensus,
debate.completed, debate.failed
```

## Testing

Run the plugin challenge scripts to verify your setup:

```bash
# Run all plugin challenges (90 tests)
./challenges/scripts/run_plugin_challenges.sh

# Run individual challenges
./challenges/scripts/plugin_transport_challenge.sh    # 25 tests
./challenges/scripts/plugin_events_challenge.sh       # 20 tests
./challenges/scripts/plugin_ui_challenge.sh           # 15 tests
./challenges/scripts/plugin_integration_challenge.sh  # 30 tests
```

## Development

### Building TypeScript Packages

```bash
# Build all packages
cd plugins/packages/transport/typescript && npx tsc
cd plugins/packages/events && npx tsc
cd plugins/packages/ui && npx tsc
cd plugins/agents/kilo_code && npm run build
cd plugins/mcp-server && npm run build
```

### Building Go Packages

```bash
cd plugins/packages/transport/go
go build ./...

cd plugins/agents/opencode/mcp
go build -o helixagent-mcp-opencode
```

## License

Apache-2.0
