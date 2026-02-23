# HelixAgent Integrations Guide

Complete guide to integrating HelixAgent with LLM providers, vector databases, messaging systems, and other services.

---

## Overview

HelixAgent is designed to integrate with a wide ecosystem of AI and infrastructure services. This guide covers configuration, authentication, and best practices for each integration category.

---

## LLM Provider Integrations

### Provider Overview

| Provider | Auth Method | Key Features | Rate Limits |
|----------|-------------|--------------|-------------|
| Claude | OAuth/API Key | Strong reasoning, safety | Varies by plan |
| DeepSeek | API Key | Coding, cost-effective | 1000 RPM |
| Gemini | API Key | Multimodal, long context | 60 RPM |
| Mistral | API Key | Efficiency, multilingual | Varies |
| OpenRouter | API Key | 100+ models | Varies by model |
| Qwen | OAuth/API Key | Chinese, reasoning | Varies |
| ZAI | API Key | Specialized tasks | Varies |
| Zen | Free/API Key | Development | Unlimited local |
| Cerebras | API Key | Speed | Varies |
| Ollama | None (local) | Privacy, offline | Unlimited |

### Claude Integration

Anthropic's Claude models with CLI proxy support for OAuth authentication.

**Configuration:**

```bash
# Option 1: API Key (recommended)
export CLAUDE_API_KEY="sk-ant-..."

# Option 2: OAuth via CLI proxy
export CLAUDE_USE_OAUTH_CREDENTIALS=true
```

**YAML Config:**

```yaml
providers:
  claude:
    name: "Anthropic Claude"
    type: "claude"
    enabled: true
    api_key: "${CLAUDE_API_KEY}"
    models:
      - name: "claude-3-5-sonnet-20241022"
        display_name: "Claude 3.5 Sonnet"
        capabilities: ["chat", "reasoning", "coding", "vision"]
        max_tokens: 200000
```

**Models Available:**
- `claude-3-5-sonnet-20241022` - Best balance of speed and capability
- `claude-3-opus-20240229` - Most capable, complex tasks
- `claude-3-haiku-20240307` - Fastest, simple tasks

---

### DeepSeek Integration

Cost-effective models excelling at coding and reasoning.

**Configuration:**

```bash
export DEEPSEEK_API_KEY="sk-..."
```

**YAML Config:**

```yaml
providers:
  deepseek:
    name: "DeepSeek"
    type: "deepseek"
    enabled: true
    api_key: "${DEEPSEEK_API_KEY}"
    base_url: "https://api.deepseek.com/v1"
    models:
      - name: "deepseek-chat"
        display_name: "DeepSeek Chat"
        capabilities: ["chat", "reasoning"]
      - name: "deepseek-coder"
        display_name: "DeepSeek Coder"
        capabilities: ["coding", "chat"]
```

---

### Gemini Integration

Google's multimodal models with long context support.

**Configuration:**

```bash
export GEMINI_API_KEY="..."
```

**YAML Config:**

```yaml
providers:
  gemini:
    name: "Google Gemini"
    type: "gemini"
    enabled: true
    api_key: "${GEMINI_API_KEY}"
    models:
      - name: "gemini-pro"
        display_name: "Gemini Pro"
        capabilities: ["chat", "reasoning"]
      - name: "gemini-1.5-pro"
        display_name: "Gemini 1.5 Pro"
        capabilities: ["chat", "multimodal", "long_context"]
        max_tokens: 1000000
```

---

### OpenRouter Integration

Access 100+ models through a single API.

**Configuration:**

```bash
export OPENROUTER_API_KEY="sk-or-..."
```

**YAML Config:**

```yaml
providers:
  openrouter:
    name: "OpenRouter"
    type: "openrouter"
    enabled: true
    api_key: "${OPENROUTER_API_KEY}"
    base_url: "https://openrouter.ai/api/v1"
    models:
      - name: "anthropic/claude-3.5-sonnet"
        display_name: "Claude 3.5 Sonnet (via OR)"
      - name: "google/gemini-2.5-flash"
        display_name: "Gemini 2.5 Flash"
      - name: "x-ai/grok-4"
        display_name: "Grok-4"
      - name: "meta-llama/llama-3.1-70b-instruct:free"
        display_name: "Llama 3.1 70B (Free)"
```

---

### Ollama Integration

Local model execution for privacy and offline use.

