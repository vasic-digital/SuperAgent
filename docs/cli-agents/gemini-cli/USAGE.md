# Gemini CLI - Usage Guide

## Table of Contents

1. [Getting Started](#getting-started)
2. [Common Workflows](#common-workflows)
3. [Advanced Features](#advanced-features)
4. [Best Practices](#best-practices)
5. [Troubleshooting](#troubleshooting)

---

## Getting Started

### First-Time Setup

1. **Install Gemini CLI**
   ```bash
   # Using npm
   npm install -g @google/gemini-cli
   
   # Or using Homebrew
   brew install gemini-cli
   ```

2. **Authenticate**
   ```bash
   gemini
   # Select "Login with Google" and follow browser OAuth flow
   ```

3. **Verify Installation**
   ```bash
   gemini --version
   ```

### Basic Navigation

```bash
# Navigate to your project
cd ~/projects/my-app

# Start Gemini CLI
gemini

# You'll see:
# > _ (cursor waiting for input)
```

### Non-Interactive Usage

```bash
# Quick query
gemini "explain the architecture of this codebase"

# JSON output for scripts
gemini -p "list all TypeScript files" --output-format json

# Resume and continue
gemini -r "latest" "fix the failing tests"
```

---

## Common Workflows

### 1. Code Exploration

**Understanding a New Codebase**

```
> What does this project do?

Gemini will:
1. Read README.md
2. Analyze package.json / dependencies
3. Explore directory structure
4. Summarize architecture
```

**Finding Specific Code**

```
> Find where user authentication is handled

Gemini will:
1. Search for auth-related files
2. Read relevant code
3. Explain the flow
4. Show key files: @src/auth/login.ts
```

**Explaining Complex Code**

```
> Explain the logic in @src/utils/cache.ts

Gemini will:
1. Read the file
2. Break down complex sections
3. Provide context
4. Suggest improvements
```

### 2. Code Modifications

**Making Edits**

```
> Add input validation to the login function

Gemini will:
1. Find the login function
2. Propose changes
3. Show diff for approval
4. Apply with your confirmation
```

**Refactoring**

```
> Refactor this code to use async/await instead of callbacks

Gemini will:
1. Identify callback patterns
2. Convert to async/await
3. Handle error cases
4. Test the changes
```

**Creating New Files**

```
> Create a middleware for JWT authentication

Gemini will:
1. Check existing middleware structure
2. Create authMiddleware.ts
3. Implement JWT verification
4. Add to exports
```

### 3. Debugging

**Finding Bugs**

```
> Why is this test failing?

> Find the source of this error: [paste error]

> Add logging to trace this issue
```

**Using Web Search**

```
> Search for the latest React best practices

> What is the current version of TypeScript?
```

### 4. Project Setup

**New Project**

```
> Set up a new React project with TypeScript

Gemini will:
1. Create project structure
2. Initialize package.json
3. Install dependencies
4. Set up build tools
5. Create initial files
```

**Adding Dependencies**

```
> Install and configure jest for testing

> Add TypeScript to this project

> Set up ESLint with recommended rules
```

### 5. Git Workflows

**Committing Changes**

```
> Commit these changes with a good message

Gemini will:
1. Run git status
2. Review diffs
3. Stage files
4. Create commit with descriptive message
```

**Using Custom Commands**

```bash
# Create a custom commit command
mkdir -p .gemini/commands/git
cat > .gemini/commands/git/commit.toml << 'EOF'
description = "Generate commit message from staged changes"
prompt = """
Generate a Conventional Commit message for these changes:

!{git diff --staged}
"""
EOF
```

Then use:
```
> /git:commit
```

---

## Advanced Features

### 1. Plan Mode

Break complex tasks into manageable steps:

```
> /plan

# Now in plan mode - Gemini will:
# 1. Analyze the task
# 2. Create a step-by-step plan
# 3. Get your approval
# 4. Execute each step
```

Configure plan mode:

```json
{
  "general": {
    "plan": {
      "modelRouting": true,
      "directory": "/tmp/gemini-plans"
    }
  }
}
```

### 2. GEMINI.md Files

Create project context files:

```markdown
<!-- GEMINI.md in project root -->
# My Project

## Tech Stack
- React 18
- TypeScript
- Vite
- Tailwind CSS

## Commands
- `npm run dev` - Start dev server
- `npm run build` - Production build
- `npm test` - Run tests

## Conventions
- Use functional components
- Prefer named exports
- Follow existing file structure
```

View loaded context:
```
> /memory show
```

### 3. Custom Commands

Create project-specific commands:

```toml
# .gemini/commands/test.toml
description = "Run tests with coverage"
prompt = """
Run the test suite with coverage and report any failures.

!{npm run test:coverage}
"""
```

Reload commands after changes:
```
> /commands reload
```

### 4. MCP Integration

Configure MCP servers in `~/.gemini/settings.json`:

```json
{
  "mcp": {
    "servers": {
      "github": {
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-github"],
        "env": {
          "GITHUB_TOKEN": "${GITHUB_TOKEN}"
        }
      }
    }
  }
}
```

Then use:
```
> List my open GitHub issues

> Create an issue for this bug
```

### 5. Sandboxing

Enable sandbox for secure execution:

```bash
# Run with Docker sandbox
gemini --sandbox

# Or in settings.json
{
  "tools": {
    "sandbox": true
  }
}
```

### 6. Session Checkpointing

Enable checkpointing for recovery:

```json
{
  "general": {
    "checkpointing": {
      "enabled": true
    }
  }
}
```

Resume sessions:
```bash
# List sessions
gemini --list-sessions

# Resume by ID
gemini -r "abc123"

# Resume latest
gemini -r "latest"
```

---

## Best Practices

### 1. Effective Communication

**Be Specific:**
```
# Good
> Add error handling for network timeouts in fetchUserData()

# Less effective
> Fix the error handling
```

**Provide Context:**
```
> This is a React app using Redux. Add a new action for user logout.
```

**Use References:**
```
> Update @src/components/Button.tsx to support a loading state
```

### 2. Managing Sessions

**Save Important Sessions:**
```bash
# Sessions are auto-saved
# Resume with:
gemini -r "latest"

# Or list and select:
gemini --list-sessions
```

**Compress Long Sessions:**
```
> /compress
# Summarizes older messages to save context space
```

### 3. Context Management

**Use GEMINI.md:**
- Create at project root for project-wide context
- Create in subdirectories for component-specific context
- Use `~/.gemini/GEMINI.md` for global preferences

**Control Context Size:**
```
> /memory refresh  # Reload context files
> /clear           # Clear conversation history
```

### 4. Security

**Review Before Approving:**
- Always review code changes
- Check shell commands before execution
- Be cautious with file deletions

**Use Approval Modes:**
```bash
# Default mode - prompts for file writes and shell
gemini --approval-mode default

# Auto-edit mode - auto-approves edits, prompts for shell
gemini --approval-mode auto_edit

# Plan mode - read-only, no tool execution
gemini --approval-mode plan
```

**Configure Trusted Folders:**
```json
{
  "security": {
    "folderTrust": {
      "enabled": true
    }
  }
}
```

### 5. Quota Management

**Monitor Usage:**
```
> Check my quota usage

> How many requests have I made today?
```

**Optimize Context:**
- Use `/compress` regularly
- Clear history with `/clear` when switching tasks
- Remove unnecessary files from context

---

## Troubleshooting

### Common Issues

**1. Authentication Problems**

```bash
# Clear cached credentials
rm -rf ~/.gemini/credentials

# Re-authenticate
gemini
```

**2. Session Won't Resume**

```bash
# List available sessions
gemini --list-sessions

# Or start fresh
gemini
```

**3. Tool Execution Fails**

```
# Check permissions
> Check the approval settings

# Review allowed/denied patterns in settings.json
```

**4. Out of Context Space**

```
> /compress
# Or
> /clear
# Or start new session
```

**5. MCP Server Not Working**

```bash
# Check MCP config
gemini mcp list

# Verify server is installed
npm list -g @server/package

# Check server logs
```

**6. Extension Not Loading**

```bash
# Validate extension
gemini extensions validate ./my-extension

# Check extension is enabled
gemini extensions list

# Reload extensions
# (use /commands reload in CLI)
```

### Getting Help

**Within Gemini CLI:**
```
> /help                    # Show commands
> How do I use X?         # Ask Gemini
> /bug description        # Report issue
```

**External Resources:**
- Documentation: https://geminicli.com/docs/
- GitHub Issues: https://github.com/google-gemini/gemini-cli/issues
- GitHub Discussions: https://github.com/google-gemini/gemini-cli/discussions

---

## Quick Reference Card

### Essential Commands

| Command | Purpose |
|---------|---------|
| `gemini` | Start interactive session |
| `gemini -p "query"` | Non-interactive mode |
| `gemini -r "latest"` | Resume session |
| `/help` | Show help |
| `/exit` | Quit |
| `/clear` | Clear chat |
| `/model` | Change model |
| `/memory` | Manage context |
| `/plan` | Enter plan mode |

### Context References

| Syntax | Meaning |
|--------|---------|
| `@file` | Reference file |
| `@dir/` | Reference directory |
| `!cmd` | Execute shell |
| `/command` | Slash command |

### Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl+C` | Cancel/Exit |
| `Ctrl+L` | Clear screen |
| `↑/↓` | Command history |
| `Tab` | Autocomplete |
| `Shift+Enter` | New line |

---

*For more details, see the [API Reference](./API.md) and [Architecture](./ARCHITECTURE.md) documentation.*
