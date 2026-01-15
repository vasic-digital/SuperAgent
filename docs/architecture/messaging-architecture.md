# Messaging Architecture

HelixAgent uses a hybrid messaging pattern combining RabbitMQ for task queuing and Apache Kafka for event streaming.

## Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        HelixAgent                                │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────────┐         ┌──────────────────┐              │
│  │  Task Queue      │         │  Event Stream    │              │
│  │  (RabbitMQ)      │         │  (Kafka)         │              │
│  │                  │         │                  │              │
│  │  - Background    │         │  - LLM Responses │              │
│  │  - LLM Tasks     │         │  - Debate Rounds │              │
│  │  - Debates       │         │  - Verification  │              │
│  │  - Verification  │         │  - Audit Logs    │              │
│  │  - Notifications │         │  - Metrics       │              │
│  └────────┬─────────┘         └────────┬─────────┘              │
│           │                            │                         │
│           └──────────┬─────────────────┘                         │
│                      │                                           │
│           ┌──────────▼──────────┐                               │
│           │  Messaging Hub      │                               │
│           │  (Abstraction Layer)│                               │
│           └──────────┬──────────┘                               │
│                      │                                           │
│           ┌──────────▼──────────┐                               │
│           │  In-Memory Fallback │                               │
│           │  (Testing/Dev)      │                               │
│           └─────────────────────┘                               │
└─────────────────────────────────────────────────────────────────┘
```

## Message Broker Abstraction

All message brokers implement the `MessageBroker` interface:

```go
type MessageBroker interface {
    Connect(ctx context.Context) error
    Close(ctx context.Context) error
    HealthCheck(ctx context.Context) error
    IsConnected() bool
    Publish(ctx context.Context, topic string, message *Message, opts ...PublishOption) error
    PublishBatch(ctx context.Context, topic string, messages []*Message, opts ...PublishOption) error
    Subscribe(ctx context.Context, topic string, handler MessageHandler, opts ...SubscribeOption) (Subscription, error)
    BrokerType() BrokerType
    GetMetrics() *BrokerMetrics
}
```

## Broker Implementations

### RabbitMQ (Task Queuing)

**Use Cases:**
- Background task execution
- LLM request queuing
- Debate round coordination
- Verification tasks
- Real-time notifications

**Key Features:**
- Publisher confirms for guaranteed delivery
- Dead letter queues for failed messages
- Priority queues for task prioritization
- Automatic reconnection
- Prefetch control for throughput management

**Configuration (`configs/messaging.yaml`):**
```yaml
rabbitmq:
  host: ${RABBITMQ_HOST:-localhost}
  port: ${RABBITMQ_PORT:-5672}
  username: ${RABBITMQ_USER:-helixagent}
  password: ${RABBITMQ_PASSWORD:-helixagent123}
  publisher_confirm: true
  prefetch_count: 10
```

### Apache Kafka (Event Streaming)

**Use Cases:**
- LLM response streaming
- Debate round events
- Verification results
- Audit logging
- Metrics collection
- Token streaming

**Key Features:**
- High-throughput event streaming
- Consumer groups for horizontal scaling
- Message compression (LZ4, Snappy, Zstd)
- Exactly-once semantics (optional)
- Event replay capability

**Configuration (`configs/messaging.yaml`):**
```yaml
kafka:
  brokers:
    - ${KAFKA_BROKER:-localhost:9092}
  client_id: ${KAFKA_CLIENT_ID:-helixagent}
  group_id: ${KAFKA_GROUP_ID:-helixagent-group}
  compression: lz4
  required_acks: all
