# Messaging Quickstart Guide

Get started with HelixAgent's messaging system in 5 minutes.

## Prerequisites

- Docker and Docker Compose installed
- Go 1.24+ installed
- HelixAgent repository cloned

## Quick Start

### 1. Start Messaging Infrastructure

```bash
# Start RabbitMQ and Kafka
docker-compose -f docker-compose.messaging.yml --profile messaging up -d

# Verify services are running
docker-compose -f docker-compose.messaging.yml ps
```

Expected output:
```
NAME                        STATUS    PORTS
helixagent-rabbitmq         Up        5672, 15672
helixagent-zookeeper        Up        2181
helixagent-kafka            Up        9092, 29092
helixagent-schema-registry  Up        8081
```

### 2. Access Management UIs

| Service | URL | Credentials |
|---------|-----|-------------|
| RabbitMQ | http://localhost:15672 | helixagent / helixagent123 |
| Kafka UI | http://localhost:8080 | - |

### 3. Run the Application

```bash
# Start HelixAgent with messaging
make run
```

HelixAgent will automatically:
- Connect to RabbitMQ and Kafka
- Declare queues and topics
- Start consuming messages

## Basic Usage

### Publishing Tasks (RabbitMQ)

```go
import (
    "context"
    "dev.helix.agent/internal/messaging"
    "dev.helix.agent/internal/messaging/rabbitmq"
)

// Create broker
cfg := rabbitmq.DefaultConfig()
broker := rabbitmq.NewBroker(cfg)

// Connect
ctx := context.Background()
if err := broker.Connect(ctx); err != nil {
    log.Fatal(err)
}
defer broker.Close(ctx)

// Publish a task
msg := messaging.NewMessage("llm.request", []byte(`{
    "prompt": "Hello, world!",
    "model": "claude-3"
}`))
msg.WithPriority(messaging.PriorityHigh)

if err := broker.Publish(ctx, "helixagent.tasks.llm", msg); err != nil {
    log.Fatal(err)
}
```

### Publishing Events (Kafka)

```go
import (
    "context"
    "dev.helix.agent/internal/messaging"
    "dev.helix.agent/internal/messaging/kafka"
)

// Create broker
cfg := kafka.DefaultConfig()
broker := kafka.NewBroker(cfg)

// Connect
ctx := context.Background()
if err := broker.Connect(ctx); err != nil {
    log.Fatal(err)
}
defer broker.Close(ctx)

// Publish an event
msg := messaging.NewMessage("debate.round", []byte(`{
    "debate_id": "123",
    "round": 1,
    "content": "..."
}`))

if err := broker.Publish(ctx, "helixagent.events.debate.rounds", msg); err != nil {
    log.Fatal(err)
}
```

### Subscribing to Messages

```go
// Subscribe to tasks
handler := func(ctx context.Context, msg *messaging.Message) error {
    fmt.Printf("Received: %s\n", string(msg.Payload))
    return nil
}

sub, err := broker.Subscribe(ctx, "helixagent.tasks.llm", handler)
if err != nil {
    log.Fatal(err)
}
defer sub.Unsubscribe()

// Block until shutdown
<-ctx.Done()
```

## Using the Messaging Hub

The MessagingHub provides a unified interface:

```go
import "dev.helix.agent/internal/messaging"

// Initialize hub
hub, err := messaging.NewHub(&messaging.HubConfig{
    RabbitMQ: rabbitmq.DefaultConfig(),
    Kafka:    kafka.DefaultConfig(),
})
if err != nil {
    log.Fatal(err)
}

ctx := context.Background()
if err := hub.Initialize(ctx); err != nil {
    log.Fatal(err)
}
defer hub.Shutdown(ctx)

// Publish task (goes to RabbitMQ)
task := &messaging.Task{
    Type:    "llm.request",
    Payload: []byte(`{"prompt": "Hello"}`),
}
hub.PublishTask(ctx, "llm", task)

// Publish event (goes to Kafka)
event := &messaging.Event{
    Type:   "debate.started",
    Source: "debate-service",
    Data:   []byte(`{"debate_id": "123"}`),
}
hub.PublishEvent(ctx, "helixagent.events.debate.rounds", event)
```

## In-Memory Fallback

For testing without infrastructure:

```go
import "dev.helix.agent/internal/messaging/inmemory"

// Create in-memory broker
broker := inmemory.NewBroker(nil)

ctx := context.Background()
broker.Connect(ctx)
defer broker.Close(ctx)

// Use exactly like RabbitMQ/Kafka
msg := messaging.NewMessage("test", []byte(`{"data": "test"}`))
broker.Publish(ctx, "test.topic", msg)
```

## Testing Your Setup

### Run Integration Tests

```bash
# Start infrastructure
docker-compose -f docker-compose.messaging.yml --profile messaging up -d

# Run tests
go test ./internal/messaging/... -v
go test ./tests/integration/... -v
```

### Run Challenge Scripts

```bash
# RabbitMQ challenge (25 tests)
./challenges/scripts/messaging_rabbitmq_challenge.sh

# Kafka challenge (30 tests)
./challenges/scripts/messaging_kafka_challenge.sh

# Full hybrid challenge (37 tests)
./challenges/scripts/messaging_hybrid_challenge.sh
```

## Configuration Reference

### Environment Variables

```bash
# RabbitMQ
RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
RABBITMQ_USER=helixagent
RABBITMQ_PASSWORD=helixagent123

# Kafka
KAFKA_BROKERS=localhost:9092
KAFKA_CLIENT_ID=helixagent
KAFKA_GROUP_ID=helixagent-group
```

### Configuration File

Create `configs/messaging.yaml`:

```yaml
# See configs/messaging.yaml for full example
rabbitmq:
  host: localhost
  port: 5672
  username: helixagent
  password: helixagent123
  publisher_confirm: true
  prefetch_count: 10

kafka:
  brokers:
    - localhost:9092
  client_id: helixagent
  group_id: helixagent-group
  compression: lz4
```

## Stopping Infrastructure

```bash
# Stop services
docker-compose -f docker-compose.messaging.yml down

# Stop and remove volumes (clean slate)
docker-compose -f docker-compose.messaging.yml down -v
```

## Troubleshooting

### Connection Refused

```bash
# Check if containers are running
docker-compose -f docker-compose.messaging.yml ps

# Check logs
docker-compose -f docker-compose.messaging.yml logs rabbitmq
docker-compose -f docker-compose.messaging.yml logs kafka
```

### Authentication Errors

Verify credentials in your configuration match `docker-compose.messaging.yml`.

### Messages Not Being Consumed

```bash
# Check RabbitMQ queues
docker exec helixagent-rabbitmq rabbitmqctl list_queues

# Check Kafka consumer groups
docker exec helixagent-kafka kafka-consumer-groups --bootstrap-server localhost:9092 --list
```

## Next Steps

- Read [RabbitMQ Integration](../architecture/rabbitmq-integration.md) for advanced usage
- Read [Kafka Integration](../architecture/kafka-integration.md) for event streaming
- Read [TOON Protocol](../architecture/toon-protocol.md) for token optimization
- Read [GraphQL API](../architecture/graphql-api.md) for API integration
