# GPTMe - Architecture Documentation

## System Overview

GPTMe is a Python-based personal AI agent that combines a terminal interface with LLM capabilities to provide an interactive coding and task automation experience. It enables AI assistants to execute code, edit files, browse the web, process images, and interact with your computer through a rich set of tools.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            GPTME ARCHITECTURE                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────────┐      ┌──────────────────┐      ┌──────────────────┐   │
│  │   Terminal UI    │◄────►│   GPTMe Core     │◄────►│  LLM Providers   │   │
│  │   (Rich/Click)   │      │   Engine         │      │  (10+ Providers) │   │
│  └──────────────────┘      └────────┬─────────┘      └──────────────────┘   │
│                                     │                                        │
│           ┌─────────────────────────┼─────────────────────────┐              │
│           │                         │                         │              │
│           ▼                         ▼                         ▼              │
│  ┌──────────────────┐    ┌──────────────────┐    ┌──────────────────┐       │
│  │  Plugin System   │    │   Tool System    │    │  Session Manager │       │
│  │  (Python pkgs)   │    │  (14+ Tools)     │    │  (LogManager)    │       │
│  └──────────────────┘    └──────────────────┘    └──────────────────┘       │
│                                                                              │
│           ┌──────────────────┐    ┌──────────────────┐                       │
│           │   MCP Client     │    │   Web Server     │                       │
│           │  (External Tools)│    │  (REST API/WS)   │                       │
│           └──────────────────┘    └──────────────────┘                       │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Core Components

### 1. Terminal UI Layer

**Technology Stack:**
- **CLI Framework**: Click for argument parsing and command structure
- **UI Library**: Rich for terminal rendering, syntax highlighting, and formatting
- **Input Handling**: Prompt Toolkit for interactive prompts with history and completion
- **Features**:
  - Real-time streaming responses from LLMs
  - Syntax highlighting for code blocks
  - Command history and tab completion
  - Diff display for file changes
  - Interactive tool confirmation dialogs
  - Sound notifications (optional)

**Key UI Elements:**
```
┌──────────────────────────────────────────────────────┐
│ gptme v0.31.0 - using anthropic/claude-sonnet-4-6   │
├──────────────────────────────────────────────────────┤
│                                                      │
│ > User: Create a fibonacci function                 │
│                                                      │
│ Assistant: I'll create a fibonacci function for you.│
│                                                      │
│ ▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔ │
│ Saving fibonacci to fib.py                          │
│ ▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔ │
│                                                      │
│ Execute shell command?                              │
│ $ python fib.py                                     │
│ [Y]es / [n]o / [N]ever                              │
│                                                      │
│ > _                                                  │
│                                                      │
└──────────────────────────────────────────────────────┘
```

**Implementation Details:**
- Located in `gptme/cli/main.py`
- Uses Click for CLI argument parsing with custom parameter types
- Rich console for formatted output and progress indicators
- Handles both interactive and non-interactive modes
- Supports piped input and output redirection

### 2. Core Engine

The central orchestrator located in `gptme/chat.py` manages:
- **LLM Communication**: Streaming HTTP requests to providers with retry logic
- **Context Management**: Assembling conversation history and file context
- **Tool Execution**: Parsing, confirming, and executing tool calls
- **Message Processing**: Formatting and parsing messages between user, assistant, and tools
- **Session State**: Persistence, resume functionality, and conversation management

**Chat Loop Flow:**
```python
while True:
    1. Get user input
    2. Build context (system prompt + history + files)
    3. Send to LLM with available tools
    4. Stream response
    5. Parse tool calls from response
    6. Execute tools with confirmation
    7. Return results to LLM
    8. Continue until no more tool calls
```

### 3. Tool System

**Tool Architecture:**
```
┌─────────────────────────────────────────────────────────────┐
│                       Tool System                           │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ToolSpec (Base Class)                                       │
│  ├── name: str              - Tool identifier                │
│  ├── desc: str              - Short description              │
│  ├── instructions: str      - Usage instructions for LLM     │
│  ├── examples: str          - Example usage                  │
│  ├── execute: Callable      - Execution function             │
│  ├── init: Optional[Callable] - Initialization function      │
│  ├── blocklist: Optional[BlockList] - Blocked patterns       │
│  └── confirmations: Optional[bool] - Require confirmation    │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Built-in Tools:**

| Tool | Purpose | File | Confirmation |
|------|---------|------|--------------|
| `shell` | Execute shell commands | `tools/shell.py` | Yes |
| `ipython` | Run Python code interactively | `tools/python.py` | Yes |
| `read` | Read file contents | `tools/read.py` | No |
| `save` | Create new files | `tools/save.py` | Yes |
| `patch` | Edit existing files | `tools/patch.py` | Yes |
| `browser` | Browse websites via Playwright | `tools/browser.py` | Yes |
| `vision` | Analyze images | `tools/vision.py` | No |
| `screenshot` | Capture screen | `tools/screenshot.py` | Yes |
| `tmux` | Terminal session management | `tools/tmux.py` | No |
| `subagent` | Spawn sub-agents | `tools/subagent/` | Yes |
| `rag` | Semantic search | `tools/rag.py` | No |
| `gh` | GitHub CLI integration | `tools/gh.py` | Yes |
| `computer` | Desktop control | `tools/computer.py` | Yes |
| `chats` | Search conversations | `tools/chats.py` | No |

**Tool Execution Flow:**
```
User Request → Tool Selection → Confirmation Check → Execute → Return Result
                    │                                        │
                    ▼                                        ▼
            ┌─────────────┐                          ┌─────────────┐
            │  Pre-Tool   │                          │  Post-Tool  │
            │   Hooks     │                          │   Hooks     │
            └─────────────┘                          └─────────────┘
