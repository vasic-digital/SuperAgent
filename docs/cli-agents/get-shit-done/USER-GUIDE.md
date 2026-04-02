# Get Shit Done (GSD) - User Guide

> A lightweight but powerful meta-prompting, context engineering, and spec-driven development system for AI coding assistants.

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

Get Shit Done (GSD) is a spec-driven development system designed to streamline building software with AI assistants like Claude Code, OpenCode, Gemini CLI, Codex, and others. It solves the "context rot" problem — the degradation of AI quality as a chat session grows — by structuring your ideas into precise, context-engineered steps that are researched, scoped, planned, executed, and verified.

### Key Features

- **Context Engineering**: Prevents AI quality degradation by isolating and refreshing context
- **Spec-Driven Development**: Structured multi-phase development lifecycle
- **Multi-Platform Support**: Works with Claude Code, OpenCode, Gemini CLI, Codex, Copilot, Cursor, Windsurf, and Antigravity
- **Atomic Commits**: Clean git history with standardized commits
- **Human-in-the-Loop**: Verification points at key stages
- **Sub-Agent Orchestration**: Uses researcher, planner, and executor agents
- **No Bureaucracy**: Focus on shipping without story points or sprint ceremonies

---

## Installation

### Prerequisites

- **Node.js 20+** (Node 20 LTS active until October 2026)
- **One of the supported AI platforms** (Claude Code, Copilot CLI, Codex CLI, etc.)

### Installation Methods

#### Option 1: Global Installation (Recommended)

Install for all supported platforms:

```bash
npx get-shit-done-cc@latest
```

#### Option 2: Platform-Specific Installation

**Claude Code:**
```bash
# Global install (all projects)
npx get-shit-done-cc --claude --global

# Local install (current project only)
npx get-shit-done-cc --claude --local
```

**OpenCode:**
```bash
# Global install
npx get-shit-done-cc --opencode --global

# Local install
npx get-shit-done-cc --opencode --local
```

**Gemini CLI:**
```bash
# Global install
npx get-shit-done-cc --gemini --global

# Local install
npx get-shit-done-cc --gemini --local
```

**Codex CLI:**
```bash
# Global install
npx get-shit-done-cc --codex --global

# Local install
npx get-shit-done-cc --codex --local
```

**GitHub Copilot CLI:**
```bash
# Global install
npx get-shit-done-cc --copilot --global

# Local install
npx get-shit-done-cc --copilot --local
```

**Cursor CLI:**
```bash
# Global install
npx get-shit-done-cc --cursor --global

# Local install
npx get-shit-done-cc --cursor --local
```

**Windsurf (Codeium):**
```bash
# Global install
npx get-shit-done-cc --windsurf --global

# Local install
npx get-shit-done-cc --windsurf --local
```

**Antigravity (Google):**
```bash
# Global install
npx get-shit-done-cc --antigravity --global

# Local install
npx get-shit-done-cc --antigravity --local
```

**Install for all platforms:**
```bash
npx get-shit-done-cc --all --global
```

#### Option 3: Install with SDK

For headless autonomous execution:

```bash
npx get-shit-done-cc --sdk
```

#### Option 4: Local Clone Installation

Clone the repository and run locally:

```bash
git clone https://github.com/gsd-build/get-shit-done.git
cd get-shit-done
node bin/install.js --claude --local
```

### Installation Locations

| Platform | Global Path | Local Path |
|----------|-------------|------------|
| Claude Code | `~/.claude/` | `./.claude/` |
| OpenCode | `~/.config/opencode/` | `./.config/opencode/` |
| Gemini CLI | `~/.gemini/` | `./.gemini/` |
| Codex CLI | `~/.codex/` | `./.codex/` |
| Copilot CLI | `~/.github/` | `./.github/` |
| Cursor | `~/.cursor/` | `./.cursor/` |
| Windsurf | `~/.windsurf/` | `./.windsurf/` |
| Antigravity | `~/.gemini/antigravity/` | `./.agent/` |

---

## Quick Start

### 1. Verify Installation

After installation, verify GSD is available:

