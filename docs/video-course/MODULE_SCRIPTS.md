# HelixAgent Video Course - Module Scripts Outline

This document provides comprehensive script outlines for all 74 videos across 14 modules. Each outline includes talking points, demonstrations, and example commands.

---

## Module 1: Introduction to HelixAgent (45 minutes)

### Video 1.1: Course Welcome and Learning Path (8 min)

**Opening Hook**:
- "What if you could harness the power of 10 different AI models with a single API call?"

**Talking Points**:
- Welcome to the HelixAgent course
- Instructor introduction and background
- Course structure overview (14 modules, 5 certification levels)
- How to navigate the course materials
- Prerequisites review:
  - Basic programming knowledge
  - Familiarity with REST APIs
  - Go 1.24+ for development modules
- What you'll build by the end
- How to get help and support

**Visual Aids**:
- Course roadmap diagram
- Certification path overview
- Module timeline

**Call to Action**:
- Set up your development environment
- Join the community Discord

---

### Video 1.2: What is HelixAgent? (12 min)

**Opening Hook**:
- Start with a common problem: vendor lock-in, single points of failure

**Talking Points**:
- Definition: AI-powered ensemble LLM service
- Core value proposition:
  - Multi-provider orchestration
  - Intelligent aggregation strategies
  - OpenAI-compatible APIs
- 10 supported providers:
  - Claude (Anthropic) - Reasoning, Analysis
  - DeepSeek - Code, Technical Content
  - Gemini (Google) - Multimodal, Scientific
  - Qwen (Alibaba) - Multilingual
  - Ollama - Local/Private Deployment
  - OpenRouter - Meta-Provider Access
  - ZAI - Specialized Tasks
  - Zen - Alternative Provider
  - Mistral - European AI
  - Cerebras - High Performance
- Comparison with single-provider approaches
- Real-world scenarios where HelixAgent excels

**Demo**:
```bash
# Show a simple API call
curl -X POST http://localhost:7061/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"helixagent-ensemble","messages":[{"role":"user","content":"Hello"}]}'
```

**Visual Aids**:
- Hub and spoke architecture diagram
- Provider comparison matrix
- Before/after workflow diagrams

---

### Video 1.3: Architecture Overview (15 min)

**Talking Points**:
- High-level system architecture
- Core components explained:
  - **API Gateway**: Request routing, authentication
  - **Ensemble Engine**: Multi-provider orchestration
  - **AI Debate Service**: Multi-agent discussions
  - **Provider Registry**: Unified provider interface
  - **Plugin System**: Extensibility
  - **Cache Layer**: Redis + in-memory
  - **Monitoring**: Prometheus, Grafana
- Internal package structure:
  - `internal/llm/` - Provider abstractions
  - `internal/services/` - Business logic
  - `internal/handlers/` - HTTP handlers
  - `internal/middleware/` - Auth, rate limiting
  - `internal/debate/` - Orchestrator framework
- Data flow walkthrough
- LLMProvider interface introduction:
  ```go
  type LLMProvider interface {
      Complete(ctx, request) (*Response, error)
      CompleteStream(ctx, request) (chan StreamChunk, error)
      HealthCheck(ctx) error
      GetCapabilities() *Capabilities
      ValidateConfig() error
  }
  ```

**Visual Aids**:
- Architecture diagram
- Component interaction flow
- Package structure tree

**Demo**:
- Walk through key directories in the repository
- Show configuration files

---

### Video 1.4: Use Cases and Applications (10 min)

**Talking Points**:
- Enterprise use cases:
  - **Content Generation**: Multi-model review for quality
  - **Code Analysis**: Cross-provider code review
  - **Research**: Multiple perspectives on complex topics
  - **Customer Support**: Intelligent routing by topic
  - **Decision Support**: AI debate for recommendations
  - **Translation**: Multi-provider verification
- Industry applications:
  - Financial services (risk assessment)
  - Healthcare (diagnostic support)
  - Legal (document analysis)
  - Education (personalized learning)
- Performance characteristics:
  - Horizontal scaling
  - Semantic caching
  - Circuit breaker patterns
- Security features overview

**Demo**:
- Show a complete request flow from API to response
- Demonstrate provider health monitoring

---

## Module 2: Installation and Setup (60 minutes)

### Video 2.1: Environment Prerequisites (10 min)

**Talking Points**:
- System requirements:
  - Linux, macOS, or Windows (WSL2)
  - 8GB RAM minimum (16GB recommended)
  - 20GB disk space
