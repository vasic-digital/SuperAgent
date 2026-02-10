# 🎯 SpecKit Auto-Activation User Guide

## 📋 Overview

SpecKit is HelixAgent's intelligent development workflow orchestrator that automatically triggers a comprehensive 7-phase development flow for large changes and refactoring tasks. It ensures proper planning, specification, and validation before implementation begins.

**Key Benefits:**
- ✅ Prevents premature implementation of complex features
- ✅ Ensures comprehensive planning and specification
- ✅ Reduces errors through multi-phase validation
- ✅ Provides structured workflow for large-scale changes
- ✅ Auto-detects work granularity and triggers appropriately

---

## 🔄 The 7-Phase Development Flow

SpecKit orchestrates development through seven sequential phases:

### Phase 1: Constitution 📜
**Purpose**: Establish project rules and constraints

- Reviews project Constitution (CLAUDE.md, AGENTS.md, CONSTITUTION.md)
- Identifies applicable mandatory principles
- Validates requirements against Constitution
- Ensures compliance with project standards

**Triggers**:
- Large functionality additions
- Whole system refactoring
- Architecture changes

**Duration**: 2-5 minutes

---

### Phase 2: Specify 📝
**Purpose**: Create detailed specifications

- Generates comprehensive specification documents
- Defines success criteria and acceptance tests
- Documents interfaces and contracts
- Creates feature specifications with examples

**Output**:
- `.speckit/specs/specification_YYYYMMDD_HHMMSS.md`
- Technical requirements document
- API contracts and schemas

**Duration**: 5-10 minutes

---

### Phase 3: Clarify ❓
**Purpose**: Resolve ambiguities and gather missing information

- Identifies unclear requirements
- Asks targeted clarification questions
- Validates assumptions with user
- Resolves conflicts in specifications

**Interaction**:
- May pause for user input via `AskUserQuestion`
- Presents multiple-choice options for decisions
- Allows custom text input for complex answers

**Duration**: Variable (depends on user response time)

---

### Phase 4: Plan 📋
**Purpose**: Design implementation strategy

- Creates step-by-step implementation plan
- Identifies affected files and modules
- Plans test strategy
- Defines rollback procedures

**Output**:
- `.speckit/plans/plan_YYYYMMDD_HHMMSS.md`
- Task breakdown with dependencies
- Risk assessment

**Duration**: 5-8 minutes

---

### Phase 5: Tasks ✅
**Purpose**: Break down into actionable tasks

- Creates granular task list
- Establishes task dependencies
- Assigns priorities
- Sets up progress tracking

**Output**:
- Task list in HelixAgent's task system
- Dependency graph
- Progress tracking dashboard

**Duration**: 2-3 minutes

---

### Phase 6: Analyze 🔍
**Purpose**: Deep code analysis and validation

- Analyzes existing codebase
- Identifies potential conflicts
- Validates architectural fit
- Performs impact analysis

**Tools Used**:
- Static code analysis
- Dependency analysis
- Test coverage analysis
- Architecture validation

**Duration**: 5-10 minutes

---

### Phase 7: Implement 🛠️
**Purpose**: Execute implementation with validation

- Implements changes according to plan
- Runs tests after each significant change
- Validates against specification
- Performs incremental commits

**Validation**:
- Unit tests
- Integration tests
- Specification compliance checks
- Constitution compliance

**Duration**: Variable (depends on complexity)

---

## 🎚️ Work Granularity Detection

SpecKit automatically detects the complexity of requested work and decides whether to activate:

### Granularity Levels

| Level | Description | SpecKit Activation | Example |
|-------|-------------|-------------------|---------|
| **1. Single Action** | One-off simple task | ❌ No | Fix typo, add log statement |
| **2. Small Creation** | Single file/function | ❌ No | Add utility function, create helper |
| **3. Big Creation** | Multiple files, new feature | ✅ Yes | New API endpoint with tests |
| **4. Whole Functionality** | Complete feature/system | ✅ Yes | Authentication system, payment integration |
| **5. Refactoring** | Large-scale restructuring | ✅ Yes | Extract module, redesign architecture |

### Detection Criteria

**Triggers SpecKit (Levels 3-5)**:
- Multiple files affected (≥3 files)
- New modules or packages created
- Architecture changes required
- Breaking changes to public APIs
- Cross-cutting concerns (security, performance, scalability)
- Complex business logic implementation
- Multi-step migrations or transformations

**Skips SpecKit (Levels 1-2)**:
- Single file edits
- Simple bug fixes
- Documentation updates
- Configuration changes
- Code formatting
- Minor refactoring (rename, extract variable)

---

## ⚙️ Configuration

