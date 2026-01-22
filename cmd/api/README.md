# HelixAgent API Server

This package contains the standalone API server entry point.

## Overview

The API server provides the REST API functionality without the full HelixAgent feature set. Use this for lightweight deployments or API-only scenarios.

## Files

- `main.go` - API server entry point
- `setup.go` - Server configuration and route setup

## Features

- OpenAI-compatible completions API
- Model listing and information
- Health check endpoints
- Metrics endpoint

## Usage

### Build and Run

```bash
# Build
go build -o bin/api-server ./cmd/api

# Run
./bin/api-server

# Run with custom port
PORT=8080 ./bin/api-server
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `7061` |
| `API_KEY` | Required API key | None |
| `PROVIDER` | Default LLM provider | `claude` |

## Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/v1/completions` | POST | Create completion |
| `/v1/chat/completions` | POST | Create chat completion |
| `/v1/models` | GET | List available models |
| `/health` | GET | Health check |
| `/metrics` | GET | Prometheus metrics |

## Example

```bash
curl -X POST http://localhost:7061/v1/chat/completions \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-sonnet-4-20250514",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

## Testing

```bash
go test -v ./cmd/api/...
```
