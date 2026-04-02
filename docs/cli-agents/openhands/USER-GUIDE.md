# OpenHands User Guide

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

### Method 1: uv (Recommended)

```bash
# Install uv first if not installed
curl -LsSf https://astral.sh/uv/install.sh | sh

# Install OpenHands
uv tool install openhands --python 3.12

# Verify installation
openhands --version

# Upgrade
uv tool upgrade openhands --python 3.12
```

### Method 2: pip

```bash
# Install with pip
pip install openhands-ai

# Or with specific Python version
python3.12 -m pip install openhands-ai

# Verify
openhands --version
```

### Method 3: Docker

```bash
# Run with Docker
docker run -it \
    --pull=always \
    -e AGENT_SERVER_IMAGE_REPOSITORY=ghcr.io/openhands/agent-server \
    -e AGENT_SERVER_IMAGE_TAG=1.15.0-python \
    -e SANDBOX_USER_ID=$(id -u) \
    -e SANDBOX_VOLUMES=$(pwd) \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v ~/.openhands:/root/.openhands \
    --add-host host.docker.internal:host-gateway \
    python:3.12-slim \
    bash -c "pip install uv && uv tool install openhands --python 3.12 && openhands"
```

### Method 4: Binary Install

```bash
# Install using install script
curl -fsSL https://install.openhands.dev/install.sh | sh

# Run OpenHands
openhands
```

### Method 5: Build from Source

```bash
# Clone repository
git clone https://github.com/All-Hands-AI/OpenHands.git
cd OpenHands

# Install with uv
uv tool install . --python 3.12

# Or with pip
pip install -e .
```

### Prerequisites

- Python 3.12 or 3.13
- uv (recommended) or pip
- Docker (optional, for sandboxed execution)
- Git

## Quick Start

### First-Time Setup

```bash
# Verify installation
openhands --version

# Run OpenHands
openhands

# On first run, configure LLM settings
# You'll be prompted to:
# 1. Select LLM provider (Anthropic, OpenAI, etc.)
# 2. Enter API key
# 3. Choose default model

# Settings are saved to ~/.openhands/settings.json
```

### Basic Usage

```bash
# Start interactive CLI
openhands

# Start with a task
openhands -t "Fix the bug in auth.py"

# Load task from file
openhands -f task.txt

# Start headless mode
openhands --headless

# Start web interface
openhands web
```

### Hello World

```bash
# Start OpenHands
openhands

# At the prompt, type:
Create a Python script that prints "Hello, World!"

# OpenHands will create the file and show results
```

## CLI Commands

### Global Options

| Option | Short | Description | Example |
|--------|-------|-------------|---------|
| --version | -v | Show version | `openhands --version` |
| --help | -h | Show help | `openhands --help` |
| --task | -t | Initial task | `openhands -t "fix bug"` |
| --file | -f | Load task from file | `openhands -f task.txt` |
| --headless | | Run without UI | `openhands --headless` |
| --resume | | Resume conversation | `openhands --resume` |
| --last | | Resume last conversation | `openhands --resume --last` |
| --always-approve | | Auto-approve all actions | `openhands --always-approve` |
| --llm-approve | | Use LLM security analyzer | `openhands --llm-approve` |
| --override-with-envs | | Override with env vars | `openhands --override-with-envs` |
| --config | -c | Config file path | `openhands -c custom.json` |

### Command: serve

**Description:** Launch the OpenHands GUI server using Docker.

**Usage:**
```bash
openhands serve [OPTIONS]
```

**Options:**
| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| --mount-cwd | boolean | No | false | Mount current directory |
| --gpu | boolean | No | false | Enable GPU support |
| --port | int | No | 3000 | Server port |
| --host | string | No | 0.0.0.0 | Server host |

**Examples:**
```bash
# Start server
openhands serve

# With current directory mounted
openhands serve --mount-cwd

# With GPU support
openhands serve --gpu

# Custom port
openhands serve --port 8080
```

**Exit Codes:**
| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Server error |
| 130 | Interrupted |

### Command: web

**Description:** Launch the CLI as a web application accessible via browser.

**Usage:**
```bash
openhands web [OPTIONS]
```

**Options:**
| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| --host | string | No | 0.0.0.0 | Host to bind |
| --port | int | No | 12000 | Port to bind |
| --debug | boolean | No | false | Enable debug mode |

**Examples:**
```bash
# Start web interface
openhands web

# Custom port
openhands web --port 8080

# Localhost only
openhands web --host 127.0.0.1
```

### Command: cloud

**Description:** Create a new conversation in OpenHands Cloud.

