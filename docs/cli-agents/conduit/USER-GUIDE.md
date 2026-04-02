# Conduit - User Guide

**Conduit** is a multi-agent terminal user interface (TUI) that lets you run AI coding assistants side-by-side. Orchestrate Claude Code, Codex CLI, and Gemini CLI in a tabbed interface with full session management, git integration, and real-time token tracking.

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

- **macOS**, **Linux**, or **Windows (WSL2)**
- Git (required for workspace management)
- At least one AI agent: Claude Code, Codex CLI, or Gemini CLI
- Terminal with Unicode and color support

### Method 1: Official Installer (Recommended)

```bash
curl -fsSL https://getconduit.sh/install | sh
```

### Method 2: Homebrew

```bash
brew install conduit
```

### Method 3: Cargo

```bash
cargo install conduit-tui
```

### Method 4: Build from Source

```bash
# Clone repository
git clone https://github.com/hey-pal/conduit.git
cd conduit

# Build with Rust
cargo build --release

# Install
cargo install --path .
```

### Verify Installation

```bash
conduit --version
```

---

## Quick Start

### First Launch

```bash
# Navigate to your project
cd ~/projects/my-app

# Start Conduit
conduit
```

### Create Your First Session

```bash
# In the TUI:
# Ctrl+N    - Open project picker
# Enter     - Select current directory
# Type      - Send prompts to the agent
```

### Open Multiple Tabs

```bash
# In Conduit TUI:
# Ctrl+T    - New tab
# Alt+2     - Switch to tab 2
# Ctrl+W    - Close tab
```

### Basic Navigation

```bash
# Tab           - Switch between preview/diff tabs
# Ctrl+Q        - Quit
# ?             - Show help
```

---

## CLI Commands

### Core Commands

| Command | Description |
|---------|-------------|
| `conduit` | Start the TUI |
| `conduit --version` | Show version |
| `conduit --help` | Show help |

### Debug Commands

| Command | Description |
|---------|-------------|
| `conduit debug-keys` | Test keyboard input |
| `conduit migrate-theme` | Migrate theme from VS Code |

### Flags

| Flag | Description |
|------|-------------|
| `--config <path>` | Use custom config file |
| `--theme <name>` | Use specific theme |
| `--agent <name>` | Default agent to use |

---

## TUI/Interactive Commands

### Global Shortcuts

| Key | Action |
|-----|--------|
| `?` | Show help |
| `Ctrl+Q` | Quit |
| `Ctrl+N` | New project/session |
| `Ctrl+T` | New tab |
| `Ctrl+W` | Close tab |
| `Tab` | Switch between tabs/views |

### Tab Navigation

| Key | Action |
|-----|--------|
| `Alt+1` to `Alt+9` | Switch to tab 1-9 |
| `Tab` / `Shift+Tab` | Next/previous tab |
| `Ctrl+W` | Close current tab |

### Session Navigation

| Key | Action |
|-----|--------|
| `↑/↓` | Navigate sessions |
| `Enter` | Attach to session |
| `Ctrl+Q` | Detach from session |

### Sidebar

| Key | Action |
|-----|--------|
| `S` or `/` | Search |
| `↑/↓` | Navigate items |
| `Enter` | Select item |

### Chat View

| Key | Action |
|-----|--------|
| `↑/↓` | Scroll messages |
| `Page Up/Down` | Fast scroll |
| `Home/End` | Jump to start/end |

### Scrolling

| Key | Action |
|-----|--------|
| `Shift+↑/↓` | Scroll in diff view |
| `Page Up/Down` | Page scroll |
| `Ctrl+Home/End` | Jump to start/end |

---

## Configuration

### Configuration File

Conduit stores configuration in:

- **macOS**: `~/.config/conduit/config.toml`
- **Linux**: `~/.config/conduit/config.toml`
- **Windows**: `%APPDATA%\conduit\config.toml`

### Basic Configuration

