# Bridle - User Guide

**Bridle** is a TUI/CLI configuration manager for AI coding harnesses. It provides a unified interface to manage profiles, install skills and agents, and switch configurations across Claude Code, OpenCode, Goose, Amp, Copilot CLI, and Crush.

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

- Node.js 18+ (for npm/bun/pnpm installation)
- OR Rust toolchain (for Cargo installation)
- macOS, Linux, or Windows

### Method 1: npx/bunx/pnpm dlx (No Install)

Run without installing:

```bash
# Using npx
npx bridle-ai

# Using bunx
bunx bridle-ai

# Using pnpm
pnpm dlx bridle-ai
```

### Method 2: npm/bun/pnpm (Global Install)

```bash
# npm
npm install -g bridle-ai

# bun
bun install -g bridle-ai

# pnpm
pnpm add -g bridle-ai
```

### Method 3: Homebrew

```bash
brew install neiii/bridle/bridle
```

### Method 4: Cargo

```bash
cargo install bridle
```

### Method 5: From Source

```bash
git clone https://github.com/neiii/bridle
cd bridle
cargo install --path .
```

### Verify Installation

```bash
bridle --version
```

---

## Quick Start

### Launch the TUI

```bash
# Start interactive TUI
bridle
```

### Check Status

```bash
# See what's configured across all harnesses
bridle status
```

### Create a Profile

```bash
# Create profile from current config
bridle profile create claude work --from-current

# Create empty profile
bridle profile create claude personal
```

### Switch Profiles

```bash
# Switch to a profile
bridle profile switch claude personal

# Verify switch
bridle status
```

### Install Skills

```bash
# Install from GitHub
bridle install owner/repo

# Example
bridle install vercel-labs/agent-skills
```

---

## CLI Commands

### Core Commands

| Command | Description |
|---------|-------------|
| `bridle` | Launch interactive TUI |
| `bridle --version` | Show version |
| `bridle --help` | Show help |
| `bridle init` | Initialize bridle config and default profiles |
| `bridle status` | Show active profiles across all harnesses |

### Profile Management

| Command | Description |
|---------|-------------|
| `bridle profile list <harness>` | List all profiles for a harness |
| `bridle profile show <harness> <name>` | Show profile details |
| `bridle profile create <harness> <name>` | Create empty profile |
| `bridle profile create <harness> <name> --from-current` | Create from current config |
| `bridle profile switch <harness> <name>` | Activate a profile |
| `bridle profile edit <harness> <name>` | Open profile in editor |
| `bridle profile diff <harness> <name> [other]` | Compare profiles |
| `bridle profile delete <harness> <name>` | Delete a profile |

### Installation Commands

| Command | Description |
|---------|-------------|
| `bridle install <source>` | Install skills/MCPs from GitHub |
| `bridle install <source> --force` | Overwrite existing installations |
| `bridle uninstall <harness> <profile>` | Interactively remove components |

### Configuration Commands

| Command | Description |
|---------|-------------|
| `bridle config get <key>` | Get a config value |
| `bridle config set <key> <value>` | Set a config value |

**Config keys:**
- `profile_marker` - Marker for active profile
- `editor` - Default editor
- `tui.view` - TUI view mode
- `default_harness` - Default harness to use

### Output Formats

All commands support `-o, --output <format>`:

```bash
# Text (default) - Human-readable
bridle status -o text

# JSON - Machine-readable
bridle status -o json

# Auto - Text for TTY, JSON for pipes
bridle status -o auto
```

---

## TUI/Interactive Commands

### Main Interface

The TUI provides a dashboard view of all configured harnesses and their active profiles.

### Navigation

| Key | Action |
|-----|--------|
| `↑/↓` | Navigate items |
| `Enter` | Select/confirm |
| `Esc` | Go back/cancel |
| `Tab` | Switch panels |
| `q` | Quit |

### Profile Picker

When creating a new session with multiple profiles defined:

| Key | Action |
|-----|--------|
| `←/→` | Navigate between profiles |
| `Enter` | Select profile |

