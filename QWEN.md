# HelixAgent: AI-Powered Ensemble LLM Service

## Project Overview

HelixAgent is a production-ready, AI-powered ensemble LLM service that intelligently combines responses from multiple language models to provide the most accurate and reliable outputs. The project is built in Go and provides a comprehensive platform for managing multiple LLM providers with advanced features like Cognee knowledge graph integration, multi-modal processing, and auto-containerization.

### Key Technologies
- **Backend**: Go 1.24.0+ with Gin web framework
- **Database**: PostgreSQL with pgx driver
- **Cache**: Redis for high-performance caching
- **Containerization**: Docker and Docker Compose
- **Monitoring**: Prometheus and Grafana
- **Knowledge Graph**: Cognee for advanced AI memory
- **API Gateway**: OpenAI-compatible API interface

### Architecture
The system follows a microservices architecture with:
- Main HelixAgent service handling API requests
- PostgreSQL database for persistent storage
- Redis for caching and session management
- Cognee knowledge graph for advanced AI memory
- Multiple LLM providers (Claude, DeepSeek, Gemini, Qwen, ZAI, Ollama)
- Monitoring stack with Prometheus and Grafana

## Building and Running

### Prerequisites
- Docker & Docker Compose
- Go 1.21+ (for local development)
- Git

### Using Docker (Recommended)
```bash
# Clone the repository
git clone https://dev.helix.agent.git
cd helixagent

# Copy environment configuration
cp .env.example .env

# Start all services
make docker-full

# Or start specific profiles
make docker-ai          # AI services only
make docker-monitoring   # Monitoring stack only
```

### Local Development
```bash
# Install dependencies
make install-deps

# Setup development environment
make setup-dev

# Run locally
make run-dev
```

### Build Commands
- `make build` - Build HelixAgent binary
- `make build-debug` - Build HelixAgent binary (debug mode)
- `make build-all` - Build for all architectures
- `make run` - Run HelixAgent locally
- `make run-dev` - Run HelixAgent in development mode

## Development Conventions

### Code Quality
- Code formatting: `make fmt` (uses go fmt)
- Static analysis: `make vet` and `make lint`
- Security scanning: `make security-scan`

### Testing
- Unit tests: `make test-unit`
- Integration tests: `make test-integration`
- Full test suite: `make test-with-infra`
- Test coverage: `make test-coverage`
- Benchmark tests: `make test-bench`
- Race condition tests: `make test-race`

### Configuration
The application uses comprehensive environment-based configuration with the following key variables:
- Server configuration (PORT, HELIXAGENT_API_KEY, GIN_MODE)
- Database configuration (DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME)
- Redis configuration (REDIS_HOST, REDIS_PORT, REDIS_PASSWORD)
- LLM provider API keys (CLAUDE_API_KEY, DEEPSEEK_API_KEY, GEMINI_API_KEY, etc.)

### API Endpoints
The service provides an OpenAI-compatible API with the following key endpoints:
- `/health` - Service health status
- `/v1/health` - Detailed health with provider status
- `/v1/models` - Available LLM models
- `/v1/providers` - Configured providers
- `/v1/completions` - Single completion request
- `/v1/chat/completions` - Chat-style completions
- `/v1/ensemble/completions` - Ensemble completions
- `/v1/cognee` - Cognee knowledge graph integration
- `/v1/protocols` - Protocol management
- `/v1/mcp` - Model Context Protocol endpoints

### Security Features
- JWT token authentication
- API key authentication
- Rate limiting with configurable limits
- CORS support with configurable origins
- Input sanitization and request validation
- Secure error handling without information leakage

### Monitoring and Observability
- Prometheus metrics at `/metrics` endpoint
- Comprehensive health checks
- Structured logging with log levels
- Performance monitoring with response time tracking

### Plugin System
The application supports a hot-reloading plugin system with standardized interfaces for extending functionality.

## Project Structure
- `cmd/` - Application entry points
- `internal/` - Internal packages with core functionality
  - `config/` - Configuration management
  - `router/` - HTTP routing and middleware
  - `handlers/` - API request handlers
  - `services/` - Business logic implementations
  - `models/` - Data models and structures
  - `database/` - Database operations
  - `cache/` - Caching mechanisms
  - `llm/` - LLM provider integrations
  - `cognee/` - Cognee knowledge graph integration
  - `optimization/` - LLM optimization features
- `pkg/` - Public packages
- `tests/` - Test suites
- `docker-compose.yml` - Multi-service Docker configuration
- `Dockerfile` - Production Docker image
- `Makefile` - Build and development automation