# HelixAgent Video Course: Complete Training Program

## Course Overview

**Title**: "Mastering HelixAgent: Multi-Provider AI Orchestration"
**Duration**: 14+ hours across 14 comprehensive modules
**Target Audience**: Developers, DevOps engineers, AI engineers, and technical decision-makers
**Prerequisites**: Basic programming knowledge, familiarity with REST APIs, Go 1.24+ for development modules
**Skill Level**: Beginner to Advanced

---

## Course Objectives

Upon completion of this course, participants will be able to:

1. Understand the architecture and design principles of HelixAgent
2. Install, configure, and deploy HelixAgent in various environments
3. Integrate 10 LLM providers (Claude, Gemini, DeepSeek, Qwen, Ollama, OpenRouter, ZAI, Zen, Mistral, Cerebras)
4. Implement ensemble voting strategies for improved AI responses
5. Configure and utilize the AI Debate System with 15 LLMs for complex problem-solving
6. Develop custom plugins for extended functionality
7. Integrate MCP/LSP/ACP protocols for enhanced capabilities
8. Apply LLM optimization techniques for better performance
9. Implement security best practices for production deployments
10. Set up comprehensive testing and CI/CD pipelines
11. **NEW**: Run and interpret RAGS, MCPS, and SKILLS challenges
12. **NEW**: Implement MCP Tool Search and discovery
13. **NEW**: Configure multi-pass validation for AI debates
14. **NEW**: Integrate with 20+ CLI agents

---

## Module 1: Introduction to HelixAgent (45 minutes)

### Learning Objectives
- Understand what HelixAgent is and its core value proposition
- Learn the benefits of multi-provider AI orchestration
- Familiarize with the system architecture and components

### Videos

#### 1.1 Course Welcome and Learning Path (10 min)
- Welcome and instructor introduction
- Course structure and navigation
- Prerequisites review
- Learning outcomes overview
- How to get the most from this course

#### 1.2 What is HelixAgent? (15 min)
- Multi-provider AI orchestration explained
- The problem of vendor lock-in
- Benefits of intelligent routing
- Real-world use cases and applications
- Comparison with single-provider approaches

#### 1.3 Architecture Overview (20 min)
- System architecture diagram walkthrough
- Core components and their responsibilities:
  - LLM Provider Abstraction Layer
  - Ensemble Orchestration Engine
  - AI Debate Service
  - Plugin System
  - Protocol Managers (MCP, LSP, ACP)
  - Caching Layer
  - Monitoring and Metrics
- Data flow and processing pipeline
- Scalability and performance design

### Hands-On Lab
- Explore the HelixAgent repository structure
- Review key configuration files
- Examine API documentation

---

## Module 2: Installation and Setup (60 minutes)

### Learning Objectives
- Set up development environment
- Install HelixAgent using different methods
- Verify successful installation

### Videos

#### 2.1 Environment Prerequisites (15 min)
- System requirements overview
- Go 1.23+ installation and verification
- Docker and Docker Compose setup
- Git configuration
- IDE recommendations (VS Code, GoLand)

#### 2.2 Quick Start with Docker (20 min)
- Docker-based installation walkthrough
- Docker Compose configuration
- Starting core services (PostgreSQL, Redis, Cognee)
- Starting with different profiles:
  - Core services
  - AI services (Ollama)
  - Monitoring (Prometheus, Grafana)
  - Full stack deployment
- Container health verification

#### 2.3 Manual Installation from Source (20 min)
- Repository cloning
- Dependency installation
- Building from source:
  ```bash
  make build
  make build-debug
  ```
- Running locally:
  ```bash
  make run
  make run-dev
  ```
- Troubleshooting common installation issues

#### 2.4 Podman Alternative Setup (5 min)
- Podman compatibility overview
- Podman-specific configuration
- Using the container-runtime script

### Hands-On Lab
- Complete installation using preferred method
- Start all services
- Verify API endpoints are accessible
- Run initial health checks

---

## Module 3: Configuration (60 minutes)

