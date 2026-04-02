# Forge - API Reference

## Command Line Interface

### Global Options

```bash
forge [OPTIONS] [COMMAND]
```

| Option | Short | Description | Example |
|--------|-------|-------------|---------|
| `--prompt` | `-p` | Direct prompt (non-interactive) | `forge -p "explain this code"` |
| `--command` | `-c` | Path to command file | `forge -c ./commands.md` |
| `--workflow` | `-w` | Path to workflow file | `forge -w ./workflow.yaml` |
| `--event` | `-e` | Dispatch event to workflow | `forge -e "task:complete"` |
| `--conversation` | | Path to conversation file | `forge --conversation ./conv.json` |
| `--model` | `-m` | Specify model | `forge -m claude-sonnet-4` |
| `--restricted` | `-r` | Enable restricted mode | `forge -r` |
| `--verbose` | | Enable verbose output | `forge --verbose` |
| `--help` | `-h` | Show help | `forge --help` |
| `--version` | `-V` | Show version | `forge --version` |

### Subcommands

#### Provider Management

```bash
forge provider <COMMAND>
```

| Command | Description | Example |
|---------|-------------|---------|
| `login` | Add/update provider credentials | `forge provider login` |
| `logout` | Remove provider credentials | `forge provider logout` |
| `list` | List supported providers | `forge provider list` |

#### Agent Commands

```bash
forge agent <AGENT_NAME> [PROMPT]
```

| Agent | Description | Example |
|-------|-------------|---------|
| `forge` | Default implementation agent | `forge agent forge` |
| `sage` | Research and analysis agent | `forge agent sage` |
| `muse` | Planning and strategy agent | `forge agent muse` |

#### MCP Commands

```bash
forge mcp <COMMAND>
```

| Command | Description | Example |
|---------|-------------|---------|
| `list` | List configured MCP servers | `forge mcp list` |
| `add` | Add MCP server interactively | `forge mcp add` |
| `add-json` | Add MCP server from JSON | `forge mcp add-json` |
| `get` | Get server details | `forge mcp get <name>` |
| `remove` | Remove MCP server | `forge mcp remove <name>` |

#### Model Commands

```bash
forge model [COMMAND]
```

| Command | Description | Example |
|---------|-------------|---------|
| (none) | Interactive model selection | `forge model` |
| `list` | List available models | `forge model list` |

#### Utility Commands

```bash
forge <COMMAND>
```

| Command | Description | Example |
|---------|-------------|---------|
| `zsh-setup` | Configure Zsh plugin | `forge zsh-setup` |
| `zsh-uninstall` | Remove Zsh plugin | `forge zsh-uninstall` |
| `update` | Check for updates | `forge update` |

---

## Configuration Reference

### forge.yaml Schema

