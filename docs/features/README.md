# Feature Documentation

This directory contains documentation for HelixAgent's core features and advanced capabilities.

## Overview

HelixAgent provides a comprehensive set of features for AI orchestration, including ensemble strategies, AI debate systems, RAG pipelines, memory management, security frameworks, and more. This index provides navigation to detailed documentation for each feature.

## Documentation Index

### AI Debate System

| Document | Description |
|----------|-------------|
| [AI Debate Orchestrator](./AI_DEBATE_ORCHESTRATOR.md) | Multi-agent debate framework with topologies and protocols |
| [AI Debate Configuration](./ai-debate-configuration.md) | Configuration options for AI debates |
| [Advanced Features Summary](./ADVANCED_FEATURES_SUMMARY.md) | Summary of advanced debate strategies and consensus algorithms |
| [Advanced Features Implementation](./ADVANCED_FEATURES_IMPLEMENTATION_SUMMARY.md) | Implementation status of all advanced features |
| [Phase 8: AI Agent Orchestration](./PHASE_8_AI_AGENT_ORCHESTRATION.md) | Future roadmap for agent orchestration |

### Intelligence Systems

| Document | Description |
|----------|-------------|
| [Semantic Routing](./SEMANTIC_ROUTING.md) | Embedding-based query routing system |
| [RAG System](./RAG_SYSTEM.md) | Retrieval Augmented Generation with hybrid retrieval |
| [Memory System](./MEMORY_SYSTEM.md) | Mem0-style persistent AI memory with entity graphs |

### Security and Testing

| Document | Description |
|----------|-------------|
| [Security Framework](./SECURITY_FRAMEWORK.md) | Red team framework, guardrails, PII detection |
| [LLM Testing](./LLM_TESTING.md) | DeepEval-style testing with RAGAS metrics |

## Feature Highlights

### AI Debate System

The AI Debate Orchestrator enables sophisticated multi-agent consensus building:

**Topologies**:
- **Mesh**: All agents communicate with all (best for small groups, complex topics)
- **Star**: Central moderator coordinates (best for large groups, structured debates)
- **Chain**: Sequential communication (best for step-by-step analysis)

**Agent Roles**:
- **Moderator**: Guides debate flow
- **Analyst**: Data-driven analysis
- **Critic**: Challenges assumptions
- **Synthesizer**: Combines perspectives
- **Expert**: Domain-specific knowledge

**Debate Strategies**:
- Socratic Method
- Devil's Advocate
- Consensus Building
- Evidence-Based
- Creative Synthesis
- Adversarial Testing

**Consensus Algorithms**:
- Weighted Average
- Median Consensus
- Fuzzy Logic
- Bayesian Inference

### Ensemble Strategies

HelixAgent supports multiple ensemble strategies for combining LLM responses:

| Strategy | Description | Use Case |
|----------|-------------|----------|
| Majority Vote | Simple voting across providers | Quick consensus |
| Confidence Weighted | Weight by provider confidence scores | Quality-focused |
| Quality Weighted | Weight by response quality metrics | Accuracy-focused |
| Best-of-N | Select highest quality single response | Performance-focused |

### Semantic Routing

Intelligent query routing based on embedding similarity:

```go
router := semantic.NewRouter(&semantic.Config{
    EmbeddingProvider:   embeddingProvider,
    SimilarityThreshold: 0.7,
    TopK:                3,
})

router.AddRoute(&semantic.Route{
    Name:        "code-review",
    Description: "Code review requests",
    Utterances:  []string{"review this code", "check for bugs"},
    Handler:     codeReviewHandler,
})
```

### RAG System

Hybrid retrieval combining dense and sparse methods:

- **Dense Retrieval**: Vector similarity with embedding models
- **Sparse Retrieval**: BM25 keyword matching
- **Reranking**: Cross-encoder reranking for precision
- **Chunking**: Intelligent document chunking with overlap

### Memory System

Mem0-style persistent memory with multiple memory types:

- **Episodic Memory**: Conversation events and experiences
- **Semantic Memory**: Facts and knowledge
- **Procedural Memory**: How-to information
- **Working Memory**: Short-term context

Features entity extraction and relationship graphs for connected memory.

### Security Framework

Comprehensive security with:

- **Red Team Framework**: 40+ attack types for vulnerability testing
- **Guardrails**: Input/output filtering pipeline
- **PII Detection**: Automatic PII identification and masking
- **MCP Security**: Tool execution security controls
- **Audit Logging**: Complete audit trail

### LLM Testing

DeepEval-style testing capabilities:

- **RAGAS Metrics**: Faithfulness, answer relevancy, context precision
- **Custom Assertions**: Define custom evaluation criteria
- **Judge LLM**: Use LLM as evaluator for subjective metrics
- **Batch Evaluation**: Parallel test execution

## Configuration Examples

### Ensemble Mode

```yaml
ensemble:
  strategy: confidence_weighted
  providers:
    - claude
    - deepseek
    - gemini
  min_responses: 2
  timeout: 60s
```

### AI Debate

```yaml
debate:
  enabled: true
  max_rounds: 3
  participants: 5
  strategy: structured
  voting: confidence_weighted
  consensus_threshold: 0.75
  timeout: 300s
```

### Semantic Routing

```yaml
routing:
  semantic:
    enabled: true
    similarity_threshold: 0.7
    top_k: 3
    embedding_model: text-embedding-ada-002
```

### RAG Pipeline

```yaml
rag:
  dense_weight: 0.7
  sparse_weight: 0.3
  initial_k: 20
  final_k: 5
  reranker: cross-encoder
```

### Memory System

```yaml
memory:
  enabled: true
  retention_days: 30
  max_entries: 10000
  entity_extraction: true
  relationship_tracking: true
```

### Security

```yaml
security:
  guardrails:
    enabled: true
    input_filtering: true
    output_filtering: true
  pii_detection:
    enabled: true
    mask_pii: true
  audit_logging:
    enabled: true
    log_level: detailed
```

## API Endpoints

| Endpoint | Description |
|----------|-------------|
| `/v1/chat/completions` | OpenAI-compatible chat endpoint |
| `/v1/embeddings` | Generate embeddings |
| `/v1/debate/start` | Start AI debate session |
| `/v1/debate/status` | Check debate status |
| `/v1/memory/store` | Store memory entry |
| `/v1/memory/search` | Search memories |
| `/v1/rag/query` | Execute RAG query |

## Challenge Scripts

Validate features with challenge scripts:

```bash
# AI Debate challenges
./challenges/scripts/debate_team_dynamic_selection_challenge.sh

# Semantic routing challenges
./challenges/scripts/semantic_intent_challenge.sh

# Full system validation
./challenges/scripts/full_system_boot_challenge.sh

# Security scanning
./challenges/scripts/security_scanning_challenge.sh
```

## Related Documentation

- [Architecture Overview](../architecture/README.md)
- [API Reference](../api/README.md)
- [Configuration Guide](../configuration/README.md)
- [Integration Guide](../integrations/README.md)
