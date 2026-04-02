# Claude Squad - User Guide

**Claude Squad** is a terminal application that manages multiple Claude Code, Codex, Gemini, and Aider sessions in separate workspaces. It enables you to work on multiple tasks simultaneously with isolated git worktrees, background processing, and a unified TUI interface.

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

- macOS or Linux (Windows via WSL)
- tmux 3.0+
- gh (GitHub CLI)
- Claude Code, Codex, or other supported agents installed

### Method 1: Homebrew (Recommended)

```bash
brew install claude-squad

# Create shortcut (optional)
ln -s "$(brew --prefix)/bin/claude-squad" "$(brew --prefix)/bin/cs"
```

### Method 2: Install Script

```bash
# Install with default name (claude-squad)
curl -fsSL https://raw.githubusercontent.com/smtg-ai/claude-squad/main/install.sh | bash

# Install with custom name
curl -fsSL https://raw.githubusercontent.com/smtg-ai/claude-squad/main/install.sh | bash -s -- --name cs
```

Binary is installed to `~/.local/bin/`.

### Method 3: From Source

```bash
git clone https://github.com/smtg-ai/claude-squad.git
cd claude-squad
go build -o cs ./cmd/claude-squad
mv cs ~/.local/bin/
```

### Verify Installation

```bash
# Check version
claude-squad version
# or
cs version

# Check debug info
claude-squad debug
```

---

## Quick Start

### Launch Claude Squad

```bash
# Start with default program (claude)
cs

# Or use full name
claude-squad
```

### Create Your First Session

```bash
# In the TUI, press 'n' to create new session
# Or 'N' to create with a prompt
```

### Work with Multiple Agents

```bash
# Launch with Codex
cs -p "codex"

# Launch with Aider
cs -p "aider --model ollama_chat/gemma3:1b"

# Launch with Gemini
cs -p "gemini"
```

### Enable Auto-Yes Mode (Experimental)

```bash
# Automatically accept prompts
cs -y
```

---

## CLI Commands

### Core Commands

| Command | Description |
|---------|-------------|
| `cs` or `claude-squad` | Launch TUI |
| `cs version` | Show version |
| `cs debug` | Show debug information |
| `cs reset` | Reset all stored instances |
| `cs completion <shell>` | Generate shell completions |

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--program` | `-p` | Program to run in new instances |
| `--autoyes` | `-y` | Auto-accept all prompts (experimental) |
| `--help` | `-h` | Show help |

### Program Examples

```bash
# Default (Claude Code)
cs

# Codex CLI
cs -p "codex"

# Aider with specific model
cs -p "aider --model ollama_chat/gemma3:1b"

# Gemini CLI
cs -p "gemini"
```

### Shell Completions

```bash
# Bash
cs completion bash > /etc/bash_completion.d/cs

# Zsh
cs completion zsh > "${fpath[1]}/_cs"

# Fish
cs completion fish > ~/.config/fish/completions/cs.fish
```

---

## TUI/Interactive Commands

### Main Interface

Claude Squad presents a TUI with:
- Session list with status indicators
- Preview tab (session output)
- Diff tab (git changes)
- Bottom menu with available commands

### Session Management

| Key | Action |
|-----|--------|
| `n` | Create new session |
| `N` | Create new session with prompt |
| `D` | Kill (delete) selected session |
| `↑/j` | Navigate up |
| `↓/k` | Navigate down |

### Actions

| Key | Action |
|-----|--------|
| `↵` or `o` | Attach to selected session |
| `Ctrl+Q` | Detach from session |
| `s` | Commit and push branch to GitHub |
| `c` | Checkout (commit changes and pause) |
| `r` | Resume paused session |

### Navigation

| Key | Action |
|-----|--------|
| `Tab` | Switch between preview and diff tabs |
| `q` | Quit application |
| `?` | Show help menu |
| `Shift+↑/↓` | Scroll in diff view |

### Status Indicators

Sessions display status indicators:
- **Running**: Agent is actively working
- **Waiting**: Agent waiting for user input
- **Idle**: Session ready
- **Error**: Session encountered an error

---

## Configuration

### Configuration File

Configuration stored in `~/.claude-squad/config.json`:

```json
{
  "default_program": "claude",
  "profiles": [
    {
      "name": "claude",
      "program": "claude"
    },
    {
      "name": "codex",
      "program": "codex"
    },
    {
      "name": "aider",
      "program": "aider --model ollama_chat/gemma3:1b"
    }
  ]
}
```

### Configuration Location

```bash
# Find config path
cs debug

