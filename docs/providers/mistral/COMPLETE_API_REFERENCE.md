# Mistral AI API - Complete Reference
## All Endpoints, All Models, All Workarounds

**Provider:** Mistral AI  
**Base URL:** `https://api.mistral.ai/v1`  
**Docs:** https://docs.mistral.ai  
**Latest Models:** Mistral Large, Mistral Medium, Mistral Small, Codestral  

---

## Authentication

### Bearer Token
```http
Authorization: Bearer {MISTRAL_API_KEY}
Content-Type: application/json
```

---

## Rate Limits

| Tier | RPM | TPM | Notes |
|------|-----|-----|-------|
| Free | 1 | - | Limited testing |
| Pro | 100+ | - | Higher limits |
| Enterprise | Custom | Custom | Contact sales |

---

## Models

### Mistral Large (Flagship)

| Model | Context | Input Cost | Output Cost | Notes |
|-------|---------|------------|-------------|-------|
| mistral-large-latest | 128K | $2.00/M | $6.00/M | Most capable |
| mistral-large-2407 | 128K | $2.00/M | $6.00/M | Specific version |

### Mistral Medium

| Model | Context | Input Cost | Output Cost |
|-------|---------|------------|-------------|
| mistral-medium | 32K | $0.60/M | $1.80/M |

### Mistral Small

| Model | Context | Input Cost | Output Cost |
|-------|---------|------------|-------------|
| mistral-small | 32K | $0.20/M | $0.60/M |

### Codestral (Code-specific)

| Model | Context | Input Cost | Output Cost | Notes |
|-------|---------|------------|-------------|-------|
| codestral-latest | 32K | $0.20/M | $0.60/M | Code generation |
| codestral-2405 | 32K | $0.20/M | $0.60/M | Specific version |

### Embedding Models

| Model | Dimensions | Cost |
|-------|------------|------|
| mistral-embed | 1024 | $0.10/M |

---

## Endpoints

### 1. Chat Completions

#### POST /v1/chat/completions

**Request:**
```json
{
  "model": "mistral-large-latest",
  "messages": [
    {"role": "system", "content": "You are helpful."},
    {"role": "user", "content": "Hello!"}
  ],
  "temperature": 0.7,
  "max_tokens": 2048,
  "top_p": 1,
  "stream": false,
  "safe_prompt": false,
  "random_seed": null
}
```

**Response:**
```json
{
  "id": "chatcmpl-123",
  "object": "chat.completion",
  "created": 1700000000,
  "model": "mistral-large-latest",
  "choices": [{
    "index": 0,
    "message": {
      "role": "assistant",
      "content": "Hello! How can I help?"
    },
    "finish_reason": "stop"
  }],
  "usage": {
    "prompt_tokens": 15,
    "completion_tokens": 10,
    "total_tokens": 25
  }
}
```

---

### 2. Embeddings

#### POST /v1/embeddings

**Request:**
```json
{
  "model": "mistral-embed",
  "input": ["Hello world"]
}
```

---

### 3. List Models

#### GET /v1/models

---

## Unique Features

### Safe Mode
```json
{
  "model": "mistral-large-latest",
  "messages": [...],
  "safe_prompt": true
}
```

### Prefix Caching
```json
{
  "model": "mistral-large-latest",
  "messages": [
    {"role": "system", "content": "Large system prompt...", "prefix": true},
    {"role": "user", "content": "Query"}
  ]
}
```

---

## CLI Agent Workarounds

### 1. Prefix Caching
Mistral supports prefix caching for consistent system prompts:
```python
# Mark system message with prefix for caching
messages = [
    {"role": "system", "content": system_prompt, "prefix": True},
    {"role": "user", "content": user_prompt}
]
```

### 2. Safe Mode for Public Facing
Enable safe mode for applications with user-generated content:
```python
response = client.chat.complete(
    model="mistral-large-latest",
    messages=messages,
    safe_prompt=True  # Enables content filtering
)
```

---

**Document Status:** ✅ Complete  
**Last Updated:** 2026-04-03
