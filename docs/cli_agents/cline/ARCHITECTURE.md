# Cline - Architecture Documentation

## System Overview

Cline is a VS Code extension that embeds an autonomous AI coding agent directly in your IDE. Unlike terminal-based AI assistants, Cline integrates seamlessly with the VS Code environment, providing a graphical interface for complex software development tasks with human-in-the-loop approval.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           CLINE SYSTEM ARCHITECTURE                          │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────────┐      ┌──────────────────┐      ┌──────────────────┐   │
│  │   VS Code        │◄────►│   Cline Core     │◄────►│  LLM Providers   │   │
│  │   (Extension     │      │   (TypeScript)   │      │  (Claude/GPT/etc)│   │
│  │    Host)         │      │                  │      │                  │   │
│  └──────────────────┘      └────────┬─────────┘      └──────────────────┘   │
│                                     │                                        │
│           ┌─────────────────────────┼─────────────────────────┐              │
│           │                         │                         │              │
│           ▼                         ▼                         ▼              │
│  ┌──────────────────┐    ┌──────────────────┐    ┌──────────────────┐       │
│  │  Webview UI      │    │   Tool System    │    │  VS Code APIs    │       │
│  │  (React-based)   │    │  (File/Terminal/ │    │  (Editor/Terminal│       │
│  │                  │    │   Browser/MCP)   │    │   /Workspace)    │       │
│  └──────────────────┘    └──────────────────┘    └──────────────────┘       │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Core Components

### 1. VS Code Extension Host

**Technology Stack:**
- **Framework**: VS Code Extension API
- **Language**: TypeScript
- **Entry Point**: `src/extension.ts`
- **Activation Events**: `onLanguage`, `onUri`, `onStartupFinished`

**Key Responsibilities:**
- Extension lifecycle management
- Command registration
- Webview panel hosting
- VS Code API integration

**Extension Manifest (package.json):**
```json
{
  "name": "claude-dev",
  "displayName": "Cline",
  "version": "3.67.1",
  "engines": { "vscode": "^1.84.0" },
  "activationEvents": [
    "onLanguage",
    "onUri", 
    "onStartupFinished"
  ],
  "main": "./dist/extension.js"
}
```

### 2. Cline Core Engine

The central orchestrator that manages AI interactions and task execution.

**Core Modules:**

```
src/core/
├── api/                    # Provider API abstractions
├── assistant-message/      # LLM message handling
├── commands/               # VS Code command implementations
├── context/                # Context management
├── controller/             # Main orchestration logic
├── hooks/                  # Lifecycle hooks
├── ignore/                 # File ignore patterns
├── mentions/               # @-mention handling
├── permissions/            # Permission system
├── prompts/                # System prompts
├── slash-commands/         # /command implementations
├── storage/                # State persistence
├── task/                   # Task management
├── webview/                # Webview communication
└── workspace/              # Workspace operations
```

**Key Functions:**
- **API Communication**: Streaming requests to LLM providers
- **Context Management**: File structure, AST analysis, code search
- **Tool Execution**: File operations, terminal commands, browser automation
- **Permission System**: Human-in-the-loop approval workflow
- **State Management**: Task history, checkpoints, settings

### 3. Webview UI Layer

**Technology Stack:**
- **Framework**: React
- **Location**: `webview-ui/`
- **Communication**: VS Code Webview API

**Key UI Elements:**
```
┌──────────────────────────────────────────────────────┐
│ Cline                                    [Settings]  │
├──────────────────────────────────────────────────────┤
│                                                      │
│  ┌──────────────────────────────────────────────┐   │
│  │  Task: Implement user authentication         │   │
│  └──────────────────────────────────────────────┘   │
│                                                      │
│  Claude: I'll help you implement user authentication.│
│  First, let me explore the project structure...      │
│                                                      │
│  [Tool Use: list_code_definition_names]             │
│  Analyzing src/auth/ directory...                    │
│                                                      │
│  [Tool Use: write_to_file]                          │
│  Creating auth.middleware.ts...                      │
│  ┌──────────────────────────────────────────────┐   │
│  │  Diff view of changes                        │   │
│  │  - Added JWT verification                    │   │
│  │  - Added error handling                      │   │
│  └──────────────────────────────────────────────┘   │
│                                                      │
│  [Approve] [Reject]                                  │
│                                                      │
│  > _                                                 │
│                                                      │
└──────────────────────────────────────────────────────┘
```