**Claude Code / Copilot:**
```
/gsd:help
```

**OpenCode:**
```
/gsd-help
```

**Codex:**
```
$gsd-help
```

### 2. Start a New Project

**Claude Code / Copilot:**
```
/gsd:new-project
```

**OpenCode:**
```
/gsd-new-project
```

**Codex:**
```
$gsd-new-project
```

### 3. Follow the Workflow

1. State your project goals
2. Let GSD research and generate a roadmap
3. Discuss and plan each phase
4. Execute with atomic commits
5. Verify the work

### 4. Quick Fix Mode

For quick bug fixes without full planning:

**Claude Code / Copilot:**
```
/gsd:quick
```

**OpenCode:**
```
/gsd-quick
```

---

## CLI Commands

### Installation Flags

| Flag | Description |
|------|-------------|
| `--claude` | Install for Claude Code |
| `--opencode` | Install for OpenCode |
| `--gemini` | Install for Gemini CLI |
| `--codex` | Install for Codex CLI |
| `--copilot` | Install for GitHub Copilot CLI |
| `--cursor` | Install for Cursor CLI |
| `--windsurf` | Install for Windsurf |
| `--antigravity` | Install for Antigravity |
| `--all` | Install for all platforms |
| `--global` | Install globally (all projects) |
| `--local` | Install locally (current project only) |
| `--sdk` | Also install GSD SDK CLI (`gsd-sdk`) |

### SDK CLI Commands (if installed with --sdk)

```bash
# Initialize new project
gsd-sdk init

# List available templates
gsd-sdk templates

# Run autonomous execution
gsd-sdk run --autonomous
```

### Update Installation

```bash
# Update to latest version
npx get-shit-done-cc@latest
```

---

## Interactive Commands

### Command Prefixes by Platform

| Platform | Command Prefix |
|----------|----------------|
| Claude Code | `/gsd:` |
| OpenCode | `/gsd-` |
| Codex CLI | `$gsd-` |
| GitHub Copilot | `/gsd:` |
| Cursor | `/gsd:` |
| Windsurf | `/gsd:` |
| Antigravity | `/gsd:` |

### Core Workflow Commands

#### Project Initialization

**Claude/Copilot:**
```
/gsd:new-project
```

**OpenCode:**
```
/gsd-new-project
```

**Codex:**
```
$gsd-new-project
```

Starts a new project with guided setup including:
- Project goal definition
- Research phase
- Requirements extraction
- Roadmap creation

#### Phase Discussion

**Claude/Copilot:**
```
/gsd:discuss-phase <N>
```

**OpenCode:**
```
/gsd-discuss-phase <N>
```

**Codex:**
```
$gsd-discuss-phase <N>
```

Explore approach for phase N through guided questions before implementation.

#### Phase Research

**Claude/Copilot:**
```
/gsd:research-phase <N>
```

Research implementation patterns and best practices for phase N.

#### Phase Planning

**Claude/Copilot:**
```
/gsd:plan-phase <N>
```

**OpenCode:**
```
/gsd-plan-phase <N>
```

**Codex:**
```
$gsd-plan-phase <N>
```

Create atomic execution plans with XML sub-tasks for phase N.

#### Phase Execution

**Claude/Copilot:**
```
/gsd:execute-phase <N>
```

**OpenCode:**
```
/gsd-execute-phase <N>
```

**Codex:**
```
$gsd-execute-phase <N>
```

Execute all plans for phase N with fresh context:
- Sub-agents write code
- Perform checks
- Commit atomically

#### Work Verification

**Claude/Copilot:**
```
/gsd:verify-work
```

**OpenCode:**
```
/gsd-verify-work
```

**Codex:**
```
$gsd-verify-work
```

Manual acceptance testing loop allowing visual verification of the feature block.

#### Complete Milestone

**Claude/Copilot:**
```
/gsd:complete-milestone
```

**OpenCode:**
```
/gsd-complete-milestone
```

Archive current milestone and prepare for the next one.

#### Autonomous Mode

**Claude/Copilot:**
```
/gsd:autonomous
```

