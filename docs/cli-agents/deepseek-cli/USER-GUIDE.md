# DeepSeek CLI User Guide

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

```bash
npm install -g deepseek-cli
```

### Method 2: Package Manager (PIP)

```bash
pip install deepseek-cli
```

### Method 3: Build from Source

```bash
git clone https://github.com/PierrunoYT/deepseek-cli.git
cd deepseek-cli
pip install -e .
```

---

## Quick Start

```bash
# Set API key
export DEEPSEEK_API_KEY="your-api-key-here"

# Verify installation
deepseek --version
deepseek --help

# Interactive mode
deepseek

# Inline mode with query
deepseek -q "What is the capital of France?"

# Read from file
deepseek --read prompt.txt

# Pipe input
echo "Explain quantum computing" | deepseek --read -
```

---

## CLI Commands

### Global Options

| Option | Short | Description | Example |
|--------|-------|-------------|---------|
| `--help` | `-h` | Show help | `deepseek --help` |
| `--version` | `-v` | Show version | `deepseek --version` |
| `--query` | `-q` | Run inline mode with query | `deepseek -q "Hello"` |
| `--read` | | Read query from file | `deepseek --read file.txt` |
| `--model` | `-m` | Model to use | `deepseek -m deepseek-coder` |
| `--raw` | `-r` | Output without token info | `deepseek -q "Hi" -r` |
| `--system` | `-S` | Set system message | `deepseek -S "Be helpful"` |
| `--stream` | `-s` | Enable streaming | `deepseek --stream` |
| `--no-stream` | | Disable streaming | `deepseek --no-stream` |
| `--json` | | Enable JSON output | `deepseek --json` |
| `--beta` | | Enable beta API | `deepseek --beta` |
| `--prefix` | | Enable prefix completion | `deepseek --prefix` |
| `--fim` | | Enable Fill-in-Middle | `deepseek --fim` |
| `--multiline` | | Enable multiline input | `deepseek --multiline` |
| `--temp` | | Set temperature (0-2) | `deepseek --temp 0.7` |
| `--freq` | | Frequency penalty (-2 to 2) | `deepseek --freq 0.5` |
| `--pres` | | Presence penalty (-2 to 2) | `deepseek --pres 0.5` |
| `--top-p` | | Top-p sampling (0-1) | `deepseek --top-p 0.9` |
| `--stop` | | Add stop sequence | `deepseek --stop "END"` |

### Command: deepseek (Interactive Mode)

**Description:** Start the interactive REPL mode.

**Usage:**
```bash
deepseek [options]
```

**Options:**
| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `--model` | string | No | deepseek-chat | Model to use |
| `--system` | string | No | "You are a helpful assistant" | System message |
| `--multiline` | flag | No | false | Enable multiline input |
| `--multiline-submit` | string | No | empty-line | Submit mode (empty-line/shift-enter) |

**Examples:**
```bash
# Start interactive mode
deepseek

# Start with specific model
deepseek --model deepseek-coder

# Start with multiline input
deepseek --multiline

# Start with shift-enter submit
deepseek --multiline --multiline-submit shift-enter
```

**Exit Codes:**
- `0` - Success
- `1` - General error
- `130` - Interrupted (Ctrl+C)

### Command: deepseek (Inline Mode)

**Description:** Run a single query and exit.

**Usage:**
```bash
deepseek -q "Your query here" [options]
```

**Options:**
| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `-q, --query` | string | Yes | - | The query text |
| `-m, --model` | string | No | deepseek-chat | Model to use |
| `-r, --raw` | flag | No | false | Raw output without token info |
| `-S, --system` | string | No | - | System message |
| `--json` | flag | No | false | JSON output mode |
| `--temp` | float | No | 1.0 | Temperature (0-2) |

**Examples:**
```bash
# Basic query
deepseek -q "What is Python?"

# Specify model
deepseek -q "Write a function" -m deepseek-coder

# Raw output
deepseek -q "Hello" -r

# Custom system message
deepseek -S "You are a Rust expert" -q "Explain lifetimes"

# Set temperature
deepseek -q "Tell me a story" --temp 1.3

# Multiple stop sequences
deepseek -q "Count to five" --stop "5" --stop "five"

# JSON output
deepseek -q "List 3 European capitals" --json
```

### Command: deepseek --read

**Description:** Read query from file or stdin.

