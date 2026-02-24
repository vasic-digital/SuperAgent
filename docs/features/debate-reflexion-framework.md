# Debate Reflexion Framework

## Overview

The Reflexion framework implements **verbal reinforcement learning** (verbal RL) within the HelixAgent debate system. Rather than relying on scalar reward signals as in traditional RL, Reflexion uses natural-language self-reflection to guide iterative code improvement. An agent generates code, runs tests, reflects on failures, and feeds those reflections into the next attempt -- creating a self-correcting loop that accumulates transferable knowledge over time.

The implementation lives in `internal/debate/reflexion/` and consists of four core components: the **Episodic Memory Buffer**, the **Reflection Generator**, the **Reflexion Loop**, and the **Accumulated Wisdom** store.

## Episodic Memory Buffer

**File:** `internal/debate/reflexion/episodic_memory.go`

The `EpisodicMemoryBuffer` is a bounded, thread-safe FIFO store of `Episode` records. Each episode captures a single attempt at solving a task:

| Field             | Type                     | Description                                  |
|-------------------|--------------------------|----------------------------------------------|
| `ID`              | `string`                 | Unique episode identifier (auto-generated)   |
| `SessionID`       | `string`                 | Parent debate session                        |
| `TurnID`          | `string`                 | Specific debate turn                         |
| `AgentID`         | `string`                 | Agent that produced the attempt (required)   |
| `TaskDescription` | `string`                 | What the agent was trying to do              |
| `AttemptNumber`   | `int`                    | Sequential attempt counter (1-based)         |
| `Code`            | `string`                 | Source code of the attempt                   |
| `TestResults`     | `map[string]interface{}` | Structured test outcomes                     |
| `FailureAnalysis` | `string`                 | Human-readable error summary                 |
| `Reflection`      | `*Reflection`            | Self-assessment generated after the attempt  |
| `Improvement`     | `string`                 | Proposed fix extracted from reflection       |
| `Confidence`      | `float64`                | Agent confidence in the fix (0-1)            |
| `Timestamp`       | `time.Time`              | When the episode was recorded                |

### FIFO eviction

The buffer enforces a configurable maximum size (default: 1000 episodes). When the buffer is at capacity, `Store()` evicts the oldest episode before inserting the new one. Eviction also cleans the by-agent and by-session indexes to prevent memory leaks.

### Indexed lookups

Three retrieval strategies are available:

- **`GetByAgent(agentID)`** -- returns all episodes for a given agent, useful for agent-specific reflection history.
- **`GetBySession(sessionID)`** -- returns all episodes within a debate session.
- **`GetRelevant(taskDescription, limit)`** -- keyword overlap search across task descriptions, returning the most similar past episodes. This enables "recall from past experience" during new tasks.

### Persistence

The buffer implements `json.Marshaler` and `json.Unmarshaler`. On unmarshal, all indexes are rebuilt from the episode list, making it safe to serialize to the `reflections` JSONB column in the `debate_turns` table.

## Reflection Generator

**File:** `internal/debate/reflexion/reflection_generator.go`

The `ReflectionGenerator` accepts a `ReflectionRequest` (code, test results, error messages, prior reflections, task description, attempt number) and produces a structured `Reflection`:

```go
type Reflection struct {
    RootCause        string    // What caused the failure
    WhatWentWrong    string    // Observable symptoms
    WhatToChangeNext string    // Concrete fix suggestion
    ConfidenceInFix  float64   // How likely the fix will work (0-1)
    GeneratedAt      time.Time
}
```

### LLM-based generation

The generator sends a structured prompt to the LLM client and parses the response by looking for four labelled fields: `ROOT_CAUSE`, `WHAT_WENT_WRONG`, `WHAT_TO_CHANGE`, and `CONFIDENCE`. If any field is missing, the response is rejected.

### Deterministic fallback

When the LLM call fails or returns unparseable output, the generator activates a **pattern-matching fallback** that categorizes errors into 10 classes:

1. Compilation / syntax errors
2. Test assertion failures
3. Timeouts / deadlines
4. Nil pointer dereferences
5. Index out of bounds
6. Permission / authorization errors
7. Import / dependency issues
8. Type mismatch / conversion errors
9. Concurrency issues (deadlock, race)
10. Memory allocation failures

