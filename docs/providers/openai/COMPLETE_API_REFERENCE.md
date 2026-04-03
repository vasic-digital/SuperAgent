# OpenAI API - Complete Reference
## All Endpoints, All Models, All Workarounds

**Provider:** OpenAI  
**Base URL:** `https://api.openai.com/v1`  
**Docs:** https://platform.openai.com/docs  
**Latest Models:** GPT-4o, GPT-4o-mini, o1, o3, Codex  

---

## Authentication

### Header Authentication
```http
Authorization: Bearer {OPENAI_API_KEY}
Content-Type: application/json
OpenAI-Organization: {ORG_ID}      # Optional
OpenAI-Project: {PROJECT_ID}       # Optional
OpenAI-Beta: assistants=v2         # For beta features
```

### Organization & Project Scoping
```bash
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -H "OpenAI-Organization: org-123" \
  -H "OpenAI-Project: proj_abc"
```

---

## Rate Limits

### By Tier

| Tier | RPM | TPM | Batch Queue Limit |
|------|-----|-----|-------------------|
| Free | 3 | 40,000 | 0 |
| Tier 1 | 500 | 200,000 | 100,000 |
| Tier 2 | 5,000 | 2,000,000 | 2,000,000 |
| Tier 3 | 10,000 | 10,000,000 | 10,000,000 |
| Tier 4 | 20,000 | 40,000,000 | 40,000,000 |
| Tier 5 | 50,000 | 150,000,000 | 150,000,000 |

### Rate Limit Headers
```http
x-ratelimit-limit-requests: 500
x-ratelimit-remaining-requests: 499
x-ratelimit-reset-requests: 1s
x-ratelimit-limit-tokens: 200000
x-ratelimit-remaining-tokens: 199472
x-ratelimit-reset-tokens: 6s
```

---

## Models

### GPT-4o Series (Flagship)

| Model | Context | Max Output | Input Cost | Output Cost | Vision | Tools | Streaming |
|-------|---------|------------|------------|-------------|--------|-------|-----------|
| gpt-4o | 128K | 16K | $2.50/M | $10.00/M | ✅ | ✅ | ✅ |
| gpt-4o-2024-11-20 | 128K | 16K | $2.50/M | $10.00/M | ✅ | ✅ | ✅ |
| gpt-4o-2024-08-06 | 128K | 16K | $2.50/M | $10.00/M | ✅ | ✅ | ✅ |
| gpt-4o-mini | 128K | 16K | $0.15/M | $0.60/M | ✅ | ✅ | ✅ |
| gpt-4o-mini-2024-07-18 | 128K | 16K | $0.15/M | $0.60/M | ✅ | ✅ | ✅ |
| gpt-4o-audio-preview | 128K | 16K | $2.50/M | $10.00/M | ✅ | ✅ | ✅ |

### GPT-4 Series

| Model | Context | Max Output | Input Cost | Output Cost | Notes |
|-------|---------|------------|------------|-------------|-------|
| gpt-4-turbo | 128K | 4K | $10.00/M | $30.00/M | Deprecated |
| gpt-4-turbo-2024-04-09 | 128K | 4K | $10.00/M | $30.00/M | Deprecated |
| gpt-4 | 8K | 8K | $30.00/M | $60.00/M | Legacy |
| gpt-4-32k | 32K | 8K | $60.00/M | $120.00/M | Legacy |

### o1 Series (Reasoning)

| Model | Context | Max Output | Input Cost | Output Cost | Notes |
|-------|---------|------------|------------|-------------|-------|
| o1 | 200K | 100K | $15.00/M | $60.00/M | Reasoning model |
| o1-2024-12-17 | 200K | 100K | $15.00/M | $60.00/M | |
| o1-preview | 128K | 32K | $15.00/M | $60.00/M | Limited access |
| o1-mini | 128K | 65K | $3.00/M | $12.00/M | Fast reasoning |
| o1-mini-2024-09-12 | 128K | 65K | $3.00/M | $12.00/M | |

### o3 Series (Next-gen Reasoning)

| Model | Context | Max Output | Input Cost | Output Cost | Notes |
|-------|---------|------------|------------|-------------|-------|
| o3-mini | 200K | 100K | $1.10/M | $4.40/M | Cost-effective reasoning |
| o3-mini-2025-01-31 | 200K | 100K | $1.10/M | $4.40/M | |

### GPT-3.5 Series (Legacy)

