# Replicate API - Complete Reference
## Model Hosting Platform

**Provider:** Replicate  
**Base URL:** `https://api.replicate.com/v1`  
**Docs:** https://replicate.com/docs  
**Specialty:** Run open-source models in the cloud  

---

## Authentication

### Bearer Token
```http
Authorization: Token {REPLICATE_API_TOKEN}
Content-Type: application/json
```

---

## Models

Thousands of open-source models available:

| Model | Owner | Type |
|-------|-------|------|
| llama-3.3-70b | meta | LLM |
| mistral-7b | mistralai | LLM |
| stable-diffusion-xl | stability-ai | Image |
| whisper | openai | Audio |

---

## Endpoints

### Run Model

#### POST /v1/models/{owner}/{name}/predictions

**Request:**
```json
{
  "input": {
    "prompt": "Hello!",
    "max_tokens": 512
  }
}
```

**Response:**
```json
{
  "id": "xyz123",
  "status": "starting",
  "urls": {
    "get": "https://api.replicate.com/v1/predictions/xyz123"
  }
}
```

### Get Prediction

#### GET /v1/predictions/{id}

---

## Unique Features

### Async by Default
All predictions are async. Poll for results or use webhooks.

### Community Models
Access thousands of community-uploaded models.

---

**Document Status:** ✅ Complete  
**Last Updated:** 2026-04-03
