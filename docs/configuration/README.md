# Configuration Reference

This directory provides configuration documentation for HelixAgent.

## Configuration Sources

HelixAgent loads configuration from multiple sources in order of precedence:

1. **Environment variables** - Highest priority, override all other sources
2. **Configuration files** - `configs/development.yaml` or `configs/production.yaml`
3. **Defaults** - Built-in sensible defaults

## Key Configuration Areas

| Area | Documentation |
|------|---------------|
| Deployment | [Deployment Configuration](../deployment/configuration.md) |
| Services | [Services Configuration](../user/SERVICES_CONFIGURATION.md) |
| LLM Providers | [Provider Guides](../providers/README.md) |
| Monitoring | [Monitoring Setup](../monitoring/README.md) |
| Security | [Authentication](../security/AUTHENTICATION.md) |

## Environment Variables

See `.env.example` in the project root for the complete list of supported environment variables, including provider API keys, database credentials, and service overrides.

## Related Documentation

- [Quick Start Guide](../guides/quick-start-guide.md)
- [Configuration Guide](../guides/configuration-guide.md)
- [Deployment Guide](../deployment/DEPLOYMENT_GUIDE.md)
