# Package: selfimprove

## Overview

The `selfimprove` package implements RLAIF (Reinforcement Learning from AI Feedback) and Constitutional AI patterns for autonomous model improvement. It provides AI-powered reward models, feedback collection, and policy optimization.

## Architecture

```
selfimprove/
├── types.go            # Core types (feedback, preferences)
├── reward_model.go     # AI reward model
├── feedback.go         # Feedback collection system
├── optimizer.go        # Policy optimizer
└── selfimprove_test.go # Unit tests (45.9% coverage)
```

## Features

- **AI Reward Model**: LLM-as-judge for response scoring
- **Debate-Based Scoring**: Multi-model consensus for quality
- **Preference Learning**: Comparison-based preference pairs
- **Constitutional Principles**: Guideline-based improvement

## Key Types

### Feedback

```go
type Feedback struct {
    ID           string
    SessionID    string
    PromptID     string
    Type         FeedbackType  // positive, negative, neutral
    Source       FeedbackSource // human, ai, debate
    Score        float64
    ProviderName string
}
```

### PreferencePair

```go
type PreferencePair struct {
    ID            string
    Prompt        string
    Chosen        string   // Preferred response
    Rejected      string   // Non-preferred response
    ChosenScore   float64
    RejectedScore float64
    Margin        float64
    Source        FeedbackSource
}
```

## Usage

### AI Reward Model

```go
import "dev.helix.agent/internal/selfimprove"

// Create reward model
config := selfimprove.DefaultSelfImprovementConfig()
model := selfimprove.NewAIRewardModel(provider, debateService, config, logger)

// Score a response
score, err := model.Score(ctx, "What is 2+2?", "The answer is 4.")
// score == 0.95

// Compare responses
pair, err := model.Compare(ctx, prompt, responseA, responseB)
// pair.Chosen == responseA (if A is better)
```

### Debate-Based Scoring

```go
config.UseDebateForReward = true
model := selfimprove.NewAIRewardModel(provider, debateService, config, logger)

// Score uses multi-model debate for consensus
score, err := model.Score(ctx, prompt, response)
// Higher confidence through multi-model agreement
```

### Feedback Collection

```go
collector := selfimprove.NewInMemoryFeedbackCollector(logger, 1000)

// Collect feedback
feedback := &selfimprove.Feedback{
    SessionID: sessionID,
    Type:      selfimprove.FeedbackTypePositive,
    Source:    selfimprove.FeedbackSourceHuman,
    Score:     0.9,
}
collector.Collect(ctx, feedback)

// Get aggregated stats
stats, _ := collector.GetAggregated(ctx, nil)
fmt.Printf("Average score: %.2f\n", stats.AverageScore)
```

### Policy Optimization

```go
optimizer := selfimprove.NewLLMPolicyOptimizer(provider, debateService, config, logger)
optimizer.SetCurrentPolicy("Be helpful and concise.")

// Get training examples
examples, _ := collector.Export(ctx, nil)

// Generate policy updates
updates, err := optimizer.Optimize(ctx, examples)
for _, update := range updates {
    fmt.Printf("Suggested change: %s\n", update.Change)
}
```

## Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| UseDebateForReward | bool | true | Use multi-model debate |
| MinExamplesForUpdate | int | 10 | Min examples for optimization |
| AutoCollectFeedback | bool | true | Auto-collect from sessions |
| ConstitutionalPrinciples | []string | [...] | Guiding principles |

## Testing

```bash
go test -v ./internal/selfimprove/...
go test -cover ./internal/selfimprove/...
```

## Dependencies

### Internal
- `internal/llm` - LLM providers
- `internal/services` - Debate service

### External
- Standard library only

## See Also

- [RLAIF Paper](https://arxiv.org/abs/2212.08073)
- [Constitutional AI Paper](https://arxiv.org/abs/2212.08073)