# Default locations:
# macOS: ~/.claude-squad/config.json
# Linux: ~/.config/claude-squad/config.json
```

### Profile Configuration

Define multiple agent profiles:

```json
{
  "default_program": "claude",
  "profiles": [
    {
      "name": "claude",
      "program": "claude"
    },
    {
      "name": "codex",
      "program": "codex"
    },
    {
      "name": "gemini",
      "program": "gemini"
    },
    {
      "name": "aider",
      "program": "aider --model ollama_chat/gemma3:1b"
    }
  ]
}
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `CS_CONFIG` | Path to config file |
| `CS_DATA` | Path to data directory |
| `OPENAI_API_KEY` | Required for Codex |
| `ANTHROPIC_API_KEY` | Alternative auth for Claude |

---

## Usage Examples

### Basic Workflow

```bash
# 1. Navigate to repository
cd ~/projects/my-app

# 2. Start Claude Squad
cs

# 3. Create new session (press 'n')
# Enter session name: "feature-auth"

# 4. Work with agent
# Press Enter to attach
# Type prompts, work with agent
# Press Ctrl+Q to detach

# 5. Create another session (press 'n')
# Name: "bugfix-login"

# 6. Switch between sessions with arrow keys

# 7. View diffs (press Tab to switch to diff view)

# 8. Push changes (press 's' to commit and push)
```

### Multi-Agent Workflow

```bash
# Terminal 1: Claude Code for backend
cs -p "claude"
# Create session "api-development"

# Terminal 2: Codex for frontend
# New window/tab
cs -p "codex"
# Create session "ui-development"

# Terminal 3: Aider for quick fixes
cs -p "aider"
# Create session "quick-fixes"
```

### Profile-Based Workflow

```bash
# Setup profiles in ~/.claude-squad/config.json
# Then launch with profile picker

# When you have multiple profiles:
# 1. Press 'n' to create session
# 2. Use ←/→ to select profile
# 3. Press Enter to confirm
```

### Background Processing

```bash
# Start with auto-yes for background work
cs -y

# Create session with task
# N (new with prompt)
# Name: "refactor-utils"
# Prompt: "Refactor all utility functions to use async/await"

# Session runs in background
# Switch to other work
# Check back later with preview tab
```

### Git Worktree Workflow

```bash
# Each session gets its own git worktree
# This means isolated branches per session

# 1. Create session for feature
cs
n
# 2. Name: "feature-new-ui"
# Worktree created at: .git/worktrees/feature-new-ui

# 3. Work with agent, make changes

# 4. Commit and push (press 's')
# Automatically commits and pushes branch

# 5. Create PR from pushed branch
```

### Checkout and Resume

```bash
# When you need to pause and review:

# 1. Select session
# 2. Press 'c' to checkout
#    - Commits current changes
#    - Pauses the session

# 3. Review changes externally
#    git log
#    git diff HEAD~1

# 4. Return to Claude Squad
# 5. Select paused session
# 6. Press 'r' to resume
```

### Session Cleanup

```bash
# Kill a session
cs
# Select session
# Press 'D' to kill

# Reset all sessions (careful!)
cs reset
```

---

## Troubleshooting

### Installation Issues

#### "command not found" after install

```bash
# Check if in PATH
which cs
which claude-squad

# Add to PATH
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

#### Install script fails

```bash
# Check prerequisites
which curl
which bash

# Manual install
curl -fsSL https://raw.githubusercontent.com/smtg-ai/claude-squad/main/install.sh -o install.sh
bash install.sh
```

### Session Issues

#### "failed to start new session: timed out"

```bash
# Update underlying agent
claude --version
# Update to latest

# Check tmux
which tmux
tmux -V  # Needs 3.0+

# Verify gh CLI
gh --version
gh auth status
```

#### Session shows "error" status

```bash
# Attach to see error
cs
# Select error session
# Press Enter to attach

