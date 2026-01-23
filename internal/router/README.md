# Router Package

The router package configures and initializes the main HTTP router for HelixAgent, wiring together all handlers, middleware, and services.

## Overview

This package provides the central routing configuration for HelixAgent's REST API. It uses the Gin web framework to handle HTTP requests and integrates with various subsystems including authentication, database, caching, and LLM providers.

## Key Components

### RouterContext

```go
type RouterContext struct {
    Engine          *gin.Engine
    protocolManager *services.UnifiedProtocolManager
    oauthMonitor    *services.OAuthTokenMonitor
    healthMonitor   *services.ProviderHealthMonitor
}
```

Wraps the Gin engine with cleanup capabilities for background services.

### Setup Functions

- `SetupRouter(cfg *config.Config) *gin.Engine` - Creates and configures the main HTTP router
- `SetupRouterWithContext(cfg *config.Config) *RouterContext` - Creates router with cleanup support for tests

## Features

- **Middleware Integration**: Logger, Recovery, Feature Flags, CORS, Rate Limiting
- **Database Fallback**: Automatic fallback to in-memory mode if PostgreSQL unavailable
- **Protocol Support**: MCP, ACP, LSP, Embeddings, Vision, Cognee
- **Health Monitoring**: Provider health checks and OAuth token monitoring
- **Prometheus Metrics**: Exposes `/metrics` endpoint

## API Endpoints

### OpenAI-Compatible API
- `POST /v1/chat/completions` - Chat completions
- `POST /v1/completions` - Text completions
- `GET /v1/models` - List available models
- `POST /v1/embeddings` - Generate embeddings

### Authentication
- `POST /v1/auth/login` - User login
- `POST /v1/auth/register` - User registration
- `POST /v1/auth/refresh` - Refresh token
- `POST /v1/auth/logout` - Logout
- `GET /v1/auth/me` - Current user info

### AI Debate System
- `POST /v1/debates` - Start new debate
- `GET /v1/debates/:id` - Get debate status
- `GET /v1/debates/:id/stream` - Stream debate events

### Protocol Endpoints
- `/v1/mcp/*` - Model Context Protocol
- `/v1/lsp/*` - Language Server Protocol
- `/v1/acp/*` - Agent Communication Protocol

## Usage

```go
import "dev.helix.agent/internal/router"

cfg := config.LoadConfig()
ctx := router.SetupRouterWithContext(cfg)
defer ctx.Shutdown()

ctx.Engine.Run(":7061")
```

## Testing

```bash
go test -v ./internal/router/...
```

## Related Packages

- `internal/handlers` - HTTP request handlers
- `internal/middleware` - Request middleware
- `internal/services` - Business logic services
- `internal/config` - Configuration management
