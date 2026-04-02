# Amazon Q Developer CLI

## Overview

**Amazon Q Developer CLI** is AWS's AI-powered coding assistant that operates directly in your terminal. It provides intelligent code suggestions, automated refactoring, security analysis, and natural language code generation integrated with AWS services.

> [!IMPORTANT]
> This open source project is no longer being actively maintained and will only receive critical security fixes. Amazon Q Developer CLI is now available as [Kiro CLI](https://kiro.dev/cli/), a closed-source product. For the latest features and updates, please use Kiro CLI.

**Official Documentation:** [https://docs.aws.amazon.com/amazonq/latest/qdeveloper-ug/command-line.html](https://docs.aws.amazon.com/amazonq/latest/qdeveloper-ug/command-line.html)

---

## Key Features

### Core Capabilities

| Feature | Description |
|---------|-------------|
| **Natural Language Interface** | Interact with your codebase using plain English commands |
| **Code Understanding** | Analyzes and comprehends complex codebases contextually |
| **Task Automation** | Executes routine development tasks automatically |
| **AWS Integration** | Native integration with AWS services and CLI |
| **Security Analysis** | Built-in code security scanning and remediation |
| **Multi-Platform** | Supports macOS and Linux |
| **MCP Support** | Model Context Protocol for external tool integration |
| **Knowledge Base** | Persistent semantic search across sessions |

### Advanced Features

- **Custom Agents**: Create specialized agents for different workflows
- **Hooks System**: Custom event handlers for tool execution
- **TODO Management**: Track multi-step tasks across sessions
- **Knowledge Management**: Semantic search with BM25 and embedding-based indexing
- **Code Transformation**: Automated code modernization (Java upgrades, etc.)
- **Test Generation**: AI-powered unit test creation

---

## Installation

### macOS

**Homebrew (Recommended):**
```bash
brew install --cask amazon-q
```

**DMG Download:**
[Download Amazon Q.dmg](https://desktop-release.q.us-east-1.amazonaws.com/latest/Amazon%20Q.dmg)

### Linux

**Ubuntu/Debian:**
```bash
# Download and install from AWS documentation
# See: https://docs.aws.amazon.com/amazonq/latest/qdeveloper-ug/command-line-installing.html
```

**AppImage:**
[Download AppImage](https://docs.aws.amazon.com/amazonq/latest/qdeveloper-ug/command-line-installing.html#command-line-installing-appimage)

### Requirements

- **Operating Systems**: macOS 10.15+, Ubuntu 20.04+, Amazon Linux 2+
- **AWS Account**: Required for authentication
- **Internet Connection**: Required for AI features

---

## Quick Start

### 1. Initial Setup

```bash
# Launch Amazon Q CLI
q chat
```

### 2. Authentication

On first run, Amazon Q will prompt for authentication:
- Follow the browser OAuth flow
- Sign in with your AWS account
- Grant necessary permissions

### 3. Basic Commands

```bash
# Start chat session
q chat

# Start with specific agent
q chat --agent my-custom-agent

# View help
q --help

# Configure settings
q settings chat.defaultAgent my-agent
```

---

## Repository Structure

This repository (`../../cli_agents/amazon-q/`) contains:

```
amazon-q/
├── crates/
│   ├── chat-cli/           # Main CLI application
│   ├── chat-cli-ui/        # Terminal UI components
│   ├── agent/              # Agent runtime and tools
│   ├── amzn-codewhisperer-client/    # AWS API client
│   ├── amzn-codewhisperer-streaming-client/  # Streaming API
│   └── ...                 # Additional service clients
├── docs/                   # Technical documentation
│   ├── agent-format.md     # Agent configuration format
│   ├── built-in-tools.md   # Tool reference
│   ├── hooks.md            # Hooks documentation
│   ├── knowledge-management.md  # Knowledge base docs
│   └── ...
├── scripts/                # Build and ops scripts
├── .github/                # GitHub workflows
├── Cargo.toml             # Rust workspace configuration
├── CONTRIBUTING.md         # Contribution guidelines
├── LICENSE.MIT            # MIT License
├── LICENSE.APACHE         # Apache 2.0 License
└── README.md              # This file
```

---

## Built-in Tools

Amazon Q CLI includes several built-in tools:

| Tool | Purpose | Permission Level |
|------|---------|------------------|
| `execute_bash` | Execute shell commands | Configurable |
| `fs_read` | Read files and directories | Trusted |
| `fs_write` | Create and edit files | Prompt |
| `introspect` | Access Q CLI documentation | Trusted |
| `knowledge` | Store/retrieve from knowledge base | Trusted |
| `report_issue` | Open GitHub issue template | Trusted |
| `thinking` | Internal reasoning mechanism | Trusted |
| `todo_list` | Manage TODO lists | Trusted |
| `use_aws` | Make AWS CLI API calls | Configurable |

---

## Agent System

Amazon Q CLI uses a flexible agent system for customization:

### Default Agent Locations

- **Global**: `~/.aws/amazonq/cli-agents/`
- **Local**: `.amazonq/cli-agents/` (project-specific)

### Creating Custom Agents

Use the `/agent generate` slash command within Q to create agents interactively, or create JSON configuration files manually.

### Agent Selection Priority

1. Command-line specified (`--agent` flag)
2. User-defined default (`chat.defaultAgent` setting)
3. Built-in default agent

---

## Documentation Links

- [Architecture Documentation](./ARCHITECTURE.md) - System design and components
- [API Reference](./API.md) - Commands and configuration
- [Usage Guide](./USAGE.md) - Tutorials and examples
- [External References](./REFERENCES.md) - Links and resources

---

## Development

### Prerequisites

- **Rust**: Latest stable toolchain
- **macOS**: Xcode 13 or later (for macOS development)

### Building from Source

```bash
# Clone repository
git clone https://github.com/aws/amazon-q-developer-cli.git
cd amazon-q-developer-cli

# Install Rust toolchain
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
rustup default stable
rustup toolchain install nightly
cargo install typos-cli

# Build and run
cargo run --bin chat_cli

# Run tests
cargo test

# Run lints
cargo clippy

# Format code
cargo +nightly fmt
```

---

## Data & Privacy

Amazon Q CLI collects usage data including:
- Code acceptance/rejection feedback
- Conversation data
- Feature usage analytics

**Privacy Safeguards:**
- Data handled according to AWS privacy policies
- Configurable data sharing settings
- Enterprise controls available

See [AWS Privacy Policy](https://aws.amazon.com/privacy/) for more information.

---

## License

This repository is dual-licensed under:
- **MIT License**: See [LICENSE.MIT](../../../cli-agents/amazon-q/LICENSE.MIT)
- **Apache 2.0 License**: See [LICENSE.APACHE](../../../cli-agents/amazon-q/LICENSE.APACHE)

"Amazon Web Services" and all related marks are trademarks of AWS.

---

*Part of the HelixAgent CLI Agent Collection*