### Learning Objectives
- Master HelixAgent configuration options
- Configure environment variables and YAML files
- Set up multi-environment configurations

### Videos

#### 3.1 Configuration Architecture (15 min)
- Configuration file structure
- Environment variable hierarchy
- Configuration loading precedence
- Secrets management best practices

#### 3.2 Core Configuration Options (20 min)
- Server configuration (PORT, GIN_MODE, JWT_SECRET)
- Database settings (PostgreSQL configuration)
- Redis cache configuration
- Logging and debugging options
- Configuration files walkthrough:
  - `configs/development.yaml`
  - `configs/production.yaml`
  - `configs/multi-provider.yaml`

#### 3.3 Provider Configuration (15 min)
- API key configuration for each provider:
  - CLAUDE_API_KEY
  - DEEPSEEK_API_KEY
  - GEMINI_API_KEY
  - QWEN_API_KEY
  - ZAI_API_KEY
  - OPENROUTER_API_KEY
- Ollama local model configuration
- Provider-specific settings and models
- Rate limiting and timeout configuration

#### 3.4 Advanced Configuration (10 min)
- AI Debate configuration
- Cognee integration settings
- Optimization feature configuration
- Cloud provider integration (AWS Bedrock, GCP Vertex AI, Azure OpenAI)

### Hands-On Lab
- Create custom configuration for your use case
- Configure multiple providers
- Set up environment-specific configurations
- Test configuration validation

---

## Module 4: LLM Provider Integration (75 minutes)

### Learning Objectives
- Integrate all supported LLM providers
- Understand provider capabilities and limitations
- Implement fallback chains for reliability

### Videos

#### 4.1 Provider Interface and Architecture (15 min)
- `LLMProvider` interface explained
- Provider lifecycle management
- Health checking and monitoring
- Error handling patterns

#### 4.2 Claude Integration (10 min)
- Claude API configuration
- Model selection (Claude 3 Opus, Sonnet, Haiku)
- Claude-specific features
- Best use cases for Claude

#### 4.3 DeepSeek Integration (10 min)
- DeepSeek API setup
- DeepSeek-Coder and DeepSeek-Chat models
- Technical content optimization
- Cost-effective processing strategies

#### 4.4 Gemini Integration (10 min)
- Google Gemini API configuration
- Gemini Pro and multimodal capabilities
- Scientific and analytical use cases
- GCP Vertex AI integration

#### 4.5 Qwen, Ollama, and Other Providers (15 min)
- Qwen API configuration
- Ollama local model setup and advantages
- OpenRouter as a meta-provider
- ZAI integration
- Provider selection guidelines

#### 4.6 Building Fallback Chains (15 min)
- Designing reliable fallback strategies
- Primary and secondary provider configuration
- Health-based routing
- Cost-optimized fallback ordering
- Testing fallback scenarios

### Hands-On Lab
- Configure at least 3 different providers
- Implement a fallback chain
- Test provider health checks
- Compare response quality across providers

---

## Module 5: Ensemble Strategies (60 minutes)

### Learning Objectives
- Understand ensemble voting mechanisms
- Implement different voting strategies
- Optimize ensemble performance

### Videos

#### 5.1 Introduction to Ensemble AI (15 min)
- What is ensemble learning in AI?
- Benefits of multi-model consensus
- Accuracy improvements through diversity
- Real-world ensemble applications

#### 5.2 Voting Strategies (20 min)
- Available voting strategies:
  - Majority voting
  - Weighted voting
  - Consensus voting
  - Confidence-weighted voting
  - Quality-weighted voting
- Selecting the right strategy for your use case
- Configuration examples

#### 5.3 Implementing Custom Strategies (15 min)
- VotingStrategy interface
- Creating custom voting algorithms
- Testing and validating strategies
- Performance considerations

#### 5.4 Performance Optimization (10 min)
- Parallel execution strategies
- Caching for ensemble responses
- Reducing latency in multi-provider calls
- Cost optimization techniques

