# HelixAgent Challenges - Documentation Index

## Overview Documents

| Document | Description |
|----------|-------------|
| [01_INTRODUCTION.md](01_INTRODUCTION.md) | Framework overview and architecture |
| [02_QUICK_START.md](02_QUICK_START.md) | 5-minute quick start guide |
| [03_CHALLENGE_CATALOG.md](03_CHALLENGE_CATALOG.md) | Complete challenge specifications |
| [04_AI_DEBATE_GROUP.md](04_AI_DEBATE_GROUP.md) | AI Debate Group formation details |
| [05_SECURITY.md](05_SECURITY.md) | Security practices and guidelines |
| [07_CONTAINER_NAMING.md](07_CONTAINER_NAMING.md) | Container naming & cleanup conventions |

## Quick Links

### Getting Started
1. Copy `.env.example` to `.env`
2. Configure your API keys
3. Run `./scripts/run_challenges.sh ai_debate_formation`

### Key Concepts
- **Challenge**: A testable verification task with defined inputs and outputs
- **Debate Group**: Collection of top-scoring LLMs working together
- **Verification**: Process of validating model capabilities
- **Scoring**: Quantitative assessment of model performance

### Challenge Categories

| Category | Description | Challenges |
|----------|-------------|------------|
| Core | Essential verification and formation | 2 |
| Validation | API quality and response testing | 1 |
| Integration | LLMsVerifier integration tests | 1 |

### Results Structure

```
results/[challenge]/[YYYY]/[MM]/[DD]/[timestamp]/
├── logs/           # Execution logs
├── results/        # Output files
└── config/         # Configuration used
```

## Related Documentation

- [IMPLEMENTATION_PLAN.md](../IMPLEMENTATION_PLAN.md) - Implementation roadmap
- [README.md](../README.md) - Project overview
