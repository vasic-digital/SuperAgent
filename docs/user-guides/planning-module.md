# Planning Module User Guide

**Module**: `digital.vasic.planning`
**Directory**: `Planning/`
**Phase**: 5 (AI/ML)

## Overview

The Planning module provides three complementary AI planning algorithms for decomposing and solving
complex tasks:

1. **HiPlan** — Hierarchical planning. Breaks a high-level goal into ordered milestones, then
   decomposes each milestone into executable steps. Supports sequential and parallel execution of
   steps within milestones.

2. **MCTS (Monte Carlo Tree Search)** — Tree-based search over a space of actions. Balances
   exploration and exploitation via UCB1 scoring. Designed for code generation and optimisation
   tasks where actions can be evaluated by a reward function.

3. **Tree of Thoughts (ToT)** — Multi-path reasoning over a tree of generated thoughts. Supports
   BFS, DFS, and beam search strategies with LLM-backed thought generation and evaluation. Best
   suited for problems that benefit from considering several independent reasoning chains in
   parallel.

## Installation

```go
import "digital.vasic.planning/planning"
```

Add to your `go.mod` (HelixAgent uses a `replace` directive for local development):

```go
require digital.vasic.planning v0.0.0

replace digital.vasic.planning => ./Planning
```

## Key Types and Interfaces

### HiPlan

```go
type HiPlan struct { /* ... */ }

func NewHiPlan(
    config HiPlanConfig,
    generator MilestoneGenerator,
    executor  StepExecutor,
    logger    *logrus.Logger,
) *HiPlan

func (h *HiPlan) CreatePlan(ctx context.Context, goal string) (*HierarchicalPlan, error)
func (h *HiPlan) ExecutePlan(ctx context.Context, plan *HierarchicalPlan) (*PlanResult, error)
```

**Configuration:**

```go
type HiPlanConfig struct {
    MaxMilestones    int
    MaxStepsPerMilestone int
    ParallelExecution    bool
    MaxParallelSteps     int
    Timeout              time.Duration
    RetryFailedSteps     bool
    MaxRetries           int
}

func DefaultHiPlanConfig() HiPlanConfig
```

**Key data types:**

```go
type HierarchicalPlan struct {
    ID         string
    Goal       string
    Milestones []*Milestone
    CreatedAt  time.Time
}

type Milestone struct {
    ID          string
    Name        string
    Description string
    Steps       []*PlanStep
    DependsOn   []string  // milestone IDs
    Parallel    bool
}

type PlanStep struct {
    ID          string
    Name        string
    Description string
    Action      string
    Parameters  map[string]interface{}
    DependsOn   []string  // step IDs within the milestone
}

type PlanResult struct {
    PlanID    string
    Success   bool
    Steps     []*StepResult
    Duration  time.Duration
    Error     error
}
```

**Strategy interfaces:**

```go
type MilestoneGenerator interface {
    Generate(ctx context.Context, goal string, config HiPlanConfig) ([]*Milestone, error)
}

type StepExecutor interface {
    Execute(ctx context.Context, step *PlanStep, state map[string]interface{}) (*StepResult, error)
}
```

Use `LLMMilestoneGenerator` for LLM-backed milestone generation.

### MCTS

```go
type MCTS struct { /* ... */ }

func NewMCTS(
    config       MCTSConfig,
    actionGen    MCTSActionGenerator,
    rewardFunc   MCTSRewardFunction,
    rolloutPolicy MCTSRolloutPolicy,
    logger       *logrus.Logger,
) *MCTS

func (m *MCTS) Search(ctx context.Context, initialState string) (*MCTSResult, error)
```

**Configuration:**

```go
type MCTSConfig struct {
    MaxIterations   int
    MaxDepth        int
    ExplorationConst float64  // UCB1 constant (default 1.414)
    Timeout         time.Duration
    RolloutDepth    int
}

func DefaultMCTSConfig() MCTSConfig
```

**Strategy interfaces:**

```go
type MCTSActionGenerator interface {
    Generate(ctx context.Context, state string) ([]string, error)
}

type MCTSRewardFunction interface {
    Evaluate(ctx context.Context, state string) (float64, error)
}

type MCTSRolloutPolicy interface {
    Rollout(ctx context.Context, state string, actions []string, depth int) (string, error)
}
```

Concrete implementations: `CodeActionGenerator`, `CodeRewardFunction`, `DefaultRolloutPolicy`.

### Tree of Thoughts

```go
type TreeOfThoughts struct { /* ... */ }

func NewTreeOfThoughts(
    config    TreeOfThoughtsConfig,
    generator ThoughtGenerator,
    evaluator ThoughtEvaluator,
    logger    *logrus.Logger,
) *TreeOfThoughts

func (t *TreeOfThoughts) Solve(ctx context.Context, problem string) (*ToTResult, error)
```

