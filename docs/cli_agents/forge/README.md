# Forge

## Overview

**Forge** is an AI-enhanced terminal development environment that integrates AI capabilities with your development workflow. It provides a comprehensive coding agent for code understanding, implementation, debugging, and more.

**Website:** https://forgecode.dev  
**Repository:** https://github.com/antinomyhq/forge

---

## Key Features

| Feature | Description |
|---------|-------------|
| **Code Understanding** | Analyze and explain complex codebases |
| **Feature Implementation** | Scaffold and implement new features |
| **Debugging** | Error analysis and solution suggestions |
| **Code Reviews** | Automated code review and improvements |
| **Multi-Provider** | Support for multiple AI providers |
| **MCP Support** | Model Context Protocol integration |
| **Skills System** | Pre-built capabilities for common tasks |

---

## Installation

```bash
curl -fsSL https://forgecode.dev/cli | sh
```

Or download from [GitHub Releases](https://github.com/antinomyhq/forge/releases)

---

## Quick Start

### 1. Login to Provider

```bash
forge provider login
```

### 2. Start Forge

```bash
forge
```

---

## Usage Examples

### Code Understanding
```
> Can you explain how the authentication system works?
```

### Feature Implementation
```
> I need to add a dark mode toggle to our React application
```

### Debugging
```
> I'm getting this error: "TypeError: Cannot read property 'map' of undefined"
```

### Code Review
```
> Please review the code in src/components/UserProfile.js
```

---

## Configuration

### forge.yaml

```yaml
provider: openai
model: gpt-4
auto-apply: false
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `FORGE_API_KEY` | API key for provider |
| `FORGE_PROVIDER` | Default provider |
| `FORGE_MODEL` | Default model |

---

## Skills

Forge includes built-in skills for common tasks:

- **github-pr-comments** - Review and respond to PR comments
- **post-forge-feature** - Create feature posts
- **test-reasoning** - Reasoning about test cases

---

*Part of the HelixAgent CLI Agent Collection*
