# Anthropic Claude API - Complete Reference
## All Endpoints, All Models, All Workarounds

**Provider:** Anthropic  
**Base URL:** `https://api.anthropic.com/v1`  
**Docs:** https://docs.anthropic.com  
**Latest Models:** Claude 3.7 Sonnet, Claude 3.5 Sonnet, Claude 3 Opus, Claude 3 Haiku  

---

## Authentication

### Header Authentication
```http
x-api-key: {ANTHROPIC_API_KEY}
anthropic-version: 2023-06-01
Content-Type: application/json
anthropic-beta: prompt-caching-2024-07-31  # Optional beta features
```

### Beta Features Headers
```http
# Prompt caching
anthropic-beta: prompt-caching-2024-07-31

# Computer use (beta)
anthropic-beta: computer-use-2024-10-22

# Extended thinking (3.7 Sonnet)
anthropic-beta: max-tokens-3-5-sonnet-2024-07-15
```

---

## Rate Limits

### By Tier

| Tier | Requests/min | Requests/day | Tokens/min | Concurrent Requests |
|------|--------------|--------------|------------|---------------------|
| Free | 5 | 100 | 25,000 | 1 |
| Build | 50 | 1,000 | 50,000 | 3 |
| Scale | 1,000 | 50,000 | 200,000 | 20 |
| Enterprise | Custom | Custom | Custom | Custom |

### Rate Limit Headers
```http
anthropic-ratelimit-requests-limit: 1000
anthropic-ratelimit-requests-remaining: 999
anthropic-ratelimit-requests-reset: 2024-01-01T00:00:00Z
anthropic-ratelimit-tokens-limit: 40000
anthropic-ratelimit-tokens-remaining: 39950
anthropic-ratelimit-tokens-reset: 2024-01-01T00:00:00Z
retry-after: 60  # seconds
```

### Rate Limit Response (429)
```json
{
  "type": "error",
  "error": {
    "type": "rate_limit_error",
    "message": "Rate limit exceeded. Please wait before retrying."
  }
}

// 529 - Overloaded (Anthropic-specific)
{
  "type": "error",
  "error": {
    "type": "overloaded_error",
    "message": "Anthropic's API is temporarily overloaded"
  }
}
```

**CLI Agent Workaround:** Implement longer backoff for 529 errors (5-10 seconds)

---

## Models

### Claude 3.7 Sonnet (Latest)

| Model | Context | Max Output | Input Cost | Output Cost | Notes |
|-------|---------|------------|------------|-------------|-------|
| claude-3-7-sonnet-20250219 | 200K | 128K | $3.00/M | $15.00/M | Extended thinking |
| claude-3-7-sonnet-latest | 200K | 128K | $3.00/M | $15.00/M | Alias to latest |

**Features:**
- Extended thinking mode (up to 128K tokens)
- Vision capabilities
- Tool use
- Computer use (beta)

### Claude 3.5 Sonnet

| Model | Context | Max Output | Input Cost | Output Cost | Notes |
|-------|---------|------------|------------|-------------|-------|
| claude-3-5-sonnet-20241022 | 200K | 8K | $3.00/M | $15.00/M | New version |
| claude-3-5-sonnet-20240620 | 200K | 8K | $3.00/M | $15.00/M | Original |
| claude-3-5-sonnet-latest | 200K | 8K | $3.00/M | $15.00/M | Alias |

### Claude 3.5 Haiku

| Model | Context | Max Output | Input Cost | Output Cost | Notes |
|-------|---------|------------|------------|-------------|-------|
| claude-3-5-haiku-20241022 | 200K | 4K | $0.80/M | $4.00/M | Fast |
| claude-3-5-haiku-latest | 200K | 4K | $0.80/M | $4.00/M | Alias |

### Claude 3 Opus

| Model | Context | Max Output | Input Cost | Output Cost | Notes |
|-------|---------|------------|------------|-------------|-------|
| claude-3-opus-20240229 | 200K | 4K | $15.00/M | $75.00/M | Most capable |
| claude-3-opus-latest | 200K | 4K | $15.00/M | $75.00/M | Alias |

### Context Window Costs (with Caching)

| Model | Input | Cache Write | Cache Read | Output |
|-------|-------|-------------|------------|--------|
| Claude 3.7 Sonnet | $3.00/M | $3.75/M | $0.30/M | $15.00/M |
| Claude 3.5 Sonnet | $3.00/M | $3.75/M | $0.30/M | $15.00/M |
| Claude 3.5 Haiku | $0.80/M | $1.00/M | $0.08/M | $4.00/M |
| Claude 3 Opus | $15.00/M | $18.75/M | $1.50/M | $75.00/M |
| Claude 3 Haiku | $0.25/M | $0.30/M | $0.03/M | $1.25/M |

