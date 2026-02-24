# Debate Specification Full Compliance Design

**Date:** 2026-02-24
**Status:** Approved
**Scope:** Full spec compliance — all 4 research documents implemented to the smallest detail

## Context

Four research and specification documents in `docs/requests/debate/` define a comprehensive multi-agent AI debate system:

1. **001 AI Debate Research Kimi.md** — Core architecture, AgentPG, Go implementation, tool integration
2. **002 2025.acl-long.421.pdf** — MultiAgentBench/MARBLE framework, coordination protocols, cognitive planning
3. **003 AI Debate Research MiniMax.md** — Theoretical foundations, Perfection Workflow, Reflexion, consensus mechanisms
4. **004 AI Debate.md** — Ultra-advanced architecture, DebateCoder, step-by-step protocol, evaluation

The current implementation covers ~60-70% of the specifications. This design closes all gaps.

## Design Decisions

- **Database**: Extend existing `debate_logs` with 3 new tables (`debate_sessions`, `debate_turns`, `code_versions`). Skip `agents`/`projects`/`tasks` tables — those concepts exist elsewhere in the system.
- **Approval Gates**: Configurable, disabled by default. When enabled, REST API + SSE/WebSocket. When disabled, auto-approve.
- **Approach**: Infrastructure-Then-Features — shared foundations first, then features grouped by dependency, then test completeness, then documentation.

---

## Section 1: Database Schema Extensions

Three new tables extending the existing `debate_logs`:

### `debate_sessions`

Tracks debate lifecycle with full metadata.

| Column | Type | Description |
|--------|------|-------------|
| `id` | UUID PK | Session identifier |
| `debate_id` | VARCHAR | Links to debate_logs |
| `topic` | TEXT | Debate topic/task description |
| `status` | VARCHAR | pending/running/paused/completed/failed/cancelled |
| `topology_type` | VARCHAR | graph_mesh/star/chain/tree |
| `coordination_protocol` | VARCHAR | CPDE/DPDE/adaptive |
| `config` | JSONB | Max rounds, timeout, consensus threshold, gates enabled |
| `initiated_by` | VARCHAR | Requester identifier |
| `created_at` | TIMESTAMPTZ | Session creation time |
| `updated_at` | TIMESTAMPTZ | Last state change |
| `completed_at` | TIMESTAMPTZ | Completion time |
| `total_rounds` | INTEGER | Rounds completed |
| `final_consensus_score` | FLOAT | Final consensus level (0-1) |
| `outcome` | JSONB | Winner, voting method, confidence |

### `debate_turns`

Granular turn-level state for replay/recovery.

| Column | Type | Description |
|--------|------|-------------|
| `id` | UUID PK | Turn identifier |
| `session_id` | UUID FK | References debate_sessions |
| `round` | INTEGER | Round number |
| `phase` | VARCHAR | proposal/critique/review/optimization/convergence/self_evolvement/dehallucination/adversarial |
| `agent_id` | VARCHAR | Agent identifier |
| `agent_role` | VARCHAR | Agent role in this turn |
| `provider` | VARCHAR | LLM provider used |
| `model` | VARCHAR | Model used |
| `content` | TEXT | Response content |
| `confidence` | FLOAT | Agent confidence (0-1) |
| `tool_calls` | JSONB | Tool invocations and results |
| `test_results` | JSONB | Test execution results |
| `reflections` | JSONB | Reflexion episodic memory entries |
| `metadata` | JSONB | Additional structured data |
| `created_at` | TIMESTAMPTZ | Turn timestamp |
| `response_time_ms` | INTEGER | Response latency |

### `code_versions`

Snapshots of code at debate milestones.

