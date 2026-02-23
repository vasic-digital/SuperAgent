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

### Video 14.1: 25-LLM Debate Team Configuration (18 min)

**Talking Points**:
- Team structure:
  - 5 positions (Analyst, Proposer, Critic, Synthesizer, Mediator)
  - 5 LLMs per position (primary + 4 fallbacks)
  - Total: 25 LLMs
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

---

## Module S7.1: Advanced AI/ML Modules — Part 1 (Agentic, LLMOps, SelfImprove)

### Section Overview

This section covers three advanced extracted modules that extend HelixAgent with agentic workflow
orchestration, LLM operations management, and AI self-improvement capabilities. Each module is an
independent Go library (`digital.vasic.agentic`, `digital.vasic.llmops`,
`digital.vasic.selfimprove`) usable standalone or via HelixAgent adapters.

---

### Video S7.1.1: Agentic Module — Graph-Based Workflow Orchestration (30 min)

#### Part 1 — Introduction (5 min)

**Opening Hook**:
- "What if an AI agent could plan, execute, and correct its own multi-step workflows without human
  intervention between each step?"

**Talking Points**:
- What the Agentic module provides:
  - Graph-based workflow definition for autonomous AI agents
  - Nodes represent individual steps; edges encode dependencies and branching
  - Mutable `WorkflowState` shared across all nodes for stateful execution
  - Planning, execution, and self-correction in a single framework
- Module identity: `digital.vasic.agentic` (Go 1.24+)
- Position in the HelixAgent ecosystem: used by `internal/agentic/` for orchestrating multi-step
  agent tasks (code analysis → planning → execution → validation loops)
- Architecture at a glance:

```
WorkflowGraph
  ├── Node (handler func + metadata)
  ├── Node
  └── Edge (source → target, condition)

WorkflowState (mutable, context-threaded)
```

**Visual Aids**:
- Directed graph diagram: nodes as circles, edges as arrows with labels
- Side-by-side comparison: linear pipeline vs graph-based workflow

---

#### Part 2 — Core Concepts (10 min)

**Talking Points**:
- `Workflow` — top-level orchestrator: builds the graph, resolves execution order, runs nodes
- `WorkflowGraph` — pure structure: nodes map + edge list, no execution logic
- `Node` — the unit of work:
  - `ID` string — unique within graph
  - `Handler NodeHandler` — the function that runs
  - `Metadata map[string]interface{}` — arbitrary annotations
- `WorkflowState` — thread-safe mutable bag passed to every handler
- `NodeHandler` signature:
  ```go
  type NodeHandler func(ctx context.Context, state *WorkflowState, input interface{}) (*NodeOutput, error)
  ```
- `NodeOutput`:
  - `NextNode string` — dynamic routing: which node runs next
  - `Data interface{}` — payload forwarded to the next node
  - `Done bool` — signals workflow termination from within a node
- Edge semantics: static edges define graph topology; `NextNode` in output enables runtime routing
- Error propagation: a node returning a non-nil error halts the workflow unless the graph defines
  an error-recovery edge

**Code Walkthrough**:
```go
import "digital.vasic.agentic/agentic"

graph := &agentic.WorkflowGraph{
    Nodes: map[string]*agentic.Node{
        "plan":     {ID: "plan",     Handler: planHandler},
        "execute":  {ID: "execute",  Handler: executeHandler},
        "validate": {ID: "validate", Handler: validateHandler},
    },
    Edges: []agentic.Edge{
        {From: "plan",    To: "execute"},
        {From: "execute", To: "validate"},
    },
}

wf := agentic.NewWorkflow(graph)
state := agentic.NewWorkflowState()
state.Set("goal", "refactor authentication module")

result, err := wf.Run(ctx, "plan", state, nil)
```

**Visual Aids**:
- Annotated code block with callouts for each key type

---

#### Part 3 — Live Demo (10 min)

**Demo: Building a Three-Node Code-Review Agent**

