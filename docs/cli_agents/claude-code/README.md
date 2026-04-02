# Claude Code

## Overview

**Claude Code** is Anthropic's official agentic coding tool that operates directly in your terminal. It understands your codebase, executes routine tasks, explains complex code, and handles git workflows through natural language commands.

**Official Documentation:** [https://code.claude.com/docs/en/overview](https://code.claude.com/docs/en/overview)

---

## Key Features

### Core Capabilities

| Feature | Description |
|---------|-------------|
| **Natural Language Interface** | Interact with your codebase using plain English commands |
| **Code Understanding** | Analyzes and comprehends complex codebases contextually |
| **Task Automation** | Executes routine development tasks automatically |
| **Git Integration** | Handles commits, branches, PRs, and git workflows |
| **Multi-Platform** | Supports macOS, Linux, and Windows |
| **IDE Integration** | Works in terminal, IDE, or via GitHub @mentions |
| **Plugin System** | Extensible architecture with custom commands and agents |
| **MCP Support** | Model Context Protocol for external tool integration |

### Advanced Features

- **Auto Mode**: Autonomous execution with permission controls
- **Subagents**: Spawn specialized agents for parallel tasks
- **Hooks**: Custom event handlers (PreToolUse, PostToolUse, Stop, etc.)
- **Skills**: Contextual knowledge injection
- **Voice Mode**: Hands-free interaction
- **Session Management**: Save, resume, and fork conversations

---

## Installation

### Recommended Methods

#### macOS / Linux
```bash
curl -fsSL https://claude.ai/install.sh | bash
```

#### Homebrew (macOS/Linux)
```bash
brew install --cask claude-code
```

#### Windows (PowerShell)
```powershell
irm https://claude.ai/install.ps1 | iex
```

#### WinGet (Windows)
```powershell
winget install Anthropic.ClaudeCode
```

### Alternative: NPM (Deprecated)
```bash
npm install -g @anthropic-ai/claude-code
```

### Requirements
- **Node.js**: 18+ (for npm installation)
- **Operating Systems**: macOS, Linux, Windows
- **Anthropic API Key**: Required for authentication

---

## Quick Start

### 1. Initial Setup
```bash
# Navigate to your project
cd /path/to/your/project

# Launch Claude Code
claude
```

### 2. Authentication
On first run, Claude Code will prompt for authentication:
- Follow the OAuth flow in your browser
- Or provide API key directly

### 3. Basic Commands
```bash
# Get help
claude --help

# Start in specific directory
claude /path/to/project

# Resume previous session
claude --resume

# Run in headless/CI mode
claude -p "your command here"
```

---

## Repository Structure

This repository (`cli_agents/claude-code/`) contains:

```
claude-code/
├── .claude/                  # Claude Code configuration
│   └── commands/             # Custom slash commands
├── .claude-plugin/           # Plugin metadata
├── .devcontainer/            # Dev container configuration
├── .github/                  # GitHub workflows and templates
│   ├── ISSUE_TEMPLATE/
│   └── workflows/            # GitHub Actions for issue management
├── examples/                 # Example configurations
│   ├── hooks/               # Hook examples
│   └── settings/            # Settings examples
├── plugins/                  # Official plugins (14 plugins)
├── scripts/                  # Automation scripts
├── CHANGELOG.md             # Version history
├── LICENSE.md               # License
├── README.md                # This file
└── SECURITY.md              # Security policy
```

---

## Plugin System

Claude Code supports a robust plugin architecture with 14 official plugins included in this repository.

### Included Plugins

| Plugin | Purpose | Components |
|--------|---------|------------|
| **agent-sdk-dev** | Agent SDK development kit | Commands, agents |
| **claude-opus-4-5-migration** | Model migration assistant | Skills |
| **code-review** | Automated PR review | Commands, 5 agents |
| **commit-commands** | Git workflow automation | Commands |
| **explanatory-output-style** | Educational output mode | Hooks |
| **feature-dev** | Feature development workflow | Commands, 3 agents |
| **frontend-design** | Frontend design guidance | Skills |
| **hookify** | Custom hook creation | Commands, agents, hooks |
| **learning-output-style** | Interactive learning mode | Hooks |
| **plugin-dev** | Plugin development toolkit | Commands, 3 agents, 7 skills |
| **pr-review-toolkit** | Comprehensive PR review | Commands, 6 agents |
| **ralph-wiggum** | Iterative development loops | Commands, hooks |
| **security-guidance** | Security reminder system | Hooks |

### Plugin Structure
```
plugin-name/
├── .claude-plugin/
│   └── plugin.json          # Plugin metadata
├── commands/                # Slash commands
├── agents/                  # Specialized agents
├── skills/                  # Contextual skills
├── hooks/                   # Event handlers
├── .mcp.json               # MCP server config
└── README.md               # Plugin docs
```

---

## Documentation Links

- [Architecture Documentation](./ARCHITECTURE.md) - System design and components
- [API Reference](./API.md) - Commands and configuration
- [Usage Guide](./USAGE.md) - Tutorials and examples
- [Development Guide](./DEVELOPMENT.md) - Contributing and development
- [External References](./REFERENCES.md) - Links and resources
- [Diagrams](./DIAGRAMS.md) - Visual documentation

---

## Version Information

- **Current Version**: 2.1.90
- **Release Frequency**: Regular updates with detailed changelogs
- **Changelog**: See [CHANGELOG.md](../../../cli_agents/claude-code/CHANGELOG.md)

---

## Support & Community

- **Discord**: [Claude Developers Discord](https://anthropic.com/discord)
- **Issues**: [GitHub Issues](https://github.com/anthropics/claude-code/issues)
- **Bug Reports**: Use `/bug` command within Claude Code

---

## Data & Privacy

Claude Code collects usage data including:
- Code acceptance/rejection feedback
- Conversation data
- Bug reports via `/bug` command

**Privacy Safeguards:**
- Limited retention periods for sensitive data
- Restricted access to user session data
- No use of feedback for model training

See [Privacy Policy](https://www.anthropic.com/legal/privacy) and [Commercial Terms](https://www.anthropic.com/legal/commercial-terms).

---

## License

See [LICENSE.md](../../../cli_agents/claude-code/LICENSE.md)

---

*Part of the HelixAgent CLI Agent Collection*
