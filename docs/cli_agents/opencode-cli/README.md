# OpenCode CLI

## Overview

**OpenCode** is an open-source AI coding agent with a terminal UI and desktop app. It provides an intelligent coding assistant that runs in your terminal with support for multiple LLM providers.

**Website:** https://opencode.ai  
**Repository:** https://github.com/anomalyco/opencode

---

## Key Features

| Feature | Description |
|---------|-------------|
| **Terminal UI** | Rich terminal interface for coding |
| **Desktop App** | Available as desktop application (beta) |
| **Multi-Agent** | Switch between agents with Tab key |
| **Multi-Provider** | Support for OpenAI, Anthropic, and more |
| **Open Source** | Fully open source under Apache 2.0 |
| **Multi-Language** | Supports 17+ languages |
| **Package Managers** | npm, brew, scoop, choco, pacman, nix |

---

## Installation

### Quick Install

```bash
curl -fsSL https://opencode.ai/install | bash
```

### Package Managers

```bash
# npm
npm i -g opencode-ai@latest

# Homebrew (recommended)
brew install anomalyco/tap/opencode

# Scoop (Windows)
scoop install opencode

# Chocolatey (Windows)
choco install opencode

# Arch Linux
sudo pacman -S opencode

# Nix
nix run nixpkgs#opencode

# Mise
mise use -g opencode
```

### Desktop App

Download from [releases page](https://github.com/anomalyco/opencode/releases):

| Platform | File |
|----------|------|
| macOS (Apple Silicon) | `opencode-desktop-darwin-aarch64.dmg` |
| macOS (Intel) | `opencode-desktop-darwin-x64.dmg` |
| Windows | `opencode-desktop-windows-x64.exe` |
| Linux | `.deb`, `.rpm`, or AppImage |

---

## Quick Start

```bash
# Start OpenCode
opencode

# The terminal UI will open with two built-in agents
# Switch between agents with Tab key
```

---

## Built-in Agents

OpenCode includes two built-in agents that you can switch between:

1. **Default Agent** - General purpose coding assistant
2. **Specialized Agent** - Task-specific capabilities

Switch agents by pressing `Tab` in the terminal UI.

---

## Configuration

### Installation Directory Priority

1. `$OPENCODE_INSTALL_DIR` - Custom installation directory
2. `$XDG_BIN_DIR` - XDG Base Directory path
3. `$HOME/bin` - Standard user binary directory
4. `$HOME/.opencode/bin` - Default fallback

### Example

```bash
OPENCODE_INSTALL_DIR=/usr/local/bin curl -fsSL https://opencode.ai/install | bash
```

---

## Architecture

OpenCode is built with:
- **TypeScript** - Primary language
- **Node.js** - Runtime
- **Terminal UI** - Rich console interface
- **Desktop App** - Electron-based (beta)

```
┌─────────────────────────────────────────────────────────────┐
│                        OpenCode                              │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐   │
│  │  Terminal UI │    │  OpenCode    │    │  LLM APIs    │   │
│  │   (User)     │◄──►│   (Node.js)  │◄──►│  (Multiple)  │   │
│  └──────────────┘    └──────┬───────┘    └──────────────┘   │
│                             │                                │
│                        ┌────┴────┐                          │
│                        │  Agents │                          │
│                        └─────────┘                          │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Documentation

- [Official Website](https://opencode.ai)
- [Discord Community](https://opencode.ai/discord)
- [npm Package](https://www.npmjs.com/package/opencode-ai)

---

*Part of the HelixAgent CLI Agent Collection*
