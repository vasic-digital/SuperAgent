# OpenCode CLI User Guide

## Table of Contents
1. [Installation](#installation)
2. [Quick Start](#quick-start)
3. [CLI Commands](#cli-commands)
4. [TUI/Interactive Commands](#tuiinteractive-commands)
5. [Configuration](#configuration)
6. [API/Protocol Endpoints](#api-protocol-endpoints)
7. [Usage Examples](#usage-examples)
8. [Troubleshooting](#troubleshooting)

---

## Installation

### Method 1: Quick Install (Recommended)

```bash
curl -fsSL https://opencode.ai/install | bash
```

### Method 2: Package Manager (NPM)

```bash
npm i -g opencode-ai@latest
```

### Method 3: Homebrew

```bash
brew install anomalyco/tap/opencode
```

### Method 4: Scoop (Windows)

```powershell
scoop install opencode
```

### Method 5: Chocolatey (Windows)

```powershell
choco install opencode
```

### Method 6: Arch Linux

```bash
sudo pacman -S opencode
```

### Method 7: Nix

```bash
nix run nixpkgs#opencode
```

### Method 8: Mise

```bash
mise use -g opencode
```

### Method 9: Build from Source

```bash
git clone https://github.com/anomalyco/opencode.git
cd opencode
npm install
npm run build
npm install -g .
```

---

## Quick Start

```bash
# Verify installation
opencode --version
opencode --help

# Start OpenCode
opencode

# The terminal UI will open with two built-in agents
# Switch between agents with Tab key
```

---

## CLI Commands

### Global Options

| Option | Short | Description | Example |
|--------|-------|-------------|---------|
| `--help` | `-h` | Show help | `opencode --help` |
| `--version` | `-v` | Show version | `opencode --version` |

### Command: opencode

**Description:** Start the OpenCode terminal UI.

**Usage:**
```bash
opencode [options]
```

**Examples:**
```bash
# Start OpenCode
opencode
```

**Exit Codes:**
- `0` - Success
- `1` - General error
- `130` - Interrupted (Ctrl+C)

---

## TUI/Interactive Commands

Once inside the OpenCode TUI:

| Command | Shortcut | Description |
|---------|----------|-------------|
| `Tab` | Tab | Switch between agents |
| `Ctrl+C` | | Exit OpenCode |

### Built-in Agents

OpenCode includes two built-in agents:

1. **Default Agent** - General purpose coding assistant
2. **Specialized Agent** - Task-specific capabilities

Switch between agents by pressing `Tab` in the terminal UI.

---

## Configuration

### Installation Directory Priority

OpenCode checks for installation in this order:

1. `$OPENCODE_INSTALL_DIR` - Custom installation directory
2. `$XDG_BIN_DIR` - XDG Base Directory path
3. `$HOME/bin` - Standard user binary directory
4. `$HOME/.opencode/bin` - Default fallback

### Custom Installation Directory

```bash
OPENCODE_INSTALL_DIR=/usr/local/bin curl -fsSL https://opencode.ai/install | bash
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `OPENCODE_INSTALL_DIR` | Custom installation directory |
| `XDG_BIN_DIR` | XDG Base Directory path |
| `OPENAI_API_KEY` | OpenAI API key |
| `ANTHROPIC_API_KEY` | Anthropic API key |

---

## API/Protocol Endpoints

### Supported Providers

OpenCode supports multiple LLM providers:

| Provider | Environment Variable |
|----------|---------------------|
| OpenAI | `OPENAI_API_KEY` |
| Anthropic | `ANTHROPIC_API_KEY` |
| Google | `GOOGLE_API_KEY` |
| Others | Provider-specific |

---

## Usage Examples

### Example 1: Basic Usage

```bash
# Install OpenCode
curl -fsSL https://opencode.ai/install | bash

# Start OpenCode
opencode

# The TUI opens with two agents
# Type your requests naturally

# Switch agents with Tab key
```

### Example 2: Multi-Agent Workflow

```bash
opencode

# With Default Agent:
> Explain the architecture of this project

# Press Tab to switch to Specialized Agent:
> Create a detailed implementation plan for the authentication system

# Press Tab to switch back:
> Write the code for the auth middleware
```

### Example 3: Project Analysis

```bash
cd my-project
opencode

# In the TUI:
> What does this project do?
> List all the API endpoints
> Find potential security issues
```

### Example 4: Code Generation

```bash
opencode

# In the TUI:
> Create a React component for user profile
> Add TypeScript types
> Write unit tests for this component
```

---

## Troubleshooting

### Issue: Command not found after installation

**Solution:**
```bash
# Check installation directory
which opencode
echo $OPENCODE_INSTALL_DIR

# Add to PATH if needed
export PATH="$HOME/.opencode/bin:$PATH"
```

### Issue: Installation script fails

**Solution:**
```bash
# Try manual installation
npm i -g opencode-ai@latest

# Or use package manager for your OS
brew install anomalyco/tap/opencode
```

### Issue: Agent switching not working

**Solution:**
- Ensure you're in the TUI (not inline mode)
- Press `Tab` to switch agents
- Check terminal supports key detection

---

## Additional Resources

- **Website:** https://opencode.ai
- **Discord:** https://opencode.ai/discord
- **npm Package:** https://www.npmjs.com/package/opencode-ai
- **GitHub:** https://github.com/anomalyco/opencode
