# AI21 Labs API - Complete Reference
## Jurassic & Jamba Models

**Provider:** AI21 Labs  
**Base URL:** `https://api.ai21.com/studio/v1`  
**Docs:** https://docs.ai21.com  

---

## Authentication

### Bearer Token
```http
Authorization: Bearer {AI21_API_KEY}
Content-Type: application/json
```

---

## Models

### Jamba (SSM-Transformer)
| Model | Context | Notes |
|-------|---------|-------|
| jamba-1.5-large | 256K | Hybrid SSM-Transformer |
| jamba-1.5-mini | 256K | Fast |

---

## Endpoints

### Chat Completions

#### POST /studio/v1/chat/completions

**Request:**
```json
{
  "model": "jamba-1.5-large",
  "messages": [
    {"role": "system", "content": "You are helpful."},
    {"role": "user", "content": "Hello!"}
  ],
  "max_tokens": 1024,
  "temperature": 0.7
}
```

---

## Unique Features

### 256K Context
Jamba models support 256K context window.

### SSM-Transformer Hybrid
Combines State Space Models with Transformers for efficiency.

---

**Document Status:** ✅ Complete  
**Last Updated:** 2026-04-03
