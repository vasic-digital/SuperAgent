# HelixAgent User Manuals

Comprehensive user documentation for HelixAgent.

## Table of Contents

1. [Getting Started](01_GETTING_STARTED.md)
   - What is HelixAgent?
   - System Requirements
   - Quick Installation
   - Configuration
   - Your First API Call

2. [API Reference](02_API_REFERENCE.md)
   - Authentication
   - Chat Completions
   - Streaming
   - Embeddings
   - AI Debate
   - MCP Protocol
   - Error Handling

3. [Provider Configuration](03_PROVIDER_CONFIG.md)
   - Supported Providers (10 LLM providers)
   - Configuration Methods
   - Provider Verification
   - AI Debate Team (25 LLMs: 5 primary + 20 fallback)
   - Health Monitoring
   - Cost Management

4. [Advanced Features](04_ADVANCED_FEATURES.md)
   - AI Debate Ensemble
   - Model Context Protocol
   - Caching System
   - Background Tasks
   - Knowledge Graph
   - Plugin System
   - Performance Tuning
   - Challenge System (45 challenges, 100% pass rate)

## Quick Links

- [Main Documentation](../README.md)
- [Testing Guide](../testing/README.md)
- [API Specification](../../api/openapi.yaml)
- [Troubleshooting](../TROUBLESHOOTING.md)
- [Developer Guide](../DEVELOPER_GUIDE.md)
- [Comprehensive Report](../COMPREHENSIVE_COMPLETION_REPORT.md)

## Latest Updates (January 2026)

### Challenge System - 45 Challenges, 100% Pass Rate
| Challenge | Tests | Status | Description |
|-----------|-------|--------|-------------|
| RAGS | 147/147 | PASSED | RAG integration across 20+ CLI agents |
| MCPS | Full | PASSED | MCP Server integration (22 servers) |
| SKILLS | Full | PASSED | Skills integration (21 categories) |
| +42 more | All | PASSED | Providers, protocols, security, etc. |

### Key Features
- **MCP Tool Search**: New API for intelligent tool discovery (`/v1/mcp/tools/search`)
- **Adapter Search**: MCP adapter discovery (`/v1/mcp/adapters/search`)
- **Tool Suggestions**: Context-aware tool recommendations
- **Strict Real-Result Validation**: FALSE SUCCESS detection prevents empty responses

### RAG Systems Validated
- Cognee (Knowledge Graph + Memory)
- Qdrant (Vector Database)
- RAG Pipeline (Hybrid Search, Reranking, HyDE)
- Embeddings Service

### Bug Fixes
- ProviderHealthMonitor mutex deadlock resolved
- CogneeService ListDatasets JSON parsing fixed
- RAGS timeout increased from 30s to 60s

## Version

This documentation covers HelixAgent v1.0.0+.

## Contributing

To contribute to the documentation, please see our [Contributing Guide](../../CONTRIBUTING.md).
