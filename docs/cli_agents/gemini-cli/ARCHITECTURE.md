# Gemini CLI - Architecture Documentation

## System Overview

Gemini CLI is a terminal-based AI assistant built on Node.js and TypeScript. It combines a React-based terminal UI with Google's Gemini API to provide an interactive coding experience with support for multimodal inputs, tool execution, and extensible architecture.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          GEMINI CLI ARCHITECTURE                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────────┐      ┌──────────────────┐      ┌──────────────────┐   │
│  │   Terminal UI    │◄────►│   Gemini CLI     │◄────►│   Gemini API     │   │
│  │   (React/Ink)    │      │   Core Engine    │      │  (Gemini Models) │   │
│  └──────────────────┘      └────────┬─────────┘      └──────────────────┘   │
│                                     │                                        │
│           ┌─────────────────────────┼─────────────────────────┐              │
│           │                         │                         │              │
│           ▼                         ▼                         ▼              │
│  ┌──────────────────┐    ┌──────────────────┐    ┌──────────────────┐       │
│  │  Extension System│    │   Tool System    │    │  Session Manager │       │
│  │  (Plugins/MCP)   │    │  (File/Shell/Web)│    │  (Checkpointing) │       │
│  └──────────────────┘    └──────────────────┘    └──────────────────┘       │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Core Components

### 1. Terminal UI Layer

**Technology Stack:**
- **Framework**: React with Ink (React for CLI)
- **Language**: TypeScript
- **Features**:
  - Real-time streaming responses
  - Syntax highlighting with themes
  - Interactive command palette
  - File browser with @ references
  - Permission confirmation dialogs

**Key UI Elements:**
```
┌──────────────────────────────────────────────────────┐
│ Gemini CLI v0.30.0                    gemini-2.5-pro │
├──────────────────────────────────────────────────────┤
│ Context: GEMINI.md, 2 MCP servers                    │
│                                                      │
│ > User message here                                  │
│                                                      │
│ ✦ Gemini: Response streams here...                   │
│                                                      │
│ [Tool: run_shell_command] ls -la                     │
│ [Output displayed...]                                │
│                                                      │
│ > _                                                  │
│                                                      │
│ ~/projects/my-app                    sandbox: docker │
└──────────────────────────────────────────────────────┘
```

### 2. Core Engine

The central orchestrator (`packages/cli/src/`) managing:

- **API Communication**: HTTP/2 streaming to Google Generative AI API
- **Context Management**: Hierarchical GEMINI.md loading, conversation history
- **Tool Execution**: File operations, shell commands, web tools
- **Permission System**: Approval modes (default, auto_edit, plan, yolo)
- **Session State**: Persistence, checkpointing, and resume functionality

### 3. Extension System

**Architecture:**
```
┌─────────────────────────────────────────────────────────────┐
│                   Extension Manager                         │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   Skills    │  │   Commands  │  │    MCP      │         │
│  │  (Context)  │  │   (Slash)   │  │  (Servers)  │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   Hooks     │  │   Themes    │  │   Agents    │         │
│  │  (Events)   │  │   (UI)      │  │ (Subagents) │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**Extension Types:**
1. **Skills**: Specialized procedural expertise (SKILL.md files)
2. **Commands**: Custom slash commands (TOML-based)
3. **MCP Servers**: External tool integration via Model Context Protocol
4. **Hooks**: Event-driven scripts for customization
5. **Themes**: Custom UI color schemes
6. **Agents**: Subagents for parallel task execution (experimental)

### 4. Tool System

**Built-in Tools:**

| Tool | Purpose | Permission Level |
|------|---------|------------------|
| `read_file` | Read file contents | Auto-allow |
| `write_file` | Create/overwrite files | Requires approval |
| `replace` | Edit existing files | Requires approval |
| `list_directory` | List directory contents | Auto-allow |
| `glob` | Find files by pattern | Auto-allow |
| `search_file_content` | Search with grep/ripgrep | Auto-allow |
| `run_shell_command` | Execute shell commands | Requires approval |
| `web_fetch` | Retrieve URL content | Requires approval |
| `google_web_search` | Search the web | Requires approval |
| `save_memory` | Save to GEMINI.md | Requires approval |
| `write_todos` | Manage task lists | Auto-allow |
| `ask_user` | Request user input | Interactive |
| `browser_agent` | Automate browser tasks | Requires approval |

**Tool Execution Flow:**
```
User Request → Tool Selection → Approval Check → Execution → Result
                    │                                        │
                    ▼                                        ▼
            ┌─────────────┐                        ┌─────────────┐
            │   Policy    │                        │  Response   │
            │   Engine    │                        │   to LLM    │
            └─────────────┘                        └─────────────┘