**Setup:**

```bash
# Install Ollama
curl -fsSL https://ollama.com/install.sh | sh

# Pull models
ollama pull llama3:70b
ollama pull codellama:34b
ollama pull mixtral:8x7b
```

**YAML Config:**

```yaml
providers:
  ollama:
    name: "Ollama"
    type: "ollama"
    enabled: true
    base_url: "http://localhost:11434"
    models:
      - name: "llama3:70b"
        display_name: "Llama 3 70B"
        capabilities: ["chat", "reasoning"]
      - name: "codellama:34b"
        display_name: "Code Llama 34B"
        capabilities: ["coding"]
```

---

### Multi-Provider Configuration

Complete example with all providers:

```yaml
# configs/multi-provider.yaml
providers:
  claude:
    enabled: true
    api_key: "${CLAUDE_API_KEY}"
    priority: 1

  deepseek:
    enabled: true
    api_key: "${DEEPSEEK_API_KEY}"
    priority: 2

  gemini:
    enabled: true
    api_key: "${GEMINI_API_KEY}"
    priority: 3

  openrouter:
    enabled: true
    api_key: "${OPENROUTER_API_KEY}"
    priority: 4

  ollama:
    enabled: true
    base_url: "http://localhost:11434"
    priority: 5

ensemble:
  strategy: "confidence_weighted"
  min_providers: 2
  fallback_to_best: true
  debate:
    enabled: true
    rounds: 3
    consensus_threshold: 0.7
```

---

## Vector Database Integrations

### Qdrant

High-performance vector database written in Rust.

**Docker Setup:**

```bash
docker run -d --name qdrant \
  -p 6333:6333 \
  -v $(pwd)/qdrant_data:/qdrant/storage \
  qdrant/qdrant:latest
```

**Configuration:**

```yaml
vectordb:
  type: qdrant
  host: localhost
  port: 6333
  collection: helixagent_embeddings
  vector_size: 1536
  distance: cosine
```

**Usage:**

```go
import "dev.helix.agent/internal/vectordb"

client, err := vectordb.NewQdrantClient(&vectordb.QdrantConfig{
    Host:       "localhost",
    Port:       6333,
    Collection: "embeddings",
})

// Upsert vectors
err = client.Upsert(ctx, vectors)

// Search
results, err := client.Search(ctx, queryVector, 10)
```

---

### Pinecone

Managed vector database service.

**Configuration:**

```bash
export PINECONE_API_KEY="..."
export PINECONE_ENVIRONMENT="us-west1-gcp"
```

```yaml
vectordb:
  type: pinecone
  api_key: "${PINECONE_API_KEY}"
  environment: "${PINECONE_ENVIRONMENT}"
  index: helixagent-index
```

---

### Milvus

Open-source vector database for scalable similarity search.

**Docker Setup:**

```bash
docker-compose -f docker-compose.milvus.yaml up -d
```

**Configuration:**

```yaml
vectordb:
  type: milvus
  host: localhost
  port: 19530
  collection: helixagent_vectors
```

---

### pgvector

PostgreSQL extension for vector operations.

**Enable Extension:**

```sql
CREATE EXTENSION IF NOT EXISTS vector;
```

**Configuration:**

```yaml
vectordb:
  type: pgvector
  host: localhost
  port: 5432
  database: helixagent
  table: embeddings
```

---

## Embedding Provider Integrations

### OpenAI Embeddings

```yaml
embedding:
  provider: openai
  model: text-embedding-3-small
  dimensions: 1536
  api_key: "${OPENAI_API_KEY}"
```

### Cohere Embeddings

```yaml
embedding:
  provider: cohere
  model: embed-english-v3.0
  api_key: "${COHERE_API_KEY}"
```

### Voyage AI Embeddings

```yaml
embedding:
  provider: voyage
  model: voyage-2
  api_key: "${VOYAGE_API_KEY}"
```

### Jina Embeddings

```yaml
embedding:
  provider: jina
  model: jina-embeddings-v2-base-en
  api_key: "${JINA_API_KEY}"
```

---

## Messaging System Integrations

### Kafka

For distributed message streaming and big data integration.

**Docker Setup:**

```bash
docker-compose -f docker-compose.kafka.yaml up -d
```

**Configuration:**

