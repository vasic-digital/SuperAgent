# OpenHands - Architecture Documentation

## System Overview

OpenHands is a modular AI software engineering platform with a Python backend, React frontend, and containerized runtime environment. It uses an event-driven architecture where agents communicate through an EventStream.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           OPENHANDS ARCHITECTURE                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────────┐      ┌──────────────────┐      ┌──────────────────┐   │
│  │   React Frontend │◄────►│   WebSocket API  │◄────►│  Agent Controller │   │
│  │   (TypeScript)   │      │   (FastAPI)      │      │  (Python)         │   │
│  └──────────────────┘      └────────┬─────────┘      └────────┬─────────┘   │
│                                     │                        │               │
│           ┌─────────────────────────┴────────────────────────┴─────────┐     │
│           │                                                            │     │
│           ▼                                                            ▼     │
│  ┌──────────────────┐                                    ┌──────────────────┐│
│  │   EventStream    │◄──────────────────────────────────►│   LLM Service    ││
│  │   (Message Bus)  │                                    │   (LiteLLM)      ││
│  └────────┬─────────┘                                    └──────────────────┘│
│           │                                                                  │
│           │     ┌─────────────────────────────────────────────────────┐     │
│           │     │              Runtime (Sandbox)                     │     │
│           │     ├─────────────────────────────────────────────────────┤     │
│           │     │                                                    │     │
│           └────►│  ┌──────────┐  ┌──────────┐  ┌──────────┐        │     │
│                 │  │  Bash    │  │ IPython  │  │ Browser  │        │     │
│                 │  │ Executor │  │  Kernel  │  │   Env    │        │     │
│                 │  └──────────┘  └──────────┘  └──────────┘        │     │
│                 │                                                    │     │
│                 │  ┌──────────┐  ┌──────────┐  ┌──────────┐        │     │
│                 │  │  File    │  │  Git     │  │  Plugins │        │     │
│                 │  │  Editor  │  │  Client  │  │ (Skills) │        │     │
│                 │  └──────────┘  └──────────┘  └──────────┘        │     │
│                 └─────────────────────────────────────────────────────┘     │
│                             ▲                                               │
│                             │                                               │
│                  ┌──────────┴──────────┐                                   │
│                  │   Docker/K8s/Modal  │                                   │
│                  │     Container       │                                   │
│                  └─────────────────────┘                                   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Core Components

### 1. Frontend (React)

**Technology Stack:**
- **Framework**: React 18 with TypeScript
- **State Management**: Zustand, TanStack Query
- **UI Components**: Custom component library
- **Build Tool**: Vite
- **Testing**: Vitest, Playwright

**Key Features:**
- Real-time WebSocket communication
- File browser with syntax highlighting
- Interactive terminal
- Chat interface with streaming responses
- Settings management UI

### 2. WebSocket API Server (FastAPI)

**Components:**
- **listen.py**: Main FastAPI application setup
- **session.py**: WebSocket session management
- **agent_session.py**: Agent lifecycle management
- **conversation_manager.py**: Multi-conversation support

**Endpoints:**
- `/ws`: WebSocket endpoint for real-time communication
- `/api/`: REST API for file operations, settings, uploads
- `/security/`: Security analysis endpoints

### 3. Agent Controller

**Core Classes:**

| Class | Purpose |
|-------|---------|
| `AgentController` | Manages agent state and execution loop |
| `Agent` | Generates actions based on state |
| `State` | Tracks agent progress, history, metrics |
| `EventStream` | Central message bus for events |

**Control Flow:**

```
┌─────────┐     ┌─────────┐     ┌─────────┐     ┌─────────┐
│  State  │────►│  Agent  │────►│  LLM    │────►│ Action  │
└─────────┘     └─────────┘     └─────────┘     └────┬────┘
                                                     │
                              ┌──────────────────────┘
                              ▼
┌─────────┐     ┌─────────┐     ┌─────────┐     ┌─────────┐
│  State  │◄────│  State  │◄────│Observation│◄──│ Runtime │
│ Update  │     │  Check  │     │   Result  │   │Execute  │
└─────────┘     └─────────┘     └─────────┘     └─────────┘
```

### 4. Event System

**Event Types:**

```python
# Actions (User → Agent)
- MessageAction          # User message
- CmdRunAction           # Execute command
- IPythonRunCellAction   # Run Python code
- FileReadAction         # Read file
- FileWriteAction        # Write file
- BrowseURLAction        # Open URL
- ThinkAction            # Agent reasoning
- FinishAction           # Task completion

# Observations (Agent → User)
- CmdOutputObservation   # Command output
- FileReadObservation    # File contents
- BrowserObservation     # Web page content
- ErrorObservation       # Error messages
```

### 5. Runtime (Sandbox)

**Runtime Types:**

| Runtime | Use Case | Isolation |
|---------|----------|-----------|
| **Docker** | Default, local development | Full container isolation |
| **Kubernetes** | Production, scaling | Pod-level isolation |
| **Modal** | Serverless execution | Function-level |
| **Runloop** | Managed infrastructure | Managed sandbox |
| **Remote** | Distributed execution | API-based |
| **Local** | Development, no Docker | None (host execution) |

**Action Execution Server:**
- Runs inside container/pod
- Executes bash commands
- Manages IPython kernel
- Handles browser interactions
- Performs file operations

