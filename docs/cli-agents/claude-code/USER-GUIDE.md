# Claude Code User Guide

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

### Method 1: Native Install (Recommended)

**macOS, Linux, WSL:**
```bash
curl -fsSL https://claude.ai/install.sh | bash
```

**Windows PowerShell:**
```powershell
irm https://claude.ai/install.ps1 | iex
```

**Windows CMD:**
```cmd
curl -fsSL https://claude.ai/install.cmd -o install.cmd && install.cmd && del install.cmd
```

### Method 2: Package Managers

**Homebrew (macOS/Linux):**
```bash
brew install --cask claude-code
```

**WinGet (Windows):**
```powershell
winget install Anthropic.ClaudeCode
```

**NPM (Deprecated - use native install instead):**
```bash
npm install -g @anthropic-ai/claude-code
```

### Method 3: Build from Source

```bash
git clone https://github.com/anthropics/claude-code.git
cd claude-code
npm install
npm run build
npm link
```

### Prerequisites

- Node.js 18 or newer
- Git (Windows requires Git for Windows)
- Claude Pro, Max, or Anthropic Console account

## Quick Start

### First-Time Setup

```bash
# Verify installation
claude --version

# Start Claude Code
claude

# Authenticate (opens browser)
# Follow the prompts to log in with your Claude.ai account
```

### Basic Usage

```bash
# Start interactive session
claude

# Run with a single prompt (non-interactive)
claude -p "Explain this codebase"

# Start in a specific directory
claude --add-dir ../shared-lib

# Resume previous session
claude resume
```

### Hello World

```bash
# Start Claude Code
claude

# At the prompt, type:
> Create a hello world program in Python

# Claude will create the file and show you the results
```

## CLI Commands

### Global Options

| Option | Short | Description | Example |
|--------|-------|-------------|---------|
| --version | -v | Show version | `claude --version` |
| --help | -h | Show help | `claude --help` |
| --print | -p | Print response (non-interactive) | `claude -p "task"` |
| --add-dir | | Add extra working directories | `claude --add-dir ../lib` |
| --allowedTools | | Allow specific tools | `claude --allowedTools "Bash(git status)"` |
| --disallowedTools | | Disallow specific tools | `claude --disallowedTools "Bash(rm)"` |
| --model | | Set model for session | `claude --model claude-sonnet-4-20250514` |
| --permission-mode | | Set permission mode | `claude --permission-mode plan` |
| --verbose | | Enable verbose logging | `claude --verbose` |
| --max-turns | | Limit agent turns | `claude -p --max-turns 3 "task"` |
| --output-format | | Output format (text/json/stream-json) | `claude -p --output-format json "task"` |
| --append-system-prompt | | Append to system prompt | `claude --append-system-prompt "custom"` |

### Command: resume

**Description:** Resume a previous conversation session.

**Usage:**
```bash
claude resume [OPTIONS]
```

**Options:**
| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| --list | boolean | No | false | List available sessions |

**Examples:**
```bash
# Resume most recent session
claude resume

# List all sessions
claude resume --list
```

**Exit Codes:**
| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | No sessions found |

### Command: doctor

**Description:** Run diagnostics to check Claude Code installation and configuration.

**Usage:**
```bash
claude doctor
```

**Examples:**
```bash
# Run diagnostics
claude doctor
```

### Command: config

**Description:** View and manage Claude Code configuration.

**Usage:**
```bash
claude config [SUBCOMMAND]
```

**Subcommands:**
| Subcommand | Description |
|------------|-------------|
| get | Get a config value |
| set | Set a config value |
| list | List all config values |

**Examples:**
```bash
# List configuration
claude config list

# Set a config value
claude config set theme dark
```

### Command: rename

**Description:** Rename the current conversation.

**Usage:**
```bash
claude rename [NEW_NAME]
```

**Examples:**
```bash
# Rename current conversation
claude rename "Feature Implementation"
```

### Command: stats

**Description:** Show usage statistics.

**Usage:**
```bash
claude stats
```

### Command: usage

**Description:** Show current usage and billing information.

**Usage:**
```bash
claude usage
```

## TUI/Interactive Commands

