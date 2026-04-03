# Google Gemini API - Complete Reference
## All Endpoints, All Models, All Workarounds

**Provider:** Google AI (Gemini)  
**Base URL:** `https://generativelanguage.googleapis.com/v1beta`  
**Docs:** https://ai.google.dev  
**Latest Models:** Gemini 2.5 Pro, Gemini 2.5 Flash, Gemini 3 Preview  

---

## Authentication

### API Key (Query Parameter)
```http
GET /v1beta/models?key={GEMINI_API_KEY}
```

### API Key (Header)
```http
x-goog-api-key: {GEMINI_API_KEY}
Content-Type: application/json
```

### OAuth 2.0 (For user data)
```http
Authorization: Bearer {ACCESS_TOKEN}
```

---

## Rate Limits

### Free Tier
| Limit | Value |
|-------|-------|
| Requests per minute | 60 |
| Requests per day | 1,000 |
| Tokens per minute | 1,000,000 |

### Paid Tier (Gemini API)
| Limit | Value |
|-------|-------|
| Requests per minute | 1,000-10,000 (varies) |
| Tokens per minute | 4,000,000+ |

### Rate Limit Headers
```http
x-ratelimit-limit: 60
x-ratelimit-remaining: 59
x-ratelimit-reset: 60
```

---

## Models

### Gemini 2.5 Pro

| Model | Context | Output | Input Cost | Output Cost | Notes |
|-------|---------|--------|------------|-------------|-------|
| gemini-2.5-pro-preview-03-25 | 1M | 64K | $1.25/M | $10.00/M | Advanced reasoning |
| gemini-2.5-pro | 1M | 64K | $1.25/M | $10.00/M | Production |

### Gemini 2.5 Flash

| Model | Context | Output | Input Cost | Output Cost | Notes |
|-------|---------|--------|------------|-------------|-------|
| gemini-2.5-flash-preview | 1M | 8K | $0.15/M | $0.60/M | Fast, cost-effective |
| gemini-2.5-flash | 1M | 8K | $0.15/M | $0.60/M | Production |

### Gemini 3 (Preview)

| Model | Context | Output | Input Cost | Output Cost | Notes |
|-------|---------|--------|------------|-------------|-------|
| gemini-3-flash-preview | 1M | 8K | $0.15/M | $0.60/M | Latest |

### Embedding Models

| Model | Dimensions | Cost |
|-------|------------|------|
| text-embedding-004 | 768 | Free tier available |
| embedding-001 | 768 | Free tier available |

---

## Endpoints

### 1. Generate Content

#### POST /v1beta/models/{model}:generateContent

**Request:**
```json
{
  "contents": [
    {
      "role": "user",
      "parts": [
        {"text": "Explain how AI works"}
      ]
    }
  ],
  "systemInstruction": {
    "parts": [{"text": "You are a helpful assistant."}]
  },
  "generationConfig": {
    "temperature": 0.9,
    "topP": 1,
    "topK": 1,
    "maxOutputTokens": 2048,
    "stopSequences": [],
    "responseMimeType": "application/json",
    "responseSchema": {...}
  },
  "safetySettings": [
    {
      "category": "HARM_CATEGORY_DANGEROUS_CONTENT",
      "threshold": "BLOCK_ONLY_HIGH"
    }
  ],
  "tools": [...],
  "toolConfig": {...}
}
```

**Response:**
```json
{
  "candidates": [
    {
      "content": {
        "role": "model",
        "parts": [
          {"text": "AI works by..."}
        ]
      },
      "finishReason": "STOP",
      "safetyRatings": [...],
      "tokenCount": 150
    }
  ],
  "usageMetadata": {
    "promptTokenCount": 10,
    "candidatesTokenCount": 150,
    "totalTokenCount": 160
  }
}
```

---

### 2. Stream Generate Content

#### POST /v1beta/models/{model}:streamGenerateContent

**SSE Response:**
```
data: {"candidates": [{"content": {"parts": [{"text": "AI"}]}}]}

data: {"candidates": [{"content": {"parts": [{"text": " works"}]}}]}

data: {"candidates": [{"finishReason": "STOP"}]}
```

---

### 3. Count Tokens

#### POST /v1beta/models/{model}:countTokens

**Request:**
```json
{
  "contents": [
    {"parts": [{"text": "Hello, world!"}]}
  ]
}
```

**Response:**
```json
{"totalTokens": 4}
```

---

### 4. Embed Content

#### POST /v1beta/models/{model}:embedContent

**Request:**
```json
{
  "content": {
    "parts": [{"text": "Hello, world!"}]
  }
}
```

**Response:**
```json
{
  "embedding": {
    "values": [0.1, 0.2, 0.3, ...]
  }
}
```

---

### 5. Batch Embed Content

#### POST /v1beta/models/{model}:batchEmbedContents

**Request:**
```json
{
  "requests": [
    {"content": {"parts": [{"text": "First text"}]}},
    {"content": {"parts": [{"text": "Second text"}]}}
  ]
}
```

---

## Multi-modal Content

### Image Input
```json
{
  "contents": [
    {
      "role": "user",
      "parts": [
        {"text": "What's in this image?"},
        {
          "inlineData": {
            "mimeType": "image/jpeg",
            "data": "base64encoded..."
          }
        }
      ]
    }
  ]
}
```

### File Upload (Large Files)
```bash
# Upload file
curl -X POST "https://generativelanguage.googleapis.com/upload/v1beta/files?key=$API_KEY" \
  -H "Content-Type: video/mp4" \
  --data-binary @video.mp4
```

