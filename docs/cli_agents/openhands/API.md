# OpenHands - API Reference

## Command Line Interface

### Core CLI Commands

```bash
# Main entry point
python -m openhands.core.main [options]

# Alternative with poetry
poetry run python -m openhands.core.main [options]
```

### CLI Options

| Option | Short | Description | Example |
|--------|-------|-------------|---------|
| `--task` | `-t` | Task to execute | `-t "Write a Python script"` |
| `--file` | `-f` | Read task from file | `-f task.txt` |
| `--config` | `-c` | Config file path | `-c config.toml` |
| `--name` | `-n` | Session name | `-n my-session` |
| `--no-auto-continue` | | Disable auto-continue | `--no-auto-continue` |
| `--help` | `-h` | Show help | `--help` |

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `LLM_API_KEY` | LLM provider API key | None |
| `LLM_MODEL` | Model to use | `gpt-4o` |
| `LLM_BASE_URL` | Custom API base URL | Provider default |
| `AGENT_MEMORY_ENABLED` | Enable memory system | `true` |
| `SANDBOX_TIMEOUT` | Command timeout (seconds) | `120` |
| `WORKSPACE_BASE` | Workspace directory | `./workspace` |
| `DEBUG` | Enable debug logging | `false` |

---

## Configuration File (TOML)

### Core Section

```toml
[core]
workspace_base = "./workspace"
cache_dir = "/tmp/cache"
debug = false
disable_color = false
save_trajectory_path = "./trajectories"
save_screenshots_in_trajectory = false
max_iterations = 500
max_budget_per_task = 0.0
enable_browser = true
default_agent = "CodeActAgent"
```

### LLM Section

```toml
[llm]
model = "gpt-4o"
api_key = "your-api-key"
base_url = ""
temperature = 0.0
top_p = 1.0
max_input_tokens = 0
max_output_tokens = 0
num_retries = 8
retry_min_wait = 15
retry_max_wait = 120
retry_multiplier = 2.0
caching_prompt = true
timeout = 0

# Reasoning effort (OpenAI o-series)
reasoning_effort = "medium"

# Safety settings
safety_settings = []

# Draft editor configuration
[llm.draft_editor]
correct_num = 5
```

### Agent Section

```toml
[agent]
enable_browsing = true
enable_jupyter = true
enable_editor = true
enable_llm_editor = false
enable_cmd = true
enable_think = true
enable_finish = true
enable_prompt_extensions = true
enable_history_truncation = true
enable_condensation_request = false
disabled_microagents = []
```

### Sandbox Section

```toml
[sandbox]
timeout = 120
user_id = 1000
base_container_image = "nikolaik/python-nodejs:python3.12-nodejs22"
use_host_network = false
enable_auto_lint = false
initialize_plugins = true
keep_runtime_alive = false
pause_closed_runtimes = false
close_delay = 300
enable_gpu = false
volumes = "/host/path:/container/path:rw"
```

### Security Section

```toml
[security]
confirmation_mode = false
security_analyzer = "llm"
enable_security_analyzer = true
```

### Condenser Section

```toml
[condenser]
type = "noop"  # Options: noop, llm, amortized, llm_attention, recent, observation_masking

# For LLM condenser
# llm_config = "condenser"
# keep_first = 1
# max_size = 100

# For recent/amortized condenser
# keep_first = 1
# max_events = 100
```

### MCP Section

```toml
[mcp]
# Stdio servers (recommended for development)
stdio_servers = [
    {name = "filesystem", command = "npx", args = ["@modelcontextprotocol/server-filesystem", "/"]},
    {name = "fetch", command = "uvx", args = ["mcp-server-fetch"], env = {DEBUG = "true"}}
]

# SHTTP servers (recommended for production)
shttp_servers = [
    {url = "https://api.example.com/mcp/shttp", timeout = 180}
]
```

### Kubernetes Section

```toml
[kubernetes]
namespace = "default"
ingress_domain = "localhost"
pvc_storage_size = "2Gi"
pvc_storage_class = "standard"
resource_cpu_request = "1"
resource_memory_request = "1Gi"
resource_memory_limit = "2Gi"
privileged = false
```

---

## Server API Endpoints

### WebSocket

**Endpoint:** `ws://host:port/ws`

**Actions (Client → Server):**

```json
// Initialize
{
  "action": "initialize",
  "args": {
    "model": "gpt-4o",
    "directory": "/workspace",
    "agent_cls": "CodeActAgent"
  }
}

// Start task
{
  "action": "start",
  "args": {
    "task": "Write a hello world program"
  }
}

// Run command
{
  "action": "run",
  "args": {
    "command": "ls -la"
  }
}

// Read file
{
  "action": "read",
  "args": {
    "path": "/workspace/file.txt"
  }
}

// Write file
{
  "action": "write",
  "args": {
    "path": "/workspace/file.txt",
    "content": "Hello World"
  }
}

// Browse
{
  "action": "browse",
  "args": {
    "url": "https://example.com"
  }
}
```