```go
// Node 1: gather context
gatherHandler := func(ctx context.Context, state *agentic.WorkflowState,
    input interface{}) (*agentic.NodeOutput, error) {

    filePath, _ := state.Get("file_path").(string)
    code, err := os.ReadFile(filePath)
    if err != nil {
        return nil, fmt.Errorf("read file: %w", err)
    }
    state.Set("code", string(code))
    return &agentic.NodeOutput{NextNode: "review", Data: string(code)}, nil
}

// Node 2: LLM code review
reviewHandler := func(ctx context.Context, state *agentic.WorkflowState,
    input interface{}) (*agentic.NodeOutput, error) {

    code := input.(string)
    // call LLM provider via HelixAgent adapter
    review, err := llmClient.Complete(ctx, "Review this Go code:\n"+code)
    if err != nil {
        return nil, err
    }
    state.Set("review", review)
    return &agentic.NodeOutput{NextNode: "report", Data: review}, nil
}

// Node 3: generate report
reportHandler := func(ctx context.Context, state *agentic.WorkflowState,
    input interface{}) (*agentic.NodeOutput, error) {

    review := input.(string)
    fmt.Printf("=== Code Review Report ===\n%s\n", review)
    return &agentic.NodeOutput{Done: true}, nil
}
```

**Running the demo**:
```bash
# Build and run the workflow
go test ./agentic/... -v -run TestWorkflow_CodeReview

# Output shows each node executing in order
# PASS: plan -> execute -> validate all ran and state threaded correctly
```

---

#### Part 4 — Integration Patterns (5 min)

**Talking Points**:
- HelixAgent adapter location: `internal/adapters/agentic/adapter.go`
- How HelixAgent uses Agentic:
  - SpecKit orchestrator (`internal/services/speckit_orchestrator.go`) uses workflow graphs for
    the 7-phase development flow (Constitution → Specify → Clarify → Plan → Tasks → Analyze →
    Implement)
  - Each SpecKit phase is a `Node`; phase caching writes to `WorkflowState`
- Standalone usage pattern:
  ```go
  // Direct import, no HelixAgent dependency required
  import "digital.vasic.agentic/agentic"
  ```
- When to use Agentic vs plain sequential code:
  - Use Agentic when: branching on LLM output, retry/self-correction loops, parallel sub-graphs
  - Use plain sequential code when: exactly one linear path, no dynamic routing
- Testing patterns: unit-test each `NodeHandler` independently; integration-test the full graph
  with real state

**Demo**:
```bash
# Run with HelixAgent adapter
curl -X POST http://localhost:7061/v1/agentic/workflows \
  -H "Content-Type: application/json" \
  -d '{"workflow_id": "code-review", "input": {"file_path": "main.go"}}'
```

---

### Video S7.1.2: LLMOps Module — Evaluation, Experiments, and Prompt Versioning (30 min)

#### Part 1 — Introduction (5 min)

**Opening Hook**:
- "Deploying an LLM is easy. Knowing whether it's getting better or worse over time — that requires
  LLMOps."

**Talking Points**:
- What LLMOps means in practice:
  - Continuous evaluation: run your LLM against a dataset and track quality over time
  - A/B experiments: compare two model configurations on live traffic
  - Dataset management: golden, synthetic, and production datasets
  - Prompt versioning: track prompt changes and their impact on quality
- Module identity: `digital.vasic.llmops` (Go 1.24+)
- Position in HelixAgent: used by the LLMsVerifier startup pipeline and provider scoring to detect
  regressions and run controlled experiments when adding new providers
- Why a standalone module: LLMOps concerns are orthogonal to inference — they belong in a separate
  deployable library that any LLM application can adopt

**Visual Aids**:
- LLMOps lifecycle diagram: Deploy → Evaluate → Experiment → Version → Iterate

---

#### Part 2 — Core Concepts (10 min)