**Usage:**
```bash
deepseek --read FILE [options]
deepseek --read - [options]  # Read from stdin
```

**Options:**
| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `--read` | string | Yes | - | File path or `-` for stdin |
| `-q, --query` | string | No | - | Prefix query (prepended to file content) |

**Examples:**
```bash
# Read from file
deepseek --read prompt.txt

# Read from stdin
echo "Explain this" | deepseek --read -

# Combine with query
git diff HEAD | deepseek --read - -q "Review this diff:" -S "You are a code reviewer"
cat report.md | deepseek --read - -q "Summarize in one paragraph"
```

### Command: deepseek --fim

**Description:** Enable Fill-in-the-Middle mode for code completion.

**Usage:**
```bash
deepseek --fim -q "<fim_prefix>...<fim_suffix>..."
```

**Examples:**
```bash
# Fill-in-the-middle
deepseek --fim -q "def add(<fim_prefix>):<fim_suffix>    pass"
```

---

## TUI/Interactive Commands

Once inside the DeepSeek CLI interactive mode, use these slash commands:

| Command | Description | Example |
|---------|-------------|---------|
| `/help` | Show help message | `/help` |
| `/models` | List available models | `/models` |
| `/model X` | Switch model | `/model deepseek-coder` |
| `/system X` | Set system message | `/system You are an expert` |
| `/system` | Show current system message | `/system` |
| `/clear` | Clear conversation history | `/clear` |
| `/history` | Display conversation history | `/history` |
| `/about` | Show API information | `/about` |
| `/balance` | Check account balance | `/balance` |
| `/multiline` | Show multiline mode info | `/multiline` |
| `/temp X` | Set temperature | `/temp 0.7` |
| `/freq X` | Set frequency penalty | `/freq 0.5` |
| `/pres X` | Set presence penalty | `/pres 0.5` |
| `/top_p X` | Set top_p sampling | `/top_p 0.9` |
| `/beta` | Toggle beta features | `/beta` |
| `/prefix` | Toggle prefix completion | `/prefix` |
| `/fim` | Toggle Fill-in-Middle | `/fim` |
| `/cache` | Toggle context caching | `/cache` |
| `/json` | Toggle JSON output | `/json` |
| `/stream` | Toggle streaming | `/stream` |
| `/stop X` | Add stop sequence | `/stop END` |
| `/clearstop` | Clear stop sequences | `/clearstop` |
| `/function {}` | Add function definition | `/function {"name": "test"}` |
| `/clearfuncs` | Clear registered functions | `/clearfuncs` |
| `/quit` | Exit the CLI | `/quit` |

---

## Configuration

### Configuration File Format

DeepSeek CLI supports configuration via environment variables.

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `DEEPSEEK_API_KEY` | DeepSeek API key | Yes |
| `DEEPSEEK_API_BASE` | Custom API base URL | No |
| `ANTHROPIC_BASE_URL` | For Claude Code compatibility | No |
| `ANTHROPIC_AUTH_TOKEN` | DeepSeek key for Claude Code | No |

### Claude Code Compatibility

DeepSeek API supports Anthropic API format for Claude Code integration:

```bash
# Install Claude Code
npm install -g @anthropic-ai/claude-code

# Configure environment
export ANTHROPIC_BASE_URL=https://api.deepseek.com/anthropic
export ANTHROPIC_AUTH_TOKEN=${DEEPSEEK_API_KEY}
export ANTHROPIC_MODEL=deepseek-chat
export ANTHROPIC_SMALL_FAST_MODEL=deepseek-chat

# Run Claude Code with DeepSeek
cd my-project
claude
```

---

## API/Protocol Endpoints

### DeepSeek API

The CLI uses the DeepSeek API endpoints:

#### Endpoint: POST /chat/completions

**Description:** Create a chat completion.

**Request:**
```json
{
  "model": "deepseek-chat",
  "messages": [
    {"role": "system", "content": "You are a helpful assistant"},
    {"role": "user", "content": "Hello"}
  ],
  "temperature": 1.0,
  "max_tokens": 4096
}
```

**Response:**
```json
{
  "id": "chatcmpl-xxx",
  "object": "chat.completion",
  "created": 1234567890,
  "model": "deepseek-chat",
  "choices": [{
    "index": 0,
    "message": {
      "role": "assistant",
      "content": "Hello! How can I help you?"
    },
    "finish_reason": "stop"
  }],
  "usage": {
    "prompt_tokens": 10,
    "completion_tokens": 20,
    "total_tokens": 30
  }
}
```

