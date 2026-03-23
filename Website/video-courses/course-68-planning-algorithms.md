# Video Course 68: AI Planning Algorithms

## Course Overview

**Duration:** 3 hours
**Level:** Advanced
**Prerequisites:** Course 01 (Fundamentals), Course 66 (Agentic Workflows)

Master the three AI planning algorithms in HelixAgent: HiPlan (Hierarchical Planning), MCTS (Monte Carlo Tree Search), and Tree of Thoughts (ToT). Learn when to use each algorithm, how to configure them, and how to invoke them through the REST API.

---

## Learning Objectives

By the end of this course, you will be able to:

1. Explain how hierarchical planning decomposes goals into milestones and steps
2. Apply Monte Carlo Tree Search for probabilistic exploration of solution spaces
3. Use Tree of Thoughts for deliberate reasoning through branching thought chains
4. Configure search depth, branching factor, and timeout for each algorithm
5. Select the appropriate algorithm for a given problem type
6. Invoke planning endpoints and interpret their results

---

## Module 1: HiPlan -- Hierarchical Planning (45 min)

### Video 1.1: What Is Hierarchical Planning? (15 min)

**Topics:**
- Top-down goal decomposition: from abstract goals to concrete actions
- Two levels: Milestones (high-level goals) and PlanSteps (low-level actions)
- Milestone dependencies: which milestones must complete before others start
- Progress tracking: per-milestone and overall plan progress
- The `MilestoneGenerator` and `StepExecutor` interfaces

**Concept:**
```
Goal: "Build a REST API service"
  |
  +-- Milestone 1: "Set up project structure"
  |     +-- Step 1.1: Create directory layout
  |     +-- Step 1.2: Initialize Go module
  |     +-- Step 1.3: Set up configuration
  |
  +-- Milestone 2: "Implement core handlers" [depends on M1]
  |     +-- Step 2.1: Create router
  |     +-- Step 2.2: Implement CRUD endpoints
  |     +-- Step 2.3: Add input validation
  |
  +-- Milestone 3: "Add tests and documentation" [depends on M2]
        +-- Step 3.1: Write unit tests
        +-- Step 3.2: Write integration tests
        +-- Step 3.3: Generate API docs
```

### Video 1.2: HiPlan Data Model (15 min)

**Topics:**
- `Milestone`: ID, Name, State (pending/in_progress/completed/failed/skipped), Priority, Dependencies, Steps, Hints
- `PlanStep`: ID, MilestoneID, Action, State, Hints, Inputs, Outputs, Duration
- `HiPlanConfig`: MaxMilestones, MaxStepsPerMilestone, EnableParallelMilestones, EnableAdaptivePlanning
- `LLMMilestoneGenerator`: LLM-backed implementation that generates milestones from natural language goals

**Key Types:**
```go
type Milestone struct {
    ID           string
    Name         string
    State        MilestoneState
    Priority     int
    Dependencies []string
    Steps        []*PlanStep
    Hints        []string
    Progress     float64
}

type PlanStep struct {
    ID          string
    MilestoneID string
    Action      string
    State       PlanStepState
    Hints       []string
    Inputs      map[string]interface{}
    Outputs     map[string]interface{}
    Duration    time.Duration
}
```

### Video 1.3: API and Configuration (15 min)

**Topics:**
- POST /v1/planning/hiplan endpoint
- Request body: `goal` (required) + optional `config`
- Response: plan ID, milestones with steps, progress, duration
- Tuning: adjust MaxMilestones and MaxStepsPerMilestone based on goal complexity
- Parallel milestones: run independent milestones concurrently

**Request:**
```json
{
  "goal": "Build a production-ready authentication system with JWT, API keys, and OAuth",
  "config": {
    "max_milestones": 8,
    "max_steps_per_milestone": 10,
    "enable_parallel_milestones": true,
    "max_parallel_milestones": 3,
    "enable_adaptive_planning": true,
    "retry_failed_steps": true,
    "max_retries": 2,
    "timeout_seconds": 300
  }
}
```

**Response:**
```json
{
  "plan_id": "plan-xyz789",
  "goal": "Build a production-ready authentication system...",
  "state": "completed",
  "progress": 1.0,
  "milestones": [
    {
      "id": "m1",
      "name": "Design authentication architecture",
      "state": "completed",
      "priority": 1,
      "progress": 1.0,
      "steps_count": 4
    },
    {
      "id": "m2",
      "name": "Implement JWT token management",
      "state": "completed",
      "priority": 2,
      "progress": 1.0,
      "steps_count": 6
    }
  ],
  "duration_ms": 12450,
  "created_at": "2026-03-23T10:00:00Z"
}
```

---

## Module 2: MCTS -- Monte Carlo Tree Search (45 min)

### Video 2.1: How MCTS Works (15 min)

**Topics:**
- The four phases: Selection, Expansion, Simulation (Rollout), Backpropagation
- UCB1 formula: balancing exploitation (high reward) vs exploration (low visits)
- UCT-DP: depth-preference variant for code generation tasks
- Why MCTS excels at problems with large, uncertain solution spaces

