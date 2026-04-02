# OpenAI Codex CLI

## Overview

**OpenAI Codex CLI** is an open-source coding agent from OpenAI that runs locally in your terminal. It combines ChatGPT-level reasoning with the ability to execute code, manipulate files, and iterate on solutions - all under version control.

**Official Documentation:** https://developers.openai.com/codex

---

## Key Features

### Core Capabilities

| Feature | Description |
|---------|-------------|
| **Terminal-Native** | Built for developers who live in the terminal |
| **Code Execution** | Runs code in sandboxed environments |
| **File Manipulation** | Reads, writes, and patches files |
| **Multimodal** | Accepts screenshots and diagrams for implementation |
| **Version Control** | Works with Git for safe experimentation |
| **Multi-Provider** | Supports OpenAI, Azure, Gemini, Ollama, and more |
| **Sandboxed Execution** | Network-disabled, directory-confined for security |
| **Fully Open Source** | Apache 2.0 license - see and contribute to development |

### Implementation Languages

Codex has two implementations:

| Implementation | Language | Status | Location |
|----------------|----------|--------|----------|
| **Legacy** | TypeScript | Deprecated | `codex-cli/` |
| **Current** | Rust | Active | `codex-rs/` |

---

## Installation

### Via Package Manager (Recommended)

```bash
# npm
npm install -g @openai/codex

# Homebrew
brew install --cask codex

# yarn
yarn global add @openai/codex

# pnpm
pnpm add -g @openai/codex

# bun
bun install -g @openai/codex
```

### Via GitHub Release

Download prebuilt binaries:

| Platform | Architecture | File |
|----------|--------------|------|
| macOS | Apple Silicon | `codex-aarch64-apple-darwin.tar.gz` |
| macOS | Intel | `codex-x86_64-apple-darwin.tar.gz` |
| Linux | x86_64 | `codex-x86_64-unknown-linux-musl.tar.gz` |
| Linux | arm64 | `codex-aarch64-unknown-linux-musl.tar.gz` |

### Build from Source

```bash
# Clone repository
git clone https://github.com/openai/codex.git
cd codex

# Build Rust implementation
cd codex-rs
cargo build --release

# Or use Bazel (repository-wide build)
bazel build //...
```

---

## Quick Start

### 1. Authentication

```bash
# Option 1: Sign in with ChatGPT (recommended)
codex
# Select "Sign in with ChatGPT" in the interactive prompt

# Option 2: API Key
export OPENAI_API_KEY="your-api-key-here"
codex
```

### 2. Interactive Mode

```bash
# Start interactive REPL
codex

# With initial prompt
codex "explain this codebase to me"
```

### 3. Approval Modes

| Mode | File Operations | Shell Commands | Network |
|------|-----------------|----------------|---------|
| **Suggest** (default) | Read-only | None | N/A |
| **Auto Edit** | Read/Write | Requires approval | Disabled |
| **Full Auto** | Read/Write | Auto-execute | Disabled |

```bash
# Full auto mode
codex --approval-mode full-auto "create a todo app"

# Auto edit mode
codex --approval-mode auto-edit "refactor utils.py"
```

### 4. Non-Interactive / CI Mode

```bash
# Quiet mode for CI/CD
codex -q --json "update CHANGELOG for release"

# With specific approval mode
codex -a auto-edit --quiet "run tests and fix failures"
```

---

## Repository Structure

```
codex/
├── codex-cli/              # Legacy TypeScript implementation
│   ├── src/               # Source code
│   ├── tests/             # Test suite
│   ├── package.json       # npm configuration
│   └── README.md          # Legacy docs
│
├── codex-rs/              # Current Rust implementation
│   ├── core/              # Core library
│   ├── tui/               # Terminal UI
│   ├── app-server/        # App server protocol
│   ├── Cargo.toml         # Rust workspace
│   └── README.md          # Rust-specific docs
│
├── .codex/                # Codex configuration
│   └── skills/            # Built-in skills
│
├── docs/                  # Documentation
│   ├── getting-started.md
│   ├── authentication.md
│   ├── contributing.md
│   └── ...
│
├── .github/               # GitHub workflows
├── MODULE.bazel           # Bazel module configuration
├── justfile               # Just command runner
└── README.md              # This file
```

---

## Security Model

### Sandboxing by Platform