```toml
# Conduit configuration

[general]
# Default agent: claude, codex, or gemini
default_agent = "claude"

# Maximum tabs
max_tabs = 10

# Auto-save sessions
auto_save = true

[ui]
# Theme: auto, dark, light, or custom
theme = "auto"

# Show token usage in status bar
show_tokens = true

# Show git status
show_git = true

# Sidebar position: left or right
sidebar_position = "left"

[agents]
# Claude Code path
claude_path = "claude"

# Codex path
codex_path = "codex"

# Gemini path
gemini_path = "gemini"

# Agent-specific settings
[agents.claude]
args = ["--dangerously-skip-permissions"]

[agents.codex]
args = []

[keybindings]
# Custom keybindings
quit = "Ctrl+Q"
new_tab = "Ctrl+T"
close_tab = "Ctrl+W"
new_session = "Ctrl+N"
help = "?"

[git]
# Git integration settings
enable_worktrees = true
auto_commit = false
show_branch_status = true
pr_tracking = true

[tracking]
# Token tracking
token_usage = true
cost_estimation = true
daily_budget = 50.00  # USD
```

### Themes

Built-in themes:
- `auto` - Follow system theme
- `dark` - Dark theme
- `light` - Light theme

Custom theme:

```toml
[theme]
name = "my-theme"
background = "#1e1e1e"
foreground = "#d4d4d4"
accent = "#007acc"
success = "#4ec9b0"
warning = "#ce9178"
error = "#f44747"
info = "#569cd6"
border = "#3c3c3c"
highlight = "#264f78"
selection = "#094771"
```

### VS Code Theme Migration

```bash
# Migrate your VS Code theme
conduit migrate-theme --from vscode

# Specify theme file
conduit migrate-theme ~/path/to/theme.json
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `CONDUIT_CONFIG` | Path to config file |
| `CONDUIT_DATA` | Path to data directory |
| `CONDUIT_THEME` | Default theme |
| `CONDUIT_AGENT` | Default agent |

---

## Usage Examples

### Basic Workflow

```bash
# 1. Navigate to project
cd ~/projects/my-app

# 2. Start Conduit
conduit

# 3. Select project (Ctrl+N, then Enter for current dir)

# 4. Work with agent
# Type prompts in the chat input

# 5. Open new tab for parallel work
# Ctrl+T

# 6. Switch between tabs
# Alt+1, Alt+2, etc.

# 7. View git diff
# Tab to switch to diff view

# 8. Quit
# Ctrl+Q
```

### Multi-Agent Workflow

```bash
# Tab 1: Claude Code for backend
# Start Conduit
# Default agent: Claude

# Tab 2: Codex for frontend
# Ctrl+T
# Switch agent to Codex in sidebar

# Tab 3: Gemini for documentation
# Ctrl+T
# Switch agent to Gemini

# Switch between tabs with Alt+1, Alt+2, Alt+3
```

### Git Worktree Workflow

```bash
# Conduit automatically uses git worktrees

# 1. Start Conduit in repo
conduit

# 2. Each session gets its own worktree
# Check git status in diff tab (Tab)

# 3. Work on feature in tab 1
# Work on bug fix in tab 2

# 4. View diffs for each session
# Switch tabs to compare

# 5. Push when ready
# Use built-in git commands
```

### Build vs Plan Mode

```bash
# Build Mode (default)
# Agent has full access to make changes

# Plan Mode (read-only analysis)
# Switch to plan mode in sidebar
# Agent analyzes without making changes
# Review plan before approving
```

### Session Management

```bash
# Conduit automatically saves sessions

# Resume work:
conduit
# Previous sessions restored

# View session history in sidebar
# Search with S or /
```

### Token Tracking

```bash
# View real-time token usage in status bar

# Daily budget warning
# Configured in config.toml
# Warning when approaching limit
```

---

## Troubleshooting

### Installation Issues

#### "command not found" after install

```bash
# Check installation
which conduit

# Add to PATH if needed
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

#### Install script fails

```bash
# Check prerequisites
which curl
which bash

# Try manual install
cargo install conduit-tui
```

### TUI Issues

#### Display garbled or colors wrong

```bash
# Set TERM
export TERM=xterm-256color

# Check terminal capabilities
echo $TERM

# Try different terminal emulator
```

#### Keys not working