Each category produces a tailored root cause, description, fix suggestion, and confidence score. When prior reflections exist, the fallback notes that the previous suggestion was insufficient and reduces confidence by 10%.

## Reflexion Loop

**File:** `internal/debate/reflexion/reflexion_loop.go`

The `ReflexionLoop` ties everything together. It accepts a `ReflexionConfig`:

| Parameter             | Default   | Description                          |
|-----------------------|-----------|--------------------------------------|
| `MaxAttempts`         | 3         | Maximum retry iterations             |
| `ConfidenceThreshold` | 0.95      | Pass rate required for early exit    |
| `Timeout`             | 5 minutes | Overall time budget for the loop     |

### Algorithm

```
1. If no initial code is provided, generate it via CodeGenerator
2. FOR attempt = 1 to MaxAttempts:
   a. Check context for timeout / cancellation
   b. Execute tests against current code (TestExecutor)
   c. IF all tests pass AND confidence >= threshold:
      RETURN success
   d. Collect error messages from failed tests
   e. Generate a Reflection (LLM or fallback)
   f. Store episode in EpisodicMemoryBuffer
   g. Accumulate reflection
   h. IF at max attempts: RETURN with accumulated data
   i. Generate improved code via CodeGenerator(task, reflections)
   j. currentCode = improved code
3. RETURN result
```

The loop returns a `ReflexionResult` containing: the final code, number of attempts, whether all tests passed, all reflections and episodes, final confidence, total duration, and the last set of test results.

### Key interfaces

- **`TestExecutor`** -- executes tests against code and returns `[]*TestResult`.
- **`CodeGenerator`** -- a function type `func(ctx, task, priorReflections) (string, error)` that generates improved code given accumulated reflections.

## Accumulated Wisdom

**File:** `internal/debate/reflexion/accumulated_wisdom.go`

The `AccumulatedWisdom` store generalizes insights from individual episodes into reusable **Wisdom** entries that persist across sessions.

### Extraction algorithm

`ExtractFromEpisodes(episodes)`:

1. Group episodes by `Reflection.RootCause`.
2. For groups with 2+ episodes, create a `Wisdom` entry:
   - **Pattern** = the shared root cause string.
   - **Frequency** = group size.
   - **Impact** = average confidence improvement within sessions (later attempts vs earlier).
   - **Domain** = inferred from code content signals (code, testing, architecture, concurrency, performance).
   - **Tags** = keywords extracted from the pattern, excluding stop words.
   - **Source** = comma-separated episode IDs.
3. Store each wisdom entry in the in-memory index.

### Relevance retrieval

`GetRelevant(taskDescription, limit)` scores each wisdom entry against the query by checking keyword overlap across three dimensions: pattern text, tags, and domain. Results are sorted by score with impact as tiebreaker.

### Usage tracking

`RecordUsage(wisdomID, success)` updates a running success rate and use count, enabling the system to learn which wisdom entries actually help in practice.

### Domain index

Wisdom is additionally indexed by domain (`code`, `testing`, `architecture`, `concurrency`, `performance`), enabling `GetByDomain(domain)` lookups for domain-specific guidance.

## Configuration

Reflexion is configured per debate session via the `ReflexionConfig` struct. Example:

```go
config := reflexion.ReflexionConfig{
    MaxAttempts:         5,
    ConfidenceThreshold: 0.90,
    Timeout:             10 * time.Minute,
}
```

The episodic memory buffer size is set at construction:

```go
memory := reflexion.NewEpisodicMemoryBuffer(2000) // 2000 episodes max
```

## Integration with Debate Protocol

The Reflexion framework integrates with the debate protocol at two points:

1. **During debate turns** -- episodes are stored in the `reflections` JSONB column of the `debate_turns` table, enabling full replay.
2. **Cross-session learning** -- the `AccumulatedWisdom` store feeds into the knowledge repository (`internal/debate/knowledge/`), making patterns from past debates available to future ones.

## Related Files

- `internal/debate/reflexion/episodic_memory.go` -- Episodic memory buffer
- `internal/debate/reflexion/reflection_generator.go` -- Reflection generator
- `internal/debate/reflexion/reflexion_loop.go` -- Main loop
- `internal/debate/reflexion/accumulated_wisdom.go` -- Cross-session wisdom
- `internal/debate/knowledge/repository.go` -- Knowledge repository integration
- `sql/schema/debate_turns.sql` -- Database schema for reflections storage
