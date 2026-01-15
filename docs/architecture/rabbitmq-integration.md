# RabbitMQ Integration

HelixAgent uses RabbitMQ as its primary task queue broker for reliable, low-latency job processing.

## Overview

RabbitMQ handles synchronous task operations where immediate reliability and acknowledgment are critical:
- Background task execution
- LLM request queuing with priority support
- Debate round coordination
- Real-time notifications
- Verification tasks

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      RabbitMQ Broker                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐         │
│  │  Exchange   │    │  Exchange   │    │  Exchange   │         │
│  │  (tasks)    │    │  (events)   │    │(notifications)│       │
│  │   direct    │    │   topic     │    │   fanout    │         │
│  └──────┬──────┘    └──────┬──────┘    └──────┬──────┘         │
│         │                  │                  │                  │
│  ┌──────▼──────┐    ┌──────▼──────┐    ┌──────▼──────┐         │
│  │   Queues    │    │   Queues    │    │   Queues    │         │
│  │ .background │    │  .debate    │    │ .websocket  │         │
│  │    .llm     │    │ .verify     │    │   .sse      │         │
│  │  .priority  │    │  .audit     │    │ .webhooks   │         │
│  └─────────────┘    └─────────────┘    └─────────────┘         │
│                                                                  │
│  ┌─────────────┐    ┌─────────────┐                             │
│  │  Dead Letter│    │   Retry     │                             │
│  │    Queue    │    │   Queue     │                             │
│  │ helixagent  │    │   60s TTL   │                             │
│  │   .dlq      │    │             │                             │
│  └─────────────┘    └─────────────┘                             │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Configuration

### Environment Variables

```bash
RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
RABBITMQ_USER=helixagent
RABBITMQ_PASSWORD=helixagent123
RABBITMQ_VHOST=/
RABBITMQ_PREFETCH_COUNT=10
RABBITMQ_PUBLISHER_CONFIRM=true
```

### Configuration File (`configs/messaging.yaml`)

```yaml
rabbitmq:
  host: ${RABBITMQ_HOST:-localhost}
  port: ${RABBITMQ_PORT:-5672}
  username: ${RABBITMQ_USER:-helixagent}
  password: ${RABBITMQ_PASSWORD:-helixagent123}
  vhost: ${RABBITMQ_VHOST:-/}

  # Connection settings
  connect_timeout: 30s
  heartbeat: 10s
  max_reconnect_attempts: 10
  reconnect_delay: 5s

  # Publisher settings
  publisher_confirm: true
  publish_timeout: 10s

  # Consumer settings
  prefetch_count: 10
  prefetch_global: false

  # TLS settings (production)
  tls:
    enabled: false
    cert_file: ""
    key_file: ""
    ca_file: ""
    skip_verify: false
```

## Queue Definitions

### Task Queues

| Queue Name | Exchange | Routing Key | Priority | Description |
|------------|----------|-------------|----------|-------------|
| `helixagent.tasks.background` | `helixagent.tasks` | `background` | No | General background jobs |
| `helixagent.tasks.llm` | `helixagent.tasks` | `llm` | Yes (0-10) | LLM API requests |
| `helixagent.tasks.debate` | `helixagent.tasks` | `debate` | No | Debate round processing |
| `helixagent.tasks.verification` | `helixagent.tasks` | `verification` | No | Provider verification |
| `helixagent.tasks.notifications` | `helixagent.tasks` | `notifications` | No | Notification dispatch |

### System Queues

| Queue Name | Purpose | TTL | Max Length |
|------------|---------|-----|------------|
| `helixagent.dlq` | Dead letter queue | 7 days | 100,000 |
| `helixagent.retry` | Retry queue | 60s | 10,000 |

## Exchange Definitions

| Exchange Name | Type | Durable | Description |
|---------------|------|---------|-------------|
| `helixagent.tasks` | direct | Yes | Task routing by type |
| `helixagent.events` | topic | Yes | Event pattern matching |
| `helixagent.notifications` | fanout | Yes | Broadcast notifications |

## Implementation

### Broker Interface