```bash
# Test key input
conduit debug-keys

# Check terminal compatibility
# Supported: iTerm2, Alacritty, Windows Terminal, GNOME Terminal

# Disable custom keybindings in config
```

#### Blank screen

```bash
# Check terminal size
stty size

# Resize terminal (minimum 80x24)

# Try different theme
conduit --theme dark
```

### Agent Issues

#### "Agent not found"

```bash
# Check agent installed
which claude
which codex
which gemini

# Configure paths in config.toml
[agents]
claude_path = "/usr/local/bin/claude"
codex_path = "/usr/local/bin/codex"
```

#### Agent fails to start

```bash
# Check agent works standalone
claude --version

# Check auth
claude login

# View Conduit logs
# Logs in ~/.config/conduit/logs/
```

### Session Issues

#### Sessions not saving

```bash
# Check data directory
ls ~/.config/conduit/sessions/

# Check permissions
chmod 755 ~/.config/conduit

# Enable auto-save in config
[general]
auto_save = true
```

#### Cannot resume session

```bash
# Check session exists
ls ~/.config/conduit/sessions/

# Try manual resume
conduit --session <session-id>
```

### Git Issues

#### Worktree errors

```bash
# Check git version
git --version  # Need 2.15+

# List worktrees
git worktree list

# Clean up stale worktrees
git worktree prune
```

#### PR tracking not working

```bash
# Check git remote
git remote -v

# Configure PR tracking
[git]
pr_tracking = true

# Requires GitHub CLI for full features
which gh
```

### Configuration Issues

#### Config not loading

```bash
# Check config location
ls ~/.config/conduit/config.toml

# Validate TOML syntax
# Use online TOML validator

# Reset to defaults
rm ~/.config/conduit/config.toml
conduit  # Creates default config
```

#### Theme not applying

```bash
# Check theme exists
ls ~/.config/conduit/themes/

# Use built-in theme
conduit --theme dark

# Check theme file format
```

### Performance Issues

#### Slow rendering

```bash
# Reduce max tabs
[general]
max_tabs = 5

# Disable token tracking
[tracking]
token_usage = false

# Use simpler theme
theme = "dark"
```

#### High CPU usage

```bash
# Check active sessions
# Close unused tabs

# Disable animations
[ui]
animations = false
```

### Common Errors

#### "Permission denied"

```bash
# Fix config permissions
chmod 755 ~/.config/conduit
chmod 644 ~/.config/conduit/config.toml
```

#### "No such file or directory"

```bash
# Create config directory
mkdir -p ~/.config/conduit

# Reinstall Conduit
```

### Debug Mode

```bash
# Enable debug logging
export CONDUIT_DEBUG=1
conduit

# Check logs
tail ~/.config/conduit/logs/conduit.log
```

### Getting Help

```bash
# In-TUI help
# Press ? for help

# CLI help
conduit --help

# Documentation
# https://getconduit.sh/docs/

# Discord community
# Join via website
```

---

## Best Practices

1. **Use Tabs Wisely**: Up to 10 concurrent sessions
2. **Name Sessions Clearly**: Descriptive names in sidebar
3. **Save Frequently**: Auto-save enabled by default
4. **Monitor Tokens**: Watch usage in status bar
5. **Git Hygiene**: Use worktrees for isolation
6. **Plan Mode First**: Analyze before building
7. **Close Unused Tabs**: Free up resources
8. **Customize Keybindings**: Match your workflow
9. **Theme Selection**: Comfortable for long sessions
10. **Regular Updates**: Keep Conduit updated

---

## Comparison with Alternatives

| Feature | Conduit | Claude Squad | Agent Deck |
|---------|---------|--------------|------------|
| Multi-agent tabs | Yes | Yes | Yes |
| Real-time streaming | Yes | Yes | Yes |
| Token tracking | Yes | No | Yes |
| Git worktrees | Yes | Yes | Yes |
| Web interface | Yes | No | Yes |
| VS Code theme import | Yes | No | No |
| PR tracking | Yes | Basic | No |
| Build/Plan modes | Yes | No | No |

---

*Last Updated: April 2026*