**The Four Phases:**
```
1. Selection:     Traverse the tree using UCB1 to find a promising leaf
2. Expansion:     Add child nodes representing possible actions
3. Simulation:    Rollout from the new node to estimate value
4. Backpropagation: Update visit counts and rewards up the tree
```

**UCB1 Formula:**
```
UCB1(node) = average_reward + C * sqrt(ln(parent_visits) / node_visits)

C = exploration constant (higher = more exploration)
```

### Video 2.2: MCTS Data Model and Configuration (15 min)

**Topics:**
- `MCTSNode`: ID, ParentID, State, Action, Visits, TotalReward, Children, Depth
- `MCTSConfig`: ExplorationConstant, DepthPreferenceAlpha, MaxDepth, MaxIterations, SimulationCount
- Strategy interfaces: `MCTSActionGenerator`, `MCTSRewardFunction`, `MCTSRolloutPolicy`
- Built-in implementations: `CodeActionGenerator`, `CodeRewardFunction`, `DefaultRolloutPolicy`
- Thread-safe reward accumulation with `sync.RWMutex`

**Key Types:**
```go
type MCTSNode struct {
    ID          string
    ParentID    string
    State       interface{}
    Action      string
    Visits      int
    TotalReward float64
    Children    []*MCTSNode
    Depth       int
}

type MCTSConfig struct {
    ExplorationConstant  float64 // C in UCB1 formula
    DepthPreferenceAlpha float64 // Alpha for UCT-DP
    MaxDepth             int
    MaxIterations        int
    SimulationCount      int
    DiscountFactor       float64
    EnableParallel       bool
    ParallelWorkers      int
    Timeout              time.Duration
}
```

### Video 2.3: API Endpoint and Usage (15 min)

**Topics:**
- POST /v1/planning/mcts endpoint
- Request: goal, actions, optional config
- Response: best path through the tree, scores, iterations completed
- Tuning ExplorationConstant: higher values for more creative exploration
- Parallel simulations: use EnableParallel for faster convergence

**Request:**
```json
{
  "goal": "Find the optimal database schema for a multi-tenant SaaS application",
  "config": {
    "exploration_constant": 1.414,
    "max_depth": 8,
    "max_iterations": 500,
    "simulation_count": 10,
    "enable_parallel": true,
    "parallel_workers": 4,
    "timeout_seconds": 60
  }
}
```

---

## Module 3: Tree of Thoughts (ToT) (45 min)

### Video 3.1: Deliberate Reasoning with ToT (15 min)

**Topics:**
- How ToT differs from chain-of-thought: branching instead of linear
- Each `Thought` node has content, reasoning, score, and confidence
- Search strategies: BFS (breadth-first), DFS (depth-first), Beam search
- Pruning: low-scoring thoughts are eliminated to focus computation
- Backtracking: recover from dead ends by returning to higher-scoring branches

**Concept:**
```
Problem: "Design a caching strategy for HelixAgent"

Thought 1.1 (score: 0.8): "Use Redis as primary with in-memory L1 cache"
  +-- Thought 2.1 (score: 0.9): "Add TTL-based invalidation"
  +-- Thought 2.2 (score: 0.6): "Use write-through for consistency"

Thought 1.2 (score: 0.5): "Use only in-memory caching"
  (pruned -- score below threshold)

Thought 1.3 (score: 0.7): "Use semantic similarity caching"
  +-- Thought 2.3 (score: 0.85): "Combine with Redis for hybrid approach"
```

### Video 3.2: ToT Data Model and Configuration (15 min)

**Topics:**
- `Thought`: ID, ParentID, Content, Reasoning, Score, Confidence, State, Depth
- `ThoughtState`: pending, active, evaluated, pruned, selected
- `TreeOfThoughtsConfig`: MaxDepth, MaxBranches, MinScore, PruneThreshold, SearchStrategy, BeamWidth
- Strategy interfaces: `ThoughtGenerator` (proposes thoughts), `ThoughtEvaluator` (scores them)
- LLM-backed implementations: `LLMThoughtGenerator`, `LLMThoughtEvaluator`

**Key Types:**
```go
type Thought struct {
    ID         string
    ParentID   string
    Content    string
    Reasoning  string
    State      ThoughtState
    Score      float64
    Confidence float64
    Depth      int
    Children   []*Thought
}

type TreeOfThoughtsConfig struct {
    MaxDepth           int
    MaxBranches        int     // Branches per node
    MinScore           float64 // Minimum score to keep
    PruneThreshold     float64 // Score below which thoughts are pruned
    SearchStrategy     string  // "bfs", "dfs", "beam"
    BeamWidth          int     // Width for beam search
    Temperature        float64 // Diversity in thought generation
    EnableBacktracking bool
    MaxIterations      int
    Timeout            time.Duration
}
```

**Default Configuration:**
```go
func DefaultTreeOfThoughtsConfig() TreeOfThoughtsConfig {
    return TreeOfThoughtsConfig{
        MaxDepth:           10,
        MaxBranches:        5,
        MinScore:           0.3,
        PruneThreshold:     0.2,
        SearchStrategy:     "beam",
        BeamWidth:          3,
        Temperature:        0.7,
        EnableBacktracking: true,
        MaxIterations:      100,
        Timeout:            5 * time.Minute,
    }
}
```

