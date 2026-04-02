# Noi - User Guide

> The ultimate AI desktop app - run multiple AI models side-by-side in a single, powerful interface.

## Table of Contents

- [Overview](#overview)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [CLI Commands](#cli-commands)
- [UI Features](#ui-features)
- [Configuration](#configuration)
- [Usage Examples](#usage-examples)
- [Troubleshooting](#troubleshooting)

---

## Overview

Noi is an open-source desktop application that unifies access to multiple AI services in a single interface. Built by `lencx` (creator of the ChatGPT Desktop Application), Noi eliminates the need to juggle multiple browser tabs for different AI services.

### Key Features

- **Multi-Model Access**: ChatGPT, Claude, Gemini, Perplexity, and more in one app
- **Noi Ask**: Send one prompt to multiple AI models simultaneously
- **Session Isolation**: Cookie isolation for multiple accounts
- **Local-First**: History, prompts, and settings stay on-device
- **Built-in Terminal**: Run local commands without leaving the app
- **Prompt Library**: Save, tag, and sync your best prompts
- **Multi-Window**: Run parallel workspaces side-by-side
- **CLI Support**: Control Noi via command line tools
- **Cross-Platform**: Available for macOS, Windows, and Linux

---

## Installation

### System Requirements

- **macOS**: 10.15 (Catalina) or later
- **Windows**: Windows 10 or later
- **Linux**: Ubuntu 18.04+, Fedora, or similar
- **RAM**: 4GB minimum (8GB recommended)
- **Storage**: 200MB free space

### Installation Methods

#### Option 1: Official Download

1. Visit https://github.com/lencx/noi/releases
2. Download the installer for your platform:
   - **macOS**: `Noi-x.x.x.dmg` or `Noi-x.x.x-mac.zip`
   - **Windows**: `Noi-x.x.x.exe` or `Noi-x.x.x-win.zip`
   - **Linux**: `Noi-x.x.x.AppImage` or `Noi-x.x.x-linux.zip`
3. Run the installer

#### Option 2: Homebrew (macOS)

```bash
brew install --cask noi
```

#### Option 3: Package Managers

**Windows (Winget):**
```powershell
winget install lencx.noi
```

**Linux (AppImage):**
```bash
# Download AppImage
wget https://github.com/lencx/noi/releases/download/v1.1.0/Noi-1.1.0.AppImage

# Make executable
chmod +x Noi-1.1.0.AppImage

# Run
./Noi-1.1.0.AppImage
```

#### Option 4: Build from Source

```bash
# Clone repository
git clone https://github.com/lencx/noi.git
cd noi

# Install dependencies
npm install

# Build
npm run build

# Run
npm start
```

### Post-Installation (macOS)

Due to macOS security, you may need to run:

```bash
xattr -cr /Applications/Noi.app
```

Or:
1. Open System Preferences → Security & Privacy
2. Click "Open Anyway" for Noi

### Verify Installation

1. Launch Noi from Applications menu
2. You should see the Noi window with a sidebar of AI services

---

## Quick Start

### 1. First Launch

When you first open Noi:

1. **Default Space**: You'll see a sidebar with various AI services
2. **Built-in Services**: ChatGPT, Claude, Gemini, Perplexity, and more
3. **Anonymous Access**: Some services work without login
4. **Sign In**: Click a service to sign in (if needed)

### 2. Basic Usage

**Chat with a single AI:**
1. Click an AI service in the sidebar (e.g., ChatGPT)
2. Start typing in the chat interface

**Use Noi Ask (multiple AIs):**
1. Type your prompt in the Noi Ask input
2. Select which AIs to query
3. View responses side-by-side

### 3. Built-in Terminal

Access the terminal:
1. Click the terminal icon in the toolbar
2. Or use keyboard shortcut (see below)
3. Run local commands alongside AI chats

### 4. Managing Spaces

Spaces are collections of services:

1. **Default Space**: All available services
2. **Custom Spaces**: Create spaces for specific workflows
3. **Switch Spaces**: Click the Spaces icon in the bottom toolbar

---

## CLI Commands

Noi includes a CLI for automation and control from terminal.

### Installation

The CLI is included with Noi. Ensure Noi is installed first.

### Global Commands

```bash
# Show version
noi --version

# Show help
noi --help

# Launch Noi
noi launch

# Launch with specific URL
noi launch --url https://chat.openai.com
```

### Window Management

```bash
# Open new window
noi window new

# Open specific service in new window
noi window new --service chatgpt

# List windows
noi window list

# Close window
noi window close --id <window-id>

# Focus window
noi window focus --id <window-id>
```

### Space Management

```bash
# List spaces
noi space list

# Create new space
noi space create --name "Development"

# Delete space
noi space delete --name "Development"

# Switch to space
noi space switch --name "Development"

# Add service to space
noi space add-service --space "Development" --service github

# Remove service from space
noi space remove-service --space "Development" --service twitter
```

### Service Management

```bash
# List available services
noi service list

# Add custom service
noi service add --name "CustomAI" --url https://custom-ai.com

# Remove custom service
noi service remove --name "CustomAI"

# Enable service
noi service enable --name "chatgpt"

# Disable service
noi service disable --name "chatgpt"
```

### Prompt Management

```bash
# List prompts
noi prompt list

# Add prompt
noi prompt add --name "Code Review" --content "Review this code for bugs..."

# Delete prompt
noi prompt delete --name "Code Review"

# Use prompt
noi prompt use --name "Code Review" --service claude

# Export prompts
noi prompt export --file ~/noi-prompts.json

# Import prompts
noi prompt import --file ~/noi-prompts.json
```

### Noi Ask Commands

```bash
# Send ask to multiple services
noi ask --prompt "Explain quantum computing" --services chatgpt,claude,gemini

# Quick ask (uses default services)
noi ask "What is machine learning?"
```

### Settings Commands

```bash
# Show settings
noi settings show

# Set setting
noi settings set --key theme --value dark

# Get setting
noi settings get --key theme

# Reset to defaults
noi settings reset
```

### Plugin Commands

```bash
# List plugins
noi plugin list

# Install plugin
noi plugin install --name <plugin-name>

# Uninstall plugin
noi plugin uninstall --name <plugin-name>

# Enable plugin
noi plugin enable --name <plugin-name>

# Disable plugin
noi plugin disable --name <plugin-name>
```

### Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Cmd/Ctrl + T` | New tab |
| `Cmd/Ctrl + W` | Close tab |
| `Cmd/Ctrl + Shift + T` | Reopen closed tab |
| `Cmd/Ctrl + L` | Focus address bar |
| `Cmd/Ctrl + `` | Toggle terminal |
| `Cmd/Ctrl + Shift + N` | New window |
| `Cmd/Ctrl + ,` | Open settings |
| `Cmd/Ctrl + Shift + A` | Open Noi Ask |

---

## UI Features

### Main Interface

```
┌─────────────────────────────────────────────────────────┐
│  Noi                                         [_][□][X]  │
├──────┬──────────────────────────────────────────────────┤
│      │                                                  │
│  🤖  │  ChatGPT                                    🔄  │
│  💬  │  ┌────────────────────────────────────────────┐ │
│  🔍  │  │ Hello! How can I help you today?          │ │
│  📝  │  └────────────────────────────────────────────┘ │
│      │                                                  │
│ Chat │  > Type your message...                         │
│ Claude│                                                 │
│ Gemini│  [Send]                                         │
│       │                                                  │
│ Perp. ├──────────────────────────────────────────────────┤
│       │  Terminal                                        │
│ ...   │  $ _                                             │
│       │                                                  │
└───────┴──────────────────────────────────────────────────┘
```

### Spaces

**Creating a Custom Space:**

1. Click the Spaces icon (bottom toolbar)
2. Click "+" to create new space
3. Name your space (e.g., "Development")
4. Add services to the space

**Space Use Cases:**
- **Development**: GitHub, Stack Overflow, Documentation
- **Research**: Perplexity, Google Scholar, Wikipedia
- **Writing**: ChatGPT, Claude, Grammarly
- **Personal**: ChatGPT, Claude, your local Ollama

### Noi Ask

**Using Noi Ask:**

1. Click the Noi Ask icon or press `Cmd/Ctrl + Shift + A`
2. Type your prompt
3. Select which AI services to query
4. View all responses side-by-side

**Benefits:**
- Compare answers from multiple models
- Get diverse perspectives
- Choose the best response
- Save time switching between tabs

### Prompt Library

**Managing Prompts:**

1. Open Prompt Library (sidebar or menu)
2. Click "+" to add new prompt
3. Add:
   - Name: "React Component Generator"
   - Tags: coding, react
   - Content: Your prompt template
4. Save and use anytime

**Using Prompts:**
1. Open Prompt Library
2. Search or browse your prompts
3. Click to insert into active chat
4. Edit as needed before sending

### Session Isolation

**Multiple Accounts:**

1. Right-click a service in sidebar
2. Select "New Session"
3. Log in with different credentials
4. Switch between sessions seamlessly

**Cookie Isolation:**
- Each session has isolated cookies
- Use personal and work accounts simultaneously
- No need for multiple browser profiles

### Built-in Terminal

**Accessing Terminal:**

1. Click terminal icon in toolbar
2. Or press `Cmd/Ctrl + `` `
3. Run commands alongside AI conversations

**Terminal Features:**
- Full shell access
- Run local LLMs (Ollama)
- Git commands
- npm/python/etc. commands
- Split view with AI chats

### Multi-Window Management

**Creating Windows:**

1. `Cmd/Ctrl + Shift + N` for new window
2. Drag tabs between windows
3. Each window can have different space

**Use Cases:**
- Research in one window, coding in another
- Compare different AI responses
- Reference documentation while coding

---

## Configuration

### Settings Location

- **macOS**: `~/Library/Application Support/Noi/config.json`
- **Windows**: `%APPDATA%/Noi/config.json`
- **Linux**: `~/.config/Noi/config.json`

### Configuration Structure

```json
{
  "version": "1.1.0",
  "theme": "dark",
  "window": {
    "width": 1400,
    "height": 900,
    "fullscreen": false,
    "alwaysOnTop": false
  },
  "sidebar": {
    "width": 250,
    "collapsed": false,
    "showIcons": true,
    "showLabels": true
  },
  "spaces": [
    {
      "id": "default",
      "name": "All",
      "services": ["chatgpt", "claude", "gemini", "perplexity"]
    },
    {
      "id": "dev",
      "name": "Development",
      "services": ["github", "stackoverflow", "docs"]
    }
  ],
  "services": {
    "chatgpt": {
      "enabled": true,
      "url": "https://chat.openai.com",
      "shortcut": "Cmd+1"
    },
    "claude": {
      "enabled": true,
      "url": "https://claude.ai",
      "shortcut": "Cmd+2"
    },
    "gemini": {
      "enabled": true,
      "url": "https://gemini.google.com",
      "shortcut": "Cmd+3"
    }
  },
  "noiAsk": {
    "defaultServices": ["chatgpt", "claude", "gemini"],
    "showDividers": true,
    "syncScroll": false
  },
  "prompts": {
    "library": [
      {
        "id": "1",
        "name": "Code Review",
        "tags": ["coding", "review"],
        "content": "Review this code for bugs and improvements..."
      }
    ]
  },
  "terminal": {
    "enabled": true,
    "defaultShell": "/bin/zsh",
    "fontSize": 14,
    "opacity": 0.9
  },
  "shortcuts": {
    "newTab": "Cmd+T",
    "closeTab": "Cmd+W",
    "toggleTerminal": "Cmd+`",
    "openNoiAsk": "Cmd+Shift+A"
  }
}
```

### Custom Services

Add custom AI services or websites:

```json
{
  "services": {
    "my-custom-ai": {
      "name": "My Custom AI",
      "url": "https://my-ai.example.com",
      "icon": "path/to/icon.png",
      "enabled": true,
      "shortcut": "Cmd+9"
    }
  }
}
```

### Themes

Available themes:
- `dark` (default)
- `light`
- `auto` (follows system)
- `frosted` (glass effect)

```bash
# Set via CLI
noi settings set --key theme --value frosted
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `NOI_HOME` | Custom config directory |
| `NOI_THEME` | Override theme |
| `NOI_DEBUG` | Enable debug mode |
| `NOI_DISABLE_GPU` | Disable GPU acceleration |

---

## Usage Examples

### Example 1: Research Task

```
1. Create new space: "Research"
2. Add services: Perplexity, Google Scholar, Wikipedia
3. Use Noi Ask: "Latest developments in quantum computing"
4. Compare responses from all three sources
5. Save useful prompts to library
```

### Example 2: Coding Workflow

```
1. Open Claude in main window
2. Open terminal (Cmd+`)
3. Start local development server
4. Ask Claude: "Review this error" + paste error
5. Switch to GitHub space to check issues
6. Use terminal to commit changes
```

### Example 3: Multi-Account Setup

```
1. Right-click ChatGPT → New Session
2. Log in with work account
3. Right-click ChatGPT → New Session again
4. Log in with personal account
5. Switch between sessions as needed
```

### Example 4: Prompt Library Workflow

```
1. Create prompt: "React Component Generator"
   Content: "Create a React functional component 
   named {name} that {description}. Include 
   PropTypes and basic styling."

2. Save with tags: react, component, boilerplate

3. Use in ChatGPT:
   - Open Prompt Library
   - Click "React Component Generator"
   - Fill: name="UserCard", description="displays user info"
   - Send
```

### Example 5: Local LLM Integration

```bash
# In Noi terminal:
$ ollama serve

# In another terminal tab:
$ ollama pull llama3.1

# Add to Noi as custom service:
noi service add --name "Local Ollama" --url http://localhost:11434
```

### Example 6: CLI Automation

```bash
# Daily workflow script
#!/bin/bash

# Open Noi with development space
noi launch
noi space switch --name "Development"

# Open new window with research space
noi window new
noi space switch --name "Research"

# Send initial prompts
noi ask "What are today's top tech news?" --services perplexity
```

### Example 7: Team Prompt Sharing

```bash
# Export your prompts
noi prompt export --file ./team-prompts.json

# Share file with team
# Team members import:
noi prompt import --file ./team-prompts.json
```

---

## Troubleshooting

### Installation Issues

#### "App is damaged" (macOS)

```bash
# Remove quarantine attribute
xattr -cr /Applications/Noi.app

# Or in System Preferences:
# Security & Privacy → Open Anyway
```

#### "Installation failed" (Windows)

1. Run installer as Administrator
2. Check Windows Defender isn't blocking
3. Install Visual C++ Redistributable if needed

#### "AppImage won't run" (Linux)

```bash
# Make executable
chmod +x Noi-*.AppImage

# Install FUSE if needed
sudo apt install libfuse2

# Or use --appimage-extract and run
./Noi-*.AppImage --appimage-extract
./squashfs-root/AppRun
```

### Runtime Issues

#### "Blank screen on launch"

```bash
# Disable GPU acceleration
export NOI_DISABLE_GPU=1
noi

# Or delete GPU cache
rm -rf ~/Library/Application\ Support/Noi/GPUCache  # macOS
```

#### "Services won't load"

1. Check internet connection
2. Try refreshing: `Cmd/Ctrl + R`
3. Clear service cache:
   ```bash
   rm -rf ~/Library/Application\ Support/Noi/Cache
   ```

#### "Cannot sign in to service"

1. Check if service requires specific browser
2. Try "New Session" for fresh cookies
3. Clear cookies for specific service:
   - Right-click service → Clear Cookies

#### "Terminal not working"

1. Check default shell exists:
   ```bash
   which $SHELL
   ```
2. Change default shell in settings:
   ```bash
   noi settings set --key terminal.defaultShell --value /bin/bash
   ```

### Performance Issues

#### "High CPU/Memory usage"

```bash
# Limit number of services
# Disable unused services
noi service disable --name <service>

# Close unused windows/tabs
# Enable "Optimize Performance" in settings
```

#### "Slow startup"

```bash
# Disable services you don't use
# Clear cache
rm -rf ~/Library/Application\ Support/Noi/Cache
rm -rf ~/Library/Application\ Support/Noi/Code\ Cache
```

### CLI Issues

#### "noi command not found"

```bash
# Ensure Noi.app is in Applications (macOS)
# Or use full path
/Applications/Noi.app/Contents/MacOS/noi

# Add to PATH
export PATH="$PATH:/Applications/Noi.app/Contents/MacOS"
```

### Debug Mode

```bash
# Enable debug logging
export NOI_DEBUG=1
noi

# Check logs
tail -f ~/Library/Logs/Noi/log.log  # macOS
```

### Getting Help

```bash
# Show help
noi --help

# Check version
noi --version

# Online resources
# GitHub: https://github.com/lencx/noi
# Discussions: https://github.com/lencx/noi/discussions
```

### Reporting Issues

1. Check existing issues: https://github.com/lencx/noi/issues
2. Include:
   - Noi version
   - Operating system and version
   - Steps to reproduce
   - Error messages
   - Screenshots if applicable

---

## Best Practices

### 1. Organize with Spaces

Create spaces for different contexts:
- Work
- Personal
- Research
- Development

### 2. Use Prompt Library

Save frequently used prompts:
- Code review templates
- Boilerplate generators
- Analysis frameworks

### 3. Session Management

- Use sessions for multiple accounts
- Name sessions descriptively
- Close unused sessions

### 4. Terminal Integration

- Use built-in terminal for quick commands
- Run local LLMs (Ollama) in terminal
- Keep terminal visible while chatting

### 5. Noi Ask for Comparison

- Use for important decisions
- Compare 3+ models when possible
- Save best responses

---

## Resources

- **GitHub**: https://github.com/lencx/noi
- **Releases**: https://github.com/lencx/noi/releases
- **Discussions**: https://github.com/lencx/noi/discussions
- **ChatGPT Desktop** (by same author): https://github.com/lencx/chatgpt

---

*Last updated: 2026-04-02*
