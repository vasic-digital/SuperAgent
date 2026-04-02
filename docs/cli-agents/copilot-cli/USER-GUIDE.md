# GitHub Copilot CLI - User Guide

> AI-powered terminal-native coding assistant that brings agentic capabilities directly to your command line.

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

GitHub Copilot CLI is a powerful terminal-native AI coding assistant that enables developers to interact with AI directly from the command line. It offers deep GitHub workflow integration, autonomous task execution, and the ability to work on complex tasks while maintaining full user control.

### Key Features

- **Natural Language Interface**: Describe what you want in plain English
- **GitHub Native Integration**: Works seamlessly with issues, PRs, and repositories
- **Parallel Execution**: Use `/fleet` to run multiple subagents simultaneously
- **MCP Support**: Extend capabilities with Model Context Protocol servers
- **Session Management**: Resume long-running work with memory and compaction
- **Multi-Model Support**: Switch between models with `/model` command
- **Custom Skills**: Define specialized behaviors with AGENTS.md and skills

---

## Installation

### Prerequisites

- **GitHub Copilot subscription** (Free, Pro, Pro+, Business, or Enterprise)
- **Node.js 22+** and **npm 10+** (for npm installation)
- **PowerShell 6+** (Windows only)

### Installation Methods

#### Option 1: NPM (Cross-Platform)

```bash
npm install -g @github/copilot
```

> **Note**: If you have `ignore-scripts=true` in your `~/.npmrc`:
> ```bash
> npm_config_ignore_scripts=false npm install -g @github/copilot
> ```

#### Option 2: Homebrew (macOS/Linux)

```bash
brew install copilot-cli
```

For prerelease version:
```bash
brew install copilot-cli@prerelease
```

#### Option 3: WinGet (Windows)

```powershell
winget install GitHub.Copilot
```

For prerelease version:
```powershell
winget install GitHub.Copilot.Prerelease
```

#### Option 4: Install Script (macOS/Linux)

```bash
curl -fsSL https://gh.io/copilot-install | bash
```

Or with wget:
```bash
wget -qO- https://gh.io/copilot-install | bash
```

**Custom installation options:**
```bash
# Install as root to /usr/local/bin
curl -fsSL https://gh.io/copilot-install | sudo bash

# Install to custom directory
curl -fsSL https://gh.io/copilot-install | PREFIX="$HOME/custom" bash

# Install specific version
curl -fsSL https://gh.io/copilot-install | VERSION="v0.0.369" bash
```

#### Option 5: Direct Download

