package messaging

import (
	"time"
)

// PublishOptions holds options for publishing messages.
type PublishOptions struct {
	// Mandatory indicates the message must be routed to a queue.
	Mandatory bool
	// Immediate indicates the message must be delivered immediately.
	Immediate bool
	// Exchange is the exchange to publish to (RabbitMQ).
	Exchange string
	// RoutingKey is the routing key for message routing.
	RoutingKey string
	// ContentType is the MIME type of the message payload.
	ContentType string
	// ContentEncoding is the encoding of the message payload.
	ContentEncoding string
	// Timeout is the publish timeout.
	Timeout time.Duration
	// Confirm enables publisher confirms (RabbitMQ).
	Confirm bool
	// Partition specifies the target partition (Kafka).
	Partition *int32
	// Key is the message key for partition assignment (Kafka).
	Key []byte
	// Compression specifies the compression type (Kafka).
	Compression CompressionType
	// BatchSize is the number of messages to batch before sending (Kafka).
	BatchSize int
	// LingerMs is the time to wait for batching (Kafka).
	LingerMs int
}

// CompressionType represents the compression algorithm for messages.
type CompressionType int

const (
	// CompressionNone disables compression.
	CompressionNone CompressionType = iota
	// CompressionGzip uses gzip compression.
	CompressionGzip
	// CompressionSnappy uses snappy compression.
	CompressionSnappy
	// CompressionLZ4 uses lz4 compression.
	CompressionLZ4
	// CompressionZstd uses zstd compression.
	CompressionZstd
)

// String returns the string representation of CompressionType.
func (c CompressionType) String() string {
	switch c {
	case CompressionNone:
		return "none"
	case CompressionGzip:
		return "gzip"
	case CompressionSnappy:
		return "snappy"
	case CompressionLZ4:
		return "lz4"
	case CompressionZstd:
		return "zstd"
	default:
		return "unknown"
	}
}

// PublishOption is a function that modifies PublishOptions.
type PublishOption func(*PublishOptions)

// DefaultPublishOptions returns default publish options.
func DefaultPublishOptions() *PublishOptions {
	return &PublishOptions{
		Mandatory:       false,
		Immediate:       false,
		Exchange:        "",
		RoutingKey:      "",
		ContentType:     "application/json",
		ContentEncoding: "utf-8",
		Timeout:         30 * time.Second,
		Confirm:         true,
		Compression:     CompressionNone,
		BatchSize:       100,
		LingerMs:        10,
	}
}

// ApplyPublishOptions applies the given options to the default options.
func ApplyPublishOptions(opts ...PublishOption) *PublishOptions {
	options := DefaultPublishOptions()
	for _, opt := range opts {
		opt(options)
	}
	return options
}

// WithMandatory sets the mandatory flag.
func WithMandatory(mandatory bool) PublishOption {
	return func(o *PublishOptions) {
		o.Mandatory = mandatory
	}
}

// WithImmediate sets the immediate flag.
func WithImmediate(immediate bool) PublishOption {
	return func(o *PublishOptions) {
		o.Immediate = immediate
	}
}

// WithExchange sets the exchange name.
func WithExchange(exchange string) PublishOption {
	return func(o *PublishOptions) {
		o.Exchange = exchange
	}
}

// WithRoutingKey sets the routing key.
func WithRoutingKey(key string) PublishOption {
	return func(o *PublishOptions) {
		o.RoutingKey = key
	}
}

// WithContentType sets the content type.
func WithContentType(contentType string) PublishOption {
	return func(o *PublishOptions) {
		o.ContentType = contentType
	}
}

// WithContentEncoding sets the content encoding.
func WithContentEncoding(encoding string) PublishOption {
	return func(o *PublishOptions) {
		o.ContentEncoding = encoding
	}
}

// WithPublishTimeout sets the publish timeout.
func WithPublishTimeout(timeout time.Duration) PublishOption {
	return func(o *PublishOptions) {
		o.Timeout = timeout
	}
}

// WithConfirm enables or disables publisher confirms.
func WithConfirm(confirm bool) PublishOption {
	return func(o *PublishOptions) {
		o.Confirm = confirm
	}
}

// WithPartition sets the target partition for Kafka.
func WithPartition(partition int32) PublishOption {
	return func(o *PublishOptions) {
		o.Partition = &partition
	}
}