| Column | Type | Description |
|--------|------|-------------|
| `id` | UUID PK | Version identifier |
| `session_id` | UUID FK | References debate_sessions |
| `turn_id` | UUID FK | References debate_turns |
| `language` | VARCHAR | Programming language |
| `code` | TEXT | Code snapshot |
| `version_number` | INTEGER | Sequential version |
| `quality_score` | FLOAT | Quality score (0-1) |
| `test_pass_rate` | FLOAT | Percentage of tests passing |
| `metrics` | JSONB | Maintainability, complexity, security score |
| `diff_from_previous` | TEXT | Diff from prior version |
| `created_at` | TIMESTAMPTZ | Snapshot timestamp |

Migrations in `sql/schema/`. Repository interfaces in `internal/database/`.

---

## Section 2: Episodic Memory & Reflexion Framework

**New package: `internal/debate/reflexion/`**

### `episodic_memory.go`

Persistent memory buffer storing failure episodes:

- `Episode`: ID, SessionID, TurnID, AgentID, TaskDescription, AttemptNumber, Code, TestResults, FailureAnalysis, Reflection (verbal lesson learned), Improvement (what to do differently), Confidence, Timestamp
- `EpisodicMemoryBuffer`: in-memory store with DB persistence (`debate_turns.reflections` JSONB), configurable max size with FIFO eviction, retrieval by agent/session/similarity

### `reflection_generator.go`

Produces verbal reflections from failures:

- Input: failed code, test results, error messages, previous reflections
- Calls agent's LLM with structured prompt: "What went wrong? Why? What should change?"
- Returns `Reflection`: RootCause, WhatWentWrong, WhatToChangeNext, ConfidenceInFix
- This is "verbal reinforcement" from the Reflexion paper — text updates to memory, not weight updates

### `reflexion_loop.go`

Core retry-with-learning loop:

- `ReflexionLoop.Execute(task)`: Generate code -> Run tests -> If fail: generate reflection -> store episode -> regenerate with accumulated reflections in context -> repeat
- Max attempts configurable (default 3)
- Termination: all tests pass, OR max attempts, OR confidence threshold met
- Each attempt sees ALL prior reflections for that task (accumulated wisdom)

### `accumulated_wisdom.go`

Cross-session learning:

- Extracts generalizable insights from episodes (not task-specific details)
- Stores in `knowledge/repository.go` as reusable patterns
- On new debates, retrieves relevant wisdom based on task similarity (embedding-based if available, keyword fallback)

Integrates into existing protocol: when test-driven debate detects failures, Reflexion loop engages before moving to next debate round.

---

## Section 3: Complete Agent Roles & Adversarial Dynamics

### 10 New Role Templates (`internal/debate/agents/templates.go`)

| Role | System Prompt Focus | Primary Tools |
|------|-------------------|---------------|
| Generator | Produce complete, production-ready implementations | Code executor, formatters |
| Refactorer | Improve structure without changing behavior, smell detection | Static analyzers, formatters |
| PerformanceAnalyzer | Profile, optimize hotpaths, memory/CPU efficiency | Profilers, benchmarks |
| Security | Find vulnerabilities, OWASP top 10, injection vectors | Security scanners, fuzzers |
| Teacher | Explain decisions, document rationale, knowledge transfer | Documentation tools |
| Compiler | Validate syntax, type safety, build correctness | Language compilers, go vet |
| Executor | Run code in sandbox, collect runtime feedback, capture metrics | Docker executor, test runner |
| Judge/Adjudicator | Score solutions objectively against rubric (0-1 per criterion), declare winner | Scoring rubrics, metrics |
| Implementer | Turn architectural specs into concrete code, follow contracts | Code executor, LSP |
| Designer | High-level system design, component decomposition, interface specs | Design analysis, RAG |

### Adversarial Dynamics (`internal/debate/agents/adversarial.go`)

- `AdversarialProtocol` orchestrates Red vs Blue interaction
- **Red Team**: systematically breaks solutions — adversarial inputs, edge cases, security exploits, stress scenarios
- **Blue Team**: defends — patches vulnerabilities, hardens edge cases, adds validation
- **Round**: Red attacks -> Blue defends -> Red re-attacks (max 3 rounds or until Red finds no new issues)
- Integrates as optional phase between Optimization and Convergence
- Configurable via `DebateConfig.AdversarialEnabled`

