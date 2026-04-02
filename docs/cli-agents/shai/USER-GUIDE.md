# Shai User Guide

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
pip install shell-ai
```

### Method 2: Package Manager (PIPX)

```bash
pipx install shell-ai
```

### Method 3: Build from Source

```bash
git clone https://github.com/ricklamers/shell-ai.git
cd shell-ai
pip install -e .
```

---

## Quick Start

```bash
# Set API key
export OPENAI_API_KEY="your-api-key-here"

# Verify installation
shai --help

# Run Shai
shai run list all files in current directory

# Pipe input
echo "sort files by size" | shai run
```

---

## CLI Commands

### Global Options

| Option | Short | Description | Example |
|--------|-------|-------------|---------|
| `--help` | `-h` | Show help | `shai --help` |
| `--version` | `-v` | Show version | `shai --version` |

### Command: shai run

**Description:** Convert natural language to shell commands.

**Usage:**
```bash
shai run [description...]
```

**Options:**
| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `description` | string | Yes | - | Natural language description |

**Examples:**
```bash
# Basic usage
shai run list all files in the current directory

# Find files
shai run find all Python files modified in the last week

# Complex commands
shai run compress all log files older than 30 days

# Terraform example
shai run terraform dry run thingy

# Output:
# terraform plan
# terraform plan -input=false
# terraform plan
```

**Exit Codes:**
- `0` - Success
- `1` - General error

---

## TUI/Interactive Commands

Shai operates in inline mode and presents command suggestions interactively:

| Action | Description |
|--------|-------------|
| Select command | Use arrow keys to navigate suggestions |
| Execute | Press Enter to execute selected command |
| Cancel | Press Ctrl+C to cancel |

---

## Configuration

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `OPENAI_API_KEY` | OpenAI API key | Yes |
| `OPENAI_API_BASE` | Custom API base URL | No |
| `SHAI_SUGGESTION_COUNT` | Number of suggestions (default: 3) | No |
| `SHAI_MODEL` | Model to use | No |

### Azure OpenAI Configuration

For Azure OpenAI deployments:

```bash
export OPENAI_API_TYPE="azure"
export OPENAI_API_BASE="https://your-resource.openai.azure.com/"
export OPENAI_API_VERSION="2023-05-15"
export OPENAI_API_KEY="your-azure-api-key"
```

### Configuration File

**File Location:** `~/.config/shai/config.json`

```json
{
  "provider": "openai",
  "model": "gpt-3.5-turbo",
  "suggestion_count": 3,
  "auto_execute": false
}
```

---

## API/Protocol Endpoints

### OpenAI API

Shai uses the OpenAI API for command generation:

#### Endpoint: POST /v1/chat/completions

**Description:** Generate command suggestions.

**Request:**
```json
{
  "model": "gpt-3.5-turbo",
  "messages": [
    {
      "role": "system",
      "content": "Convert natural language to shell commands. Provide 3 alternatives."
    },
    {
      "role": "user",
      "content": "list all files in current directory"
    }
  ],
  "temperature": 0.3
}
```

**Response:**
```json
{
  "choices": [{
    "message": {
      "content": "1. ls -la\n2. ls -lah\n3. find . -maxdepth 1 -type f -ls"
    }
  }]
}
```

---

## Usage Examples

### Example 1: Basic Command Generation

```bash
# Set API key
export OPENAI_API_KEY="sk-..."

# Generate commands
shai run show me disk usage
# Output:
# df -h
# du -sh *
# ncdu

# Select and execute
```

### Example 2: Git Commands

```bash
shai run show git log with graph
# Output:
# git log --graph --oneline
# git log --graph --decorate --oneline
# git log --graph --all --oneline

shai run undo last git commit
# Output:
# git reset --soft HEAD~1
# git reset --mixed HEAD~1
# git reset --hard HEAD~1
```

### Example 3: File Operations

```bash
shai run find large files
# Output:
# find . -type f -size +100M
# du -ah . | sort -rh | head -20
# find . -type f -exec ls -lh {} \; | sort -k5 -rh

shai run delete old log files
# Output:
# find . -name "*.log" -mtime +30 -delete
# find /var/log -name "*.log" -type f -mtime +30 -exec rm {} \;
```

### Example 4: System Administration

```bash
shai run check memory usage
# Output:
# free -h
# vmstat -s
# cat /proc/meminfo

shai run list running processes
# Output:
# ps aux
# top
# htop
```

### Example 5: Docker Commands

```bash
shai run list docker containers
# Output:
# docker ps
# docker ps -a
# docker container ls

shai run clean up docker
# Output:
# docker system prune
# docker volume prune
# docker image prune -a
```

### Example 6: Piped Input

```bash
# Pipe description to shai
echo "sort files by modification time" | shai run

# Use with other commands
cat todo.txt | shai run "create shell commands from these tasks"
```

---

## Troubleshooting

### Issue: API key not set

**Solution:**
```bash
export OPENAI_API_KEY="sk-..."

# Add to shell profile
echo 'export OPENAI_API_KEY="sk-..."' >> ~/.bashrc
```

### Issue: Command not found after installation

**Solution:**
```bash
# Check pip installation location
which shai
pip show shell-ai

# Ensure local bin is in PATH
export PATH="$HOME/.local/bin:$PATH"
```

### Issue: Permission denied (Linux Python 3.10+)

**Solution:**
```bash
# Use pipx instead of pip
pipx install shell-ai

# Or use --user flag
pip install --user shell-ai
```

### Issue: Azure OpenAI not working

**Solution:**
```bash
# Set all required Azure variables
export OPENAI_API_TYPE="azure"
export OPENAI_API_BASE="https://your-resource.openai.azure.com/"
export OPENAI_API_VERSION="2023-05-15"
export OPENAI_API_KEY="your-key"
```

### Issue: Too many/few suggestions

**Solution:**
```bash
# Set suggestion count
export SHAI_SUGGESTION_COUNT=5
```

### Issue: Suggestions are not relevant

**Solution:**
- Be more specific in your description
- Include context (e.g., "in this directory", "for this project")
- Try different phrasing

---

## Additional Resources

- **GitHub Repository:** https://github.com/ricklamers/shell-ai
- **PyPI Package:** https://pypi.org/project/shell-ai/
- **InquirerPy:** https://github.com/kazhala/InquirerPy