When running in interactive/TUI mode, use these slash commands:

| Command | Shortcut | Description | Example |
|---------|----------|-------------|---------|
| /help | ? | Show available commands | `/help` |
| /exit | Ctrl+D | Exit Claude Code | `/exit` |
| /clear | Ctrl+L | Clear screen | `/clear` |
| /add | | Add files to context | `/add src/main.js` |
| /add-dir | | Add directory to context | `/add-dir ./lib` |
| /drop | | Remove files from context | `/drop src/main.js` |
| /commit | | Create git commit | `/commit "Fix login bug"` |
| /diff | | Show git diff | `/diff` |
| /log | | Show conversation history | `/log` |
| /undo | | Undo last action | `/undo` |
| /model | | Switch model | `/model claude-sonnet-4` |
| /mode | | Change permission mode | `/mode auto` |
| /cost | | Show token usage and cost | `/cost` |
| /tokens | | Show token count | `/tokens` |
| /compact | | Summarize conversation | `/compact` |
| /mcp | | Manage MCP servers | `/mcp list` |
| /statusline | | Toggle status line | `/statusline` |
| /terminal-setup | | Configure terminal | `/terminal-setup` |
| /teleport | | Navigate to directory | `/teleport ~/projects` |
| /remote-env | | Configure remote environment | `/remote-env` |
| /rewind | | Go back to previous state | `/rewind 5` |
| /stats | | Show session statistics | `/stats` |
| /usage | | Show usage info | `/usage` |
| /config | | Open config UI | `/config` |
| /doctor | | Run diagnostics | `/doctor` |
| /resume | | Resume conversation | `/resume` |
| /rename | | Rename conversation | `/rename "New Name"` |

### Permission Modes

| Mode | Description |
|------|-------------|
| auto | Full automatic mode (approve safe actions) |
| plan | Plan mode (approve each step) |
| ask | Ask before every action |

### Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| Tab | Accept completion |
| Ctrl+C | Cancel current operation |
| Ctrl+D | Exit Claude Code |
| Ctrl+L | Clear screen |
| Ctrl+B | Background tasks |
| Alt+T | Toggle thinking mode |
| Tab (in thinking) | Toggle ultrathink |

## Configuration

### Configuration File Format

Claude Code uses `.claude/settings.json` for configuration:

```json
{
  "env": {
    "CLAUDE_CODE_USE_BEDROCK": "1",
    "AWS_REGION": "us-east-1",
    "AWS_PROFILE": "your-profile"
  },
  "allowedTools": [
    "Bash(git status:*)",
    "Read",
    "Edit"
  ],
  "disallowedTools": [
    "Bash(rm -rf *)"
  ],
  "theme": "dark",
  "language": "en",
  "attribution": true,
  "respectGitignore": true,
  "plansDirectory": ".claude/plans"
}
```

### Project-Specific Configuration (CLAUDE.md)

Create a `CLAUDE.md` file in your project root:

```markdown
# Project Guidelines

## Coding Standards
- Use TypeScript for all new code
- Follow ESLint rules
- Write tests for all features

## Architecture
- Frontend: React + Vite
- Backend: Node.js + Express
- Database: PostgreSQL

## Common Commands
- `npm run dev` - Start development server
- `npm test` - Run tests
- `npm run build` - Build for production
```

### Environment Variables

| Variable | Description | Required | Example |
|----------|-------------|----------|---------|
| ANTHROPIC_API_KEY | Anthropic API key | Yes* | `sk-ant-...` |
| CLAUDE_CODE_USE_BEDROCK | Use AWS Bedrock | No | `1` |
| AWS_REGION | AWS region | No | `us-east-1` |
| AWS_PROFILE | AWS profile | No | `default` |
| CLAUDE_SESSION_ID | Session identifier | No | `abc123` |

*Required if not using Claude.ai account

### Configuration Locations (in order of precedence)

1. Command-line flags
2. Environment variables
3. Project config (`./.claude/settings.json`)
4. User config (`~/.claude/settings.json`)
5. System config (`/etc/claude/settings.json`)
6. Project CLAUDE.md (`./CLAUDE.md`)
7. User CLAUDE.md (`~/.claude/CLAUDE.md`)

