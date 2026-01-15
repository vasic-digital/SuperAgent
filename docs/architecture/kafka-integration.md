# Kafka Integration

HelixAgent uses Apache Kafka as its event streaming platform for high-throughput, durable event processing.

## Overview

Kafka handles asynchronous event streaming where persistence, replay capability, and horizontal scaling are critical:
- LLM response streaming
- Debate round events
- Verification results publishing
- Audit logging
- Metrics collection
- Token streaming

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        Apache Kafka Cluster                              │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                      Topics                                       │   │
│  │                                                                   │   │
│  │  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │   │
│  │  │ llm.responses   │  │ debate.rounds   │  │ verification    │ │   │
│  │  │  (6 partitions) │  │  (6 partitions) │  │  (3 partitions) │ │   │
│  │  └─────────────────┘  └─────────────────┘  └─────────────────┘ │   │
│  │                                                                   │   │
│  │  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │   │
│  │  │ audit           │  │ metrics         │  │ errors          │ │   │
│  │  │  (6 partitions) │  │  (3 partitions) │  │  (3 partitions) │ │   │
│  │  │  (30-day TTL)   │  │                 │  │                 │ │   │
│  │  └─────────────────┘  └─────────────────┘  └─────────────────┘ │   │
│  │                                                                   │   │
│  │  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │   │
│  │  │ stream.tokens   │  │ stream.sse      │  │ stream.websocket│ │   │
│  │  │ (12 partitions) │  │  (6 partitions) │  │  (6 partitions) │ │   │
│  │  └─────────────────┘  └─────────────────┘  └─────────────────┘ │   │
│  │                                                                   │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                          │
│  ┌─────────────────┐     ┌─────────────────┐                           │
│  │   Zookeeper     │     │ Schema Registry │                           │
│  │   (Metadata)    │     │    (Avro)       │                           │
│  └─────────────────┘     └─────────────────┘                           │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

## Configuration

### Environment Variables

```bash
KAFKA_BROKERS=localhost:9092
KAFKA_CLIENT_ID=helixagent
KAFKA_GROUP_ID=helixagent-group
KAFKA_COMPRESSION=lz4
KAFKA_REQUIRED_ACKS=all
KAFKA_MAX_WAIT_TIME=500ms
```

### Configuration File (`configs/messaging.yaml`)

```yaml
kafka:
  brokers:
    - ${KAFKA_BROKER:-localhost:9092}
  client_id: ${KAFKA_CLIENT_ID:-helixagent}
  group_id: ${KAFKA_GROUP_ID:-helixagent-group}

  # Producer settings
  compression: lz4  # none, gzip, snappy, lz4, zstd
  required_acks: all  # none, leader, all
  max_wait_time: 500ms
  batch_size: 16384
  linger_ms: 5

  # Consumer settings
  auto_offset_reset: earliest  # earliest, latest
  enable_auto_commit: false
  session_timeout: 30s
  heartbeat_interval: 3s
  max_poll_records: 500

  # TLS settings (production)
  tls:
    enabled: false
    cert_file: ""
    key_file: ""
    ca_file: ""

  # SASL settings (production)
  sasl:
    enabled: false
    mechanism: SCRAM-SHA-512
    username: ""
    password: ""
```

## Topic Definitions

### Event Topics

| Topic Name | Partitions | Replication | Retention | Description |
|------------|------------|-------------|-----------|-------------|
| `helixagent.events.llm.responses` | 6 | 3 | 7 days | LLM response events |
| `helixagent.events.debate.rounds` | 6 | 3 | 7 days | Debate round events |
| `helixagent.events.verification.results` | 3 | 3 | 7 days | Verification results |
| `helixagent.events.provider.health` | 3 | 3 | 1 day | Provider health updates |
| `helixagent.events.audit` | 6 | 3 | 30 days | Audit logs |
| `helixagent.events.metrics` | 3 | 3 | 1 day | System metrics |
| `helixagent.events.errors` | 3 | 3 | 7 days | Error events |

### Streaming Topics

| Topic Name | Partitions | Replication | Retention | Description |
|------------|------------|-------------|-----------|-------------|
| `helixagent.stream.tokens` | 12 | 3 | 1 hour | Token streaming |
| `helixagent.stream.sse` | 6 | 3 | 1 hour | SSE events |
| `helixagent.stream.websocket` | 6 | 3 | 1 hour | WebSocket messages |

