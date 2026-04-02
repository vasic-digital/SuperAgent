# Forge - Architecture Documentation

## System Overview

Forge is built as a modular Rust workspace with clean architecture principles. The system follows a layered architecture with clear separation between domain logic, application services, and infrastructure concerns.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           FORGE SYSTEM ARCHITECTURE                          │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   ┌──────────────┐         ┌──────────────┐         ┌──────────────┐        │
│   │    USER      │◄───────►│  TERMINAL    │◄───────►│    FORGE     │        │
│   │              │         │     UI       │         │    ENGINE    │        │
│   └──────────────┘         └──────────────┘         └──────┬───────┘        │
│                                                            │                │
│                              ┌─────────────────────────────┼─────────┐      │
│                              │                             │         │      │
│                              ▼                             ▼         ▼      │
│   ┌──────────────┐    ┌──────────────┐    ┌──────────────┐ ┌──────────────┐ │
│   │   AGENTS     │    │    TOOLS     │    │   CONTEXT    │ │  LLM PROVIDER│ │
│   │  (Forge/     │    │  (Shell/     │    │  MANAGER     │ │   (300+      │ │
│   │   Sage/Muse) │    │   File/MCP)  │    │              │ │   models)    │ │
│   └──────────────┘    └──────────────┘    └──────────────┘ └──────────────┘ │
│                              │                                              │
│                              ▼                                              │
│   ┌──────────────┐    ┌──────────────┐    ┌──────────────┐                  │
│   │     MCP      │    │   FILESYSTEM │    │   SESSION    │                  │
│   │   SERVERS    │    │   (Project)  │    │   STORAGE    │                  │
│   └──────────────┘    └──────────────┘    └──────────────┘                  │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Workspace Structure

Forge uses a Cargo workspace with 20+ crates organized by responsibility:

### Core Crates

| Crate | Purpose | Key Components |
|-------|---------|----------------|
| **forge_main** | Application entry point | CLI parsing, UI initialization |
| **forge_domain** | Domain types and models | Agents, messages, tools, events |
| **forge_api** | Public API abstractions | API traits, service boundaries |
| **forge_app** | Application orchestration | Workflow execution, coordination |

### Service Crates

| Crate | Purpose | Key Components |
|-------|---------|----------------|
| **forge_services** | Business logic | Conversation service, tool orchestration |
| **forge_infra** | Infrastructure | HTTP clients, persistence |
| **forge_repo** | Repository layer | Conversation storage, file management |
| **forge_config** | Configuration | Config parsing, validation, defaults |

### Tool Crates

| Crate | Purpose | Key Components |
|-------|---------|----------------|
| **forge_fs** | File system tools | Read, write, patch, search |
| **forge_embed** | Embedding generation | Semantic search, vector storage |
| **forge_walker** | File system traversal | Directory walking, filtering |
| **forge_template** | Template engine | Prompt templating, rendering |

### UI Crates

| Crate | Purpose | Key Components |
|-------|---------|----------------|
| **forge_display** | Terminal display | Markdown rendering, syntax highlighting |
| **forge_stream** | Response streaming | Real-time output handling |
| **forge_spinner** | Progress indicators | Loading states, progress bars |
| **forge_markdown_stream** | Streaming markdown | Incremental markdown parsing |

### Utility Crates

| Crate | Purpose | Key Components |
|-------|---------|----------------|
| **forge_snaps** | Snapshot testing | Test fixtures, assertions |
| **forge_test_kit** | Testing utilities | Mock services, test helpers |
| **forge_tracker** | Telemetry | Usage tracking, analytics |
| **forge_json_repair** | JSON utilities | Repair malformed JSON |
| **forge_select** | Interactive selection | Fuzzy finder, prompts |
| **forge_ci** | CI/CD utilities | GitHub Actions integration |
| **forge_tool_macros** | Procedural macros | Tool definition macros |

---

## Domain Model

### Core Entities

```rust
// Simplified domain types from forge_domain

pub struct Agent {
    pub id: String,
    pub title: String,
    pub description: String,
    pub reasoning: ReasoningConfig,
    pub tools: Vec<String>,
    pub user_prompt: String,
    pub system_prompt: String,
}

pub struct Conversation {
    pub id: ConversationId,
    pub title: Option<String>,
    pub messages: Vec<Message>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

pub struct Message {
    pub role: Role,
    pub content: Content,
    pub tool_calls: Option<Vec<ToolCall>>,
}

pub enum Role {
    User,
    Assistant,
    System,
}
```

### Event System

Forge uses an event-driven architecture for tool execution and agent coordination:

```rust
pub enum Event {
    User(UserEvent),
    Assistant(AssistantEvent),
    Tool(ToolEvent),
    System(SystemEvent),
}

pub struct ToolEvent {
    pub tool_name: String,
    pub input: Value,
    pub output: Result<Value, ToolError>,
}
```

