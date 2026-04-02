# Forge User Guide

## Table of Contents
1. [Installation](#installation)
2. [Quick Start](#quick-start)
3. [CLI Commands](#cli-commands)
4. [TUI/Interactive Commands](#tui-interactive-commands)
5. [Configuration](#configuration)
6. [Usage Examples](#usage-examples)
7. [Troubleshooting](#troubleshooting)

## Installation

### Method 1: Quick Install
```bash
curl -fsSL https://forgecode.dev/cli | sh
```

### Method 2: Homebrew
```bash
brew tap antinomyhq/forge
brew install forge
```

### Method 3: Nix
```bash
nix run github:antinomyhq/forge
```

### Method 4: Cargo
```bash
cargo install forge
```

## Quick Start

```bash
# Setup API key
export FORGE_API_KEY=your-key
# Or use: forge config set api_key your-key

# Start Forge
forge

# Execute a task
forge "Create a REST API in Rust"

# Use specific agent
forge --agent sage "Review this code"
```

## CLI Commands

### Global Options
| Option | Short | Description | Example |
|--------|-------|-------------|---------|
| --help | -h | Show help | `forge --help` |
| --version | -v | Show version | `forge --version` |
| --config | -c | Config file | `--config ~/.forge.yml` |
| --agent | -a | Select agent | `--agent forge` |
| --model | -m | Model to use | `--model gpt-4` |
| --dry-run | | Simulate only | `--dry-run` |

### Command: (default)
**Description:** Execute a task with Forge

**Usage:**
```bash
forge "Your task description"
forge --file task.md
```

### Command: config
**Description:** Manage configuration

**Subcommands:**
```bash
forge config get <key>          # Get config value
forge config set <key> <value>  # Set config value
forge config list               # List all configs
forge config edit               # Edit config file
```

### Command: agent
**Description:** Agent management

**Subcommands:**
```bash
forge agent list          # List available agents
forge agent info <name>   # Show agent info
forge agent use <name>    # Set default agent
```

### Command: workflow
**Description:** Workflow management

**Subcommands:**
```bash
forge workflow list       # List workflows
forge workflow run <name> # Run a workflow
forge workflow show <name># Show workflow details
```

### Command: context
**Description:** Context management

**Subcommands:**
```bash
forge context list        # List contexts
forge context use <name>  # Switch context
forge context clear       # Clear current context
```

## TUI/Interactive Commands

When in interactive mode:

| Command | Shortcut | Description |
|---------|----------|-------------|
| /help | | Show help |
| /exit | Ctrl+D | Exit Forge |
| /clear | | Clear context |
| /agent | | Switch agent |
| /model | | Change model |
| /compact | | Compact context |
| /undo | | Undo last action |
| /retry | | Retry failed action |

## Configuration

### Configuration File Format (YAML)

```yaml
# ~/.forge/forge.yml
api_key: your-api-key
model: gpt-4
agent: forge

agents:
  forge:
    model: gpt-4
    temperature: 0.7
  sage:
    model: claude-3-opus
    temperature: 0.5
  muse:
    model: gemini-pro
    temperature: 0.8

features:
  tool_use: true
  git: true
  web_search: false
```

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| FORGE_API_KEY | API key | Yes |
| FORGE_MODEL | Default model | No |
| FORGE_AGENT | Default agent | No |
| OPENAI_API_KEY | OpenAI key | Alternative |
| ANTHROPIC_API_KEY | Anthropic key | Alternative |

### Configuration Locations
1. `~/.forge/forge.yml`
2. Project `.forge.yml`
3. Environment variables
4. Command-line flags

## Usage Examples

### Example 1: Code Implementation
```bash
forge "Create a Python function to parse CSV files with validation"
```

### Example 2: Code Review
```bash
forge --agent sage "Review src/main.rs for safety issues"
```

### Example 3: Research Task
```bash
forge --agent muse "Research best practices for Rust error handling"
```

### Example 4: Workflow
```bash
forge workflow run refactor
# Executes predefined refactoring workflow
```

### Example 5: With File Input
```bash
cat task.md | forge
# or
forge --file task.md
```

## Troubleshooting

### Issue: API Key Not Set
**Solution:**
```bash
forge config set api_key your-key
# Or:
export FORGE_API_KEY=your-key
```

### Issue: Agent Not Found
**Solution:**
```bash
forge agent list
forge agent use forge
```

### Issue: Context Too Long
**Solution:**
```bash
forge context compact
# Or in interactive mode:
/compact
```

### Issue: Tool Execution Failed
**Solution:**
```bash
# Check tool permissions
forge config get features.tool_use
# Enable if needed:
forge config set features.tool_use true
```

---

**Last Updated:** 2026-04-02