- Required software:
  - Go 1.24+ installation
  - Docker and Docker Compose
  - Git
- IDE recommendations:
  - VS Code with Go extension
  - GoLand
  - Vim/Neovim with gopls

**Demo**:
```bash
# Verify Go installation
go version

# Verify Docker
docker --version
docker-compose --version

# Clone repository
git clone git@github.com:your-org/helix-agent.git
cd helix-agent
```

---

### Video 2.2: Docker Quick Start (15 min)

**Talking Points**:
- Docker-based installation advantages
- Docker Compose configuration structure
- Available profiles:
  - Core services (PostgreSQL, Redis)
  - AI services (Ollama)
  - Monitoring (Prometheus, Grafana)
  - Full stack

**Demo**:
```bash
# Start core infrastructure
make infra-core

# Start full stack
docker-compose up -d

# Verify services
docker-compose ps

# Check health
curl http://localhost:7061/health | jq

# View logs
docker-compose logs -f helixagent
```

**Visual Aids**:
- Docker architecture diagram
- Service dependency graph

---

### Video 2.3: Manual Installation from Source (15 min)

**Talking Points**:
- When to use source installation
- Build process explained
- Configuration options

**Demo**:
```bash
# Install dependencies
make install-deps

# Build binary
make build

# Build with debug symbols
make build-debug

# Run locally
make run

# Run in development mode
make run-dev

# Verify API
curl http://localhost:7061/health
curl http://localhost:7061/v1/models | jq
```

---

### Video 2.4: Podman Alternative Setup (8 min)

**Talking Points**:
- Podman vs Docker
- When to use Podman
- Configuration differences

**Demo**:
```bash
# Using container runtime script
./scripts/container-runtime.sh

# Podman-specific commands
podman-compose up -d
podman ps
```

---

### Video 2.5: Troubleshooting Installation Issues (12 min)

**Talking Points**:
- Common installation problems:
  - Port conflicts
  - Docker daemon issues
  - Go version mismatches
  - API key problems
- Diagnostic commands
- Log analysis

**Demo**:
```bash
# Check port availability
lsof -i :7061

# Check Docker
docker info

# View detailed logs
make run-dev 2>&1 | tee debug.log

# Test database connection
PGPASSWORD=helixagent123 psql -h localhost -p 15432 -U helixagent -d helixagent_db -c "SELECT 1"
```

---

## Module 3: Configuration (60 minutes)

### Video 3.1: Configuration Architecture (12 min)

**Talking Points**:
- Configuration file hierarchy
- Environment variable precedence
- Configuration loading order:
  1. Default values
  2. Configuration files
  3. Environment variables
- Secrets management best practices

**Visual Aids**:
- Configuration hierarchy diagram
- File structure overview

---

### Video 3.2: Core Configuration Options (15 min)

**Talking Points**:
- Server configuration:
  - PORT, GIN_MODE, JWT_SECRET
- Database settings:
  - PostgreSQL connection parameters
  - Connection pooling
- Redis configuration:
  - Cache settings
  - Session management
- Logging options

**Demo**:
```bash
# Show configuration files
cat configs/development.yaml
cat configs/production.yaml

# Environment variables
export PORT=7061
export GIN_MODE=debug
export JWT_SECRET=your-secret-key
```

---

### Video 3.3: Provider Configuration (12 min)

**Talking Points**:
- API key configuration for each provider:
  - CLAUDE_API_KEY
  - DEEPSEEK_API_KEY
  - GEMINI_API_KEY
  - QWEN_API_KEY
  - ZAI_API_KEY
  - OPENROUTER_API_KEY
  - MISTRAL_API_KEY
  - CEREBRAS_API_KEY
- OAuth credentials:
  - CLAUDE_USE_OAUTH_CREDENTIALS
  - QWEN_USE_OAUTH_CREDENTIALS
- Ollama local setup
- Rate limiting and timeouts

**Demo**:
```bash
# Create .env file
cat > .env << 'EOF'
CLAUDE_API_KEY=sk-ant-xxxxx
DEEPSEEK_API_KEY=sk-xxxxx
GEMINI_API_KEY=AIza-xxxxx
EOF

# Verify provider status
curl http://localhost:7061/v1/providers/status | jq
```

---

### Video 3.4: Advanced Configuration (12 min)

**Talking Points**:
- AI Debate configuration
- Cognee integration settings
- BigData components:
  - BIGDATA_ENABLE_* environment variables
- Service overrides:
  - SVC_<SERVICE>_<FIELD> pattern
