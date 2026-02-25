# Cloud Provider Integration

HelixAgent supports enterprise cloud LLM providers for organizations that require managed, scalable AI infrastructure.

---

## Overview

Cloud provider integration enables HelixAgent to use enterprise-grade LLM services from AWS, Google Cloud, and Microsoft Azure.

| Provider | Service | Models | Auth Method |
|----------|---------|--------|-------------|
| AWS | Bedrock | Claude, Llama, Titan | IAM credentials |
| Google Cloud | Vertex AI | Gemini, PaLM, Llama | OAuth2 / Service Account |
| Microsoft Azure | OpenAI Service | GPT-4, GPT-3.5 | API Key |

---

## AWS Bedrock

### Configuration

```env
# AWS Bedrock Configuration
AWS_BEDROCK_REGION=us-east-1
AWS_BEDROCK_ACCESS_KEY_ID=your-access-key
AWS_BEDROCK_SECRET_ACCESS_KEY=your-secret-key
AWS_BEDROCK_SESSION_TOKEN=your-session-token  # Optional
AWS_BEDROCK_TIMEOUT=60s
```

### Supported Models

- `anthropic.claude-3-sonnet-20240229-v1:0`
- `anthropic.claude-3-opus-20240229-v1:0`
- `meta.llama3-70b-instruct-v1:0`
- `amazon.titan-text-express-v1`

### Usage

```bash
curl -X POST http://localhost:7061/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-helixagent-key" \
  -d '{
    "model": "aws-bedrock/anthropic.claude-3-sonnet",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

---

## GCP Vertex AI

### Configuration

```env
# GCP Vertex AI Configuration
GCP_VERTEX_PROJECT_ID=your-project-id
GCP_VERTEX_LOCATION=us-central1
GCP_VERTEX_ACCESS_TOKEN=your-access-token
GCP_VERTEX_TIMEOUT=60s
```

### Supported Models

- `gemini-1.5-pro`
- `gemini-1.0-pro`
- `text-bison`
- `code-bison`

### Usage

```bash
curl -X POST http://localhost:7061/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-helixagent-key" \
  -d '{
    "model": "gcp-vertex/gemini-1.5-pro",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

---

## Azure OpenAI Service

### Configuration

```env
# Azure OpenAI Configuration
AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com/
AZURE_OPENAI_API_KEY=your-api-key
AZURE_OPENAI_TIMEOUT=60s
```

### Supported Models

- `gpt-4`
- `gpt-4-turbo`
- `gpt-35-turbo`
- `gpt-35-turbo-16k`

### Usage

```bash
curl -X POST http://localhost:7061/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-helixagent-key" \
  -d '{
    "model": "azure-openai/gpt-4",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

---

## Model Selection Priority

HelixAgent uses the following priority for cloud provider models:

1. **Explicit Selection** - User specifies provider in model name
2. **Score-Based** - Providers ranked by LLMsVerifier scores
3. **Cost Optimization** - Cheaper providers preferred for non-critical requests
4. **Availability** - Fallback to alternative providers on failure

---

## Cost Considerations

| Provider | Average Cost (1K tokens) | Best For |
|----------|-------------------------|----------|
| AWS Bedrock | $0.003 - $0.015 | Enterprise, compliance |
| GCP Vertex AI | $0.001 - $0.012 | Multimodal, long context |
| Azure OpenAI | $0.002 - $0.06 | GPT-4 access |

---

## Security Best Practices

### AWS Bedrock
- Use IAM roles instead of access keys when possible
- Enable CloudTrail for audit logging
- Implement least-privilege policies

### GCP Vertex AI
- Use service accounts with minimal permissions
- Enable VPC Service Controls
- Configure organization policies

### Azure OpenAI
- Use Managed Identity for Azure resources
- Enable Azure Monitor logging
- Implement network isolation

---

## Failover Configuration

Cloud providers can be configured as fallbacks:

```yaml
providers:
  primary:
    - openai
    - anthropic
  fallback:
    - aws-bedrock
    - azure-openai
```

---

## Related Documentation

- [Provider Configuration](../guides/configuration-guide.md)
- [API Reference](./api-documentation.md)
- [Security Hardening](../deployment/PRODUCTION_DEPLOYMENT_CHECKLIST.md)
