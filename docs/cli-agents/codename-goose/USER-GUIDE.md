# Codename Goose - User Guide

**Codename Goose** is an open-source, extensible AI agent designed to automate engineering tasks. It runs locally on your machine and integrates with any LLM via MCP (Model Context Protocol) servers, providing a powerful CLI and desktop interface for AI-assisted development.

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

- **macOS 10.15+**, **Linux**, or **Windows (WSL2)**
- Terminal with Unicode support
- Git

### Method 1: Desktop App (macOS)

**Download:**
1. Download the latest release from [GitHub Releases](https://github.com/block/goose/releases)
2. Unzip the downloaded file
3. Run the executable

**Or via Homebrew:**
```bash
brew install --cask block-goose
```

### Method 2: CLI Installation (macOS/Linux)

```bash
curl -fsSL https://github.com/block/goose/releases/download/stable/download_cli.sh | bash
```

### Method 3: Windows (WSL2)

```bash
# In WSL2 terminal
curl -fsSL https://github.com/block/goose/releases/download/stable/download_cli.sh | bash
```

### Method 4: Build from Source

```bash
# Clone repository
git clone https://github.com/block/goose.git
cd goose

# Install Rust toolchain
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh

# Build
cargo build --release

# Install
cargo install --path .
```

### Verify Installation

```bash
# Check version
goose --version

# Check debug info
goose debug
```

---

## Quick Start

### First Launch

```bash
# Configure LLM provider (first time)
goose configure
```

Configuration options:
- **Quick Setup with API Key** - Automatic provider detection
- **ChatGPT Subscription** - Use ChatGPT Plus/Pro credentials
- **Agent Router by Tetrate** - Multi-model with automatic failover
- **OpenRouter** - Access 200+ models
- **Other Providers** - Manual configuration

### Start a Session

```bash
# Start interactive session
goose session start

# Start with name
goose session start --name "feature-auth"

# Start with extensions
goose session start --with-extension developer
```

### Your First Interaction

```bash
# In the session, type natural language:
> Create a React component for a login form

> Fix the bug in the authentication middleware

> Write tests for the user API endpoints
```

### Exit Session

```bash
# Type in session
/exit
# or
Ctrl+D
```

---

## CLI Commands

### Core Commands

| Command | Description |
|---------|-------------|
| `goose` | Start interactive session (alias for session start) |
| `goose --version` | Show version |
| `goose --help` | Show help |

### Session Management

| Command | Description |
|---------|-------------|
| `goose session start` | Start new session |
| `goose session start --name <name>` | Start named session |
| `goose session start --with-extension <ext>` | Start with extension |
| `goose session resume` | Resume last session |
| `goose session resume --id <id>` | Resume specific session |
| `goose session list` | List all sessions |

### Configuration

| Command | Description |
|---------|-------------|
| `goose configure` | Configure LLM provider |
| `goose configure --provider <provider>` | Configure specific provider |
| `goose config get <key>` | Get config value |
| `goose config set <key> <value>` | Set config value |

### Extension Management

| Command | Description |
|---------|-------------|
| `goose extension list` | List available extensions |
| `goose extension enable <name>` | Enable extension |
| `goose extension disable <name>` | Disable extension |
| `goose extension install <path>` | Install custom extension |

### Project Management

| Command | Alias | Description |
|---------|-------|-------------|
| `goose project` | `goose p` | Open last project or create new |
| `goose projects` | `goose ps` | Choose project to work on |

### Completion

| Command | Description |
|---------|-------------|
| `goose completion bash` | Generate bash completions |
| `goose completion zsh` | Generate zsh completions |
| `goose completion fish` | Generate fish completions |
| `goose completion powershell` | Generate PowerShell completions |

---

## TUI/Interactive Commands

### Interactive Session Features

### Slash Commands

| Command | Description |
|---------|-------------|
| `/help` | Show available commands |
| `/clear` | Clear conversation history |
| `/mode chat` | Switch to chat mode |
| `/mode plan` | Switch to plan mode |
| `/plan` | Create execution plan |
| `/prompts` | List prompts from extensions |
| `/prompts --extension <name>` | List extension prompts |
| `/builtin <extension>` | Add builtin extension |

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `↑/↓` | Navigate history |
| `Tab` | Autocomplete |
| `Ctrl+C` | Cancel operation |
| `Ctrl+D` | Exit session |

### Terminal Integration

Setup terminal aliases for direct access:

```bash
# Add to ~/.bashrc or ~/.zshrc
eval "$(goose completion bash)"

# Use @goose or @g alias
@goose create a python script to process these files
@g how do I fix these permission errors?
```

---

## Configuration

### Configuration Files

Goose stores configuration in:

- **macOS**: `~/.config/goose/config.yaml`
- **Linux**: `~/.config/goose/config.yaml`
- **Windows**: `%APPDATA%\goose\config.yaml`

### Basic Configuration

```yaml
# Goose configuration

# LLM Provider
provider:
  name: "openai"  # or "anthropic", "google", etc.
  api_key: ""     # Uses env var if empty
  model: "gpt-4"  # Model to use
  base_url: ""    # Custom API base URL

# Session settings
session:
  save_history: true
  max_history: 100
  
# Extensions
extensions:
  - name: "developer"
    enabled: true
  - name: "github"
    enabled: false
    
# MCP Servers
mcp:
  servers:
    - name: "filesystem"
      command: "npx"
      args: ["-y", "@modelcontextprotocol/server-filesystem", "/home/user"]
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `GOOSE_API_KEY` | Default API key |
| `OPENAI_API_KEY` | OpenAI API key |
| `ANTHROPIC_API_KEY` | Anthropic API key |
| `GOOGLE_API_KEY` | Google API key |
| `GOOSE_CONFIG` | Path to config file |
| `GOOSE_DATA` | Path to data directory |

### Supported Providers

| Provider | Configuration |
|----------|---------------|
| **OpenAI** | `provider: openai`, set `OPENAI_API_KEY` |
| **Anthropic** | `provider: anthropic`, set `ANTHROPIC_API_KEY` |
| **Google** | `provider: google`, set `GOOGLE_API_KEY` |
| **OpenRouter** | `provider: openrouter`, set `OPENROUTER_API_KEY` |
| **Tetrate** | `provider: tetrate` (built-in) |
| **Ollama** | `provider: ollama`, `base_url: http://localhost:11434` |

### Extensions

Built-in extensions:

| Extension | Description |
|-----------|-------------|
| `developer` | Core development tools (default) |
| `github` | GitHub integration |
| `google-drive` | Google Drive access |
| `jetbrains` | JetBrains IDE integration |

Enable extensions:

```bash
# In config
goose config set extensions.github.enabled true

# Or in session
/builtin github
```

### MCP Server Configuration

Add MCP servers in `~/.config/goose/config.yaml`:

```yaml
mcp:
  servers:
    - name: "fetch"
      command: "uvx"
      args: ["mcp-server-fetch"]
      
    - name: "sqlite"
      command: "uvx"
      args: ["mcp-server-sqlite", "--db-path", "~/data.db"]
      
    - name: "puppeteer"
      command: "npx"
      args: ["-y", "@modelcontextprotocol/server-puppeteer"]
```

### Permission Modes

Configure autonomy levels:

```yaml
permissions:
  # Options: ask, auto, danger
  file_edits: "ask"
  shell_commands: "ask"
  extensions: "ask"
```

Modes:
- **ask** - Always prompt before action
- **auto** - Auto-approve safe operations
- **danger** - Auto-approve all (use with caution)

---

## Usage Examples

### Basic Development Workflow

```bash
# 1. Navigate to project
cd ~/projects/my-app

# 2. Configure provider (first time)
goose configure

# 3. Start session
goose session start

# 4. Work with Goose
> Create a new API endpoint for user registration

# 5. Exit
/exit
```

### Project-Based Workflow

```bash
# Start project mode
goose project

# Or select from existing
goose projects

# Goose remembers project context
# Resume with:
goose project
```

### Session Management

```bash
# Start named session
goose session start --name "bugfix-123"

# Work on multiple tasks
# Switch between sessions
goose session list
goose session resume --id <id>
```

### Using Extensions

```bash
# Start with specific extensions
goose session start --with-extension developer --with-extension github

# Or enable during session
/builtin github

# Now you can:
> Create a PR for the current branch
> List open issues assigned to me
```

### MCP Server Usage

```bash
# With MCP servers configured, Goose can:
> Fetch the latest documentation from https://example.com/api

> Query the database for users created this week

> Take a screenshot of the login page

> Read my email about the deployment status
```

### Plan Mode

```bash
# Switch to plan mode
/mode plan

# Or use /plan command
/plan create a new authentication system

# Goose creates a plan:
# 1. Set up database schema
# 2. Create API endpoints
# 3. Implement JWT tokens
# 4. Add middleware
# 5. Write tests

# Review and approve each step
```

### Subagents

```bash
# Spawn subagent for parallel work
> Create a subagent to research the best React form libraries

# Main session continues while subagent works
# Subagent reports back with findings
```

### CI/CD Integration

```bash
# In CI environment, pin version
export GOOSE_VERSION=1.0.0
curl -fsSL https://github.com/block/goose/releases/download/stable/download_cli.sh | bash

# Non-interactive mode
echo "Fix the build errors" | goose
```

---

## Troubleshooting

### Installation Issues

#### "command not found" after install

```bash
# Check installation
which goose
ls ~/.local/bin/goose

# Add to PATH
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

#### Install script fails

```bash
# Check dependencies
which curl
which bash

# Try manual download
curl -L https://github.com/block/goose/releases/latest/download/goose-$(uname -s)-$(uname -m) -o goose
chmod +x goose
mv goose ~/.local/bin/
```

#### macOS "cannot be opened" warning

```bash
# Remove quarantine
xattr -d com.apple.quarantine ~/.local/bin/goose

# Or allow in System Preferences > Security & Privacy
```

### Configuration Issues

#### "No provider configured"

```bash
# Run configuration
goose configure

# Or set environment variable
export OPENAI_API_KEY="sk-..."
```

#### Provider connection fails

```bash
# Test API key
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer $OPENAI_API_KEY"

# Check config
cat ~/.config/goose/config.yaml

# Reconfigure
goose configure --provider openai
```

### Session Issues

#### Session won't start

```bash
# Check logs
tail ~/.config/goose/logs/goose.log

# Reset session state
rm -rf ~/.config/goose/sessions/

# Try again
goose session start
```

#### Extensions not loading

```bash
# List extensions
goose extension list

# Check enabled
goose config get extensions

# Enable extension
goose extension enable developer
```

### MCP Issues

#### MCP server not connecting

```bash
# Check MCP config
cat ~/.config/goose/config.yaml | grep -A5 mcp:

# Test MCP server manually
npx -y @modelcontextprotocol/server-fetch --help

# Check logs for errors
```

### Desktop App Issues

#### App won't launch (macOS M3)

```bash
# Check permissions
ls -la ~/.config/

# Fix permissions
chmod 755 ~/.config
chmod 755 ~/.config/goose

# Try launching again
```

#### Blank window on startup

```bash
# Clear app data
rm -rf ~/.config/goose/desktop/

# Restart app
```

### Common Errors

#### "Permission denied"

```bash
# Check config directory permissions
ls -la ~/.config/

# Fix
chmod 755 ~/.config/goose
chmod 644 ~/.config/goose/config.yaml
```

#### "Rate limit exceeded"

```bash
# Check provider dashboard
# Wait and retry

# Use different provider
goose configure --provider openrouter
```

### Debug Mode

```bash
# Enable debug logging
export GOOSE_DEBUG=1
goose session start

# Check logs
ls ~/.config/goose/logs/
tail -f ~/.config/goose/logs/goose.log
```

### Getting Help

```bash
# In-session help
/help

# CLI help
goose --help
goose session --help

# Documentation
# https://block.github.io/goose/

# GitHub
# https://github.com/block/goose/issues
```

---

## Best Practices

1. **Use Named Sessions**: Organize work with descriptive session names
2. **Enable Relevant Extensions**: Only load extensions you need
3. **Configure MCP Servers**: Extend capabilities with MCP
4. **Review Plans**: In plan mode, review before execution
5. **Version Control**: Commit before major changes
6. **Test Changes**: Verify AI-generated code works
7. **Use Subagents**: Delegate parallel tasks
8. **Set Permissions**: Choose appropriate permission level
9. **Update Regularly**: Keep Goose updated
10. **Backup Config**: Track config in dotfiles

---

## Advanced Features

### Custom Extensions

Create custom extensions:

```yaml
# ~/.config/goose/extensions/my-extension.yaml
name: my-extension
description: "My custom extension"
commands:
  - name: "custom-command"
    description: "Does something custom"
    script: |
      echo "Running custom command"
```

### Persistent Instructions

Add to `~/.config/goose/instructions.md`:

```markdown
Always follow these guidelines:
- Use TypeScript strict mode
- Write tests for all functions
- Follow existing code style
```

### Recipes

Create reusable prompt templates:

```bash
# Create recipe
mkdir -p ~/.config/goose/recipes
cat > ~/.config/goose/recipes/refactor.yaml << 'EOF'
name: refactor
description: "Refactor code pattern"
prompt: |
  Refactor the following code to use {{pattern}}:
  {{code}}
EOF

# Use in session
> /recipe refactor pattern=async/await code=@src/utils.ts
```

---

*Last Updated: April 2026*
