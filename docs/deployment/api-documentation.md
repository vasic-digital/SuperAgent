# Deployment API Documentation

This document covers API-related configuration and endpoints relevant to deploying HelixAgent in production environments.

## Base URL Configuration

Set the base URL via the `PORT` environment variable (default `7061`):

```bash
PORT=7061 ./bin/helixagent
```

All API endpoints are served under `/v1/`. The OpenAI-compatible chat endpoint is `POST /v1/chat/completions`.

## Health and Readiness Endpoints

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/v1/health` | GET | Liveness probe |
| `/v1/monitoring/status` | GET | Full system status |
| `/v1/startup/verification` | GET | Provider verification results |

## Authentication

All API requests require a valid API key passed via the `Authorization: Bearer <key>` header or `x-api-key` header. Configure the accepted keys via `JWT_SECRET` and API key middleware.

## Rate Limiting

Rate limiting is enforced per API key. Defaults can be overridden in `configs/production.yaml`. Responses include `X-RateLimit-Remaining` and `Retry-After` headers when limits are reached.

## Deployment Checklist

1. Set `GIN_MODE=release` for production
2. Configure TLS certificates for HTTPS
3. Enable HTTP/3 (QUIC) transport
4. Verify all required providers pass health checks
5. Confirm database migrations are applied

## Related Documentation

- [General Deployment Guide](./general-deployment-guide.md)
- [Production Deployment](./production-deployment.md)
- [API Reference](../api/api-documentation.md)
