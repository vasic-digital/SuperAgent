# GPT-Engineer Documentation

Complete documentation for [GPT-Engineer](https://github.com/gpt-engineer-org/gpt-engineer) - the open-source AI code generation platform.

## Overview

GPT-Engineer is a Python-based CLI tool that generates complete codebases from natural language descriptions. It uses Large Language Models (LLMs) to automate software development workflows.

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Natural       │────►│  GPT-Engineer   │────►│   Generated     │
│   Language      │     │  (Python CLI)   │     │   Codebase      │
│   Prompt        │     │                 │     │   (Files)       │
└─────────────────┘     └─────────────────┘     └─────────────────┘
```

## Documentation Structure

| Document | Description | Lines |
|----------|-------------|-------|
| [ARCHITECTURE.md](./ARCHITECTURE.md) | System design, components, data flow | ~400 |
| [API.md](./API.md) | Commands, settings, configuration reference | ~300 |
| [USAGE.md](./USAGE.md) | Workflows, examples, best practices | ~350 |
| [REFERENCES.md](./REFERENCES.md) | External links, tutorials, community | ~280 |
| [DIAGRAMS.md](./DIAGRAMS.md) | Visual documentation, architecture diagrams | ~600 |
| [GAP_ANALYSIS.md](./GAP_ANALYSIS.md) | Improvement opportunities, feature gaps | ~320 |

## Quick Start

### Installation

```bash
# Stable release
pip install gpt-engineer

# Development
git clone https://github.com/gpt-engineer-org/gpt-engineer.git
cd gpt-engineer
poetry install
poetry shell
```

### Basic Usage

```bash
# Create project
echo "Create a Python script that prints 'Hello World'" > prompt

# Generate code
gpte .

# Check output
ls workspace/
```

### Improve Existing Code

```bash
# Add improvement instructions
echo "Add error handling and logging" > prompt

# Run in improve mode
gpte . -i
```

## Key Features

- **Natural Language to Code**: Describe what you want in plain English
- **Multiple LLM Support**: OpenAI, Anthropic, Azure, local LLMs (llama.cpp)
- **Multiple Modes**: Default, clarify, lite, improve, self-heal
- **Vision Support**: Use images as context for generation
- **Benchmarking**: Built-in APPS and MBPP benchmarks
- **Git Integration**: Automatic staging of changes
- **Customizable**: Override preprompts for custom AI identity

## Project Structure

```
gpt-engineer/
├── gpt_engineer/
│   ├── applications/cli/    # CLI implementation
│   ├── core/                # AI, files, diff processing
│   ├── preprompts/          # AI identity/personality
│   ├── benchmark/           # Benchmarking framework
│   └── tools/               # Custom steps
├── projects/                # Example projects
├── docs/                    # Documentation
└── docker/                  # Docker configuration
```

## Environment Setup

```bash
# Required
export OPENAI_API_KEY="sk-xxxxxxxx"

# Optional
export ANTHROPIC_API_KEY="sk-ant-xxx"
export MODEL_NAME="gpt-4o"
```

## Command Reference

```bash
# Generate new code
gpte <project_path> [model]

# Improve existing code
gpte <project_path> -i

# Clarify mode (ask questions first)
gpte <project_path> -c

# Lite mode (faster, simpler)
gpte <project_path> -l

# Self-heal mode (auto-fix errors)
gpte <project_path> -sh

# Run benchmark
bench <agent_file> [config.toml]
```

## Commercial Evolution

The project maintains two paths:

1. **Open Source** (this repository): Code generation experimentation platform
2. **Managed Service**: [gptengineer.app](https://gptengineer.app) - Opinionated web app builder

For a well-maintained CLI alternative, consider [Aider](https://aider.chat).

## Community

- **GitHub**: https://github.com/gpt-engineer-org/gpt-engineer
- **Discord**: https://discord.gg/8tcDQ89Ej2
- **Documentation**: https://gpt-engineer.readthedocs.io
- **PyPI**: https://pypi.org/project/gpt-engineer

## Requirements

- Python 3.10 - 3.12
- OpenAI API key (or Anthropic/Azure/local LLM)
- Git (recommended)

## License

MIT License - See [TERMS_OF_USE.md](../../../cli_agents/gpt-engineer/TERMS_OF_USE.md) for details.

---

*For detailed information, explore the documentation files listed above.*
