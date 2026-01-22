# HelixAgent Main Entry Point

This package contains the main entry point for the HelixAgent server application.

## Overview

The HelixAgent server provides:
- OpenAI-compatible REST API
- AI Debate System with 10 LLM providers
- MCP/LSP/ACP protocol support
- Background task execution
- Real-time notifications (SSE/WebSocket)

## Files

- `main.go` - Application entry point and server startup
- `init.go` - Initialization and dependency injection

## Startup Sequence

1. **Load Configuration**
   - Read environment variables
   - Load configuration files
   - Validate settings

2. **Initialize Infrastructure**
   - Connect to PostgreSQL
   - Connect to Redis
   - Initialize HTTP client pool

3. **Startup Verification Pipeline**
   - Discover all LLM providers
   - Verify provider credentials
   - Score providers using LLMsVerifier
   - Select AI Debate Team (15 LLMs)

4. **Initialize Services**
   - Provider Registry
   - Ensemble Service
   - Debate Service
   - Background Task Queue
   - Notification Hub

5. **Start Server**
   - Configure routes
   - Start HTTP server
   - Begin health monitoring

## Usage

### Build and Run

```bash
# Build
make build

# Run with default config
./bin/helixagent

# Run with custom config
./bin/helixagent --config configs/production.yaml

# Run in debug mode
GIN_MODE=debug ./bin/helixagent
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `7061` |
| `GIN_MODE` | Gin mode (debug/release) | `release` |
| `JWT_SECRET` | JWT signing secret | Required |
| `DB_HOST` | PostgreSQL host | `localhost` |
| `REDIS_HOST` | Redis host | `localhost` |
| `CLAUDE_API_KEY` | Claude API key | Optional |
| `DEEPSEEK_API_KEY` | DeepSeek API key | Optional |
| `GEMINI_API_KEY` | Gemini API key | Optional |

### Command Line Flags

```bash
./bin/helixagent --help

Flags:
  --config string    Configuration file path
  --port int         Server port (overrides config)
  --debug            Enable debug mode
  --version          Show version information
```

## Health Checks

```bash
# Liveness probe
curl http://localhost:7061/healthz/live

# Readiness probe
curl http://localhost:7061/healthz/ready

# Full health status
curl http://localhost:7061/health
```

## Graceful Shutdown

The server handles SIGTERM and SIGINT for graceful shutdown:
1. Stop accepting new requests
2. Wait for active requests to complete
3. Close database connections
4. Flush pending notifications
5. Exit cleanly

## Testing

```bash
go test -v ./cmd/helixagent/...
```
