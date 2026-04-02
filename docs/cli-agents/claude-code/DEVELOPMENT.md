# Claude Code - Development Guide

## Overview

This document covers development of **Claude Code plugins and extensions**, not the core Claude Code tool itself (which is closed-source).

---

## Plugin Development

### Getting Started

1. **Install Claude Code**
   ```bash
   curl -fsSL https://claude.ai/install.sh | bash
   ```

2. **Set Up Plugin Development Environment**
   ```bash
   # Create plugin directory
   mkdir -p my-plugin/.claude-plugin
   cd my-plugin
   ```

3. **Create Plugin Metadata**
   ```json
   // .claude-plugin/plugin.json
   {
     "name": "my-plugin",
     "version": "1.0.0",
     "description": "My custom Claude Code plugin",
     "author": "Your Name"
   }
   ```

### Plugin Structure

```
my-plugin/
├── .claude-plugin/
│   └── plugin.json           # Plugin metadata
├── commands/                 # Custom slash commands
│   └── my-command.md
├── agents/                   # Specialized agents
│   └── my-agent.md
├── skills/                   # Contextual skills
│   └── my-skill.md
├── hooks/                    # Event hooks
│   ├── hooks.json
│   └── my-hook.py
├── .mcp.json                # MCP server config
└── README.md                # Plugin documentation
```

---

## Creating Commands

Commands are Markdown files with YAML frontmatter:

```markdown
---
allowed-tools: Bash(*), Read(*), Edit(*)
description: Description shown in /help
---

## Context
Optional context gathering

- Current git status: !`git status`
- File contents: !`cat config.json`

## Your task
Instructions for Claude when command is invoked

1. Step one
2. Step two
3. Step three
```

### Command Best Practices

1. **Limit Allowed Tools**
   ```yaml
   # Good - specific tools
   allowed-tools: Bash(git:*), Read(*), Grep(*)
   
   # Avoid - too broad
   allowed-tools: Bash(*), Write(*), Edit(*)
   ```

2. **Use Dynamic Context**
   ```markdown
   ## Context
   - Current branch: !`git branch --show-current`
   - Last commit: !`git log -1 --oneline`
   ```

3. **Clear Instructions**
   - Be specific about steps
   - Define expected output
   - Handle edge cases

### Example Commands

**Git Workflow Command:**
```markdown
---
allowed-tools: Bash(git:*), Bash(gh:*), Read(*)
description: Create a commit and PR
---

## Context
- Current status: !`git status`
- Current branch: !`git branch --show-current`

## Your task
If on main branch:
1. Create new branch named "feature/" + commit message
2. Stage all changes
3. Create commit with descriptive message
4. Push to origin
5. Create PR using gh pr create

If on feature branch:
1. Stage all changes
2. Amend last commit
3. Force push
```

---

## Creating Agents

Agents are specialized AI workers for parallel tasks:

```markdown
---
model: claude-3-5-sonnet-20241022
description: Specialized agent description
---

## Your role
You are a [specific role]. Your task is to...

## Guidelines
1. Focus on [specific aspect]
2. Look for [specific patterns]
3. Report [specific findings]

## Output format
Provide results in this format:
- Summary: ...
- Details: ...
- Recommendations: ...
```

### Agent Types

**Analyzer Agent:**
```markdown
---
model: claude-3-5-sonnet
description: Code quality analyzer
---

Analyze code for:
- Security vulnerabilities
- Performance issues
- Anti-patterns
- Style violations

Provide specific line references and fix suggestions.
```

**Reviewer Agent:**
```markdown
---
model: claude-3-5-sonnet
description: PR reviewer
---

Review the changes for:
1. Logic errors
2. Missing tests
3. Documentation gaps
4. Breaking changes

Assign confidence score (0-100) for each finding.
```

---

## Creating Skills

Skills provide contextual knowledge:

```markdown
---
name: my-skill
description: Context for specific tasks
---

## When to use
This skill applies when:
- Working with [technology]
- Implementing [feature type]
- Reviewing [code type]

## Key concepts
1. Concept one
2. Concept two

## Best practices
- Practice one
- Practice two

## Common patterns
```[language]
// Example code pattern
```

## References
- Link to docs
- Link to examples
```

### Skill Activation

Skills can be:
- **Auto-invoked**: Based on file patterns
- **Manually triggered**: Via `/skill` command
- **Always active**: Global skills

---

## Creating Hooks

Hooks intercept events and modify behavior:

### Hook Types

**PreToolUse Hook:**
```python
#!/usr/bin/env python3
# hooks/security-hook.py
import json
import sys

def main():
    event = json.load(sys.stdin)
    tool = event.get('tool')
    params = event.get('params', {})
    
    # Block dangerous patterns
    if tool == 'Bash':
        command = params.get('command', '')
        dangerous = ['rm -rf /', '>:', 'mkfs']
        if any(d in command for d in dangerous):
            print(json.dumps({
                "decision": "block",
                "reason": "Potentially dangerous command detected"
            }))
            return 2
    
    # Allow with modification
    if tool == 'Write':
        path = params.get('file_path', '')
        if path.endswith('.env'):
            print(json.dumps({
                "decision": "allow",
                "note": "Writing to .env file - ensure secrets are not committed"
            }))
            return 0
    
    # Default allow
    print(json.dumps({"decision": "allow"}))
    return 0

if __name__ == '__main__':
    sys.exit(main())
```

