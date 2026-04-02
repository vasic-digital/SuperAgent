# Kilo-Code User Guide

## Table of Contents
1. [Installation](#installation)
2. [Quick Start](#quick-start)
3. [CLI Commands](#cli-commands)
4. [TUI/Interactive Commands](#tuiinteractive-commands)
5. [Configuration](#configuration)
6. [API/Protocol Endpoints](#api-protocol-endpoints)
7. [Usage Examples](#usage-examples)
8. [Troubleshooting](#troubleshooting)

---

## Installation

### Method 1: Package Manager (NPM)

The recommended way to install Kilo-Code CLI:

```bash
npm install -g @kilocode/cli
```

Update to the latest version:
```bash
npm update -g @kilocode/cli
```

Or upgrade via CLI:
```bash
kilo upgrade
```

### Method 2: Direct Download (Baseline for Older CPUs)

For older CPUs without AVX support (e.g., Intel Xeon Nehalem, AMD Bulldozer):

1. Go to Kilo Releases on GitHub
2. Download the `-baseline` variant for your platform:
   - Linux x64: `kilo-linux-x64-baseline.tar.gz`
   - macOS x64: `kilo-darwin-x64-baseline.zip`
   - Windows x64: `kilo-windows-x64-baseline.zip`
3. Extract and run the `kilo` binary directly

### Method 3: Build from Source

```bash
git clone https://github.com/Kilo-Org/kilocode.git
cd kilocode/packages/kilo-cli
npm install
npm run build
npm install -g .
```

---

## Quick Start

```bash
# Verify installation
kilo --version
kilo --help

# Start interactive TUI in current directory
kilo

# Start with specific mode
kilo --mode architect

# Start with specific workspace
kilo --workspace /path/to/project

# Resume last conversation
kilo --continue
```

After installation, run `kilo` and use the `/connect` command to add your first provider credentials.

---

## CLI Commands

### Global Options

| Option | Short | Description | Example |
|--------|-------|-------------|---------|
| `--help` | `-h` | Show help | `kilo --help` |
| `--version` | `-v` | Show version | `kilo --version` |
| `--print-logs` | | Print logs to stderr | `kilo --print-logs` |
| `--log-level` | | Log level (DEBUG, INFO, WARN, ERROR) | `kilo --log-level DEBUG` |
| `--mode` | | Start with specific mode | `kilo --mode code` |
| `--workspace` | | Specify workspace path | `kilo --workspace ./my-project` |
| `--continue` | | Resume last conversation | `kilo --continue` |

### Command: kilo [project]

**Description:** Start the TUI (Terminal User Interface) for interactive coding.

**Usage:**
```bash
kilo [path/to/project]
```

**Options:**
| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `project` | string | No | Current directory | Path to project directory |

**Examples:**
```bash
# Start in current directory
kilo

# Start with specific project
kilo ~/projects/my-app
```

**Exit Codes:**
- `0` - Success
- `1` - General error
- `130` - Interrupted (Ctrl+C)

### Command: kilo run

**Description:** Run with a message (non-interactive mode).

**Usage:**
```bash
kilo run [message...] [options]
```

**Options:**
| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `--model` | string | No | Default | Specify model to use |
| `--mode` | string | No | code | Agent mode to use |

**Examples:**
```bash
kilo run "Create a React component for user login"
kilo run "Refactor authentication module" --mode architect
```

### Command: kilo auth

**Description:** Manage credentials (login, logout, list).

**Usage:**
```bash
kilo auth [subcommand]
```

**Subcommands:**
| Subcommand | Description |
|------------|-------------|
| `login` | Authenticate with a provider |
| `logout` | Remove stored credentials |
| `list` | List configured providers |

**Examples:**
```bash
kilo auth login
kilo auth list
kilo auth logout
```

### Command: kilo models

**Description:** List available models from configured providers.

**Usage:**
```bash
kilo models [provider]
```

**Examples:**
```bash
# List all available models
kilo models

# List models from specific provider
kilo models openai
```

### Command: kilo mcp

**Description:** Manage MCP servers (list, add, auth).

**Usage:**
```bash
kilo mcp [subcommand] [options]
```

**Subcommands:**
| Subcommand | Description |
|------------|-------------|
| `list` | List configured MCP servers |
| `add` | Add a new MCP server |
| `auth` | Authenticate with MCP server |

**Examples:**
```bash
kilo mcp list
kilo mcp add
```

### Command: kilo session

**Description:** Manage sessions (list, export, import).

**Usage:**
```bash
kilo session [subcommand]
```

**Subcommands:**
| Subcommand | Description |
|------------|-------------|
| `list` | List available sessions |
| `export` | Export session data |
| `import` | Import session data |

**Examples:**
```bash
kilo session list
kilo export my-session-id > session.json
kilo import session.json
```

### Command: kilo stats

**Description:** Show token usage and cost statistics.

**Usage:**
```bash
kilo stats [options]
```

**Options:**
| Option | Type | Description |
|--------|------|-------------|
| `--format` | string | Output format (table, json) |

**Examples:**
```bash
kilo stats
kilo stats --format json
```

### Command: kilo serve

**Description:** Start a headless server.

**Usage:**
```bash
kilo serve [options]
```

**Options:**
| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `--port` | number | No | 3000 | Server port |
| `--host` | string | No | localhost | Server host |

**Examples:**
```bash
kilo serve
kilo serve --port 8080
```

### Command: kilo web

**Description:** Start server and open web interface.

**Usage:**
```bash
kilo web [options]
```

**Options:**
| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `--port` | number | No | 3000 | Server port |

**Examples:**
```bash
kilo web
kilo web --port 8080
```

---

## TUI/Interactive Commands

Once inside the Kilo-Code TUI, you can use the following slash commands:

| Command | Shortcut | Description | Example |
|---------|----------|-------------|---------|
| `/help` | `/?` | Show help message | `/help` |
| `/connect` | | Add provider credentials | `/connect` |
| `/models` | | List available models | `/models` |
| `/mode` | | Switch agent mode | `/mode architect` |
| `/clear` | | Clear conversation history | `/clear` |
| `/local-review` | | Review current branch changes | `/local-review` |
| `/local-review-uncommitted` | | Review uncommitted changes | `/local-review-uncommitted` |
| `/newtask` | | Start a new task | `/newtask` |
| `/smol` | | Use smol agent | `/smol` |
| `/quit` | | Exit Kilo-Code | `/quit` |

### Agent Modes

| Mode | Description |
|------|-------------|
| `ask` | Answer questions without modifying code |
| `architect` | Plan and design code structure |
| `code` | Write and edit code |
| `debug` | Debug and fix issues |
| `orchestrator` | Coordinate multiple agents |

---

## Configuration

### Configuration File Format

Kilo-Code uses JSONC (JSON with Comments) for configuration:

**File Location:** `~/.config/kilo/kilo.jsonc`

```jsonc
{
  "providers": {
    "openai": {
      "apiKey": "sk-...",
      "model": "gpt-4"
    },
    "anthropic": {
      "apiKey": "sk-ant-...",
      "model": "claude-3-opus"
    }
  },
  "execute": {
    "enabled": true,
    "allowed": [
      "npm",
      "git status",
      "ls -la"
    ],
    "denied": [
      "git push --force",
      "rm -rf /"
    ]
  },
  "skills": {
    "enabled": true
  }
}
```

### Command Approval Patterns

The `execute.allowed` and `execute.denied` lists support hierarchical pattern matching:

- **Base command:** `"git"` matches any git command
- **Command + subcommand:** `"git status"` matches any git status command
- **Full command:** `"git status --short"` only matches exactly that command

### Environment Variables

| Variable | Description |
|----------|-------------|
| `KILO_API_KEY` | Default API key for providers |
| `OPENAI_API_KEY` | OpenAI provider API key |
| `ANTHROPIC_API_KEY` | Anthropic provider API key |
| `KILO_CONFIG_DIR` | Custom config directory path |
| `KILO_LOG_LEVEL` | Log level (DEBUG, INFO, WARN, ERROR) |

### Configuration Locations

| Platform | Path |
|----------|------|
| Linux | `~/.config/kilo/` |
| macOS | `~/Library/Application Support/kilo/` |
| Windows | `%APPDATA%\kilo\` |

### Skills System

Skills are discovered from:
- **Global skills:** `~/.kilocode/skills/` (available in all projects)
- **Project skills:** `.kilocode/skills/` (project-specific)

Skills can be:
- **Generic** - Available in all modes
- **Mode-specific** - Only loaded when using a particular mode

Example structure:
```
your-project/
└── .kilocode/
    ├── skills/                # Generic skills
    │   └── project-conventions/
    │       └── SKILL.md
    └── skills-code/           # Code mode skills
        └── linting-rules/
            └── SKILL.md
```

---

## API/Protocol Endpoints

### Local Server Mode

When running `kilo serve`, the following endpoints are available:

#### Endpoint: POST /v1/chat

**Description:** Send a chat message to the AI agent.

**Request:**
```json
{
  "message": "Create a function to calculate fibonacci",
  "context": {
    "mode": "code",
    "workspace": "/path/to/project"
  }
}
```

**Response:**
```json
{
  "response": "Here's a fibonacci function...",
  "actions": [
    {
      "type": "write_file",
      "path": "fibonacci.js",
      "content": "function fibonacci(n) {...}"
    }
  ]
}
```

**Example:**
```bash
curl -X POST http://localhost:3000/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello", "context": {"mode": "ask"}}'
```

#### Endpoint: GET /v1/models

**Description:** List available models.

**Response:**
```json
{
  "models": [
    {
      "id": "gpt-4",
      "provider": "openai",
      "capabilities": ["chat", "code"]
    }
  ]
}
```

---

## Usage Examples

### Example 1: Code Generation

```bash
# Start Kilo-Code
cd my-project
kilo

# In the TUI, type:
Create a React component that displays a user profile with name, email, and avatar

# Review the generated code and approve changes
```

### Example 2: Code Review

```bash
# Review current branch changes
kilo

# In the TUI:
/local-review

# Or review uncommitted changes
/local-review-uncommitted
```

### Example 3: Refactoring with Architect Mode

```bash
# Start in architect mode
kilo --mode architect

# In the TUI:
Refactor the authentication system to use JWT tokens instead of session cookies

# Review the plan, then switch to code mode to implement
/mode code
```

### Example 4: Custom Skills

Create a custom skill for your project's conventions:

```bash
mkdir -p .kilocode/skills/project-conventions
cat > .kilocode/skills/project-conventions/SKILL.md << 'EOF'
---
description: Project coding conventions
---

# Project Conventions

- Use TypeScript for all new files
- Follow the existing folder structure
- Use async/await instead of callbacks
- Write tests for all new functions
EOF
```

### Example 5: Custom Commands

Create a custom slash command:

```bash
mkdir -p ~/.kilocode/commands
cat > ~/.kilocode/commands/component.md << 'EOF'
---
description: Create a new React component
arguments:
  - ComponentName
---

Create a new React component named $1.
Include:
- Proper TypeScript typing
- Basic component structure
- Export statement
- A simple props interface if appropriate
EOF
```

Use it in the TUI:
```
/component UserCard
```

---

## Troubleshooting

### Issue: "Illegal instruction" error on startup

**Cause:** Your CPU doesn't support AVX instructions.

**Solution:** Download the baseline variant from GitHub releases instead of using npm.

### Issue: API key not recognized

**Solution:**
```bash
# Check if environment variable is set
echo $OPENAI_API_KEY

# Or use /connect command in TUI
kilo
/connect
```

### Issue: Commands not executing

**Solution:** Check your `execute` configuration in `~/.config/kilo/kilo.jsonc`. Ensure the command is in the `allowed` list and not in the `denied` list.

### Issue: Skills not loading

**Solution:**
1. Verify skill file location: `~/.kilocode/skills/` or `.kilocode/skills/`
2. Check SKILL.md file format
3. Restart Kilo-Code after adding skills

### Issue: MCP server connection failed

**Solution:**
```bash
# List configured MCP servers
kilo mcp list

# Re-authenticate
kilo mcp auth <server-name>
```

### Issue: Session not resuming

**Solution:**
```bash
# List available sessions
kilo session list

# Start with specific session
kilo --continue
```

### Issue: High token usage

**Solution:**
- Use `/compress` to summarize conversation history
- Switch to a more efficient model
- Clear conversation history with `/clear`

---

## Additional Resources

- **Website:** https://kilo.ai
- **Documentation:** https://kilo.ai/docs
- **GitHub:** https://github.com/Kilo-Org/kilocode
- **VS Code Extension:** Available on the VS Code Marketplace