**UI Components:**
- Chat interface with streaming responses
- Diff view for file changes
- Tool execution previews
- Settings panel
- MCP server management
- Task history browser

### 4. Tool System

**Built-in Tools:**

| Tool | Purpose | Permission Level |
|------|---------|------------------|
| `read_file` | Read file contents | Auto-allow |
| `write_to_file` | Create/modify files | Requires approval |
| `replace_in_file` | Search/replace edits | Requires approval |
| `search_files` | Ripgrep code search | Auto-allow |
| `list_files` | Directory listing | Auto-allow |
| `list_code_definition_names` | AST analysis | Auto-allow |
| `execute_command` | Terminal commands | Requires approval |
| `browser_action` | Browser automation | Requires approval |
| `ask_followup_question` | Clarification | Auto-allow |
| `attempt_completion` | Task completion | Auto-allow |

**Tool Execution Flow:**
```
User Request → Context Analysis → Tool Selection → Permission Check → Execution
                     │                                                        │
                     ▼                                                        ▼
            ┌─────────────┐                                          ┌─────────────┐
            │ File Structure│                                          │ Result Return │
            │ AST Analysis  │                                          │ to LLM        │
            │ Code Search   │                                          │               │
            └─────────────┘                                          └─────────────┘
```

### 5. LLM Provider System

**Supported Providers:**

| Provider | Authentication | Models |
|----------|---------------|--------|
| **Anthropic** | API Key | Claude 3.7/3.5 Sonnet, Opus, Haiku |
| **OpenAI** | API Key | GPT-4o, GPT-4, GPT-3.5 |
| **OpenRouter** | API Key | 100+ models aggregated |
| **Google** | API Key | Gemini Pro, Ultra |
| **AWS Bedrock** | IAM/Keys | Claude, Llama via AWS |
| **Azure** | Azure AD | OpenAI models |
| **GCP Vertex** | Service Account | Claude, Gemini |
| **Local (Ollama)** | Local | Llama, DeepSeek, Qwen |
| **VS Code LM API** | GitHub | Copilot-integrated models |

**Provider Architecture:**
```
┌─────────────────────────────────────────────────────────────┐
│                     LLM Provider System                      │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐   │
│  │   Provider   │───►│   Request    │───►│   Stream     │   │
│  │   Factory    │    │   Builder    │    │   Handler    │   │
│  └──────────────┘    └──────────────┘    └──────────────┘   │
│                                                              │
│  ┌────────────┐ ┌────────────┐ ┌────────────┐ ┌───────────┐ │
│  │ Anthropic  │ │  OpenAI    │ │ OpenRouter │ │   Local   │ │
│  │  Handler   │ │  Handler   │ │  Handler   │ │  Handler  │ │
│  └────────────┘ └────────────┘ └────────────┘ └───────────┘ │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### 6. MCP (Model Context Protocol) Integration

**MCP Architecture:**
```
┌─────────────────────────────────────────────────────────────┐
│                    MCP INTEGRATION                           │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Cline Extension ◄──► MCP Client ◄──► MCP Servers           │
│                                                              │
│  Transport Types:                                            │
│  ├─► stdio (local processes)                                 │
│  ├─► HTTP (remote servers)                                   │
│  └─► SSE (server-sent events)                                │
│                                                              │
│  Configuration: ~/Library/Application Support/Code/User/     │
│                 globalStorage/saoudrizwan.claude-dev/        │
│                 settings/cline_mcp_settings.json             │
│                                                              │
│  Example Servers:                                            │
│  ├─► GitHub MCP (repo management)                           │
│  ├─► PostgreSQL MCP (database access)                       │
│  ├─► Brave Search MCP (web search)                          │
│  ├─► Filesystem MCP (file operations)                       │
│  └─► Context7 MCP (documentation)                           │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Data Flow

### Request-Response Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         REQUEST LIFECYCLE                               │
└─────────────────────────────────────────────────────────────────────────┘