- Remote service configuration

**Demo**:
```bash
# Advanced configuration example
cat configs/multi-provider.yaml
```

---

### Video 3.5: Configuration Best Practices (9 min)

**Talking Points**:
- Environment-specific configurations
- Secrets management:
  - Never commit secrets
  - Use .env.example templates
  - Rotate API keys regularly
- Configuration validation
- Documentation practices

**Demo**:
```bash
# Validate configuration
make fmt vet lint

# Test configuration loading
go test -v ./internal/config/...
```

---

## Module 4: LLM Provider Integration (75 minutes)

### Video 4.1: Provider Interface Architecture (12 min)

**Talking Points**:
- LLMProvider interface deep dive:
  ```go
  type LLMProvider interface {
      Complete(ctx, request) (*Response, error)
      CompleteStream(ctx, request) (chan StreamChunk, error)
      HealthCheck(ctx) error
      GetCapabilities() *Capabilities
      ValidateConfig() error
  }
  ```
- Provider lifecycle management
- Health checking mechanisms
- Error handling patterns

**Visual Aids**:
- Interface diagram
- Provider state machine

---

### Video 4.2: Claude Integration (10 min)

**Talking Points**:
- Claude API configuration
- Model selection:
  - claude-3-opus
  - claude-3-sonnet
  - claude-3-haiku
  - claude-3.5-sonnet
- OAuth CLI proxy option
- Best use cases for Claude

**Demo**:
```bash
# Configure Claude
export CLAUDE_API_KEY=sk-ant-xxxxx

# Test Claude directly
curl -X POST http://localhost:7061/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"claude-3.5-sonnet","messages":[{"role":"user","content":"Explain recursion"}]}'
```

---

### Video 4.3: DeepSeek Integration (10 min)

**Talking Points**:
- DeepSeek API setup
- Available models:
  - deepseek-chat
  - deepseek-coder
- Technical content optimization
- Cost-effective processing strategies

**Demo**:
```bash
# Configure DeepSeek
export DEEPSEEK_API_KEY=sk-xxxxx

# Test code generation
curl -X POST http://localhost:7061/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"deepseek-coder","messages":[{"role":"user","content":"Write a Go function to sort integers"}]}'
```

---

### Video 4.4: Gemini Integration (10 min)

**Talking Points**:
- Google Gemini API configuration
- Model options:
  - gemini-pro
  - gemini-2.0-flash
- Multimodal capabilities
- GCP Vertex AI integration

**Demo**:
```bash
# Configure Gemini
export GEMINI_API_KEY=AIza-xxxxx

# Test multimodal request
curl -X POST http://localhost:7061/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"gemini-2.0-flash","messages":[{"role":"user","content":"Analyze this scientific concept"}]}'
```

---

### Video 4.5: Qwen, Ollama, and Other Providers (18 min)

**Talking Points**:
- Qwen API configuration (Alibaba Cloud)
- Ollama local model setup:
  - Installation
  - Model download
  - Performance tuning
- OpenRouter as meta-provider
- ZAI, Zen, Mistral, Cerebras overview
- Provider selection guidelines

**Demo**:
```bash
# Ollama setup
ollama pull llama2
ollama run llama2

# OpenRouter configuration
export OPENROUTER_API_KEY=sk-or-xxxxx

# Test multiple providers
for model in "claude-3.5-sonnet" "deepseek-chat" "gemini-2.0-flash"; do
  echo "Testing $model..."
  curl -s -X POST http://localhost:7061/v1/chat/completions \
    -H "Content-Type: application/json" \
    -d "{\"model\":\"$model\",\"messages\":[{\"role\":\"user\",\"content\":\"What is 2+2?\"}]}" | jq '.choices[0].message.content'
done
```

---

### Video 4.6: Building Fallback Chains (15 min)

**Talking Points**:
- Fallback chain design principles
- Primary and secondary provider selection
- Health-based routing
- Cost-optimized ordering
- Error categorization:
  - rate_limit
  - timeout
  - auth
  - connection
  - unavailable
  - overloaded

**Demo**:
```bash
# Configure fallback in YAML
cat configs/fallback-example.yaml

# Test fallback behavior
# Simulate primary failure and observe fallback
```

**Visual Aids**:
- Fallback chain diagram
- Error handling flow

---

## Module 5: Ensemble Strategies (60 minutes)

### Video 5.1: Introduction to Ensemble AI (12 min)

**Talking Points**:
- What is ensemble learning?
- Benefits of multi-model consensus:
  - Improved accuracy
  - Reduced bias
  - Error detection
