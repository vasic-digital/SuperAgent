# Video Course 66: Agentic Workflows Deep Dive

## Course Overview

**Duration:** 3 hours
**Level:** Intermediate to Advanced
**Prerequisites:** Course 01 (Fundamentals), Course 12 (Advanced Workflows), Course 15 (Advanced Agentic Workflows)

Deep dive into the HelixAgent graph-based agentic workflow system. Learn to design, build, and operate autonomous multi-step AI workflows with conditional branching, checkpointing, self-correction, and integration with the AI debate ensemble.

---

## Learning Objectives

By the end of this course, you will be able to:

1. Design graph-based workflows with nodes, edges, and conditional routing
2. Choose the correct node type for each workflow step (agent, tool, condition, parallel, human, subgraph)
3. Configure checkpointing and self-correction for fault-tolerant workflows
4. Use the REST API to create and monitor workflows
5. Integrate agentic workflows with the AI debate system
6. Write tests that validate workflow execution and state transitions

---

## Module 1: What Are Agentic Workflows? (25 min)

### Video 1.1: From Simple Chains to Graphs (10 min)

**Topics:**
- Linear prompt chains and their limitations
- Why sequential pipelines break on complex tasks
- Graph-based orchestration: nodes as steps, edges as transitions
- How HelixAgent models workflows as directed graphs

**Comparison:**
```
Linear Chain:           Graph Workflow:
A -> B -> C             A -> B -> C
                             \-> D -> E (parallel branch)
                        B -> F (conditional routing)
```

### Video 1.2: Core Concepts in the Agentic Module (15 min)

**Topics:**
- `Workflow` struct: ID, Name, Graph, State, Config
- `WorkflowGraph`: map of Nodes, list of Edges, EntryPoint, EndNodes
- `WorkflowState`: runtime state threaded through all nodes
- `NodeHandler`: the function signature `func(ctx, state, input) (*NodeOutput, error)`
- The execution loop: resolve next node, invoke handler, update state, repeat

**Key Types:**
```go
type Workflow struct {
    ID          string
    Name        string
    Graph       *WorkflowGraph
    State       *WorkflowState
    Config      *WorkflowConfig
}

type WorkflowGraph struct {
    Nodes      map[string]*Node
    Edges      []*Edge
    EntryPoint string
    EndNodes   []string
}

type WorkflowState struct {
    CurrentNode string
    Messages    []Message
    Variables   map[string]interface{}
    History     []NodeExecution
    Checkpoints []Checkpoint
    Status      WorkflowStatus
}
```

**Statuses:**
- `pending` -- created, not yet started
- `running` -- currently executing nodes
- `paused` -- suspended, awaiting external input
- `completed` -- all end nodes reached
- `failed` -- unrecoverable error

---

## Module 2: Node Types and Graph Design (40 min)

### Video 2.1: The Six Node Types (15 min)

**Topics:**
- `agent` -- LLM-based agent that reasons and produces output
- `tool` -- Execute a tool (code formatter, web search, database query)
- `condition` -- Evaluate a condition and route to a branch
- `parallel` -- Execute multiple downstream nodes simultaneously
- `human` -- Pause workflow and wait for human input or approval
- `subgraph` -- Embed another workflow as a single node

**Node Definition:**
```go
type Node struct {
    ID          string
    Name        string
    Type        NodeType
    Handler     NodeHandler
    Condition   ConditionFunc
    Config      map[string]interface{}
    RetryPolicy *RetryPolicy
}
```

### Video 2.2: Edges and Conditional Routing (10 min)

**Topics:**
- Directed edges: From and To node IDs
- Optional edge conditions: `ConditionFunc` evaluated at runtime
- Label annotations for observability
- `NextNode` override in `NodeOutput` for dynamic routing

**Edge Definition:**
```go
type Edge struct {
    From      string
    To        string
    Condition ConditionFunc
    Label     string
}
```

### Video 2.3: Designing Effective Workflow Graphs (15 min)

**Topics:**
- Start from the desired outcome and work backward
- Keep graphs shallow (prefer width over depth)
- Use condition nodes to avoid unnecessary LLM calls
- Identify parallelizable branches for throughput
- Use subgraphs to encapsulate reusable logic
- Always define explicit end nodes

**Example: Code Review Workflow:**
```
[analyze_code] --> [check_security] --> [decision: is_critical?]
                                             |
                           yes: [human_review] --> [apply_fixes]
                           no:  [auto_fix]     --> [apply_fixes]
                                                       |
                                               [run_tests] --> [report]
```

