# Messaging Package

The messaging package provides enterprise message broker integration for HelixAgent, supporting RabbitMQ, Kafka, and in-memory fallback.

## Overview

This package implements:

- **Messaging Hub**: Central hub for message routing
- **RabbitMQ Integration**: Task queue with reliable delivery
- **Kafka Integration**: Event streaming with partitioning
- **In-Memory Fallback**: Development and testing mode
- **Automatic Failover**: Seamless broker switching

## Key Components

### Messaging System

Main entry point for messaging operations:

```go
config := messaging.LoadMessagingConfigFromEnv()
system := messaging.NewMessagingSystem(config, logger)

// Set fallback broker factory
system.FallbackBrokerFactory = func() messaging.MessageBroker {
    return messaging.NewInMemoryBroker()
}

err := system.Initialize(ctx)
defer system.Close(ctx)
```

### Messaging Hub

Central message routing:

```go
hub := messaging.NewMessagingHub(hubConfig)

// Set brokers
hub.SetTaskQueueBroker(rabbitBroker)
hub.SetEventStreamBroker(kafkaBroker)
hub.SetFallbackBroker(inMemoryBroker)

// Publish task
err := hub.PublishTask(ctx, "task.process", task)

// Publish event
err := hub.PublishEvent(ctx, "event.completed", event)
```

### Message Broker Interface

```go
type MessageBroker interface {
    Connect(ctx context.Context) error
    Close(ctx context.Context) error
    Publish(ctx context.Context, topic string, msg *Message) error
    Subscribe(ctx context.Context, topic string, handler MessageHandler) error
    Unsubscribe(ctx context.Context, topic string) error
}
```

## Broker Types

### RabbitMQ (Task Queue)

Best for:
- Reliable task delivery
- Work queues with acknowledgments
- Fan-out patterns

```go
broker := messaging.NewRabbitMQBroker(&messaging.RabbitMQConfig{
    Host:     "localhost",
    Port:     5672,
    Username: "guest",
    Password: "guest",
    VHost:    "/",
})
```

### Kafka (Event Stream)

Best for:
- High-throughput event streaming
- Log aggregation
- Event sourcing

```go
broker := messaging.NewKafkaBroker(&messaging.KafkaConfig{
    Brokers:  []string{"localhost:9092"},
    ClientID: "helixagent",
    GroupID:  "helixagent-group",
})
```

### In-Memory (Development)

For testing and development:

```go
broker := messaging.NewInMemoryBroker()
```

## Configuration

```go
type MessagingConfig struct {
    Enabled            bool
    RabbitMQ           RabbitMQConfig
    Kafka              KafkaConfig
    FallbackToInMemory bool
    ConnectionTimeout  time.Duration
}
```

### Environment Variables

```bash
# General
MESSAGING_ENABLED=true
MESSAGING_FALLBACK_INMEMORY=true

# RabbitMQ
RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
RABBITMQ_USER=helixagent
RABBITMQ_PASSWORD=secret
RABBITMQ_VHOST=/

# Kafka
KAFKA_BROKERS=localhost:9092
KAFKA_CLIENT_ID=helixagent
KAFKA_GROUP_ID=helixagent-group
```

## Usage Examples

### Publish/Subscribe

```go
// Subscribe to messages
hub.Subscribe(ctx, "task.created", func(ctx context.Context, msg *messaging.Message) error {
    var task Task
    if err := json.Unmarshal(msg.Body, &task); err != nil {
        return err
    }
    return processTask(task)
})

// Publish message
err := hub.Publish(ctx, "task.created", &messaging.Message{
    ID:        uuid.New().String(),
    Body:      taskJSON,
    Timestamp: time.Now(),
})
```

### Task Queue Pattern

```go
// Worker pool
for i := 0; i < numWorkers; i++ {
    go func() {
        hub.Subscribe(ctx, "tasks", func(ctx context.Context, msg *messaging.Message) error {
            // Process task
            return nil // Acknowledge
        })
    }()
}

// Enqueue task
hub.PublishTask(ctx, "tasks", task)
```

### Event Sourcing

```go
// Publish events
hub.PublishEvent(ctx, "user.created", UserCreatedEvent{...})
hub.PublishEvent(ctx, "user.updated", UserUpdatedEvent{...})

// Replay events
hub.ReplayEvents(ctx, "user.*", startTime, handler)
```

## Automatic Failover

The hub automatically fails over to fallback broker:

```go
hubConfig := messaging.DefaultHubConfig()
hubConfig.UseFallbackOnError = true

hub := messaging.NewMessagingHub(hubConfig)
hub.SetTaskQueueBroker(rabbitBroker)
hub.SetFallbackBroker(inMemoryBroker)

// If RabbitMQ fails, automatically uses in-memory
```

## Testing

```bash
# Run messaging tests
go test -v ./internal/messaging/...

# Test with infrastructure
make test-infra-start
go test -v ./internal/messaging/...
make test-infra-stop
```

## See Also

- `internal/background/` - Background task execution
- `internal/notifications/` - Notification delivery
- Docker Compose for infrastructure setup
