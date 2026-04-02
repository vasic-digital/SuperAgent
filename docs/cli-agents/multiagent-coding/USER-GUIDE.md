# Multi-Agent Coding - User Guide

> Orchestrate multiple AI coding agents working in parallel for faster, more reliable software development.

## Table of Contents

- [Overview](#overview)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [CLI Commands](#cli-commands)
- [Interactive Commands](#interactive-commands)
- [Configuration](#configuration)
- [Usage Examples](#usage-examples)
- [Troubleshooting](#troubleshooting)

---

## Overview

Multi-Agent Coding systems enable you to run multiple AI coding agents simultaneously, either working on different aspects of a project or approaching the same problem from multiple angles. This approach provides faster iteration, better code quality through comparison, and more comprehensive solutions.

### Key Features

- **Parallel Execution**: Run multiple agents simultaneously
- **Hierarchical Coordination**: Orchestrator manages worker agents
- **Role-Based Agents**: Architect, Developer, Reviewer roles
- **Context Isolation**: Each agent works with fresh context
- **Git Worktree Integration**: Isolated workspaces for conflict-free parallel work
- **Reflexion Pattern**: Self-critique loops for quality improvement
- **AI-Driven Evaluation**: Automated quality assessment of agent outputs

### Available Tools

| Tool | Description |
|------|-------------|
| `multiagent-code` | Agent-based toolsuite for Claude Code |
| `/fleet` (Copilot) | GitHub Copilot's parallel subagent execution |
| `cmux` | Run multiple coding agents in parallel |
| `aiswarm` | AI Swarm Agent Launcher with MCP coordination |

---

## Installation

### Prerequisites

- **Node.js 18+**
- **Git** with worktree support
- **One of**: Claude Code, GitHub Copilot CLI, or other supported AI assistants

### multiagent-code Installation

```bash
# Clone and install
git clone https://github.com/willsmanley/multiagent-code.git
chmod +x multiagent-code/install.sh
./multiagent-code/install.sh
```

This installs:
- Shell scripts for job management
- Orchestration prompts
- Claude Command `/orchestrator`

### cmux Installation

```bash
# Install cmux for parallel agent execution
npm install -g cmux

# Or from source
git clone https://github.com/yourusername/cmux.git
cd cmux
npm install
npm link
```

### aiswarm Installation

```bash
# Clone repository
git clone https://github.com/mrlarson2007/aiswarm.git
cd aiswarm

# Install dependencies
npm install

# Build
npm run build

# Link globally
npm link
```

### GitHub Copilot CLI

Fleet mode is built-in to Copilot CLI:
```bash
npm install -g @github/copilot
```

---

## Quick Start

### 1. Verify Installation

```bash
# Check multiagent-code
which orchestrator

# Check cmux
cmux --version

# Check aiswarm
aiswarm --version
```

### 2. Start Interactive Mode

**With Claude Code:**
```bash
claude --dangerously-skip-permissions

# Then use:
/orchestrator
```

**With Copilot CLI:**
```bash
copilot

# Then use:
/fleet
```

### 3. First Multi-Agent Task

**Example with Copilot /fleet:**
```
/fleet Create a REST API with the following:
1. User authentication endpoints (depends on: none)
2. Database models for users and posts (depends on: none)
3. CRUD endpoints for posts (depends on: 1, 2)
4. API documentation (depends on: 3)
```

---

## CLI Commands

### multiagent-code Commands

#### Installation

```bash
# Install the system
multiagent-code/install.sh

# Options:
# --claude    Install for Claude Code only
# --global    Install globally
# --local     Install locally
```

#### Orchestrator

```bash
# Start orchestrator (inside Claude Code)
/orchestrator

# With options
/orchestrator --max-workers 5 --timeout 300
```

### cmux Commands

```bash
# Run multiple agents
cmux run --agents claude,codex --task "Implement auth system"

# List running agents
cmux list

# Stop all agents
cmux stop

# View logs
cmux logs --agent claude
```

#### cmux Flags

| Flag | Description |
|------|-------------|
| `--agents <list>` | Comma-separated agent list |
| `--task <text>` | Task description |
| `--workspace <path>` | Working directory |
| `--timeout <seconds>` | Task timeout |
| `--parallel` | Run in parallel |
| `--sequential` | Run sequentially |

### aiswarm Commands

```bash
# Launch agent swarm
aiswarm launch --config swarm-config.json

# List active agents
aiswarm list

# Check agent status
aiswarm status --agent <name>

# Terminate agent
aiswarm terminate --agent <name>
```

#### aiswarm Launch Options

| Option | Description |
|--------|-------------|
| `--config <file>` | Configuration file |
| `--persona <name>` | Built-in persona |
| `--workspace <path>` | Git worktree path |
| `--model <name>` | LLM model to use |
| `--mcp <servers>` | MCP servers to connect |

### GitHub Copilot /fleet

```bash
# Start Copilot
copilot

# Use fleet command
/fleet <task description with dependencies>
```

#### Fleet Syntax

```
/fleet <task description>:
1. <subtask> (depends on: <deps>)
2. <subtask> (depends on: <deps>)
```

---

## Interactive Commands

### multiagent-code Orchestrator

#### Inside Claude Code

| Command | Description |
|---------|-------------|
| `/orchestrator` | Start multi-agent orchestration |
| `/orchestrator init` | Initialize new orchestration |
| `/orchestrator status` | Check running jobs |
| `/orchestrator cancel` | Cancel all jobs |

#### Orchestrator Workflow

```bash
# 1. Initialize
/orchestrator init

# 2. Describe task
# "Refactor the database layer to use Repository pattern"

# 3. Orchestrator breaks down task:
# - Create UserRepository
# - Create PostRepository  
# - Update UserController (depends on: UserRepository)
# - Update PostController (depends on: PostRepository)
# - Write tests (depends on: all)

# 4. Monitor progress
/orchestrator status
```

### cmux Interactive

```bash
# Start cmux TUI
cmux tui

# Commands in TUI:
p        # Pause/resume agent
s        # Stop agent
l        # View logs
q        # Quit
```

### aiswarm Interactive

```bash
# Start MCP server
aiswarm mcp-server

# In another terminal, connect via MCP client
# Tools available:
- create_task
- assign_agent
- get_agent_status
- terminate_agent
```

### Copilot /fleet Interactive

```bash
# Start fleet mode
copilot
/fleet <task>

# While running:
Ctrl+C     # Cancel current operation
/tasks     # View background tasks
/context   # View context usage
```

---

## Configuration

### multiagent-code Configuration

Configuration stored in:
```
~/.multiagent-code/config.json
```

Example:
```json
{
  "maxWorkers": 5,
  "timeout": 300,
  "roles": {
    "orchestrator": {
      "model": "claude-3-opus",
      "systemPrompt": "You are an orchestrator..."
    },
    "manager": {
      "model": "claude-3-sonnet",
      "systemPrompt": "You are a manager..."
    },
    "worker": {
      "model": "claude-3-haiku",
      "systemPrompt": "You are a worker..."
    }
  },
  "reflexion": {
    "enabled": true,
    "maxIterations": 3
  }
}
```

### cmux Configuration

```bash
# Config file
~/.cmux/config.json
```

Example:
```json
{
  "agents": [
    {
      "name": "claude",
      "command": "claude",
      "args": ["--dangerously-skip-permissions"]
    },
    {
      "name": "codex",
      "command": "codex",
      "args": []
    }
  ],
  "defaults": {
    "timeout": 600,
    "parallel": true
  }
}
```

### aiswarm Configuration

```json
{
  "swarm": {
    "coordinator": "hierarchical",
    "maxAgents": 5
  },
  "agents": [
    {
      "name": "planner",
      "persona": "architect",
      "model": "claude-3-opus"
    },
    {
      "name": "implementer-1",
      "persona": "developer",
      "model": "claude-3-sonnet"
    },
    {
      "name": "implementer-2",
      "persona": "developer",
      "model": "claude-3-sonnet"
    },
    {
      "name": "reviewer",
      "persona": "reviewer",
      "model": "claude-3-opus"
    }
  ],
  "mcp": {
    "servers": [
      "filesystem",
      "git"
    ]
  }
}
```

### Persona Definitions

**architect.md:**
```markdown
---
name: architect
description: System architecture specialist
---

You are a system architect. Your role is to:
- Design high-level system architecture
- Define component interfaces
- Plan data models
- Consider scalability and performance

Always provide detailed design documents before implementation.
```

**developer.md:**
```markdown
---
name: developer
description: Implementation specialist
---

You are a developer. Your role is to:
- Implement features according to specifications
- Write clean, tested code
- Follow existing code patterns
- Ask for clarification when needed
```

**reviewer.md:**
```markdown
---
name: reviewer
description: Code quality specialist
---

You are a code reviewer. Your role is to:
- Review code for bugs and issues
- Check for security vulnerabilities
- Ensure code follows best practices
- Provide constructive feedback
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `MULTIAGENT_MAX_WORKERS` | Maximum parallel agents |
| `MULTIAGENT_TIMEOUT` | Default timeout (seconds) |
| `MULTIAGENT_LOG_LEVEL` | Logging verbosity |
| `CMUX_CONFIG` | cmux config file path |
| `AISWARM_CONFIG` | aiswarm config file path |

---

## Usage Examples

### Example 1: Parallel Feature Development (Copilot /fleet)

```
/fleet Implement user management system:

1. Create database schema for users table (depends on: none)
2. Implement User model with validation (depends on: 1)
3. Create registration API endpoint (depends on: 2)
4. Create login API endpoint (depends on: 2)
5. Implement password reset flow (depends on: 2)
6. Write API documentation (depends on: 3, 4, 5)

Use default agent for tasks 1, 2
Use @security-specialist.md for task 5
```

### Example 2: Multi-Agent Code Review

```bash
# With cmux
cmux run \
  --agents claude,codex,gemini \
  --task "Review src/auth.js for security issues" \
  --parallel

# Each agent reviews independently
# Results are compared and consolidated
```

### Example 3: Repository Refactoring

```bash
# With multiagent-code
/orchestrator

# Task: "Refactor monolithic app to microservices"

# Orchestrator creates plan:
# Phase 1: Extract user service
# Phase 2: Extract order service  
# Phase 3: Extract inventory service
# Phase 4: Setup API gateway
# Phase 5: Update deployments

# Each phase spawns:
# - 1 Architect agent (design)
# - 2 Developer agents (implement)
# - 1 Reviewer agent (validate)
```

### Example 4: Multi-Model Solution Comparison

```bash
# With cmux
cmux run \
  --agents claude-sonnet,claude-opus,gpt-4o \
  --task "Implement a red-black tree in Python" \
  --workspace ./algorithm-challenge \
  --parallel

# Each model implements independently
# Results compared for:
# - Correctness
# - Performance
# - Code clarity
```

### Example 5: AI Swarm with MCP

```bash
# Start MCP server
aiswarm mcp-server

# In MCP client (e.g., Claude Desktop):
# Use tools:
# - create_task: "Build React components"
# - assign_agent: planner → design components
# - assign_agent: implementer-1 → build Button
# - assign_agent: implementer-2 → build Input
# - assign_agent: reviewer → review all

# Monitor progress via get_agent_status
```

### Example 6: Hierarchical Task Delegation

```bash
# Configuration
{
  "coordination": "hierarchical",
  "roles": ["architect", "developer", "reviewer"]
}

# Workflow:
# 1. Architect designs API
# 2. Architect delegates to 3 developers:
#    - Developer 1: Authentication endpoints
#    - Developer 2: CRUD endpoints
#    - Developer 3: Documentation
# 3. Reviewer validates all outputs
# 4. Architect approves final design
```

### Example 7: Cross-Platform Development

```bash
# With cmux - develop for multiple platforms
cmux run \
  --agents claude,codex \
  --task "Create Todo app" \
  --workspace ./todo-app

# Agent 1 (claude): Build iOS version
# Agent 2 (codex): Build Android version
# Both work in parallel in separate worktrees
```

---

## Troubleshooting

### Installation Issues

#### "install.sh permission denied"

```bash
# Make executable
chmod +x multiagent-code/install.sh

# Run with bash
bash multiagent-code/install.sh
```

#### "command not found after installation"

```bash
# Check PATH
echo $PATH

# Add to PATH
export PATH="$PATH:$HOME/.local/bin"

# Or use full path
~/.multiagent-code/bin/orchestrator
```

### Runtime Issues

#### "Too many agents, resource exhausted"

```bash
# Reduce max workers
export MULTIAGENT_MAX_WORKERS=3

# Or in config:
{
  "maxWorkers": 3
}
```

#### "Git worktree creation failed"

```bash
# Check git version
git --version  # Need 2.15+

# Clean up existing worktrees
git worktree list
git worktree remove <path>

# Check disk space
df -h
```

#### "Agent communication failed"

```bash
# Check if agents are running
cmux list

# Restart agents
cmux stop
cmux run --agents claude,codex --task "..."

# Check logs
cmux logs --agent claude
```

### Configuration Issues

#### "Invalid config file"

```bash
# Validate JSON
jsonlint ~/.multiagent-code/config.json

# Reset to defaults
rm ~/.multiagent-code/config.json
# Re-run install
```

#### "Persona not found"

```bash
# Check persona directory
ls ~/.multiagent-code/personas/

# Create default personas
multiagent-code/setup-personas
```

### Agent-Specific Issues

#### "Claude Code not responding"

```bash
# Check Claude is installed
which claude

# Restart Claude
pkill -f claude
claude

# Check permissions
claude --dangerously-skip-permissions --help
```

#### "Copilot CLI auth expired"

```bash
# Re-authenticate
copilot /login
```

#### "Model rate limit exceeded"

```bash
# Add delays between requests
# Use different models for different agents
{
  "agents": [
    {"model": "claude-3-opus"},    # Premium for architect
    {"model": "claude-3-haiku"},   # Cheaper for workers
    {"model": "claude-3-haiku"}    # Cheaper for workers
  ]
}
```

### Debug Mode

```bash
# Enable debug logging
export MULTIAGENT_LOG_LEVEL=debug

# Verbose output
cmux run --verbose --agents claude --task "test"

# With stack traces
DEBUG=* cmux run --agents claude --task "test"
```

### Getting Help

```bash
# Show help
orchestrator --help
cmux --help
aiswarm --help

# Check version
orchestrator --version
cmux --version
aiswarm --version
```

### Reporting Issues

1. Check existing issues on GitHub
2. Include:
   - Tool version
   - AI assistant version (Claude, Copilot, etc.)
   - Configuration file
   - Log output (with MULTIAGENT_LOG_LEVEL=debug)
   - Steps to reproduce

---

## Best Practices

### 1. Define Clear Dependencies

```
# Good:
1. Create database schema (depends on: none)
2. Implement models (depends on: 1)
3. Create API endpoints (depends on: 2)

# Bad:
1. Do everything
2. Do more things
```

### 2. Use Appropriate Granularity

Break tasks into 1-2 hour chunks. Too small = overhead, too large = context issues.

### 3. Assign Appropriate Roles

```
- Complex architecture → Architect persona
- Implementation → Developer persona  
- Security review → Reviewer persona
- Documentation → Technical writer persona
```

### 4. Monitor Resource Usage

```bash
# Watch system resources
htop

# Limit concurrent agents if needed
export MULTIAGENT_MAX_WORKERS=2
```

### 5. Clean Up Worktrees

```bash
# After completion
git worktree prune

# Or automatically
cmux run --cleanup --agents claude --task "..."
```

---

## Resources

- **multiagent-code**: https://github.com/willsmanley/multiagent-code
- **cmux**: https://github.com/yourusername/cmux
- **aiswarm**: https://github.com/mrlarson2007/aiswarm
- **GitHub Copilot Fleet**: https://github.blog/ai-and-ml/github-copilot/run-multiple-agents-at-once-with-fleet-in-copilot-cli/

---

*Last updated: 2026-04-02*
