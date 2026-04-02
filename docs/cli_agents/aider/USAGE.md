# Aider - Usage Guide

## Table of Contents

1. [Getting Started](#getting-started)
2. [Common Workflows](#common-workflows)
3. [Advanced Features](#advanced-features)
4. [Best Practices](#best-practices)
5. [Troubleshooting](#troubleshooting)

---

## Getting Started

### First-Time Setup

1. **Install Aider**
   ```bash
   # Recommended method
   curl -LsSf https://aider.chat/install.sh | sh
   ```

2. **Get API Keys**
   - [Anthropic](https://console.anthropic.com) for Claude
   - [OpenAI](https://platform.openai.com) for GPT-4o, o3-mini
   - [DeepSeek](https://platform.deepseek.com) for DeepSeek models
   - [Google](https://ai.google.dev) for Gemini

3. **Create Project Configuration**
   ```bash
   cd ~/my-project
   
   # Create .env file
   echo "ANTHROPIC_API_KEY=sk-ant-..." > .env
   echo "AIDER_MODEL=sonnet" >> .env
   ```

4. **Verify Installation**
   ```bash
   aider --version
   ```

### Basic Navigation

```bash
# Start in your project directory
cd ~/my-project

# Launch with your preferred model
aider --model sonnet

# You'll see:
# Aider v0.78.0
# Models: claude-3-7-sonnet with diff edit format
# Git repo: .git with 258 files
# Repo-map: using 1024 tokens
# > _ (prompt waiting for input)
```

---

## Common Workflows

### 1. Starting a New Feature

```bash
# Add the files you'll be editing
aider src/auth.py src/user.py tests/test_auth.py

# Or add files in chat
> /add src/auth.py src/user.py

# Request the feature
> Add JWT authentication to the auth module
  Create login and register endpoints in user.py
  Add corresponding tests

# Aider will:
# 1. Read the existing code
# 2. Implement JWT authentication
# 3. Create the endpoints
# 4. Write tests
# 5. Show you diffs
# 6. Commit changes
```

### 2. Understanding Code

```bash
# Switch to ask mode for questions
> /ask How does the caching system work?

# Or ask about specific files
> /add src/cache.py
> Explain the cache invalidation logic

# Request diagrams
> Show me the data flow for user registration
```

### 3. Refactoring

```bash
# Add files to refactor
> /add src/old_module.py

# Request refactoring
> Refactor this to use async/await instead of callbacks

# Aider will:
# - Convert callback patterns
# - Update error handling
# - Maintain existing behavior
# - Run tests if configured
```

### 4. Bug Fixing

```bash
# Add relevant files
> /add src/api.py src/errors.py

# Describe the bug
> Fix the timeout error in the API client
  The error occurs when the server takes >30s to respond
  Handle it gracefully and retry once

# Or paste an error
> Fix this error: ConnectionTimeout: Request timed out
```

### 5. Adding Tests

```bash
# Add source file (read-only) and test file
> /read-only src/calculator.py
> /add tests/test_calculator.py

# Request tests
> Add comprehensive tests for the Calculator class
  Cover edge cases like division by zero

# Aider will create tests and run them if --test-cmd is set
```

### 6. Documentation

```bash
# Add code file
> /add src/complex_module.py

# Generate documentation
> Add docstrings to all public functions
  Follow Google style guide

# Or generate README
> /add src/main.py src/config.py
> Create a README.md explaining how to use this project
```

### 7. Git Workflows

```bash
# See what changed
> /diff

# Undo last Aider commit
> /undo

# Commit external changes
> /commit "Manual updates to config"

# Run git commands
> /git log --oneline -10
```

---

## Advanced Features

### 1. Architect Mode

Use two models: one to design, one to implement:

```bash
# Start with architect mode
aider --architect --model o1 --editor-model sonnet

# Or switch in chat
> /architect Design a REST API for user management

# The architect model will design the API
# The editor model will implement the code
```

### 2. Context Mode

Let Aider automatically identify files:

```bash
# Start in context mode
aider --edit-format context

# Or switch in chat
> /context Implement user authentication

# Aider will:
# 1. Analyze your request
# 2. Find relevant files
# 3. Add them to the chat
# 4. Make the changes
```

### 3. Voice Input

```bash
# Enable voice
aider --voice-language en

# Or in chat
> /voice

# Speak your request
# Aider transcribes and processes
```

### 4. Web Pages & Images

```bash
# Add documentation
> /web https://docs.example.com/api

# Add an image (screenshot, diagram)
> /paste my-screenshot.png

# Or from file
> Add the API from this screenshot: /path/to/image.png
```

### 5. Watch Mode (IDE Integration)

```bash
# Watch files for changes
aider --watch-files src/

# Now in your IDE:
# 1. Add comment: "// Add error handling here"
# 2. Aider will implement the change
# 3. Review and commit
```

### 6. Copy/Paste Mode

For web-based LLMs without API access:

```bash
# Use copy/paste mode
aider --edit-format copypaste

# Aider will:
# 1. Format context for copying
# 2. Prompt you to paste the LLM response
# 3. Parse and apply changes
```

### 7. Prompt Caching (Cost Reduction)

```bash
# Enable caching
aider --cache-prompts --cache-keepalive-pings 3

# Aider will cache the system prompt
# Reduces costs for repeated similar requests
```

### 8. Multi-Model Reasoning

```bash
# Use reasoning models
aider --model o1 --reasoning-effort high
aider --model deepseek-r1

# Control thinking
> /think-tokens 8192
> /reasoning-effort medium
```

---

## Best Practices

### 1. Effective Communication

**Be Specific:**
```
# Good
> Add input validation for email in the User.create() method
  Return 400 status for invalid emails

# Less effective
> Fix the user code
```

**Provide Context:**
```
# Good
> This is a FastAPI app. Add rate limiting to the login endpoint
  using slowapi with a limit of 5 requests per minute

# Less effective
> Add rate limiting
```

**Use File References:**
```
# Good
> Update src/utils.py to use the new logging format
  defined in src/config.py

# Less effective
> Fix the logging
```

### 2. Managing Context

**Add Only Necessary Files:**
```
# Good - 2-4 relevant files
> /add src/auth.py src/middleware.py tests/test_auth.py

# Avoid - too many files
> /add src/*.py  # Don't do this
```

**Drop Files When Done:**
```
> /drop src/completed_feature.py
```

**Use Read-Only for Reference:**
```
> /read-only src/config.py  # Reference, won't be edited
> /add src/main.py          # Will be edited
```

### 3. Working with Git

**Review Before Committing:**
```
# Disable auto-commits to review first
aider --no-auto-commits

# Review changes
> /diff

# Then commit manually
> /commit "Implement feature X"
```

**Use Descriptive Messages:**
```
# Let Aider generate good messages
# Or specify your own
> /commit "Add JWT auth with refresh tokens"
```

**Undo When Needed:**
```
> /undo  # Removes last Aider commit
```

### 4. Cost Management

**Use Appropriate Models:**
```bash
# Cheap model for simple tasks
aider --model haiku

# Strong model for complex changes
aider --model sonnet

# Reasoning model for design
aider --model o1 --architect --editor-model haiku
```

**Enable Caching:**
```bash
aider --cache-prompts
```

**Monitor Usage:**
```
> /tokens  # Show current token count
```

### 5. Code Quality

**Enable Linting:**
```bash
aider --lint-cmd "python -m flake8" --auto-lint
```

**Run Tests:**
```bash
aider --test-cmd "pytest" --auto-test
```

**Check Changes:**
```
> /lint  # Run linter
> /test  # Run tests
```

---

## Troubleshooting

### Common Issues

**1. API Key Not Found**

```bash
# Check your .env file
cat .env

# Or set explicitly
export ANTHROPIC_API_KEY=sk-ant-...
aider --model sonnet

# Or pass on command line
aider --model sonnet --api-key anthropic=sk-ant-...
```

**2. Model Not Responding**

```bash
# Check model availability
aider --list-models claude

# Try a different model
aider --model gpt-4o

# Check internet connection
```

**3. Changes Not Applied**

```
# Check for edit format issues
> /settings  # Show current edit format

# Try different format
aider --edit-format whole

# Check if files are writable
ls -la src/
```

**4. Git Issues**

```bash
# Ensure you're in a git repo
git status

# Initialize if needed
git init

# Check git config
git config user.name
git config user.email
```

**5. Out of Context Space**

```
# Drop unnecessary files
> /drop src/unused.py

# Clear history
> /clear

# Use weaker model for summaries (already default)
```

### Getting Help

**Within Aider:**
```
> /help              # Show all commands
> /help /add        # Help for specific command
> /settings          # Show current settings
```

**External Resources:**
- Documentation: [aider.chat/docs](https://aider.chat/docs)
- GitHub Issues: [Aider-AI/aider/issues](https://github.com/Aider-AI/aider/issues)
- Discord: [Aider Discord](https://discord.gg/Y7X7bhMQFV)

---

## Quick Reference Card

### Essential Commands

| Command | Purpose |
|---------|---------|
| `aider` | Start session |
| `aider --model X` | Start with model X |
| `/add file.py` | Add file to chat |
| `/drop file.py` | Remove file from chat |
| `/model X` | Switch model |
| `/ask` | Ask mode |
| `/code` | Code mode |
| `/architect` | Architect mode |
| `/undo` | Undo last commit |
| `/diff` | Show changes |
| `/help` | Show help |
| `/exit` | Quit |

### Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl+C` | Cancel/Exit |
| `Ctrl+L` | Clear screen |
| `Ctrl+R` | Search history |
| `↑/↓` | Scroll history |
| `Tab` | Autocomplete |
| `Meta+Enter` | Submit (multiline mode) |

---

*For more details, see the [API Reference](./API.md) and [Architecture](./ARCHITECTURE.md) documentation.*
