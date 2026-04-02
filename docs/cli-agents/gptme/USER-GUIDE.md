# GPTMe User Guide

## Table of Contents
1. [Installation](#installation)
2. [Quick Start](#quick-start)
3. [CLI Commands](#cli-commands)
4. [TUI/Interactive Commands](#tui-interactive-commands)
5. [Configuration](#configuration)
6. [Usage Examples](#usage-examples)
7. [Troubleshooting](#troubleshooting)

## Installation

### Method 1: pip
```bash
pip install gptme
```

### Method 2: pipx
```bash
pipx install gptme
```

### Method 3: Source
```bash
git clone https://github.com/ErikBjare/gptme.git
cd gptme
pip install -e .
```

## Quick Start

```bash
# Set API key
export OPENAI_API_KEY=your-key

# Start interactive mode
gptme

# Run a command
gptme "Create a Python script that fetches weather data"

# Resume conversation
gptme --resume
```

## CLI Commands

### Global Options
| Option | Short | Description | Example |
|--------|-------|-------------|---------|
| --help | -h | Show help | `gptme --help` |
| --version | -v | Show version | `gptme --version` |
| --model | -m | Model to use | `--model gpt-4` |
| --resume | -r | Resume chat | `--resume` |
| --name | -n | Chat name | `--name myproject` |
| --no-confirm | -y | Auto-confirm | `--no-confirm` |

### Command: (default)
**Description:** Start chat or execute prompt

**Usage:**
```bash
gptme                    # Interactive mode
gptme "prompt"          # Execute and exit
gptme -r                # Resume last chat
gptme -n project "task" # Named chat
```

### Command: --resume
**Description:** Resume previous conversation

**Usage:**
```bash
gptme --resume
gptme -r -n projectname
```

### Command: --list
**Description:** List conversations

**Usage:**
```bash
gptme --list
gptme -l
```

### Command: --tools
**Description:** List available tools

**Usage:**
```bash
gptme --tools
```

## TUI/Interactive Commands

In interactive mode, these slash commands available:

| Command | Description |
|---------|-------------|
| /help | Show help |
| /exit | Exit gptme |
| /save | Save conversation |
| /clear | Clear screen |
| /undo | Undo last change |
| /tools | List tools |
| /model | Show/change model |

## Configuration

### Configuration File Format (TOML)

```toml
# ~/.config/gptme/config.toml
[openai]
api_key = "your-key"
model = "gpt-4"
temperature = 0.7

[gptme]
confirm = true
stream = true
tools = ["python", "shell", "save"]
```

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| OPENAI_API_KEY | OpenAI key | Yes |
| ANTHROPIC_API_KEY | Anthropic key | Alternative |
| GPTME_MODEL | Default model | No |

### Configuration Locations
1. `~/.config/gptme/config.toml`
2. Environment variables
3. Command-line flags

## Usage Examples

### Example 1: Interactive Coding
```bash
gptme
> Create a script to process CSV files
> Add error handling
> Save it as processor.py
```

### Example 2: One-shot Task
```bash
gptme "Write a bash script that backs up ~/Documents to S3"
```

### Example 3: Project Session
```bash
gptme -n myproject
# Work on project across multiple sessions
gptme -r -n myproject
```

### Example 4: Auto-confirm Mode
```bash
gptme -y "Update all dependencies"
# Executes without asking for confirmation
```

## Troubleshooting

### Issue: API Key Not Set
**Solution:**
```bash
export OPENAI_API_KEY=sk-...
```

### Issue: Cannot Resume Chat
**Solution:**
```bash
gptme --list  # See available chats
gptme -r -n exact_name
```

---

**Last Updated:** 2026-04-02