---

## Configuration

### Configuration File

Bridle stores configuration in platform-specific locations:

- **macOS**: `~/Library/Application Support/bridle/config.toml`
- **Linux**: `~/.config/bridle/config.toml`
- **Windows**: `%APPDATA%\bridle\config.toml`

### Config Structure

```toml
# Bridle configuration

[general]
editor = "vim"
default_harness = "claude"
profile_marker = "🔷"

[tui]
view = "compact"  # or "detailed"

color_scheme = "auto"

[paths]
# Override default harness paths (optional)
claude_skills = "~/.claude/skills"
opencode_skills = "~/.config/opencode/skill"

# Profile definitions
[[profiles]]
name = "work"
description = "Work configuration"

[[profiles]]
name = "personal"
description = "Personal projects"
```

### Harness-Specific Paths

Bridle automatically handles path translations between harnesses:

| Component | Claude Code | OpenCode | Goose | Copilot CLI | Crush |
|-----------|-------------|----------|-------|-------------|-------|
| Skills | `~/.claude/skills/` | `~/.config/opencode/skill/` | `~/.config/goose/skills/` | `~/.copilot/skills/` | `~/.config/crush/skills/` |
| Agents | `~/.claude/plugins/*/agents/` | `~/.config/opencode/agent/` | — | `~/.copilot/agents/` | — |
| Commands | `~/.claude/plugins/*/commands/` | `~/.config/opencode/command/` | — | — | — |
| MCPs | `~/.claude/.mcp.json` | `opencode.jsonc` | `config.yaml` | `~/.copilot/mcp-config.json` | `crush.json` |

### Environment Variables

| Variable | Description |
|----------|-------------|
| `BRIDLE_CONFIG` | Path to config file |
| `BRIDLE_DATA` | Path to data directory |
| `BRIDLE_EDITOR` | Default editor |
| `EDITOR` | Fallback editor |

### Profile Structure

Profiles are stored in the data directory:

```
~/.local/share/bridle/profiles/
├── claude/
│   ├── work/
│   │   ├── settings.json
│   │   ├── CLAUDE.md
│   │   └── skills/
│   └── personal/
│       ├── settings.json
│       └── CLAUDE.md
├── opencode/
│   ├── work/
│   │   └── opencode.jsonc
│   └── personal/
│       └── opencode.jsonc
└── goose/
    ├── work/
    │   └── config.yaml
    └── personal/
        └── config.yaml
```

---

## Usage Examples

### Basic Workflow

```bash
# 1. Initialize bridle
bridle init

# 2. Check current status
bridle status

# 3. Create work profile from current config
bridle profile create claude work --from-current

# 4. Modify work profile settings
bridle profile edit claude work

# 5. Switch to work profile
bridle profile switch claude work
```

### Managing Multiple Harnesses

```bash
# View all harnesses
bridle status

# Create profiles for each harness
bridle profile create claude work --from-current
bridle profile create opencode work --from-current
bridle profile create goose work --from-current

# Switch all to work context
bridle profile switch claude work
bridle profile switch opencode work
bridle profile switch goose work
```

### Installing Skills

```bash
# Install skill from GitHub
bridle install vercel-labs/agent-skills

# Install specific skill from repo
bridle install vercel-labs/agent-skills --skill frontend-design

# Force reinstall
bridle install vercel-labs/agent-skills --force
```

**What happens during install:**
1. Bridle scans the repo for skills, agents, commands, and MCPs
2. You select which components to install
3. You choose target harnesses and profiles
4. Bridle translates paths and configs for each harness automatically

### Comparing Profiles

```bash
# Compare current profile with another
bridle profile diff claude work personal

# Output as JSON for scripting
bridle profile diff claude work personal -o json
```

### CI/CD Integration

```bash
# Non-interactive profile switch
bridle profile switch claude work -o json

# Check status in JSON format
bridle status -o json | jq '.harnesses.claude.active_profile'
```

### Backup and Restore

