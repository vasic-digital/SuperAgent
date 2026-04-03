# DeepSeek API - Complete Reference
## All Endpoints, All Models, Stability Workarounds

**Provider:** DeepSeek  
**Base URL:** `https://api.deepseek.com/v1`  
**Docs:** https://platform.deepseek.com  
**Latest Models:** DeepSeek V3, DeepSeek R1, DeepSeek Coder  

---

## ⚠️ Stability Notice

**DeepSeek API is known for instability.** All CLI agents implement aggressive workarounds:
- 10+ retry attempts (vs 3-5 for other providers)
- Longer backoff delays
- Fallback provider switching
- Connection keep-alive
- Request coalescing

---

## Authentication

### Bearer Token
```http
Authorization: Bearer {DEEPSEEK_API_KEY}
Content-Type: application/json
```

---

## Rate Limits

| Tier | RPM | TPM | Notes |
|------|-----|-----|-------|
| Free | 10 | - | Limited access |
| Pro | 100+ | - | Higher limits |
| Enterprise | Custom | Custom | Contact sales |

**Note:** Rate limits are subject to change based on API stability.

---

## Models

### DeepSeek V3 (Chat)

| Model | Context | Output | Input Cost | Output Cost | Notes |
|-------|---------|--------|------------|-------------|-------|
| deepseek-chat | 64K | 8K | $0.14/M | $0.28/M | General purpose |
| deepseek-v3 | 64K | 8K | $0.14/M | $0.28/M | Alias |

**Features:**
- General conversation
- Code generation
- Reasoning
- Chinese language optimized

### DeepSeek R1 (Reasoning)

| Model | Context | Output | Input Cost | Output Cost | Notes |
|-------|---------|--------|------------|-------------|-------|
| deepseek-reasoner | 64K | 8K | $0.55/M | $2.19/M | Chain-of-thought |
| deepseek-r1 | 64K | 8K | $0.55/M | $2.19/M | Alias |

**Features:**
- Shows reasoning process
- Step-by-step problem solving
- Mathematical reasoning
- Code reasoning

**Important:** R1 returns reasoning content separately:
```json
{
  "choices": [{
    "message": {
      "content": "Final answer...",
      "reasoning_content": "Step 1: Analyze... Step 2: ..."
    }
  }]
}
```

### DeepSeek Coder

| Model | Context | Output | Notes |
|-------|---------|--------|-------|
| deepseek-coder | 16K | 4K | Code-specific |

**Features:**
- Code completion
- Code explanation
- Bug fixing
- Multiple languages

---

## Endpoints

### 1. Chat Completions

#### POST /v1/chat/completions

**Request:**
```json
{
  "model": "deepseek-chat",
  "messages": [
    {"role": "system", "content": "You are a helpful assistant."},
    {"role": "user", "content": "Hello!"}
  ],
  "max_tokens": 4096,
  "temperature": 1.0,
  "top_p": 1.0,
  "stream": false,
  "presence_penalty": 0,
  "frequency_penalty": 0,
  "response_format": {"type": "text"}
}
```

**Response (V3):**
```json
{
  "id": "chatcmpl-123",
  "object": "chat.completion",
  "created": 1677652288,
  "model": "deepseek-chat",
  "choices": [{
    "index": 0,
    "message": {
      "role": "assistant",
      "content": "Hello! How can I help you today?"
    },
    "finish_reason": "stop"
  }],
  "usage": {
    "prompt_tokens": 20,
    "completion_tokens": 10,
    "total_tokens": 30
  }
}
```

**Response (R1 with reasoning):**
```json
{
  "id": "chatcmpl-123",
  "model": "deepseek-reasoner",
  "choices": [{
    "message": {
      "role": "assistant",
      "content": "The answer is 42.",
      "reasoning_content": "Let me think through this step by step. First, I need to understand the question..."
    },
    "finish_reason": "stop"
  }],
  "usage": {
    "prompt_tokens": 20,
    "completion_tokens": 150,
    "total_tokens": 170
  }
}
```

---

### 2. Completions (Legacy)