## Implementation

### Broker Interface

```go
// internal/messaging/kafka/broker.go
type Broker struct {
    config    *Config
    producer  *kafka.Writer
    consumers map[string]*kafka.Reader
    admin     *kafka.Client
    metrics   *messaging.BrokerMetrics
    mu        sync.RWMutex
    connected atomic.Bool
}

func NewBroker(cfg *Config) *Broker
func (b *Broker) Connect(ctx context.Context) error
func (b *Broker) Close(ctx context.Context) error
func (b *Broker) Publish(ctx context.Context, topic string, msg *messaging.Message, opts ...messaging.PublishOption) error
func (b *Broker) Subscribe(ctx context.Context, topic string, handler messaging.MessageHandler, opts ...messaging.SubscribeOption) (messaging.Subscription, error)
```

### Producer Implementation

```go
func (b *Broker) newProducer() *kafka.Writer {
    return &kafka.Writer{
        Addr:         kafka.TCP(b.config.Brokers...),
        Topic:        "", // Set per message
        Balancer:     &kafka.LeastBytes{},
        BatchSize:    b.config.BatchSize,
        BatchTimeout: time.Duration(b.config.LingerMs) * time.Millisecond,
        Compression:  b.compressionCodec(),
        RequiredAcks: b.requiredAcks(),
    }
}

func (b *Broker) Publish(ctx context.Context, topic string, msg *messaging.Message, opts ...messaging.PublishOption) error {
    kafkaMsg := kafka.Message{
        Topic:   topic,
        Key:     []byte(msg.ID),
        Value:   msg.Payload,
        Time:    msg.Timestamp,
        Headers: b.convertHeaders(msg.Headers),
    }

    start := time.Now()
    err := b.producer.WriteMessages(ctx, kafkaMsg)
    b.metrics.PublishLatencyTotal.Add(int64(time.Since(start)))
    b.metrics.PublishLatencyCount.Add(1)

    if err != nil {
        b.metrics.PublishFailures.Add(1)
        return &messaging.Error{
            Code:      messaging.ErrPublishFailed,
            Message:   "kafka publish failed",
            Cause:     err,
            Retryable: true,
        }
    }

    b.metrics.PublishSuccesses.Add(1)
    b.metrics.MessagesPublished.Add(1)
    b.metrics.BytesPublished.Add(int64(len(msg.Payload)))
    return nil
}
```

### Consumer Group Implementation

```go
func (b *Broker) newConsumer(topic, groupID string) *kafka.Reader {
    return kafka.NewReader(kafka.ReaderConfig{
        Brokers:        b.config.Brokers,
        Topic:          topic,
        GroupID:        groupID,
        MinBytes:       1,
        MaxBytes:       10e6, // 10MB
        MaxWait:        b.config.MaxWaitTime,
        StartOffset:    b.startOffset(),
        CommitInterval: 0, // Manual commit
    })
}

func (b *Broker) Subscribe(ctx context.Context, topic string, handler messaging.MessageHandler, opts ...messaging.SubscribeOption) (messaging.Subscription, error) {
    options := messaging.DefaultSubscribeOptions()
    for _, opt := range opts {
        opt(options)
    }

    consumer := b.newConsumer(topic, options.GroupID)
    b.mu.Lock()
    b.consumers[topic] = consumer
    b.mu.Unlock()

    sub := &kafkaSubscription{
        topic:    topic,
        consumer: consumer,
        handler:  handler,
        active:   atomic.Bool{},
        done:     make(chan struct{}),
    }
    sub.active.Store(true)

    go sub.consumeLoop(ctx, b.metrics)

    return sub, nil
}
```

### Offset Management