// WithMessageKey sets the message key for Kafka partitioning.
func WithMessageKey(key []byte) PublishOption {
	return func(o *PublishOptions) {
		o.Key = key
	}
}

// WithCompression sets the compression type.
func WithCompression(compression CompressionType) PublishOption {
	return func(o *PublishOptions) {
		o.Compression = compression
	}
}

// WithBatchSize sets the batch size for Kafka.
func WithBatchSize(size int) PublishOption {
	return func(o *PublishOptions) {
		o.BatchSize = size
	}
}

// WithLingerMs sets the linger time for Kafka batching.
func WithLingerMs(ms int) PublishOption {
	return func(o *PublishOptions) {
		o.LingerMs = ms
	}
}

// SubscribeOptions holds options for subscribing to messages.
type SubscribeOptions struct {
	// ConsumerTag is the consumer identifier (RabbitMQ).
	ConsumerTag string
	// AutoAck enables automatic acknowledgment.
	AutoAck bool
	// Exclusive makes this consumer exclusive to the queue (RabbitMQ).
	Exclusive bool
	// NoLocal prevents receiving messages from the same connection (RabbitMQ).
	NoLocal bool
	// NoWait doesn't wait for server confirmation (RabbitMQ).
	NoWait bool
	// QueueArgs are additional queue arguments (RabbitMQ).
	QueueArgs map[string]interface{}
	// Prefetch is the number of messages to prefetch (RabbitMQ).
	Prefetch int
	// PrefetchSize is the size of messages to prefetch (RabbitMQ).
	PrefetchSize int
	// GroupID is the consumer group ID (Kafka).
	GroupID string
	// SessionTimeout is the session timeout for consumer groups (Kafka).
	SessionTimeout time.Duration
	// HeartbeatInterval is the heartbeat interval (Kafka).
	HeartbeatInterval time.Duration
	// MaxPollRecords is the maximum records per poll (Kafka).
	MaxPollRecords int
	// OffsetReset is the offset reset policy (Kafka).
	OffsetReset OffsetReset
	// CommitInterval is the interval for auto-committing offsets (Kafka).
	CommitInterval time.Duration
	// Filter is a message filter function.
	Filter MessageFilter
	// BufferSize is the size of the message channel buffer.
	BufferSize int
	// RetryOnError enables automatic retry on handler error.
	RetryOnError bool
	// MaxRetries is the maximum number of retries.
	MaxRetries int
	// RetryDelay is the delay between retries.
	RetryDelay time.Duration
}

// OffsetReset specifies the offset reset policy for Kafka.
type OffsetReset string

const (
	// OffsetResetEarliest starts from the earliest offset.
	OffsetResetEarliest OffsetReset = "earliest"
	// OffsetResetLatest starts from the latest offset.
	OffsetResetLatest OffsetReset = "latest"
	// OffsetResetNone fails if no offset is found.
	OffsetResetNone OffsetReset = "none"
)

// SubscribeOption is a function that modifies SubscribeOptions.
type SubscribeOption func(*SubscribeOptions)

// DefaultSubscribeOptions returns default subscribe options.
func DefaultSubscribeOptions() *SubscribeOptions {
	return &SubscribeOptions{
		ConsumerTag:       "",
		AutoAck:           false,
		Exclusive:         false,
		NoLocal:           false,
		NoWait:            false,
		QueueArgs:         nil,
		Prefetch:          10,
		PrefetchSize:      0,
		GroupID:           "",
		SessionTimeout:    10 * time.Second,
		HeartbeatInterval: 3 * time.Second,
		MaxPollRecords:    500,
		OffsetReset:       OffsetResetLatest,
		CommitInterval:    5 * time.Second,
		Filter:            nil,
		BufferSize:        100,
		RetryOnError:      true,
		MaxRetries:        3,
		RetryDelay:        1 * time.Second,
	}
}

// ApplySubscribeOptions applies the given options to the default options.
func ApplySubscribeOptions(opts ...SubscribeOption) *SubscribeOptions {
	options := DefaultSubscribeOptions()
	for _, opt := range opts {
		opt(options)
	}
	return options
}

// WithConsumerTag sets the consumer tag.
func WithConsumerTag(tag string) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.ConsumerTag = tag
	}
}

// WithAutoAck enables or disables automatic acknowledgment.
func WithAutoAck(autoAck bool) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.AutoAck = autoAck
	}
}