```yaml
# yaml-language-server: $schema=./forge.schema.json

# Model Configuration
model: "anthropic/claude-sonnet-4"
session:
  model_id: "claude-sonnet-4"
  provider_id: "anthropic"
commit:
  model_id: "claude-sonnet-4"
  provider_id: "anthropic"
suggest:
  model_id: "gpt-4o-mini"
  provider_id: "openai"

# Generation Parameters
temperature: 0.7
top_p: 0.8
top_k: 30
max_tokens: 20480
max_requests_per_turn: 100
max_tool_failure_per_turn: 3

# Tool Settings
tool_supported: true
tool_timeout_secs: 300
max_file_size_bytes: 10485760
max_image_size_bytes: 10485760
max_line_chars: 2000
max_read_lines: 1000
max_stdout_line_chars: 2000

# File Operations
max_file_read_batch_size: 10
max_parallel_file_reads: 5
max_walker_depth: 10
max_extensions: 100

# Search Settings
max_search_lines: 100
max_search_result_bytes: 10240
max_sem_search_results: 200
sem_search_top_k: 20

# Context Compaction
compact:
  max_tokens: 2000
  token_threshold: 100000
  message_threshold: 200
  retention_window: 6
  eviction_window: 0.2
  turn_threshold: 50
  on_turn_end: false
  model: null

# Reasoning Configuration
reasoning:
  enabled: true
  effort: "medium"  # none, minimal, low, medium, high, xhigh, max
  max_tokens: null
  exclude: false

# Update Settings
updates:
  frequency: "daily"  # daily, weekly, always
  auto_update: false

# Retry Configuration
retry:
  initial_backoff_ms: 1000
  min_delay_ms: 100
  backoff_factor: 2
  max_attempts: 3
  max_delay_secs: 60
  status_codes: [429, 500, 502, 503, 504]
  suppress_errors: false

# HTTP Configuration
http:
  connect_timeout_secs: 30
  read_timeout_secs: 900
  pool_idle_timeout_secs: 90
  pool_max_idle_per_host: 5
  max_redirects: 10
  hickory: false
  tls_backend: "default"  # default, rustls
  min_tls_version: null   # "1.0", "1.1", "1.2", "1.3"
  max_tls_version: null
  adaptive_window: true
  keep_alive_interval_secs: 60
  keep_alive_timeout_secs: 10
  keep_alive_while_idle: true
  accept_invalid_certs: false
  root_cert_paths: null

# Session Settings
max_conversations: 100
custom_history_path: null
restricted: false
auto_dump: null  # "json" or "html"
auto_open_dump: false
debug_requests: null

# Services
services_url: "https://api.forgecode.dev"
model_cache_ttl_secs: 3600

# Custom Rules
custom_rules: |
  1. Always add error handling to new code
  2. Include tests for all functions
  3. Follow the team's style guide

# Custom Commands
commands:
  - name: "test"
    description: "Run the test suite"
    prompt: "Run all tests and report results"
  - name: "lint"
    description: "Run linter"
    prompt: "Run the linter and fix any issues"
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `FORGE_LOG` | Logging level filter | `forge=info` |
| `FORGE_TRACKER` | Enable telemetry tracking | `true` |
| `FORGE_API_URL` | Custom Forge API URL | `https://api.forgecode.dev` |
| `FORGE_WORKSPACE_SERVER_URL` | Indexing server URL | `https://api.forgecode.dev/` |
| `FORGE_RETRY_INITIAL_BACKOFF_MS` | Initial retry backoff | `1000` |
| `FORGE_RETRY_BACKOFF_FACTOR` | Retry backoff multiplier | `2` |
| `FORGE_RETRY_MAX_ATTEMPTS` | Maximum retry attempts | `3` |
| `FORGE_HTTP_CONNECT_TIMEOUT` | Connection timeout (secs) | `30` |
| `FORGE_HTTP_READ_TIMEOUT` | Read timeout (secs) | `900` |
| `FORGE_TOOL_TIMEOUT` | Tool execution timeout (secs) | `300` |
| `FORGE_MAX_IMAGE_SIZE` | Max image size (bytes) | `10485760` |
| `FORGE_DUMP_AUTO_OPEN` | Auto-open dump files | `false` |
| `FORGE_DEBUG_REQUESTS` | Debug request log path | - |
| `FORGE_SEM_SEARCH_LIMIT` | Semantic search limit | `200` |
| `FORGE_SEM_SEARCH_TOP_K` | Semantic search top-k | `20` |
| `FORGE_MAX_SEARCH_RESULT_BYTES` | Max search result bytes | `10240` |
| `FORGE_HISTORY_FILE` | Custom history file path | - |
| `FORGE_BANNER` | Custom startup banner | - |
| `FORGE_MAX_CONVERSATIONS` | Max conversations in list | `100` |
| `FORGE_MAX_LINE_LENGTH` | Max chars per line | `2000` |
| `FORGE_STDOUT_MAX_LINE_LENGTH` | Max stdout line length | `2000` |
| `FORGE_CURRENCY_SYMBOL` | Currency symbol for costs | `$` |
| `FORGE_CURRENCY_CONVERSION_RATE` | Currency conversion rate | `1.0` |
| `NERD_FONT` | Enable Nerd Font icons | auto-detect |
| `FORGE_BIN` | Forge binary command | `forge` |

---

## Tool Reference

### File Operations

#### read

Read file contents with optional line range.

```json
{
  "path": "string",
  "offset": 0,
  "limit": 100
}
```

**Examples:**
```json
{"path": "src/main.rs"}
{"path": "config.yaml", "offset": 10, "limit": 20}
```

#### write

Create or overwrite a file.

```json
{
  "path": "string",
  "content": "string"
}
```

#### patch

Apply edits to an existing file using search/replace.

```json
{
  "path": "string",
  "search": "string",
  "replace": "string"
}
```

#### remove

Delete a file or directory.

```json
{
  "path": "string"
}
```

### Search Operations