```

### 5. Session Management

**Session Lifecycle:**
```
┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐
│  Start  │───►│  Active │───►│Checkpoint│───►│ Resume  │───►│  End    │
│         │    │         │    │         │    │         │    │         │
└─────────┘    └─────────┘    └─────────┘    └─────────┘    └─────────┘
      │              │              │              │              │
      ▼              ▼              ▼              ▼              ▼
  Initialize   Transcript    Persist to    Load from    Cleanup
  Context      History       ~/.gemini/    ~/.gemini/   Resources
```

**Session Storage:**
- Location: `~/.gemini/`
- Files:
  - `settings.json` - User preferences
  - `sessions/` - Session transcripts
  - `extensions/` - Installed extensions
  - `credentials/` - Authentication tokens

---

## Data Flow

### Request-Response Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         REQUEST LIFECYCLE                               │
└─────────────────────────────────────────────────────────────────────────┘

1. USER INPUT
   └─► Natural language command, @file references, or !shell commands

2. CONTEXT ASSEMBLY
   ├─► Hierarchical GEMINI.md discovery and loading
   ├─► Project structure (via glob/list_directory)
   ├─► MCP server tools and capabilities
   ├─► Active skills and extensions
   └─► Conversation history

3. API REQUEST
   └─► Streaming HTTP/2 to Gemini API
       ├─► Model: gemini-2.5-pro, gemini-3-pro-preview, etc.
       ├─► Tools: Available tool definitions
       ├─► System: Combined context from GEMINI.md files
       └─► Contents: Conversation messages

4. RESPONSE PROCESSING
   └─► Streaming text and function calls
       ├─► Text: Displayed to user
       └─► Function calls: Approval check → Execute

5. TOOL EXECUTION
   └─► If tools called:
       ├─► Policy engine evaluation
       ├─► User approval (if required)
       ├─► Sandboxed execution (if enabled)
       └─► Return result to model

6. FOLLOW-UP
   └─► Model may make additional tool calls
   └─► Or provide final response
```

### Context Management

**Hierarchical Context Loading:**
```
┌────────────────────────────────────────────────────────────┐
│                    CONTEXT HIERARCHY                       │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  1. System Defaults                                        │
│     └─► Base instructions, tool schemas                   │
│                                                            │
│  2. Global GEMINI.md (~/.gemini/GEMINI.md)                │
│     └─► User-wide preferences and instructions            │
│                                                            │
│  3. Workspace GEMINI.md (discovered up tree)              │
│     └─► Project-specific context                          │
│                                                            │
│  4. JIT Context (accessed directories)                    │
│     └─► Component-specific instructions                   │
│                                                            │
│  5. Session Context                                        │
│     └─► Conversation history                              │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

---

## Permission System

### Approval Modes

| Mode | Description | Use Case |
|------|-------------|----------|
| **default** | Prompt for file writes and shell commands | Interactive development |
| **auto_edit** | Auto-approve edit tools, prompt for shell | Faster editing |
| **plan** | Read-only mode, no tool execution | Planning and analysis |
| **yolo** | Auto-approve all actions (requires flag) | Trusted environments |

### Policy Engine

The policy engine evaluates tool executions against configured rules:

```
Tool Request
     │
     ▼
┌─────────────┐
│  Policy     │──Block──► Cancel tool
│  Engine     │
└─────────────┘
     │ Allow
     ▼
┌─────────────┐
│   User      │──Deny────► Show dialog
│ Confirmation│
└─────────────┘
     │ Approve
     ▼
  Execute tool
