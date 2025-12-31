# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

SuperAgent is an AI-powered ensemble LLM service written in Go that combines responses from multiple language models. It provides OpenAI-compatible APIs and supports 6+ LLM providers (Claude, DeepSeek, Gemini, Qwen, ZAI, Ollama, OpenRouter).

## Build Commands

```bash
make build              # Build SuperAgent binary
make build-debug        # Build with debug symbols
make run                # Run locally
make run-dev            # Run in development mode (GIN_MODE=debug)
make docker-build       # Build Docker image
docker-compose up -d    # Start full stack
```

## Testing

```bash
make test                  # Run all tests
make test-coverage         # Tests with HTML coverage report
make test-unit             # Unit tests only (./internal/... -short)
make test-integration      # Integration tests (./tests/integration)
make test-e2e              # End-to-end tests (./tests/e2e)
make test-security         # Security tests (./tests/security)
make test-stress           # Stress tests (./tests/stress)
make test-chaos            # Chaos/challenge tests (./tests/challenge)
make test-bench            # Benchmark tests
make test-race             # Race condition detection
```

Run a single test:
```bash
go test -v -run TestName ./path/to/package
```

## Code Quality

```bash
make fmt              # Format code (go fmt)
make vet              # Static analysis (go vet)
make lint             # Run golangci-lint
make security-scan    # Security scanning (gosec)
make install-deps     # Install dev dependencies (golangci-lint, gosec)
```

## Architecture

### Entry Points
- `cmd/superagent/` - Main SuperAgent application
- `cmd/api/` - API server
- `cmd/grpc-server/` - gRPC server

### Core Packages (`internal/`)
- `llm/` - LLM provider abstractions and ensemble orchestration
  - `providers/` - Individual implementations (claude, deepseek, gemini, ollama, qwen, zai, openrouter)
  - `ensemble.go` - Ensemble orchestration logic
- `services/` - Business logic
  - `provider_registry.go` - Provider management
  - `ensemble.go` - Ensemble service
  - `context_manager.go` - Multi-source context aggregation
  - `mcp_client.go` - Model Context Protocol client
  - `lsp_manager.go` - Language Server Protocol manager
  - `plugin_system.go` - Hot-reloadable plugin architecture
- `handlers/` - HTTP handlers & API endpoints
- `middleware/` - Auth, rate limiting, CORS
- `cache/` - Caching layer (Redis, in-memory)
- `database/` - PostgreSQL connections
- `repository/` - Data access layer
- `models/` - Data models and enums

### Key Interfaces (Extensibility Points)
- `LLMProvider` - Provider implementation contract
- `VotingStrategy` - Ensemble voting strategies
- `PluginRegistry` / `PluginLoader` - Plugin system
- `CacheInterface` - Caching abstraction
- `CloudProvider` - Cloud integration

### Architectural Patterns
- **Provider Registry**: Unified interface for multiple LLM providers
- **Ensemble Strategy**: Confidence-weighted voting, majority vote
- **Plugin System**: Hot-reloadable plugins
- **Circuit Breaker**: Fault tolerance for provider failures
- **Middleware Chain**: Auth, rate limiting, logging pipeline

## Technology Stack

- **Framework**: Gin (v1.11.0)
- **Database**: PostgreSQL 15 with pgx/v5 driver
- **Cache**: Redis 7
- **Protocols**: OpenAI-compatible REST, gRPC, MCP, LSP
- **Testing**: testify (v1.11.1)
- **Monitoring**: Prometheus, Grafana

## Configuration

Environment variables defined in `.env.example`. Key categories:
- Server: `PORT`, `GIN_MODE`, `JWT_SECRET`
- Database: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- Redis: `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`
- LLM providers: `CLAUDE_API_KEY`, `DEEPSEEK_API_KEY`, `GEMINI_API_KEY`, etc.
- Ollama: `OLLAMA_ENABLED`, `OLLAMA_BASE_URL`, `OLLAMA_MODEL`

Configuration files in `/configs`: `development.yaml`, `production.yaml`, `multi-provider.yaml`

## Docker Compose Profiles

```bash
docker-compose up -d                    # Core services (postgres, redis, cognee, chromadb)
docker-compose --profile ai up -d       # Add AI services (ollama)
docker-compose --profile monitoring up -d  # Add monitoring (prometheus, grafana)
docker-compose --profile full up -d     # Everything
```
