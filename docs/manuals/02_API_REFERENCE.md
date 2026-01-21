# Chapter 2: API Reference

Complete API reference for HelixAgent's OpenAI-compatible endpoints.

## Base URL

```
http://localhost:7061
```

## Authentication

HelixAgent supports multiple authentication methods:

### API Key Authentication

```bash
curl -H "Authorization: Bearer YOUR_API_KEY" \
  http://localhost:7061/v1/chat/completions
```

### JWT Token Authentication

```bash
# Login to get token
TOKEN=$(curl -X POST http://localhost:7061/auth/login \
  -d '{"email":"user@example.com","password":"secret"}' | jq -r '.token')

# Use token in requests
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:7061/v1/chat/completions
```

## Endpoints

### Health & Status

#### GET /health

Check service health.

```bash
curl http://localhost:7061/health
```

Response:
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "providers": {
    "claude": "healthy",
    "deepseek": "healthy",
    "gemini": "healthy"
  }
}
```

#### GET /v1/models

List available models.

```bash
curl http://localhost:7061/v1/models
```

Response:
```json
{
  "data": [
    {
      "id": "helixagent-debate",
      "object": "model",
      "owned_by": "helixagent"
    }
  ]
}
```

### Chat Completions

#### POST /v1/chat/completions

Create a chat completion.

**Request:**

```json
{
  "model": "helixagent-debate",
  "messages": [
    {"role": "system", "content": "You are a helpful assistant."},
    {"role": "user", "content": "What is machine learning?"}
  ],
  "temperature": 0.7,
  "max_tokens": 1000,
  "stream": false
}
```

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| model | string | Yes | Model ID (e.g., "helixagent-debate") |
| messages | array | Yes | Array of message objects |
| temperature | float | No | Sampling temperature (0-2). Default: 0.7 |
| max_tokens | int | No | Maximum tokens to generate |
| stream | bool | No | Enable streaming. Default: false |
| top_p | float | No | Nucleus sampling parameter |
| stop | array | No | Stop sequences |

**Response:**

```json
{
  "id": "chatcmpl-abc123",
  "object": "chat.completion",
  "created": 1699000000,
  "model": "helixagent-debate",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Machine learning is..."
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 20,
    "completion_tokens": 150,
    "total_tokens": 170
  }
}
```

### Streaming

Enable streaming for real-time responses:

```bash
curl -X POST http://localhost:7061/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "helixagent-debate",
    "messages": [{"role": "user", "content": "Hello"}],
    "stream": true
  }'
```

Streamed response format (Server-Sent Events):

```
data: {"id":"chatcmpl-abc123","choices":[{"delta":{"content":"Hello"}}]}

data: {"id":"chatcmpl-abc123","choices":[{"delta":{"content":" there"}}]}

data: [DONE]
```

### Embeddings

#### POST /v1/embeddings

Generate vector embeddings.

```bash
curl -X POST http://localhost:7061/v1/embeddings \
  -H "Content-Type: application/json" \
  -d '{
    "model": "helixagent-embedding",
    "input": "Hello world"
  }'
```

Response:
```json
{
  "object": "list",
  "data": [
    {
      "object": "embedding",
      "index": 0,
      "embedding": [0.123, -0.456, ...]
    }
  ],
  "model": "helixagent-embedding",
  "usage": {
    "prompt_tokens": 2,
    "total_tokens": 2
  }
}
```

### AI Debate

#### POST /v1/debates

Create a new AI debate session.

```bash
curl -X POST http://localhost:7061/v1/debates \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "Should AI be regulated?",
    "rounds": 3,
    "style": "theater"
  }'
```

Response:
```json
{
  "id": "debate-abc123",
  "status": "created",
  "topic": "Should AI be regulated?",
  "rounds": 3,
  "created_at": "2024-01-15T10:00:00Z"
}
```

#### GET /v1/debates/:id

Get debate status and results.

```bash
curl http://localhost:7061/v1/debates/debate-abc123
```

#### GET /v1/debates/:id/status

Poll debate status.

```bash
curl http://localhost:7061/v1/debates/debate-abc123/status
```

### MCP (Model Context Protocol)

#### GET /v1/mcp/capabilities

Get MCP server capabilities.

```bash
curl http://localhost:7061/v1/mcp/capabilities
```

Response:
```json
{
  "version": "1.0.0",
  "capabilities": {
    "tools": {"listChanged": true},
    "prompts": {"listChanged": true},
    "resources": {"listChanged": true}
  },
  "providers": ["claude", "deepseek", "gemini"],
  "mcp_servers": ["filesystem", "git", "github", "memory"]
}
```

#### GET /v1/mcp/tools

List available MCP tools.

```bash
curl http://localhost:7061/v1/mcp/tools
```

#### POST /v1/mcp/tools/call

Execute an MCP tool.

```bash
curl -X POST http://localhost:7061/v1/mcp/tools/call \
  -H "Content-Type: application/json" \
  -d '{
    "name": "filesystem_read",
    "arguments": {"path": "/path/to/file"}
  }'
