# Together AI API - Complete Reference
## 100+ Open Source Models

**Provider:** Together AI  
**Base URL:** `https://api.together.xyz/v1`  
**Docs:** https://docs.together.ai  
**Specialty:** Open source model hosting  

---

## Authentication

### Bearer Token
```http
Authorization: Bearer {TOGETHER_API_KEY}
Content-Type: application/json
```

---

## Models (100+)

### Llama Series
| Model | Context | Notes |
|-------|---------|-------|
| meta-llama/Llama-3.3-70B-Instruct-Turbo | 128K | Latest |
| meta-llama/Llama-3.2-90B-Vision-Instruct | 128K | Vision |
| meta-llama/Llama-3.1-405B-Instruct-Turbo | 128K | Largest |
| meta-llama/Llama-3.1-70B-Instruct-Turbo | 128K | Balanced |
| meta-llama/Llama-3.1-8B-Instruct-Turbo | 128K | Fast |

### Mixtral
| Model | Context | Notes |
|-------|---------|-------|
| mistralai/Mixtral-8x22B-Instruct-v0.1 | 64K | MoE |
| mistralai/Mixtral-8x7B-Instruct-v0.1 | 32K | Popular |

### Qwen
| Model | Context | Notes |
|-------|---------|-------|
| Qwen/Qwen2.5-72B-Instruct-Turbo | 128K | Alibaba |
| Qwen/Qwen2.5-7B-Instruct-Turbo | 128K | Fast |

### DeepSeek
| Model | Context | Notes |
|-------|---------|-------|
| deepseek-ai/DeepSeek-V3 | 64K | Reasoning |
| deepseek-ai/DeepSeek-R1 | 64K | Chain-of-thought |

### Specialized
| Model | Type | Notes |
|-------|------|-------|
| Nexusflow/Athene-V2-Chat | Coding | Code generation |
| databricks/dbrx-instruct | General | 132B params |
| google/gemma-2-27b-it | Lightweight | Google |

---

## Endpoints

### Chat Completions

#### POST /v1/chat/completions

**Request:**
```json
{
  "model": "meta-llama/Llama-3.3-70B-Instruct-Turbo",
  "messages": [
    {"role": "system", "content": "You are helpful."},
    {"role": "user", "content": "Hello!"}
  ],
  "max_tokens": 512,
  "temperature": 0.7,
  "top_p": 0.7,
  "top_k": 50,
  "repetition_penalty": 1,
  "stream": false,
  "stop": ["<|eot_id|>"]
}
```

**Response:**
```json
{
  "id": "8f6b37e6-1",
  "choices": [{
    "message": {
      "role": "assistant",
      "content": "Hello! How can I help?"
    }
  }],
  "usage": {
    "prompt_tokens": 20,
    "completion_tokens": 10,
    "total_tokens": 30
  }
}
```

---

## Unique Features

### Quantization Selection
```json
{
  "model": "meta-llama/Llama-3.1-70B-Instruct-Turbo",
  "messages": [...],
  "quantization": "fp16"  // fp16, int8, int4, awq
}
```

### JSON Mode
```json
{
  "model": "meta-llama/Llama-3.3-70B-Instruct-Turbo",
  "messages": [...],
  "response_format": {"type": "json_object"}
}
```

---

## CLI Agent Workarounds

### Model Selection Strategy
```python
# Together has many models - cache available models
@lru_cache(maxsize=1)
def get_available_models():
    return client.models.list()

# Select based on task
def select_model(task_type):
    if task_type == "coding":
        return "Nexusflow/Athene-V2-Chat"
    elif task_type == "reasoning":
        return "deepseek-ai/DeepSeek-V3"
    else:
        return "meta-llama/Llama-3.3-70B-Instruct-Turbo"
```

### Cost Optimization
```python
# Use smaller models for simple tasks
# Quantize to int8/int4 for speed
# Use repetition_penalty to reduce token waste
```

---

**Document Status:** ✅ Complete  
**Last Updated:** 2026-04-03
