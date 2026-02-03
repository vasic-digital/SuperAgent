// Package messaging provides adapters bridging HelixAgent's internal messaging
// types to the generic digital.vasic.messaging module.
//
// The adapter package maintains backward compatibility with code using
// dev.helix.agent/internal/messaging while delegating core broker operations
// to digital.vasic.messaging.
package messaging

import (
	"context"

	"digital.vasic.messaging/pkg/broker"

	"dev.helix.agent/internal/messaging"
)

// BrokerAdapter wraps the generic broker.MessageBroker to implement
// HelixAgent's messaging.MessageBroker interface.
type BrokerAdapter struct {
	broker  broker.MessageBroker
	metrics *messaging.BrokerMetrics
}

// NewBrokerAdapter creates a new adapter wrapping a generic broker.
func NewBrokerAdapter(b broker.MessageBroker) *BrokerAdapter {
	return &BrokerAdapter{
		broker:  b,
		metrics: messaging.NewBrokerMetrics(),
	}
}

// Connect establishes a connection to the broker.
func (a *BrokerAdapter) Connect(ctx context.Context) error {
	a.metrics.RecordConnectionAttempt()
	if err := a.broker.Connect(ctx); err != nil {
		a.metrics.RecordConnectionFailure()
		return err
	}
	a.metrics.RecordConnectionSuccess()
	return nil
}

// Close closes the connection to the broker.
func (a *BrokerAdapter) Close(ctx context.Context) error {
	a.metrics.RecordDisconnection()
	return a.broker.Close(ctx)
}

// HealthCheck checks if the broker is healthy.
func (a *BrokerAdapter) HealthCheck(ctx context.Context) error {
	return a.broker.HealthCheck(ctx)
}

// IsConnected returns true if connected to the broker.
func (a *BrokerAdapter) IsConnected() bool {
	return a.broker.IsConnected()
}

// Publish sends a message to a topic or queue.
func (a *BrokerAdapter) Publish(ctx context.Context, topic string, message *messaging.Message, opts ...messaging.PublishOption) error {
	// Convert internal message to generic message
	genericMsg := InternalToGenericMessage(message)

	if err := a.broker.Publish(ctx, topic, genericMsg); err != nil {
		a.metrics.RecordPublish(int64(len(message.Payload)), 0, false)
		return err
	}

	a.metrics.RecordPublish(int64(len(message.Payload)), 0, true)
	return nil
}

// PublishBatch sends multiple messages to a topic or queue.
func (a *BrokerAdapter) PublishBatch(ctx context.Context, topic string, messages []*messaging.Message, opts ...messaging.PublishOption) error {
	for _, msg := range messages {
		if err := a.Publish(ctx, topic, msg, opts...); err != nil {
			return err
		}
	}
	return nil
}

// Subscribe creates a subscription to a topic or queue.
func (a *BrokerAdapter) Subscribe(ctx context.Context, topic string, handler messaging.MessageHandler, opts ...messaging.SubscribeOption) (messaging.Subscription, error) {
	// Wrap the internal handler to work with generic messages
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

// BrokerType returns the type of this broker.
func (a *BrokerAdapter) BrokerType() messaging.BrokerType {
	bt := a.broker.Type()
	switch bt {
	case broker.BrokerTypeKafka:
		return messaging.BrokerTypeKafka
	case broker.BrokerTypeRabbitMQ:
		return messaging.BrokerTypeRabbitMQ
	case broker.BrokerTypeInMemory:
		return messaging.BrokerTypeInMemory
	default:
		return messaging.BrokerTypeInMemory
	}
}

// GetMetrics returns broker metrics.
func (a *BrokerAdapter) GetMetrics() *messaging.BrokerMetrics {
	return a.metrics
}

// Unwrap returns the underlying generic broker.
func (a *BrokerAdapter) Unwrap() broker.MessageBroker {
	return a.broker
}

// SubscriptionAdapter wraps a generic subscription.
type SubscriptionAdapter struct {
	sub     broker.Subscription
	metrics *messaging.BrokerMetrics
}

// Unsubscribe cancels the subscription.
func (s *SubscriptionAdapter) Unsubscribe() error {
	s.metrics.RecordUnsubscription()
	return s.sub.Unsubscribe()
}

// IsActive returns true if the subscription is still active.
func (s *SubscriptionAdapter) IsActive() bool {
	return s.sub.IsActive()
}

// Topic returns the subscribed topic or queue name.
func (s *SubscriptionAdapter) Topic() string {
	return s.sub.Topic()
}

// ID returns the subscription identifier.
func (s *SubscriptionAdapter) ID() string {
	return s.sub.ID()
}
