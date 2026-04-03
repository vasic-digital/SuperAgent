# Perplexity API - Complete Reference
## Search-Integrated LLM

**Provider:** Perplexity  
**Base URL:** `https://api.perplexity.ai`  
**Docs:** https://docs.perplexity.ai  
**Specialty:** Real-time web search + LLM  

---

## Authentication

### Bearer Token
```http
Authorization: Bearer {PERPLEXITY_API_KEY}
Content-Type: application/json
```

---

## Rate Limits

| Tier | RPM | Notes |
|------|-----|-------|
| Free | 5 | Limited testing |
| Pro | 100+ | Production use |

---

## Models (Sonar Series)

| Model | Context | Search | Notes |
|-------|---------|--------|-------|
| sonar-pro | 200K | ✅ | Best performance |
| sonar | 128K | ✅ | Balanced |
| sonar-reasoning | 128K | ✅ | Extended thinking |
| sonar-deep-research | 128K | ✅ | Deep research |

---

## Endpoints

### Chat Completions

#### POST /chat/completions

**Request:**
```json
{
  "model": "sonar-pro",
  "messages": [
    {"role": "system", "content": "Be precise and concise."},
    {"role": "user", "content": "What happened yesterday?"}
  ],
  "temperature": 0.7,
  "max_tokens": 2048,
  "search_recency_filter": "day",
  "return_images": false,
  "return_related_questions": true
}
```

**Response:**
```json
{
  "choices": [{
    "message": {
      "content": "Yesterday, ...",
      "role": "assistant"
    }
  }],
  "citations": [
    {"url": "https://...", "title": "..."}
  ]
}
```

---

## Search Parameters

| Parameter | Options | Description |
|-----------|---------|-------------|
| search_recency_filter | day, week, month, year | Time filter |
| search_domain_filter | ["domain.com"] | Domain whitelist |
| return_images | boolean | Include images |
| return_related_questions | boolean | Suggest follow-ups |

---

## CLI Agent Integration

Perplexity is already implemented as search provider in HelixAgent:

```go
// From internal/search/web_search.go
provider := search.NewPerplexityProvider(logger, apiKey)
result, err := provider.Search(ctx, "query", options)
```

---

**Document Status:** ✅ Complete  
**Last Updated:** 2026-04-03
