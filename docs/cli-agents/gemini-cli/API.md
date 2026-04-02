# Gemini CLI - API Reference

## Command Line Interface

### Global Commands

```bash
gemini [options] [query]
```

### CLI Options

| Option | Short | Type | Default | Description |
|--------|-------|------|---------|-------------|
| `--help` | `-h` | - | - | Show help information |
| `--version` | `-v` | - | - | Show version number |
| `--debug` | `-d` | boolean | `false` | Run with verbose logging |
| `--model` | `-m` | string | `auto` | Model to use (see Model Selection) |
| `--prompt` | `-p` | string | - | Non-interactive prompt (deprecated, use positional) |
| `--prompt-interactive` | `-i` | string | - | Execute and continue interactively |
| `--resume` | `-r` | string | - | Resume session (use "latest" or session ID) |
| `--sandbox` | `-s` | boolean | `false` | Run in sandboxed environment |
| `--approval-mode` | - | string | `default` | Approval mode: default, auto_edit, plan, yolo |
| `--yolo` | `-y` | boolean | `false` | Auto-approve all actions (deprecated) |
| `--include-directories` | - | array | - | Additional directories to include |
| `--output-format` | `-o` | string | `text` | Output format: text, json, stream-json |
| `--screen-reader` | - | boolean | - | Enable screen reader mode |
| `--list-sessions` | - | boolean | - | List available sessions |
| `--delete-session` | - | string | - | Delete session by index |
| `--allowed-mcp-server-names` | - | array | - | Allowed MCP servers |
| `--extensions` | `-e` | array | - | List of extensions to use |
| `--list-extensions` | `-l` | boolean | - | List available extensions |

### Positional Arguments

| Argument | Type | Description |
|----------|------|-------------|
| `query` | string (variadic) | Prompt for non-interactive mode |

### Non-Interactive Mode Examples

```bash
# Simple query
gemini "explain this codebase"

# Query with specific model
gemini -m gemini-2.5-flash "explain this codebase"

# JSON output for scripting
gemini -p "explain this codebase" --output-format json

# Streaming JSON for real-time monitoring
gemini -p "run tests" --output-format stream-json

# Execute and continue interactively
gemini -i "What is the purpose of this project?"

# Resume and continue
gemini -r "latest" "Check for type errors"
```

---

## Slash Commands

### Built-in Commands

| Command | Description | Usage |
|---------|-------------|-------|
| `/help` | Show available commands | `/help` |
| `/clear` | Clear conversation history | `/clear` |
| `/reset` | Reset session context | `/reset` |
| `/exit` | Exit Gemini CLI | `/exit` or `Ctrl+C` |
| `/model` | Change active model | `/model` or `/model gemini-2.5-pro` |
| `/chat` | Start new conversation | `/chat` |
| `/plan` | Enter plan mode | `/plan` |
| `/compress` | Compress conversation | `/compress` |
| `/undo` | Undo last action | `/undo` |
| `/redo` | Redo last undone action | `/redo` |
| `/stats` | View session statistics | `/stats` |
| `/tools` | List available tools | `/tools` |
| `/memory` | Manage GEMINI.md context | `/memory show`, `/memory refresh`, `/memory add <text>` |
| `/bug` | Report a bug | `/bug description` |

### Git Commands

| Command | Description | Usage |
|---------|-------------|-------|
| `/git status` | Git status | `/git status` |
| `/git diff` | Show git diff | `/git diff` |
| `/git log` | Show git log | `/git log` |
| `/git commit` | Create commit | `/git commit message` |

### Context References

| Syntax | Description | Example |
|--------|-------------|---------|
| `@filename` | Reference file | `@src/main.ts` |
| `@directory/` | Reference directory | `@src/` |
| `!command` | Execute shell | `!npm test` |

---

## Model Selection

### Model Aliases

