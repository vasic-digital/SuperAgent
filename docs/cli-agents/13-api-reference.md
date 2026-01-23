# API Reference

Complete API reference for HelixAgent CLI agent integration.

## REST API Endpoints

### Chat Completions

**POST /v1/chat/completions**

Create a chat completion with the AI Debate Ensemble.

**Request:**
```json
{
  "model": "helixagent-debate",
  "messages": [
    {"role": "system", "content": "You are a helpful assistant."},
    {"role": "user", "content": "What is the capital of France?"}
  ],
  "temperature": 0.7,
  "max_tokens": 1024,
  "stream": false,
  "enable_debate": true,
  "validation_config": {
    "enable_validation": true,
    "enable_polish": true,
    "max_validation_rounds": 3
  }
}
```

**Response:**
```json
{
  "id": "chatcmpl-abc123",
  "object": "chat.completion",
  "created": 1705312200,
  "model": "helixagent-debate",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "The capital of France is Paris."
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 25,
    "completion_tokens": 10,
    "total_tokens": 35
  },
  "debate": {
    "rounds": 3,
    "participants": ["Claude", "Gemini", "DeepSeek"],
    "consensus": "Paris is the capital of France",
    "confidence": 0.98,
    "votes": {"for": 15, "total": 15}
  }
}
```

**Headers:**
| Header | Value | Description |
|--------|-------|-------------|
| `Content-Type` | `application/json` or `application/toon+json` | Request format |
| `Accept` | `application/json` or `application/toon+json` | Response format |
| `Accept-Encoding` | `br, gzip` | Compression preference |
| `Authorization` | `Bearer <api_key>` | API authentication |

---

### Embeddings

**POST /v1/embeddings**

Generate vector embeddings for text.

**Request:**
```json
{
  "input": ["Hello, world!", "How are you?"],
  "model": "text-embedding-3-small",
  "encoding_format": "float"
}
```

**Response:**
```json
{
  "object": "list",
  "data": [
    {
      "object": "embedding",
      "index": 0,
      "embedding": [0.123, -0.456, 0.789, ...]
    },
    {
      "object": "embedding",
      "index": 1,
      "embedding": [0.321, -0.654, 0.987, ...]
    }
  ],
  "model": "text-embedding-3-small",
  "usage": {
    "prompt_tokens": 8,
    "total_tokens": 8
  }
}
```

---

### Vision

**POST /v1/vision/analyze**

Analyze images using vision models.

**Request:**
```json
{
  "image": "base64-encoded-image-data",
  "prompt": "Describe this image in detail",
  "model": "gpt-4-vision-preview"
}
```

**Response:**
```json
{
  "id": "vision-abc123",
  "analysis": "The image shows a cat sitting on a windowsill...",
  "objects_detected": ["cat", "window", "plant"],
  "confidence": 0.95
}
```

---

### Background Tasks

**POST /v1/tasks**

Create a background task.

**Request:**
```json
{
  "command": "analyze_codebase",
  "parameters": {
    "path": "/path/to/code",
    "depth": 3
  },
  "priority": "normal",
  "timeout": 300000
}
```

**Response:**
```json
{
  "id": "task_abc123",
  "status": "pending",
  "created_at": "2024-01-15T10:30:00Z"
}
```

**GET /v1/tasks/:id**

Get task status.

**Response:**
```json
{
  "id": "task_abc123",
  "status": "running",
  "progress": 45,
  "message": "Analyzing files...",
  "created_at": "2024-01-15T10:30:00Z",
  "started_at": "2024-01-15T10:30:05Z"
}
```

**GET /v1/tasks/:id/events**

Stream task events via SSE.

**Response (SSE):**
```
event: task.progress
id: 1
data: {"taskId":"task_abc123","progress":45,"message":"Analyzing..."}

event: task.completed
id: 2
data: {"taskId":"task_abc123","result":{...},"duration":5200}
```

---

### Debates

**POST /v1/debates**

Start an AI debate session.

