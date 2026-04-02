# Amazon Q CLI - Architecture Documentation

## System Overview

Amazon Q CLI is a terminal-based AI coding assistant built in Rust. It combines a terminal UI with AWS AI services to provide an interactive coding experience with deep AWS integration.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        AMAZON Q CLI ARCHITECTURE                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────────┐      ┌──────────────────┐      ┌──────────────────┐   │
│  │   Terminal UI    │◄────►│   Amazon Q CLI   │◄────►│   AWS Services   │   │
│  │   (Rust/TUI)     │      │   Core Engine    │      │  (CodeWhisperer) │   │
│  └──────────────────┘      └────────┬─────────┘      └──────────────────┘   │
│                                     │                                        │
│           ┌─────────────────────────┼─────────────────────────┐              │
│           │                         │                         │              │
│           ▼                         ▼                         ▼              │
│  ┌──────────────────┐    ┌──────────────────┐    ┌──────────────────┐       │
│  │  Agent System    │    │   Tool System    │    │ Knowledge Base   │       │
│  │  (Custom Agents) │    │  (Built-in/MCP)  │    │  (Semantic)      │       │
│  └──────────────────┘    └──────────────────┘    └──────────────────┘       │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Core Components

### 1. Terminal UI Layer

**Technology Stack:**
- **Language**: Rust
- **UI Framework**: Custom terminal UI (chat-cli-ui crate)
- **Features**:
  - Real-time streaming responses
  - Syntax highlighting
  - Interactive command palette
  - File browser integration
  - Permission dialogs

**Key UI Elements:**
```
┌──────────────────────────────────────────────────────┐
│ Amazon Q Developer CLI                               │
├──────────────────────────────────────────────────────┤
│                                                      │
│ > User message here                                  │
│                                                      │
│ Q: Response appears here with streaming...           │
│                                                      │
│ [Tool: execute_bash] $ ls -la                        │
│ [Output shown...]                                    │
│                                                      │
│ > _                                                  │
│                                                      │
└──────────────────────────────────────────────────────┘
```

### 2. Core Engine (chat-cli)

The central orchestrator manages:
- **API Communication**: HTTP/2 streaming to AWS CodeWhisperer API
- **Context Management**: Conversation history and file context
- **Tool Execution**: Built-in tools and MCP servers
- **Permission System**: Configurable tool permissions
- **Session State**: Persistence and resume functionality

**Module Structure:**
```
crates/chat-cli/src/
├── main.rs           # Entry point
├── cli.rs            # CLI argument parsing
├── agent/            # Agent management
├── api_client/       # AWS API client
├── auth/             # Authentication handling
├── aws_common/       # AWS common utilities
├── constants/        # Application constants
├── database/         # Local storage
├── logging/          # Logging infrastructure
├── mcp_client/       # MCP protocol client
├── os/               # OS-specific code
├── request/          # Request handling
├── telemetry/        # Usage telemetry
├── theme/            # UI theming
└── util/             # Utility functions
```

### 3. Agent System

**Agent Architecture:**
```
┌─────────────────────────────────────────────────────────────┐
│                     Agent Manager                           │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   Prompt    │  │    Tools    │  │   MCP Svrs  │         │
│  │  (Context)  │  │  (Built-in) │  │  (External) │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   Hooks     │  │  Resources  │  │  Settings   │         │
│  │  (Events)   │  │   (Files)   │  │   (Config)  │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**Agent Configuration Files:**
- Location: `~/.aws/amazonq/cli-agents/` (global) or `.amazonq/cli-agents/` (local)
- Format: JSON with fields for prompt, tools, MCP servers, hooks, etc.

### 4. Tool System

**Built-in Tools:**

| Tool | Purpose | Permission Level |
|------|---------|------------------|
| `execute_bash` | Execute shell commands | Configurable |
| `fs_read` | Read file contents | Trusted |
| `fs_write` | Create/edit files | Prompt |
| `introspect` | Q CLI documentation | Trusted |
| `knowledge` | Knowledge base access | Trusted |
| `report_issue` | GitHub issue template | Trusted |
| `thinking` | Reasoning mechanism | Trusted |
| `todo_list` | Task management | Trusted |
| `use_aws` | AWS CLI calls | Configurable |

**Tool Execution Flow:**
```
User Request → Tool Selection → Permission Check → Execution → Result
                    │                                        │
                    ▼                                        ▼
            ┌─────────────┐                        ┌─────────────┐
            │   Hooks     │                        │   Hooks     │
            │  PreToolUse │                        │ PostToolUse │
            └─────────────┘                        └─────────────┘
```

### 5. MCP (Model Context Protocol) Integration

```
┌─────────────────────────────────────────────────────────────┐
│                    MCP ARCHITECTURE                         │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Amazon Q CLI ◄──► MCP Client ◄──► MCP Servers             │
│                                                             │
│  Configuration:                                             │
│  ├─~/.aws/amazonq/mcp.json (global)                        │
│  ├─./.amazonq/mcp.json (local)                             │
│  └─Agent mcpServers field                                   │
│                                                             │
│  Example Servers:                                           │
│  ├─ Git MCP                                                │
│  ├─ Fetch MCP                                              │
│  └─ Custom servers                                         │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 6. Knowledge Base System