**Configuration:**

```go
type TreeOfThoughtsConfig struct {
    MaxDepth         int
    BranchingFactor  int     // thoughts generated per node
    BeamWidth        int     // beam search width
    SearchStrategy   string  // "bfs" | "dfs" | "beam"
    EvalThreshold    float64 // prune thoughts below this score
    Timeout          time.Duration
}

func DefaultTreeOfThoughtsConfig() TreeOfThoughtsConfig
```

**Strategy interfaces:**

```go
type ThoughtGenerator interface {
    Generate(ctx context.Context, problem, parentThought string, n int) ([]string, error)
}

type ThoughtEvaluator interface {
    Evaluate(ctx context.Context, problem, thought string) (float64, error)
}
```

Concrete implementations: `LLMThoughtGenerator`, `LLMThoughtEvaluator`.

## Usage Examples

### HiPlan — Break Down a Complex Goal

```go
package main

import (
    "context"
    "fmt"

    "digital.vasic.planning/planning"
    "github.com/sirupsen/logrus"
)

func main() {
    logger := logrus.New()
    cfg := planning.DefaultHiPlanConfig()
    cfg.ParallelExecution = true
    cfg.MaxParallelSteps = 3

    // LLM-backed milestone generator
    generator := planning.NewLLMMilestoneGenerator(func(ctx context.Context, prompt string) (string, error) {
        return myLLM.Complete(ctx, prompt)
    }, logger)

    // Step executor (call real tools/APIs)
    executor := myStepExecutor{}

    hp := planning.NewHiPlan(cfg, generator, executor, logger)

    plan, err := hp.CreatePlan(context.Background(), "Build a REST API for user management")
    if err != nil {
        panic(err)
    }
    fmt.Printf("Plan: %d milestones\n", len(plan.Milestones))

    result, err := hp.ExecutePlan(context.Background(), plan)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Success: %v, duration: %s\n", result.Success, result.Duration)
}
```

### MCTS — Code Generation Search

```go
cfg := planning.DefaultMCTSConfig()
cfg.MaxIterations = 200
cfg.MaxDepth = 5

mcts := planning.NewMCTS(
    cfg,
    planning.NewCodeActionGenerator(myLLM, logger),
    planning.NewCodeRewardFunction(myCodeRunner, logger),
    planning.NewDefaultRolloutPolicy(),
    logger,
)

result, err := mcts.Search(context.Background(), "def fibonacci(n):")
if err != nil {
    panic(err)
}
fmt.Printf("Best action: %s (reward: %.2f)\n", result.BestAction, result.BestReward)
```

### Tree of Thoughts — Multi-Path Reasoning

```go
cfg := planning.DefaultTreeOfThoughtsConfig()
cfg.BranchingFactor = 3
cfg.SearchStrategy = "beam"
cfg.BeamWidth = 2

thoughtGen := planning.NewLLMThoughtGenerator(myLLMGenerate, 0.7, logger)
thoughtEval := planning.NewLLMThoughtEvaluator(myLLMEvaluate, logger)

tot := planning.NewTreeOfThoughts(cfg, thoughtGen, thoughtEval, logger)
result, err := tot.Solve(context.Background(), "Design a caching strategy for a high-traffic API")
if err != nil {
    panic(err)
}
fmt.Println("Best reasoning path:", result.BestPath)
fmt.Printf("Confidence: %.2f\n", result.Confidence)
```

## Integration with HelixAgent Adapter

HelixAgent wraps the module through `internal/adapters/planning/adapter.go`.

```go
import planningadapter "dev.helix.agent/internal/adapters/planning"

adapter := planningadapter.New(logger)

// HiPlan
hiPlanCfg := adapter.DefaultHiPlanConfig()
generator := adapter.NewLLMMilestoneGenerator(myLLMFunc)
hp := adapter.NewHiPlan(hiPlanCfg, generator, myExecutor)

// MCTS
mctsCfg := adapter.DefaultMCTSConfig()
mcts := adapter.NewMCTS(mctsCfg, actionGen, rewardFunc, rolloutPolicy)

// Tree of Thoughts
totCfg := adapter.DefaultTreeOfThoughtsConfig()
thoughtGen := adapter.NewLLMThoughtGenerator(myGenFunc, 0.7)
thoughtEval := adapter.NewLLMThoughtEvaluator(myEvalFunc)
tot := adapter.NewTreeOfThoughts(totCfg, thoughtGen, thoughtEval)
```

The adapter injects the HelixAgent logger automatically into all created instances.

## Build and Test

```bash
cd Planning
go build ./...
go test ./... -count=1 -race
```