```

### 4. Configuration System

**Configuration Hierarchy:**
```
Configuration Precedence (lowest to highest):

1. Default Settings
   └─► Built into GPTMe

2. Global User Config
   └─► ~/.config/gptme/config.toml

3. Global Local Config
   └─► ~/.config/gptme/config.local.toml

4. Project Config
   └─► ./gptme.toml

5. Project Local Config
   └─► ./gptme.local.toml

6. Environment Variables
   └─► GPTME_* variables

7. CLI Arguments
   └─► --flag options

8. Chat Config
   └─► ~/.local/share/gptme/logs/<chat>/config.toml
```

**Config Classes:**
- `UserConfig` (`gptme/config/user.py`) - Global user preferences
- `ProjectConfig` (`gptme/config/project.py`) - Project-specific settings
- `ChatConfig` (`gptme/config/chat.py`) - Per-conversation settings
- `MCPConfig` (`gptme/config/models.py`) - MCP server configuration

**Key Configuration Files:**

**~/.config/gptme/config.toml:**
```toml
[user]
name = "User"
about = "I am a programmer"
response_preference = "Be concise"

[prompt]
files = ["~/notes/tips.md"]

[env]
MODEL = "anthropic/claude-sonnet-4-6"
ANTHROPIC_API_KEY = "sk-ant-..."
```

**./gptme.toml (Project Config):**
```toml
files = ["README.md", "Makefile"]
prompt = "This is a Python project"
context_cmd = "git status -v"

[rag]
enabled = true

[plugins]
enabled = ["my_plugin"]
```

### 5. LLM Provider System

**Supported Providers:**

| Provider | Module | Auth | Notes |
|----------|--------|------|-------|
| Anthropic | `llm/llm_anthropic.py` | API Key | Claude models |
| OpenAI | `llm/llm_openai.py` | API Key | GPT models |
| OpenRouter | `llm/llm_openai.py` | API Key | 100+ models |
| Google (Gemini) | `llm/llm_openai.py` | API Key | Gemini models |
| xAI (Grok) | `llm/llm_openai.py` | API Key | Grok models |
| DeepSeek | `llm/llm_openai.py` | API Key | DeepSeek models |
| Groq | `llm/llm_openai.py` | API Key | Fast inference |
| Local | `llm/llm_openai.py` | None | llama.cpp, etc. |

**Provider Resolution:**
```
Model String: "anthropic/claude-sonnet-4-6"
                    │              │
                    ▼              ▼
              Provider      Model Name
```

**Model Selection Priority:**
1. CLI `--model` argument
2. Environment variable `MODEL`
3. Config file `[env]` section
4. Provider-specific default

### 6. Session Management

**Session Lifecycle:**
```
┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐
│  Start  │───►│  Active │───►│  Save   │───►│ Resume  │───►│  End    │
│         │    │         │    │         │    │         │    │         │
└─────────┘    └─────────┘    └─────────┘    └─────────┘    └─────────┘
      │              │              │              │              │
      ▼              ▼              ▼              ▼              ▼
  Initialize   Transcript    Persist to    Load from    Cleanup
  Context      History       ~/.local/     ~/.local/    Resources
                              share/        share/
