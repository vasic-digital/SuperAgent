# GPTMe User Guide

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

### Method 1: pipx (Recommended)

```bash
# Install with pipx
pipx install gptme

# Verify installation
gptme --version

# Upgrade
pipx upgrade gptme
```

### Method 2: uv

```bash
# Install with uv
uv tool install gptme

# Upgrade
uv tool upgrade gptme
```

### Method 3: pip

```bash
# Install with pip
pip install gptme

# With optional extras
pip install 'gptme[browser]'  # Playwright support
pip install 'gptme[all]'      # All extras
```

### Method 4: Build from Source

```bash
# Clone repository
git clone https://github.com/gptme/gptme.git
cd gptme

# Install with pipx
pipx install .

# Or with uv
uv tool install .

# Latest from git with all extras
uv tool install 'git+https://github.com/gptme/gptme.git[all]'
```

### Prerequisites

- Python 3.10 or newer
- API key from Anthropic, OpenAI, or OpenRouter
- Git (optional, for git integration)

## Quick Start

### First-Time Setup

```bash
# Verify installation
gptme --version

# Set up API key
export ANTHROPIC_API_KEY="sk-ant-..."
# OR
export OPENAI_API_KEY="sk-..."
# OR
export OPENROUTER_API_KEY="sk-or-..."

# Run setup
gptme /setup

# Start interactive session
gptme
```

### Basic Usage

```bash
# Start interactive session
gptme

# Run with prompt
gptme "explain this codebase"

# Run with specific model
gptme -m anthropic/claude-sonnet-4-20250514 "task"

# Non-interactive mode
gptme -y "run tests and fix failures"

# Resume conversation
gptme -r
```

### Hello World

```bash
# Start GPTMe
gptme

# At the prompt:
> Create a Python script that prints "Hello, World!"

# Or non-interactive:
gptme "Create a Python script that prints 'Hello, World!'"
```

## CLI Commands

### Global Options

| Option | Short | Description | Example |
|--------|-------|-------------|---------|
| --version | | Show version | `gptme --version` |
| --help | -h | Show help | `gptme --help` |
| --name | | Conversation name | `gptme --name my-chat` |
| --model | -m | Select model | `gptme -m gpt-4o` |
| --workspace | -w | Workspace directory | `gptme -w ./project` |
| --agent-path | | Agent workspace path | `gptme --agent-path ./agent` |
| --resume | -r | Resume conversation | `gptme -r` |
| --no-confirm | -y | Skip confirmations | `gptme -y "task"` |
| --non-interactive | -n | Non-interactive mode | `gptme -n "task"` |
| --system | | System prompt | `gptme --system short` |
| --tools | -t | Enable tools | `gptme -t shell,read` |
| --tool-format | | Tool format | `gptme --tool-format markdown` |
| --no-stream | | Disable streaming | `gptme --no-stream` |
| --show-hidden | | Show hidden messages | `gptme --show-hidden` |
| --verbose | -v | Verbose output | `gptme -v` |

### Command: (default - interactive)

**Description:** Start interactive chat session.

**Usage:**
```bash
gptme [OPTIONS] [PROMPTS...]
```

**Examples:**
```bash
# Start interactive session
gptme

# With initial prompt
gptme "explain this codebase"

# Chain prompts
gptme "create tests" - "run tests" - "fix failures"

# With files
gptme "review this" main.py

# With specific tools only
gptme -t read,save "analyze code"
```

**Exit Codes:**
| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 130 | Interrupted |

### Command: gptme-server

**Description:** Start web server for browser access.

**Usage:**
```bash
gptme-server [OPTIONS]
```

**Options:**
| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| --host | string | No | 0.0.0.0 | Server host |
| --port | int | No | 5000 | Server port |
| --debug | boolean | No | false | Debug mode |

**Examples:**
```bash
# Start server
gptme-server

# Custom port
gptme-server --port 8080

# Localhost only
gptme-server --host 127.0.0.1
```

### Command: gptme-eval

**Description:** Run evaluation suite.

**Usage:**
```bash
gptme-eval [OPTIONS]
```

**Examples:**
```bash
# Run evals
gptme-eval

# With specific model
gptme-eval --model claude-sonnet-4
```

### Command: gptme-util

**Description:** Utility commands for GPTMe.

**Usage:**
```bash
gptme-util [SUBCOMMAND]
```

**Subcommands:**
| Subcommand | Description |
|------------|-------------|
| tools | Tool management |
| chats | Chat management |
| models | Model management |
| context | Context management |
| llm | Direct LLM access |

**Examples:**
```bash
# List tools
gptme-util tools list

# Tool info
gptme-util tools info shell

# List chats
gptme-util chats list

# List models
gptme-util models list
```

