# learning

Package `learning` implements cross-session learning for HelixAgent. It consumes completed conversation events from Kafka, extracts behavioral patterns (user intent, debate strategy, entity co-occurrence, user preferences, conversation flow), and generates actionable insights that are published back to Kafka for downstream consumers.

## Architecture

The system follows a consume-process-publish pipeline:

1. `CrossSessionLearner` subscribes to a Kafka topic for completed conversations.
2. Each completion is analyzed to extract multiple `Pattern` types.
3. Patterns are stored and their frequencies updated in the `InsightStore`.
4. High-frequency, high-confidence patterns generate `Insight` objects published to an insights topic.

### Key Types

- **`CrossSessionLearner`** -- Main orchestrator: subscribes to completions, extracts patterns, publishes insights.
- **`InsightStore`** -- In-memory store for patterns and insights with frequency tracking and type-based queries.
- **`Pattern`** -- A learned pattern with type, frequency, confidence, examples, and time range.
- **`Insight`** -- A derived insight with impact level, associated patterns, and user scope.
- **`ConversationCompletion`** -- Input event containing messages, entities, debate rounds, and outcome.
- **`CrossSessionConfig`** -- Kafka topics and thresholds for pattern extraction.

### Pattern Types

| Type                     | Extraction Logic                                      |
|--------------------------|-------------------------------------------------------|
| `user_intent`            | Keyword analysis of first user message (help, explain, fix, create) |
| `debate_strategy`        | Best-performing provider by cumulative confidence      |
| `entity_cooccurrence`    | Most frequent entity-type pair co-occurrences          |
| `user_preference`        | Communication style classification (concise/moderate/detailed) |
| `conversation_flow`      | Pace classification based on average time per message   |
| `provider_performance`   | Provider performance tracking across sessions           |

### Insight Generation

Insights are generated when a pattern reaches both minimum frequency (default: 3) and minimum confidence (default: 0.7). Impact is classified as:
- **high**: frequency >= 10, confidence >= 0.9
- **medium**: frequency >= 5, confidence >= 0.7
- **low**: all others

## Public API

```go
// Construction
NewCrossSessionLearner(config CrossSessionConfig, broker messaging.MessageBroker, logger *logrus.Logger) *CrossSessionLearner

// Lifecycle
StartLearning(ctx context.Context) error

// Query
GetInsights(limit int) []Insight
GetPatterns(patternType string) []Pattern  // "all" returns all types
GetLearningStats() map[string]interface{}

// InsightStore
NewInsightStore(logger *logrus.Logger) *InsightStore
UpdatePattern(pattern Pattern)
GetPattern(patternID string) *Pattern
GetPatternsByType(patternType PatternType) []Pattern
GetTopPatterns(limit int) []Pattern
AddInsight(insight Insight)
GetInsight(insightID string) *Insight
GetInsightsByUser(userID string) []Insight
GetStats() map[string]interface{}
```

## Configuration

```go
config := learning.CrossSessionConfig{
    CompletedTopic: "helixagent.conversations.completed",
    InsightsTopic:  "helixagent.learning.insights",
    MinConfidence:  0.7,
    MinFrequency:   3,
}
```

## Usage

```go
learner := learning.NewCrossSessionLearner(config, kafkaBroker, logger)
err := learner.StartLearning(ctx)

// Query learned patterns
patterns := learner.GetPatterns("user_intent")
insights := learner.GetInsights(10)
stats := learner.GetLearningStats()
```

Insights and patterns are also exposed via the bigdata HTTP handler at `GET /v1/learning/insights` and `GET /v1/learning/patterns`.

## Testing

```bash
go test -v ./internal/learning/
go test -v -run TestCrossSession ./internal/learning/
```

Integration tests require a running Kafka broker. Use `make test-infra-start` to provision one.
