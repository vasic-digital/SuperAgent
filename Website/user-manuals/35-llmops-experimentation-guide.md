# User Manual 35: LLMOps Experimentation Guide

## Overview

HelixAgent's LLMOps system provides continuous evaluation, A/B experimentation, and prompt versioning for LLM operations. Run experiments to compare providers, models, and prompt strategies with real data.

## API Endpoints

### Create an Experiment

```
POST /v1/llmops/experiments
```

```json
{
  "name": "claude-vs-deepseek-coding",
  "description": "Compare Claude and DeepSeek for code generation",
  "variants": [
    {"name": "claude-3", "provider": "claude", "model": "claude-3-sonnet", "weight": 50},
    {"name": "deepseek-v2", "provider": "deepseek", "model": "deepseek-coder", "weight": 50}
  ]
}
```

### List Experiments

```
GET /v1/llmops/experiments
GET /v1/llmops/experiments?status=running
```

### Run Continuous Evaluation

```
POST /v1/llmops/evaluate
```

```json
{
  "name": "weekly-quality-check",
  "dataset": "coding-benchmarks-v2",
  "metrics": ["accuracy", "latency", "cost"]
}
```

### Prompt Versioning

```
POST /v1/llmops/prompts
```

```json
{
  "name": "code-review-prompt",
  "version": "2.1.0",
  "content": "Review the following code for security vulnerabilities...",
  "metadata": {"author": "team-security", "tags": ["security", "review"]}
}
```

```
GET /v1/llmops/prompts
```

## Key Concepts

- **Experiments**: A/B tests with weighted traffic split across variants
- **Evaluations**: Automated quality scoring against reference datasets
- **Prompt Versions**: Immutable prompt templates with semantic versioning
- **Metrics**: latency, accuracy, cost, token usage tracked per variant

## Best Practices

1. Start with 50/50 splits, then adjust based on results
2. Run evaluations for at least 100 samples for statistical significance
3. Version prompts semantically (major.minor.patch)
4. Use evaluation datasets that represent real production traffic