**Architecture:**
```
┌─────────────────────────────────────────────────────────────┐
│                  KNOWLEDGE BASE SYSTEM                      │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Indexing Types:                                            │
│  ├─ Fast (BM25): Lexical search, fast indexing             │
│  └─ Best (Embeddings): Semantic search, AI-powered         │
│                                                             │
│  Storage:                                                   │
│  ~/.aws/amazonq/knowledge_bases/                            │
│  ├── <agent-name>/                                          │
│  │   ├── contexts.json                                      │
│  │   └── <context-id>/                                      │
│  │       ├── data.json                                      │
│  │       └── bm25_data.json                                 │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**Features:**
- Agent-isolated knowledge bases
- Background indexing
- Pattern filtering (include/exclude)
- Semantic and lexical search

---

## Data Flow

### Request-Response Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         REQUEST LIFECYCLE                               │
└─────────────────────────────────────────────────────────────────────────┘

1. USER INPUT
   └─► Natural language command or /slash command

2. CONTEXT ASSEMBLY
   ├─► Current file context
   ├─► Project structure (via fs_read)
   ├─► Agent prompt and resources
   ├─► Active knowledge bases
   └─► Conversation history

3. API REQUEST
   └─► Streaming HTTP/2 to AWS CodeWhisperer
       ├─► Model: Claude models via AWS
       ├─► Tools: Available tool definitions
       └─► Messages: Conversation context

4. RESPONSE PROCESSING
   └─► Streaming text and tool calls
       ├─► Text: Displayed to user
       └─► Tool calls: Permission check → Execute

5. TOOL EXECUTION
   └─► If tools called:
       ├─► PreToolUse hooks
       ├─► User permission (if required)
       ├─► Execute tool
       ├─► PostToolUse hooks
       └─► Return result to model

6. FOLLOW-UP
   └─► Model may make additional tool calls
   └─► Or provide final response
```

---

## Hook System

### Event Types

```
┌─────────────────────────────────────────────────────────────┐
│                      HOOK EVENTS                            │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  agentSpawn         ──► When agent initializes              │
│                                                             │
│  userPromptSubmit   ──► When user submits message           │
│                                                             │
│  preToolUse         ──► Before tool execution               │
│       │                                                     │
│       ├── Can block tool (exit code 2)                      │
│       ├── Can audit/log                                     │
│       └── Can modify context                                │
│                                                             │
│  postToolUse        ──► After tool execution                │
│       │                                                     │
│       ├── Can format output                                 │
│       └── Can trigger follow-up                             │
│                                                             │
│  stop               ──► When response completes             │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Hook Execution

Hooks receive JSON via stdin and output JSON to stdout:
- **Exit 0**: Success
- **Exit 2** (preToolUse): Block tool execution
- **Other**: Warning shown to user

---

## Configuration Architecture

### Configuration Hierarchy

```
Configuration Precedence (lowest to highest):

1. Built-in Default Agent
   └─► Hardcoded fallback configuration

2. Global Settings
   └─► ~/.aws/amazonq/settings

3. Project Settings
   └─► .amazonq/settings

4. Environment Variables
   └─► Q_* variables

5. Command Line Flags
   └─► --flag options

6. Session Overrides
   └─► /settings commands
```

### Key Configuration Files

| File | Purpose | Location |
|------|---------|----------|
| `settings` | User preferences | `~/.aws/amazonq/` or `.amazonq/` |
| `mcp.json` | MCP server config | `~/.aws/amazonq/` or `.amazonq/` |
| `<agent>.json` | Agent definitions | `cli-agents/` directories |
| `AmazonQ.md` | Project context | Project root |

---

## Security Architecture

### Permission System

```
┌────────────────────────────────────────────────────────────┐
│                 PERMISSION DECISION TREE                   │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  Tool Request                                              │
│       │                                                    │
│       ▼                                                    │
│  ┌─────────────┐                                           │
│  │ In allowed  │──Yes──► Execute without prompt            │
│  │   Tools?    │                                           │
│  └─────────────┘                                           │
│       │ No                                                 │
│       ▼                                                    │
│  ┌─────────────┐                                           │
│  │ PreToolUse  │──Block──► Cancel tool                     │
│  │   Hooks     │                                           │
│  └─────────────┘                                           │
│       │ Allow                                              │
│       ▼                                                    │
│  ┌─────────────┐                                           │
│  │ Tool Config │──Deny────► Show permission dialog         │
│  │   Rules     │                                           │
│  └─────────────┘                                           │
│       │ Allowed                                            │
│       ▼                                                    │
│   Execute tool                                             │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

### Security Layers

1. **Tool Permissions**: Configurable allow/deny lists
2. **Hook System**: Custom validation and blocking
3. **Path Restrictions**: Allowed/denied paths for file operations
4. **AWS Service Controls**: Allowed/denied AWS services
5. **Command Filtering**: Regex-based command restrictions

---

## AWS Service Integration

### CodeWhisperer Client

The `amzn-codewhisperer-client` crate provides:
- Authentication with AWS
- Streaming API communication
- Code completion generation
- Code analysis and security scanning
- Test generation
- Code transformation

### Telemetry

The `amzn-toolkit-telemetry-client` handles:
- Usage analytics
- Feature evaluation
- Error reporting
- Performance metrics

---

## Performance Considerations

### Optimization Strategies

1. **Streaming Responses**: Real-time display without buffering
2. **Background Indexing**: Knowledge base indexing happens async
3. **Tool Result Caching**: Hook results cached with TTL
4. **Context Management**: Efficient conversation context handling
5. **Lazy Loading**: Agents and tools loaded on demand

### Resource Management

```
Memory Usage:
├─► Conversation history (context window managed)
├─► Knowledge base embeddings
├─► MCP connection pool
└─► Tool execution buffers

CPU Usage:
├─► Terminal UI rendering
├─► Background indexing
├─► Hook script execution
└─► File system operations
```

---

## Related Documentation

- [API Reference](./API.md) - Commands and configuration
- [Usage Guide](./USAGE.md) - Practical examples
- [External References](./REFERENCES.md) - Links and resources