// WithExclusive makes the consumer exclusive.
func WithExclusive(exclusive bool) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.Exclusive = exclusive
	}
}

// WithNoLocal prevents receiving messages from the same connection.
func WithNoLocal(noLocal bool) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.NoLocal = noLocal
	}
}

// WithNoWait doesn't wait for server confirmation.
func WithNoWait(noWait bool) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.NoWait = noWait
	}
}

// WithQueueArgs sets additional queue arguments.
func WithQueueArgs(args map[string]interface{}) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.QueueArgs = args
	}
}

// WithPrefetch sets the prefetch count.
func WithPrefetch(prefetch int) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.Prefetch = prefetch
	}
}

// WithPrefetchSize sets the prefetch size.
func WithPrefetchSize(size int) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.PrefetchSize = size
	}
}

// WithGroupID sets the consumer group ID.
func WithGroupID(groupID string) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.GroupID = groupID
	}
}

// WithSessionTimeout sets the session timeout.
func WithSessionTimeout(timeout time.Duration) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.SessionTimeout = timeout
	}
}

// WithHeartbeatInterval sets the heartbeat interval.
func WithHeartbeatInterval(interval time.Duration) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.HeartbeatInterval = interval
	}
}

// WithMaxPollRecords sets the maximum poll records.
func WithMaxPollRecords(max int) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.MaxPollRecords = max
	}
}

// WithOffsetReset sets the offset reset policy.
func WithOffsetReset(reset OffsetReset) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.OffsetReset = reset
	}
}

// WithCommitInterval sets the commit interval.
func WithCommitInterval(interval time.Duration) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.CommitInterval = interval
	}
}

// WithMessageFilter sets the message filter.
func WithMessageFilter(filter MessageFilter) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.Filter = filter
	}
}

// WithBufferSize sets the message buffer size.
func WithBufferSize(size int) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.BufferSize = size
	}
}

// WithRetryOnError enables or disables retry on error.
func WithRetryOnError(retry bool) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.RetryOnError = retry
	}
}

// WithMaxSubscribeRetries sets the maximum number of retries.
func WithMaxSubscribeRetries(max int) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.MaxRetries = max
	}
}

// WithRetryDelay sets the delay between retries.
func WithRetryDelay(delay time.Duration) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.RetryDelay = delay
	}
}

// QueueOptions holds options for declaring queues (RabbitMQ).
type QueueOptions struct {
	// Durable makes the queue survive broker restart.
	Durable bool
	// AutoDelete deletes the queue when no consumers are connected.
	AutoDelete bool
	// Exclusive makes the queue exclusive to the connection.
	Exclusive bool
	// NoWait doesn't wait for server confirmation.
	NoWait bool
	// Args are additional queue arguments.
	Args map[string]interface{}
	// DeadLetterExchange is the exchange for dead-lettered messages.
	DeadLetterExchange string
	// DeadLetterRoutingKey is the routing key for dead-lettered messages.
	DeadLetterRoutingKey string
	// MessageTTL is the default message TTL in milliseconds.
	MessageTTL int64
	// MaxLength is the maximum queue length.
	MaxLength int64
	// MaxLengthBytes is the maximum queue size in bytes.
	MaxLengthBytes int64
	// MaxPriority enables priority queue (0-255).
	MaxPriority *int
}

// QueueOption is a function that modifies QueueOptions.
type QueueOption func(*QueueOptions)

// DefaultQueueOptions returns default queue options.
func DefaultQueueOptions() *QueueOptions {
	return &QueueOptions{
		Durable:              true,
		AutoDelete:           false,
		Exclusive:            false,
		NoWait:               false,
		Args:                 nil,
		DeadLetterExchange:   "",
		DeadLetterRoutingKey: "",
		MessageTTL:           0,
		MaxLength:            0,
		MaxLengthBytes:       0,
		MaxPriority:          nil,
	}
}

// ApplyQueueOptions applies the given options to the default options.
func ApplyQueueOptions(opts ...QueueOption) *QueueOptions {
	options := DefaultQueueOptions()
	for _, opt := range opts {
		opt(options)
	}
	return options
}

// WithDurable makes the queue durable.
func WithDurable(durable bool) QueueOption {
	return func(o *QueueOptions) {
		o.Durable = durable
	}
}

