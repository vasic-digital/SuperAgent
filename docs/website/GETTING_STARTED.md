# Getting Started with HelixAgent

Get up and running with HelixAgent in minutes. This guide covers installation, basic configuration, and your first API call.

---

## Quick Start

### Prerequisites

- **Go 1.24+** - [Install Go](https://go.dev/doc/install)
- **Docker** (optional) - For containerized deployment
- **API Keys** - At least one LLM provider API key

### 1. Install HelixAgent

**Option A: Go Install**

```bash
go install dev.helix.agent/cmd/helixagent@latest
```

**Option B: Build from Source**

```bash
git clone git@github.com:helixagent/helixagent.git
cd helixagent
make build
```

**Option C: Docker**

```bash
docker pull helixagent/helixagent:latest
```

### 2. Configure API Keys

Set at least one provider API key:

```bash
# Option 1: DeepSeek (recommended for getting started - cost effective)
export DEEPSEEK_API_KEY="sk-your-deepseek-key"

# Option 2: OpenRouter (access to many models)
export OPENROUTER_API_KEY="sk-or-your-key"

# Option 3: Gemini
export GEMINI_API_KEY="your-gemini-key"

# For authentication
export JWT_SECRET="your-secret-key-at-least-32-chars"
```

### 3. Start the Server

```bash
# Using binary
helixagent serve

# Using make
make run

# Using Docker
docker run -d -p 8080:8080 \
  -e DEEPSEEK_API_KEY=$DEEPSEEK_API_KEY \
  -e JWT_SECRET=$JWT_SECRET \
  helixagent/helixagent:latest
```

### 4. Test the API

```bash
# Health check
curl http://localhost:8080/health

# Chat completion
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test-key" \
  -d '{
    "model": "helixagent-debate",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

You should see a response from the AI debate ensemble.

---

## Basic Configuration

### Configuration File

Create `configs/development.yaml`:

```yaml
server:
  port: 8080
  mode: development
  timeout: 120s

providers:
  deepseek:
    enabled: true
    api_key: "${DEEPSEEK_API_KEY}"

  openrouter:
    enabled: true
    api_key: "${OPENROUTER_API_KEY}"

ensemble:
  strategy: confidence_weighted
  min_providers: 2
  debate:
    enabled: true
    rounds: 3

cache:
  enabled: true
  type: memory
  ttl: 3600

logging:
  level: info
  format: json
```

### Environment Variables

Create a `.env` file:

```bash
# Server
PORT=8080
GIN_MODE=release
JWT_SECRET=your-secret-key-at-least-32-chars

# Providers
DEEPSEEK_API_KEY=sk-...
OPENROUTER_API_KEY=sk-or-...
GEMINI_API_KEY=...

# Optional: Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=helixagent
DB_PASSWORD=secret
DB_NAME=helixagent

# Optional: Cache
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=secret
```

---

## Using the API

### OpenAI-Compatible Endpoints

HelixAgent is 100% compatible with OpenAI's API format.

#### Chat Completions

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-api-key" \
  -d '{
    "model": "helixagent-debate",
    "messages": [
      {"role": "system", "content": "You are a helpful assistant."},
      {"role": "user", "content": "Explain quantum computing in simple terms."}
    ],
    "temperature": 0.7,
    "max_tokens": 500
  }'
```

#### Streaming

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-api-key" \
  -d '{
    "model": "helixagent-debate",
    "messages": [{"role": "user", "content": "Tell me a story."}],
    "stream": true
  }'
```

#### List Models

```bash
curl http://localhost:8080/v1/models \
  -H "Authorization: Bearer your-api-key"
```

### Using with SDKs

#### Python (OpenAI SDK)

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/v1",
    api_key="your-api-key"
)

response = client.chat.completions.create(
    model="helixagent-debate",
    messages=[
        {"role": "system", "content": "You are a helpful assistant."},
        {"role": "user", "content": "What is machine learning?"}
    ]
)

print(response.choices[0].message.content)
```

