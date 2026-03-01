# Comprehensive Debate System - Master Implementation Plan

## Project Overview

Complete implementation of the multi-agent debate system from:
- `docs/requests/debate/001 AI Debate Research Kimi.md`
- `docs/requests/debate/003 AI Debate Research MiniMax.md`
- `docs/requests/debate/004 AI Debate.md`

**Goal**: Production-ready system with 100% test coverage and comprehensive challenges.

---

## Phase Tracking

| Phase | Status | Files | Tests | Challenges | Progress |
|-------|--------|-------|-------|------------|----------|
| 0 | 🔄 In Progress | - | - | - | Creating plan |
| 1 | ⏳ Pending | 15+ | 50+ | 3 | 0% |
| 2 | ⏳ Pending | 20+ | 60+ | 3 | 0% |
| 3 | ⏳ Pending | 30+ | 80+ | 4 | 0% |
| 4 | ⏳ Pending | 25+ | 70+ | 3 | 0% |
| 5 | ⏳ Pending | 20+ | 60+ | 3 | 0% |
| 6 | ⏳ Pending | 15+ | 50+ | 2 | 0% |
| 7 | ⏳ Pending | 15+ | 40+ | 2 | 0% |
| 8 | ⏳ Pending | 10+ | 30+ | 2 | 0% |
| 9 | ⏳ Pending | - | 200+ | - | 0% |
| 10 | ⏳ Pending | - | 50+ | - | 0% |
| 11 | ⏳ Pending | - | 30+ | - | 0% |
| 12 | ⏳ Pending | - | - | 15+ | 0% |
| 13 | ⏳ Pending | 10+ | - | - | 0% |
| 14 | ⏳ Pending | - | - | 5+ | 0% |

**Total Estimated**: 200+ source files, 700+ tests, 40+ challenges

---

## Phase 0: Master Plan Creation (Current)

**Duration**: 30 minutes
**Deliverables**:
- [x] Master plan document (this file)
- [x] Directory structure
- [ ] Phase 1 plan document
- [ ] Progress tracking system
- [ ] Commit strategy

**Commits**: 1 (plan creation)

---

## Phase 1: Core Agent Framework & Types

**Duration**: 2-3 hours
**Goal**: Foundation types, interfaces, and base agent functionality

### 1.1 Core Types (`internal/debate/comprehensive/types/`)
- [ ] `agent.go` - Agent interface and base implementation
- [ ] `message.go` - Inter-agent communication types
- [ ] `context.go` - Shared context and memory
- [ ] `consensus.go` - Consensus algorithms
- [ ] `score.go` - Scoring and confidence metrics

### 1.2 Agent Framework (`internal/debate/comprehensive/agents/`)
- [ ] `base.go` - Base agent with common functionality
- [ ] `pool.go` - Agent pool management
- [ ] `factory.go` - Agent creation factory
- [ ] `coordinator.go` - Inter-agent coordination
- [ ] `roles.go` - Role definitions and capabilities

### 1.3 Memory System (`internal/debate/comprehensive/memory/`)
- [ ] `short_term.go` - Short-term memory (conversation history)
- [ ] `long_term.go` - Long-term memory (lessons learned)
- [ ] `episodic.go` - Episodic memory (Reflexion)
- [ ] `storage.go` - PostgreSQL storage implementation

### 1.4 Utilities (`internal/debate/comprehensive/utils/`)
- [ ] `prompts.go` - Prompt templates
- [ ] `parsing.go` - Response parsing utilities
- [ ] `validation.go` - Input validation
- [ ] `crypto.go` - ID generation, hashing

**Tests**: 50+ unit tests
**Challenges**: 3 (agent_creation, memory_system, coordination)

**Commits**: 5-7 commits

---

## Phase 2: Tool Suite Implementation

**Duration**: 3-4 hours
**Goal**: Complete tool suite for agent actions

### 2.1 Code Tool (`internal/debate/comprehensive/tools/code/`)
- [ ] `reader.go` - Read files with line tracking
- [ ] `writer.go` - Write/create files with backups
- [ ] `updater.go` - Update file sections
- [ ] `deleter.go` - Safe file deletion
- [ ] `search.go` - File search and glob patterns
- [ ] `schema.go` - Tool schemas for LLMs

