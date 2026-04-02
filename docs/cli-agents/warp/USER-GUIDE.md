# Warp Terminal User Guide

## Overview

Warp is an AI-powered terminal built for modern development workflows. It combines a Rust-based terminal emulator with integrated AI capabilities, enabling developers to use natural language for command execution, get intelligent command suggestions, and collaborate with team members. Warp features Agent Mode for autonomous task execution, Warp AI for command assistance, and Warp Drive for team collaboration.

**Key Features:**
- AI-powered command suggestions and error explanations
- Agent Mode for natural language task execution
- Modern text editing (IDE-like features in terminal)
- Block-based command output
- Warp Drive for workflows and team collaboration
- Multi-platform support (macOS, Linux, Windows via WSL)
- Native AI integration (Claude, GPT-4o)
- MCP (Model Context Protocol) support
- Custom themes and keybindings
- Real-time collaboration

---

## Installation Methods

### Method 1: macOS (Official Installer)

```bash
# Using Homebrew
brew install --cask warp

# Or download from website
# https://app.warp.dev/get_warp
```

### Method 2: Linux

```bash
# Debian/Ubuntu (.deb)
wget https://releases.warp.dev/stable/v0.2024.06.11.08.02.stable_03/Warp.deb
sudo dpkg -i Warp.deb
sudo apt-get install -f  # Fix dependencies if needed

# Or AppImage
wget https://releases.warp.dev/stable/v0.2024.06.11.08.02.stable_03/Warp-x86_64.AppImage
chmod +x Warp-x86_64.AppImage
./Warp-x86_64.AppImage

# Or via package manager (when available)
# Arch Linux: yay -S warp-terminal
```

### Method 3: Windows (WSL)

```bash
# In WSL terminal
wget https://releases.warp.dev/stable/v0.2024.06.11.08.02.stable_03/Warp.deb
sudo dpkg -i Warp.deb
```

**Note:** Warp requires WSL2 and a Linux distribution.

### Method 4: Update Existing Installation

```bash
# macOS
brew upgrade --cask warp

# Linux (Debian)
sudo apt update
sudo apt upgrade warp
```

### Verify Installation

```bash
# Check version
warp --version

# Or in Warp terminal
# Settings > About
```

---

## Quick Start

### 1. First Launch

1. Open Warp from Applications menu or Launchpad
2. Complete onboarding wizard
3. Sign in with email or SSO (optional)
4. Choose theme (Dark/Light)
5. Configure default shell (bash/zsh/fish)

### 2. Basic Navigation

```bash
# Warp opens with input at the bottom
# Type commands as usual
ls -la
cd projects/my-app
git status

# Use IDE features:
# - Cmd/Ctrl + Click to open files
# - Syntax highlighting
# - Auto-suggestions
# - Command completions
```

### 3. Using Warp AI

```bash
# Type # to ask AI
# "How do I find large files in this directory?"

# Right-click output and "Ask Warp AI"
# To explain errors

# Select text and Cmd/Ctrl+K to ask about it
```

### 4. Using Agent Mode

```bash
# Type natural language
"Delete all merged git branches"

# Or use # prefix
# "Set up a new React project with TypeScript"

# Agent will:
# 1. Understand your request
# 2. Suggest commands
# 3. Ask for approval
# 4. Execute and verify
```

---

## CLI Commands Reference

### Core Commands (In Warp Terminal)

| Command | Description |
|---------|-------------|
| `warp` | Launch Warp |
| `warp --version` | Show version |
| `warp --help` | Show help |
| `warp --hidden` | Launch without window |

### Window Management

| Shortcut | Action |
|----------|--------|
| `Cmd/Ctrl + N` | New window |
| `Cmd/Ctrl + T` | New tab |
| `Cmd/Ctrl + W` | Close tab |
| `Cmd/Ctrl + Shift + W` | Close window |
| `Cmd/Ctrl + Shift + T` | Reopen closed tab |
| `Cmd/Ctrl + 1-9` | Switch to tab |
| `Cmd/Ctrl + Shift + ]` | Next tab |
| `Cmd/Ctrl + Shift + [` | Previous tab |

### Pane Management

