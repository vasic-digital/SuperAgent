# Plugin Architecture

Detailed architecture documentation for CLI agent plugins.

## Overview

HelixAgent's plugin system enables deep integration with Tier 1 CLI agents, providing:

- **HTTP/3 + QUIC Transport** - Modern, efficient protocol
- **TOON Protocol** - 40-70% token savings
- **Brotli Compression** - Additional bandwidth reduction
- **Real-time Events** - SSE, WebSocket, Webhooks
- **UI Extensions** - AI Debate visualization, progress bars

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           Plugin Architecture                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                        CLI Agent Host                                │    │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐               │    │
│  │  │ Claude Code  │  │   OpenCode   │  │    Cline     │  ...          │    │
│  │  │   Plugin     │  │  MCP Server  │  │  Extension   │               │    │
│  │  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘               │    │
│  └─────────┼─────────────────┼─────────────────┼───────────────────────┘    │
│            │                 │                 │                             │
│            └─────────────────┼─────────────────┘                             │
│                              │                                               │
│  ┌───────────────────────────┴───────────────────────────────────────────┐  │
│  │                    HelixAgent Plugin Libraries                         │  │
│  │  ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐          │  │
│  │  │   Transport     │ │     Events      │ │       UI        │          │  │
│  │  │  ┌───────────┐  │ │  ┌───────────┐  │ │  ┌───────────┐  │          │  │
│  │  │  │  HTTP/3   │  │ │  │    SSE    │  │ │  │  Debate   │  │          │  │
│  │  │  │  + QUIC   │  │ │  │  Client   │  │ │  │ Renderer  │  │          │  │
│  │  │  └───────────┘  │ │  └───────────┘  │ │  └───────────┘  │          │  │
│  │  │  ┌───────────┐  │ │  ┌───────────┐  │ │  ┌───────────┐  │          │  │
│  │  │  │   TOON    │  │ │  │ WebSocket │  │ │  │ Progress  │  │          │  │
│  │  │  │  Codec    │  │ │  │  Client   │  │ │  │   Bars    │  │          │  │
│  │  │  └───────────┘  │ │  └───────────┘  │ │  └───────────┘  │          │  │
│  │  │  ┌───────────┐  │ │  ┌───────────┐  │ │  ┌───────────┐  │          │  │
│  │  │  │  Brotli   │  │ │  │  Webhook  │  │ │  │  Notifs   │  │          │  │
│  │  │  │Compression│  │ │  │  Handler  │  │ │  │           │  │          │  │
│  │  │  └───────────┘  │ │  └───────────┘  │ │  └───────────┘  │          │  │
│  │  └─────────────────┘ └─────────────────┘ └─────────────────┘          │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
│                              │                                               │
│                              ▼                                               │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │                         HelixAgent Server                              │  │
│  │  ┌─────────────────────────────────────────────────────────────────┐  │  │
│  │  │                    AI Debate Ensemble                            │  │  │
│  │  │     15 LLMs (5 positions × 3 per position)                       │  │  │
│  │  │     Multi-pass validation for consensus                          │  │  │
│  │  └─────────────────────────────────────────────────────────────────┘  │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Plugin Components

### 1. Transport Layer

The transport layer handles communication between the CLI agent and HelixAgent.

#### Protocol Negotiation

```
Client → Server: HTTP/3 Request
  ↓ (fallback if unsupported)
Client → Server: HTTP/2 Request
  ↓ (fallback if unsupported)
Client → Server: HTTP/1.1 Request
```

#### Content Negotiation

```
Accept: application/toon+json
Accept-Encoding: br, gzip
```

Fallback chain:
1. TOON format (40-70% smaller)
2. JSON (standard)

Compression fallback:
1. Brotli (best compression)
2. gzip (widely supported)
3. none

#### Transport Interface (Go)

