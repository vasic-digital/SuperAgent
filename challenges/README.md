# SuperAgent Challenges System

A comprehensive challenge framework for testing, verifying, and validating LLM providers, AI debate groups, and API quality.

## Overview

The SuperAgent Challenges System provides:

- **Automated Provider Verification**: Test and score all configured LLM providers
- **AI Debate Group Formation**: Create optimized groups of top-performing models
- **API Quality Testing**: Validate response quality with comprehensive assertions
- **Comprehensive Logging**: Full audit trail of all API communications
- **Execution Reports**: Detailed reports and master summaries

## Quick Start

```bash
# 1. Copy environment template and configure API keys
cp .env.example .env
# Edit .env with your actual API keys

# 2. Run the AI Debate Group Formation challenge
./scripts/run_challenges.sh ai_debate_formation

# 3. View results
cat master_results/master_summary_*.md
```

## Directory Structure

```
challenges/
├── data/                     # Challenge registry and data
├── docs/                     # Comprehensive documentation
├── codebase/                 # Challenge implementations
│   ├── challenge_runners/    # Shell script runners
│   └── go_files/             # Go implementations
├── results/                  # Timestamped execution results
├── master_results/           # Master summary reports
├── scripts/                  # Utility scripts
├── config/                   # Configuration files
└── *.go                      # Core framework code
```

## Available Challenges

| Challenge | Description | Category |
|-----------|-------------|----------|
| `provider_verification` | Verify all LLM providers and score models | Core |
| `ai_debate_formation` | Form AI debate group from top models | Core |
| `api_quality_test` | Test API quality with assertions | Validation |

## Documentation

- [00_INDEX.md](docs/00_INDEX.md) - Documentation index
- [01_INTRODUCTION.md](docs/01_INTRODUCTION.md) - Framework introduction
- [02_QUICK_START.md](docs/02_QUICK_START.md) - Quick start guide
- [03_CHALLENGE_CATALOG.md](docs/03_CHALLENGE_CATALOG.md) - All challenges
- [04_AI_DEBATE_GROUP.md](docs/04_AI_DEBATE_GROUP.md) - Debate group details
- [05_SECURITY.md](docs/05_SECURITY.md) - Security practices

## Security

**IMPORTANT**: Never commit API keys or sensitive information!

- `.env` - Contains actual API keys (git-ignored)
- `.env.example` - Template with placeholder values (git-versioned)
- `config/config.yaml` - Actual configuration (git-ignored)
- `config/config.yaml.example` - Template configuration (git-versioned)

All generated reports and logs automatically redact sensitive information.

## License

Part of the SuperAgent project.
