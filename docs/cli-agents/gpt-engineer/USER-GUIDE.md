# GPT-Engineer User Guide

## Table of Contents
1. [Installation](#installation)
2. [Quick Start](#quick-start)
3. [CLI Commands](#cli-commands)
4. [TUI/Interactive Commands](#tui-interactive-commands)
5. [Configuration](#configuration)
6. [API/Protocol Endpoints](#api-protocol-endpoints)
7. [Usage Examples](#usage-examples)
8. [Troubleshooting](#troubleshooting)

## Installation

### Method 1: pip (Recommended)

```bash
# Install stable release
pip install gpt-engineer

# Or with specific Python
python -m pip install gpt-engineer

# Verify installation
gpt-engineer --version
```

### Method 2: pipx (Isolated)

```bash
# Install with pipx
pipx install gpt-engineer

# Verify
gpt-engineer --version

# Upgrade
pipx upgrade gpt-engineer
```

### Method 3: Docker

```bash
# Pull Docker image
docker pull paulgauthier/gpt-engineer

# Run with current directory
docker run -it -v $(pwd):/project paulgauthier/gpt-engineer
```

### Method 4: Build from Source

```bash
# Clone repository
git clone https://github.com/gpt-engineer-org/gpt-engineer.git
cd gpt-engineer

# Install with Poetry
pip install poetry
poetry install
poetry shell

# Or with pip
pip install -e .
```

### Method 5: Using aider-install

```bash
# Install using aider installer
python -m pip install aider-install

# Then use gpt-engineer
```

### Prerequisites

- Python 3.10 - 3.12
- Git
- API key from OpenAI, Anthropic, or other supported provider

## Quick Start

### First-Time Setup

```bash
# Verify installation
gpt-engineer --version

# Set up API key
export OPENAI_API_KEY="sk-..."
# OR
export ANTHROPIC_API_KEY="sk-ant-..."

# Create a new project
mkdir my-project
cd my-project

# Create prompt file
echo "Create a simple REST API in Python using Flask" > prompt

# Run gpt-engineer
gpt-engineer .
```

### Basic Usage

```bash
# Create new code (default)
gpt-engineer <project_dir>

# Improve existing code
gpt-engineer <project_dir> -i

# With specific model
gpt-engineer . --model gpt-4o

# List available steps
gpt-engineer --list-steps
```

### Hello World

```bash
# Create project directory
mkdir hello-world
cd hello-world

# Create prompt
echo "Create a Python script that prints Hello World" > prompt

# Run gpt-engineer
gpt-engineer .

# Check generated code in workspace/
```

## CLI Commands

### Global Options

| Option | Short | Description | Example |
|--------|-------|-------------|---------|
| --version | -v | Show version | `gpt-engineer --version` |
| --help | -h | Show help | `gpt-engineer --help` |
| --model | | Select model | `gpt-engineer . --model gpt-4o` |
| --temperature | | Set temperature | `gpt-engineer . --temperature 0.7` |
| --steps | | Select steps | `gpt-engineer . --steps use_feedback` |
| --list-steps | | List available steps | `gpt-engineer --list-steps` |
| --verbose | | Verbose output | `gpt-engineer . --verbose` |
| --no-execution | | Don't execute code | `gpt-engineer . --no-execution` |
| --port | | Port for web UI | `gpt-engineer . --port 8080` |

### Command: (default - new project)

**Description:** Generate new code based on prompt file.

**Usage:**
```bash
gpt-engineer <project_directory> [OPTIONS]
```

**Options:**
| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| --model | string | No | gpt-4o | Model to use |
| --temperature | float | No | 0.1 | Temperature (0-2) |
| --steps | string | No | default | Steps to run |

**Examples:**
```bash
# Generate new code
gpt-engineer my-project

# With specific model
gpt-engineer my-project --model claude-sonnet-4

# With custom temperature
gpt-engineer my-project --temperature 0.5
```

**Exit Codes:**
| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid arguments |

### Command: -i / --improve

**Description:** Improve existing code.

**Usage:**
```bash
gpt-engineer <project_directory> -i [OPTIONS]
gpt-engineer <project_directory> --improve [OPTIONS]
```

**Examples:**
```bash
# Improve existing code
gpt-engineer my-project -i

# With feedback
gpt-engineer my-project --steps use_feedback
```

### Command: --list-steps

**Description:** List available execution steps.

**Usage:**
```bash
gpt-engineer --list-steps
```

**Available Steps:**
| Step | Description |
|------|-------------|
| default | Run default generation pipeline |
| clarify | Ask clarifying questions |
| gen_clarified_code | Generate code from clarified requirements |
| gen_entrypoint | Generate entrypoint script |
| execute_entrypoint | Execute the entrypoint |
| use_feedback | Use feedback to improve code |
| fix_code | Fix code based on errors |

### Command: --self-heal

**Description:** Enable self-healing mode for error correction.

**Usage:**
```bash
gpt-engineer <project_directory> --self-heal
```

**Examples:**
```bash
# Enable self-healing
gpt-engineer my-project --self-heal
```

## TUI/Interactive Commands

During generation, GPT-Engineer may ask clarifying questions:

### Interactive Prompts

| Prompt | Description |
|--------|-------------|
| Clarifying questions | Answer questions about requirements |
| Feedback | Provide feedback on generated code |
| Continue | Press Enter to continue |

### Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| Ctrl+C | Cancel generation |
| Enter | Continue / Submit answer |
| Up/Down | Navigate options |

## Configuration

### Configuration File Format

GPT-Engineer uses environment variables and optional `.env` file:

```bash
# .env file
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...
MODEL=gpt-4o
TEMPERATURE=0.1
```

### Project Structure

```
my-project/
├── prompt              # Main prompt file (required)
├── .env               # Environment variables (optional)
├── identity/          # Identity files (optional)
│   └── setup.py       # Setup identity
├── memory/            # Conversation memory
│   └── log.json       # Conversation log
├── workspace/         # Generated code output
│   └── main.py        # Generated files
└── .gpteng/           # GPT-Engineer internal files
```

### Environment Variables

| Variable | Description | Required | Example |
|----------|-------------|----------|---------|
| OPENAI_API_KEY | OpenAI API key | Yes* | `sk-...` |
| ANTHROPIC_API_KEY | Anthropic API key | Yes* | `sk-ant-...` |
| MODEL | Default model | No | `gpt-4o` |
| TEMPERATURE | Generation temperature | No | `0.1` |
| MAX_TOKENS | Max tokens per request | No | `4000` |

*At least one provider required

### Prompt File Format

Create a `prompt` file in your project directory:

```markdown
# Simple Python API

Create a REST API using Flask with the following endpoints:

- GET /users - List all users
- POST /users - Create a new user
- GET /users/<id> - Get user by ID
- PUT /users/<id> - Update user
- DELETE /users/<id> - Delete user

Requirements:
- Use SQLite for database
- Include input validation
- Add error handling
- Write tests using pytest
```

### Identity Files

Create `identity/setup.py` to customize behavior:

```python
# identity/setup.py
"""
You are an expert Python developer specializing in FastAPI.
Always use type hints and Pydantic models.
Follow PEP 8 style guidelines.
Include comprehensive docstrings.
"""
```

## API/Protocol Endpoints

GPT-Engineer connects to LLM provider APIs:

### Supported Providers

| Provider | Environment Variable | Example Model |
|----------|---------------------|---------------|
| OpenAI | OPENAI_API_KEY | gpt-4o, gpt-5 |
| Anthropic | ANTHROPIC_API_KEY | claude-sonnet-4 |
| Azure | AZURE_OPENAI_KEY | gpt-4 |
| Local | OPENAI_API_BASE | local models |

### Custom Model Configuration

```bash
# Use local model
export OPENAI_API_KEY="dummy"
export OPENAI_API_BASE="http://localhost:11434/v1"
export MODEL="codellama"

gpt-engineer my-project
```

## Usage Examples

### Example 1: Simple Web Application

```bash
# Create project
mkdir todo-app
cd todo-app

# Create prompt
cat > prompt << 'EOF'
Create a Todo web application using:
- Backend: Python Flask with SQLAlchemy
- Frontend: HTML with vanilla JavaScript
- Database: SQLite

Features:
- Add, edit, delete todos
- Mark todos as complete
- Filter by status (all/active/completed)
- Responsive design with CSS
EOF

# Generate
gpt-engineer .

# Run the application
cd workspace
pip install -r requirements.txt
python app.py
```

### Example 2: Improving Existing Code

```bash
# Navigate to existing project
cd existing-project

# Create improvement prompt
cat > prompt << 'EOF'
Improve the existing codebase:
- Add input validation
- Implement proper error handling
- Add unit tests
- Refactor for better code organization
- Add type hints
EOF

# Run in improve mode
gpt-engineer . -i
```

### Example 3: Using Feedback Loop

```bash
# Initial generation
gpt-engineer my-project

# Review and provide feedback
# Create feedback file
cat > prompt << 'EOF'
The generated code works well but:
- Add authentication with JWT
- Implement rate limiting
- Add request logging
- Improve error messages
EOF

# Use feedback step
gpt-engineer . --steps use_feedback
```

### Example 4: Multi-Step Generation

```bash
# Step 1: Clarify requirements
gpt-engineer my-project --steps clarify

# Step 2: Generate from clarified requirements
gpt-engineer my-project --steps gen_clarified_code

# Step 3: Generate entrypoint
gpt-engineer my-project --steps gen_entrypoint

# Step 4: Execute
gpt-engineer my-project --steps execute_entrypoint
```

### Example 5: Custom Identity

```bash
# Create project with custom identity
mkdir api-project
cd api-project
mkdir identity

# Create custom identity
cat > identity/setup.py << 'EOF'
"""
You are a senior backend engineer specializing in:
- FastAPI for high-performance APIs
- PostgreSQL with SQLAlchemy async
- Docker and Kubernetes deployment
- Comprehensive test coverage

Always include:
- OpenAPI documentation
- Pydantic models
- Async/await patterns
- Docker compose setup
"""
EOF

# Create prompt
cat > prompt << 'EOF'
Create a microservice for user management with:
- RESTful API endpoints
- PostgreSQL database
- JWT authentication
- Docker setup
- Kubernetes manifests
- Comprehensive tests
EOF

# Generate
gpt-engineer .
```

### Example 6: Non-Interactive Mode

```bash
# Generate without user interaction
gpt-engineer my-project --no-execution

# Or with all defaults
echo "" | gpt-engineer my-project
```

## Troubleshooting

### Issue: API Key Not Found

**Symptoms:** "API key not found" error

**Solution:**
```bash
# Set environment variable
export OPENAI_API_KEY="sk-..."

# Or create .env file
echo "OPENAI_API_KEY=sk-..." > .env

# Verify
echo $OPENAI_API_KEY
```

### Issue: Model Not Available

**Symptoms:** "Model not found" or rate limit errors

**Solution:**
```bash
# Use different model
gpt-engineer . --model gpt-3.5-turbo

# Check API key has access to model
# Verify rate limits not exceeded
```

### Issue: No prompt File

**Symptoms:** "No prompt file found" error

**Solution:**
```bash
# Create prompt file
echo "Your requirements here" > prompt

# Or specify custom location
gpt-engineer . --prompt-file custom_prompt.txt
```

### Issue: Generation Stops Unexpectedly

**Symptoms:** Incomplete code generation

**Solution:**
```bash
# Use continue from feedback
gpt-engineer . --steps use_feedback

# Increase max tokens
export MAX_TOKENS=8000

# Check API rate limits
```

### Issue: Code Doesn't Execute

**Symptoms:** Generated code fails to run

**Solution:**
```bash
# Use self-heal mode
gpt-engineer . --self-heal

# Manual fix with improve mode
gpt-engineer . -i

# Check generated requirements.txt
pip install -r workspace/requirements.txt
```

### Issue: Workspace Directory Empty

**Symptoms:** No files in workspace/

**Solution:**
```bash
# Check for errors in output
# Verify write permissions
ls -la workspace/

# Try with verbose mode
gpt-engineer . --verbose
```

### Issue: Python Version Error

**Symptoms:** "Python 3.10+ required" error

**Solution:**
```bash
# Check Python version
python --version

# Use specific Python version
python3.12 -m pip install gpt-engineer
python3.12 -m gpt_engineer .
```

---

**Last Updated:** 2026-04-02
**Version:** 0.3.1
