# OpenAPI Specification Generation Guide

## Overview

This document provides instructions for generating and maintaining the OpenAPI specification for HelixAgent's REST API.

## Current Status

**Status**: ⚠️ **To Be Generated**

The OpenAPI specification should be generated from source code using Swagger annotations rather than manually written. This ensures the spec stays synchronized with the actual API implementation.

## Recommended Approach

### 1. Add Swagger Annotations to Code

Install swag:
```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

Add annotations to `internal/handlers/*.go`:
```go
// @title HelixAgent API
// @version 1.0
// @description AI-powered ensemble LLM service with multi-provider support
// @termsOfService http://helixagent.ai/terms/

// @contact.name API Support
// @contact.url http://helixagent.ai/support
// @contact.email support@helixagent.ai

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:7061
// @BasePath /v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// Health godoc
// @Summary Health check endpoint
// @Description Get service health status
// @Tags health
// @Accept  json
// @Produce  json
// @Success 200 {object} models.HealthResponse
// @Router /health [get]
func (h *Handler) Health(c *gin.Context) {
    // Implementation
}
```

### 2. Generate OpenAPI Spec

```bash
# Generate swagger docs
swag init -g cmd/api/main.go -o docs/api/swagger

# This creates:
# - docs/api/swagger/swagger.json
# - docs/api/swagger/swagger.yaml
# - docs/api/swagger/docs.go
```

### 3. Serve Swagger UI

Add to `main.go`:
```go
import (
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
)

// Add route
router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
```

Access at: `http://localhost:7061/docs/index.html`

## Endpoints to Document

### Core API
- `GET /health` - Service health
- `GET /v1/health` - Detailed health
- `GET /v1/models` - Available models
- `GET /v1/providers` - Provider list
- `GET /metrics` - Prometheus metrics

### Completions
- `POST /v1/completions` - Single completion
- `POST /v1/chat/completions` - Chat completions
- `POST /v1/completions/stream` - Streaming
- `POST /v1/ensemble/completions` - Ensemble

### AI Debate
- `POST /v1/debates` - Start debate
- `GET /v1/debates/:id` - Get debate
- `POST /v1/debates/:id/rounds` - Add round

### Protocols
- `POST /v1/mcp` - MCP requests
- `POST /v1/acp` - ACP requests
- `POST /v1/lsp` - LSP requests

### Monitoring
- `GET /v1/monitoring/status` - System status
- `GET /v1/monitoring/circuit-breakers` - Circuit breaker status
- `GET /v1/monitoring/provider-health` - Provider health

### SpecKit
- `GET /v1/speckit/status` - SpecKit status
- `POST /v1/speckit/force-activate` - Force activation
- `POST /v1/speckit/resume` - Resume session

### Constitution Watcher
- `GET /v1/constitution/watcher/health` - Watcher health
- `GET /v1/constitution/watcher/status` - Watcher status
- `GET /v1/constitution/watcher/history` - Change history
- `POST /v1/constitution/watcher/check` - Force check
- `POST /v1/constitution/watcher/sync` - Force sync

## Manual OpenAPI Spec (Temporary)

Until automatic generation is implemented, here's a minimal OpenAPI 3.0 spec:

```yaml
openapi: 3.0.3
info:
  title: HelixAgent API
  description: AI-powered ensemble LLM service
  version: 1.0.0
  contact:
    email: support@helixagent.ai

servers:
  - url: http://localhost:7061
    description: Local development
  - url: https://api.helixagent.ai
    description: Production

paths:
  /health:
    get:
      summary: Health check
      responses:
        '200':
          description: Service is healthy
          content:
            application/json:
              schema:
                type: object
                properties:
                  status:
                    type: string
                    example: healthy

  /v1/completions:
    post:
      summary: Single completion request
      security:
        - BearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                prompt:
                  type: string
                model:
                  type: string
                max_tokens:
                  type: integer
      responses:
        '200':
          description: Completion response
          content:
            application/json:
              schema:
                type: object

components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
```

Save as: `docs/api/openapi.yaml`

## Next Steps

1. **Priority 1**: Add Swagger annotations to all handlers
2. **Priority 2**: Generate OpenAPI spec with `swag init`
3. **Priority 3**: Deploy Swagger UI
4. **Priority 4**: Validate spec with OpenAPI validator
5. **Priority 5**: Generate client SDKs from spec

## Validation

```bash
# Install validator
npm install -g @apidevtools/swagger-cli

# Validate spec
swagger-cli validate docs/api/swagger/swagger.yaml
```

## Client SDK Generation

```bash
# Install OpenAPI Generator
npm install -g @openapitools/openapi-generator-cli

# Generate Go client
openapi-generator-cli generate \
  -i docs/api/swagger/swagger.yaml \
  -g go \
  -o sdk/go

# Generate Python client
openapi-generator-cli generate \
  -i docs/api/swagger/swagger.yaml \
  -g python \
  -o sdk/python
```

---

**Last Updated**: February 10, 2026
**Status**: 📝 Guide Created - Implementation Pending
**Recommended Tool**: swaggo/swag