Advance directly through available phases without manual intervention.

### Quick Commands

#### Quick Fix

**Claude/Copilot:**
```
/gsd:quick
```

**OpenCode:**
```
/gsd-quick
```

Quick bug fix without the full planning bureaucracy.

#### Help

**Claude/Copilot:**
```
/gsd:help
```

**OpenCode:**
```
/gsd-help
```

**Codex:**
```
$gsd-help
```

Display help information about GSD commands.

---

## Configuration

### Project Structure

After initialization, GSD creates the following structure:

```
project-root/
├── .gsd/                     # GSD configuration directory
│   ├── roadmap.md            # Project roadmap
│   ├── phases/               # Phase definitions
│   │   ├── phase-01.md
│   │   ├── phase-02.md
│   │   └── ...
│   ├── specs/                # Technical specifications
│   └── state.json            # Current project state
├── .claude/                  # Claude-specific files (if using Claude)
│   └── skills/               # GSD skills
│       └── gsd-*/
│           └── SKILL.md
└── ...
```

### Roadmap Format

The `roadmap.md` file contains:

```markdown
# Project Roadmap

## Goal
Brief description of the project goal.

## Phases

### Phase 1: Foundation
- [ ] Setup project structure
- [ ] Configure build system
- [ ] Set up testing framework

### Phase 2: Core Features
- [ ] Implement authentication
- [ ] Create user management
- [ ] Add basic CRUD operations

### Phase 3: Polish
- [ ] Error handling
- [ ] Documentation
- [ ] Performance optimization
```

### Phase Specification Format

Each phase file (e.g., `phases/phase-01.md`):

```markdown
# Phase 1: Foundation

## Objective
Set up the project foundation with proper structure and tooling.

## Requirements
- TypeScript configuration
- Testing framework (Jest)
- Linting (ESLint + Prettier)
- Build system (Vite)

## Implementation Plan
1. Initialize npm project
2. Install TypeScript and configure tsconfig.json
3. Setup Jest for testing
4. Configure ESLint and Prettier
5. Setup Vite build configuration

## Verification Criteria
- [ ] `npm run build` succeeds
- [ ] `npm test` passes
- [ ] `npm run lint` passes
- [ ] TypeScript compiles without errors
```

### State File

The `state.json` file tracks progress:

```json
{
  "currentPhase": 1,
  "completedPhases": [],
  "currentTask": "setup-typescript",
  "context": {
    "discussed": true,
    "researched": true,
    "planned": true,
    "executed": false,
    "verified": false
  }
}
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `GSD_DEBUG` | Enable debug logging |
| `GSD_NO_COLOR` | Disable colored output |
| `GSD_AUTO_COMMIT` | Automatically commit changes |

---

## Usage Examples

### Example 1: New Web Application

```bash
# Start new project
/gsd:new-project

# Answer prompts:
# - Project name: "Task Manager App"
# - Description: "A simple task management application with React and Node.js"
# - Tech stack: React, TypeScript, Express, MongoDB

# GSD generates roadmap with phases:
# Phase 1: Project Setup
# Phase 2: Backend API
# Phase 3: Frontend UI
# Phase 4: Integration & Testing

# Work through phases
/gsd:discuss-phase 1
/gsd:plan-phase 1
/gsd:execute-phase 1
/gsd:verify-work 1

# Continue to next phase
/gsd:discuss-phase 2
/gsd:plan-phase 2
/gsd:execute-phase 2
/gsd:verify-work 2
```

### Example 2: Adding a Feature

```bash
# For an existing GSD project

# Discuss the new feature
/gsd:discuss-phase 3
# "I want to add user authentication with JWT tokens"

# GSD suggests:
# - Login/Register endpoints
# - Password hashing
# - JWT middleware
# - Protected routes

# Plan the implementation
/gsd:plan-phase 3

# Review the XML sub-tasks created

# Execute the plan
/gsd:execute-phase 3

# Verify the implementation
/gsd:verify-work 3
```

### Example 3: Quick Bug Fix

```bash
# Navigate to the file with the bug

