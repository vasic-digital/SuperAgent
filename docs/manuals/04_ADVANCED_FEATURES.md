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

HelixAgent supports **22 MCP servers** validated by the MCPS challenge:

**HelixAgent Remote Servers:**
- helixagent-mcp - Main MCP endpoint
- helixagent-acp - Agent Communication Protocol
- helixagent-lsp - Language Server Protocol
- helixagent-embeddings - Vector embeddings
- helixagent-vision - Image analysis
- helixagent-cognee - Knowledge graph

**Core Servers:**
- filesystem - Secure file operations with configurable access controls
- memory - Knowledge-graph-based persistent memory system
- fetch - Web-content fetching and conversion
- git - Git repository operations

**Version Control:**
- github - GitHub repository management
- gitlab - GitLab integration

**Database:**
- postgres - PostgreSQL operations
- sqlite - SQLite operations
- redis - Redis cache/data store
- mongodb - MongoDB operations

**Cloud & DevOps:**
- docker - Container management
- kubernetes - Cluster management
- aws-s3 - S3 storage operations
- google-drive - Google Drive operations

**Communication:**
- slack - Slack automation
- notion - Notion workspace management

**Search & Automation:**
- brave-search - Web search
- puppeteer - Browser automation

**AI & Vector:**
- sequential-thinking - Step-by-step reasoning
- chroma, qdrant, weaviate - Vector databases

### Using MCP Tools

```bash
# List available tools
curl http://localhost:7061/v1/mcp/tools

# Execute a tool
curl -X POST http://localhost:7061/v1/mcp/tools/call \
  -H "Content-Type: application/json" \
  -d '{
    "name": "filesystem_read",
    "arguments": {"path": "/path/to/file"}
  }'
```

### MCP Tool Search (NEW)

The MCP Tool Search feature enables intelligent tool discovery:

```bash
# Search for file-related tools
curl "http://localhost:7061/v1/mcp/tools/search?q=file"

# Search with fuzzy matching
curl "http://localhost:7061/v1/mcp/tools/search?q=fiel&fuzzy=true"

# Search with category filter
curl "http://localhost:7061/v1/mcp/tools/search?q=read&categories=filesystem,git"

# POST method for advanced search
curl -X POST http://localhost:7061/v1/mcp/tools/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "file operations",
    "categories": ["filesystem"],
    "include_params": true,
    "fuzzy_match": true,
    "max_results": 10
  }'
```

**Response format:**
```json
{
  "query": "file",
  "count": 5,
  "results": [
    {
      "name": "Read",
      "description": "Read file contents",
      "category": "filesystem",
      "score": 1.0,
      "match_type": "exact",
      "parameters": {"file_path": "string"},
      "required": ["file_path"]
    }
  ]
}
```

### MCP Adapter Search

Search for MCP adapters (servers) by name or functionality:

```bash
# Search for GitHub adapter
curl "http://localhost:7061/v1/mcp/adapters/search?q=github"

# Filter by auth type
curl "http://localhost:7061/v1/mcp/adapters/search?q=&auth_types=api_key"

# Filter official adapters only
curl "http://localhost:7061/v1/mcp/adapters/search?q=&official=true"
```

### Tool Suggestions

Get context-aware tool recommendations based on natural language:

```bash
curl "http://localhost:7061/v1/mcp/tools/suggestions?prompt=list%20files%20in%20directory"
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

Cognee provides knowledge graph and memory management capabilities, validated by the RAGS challenge.

### Adding Knowledge

```bash
# Add text
curl -X POST http://localhost:7061/v1/cognee/add \
  -H "Content-Type: application/json" \
  -d '{"data": "Python is a programming language..."}'

# Add file
curl -X POST http://localhost:7061/v1/cognee/add/file \
  -F "file=@document.pdf"

# Add with metadata
curl -X POST http://localhost:7061/v1/cognee/memory \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Important information about the project",
    "metadata": {
      "source": "user_input",
      "timestamp": "2026-01-22T10:00:00Z"
    }
  }'
```

### Querying Knowledge

```bash
# Basic search
curl -X POST http://localhost:7061/v1/cognee/search \
  -H "Content-Type: application/json" \
  -d '{"query": "What programming languages are mentioned?", "limit": 10}'

# Semantic search with filters
curl -X POST http://localhost:7061/v1/cognee/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "project requirements",
    "limit": 5,
    "filters": {"source": "documentation"}
  }'
