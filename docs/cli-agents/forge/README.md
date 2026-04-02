# Forge CLI Agent

## Overview

**Forge** is a Rust-based AI-enhanced terminal development environment that integrates AI capabilities directly into your command-line workflow. Unlike browser-based AI coding tools, Forge operates natively in your terminal, providing seamless integration with your existing development environment.

Forge is developed by [Antinomy](https://antinomy.ai) (formerly ForgeCode) and is designed as a comprehensive coding agent that works with multiple LLM providers including Claude, GPT, Gemini, DeepSeek, Grok, and 300+ models through OpenRouter.

**Official Website:** [https://forgecode.dev](https://forgecode.dev)  
**GitHub Repository:** [https://github.com/antinomyhq/forge](https://github.com/antinomyhq/forge)

---

## Key Features

### Core Capabilities

| Feature | Description |
|---------|-------------|
| **Terminal-Native** | Works directly in your terminal without context switching |
| **Multi-Provider Support** | 300+ models via OpenRouter, plus native Anthropic, OpenAI, Google, and more |
| **Multi-Agent System** | Built-in agents (Forge, Sage, Muse) for different tasks |
| **Custom Agents** | Create your own agents with specific capabilities and tools |
| **MCP Integration** | Model Context Protocol support for external tools |
| **Zsh Plugin** | Shell integration with `: ` trigger for quick prompts |
| **Restricted Mode** | Secure shell mode with configurable permissions |
| **Semantic Search** | AI-powered code search using embeddings |
| **Automatic Context** | AGENTS.md discovery and context loading |

### Advanced Features

- **Workflow Orchestration** - Complex multi-step development workflows
- **Context Compaction** - Automatic context window management
- **Reasoning Mode** - Extended thinking for complex problems
- **Git Integration** - Automated commits, PRs, and conflict resolution
- **Skills System** - Reusable capabilities for agents
- **Shell Command Suggestions** - AI-powered command completion
- **Session Management** - Persistent conversation history

---

## Installation

### Quick Install (Recommended)

```bash
curl -fsSL https://forgecode.dev/cli | sh
```

### Alternative Methods

#### Nix
```bash
nix run github:antinomyhq/forge  # Latest dev branch
```

#### Cargo (from source)
```bash
git clone https://github.com/antinomyhq/forge.git
cd forge
cargo build --release
```

### Prerequisites

- **Nerd Font** - For terminal icons (e.g., FiraCode Nerd Font)
- **Zsh** - Recommended shell (Bash also supported)
- **Rust toolchain** - Only for building from source (1.92+)

---

## Quick Start

### 1. Configure Provider

```bash
# Interactive provider setup
forge provider login

# List available providers
forge provider list
```

### 2. Select Model

```bash
# Interactive model selection
forge model

# Or specify directly
forge --model claude-sonnet-4
```

### 3. Start Using Forge

```bash
# Interactive mode
forge

# With a prompt
forge -p "Explain this codebase"

# Zsh plugin (after setup)
: explain how the auth system works
```

### 4. Setup Zsh Plugin (Optional but Recommended)

```bash
forge zsh-setup
# Then restart your terminal
```

---

## Built-in Agents

### Forge (Implementation Agent)

The default hands-on coding agent for executing development tasks.

```bash
forge
> Refactor the authentication module to use JWT tokens
```

**Capabilities:**
- Read/write/patch files
- Execute shell commands
- Semantic code search
- MCP tool integration

### Sage (Research Agent)

Read-only agent for codebase exploration and analysis.

```bash
forge agent sage
> Analyze the architecture of this project
```

**Capabilities:**
- Deep codebase analysis
- Pattern recognition
- Documentation generation
- Read-only investigation

### Muse (Planning Agent)

Strategic planning agent for creating implementation plans.

```bash
forge agent muse
> Create a plan for adding OAuth2 authentication
```

**Capabilities:**
- Implementation roadmaps
- Risk assessment
- Task breakdown
- Alternative approaches

---

## Repository Structure

```
cli-agents/forge/
├── crates/                      # Rust workspace crates
│   ├── forge_main/             # Main application entry
│   ├── forge_domain/           # Domain types and models
│   ├── forge_api/              # API abstractions
│   ├── forge_app/              # Application logic
│   ├── forge_services/         # Business services
│   ├── forge_infra/            # Infrastructure layer
│   ├── forge_fs/               # File system operations
│   ├── forge_repo/             # Repository management
│   ├── forge_embed/            # Embedding generation
│   ├── forge_display/          # Terminal display
│   ├── forge_config/           # Configuration management
│   ├── forge_template/         # Template engine
│   ├── forge_walker/           # File system walker
│   └── ...                     # Additional crates
├── .forge/                      # Forge configuration
│   ├── agents/                 # Custom agent definitions
│   ├── commands/               # Custom commands
│   └── skills/                 # Reusable skills
├── templates/                   # Prompt templates
├── benchmarks/                  # Performance benchmarks
├── Cargo.toml                  # Workspace manifest
├── forge.default.yaml          # Default configuration
└── forge.schema.json           # JSON Schema for validation
```

---

## Configuration

### Main Config File (`forge.yaml`)

```yaml
# yaml-language-server: $schema=./forge.schema.json
variables:
  operating_agent: Forge
  advanced_model: &advanced_model anthropic/claude-sonnet-4

model: *advanced_model
max_requests_per_turn: 100
max_tool_failure_per_turn: 3
temperature: 0.7
top_p: 0.8
top_k: 30
max_tokens: 20480
max_walker_depth: 1
tool_supported: true

# Context compaction settings
compact:
  max_tokens: 2000
  token_threshold: 100000
  retention_window: 6
  message_threshold: 200
  eviction_window: 0.2
  on_turn_end: false

# Update settings
updates:
  frequency: "daily"
  auto_update: false
```

### Provider Configuration

Stored securely via keychain/keyring after `forge provider login`:

```bash
# Manage providers
forge provider login    # Add/update credentials
forge provider logout   # Remove credentials
forge provider list     # List supported providers
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `FORGE_LOG` | Log level (error/warn/info/debug/trace) | `forge=info` |
| `FORGE_TRACKER` | Enable tracking telemetry | `true` |
| `FORGE_RETRY_MAX_ATTEMPTS` | Max retry attempts | `3` |
| `FORGE_TOOL_TIMEOUT` | Tool execution timeout (seconds) | `300` |
| `FORGE_HTTP_READ_TIMEOUT` | HTTP read timeout (seconds) | `900` |

---

## Documentation Links

- [Architecture Documentation](./ARCHITECTURE.md) - System design and Rust components
- [API Reference](./API.md) - Commands, settings, and configuration
- [Usage Guide](./USAGE.md) - Workflows, examples, and best practices
- [External References](./REFERENCES.md) - Tutorials and community resources
- [Visual Diagrams](./DIAGRAMS.md) - Architecture diagrams
- [Gap Analysis](./GAP_ANALYSIS.md) - Improvement opportunities

---

## Version Information

- **Current Version:** See [GitHub releases](https://github.com/antinomyhq/forge/releases)
- **Rust Edition:** 2024
- **Minimum Rust Version:** 1.92

---

## Support & Community

| Channel | Link |
|---------|------|
| **Discord** | [https://discord.gg/kRZBPpkgwq](https://discord.gg/kRZBPpkgwq) |
| **GitHub Issues** | [https://github.com/antinomyhq/forge/issues](https://github.com/antinomyhq/forge/issues) |
| **Documentation** | [https://forgecode.dev/docs](https://forgecode.dev/docs) |

---

## License

Forge is open-source software licensed under the [Apache-2.0 License](../../../cli-agents/forge/LICENSE).

---

*Part of the HelixAgent CLI Agent Collection*
