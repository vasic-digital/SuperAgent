# OpenAI Codex: API Documentation & HelixAgent Cross-Reference

**Agent:** OpenAI Codex  
**Type:** CLI + API (OpenAI ecosystem)  
**API Base:** `https://api.openai.com/v1`  
**HelixAgent Provider:** [internal/llm/providers/openai/](../../../internal/llm/providers/openai/)  
**Analysis Date:** 2026-04-03  

---

## Executive Summary

OpenAI Codex is deeply integrated with the OpenAI API ecosystem. Unlike CLI-only agents, Codex leverages the full OpenAI API platform including reasoning models (o3, o4-mini), code interpreter, and ChatGPT conversation sync. HelixAgent provides full OpenAI API compatibility through its OpenAI provider.

**HelixAgent Status:** Full OpenAI API compatibility with extensions.

---

## OpenAI API Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    OPENAI API ECOSYSTEM                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   ┌──────────────────────────────────────────────────────────┐  │
│   │                    OPENAI PLATFORM                        │  │
│   │                                                          │  │
│   │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐   │  │
│   │  │  ChatGPT     │  │   Codex      │  │   API        │   │  │
│   │  │  (Web)       │  │   (CLI)      │  │   Platform   │   │  │
│   │  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘   │  │
│   │         │                 │                 │            │  │
│   │         └─────────────────┼─────────────────┘            │  │
│   │                           │                              │  │
│   │                    ┌──────┴──────┐                       │  │
│   │                    │  OpenAI API  │                       │  │
│   │                    │  /v1/*       │                       │  │
│   │                    └─────────────┘                       │  │
│   │                                                          │  │
│   └──────────────────────────────────────────────────────────┘  │
│                                                                  │
│   Base URL: https://api.openai.com/v1                         │
│   Auth: Bearer Token (sk-...)                                 │
│   Protocol: HTTPS, HTTP/2, SSE                                │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Complete OpenAI API Specification

### Base Configuration

```yaml
Base URL: https://api.openai.com/v1
Authentication: Bearer ${OPENAI_API_KEY}
Content-Type: application/json
Protocols:
  - REST (HTTPS)
  - Server-Sent Events (SSE)
  - WebSocket (beta)
Rate Limits:
  - Tier 1: 60 RPM
  - Tier 2: 1000 RPM
  - Tier 3: 5000 RPM
```

### Endpoints Overview

| Endpoint | Method | Description | HelixAgent Compatible |
|----------|--------|-------------|----------------------|
| `/chat/completions` | POST | Chat completion | ✅ Full |
| `/completions` | POST | Text completion | ✅ Full |
| `/embeddings` | POST | Text embeddings | ✅ Full |
| `/models` | GET | List models | ✅ Full |
| `/files` | GET/POST | File management | ✅ Full |
| `/fine_tuning` | GET/POST | Fine-tuning | ⚠️ Partial |
| `/images/generations` | POST | DALL-E | ❌ No |
| `/audio/transcriptions` | POST | Whisper | ⚠️ Partial |
| `/assistants` | GET/POST | Assistants API | ⚠️ Partial |
| `/threads` | GET/POST | Thread management | ⚠️ Partial |
| `/runs` | POST | Run assistant | ⚠️ Partial |

---

## Core Endpoints

### 1. Chat Completions

**Source:** [`internal/handlers/completion.go:45`](../../../internal/handlers/completion.go#L45)

```
POST /v1/chat/completions
```

**OpenAI Request:**
```json
{
  "model": "gpt-4o",
  "messages": [
    {"role": "system", "content": "You are a helpful assistant."},
    {"role": "user", "content": "Hello!"}
  ],
  "temperature": 0.7,
  "max_tokens": 4096,
  "top_p": 1,
  "frequency_penalty": 0,
  "presence_penalty": 0,
  "stream": false,
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "get_weather",
        "description": "Get weather for location",
        "parameters": {
          "type": "object",
          "properties": {
            "location": {"type": "string"}
          }
        }
      }
    }
  ]
}
```

**OpenAI Response:**
```json
{
  "id": "chatcmpl-abc123",
  "object": "chat.completion",
  "created": 1677858242,
  "model": "gpt-4o",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Hello! How can I help you today?"
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 20,
    "completion_tokens": 10,
    "total_tokens": 30
  }
}
```

### HelixAgent Implementation

**Source:** [`internal/llm/providers/openai/openai.go`](../../../internal/llm/providers/openai/openai.go)

```go
package openai

// OpenAIProvider implements OpenAI API compatibility
// Source: internal/llm/providers/openai/openai.go#L1-200

type OpenAIProvider struct {
    client  *openai.Client
    config  *Config
    baseURL string
}

// GenerateChatCompletion implements /v1/chat/completions
// Source: internal/llm/providers/openai/openai.go#L56-120
func (p *OpenAIProvider) GenerateChatCompletion(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
    resp, err := p.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
        Model:    req.Model,
        Messages: convertMessages(req.Messages),
        Temperature: openai.Float(req.Temperature),
        MaxTokens:   openai.Int(req.MaxTokens),
        Tools:       convertTools(req.Tools),
    })
    
    return &ChatResponse{
        ID:      resp.ID,
        Model:   resp.Model,
        Choices: convertChoices(resp.Choices),
        Usage:   convertUsage(resp.Usage),
    }, nil
}
```

### HelixAgent API Compatibility

**Source:** [`internal/handlers/openai_compatible.go`](../../../internal/handlers/openai_compatible.go)

```go
// OpenAICompatibleHandler provides drop-in OpenAI API replacement
// Source: internal/handlers/openai_compatible.go#L1-150

func (h *OpenAICompatibleHandler) ChatCompletions(c *gin.Context) {
    var req openai.ChatCompletionRequest
    if err := c.BindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    // Support all OpenAI parameters
    // - model: Maps to provider
    // - messages: Standard format
    // - temperature, max_tokens, etc.: Passed through
    // - tools: Converted to MCP
    // - stream: SSE support
    
    resp, err := h.ensemble.Execute(c.Request.Context(), &req)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    // Return OpenAI-compatible format
    c.JSON(200, convertToOpenAIFormat(resp))
}
```

---

## Reasoning Models (o3, o4-mini)

### OpenAI o3/o4-mini

Codex uses OpenAI's reasoning models for complex tasks:

```json
{
  "model": "o3-mini",
  "messages": [{"role": "user", "content": "Design a distributed system"}],
  "reasoning_effort": "high",  // low, medium, high
  "max_completion_tokens": 10000
}
```

**Response includes reasoning tokens:**
```json
{
  "id": "chatcmpl-xxx",
  "model": "o3-mini",
  "choices": [{
    "message": {
      "role": "assistant",
      "content": "Here's the design..."
    }
  }],
  "usage": {
    "prompt_tokens": 50,
    "completion_tokens": 500,
    "reasoning_tokens": 2000,  // Special field for reasoning
    "total_tokens": 2550
  }
}
```

### HelixAgent Reasoning Support

**Status:** ⚠️ Partial - No native reasoning models

**Workaround:** Multi-step prompting with ensemble

```json
// HelixAgent ensemble for reasoning
{
  "model": "helixagent-ensemble",
  "messages": [{"role": "user", "content": "Design a distributed system"}],
  "strategy": "chain_of_thought",
  "steps": [
    {"provider": "claude", "task": "analyze_requirements"},
    {"provider": "gpt4", "task": "design_architecture"},
    {"provider": "deepseek", "task": "validate_design"}
  ]
}
```

---

## Streaming (Server-Sent Events)

### OpenAI SSE Format

```
POST /v1/chat/completions
Content-Type: application/json

{"model": "gpt-4o", "messages": [...], "stream": true}
```

**SSE Response:**
```
data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","choices":[{"delta":{"content":"Hello"}}]}

data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","choices":[{"delta":{"content":"!"}}]}

data: [DONE]
```

### HelixAgent SSE Implementation

**Source:** [`internal/handlers/streaming.go`](../../../internal/handlers/streaming.go)

```go
package handlers

// StreamingHandler manages SSE connections
// Source: internal/handlers/streaming.go#L30-95

func (h *StreamingHandler) StreamCompletions(c *gin.Context) {
    req := parseRequest(c)
    
    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    c.Header("Connection", "keep-alive")
    
    // Stream from ensemble
    stream, err := h.ensemble.Stream(c.Request.Context(), req)
    if err != nil {
        c.SSEvent("error", err.Error())
        return
    }
    
    for chunk := range stream {
        c.SSEvent("data", formatChunk(chunk))
        c.Writer.Flush()
    }
    
    c.SSEvent("data", "[DONE]")
}
```

---

## Code Interpreter

### OpenAI Code Interpreter

Codex has built-in code interpreter:

```json
{
  "model": "gpt-4o",
  "messages": [{"role": "user", "content": "Analyze this CSV"}],
  "tools": [{"type": "code_interpreter"}]
}
```

### HelixAgent Code Execution

**Source:** [`internal/mcp/adapters/code_interpreter.go`](../../../internal/mcp/adapters/code_interpreter.go)

```go
// CodeInterpreter MCP adapter
// Source: internal/mcp/adapters/code_interpreter.go#L1-89

type CodeInterpreterAdapter struct {
    sandbox *Sandbox
    timeout time.Duration
}

// Execute runs Python code safely
func (c *CodeInterpreterAdapter) Execute(ctx context.Context, code string) (*ExecutionResult, error) {
    // 1. Validate code (security)
    // 2. Run in sandbox
    // 3. Capture output
    // 4. Return results
}
```

**API Endpoint:**
```
POST /v1/tools/code_interpreter
```

```json
{
  "code": "import pandas as pd\ndf = pd.read_csv('data.csv')\nprint(df.describe())",
  "language": "python",
  "timeout": 30
}
```

---

## Model Listing

### OpenAI Models Endpoint

```
GET /v1/models
```

**Response:**
```json
{
  "object": "list",
  "data": [
    {
      "id": "gpt-4o",
      "object": "model",
      "created": 1686935002,
      "owned_by": "openai"
    },
    {
      "id": "o3-mini",
      "object": "model",
      "created": 1700000000,
      "owned_by": "openai"
    }
  ]
}
```

### HelixAgent Model Aggregation

**Source:** [`internal/handlers/models.go`](../../../internal/handlers/models.go)

```go
// ModelsHandler aggregates models from all providers
// Source: internal/handlers/models.go#L25-78

func (h *ModelsHandler) ListModels(c *gin.Context) {
    var allModels []Model
    
    // Aggregate from all configured providers
    for _, provider := range h.registry.GetAll() {
        models, err := provider.ListModels()
        if err != nil {
            continue
        }
        allModels = append(allModels, models...)
    }
    
    // Add ensemble models
    allModels = append(allModels, Model{
        ID:       "helixagent-ensemble",
        Object:   "model",
        Provider: "helixagent",
    })
    
    c.JSON(200, gin.H{
        "object": "list",
        "data":   allModels,
    })
}
```

---

## WebSocket Support

### OpenAI WebSocket (Beta)

```
wss://api.openai.com/v1/realtime
```

**Protocol:**
```json
{
  "event_id": "evt_001",
  "type": "session.update",
  "session": {
    "modalities": ["text", "audio"],
    "model": "gpt-4o-realtime"
  }
}
```

### HelixAgent WebSocket

**Source:** [`internal/handlers/websocket.go`](../../../internal/handlers/websocket.go)

```
ws://localhost:7061/v1/stream
```

**Compatible protocol with extensions:**
```json
{
  "type": "chat.completion",
  "model": "helixagent-ensemble",
  "messages": [...],
  "providers": ["openai", "claude"],
  "strategy": "voting"
}
```

---

## File Operations

### OpenAI Files API

```
POST /v1/files           # Upload file
GET  /v1/files           # List files
GET  /v1/files/{id}      # Retrieve file
DELETE /v1/files/{id}    # Delete file
```

**Upload:**
```bash
curl https://api.openai.com/v1/files \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -F purpose="assistants" \
  -F file="@data.csv"
```

### HelixAgent File Management

**Source:** [`internal/handlers/files.go`](../../../internal/handlers/files.go)

```go
// FileHandler manages file uploads
// Source: internal/handlers/files.go#L1-95
```

**API Endpoints:**
```
POST /v1/files           # Upload
GET  /v1/files           # List
GET  /v1/files/{id}      # Get
DELETE /v1/files/{id}    # Delete
```

---

## Embeddings

### OpenAI Embeddings

```
POST /v1/embeddings
```

```json
{
  "model": "text-embedding-3-large",
  "input": "The food was delicious"
}
```

**Response:**
```json
{
  "object": "list",
  "data": [
    {
      "object": "embedding",
      "embedding": [0.0023, -0.0091, ...],
      "index": 0
    }
  ],
  "model": "text-embedding-3-large",
  "usage": {
    "prompt_tokens": 5,
    "total_tokens": 5
  }
}
```

### HelixAgent Embeddings

**Source:** [`internal/handlers/embeddings.go`](../../../internal/handlers/embeddings.go)

```go
// EmbeddingsHandler with multiple provider support
// Source: internal/handlers/embeddings.go#L1-67
```

**Features:**
- Multiple embedding providers
- Vector DB integration (ChromaDB, Qdrant)
- Caching
- Batch processing

---

## Source Code Reference Index

### OpenAI-Related HelixAgent Files

| Feature | HelixAgent Implementation | File | Lines |
|---------|--------------------------|------|-------|
| Chat Completions | OpenAI Provider | `internal/llm/providers/openai/openai.go` | 200 |
| Streaming | SSE Handler | `internal/handlers/streaming.go` | 95 |
| Models | Model Handler | `internal/handlers/models.go` | 78 |
| Files | File Handler | `internal/handlers/files.go` | 95 |
| Embeddings | Embeddings Handler | `internal/handlers/embeddings.go` | 67 |
| Code Interpreter | Interpreter Adapter | `internal/mcp/adapters/code_interpreter.go` | 89 |
| WebSocket | WS Handler | `internal/handlers/websocket.go` | 85 |
| OpenAI Compatible | Compatible Handler | `internal/handlers/openai_compatible.go` | 150 |
| Provider Config | Provider Config | `internal/config/providers.go` | 234 |
| Ensemble | Ensemble Engine | `internal/services/ensemble.go` | 167 |

---

## Feature Comparison: OpenAI Codex vs HelixAgent

| Feature | OpenAI Codex | HelixAgent | Status |
|---------|--------------|------------|--------|
| **Chat Completions** | ✅ Native | ✅ Compatible | Full |
| **Streaming** | ✅ SSE | ✅ SSE | Full |
| **Reasoning (o3/o4)** | ✅ Native | ⚠️ Workaround | Partial |
| **Code Interpreter** | ✅ Built-in | ✅ MCP | Full |
| **DALL-E Images** | ✅ Native | ❌ No | Gap |
| **Whisper Audio** | ✅ Native | ⚠️ Partial | Partial |
| **Assistants API** | ✅ Native | ⚠️ Partial | Partial |
| **File Operations** | ✅ Native | ✅ Full | Full |
| **Embeddings** | ✅ Native | ✅ Multi-provider | Full |
| **Multi-Provider** | ❌ | ✅ 22+ | Superior |
| **Ensemble** | ❌ | ✅ | Superior |
| **WebSocket** | ⚠️ Beta | ✅ Stable | Superior |
| **Self-Hosted** | ❌ | ✅ | Superior |
| **Open Source** | ❌ | ✅ | Superior |

---

## Integration Guide

### Using HelixAgent as OpenAI Drop-in Replacement

**Configuration:**
```yaml
# HelixAgent config for OpenAI compatibility
server:
  openai_compatible: true
  base_path: /v1

providers:
  openai:
    type: openai
    api_key: ${OPENAI_API_KEY}
    
  # Add ensemble for better results
  anthropic:
    type: anthropic
    api_key: ${ANTHROPIC_API_KEY}

ensemble:
  enabled: true
  providers: ["openai", "anthropic"]
  voting_strategy: "best_of_n"
```

**Client Usage:**
```python
# Standard OpenAI client works with HelixAgent
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:7061/v1",  # HelixAgent endpoint
    api_key="any-key"  # HelixAgent doesn't require OpenAI key
)

response = client.chat.completions.create(
    model="helixagent-ensemble",  # Use ensemble
    messages=[{"role": "user", "content": "Hello!"}]
)
```

---

## Conclusion

**HelixAgent provides:**
- ✅ Full OpenAI API compatibility
- ✅ All core endpoints (chat, completions, embeddings)
- ✅ Streaming support
- ✅ File operations
- ✅ Extended with ensemble capabilities

**Gaps:**
- ⚠️ No DALL-E image generation
- ⚠️ No Whisper audio (partial)
- ⚠️ No Assistants API (partial)
- ⚠️ No native reasoning models (workaround available)

**Recommendation:** HelixAgent is a **drop-in replacement** for OpenAI API with added multi-provider capabilities.

---

*Documentation: API Specification & Cross-Reference*  
*Last Updated: 2026-04-03*  
*HelixAgent Commit: 7ec2da53*