```go
// internal/messaging/rabbitmq/broker.go
type Broker struct {
    config     *Config
    conn       *amqp.Connection
    channel    *amqp.Channel
    confirms   chan amqp.Confirmation
    metrics    *messaging.BrokerMetrics
    reconnectMu sync.RWMutex
    connected  atomic.Bool
}

func NewBroker(cfg *Config) *Broker
func (b *Broker) Connect(ctx context.Context) error
func (b *Broker) Close(ctx context.Context) error
func (b *Broker) Publish(ctx context.Context, topic string, msg *messaging.Message, opts ...messaging.PublishOption) error
func (b *Broker) Subscribe(ctx context.Context, topic string, handler messaging.MessageHandler, opts ...messaging.SubscribeOption) (messaging.Subscription, error)
```

### Publisher Confirms

RabbitMQ publisher confirms ensure guaranteed delivery:

```go
func (b *Broker) publishWithConfirm(ctx context.Context, exchange, routingKey string, msg *messaging.Message) error {
    // Enable confirm mode
    if err := b.channel.Confirm(false); err != nil {
        return err
    }

    // Publish message
    if err := b.channel.PublishWithContext(ctx, exchange, routingKey, false, false, amqp.Publishing{
        DeliveryMode: amqp.Persistent,
        ContentType:  "application/json",
        Body:         msg.Payload,
        MessageId:    msg.ID,
        Timestamp:    msg.Timestamp,
        Headers:      amqp.Table(msg.Headers),
    }); err != nil {
        return err
    }

    // Wait for confirm
    select {
    case confirm := <-b.confirms:
        if !confirm.Ack {
            return ErrPublishNack
        }
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

### Dead Letter Queue Handling

Failed messages are automatically routed to the DLQ:

```go
func (b *Broker) declareQueueWithDLQ(name string) error {
    args := amqp.Table{
        "x-dead-letter-exchange":    "",
        "x-dead-letter-routing-key": "helixagent.dlq",
    }

    _, err := b.channel.QueueDeclare(
        name,  // queue name
        true,  // durable
        false, // auto-delete
        false, // exclusive
        false, // no-wait
        args,  // arguments
    )
    return err
}
```

### Priority Queue Support

LLM requests use priority queues for task prioritization:

```go
func (b *Broker) declarePriorityQueue(name string, maxPriority int) error {
    args := amqp.Table{
        "x-max-priority": maxPriority,
        "x-dead-letter-exchange":    "",
        "x-dead-letter-routing-key": "helixagent.dlq",
    }

    _, err := b.channel.QueueDeclare(
        name,
        true,
        false,
        false,
        false,
        args,
    )
    return err
}
```

### Automatic Reconnection

```go
func (b *Broker) reconnect(ctx context.Context) error {
    b.reconnectMu.Lock()
    defer b.reconnectMu.Unlock()

    for attempt := 1; attempt <= b.config.MaxReconnectAttempts; attempt++ {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(b.config.ReconnectDelay):
        }

        if err := b.connect(ctx); err != nil {
            b.metrics.ConnectionFailures.Add(1)
            continue
        }

        b.connected.Store(true)
        b.metrics.ConnectionSuccesses.Add(1)
        return nil
    }

    return ErrMaxReconnectAttempts
}
```

## Message Flow

### Task Publishing

```
┌──────────┐     ┌──────────────┐     ┌──────────┐     ┌──────────┐
│ Producer │ ──► │  Exchange    │ ──► │  Queue   │ ──► │ Consumer │
│          │     │ (direct)     │     │          │     │          │
└──────────┘     └──────────────┘     └──────────┘     └──────────┘
     │                                                        │
     │ Publish with confirm                         Ack/Nack  │
     │◄──────────────────────────────────────────────────────►│
```

### Failure Handling

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│ Consumer │ ──► │  Nack    │ ──► │  Retry   │ ──► │  Queue   │
│  (fail)  │     │          │     │  Queue   │     │ (retry)  │
└──────────┘     └──────────┘     └──────────┘     └──────────┘
                                       │
                                       │ After TTL
                                       ▼
                                 ┌──────────┐
                                 │   DLQ    │
                                 │ (7 days) │
                                 └──────────┘
```

