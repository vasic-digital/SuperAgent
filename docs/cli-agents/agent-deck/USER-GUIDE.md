# Agent Deck - User Guide

**Agent Deck** is a terminal session manager for AI coding agents. It provides a unified TUI (Terminal User Interface) to manage multiple AI agent sessions including Claude, Gemini, OpenCode, Codex, and more from a single terminal window.

---

## Table of Contents

1. [Installation](#installation)
2. [Quick Start](#quick-start)
3. [CLI Commands](#cli-commands)
4. [TUI/Interactive Commands](#tuiinteractive-commands)
5. [Configuration](#configuration)
6. [Usage Examples](#usage-examples)
7. [Troubleshooting](#troubleshooting)

---

## Installation

### Prerequisites

- macOS, Linux, or Windows (WSL)
- tmux (for session management)
- Git

### Method 1: Official Installer (Recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/asheshgoplani/agent-deck/main/install.sh | bash
```

Then run: `agent-deck`

### Method 2: Homebrew

```bash
brew install asheshgoplani/tap/agent-deck
```

### Method 3: Go Install

```bash
go install github.com/asheshgoplani/agent-deck/cmd/agent-deck@latest
```

### Method 4: From Source

```bash
git clone https://github.com/asheshgoplani/agent-deck.git
cd agent-deck
make install
```

### Verify Installation

```bash
agent-deck --version
```

---

## Quick Start

### Launch Agent Deck

```bash
# Start the TUI
agent-deck
```

### Add Your First Session

```bash
# Add current directory with Claude
agent-deck add . -c claude

# Add with a specific agent
agent-deck add /path/to/project -c claude
agent-deck add /path/to/project -c codex
agent-deck add /path/to/project -c opencode
```

### Fork a Session

```bash
# Fork an existing session
agent-deck session fork my-proj
```

### Attach MCP Server

```bash
# Attach an MCP server to a session
agent-deck mcp attach my-proj exa
```

### Attach Skills

```bash
# Attach a skill and restart session
agent-deck skill attach my-proj docs --source pool --restart
```

### Start Web UI

```bash
# Start web interface
agent-deck web

# Read-only mode
agent-deck web --read-only

# Custom port
agent-deck web --listen 127.0.0.1:9000

# With authentication
agent-deck web --token my-secret
```

---

## CLI Commands

### Core Commands

| Command | Description |
|---------|-------------|
| `agent-deck` | Launch the TUI |
| `agent-deck --version` | Show version information |
| `agent-deck update` | Update to latest version |
| `agent-deck uninstall` | Uninstall agent-deck |
| `agent-deck uninstall --keep-data` | Uninstall but keep sessions |

### Session Management

| Command | Description |
|---------|-------------|
| `agent-deck add <path> -c <agent>` | Add a new session |
| `agent-deck session fork <name>` | Fork an existing session |
| `agent-deck session restart <name>` | Restart a session |
| `agent-deck session delete <name>` | Delete a session |
| `agent-deck session list` | List all sessions |

### MCP Management

| Command | Description |
|---------|-------------|
| `agent-deck mcp attach <session> <server>` | Attach MCP server |
| `agent-deck mcp detach <session> <server>` | Detach MCP server |
| `agent-deck mcp list` | List available MCP servers |
| `agent-deck mcp pool` | Manage MCP socket pooling |

### Skill Management

| Command | Description |
|---------|-------------|
| `agent-deck skill attach <session> <skill>` | Attach a skill |
| `agent-deck skill detach <session> <skill>` | Detach a skill |
| `agent-deck skill list` | List available skills |

### Conductor Management

Conductors are persistent agent sessions that monitor and orchestrate other sessions:

```bash
# Setup a conductor
agent-deck -p work conductor setup ops --description "Ops monitor"

# Setup with specific agent
agent-deck conductor setup review --agent codex --description "Codex reviewer"

# Setup with custom environment
agent-deck conductor setup glm-bot \
  -env ANTHROPIC_BASE_URL=https://api.z.ai/api/anthropic \
  -env ANTHROPIC_AUTH_TOKEN=<token> \
  -env ANTHROPIC_DEFAULT_OPUS_MODEL=glm-5

# List conductors
agent-deck conductor list
agent-deck conductor list --profile work

# Check status
agent-deck conductor status
agent-deck conductor status ops

# Teardown
agent-deck conductor teardown ops
agent-deck conductor teardown --all --remove
```

### Web Interface

| Command | Description |
|---------|-------------|
| `agent-deck web` | Start web UI (default: 127.0.0.1:8420) |
| `agent-deck web --read-only` | Read-only mode |
| `agent-deck web --listen <host:port>` | Custom listen address |
| `agent-deck web --token <secret>` | Enable token authentication |

---

## TUI/Interactive Commands

### Keyboard Shortcuts

#### Navigation

| Key | Action |
|-----|--------|
| `↑/↓` or `j/k` | Navigate between sessions |
| `Enter` | Attach to selected session |
| `Tab` | Switch between preview and diff tabs |
| `/` or `G` | Search / Global search |

#### Session Management

| Key | Action |
|-----|--------|
| `n` | New session |
| `N` | New session with prompt |
| `f` | Fork session (quick) |
| `F` | Fork session (dialog) |
| `r` | Restart session |
| `d` | Delete session |
| `D` | Kill session |

#### Actions

| Key | Action |
|-----|--------|
| `↵` or `o` | Attach to session to reprompt |
| `Ctrl+Q` | Detach from session |
| `s` | Commit and push branch to GitHub |
| `c` | Checkout (commit changes and pause) |
| `r` | Resume paused session |
| `M` | Move session to group |

#### Tools & Settings

| Key | Action |
|-----|--------|
| `m` | MCP Manager |
| `$` | Cost Dashboard |
| `S` | Settings |
| `T` | Container shell (sandboxed sessions) |
| `?` | Show full help |
| `q` | Quit application |

### Status Indicators

- **Running**: Agent is actively working
- **Waiting**: Agent waiting for user input
- **Idle**: Agent is idle/ready
- **Error**: Session encountered an error

---

## Configuration

### Configuration File

Agent Deck stores configuration in `~/.agent-deck/config.toml`:

```toml
# Profile settings
[claude]
config_dir = "~/.claude"  # Global default
allow_dangerous_mode = false

# Per-profile overrides
[profiles.work.claude]
config_dir = "~/.claude-work"

[profiles.personal.claude]
config_dir = "~/.claude-personal"

# Auto-update
auto_update = true

# Default settings
default_agent = "claude"
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `AGENT_DECK_PROFILE` | Default profile to use |
| `AGENT_DECK_CONFIG` | Path to config file |
| `AGENT_DECK_DATA` | Path to data directory |

### Profile Configuration

Create profiles for different contexts:

```bash
# Use a specific profile
agent-deck -p work

# Profile-specific Claude config
mkdir -p ~/.claude-work
# Copy or create settings.json
```

### Conductor Configuration

Conductors store configuration in `~/.agent-deck/conductor/`:

```
~/.agent-deck/conductor/
├── CLAUDE.md           # Shared knowledge for Claude conductors
├── AGENTS.md           # Shared knowledge for Codex conductors
├── bridge.py           # Bridge daemon (Telegram/Slack)
├── ops/                # Conductor instance
│   ├── CLAUDE.md       # Identity definition
│   ├── meta.json       # Config: name, profile, description
│   ├── state.json      # Runtime state
│   └── task-log.md     # Action log
└── review/
    └── ...
```

### Telegram Bridge Setup

Configure Telegram for remote monitoring:

1. Create a Telegram bot via @BotFather
2. Get your bot token
3. Run `agent-deck conductor setup <name>` and enter token
4. Messages route using `name: message` prefix

### Slack Bridge Setup

Configure Slack for channel-based monitoring:

1. Create a Slack app at api.slack.com/apps
2. Enable Socket Mode and generate app-level token (xapp-...)
3. Add bot scopes: `chat:write`, `channels:history`, `channels:read`
4. Subscribe to bot events: `message.channels`, `app_mention`
5. Install to workspace
6. Run `agent-deck conductor setup <name>` and enter tokens

---

## Usage Examples

### Basic Workflow

```bash
# 1. Navigate to project
cd ~/projects/my-app

# 2. Launch Agent Deck
agent-deck

# 3. Create new session (press 'n' in TUI)
# or from CLI:
agent-deck add . -c claude

# 4. Attach to session (press Enter)

# 5. Work with the agent, then detach (Ctrl+Q)
```

### Managing Multiple Projects

```bash
# Add multiple projects
agent-deck add ~/projects/frontend -c claude --name frontend
agent-deck add ~/projects/backend -c codex --name backend
agent-deck add ~/projects/docs -c opencode --name docs

# Switch between them in TUI with arrow keys
```

### Using MCP Servers

```bash
# Attach Exa search to a session
agent-deck mcp attach my-proj exa

# Attach with restart
agent-deck mcp attach my-proj exa --restart

# List attached MCPs
agent-deck mcp list my-proj
```

### Working with Skills

```bash
# Install skill from marketplace
agent-deck skill install asheshgoplani/agent-deck

# Attach skill to session
agent-deck skill attach my-proj docs --source pool --restart

# List available skills
agent-deck skill list
```

### Forking Sessions

```bash
# Fork for experimental changes
agent-deck session fork my-proj --name my-proj-experiment

# Work on experiment, then merge or discard
```

### Using Conductors

```bash
# Setup monitoring conductor
agent-deck -p work conductor setup ops --description "Monitor all work sessions"

# Conductor will watch sessions and auto-respond when confident
# Escalates to you when uncertain
```

### CI/CD Integration

```bash
# Non-interactive session creation
agent-deck add . -c claude --non-interactive

# Script agent interactions
agent-deck exec my-proj "run tests and fix failures"
```

---

## Troubleshooting

### Installation Issues

#### "command not found" after installation

```bash
# Add to PATH
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

#### Permission denied on macOS

```bash
# Remove quarantine attribute
xattr -d com.apple.quarantine ~/.local/bin/agent-deck
```

### Session Issues

#### Failed to start new session

If you get `failed to start new session: timed out waiting for tmux session`:

1. Update the underlying agent (claude, codex, etc.) to latest version
2. Check tmux is installed: `which tmux`
3. Verify tmux version: `tmux -V` (needs 3.0+)

#### Session not responding

```bash
# Force restart
agent-deck session restart <name> --force

# Or kill and recreate
agent-deck session kill <name>
agent-deck add . -c claude
```

### TUI Issues

#### Display garbled or colors wrong

```bash
# Set TERM explicitly
export TERM=xterm-256color
agent-deck
```

#### Keys not working

Check terminal emulator compatibility:
- iTerm2: Fully supported
- Windows Terminal: Supported (WSL)
- Alacritty: Supported
- VS Code Terminal: Supported

### MCP Issues

#### MCP server fails to attach

```bash
# Check MCP config
agent-deck mcp config validate

# Restart MCP daemon
agent-deck mcp daemon restart
```

#### Socket pooling issues

```bash
# Reset MCP pool
agent-deck mcp pool reset

# Reconfigure pool
agent-deck mcp pool setup
```

### Conductor Issues

#### Conductor not responding

```bash
# Check conductor status
agent-deck conductor status <name>

# View logs
tail -f ~/.agent-deck/conductor/<name>/task-log.md

# Restart conductor
agent-deck conductor teardown <name>
agent-deck conductor setup <name>
```

#### Telegram bridge not working

1. Verify bot token is correct
2. Check bot is not in privacy mode
3. Ensure you've started a conversation with the bot

### Configuration Issues

#### Profile not loading

```bash
# Verify config
agent-deck config validate

# Check profile exists
agent-deck profile list

# Debug hooks
agent-deck hooks status
agent-deck hooks status -p work
```

### Update Issues

#### Auto-update failing

```bash
# Manual update
agent-deck update

# Or via Homebrew
brew upgrade asheshgoplani/tap/agent-deck
```

### Getting Help

```bash
# Show help
agent-deck --help

# Debug information
agent-deck debug

# Check version
agent-deck --version
```

### Community Resources

- GitHub Issues: https://github.com/asheshgoplani/agent-deck/issues
- Documentation: https://github.com/asheshgoplani/agent-deck/tree/main/docs
- Discord: Join via GitHub README

---

## Best Practices

1. **Use Profiles**: Separate work and personal configurations
2. **Name Sessions Clearly**: Use descriptive names for easy identification
3. **Regular Updates**: Keep agent-deck and agents updated
4. **MCP Pooling**: Use MCP socket pooling for better performance
5. **Conductor Monitoring**: Set up conductors for long-running tasks
6. **Git Integration**: Use git worktrees for isolated session workspaces
7. **Cost Tracking**: Monitor token usage with the cost dashboard (`$` key)

---

*Last Updated: April 2026*
