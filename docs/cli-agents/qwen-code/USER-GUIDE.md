# Qwen Code User Guide

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

### Method 1: Quick Install (Recommended)

**Linux / macOS:**
```bash
bash -c "$(curl -fsSL https://qwen-code-assets.oss-cn-hangzhou.aliyuncs.com/installation/install-qwen.sh)"
```

**Windows (Run as Administrator CMD):**
```cmd
curl -fsSL -o %TEMP%\install-qwen.bat https://qwen-code-assets.oss-cn-hangzhou.aliyuncs.com/installation/install-qwen.bat && %TEMP%\install-qwen.bat
```

> **Note:** Restart your terminal after installation to ensure environment variables take effect.

### Method 2: Package Manager (NPM)

```bash
npm install -g @qwen-code/qwen-code@latest
```

### Method 3: Homebrew

```bash
brew install qwen-code
```

### Method 4: Build from Source

```bash
git clone https://github.com/QwenLM/qwen-code.git
cd qwen-code
npm install
npm run build
npm install -g .
```

---

## Quick Start

```bash
# Verify installation
qwen --version

# Start interactive session
qwen

# First-time setup
# 1. Run /auth to configure authentication
# 2. Select authentication method (Alibaba Cloud / OAuth / API Key)
# 3. Enter credentials

# Start coding
qwen
> Explain this codebase
```

---

## CLI Commands

### Global Options

| Option | Short | Description | Example |
|--------|-------|-------------|---------|
| `--help` | `-h` | Show help | `qwen --help` |
| `--version` | `-v` | Show version | `qwen --version` |
| `--acp` | | Enable Agent Client Protocol | `qwen --acp` |

### Command: qwen

**Description:** Start the interactive Qwen Code session.

**Usage:**
```bash
qwen [options]
```

**Options:**
| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `--acp` | flag | No | false | Enable ACP mode for IDE integration |

**Examples:**
```bash
# Start interactive mode
qwen

# Start with ACP enabled (for IDE integration)
qwen --acp
```

**Exit Codes:**
- `0` - Success
- `1` - General error
- `130` - Interrupted (Ctrl+C)

### Command: qwen --version

**Description:** Show version information.

**Usage:**
```bash
qwen --version
```

---

## TUI/Interactive Commands

Once inside the Qwen Code TUI, use these slash commands:

| Command | Description | Example |
|---------|-------------|---------|
| `/help` | Show help | `/help` |
| `/?` | Show help (alias) | `/?` |
| `/auth` | Configure authentication | `/auth` |
| `/model` | Switch model | `/model` |
| `/init` | Analyze directory and create QWEN.md | `/init` |
| `/clear` | Clear terminal and start new session | `/clear` |
| `/compress` | Replace chat history with summary | `/compress` |
| `/settings` | Open settings (language, theme) | `/settings` |
| `/summary` | Generate project summary | `/summary` |
| `/resume` | Resume previous session | `/resume` |
| `/stats` | Show session statistics | `/stats` |
| `/quit` | Exit Qwen Code | `/quit` |
| `/exit` | Exit Qwen Code (alias) | `/exit` |

---

## Configuration

### Authentication Methods

Qwen Code supports multiple authentication methods:

#### 1. Qwen OAuth (Recommended)

- 2,000 requests per day (no token counting)
- 60 requests per minute rate limit
- Automatic credential refresh
- Zero cost for individual users

Setup:
```bash
qwen
> /auth
# Select "Qwen OAuth"
# Complete browser authentication
```

#### 2. Alibaba Cloud Coding Plan

Setup:
```bash
qwen
> /auth
# Select "Alibaba Cloud"
# Enter API key from Alibaba Cloud Model Studio
```

#### 3. OpenAI-Compatible API

Works with:
- Alibaba Cloud Bailian/ModelStudio
- OpenRouter
- ModelScope

Setup:
```bash
qwen
> /auth
# Select "API Key"
# Enter your API key and base URL
```

### Configuration File Format

**File Location:** `~/.config/qwen/config.json`

```json
{
  "auth": {
    "method": "oauth",
    "provider": "qwen"
  },
  "model": {
    "default": "qwen3-coder-plus",
    "fallback": "qwen3-coder"
  },
  "ui": {
    "theme": "dark",
    "language": "en"
  },
  "features": {
    "autoComplete": true,
    "inlineSuggestions": true
  }
}
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `QWEN_API_KEY` | API key for authentication |
| `QWEN_BASE_URL` | Custom API base URL |
| `QWEN_MODEL` | Default model to use |

### QWEN.md Context File

Qwen Code supports project-level context via `QWEN.md`:

```markdown
# Project Context

## Architecture
- Frontend: React + TypeScript
- Backend: Node.js + Express
- Database: PostgreSQL

## Coding Standards
- Use functional components
- Follow ESLint rules
- Write tests with Jest

## Important Files
- src/config.ts - Configuration
- src/utils/ - Utility functions
```

Generate with:
```bash
qwen
> /init
```

### IDE Integration

#### JetBrains IDE Setup

1. Install Qwen Code CLI and configure Coding Plan
2. Open JetBrains IDE, navigate to **AI Chat** panel
3. Click **Install Plugin** to install JetBrains AI Assistant
4. Click three-dot menu → **Add Custom Agent**
5. Enter configuration:

```json
{
  "agent_servers": {
    "qwen": {
      "command": "/path/to/qwen",
      "args": ["--acp"],
      "env": {}
    }
  }
}
```

Find path with:
```bash
# macOS/Linux
which qwen

