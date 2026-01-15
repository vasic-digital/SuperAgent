// Package rabbitmq provides a RabbitMQ message broker implementation.
package rabbitmq

import (
	"context"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"dev.helix.agent/internal/messaging"
)

// Subscription represents a RabbitMQ subscription.
type Subscription struct {
	id         string
	topic      string
	broker     *Broker
	channel    *amqp.Channel
	deliveries <-chan amqp.Delivery
	handler    messaging.MessageHandler
	options    *messaging.SubscribeOptions
	active     bool
	stopCh     chan struct{}
	mu         sync.RWMutex
}

// ID returns the subscription ID.
func (s *Subscription) ID() string {
	return s.id
}

// Topic returns the subscribed topic.
func (s *Subscription) Topic() string {
	return s.topic
}

// IsActive returns true if the subscription is active.
func (s *Subscription) IsActive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.active
}

// Unsubscribe cancels the subscription.
func (s *Subscription) Unsubscribe() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.active {
		return nil
	}

	s.active = false
	close(s.stopCh)

	// Cancel consumer
	if s.channel != nil {
		if err := s.channel.Cancel(s.options.ConsumerTag, false); err != nil {
			// Log but don't return error - channel might already be closed
		}
		if err := s.channel.Close(); err != nil {
			// Log but don't return error
		}
	}

	// Remove from broker's subscriber list
	s.broker.mu.Lock()
	delete(s.broker.subscribers, s.id)
	s.broker.metrics.RecordUnsubscription()
	s.broker.mu.Unlock()

	return nil
}

// consume processes incoming deliveries.
func (s *Subscription) consume(ctx context.Context) {
	for {
		select {
		case <-s.stopCh:
			return
		case <-ctx.Done():
			s.cancel()
			return
		case delivery, ok := <-s.deliveries:
			if !ok {
				s.handleChannelClosed()
				return
			}

			s.processDelivery(ctx, delivery)
		}
	}
}

// processDelivery processes a single delivery.
func (s *Subscription) processDelivery(ctx context.Context, delivery amqp.Delivery) {
	start := time.Now()

	// Convert AMQP delivery to Message
	msg := s.deliveryToMessage(delivery)

	// Record message consumption
	s.broker.metrics.RecordConsume(int64(len(delivery.Body)), time.Since(start), true)

	// Process with handler
	var err error
	if s.options.Timeout > 0 {
		handlerCtx, cancel := context.WithTimeout(ctx, s.options.Timeout)
		err = s.handler(handlerCtx, msg)
		cancel()
	} else {
		err = s.handler(ctx, msg)
	}

	// Handle result
	if err != nil {
		s.handleError(ctx, delivery, msg, err)
	} else {
		s.handleSuccess(ctx, delivery, msg)
	}
}

// deliveryToMessage converts an AMQP delivery to a Message.
func (s *Subscription) deliveryToMessage(delivery amqp.Delivery) *messaging.Message {
	msg := &messaging.Message{
		ID:            delivery.MessageId,
		Type:          delivery.Type,
		Payload:       delivery.Body,
		Headers:       make(map[string]string),
		Timestamp:     delivery.Timestamp,
		Priority:      messaging.MessagePriority(delivery.Priority),
		DeliveryMode:  messaging.DeliveryMode(delivery.DeliveryMode),
		CorrelationID: delivery.CorrelationId,
		ReplyTo:       delivery.ReplyTo,
		DeliveryTag:   delivery.DeliveryTag,
		Redelivered:   delivery.Redelivered,
	}

	// Convert headers
	for k, v := range delivery.Headers {
		if str, ok := v.(string); ok {
			msg.Headers[k] = str
		}
	}

	// Extract trace ID from headers
	if traceID, ok := delivery.Headers["trace_id"]; ok {
		if str, ok := traceID.(string); ok {
			msg.TraceID = str
		}
	}

	return msg
}

// handleSuccess handles successful message processing.
func (s *Subscription) handleSuccess(ctx context.Context, delivery amqp.Delivery, msg *messaging.Message) {
	if !s.options.AutoAck {
		if err := delivery.Ack(false); err != nil {
			s.broker.metrics.RecordError()
		} else {
			s.broker.metrics.RecordAck()
		}
	}
}

// handleError handles message processing errors.
func (s *Subscription) handleError(ctx context.Context, delivery amqp.Delivery, msg *messaging.Message, err error) {
	s.broker.metrics.RecordError()

	// Check if retry is enabled
	if s.options.RetryOnError && msg.RetryCount < s.options.MaxRetries {
		msg.RetryCount++

		// Retry delay
		if s.options.RetryDelay > 0 {
			time.Sleep(s.options.RetryDelay)
		}

		// Requeue for retry
		if !s.options.AutoAck {
			if err := delivery.Nack(false, true); err != nil {
				s.broker.metrics.RecordError()
			} else {
				s.broker.metrics.RecordNack()
			}
		}
		return
	}

	// No retry or max retries exceeded
	if !s.options.AutoAck {
		// Don't requeue - send to DLQ
		if err := delivery.Nack(false, false); err != nil {
			s.broker.metrics.RecordError()
		} else {
			s.broker.metrics.RecordNack()
		}
	}
}

// handleChannelClosed handles when the channel is closed.
func (s *Subscription) handleChannelClosed() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.active {
		return
	}

	// Channel closed - mark as inactive but don't close stopCh
	// The broker's reconnection handler will restart this subscription
}

// cancel cancels the subscription.
func (s *Subscription) cancel() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.active {
		return
	}

	s.active = false

	select {
	case <-s.stopCh:
		// Already closed
	default:
		close(s.stopCh)
	}
}

// restart restarts the subscription after reconnection.
func (s *Subscription) restart(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a new channel
	ch, err := s.broker.conn.Channel()
	if err != nil {
		return err
	}

	// Set QoS
	if err := ch.Qos(s.options.Prefetch, s.options.PrefetchSize, false); err != nil {
		ch.Close()
		return err
	}

	// Re-declare queue
	if _, err := ch.QueueDeclare(
		s.topic,
		s.broker.config.DefaultQueueDurable,
		s.broker.config.DefaultQueueAutoDelete,
		s.broker.config.DefaultQueueExclusive,
		false,
		nil,
	); err != nil {
		ch.Close()
		return err
	}

	// Start consuming
	deliveries, err := ch.Consume(
		s.topic,
		s.options.ConsumerTag,
		s.options.AutoAck,
		s.options.Exclusive,
		s.options.NoLocal,
		s.options.NoWait,
		nil,
	)
	if err != nil {
		ch.Close()
		return err
	}

	// Update subscription
	s.channel = ch
	s.deliveries = deliveries
	s.stopCh = make(chan struct{})
	s.active = true

	// Start consuming in background
	go s.consume(ctx)

	return nil
}