### Environment Variables

```bash
# Enable/disable SpecKit auto-activation (default: enabled)
SPECKIT_AUTO_ACTIVATION=true

# Minimum granularity level to trigger SpecKit (default: 3)
# 1=SingleAction, 2=SmallCreation, 3=BigCreation, 4=WholeFunctionality, 5=Refactoring
SPECKIT_MIN_GRANULARITY=3

# Enable phase caching for resumption (default: enabled)
SPECKIT_PHASE_CACHING=true

# Cache directory (default: .speckit/cache/)
SPECKIT_CACHE_DIR=.speckit/cache

# Timeout for each phase in minutes (default: 30)
SPECKIT_PHASE_TIMEOUT=30

# Enable verbose logging (default: false)
SPECKIT_VERBOSE=false
```

### Configuration File

Create `.speckit/config.yaml`:

```yaml
speckit:
  auto_activation:
    enabled: true
    min_granularity: 3  # BigCreation and above

  phases:
    constitution:
      enabled: true
      timeout_minutes: 5
    specify:
      enabled: true
      timeout_minutes: 10
      output_dir: .speckit/specs
    clarify:
      enabled: true
      max_questions: 5
      allow_skip: false
    plan:
      enabled: true
      timeout_minutes: 10
      output_dir: .speckit/plans
    tasks:
      enabled: true
      create_task_list: true
    analyze:
      enabled: true
      timeout_minutes: 15
      tools:
        - static_analysis
        - dependency_check
        - test_coverage
    implement:
      enabled: true
      incremental_commits: true
      run_tests: true

  caching:
    enabled: true
    cache_dir: .speckit/cache
    ttl_hours: 24

  logging:
    level: info  # debug, info, warn, error
    file: .speckit/logs/speckit.log
    console: true
```

---

## 🚀 Usage Examples

### Example 1: Adding Authentication System

**User Request:**
> "Add JWT-based authentication to the API with role-based access control"

**SpecKit Activation:**
```
🎯 SpecKit Auto-Activation Triggered
Granularity: WholeFunctionality (Level 4)
Reason: New authentication system requires multiple components

Phase 1/7: Constitution - Reviewing security requirements...
Phase 2/7: Specify - Creating authentication specification...
Phase 3/7: Clarify - Asking clarification questions...
  ❓ Which JWT library should we use?
    1. golang-jwt/jwt (Recommended)
    2. dgrijalva/jwt-go
    3. Custom implementation
  ❓ Where should user credentials be stored?
    1. PostgreSQL with bcrypt (Recommended)
    2. External auth service (OAuth)
Phase 4/7: Plan - Creating implementation plan...
Phase 5/7: Tasks - Breaking down into tasks...
  ✅ Task #1: Create JWT middleware
  ✅ Task #2: Add user authentication table
  ✅ Task #3: Implement login endpoint
  ✅ Task #4: Add role-based authorization
  ✅ Task #5: Write tests
Phase 6/7: Analyze - Analyzing codebase...
Phase 7/7: Implement - Implementing authentication...
```

**Output Files:**
- `.speckit/specs/authentication_spec_20260210_143022.md`
- `.speckit/plans/authentication_plan_20260210_143530.md`
- `.speckit/cache/phase_3_clarify.json` (for resumption)

---

### Example 2: Refactoring Provider System

**User Request:**
> "Refactor the LLM provider system to use a plugin architecture"

**SpecKit Activation:**
```
🎯 SpecKit Auto-Activation Triggered
Granularity: Refactoring (Level 5)
Reason: Large-scale architectural refactoring

Phase 1/7: Constitution - Checking architecture principles...
  ⚠️ Constitution Rule: "No broken components allowed"
  ⚠️ Constitution Rule: "100% test coverage required"
Phase 2/7: Specify - Defining plugin interface...
Phase 3/7: Clarify - Validating approach...
Phase 4/7: Plan - Creating migration plan...
  📋 Plan includes:
    - Backward compatibility layer
    - Incremental migration strategy
    - Rollback procedure
Phase 5/7: Tasks - Creating task breakdown...
Phase 6/7: Analyze - Impact analysis...
  🔍 Affected files: 47 files
  🔍 Affected tests: 156 tests
  🔍 Breaking changes: 3 public APIs
Phase 7/7: Implement - Executing refactoring...
```

---

### Example 3: Simple Bug Fix (No SpecKit)

**User Request:**
> "Fix the typo in the error message on line 42 of handlers.go"

**SpecKit Decision:**
```
ℹ️ SpecKit Auto-Activation Skipped
Granularity: SingleAction (Level 1)
Reason: Simple single-file change

Proceeding with direct implementation...
```

---

