# Course 17: Cloud Provider Integration

## Course Overview

**Duration:** 45 minutes  
**Level:** Intermediate  
**Prerequisites:** Course 01-Fundamentals, Course 07-Advanced Providers

Learn how to integrate HelixAgent with enterprise cloud LLM providers including AWS Bedrock, Google Cloud Vertex AI, and Azure OpenAI Service.

---

## Learning Objectives

By the end of this course, you will be able to:

1. Configure AWS Bedrock integration
2. Set up GCP Vertex AI connectivity
3. Configure Azure OpenAI Service
4. Implement failover between providers
5. Optimize costs across cloud providers
6. Apply security best practices

---

## Module 1: Introduction to Cloud Providers (5 min)

### Why Cloud Providers?

- Enterprise compliance requirements
- Managed infrastructure
- Scalability and reliability
- Regional data residency

### Supported Providers

| Provider | Service | Best For |
|----------|---------|----------|
| AWS | Bedrock | Enterprise workloads, compliance |
| GCP | Vertex AI | Multimodal, long context |
| Azure | OpenAI | GPT-4 access, Microsoft ecosystem |

---

## Module 2: AWS Bedrock Setup (10 min)

### Prerequisites

- AWS Account with Bedrock access
- IAM user with Bedrock permissions
- Access keys configured

### Configuration

```bash
# .env configuration
AWS_BEDROCK_REGION=us-east-1
AWS_BEDROCK_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
AWS_BEDROCK_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
```

### Available Models

- Claude 3 (Sonnet, Opus, Haiku)
- Llama 3 (8B, 70B)
- Amazon Titan
- Cohere Command
- AI21 Jurassic

### Demo: Invoke Claude via Bedrock

```bash
curl -X POST http://localhost:7061/v1/chat/completions \
  -H "Authorization: Bearer $HELIXAGENT_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "aws-bedrock/anthropic.claude-3-sonnet",
    "messages": [{"role": "user", "content": "Analyze this code for security issues"}]
  }'
```

---

## Module 3: GCP Vertex AI Setup (10 min)

### Prerequisites

- GCP Project with Vertex AI API enabled
- Service account with Vertex AI permissions
- OAuth2 access token

### Configuration

```bash
# .env configuration
GCP_VERTEX_PROJECT_ID=my-project-id
GCP_VERTEX_LOCATION=us-central1
GCP_VERTEX_ACCESS_TOKEN=$(gcloud auth print-access-token)
```

### Available Models

- Gemini 1.5 Pro
- Gemini 1.0 Pro
- PaLM 2 (Text, Chat)
- Codey (Code generation)

### Demo: Multimodal with Gemini

```bash
curl -X POST http://localhost:7061/v1/chat/completions \
  -H "Authorization: Bearer $HELIXAGENT_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gcp-vertex/gemini-1.5-pro",
    "messages": [{"role": "user", "content": "Describe this image", "image": "base64..."}]
  }'
```

---

## Module 4: Azure OpenAI Service Setup (10 min)

### Prerequisites

- Azure subscription
- Azure OpenAI resource deployed
- API key from Azure portal

### Configuration

```bash
# .env configuration
AZURE_OPENAI_ENDPOINT=https://my-resource.openai.azure.com/
AZURE_OPENAI_API_KEY=your-azure-api-key
```

### Available Models

- GPT-4
- GPT-4 Turbo
- GPT-3.5-Turbo
- GPT-3.5-Turbo-16k

### Demo: GPT-4 via Azure

```bash
curl -X POST http://localhost:7061/v1/chat/completions \
  -H "Authorization: Bearer $HELIXAGENT_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "azure-openai/gpt-4",
    "messages": [{"role": "user", "content": "Write a complex SQL query"}]
  }'
```

---

## Module 5: Failover and Cost Optimization (5 min)

### Provider Priority

```yaml
providers:
  primary:
    - openai
    - anthropic
  fallback:
    - aws-bedrock
    - azure-openai
```

### Cost Comparison

| Task | Recommended Provider | Cost per 1K tokens |
|------|---------------------|-------------------|
| Simple chat | GCP Gemini Flash | $0.00019 |
| Code generation | Azure GPT-4 | $0.03 |
| Enterprise compliance | AWS Bedrock Claude | $0.003 |
| Multimodal | GCP Gemini Pro | $0.00125 |

### Auto-Selection Rules

```yaml
routing:
  rules:
    - match: "compliance.required"
      provider: aws-bedrock
    - match: "cost.sensitive"
      provider: gcp-vertex/gemini-flash
    - match: "quality.critical"
      provider: azure-openai/gpt-4
```

---

## Module 6: Security Best Practices (5 min)

### AWS Bedrock

- Use IAM roles, not access keys
- Enable CloudTrail logging
- Implement resource tags
- Configure VPC endpoints

### GCP Vertex AI

- Use service accounts with minimal scopes
- Enable VPC Service Controls
- Configure organization policies
- Use Customer-Managed Encryption Keys

### Azure OpenAI

- Use Managed Identity
- Enable Private Link
- Configure network rules
- Monitor with Azure Monitor

---

## Hands-on Lab

### Exercise 1: Multi-Cloud Configuration

Configure all three cloud providers and test failover:

1. Set up AWS Bedrock credentials
2. Configure GCP Vertex AI
3. Add Azure OpenAI
4. Test automatic failover

### Exercise 2: Cost-Optimized Routing

Create routing rules that optimize for cost:

1. Route simple queries to cheapest provider
2. Route complex queries to best capability
3. Implement budget limits

---

## Quiz

1. Which cloud provider offers Claude models?
2. What authentication method does GCP Vertex AI use?
3. How do you configure provider failover?
4. What is the recommended way to secure AWS credentials?

---

## Resources

- [Cloud Provider Documentation](../CLOUD_PROVIDERS.md)
- [Configuration Guide](../guides/configuration-guide.md)
- [Security Best Practices](../deployment/PRODUCTION_DEPLOYMENT_CHECKLIST.md)

---

**Next Course:** Course 18 - Advanced Security Scanning
