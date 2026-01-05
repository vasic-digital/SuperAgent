# SuperAgent API Reference

## Introduction

This comprehensive API reference documents all endpoints available in SuperAgent's REST API. The API is OpenAI-compatible, meaning existing OpenAI client libraries can be used with minimal modifications. SuperAgent extends the standard API with additional endpoints for ensemble operations, AI debates, provider management, and knowledge graph integration.

---

## Table of Contents

1. [Base URL and Authentication](#base-url-and-authentication)
2. [Authentication Endpoints](#authentication-endpoints)
3. [Chat Completions](#chat-completions)
4. [Text Completions](#text-completions)
5. [Embeddings](#embeddings)
6. [Models](#models)
7. [Providers](#providers)
8. [AI Debates](#ai-debates)
9. [Cognee Knowledge Graph](#cognee-knowledge-graph)
10. [Health and Monitoring](#health-and-monitoring)
11. [Rate Limiting](#rate-limiting)
12. [Error Handling](#error-handling)

---

## Base URL and Authentication

### Base URL

```
Production:  https://api.superagent.ai/v1
Development: http://localhost:8080/v1
```

### Authentication Methods

SuperAgent supports multiple authentication methods:

#### JWT Bearer Token

```bash
curl -X GET http://localhost:8080/v1/models \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

#### API Key

```bash
curl -X GET http://localhost:8080/v1/models \
  -H "X-API-Key: your-api-key-here"
```

#### Basic Authentication

```bash
curl -X GET http://localhost:8080/v1/models \
  -u "username:password"
```

### Request Headers

| Header | Required | Description |
|--------|----------|-------------|
| `Authorization` | Yes | Bearer token or API key |
| `Content-Type` | Yes | `application/json` for JSON requests |
| `Accept` | No | `application/json` or `text/event-stream` for streaming |
| `X-Request-ID` | No | Custom request ID for tracing |

---

## Authentication Endpoints

### Register User

Creates a new user account.

```
POST /v1/auth/register
```

**Request Body**

```json
{
  "username": "newuser",
  "email": "user@example.com",
  "password": "SecurePassword123!",
  "full_name": "John Doe"
}
```

**Response**

```json
{
  "success": true,
  "message": "User registered successfully",
  "user": {
    "id": "user_abc123",
    "username": "newuser",
    "email": "user@example.com",
    "full_name": "John Doe",
    "created_at": "2024-01-15T10:30:00Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Login

Authenticates user and returns JWT token.

```
POST /v1/auth/login
```

**Request Body**

```json
{
  "username": "myuser",
  "password": "SecurePassword123!"
}
```

**Response**

```json
{
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2024-01-16T10:30:00Z",
  "user": {
    "id": "user_abc123",
    "username": "myuser",
    "email": "user@example.com"
  }
}
```

### Refresh Token

Refreshes an existing JWT token.

```
POST /v1/auth/refresh
```

**Request Headers**

```
Authorization: Bearer <current_token>
```

**Response**

```json
{
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2024-01-17T10:30:00Z"
}
```

### Logout

Invalidates the current token.

```
POST /v1/auth/logout
```

**Response**

```json
{
  "success": true,
  "message": "Logged out successfully"
}
```

### Get Current User

Returns the authenticated user's information.

```
GET /v1/auth/me
```

**Response**

```json
{
  "id": "user_abc123",
  "username": "myuser",
  "email": "user@example.com",
  "full_name": "John Doe",
  "role": "user",
  "created_at": "2024-01-01T00:00:00Z",
  "last_login": "2024-01-15T10:30:00Z"
}
```

---

## Chat Completions

### Create Chat Completion

Generates a chat completion response.

```
POST /v1/chat/completions
```

**Request Body**

```json
{
  "model": "superagent-ensemble",
  "messages": [
    {
      "role": "system",
      "content": "You are a helpful AI assistant."
    },
    {
      "role": "user",
      "content": "What are the benefits of microservices architecture?"
    }
  ],
  "temperature": 0.7,
  "max_tokens": 1000,
  "top_p": 0.9,
  "frequency_penalty": 0.0,
  "presence_penalty": 0.0,
  "stop": ["\n\n"],
  "stream": false,
  "user": "user-identifier"
}
```

**Parameters**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `model` | string | Yes | - | Model ID or ensemble name |
| `messages` | array | Yes | - | Array of message objects |
| `temperature` | number | No | 0.7 | Sampling temperature (0-2) |
| `max_tokens` | integer | No | 4096 | Maximum tokens to generate |
| `top_p` | number | No | 1.0 | Nucleus sampling parameter |
| `frequency_penalty` | number | No | 0.0 | Frequency penalty (-2 to 2) |
| `presence_penalty` | number | No | 0.0 | Presence penalty (-2 to 2) |
| `stop` | array | No | null | Stop sequences |
| `stream` | boolean | No | false | Enable streaming |
| `user` | string | No | null | User identifier for tracking |

**Message Object**

```json
{
  "role": "user | assistant | system",
  "content": "Message content",
  "name": "optional_name"
}
```

**Response**

```json
{
  "id": "chatcmpl-superagent-abc123",
  "object": "chat.completion",
  "created": 1704067200,
  "model": "superagent-ensemble",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Microservices architecture offers several key benefits..."
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 45,
    "completion_tokens": 250,
    "total_tokens": 295
  },
  "ensemble": {
    "providers_used": ["deepseek", "claude"],
    "voting_strategy": "confidence_weighted",
    "confidence_score": 0.92,
    "provider_responses": [
      {
        "provider": "deepseek",
        "confidence": 0.90,
        "latency_ms": 1200
      },
      {
        "provider": "claude",
        "confidence": 0.94,
        "latency_ms": 1500
      }
    ]
  }
}
```

### Streaming Chat Completion

Enable streaming for real-time response generation.

```
POST /v1/chat/completions
```

**Request Body**

```json
{
  "model": "deepseek-chat",
  "messages": [{"role": "user", "content": "Tell me a story"}],
  "stream": true
}
```

**Response (Server-Sent Events)**

```
data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1704067200,"model":"deepseek-chat","choices":[{"index":0,"delta":{"role":"assistant"},"finish_reason":null}]}

data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1704067200,"model":"deepseek-chat","choices":[{"index":0,"delta":{"content":"Once"},"finish_reason":null}]}

data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1704067200,"model":"deepseek-chat","choices":[{"index":0,"delta":{"content":" upon"},"finish_reason":null}]}

data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1704067200,"model":"deepseek-chat","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}

data: [DONE]
```

### Ensemble Chat Completion

Use multiple providers with ensemble voting.

```
POST /v1/chat/completions
```

**Request Body**

```json
{
  "model": "superagent-ensemble",
  "messages": [
    {"role": "user", "content": "Explain quantum computing"}
  ],
  "ensemble": {
    "providers": ["claude", "deepseek", "gemini"],
    "strategy": "confidence_weighted",
    "min_responses": 2,
    "timeout": "30s"
  }
}
```

---

## Text Completions

### Create Completion

Generates a text completion for a given prompt.

```
POST /v1/completions
```

**Request Body**

```json
{
  "model": "deepseek-chat",
  "prompt": "The future of artificial intelligence is",
  "max_tokens": 100,
  "temperature": 0.7,
  "top_p": 0.9,
  "n": 1,
  "stop": ["\n"],
  "echo": false,
  "logprobs": null,
  "suffix": null
}
```

**Parameters**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `model` | string | Yes | - | Model ID |
| `prompt` | string/array | Yes | - | Text prompt(s) |
| `max_tokens` | integer | No | 256 | Maximum tokens |
| `temperature` | number | No | 0.7 | Sampling temperature |
| `top_p` | number | No | 1.0 | Nucleus sampling |
| `n` | integer | No | 1 | Number of completions |
| `stop` | array | No | null | Stop sequences |
| `echo` | boolean | No | false | Echo prompt in response |
| `logprobs` | integer | No | null | Log probabilities |

**Response**

```json
{
  "id": "cmpl-abc123",
  "object": "text_completion",
  "created": 1704067200,
  "model": "deepseek-chat",
  "choices": [
    {
      "text": " likely to revolutionize every industry...",
      "index": 0,
      "logprobs": null,
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 10,
    "completion_tokens": 50,
    "total_tokens": 60
  }
}
```

---

## Embeddings

### Create Embedding

Generates embedding vectors for input text.

```
POST /v1/embeddings
```

**Request Body**

```json
{
  "model": "text-embedding-3-small",
  "input": "The quick brown fox jumps over the lazy dog",
  "encoding_format": "float"
}
```

**Parameters**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `model` | string | Yes | - | Embedding model |
| `input` | string/array | Yes | - | Text(s) to embed |
| `encoding_format` | string | No | "float" | "float" or "base64" |
| `dimensions` | integer | No | null | Output dimensions |

**Response**

```json
{
  "object": "list",
  "data": [
    {
      "object": "embedding",
      "index": 0,
      "embedding": [0.0023, -0.0089, 0.0156, ...]
    }
  ],
  "model": "text-embedding-3-small",
  "usage": {
    "prompt_tokens": 12,
    "total_tokens": 12
  }
}
```

---

## Models

### List Models

Returns all available models.

```
GET /v1/models
```

**Response**

```json
{
  "object": "list",
  "data": [
    {
      "id": "superagent-ensemble",
      "object": "model",
      "created": 1704067200,
      "owned_by": "superagent",
      "permission": [],
      "capabilities": {
        "chat": true,
        "completion": true,
        "embedding": false,
        "ensemble": true
      }
    },
    {
      "id": "claude-3-sonnet-20240229",
      "object": "model",
      "created": 1704067200,
      "owned_by": "anthropic",
      "provider": "claude",
      "capabilities": {
        "chat": true,
        "completion": true,
        "embedding": false
      }
    },
    {
      "id": "deepseek-chat",
      "object": "model",
      "created": 1704067200,
      "owned_by": "deepseek",
      "provider": "deepseek",
      "capabilities": {
        "chat": true,
        "completion": true,
        "embedding": false
      }
    }
  ]
}
```

### Get Model

Retrieves information about a specific model.

```
GET /v1/models/{model_id}
```

**Response**

```json
{
  "id": "claude-3-sonnet-20240229",
  "object": "model",
  "created": 1704067200,
  "owned_by": "anthropic",
  "provider": "claude",
  "capabilities": {
    "chat": true,
    "completion": true,
    "embedding": false,
    "function_calling": true,
    "vision": false
  },
  "limits": {
    "max_tokens": 4096,
    "context_window": 200000
  }
}
```

---

## Providers

### List Providers

Returns all configured LLM providers.

```
GET /v1/providers
```

**Response**

```json
{
  "object": "list",
  "data": [
    {
      "id": "claude",
      "name": "Claude (Anthropic)",
      "status": "active",
      "health": "healthy",
      "models": ["claude-3-opus", "claude-3-sonnet", "claude-3-haiku"],
      "rate_limit": {
        "requests_per_minute": 60,
        "remaining": 45
      }
    },
    {
      "id": "deepseek",
      "name": "DeepSeek",
      "status": "active",
      "health": "healthy",
      "models": ["deepseek-chat", "deepseek-coder"],
      "rate_limit": {
        "requests_per_minute": 100,
        "remaining": 89
      }
    }
  ]
}
```

### Get Provider Status

Returns detailed status for a specific provider.

```
GET /v1/providers/{provider_id}
```

**Response**

```json
{
  "id": "deepseek",
  "name": "DeepSeek",
  "status": "active",
  "health": {
    "status": "healthy",
    "latency_ms": 245,
    "last_check": "2024-01-15T10:30:00Z",
    "uptime_percentage": 99.9
  },
  "models": [
    {
      "id": "deepseek-chat",
      "available": true,
      "max_tokens": 4096
    }
  ],
  "rate_limit": {
    "requests_per_minute": 100,
    "tokens_per_minute": 200000,
    "remaining_requests": 89,
    "remaining_tokens": 195000,
    "reset_at": "2024-01-15T10:31:00Z"
  },
  "metrics": {
    "requests_today": 1523,
    "tokens_today": 2500000,
    "errors_today": 2,
    "average_latency_ms": 280
  }
}
```

### Provider Health Check

Performs a health check for a specific provider.

```
GET /v1/providers/{provider_id}/health
```

**Response**

```json
{
  "provider": "deepseek",
  "status": "healthy",
  "latency_ms": 189,
  "timestamp": "2024-01-15T10:30:00Z",
  "checks": {
    "api_reachable": true,
    "authentication": true,
    "model_available": true,
    "rate_limit_ok": true
  }
}
```

---

## AI Debates

### Create Debate

Initiates a new AI debate session.

```
POST /v1/debates
```

**Request Body**

```json
{
  "topic": "Should AI systems be required to explain their decisions?",
  "context": "Consider technical feasibility, user trust, and regulatory requirements",
  "participants": [
    {
      "provider": "claude",
      "role": "proposer",
      "model": "claude-3-sonnet-20240229"
    },
    {
      "provider": "deepseek",
      "role": "critic",
      "model": "deepseek-chat"
    },
    {
      "provider": "gemini",
      "role": "synthesizer",
      "model": "gemini-pro"
    }
  ],
  "settings": {
    "max_rounds": 3,
    "consensus_threshold": 0.7,
    "consensus_strategy": "synthesized",
    "round_timeout": "60s"
  }
}
```

**Response**

```json
{
  "id": "debate_abc123",
  "object": "debate",
  "topic": "Should AI systems be required to explain their decisions?",
  "status": "in_progress",
  "current_round": 1,
  "participants": [
    {"id": "p1", "provider": "claude", "role": "proposer"},
    {"id": "p2", "provider": "deepseek", "role": "critic"},
    {"id": "p3", "provider": "gemini", "role": "synthesizer"}
  ],
  "created_at": "2024-01-15T10:30:00Z",
  "estimated_completion": "2024-01-15T10:35:00Z"
}
```

### Get Debate Status

Retrieves the current status of a debate.

```
GET /v1/debates/{debate_id}
```

**Response**

```json
{
  "id": "debate_abc123",
  "object": "debate",
  "topic": "Should AI systems be required to explain their decisions?",
  "status": "completed",
  "current_round": 3,
  "total_rounds": 3,
  "consensus_reached": true,
  "consensus_score": 0.87,
  "final_response": "Based on the debate, AI systems should provide...",
  "participants": [
    {"id": "p1", "provider": "claude", "role": "proposer", "contributions": 3},
    {"id": "p2", "provider": "deepseek", "role": "critic", "contributions": 2},
    {"id": "p3", "provider": "gemini", "role": "synthesizer", "contributions": 1}
  ],
  "created_at": "2024-01-15T10:30:00Z",
  "completed_at": "2024-01-15T10:34:23Z"
}
```

### Get Debate Transcript

Retrieves the full transcript of a debate.

```
GET /v1/debates/{debate_id}/transcript
```

**Response**

```json
{
  "debate_id": "debate_abc123",
  "topic": "Should AI systems be required to explain their decisions?",
  "rounds": [
    {
      "round_number": 1,
      "type": "proposal",
      "contributions": [
        {
          "participant_id": "p1",
          "provider": "claude",
          "role": "proposer",
          "content": "I propose that AI systems should...",
          "timestamp": "2024-01-15T10:30:15Z",
          "confidence": 0.88
        }
      ]
    },
    {
      "round_number": 2,
      "type": "critique",
      "contributions": [
        {
          "participant_id": "p2",
          "provider": "deepseek",
          "role": "critic",
          "content": "While the proposal has merit...",
          "timestamp": "2024-01-15T10:31:30Z",
          "confidence": 0.82
        }
      ]
    }
  ],
  "consensus": {
    "reached": true,
    "score": 0.87,
    "strategy": "synthesized",
    "final_synthesis": "Based on all perspectives..."
  }
}
```

### Stream Debate Progress

Streams real-time debate progress.

```
GET /v1/debates/{debate_id}/stream
```

**Headers**

```
Accept: text/event-stream
```

**Response (Server-Sent Events)**

```
event: round_start
data: {"round": 1, "type": "proposal"}

event: contribution
data: {"participant": "claude", "role": "proposer", "content": "I propose..."}

event: round_end
data: {"round": 1, "contributions": 1}

event: round_start
data: {"round": 2, "type": "critique"}

event: contribution
data: {"participant": "deepseek", "role": "critic", "content": "While..."}

event: consensus
data: {"reached": true, "score": 0.87, "final": "Based on..."}

event: debate_complete
data: {"debate_id": "debate_abc123", "status": "completed"}
```

### List Debates

Lists all debates for the authenticated user.

```
GET /v1/debates
```

**Query Parameters**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `status` | string | all | Filter by status |
| `limit` | integer | 20 | Max results |
| `offset` | integer | 0 | Pagination offset |

**Response**

```json
{
  "object": "list",
  "data": [
    {
      "id": "debate_abc123",
      "topic": "Should AI...",
      "status": "completed",
      "consensus_reached": true,
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 45,
  "limit": 20,
  "offset": 0
}
```

---

## Cognee Knowledge Graph

### Add Knowledge

Adds content to the knowledge graph.

```
POST /v1/cognee/add
```

**Request Body**

```json
{
  "content": "SuperAgent is an AI orchestration platform...",
  "metadata": {
    "source": "documentation",
    "category": "product_info",
    "tags": ["ai", "orchestration", "llm"]
  },
  "dataset": "superagent_docs"
}
```

**Response**

```json
{
  "success": true,
  "document_id": "doc_xyz789",
  "chunks_created": 3,
  "entities_extracted": 5,
  "relationships_created": 8
}
```

### Search Knowledge

Searches the knowledge graph.

```
POST /v1/cognee/search
```

**Request Body**

```json
{
  "query": "How does ensemble voting work?",
  "search_type": "vector",
  "limit": 10,
  "filters": {
    "dataset": "superagent_docs",
    "category": "technical"
  }
}
```

**Response**

```json
{
  "results": [
    {
      "id": "chunk_abc123",
      "content": "Ensemble voting aggregates responses...",
      "score": 0.92,
      "metadata": {
        "source": "documentation",
        "chunk_index": 5
      }
    }
  ],
  "total": 15,
  "search_type": "vector"
}
```

### Cognify Content

Processes content through Cognee's cognification pipeline.

```
POST /v1/cognee/cognify
```

**Request Body**

```json
{
  "content": "Large language models are...",
  "options": {
    "extract_entities": true,
    "build_relationships": true,
    "generate_embeddings": true
  }
}
```

**Response**

```json
{
  "success": true,
  "entities": [
    {"name": "Large language models", "type": "CONCEPT"},
    {"name": "AI", "type": "TECHNOLOGY"}
  ],
  "relationships": [
    {"source": "Large language models", "target": "AI", "type": "PART_OF"}
  ],
  "processing_time_ms": 1250
}
```

---

## Health and Monitoring

### Health Check

Basic health check endpoint.

```
GET /health
```

**Response**

```json
{
  "status": "ok",
  "timestamp": 1704067200,
  "version": "1.0.0"
}
```

### Detailed Health

Comprehensive health check with component status.

```
GET /v1/health
```

**Response**

```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0.0",
  "uptime_seconds": 86400,
  "components": {
    "database": {
      "status": "healthy",
      "latency_ms": 5
    },
    "redis": {
      "status": "healthy",
      "latency_ms": 2
    },
    "providers": {
      "status": "healthy",
      "active": 5,
      "total": 7
    },
    "cognee": {
      "status": "healthy",
      "latency_ms": 15
    }
  }
}
```

### Metrics

Prometheus-compatible metrics endpoint.

```
GET /metrics
```

**Response**

```
# HELP superagent_requests_total Total number of requests
# TYPE superagent_requests_total counter
superagent_requests_total{endpoint="/v1/chat/completions",status="200"} 15234

# HELP superagent_request_duration_seconds Request duration in seconds
# TYPE superagent_request_duration_seconds histogram
superagent_request_duration_seconds_bucket{endpoint="/v1/chat/completions",le="0.5"} 12000
superagent_request_duration_seconds_bucket{endpoint="/v1/chat/completions",le="1"} 14500
superagent_request_duration_seconds_bucket{endpoint="/v1/chat/completions",le="5"} 15200
```

---

## Rate Limiting

### Rate Limit Headers

All responses include rate limit headers:

```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1704067260
X-RateLimit-Reset-After: 30
```

### Rate Limit Response

When rate limited, the API returns:

```json
{
  "error": {
    "type": "rate_limit_error",
    "message": "Rate limit exceeded. Please retry after 30 seconds.",
    "code": "rate_limit_exceeded",
    "retry_after": 30
  }
}
```

---

## Error Handling

### Error Response Format

All errors follow a consistent format:

```json
{
  "error": {
    "type": "invalid_request_error",
    "message": "The 'model' parameter is required",
    "code": "missing_parameter",
    "param": "model",
    "request_id": "req_abc123"
  }
}
```

### Error Codes

| HTTP Status | Error Type | Description |
|-------------|------------|-------------|
| 400 | `invalid_request_error` | Malformed request |
| 401 | `authentication_error` | Invalid or missing auth |
| 403 | `permission_error` | Insufficient permissions |
| 404 | `not_found_error` | Resource not found |
| 429 | `rate_limit_error` | Rate limit exceeded |
| 500 | `server_error` | Internal server error |
| 503 | `service_unavailable` | Provider unavailable |

### Common Error Codes

| Code | Description |
|------|-------------|
| `missing_parameter` | Required parameter missing |
| `invalid_parameter` | Parameter value invalid |
| `invalid_api_key` | API key invalid or expired |
| `model_not_found` | Requested model not available |
| `context_length_exceeded` | Input too long |
| `provider_unavailable` | LLM provider down |
| `consensus_failed` | Debate failed to reach consensus |

---

## Summary

This API reference covers all SuperAgent endpoints. Key points:

1. **OpenAI-Compatible**: Standard chat/completion endpoints work with existing clients
2. **Extended Features**: Ensemble, debate, and knowledge endpoints add unique capabilities
3. **Comprehensive Auth**: Multiple authentication methods supported
4. **Full Monitoring**: Health, metrics, and provider status endpoints available

For implementation examples, see the [Getting Started Guide](01-getting-started.md). For deployment, see the [Deployment Guide](05-deployment-guide.md).