### Command: gptme-util skills

**Description:** Manage GPTMe skills.

**Usage:**
```bash
gptme-util skills [SUBCOMMAND]
```

**Subcommands:**
| Subcommand | Description |
|------------|-------------|
| list | List available skills |
| installed | List installed skills |
| install | Install a skill |
| init | Create new skill |
| check | Validate skill |
| publish | Package skill |
| dirs | Show skill directories |

**Examples:**
```bash
# List skills
gptme-util skills list

# Install skill
gptme-util skills install code-review-helper

# Create new skill
gptme-util skills init ./my-skill --name my-skill -d "Does cool things"

# Validate skill
gptme-util skills check
```

## TUI/Interactive Commands

When running in interactive mode, use these commands (prefix with `/`):

| Command | Description | Example |
|---------|-------------|---------|
| /help | Show help | `/help` |
| /exit | Exit GPTMe | `/exit` |
| /undo | Undo last action | `/undo` |
| /log | Show conversation log | `/log` |
| /tools | Show available tools | `/tools` |
| /model | Show/switch model | `/model` or `/model gpt-4o` |
| /models | List models | `/models` |
| /edit | Edit conversation | `/edit` |
| /rename | Rename conversation | `/rename new-name` |
| /fork | Fork conversation | `/fork copy-name` |
| /delete | Delete conversation | `/delete <id>` |
| /summarize | Summarize conversation | `/summarize` |
| /replay | Replay tool operations | `/replay` |
| /export | Export as HTML | `/export chat.html` |
| /commit | Git commit | `/commit` |
| /compact | Compact conversation | `/compact` |
| /impersonate | Impersonate assistant | `/impersonate` |
| /plugin | Manage plugins | `/plugin list` |
| /clear | Clear screen | `/clear` |
| /setup | Run setup | `/setup` |
| /doctor | Run diagnostics | `/doctor` |
| /restart | Restart GPTMe | `/restart` |
| /tokens | Show token usage | `/tokens` |
| /context | Show context breakdown | `/context` |

### Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| Ctrl+X Ctrl+E | Edit prompt in editor |
| Ctrl+J | Insert new line |
| Ctrl+C | Cancel operation |
| Ctrl+D | Exit |
| Up/Down | Navigate history |

### Tool Shortcuts

| Shortcut | Action |
|----------|--------|
| @file | Include file in context |
| @url | Include URL content |
| @gh | Include GitHub PR/issue |

## Configuration

### Configuration File Format

GPTMe uses `~/.config/gptme/config.toml`:

```toml
[model]
default = "anthropic/claude-sonnet-4-20250514"

[api_keys]
anthropic = "sk-ant-..."
openai = "sk-..."
openrouter = "sk-or-..."

[tools]
enabled = ["shell", "read", "save", "patch", "browser", "vision"]
disabled = ["tts", "youtube"]

[behavior]
auto_confirm = false
stream = true
tool_format = "markdown"

[display]
theme = "dark"
show_cost = true
show_tokens = true

[server]
host = "0.0.0.0"
port = 5000
```

### Project Configuration (gptme.toml)

Create `gptme.toml` in project root:

```toml
[project]
name = "My Project"
description = "A sample project"

[context]
files = [
    "README.md",
    "src/main.py",
    "docs/API.md"
]
cmd = "git status"

[tools]
enabled = ["shell", "read", "save", "patch", "ipython"]

[behavior]
system_prompt = "You are an expert Python developer."
```

### Environment Variables

| Variable | Description | Required | Example |
|----------|-------------|----------|---------|
| ANTHROPIC_API_KEY | Anthropic API key | Yes* | `sk-ant-...` |
| OPENAI_API_KEY | OpenAI API key | Yes* | `sk-...` |
| OPENROUTER_API_KEY | OpenRouter API key | Yes* | `sk-or-...` |
| GPTME_CONFIG | Config file path | No | `~/.config/gptme/config.toml` |
| EDITOR | Editor for /edit | No | `vim` |
| PAGER | Pager for output | No | `less` |

*At least one provider required

### Configuration Locations (in order of precedence)

1. Command-line flags
2. Environment variables
3. Project config (`./gptme.toml`)
4. User config (`~/.config/gptme/config.toml`)

## API/Protocol Endpoints

### gptme-server REST API

When running `gptme-server`:

**Health Check:**
```bash
curl http://localhost:5000/health
```

**Chat Request:**
```bash
curl -X POST http://localhost:5000/api/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Hello, how are you?",
    "conversation_id": "abc123"
  }'
```

**Response:**
```json
{
  "response": "I'm doing well, thank you!",
  "conversation_id": "abc123",
  "tokens_used": 25
}
```

### Supported Providers

