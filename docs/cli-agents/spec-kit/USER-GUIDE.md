# Spec Kit User Guide

## Overview

Spec Kit (also known as Specify CLI) is GitHub's open-source toolkit for Spec-Driven Development (SDD). It provides a structured workflow for AI-assisted software development, enabling developers to create comprehensive specifications, implementation plans, and task lists before writing code. Spec Kit integrates with multiple AI coding assistants including Claude Code, GitHub Copilot, Cursor, Gemini CLI, and many others.

**Key Features:**
- Structured specification workflow (Constitution → Spec → Plan → Tasks → Implement)
- Multi-agent support (25+ AI assistants)
- Agent-agnostic design
- GitHub Issues integration
- Specification validation and analysis
- Quality checklists
- Branch management
- Template-based generation

---

## Installation Methods

### Method 1: Using uv (Recommended)

The modern Python package manager:

```bash
# Install uv first
curl -sSL https://astral.sh/uv/install.sh | sh

# Install Spec Kit
uv tool install specify-cli --from git+https://github.com/github/spec-kit.git
```

### Method 2: Using pip

```bash
pip install specify-cli

# Or install from source
git clone https://github.com/github/spec-kit.git
cd spec-kit
pip install -e .
```

### Method 3: Using pipx (Isolated Install)

```bash
pipx install git+https://github.com/github/spec-kit.git
```

### Method 4: Clone and Run

```bash
git clone https://github.com/github/spec-kit.git
cd spec-kit
python -m specify --help
```

### Verify Installation

```bash
specify --version
specify --help
```

---

## Quick Start

### 1. Initialize a Project

```bash
# Create new project directory
specify init my-project
cd my-project

# Or initialize in current directory
specify init .
# or
specify init --here
```

### 2. Select Your AI Assistant

During initialization, you'll be prompted to select your AI assistant:

```bash
# Or specify directly
specify init my-project --ai claude
specify init my-project --ai copilot
specify init my-project --ai cursor-agent
```

### 3. Establish Project Principles

Use the constitution command to define project guidelines:

```bash
# In your AI assistant chat
/speckit.constitution Create principles focused on code quality, testing 
standards, user experience consistency, and performance requirements.
```

### 4. Create a Specification

```bash
# In AI assistant chat
/speckit.specify Build a task management application with real-time 
collaboration, dark mode support, and offline capabilities.
```

### 5. Create Implementation Plan

```bash
# In AI assistant chat
/speckit.plan Use React with TypeScript, Node.js backend, PostgreSQL 
database, and Socket.io for real-time features.
```

### 6. Generate Tasks

```bash
# In AI assistant chat
/speckit.tasks
```

### 7. Execute Implementation

```bash
# In AI assistant chat
/speckit.implement
```

---

## CLI Commands Reference

### Core Commands

| Command | Description |
|---------|-------------|
| `specify init <name>` | Initialize new project |
| `specify init .` | Initialize in current directory |
| `specify init --here` | Initialize in current directory (flag style) |
| `specify check` | Check system requirements |
| `specify --version` | Show version |
| `specify --help` | Show help |

### Init Command Options

| Flag | Description |
|------|-------------|
| `--ai <assistant>` | Specify AI assistant |
| `--here` | Initialize in current directory |
| `--force` | Force merge into non-empty directory |
| `--no-git` | Skip git initialization |
| `--script <type>` | Script variant (sh/ps) |
| `--ai-skills` | Install as agent skills |
| `--branch-numbering <mode>` | Sequential or timestamp |
| `--github-token <token>` | GitHub token for API |
| `--debug` | Enable debug output |
| `--skip-tls` | Skip TLS verification |
| `--ignore-agent-tools` | Skip AI tool checks |
| `--ai-commands-dir <path>` | Custom commands directory |

### Supported AI Assistants

```bash
# Claude Code
specify init my-project --ai claude

# GitHub Copilot
specify init my-project --ai copilot

# Cursor
specify init my-project --ai cursor-agent

# Gemini CLI
specify init my-project --ai gemini

# Codex CLI
specify init my-project --ai codex --ai-skills

# Qwen Code
specify init my-project --ai qwen

# Windsurf
specify init my-project --ai windsurf

# Kiro CLI
specify init my-project --ai kiro-cli

# Junie
specify init my-project --ai junie

# OpenCode
specify init my-project --ai opencode

# Roo Code
specify init my-project --ai roo

# Full list
specify init my-project --ai generic --ai-commands-dir .myagent/commands/
```

