# Ollama Code User Guide

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

### Prerequisites

Before installing Ollama Code, you need Ollama running locally:

```bash
# Install Ollama
# macOS/Linux:
curl -fsSL https://ollama.com/install.sh | sh

# Or use Homebrew (macOS)
brew install ollama

# Windows: Download from https://ollama.com/download
```

### Method 1: Package Manager (NPM)

```bash
npm install -g ollama-code
```

### Method 2: Package Manager (PIP)

```bash
pip install ollama-code-cli
```

### Method 3: Build from Source

```bash
git clone https://github.com/tcsenpai/ollama-code.git
cd ollama-code
npm install
npm run build
npm install -g .
```

### macOS Installation Notes

If you see "externally-managed-environment" error:

**Option 1: Use pipx**
```bash
brew install pipx
pipx ensurepath
pipx install ollama-code-cli
```

**Option 2: Virtual environment**
```bash
python3 -m venv ~/.venvs/ollama-code
source ~/.venvs/ollama-code/bin/activate
pip install ollama-code-cli
```

**Option 3: Use uv**
```bash
brew install uv
uv tool install ollama-code-cli
```

---

## Quick Start

```bash
# Start Ollama server
ollama serve

# Pull a coding model
ollama pull qwen3:4b
ollama pull qwen2.5:3b
ollama pull codellama:7b

# Start Ollama Code
ollama-code

# Or with specific model
ollama-code --model qwen3:4b
```

---

## CLI Commands

### Global Options

| Option | Short | Description | Example |
|--------|-------|-------------|---------|
| `--help` | `-h` | Show help | `ollama-code --help` |
| `--version` | `-v` | Show version | `ollama-code --version` |
| `--model` | `-m` | Model to use | `ollama-code -m qwen3:4b` |
| `--host` | | Ollama host | `ollama-code --host localhost:11434` |

### Command: ollama-code

**Description:** Start the interactive Ollama Code session.

**Usage:**
```bash
ollama-code [options]
```

**Options:**
| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `--model` | string | No | default | Model to use |
| `--host` | string | No | localhost:11434 | Ollama server host |

**Examples:**
```bash
# Interactive mode with default model
ollama-code

# With specific model
ollama-code --model qwen3:4b

# With custom Ollama host
ollama-code --host 192.168.1.100:11434
```

**Exit Codes:**
- `0` - Success
- `1` - General error
- `130` - Interrupted (Ctrl+C)

### Command: ollama-code (Inline Mode)

**Description:** Run a single query and exit.

**Usage:**
```bash
ollama-code "Your prompt here" [options]
```

**Examples:**
```bash
ollama-code "Explain what this project does"
ollama-code "Create a Python function to sort a list" --model codellama
```

---

## TUI/Interactive Commands

Once inside the Ollama Code TUI, use these commands:

| Command | Description | Example |
|---------|-------------|---------|
| `/help` | Show help | `/help` |
| `/model` | List or switch models | `/model` or `/model qwen3:4b` |
| `/clear` | Clear conversation | `/clear` |
| `/exit` | Exit the application | `/exit` |

---

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `OLLAMA_HOST` | Ollama server host | `localhost:11434` |
| `OLLAMA_MODEL` | Default model to use | - |

### Configuration File

**File Location:** `~/.config/ollama-code/config.json`

```json
{
  "ollama": {
    "host": "localhost:11434",
    "default_model": "qwen3:4b"
  },
  "ui": {
    "theme": "dark",
    "show_thinking": true
  },
  "tools": {
    "enabled": true,
    "auto_execute": false
  }
}
```

### Available Tools

Ollama Code supports tool calling with compatible models:

| Tool | Description |
|------|-------------|
| `file_read` | Read file contents |
| `file_write` | Write to files |
| `command_execute` | Execute shell commands |
| `directory_list` | List directory contents |

### MCP Integration

Ollama Code supports MCP (Model Context Protocol) servers:

```json
{
  "mcp": {
    "servers": {
      "filesystem": {
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/allowed/dir"]
      }
    }
  }
}
```

---

## API/Protocol Endpoints

### Ollama API

Ollama Code uses the Ollama API for local inference:

#### Endpoint: POST /api/chat

**Description:** Generate a chat completion.

**Request:**
```json
{
  "model": "qwen3:4b",
  "messages": [
    {"role": "system", "content": "You are a helpful coding assistant"},
    {"role": "user", "content": "Write a Python function"}
  ],
  "stream": true,
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "file_read",
        "description": "Read a file",
        "parameters": {
          "type": "object",
          "properties": {
            "path": {"type": "string"}
          }
        }
      }
    }
  ]
}
```

**Response:**
```json
{
  "model": "qwen3:4b",
  "created_at": "2024-01-01T00:00:00Z",
  "message": {
    "role": "assistant",
    "content": "Here's the function..."
  },
  "done": true
}
```