### New Role Constants

`RoleCompiler`, `RoleExecutor`, `RoleJudge`, `RoleImplementer`, `RoleDesigner` added to `internal/debate/topology/topology.go`. Domain-role affinities updated.

---

## Section 4: Tree Topology & Planning Styles

### Tree Topology (`internal/debate/topology/tree.go`)

- `TreeTopology` implements `Topology` interface with hierarchical parent-child relationships
- Structure: root (Architect/Lead), intermediate (team leads per concern), leaf (specialist agents)
- `TreeNode`: AgentID, Role, Parent, Children, Level, Subtree responsibility
- Message routing: up (escalation), down (delegation), siblings through common parent
- `SelectLeader(phase)` returns responsible node for that phase's concern area
- `GetParallelGroups(phase)` returns independent subtrees for concurrent execution
- `BuildTree(agents, config)` auto-constructs from agent roles
- Rebalancing: failed agent's subtree redistributed to siblings

### Planning Styles (`internal/debate/topology/planning.go`)

- `PlanningStyle` enum: `CPDE`, `DPDE`
- **CPDE** (Centralized Planning, Decentralized Execution): root creates comprehensive plan, distributes to leaves, each executes independently, results aggregate up
- **DPDE** (Decentralized Planning, Decentralized Execution): each agent plans own subtask, coordinates through negotiation, consensus emerges bottom-up
- `PlanningStyleSelector`: auto-selects based on task characteristics. Configurable override
- Integrates with `CognitivePlanner` from `cognitive/cognitive_planning.go`

### Factory Update

`TopologyFactory.Create("tree", config)` support. `TopologyTree = "tree"` added to enum.

---

## Section 5: Communicative Dehallucination & Self-Evolvement

### Dehallucination Phase (`internal/debate/protocol/dehallucination.go`)

- Pre-debate clarification protocol following ChatDev's pattern
- Flow: Instructor sends task -> Assistant asks clarifying questions -> Instructor answers -> repeat until confidence > 0.9 or max 3 rounds
- `ClarificationRequest`: Question, Category (requirements/constraints/edge_cases/performance/integration), Priority
- `ClarificationResponse`: Answer, Confidence, RemainingAmbiguities
- Integrates as Phase -1 (before Proposal). Enabled by default for complex tasks, skipped for simple

### Self-Evolvement Phase (`internal/debate/protocol/self_evolvement.go`)

- Agents validate own code before entering debate
- Flow: generate -> self-test -> analyze failures -> refine -> refined solution enters debate
- Uses existing `SandboxedTestExecutor` and `LLMTestCaseGenerator`
- Integrates as Phase 0 (after Dehallucination, before Proposal)
- Max self-refinement iterations configurable (default 2)

### Protocol Update

Phase enum: `PhaseDehallucination`, `PhaseSelfEvolvement` added. Full order: Dehallucination -> Self-Evolvement -> Proposal -> Critique -> Review -> Optimization -> Adversarial -> Convergence. Each phase independently configurable.

---

## Section 6: Complete Voting Methods & Approval Gates

### Voting Methods

**Borda Count** (complete stub): Each voter ranks N candidates. Rank 1 = N-1 points, Rank 2 = N-2, etc. Highest total wins. Tie-breaking via existing strategies.

**Condorcet** (new): Pairwise comparison matrix. Condorcet winner = beats every other head-to-head. Cycle fallback to Borda with `FallbackUsed: true`.

**Plurality** (ensure complete): Most first-place votes wins.

**Unanimous** (ensure complete): All must agree. Returns `ConsensusReached: false` with disagreements if not.

**Auto-selection**: <3 agents = Unanimous, 3-5 = Weighted, 6+ = Borda. Override via config.

### Approval Gates (`internal/debate/gates/approval_gate.go`)

