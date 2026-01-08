# HelixAgent Challenges - AI Debate Group

Comprehensive documentation for the AI Debate Group formation and operation.

## Overview

The AI Debate Group is HelixAgent's core ensemble mechanism, combining multiple top-performing LLMs into a single virtual model. This document details how groups are formed, configured, and operated.

## Group Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      AI DEBATE GROUP                             │
│                    (Virtual LLM Endpoint)                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │  Primary 1   │  │  Primary 2   │  │  Primary 3   │  ...     │
│  │  (Score 9.5) │  │  (Score 9.3) │  │  (Score 9.1) │          │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘          │
│         │                 │                 │                    │
│    ┌────┴────┐       ┌────┴────┐       ┌────┴────┐              │
│    │Fallback │       │Fallback │       │Fallback │              │
│    │  1a/1b  │       │  2a/2b  │       │  3a/3b  │              │
│    └─────────┘       └─────────┘       └─────────┘              │
│                                                                  │
│  Total: 5 Primaries + 10 Fallbacks = 15 Models                  │
└─────────────────────────────────────────────────────────────────┘
```

## Formation Process

### Step 1: Model Discovery

All configured providers are queried for available models:

```json
{
  "providers_discovered": 15,
  "models_discovered": 127,
  "providers": [
    {"name": "anthropic", "models": 8},
    {"name": "openai", "models": 12},
    {"name": "openrouter", "models": 50},
    ...
  ]
}
```

### Step 2: Verification

Each model undergoes verification:

1. **Connectivity Test**: Can we reach the API?
2. **Authentication Test**: Is the API key valid?
3. **Capability Test**: What features does it support?
4. **Code Visibility Test**: "Can you see my code?"

```json
{
  "model_id": "anthropic/claude-3-opus",
  "verification": {
    "connectivity": true,
    "authentication": true,
    "capabilities": ["code_generation", "streaming", "function_calling"],
    "code_visibility": true,
    "response_time_ms": 1250
  }
}
```

### Step 3: Scoring

Models are scored on multiple criteria (0-10 scale):

| Criterion | Weight | Description |
|-----------|--------|-------------|
| Response Speed | 0.25 | Latency and throughput |
| Model Efficiency | 0.20 | Context window, parameters |
| Cost Effectiveness | 0.25 | Price per token |
| Capability | 0.20 | Verified features |
| Recency | 0.10 | Model freshness |

**Final Score**: Weighted sum of all criteria

### Step 4: Selection

Models are sorted by score and selected:

```
Rank 1:  anthropic/claude-3-opus     Score: 9.5 → Primary 1
Rank 2:  openai/gpt-4-turbo          Score: 9.3 → Primary 2
Rank 3:  anthropic/claude-3-sonnet   Score: 9.1 → Primary 3
Rank 4:  deepseek/deepseek-coder     Score: 9.0 → Primary 4
Rank 5:  openai/gpt-4o               Score: 8.9 → Primary 5
Rank 6:  gemini/gemini-pro           Score: 8.7 → Fallback 1a
Rank 7:  anthropic/claude-3-haiku    Score: 8.6 → Fallback 1b
...
Rank 15: qwen/qwen-72b               Score: 8.0 → Fallback 5b
```

### Step 5: Fallback Assignment

Fallbacks are assigned based on:
1. **Compatibility**: Similar capabilities to primary
2. **Diversity**: Different provider preferred
3. **Score**: Next highest scores

```json
{
  "primary_1": {
    "model": "anthropic/claude-3-opus",
    "score": 9.5,
    "fallbacks": [
      {"model": "gemini/gemini-pro", "score": 8.7},
      {"model": "anthropic/claude-3-haiku", "score": 8.6}
    ]
  }
}
```

## Group Configuration

### debate_group.json

```json
{
  "id": "dg_20250104_103000",
  "name": "HelixAgent Debate Group",
  "created_at": "2025-01-04T10:30:00Z",
  "members": [
    {
      "position": 1,
      "role": "primary",
      "model": {
        "provider": "anthropic",
        "model_id": "claude-3-opus-20240229",
        "display_name": "Claude 3 Opus",
        "score": 9.5,
        "capabilities": ["code_generation", "streaming", "vision"]
      },
      "fallbacks": [
        {
          "role": "fallback_1",
          "model": {
            "provider": "google",
            "model_id": "gemini-pro",
            "score": 8.7
          }
        },
        {
          "role": "fallback_2",
          "model": {
            "provider": "anthropic",
            "model_id": "claude-3-haiku",
            "score": 8.6
          }
        }
      ]
    }
    // ... 4 more members
  ],
  "total_models": 15,
  "average_score": 8.8,
  "configuration": {
    "debate_rounds": 3,
    "consensus_threshold": 0.7,
    "timeout_seconds": 60,
    "fallback_strategy": "sequential"
  }
}
```

## Debate Operation

### Request Flow

```
User Request
     │
     ▼
