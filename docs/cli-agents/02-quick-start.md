# Quick Start Guide

Get HelixAgent working with your CLI agent in 5 minutes.

## Prerequisites

- Go 1.24+ installed
- Docker (for HelixAgent dependencies)
- Your preferred CLI agent installed

## Step 1: Build HelixAgent

```bash
cd /path/to/HelixAgent
make build
```

## Step 2: Start HelixAgent

```bash
# Start with default settings (auto-starts Docker dependencies)
./bin/helixagent

# Or start without Docker auto-start
./bin/helixagent --auto-start-docker=false
```

HelixAgent will start on `http://localhost:7061` by default.

## Step 3: Generate Agent Configuration

### Option A: Generate for Specific Agent

```bash
# List all available agents
./bin/helixagent --list-agents

# Generate configuration (outputs to stdout)
./bin/helixagent --generate-agent-config=opencode

# Generate and save to file
./bin/helixagent --generate-agent-config=opencode --agent-config-output=opencode.json
```

### Option B: Generate All Configurations

```bash
./bin/helixagent --generate-all-agents --all-agents-output-dir=~/helix-configs/
```

## Step 4: Install Configuration

Copy the generated configuration to your agent's config directory:

### OpenCode

```bash
cp opencode.json ~/.config/opencode/opencode.json
```

### Claude Code

```bash
mkdir -p ~/.claude
cp claude-code-settings.json ~/.claude/settings.json
```

### Cline

```bash
mkdir -p ~/.config/cline
cp cline.json ~/.config/cline/cline.json
```

### Aider

```bash
cp .aider.conf.yml ~/.aider.conf.yml
```

### Continue

```bash
mkdir -p ~/.continue
cp config.json ~/.continue/config.json
```

### Codex

```bash
mkdir -p ~/.config/codex
cp codex.json ~/.config/codex/codex.json
```

## Step 5: Verify Configuration

```bash
./bin/helixagent --validate-agent-config=opencode:~/.config/opencode/opencode.json
```

Expected output:
```
✓ Config file is valid for opencode
```

## Step 6: Use Your Agent

Now use your CLI agent normally. All requests will route through HelixAgent's AI Debate Ensemble!

```bash
# OpenCode
opencode

# Claude Code
claude

# Aider
aider

# Cline
cline
```

## Quick Examples

### Example 1: OpenCode with HelixAgent

```bash
# Generate config
./bin/helixagent --generate-agent-config=opencode --agent-config-output=/tmp/opencode.json

# View config
cat /tmp/opencode.json

# Install
cp /tmp/opencode.json ~/.config/opencode/opencode.json

# Run OpenCode
opencode
```

### Example 2: All Agents Batch Setup

```bash
# Generate all 48 configs
./bin/helixagent --generate-all-agents --all-agents-output-dir=~/helix-configs/

# List generated files
ls -la ~/helix-configs/

# Output:
# opencode.json
# crush.json
# helixcode.json
# aider.conf.yml
# cline.json
# codex.json
# ... (48 files total)
```

### Example 3: Custom HelixAgent Host

If HelixAgent is running on a different host:

```bash
# Edit the generated config to change the base URL
# Change: "base_url": "http://localhost:7061/v1"
# To:     "base_url": "http://myserver:7061/v1"
```

## Configuration Structure

All generated configurations follow this pattern:

```json
{
  "version": "1.0",
  "provider": {
    "type": "openai-compatible",
    "name": "helixagent",
    "base_url": "http://localhost:7061/v1",
    "api_key_env": "HELIXAGENT_API_KEY"
  },
  "models": [
    {
      "id": "helixagent-debate",
      "name": "HelixAgent AI Debate Ensemble",
      "max_tokens": 128000,
      "capabilities": ["vision", "streaming", "function_calls", "embeddings", "mcp", "acp", "lsp"]
    }
  ],
  "mcp": {
    "helixagent-mcp": {"type": "remote", "url": "http://localhost:7061/v1/mcp"},
    "helixagent-acp": {"type": "remote", "url": "http://localhost:7061/v1/acp"},
    "filesystem": {"type": "local", "command": ["npx", "-y", "@modelcontextprotocol/server-filesystem"]},
    "github": {"type": "local", "command": ["npx", "-y", "@modelcontextprotocol/server-github"]}
  },
  "settings": {
    // Agent-specific settings
  }
}
```

## Environment Variables

Set these environment variables for authentication:

```bash
# HelixAgent API key (optional, generated if not set)
export HELIXAGENT_API_KEY="your-api-key"

# For specific providers (if using directly)
export CLAUDE_API_KEY="..."
export OPENAI_API_KEY="..."
export DEEPSEEK_API_KEY="..."
```

## Troubleshooting

### Agent can't connect to HelixAgent

1. Verify HelixAgent is running: `curl http://localhost:7061/health`
2. Check the config has correct `base_url`
3. Ensure no firewall blocking port 7061

### Configuration validation fails

```bash
# Get detailed validation errors
./bin/helixagent --validate-agent-config=opencode:config.json
```

### MCP servers not working

```bash
# Pre-install MCP packages
./bin/helixagent --preinstall-mcp
```

## Next Steps

- [Agent Reference](./03-agent-reference.md) - Details on all 48 agents
- [Configuration Guide](./04-configuration-guide.md) - Advanced configuration options
- [Plugin Architecture](./05-plugin-architecture.md) - For Tier 1 plugin development