### 6. LLM Integration (LiteLLM)

**Features:**
- Universal API for 100+ LLM providers
- Automatic provider routing
- Token usage tracking
- Cost estimation
- Retry logic with exponential backoff
- Prompt caching support

**Supported Providers:**
- Anthropic (Claude)
- OpenAI (GPT-4, o-series)
- Google (Gemini)
- Azure OpenAI
- Local models (Ollama)
- And 90+ more via LiteLLM

### 7. Agent Types

#### CodeAct Agent
The primary agent implementing the [CodeAct framework](https://arxiv.org/abs/2402.01030):

**Tools:**
- `execute_bash` - Linux command execution
- `execute_ipython_cell` - Python code execution
- `str_replace_editor` - File viewing and editing
- `browser` - Web page interaction
- `web_read` - Web content extraction

#### Browsing Agent
Specialized for web navigation tasks:
- Automated web browsing
- Form filling and clicking
- Content extraction
- Multi-step navigation

#### VisualBrowsing Agent
Vision-enabled browsing agent:
- Screenshot analysis
- Visual element detection
- Accessibility tree navigation

#### RepoExplorer Agent
Repository analysis specialist:
- Codebase exploration
- Structure analysis
- Cross-reference navigation

### 8. Memory System

**Components:**
- **Conversation Memory**: Chat history and context
- **Repository Memory**: Project structure, files
- **Microagents**: Specialized domain knowledge

**Condenser Types:**

| Condenser | Description |
|-----------|-------------|
| `noop` | No compression (default) |
| `llm` | LLM-based summarization |
| `amortized` | Intelligent forgetting |
| `llm_attention` | Context prioritization |
| `recent` | Keep only recent events |
| `observation_masking` | Mask older observations |

### 9. Security System

**Security Analyzers:**
- **LLM-based**: Uses LLM to analyze actions for risks
- **Invariant**: Rule-based security checks

**Features:**
- Confirmation mode for sensitive actions
- Command injection detection
- File access restrictions
- Security policy enforcement

### 10. MCP Integration

**MCP Server Support:**
- **stdio**: Local process communication
- **shttp**: Streamable HTTP transport
- **sse**: Server-Sent Events (legacy)

**Configuration:**
```toml
[mcp]
stdio_servers = [
    {name = "filesystem", command = "npx", args = ["@modelcontextprotocol/server-filesystem", "/"]}
]
```

---

## Data Flow

### Request Lifecycle

```
1. USER INPUT
   └─► Frontend captures user message or file operation

2. WEBSOCKET TRANSPORT
   └─► Message sent via WebSocket to server

3. SESSION MANAGEMENT
   └─► Session routes to appropriate AgentSession

4. EVENT PUBLICATION
   └─► Action published to EventStream

5. AGENT PROCESSING
   └─► AgentController retrieves action
   └─► Agent generates response using LLM
   └─► Tool calls extracted from response

6. RUNTIME EXECUTION
   └─► Action sent to Runtime
   └─► ActionExecutor runs in sandbox
   └─► Observation generated

7. RESPONSE STREAMING
   └─► Observation published to EventStream
   └─► Streamed back to frontend via WebSocket

8. UI UPDATE
   └─► Frontend renders response
   └─► User can respond or approve actions
```

### State Management

```
┌──────────────────────────────────────────────────────────────┐
│                     STATE LIFECYCLE                          │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌────────┐ │
│  │  INIT    │───►│ RUNNING  │───►│AWAITING  │───►│FINISHED│ │
│  │          │    │          │    │  INPUT   │    │        │ │
│  └──────────┘    └──────────┘    └──────────┘    └────────┘ │
│       │               │                │              │      │
│       ▼               ▼                ▼              ▼      │
│   Initialize     Execute tool    User provides    Task done │
│   context        calls           response                   │
│                                                              │
│   Other states: LOADING, REJECTED, ERROR, PAUSED, STOPPED   │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

---

## Configuration Architecture

### Configuration Hierarchy

```
Priority (lowest to highest):

1. Default Values
   └─► Built into config classes

2. TOML File
   └─► config.toml in project root

3. Environment Variables
   └─► LLM_API_KEY, AGENT_*, etc.

4. Command Line Arguments
   └─► --model, --task, etc.

5. Session Overrides
   └─► Runtime configuration changes
```

### Key Configuration Classes

| Class | Purpose |
|-------|---------|
| `AppConfig` | Root configuration |
| `LLMConfig` | LLM provider settings |
| `AgentConfig` | Agent behavior settings |
| `SandboxConfig` | Runtime environment |
| `SecurityConfig` | Security settings |

---

## Performance Considerations

### Optimization Strategies

1. **Prompt Caching**: Cache system prompts across turns
2. **Context Compaction**: Auto-summarize long conversations
3. **Lazy Loading**: Load plugins on demand
4. **Streaming**: Real-time response display
5. **Parallel Execution**: Multiple runtime instances

### Resource Management

```
Memory Usage:
├─► Event history (configurable retention)
├─► Runtime containers (per-session)
├─► LLM context window
└─► File cache

CPU Usage:
├─► Agent reasoning loops
├─► Runtime command execution
└─► Frontend rendering
```

---

## Related Documentation

- [API Reference](./API.md) - Commands and configuration
- [Usage Guide](./USAGE.md) - Practical examples
- [References](./REFERENCES.md) - External resources