### Hands-On Lab
- Configure different voting strategies
- Compare results across strategies
- Implement a simple custom strategy
- Benchmark ensemble performance

---

## Module 6: AI Debate System (90 minutes)

### Learning Objectives
- Master the AI Debate configuration system
- Configure participants and LLM chains
- Implement advanced debate strategies

### Videos

#### 6.1 AI Debate Concepts (20 min)
- What is AI debate?
- Benefits of multi-agent discussions
- Consensus building through debate
- Use cases: complex reasoning, decision making, analysis

#### 6.2 Participant Configuration (20 min)
- Participant structure and roles
- Configuring participant properties:
  - Name and role
  - Debate style (analytical, creative, balanced, etc.)
  - Argumentation style (logical, evidence-based, etc.)
  - Weight and priority
  - Quality thresholds
- LLM fallback chains per participant

#### 6.3 Debate Strategies (20 min)
- Available strategies:
  - Round-robin
  - Free-form
  - Structured
  - Adversarial
  - Collaborative
- Consensus threshold configuration
- Maximal repeat rounds
- Timeout management

#### 6.4 Cognee AI Integration (15 min)
- Enhanced responses with Cognee
- Semantic analysis and insights
- Dataset configuration
- Memory integration

#### 6.5 Programmatic Debate Execution (15 min)
- Using the Debate API
- Handling debate responses
- Interpreting consensus results
- Error handling and retries

### Hands-On Lab
- Configure a multi-participant debate
- Implement different debate strategies
- Test consensus building
- Analyze debate results and quality scores

---

## Module 7: Plugin Development (75 minutes)

### Learning Objectives
- Understand the plugin architecture
- Develop custom plugins
- Implement hot-reloadable extensions

### Videos

#### 7.1 Plugin Architecture Overview (15 min)
- Plugin system design
- Hot-reloading mechanism
- Plugin lifecycle management
- Discovery and registration

#### 7.2 Plugin Interfaces (20 min)
- PluginRegistry interface
- PluginLoader interface
- Plugin metadata and configuration
- Health and metrics integration

#### 7.3 Developing Your First Plugin (25 min)
- Plugin project structure
- Implementing required interfaces
- Plugin configuration
- Building and packaging
- Testing plugins locally

#### 7.4 Advanced Plugin Topics (15 min)
- Dependency resolution
- Plugin communication
- Error handling best practices
- Performance considerations
- Deployment strategies

### Hands-On Lab
- Create a simple plugin
- Implement plugin configuration
- Test hot-reloading
- Deploy and verify in running system

---

## Module 8: MCP/LSP Integration (60 minutes)

### Learning Objectives
- Understand protocol support in HelixAgent
- Configure MCP, LSP, and ACP servers
- Implement protocol-based workflows

### Videos

#### 8.1 Protocol Support Overview (15 min)
- Unified Protocol Manager architecture
- Supported protocols:
  - MCP (Model Context Protocol)
  - LSP (Language Server Protocol)
  - ACP (Agent Client Protocol)
  - Embeddings
- Use cases for each protocol

#### 8.2 MCP Integration (15 min)
- MCP server configuration
- Tool execution and management
- MCP API endpoints
- Building MCP-enabled workflows

#### 8.3 LSP Integration (15 min)
- Language Server Protocol basics
- LSP server configuration
- Code intelligence features
- IDE integration possibilities

#### 8.4 ACP and Embeddings (15 min)
- Agent Client Protocol configuration
- Agent communication patterns
- Embedding generation and comparison
- Vector operations and semantic search

### Hands-On Lab
- Configure an MCP server
- Execute MCP tools via API
- Set up embedding generation
- Test protocol metrics and health

---

## Module 9: Optimization Features (75 minutes)

### Learning Objectives
- Implement LLM optimization techniques
- Configure semantic caching
- Use structured output generation

### Videos

#### 9.1 Optimization Framework Overview (15 min)
- 8 optimization tools integrated
- Native Go vs HTTP client implementations
- Docker services for optimization
- Configuration overview

