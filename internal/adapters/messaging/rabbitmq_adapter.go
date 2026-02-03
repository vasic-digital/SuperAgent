package messaging

import (
	"context"

	"digital.vasic.messaging/pkg/rabbitmq"

	"dev.helix.agent/internal/messaging"
)

// RabbitMQProducerAdapter wraps the generic rabbitmq.Producer for HelixAgent.
type RabbitMQProducerAdapter struct {
	producer *rabbitmq.Producer
	metrics  *messaging.BrokerMetrics
}

// NewRabbitMQProducerAdapter creates a new RabbitMQ producer adapter.
func NewRabbitMQProducerAdapter(config *rabbitmq.Config) *RabbitMQProducerAdapter {
	return &RabbitMQProducerAdapter{
		producer: rabbitmq.NewProducer(config),
		metrics:  messaging.NewBrokerMetrics(),
	}
}

// Connect establishes connection to RabbitMQ.
func (a *RabbitMQProducerAdapter) Connect(ctx context.Context) error {
	a.metrics.RecordConnectionAttempt()
	if err := a.producer.Connect(ctx); err != nil {
		a.metrics.RecordConnectionFailure()
		return err
	}
	a.metrics.RecordConnectionSuccess()
	return nil
}

// Close closes the producer.
func (a *RabbitMQProducerAdapter) Close(ctx context.Context) error {
	a.metrics.RecordDisconnection()
	return a.producer.Close(ctx)
}

// IsConnected returns true if connected.
func (a *RabbitMQProducerAdapter) IsConnected() bool {
	return a.producer.IsConnected()
}

// Publish sends a message to a RabbitMQ queue/exchange.
func (a *RabbitMQProducerAdapter) Publish(ctx context.Context, topic string, message *messaging.Message) error {
	genericMsg := InternalToGenericMessage(message)
	return a.producer.Publish(ctx, topic, genericMsg)
}

// GetMetrics returns broker metrics.
func (a *RabbitMQProducerAdapter) GetMetrics() *messaging.BrokerMetrics {
	return a.metrics
}

// Config returns the producer configuration.
func (a *RabbitMQProducerAdapter) Config() *rabbitmq.Config {
	return a.producer.Config()
}

// RabbitMQConsumerAdapter wraps the generic rabbitmq.Consumer for HelixAgent.
type RabbitMQConsumerAdapter struct {
	consumer *rabbitmq.Consumer
	metrics  *messaging.BrokerMetrics
}

// NewRabbitMQConsumerAdapter creates a new RabbitMQ consumer adapter.
func NewRabbitMQConsumerAdapter(config *rabbitmq.Config) *RabbitMQConsumerAdapter {
	return &RabbitMQConsumerAdapter{
		consumer: rabbitmq.NewConsumer(config),
		metrics:  messaging.NewBrokerMetrics(),
	}
}

// Connect establishes connection to RabbitMQ.
func (a *RabbitMQConsumerAdapter) Connect(ctx context.Context) error {
	a.metrics.RecordConnectionAttempt()
	if err := a.consumer.Connect(ctx); err != nil {
		a.metrics.RecordConnectionFailure()
		return err
	}
	a.metrics.RecordConnectionSuccess()
	return nil
}

// Close closes the consumer.
func (a *RabbitMQConsumerAdapter) Close(ctx context.Context) error {
	a.metrics.RecordDisconnection()
	return a.consumer.Close(ctx)
}

// IsConnected returns true if connected.
func (a *RabbitMQConsumerAdapter) IsConnected() bool {
	return a.consumer.IsConnected()
}

// Subscribe creates a subscription to a RabbitMQ queue.
func (a *RabbitMQConsumerAdapter) Subscribe(ctx context.Context, topic string, handler messaging.MessageHandler) (messaging.Subscription, error) {
	genericHandler := InternalToGenericHandler(handler)
	sub, err := a.consumer.Subscribe(ctx, topic, genericHandler)
	if err != nil {
		return nil, err
	}
	a.metrics.RecordSubscription()
	return &SubscriptionAdapter{sub: sub, metrics: a.metrics}, nil
}

// Unsubscribe cancels a subscription.
func (a *RabbitMQConsumerAdapter) Unsubscribe(topic string) error {
	a.metrics.RecordUnsubscription()
	return a.consumer.Unsubscribe(topic)
}

// GetMetrics returns broker metrics.
func (a *RabbitMQConsumerAdapter) GetMetrics() *messaging.BrokerMetrics {
	return a.metrics
}

// Config returns the consumer configuration.
func (a *RabbitMQConsumerAdapter) Config() *rabbitmq.Config {
	return a.consumer.Config()
}
