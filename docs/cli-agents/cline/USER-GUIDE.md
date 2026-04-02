# Cline User Guide

## Table of Contents
1. [Installation](#installation)
2. [Quick Start](#quick-start)
3. [CLI Commands](#cli-commands)
4. [TUI/Interactive Commands](#tui-interactive-commands)
5. [Configuration](#configuration)
6. [API/Protocol Endpoints](#api-protocol-endpoints)
7. [Usage Examples](#usage-examples)
8. [Troubleshooting](#troubleshooting)

## Installation

### Method 1: VS Code Extension (Recommended)

```bash
# Install via VS Code CLI
code --install-extension saoudrizwan.claude-dev

# Or search "Cline" in VS Code Extensions marketplace
# Press Ctrl+Shift+X, search "Cline", click Install
```

### Method 2: Cursor Extension

```bash
# Install in Cursor
cursor --install-extension saoudrizwan.claude-dev

# Or from Extensions marketplace in Cursor
```

### Method 3: VSCodium/Open VSX

```bash
# Install from Open VSX
codium --install-extension saoudrizwan.claude-dev
```

### Method 4: Nightly Version

```bash
# Install nightly build for latest features
code --install-extension saoudrizwan.cline-nightly
```

### Method 5: Build from Source

```bash
# Clone repository
git clone https://github.com/cline/cline.git
cd cline

# Install dependencies
npm install

# Build
npm run build

# Package extension
npx vsce package

# Install from VSIX
code --install-extension cline-*.vsix
```

### Prerequisites

- VS Code 1.84+ or Cursor
- Node.js 18+ (for building from source)
- API key from supported provider (Anthropic, OpenAI, etc.)

## Quick Start

### First-Time Setup

```bash
# Open VS Code
code .

# Open Cline panel
# Click Cline icon in Activity Bar (left sidebar)
# Or press Ctrl+Shift+P → "Cline: Focus on Cline View"

# Configure API provider
# 1. Click settings icon in Cline panel
# 2. Select provider (Anthropic, OpenAI, OpenRouter, etc.)
# 3. Enter API key
```

### Basic Usage

```bash
# Open Cline in VS Code
# Click Cline icon in sidebar or use command palette

# Start a task
# Type in Cline chat input

# Use slash commands
# Type / to see available commands
```

### Hello World

```bash
# Open VS Code with Cline installed
code .

# Open Cline panel (click icon in left sidebar)

# In Cline chat:
> Create a Python script that prints "Hello, World!"

# Cline will create the file and show diff
```

## CLI Commands

Cline is primarily a VS Code extension, but provides command palette commands:

### VS Code Command Palette Commands

| Command | Description | Shortcut |
|---------|-------------|----------|
| Cline: Focus on Cline View | Open Cline panel | |
| Cline: New Task | Start new task | |
| Cline: Open Settings | Configure Cline | |
| Cline: Export Conversation | Save chat history | |
| Cline: Import Conversation | Load chat history | |

### Access via Command Palette

```bash
# Open Command Palette
Ctrl+Shift+P (Linux/Windows)
Cmd+Shift+P (macOS)

# Type "Cline" to see all commands
```

### Extension Settings

| Setting | Description | Default |
|---------|-------------|---------|
| cline.provider | API provider | anthropic |
| cline.apiKey | API key | |
| cline.model | Default model | claude-sonnet-4 |
| cline.autoApprove | Auto-approve actions | false |
| cline.customInstructions | Custom system prompt | |
| cline.theme | Color theme | auto |

## TUI/Interactive Commands

When using Cline in VS Code, use these slash commands:

| Command | Description | Example |
|---------|-------------|---------|
| /help | Show available commands | `/help` |
| /newtask | Start fresh task with context | `/newtask` |
| /smol | Compact conversation | `/smol` |
| /compact | Alias for /smol | `/compact` |
| /newrule | Create rule file | `/newrule` |
| /deep-planning | Thorough planning mode | `/deep-planning` |
| /explain-changes | Explain git diff | `/explain-changes` |
| /reportbug | Report issue with diagnostics | `/reportbug` |

### Command Details

#### /newtask

Start fresh with distilled context from current conversation. Useful for:
- Breaking complex tasks into steps
- Preserving important context while removing noise
- Handing off work between sessions

**Example:**
```
/newtask
```

#### /smol (or /compact)

Compress conversation history while preserving essential context:
- Frees up context window space
- Maintains key insights
- Allows continuing same task

**Example:**
```
/smol
```

#### /newrule

Create a `.clinerules` file that teaches Cline your preferences:
- Coding standards
- Project context
- Communication style

**Example:**
```
/newrule
# Cline will guide you through creating the rule
```

#### /deep-planning

Transform Cline into meticulous architect mode:
1. Silent investigation of codebase
2. Discussion of requirements
3. Plan creation (implementation_plan.md)
4. Task creation with trackable steps

**Example:**
```
/deep-planning
# Then describe your feature request
```

#### /explain-changes

Generate AI explanations for git diffs:
- Last commit
- Uncommitted changes
- Specific commits
- Pull requests

**Example:**
```
/explain-changes
/explain-changes HEAD~3
/explain-changes --pr 42
```

### Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| Ctrl+Enter | Send message |
| Shift+Enter | New line in input |
| Up | Navigate history |
| Escape | Cancel current operation |

### Context Menu Actions

Right-click in editor for Cline actions:
- Explain this code
- Fix this code
- Generate tests
- Refactor
- Document this

## Configuration

### Configuration File Format

Cline stores settings in VS Code settings:

```json
// settings.json
{
  "cline.provider": "anthropic",
  "cline.apiKey": "sk-ant-...",
  "cline.model": "claude-sonnet-4-20250514",
  "cline.autoApprove": false,
  "cline.customInstructions": "Always use TypeScript. Follow ESLint rules.",
  "cline.theme": "dark",
  "cline.approvalSettings": {
    "readFiles": true,
    "editFiles": false,
    "runCommands": false,
    "useBrowser": false,
    "useMcp": false
  },
  "cline.mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/workspace"]
    },
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "ghp_..."
      }
    }
  }
}
```

### Project-Specific Rules (.clinerules)

Create `.clinerules` in project root:

```markdown
# Cline Rules for This Project

## Coding Standards
- Use TypeScript for all new code
- Follow ESLint configuration
- Write tests for all features
- Use functional components with hooks

## Project Context
- Frontend: React + Vite + Tailwind CSS
- Backend: Node.js + Express
- Database: PostgreSQL with Prisma ORM
- Testing: Jest + React Testing Library

## Workflow
- Create feature branches from main
- Write descriptive commit messages
- Run tests before committing
- Update documentation for API changes
```

### Environment Variables

| Variable | Description | Required | Example |
|----------|-------------|----------|---------|
| ANTHROPIC_API_KEY | Anthropic API key | Yes* | `sk-ant-...` |
| OPENAI_API_KEY | OpenAI API key | Yes* | `sk-...` |
| OPENROUTER_API_KEY | OpenRouter API key | Yes* | `sk-or-...` |
| CLINE_CUSTOM_INSTRUCTIONS | Custom prompt | No | `Always use TypeScript` |

*One provider required

### Configuration Locations (in order of precedence)

1. VS Code UI settings (highest priority)
2. Workspace settings (`.vscode/settings.json`)
3. User settings (`~/.config/Code/User/settings.json`)
4. `.clinerules` file in project root
5. Default settings (lowest priority)

## API/Protocol Endpoints

Cline connects to LLM provider APIs:

### Supported Providers

| Provider | Base URL | Example Model |
|----------|----------|---------------|
| Anthropic | api.anthropic.com | claude-sonnet-4 |
| OpenAI | api.openai.com | gpt-4o |
| OpenRouter | openrouter.ai/api | various |
| AWS Bedrock | bedrock.amazonaws.com | claude-sonnet-4 |
| GCP Vertex | vertexai.googleapis.com | gemini-pro |
| Azure | {your-resource}.openai.azure.com | gpt-4 |
| Ollama | localhost:11434 | local models |

### Local Models (Ollama)

```bash
# Start Ollama
ollama serve

# Pull a model
ollama pull codellama:7b-code

# Configure Cline
# Set provider to "Ollama"
# Set model to "codellama:7b-code"
# Set base URL to "http://localhost:11434"
```

### MCP Server Configuration

Configure MCP in VS Code settings:

```json
{
  "cline.mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/home/user/projects"]
    },
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "ghp_..."
      }
    },
    "fetch": {
      "command": "uvx",
      "args": ["mcp-server-fetch"]
    }
  }
}
```

## Usage Examples

### Example 1: Feature Implementation

```bash
# Open VS Code
code .

# Open Cline panel

# Request feature
> Create a user authentication system with login, signup, and password reset
> Use JWT tokens
> Include input validation
> Write unit tests

# Review each change in diff view
# Approve or request changes
```

### Example 2: Code Refactoring

```bash
# In VS Code with Cline

# Select code to refactor
# Right-click → "Refactor with Cline"

# Or in chat
> Refactor the UserService class to use dependency injection
> Extract database operations into a repository pattern

# Review diff before applying
```

### Example 3: Debugging

```bash
# Open Cline panel

# Describe issue
> The login form is not working. Debug and fix the issue.

# Cline will:
# 1. Investigate the code
# 2. Identify the problem
# 3. Propose a fix
# 4. Show the diff

# Interactive debugging
> What error messages are in the console?
> Check the network requests
```

### Example 4: Test Generation

```bash
# Select function to test
# Right-click → "Generate tests with Cline"

# Or in chat
> Generate comprehensive unit tests for the calculateTotal function
> Include edge cases and error scenarios
> Use Jest with 100% coverage
```

### Example 5: Documentation

```bash
# In Cline chat
> Generate API documentation for the UserController
> Include all endpoints, request/response examples
> Format as OpenAPI/Swagger spec

> Create a README with setup instructions
```

### Example 6: Using MCP Tools

```bash
# Configure MCP in settings first

# In Cline chat
> Search for issues labeled "bug" in the GitHub repository
> Create a new issue for the authentication bug we found

> Fetch the content from https://api.example.com/docs
> Implement a client based on this API documentation
```

### Example 7: Multi-File Changes

```bash
# In Cline chat
> Add a new field "phone" to the User model
> Update the database schema
> Add validation in the API
> Update the frontend form
> Write tests for all changes

# Cline will coordinate changes across files
# Review each diff before approval
```

## Troubleshooting

### Issue: Extension Not Loading

**Symptoms:** Cline icon not visible or extension fails to activate

**Solution:**
```bash
# Check extension is installed
code --list-extensions | grep cline

# Reinstall if needed
code --uninstall-extension saoudrizwan.claude-dev
code --install-extension saoudrizwan.claude-dev

# Reload VS Code window
# Ctrl+Shift+P → "Developer: Reload Window"
```

### Issue: API Key Not Working

**Symptoms:** "Invalid API key" or authentication errors

**Solution:**
```bash
# Verify API key
# Open Cline settings
# Check key is correct

# Test key manually
curl -H "x-api-key: sk-ant-..." https://api.anthropic.com/v1/me

# Check provider selection matches key type
```

### Issue: Model Not Responding

**Symptoms:** No response or timeout errors

**Solution:**
```bash
# Check internet connection
# Try different model
# Check model is available for your API key
# Check rate limits not exceeded
```

### Issue: Diff View Not Showing

**Symptoms:** Changes applied without showing diff

**Solution:**
```bash
# Check settings
"cline.showDiffBeforeApply": true

# Or in Cline panel settings
# Enable "Show diff before applying changes"
```

### Issue: MCP Tools Not Available

**Symptoms:** MCP commands not working

**Solution:**
```bash
# Check MCP configuration in settings
# Verify MCP server is installed
# Check environment variables for MCP
# Restart VS Code after MCP config changes
```

### Issue: Slow Performance

**Symptoms:** Long response times

**Solution:**
```bash
# Use faster model (Claude Haiku, GPT-3.5)
# Compact conversation with /smol
# Start new task with /newtask
# Check internet connection
```

### Issue: Context Window Full

**Symptoms:** "Context window exceeded" errors

**Solution:**
```bash
# Use /smol to compact conversation
# Use /newtask to start fresh with context
# Remove unnecessary files from context
# Use model with larger context (Claude Opus)
```

---

**Last Updated:** 2026-04-02
**Version:** 3.5.0
