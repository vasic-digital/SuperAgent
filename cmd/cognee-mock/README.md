# Cognee Mock Server

This package contains a mock Cognee server for testing and development.

## Overview

The Cognee mock server simulates the Cognee API for:
- Local development without Cognee credentials
- Integration testing
- CI/CD pipelines

## Files

- `main.go` - Mock server entry point
- `handler.go` - API endpoint handlers

## Simulated Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/auth` | POST | Mock authentication |
| `/api/v1/add` | POST | Simulate document addition |
| `/api/v1/search` | POST | Simulate knowledge search |
| `/api/v1/cognify` | POST | Simulate cognification |
| `/health` | GET | Health check |

## Usage

### Build and Run

```bash
# Build
go build -o bin/cognee-mock ./cmd/cognee-mock

# Run
./bin/cognee-mock

# Run with custom port
PORT=8888 ./bin/cognee-mock
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Mock server port | `8888` |
| `RESPONSE_DELAY` | Simulated latency | `100ms` |
| `FAILURE_RATE` | Random failure rate | `0` |

## Mock Responses

### Authentication

```bash
curl -X POST http://localhost:8888/api/v1/auth \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "username=admin@helixagent.ai&password=HelixAgentPass123"

# Response
{
    "access_token": "mock-token-xxx",
    "token_type": "Bearer",
    "expires_in": 3600
}
```

### Search

```bash
curl -X POST http://localhost:8888/api/v1/search \
  -H "Authorization: Bearer mock-token-xxx" \
  -d '{"query": "test query"}'

# Response
{
    "results": [
        {"content": "Mock result 1", "score": 0.95},
        {"content": "Mock result 2", "score": 0.87}
    ]
}
```

## Testing Configuration

Use the mock server in tests:

```yaml
cognee:
  base_url: "http://localhost:8888"
  auth_email: "admin@helixagent.ai"
  auth_password: "HelixAgentPass123"
```

## Testing

```bash
go test -v ./cmd/cognee-mock/...
```