┌─────────────────────────────────────────┐
│          Debate Orchestrator            │
└─────────────────────────────────────────┘
     │
     ├──→ Primary 1 ──→ Response 1
     ├──→ Primary 2 ──→ Response 2
     ├──→ Primary 3 ──→ Response 3
     ├──→ Primary 4 ──→ Response 4
     └──→ Primary 5 ──→ Response 5
     │
     ▼
┌─────────────────────────────────────────┐
│         Consensus Engine                 │
│  (Confidence-weighted voting)            │
└─────────────────────────────────────────┘
     │
     ▼
Final Response
```

### Fallback Activation

Fallbacks are activated when:
1. Primary model fails to respond
2. Primary model times out
3. Response confidence below threshold
4. Rate limit exceeded

```
Primary 1 ─[FAIL]──→ Fallback 1a ─[FAIL]──→ Fallback 1b
                          │
                          └──→ Response
```

### Debate Rounds

For complex queries, multiple debate rounds may occur:

**Round 1**: Initial responses from all primaries
**Round 2**: Refinement based on other responses
**Round 3**: Final consensus

## Metrics & Monitoring

### Group Health Metrics

| Metric | Description | Threshold |
|--------|-------------|-----------|
| `group_availability` | % of primaries responding | >= 80% |
| `fallback_rate` | % of requests using fallbacks | <= 20% |
| `avg_response_time` | Average response latency | <= 5s |
| `consensus_rate` | % of requests with consensus | >= 70% |

### Per-Model Metrics

```json
{
  "model_id": "anthropic/claude-3-opus",
  "metrics": {
    "requests_handled": 1250,
    "avg_response_time_ms": 2100,
    "error_rate": 0.02,
    "fallback_triggers": 15,
    "consensus_contribution": 0.35
  }
}
```

## API Integration

### Exposed Endpoint

The debate group is exposed via OpenAI-compatible API:

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "helixagent-debate",
    "messages": [{"role": "user", "content": "Write a Go function"}]
  }'
```

### Response Format

```json
{
  "id": "chatcmpl-abc123",
  "object": "chat.completion",
  "created": 1704362400,
  "model": "helixagent-debate",
  "choices": [{
    "index": 0,
    "message": {
      "role": "assistant",
      "content": "Here is a Go function..."
    },
    "finish_reason": "stop"
  }],
  "usage": {
    "prompt_tokens": 50,
    "completion_tokens": 200,
    "total_tokens": 250
  },
  "debate_metadata": {
    "primaries_used": 5,
    "fallbacks_used": 0,
    "debate_rounds": 2,
    "consensus_score": 0.85,
    "winning_model": "anthropic/claude-3-opus"
  }
}
```

## Best Practices

### Group Composition

1. **Provider Diversity**: Include models from multiple providers
2. **Capability Coverage**: Ensure all needed capabilities are present
3. **Score Balance**: Maintain reasonable score distribution
4. **Fallback Readiness**: Test fallbacks regularly

### Performance Optimization

1. **Parallel Execution**: Query all primaries simultaneously
2. **Early Termination**: Stop if consensus reached early
3. **Caching**: Cache common responses
4. **Streaming**: Use streaming for real-time feedback

### Reliability

1. **Regular Verification**: Re-verify models periodically
2. **Score Refresh**: Update scores based on actual performance
3. **Fallback Testing**: Ensure fallbacks work when needed
4. **Monitoring**: Watch health metrics continuously
