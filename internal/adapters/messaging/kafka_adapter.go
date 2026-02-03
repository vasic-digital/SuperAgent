package messaging

import (
	"context"

	"digital.vasic.messaging/pkg/kafka"

	"dev.helix.agent/internal/messaging"
)

// KafkaProducerAdapter wraps the generic kafka.Producer for HelixAgent.
type KafkaProducerAdapter struct {
	producer *kafka.Producer
	metrics  *messaging.BrokerMetrics
}

// NewKafkaProducerAdapter creates a new Kafka producer adapter.
func NewKafkaProducerAdapter(config *kafka.Config) *KafkaProducerAdapter {
	return &KafkaProducerAdapter{
		producer: kafka.NewProducer(config),
		metrics:  messaging.NewBrokerMetrics(),
	}
}

// Connect establishes connection to Kafka.
func (a *KafkaProducerAdapter) Connect(ctx context.Context) error {
	a.metrics.RecordConnectionAttempt()
	if err := a.producer.Connect(ctx); err != nil {
		a.metrics.RecordConnectionFailure()
		return err
	}
	a.metrics.RecordConnectionSuccess()
	return nil
}

// Close closes the producer.
func (a *KafkaProducerAdapter) Close(ctx context.Context) error {
	a.metrics.RecordDisconnection()
	return a.producer.Close(ctx)
}

// IsConnected returns true if connected.
func (a *KafkaProducerAdapter) IsConnected() bool {
	return a.producer.IsConnected()
}

// Publish sends a message to a Kafka topic.
func (a *KafkaProducerAdapter) Publish(ctx context.Context, topic string, message *messaging.Message) error {
	genericMsg := InternalToGenericMessage(message)
	return a.producer.Publish(ctx, topic, genericMsg)
}

// GetMetrics returns broker metrics.
func (a *KafkaProducerAdapter) GetMetrics() *messaging.BrokerMetrics {
	return a.metrics
}

// Config returns the producer configuration.
func (a *KafkaProducerAdapter) Config() *kafka.Config {
	return a.producer.Config()
}

// KafkaConsumerAdapter wraps the generic kafka.Consumer for HelixAgent.
type KafkaConsumerAdapter struct {
	consumer *kafka.Consumer
	metrics  *messaging.BrokerMetrics
}

// NewKafkaConsumerAdapter creates a new Kafka consumer adapter.
func NewKafkaConsumerAdapter(config *kafka.Config) *KafkaConsumerAdapter {
	return &KafkaConsumerAdapter{
		consumer: kafka.NewConsumer(config),
		metrics:  messaging.NewBrokerMetrics(),
	}
}

// Connect establishes connection to Kafka.
func (a *KafkaConsumerAdapter) Connect(ctx context.Context) error {
	a.metrics.RecordConnectionAttempt()
	if err := a.consumer.Connect(ctx); err != nil {
		a.metrics.RecordConnectionFailure()
		return err
	}
	a.metrics.RecordConnectionSuccess()
	return nil
}

// Close closes the consumer.
func (a *KafkaConsumerAdapter) Close(ctx context.Context) error {
	a.metrics.RecordDisconnection()
	return a.consumer.Close(ctx)
}

// IsConnected returns true if connected.
func (a *KafkaConsumerAdapter) IsConnected() bool {
	return a.consumer.IsConnected()
}

// Subscribe creates a subscription to a Kafka topic.
func (a *KafkaConsumerAdapter) Subscribe(ctx context.Context, topic string, handler messaging.MessageHandler) (messaging.Subscription, error) {
	genericHandler := InternalToGenericHandler(handler)
	sub, err := a.consumer.Subscribe(ctx, topic, genericHandler)
	if err != nil {
		return nil, err
	}
	a.metrics.RecordSubscription()
	return &SubscriptionAdapter{sub: sub, metrics: a.metrics}, nil
}

// Unsubscribe cancels a subscription.
func (a *KafkaConsumerAdapter) Unsubscribe(topic string) error {
	a.metrics.RecordUnsubscription()
	return a.consumer.Unsubscribe(topic)
}

// GetMetrics returns broker metrics.
func (a *KafkaConsumerAdapter) GetMetrics() *messaging.BrokerMetrics {
	return a.metrics
}

// Config returns the consumer configuration.
func (a *KafkaConsumerAdapter) Config() *kafka.Config {
	return a.consumer.Config()
}