```bash
# List all profiles
bridle profile list claude
bridle profile list opencode

# Backup (copy profile directory)
cp -r ~/.local/share/bridle/profiles ~/bridle-backup

# Restore
cp -r ~/bridle-backup/profiles ~/.local/share/bridle/
```

### Scripting Examples

```bash
#!/bin/bash
# Switch to work profile during work hours

HOUR=$(date +%H)
if [ "$HOUR" -ge 9 ] && [ "$HOUR" -lt 17 ]; then
    bridle profile switch claude work
    bridle profile switch opencode work
else
    bridle profile switch claude personal
    bridle profile switch opencode personal
fi
```

---

## Troubleshooting

### Installation Issues

#### "command not found" after npm install

```bash
# Check npm global bin
npm bin -g

# Add to PATH
echo 'export PATH="$(npm bin -g):$PATH"' >> ~/.bashrc
source ~/.bashrc
```

#### Cargo install fails

```bash
# Update Rust toolchain
rustup update

# Install with specific features
cargo install bridle --features tui
```

#### Homebrew tap not found

```bash
# Tap manually
brew tap neiii/bridle

# Then install
brew install bridle
```

### Configuration Issues

#### Config file not found

```bash
# Initialize config
bridle init

# Check config location
bridle config get --help
```

#### Profile switch not working

```bash
# Verify profile exists
bridle profile list claude

# Check current status
bridle status

# Debug with verbose output
bridle profile switch claude work --verbose
```

#### Editor not opening

```bash
# Set editor explicitly
bridle config set editor vim
# or
export EDITOR=vim
```

### TUI Issues

#### Display issues

```bash
# Force basic output
bridle status -o text

# Check terminal capabilities
echo $TERM
```

#### Colors not rendering

```bash
# Set color scheme
bridle config set tui.color_scheme "basic"

# Or use NO_COLOR
export NO_COLOR=1
```

### Path Translation Issues

#### Skills not appearing in harness

```bash
# Verify skill was installed
ls ~/.local/share/bridle/installed/

# Check harness path
bridle config get paths.claude_skills

# Re-link skills
bridle install <repo> --force
```

#### MCP config not syncing

```bash
# Check MCP config locations
ls ~/.claude/.mcp.json
ls ~/.config/opencode/opencode.jsonc

# Manual sync
bridle profile edit claude work
# Edit MCP section
```

### Common Errors

#### "Profile already exists"

```bash
# Use force to overwrite
bridle profile create claude work --force

# Or delete first
bridle profile delete claude work
bridle profile create claude work
```

#### "Harness not detected"

```bash
# Check harness installation
which claude
which opencode
which goose

# Install harness first, then retry
```

#### "No profiles found"

```bash
# Create initial profile
bridle profile create claude default

# Or init with defaults
bridle init
```

### Getting Help

```bash
# Show help
bridle --help

# Command-specific help
bridle profile --help
bridle install --help

# Check version
bridle --version
```

### Debug Mode

```bash
# Enable debug logging
export BRIDLE_DEBUG=1
bridle status

# Verbose output
bridle -v profile switch claude work
```

---

## Best Practices

1. **Use Descriptive Profile Names**: `work`, `personal`, `client-a`, etc.
2. **Version Control**: Track profile configs in dotfiles repo
3. **Regular Backups**: Backup `~/.local/share/bridle/profiles/`
4. **Test Before Switching**: Use `profile diff` to review changes
5. **Cross-Harness Consistency**: Keep skill sets similar across harnesses
6. **Clean Up**: Delete unused profiles to avoid confusion
7. **Documentation**: Document custom paths in profile descriptions

---

## Core Concepts

### Harnesses

AI coding assistants: `claude`, `opencode`, `goose`, `amp`, `copilot`, `crush`

### Profiles

Saved configurations per harness. Each harness can have multiple profiles (e.g., `work`, `personal`, `minimal`).

### Auto-Translation

Bridle automatically handles:
- Path differences between harnesses
- Configuration schema differences
- File naming conventions
- MCP configuration formats

---

*Last Updated: April 2026*
