# GPT-Engineer User Guide

## Table of Contents
1. [Installation](#installation)
2. [Quick Start](#quick-start)
3. [CLI Commands](#cli-commands)
4. [Configuration](#configuration)
5. [Usage Examples](#usage-examples)
6. [Troubleshooting](#troubleshooting)

## Installation

### Method 1: pip (Recommended)
```bash
pip install gpt-engineer
```

### Method 2: pipx
```bash
pipx install gpt-engineer
```

### Method 3: Homebrew
```bash
brew install gpt-engineer
```

### Method 4: Source
```bash
git clone https://github.com/gpt-engineer-org/gpt-engineer.git
cd gpt-engineer
pip install -e .
```

## Quick Start

```bash
# Create new project
gpt-engineer

# Or with prompt directly
gpt-engineer "Create a Python CLI tool that converts JSON to CSV"

# Improve existing code
gpt-engineer --improve
```

## CLI Commands

### Global Options
| Option | Short | Description | Example |
|--------|-------|-------------|---------|
| --help | -h | Show help | `gpt-engineer --help` |
| --version | -v | Show version | `gpt-engineer --version` |
| --model | -m | Model to use | `--model gpt-4` |
| --temperature | | Temperature | `--temperature 0.7` |
| --verbose | | Verbose output | `--verbose` |

### Command: (default)
**Description:** Create new project from prompt

**Usage:**
```bash
gpt-engineer "Your project description"
gpt-engineer -m claude-3-opus "Description"
```

### Command: --improve
**Description:** Improve existing code

**Usage:**
```bash
gpt-engineer --improve
gpt-engineer --improve -m gpt-4
```

### Command: --lite
**Description:** Lite mode (faster, cheaper)

**Usage:**
```bash
gpt-engineer --lite "Quick feature"
```

### Command: --clarify
**Description:** Clarify mode (asks questions)

**Usage:**
```bash
gpt-engineer --clarify "Build a web app"
# Will ask clarifying questions
```

### Command: --self-heal
**Description:** Self-healing mode (fixes errors)

**Usage:**
```bash
gpt-engineer --self-heal "Create API"
# Automatically fixes issues
```

## Configuration

### Configuration File Format

GPT-Engineer uses prompt files in project directory:

```
project/
├── prompt          # Main prompt
├── .env            # Environment variables
└── preprompts/     # Optional preprompts
    ├── roadmapper  # Architecture planning
    ├── philosopher # Clarification
    └── coder       # Implementation
```

### prompt file
```markdown
# Project Description

We are building a web scraper that:
- Accepts a URL as input
- Extracts article content
- Saves to markdown
- Handles pagination

## Technical Requirements
- Python 3.11+
- Use requests and BeautifulSoup
- Include error handling
- Add CLI interface
```

### .env file
```bash
OPENAI_API_KEY=your-key
MODEL_NAME=gpt-4
TEMPERATURE=0.7
```

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| OPENAI_API_KEY | OpenAI API key | Yes |
| ANTHROPIC_API_KEY | Anthropic key | Alternative |
| MODEL_NAME | Model to use | No |
| TEMPERATURE | Temperature | No |

## Usage Examples

### Example 1: New Project
```bash
mkdir my-api && cd my-api
echo "Create a FastAPI REST API with CRUD for users" > prompt
gpt-engineer
```

### Example 2: With Specific Model
```bash
gpt-engineer -m claude-3-opus "Build a React dashboard"
```

### Example 3: Improve Existing Code
```bash
cd existing-project
gpt-engineer --improve
# Asks what to improve
```

### Example 4: Clarify Mode
```bash
gpt-engineer --clarify "Build an e-commerce site"
# Answers questions about features
# Then generates code
```

### Example 5: Self-Heal Mode
```bash
gpt-engineer --self-heal "Create a compiler"
# If build fails, automatically fixes
```

### Example 6: Lite Mode for Quick Tasks
```bash
gpt-engineer --lite "Add logging to main.py"
```

## Troubleshooting

### Issue: API Key Not Found
**Solution:**
```bash
export OPENAI_API_KEY=your-key
# Or create .env file
```

### Issue: Model Not Available
**Solution:**
```bash
# Check available models
gpt-engineer --help | grep model

# Use valid model
gpt-engineer -m gpt-4 "Task"
```

### Issue: Generated Code Doesn't Run
**Solution:**
```bash
# Use self-heal mode
gpt-engineer --self-heal "Original prompt"

# Or run with --improve
gpt-engineer --improve
```

### Issue: Prompt Too Long
**Solution:**
- Break into smaller prompts
- Use preprompts directory
- Focus on core requirements first

---

**Last Updated:** 2026-04-02
