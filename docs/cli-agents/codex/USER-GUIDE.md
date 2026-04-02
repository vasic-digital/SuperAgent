# Codex CLI User Guide

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
npm install -g @openai/codex

# Verify installation
codex --version

# Upgrade to latest
npm install -g @openai/codex@latest
```

### Method 2: Homebrew (macOS/Linux)

```bash
# Install using Homebrew
brew install --cask codex

# Or traditional formula
brew install codex
```

### Method 3: Direct Binary Download

```bash
# macOS ARM (Apple Silicon)
curl -L -o codex "https://github.com/openai/codex/releases/latest/download/codex-darwin-arm64"
chmod +x codex
sudo mv codex /usr/local/bin/

# macOS x86_64
curl -L -o codex "https://github.com/openai/codex/releases/latest/download/codex-darwin-x64"
chmod +x codex
sudo mv codex /usr/local/bin/

# Linux x86_64
curl -L -o codex "https://github.com/openai/codex/releases/latest/download/codex-linux-x64"
chmod +x codex
sudo mv codex /usr/local/bin/

# Linux ARM64
curl -L -o codex "https://github.com/openai/codex/releases/latest/download/codex-linux-arm64"
chmod +x codex
sudo mv codex /usr/local/bin/
```

### Method 4: Build from Source

```bash
# Clone repository
git clone https://github.com/openai/codex.git
cd codex

# Install dependencies (requires Rust)
cargo build --release

# The binary will be at target/release/codex
sudo cp target/release/codex /usr/local/bin/
```

### Prerequisites

- Node.js 18+ (for NPM install)
- GitHub account or OpenAI API key
- ChatGPT Plus, Pro, Business, Edu, or Enterprise plan

## Quick Start

### First-Time Setup

```bash
# Verify installation
codex --version

# Start Codex CLI
codex

# Authenticate (opens browser)
# Choose:
# - ChatGPT account (for Plus/Pro/Team/Edu/Enterprise users)
# - API key (for pay-per-use billing)
```

### Basic Usage

```bash
# Start interactive session
codex

# Run with a direct prompt
codex "explain this codebase"

# Run in non-interactive mode
codex exec "fix the CI failure"

# Run with image input
codex "describe this UI" --image screenshot.png
```

### Hello World

```bash
# Start Codex CLI
codex

# At the prompt, type:
> Create a Python script that prints "Hello, World!"

# Or non-interactive:
codex "Create a Python script that prints 'Hello, World!'"
```

## CLI Commands

### Global Options

| Option | Short | Description | Example |
|--------|-------|-------------|---------|
| --version | -v | Show version | `codex --version` |
| --help | -h | Show help | `codex --help` |
| --image | -i | Attach image | `codex -i screenshot.png "analyze"` |
| --approval-mode | -a | Set approval mode | `codex -a auto` |
| --model | -m | Select model | `codex -m gpt-5.4` |
| --reasoning | -r | Set reasoning effort | `codex -r high` |
| --output | -o | Output file | `codex -o output.txt "task"` |
| --quiet | -q | Suppress output | `codex -q "task"` |
| --dry-run | | Show plan without executing | `codex --dry-run "task"` |
| --no-tty | | Non-interactive mode | `codex --no-tty "task"` |

### Command: exec

**Description:** Execute a task non-interactively and exit.

**Usage:**
```bash
codex exec [OPTIONS] "PROMPT"
```

**Options:**
| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| --approval-mode | string | No | "suggest" | auto, suggest, or manual |
| --model | string | No | "gpt-5.3-codex" | Model to use |
| --max-turns | int | No | 50 | Maximum agent turns |

**Examples:**
```bash
# Execute a task
codex exec "fix linting errors"

# Auto-approve all changes
codex exec --approval-mode auto "refactor auth module"

# Use specific model
codex exec --model gpt-5.4 "implement feature"
```

**Exit Codes:**
| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid arguments |
| 3 | Authentication error |
| 4 | Rate limit exceeded |

### Command: review

**Description:** Review code changes before committing.

**Usage:**
```bash
codex review [OPTIONS] [PATH]
```

**Options:**
| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| --diff | boolean | No | false | Review staged changes |
| --pr | string | No | | Review a pull request |

**Examples:**
```bash
# Review current changes
codex review

# Review staged changes
codex review --diff