#### Endpoint: POST /api/generate

**Description:** Generate a completion.

**Request:**
```json
{
  "model": "codellama:7b",
  "prompt": "Write a Python function to calculate factorial:",
  "stream": false
}
```

**Response:**
```json
{
  "model": "codellama:7b",
  "response": "def factorial(n):\n    if n <= 1:\n        return 1\n    return n * factorial(n - 1)",
  "done": true
}
```

### Recommended Models

| Model | Size | Best For |
|-------|------|----------|
| `qwen3:4b` | 4B | General coding |
| `qwen2.5:3b` | 3B | Fast responses |
| `qwen2.5-coder:7b` | 7B | Code generation |
| `codellama:7b` | 7B | Code completion |
| `codellama:13b` | 13B | Complex tasks |
| `deepseek-coder:6.7b` | 6.7B | Code-specific |
| `llama3.1:8b` | 8B | General purpose |

---

## Usage Examples

### Example 1: Basic Setup

```bash
# Start Ollama
ollama serve

# Pull a model
ollama pull qwen3:4b

# Test the model
ollama run qwen3:4b "Write a hello world in Python"

# Start Ollama Code
ollama-code

# In the TUI:
> Explain what this project does
> Create a function to parse JSON
```

### Example 2: Code Generation

```bash
# Start with coding model
ollama-code --model qwen2.5-coder:7b

# In the TUI:
> Create a React component for a todo list with TypeScript
> Add local storage persistence
> Style it with Tailwind CSS
```

### Example 3: Code Review

```bash
cd my-project
ollama-code

# In the TUI:
> Review the src/auth.js file for security issues
> Find potential bugs in the database layer
> Suggest improvements for the API design
```

### Example 4: Project Analysis

```bash
ollama-code

# In the TUI:
> What is the architecture of this project?
> List all the dependencies and their purposes
> Identify areas that could benefit from refactoring
```

### Example 5: Learning a Codebase

```bash
cd unfamiliar-project
ollama-code --model qwen3:4b

# In the TUI:
> Give me an overview of this codebase
> What are the main entry points?
> How does the authentication work?
> Explain the data flow from UI to database
```

### Example 6: Tool Calling

With models that support tool calling (Qwen3, Qwen2.5):

```bash
ollama-code --model qwen3:4b

# In the TUI:
> Read the package.json and explain the project dependencies
# (AI uses file_read tool automatically)

> List all JavaScript files in the src directory
# (AI uses directory_list tool)
```

### Example 7: Refactoring

```bash
ollama-code

# In the TUI:
> Refactor the User class to use TypeScript interfaces
> Convert callback-based code to use async/await
> Extract the validation logic into a separate module
```

---

## Troubleshooting

### Issue: Ollama server not running

**Solution:**
```bash
# Start Ollama
ollama serve

# Or as a background service
# macOS: Ollama runs automatically after installation
# Linux: sudo systemctl start ollama
```

### Issue: Model not found

**Solution:**
```bash
# List available models
ollama list

# Pull the model you need
ollama pull qwen3:4b
ollama pull qwen2.5-coder:7b
```

### Issue: Connection refused

**Solution:**
```bash
# Check Ollama is running
curl http://localhost:11434/api/tags

# Verify OLLAMA_HOST
export OLLAMA_HOST=localhost:11434

# Check firewall settings
```

### Issue: Out of memory

**Solution:**
- Use a smaller model (3B instead of 7B)
- Close other applications
- Increase system RAM
- Use a machine with GPU

### Issue: Slow responses

**Solution:**
- Use a smaller/faster model
- Ensure GPU is being used (check `ollama ps`)
- Close unnecessary applications
- Use a quantized model (Q4_K_M)

### Issue: Tool calling not working

**Solution:**
- Ensure model supports tool calling (Qwen3, Qwen2.5)
- Check tools are enabled in config
- Update to latest Ollama version

### Issue: Import errors (Python)

**Solution:**
```bash
# Use virtual environment
python3 -m venv ~/.venvs/ollama-code
source ~/.venvs/ollama-code/bin/activate
pip install ollama-code-cli
```

### Issue: VS Code extension integration

**Solution:**
1. Install Continue extension in VS Code
2. Configure Continue to use Ollama:

```json
{
  "models": [
    {
      "title": "Ollama",
      "provider": "ollama",
      "model": "qwen2.5-coder:7b"
    }
  ]
}
```

---

## Additional Resources

- **Ollama Website:** https://ollama.ai
- **Ollama Documentation:** https://github.com/ollama/ollama/tree/main/docs
- **Ollama Code GitHub:** https://github.com/tcsenpai/ollama-code
- **Model Library:** https://ollama.ai/library