**Talking Points**:
- `InMemoryContinuousEvaluator`:
  - Runs an evaluation pipeline against a `Dataset` using a pluggable `LLMClient`
  - Produces an `EvaluationRun` with per-example scores and aggregate metrics
  - "Continuous" means it can be scheduled (cron or event-driven) to detect drift
- `InMemoryExperimentManager`:
  - Creates and tracks A/B experiments between two model configurations (control vs treatment)
  - Assigns traffic splits, collects results, computes statistical significance
  - `Experiment` struct: `ID`, `Name`, `ControlConfig`, `TreatmentConfig`, `TrafficSplit float64`
- `Dataset`:
  ```go
  type Dataset struct {
      ID       string
      Name     string
      Type     DatasetType // golden | synthetic | production
      Examples []Example
  }

  type Example struct {
      ID       string
      Input    string
      Expected string // optional: ground truth for scored evaluation
  }
  ```
- `EvaluationRun`:
  ```go
  type EvaluationRun struct {
      ID          string
      DatasetID   string
      ModelConfig ModelConfig
      StartedAt   time.Time
      CompletedAt time.Time
      Results     []ExampleResult
      AggMetrics  AggregateMetrics
  }
  ```
- Prompt versioning: store prompt templates with semantic version tags, diff between versions,
  roll back when a new version causes score regressions

**Code Walkthrough**:
```go
import "digital.vasic.llmops/llmops"

// Create evaluator
evaluator := llmops.NewInMemoryContinuousEvaluator(llmClient)

// Register a golden dataset
dataset := &llmops.Dataset{
    ID:   "code-generation-v1",
    Type: llmops.DatasetTypeGolden,
    Examples: []llmops.Example{
        {ID: "ex1", Input: "Write a Go HTTP handler", Expected: "func handler(w http.ResponseWriter, r *http.Request)"},
    },
}
evaluator.RegisterDataset(dataset)

// Run evaluation
run, err := evaluator.Evaluate(ctx, "code-generation-v1", modelCfg)
fmt.Printf("Mean score: %.3f\n", run.AggMetrics.MeanScore)
```

---

#### Part 3 — Live Demo (10 min)

**Demo: Running an A/B Experiment**

```go
mgr := llmops.NewInMemoryExperimentManager()

exp, err := mgr.CreateExperiment(ctx, &llmops.Experiment{
    Name: "deepseek-vs-claude-code-gen",
    ControlConfig: llmops.ModelConfig{
        Provider: "deepseek",
        Model:    "deepseek-coder",
    },
    TreatmentConfig: llmops.ModelConfig{
        Provider: "claude",
        Model:    "claude-3.5-sonnet",
    },
    TrafficSplit: 0.5, // 50/50
})

// Record results as traffic flows through
for _, result := range liveResults {
    mgr.RecordResult(ctx, exp.ID, result)
}

// Get winner
summary, err := mgr.GetSummary(ctx, exp.ID)
fmt.Printf("Winner: %s (p-value: %.4f)\n", summary.Winner, summary.PValue)
```

```bash
# Run tests
go test ./llmops/... -v -run TestExperimentManager

# Check HelixAgent experiment endpoint
curl http://localhost:7061/v1/llmops/experiments | jq
curl http://localhost:7061/v1/llmops/evaluations/latest | jq
```

---

#### Part 4 — Integration Patterns (5 min)

**Talking Points**:
- Adapter location: `internal/adapters/llmops/adapter.go`
- How HelixAgent uses LLMOps:
  - LLMsVerifier startup pipeline runs a mini evaluation against each provider using a built-in
    golden dataset (the 8-test verification pipeline)
  - Provider scoring (ResponseSpeed, CostEffectiveness, etc.) feeds into `EvaluationRun` metrics
  - Experiment manager can be used to A/B test provider upgrades (e.g., testing a new Claude model
    against the current one before promoting it)
