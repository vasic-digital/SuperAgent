# OpenAI API Documentation

## Overview

OpenAI provides industry-leading language models including GPT-4o, GPT-4, and the o-series reasoning models. This documentation covers the complete OpenAI API with cutting-edge protocol support.

**Base URL:** `https://api.openai.com/v1`

**Documentation:** https://platform.openai.com/docs/api-reference

---

## Feature Support Matrix

| Feature | Status | Notes |
|---------|--------|-------|
| HTTP/3 (QUIC) | ✅ Supported | HTTP/3 enabled for lower latency |
| HTTP/2 | ✅ Supported | Multiplexing, server push |
| Brotli Compression | ✅ Supported | `br` encoding accepted |
| Gzip Compression | ✅ Supported | Standard gzip encoding |
| Streaming (SSE) | ✅ Supported | Server-sent events for real-time |
| WebSocket | ❌ Not Supported | Use SSE instead |
| gRPC | ❌ Not Supported | REST API only |
| Toon/Binary JSON | ❌ Not Supported | JSON only |

---

## Authentication

### Bearer Token
```
Authorization: Bearer <api_key>
```

### Organization Header (Optional)
```
OpenAI-Organization: <org_id>
```

### Project Header (Optional)
```
OpenAI-Project: <project_id>
```

---

## Complete Endpoint Reference

### Chat Completions
Create chat completions with streaming support.

**Endpoint:** `POST /chat/completions`

**Features:**
- Multi-turn conversations
- Streaming (SSE)
- Tool/function calling
- JSON mode
- Vision (multimodal)
- Reasoning (o-series models)

### Assistants
Build AI assistants with persistent threads and tools.

**Endpoints:**
- `POST /assistants` - Create assistant
- `GET /assistants` - List assistants
- `GET /assistants/{assistant_id}` - Retrieve assistant
- `POST /assistants/{assistant_id}` - Modify assistant
- `DELETE /assistants/{assistant_id}` - Delete assistant

### Threads
Conversation threads for Assistants API.

**Endpoints:**
- `POST /threads` - Create thread
- `GET /threads/{thread_id}` - Retrieve thread
- `POST /threads/{thread_id}` - Modify thread
- `DELETE /threads/{thread_id}` - Delete thread

### Messages
Messages within threads.

**Endpoints:**
- `POST /threads/{thread_id}/messages` - Create message
- `GET /threads/{thread_id}/messages` - List messages
- `GET /threads/{thread_id}/messages/{message_id}` - Retrieve message
- `POST /threads/{thread_id}/messages/{message_id}` - Modify message

### Runs
Execute runs on threads.

**Endpoints:**
- `POST /threads/{thread_id}/runs` - Create run
- `GET /threads/{thread_id}/runs` - List runs
- `GET /threads/{thread_id}/runs/{run_id}` - Retrieve run
- `POST /threads/{thread_id}/runs/{run_id}` - Modify run
- `POST /threads/{thread_id}/runs/{run_id}/submit_tool_outputs` - Submit tool outputs
- `POST /threads/{thread_id}/runs/{run_id}/cancel` - Cancel run
- `POST /threads/{thread_id}/runs/{run_id}/steps` - List run steps

### Vector Stores
Store and search documents for retrieval.

**Endpoints:**
- `POST /vector_stores` - Create vector store
- `GET /vector_stores` - List vector stores
- `GET /vector_stores/{vector_store_id}` - Retrieve vector store
- `POST /vector_stores/{vector_store_id}` - Modify vector store
- `DELETE /vector_stores/{vector_store_id}` - Delete vector store

### Embeddings
Generate text embeddings.

**Endpoint:** `POST /embeddings`

### Fine-tuning
Fine-tune models on custom data.

**Endpoints:**
- `POST /fine_tuning/jobs` - Create fine-tuning job
- `GET /fine_tuning/jobs` - List fine-tuning jobs
- `GET /fine_tuning/jobs/{job_id}` - Retrieve fine-tuning job
- `POST /fine_tuning/jobs/{job_id}/cancel` - Cancel fine-tuning job
- `GET /fine_tuning/jobs/{job_id}/events` - List fine-tuning events
- `GET /fine_tuning/jobs/{job_id}/checkpoints` - List fine-tuning checkpoints

### Batch
Process batches of API requests.

