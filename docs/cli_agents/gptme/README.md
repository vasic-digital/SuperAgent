# GPTMe

## Overview

**GPTMe** is a personal AI agent that runs in your terminal, equipped with powerful tools to execute code, edit files, browse the web, process images, and more. It acts as an intelligent copilot for your computer, enabling AI-assisted coding and general task automation.

**Official Website:** [https://gptme.org](https://gptme.org)  
**Documentation:** [https://gptme.org/docs/](https://gptme.org/docs/)  
**GitHub:** [https://github.com/gptme/gptme](https://github.com/gptme/gptme)

---

## Key Features

### Core Capabilities

| Feature | Description |
|---------|-------------|
| **Code Execution** | Execute Python and shell commands directly in your environment |
| **File Operations** | Read, write, and patch files with incremental changes |
| **Web Browsing** | Browse websites and extract information using Playwright |
| **Vision** | Analyze images, screenshots, and visual content |
| **Computer Use** | Control desktop applications through visual interface |
| **Multi-Provider** | Support for 10+ LLM providers (Anthropic, OpenAI, Google, etc.) |
| **Extensible** | Plugin system, MCP support, custom tools, and hooks |
| **Autonomous Mode** | Run as persistent agents with scheduled tasks |

### Advanced Features

- **Subagents**: Spawn specialized agents for parallel tasks
- **Context Compression**: Automatic context management for long conversations
- **Lessons System**: Contextual guidance that auto-injects based on keywords
- **RAG Integration**: Retrieve context from local files
- **Web UI**: Modern React-based interface at [chat.gptme.org](https://chat.gptme.org)
- **MCP Support**: Model Context Protocol for external tool integration
- **ACP Support**: Agent Client Protocol for editor integration

---

## Installation

### Recommended Methods

#### Using pipx (Recommended)
```bash
pipx install gptme
```

#### Using uv
```bash
uv tool install gptme
```

#### With Optional Extras
```bash
# Web browsing support
pipx install 'gptme[browser]'

# All extras
pipx install 'gptme[all]'

# From git with all extras
uv tool install 'git+https://github.com/gptme/gptme.git[all]'
```

### Requirements
- **Python**: 3.10 or newer
- **Operating Systems**: macOS, Linux (Windows via WSL/Docker)
- **API Key**: At least one LLM provider API key

---

## Quick Start

### 1. Initial Setup
```bash
# Start gptme
gptme

# You'll be prompted for an API key if not configured
```

### 2. Basic Usage
```bash
# Simple chat
gptme "Hello, what can you do?"

# With file context
gptme "explain this code" main.py

# Resume conversation
gptme -r

# Non-interactive mode
gptme -n "fix the failing tests"
```

### 3. Example Commands
```bash
# Create applications
gptme 'write a particle effect using three.js to particles.html'

# Process files
gptme 'summarize this' README.md
gptme 'what do you see?' screenshot.png

# Development workflows
git diff | gptme 'complete the TODOs in this diff'
make test | gptme 'fix the failing tests'

# Chain multiple tasks
gptme 'make a change' - 'test it' - 'commit it'
```

---

## Repository Structure

This repository (`cli_agents/gptme/`) contains:

```
gptme/
├── gptme/                      # Main Python package
│   ├── cli/                    # CLI implementations
│   ├── tools/                  # Built-in tools (shell, python, browser, etc.)
│   ├── llm/                    # LLM provider implementations
│   ├── server/                 # REST API server
│   ├── config/                 # Configuration management
│   ├── hooks/                  # Hook system
│   ├── plugins/                # Plugin system
│   └── eval/                   # Evaluation framework
├── docs/                       # Sphinx documentation
├── tests/                      # Test suite
├── webui/                      # React web interface
├── tauri/                      # Desktop app (Tauri)
├── scripts/                    # Automation scripts
├── media/                      # Audio files for tool sounds
├── pyproject.toml             # Python package configuration
└── README.md                  # This file
```

---

## Tool System

GPTMe equips the AI with a rich set of built-in tools:

| Tool | Description |
|------|-------------|
| `shell` | Execute shell commands directly in your terminal |
| `ipython` | Run Python code interactively with full library access |
| `read` | Read files and directories |
| `save` / `append` | Create or update files |
| `patch` / `morph` | Make incremental edits to existing files |
| `browser` | Browse websites via Playwright |
| `vision` | Process and analyze images |
| `screenshot` | Capture screenshots of your desktop |
| `rag` | Retrieve context from local files |
| `gh` | Interact with GitHub via CLI |
| `tmux` | Run long-lived commands in persistent sessions |
| `computer` | Full desktop access for GUI interactions |
| `subagent` | Spawn sub-agents for parallel tasks |
| `chats` | Reference and search past conversations |

Use `/tools` during a conversation to see all available tools.

---

## Extensibility

### Plugins
Extend gptme with custom tools, hooks, and commands via Python packages:

```toml
# gptme.toml
[plugins]
paths = ["~/.config/gptme/plugins", "./plugins"]
enabled = ["my_plugin"]
```

### MCP (Model Context Protocol)
Use any MCP server as a tool source:
```bash
# MCP support included by default
pipx install gptme
```

### Skills
Lightweight workflow bundles that auto-load when mentioned by name.

### Hooks
Run custom code at key lifecycle events without a full plugin.

### Lessons
Contextual guidance that auto-injects into conversations based on keywords, tools, and patterns.

---

## Autonomous Agents

GPTMe is designed to run as a **persistent autonomous agent**:

```bash
# Create and run your own agent
gptme-agent create ~/my-agent --name MyAgent
gptme-agent install   # runs on a schedule
gptme-agent status    # check on it
```

**Reference Implementation**: [Bob](https://github.com/TimeToBuildBob) is a production autonomous agent with 1700+ completed sessions that opens PRs, reviews code, manages tasks, posts on Twitter, and writes blog posts.

---

## Documentation Links

- [Architecture Documentation](./ARCHITECTURE.md) - System design and Python components
- [API Reference](./API.md) - Commands, settings, configuration reference
- [Usage Guide](./USAGE.md) - Workflows, examples, best practices
- [External References](./REFERENCES.md) - Links and resources
- [Diagrams](./DIAGRAMS.md) - Visual documentation
- [Gap Analysis](./GAP_ANALYSIS.md) - Improvement opportunities

---

## Version Information

- **Current Version**: 0.31.0
- **License**: MIT
- **Release Frequency**: Regular updates with detailed changelogs
- **Changelog**: See [docs/changelog.rst](../../../cli_agents/gptme/docs/changelog.rst)

---

## Support & Community

- **Discord**: [https://discord.gg/NMaCmmkxWv](https://discord.gg/NMaCmmkxWv)
- **GitHub Issues**: [https://github.com/gptme/gptme/issues](https://github.com/gptme/gptme/issues)
- **X/Twitter**: [@gptmeorg](https://x.com/gptmeorg)

---

## Ecosystem

| Project | Description |
|---------|-------------|
| [gptme-webui](https://github.com/gptme/gptme-webui) | Modern React web interface |
| [gptme-contrib](https://github.com/gptme/gptme-contrib) | Community plugins and scripts |
| [gptme-agent-template](https://github.com/gptme/gptme-agent-template) | Template for autonomous agents |
| [gptme-rag](https://github.com/gptme/gptme-rag) | RAG integration |
| [gptme.vim](https://github.com/gptme/gptme.vim) | Vim plugin |
| [gptme-tauri](https://github.com/gptme/gptme-tauri) | Desktop app (WIP) |

---

## License

See [LICENSE](../../../cli_agents/gptme/LICENSE)

---

*Part of the HelixAgent CLI Agent Collection*