#### 9.2 Semantic Caching with GPTCache (15 min)
- Vector similarity caching
- LRU eviction strategies
- TTL configuration
- Cache hit optimization

#### 9.3 Structured Output with Outlines (15 min)
- JSON schema validation
- Regex pattern constraints
- Choice constraints
- Ensuring output format compliance

#### 9.4 Enhanced Streaming (10 min)
- Word and sentence buffering
- Progress tracking
- Rate limiting
- Real-time response handling

#### 9.5 Advanced Optimization (SGLang, LlamaIndex) (20 min)
- RadixAttention prefix caching
- Document retrieval with HyDE
- Reranking strategies
- Cognee integration for RAG

### Hands-On Lab
- Enable semantic caching
- Configure structured output schemas
- Test enhanced streaming
- Measure optimization improvements

---

## Module 10: Security Best Practices (60 minutes)

### Learning Objectives
- Implement security best practices
- Configure authentication and authorization
- Secure production deployments

### Videos

#### 10.1 Security Architecture (15 min)
- Security-first design principles
- Authentication mechanisms
- Authorization and RBAC
- Threat model overview

#### 10.2 API Security (15 min)
- JWT token configuration
- API key management
- Rate limiting implementation
- Input validation
- Request size limits

#### 10.3 Secrets Management (15 min)
- Environment variable best practices
- API key rotation
- Secure configuration storage
- Secrets in containers

#### 10.4 Production Security Hardening (15 min)
- Network security
- Container security
- Database security
- Logging and audit trails
- Security testing with `make test-security`

### Hands-On Lab
- Configure JWT authentication
- Implement rate limiting
- Set up secure secrets management
- Run security scan with gosec

---

## Module 11: Testing and CI/CD (75 minutes)

### Learning Objectives
- Master HelixAgent testing strategies
- Set up comprehensive CI/CD pipelines
- Implement quality gates

### Videos

#### 11.1 Testing Strategy Overview (15 min)
- Test pyramid for HelixAgent
- Test types:
  - Unit tests
  - Integration tests
  - E2E tests
  - Security tests
  - Stress tests
  - Chaos tests
- Coverage targets and metrics

#### 11.2 Running Tests (20 min)
- Make commands:
  ```bash
  make test
  make test-coverage
  make test-unit
  make test-integration
  make test-e2e
  make test-security
  make test-stress
  make test-chaos
  make test-bench
  make test-race
  ```
- Test infrastructure setup
- Using Docker for test dependencies

#### 11.3 Writing Effective Tests (20 min)
- Test patterns for LLM providers
- Mocking external services
- Integration test best practices
- Performance testing guidelines

#### 11.4 CI/CD Pipeline Setup (20 min)
- GitHub Actions configuration
- Pipeline stages:
  - Linting and formatting
  - Unit testing
  - Integration testing
  - Security scanning
  - Docker image building
  - Deployment automation
- Quality gates and approvals

### Hands-On Lab
- Run all test suites
- Analyze coverage reports
- Write a custom integration test
- Review CI/CD pipeline configuration

---

## Module 12: Challenge System and Validation (90 minutes)

### Learning Objectives
- Master the HelixAgent Challenge System
- Run RAGS, MCPS, and SKILLS challenges
- Understand strict real-result validation
- Validate system integration across 20+ CLI agents

### Videos

#### 12.1 Challenge System Architecture (20 min)
- What is the Challenge System?
- Challenge types:
  - RAGS Challenge (RAG Integration)
  - MCPS Challenge (MCP Server Integration)
  - SKILLS Challenge (Skills Integration)
- Challenge execution flow
- Results directory structure
- 100% test pass rate methodology

#### 12.2 RAGS Challenge - RAG Integration (20 min)
- RAG systems tested:
  - Cognee (Knowledge Graph + Memory)
  - Qdrant (Vector Database)
  - RAG Pipeline (Hybrid Search, Reranking, HyDE)
  - Embeddings Service