1. USER INPUT
   └─► Natural language task description entered in Cline UI

2. CONTEXT ASSEMBLY
   ├─► Current workspace file structure (via list_files)
   ├─► Source code AST analysis (via list_code_definition_names)
   ├─► Relevant file contents (via read_file)
   ├─► Search results (via search_files)
   ├─► .clinerules content (if exists)
   └─► Active MCP server capabilities

3. API REQUEST
   └─► Streaming request to selected LLM provider
       ├─► Model: Claude 3.7 Sonnet (default)
       ├─► Tools: Available tool definitions
       └─► Messages: Conversation context + system prompt

4. RESPONSE PROCESSING
   └─► Streaming text and tool call requests
       ├─► Text: Displayed in chat UI
       └─► Tool calls: Permission check → Execute → Return result

5. TOOL EXECUTION
   └─► If tools requested:
       ├─► Show permission dialog (if required)
       ├─► Execute tool via VS Code APIs
       ├─► Display results/diffs
       └─► Return result to LLM

6. ITERATION
   └─► LLM may request additional tools
   └─► Or provide final response with attempt_completion
```

### Context Management Strategy

```
┌────────────────────────────────────────────────────────────┐
│                    CONTEXT WINDOW                          │
├────────────────────────────────────────────────────────────┤
│                                                              │
│  System Prompt (~2K tokens)                                  │
│  ├─► Base instructions and capabilities                    │
│  ├─► Tool definitions                                        │
│  └─► .clinerules content                                     │
│                                                              │
│  Conversation History (~100K tokens)                         │
│  ├─► User messages                                           │
│  ├─► Assistant responses                                     │
│  ├─► Tool results                                            │
│  └─► File contents                                           │
│                                                              │
│  Checkpoint System                                           │
│  └─► Workspace snapshots at each step                       │
│  └─► Compare/Restore functionality                          │
│                                                              │
└────────────────────────────────────────────────────────────┘
```

---

## Permission System

### Human-in-the-Loop Workflow

```
┌────────────────────────────────────────────────────────────┐
│                 PERMISSION DECISION TREE                   │
├────────────────────────────────────────────────────────────┘
│
│  Tool Request
│       │
│       ▼
│  ┌─────────────────────────────────────────────┐
│  │ Safe Operations (read/search/list)          │
│  │ read_file, search_files, list_files         │
│  └──────────────┬──────────────────────────────┘
│                 │ YES
│                 ▼
│           Auto-Execute
│                 │
│                 │ NO
│                 ▼
│  ┌─────────────────────────────────────────────┐
│  │ Destructive Operations                      │
│  │ write_to_file, replace_in_file              │
│  │ execute_command, browser_action             │
│  └──────────────┬──────────────────────────────┘
│                 │
│                 ▼
│         Show Approval Dialog
│                 │
│       ┌─────────┴─────────┐
│       │                   │
│       ▼                   ▼
│   [Approve]          [Reject]
│       │                   │
│       ▼                   ▼
│   Execute            Skip Tool
│       │                   │
│       ▼                   ▼
│  Return Result      Explain Why
│
└────────────────────────────────────────────────────────────┘
```

### Auto-Approve Settings

Users can configure auto-approval for specific operations:

| Operation | Risk Level | Auto-Approve Option |
|-----------|------------|---------------------|
| Read files | Low | Available |
| Edit files | Medium | Available with restrictions |
| Execute commands | High | Available for safe commands |
| Browser actions | Medium | Available |

---

## Checkpoint System

### Workspace Snapshots

```
Task Progress Timeline:
────────────────────────────────────────────────────────────►

    Start        Step 1        Step 2        Step 3       End
      │            │             │             │           │
      ▼            ▼             ▼             ▼           ▼
  ┌────────┐   ┌────────┐    ┌────────┐    ┌────────┐  ┌────────┐
  │Initial │   │Snapshot│    │Snapshot│    │Snapshot│  │ Final  │
  │ State  │   │   1    │    │   2    │    │   3    │  │ State  │
  └────────┘   └────────┘    └────────┘    └────────┘  └────────┘
                    │             │             │
                    ▼             ▼             ▼
               [Compare]     [Compare]     [Compare]
               [Restore]     [Restore]     [Restore]