#### POST /v1/completions

**Request:**
```json
{
  "model": "deepseek-coder",
  "prompt": "def fibonacci(n):",
  "max_tokens": 100,
  "temperature": 0.7
}
```

---

### 3. List Models

#### GET /v1/models

**Response:**
```json
{
  "object": "list",
  "data": [
    {"id": "deepseek-chat", "object": "model"},
    {"id": "deepseek-reasoner", "object": "model"},
    {"id": "deepseek-coder", "object": "model"}
  ]
}
```

---

## Streaming

**SSE Format:**
```
data: {"choices": [{"delta": {"content": "Hello"}}]}

data: {"choices": [{"delta": {"content": "!"}}]}

data: {"choices": [{"finish_reason": "stop"}]}

data: [DONE]
```

**R1 Streaming with Reasoning:**
```
data: {"choices": [{"delta": {"reasoning_content": "Step 1:"}}]}

data: {"choices": [{"delta": {"reasoning_content": " Analyze"}}]}

data: {"choices": [{"delta": {"content": "The answer"}}]}

data: [DONE]
```

---

## JSON Mode

```json
{
  "model": "deepseek-chat",
  "messages": [{"role": "user", "content": "List 3 colors in JSON"}],
  "response_format": {"type": "json_object"}
}
```

---

## Function Calling

### Request
```json
{
  "model": "deepseek-chat",
  "messages": [{"role": "user", "content": "What's the weather?"}],
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "get_weather",
        "description": "Get weather",
        "parameters": {
          "type": "object",
          "properties": {
            "location": {"type": "string"}
          },
          "required": ["location"]
        }
      }
    }
  ]
}
```

### Response
```json
{
  "choices": [{
    "message": {
      "tool_calls": [{
        "id": "call_123",
        "type": "function",
        "function": {
          "name": "get_weather",
          "arguments": "{\"location\":\"San Francisco\"}"
        }
      }]
    }
  }]
}
```

---

## ⚠️ CLI Agent Stability Workarounds (CRITICAL)

### 1. Aggressive Retry Logic

```python
import time
import random

def deepseek_request_with_retry(func, max_retries=10):  # More than other providers
    for attempt in range(max_retries):
        try:
            return func()
        except (RateLimitError, TimeoutError, ConnectionError) as e:
            if attempt == max_retries - 1:
                raise
            
            # Exponential backoff with longer delays
            base_delay = 0.5 * (2 ** attempt)
            jitter = random.uniform(0, base_delay * 0.5)
            delay = base_delay + jitter
            
            # Cap at 60 seconds
            delay = min(delay, 60)
            
            print(f"DeepSeek request failed (attempt {attempt + 1}), retrying in {delay:.1f}s...")
            time.sleep(delay)
```

### 2. Fallback Provider Switching

```python
providers = ["deepseek", "qwen", "yi", "openai"]

def generate_with_fallback(prompt):
    for provider in providers:
        try:
            if provider == "deepseek":
                return deepseek_generate(prompt)
            elif provider == "qwen":
                return qwen_generate(prompt)
            # etc.
        except Exception as e:
            print(f"{provider} failed: {e}")
            continue
    
    raise Exception("All providers failed")
```

### 3. Connection Keep-Alive

```http
Connection: keep-alive
Keep-Alive: timeout=120, max=1000
```

```python
# Persistent session
session = requests.Session()
session.headers.update({
    "Connection": "keep-alive",
    "Keep-Alive": "timeout=120, max=1000"
})
```

### 4. Request Coalescing

```python
# Combine multiple small requests
class RequestBatcher:
    def __init__(self, batch_size=5, window_ms=100):
        self.batch_size = batch_size
        self.window_ms = window_ms
        self.pending = []
    
    def add(self, request):
        self.pending.append(request)
        if len(self.pending) >= self.batch_size:
            return self.flush()
    
    def flush(self):
        # Send batched request
        combined = self.combine_requests(self.pending)
        self.pending = []
        return self.send(combined)
```

### 5. Circuit Breaker Pattern