// WithAutoDelete enables auto-delete.
func WithAutoDelete(autoDelete bool) QueueOption {
	return func(o *QueueOptions) {
		o.AutoDelete = autoDelete
	}
}

// WithQueueExclusive makes the queue exclusive.
func WithQueueExclusive(exclusive bool) QueueOption {
	return func(o *QueueOptions) {
		o.Exclusive = exclusive
	}
}

// WithQueueNoWait doesn't wait for server confirmation.
func WithQueueNoWait(noWait bool) QueueOption {
	return func(o *QueueOptions) {
		o.NoWait = noWait
	}
}

// WithDeadLetterExchange sets the dead letter exchange.
func WithDeadLetterExchange(exchange string) QueueOption {
	return func(o *QueueOptions) {
		o.DeadLetterExchange = exchange
	}
}

// WithDeadLetterRoutingKey sets the dead letter routing key.
func WithDeadLetterRoutingKey(key string) QueueOption {
	return func(o *QueueOptions) {
		o.DeadLetterRoutingKey = key
	}
}

// WithMessageTTL sets the message TTL.
func WithMessageTTL(ttl time.Duration) QueueOption {
	return func(o *QueueOptions) {
		o.MessageTTL = int64(ttl.Milliseconds())
	}
}

// WithMaxLength sets the maximum queue length.
func WithMaxLength(max int64) QueueOption {
	return func(o *QueueOptions) {
		o.MaxLength = max
	}
}

// WithMaxLengthBytes sets the maximum queue size in bytes.
func WithMaxLengthBytes(max int64) QueueOption {
	return func(o *QueueOptions) {
		o.MaxLengthBytes = max
	}
}

// WithMaxPriority enables priority queue.
func WithMaxPriority(max int) QueueOption {
	return func(o *QueueOptions) {
		o.MaxPriority = &max
	}
}

// StreamOptions holds options for streaming messages (Kafka).
type StreamOptions struct {
	// StartOffset is the starting offset.
	StartOffset int64
	// StartTime is the starting timestamp.
	StartTime *time.Time
	// BufferSize is the channel buffer size.
	BufferSize int
	// MinBytes is the minimum bytes to fetch.
	MinBytes int
	// MaxBytes is the maximum bytes to fetch.
	MaxBytes int
	// MaxWait is the maximum wait time for new messages.
	MaxWait time.Duration
	// Partition specifies a specific partition to stream from.
	Partition *int32
}

// StreamOption is a function that modifies StreamOptions.
type StreamOption func(*StreamOptions)

// DefaultStreamOptions returns default stream options.
func DefaultStreamOptions() *StreamOptions {
	return &StreamOptions{
		StartOffset: -1, // Latest
		StartTime:   nil,
		BufferSize:  1000,
		MinBytes:    1,
		MaxBytes:    10 * 1024 * 1024, // 10MB
		MaxWait:     500 * time.Millisecond,
		Partition:   nil,
	}
}

// ApplyStreamOptions applies the given options to the default options.
func ApplyStreamOptions(opts ...StreamOption) *StreamOptions {
	options := DefaultStreamOptions()
	for _, opt := range opts {
		opt(options)
	}
	return options
}

// WithStartOffset sets the starting offset.
func WithStartOffset(offset int64) StreamOption {
	return func(o *StreamOptions) {
		o.StartOffset = offset
	}
}

// WithStartTime sets the starting timestamp.
func WithStartTime(t time.Time) StreamOption {
	return func(o *StreamOptions) {
		o.StartTime = &t
	}
}

// WithStreamBufferSize sets the buffer size.
func WithStreamBufferSize(size int) StreamOption {
	return func(o *StreamOptions) {
		o.BufferSize = size
	}
}

// WithMinBytes sets the minimum bytes to fetch.
func WithMinBytes(min int) StreamOption {
	return func(o *StreamOptions) {
		o.MinBytes = min
	}
}

// WithMaxBytes sets the maximum bytes to fetch.
func WithMaxBytes(max int) StreamOption {
	return func(o *StreamOptions) {
		o.MaxBytes = max
	}
}

// WithMaxWait sets the maximum wait time.
func WithMaxWait(wait time.Duration) StreamOption {
	return func(o *StreamOptions) {
		o.MaxWait = wait
	}
}

// WithStreamPartition sets the partition to stream from.
func WithStreamPartition(partition int32) StreamOption {
	return func(o *StreamOptions) {
		o.Partition = &partition
	}
}