| Model | Context | Max Output | Input Cost | Output Cost | Status |
|-------|---------|------------|------------|-------------|--------|
| gpt-3.5-turbo | 16K | 4K | $0.50/M | $1.50/M | Deprecated |
| gpt-3.5-turbo-0125 | 16K | 4K | $0.50/M | $1.50/M | Deprecated |
| gpt-3.5-turbo-1106 | 16K | 4K | $0.50/M | $1.50/M | Deprecated |

### Embeddings

| Model | Dimensions | Max Input | Cost |
|-------|------------|-----------|------|
| text-embedding-3-large | 3072 | 8K | $0.13/M |
| text-embedding-3-small | 1536 | 8K | $0.02/M |
| text-embedding-ada-002 | 1536 | 8K | $0.10/M |

### Audio

| Model | Cost | Notes |
|-------|------|-------|
| whisper-1 | $0.006/minute | Speech-to-text |
| tts-1 | $15.00/1M chars | Text-to-speech |
| tts-1-hd | $30.00/1M chars | HD text-to-speech |
| gpt-4o-transcribe | $0.006/minute | End-to-end transcription |
| gpt-4o-mini-transcribe | $0.003/minute | Fast transcription |

---

## Endpoints

### 1. Chat Completions

#### POST /v1/chat/completions

**Description:** Main endpoint for chat-based interactions

**Request:**
```json
{
  "model": "gpt-4o",
  "messages": [
    {"role": "system", "content": "You are a helpful assistant."},
    {"role": "user", "content": "Hello!"}
  ],
  "max_tokens": 4096,
  "temperature": 1.0,
  "top_p": 1.0,
  "n": 1,
  "stream": false,
  "stop": null,
  "presence_penalty": 0,
  "frequency_penalty": 0,
  "logit_bias": {},
  "user": "user-id",
  "response_format": {"type": "text"},
  "seed": null,
  "tools": null,
  "tool_choice": "auto",
  "parallel_tool_calls": true
}
```

**Parameters:**

| Parameter | Type | Required | Default | Range | Description |
|-----------|------|----------|---------|-------|-------------|
| model | string | ✅ | - | - | Model ID |
| messages | array | ✅ | - | - | Conversation history |
| max_tokens | integer | ❌ | inf | 1-4096 | Max output tokens |
| temperature | number | ❌ | 1.0 | 0-2 | Sampling randomness |
| top_p | number | ❌ | 1.0 | 0-1 | Nucleus sampling |
| n | integer | ❌ | 1 | 1-128 | Number of completions |
| stream | boolean | ❌ | false | - | SSE streaming |
| stop | string/array | ❌ | null | - | Stop sequences |
| presence_penalty | number | ❌ | 0 | -2 to 2 | Repeat penalty |
| frequency_penalty | number | ❌ | 0 | -2 to 2 | Frequency penalty |
| logit_bias | object | ❌ | {} | -100 to 100 | Token bias |
| user | string | ❌ | - | - | End-user ID |
| response_format | object | ❌ | text | - | Output format |
| seed | integer | ❌ | null | - | Determinism seed |
| tools | array | ❌ | null | - | Available tools |
| tool_choice | string/object | ❌ | auto | - | Tool selection |
| parallel_tool_calls | boolean | ❌ | true | - | Parallel tool calls |

**Response (Non-streaming):**
```json
{
  "id": "chatcmpl-123",
  "object": "chat.completion",
  "created": 1677652288,
  "model": "gpt-4o",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Hello! How can I help you today?",
        "tool_calls": null,
        "function_call": null
      },
      "logprobs": null,
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 20,
    "completion_tokens": 10,
    "total_tokens": 30,
    "prompt_tokens_details": {
      "cached_tokens": 0
    }
  },
  "system_fingerprint": "fp_44709d6fcb"
}
```

**Response (Streaming - SSE):**
```
data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1677652288,"model":"gpt-4o","choices":[{"index":0,"delta":{"role":"assistant"},"finish_reason":null}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1677652288,"model":"gpt-4o","choices":[{"index":0,"delta":{"content":"Hello"},"finish_reason":null}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1677652288,"model":"gpt-4o","choices":[{"index":0,"delta":{"content":"!"},"finish_reason":null}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1677652288,"model":"gpt-4o","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}

data: [DONE]
```