- 6 test sections:
  - RAG Endpoint Availability
  - CLI Agents RAG Access
  - RAG Trigger via AI Debate
  - Cognee Integration Depth
  - Qdrant/Vector DB Integration
  - RAG Pipeline Advanced Features
- Running the RAGS challenge:
  ```bash
  ./challenges/scripts/rags_challenge.sh
  ```

#### 12.3 MCPS Challenge - MCP Server Integration (20 min)
- 22 MCP servers tested:
  - Core: filesystem, memory, fetch, git, github, gitlab
  - Database: postgres, sqlite, redis, mongodb
  - Cloud: docker, kubernetes, aws-s3, google-drive
  - Communication: slack, notion
  - Search: brave-search
  - Vector: chroma, qdrant, weaviate
- MCP Tool Search integration
- Protocol endpoints (MCP, LSP, ACP)
- Running the MCPS challenge:
  ```bash
  ./challenges/scripts/mcps_challenge.sh
  ```

#### 12.4 SKILLS Challenge - Skills Integration (15 min)
- 21 skills across 8 categories:
  - Code (generate, refactor, optimize)
  - Debug (trace, profile, analyze)
  - Search (find, grep, semantic-search)
  - Git (commit, branch, merge)
  - Deploy (build, deploy)
  - Docs (document, explain, readme)
  - Test (unit-test, integration-test)
  - Review (lint, security-scan)
- Running the SKILLS challenge:
  ```bash
  ./challenges/scripts/skills_challenge.sh
  ```

#### 12.5 Strict Real-Result Validation (15 min)
- What is strict validation?
- FALSE SUCCESS detection:
  - HTTP 200 with no real content
  - Empty choices array
  - Error messages in responses
- Content length validation
- RAG evidence detection
- Real vs mock response differentiation
- Validation code walkthrough

### Hands-On Lab
- Run all three challenges (RAGS, MCPS, SKILLS)
- Analyze challenge reports
- Review test_results.csv outputs
- Interpret pass rates and failures
- Debug a failing challenge test

---

## Module 13: MCP Tool Search and Discovery (60 minutes)

### Learning Objectives
- Master MCP Tool Search functionality
- Implement semantic tool discovery
- Configure tool suggestions for prompts

### Videos

#### 13.1 MCP Tool Search Overview (15 min)
- What is MCP Tool Search?
- Search endpoints:
  - `/v1/mcp/tools/search` - Tool search by query
  - `/v1/mcp/tools/suggestions` - AI-powered suggestions
  - `/v1/mcp/adapters/search` - Adapter search
  - `/v1/mcp/categories` - Tool categories
  - `/v1/mcp/stats` - Usage statistics
- Search result structure

#### 13.2 Tool Search Implementation (20 min)
- GET and POST search methods
- Query parameters:
  - `q` - Search query
  - `limit` - Result limit
  - `category` - Filter by category
- Search result validation
- Real-time tool discovery
- Example searches:
  ```bash
  # Search for file tools
  curl "${HELIXAGENT_URL}/v1/mcp/tools/search?q=file"

  # Search for git tools
  curl "${HELIXAGENT_URL}/v1/mcp/tools/search?q=git"

  # POST search with options
  curl -X POST "${HELIXAGENT_URL}/v1/mcp/tools/search" \
    -d '{"query": "file operations", "limit": 10}'
  ```

#### 13.3 Tool Suggestions (15 min)
- AI-powered tool suggestions
- Prompt-based recommendation
- Suggestion endpoints:
  ```bash
  curl "${HELIXAGENT_URL}/v1/mcp/tools/suggestions?prompt=list%20files"
  ```
- Integration with chat completions
- Automatic tool selection

#### 13.4 Adapter Search (10 min)
- MCP adapter discovery
- Pre-built adapters:
  - GitHub, GitLab
  - PostgreSQL, MongoDB
  - Slack, Notion
  - Filesystem, Git
- Finding adapters for your use case

### Hands-On Lab
- Search for tools using different queries
- Test tool suggestions with various prompts
- Explore adapter search functionality
- Build a custom tool discovery workflow

---

