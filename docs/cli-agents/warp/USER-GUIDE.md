# Warp User Guide

## Table of Contents
1. [Installation](#installation)
2. [Quick Start](#quick-start)
3. [CLI Commands](#cli-commands)
4. [TUI/Interactive Commands](#tui-interactive-commands)
5. [Configuration](#configuration)
6. [Usage Examples](#usage-examples)
7. [Troubleshooting](#troubleshooting)

## Installation

### Method 1: Homebrew
```bash
brew install --cask warp
```

### Method 2: Direct Download
Download from https://www.warp.dev/

### Method 3: Shell Integration
```bash
echo 'source ~/.warp/warp.sh' >> ~/.bashrc
```

## Quick Start

```bash
# Launch Warp
warp

# Or open in directory
warp .

# Use AI command
# Press # in Warp terminal, then type natural language
```

## CLI Commands

### Global Options
| Option | Description | Example |
|--------|-------------|---------|
| --help | Show help | `warp --help` |
| --version | Show version | `warp --version` |

### Command: (launch)
**Description:** Launch Warp terminal

**Usage:**
```bash
warp [directory]
```

### Command: ai
**Description:** AI command (in-app)

**Usage:**
```bash
# In Warp terminal, press:
# - # for AI command
# - Ctrl+` for AI chat
```

## TUI/Interactive Commands

In Warp terminal:

| Shortcut | Description |
|----------|-------------|
| # | AI command input |
| Ctrl+` | AI chat panel |
| Ctrl+L | Clear screen |
| Ctrl+Shift+C | Copy |
| Ctrl+Shift+V | Paste |
| Cmd+T | New tab |
| Cmd+W | Close tab |

### AI Command Prefix
Type `#` followed by natural language:
```
# List all docker containers
# Find files modified today
# Explain this error
```

## Configuration

### Settings (In-App)
Access via Settings menu or `~/.warp/settings.json`:

```json
{
  "theme": "Dark",
  "fontSize": 14,
  "ai": {
    "enabled": true,
    "provider": "openai",
    "apiKey": "your-key"
  }
}
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| WARP_API_KEY | Warp API key |
| OPENAI_API_KEY | OpenAI key for AI features |

## Usage Examples

### Example 1: AI Command
```bash
# In Warp terminal:
# list all processes using port 3000
```

### Example 2: AI Chat
```
Press Ctrl+`
Type: "How do I fix npm permission errors?"
```

### Example 3: Workflows
```
Press Cmd+Shift+W
Select or create a workflow
```

## Troubleshooting

### Issue: AI Not Working
**Solution:**
1. Check API key in settings
2. Verify internet connection
3. Restart Warp

### Issue: Shell Not Detected
**Solution:**
```bash
warp --shell bash
# or
warp --shell zsh
```

---

**Last Updated:** 2026-04-02