| Alias | Resolves To | Description |
|-------|-------------|-------------|
| `auto` | gemini-2.5-pro or gemini-3-pro-preview | Default, uses preview if enabled |
| `pro` | gemini-2.5-pro or gemini-3-pro-preview | Complex reasoning tasks |
| `flash` | gemini-2.5-flash | Fast, balanced for most tasks |
| `flash-lite` | gemini-2.5-flash-lite | Fastest for simple tasks |

### Concrete Model Names

| Model | Context Window | Best For |
|-------|----------------|----------|
| `gemini-3-pro-preview` | 1M tokens | Complex reasoning, coding |
| `gemini-3-flash-preview` | 1M tokens | Fast responses, simple tasks |
| `gemini-2.5-pro` | 1M tokens | General purpose, coding |
| `gemini-2.5-flash` | 1M tokens | Balanced performance |
| `gemini-2.5-flash-lite` | 1M tokens | Quick tasks, low latency |

---

## Configuration Reference

### Settings Schema

```json
{
  // General Settings
  "general": {
    "preferredEditor": "code",
    "vimMode": false,
    "defaultApprovalMode": "default",
    "enableAutoUpdate": true,
    "enableNotifications": false,
    "checkpointing": {
      "enabled": false
    },
    "plan": {
      "directory": "/tmp/gemini-plans",
      "modelRouting": true
    }
  },

  // Model Configuration
  "model": {
    "name": "gemini-2.5-pro",
    "maxSessionTurns": -1,
    "compressionThreshold": 0.5
  },

  // Model Configs & Aliases
  "modelConfigs": {
    "customAliases": {
      "my-model": {
        "modelConfig": {
          "model": "gemini-2.5-pro",
          "generateContentConfig": {
            "temperature": 0.7
          }
        }
      }
    }
  },

  // UI Settings
  "ui": {
    "theme": "default",
    "autoThemeSwitching": true,
    "showLineNumbers": true,
    "showCitations": false,
    "loadingPhrases": "tips"
  },

  // Context Settings
  "context": {
    "fileName": "GEMINI.md",
    "includeDirectoryTree": true,
    "includeDirectories": [],
    "fileFiltering": {
      "respectGitIgnore": true,
      "respectGeminiIgnore": true
    }
  },

  // Tool Settings
  "tools": {
    "sandbox": false,
    "shell": {
      "enableInteractiveShell": true,
      "inactivityTimeout": 300
    },
    "allowed": ["run_shell_command(git status)"],
    "useRipgrep": true
  },

  // MCP Settings
  "mcp": {
    "serverCommand": "npx",
    "allowed": ["github", "postgres"],
    "excluded": []
  },

  // Security Settings
  "security": {
    "disableYoloMode": false,
    "folderTrust": {
      "enabled": true
    },
    "environmentVariableRedaction": {
      "enabled": false
    }
  },

  // Privacy Settings
  "privacy": {
    "usageStatisticsEnabled": true
  }
}
```

### Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `GEMINI_API_KEY` | API key for AI Studio | `AIza...` |
| `GOOGLE_API_KEY` | Google Cloud API key | `...` |
| `GOOGLE_CLOUD_PROJECT` | Google Cloud project ID | `my-project-123` |
| `GOOGLE_CLOUD_LOCATION` | Google Cloud region | `us-central1` |
| `GOOGLE_APPLICATION_CREDENTIALS` | Service account key path | `/path/to/key.json` |
| `GOOGLE_GENAI_USE_VERTEXAI` | Enable Vertex AI | `true` |
| `GEMINI_CLI_SYSTEM_DEFAULTS_PATH` | System defaults path | `/etc/gemini-cli/defaults.json` |
| `GEMINI_CLI_SYSTEM_SETTINGS_PATH` | System settings path | `/etc/gemini-cli/settings.json` |

---

## Extension Management

### Commands