## API/Protocol Endpoints

Claude Code primarily uses the Anthropic API. For MCP (Model Context Protocol) integrations:

### MCP Server Configuration

Create `~/.claude/mcp.json`:

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "ghp_..."
      }
    },
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/home/user/docs"]
    }
  }
}
```

### Using MCP in Claude Code

```bash
# List MCP servers
/mcp list

# Enable/disable MCP servers
/mcp enable github
/mcp disable filesystem
```

## Usage Examples

### Example 1: Code Analysis and Refactoring

```bash
# Start in your project directory
claude

# Ask Claude to analyze the codebase
> Analyze this codebase and identify the main components

# Request refactoring
> Refactor the authentication module to use JWT tokens

# Review the changes
> Show me the diff of all changes made
```

### Example 2: Feature Implementation

```bash
# Start with a specific task
claude -p "Create a REST API endpoint for user registration with validation"

# Or interactively
claude
> Create a new React component for a login form with email and password fields
> Add form validation using Formik
> Style it with Tailwind CSS
```

### Example 3: Git Workflow Automation

```bash
# Start Claude Code
claude

# Check status
> What files have I changed?

# Create commits
> Commit all changes with message "Add user authentication"

# Create branch
> Create a new branch for the feature payment-integration

# Generate PR description
> Generate a pull request description for the current branch
```

### Example 4: Advanced Usage with Multiple Directories

```bash
# Work with multiple directories
claude --add-dir ../shared-lib --add-dir ../common-utils

# Or in TUI
/add-dir ../shared-lib

# Request cross-module changes
> Update the API client in src/api to use the shared utilities from shared-lib
```

### Example 5: Using with MCP Servers

```bash
# Configure MCP first, then:
claude

# Use GitHub MCP
> Create an issue for the bug we just found

# Use filesystem MCP
> Read the documentation from /home/user/docs/api.md
```

## Troubleshooting

### Issue: Authentication Fails

**Symptoms:** Claude Code prompts for login repeatedly or shows "Authentication failed"

**Solution:**
```bash
# Log out and log back in
claude logout
claude login

# Or use API key instead
echo "sk-ant-..." > ~/.claude/api_key
```

### Issue: Permission Denied Errors

**Symptoms:** Claude Code cannot edit files or run commands

**Solution:**
```bash
# Check file permissions
ls -la

# Run with appropriate permissions
# On macOS/Linux, ensure you own the files
sudo chown -R $(whoami) .

# In Claude Code, adjust permission mode
/mode auto
```

### Issue: High Token Usage / Costs

**Symptoms:** Sessions becoming expensive quickly

**Solution:**
```bash
# Use compact regularly
/compact

# Drop unnecessary files from context
/drop large-file.js

# Start fresh with a new task
/newtask

# Use --max-turns for limited interactions
claude -p --max-turns 5 "quick task"
```

### Issue: Claude Code Freezes on Windows

**Symptoms:** Terminal becomes unresponsive

**Solution:**
- Use Windows Terminal or PowerShell instead of CMD
- Ensure Git for Windows is installed
- Run in WSL for better compatibility
- Update to latest version: `npm upgrade -g @anthropic-ai/claude-code`

### Issue: MCP Servers Not Working

**Symptoms:** MCP commands fail or timeout

**Solution:**
```bash
# Check MCP configuration
/doctor

# Verify MCP server is installed
which npx

# Check MCP logs
# In settings.json, enable verbose logging:
{
  "verbose": true
}
```

### Issue: Model Not Available

**Symptoms:** "Model not found" or "Invalid model" errors

**Solution:**
```bash
# List available models
/model

# Use a supported model
claude --model claude-sonnet-4-20250514
```

### Issue: Session Resumption Fails

**Symptoms:** `claude resume` doesn't restore previous context

**Solution:**
```bash
# List available sessions
claude resume --list

# Resume specific session
claude resume <session-id>

# Clear corrupted sessions
rm -rf ~/.claude/sessions/
```

---

**Last Updated:** 2026-04-02
**Version:** 2.1.x