### Complete Init Examples

```bash
# Basic initialization
specify init my-project

# With specific AI
specify init my-project --ai claude

# Current directory, force merge
specify init . --force --ai claude

# With PowerShell scripts
specify init my-project --ai copilot --script ps

# Skip git, enable debug
specify init my-project --ai gemini --no-git --debug

# Timestamp-based branch numbering
specify init my-project --ai claude --branch-numbering timestamp

# With GitHub token
specify init my-project --ai claude --github-token ghp_xxx

# Install as skills (for Codex, Antigravity)
specify init my-project --ai codex --ai-skills

# Custom agent
specify init my-project --ai generic --ai-commands-dir .custom/commands/
```

---

## Slash Commands (AI Assistant Integration)

### Core Workflow Commands

| Command | Purpose | File Created |
|---------|---------|--------------|
| `/speckit.constitution` | Define project principles | `.specify/memory/constitution.md` |
| `/speckit.specify` | Create feature specification | `specs/<feature>/spec.md` |
| `/speckit.plan` | Technical implementation plan | `specs/<feature>/plan.md` |
| `/speckit.tasks` | Generate task list | `specs/<feature>/tasks.md` |
| `/speckit.implement` | Execute implementation | Code files |

### Optional Enhancement Commands

| Command | Purpose | When to Use |
|---------|---------|-------------|
| `/speckit.clarify` | Ask clarifying questions | Before planning, if spec is unclear |
| `/speckit.analyze` | Cross-artifact analysis | After tasks, before implementation |
| `/speckit.checklist` | Quality checklist | Before implementation for validation |
| `/speckit.taskstoissues` | Convert to GitHub Issues | For project management |

### Command Variations by Agent

**Claude Code:**
```
/speckit.constitution
/speckit.specify
/speckit.plan
/speckit.tasks
/speckit.implement
```

**Codex CLI:**
```
$speckit-constitution
$speckit-specify
$speckit-plan
$speckit-tasks
$speckit-implement
```

**Kiro / Copilot:**
```
/ui-ux-pro-max <request>
```

---

## Configuration

### Project Structure

After initialization:

```
my-project/
├── .specify/
│   ├── memory/
│   │   └── constitution.md      # Project principles
│   └── config.yaml              # Spec Kit configuration
├── specs/
│   └── 001-feature-name/
│       ├── spec.md              # Feature specification
│       ├── plan.md              # Implementation plan
│       └── tasks.md             # Task list
├── .claude/                     # Claude Code skills (if --ai claude)
│   └── skills/
│       └── speckit-*.md
├── .cursor/                     # Cursor rules (if --ai cursor)
│   └── rules/
│       └── *.md
└── .github/
    └── workflows/               # Optional CI/CD
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `SPECIFY_FEATURE` | Override feature detection for non-Git repos |
| `GITHUB_TOKEN` | GitHub token for API requests |
| `GH_TOKEN` | Alternative GitHub token |

### Configuration File (.specify/config.yaml)

```yaml
project:
  name: my-project
  ai_assistant: claude
  
workflow:
  branch_numbering: sequential  # or timestamp
  auto_create_pr: false
  
git:
  commit_style: conventional
  auto_commit: false
  
specifications:
  template_dir: .specify/templates
  required_sections:
    - overview
    - requirements
    - acceptance_criteria
    
agents:
  claude:
    skills_dir: .claude/skills
  codex:
    skills_dir: .agents/skills
  copilot:
    prompts_dir: .github/prompts
```

### Constitution Template

The constitution defines project principles:

```markdown
# Project Constitution

## Code Quality
- All code must have unit tests
- Minimum 80% code coverage
- Follow language style guides

## Architecture
- Prefer composition over inheritance
- Use dependency injection
- Separate concerns clearly

## User Experience
- Mobile-first design
- Accessibility compliance (WCAG 2.1)
- Performance budget: < 3s load time