**Usage:**
```bash
openhands cloud [OPTIONS]
```

**Options:**
| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| --task | -t | string | No | Initial task |
| --file | -f | path | No | Task file path |
| --server-url | URL | No | https://app.all-hands.dev | Server URL |

**Examples:**
```bash
# Start cloud task
openhands cloud -t "Fix the bug"

# From file
openhands cloud -f task.txt

# Custom server
openhands cloud --server-url https://custom.server.com -t "Task"
```

### Command: acp

**Description:** Start the Agent Client Protocol server for IDE integrations.

**Usage:**
```bash
openhands acp [OPTIONS]
```

**Options:**
| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| --resume | [ID] | No | | Resume conversation by ID |
| --last | boolean | No | false | Resume most recent |
| --always-approve | boolean | No | false | Auto-approve all actions |
| --llm-approve | boolean | No | false | LLM security analyzer |
| --streaming | boolean | No | false | Enable token streaming |

**Examples:**
```bash
# Start ACP server
openhands acp

# With LLM approval
openhands acp --llm-approve

# Resume specific conversation
openhands acp --resume abc123def456
```

### Command: resume

**Description:** Resume a previous conversation.

**Usage:**
```bash
openhands --resume [OPTIONS] [CONVERSATION_ID]
```

**Examples:**
```bash
# List and select
openhands --resume

# Resume most recent
openhands --resume --last

# Resume specific
openhands --resume abc123def456
```

## TUI/Interactive Commands

When running in interactive/TUI mode, use these commands (prefix with `/`):

| Command | Shortcut | Description | Example |
|---------|----------|-------------|---------|
| /help | | Display available commands | `/help` |
| /new | | Start new conversation | `/new` |
| /history | | Toggle conversation history | `/history` |
| /confirm | | Configure confirmation settings | `/confirm` |
| /condense | | Condense conversation history | `/condense` |
| /skills | | View loaded skills, hooks, MCPs | `/skills` |
| /feedback | | Send feedback | `/feedback` |
| /exit | Ctrl+Q | Exit the application | `/exit` |

### Command Palette

Press `Ctrl+P` (or `Ctrl+\`) to open the command palette:

| Option | Description |
|--------|-------------|
| History | Toggle conversation history panel |
| Keys | Show keyboard shortcuts |
| MCP | View MCP server configurations |
| Maximize | Maximize/restore window |
| Plan | View agent plan |
| Quit | Quit the application |
| Screenshot | Take a screenshot |
| Settings | Configure LLM, API keys, settings |
| Theme | Toggle color theme |

### Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| Ctrl+P | Open command palette |
| Ctrl+\ | Open command palette (alternative) |
| Esc | Pause the running agent |
| Ctrl+Q | Exit the CLI |
| Up/Down | Navigate history |

### Confirmation Modes

| Mode | Description |
|------|-------------|
| default | Always ask for confirmation |
| always-approve | Auto-approve all actions |
| llm-approve | Use LLM-based security analyzer |

## Configuration

### Configuration File Format

OpenHands uses `~/.openhands/settings.json`:

```json
{
  "llm": {
    "model": "claude-sonnet-4-20250514",
    "api_key": "sk-ant-...",
    "base_url": "https://api.anthropic.com",
    "temperature": 0.7,
    "max_tokens": 4096
  },
  "agent": {
    "name": "CodeActAgent",
    "version": "1.0"
  },
  "sandbox": {
    "type": "docker",
    "image": "python:3.12-slim",
    "timeout": 120
  },
  "confirmation_mode": "default",
  "conversation_history": true,
  "mcp_servers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "ghp_..."
      }
    }
  }
}
```

### Environment Variables

| Variable | Description | Required | Example |
|----------|-------------|----------|---------|
| LLM_MODEL | Default model | No | `claude-sonnet-4-20250514` |
| LLM_API_KEY | API key | Yes* | `sk-ant-...` |
| LLM_BASE_URL | Custom base URL | No | `https://api.anthropic.com` |
| SANDBOX_VOLUMES | Volumes to mount | No | `/path/to/project` |
| SANDBOX_USER_ID | User ID for sandbox | No | `1000` |
| OPENHANDS_CONFIG | Config file path | No | `~/.openhands/custom.json` |

*Required for cloud LLM providers

### Configuration Locations (in order of precedence)

1. Command-line flags
2. Environment variables
3. `--override-with-envs` flag
4. User config (`~/.openhands/settings.json`)
5. Project config (`./.openhands.json`)

## API/Protocol Endpoints

### ACP (Agent Client Protocol)

OpenHands implements ACP for IDE integration:

**Start ACP Server:**
```bash
openhands acp --streaming
```

**Connect from IDE:**
- VS Code: Install OpenHands extension
- JetBrains: Install OpenHands plugin
- Zed: Built-in ACP support

### REST API

When running `openhands serve`:

**Health Check:**
```bash
curl http://localhost:3000/api/health
```

**Response:**
```json
{
  "status": "healthy",
  "version": "0.36.0"
}
```

**Create Task:**
```bash
curl -X POST http://localhost:3000/api/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "task": "Fix the bug in auth.py",
    "workspace": "/path/to/project"
  }'
```

**Response:**
```json
{
  "task_id": "task_abc123",
  "status": "pending",
  "created_at": "2026-04-02T10:00:00Z"
}
```

### MCP Server Configuration

Create `~/.openhands/mcp.json`:

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/workspace"]
    },
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "ghp_..."
      }
    },
    "fetch": {
      "command": "uvx",
      "args": ["mcp-server-fetch"]
    }
  }
}
```

## Usage Examples

### Example 1: Bug Fix Task

```bash
# Start OpenHands with a task
openhands -t "Fix the authentication bug where users can't login with special characters"

# Or interactive
openhands
> Fix the authentication bug where users can't login with special characters

# Watch the agent work
# Review changes when complete
```

### Example 2: Feature Implementation

```bash
# Start OpenHands
openhands

# Request feature
> Create a REST API endpoint for user profile management
> Implement GET, POST, PUT, DELETE methods
> Add validation and error handling

# Review and approve each step
```

### Example 3: Code Review

```bash
# Start in ask mode
openhands

> Review the codebase for security vulnerabilities
> Focus on authentication and authorization
> Suggest specific fixes for any issues found
```

### Example 4: Web Development

```bash
# Start OpenHands
openhands

> Create a React component for a data table with sorting and pagination
> Add TypeScript types
> Style with Tailwind CSS
> Write unit tests
```

### Example 5: DevOps Tasks

```bash
# Start with a DevOps task
openhands -t "Create a Docker Compose setup for this application with PostgreSQL and Redis"

# Or
openhands
> Write a GitHub Actions workflow for CI/CD
> Include build, test, and deploy steps
```

### Example 6: Headless Automation

```bash
# Run non-interactively
openhands --headless -t "Run tests and fix any failures"

# With auto-approval
openhands --headless --always-approve -t "Update all dependencies"

# From task file
openhands --headless -f automated-task.txt
```

## Troubleshooting

### Issue: Python Version Error

**Symptoms:** "Python 3.12+ required" error

**Solution:**
```bash
# Check Python version
python --version

# Install Python 3.12
# Using uv
uv python install 3.12

# Or using pyenv
pyenv install 3.12.0
pyenv local 3.12.0
```

### Issue: Docker Sandbox Errors

**Symptoms:** Sandbox cannot start or permission errors

**Solution:**
```bash
# Check Docker is running
docker ps

# Set sandbox user ID
export SANDBOX_USER_ID=$(id -u)

# Set volumes
export SANDBOX_VOLUMES=$(pwd)

# Run with explicit settings
openhands --override-with-envs
```

### Issue: LLM Configuration Errors

**Symptoms:** "No LLM configured" or API errors

**Solution:**
```bash
# Set environment variables
export LLM_MODEL="claude-sonnet-4-20250514"
export LLM_API_KEY="sk-ant-..."

# Or configure interactively
openhands
# Press Ctrl+P → Settings

# Or edit config directly
vim ~/.openhands/settings.json
```

### Issue: Resume Not Working

**Symptoms:** Cannot resume previous conversations

**Solution:**
```bash
# List conversations
ls ~/.openhands/conversations/

# Resume with --last
openhands --resume --last

# Or specify ID
openhands --resume <conversation-id>
```

### Issue: High Token Usage

**Symptoms:** High API costs or rate limiting

**Solution:**
```bash
# Use cheaper model
export LLM_MODEL="gpt-3.5-turbo"

# Condense conversation
/condense

# Start new conversation
/new
```

### Issue: Windows Compatibility

**Symptoms:** OpenHands not working on Windows

**Solution:**
- Use WSL (Windows Subsystem for Linux)
- Or use Docker mode
- Native Windows support is limited

### Issue: MCP Servers Not Loading

**Symptoms:** MCP tools not available

**Solution:**
```bash
# Check MCP config
/skills

# Verify MCP servers in config
cat ~/.openhands/settings.json | jq .mcp_servers

# Restart OpenHands after config changes
```

---

**Last Updated:** 2026-04-02
**Version:** 0.36.0