```

---

## MCP (Model Context Protocol) Integration

```
┌─────────────────────────────────────────────────────────────┐
│                    MCP ARCHITECTURE                         │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Gemini CLI ◄──► MCP Client ◄──► MCP Servers               │
│                                                             │
│  Transport Types:                                           │
│  ├─► stdio (local processes)                                │
│  ├─► HTTP (remote servers)                                  │
│  └─► SSE (server-sent events)                               │
│                                                             │
│  Configuration: ~/.gemini/settings.json                    │
│                                                             │
│  Example Servers:                                           │
│  ├─► GitHub MCP                                            │
│  ├─► PostgreSQL MCP                                        │
│  ├─► Google Search MCP                                     │
│  └─► Custom servers                                        │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Configuration Architecture

### Configuration Hierarchy

```
Configuration Precedence (lowest to highest):

1. Default Settings
   └─► Hardcoded defaults in the application

2. System Defaults File
   └─► /etc/gemini-cli/system-defaults.json

3. User Settings File
   └─► ~/.gemini/settings.json

4. Project Settings File
   └─► .gemini/settings.json

5. System Settings File
   └─► /etc/gemini-cli/settings.json

6. Environment Variables
   └─► GOOGLE_CLOUD_PROJECT, GEMINI_API_KEY, etc.

7. Command Line Flags
   └─► --model, --sandbox, etc.
```

### Key Configuration Files

| File | Purpose | Location |
|------|---------|----------|
| `settings.json` | User preferences | `~/.gemini/` or project `.gemini/` |
| `GEMINI.md` | Project context | Project root or subdirectories |
| `commands/*.toml` | Custom commands | `~/.gemini/commands/` or `.gemini/commands/` |
| `.geminiignore` | Ignore patterns | Project root |

---

## Security Architecture

### Sandboxing

```
┌─────────────────────────────────────────────────────────────┐
│                   SECURITY LAYERS                           │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Layer 1: Permission System                                 │
│  ├─► Approval modes (default, auto_edit, plan)             │
│  ├─► User confirmation for sensitive operations            │
│  └─► Configurable allowed/denied tool patterns             │
│                                                             │
│  Layer 2: Policy Engine                                     │
│  ├─► Pre-execution validation                              │
│  ├─► Pattern matching for dangerous commands               │
│  └─► Custom security policies                              │
│                                                             │
│  Layer 3: Sandboxing                                        │
│  ├─► Docker/Podman container execution                     │
│  ├─► macOS Seatbelt profiles                               │
│  └─► Filesystem isolation                                  │
│                                                             │
│  Layer 4: Trusted Folders                                   │
│  ├─► Folder trust confirmation                             │
│  ├─► Workspace boundary enforcement                        │
│  └─► .gitignore/.geminiignore respect                      │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Performance Considerations

### Optimization Strategies

1. **Context Compaction**: Auto-summarize when near token limits
2. **Token Caching**: Reuse context across turns
3. **Lazy Loading**: Load extensions/tools on demand
4. **Streaming**: Real-time response display
5. **Ripgrep Integration**: Fast file search
6. **Tool Output Masking**: Reduce token usage for large outputs

### Resource Management

```
Memory Usage:
├─► Conversation history (configurable retention)
├─► Extension cache
├─► MCP connection pool
└─► File operation buffers

CPU Usage:
├─► Terminal UI rendering (React/Ink)
├─► File system operations
└─► Hook script execution
```

---

## Package Structure

```
packages/
├── cli/                      # Main CLI application
│   ├── src/
│   │   ├── commands/         # CLI commands (extensions, mcp, skills)
│   │   ├── config/           # Configuration management
│   │   ├── services/         # Core services
│   │   ├── ui/               # React/Ink components
│   │   └── tools/            # Tool implementations
│   └── package.json
└── a2a-server/               # Agent-to-Agent server (experimental)
    └── src/
```

---

## Related Documentation

- [API Reference](./API.md) - Commands and configuration
- [Usage Guide](./USAGE.md) - Practical examples
- [External References](./REFERENCES.md) - Links and resources