```go
func (s *kafkaSubscription) consumeLoop(ctx context.Context, metrics *messaging.BrokerMetrics) {
    for {
        select {
        case <-ctx.Done():
            return
        case <-s.done:
            return
        default:
        }

        kafkaMsg, err := s.consumer.FetchMessage(ctx)
        if err != nil {
            if err == context.Canceled {
                return
            }
            metrics.MessagesDeadLettered.Add(1)
            continue
        }

        msg := &messaging.Message{
            ID:        string(kafkaMsg.Key),
            Payload:   kafkaMsg.Value,
            Timestamp: kafkaMsg.Time,
            Headers:   convertKafkaHeaders(kafkaMsg.Headers),
        }

        metrics.MessagesReceived.Add(1)

        if err := s.handler(ctx, msg); err != nil {
            metrics.MessagesFailed.Add(1)
            // Don't commit - message will be redelivered
            continue
        }

        // Commit offset after successful processing
        if err := s.consumer.CommitMessages(ctx, kafkaMsg); err != nil {
            metrics.MessagesFailed.Add(1)
            continue
        }

        metrics.MessagesProcessed.Add(1)
    }
}
```

## Message Flow

### Event Publishing

```
┌──────────┐     ┌──────────┐     ┌──────────────┐     ┌──────────┐
│ Producer │ ──► │ Kafka    │ ──► │ Partition 0  │ ──► │ Consumer │
│          │     │ Topic    │     │ Partition 1  │     │ Group    │
│          │     │          │     │ Partition 2  │     │          │
└──────────┘     └──────────┘     └──────────────┘     └──────────┘
                                         │
                                         │ Offset commit
                                         ▼
                              ┌─────────────────────┐
                              │ Consumer Offsets    │
                              │ __consumer_offsets  │
                              └─────────────────────┘
```

### Consumer Group Rebalancing

```
┌─────────────────────────────────────────────────────────────────┐
│                    Consumer Group: helixagent-group             │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Consumer 1          Consumer 2          Consumer 3              │
│  ┌───────────┐       ┌───────────┐       ┌───────────┐         │
│  │ P0, P1    │       │ P2, P3    │       │ P4, P5    │         │
│  └───────────┘       └───────────┘       └───────────┘         │
│                                                                  │
│  ◄──────────────── Partition Assignment ─────────────────►      │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Partitioning Strategy

### Key-Based Partitioning

Messages with the same key go to the same partition:

```go
func (b *Broker) PublishWithKey(ctx context.Context, topic string, key string, msg *messaging.Message) error {
    kafkaMsg := kafka.Message{
        Topic: topic,
        Key:   []byte(key),  // e.g., debate_id, user_id
        Value: msg.Payload,
    }
    return b.producer.WriteMessages(ctx, kafkaMsg)
}
```

### Partition Strategies

| Strategy | Use Case | Example |
|----------|----------|---------|
| Round-robin | Even distribution | Metrics events |
| Key-based | Ordering guarantee | Debate rounds (by debate_id) |
| Custom | Special routing | LLM responses (by provider) |

## Compression

### Supported Codecs

| Codec | CPU | Compression | Use Case |
|-------|-----|-------------|----------|
| none | - | 1:1 | Low latency |
| gzip | High | ~70% | Cold storage |
| snappy | Low | ~50% | Balanced |
| lz4 | Very Low | ~60% | **Recommended** |
| zstd | Medium | ~75% | Best ratio |

### Configuration

```yaml
kafka:
  compression: lz4  # Best balance of speed and compression
```

## Monitoring

### Kafka UI

Access the Kafka UI at `http://localhost:8080`:
- View topics, partitions, consumer groups
- Monitor lag and throughput
- Browse messages

### Key Metrics

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| `consumer_lag` | Messages behind | > 10,000 |
| `request_rate` | Requests/sec | > 50,000 |
| `byte_rate` | Bytes/sec | > 100MB/s |
| `partition_count` | Partitions per topic | > 100 |

### Prometheus Metrics

```go
// Exposed at /metrics
kafka_messages_produced_total
kafka_messages_consumed_total
kafka_consumer_lag
kafka_produce_latency_seconds
kafka_consume_latency_seconds
kafka_connection_state
kafka_partition_offset
```

## Docker Deployment

### Development

```bash
docker-compose -f docker-compose.messaging.yml --profile messaging up -d zookeeper kafka
```

### Production

```yaml
# docker-compose.prod.yml
services:
  kafka:
    image: confluentinc/cp-kafka:7.5.0
    deploy:
      replicas: 3
      resources:
        limits:
          memory: 4G
          cpus: '4'
    environment:
      KAFKA_BROKER_ID: ${KAFKA_BROKER_ID}
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_LISTENER_SECURITY_PROTOCOL_MAP: PLAINTEXT:PLAINTEXT,SSL:SSL
      KAFKA_SSL_KEYSTORE_LOCATION: /etc/kafka/secrets/kafka.keystore.jks
      KAFKA_SSL_TRUSTSTORE_LOCATION: /etc/kafka/secrets/kafka.truststore.jks
    volumes:
      - kafka_data:/var/lib/kafka/data
      - ./certs:/etc/kafka/secrets:ro
```

