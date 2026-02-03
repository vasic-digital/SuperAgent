# External Integrations

This directory contains documentation for all external service integrations supported by HelixAgent.

## Overview

HelixAgent provides extensive integration capabilities with third-party services, LLM providers, knowledge systems, and external APIs. These integrations enable powerful workflows combining multiple services and data sources.

## Documentation Index

### LLM Provider Integrations

| Document | Description |
|----------|-------------|
| [Multi-Provider Setup](./MULTI_PROVIDER_SETUP.md) | Guide for configuring multiple LLM providers with ensemble support |
| [OpenRouter Integration](./OPENROUTER_INTEGRATION.md) | OpenRouter AI provider with 50+ models and routing strategies |
| [OAuth Credentials Integration](./OAUTH_CREDENTIALS_INTEGRATION.md) | OAuth2 credential support for Claude Code and Qwen Code CLI agents |

### Knowledge Systems

| Document | Description |
|----------|-------------|
| [Cognee Integration](./COGNEE_INTEGRATION.md) | Knowledge graph, vector search, and AI memory engine integration |
| [Cognee Integration Guide](./COGNEE_INTEGRATION_GUIDE.md) | Step-by-step guide for Cognee setup and advanced features |

### Models.dev Integration

| Document | Description |
|----------|-------------|
| [Models.dev Integration Plan](./MODELSDEV_INTEGRATION_PLAN.md) | Comprehensive implementation plan for Models.dev integration |
| [Models.dev Implementation Guide](./MODELSDEV_IMPLEMENTATION_GUIDE.md) | Technical implementation guide |
| [Models.dev Implementation Status](./MODELSDEV_IMPLEMENTATION_STATUS.md) | Current implementation status and progress |
| [Models.dev Summary](./MODELSDEV_SUMMARY.md) | High-level summary of the integration |
| [Models.dev Final Summary](./MODELSDEV_FINAL_SUMMARY.md) | Final summary and outcomes |
| [Models.dev Test Summary](./MODELSDEV_TEST_SUMMARY.md) | Test results and validation |
| [Models.dev Completion Report](./MODELSDEV_COMPLETION_REPORT.md) | Completion report and metrics |

### LLM Utilities

| Document | Description |
|----------|-------------|
| [Request LLM Utils Integrations](./Request_LLM_Utils_Integrations.md) | LLM utility libraries: GPTCache, LangChain, LlamaIndex, Outlines, SGLang |

## Third-Party Service Integrations

### Message Queues

HelixAgent supports integration with message queue systems for event-driven architectures:

- **Kafka**: High-throughput distributed streaming platform
  - Topic-based message routing
  - Consumer group management
  - Exactly-once semantics support

- **RabbitMQ**: Advanced message queuing
  - Exchange and queue management
  - Dead letter handling
  - Priority message support

### API Integrations

- **GraphQL**: Full GraphQL support for flexible API queries
  - Schema introspection
  - Subscription support
  - Batched queries

- **gRPC**: High-performance RPC framework
  - Bidirectional streaming
  - Protocol buffer serialization
  - Load balancing support

## Custom Integration Patterns

### Creating Custom Integrations

HelixAgent provides interfaces for building custom integrations:

```go
// Example: Custom service integration
type ServiceIntegration interface {
    Connect(ctx context.Context, config *Config) error
    Disconnect(ctx context.Context) error
    HealthCheck(ctx context.Context) error
    Execute(ctx context.Context, request *Request) (*Response, error)
}
```

### Integration Best Practices

1. **Connection Management**: Use connection pooling for external services
2. **Error Handling**: Implement circuit breakers and retry logic
3. **Health Checks**: Register health check endpoints for all integrations
4. **Configuration**: Use environment variables for sensitive credentials
5. **Observability**: Emit metrics and traces for all integration calls

### Webhook Support

Configure webhooks for event notifications:

```yaml
webhooks:
  - url: "https://your-service.com/webhook"
    events: ["debate.completed", "consensus.reached"]
    secret: "${WEBHOOK_SECRET}"
    retry_count: 3
```

## Environment Variables

Key environment variables for integrations:

| Variable | Description |
|----------|-------------|
| `COGNEE_BASE_URL` | Cognee service URL |
| `COGNEE_API_KEY` | Cognee API key |
| `OPENROUTER_API_KEY` | OpenRouter API key |
| `CLAUDE_CODE_USE_OAUTH_CREDENTIALS` | Enable Claude OAuth |
| `QWEN_CODE_USE_OAUTH_CREDENTIALS` | Enable Qwen OAuth |
| `KAFKA_BROKERS` | Kafka broker addresses |
| `RABBITMQ_URL` | RabbitMQ connection URL |

## Related Documentation

- [Provider Setup Guide](../deployment/PROVIDER_SETUP.md)
- [Configuration Reference](../configuration/README.md)
- [API Documentation](../api/README.md)
