# Quiz 7: New Features Assessment

## Section A: Agentic Workflows (5 questions)

1. What is the minimum required field set for a workflow creation request?
   a) name, nodes, edges
   b) name, nodes, edges, entry_point, end_nodes
   c) name, nodes
   d) nodes, edges, entry_point

2. What happens when a workflow node panics during execution?
   a) The entire server crashes
   b) The panic is recovered and the workflow reports failure
   c) The node is retried automatically
   d) The workflow hangs indefinitely

3. Which endpoint retrieves workflow execution history?
   a) POST /v1/agentic/workflows
   b) GET /v1/agentic/workflows/:id
   c) GET /v1/agentic/status/:id
   d) POST /v1/agentic/workflows/:id/status

4. What is the purpose of `end_nodes` in a workflow definition?
   a) Nodes that are deleted after execution
   b) Terminal nodes where execution completes
   c) Nodes that run last regardless of graph structure
   d) Error handling nodes

5. How does the agentic system handle cycles in the edge graph?
   a) Infinite loop until timeout
   b) Detected and rejected at creation time
   c) Executed once then skipped
   d) Converted to linear sequence

## Section B: LLMOps (5 questions)

6. What is the minimum number of variants required for an A/B experiment?
   a) 1
   b) 2
   c) 3
   d) No minimum

7. Which endpoint creates a versioned prompt template?
   a) POST /v1/llmops/evaluate
   b) POST /v1/llmops/prompts
   c) POST /v1/llmops/experiments
   d) PUT /v1/llmops/prompts/:id

8. What uniquely identifies a prompt version?
   a) id
   b) name + content hash
   c) name + version
   d) version alone

9. An experiment with status "created" means:
   a) It is actively routing traffic
   b) It has been defined but not started
   c) It completed successfully
   d) It was cancelled

10. What does the `dataset` field in an evaluation request specify?
    a) The database table to write results to
    b) The reference dataset for quality scoring
    c) The list of providers to test
    d) The output format

## Section C: Planning Algorithms (5 questions)

11. Which planning algorithm is best for breaking a feature into subtasks?
    a) MCTS
    b) Tree of Thoughts
    c) HiPlan
    d) Beam Search

12. What does the `exploration_weight` parameter control in MCTS?
    a) How many nodes to explore total
    b) Balance between exploring new actions vs exploiting known good ones
    c) Maximum tree depth
    d) Number of simulations per node

13. Tree of Thoughts supports which search strategies?
    a) BFS only
    b) DFS only
    c) BFS, DFS, and Beam Search
    d) A* and Dijkstra

14. What is the `beam_width` parameter in ToT?
    a) Maximum number of child thoughts per node
    b) Number of top thoughts kept at each level
    c) Total thoughts generated
    d) Depth limit for the search

15. Which algorithm uses reward-based evaluation to select actions?
    a) HiPlan
    b) MCTS
    c) Tree of Thoughts with BFS
    d) All of the above

---
**Answer Key:** 1-b, 2-b, 3-b, 4-b, 5-b, 6-b, 7-b, 8-c, 9-b, 10-b, 11-c, 12-b, 13-c, 14-b, 15-b