## 🔄 Phase Caching & Resumption

SpecKit caches phase results to allow resumption if interrupted:

### Cache Structure

```
.speckit/
├── cache/
│   ├── phase_1_constitution.json
│   ├── phase_2_specify.json
│   ├── phase_3_clarify.json
│   ├── phase_4_plan.json
│   ├── phase_5_tasks.json
│   ├── phase_6_analyze.json
│   └── session_metadata.json
├── specs/
│   └── specification_20260210_143022.md
├── plans/
│   └── plan_20260210_143530.md
└── logs/
    └── speckit.log
```

### Resuming from Interruption

```bash
# SpecKit automatically detects cached phases
# and resumes from last completed phase

# Manual resumption
helixagent speckit resume --session-id <session-id>

# Clear cache and restart
helixagent speckit clear-cache
```

---

## 📊 Monitoring & Observability

### Progress Tracking

```bash
# Check current SpecKit status
curl http://localhost:7061/v1/speckit/status

# Response:
{
  "active": true,
  "session_id": "speckit_20260210_143022",
  "current_phase": "implement",
  "phase_number": 7,
  "total_phases": 7,
  "progress": 85,
  "started_at": "2026-02-10T14:30:22Z",
  "estimated_completion": "2026-02-10T15:45:00Z",
  "phases_completed": [
    "constitution",
    "specify",
    "clarify",
    "plan",
    "tasks",
    "analyze"
  ]
}
```

### Metrics

```bash
# Prometheus metrics
speckit_phase_duration_seconds{phase="constitution"} 180.5
speckit_phase_duration_seconds{phase="specify"} 420.2
speckit_activation_total{granularity="whole_functionality"} 47
speckit_cache_hit_ratio 0.73
```

---

## ❓ Troubleshooting

### SpecKit Not Activating

**Problem**: SpecKit doesn't activate for complex tasks

**Solutions**:
1. Check granularity detection:
   ```bash
   export SPECKIT_VERBOSE=true
   # Review logs for granularity decision
   ```

2. Lower activation threshold:
   ```bash
   export SPECKIT_MIN_GRANULARITY=2  # Activate for SmallCreation
   ```

3. Force activation:
   ```bash
   helixagent speckit force-activate
   ```

---

### Phase Timeout

**Problem**: Phase exceeds timeout limit

**Solutions**:
1. Increase timeout:
   ```bash
   export SPECKIT_PHASE_TIMEOUT=60  # 60 minutes
   ```

2. Resume from cache:
   ```bash
   helixagent speckit resume
   ```

---

### Cache Corruption

**Problem**: Cached phase data is corrupted

**Solutions**:
1. Clear cache and restart:
   ```bash
   rm -rf .speckit/cache/*
   helixagent speckit restart
   ```

2. Disable caching temporarily:
   ```bash
   export SPECKIT_PHASE_CACHING=false
   ```

---

## 🎓 Best Practices

### 1. Let SpecKit Decide
- Trust granularity detection for most cases
- Only override for exceptional situations
- Review activation logs to understand decisions

### 2. Answer Clarification Questions Thoroughly
- Provide detailed context in clarify phase
- Use custom text input for nuanced answers
- Don't skip clarification questions

### 3. Review Specifications Before Implementation
- Check generated specs in `.speckit/specs/`
- Validate against requirements
- Request regeneration if needed

### 4. Monitor Progress
- Use `/v1/speckit/status` endpoint
- Set up alerting for long-running phases
- Review logs regularly

### 5. Maintain Cache Hygiene
- Clear old caches periodically
- Archive important specifications
- Don't rely on cache for >24 hours

---

## 📚 Related Documentation

- **[CLAUDE.md](../../CLAUDE.md)** - SpecKit implementation details
- **[AGENTS.md](../../AGENTS.md)** - SpecKit agent configuration
- **[Constitution](../../CONSTITUTION.md)** - Project rules and constraints
- **[API Reference](../api/API_REFERENCE.md)** - SpecKit API endpoints

---

## 🔗 Quick Reference

### Commands

```bash
# Check SpecKit status
curl http://localhost:7061/v1/speckit/status

# Force activation
helixagent speckit force-activate

# Resume from cache
helixagent speckit resume

# Clear cache
helixagent speckit clear-cache

# View logs
tail -f .speckit/logs/speckit.log
```

### Key Files

- Configuration: `.speckit/config.yaml`
- Cache: `.speckit/cache/`
- Specifications: `.speckit/specs/`
- Plans: `.speckit/plans/`
- Logs: `.speckit/logs/`

---

**Last Updated**: February 10, 2026
**Version**: 1.0.0
**Status**: ✅ Production Ready
