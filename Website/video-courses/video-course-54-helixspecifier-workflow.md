# Video Course 54: HelixSpecifier Spec-Driven Development Workflow

## Course Overview

**Duration**: 2 hours 30 minutes
**Level**: Advanced
**Prerequisites**: Course 01-Fundamentals, Course 12-Advanced-Workflows, familiarity with software development lifecycle practices

HelixSpecifier is the Spec-Driven Development (SDD) fusion engine for HelixAgent. This course covers its 3-pillar architecture, the 7-phase SDD flow, effort classification, auto-activation, and a complete spec-to-code walkthrough.

---

## Learning Objectives

By the end of this course, you will be able to:

1. Explain the 3-pillar architecture (SpecKit, Superpowers, GSD) and how they fuse
2. Walk through the 7-phase SDD flow from Constitution to Implementation
3. Configure effort classification and work granularity detection
4. Understand auto-activation triggers and phase caching
5. Execute a complete spec-driven development workflow end to end
6. Integrate HelixSpecifier with the AI debate ensemble via DebateFunc injection

---

## Module 1: Architecture Overview (30 min)

### 1.1 What Is HelixSpecifier?

**Video: Spec-Driven Development Fusion Engine** (10 min)

- Module path: `digital.vasic.helixspecifier` located in `HelixSpecifier/`
- 27 packages (21 core + 6 test suites), 835+ tests
- Active by default; opt out with `-tags nohelixspecifier`
- Bridges specification, testing, and execution into a unified workflow

### 1.2 The 3-Pillar Architecture

**Video: SpecKit + Superpowers + GSD** (12 min)

| Pillar       | Role                                    | Key Capability                        |
|--------------|-----------------------------------------|---------------------------------------|
| SpecKit      | 7-phase SDD lifecycle                   | Constitution, Specify, Clarify, Plan, Tasks, Analyze, Implement |
| Superpowers  | TDD and subagent orchestration          | Test-first validation, parallel subagents |
| GSD          | Milestone tracking and delivery         | Get Stuff Done milestones, progress gates  |

- 3-pillar fusion combines all three into adaptive ceremony scaling
- Ceremony level adjusts based on work granularity (5 levels)

### 1.3 Integration Points

**Video: HelixSpecifier in the HelixAgent Ecosystem** (8 min)

- DebateFunc injection allows real multi-LLM debate during specification phases
- Intent classifier routes requests to appropriate SDD phases
- Spec memory persists specifications across sessions
- CLI agent adapters expose SDD workflow to all 48 agents

### Hands-On Lab 1

Verify HelixSpecifier is active and inspect its configuration:

```bash
# Build with HelixSpecifier (default)
make build

# Check HelixSpecifier status
curl http://localhost:7061/v1/helixspecifier/status

# Build without HelixSpecifier
go build -tags nohelixspecifier ./cmd/helixagent
```

---

## Module 2: The 7-Phase SDD Flow (35 min)

### 2.1 Phase 1 -- Constitution

**Video: Establishing Project Principles** (5 min)

- Loads the project Constitution as the foundational contract
- All subsequent phases must comply with Constitution rules
- Constitution auto-update via ConstitutionWatcher feeds changes into HelixSpecifier

### 2.2 Phase 2 -- Specify

**Video: Writing Specifications** (5 min)

- Captures functional and non-functional requirements
- Structured specification templates with acceptance criteria
- DebateFunc injection enables multi-LLM refinement of specifications

### 2.3 Phase 3 -- Clarify

**Video: Resolving Ambiguities** (5 min)

- Identifies gaps and contradictions in specifications
- Generates clarification questions ranked by impact
- Interactive or automated resolution based on ceremony level

### 2.4 Phase 4 -- Plan

**Video: Architecture and Design Planning** (5 min)

- Produces architecture decisions and component decomposition
- Dependency analysis between components
- Risk identification and mitigation strategies

### 2.5 Phase 5 -- Tasks

**Video: Task Breakdown and Sequencing** (5 min)

- Decomposes plan into ordered implementation tasks
- Dependency graph generation with critical path analysis
- Effort estimation per task using historical data

### 2.6 Phase 6 -- Analyze

**Video: Pre-Implementation Analysis** (5 min)

- Static analysis of existing codebase for impact assessment
- Test coverage gap identification
- Security and performance implication review

### 2.7 Phase 7 -- Implement

**Video: Guided Implementation** (5 min)

- Task-by-task implementation with Superpowers TDD validation
- GSD milestone gates at key checkpoints
- Automated verification against original specifications

### Hands-On Lab 2

Walk through a small feature using all 7 phases:

1. Define a Constitution rule for the feature
2. Write a specification with acceptance criteria
3. Generate and review clarification questions
4. Create an architectural plan
5. Break the plan into tasks
6. Run pre-implementation analysis
7. Implement with TDD validation

---

## Module 3: Effort Classification and Auto-Activation (25 min)

### 3.1 Work Granularity Detection

**Video: 5 Levels of Granularity** (10 min)

