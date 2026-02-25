# HelixAgent API Documentation

**Version:** 1.3.0  
**Base URL:** `http://localhost:7061`  
**Protocol:** HTTP/3 (QUIC) with Brotli compression

---

## Table of Contents

1. [Authentication](#authentication)
2. [LLM Endpoints](#llm-endpoints)
3. [Debate Endpoints](#debate-endpoints)
4. [Provider Endpoints](#provider-endpoints)
5. [MCP Endpoints](#mcp-endpoints)
6. [Monitoring Endpoints](#monitoring-endpoints)
7. [Embeddings Endpoints](#embeddings-endpoints)
8. [Memory Endpoints](#memory-endpoints)
9. [RAG Endpoints](#rag-endpoints)
10. [Error Handling](#error-handling)

---

## Authentication

All API requests require authentication via Bearer token or API key.

### Headers

```http
Authorization: Bearer <your-api-key>
Content-Type: application/json
```

### API Key Management

```http
POST /v1/api-keys
GET /v1/api-keys
DELETE /v1/api-keys/{key_id}
```

---

## LLM Endpoints

### Chat Completion

**POST** `/v1/chat/completions`

OpenAI-compatible chat completion endpoint.

**Request Body:**

```json
{
  "model": "string (required)",
  "messages": [
    {
      "role": "system|user|assistant",
      "content": "string"
    }
  ],
  "temperature": "number (0-2, default: 1)",
  "max_tokens": "integer (default: 4096)",
  "stream": "boolean (default: false)",
  "provider": "string (optional)",
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "string",
        "description": "string",
        "parameters": {}
      }
    }
  ]
}
```

**Response:**

```json
{
  "id": "string",
  "object": "chat.completion",
  "created": 1234567890,
  "model": "string",
  "provider": "string",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "string"
      },
      "finish_reason": "stop|length|tool_calls"
    }
  ],
  "usage": {
    "prompt_tokens": 100,
    "completion_tokens": 50,
    "total_tokens": 150
  }
}
```

### Stream Chat Completion

**POST** `/v1/chat/completions` with `stream: true`

Returns Server-Sent Events (SSE):

```
data: {"id":"chunk-1","choices":[{"delta":{"content":"Hello"}}]}

data: {"id":"chunk-2","choices":[{"delta":{"content":" world"}}]}

data: [DONE]
```

### List Models

**GET** `/v1/models`

**Response:**

```json
{
  "object": "list",
  "data": [
    {
      "id": "gpt-4",
      "object": "model",
      "created": 1234567890,
      "owned_by": "openai",
      "provider": "openai"
    }
  ]
}
```

---

## Debate Endpoints

### Conduct Debate

**POST** `/v1/debate/conduct`

**Request Body:**

```json
{
  "topic": "string (required)",
  "context": "string (optional)",
  "config": {
    "max_rounds": "integer (default: 3)",
    "participants": ["string"],
    "voting_method": "weighted|majority|borda|condorcet",
    "model": "string (default: provider default)"
  }
}
```

**Response:**

```json
{
  "debate_id": "string",
  "status": "completed|in_progress|failed",
  "result": {
    "winner": "string",
    "consensus": "string",
    "turns": [
      {
        "round": 1,
        "participant": "string",
        "content": "string"
      }
    ]
  },
  "scores": {
    "participant_name": 0.95
  }
}
```

### Get Debate Status

**GET** `/v1/debate/{debate_id}/status`

### Get Debate Result

**GET** `/v1/debate/{debate_id}/result`

### List Debates

**GET** `/v1/debates`

---

## Provider Endpoints

### List Providers

**GET** `/v1/providers`

**Response:**

```json
{
  "providers": [
    {
      "id": "openai",
      "name": "OpenAI",
      "enabled": true,
      "models": ["gpt-4", "gpt-3.5-turbo"],
      "score": 95.5,
      "status": "healthy"
    }
  ]
}
```

### Get Provider

**GET** `/v1/providers/{provider_id}`

### Provider Health Check

**GET** `/v1/providers/{provider_id}/health`

---

## MCP Endpoints

### Execute MCP Tool

**POST** `/v1/mcp`

**Request Body:**

```json
{
  "server": "string (required)",
  "tool": "string (required)",
  "arguments": {}
}
```

### List MCP Servers

**GET** `/v1/mcp/servers`

### List MCP Tools

**GET** `/v1/mcp/tools`

---

## Monitoring Endpoints

### System Status

**GET** `/v1/monitoring/status`

**Response:**

```json
{
  "status": "healthy",
  "version": "1.3.0",
  "uptime_seconds": 3600,
  "providers": {
    "healthy": 20,
    "total": 22
  },
  "memory": {
    "alloc_mb": 100,
    "sys_mb": 200
  },
  "goroutines": 50
}
```

### Circuit Breaker Status

**GET** `/v1/monitoring/circuit-breakers`

### Provider Health

**GET** `/v1/monitoring/provider-health`

### Metrics (Prometheus)

**GET** `/metrics`

---

## Embeddings Endpoints

### Generate Embeddings

**POST** `/v1/embeddings`

**Request Body:**

```json
{
  "model": "text-embedding-3-small",
  "input": "string or string[]",
  "encoding_format": "float|base64"
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
      "embedding": [0.1, 0.2, 0.3]
    }
  ],
  "model": "text-embedding-3-small",
  "usage": {
    "prompt_tokens": 10,
    "total_tokens": 10
  }
}
```

---

## Memory Endpoints

### Add Memory

**POST** `/v1/memory`

**Request Body:**

```json
{
  "content": "string",
  "metadata": {},
  "scope": "user|session|global"
}
```

### Search Memory

**GET** `/v1/memory/search?q={query}`

### Get Memory

**GET** `/v1/memory/{memory_id}`

### Delete Memory

**DELETE** `/v1/memory/{memory_id}`

---

## RAG Endpoints

### Ingest Document

**POST** `/v1/rag/ingest`

**Request Body:**

```json
{
  "content": "string",
  "metadata": {},
  "chunk_size": 512,
  "chunk_overlap": 50
}
```

### Search Documents

**POST** `/v1/rag/search`

**Request Body:**

```json
{
  "query": "string",
  "top_k": 10,
  "filters": {}
}
```

### Hybrid Search

**POST** `/v1/rag/hybrid-search`

**Request Body:**

```json
{
  "query": "string",
  "alpha": 0.7,
  "top_k": 10
}
```

---

## LSP Endpoints

### Get Diagnostics

**GET** `/v1/lsp/diagnostics?file_path={path}`

### Go to Definition

**GET** `/v1/lsp/definition?file_path={path}&line={line}&character={char}`

### Find References

**GET** `/v1/lsp/references?file_path={path}&line={line}&character={char}`

---

## ACP Endpoints

### Send Message to Agent

**POST** `/v1/acp`

**Request Body:**

```json
{
  "agent_id": "string",
  "message": "string"
}
```

### List Agents

**GET** `/v1/acp/agents`

---

## Vision Endpoints

### Analyze Image

**POST** `/v1/vision/analyze`

**Request Body:**

```json
{
  "image_url": "string",
  "prompt": "string (optional)"
}
```

### OCR

**POST** `/v1/vision/ocr`

**Request Body:**

```json
{
  "image_url": "string"
}
```

---

## Formatters Endpoints

### Format Code

**POST** `/v1/format`

**Request Body:**

```json
{
  "code": "string",
  "language": "go|python|javascript|...",
  "formatter": "gofmt|black|prettier|..."
}
```

### List Formatters

**GET** `/v1/formatters`

---

## Error Handling

### Error Response Format

```json
{
  "error": {
    "type": "invalid_request_error|authentication_error|rate_limit_error|provider_error",
    "message": "string",
    "code": "ERROR_CODE",
    "param": "string (optional)",
    "details": {}
  }
}
```

### HTTP Status Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not Found |
| 429 | Rate Limit Exceeded |
| 500 | Internal Server Error |
| 502 | Provider Error |
| 503 | Service Unavailable |

### Error Types

- `invalid_request_error` - Malformed request
- `authentication_error` - Invalid or missing API key
- `rate_limit_error` - Too many requests
- `provider_error` - Upstream provider error
- `model_not_found` - Requested model doesn't exist
- `context_length_exceeded` - Input too long

---

## Rate Limiting

Default rate limits:

- **Authenticated:** 1000 requests/hour
- **Anonymous:** 100 requests/hour

Rate limit headers:

```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1234567890
```

---

## Streaming

All streaming endpoints use Server-Sent Events (SSE):

```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
```

---

## WebSocket

WebSocket endpoint for real-time updates:

```
ws://localhost:7061/ws
```

### Message Format

```json
{
  "type": "chat|debate|monitoring",
  "action": "subscribe|unsubscribe|message",
  "data": {}
}
```

---

## OpenAPI Specification

Full OpenAPI 3.0 specification available at:

```
GET /v1/openapi.json
GET /v1/openapi.yaml
```

---

## SDK Examples

### Go

```go
import "github.com/vasic-digital/helixagent-sdk"

client := helixagent.NewClient("your-api-key")
response, err := client.Chat.Completions.Create(ctx, &helixagent.ChatRequest{
    Model: "gpt-4",
    Messages: []helixagent.Message{
        {Role: "user", Content: "Hello!"},
    },
})
```

### Python

```python
import helixagent

client = helixagent.Client(api_key="your-api-key")
response = client.chat.completions.create(
    model="gpt-4",
    messages=[{"role": "user", "content": "Hello!"}]
)
```

### cURL

```bash
curl -X POST http://localhost:7061/v1/chat/completions \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

---

## Versioning

API version is included in the URL path: `/v1/...`

Current version: **v1**

---

## Changelog

### v1.3.0 (2026-02-25)
- Added Kimi Code provider
- Added Qwen Code provider
- Added profiling tools
- Enhanced security scanning
- Memory leak detection
- Deadlock prevention

### v1.2.0 (2026-02-20)
- Added 14 new LLM providers
- Enhanced debate orchestration
- Improved streaming support

### v1.1.0 (2026-02-15)
- Added MCP support
- Added RAG pipeline
- Added memory system

### v1.0.0 (2026-01-01)
- Initial release
- Core LLM integration
- Basic API endpoints