| Shortcut | Action |
|----------|--------|
| `Cmd/Ctrl + D` | Split pane vertically |
| `Cmd/Ctrl + Shift + D` | Split pane horizontally |
| `Cmd/Ctrl + Option + Arrow` | Navigate panes |
| `Cmd/Ctrl + Shift + Enter` | Toggle pane zoom |
| `Cmd/Ctrl + W` | Close pane |

### Text Editing

| Shortcut | Action |
|----------|--------|
| `Cmd/Ctrl + A` | Beginning of line |
| `Cmd/Ctrl + E` | End of line |
| `Cmd/Ctrl + K` | Clear to end of line |
| `Cmd/Ctrl + U` | Clear entire line |
| `Cmd/Ctrl + L` | Clear screen |
| `Cmd/Ctrl + F` | Find in buffer |
| `Cmd/Ctrl + Shift + F` | Find in all blocks |

### Block Operations

| Shortcut | Action |
|----------|--------|
| `Cmd/Ctrl + Click` | Open file/link |
| `Cmd/Ctrl + Shift + C` | Copy block output |
| `Cmd/Ctrl + Shift + G` | Copy block command |
| `Cmd/Ctrl + Shift + B` | Bookmark block |
| `Click block header` | Select entire block |

### AI Commands

| Action | How To |
|--------|--------|
| Ask AI | Type `#` followed by question |
| Explain error | Right-click error > Ask Warp AI |
| Command suggestion | Type description, press Tab |
| Agent Mode | Type natural language request |
| Select AI model | Settings > AI > Model |

### Warp Drive Commands

| Action | How To |
|--------|--------|
| Open Drive | `Cmd/Ctrl + Shift + D` |
| Search workflows | Type in Drive search |
| Run workflow | Click or use assigned shortcut |
| Create workflow | `Cmd/Ctrl + Shift + W` |
| Edit workflow | Right-click > Edit |

---

## Configuration

### Settings File Location

```
macOS:     ~/Library/Application Support/dev.warp.Warp-Stable/settings.json
Linux:     ~/.config/warp-terminal/settings.json
Windows:   %APPDATA%\warp-terminal\settings.json
```

### Key Configuration Options

```json
{
  "theme": "Dark",
  "font": {
    "family": "JetBrains Mono",
    "size": 14,
    "ligatures": true
  },
  "shell": {
    "default": "zsh",
    "args": ["-l"]
  },
  "ai": {
    "enabled": true,
    "default_model": "claude-3.5-sonnet",
    "available_models": [
      "claude-3.5-sonnet",
      "claude-3-haiku",
      "gpt-4o",
      "gpt-4o-mini"
    ],
    "agent_mode": {
      "enabled": true,
      "auto_detect": true,
      "confirmation_required": true
    }
  },
  "features": {
    "blocks": true,
    "completions": true,
    "auto_suggestions": true,
    "syntax_highlighting": true,
    "warp_drive": true
  },
  "appearance": {
    "window_opacity": 1.0,
    "background_blur": false,
    "show_tab_bar": true,
    "show_status_bar": true
  },
  "keybindings": {
    "preset": "default",
    "custom": {
      "ctrl+t": "new_tab",
      "ctrl+shift+f": "search"
    }
  },
  "network": {
    "proxy": {
      "enabled": false,
      "url": ""
    }
  }
}
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `WARP_THEMES_DIR` | Custom themes directory |
| `WARP_WORKFLOWS_DIR` | Custom workflows directory |
| `WARP_DISABLE_AI` | Disable AI features |
| `OPENAI_API_KEY` | Custom OpenAI API key |
| `ANTHROPIC_API_KEY` | Custom Anthropic API key |

### Shell Integration

**For bash:**
```bash
# Add to ~/.bashrc
source ~/.warp/bash_warp.sh
```

**For zsh:**
```zsh
# Add to ~/.zshrc
source ~/.warp/zsh_warp.zsh
```

**For fish:**
```fish
# Add to ~/.config/fish/config.fish
source ~/.warp/fish_warp.fish
```

---

## Usage Examples

### Example 1: Basic AI Command Suggestion

```bash
# Type description followed by Tab
# Input: "show git log with graph"
# Suggestion: git log --graph --oneline --all

# Press Tab to accept, or keep typing
```

### Example 2: Agent Mode Task

```bash
# Type natural language request:
"Delete all local git branches that have been merged"