## Module 14: AI Debate System Advanced (90 minutes)

### Learning Objectives
- Configure the 15 LLM AI Debate Ensemble
- Implement multi-pass validation
- Integrate with LLMsVerifier scoring

### Videos

#### 14.1 AI Debate with 15 LLMs (25 min)
- Debate team configuration:
  - 5 positions (Analyst, Proposer, Critic, Synthesizer, Mediator)
  - 3 LLMs per position (1 primary + 2 fallbacks)
  - Total: 15 LLMs in the ensemble
- Dynamic selection via LLMsVerifier scores
- OAuth providers priority (Claude, Qwen)
- Scoring algorithm (5 weighted components):
  - ResponseSpeed (25%)
  - ModelEfficiency (20%)
  - CostEffectiveness (25%)
  - Capability (20%)
  - Recency (10%)

#### 14.2 Multi-Pass Validation System (25 min)
- Validation phases:
  1. INITIAL RESPONSE - Initial perspectives
  2. VALIDATION - Cross-validation for accuracy
  3. POLISH & IMPROVE - Refinement based on feedback
  4. FINAL CONCLUSION - Synthesized consensus
- Configuration:
  ```json
  {
    "enable_multi_pass_validation": true,
    "validation_config": {
      "enable_validation": true,
      "enable_polish": true,
      "validation_timeout": 120,
      "polish_timeout": 60,
      "min_confidence_to_skip": 0.9,
      "max_validation_rounds": 3
    }
  }
  ```
- Quality improvement metrics
- Confidence scoring

#### 14.3 Debate Orchestrator Framework (25 min)
- New framework architecture:
  - Agent Pool
  - Team Building
  - Protocol Manager
  - Knowledge Repository
- Topologies:
  - Mesh (parallel)
  - Star (hub-spoke)
  - Chain (sequential)
- Phase-based protocol:
  - Proposal -> Critique -> Review -> Synthesis
- Learning system and cross-debate knowledge

#### 14.4 Integration with 20+ CLI Agents (15 min)
- CLI agent registry (20 agents):
  - OpenCode, Crush, HelixCode, Kiro
  - Aider, ClaudeCode, Cline, CodenameGoose
  - DeepSeekCLI, Forge, GeminiCLI, GPTEngineer
  - KiloCode, MistralCode, OllamaCode, Plandex
  - QwenCode, AmazonQ, CursorAI, Windsurf
- Agent-specific configurations
- X-CLI-Agent header support
- User-Agent pattern matching

### Hands-On Lab
- Configure a 15 LLM debate team
- Enable multi-pass validation
- Test different topologies
- Verify agent integration

---

## Bonus Content

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

## Certification Path

### Level 1: HelixAgent Fundamentals
- Modules 1-3
- Basic installation and configuration
- Assessment: Written quiz + Lab exercise

### Level 2: Provider Expert
- Modules 4-6
- Multi-provider and ensemble mastery
- Assessment: Implementation project

### Level 3: Advanced Practitioner
- Modules 7-9
- Plugin development and optimization
- Assessment: Custom plugin submission

### Level 4: Master Engineer
- Modules 10-11
- Security and CI/CD mastery
- Assessment: Full production deployment review

### Level 5: Challenge Expert (NEW)
- Modules 12-14
- Challenge system mastery and advanced AI debate
- Assessment: 100% pass rate on all challenge scripts
- Requirements:
  - Run RAGS, MCPS, and SKILLS challenges successfully
  - Configure 15 LLM debate team
  - Demonstrate MCP Tool Search integration
  - Document strict validation methodology

---

## Course Resources

### Downloadable Materials
- Complete configuration examples
- API reference sheets
- Best practices checklists
- Troubleshooting guides
- Code templates and snippets

### Support Channels
- Course discussion forums
- GitHub Issues
- Community Discord
- Office hours sessions

---

*Course Version: 3.0.0*
*Last Updated: January 2026*
*Total Duration: 14+ hours*
*Modules: 14 (including 3 new challenge and advanced modules)*