## Security
- No secrets in code
- Input validation on all boundaries
- Use parameterized queries
```

---

## Usage Examples

### Example 1: Greenfield Web Application

```bash
# Initialize project
specify init task-manager --ai claude
cd task-manager

# In Claude Code:
/speckit.constitution Create principles for a modern web app with 
focus on performance, accessibility, and clean architecture.

/speckit.specify Build a task management app with:
- User authentication
- Task CRUD operations  
- Due date reminders
- Team collaboration
- Dark mode support

/speckit.plan Use Next.js 14 with App Router, Prisma ORM, 
PostgreSQL, and Tailwind CSS. Deploy to Vercel.

/speckit.tasks

/speckit.implement
```

### Example 2: API Development

```bash
specify init payment-api --ai copilot
cd payment-api

# In Copilot Chat:
/speckit.specify Create a REST API for payment processing with 
Stripe integration, webhook handling, and transaction logging.

/speckit.plan Use Node.js, Express, TypeScript, and MongoDB.
Include comprehensive error handling and idempotency keys.

/speckit.tasks

/speckit.implement
```

### Example 3: Brownfield Enhancement

```bash
# In existing project
cd existing-project
specify init . --ai claude --force

# In Claude Code:
/speckit.specify Add real-time notifications to the existing 
application using WebSockets and Redis pub/sub.

/speckit.plan Integrate Socket.io with existing Express server.
Use Redis adapter for multi-server deployment.

/speckit.tasks

/speckit.implement
```

### Example 4: Multi-Agent Workflow

```bash
# Initialize for multiple agents
specify init shared-project --ai all

# Work with different agents on different tasks
# Claude Code: Architecture and core features
# Copilot: Testing and documentation
# Cursor: UI components
```

### Example 5: GitHub Issues Integration

```bash
# After creating tasks
/speckit.taskstoissues

# This creates GitHub Issues from tasks.md
# Each task becomes an issue with:
# - Title and description
# - Labels (feature, bug, etc.)
# - Assignees (if configured)
# - Milestone (if configured)
```

### Example 6: Quality Gates

```bash
# Before implementation
/speckit.checklist

# Generates quality checklist:
# - [ ] All requirements have acceptance criteria
# - [ ] Error cases are handled
# - [ ] Security considerations addressed
# - [ ] Performance requirements specified
```

---

## TUI / Interactive Features

### Workflow Visualization

Spec Kit creates a visual workflow representation:

```
┌─────────────────────────────────────────┐
│  Spec Kit Workflow                      │
├─────────────────────────────────────────┤
│  ✓ Constitution  →  Principles defined  │
│  ✓ Specify       →  Requirements set    │
│  ✓ Plan          →  Technical design    │
│  ○ Tasks         →  Pending...          │
│  ○ Implement     →  Waiting             │
│  ○ Review        →  Waiting             │
└─────────────────────────────────────────┘
```

### Task Dependency Graph

```bash
# Visualize task dependencies
specify visualize --tasks

# Output:
# Task 1: Setup project [ROOT]
#   └─ Task 2: Configure database
#       └─ Task 3: Create models
#           └─ Task 4: Implement API
#               └─ Task 5: Add tests
```

### Progress Tracking

```bash
# Check project status
specify status

# Output:
# Project: my-project
# Current Feature: 001-user-authentication
# Status: In Progress
# Tasks: 5/12 completed
# Last Updated: 2026-04-02
```

---

## Troubleshooting

### Installation Issues

**Problem:** `specify` command not found

**Solutions:**
```bash
# Check if installed
which specify

# If using uv, ensure path is set
export PATH="$HOME/.local/bin:$PATH"

# Reinstall with uv
uv tool uninstall specify-cli
uv tool install specify-cli --from git+https://github.com/github/spec-kit.git

# Or use Python module
python -m specify --help
```

### AI Assistant Not Detected

**Problem:** Spec Kit doesn't recognize installed AI assistant

**Solutions:**
```bash
# Skip detection
specify init my-project --ai claude --ignore-agent-tools

# Or use generic mode
specify init my-project --ai generic --ai-commands-dir .myagent/commands/
```

### Git Issues

**Problem:** Git initialization fails

**Solutions:**
```bash
# Skip git init
specify init my-project --no-git

