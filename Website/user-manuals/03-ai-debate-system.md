# AI Debate System Guide

## Introduction

SuperAgent's AI Debate System is a groundbreaking feature that enables multiple AI participants to engage in structured debates, analyze topics from different perspectives, and reach consensus through collaborative reasoning. This system leverages the diverse strengths of multiple LLM providers to produce more balanced, thoroughly considered responses.

The debate system is particularly valuable for complex decision-making, fact verification, creative brainstorming, and scenarios where multiple perspectives improve outcome quality.

---

## Table of Contents

1. [Concept Overview](#concept-overview)
2. [Architecture](#architecture)
3. [Configuration](#configuration)
4. [Participants and Roles](#participants-and-roles)
5. [Debate Rounds and Flow](#debate-rounds-and-flow)
6. [Consensus Mechanisms](#consensus-mechanisms)
7. [Cognee Knowledge Integration](#cognee-knowledge-integration)
8. [API Usage](#api-usage)
9. [Advanced Configurations](#advanced-configurations)
10. [Monitoring and Analytics](#monitoring-and-analytics)
11. [Best Practices](#best-practices)
12. [Troubleshooting](#troubleshooting)

---

## Concept Overview

### What is AI Debate?

The AI Debate System orchestrates multiple AI models to:

1. **Propose**: Generate initial responses to a topic or question
2. **Critique**: Analyze and provide feedback on other participants' responses
3. **Defend**: Respond to critiques and refine positions
4. **Synthesize**: Combine insights to reach a consensus or final answer

### Benefits

| Benefit | Description |
|---------|-------------|
| **Reduced Bias** | Multiple perspectives minimize individual model biases |
| **Higher Quality** | Collaborative refinement improves response quality |
| **Fact Verification** | Cross-checking between models improves accuracy |
| **Creative Solutions** | Diverse viewpoints generate more innovative ideas |
| **Transparency** | Debate transcript provides reasoning audit trail |

### Use Cases

- **Complex Analysis**: Research questions, market analysis, technical decisions
- **Content Validation**: Fact-checking, claim verification, source analysis
- **Creative Projects**: Brainstorming, story development, design critique
- **Decision Support**: Risk assessment, pros/cons analysis, recommendations
- **Educational**: Learning scenarios with multiple teaching perspectives

---

## Architecture

### System Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                        AI Debate Orchestrator                        │
├─────────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐               │
│  │ Debate       │  │  Round       │  │  Consensus   │               │
│  │ Manager      │  │  Controller  │  │  Engine      │               │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘               │
│         │                 │                 │                        │
│  ┌──────▼─────────────────▼─────────────────▼───────┐               │
│  │                Participant Pool                    │               │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌───────┐│               │
│  │  │Claude   │  │DeepSeek │  │Gemini   │  │Qwen   ││               │
│  │  │Debater  │  │Debater  │  │Debater  │  │Debater││               │
│  │  └─────────┘  └─────────┘  └─────────┘  └───────┘│               │
│  └──────────────────────────────────────────────────┘               │
│         │                                                            │
│  ┌──────▼──────────────────────────────────────────┐                │
│  │              Knowledge Integration               │                │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐         │                │
│  │  │ Cognee  │  │ Memory  │  │ Context │         │                │
│  │  │ Graph   │  │ Store   │  │ Builder │         │                │
│  │  └─────────┘  └─────────┘  └─────────┘         │                │
│  └─────────────────────────────────────────────────┘                │
└─────────────────────────────────────────────────────────────────────┘
```

### Core Components

| Component | Responsibility |
|-----------|----------------|
| **Debate Manager** | Initializes debates, manages lifecycle |
| **Round Controller** | Orchestrates debate rounds and turn-taking |
| **Consensus Engine** | Analyzes responses and determines consensus |
| **Participant Pool** | Manages AI debater instances and assignments |
| **Knowledge Integration** | Connects to Cognee for context and memory |

---

## Configuration

### Basic Configuration

```yaml
# configs/development.yaml
ai_debate:
  # Enable/disable the debate system
  enabled: true

  # Maximum number of debate participants
  max_participants: 5

  # Maximum number of debate rounds
  max_rounds: 3

  # Consensus threshold (0.0 - 1.0)
  consensus_threshold: 0.7

  # Timeout for individual debate rounds
  round_timeout: "60s"

  # Overall debate timeout
  debate_timeout: "300s"
```

### Advanced Configuration

```yaml
ai_debate:
  enabled: true
  max_participants: 5
  max_rounds: 5
  consensus_threshold: 0.7

  # Debate behavior settings
  behavior:
    # Allow participants to skip rounds if they agree
    allow_skip: true

    # Require minimum participation per round
    min_contributions_per_round: 2

    # Enable parallel response generation
    parallel_execution: true

    # Enable moderator role
    moderator_enabled: true

    # Temperature adjustment per round
    temperature_decay: 0.1  # Reduce temperature each round

  # Scoring and weights
  scoring:
    # Weight for relevance to topic
    relevance_weight: 0.3

    # Weight for argument quality
    quality_weight: 0.25

    # Weight for supporting evidence
    evidence_weight: 0.2

    # Weight for novelty of contribution
    novelty_weight: 0.15

    # Weight for agreement with other participants
    agreement_weight: 0.1

  # Cognee integration
  cognee:
    enabled: true
    api_url: "http://cognee:8000"
    dataset_name: "debate_knowledge"

    # Knowledge retrieval settings
    retrieval:
      max_results: 10
      relevance_threshold: 0.7
      include_graph: true

    # Memory settings
    memory:
      store_debates: true
      retention_days: 90
      enable_learning: true

  # Logging and analytics
  analytics:
    enabled: true
    store_transcripts: true
    calculate_metrics: true
```

---

## Participants and Roles

### Participant Types

The debate system supports various participant roles:

#### Proposer

Initiates discussion by generating initial response to the topic.

```yaml
participants:
  - name: "claude_proposer"
    role: "proposer"
    provider: "claude"
    model: "claude-3-sonnet-20240229"
    settings:
      temperature: 0.8
      max_tokens: 1000
      system_prompt: |
        You are a thoughtful proposer in a debate. Present your initial
        perspective on the topic with clear reasoning and supporting arguments.
```

#### Critic

Analyzes and critiques other participants' arguments.

```yaml
participants:
  - name: "deepseek_critic"
    role: "critic"
    provider: "deepseek"
    model: "deepseek-chat"
    settings:
      temperature: 0.6
      max_tokens: 800
      system_prompt: |
        You are a constructive critic in a debate. Identify weaknesses,
        logical fallacies, and areas for improvement in other arguments.
        Be respectful but thorough in your analysis.
```

#### Defender

Responds to critiques and refines positions.

```yaml
participants:
  - name: "gemini_defender"
    role: "defender"
    provider: "gemini"
    model: "gemini-pro"
    settings:
      temperature: 0.7
      max_tokens: 900
      system_prompt: |
        You are a skilled defender in a debate. Address critiques directly,
        provide additional evidence, and strengthen arguments against objections.
```

#### Synthesizer

Combines insights from all participants to form consensus.

```yaml
participants:
  - name: "qwen_synthesizer"
    role: "synthesizer"
    provider: "qwen"
    model: "qwen-max"
    settings:
      temperature: 0.5
      max_tokens: 1200
      system_prompt: |
        You are a neutral synthesizer in a debate. Your role is to identify
        common ground, integrate the best ideas from all participants, and
        formulate a balanced conclusion.
```

#### Moderator

Guides the debate, ensures fairness, and manages flow.

```yaml
participants:
  - name: "claude_moderator"
    role: "moderator"
    provider: "claude"
    model: "claude-3-opus-20240229"
    settings:
      temperature: 0.4
      max_tokens: 500
      system_prompt: |
        You are an impartial moderator. Guide the debate, ensure all
        participants are heard, redirect off-topic discussions, and
        summarize key points between rounds.
```

### Full Participant Configuration Example

```yaml
ai_debate:
  participants:
    # Primary proposer
    - name: "primary_analyst"
      role: "proposer"
      provider: "claude"
      model: "claude-3-sonnet-20240229"
      weight: 1.0
      settings:
        temperature: 0.7
        max_tokens: 1000

    # Alternative proposer
    - name: "alternative_analyst"
      role: "proposer"
      provider: "deepseek"
      model: "deepseek-chat"
      weight: 0.9
      settings:
        temperature: 0.8
        max_tokens: 1000

    # Critical reviewer
    - name: "critical_reviewer"
      role: "critic"
      provider: "gemini"
      model: "gemini-pro"
      weight: 0.85
      settings:
        temperature: 0.6
        max_tokens: 800

    # Expert synthesizer
    - name: "expert_synthesizer"
      role: "synthesizer"
      provider: "claude"
      model: "claude-3-opus-20240229"
      weight: 1.0
      settings:
        temperature: 0.5
        max_tokens: 1500

    # Debate moderator
    - name: "debate_moderator"
      role: "moderator"
      provider: "claude"
      model: "claude-3-sonnet-20240229"
      weight: 1.0
      settings:
        temperature: 0.4
        max_tokens: 600
```

---

## Debate Rounds and Flow

### Standard Debate Flow

```
┌─────────────────────────────────────────────────────────────────────┐
│                         DEBATE LIFECYCLE                             │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐      │
│  │  ROUND 1 │───▶│  ROUND 2 │───▶│  ROUND 3 │───▶│ CONSENSUS│      │
│  │ Initial  │    │ Critique │    │ Refine   │    │ Synthesis│      │
│  │ Proposals│    │ & Defend │    │ & Defend │    │ & Vote   │      │
│  └──────────┘    └──────────┘    └──────────┘    └──────────┘      │
│       │               │               │               │             │
│       ▼               ▼               ▼               ▼             │
│  All participants  Critics analyze  Defenders refine  Synthesizer  │
│  propose initial   proposals,       arguments,        creates final│
│  responses        defenders         critics respond   consensus    │
│                   respond                                           │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### Round Configuration

```yaml
debate_rounds:
  # Round 1: Initial Proposals
  round_1:
    name: "Initial Proposals"
    type: "proposal"
    active_roles: ["proposer"]
    timeout: "60s"
    settings:
      allow_parallel: true
      require_all_participants: true

  # Round 2: Critique and Response
  round_2:
    name: "Critique Phase"
    type: "critique"
    active_roles: ["critic", "defender"]
    timeout: "90s"
    settings:
      critique_targets: "all_proposals"
      allow_rebuttals: true

  # Round 3: Refinement
  round_3:
    name: "Refinement Phase"
    type: "refinement"
    active_roles: ["proposer", "defender"]
    timeout: "60s"
    settings:
      incorporate_feedback: true
      temperature_reduction: 0.1

  # Final: Synthesis
  final_round:
    name: "Consensus Synthesis"
    type: "synthesis"
    active_roles: ["synthesizer", "moderator"]
    timeout: "120s"
    settings:
      produce_summary: true
      vote_on_consensus: true
```

### Round Execution API

```bash
# Start a debate
curl -X POST http://localhost:8080/v1/debates \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "Should AI systems be required to explain their decisions?",
    "participants": ["claude", "deepseek", "gemini"],
    "max_rounds": 3,
    "consensus_threshold": 0.7
  }'

# Response
{
  "debate_id": "debate_abc123",
  "status": "in_progress",
  "current_round": 1,
  "participants": [
    {"name": "claude_proposer", "provider": "claude", "role": "proposer"},
    {"name": "deepseek_critic", "provider": "deepseek", "role": "critic"},
    {"name": "gemini_defender", "provider": "gemini", "role": "defender"}
  ]
}
```

---

## Consensus Mechanisms

### Available Strategies

SuperAgent supports multiple consensus strategies:

#### 1. Confidence-Weighted Voting

Aggregates responses weighted by model confidence scores.

```yaml
consensus:
  strategy: "confidence_weighted"
  settings:
    confidence_threshold: 0.7
    tie_breaker: "highest_confidence"
    require_supermajority: false
```

#### 2. Majority Vote

Simple majority wins approach.

```yaml
consensus:
  strategy: "majority_vote"
  settings:
    minimum_agreement: 0.51
    abstention_handling: "exclude"
```

#### 3. Unanimous Agreement

Requires all participants to agree.

```yaml
consensus:
  strategy: "unanimous"
  settings:
    allow_minor_disagreements: true
    similarity_threshold: 0.9
```

#### 4. Quality-Weighted

Weights votes by response quality scores.

```yaml
consensus:
  strategy: "quality_weighted"
  settings:
    quality_metrics:
      - coherence
      - relevance
      - evidence_support
      - logical_consistency
    weights:
      coherence: 0.25
      relevance: 0.3
      evidence_support: 0.25
      logical_consistency: 0.2
```

#### 5. Synthesized Consensus

Uses a designated synthesizer to create unified response.

```yaml
consensus:
  strategy: "synthesized"
  settings:
    synthesizer_provider: "claude"
    synthesizer_model: "claude-3-opus-20240229"
    include_minority_views: true
    highlight_disagreements: true
```

### Consensus Response Format

```json
{
  "consensus_reached": true,
  "confidence": 0.87,
  "strategy_used": "confidence_weighted",
  "final_response": "Based on the debate, AI systems should...",
  "agreement_breakdown": {
    "claude": {
      "vote": "agree",
      "confidence": 0.92,
      "key_points": ["transparency", "accountability"]
    },
    "deepseek": {
      "vote": "agree",
      "confidence": 0.85,
      "key_points": ["user trust", "debugging"]
    },
    "gemini": {
      "vote": "partial_agree",
      "confidence": 0.78,
      "key_points": ["complexity concerns", "performance trade-offs"]
    }
  },
  "minority_views": [
    "Performance impact may be significant for real-time systems"
  ],
  "debate_summary": "All participants agreed on the principle while...",
  "transcript_id": "transcript_xyz789"
}
```

---

## Cognee Knowledge Integration

### Overview

Cognee provides knowledge graph and memory capabilities to enhance debates:

- **Historical Context**: Reference past debates and decisions
- **Fact Grounding**: Verify claims against knowledge base
- **Entity Awareness**: Understand relationships and context
- **Memory Persistence**: Learn from debate outcomes

### Configuration

```yaml
ai_debate:
  cognee:
    enabled: true
    api_url: "http://cognee:8000"

    # Knowledge retrieval
    retrieval:
      enabled: true
      sources:
        - type: "knowledge_graph"
          priority: 1
        - type: "vector_search"
          priority: 2
        - type: "historical_debates"
          priority: 3

      settings:
        max_results: 10
        relevance_threshold: 0.7
        include_metadata: true

    # Memory storage
    memory:
      enabled: true
      store_debates: true
      store_outcomes: true
      store_participants: true

      indexing:
        extract_entities: true
        build_relationships: true
        calculate_embeddings: true

    # Learning from debates
    learning:
      enabled: true
      update_knowledge_graph: true
      track_accuracy: true
      feedback_loop: true
```

### Using Cognee in Debates

```bash
# Create debate with knowledge context
curl -X POST http://localhost:8080/v1/debates \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "How should we approach renewable energy transition?",
    "participants": ["claude", "deepseek"],
    "cognee": {
      "enabled": true,
      "retrieve_context": true,
      "context_query": "renewable energy policies, climate change impacts",
      "max_context_items": 5
    }
  }'
```

---

## API Usage

### Starting a Debate

```bash
curl -X POST http://localhost:8080/v1/debates \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "What is the best approach to AI safety?",
    "context": "Consider both technical and governance approaches",
    "participants": [
      {
        "provider": "claude",
        "role": "proposer",
        "model": "claude-3-sonnet-20240229"
      },
      {
        "provider": "deepseek",
        "role": "critic",
        "model": "deepseek-chat"
      },
      {
        "provider": "gemini",
        "role": "synthesizer",
        "model": "gemini-pro"
      }
    ],
    "settings": {
      "max_rounds": 3,
      "consensus_threshold": 0.7,
      "consensus_strategy": "synthesized"
    }
  }'
```

### Getting Debate Status

```bash
curl http://localhost:8080/v1/debates/debate_abc123 \
  -H "Authorization: Bearer $TOKEN"
```

### Retrieving Transcript

```bash
curl http://localhost:8080/v1/debates/debate_abc123/transcript \
  -H "Authorization: Bearer $TOKEN"
```

### Streaming Debate Progress

```bash
curl -N http://localhost:8080/v1/debates/debate_abc123/stream \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept: text/event-stream"
```

---

## Advanced Configurations

### Custom Debate Templates

```yaml
debate_templates:
  # Technical analysis template
  technical_analysis:
    participants:
      - role: "technical_expert"
        count: 2
        providers: ["claude", "deepseek"]
      - role: "critic"
        count: 1
        providers: ["gemini"]
      - role: "synthesizer"
        count: 1
        providers: ["claude"]
    rounds: 4
    consensus_strategy: "quality_weighted"

  # Creative brainstorming template
  brainstorming:
    participants:
      - role: "idea_generator"
        count: 3
        providers: ["claude", "gemini", "qwen"]
      - role: "evaluator"
        count: 1
        providers: ["deepseek"]
    rounds: 2
    consensus_strategy: "synthesized"
    settings:
      temperature_boost: 0.2
      encourage_novelty: true
```

### Dynamic Participant Adjustment

```yaml
dynamic_adjustment:
  enabled: true

  # Add participants if debate stalls
  add_on_stall:
    enabled: true
    stall_rounds: 2
    new_participant:
      role: "mediator"
      provider: "claude"

  # Remove low-contribution participants
  remove_inactive:
    enabled: true
    inactivity_threshold: 1
    minimum_participants: 2
```

---

## Monitoring and Analytics

### Debate Metrics

```bash
# Get debate metrics
curl http://localhost:8080/v1/debates/metrics \
  -H "Authorization: Bearer $TOKEN"
```

Response:
```json
{
  "total_debates": 1523,
  "consensus_rate": 0.87,
  "average_rounds": 2.4,
  "average_duration_seconds": 145,
  "provider_performance": {
    "claude": {"participation": 1200, "synthesis_quality": 0.92},
    "deepseek": {"participation": 1100, "critique_quality": 0.88},
    "gemini": {"participation": 900, "defense_quality": 0.85}
  }
}
```

### Analytics Dashboard Configuration

```yaml
analytics:
  enabled: true

  metrics:
    - name: "debate_consensus_rate"
      type: "gauge"
    - name: "debate_rounds_total"
      type: "counter"
    - name: "debate_duration_seconds"
      type: "histogram"
    - name: "participant_contribution_score"
      type: "gauge"

  dashboards:
    grafana:
      enabled: true
      refresh_interval: "30s"
```

---

## Best Practices

### 1. Participant Selection

- **Diversity**: Include providers with different training approaches
- **Complementary Roles**: Mix analytical and creative participants
- **Quality Focus**: Use higher-quality models for synthesis roles

### 2. Round Configuration

- **Start Broad**: Allow open proposals in early rounds
- **Narrow Focus**: Increase specificity in later rounds
- **Temperature Decay**: Reduce temperature to converge on consensus

### 3. Consensus Strategy

- **Complex Topics**: Use synthesized consensus with strong synthesizer
- **Factual Questions**: Use majority vote with fact-checking
- **Creative Tasks**: Use quality-weighted to reward innovation

### 4. Performance Optimization

- **Parallel Execution**: Enable parallel responses where possible
- **Caching**: Cache common debate patterns
- **Timeouts**: Set appropriate timeouts to prevent runaway debates

---

## Troubleshooting

### Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| No consensus reached | Low threshold or divergent participants | Increase rounds or adjust threshold |
| Slow debates | Sequential execution | Enable parallel_execution |
| Repetitive responses | High temperature or no feedback | Enable feedback incorporation |
| Participant timeout | Provider issues | Check provider health, increase timeout |

### Debug Mode

```yaml
ai_debate:
  debug:
    enabled: true
    log_level: "debug"
    store_intermediate_states: true
    trace_decision_path: true
```

---

## Summary

The AI Debate System is a powerful feature for generating higher-quality, more balanced AI responses through collaborative reasoning. Key points:

1. **Configure Participants**: Choose diverse providers with complementary roles
2. **Define Rounds**: Structure debate flow for optimal convergence
3. **Select Consensus Strategy**: Match strategy to use case requirements
4. **Integrate Knowledge**: Use Cognee for context and memory
5. **Monitor Performance**: Track metrics for continuous improvement

Continue to the [API Reference](04-api-reference.md) for complete endpoint documentation.
