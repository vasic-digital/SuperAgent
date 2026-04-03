# xAI API - Complete Reference
## Grok Models

**Provider:** xAI  
**Base URL:** `https://api.x.ai/v1`  
**Docs:** https://docs.x.ai  
**Models:** Grok 2, Grok 3  

---

## Authentication

### Bearer Token
```http
Authorization: Bearer {XAI_API_KEY}
Content-Type: application/json
```

---

## Models

| Model | Context | Notes |
|-------|---------|-------|
| grok-3-latest | 128K | Latest |
| grok-3 | 128K | Standard |
| grok-2-latest | 128K | Previous gen |
| grok-2 | 128K | Standard |

---

## Endpoints

### Chat Completions

#### POST /v1/chat/completions

**Request:**
```json
{
  "model": "grok-3-latest",
  "messages": [
    {"role": "system", "content": "You are Grok, helpful AI."},
    {"role": "user", "content": "Hello!"}
  ],
  "max_tokens": 1024,
  "temperature": 0.7,
  "stream": false
}
```

---

## Unique Features

### X Integration
Grok has real-time access to X (Twitter) data for current events.

---

**Document Status:** ✅ Complete  
**Last Updated:** 2026-04-03
