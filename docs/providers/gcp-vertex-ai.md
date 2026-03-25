# GCP Vertex AI Integration

HelixAgent supports Google Gemini models through both the direct Gemini API and GCP Vertex AI.

## Overview

Vertex AI provides enterprise-grade access to Gemini models with additional features like VPC-SC support, CMEK encryption, and regional endpoints. Use Vertex AI when you need compliance with enterprise security requirements.

## Configuration

```bash
# Vertex AI endpoint
GEMINI_API_KEY=your-api-key
GEMINI_USE_VERTEX=true
GCP_PROJECT_ID=your-project-id
GCP_REGION=us-central1
```

## Authentication

Vertex AI uses Google Cloud IAM for authentication. Ensure your service account has the `aiplatform.endpoints.predict` permission. Application Default Credentials (ADC) are supported.

## Differences from Direct Gemini API

| Feature | Direct API | Vertex AI |
|---------|-----------|-----------|
| Authentication | API key | IAM / Service Account |
| Regional endpoints | No | Yes |
| VPC-SC support | No | Yes |
| Data residency | Limited | Full control |
| Pricing | Consumer | Enterprise |

## Related Documentation

- [Gemini Provider Guide](./gemini.md)
- [Provider Overview](./README.md)
