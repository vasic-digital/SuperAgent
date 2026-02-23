# SelfImprove Module User Guide

**Module**: `digital.vasic.selfimprove`
**Directory**: `SelfImprove/`
**Phase**: 5 (AI/ML)

## Overview

The SelfImprove module implements an AI self-improvement loop inspired by Reinforcement Learning
from Human Feedback (RLHF). It provides three collaborating components:

1. **Reward Model** (`AIRewardModel`) — Scores LLM responses on a -1.0 to 1.0 scale, broken down
   by quality dimensions (accuracy, relevance, helpfulness, harmlessness, honesty, coherence,
   creativity, formatting). Optionally uses HelixAgent's AI debate system as an ensemble evaluator
   for higher-quality scoring.
2. **Feedback Collector** — Collects structured feedback from human users, other AI models, the
   debate system, or automated metrics. Aggregates statistics and exports training examples.
3. **Policy Optimizer** (`SelfImprovementOptimizer`) — Analyses accumulated training examples and
   proposes targeted policy/prompt updates (refinements, guideline additions, example injections,
   constraint updates, tone adjustments) with configurable safety limits.

## Installation

```go
import "digital.vasic.selfimprove/selfimprove"
```

Add to your `go.mod` (HelixAgent uses a `replace` directive for local development):

```go
require digital.vasic.selfimprove v0.0.0

replace digital.vasic.selfimprove => ./SelfImprove
```

## Key Types and Interfaces

### RewardModel

Evaluates response quality.

```go
type RewardModel interface {
    Score(ctx context.Context, prompt, response string) (float64, error)
    ScoreWithDimensions(ctx context.Context, prompt, response string) (map[DimensionType]float64, error)
    Compare(ctx context.Context, prompt, response1, response2 string) (*PreferencePair, error)
    Train(ctx context.Context, examples []*TrainingExample) error
}
```

`AIRewardModel` is the concrete implementation backed by an `LLMProvider` and an optional
`DebateService` for ensemble scoring.

### DimensionType

Eight quality dimensions tracked per response:

```go
const (
    DimensionAccuracy    DimensionType = "accuracy"
    DimensionRelevance   DimensionType = "relevance"
    DimensionHelpfulness DimensionType = "helpfulness"
    DimensionHarmless    DimensionType = "harmlessness"
    DimensionHonest      DimensionType = "honesty"
    DimensionCoherence   DimensionType = "coherence"
    DimensionCreativity  DimensionType = "creativity"
    DimensionFormatting  DimensionType = "formatting"
)
```

### FeedbackCollector

Collects and aggregates feedback.

```go
type FeedbackCollector interface {
    Collect(ctx context.Context, feedback *Feedback) error
    GetBySession(ctx context.Context, sessionID string) ([]*Feedback, error)
    GetByPrompt(ctx context.Context, promptID string) ([]*Feedback, error)
    GetAggregated(ctx context.Context, filter *FeedbackFilter) (*AggregatedFeedback, error)
    Export(ctx context.Context, filter *FeedbackFilter) ([]*TrainingExample, error)
}
```

### Feedback

```go
type Feedback struct {
    ID           string
    SessionID    string
    PromptID     string
    ResponseID   string
    Type         FeedbackType    // positive | negative | neutral | suggestion | correction
    Source       FeedbackSource  // human | ai | debate | verifier | metric
    Score        float64         // -1.0 to 1.0
    Dimensions   map[DimensionType]float64
    Comment      string
    Correction   string          // corrected response if applicable
    ProviderName string
    Model        string
    CreatedAt    time.Time
}
```

### TrainingExample

Prepared for RLHF-style training.

```go
type TrainingExample struct {
    ID                string
    Prompt            string
    Response          string
    PreferredResponse string
    RejectedResponse  string
    Feedback          []*Feedback
    RewardScore       float64
    Dimensions        map[DimensionType]float64
    ProviderName      string
    Model             string
}
```

### PolicyOptimizer

Generates and applies policy updates.

```go
type PolicyOptimizer interface {
    Optimize(ctx context.Context, examples []*TrainingExample) ([]*PolicyUpdate, error)
    Apply(ctx context.Context, update *PolicyUpdate) error
    Rollback(ctx context.Context, updateID string) error
    GetHistory(ctx context.Context, limit int) ([]*PolicyUpdate, error)
    GetCurrentPolicy() string
    SetCurrentPolicy(policy string)
}
```

### SelfImprovementConfig