- `GateConfig`: Enabled (default false), GatePoints (phases list), Timeout (default 30min), NotificationChannels
- `GateRequest`: DebateID, SessionID, Phase, Summary, Artifacts (diff, tests, scores), RequestedAt
- `GateDecision`: Approved/Rejected/TimedOut, Reviewer, Reason, DecidedAt
- When enabled: pause -> store request -> SSE/WebSocket notify -> wait for `POST /v1/debates/{id}/approve` or timeout
- When disabled: auto-approve immediately
- New endpoints: `POST /v1/debates/{id}/approve`, `POST /v1/debates/{id}/reject`, `GET /v1/debates/{id}/gates`

---

## Section 7: External Integrations

### Benchmark Bridge (`internal/debate/evaluation/benchmark_bridge.go`)

- Connects `Benchmark` module to debate system
- `EvaluateDebateResult(result, benchmarkType) EvaluationScore`
- Types: `swe_bench`, `human_eval`, `mmlu`, `custom`
- `DebateBenchmarkSuite`: runs problems through debate, compares single vs multi-agent
- Custom metrics: correctness, maintainability, performance, security, test coverage

### Git Tool (`internal/debate/tools/git_tool.go`)

- Version control within debates via temporary git worktrees
- Operations: CreateBranch, CommitSnapshot, CreateDiff, GetFileHistory, GetBlame
- Each debate session gets branch: `debate/<session-id>`
- Cleanup after debate completes

### CI/CD Hooks (`internal/debate/tools/cicd_hook.go`)

- Trigger validation at hook points: PostProposal, PostOptimization, PostAdversarial, PostConvergence
- Actions: unit tests, linter, static analysis, security scan, benchmarks
- All local execution through existing containerized tools (matches "Manual CI/CD Only" Constitution rule)
- Results fed back as structured feedback

### Provenance & Audit (`internal/debate/audit/provenance.go`)

- Full audit trail: prompts, model versions, responses, tool calls, test results, votes, reflections, gate decisions
- Stored in `debate_sessions.metadata` and `debate_turns.metadata` JSONB
- `GetAuditTrail(sessionID) AuditTrail` for complete replay
- Enables reproducibility

---

## Section 8: Test Coverage

### Unit Tests

Every new and existing untested file gets `*_test.go` with table-driven tests:

- `agents/`: factory_test, specialization_test, templates_test, adversarial_test
- `cognitive/`: cognitive_planning_test
- `knowledge/`: repository_test, integration_test
- `testing/`: contrastive_analyzer_test, protocol_integration_test, test_case_generator_test, test_executor_test
- `tools/`: service_bridge_test, tool_integration_test, git_tool_test, cicd_hook_test
- `topology/`: chain_test, factory_test, graph_mesh_test, star_test, tree_test, planning_test
- `reflexion/`: episodic_memory_test, reflection_generator_test, reflexion_loop_test, accumulated_wisdom_test
- `protocol/`: dehallucination_test, self_evolvement_test
- `voting/`: borda_test, condorcet_test, plurality_test, unanimous_test
- `gates/`: approval_gate_test
- `evaluation/`: benchmark_bridge_test
- `audit/`: provenance_test

### Integration Tests (`tests/integration/`)

- `debate_reflexion_integration_test.go` — failure -> reflection -> retry -> success
- `debate_adversarial_integration_test.go` — Red attack -> Blue defend -> re-attack
- `debate_tree_topology_integration_test.go` — hierarchical coordination
- `debate_cross_learning_integration_test.go` — lessons from debate A improve debate B
- `debate_approval_gate_integration_test.go` — pause/resume/approve/reject/timeout
- `debate_full_protocol_integration_test.go` — all 8 phases end-to-end
- `debate_persistence_integration_test.go` — DB write/read/recovery
- `debate_tool_integration_test.go` — MCP/LSP/RAG/formatter calls within debate

### E2E Tests (`tests/e2e/`)

