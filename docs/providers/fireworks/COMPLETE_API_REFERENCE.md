# Fireworks AI API - Complete Reference
## Fast Inference Platform

**Provider:** Fireworks AI  
**Base URL:** `https://api.fireworks.ai/inference/v1`  
**Docs:** https://docs.fireworks.ai  
**Specialty:** Optimized inference, fine-tuning  

---

## Authentication

### Bearer Token
```http
Authorization: Bearer {FIREWORKS_API_KEY}
Content-Type: application/json
```

---

## Models

### Fireworks Hosted
| Model | ID | Context |
|-------|-----|---------|
| Llama 3.3 70B | accounts/fireworks/models/llama-v3p3-70b-instruct | 128K |
| Llama 3.1 405B | accounts/fireworks/models/llama-v3p1-405b-instruct | 128K |
| Mixtral 8x22B | accounts/fireworks/models/mixtral-8x22b-instruct | 64K |
| Qwen 2.5 72B | accounts/fireworks/models/qwen2p5-72b-instruct | 128K |
| DeepSeek V3 | accounts/fireworks/models/deepseek-v3 | 64K |

### Custom Fine-tuned
Access your fine-tuned models:
```
accounts/{account-id}/models/{model-id}
```

---

## Endpoints

### Chat Completions

#### POST /inference/v1/chat/completions

**Request:**
```json
{
  "model": "accounts/fireworks/models/llama-v3p3-70b-instruct",
  "messages": [
    {"role": "system", "content": "You are helpful."},
    {"role": "user", "content": "Hello!"}
  ],
  "max_tokens": 512,
  "temperature": 0.7,
  "top_p": 1.0,
  "top_k": 50,
  "stream": false
}
```

---

## Unique Features

### Fine-tuning API
```bash
curl -X POST https://api.fireworks.ai/v1/fine-tuning/jobs \
  -H "Authorization: Bearer $KEY" \
  -d '{
    "model": "llama-v3p1-8b-instruct",
    "training_dataset": "dataset-id",
    "validation_dataset": "val-dataset-id"
  }'
```

---

**Document Status:** ✅ Complete  
**Last Updated:** 2026-04-03
