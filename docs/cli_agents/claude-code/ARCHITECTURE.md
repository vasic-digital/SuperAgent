# Claude Code - Architecture Documentation

## System Overview

Claude Code is a terminal-based AI coding assistant built on Node.js. It combines a terminal UI with the Claude API to provide an interactive coding experience.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           CLAUDE CODE ARCHITECTURE                           │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────────┐      ┌──────────────────┐      ┌──────────────────┐   │
│  │   Terminal UI    │◄────►│   Claude Code    │◄────►│  Anthropic API   │   │
│  │   (React/Blessed)│      │   Core Engine    │      │  (Claude Models) │   │
│  └──────────────────┘      └────────┬─────────┘      └──────────────────┘   │
│                                     │                                        │
│           ┌─────────────────────────┼─────────────────────────┐              │
│           │                         │                         │              │
│           ▼                         ▼                         ▼              │
│  ┌──────────────────┐    ┌──────────────────┐    ┌──────────────────┐       │
│  │  Plugin System   │    │   Tool System    │    │  Session Manager │       │
│  │  (Extensions)    │    │  (Bash/Edit/etc) │    │  (Persistence)   │       │
│  └──────────────────┘    └──────────────────┘    └──────────────────┘       │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Core Components

### 1. Terminal UI Layer

**Technology Stack:**
- **Framework**: React-based terminal rendering
- **Library**: Blessed or similar terminal UI library
- **Features**:
  - Real-time streaming responses
  - Syntax highlighting
  - Interactive command palette
  - File browser integration
  - Permission dialogs

**Key UI Elements:**
```
┌──────────────────────────────────────────────────────┐
│ Claude Code v2.1.90                          [Model] │
├──────────────────────────────────────────────────────┤
│                                                      │
│ > User message here                                  │
│                                                      │
│ Claude: Response appears here with streaming...      │
│                                                      │
│ [Tool Use: Bash] $ ls -la                            │
│ [Output shown...]                                    │
│                                                      │
│ > _                                                  │
│                                                      │
└──────────────────────────────────────────────────────┘
```

### 2. Core Engine

The central orchestrator that manages:
- **API Communication**: HTTP/2 streaming to Anthropic API
- **Context Management**: Conversation history and file context
- **Tool Execution**: Bash, Edit, Read, Write, etc.
- **Permission System**: Auto-mode classifier and user approvals
- **Session State**: Persistence and resume functionality

### 3. Plugin System

**Plugin Architecture:**
```
┌─────────────────────────────────────────────────────────────┐
│                     Plugin Manager                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │  Commands   │  │   Agents    │  │    Hooks    │         │
│  │  (Slash)    │  │ (Subagents) │  │  (Events)   │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   Skills    │  │    MCP      │  │   Config    │         │
│  │ (Context)   │  │  (Servers)  │  │   (JSON)    │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**Plugin Types:**
1. **Commands**: Custom slash commands (`/command-name`)
2. **Agents**: Specialized subagents for parallel tasks
3. **Hooks**: Event-driven handlers (PreToolUse, PostToolUse, Stop, etc.)
4. **Skills**: Contextual knowledge injection
5. **MCP Servers**: External tool integration

### 4. Tool System

**Built-in Tools:**

| Tool | Purpose | Permission Level |
|------|---------|------------------|
| `Bash` | Execute shell commands | Requires approval |
| `Read` | Read file contents | Auto-allow |
| `Write` | Create new files | Requires approval |
| `Edit` | Modify existing files | Requires approval |
| `Grep` | Search file contents | Auto-allow |
| `Glob` | Find files by pattern | Auto-allow |
| `LS` | List directory contents | Auto-allow |
| `Think` | Reasoning/thinking step | Auto-allow |

**Tool Execution Flow:**
```
User Request → Tool Selection → Permission Check → Execution → Result
                    │                                        │
                    ▼                                        ▼
            ┌─────────────┐                        ┌─────────────┐
            │  PreToolUse │                        │ PostToolUse │
            │   Hooks     │                        │    Hooks    │
            └─────────────┘                        └─────────────┘
```

### 5. Session Management

**Session Lifecycle:**
```
┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐
│  Start  │───►│  Active │───►│  Save   │───►│ Resume  │───►│  End    │
│         │    │         │    │         │    │         │    │         │
└─────────┘    └─────────┘    └─────────┘    └─────────┘    └─────────┘
      │              │              │              │              │
      ▼              ▼              ▼              ▼              ▼
  Initialize   Transcript    Persist to    Load from    Cleanup
  Context      History       ~/.claude/    ~/.claude/   Resources
```

**Session Storage:**
- Location: `~/.claude/`
- Files:
  - `history.jsonl` - Command history
  - `settings.json` - User preferences
  - Transcript directories for each session

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
   ├─► Project structure (via Glob/LS)
   ├─► Relevant CLAUDE.md files
   ├─► Active skills context
   └─► Conversation history

3. API REQUEST
   └─► Streaming HTTP/2 to Anthropic API
       ├─► Model: Claude 3.5/3.7 Sonnet, Opus, etc.
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

### Context Management

**Context Window Strategy:**
```
┌────────────────────────────────────────────────────────────┐
│                    CONTEXT WINDOW                          │
├────────────────────────────────────────────────────────────┤
│ System Prompt                                              │
│ ├─► Base instructions                                      │
│ ├─► Tool definitions                                       │
│ └─► CLAUDE.md content                                      │
├────────────────────────────────────────────────────────────┤
│ Conversation History                                       │
│ ├─► User messages                                          │
│ ├─► Assistant responses                                    │
│ └─► Tool results                                           │
├────────────────────────────────────────────────────────────┤
│ Auto-Compaction (when near limit)                          │
│ └─► Summarize older messages                               │
└────────────────────────────────────────────────────────────┘
```

---

## Permission System

### Auto Mode Classifier

```
┌────────────────────────────────────────────────────────────┐
│                 PERMISSION DECISION TREE                   │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  Tool Request                                              │
│       │                                                    │
│       ▼                                                    │
│  ┌─────────────┐                                           │
│  │ Auto Mode   │──Yes──► Execute without prompt            │
│  │ Enabled?    │                                           │
│  └─────────────┘                                           │
│       │ No                                                 │
│       ▼                                                    │
│  ┌─────────────┐                                           │
│  │ PreToolUse  │──Block──► Cancel tool                     │
│  │ Hooks       │                                           │
│  └─────────────┘                                           │
│       │ Allow                                              │
│       ▼                                                    │
│  ┌─────────────┐                                           │
│  │ Classifier  │──Deny────► Show permission dialog         │
│  │ Confidence  │                                           │
│  └─────────────┘                                           │
│       │ High confidence                                    │
│       ▼                                                    │
│   Execute tool                                             │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

