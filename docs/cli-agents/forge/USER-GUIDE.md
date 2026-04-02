# Forge User Guide

## Table of Contents
1. [Installation](#installation)
2. [Quick Start](#quick-start)
3. [CLI Commands](#cli-commands)
4. [TUI/Interactive Commands](#tui-interactive-commands)
5. [Configuration](#configuration)
6. [API/Protocol Endpoints](#api-protocol-endpoints)
7. [Usage Examples](#usage-examples)
8. [Troubleshooting](#troubleshooting)

## Installation

### Method 1: NPM (Recommended)

```bash
# Install globally
npm install -g @forge-agents/forge

# Verify installation
forge --version

# Upgrade
npm install -g @forge-agents/forge@latest
```

### Method 2: npx (No Installation)

```bash
# Run without installing
npx @forge-agents/forge
```

### Method 3: Binary Download

```bash
# macOS
curl -L -o forge "https://github.com/forge-agents/forge/releases/latest/download/forge-darwin-amd64"
chmod +x forge
sudo mv forge /usr/local/bin/

# Linux
curl -L -o forge "https://github.com/forge-agents/forge/releases/latest/download/forge-linux-amd64"
chmod +x forge
sudo mv forge /usr/local/bin/
```

### Method 4: Build from Source

```bash
# Clone repository
git clone https://github.com/forge-agents/forge.git
cd forge

# Install dependencies
npm install

# Build
npm run build

# Link
npm link
```

### Prerequisites

- Node.js 18 or higher
- Zsh (recommended for full features)
- Nerd Font (for icons)
- API key for at least one LLM provider

## Quick Start

### First-Time Setup

```bash
# Verify installation
forge --version

# Run setup wizard
forge setup

# Configure Zsh plugin
# Follow interactive prompts
# Restart terminal after setup

# Log in to AI provider
forge login
```

### Basic Usage

```bash
# Start interactive TUI
forge

# Run with prompt
forge "create a todo app"

# Run specific agent
forge claude "refactor auth module"

# Continue last session
forge --continue
```

### Hello World

```bash
# Start Forge
forge

# Or with prompt
forge "Create a Python script that prints Hello World"

# Using Zsh plugin (after setup)
: create a hello world program in Python
```

## CLI Commands

### Global Options

| Option | Short | Description | Example |
|--------|-------|-------------|---------|
| --version | -v | Show version | `forge --version` |
| --help | -h | Show help | `forge --help` |
| --continue | -c | Continue last session | `forge --continue` |
| --session | -s | Resume specific session | `forge --session abc123` |
| --project | | Set project path | `forge --project /path/to/project` |
| --print-logs | | Print logs to stderr | `forge --print-logs` |
| --log-level | | Set log level | `forge --log-level DEBUG` |
| --porcelain | | Machine-readable output | `forge --porcelain` |

### Command: (default - TUI mode)

**Description:** Start interactive TUI mode.

**Usage:**
```bash
forge [OPTIONS]
```

**Examples:**
```bash
# Start TUI
forge

# Continue session
forge --continue

# With specific project
forge --project ~/my-project
```

**Exit Codes:**
| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 130 | Interrupted |

### Command: agents

**Description:** List all available agents.

**Usage:**
```bash
forge agents [OPTIONS]
```

**Options:**
| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| --porcelain | boolean | No | false | Machine-readable output |

**Examples:**
```bash
# List agents
forge agents

# JSON output
forge agents --porcelain
```

**Available Agents:**
| Agent | Description |
|-------|-------------|
| claude | Claude Code via ACP |
| codex | Codex CLI via ACP |
| gemini | Gemini CLI |
| augment | Augment Code |
| goose | Goose agent |
| openhands | OpenHands |
| opencode | OpenCode |
| kimi | Kimi CLI |
| mistral | Mistral Vibe |
| qwen | Qwen Code |

### Command: <agent>

**Description:** Run a specific agent.

**Usage:**
```bash
forge <agent> [SUBCOMMAND|PROMPT]
```

**Subcommands:**
| Subcommand | Description |
|------------|-------------|
| install | Install the agent |
| uninstall | Uninstall the agent |
| check | Check agent installation |
| models | List available models |
| modes | List available modes |

**Options:**
| Option | Type | Description |
|--------|------|-------------|
| --model | string | Model to use |
| --mode | string | Session mode |
| --print | boolean | Run headless |
| --continue | boolean | Continue session |
| --session | string | Session ID |

**Examples:**
```bash
# Install agent
forge claude install

# Run agent
forge claude "refactor auth module"

# With specific model/mode
forge claude --model opus --mode acceptEdits "create tests"

# List models
forge claude models

# Headless mode
forge claude --print "fix bug"
```

### Command: setup

**Description:** Setup Zsh integration.

**Usage:**
```bash
forge setup
```

**Examples:**
```bash
# Run setup wizard
forge setup

# Follow prompts
# Restart terminal when complete
```

### Command: doctor

**Description:** Run diagnostics.

**Usage:**
```bash
forge doctor
```

**Examples:**
```bash
# Check environment
forge doctor
```

### Command: update

**Description:** Update Forge to latest version.

**Usage:**
```bash
forge update
```

### Command: login

**Description:** Log in to AI provider.

**Usage:**
```bash
forge login [PROVIDER]
```

**Examples:**
```bash
# Interactive login
forge login

# Login to specific provider
forge login openrouter
forge login anthropic
forge login openai
```

### Command: info

**Description:** Show configuration and environment status.

**Usage:**
```bash
forge info [OPTIONS]
```

**Options:**
| Option | Type | Description |
|--------|------|-------------|
| --porcelain | boolean | Machine-readable output |

### Command: list

**Description:** List various resources.

**Usage:**
```bash
forge list <resource> [OPTIONS]
```

**Resources:**
| Resource | Description |
|----------|-------------|
| agent | List installed agents |
| model | List available models |
| conversation | List conversations |

**Examples:**
```bash
# List agents
forge list agent

# List conversations
forge list conversation

# With JSON output
forge list agent --porcelain
```

### Command: conversation

**Description:** Manage conversations.

**Usage:**
```bash
forge conversation <subcommand> [OPTIONS]
```

**Subcommands:**
| Subcommand | Description |
|------------|-------------|
| list | List conversations |
| show | Show conversation details |
| delete | Delete conversation |

**Examples:**
```bash
# List conversations
forge conversation list

# Show specific conversation
forge conversation show <id>

# Delete conversation
forge conversation delete <id>
```

## TUI/Interactive Commands

When running in TUI mode, use these commands:

| Command | Description | Example |
|---------|-------------|---------|
| /help | Show help | `/help` |
| /new | Start new session | `/new` |
| /agent | Switch agent | `/agent claude` |
| /model | Switch model | `/model gpt-4o` |
| /mode | Change mode | `/mode auto` |
| /history | Show history | `/history` |
| /clear | Clear screen | `/clear` |
| /exit | Exit Forge | `/exit` |

### Built-in Agents

| Agent | Purpose | Command |
|-------|---------|---------|
| /muse | Planning and reviewing | `/muse` |
| /forge | Implementation | `/forge` |

**Example:**
```bash
# Plan with muse
/muse
> Plan the architecture for a new feature

# Implement with forge
/forge
> Implement the planned feature
```

### Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| Ctrl+C | Cancel / Exit |
| Ctrl+L | Clear screen |
| Tab | Autocomplete |
| Up/Down | Navigate history |

## Configuration

### Configuration File Format

Forge uses `~/.forge/config.json`:

```json
{
  "defaultAgent": "claude",
  "defaultModel": "claude-sonnet-4-20250514",
  "providers": {
    "openrouter": {
      "apiKey": "sk-or-...",
      "enabled": true
    },
    "anthropic": {
      "apiKey": "sk-ant-...",
      "enabled": true
    },
    "openai": {
      "apiKey": "sk-...",
      "enabled": false
    }
  },
  "agents": {
    "claude": {
      "installed": true,
      "defaultModel": "claude-sonnet-4"
    },
    "codex": {
      "installed": false
    }
  },
  "settings": {
    "autoApprove": false,
    "theme": "dark",
    "streamOutput": true
  }
}
```

### Zsh Plugin Configuration

After running `forge setup`, configure in `~/.zshrc`:

```bash
# Forge Zsh plugin
source ~/.forge/zsh-plugin.zsh

# Now you can use:
# : <prompt>  - Send prompt to Forge
# :help       - Show Forge help
```

### Environment Variables

| Variable | Description | Required | Example |
|----------|-------------|----------|---------|
| FORGE_API_KEY | Default API key | No | `sk-...` |
| OPENROUTER_API_KEY | OpenRouter key | Yes* | `sk-or-...` |
| ANTHROPIC_API_KEY | Anthropic key | Yes* | `sk-ant-...` |
| OPENAI_API_KEY | OpenAI key | Yes* | `sk-...` |
| FORGE_CONFIG | Config file path | No | `~/.forge/config.json` |
| FORGE_LOG_LEVEL | Log level | No | `DEBUG` |

*One provider required

### Configuration Locations (in order of precedence)

1. Command-line flags
2. Environment variables
3. User config (`~/.forge/config.json`)
4. Project config (`./.forge/config.json`)

## API/Protocol Endpoints

### ACP (Agent Client Protocol)

Forge implements ACP for agent interoperability:

```bash
# Start ACP server
forge acp

# Connect IDE via ACP
# VS Code, JetBrains, Zed supported
```

### Provider APIs

Forge connects to various LLM provider APIs:

| Provider | Base URL |
|----------|----------|
| OpenRouter | https://openrouter.ai/api/v1 |
| Anthropic | https://api.anthropic.com |
| OpenAI | https://api.openai.com |
| Local | http://localhost:11434/v1 |

## Usage Examples

### Example 1: Agent Selection

```bash
# List available agents
forge agents

# Install an agent
forge openhands install

# Run with specific agent
forge openhands "fix the bug in auth.py"

# Or in TUI
forge
/agent openhands
> fix the bug in auth.py
```

### Example 2: Multi-Agent Workflow

```bash
# Start Forge
forge

# Plan with muse
/muse
> Design a microservices architecture for an e-commerce platform

# Switch to forge for implementation
/forge
> Implement the API gateway service based on the plan

# Switch to claude for detailed work
/agent claude
> Add authentication middleware to the gateway
```

### Example 3: Zsh Integration

```bash
# After setup, use : prefix
: create a Python script that fetches weather data

# Continue conversation
: --continue

# With specific agent
: --agent claude refactor the database layer

# List available commands
: --help
```

### Example 4: Headless Usage

```bash
# Non-interactive mode
forge claude --print "explain this codebase"

# Save output
forge claude --print "generate tests" > tests.txt

# With specific model
forge claude --model opus --print "complex refactoring"
```

### Example 5: Session Management

```bash
# Start session
forge

# Do some work...

# Exit and continue later
# Ctrl+C or /exit

# Continue session
forge --continue

# Or specify session
forge --session <session-id>

# List sessions
forge conversation list
```

### Example 6: OpenRouter Multi-Model

```bash
# Login to OpenRouter
forge login openrouter

# Use any model through OpenRouter
forge --model anthropic/claude-sonnet-4 "task"
forge --model openai/gpt-4o "task"
forge --model google/gemini-pro "task"

# List available models
forge list model
```

## Troubleshooting

### Issue: Command Not Found

**Symptoms:** `forge: command not found`

**Solution:**
```bash
# Check installation
npm list -g @forge-agents/forge

# Reinstall
npm install -g @forge-agents/forge

# Check PATH
export PATH="$PATH:$(npm bin -g)"
```

### Issue: Zsh Plugin Not Working

**Symptoms:** `:` command not recognized

**Solution:**
```bash
# Check Zsh is being used
echo $SHELL

# Run setup again
forge setup

# Restart terminal
# Or reload config
source ~/.zshrc

# Check plugin loaded
which _forge_complete
```

### Issue: Agent Not Found

**Symptoms:** "Agent not installed" error

**Solution:**
```bash
# List available agents
forge agents

# Install agent
forge claude install

# Check agent status
forge claude check
```

### Issue: Authentication Errors

**Symptoms:** "Authentication failed" errors

**Solution:**
```bash
# Check login status
forge info

# Re-login
forge login

# Or set API key
export OPENROUTER_API_KEY="sk-or-..."
```

### Issue: Model Not Available

**Symptoms:** "Model not found" error

**Solution:**
```bash
# List models for agent
forge claude models

# Use available model
forge claude --model claude-sonnet-4 "task"
```

### Issue: Session Resume Fails

**Symptoms:** Cannot continue previous session

**Solution:**
```bash
# List conversations
forge conversation list

# Resume specific session
forge --session <id>

# Or use continue flag
forge --continue
```

---

**Last Updated:** 2026-04-02
**Version:** 1.0.x
