# CLAUDE.md - Planning Module

## Overview

`digital.vasic.planning` is a Go module for AI planning algorithms: HiPlan (hierarchical planning),
Monte Carlo Tree Search (MCTS), and Tree of Thoughts (ToT).

**Module**: `digital.vasic.planning` (Go 1.24+)

## Build & Test

```bash
go build ./...
go test ./... -count=1 -race
```

## Package Structure

| Package | Purpose |
|---------|---------|
| `planning` | Core planning algorithms: HiPlan, MCTS, ToT |

## Key Types

### HiPlan (Hierarchical Planning)

- `HiPlan` — Main hierarchical planner struct
- `HiPlanConfig` / `DefaultHiPlanConfig()` — Configuration with defaults
- `MilestoneGenerator` — Interface for generating plan milestones
- `StepExecutor` — Interface for executing individual steps
- `HierarchicalPlan` / `Milestone` / `PlanStep` / `PlanResult` — Plan data types
- `LLMMilestoneGenerator` — LLM-backed milestone generator implementation

### MCTS (Monte Carlo Tree Search)

- `MCTS` — Main MCTS planner struct
- `MCTSConfig` / `DefaultMCTSConfig()` — Configuration with defaults
- `MCTSActionGenerator` / `MCTSRewardFunction` / `MCTSRolloutPolicy` — Strategy interfaces
- `MCTSNode` / `MCTSResult` — Tree node and result types
- `CodeActionGenerator` / `CodeRewardFunction` / `DefaultRolloutPolicy` — Concrete implementations

### Tree of Thoughts

- `TreeOfThoughts` — Main ToT planner struct
- `TreeOfThoughtsConfig` / `DefaultTreeOfThoughtsConfig()` — Configuration with defaults
- `ThoughtGenerator` / `ThoughtEvaluator` — Strategy interfaces
- `Thought` / `ThoughtNode` / `ToTResult` — Thought tree data types
- `LLMThoughtGenerator` / `LLMThoughtEvaluator` — LLM-backed implementations

## Mandatory Development Standards

- 100% test coverage across unit, integration, and benchmark tests
- No mocks outside unit tests — all other tests use real implementations
- Challenges must validate real-life use cases, not just return codes
- Follow Conventional Commits: `feat(planning): ...`, `fix(planning): ...`