```

**LogManager:**
- Location: `~/.local/share/gptme/logs/`
- Format: `<YYYY-MM-DD>-<name>/`
- Files:
  - `conversation.jsonl` - Message history (append-only)
  - `config.toml` - Chat-specific configuration
  - `workspace/` - Working directory snapshot

**Conversation Structure:**
```json
{"role": "system", "content": "...", "timestamp": "2025-01-01T00:00:00Z"}
{"role": "user", "content": "Hello", "timestamp": "..."}
{"role": "assistant", "content": "Hi!", "timestamp": "..."}
{"role": "assistant", "tool_calls": [{"function": {"name": "shell", "arguments": "..."}}]}
{"role": "system", "name": "shell", "content": "output", "timestamp": "..."}
```

### 7. Plugin System

**Plugin Architecture:**
```
┌─────────────────────────────────────────────────────────────┐
│                     Plugin System                           │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Entry Points (pyproject.toml)                               │
│  └── gptme.plugins                                           │
│                                                              │
│  Plugin Module                                               │
│  ├── ToolSpec instances                                      │
│  ├── Hook functions                                          │
│  └── Command handlers                                        │
│                                                              │
│  Loading Flow:                                               │
│  1. Discover entry points                                    │
│  2. Import modules                                           │
│  3. Extract ToolSpec objects                                 │
│  4. Initialize and register                                  │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Creating a Plugin:**
```python
# my_plugin/__init__.py
from gptme.tools import ToolSpec

my_tool = ToolSpec(
    name="my_tool",
    desc="My custom tool",
    instructions="Use this tool to...",
    examples="Example usage...",
    execute=lambda code, **kwargs: f"Result: {code}",
)

# In pyproject.toml:
[project.entry-points."gptme.plugins"]
my_plugin = "my_package.my_plugin"
```

### 8. Hook System

**Hook Events:**

```
┌─────────────────────────────────────────────────────────────┐
│                      HOOK EVENTS                            │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  pre_prompt         ──► Before sending prompt to LLM        │
│       │                                                      │
│       └── Can modify prompt context                         │
│                                                              │
│  post_response      ──► After receiving LLM response        │
│       │                                                      │
│       └── Can modify response before display                │
│                                                              │
│  pre_tool_execute   ──► Before executing a tool             │
│       │                                                      │
│       ├── Can block execution                                │
│       ├── Can modify parameters                              │
│       └── Can add confirmation requirements                  │
│                                                              │
│  post_tool_execute  ──► After tool execution                │
│       │                                                      │
│       ├── Can modify output                                  │
│       └── Can trigger side effects                           │
│                                                              │
│  on_start           ──► When conversation starts            │
│                                                              │
│  on_exit            ──► When conversation ends              │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Hook Registration:**
```python
from gptme.hooks import hook

@hook("pre_tool_execute")
def my_pre_hook(tool, **kwargs):
    if tool.name == "shell":
        print(f"About to execute: {kwargs.get('code', '')}")
    return kwargs
```

### 9. MCP Integration

**MCP Architecture:**
```
┌─────────────────────────────────────────────────────────────┐
│                    MCP ARCHITECTURE                         │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  GPTMe ◄──► MCP Client ◄──► MCP Servers                    │
│                                                              │
│  Transport Types:                                            │
│  ├─► stdio (local processes)                                 │
│  ├─► HTTP (remote servers)                                   │
│  └─► SSE (server-sent events)                                │
│                                                              │
│  Configuration:                                              │
│  - ~/.config/gptme/config.toml                               │
│  - ./gptme.toml                                              │
│                                                              │
│  Dynamic Loading:                                            │
│  - Auto-discover MCP servers                                 │
│  - Load tools dynamically                                    │
│  - Support for custom transports                             │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**MCP Configuration Example:**
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
```

### 10. Server Architecture

**REST API Structure:**
```
┌─────────────────────────────────────────────────────────────┐
│                    GPTME SERVER                             │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Flask Application                                           │
│  ├── /api/v2/conversations      # List conversations        │
│  ├── /api/v2/conversations/new  # Create conversation       │
│  ├── /api/v2/conversation/{id}  # Get conversation          │
│  ├── /api/v2/conversation/{id}/step  # Execute step         │
│  ├── /api/v2/models             # List models               │
│  ├── /api/v2/tools              # List tools                │
│  └── /api/v2/mcp                # MCP management            │
│                                                              │
│  WebSocket Support:                                          │
│  └── Real-time streaming responses                          │
│                                                              │
│  Static Files:                                               │
│  └── Built-in web UI (gptme/server/static/)                 │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Server Components:**
- `gptme/server/app.py` - Flask application setup
- `gptme/server/api_v2.py` - API v2 endpoints
- `gptme/server/api_v2_sessions.py` - Session management
- `gptme/server/api_v2_agents.py` - Agent endpoints

---

## Data Flow

### Request-Response Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         REQUEST LIFECYCLE                               │
└─────────────────────────────────────────────────────────────────────────┘

1. USER INPUT
   └─► Natural language command or /command

2. CONTEXT ASSEMBLY
   ├─► Load system prompt with tool definitions
   ├─► Include user config (name, about, preferences)
   ├─► Add project context (gptme.toml, files)
   ├─► Load relevant lessons based on keywords
   ├─► Include conversation history
   └─► Apply context compression if needed

3. LLM REQUEST
   └─► Streaming request to provider
       ├─► Model: Per configuration
       ├─► Tools: Available tool definitions
       └─► Messages: Full conversation context