```bash
# Install extension
gemini extensions install <source>
gemini extensions install https://github.com/user/my-extension
gemini extensions install https://github.com/user/my-extension --ref develop
gemini extensions install https://github.com/user/my-extension --auto-update

# Uninstall extension
gemini extensions uninstall <name>

# List extensions
gemini extensions list

# Update extensions
gemini extensions update <name>
gemini extensions update --all

# Enable/Disable
gemini extensions enable <name>
gemini extensions disable <name>

# Development
gemini extensions link <path>
gemini extensions new <path>
gemini extensions validate <path>
```

### Extension Structure

```
extension-name/
├── gemini-extension.json    # Extension metadata
├── commands/                # TOML custom commands
├── skills/                  # SKILL.md files
├── hooks/                   # Event hooks
├── themes/                  # Custom themes
└── README.md               # Documentation
```

### Extension Manifest

```json
{
  "name": "my-extension",
  "version": "1.0.0",
  "description": "Extension description",
  "author": "Author Name",
  "commands": ["command-name"],
  "skills": ["skill-name"],
  "hooks": ["hook-name"],
  "themes": ["theme-name"]
}
```

---

## MCP Server Management

### Commands

```bash
# Add stdio MCP server
gemini mcp add <name> <command>
gemini mcp add github npx -y @modelcontextprotocol/server-github

# Add HTTP MCP server
gemini mcp add api-server http://localhost:3000 --transport http

# Add with environment variables
gemini mcp add slack node server.js --env SLACK_TOKEN=xoxb-xxx

# Add with user scope
gemini mcp add db node db-server.js --scope user

# Add with specific tools
gemini mcp add github npx -y @modelcontextprotocol/server-github --include-tools list_repos,get_pr

# Remove MCP server
gemini mcp remove <name>

# List MCP servers
gemini mcp list
```

### MCP Configuration in settings.json

```json
{
  "mcp": {
    "servers": {
      "github": {
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-github"],
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
      }
    }
  }
}
```

---

## Skills Management

### Commands

```bash
# List skills
gemini skills list

# Install skill
gemini skills install <source>
gemini skills install https://github.com/user/skill-repo

# Link local skill
gemini skills link <path>

# Uninstall skill
gemini skills uninstall <name>

# Enable/Disable
gemini skills enable <name>
gemini skills disable <name>
gemini skills enable --all
gemini skills disable --all
```

### Skill Structure

```
skill-name/
├── SKILL.md                 # Skill definition and instructions
└── README.md               # Documentation
```

---

## Custom Commands

### TOML Format

```toml
# ~/.gemini/commands/git/commit.toml
# Invoked via: /git:commit

description = "Generates a Git commit message based on staged changes."

prompt = """
Please generate a Conventional Commit message based on the following git diff:

```diff
!{git diff --staged}
```
"""
```

### Argument Handling

```toml
# With {{args}} placeholder
description = "Search for pattern"
prompt = """
Search for "{{args}}" in the codebase.

Results:
!{grep -r {{args}} .}
"""
```

### File Injection

```toml
# With @{...} syntax
description = "Review with best practices"
prompt = """
Review {{args}} using these best practices:

@{docs/best-practices.md}
"""
```

---

## GEMINI.md Format

### Basic Structure

```markdown
# Project Context

## Overview
Brief project description

## Tech Stack
- React 18
- TypeScript
- Node.js 20+

## Commands
- `npm run dev` - Start dev server
- `npm run build` - Production build
- `npm test` - Run tests

## Conventions
- Use functional components
- Follow ESLint rules
- Prefer named exports
```

### Context Hierarchy

1. `~/.gemini/GEMINI.md` - Global context
2. Workspace GEMINI.md files (discovered up tree)
3. JIT context from accessed directories

---

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | General error |
| `2` | Invalid arguments |
| `130` | Interrupted (Ctrl+C) |

---

## Related Documentation

- [Architecture](./ARCHITECTURE.md) - System design
- [Usage Guide](./USAGE.md) - Practical examples
- [External References](./REFERENCES.md) - Links and resources