---

## Endpoints

### 1. Messages

#### POST /v1/messages

**Description:** Primary endpoint for Claude interactions

**Request:**
```json
{
  "model": "claude-3-5-sonnet-20241022",
  "max_tokens": 4096,
  "messages": [
    {"role": "user", "content": "Hello, Claude!"}
  ],
  "system": "You are a helpful assistant.",
  "temperature": 1.0,
  "top_p": 0.999,
  "top_k": 0,
  "metadata": {"user_id": "user_123"},
  "stop_sequences": [],
  "stream": false,
  "tools": null,
  "tool_choice": {"type": "auto"},
  "thinking": null
}
```

**Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| model | string | ✅ | - | Model ID |
| messages | array | ✅ | - | Conversation history |
| max_tokens | integer | ✅ | - | Max output tokens (1-128K) |
| system | string | ❌ | - | System prompt |
| temperature | number | ❌ | 1.0 | Sampling randomness (0-1) |
| top_p | number | ❌ | 0.999 | Nucleus sampling |
| top_k | integer | ❌ | 0 | Top-k sampling (0-500) |
| metadata | object | ❌ | {} | Request metadata |
| stop_sequences | array | ❌ | [] | Stop sequences |
| stream | boolean | ❌ | false | SSE streaming |
| tools | array | ❌ | null | Available tools |
| tool_choice | object | ❌ | auto | Tool selection |
| thinking | object | ❌ | null | Extended thinking config |

**Multi-modal Content:**
```json
{
  "role": "user",
  "content": [
    {
      "type": "text",
      "text": "What's in this image?"
    },
    {
      "type": "image",
      "source": {
        "type": "base64",
        "media_type": "image/jpeg",
        "data": "/9j/4AAQ..."
      }
    }
  ]
}
```

**Image Types:** image/jpeg, image/png, image/gif, image/webp

**Response (Non-streaming):**
```json
{
  "id": "msg_01Xg9X4qZ7fF1xGH8dQjP1JY",
  "type": "message",
  "role": "assistant",
  "model": "claude-3-5-sonnet-20241022",
  "content": [
    {
      "type": "text",
      "text": "Hello! How can I help you today?"
    }
  ],
  "stop_reason": "end_turn",
  "stop_sequence": null,
  "usage": {
    "input_tokens": 12,
    "output_tokens": 10,
    "cache_creation_input_tokens": 0,
    "cache_read_input_tokens": 0
  }
}
```

**Response with Tool Use:**
```json
{
  "id": "msg_01Xg9X4qZ7fF1xGH8dQjP1JY",
  "type": "message",
  "role": "assistant",
  "content": [
    {
      "type": "text",
      "text": "I'll check the weather for you."
    },
    {
      "type": "tool_use",
      "id": "toolu_01THpBzsV4aXCBPNY8K9hPfS",
      "name": "get_weather",
      "input": {"location": "San Francisco, CA"}
    }
  ],
  "stop_reason": "tool_use",
  "usage": {"input_tokens": 45, "output_tokens": 32}
}
```

**Extended Thinking (3.7 Sonnet):**
```json
{
  "content": [
    {
      "type": "thinking",
      "thinking": "Let me break this down step by step...",
      "signature": "Ep8DCkYICxgCKkAC..."
    },
    {
      "type": "text",
      "text": "The answer is 42."
    }
  ],
  "usage": {
    "input_tokens": 20,
    "output_tokens": 150,
    "thinking_tokens": 100
  }
}
```

**Streaming Response (SSE):**
```
event: message_start
data: {"type":"message_start","message":{"id":"msg_01...","type":"message","role":"assistant"}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":10}}

event: message_stop
data: {"type":"message_stop"}
```

---

### 2. Message Counting

#### POST /v1/messages/count_tokens

**Request:**
```json
{
  "model": "claude-3-5-sonnet-20241022",
  "messages": [{"role": "user", "content": "Hello, world!"}],
  "system": "You are helpful.",
  "tools": null
}
```

**Response:**
```json
{"input_tokens": 15}
```

**CLI Agent Workaround:** Pre-count tokens before every request to avoid limit errors

---

## Tool Use