**Error Responses:**
```json
// 400 - Bad Request
{
  "error": {
    "message": "Invalid model ID",
    "type": "invalid_request_error",
    "param": "model",
    "code": "model_not_found"
  }
}

// 401 - Unauthorized
{
  "error": {
    "message": "Incorrect API key provided",
    "type": "authentication_error",
    "code": "invalid_api_key"
  }
}

// 429 - Rate Limit
{
  "error": {
    "message": "Rate limit exceeded",
    "type": "rate_limit_error",
    "code": "rate_limit_exceeded"
  }
}

// 500 - Server Error
{
  "error": {
    "message": "The server had an error",
    "type": "server_error",
    "code": null
  }
}
```

**CLI Agent Workarounds:**

1. **Claude Code:** Implements request prediction to pre-warm connections
2. **Codex:** Uses aggressive streaming with 8KB buffers
3. **Aider:** Batches similar requests when possible
4. **Gemini CLI:** Falls back to non-streaming on 429 errors

---

### 2. Completions (Legacy)

#### POST /v1/completions

**Description:** Legacy endpoint for text completion (not chat)

**Request:**
```json
{
  "model": "gpt-3.5-turbo-instruct",
  "prompt": "Once upon a time",
  "max_tokens": 100,
  "temperature": 0.7,
  "top_p": 1,
  "n": 1,
  "stream": false,
  "logprobs": null,
  "echo": false,
  "stop": null,
  "presence_penalty": 0,
  "frequency_penalty": 0,
  "best_of": 1,
  "logit_bias": {},
  "user": "user-id"
}
```

**Note:** Only works with `gpt-3.5-turbo-instruct` and legacy models.

---

### 3. Embeddings

#### POST /v1/embeddings

**Request:**
```json
{
  "input": "The quick brown fox",
  "model": "text-embedding-3-large",
  "encoding_format": "float",
  "dimensions": 3072,
  "user": "user-id"
}
```

**Response:**
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
  "model": "text-embedding-3-large",
  "usage": {
    "prompt_tokens": 8,
    "total_tokens": 8
  }
}
```

**Batch Embeddings:**
```json
{
  "input": [
    "First text",
    "Second text",
    "Third text"
  ],
  "model": "text-embedding-3-small"
}
```

---

### 4. Models

#### GET /v1/models

**Response:**
```json
{
  "object": "list",
  "data": [
    {
      "id": "gpt-4o",
      "object": "model",
      "created": 1687882411,
      "owned_by": "openai"
    },
    {
      "id": "gpt-4o-mini",
      "object": "model",
      "created": 1721172741,
      "owned_by": "openai"
    }
  ]
}
```

#### GET /v1/models/{model}

**Response:**
```json
{
  "id": "gpt-4o",
  "object": "model",
  "created": 1687882411,
  "owned_by": "openai"
}
```

---

### 5. Images (DALL-E)

#### POST /v1/images/generations

**Request:**
```json
{
  "model": "dall-e-3",
  "prompt": "A cute baby sea otter",
  "n": 1,
  "size": "1024x1024",
  "quality": "standard",
  "response_format": "url",
  "style": "vivid",
  "user": "user-id"
}
```

**Parameters:**
| Parameter | Type | Options | Default |
|-----------|------|---------|---------|
| model | string | dall-e-2, dall-e-3 | dall-e-2 |
| prompt | string | - | - |
| n | integer | 1-10 (dall-e-2), 1 (dall-e-3) | 1 |
| size | string | 256x256, 512x512, 1024x1024, 1792x1024, 1024x1792 | 1024x1024 |
| quality | string | standard, hd | standard |
| style | string | vivid, natural | vivid |
| response_format | string | url, b64_json | url |

**Response:**
```json
{
  "created": 1677721600,
  "data": [
    {
      "url": "https://...",
      "revised_prompt": "A detailed description..."
    }
  ]
}
```

#### POST /v1/images/edits

Edit images with masks.

#### POST /v1/images/variations

Generate variations of an image.

---

### 6. Audio

#### POST /v1/audio/transcriptions

**Request (multipart/form-data):**
```bash
curl https://api.openai.com/v1/audio/transcriptions \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -H "Content-Type: multipart/form-data" \
  -F file="@audio.mp3" \
  -F model="whisper-1" \
  -F language="en" \
  -F prompt="" \
  -F response_format="json" \
  -F temperature=0 \
  -F timestamp_granularities[]=word