**Request:**
```json
{
  "topic": "Should we use microservices or monolith?",
  "participants": [
    {"name": "Claude", "role": "advocate"},
    {"name": "Gemini", "role": "critic"},
    {"name": "DeepSeek", "role": "synthesizer"}
  ],
  "rounds": 3,
  "enable_multi_pass_validation": true,
  "validation_config": {
    "enable_validation": true,
    "enable_polish": true,
    "show_phase_indicators": true
  }
}
```

**Response:**
```json
{
  "id": "debate_abc123",
  "status": "started",
  "topic": "Should we use microservices or monolith?",
  "current_round": 1,
  "total_rounds": 3
}
```

**GET /v1/debates/:id**

Get debate status and results.

**Response:**
```json
{
  "id": "debate_abc123",
  "status": "completed",
  "topic": "Should we use microservices or monolith?",
  "rounds": [
    {
      "number": 1,
      "positions": [
        {
          "participant": "Claude",
          "role": "advocate",
          "argument": "Microservices offer better scalability...",
          "confidence": 0.85
        }
      ]
    }
  ],
  "consensus": "A hybrid approach is recommended...",
  "confidence": 0.92,
  "multi_pass_result": {
    "phases_completed": 3,
    "overall_confidence": 0.95,
    "quality_improvement": 0.12
  }
}
```

**GET /v1/debates/:id/events**

Stream debate events via SSE.

---

### Models

**GET /v1/models**

List available models.

**Response:**
```json
{
  "object": "list",
  "data": [
    {
      "id": "helixagent-debate",
      "object": "model",
      "owned_by": "helixagent",
      "capabilities": ["vision", "streaming", "function_calls", "embeddings"]
    }
  ]
}
```

---

### Health

**GET /health**

Health check endpoint.