# Review specific PR
codex review --pr 123
```

### Command: init

**Description:** Initialize a new Codex environment.

**Usage:**
```bash
codex init [OPTIONS]
```

**Examples:**
```bash
# Initialize in current directory
codex init
```

### Command: configure

**Description:** Configure Codex CLI settings.

**Usage:**
```bash
codex configure [SUBCOMMAND]
```

**Subcommands:**
| Subcommand | Description |
|------------|-------------|
| set | Set a configuration value |
| get | Get a configuration value |
| list | List all configuration |

**Examples:**
```bash
# Set approval mode
codex configure set approval-mode auto

# Get current model
codex configure get model

# List all settings
codex configure list
```

### Command: cloud

**Description:** Launch a Codex Cloud task.

**Usage:**
```bash
codex cloud [OPTIONS] "PROMPT"
```

**Options:**
| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| --environment | string | No | "default" | Cloud environment |
| --timeout | int | No | 3600 | Timeout in seconds |

**Examples:**
```bash
# Run task in cloud
codex cloud "implement payment processing"

# With specific environment
codex cloud --environment production "deploy"
```

### Command: hooks

**Description:** Manage Codex hooks.

**Usage:**
```bash
codex hooks [SUBCOMMAND]
```

**Subcommands:**
| Subcommand | Description |
|------------|-------------|
| list | List installed hooks |
| install | Install a hook |
| uninstall | Uninstall a hook |

**Examples:**
```bash
# List hooks
codex hooks list

# Install pre-commit hook
codex hooks install pre-commit
```

## TUI/Interactive Commands

When running in interactive/TUI mode, use these commands:

| Command | Shortcut | Description | Example |
|---------|----------|-------------|---------|
| /help | ? | Show available commands | `/help` |
| /exit | Ctrl+D | Exit Codex CLI | `/exit` |
| /clear | Ctrl+L | Clear screen | `/clear` |
| /model | | Switch model | `/model gpt-5.4` |
| /mode | | Change approval mode | `/mode auto` |
| /approvals | | Show approval settings | `/approvals` |
| /history | | Show conversation history | `/history` |
| /reset | | Reset conversation | `/reset` |
| /save | | Save conversation | `/save session.md` |
| /load | | Load conversation | `/load session.md` |
| /subagent | | Create subagent | `/subagent "task"` |
| /search | | Web search | `/search "documentation"` |

### Approval Modes

| Mode | Description |
|------|-------------|
| auto | Automatically approve safe actions |
| suggest | Suggest actions but require approval |
| manual | Require approval for every action |

### Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| Tab | Accept suggestion |
| Ctrl+C | Cancel current operation |
| Ctrl+D | Exit |
| Ctrl+L | Clear screen |
| Up/Down | Navigate history |

## Configuration

### Configuration File Format

Codex CLI uses `~/.codex/config.json`:

```json
{
  "model": "gpt-5.3-codex",
  "approvalMode": "suggest",
  "reasoningEffort": "medium",
  "defaultTimeout": 3600,
  "hooks": {
    "pre-command": "echo 'Starting task'",
    "post-command": "echo 'Task complete'"
  },
  "mcp": {
    "servers": {
      "github": {
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-github"]
      }
    }
  },
  "ssl": {
    "certFile": "/path/to/cert.pem",
    "keyFile": "/path/to/key.pem"
  }
}
```

### Environment Variables

| Variable | Description | Required | Example |
|----------|-------------|----------|---------|
| OPENAI_API_KEY | OpenAI API key | Yes* | `sk-...` |
| CODEX_API_KEY | Codex-specific API key | No | `codex-...` |
| SSL_CERT_FILE | SSL certificate file | No | `/path/to/cert.pem` |
| SSL_KEY_FILE | SSL key file | No | `/path/to/key.pem` |
| HTTP_PROXY | HTTP proxy URL | No | `http://proxy:8080` |
| HTTPS_PROXY | HTTPS proxy URL | No | `https://proxy:8080` |
| CODEX_CONFIG_PATH | Custom config path | No | `~/.codex/custom.json` |

*Required if not using ChatGPT account

### Configuration Locations (in order of precedence)

1. Command-line flags
2. Environment variables
3. Project config (`./.codex/config.json`)
4. User config (`~/.codex/config.json`)
5. System config (`/etc/codex/config.json`)