- Operational pattern — running evaluation on a schedule:
  ```go
  ticker := time.NewTicker(1 * time.Hour)
  for range ticker.C {
      run, _ := evaluator.Evaluate(ctx, "production-golden", currentConfig)
      if run.AggMetrics.MeanScore < threshold {
          alertOps("Score regression detected")
      }
  }
  ```
- Prompt versioning workflow:
  1. `mgr.RegisterPromptVersion(name, version, template)` — stores the prompt
  2. Run evaluation against both old and new versions
  3. `mgr.PromotePrompt(name, version)` if new version wins
  4. `mgr.RollbackPrompt(name)` on regression

---

### Video S7.1.3: SelfImprove Module — RLHF, Reward Modeling, and Preference Optimization (30 min)

#### Part 1 — Introduction (5 min)

**Opening Hook**:
- "The most powerful AI systems don't just respond — they learn from every interaction to become
  measurably better over time."

**Talking Points**:
- What AI self-improvement means in a production system:
  - Reinforcement Learning from Human Feedback (RLHF): collect preference signals, train reward
    models, update response policies
  - Reward modeling: learn to predict which responses humans prefer without asking them every time
  - Preference optimization: directly optimize the model to generate higher-rewarded outputs
  - Continuous self-refinement: iterative loops where the system critiques and improves its own
    outputs
- Module identity: `digital.vasic.selfimprove` (Go 1.24+)
- This module provides the infrastructure layer — it does not train neural networks itself; instead
  it manages the data collection, scoring, and feedback routing that feeds into fine-tuning
  pipelines
- Real-world use case: HelixAgent collects thumbs-up/thumbs-down signals from CLI agent users →
  SelfImprove aggregates these → generates training data → triggers provider fine-tuning jobs

**Visual Aids**:
- RLHF loop diagram: Human Feedback → Reward Model → Policy Update → Better Responses → repeat

---

#### Part 2 — Core Concepts (10 min)

**Talking Points**:
- Core package: `selfimprove` — reward models, feedback collection, optimization
- Feedback collection types:
  - `ExplicitFeedback` — user-provided (thumbs up/down, star ratings, corrections)
  - `ImplicitFeedback` — inferred from behavior (copy, follow-up questions, abandonment)
  - `PreferencePair` — two responses where one is preferred (used for DPO/RLHF training)
- `RewardModel` interface:
  ```go
  type RewardModel interface {
      Score(ctx context.Context, prompt, response string) (float64, error)
      Train(ctx context.Context, pairs []PreferencePair) error
      Evaluate(ctx context.Context, dataset []PreferencePair) (*RewardMetrics, error)
  }
  ```
- `FeedbackCollector`:
  - Buffers incoming feedback signals
  - Deduplicates by session and prompt hash
  - Exports batches for reward model training
- `PreferenceOptimizer`:
  - Takes a `RewardModel` and a set of candidate responses
  - Returns the highest-scored response (at inference time)
  - Can trigger DPO (Direct Preference Optimization) training runs
- `SelfRefinementLoop`:
  - Generates an initial response
  - Critiques it using a second LLM call or a reward model
  - Regenerates with the critique as context
  - Stops when score exceeds threshold or max iterations reached

**Code Walkthrough**:
```go
import "digital.vasic.selfimprove/selfimprove"

// Collect feedback
collector := selfimprove.NewFeedbackCollector()
collector.Record(selfimprove.ExplicitFeedback{
    PromptHash: hash(prompt),
    ResponseID: resp.ID,
    Signal:     selfimprove.SignalThumbsUp,
    UserID:     "user-123",
})

// Score a candidate response
rewardModel := selfimprove.NewInMemoryRewardModel()
score, err := rewardModel.Score(ctx, prompt, candidateResponse)
fmt.Printf("Reward score: %.3f\n", score)
```

---

#### Part 3 — Live Demo (10 min)

**Demo: Self-Refinement Loop**