| Platform | Mechanism | Features |
|----------|-----------|----------|
| **macOS 12+** | Apple Seatbelt (`sandbox-exec`) | Read-only jail, writable roots, blocked outbound network |
| **Linux** | Docker container | Minimal image, custom iptables firewall, OpenAI API only |
| **Windows** | WSL2 required | Uses Linux sandboxing via WSL2 |

### Approval Mode Matrix

```
┌─────────────────┬─────────────┬─────────────────┬─────────┐
│ Mode            │ File Read   │ File Write      │ Shell   │
├─────────────────┼─────────────┼─────────────────┼─────────┤
│ Suggest         │ Auto        │ Prompt          │ Prompt  │
│ Auto Edit       │ Auto        │ Auto            │ Prompt  │
│ Full Auto       │ Auto        │ Auto            │ Auto*   │
└─────────────────┴─────────────┴─────────────────┴─────────┘
* Network disabled, directory-confined
```

---

## Multi-Provider Support

Codex supports multiple AI providers via the OpenAI-compatible API:

| Provider | Environment Variable | Base URL |
|----------|---------------------|----------|
| **OpenAI** (default) | `OPENAI_API_KEY` | https://api.openai.com/v1 |
| **Azure** | `AZURE_OPENAI_API_KEY` | https://YOUR_PROJECT.openai.azure.com/openai |
| **OpenRouter** | `OPENROUTER_API_KEY` | https://openrouter.ai/api/v1 |
| **Gemini** | `GEMINI_API_KEY` | https://generativelanguage.googleapis.com/v1beta/openai |
| **Ollama** | `OLLAMA_API_KEY` | http://localhost:11434/v1 |
| **Mistral** | `MISTRAL_API_KEY` | https://api.mistral.ai/v1 |
| **DeepSeek** | `DEEPSEEK_API_KEY` | https://api.deepseek.com |
| **xAI** | `XAI_API_KEY` | https://api.x.ai/v1 |
| **Groq** | `GROQ_API_KEY` | https://api.groq.com/openai/v1 |
| **ArceeAI** | `ARCEEAI_API_KEY` | https://conductor.arcee.ai/v1 |

```bash
# Use specific provider
codex --provider gemini "explain this code"

# With custom model
codex --provider openai --model gpt-4.1 "refactor this"
```

---

## Configuration

### Configuration File

Location: `~/.codex/config.yaml` or `~/.codex/config.json`

```yaml
# Basic configuration
model: o4-mini
approvalMode: suggest
fullAutoErrorMode: ask-user
notify: true

# Custom providers
providers:
  openai:
    name: OpenAI
    baseURL: https://api.openai.com/v1
    envKey: OPENAI_API_KEY
  
  azure:
    name: Azure OpenAI
    baseURL: https://myproject.openai.azure.com/openai
    envKey: AZURE_OPENAI_API_KEY

# History settings
history:
  maxSize: 1000
  saveHistory: true
  sensitivePatterns: []
```

### Project Documentation (AGENTS.md)

Codex loads `AGENTS.md` files for context:

```
~/.codex/AGENTS.md          # Personal global guidance
./AGENTS.md                 # Project-level notes
./subdir/AGENTS.md          # Subdirectory-specific
```

Example `AGENTS.md`:
```markdown
# Project Guidelines

## Coding Standards
- Use TypeScript strict mode
- Prefer functional components
- Add tests for all new features

## Commands
- Build: `npm run build`
- Test: `npm test`
- Lint: `npm run lint`
```

---

## Documentation

- [Architecture](./ARCHITECTURE.md) - System design and components
- [API Reference](./API.md) - Commands and configuration
- [Usage Guide](./USAGE.md) - Workflows and examples
- [Development Guide](./DEVELOPMENT.md) - Contributing guidelines
- [External References](./REFERENCES.md) - Resources and tutorials
- [Diagrams](./DIAGRAMS.md) - Visual documentation

---

## Related Projects

| Project | Description | Link |
|---------|-------------|------|
| **Codex Web** | Cloud-based agent | https://chatgpt.com/codex |
| **Codex IDE** | VS Code/Cursor/Windsurf extension | https://developers.openai.com/codex/ide |
| **Codex App** | Desktop application | `codex app` or https://chatgpt.com/codex |

---

## License

Apache 2.0 - See [LICENSE](../../../cli_agents/codex/LICENSE)

---

*Part of the HelixAgent CLI Agent Collection*
