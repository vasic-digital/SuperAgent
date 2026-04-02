# Gemini CLI User Guide

## Table of Contents
1. [Installation](#installation)
2. [Quick Start](#quick-start)
3. [CLI Commands](#cli-commands)
4. [TUI/Interactive Commands](#tui-interactive-commands)
5. [Configuration](#configuration)
6. [Usage Examples](#usage-examples)
7. [Troubleshooting](#troubleshooting)

## Installation

### Method 1: npm
```bash
npm install -g @google/genai-cli
```

### Method 2: Direct Download
```bash
curl -fsSL https://dl.google.com/genai-cli/install.sh | bash
```

### Method 3: Build from Source
```bash
git clone https://github.com/google/genai-cli.git
cd genai-cli
npm install
npm run build
npm link
```

## Quick Start

```bash
# Authenticate
gemini auth login

# Start interactive mode
gemini chat

# Single prompt
gemini ask "Explain quantum computing"

# With file
gemini ask --file code.py "Review this code"
```

## CLI Commands

### Global Options
| Option | Short | Description | Example |
|--------|-------|-------------|---------|
| --help | -h | Show help | `gemini --help` |
| --version | -v | Show version | `gemini --version` |
| --api-key | | API key | `--api-key YOUR_KEY` |
| --model | -m | Model to use | `--model gemini-pro` |

### Command: auth
**Description:** Authentication commands

**Subcommands:**
```bash
gemini auth login      # Login with Google
gemini auth logout     # Logout
gemini auth status     # Check auth status
```

### Command: chat
**Description:** Start interactive chat

**Usage:**
```bash
gemini chat [options]
```

**Options:**
| Option | Description | Default |
|--------|-------------|---------|
| --model | Model to use | gemini-pro |
| --system | System prompt | "" |
| --temperature | Temperature | 0.7 |

### Command: ask
**Description:** Single-turn query

**Usage:**
```bash
gemini ask "Your question here"
gemini ask --file document.txt "Summarize this"
```

**Options:**
| Option | Description |
|--------|-------------|
| --file | File to include |
| --model | Model selection |
| --output | Output file |

### Command: models
**Description:** List available models

**Usage:**
```bash
gemini models list
```

## TUI/Interactive Commands

In interactive chat mode:

| Command | Shortcut | Description |
|---------|----------|-------------|
| /help | | Show commands |
| /exit | Ctrl+D | Exit chat |
| /clear | | Clear context |
| /file | | Add file to context |
| /model | | Change model |
| /temp | | Set temperature |

## Configuration

### Configuration File Format (JSON)

```json
{
  "defaultModel": "gemini-pro",
  "temperature": 0.7,
  "maxOutputTokens": 2048,
  "safetySettings": {
    "harmCategory": "HARM_CATEGORY_DANGEROUS_CONTENT",
    "threshold": "BLOCK_ONLY_HIGH"
  }
}
```

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| GEMINI_API_KEY | API key | Yes |
| GOOGLE_API_KEY | Alternative API key | Yes |
| GEMINI_MODEL | Default model | No |

### Configuration Locations
1. `~/.config/gemini/config.json`
2. Project `.gemini.json`
3. Environment variables

## Usage Examples

### Example 1: Code Explanation
```bash
gemini ask --file script.js "Explain what this code does"
```

### Example 2: Interactive Session
```bash
gemini chat --model gemini-pro
> Help me write a Python function
> Add error handling to it
> Write tests for it
```

### Example 3: Batch Processing
```bash
for file in *.py; do
  gemini ask --file "$file" "Generate docstrings" >> docs.txt
done
```

### Example 4: With System Prompt
```bash
gemini chat --system "You are a code reviewer. Be concise and critical."
```

## Troubleshooting

### Issue: Authentication Failed
**Solution:**
```bash
gemini auth logout
gemini auth login
```

### Issue: API Key Not Found
**Solution:**
```bash
export GEMINI_API_KEY=your-key
# Or create ~/.config/gemini/config.json
```

### Issue: Model Not Available
**Solution:**
```bash
gemini models list  # Check available models
gemini ask --model gemini-pro "Hello"
```

---

**Last Updated:** 2026-04-02