**PostToolUse Hook:**
```python
#!/usr/bin/env python3
import json
import sys

def main():
    event = json.load(sys.stdin)
    tool = event.get('tool')
    result = event.get('result', {})
    
    # Log tool usage
    if tool == 'Bash':
        command = event.get('params', {}).get('command', '')
        print(f"[LOG] Executed: {command}", file=sys.stderr)
    
    # Modify output
    if tool == 'Read' and 'error' in result:
        print(json.dumps({
            "output": f"File read failed: {result['error']}\nSuggestion: Check file path"
        }))
    
    return 0

if __name__ == '__main__':
    sys.exit(main())
```

### Hook Configuration

```json
// hooks/hooks.json
{
  "hooks": [
    {
      "event": "PreToolUse",
      "script": "./hooks/security.py",
      "if": "tool_name == 'Bash' || tool_name == 'Write'"
    },
    {
      "event": "PostToolUse",
      "script": "./hooks/logger.sh"
    },
    {
      "event": "SessionStart",
      "script": "./hooks/welcome.py"
    },
    {
      "event": "Stop",
      "script": "./hooks/confirm-exit.py"
    }
  ]
}
```

---

## MCP Server Development

### Creating MCP Servers

Claude Code can use custom MCP servers:

**Python MCP Server:**
```python
# server.py
from mcp.server import Server
from mcp.types import TextContent

app = Server("my-server")

@app.tool()
def search_code(query: str) -> str:
    """Search codebase for pattern"""
    # Implementation
    return results

@app.resource("docs://{topic}")
def get_documentation(topic: str) -> str:
    """Get documentation for topic"""
    # Implementation
    return docs

if __name__ == "__main__":
    app.run()
```

**Configuration:**
```json
// .mcp.json
{
  "mcpServers": {
    "my-server": {
      "command": "python",
      "args": ["./server.py"]
    }
  }
}
```

---

## Testing Plugins

### Local Testing

1. **Install Plugin Locally:**
   ```bash
   # Create symlink or copy to project
   ln -s ~/my-plugin ./my-plugin
   ```

2. **Test Commands:**
   ```
   > /my-command
   ```

3. **Test Hooks:**
   ```
   # Trigger events that hooks intercept
   > !ls  # Triggers PreToolUse for Bash
   ```

### Debugging

**Enable Debug Logging:**
```bash
export CLAUDE_CODE_DEBUG=1
claude
```

**Hook Debugging:**
```python
import sys
# Add logging
print(f"Hook triggered: {event}", file=sys.stderr)
```

---

## Publishing Plugins

### Distribution Methods

1. **Git Repository:**
   ```bash
   # Users install via:
   git clone https://github.com/user/plugin.git
   ```

2. **NPM Package:**
   ```bash
   npm publish
   # Users install via:
   npm install -g your-plugin
   ```

3. **Marketplace:**
   - Submit to Claude Code marketplace (when available)
   - Include plugin.json with full metadata

### Plugin Manifest

```json
{
  "name": "plugin-name",
  "version": "1.0.0",
  "description": "Clear description",
  "author": "Your Name <email@example.com>",
  "license": "MIT",
  "repository": "https://github.com/user/repo",
  "homepage": "https://docs.example.com",
  "keywords": ["claude-code", "plugin", "category"],
  "commands": ["command1", "command2"],
  "agents": ["agent1", "agent2"],
  "hooks": ["hook1"],
  "skills": ["skill1"],
  "mcpServers": ["server1"],
  "requirements": {
    "claude-code": ">=2.0.0",
    "os": ["macos", "linux", "windows"]
  }
}
```

---

## Best Practices

### 1. Code Quality

- Follow consistent naming conventions
- Document all commands and agents
- Include usage examples
- Handle errors gracefully

### 2. Security

- Validate all inputs
- Sanitize file paths
- Don't expose secrets
- Use least-privilege permissions

### 3. Performance

- Minimize tool calls in hooks
- Cache expensive operations
- Use specific glob patterns
- Avoid blocking operations

### 4. User Experience

- Provide clear feedback
- Show progress for long tasks
- Offer undo when possible
- Document prerequisites

---

## Examples

### Complete Plugin Example

```
eslint-plugin/
├── .claude-plugin/
│   └── plugin.json
├── commands/
│   └── lint-project.md
├── agents/
│   └── rule-expert.md
├── skills/
│   └── eslint-patterns.md
├── hooks/
│   ├── hooks.json
│   └── pre-commit.py
└── README.md
```

**plugin.json:**
```json
{
  "name": "eslint-enhanced",
  "version": "1.0.0",
  "description": "Enhanced ESLint integration",
  "author": "Developer Name",
  "commands": ["lint-project"],
  "agents": ["rule-expert"],
  "skills": ["eslint-patterns"],
  "hooks": ["pre-commit"]
}
```

**commands/lint-project.md:**
```markdown
---
allowed-tools: Bash(npm:*), Bash(yarn:*), Bash(pnpm:*), Read(*)
description: Run ESLint with enhanced reporting
---

## Context
- Package manager: !`cat package.json | grep -E '"packageManager"|"lockfileVersion"' | head -1`
- ESLint config: !`ls -la .eslintrc* eslint.config.* 2>/dev/null || echo "No config found"`

## Your task
1. Detect package manager (npm/yarn/pnpm)
2. Run lint command
3. If errors found:
   - Categorize by rule
   - Suggest fixes for common issues
   - Offer to apply auto-fixes
4. Generate summary report
```

---

## Resources

- [Official Plugin Docs](https://docs.claude.com/en/docs/claude-code/plugins)
- [Agent SDK](https://docs.claude.com/en/api/agent-sdk/overview)
- [MCP Specification](https://modelcontextprotocol.io/)
- [Example Plugins](../../../cli-agents/claude-code/plugins/)

---

*For core Claude Code development (not plugins), see the [official repository](https://github.com/anthropics/claude-code).*