### Supported Models

| Model | Context | Output | Features |
|-------|---------|--------|----------|
| `deepseek-chat` | 128K | 4K/8K | JSON, Function Calling, FIM |
| `deepseek-reasoner` | 128K | 32K/64K | Chain-of-Thought, JSON |
| `deepseek-coder` | 128K | 4K/8K | JSON, Function Calling, FIM |

### Model-Specific Features

**DeepSeek-V3.2 (deepseek-chat):**
- Context: 128K tokens
- Output: Default 4K, Max 8K
- Features: JSON Output, Function Calling (128 functions), Chat Prefix, FIM

**DeepSeek-V3.2 (deepseek-reasoner):**
- Context: 128K tokens
- Output: Default 32K, Max 64K
- Chain-of-Thought: Displays reasoning before answer
- Features: JSON Output, Chat Prefix
- Limitations: No Function Calling, No FIM, No temperature/top_p/penalties

---

## Usage Examples

### Example 1: Basic Chat

```bash
# Set API key
export DEEPSEEK_API_KEY="sk-..."

# Start interactive mode
deepseek

# Chat with the model
> What is machine learning?
> Write a Python script to sort a list
```

### Example 2: Code Generation

```bash
# Generate code with coder model
deepseek -q "Write a Python function to calculate factorial" -m deepseek-coder

# With custom system message
deepseek -S "You are a JavaScript expert" -q "Create a debounce function"

# With Fill-in-the-Middle
deepseek --fim -q "def fibonacci(<fim_prefix>):<fim_suffix>    pass"
```

### Example 3: Code Review

```bash
# Review git diff
git diff HEAD | deepseek --read - -q "Review this diff:" -S "You are a code reviewer"

# Review a file
deepseek --read mycode.py -q "Find bugs in this code"
```

### Example 4: Multiline Input

```bash
# Enable multiline mode
deepseek --multiline

# Then type:
# def calculate_sum(a, b):
#     return a + b
# 
# print(calculate_sum(2, 3))
# (Press Enter on blank line to submit)
```

### Example 5: Structured Output

```bash
# JSON output mode
deepseek -q "List 3 European capitals with their countries" --json

# With specific model
deepseek -m deepseek-chat -q "Generate a JSON schema for a user" --json
```

### Example 6: Interactive Session with Model Switching

```bash
deepseek

# In the session:
> /model deepseek-coder
> Write a React component

> /temp 0.2
> /model deepseek-chat
> Explain how React works
```

---

## Troubleshooting

### Issue: API key not recognized

**Solution:**
```bash
# Check environment variable
echo $DEEPSEEK_API_KEY

# Set it correctly
export DEEPSEEK_API_KEY="sk-..."

# Verify (Unix)
echo $DEEPSEEK_API_KEY

# Verify (Windows CMD)
echo %DEEPSEEK_API_KEY%

# Verify (Windows PowerShell)
echo $env:DEEPSEEK_API_KEY

# Try closing and reopening your terminal
```

### Issue: Import errors

**Solution:**
```bash
# Check installation
pip list | grep deepseek-cli

# Reinstall
pip install --force-reinstall deepseek-cli

# For development installation
pip install -e . --upgrade
```

### Issue: Streaming not working

**Solution:**
```bash
# Explicitly enable streaming
deepseek --stream

# Or disable if causing issues
deepseek --no-stream
```

### Issue: Model not found

**Solution:**
```bash
# List available models
deepseek
> /models

# Use correct model name
deepseek -m deepseek-chat  # or deepseek-coder, deepseek-reasoner
```

### Issue: Reasoner model limitations

**Cause:** DeepSeek-R1 (reasoner) has limitations.

**Solution:**
- Function Calling: Automatically falls back to deepseek-chat
- FIM: Not supported, use deepseek-coder instead
- Temperature/Penalties: Not supported for reasoning model

### Issue: Multiline mode not working

**Solution:**
```bash
# Use shift-enter submit for terminals that support it
deepseek --multiline --multiline-submit shift-enter

# Or use empty-line (default) - press Enter twice to submit
```

---

## Additional Resources

- **DeepSeek Platform:** https://platform.deepseek.com
- **API Documentation:** https://platform.deepseek.com/api-docs
- **GitHub Repository:** https://github.com/PierrunoYT/deepseek-cli