#### Node.js (OpenAI SDK)

```javascript
import OpenAI from 'openai';

const client = new OpenAI({
  baseURL: 'http://localhost:8080/v1',
  apiKey: 'your-api-key',
});

const response = await client.chat.completions.create({
  model: 'helixagent-debate',
  messages: [
    { role: 'system', content: 'You are a helpful assistant.' },
    { role: 'user', content: 'What is machine learning?' }
  ],
});

console.log(response.choices[0].message.content);
```

#### LangChain

```python
from langchain_openai import ChatOpenAI

llm = ChatOpenAI(
    base_url="http://localhost:8080/v1",
    api_key="your-api-key",
    model="helixagent-debate"
)

response = llm.invoke("Explain the concept of recursion.")
print(response.content)
```

---

## Adding More Providers

Add more providers to improve debate quality:

### 1. Get API Keys

| Provider | Get Key |
|----------|---------|
| DeepSeek | [platform.deepseek.com](https://platform.deepseek.com) |
| OpenRouter | [openrouter.ai/keys](https://openrouter.ai/keys) |
| Gemini | [makersuite.google.com](https://makersuite.google.com/app/apikey) |
| Claude | [console.anthropic.com](https://console.anthropic.com) |
| Mistral | [console.mistral.ai](https://console.mistral.ai) |

### 2. Add to Configuration

```bash
# Add to .env
export CLAUDE_API_KEY=sk-ant-...
export GEMINI_API_KEY=...
export MISTRAL_API_KEY=...
```

### 3. Update Config

```yaml
providers:
  claude:
    enabled: true
    api_key: "${CLAUDE_API_KEY}"

  gemini:
    enabled: true
    api_key: "${GEMINI_API_KEY}"

  mistral:
    enabled: true
    api_key: "${MISTRAL_API_KEY}"
```

### 4. Restart

```bash
helixagent serve
```

HelixAgent will automatically verify each provider and include the best performers in the debate team.

---

## Optional: Database Setup

For production use, set up PostgreSQL and Redis:

### PostgreSQL

```bash
docker run -d --name postgresql \
  -e POSTGRES_USER=helixagent \
  -e POSTGRES_PASSWORD=secret \
  -e POSTGRES_DB=helixagent \
  -p 5432:5432 \
  postgres:15
```

### Redis

```bash
docker run -d --name redis \
  -p 6379:6379 \
  redis:7 redis-server --requirepass secret
```

### Configuration

```yaml
database:
  host: localhost
  port: 5432
  user: helixagent
  password: secret
  database: helixagent

cache:
  type: redis
  host: localhost
  port: 6379
  password: secret
```

---

## Docker Compose Setup

For a complete local environment:

```yaml
# docker-compose.yaml
version: '3.8'

services:
  helixagent:
    image: helixagent/helixagent:latest
    ports:
      - "8080:8080"
    environment:
      - JWT_SECRET=${JWT_SECRET}
      - DEEPSEEK_API_KEY=${DEEPSEEK_API_KEY}
      - DB_HOST=postgresql
      - DB_PORT=5432
      - DB_USER=helixagent
      - DB_PASSWORD=secret
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - REDIS_PASSWORD=secret
    depends_on:
      - postgresql
      - redis

  postgresql:
    image: postgres:15
    environment:
      - POSTGRES_USER=helixagent
      - POSTGRES_PASSWORD=secret
      - POSTGRES_DB=helixagent
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7
    command: redis-server --requirepass secret
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

Start:

```bash
docker-compose up -d
```

---

## Verify Installation

### Run Health Check

```bash
curl http://localhost:8080/health
```

Expected response:

```json
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime": "1h23m45s"
}
```

### Check Providers

```bash
curl http://localhost:8080/v1/monitoring/provider-health \
  -H "Authorization: Bearer your-api-key"
```

### Run Test Chat

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-api-key" \
  -d '{
    "model": "helixagent-debate",
    "messages": [{"role": "user", "content": "What is 2+2?"}]
  }'
```

---

## Using AI/ML Modules

HelixAgent ships with five advanced AI/ML modules accessible via adapter bridges in `internal/adapters/`.

### Agentic Workflows

Run multi-step graph workflows:

```go
import agenticadapter "dev.helix.agent/internal/adapters/agentic"

adapter := agenticadapter.NewAdapter(cfg)

// Execute a workflow defined as a directed task graph
result, err := adapter.Execute(ctx, workflow)
```

### LLMOps — Evaluation and Experiments

Measure and compare response quality:

```go
import llmopsadapter "dev.helix.agent/internal/adapters/llmops"

adapter := llmopsadapter.NewAdapter(cfg)

// Run an evaluation pipeline against a named dataset
report, err := adapter.Evaluate(ctx, "my-dataset", evalConfig)

// Start an A/B experiment between two prompt variants
experiment, err := adapter.StartExperiment(ctx, variantA, variantB)
```

### SelfImprove — Feedback-Driven Optimization

Collect feedback and improve response selection:

```go
import selfimproveadapter "dev.helix.agent/internal/adapters/selfimprove"

adapter := selfimproveadapter.NewAdapter(cfg)

// Submit human feedback for a completed response
err := adapter.SubmitFeedback(ctx, responseID, feedbackScore)

// Retrieve optimized preferences for use in prompts
prefs, err := adapter.GetOptimizedPreferences(ctx, userID)
```

### Planning — Algorithmic Problem Solving

Decompose complex goals into executable plans:

```go
import planningadapter "dev.helix.agent/internal/adapters/planning"

adapter := planningadapter.NewAdapter(cfg)

// Run hierarchical planning
plan, err := adapter.HiPlan(ctx, goal, constraints)

// Use Monte Carlo Tree Search for exploration
result, err := adapter.MCTS(ctx, initialState, simulations)
```

### Benchmarking

Run quality benchmarks against industry-standard suites:

```go
import benchmarkadapter "dev.helix.agent/internal/adapters/benchmark"

adapter := benchmarkadapter.NewAdapter(cfg)

// Run HumanEval coding benchmark
scores, err := adapter.RunHumanEval(ctx, providerName)

// Run MMLU reasoning benchmark
scores, err := adapter.RunMMLU(ctx, providerName)
```

All five modules are independent Go modules with their own `go.mod`, tests, CLAUDE.md, and challenges. See the submodule directories (`Agentic/`, `LLMOps/`, `SelfImprove/`, `Planning/`, `Benchmark/`) for module-specific documentation.

---

## Next Steps

Now that HelixAgent is running, explore:

1. **[Features](./FEATURES.md)** - Discover all capabilities
2. **[Architecture](./ARCHITECTURE.md)** - Understand the system design
3. **[Integrations](./INTEGRATIONS.md)** - Add more providers and services
4. **[API Reference](/docs/api/README.md)** - Complete API documentation
5. **[Security](./SECURITY.md)** - Security configuration

---

## Common Issues

### Port Already in Use

```bash
# Find process using port 8080
lsof -i :8080

# Kill it or use a different port
PORT=9000 helixagent serve
```

### Provider Verification Failed

1. Check API key is correct
2. Verify network connectivity
3. Check provider status page
4. Review logs: `helixagent serve --log-level=debug`

### Database Connection Error

1. Verify PostgreSQL is running
2. Check credentials
3. Ensure database exists
4. Test connection: `psql -h localhost -U helixagent -d helixagent`

### Redis Connection Error

1. Verify Redis is running
2. Check password
3. Test connection: `redis-cli -h localhost ping`

---

## Getting Help

- **Documentation**: [docs.helixagent.ai](https://docs.helixagent.ai)
- **GitHub Issues**: [github.com/helixagent/helixagent/issues](https://github.com/helixagent/helixagent/issues)
- **Community Discord**: [discord.gg/helixagent](https://discord.gg/helixagent)
- **Email Support**: support@helixagent.ai

---

**Last Updated**: February 2026
**Version**: 1.0.0
