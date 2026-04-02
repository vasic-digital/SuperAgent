# Gemini CLI User Guide

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
npm install -g @google/gemini-cli

# Verify installation
gemini --version

# Upgrade
npm install -g @google/gemini-cli@latest
```

### Method 2: npx (No Installation)

```bash
# Run without installing
npx @google/gemini-cli

# Run latest from GitHub main branch
npx https://github.com/google-gemini/gemini-cli
```

### Method 3: Homebrew (macOS/Linux)

```bash
# Install using Homebrew
brew install gemini-cli
```

### Method 4: MacPorts (macOS)

```bash
sudo port install gemini-cli
```

### Method 5: Anaconda

```bash
# Create and activate environment
conda create -y -n gemini_env -c conda-forge nodejs
conda activate gemini_env

# Install Gemini CLI
npm install -g @google/gemini-cli
```

### Method 6: Build from Source

```bash
# Clone repository
git clone https://github.com/google-gemini/gemini-cli.git
cd gemini-cli

# Install dependencies
npm install

# Build
npm run build

# Link for development
npm link packages/cli
```

### Prerequisites

- Node.js 18 or higher
- Google account (for free tier)
- Google AI Studio API key (optional, for higher limits)

## Quick Start

### First-Time Setup

```bash
# Verify installation
gemini --version

# Start Gemini CLI
gemini

# Select theme (on first run)
# Choose from available color themes using arrow keys

# Authenticate
# Browser will open for Google account sign-in
```

### Basic Usage

```bash
# Start interactive session
gemini

# Run with a prompt
gemini "explain this codebase"

# Run with specific model
gemini -m qwen3-coder "write a Python function"

# Run with image
gemini "analyze this" --image screenshot.png
```

### Hello World

```bash
# Start Gemini CLI
gemini

# At the prompt, type:
> Create a Python script that prints "Hello, World!"

# Or non-interactive:
gemini "Create a Python script that prints 'Hello, World!'"
```

## CLI Commands

### Global Options

| Option | Short | Description | Example |
|--------|-------|-------------|---------|
| --version | -v | Show version | `gemini --version` |
| --help | -h | Show help | `gemini --help` |
| --model | -m | Select model | `gemini -m gemini-2.5-pro` |
| --image | -i | Attach image | `gemini -i image.png "analyze"` |
| --sandbox | -s | Run in sandbox | `gemini --sandbox "task"` |
| --yes | -y | Auto-approve | `gemini -y "task"` |
| --output | -o | Output file | `gemini -o result.md "task"` |
| --json | | JSON output | `gemini --json "task"` |
| --stream | | Stream output | `gemini --stream "task"` |
| --no-stream | | No streaming | `gemini --no-stream "task"` |

### Command: extensions

**Description:** Manage Gemini CLI extensions.

**Usage:**
```bash
gemini extensions [SUBCOMMAND]
```

**Subcommands:**
| Subcommand | Description |
|------------|-------------|
| list | List installed extensions |
| install | Install an extension |
| uninstall | Uninstall an extension |
| update | Update extensions |

**Examples:**
```bash
# List extensions
gemini extensions list

# Install extension
gemini extensions install https://github.com/user/extension

# Install from GitHub
gemini extensions install gh:stripe/stripe-cli-gemini

# Uninstall
gemini extensions uninstall extension-name
```

### Command: config

**Description:** Manage Gemini CLI configuration.

**Usage:**
```bash
gemini config [SUBCOMMAND]
```

**Subcommands:**
| Subcommand | Description |
|------------|-------------|
| get | Get config value |
| set | Set config value |
| list | List all config |

**Examples:**
```bash
# Set default model
gemini config set model gemini-2.5-pro

# Get current theme
gemini config get theme

# List all settings
gemini config list
```

### Command: auth

**Description:** Manage authentication.

**Usage:**
```bash
gemini auth [SUBCOMMAND]
```

**Subcommands:**
| Subcommand | Description |
|------------|-------------|
| login | Log in to Google |
| logout | Log out |
| status | Check auth status |

**Examples:**
```bash
# Log in
gemini auth login

# Check status
gemini auth status