**Response:**
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime": 3600,
  "providers": {
    "claude": {"status": "healthy", "latency": 120},
    "gemini": {"status": "healthy", "latency": 150},
    "deepseek": {"status": "healthy", "latency": 180}
  }
}
```

---

## MCP Protocol

### Tool Definitions

**helix_chat**

Send a message to the AI Debate Ensemble.

```json
{
  "name": "helix_chat",
  "description": "Chat with HelixAgent AI Debate Ensemble",
  "inputSchema": {
    "type": "object",
    "properties": {
      "message": {
        "type": "string",
        "description": "The message to send"
      },
      "enableDebate": {
        "type": "boolean",
        "description": "Enable multi-agent debate",
        "default": true
      },
      "temperature": {
        "type": "number",
        "description": "Sampling temperature",
        "minimum": 0,
        "maximum": 2,
        "default": 0.7
      }
    },
    "required": ["message"]
  }
}
```

**helix_embeddings**

Generate embeddings for text.

```json
{
  "name": "helix_embeddings",
  "description": "Generate vector embeddings",
  "inputSchema": {
    "type": "object",
    "properties": {
      "texts": {
        "type": "array",
        "items": {"type": "string"},
        "description": "Texts to embed"
      },
      "model": {
        "type": "string",
        "description": "Embedding model",
        "default": "text-embedding-3-small"
      }
    },
    "required": ["texts"]
  }
}
```

**helix_vision**

Analyze images.

```json
{
  "name": "helix_vision",
  "description": "Analyze images using AI",
  "inputSchema": {
    "type": "object",
    "properties": {
      "image": {
        "type": "string",
        "description": "Base64 encoded image or URL"
      },
      "prompt": {
        "type": "string",
        "description": "Analysis prompt"
      }
    },
    "required": ["image"]
  }
}
```

**helix_debate**

Start a multi-agent debate.

```json
{
  "name": "helix_debate",
  "description": "Start AI debate session",
  "inputSchema": {
    "type": "object",
    "properties": {
      "topic": {
        "type": "string",
        "description": "Debate topic"
      },
      "rounds": {
        "type": "integer",
        "description": "Number of debate rounds",
        "default": 3
      },
      "enableValidation": {
        "type": "boolean",
        "description": "Enable multi-pass validation",
        "default": true
      }
    },
    "required": ["topic"]
  }
}
```

**helix_task**

Create a background task.

```json
{
  "name": "helix_task",
  "description": "Create background task",
  "inputSchema": {
    "type": "object",
    "properties": {
      "command": {
        "type": "string",
        "description": "Task command"
      },
      "parameters": {
        "type": "object",
        "description": "Task parameters"
      },
      "timeout": {
        "type": "integer",
        "description": "Timeout in milliseconds"
      }
    },
    "required": ["command"]
  }
}
```

**helix_task_status**

Get task status.

```json
{
  "name": "helix_task_status",
  "description": "Get background task status",
  "inputSchema": {
    "type": "object",
    "properties": {
      "taskId": {
        "type": "string",
        "description": "Task ID"
      }
    },
    "required": ["taskId"]
  }
}
```

---

## Event Types

### Task Events

| Event | Description | Payload |
|-------|-------------|---------|
| `task.created` | Task created | `{taskId, command, priority}` |
| `task.started` | Execution started | `{taskId, workerId, startTime}` |
| `task.progress` | Progress update | `{taskId, percent, message}` |
| `task.heartbeat` | Keepalive | `{taskId, timestamp}` |
| `task.paused` | Task paused | `{taskId, reason}` |
| `task.resumed` | Task resumed | `{taskId}` |
| `task.completed` | Success | `{taskId, result, duration}` |
| `task.failed` | Failure | `{taskId, error, retryCount}` |
| `task.stuck` | Stuck detected | `{taskId, lastActivity}` |
| `task.cancelled` | Cancelled | `{taskId, reason}` |
| `task.retrying` | Retry attempt | `{taskId, attempt, maxRetries}` |
| `task.deadletter` | DLQ'd | `{taskId, error, attempts}` |
| `task.log` | Log output | `{taskId, level, message}` |
| `task.resource` | Resource usage | `{taskId, cpu, memory, io}` |

### Debate Events

| Event | Description | Payload |
|-------|-------------|---------|
| `debate.started` | Debate initiated | `{debateId, topic, participants}` |
| `debate.round_started` | New round | `{debateId, round, totalRounds}` |
| `debate.position_submitted` | Position from LLM | `{debateId, position, participant, confidence}` |
| `debate.validation_phase` | Entering validation | `{debateId, round, positions}` |
| `debate.polish_phase` | Entering polish | `{debateId, round}` |
| `debate.consensus` | Consensus reached | `{debateId, consensus, confidence}` |
| `debate.completed` | Debate finished | `{debateId, result, duration}` |
| `debate.failed` | Debate failed | `{debateId, error}` |

---

## TOON Protocol

### Key Mappings

| Full Key | TOON Code |
|----------|-----------|
| `message` | `m` |
| `messages` | `ms` |
| `role` | `r` |
| `content` | `c` |
| `model` | `mo` |
| `metadata` | `md` |
| `timestamp` | `ts` |
| `confidence` | `cf` |
| `temperature` | `tp` |
| `max_tokens` | `mt` |
| `tools` | `t` |
| `function` | `f` |
| `arguments` | `a` |

### Value Mappings

| Full Value | TOON Code |
|------------|-----------|
| `assistant` | `a` |
| `user` | `u` |
| `system` | `s` |
| `helixagent-debate` | `hd` |
| `function` | `fn` |
| `tool` | `tl` |

### Format

```
# Object: key=value pairs separated by ;
m=Hello;r=u;mo=hd

# Array: items separated by |
item1|item2|item3

# Nested: use : for nesting
m:r=u;c=Hello

# Escaped characters: \; \| \= \:
```

---

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `invalid_api_key` | 401 | API key missing or invalid |
| `rate_limit_exceeded` | 429 | Too many requests |
| `invalid_request` | 400 | Malformed request body |
| `model_not_found` | 404 | Requested model not available |
| `context_length_exceeded` | 400 | Input too long |
| `server_error` | 500 | Internal server error |
| `provider_unavailable` | 503 | LLM provider unavailable |
| `timeout` | 504 | Request timeout |

**Error Response Format:**
```json
{
  "error": {
    "code": "invalid_api_key",
    "message": "The API key provided is invalid.",
    "type": "authentication_error"
  }
}
```

---

## Rate Limits

| Tier | Requests/min | Tokens/min | Concurrent |
|------|--------------|------------|------------|
| Free | 10 | 10,000 | 2 |
| Basic | 60 | 100,000 | 10 |
| Pro | 300 | 500,000 | 50 |
| Enterprise | Unlimited | Unlimited | Unlimited |

**Rate Limit Headers:**
```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1705312260
```

---

## SDK Reference

### TypeScript

```typescript
import { HelixClient, SSEClient, DebateRenderer } from '@helixagent/sdk';

