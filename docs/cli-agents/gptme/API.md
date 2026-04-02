# GPTMe - API Reference

## Command Line Interface

### Main Command: gptme

```bash
gptme [OPTIONS] [PROMPTS]...
```

### Global Options

| Option | Short | Description | Example |
|--------|-------|-------------|---------|
| `--name` | | Conversation name | `--name my-chat` |
| `--model` | `-m` | Model to use | `-m anthropic/claude-sonnet-4-6` |
| `--workspace` | `-w` | Workspace directory | `-w ./project` |
| `--agent-path` | | Agent workspace directory | `--agent-path ~/my-agent` |
| `--resume` | `-r` | Resume most recent conversation | `-r` |
| `--no-confirm` | `-y` | Skip all confirmations | `-y` |
| `--non-interactive` | `-n` | Non-interactive mode | `-n` |
| `--system` | | System prompt type | `--system short` |
| `--tools` | `-t` | Allowed tools | `-t shell,python,read` |
| `--tool-format` | | Tool format | `--tool-format xml` |
| `--no-stream` | | Disable streaming | `--no-stream` |
| `--show-hidden` | | Show system messages | `--show-hidden` |
| `--verbose` | `-v` | Verbose output | `-v` |
| `--version` | | Show version | `--version` |
| `--help` | `-h` | Show help | `--help` |

### Tool Allowlist Syntax

```bash
# Only specific tools
gptme -t shell,python,read "analyze this"

# Add tools to default set
gptme -t +subagent,+browser "research this"

# Remove tools from default set
gptme -t=-browser "work offline"

# Mixed operations
gptme -t=-browser,+subagent "complex task"
```

### Multi-Prompt Separator

Chain multiple prompts with `-`:
```bash
gptme 'make a change' - 'test it' - 'commit it'
```

---

## Utility Commands

### gptme-util

```bash
gptme-util [COMMAND] [OPTIONS]
```

**Available Commands:**

| Command | Description |
|---------|-------------|
| `tools list` | List all tools and availability |
| `tools info TOOL` | Show detailed tool information |
| `chats list` | List past conversations |
| `chats search QUERY` | Search conversations |
| `chats rename OLD NEW` | Rename a conversation |
| `models list` | List available models |
| `context index` | Index project files for RAG |
| `llm generate` | Direct LLM generation |

### gptme-server

```bash
gptme-server [OPTIONS]
```

| Option | Description | Default |
|--------|-------------|---------|
| `--host` | Host to bind to | `127.0.0.1` |
| `--port` | Port to listen on | `5000` |
| `--debug` | Enable debug mode | `False` |

### gptme-agent

```bash
gptme-agent [COMMAND] [OPTIONS]
```

**Commands:**

| Command | Description |
|---------|-------------|
| `create PATH` | Create new agent |
| `install` | Install agent (systemd/launchd) |
| `uninstall` | Uninstall agent |
| `status` | Check agent status |
| `run` | Run agent once |

### gptme-auth

```bash
gptme-auth [OPTIONS]
```

Manage authentication for LLM providers.

### gptme-doctor

```bash
gptme-doctor [OPTIONS]
```

Check system configuration and dependencies.

### gptme-eval

```bash
gptme-eval [OPTIONS] [SUITES...]
```

Run evaluation suites to test capabilities.

---

## Slash Commands

During a conversation, use these commands:

| Command | Description |
|---------|-------------|
| `/undo` | Undo the last action |
| `/log` | Show conversation log |
| `/edit` | Edit conversation in editor |
| `/rename` | Rename conversation |
| `/fork` | Create copy of conversation |
| `/summarize` | Summarize conversation |
| `/replay` | Replay tool operations |
| `/export` | Export as HTML |
| `/model` | Show/switch model |
| `/models` | List available models |
| `/tokens` | Show token usage |
| `/context` | Show context breakdown |
| `/tools` | Show available tools |
| `/commit` | Ask assistant to git commit |
| `/compact` | Compact conversation |
| `/impersonate` | Impersonate assistant |
| `/restart` | Restart gptme process |
| `/setup` | Run setup wizard |
| `/help` | Show help |
| `/exit` | Exit program |

### Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl+X Ctrl+E` | Edit prompt in editor |
| `Ctrl+J` | Insert new line |
| `↑/↓` | Command history |
| `Tab` | Autocomplete |

---

## Configuration Reference

### Global Config (~/.config/gptme/config.toml)

```toml
[user]
name = "User"
about = "I am a curious human programmer."
response_preference = "Don't explain basic concepts"
avatar = "~/Pictures/avatar.jpg"

[prompt]
files = ["~/notes/llm-tips.md"]

[prompt.project]
myproject = "A description of my project."

[env]
MODEL = "anthropic/claude-sonnet-4-6"
ANTHROPIC_API_KEY = "sk-ant-..."
OPENAI_API_KEY = "sk-..."
OPENROUTER_API_KEY = "sk-or-..."
GEMINI_API_KEY = "..."
XAI_API_KEY = "..."
GROQ_API_KEY = "..."
DEEPSEEK_API_KEY = "..."

# Tool configuration
TOOL_FORMAT = "markdown"  # markdown, xml, tool
TOOL_ALLOWLIST = "save,append,patch,ipython,shell,browser"

# MCP servers
[[mcp.servers]]
name = "filesystem"
command = "npx"
args = ["-y", "@modelcontextprotocol/server-filesystem", "/home/user"]
```

### Project Config (gptme.toml)

```toml
files = ["README.md", "Makefile"]
prompt = "This is gptme."
base_prompt = "You are a coding assistant..."
context_cmd = "git status -v"

[rag]
enabled = true
index_path = ".gptme-rag"

[plugins]
paths = ["./plugins", "~/.config/gptme/plugins"]
enabled = ["my_plugin"]

[agent]
name = "Bob"
avatar = "assets/avatar.png"

[env]
CUSTOM_VAR = "value"

[[mcp.servers]]
name = "my-server"
command = "python"
args = ["-m", "my_mcp_server"]
env = { API_KEY = "secret" }
```

### Chat Config (~/.local/share/gptme/logs/<chat>/config.toml)

Automatically generated per conversation:

```toml
model = "anthropic/claude-sonnet-4-6"
tools = ["shell", "python", "read", "save", "patch"]
tool_format = "markdown"
stream = true
interactive = true
workspace = "/path/to/workspace"
```

---

## Environment Variables

### Feature Flags

| Variable | Description | Default |
|----------|-------------|---------|
| `GPTME_CHECK` | Enable pre-commit checks | `true` (if `.pre-commit-config.yaml`) |
| `GPTME_CHAT_HISTORY` | Cross-conversation context | `false` |
| `GPTME_COSTS` | Enable cost reporting | `false` |
| `GPTME_FRESH` | Fresh context mode | `false` |
| `GPTME_BREAK_ON_TOOLUSE` | Interrupt on tool use | Model-dependent |
| `GPTME_PATCH_RECOVERY` | Return file content on patch fail | `false` |
| `GPTME_SUGGEST_LLM` | LLM-powered prompt completion | `false` |
| `GPTME_TOOL_SOUNDS` | Enable tool sounds | `false` |

### API Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `LLM_API_TIMEOUT` | LLM request timeout (seconds) | `600` |

### Paths

| Variable | Description |
|----------|-------------|
| `GPTME_LOGS_HOME` | Override logs folder location |

### CLI Options via Environment

All CLI options can be set via `GPTME_*` prefixed variables:

- `GPTME_MODEL` - Set default model
- `GPTME_TOOL_FORMAT` - Set tool format
- `GPTME_WORKSPACE` - Set workspace
- `GPTME_TOOL_ALLOWLIST` - Set allowed tools

---

## Model Reference

### Provider/Model Format

```
<provider>/<model-name>
```

