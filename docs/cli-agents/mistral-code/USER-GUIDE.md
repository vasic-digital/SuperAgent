# Mistral Code User Guide

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

### Method 1: Package Manager (PIP)

```bash
pip install mistral-code
```

### Method 2: Package Manager (UV)

```bash
uv tool install mistral-vibe
```

### Method 3: Install Script (Recommended)

**Linux and macOS:**
```bash
curl -LsSf https://mistral.ai/vibe/install.sh | bash
```

**Windows:**
```powershell
# First install uv
powershell -ExecutionPolicy ByPass -c "irm https://astral.sh/uv/install.ps1 | iex"

# Then install mistral-vibe
uv tool install mistral-vibe
```

### Method 4: Build from Source

```bash
git clone https://github.com/mistralai/mistral-vibe.git
cd mistral-vibe
pip install -e .
```

---

## Quick Start

```bash
# Set API key
export MISTRAL_API_KEY="your-api-key-here"

# Verify installation
mistral-vibe --version
mistral-vibe --help

# Start interactive mode
mistral-vibe

# Run with a prompt
mistral-vibe "Create a Python function to parse JSON"
```

---

## CLI Commands

### Global Options

| Option | Short | Description | Example |
|--------|-------|-------------|---------|
| `--help` | `-h` | Show help | `mistral-vibe --help` |
| `--version` | `-v` | Show version | `mistral-vibe --version` |
| `--model` | `-m` | Model to use | `mistral-vibe -m codestral-latest` |

### Command: mistral-vibe

**Description:** Start the interactive TUI or run with a prompt.

**Usage:**
```bash
mistral-vibe [prompt] [options]
```

**Options:**
| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `prompt` | string | No | - | Initial prompt to send |
| `--model` | string | No | codestral-latest | Model to use |

**Examples:**
```bash
# Interactive mode
mistral-vibe

# Run with prompt
mistral-vibe "Explain this codebase"

# With specific model
mistral-vibe "Write a function" --model mistral-large-latest
```

**Exit Codes:**
- `0` - Success
- `1` - General error
- `130` - Interrupted (Ctrl+C)

### Command: vibe

**Description:** Alternative command alias for mistral-vibe.

**Usage:**
```bash
vibe [prompt] [options]
```

**Examples:**
```bash
vibe
vibe "Create a REST API"
```

---

## TUI/Interactive Commands

Once inside the Mistral Vibe TUI, use these slash commands:

| Command | Description | Example |
|---------|-------------|---------|
| `/help` | Show help message | `/help` |
| `/models` | List available models | `/models` |
| `/model` | Switch model | `/model codestral-latest` |
| `/clear` | Clear conversation | `/clear` |
| `/exit` | Exit the application | `/exit` |

### Built-in Agents

Mistral Vibe includes built-in agents for different tasks:

| Agent | Description |
|-------|-------------|
| Default | General purpose coding assistant |
| Code | Focused on code generation |
| Ask | Q&A without code modifications |

### Subagents and Task Delegation

Mistral Vibe supports delegating tasks to specialized subagents for parallel execution.

### Skills System

Mistral Vibe uses a skills system for extending capabilities:

**Skill Discovery:**
- Global skills: `~/.mistral/skills/`
- Project skills: `.mistral/skills/`

**Creating Skills:**

Create a `SKILL.md` file in the skills directory:

```markdown
---
description: Database optimization skill
---

# Database Optimization

When working with database queries:
- Always use indexes
- Avoid N+1 queries
- Use EXPLAIN to analyze performance
```

---

## Configuration

### Configuration File Format

Mistral Vibe uses TOML configuration:

**File Location:** `~/.config/mistral/config.toml`

```toml
[api]
api_key = "your-api-key-here"
base_url = "https://api.mistral.ai"

[models]
default = "codestral-latest"

[ui]
theme = "dark"
show_tokens = true

[execution]
auto_approve = false
allowed_commands = ["git", "npm", "pip"]
```

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `MISTRAL_API_KEY` | Mistral API key | Yes |
| `MISTRAL_MODEL` | Default model | No |
| `MISTRAL_MAX_TOKENS` | Max tokens per request | No |

### API Key Configuration

```bash
# Set via environment
export MISTRAL_API_KEY="your-key"

# Or configure in the TUI
mistral-vibe
# Then use /config command
```

### Custom System Prompts

Create `~/.config/mistral/system_prompt.md`:

```markdown
You are a helpful coding assistant specialized in Python.
Always provide type hints and docstrings.
Follow PEP 8 style guidelines.
```

### Custom Agent Configurations

Create custom agents in `~/.config/mistral/agents/`:

```toml
# ~/.config/mistral/agents/python-expert.toml
name = "python-expert"
description = "Python expert agent"
system_prompt = """You are a Python expert.
Focus on clean, efficient, and well-documented code."""
```

### Tool Management

Configure allowed tools in `~/.config/mistral/tools.toml`:

```toml
[tools]
file_read = true
file_write = true
command_execute = false
web_search = true
```