- Diversity in AI models
- Real-world ensemble applications

**Visual Aids**:
- Ensemble concept diagram
- Accuracy comparison charts

---

### Video 5.2: Voting Strategies Explained (15 min)

**Talking Points**:
- Available voting strategies:
  - **Majority Voting**: Simple vote count
  - **Weighted Voting**: Provider weight-based
  - **Consensus Voting**: High agreement required
  - **Confidence-Weighted**: Based on confidence scores
  - **Quality-Weighted**: Based on quality metrics
- Strategy selection criteria
- Configuration examples

**Demo**:
```bash
# Configure ensemble strategy
cat configs/ensemble-example.yaml

# Test different strategies
curl -X POST http://localhost:7061/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"helixagent-ensemble","messages":[{"role":"user","content":"Explain quantum computing"}]}'
```

---

### Video 5.3: Implementing Custom Strategies (15 min)

**Talking Points**:
- VotingStrategy interface:
  ```go
  type VotingStrategy interface {
      Vote(responses []*Response) (*Response, error)
      Name() string
  }
  ```
- Creating custom algorithms
- Testing strategies
- Performance considerations

**Demo**:
- Walk through `internal/llm/ensemble.go`
- Show custom strategy implementation

---

### Video 5.4: Performance Optimization (10 min)

**Talking Points**:
- Parallel execution strategies
- Caching for ensemble responses
- Latency reduction techniques
- Cost optimization

**Demo**:
```bash
# Monitor performance
curl http://localhost:7061/metrics | grep ensemble
```

---

### Video 5.5: Ensemble Benchmarking (8 min)

**Talking Points**:
- Benchmark methodology
- Metrics to track
- Comparison analysis
- Tuning recommendations

**Demo**:
```bash
# Run benchmarks
make test-bench

# Analyze results
go test -bench=. ./internal/llm/...
```

---

## Module 6: AI Debate System (90 minutes)

### Video 6.1: AI Debate Concepts (15 min)

**Talking Points**:
- What is AI debate?
- Multi-agent discussion system
- Benefits over single-model responses:
  - Multiple viewpoints
  - Cross-validation
  - Dynamic discussion
  - Deeper reasoning
- Use cases: complex reasoning, decision making

**Visual Aids**:
- Debate architecture diagram
- Participant interaction flow

---

### Video 6.2: Participant Configuration (18 min)

**Talking Points**:
- Participant structure:
  ```yaml
  participants:
    - name: "Analyst"
      role: "Primary Analyst"
      enabled: true
      weight: 1.5
      priority: 1
      debate_style: analytical
      argumentation_style: logical
      llms:
        - provider: claude
          model: claude-3-opus
        - provider: deepseek
          model: deepseek-coder
  ```
- Debate styles:
  - analytical, creative, balanced
  - aggressive, diplomatic, technical
- Argumentation styles:
  - logical, emotional, evidence_based
  - hypothetical, socratic
- LLM fallback chains per participant

**Demo**:
```bash
# Show configuration
cat configs/ai-debate-example.yaml

# Explain each participant role
```

---

### Video 6.3: Debate Strategies Deep Dive (15 min)

**Talking Points**:
- Available strategies:
  - **round_robin**: Fixed turn order
  - **free_form**: Dynamic order
  - **structured**: Organized rounds
  - **adversarial**: Opposing views
  - **collaborative**: Consensus building
- Consensus threshold configuration
- Timeout management

**Demo**:
```bash
# Create debate with specific strategy
curl -X POST http://localhost:7061/v1/debates \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "Should AI development be regulated?",
    "strategy": "adversarial",
    "rounds": 3,
    "consensus_threshold": 0.8
  }'
```

---

### Video 6.4: Cognee AI Integration (12 min)

**Talking Points**:
- What Cognee provides:
  - Semantic enhancement
  - Contextual analysis
  - Knowledge integration
  - Memory across debates
- Configuration:
  ```yaml
  cognee_config:
    enabled: true
    enhance_responses: true
    analyze_consensus: true
    generate_insights: true
  ```

**Demo**:
```bash
# Enable Cognee
export COGNEE_ENABLED=true

# Test enhanced debate
```

---

### Video 6.5: Programmatic Debate Execution (15 min)

**Talking Points**:
- Using the Debate API
- Go SDK usage:
  ```go
  debateService, err := services.NewAIDebateService(cfg, nil, nil)
  result, err := debateService.ConductDebate(ctx, topic, context)
  ```