# Use quick mode
/gsd:quick

# Describe the bug:
# "The login form doesn't show error messages when authentication fails"

# GSD will:
# - Analyze the code
# - Identify the issue
# - Propose a fix
# - Apply the fix with your approval
```

### Example 4: Refactoring

```bash
# Plan a refactoring task
/gsd:plan-phase 2

# Describe:
# "Refactor the database layer to use Repository pattern instead of direct queries"

# GSD creates sub-tasks:
# 1. Create UserRepository class
# 2. Create TaskRepository class
# 3. Update all database calls in controllers
# 4. Write tests for repositories
# 5. Remove old direct query code

# Execute with atomic commits
/gsd:execute-phase 2
```

### Example 5: Autonomous Mode

```bash
# For well-defined tasks, use autonomous mode
/gsd:autonomous

# GSD will:
# 1. Check current phase status
# 2. Automatically discuss if needed
# 3. Research if needed
# 4. Plan if needed
# 5. Execute all pending tasks
# 6. Verify the work
# 7. Move to next phase

# You can intervene at any verification point
```

---

## Troubleshooting

### Installation Issues

#### "npx command not found"

```bash
# Ensure Node.js 20+ is installed
node --version

# Install Node.js from https://nodejs.org/
# Or use a version manager:
nvm install 20
nvm use 20
```

#### "Installation fails with permission error"

```bash
# Use --local flag for local installation
npx get-shit-done-cc --claude --local

# Or fix npm permissions
sudo chown -R $(whoami) ~/.npm
```

#### "Platform not detected"

```bash
# Explicitly specify the platform
npx get-shit-done-cc --claude --global
```

### Runtime Issues

#### "Command not recognized"

1. Verify GSD is installed:
   ```bash
   ls -la ~/.claude/skills/  # For Claude Code
   ```

2. Check installation location matches your platform

3. Restart your AI assistant

#### "Phase execution fails"

```bash
# Check the roadmap file exists
cat .gsd/roadmap.md

# Verify state file is valid
cat .gsd/state.json

# Reset state if needed
rm .gsd/state.json
/gsd:new-project  # Re-initialize
```

#### "Context is too large"

```bash
# Use /compact command in your AI assistant
# Or start a new session
/gsd:complete-milestone
```

### Git Issues

#### "Git commit fails"

```bash
# Ensure git is configured
git config --global user.name "Your Name"
git config --global user.email "your@email.com"

# Check for merge conflicts
git status
```

#### "Changes not being committed atomically"

- Check `GSD_AUTO_COMMIT` is not set to false
- Verify git is initialized: `git init`
- Ensure there are staged changes: `git add .`

### Model Issues

#### "AI doesn't follow the GSD format"

1. Verify GSD skills are properly installed
2. Check the SKILL.md files exist in the skills directory
3. Restart the AI assistant to reload skills

#### "Sub-agents don't work correctly"

- Ensure your AI platform supports skill/commands
- Check the platform-specific prefix is correct
- Verify the agent has access to all required tools

### Getting Help

```bash
# Show help
/gsd:help

# Check documentation
# https://github.com/gsd-build/get-shit-done
```

### Reporting Issues

1. Check existing issues at https://github.com/gsd-build/get-shit-done/issues
2. Include:
   - Platform (Claude Code, Codex, etc.)
   - Node.js version
   - Installation method used
   - Error messages
   - Steps to reproduce

---

## Best Practices

### 1. Keep Phases Small

Break work into small, manageable phases that can be completed in 1-2 hours.

### 2. Verify Before Continuing

Always use `/gsd:verify-work` before marking a phase complete.

### 3. Use Descriptive Goals

When starting a project, be specific about what you want to achieve.

### 4. Commit Messages

GSD uses conventional commits. Review the generated messages before confirming.

### 5. Context Management

If the AI seems to be losing context, complete the current phase and start fresh.

---

## Resources

- **GitHub Repository**: https://github.com/gsd-build/get-shit-done
- **Documentation**: See repository README
- **Community**: GitHub Discussions

---

*Last updated: 2026-04-02*
