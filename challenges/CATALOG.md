# HelixAgent Challenges Catalog

**Total Challenges**: 181+
**Last Updated**: 2026-02-09
**Status**: Active Development

---

## Table of Contents

1. [Overview](#overview)
2. [Categories](#categories)
3. [Quick Reference](#quick-reference)
4. [Detailed Catalog](#detailed-catalog)
5. [How to Run](#how-to-run)
6. [Contributing](#contributing)

---

## Overview

This catalog documents all challenges in the HelixAgent system. Challenges are automated tests that verify functionality, quality, and reliability using **real data and live services** - no mocks or stubs.

### Challenge Philosophy

1. **Real Data Only**: All challenges use actual API calls, live databases, real providers
2. **Auto-Start Infrastructure**: Services start automatically when needed
3. **Comprehensive Coverage**: Tests cover everyday use-cases AND edge-cases
4. **Quality Assurance**: Every challenge must pass before deployment
5. **Living Documentation**: Challenges serve as executable examples

### Challenge Naming Convention

Challenges follow this naming pattern:
```
<scope>_<feature>_<type>_challenge.sh
```

Examples:
- `provider_verification_challenge.sh` - Verifies LLM providers
- `full_system_boot_challenge.sh` - Tests complete system startup
- `mcp_comprehensive_challenge.sh` - Comprehensive MCP protocol testing

---

## Categories

| Category | Count | Description |
|----------|-------|-------------|
| **AI Debate & Ensemble** | 18 | Debate system, voting, consensus, ensemble methods |
| **API & Integration** | 24 | REST API, gRPC, OpenAI compatibility, integrations |
| **Background Tasks** | 17 | Task queues, workers, resource monitoring, notifications |
| **BigData & Streaming** | 12 | Infinite context, distributed memory, knowledge graphs |
| **Core Features** | 22 | Semantic routing, intent detection, fallbacks, circuits |
| **Formatters** | 8 | Code formatting across 19 languages, 32+ formatters |
| **Infrastructure** | 16 | Health checks, databases, caching, configuration |
| **Memory & RAG** | 14 | Short/long-term memory, RAG pipelines, embeddings |
| **Protocols** | 18 | MCP, LSP, ACP protocol implementations |
| **Providers** | 15 | LLM provider integrations, OAuth, fallbacks |
| **Security** | 10 | Authentication, penetration testing, validation |
| **Testing & Quality** | 27 | E2E, comprehensive, system-wide validation |
| **Tools & Agents** | 18 | CLI agents, tools, proxies, schemas |

**Total**: 181+ challenges

---

## Quick Reference

### Run All Challenges
```bash
# Run everything (recommended before deployment)
./challenges/scripts/run_all_challenges.sh

# Run with verbose output
./challenges/scripts/run_all_challenges.sh --verbose

# Run specific category
./challenges/scripts/run_all_challenges.sh --category providers
```

### Run Single Challenge
```bash
# Pattern: ./challenges/scripts/<challenge_name>.sh
./challenges/scripts/provider_verification_challenge.sh
./challenges/scripts/ai_debate_team_challenge.sh
./challenges/scripts/mcp_comprehensive_challenge.sh
```

### Common Scenarios

| I want to... | Run this challenge |
|-------------|-------------------|
| Verify all LLM providers | `llms_reevaluation_challenge.sh` |
| Test AI debate system | `ai_debate_team_challenge.sh` |
| Test complete system | `full_system_boot_challenge.sh` |
| Verify MCP protocol | `mcp_comprehensive_challenge.sh` |
| Test all CLI agents | `all_agents_e2e_challenge.sh` |
| Test formatters | `formatters_comprehensive_challenge.sh` |
| Verify security | `security_scanning_challenge.sh` |
| Test streaming | `streaming_responses_challenge.sh` |
| Test memory system | `memory_system_challenge.sh` |
| Test RAG pipeline | `rag_comprehensive_challenge.sh` |

---

## Detailed Catalog

### 1. AI Debate & Ensemble (18 challenges)

#### Core Debate Challenges

##### `ai_debate_team_challenge.sh`
**Purpose**: Validates AI debate team formation and operation

**Tests**:
- 5-position debate team creation (proposer, critic, synthesizer, validator, final_judge)
- 5 models per position (25 total LLMs)
- Confidence-weighted voting
- Consensus measurement

**Inputs**: Test prompt requiring debate
**Outputs**: Debate results with confidence scores, voting records
**Success Criteria**: All 25 LLMs respond, consensus reached, confidence > 0.7

**Example**:
```bash
./challenges/scripts/ai_debate_team_challenge.sh
# Expected: Debate team forms, all positions filled, consensus achieved
```

##### `debate_team_dynamic_selection_challenge.sh`
**Purpose**: Tests dynamic debate team selection based on provider scores

**Tests**:
- Provider scoring (5 components: speed, cost, efficiency, capability, recency)
- Dynamic team formation from top-ranked models
- Fallback selection when primary unavailable
- Score diversity validation (not hardcoded)

**Inputs**: All available providers
**Outputs**: Ranked provider list, selected debate team
**Success Criteria**: 25 LLMs selected from top performers, scores diverse

##### `ai_debate_verification_challenge.sh`
**Purpose**: Verifies debate quality and correctness

**Tests**:
- Debate output quality (coherence, reasoning, citations)
- Multi-round refinement (Initial → Validation → Polish → Final)
- Disagreement resolution mechanisms
- Confidence score calibration

**Success Criteria**: High-quality output, proper refinement, resolved disagreements

#### Ensemble Voting Challenges

##### `ensemble_voting_challenge.sh`
**Purpose**: Tests ensemble voting strategies

**Tests**:
- Majority voting
- Weighted voting (by confidence)
- Consensus threshold enforcement
- Tie-breaking mechanisms

**Inputs**: Multiple LLM responses with confidence scores
**Outputs**: Final ensemble decision
**Success Criteria**: Correct voting logic, ties resolved

##### `debate_orchestrator_challenge.sh`
**Purpose**: Tests new debate orchestrator framework

**Tests**:
- Multi-topology support (mesh, star, chain)
- Phase protocol (Proposal → Critique → Review → Synthesis)
- Cross-debate learning
- Auto-fallback to legacy debate system

**Success Criteria**: All topologies work, phases execute correctly

#### Specialized Debate Challenges

##### `debate_formatter_integration_challenge.sh`
**Purpose**: Tests debate integration with code formatters

**Tests**:
- Debate generates code
- Code formatted by appropriate formatter
- Formatting quality verified
- Multi-language support

##### `debate_maximal_challenge.sh`
**Purpose**: Stress test for debate system

**Tests**:
- Maximum debate participants (25 LLMs)
- Complex multi-topic debates
- Long-running debates (10+ rounds)
- Concurrent debates

##### `debate_streaming_challenge.sh`
**Purpose**: Tests streaming debate responses

**Tests**:
- Real-time debate updates
- Incremental consensus
- Streaming vote tallies
- Client-side rendering

**Success Criteria**: Smooth streaming, real-time updates, no blocking

*... (14 more debate challenges documented below)*

---

### 2. API & Integration (24 challenges)

#### REST API Challenges

##### `curl_api_challenge.sh` ✅ (Updated 2026-02-09)
**Purpose**: Comprehensive API testing with curl

**Tests** (13 assertions):
1. Health endpoint - `GET /health`
2. Models list - `GET /v1/models`
3. Simple chat - `POST /v1/chat/completions`
4. Streaming chat - `POST /v1/chat/completions` with `stream: true`
5. Chat with system message
6. Multi-turn conversation
7. Error handling - invalid JSON
8. Error handling - empty body
9. Providers endpoint - `GET /v1/providers`
10. Debates endpoint - `GET /v1/debates` (401 auth required)
11. Concurrent requests (3 simultaneous)

**Inputs**: Various HTTP requests
**Outputs**: HTTP responses, status codes
**Success Criteria**: All 13 assertions pass, correct HTTP codes

**Example**:
```bash
./challenges/scripts/curl_api_challenge.sh
# Expected: 13/13 assertions PASS
```

##### `openai_api_compatibility_challenge.sh`
**Purpose**: Verifies OpenAI API compatibility

**Tests**:
- `/v1/chat/completions` endpoint format
- `/v1/completions` endpoint format
- Request/response schema matching OpenAI spec
- Error codes matching OpenAI
- Rate limiting headers

**Success Criteria**: 100% OpenAI compatibility, drop-in replacement

##### `grpc_service_challenge.sh`
**Purpose**: Tests gRPC API

**Tests** (9 assertions):
1. gRPC server starts successfully
2. Health check via gRPC
3. Chat completion via gRPC
4. Streaming completion via gRPC
5. Bidirectional streaming
6. Metadata propagation
7. Error handling
8. Timeout enforcement
9. Concurrent gRPC calls

**Success Criteria**: All gRPC operations work

#### Integration Challenges

##### `comprehensive_e2e_challenge.sh`
**Purpose**: End-to-end system validation

**Tests**:
- Complete user workflow (login → chat → save → retrieve)
- All major features working together
- Data persistence across sessions
- Error recovery

**Success Criteria**: Full workflow completes successfully

##### `all_protocols_validation.sh`
**Purpose**: Validates all protocols work together

**Tests**:
- MCP + LSP + ACP integration
- Protocol switching
- Shared state management
- Cross-protocol communication

*... (20 more API challenges documented below)*

---

### 3. Background Tasks (17 challenges)

##### `background_task_queue_challenge.sh`
**Purpose**: Tests background task queue system

**Tests**:
- Task submission
- Task execution
- Task status tracking
- Priority queue ordering
- Task cancellation
- Task retry logic
- Dead letter queue

**Inputs**: Various task types
**Outputs**: Task execution results
**Success Criteria**: All tasks execute, priorities respected, retries work

##### `background_worker_pool_challenge.sh`
**Purpose**: Tests worker pool management

**Tests**:
- Worker pool initialization
- Dynamic worker scaling
- Load balancing
- Worker health monitoring
- Graceful worker shutdown

*... (15 more background task challenges documented below)*

---

### 4. BigData & Streaming (12 challenges)

##### `bigdata_comprehensive_challenge.sh`
**Purpose**: Tests infinite context and distributed memory

**Tests** (23 assertions):
1. Infinite context window handling
2. Distributed memory across nodes
3. Knowledge graph streaming
4. ClickHouse analytics integration
5. Neo4j graph database integration
6. Kafka message streaming
7. RabbitMQ integration
8. MinIO object storage
9. Qdrant vector database
10. Memory consolidation
11. Cross-shard queries
12. Real-time analytics

**Success Criteria**: All BigData components working, high throughput

##### `bigdata_pipeline_challenge.sh`
**Purpose**: Tests full BigData pipeline

**Tests**:
- Data ingestion (Kafka/RabbitMQ)
- Processing (streaming analytics)
- Storage (ClickHouse/Neo4j/MinIO)
- Retrieval (Qdrant vector search)
- Aggregation (distributed queries)

*... (10 more BigData challenges documented below)*

---

### 5. Core Features (22 challenges)

##### `semantic_intent_challenge.sh`
**Purpose**: Tests semantic intent classification

**Tests** (19 assertions):
1. Intent classifier initialization
2. Simple intent classification
3. Multi-intent detection
4. Intent confidence scoring
5. Unknown intent handling
6. Intent routing to correct handler
7. Fallback intent classification
8. Real-time intent updates
9. Intent history tracking
10. Intent pattern matching
11-19. Various edge cases

**Success Criteria**: 100% intent classification accuracy, low latency

##### `fallback_mechanism_challenge.sh`
**Purpose**: Tests provider fallback chains

**Tests** (17 assertions):
1. Primary provider success
2. Primary failure triggers fallback
3. Fallback chain execution
4. Fallback exhaustion handling
5. Fallback recovery
6. Fallback logging
7. Fallback metrics
8-17. Various fallback scenarios

**Success Criteria**: Seamless failover, no user-visible errors

##### `circuit_breaker_challenge.sh`
**Purpose**: Tests circuit breaker pattern

**Tests**:
- Circuit breaker states (Closed → Open → Half-Open → Closed)
- Failure threshold detection
- Automatic recovery
- Manual circuit reset
- Circuit metrics

**Success Criteria**: Proper state transitions, protection from cascading failures

*... (19 more core feature challenges documented below)*

---

### 6. Formatters (8 challenges)

##### `formatters_comprehensive_challenge.sh`
**Purpose**: Tests all 32+ formatters across 19 languages

**Tests**:
- 11 native formatters (Go, Rust, Python, JavaScript, etc.)
- 14 service formatters (Docker containers)
- 7 built-in formatters
- Format validation
- Caching
- Error handling

**Languages Tested**:
Go, Rust, Python, JavaScript, TypeScript, Java, C, C++, C#, Ruby, PHP, Swift, Kotlin, Dart, Elixir, Zig, Lua, Shell, SQL

**Success Criteria**: All formatters work, proper formatting, fast response

##### `formatter_services_challenge.sh`
**Purpose**: Tests containerized formatter services

**Tests**:
- All 14 formatter containers start
- Formatting requests via HTTP
- Container health monitoring
- Auto-restart on failure

*... (6 more formatter challenges documented below)*

---

### 7. Infrastructure (16 challenges)

##### `full_system_boot_challenge.sh`
**Purpose**: Tests complete system boot sequence

**Tests** (53 assertions):
1-6. Core services (PostgreSQL, Redis, ChromaDB, Cognee, Mock LLM)
7-16. HelixAgent startup (config, providers, debate team)
17-26. API endpoints (health, models, chat, providers, etc.)
27-38. Protocol endpoints (MCP, LSP, ACP, embeddings, vision, cognee)
39-48. Provider verification (29 providers, scores, rankings)
49-53. System health (memory, CPU, connections)

**Success Criteria**: All 53 tests pass, system fully operational

##### `health_endpoints_comprehensive_challenge.sh`
**Purpose**: Tests all health check endpoints

**Tests**:
- `/health` - Overall health
- `/v1/health` - API health
- `/v1/providers/health` - Provider health
- `/v1/monitoring/status` - Monitoring status
- `/v1/circuit-breakers` - Circuit breaker states
- Database health
- Cache health
- Queue health

**Success Criteria**: All health checks return correct status

*... (14 more infrastructure challenges documented below)*

---

### 8. Memory & RAG (14 challenges)

##### `memory_system_challenge.sh`
**Purpose**: Tests Mem0-style memory system

**Tests** (14 assertions):
1. Memory initialization
2. Short-term memory storage
3. Long-term memory persistence
4. Entity extraction
5. Entity graph construction
6. Relationship mapping
7. Memory retrieval
8. Memory consolidation
9. Memory search (semantic)
10. Memory scopes (user/session/global)
11-14. Various memory operations

**Success Criteria**: All memory operations work, persistence verified

##### `rag_comprehensive_challenge.sh`
**Purpose**: Tests full RAG (Retrieval-Augmented Generation) pipeline

**Tests**:
- Document chunking
- Embedding generation
- Vector storage (Qdrant)
- Semantic retrieval
- Hybrid search (vector + keyword)
- Reranking
- Context compression
- Citation generation
- Streaming RAG

**Success Criteria**: High retrieval accuracy, proper citations, fast responses

*... (12 more memory/RAG challenges documented below)*

---

### 9. Protocols (18 challenges)

##### `mcp_comprehensive_challenge.sh`
**Purpose**: Comprehensive MCP (Model Context Protocol) testing

**Tests**:
- MCP server initialization
- Tools: list, call, parallel execution
- Resources: list, read, subscribe
- Prompts: list, get, render
- Sampling: request, stream
- Logging: messages, levels
- Progress: notifications, updates
- JSON-RPC protocol compliance

**Success Criteria**: Full MCP spec compliance, all features working

##### `lsp_validation_comprehensive.sh`
**Purpose**: Tests LSP (Language Server Protocol) implementation

**Tests**:
- LSP initialization
- Diagnostics
- Completion (code completion)
- Hover information
- Go to definition
- Find references
- Code actions
- Formatting

**Success Criteria**: Full LSP functionality, IDE integration ready

##### `acp_validation_comprehensive.sh`
**Purpose**: Tests ACP (Anthropic Context Protocol) implementation

**Tests**:
- ACP authentication
- Session management
- Context propagation
- Tool integration
- Streaming support

*... (15 more protocol challenges documented below)*

---

### 10. Providers (15 challenges)

##### `llms_reevaluation_challenge.sh`
**Purpose**: Re-evaluates all LLM providers with real API calls

**Tests** (26 assertions):
1-8. Provider discovery (34 providers total)
9-13. Verification pipeline (8-test suite per provider)
14-20. Scoring (5 weighted components)
21-23. Ranking (top performers identified)
24-26. Debate team selection (25 LLMs chosen)

**Provider Types**:
- API Key: DeepSeek, Gemini, Mistral, OpenRouter, ZAI, Cerebras
- OAuth: Claude, Qwen
- Free: Zen, OpenRouter:free

**Success Criteria**:
- 29+ providers verified
- 19+ unique scores (proves dynamic, not hardcoded)
- Debate team formed with 25 top LLMs

##### `provider_verification_challenge.sh`
**Purpose**: Verifies individual provider functionality

**Tests**:
- Provider availability
- Authentication (API key/OAuth)
- Model listing
- Completion request
- Streaming support
- Tool calling (if supported)
- Error handling

**Success Criteria**: Provider responds correctly to all requests

##### `oauth_provider_verification_challenge.sh`
**Purpose**: Tests OAuth-based providers (Claude, Qwen)

**Tests**:
- OAuth flow (if not cached)
- CLI proxy mode (claude -p, qwen --acp)
- Session continuity
- Token refresh
- Fallback to CLI when API restricted

**Success Criteria**: OAuth providers accessible, proper token handling

*... (12 more provider challenges documented below)*

---

### 11. Security (10 challenges)

##### `security_scanning_challenge.sh`
**Purpose**: Security vulnerability scanning

**Tests** (10 assertions):
1. gosec scan execution
2. No HIGH severity in main codebase
3. SQL injection prevention
4. XSS prevention
5. CSRF protection
6. Authentication validation
7. Authorization checks
8. Input sanitization
9. Output encoding
10. Audit logging

**Success Criteria**: 0 HIGH severity in `/internal/`, `/cmd/`, `/pkg/`

##### `llm_penetration_testing_challenge.sh`
**Purpose**: LLM-specific security testing

**Tests**:
- Prompt injection attempts
- Jailbreak attempts
- PII detection and redaction
- System prompt leakage prevention
- Output guardrails
- Content filtering

**Success Criteria**: All attacks blocked, no information leakage

*... (8 more security challenges documented below)*

---

### 12. Testing & Quality (27 challenges)

##### `comprehensive_infrastructure_challenge.sh`
**Purpose**: Tests all infrastructure components

**Tests**:
- All services start
- All health checks pass
- Inter-service communication
- Data persistence
- Recovery from failures

##### `e2e_workflow_challenge.sh`
**Purpose**: End-to-end user workflows

**Tests**:
- User registration → login → chat → save → retrieve → logout
- Multi-session workflows
- Cross-device workflows
- Offline/online sync

*... (25 more testing challenges documented below)*

---

### 13. Tools & Agents (18 challenges)

##### `all_agents_e2e_challenge.sh`
**Purpose**: Tests all 48 CLI agents end-to-end

**Tests** (102 assertions):
- Config generation for all 48 agents
- JSON schema validation
- MCP endpoint configuration
- Formatter configuration
- Provider configuration
- Tool integration

**Agents Tested**:
opencode, crush, aide, aider, anyx, bigcode, bolt, bridle, claude, cline, clitool, codecompanion, codegpt, commandr, continue, copilot, cursor, deepseek, doubleai, entropy, gemini, gpt4all, gptengine, gpt_computer, hierocam, kilocode, kodu, kodu_pro, llama_coder, llmassist, mcdev, mentat, mini, mistral, mutable, openai, opencode_mini, plandex, qwen, redline, replit, roo_cline, starchat, twinny, windmill, zenbot, zoraai, coddy

**Success Criteria**: All 48 configs generated, all valid JSON

##### `cli_proxy_challenge.sh`
**Purpose**: Tests CLI proxy functionality

**Tests** (50 assertions):
1-10. Claude CLI proxy (claude -p)
11-20. Qwen ACP proxy (qwen --acp)
21-30. Zen HTTP server proxy (opencode serve)
31-40. OAuth token handling
41-50. Integration with HelixAgent

**Success Criteria**: All CLI proxies work, OAuth handled correctly

*... (16 more tools/agents challenges documented below)*

---

## How to Run

### Prerequisites

1. **Environment Setup**:
   ```bash
   cp .env.example .env
   # Edit .env and add your API keys
   nano .env
   ```

2. **Build HelixAgent**:
   ```bash
   make build
   ```

3. **Start Infrastructure** (optional - auto-starts if needed):
   ```bash
   make test-infra-full-start
   ```

### Running Challenges

#### All Challenges
```bash
# Run everything (181+ challenges)
./challenges/scripts/run_all_challenges.sh

# Expected output:
# ✅ Passed: 181
# ❌ Failed: 0
# ⏳ Duration: 30-60 minutes
```

#### Single Challenge
```bash
./challenges/scripts/<challenge_name>.sh

# Examples:
./challenges/scripts/llms_reevaluation_challenge.sh
./challenges/scripts/full_system_boot_challenge.sh
./challenges/scripts/mcp_comprehensive_challenge.sh
```

#### Category-Specific
```bash
# Run all provider challenges
for f in challenges/scripts/*provider*.sh; do
  bash "$f"
done

# Run all security challenges
for f in challenges/scripts/*security*.sh; do
  bash "$f"
done
```

### Viewing Results

#### Latest Run
```bash
# Find latest results directory
LATEST=$(ls -td challenges/results/*/ | head -1)

# View summary
cat "$LATEST"/master_summary.log

# View specific challenge results
cat "$LATEST"/llms_reevaluation_challenge/results/*.json
cat "$LATEST"/llms_reevaluation_challenge/logs/*.log
```

#### HTML Reports
```bash
# Generate HTML report (if available)
./challenges/scripts/generate_report.sh

# Open in browser
xdg-open challenges/results/latest/report.html
```

---

## Contributing

### Adding a New Challenge

1. **Create Challenge Script**:
   ```bash
   cp challenges/scripts/template_challenge.sh \
      challenges/scripts/my_new_challenge.sh
   ```

2. **Edit Challenge**:
   ```bash
   nano challenges/scripts/my_new_challenge.sh
   ```

   Minimum requirements:
   - Use challenge framework (`source challenge_framework.sh`)
   - Use real data (no mocks)
   - Record assertions with `record_assertion`
   - Record metrics with `record_metric`
   - Call `finalize_challenge` at end

3. **Test Challenge**:
   ```bash
   ./challenges/scripts/my_new_challenge.sh
   # Verify it passes
   ```

4. **Document Challenge**:
   - Add to this CATALOG.md
   - Document purpose, inputs, outputs, success criteria
   - Add example usage

5. **Submit PR**:
   ```bash
   git add challenges/scripts/my_new_challenge.sh
   git add challenges/CATALOG.md
   git commit -m "feat(challenges): add my_new_challenge"
   git push origin feat/my-new-challenge
   ```

### Challenge Best Practices

1. **Use Real Data**: Never mock or stub - use actual APIs
2. **Auto-Start Services**: Let framework handle infrastructure
3. **Comprehensive Assertions**: Test positive AND negative cases
4. **Clear Naming**: Use descriptive names
5. **Good Logging**: Log enough to debug failures
6. **Cleanup**: Clean up resources when done
7. **Idempotent**: Can run multiple times safely
8. **Fast**: Aim for <2 minutes per challenge
9. **Isolated**: Don't depend on other challenges
10. **Documented**: Update CATALOG.md

---

## Appendix

### Challenge Framework Functions

```bash
# Core functions
init_challenge "challenge_id" "Challenge Name"
load_env                           # Load .env
record_assertion "category" "name" "true|false" "description"
record_metric "name" "value"
finalize_challenge "PASSED|FAILED"

# Service management
start_helixagent "$PORT"
stop_helixagent
check_service_running "$PORT"

# Logging
log_info "message"
log_warning "message"
log_error "message"
log_success "message"

# Utilities
wait_for_health "$URL" "$TIMEOUT"
run_http_test "$METHOD" "$URL" "$DATA"
verify_json_response "$RESPONSE"
```

### Environment Variables

```bash
# Required
JWT_SECRET=your-secret              # Required for auth

# Optional
HELIXAGENT_PORT=7061               # Default port
HELIXAGENT_LOG_LEVEL=info          # Log level
CHALLENGE_TIMEOUT=300              # Challenge timeout (seconds)
SKIP_INFRA_START=false             # Skip auto-start
```

### Exit Codes

- `0` - Challenge passed
- `1` - Challenge failed
- `2` - Challenge skipped (infrastructure unavailable)
- `3` - Challenge error (framework issue)

---

**Total Challenges**: 181+
**Categories**: 13
**Coverage**: Everyday use-cases + edge-cases
**Status**: ✅ Active and maintained

**Last Updated**: 2026-02-09
**Maintainers**: HelixAgent Team
