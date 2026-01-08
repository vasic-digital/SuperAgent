# HelixAgent Getting Started Guide

## Introduction

HelixAgent is an enterprise-grade AI-powered ensemble LLM service that aggregates responses from multiple language model providers using intelligent orchestration strategies. Built with Go 1.23+ and designed for production environments, HelixAgent provides OpenAI-compatible APIs while offering advanced features like AI debates, knowledge graph integration, and multi-provider ensemble capabilities.

This comprehensive guide will walk you through the installation, configuration, and your first API requests with HelixAgent.

---

## Table of Contents

1. [System Requirements](#system-requirements)
2. [Installation Methods](#installation-methods)
3. [Initial Configuration](#initial-configuration)
4. [Starting HelixAgent](#starting-helixagent)
5. [Your First API Request](#your-first-api-request)
6. [Basic Usage Examples](#basic-usage-examples)
7. [Understanding the Architecture](#understanding-the-architecture)
8. [Next Steps](#next-steps)
9. [Troubleshooting](#troubleshooting)

---

## System Requirements

### Minimum Requirements

Before installing HelixAgent, ensure your system meets these minimum requirements:

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| CPU | 4 cores | 8+ cores |
| RAM | 8 GB | 16+ GB |
| Storage | 50 GB SSD | 100+ GB NVMe SSD |
| Network | 100 Mbps | 1 Gbps |
| OS | Linux (Ubuntu 20.04+), macOS 12+, Windows Server 2019+ | Ubuntu 22.04 LTS |

### Software Dependencies

- **Go 1.23+**: Required for building from source
- **Docker 24.0+**: Required for containerized deployment (recommended)
- **Docker Compose 2.20+**: For orchestrating multi-container deployments
- **PostgreSQL 15+**: Primary database (included in Docker deployment)
- **Redis 7+**: Caching and session management (included in Docker deployment)
- **Git 2.40+**: For cloning the repository

### API Key Requirements

HelixAgent supports 7 LLM providers. While you can run locally with Ollama (no API key required), production deployments typically require at least one cloud provider API key:

| Provider | API Key Environment Variable | Required |
|----------|------------------------------|----------|
| Claude (Anthropic) | `CLAUDE_API_KEY` | Optional |
| DeepSeek | `DEEPSEEK_API_KEY` | Optional |
| Gemini (Google) | `GEMINI_API_KEY` | Optional |
| Qwen (Alibaba) | `QWEN_API_KEY` | Optional |
| ZAI | `ZAI_API_KEY` | Optional |
| Ollama | None (local) | Optional |
| OpenRouter | `OPENROUTER_API_KEY` | Optional |

---

## Installation Methods

HelixAgent offers multiple installation methods to suit different deployment scenarios.

### Method 1: Docker Compose (Recommended)

Docker Compose provides the simplest and most reliable way to deploy HelixAgent with all its dependencies.

#### Step 1: Clone the Repository

```bash
# Clone the repository
git clone https://dev.helix.agent.git
cd helixagent
```

#### Step 2: Configure Environment Variables

Create your environment file by copying the example configuration:

```bash
# Copy the example environment file
cp .env.example .env
```

Edit the `.env` file with your preferred settings:

```bash
# Server Configuration
PORT=8080
GIN_MODE=release
JWT_SECRET=your-secure-jwt-secret-change-in-production

# Database Configuration
DB_HOST=postgres
DB_PORT=5432
DB_USER=helixagent
DB_PASSWORD=your-secure-password
DB_NAME=helixagent_db

# Redis Configuration
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=your-secure-redis-password

# LLM Provider API Keys (add the ones you have)
CLAUDE_API_KEY=sk-ant-your-claude-key
DEEPSEEK_API_KEY=sk-your-deepseek-key
GEMINI_API_KEY=your-gemini-key
OPENROUTER_API_KEY=sk-or-your-openrouter-key

# Ollama (for local LLM testing)
OLLAMA_ENABLED=true
OLLAMA_BASE_URL=http://ollama:11434
OLLAMA_MODEL=llama2
```

#### Step 3: Start Services

```bash
# Start core services (PostgreSQL, Redis, HelixAgent)
docker-compose up -d

# Or start with AI services (adds Ollama)
docker-compose --profile ai up -d

# Or start full stack (adds monitoring with Prometheus/Grafana)
docker-compose --profile full up -d
```

#### Step 4: Verify Installation

```bash
# Check if all services are running
docker-compose ps

# Test the health endpoint
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "ok",
  "timestamp": 1704067200,
  "version": "1.0.0"
}
```

### Method 2: Build from Source

For development or custom deployments, you can build HelixAgent from source.

#### Step 1: Install Go

```bash
# Download and install Go 1.23+
wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify installation
go version
```

#### Step 2: Clone and Build

```bash
# Clone the repository
git clone https://dev.helix.agent.git
cd helixagent

# Download dependencies
go mod download

# Build the binary
make build

# Or build with debug symbols
make build-debug
```

#### Step 3: Set Up Database

You need PostgreSQL and Redis running before starting HelixAgent:

```bash
# Start PostgreSQL
sudo systemctl start postgresql

# Create database and user
sudo -u postgres psql << EOF
CREATE DATABASE helixagent_db;
CREATE USER helixagent WITH ENCRYPTED PASSWORD 'your-password';
GRANT ALL PRIVILEGES ON DATABASE helixagent_db TO helixagent;
EOF

# Start Redis
sudo systemctl start redis-server
```

#### Step 4: Run HelixAgent

```bash
# Set environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=helixagent
export DB_PASSWORD=your-password
export DB_NAME=helixagent_db
export REDIS_HOST=localhost
export REDIS_PORT=6379
export JWT_SECRET=your-jwt-secret

# Run HelixAgent
./helixagent

# Or run in development mode
make run-dev
```

### Method 3: Podman (Alternative to Docker)

HelixAgent fully supports Podman as an alternative container runtime:

```bash
# Enable Podman socket for Docker compatibility
systemctl --user enable --now podman.socket

# Use podman-compose
pip install podman-compose
podman-compose up -d

# Or run directly
podman build -t helixagent:latest .
podman run -d --name helixagent -p 8080:8080 helixagent:latest
```

---

## Initial Configuration

### Configuration Files

HelixAgent uses YAML configuration files located in the `configs/` directory:

| File | Purpose |
|------|---------|
| `development.yaml` | Development environment settings |
| `production.yaml` | Production environment settings |
| `multi-provider.yaml` | Multi-provider configuration |

### Key Configuration Sections

#### Server Configuration

```yaml
server:
  port: 8080
  host: "0.0.0.0"
  environment: "production"
  log_level: "info"
```

#### Database Configuration

```yaml
database:
  host: "${DB_HOST}"
  port: ${DB_PORT}
  user: "${DB_USER}"
  password: "${DB_PASSWORD}"
  name: "${DB_NAME}"
  sslmode: "require"
  max_open_connections: 50
  max_idle_connections: 10
```

#### LLM Provider Configuration

```yaml
llm_providers:
  claude:
    enabled: true
    api_key: "${CLAUDE_API_KEY}"
    model: "claude-3-sonnet-20240229"
    temperature: 0.7
    max_tokens: 4096
    timeout: "60s"

  deepseek:
    enabled: true
    api_key: "${DEEPSEEK_API_KEY}"
    model: "deepseek-chat"
    temperature: 0.7
    max_tokens: 4096

  ollama:
    enabled: true
    base_url: "http://localhost:11434"
    model: "llama2"
```

### Security Configuration

```yaml
security:
  jwt:
    secret: "${JWT_SECRET}"
    expiration: "24h"

  rate_limit:
    requests_per_minute: 60
    burst: 10

  cors:
    allowed_origins:
      - "https://your-frontend.com"
```

---

## Starting HelixAgent

### Using Make Commands

HelixAgent provides convenient Make commands:

```bash
# Build the binary
make build

# Run locally
make run

# Run in development mode (with debug logging)
make run-dev

# Run tests
make test

# Run with coverage
make test-coverage
```

### Using Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f helixagent

# Stop all services
docker-compose down

# Rebuild after code changes
docker-compose build --no-cache
docker-compose up -d
```

### Verifying the Installation

After starting HelixAgent, verify it is running correctly:

```bash
# Basic health check
curl http://localhost:8080/health

# Enhanced health check with provider status
curl http://localhost:8080/v1/health

# List available models
curl http://localhost:8080/v1/models
```

---

## Your First API Request

### Step 1: Register a User Account

```bash
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "myuser",
    "password": "SecurePassword123!",
    "email": "user@example.com"
  }'
```

Response:
```json
{
  "success": true,
  "message": "User registered successfully",
  "user_id": "user_abc123",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Step 2: Authenticate and Get Token

```bash
# Login to get JWT token
TOKEN=$(curl -s -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "myuser",
    "password": "SecurePassword123!"
  }' | jq -r '.token')

echo "Your token: $TOKEN"
```

### Step 3: Make Your First Chat Completion Request

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "helixagent-ensemble",
    "messages": [
      {
        "role": "system",
        "content": "You are a helpful AI assistant."
      },
      {
        "role": "user",
        "content": "What is HelixAgent and what are its key features?"
      }
    ],
    "temperature": 0.7,
    "max_tokens": 500
  }'
```

Response:
```json
{
  "id": "chatcmpl-helixagent-abc123",
  "object": "chat.completion",
  "created": 1704067200,
  "model": "helixagent-ensemble",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "HelixAgent is an AI-powered ensemble LLM service..."
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 35,
    "completion_tokens": 150,
    "total_tokens": 185
  },
  "ensemble": {
    "providers_used": ["deepseek", "claude"],
    "confidence_score": 0.92,
    "voting_strategy": "confidence_weighted"
  }
}
```

---

## Basic Usage Examples

### Using with curl

#### Text Completion

```bash
curl -X POST http://localhost:8080/v1/completions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "deepseek-chat",
    "prompt": "The future of artificial intelligence is",
    "max_tokens": 100,
    "temperature": 0.5
  }'
```

#### Streaming Response

```bash
curl -X POST http://localhost:8080/v1/chat/completions/stream \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "helixagent-ensemble",
    "messages": [
      {"role": "user", "content": "Tell me a short story."}
    ],
    "stream": true
  }'
```

### Using with Python

```python
import requests

BASE_URL = "http://localhost:8080"
TOKEN = "your-jwt-token"

def chat_completion(messages, model="helixagent-ensemble"):
    response = requests.post(
        f"{BASE_URL}/v1/chat/completions",
        headers={
            "Authorization": f"Bearer {TOKEN}",
            "Content-Type": "application/json"
        },
        json={
            "model": model,
            "messages": messages,
            "temperature": 0.7,
            "max_tokens": 500
        }
    )
    return response.json()

# Example usage
result = chat_completion([
    {"role": "user", "content": "Explain quantum computing simply."}
])
print(result["choices"][0]["message"]["content"])
```

### Using with JavaScript/Node.js

```javascript
const axios = require('axios');

const client = axios.create({
  baseURL: 'http://localhost:8080',
  headers: {
    'Authorization': 'Bearer your-jwt-token',
    'Content-Type': 'application/json'
  }
});

async function chatCompletion(messages, model = 'helixagent-ensemble') {
  const response = await client.post('/v1/chat/completions', {
    model,
    messages,
    temperature: 0.7,
    max_tokens: 500
  });
  return response.data;
}

// Example usage
chatCompletion([
  { role: 'user', content: 'What is machine learning?' }
]).then(result => {
  console.log(result.choices[0].message.content);
});
```

### Using with Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

func main() {
    url := "http://localhost:8080/v1/chat/completions"

    payload := map[string]interface{}{
        "model": "helixagent-ensemble",
        "messages": []map[string]string{
            {"role": "user", "content": "Hello, how are you?"},
        },
        "temperature": 0.7,
        "max_tokens":  500,
    }

    body, _ := json.Marshal(payload)
    req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
    req.Header.Set("Authorization", "Bearer your-jwt-token")
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    fmt.Println(result)
}
```

---

## Understanding the Architecture

### Core Components

HelixAgent consists of several key components:

```
                    ┌─────────────────────────────────────┐
                    │           Load Balancer             │
                    └─────────────────┬───────────────────┘
                                      │
                    ┌─────────────────▼───────────────────┐
                    │         HelixAgent Server           │
                    │  ┌──────────┬──────────┬──────────┐ │
                    │  │ Handlers │ Services │ Ensemble │ │
                    │  └──────────┴──────────┴──────────┘ │
                    └─────────────────┬───────────────────┘
                                      │
        ┌─────────────┬───────────────┼───────────────┬────────────┐
        │             │               │               │            │
        ▼             ▼               ▼               ▼            ▼
   ┌─────────┐  ┌─────────┐    ┌─────────┐    ┌─────────┐   ┌─────────┐
   │PostgreSQL│  │  Redis  │    │ Cognee  │    │ Providers│   │Prometheus│
   │    DB    │  │  Cache  │    │Knowledge│    │ Claude   │   │ Metrics  │
   │          │  │         │    │  Graph  │    │ DeepSeek │   │          │
   └─────────┘  └─────────┘    └─────────┘    │ Gemini   │   └─────────┘
                                              │ Qwen     │
                                              │ Ollama   │
                                              │OpenRouter│
                                              └─────────┘
```

### Request Flow

1. **Client Request**: API request arrives at HelixAgent
2. **Authentication**: JWT token validated via middleware
3. **Rate Limiting**: Request rate checked against limits
4. **Model Selection**: Target model(s) determined
5. **Ensemble Orchestration**: Multiple providers queried in parallel
6. **Response Aggregation**: Best response selected via voting strategy
7. **Response**: Final response returned to client

### Key Features

| Feature | Description |
|---------|-------------|
| Ensemble Orchestration | Query multiple LLMs and select best response |
| Provider Registry | Unified interface for 7+ LLM providers |
| AI Debate System | Multi-participant AI debates with consensus |
| Cognee Integration | Knowledge graph and memory capabilities |
| Circuit Breaker | Fault tolerance for provider failures |
| Semantic Caching | Cache similar queries for faster responses |
| Streaming | Real-time response streaming support |

---

## Next Steps

After completing the basic setup, explore these advanced features:

1. **[Provider Configuration](02-provider-configuration.md)**: Learn to configure all 7 LLM providers
2. **[AI Debate System](03-ai-debate-system.md)**: Set up multi-participant AI debates
3. **[API Reference](04-api-reference.md)**: Complete API documentation
4. **[Deployment Guide](05-deployment-guide.md)**: Production deployment strategies
5. **[Administration Guide](06-administration-guide.md)**: User and system management

### Quick Links

- Test different models: `GET /v1/models`
- Check provider health: `GET /v1/providers`
- View metrics: `GET /metrics`
- Access Grafana dashboard: `http://localhost:3000` (if monitoring profile enabled)

---

## Troubleshooting

### Common Issues

#### Port Already in Use

```bash
# Error: port 8080 already in use
# Solution: Find and stop the process
lsof -i :8080
kill -9 <PID>

# Or change the port
export PORT=8081
```

#### Database Connection Failed

```bash
# Check if PostgreSQL is running
docker-compose ps postgres

# Check logs
docker-compose logs postgres

# Verify connection
psql -h localhost -U helixagent -d helixagent_db
```

#### Redis Connection Failed

```bash
# Check if Redis is running
docker-compose ps redis

# Test connection
redis-cli -h localhost -p 6379 ping
```

#### Authentication Errors

```bash
# Ensure JWT_SECRET is set
echo $JWT_SECRET

# Regenerate token
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "myuser", "password": "mypassword"}'
```

#### Provider Not Responding

```bash
# Check provider health
curl http://localhost:8080/v1/providers/deepseek/health

# Verify API key is set
echo $DEEPSEEK_API_KEY

# Check logs for errors
docker-compose logs helixagent | grep -i error
```

### Getting Help

- **Documentation**: Check the `/docs` directory for detailed guides
- **Issues**: Report bugs at [GitHub Issues](https://dev.helix.agent/issues)
- **Community**: Join discussions at [GitHub Discussions](https://dev.helix.agent/discussions)
- **Support**: Contact support@helixagent.ai for enterprise support

---

## Summary

You have successfully installed HelixAgent and made your first API request. HelixAgent provides a powerful platform for orchestrating multiple LLM providers with features like ensemble voting, AI debates, and knowledge graph integration.

Continue to the [Provider Configuration Guide](02-provider-configuration.md) to learn how to configure and optimize each LLM provider for your use case.
