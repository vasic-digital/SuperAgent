# CLI Agents Documentation

HelixAgent supports **48 CLI agents** with unified configuration generation, validation, and plugin integration.

## Table of Contents

1. [Overview](./01-overview.md) - Introduction to CLI agents support
2. [Quick Start](./02-quick-start.md) - Get started in 5 minutes
3. [Agent Reference](./03-agent-reference.md) - Complete list of 48 agents
4. [Configuration Guide](./04-configuration-guide.md) - Configuration formats and options
5. [Plugin Architecture](./05-plugin-architecture.md) - How plugins work
6. [Tier 1 Plugins](./06-tier1-plugins.md) - Claude Code, OpenCode, Cline, Kilo-Code
7. [Tier 2-3 Integration](./07-tier2-tier3-integration.md) - Generic MCP server approach
8. [Transport Layer](./08-transport-layer.md) - HTTP/3, TOON, Brotli
9. [Event System](./09-event-system.md) - SSE, WebSocket, Webhooks
10. [UI Extensions](./10-ui-extensions.md) - Debate visualization, progress bars
11. [Development Guide](./11-development-guide.md) - Creating new agent support
12. [Troubleshooting](./12-troubleshooting.md) - Common issues and solutions
13. [API Reference](./13-api-reference.md) - CLI flags and programmatic API

## Quick Links

### Generate Configurations

```bash
# List all 48 agents
./bin/helixagent --list-agents

# Generate for specific agent
./bin/helixagent --generate-agent-config=codex --agent-config-output=codex.json

# Generate all 48 at once
./bin/helixagent --generate-all-agents --all-agents-output-dir=~/agent-configs/
```

### Validate Configurations

```bash
./bin/helixagent --validate-agent-config=opencode:~/.config/opencode/opencode.json
```

### Run Challenge Tests

```bash
./challenges/scripts/all_agents_e2e_challenge.sh  # 102 tests
```

## Agent Categories

| Category | Count | Examples |
|----------|-------|----------|
| Original CLI Agents | 18 | OpenCode, Claude Code, Aider, Cline |
| Extended CLI Agents | 30 | Codex, OpenHands, TaskWeaver, Continue |
| **Total** | **48** | |

## Support Matrix

| Feature | Tier 1 | Tier 2 | Tier 3 |
|---------|--------|--------|--------|
| Full Plugin Support | Yes | Partial | Config Only |
| HTTP/3 Transport | Yes | Yes | Via MCP |
| TOON Protocol | Yes | Yes | Via MCP |
| Event Streaming | Yes | Yes | Limited |
| UI Extensions | Yes | Limited | No |
| Agents | 4 | 8 | 36 |
