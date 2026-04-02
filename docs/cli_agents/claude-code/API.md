# Claude Code - API Reference

## Command Line Interface

### Global Commands

```bash
claude [options] [directory]
```

### Options

| Option | Short | Description | Example |
|--------|-------|-------------|---------|
| `--help` | `-h` | Show help information | `claude --help` |
| `--version` | `-v` | Show version number | `claude --version` |
| `--resume` | `-r` | Resume previous session | `claude --resume` |
| `--continue` | `-c` | Continue with last message | `claude --continue` |
| `--print` | `-p` | Headless/non-interactive mode | `claude -p "command"` |
| `--bare` | `-b` | Minimal output mode | `claude --bare` |
| `--mcp-config` | | Specify MCP config file | `--mcp-config ./mcp.json` |
| `--model` | `-m` | Select model | `--model claude-3-5-sonnet` |

### Headless Mode

For CI/CD and scripting:

```bash
# Execute a single command
claude -p "explain the auth module"

# Execute with context
claude -p --resume "continue the task"

# With deferred tools (pauses for approval)
claude -p --resume --continue
```

---

## Slash Commands

### Built-in Commands

| Command | Description | Usage |
|---------|-------------|-------|
| `/help` | Show available commands | `/help` |
| `/clear` | Clear conversation history | `/clear` |
| `/reset` | Reset session (keep context) | `/reset` |
| `/exit` | Exit Claude Code | `/exit` or `Ctrl+C` |
| `/model` | Change active model | `/model` |
| `/config` | View/edit configuration | `/config` |
| `/usage` | View usage statistics | `/usage` |
| `/stats` | View session statistics | `/stats` |
| `/cost` | View estimated cost | `/cost` |
| `/resume` | List and resume sessions | `/resume` |
| `/bug` | Report a bug | `/bug description` |
| `/feedback` | Send feedback | `/feedback message` |
| `/compact` | Compact conversation | `/compact` |
| `/review` | Review pending changes | `/review` |
| `/undo` | Undo last action | `/undo` |
| `/permissions` | View permission settings | `/permissions` |
| `/env` | View/modify environment | `/env` |
| `/history` | View command history | `/history` |
| `/diff` | Show pending diffs | `/diff` |
| `/lint` | Run linter | `/lint` |
| `/test` | Run tests | `/test` |
| `/build` | Build project | `/build` |

### Git Commands

| Command | Description | Usage |
|---------|-------------|-------|
| `/git status` | Git status | `/git status` |
| `/git diff` | Show git diff | `/git diff` |
| `/git log` | Show git log | `/git log` |
| `/git commit` | Create commit | `/git commit message` |
| `/git push` | Push changes | `/git push` |
| `/git pr` | Create pull request | `/git pr` |

### Context Commands

| Command | Description | Usage |
|---------|-------------|-------|
| `@filename` | Reference file | `@src/main.py` |
| `@directory/` | Reference directory | `@src/` |
| `#symbol` | Reference symbol | `#MyClass` |
| `!command` | Execute bash | `!npm test` |

---

## Tool Reference

### Bash Tool

Execute shell commands with permission controls.

**Parameters:**
```json
{
  "command": "string",      // Command to execute
  "timeout": 60000,         // Timeout in milliseconds
  "description": "string"   // Human-readable description
}
```

**Examples:**
```bash
# Simple command
Bash({"command": "ls -la", "description": "List directory contents"})

# With timeout
Bash({"command": "npm test", "timeout": 120000})

# Piped commands
Bash({"command": "cat file.txt | grep pattern"})
```

### Read Tool

Read file contents.

**Parameters:**
```json
{
  "file_path": "string",    // Path to file
  "offset": 0,              // Start line (0-indexed)
  "limit": 100              // Number of lines to read
}
```

**Examples:**
```bash
# Read entire file
Read({"file_path": "README.md"})

# Read specific lines
Read({"file_path": "src/main.py", "offset": 10, "limit": 20})
```

### Write Tool

Create new files.

**Parameters:**
```json
{
  "file_path": "string",    // Path to create
  "content": "string"       // File content
}
```

### Edit Tool

Modify existing files using search/replace.

**Parameters:**
```json
{
  "file_path": "string",
  "old_string": "string",   // Text to find
  "new_string": "string"    // Replacement text
}
```

**Examples:**
```bash
# Simple replacement
Edit({
  "file_path": "config.js",
  "old_string": "const PORT = 3000;",
  "new_string": "const PORT = 8080;"
})

# Multi-line edit
Edit({
  "file_path": "app.js",
  "old_string": "function old() {\n  return 1;\n}",
  "new_string": "function new() {\n  return 2;\n}"
})
```

### Grep Tool

Search file contents.

**Parameters:**
```json
{
  "pattern": "string",      // Regex pattern
  "path": "string",         // Directory to search
  "include": "*.js",        // File pattern
  "exclude": "node_modules" // Exclude pattern
}
```

### Glob Tool

Find files by pattern.

**Parameters:**
```json
{
  "pattern": "string",      // Glob pattern
  "path": "string"          // Base directory
}
```

