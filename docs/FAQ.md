# HelixAgent Frequently Asked Questions

## Getting Started

### What is HelixAgent?

HelixAgent is an AI-powered ensemble LLM service that combines responses from multiple language models using intelligent aggregation strategies. It provides OpenAI-compatible APIs and supports 43 LLM providers with dynamic selection based on real-time verification scores.

### How do I run HelixAgent for the first time?

```bash
# 1. Copy environment file and add your API keys
cp .env.example .env
# Edit .env with your provider API keys

# 2. Build the binary
make build

# 3. Run (containers start automatically)
./bin/helixagent
```

The binary automatically orchestrates all required containers (PostgreSQL, Redis, etc.) on startup. Do not start containers manually.

### What are the minimum requirements?

- Go 1.25.3+
- Docker or Podman (for container orchestration)
- At least one LLM provider API key
- 4 GB RAM recommended (8 GB for full stack)

### Which LLM providers are supported?

43 providers including: Claude, Chutes, DeepSeek, Gemini, Mistral, OpenRouter, Qwen, ZAI, Zen, Cerebras, Ollama, AI21, Anthropic, Cohere, Fireworks, GitHub Models, Groq, HuggingFace, OpenAI, Perplexity, Replicate, Together, Venice, xAI, Junie, Cloudflare, and more.

---

## Provider Configuration

### How do I configure a provider API key?

Add the key to your `.env` file:

```bash
OPENAI_API_KEY=sk-...
DEEPSEEK_API_KEY=...
GEMINI_API_KEY=...
```

See `.env.example` for the full list of supported environment variables.

### How do OAuth-based providers (Claude, Qwen, Gemini) work?

These providers can use CLI proxies when direct API access is restricted:

```bash
# Use OAuth credentials instead of API key
CLAUDE_USE_OAUTH_CREDENTIALS=true
QWEN_USE_OAUTH_CREDENTIALS=true
```

The provider must be installed and authenticated locally (e.g., `claude`, `qwen`, `gemini` CLI tools).

### How does dynamic provider selection work?

On startup, LLMsVerifier runs an 8-test verification pipeline for each configured provider and scores them across 5 weighted dimensions: ResponseSpeed (25%), CostEffectiveness (25%), ModelEfficiency (20%), Capability (20%), Recency (10%). Providers scoring below 5.0 are excluded. The top-scoring providers form the debate team.

### Why is a provider not being selected?

Check the startup verification log. Common reasons:
- API key missing or invalid
- Provider scored below 5.0 minimum
- Network connectivity issue
- Provider rate limit hit during verification

```bash
# Check startup verification status
curl http://localhost:7061/v1/startup/verification
```

---

## Debugging

### How do I see what HelixAgent is doing?

```bash
# Server log
tail -f /tmp/helixagent-server.log

# Health status
curl http://localhost:7061/v1/health

# Provider health
curl http://localhost:7061/v1/monitoring/status

# Circuit breaker states
make monitoring-circuit-breakers
```

### Why is the server slow to start?

Startup runs provider verification in parallel (typically 1-2 minutes). This is normal. The server is not ready until `/v1/health` returns `200 OK`.

### How do I run tests?

```bash
# Unit tests only (no infrastructure needed)
make test-unit

# All tests (requires running containers)
make test-with-infra

# Single test
go test -v -run TestName ./path/to/package
```

### How do I debug a failing provider?

```bash
# Check provider health endpoint
curl http://localhost:7061/v1/verification

# Run verification manually
make verifier-verify MODEL=gpt-4 PROVIDER=openai

# Check circuit breaker state
make monitoring-reset-circuits
```

---

## Deployment

### How do I deploy to production?

See [Deployment Guide](deployment/DEPLOYMENT_GUIDE.md). Key steps:
1. Build release binary inside container: `make release`
2. Configure `Containers/.env` for remote or local orchestration
3. Set `GIN_MODE=release` and strong `JWT_SECRET`
4. Run `./bin/helixagent` — all containers start automatically

### Can I deploy containers to a remote host?

Yes. Set in `Containers/.env`:

```bash
CONTAINERS_REMOTE_ENABLED=true
CONTAINERS_REMOTE_HOST_1=user@hostname
```

HelixAgent will distribute all containers to the remote host automatically on startup.

### How do I build release binaries?

```bash
# Build for all platforms (runs inside container for reproducibility)
make release

# Build for all 7 apps
make release-all
```

### What ports does HelixAgent use?

- `7061` — Main API (HTTP/3 + HTTP/2 fallback)
- `15432` — PostgreSQL (test infra)
- `16379` — Redis (test infra)
- `18081` — Mock LLM (test infra)
- `9101-9999` — MCP servers

---

## Architecture

### What is the AI Debate system?

The debate system runs a query through multiple LLM providers simultaneously (5 positions × 5 LLMs = 25 total responses), then uses voting strategies (Weighted, Majority, Borda Count, Condorcet, Plurality, Unanimous) to produce a consensus answer. Use the `/v1/chat/completions` endpoint with model `helixagent-debate`.

### What is SpecKit?

SpecKit is a 7-phase Spec-Driven Development workflow (Constitution → Specify → Clarify → Plan → Tasks → Analyze → Implement) that auto-activates for large changes or refactoring. It integrates with the AI debate system for specification refinement. See [SpecKit User Guide](guides/SPECKIT_USER_GUIDE.md).

### What are the extracted modules?

HelixAgent's functionality is decomposed into 41 independent Go modules (EventBus, Concurrency, Observability, Auth, Storage, Streaming, Security, VectorDB, Embeddings, Database, Cache, Messaging, Formatters, MCP, RAG, Memory, Optimization, Plugins, Agentic, LLMOps, SelfImprove, Planning, Benchmark, HelixMemory, HelixSpecifier, and more). See [MODULES.md](MODULES.md) for the full catalog.

### How does the container orchestration work?

All container management goes through the `Containers` module adapter (`internal/adapters/containers/adapter.go`). The adapter reads `Containers/.env` on startup and either starts containers locally or distributes them to remote hosts. Never manipulate containers directly — always let HelixAgent manage them.

### What is HelixMemory?

HelixMemory is a unified cognitive memory engine that fuses Mem0 (facts), Cognee (knowledge graphs), Letta (stateful agent runtime), and Graphiti (temporal graph) into a single orchestrated system with a 3-stage fusion pipeline. It is active by default; opt out with `-tags nohelixmemory`.

---

## Code Formatters

### How many code formatters are supported?

32+ formatters for 19 languages: 11 native (run in-process), 14 service-based (run in Docker containers on ports 9210-9300), and 7 built-in.

### How do I format code via the API?

```bash
curl -X POST http://localhost:7061/v1/format \
  -H "Content-Type: application/json" \
  -d '{"language": "go", "code": "package main\nfunc main(){}"}'
```

---

## Security

### How do I run security scans?

```bash
make security-scan-all     # All scanners
make security-scan-gosec   # Go source analysis
make security-scan-snyk    # Dependency vulnerabilities
make security-scan-sonarqube  # Code quality + security
```

See [Security Scanning Guide](security/scanning-guide.md) for details.

### How do I report a vulnerability?

See [Vulnerability Disclosure Policy](security/vulnerability-disclosure.md). Do not open public GitHub issues for security vulnerabilities.