### 2.2 Command Tool (`internal/debate/comprehensive/tools/command/`)
- [ ] `executor.go` - Bash/Go command execution
- [ ] `sandbox.go` - Sandboxed execution (Docker)
- [ ] `limits.go` - Resource limits (time, memory)
- [ ] `parser.go` - Output parsing
- [ ] `security.go` - Security validation

### 2.3 Database Tool (`internal/debate/comprehensive/tools/database/`)
- [ ] `query.go` - SQL query execution
- [ ] `schema.go` - Schema introspection
- [ ] `migration.go` - Migration support
- [ ] `security.go` - SQL injection prevention

### 2.4 Analysis Tools (`internal/debate/comprehensive/tools/analysis/`)
- [ ] `static.go` - Static analysis (linting, complexity)
- [ ] `security.go` - Security scanning (SAST)
- [ ] `performance.go` - Performance profiling
- [ ] `coverage.go` - Test coverage analysis
- [ ] `smells.go` - Code smell detection

### 2.5 Tool Registry (`internal/debate/comprehensive/tools/`)
- [ ] `registry.go` - Tool registration and discovery
- [ ] `dispatcher.go` - Tool call dispatching
- [ ] `validation.go` - Tool input validation
- [ ] `result.go` - Tool result handling
- [ ] `integration.go` - Integration with LLM providers

**Tests**: 60+ unit tests
**Challenges**: 3 (code_tools, command_tools, analysis_tools)

**Commits**: 6-8 commits

---

## Phase 3: Agent Roles Implementation (8 Agents)

**Duration**: 6-8 hours
**Goal**: Complete implementation of all 8 specialized agents

### 3.1 Architect Agent (`internal/debate/comprehensive/agents/architect/`)
- [ ] `agent.go` - Main agent implementation
- [ ] `design.go` - System design capabilities
- [ ] `decomposition.go` - Task decomposition
- [ ] `planning.go` - Architecture planning
- [ ] `patterns.go` - Design pattern knowledge
- [ ] Tests: 10+ tests
- [ ] Prompts: 5+ specialized prompts

### 3.2 Generator Agent (`internal/debate/comprehensive/agents/generator/`)
- [ ] `agent.go` - Main agent implementation
- [ ] `codegen.go` - Code generation
- [ ] `templates.go` - Template-based generation
- [ ] `languages.go` - Language-specific support (Go, Bash, SQL)
- [ ] `quality.go` - Initial quality checks
- [ ] Tests: 10+ tests
- [ ] Prompts: 5+ specialized prompts

### 3.3 Critic Agent (`internal/debate/comprehensive/agents/critic/`)
- [ ] `agent.go` - Main agent implementation
- [ ] `analysis.go` - Code analysis
- [ ] `bugs.go` - Bug identification
- [ ] `security.go` - Security flaw detection
- [ ] `performance.go` - Performance issue detection
- [ ] `style.go` - Style guide compliance
- [ ] Tests: 10+ tests
- [ ] Prompts: 5+ specialized prompts

### 3.4 Refactoring Agent (`internal/debate/comprehensive/agents/refactorer/`)
- [ ] `agent.go` - Main agent implementation
- [ ] `transform.go` - Code transformation
- [ ] `patterns.go` - Refactoring patterns
- [ ] `smells.go` - Code smell fixes
- [ ] `preservation.go` - Behavior preservation
- [ ] Tests: 10+ tests
- [ ] Prompts: 5+ specialized prompts

### 3.5 Tester Agent (`internal/debate/comprehensive/agents/tester/`)
- [ ] `agent.go` - Main agent implementation
- [ ] `generation.go` - Test case generation
- [ ] `edge_cases.go` - Edge case identification
- [ ] `coverage.go` - Coverage maximization
- [ ] `execution.go` - Test execution
- [ ] Tests: 10+ tests
- [ ] Prompts: 5+ specialized prompts

### 3.6 Validator Agent (`internal/debate/comprehensive/agents/validator/`)
- [ ] `agent.go` - Main agent implementation
- [ ] `correctness.go` - Correctness verification
- [ ] `formal.go` - Formal verification support
- [ ] `equivalence.go` - Behavioral equivalence
- [ ] `metrics.go` - Quality metrics
- [ ] Tests: 10+ tests
- [ ] Prompts: 5+ specialized prompts

### 3.7 Security Agent (`internal/debate/comprehensive/agents/security/`)
- [ ] `agent.go` - Main agent implementation
- [ ] `vulnerabilities.go` - Vulnerability detection
- [ ] `scanning.go` - Security scanning
- [ ] `attack_surface.go` - Attack surface analysis
- [ ] `mitigation.go` - Mitigation strategies
- [ ] Tests: 10+ tests
- [ ] Prompts: 5+ specialized prompts

