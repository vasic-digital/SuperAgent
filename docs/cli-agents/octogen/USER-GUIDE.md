# Octogen User Guide

## Table of Contents
1. [Installation](#installation)
2. [Quick Start](#quick-start)
3. [CLI Commands](#cli-commands)
4. [TUI/Interactive Commands](#tuiinteractive-commands)
5. [Configuration](#configuration)
6. [API/Protocol Endpoints](#api-protocol-endpoints)
7. [Usage Examples](#usage-examples)
8. [Troubleshooting](#troubleshooting)

---

## Installation

### Prerequisites

- Python 3.10+
- pip
- Docker 24.0.0+ or Podman

### Method 1: Install Script (Recommended)

```bash
curl --proto '=https' --tlsv1.2 -sSf https://up.dbpunk.xyz | sh
```

This script will:
1. Install the `og_up` launcher
2. Guide you through setup
3. Configure API service (OpenAI, Azure, CodeLlama, or Octogen)
4. Set installation directory
5. Set kernel workspace directory

### Method 2: Docker Installation

```bash
# Install og_up
curl --proto '=https' --tlsv1.2 -sSf https://up.dbpunk.xyz | sh

# Run setup with Docker
og_up
# Select Docker option during setup
```

### Method 3: Podman Installation

```bash
# Install og_up
curl --proto '=https' --tlsv1.2 -sSf https://up.dbpunk.xyz | sh

# Run setup with Podman
og_up --use_podman
```

### Method 4: Build from Source

```bash
git clone https://github.com/dbpunk-labs/octogen.git
cd octogen
python3 -m venv octogen_venv
source octogen_venv/bin/activate
pip install -r requirements.txt
```

---

## Quick Start

```bash
# Run setup (first time only)
og_up

# Start Octogen
og

# You'll see:
# Welcome to use octogen❤️ . To ask a programming question, 
# simply type your question and press esc + enter
# You can use /help to look for help
#
# [1]🎧>

# Type your question and press Esc+Enter to submit
```

---

## CLI Commands

### Global Options

| Option | Description | Example |
|--------|-------------|---------|
| `--help` | Show help | `og --help` |
| `--version` | Show version | `og --version` |

### Command: og

**Description:** Start the Octogen interactive CLI.

**Usage:**
```bash
og [options]
```

**Examples:**
```bash
# Start Octogen
og

# You'll see the welcome message and prompt
```

**Exit Codes:**
- `0` - Success
- `1` - General error
- `130` - Interrupted (Ctrl+C)

### Command: og_up

**Description:** Setup and configure Octogen.

**Usage:**
```bash
og_up [options]
```

**Options:**
| Option | Description |
|--------|-------------|
| `--use_podman` | Use Podman instead of Docker |

**Examples:**
```bash
# Run setup
og_up

# With Podman
og_up --use_podman
```

During setup, you'll select:
1. **OpenAI** - Use OpenAI GPT-3.5/4 (recommended for daily use)
2. **Azure OpenAI** - Use Azure OpenAI service
3. **CodeLlama** - Use local CodeLlama model (requires 8+ CPUs, 16GB+ RAM)
4. **Octogen** - Use Octogen agent services

---

## TUI/Interactive Commands

Once inside the Octogen CLI, use these slash commands:

| Command | Description | Example |
|---------|-------------|---------|
| `/help` | Show help message | `/help` |
| `/up` | Upload files | `/up /path/to/file.txt` |
| `/run` | Run assembled application | `/run` |
| `/cc` | Copy output to clipboard | `/cc` |
| `/quit` | Exit Octogen | `/quit` |

### Submitting Queries

In Octogen, type your question and press **Esc + Enter** to submit (not just Enter).

---

## Configuration

### Supported API Services

| Service | Type | Status | Setup Command |
|---------|------|--------|---------------|
| OpenAI GPT 3.5/4 | LLM | Fully supported | `og_up` → Select OpenAI |
| Azure OpenAI GPT 3.5/4 | LLM | Fully supported | `og_up` → Select Azure |
| LLama.cpp Server | LLM | Supported | `og_up` → Select CodeLlama |
| Octogen Agent Service | Code Interpreter | Supported | `og_up` → Select Octogen |

### Environment Variables

| Variable | Description |
|----------|-------------|
| `OPENAI_API_KEY` | OpenAI API key |
| `AZURE_OPENAI_KEY` | Azure OpenAI key |
| `AZURE_OPENAI_ENDPOINT` | Azure OpenAI endpoint |
| `OCTOGEN_API_KEY` | Octogen service API key |

### Configuration Files

**OpenAI Configuration:**
```bash
# Set in environment
export OPENAI_API_KEY="sk-..."
```

**Azure OpenAI Configuration:**
```bash
export AZURE_OPENAI_KEY="..."
export AZURE_OPENAI_ENDPOINT="https://your-resource.openai.azure.com/"
```

**CodeLlama Configuration:**
- Requires local setup
- Uses LLama.cpp server
- Minimum 8 CPUs, 16GB RAM

### Docker Compose Configuration

When using Docker deployment, configuration is in `docker-compose.yaml`:

```yaml
version: '3.8'
services:
  octogen-agent:
    image: octogen/agent:latest
    environment:
      - OPENAI_API_KEY=${OPENAI_API_KEY}
    volumes:
      - ./workspace:/workspace
  
  octogen-kernel:
    image: octogen/kernel:latest
    volumes:
      - ./workspace:/workspace
```

---

## API/Protocol Endpoints

### Internal Architecture

Octogen consists of three components:

1. **Octogen Kernel** - Code execution engine based on notebook kernels
2. **Octogen Agent** - Manages requests, uses ReAct for complex tasks
3. **Octogen Terminal CLI** - Accepts user requests and renders results

Data flows between components using streaming for real-time updates.

### Supported Terminals

| Terminal | Image Rendering | Notes |
|----------|-----------------|-------|
| iTerm2 | ✅ | Full image support |
| Kitty | ✅ | Full image support |
| Discord | ✅ | Via bot integration |
| Standard | ⚠️ | Text only |

---

## Usage Examples

### Example 1: Basic Setup with OpenAI

```bash
# Set API key
export OPENAI_API_KEY="sk-..."

# Run setup
og_up
# Select "OpenAI" when prompted

# Start Octogen
og

# Ask a question
[1]🎧> Write a Python function to calculate fibonacci numbers
# Press Esc+Enter to submit
```

### Example 2: Data Analysis

```bash
# Start Octogen
og

# Upload a file
[1]🎧> /up /path/to/data.csv

# Analyze the data
[2]🎧> Analyze this CSV and create visualizations

# Octogen will:
# 1. Read the file
# 2. Generate analysis code
# 3. Execute in Docker environment
# 4. Display results (including images in iTerm2/Kitty)
```

### Example 3: Code Generation and Execution

```bash
og

[1]🎧> Create a web scraper for Hacker News
# Octogen generates code

[2]🎧> Run it and save the top 10 stories to a file
# Octogen executes the code

[3]🎧> /cc
# Copy output to clipboard
```

### Example 4: Local CodeLlama Setup

```bash
# Ensure you have 8+ CPUs and 16GB+ RAM

# Run setup
og_up
# Select "CodeLlama"

# Setup will configure LLama.cpp server

# Start Octogen
og

[1]🎧> Explain how recursion works
# Uses local CodeLlama model
```

### Example 5: Application Assembly

```bash
og

[1]🎧> Create a calculator app in Python with GUI

[2]🎧> Add support for scientific functions

[3]🎧> /run
# Run the assembled application
```

### Example 6: File Processing Pipeline

```bash
og

[1]🎧> /up raw_data.json

[2]🎧> Clean this JSON data and convert to CSV

[3]🎧> Generate summary statistics

[4]🎧> Create a visualization of the key metrics
```

### Example 7: Multi-step Development

```bash
og

[1]🎧> Create a Flask API for a todo app

[2]🎧> Add SQLAlchemy models for the database

[3]🎧> Write CRUD endpoints

[4]🎧> Create a simple HTML frontend

[5]🎧> Run the application locally
```

---

## Troubleshooting

### Issue: "og command not found"

**Solution:**
```bash
# Check installation
which og
which og_up

# If not found, re-run setup
curl --proto '=https' --tlsv1.2 -sSf https://up.dbpunk.xyz | sh

# Add to PATH if needed
export PATH="$HOME/.local/bin:$PATH"
```

### Issue: Docker not running

**Solution:**
```bash
# Start Docker
sudo systemctl start docker

# Or use Podman
og_up --use_podman
```

### Issue: API key not set

**Solution:**
```bash
# For OpenAI
export OPENAI_API_KEY="sk-..."

# For Azure
export AZURE_OPENAI_KEY="..."
export AZURE_OPENAI_ENDPOINT="..."

# Re-run setup
og_up
```

### Issue: CodeLlama requires too much resources

**Solution:**
- Ensure minimum 8 CPUs and 16GB RAM
- Use OpenAI or Azure instead
- Or use Octogen agent service

### Issue: Images not displaying

**Solution:**
- Use iTerm2 or Kitty terminal for image support
- Standard terminals show text descriptions instead
- Images are still generated and saved to workspace

### Issue: Kernel not responding

**Solution:**
```bash
# Check Docker containers
docker ps
docker logs octogen-kernel

# Restart services
docker compose restart
```

### Issue: Prompt history lost

**Solution:**
- History is stored in `~/.octogen/history/`
- Check directory permissions
- Ensure disk space available

### Issue: File upload fails

**Solution:**
```bash
# Check file exists
ls -la /path/to/file

# Check file size (large files may timeout)
# Check workspace directory permissions
ls -la ~/.octogen/workspace/
```

### Issue: Code execution errors

**Solution:**
- Check Docker container has necessary packages
- Review error messages in output
- Try breaking task into smaller steps

---

## Additional Resources

- **GitHub Repository:** https://github.com/dbpunk-labs/octogen
- **Installation Script:** https://up.dbpunk.xyz
- **Docker Hub:** https://hub.docker.com/r/octogen
- **OpenAI API:** https://platform.openai.com
- **Azure OpenAI:** https://azure.microsoft.com/en-us/services/cognitive-services/openai-service