---

## Agent System

### Agent Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           AGENT SYSTEM                                       │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │                     AGENT REGISTRY                                   │   │
│   │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌────────────┐  │   │
│   │  │   Forge     │  │    Sage     │  │    Muse     │  │   Custom   │  │   │
│   │  │ (Implement) │  │  (Research) │  │   (Plan)    │  │   Agents   │  │   │
│   │  └─────────────┘  └─────────────┘  └─────────────┘  └────────────┘  │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                      │                                       │
│   ┌──────────────────────────────────┼──────────────────────────────────┐   │
│   │                                  ▼                                  │   │
│   │   ┌──────────────┐    ┌──────────────┐    ┌──────────────┐         │   │
│   │   │   SYSTEM     │    │   USER       │    │   TOOL       │         │   │
│   │   │   PROMPT     │    │   PROMPT     │    │   RESULTS    │         │   │
│   │   └──────────────┘    └──────────────┘    └──────────────┘         │   │
│   │                                                                     │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                      │                                       │
│                                      ▼                                       │
│                           LLM Provider API                                   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Agent Definition Format

Agents are defined in markdown files with YAML frontmatter:

```markdown
---
id: "forge"
title: "Perform technical development tasks"
description: "Hands-on implementation agent..."
reasoning:
  enabled: true
tools:
  - sem_search
  - read
  - write
  - patch
  - shell
  - mcp_*
user_prompt: |-
  <{{event.name}}>{{event.value}}</{{event.name}}>
  <system_date>{{current_date}}</system_date>
---

You are Forge, an expert software engineering assistant...

## Core Principles:
1. **Solution-Oriented**: Focus on providing effective solutions
2. **Professional Tone**: Maintain professional yet conversational tone
...
```

---

## Tool System

### Built-in Tools

| Tool | Purpose | Permission Level |
|------|---------|------------------|
| `read` | Read file contents | Auto-allow |
| `write` | Create new files | Requires approval |
| `patch` | Edit existing files | Requires approval |
| `remove` | Delete files | Requires approval |
| `shell` | Execute shell commands | Requires approval |
| `fs_search` | Regex search | Auto-allow |
| `sem_search` | Semantic code search | Auto-allow |
| `fetch` | Fetch URL content | Requires approval |
| `sage` | Invoke research agent | Auto-allow |
| `skill` | Load skill context | Auto-allow |
| `mcp_*` | MCP tool calls | Varies |

### Tool Execution Flow

```
User Request → Tool Selection → Permission Check → Execution → Result
                    │                                        │
                    ▼                                        ▼
            ┌─────────────┐                          ┌─────────────┐
            │   Tool      │                          │   Context   │
            │   Registry  │                          │   Update    │
            └─────────────┘                          └─────────────┘
```

### Tool Implementation Pattern

```rust
#[async_trait]
pub trait Tool: Send + Sync {
    fn name(&self) -> &str;
    fn description(&self) -> &str;
    fn parameters(&self) -> Value;
    
    async fn execute(&self, input: Value) -> Result<Value, ToolError>;
}
```

---

## Context Management

### Context Window Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         CONTEXT COMPOSITION                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  [System Context - ~2K tokens]                                               │
│  ├─ Base agent instructions                                                  │
│  ├─ Tool definitions                                                         │
│  ├─ AGENTS.md content (if present)                                           │
│  └─ Operating system / environment info                                      │
│                                                                              │
│  [Conversation History - ~150K tokens]                                       │
│  ├─ User messages                                                            │
│  ├─ Assistant responses                                                      │
│  ├─ Tool call results                                                        │
│  └─ File contents from read operations                                       │
│                                                                              │
│  [Current Turn - ~48K tokens]                                                │
│  ├─ Latest user message                                                      │
│  └─ Pending tool calls                                                       │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Context Compaction

When approaching token limits, Forge automatically compacts context:

```rust
pub struct CompactConfig {
    pub max_tokens: usize,        // Target token count after compaction
    pub token_threshold: usize,   // Trigger threshold
    pub retention_window: usize,  // Recent messages to preserve
    pub eviction_window: f32,     // % of context that can be summarized
    pub on_turn_end: bool,        // Compact after each turn
}
```

---

## LLM Provider Integration

### Provider Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                       PROVIDER ABSTRACTION LAYER                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │                     PROVIDER TRAIT                                   │   │
│   │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌────────────┐  │   │
│   │  │   chat()    │  │   stream()  │  │   models()  │  │  health()  │  │   │
│   │  └─────────────┘  └─────────────┘  └─────────────┘  └────────────┘  │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                      │                                       │
│   ┌──────────────────────────────────┼──────────────────────────────────┐   │
│   │                                  ▼                                  │   │
│   │   ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐  │   │
│   │   │Anthropic │ │  OpenAI  │ │OpenRouter│ │  Google  │ │  Custom  │  │   │
│   │   └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘  │   │
│   │                                                                     │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Supported Providers