- Handling responses
- Interpreting consensus
- Error handling

**Demo**:
```bash
# Full debate via API
curl -X POST http://localhost:7061/v1/chat/completions \
  -H "Content-Type: application/json" \
  -N \
  -d '{
    "model": "helixagent-debate",
    "messages": [{"role": "user", "content": "What are the pros and cons of microservices?"}],
    "stream": true
  }'
```

---

### Video 6.6: Monitoring and Metrics (15 min)

**Talking Points**:
- Available metrics:
  - Total debates conducted
  - Success and consensus rates
  - Response times per participant
  - Provider usage
  - Quality score distribution
- Monitoring endpoints
- Grafana dashboard setup

**Demo**:
```bash
# Get debate metrics
curl http://localhost:7061/v1/debate/metrics | jq

# Prometheus metrics
curl http://localhost:7061/metrics | grep debate
```

---

## Module 7: Plugin Development (75 minutes)

### Video 7.1: Plugin Architecture Overview (12 min)

**Talking Points**:
- Plugin system design goals
- Hot-reloading mechanism
- Plugin lifecycle:
  - Discovery
  - Loading
  - Initialization
  - Execution
  - Unloading
- Plugin isolation

**Visual Aids**:
- Plugin architecture diagram
- Lifecycle state machine

---

### Video 7.2: Plugin Interfaces Deep Dive (15 min)

**Talking Points**:
- PluginRegistry interface
- PluginLoader interface
- Plugin metadata structure
- Health and metrics integration

**Demo**:
- Walk through `internal/plugins/` code
- Show interface definitions

---

### Video 7.3: Developing Your First Plugin (20 min)

**Talking Points**:
- Plugin project structure
- Implementing required interfaces
- Plugin configuration
- Building and packaging
- Testing locally

**Demo**:
```go
// Example plugin implementation
type MyPlugin struct {
    config *Config
}

func (p *MyPlugin) Init(cfg interface{}) error {
    // Initialize plugin
}

func (p *MyPlugin) Execute(ctx context.Context, req *Request) (*Response, error) {
    // Plugin logic
}
```

---

### Video 7.4: Advanced Plugin Topics (15 min)

**Talking Points**:
- Dependency resolution
- Plugin communication
- Error handling best practices
- Performance considerations
- Security constraints

---

### Video 7.5: Plugin Deployment and Testing (13 min)

**Talking Points**:
- Deployment strategies
- Hot-reload testing
- Production considerations
- Monitoring plugins

**Demo**:
```bash
# Deploy plugin
cp myplugin.so plugins/

# Verify loading
curl http://localhost:7061/v1/plugins | jq

# Test hot-reload
```

---

## Module 8: MCP/LSP Integration (60 minutes)

### Video 8.1: Protocol Support Overview (10 min)

**Talking Points**:
- Unified Protocol Manager
- Supported protocols:
  - MCP (Model Context Protocol)
  - LSP (Language Server Protocol)
  - ACP (Agent Client Protocol)
- Use cases for each

**Visual Aids**:
- Protocol interaction diagram
- Endpoint mapping

---

### Video 8.2: MCP Integration Deep Dive (15 min)

**Talking Points**:
- MCP server configuration
- 45+ available adapters
- Tool execution
- API endpoints

**Demo**:
```bash
# List MCP tools
curl http://localhost:7061/v1/mcp/tools | jq

# Execute tool
curl -X POST http://localhost:7061/v1/mcp/execute \
  -H "Content-Type: application/json" \
  -d '{"tool":"filesystem.read_file","arguments":{"path":"/etc/hostname"}}'
```

---

### Video 8.3: LSP Integration (12 min)

**Talking Points**:
- Language Server Protocol basics
- Configuration
- Code intelligence features
- IDE integration

**Demo**:
```bash
# LSP endpoint
curl http://localhost:7061/v1/lsp/status | jq
```

---

### Video 8.4: ACP and Embeddings (13 min)

**Talking Points**:
- Agent Client Protocol
- Agent communication patterns
- Embedding generation
- Vector operations
- 6 embedding providers

**Demo**:
```bash
# Generate embeddings
curl -X POST http://localhost:7061/v1/embeddings \
  -H "Content-Type: application/json" \
  -d '{"model":"text-embedding-ada-002","input":"Hello world"}'
```

---

### Video 8.5: Building Protocol Workflows (10 min)

**Talking Points**:
- Combining protocols
- Workflow design patterns
- Error handling
- Performance optimization

**Demo**:
- Build complete workflow using MCP + embeddings