### 3.8 Performance Agent (`internal/debate/comprehensive/agents/performance/`)
- [ ] `agent.go` - Main agent implementation
- [ ] `profiling.go` - Performance profiling
- [ ] `optimization.go` - Optimization strategies
- [ ] `benchmarking.go` - Benchmark creation
- [ ] `resources.go` - Resource usage analysis
- [ ] Tests: 10+ tests
- [ ] Prompts: 5+ specialized prompts

**Tests**: 80+ unit tests
**Challenges**: 4 (architect, critic, tester, security)

**Commits**: 16-20 commits

---

## Phase 4: Phase Orchestration (6 Phases)

**Duration**: 4-5 hours
**Goal**: Implement the 6-phase debate workflow

### 4.1 Planning Phase (`internal/debate/comprehensive/phases/planning/`)
- [ ] `phase.go` - Phase orchestration
- [ ] `decomposition.go` - Task decomposition logic
- [ ] `assignment.go` - Agent assignment
- [ ] `dependencies.go` - Dependency analysis
- [ ] Tests: 15+ tests

### 4.2 Generation Phase (`internal/debate/comprehensive/phases/generation/`)
- [ ] `phase.go` - Phase orchestration
- [ ] `parallel.go` - Parallel generation
- [ ] `variants.go` - Multiple variant generation
- [ ] `sampling.go` - Probabilistic sampling
- [ ] Tests: 15+ tests

### 4.3 Debate Phase (`internal/debate/comprehensive/phases/debate/`)
- [ ] `phase.go` - Phase orchestration
- [ ] `rounds.go` - Round management
- [ ] `critique.go` - Critique mechanisms
- [ ] `defense.go` - Defense mechanisms
- [ ] `convergence.go` - Convergence detection
- [ ] Tests: 20+ tests

### 4.4 Validation Phase (`internal/debate/comprehensive/phases/validation/`)
- [ ] `phase.go` - Phase orchestration
- [ ] `testing.go` - Test execution
- [ ] `static.go` - Static analysis
- [ ] `coverage.go` - Coverage validation
- [ ] Tests: 15+ tests

### 4.5 Refactoring Phase (`internal/debate/comprehensive/phases/refactoring/`)
- [ ] `phase.go` - Phase orchestration
- [ ] `pipeline.go` - Refactoring pipeline
- [ ] `validation.go` - Post-refactor validation
- [ ] `metrics.go` - Quality metrics
- [ ] Tests: 15+ tests

### 4.6 Integration Phase (`internal/debate/comprehensive/phases/integration/`)
- [ ] `phase.go` - Phase orchestration
- [ ] `consistency.go` - Cross-file consistency
- [ ] `compilation.go` - Compilation verification
- [ ] `deployment.go` - Deployment preparation
- [ ] Tests: 15+ tests

**Tests**: 70+ unit tests
**Challenges**: 3 (phase_orchestration, debate_flow, validation_pipeline)

**Commits**: 12-15 commits

---

## Phase 5: Debate Engine & Algorithms

**Duration**: 4-5 hours
**Goal**: Core debate algorithms and consensus mechanisms

### 5.1 Engine Core (`internal/debate/comprehensive/engine/`)
- [ ] `engine.go` - Main debate engine
- [ ] `scheduler.go` - Turn scheduling
- [ ] `dispatcher.go` - Message dispatching
- [ ] `state.go` - State management
- [ ] `events.go` - Event system

### 5.2 Consensus Algorithms (`internal/debate/comprehensive/engine/consensus/`)
- [ ] `voting.go` - Weighted voting
- [ ] `aggregation.go` - Confidence aggregation
- [ ] `judge.go` - Judge agent logic
- [ ] `threshold.go` - Threshold-based decisions
- [ ] Tests: 15+ tests

### 5.3 Conflict Resolution (`internal/debate/comprehensive/engine/conflict/`)
- [ ] `detection.go` - Conflict detection
- [ ] `resolution.go` - Resolution strategies
- [ ] `escalation.go` - Escalation to humans
- [ ] `mediation.go` - Mediation logic
- [ ] Tests: 15+ tests

