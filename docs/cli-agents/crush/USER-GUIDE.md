# Crush - User Guide

> Glamorous agentic coding for all - A beautiful terminal-based AI coding assistant by Charmbracelet.

## Table of Contents

- [Overview](#overview)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [CLI Commands](#cli-commands)
- [TUI/Interactive Commands](#tuiinteractive-commands)
- [Configuration](#configuration)
- [Usage Examples](#usage-examples)
- [Troubleshooting](#troubleshooting)

---

## Overview

Crush is a glamorous AI coding agent that lives in your terminal, seamlessly connecting your tools, code, and workflows with any Large Language Model (LLM) of your choice. Built by Charmbracelet (the creators of popular terminal tools like Gum, Glow, and VHS), Crush combines beautiful TUI aesthetics with powerful coding capabilities.

### Key Features

- **Multi-Model Flexibility**: Choose from various LLMs or add your own using OpenAI or Anthropic-compatible APIs
- **Session-Based Context**: Multiple project-specific contexts can coexist
- **LSP Enhancement**: Language Server Protocol integration for coding-aware context
- **MCP Extensibility**: Model Context Protocol plugins via HTTP, stdio, or SSE
- **Mid-Session Model Switching**: Change models while preserving context
- **Beautiful TUI**: Built with Charm's Bubble Tea framework for a polished terminal experience
- **Fast & Reliable**: Written in Go for exceptional speed and responsiveness
- **Skill System**: Supports Agent Skills open standard for extending capabilities

---

## Installation

### System Requirements

- **Operating System**: Windows 10+, macOS 10.15+, or Linux (Ubuntu 18.04+)
- **Memory**: Minimum 2GB RAM (4GB recommended)
- **Storage**: 100MB free space
- **Network**: Internet connection for AI model access
- **Optional**: Git for version control integration

### Installation Methods

#### Option 1: Homebrew (macOS/Linux - Recommended)

```bash
brew install charmbracelet/tap/crush
```

#### Option 2: NPM (Cross-Platform)

```bash
npm install -g @charmland/crush
```

#### Option 3: Go Install

```bash
go install github.com/charmbracelet/crush@latest
```

#### Option 4: Arch Linux

```bash
yay -S crush-bin
```

#### Option 5: Nix

```bash
nix run github:numtide/nix-ai-tools#crush
```

Or via NUR:
```bash
# Add the NUR channel
nix-channel --add https://github.com/nix-community/NUR/archive/main.tar.gz nur
nix-channel --update

# Get Crush in a Nix shell
nix-shell -p '(import <nur> { pkgs = import <nixpkgs> {}; }).repos.charmbracelet.crush'
```

#### Option 6: Windows (Winget)

```powershell
winget install charmbracelet.crush
```

#### Option 7: Windows (Scoop)

```powershell
scoop bucket add charm https://github.com/charmbracelet/scoop-bucket.git
scoop install crush
```

#### Option 8: FreeBSD

```bash
pkg install crush
```

#### Option 9: APT/YUM (Linux)

**Debian/Ubuntu:**
```bash
echo 'deb [trusted=yes] https://repo.charm.sh/apt/ /' | sudo tee /etc/apt/sources.list.d/charm.list
sudo apt update && sudo apt install crush
```

**RHEL/CentOS/Fedora:**
```bash
echo '[charm]
name=Charm
baseurl=https://repo.charm.sh/yum/
enabled=1
gpgcheck=0' | sudo tee /etc/yum.repos.d/charm.repo
sudo yum install crush
```

#### Option 10: Direct Binary Download

1. Visit the [Crush releases page](https://github.com/charmbracelet/crush/releases)
2. Download the appropriate binary for your system
3. Extract to your PATH directory
4. Make executable (Linux/macOS): `chmod +x crush`

---

## Quick Start

### 1. Verify Installation

```bash
crush --version
```

### 2. First Launch

```bash
crush
```

### 3. Initial Configuration

On first launch, Crush will guide you through:

1. **Select Provider**: Choose from available LLM providers
   - OpenAI
   - Anthropic (Claude)
   - OpenRouter (free models)
   - Vercel AI Gateway
   - Z.AI
   - Custom OpenAI-compatible API

2. **Select Model**: Pick your preferred model

3. **Enter API Key**: Provide your API key (stored securely in `~/.local/share/crush/crush.json`)

### 4. Start Coding

Once configured, you can immediately start interacting:

```
> Create a simple "Hello World" program in Python with proper documentation

> Refactor the main function to use async/await

> Explain what this code does
```

---

## CLI Commands

### Global Flags

| Flag | Description |
|------|-------------|
| `--version` | Show version information |
| `--help` | Display help message |
| `--model <name>` | Specify model to use |
| `--provider <name>` | Specify provider to use |
| `--config <path>` | Use custom configuration file |

### Non-Interactive Mode

Execute a single prompt and exit:

```bash
crush "Create a React component for a todo list"
```

### Project Initialization

```bash
# Navigate to project directory
cd /path/to/project

# Launch Crush in project context
crush

# When prompted, select "Yes" to initialize project
```

This creates:
- `crush.md` - Project context file
- LSP integration (if applicable)
- Session configuration

---

## TUI/Interactive Commands

### Key Bindings

| Shortcut | Action |
|----------|--------|
| `Ctrl+P` | Open command palette |
| `Ctrl+G` | Focus chat input |
| `Ctrl+S` | Session management |
| `Ctrl+F` | File attachment |
| `Ctrl+O` | Open editor |
| `Ctrl+C` | Exit/Cancel |
| `Tab` | Auto-completion |
| `↑` / `↓` | Navigate history |

### Command Palette Commands

Press `Ctrl+P` to open the command palette, then select:

#### Session Management

| Command | Description |
|---------|-------------|
| `New Session` | Create a new work session |
| `Switch Session` | Change to a different session |
| `Delete Session` | Remove a session |
| `Rename Session` | Rename current session |

#### Model Operations

| Command | Description |
|---------|-------------|
| `Switch Model` | Change LLM model mid-session |
| `Add Provider` | Configure a new AI provider |
| `Edit Provider` | Modify provider settings |

#### File Operations

| Command | Description |
|---------|-------------|
| `Attach File` | Include file in context (@filename) |
| `Open File` | Open file in external editor |
| `Refresh Context` | Reload project context |

#### MCP & Extensions

| Command | Description |
|---------|-------------|
| `Manage MCP Servers` | Add/remove MCP servers |
| `Install Skill` | Install Agent Skills |
| `Browse Skills` | View available skills |

### Session Management

```bash
# Create dedicated session
Ctrl+P → "New Session" → Name: "Backend API"

# Switch between sessions
Ctrl+S → Select session

# Create another session
Ctrl+P → "New Session" → Name: "Frontend Components"
```

### Model Switching

```bash
# Open command palette
Ctrl+P

# Select "Switch Model"
# Choose from available models:
# - OpenAI GPT-4o
# - Anthropic Claude 3.5 Sonnet
# - OpenRouter Qwen 3 Coder (free)
# - etc.
```

### Context Management

Mention files using `@` symbol:

```
> Review @src/app.ts for potential bugs

> Add tests for the functions in @src/utils.js

> Refactor @src/components/Header.tsx to use hooks
```

---

## Configuration

### Configuration File Location

```
~/.local/share/crush/crush.json
```

### Configuration Structure

```json
{
  "version": "1.0",
  "providers": {
    "openai": {
      "id": "openai",
      "name": "OpenAI",
      "api_key": "sk-...",
      "base_url": "https://api.openai.com/v1"
    },
    "anthropic": {
      "id": "anthropic",
      "name": "Anthropic",
      "api_key": "sk-ant-...",
      "base_url": "https://api.anthropic.com/v1"
    },
    "openrouter": {
      "id": "openrouter",
      "name": "OpenRouter",
      "api_key": "sk-or-...",
      "base_url": "https://openrouter.ai/api/v1"
    }
  },
  "default_provider": "openai",
  "default_model": "gpt-4o",
  "theme": "auto",
  "stream_responses": true,
  "show_progress": true,
  "confirm_commands": true
}
```

### MCP Server Configuration

Add MCP servers to extend capabilities:

```json
{
  "mcp": {
    "filesystem": {
      "type": "local",
      "command": ["npx", "-y", "@modelcontextprotocol/server-filesystem", "."]
    },
    "git": {
      "type": "local",
      "command": ["npx", "-y", "@modelcontextprotocol/server-git"]
    },
    "puppeteer": {
      "type": "local",
      "command": ["npx", "-y", "@modelcontextprotocol/server-puppeteer"]
    }
  }
}
```

### Project-Specific Configuration

Create `crush.md` in your project root:

```markdown
# Project Context

## Overview
This is a React-based e-commerce application built with TypeScript.

## Tech Stack
- React 18
- TypeScript
- Vite
- Tailwind CSS
- React Query

## Architecture
- `/src/components` - Reusable UI components
- `/src/pages` - Page components
- `/src/hooks` - Custom React hooks
- `/src/api` - API client and queries

## Coding Standards
- Use functional components with hooks
- Follow the existing naming conventions
- Add JSDoc for complex functions
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `OPENAI_API_KEY` | OpenAI API key |
| `ANTHROPIC_API_KEY` | Anthropic API key |
| `OPENROUTER_API_KEY` | OpenRouter API key |
| `CRUSH_CONFIG` | Custom config file path |
| `CRUSH_THEME` | Override theme setting |

---

## Usage Examples

### Basic Coding Tasks

```bash
# Start Crush
crush

# Inside Crush:
> Create a Python script that fetches weather data from an API

> Refactor this function to use async/await instead of callbacks

> Write unit tests for the UserService class
```

### Multi-Session Workflow

```bash
# Session 1: Backend Development
Ctrl+P → "New Session" → "Backend API"
> Create a REST API endpoint for user authentication

# Session 2: Frontend Development  
Ctrl+P → "New Session" → "Frontend UI"
> Create a login form component that uses the auth API

# Switch between sessions
Ctrl+S → Select session
```

### Model Comparison

```bash
# Try different models for the same task
Ctrl+P → "Switch Model" → "GPT-4o"
> Implement a binary search algorithm

Ctrl+P → "Switch Model" → "Claude 3.5 Sonnet"
> Implement a binary search algorithm

Ctrl+P → "Switch Model" → "Qwen 3 Coder"
> Implement a binary search algorithm
```

### Project Initialization Example

```bash
mkdir my-new-app
cd my-new-app
crush

# When prompted, initialize project

> Create a full-stack application with:
> - React frontend with TypeScript
> - Express backend with MongoDB
> - Docker Compose setup
> - README with setup instructions
```

### Using Skills

```bash
# Install a skill
Ctrl+P → "Install Skill" → "code-review"

# Use the skill
> /skill code-review @src/components/Button.tsx
```

### Working with Multiple Files

```bash
> Create a database schema in @src/models/user.ts
> Then create the corresponding API routes in @src/routes/users.ts
> And finally add tests in @tests/users.test.ts
```

### Git Integration

```bash
> Summarize the changes in the current git branch

> Create a commit message for the staged changes

> Review the diff and suggest improvements
```

---

## Troubleshooting

### Installation Issues

#### "brew install fails"

```bash
# Update Homebrew first
brew update
brew install charmbracelet/tap/crush
```

#### "npm install fails with permission error"

```bash
# Use npx instead
npx @charmland/crush

# Or fix npm permissions
sudo chown -R $(whoami) ~/.npm
```

#### "go install fails"

```bash
# Ensure Go 1.19+ is installed
go version

# Set GOPATH if needed
export PATH="$PATH:$(go env GOPATH)/bin"
```

### Configuration Issues

#### "API key not found"

```bash
# Set environment variable
export OPENAI_API_KEY="your-key-here"

# Or add to config file directly
crush --config
```

#### "Configuration file is invalid"

```bash
# Reset configuration
rm ~/.local/share/crush/crush.json

# Reconfigure
crush
```

### Runtime Issues

#### "Crush hangs or freezes"

```bash
# Kill the process
pkill -f crush

# Clear temporary files
rm -rf ~/.local/share/crush/sessions

# Restart
crush
```

#### "LSP not working"

1. Ensure LSP server is installed for your language
2. Check LSP configuration in project settings
3. Verify project has valid configuration files

#### "MCP server connection failed"

```bash
# Check MCP server status
# Review MCP configuration syntax
# Ensure required tools are installed (npx, etc.)
```

### Model Issues

#### "Model not available"

```bash
# List configured providers
Ctrl+P → "Switch Model"

# Add new provider
Ctrl+P → "Add Provider"
```

#### "API rate limit exceeded"

- Wait before making more requests
- Switch to a different provider
- Check your API usage dashboard

### Display Issues

#### "UI looks broken in terminal"

```bash
# Ensure your terminal supports Unicode
# Try a different terminal emulator
# Set TERM variable:
export TERM=xterm-256color
```

#### "Colors not rendering correctly"

```bash
# Force theme
export CRUSH_THEME=dark
# or
export CRUSH_THEME=light
```

### Getting Help

```bash
# Show help
crush --help

# Check version
crush --version

# Visit documentation
# https://github.com/charmbracelet/crush
```

### Reporting Issues

1. Check existing issues at https://github.com/charmbracelet/crush/issues
2. Include:
   - Crush version
   - Operating system
   - Terminal emulator
   - Configuration (redact API keys)
   - Steps to reproduce

---

## Resources

- **GitHub Repository**: https://github.com/charmbracelet/crush
- **Charm Website**: https://charm.sh
- **Documentation**: https://github.com/charmbracelet/crush#readme
- **Community**: https://github.com/charmbracelet/crush/discussions

---

*Last updated: 2026-04-02*