```go
type HelixTransport interface {
    Connect(endpoint string, opts *ConnectOptions) error
    NegotiateProtocol() (Protocol, error)
    NegotiateContent() (ContentType, error)
    NegotiateCompression() (Compression, error)
    Do(ctx context.Context, req *Request) (*Response, error)
    Stream(ctx context.Context, req *Request) (<-chan *Event, error)
}
```

#### Transport Interface (TypeScript)

```typescript
interface HelixTransport {
  connect(endpoint: string, opts?: ConnectOptions): Promise<void>;
  negotiateProtocol(): Promise<Protocol>;
  negotiateContent(): Promise<ContentType>;
  negotiateCompression(): Promise<Compression>;
  request<T>(ctx: Context, req: Request): Promise<Response<T>>;
  stream(ctx: Context, req: Request): AsyncIterable<Event>;
}
```

### 2. Event System

The event system provides real-time updates from HelixAgent.

#### Event Types

**Task Events (14 types):**
```
task.created      - New task created
task.started      - Task execution started
task.progress     - Progress update
task.heartbeat    - Keepalive signal
task.paused       - Task paused
task.resumed      - Task resumed
task.completed    - Task completed successfully
task.failed       - Task failed
task.stuck        - Task detected as stuck
task.cancelled    - Task cancelled
task.retrying     - Task retrying after failure
task.deadletter   - Task moved to dead letter queue
task.log          - Log message from task
task.resource     - Resource usage update
```

**Debate Events (8 types):**
```
debate.started           - Debate session started
debate.round_started     - New debate round
debate.position_submitted - Position submitted by LLM
debate.validation_phase  - Entering validation
debate.polish_phase      - Entering polish/improve
debate.consensus         - Consensus reached
debate.completed         - Debate completed
debate.failed            - Debate failed
```

#### Event Subscription

**SSE (Server-Sent Events):**
```typescript
const events = new EventSource('/v1/tasks/123/events');
events.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log(data.type, data.payload);
};
```

**WebSocket:**
```typescript
const ws = new WebSocket('wss://localhost:7061/v1/ws/tasks/123');
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  handleEvent(data);
};
```

**Webhook:**
```json
POST /webhook
{
  "type": "debate.completed",
  "task_id": "123",
  "payload": {
    "consensus": "...",
    "confidence": 0.95
  }
}
```

### 3. UI Extensions

UI extensions render HelixAgent-specific visualizations in the CLI.

#### Debate Renderer

```
┌─────────────────────────────────────────────────────────────┐
│                   AI Debate in Progress                      │
├─────────────────────────────────────────────────────────────┤
│ Round 2/3                                        [████░░] 67%│
├─────────────────────────────────────────────────────────────┤
│ Position 1: Claude (Advocate)                               │
│ ├─ Argument: "The proposed solution handles edge cases..."  │
│ └─ Confidence: 0.92                                         │
├─────────────────────────────────────────────────────────────┤
│ Position 2: Gemini (Critic)                                 │
│ ├─ Argument: "However, performance could be improved..."    │
│ └─ Confidence: 0.85                                         │
├─────────────────────────────────────────────────────────────┤
│ Position 3: DeepSeek (Synthesizer)                          │
│ ├─ Argument: "Combining both perspectives..."               │
│ └─ Confidence: 0.88                                         │
├─────────────────────────────────────────────────────────────┤
│ Current Phase: VALIDATION                    Votes: 12/15   │
└─────────────────────────────────────────────────────────────┘
```

#### Render Styles

| Style | Description |
|-------|-------------|
| `theater` | Full visualization with animations |
| `novel` | Narrative format with prose |
| `screenplay` | Script-like format |
| `minimal` | Compact, essential info only |
| `plain` | Text only, no formatting |

#### Progress Bars

| Style | Example |
|-------|---------|
| `ascii` | `[========>   ] 75%` |
| `unicode` | `[████████░░░] 75%` |
| `block` | `▓▓▓▓▓▓▓▓░░░ 75%` |
| `dots` | `●●●●●●●●○○○ 75%` |

## Plugin Directory Structure

### Claude Code Plugin