```go
refiner := selfimprove.NewSelfRefinementLoop(selfimprove.SelfRefinementConfig{
    MaxIterations:  3,
    ScoreThreshold: 0.85,
    CritiquePrompt: "Critique this response for accuracy, clarity, and completeness: ",
    RefinePrompt:   "Improve this response based on the critique: ",
})

initial := "The quick sort algorithm works by partitioning arrays."
refined, metrics, err := refiner.Refine(ctx, prompt, initial, llmClient, rewardModel)

fmt.Printf("Iterations: %d\n", metrics.Iterations)
fmt.Printf("Initial score: %.3f → Final score: %.3f\n",
    metrics.InitialScore, metrics.FinalScore)
fmt.Printf("Refined response: %s\n", refined)
```

```bash
# Run the self-improvement tests
go test ./selfimprove/... -v -run TestSelfRefinement

# Check improvement metrics via HelixAgent
curl http://localhost:7061/v1/selfimprove/metrics | jq
```

---

#### Part 4 — Integration Patterns (5 min)

**Talking Points**:
- Adapter location: `internal/adapters/selfimprove/adapter.go`
- How HelixAgent integrates SelfImprove:
  - The AI Debate multi-pass validation (Modules 14.2) is conceptually a self-refinement loop;
    SelfImprove provides the formal framework
  - CLI agent interactions generate implicit feedback (session length, follow-up rate)
  - The `FeedbackCollector` is wired into the streaming response handler
- End-to-end flow with HelixAgent:
  ```
  User request → HelixAgent streams response
  → FeedbackCollector captures implicit signals
  → Nightly batch: RewardModel.Train(pairs)
  → Next day: PreferenceOptimizer uses updated model
  → Better responses on similar prompts
  ```
- Production considerations:
  - Feedback data is PII-sensitive: anonymize before training
  - Reward model drift: re-evaluate monthly against a held-out golden set
  - Cold-start problem: use rule-based scoring for the first 1000 interactions

---

## Module S7.2: Advanced AI/ML Modules — Part 2 (Planning, Benchmark)

### Section Overview

This section covers two more advanced modules: Planning (HiPlan, MCTS, Tree of Thoughts) and
Benchmark (standardized LLM evaluation across MMLU, HumanEval, GSM8K, and custom suites).

---

### Video S7.2.1: Planning Module — HiPlan, MCTS, and Tree of Thoughts (30 min)

#### Part 1 — Introduction (5 min)

**Opening Hook**:
- "An LLM that can only respond to prompts is a calculator. An LLM that can plan ahead is an
  agent."

**Talking Points**:
- The problem with single-shot LLM responses for complex tasks:
  - No lookahead: the model commits to a path without exploring alternatives
  - No backtracking: a wrong sub-step contaminates all subsequent reasoning
  - No hierarchical decomposition: large tasks exceed context windows
- What the Planning module solves:
  - Three complementary algorithms covering different planning regimes
  - **HiPlan** — hierarchical decomposition for structured, known-topology tasks
  - **MCTS** — Monte Carlo Tree Search for exploratory planning with uncertainty
  - **Tree of Thoughts** — breadth-first thought exploration with LLM-powered evaluation
- Module identity: `digital.vasic.planning` (Go 1.24+)
- When to use which algorithm:
  - HiPlan: multi-phase projects with known milestones (software development, research pipelines)
  - MCTS: game-like tasks where you can simulate outcomes (code generation with test feedback)
  - ToT: open-ended reasoning where the "right" path is unknown (math proofs, strategic decisions)

**Visual Aids**:
- Side-by-side comparison: HiPlan tree (top-down), MCTS game tree, ToT thought branches

---

#### Part 2 — Core Concepts (10 min)

**Talking Points**:

**HiPlan (Hierarchical Planning)**:
```go
type HiPlan struct {
    config    *HiPlanConfig
    generator MilestoneGenerator
    executor  StepExecutor
}

type HiPlanConfig struct {
    MaxMilestones   int
    MaxStepsPerMile int
    TimeoutPerStep  time.Duration
}

// Interfaces for extension
type MilestoneGenerator interface {
    GenerateMilestones(ctx context.Context, goal string) ([]Milestone, error)
}

type StepExecutor interface {
    ExecuteStep(ctx context.Context, step PlanStep, state interface{}) (interface{}, error)
}
```

**MCTS**:
```go
type MCTSConfig struct {
    MaxIterations  int
    ExplorationC   float64 // UCB exploration constant
    MaxDepth       int
    RolloutDepth   int
}

// Strategy interfaces
type MCTSActionGenerator interface {
    GenerateActions(ctx context.Context, node *MCTSNode) ([]interface{}, error)
}
type MCTSRewardFunction interface {
    Reward(ctx context.Context, node *MCTSNode, action interface{}) (float64, error)
}
```

**Tree of Thoughts**:
```go
type TreeOfThoughtsConfig struct {
    BranchingFactor  int
    MaxDepth         int
    BeamWidth        int
    EvaluationPrompt string
}

type ThoughtGenerator interface {
    Generate(ctx context.Context, parent *ThoughtNode) ([]Thought, error)
}
type ThoughtEvaluator interface {
    Evaluate(ctx context.Context, thought Thought) (float64, error)
}
```

---

#### Part 3 — Live Demo (10 min)

**Demo 1 — HiPlan for a software feature**:
```go
plan := planning.NewHiPlan(planning.DefaultHiPlanConfig(),
    &planning.LLMMilestoneGenerator{Client: llmClient},
    &myStepExecutor{},
)

result, err := plan.Execute(ctx, "Implement user authentication with JWT")
for _, milestone := range result.Milestones {
    fmt.Printf("[Milestone] %s\n", milestone.Title)
    for _, step := range milestone.Steps {
        fmt.Printf("  [Step] %s: %s\n", step.ID, step.Description)
    }
}
```

**Demo 2 — MCTS for code optimization**:
```go
mcts := planning.NewMCTS(planning.DefaultMCTSConfig(),
    &planning.CodeActionGenerator{},
    &planning.CodeRewardFunction{TestRunner: runner},
    &planning.DefaultRolloutPolicy{},
)

result, err := mcts.Search(ctx, initialState)
fmt.Printf("Best action: %v (score: %.3f)\n", result.BestAction, result.BestScore)
```

```bash
# Run planning tests
go test ./planning/... -v -run TestHiPlan_Execute
go test ./planning/... -v -run TestMCTS_Search
go test ./planning/... -v -run TestTreeOfThoughts_Run

# HelixAgent planning endpoint
curl -X POST http://localhost:7061/v1/planning/run \
  -H "Content-Type: application/json" \
  -d '{"algorithm": "hiplan", "goal": "Build REST API for user management"}'
```

---

#### Part 4 — Integration Patterns (5 min)

**Talking Points**:
- Adapter location: `internal/adapters/planning/adapter.go`
- How HelixAgent uses Planning:
  - SpecKit orchestrator uses HiPlan to decompose large refactoring tasks into milestones and steps
    (`GranularityRefactoring` triggers a HiPlan execution before task assignment)
  - The AI Debate system optionally uses Tree of Thoughts to explore multiple response strategies
    before committing to the Proposal phase
  - MCTS is used in the code-generation skill to search for optimal implementations given unit test
    feedback
- Choosing algorithm parameters:
  - HiPlan: `MaxMilestones=5` for most tasks; increase only for very large projects
  - MCTS: `MaxIterations=100` is a good start; `ExplorationC=1.414` (sqrt(2)) is the classic UCB
    constant
  - ToT: `BranchingFactor=3`, `MaxDepth=4`, `BeamWidth=2` balances quality vs cost
- Cost control: each tree node = 1 LLM call; set `MaxDepth` conservatively in production

---

