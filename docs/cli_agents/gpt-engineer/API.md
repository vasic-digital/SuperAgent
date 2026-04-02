# GPT-Engineer - API Reference

## Command Line Interface

### Global Commands

```bash
gpte [OPTIONS] [PROJECT_PATH] [MODEL]
```

### Positional Arguments

| Argument | Description | Default |
|----------|-------------|---------|
| `PROJECT_PATH` | Path to project directory | `.` (current directory) |
| `MODEL` | Model ID string | `gpt-4o` (or `$MODEL_NAME` env) |

### Options Reference

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--help` | `-h` | Show help message | - |
| `--model` | `-m` | Model ID to use | `gpt-4o` |
| `--temperature` | `-t` | Sampling temperature (0.0-2.0) | `0.1` |
| `--improve` | `-i` | Improve existing code | `False` |
| `--lite` | `-l` | Lite mode (skip preprompts) | `False` |
| `--clarify` | `-c` | Clarify requirements first | `False` |
| `--self-heal` | `-sh` | Auto-fix execution failures | `False` |
| `--azure` | `-a` | Azure OpenAI endpoint | `""` |
| `--use-custom-preprompts` | - | Use custom preprompts | `False` |
| `--llm-via-clipboard` | - | Manual LLM via clipboard | `False` |
| `--verbose` | `-v` | Enable verbose logging | `False` |
| `--debug` | `-d` | Enable debug mode (pdb) | `False` |
| `--prompt_file` | - | Path to prompt file | `prompt` |
| `--entrypoint_prompt` | - | Entrypoint requirements file | `""` |
| `--image_directory` | - | Directory with images | `""` |
| `--use_cache` | - | Enable LLM response caching | `False` |
| `--skip-file-selection` | `-s` | Skip interactive file selection | `False` |
| `--no_execution` | - | Setup only, don't run | `False` |
| `--sysinfo` | - | Output system information | `False` |
| `--diff_timeout` | - | Diff regex timeout (seconds) | `3` |

---

## Environment Variables

### Required Variables

| Variable | Description | Required For |
|----------|-------------|--------------|
| `OPENAI_API_KEY` | OpenAI API key | OpenAI models |
| `ANTHROPIC_API_KEY` | Anthropic API key | Claude models |

### Optional Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `MODEL_NAME` | Default model name | `gpt-4o` |
| `OPENAI_API_BASE` | Custom API base URL | OpenAI default |
| `LOCAL_MODEL` | Flag for local LLM | `false` |
| `WANDB_API_KEY` | Weights & Biases API key | - |
| `LANGCHAIN_WANDB_TRACING` | Enable W&B tracing | `false` |

### Azure-Specific Variables

| Variable | Description |
|----------|-------------|
| `OPENAI_API_VERSION` | Azure API version | `2024-05-01-preview` |

---

## Usage Examples

### Basic Code Generation

```bash
# Generate new project
gpte projects/my-app

# With specific model
gpte projects/my-app gpt-4-turbo

# With custom temperature
gpte projects/my-app --temperature 0.5
```

### Improve Existing Code

```bash
# Improve mode with file selection
gpte projects/my-app -i

# Skip file selection (use existing toml)
gpte projects/my-app -i --skip-file-selection
```

### Alternative Modes

```bash
# Clarify mode - ask questions before coding
gpte projects/my-app -c

# Lite mode - faster, simpler generation
gpte projects/my-app -l

# Self-heal mode - auto-fix execution errors
gpte projects/my-app -sh
```

### Vision Support

```bash
# Generate with image context
gpte projects/my-app gpt-4-vision-preview \
  --prompt_file prompt/text \
  --image_directory prompt/images
```

### Azure OpenAI

```bash
# Use Azure deployment
gpte projects/my-app my-deployment-name \
  --azure https://my-resource.openai.azure.com
```

### Local LLM

```bash
# Setup environment
export OPENAI_API_BASE="http://localhost:8000/v1"
export OPENAI_API_KEY="sk-xxx"
export MODEL_NAME="CodeLlama-70B"
export LOCAL_MODEL=true

# Run with lite mode (recommended for local models)
gpte projects/my-app $MODEL_NAME --lite --temperature 0.1
```

### Docker Usage

```bash
# Build image
docker build --rm -t gpt-engineer -f docker/Dockerfile .

# Run with API key
docker run -it --rm \
  -e OPENAI_API_KEY="YOUR_KEY" \
  -v ./your-project:/project \
  gpt-engineer

# Using docker-compose
docker-compose up -d --build
docker-compose run --rm gpt-engineer
```

---

## Benchmark Command

### Syntax

```bash
bench <path_to_agent> [bench_config.toml] [OPTIONS]
```

### Arguments

| Argument | Description | Default |
|----------|-------------|---------|
| `path_to_agent` | Python file with `default_config_agent()` function | Required |
| `bench_config` | TOML configuration file | `default_bench_config.toml` |

### Options

| Option | Description | Default |
|--------|-------------|---------|
| `--yaml-output` | Export results to YAML file | `None` |
| `--verbose` | Print results for each task | `False` |
| `--use-cache` | Enable LLM caching | `True` |

### Example

```bash
# Run benchmark with default config
bench my_agent.py

# With custom config and output
bench my_agent.py my_config.toml \
  --yaml-output results.yaml \
  --verbose
```

---

## Configuration Files

### Project Prompt File

**File**: `prompt` (no extension)

```
Create a Python REST API with the following features:
- User authentication (JWT)
- CRUD operations for tasks
- PostgreSQL database
- FastAPI framework
- Include tests
```

### Environment File

**File**: `.env`

```bash
# OpenAI
OPENAI_API_KEY=sk-xxxxxxxxxxxxxxxxxxxxxxxx