```

### MCP Tool Search API (NEW)

The MCP Tool Search API enables intelligent discovery of tools and MCP adapters.

#### GET /v1/mcp/tools/search

Search for tools by query string. Supports fuzzy matching and category filtering.

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| q (or query) | string | Yes | Search query |
| categories | string | No | Comma-separated category filter |
| include_params | bool | No | Include parameter details (default: false) |
| fuzzy | bool | No | Enable fuzzy matching (default: false) |
| max_results | int | No | Limit results (default: unlimited) |

**Example - Search for file-related tools:**

```bash
curl "http://localhost:7061/v1/mcp/tools/search?q=file"
```

Response:
```json
{
  "query": "file",
  "count": 5,
  "results": [
    {
      "name": "Read",
      "description": "Read file contents",
      "category": "filesystem",
      "score": 1.0,
      "match_type": "exact",
      "parameters": {"file_path": "string"},
      "required": ["file_path"],
      "aliases": ["read_file", "cat"]
    },
    {
      "name": "Write",
      "description": "Write content to a file",
      "category": "filesystem",
      "score": 0.9,
      "match_type": "description"
    }
  ]
}
```

**Example - Search with fuzzy matching:**

```bash
curl "http://localhost:7061/v1/mcp/tools/search?q=fiel&fuzzy=true"
```

**Example - Filter by category:**

```bash
curl "http://localhost:7061/v1/mcp/tools/search?q=read&categories=filesystem,git"
```

#### POST /v1/mcp/tools/search

Search tools with JSON body (supports advanced options).

```bash
curl -X POST http://localhost:7061/v1/mcp/tools/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "file operations",
    "categories": ["filesystem"],
    "include_params": true,
    "fuzzy_match": true,
    "max_results": 10
  }'
```

#### GET /v1/mcp/adapters/search

Search for MCP adapters (servers) by name or functionality.

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| q (or query) | string | No | Search query |
| categories | string | No | Comma-separated category filter |
| auth_types | string | No | Filter by auth type (api_key, oauth, none) |
| official | bool | No | Filter official adapters only |
| supported | bool | No | Filter supported adapters only |
| max_results | int | No | Limit results |

**Example - Search for GitHub adapter:**

```bash
curl "http://localhost:7061/v1/mcp/adapters/search?q=github"
```

Response:
```json
{
  "query": "github",
  "count": 1,
  "results": [
    {
      "name": "github",
      "description": "GitHub repository management, commits, branches, PRs",
      "category": "version_control",
      "auth_type": "api_key",
      "official": true,
      "supported": true
    }
  ]
}
```

#### GET /v1/mcp/tools/suggestions

Get tool suggestions based on a natural language prompt.

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| prompt | string | Yes | Natural language description of task |

**Example:**

```bash
curl "http://localhost:7061/v1/mcp/tools/suggestions?prompt=list%20files%20in%20directory"
```

Response:
```json
{
  "prompt": "list files in directory",
  "suggestions": [
    {
      "tool": "Glob",
      "confidence": 0.95,
      "reason": "Glob is ideal for listing files matching patterns"
    },
    {
      "tool": "Bash",
      "confidence": 0.7,
      "reason": "Can execute ls command"
    }
  ]
}
```

#### GET /v1/mcp/categories

List all available tool categories.

```bash
curl http://localhost:7061/v1/mcp/categories
```

#### GET /v1/mcp/stats

Get MCP statistics including tool counts and usage.

```bash
curl http://localhost:7061/v1/mcp/stats
```

### Vision

#### POST /v1/vision/analyze

Analyze an image.

```bash
curl -X POST http://localhost:7061/v1/vision/analyze \
  -H "Content-Type: application/json" \
  -d '{
    "image_url": "https://example.com/image.jpg",
    "prompt": "Describe this image"
  }'
```

### Cognee (Knowledge Graph)

#### POST /v1/cognee/add

Add knowledge to the graph.

```bash
curl -X POST http://localhost:7061/v1/cognee/add \
  -H "Content-Type: application/json" \
  -d '{
    "data": "Your knowledge content here"
  }'
```

#### POST /v1/cognee/search

Search the knowledge graph.

```bash
curl -X POST http://localhost:7061/v1/cognee/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "What do you know about X?"
  }'
```

## Error Handling

### Error Response Format

```json
{
  "error": {
    "message": "Invalid request",
    "type": "invalid_request_error",
    "code": "invalid_api_key"
  }
}
```

### Common Error Codes

| Status | Code | Description |
|--------|------|-------------|
| 400 | invalid_request | Malformed request |
| 401 | invalid_api_key | Authentication failed |
| 403 | access_denied | Insufficient permissions |
| 404 | not_found | Resource not found |
| 429 | rate_limit_exceeded | Too many requests |
| 500 | internal_error | Server error |

## Rate Limiting

Default rate limits:
- 100 requests per minute per API key
- 10,000 tokens per minute per API key

Rate limit headers:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1699000060
```

## Pagination

For list endpoints:

```bash
curl "http://localhost:7061/v1/debates?limit=10&offset=0"
```

Response includes pagination info:
```json
{
  "data": [...],
  "has_more": true,
  "total": 100
}
```

## OpenAPI Specification

Full OpenAPI spec available at:
```
http://localhost:7061/swagger/doc.json
```

Swagger UI:
```
http://localhost:7061/swagger/index.html
```
