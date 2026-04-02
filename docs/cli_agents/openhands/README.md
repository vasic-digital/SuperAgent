# OpenHands

## Overview

**OpenHands** is an AI-powered software engineering agent developed by [All-Hands-AI](https://www.all-hands.dev/). It performs real development tasks including coding, debugging, testing, and git operations through natural language commands. OpenHands can interact with files, execute commands, browse the web, and integrate with various development workflows.

**Official Documentation:** [https://docs.openhands.dev](https://docs.openhands.dev)

---

## Key Features

### Core Capabilities

| Feature | Description |
|---------|-------------|
| **CodeAct Framework** | Unified code action space for AI agents (bash, Python, browser, file editing) |
| **Multi-Runtime Support** | Docker, Kubernetes, Modal, Runloop, Remote, and Local runtimes |
| **Multi-Agent System** | Specialized agents: CodeAct, Browsing, VisualBrowsing, RepoExplorer |
| **Sandboxed Execution** | Secure containerized environment for code execution |
| **Web Browsing** | Built-in browser automation for web-based tasks |
| **MCP Support** | Model Context Protocol for external tool integration |
| **Microagents** | Specialized prompts for domain-specific tasks |
| **IDE Integration** | VSCode extension available |

### Advanced Features

- **Runtime Plugins**: Jupyter integration, agent skills, file editing
- **Memory System**: Persistent conversation and repository context
- **Condensers**: Intelligent context compression (LLM-based, amortized forgetting)
- **Security Analysis**: Optional LLM-based or invariant security analyzers
- **Git Integration**: Automatic repository cloning and branch management
- **Multi-Modal**: Support for vision-capable models

---

## Installation

### Prerequisites

- **Operating System**: Linux, macOS, or WSL on Windows
- **Docker**: Required for sandboxed execution
- **Python**: 3.12+ (for development)
- **Node.js**: 22.x+ (for frontend/UI)
- **Poetry**: 1.8+ (for Python dependencies)

### Method 1: Docker (Recommended)

```bash
# Clone the repository
git clone https://github.com/OpenHands/OpenHands.git
cd OpenHands

# Build and run with Docker Compose
docker-compose up -d
```

### Method 2: Local Development Setup

```bash
# Clone the repository
git clone https://github.com/OpenHands/OpenHands.git
cd OpenHands

# Build the project (installs all dependencies)
make build

# Run the full application (backend + frontend)
make run
```

### Method 3: PyPI Installation

```bash
# Install from PyPI
pip install openhands-ai

# Run the CLI
openhands -t "your task here"
```

---

## Quick Start

### 1. Initial Setup

```bash
# Configure LLM settings
make setup-config

# Or set environment variables
export LLM_API_KEY="your-api-key"
export LLM_MODEL="gpt-4o"
```

### 2. Running OpenHands

**Web UI (Default):**
```bash
# Start the server
make run

# Access the UI at http://localhost:3001
```

**Headless/CLI Mode:**
```bash
# Run a single task
python -m openhands.core.main -t "Write a hello world program in Python"

# With custom configuration
python -m openhands.core.main -t "Fix the bug" -c config.toml
```

### 3. Docker Quick Run

```bash
# Run with Docker
export WORKSPACE_BASE=$(pwd)/workspace
docker run -it --rm \
  -p 3000:3000 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v ~/.openhands:/.openhands \
  -v $WORKSPACE_BASE:/opt/workspace_base \
  -e LLM_API_KEY="your-key" \
  -e LLM_MODEL="gpt-4o" \
  openhands:latest
```

---

## Repository Structure

This repository (`cli_agents/openhands/`) contains:

```
openhands/
├── openhands/                # Core Python backend
│   ├── agenthub/            # Agent implementations
│   │   ├── codeact_agent/   # Main CodeAct agent
│   │   ├── browsing_agent/  # Web browsing agent
│   │   └── visualbrowsing_agent/  # Vision-enabled browsing
│   ├── core/                # Core configuration and main loop
│   ├── events/              # Event system (actions/observations)
│   ├── runtime/             # Runtime implementations
│   │   ├── impl/docker/     # Docker runtime
│   │   ├── impl/remote/     # Remote runtime
│   │   └── impl/kubernetes/ # Kubernetes runtime
│   ├── server/              # WebSocket API server
│   ├── memory/              # Memory and context management
│   ├── llm/                 # LLM integration (via LiteLLM)
│   ├── security/            # Security analysis
│   └── mcp/                 # MCP protocol support
├── frontend/                 # React frontend application
├── enterprise/               # Enterprise features (separate license)
├── containers/               # Docker configurations
│   ├── app/                 # Production Docker image
│   ├── dev/                 # Development environment
│   └── runtime/             # Runtime base images
├── skills/                   # Agent skills
├── microagents/              # Public microagents
├── config.template.toml      # Configuration template
├── docker-compose.yml        # Docker Compose setup
├── Makefile                  # Build automation
└── pyproject.toml           # Python dependencies
```

---

## Usage Modes

### 1. Web GUI Mode

Full-featured web interface with:
- Interactive chat interface
- File browser and editor
- Terminal access
- Settings management
- Conversation history

```bash
make run  # Starts on http://localhost:3001
```

### 2. Headless/CLI Mode

For automation and CI/CD pipelines:

```bash
# Single task execution
python -m openhands.core.main -t "Create a Python script"

# With file input
python -m openhands.core.main -f task.txt

# Disable auto-continue (interactive mode)
python -m openhands.core.main -t "Help me" --no-auto-continue
```

### 3. OpenHands Cloud

Managed cloud service available at [app.all-hands.dev](https://app.all-hands.dev):
- Free tier with Minimax model
- GitHub/GitLab integration
- Multi-user support
- Enterprise features

### 4. OpenHands Enterprise

Self-hosted enterprise deployment:
- VPC deployment via Kubernetes
- SSO and RBAC
- Slack, Jira, Linear integrations
- Advanced analytics

---

## Documentation Links

- [Architecture Documentation](./ARCHITECTURE.md) - System design and components
- [API Reference](./API.md) - Commands and configuration
- [Usage Guide](./USAGE.md) - Tutorials and examples
- [References](./REFERENCES.md) - External links and resources

---

## Version Information

- **Current Version**: 1.4.0
- **Python Support**: 3.12, 3.13
- **License**: MIT (core), Polyform Free Trial (enterprise)

---

## Support & Community

- **Slack**: [Join OpenHands Slack](https://dub.sh/openhands)
- **GitHub Issues**: [OpenHands Issues](https://github.com/OpenHands/OpenHands/issues)
- **Documentation**: [docs.openhands.dev](https://docs.openhands.dev)

---

## SWE-bench Performance

OpenHands achieves state-of-the-art results on [SWE-bench](https://www.swebench.com/) (software engineering benchmark):

| Benchmark | Score |
|-----------|-------|
| SWE-bench Verified | 77.6% |
| SWE-bench Lite | 64.0% |

---

## License

- **Core**: MIT License (see [LICENSE](../../../cli_agents/openhands/LICENSE))
- **Enterprise**: Polyform Free Trial License (30-day limit, see [enterprise/LICENSE](../../../cli_agents/openhands/enterprise/LICENSE))

---

*Part of the HelixAgent CLI Agent Collection*