# Anthropic (optional)
ANTHROPIC_API_KEY=sk-ant-xxxxxxxx

# Model configuration
MODEL_NAME=gpt-4o
```

### Custom Preprompts

**Directory**: `preprompts/`

```
preprompts/
├── clarify           # Clarification questions
├── generate          # Code generation identity
├── improve           # Improvement instructions
├── file_format       # Output format
├── file_format_diff  # Diff format
├── file_format_fix   # Fix format
├── entrypoint        # Execution instructions
└── philosophy        # AI behavior
```

### File Selection Config

**File**: `.gpteng/file_selection.toml` (auto-generated)

```toml
[files]
selected = [
    "src/main.py",
    "src/utils.py",
    "tests/test_main.py"
]
linting_enabled = true
```

---

## AI Class Reference

### Constructor

```python
AI(
    model_name: str = "gpt-4-turbo",
    temperature: float = 0.1,
    azure_endpoint: Optional[str] = None,
    streaming: bool = True,
    vision: bool = False
)
```

### Methods

| Method | Description | Parameters |
|--------|-------------|------------|
| `start()` | Initialize conversation | `system`, `user`, `step_name` |
| `next()` | Continue conversation | `messages`, `prompt`, `step_name` |
| `backoff_inference()` | LLM call with retry | `messages` |
| `serialize_messages()` | Serialize to JSON | `messages` |
| `deserialize_messages()` | Deserialize from JSON | `jsondictstr` |

### Example

```python
from gpt_engineer.core.ai import AI

# Initialize
ai = AI(
    model_name="gpt-4o",
    temperature=0.1,
    streaming=True
)

# Start conversation
messages = ai.start(
    system="You are a Python expert",
    user="Create a Flask app",
    step_name="initial"
)

# Continue
messages = ai.next(
    messages,
    prompt="Add user authentication",
    step_name="iteration"
)
```

---

## Core Classes

### FilesDict

In-memory representation of a file collection.

```python
from gpt_engineer.core.files_dict import FilesDict

# Create
files = FilesDict({
    "main.py": "print('hello')",
    "README.md": "# Project"
})

# Access
content = files["main.py"]

# Iterate
for path, content in files.items():
    print(f"{path}: {len(content)} chars")
```

### DiskMemory

Persistent storage for conversations and logs.

```python
from gpt_engineer.core.default.disk_memory import DiskMemory

memory = DiskMemory("path/to/memory")
memory["conversation"] = "chat log content"
content = memory.get("conversation")
```

### DiskExecutionEnv

Environment for executing generated code.

```python
from gpt_engineer.core.default.disk_execution_env import DiskExecutionEnv

env = DiskExecutionEnv()
env.execute_program("python main.py")
```

---

## Prompt Class

### Constructor

```python
Prompt(
    text: str,
    image_urls: Optional[Dict[str, bytes]] = None,
    entrypoint_prompt: str = ""
)
```

### Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `text` | `str` | Main prompt text |
| `image_urls` | `Dict[str, bytes]` | Image data for vision models |
| `entrypoint_prompt` | `str` | Additional entrypoint requirements |

---

## Token Usage

### Tracking

```python
from gpt_engineer.core.ai import AI

ai = AI(model_name="gpt-4o")

# After generation
cost = ai.token_usage_log.usage_cost()
total_tokens = ai.token_usage_log.total_tokens()

print(f"Cost: ${cost}")
print(f"Tokens: {total_tokens}")
```

### Cost Calculation

| Model | Input Cost | Output Cost |
|-------|------------|-------------|
| GPT-4o | $5 / 1M tokens | $15 / 1M tokens |
| GPT-4-turbo | $10 / 1M tokens | $30 / 1M tokens |
| GPT-3.5-turbo | $0.5 / 1M tokens | $1.5 / 1M tokens |
| Claude-3 | Varies | Varies |
| Local LLM | $0 | $0 |

---

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | General error / Invalid arguments |
| `130` | Interrupted (Ctrl+C) |

---

## Python API Usage

### Basic Example

```python
from pathlib import Path
from gpt_engineer.core.ai import AI
from gpt_engineer.core.default.disk_memory import DiskMemory
from gpt_engineer.core.default.disk_execution_env import DiskExecutionEnv
from gpt_engineer.core.default.file_store import FileStore
from gpt_engineer.core.default.steps import gen_code, execute_entrypoint
from gpt_engineer.core.preprompts_holder import PrepromptsHolder
from gpt_engineer.core.prompt import Prompt
from gpt_engineer.applications.cli.cli_agent import CliAgent

# Setup
path = Path("projects/my-app")
ai = AI(model_name="gpt-4o", temperature=0.1)
prompt = Prompt("Create a Python script that prints 'Hello World'")

# Configure
memory = DiskMemory(path / "memory")
execution_env = DiskExecutionEnv()
preprompts_holder = PrepromptsHolder(path / "preprompts")

# Create agent
agent = CliAgent.with_default_config(
    memory=memory,
    execution_env=execution_env,
    ai=ai,
    preprompts_holder=preprompts_holder
)

# Generate
files_dict = agent.init(prompt)

# Save
file_store = FileStore(path)
file_store.push(files_dict)
```

---

## Related Documentation

- [Architecture](./ARCHITECTURE.md) - System design
- [Usage Guide](./USAGE.md) - Practical examples
- [References](./REFERENCES.md) - External resources
- [Diagrams](./DIAGRAMS.md) - Visual documentation
- [Gap Analysis](./GAP_ANALYSIS.md) - Improvement opportunities
