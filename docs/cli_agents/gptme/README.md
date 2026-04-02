# gptme

## Overview

**gptme** is a personal AI agent in your terminal, with tools enabling it to use the terminal, run code, edit files, browse the web, use vision, and much more. It's a general-purpose assistant for all kinds of knowledge work and coding tasks.

**Website:** https://gptme.org  
**Documentation:** https://gptme.org/docs/

---

## Key Features

| Feature | Description |
|---------|-------------|
| **Terminal Integration** | Native terminal experience with powerful CLI |
| **Code Execution** | Run code in multiple languages |
| **File Operations** | Read, write, and edit files |
| **Web Browsing** | Browse websites and extract information |
| **Vision Support** | Analyze images and screenshots |
| **MCP Support** | Model Context Protocol for extensibility |
| **Plugin System** | Extend with custom tools |
| **Multi-Provider** | OpenAI, Anthropic, local models |
| **Autonomous Mode** | Self-directed task execution |

---

## Installation

### Via pip

```bash
pip install gptme
```

### Via pipx (recommended)

```bash
pipx install gptme
```

### From source

```bash
git clone https://github.com/gptme/gptme.git
cd gptme
pip install -e .
```

---

## Quick Start

### 1. Set API Key

```bash
export OPENAI_API_KEY="your-key-here"
# or
export ANTHROPIC_API_KEY="your-key-here"
```

### 2. Start gptme

```bash
# Interactive mode
gptme

# With prompt
gptme "create a React todo app"

# With files
gptme --files src/main.py "explain this code"
```

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                          gptme                               │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐   │
│  │    CLI       │◄──►│   gptme      │◄──►│  LLM APIs    │   │
│  │   (User)     │    │   (Python)   │    │  (OpenAI)    │   │
│  └──────────────┘    └──────┬───────┘    └──────────────┘   │
│                             │                                │
│        ┌────────────────────┼────────────────────┐          │
│        │                    │                    │          │
│        ▼                    ▼                    ▼          │
│  ┌──────────┐        ┌──────────┐        ┌──────────┐      │
│  │  Tools   │        │  Files   │        │   MCP    │      │
│  │(shell)   │        │(R/W/Edit)│        │(External)│      │
│  └──────────┘        └──────────┘        └──────────┘      │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## Tools

| Tool | Purpose |
|------|---------|
| `shell` | Execute shell commands |
| `python` | Run Python code |
| `read` | Read files |
| `save` | Write files |
| `patch` | Apply patches |
| `browser` | Browse the web |
| `vision` | Analyze images |
| `tmux` | Terminal multiplexing |
| `mcp` | MCP server tools |

---

## Documentation

- [Getting Started](https://gptme.org/docs/getting-started.html)
- [Documentation](https://gptme.org/docs/)
- [Examples](https://gptme.org/docs/examples.html)

---

*Part of the HelixAgent CLI Agent Collection*