```
plugins/helixagent-integration/
├── .claude-plugin/
│   └── plugin.json           # Plugin manifest
├── hooks/
│   ├── hooks.json            # Hook definitions
│   ├── session_start.js      # Initialize connection
│   ├── session_end.js        # Cleanup
│   ├── pre_tool.js           # Transform to TOON
│   └── post_tool.js          # Render debate results
├── lib/
│   ├── transport.js          # HTTP/3 + TOON + Brotli
│   ├── events.js             # SSE/WebSocket client
│   ├── debate_renderer.js    # AI debate visualization
│   └── progress_renderer.js  # Task progress
└── MANIFEST.md
```

### OpenCode MCP Server

```
cmd/helixagent-mcp-go/
├── main.go                   # Entry point
├── handler.go                # Tool handlers
├── transport.go              # HTTP/3 + TOON
└── ui.go                     # CLI rendering
```

### Cline Extension

```
.clinerules/hooks/
├── task_start.js
├── task_resume.js
├── task_cancel.js
├── task_complete.js
├── user_prompt_submit.js
├── pre_tool_use.js
├── post_tool_use.js
└── pre_compact.js
```

### Kilo-Code Package

```
packages/helixagent/
├── package.json
├── src/
│   ├── index.ts
│   ├── transport/
│   │   ├── quic_transport.ts
│   │   ├── toon_codec.ts
│   │   └── brotli.ts
│   ├── events/
│   │   └── event_client.ts
│   └── ui/
│       ├── debate_renderer.ts
│       └── progress_bar.ts
├── cli/                      # CLI-specific
├── vscode/                   # VSCode-specific
└── jetbrains/                # JetBrains-specific
```

## Configuration Schema

### Plugin Configuration

```json
{
  "$schema": "helixagent-plugin-schema.json",
  "endpoint": "https://localhost:7061",
  "transport": {
    "preferHTTP3": true,
    "enableTOON": true,
    "enableBrotli": true,
    "fallbackChain": ["http3", "http2", "http1.1"],
    "contentFallback": ["toon", "json"],
    "compressionFallback": ["brotli", "gzip", "none"],
    "timeout": 30000
  },
  "events": {
    "transport": "sse",
    "subscribeToDebates": true,
    "subscribeToTasks": true,
    "reconnectInterval": 5000
  },
  "ui": {
    "renderStyle": "theater",
    "progressStyle": "unicode",
    "colorScheme": "256",
    "showResources": true,
    "animate": true
  },
  "debate": {
    "showPhaseIndicators": true,
    "showConfidenceScores": true,
    "showValidationPhase": true
  },
  "notifications": {
    "debateStarted": true,
    "debateCompleted": true,
    "taskProgress": true,
    "errors": true
  }
}
```

## HelixAgent Source Files for Plugin Development

Key files to reference when building plugins:

| File | Purpose |
|------|---------|
| `internal/toon/types.go` | TOON value types |
| `internal/toon/encoder.go` | JSON-based TOON encoder |
| `internal/toon/native_encoder.go` | Native TOON format |
| `internal/http/quic_client.go` | QUIC/HTTP3 client |
| `internal/notifications/cli/renderer.go` | CLI renderer |
| `internal/notifications/sse_manager.go` | SSE server patterns |
| `internal/background/events.go` | Event type definitions |

## Plugin Lifecycle

### Initialization

1. Plugin loads configuration
2. Transport layer connects to HelixAgent
3. Protocol/content/compression negotiation
4. Event subscription established
5. UI renderer initialized

### Request Flow

1. User input received by CLI agent
2. Plugin intercepts request
3. Request encoded (TOON if supported)
4. Request compressed (Brotli if supported)
5. Request sent via HTTP/3 (with fallback)
6. Response received and decoded
7. Events processed for real-time updates
8. UI rendered with debate visualization
9. Response returned to CLI agent

### Cleanup

1. Event subscriptions cancelled
2. Active requests cancelled
3. Transport connection closed
4. Resources released