```yaml
messaging:
  type: kafka
  brokers:
    - localhost:9092
  topics:
    events: helixagent-events
    analytics: helixagent-analytics
    learning: helixagent-learning
  consumer_group: helixagent-group
```

**Usage:**

```go
import "dev.helix.agent/internal/messaging"

broker, err := messaging.NewKafkaBroker(&messaging.KafkaConfig{
    Brokers: []string{"localhost:9092"},
})

// Publish
err = broker.Publish(ctx, "events", event)

// Subscribe
err = broker.Subscribe(ctx, "events", handler)
```

---

### RabbitMQ

For traditional message queuing.

**Docker Setup:**

```bash
docker run -d --name rabbitmq \
  -p 5672:5672 \
  -p 15672:15672 \
  rabbitmq:3-management
```

**Configuration:**

```yaml
messaging:
  type: rabbitmq
  url: "amqp://guest:guest@localhost:5672/"
  queues:
    tasks: helixagent-tasks
    notifications: helixagent-notifications
```

---

## Database Integrations

### PostgreSQL

Primary database for HelixAgent.

**Docker Setup:**

```bash
docker run -d --name postgresql \
  -e POSTGRES_USER=helixagent \
  -e POSTGRES_PASSWORD=secret \
  -e POSTGRES_DB=helixagent \
  -p 5432:5432 \
  postgres:15
```

**Configuration:**

```yaml
database:
  type: postgresql
  host: localhost
  port: 5432
  user: helixagent
  password: "${DB_PASSWORD}"
  database: helixagent
  ssl_mode: prefer
  pool_size: 20
```

---

### Redis

Caching and session management.

**Docker Setup:**

```bash
docker run -d --name redis \
  -p 6379:6379 \
  redis:7 \
  redis-server --requirepass secret
```

**Configuration:**

```yaml
cache:
  type: redis
  host: localhost
  port: 6379
  password: "${REDIS_PASSWORD}"
  database: 0
  ttl: 3600
```

---

## Analytics Integrations

### ClickHouse

For high-performance analytics.

**Docker Setup:**

```bash
docker run -d --name clickhouse \
  -p 8123:8123 \
  -p 9000:9000 \
  clickhouse/clickhouse-server:latest
```

**Configuration:**

```yaml
analytics:
  type: clickhouse
  host: localhost
  port: 8123
  database: helixagent
  tables:
    events: debate_events
    metrics: provider_metrics
```

---

### Neo4j

For knowledge graph storage.

**Docker Setup:**

```bash
docker run -d --name neo4j \
  -p 7474:7474 \
  -p 7687:7687 \
  -e NEO4J_AUTH=neo4j/password \
  neo4j:5
```

**Configuration:**

```yaml
knowledge_graph:
  type: neo4j
  uri: "bolt://localhost:7687"
  user: neo4j
  password: "${NEO4J_PASSWORD}"
```

---

## Observability Integrations

### Prometheus

```yaml
observability:
  prometheus:
    enabled: true
    port: 9091
    path: /metrics
```

### Jaeger

```yaml
observability:
  tracing:
    provider: jaeger
    endpoint: "http://localhost:14268/api/traces"
    service_name: helixagent
    sample_rate: 0.1
```

### Langfuse

```yaml
observability:
  langfuse:
    enabled: true
    public_key: "${LANGFUSE_PUBLIC_KEY}"
    secret_key: "${LANGFUSE_SECRET_KEY}"
    host: "https://cloud.langfuse.com"
```

---

## MCP Server Integrations

HelixAgent includes 45+ pre-built MCP adapters:

| Category | Adapters |
|----------|----------|
| File System | filesystem, directory, file-search |
| Development | git, github, gitlab, npm |
| Databases | postgres, mysql, sqlite, redis |
| Cloud | aws, gcp, azure |
| Communication | slack, email, discord |
| Productivity | notion, linear, jira |
| Search | brave-search, google-search |

**Configuration:**

```yaml
mcp:
  adapters:
    filesystem:
      enabled: true
      config:
        root: "/workspace"
        allowed_extensions: [".go", ".py", ".js"]

    github:
      enabled: true
      config:
        token: "${GITHUB_TOKEN}"
        repos: ["org/repo1", "org/repo2"]
```

---

## Environment Variable Reference

### Required

| Variable | Purpose |
|----------|---------|
| `JWT_SECRET` | API authentication |

### Provider Keys