## Event Replay

### Seek to Offset

```go
func (b *Broker) SeekToOffset(ctx context.Context, topic string, partition int32, offset int64) error {
    consumer := b.consumers[topic]
    if consumer == nil {
        return ErrConsumerNotFound
    }
    return consumer.SetOffset(offset)
}
```

### Seek to Timestamp

```go
func (b *Broker) SeekToTimestamp(ctx context.Context, topic string, partition int32, ts time.Time) error {
    // Use admin client to get offset for timestamp
    offsets, err := b.admin.ListOffsets(ctx, map[string][]kafka.OffsetRequest{
        topic: {{Partition: partition, Timestamp: ts}},
    })
    if err != nil {
        return err
    }

    return b.SeekToOffset(ctx, topic, partition, offsets[topic][partition].Offset)
}
```

## Best Practices

### Producer Design

1. **Use batching** - Improve throughput
2. **Enable compression** - Reduce bandwidth
3. **Set appropriate acks** - Balance durability vs latency
4. **Handle failures** - Implement retry logic

### Consumer Design

1. **Manual commits** - Control exactly-once processing
2. **Idempotent handlers** - Messages may be redelivered
3. **Parallel processing** - One consumer per partition max
4. **Graceful shutdown** - Commit offsets before exit

### Topic Design

1. **Partition by key** - Maintain ordering where needed
2. **Right-size partitions** - ~1GB/partition/day
3. **Set retention** - Based on replay requirements
4. **Use compaction** - For stateful topics

## Troubleshooting

### Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| Connection refused | Kafka not running | Start Kafka container |
| Leader not available | Broker not ready | Wait for election |
| Consumer lag growing | Slow processing | Add consumers |
| Message too large | Default 1MB limit | Increase `max.message.bytes` |
| Offset out of range | Retention expired | Use `earliest` offset |

### Debug Commands

```bash
# List topics
docker exec helixagent-kafka kafka-topics --bootstrap-server localhost:9092 --list

# Describe topic
docker exec helixagent-kafka kafka-topics --bootstrap-server localhost:9092 --describe --topic helixagent.events.audit

# Consumer groups
docker exec helixagent-kafka kafka-consumer-groups --bootstrap-server localhost:9092 --list

# Consumer lag
docker exec helixagent-kafka kafka-consumer-groups --bootstrap-server localhost:9092 --describe --group helixagent-group

# Produce test message
docker exec helixagent-kafka kafka-console-producer --bootstrap-server localhost:9092 --topic test

# Consume messages
docker exec helixagent-kafka kafka-console-consumer --bootstrap-server localhost:9092 --topic test --from-beginning
```

## Testing

```bash
# Run Kafka unit tests
go test ./internal/messaging/kafka/... -v

# Run Kafka challenge
./challenges/scripts/messaging_kafka_challenge.sh

# Integration tests (requires Kafka running)
KAFKA_BROKERS=localhost:9092 go test ./tests/integration/... -run TestKafka
```

## Security

### Authentication (SASL)

```yaml
kafka:
  sasl:
    enabled: true
    mechanism: SCRAM-SHA-512
    username: helixagent
    password: ${KAFKA_PASSWORD}
```

### Authorization (ACLs)

```bash
# Grant producer permission
docker exec helixagent-kafka kafka-acls --bootstrap-server localhost:9092 \
  --add --allow-principal User:helixagent \
  --producer --topic 'helixagent.*'

# Grant consumer permission
docker exec helixagent-kafka kafka-acls --bootstrap-server localhost:9092 \
  --add --allow-principal User:helixagent \
  --consumer --topic 'helixagent.*' --group 'helixagent-*'
```

### TLS Configuration

```yaml
kafka:
  tls:
    enabled: true
    cert_file: /etc/kafka/certs/client.crt
    key_file: /etc/kafka/certs/client.key
    ca_file: /etc/kafka/certs/ca.crt
```