```

**Response:**
```json
{
  "text": "Hello, this is a test transcription.",
  "words": [
    {"word": "Hello", "start": 0.0, "end": 0.5},
    {"word": "this", "start": 0.6, "end": 0.8}
  ]
}
```

#### POST /v1/audio/translations

Translate audio to English.

#### POST /v1/audio/speech

**Request:**
```json
{
  "model": "tts-1",
  "input": "Hello world",
  "voice": "alloy",
  "response_format": "mp3",
  "speed": 1.0
}
```

**Voices:** alloy, echo, fable, onyx, nova, shimmer

---

### 7. Files

#### POST /v1/files

Upload files for fine-tuning or assistants.

#### GET /v1/files

List files.

#### GET /v1/files/{file_id}

Retrieve file.

#### DELETE /v1/files/{file_id}

Delete file.

#### GET /v1/files/{file_id}/content

Retrieve file content.

---

### 8. Fine-tuning

#### POST /v1/fine_tuning/jobs

Create fine-tuning job.

#### GET /v1/fine_tuning/jobs

List fine-tuning jobs.

#### GET /v1/fine_tuning/jobs/{fine_tuning_job_id}

Retrieve job.

#### POST /v1/fine_tuning/jobs/{fine_tuning_job_id}/cancel

Cancel job.

---

### 9. Batches

#### POST /v1/batches

Create batch job for async processing.

**Request:**
```json
{
  "input_file_id": "file-abc123",
  "endpoint": "/v1/chat/completions",
  "completion_window": "24h",
  "metadata": {
    "customer_id": "user_123"
  }
}
```

**CLI Agent Workaround:**
- Aider uses batch API for large-scale evaluations
- Claude Code doesn't use batch (real-time focus)
- Codex uses batch for code analysis jobs

---

### 10. Assistants API (Beta)

#### POST /v1/assistants

Create assistant.

#### GET /v1/assistants

List assistants.

#### POST /v1/assistants/{assistant_id}

Modify assistant.

#### DELETE /v1/assistants/{assistant_id}

Delete assistant.

#### POST /v1/threads

Create thread.

#### POST /v1/threads/{thread_id}/messages

Add message to thread.

#### POST /v1/threads/{thread_id}/runs

Create run.

---

## Message Format Deep Dive

### System Messages
```json
{
  "role": "system",
  "content": "You are a helpful assistant.",
  "name": "example_assistant"  // Optional
}
```

**Note:** o1/o3 models use "developer" role instead of "system":
```json
{
  "role": "developer",
  "content": "Formatting re-enabled"
}
```

### User Messages

**Text only:**
```json
{
  "role": "user",
  "content": "Hello!"
}
```

**Multi-modal (Vision):**
```json
{
  "role": "user",
  "content": [
    {"type": "text", "text": "What's in this image?"},
    {
      "type": "image_url",
      "image_url": {
        "url": "https://example.com/image.jpg",
        "detail": "high"  // low, high, auto
      }
    }
  ]
}
```

**Base64 image:**
```json
{
  "type": "image_url",
  "image_url": {
    "url": "data:image/jpeg;base64,/9j/4AAQ..."
  }
}
```

### Assistant Messages

**Text response:**
```json
{
  "role": "assistant",
  "content": "The image shows a cat."
}
```

**With tool calls:**
```json
{
  "role": "assistant",
  "content": null,
  "tool_calls": [
    {
      "id": "call_abc123",
      "type": "function",
      "function": {
        "name": "get_weather",
        "arguments": "{\"location\":\"San Francisco\"}"
      }
    }
  ]
}
```

### Tool Messages
```json
{
  "role": "tool",
  "tool_call_id": "call_abc123",
  "content": "{\"temperature\": 72, \"unit\": \"fahrenheit\"}"
}
```

---

## Tool/Function Calling

### Function Definition
```json
{
  "type": "function",
  "function": {
    "name": "get_weather",
    "description": "Get current weather for a location",
    "parameters": {
      "type": "object",
      "properties": {
        "location": {
          "type": "string",
          "description": "City and country"
        },
        "unit": {
          "type": "string",
          "enum": ["celsius", "fahrenheit"]
        }
      },
      "required": ["location"]
    }
  }
}
```

### Tool Choice Options
```json
// Auto - let model decide
{"tool_choice": "auto"}

// None - don't use tools
{"tool_choice": "none"}

// Required - must use a tool
{"tool_choice": "required"}

