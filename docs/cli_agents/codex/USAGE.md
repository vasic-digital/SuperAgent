# OpenAI Codex - Usage Guide

## Common Workflows

### Code Exploration

```bash
# Understand a codebase
codex "explain this project to me"

# Find specific code
codex "where is authentication handled?"

# Review file
codex "explain utils/date.ts"
```

### Code Generation

```bash
# Create new feature
codex -a full-auto "create a login page with React"

# Generate tests
codex "write unit tests for auth.ts"

# Create migrations
codex "generate SQL migrations for users table"
```

### Refactoring

```bash
# Refactor with auto-apply
codex -a auto-edit "convert to async/await"

# Bulk rename
codex "rename all .jpeg to .jpg with git mv"

# TypeScript conversion
codex "add types to untyped functions"
```

### Debugging

```bash
# Fix errors
codex "fix the failing tests"

# Security audit
codex "find vulnerabilities in this codebase"

# Performance check
codex "optimize slow database queries"
```

## Best Practices

### 1. Project Documentation

Create `AGENTS.md` at project root:

```markdown
# Project Guidelines

## Stack
- React + TypeScript
- Node.js backend
- PostgreSQL

## Commands
- npm run dev - Start dev server
- npm test - Run tests
- npm run build - Production build

## Conventions
- Use functional components
- Prefer async/await
- Add JSDoc comments
```

### 2. Safety First

```bash
# Start with suggest mode (default)
codex "make changes"  # Review each change

# Use auto-edit for trusted operations
codex -a auto-edit "apply lint fixes"

# Full auto only when confident
codex -a full-auto "scaffold new project"
```

### 3. Git Integration

```bash
# Create branch first
git checkout -b feature/new-thing

# Use codex
codex -a full-auto "implement feature"

# Review changes
git diff

# Commit
git commit -m "Add new feature"
```

## Recipes

| Task | Command |
|------|---------|
| Refactor component | `codex "refactor Dashboard to use Hooks"` |
| Generate migrations | `codex "create SQL migrations for users table"` |
| Write tests | `codex "write tests for utils/date.ts"` |
| Bulk rename | `codex "rename *.jpeg to *.jpg with git mv"` |
| Explain regex | `codex "explain: ^(?=.*[A-Z]).{8,}$"` |
| Security audit | `codex "find vulnerabilities"` |
| Propose PRs | `codex "propose 3 high-impact PRs"` |

## CI/CD Integration

```yaml
# GitHub Actions example
- name: Update changelog
  run: |
    npm install -g @openai/codex
    export OPENAI_API_KEY="${{ secrets.OPENAI_KEY }}"
    codex -a auto-edit --quiet "update CHANGELOG"
```

---

*See [API Reference](./API.md) for command details*
