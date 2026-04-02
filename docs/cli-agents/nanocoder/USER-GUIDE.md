# Nanocoder - User Guide

> A beautiful local-first CLI coding agent running in your terminal - built by the community for the community.

## Table of Contents

- [Overview](#overview)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [CLI Commands](#cli-commands)
- [Interactive Commands](#interactive-commands)
- [Configuration](#configuration)
- [Usage Examples](#usage-examples)
- [Troubleshooting](#troubleshooting)

---

## Overview

Nanocoder is a local-first CLI coding agent that brings the power of agentic coding tools like Claude Code and Gemini CLI to local models or controlled APIs like OpenRouter. Built by the Nano Collective, it's designed with privacy and control in mind, supporting multiple AI providers with comprehensive tool support.

### Key Features

- **Local-First**: Run local models (Ollama) for maximum privacy
- **Multi-Provider Support**: OpenAI, Anthropic, OpenRouter, Ollama, Google Gemini, and more
- **MCP Integration**: Model Context Protocol for extended capabilities
- **Custom Commands**: Create reusable command templates
- **Task Management**: Built-in task tracking for complex work
- **Checkpointing**: Save and restore conversation states
- **Beautiful TUI**: Terminal UI with themes and customization
- **VS Code Extension**: Editor integration available
- **Skill System**: Agent Skills open standard support

---

## Installation

### Prerequisites

- **Node.js 18+**
- **npm or yarn**
- **Git** (for repository operations)

### Installation Methods

#### Option 1: NPM (Recommended)

```bash
# Install globally
npm install -g @nanocollective/nanocoder

# Or use npx (no installation)
npx @nanocollective/nanocoder
```

#### Option 2: Homebrew

```bash
# Add tap (if needed)
brew tap nanocollective/tap

# Install
brew install nanocoder
```

#### Option 3: Nix Flakes

```bash
# Run directly
nix run github:nanocollective/nanocoder

# Or install
nix profile install github:nanocollective/nanocoder
```

#### Option 4: Local Development

```bash
# Clone repository
git clone https://github.com/nanocollective/nanocoder.git
cd nanocoder

# Install dependencies
npm install

# Build
npm run build

# Link for global use
npm link

# Or run directly
npm start
```

### Verify Installation

```bash
# Check version
nanocoder --version

# Show help
nanocoder --help
```

---

## Quick Start

### 1. First Launch

```bash
# Start Nanocoder
nanocoder
```

### 2. Configuration Wizard

On first launch, you'll be guided through:

1. **Select Provider**: Choose your AI provider
   - OpenAI (GPT models)
   - Anthropic (Claude models)
   - OpenRouter (access to many models)
   - Ollama (local models)
   - Google (Gemini models)
   - Custom OpenAI-compatible API

2. **Enter API Key**: Securely store your API credentials

3. **Select Model**: Choose default model for conversations

### 3. Basic Interaction

```bash
# Start interactive mode
nanocoder

# Then type natural language:
> Create a Python script that fetches weather data

> Refactor this function to use async/await

> Explain the codebase structure
```

### 4. Non-Interactive Mode

```bash
# Execute a single command
nanocoder run "Create a React component for a button"

# With specific provider/model
nanocoder --provider openrouter --model google/gemini-3.1-flash run "Analyze src/app.ts"
```

---

## CLI Commands

### Global Flags

| Flag | Description |
|------|-------------|
| `--version` | Show version information |
| `--help` | Display help message |
| `--provider <name>` | Specify AI provider |
| `--model <name>` | Specify model to use |
| `--config <path>` | Use custom config file |
| `--verbose` | Enable verbose logging |
| `--quiet` | Suppress non-essential output |

### Core Commands

#### Interactive Mode

```bash
# Start interactive session
nanocoder

# With specific provider
nanocoder --provider ollama --model llama3.1
```

#### Non-Interactive Mode

```bash
# Run single command
nanocoder run "Your prompt here"

# Flags can appear before or after 'run'
nanocoder run --provider openrouter "refactor database module"
nanocoder --provider openrouter run "refactor database module"
```

#### Configuration

```bash
# Open configuration wizard
nanocoder config

# Set specific config value
nanocoder config set provider openai
nanocoder config set model gpt-4o

# Show current config
nanocoder config show

# Reset to defaults
nanocoder config reset
```

#### Model Management

```bash
# List available models
nanocoder models

# Switch model
nanocoder models set gpt-4o

# List models for specific provider
nanocoder models --provider openrouter
```

#### Provider Management

```bash
# List configured providers
nanocoder providers

# Add new provider
nanocoder providers add

# Remove provider
nanocoder providers remove <name>
```

#### MCP Server Management

```bash
# List MCP servers
nanocoder mcp list

# Add MCP server
nanocoder mcp add

# Remove MCP server
nanocoder mcp remove <name>

# Test MCP connection
nanocoder mcp test <name>
```

#### Custom Commands

```bash
# List custom commands
nanocoder commands

# Create new command
nanocoder commands create <name>

# Edit command
nanocoder commands edit <name>

# Delete command
nanocoder commands delete <name>
```

#### Checkpoint Management

```bash
# Save current state
nanocoder checkpoint save <name>

# List checkpoints
nanocoder checkpoint list

# Restore checkpoint
nanocoder checkpoint restore <name>

# Delete checkpoint
nanocoder checkpoint delete <name>
```

### Developer Commands

```bash
# Run in development mode
nanocoder dev

# Run tests
nanocoder test

# Show logs
nanocoder logs

# Clear cache
nanocoder cache clear
```

---

## Interactive Commands

### Built-in Slash Commands

Once inside Nanocoder interactive mode, use these commands:

#### Navigation

| Command | Description |
|---------|-------------|
| `/help` | Show available commands |
| `/exit` or `/quit` | Exit Nanocoder |
| `/clear` | Clear conversation history |

#### Model & Provider

| Command | Description |
|---------|-------------|
| `/provider` | Show current provider |
| `/provider <name>` | Switch provider |
| `/model` | Show current model |
| `/model <name>` | Switch model |

#### Context Management

| Command | Description |
|---------|-------------|
| `/context` | Show current context |
| `/compact` | Compress conversation history |
| `/usage` | Show token usage statistics |

#### File Operations

| Command | Description |
|---------|-------------|
| `/explorer` | Open interactive file browser |
| `@<file>` | Include file in context |
| `@<folder>/` | Include folder contents |

#### MCP Operations

| Command | Description |
|---------|-------------|
| `/mcp` | List connected MCP servers |
| `/mcp <name>` | Use specific MCP server |

#### Task Management

| Command | Description |
|---------|-------------|
| `/tasks` | List active tasks |
| `/tasks create <name>` | Create new task |
| `/tasks complete <id>` | Mark task complete |
| `/tasks delete <id>` | Delete task |

#### Settings

| Command | Description |
|---------|-------------|
| `/settings` | Open settings menu |
| `/theme <name>` | Change UI theme |

#### Custom Commands

| Command | Description |
|---------|-------------|
| `/<custom-command>` | Execute custom command |

### Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Tab` | Auto-complete |
| `↑` / `↓` | Navigate history |
| `Ctrl+C` | Cancel/Clear/Exit |
| `Ctrl+L` | Clear screen |
| `Ctrl+R` | Search history |
| `Esc` | Cancel current operation |

### File Explorer

The `/explorer` command provides interactive file browsing:

```
/commands
  /create
  /edit
  /delete
/src
  /components
    Button.tsx
    Input.tsx
  app.tsx
  main.tsx
/tests
  app.test.tsx
```

Features:
- Tree view navigation
- File preview with syntax highlighting
- Multi-file selection
- Search mode
- VS Code integration (open files in editor)

---

## Configuration

### Configuration File Location

```
~/.config/nanocoder/config.json
```

Or platform-specific:
- **macOS**: `~/Library/Application Support/nanocoder/config.json`
- **Windows**: `%APPDATA%/nanocoder/config.json`
- **Linux**: `~/.config/nanocoder/config.json`

### Configuration Structure

```json
{
  "version": "1.0",
  "providers": {
    "openai": {
      "id": "openai",
      "name": "OpenAI",
      "apiKey": "sk-...",
      "baseUrl": "https://api.openai.com/v1",
      "models": ["gpt-4o", "gpt-4o-mini", "gpt-3.5-turbo"]
    },
    "anthropic": {
      "id": "anthropic",
      "name": "Anthropic",
      "apiKey": "sk-ant-...",
      "baseUrl": "https://api.anthropic.com/v1",
      "models": ["claude-3-5-sonnet", "claude-3-opus", "claude-3-haiku"]
    },
    "openrouter": {
      "id": "openrouter",
      "name": "OpenRouter",
      "apiKey": "sk-or-...",
      "baseUrl": "https://openrouter.ai/api/v1"
    },
    "ollama": {
      "id": "ollama",
      "name": "Ollama",
      "apiKey": "",
      "baseUrl": "http://localhost:11434",
      "models": ["llama3.1", "codellama", "mistral"]
    }
  },
  "defaultProvider": "openai",
  "defaultModel": "gpt-4o",
  "settings": {
    "theme": "auto",
    "streamResponses": true,
    "showTokenCount": true,
    "confirmDestructive": true,
    "autoFormat": true
  },
  "mcp": {
    "servers": {
      "filesystem": {
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-filesystem", "."]
      },
      "git": {
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-git"]
      }
    }
  },
  "customCommands": {
    "test": {
      "description": "Create tests for code",
      "template": "Write comprehensive tests for {file} covering edge cases and error scenarios."
    }
  }
}
```

### Project-Specific Configuration

Create `agents.config.json` or `.nanocoder/config.json` in project root:

```json
{
  "provider": "ollama",
  "model": "llama3.1",
  "mcp": {
    "servers": {
      "project-context": {
        "command": "node",
        "args": ["./scripts/mcp-server.js"]
      }
    }
  },
  "commands": {
    "test": {
      "description": "Run project tests",
      "template": "Run the test suite and report any failures."
    }
  }
}
```

### Custom Commands Directory

```
project-root/
├── .nanocoder/
│   ├── commands/
│   │   ├── test.md
│   │   ├── refactor.md
│   │   └── document.md
│   └── tasks.json
└── src/
```

### Custom Command Format

**`.nanocoder/commands/test.md`:**

```markdown
---
name: test
description: Create comprehensive tests
parameters:
  - name: file
    type: string
    required: true
    description: File to test
  - name: type
    type: string
    required: false
    default: unit
    description: Test type (unit, integration, e2e)
---

Write {type} tests for {file} that cover:
- Happy path scenarios
- Edge cases and boundary conditions
- Error handling and exceptions
- Integration with dependencies

Use the project's testing framework and follow existing patterns.
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `OPENAI_API_KEY` | OpenAI API key |
| `ANTHROPIC_API_KEY` | Anthropic API key |
| `OPENROUTER_API_KEY` | OpenRouter API key |
| `NANOCODER_CONFIG` | Custom config file path |
| `NANOCODER_THEME` | Override theme |
| `NANOCODER_LOG_LEVEL` | Logging level (debug, info, warn, error) |

### User Preferences

```bash
# Open settings UI
/settings

# Or edit directly
~/.config/nanocoder/preferences.json
```

Preferences include:
- UI theme (dark, light, auto)
- Message shapes (rounded, square)
- Compact mode
- Token count display
- Keyboard shortcuts

---

## Usage Examples

### Example 1: Basic Coding Tasks

```bash
# Start interactive mode
nanocoder

# Then:
> Create a Python function to validate email addresses

> Refactor the UserService class to use dependency injection

> Explain how the authentication middleware works
```

### Example 2: Using Local Models (Ollama)

```bash
# Start Ollama first
ollama serve

# Use with Nanocoder
nanocoder --provider ollama --model llama3.1

# All requests stay local - maximum privacy
```

### Example 3: Custom Commands

```bash
# Create custom command
nanocoder commands create review

# Edit .nanocoder/commands/review.md
# Use with:
/review file="src/app.ts"

# Or another example:
/refactor target="old-pattern" replacement="new-pattern"
```

### Example 4: File Operations

```bash
# Use @ to mention files
> Review @src/components/Header.tsx for accessibility issues

> Add error handling to @src/utils/api.ts

> Create tests for all files in @src/helpers/
```

### Example 5: MCP Integration

```bash
# With filesystem MCP configured
> Read the contents of @README.md

> List all TypeScript files in the project

> Search for functions that use fetch API
```

### Example 6: Task Management

```bash
# In interactive mode:
/tasks create "Implement user authentication"
/tasks create "Add password reset flow"
/tasks create "Write API documentation"

# Work on tasks
> Implement the login endpoint

# Mark complete
/tasks complete 1

# View remaining
/tasks
```

### Example 7: Checkpoints

```bash
# Save before major refactoring
nanocoder checkpoint save "before-refactor"

# Do refactoring work...

# If something goes wrong, restore
nanocoder checkpoint restore "before-refactor"
```

### Example 8: Multi-Provider Workflow

```bash
# Use GPT-4 for architecture
nanocoder --provider openai --model gpt-4o run "Design a microservices architecture"

# Use Claude for implementation
nanocoder --provider anthropic --model claude-3-5-sonnet run "Implement the API gateway"

# Use local model for quick checks
nanocoder --provider ollama --model llama3.1 run "Review this function"
```

### Example 9: VS Code Integration

```bash
# Install VS Code extension
# Extensions → Search "Nanocoder" → Install

# Use within VS Code:
# - Select code → Right click → "Ask Nanocoder"
# - Command Palette → "Nanocoder: Start Session"
# - Integrated terminal with Nanocoder
```

---

## Troubleshooting

### Installation Issues

#### "npm install fails with permission error"

```bash
# Use npx instead
npx @nanocollective/nanocoder

# Or fix npm permissions
sudo chown -R $(whoami) ~/.npm

# Or use node version manager
nvm use 20
```

#### "command not found: nanocoder"

```bash
# Ensure npm global bin is in PATH
export PATH="$PATH:$(npm bin -g)"

# Or use npx
npx @nanocollective/nanocoder
```

#### "Homebrew install fails"

```bash
# Update Homebrew
brew update
brew install nanocoder

# Or try from npm
npm install -g @nanocollective/nanocoder
```

### Configuration Issues

#### "API key not found"

```bash
# Run config wizard
nanocoder config

# Or set environment variable
export OPENAI_API_KEY="sk-..."
```

#### "Model not available"

```bash
# List available models
nanocoder models

# Switch to available model
nanocoder models set gpt-4o
```

#### "Config file is corrupted"

```bash
# Reset configuration
nanocoder config reset

# Or manually delete
rm ~/.config/nanocoder/config.json
```

### Runtime Issues

#### "Connection timeout"

```bash
# Check internet connection
ping google.com

# Try different provider
nanocoder --provider ollama

# Check API status (for cloud providers)
```

#### "Context limit exceeded"

```bash
# Compact conversation
/compact

# Or clear and start fresh
/clear
```

#### "Tool execution failed"

```bash
# Check MCP servers are running
/mcp

# Restart MCP server
nanocoder mcp test <name>

# Check permissions
ls -la .
```

#### "Token count not showing"

```bash
# Enable in settings
/settings → Show token count: true

# Or in config:
{
  "settings": {
    "showTokenCount": true
  }
}
```

### Ollama-Specific Issues

#### "Ollama connection refused"

```bash
# Start Ollama server
ollama serve

# Check if Ollama is running
curl http://localhost:11434/api/tags

# Verify model is pulled
ollama pull llama3.1
```

#### "Model not found"

```bash
# List available models
ollama list

# Pull model
ollama pull llama3.1
```

### Display Issues

#### "UI looks broken"

```bash
# Ensure terminal supports Unicode
# Try different terminal emulator

# Set TERM
export TERM=xterm-256color

# Check terminal size
stty size
```

#### "Colors not rendering"

```bash
# Force theme
nanocoder --theme dark

# Or environment variable
export NANOCODER_THEME=dark
```

### Debug Mode

```bash
# Enable debug logging
export NANOCODER_LOG_LEVEL=debug
nanocoder

# Verbose output
nanocoder --verbose

# Show config
nanocoder config show
```

### Getting Help

```bash
# Built-in help
/help

# Command-specific help
/help /explorer

# Online documentation
# https://docs.nanocollective.org

# Community Discord
# Link in GitHub repository
```

### Reporting Issues

1. Check existing issues: https://github.com/nanocollective/nanocoder/issues
2. Include:
   - Nanocoder version
   - Node.js version
   - Operating system
   - Config (redact API keys)
   - Debug logs
   - Steps to reproduce

---

## Resources

- **GitHub**: https://github.com/nanocollective/nanocoder
- **Documentation**: https://docs.nanocollective.org
- **NPM Package**: https://www.npmjs.com/package/@nanocollective/nanocoder
- **Discord**: Community link in repository
- **VS Code Extension**: Search "Nanocoder" in VS Code marketplace

---

*Last updated: 2026-04-02*