# Windows
where qwen
```

---

## API/Protocol Endpoints

### Alibaba Cloud Model Studio API

#### Endpoint: POST /v1/chat/completions

**Description:** Create a chat completion.

**Request:**
```json
{
  "model": "qwen3-coder-plus",
  "messages": [
    {"role": "system", "content": "You are a helpful assistant"},
    {"role": "user", "content": "Write a Python function"}
  ],
  "temperature": 0.7
}
```

**Response:**
```json
{
  "id": "chatcmpl-xxx",
  "choices": [{
    "message": {
      "role": "assistant",
      "content": "Here's the function..."
    }
  }],
  "usage": {
    "prompt_tokens": 20,
    "completion_tokens": 50
  }
}
```

### Agent Client Protocol (ACP)

Enable ACP mode for IDE integration:

```bash
qwen --acp
```

### Supported Models

| Model | Description |
|-------|-------------|
| `qwen3-coder-plus` | Enhanced coding model |
| `qwen3-coder` | Standard coding model |
| `qwen-max` | Most capable model |
| `qwen-plus` | Balanced capability |
| `qwen-flash` | Fast responses |
| `deepseek-chat` | DeepSeek model |
| `MiniMax-M2.5` | MiniMax model |
| `glm-5` | GLM model |

---

## Usage Examples

### Example 1: First-Time Setup

```bash
# Install Qwen Code
bash -c "$(curl -fsSL https://qwen-code-assets.oss-cn-hangzhou.aliyuncs.com/installation/install-qwen.sh)"

# Restart terminal, then:
qwen

# In the TUI:
> /auth
# Select authentication method
# Follow prompts to complete setup

> What does this project do?
```

### Example 2: Code Generation

```bash
qwen

# In the TUI:
> Create a Python Flask API with CRUD operations for users
> Generate unit tests for the auth module
> Refactor this function to use async/await
```

### Example 3: Model Switching

```bash
qwen

# In the TUI:
> /model
# (Select from available models)

# Or directly specify:
> /model qwen3-coder-plus

# If model not in list, update Qwen Code:
> /quit
npm install -g @qwen-code/qwen-code@latest
qwen
```

### Example 4: Project Analysis

```bash
cd my-project
qwen

# In the TUI:
> /init
# (Creates QWEN.md with project context)

> Explain the architecture of this project
> Find potential bugs in the codebase
> Suggest improvements for performance
```

### Example 5: Session Management

```bash
qwen

# In the TUI:
> Explain this function
# (Conversation continues...)

> /compress
# (Compresses history to save tokens)

> /summary
# (Generates project summary)

> /resume
# (Resumes previous session)

> /quit
```

### Example 6: Multi-turn Conversation

```bash
qwen

# In the TUI:
> Create a React component for a shopping cart
# (AI generates component)

> Add TypeScript types to that component
# (AI updates with types)

> Now add a function to calculate total price
# (AI adds the function)

> Write unit tests for this component
# (AI generates tests)
```

### Example 7: Working with Large Codebases

```bash
cd large-project
qwen

# In the TUI:
> /init

> Find all files that use the deprecated API
> Migrate the authentication to use JWT
> Create a migration guide for the team
```

---

## Troubleshooting

### Issue: Command not found after installation

**Solution:**
```bash
# Restart terminal
# Or source shell profile
source ~/.bashrc  # or ~/.zshrc

# Check installation
which qwen
qwen --version
```

### Issue: Authentication fails

**Solution:**
```bash
# Re-authenticate
qwen
> /auth

# Select different auth method
# Verify API key is correct
# Check internet connection
```

### Issue: Model not in list

**Solution:**
```bash
# Exit Qwen Code
> /quit

# Update to latest version
npm install -g @qwen-code/qwen-code@latest

# Restart
qwen
> /model
```

### Issue: IDE integration not working

**Solution:**
1. Verify Qwen Code path is correct in IDE settings
2. Ensure `--acp` flag is included
3. Check JetBrains AI Assistant plugin is installed
4. Verify Qwen Code CLI is properly configured

### Issue: Rate limit exceeded

**Solution:**
- Qwen OAuth: 2,000 requests/day, 60/minute
- Wait before making more requests
- Consider upgrading to paid plan

### Issue: VLM (Vision Language Model) issues

**Solution:**
```bash
# Enable VLM switch mode
qwen --vlm-switch-mode once
```

### Issue: Compressed history loses context

**Solution:**
- Use `/compress` sparingly
- Start new session with `/clear` if needed
- Use `/summary` to preserve key information

### Issue: Windows installation problems

**Solution:**
```cmd
# Run CMD as Administrator
curl -fsSL -o %TEMP%\install-qwen.bat https://qwen-code-assets.oss-cn-hangzhou.aliyuncs.com/installation/install-qwen.bat && %TEMP%\install-qwen.bat

# Restart CMD after installation
```

---

## Additional Resources

- **Official Website:** https://github.com/QwenLM/qwen-code
- **Alibaba Cloud Model Studio:** https://www.alibabacloud.com/help/en/model-studio
- **Documentation:** https://www.alibabacloud.com/help/en/model-studio/qwen-code-coding-plan
- **Community Blog:** https://www.alibabacloud.com/blog/boost-your-coding-workflow-with-qwen-code