### MCP Server Configuration

```toml
# ~/.config/mistral/mcp.toml
[mcp.servers]
github = { command = "npx", args = ["-y", "@github/mcp-server"] }
```

### Session Management

Sessions are automatically saved to `~/.local/share/mistral/sessions/`.

### Custom Vibe Home Directory

```bash
export VIBE_HOME="/custom/path"
```

---

## API/Protocol Endpoints

### Mistral API

Mistral Vibe uses the Mistral AI API:

#### Endpoint: POST /v1/chat/completions

**Description:** Create a chat completion.

**Request:**
```json
{
  "model": "codestral-latest",
  "messages": [
    {"role": "system", "content": "You are a helpful assistant"},
    {"role": "user", "content": "Write a Python function"}
  ],
  "temperature": 0.7,
  "max_tokens": 4096
}
```

**Response:**
```json
{
  "id": "chatcmpl-xxx",
  "object": "chat.completion",
  "created": 1234567890,
  "model": "codestral-latest",
  "choices": [{
    "index": 0,
    "message": {
      "role": "assistant",
      "content": "Here's the function..."
    },
    "finish_reason": "stop"
  }],
  "usage": {
    "prompt_tokens": 15,
    "completion_tokens": 100,
    "total_tokens": 115
  }
}
```

### Supported Models

| Model | Description | Best For |
|-------|-------------|----------|
| `codestral-latest` | Latest Codestral model | Code generation |
| `mistral-large-latest` | Most capable model | Complex tasks |
| `mistral-small-latest` | Fast and efficient | Quick tasks |
| `pixtral-large-latest` | Vision-capable | Image + code tasks |

---

## Usage Examples

### Example 1: Interactive Coding

```bash
# Start Mistral Vibe
cd my-project
mistral-vibe

# In the TUI:
> Explain the structure of this project
> Create a function to handle user authentication
> Refactor the database layer to use async/await
```

### Example 2: Quick Code Generation

```bash
# Generate code without entering TUI
mistral-vibe "Create a Python class for a blog post with title, content, and author"

# With specific model
mistral-vibe "Write a React component for a todo list" --model codestral-latest
```

### Example 3: Code Review

```bash
# Review specific files
mistral-vibe "Review src/auth.js for security issues"

# Review entire codebase
mistral-vibe "Analyze this codebase for best practices"
```

### Example 4: Custom Skills

```bash
# Create a skill for your project
mkdir -p .mistral/skills/project-conventions
cat > .mistral/skills/project-conventions/SKILL.md << 'EOF'
---
description: Project conventions
---

# Our Project Conventions

- Use TypeScript for all new files
- Follow the existing folder structure
- Use pnpm for package management
- Write tests with Vitest
EOF

# Start vibe - skill will be automatically loaded
mistral-vibe
```

### Example 5: Working with Skills

```bash
# List available skills
mistral-vibe
> /skills list

# Enable a skill
> /skills enable project-conventions

# Use skill in conversation
> Create a new component following our project conventions
```

### Example 6: Switching Models

```bash
mistral-vibe

# In the TUI:
> /model
# (Select from available models)

# Or directly:
> /model mistral-large-latest
```

---

## Troubleshooting

### Issue: API key not set

**Solution:**
```bash
export MISTRAL_API_KEY="your-api-key"

# Or add to shell profile
echo 'export MISTRAL_API_KEY="your-api-key"' >> ~/.bashrc
```

### Issue: Command not found after installation

**Solution:**
```bash
# Ensure local bin is in PATH
export PATH="$HOME/.local/bin:$PATH"

# For uv installations
export PATH="$HOME/.cargo/bin:$PATH"
```

### Issue: Model not available

**Solution:**
```bash
# List available models
mistral-vibe
> /models

# Use valid model name
mistral-vibe --model codestral-latest
```

### Issue: Skills not loading

**Solution:**
1. Check skill location: `~/.mistral/skills/` or `.mistral/skills/`
2. Verify SKILL.md format
3. Check skill file permissions
4. Restart Mistral Vibe

### Issue: Session not saving

**Solution:**
```bash
# Check write permissions
ls -la ~/.local/share/mistral/

# Create directory if needed
mkdir -p ~/.local/share/mistral/sessions
```

### Issue: Windows compatibility

**Note:** Mistral Vibe works on Windows but officially targets UNIX environments.

**Solution:** Use WSL2 for best experience:
```bash
# In WSL2
curl -LsSf https://mistral.ai/vibe/install.sh | bash
```

### Issue: Trust folder system

Mistral Vibe uses a trust folder system for security.

**Solution:**
```bash
# Trust current directory
mistral-vibe
> /trust add .

# Or trust specific path
> /trust add /path/to/project
```

---

## Additional Resources

- **Mistral AI Platform:** https://console.mistral.ai
- **Documentation:** https://docs.mistral.ai
- **GitHub:** https://github.com/mistralai/mistral-vibe
- **Models:** https://docs.mistral.ai/getting-started/models
