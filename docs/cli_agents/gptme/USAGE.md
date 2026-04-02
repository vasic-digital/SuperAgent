# GPTMe - Usage Guide

## Table of Contents

1. [Getting Started](#getting-started)
2. [Common Workflows](#common-workflows)
3. [Advanced Features](#advanced-features)
4. [Best Practices](#best-practices)
5. [Troubleshooting](#troubleshooting)

---

## Getting Started

### First-Time Setup

1. **Install GPTMe**
   ```bash
   # Using pipx (recommended)
   pipx install gptme
   
   # Using uv
   uv tool install gptme
   
   # With all extras
   pipx install 'gptme[all]'
   ```

2. **Configure API Key**
   ```bash
   # Option 1: Environment variable
   export ANTHROPIC_API_KEY="sk-ant-..."
   
   # Option 2: Config file
   mkdir -p ~/.config/gptme
   cat > ~/.config/gptme/config.toml << 'EOF'
   [env]
   ANTHROPIC_API_KEY = "sk-ant-..."
   MODEL = "anthropic/claude-sonnet-4-6"
   EOF
   ```

3. **Verify Installation**
   ```bash
   gptme --version
   gptme-doctor
   ```

### Basic Navigation

```bash
# Start interactive session
gptme

# Start with a prompt
gptme "Hello, what can you do?"

# Include files in context
gptme "explain this" main.py
gptme "review these" src/*.py

# Resume last conversation
gptme -r

# List all conversations
gptme-util chats list
```

---

## Common Workflows

### 1. Code Exploration

**Understanding a New Codebase**

```bash
gptme "What does this project do?"
```

GPTMe will:
1. Read README.md and project files
2. Analyze directory structure
3. Examine key configuration files
4. Summarize architecture

**Finding Specific Code**

```bash
# Search for authentication logic
gptme "Find where user authentication is handled"

# Find specific function
gptme "Where is the login function defined?"
```

**Explaining Complex Code**

```bash
# Explain a file
gptme "Explain the logic in" src/utils/cache.py

# Explain with line numbers
gptme "Explain lines 50-100 of" main.py
```

### 2. Code Modifications

**Making Edits**

```bash
# Simple edit request
gptme "Add input validation to the login function"

# With file context
gptme "Refactor main.py to use async/await" main.py
```

**Creating New Files**

```bash
# Create a new module
gptme "Create a middleware for JWT authentication"

# Create with specific requirements
gptme "Create a React component for a todo list with TypeScript"
```

**Batch Operations**

```bash
# Process multiple files
gptme "Add docstrings to all functions in" src/*.py
```

### 3. Development Workflows

**Testing**

```bash
# Run and fix tests
gptme "Run the test suite and fix any failing tests"

# Pipe test output
gptme -n "fix these tests" < pytest_output.txt

# Fix specific test
gptme "Fix the test_user_login test" tests/test_auth.py
```

**Git Workflows**

```bash
# Generate commit message
git diff | gptme "Write a good commit message for these changes"

# Review changes
git diff | gptme "Review these changes"

# Complete TODOs
git diff | gptme "Complete the TODOs in this diff"
```

**Code Review**

```bash
# Review PR
gptme "Review this PR" https://github.com/user/repo/pull/123

# Review local changes
gptme "Review my changes" $(git diff --name-only)
```

### 4. Shell & System Tasks

**Shell Expert Mode**

```bash
# Get the right command
gptme "How do I find all Python files modified in the last 24 hours?"

# Explain a command
gptme "What does 'find . -type f -name '*.py' -mtime -1' do?"

# Complex pipeline
gptme "Show me disk usage by directory, sorted by size"
```

**File Operations**

```bash
# Batch rename
gptme "Rename all .txt files to .md"

# Data processing
gptme "Extract all email addresses from" contacts.txt

# Format conversion
gptme "Convert this CSV to JSON" data.csv
```

### 5. Web & Research

**Web Browsing**

```bash
# Research a topic
gptme "What are the latest features in Python 3.12?"

# Check documentation
gptme "Look up the FastAPI documentation for dependency injection"

# Answer from URL
gptme "What does this article say about AI?" https://example.com/article
```

**Data Extraction**

```bash
# Extract information
gptme "Extract all links from" https://example.com

# Summarize webpage
gptme "Summarize this article" https://example.com/long-article
```

### 6. Vision & Images

**Image Analysis**

```bash
# Analyze an image
gptme "What do you see in this image?" screenshot.png

# Extract text
gptme "Extract the text from this image" document.jpg

# Code from screenshot
gptme "Write the code shown in this screenshot" ui-mockup.png
```

**Screenshots**

```bash
# Analyze current screen
gptme "What's on my screen?" --screenshot

# Debug UI issues
gptme "Why is this button not working?" error-screenshot.png
```

### 7. Python & Data Analysis

**Interactive Python**

```bash
# Data analysis
gptme "Analyze this CSV and show me statistics" data.csv

# Visualization
gptme "Create a plot of this data" data.csv

# Complex computation
gptme "Calculate the Fibonacci sequence up to 1000"
```

**Jupyter-like Experience**

```bash
# Start analysis session
gptme "Load pandas and analyze this dataset" sales_data.csv

# Continue with context
gptme -r "Now create a visualization of the trends"
```

---

## Advanced Features

### 1. Non-Interactive Mode

**For Scripts and CI/CD**

```bash
# Auto-approve all confirmations
gptme -y "fix the failing tests"

# Fully non-interactive (no user interaction)
gptme -n "run the test suite and fix failures"

# Pipe input
echo "Hello" | gptme -n "respond to this"
```

**Combining with Other Tools**

```bash
# Process git diff
git diff | gptme -n "generate commit message"

# Process test output
pytest 2>&1 | gptme -n "fix the failing tests"

# Batch process files
find . -name "*.py" -exec gptme -n "add type hints to {}" \;
```

### 2. Subagents

**Parallel Task Execution**

```bash
# Use subagent tool
gptme "Analyze this codebase using subagents for different components"

# Complex multi-step task
gptme "Create a full-stack app with subagents for frontend and backend"
```

### 3. Chaining Tasks

**Multi-Prompt Workflows**

```bash
# Chain multiple prompts
gptme 'Create a fibonacci function' - 'Write tests for it' - 'Commit the changes'

# Complex workflow
gptme 'Implement feature X' - 'Test it' - 'Refactor if needed' - 'Create PR'
```

### 4. Tool Selection

**Selective Tool Access**

```bash
# Restrict to safe tools
gptme -t read,python "Analyze this code without making changes"

# Add specific tools
gptme -t +browser "Research this topic online"

# Exclude dangerous tools
gptme -t=-shell "Help me understand this code safely"

# Custom tool set
gptme -t shell,python,read,save "Work in restricted mode"
```

### 5. Context Management

**Workspace Management**

```bash
# Use specific workspace
gptme -w /path/to/project "Analyze this project"

# Create workspace in logs
gptme -w @log "Create a new project"
```

**Context Compression**

```bash
# Manually compact conversation
gptme
> /compact

# Start fresh while keeping workspace
gptme
> /summarize
```

### 6. MCP Integration

**Using MCP Servers**

```toml
# ~/.config/gptme/config.toml
[[mcp.servers]]
name = "github"
command = "npx"
args = ["-y", "@github/mcp-server"]
env = { GITHUB_TOKEN = "ghp_..." }
```

```bash
# Use MCP tools
gptme "List my open GitHub issues"
gptme "Create an issue for this bug"
```

### 7. Autonomous Agents

**Creating an Agent**

```bash
# Create agent workspace
gptme-agent create ~/my-agent --name MyAgent

# Configure the agent
cd ~/my-agent
# Edit gptme.toml, tasks/, journal/, etc.

# Install (schedule to run)
gptme-agent install

# Check status
gptme-agent status

# Run manually
gptme-agent run
```

### 8. RAG (Retrieval Augmented Generation)

**Index and Search Local Files**

```bash
# Index project files
gptme-util context index

# Query with RAG
gptme "Find code related to authentication"
```

---

## Best Practices

### 1. Effective Communication

**Be Specific:**
```bash
# Good
gptme "Add error handling for network timeouts in fetchUserData()"

# Less effective
gptme "Fix the error handling"
```

**Provide Context:**
```bash
# Good
gptme "This is a React app using Redux. Add a new action for user logout."

# Better with files
gptme "Add Redux action for logout" src/store/actions.js src/store/reducers.js
```

**Use References:**
```bash
# Reference specific files
gptme "Update src/components/Button.tsx to support a loading state"

# Reference with wildcards
gptme "Update all test files" tests/**/*.test.ts
```

### 2. Managing Sessions

**Naming Conventions:**
```bash
# Start with meaningful name
gptme --name "auth-refactor-2025-01" "Refactor authentication"

# Resume specific conversation
gptme -r  # Shows list to choose from
```

**Session Hygiene:**
```bash
# List conversations
gptme-util chats list

# Search past conversations
gptme-util chats search "authentication"

# Export important conversations
gptme
> /export
```

### 3. Security

**Review Before Approving:**
- Always review code changes before confirming
- Check shell commands before execution
- Be cautious with file deletions

**Use Restricted Mode:**
```bash
# For untrusted code
gptme -t read,python "Analyze this without executing shell commands"

# For sensitive environments
gptme -y --no-confirm  # Never use this blindly!
```

**Environment Isolation:**
```bash
# Use Docker for isolation
docker run -it -v $(pwd):/workspace gptme/gptme "work in container"

# Use workspace isolation
gptme -w /tmp/isolated-workspace "Run untrusted code"
```

### 4. Cost Management

**Monitor Usage:**
```bash
# Enable cost tracking
export GPTME_COSTS=true

# Check tokens
gptme
> /tokens
```

**Optimize Context:**
```bash
# Compact regularly
gptme
> /compact

# Start fresh for new tasks
gptme  # New conversation

# Use appropriate model
gptme -m openai/gpt-4o-mini "Simple task"
gptme -m anthropic/claude-opus-4 "Complex task"
```

### 5. Tool Workflows

**Tool Combinations:**

| Task | Tools Used | Command |
|------|------------|---------|
| Web research + code | browser, python | `gptme 'Research X and implement'` |
| Debug + fix | screenshot, patch | `gptme 'Fix this UI bug' screenshot.png` |
| Test + iterate | shell, python | `gptme 'Fix tests' - 'run again'` |
| Multi-file refactor | read, patch, subagent | `gptme 'Refactor with subagents'` |

---

## Troubleshooting

### Common Issues

**1. API Key Not Found**

```bash
# Check if set
echo $ANTHROPIC_API_KEY

# Set in config
cat > ~/.config/gptme/config.toml << 'EOF'
[env]
ANTHROPIC_API_KEY = "sk-ant-..."
EOF
```

**2. Tool Not Available**

```bash
# List available tools
gptme-util tools list

# Install missing dependencies
pipx install 'gptme[browser]'  # For browser tool
pipx install 'gptme[all]'       # For all tools
```

**3. Conversation Won't Resume**

```bash
# List available sessions
gptme-util chats list

# Start fresh
gptme --name "new-conversation" "Start fresh"
```

**4. Out of Context Space**

```bash
# Compact conversation
gptme
> /compact

# Or start new
gptme
```

**5. MCP Server Not Working**

```bash
# Check MCP config
gptme
> /tools  # See MCP tools

# Verify server installation
which npx
npm list -g @modelcontextprotocol/server-filesystem
```

### Getting Help

**Within GPTMe:**
```bash
> /help                    # Show commands
> How do I use X?         # Ask GPTMe
```

**External Resources:**
- Documentation: https://gptme.org/docs
- Discord: https://discord.gg/NMaCmmkxWv
- GitHub Issues: https://github.com/gptme/gptme/issues

---

## Quick Reference Card

### Essential Commands

| Command | Purpose |
|---------|---------|
| `gptme` | Start session |
| `gptme -r` | Resume session |
| `gptme -n "cmd"` | Non-interactive mode |
| `gptme -y "cmd"` | Auto-confirm mode |
| `/help` | Show help |
| `/exit` | Quit |
| `/undo` | Undo last action |
| `/compact` | Compact conversation |
| `/tools` | List tools |
| `/tokens` | Show token usage |

### File References

| Syntax | Meaning |
|--------|---------|
| `filename.py` | Include file |
| `*.py` | Include all matching files |
| `dir/` | Include directory |
| `@workspace` | Use workspace |

### Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl+C` | Cancel/Exit |
| `Ctrl+X Ctrl+E` | Edit in editor |
| `Ctrl+J` | New line |
| `↑/↓` | Command history |
| `Tab` | Autocomplete |

---

*For more details, see the [API Reference](./API.md) and [Architecture](./ARCHITECTURE.md) documentation.*
