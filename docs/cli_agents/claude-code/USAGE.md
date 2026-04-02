# Claude Code - Usage Guide

## Table of Contents

1. [Getting Started](#getting-started)
2. [Common Workflows](#common-workflows)
3. [Advanced Features](#advanced-features)
4. [Best Practices](#best-practices)
5. [Troubleshooting](#troubleshooting)

---

## Getting Started

### First-Time Setup

1. **Install Claude Code**
   ```bash
   # macOS/Linux
   curl -fsSL https://claude.ai/install.sh | bash
   
   # Windows
   irm https://claude.ai/install.ps1 | iex
   ```

2. **Authenticate**
   ```bash
   claude
   # Follow browser OAuth flow
   ```

3. **Verify Installation**
   ```bash
   claude --version
   ```

### Basic Navigation

```bash
# Navigate to your project
cd ~/projects/my-app

# Start Claude Code
claude

# You'll see:
# > _ (cursor waiting for input)
```

---

## Common Workflows

### 1. Code Exploration

**Understanding a New Codebase**

```
> What does this project do?

Claude will:
1. Read README.md
2. Analyze package.json / dependencies
3. Explore directory structure
4. Summarize architecture
```

**Finding Specific Code**

```
> Find where user authentication is handled

Claude will:
1. Search for auth-related files
2. Read relevant code
3. Explain the flow
4. Show key files: @src/auth/login.ts
```

**Explaining Complex Code**

```
> Explain the logic in @src/utils/cache.ts

Claude will:
1. Read the file
2. Break down complex sections
3. Provide context
4. Suggest improvements
```

### 2. Code Modifications

**Making Edits**

```
> Add input validation to the login function

Claude will:
1. Find the login function
2. Propose changes
3. Show diff for approval
4. Apply with your confirmation
```

**Refactoring**

```
> Refactor this code to use async/await instead of callbacks

Claude will:
1. Identify callback patterns
2. Convert to async/await
3. Handle error cases
4. Test the changes
```

**Creating New Files**

```
> Create a middleware for JWT authentication

Claude will:
1. Check existing middleware structure
2. Create authMiddleware.ts
3. Implement JWT verification
4. Add to exports
```

### 3. Git Workflows

**Committing Changes**

```
> Commit these changes with a good message

Claude will:
1. Run git status
2. Review diffs
3. Stage files
4. Create commit with descriptive message
```

**Creating Pull Requests**

```
> Create a PR for these changes

Claude will:
1. Create branch if on main
2. Commit changes
3. Push to origin
4. Open PR with gh pr create
```

**Reviewing Code**

```
> Review the changes in this branch

Claude will:
1. Show git diff
2. Analyze changes
3. Suggest improvements
4. Check for issues
```

### 4. Testing & Debugging

**Running Tests**

```
> Run the test suite

> Run tests for the auth module only

> Fix the failing test in users.test.ts
```

**Debugging**

```
> Why is this test failing?

> Find the source of this error: [paste error]

> Add logging to trace this issue
```

### 5. Project Setup

**New Project**

```
> Set up a new React project with TypeScript

Claude will:
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

---

## Advanced Features

### 1. Auto Mode

Enable autonomous execution:

```
> /config
# Enable auto mode in settings

# Now Claude can:
# - Auto-execute safe commands
# - Skip repetitive approvals
# - Work more efficiently
```

Configure auto mode permissions:

```json
{
  "autoMode": {
    "enabled": true,
    "allowEdits": true,
    "allowBash": true,
    "allowWrite": false
  }
}
```

### 2. Subagents

Spawn parallel agents for complex tasks:

```
> Analyze this codebase using subagents

Claude will:
1. Spawn multiple agents
2. Each analyzes different aspects
3. Aggregate results
4. Provide comprehensive summary
```

### 3. Custom Commands

Create project-specific commands in `.claude/commands/`:

```markdown
---
allowed-tools: Bash(npm:*), Read(*)
description: Run full test suite with coverage
---

## Your task
1. Run npm run test:coverage
2. Check coverage report
3. Report any failures
```

Use with: `/run-tests`

### 4. Hooks

Create custom behavior hooks:

**PreToolUse Hook Example:**
```python
#!/usr/bin/env python3
# .claude/hooks/pretooluse.py
import json
import sys

event = json.load(sys.stdin)

if event['tool'] == 'Bash':
    command = event['params']['command']
    # Block dangerous commands
    if 'rm -rf /' in command:
        print(json.dumps({"decision": "block", "reason": "Dangerous command"}))
        sys.exit(2)

print(json.dumps({"decision": "allow"}))
```

### 5. CLAUDE.md Files

Create project context files:

```markdown
<!-- CLAUDE.md in project root -->
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

Claude automatically reads CLAUDE.md files in:
- Project root
- Current working directory
- Parent directories

### 6. MCP Integration

Use external tools via MCP:

```json
// ~/.claude/.mcp.json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@github/mcp-server"]
    }
  }
}
```

Then use:
```
> List my open GitHub issues

> Create an issue for this bug
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
```
# Sessions are auto-saved
# Resume with:
claude --resume

# Or list and select:
> /resume
```

**Compact Long Sessions:**
```
> /compact
# Summarizes older messages to save context space
```

### 3. Security

**Review Before Approving:**
- Always review code changes
- Check bash commands before execution
- Be cautious with file deletions

**Use Protected Directories:**
```json
{
  "protectedDirectories": [
    "~/.ssh",
    "~/.aws",
    "~/.gnupg",
    "**/node_modules"
  ]
}
```

### 4. Cost Management

**Monitor Usage:**
```
> /usage
# Shows token usage and estimated cost
```

**Optimize Context:**
- Use `/compact` regularly
- Clear history with `/clear` when switching tasks
- Remove unnecessary files from context

---

## Troubleshooting

### Common Issues

**1. Authentication Problems**

```bash
# Clear cached credentials
rm -rf ~/.claude/credentials

# Re-authenticate
claude
```

**2. Session Won't Resume**

```bash
# List available sessions
claude --resume

# Or start fresh
claude --reset
```

**3. Tool Execution Fails**

```
# Check permissions
> /permissions

# Review allowed/denied patterns
```

**4. Out of Context Space**

```
> /compact
# Or
> /clear
# Or start new session
```

**5. MCP Server Not Working**

```bash
# Check MCP config
> /config

# Verify server is installed
npm list -g @server/package

# Check server logs
```

### Getting Help

**Within Claude Code:**
```
> /help                    # Show commands
> How do I use X?         # Ask Claude
> /bug description        # Report issue
```

**External Resources:**
- Documentation: https://code.claude.com/docs
- Discord: https://anthropic.com/discord
- GitHub Issues: https://github.com/anthropics/claude-code/issues

---

## Quick Reference Card

### Essential Commands

| Command | Purpose |
|---------|---------|
| `claude` | Start session |
| `claude --resume` | Resume session |
| `claude -p "cmd"` | Headless mode |
| `/help` | Show help |
| `/exit` | Quit |
| `/clear` | Clear chat |
| `/reset` | Reset session |
| `/config` | Settings |
| `/usage` | View usage |
| `/model` | Change model |

### Context References

| Syntax | Meaning |
|--------|---------|
| `@file` | Reference file |
| `@dir/` | Reference directory |
| `#symbol` | Reference symbol |
| `!cmd` | Execute bash |
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