**Endpoints:**
- `POST /batches` - Create batch
- `GET /batches` - List batches
- `GET /batches/{batch_id}` - Retrieve batch
- `POST /batches/{batch_id}/cancel` - Cancel batch

### Files
Upload and manage files.

**Endpoints:**
- `POST /files` - Upload file
- `GET /files` - List files
- `GET /files/{file_id}` - Retrieve file
- `DELETE /files/{file_id}` - Delete file
- `GET /files/{file_id}/content` - Retrieve file content

### Uploads
Large file uploads with resumable transfers.

**Endpoints:**
- `POST /uploads` - Create upload
- `POST /uploads/{upload_id}/parts` - Add upload part
- `POST /uploads/{upload_id}/complete` - Complete upload
- `POST /uploads/{upload_id}/cancel` - Cancel upload

### Images
Generate and edit images with DALL-E.

**Endpoints:**
- `POST /images/generations` - Create image
- `POST /images/edits` - Edit image
- `POST /images/variations` - Create image variation

### Audio
Speech-to-text and text-to-speech.

**Endpoints:**
- `POST /audio/speech` - Create speech (TTS)
- `POST /audio/transcriptions` - Create transcription (STT)
- `POST /audio/translations` - Create translation

### Models
List and retrieve models.

**Endpoints:**
- `GET /models` - List models
- `GET /models/{model}` - Retrieve model
- `DELETE /models/{model}` - Delete fine-tuned model

### Moderations
Check content against usage policies.

**Endpoint:** `POST /moderations`

---

## Advanced Protocol Support

### HTTP/3 (QUIC)
OpenAI supports HTTP/3 for reduced latency and improved performance.

**Client Requirements:**
- QUIC protocol support
- TLS 1.3
- Connection migration support

**Implementation:**
```go
// Using quic-go
import "github.com/quic-go/quic-go/http3"

client := &http3.Client{
    // HTTP/3 configuration
}
```

### Brotli Compression
Enable Brotli compression for smaller payload sizes.

**Request Header:**
```
Accept-Encoding: br, gzip
```

**Response:**
- Content-Encoding: `br` (Brotli compressed)
- Content-Encoding: `gzip` (Gzip compressed)

**Implementation:**
```go
import "github.com/andybalholm/brotli"

// Automatic decompression
reader := brotli.NewReader(resp.Body)
```

### Streaming (SSE)
Server-sent events for real-time responses.

**Request:**
```json
{
  "model": "gpt-4o",
  "messages": [{"role": "user", "content": "Hello"}],
  "stream": true
}
```

**Response:**
```
data: {"id":"...","object":"chat.completion.chunk","choices":[{"delta":{"content":"Hello"}}]}

data: {"id":"...","object":"chat.completion.chunk","choices":[{"delta":{"content":"!"}}]}

data: [DONE]
```

---

## SDK Integration

See [sdk-integration.md](./sdk-integration.md) for HelixAgent SDK integration details.

---

## Models Reference

See [models.md](./models.md) for complete model specifications.

---

## Rate Limits

| Tier | RPM | TPM | Batch Queue Limit |
|------|-----|-----|-------------------|
| Free | 3 | 40,000 | 3 requests |
| Tier 1 | 500 | 200,000 | 5 requests |
| Tier 2 | 5,000 | 2,000,000 | 100 requests |
| Tier 3 | 10,000 | 10,000,000 | 1,000 requests |
| Tier 4 | 10,000 | 40,000,000 | 2,000 requests |
| Tier 5 | 10,000 | 80,000,000 | 2,000 requests |

---

## Error Codes

| Status | Code | Description |
|--------|------|-------------|
| 400 | `invalid_request_error` | Invalid request parameters |
| 401 | `authentication_error` | Invalid API key |
| 403 | `permission_error` | Insufficient permissions |
| 404 | `not_found_error` | Resource not found |
| 409 | `conflict_error` | Resource conflict |
| 422 | `unprocessable_entity_error` | Validation error |
| 429 | `rate_limit_error` | Rate limit exceeded |
| 500 | `internal_server_error` | OpenAI server error |
| 503 | `service_unavailable` | Service temporarily unavailable |

---

## Additional Resources

- [Official API Reference](https://platform.openai.com/docs/api-reference)
- [OpenAI Cookbook](https://cookbook.openai.com)
- [Community Forum](https://community.openai.com)