### 5.4 Convergence (`internal/debate/comprehensive/engine/convergence/`)
- [ ] `criteria.go` - Convergence criteria
- [ ] `detection.go` - Loop detection
- [ ] `stagnation.go` - Stagnation detection
- [ ] `termination.go` - Graceful termination
- [ ] Tests: 15+ tests

### 5.5 Optimization (`internal/debate/comprehensive/engine/optimization/`)
- [ ] `parallel.go` - Parallel execution
- [ ] `caching.go` - Response caching
- [ ] `pruning.go` - Probabilistic pruning
- [ ] `cost.go` - Cost optimization
- [ ] Tests: 15+ tests

**Tests**: 60+ unit tests
**Challenges**: 3 (consensus, convergence, optimization)

**Commits**: 10-12 commits

---

## Phase 6: Quality Gates & Validation

**Duration**: 3-4 hours
**Goal**: Quality assurance and validation pipeline

### 6.1 Quality Gates (`internal/debate/comprehensive/validation/`)
- [ ] `gates.go` - Quality gate definitions
- [ ] `correctness.go` - Correctness checks
- [ ] `coverage.go` - Coverage requirements
- [ ] `performance.go` - Performance thresholds
- [ ] `security.go` - Security requirements
- [ ] Tests: 20+ tests

### 6.2 Test Integration (`internal/debate/comprehensive/validation/testing/`)
- [ ] `runner.go` - Test runner integration
- [ ] `frameworks.go` - Framework support (Go test, etc.)
- [ ] `reporting.go` - Test reporting
- [ ] `coverage.go` - Coverage analysis
- [ ] Tests: 15+ tests

### 6.3 Static Analysis (`internal/debate/comprehensive/validation/static/`)
- [ ] `linters.go` - Linter integration
- [ ] `complexity.go` - Complexity analysis
- [ ] `smells.go` - Code smell detection
- [ ] `standards.go` - Standard compliance
- [ ] Tests: 15+ tests

**Tests**: 50+ unit tests
**Challenges**: 2 (quality_gates, validation_pipeline)

**Commits**: 8-10 commits

---

## Phase 7: Advanced Features

**Duration**: 3-4 hours
**Goal**: Red/Blue team, Reflexion, and other advanced features

### 7.1 Red/Blue Team (`internal/debate/comprehensive/adversarial/`)
- [ ] `red_team.go` - Red team agent (attacker)
- [ ] `blue_team.go` - Blue team agent (defender)
- [ ] `coordinator.go` - Adversarial coordinator
- [ ] `strategies.go` - Attack/defense strategies
- [ ] Tests: 15+ tests

### 7.2 Reflexion (`internal/debate/comprehensive/reflexion/`)
- [ ] `memory.go` - Episodic memory buffer
- [ ] `reflection.go` - Reflection generation
- [ ] `learning.go` - Learning from failures
- [ ] `feedback.go` - Feedback incorporation
- [ ] Tests: 15+ tests

### 7.3 Cross-File Consistency (`internal/debate/comprehensive/consistency/`)
- [ ] `analyzer.go` - Cross-file analysis
- [ ] `dependencies.go` - Dependency tracking
- [ ] `propagation.go` - Change propagation
- [ ] `validation.go` - Consistency validation
- [ ] Tests: 10+ tests

**Tests**: 40+ unit tests
**Challenges**: 2 (adversarial_debate, reflexion_learning)

**Commits**: 8-10 commits

---

## Phase 8: Integration & Wiring

**Duration**: 2-3 hours
**Goal**: Connect everything together

### 8.1 System Integration (`internal/debate/comprehensive/`)
- [ ] Update `system.go` with full implementation
- [ ] `wiring.go` - Dependency injection
- [ ] `config.go` - Configuration management
- [ ] `lifecycle.go` - System lifecycle management

### 8.2 Handler Integration (`internal/handlers/`)
- [ ] Update `openai_compatible.go` to use new system
- [ ] `debate_handler.go` - Debate-specific endpoints
- [ ] `monitoring.go` - Debate monitoring endpoints

### 8.3 Provider Integration (`internal/llm/`)
- [ ] Provider selection for each agent role
- [ ] Fallback chains per agent
- [ ] Score-based routing

### 8.4 Router Integration (`internal/router/`)
- [ ] Register debate endpoints
- [ ] Wire up handlers
- [ ] Add middleware

**Tests**: 30+ integration tests
**Challenges**: 2 (system_integration, end_to_end)

