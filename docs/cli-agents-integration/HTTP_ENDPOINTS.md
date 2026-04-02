# HTTP Endpoints Reference

## Base URL

All endpoints are prefixed with:
```
http://localhost:7061/v1
```

## Authentication

Most endpoints require authentication via Bearer token:
```
Authorization: Bearer <HELIXAGENT_API_KEY>
```

## Core Endpoints

### Chat Completions

**Endpoint:** `POST /chat/completions`

**Description:** Generate chat completions using AI Debate Ensemble

**Request:**
```json
{
  "model": "helixagent-debate",
  "messages": [
    {"role": "system", "content": "You are a helpful assistant."},
    {"role": "user", "content": "Hello!"}
  ],
  "stream": true,
  "max_tokens": 4096,
  "temperature": 0.7
}
```

**Response:**
```json
{
  "id": "chatcmpl-123",
  "object": "chat.completion",
  "created": 1677652288,
  "model": "helixagent-debate",
  "choices": [{
    "index": 0,
    "message": {
      "role": "assistant",
      "content": "Hello! How can I help you today?"
    },
    "finish_reason": "stop"
  }]
}
```

### MCP Tool Call

**Endpoint:** `POST /mcp`

**Description:** Execute MCP tool calls

**Request:**
```json
{
  "server": "filesystem",
  "tool": "read_file",
  "arguments": {
    "path": "/path/to/file.txt"
  }
}
```

**Response:**
```json
{
  "content": [
    {"type": "text", "text": "file contents"}
  ]
}
```

### Embeddings

**Endpoint:** `POST /embeddings`

**Description:** Generate text embeddings

**Request:**
```json
{
  "input": "text to embed",
  "model": "helixagent-embedding"
}
```

**Response:**
```json
{
  "object": "list",
  "data": [{
    "object": "embedding",
    "embedding": [0.0023, -0.0091, ...],
    "index": 0
  }]
}
```

### LSP (Language Server Protocol)

**Endpoint:** `POST /lsp`

**Description:** LSP operations

**Request:**
```json
{
  "method": "textDocument/completion",
  "params": {
    "textDocument": {"uri": "file:///path/to/file.py"},
    "position": {"line": 10, "character": 15}
  }
}
```

**Response:**
```json
{
  "items": [
    {"label": "function_name", "kind": 3}
  ]
}
```

### ACP (Agent Communication Protocol)

**Endpoint:** `POST /acp`

**Description:** Inter-agent communication

**Request:**
```json
{
  "action": "send_message",
  "target_agent": "agent-123",
  "message": "Hello from another agent!"
}
```

**Response:**
```json
{
  "status": "delivered",
  "message_id": "msg-456"
}
```

### Vision

**Endpoint:** `POST /vision`

**Description:** Image analysis

**Request:**
```json
{
  "image": "base64encodedimage...",
  "prompt": "What's in this image?"
}
```

**Response:**
```json
{
  "description": "The image shows a sunset over mountains..."
}
```

### RAG Retrieval

**Endpoint:** `POST /rag`

**Description:** Retrieve context from knowledge base

**Request:**
```json
{
  "query": "How does authentication work?",
  "top_k": 5
}
```

**Response:**
```json
{
  "results": [
    {"text": "Authentication uses JWT tokens...", "score": 0.95}
  ]
}
```

### Code Formatting

**Endpoint:** `POST /format`

**Description:** Format code using 32+ formatters

**Request:**
```json
{
  "code": "def hello():\n    print('world')",
  "language": "python",
  "formatter": "ruff"
}
```

**Response:**
```json
{
  "formatted_code": "def hello():\n    print(\"world\")\n"
}
```

### Cognee (Memory)

**Endpoint:** `POST /cognee`

**Description:** Memory/knowledge graph operations

**Request:**
```json
{
  "action": "remember",
  "content": "User prefers Python for data processing"
}
```

**Response:**
```json
{
  "status": "stored",
  "memory_id": "mem-789"
}
```

## Utility Endpoints

### Health Check

**Endpoint:** `GET /health`

**Response:**
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "services": {
    "debate": "up",
    "mcp": "up",
    "lsp": "up"
  }
}
```

### Models List

**Endpoint:** `GET /models`

**Response:**
```json
{
  "object": "list",
  "data": [
    {
      "id": "helixagent-debate",
      "object": "model",
      "created": 1677610602,
      "owned_by": "helixagent"
    }
  ]
}
```

### Debate Status

**Endpoint:** `GET /debate/status/{debate_id}`

**Response:**
```json
{
  "debate_id": "deb-123",
  "status": "completed",
  "participants": ["claude", "gpt-4", "gemini"],
  "consensus": "achieved"
}
```

## Error Responses

### 400 Bad Request
```json
{
  "error": {
    "message": "Invalid request format",
    "type": "invalid_request_error",
    "code": "invalid_json"
  }
}
```

### 401 Unauthorized
```json
{
  "error": {
    "message": "Invalid API key",
    "type": "authentication_error",
    "code": "invalid_api_key"
  }
}
```

### 500 Internal Server Error
```json
{
  "error": {
    "message": "Internal server error",
    "type": "server_error",
    "code": "internal_error"
  }
}
```

## Rate Limits

| Endpoint | Rate Limit |
|----------|-----------|
| /chat/completions | 100/minute |
| /mcp | 200/minute |
| /embeddings | 300/minute |
| Others | 500/minute |

## Streaming

For streaming responses, set `stream: true` in the request:

```bash
curl -N http://localhost:7061/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $HELIXAGENT_API_KEY" \
  -d '{
    "model": "helixagent-debate",
    "messages": [{"role": "user", "content": "Hello"}],
    "stream": true
  }'
```

---

**Last Updated:** 2026-04-02
