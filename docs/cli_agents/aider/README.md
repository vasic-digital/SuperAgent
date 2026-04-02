# Aider

## Overview

**Aider** is an AI pair programming tool that operates directly in your terminal. It enables you to edit code in your local git repository using natural language commands, integrating seamlessly with multiple LLM providers to provide intelligent code assistance.

**Official Website:** [https://aider.chat](https://aider.chat)

---

## Key Features

### Core Capabilities

| Feature | Description |
|---------|-------------|
| **Multi-LLM Support** | Works with 20+ LLM providers including Claude, GPT-4o, DeepSeek, Gemini |
| **Git Integration** | Automatic commits with sensible messages, tracks all AI changes |
| **Repository Mapping** | Creates codebase map for better context in large projects |
| **100+ Languages** | Python, JavaScript, Rust, Go, Java, C++, and dozens more |
| **Multiple Edit Formats** | diff, whole file, udiff, editor, architect modes |
| **Voice-to-Code** | Speak your requests, Aider implements them |
| **Web Page/Images** | Add URLs and images as context for the LLM |
| **Linting & Testing** | Auto-run linters and tests on AI-generated code |

### Advanced Features

- **Architect Mode**: Design with one model, edit with another
- **Context Mode**: Automatically identify files that need editing
- **Watch Mode**: IDE integration via file watching
- **Copy/Paste Mode**: Work with web-based LLMs
- **Prompt Caching**: Reduce API costs with cached prompts
- **Reasoning Support**: Works with reasoning models (o1, DeepSeek R1)

---

## Installation

### Recommended: aider-install

```bash
# Install using the official installer (Mac & Linux)
curl -LsSf https://aider.chat/install.sh | sh

# Or with wget
wget -qO- https://aider.chat/install.sh | sh

# Windows
powershell -ExecutionPolicy ByPass -c "irm https://aider.chat/install.ps1 | iex"
```

### Alternative: uv (Recommended)

```bash
python -m pip install uv
uv tool install --force --python python3.12 --with pip aider-chat@latest
```

### Alternative: pipx

```bash
python -m pip install pipx
pipx install aider-chat
```

### Requirements

- **Python**: 3.10 - 3.14 (3.12 recommended)
- **Git**: Required for repository tracking
- **API Key**: From at least one LLM provider

---

## Quick Start

### 1. Initial Setup

```bash
# Navigate to your project
cd /path/to/your/project

# Start with a specific model
aider --model sonnet --api-key anthropic=your_key

# Or use DeepSeek
aider --model deepseek --api-key deepseek=your_key

# Or OpenAI
aider --model gpt-4o --api-key openai=your_key
```

### 2. Basic Usage

```bash
# Start with specific files
aider src/main.py src/utils.py

# Aider will:
# - Load the files into context
# - Show you a prompt (>) where you can ask for changes
# - Edit files based on your requests
# - Commit changes with descriptive messages
```

### 3. Example Session

```bash
$ aider factorial.py

Aider v0.78.0
Models: claude-3-7-sonnet with diff edit format
Git repo: .git with 258 files
Repo-map: using 1024 tokens
Use /help to see in-chat commands
────────────────────────────────────────────────────────
> Make a program that asks for a number and prints its factorial

[Claude will implement the code and show you the diff]

[Changes will be automatically committed]
```

---

## Repository Structure

This repository (`cli_agents/aider/`) contains:

```
aider/
├── aider/                      # Main Python package
│   ├── coders/                 # Edit format implementations
│   │   ├── base_coder.py       # Base coder class
│   │   ├── editblock_coder.py  # Search/replace edit format
│   │   ├── wholefile_coder.py  # Whole file edit format
│   │   ├── udiff_coder.py      # Unified diff format
│   │   ├── architect_coder.py  # Architect/editor mode
│   │   └── ...                 # Additional coders
│   ├── commands.py             # In-chat slash commands
│   ├── main.py                 # Entry point
│   ├── args.py                 # CLI argument parsing
│   ├── models.py               # LLM model definitions
│   ├── repomap.py              # Repository mapping
│   ├── repo.py                 # Git integration
│   ├── io.py                   # Input/output handling
│   ├── llm.py                  # LLM communication
│   ├── voice.py                # Voice input
│   ├── linter.py               # Linting support
│   └── watch.py                # File watching
├── benchmark/                  # Benchmarking tools
├── tests/                      # Test suite
├── website/                    # Documentation website
│   └── docs/                   # Markdown documentation
├── requirements.txt            # Dependencies
└── pyproject.toml              # Package configuration
```

---

## LLM Support

### Best Performing Models

| Model | Provider | Notes |
|-------|----------|-------|
| Claude 3.7 Sonnet | Anthropic | Excellent code editing |
| GPT-4o / o3-mini | OpenAI | Strong reasoning |
| DeepSeek R1/V3 | DeepSeek | Open source, capable |
| Gemini 2.5 Pro | Google | Large context window |

### Quick Model Selection

```bash
# Claude
aider --model sonnet --api-key anthropic=sk-...

# OpenAI
aider --model gpt-4o --api-key openai=sk-...
aider --model o3-mini --api-key openai=sk-...

# DeepSeek
aider --model deepseek --api-key deepseek=sk-...

# Gemini
aider --model gemini-2.5-pro --api-key gemini=...

# Local (Ollama)
aider --model ollama/llama3.2
```

---

## Documentation Links

- [Architecture Documentation](./ARCHITECTURE.md) - System design and components
- [API Reference](./API.md) - Commands and configuration
- [Usage Guide](./USAGE.md) - Tutorials and examples
- [External References](./REFERENCES.md) - Links and resources

---

## Version Information

- **Current Version**: 0.78.0+
- **Python Support**: 3.10 - 3.14
- **License**: Apache License 2.0
- **Release Frequency**: Regular updates with detailed changelogs

---

## Support & Community

- **GitHub**: [Aider-AI/aider](https://github.com/Aider-AI/aider)
- **Discord**: [Aider Discord](https://discord.gg/Y7X7bhMQFV)
- **Issues**: [GitHub Issues](https://github.com/Aider-AI/aider/issues)
- **Documentation**: [aider.chat/docs](https://aider.chat/docs)

---

## Data & Privacy

Aider collects optional analytics:
- Code acceptance/rejection feedback
- Model usage statistics
- Error reports

**Privacy Features:**
- All data collection is opt-in
- Local processing where possible
- API keys stored in `.env` files (not sent anywhere)

See [Privacy Policy](https://aider.chat/docs/legal/privacy.html)

---

## License

Apache License 2.0 - See [LICENSE.txt](../../../cli_agents/aider/LICENSE.txt)

---

*Part of the HelixAgent CLI Agent Collection*