# Agent Mode will:
# 1. Detect this as natural language
# 2. Suggest command: git branch --merged | grep -v \\* | xargs -n 1 git branch -d
# 3. Ask for approval
# 4. Execute if approved
# 5. Verify result
```

### Example 3: Error Explanation

```bash
# Run a command that fails
npm install some-package

# Error appears in red block
# Right-click on error output
# Select "Ask Warp AI"

# Warp AI explains:
# "This error indicates the package doesn't exist. 
# Did you mean 'some-package-name'?"
```

### Example 4: Creating Workflows

```bash
# Open Warp Drive
Cmd/Ctrl + Shift + D

# Click "New Workflow"
Name: Deploy to Production
Description: Full deployment pipeline

# Add commands:
git pull origin main
npm ci
npm run build
npm run test
npm run deploy

# Save with shortcut:
Cmd/Ctrl + Shift + W

# Use workflow:
# Click in Drive or use shortcut
```

### Example 5: Team Collaboration

```bash
# Create team in Warp Drive
# Invite team members via email

# Share workflows:
# 1. Create workflow
# 2. Click Share
# 3. Select team members

# Shared workflows appear in team Drive
# Everyone can run them
```

### Example 6: MCP Server Integration

```bash
# Add MCP server in Settings > AI > MCP

# Example: PostgreSQL MCP
Name: postgres-local
Command: docker run -i --rm mcp/postgres postgresql://localhost/mydb

# Use in Agent Mode:
"Query the database for active users"
# Agent will use MCP server to execute
```

### Example 7: Custom Themes

```bash
# Create custom theme
mkdir -p ~/.warp/themes
cat > ~/.warp/themes/my-theme.yaml << 'EOF'
name: My Custom Theme
details: darker
background: "#1a1a2e"
foreground: "#eee"
accent: "#e94560"
cursor: "#e94560"
selection: "#16213e"
EOF

# Select in Settings > Appearance > Theme
```

### Example 8: SSH with Warp

```bash
# SSH to remote server
ssh user@server

# Warp features work over SSH:
# - AI suggestions
# - Syntax highlighting
# - Block navigation
# - Command completions

# Note: Requires Warp SSH wrapper
# Settings > Features > SSH
```

### Example 9: Git Workflow with AI

```bash
# Natural language git operations
"Show me commits from last week by author"
"Create a feature branch for user authentication"
"Revert the last commit but keep changes"

# Agent Mode handles git operations
# with proper verification
```

### Example 10: Multi-Step Automation

```bash
# Complex request:
"Set up a new Python project with poetry, 
 add pytest, black, and flake8, 
 create src/ and tests/ directories,
 and initialize git"

# Agent Mode:
# 1. Creates directory structure
# 2. Initializes poetry
# 3. Adds dependencies
# 4. Creates config files
# 5. Initializes git
# 6. Verifies setup
```

---

## TUI / Interactive Features

### Command Palette

```bash
# Open with Cmd/Ctrl + Shift + P

Available actions:
- New tab/window/pane
- Search commands
- Open workflows
- Change theme
- Access settings
- Run AI commands
```

### Block Navigation

```bash
# Navigate command history as blocks
# Each command+output is a block

# Click block header to select
# Scroll through blocks
# Copy individual blocks
# Bookmark important blocks
```

### AI Chat Panel

```bash
# Open with Cmd/Ctrl + L

# Persistent AI conversation
# Reference previous commands
# Get contextual help
# Multi-turn conversations
```

### Settings UI

```bash
# Open with Cmd/Ctrl + ,

# Configure:
- Appearance (themes, fonts)
- AI settings (models, preferences)
- Features (blocks, completions)
- Keybindings
- Network (proxy)
- Integrations (SSH, shells)
```

---

## Warp Drive (Workflows)

### Workflow Structure

```yaml
# Example workflow
name: Database Backup
description: Backup PostgreSQL database
author: john@example.com
tags: [database, backup, postgres]

parameters:
  - name: database
    description: Database name
    default: production
  
  - name: backup_dir
    description: Backup directory
    default: /backups

commands:
  - echo "Starting backup of {{database}}..."
  - pg_dump {{database}} > {{backup_dir}}/{{database}}_$(date +%Y%m%d).sql
  - echo "Backup complete!"

