# Planning

AI planning algorithms module: HiPlan, MCTS, Tree of Thoughts.

**Module**: `digital.vasic.planning`

## Algorithms

- **HiPlan** — Hierarchical planning with milestone decomposition and parallel/sequential execution
- **MCTS** — Monte Carlo Tree Search for code generation and exploration
- **Tree of Thoughts** — BFS/DFS/Beam search over generated thought trees

## Usage

```go
import "digital.vasic.planning/planning"

// HiPlan
config := planning.DefaultHiPlanConfig()
hp := planning.NewHiPlan(config, generator, executor, logger)
plan, err := hp.CreatePlan(ctx, "Build a REST API")

// MCTS
mctsConfig := planning.DefaultMCTSConfig()
mcts := planning.NewMCTS(mctsConfig, actionGen, rewardFunc, rollout, logger)
result, err := mcts.Search(ctx, initialState)

// Tree of Thoughts
totConfig := planning.DefaultTreeOfThoughtsConfig()
tot := planning.NewTreeOfThoughts(totConfig, thoughtGen, thoughtEval, logger)
result, err := tot.Solve(ctx, "Design a data model")
```

## Build & Test

```bash
go build ./...
go test ./... -count=1 -race
```
