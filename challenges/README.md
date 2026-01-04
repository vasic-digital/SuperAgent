# SuperAgent Challenges System

A comprehensive challenge framework for testing, verifying, and validating LLM providers, AI debate groups, and API quality.

## Overview

The SuperAgent Challenges System provides:

- **Automated Provider Verification**: Test and score all configured LLM providers
- **AI Debate Group Formation**: Create optimized groups of top-performing models
- **API Quality Testing**: Validate response quality with comprehensive assertions
- **Comprehensive Logging**: Full audit trail of all API communications
- **Execution Reports**: Detailed reports and master summaries with history tracking

## Quick Start

```bash
# 1. Copy environment template and configure API keys
cp config/.env.example config/.env
# Edit .env with your actual API keys

# 2. Verify configuration
./scripts/verify_config.sh

# 3. Run all challenges
./scripts/run_all_challenges.sh

# 4. View results
cat results/latest_summary.md
```

## Directory Structure

```
challenges/
├── README.md                   # This file
├── docs/                       # Documentation
│   ├── 00_INDEX.md             # Documentation index
│   ├── 01_INTRODUCTION.md      # Framework introduction
│   └── ...                     # Additional docs
├── config/                     # Configuration files
│   ├── .env.example            # Environment template (git-versioned)
│   └── providers.yaml.example  # Provider config template
├── data/                       # Challenge definitions
│   ├── challenges_bank.json    # Challenge registry
│   └── test_prompts.json       # Test prompts by category
├── scripts/                    # Shell scripts
│   ├── common.sh               # Shared functions
│   ├── run_challenges.sh       # Run specific challenges
│   ├── run_all_challenges.sh   # Run all challenges
│   ├── verify_config.sh        # Validate configuration
│   ├── generate_report.sh      # Generate reports
│   └── cleanup_results.sh      # Clean old results
├── codebase/go_files/          # Go implementation
│   ├── framework/              # Core framework (61 tests)
│   │   ├── types.go            # Core types
│   │   ├── interfaces.go       # Interface definitions
│   │   ├── registry.go         # Challenge registry
│   │   ├── runner.go           # Challenge execution
│   │   ├── assertions.go       # Assertion engine (15+ evaluators)
│   │   ├── env.go              # Environment handling
│   │   ├── logger.go           # Logging system
│   │   ├── reporter.go         # Report generation
│   │   └── integration_test.go # Integration tests
│   ├── provider_verification/  # Provider verification challenge
│   ├── ai_debate_formation/    # Debate group formation challenge
│   └── api_quality_test/       # API quality testing challenge
└── results/                    # Execution results (git-ignored)
    └── <timestamp>/            # Per-run results
```

## Available Challenges

| Challenge | Description | Category |
|-----------|-------------|----------|
| `provider_verification` | Verify all LLM providers and score models | Core |
| `ai_debate_formation` | Form AI debate group from top models | Core |
| `api_quality_test` | Test API quality with assertions | Validation |

## Framework Components

### Core Framework (`codebase/go_files/framework/`)

| File | Purpose | Tests |
|------|---------|-------|
| `types.go` | Core types for challenges, results, assertions | - |
| `interfaces.go` | Challenge, Registry, Runner, Logger interfaces | - |
| `registry.go` | Challenge registration with dependency ordering | 7 |
| `runner.go` | Challenge execution lifecycle | - |
| `assertions.go` | 15+ assertion evaluators | 14 |
| `env.go` | Secure environment handling with redaction | 12 |
| `logger.go` | JSON/Console/Multi/Redacting loggers | 13 |
| `reporter.go` | Markdown/JSON reports with history | 14 |
| `integration_test.go` | Full workflow integration tests | 6 |

### Assertion Types

- `not_empty` - Value is not empty
- `not_mock` - Response is not mocked/placeholder
- `contains` / `contains_any` - Contains expected substrings
- `min_length` / `max_length` - Length constraints
- `quality_score` - Meets quality threshold
- `reasoning_present` - Contains reasoning indicators
- `code_valid` - Contains valid code structure
- `response_time` / `max_latency` - Time constraints
- `format_valid` - Valid JSON format
- `no_errors` - No error indicators
- `min_count` / `no_duplicates` - Array constraints

### Logging System

| Logger | Purpose |
|--------|---------|
| `JSONLogger` | JSON Lines format to files |
| `ConsoleLogger` | Colored terminal output |
| `MultiLogger` | Write to multiple destinations |
| `RedactingLogger` | Automatic secret redaction |
| `NullLogger` | Discard all output |

## Security

**IMPORTANT**: Never commit API keys or sensitive information!

### Protected Files (git-ignored)
- `config/.env` - Contains actual API keys
- `config/providers.yaml` - Actual configuration
- `results/` - All execution results and logs

### Git-versioned Templates
- `config/.env.example` - Environment template with placeholders
- `config/providers.yaml.example` - Configuration template

### Automatic Redaction
- API keys: `sk-ant-api-123...` → `sk-a*********`
- Authorization headers: `Bearer sk-...` → `Bearer ***`
- URL parameters: `?api_key=secret` → `?api_key=***`

## Configuration

### Environment Variables (.env)

```bash
# Provider API Keys
ANTHROPIC_API_KEY=sk-ant-api-...
OPENAI_API_KEY=sk-...
OPENROUTER_API_KEY=sk-or-v1-...
DEEPSEEK_API_KEY=sk-...
GEMINI_API_KEY=...

# SuperAgent Configuration
SUPERAGENT_BASE_URL=http://localhost:8080
SUPERAGENT_API_KEY=...
```

### Provider Configuration (providers.yaml)

```yaml
providers:
  anthropic:
    enabled: true
    model: claude-3-opus-20240229
    weight: 1.0
  openai:
    enabled: true
    model: gpt-4-turbo
    weight: 1.0

scoring:
  response_time_weight: 0.3
  quality_weight: 0.5
  reliability_weight: 0.2

debate_group:
  primary_count: 5
  fallbacks_per_primary: 2
  diversity_preference: true
```

## Running Tests

```bash
# Run all framework tests
cd challenges/codebase/go_files/framework
go test -v ./...

# Run integration tests only
go test -v -run Integration ./...

# Run with coverage
go test -cover ./...
```

## Documentation

- [00_INDEX.md](docs/00_INDEX.md) - Documentation index
- [01_INTRODUCTION.md](docs/01_INTRODUCTION.md) - Framework introduction
- [02_QUICK_START.md](docs/02_QUICK_START.md) - Quick start guide
- [03_CHALLENGE_CATALOG.md](docs/03_CHALLENGE_CATALOG.md) - All challenges
- [04_AI_DEBATE_GROUP.md](docs/04_AI_DEBATE_GROUP.md) - Debate group details
- [05_SECURITY.md](docs/05_SECURITY.md) - Security practices

## License

Part of the SuperAgent project.
