package messaging

import (
	"context"
	"time"

	"digital.vasic.messaging/pkg/broker"
	"digital.vasic.messaging/pkg/consumer"

	"dev.helix.agent/internal/messaging"
)

// ConsumerGroupAdapter wraps the generic consumer.ConsumerGroup for HelixAgent.
type ConsumerGroupAdapter struct {
	cg      *consumer.ConsumerGroup
	metrics *messaging.BrokerMetrics
}

// NewConsumerGroupAdapter creates a new consumer group adapter.
func NewConsumerGroupAdapter(id string, b broker.MessageBroker) *ConsumerGroupAdapter {
	return &ConsumerGroupAdapter{
		cg:      consumer.NewConsumerGroup(id, b),
		metrics: messaging.NewBrokerMetrics(),
	}
}

// ID returns the consumer group identifier.
func (a *ConsumerGroupAdapter) ID() string {
	return a.cg.ID()
}

// Add registers a handler for a topic.
func (a *ConsumerGroupAdapter) Add(topic string, handler messaging.MessageHandler) {
	genericHandler := InternalToGenericHandler(handler)
	a.cg.Add(topic, genericHandler)
}

// Start subscribes to all registered topics.
func (a *ConsumerGroupAdapter) Start(ctx context.Context) error {
	return a.cg.Start(ctx)
}

// Stop unsubscribes from all topics.
func (a *ConsumerGroupAdapter) Stop() error {
	return a.cg.Stop()
}

// IsRunning returns true if the consumer group is running.
func (a *ConsumerGroupAdapter) IsRunning() bool {
	return a.cg.IsRunning()
}

// Topics returns the list of subscribed topics.
func (a *ConsumerGroupAdapter) Topics() []string {
	return a.cg.Topics()
}

// RetryPolicyAdapter wraps the generic consumer.RetryPolicy.
type RetryPolicyAdapter struct {
	rp *consumer.RetryPolicy
}

// NewRetryPolicyAdapter creates a new retry policy adapter.
func NewRetryPolicyAdapter() *RetryPolicyAdapter {
	return &RetryPolicyAdapter{
		rp: consumer.DefaultRetryPolicy(),
	}
}

// NewRetryPolicyAdapterWithConfig creates a retry policy adapter with custom config.
func NewRetryPolicyAdapterWithConfig(maxRetries int, backoffBase, backoffMax time.Duration, multiplier float64) *RetryPolicyAdapter {
	return &RetryPolicyAdapter{
		rp: &consumer.RetryPolicy{
			MaxRetries:        maxRetries,
			BackoffBase:       backoffBase,
			BackoffMax:        backoffMax,
			BackoffMultiplier: multiplier,
		},
	}
}

// Delay calculates the delay for the given attempt number.
func (a *RetryPolicyAdapter) Delay(attempt int) time.Duration {
	return a.rp.Delay(attempt)
}

// ShouldRetry returns true if the given attempt is within the retry limit.
func (a *RetryPolicyAdapter) ShouldRetry(attempt int) bool {
	return a.rp.ShouldRetry(attempt)
}

// Unwrap returns the underlying retry policy.
func (a *RetryPolicyAdapter) Unwrap() *consumer.RetryPolicy {
	return a.rp
}

// WithRetryAdapter wraps a handler with retry logic.
func WithRetryAdapter(handler messaging.MessageHandler, policy *RetryPolicyAdapter) messaging.MessageHandler {
	if policy == nil {
		policy = NewRetryPolicyAdapter()
	}

	genericHandler := InternalToGenericHandler(handler)
	retryHandler := consumer.WithRetry(genericHandler, policy.rp)
	return GenericToInternalHandler(retryHandler)
}

// DeadLetterHandlerAdapter wraps the generic consumer.DeadLetterHandler.
type DeadLetterHandlerAdapter struct {
	dlh     *consumer.DeadLetterHandler
	metrics *messaging.BrokerMetrics
}

// NewDeadLetterHandlerAdapter creates a new dead letter handler adapter.
func NewDeadLetterHandlerAdapter(b broker.MessageBroker, dlqTopic string) *DeadLetterHandlerAdapter {
	return &DeadLetterHandlerAdapter{
		dlh:     consumer.NewDeadLetterHandler(b, dlqTopic),
		metrics: messaging.NewBrokerMetrics(),
	}
}

// SetOnFailure sets a callback for when a message is sent to the DLQ.
func (a *DeadLetterHandlerAdapter) SetOnFailure(fn func(ctx context.Context, msg *messaging.Message, err error)) {
	genericFn := func(ctx context.Context, msg *broker.Message, err error) {
		internalMsg := GenericToInternalMessage(msg)
		fn(ctx, internalMsg, err)
	}
	a.dlh.SetOnFailure(genericFn)
}

// Handle sends a failed message to the dead letter queue.
func (a *DeadLetterHandlerAdapter) Handle(ctx context.Context, msg *messaging.Message, originalErr error) error {
	genericMsg := InternalToGenericMessage(msg)
	return a.dlh.Handle(ctx, genericMsg, originalErr)
}

// Count returns the number of messages sent to the DLQ.
func (a *DeadLetterHandlerAdapter) Count() int64 {
	return a.dlh.Count()
}

// DLQTopic returns the dead letter queue topic.
func (a *DeadLetterHandlerAdapter) DLQTopic() string {
	return a.dlh.DLQTopic()
}

// BatchConsumerAdapter wraps the generic consumer.BatchConsumer.
type BatchConsumerAdapter struct {
	bc      *consumer.BatchConsumer
	metrics *messaging.BrokerMetrics
}

// NewBatchConsumerAdapter creates a new batch consumer adapter.
func NewBatchConsumerAdapter(
	batchSize int,
	flushAfter time.Duration,
	handler func(ctx context.Context, msgs []*messaging.Message) error,
) *BatchConsumerAdapter {
	genericHandler := func(ctx context.Context, msgs []*broker.Message) error {
		internalMsgs := InternalMessageBatch(msgs)
		return handler(ctx, internalMsgs)
	}

	return &BatchConsumerAdapter{
		bc:      consumer.NewBatchConsumer(batchSize, flushAfter, genericHandler),
		metrics: messaging.NewBrokerMetrics(),
	}
}

// Add adds a message to the batch buffer.
func (a *BatchConsumerAdapter) Add(msg *messaging.Message) {
	genericMsg := InternalToGenericMessage(msg)
	a.bc.Add(genericMsg)
}

// Flush processes all buffered messages immediately.
func (a *BatchConsumerAdapter) Flush(ctx context.Context) error {
	return a.bc.Flush(ctx)
}

// Start starts the background flush loop.
func (a *BatchConsumerAdapter) Start(ctx context.Context) {
	a.bc.Start(ctx)
}

// Stop stops the background flush loop.
func (a *BatchConsumerAdapter) Stop(ctx context.Context) error {
	return a.bc.Stop(ctx)
}

// BufferLen returns the current number of buffered messages.
func (a *BatchConsumerAdapter) BufferLen() int {
	return a.bc.BufferLen()
}

// BatchSize returns the configured batch size.
func (a *BatchConsumerAdapter) BatchSize() int {
	return a.bc.BatchSize()
}

// AsHandler returns an internal MessageHandler that adds messages to the batch.
func (a *BatchConsumerAdapter) AsHandler() messaging.MessageHandler {
	return func(ctx context.Context, msg *messaging.Message) error {
		a.Add(msg)
		return nil
	}
}