---

## Module 9: Optimization Features (75 minutes)

### Video 9.1: Optimization Framework Overview (12 min)

**Talking Points**:
- 8 optimization tools integrated
- Native Go vs HTTP implementations
- Docker services for optimization
- Configuration overview

**Visual Aids**:
- Optimization stack diagram

---

### Video 9.2: Semantic Caching with GPTCache (15 min)

**Talking Points**:
- Vector similarity caching
- LRU eviction strategies
- TTL configuration
- Cache hit optimization

**Demo**:
```bash
# Configure caching
# Show cache hits in metrics
curl http://localhost:7061/metrics | grep cache
```

---

### Video 9.3: Structured Output with Outlines (12 min)

**Talking Points**:
- JSON schema validation
- Regex pattern constraints
- Choice constraints
- Ensuring format compliance

**Demo**:
```bash
# Request with schema
curl -X POST http://localhost:7061/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model":"helixagent-ensemble",
    "messages":[{"role":"user","content":"List 3 colors"}],
    "response_format":{"type":"json_object","schema":{"type":"array","items":{"type":"string"}}}
  }'
```

---

### Video 9.4: Enhanced Streaming (12 min)

**Talking Points**:
- Word and sentence buffering
- Progress tracking
- Rate limiting
- Real-time handling

**Demo**:
```bash
# Streaming request
curl -N -X POST http://localhost:7061/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"helixagent-debate","messages":[{"role":"user","content":"Write a story"}],"stream":true}'
```

---

### Video 9.5: Advanced Optimization (SGLang, LlamaIndex) (15 min)

**Talking Points**:
- RadixAttention prefix caching
- Document retrieval with HyDE
- Reranking strategies
- Cognee integration for RAG

**Demo**:
- Show RAG pipeline configuration
- Demonstrate reranking

---

### Video 9.6: Measuring Optimization Impact (9 min)

**Talking Points**:
- Benchmark methodology
- Metrics to track
- Before/after comparison
- Tuning recommendations

**Demo**:
```bash
# Run optimization benchmarks
make test-bench
```

---

## Module 10: Security Best Practices (60 minutes)

### Video 10.1: Security Architecture (12 min)

**Talking Points**:
- Security-first design
- Authentication mechanisms
- Authorization and RBAC
- Threat model

**Visual Aids**:
- Security architecture diagram

---

### Video 10.2: API Security Configuration (12 min)

**Talking Points**:
- JWT token configuration
- API key management
- Rate limiting implementation
- Input validation
- Request size limits

**Demo**:
```bash
# Configure JWT
export JWT_SECRET=your-secret-key

# Test authentication
curl -H "Authorization: Bearer $TOKEN" http://localhost:7061/v1/models
```

---

### Video 10.3: Secrets Management (12 min)

**Talking Points**:
- Environment variable best practices
- API key rotation
- Secure configuration storage
- Secrets in containers

**Demo**:
```bash
# Using secrets file
cat .env.example

# Never commit secrets
echo ".env" >> .gitignore
```

---

### Video 10.4: Production Security Hardening (15 min)

**Talking Points**:
- Network security
- Container security
- Database security
- Logging and audit trails

**Demo**:
```bash
# Security scan
make security-scan

# Review gosec output
```

---

### Video 10.5: Security Testing (9 min)

**Talking Points**:
- Security test suite
- Penetration testing
- Vulnerability scanning
- Compliance verification

**Demo**:
```bash
# Run security tests
make test-security
```

---

## Module 11: Testing and CI/CD (75 minutes)

### Video 11.1: Testing Strategy Overview (12 min)

**Talking Points**:
- Test pyramid
- Test types:
  - Unit, Integration, E2E
  - Security, Stress, Chaos
- Coverage targets

**Visual Aids**:
- Test pyramid diagram

---

### Video 11.2: Running All Test Types (18 min)

**Talking Points**:
- Make commands for testing
- Test infrastructure setup
- Docker dependencies

**Demo**:
```bash
# All tests
make test

# Specific test types
make test-unit
make test-integration
make test-e2e
make test-security
make test-stress
make test-chaos
make test-bench
make test-race

# Coverage
make test-coverage
```

---

### Video 11.3: Writing Effective Tests (18 min)

**Talking Points**:
- Test patterns for LLM providers
- Mocking external services
- Integration test best practices
- Table-driven tests

**Demo**:
```go
// Example test structure
func TestProvider_Complete(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        // test cases
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test logic
        })
    }
}
```

---

