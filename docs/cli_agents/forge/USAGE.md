# Forge - Usage Guide

## Table of Contents

1. [Getting Started](#getting-started)
2. [Common Workflows](#common-workflows)
3. [Agent-Specific Usage](#agent-specific-usage)
4. [Advanced Features](#advanced-features)
5. [Best Practices](#best-practices)
6. [Troubleshooting](#troubleshooting)

---

## Getting Started

### First-Time Setup

1. **Install Forge**
   ```bash
   curl -fsSL https://forgecode.dev/cli | sh
   ```

2. **Configure Provider**
   ```bash
   forge provider login
   # Select your provider (Anthropic, OpenAI, OpenRouter, etc.)
   # Enter your API key
   ```

3. **Select Model**
   ```bash
   forge model
   # Browse and select your preferred model
   ```

4. **Optional: Setup Zsh Plugin**
   ```bash
   forge zsh-setup
   # Restart your terminal
   ```

### Basic Usage

```bash
# Interactive mode
forge

# One-shot prompt
forge -p "explain the auth module"

# With specific model
forge -m claude-sonnet-4 -p "refactor this function"

# Using Zsh plugin
: explain how the database layer works
```

---

## Common Workflows

### 1. Code Exploration

**Understanding a New Codebase**

```bash
forge
> What does this project do?
```

Forge will:
1. Read AGENTS.md or README.md if present
2. Analyze the project structure
3. Identify key technologies and patterns
4. Provide a comprehensive overview

**Finding Specific Code**

```bash
# Find authentication logic
> Find where user authentication is handled

# Search for specific patterns
> Show me all database query implementations

# Understand data flow
> Trace how a request flows from API to database
```

**Analyzing Dependencies**

```bash
> Analyze the project's dependencies and architecture
> What are the main external libraries used?
> Show me the module hierarchy
```

### 2. Code Modifications

**Implementing Features**

```bash
# Add new functionality
> Add JWT authentication to the API

# Create new components
> Create a React component for user profiles

# Add API endpoints
> Add CRUD endpoints for the User resource
```

**Refactoring**

```bash
# Modernize code
> Refactor callback-based code to use async/await

# Improve structure
> Extract this logic into a separate service

# Optimize performance
> Optimize this database query
```

**Fixing Issues**

```bash
# Debug errors
> Fix this error: "cannot borrow as mutable more than once"

# Address warnings
> Fix all clippy warnings in this module

# Resolve type errors
> Fix the type mismatch in the auth module
```

### 3. Git Operations

**Commit Management**

```bash
# Generate commit message
> Write a conventional commit message for these changes

# Review before commit
> Review my changes before I commit

# Stage and commit
> Stage all changes and commit with a good message
```

**Pull Requests**

```bash
# Create PR description
> Create a pull request description for these changes

# Review PR
> Review this PR for potential issues

# Handle merge conflicts
> Help me resolve these merge conflicts
```

**Branch Management**

```bash
# Create feature branch
> Create a feature branch for user authentication

# Compare branches
> Show me the differences between main and this branch

# Clean up branches
> Delete merged branches
```

### 4. Testing Workflows

**Test Creation**

```bash
# Generate tests
> Add unit tests for the UserService

# Improve coverage
> Add missing test cases for edge cases

# Integration tests
> Create integration tests for the API endpoints
```

**Test Execution**

```bash
# Run tests
> Run the test suite

# Debug failures
> Why is this test failing?

# Specific tests
> Run only the auth module tests
```

### 5. Documentation

**Code Documentation**

```bash
# Add documentation
> Add rustdoc comments to this module

# Generate docs
> Generate API documentation

# Update README
> Update the README with the new features
```

**Architecture Documentation**

```bash
# Create diagrams
> Create an architecture diagram description

# Document decisions
> Document why we chose this database

> Update AGENTS.md with the new patterns
```

---

## Agent-Specific Usage

### Forge (Default Agent)

Best for: Implementation tasks, coding, file operations

```bash
forge
> Implement a rate limiter middleware
> Create a database migration for user profiles
> Refactor the error handling module
```

**Capabilities:**
- Read/write/patch files
- Execute shell commands
- Semantic search
- Full tool access

### Sage (Research Agent)

Best for: Codebase analysis, understanding, documentation

```bash
forge agent sage
> Analyze the overall architecture
> Find all places that use the User struct
> Explain the authentication flow
> Identify potential security issues
```

**Capabilities:**
- Deep codebase search
- Read-only analysis
- Pattern recognition
- Documentation generation

**Limitations:**
- Cannot modify files
- Cannot execute commands

### Muse (Planning Agent)

Best for: Strategic planning, task breakdown, design decisions

```bash
forge agent muse
> Create an implementation plan for OAuth2
> Design the database schema for a blog
> Plan the migration from REST to GraphQL
```

**Capabilities:**
- Create detailed plans
- Task breakdown
- Risk assessment
- Alternative approaches

**Limitations:**
- Cannot implement changes
- Advisory only

---

## Advanced Features

### 1. Custom Agents

Create specialized agents for your team:

```bash
# Create custom agent definition
mkdir -p .forge/agents
cat > .forge/agents/security-reviewer.md << 'EOF'
---
id: "security-reviewer"
title: "Security code reviewer"
description: "Specialized agent for security code reviews"
reasoning:
  enabled: true
tools:
  - sem_search
  - read
  - sage
user_prompt: |-
  <{{event.name}}>{{event.value}}</{{event.name}}>
  <system_date>{{current_date}}</system_date>
---

You are a security-focused code reviewer...
EOF
```

### 2. Custom Commands

Add project-specific commands:

```yaml
# forge.yaml
commands:
  - name: "test-coverage"
    description: "Run tests with coverage"
    prompt: "Run cargo test with coverage and generate report"
  
  - name: "security-audit"
    description: "Run security audit"
    prompt: "Run cargo audit and review security advisories"
  
  - name: "release-check"
    description: "Pre-release checks"
    prompt: "Run all checks: format, clippy, test, and audit"
```

Usage:
```bash
: test-coverage
: security-audit
: release-check
```

### 3. Context Management

**Using AGENTS.md**

Create context files for your project:

```markdown
# My Project

## Tech Stack
- Rust with Axum
- PostgreSQL with SQLx
- React frontend

## Commands
- `cargo run` - Start server
- `cargo test` - Run tests
- `cargo sqlx prepare` - Update query metadata

## Architecture
- `src/handlers/` - HTTP handlers
- `src/models/` - Database models
- `src/services/` - Business logic

## Conventions
- Use snake_case for functions
- Add tests for all handlers
- Use thiserror for errors
```

**Context Compaction**

When conversations get long:

```bash
# Automatic compaction happens at thresholds
# Or manually request:
> Summarize our conversation so far

# Start fresh but keep context
> Start a new task
```

### 4. MCP Integration

**Setup GitHub MCP**

```bash
# Create MCP config
mkdir -p ~/.config/forge
cat > ~/.config/forge/.mcp.json << 'EOF'
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@github/mcp-server"],
      "env": {
        "GITHUB_TOKEN": "your-token"
      }
    }
  }
}
EOF
```

Usage:
```bash
> List my open issues
> Create an issue for this bug
> Review pull request #123
```

**Setup Database MCP**

```bash
cat > ~/.config/forge/.mcp.json << 'EOF'
{
  "mcpServers": {
    "postgres": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-postgres"],
      "env": {
        "DATABASE_URL": "postgres://..."
      }
    }
  }
}
EOF
```

Usage:
```bash
> Show me the database schema
> Run a query to count users
> Analyze slow queries
```

### 5. Semantic Search

**Index Your Codebase**

```bash
# Automatic indexing on first use
# Or force reindex:
> Index the codebase for semantic search
```

**Search Usage**

```bash
# Natural language search
> Find code related to "user authentication"
> Show me where we handle database connections
> Find error handling patterns
```

### 6. Reasoning Mode

Enable extended thinking for complex problems:

```yaml
# forge.yaml
reasoning:
  enabled: true
  effort: "high"  # none, minimal, low, medium, high, xhigh, max
```

Usage:
```bash
> Design the architecture for a new microservice
# With reasoning enabled, Forge will think through the design
```

---

## Best Practices

### 1. Effective Communication

**Be Specific**

```bash
# Good
> Add input validation for email format in the register function

# Less effective
> Fix the validation
```

**Provide Context**

```bash
# Good
> This is a Rust Axum API. Add middleware for request logging.

# Less effective
> Add logging
```

**Reference Files**

```bash
# Good
> Update src/auth/middleware.rs to add JWT verification

# Less effective
> Update the auth code
```

### 2. Session Management

**Keep Sessions Focused**

```bash
# One task per session
> Implement user authentication
# Complete task, then start new session for next feature

# Or clear context between tasks
> /clear
```

**Use Meaningful Session Names**

Sessions are auto-named based on first prompt. Start with clear intent:

```bash
> Implement OAuth2 login with Google provider
# Session will be named "Implement OAuth2 login..."
```

### 3. Security

**Review Before Applying**

- Always review file changes before confirming
- Check shell commands before execution
- Be cautious with `remove` operations

**Use Restricted Mode for Sensitive Projects**

```bash
forge -r  # Enable restricted mode
```

**Validate Generated Code**

```bash
> Add error handling to this function
# After implementation:
> Run the tests to verify the changes
```

### 4. Cost Management

**Monitor Usage**

```bash
# Check token usage
> How many tokens have we used?

# Use cheaper models for simple tasks
forge -m gpt-4o-mini -p "fix this typo"
```

**Optimize Context**

```bash
# Clear unnecessary context
> /clear

# Compact long conversations
> Summarize what we've done
```

### 5. Team Collaboration

**Share AGENTS.md**

Commit AGENTS.md to your repository:

```bash
git add AGENTS.md
git commit -m "docs: Add AI assistant context"
```

**Standardize Commands**

Create shared custom commands in `forge.yaml`:

```yaml
commands:
  - name: "onboard"
    description: "Help new team members understand the project"
    prompt: "Explain this codebase to a new team member"
```

---

## Troubleshooting

### Common Issues

**1. Provider Authentication Fails**

```bash
# Check provider configuration
forge provider list

# Re-authenticate
forge provider login

# Check environment variables aren't overriding
env | grep -i forge
```

**2. Model Not Available**

```bash
# List available models
forge model list

# Check provider has model access
# Some models require specific subscriptions
```

**3. Context Too Large**

```bash
# Clear conversation
> Start a new task

# Or compact context
> Summarize our progress

# Check token usage
> Show token usage
```

**4. Tool Execution Fails**

```bash
# Check permissions
> Why can't you edit that file?

# Verify file exists
> Check if the file exists

# Check disk space
df -h
```

**5. MCP Server Not Working**

```bash
# List configured servers
forge mcp list

# Test server connection
forge mcp get <server-name>

# Check server logs
# Review .mcp.json syntax
```

**6. Slow Performance**

```bash
# Use faster model
forge -m gpt-4o-mini

# Reduce context
> Focus on just this file

# Check network connection
ping forgecode.dev
```

### Getting Help

**Within Forge**

```bash
# Get help
> How do I use the semantic search?

# Ask about capabilities
> What can you help me with?
```

**External Resources**

- Discord: https://discord.gg/kRZBPpkgwq
- GitHub Issues: https://github.com/antinomyhq/forge/issues
- Documentation: https://forgecode.dev/docs

---

## Quick Reference Card

### Essential Commands

| Command | Purpose |
|---------|---------|
| `forge` | Start interactive session |
| `forge -p "cmd"` | One-shot command |
| `forge -m <model>` | Use specific model |
| `forge provider login` | Configure provider |
| `forge model` | Select model |
| `forge agent sage` | Research mode |
| `forge agent muse` | Planning mode |
| `: <prompt>` | Zsh plugin quick prompt |

### Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl+C` | Cancel/Clear |
| `Ctrl+D` | Exit |
| `↑/↓` | Command history |
| `Tab` | Autocomplete |

---

*For more details, see [API Reference](./API.md) and [Architecture](./ARCHITECTURE.md).*
