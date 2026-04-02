# GPT-Engineer - Usage Guide

## Table of Contents

1. [Getting Started](#getting-started)
2. [Basic Workflows](#basic-workflows)
3. [Advanced Features](#advanced-features)
4. [Best Practices](#best-practices)
5. [Troubleshooting](#troubleshooting)

---

## Getting Started

### Installation

**Stable Release (Recommended):**
```bash
pip install gpt-engineer
```

**Development Version:**
```bash
git clone https://github.com/gpt-engineer-org/gpt-engineer.git
cd gpt-engineer
poetry install
poetry shell
```

**Requirements:**
- Python 3.10 - 3.12
- OpenAI API key (or Anthropic/Azure/local LLM)

### First-Time Setup

1. **Get API Key:**
   - OpenAI: https://platform.openai.com/account/api-keys
   - Anthropic: https://console.anthropic.com/
   - Set rate limits to control costs!

2. **Configure Environment:**
   ```bash
   # Option 1: Export (recommended)
   export OPENAI_API_KEY="sk-xxxxxxxx"
   
   # Option 2: .env file
   cp .env.template .env
   # Edit .env and add your key
   ```

3. **Verify Installation:**
   ```bash
   gpte --help
   ```

### Your First Project

```bash
# 1. Create project directory
mkdir -p projects/my-first-app
cd projects/my-first-app

# 2. Create prompt file
cat > prompt << 'EOF'
Create a Python script that:
- Fetches weather data from OpenWeatherMap API
- Takes a city name as input
- Displays temperature, humidity, and description
- Includes error handling
EOF

# 3. Run GPT-Engineer
gpte .

# 4. Check results
ls workspace/
```

---

## Basic Workflows

### 1. New Project Generation

**Simple Web Application:**
```bash
# Setup
mkdir projects/web-app
echo "Create a Flask web app with user login and a todo list" > projects/web-app/prompt

# Generate
gpte projects/web-app

# Results in: projects/web-app/workspace/
```

**API Service:**
```bash
echo "Create a FastAPI REST API for managing books with CRUD operations and SQLite" > projects/api/prompt
gpte projects/api
```

**CLI Tool:**
```bash
echo "Create a command-line tool that converts JSON to YAML with validation" > projects/cli-tool/prompt
gpte projects/cli-tool
```

### 2. Code Improvement

**Workflow:**
```bash
# 1. Navigate to existing project
cd projects/existing-app

# 2. Create improvement prompt
cat > prompt << 'EOF'
Improve this codebase by:
- Adding input validation
- Implementing proper error handling
- Adding unit tests
- Using type hints throughout
EOF

# 3. Run in improve mode
gpte . -i

# 4. Review changes and approve
```

**Common Improvements:**
- Add logging
- Implement authentication
- Add database migrations
- Optimize performance
- Add documentation
- Refactor for readability

### 3. Mode Selection Guide

| Scenario | Recommended Mode | Command |
|----------|-----------------|---------|
| New simple project | Default | `gpte project/` |
| New complex project | Clarify | `gpte project/ -c` |
| Quick prototype | Lite | `gpte project/ -l` |
| Fix/extend existing | Improve | `gpte project/ -i` |
| Need reliable execution | Self-heal | `gpte project/ -sh` |
| Local LLM | Lite | `gpte project/ -l --temperature 0.1` |

---

## Advanced Features

### 1. Clarify Mode

**When to Use:** Complex requirements that need refinement

```bash
# Create detailed prompt
cat > projects/complex-app/prompt << 'EOF'
Build a microservices architecture with:
- User service (authentication)
- Order service (business logic)
- Notification service (async processing)
- API Gateway (routing)
EOF

# Run with clarify mode
gpte projects/complex-app -c
```

**Process:**
1. GPT-Engineer analyzes requirements
2. Asks clarifying questions
3. Generates specification
4. Proceeds with implementation

### 2. Vision Support

**Use Case:** Generate code from UI mockups or diagrams

```bash
# Setup directory structure
mkdir -p projects/ui-app/prompt
cp mockup.png projects/ui-app/prompt/

# Create text prompt
echo "Implement this UI in React with Tailwind CSS" > projects/ui-app/prompt/text

# Run with vision model
gpte projects/ui-app gpt-4-vision-preview \
  --prompt_file prompt/text \
  --image_directory prompt
```

### 3. Custom Preprompts

**Customize AI Behavior:**

```bash
# 1. Copy default preprompts
gpte projects/my-app --use-custom-preprompts

# 2. Edit preprompts/generate
cat > projects/my-app/preprompts/generate << 'EOF'
Think step by step and reason yourself to the correct decisions.
You are an expert Python developer who follows PEP 8 and uses type hints.
Always include docstrings and unit tests.
Prefer async/await for I/O operations.

FILE_FORMAT

You will start with the "entrypoint" file, then go to the ones that are imported.
Please note that the code should be fully functional. No placeholders.
Follow best practices and include proper error handling.
When you are done, write "this concludes a fully working implementation".
EOF

# 3. Regenerate with custom identity
gpte projects/my-app --use-custom-preprompts
```

### 4. Local LLM Integration

**Setup llama.cpp:**

```bash
# Install with hardware acceleration
# macOS (Metal)
CMAKE_ARGS="-DLLAMA_METAL=on" pip install llama-cpp-python

# Linux (OpenBLAS)
CMAKE_ARGS="-DLLAMA_BLAS=ON -DLLAMA_BLAS_VENDOR=OpenBLAS" \
  pip install llama-cpp-python

# Start server
python -m llama_cpp.server \
  --model /path/to/codellama-70b.Q6_K.gguf \
  --n_batch 256 \
  --n_gpu_layers 30
```

**Use with GPT-Engineer:**
```bash
export OPENAI_API_BASE="http://localhost:8000/v1"
export OPENAI_API_KEY="sk-xxx"
export MODEL_NAME="CodeLlama-70B"
export LOCAL_MODEL=true

gpte projects/my-app $MODEL_NAME --lite --temperature 0.1
```

### 5. Azure OpenAI

```bash
# Configure
export OPENAI_API_KEY="your-azure-key"
export OPENAI_API_VERSION="2024-05-01-preview"

# Run with Azure endpoint
gpte projects/my-app my-deployment-name \
  --azure https://my-resource.openai.azure.com
```

### 6. Benchmarking

**Create Custom Agent:**
```python
# my_agent.py
from gpt_engineer.applications.cli.cli_agent import CliAgent
from gpt_engineer.core.default.disk_memory import DiskMemory
from gpt_engineer.core.default.disk_execution_env import DiskExecutionEnv
from gpt_engineer.core.ai import AI

def default_config_agent():
    ai = AI(model_name="gpt-4o", temperature=0.1)
    memory = DiskMemory(".gpteng/memory")
    execution_env = DiskExecutionEnv()
    return CliAgent.with_default_config(memory, execution_env, ai=ai)
```

**Run Benchmark:**
```bash
bench my_agent.py bench_config.toml --verbose
```

---

## Best Practices

### 1. Prompt Engineering

**Be Specific:**
```
# Good
Create a Python script that fetches weather data from the
OpenWeatherMap API. It should accept a city name as a command-line
argument, handle API errors gracefully, and display temperature
in both Celsius and Fahrenheit.

# Less effective
Create a weather app
```

**Include Constraints:**
```
Build a React component that:
- Uses functional components and hooks
- Follows Airbnb ESLint rules
- Includes PropTypes validation
- Has 90%+ test coverage
- Supports dark mode
```

**Specify Tech Stack:**
```
Create a REST API using:
- FastAPI framework
- SQLAlchemy ORM
- PostgreSQL database
- Pydantic for validation
- pytest for testing
```

### 2. Cost Management

**Monitor Usage:**
```bash
# GPT-Engineer shows cost after each run
# Example output:
# Total api cost: $ 0.14
```

**Reduce Costs:**
- Use GPT-3.5-turbo for simple tasks: `-m gpt-3.5-turbo`
- Enable caching: `--use_cache`
- Use lite mode for prototypes: `-l`
- Set OpenAI rate limits

**Typical Costs:**
| Project Type | Model | Approximate Cost |
|-------------|-------|------------------|
| Simple script | GPT-3.5 | $0.01 - $0.05 |
| Web app | GPT-4o | $0.10 - $0.50 |
| Complex API | GPT-4-turbo | $0.50 - $2.00 |
| Large refactor | GPT-4o | $0.20 - $1.00 |

### 3. Version Control

**Always Use Git:**
```bash
cd projects/my-app
git init
git add .
git commit -m "Initial commit"

# GPT-Engineer auto-stages changes
gpte . -i
# Review changes:
git diff --cached
```

**Recovery:**
```bash
# If something goes wrong
git checkout -- .
git clean -fd
```

### 4. Iterative Development

**Step-by-Step Approach:**
1. Start with core functionality
2. Test generated code
3. Create improvement prompt
4. Run in improve mode
5. Repeat until satisfied

**Example Iteration:**
```bash
# Iteration 1: Basic structure
echo "Create a todo app with Flask" > prompt
gpte .

# Iteration 2: Add database
echo "Add PostgreSQL database with SQLAlchemy models" > prompt
gpte . -i

# Iteration 3: Add auth
echo "Add JWT authentication to protect todo endpoints" > prompt
gpte . -i

# Iteration 4: Add tests
echo "Add pytest unit tests with 90% coverage" > prompt
gpte . -i
```

### 5. Code Review

**Always Review Generated Code:**
- Check for security issues
- Verify error handling
- Review dependencies
- Test before deploying

**Common Issues to Check:**
- Hardcoded credentials
- Missing input validation
- SQL injection vulnerabilities
- Insecure HTTP configurations

---

## Troubleshooting

### Common Issues

**1. API Key Errors**

```
Error: No API key provided
```

**Solution:**
```bash
# Verify key is set
echo $OPENAI_API_KEY

# Set if missing
export OPENAI_API_KEY="sk-xxxxxxxx"

# Or use .env file
cp .env.template .env
# Edit and add your key
```

**2. Rate Limit Errors**

```
RateLimitError: Rate limit exceeded
```

**Solution:**
- GPT-Engineer has built-in retry with exponential backoff
- Wait a minute and retry
- Check your OpenAI rate limits
- Consider upgrading your plan

**3. Model Not Available**

```
Error: Model gpt-4 not available
```

**Solution:**
```bash
# Use available model
gpte projects/my-app gpt-3.5-turbo

# Or check your OpenAI account has GPT-4 access
```

**4. Improve Mode - No Changes**

```
No changes applied. Could you please upload the debug_log_file.txt...
```

**Solution:**
- Check `memory/logs/debug_log_file.txt`
- Try rephrasing your prompt
- Ensure files are properly selected
- Check diff_timeout setting: `--diff_timeout 10`

**5. Local LLM Connection Failed**

```
Connection refused to localhost:8000
```

**Solution:**
```bash
# Verify server is running
curl http://localhost:8000/v1/models

# Start llama.cpp server
python -m llama_cpp.server --model /path/to/model.gguf

# Check environment variables
export OPENAI_API_BASE="http://localhost:8000/v1"
export LOCAL_MODEL=true
```

**6. Vision Mode Not Working**

**Solution:**
```bash
# Use vision-capable model
gpte project/ gpt-4-vision-preview --image_directory images/

# Check images exist
ls project/prompt/images/
```

### Debug Mode

```bash
# Enable verbose logging
gpte projects/my-app -v

# Enable debug mode (pdb on error)
gpte projects/my-app -d

# Output system info
gpte projects/my-app --sysinfo
```

### Getting Help

**Resources:**
- Documentation: https://gpt-engineer.readthedocs.io
- GitHub Issues: https://github.com/gpt-engineer-org/gpt-engineer/issues
- Discord: https://discord.gg/8tcDQ89Ej2

**Report Bug:**
```bash
# Include debug logs when reporting issues
cat projects/my-app/memory/logs/debug_log_file.txt
```

---

## Quick Reference Card

### Essential Commands

| Command | Purpose |
|---------|---------|
| `gpte project/` | Generate new code |
| `gpte project/ -i` | Improve existing code |
| `gpte project/ -l` | Lite mode (fast) |
| `gpte project/ -c` | Clarify mode |
| `gpte project/ -sh` | Self-heal mode |
| `bench agent.py` | Run benchmark |

### Common Flags

| Flag | Purpose |
|------|---------|
| `-m gpt-4o` | Specify model |
| `-t 0.5` | Set temperature |
| `-v` | Verbose output |
| `--use_cache` | Enable caching |
| `--skip-file-selection` | Skip prompts |

### File Structure

```
project/
├── prompt              # Your requirements
├── preprompts/         # Custom AI identity
├── memory/             # Conversation history
│   └── logs/           # Debug logs
└── workspace/          # Generated code
```

---

*For more details, see the [API Reference](./API.md) and [Architecture](./ARCHITECTURE.md) documentation.*