### Video 11.4: CI/CD Pipeline Setup (18 min)

**Talking Points**:
- GitHub Actions configuration
- Pipeline stages
- Quality gates
- Deployment automation

**Demo**:
```bash
# View CI configuration
cat .github/workflows/ci.yml

# Run CI validation locally
make ci-validate-all
```

---

### Video 11.5: Quality Gates and Automation (9 min)

**Talking Points**:
- Pre-commit hooks
- Pre-push validation
- Automated testing
- Deployment gates

**Demo**:
```bash
# Pre-commit
make ci-pre-commit

# Pre-push
make ci-pre-push
```

---

## Module 12: Challenge System and Validation (90 minutes)

### Video 12.1: Challenge System Architecture (15 min)

**Talking Points**:
- What is the Challenge System?
- Challenge types:
  - RAGS (RAG Integration)
  - MCPS (MCP Server Integration)
  - SKILLS (Skills Integration)
- Execution flow
- Results directory structure
- 100% test pass rate methodology

**Visual Aids**:
- Challenge architecture diagram
- Results file structure

---

### Video 12.2: RAGS Challenge - RAG Integration (18 min)

**Talking Points**:
- RAG systems tested:
  - Cognee (Knowledge Graph + Memory)
  - Qdrant (Vector Database)
  - RAG Pipeline (Hybrid Search, Reranking, HyDE)
  - Embeddings Service
- 6 test sections
- Running the challenge

**Demo**:
```bash
# Run RAGS challenge
./challenges/scripts/rags_challenge.sh

# View results
cat challenges/results/rags/test_results.csv
```

---

### Video 12.3: MCPS Challenge - MCP Server Integration (18 min)

**Talking Points**:
- 22 MCP servers tested
- Categories:
  - Core: filesystem, memory, fetch, git
  - Database: postgres, sqlite, redis
  - Cloud: docker, kubernetes
  - Communication: slack, notion
  - Vector: chroma, qdrant, weaviate
- MCP Tool Search integration

**Demo**:
```bash
# Run MCPS challenge
./challenges/scripts/mcps_challenge.sh

# View results
cat challenges/results/mcps/test_results.csv
```

---

### Video 12.4: SKILLS Challenge - Skills Integration (12 min)

**Talking Points**:
- 21 skills across 8 categories:
  - Code, Debug, Search, Git
  - Deploy, Docs, Test, Review
- Running the challenge

**Demo**:
```bash
# Run SKILLS challenge
./challenges/scripts/skills_challenge.sh

# View results
```

---

### Video 12.5: Strict Real-Result Validation (15 min)

**Talking Points**:
- What is strict validation?
- FALSE SUCCESS detection:
  - HTTP 200 with no content
  - Empty choices array
  - Error messages in responses
- Content length validation
- RAG evidence detection
- Real vs mock differentiation

**Demo**:
```bash
# Show validation logic
# Demonstrate false positive detection
```

---

### Video 12.6: Debugging Challenge Failures (12 min)

**Talking Points**:
- Analyzing failure reports
- Common failure causes
- Debugging techniques
- Fixing issues

**Demo**:
```bash
# Analyze failures
cat challenges/results/*/test_results.csv | grep FAIL

# Debug specific test
```

---

## Module 13: MCP Tool Search and Discovery (60 minutes)

### Video 13.1: MCP Tool Search Overview (12 min)

**Talking Points**:
- What is MCP Tool Search?
- Search endpoints:
  - `/v1/mcp/tools/search`
  - `/v1/mcp/tools/suggestions`
  - `/v1/mcp/adapters/search`
  - `/v1/mcp/categories`
  - `/v1/mcp/stats`
- Search result structure

**Visual Aids**:
- Endpoint diagram

---

### Video 13.2: Tool Search Implementation (15 min)

**Talking Points**:
- GET and POST search methods
- Query parameters: q, limit, category
- Search result validation

**Demo**:
```bash
# Search for file tools
curl "http://localhost:7061/v1/mcp/tools/search?q=file" | jq

# Search for git tools
curl "http://localhost:7061/v1/mcp/tools/search?q=git" | jq

# POST search
curl -X POST "http://localhost:7061/v1/mcp/tools/search" \
  -H "Content-Type: application/json" \
  -d '{"query": "file operations", "limit": 10}'
```

---

### Video 13.3: AI-Powered Tool Suggestions (12 min)

**Talking Points**:
- Prompt-based recommendation
- Integration with chat
- Automatic tool selection