### Permission Levels

| Level | Description | Examples |
|-------|-------------|----------|
| **Auto-allow** | No permission needed | Read, Grep, LS, Glob |
| **Auto-mode** | AI decides based on context | Bash (safe commands) |
| **Prompt** | Always ask user | Write, Edit, Bash (dangerous) |
| **Block** | Never allowed (hooks) | Configured via PreToolUse |

---

## Hook System

### Event Types

```
┌─────────────────────────────────────────────────────────────┐
│                      HOOK EVENTS                            │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  SessionStart       ──► When session begins                 │
│                                                             │
│  PreToolUse         ──► Before tool execution               │
│       │                                                     │
│       ├── Can block tool                                    │
│       ├── Can modify parameters                             │
│       └── Can add context                                   │
│                                                             │
│  PostToolUse        ──► After tool execution                │
│       │                                                     │
│       ├── Can modify output                                 │
│       └── Can trigger actions                               │
│                                                             │
│  PermissionDenied   ──► When auto-mode denies               │
│       │                                                     │
│       └── Can retry with {retry: true}                      │
│                                                             │
│  TaskCreated        ──► When subagent spawned               │
│                                                             │
│  Stop               ──► When user tries to exit             │
│       │                                                     │
│       └── Can intercept and continue                        │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Hook Execution Model

Hooks are executable scripts (Python, Bash, etc.) that:
1. Receive event data via stdin
2. Output JSON to stdout
3. Exit codes determine behavior:
   - `0`: Success, use output
   - `1`: Error
   - `2`: Block (for PreToolUse)

---

## MCP (Model Context Protocol) Integration

```
┌─────────────────────────────────────────────────────────────┐
│                    MCP ARCHITECTURE                         │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Claude Code ◄──► MCP Client ◄──► MCP Servers              │
│                                                             │
│  Transport Types:                                           │
│  ├─► stdio (local processes)                                │
│  ├─► HTTP (remote servers)                                  │
│  └─► SSE (server-sent events)                               │
│                                                             │
│  Configuration: ~/.claude/.mcp.json                        │
│                                                             │
│  Example Servers:                                           │
│  ├─► GitHub MCP                                            │
│  ├─► PostgreSQL MCP                                        │
│  ├─► Brave Search MCP                                      │
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
   └─► Built into Claude Code

2. Global Settings
   └─► ~/.claude/settings.json

3. Project Settings
   └─► .claude/settings.json

4. Environment Variables
   └─► CLAUDE_CODE_* variables

5. Command Line Flags
   └─► --flag options

6. Session Overrides
   └─► /config commands
```

### Key Configuration Files

| File | Purpose | Location |
|------|---------|----------|
| `settings.json` | User preferences | `~/.claude/` or project `.claude/` |
| `.mcp.json` | MCP server config | `~/.claude/` or project root |
| `CLAUDE.md` | Project context | Project root or subdirectories |
| `hooks.json` | Hook definitions | Plugin directory |

---

## Security Architecture

### Sandboxing

```
┌─────────────────────────────────────────────────────────────┐
│                   SECURITY LAYERS                           │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Layer 1: Permission System                                 │
│  ├─► Every tool call evaluated                              │
│  ├─► User approval for dangerous operations                 │
│  └─► Configurable allow lists                               │
│                                                             │
│  Layer 2: Hook System                                       │
│  ├─► PreToolUse hooks can block                             │
│  ├─► Pattern matching for dangerous commands                │
│  └─► Custom security policies                               │
│                                                             │
│  Layer 3: Auto Mode Classifier                              │
│  ├─► AI-powered safety evaluation                           │
│  ├─► Confidence scoring                                     │
│  └─► Conservative defaults                                  │
│                                                             │
│  Layer 4: Protected Directories                             │
│  ├─► ~/.ssh, ~/.aws, ~/.gnupg                              │
│  ├─► System directories                                     │
│  └─► Configurable protections                               │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Performance Considerations

### Optimization Strategies

1. **Prompt Caching**: Reuse context across turns
2. **Lazy Loading**: Load plugins/tools on demand
3. **Streaming**: Real-time response display
4. **Context Compaction**: Auto-summarize long sessions
5. **Parallel Agents**: Spawn subagents for concurrent tasks

### Resource Management

```
Memory Usage:
├─► Transcript files (limited retention)
├─► Plugin cache
├─► MCP connection pool
└─► LRU caches for tool results

CPU Usage:
├─► Terminal UI rendering
├─► Hook script execution
└─► File system operations
```

---

## Related Documentation

- [API Reference](./API.md) - Commands and configuration
- [Usage Guide](./USAGE.md) - Practical examples
- [Development Guide](./DEVELOPMENT.md) - Contributing
- [Plugin Documentation](../plugins/README.md) - Plugin development