- `debate_full_workflow_e2e_test.go` — HTTP API full cycle with SSE monitoring
- `debate_concurrent_e2e_test.go` — multiple simultaneous debates, verify isolation
- `debate_recovery_e2e_test.go` — start -> kill server -> restart -> verify recovery

### Security Tests (`tests/security/`)

- `debate_security_test.go` — prompt injection, code injection, resource exhaustion, information leakage, malformed responses, debate poisoning

### Stress Tests (`tests/stress/`)

- `debate_stress_test.go` — 100+ concurrent debates, agent failure recovery, memory leak detection, deadlock detection, provider cascade, topology reconfiguration

### Benchmark Tests (`tests/performance/`)

- `debate_benchmark_test.go` — round latency, voting algorithm comparison, consensus scaling, topology init time, message throughput, memory under load

### Challenge Scripts (`challenges/scripts/`)

12 new challenges validating actual behavior (HTTP calls, DB queries, real outputs):

- `debate_reflexion_challenge.sh`
- `debate_adversarial_dynamics_challenge.sh`
- `debate_tree_topology_challenge.sh`
- `debate_dehallucination_challenge.sh`
- `debate_self_evolvement_challenge.sh`
- `debate_condorcet_voting_challenge.sh`
- `debate_approval_gate_challenge.sh`
- `debate_persistence_challenge.sh`
- `debate_benchmark_integration_challenge.sh`
- `debate_provenance_audit_challenge.sh`
- `debate_deadlock_detection_challenge.sh`
- `debate_git_integration_challenge.sh`

Existing grep-only challenge scripts upgraded to include HTTP behavioral validation.

---

## Section 9: Documentation & Diagrams

### Updated Documentation

- `docs/features/AI_DEBATE_ORCHESTRATOR.md` — Major rewrite: all 8 phases, 16+ roles, 4 topologies, 6 voting methods, Reflexion, adversarial, gates, API reference
- `docs/features/ai-debate-configuration.md` — All new config options
- `CLAUDE.md` — Architecture section updated for new packages, tables, challenges, phases, roles
- `AGENTS.md` — Synchronized per Constitution
- `docs/MODULES.md` — Updated debate package catalog

### New Documentation

- `docs/features/debate-reflexion-framework.md` — Episodic memory, verbal reinforcement, mathematical model
- `docs/features/debate-adversarial-dynamics.md` — Red/Blue protocol, attack categories, defense strategies
- `docs/features/debate-approval-gates.md` — Config guide, API reference, SSE format, timeout behavior
- `docs/guides/debate-database-schema.md` — ER diagram, table descriptions, migration guide, query examples

### New Diagrams (`docs/diagrams/`)

- `debate-full-protocol.mmd` — All 8 phases with decision points
- `debate-reflexion-loop.mmd` — Generate -> test -> fail -> reflect -> retry
- `debate-tree-topology.mmd` — Tree structure with message routing
- `debate-adversarial-round.mmd` — Red attack -> Blue defend cycle
- `debate-voting-methods.mmd` — All 6 methods with selection criteria
- `debate-database-er.mmd` — ER diagram for new tables
- `debate-approval-gate-flow.mmd` — Gate pause -> notify -> decide

---

## Implementation Approach

**Phase 1 — Shared Infrastructure:**
DB schema migrations, episodic memory store, approval gate framework, missing role/template registry

**Phase 2 — Feature Implementation (grouped by dependency):**
- Group A (independent): Tree topology, Condorcet/Borda voting, all missing role templates
- Group B (depends on episodic memory): Reflexion framework, cross-debate learning enhancement
- Group C (depends on roles): Red/Blue Team adversarial dynamics, Self-Evolvement phase, Communicative Dehallucination
- Group D (integration): SWE-bench/HumanEval, Git/CI-CD hooks, human-in-the-loop gates

**Phase 3 — Test Completeness:**
Fill all unit test gaps (27+ files), stress/security/benchmark tests, update challenge scripts

**Phase 4 — Documentation:**
Design docs, diagrams, manuals, guide updates, CLAUDE.md/AGENTS.md sync