| Provider | Models | Authentication |
|----------|--------|----------------|
| **Anthropic** | Claude 3.5/4 Sonnet, Opus | API Key |
| **OpenAI** | GPT-4, GPT-4o, o1, o3 | API Key |
| **OpenRouter** | 300+ models | API Key |
| **Google** | Gemini Pro, Flash | API Key / OAuth |
| **xAI** | Grok | API Key |
| **Cerebras** | Llama models | API Key |
| **Groq** | Mixtral, Llama | API Key |
| **Local** | Ollama, LM Studio | Local endpoint |

---

## MCP Integration

### Model Context Protocol Support

Forge implements MCP for external tool integration:

```json
// .mcp.json configuration
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@github/mcp-server"],
      "env": { "GITHUB_TOKEN": "..." }
    },
    "postgres": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-postgres"],
      "env": { "DATABASE_URL": "..." }
    }
  }
}
```

### MCP Tool Registration

MCP tools are dynamically registered with the `mcp_*` wildcard:

```yaml
# In agent definition
tools:
  - mcp_*          # All MCP tools
  - mcp_github_*   # GitHub-specific tools
  - mcp_postgres_query  # Specific tool
```

---

## Session Management

### Session Lifecycle

```
┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐
│  Start  │───►│  Active │───►│  Save   │───►│ Resume  │───►│  End    │
│         │    │         │    │         │    │         │    │         │
└─────────┘    └─────────┘    └─────────┘    └─────────┘    └─────────┘
     │              │              │              │              │
     ▼              ▼              ▼              ▼              ▼
 Load AGENTS   Tool calls    Persist to     Load from      Cleanup
 Create conv   File ops      ~/.forge/      ~/.forge/      Archive
```

### Storage Locations

| Platform | Path |
|----------|------|
| **macOS** | `~/Library/Application Support/forge/` |
| **Linux** | `~/.config/forge/` or `$XDG_CONFIG_HOME/forge/` |
| **Windows** | `%APPDATA%/forge/` |

---

## Security Model

### Permission Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         SECURITY LAYERS                                      │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  Layer 4: User Control                                                       │
│  ├─ Explicit approval for write operations                                 │
│  ├─ Confirmation for shell commands                                        │
│  └─ Configurable auto-approve patterns                                     │
│                                                                              │
│  Layer 3: Agent Restrictions                                               │
│  ├─ Tool allowlists per agent                                              │
│  ├─ Read-only agents (Sage)                                                │
│  └─ Planning-only agents (Muse)                                            │
│                                                                              │
│  Layer 2: Restricted Mode                                                  │
│  ├─ Sandbox environment                                                    │
│  ├─ Limited file system access                                             │
│  └─ Network restrictions                                                   │
│                                                                              │
│  Layer 1: Protected Paths                                                  │
│  ├─ ~/.ssh, ~/.aws, ~/.gnupg                                              │
│  ├─ System directories                                                     │
│  └─ Configurable protections                                               │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Secure Credential Storage

Provider credentials are stored using platform-native secure storage:

- **macOS**: Keychain
- **Linux**: Secret Service API / libsecret
- **Windows**: Credential Manager

---

## Performance Considerations

### Optimization Strategies

1. **Lazy Loading** - Components loaded on demand
2. **Connection Pooling** - HTTP client reuse
3. **Streaming** - Real-time response processing
4. **Parallel Tool Execution** - Concurrent tool calls
5. **Semantic Search Caching** - Embedding result caching

### Resource Management

| Resource | Strategy |
|----------|----------|
| **Memory** | Streaming responses, bounded channels |
| **Connections** | HTTP/2 multiplexing, connection pools |
| **Disk I/O** | Async file operations, buffering |
| **Tokens** | Context compaction, intelligent truncation |

---

## Build System

### Cross-Compilation

Forge uses `cross` for multi-platform builds:

```toml
# Cross.toml
[target.x86_64-unknown-linux-gnu]
image = "ghcr.io/cross-rs/x86_64-unknown-linux-gnu:latest"

[target.aarch64-unknown-linux-gnu]
image = "ghcr.io/cross-rs/aarch64-unknown-linux-gnu:latest"
```

### Release Profile

```toml
[profile.release]
lto = true              # Link-time optimization
codegen-units = 1       # Single codegen unit for max optimization
opt-level = 3           # Maximum optimization
strip = true            # Strip symbols for smaller binary
```

---

## Related Documentation

- [API Reference](./API.md) - Commands and configuration
- [Usage Guide](./USAGE.md) - Practical examples
- [External References](./REFERENCES.md) - Tutorials and resources
- [Diagrams](./DIAGRAMS.md) - Visual documentation
- [Gap Analysis](./GAP_ANALYSIS.md) - Improvement opportunities