Restore Options:
- Restore Workspace Only: Revert files, keep chat
- Restore Task and Workspace: Revert everything
```

---

## Storage Architecture

### State Persistence

```
~/Library/Application Support/Code/User/
└── globalStorage/
    └── saoudrizwan.claude-dev/
        ├── settings/
        │   ├── cline_mcp_settings.json    # MCP server config
        │   └── custom-instructions.md     # Custom system prompt
        ├── tasks/                         # Task history
        │   ├── task-uuid-1/
        │   │   ├── api_conversation_history.json
        │   │   ├── ui_messages.json
        │   │   └── checkpoints/
        │   └── task-uuid-2/
        │       └── ...
        └── cache/
            └── ...
```

---

## Browser Integration

### Computer Use Capability

```
┌─────────────────────────────────────────────────────────────┐
│                  BROWSER AUTOMATION                          │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Cline Extension ◄──► Puppeteer ◄──► Headless Chrome        │
│                                                              │
│  Supported Actions:                                          │
│  ├─► launch (start browser)                                  │
│  ├─► click (element interaction)                             │
│  ├─► type (text input)                                       │
│  ├─► scroll (page navigation)                                │
│  ├─► screenshot (visual capture)                             │
│  └─► close (shutdown browser)                                │
│                                                              │
│  Use Cases:                                                  │
│  ├─► Visual bug debugging                                    │
│  ├─► End-to-end testing                                      │
│  ├─► Runtime error capture                                   │
│  └─► General web automation                                  │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Configuration Architecture

### Configuration Hierarchy

```
Configuration Precedence (lowest to highest):

1. Default Settings
   └─► Built into Cline extension

2. Global Settings
   └─► VS Code settings (cline.*)

3. Project Settings
   └─► .clinerules file in project root

4. MCP Settings
   └─► cline_mcp_settings.json

5. Session Settings
   └─► Per-task configurations
```

### Key Configuration Files

| File | Purpose | Location |
|------|---------|----------|
| `.clinerules` | Project context rules | Project root |
| `cline_mcp_settings.json` | MCP server config | VS Code global storage |
| VS Code Settings | Extension preferences | User/Workspace settings |

---

## Security Architecture

### Security Layers

```
┌─────────────────────────────────────────────────────────────┐
│                   SECURITY ARCHITECTURE                      │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Layer 4: User Control                                       │
│  ├─► Approval dialogs for all destructive operations        │
│  ├─► Diff preview before file modifications                 │
│  ├─► Checkpoint restore capability                          │
│  └─► Auto-approve opt-in only                               │
│                                                              │
│  Layer 3: VS Code Sandbox                                    │
│  ├─► Extension runs in VS Code context                      │
│  ├─► File access limited to workspace                       │
│  └─► Terminal access requires user approval                 │
│                                                              │
│  Layer 2: Protected Operations                               │
│  ├─► Write operations require explicit approval             │
│  ├─► Command execution prompts for confirmation             │
│  └─► Browser automation is opt-in                           │
│                                                              │
│  Layer 1: Read-Only Defaults                                 │
│  ├─► File reads are safe by default                         │
│  ├─► Search operations are safe                             │
│  └─► AST analysis is read-only                              │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Performance Considerations

### Optimization Strategies

1. **Streaming Responses**: Real-time token display
2. **Lazy Context Loading**: Load files only when needed
3. **Incremental Updates**: Partial UI updates
4. **Checkpoint Compression**: Efficient snapshot storage
5. **MCP Connection Pooling**: Reuse server connections

### Resource Management

```
Memory Usage:
├─► Webview React app
├─► Conversation history
├─► File cache
├─► MCP connection pool
└─► Browser instances (when active)

CPU Usage:
├─► UI rendering
├─► File system operations
├─► Code search (ripgrep)
└─► AST parsing
```

---

## Related Documentation

- [API Reference](./API.md) - Commands and configuration
- [Usage Guide](./USAGE.md) - Practical examples
- [References](./REFERENCES.md) - External resources
- [Diagrams](./DIAGRAMS.md) - Visual documentation
- [Gap Analysis](./GAP_ANALYSIS.md) - Improvement opportunities
