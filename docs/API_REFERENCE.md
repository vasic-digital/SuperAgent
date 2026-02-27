# HelixAgent API Documentation

## Overview

HelixAgent provides a comprehensive REST API for LLM aggregation, ensemble processing, memory management, and debate orchestration. All endpoints are OpenAI-compatible unless specified otherwise.

**Base URL:** `http://localhost:7061`  
**API Version:** v1  
**Protocol:** HTTP/3 (QUIC) with Brotli compression, HTTP/2 fallback  
**Authentication:** API Key or Bearer Token  

---

## Authentication

### API Key Authentication
Include your API key in the header:
```
X-API-Key: your-api-key-here
```

### Bearer Token Authentication
Include a JWT token in the Authorization header:
```
Authorization: Bearer your-jwt-token-here
```

---

## Core Endpoints

### 1. Chat Completions (OpenAI Compatible)

**Endpoint:** `POST /v1/chat/completions`  
**Description:** Generate chat completions using the ensemble of LLMs

#### Request Body
```json
{
  "model": "helixagent-debate",
  "messages": [
    {"role": "system", "content": "You are a helpful assistant."},
    {"role": "user", "content": "Hello!"}
  ],
  "temperature": 0.7,
  "max_tokens": 1000,
  "stream": false,
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "get_weather",
        "description": "Get weather information",
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

#### Response
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
  }],
  "usage": {
    "prompt_tokens": 9,
    "completion_tokens": 12,
    "total_tokens": 21
  }
}
```

#### Streaming Response
Set `"stream": true` for Server-Sent Events (SSE) streaming:
```
data: {"id":"chatcmpl-123","object":"chat.completion.chunk",...}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk",...}

data: [DONE]
```

---

### 2. Models List (OpenAI Compatible)

**Endpoint:** `GET /v1/models`  
**Description:** List all available models from all providers

#### Response
```json
{
  "object": "list",
  "data": [
    {
      "id": "helixagent-debate",
      "object": "model",
      "created": 1677610602,
      "owned_by": "helixagent"
    },
    {
      "id": "gpt-4",
      "object": "model", 
      "created": 1687882411,
      "owned_by": "openai"
    }
  ]
}
```

---

### 3. Health Check

**Endpoint:** `GET /health`  
**Description:** Basic health check

#### Response
```json
{
  "status": "healthy"
}
```

---

### 4. Detailed Health Check

**Endpoint:** `GET /v1/health`  
**Description:** Enhanced health check with provider status

#### Response
```json
{
  "status": "healthy",
  "providers": {
    "total": 22,
    "healthy": 20,
    "unhealthy": 2
  },
  "timestamp": 1677652288
}
```

---

### 5. Prometheus Metrics

**Endpoint:** `GET /metrics`  
**Description:** Prometheus-compatible metrics endpoint

#### Response
Prometheus exposition format with metrics:
- `helixagent_requests_total` - Total requests
- `helixagent_request_duration_seconds` - Request latency
- `helixagent_provider_health` - Provider health status
- `helixagent_debate_votes_total` - Debate voting metrics

---

## Debate & Ensemble Endpoints

### 6. Start Debate Session

**Endpoint:** `POST /v1/debate/start`  
**Description:** Start a new AI debate session

#### Request Body
```json
{
  "topic": "Best programming language for AI development",
  "participants": 5,
  "rounds": 3,
  "voting_method": "weighted",
  "positions": ["pro_python", "pro_rust", "pro_julia", "pro_cpp", "neutral"]
}
```

#### Response
```json
{
  "session_id": "debate-123-abc",
  "status": "started",
  "participants": [
    {"id": "claude", "position": "pro_python", "model": "claude-sonnet-4"},
    {"id": "gpt4", "position": "pro_rust", "model": "gpt-4"}
  ],
  "current_round": 1
}
```

---

### 7. Get Debate Status

**Endpoint:** `GET /v1/debate/{session_id}/status`  
**Description:** Get current debate status and results

#### Response
```json
{
  "session_id": "debate-123-abc",
  "status": "in_progress",
  "current_round": 2,
  "total_rounds": 3,
  "votes": {
    "round_1": {
      "pro_python": 0.35,
      "pro_rust": 0.25,
      "pro_julia": 0.20,
      "pro_cpp": 0.15,
      "neutral": 0.05
    }
  },
  "transcript": [...]
}
```

---

## Memory Endpoints

### 8. Store Memory

**Endpoint:** `POST /v1/memory/store`  
**Description:** Store information in Mem0 memory system

#### Request Body
```json
{
  "user_id": "user-123",
  "content": "User prefers Python for data science projects",
  "metadata": {
    "category": "preference",
    "confidence": 0.95
  }
}
```

#### Response
```json
{
  "memory_id": "mem-456",
  "status": "stored",
  "entities_extracted": ["Python", "data science"]
}
```

---

### 9. Retrieve Memory

**Endpoint:** `POST /v1/memory/retrieve`  
**Description:** Retrieve relevant memories for context

#### Request Body
```json
{
  "user_id": "user-123",
  "query": "What programming languages does the user prefer?",
  "limit": 5
}
```

#### Response
```json
{
  "memories": [
    {
      "id": "mem-456",
      "content": "User prefers Python for data science projects",
      "relevance": 0.92,
      "timestamp": "2026-02-27T10:00:00Z"
    }
  ]
}
```

---

## MCP (Model Context Protocol) Endpoints

### 10. MCP Invoke

**Endpoint:** `POST /v1/mcp/invoke`  
**Description:** Invoke an MCP server capability

#### Request Body
```json
{
  "server": "filesystem",
  "capability": "read_file",
  "parameters": {
    "path": "/home/user/document.txt"
  }
}
```

