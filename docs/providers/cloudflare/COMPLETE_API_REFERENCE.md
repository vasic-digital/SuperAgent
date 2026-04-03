# Cloudflare Workers AI - Complete Reference
## Edge Inference

**Provider:** Cloudflare  
**Base URL:** `https://api.cloudflare.com/client/v4/accounts/{account_id}/ai/run`  
**Docs:** https://developers.cloudflare.com/workers-ai  
**Specialty:** Edge deployment, low latency  

---

## Authentication

### API Token
```http
Authorization: Bearer {CF_API_TOKEN}
Content-Type: application/json
```

---

## Models

| Model | Type | Notes |
|-------|------|-------|
| @cf/meta/llama-3.3-70b-instruct | Chat | Meta Llama |
| @cf/meta/llama-3.1-8b-instruct | Chat | Fast |
| @cf/mistral/mistral-7b-instruct | Chat | Mistral |
| @cf/baai/bge-base-en-v1.5 | Embedding | BGE |

---

## Endpoints

### Text Generation

#### POST /accounts/{account_id}/ai/run/@cf/meta/llama-3.3-70b-instruct

**Request:**
```json
{
  "messages": [
    {"role": "system", "content": "You are helpful."},
    {"role": "user", "content": "Hello!"}
  ],
  "max_tokens": 512
}
```

---

## Unique Features

### Edge Deployment
Runs on Cloudflare's edge network (300+ cities).

---

**Document Status:** ✅ Complete  
**Last Updated:** 2026-04-03
