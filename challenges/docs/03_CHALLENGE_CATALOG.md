# HelixAgent Challenges - Challenge Catalog

Complete specification of all available challenges.

---

## Challenge Overview

| ID | Name | Category | Dependencies | Duration |
|----|------|----------|--------------|----------|
| `provider_verification` | Provider Verification | Core | None | 2-5 min |
| `ai_debate_formation` | AI Debate Group Formation | Core | provider_verification | 3-8 min |
| `api_quality_test` | API Quality Testing | Validation | ai_debate_formation | 5-15 min |

---

## Core Challenges

### 1. Provider Verification (`provider_verification`)

**Purpose**: Verify all configured LLM providers and score their models.

**Description**:
This challenge discovers, verifies, and scores all LLM providers configured in the environment. It integrates with LLMsVerifier to perform comprehensive capability testing.

**Inputs**:
- API keys from `.env`
- Provider configurations

**Outputs**:
- `providers_verified.json` - List of verified providers
- `models_scored.json` - Scored model list
- `verification_report.md` - Human-readable report

**Assertions**:
- At least 1 provider verified successfully
- All API keys valid
- Response times within acceptable limits

**Metrics Collected**:
- Provider count (verified/failed)
- Model count per provider
- Average response time
- Capability coverage

---

### 2. AI Debate Group Formation (`ai_debate_formation`)

**Purpose**: Form an optimal AI debate group from top-scoring models.

**Description**:
Creates an AI debate group consisting of 5 primary members with 2 fallback models each, totaling 15 LLMs. Selection is based on verification scores from the Provider Verification challenge.

**Inputs**:
- Scored models from `provider_verification`
- Group configuration parameters

**Outputs**:
- `debate_group.json` - Complete group configuration
- `member_assignments.json` - Member/fallback assignments
- `formation_report.md` - Formation details

**Group Structure**:
```
Debate Group (15 models total)
├── Primary Member 1 (Highest Score)
│   ├── Fallback 1a
│   └── Fallback 1b
├── Primary Member 2
│   ├── Fallback 2a
│   └── Fallback 2b
├── Primary Member 3
│   ├── Fallback 3a
│   └── Fallback 3b
├── Primary Member 4
│   ├── Fallback 4a
│   └── Fallback 4b
└── Primary Member 5
    ├── Fallback 5a
    └── Fallback 5b
```

**Selection Criteria**:
1. Verification score (weight: 0.4)
2. Capability coverage (weight: 0.3)
3. Response speed (weight: 0.2)
4. Provider diversity (weight: 0.1)

**Assertions**:
- Exactly 5 primary members
- Exactly 2 fallbacks per primary
- No duplicate models
- Minimum average score threshold

---

## Validation Challenges

### 3. API Quality Testing (`api_quality_test`)

**Purpose**: Validate response quality of the AI debate group.

**Description**:
Sends comprehensive test requests to the formed debate group via the OpenAI-compatible API and validates responses using assertions.

**Test Categories**:

#### Code Generation Tests
| Test | Prompt | Assertions |
|------|--------|------------|
| Simple Function | "Write a Go function to calculate factorial" | contains:func, not_empty, min_length:50 |
| Algorithm | "Implement binary search in Python" | contains:def, contains:return, code_valid |
| Class Design | "Create a TypeScript class for user management" | contains:class, contains:constructor |

#### Code Review Tests
| Test | Prompt | Assertions |
|------|--------|------------|
| Bug Detection | "Review: `func divide(a,b int) int { return a/b }`" | contains_any:[zero,division], quality_score:0.7 |
| Security Review | "Find security issues in: [SQL query code]" | contains_any:[injection,sanitize] |

#### Reasoning Tests
| Test | Prompt | Assertions |
|------|--------|------------|
| Math Problem | "A farmer has 17 sheep. All but 9 run away. How many left?" | contains:9, reasoning_present |
| Logic Puzzle | "If all A are B, and all B are C, are all A also C?" | contains:yes, reasoning_present |

#### Quality Tests
| Test | Prompt | Assertions |
|------|--------|------------|
| Completeness | "Explain REST API best practices" | min_length:200, not_mock |
| Accuracy | "What is the capital of France?" | contains:Paris, not_empty |

**Assertion Types**:

| Type | Description | Example |
|------|-------------|---------|
| `not_empty` | Response must not be empty | `{"type": "not_empty"}` |
| `not_mock` | Response must not be mocked | `{"type": "not_mock"}` |
| `contains` | Must contain text | `{"type": "contains", "value": "func"}` |
| `contains_any` | Must contain any of values | `{"type": "contains_any", "values": ["yes", "true"]}` |
| `min_length` | Minimum character count | `{"type": "min_length", "value": 100}` |
| `quality_score` | Quality >= threshold | `{"type": "quality_score", "min_value": 0.7}` |
| `reasoning_present` | Shows reasoning steps | `{"type": "reasoning_present"}` |
| `code_valid` | Contains valid code | `{"type": "code_valid"}` |

**Outputs**:
- `test_results.json` - Detailed test results
- `assertion_report.json` - Assertion outcomes
- `api_quality_report.md` - Human-readable report
- `api_requests.log` - All API requests
- `api_responses.log` - All API responses

**Pass Criteria**:
- All tests must pass
- No mock responses detected
- Average quality score >= 0.8
- Response time < 30s per request

---

## Challenge Execution

### Running Challenges

```bash
# Single challenge
./scripts/run_challenges.sh <challenge_id>

# With options
./scripts/run_challenges.sh ai_debate_formation --verbose --timeout=600

# All challenges in sequence
./scripts/run_all_challenges.sh
```

### Challenge Dependencies

```
provider_verification (no dependencies)
        │
        ▼
ai_debate_formation (requires: provider_verification)
        │
        ▼
api_quality_test (requires: ai_debate_formation)
```

### Result Storage

Results are stored in timestamped directories:
```
results/<challenge_id>/<YYYY>/<MM>/<DD>/<timestamp>/
├── logs/
│   ├── challenge.log
│   ├── commands.log
│   ├── api_requests.log
│   └── api_responses.log
├── results/
│   ├── *.json
│   └── *.md
└── config/
    └── config.yaml.redacted
```

---

## Adding New Challenges

### 1. Define Challenge

Add to `data/challenges_bank.json`:
```json
{
  "id": "my_challenge",
  "name": "My Challenge",
  "description": "Description here",
  "category": "custom",
  "dependencies": ["provider_verification"],
  "estimated_duration": "5-10 minutes",
  "outputs": ["results.json", "report.md"]
}
```

### 2. Implement Runner

Create `codebase/challenge_runners/my_challenge/run.sh`:
```bash
#!/bin/bash
# Challenge runner script
source ../common.sh
setup_challenge "my_challenge"
# Implementation here
finalize_challenge
```

### 3. Implement Logic

Create `codebase/go_files/my_challenge/main.go`:
```go
package main

import "dev.helix.agent/challenges"

func main() {
    runner := challenges.NewRunner("my_challenge")
    runner.Execute()
}
```

### 4. Add Tests

Create `my_challenge_test.go` with unit tests.
