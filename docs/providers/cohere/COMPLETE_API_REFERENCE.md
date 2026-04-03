# Cohere API - Complete Reference
## Command Models & Enterprise Features

**Provider:** Cohere  
**Base URL:** `https://api.cohere.com/v1`  
**Docs:** https://docs.cohere.com  
**Latest Models:** Command R+, Command R, Embed  

---

## Authentication

### Bearer Token
```http
Authorization: Bearer {COHERE_API_KEY}
Content-Type: application/json
```

---

## Rate Limits

| Tier | RPM | TPM | Notes |
|------|-----|-----|-------|
| Trial | 20 | 100,000 | Limited testing |
| Production | 1,000+ | 10,000,000+ | Paid tier |

---

## Models

### Command Series

| Model | Context | Input Cost | Output Cost | Notes |
|-------|---------|------------|-------------|-------|
| command-r-plus | 128K | $2.50/M | $10.00/M | Most capable |
| command-r | 128K | $0.15/M | $0.60/M | Balanced |
| command | 4K | $1.00/M | $2.00/M | Legacy |
| command-nightly | 4K | $1.00/M | $2.00/M | Latest updates |

### Embedding

| Model | Dimensions | Cost |
|-------|------------|------|
| embed-english-v3.0 | 1024 | $0.10/M |
| embed-multilingual-v3.0 | 1024 | $0.10/M |

---

## Endpoints

### 1. Chat

#### POST /v1/chat

**Request:**
```json
{
  "model": "command-r-plus",
  "message": "Hello!",
  "preamble": "You are a helpful assistant.",
  "temperature": 0.7,
  "max_tokens": 2048,
  "stream": false,
  "chat_history": [],
  "documents": [],
  "tools": []
}
```

**Response:**
```json
{
  "text": "Hello! How can I help you?",
  "generation_id": "gen-123",
  "finish_reason": "COMPLETE",
  "tool_calls": [],
  "documents": []
}
```

---

### 2. Generate (Legacy)

#### POST /v1/generate

---

### 3. Embed

#### POST /v1/embed

---

## Unique Features

### RAG (Retrieval Augmented Generation)
```json
{
  "model": "command-r-plus",
  "message": "What is in the documents?",
  "documents": [
    {
      "id": "doc1",
      "title": "Document 1",
      "text": "Document content..."
    }
  ],
  "prompt_truncation": "AUTO"
}
```

### Connectors
```json
{
  "model": "command-r-plus",
  "message": "Search for information",
  "connectors": [
    {"id": "web-search"}
  ]
}
```

---

## CLI Agent Workarounds

### Document Pre-loading
```python
# Pre-load documents for RAG
documents = [
    {"id": "1", "title": "README", "text": readme_content},
    {"id": "2", "title": "API", "text": api_content}
]

response = co.chat(
    model="command-r-plus",
    message=user_query,
    documents=documents
)
```

---

**Document Status:** ✅ Complete  
**Last Updated:** 2026-04-03