| Variable | Provider |
|----------|----------|
| `CLAUDE_API_KEY` | Anthropic Claude |
| `DEEPSEEK_API_KEY` | DeepSeek |
| `GEMINI_API_KEY` | Google Gemini |
| `MISTRAL_API_KEY` | Mistral AI |
| `OPENROUTER_API_KEY` | OpenRouter |
| `QWEN_API_KEY` | Alibaba Qwen |
| `ZAI_API_KEY` | ZAI |
| `CEREBRAS_API_KEY` | Cerebras |

### Infrastructure

| Variable | Purpose |
|----------|---------|
| `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD` | PostgreSQL |
| `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD` | Redis |
| `KAFKA_BROKERS` | Kafka |
| `NEO4J_URI`, `NEO4J_PASSWORD` | Neo4j |

---

## AI/ML Module Integrations

HelixAgent includes five advanced AI/ML modules (Phase 5 extracted modules). Each is accessible via an adapter in `internal/adapters/`.

### Agentic Module (`digital.vasic.agentic`)

Graph-based workflow orchestration for multi-step AI automation.

**Configuration:**

```yaml
agentic:
  enabled: true
  max_concurrent_nodes: 10
  default_timeout: 120s
  retry_policy:
    max_attempts: 3
    backoff: exponential
```

**Usage:**

```go
import agenticadapter "dev.helix.agent/internal/adapters/agentic"

adapter := agenticadapter.NewAdapter(&agenticadapter.Config{
    MaxConcurrentNodes: 10,
    DefaultTimeout:     120 * time.Second,
})

// Define a workflow
workflow := &agentic.Workflow{
    ID:    "code-review-pipeline",
    Nodes: []agentic.Node{
        {ID: "fetch-pr",    Task: fetchPRTask},
        {ID: "review-code", Task: reviewTask,  DependsOn: []string{"fetch-pr"}},
        {ID: "post-review", Task: postTask,    DependsOn: []string{"review-code"}},
    },
}

result, err := adapter.Execute(ctx, workflow)
```

---

### LLMOps Module (`digital.vasic.llmops`)

Production operations tooling: evaluation pipelines, A/B experiments, dataset management, and prompt versioning.

**Configuration:**

```yaml
llmops:
  enabled: true
  evaluation:
    provider: claude
    metrics:
      - accuracy
      - relevance
      - coherence
  experiments:
    significance_threshold: 0.95
    min_sample_size: 100
```

**Usage:**

```go
import llmopsadapter "dev.helix.agent/internal/adapters/llmops"

adapter := llmopsadapter.NewAdapter(cfg)

// Run evaluation pipeline
report, err := adapter.Evaluate(ctx, &llmops.EvalConfig{
    Dataset:  "qa-benchmark-v2",
    Provider: "helixagent-debate",
    Metrics:  []string{"accuracy", "relevance"},
})

// Create A/B experiment
exp, err := adapter.StartExperiment(ctx, &llmops.Experiment{
    Name:     "prompt-v2-vs-v3",
    VariantA: promptV2,
    VariantB: promptV3,
    TrafficSplit: 0.5,
})

// Version a prompt
version, err := adapter.VersionPrompt(ctx, "system-prompt", newPromptContent)
```

---

### SelfImprove Module (`digital.vasic.selfimprove`)

RLHF-based feedback loops, reward modelling, and response optimization.

**Configuration:**

```yaml
selfimprove:
  enabled: true
  reward_model:
    update_interval: 1h
    min_feedback_samples: 50
  rlhf:
    learning_rate: 0.001
    feedback_window: 30d
```

**Usage:**

```go
import selfimproveadapter "dev.helix.agent/internal/adapters/selfimprove"

adapter := selfimproveadapter.NewAdapter(cfg)

// Submit feedback after a response
err := adapter.SubmitFeedback(ctx, &selfimprove.Feedback{
    ResponseID: responseID,
    UserID:     userID,
    Score:      0.9,
    Comment:    "Accurate and concise",
})

// Retrieve learned preferences for prompt enrichment
prefs, err := adapter.GetOptimizedPreferences(ctx, userID)

// Train reward model on collected feedback
err = adapter.TrainRewardModel(ctx)
```

---

### Planning Module (`digital.vasic.planning`)

Hierarchical planning, Monte Carlo Tree Search, and Tree of Thoughts algorithms.

**Configuration:**