### Tool Definition
```json
{
  "name": "get_weather",
  "description": "Get the current weather in a given location",
  "input_schema": {
    "type": "object",
    "properties": {
      "location": {
        "type": "string",
        "description": "The city and state, e.g. San Francisco, CA"
      },
      "unit": {
        "type": "string",
        "enum": ["celsius", "fahrenheit"]
      }
    },
    "required": ["location"]
  }
}
```

### Tool Choice Options
```json
// Auto - let Claude decide
{"tool_choice": {"type": "auto"}}

// Any - must use at least one tool
{"tool_choice": {"type": "any"}}

// Tool - must use specific tool
{"tool_choice": {"type": "tool", "name": "get_weather"}}
```

### Tool Result Message
```json
{
  "role": "user",
  "content": [
    {
      "type": "tool_result",
      "tool_use_id": "toolu_01THpBzsV4aXCBPNY8K9hPfS",
      "content": "{\"temperature\": 72, \"unit\": \"fahrenheit\"}",
      "is_error": false
    }
  ]
}
```

---

## Extended Thinking (Claude 3.7 Sonnet)

### Configuration
```json
{
  "model": "claude-3-7-sonnet-20250219",
  "max_tokens": 128000,
  "thinking": {
    "type": "enabled",
    "budget_tokens": 32000  // Must be >= 1024, <= max_tokens - 1024
  },
  "messages": [...]
}
```

---

## Prompt Caching (Beta)

### Cache Control
```json
{
  "role": "user",
  "content": [
    {
      "type": "text",
      "text": "Large document content...",
      "cache_control": {"type": "ephemeral"}
    }
  ]
}
```

**Cost Savings:**
- Cache write: 1.25x base cost
- Cache read: 0.10x base cost (90% savings!)
- Regular input: 1.0x base cost

**Usage in Response:**
```json
{
  "usage": {
    "input_tokens": 10000,
    "output_tokens": 500,
    "cache_creation_input_tokens": 5000,
    "cache_read_input_tokens": 5000
  }
}
```

---

## CLI Agent Workarounds

### 1. Claude Code Optimizations

**Request Queuing:**
```python
class ClaudeRateLimiter:
    def __init__(self):
        self.semaphore = asyncio.Semaphore(3)  # Concurrent limit
    
    async def request(self, func):
        async with self.semaphore:
            return await self._retry_with_backoff(func)
```

**Context Management:**
- Automatic conversation compression with `/compact`
- Repository map for large codebases
- Token counting before every request

### 2. 200K Context Window Management

**Intelligent Truncation:**
```python
# Keep system prompt + recent messages
# Summarize older messages
messages = [system_message] + last_10_messages + summary_of_older
```

**File Chunking:**
```python
def chunk_file(content, chunk_size=100000):  # tokens
    tokens = tokenize(content)
    for i in range(0, len(tokens), chunk_size):
        yield detokenize(tokens[i:i+chunk_size])
```

### 3. Retry Strategy (Critical for 529 errors)

```python
import random

def anthropic_retry(func, max_retries=5):
    for attempt in range(max_retries):
        try:
            return func()
        except RateLimitError:
            delay = (2 ** attempt) + random.uniform(1, 3)
            time.sleep(delay)
        except APIError as e:
            if e.status_code == 529:  # Overloaded - special handling
                delay = 5 + random.uniform(0, 5)  # Longer backoff
                time.sleep(delay)
                continue
            if e.status_code >= 500:
                time.sleep(1)
                continue
            raise
```

### 4. SSE Buffer Optimization

```go
// Claude Code uses 4KB buffers
bufferSize := 4 * 1024  // Optimal for token streaming

// Aider uses 1KB for fast typing feedback
bufferSize := 1 * 1024
```

---

## SDK Examples

### Python
```python
import anthropic

client = anthropic.Anthropic()

# Basic completion
message = client.messages.create(
    model="claude-3-5-sonnet-20241022",
    max_tokens=4096,
    messages=[{"role": "user", "content": "Hello!"}]
)

# With thinking
message = client.messages.create(
    model="claude-3-7-sonnet-20250219",
    max_tokens=64000,
    thinking={"type": "enabled", "budget_tokens": 16000},
    messages=[{"role": "user", "content": "Complex problem..."}]
)

# Streaming
stream = client.messages.create(
    model="claude-3-5-sonnet-20241022",
    max_tokens=4096,
    messages=[{"role": "user", "content": "Count to 10"}],
    stream=True
)

for chunk in stream:
    if chunk.type == "content_block_delta":
        print(chunk.delta.text, end="")
```

---

**Document Status:** ✅ Complete  
**Last Updated:** 2026-04-03
