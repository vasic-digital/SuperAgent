# User Manual 36: Planning Algorithms Guide

## Overview

HelixAgent provides three AI planning algorithms for complex task decomposition: HiPlan (Hierarchical Planning), MCTS (Monte Carlo Tree Search), and Tree of Thoughts (ToT). Each is exposed via HTTP API.

## Algorithms

### HiPlan — Hierarchical Planning

Decomposes a high-level goal into milestones with ordered steps.

```
POST /v1/planning/hiplan
```

```json
{
  "goal": "Migrate the authentication system from JWT to OAuth2",
  "constraints": ["no downtime", "backward compatible for 30 days"],
  "config": {
    "max_depth": 3,
    "max_milestones": 5
  }
}
```

**Response:** A plan with milestones, each containing ordered steps with descriptions and estimated effort.

**Best for:** Project planning, feature decomposition, migration strategies.

### MCTS — Monte Carlo Tree Search

Explores action sequences to find the optimal path from an initial state to a goal.

```
POST /v1/planning/mcts
```

```json
{
  "initial_state": {"codebase": "monolith", "tests": "passing"},
  "available_actions": ["extract-service", "add-api-gateway", "setup-ci", "deploy-canary"],
  "config": {
    "iterations": 1000,
    "exploration_weight": 1.41,
    "max_depth": 5
  }
}
```

**Response:** Best action sequence with expected reward score and visit statistics.

**Best for:** Decision-making under uncertainty, optimization problems, game-tree search.

### Tree of Thoughts (ToT)

Solves problems by exploring a tree of intermediate reasoning steps.

```
POST /v1/planning/tot
```

```json
{
  "problem": "Design a caching strategy for 10M daily requests with 99.9% hit rate",
  "strategy": "bfs",
  "config": {
    "max_depth": 4,
    "branching_factor": 3,
    "beam_width": 2
  }
}
```

**Strategies:**
- `bfs` — Breadth-first: explores all thoughts at each level before going deeper
- `dfs` — Depth-first: follows one chain of thought to completion before backtracking
- `beam` — Beam search: keeps top-K thoughts at each level (controlled by `beam_width`)

**Response:** Solution path through the thought tree with evaluation scores.

**Best for:** Creative problem-solving, multi-step reasoning, architecture design.

## Choosing an Algorithm

| Scenario | Algorithm | Why |
|----------|-----------|-----|
| Break down a feature into tasks | HiPlan | Natural hierarchical decomposition |
| Find optimal deployment sequence | MCTS | Handles uncertainty and exploration |
| Design system architecture | ToT | Creative exploration of alternatives |
| Migration planning | HiPlan | Clear milestones and dependencies |
| Performance optimization | MCTS | Reward-based action selection |
| Debugging complex issues | ToT (DFS) | Deep reasoning chains |