```yaml
planning:
  enabled: true
  hiplan:
    max_depth: 5
    max_subtasks: 20
  mcts:
    simulations: 1000
    exploration_constant: 1.414
  tree_of_thoughts:
    branches: 3
    depth: 4
    eval_provider: claude
```

**Usage:**

```go
import planningadapter "dev.helix.agent/internal/adapters/planning"

adapter := planningadapter.NewAdapter(cfg)

// Hierarchical planning
plan, err := adapter.HiPlan(ctx, &planning.Goal{
    Description: "Refactor the authentication module",
    Constraints: []string{"no breaking API changes", "maintain test coverage"},
})

// Monte Carlo Tree Search for exploration problems
result, err := adapter.MCTS(ctx, initialState, 500 /* simulations */)

// Tree of Thoughts for deliberate reasoning
thoughts, err := adapter.TreeOfThoughts(ctx, &planning.ToTConfig{
    Problem:  "Design a caching strategy for the API",
    Branches: 3,
    Depth:    4,
})
```

---

### Benchmark Module (`digital.vasic.benchmark`)

Standardized quality benchmarking against SWE-bench, HumanEval, and MMLU.

**Configuration:**

```yaml
benchmark:
  enabled: true
  suites:
    - humaneval
    - mmlu
    - swe-bench
  parallel: 4
  result_store:
    type: postgresql
    table: benchmark_results
```

**Usage:**

```go
import benchmarkadapter "dev.helix.agent/internal/adapters/benchmark"

adapter := benchmarkadapter.NewAdapter(cfg)

// Run HumanEval (code generation)
humanEvalResult, err := adapter.RunHumanEval(ctx, &benchmark.RunConfig{
    Provider: "helixagent-debate",
    Problems: 164, // full suite
})

fmt.Printf("HumanEval pass@1: %.2f%%\n", humanEvalResult.PassAt1 * 100)

// Run MMLU (reasoning across 57 subjects)
mmluResult, err := adapter.RunMMLU(ctx, &benchmark.RunConfig{
    Provider: "helixagent-debate",
    Subjects: benchmark.AllSubjects,
})

fmt.Printf("MMLU accuracy: %.2f%%\n", mmluResult.Accuracy * 100)

// Run SWE-bench (software engineering tasks)
sweResult, err := adapter.RunSWEBench(ctx, &benchmark.RunConfig{
    Provider: "helixagent-debate",
    Split:    "test",
})

fmt.Printf("SWE-bench resolved: %.2f%%\n", sweResult.ResolvedRate * 100)

// Compare providers across all benchmarks
comparison, err := adapter.CompareProviders(ctx, []string{
    "claude", "deepseek", "gemini", "helixagent-debate",
})
```

---

## Integration Testing

Verify integrations with challenges:

```bash
# Test all providers
./challenges/scripts/integration_providers_challenge.sh

# Test specific provider
./challenges/scripts/test_provider.sh claude

# Test vector database
./challenges/scripts/vectordb_challenge.sh qdrant

# Test messaging
./challenges/scripts/messaging_challenge.sh kafka
```

---

## Best Practices

### Provider Selection

1. **Diversity**: Include providers from different vendors
2. **Capability Match**: Select models suited to your use case
3. **Cost Awareness**: Balance quality with cost
4. **Fallback Coverage**: Ensure fallbacks for each primary

### Security

1. **Never hardcode** API keys in code or configs
2. **Use secrets management** (Vault, AWS Secrets Manager)
3. **Rotate keys** regularly
4. **Audit access** to sensitive integrations

### Performance

1. **Connection pooling** for databases
2. **Cache frequently** accessed data
3. **Use async** for non-blocking operations
4. **Monitor latencies** for all integrations

---

## Troubleshooting

### Provider Connection Issues

```bash
# Check provider health
curl http://localhost:8080/v1/monitoring/provider-health

# Verify API key
curl -H "Authorization: Bearer $PROVIDER_API_KEY" \
  https://api.provider.com/health
```

### Database Issues

```bash
# Check connection
psql -h localhost -U helixagent -d helixagent -c "SELECT 1"

# Check Redis
redis-cli -h localhost ping
```

### Observability

```bash
# Check Prometheus metrics
curl http://localhost:9091/metrics

# Check health endpoint
curl http://localhost:8080/health
```

---

**Last Updated**: February 2026
**Version**: 1.0.0
