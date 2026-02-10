# HelixAgent Documentation

## Overview
HelixAgent is an advanced AI-powered multi-provider LLM orchestration platform with AI debate capabilities, comprehensive monitoring, and enterprise-grade features.

## Documentation Structure

### Getting Started
- [Quick Start Guide](guides/quick-start-guide.md) - Get up and running in 5 minutes
- [Configuration Guide](guides/configuration-guide.md) - Configuration options and examples
- [Best Practices](guides/best-practices-guide.md) - Recommended practices

### API Documentation
- [API Documentation](api/api-documentation.md) - Complete API reference
- [API Reference Examples](api/api-reference-examples.md) - Usage examples
- [OpenAPI Specification](api/openapi.yaml) - Machine-readable API specification

### User Guides
- [Troubleshooting](guides/troubleshooting-guide.md) - Common issues and solutions
- [AI Coding CLI Agents](guides/AI_CODING_CLI_AGENTS_GUIDE.md) - CLI agent usage
- [Operational Guide](guides/OPERATIONAL_GUIDE.md) - Day-to-day operations
- [Analytics Configuration](guides/ANALYTICS_CONFIGURATION_GUIDE.md) - Metrics setup
- [SpecKit User Guide](guides/SPECKIT_USER_GUIDE.md) - 7-phase development flow and auto-activation
- [Constitution Watcher Guide](guides/CONSTITUTION_WATCHER_GUIDE.md) - Auto-update Constitution management

### Deployment
- [Deployment Guide](deployment/DEPLOYMENT_GUIDE.md) - Main deployment guide
- [Production Deployment](deployment/production-deployment.md) - Production setup
- [Deployment Readiness](deployment/DEPLOYMENT_READINESS_REPORT.md) - Pre-deployment checklist
- [Protocol Deployment](deployment/PROTOCOL_DEPLOYMENT_GUIDE.md) - Protocol-specific deployment

### Development
- [Development Status](development/DEVELOPMENT_STATUS.md) - Current development state
- [Testing Strategy](development/DETAILED_TESTING_STRATEGY.md) - Testing approach
- [Implementation Guide Phase 1](development/DETAILED_IMPLEMENTATION_GUIDE_PHASE1.md)
- [Implementation Guide Phase 2](development/DETAILED_IMPLEMENTATION_GUIDE_PHASE2.md)

### Features
- **[Comprehensive Features Reference](FEATURES.md)** - Complete list of all providers, protocols, and capabilities
- [AI Debate Configuration](features/ai-debate-configuration.md) - Debate system setup
- [Advanced Features Summary](features/ADVANCED_FEATURES_SUMMARY.md)
- [AI Agent Orchestration](features/PHASE_8_AI_AGENT_ORCHESTRATION.md)

### Integrations
- [Cognee Integration](integrations/COGNEE_INTEGRATION_GUIDE.md) - Cognee AI setup
- [Multi-Provider Setup](integrations/MULTI_PROVIDER_SETUP.md) - Multiple LLM providers
- [OpenRouter Integration](integrations/OPENROUTER_INTEGRATION.md) - OpenRouter setup
- [ModelsDevAPI Integration](integrations/MODELSDEV_IMPLEMENTATION_GUIDE.md)

### Architecture
- [Architecture Overview](architecture/architecture.md) - System architecture
- [Agents](architecture/AGENTS.md) - AI agent architecture
- [Protocol Support](architecture/PROTOCOL_SUPPORT_DOCUMENTATION.md) - MCP, LSP, ACP protocols

### SDK Documentation
- [Go SDK](sdk/go-sdk.md) - Go client library
- [Python SDK](sdk/python-sdk.md) - Python client library
- [JavaScript SDK](sdk/javascript-sdk.md) - JavaScript/TypeScript client
- [Mobile SDKs](sdk/mobile-sdks.md) - iOS and Android clients

### Tutorials
- [Hello World](tutorials/HELLO_WORLD.md) - First steps tutorial
- [Video Course Content](tutorials/VIDEO_COURSE_CONTENT.md) - Video tutorial scripts

### Reports
- [Test Report](reports/TEST_REPORT.md) - Test results
- [AI Debate Test Report](reports/AI_DEBATE_TEST_REPORT.md) - Debate system tests
- [Comprehensive Audit](reports/COMPREHENSIVE_AUDIT_REPORT.md)

### Testing & Security
- **Security Penetration Tests**: `tests/security/penetration_test.go`
  - Prompt injection (system prompt extraction, role manipulation)
  - Jailbreaking (multi-language attacks, hypothetical scenarios)
  - Data exfiltration (PII extraction, credential probing)
  - Indirect injection (markdown/HTML injection, encoded payloads)
- **AI Debate Challenge Tests**: `tests/challenge/ai_debate_maximal_challenge_test.go`
- **LLM+Cognee Integration Tests**: `tests/integration/llm_cognee_verification_test.go`

### Additional Resources
- [Specifications](specs/) - Project specifications
- [Toolkit Documentation](toolkit/) - Provider implementations
- [Marketing Materials](marketing/) - Launch documentation
- [Archive](archive/) - Historical documentation

## Quick Navigation

| Section | Description |
|---------|-------------|
| **Getting Started** | New user onboarding |
| **API Docs** | API reference and examples |
| **User Guides** | End-user documentation |
| **Deployment** | Production deployment |
| **Development** | Developer resources |
| **Tutorials** | Learning materials |
| **Architecture** | Technical design docs |

## Key Features

- **Multi-Provider Support**: 21 LLM providers including Claude, DeepSeek, Gemini, Mistral, OpenRouter, Qwen, xAI/Grok, Cohere, Perplexity, Groq, and more
- **Embedding Providers**: 13 embedding providers with 40+ models (OpenAI, Cohere, Voyage, Jina, Google, Bedrock, etc.)
- **Protocol Support**: 35 MCP implementations (19 adapters + 16 servers), 10 LSP servers, ACP integration
- **Modular Architecture**: 20 extracted modules (EventBus, Concurrency, Observability, Auth, Storage, Streaming, Security, VectorDB, Embeddings, Database, Cache, Messaging, Formatters, MCP, RAG, Memory, Optimization, Plugins, Containers, Challenges)
- **Dynamic Provider Selection**: Real-time LLMsVerifier scores for optimal provider routing
- **SpecKit Auto-Activation**: Intelligent 7-phase development flow triggered automatically for large changes
- **Constitution Watcher**: Auto-update Constitution on project changes with background monitoring
- **Cognee Integration**: AI Memory Engine with knowledge graphs and semantic search
- **AI Debate System**: Advanced multi-agent debate orchestration (5 positions x 3 LLMs = 15 total)
- **Power Features**: RAG, Memory Management, Semantic Routing, Agentic Workflows, Security Framework, and more
- **Enterprise Monitoring**: Comprehensive metrics and observability with OpenTelemetry
- **BigData Integration**: Infinite context, distributed memory, knowledge graph streaming

> See **[FEATURES.md](FEATURES.md)** for the complete reference with all 24+ power features.
- **Protocol Support**: MCP, LSP, and ACP integration
- **Security Testing**: LLM penetration testing framework with guardrails
- **Extensible**: Plugin architecture for custom integrations

## Quick Links

- [Main README](../README.md) - Project overview
- [CLAUDE.md](../CLAUDE.md) - Claude Code development guide

---

*Last updated: February 10, 2026*
