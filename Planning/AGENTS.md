# AGENTS.md - Planning Module

## Overview

Planning module provides AI planning algorithms including hierarchical planning,
Monte Carlo Tree Search, and Tree of Thoughts for use in AI agent workflows.

## Key Files

- `planning/hiplan.go` — Hierarchical planning (HiPlan, MilestoneGenerator, StepExecutor)
- `planning/mcts.go` — Monte Carlo Tree Search (MCTS, MCTSActionGenerator, MCTSRewardFunction)
- `planning/tree_of_thoughts.go` — Tree of Thoughts (TreeOfThoughts, ThoughtGenerator, ThoughtEvaluator)

## Exported Types Summary

### hiplan.go
- `MilestoneState`, `PlanStepState` — State enums
- `Milestone`, `PlanStep`, `HierarchicalPlan`, `PlanResult`, `MilestoneResult`, `StepResult` — Data types
- `HiPlanConfig`, `DefaultHiPlanConfig()` — Configuration
- `MilestoneGenerator`, `StepExecutor` — Interfaces
- `HiPlan`, `NewHiPlan()` — Core planner
- `LLMMilestoneGenerator`, `NewLLMMilestoneGenerator()` — LLM-backed generator

### mcts.go
- `MCTSNodeState` — State enum
- `MCTSNode`, `MCTSResult` — Tree types
- `MCTSConfig`, `DefaultMCTSConfig()` — Configuration
- `MCTSActionGenerator`, `MCTSRewardFunction`, `MCTSRolloutPolicy` — Interfaces
- `MCTS`, `NewMCTS()` — Core planner
- `CodeActionGenerator`, `CodeRewardFunction`, `DefaultRolloutPolicy` — Concrete implementations

### tree_of_thoughts.go
- `ThoughtState` — State enum
- `Thought`, `ThoughtNode`, `ToTResult` — Tree types
- `TreeOfThoughtsConfig`, `DefaultTreeOfThoughtsConfig()` — Configuration
- `ThoughtGenerator`, `ThoughtEvaluator` — Interfaces
- `TreeOfThoughts`, `NewTreeOfThoughts()` — Core planner
- `LLMThoughtGenerator`, `LLMThoughtEvaluator` — LLM-backed implementations

## Integration with HelixAgent

The adapter at `internal/adapters/planning/adapter.go` bridges the internal
`dev.helix.agent/internal/planning` package to this extracted module.
Use `planningadapter.New(logger)` to obtain an adapter instance.

## Development Standards

- All code must compile and pass `go vet ./...`
- Tests must use table-driven style with `testify`
- No mocks outside unit tests
- Run challenges before submitting: `./challenges/scripts/planning_challenge.sh`
