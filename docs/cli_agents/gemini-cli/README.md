# Gemini CLI

## Overview

**Gemini CLI** is Google's official open-source AI agent that brings the power of Gemini directly into your terminal. It provides lightweight access to Google's Gemini models, giving you the most direct path from your prompt to the model.

**Official Documentation:** [https://geminicli.com/docs/](https://geminicli.com/docs/)  
**GitHub Repository:** [https://github.com/google-gemini/gemini-cli](https://github.com/google-gemini/gemini-cli)  
**NPM Package:** [@google/gemini-cli](https://www.npmjs.com/package/@google/gemini-cli)

---

## Key Features

### Core Capabilities

| Feature | Description |
|---------|-------------|
| **Natural Language Interface** | Interact with your codebase using plain English commands |
| **Code Understanding** | Query and edit large codebases with contextual awareness |
| **Multimodal Support** | Generate apps from PDFs, images, or sketches |
| **Built-in Tools** | File operations, shell commands, web fetch, Google Search |
| **MCP Support** | Model Context Protocol for custom tool integrations |
| **Session Management** | Save, resume, and checkpoint conversations |
| **Extensible** | Custom commands, skills, and extensions |
| **Open Source** | Apache 2.0 licensed |

### Advanced Features

- **Gemini 3 Models**: Access to improved reasoning with 1M token context window
- **Google Search Grounding**: Real-time information with search citations
- **Checkpointing**: Save and resume complex sessions
- **Custom Commands**: Create reusable TOML-based slash commands
- **GEMINI.md Files**: Project-specific context and instructions
- **Sandboxing**: Secure containerized execution environments
- **Browser Agent**: Automate web browser tasks (experimental)

---

## Installation

### System Requirements

- **Operating System**: macOS 15+, Windows 11 24H2+, Ubuntu 20.04+
- **Hardware**: 4GB+ RAM (casual), 16GB+ RAM (power usage)
- **Runtime**: Node.js 20.0.0+
- **Shell**: Bash or Zsh

### Installation Methods

#### npm (Recommended)
```bash
npm install -g @google/gemini-cli
```

#### Homebrew (macOS/Linux)
```bash
brew install gemini-cli
```

#### MacPorts (macOS)
```bash
sudo port install gemini-cli
```

#### npx (No Installation)
```bash
npx @google/gemini-cli
```

#### Anaconda (Restricted Environments)
```bash
conda create -y -n gemini_env -c conda-forge nodejs
conda activate gemini_env
npm install -g @google/gemini-cli
```

### Release Channels

| Channel | Install Command | Description |
|---------|-----------------|-------------|
| **Stable** | `@latest` (default) | Production-ready, fully vetted |
| **Preview** | `@preview` | Weekly preview with latest features |
| **Nightly** | `@nightly` | Daily builds from main branch |

```bash
# Install preview version
npm install -g @google/gemini-cli@preview

# Install nightly version
npm install -g @google/gemini-cli@nightly
```

---

## Quick Start

### 1. Initial Setup

```bash
# Navigate to your project
cd /path/to/your/project

# Launch Gemini CLI
gemini
```

### 2. Authentication

Choose your authentication method on first run:

#### Option A: Login with Google (Recommended)
```bash
gemini
# Select "Login with Google" and follow browser flow
```

#### Option B: Gemini API Key
```bash
export GEMINI_API_KEY="your-api-key"
gemini
# Select "Use Gemini API key"
```

#### Option C: Vertex AI (Enterprise)
```bash
export GOOGLE_CLOUD_PROJECT="your-project-id"
export GOOGLE_CLOUD_LOCATION="us-central1"
gcloud auth application-default login
gemini
# Select "Vertex AI"
```

### 3. Basic Usage

```bash
# Start interactive session
gemini

# Non-interactive mode
gemini -p "explain this codebase"

# Resume previous session
gemini -r "latest"

# Use specific model
gemini -m gemini-2.5-flash

# Include multiple directories
gemini --include-directories ../lib,../docs
```

---

## Repository Structure

```
gemini-cli/
├── .gemini/                  # Gemini CLI configuration
│   ├── commands/             # Custom slash commands
│   ├── settings.json         # User settings
│   └── GEMINI.md            # Global context file
├── docs/                     # Documentation
├── packages/                 # Monorepo packages
│   ├── cli/                  # Main CLI package
│   └── a2a-server/           # Agent-to-Agent server
├── integration-tests/        # Integration test suite
├── evals/                    # Evaluation framework
└── schemas/                  # JSON schemas
```

---

## Free Tier Limits

| Plan | Requests/Min | Requests/Day | Models |
|------|--------------|--------------|--------|
| **Google Account** | 60 | 1,000 | Gemini 3, 1M context |
| **Gemini API Key** | Varies | 1,000 | Model selection available |
| **Vertex AI** | Higher limits | Billing-based | Enterprise features |

---

## Documentation Links

- [Architecture Documentation](./ARCHITECTURE.md) - System design and components
- [API Reference](./API.md) - Commands and configuration
- [Usage Guide](./USAGE.md) - Tutorials and examples
- [External References](./REFERENCES.md) - Links and resources

---

## Key Commands Quick Reference

| Command | Purpose |
|---------|---------|
| `gemini` | Start interactive session |
| `gemini -p "query"` | Non-interactive mode |
| `gemini -r "latest"` | Resume session |
| `/help` | Show available commands |
| `/exit` | Quit CLI |
| `/model` | Change model |
| `/memory` | Manage context files |
| `/plan` | Enter plan mode |

---

## Support & Community

- **GitHub Issues**: [Report bugs](https://github.com/google-gemini/gemini-cli/issues)
- **GitHub Discussions**: [Community forum](https://github.com/google-gemini/gemini-cli/discussions)
- **Official Roadmap**: [Project board](https://github.com/orgs/google-gemini/projects/11)
- **Bug Reports**: Use `/bug` command within Gemini CLI

---

## Data & Privacy

- **Usage Statistics**: Optional, can be disabled in settings
- **Code Retention**: Limited retention for service improvement
- **Privacy Policy**: See [Terms & Privacy](https://geminicli.com/docs/resources/tos-privacy)

---

## Contributing

We welcome contributions! See [CONTRIBUTING.md](https://github.com/google-gemini/gemini-cli/blob/main/CONTRIBUTING.md) for:
- Development setup
- Code contribution process
- Documentation contributions
- Pull request guidelines

---

## License

Apache License 2.0 - See [LICENSE](https://github.com/google-gemini/gemini-cli/blob/main/LICENSE)

---

*Part of the HelixAgent CLI Agent Collection*
