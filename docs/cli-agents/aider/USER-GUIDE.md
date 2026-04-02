# Aider User Guide

## Table of Contents
1. [Installation](#installation)
2. [Quick Start](#quick-start)
3. [CLI Commands](#cli-commands)
4. [TUI/Interactive Commands](#tui-interactive-commands)
5. [Configuration](#configuration)
6. [Usage Examples](#usage-examples)
7. [Troubleshooting](#troubleshooting)

## Installation

### Method 1: pip (Recommended)
```bash
pip install aider-chat
```

### Method 2: pipx
```bash
pipx install aider-chat
```

### Method 3: Homebrew
```bash
brew install aider
```

## Quick Start

```bash
# Set API key
export OPENAI_API_KEY=your-key-here

# Start Aider with files
aider file1.py file2.py

# Start with specific model
aider --model gpt-4 file.py
```

## CLI Commands

### Global Options
| Option | Short | Description | Example |
|--------|-------|-------------|---------|
| --model | -m | Model to use | `aider --model gpt-4` |
| --config | | Config file path | `aider --config ~/.aider.conf.yml` |
| --help | -h | Show help | `aider --help` |
| --version | -v | Show version | `aider --version` |

### Command: /add
**Description:** Add files to the chat session

**Usage:**
```bash
/add <file1> [file2] ...
```

**Examples:**
```bash
/add src/main.py
/add src/*.py tests/*.py
```

### Command: /drop
**Description:** Remove files from the chat session

**Usage:**
```bash
/drop <file1> [file2] ...
```

### Command: /commit
**Description:** Commit changes with AI-generated message

**Usage:**
```bash
/commit [message]
```

### Command: /undo
**Description:** Undo last change

**Usage:**
```bash
/undo
```

### Command: /diff
**Description:** Show diff of changes

**Usage:**
```bash
/diff
```

### Command: /ls
**Description:** List files in chat session

**Usage:**
```bash
/ls
```

### Command: /voice
**Description:** Enter voice coding mode

**Usage:**
```bash
/voice
```

## TUI/Interactive Commands

When in interactive mode, use these slash commands:

| Command | Description | Example |
|---------|-------------|---------|
| /help | Show available commands | `/help` |
| /exit | Exit Aider | `/exit` |
| /add | Add files to context | `/add file.py` |
| /drop | Remove files from context | `/drop file.py` |
| /commit | Commit with AI message | `/commit` |
| /undo | Undo last change | `/undo` |
| /diff | Show changes | `/diff` |
| /ls | List tracked files | `/ls` |
| /voice | Voice coding mode | `/voice` |
| /run | Run shell command | `/run make test` |
| /test | Run test command | `/test pytest` |
| /map | Show repo map | `/map` |
| /tokens | Show token count | `/tokens` |
| /clear | Clear chat history | `/clear` |
| /reset | Reset chat history | `/reset` |

## Configuration

### Configuration File Format (YAML)

```yaml
# ~/.aider.conf.yml
model: gpt-4
auto-commits: true
git: true
pretty: true
stream: true
check-update: true
show-diffs: true

# API Keys
openai-api-key: sk-...
anthropic-api-key: sk-...
```

### Environment Variables

| Variable | Description | Required | Example |
|----------|-------------|----------|---------|
| OPENAI_API_KEY | OpenAI API key | For OpenAI | `sk-...` |
| ANTHROPIC_API_KEY | Anthropic API key | For Claude | `sk-...` |
| DEEPSEEK_API_KEY | DeepSeek API key | For DeepSeek | `sk-...` |
| AIDER_MODEL | Default model | No | `gpt-4` |

### Configuration Locations (precedence order)
1. Command-line arguments
2. Environment variables
3. `.aider.conf.yml` in current directory
4. `~/.aider.conf.yml`

## Usage Examples

### Example 1: Basic Code Editing
```bash
# Start Aider with a Python file
aider src/main.py

# Ask for changes
> Add error handling to the main function
```

### Example 2: Multi-file Refactoring
```bash
# Add multiple files
aider src/*.py

# Request refactoring
> Refactor all these files to use a common base class
```

### Example 3: Voice Coding
```bash
# Start Aider
aider

# Enter voice mode
/voice

# Speak your request
"Create a function to calculate factorial"
```

### Example 4: Test-Driven Development
```bash
# Add test and implementation files
aider tests/test_feature.py src/feature.py

# Ask to implement based on tests
> Implement the feature to pass all tests
```

### Example 5: Git Workflow
```bash
# Make changes with auto-commit
aider --auto-commits src/file.py

> Fix the bug in the parse function
# Changes committed automatically
```

## Troubleshooting

### Issue: API Key Not Found
**Symptoms:** Error about missing API key
**Solution:**
```bash
export OPENAI_API_KEY=your-key
# Or add to ~/.bashrc
```

### Issue: Model Not Available
**Symptoms:** Model not found error
**Solution:**
```bash
# Check available models
aider --list-models

# Use a valid model
aider --model gpt-4
```

### Issue: Git Not Initialized
**Symptoms:** Warning about git repository
**Solution:**
```bash
git init
git add .
git commit -m "Initial commit"
```

### Issue: Token Limit Exceeded
**Symptoms:** Context too long error
**Solution:**
- Use `/drop` to remove unnecessary files
- Use `/clear` to reset chat history
- Use a model with larger context

---

**Last Updated:** 2026-04-02
