# debate/reflexion - Reflexion Framework

Implements episodic memory, verbal reflection, retry-and-learn loops, and accumulated wisdom for cross-session debate learning.

## Purpose

The reflexion package enables the debate system to learn from past experiences by maintaining episodic memory of debate outcomes, generating verbal reflections on failures, and accumulating wisdom that persists across debate sessions.

## Key Components

### EpisodicMemory

Short-term memory buffer storing recent debate experiences for reflection.

```go
memory := reflexion.NewEpisodicMemory(bufferSize, logger)
memory.Store(ctx, episode)
recent := memory.GetRecent(ctx, count)
similar := memory.SearchSimilar(ctx, query, topK)
```

### ReflectionGenerator

Generates verbal reflections analyzing why a debate outcome succeeded or failed.

```go
generator := reflexion.NewReflectionGenerator(llmProvider, logger)
reflection, err := generator.Reflect(ctx, episode, outcome)
// Returns: "The debate failed because participants focused on surface-level
// arguments without addressing the core algorithmic complexity..."
```

### ReflexionLoop

Retry-and-learn loop that uses reflections to improve subsequent attempts.

```go
loop := reflexion.NewReflexionLoop(config, generator, memory, logger)
result, err := loop.Run(ctx, task, maxAttempts)
```

The loop:
1. Attempts the task
2. If it fails, generates a reflection on why
3. Stores the reflection in episodic memory
4. Retries the task with the reflection as additional context
5. Repeats until success or max attempts reached

### AccumulatedWisdom

Long-term knowledge base that persists lessons learned across debate sessions.

```go
wisdom := reflexion.NewAccumulatedWisdom(logger)
wisdom.AddLesson(ctx, lesson)
relevant := wisdom.GetRelevantLessons(ctx, topic, topK)
summary := wisdom.Summarize(ctx)
```

## Key Types

- **Episode** -- A recorded debate experience with context, actions, outcome, and metadata
- **Reflection** -- Verbal analysis of an episode identifying what went wrong/right and why
- **Lesson** -- Distilled wisdom from multiple reflections on similar topics
- **WisdomEntry** -- A lesson with confidence score, usage count, and timestamp

## Usage within Debate System

The reflexion framework is integrated into the debate orchestrator's convergence phase. When debates fail to reach consensus, the reflexion loop generates insights that are fed into subsequent debate rounds. Accumulated wisdom from past sessions is injected as additional context for new debates on similar topics.

## Files

- `episodic_memory.go` -- EpisodicMemory buffer with search
- `episodic_memory_test.go` -- Unit tests
- `reflection_generator.go` -- LLM-based reflection generation
- `reflection_generator_test.go` -- Unit tests
- `reflexion_loop.go` -- Retry-and-learn loop orchestrator
- `reflexion_loop_test.go` -- Unit tests
- `accumulated_wisdom.go` -- Cross-session wisdom persistence
- `accumulated_wisdom_test.go` -- Unit tests
