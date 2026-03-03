# Planning Module Architecture

**Module:** `digital.vasic.planning`

## Overview

AI planning algorithms for complex task decomposition and solution search: Hierarchical Planning (HiPlan), Monte Carlo Tree Search (MCTS), and Tree of Thoughts (ToT).

## Core Algorithms

### HiPlan (`hiplan.go`)
Hierarchical task network planning. Decomposes high-level goals into subtasks recursively, building execution plans with dependency ordering.

**Key concepts:**
- Task decomposition (abstract → concrete)
- Precondition/effect tracking
- Plan validation and repair

### MCTS (`mcts.go`)
Monte Carlo Tree Search for exploring solution spaces. Uses selection, expansion, simulation, and backpropagation phases to find optimal action sequences.

**Key concepts:**
- UCB1 exploration/exploitation balance
- Rollout simulation
- Node statistics tracking

### Tree of Thoughts (`tree_of_thoughts.go`)
Structured reasoning by exploring multiple thought paths simultaneously. Evaluates intermediate thoughts and prunes unpromising branches.

**Key concepts:**
- Thought generation (breadth-first or depth-first)
- Thought evaluation scoring
- Branch pruning strategies

## Usage Pattern

```
Problem Definition → Algorithm Selection → Search/Planning
                                               ↓
                                          Solution Path
                                               ↓
                                          Execution Plan
```

## Package Structure

```
planning/
├── hiplan.go            # Hierarchical planning
├── mcts.go              # Monte Carlo Tree Search
└── tree_of_thoughts.go  # Tree of Thoughts reasoning
```

## Integration

Used by HelixAgent's debate orchestrator for planning debate strategies and by the agentic workflow system for task decomposition.
Adapter: `internal/adapters/planning/adapter.go`