## Monitoring

### RabbitMQ Management UI

Access the management interface at `http://localhost:15672`:
- Default credentials: `helixagent:helixagent123`

### Key Metrics

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| `queue_messages_ready` | Messages waiting to be processed | > 10,000 |
| `queue_messages_unacked` | Messages being processed | > 1,000 |
| `channel_count` | Active channels | > 100 |
| `connection_count` | Active connections | > 50 |

### Prometheus Metrics

```go
// Exposed at /metrics
rabbitmq_messages_published_total
rabbitmq_messages_consumed_total
rabbitmq_messages_acked_total
rabbitmq_messages_nacked_total
rabbitmq_messages_dead_lettered_total
rabbitmq_publish_latency_seconds
rabbitmq_connection_state
```

## Docker Deployment

### Development

```bash
docker-compose -f docker-compose.messaging.yml --profile messaging up -d rabbitmq
```

### Production

```yaml
# docker-compose.prod.yml
services:
  rabbitmq:
    image: rabbitmq:3.12-management
    deploy:
      replicas: 3
      resources:
        limits:
          memory: 2G
          cpus: '2'
    environment:
      RABBITMQ_ERLANG_COOKIE: ${RABBITMQ_ERLANG_COOKIE}
    volumes:
      - rabbitmq_data:/var/lib/rabbitmq
      - ./configs/rabbitmq/definitions.json:/etc/rabbitmq/definitions.json
      - ./certs:/etc/rabbitmq/certs:ro
```

## Best Practices

### Message Design

1. **Keep messages small** - Large payloads impact queue performance
2. **Use correlation IDs** - Track message flow across services
3. **Set message TTL** - Prevent queue buildup
4. **Enable persistence** - Durable queues and messages for reliability

### Consumer Design

1. **Prefetch wisely** - Balance throughput vs memory
2. **Ack after processing** - Not before
3. **Handle duplicates** - Messages may be redelivered
4. **Graceful shutdown** - Complete in-flight messages

### Connection Management

1. **Connection pooling** - Reuse connections
2. **Channel per thread** - Don't share channels
3. **Heartbeats** - Detect dead connections
4. **Automatic recovery** - Handle network failures

## Troubleshooting

### Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| Connection refused | RabbitMQ not running | Start RabbitMQ container |
| Access refused | Wrong credentials | Check username/password |
| Channel exception | Resource limit | Increase prefetch limit |
| Message loss | Publisher not confirming | Enable publisher confirms |
| DLQ growing | Processing failures | Check consumer logs |

### Debug Commands

```bash
# Check queue status
docker exec helixagent-rabbitmq rabbitmqctl list_queues name messages consumers

# Check connections
docker exec helixagent-rabbitmq rabbitmqctl list_connections user peer_host state

# Check channels
docker exec helixagent-rabbitmq rabbitmqctl list_channels name number prefetch_count

# Purge a queue
docker exec helixagent-rabbitmq rabbitmqctl purge_queue helixagent.dlq
```

## Testing

```bash
# Run RabbitMQ unit tests
go test ./internal/messaging/rabbitmq/... -v

# Run RabbitMQ challenge
./challenges/scripts/messaging_rabbitmq_challenge.sh

# Integration tests (requires RabbitMQ running)
RABBITMQ_HOST=localhost go test ./tests/integration/... -run TestRabbitMQ
```

## Security

### Authentication

- SASL/PLAIN (default)
- SASL/EXTERNAL (TLS client certificates)
- LDAP integration (enterprise)

### Authorization

Configure permissions in `configs/rabbitmq/definitions.json`:

```json
{
  "permissions": [
    {
      "user": "helixagent",
      "vhost": "/",
      "configure": "helixagent\\..*",
      "write": "helixagent\\..*",
      "read": "helixagent\\..*"
    }
  ]
}
```

### TLS Configuration

```yaml
rabbitmq:
  tls:
    enabled: true
    cert_file: /etc/rabbitmq/certs/server.crt
    key_file: /etc/rabbitmq/certs/server.key
    ca_file: /etc/rabbitmq/certs/ca.crt
    verify_peer: true
```
