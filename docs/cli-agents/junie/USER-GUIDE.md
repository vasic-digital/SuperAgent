# Junie User Guide

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

### Method 1: Install Script (Recommended)

**Linux / macOS:**
```bash
curl -fsSL https://junie.jetbrains.com/install.sh | bash
```

### Method 2: Homebrew

```bash
brew install junie
```

### Method 3: Direct Download

Download from [junie.jetbrains.com](https://junie.jetbrains.com) and extract to a directory in your PATH.

---

## Quick Start

```bash
# Verify installation
junie --version
junie --help

# Authenticate (supports GitHub, Google, AWS Builder ID, IAM Identity Center)
junie login

# Start interactive session
junie

# Headless mode
junie "Review and fix code quality issues"
```

---

## CLI Commands

### Global Options

| Option | Short | Description | Example |
|--------|-------|-------------|---------|
| `--help` | `-h` | Show help | `junie --help` |
| `--version` | `-v` | Show version | `junie --version` |

### Command: junie

**Description:** Start the interactive Junie session or run with a prompt.

**Usage:**
```bash
junie [prompt] [options]
```

**Options:**
| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `prompt` | string | No | - | Prompt for headless mode |

**Examples:**
```bash
# Interactive mode
junie

# Headless mode with prompt
junie "Explain this codebase"
junie "Refactor authentication module"
```

**Exit Codes:**
- `0` - Success
- `1` - General error
- `130` - Interrupted (Ctrl+C)

### Command: junie login

**Description:** Authenticate with your identity provider.

**Usage:**
```bash
junie login [options]
```

**Options:**
| Option | Type | Description |
|--------|------|-------------|
| `--provider` | string | Identity provider (github, google, aws) |

**Examples:**
```bash
junie login
junie login --provider github
```

### Command: junie chat

**Description:** Start an interactive chat session.

**Usage:**
```bash
junie chat [options]
```

**Options:**
| Option | Type | Description |
|--------|------|-------------|
| `--resume` | flag | Resume last session |

**Examples:**
```bash
junie chat
junie chat --resume
```

### Command: junie --account

**Description:** Manage account and credentials.

**Usage:**
```bash
junie --account
```

**Examples:**
```bash
# In interactive mode:
> /account
```

---

## TUI/Interactive Commands

Once inside the Junie CLI, use these slash commands:

| Command | Description | Example |
|---------|-------------|---------|
| `/help` | Show help | `/help` |
| `/account` | Manage credentials | `/account` |
| `/clear` | Clear conversation | `/clear` |
| `/exit` | Exit session | `/exit` |

### Context Attachment

Use `@` to attach files or folders to your request:

```
> Explain the authentication in @src/auth.js
> Review @tests/ for coverage gaps
> Compare @src/old with @src/new
```

---

## Configuration

### Authentication

Junie supports multiple authentication methods:
- JetBrains Account
- GitHub
- Google
- AWS Builder ID
- IAM Identity Center

### OpenRouter Integration

Configure Junie to use OpenRouter for model access:

```bash
# Set environment variable
export JUNIE_OPENROUTER_API_KEY="sk-or-v1-..."

# Or add to shell profile
echo 'export JUNIE_OPENROUTER_API_KEY="sk-or-v1-..."' >> ~/.bashrc
```

Benefits:
- Access to hundreds of models
- Provider failover
- Centralized billing
- Team controls

### Environment Variables

| Variable | Description |
|----------|-------------|
| `JUNIE_OPENROUTER_API_KEY` | OpenRouter API key |
| `JUNIE_API_KEY` | JetBrains API key |

### Configuration File

**File Location:** `~/.config/junie/config.json`

```json
{
  "auth": {
    "provider": "openrouter",
    "api_key": "sk-or-v1-..."
  },
  "models": {
    "default": "anthropic/claude-sonnet-4",
    "fallback": "openai/gpt-4o"
  },
  "features": {
    "auto_execute": false,
    "brave_mode": false
  }
}
```

### Guidelines and Skills

Junie supports customization via:
- Guidelines - Project-specific instructions
- Custom agents - Specialized behavior
- Skills - Reusable capabilities
- MCP (Model Context Protocol) - External tool integration

---

## API/Protocol Endpoints

### OpenRouter API

When using OpenRouter integration:

#### Endpoint: POST /api/v1/chat/completions

**Description:** Create a chat completion.

**Request:**
```json
{
  "model": "anthropic/claude-sonnet-4",
  "messages": [
    {"role": "user", "content": "Write a Python function"}
  ]
}
```

**Response:**
```json
{
  "choices": [{
    "message": {
      "content": "Here's the function..."
    }
  }]
}
```

### Supported Models

Junie supports models from:
- OpenAI (GPT-4, GPT-3.5)
- Anthropic (Claude 3, Claude 2)
- Google (Gemini)
- xAI (Grok)
- Meta (Llama)
- And more via OpenRouter

---

## Usage Examples

### Example 1: Basic Setup with OpenRouter

```bash
# Install Junie
curl -fsSL https://junie.jetbrains.com/install.sh | bash

# Set OpenRouter API key
export JUNIE_OPENROUTER_API_KEY="sk-or-v1-..."

# Start Junie
junie

# In the TUI:
> Give me an overview of this codebase
> Create a function to handle user authentication
```

### Example 2: Headless Mode for CI/CD

```bash
# Set API key
export JUNIE_OPENROUTER_API_KEY="$OPENROUTER_API_KEY"

# Run in headless mode
junie "Review and fix any code quality issues in the latest commit"

# Or with specific prompt
junie "Run tests and fix any failures"
```

### Example 3: GitHub Actions Integration

```yaml
# .github/workflows/junie-review.yml
name: Code Review
on:
  pull_request:
    types: [opened, synchronize]

jobs:
  review:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      pull-requests: write
      issues: write
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 1
      
      - uses: JetBrains/junie-github-action@v0
        with:
          openrouter_api_key: ${{ secrets.OPENROUTER_API_KEY }}
          prompt: "code-review"
```

### Example 4: Code Analysis

```bash
cd my-project
junie

# In the TUI:
> What does this project do?
> Explain the architecture
> Find potential bugs in @src/
> Suggest improvements for @src/utils.js
```

### Example 5: Refactoring

```bash
junie

# In the TUI:
> Refactor the authentication system to use JWT
> @src/auth.js

# Review changes and approve
> Apply the changes
```

### Example 6: Test Generation

```bash
junie

# In the TUI:
> Write unit tests for @src/calculator.js
> Ensure 100% coverage
> Use Jest testing framework
```

### Example 7: Documentation

```bash
junie

# In the TUI:
> Generate API documentation for @src/routes/
> Create a README for this project
> Document the database schema
```

---

## Troubleshooting

### Issue: Authentication fails

**Solution:**
```bash
# Re-authenticate
junie login

# Check environment variable
env | grep JUNIE

# Verify API key is valid
```

### Issue: Model not available

**Solution:**
```bash
# Check OpenRouter dashboard for available models
# Update Junie to latest version
# Verify API key has access to requested model
```

### Issue: Headless mode hangs

**Solution:**
```bash
# Ensure prompt is specific
junie "Fix the typo in README.md"

# Check logs for errors
junie --verbose "Run tests"
```

### Issue: Context not loading

**Solution:**
```bash
# Verify file paths
junie
> Explain @src/main.js

# Use absolute paths if needed
> Review @/home/user/project/src/
```

### Issue: GitHub Actions fails

**Solution:**
- Verify `OPENROUTER_API_KEY` secret is set
- Check permissions in workflow
- Ensure fetch-depth is sufficient

---

## Additional Resources

- **Junie Website:** https://junie.jetbrains.com
- **Documentation:** https://junie.jetbrains.com/docs
- **OpenRouter:** https://openrouter.ai
- **JetBrains:** https://jetbrains.com