**Observations (Server → Client):**

```json
// Command output
{
  "observation": "run",
  "content": "file1.txt file2.txt",
  "extras": {
    "command": "ls",
    "exit_code": 0
  }
}

// File content
{
  "observation": "read",
  "content": "file contents here",
  "extras": {
    "path": "/workspace/file.txt"
  }
}

// Browser content
{
  "observation": "browse",
  "content": "HTML content",
  "extras": {
    "url": "https://example.com"
  }
}

// Chat message
{
  "observation": "chat",
  "content": "Hello! How can I help?"
}
```

### REST API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/default-settings` | GET | Get default configuration |
| `/api/options/models` | GET | List available LLM models |
| `/api/options/agents` | GET | List available agents |
| `/api/options/security-analyzers` | GET | List security analyzers |
| `/api/options/config` | GET | Get config options |
| `/api/list-files` | GET | List workspace files |
| `/api/select-file` | POST | Select a file |
| `/api/upload-files` | POST | Upload files |
| `/api/security/analyze` | POST | Run security analysis |

---

## Agent Actions Reference

### CodeAct Agent Tools

#### execute_bash

Execute Linux bash commands.

```python
{
  "name": "execute_bash",
  "arguments": {
    "command": "ls -la",
    "description": "List directory contents"
  }
}
```

#### execute_ipython_cell

Run Python code in IPython environment.

```python
{
  "name": "execute_ipython_cell",
  "arguments": {
    "code": "import pandas as pd\ndf = pd.read_csv('data.csv')\nprint(df.head())"
  }
}
```

#### str_replace_editor

View and edit files.

```python
# View file
{
  "name": "str_replace_editor",
  "arguments": {
    "command": "view",
    "path": "/workspace/file.py"
  }
}

# Create file
{
  "name": "str_replace_editor",
  "arguments": {
    "command": "create",
    "path": "/workspace/new_file.py",
    "file_text": "print('hello')"
  }
}

# Edit file
{
  "name": "str_replace_editor",
  "arguments": {
    "command": "str_replace",
    "path": "/workspace/file.py",
    "old_str": "def old():\n    pass",
    "new_str": "def new():\n    return 42"
  }
}
```

#### browser

Interact with web pages.

```python
{
  "name": "browser",
  "arguments": {
    "code": "goto('https://example.com')\nclick('button#submit')"
  }
}
```

#### web_read

Read web page content.

```python
{
  "name": "web_read",
  "arguments": {
    "url": "https://example.com"
  }
}
```

#### think

Record agent thoughts.

```python
{
  "name": "think",
  "arguments": {
    "thought": "I need to analyze the error..."
  }
}
```

#### finish

Signal task completion.

```python
{
  "name": "finish",
  "arguments": {
    "message": "Task completed successfully"
  }
}
```

---

## Makefile Commands

| Command | Description |
|---------|-------------|
| `make build` | Full build (Python + frontend + hooks) |
| `make run` | Start backend and frontend |
| `make start-backend` | Start backend only |
| `make start-frontend` | Start frontend only |
| `make setup-config` | Interactive config setup |
| `make test` | Run all tests |
| `make lint` | Run linting |
| `make docker-build` | Build Docker image |
| `make docker-run` | Run with Docker |
| `make help` | Show all commands |

---

## Docker Commands

### Build

```bash
# Build Docker image
docker build -t openhands:latest -f containers/app/Dockerfile .

# Build with specific version
docker build --build-arg OPENHANDS_BUILD_VERSION=1.4.0 -t openhands:1.4.0 .
```

### Run

```bash
# Basic run
docker run -it --rm \
  -p 3000:3000 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v ~/.openhands:/.openhands \
  -v $(pwd)/workspace:/opt/workspace_base \
  -e LLM_API_KEY="your-key" \
  -e LLM_MODEL="gpt-4o" \
  openhands:latest

# With Docker Compose
docker-compose up -d
```

---

## Python API

### Programmatic Usage

```python
import asyncio
from openhands.core.main import run_controller
from openhands.core.config import load_openhands_config
from openhands.events.action import MessageAction

async def main():
    # Load configuration
    config = load_openhands_config()
    
    # Create initial action
    action = MessageAction(content="Write a hello world program")
    
    # Run controller
    state = await run_controller(
        config=config,
        initial_user_action=action,
        exit_on_message=False
    )
    
    print(f"Final state: {state.agent_state}")

asyncio.run(main())
```

### Configuration Loading

```python
from openhands.core.config import load_app_config

# Load from all sources (TOML, env, defaults)
config = load_app_config()

# Access configuration
llm_config = config.get_llm_config()
agent_config = config.get_agent_config()
sandbox_config = config.sandbox

print(f"Model: {llm_config.model}")
print(f"Agent: {config.default_agent}")
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | General error |
| `2` | Invalid arguments |
| `130` | Interrupted (Ctrl+C) |

---

## Related Documentation

- [Architecture](./ARCHITECTURE.md) - System design
- [Usage Guide](./USAGE.md) - Practical examples