4. RESPONSE PROCESSING
   └─► Streaming text and tool calls
       ├─► Text: Display to user
       └─► Tool calls: Extract and execute

5. TOOL EXECUTION
   └─► If tools called:
       ├─► Check blocklists
       ├─► Run pre-tool hooks
       ├─► Get user approval if needed
       ├─► Execute tool
       ├─► Run post-tool hooks
       └─► Return result to LLM

6. FOLLOW-UP
   └─► LLM may make additional tool calls
   └─► Or provide final response
```

### Context Management

**Context Window Strategy:**
```
┌────────────────────────────────────────────────────────────┐
│                    CONTEXT WINDOW                          │
├────────────────────────────────────────────────────────────┤
│ System Prompt (~2K tokens)                                 │
│ ├─► Base assistant instructions                             │
│ ├─► Tool definitions                                        │
│ ├─► User preferences                                        │
│ └─► Project context                                         │
├────────────────────────────────────────────────────────────┤
│ Conversation History                                       │
│ ├─► User messages                                           │
│ ├─► Assistant responses                                     │
│ ├─► Tool calls and results                                  │
│ └─► System messages                                         │
├────────────────────────────────────────────────────────────┤
│ Auto-Compaction (when near limit)                          │
│ └─► Summarize older messages                                │
│ └─► Preserve recent context                                 │
└────────────────────────────────────────────────────────────┘
```

---

## Security Architecture

### Permission System

```
┌─────────────────────────────────────────────────────────────┐
│                 PERMISSION DECISION TREE                    │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Tool Request                                                │
│       │                                                      │
│       ▼                                                      │
│  ┌─────────────┐                                             │
│  │  Blocklist  │──Match──► Block tool                        │
│  │   Check     │                                             │
│  └─────────────┘                                             │
│       │ No match                                             │
│       ▼                                                      │
│  ┌─────────────┐                                             │
│  │ Pre-Tool    │──Block──► Cancel execution                   │
│  │   Hooks     │                                             │
│  └─────────────┘                                             │
│       │ Allow                                                │
│       ▼                                                      │
│  ┌─────────────┐                                             │
│  │ Interactive │──Yes──► Show confirmation prompt             │
│  │    Mode?    │                                             │
│  └─────────────┘                                             │
│       │ No (-y flag)                                         │
│       ▼                                                      │
│   Execute tool                                               │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Security Layers

1. **Tool Blocklists**: Each tool can define blocked patterns (e.g., shell tool blocks `rm -rf /`)
2. **Hook System**: Pre-tool hooks can block or modify execution
3. **Interactive Confirmation**: User approval for destructive operations
4. **Non-interactive Mode**: `-y` for auto-approval, `-n` for scripts
5. **Environment Isolation**: Tools run in user's environment

---

## Performance Considerations

### Optimization Strategies

1. **Context Compression**: Automatically summarize older messages when near token limit
2. **Lazy Loading**: Load tools and MCP servers on demand
3. **Streaming**: Real-time response display without waiting for full response
4. **Caching**: Conversation history cached in memory during session
5. **Parallel Agents**: Spawn subagents for concurrent tasks

### Resource Management

```
Memory Usage:
├─► Conversation history (configurable via context compression)
├─► Tool execution context
├─► MCP connection pool
└─► Plugin cache

CPU Usage:
├─► Terminal UI rendering (Rich)
├─► Hook script execution
├─► File system operations
└─► LLM tokenization (tiktoken)
```

---

## Module Dependencies

### Core Dependencies

```
gptme/
├── cli/              → click, rich, prompt-toolkit
├── tools/            → tool-specific (playwright, ipython, etc.)
├── llm/              → anthropic, openai, tiktoken
├── server/           → flask, flask-cors, pydantic
├── config/           → tomlkit, platformdirs
├── hooks/            → (internal)
├── plugins/          → importlib, pkgutil
└── mcp/              → mcp (official SDK)
```

### Optional Dependencies

| Extra | Packages | Use Case |
|-------|----------|----------|
| `browser` | playwright | Web browsing |
| `server` | flask, flask-cors | REST API |
| `datascience` | pandas, numpy, matplotlib | Data analysis |
| `sounds` | sounddevice, scipy | Audio notifications |
| `computer` | python-xlib | Desktop control |
| `acp` | agent-client-protocol | Editor integration |
| `telemetry` | opentelemetry-* | Observability |
| `eval` | dspy, datasets, swebench | Evaluation |

---

## Related Documentation

- [API Reference](./API.md) - Commands and configuration
- [Usage Guide](./USAGE.md) - Practical examples
- [External References](./REFERENCES.md) - Links and resources
- [Diagrams](./DIAGRAMS.md) - Visual documentation
- [Gap Analysis](./GAP_ANALYSIS.md) - Improvement opportunities