#### Response
```json
{
  "result": "File contents here...",
  "server": "filesystem",
  "capability": "read_file"
}
```

---

### 11. List MCP Servers

**Endpoint:** `GET /v1/mcp/servers`  
**Description:** List all available MCP servers

#### Response
```json
{
  "servers": [
    {
      "name": "filesystem",
      "status": "running",
      "port": 9101,
      "capabilities": ["read_file", "write_file", "list_directory"]
    },
    {
      "name": "memory",
      "status": "running", 
      "port": 9102,
      "capabilities": ["store", "retrieve", "search"]
    }
  ]
}
```

---

## ACP (Agent Communication Protocol) Endpoints

### 12. ACP Send Message

**Endpoint:** `POST /v1/acp/send`  
**Description:** Send message to another agent via ACP

#### Request Body
```json
{
  "target_agent": "agent-456",
  "message_type": "task_request",
  "payload": {
    "task": "analyze_code",
    "code": "def foo(): pass"
  },
  "timeout": 30000
}
```

#### Response
```json
{
  "message_id": "acp-msg-789",
  "status": "delivered",
  "response": {
    "analysis": "Function is empty and does nothing"
  }
}
```

---

## Embeddings Endpoints

### 13. Create Embeddings

**Endpoint:** `POST /v1/embeddings`  
**Description:** Create vector embeddings for text (OpenAI compatible)

#### Request Body
```json
{
  "input": "The quick brown fox jumps over the lazy dog",
  "model": "text-embedding-3-small"
}
```

#### Response
```json
{
  "object": "list",
  "data": [
    {
      "object": "embedding",
      "embedding": [0.0023064255, -0.009327292, ...],
      "index": 0
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

## RAG (Retrieval Augmented Generation) Endpoints

### 14. RAG Query

**Endpoint:** `POST /v1/rag/query`  
**Description:** Query the RAG system with hybrid retrieval

#### Request Body
```json
{
  "query": "What are the best practices for microservices?",
  "top_k": 5,
  "filters": {
    "source": "documentation",
    "date_after": "2025-01-01"
  }
}
```

#### Response
```json
{
  "results": [
    {
      "content": "Microservices should be loosely coupled...",
      "source": "docs/architecture.md",
      "score": 0.89,
      "metadata": {...}
    }
  ],
  "total_results": 42,
  "query_time_ms": 150
}
```

---

## Authentication Endpoints

### 15. Register User

**Endpoint:** `POST /v1/auth/register`  
**Description:** Register a new user account

#### Request Body
```json
{
  "username": "johndoe",
  "email": "john@example.com",
  "password": "securepassword123"
}
```

#### Response
```json
{
  "user_id": "user-123",
  "api_key": "ha_abc123xyz789",
  "message": "User registered successfully"
}
```

---

### 16. Login

**Endpoint:** `POST /v1/auth/login`  
**Description:** Authenticate and get JWT token

#### Request Body
```json
{
  "username": "johndoe",
  "password": "securepassword123"
}
```

#### Response
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 86400,
  "api_key": "ha_abc123xyz789"
}
```

---

## Error Handling

### Standard Error Format
```json
{
  "error": {
    "code": "invalid_api_key",
    "message": "The provided API key is invalid or expired",
    "type": "authentication_error"
  }
}
```

### Error Codes
- `invalid_request` - Malformed request
- `authentication_error` - Invalid credentials
- `rate_limit_exceeded` - Too many requests
- `provider_unavailable` - LLM provider down
- `invalid_model` - Model not found
- `context_length_exceeded` - Input too long
- `debate_not_found` - Debate session doesn't exist

---

## Rate Limiting

- **Authenticated requests:** 1000 requests/minute
- **Unauthenticated requests:** 60 requests/minute
- **Streaming requests:** 100 concurrent streams

Rate limit headers included in all responses:
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1677652348
```

---

## SDK Examples

### Python
```python
import requests

headers = {
    "X-API-Key": "your-api-key",
    "Content-Type": "application/json"
}

response = requests.post(
    "http://localhost:7061/v1/chat/completions",
    headers=headers,
    json={
        "model": "helixagent-debate",
        "messages": [{"role": "user", "content": "Hello!"}]
    }
)

print(response.json()["choices"][0]["message"]["content"])
```

### JavaScript
```javascript
const response = await fetch('http://localhost:7061/v1/chat/completions', {
  method: 'POST',
  headers: {
    'X-API-Key': 'your-api-key',
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    model: 'helixagent-debate',
    messages: [{ role: 'user', content: 'Hello!' }]
  })
});

const data = await response.json();
console.log(data.choices[0].message.content);
```

### cURL
```bash
curl -X POST http://localhost:7061/v1/chat/completions \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "helixagent-debate",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

---

## WebSocket Support

For real-time streaming and bidirectional communication:

**Endpoint:** `ws://localhost:7061/v1/stream`  
**Protocol:** JSON messages

### Connect
```javascript
const ws = new WebSocket('ws://localhost:7061/v1/stream');
ws.onopen = () => {
  ws.send(JSON.stringify({
    type: 'subscribe',
    channel: 'debate_updates',
    session_id: 'debate-123'
  }));
};
```

---

## Version History

- **v1.0.0** (2026-02-27) - Initial API release
- **v1.1.0** (2026-02-27) - Added debate endpoints
- **v1.2.0** (2026-02-27) - Added MCP and ACP endpoints
- **v1.3.0** (2026-02-27) - Added RAG and memory endpoints

---

## Support

- **Documentation:** https://docs.helixagent.dev
- **GitHub:** https://github.com/HelixDevelopment/HelixAgent
- **Issues:** https://github.com/HelixDevelopment/HelixAgent/issues
- **Discord:** https://discord.gg/helixagent

---

**Last Updated:** February 27, 2026  
**API Version:** v1.3.0