```python
from datetime import datetime, timedelta

class CircuitBreaker:
    def __init__(self, failure_threshold=5, timeout=300):
        self.failure_threshold = failure_threshold
        self.timeout = timeout
        self.failures = 0
        self.last_failure = None
        self.state = "CLOSED"  # CLOSED, OPEN, HALF_OPEN
    
    def call(self, func):
        if self.state == "OPEN":
            if datetime.now() - self.last_failure > timedelta(seconds=self.timeout):
                self.state = "HALF_OPEN"
            else:
                raise Exception("Circuit breaker is OPEN")
        
        try:
            result = func()
            self.on_success()
            return result
        except Exception as e:
            self.on_failure()
            raise e
    
    def on_success(self):
        self.failures = 0
        self.state = "CLOSED"
    
    def on_failure(self):
        self.failures += 1
        self.last_failure = datetime.now()
        if self.failures >= self.failure_threshold:
            self.state = "OPEN"
```

### 6. Timeout Configuration

```python
# Longer timeouts for DeepSeek
timeout_config = {
    "connect_timeout": 30,      # Connection establishment
    "read_timeout": 120,        # Response reading
    "total_timeout": 180        # Total request time
}

response = requests.post(
    url,
    headers=headers,
    json=data,
    timeout=(timeout_config["connect_timeout"], timeout_config["read_timeout"])
)
```

### 7. Health Check and Pre-warming

```python
class DeepSeekHealthMonitor:
    def __init__(self):
        self.healthy = True
        self.last_check = None
    
    def check_health(self):
        try:
            response = requests.get(
                "https://api.deepseek.com/v1/models",
                headers={"Authorization": f"Bearer {api_key}"},
                timeout=10
            )
            self.healthy = response.status_code == 200
            self.last_check = datetime.now()
        except:
            self.healthy = False
    
    def is_healthy(self):
        if self.last_check is None or datetime.now() - self.last_check > timedelta(minutes=1):
            self.check_health()
        return self.healthy
```

---

## SDK Examples

### Python
```python
import openai

client = openai.OpenAI(
    api_key="your-deepseek-api-key",
    base_url="https://api.deepseek.com/v1"
)

# Basic completion
response = client.chat.completions.create(
    model="deepseek-chat",
    messages=[
        {"role": "system", "content": "You are a helpful assistant."},
        {"role": "user", "content": "Hello!"}
    ]
)
print(response.choices[0].message.content)

# With reasoning (R1)
response = client.chat.completions.create(
    model="deepseek-reasoner",
    messages=[{"role": "user", "content": "Solve 2x + 5 = 13"}]
)
print("Reasoning:", response.choices[0].message.reasoning_content)
print("Answer:", response.choices[0].message.content)

# Streaming
response = client.chat.completions.create(
    model="deepseek-chat",
    messages=[{"role": "user", "content": "Count to 10"}],
    stream=True
)
for chunk in response:
    content = chunk.choices[0].delta.content
    if content:
        print(content, end="")
```

---

## Error Handling

### Common Errors

| Error | Cause | Workaround |
|-------|-------|------------|
| 429 Rate Limit | Too many requests | Exponential backoff |
| 503 Service Unavailable | API overloaded | Retry with longer delay |
| 504 Gateway Timeout | Request timeout | Increase timeout, retry |
| Connection Reset | Network issue | Keep-alive, retry |
| Empty Response | API error | Fallback provider |

### Error Response
```json
{
  "error": {
    "message": "Rate limit exceeded",
    "type": "rate_limit_error",
    "code": "rate_limit_exceeded"
  }
}
```

---

## Best Practices

1. **Always implement aggressive retry logic** (10+ attempts)
2. **Use fallback providers** for critical operations
3. **Enable connection keep-alive**
4. **Implement circuit breaker** to prevent cascade failures
5. **Monitor API health** before sending requests
6. **Use request coalescing** for batch operations
7. **Set longer timeouts** than other providers
8. **Cache responses** when possible

---

**Document Status:** ✅ Complete  
**Last Updated:** 2026-04-03  
**Stability Warning:** DeepSeek API requires special handling for production use
