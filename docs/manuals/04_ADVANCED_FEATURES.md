# Chapter 4: Advanced Features

Explore HelixAgent's advanced capabilities.

## AI Debate Ensemble

### How It Works

The AI Debate Ensemble brings together 5 AI minds:

1. **The Analyst** - Systematically analyzes the problem
2. **The Proposer** - Proposes solutions and approaches
3. **The Critic** - Challenges assumptions and finds weaknesses
4. **The Synthesizer** - Combines perspectives into a coherent view
5. **The Mediator** - Weighs arguments and reaches consensus

### Debate Styles

Choose from multiple presentation styles:

| Style | Description |
|-------|-------------|
| `theater` | Theatrical dialogue (default) |
| `novel` | Novel-style prose narration |
| `screenplay` | Screenplay/script format |
| `minimal` | Minimal formatting |

```bash
curl -X POST http://localhost:7061/v1/debates \
  -d '{"topic": "...", "style": "novel"}'
```

### Multi-Round Debates

Configure debate rounds:

```json
{
  "topic": "What is the best approach to AI safety?",
  "rounds": 5,
  "max_tokens_per_round": 500
}
```

## Model Context Protocol (MCP)

### Available MCP Servers

HelixAgent supports 12 MCP servers:

**HelixAgent Remote Servers:**
- helixagent-mcp - Main MCP endpoint
- helixagent-acp - Agent Communication Protocol
- helixagent-lsp - Language Server Protocol
- helixagent-embeddings - Vector embeddings
- helixagent-vision - Image analysis
- helixagent-cognee - Knowledge graph

**Standard Local Servers:**
- filesystem - File system access
- github - GitHub operations
- memory - Persistent memory
- fetch - Web fetching
- puppeteer - Browser automation
- sqlite - SQLite database

### Using MCP Tools

```bash
# List available tools
curl http://localhost:7061/v1/mcp/tools

# Execute a tool
curl -X POST http://localhost:7061/v1/mcp/execute \
  -H "Content-Type: application/json" \
  -d '{
    "tool": "filesystem.read_file",
    "arguments": {"path": "/path/to/file"}
  }'
```

### MCP Resources

```bash
# List resources
curl http://localhost:7061/v1/mcp/resources

# Get resource
curl http://localhost:7061/v1/mcp/resources/my-resource
```

## Caching System

### Tiered Cache

HelixAgent uses a two-tier caching system:

- **L1 Cache** - In-memory (fastest, limited size)
- **L2 Cache** - Redis (persistent, larger capacity)

### Cache Configuration

```yaml
cache:
  l1:
    max_size: 10000
    ttl: 5m

  l2:
    enabled: true
    ttl: 1h
    compression: true
```

### Semantic Caching

Cache similar queries:

```yaml
cache:
  semantic:
    enabled: true
    similarity_threshold: 0.85
    embedding_model: text-embedding-ada-002
```

### Cache Management

```bash
# View cache stats
curl http://localhost:7061/v1/cache/stats

# Clear cache
curl -X DELETE http://localhost:7061/v1/cache

# Invalidate specific keys
curl -X DELETE http://localhost:7061/v1/cache/pattern/provider:*
```

## Background Tasks

### Creating Tasks

```bash
curl -X POST http://localhost:7061/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "type": "document_processing",
    "payload": {
      "file_path": "/path/to/document.pdf"
    },
    "priority": "high"
  }'
```

### Task Monitoring

```bash
# Get task status
curl http://localhost:7061/v1/tasks/task-123

# List all tasks
curl http://localhost:7061/v1/tasks

# Cancel a task
curl -X DELETE http://localhost:7061/v1/tasks/task-123
```

### Notifications

Subscribe to task notifications:

```bash
# SSE notifications
curl -N http://localhost:7061/v1/tasks/task-123/events

# Polling
curl http://localhost:7061/v1/tasks/task-123/poll
```

## Knowledge Graph (Cognee)

### Adding Knowledge

```bash
# Add text
curl -X POST http://localhost:7061/v1/cognee/add \
  -d '{"data": "Python is a programming language..."}'

# Add file
curl -X POST http://localhost:7061/v1/cognee/add/file \
  -F "file=@document.pdf"
```

### Querying Knowledge

```bash
curl -X POST http://localhost:7061/v1/cognee/search \
  -d '{"query": "What programming languages are mentioned?"}'
```

### Knowledge Graph Operations

```bash
# View graph structure
curl http://localhost:7061/v1/cognee/graph

# Get insights
curl http://localhost:7061/v1/cognee/insights
```

## Plugin System

### Available Plugins

Check available plugins:

```bash
curl http://localhost:7061/v1/plugins
```

### Plugin Management

```bash
# Enable plugin
curl -X POST http://localhost:7061/v1/plugins/my-plugin/enable

# Disable plugin
curl -X POST http://localhost:7061/v1/plugins/my-plugin/disable

# Plugin health
curl http://localhost:7061/v1/plugins/my-plugin/health
```

### Hot Reloading

Plugins can be updated without restart:

```bash
# Reload plugin
curl -X POST http://localhost:7061/v1/plugins/my-plugin/reload
```

## Database Optimization

### Connection Pooling

```yaml
database:
  pool:
    max_conns: 25
    min_conns: 5
    max_conn_lifetime: 1h
    health_check_period: 30s
```

### Materialized Views

HelixAgent uses materialized views for performance:

```bash
# Refresh views
curl -X POST http://localhost:7061/v1/admin/refresh-views
```

### Query Optimization

```yaml
database:
  query_optimizer:
    cache_ttl: 5m
    max_prepared_statements: 100
    batch_size: 1000
```

## Performance Tuning

### Worker Pool

```yaml
worker_pool:
  min_workers: 2
  max_workers: 16
  scale_interval: 30s
  idle_timeout: 5m
```

### Event Bus

```yaml
event_bus:
  buffer_size: 10000
  worker_count: 5
```

### HTTP/3 Support

```yaml
server:
  http3:
    enabled: true
    cert_file: /path/to/cert.pem
    key_file: /path/to/key.pem
```

## Monitoring

### Prometheus Metrics

```bash
curl http://localhost:7061/metrics
```

Key metrics:
- `helixagent_requests_total`
- `helixagent_request_duration_seconds`
- `helixagent_provider_health`
- `helixagent_cache_hit_rate`

### Grafana Dashboard

Import the dashboard from `monitoring/grafana/dashboard.json`.

### Health Checks

```bash
# Liveness
curl http://localhost:7061/healthz/live

# Readiness
curl http://localhost:7061/healthz/ready

# Full health
curl http://localhost:7061/health
```

## Security

### Rate Limiting

```yaml
rate_limit:
  requests_per_minute: 100
  tokens_per_minute: 10000
  burst: 20
```

### CORS

```yaml
cors:
  allowed_origins:
    - "https://your-domain.com"
  allowed_methods:
    - GET
    - POST
  max_age: 3600
```

### Input Validation

HelixAgent validates all inputs against OWASP guidelines:
- SQL injection prevention
- XSS protection
- Command injection prevention
- Path traversal protection
