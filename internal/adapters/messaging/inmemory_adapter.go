package messaging

import (
	"context"

	"digital.vasic.messaging/pkg/broker"

	"dev.helix.agent/internal/messaging"
)

// InMemoryBrokerAdapter wraps the generic broker.InMemoryBroker for HelixAgent.
// It provides a simple in-memory broker for testing and development.
type InMemoryBrokerAdapter struct {
	broker  *broker.InMemoryBroker
	metrics *messaging.BrokerMetrics
}

// NewInMemoryBrokerAdapter creates a new in-memory broker adapter.
func NewInMemoryBrokerAdapter() *InMemoryBrokerAdapter {
	return &InMemoryBrokerAdapter{
		broker:  broker.NewInMemoryBroker(),
		metrics: messaging.NewBrokerMetrics(),
	}
}

// Connect establishes a connection (no-op for in-memory).
func (a *InMemoryBrokerAdapter) Connect(ctx context.Context) error {
	a.metrics.RecordConnectionAttempt()
	if err := a.broker.Connect(ctx); err != nil {
		a.metrics.RecordConnectionFailure()
		return err
	}
	a.metrics.RecordConnectionSuccess()
	return nil
}

// Close closes the broker.
func (a *InMemoryBrokerAdapter) Close(ctx context.Context) error {
	a.metrics.RecordDisconnection()
	return a.broker.Close(ctx)
}

// HealthCheck checks if the broker is healthy.
func (a *InMemoryBrokerAdapter) HealthCheck(ctx context.Context) error {
	return a.broker.HealthCheck(ctx)
}

// IsConnected returns true if connected.
func (a *InMemoryBrokerAdapter) IsConnected() bool {
	return a.broker.IsConnected()
}

// Publish sends a message to a topic or queue.
func (a *InMemoryBrokerAdapter) Publish(ctx context.Context, topic string, message *messaging.Message, opts ...messaging.PublishOption) error {
	genericMsg := InternalToGenericMessage(message)
	if err := a.broker.Publish(ctx, topic, genericMsg); err != nil {
		a.metrics.RecordPublish(int64(len(message.Payload)), 0, false)
		return err
	}
	a.metrics.RecordPublish(int64(len(message.Payload)), 0, true)
	return nil
}

// PublishBatch sends multiple messages to a topic or queue.
func (a *InMemoryBrokerAdapter) PublishBatch(ctx context.Context, topic string, messages []*messaging.Message, opts ...messaging.PublishOption) error {
	for _, msg := range messages {
		if err := a.Publish(ctx, topic, msg, opts...); err != nil {
			return err
		}
	}
	return nil
}

// Subscribe creates a subscription to a topic or queue.
func (a *InMemoryBrokerAdapter) Subscribe(ctx context.Context, topic string, handler messaging.MessageHandler, opts ...messaging.SubscribeOption) (messaging.Subscription, error) {
	genericHandler := func(ctx context.Context, msg *broker.Message) error {
		internalMsg := GenericToInternalMessage(msg)
		return handler(ctx, internalMsg)
	}

	sub, err := a.broker.Subscribe(ctx, topic, genericHandler)
	if err != nil {
		return nil, err
	}

	a.metrics.RecordSubscription()
	return &SubscriptionAdapter{sub: sub, metrics: a.metrics}, nil
}

// BrokerType returns the broker type.
func (a *InMemoryBrokerAdapter) BrokerType() messaging.BrokerType {
	return messaging.BrokerTypeInMemory
}

// GetMetrics returns broker metrics.
func (a *InMemoryBrokerAdapter) GetMetrics() *messaging.BrokerMetrics {
	return a.metrics
}

// Unwrap returns the underlying generic broker.
func (a *InMemoryBrokerAdapter) Unwrap() *broker.InMemoryBroker {
	return a.broker
}
