# HelixAgent Challenges System - Implementation Plan

## Overview

This document outlines the comprehensive implementation plan for the HelixAgent Challenges System, modeled after the LLMsVerifier challenges framework. The system will enable automated testing, verification, and scoring of LLM providers, formation of AI debate groups, and comprehensive quality assurance.

## Primary Challenge: AI Debate Group Formation & Verification

### Objective
Create and validate an AI debate group consisting of:
- **5 Primary Members**: Top-scoring verified LLMs
- **2 Fallback LLMs per Member**: Backup models for reliability
- **Total**: 15 best-scored LLMs working together

### Flow
1. Read API keys from git-versioned `.env` file
2. Use LLMsVerifier to test, verify, and score all providers/models
3. Form AI debate group from top 15 scored models
4. Expose as virtual LLM via OpenAI-compatible API
5. Run comprehensive test requests against the debate group
6. Assert response quality (no mocks, empty, or fake responses)
7. Log all API communications
8. Generate execution reports and documentation

---

## Phase 1: Directory Structure & Documentation

### Directory Layout
```
challenges/
├── data/
│   └── challenges_bank.json          # Challenge registry
├── docs/
│   ├── 00_INDEX.md                   # Documentation index
│   ├── 01_INTRODUCTION.md            # Framework overview
│   ├── 02_QUICK_START.md             # Quick start guide
│   ├── 03_CHALLENGE_CATALOG.md       # All challenges spec
│   ├── 04_AI_DEBATE_GROUP.md         # Debate group challenge
│   └── 05_SECURITY.md                # Security practices
├── codebase/
│   ├── challenge_runners/            # Shell script runners
│   │   ├── ai_debate_formation/
│   │   │   └── run.sh
│   │   ├── provider_verification/
│   │   │   └── run.sh
│   │   └── api_quality_test/
│   │       └── run.sh
│   └── go_files/                     # Go implementations
│       ├── ai_debate_formation/
│       │   └── main.go
│       ├── provider_verification/
│       │   └── main.go
│       └── api_quality_test/
│           └── main.go
├── results/                          # Timestamped results
│   └── [challenge_name]/[YYYY]/[MM]/[DD]/[timestamp]/
│       ├── logs/
│       │   ├── challenge.log
│       │   ├── commands.log
│       │   ├── api_requests.log
│       │   └── api_responses.log
│       ├── results/
│       │   ├── results_opencode.json
│       │   ├── results_crush.json
│       │   └── debate_group.json
│       └── config/
│           └── config.yaml.redacted
├── master_results/                   # Master summaries
│   └── master_summary_*.md
├── scripts/
│   ├── run_challenges.sh             # Main runner
│   ├── run_all_challenges.sh         # Run all
│   └── generate_report.sh            # Report generator
├── config/
│   ├── config.yaml.example           # Example config (redacted)
│   └── .gitignore                    # Ignore actual config
├── challenge_framework.go            # Core framework
├── challenge_types.go                # Type definitions
├── challenge_runner.go               # Execution engine
├── challenge_logger.go               # Logging system
├── challenge_reporter.go             # Report generation
├── challenges_test.go                # Tests
├── README.md                         # Overview
└── .gitignore                        # Security gitignore
```

### Deliverables
- [ ] Create directory structure
- [ ] Create README.md with overview
- [ ] Create documentation index (00_INDEX.md)
- [ ] Create introduction document (01_INTRODUCTION.md)
- [ ] Create .gitignore for secrets

---

## Phase 2: Challenge Framework Core

### Go Interface Definitions

```go
// Challenge represents a single challenge
type Challenge struct {
    ID                string            `json:"id"`
    Name              string            `json:"name"`
    Description       string            `json:"description"`
    Category          string            `json:"category"`
    Dependencies      []string          `json:"dependencies"`
    EstimatedDuration string            `json:"estimated_duration"`
    Outputs           []string          `json:"outputs"`
    Config            ChallengeConfig   `json:"config"`
}

// ChallengeResult represents execution result
type ChallengeResult struct {
    ChallengeID     string                 `json:"challenge_id"`
    Status          string                 `json:"status"` // pending, running, passed, failed
    StartTime       time.Time              `json:"start_time"`
    EndTime         time.Time              `json:"end_time"`
    Duration        time.Duration          `json:"duration"`
    Logs            []LogEntry             `json:"logs"`
    Artifacts       map[string]string      `json:"artifacts"`
    Metrics         ChallengeMetrics       `json:"metrics"`
    Error           string                 `json:"error,omitempty"`
}

// ChallengeRunner executes challenges
type ChallengeRunner interface {
    Run(ctx context.Context, challenge *Challenge) (*ChallengeResult, error)
    Validate(challenge *Challenge) error
}
```