| Level | Name                    | Example                        | SDD Triggered |
|-------|-------------------------|--------------------------------|---------------|
| 1     | Single Action           | Fix a typo                     | No            |
| 2     | Small Creation          | Add a helper function          | No            |
| 3     | Big Creation            | Add a new API endpoint         | Yes           |
| 4     | Whole Functionality     | Implement a new provider       | Yes           |
| 5     | Refactoring             | Extract a module               | Yes           |

- Auto-activation triggers for levels 3, 4, and 5
- Configurable thresholds in HelixSpecifier settings

### 3.2 Adaptive Ceremony Scaling

**Video: Ceremony Levels** (8 min)

- Minimal ceremony for Big Creation (skip Clarify, lightweight Plan)
- Standard ceremony for Whole Functionality (all phases, moderate depth)
- Full ceremony for Refactoring (exhaustive analysis, cross-impact review)

### 3.3 Phase Caching

**Video: Resumption and Persistence** (7 min)

- Phase outputs cached in `.speckit/cache/` for resumption
- Interrupted workflows resume from the last completed phase
- Cache invalidation on Constitution or specification changes

### Hands-On Lab 3

Test auto-activation with different work granularity levels:

```bash
# Simulate a small change (no SDD activation)
# vs. a large change (SDD auto-activates)
# Observe the intent classifier routing
curl -X POST http://localhost:7061/v1/helixspecifier/classify \
  -H "Content-Type: application/json" \
  -d '{"description": "Extract the caching layer into a separate module"}'
```

---

## Module 4: DebateFunc Injection and Fusion (20 min)

### 4.1 DebateFunc Injection

**Video: Real Multi-LLM Debate in SDD** (10 min)

- HelixSpecifier injects a DebateFunc into specification phases
- Multiple LLMs debate the quality and completeness of specifications
- Debate results feed back into phase outputs as refinements
- Configuration: enable/disable per phase, set debate round count

### 4.2 3-Pillar Fusion

**Video: SpecKit + Superpowers + GSD in Action** (10 min)

- SpecKit produces the specification and plan
- Superpowers validates with TDD subagents at each milestone
- GSD tracks progress against milestones and enforces delivery gates
- Fusion orchestrator sequences the three pillars per phase

### Hands-On Lab 4

Run a specification phase with DebateFunc injection enabled:

1. Submit a feature specification to HelixSpecifier
2. Observe multi-LLM debate logs during the Specify phase
3. Compare output with and without debate injection
4. Review the debate provenance in the specification output

---

## Module 5: Complete Spec-to-Code Walkthrough (30 min)

### 5.1 Scenario Definition

**Video: Planning the Walkthrough** (5 min)

- Scenario: Add a new health check aggregator endpoint
- Complexity: Whole Functionality (level 4)
- Expected ceremony: Standard (all phases, moderate depth)

### 5.2 Live Walkthrough

**Video: End-to-End Execution** (20 min)

- Phase 1: Constitution check confirms health endpoint requirements
- Phase 2: Specification defines aggregation logic and response format
- Phase 3: Clarification resolves timeout behavior for slow backends
- Phase 4: Plan decomposes into handler, service, and test components
- Phase 5: Tasks ordered with dependency graph
- Phase 6: Analysis identifies existing health check code to extend
- Phase 7: Implementation with TDD validation at each task

### 5.3 Results Review

**Video: Reviewing Outputs** (5 min)

- Specification document with traceability to implementation
- Test coverage report from Superpowers TDD validation
- GSD milestone completion summary
- Phase cache contents for future reference

### Hands-On Lab 5

Execute the complete walkthrough yourself:

1. Choose a feature for your project
2. Run it through all 7 SDD phases
3. Review phase caches in `.speckit/cache/`
4. Verify implementation passes all generated tests
5. Check GSD milestone completion status

---

## Course Summary

### Key Takeaways

1. HelixSpecifier implements Spec-Driven Development through a 3-pillar fusion of SpecKit, Superpowers, and GSD
2. The 7-phase SDD flow (Constitution, Specify, Clarify, Plan, Tasks, Analyze, Implement) provides structured development guidance
3. Work granularity detection with 5 levels enables auto-activation for significant changes
4. Adaptive ceremony scaling adjusts process overhead based on change complexity
5. DebateFunc injection brings real multi-LLM debate into specification phases
6. Phase caching enables workflow resumption after interruption

### Assessment Questions

1. Name the three pillars of HelixSpecifier and describe the role of each.
2. At which work granularity levels does SDD auto-activate, and why?
3. What is adaptive ceremony scaling and how does it affect the SDD workflow?
4. How does DebateFunc injection improve specification quality?
5. What happens when a phase cache becomes invalid due to Constitution changes?

### Related Courses

- Course 02: AI Debate System
- Course 12: Advanced Workflows
- Course 40: Module Development
- Course 53: HelixMemory Deep Dive

---

**Course Version**: 1.0
**Last Updated**: March 8, 2026