### Video S7.2.2: Benchmark Module — Standardized LLM Evaluation (30 min)

#### Part 1 — Introduction (5 min)

**Opening Hook**:
- "Everyone claims their LLM is the best. The Benchmark module gives you the data to know for
  sure — against your actual workloads, not marketing benchmarks."

**Talking Points**:
- Why standardized benchmarks matter:
  - Published benchmarks (MMLU, HumanEval, GSM8K) measure specific capabilities under controlled
    conditions — they may not reflect your use case
  - You need both: standard benchmarks for apples-to-apples provider comparison, AND custom
    benchmarks for your domain
  - Benchmarking is essential before promoting a provider upgrade (ties into LLMOps evaluation)
- Supported benchmarks out of the box:
  - **MMLU** — Massive Multitask Language Understanding (57 subjects, multiple choice)
  - **HumanEval** — Python code generation (164 problems, pass@k metric)
  - **GSM8K** — Grade school math word problems (8500 problems)
  - **SWE-Bench** — Software engineering tasks on real GitHub issues
  - **MBPP** — Mostly Basic Python Problems
  - **LMSYS** — Chatbot Arena style head-to-head comparison
  - **HellaSwag** — Commonsense NLI
  - **MATH** — Competition mathematics
  - **Custom** — Bring your own benchmark suite
- Module identity: `digital.vasic.benchmark` (Go 1.24+)

**Visual Aids**:
- Benchmark comparison table: which benchmark measures what capability

---

#### Part 2 — Core Concepts (10 min)

**Talking Points**:
- Core package: `benchmark` — runner, types, integration adapters, metrics
- `BenchmarkRunner`:
  ```go
  type BenchmarkRunner interface {
      Run(ctx context.Context, cfg *RunConfig) (*BenchmarkResult, error)
      RunSuite(ctx context.Context, cfgs []*RunConfig) ([]*BenchmarkResult, error)
      Compare(ctx context.Context, results []*BenchmarkResult) (*ComparisonReport, error)
  }
  ```
- `RunConfig`:
  ```go
  type RunConfig struct {
      BenchmarkID  string          // "mmlu", "humaneval", "gsm8k", "custom"
      Provider     string
      Model        string
      MaxExamples  int             // limit for fast runs
      Temperature  float64
      Timeout      time.Duration
      CustomDataset *Dataset       // for custom benchmarks
  }
  ```
- `BenchmarkResult`:
  ```go
  type BenchmarkResult struct {
      BenchmarkID  string
      Provider     string
      Model        string
      Score        float64         // primary metric (accuracy, pass@1, etc.)
      SubScores    map[string]float64 // per-category or per-subject scores
      Latency      LatencyStats    // p50, p95, p99
      Cost         CostStats       // tokens, estimated USD
      RunAt        time.Time
  }
  ```
- `ComparisonReport`:
  - Side-by-side provider comparison
  - Statistical significance testing (bootstrapped confidence intervals)
  - Recommendation: which provider to use for which task type
- Integration adapters: each benchmark has a dedicated adapter that handles dataset loading,
  prompt formatting, answer extraction, and scoring

---

#### Part 3 — Live Demo (10 min)

**Demo: Running benchmarks and comparing providers**

```go
runner := benchmark.NewBenchmarkRunner(benchmark.RunnerConfig{
    Providers: []benchmark.ProviderConfig{
        {Name: "deepseek", Model: "deepseek-coder"},
        {Name: "claude",   Model: "claude-3.5-sonnet"},
    },
    Parallelism: 4,
    OutputDir:   "./benchmark-results",
})

// Run HumanEval on two providers
results, err := runner.RunSuite(ctx, []*benchmark.RunConfig{
    {BenchmarkID: "humaneval", Provider: "deepseek", Model: "deepseek-coder", MaxExamples: 50},
    {BenchmarkID: "humaneval", Provider: "claude",   Model: "claude-3.5-sonnet", MaxExamples: 50},
})

// Compare
report, err := runner.Compare(ctx, results)
for _, entry := range report.Ranking {
    fmt.Printf("%s/%s: %.1f%% (p95 latency: %dms)\n",
        entry.Provider, entry.Model,
        entry.Score*100,
        entry.Latency.P95.Milliseconds())
}
```