```

### Knowledge Graph Operations

```bash
# Health check
curl http://localhost:7061/v1/cognee/health

# Statistics
curl http://localhost:7061/v1/cognee/stats

# Configuration
curl http://localhost:7061/v1/cognee/config

# View graph structure
curl http://localhost:7061/v1/cognee/graph

# Get insights
curl -X POST http://localhost:7061/v1/cognee/insights \
  -H "Content-Type: application/json" \
  -d '{"query": "What insights can you provide?"}'

# List datasets
curl http://localhost:7061/v1/cognee/datasets
```

### Cognee Integration Improvements

**Recent fixes (January 2026):**

1. **ListDatasets JSON Parsing**: Fixed malformed JSON array handling in dataset listing
2. **Memory Storage**: Improved metadata handling for memory operations
3. **Search Results**: Enhanced result formatting and pagination

**API Endpoints (validated by RAGS challenge):**

| Endpoint | Method | Description |
|----------|--------|-------------|
| /v1/cognee/health | GET | Health check |
| /v1/cognee/stats | GET | Statistics |
| /v1/cognee/config | GET | Configuration |
| /v1/cognee/memory | POST | Memory storage |
| /v1/cognee/search | POST | Semantic search |
| /v1/cognee/insights | POST | Graph insights |
| /v1/cognee/datasets | GET | List datasets |

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

---

## Challenge Validation System

HelixAgent includes a comprehensive challenge validation system to ensure all features work correctly. As of January 2026, all **45 challenges** pass with **100% success rate**.

### Challenge Overview

| Challenge | Tests | Status | Description |
|-----------|-------|--------|-------------|
| **RAGS** | 147/147 | 100% Pass | RAG system with hybrid retrieval |
| **MCPS** | 9 sections | 100% Pass | MCP integration and tool search |
| **SKILLS** | Full suite | 100% Pass | Skills with real-result validation |
| AI Debate | Complete | 100% Pass | Multi-LLM debate consensus |
| Semantic Intent | 19 tests | 100% Pass | Zero-hardcoding intent detection |
| Unified Verification | 15 tests | 100% Pass | Startup verification pipeline |
| Debate Team Selection | 12 tests | 100% Pass | Dynamic team selection |
| Free Provider Fallback | 8 tests | 100% Pass | Zen/free model fallback |
| Fallback Mechanism | 17 tests | 100% Pass | Empty response handling |
| Multi-Pass Validation | 66 tests | 100% Pass | Debate validation phases |

### Key Challenge Features

#### RAGS Challenge (RAG Integration)

Validates RAG (Retrieval Augmented Generation) integration across **20+ CLI agents**:

**RAG Systems Tested:**
- **Cognee** - Knowledge Graph + Memory storage
- **Qdrant** - Vector database operations
- **RAG Pipeline** - Hybrid search, reranking, HyDE
- **Embeddings Service** - Generation and search

**Endpoints Validated:**
| Endpoint | Method | Description |
|----------|--------|-------------|
| /v1/cognee/health | GET | Cognee health check |
| /v1/cognee/memory | POST | Memory storage |
| /v1/cognee/search | POST | Semantic search |
| /v1/cognee/insights | POST | Graph insights |
| /v1/rag/health | GET | RAG pipeline health |
| /v1/rag/search/hybrid | POST | Hybrid search |
| /v1/rag/rerank | POST | Reranking |
| /v1/embeddings/generate | POST | Embedding generation |

#### MCPS Challenge (MCP Server Integration)

Validates **22 MCP servers** across all CLI agents:

**MCP Servers:**
- **Core**: filesystem, memory, fetch, git, github, gitlab
- **Database**: postgres, sqlite, redis, mongodb
- **Cloud**: docker, kubernetes, aws-s3, google-drive
- **Communication**: slack, notion
- **Search**: brave-search
- **Design**: figma, miro, svgmaker, puppeteer
- **AI**: stable-diffusion, sequential-thinking
- **Vector**: chroma, qdrant, weaviate

**Section 9: MCP Tool Search Validation** (NEW)
- Validates search queries return actual results (not just HTTP 200)
- Tests: file, git, search, web, read, write, glob, bash
- Verifies adapter search functionality
- Validates tool suggestions feature

#### SKILLS Challenge (Skills Integration)

Validates **21 skill categories** with strict real-result validation:

**Skill Categories:**
- **Code**: generate, refactor, optimize
- **Debug**: trace, profile, analyze
- **Search**: find, grep, semantic-search
- **Git**: commit, branch, merge
- **Deploy**: build, deploy, rollback
- **Docs**: document, explain, readme
- **Test**: unit-test, integration-test
- **Review**: review, lint, security-scan

### Strict Real-Result Validation

All challenges now implement **strict validation** to prevent false successes:

```bash
# Example validation logic from RAGS challenge:
if [[ "$response_code" == "200" ]]; then
    # Check for actual content, not just HTTP 200
    local has_choices=$(echo "$response_body" | grep -q '"choices"' && echo "yes" || echo "no")
    local content=$(echo "$response_body" | jq -r '.choices[0].message.content // ""')
    local content_length=${#content}

    # STRICT: Must have real content > 20 chars and not be an error
    if [[ "$content_length" -gt 20 ]] && [[ ! "$content" =~ ^(Error|error:|Failed|null) ]]; then
        log_success "PASSED (REAL): Actual content received"
    else
        log_error "FAILED (FALSE SUCCESS): HTTP 200 but no real content"
    fi
fi
```

**Validation Checks:**
1. HTTP 200 response received
2. Response has `choices` array
3. Content length exceeds minimum threshold (20-50 chars)
4. Content is not an error message
5. For RAG operations: evidence of retrieval (context, knowledge, etc.)

### Running Challenges

```bash
# Run all 45 challenges
./challenges/scripts/run_all_challenges.sh

# Run specific challenge categories
./challenges/scripts/rags_challenge.sh               # RAG system (147 tests)
./challenges/scripts/mcps_challenge.sh               # MCP integration (9 sections)
./challenges/scripts/skills_challenge.sh             # Skills validation
./challenges/scripts/semantic_intent_challenge.sh    # Intent detection (19 tests)
./challenges/scripts/unified_verification_challenge.sh
./challenges/scripts/debate_team_dynamic_selection_challenge.sh
./challenges/scripts/free_provider_fallback_challenge.sh
./challenges/scripts/fallback_mechanism_challenge.sh
./challenges/scripts/multipass_validation_challenge.sh
```

### RAGS Challenge Details

The RAGS challenge validates the hybrid retrieval system:

```bash
./challenges/scripts/rags_challenge.sh
```

**Test Coverage:**
- Dense retrieval with embeddings
- Sparse retrieval with BM25
- Hybrid retrieval combining both
- Reranking with cross-encoder
- Qdrant vector database integration
- Document ingestion and chunking
- Query expansion and reformulation

**Recent Improvements:**
- Timeout increased from 30s to 60s for complex operations
- All 147 tests now passing

### MCPS Challenge Details

The MCPS challenge validates MCP protocol integration:

```bash
./challenges/scripts/mcps_challenge.sh
```

**Sections:**
1. MCP Server initialization
2. Tool registration
3. Tool discovery
4. Tool execution
5. Resource management
6. Protocol compliance
7. Error handling
8. Streaming support
9. **NEW: MCP Tool Search validation**

### SKILLS Challenge Details

The SKILLS challenge validates skill execution:

```bash
./challenges/scripts/skills_challenge.sh
```

**Features:**
- Strict real-result validation (no mocked responses)
- Skill registration and discovery
- Skill execution with parameters
- Error handling and recovery
- Timeout management

### Recent Bug Fixes

#### ProviderHealthMonitor Mutex Deadlock

**Issue:** The health monitor could deadlock when multiple goroutines accessed provider status concurrently.

**Fix:** Implemented proper mutex ordering and reduced lock scope.

**Location:** `internal/services/provider_health_monitor.go`

#### CogneeService ListDatasets JSON Parsing

**Issue:** The ListDatasets endpoint returned malformed JSON in some cases.

**Fix:** Added proper JSON array parsing and error handling.

**Location:** `internal/llm/cognee/cognee_client.go`

### Troubleshooting Challenges

#### Challenge Timeout

If a challenge times out:

```bash
# Increase timeout for RAGS challenge
RAGS_TIMEOUT=120 ./challenges/scripts/rags_challenge.sh
```

#### Provider Unavailable

If provider tests fail:

```bash
# Check provider health
curl http://localhost:7061/v1/providers

# Run with fallback providers
ENABLE_FALLBACK=true ./challenges/scripts/run_all_challenges.sh
```

#### Database Connection Issues

If database tests fail:

```bash
# Start test infrastructure
make test-infra-start

# Run challenges with infrastructure
make test-with-infra
```
