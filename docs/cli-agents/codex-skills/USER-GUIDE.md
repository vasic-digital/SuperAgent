# Codex Skills - User Guide

**Codex Skills** is a skill management system for OpenAI's Codex CLI. It provides structured agentic workflows, composable skills, and expert guidance to enhance Codex's capabilities for software development tasks.

---

## Table of Contents

1. [Installation](#installation)
2. [Quick Start](#quick-start)
3. [CLI Commands](#cli-commands)
4. [TUI/Interactive Commands](#tuiinteractive-commands)
5. [Configuration](#configuration)
6. [Usage Examples](#usage-examples)
7. [Troubleshooting](#troubleshooting)

---

## Installation

### Prerequisites

- OpenAI Codex CLI installed and configured
- Node.js 18+
- Git (for skill repositories)

### Method 1: Skills CLI (Recommended)

```bash
# Install skill package
npx skills add owner/repo

# Install specific skill
npx skills add owner/repo --skill skill-name

# Install all skills from repo
npx skills add owner/repo --all
```

### Method 2: Direct Installation

```bash
# Using npx
npx superpowers-codex@latest install

# Install to user scope
npx superpowers-codex@latest install --user

# Install to project scope
npx superpowers-codex@latest install --project
```

### Method 3: Global Install

```bash
# Install globally
npm install -g superpowers-codex

# Then use without npx
superpowers-codex install
superpowers-codex status
```

### Method 4: Manual Installation

```bash
# Clone repository
git clone https://github.com/owner/repo.git

# Copy skills to Codex directory
mkdir -p ~/.codex/skills
cp -r repo/skills/* ~/.codex/skills/

# Or for project-specific
mkdir -p .agents/skills
cp -r repo/skills/* .agents/skills/
```

### Verify Installation

```bash
# Check installed skills
npx skills list

# Or in Codex session
/skills list
```

---

## Quick Start

### Install Your First Skill

```bash
# Install from Vercel's skill repo
npx skills add vercel-labs/agent-skills

# Install specific skill
npx skills add vercel-labs/agent-skills --skill react-testing
```

### Use Skills in Codex

```bash
# Start Codex
codex

# Skills auto-activate based on context
> Help me write tests for this React component

# Or invoke explicitly
> $react-testing Write tests for the Button component
```

### Check Skill Status

```bash
# Using superpowers
npx superpowers-codex status

# Shows installed skills and their scope (user/project)
```

### Update Skills

```bash
# Check for updates
npx skills check

# Update all skills
npx skills update

# Update specific skill
npx skills update skill-name
```

---

## CLI Commands

### Skills CLI Commands

| Command | Description |
|---------|-------------|
| `npx skills add <repo>` | Add skills from repository |
| `npx skills add <repo> --skill <name>` | Add specific skill |
| `npx skills add <repo> --all` | Add all skills |
| `npx skills add <repo> -a <agent>` | Add to specific agent |
| `npx skills list` | List installed skills |
| `npx skills remove` | Remove skills (interactive) |
| `npx skills check` | Check for updates |
| `npx skills update` | Update all skills |
| `npx skills find <query>` | Search for skills |

### Superpowers CLI Commands

| Command | Description |
|---------|-------------|
| `npx superpowers-codex install` | Install superpowers |
| `npx superpowers-codex install --user` | Install to user scope |
| `npx superpowers-codex install --project` | Install to project scope |
| `npx superpowers-codex uninstall --user` | Uninstall from user scope |
| `npx superpowers-codex status` | Show installation status |

### In-Codex Commands

| Command | Description |
|---------|-------------|
| `/skills` | List available skills |
| `/skills list` | Show all skills |
| `$skillname` | Invoke specific skill |
| `$skill-installer` | Built-in skill installer |
| `$skill-creator` | Create new skill |

---

## TUI/Interactive Commands

### Skills Browser

```bash
# In Codex session
/skills

# Navigate skills
# ↑/↓ - Navigate
# Enter - View details
# q - Quit
```

### Skill Invocation

```bash
# Explicit skill invocation with $ prefix
> $brainstorming Help me design a new feature

> $writing-plans Create an implementation plan

> $subagent-driven-development Execute the plan
```

---

## Configuration

### Skill Directory Structure

```
~/.codex/
├── skills/
│   └── skill-name/
│       ├── SKILL.md          # Main skill file
│       └── reference.md      # Optional reference
├── config.toml               # Codex config
└── agents/                   # Agent definitions

# Project-specific
./
├── .agents/
│   └── skills/
│       └── skill-name/
│           └── SKILL.md
└── codex.md                  # Project instructions
```

### Skill Format (SKILL.md)

```markdown
---
name: skill-name
description: "When this skill should activate"
tags: ["coding", "frontend", "testing"]
triggers:
  - "write tests"
  - "test this"
---

# Skill Title

Detailed instructions for Codex when this skill is active.

## Guidelines

1. Follow these specific guidelines
2. Use these patterns
3. Avoid these anti-patterns

## Examples

### Good Example
```typescript
// Good code example
```

### Bad Example
```typescript
// Bad code example
```

## Tools

Available tools and when to use them...

## References

- [Link to docs](url)
- Internal reference
```

### Codex Configuration

`~/.codex/config.toml`:

```toml
[general]
default_model = "gpt-5-codex"
auto_apply = false

[skills]
# Skill activation settings
auto_activate = true
strict_matching = false

# Custom skill directories
skill_paths = [
    "~/.codex/skills",
    "~/custom-skills"
]

[agents]
# Agent configuration
max_subagents = 3
```

### Project Configuration

`.codex/config.toml` in project root:

```toml
[project]
name = "my-project"
type = "nodejs"  # nodejs, python, go, rust, etc.

[skills]
# Project-specific skills
required = ["react-testing", "nodejs-best-practices"]

[context]
# Files to always include
include = ["README.md", "ARCHITECTURE.md"]

# Patterns to exclude
exclude = ["*.test.ts", "node_modules/**"]
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `CODEX_SKILLS_DIR` | Custom skills directory |
| `CODEX_CONFIG` | Path to config file |
| `CODEX_DEBUG_SKILLS` | Enable skill debugging |
| `SKILLS_CLI_CONFIG` | Skills CLI config path |

---

## Usage Examples

### Installing Common Skills

```bash
# Vercel's agent skills
npx skills add vercel-labs/agent-skills

# React and Next.js skills
npx skills add vercel-labs/agent-skills --skill react-best-practices
npx skills add vercel-labs/agent-skills --skill web-design-guidelines

# DataHub skills
npx skills add datahub-project/datahub-skills

# Databricks skills
npx skills add databricks/databricks-agent-skills
```

### Using Superpowers Workflow

```bash
# 1. Install superpowers
npx superpowers-codex install --user

# 2. Restart Codex

# 3. Use skill invocation
> $brainstorming Help me design a login page

# 4. Continue workflow
> $writing-plans Turn this design into an implementation plan

# 5. Execute
> $subagent-driven-development Execute the approved plan

# 6. Finish
> $finishing-a-development-branch Clean up and prepare for PR
```

### Creating Custom Skills

```bash
# Method 1: Using $skill-creator
# In Codex:
> $skill-creator

# Follow prompts to create skill

# Method 2: Manual creation
mkdir -p ~/.codex/skills/my-skill
cat > ~/.codex/skills/my-skill/SKILL.md << 'EOF'
---
name: my-skill
description: "When working with our internal framework"
---

# Internal Framework Guidelines

Always follow these patterns when working with our codebase:

1. Use the ApiClient from @lib/api
2. Handle errors with tryCatch wrapper
3. Log to DataDog, not console
4. Use TypeScript strict types

## Common Patterns

### API Call
```typescript
const result = await tryCatch(
  () => apiClient.get('/endpoint'),
  (error) => logger.error('API failed', error)
);
```
EOF
```

### Project-Specific Skills

```bash
# Create project skill directory
mkdir -p .agents/skills/project-conventions

# Create skill
cat > .agents/skills/project-conventions/SKILL.md << 'EOF'
---
name: project-conventions
description: "When working on this specific project"
---

# Project Conventions

This project uses:
- Next.js 14 with App Router
- Prisma ORM
- Tailwind CSS
- React Query for data fetching

## File Structure

- `app/` - Next.js app router pages
- `components/` - React components
- `lib/` - Utility functions
- `db/` - Database schema and queries

## Code Style

- Use named exports
- Prefix components with feature name
- Add JSDoc for public APIs
EOF

# Install to project
npx skills add . --path .agents/skills/project-conventions
```

### Skill Development Workflow

```bash
# 1. Create skill locally
mkdir -p ~/dev/my-skill
cat > ~/dev/my-skill/SKILL.md << 'EOF'
---
name: my-awesome-skill
description: "When to use this skill"
---

# My Awesome Skill

Instructions here...
EOF

# 2. Test locally
cp -r ~/dev/my-skill ~/.codex/skills/
codex
> $my-awesome-skill Test prompt

# 3. Publish to GitHub
cd ~/dev/my-skill
git init
git add .
git commit -m "Initial skill"
git remote add origin https://github.com/username/my-skill.git
git push -u origin main

# 4. Others can install
npx skills add username/my-skill
```

### CI/CD Integration

```bash
# Install skills in CI
npx skills add vercel-labs/agent-skills --all --yes

# Or use config file
# .github/workflows/codex.yml
# - name: Install Codex Skills
#   run: npx skills add owner/repo --all --yes
```

---

## Troubleshooting

### Installation Issues

#### "command not found: skills"

```bash
# Use npx
npx skills --help

# Or install globally
npm install -g @vercel/skills
```

#### "Repository not found"

```bash
# Verify repo exists
curl -s https://api.github.com/repos/owner/repo | grep full_name

# Check spelling
npx skills add correct-owner/correct-repo
```

#### "Skill already exists"

```bash
# Force install
npx skills add owner/repo --force

# Or remove first
npx skills remove
# Select skill to remove
```

### Loading Issues

#### Skills not appearing in Codex

```bash
# Check installation location
ls ~/.codex/skills/
ls .agents/skills/

# Verify SKILL.md exists
cat ~/.codex/skills/skill-name/SKILL.md

# Check frontmatter format
# Should have --- at start and end of frontmatter

# Restart Codex
```

#### Skills not auto-activating

```bash
# Check description in frontmatter
# Description should match your prompt keywords

# Invoke explicitly with $
> $skill-name your prompt

# Check config
# ~/.codex/config.toml should have auto_activate = true
```

#### "Invalid SKILL.md format"

```bash
# Validate frontmatter
cat ~/.codex/skills/skill-name/SKILL.md | head -20

# Should be:
# ---
# name: skill-name
# description: "Description"
# ---

# Check for special characters
# Use quotes around strings with special chars
```

### Invocation Issues

#### "$skill not recognized"

```bash
# Check skill exists
ls ~/.codex/skills/ | grep skill-name

# Verify name in frontmatter matches
# name: should match what you type after $

# Try /skills list to see available skills
```

#### Skill activates but doesn't work

```bash
# Check skill content
# Instructions may be unclear

# Review reference.md if exists
# May have additional context

# Test with explicit context
> Following the skill-name skill, help me with...
```

### Update Issues

#### "No updates found" but new version exists

```bash
# Clear cache
rm -rf ~/.cache/skills/

# Force reinstall
npx skills add owner/repo --force

# Check version manually
curl -s https://api.github.com/repos/owner/repo/releases/latest
```

### Superpowers Issues

#### "superpowers-codex not found"

```bash
# Install globally
npm install -g superpowers-codex

# Or use npx
npx superpowers-codex@latest install
```

#### Config not applied

```bash
# Check scope
# --user: ~/.codex/config.toml
# --project: ./.codex/config.toml

# Verify config location
npx superpowers-codex status

# Reinstall with correct scope
npx superpowers-codex uninstall --user
npx superpowers-codex install --user
```

### Common Errors

#### "ENOENT: no such file or directory"

```bash
# Create directories
mkdir -p ~/.codex/skills
mkdir -p .agents/skills

# Then retry installation
```

#### "EACCES: permission denied"

```bash
# Fix permissions
chmod 755 ~/.codex
chmod -R 644 ~/.codex/skills/*
```

### Debug Mode

```bash
# Enable skill debugging
export CODEX_DEBUG_SKILLS=1
codex

# Check logs
# Logs show which skills loaded and why
```

### Getting Help

```bash
# Skills CLI help
npx skills --help

# Superpowers help
npx superpowers-codex --help

# In Codex
/skills help

# GitHub
# https://github.com/vercel/skills
```

---

## Best Practices

1. **Clear Descriptions**: Write specific trigger descriptions
2. **Progressive Disclosure**: Keep core instructions brief
3. **Examples**: Include good and bad examples
4. **Test Skills**: Verify skills work before sharing
5. **Version Control**: Track skills in git
6. **Scope Appropriately**: Use --user for personal, --project for team
7. **Keep Updated**: Regularly update skills
8. **Document References**: Link to external docs
9. **Namespace Carefully**: Use unique skill names
10. **Share Knowledge**: Publish useful skills

---

## Popular Skill Repositories

| Repository | Description |
|------------|-------------|
| `vercel-labs/agent-skills` | Core Vercel skills (React, Next.js, testing) |
| `datahub-project/datahub-skills` | DataHub connector development |
| `databricks/databricks-agent-skills` | Databricks development |
| `microsoft/power-platform-skills` | Power Platform development |
| `chriscox/agent-skills` | Project planning and docs sync |
| `yigitkonur/superpowers-codex` | Complete development workflow |

---

*Last Updated: April 2026*