// Force specific tool
{
  "tool_choice": {
    "type": "function",
    "function": {"name": "get_weather"}
  }
}
```

**CLI Agent Workaround:**
- Claude Code: Always sets `parallel_tool_calls: true` for speed
- Aider: Uses strict mode for deterministic tool selection
- Codex: Implements tool result caching

---

## Response Formats

### JSON Mode
```json
{
  "model": "gpt-4o",
  "messages": [...],
  "response_format": {"type": "json_object"}
}
```

**Workaround for older models:**
```json
{
  "messages": [
    {"role": "system", "content": "You must respond with valid JSON only."},
    {"role": "user", "content": "List 3 colors"}
  ]
}
```

### Structured Outputs (JSON Schema)
```json
{
  "model": "gpt-4o-2024-08-06",
  "messages": [...],
  "response_format": {
    "type": "json_schema",
    "json_schema": {
      "name": "color_list",
      "strict": true,
      "schema": {
        "type": "object",
        "properties": {
          "colors": {
            "type": "array",
            "items": {"type": "string"}
          }
        },
        "required": ["colors"],
        "additionalProperties": false
      }
    }
  }
}
```

**CLI Agent Optimization:**
- Aider uses strict schemas for code generation
- Claude Code: Disables strict mode for flexibility
- Gemini CLI: Falls back to JSON mode if schema fails

---

## Advanced Features

### Predicted Outputs (Optimization)
```json
{
  "model": "gpt-4o",
  "messages": [...],
  "prediction": {
    "type": "content",
    "content": "Expected output prefix..."
  }
}
```

**Use case:** Code editing - provide original code as prediction

### Context Caching (Beta)
```http
// Not yet publicly available
// Used internally by some CLI agents
```

### Service Tiers
```json
{
  "model": "gpt-4o",
  "messages": [...],
  "service_tier": "flex"  // default, flex, priority
}
```

**Cost savings:** Flex tier is cheaper but may have higher latency

---

## Error Handling Deep Dive

### Retry Strategy
```python
import time
import random

def openai_request_with_retry(func, max_retries=5):
    for attempt in range(max_retries):
        try:
            return func()
        except RateLimitError as e:
            if attempt == max_retries - 1:
                raise
            # Exponential backoff with jitter
            delay = (2 ** attempt) + random.uniform(0, 1)
            time.sleep(delay)
        except APIError as e:
            if e.http_status >= 500:
                # Retry server errors
                time.sleep(1)
                continue
            raise
```

### CLI Agent Specific Handling

**Claude Code:**
- Implements request queuing during 429
- Falls back to smaller model on context limit
- Uses streaming cancellation for user interrupts

**Codex:**
- Aggressive prefetching of completions
- Local caching of common responses
- Connection warming on startup

**Aider:**
- Repository map compression for large codebases
- Automatic token counting before requests
- Split large files across multiple requests

**Gemini CLI:**
- Implements circuit breaker for reliability
- Brotli compression for large payloads
- HTTP/2 multiplexing for parallel requests

---

## SDK Examples

### Python (OpenAI)
```python
from openai import OpenAI

client = OpenAI(api_key="your-key")

# Chat completion
response = client.chat.completions.create(
    model="gpt-4o",
    messages=[
        {"role": "system", "content": "You are helpful."},
        {"role": "user", "content": "Hello!"}
    ],
    stream=True
)

for chunk in response:
    print(chunk.choices[0].delta.content or "", end="")
```

### JavaScript/Node.js
```javascript
import OpenAI from 'openai';

const openai = new OpenAI({
  apiKey: process.env.OPENAI_API_KEY,
});

const stream = await openai.chat.completions.create({
  model: 'gpt-4o',
  messages: [{ role: 'user', content: 'Say hello!' }],
  stream: true,
});

for await (const chunk of stream) {
  process.stdout.write(chunk.choices[0]?.delta?.content || '');
}
```

### Go
```go
import "github.com/openai/openai-go"

client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

resp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
    Model: openai.F(openai.ChatModelGPT4o),
    Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
        openai.SystemMessage("You are helpful."),
        openai.UserMessage("Hello!"),
    }),
})

if err != nil {
    log.Fatal(err)
}

fmt.Println(resp.Choices[0].Message.Content)
```

---

## Version History

| Date | Change |
|------|--------|
| 2024-11 | GPT-4o released |
| 2024-09 | o1 models released |
| 2024-08 | Structured outputs, JSON schema |
| 2024-07 | GPT-4o mini released |
| 2024-04 | GPT-4 Turbo with Vision |
| 2023-11 | GPT-4 Turbo, JSON mode |

---

**Document Status:** ✅ Complete  
**Last Updated:** 2026-04-03  
**Next Review:** 2026-05-03
