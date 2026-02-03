package messaging

import (
	"context"
	"time"

	"digital.vasic.messaging/pkg/broker"
	"digital.vasic.messaging/pkg/producer"

	"dev.helix.agent/internal/messaging"
)

// AsyncProducerAdapter wraps the generic producer.AsyncProducer for HelixAgent.
type AsyncProducerAdapter struct {
	ap      *producer.AsyncProducer
	metrics *messaging.BrokerMetrics
}

// NewAsyncProducerAdapter creates a new async producer adapter.
func NewAsyncProducerAdapter(b broker.MessageBroker, bufferSize int) *AsyncProducerAdapter {
	return &AsyncProducerAdapter{
		ap:      producer.NewAsyncProducer(b, bufferSize),
		metrics: messaging.NewBrokerMetrics(),
	}
}

// Start starts the background send loop.
func (a *AsyncProducerAdapter) Start(ctx context.Context) {
	a.ap.Start(ctx)
}

// Send queues a message for asynchronous sending.
func (a *AsyncProducerAdapter) Send(topic string, msg *messaging.Message) error {
	genericMsg := InternalToGenericMessage(msg)
	return a.ap.Send(topic, genericMsg)
}

// Errors returns the error channel for receiving send errors.
func (a *AsyncProducerAdapter) Errors() <-chan error {
	return a.ap.Errors()
}

// Stop stops the producer and waits for pending messages to drain.
func (a *AsyncProducerAdapter) Stop() {
	a.ap.Stop()
}

// SentCount returns the number of successfully sent messages.
func (a *AsyncProducerAdapter) SentCount() int64 {
	return a.ap.SentCount()
}

// FailedCount returns the number of failed messages.
func (a *AsyncProducerAdapter) FailedCount() int64 {
	return a.ap.FailedCount()
}

// SyncProducerAdapter wraps the generic producer.SyncProducer for HelixAgent.
type SyncProducerAdapter struct {
	sp      *producer.SyncProducer
	metrics *messaging.BrokerMetrics
}

// NewSyncProducerAdapter creates a new sync producer adapter.
func NewSyncProducerAdapter(b broker.MessageBroker, timeout time.Duration) *SyncProducerAdapter {
	return &SyncProducerAdapter{
		sp:      producer.NewSyncProducer(b, timeout),
		metrics: messaging.NewBrokerMetrics(),
	}
}

// Send publishes a message and waits for acknowledgment.
func (a *SyncProducerAdapter) Send(ctx context.Context, topic string, msg *messaging.Message) error {
	genericMsg := InternalToGenericMessage(msg)
	return a.sp.Send(ctx, topic, genericMsg)
}

// SendValue serializes a value and publishes it.
func (a *SyncProducerAdapter) SendValue(ctx context.Context, topic string, value interface{}) error {
	return a.sp.SendValue(ctx, topic, value)
}

// SentCount returns the number of successfully sent messages.
func (a *SyncProducerAdapter) SentCount() int64 {
	return a.sp.SentCount()
}

// FailedCount returns the number of failed messages.
func (a *SyncProducerAdapter) FailedCount() int64 {
	return a.sp.FailedCount()
}

// SetSerializer sets a serializer for the producer.
func (a *SyncProducerAdapter) SetSerializer(s producer.Serializer) {
	a.sp.SetSerializer(s)
}

// SetCompressor sets a compressor for the producer.
func (a *SyncProducerAdapter) SetCompressor(c producer.Compressor) {
	a.sp.SetCompressor(c)
}

// JSONSerializer is an alias for producer.JSONSerializer.
type JSONSerializer = producer.JSONSerializer

// GzipCompressor is an alias for producer.GzipCompressor.
type GzipCompressor = producer.GzipCompressor

// NewJSONSerializer creates a new JSON serializer.
func NewJSONSerializer() *JSONSerializer {
	return &JSONSerializer{}
}

// NewGzipCompressor creates a new gzip compressor.
func NewGzipCompressor(level int) *GzipCompressor {
	return &GzipCompressor{Level: level}
}