**Use uploaded file:**
```json
{
  "contents": [
    {
      "parts": [
        {"fileData": {"mimeType": "video/mp4", "fileUri": "files/abc123"}}
      ]
    }
  ]
}
```

---

## Function Calling

### Function Declaration
```json
{
  "tools": [
    {
      "functionDeclarations": [
        {
          "name": "get_weather",
          "description": "Get current weather",
          "parameters": {
            "type": "object",
            "properties": {
              "location": {"type": "string"},
              "unit": {"type": "string", "enum": ["C", "F"]}
            },
            "required": ["location"]
          }
        }
      ]
    }
  ],
  "toolConfig": {
    "functionCallingConfig": {
      "mode": "AUTO"  // AUTO, ANY, NONE
    }
  }
}
```

### Function Call Response
```json
{
  "candidates": [{
    "content": {
      "role": "model",
      "parts": [{
        "functionCall": {
          "name": "get_weather",
          "args": {"location": "San Francisco", "unit": "C"}
        }
      }]
    }
  }]
}
```

### Function Response
```json
{
  "contents": [
    {
      "role": "user",
      "parts": [{"functionResponse": {
        "name": "get_weather",
        "response": {"temperature": 22, "unit": "C"}
      }}]
    }
  ]
}
```

---

## Grounding with Google Search

### Enable Search
```json
{
  "tools": [
    {"googleSearch": {}}
  ],
  "contents": [...]
}
```

### Dynamic Retrieval
```json
{
  "tools": [
    {
      "googleSearch": {
        "dynamicRetrievalConfig": {
          "mode": "MODE_DYNAMIC",
          "dynamicThreshold": 0.5
        }
      }
    }
  ]
}
```

**Grounding Metadata in Response:**
```json
{
  "candidates": [{
    "groundingMetadata": {
      "webSearchQueries": ["current weather San Francisco"],
      "groundingChunks": [
        {
          "web": {
            "uri": "https://weather.com/...",
            "title": "Weather Forecast"
          }
        }
      ],
      "searchEntryPoint": {
        "renderedContent": "<html>...</html>"
      }
    }
  }]
}
```

---

## Code Execution

### Enable Code Execution
```json
{
  "tools": [
    {"codeExecution": {}}
  ],
  "contents": [{"parts": [{"text": "Calculate fibonacci(10)"}]}]
}
```

**Response with Code Execution:**
```json
{
  "candidates": [{
    "content": {
      "parts": [
        {"executableCode": {"language": "PYTHON", "code": "def fib(n)..."}},
        {"codeExecutionResult": {"outcome": "OUTCOME_OK", "output": "55"}}
      ]
    }
  }]
}
```

---

## Safety Settings

### Categories
- `HARM_CATEGORY_HATE_SPEECH`
- `HARM_CATEGORY_SEXUALLY_EXPLICIT`
- `HARM_CATEGORY_DANGEROUS_CONTENT`
- `HARM_CATEGORY_HARASSMENT`
- `HARM_CATEGORY_CIVIC_INTEGRITY`

### Thresholds
- `BLOCK_NONE` - Allow all
- `BLOCK_ONLY_HIGH` - Block only high probability
- `BLOCK_MEDIUM_AND_ABOVE` - Block medium and high
- `BLOCK_LOW_AND_ABOVE` - Block low, medium, high

**CLI Agent Workaround for Code Generation:**
```json
{
  "safetySettings": [
    {"category": "HARM_CATEGORY_DANGEROUS_CONTENT", "threshold": "BLOCK_NONE"},
    {"category": "HARM_CATEGORY_HARASSMENT", "threshold": "BLOCK_NONE"}
  ]
}
```

---

## CLI Agent Workarounds

### 1. Gemini CLI Optimizations

**Brotli Compression:**
```http
Accept-Encoding: br, gzip, deflate
```

**Gemini CLI always requests Brotli when available for large contexts.**

### 2. 1M Context Window Management

**File API for Large Documents:**
```python
# Upload large file
file = genai.upload_file("large_document.pdf")

# Use in conversation
response = model.generate_content([file, "Summarize this"])
```

**Automatic Chunking:**
```python
# Gemini CLI implements automatic chunking for large uploads
chunk_size = 10 * 1024 * 1024  # 10MB chunks
```

### 3. Grounding Optimization

**Cache Search Results:**
```python
# Gemini CLI caches search results for follow-up queries
@lru_cache(maxsize=100)
def search_with_cache(query):
    return model.generate_content(
        contents=[query],
        tools=[{"googleSearch": {}}]
    )
```

### 4. Connection Pooling

```python
# Persistent connections for multiple requests
session = requests.Session()
adapter = HTTPAdapter(pool_connections=50, pool_maxsize=50)
session.mount('https://', adapter)
```

---

## SDK Examples

### Python
```python
import google.generativeai as genai

genai.configure(api_key="YOUR_API_KEY")

# Basic generation
model = genai.GenerativeModel('gemini-2.5-pro')
response = model.generate_content("Explain AI")
print(response.text)

# With generation config
response = model.generate_content(
    "Write a poem",
    generation_config=genai.types.GenerationConfig(
        temperature=0.9,
        max_output_tokens=200
    )
)

# Streaming
response = model.generate_content("Count to 10", stream=True)
for chunk in response:
    print(chunk.text, end="")

# Function calling
model = genai.GenerativeModel(
    'gemini-2.5-pro',
    tools=[get_weather_function]
)
response = model.generate_content("What's the weather in Paris?")
```

---

**Document Status:** ✅ Complete  
**Last Updated:** 2026-04-03