# Log out
gemini auth logout
```

### Command: mcp

**Description:** Manage MCP (Model Context Protocol) servers.

**Usage:**
```bash
gemini mcp [SUBCOMMAND]
```

**Subcommands:**
| Subcommand | Description |
|------------|-------------|
| list | List MCP servers |
| add | Add MCP server |
| remove | Remove MCP server |
| enable | Enable MCP server |
| disable | Disable MCP server |

**Examples:**
```bash
# List MCP servers
gemini mcp list

# Add filesystem MCP
gemini mcp add filesystem npx -y @modelcontextprotocol/server-filesystem /path

# Enable MCP server
gemini mcp enable github
```

### Command: theme

**Description:** Change color theme.

**Usage:**
```bash
gemini theme [THEME_NAME]
```

**Examples:**
```bash
# List themes
gemini theme

# Set dark theme
gemini theme dark

# Set light theme
gemini theme light
```

## TUI/Interactive Commands

When running in interactive/TUI mode, use these commands:

| Command | Shortcut | Description | Example |
|---------|----------|-------------|---------|
| /help | ? | Show help | `/help` |
| /exit | Ctrl+C | Exit Gemini CLI | `/exit` |
| /clear | Ctrl+L | Clear screen | `/clear` |
| /model | | Switch model | `/model gemini-2.5-pro` |
| /web | | Web search | `/web "documentation"` |
| /file | | Include file | `/file path/to/file.js` |
| /image | | Include image | `/image screenshot.png` |
| /reset | | Reset conversation | `/reset` |
| /undo | | Undo last action | `/undo` |
| /redo | | Redo last action | `/redo` |
| /save | | Save conversation | `/save chat.md` |
| /load | | Load conversation | `/load chat.md` |
| /extensions | | Manage extensions | `/extensions list` |
| /mcp | | Manage MCP | `/mcp list` |

### Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| Ctrl+C | Cancel / Exit |
| Ctrl+L | Clear screen |
| Up/Down | Navigate history |
| Tab | Autocomplete |
| Ctrl+K | Clear line |

## Configuration

### Configuration File Format

Gemini CLI uses `~/.gemini/config.json`:

```json
{
  "model": "gemini-2.5-pro",
  "theme": "dark",
  "autoApprove": false,
  "sandbox": true,
  "stream": true,
  "outputFormat": "text",
  "extensions": {
    "enabled": ["stripe", "github"]
  },
  "mcp": {
    "servers": {
      "filesystem": {
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-filesystem", "/home/user"]
      },
      "github": {
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-github"],
        "env": {
          "GITHUB_PERSONAL_ACCESS_TOKEN": "ghp_..."
        }
      }
    }
  },
  "api": {
    "key": "AIza...",
    "baseUrl": "https://generativelanguage.googleapis.com"
  }
}
```

### Environment Variables

| Variable | Description | Required | Example |
|----------|-------------|----------|---------|
| GOOGLE_API_KEY | Google AI API key | Yes* | `AIza...` |
| GEMINI_MODEL | Default model | No | `gemini-2.5-pro` |
| GEMINI_THEME | Color theme | No | `dark` |
| GEMINI_SANDBOX | Use sandbox | No | `true` |
| GEMINI_CONFIG | Config file path | No | `~/.gemini/config.json` |
| GOOGLE_GENAI_USE_VERTEXAI | Use Vertex AI | No | `true` |

*Required for API key authentication (optional for OAuth)

### Configuration Locations (in order of precedence)

1. Command-line flags
2. Environment variables (`GEMINI_*`)
3. User config (`~/.gemini/config.json`)
4. Project config (`./.gemini/config.json`)

## API/Protocol Endpoints

### Gemini API

For programmatic access with API key:

**Base URL:** `https://generativelanguage.googleapis.com/v1beta`

**Example Request:**
```bash
curl "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-pro:generateContent?key=$GOOGLE_API_KEY" \
  -H 'Content-Type: application/json' \
  -X POST \
  -d '{
    "contents": [{
      "parts":[{
        "text": "Write a Python hello world program"
      }]
    }]
  }'
```

**Response:**
```json
{
  "candidates": [{
    "content": {
      "parts": [{
        "text": "print('Hello, World!')"
      }]
    }
  }]
}
```