Download executables directly from the [copilot-cli releases page](https://github.com/github/copilot-cli/releases).

---

## Quick Start

### 1. Authenticate

Start the CLI and login:

```bash
copilot
```

Then type:
```
/login
```

Follow the on-screen prompts to authenticate with your GitHub account. You'll only need to do this once.

### 2. Verify Installation

```bash
copilot --version
```

### 3. First Interaction

Navigate to a project directory and start Copilot:

```bash
cd /path/to/your/project
copilot
```

Trust the directory when prompted, then try:

```
Give me an overview of this project.
```

---

## CLI Commands

### Global Flags

| Flag | Description |
|------|-------------|
| `--version` | Show version information |
| `--help` | Display help message |
| `--prompt "<text>"` | Execute a single prompt and exit |
| `--allow-all` | Enable all permissions without prompting |
| `--yolo` | Alias for `--allow-all` |

### Non-Interactive Mode

Execute a single command without entering interactive mode:

```bash
# Execute a prompt directly
copilot "Explain the main function in src/app.js"

# Equivalent syntax
copilot --prompt "Create a React component for a login form"
```

### Authentication Commands

```bash
# Login to GitHub
copilot /login

# Check authentication status
copilot /whoami
```

### Session Management

```bash
# Start interactive session
copilot

# Resume previous session
copilot /resume

# List available sessions to resume
copilot /resume list
```

---

## Interactive Commands

Once inside the Copilot CLI interactive session, use these commands:

### Core Shortcuts

| Shortcut | Action |
|----------|--------|
| `Esc` | Cancel current operation |
| `Ctrl+C` | Cancel if thinking, clear input, or exit |
| `Ctrl+L` | Clear the screen |
| `↑` / `↓` | Navigate command history |
| `Tab` | Auto-complete |
| `@` | Mention files to include in context |
| `/` | Show slash commands |
| `?` | Show tabbed help |

### Slash Commands

#### Navigation & Help

| Command | Description |
|---------|-------------|
| `/help` | Show all available commands |
| `/exit` | Exit the CLI |
| `/clear` | Clear the conversation history |

#### Model Management

| Command | Description |
|---------|-------------|
| `/model` | List available models |
| `/model <name>` | Switch to a specific model |

#### Context Management

| Command | Description |
|---------|-------------|
| `/context` | Show current context overview |
| `/usage` | Display session statistics (requests, duration, LOC edited) |
| `/compact` | Manually compress conversation history |

#### Planning & Execution

| Command | Description |
|---------|-------------|
| `/plan` | Create a structured plan for a task |
| `/fleet` | Execute with parallelized subagents |
| `/delegate` | Delegate work to a subagent |
| `/diff` | Show changes made |

#### GitHub Integration

| Command | Description |
|---------|-------------|
| `/mcp` | Interact with GitHub MCP server |
| `/mcp add` | Add a new MCP server |

#### Agent Customization

| Command | Description |
|---------|-------------|
| `/agent` | Configure custom agent behavior |
| `/skills` | List available skills |

#### Experimental Features

| Command | Description |
|---------|-------------|
| `/experimental show` | Access preview features |
| `/changelog` | View latest updates |

---

## Configuration

### Configuration File Location

```
~/.copilot/config.json
```

Set custom location with:
```bash
export COPILOT_HOME=/path/to/custom/copilot
```

### MCP Configuration

MCP servers are configured in:
```
~/.copilot/mcp-config.json
```

#### Adding an MCP Server

Interactive method:
```
/mcp add
```

Manual configuration example:
```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@github/mcp-server"]
    },
    "custom-server": {
      "command": "node",
      "args": ["/path/to/server.js"],
      "env": {
        "API_KEY": "your-api-key"
      }
    }
  }
}
```

### Custom Instructions

Create `.github/copilot-instructions.md` in your repository:

```markdown
# Copilot Instructions

## Coding Standards
- Always use TypeScript for new files
- Follow the existing naming conventions
- Add JSDoc comments for public APIs

## Architecture Preferences
- Prefer functional components over class components
- Use React hooks for state management
- Implement proper error handling

## Testing Requirements
- Write unit tests for all new functions
- Maintain minimum 80% code coverage
```

### Custom Agents

Create custom agents in `.github/agents/`:

```markdown
---
name: security-auditor
description: Security-focused code reviewer
model: claude-sonnet-4
tools: ["bash", "view", "edit"]
---

You are a security-focused code reviewer. Analyze all code for:
- SQL injection vulnerabilities
- XSS vulnerabilities
- Insecure authentication patterns
- Hardcoded secrets

Provide detailed explanations of any issues found.
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `COPILOT_GITHUB_TOKEN` | Personal access token for authentication |
| `GH_TOKEN` | Alternative token (checked second) |
| `GITHUB_TOKEN` | Alternative token (checked third) |
| `COPILOT_HOME` | Custom configuration directory |

---

## Usage Examples

### Basic Code Tasks

```bash
# Start interactive session
copilot

# Inside the session:
> Create a Python function to parse JSON with error handling

> Refactor the authentication middleware to use JWT

> Explain how the caching layer works in this codebase
```

### Working with Files

```bash
# Mention files using @ symbol
> Review @src/app.js and @src/utils.js for bugs

> Add TypeScript types to @src/models/user.ts

> Create tests for the functions in @src/helpers.js
```

### GitHub Workflow Integration

```bash
# Work with issues
> Find all open issues labeled "bug" and summarize them

> Create a branch for issue #123 and implement the fix

# Pull request operations
> Review the open pull requests and suggest which to merge

> Create a PR for the current branch with a detailed description
```

### Parallel Execution with /fleet

```bash
# Execute multiple tasks in parallel
/fleet Implement these features:
1. Add user authentication (depends on: none)
2. Create database schema for users (depends on: none)
3. Build login UI components (depends on: 1)
4. Write API tests (depends on: 1, 2)
```

### Using Custom Skills

```bash
# Apply a skill
> Use @security-auditor to review the auth module

> Apply @docs-writer to document the API endpoints
```

### Azure/DevOps Examples

```bash
# Azure CLI tasks
> Create an Azure App Service with Linux runtime and Node.js

> Update the Application Gateway backend pool

# Terraform
> Create a Terraform module for Azure VNet with 3 subnets

> Fix issues in @main.tf

# Pipeline debugging
> Analyze this Azure DevOps pipeline YAML and fix errors
```

---

## Troubleshooting

### Installation Issues

#### "npm install fails with permission error"

```bash
# Use npx instead
npx @github/copilot

# Or fix npm permissions
sudo chown -R $(whoami) ~/.npm
```

#### "command not found: copilot"

```bash
# Ensure npm global bin is in PATH
export PATH="$PATH:$(npm bin -g)"

# Or use npx
npx @github/copilot
```

### Authentication Issues

#### "Authentication failed"

1. Check your Copilot subscription is active
2. Verify organization hasn't disabled CLI access
3. Try re-authenticating:
   ```
   /login
   ```

#### "Token expired"

```bash
# Re-authenticate
copilot /login
```

### Runtime Issues

#### "Copilot CLI not responding"

```bash
# Kill any hanging processes
pkill -f copilot

# Clear session data
rm -rf ~/.copilot/sessions

# Restart
copilot
```

#### "Context limit reached"

```bash
# Compact conversation manually
/compact

# Or start a new session
/exit
copilot
```

#### "Model not available"

```bash
# List available models
/model

# Switch to available model
/model gpt-4
```

### Permission Issues

#### "Permission denied" errors

```bash
# Run with --allow-all for trusted projects only
copilot --allow-all

# Or in interactive mode, approve each action
```

### Network Issues

#### "Cannot connect to GitHub"

1. Check internet connection
2. Verify GitHub status at https://www.githubstatus.com
3. Check proxy settings if behind corporate firewall

### Getting Help

```bash
# Show help
/help

# Show detailed help for a command
/help /fleet

# View changelog
/changelog
```

### Reporting Issues

1. Check existing issues at https://github.com/github/copilot-cli/issues
2. Include:
   - Copilot CLI version (`copilot --version`)
   - Operating system
   - Node.js version (`node --version`)
   - Steps to reproduce
   - Error messages

---

## Resources

- **Documentation**: https://docs.github.com/copilot
- **GitHub Repository**: https://github.com/github/copilot-cli
- **Release Notes**: https://github.com/github/copilot-cli/releases
- **Community Discussions**: https://github.com/orgs/community/discussions

---

*Last updated: 2026-04-02*