# Keyboard shortcut (optional)
shortcut: ctrl+shift+b
```

### Creating Parameterized Workflows

```bash
# In Warp Drive:
1. Click "New Workflow"
2. Add parameters with {{param_name}}
3. Define default values
4. Test workflow
5. Save and share
```

### Workflow Commands

| Action | Command/Shortcut |
|--------|------------------|
| Open Drive | Cmd/Ctrl + Shift + D |
| New Workflow | Cmd/Ctrl + Shift + W |
| Run Workflow | Click or assigned shortcut |
| Search Workflows | Cmd/Ctrl + Shift + D, then type |
| Edit Workflow | Right-click > Edit |
| Share Workflow | Right-click > Share |

---

## Troubleshooting

### Installation Issues

**Problem:** Installation fails on Linux

**Solutions:**
```bash
# Install dependencies
sudo apt-get update
sudo apt-get install -f

# Check architecture
uname -m  # Should be x86_64

# Download correct package
# For ARM64: Warp-arm64.deb
```

**Problem:** "Application is damaged" on macOS

**Solutions:**
```bash
# Remove quarantine attribute
xattr -cr /Applications/Warp.app

# Or allow in System Preferences
# Security & Privacy > Open Anyway
```

### AI Features Not Working

**Problem:** AI suggestions not appearing

**Solutions:**
```bash
# Check AI is enabled
# Settings > AI > Enabled

# Check internet connection
ping warp.dev

# Verify not behind restrictive proxy
# Settings > Network > Proxy

# Check account status
# Settings > Account
```

**Problem:** Agent Mode not activating

**Solutions:**
```bash
# Check Agent Mode enabled
# Settings > AI > Agent Mode

# Verify auto-detection on
# Settings > AI > Auto-detect natural language

# Manual activation: type # before request

# Check for denylist conflicts
# Settings > AI > Denylist
```

### Performance Issues

**Problem:** Warp is slow or laggy

**Solutions:**
```bash
# Disable GPU acceleration
# Settings > Appearance > Advanced > Disable GPU

# Reduce scrollback buffer
# Settings > Features > Scrollback

# Close unused tabs/panes

# Check system resources
htop

# Update to latest version
```

### Shell Integration Issues

**Problem:** Prompt not showing correctly

**Solutions:**
```bash
# Reinstall shell integration
# Settings > Features > Shell Integration

# Check shell config
cat ~/.zshrc | grep warp

# Manually add integration
curl -sSL https://warp.dev/install.sh | bash
```

**Problem:** Environment variables not loaded

**Solutions:**
```bash
# Check shell startup files
cat ~/.zshrc
cat ~/.zprofile

# Ensure Warp sources them
# Settings > Shell > Load login environment

# Restart Warp after changes
```

### Common Error Messages

| Error | Solution |
|-------|----------|
| "Connection failed" | Check internet/proxy settings |
| "AI unavailable" | Check subscription/API keys |
| "Command not found" | Install missing tool |
| "Permission denied" | Check file permissions |
| "Port already in use" | Kill process using port |

### Getting Help

```bash
# Built-in help
# Cmd/Ctrl + Shift + P > Help

# Check for updates
# Settings > About > Check for Updates

# Debug mode
RUST_LOG=debug warp

# Community
# https://warp.dev/discord
# https://github.com/warpdotdev/Warp

# Support
# https://warp.dev/support
```

---

## Best Practices

### 1. Security
- Review Agent Mode commands before approving
- Use `warp disable-agent` for sensitive work
- Enable confirmation for destructive operations
- Keep Warp updated for security patches

### 2. Productivity
- Learn keyboard shortcuts
- Create reusable workflows
- Use bookmarks for important outputs
- Organize workflows with tags

### 3. Team Collaboration
- Share useful workflows
- Document workflow parameters
- Use consistent naming conventions
- Sync team settings

### 4. AI Usage
- Be specific in natural language requests
- Review generated commands
- Use # for AI queries
- Combine AI with manual commands

### 5. Customization
- Create themes matching your IDE
- Set up custom keybindings
- Configure default shells per directory
- Use project-specific settings

---

## Resources

- **Website:** https://www.warp.dev
- **GitHub:** https://github.com/warpdotdev/Warp
- **Documentation:** https://docs.warp.dev
- **Discord:** https://warp.dev/discord
- **Blog:** https://warp.dev/blog

---

*Last Updated: April 2026*