**Commits**: 6-8 commits

---

## Phase 9: Unit Tests (100% Coverage)

**Duration**: 4-5 hours
**Goal**: Comprehensive unit test suite

### 9.1 Agent Tests
- [ ] Test each agent role independently
- [ ] Mock LLM responses
- [ ] Test tool usage
- [ ] Test error handling
- [ ] **Target**: 100+ tests

### 9.2 Tool Tests
- [ ] Test each tool
- [ ] Mock file system
- [ ] Mock command execution
- [ ] Test sandboxing
- [ ] **Target**: 100+ tests

### 9.3 Phase Tests
- [ ] Test each phase
- [ ] Test phase transitions
- [ ] Test convergence
- [ ] **Target**: 80+ tests

### 9.4 Engine Tests
- [ ] Test consensus algorithms
- [ ] Test scheduling
- [ ] Test conflict resolution
- [ ] **Target**: 60+ tests

**Total Unit Tests**: 340+
**Coverage Target**: 100%

**Commits**: 10-12 commits

---

## Phase 10: Integration Tests

**Duration**: 2-3 hours
**Goal**: Test component interactions

### 10.1 Multi-Agent Tests
- [ ] Test agent communication
- [ ] Test role-based interactions
- [ ] Test debate rounds
- [ ] **Target**: 20+ tests

### 10.2 Tool Integration Tests
- [ ] Test tool-augmented agents
- [ ] Test real command execution (sandboxed)
- [ ] Test file operations
- [ ] **Target**: 15+ tests

### 10.3 Phase Integration Tests
- [ ] Test full phase sequences
- [ ] Test phase transitions
- [ ] Test error recovery
- [ ] **Target**: 15+ tests

**Total Integration Tests**: 50+

**Commits**: 5-6 commits

---

## Phase 11: E2E Tests

**Duration**: 2-3 hours
**Goal**: End-to-end testing with real providers

### 11.1 Full Debate Tests
- [ ] Test complete debate workflow
- [ ] Test with real LLM providers
- [ ] Test code generation and validation
- [ ] **Target**: 10+ tests

### 11.2 Performance Tests
- [ ] Test debate performance
- [ ] Test parallel execution
- [ ] Test caching effectiveness
- [ ] **Target**: 10+ tests

### 11.3 Reliability Tests
- [ ] Test failure recovery
- [ ] Test timeout handling
- [ ] Test provider fallbacks
- [ ] **Target**: 10+ tests

**Total E2E Tests**: 30+

**Commits**: 4-5 commits

---

## Phase 12: Challenges (15+ Scripts)

**Duration**: 3-4 hours
**Goal**: Comprehensive validation challenges

### 12.1 Core Challenges
- [ ] `agent_roles_challenge.sh` - Validate all 8 agents
- [ ] `tool_suite_challenge.sh` - Validate tool suite
- [ ] `phase_orchestration_challenge.sh` - Validate 6 phases
- [ ] `consensus_challenge.sh` - Validate consensus mechanisms
- [ ] `quality_gates_challenge.sh` - Validate quality gates

### 12.2 Feature Challenges
- [ ] `adversarial_debate_challenge.sh` - Red/Blue team
- [ ] `reflexion_challenge.sh` - Self-correction
- [ ] `cross_file_consistency_challenge.sh` - Multi-file support
- [ ] `convergence_challenge.sh` - Convergence detection
- [ ] `performance_challenge.sh` - Performance optimization

### 12.3 Integration Challenges
- [ ] `system_integration_challenge.sh` - Full system
- [ ] `debate_api_challenge.sh` - API endpoints
- [ ] `provider_integration_challenge.sh` - LLM providers
- [ ] `end_to_end_challenge.sh` - Complete workflow

### 12.4 Quality Challenges
- [ ] `test_coverage_challenge.sh` - 100% coverage validation
- [ ] `documentation_challenge.sh` - Documentation completeness
- [ ] `code_quality_challenge.sh` - Code quality metrics

**Total Challenges**: 15+

**Commits**: 5-6 commits

---

## Phase 13: Documentation

**Duration**: 2-3 hours
**Goal**: Comprehensive documentation

### 13.1 API Documentation (`internal/debate/comprehensive/docs/`)
- [ ] `API.md` - Public API reference
- [ ] `AGENTS.md` - Agent roles and capabilities
- [ ] `TOOLS.md` - Tool reference
- [ ] `PHASES.md` - Phase documentation
- [ ] `CONFIGURATION.md` - Configuration guide