---

## Module 3: The REST API (35 min)

### Video 3.1: Creating a Workflow (POST /v1/agentic/workflows) (15 min)

**Topics:**
- Request body structure: name, description, nodes, edges, entry_point, end_nodes
- Optional config: max_iterations, timeout, checkpoints, self-correction
- Optional input: initial query and context variables
- Response: workflow ID, status, and created_at timestamp

**Request Example:**
```json
{
  "name": "code-review-pipeline",
  "description": "Automated code review with security analysis",
  "nodes": [
    {"id": "analyze", "name": "Code Analyzer", "type": "agent"},
    {"id": "security", "name": "Security Check", "type": "tool"},
    {"id": "decision", "name": "Severity Check", "type": "condition"},
    {"id": "auto_fix", "name": "Auto Fixer", "type": "agent"},
    {"id": "human_review", "name": "Human Review", "type": "human"},
    {"id": "report", "name": "Report Generator", "type": "agent"}
  ],
  "edges": [
    {"from": "analyze", "to": "security"},
    {"from": "security", "to": "decision"},
    {"from": "decision", "to": "auto_fix", "label": "low_severity"},
    {"from": "decision", "to": "human_review", "label": "high_severity"},
    {"from": "auto_fix", "to": "report"},
    {"from": "human_review", "to": "report"}
  ],
  "entry_point": "analyze",
  "end_nodes": ["report"],
  "config": {
    "max_iterations": 50,
    "timeout_seconds": 600,
    "enable_checkpoints": true,
    "enable_self_correction": true
  },
  "input": {
    "query": "Review the authentication module for vulnerabilities",
    "context": {"repository": "helixagent", "branch": "main"}
  }
}
```

**Response:**
```json
{
  "id": "wf-abc123",
  "name": "code-review-pipeline",
  "description": "Automated code review with security analysis",
  "status": "running",
  "created_at": "2026-03-23T10:30:00Z"
}
```

### Video 3.2: Monitoring a Workflow (GET /v1/agentic/workflows/:id) (10 min)

**Topics:**
- Polling workflow status by ID
- Response includes current state, history of executed nodes, and checkpoints
- Error field populated on failure
- Completed workflows include final output and duration

**Response:**
```json
{
  "id": "wf-abc123",
  "name": "code-review-pipeline",
  "status": "completed",
  "state": {
    "current_node": "report",
    "variables": {
      "findings_count": 3,
      "severity": "medium"
    },
    "history": [
      {"node_id": "analyze", "start_time": "...", "end_time": "..."},
      {"node_id": "security", "start_time": "...", "end_time": "..."},
      {"node_id": "decision", "start_time": "...", "end_time": "..."},
      {"node_id": "auto_fix", "start_time": "...", "end_time": "..."},
      {"node_id": "report", "start_time": "...", "end_time": "..."}
    ]
  },
  "created_at": "2026-03-23T10:30:00Z",
  "completed_at": "2026-03-23T10:35:12Z"
}
```

### Video 3.3: Error Handling and Retry (10 min)

**Topics:**
- Per-node retry policies (max retries, delay, exponential backoff)
- Self-correction: re-run failed nodes with feedback from the error
- Workflow-level timeout for safety
- Checking error details in the response

**RetryPolicy:**
```go
type RetryPolicy struct {
    MaxRetries int
    Delay      time.Duration
    Backoff    float64 // Multiplier for exponential backoff
}
```

---

## Module 4: Configuration and Checkpointing (30 min)

### Video 4.1: Workflow Configuration Options (10 min)

**Topics:**
- `MaxIterations` (default 100): guards against infinite loops
- `Timeout` (default 30 minutes): hard time limit
- `EnableCheckpoints` (default true): save state at intervals
- `CheckpointInterval` (default 5): save after every N nodes
- `EnableSelfCorrection` (default true): retry failed nodes with error context
- `MaxRetries` (default 3) and `RetryDelay` (default 1s)

**Default Config:**
```go
func DefaultWorkflowConfig() *WorkflowConfig {
    return &WorkflowConfig{
        MaxIterations:        100,
        Timeout:              30 * time.Minute,
        EnableCheckpoints:    true,
        CheckpointInterval:   5,
        EnableSelfCorrection: true,
        MaxRetries:           3,
        RetryDelay:           1 * time.Second,
    }
}
```

### Video 4.2: Checkpoints and Recovery (10 min)