#### fs_search

Regex-based file content search.

```json
{
  "pattern": "string",
  "path": ".",
  "include": "*.rs",
  "exclude": "target/"
}
```

#### sem_search

Semantic code search using embeddings.

```json
{
  "query": "string",
  "path": ".",
  "limit": 10
}
```

### Shell Operations

#### shell

Execute shell commands.

```json
{
  "command": "string",
  "cwd": ".",
  "timeout": 300
}
```

**Example:**
```json
{"command": "cargo build", "timeout": 120}
```

### Network Operations

#### fetch

Fetch content from URLs.

```json
{
  "url": "string",
  "max_chars": 10000
}
```

### Agent Operations

#### sage

Invoke the research agent.

```json
{
  "query": "string",
  "path": "."
}
```

### Utility Operations

#### skill

Load a skill context.

```json
{
  "name": "string"
}
```

#### undo

Revert the last file operation.

```json
{}
```

---

## Agent Definition Format

### YAML Frontmatter

```yaml
---
# Required fields
id: "agent-id"
title: "Agent Display Name"
description: "Detailed description of agent purpose"
reasoning:
  enabled: true
tools:
  - tool1
  - tool2
  - mcp_*
user_prompt: |-
  <{{event.name}}>{{event.value}}</{{event.name}}>
  <system_date>{{current_date}}</system_date>
---
```

### Field Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | Yes | Unique identifier (lowercase, hyphens) |
| `title` | string | Yes | Display name |
| `description` | string | Yes | Detailed purpose description |
| `reasoning` | object | Yes | `{ enabled: boolean }` |
| `tools` | array | Yes | List of allowed tools |
| `user_prompt` | string | Yes | Template for user context |

### Available Tools

| Tool | Description |
|------|-------------|
| `read` | Read file contents |
| `write` | Create/overwrite files |
| `patch` | Edit existing files |
| `remove` | Delete files/directories |
| `shell` | Execute shell commands |
| `fs_search` | Regex file search |
| `sem_search` | Semantic code search |
| `fetch` | Fetch URL content |
| `sage` | Invoke research agent |
| `skill` | Load skill context |
| `undo` | Revert changes |
| `mcp_*` | All MCP tools |
| `mcp_<name>` | Specific MCP tool |

---

## MCP Configuration

### .mcp.json Schema

```json
{
  "mcpServers": {
    "server_name": {
      "command": "command_to_execute",
      "args": ["arg1", "arg2"],
      "env": {
        "ENV_VAR": "value"
      }
    },
    "http_server": {
      "url": "http://localhost:3000/events"
    }
  }
}
```

### Configuration Locations

| Location | Scope | Priority |
|----------|-------|----------|
| `./.mcp.json` | Project-specific | Highest |
| `~/.config/forge/.mcp.json` | User global | Lower |

### Common MCP Servers

| Server | Install Command | Purpose |
|--------|-----------------|---------|
| GitHub | `npx -y @github/mcp-server` | Repository management |
| PostgreSQL | `npx -y @modelcontextprotocol/server-postgres` | Database access |
| SQLite | `npx -y @modelcontextprotocol/server-sqlite` | SQLite operations |
| Brave Search | `npx -y @modelcontextprotocol/server-brave-search` | Web search |
| Filesystem | `npx -y @modelcontextprotocol/server-filesystem` | File operations |

---

## Skills System

### Skill Definition

Skills are reusable capabilities defined in `.forge/skills/`:

```markdown
---
name: skill-name
description: What this skill does
---

Skill instructions here...
```

### Loading Skills

```bash
# In agent interaction
> Use skill create-agent to make a new agent

# Or via tool call
skill({"name": "create-agent"})
```

---

## Zsh Plugin Commands

When the Zsh plugin is active, use `:` as a trigger:

```bash
# Quick prompt
: explain this codebase

# List available commands
: <Tab>

# Use specific command
: commit
: test
: lint
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | General error |
| `2` | Invalid arguments |
| `3` | Provider API error |
| `4` | Authentication error |
| `5` | Tool execution failed |
| `130` | Interrupted (Ctrl+C) |

---

## Related Documentation

- [Architecture](./ARCHITECTURE.md) - System design
- [Usage Guide](./USAGE.md) - Practical examples
- [External References](./REFERENCES.md) - Tutorials
- [Diagrams](./DIAGRAMS.md) - Visual documentation