### Video 3.3: API Endpoint and Search Strategy Selection (15 min)

**Topics:**
- POST /v1/planning/tot endpoint
- Request: problem statement, search strategy, optional config
- Response: best thought path, all evaluated thoughts, final score
- When to use BFS: exploring all options at each depth before going deeper
- When to use DFS: quickly reaching a deep solution, useful for well-structured problems
- When to use Beam: best balance of quality and speed (recommended default)

**Request:**
```json
{
  "problem": "Design an error handling strategy for a distributed microservices system",
  "config": {
    "max_depth": 6,
    "max_branches": 4,
    "min_score": 0.4,
    "prune_threshold": 0.3,
    "search_strategy": "beam",
    "beam_width": 3,
    "temperature": 0.7,
    "enable_backtracking": true,
    "timeout_seconds": 120
  }
}
```

---

## Module 4: When to Use Each Algorithm (20 min)

### Video 4.1: Algorithm Selection Guide (20 min)

**Decision Matrix:**

| Criterion | HiPlan | MCTS | ToT |
|-----------|--------|------|-----|
| **Best for** | Task decomposition | Optimization under uncertainty | Deliberate reasoning |
| **Input** | High-level goal | Goal + action space | Problem statement |
| **Output** | Milestone/step plan | Best action sequence | Best reasoning chain |
| **Structure** | Tree (2 levels) | Search tree (N levels) | Thought tree (N levels) |
| **Parallelism** | Milestone-level | Simulation-level | Branch-level |
| **Determinism** | High | Low (probabilistic) | Medium |
| **Latency** | Low-Medium | Medium-High | Medium |
| **Use when** | You need a project plan | You need to explore many options | You need deep analysis |

**Guidelines:**
- **HiPlan**: "Build me X" -- clear goal, need a structured plan
- **MCTS**: "Find the best X among many options" -- optimization with uncertainty
- **ToT**: "Analyze X from multiple angles" -- complex reasoning with trade-offs

**Combined Usage:**
```
1. Use HiPlan to decompose the overall project into milestones
2. For each milestone's key decisions, use ToT to reason through alternatives
3. For optimization steps, use MCTS to explore the solution space
4. Feed the results into agentic workflows for execution
```

---

## Module 5: Hands-On Labs (25 min)

### Lab 1: Hierarchical Planning for a Feature (10 min)

**Objective:** Use HiPlan to decompose a feature into milestones and steps.

**Steps:**
1. Send a POST /v1/planning/hiplan request with goal: "Add rate limiting to the API"
2. Review the generated milestones and their dependencies
3. Increase MaxMilestones and compare the plan granularity
4. Enable parallel milestones and observe the execution order

### Lab 2: MCTS for Architecture Decisions (10 min)

**Objective:** Use MCTS to explore database architecture options.

**Steps:**
1. Send a POST /v1/planning/mcts request for a database selection problem
2. Adjust ExplorationConstant and compare the search behavior
3. Enable parallel simulations and measure the speedup
4. Analyze the best path and its reward score

### Lab 3: Tree of Thoughts for Design Analysis (5 min)

**Objective:** Use ToT to reason through a system design problem.

**Steps:**
1. Send a POST /v1/planning/tot request with a design question
2. Compare BFS, DFS, and beam search results
3. Adjust PruneThreshold and observe how it affects solution quality
4. Identify the highest-scoring thought chain

---

## Assessment

### Quiz (10 questions)

1. What are the two levels in HiPlan's hierarchy?
2. What do the four MCTS phases do?
3. What is the UCB1 formula used for in MCTS?
4. What are the three search strategies in Tree of Thoughts?
5. When should you use beam search over BFS or DFS?
6. What does the PruneThreshold configuration control in ToT?
7. What is the role of the `MilestoneGenerator` interface?
8. How does DepthPreferenceAlpha affect MCTS exploration?
9. What is the difference between `Score` and `Confidence` in a Thought?
10. When would you combine HiPlan with ToT in a workflow?

### Practical Assessment

Solve a complex system design problem using all three algorithms:

1. Use HiPlan to decompose "Migrate a monolithic application to microservices" into milestones
2. For the "service boundary" decision, use ToT with beam search to reason through options
3. For the "technology stack" selection, use MCTS to explore combinations
4. Document which algorithm you chose for each sub-problem and why

Deliverables:
1. HiPlan output with milestone dependency graph
2. ToT output showing the highest-scoring thought chain
3. MCTS output showing the best action sequence
4. Written analysis comparing the three approaches

---

## Resources

- [Planning Module Source: HiPlan](../../Planning/planning/hiplan.go)
- [Planning Module Source: MCTS](../../Planning/planning/mcts.go)
- [Planning Module Source: ToT](../../Planning/planning/tree_of_thoughts.go)
- [Planning Handler API](../../internal/handlers/planning_handler.go)
- [Course 66: Agentic Workflows](course-66-agentic-workflows.md)
- [HelixAgent Features Overview](../../docs/website/FEATURES.md)
