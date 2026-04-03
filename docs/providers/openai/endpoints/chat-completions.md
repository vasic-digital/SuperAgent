# OpenAI Chat Completions API

## Endpoint

```
POST https://api.openai.com/v1/chat/completions
```

Creates a model response for the given chat conversation.

---

## Request

### Headers

| Header | Required | Description |
|--------|----------|-------------|
| `Authorization` | Yes | `Bearer <api_key>` |
| `Content-Type` | Yes | `application/json` |
| `OpenAI-Organization` | No | Organization ID |
| `OpenAI-Project` | No | Project ID |

### Request Body

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `model` | string | Yes | - | Model ID (e.g., `gpt-4o`) |
| `messages` | array | Yes | - | Array of message objects |
| `frequency_penalty` | number | No | 0 | -2.0 to 2.0 |
| `logit_bias` | object | No | null | Token likelihood modification |
| `logprobs` | boolean | No | false | Return log probabilities |
| `top_logprobs` | integer | No | null | 0-20 most likely tokens |
| `max_tokens` | integer | No | null | Maximum tokens to generate |
| `max_completion_tokens` | integer | No | null | Max completion tokens (o1/o3) |
| `n` | integer | No | 1 | Number of completions |
| `presence_penalty` | number | No | 0 | -2.0 to 2.0 |
| `response_format` | object | No | null | Output format |
| `seed` | integer | No | null | Deterministic sampling seed |
| `stop` | string/array | No | null | Stop sequences |
| `stream` | boolean | No | false | Stream via SSE |
| `stream_options` | object | No | null | Streaming options |
| `temperature` | number | No | 1 | 0 to 2 |
| `top_p` | number | No | 1 | Nucleus sampling |
| `tools` | array | No | null | Available tools |
| `tool_choice` | string/object | No | "auto" | Tool selection |
| `parallel_tool_calls` | boolean | No | true | Parallel tool calls |
| `user` | string | No | null | End-user identifier |

### Message Object

```typescript
{
  "role": "system" | "user" | "assistant" | "tool",
  "content": string | Array<ContentPart>,
  "name"?: string,
  "tool_calls"?: Array<ToolCall>,
  "tool_call_id"?: string
}
```

### Content Part Types

**Text:**
```json
{"type": "text", "text": "Hello"}
```

**Image:**
```json
{
  "type": "image_url",
  "image_url": {
    "url": "https://..." | "data:image/jpeg;base64,...",
    "detail": "auto" | "low" | "high"
  }
}
```

**Audio:**
```json
{
  "type": "input_audio",
  "input_audio": {
    "data": "base64_audio",
    "format": "wav" | "mp3"
  }
}
```

---

## Response

### Success (200 OK)

```json
{
  "id": "chatcmpl-123",
  "object": "chat.completion",
  "created": 1677652288,
  "model": "gpt-4o",
  "system_fingerprint": "fp_44709d6fcb",
  "choices": [{
    "index": 0,
    "message": {
      "role": "assistant",
      "content": "Hello! How can I help?"
    },
    "logprobs": null,
    "finish_reason": "stop"
  }],
  "usage": {
    "prompt_tokens": 9,
    "completion_tokens": 12,
    "total_tokens": 21,
    "prompt_tokens_details": {"cached_tokens": 0},
    "completion_tokens_details": {"reasoning_tokens": 0}
  }
}
```

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique identifier |
| `object` | string | `chat.completion` |
| `created` | integer | Unix timestamp |
| `model` | string | Model used |
| `system_fingerprint` | string | Backend config |
| `choices` | array | Completions |
| `usage` | object | Token stats |

---

## Streaming (SSE)

Request:
```json
{"model": "gpt-4o", "messages": [...], "stream": true}
```

Response:
```
data: {"id":"...","object":"chat.completion.chunk","choices":[{"delta":{"content":"Hello"}}]}

data: {"id":"...","object":"chat.completion.chunk","choices":[{"delta":{"content":"!"}}]}

data: [DONE]
```

---

## Examples

### Basic
```bash
curl https://api.openai.com/v1/chat/completions \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-4o","messages":[{"role":"user","content":"Hello"}]}'
```

### Streaming
```bash
curl https://api.openai.com/v1/chat/completions \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-4o","messages":[{"role":"user","content":"Count"}],"stream":true}'
```

### Vision
```bash
curl https://api.openai.com/v1/chat/completions \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [{
      "role": "user",
      "content": [
        {"type": "text", "text": "Describe"},
        {"type": "image_url", "image_url": {"url": "https://..."}}
      ]
    }]
  }'
```

### JSON Mode
```bash
curl https://api.openai.com/v1/chat/completions \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "List colors"}],
    "response_format": {"type": "json_object"}
  }'
```

---

## Error Codes

| Code | Description |
|------|-------------|
| 400 | Invalid request |
| 401 | Invalid API key |
| 429 | Rate limit exceeded |
| 500 | Server error |