**Examples:**
```bash
# Find all JS files
Glob({"pattern": "**/*.js", "path": "."})

# Find test files
Glob({"pattern": "**/*.test.ts", "path": "src"})
```

### LS Tool

List directory contents.

**Parameters:**
```json
{
  "path": "string"          // Directory path
}
```

---

## Configuration Reference

### Settings Schema

```json
{
  // Model Configuration
  "model": "claude-3-5-sonnet-20241022",
  
  // Auto Mode Settings
  "autoMode": {
    "enabled": true,
    "allowEdits": true,
    "allowBash": true,
    "allowWrite": false
  },
  
  // Permission Settings
  "permissions": {
    "allow": [
      "Bash(npm test)",
      "Bash(git status)",
      "Read(**)"
    ],
    "deny": [
      "Bash(rm -rf *)",
      "Write(**/.env)"
    ]
  },
  
  // Behavior Settings
  "behavior": {
    "compactAfterTurns": 50,
    "showThinkingSummaries": false,
    "cleanupPeriodDays": 30
  },
  
  // Hooks Configuration
  "hooks": {
    "enabled": true,
    "directories": [".claude/hooks"]
  },
  
  // MCP Configuration
  "mcp": {
    "enabled": true,
    "configPath": "~/.claude/.mcp.json"
  },
  
  // Output Style
  "outputStyle": "normal",
  
  // Protected Directories
  "protectedDirectories": [
    "~/.ssh",
    "~/.aws",
    "~/.gnupg"
  ]
}
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `CLAUDE_API_KEY` | Anthropic API key | None |
| `ANTHROPIC_API_KEY` | Alternative API key | None |
| `CLAUDE_CODE_DEBUG` | Enable debug logging | `false` |
| `CLAUDE_CODE_NO_FLICKER` | Flicker-free rendering | `false` |
| `CLAUDE_CODE_THEME` | UI theme | `dark` |
| `CLAUDE_CODE_PLUGIN_KEEP_MARKETPLACE_ON_FAILURE` | Keep cache on pull failure | `false` |
| `MCP_CONNECTION_NONBLOCKING` | Non-blocking MCP | `false` |
| `HTTP_PROXY` / `HTTPS_PROXY` | Proxy settings | None |

### Model Options

| Model | Description | Context |
|-------|-------------|---------|
| `claude-3-7-sonnet-latest` | Latest Sonnet (3.7) | 200K |
| `claude-3-5-sonnet-latest` | Sonnet 3.5 | 200K |
| `claude-3-opus-latest` | Opus (most capable) | 200K |
| `claude-3-haiku-latest` | Haiku (fastest) | 200K |

---

## Plugin Configuration

### Plugin Structure

```json
// .claude-plugin/plugin.json
{
  "name": "plugin-name",
  "version": "1.0.0",
  "description": "Plugin description",
  "author": "Author Name",
  "commands": ["command-name"],
  "agents": ["agent-name"],
  "hooks": ["hook-name"],
  "skills": ["skill-name"],
  "mcpServers": ["server-name"]
}
```

### Command Definition

```markdown
---
allowed-tools: Bash(git:*), Bash(gh:*), Read(*), Edit(*)
description: Commit and create PR
---

## Context
- Current git status: !`git status`

## Your task
1. Stage changes
2. Create commit
3. Push to origin
4. Create PR
```

### Hook Definition

```json
// hooks/hooks.json
{
  "hooks": [
    {
      "event": "PreToolUse",
      "script": "./hooks/pretooluse.py",
      "if": "tool_name == 'Bash'"
    },
    {
      "event": "PostToolUse",
      "script": "./hooks/posttooluse.sh"
    }
  ]
}
```

---

## CLAUDE.md Format

Project context files that auto-load.

```markdown
# Project Context

## Overview
Brief project description

## Common Commands
- Build: `npm run build`
- Test: `npm test`
- Lint: `npm run lint`

## Architecture
Key architectural decisions

## Style Guidelines
- Use TypeScript
- Follow ESLint rules
- Prefer functional components

## Important Files
- `src/index.ts` - Entry point
- `src/types/` - Type definitions
```

---

## MCP Server Configuration

### Example .mcp.json

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@github/mcp-server"],
      "env": {
        "GITHUB_TOKEN": "${GITHUB_TOKEN}"
      }
    },
    "postgres": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-postgres"],
      "env": {
        "DATABASE_URL": "${DATABASE_URL}"
      }
    },
    "brave-search": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-brave-search"],
      "env": {
        "BRAVE_API_KEY": "${BRAVE_API_KEY}"
      }
    }
  }
}
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | General error |
| `2` | Invalid arguments |
| `3` | API error |
| `4` | Permission denied |
| `5` | Tool execution failed |
| `130` | Interrupted (Ctrl+C) |

---

## Related Documentation

- [Architecture](./ARCHITECTURE.md) - System design
- [Usage Guide](./USAGE.md) - Practical examples
- [Development Guide](./DEVELOPMENT.md) - Contributing
