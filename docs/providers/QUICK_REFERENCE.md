# Provider Quick Reference
## All LLM Providers at a Glance

**Last Updated:** 2026-04-03

---

## Tier 1 (Enterprise)

| Provider | Base URL | Best For | Context | Key Feature |
|----------|----------|----------|---------|-------------|
| **OpenAI** | `api.openai.com/v1` | General purpose | 128K-200K | GPT-4o, o1, o3 |
| **Anthropic** | `api.anthropic.com/v1` | Code, reasoning | 200K | Claude 3.7 Sonnet |
| **Google** | `generativelanguage.googleapis.com/v1beta` | Multimodal | 1M | Gemini 2.5 Pro |

## Tier 2 (High Performance)

| Provider | Base URL | Best For | Context | Key Feature |
|----------|----------|----------|---------|-------------|
| **DeepSeek** | `api.deepseek.com/v1` | Code, reasoning | 64K | DeepSeek R1 |
| **Mistral** | `api.mistral.ai/v1` | European | 128K | Codestral |
| **Groq** | `api.groq.com/openai/v1` | Speed | 128K | 100+ tok/sec |
| **Cohere** | `api.cohere.com/v1` | Enterprise | 128K | Command R+ |

## Tier 3 (Specialized)

| Provider | Base URL | Best For | Key Feature |
|----------|----------|----------|-------------|
| **Perplexity** | `api.perplexity.ai` | Search | Real-time web |
| **Together** | `api.together.xyz/v1` | Open source | 100+ models |
| **Fireworks** | `api.fireworks.ai/inference/v1` | Fast inference | Optimized |
| **Cerebras** | `api.cerebras.ai/v1` | Wafer-scale | Fast training |
| **xAI** | `api.x.ai/v1` | Grok | X integration |

---

## Authentication Pattern

```http
Authorization: Bearer {API_KEY}
Content-Type: application/json
```

**Exceptions:**
- Google: `x-goog-api-key: {KEY}`

---

## Request Format

### OpenAI-compatible (Most Common)
```json
{
  "model": "model-name",
  "messages": [
    {"role": "system", "content": "..."},
    {"role": "user", "content": "..."}
  ],
  "temperature": 0.7,
  "max_tokens": 2048
}
```

### Anthropic Format
```json
{
  "model": "claude-3-5-sonnet",
  "max_tokens": 4096,
  "messages": [
    {"role": "user", "content": "..."}
  ]
}
```

### Google Format
```json
{
  "contents": [
    {"role": "user", "parts": [{"text": "..."}]}
  ]
}
```

---

## Streaming Response

### SSE Format
```
data: {"choices": [{"delta": {"content": "Hello"}}]}

data: [DONE]
```

---

## Rate Limit Handling

```python
# Exponential backoff with jitter
delay = (2 ** attempt) + random.uniform(0, 1)
time.sleep(delay)
```

---

## See Full Documentation

- [OpenAI](openai/COMPLETE_API_REFERENCE.md)
- [Anthropic](anthropic/COMPLETE_API_REFERENCE.md)
- [Google](google/COMPLETE_API_REFERENCE.md)
- [DeepSeek](deepseek/COMPLETE_API_REFERENCE.md)
- [Mistral](mistral/COMPLETE_API_REFERENCE.md)
- [Groq](groq/COMPLETE_API_REFERENCE.md)
- [Cohere](cohere/COMPLETE_API_REFERENCE.md)
- [Perplexity](perplexity/COMPLETE_API_REFERENCE.md)

---

**Document Status:** ✅ Complete  
**Total Providers Documented:** 20+