### Deliverables
- [ ] Implement challenge_types.go
- [ ] Implement challenge_framework.go
- [ ] Implement challenge_runner.go
- [ ] Create challenges_bank.json registry
- [ ] Add unit tests

---

## Phase 3: Secure .env Handling

### Security Requirements
1. **Git-Versioned .env.example**: Template with placeholder values
2. **Git-Ignored .env**: Actual API keys (never committed)
3. **Redacted Configs**: All generated configs have redacted keys
4. **Runtime Resolution**: Keys loaded at runtime from environment

### File Structure
```
challenges/
├── .env.example              # Git versioned - placeholder keys
├── .env                      # Git ignored - actual keys
├── config/
│   ├── config.yaml.example   # Git versioned - redacted
│   └── config.yaml           # Git ignored - actual
```

### .env.example Format
```env
# HelixAgent Challenges - API Keys Configuration
# Copy this file to .env and fill in actual values
# NEVER commit .env with real keys!

# LLM Provider API Keys
ANTHROPIC_API_KEY=sk-ant-xxxxx
OPENAI_API_KEY=sk-xxxxx
DEEPSEEK_API_KEY=sk-xxxxx
GEMINI_API_KEY=AIzaSyxxxxx
OPENROUTER_API_KEY=sk-or-v1-xxxxx
QWEN_API_KEY=sk-xxxxx
ZAI_API_KEY=xxxxx
OLLAMA_BASE_URL=http://localhost:11434

# Additional Providers (from LLMsVerifier integration)
HUGGINGFACE_API_KEY=hf_xxxxx
NVIDIA_API_KEY=nvapi-xxxxx
CHUTES_API_KEY=cpk_xxxxx
SILICONFLOW_API_KEY=sk-xxxxx
KIMI_API_KEY=sk-kimi-xxxxx
```

### Redaction Logic
```go
func RedactAPIKey(key string) string {
    if len(key) <= 8 {
        return "*****"
    }
    prefix := key[:4]
    return prefix + strings.Repeat("*", len(key)-4)
}
```

### Deliverables
- [ ] Create .env.example with all provider keys
- [ ] Create .gitignore entries
- [ ] Implement env_loader.go
- [ ] Implement redaction utilities
- [ ] Add validation for required keys

---

## Phase 4: LLMsVerifier Integration

### Integration Points
1. **Model Discovery**: Use LLMsVerifier's provider discovery
2. **Verification**: Run "Can you see my code?" tests
3. **Scoring**: Use LLMsVerifier's scoring engine
4. **Results**: Import verification scores

### Integration Service
```go
type LLMsVerifierIntegration struct {
    verifierPath    string
    configPath      string
    resultsPath     string
}

func (l *LLMsVerifierIntegration) DiscoverProviders(ctx context.Context) ([]Provider, error)
func (l *LLMsVerifierIntegration) VerifyModel(ctx context.Context, model Model) (*VerificationResult, error)
func (l *LLMsVerifierIntegration) ScoreModel(ctx context.Context, model Model) (*Score, error)
func (l *LLMsVerifierIntegration) GetTopScoredModels(ctx context.Context, count int) ([]ScoredModel, error)
```

### Deliverables
- [ ] Create llmsverifier_integration.go
- [ ] Implement provider discovery wrapper
- [ ] Implement verification wrapper
- [ ] Implement scoring integration
- [ ] Add result parsing

---

## Phase 5: AI Debate Group Formation Challenge

### Challenge Specification

**ID**: `ai_debate_group_formation`
**Name**: AI Debate Group Formation Challenge
**Category**: core

### Formation Algorithm
1. Run verification on all configured providers
2. Score all verified models
3. Sort by score (descending)
4. Select top 5 as primary debate members
5. For each primary, select next 2 highest-scored compatible models as fallbacks
6. Validate group configuration
7. Store debate group configuration

### Debate Group Structure
```go
type DebateGroup struct {
    ID              string           `json:"id"`
    Name            string           `json:"name"`
    CreatedAt       time.Time        `json:"created_at"`
    Members         []DebateMember   `json:"members"`
    TotalModels     int              `json:"total_models"`
    AverageScore    float64          `json:"average_score"`
    Configuration   DebateConfig     `json:"configuration"`
}

type DebateMember struct {
    Role            string           `json:"role"` // primary, fallback_1, fallback_2
    Position        int              `json:"position"` // 1-5 for primaries
    Model           ScoredModel      `json:"model"`
    Fallbacks       []ScoredModel    `json:"fallbacks,omitempty"`
}

type ScoredModel struct {
    ProviderName    string           `json:"provider_name"`
    ModelID         string           `json:"model_id"`
    ModelName       string           `json:"model_name"`
    Score           float64          `json:"score"`
    VerificationID  string           `json:"verification_id"`
    Capabilities    []string         `json:"capabilities"`
}
```

