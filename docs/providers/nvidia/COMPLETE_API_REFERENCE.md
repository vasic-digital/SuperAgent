# NVIDIA NIM API - Complete Reference
## Optimized Inference Microservices

**Provider:** NVIDIA  
**Base URL:** `https://integrate.api.nvidia.com/v1`  
**Docs:** https://docs.nvidia.com/nim  
**Specialty:** GPU-optimized inference  

---

## Authentication

### Bearer Token
```http
Authorization: Bearer {NVIDIA_API_KEY}
Content-Type: application/json
```

---

## Models

| Model | Notes |
|-------|-------|
| meta/llama-3.3-70b-instruct | Optimized Llama |
| meta/llama-3.1-405b-instruct | Largest Llama |
| microsoft/phi-3-medium-128k-instruct | Phi 3 |
| mistralai/mixtral-8x22b-instruct-v0.1 | Mixtral |
| google/gemma-2-27b-it | Gemma 2 |
| databricks/dbrx-instruct | DBRX |

---

## Endpoints

### Chat Completions

#### POST /v1/chat/completions

**Request:**
```json
{
  "model": "meta/llama-3.3-70b-instruct",
  "messages": [
    {"role": "user", "content": "Hello!"}
  ],
  "max_tokens": 1024
}
```

---

## Unique Features

### GPU Optimization
Models optimized for NVIDIA GPUs (H100, A100, etc.).

### Enterprise Support
Full NVIDIA enterprise support available.

---

**Document Status:** ✅ Complete  
**Last Updated:** 2026-04-03
