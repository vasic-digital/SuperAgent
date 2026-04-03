# Azure OpenAI API - Complete Reference
## Enterprise OpenAI

**Provider:** Microsoft Azure  
**Base URL:** `https://{resource-name}.openai.azure.com/openai/deployments/{deployment-id}`  
**Docs:** https://learn.microsoft.com/azure/ai-services/openai  
**Specialty:** Enterprise security, compliance  

---

## Authentication

### API Key
```http
api-key: {AZURE_OPENAI_API_KEY}
Content-Type: application/json
```

### Entra ID (OAuth)
```http
Authorization: Bearer {ACCESS_TOKEN}
```

---

## Models

| Model | Version | Context | Notes |
|-------|---------|---------|-------|
| gpt-4o | 2024-08-06 | 128K | Latest |
| gpt-4o-mini | 2024-07-18 | 128K | Cost-effective |
| gpt-4 | turbo-2024-04-09 | 128K | Stable |
| gpt-35-turbo | 0125 | 16K | Legacy |
| text-embedding-3-large | 1 | 8K | Embeddings |
| text-embedding-3-small | 1 | 8K | Small embeddings |
| dall-e-3 | 3.0 | - | Images |
| whisper | 001 | - | Speech |

---

## Endpoints

### Chat Completions

#### POST /openai/deployments/{deployment}/chat/completions?api-version=2024-06-01

**Request:**
```json
{
  "messages": [
    {"role": "system", "content": "You are helpful."},
    {"role": "user", "content": "Hello!"}
  ],
  "max_tokens": 800,
  "temperature": 0.7,
  "stream": false
}
```

**Note:** Model is specified in the deployment, not the request body.

---

## Unique Features

### Content Filtering
Azure applies content filtering by default. Configure in Azure Portal.

### Private Endpoints
Deploy to private VNET for enterprise security.

### Data Residency
Choose region for data residency compliance.

---

## CLI Agent Workarounds

### Deployment Mapping
```python
# Map models to Azure deployments
AZURE_DEPLOYMENTS = {
    "gpt-4o": "gpt-4o-deployment",
    "gpt-4o-mini": "gpt-4o-mini-deployment"
}

def get_deployment(model):
    return AZURE_DEPLOYMENTS.get(model, model)
```

---

**Document Status:** ✅ Complete  
**Last Updated:** 2026-04-03