### Deliverables
- [ ] Implement debate_group_formation.go
- [ ] Create formation algorithm
- [ ] Implement group validation
- [ ] Create group persistence
- [ ] Add formation tests

---

## Phase 6: OpenAI API Test Suite

### Test Categories

#### 1. Code Generation Tests
```go
var codeGenerationTests = []APITest{
    {
        Name: "Simple Function Generation",
        Prompt: "Write a Go function that calculates the factorial of a number",
        Assertions: []Assertion{
            {Type: "contains", Value: "func"},
            {Type: "contains", Value: "factorial"},
            {Type: "not_empty"},
            {Type: "min_length", Value: 50},
        },
    },
    {
        Name: "Algorithm Implementation",
        Prompt: "Implement a binary search algorithm in Python with proper error handling",
        Assertions: []Assertion{
            {Type: "contains", Value: "def"},
            {Type: "contains", Value: "binary"},
            {Type: "contains", Value: "return"},
            {Type: "not_mock"},
        },
    },
}
```

#### 2. Code Review Tests
```go
var codeReviewTests = []APITest{
    {
        Name: "Bug Detection",
        Prompt: `Review this code and identify bugs:
func divide(a, b int) int {
    return a / b
}`,
        Assertions: []Assertion{
            {Type: "contains_any", Values: []string{"zero", "divide by zero", "division"}},
            {Type: "quality_score", MinValue: 0.7},
        },
    },
}
```

#### 3. Reasoning Tests
```go
var reasoningTests = []APITest{
    {
        Name: "Multi-step Problem",
        Prompt: "A farmer has 17 sheep. All but 9 run away. How many are left?",
        Assertions: []Assertion{
            {Type: "contains", Value: "9"},
            {Type: "reasoning_present"},
        },
    },
}
```

#### 4. Quality Assertions
```go
type Assertion struct {
    Type     string      `json:"type"`
    Value    interface{} `json:"value,omitempty"`
    MinValue float64     `json:"min_value,omitempty"`
    Values   []string    `json:"values,omitempty"`
}

// Assertion Types:
// - not_empty: Response must not be empty
// - not_mock: Response must not contain mock indicators
// - contains: Response must contain specific text
// - contains_any: Response must contain at least one of the values
// - min_length: Response must have minimum character count
// - quality_score: Response quality score >= min_value
// - reasoning_present: Response shows reasoning steps
// - code_valid: Response contains syntactically valid code
```

### Deliverables
- [ ] Implement api_test_suite.go
- [ ] Create test definitions
- [ ] Implement assertion engine
- [ ] Create quality scoring
- [ ] Add comprehensive test cases (20+)

---

## Phase 7: Comprehensive Logging System

### Log Types

#### 1. Challenge Log
```
[2025-01-04 10:30:00] ========================================
[2025-01-04 10:30:00] AI DEBATE GROUP FORMATION CHALLENGE
[2025-01-04 10:30:00] ========================================
[2025-01-04 10:30:00] Challenge ID: ai_debate_group_formation
[2025-01-04 10:30:01] Loading API keys from environment...
[2025-01-04 10:30:01] Found 15 configured providers
[2025-01-04 10:30:02] Starting LLMsVerifier integration...
```

#### 2. API Request Log
```json
{
  "timestamp": "2025-01-04T10:30:05Z",
  "direction": "outbound",
  "provider": "openrouter",
  "model": "anthropic/claude-3-opus",
  "endpoint": "/chat/completions",
  "method": "POST",
  "request_id": "req_abc123",
  "prompt_length": 150,
  "prompt_preview": "Write a Go function...",
  "headers_redacted": true
}
```

#### 3. API Response Log
```json
{
  "timestamp": "2025-01-04T10:30:07Z",
  "direction": "inbound",
  "request_id": "req_abc123",
  "provider": "openrouter",
  "model": "anthropic/claude-3-opus",
  "status_code": 200,
  "response_time_ms": 2150,
  "tokens_used": 450,
  "response_length": 1200,
  "response_preview": "Here is a Go function..."
}
```

