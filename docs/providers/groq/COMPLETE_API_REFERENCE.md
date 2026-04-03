# Groq API - Complete Reference
## Ultra-Low Latency Inference

**Provider:** Groq  
**Base URL:** `https://api.groq.com/openai/v1`  
**Docs:** https://console.groq.com/docs  
**Specialty:** Ultra-fast inference (100+ tok/sec)  

---

## Authentication

### Bearer Token
```http
Authorization: Bearer {GROQ_API_KEY}
Content-Type: application/json
```

---

## Rate Limits

| Tier | RPM | RPD | TPM | Notes |
|------|-----|-----|-----|-------|
| Free | 20 | 1,440 | 25,000 | Development |
| Production | 100+ | - | 500,000+ | Paid tier |

---

## Models

### Llama 3 Series

| Model | Context | Developer | Notes |
|-------|---------|-----------|-------|
| llama-3.3-70b-versatile | 128K | Meta | Most capable |
| llama-3.1-8b-instant | 128K | Meta | Fast |
| llama3-70b-8192 | 8K | Meta | Legacy |
| llama3-8b-8192 | 8K | Meta | Legacy |

### Mixtral

| Model | Context | Developer |
|-------|---------|-----------|
| mixtral-8x7b-32768 | 32K | Mistral |

### Gemma

| Model | Context | Developer |
|-------|---------|-----------|
| gemma-7b-it | 8K | Google |
| gemma2-9b-it | 8K | Google |

### Other

| Model | Context | Developer |
|-------|---------|-----------|
| qwen-2.5-32b | 32K | Alibaba |
| deepseek-r1-distill-llama-70b | 128K | DeepSeek |

---

## Endpoints

### OpenAI Compatible
Groq uses OpenAI-compatible endpoints:

#### POST /openai/v1/chat/completions

**Request:**
```json
{
  "model": "llama-3.3-70b-versatile",
  "messages": [
    {"role": "system", "content": "You are helpful."},
    {"role": "user", "content": "Hello!"}
  ],
  "temperature": 0.7,
  "max_tokens": 1024,
  "stream": true
}
```

---

## ⚡ Ultra-Low Latency Optimizations (CRITICAL)

Groq is optimized for speed. CLI agents implement special handling:

### 1. Pre-warm Connections
```python
import requests
from requests.adapters import HTTPAdapter

session = requests.Session()
adapter = HTTPAdapter(
    pool_connections=100,
    pool_maxsize=100,
    max_retries=3
)
session.mount('https://', adapter)
```

### 2. Streaming Always
```python
# Always use streaming for real-time feedback
stream = client.chat.completions.create(
    model="llama-3.3-70b-versatile",
    messages=messages,
    stream=True  # Essential for speed perception
)
```

### 3. JSON Mode
```python
# JSON mode reduces parsing overhead
response = client.chat.completions.create(
    model="llama-3.3-70b-versatile",
    messages=messages,
    response_format={"type": "json_object"}
)
```

### 4. Tool Use (Llama 3.1+)
```python
# Native tool calling support
response = client.chat.completions.create(
    model="llama-3.3-70b-versatile",
    messages=messages,
    tools=tools,
    tool_choice="auto"
)
```

---

## CLI Agent Workarounds

### Speed Optimization
```python
# Groq-specific optimizations
config = {
    "model": "llama-3.3-70b-versatile",
    "temperature": 0.5,  # Lower for faster convergence
    "max_tokens": 1024,  # Limit for speed
    "stream": True,      # Always stream
}
```

### Error Handling
Groq has strict rate limits - implement aggressive handling:
```python
max_retries = 10
for attempt in range(max_retries):
    try:
        return groq_request()
    except RateLimitError:
        # Groq requires precise backoff
        time.sleep(0.5 * (2 ** attempt))
```

---

**Document Status:** ✅ Complete  
**Last Updated:** 2026-04-03