```bash
# Run via Make
GOMAXPROCS=2 nice -n 19 go test ./benchmark/... -v -run TestBenchmarkRunner

# Run via HelixAgent CLI
curl -X POST http://localhost:7061/v1/benchmark/run \
  -H "Content-Type: application/json" \
  -d '{
    "benchmark": "humaneval",
    "providers": ["deepseek", "claude"],
    "max_examples": 50
  }' | jq '.ranking'

# Get latest comparison report
curl http://localhost:7061/v1/benchmark/reports/latest | jq
```

---

#### Part 4 — Integration Patterns (5 min)

**Talking Points**:
- Adapter location: `internal/adapters/benchmark/adapter.go`
- How HelixAgent uses Benchmark:
  - LLMsVerifier startup pipeline runs a lightweight 8-test verification using the Benchmark runner
    under the hood (the verification pipeline IS a benchmark suite)
  - The `CostEffectiveness` and `ModelEfficiency` scoring components pull from BenchmarkResult
    metrics
  - Provider promotions: before `SVC_*` overrides promote a new model to production, the Benchmark
    adapter runs a quick HumanEval/GSM8K suite to gate the promotion
- Custom benchmark workflow:
  ```go
  // 1. Define your domain dataset
  dataset := &benchmark.Dataset{
      ID: "my-company-sql-tasks",
      Examples: []benchmark.Example{
          {Input: "Write SQL to find top 10 customers by revenue", Expected: "SELECT ..."},
      },
  }

  // 2. Register and run
  runner.RegisterCustomDataset(dataset)
  result, _ := runner.Run(ctx, &benchmark.RunConfig{
      BenchmarkID:   "custom",
      CustomDataset: dataset,
      Provider:      "deepseek",
      Model:         "deepseek-chat",
  })
  ```
- Resource management: benchmark runs are compute-intensive; always use `GOMAXPROCS=2`, `nice -n
  19`, `MaxExamples` limits in CI/CD. Full benchmark suites belong in scheduled nightly jobs, not
  pre-commit hooks.
- Storing results: `BenchmarkResult` objects are persisted to PostgreSQL via the Database adapter,
  enabling trend analysis over weeks and months

---

## Appendix E: AI/ML Module Integration Reference

### Module Dependency Map

```
HelixAgent core
  ├── internal/adapters/agentic/     → digital.vasic.agentic
  ├── internal/adapters/llmops/      → digital.vasic.llmops
  ├── internal/adapters/selfimprove/ → digital.vasic.selfimprove
  ├── internal/adapters/planning/    → digital.vasic.planning
  └── internal/adapters/benchmark/  → digital.vasic.benchmark
```

### go.mod replace directives (development)

```go
replace (
    digital.vasic.agentic      => ./Agentic
    digital.vasic.llmops       => ./LLMOps
    digital.vasic.selfimprove  => ./SelfImprove
    digital.vasic.planning     => ./Planning
    digital.vasic.benchmark    => ./Benchmark
)
```

### Challenge Scripts

```bash
./challenges/scripts/agentic_challenge.sh       # Workflow orchestration validation
./challenges/scripts/llmops_challenge.sh        # Evaluation pipeline validation
./challenges/scripts/selfimprove_challenge.sh   # RLHF loop validation
./challenges/scripts/planning_challenge.sh      # HiPlan/MCTS/ToT validation
./challenges/scripts/benchmark_challenge.sh     # Provider benchmark comparison
```

---

*Module Scripts Version: 1.1.0*
*Last Updated: February 2026*
*Total Videos: 84 (74 original + 5 new AI/ML module videos + 5 demos)*