```go
type SelfImprovementConfig struct {
    RewardModelProvider      string
    RewardModelName          string
    MinRewardThreshold       float64        // minimum acceptable score
    AutoCollectFeedback      bool
    FeedbackBatchSize        int
    MinConfidenceForAuto     float64
    OptimizationInterval     time.Duration  // how often to run optimization
    MinExamplesForUpdate     int            // minimum examples before update
    MaxPolicyUpdatesPerDay   int            // safety limit
    ConstitutionalPrinciples []string
    EnableSelfCritique       bool
    UseDebateForReward       bool
    UseDebateForOptimize     bool
    MaxBufferSize            int
}
```

Use `DefaultSelfImprovementConfig()` for production defaults (24h interval, 3 updates/day,
constitutional AI principles pre-loaded).

## Usage Examples

### Scoring a Response

```go
package main

import (
    "context"
    "fmt"

    "digital.vasic.selfimprove/selfimprove"
    "github.com/sirupsen/logrus"
)

func main() {
    cfg := selfimprove.DefaultSelfImprovementConfig()
    cfg.UseDebateForReward = false  // use single LLM for scoring
    model := selfimprove.NewAIRewardModel(myLLMProvider, nil, cfg, logrus.New())

    score, err := model.Score(context.Background(), "What is Go?", "Go is a statically typed language...")
    if err != nil {
        panic(err)
    }
    fmt.Printf("Overall score: %.2f\n", score)

    dims, _ := model.ScoreWithDimensions(context.Background(), "What is Go?", "Go is a statically typed language...")
    for dim, val := range dims {
        fmt.Printf("  %s: %.2f\n", dim, val)
    }
}
```

### Collecting Human Feedback

```go
collector := selfimprove.NewInMemoryFeedbackCollector(logrus.New())

fb := &selfimprove.Feedback{
    SessionID:    "sess-001",
    PromptID:     "prompt-abc",
    ResponseID:   "resp-xyz",
    Type:         selfimprove.FeedbackTypePositive,
    Source:       selfimprove.FeedbackSourceHuman,
    Score:        0.85,
    Dimensions: map[selfimprove.DimensionType]float64{
        selfimprove.DimensionAccuracy:    0.9,
        selfimprove.DimensionHelpfulness: 0.8,
    },
    Comment:      "Clear and accurate explanation.",
    ProviderName: "claude",
    Model:        "claude-3-opus",
}
_ = collector.Collect(context.Background(), fb)

// Export training examples for RLHF
examples, _ := collector.Export(context.Background(), &selfimprove.FeedbackFilter{
    Sources: []selfimprove.FeedbackSource{selfimprove.FeedbackSourceHuman},
    Limit:   1000,
})
fmt.Printf("Exported %d training examples\n", len(examples))
```

### Comparing Two Responses (Preference)

```go
pair, err := model.Compare(
    context.Background(),
    "Explain recursion",
    "Recursion is when a function calls itself...",
    "A recursive function calls itself with a smaller input until a base case is reached...",
)
if err != nil {
    panic(err)
}
fmt.Printf("Preferred: response %d (margin: %.2f)\n",
    func() int {
        if pair.ChosenScore > pair.RejectedScore { return 1 }
        return 2
    }(),
    pair.Margin,
)
```

### Training and Policy Optimization

```go
// Train the reward model with collected examples
_ = model.Train(context.Background(), examples)

// Optimize policy
optimizer := selfimprove.NewSelfImprovementOptimizer(myLLMProvider, nil, cfg, logrus.New())
updates, _ := optimizer.Optimize(context.Background(), examples)
for _, u := range updates {
    fmt.Printf("Proposed update: %s (%s)\n", u.Change, u.UpdateType)
    _ = optimizer.Apply(context.Background(), u)
}
```

## Integration with HelixAgent Adapter

HelixAgent wraps the module through `internal/adapters/selfimprove/adapter.go`.

```go
import selfimproveadapter "dev.helix.agent/internal/adapters/selfimprove"

adapter := selfimproveadapter.New(logger)

// Create a reward model (LLM provider injected by HelixAgent)
model := adapter.NewRewardModel(nil)  // nil uses DefaultSelfImprovementConfig

// Train with examples
err := adapter.Train(ctx, model, examples)
```

The adapter wires the internal HelixAgent LLM provider and debate service automatically when
the reward model is constructed through the HelixAgent dependency injection layer.

## Build and Test

```bash
cd SelfImprove
go build ./...
go test ./... -count=1 -race
```