# Or manually init git first
git init
specify init . --ai claude
```

### Permission Issues

**Problem:** Cannot write to directory

**Solutions:**
```bash
# Check permissions
ls -la

# Fix ownership
sudo chown -R $USER:$USER .

# Or use force
specify init . --force
```

### GitHub API Rate Limits

**Problem:** API rate limit exceeded

**Solutions:**
```bash
# Set GitHub token
export GITHUB_TOKEN=ghp_xxx

# Or pass directly
specify init my-project --github-token ghp_xxx
```

### Feature Detection Issues

**Problem:** Working on wrong feature

**Solutions:**
```bash
# Set feature explicitly
export SPECIFY_FEATURE=002-payment-integration

# Or in AI assistant context before running commands
```

### Slash Commands Not Working

**Problem:** AI assistant doesn't recognize Spec Kit commands

**Solutions:**

For **Claude Code:**
- Ensure `.claude/skills/` directory exists
- Check skill files are present
- Restart Claude Code

For **Codex CLI:**
- Verify `.agents/skills/` directory
- Use `$speckit-*` syntax, not `/speckit.*`
- Ensure `--ai-skills` was used during init

For **Cursor:**
- Check `.cursor/rules/` directory
- Restart Cursor

### Specification Quality Issues

**Problem:** Generated specs are too generic

**Solutions:**
- Provide more detailed descriptions
- Use `/speckit.clarify` before planning
- Review and refine constitution
- Run `/speckit.analyze` for consistency check

### Common Error Messages

| Error | Solution |
|-------|----------|
| "Not a Spec Kit project" | Run `specify init` first |
| "AI assistant not found" | Use `--ignore-agent-tools` flag |
| "Git repository required" | Run `git init` or use `--no-git` |
| "Invalid configuration" | Check `.specify/config.yaml` syntax |
| "Branch numbering conflict" | Use `--branch-numbering timestamp` |

### Getting Help

```bash
# Built-in help
specify --help
specify init --help

# Check version
specify --version

# Debug mode
specify init my-project --debug

# Community resources
# GitHub: https://github.com/github/spec-kit
# Issues: https://github.com/github/spec-kit/issues
```

---

## Best Practices

### 1. Start with Constitution
Always establish project principles first to ensure consistency.

### 2. Be Specific in Specifications
Provide detailed requirements, user stories, and acceptance criteria.

### 3. Use Clarify Before Plan
Run `/speckit.clarify` to fill gaps in specifications.

### 4. Validate with Analyze
Use `/speckit.analyze` to check consistency across artifacts.

### 5. Break Down Complex Features
Split large features into smaller, manageable specifications.

### 6. Review Before Implement
Always review the plan and tasks before starting implementation.

### 7. Use Version Control
Commit after each phase (constitution, spec, plan, tasks).

### 8. Iterate and Refine
Specs are living documents - update them as requirements evolve.

---

## Advanced Topics

### Custom Templates

Create custom specification templates:

```yaml
# .specify/templates/custom-spec.md
## Overview
{{overview}}

## User Stories
{{user_stories}}

## Technical Constraints
{{constraints}}

## Security Requirements
{{security}}
```

### CI/CD Integration

```yaml
# .github/workflows/spec-kit.yml
name: Spec Kit Validation

on: [pull_request]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Validate Specifications
        run: |
          pip install specify-cli
          specify validate --strict
```

### Team Workflow

```bash
# Feature branch workflow
git checkout -b feature/001-user-auth

# Team member 1: Create spec
/speckit.specify

# Team member 2: Review and add plan
/speckit.plan

# Team member 3: Generate tasks
/speckit.tasks

# Convert to issues for tracking
/speckit.taskstoissues

# Implement
/speckit.implement
```

---

## Resources

- **GitHub:** https://github.com/github/spec-kit
- **Documentation:** https://github.com/github/spec-kit/blob/main/README.md
- **Demo Projects:**
  - .NET CLI: https://github.com/mnriem/spec-kit-dotnet-cli-demo
  - Spring + React: https://github.com/mnriem/spec-kit-spring-react-demo

---

*Last Updated: April 2026*
