# Lab 14: Agentic Workflows

## Objective
Build and execute a multi-step agentic workflow using HelixAgent's graph-based orchestration API.

## Prerequisites
- HelixAgent running locally (`./bin/helixagent`)
- `curl` or Postman available

## Exercise 1: Create a Simple Chain Workflow

Create a 3-node linear workflow:

```bash
curl -X POST http://localhost:7061/v1/agentic/workflows \
  -H "Content-Type: application/json" \
  -d '{
    "name": "simple-chain",
    "nodes": [
      {"id": "step1", "type": "process", "config": {"action": "analyze"}},
      {"id": "step2", "type": "process", "config": {"action": "transform"}},
      {"id": "step3", "type": "process", "config": {"action": "output"}}
    ],
    "edges": [
      {"from": "step1", "to": "step2"},
      {"from": "step2", "to": "step3"}
    ],
    "entry_point": "step1",
    "end_nodes": ["step3"],
    "input": {"data": "Hello World"}
  }'
```

**Expected:** Response with `status: completed` and `nodes_executed: 3`.

## Exercise 2: Query Workflow Status

Using the workflow ID from Exercise 1:

```bash
curl http://localhost:7061/v1/agentic/workflows/<WORKFLOW_ID>
```

**Verify:** Response contains execution history for all 3 nodes.

## Exercise 3: Branching Workflow

Create a workflow with parallel branches (diamond pattern).

## Assessment Questions
1. What happens if an entry_point references a non-existent node?
2. How does the system handle cycles in the edge graph?
3. What is the purpose of checkpointing in long-running workflows?
