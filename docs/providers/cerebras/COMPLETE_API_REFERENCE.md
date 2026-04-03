# Cerebras API - Complete Reference
## Wafer-Scale Inference

**Provider:** Cerebras  
**Base URL:** `https://api.cerebras.ai/v1`  
**Docs:** https://inference-docs.cerebras.ai  
**Specialty:** Wafer-scale hardware, instant inference  

---

## Authentication

### Bearer Token
```http
Authorization: Bearer {CEREBRAS_API_KEY}
Content-Type: application/json
```

---

## Models

| Model | Context | Speed | Notes |
|-------|---------|-------|-------|
| llama3.1-70b | 128K | 450+ tok/sec | Fastest 70B |
| llama3.1-8b | 128K | 1800+ tok/sec | Ultra-fast |
| llama-3.3-70b | 128K | 450+ tok/sec | Latest |

---

## Endpoints

### Chat Completions

#### POST /v1/chat/completions

**Request:**
```json
{
  "model": "llama3.1-70b",
  "messages": [
    {"role": "system", "content": "You are helpful."},
    {"role": "user", "content": "Hello!"}
  ],
  "max_tokens": 1024,
  "temperature": 0.7,
  "stream": true
}
```

---

## CLI Agent Workarounds

### Speed Optimization
Cerebras is all about speed - optimize for it:
```python
# Always stream for real-time feedback
# Use larger batch sizes
# Minimize prompt engineering overhead
```

---

**Document Status:** ✅ Complete  
**Last Updated:** 2026-04-03