**Topics:**
- What a checkpoint contains: node ID, full state snapshot, timestamp
- Automatic checkpoint creation at configured intervals
- Restoring from a checkpoint after crash or restart
- The `Checkpoint` struct and its role in WorkflowState

**Checkpoint Structure:**
```go
type Checkpoint struct {
    ID        string
    NodeID    string
    State     map[string]interface{}
    Timestamp time.Time
}
```

### Video 4.3: Self-Correction and Adaptive Execution (10 min)

**Topics:**
- When a node fails, the error is fed back as context for retry
- The self-correction loop: fail, capture error, regenerate, retry
- Limiting self-correction attempts to prevent loops
- Combining self-correction with human-in-the-loop for complex failures

---

## Module 5: Integration with the Debate System (25 min)

### Video 5.1: Using Debate as a Node Handler (10 min)

**Topics:**
- Wrapping the debate service as an agent node handler
- The debate produces a consensus response from multiple LLMs
- Higher quality at the cost of latency
- When to use single-LLM agent nodes vs debate nodes

**Pattern:**
```go
debateHandler := func(ctx context.Context, state *WorkflowState,
    input *NodeInput) (*NodeOutput, error) {

    result, err := debateService.RunDebate(ctx, &DebateRequest{
        Topic:   input.Query,
        Context: state.Variables,
    })
    if err != nil {
        return nil, err
    }
    return &NodeOutput{
        Result:   result.Consensus,
        Messages: result.Messages,
    }, nil
}
```

### Video 5.2: Workflow-Debate Architecture Patterns (15 min)

**Topics:**
- Pattern 1: Debate for critical decisions, single LLM for routine steps
- Pattern 2: Parallel debate branches with majority vote aggregation
- Pattern 3: Debate as a validation step after agent-generated output
- Balancing quality, latency, and cost across workflow steps
- Monitoring debate execution within workflow state history

---

## Module 6: Hands-On Labs (25 min)

### Lab 1: Build a Research Workflow (10 min)

**Objective:** Create a multi-step research workflow with parallel search and synthesis.

**Steps:**
1. Define a graph with nodes: query_decomposition, web_search, paper_search, synthesize, format_report
2. web_search and paper_search run in parallel from query_decomposition
3. synthesize aggregates results from both search nodes
4. format_report generates the final output
5. Submit via POST /v1/agentic/workflows
6. Monitor progress via GET /v1/agentic/workflows/:id

### Lab 2: Add Human-in-the-Loop Approval (10 min)

**Objective:** Add a human approval gate to an existing workflow.

**Steps:**
1. Insert a `human` node between the analysis and action steps
2. Configure the workflow to pause at the human node
3. Submit the workflow and verify it reaches `paused` status
4. Simulate human approval and verify the workflow resumes

### Lab 3: Test Checkpoint Recovery (5 min)

**Objective:** Verify that a workflow can be restored from a checkpoint.

**Steps:**
1. Create a workflow with checkpointing enabled (interval = 2)
2. Execute the workflow and verify checkpoints are created
3. Simulate a failure mid-execution
4. Restore from the last checkpoint and verify the workflow resumes

---

## Assessment

### Quiz (10 questions)

1. What are the six node types in the agentic workflow system?
2. How does conditional routing work with edge conditions?
3. What is the default `MaxIterations` value and why does it exist?
4. What does a checkpoint contain?
5. How does self-correction differ from simple retry?
6. When should you use a `subgraph` node type?
7. What REST endpoint creates a new workflow?
8. How do you integrate the debate system with a workflow?
9. What happens when a workflow exceeds its configured timeout?
10. What is the role of `WorkflowState.Variables`?

### Practical Assessment

Build a complete agentic workflow that:
1. Accepts a code repository URL as input
2. Analyzes the codebase structure (agent node)
3. Runs security scanning and test coverage in parallel (tool nodes)
4. Routes to human review if critical vulnerabilities are found (condition + human nodes)
5. Generates an improvement plan using the debate system (agent node with debate handler)
6. Produces a formatted report (agent node)

Deliverables:
1. Workflow graph definition as JSON
2. API calls to create and monitor the workflow
3. Test cases validating each node transition
4. Documentation of the checkpoint and recovery strategy

---

## Resources

- [Agentic Module Source](../../Agentic/agentic/workflow.go)
- [Agentic Handler API](../../internal/handlers/agentic_handler.go)
- [Course 12: Advanced Workflows](course-12-advanced-workflows.md)
- [Course 15: Advanced Agentic Workflows](course-15-bigdata-analytics.md)
- [HelixAgent Architecture Documentation](../../docs/website/FEATURES.md)
