# OpenRouter API - Complete Reference
## Universal Model Router

**Provider:** OpenRouter  
**Base URL:** `https://openrouter.ai/api/v1`  
**Docs:** https://openrouter.ai/docs  
**Specialty:** Route to 100+ providers with one API  

---

## Authentication

### Bearer Token
```http
Authorization: Bearer {OPENROUTER_API_KEY}
HTTP-Referer: https://your-site.com
X-Title: Your App Name
```

---

## Models (100+)

Access models from multiple providers:

| Provider | Model |
|----------|-------|
| OpenAI | openai/gpt-4o |
| Anthropic | anthropic/claude-3.5-sonnet |
| Google | google/gemini-2.5-pro |
| Meta | meta-llama/llama-3.3-70b |
| Mistral | mistralai/mistral-large |
| DeepSeek | deepseek/deepseek-chat |
| Qwen | qwen/qwen-2.5-72b |
| + many more | ... |

---

## Endpoints

### Chat Completions

#### POST /api/v1/chat/completions

**Request:**
```json
{
  "model": "openai/gpt-4o",
  "messages": [
    {"role": "user", "content": "Hello!"}
  ],
  "max_tokens": 1024
}
```

---

## Unique Features

### Provider Routing
```json
{
  "model": "openai/gpt-4o",
  "provider": {
    "order": ["OpenAI", "Together"],
    "allow_fallbacks": true
  }
}
```

### Model Fallbacks
Automatic fallback if primary provider fails.

---

## CLI Agent Workarounds

### Universal Client
Use OpenRouter as universal fallback:
```python
# Try primary provider first
try:
    return primary_provider.generate(prompt)
except:
    # Fallback to OpenRouter
    return openrouter.generate(prompt)
```

---

**Document Status:** ✅ Complete  
**Last Updated:** 2026-04-03