# Or kill and recreate
# Press 'D' on error session
# Press 'n' to create new
```

#### Detach not working

```bash
# Use Ctrl+B then D (tmux detach)
# Or
ctrl+q

# If stuck, kill tmux session manually
tmux kill-session -t claude-squad-<name>
```

### TUI Issues

#### Display garbled

```bash
# Set TERM
export TERM=xterm-256color

# Check terminal size
stty size

# Resize terminal window
```

#### Colors wrong

```bash
# Force color mode
export FORCE_COLOR=1

# Or disable colors
export NO_COLOR=1
```

#### Keys not responding

```bash
# Check terminal emulator
# Supported: iTerm2, Terminal.app, Alacritty, Windows Terminal

# Try different terminal
# Check for conflicting key bindings
```

### Git Issues

#### "failed to push"

```bash
# Check auth
gh auth status

# Login to GitHub
gh auth login

# Check remote
gh repo view
```

#### Worktree conflicts

```bash
# List worktrees
git worktree list

# Clean up stale worktrees
git worktree prune

# Remove manually if needed
rm -rf .git/worktrees/<name>
git worktree prune
```

### Configuration Issues

#### Config not loading

```bash
# Check config location
cs debug

# Validate JSON
jq . ~/.claude-squad/config.json

# Reset config
rm ~/.claude-squad/config.json
cs
# Reconfigure
```

#### Profiles not showing

```bash
# Check config format
cat ~/.claude-squad/config.json

# Should have profiles array
# Default program should match profile name
```

### Agent-Specific Issues

#### Claude not starting

```bash
# Check Claude Code installed
which claude
claude --version

# Check auth
claude login
```

#### Codex not starting

```bash
# Check API key
export OPENAI_API_KEY=sk-...
# Or
echo $OPENAI_API_KEY

# Verify Codex installed
which codex
codex --version
```

#### Aider not starting

```bash
# Check Aider installed
which aider
aider --version

# Check model configuration
aider --list-models
```

### Performance Issues

#### Slow TUI response

```bash
# Reduce number of sessions
# Kill unused sessions

# Check system resources
htop

# Check tmux sessions
tmux ls
```

#### High memory usage

```bash
# Kill old sessions regularly
# Don't keep too many sessions active
# Use 'c' (checkout) to pause idle sessions
```

### Common Errors

#### "tmux session already exists"

```bash
# List tmux sessions
tmux ls

# Kill specific session
tmux kill-session -t claude-squad-<name>

# Or kill all
tmux kill-server
```

#### "git worktree already exists"

```bash
# Remove existing worktree
git worktree remove .git/worktrees/<name>
# or
git worktree remove --force .git/worktrees/<name>
```

### Debug Mode

```bash
# Run with debug output
cs debug

# Check logs
# Logs stored per session in ~/.claude-squad/
```

### Getting Help

```bash
# Show help in TUI
# Press '?' for help menu

# CLI help
cs --help

# GitHub Issues
# https://github.com/smtg-ai/claude-squad/issues
```

---

## Best Practices

1. **Name Sessions Clearly**: Use descriptive names like "feature-auth" or "bugfix-123"
2. **Regular Checkout**: Use 'c' to checkpoint and pause sessions
3. **Clean Up**: Kill sessions you're done with
4. **Use Profiles**: Set up profiles for different agents
5. **Auto-Yes Carefully**: Only use `-y` in trusted environments
6. **Git Hygiene**: Commit regularly, push when ready
7. **Monitor Resources**: Don't run too many sessions simultaneously
8. **Update Regularly**: Keep cs and agents updated

---

## Comparison with Alternatives

| Feature | Claude Squad | CCManager | Agent Deck |
|---------|-------------|-----------|------------|
| tmux-based | Yes | No | Yes |
| Multi-agent | Yes | Yes | Yes |
| Git worktrees | Yes | Yes | Yes |
| Real-time status | Basic | Detailed | Detailed |
| Web UI | No | No | Yes |
| Auto-yes | Yes | No | No |
| Setup complexity | Low | Low | Medium |

---

*Last Updated: April 2026*