## API/Protocol Endpoints

### Codex Cloud API

For programmatic access, Codex CLI supports cloud tasks:

**Request:**
```bash
curl -X POST https://api.openai.com/v1/codex/cloud \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Implement authentication",
    "environment": "production",
    "timeout": 3600
  }'
```

**Response:**
```json
{
  "id": "task_abc123",
  "status": "running",
  "url": "https://codex.openai.com/tasks/abc123"
}
```

### MCP Server Configuration

Create `~/.codex/mcp.json`:

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/allowed/path"]
    },
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "ghp_..."
      }
    }
  }
}
```

## Usage Examples

### Example 1: Codebase Exploration

```bash
# Start Codex CLI
codex

# Ask about the codebase
> Explain the architecture of this project

# Find specific code
> Where is the authentication middleware defined?

# Understand complex code
> Explain what this function does and trace its dependencies
```

### Example 2: Bug Fixing

```bash
# Non-interactive bug fix
codex exec "fix the null pointer exception in auth.js"

# Interactive debugging
codex
> The tests are failing with timeout errors. Debug and fix the issue.
> Show me the stack trace
> Fix the async/await issue in the database connection
```

### Example 3: Feature Implementation

```bash
# Implement a complete feature
codex exec --approval-mode suggest "implement a REST API for user management with CRUD operations"

# Step-by-step implementation
codex
> Create a new React component for a data table
> Add sorting functionality
> Add pagination
> Style it with CSS modules
```

### Example 4: Code Review

```bash
# Review current changes
codex review

# Review staged changes
codex review --diff

# Review PR
codex review --pr 42

# Interactive review
codex
> Review the changes in src/components/ and suggest improvements
```

### Example 5: Working with Subagents

```bash
# In interactive mode
codex
> /subagent "optimize database queries"

# Create parallel subagents for complex tasks
> /subagent "write unit tests for auth module"
> /subagent "write integration tests for auth module"
```

### Example 6: Web Search Integration

```bash
# Search for documentation
codex
> /search "React 19 use hook documentation"

# Use search results in implementation
> Find the latest best practices for error handling in Node.js and implement them
```

## Troubleshooting

### Issue: Authentication Errors

**Symptoms:** "Authentication failed" or "Invalid API key"

**Solution:**
```bash
# Check if logged in
codex configure get api_key

# Re-authenticate
codex login

# Or set API key manually
export OPENAI_API_KEY="sk-..."
```

### Issue: Rate Limiting

**Symptoms:** "Rate limit exceeded" errors

**Solution:**
```bash
# Check current usage
codex usage

# Wait before retrying
# Or upgrade your plan for higher limits

# Use a different model temporarily
codex -m gpt-5-mini "task"
```

### Issue: Command Execution Blocked

**Symptoms:** Codex cannot run shell commands

**Solution:**
```bash
# Check approval mode
codex configure get approval-mode

# Change to auto mode
codex configure set approval-mode auto

# Or use suggest mode with specific allowed commands
codex -a suggest "run tests"
```

### Issue: Model Not Available

**Symptoms:** "Model not found" or "Invalid model" errors

**Solution:**
```bash
# List available models
/model

# Use default model
codex -m gpt-5.3-codex "task"

# Check model access in your plan
```

### Issue: SSL/Certificate Errors

**Symptoms:** SSL certificate verification errors

**Solution:**
```bash
# Set custom CA certificate
export SSL_CERT_FILE=/path/to/ca-cert.pem

# For enterprise proxies
codex configure set ssl.certFile /path/to/cert.pem
codex configure set ssl.keyFile /path/to/key.pem
```

### Issue: Windows Compatibility

**Symptoms:** Codex CLI not working on Windows

**Solution:**
- Use WSL (Windows Subsystem for Linux) for best experience
- Install Git for Windows
- Use Windows Terminal or PowerShell
- Note: Windows support is experimental

### Issue: Slow Performance

**Symptoms:** Long response times

**Solution:**
```bash
# Use faster model
codex -m gpt-5-mini "task"

# Reduce context by focusing on specific files
codex exec "analyze src/components/Button.js"

# Enable caching
codex configure set cache.enabled true
```

---

**Last Updated:** 2026-04-02
**Version:** 0.116.0