### 13.2 Architecture Documentation
- [ ] `ARCHITECTURE.md` - System architecture
- [ ] `DESIGN.md` - Design decisions
- [ ] `DATA_FLOW.md` - Data flow diagrams
- [ ] `SEQUENCE.md` - Sequence diagrams

### 13.3 User Documentation
- [ ] `README.md` - Main documentation
- [ ] `QUICKSTART.md` - Getting started guide
- [ ] `EXAMPLES.md` - Usage examples
- [ ] `TROUBLESHOOTING.md` - Troubleshooting guide

### 13.4 Developer Documentation
- [ ] `CONTRIBUTING.md` - Contribution guide
- [ ] `TESTING.md` - Testing guide
- [ ] `DEPLOYMENT.md` - Deployment guide

**Total Documents**: 15+

**Commits**: 4-5 commits

---

## Phase 14: Final Validation & Deployment

**Duration**: 2-3 hours
**Goal**: Final checks and deployment

### 14.1 Validation
- [ ] Run all unit tests
- [ ] Run all integration tests
- [ ] Run all E2E tests
- [ ] Run all challenges
- [ ] Verify 100% test coverage
- [ ] Verify all documentation

### 14.2 Build & Deploy
- [ ] Build Docker image
- [ ] Run in test environment
- [ ] Validate functionality
- [ ] Deploy to production
- [ ] Monitor metrics

### 14.3 Handoff
- [ ] Create handoff document
- [ ] Document known issues
- [ ] Document future improvements
- [ ] Archive progress tracking

**Commits**: 3-4 commits

---

## Progress Tracking

### How to Use This Plan

1. **Start Phase**: Update status to "🔄 In Progress"
2. **Complete Task**: Check off [ ] → [x]
3. **Commit**: After each sub-phase, commit with message format:
   ```
   feat(debate): Phase X.Y - Description
   
   - Implemented feature A
   - Implemented feature B
   - Added tests
   
   Progress: X/Y tasks complete
   ```
4. **Pause**: Commit current work, update status to "⏸️ Paused"
5. **Resume**: Update status to "🔄 In Progress", continue

### Current Status

**Phase**: 0 - Master Plan Creation
**Status**: 🔄 In Progress
**Progress**: 1/1 tasks (100%)

### Daily Log

| Date | Phase | Tasks Completed | Commits | Notes |
|------|-------|-----------------|---------|-------|
| 2026-03-01 | 0 | 1/1 | 0 | Plan creation |

---

## Commit Strategy

### Commit Message Format
```
type(scope): Brief description

Detailed description of what was implemented:
- Feature A
- Feature B
- Tests added

Progress tracking:
- Phase: X
- Tasks: Y/Z complete
- Tests: N added
- Coverage: P%
```

### Types
- `feat`: New features
- `fix`: Bug fixes
- `test`: Tests only
- `docs`: Documentation
- `refactor`: Code refactoring
- `chore`: Maintenance

### Frequency
- Commit after each sub-phase
- Commit after major test additions
- Commit after documentation updates
- Never leave uncommitted work for >2 hours

---

## Estimated Timeline

| Phase | Duration | Cumulative |
|-------|----------|------------|
| 0 | 30 min | 30 min |
| 1 | 3 hrs | 3.5 hrs |
| 2 | 4 hrs | 7.5 hrs |
| 3 | 8 hrs | 15.5 hrs |
| 4 | 5 hrs | 20.5 hrs |
| 5 | 5 hrs | 25.5 hrs |
| 6 | 4 hrs | 29.5 hrs |
| 7 | 4 hrs | 33.5 hrs |
| 8 | 3 hrs | 36.5 hrs |
| 9 | 5 hrs | 41.5 hrs |
| 10 | 3 hrs | 44.5 hrs |
| 11 | 3 hrs | 47.5 hrs |
| 12 | 4 hrs | 51.5 hrs |
| 13 | 3 hrs | 54.5 hrs |
| 14 | 3 hrs | 57.5 hrs |

**Total Estimated Time**: ~58 hours (7.25 working days)

---

## Notes

- This is an ambitious implementation requiring 7+ days of focused work
- Can be paused and resumed at any phase boundary
- Each phase produces working, testable code
- Challenges validate each phase independently
- Final system will be production-ready

**Next Step**: Complete Phase 0, then start Phase 1
