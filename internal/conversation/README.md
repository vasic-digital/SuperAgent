# conversation

Package `conversation` provides conversation management with LLM-based context compression, event sourcing, and infinite context support. It enables HelixAgent to handle arbitrarily long conversations by compressing older messages while preserving entity relationships and key context.

## Architecture

The package is built around three core components:

- **`ContextCompressor`** -- LLM-powered conversation compression with four strategies.
- **`InfiniteContextEngine`** -- Kafka-backed conversation replay and snapshot management.
- **Event sourcing** -- Append-only event log for conversation state reconstruction.

### Key Types

- **`ContextCompressor`** -- Compresses conversation history using an `LLMClient` for summarization.
- **`LLMClient`** -- Interface for LLM providers: `Complete(ctx, prompt, maxTokens) (string, int, error)`.
- **`CompressionStrategy`** -- Enum: `window_summary`, `entity_graph`, `full`, `hybrid`.
- **`CompressionConfig`** -- Window size, target ratio, entity/topic preservation flags.
- **`CompressionData`** -- Result metadata: original/compressed counts, ratio, preserved entities, duration.
- **`MessageData`** -- Message structure with role, content, tokens, timestamps.
- **`EntityData`** -- Extracted entity with name, value, type, confidence.
- **`InfiniteContextEngine`** -- Manages conversation snapshots via Kafka message broker.

### Compression Strategies

| Strategy         | Behavior                                                      |
|------------------|---------------------------------------------------------------|
| `window_summary` | Summarizes older message windows, keeps recent 25% intact     |
| `entity_graph`   | Preserves entity-rich messages, summarizes non-entity content  |
| `full`           | Creates a single comprehensive summary plus recent 20% of messages |
| `hybrid`         | Cascades: entity_graph -> window_summary -> full until within token limit |

## Public API

```go
// Construction
NewContextCompressor(llmClient LLMClient, logger *logrus.Logger) *ContextCompressor

// Compression
Compress(ctx context.Context, messages []MessageData, entities []EntityData, maxTokens int) ([]MessageData, *CompressionData, error)

// Configuration
DefaultCompressionConfig() *CompressionConfig

// Infinite context engine
NewInfiniteContextEngine(broker messaging.MessageBroker, compressor *ContextCompressor, logger *logrus.Logger) *InfiniteContextEngine
GetConversationSnapshot(ctx context.Context, conversationID string) (*ConversationSnapshot, error)
```

## Usage

```go
compressor := conversation.NewContextCompressor(llmClient, logger)

compressed, stats, err := compressor.Compress(ctx, messages, entities, 4000)
// stats.CompressionRatio => 0.3 (compressed to 30% of original)
// stats.PreservedEntities => ["user_name", "project_id"]
```

The hybrid strategy (default) automatically escalates compression aggressiveness until the result fits within `maxTokens`. If no `LLMClient` is provided, a fallback produces placeholder summaries.

## Testing

```bash
go test -v ./internal/conversation/
go test -v -run TestContextCompressor ./internal/conversation/
go test -v -run TestInfiniteContext ./internal/conversation/
go test -v -run TestEventSourcing ./internal/conversation/
```

Unit tests use mock LLM clients. Integration tests require Kafka via `make test-infra-start`.