### MCP Server Configuration

Extensions/MCP servers in `~/.gemini/mcp.json`:

```json
{
  "mcpServers": {
    "context7": {
      "command": "npx",
      "args": ["-y", "@upstash/context7-mcp"],
      "env": {
        "DEFAULT_MINIMUM_TOKENS": "6000"
      }
    },
    "fetch": {
      "command": "uvx",
      "args": ["mcp-server-fetch"]
    }
  }
}
```

### Extensions

Install extensions for additional capabilities:

```bash
# Install Stripe extension
gemini extensions install gh:stripe/stripe-cli-gemini

# Install Figma extension
gemini extensions install gh:figma/figma-cli-gemini

# List available extensions
gemini extensions list
```

## Usage Examples

### Example 1: Code Generation

```bash
# Start Gemini CLI
gemini

# Request code
> Create a React component for a todo list with TypeScript

# With image reference
> Create CSS to match this design [paste image]
```

### Example 2: Codebase Analysis

```bash
# Analyze codebase
gemini "explain the architecture of this codebase"

# With specific focus
gemini "find security vulnerabilities in the authentication module"
```

### Example 3: Web Search Integration

```bash
# Search and implement
gemini
> /web "React 19 new features"
> Implement an example using the new use hook
```

### Example 4: Extension Usage

```bash
# Install and use extension
gemini extensions install gh:stripe/stripe-cli-gemini

# Use in session
gemini
> Create a Stripe checkout session for a $20 product
```

### Example 5: Sandbox Mode

```bash
# Run in sandbox for safety
gemini --sandbox -y "install dependencies and run tests"

# Or in TUI
/sandbox enable
```

### Example 6: MCP Integration

```bash
# Configure MCP first
gemini mcp add github npx -y @modelcontextprotocol/server-github

# Use MCP tools
gemini
> List my recent GitHub repositories
> Create an issue for this bug
```

## Troubleshooting

### Issue: Command Not Found

**Symptoms:** `gemini: command not found`

**Solution:**
```bash
# Check npm global installation
npm list -g @google/gemini-cli

# Reinstall
npm install -g @google/gemini-cli

# Check PATH
export PATH="$PATH:$(npm bin -g)"

# Or use npx
npx @google/gemini-cli
```

### Issue: Authentication Fails

**Symptoms:** "Authentication failed" or "Sign in required"

**Solution:**
```bash
# Check auth status
gemini auth status

# Log out and back in
gemini auth logout
gemini auth login

# Or use API key
export GOOGLE_API_KEY="AIza..."
```

### Issue: Model Not Available

**Symptoms:** "Model not found" or "Invalid model"

**Solution:**
```bash
# List available models
/model

# Use default model
gemini -m gemini-2.5-pro "task"

# Check model access
# Free tier: gemini-2.5-pro
# With API key: More models available
```

### Issue: Rate Limit Exceeded

**Symptoms:** "Rate limit exceeded" error

**Solution:**
- Free tier: 60 requests/minute, 1,000/day
- Wait before retrying
- Add API key for higher limits
- Upgrade to Gemini Code Assist

### Issue: Extensions Not Working

**Symptoms:** Extension commands fail

**Solution:**
```bash
# List installed extensions
gemini extensions list

# Update extensions
gemini extensions update

# Reinstall extension
gemini extensions uninstall <name>
gemini extensions install <url>
```

### Issue: MCP Servers Not Connecting

**Symptoms:** MCP tools unavailable

**Solution:**
```bash
# Check MCP config
gemini mcp list

# Verify server is installed
which npx

# Check MCP logs
# Enable debug mode
gemini --verbose
```

### Issue: Node.js Version Error

**Symptoms:** "Node.js 18+ required"

**Solution:**
```bash
# Check Node version
node --version

# Update Node.js
# Using nvm
nvm install 20
nvm use 20

# Or download from nodejs.org
```

### Issue: Slow Performance

**Symptoms:** Long response times

**Solution:**
```bash
# Disable streaming (for debugging)
gemini --no-stream "task"

# Use faster model
gemini -m gemini-2.5-flash "task"

# Check internet connection
```

---

**Last Updated:** 2026-04-02
**Version:** 0.35.1
