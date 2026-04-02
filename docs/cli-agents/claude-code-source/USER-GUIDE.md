# Claude Code Source - User Guide

**Claude Code Source** is the open-source distribution of Anthropic's Claude Code CLI, providing a terminal-based AI coding assistant with full source code access for customization, auditing, and contribution.

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

- **Operating System**: macOS 10.15+, Ubuntu 20.04+, Windows 10+
- **Node.js**: Version 18.0 or higher
- **Git**: Required for version control integration
- **Account**: Claude Pro / Max / Teams / Enterprise (free plan not supported)

### Method 1: Native Installation (Recommended)

**macOS/Linux:**
```bash
curl -fsSL https://claude.ai/install.sh | bash
```

**Windows (PowerShell):**
```powershell
irm https://claude.ai/install.ps1 | iex
```

### Method 2: npm Installation

```bash
npm install -g @anthropic-ai/claude-code
```

### Method 3: WinGet (Windows)

```powershell
winget install Anthropic.ClaudeCode
```

### Method 4: From Source

```bash
# Clone the repository
git clone https://github.com/anthropics/claude-code.git
cd claude-code

# Install dependencies
npm install

# Build
npm run build

# Link for global access
npm link
```

### Verify Installation

```bash
claude --version
```

### Post-Installation PATH Setup

**If `claude` command not found:**

**macOS/Linux:**
```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

**Windows:**
```powershell
[Environment]::SetEnvironmentVariable("PATH", "$env:PATH;$env:USERPROFILE\.local\bin", [EnvironmentVariableTarget]::User)
```

---

## Quick Start

### First Launch

```bash
# Navigate to your project
cd ~/projects/my-app

# Start Claude Code
claude
```

On first launch:
1. A browser window opens automatically
2. Sign in with your Claude.ai account
3. Return to terminal - authentication is complete

### Your First Command

```bash
# Explore the codebase
> Read the src/ directory and explain the project architecture

# Make a change
> Add input validation to the POST /api/users endpoint
```

### Resume Session

```bash
# Continue previous session
/resume
```

### Exit

```bash
# Exit Claude Code
/exit
# or press Ctrl+C
```

---

## CLI Commands

### Core Commands

| Command | Description |
|---------|-------------|
| `claude` | Start interactive session |
| `claude --version` | Show version information |
| `claude --help` | Show help |
| `claude login` | Authenticate with Claude.ai |
| `claude logout` | Sign out |

### Non-Interactive Mode

```bash
# Single command execution
echo "Explain this codebase" | claude

# Process file
claude < instructions.md

# With timeout
claude --timeout 30s "Run tests"
```

### Session Management

| Command | Description |
|---------|-------------|
| `/resume` | Resume previous session |
| `/clear` | Clear conversation history |
| `/compact` | Compact context window |
| `/exit` or `/quit` | End session |

### Slash Commands (In-Session)

| Command | Description |
|---------|-------------|
| `/help` | Show available commands |
| `/clear` | Clear conversation |
| `/compact` | Compact context |
| `/review` | Review recent changes |
| `/diff` | Show git diff |
| `/commit` | Generate commit message |
| `/pr` | Generate PR description |
| `/test` | Run tests |

---

## TUI/Interactive Commands

### Interactive Session Controls

| Key | Action |
|-----|--------|
| `↑/↓` | Navigate command history |
| `Tab` | Autocomplete |
| `Ctrl+C` | Cancel current operation |
| `Ctrl+D` | Exit (same as `/exit`) |
| `Ctrl+G` | Open prompt in editor |
| `@` | Reference files/directories |

### File References

Use `@` to reference files and directories:

```
> Review @src/components/Button.tsx for accessibility issues
> Explain how @utils/auth.ts works
> Compare @v1/api.ts with @v2/api.ts
```

### Context Management

```bash
# Clear context when it gets too large
/clear

# Compact to save tokens
/compact

# View context usage
/context
```

---

## Configuration

### Global Configuration File

Configuration stored in `~/.claude/settings.json`:

```json
{
  "env": {
    "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1"
  },
  "acceptEdits": "auto",
  "statusLine": true,
  "teammateMode": "auto",
  "hooks": {
    "pre-edit": [],
    "post-edit": []
  }
}
```

### Project Configuration (CLAUDE.md)

Create `CLAUDE.md` in project root:

```markdown
# Project: MyApp

## Coding Standards
- Use TypeScript strict mode
- Follow ESLint configuration
- Write tests for all new features

## Architecture
- Next.js App Router
- Prisma ORM
- Tailwind CSS

## Commands
- `npm run dev` - Start development server
- `npm test` - Run tests
- `npm run lint` - Run linter
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `ANTHROPIC_API_KEY` | API key for authentication |
| `CLAUDE_CONFIG_DIR` | Custom config directory |
| `CLAUDE_CODE_DEBUG` | Enable debug logging |
| `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` | Enable agent teams |

### Hierarchical CLAUDE.md

Claude Code supports multi-level context files:

```
~/.claude/CLAUDE.md          # Global preferences
~/projects/CLAUDE.md         # Organization standards
~/projects/my-app/CLAUDE.md  # Project-specific
~/projects/my-app/src/CLAUDE.md  # Module-specific
```

### Agent Teams Configuration

Enable multi-agent teams in `settings.json`:

```json
{
  "env": {
    "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1"
  },
  "teammateMode": "auto"
}
```

Modes:
- `"auto"` - Use split panes if in tmux, in-process otherwise
- `"in-process"` - All teammates in main terminal
- `"tmux"` - Each teammate in separate tmux pane

---

## Usage Examples

### Code Exploration

```bash
# Understand project structure
> Read the src/ directory and explain the architecture

# Find specific code
> Find where authentication is handled

# Explain complex code
> Explain the logic in @utils/algorithm.ts
```

### Feature Development

```bash
# Implement a feature
> Add a user profile page with edit functionality

# Follow existing patterns
> Check how other forms handle validation and do the same for the login form

# Add tests
> Write unit tests for the new auth module
```

### Debugging

```bash
# Debug failing tests
> The tests in @tests/api.test.ts are failing. Debug and fix.

# Analyze errors
> I got this error: [paste error]. What's causing it?

# Check logs
> Look at the server logs and identify any issues
```

### Refactoring

```bash
# Refactor with plan
> Create a plan to refactor the auth module to use services

# Execute refactor
> Refactor @components/Button/ to use compound component pattern

# Update imports
> Update all imports after the file move
```

### Git Integration

```bash
# Review changes
> /diff

# Commit changes
> /commit

# Generate PR description
> /pr
```

### Agent Teams Workflow

```bash
# Enable agent teams (one-time setup)
# Add to ~/.claude/settings.json:
# { "env": { "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1" } }

# Create a team
> Create an agent team with 3 teammates to explore this feature from different angles

# Manage team
# Shift+Down - Cycle through teammates
# Type message - Send to selected teammate
```

### Sub-agents

```bash
# Spawn sub-agent for parallel work
> Task: Research best practices for React forms

# Use results
> Based on that research, implement the form component
```

---

## Troubleshooting

### Installation Issues

#### "claude is not recognized"

```bash
# Check if in PATH
which claude

# Add to PATH (macOS/Linux)
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# Windows
# Add %USERPROFILE%\.local\bin to PATH via System Settings
```

#### Node.js version too old

```bash
# Check version
node --version

# Update via nvm
nvm install 18
nvm use 18

# Or download from nodejs.org
```

### Authentication Issues

#### "Error: Not authenticated"

```bash
# Re-authenticate
claude login

# Or use API key
export ANTHROPIC_API_KEY=sk-ant-...
```

#### "Authentication failed"

1. Check internet connection
2. Verify Claude.ai account is active
3. Try logging out and back in: `claude logout && claude login`

### Session Issues

#### Context window full

```bash
# Compact context
/compact

# Or clear and start fresh
/clear
```

#### Claude seems confused

```bash
# Clear context
/clear

# Re-explain with more context
> Let me provide more context about this project...
```

### Permission Issues

#### File access denied

```bash
# Check file permissions
ls -la file.txt

# Fix permissions
chmod 644 file.txt
```

#### Command execution blocked

- Claude will ask for permission before running commands
- Press `y` to approve, `n` to deny
- Use `--dangerously-skip-permissions` for automation (not recommended)

### Performance Issues

#### Slow responses

1. Check internet connection
2. Compact context: `/compact`
3. Clear and restart: `/clear`
4. Use smaller context window

#### High token usage

```bash
# Monitor usage in status line
# Enable in settings.json:
# { "statusLine": true }

# Use /compact regularly
/compact
```

### Common Errors

#### "Command not found" in Claude

Claude may not have access to certain commands. Use full paths:

```bash
> Run /usr/local/bin/custom-tool
```

#### Git operations failing

```bash
# Check git status
> Run git status

# Configure git user if needed
> Run git config user.email "you@example.com"
> Run git config user.name "Your Name"
```

### Debug Mode

```bash
# Enable debug logging
export CLAUDE_CODE_DEBUG=1
claude

# Check logs
# Logs stored in ~/.claude/logs/
```

### Getting Help

```bash
# In-session help
/help

# Command help
claude --help

# Anthropic documentation
# https://docs.anthropic.com/claude-code
```

---

## Best Practices

1. **Start with CLAUDE.md**: Create project documentation for better results
2. **Use @ References**: Reference specific files for precise context
3. **Compact Regularly**: Use `/compact` to manage token usage
4. **Review Changes**: Always review before committing
5. **Use Agent Teams**: For complex tasks, leverage multiple agents
6. **Git Hygiene**: Commit regularly using `/commit`
7. **Test Changes**: Run tests after modifications
8. **Be Specific**: Clear, specific prompts yield better results

---

## Advanced Features

### Hooks System

Configure pre/post execution hooks in `settings.json`:

```json
{
  "hooks": {
    "pre-edit": ["npm run lint:check"],
    "post-edit": ["npm run format", "npm run test:related"]
  }
}
```

### MCP Integration

Configure MCP servers in `~/.claude/.mcp.json`:

```json
{
  "mcpServers": {
    "fetch": {
      "command": "uvx",
      "args": ["mcp-server-fetch"]
    }
  }
}
```

### Custom Slash Commands

Create custom commands in `~/.claude/commands/`:

```markdown
# ~/.claude/commands/deploy.md
---
description: Deploy to production
---

Run tests, build, and deploy to production.
```

---

*Last Updated: April 2026*