**Demo**:
```bash
# Get suggestions
curl "http://localhost:7061/v1/mcp/tools/suggestions?prompt=list%20files" | jq
```

---

### Video 13.4: Adapter Search and Discovery (12 min)

**Talking Points**:
- MCP adapter discovery
- Pre-built adapters
- Finding adapters for your use case

**Demo**:
```bash
# Search adapters
curl "http://localhost:7061/v1/mcp/adapters/search?q=github" | jq

# Get categories
curl "http://localhost:7061/v1/mcp/categories" | jq

# Get stats
curl "http://localhost:7061/v1/mcp/stats" | jq
```

---

### Video 13.5: Building Discovery Workflows (9 min)

**Talking Points**:
- Dynamic tool discovery
- Workflow integration
- Best practices

**Demo**:
- Build end-to-end workflow with tool discovery

---

## Module 14: AI Debate System Advanced (90 minutes)

### Video 14.1: 15-LLM Debate Team Configuration (18 min)

**Talking Points**:
- Team structure:
  - 5 positions (Analyst, Proposer, Critic, Synthesizer, Mediator)
  - 3 LLMs per position (primary + 2 fallbacks)
  - Total: 15 LLMs
- Dynamic selection via LLMsVerifier
- OAuth provider priority (Claude, Qwen)

**Demo**:
```bash
# Show debate team configuration
cat configs/debate-team-15llm.yaml
```

---

### Video 14.2: Multi-Pass Validation System (18 min)

**Talking Points**:
- Validation phases:
  1. INITIAL RESPONSE
  2. VALIDATION
  3. POLISH & IMPROVE
  4. FINAL CONCLUSION
- Configuration:
  ```json
  {
    "enable_multi_pass_validation": true,
    "validation_config": {
      "enable_validation": true,
      "enable_polish": true,
      "min_confidence_to_skip": 0.9,
      "max_validation_rounds": 3
    }
  }
  ```
- Quality metrics

**Demo**:
```bash
# Create debate with multi-pass validation
curl -X POST http://localhost:7061/v1/debates \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "Should AI development be open source?",
    "enable_multi_pass_validation": true,
    "validation_config": {
      "enable_validation": true,
      "enable_polish": true,
      "show_phase_indicators": true
    }
  }'
```

---

### Video 14.3: Debate Orchestrator Framework (18 min)

**Talking Points**:
- Framework components:
  - Agent Pool
  - Team Building
  - Protocol Manager
  - Knowledge Repository
- Topologies:
  - Mesh (parallel)
  - Star (hub-spoke)
  - Chain (sequential)
- Phase protocol

**Visual Aids**:
- Topology diagrams
- Protocol flow

---

### Video 14.4: LLMsVerifier Integration (15 min)

**Talking Points**:
- Scoring algorithm (5 components):
  - ResponseSpeed (25%)
  - CostEffectiveness (25%)
  - ModelEfficiency (20%)
  - Capability (20%)
  - Recency (10%)
- OAuth bonus (+0.5)
- Minimum score (5.0)

**Demo**:
```bash
# View verification scores
curl http://localhost:7061/v1/startup/verification | jq
```

---

### Video 14.5: CLI Agent Integration (12 min)

**Talking Points**:
- 48 CLI agents supported
- Agent-specific configurations
- X-CLI-Agent header
- User-Agent pattern matching

**Demo**:
```bash
# Request with CLI agent header
curl -X POST http://localhost:7061/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "X-CLI-Agent: claude-code" \
  -d '{"model":"helixagent-debate","messages":[{"role":"user","content":"Hello"}]}'
```

---

### Video 14.6: Production Debate Deployment (9 min)

**Talking Points**:
- Scaling considerations
- Monitoring in production
- Cost optimization
- Best practices

**Demo**:
```bash
# Production configuration review
cat configs/production-debate.yaml
```

---

## Appendix: Bonus Content Outlines

### Appendix A: Cloud Integrations
- AWS Bedrock integration
- GCP Vertex AI integration
- Azure OpenAI integration

### Appendix B: Kubernetes Deployment
- Kubernetes manifests
- Helm charts
- Scaling configurations
- Monitoring in K8s

### Appendix C: Troubleshooting Guide
- Common issues and solutions
- Debug mode configuration
- Log analysis
- Performance diagnostics

### Appendix D: API Reference Quick Guide
- OpenAI-compatible endpoints
- Protocol endpoints
- Debate API
- Monitoring endpoints

---

*Module Scripts Version: 1.0.0*
*Last Updated: February 2026*
*Total Videos: 74*