```

### In-Memory (Testing/Fallback)

**Use Cases:**
- Unit testing
- Integration testing
- Development without infrastructure
- Fallback when brokers unavailable

## Queue Definitions

### RabbitMQ Queues

| Queue | Purpose | Priority |
|-------|---------|----------|
| `helixagent.tasks.background` | Background jobs | Normal |
| `helixagent.tasks.llm` | LLM API requests | High (0-10) |
| `helixagent.tasks.debate` | Debate rounds | Normal |
| `helixagent.tasks.verification` | Provider verification | Normal |
| `helixagent.tasks.notifications` | User notifications | Normal |
| `helixagent.dlq` | Dead letter queue | N/A |
| `helixagent.retry` | Retry queue (60s TTL) | N/A |

### Kafka Topics

| Topic | Partitions | Purpose |
|-------|------------|---------|
| `helixagent.events.llm.responses` | 6 | LLM response events |
| `helixagent.events.debate.rounds` | 6 | Debate round events |
| `helixagent.events.verification.results` | 3 | Verification results |
| `helixagent.events.provider.health` | 3 | Provider health updates |
| `helixagent.events.audit` | 6 | Audit logs (30-day retention) |
| `helixagent.events.metrics` | 3 | System metrics |
| `helixagent.events.errors` | 3 | Error events |
| `helixagent.stream.tokens` | 12 | Token streaming |
| `helixagent.stream.sse` | 6 | SSE events |
| `helixagent.stream.websocket` | 6 | WebSocket messages |

## Message Structure

```go
type Message struct {
    ID            string            // Unique identifier
    Type          string            // Message type for routing
    Payload       []byte            // Message content
    Headers       map[string]string // Metadata
    Timestamp     time.Time         // Creation time
    Priority      MessagePriority   // 1-10 (RabbitMQ)
    RetryCount    int               // Current retry attempt
    MaxRetries    int               // Maximum retries
    TraceID       string            // Distributed tracing
    CorrelationID string            // Related message linking
    Expiration    time.Time         // TTL
    DeliveryMode  DeliveryMode      // Persistent/Transient
}
```

## Error Handling

### Retry Strategy

1. Messages are retried up to `MaxRetries` times
2. Failed messages after max retries go to dead letter queue
3. Retry queue has 60-second TTL before re-delivery

### Error Types

| Error Code | Description | Retryable |
|------------|-------------|-----------|
| `CONNECTION_FAILED` | Broker connection lost | Yes |
| `PUBLISH_FAILED` | Message publish failed | Yes |
| `SUBSCRIBE_FAILED` | Subscription failed | Yes |
| `CONSUME_FAILED` | Message consumption failed | Yes |
| `ACK_FAILED` | Acknowledgment failed | Yes |
| `TIMEOUT` | Operation timed out | Yes |
| `INVALID_CONFIG` | Invalid configuration | No |
| `SERIALIZATION_ERROR` | Message serialization failed | No |

## Metrics

The messaging system collects comprehensive metrics:

```go
type BrokerMetrics struct {
    // Connection metrics
    ConnectionAttempts   int64
    ConnectionSuccesses  int64
    ConnectionFailures   int64

    // Publish metrics
    MessagesPublished    int64
    PublishSuccesses     int64
    PublishFailures      int64
    BytesPublished       int64

    // Consume metrics
    MessagesReceived     int64
    MessagesProcessed    int64
    MessagesFailed       int64
    MessagesRetried      int64
    MessagesDeadLettered int64

    // Latency metrics
    PublishLatencyTotal  int64  // nanoseconds
    PublishLatencyCount  int64
}
```

## Docker Deployment

Start messaging infrastructure:

```bash
# Start all messaging services
docker-compose -f docker-compose.messaging.yml --profile messaging up -d

# Start with Kafka UI (debugging)
docker-compose -f docker-compose.messaging.yml --profile messaging-ui up -d

# Start full stack
docker-compose -f docker-compose.messaging.yml --profile full up -d
```

### Services

| Service | Port | Description |
|---------|------|-------------|
| RabbitMQ | 5672 | AMQP protocol |
| RabbitMQ Management | 15672 | Web UI |
| Zookeeper | 2181 | Kafka coordination |
| Kafka | 9092 | Event streaming |
| Schema Registry | 8081 | Avro schemas |
| Kafka UI | 8080 | Web UI |

## GraphQL Integration

The messaging system integrates with GraphQL for real-time subscriptions:

```graphql
type Subscription {
    debateUpdates(debateId: ID!): DebateUpdate!
    taskProgress(taskId: ID!): TaskProgress!
    providerHealth: ProviderHealthUpdate!
    tokenStream(requestId: ID!): TokenStreamEvent!
}
```

## TOON Protocol

Token-Optimized Object Notation (TOON) reduces token usage for AI communication:

**Key Compression:**
- `id` → `i`
- `name` → `n`
- `status` → `s`
- `created_at` → `ca`

**Value Compression:**
- `healthy` → `H`
- `pending` → `P`
- `completed` → `C`
- `failed` → `F`

**Compression Levels:**
- `None` - No compression
- `Minimal` - Key compression only
- `Standard` - Key + value compression
- `Aggressive` - Full compression + gzip

## LLMsVerifier Integration

The LLMsVerifier publishes events to Kafka topics:

| Event Type | Description |
|------------|-------------|
| `verification.started` | Verification process started |
| `verification.completed` | Verification completed |
| `verification.failed` | Verification failed |
| `provider.discovered` | New provider discovered |
| `provider.scored` | Provider score calculated |
| `model.ranked` | Model ranking updated |
| `team.selected` | AI debate team selected |

## Testing

Run messaging tests:

```bash
# Unit tests
go test ./internal/messaging/...

# Integration tests
go test ./tests/integration/messaging_integration_test.go

# Challenge scripts
./challenges/scripts/messaging_hybrid_challenge.sh
./challenges/scripts/messaging_rabbitmq_challenge.sh
./challenges/scripts/messaging_kafka_challenge.sh
```

## Performance Targets

| Metric | Target |
|--------|--------|
| Publish latency | < 10ms (p99) |
| Message throughput | 100K+ msg/sec |
| Connection recovery | < 5 seconds |
| Dead letter rate | < 0.1% |