Examples:
- `anthropic/claude-sonnet-4-6`
- `openai/gpt-4o`
- `openrouter/anthropic/claude-3-opus`
- `local/llama3.1`

### Supported Providers

| Provider | Environment Variable | Notes |
|----------|---------------------|-------|
| Anthropic | `ANTHROPIC_API_KEY` | Claude models |
| OpenAI | `OPENAI_API_KEY` | GPT models |
| OpenRouter | `OPENROUTER_API_KEY` | 100+ models |
| Google | `GEMINI_API_KEY` | Gemini models |
| xAI | `XAI_API_KEY` | Grok models |
| DeepSeek | `DEEPSEEK_API_KEY` | DeepSeek models |
| Groq | `GROQ_API_KEY` | Fast inference |
| Local | None | llama.cpp, etc. |

---

## Tool Reference

### Tool Formats

| Format | Description |
|--------|-------------|
| `markdown` | Markdown code blocks (default) |
| `xml` | XML-style tags |
| `tool` | Function calling format |

### Tool Examples

**Markdown Format (default):**
```
▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔
Saving code to fib.py
▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔

```python
def fib(n):
    if n <= 1:
        return n
    return fib(n-1) + fib(n-2)
```
```

**XML Format:**
```xml
<tool>
  <name>save</name>
  <path>fib.py</path>
  <content>
def fib(n):
    if n <= 1:
        return n
    return fib(n-1) + fib(n-2)
  </content>
</tool>
```

---

## Plugin Configuration

### Plugin Structure

```python
# my_plugin/__init__.py
from gptme.tools import ToolSpec
from gptme.hooks import hook

# Define a tool
my_tool = ToolSpec(
    name="my_tool",
    desc="My custom tool",
    instructions="Use this tool to...",
    examples="",
    execute=lambda code, **kwargs: "Result",
)

# Define a hook
@hook("pre_tool_execute")
def my_hook(tool, **kwargs):
    print(f"Executing: {tool.name}")
    return kwargs
```

### Entry Points (pyproject.toml)

```toml
[project.entry-points."gptme.plugins"]
my_plugin = "my_plugin"
```

---

## MCP Server Configuration

### Example Config

```toml
[[mcp.servers]]
name = "filesystem"
command = "npx"
args = ["-y", "@modelcontextprotocol/server-filesystem", "/home/user"]
auto_start = true

[[mcp.servers]]
name = "postgres"
command = "npx"
args = ["-y", "@modelcontextprotocol/server-postgres", "postgresql://localhost/db"]
env = { DATABASE_URL = "postgresql://localhost/db" }
```

### Transport Types

| Type | Description |
|------|-------------|
| `stdio` | Local process communication |
| `http` | HTTP REST API |
| `sse` | Server-sent events |

---

## REST API

### Endpoints

#### Conversations

```
GET    /api/v2/conversations        # List conversations
POST   /api/v2/conversations/new    # Create conversation
GET    /api/v2/conversation/{id}    # Get conversation
POST   /api/v2/conversation/{id}/step  # Execute step
```

#### Models & Tools

```
GET    /api/v2/models               # List available models
GET    /api/v2/tools                # List available tools
```

#### MCP

```
GET    /api/v2/mcp/servers          # List MCP servers
POST   /api/v2/mcp/servers/{id}/start   # Start MCP server
POST   /api/v2/mcp/servers/{id}/stop    # Stop MCP server
```

### Response Format

```json
{
  "conversation_id": "uuid",
  "messages": [
    {
      "role": "user",
      "content": "Hello",
      "timestamp": "2025-01-01T00:00:00Z"
    }
  ],
  "tools_used": [],
  "tokens_used": 150
}
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | General error |
| `2` | Keyboard interrupt |
| `3` | Configuration error |
| `4` | API error |
| `5` | Tool execution failed |

---

## Related Documentation

- [Architecture](./ARCHITECTURE.md) - System design
- [Usage Guide](./USAGE.md) - Practical examples
- [Plugin Development](./plugins/README.md) - Creating plugins