### Log Storage
```
results/ai_debate_group_formation/2025/01/04/1704362400/
├── logs/
│   ├── challenge.log           # Main execution log
│   ├── commands.log            # Commands executed
│   ├── api_requests.log        # All API requests (JSON lines)
│   ├── api_responses.log       # All API responses (JSON lines)
│   ├── verification.log        # Model verification details
│   └── errors.log              # Error details
```

### Deliverables
- [ ] Implement challenge_logger.go
- [ ] Create structured logging format
- [ ] Implement API request/response logging
- [ ] Add log rotation support
- [ ] Create log aggregation utilities

---

## Phase 8: Execution Reports & Master Summaries

### Report Types

#### 1. Challenge Execution Report
```markdown
# Challenge Execution Report: AI Debate Group Formation

**Challenge ID**: ai_debate_group_formation
**Execution Date**: 2025-01-04 10:30:00 UTC
**Duration**: 5m 32s
**Status**: PASSED

## Summary
| Metric | Value |
|--------|-------|
| Providers Tested | 15 |
| Models Verified | 47 |
| Models Scored | 47 |
| Debate Group Size | 15 |
| Average Score | 8.7/10 |

## Debate Group Composition

### Primary Members
| Position | Provider | Model | Score |
|----------|----------|-------|-------|
| 1 | Anthropic | claude-3-opus | 9.5 |
| 2 | OpenAI | gpt-4-turbo | 9.3 |
...

## API Test Results
| Test Name | Status | Duration | Assertions |
|-----------|--------|----------|------------|
| Simple Function Generation | PASSED | 2.1s | 4/4 |
| Bug Detection | PASSED | 1.8s | 2/2 |
...
```

#### 2. Master Summary
```markdown
# HelixAgent Challenges - Master Summary

**Generated**: 2025-01-04T15:45:00Z
**Total Challenges Run**: 5

## Challenge Results

| Challenge | Status | Tests | Success Rate | Duration |
|-----------|--------|-------|--------------|----------|
| Provider Verification | PASSED | 15/15 | 100% | 2m 15s |
| AI Debate Formation | PASSED | 5/5 | 100% | 5m 32s |
| API Quality Tests | PASSED | 20/20 | 100% | 8m 45s |

## Overall Statistics
- **Total Challenges**: 5
- **Passed**: 5
- **Failed**: 0
- **Success Rate**: 100%

## Debate Group Status
- **Active**: Yes
- **Members**: 25 (5 primary + 20 fallbacks)
- **Average Score**: 8.7/10
- **Last Verified**: 2025-01-04T10:35:32Z
```

### Deliverables
- [ ] Implement challenge_reporter.go
- [ ] Create report templates
- [ ] Implement master summary generation
- [ ] Add HTML report option
- [ ] Create report archiving

---

## Phase 9: Final Integration & Testing

### Integration Tests
1. End-to-end challenge execution
2. LLMsVerifier integration validation
3. Debate group formation verification
4. API test suite execution
5. Report generation verification

### Documentation
- [ ] Complete all docs/*.md files
- [ ] Add usage examples
- [ ] Create troubleshooting guide
- [ ] Add contribution guidelines

### Git Operations
After each phase:
```bash
# Stage changes
git add challenges/

# Commit with descriptive message
git commit -m "challenges: Phase X - [description]"

# Push to upstream
git push origin main

# Handle submodules if needed
git submodule update --remote
git add LLMsVerifier Toolkit
git commit -m "submodules: Update to latest"
git push origin main
```

---

## Security Checklist

- [ ] .env file is in .gitignore
- [ ] config.yaml is in .gitignore (only .example versioned)
- [ ] All generated configs have redacted API keys
- [ ] Logs don't contain raw API keys
- [ ] No hardcoded credentials in code
- [ ] Pre-commit hook checks for secrets (optional)

---

## Timeline Estimates

| Phase | Description | Estimated Effort |
|-------|-------------|------------------|
| 1 | Directory Structure & Docs | 1 hour |
| 2 | Challenge Framework Core | 2 hours |
| 3 | Secure .env Handling | 1 hour |
| 4 | LLMsVerifier Integration | 2 hours |
| 5 | AI Debate Group Formation | 3 hours |
| 6 | OpenAI API Test Suite | 2 hours |
| 7 | Logging System | 1 hour |
| 8 | Reports & Summaries | 1 hour |
| 9 | Final Integration | 2 hours |

**Total Estimated**: ~15 hours

---

## Progress Tracking

Progress is tracked via:
1. This document's checklist items
2. Git commits after each phase
3. TodoWrite tool in Claude Code session
4. master_results/ summary files

To resume work:
1. Check last commit message for completed phase
2. Review this document's checklist
3. Check master_results/ for last execution status
4. Continue from next uncompleted phase