// Initialize client
const client = new HelixClient({
  endpoint: 'http://localhost:7061',
  apiKey: process.env.HELIXAGENT_API_KEY,
});

// Chat
const response = await client.chat('Hello', { enableDebate: true });

// Embeddings
const embeddings = await client.embed(['text1', 'text2']);

// Events
const events = new SSEClient({ endpoint: 'http://localhost:7061/v1/events' });
events.on('debate.consensus', (data) => console.log(data));

// UI
const renderer = new DebateRenderer({ style: 'theater' });
console.log(renderer.render(debateState));
```

### Go

```go
import (
    "github.com/helixagent/sdk-go"
)

// Initialize client
client := sdk.NewClient(sdk.ClientOptions{
    Endpoint: "http://localhost:7061",
    APIKey:   os.Getenv("HELIXAGENT_API_KEY"),
})

// Chat
response, err := client.Chat(ctx, "Hello", &sdk.ChatOptions{
    EnableDebate: true,
})

// Embeddings
embeddings, err := client.Embed(ctx, []string{"text1", "text2"})

// Events
events := sdk.NewSSEClient(client.Endpoint + "/v1/events")
events.On("debate.consensus", func(data json.RawMessage) {
    // Handle event
})

// UI
renderer := sdk.NewDebateRenderer(sdk.DebateRendererOptions{
    Style: sdk.StyleTheater,
})
fmt.Println(renderer.Render(debateState))
```

### Python

```python
from helixagent import HelixClient, SSEClient, DebateRenderer

# Initialize client
client = HelixClient(
    endpoint="http://localhost:7061",
    api_key=os.environ.get("HELIXAGENT_API_KEY"),
)

# Chat
response = await client.chat("Hello", enable_debate=True)

# Embeddings
embeddings = await client.embed(["text1", "text2"])

# Events
events = SSEClient(endpoint="http://localhost:7061/v1/events")
events.on("debate.consensus", lambda data: print(data))

# UI
renderer = DebateRenderer(style="theater")
print(renderer.render(debate_state))
```

---

## Configuration Schema

### Full Configuration

```json
{
  "$schema": "https://helixagent.dev/schema/config.json",
  "version": "1.0",

  "endpoint": "http://localhost:7061",

  "provider": {
    "type": "openai-compatible",
    "name": "helixagent",
    "base_url": "http://localhost:7061/v1",
    "api_key_env": "HELIXAGENT_API_KEY"
  },

  "models": [
    {
      "id": "helixagent-debate",
      "name": "HelixAgent AI Debate Ensemble",
      "max_tokens": 128000,
      "capabilities": ["vision", "streaming", "function_calls", "embeddings"]
    }
  ],

  "mcp": {
    "helixagent-mcp": {"type": "remote", "url": "http://localhost:7061/v1/mcp"},
    "helixagent-acp": {"type": "remote", "url": "http://localhost:7061/v1/acp"},
    "helixagent-lsp": {"type": "remote", "url": "http://localhost:7061/v1/lsp"}
  },

  "transport": {
    "preferHTTP3": true,
    "enableTOON": true,
    "enableBrotli": true,
    "timeout": 30000
  },

  "events": {
    "transport": "sse",
    "reconnectInterval": 5000,
    "subscriptions": ["task.*", "debate.*"]
  },

  "ui": {
    "renderStyle": "theater",
    "progressStyle": "unicode",
    "colorScheme": "256"
  },

  "settings": {
    "streaming": true,
    "autoApprove": false
  }
}
```