| Provider | Environment Variable | Base URL |
|----------|---------------------|----------|
| Anthropic | ANTHROPIC_API_KEY | api.anthropic.com |
| OpenAI | OPENAI_API_KEY | api.openai.com |
| OpenRouter | OPENROUTER_API_KEY | openrouter.ai/api |
| Local/Ollama | | localhost:11434 |

### Local Models (Ollama)

```bash
# Start Ollama
ollama serve

# Pull model
ollama pull codellama

# Use with GPTMe
gptme -m ollama/codellama "task"
```

## Usage Examples

### Example 1: File Operations

```bash
# Start GPTMe
gptme

# Read and edit files
> Read the contents of main.py

> Add a function to calculate factorial

> Save the changes

# Or with files in prompt
gptme main.py "add error handling to this file"
```

### Example 2: Shell Commands

```bash
# Execute shell commands
> Run the test suite

> Show git status

> Find all Python files modified today
> Run flake8 on them
```

### Example 3: Web Browsing

```bash
# Browse web (requires browser extra)
gptme 'gptme[browser]'

# Or if already installed
> Fetch the content from https://docs.python.org/3/library/asyncio.html
> Summarize the key concepts
```

### Example 4: Git Integration

```bash
# Git workflow
> Show the git diff

> Generate a commit message for these changes
/commit

> Create a new branch for the feature
> Push the branch to origin
```

### Example 5: Python Execution

```bash
# Execute Python code
> Calculate the first 10 prime numbers using Python

> Plot a sine wave and save it to plot.png

> Analyze this CSV file [attach file]
```

### Example 6: Piped Input

```bash
# Process command output
git diff | gptme "explain these changes"

# Process file
cat error.log | gptme "analyze these errors"

# Fix failing tests
make test 2>&1 | gptme "fix the failing tests"
```

### Example 7: Non-Interactive Automation

```bash
# Auto-approve all actions
gptme -y "run tests and fix failures"

# Fully non-interactive (for CI/CD)
gptme -n "generate documentation"

# Chain multiple prompts
gptme "create tests" - "run tests" - "fix failures" -y
```

### Example 8: Working with URLs and GitHub

```bash
# Include URL content
gptme @https://example.com/docs "implement this API"

# Include GitHub PR
gptme @https://github.com/org/repo/pull/123 "review this PR"

# Include GitHub issue
gptme @https://github.com/org/repo/issues/456 "implement this feature"
```

### Example 9: Agent Workspaces

```bash
# Create agent workspace
gptme --agent-path ./my-agent "create a coding assistant"

# The agent maintains its own:
# - TASKS.md for tracking
# - journal/ for logs
# - knowledge/ for learned info
# - people/ for contacts
```

### Example 10: Skills System

```bash
# Install a skill
gptme-util skills install python-best-practices

# Use in conversation
gptme
> /skills
> Apply Python best practices to refactor this code
```

## Troubleshooting

### Issue: API Key Not Found

**Symptoms:** "API key not found" error

**Solution:**
```bash
# Set environment variable
export ANTHROPIC_API_KEY="sk-ant-..."

# Or configure interactively
gptme /setup

# Or edit config
vim ~/.config/gptme/config.toml
```

### Issue: Tool Not Available

**Symptoms:** "Tool not available" error

**Solution:**
```bash
# List available tools
/tools

# Check tool requirements
gptme-util tools info <tool>

# Install extras if needed
pipx install 'gptme[all]'
```

### Issue: Browser Tool Not Working

**Symptoms:** Browser commands fail

**Solution:**
```bash
# Install browser extra
pipx install 'gptme[browser]'

# Or with uv
uv tool install 'gptme[browser]'

# Install Playwright
playwright install
```

### Issue: Context Window Full

**Symptoms:** "Context window exceeded" error

**Solution:**
```bash
# Compact conversation
/compact

# Start new conversation
/new

# Use model with larger context
gptme -m claude-opus-4 "task"
```

### Issue: Resume Not Working

**Symptoms:** Cannot resume previous conversation

**Solution:**
```bash
# List conversations
gptme-util chats list

# Resume most recent
gptme -r

# Resume specific
gptme-util chats list
gptme --resume <name>
```

### Issue: Server Won't Start

**Symptoms:** `gptme-server` fails to start

**Solution:**
```bash
# Check port available
lsof -i :5000

# Use different port
gptme-server --port 8080

# Check Python version
python --version  # Need 3.10+
```

### Issue: Skills Not Loading

**Symptoms:** Skills commands not working

**Solution:**
```bash
# List installed skills
gptme-util skills installed

# Check skill directories
gptme-util skills dirs

# Validate skill
gptme-util skills check
```

---

**Last Updated:** 2026-04-02
**Version:** 0.28.0
