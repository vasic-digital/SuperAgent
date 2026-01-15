package rabbitmq

import (
	"context"
	"crypto/tls"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"dev.helix.agent/internal/messaging"
)

// Broker is a RabbitMQ message broker implementation.
type Broker struct {
	config      *Config
	conn        *amqp.Connection
	pubChannel  *amqp.Channel
	confirms    chan amqp.Confirmation
	metrics     *messaging.BrokerMetrics
	connected   bool
	reconnecting bool
	mu          sync.RWMutex
	stopCh      chan struct{}
	closeCh     chan *amqp.Error
	subscribers map[string]*Subscription
	topology    *TopologyConfig
}

// NewBroker creates a new RabbitMQ broker.
func NewBroker(config *Config) *Broker {
	if config == nil {
		config = DefaultConfig()
	}
	return &Broker{
		config:      config,
		metrics:     messaging.NewBrokerMetrics(),
		subscribers: make(map[string]*Subscription),
		stopCh:      make(chan struct{}),
		topology:    DefaultTopologyConfig(),
	}
}

// Connect establishes a connection to RabbitMQ.
func (b *Broker) Connect(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.connected {
		return nil
	}

	b.metrics.RecordConnectionAttempt()

	if err := b.config.Validate(); err != nil {
		b.metrics.RecordConnectionFailure()
		return err
	}

	// Create AMQP config
	amqpConfig := amqp.Config{
		Heartbeat: b.config.Heartbeat,
		Locale:    "en_US",
	}

	if b.config.ChannelMax > 0 {
		amqpConfig.ChannelMax = uint16(b.config.ChannelMax)
	}
	if b.config.FrameSize > 0 {
		amqpConfig.FrameSize = b.config.FrameSize
	}

	// Configure TLS if enabled
	if b.config.TLS {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: b.config.TLSInsecure,
		}
		amqpConfig.TLSClientConfig = tlsConfig
	}

	// Connect with timeout
	connCh := make(chan *amqp.Connection, 1)
	errCh := make(chan error, 1)

	go func() {
		conn, err := amqp.DialConfig(b.config.URL(), amqpConfig)
		if err != nil {
			errCh <- err
			return
		}
		connCh <- conn
	}()

	select {
	case conn := <-connCh:
		b.conn = conn
	case err := <-errCh:
		b.metrics.RecordConnectionFailure()
		return messaging.ConnectionError("failed to connect to RabbitMQ", err)
	case <-ctx.Done():
		b.metrics.RecordConnectionFailure()
		return messaging.ConnectionTimeoutError(ctx.Err())
	case <-time.After(b.config.ConnectTimeout):
		b.metrics.RecordConnectionFailure()
		return messaging.ConnectionTimeoutError(nil)
	}

	// Create publisher channel
	pubChannel, err := b.conn.Channel()
	if err != nil {
		b.conn.Close()
		b.metrics.RecordConnectionFailure()
		return messaging.ConnectionError("failed to create publisher channel", err)
	}
	b.pubChannel = pubChannel

	// Enable publisher confirms if configured
	if b.config.PublisherConfirm {
		if err := b.pubChannel.Confirm(false); err != nil {
			b.conn.Close()
			b.metrics.RecordConnectionFailure()
			return messaging.ConnectionError("failed to enable publisher confirms", err)
		}
		b.confirms = b.pubChannel.NotifyPublish(make(chan amqp.Confirmation, 100))
	}

	// Set up QoS
	if err := b.pubChannel.Qos(
		b.config.PrefetchCount,
		b.config.PrefetchSize,
		b.config.GlobalQos,
	); err != nil {
		b.conn.Close()
		b.metrics.RecordConnectionFailure()
		return messaging.ConnectionError("failed to set QoS", err)
	}

	// Set up topology
	if err := b.setupTopology(); err != nil {
		b.conn.Close()
		b.metrics.RecordConnectionFailure()
		return err
	}

	// Set up connection close notification
	b.closeCh = make(chan *amqp.Error, 1)
	b.conn.NotifyClose(b.closeCh)

	b.connected = true
	b.metrics.RecordConnectionSuccess()

	// Start reconnection handler
	go b.handleReconnection()

	return nil
}

// setupTopology creates exchanges, queues, and bindings.
func (b *Broker) setupTopology() error {
	// Create exchanges
	for _, ex := range b.topology.Exchanges {
		if err := b.pubChannel.ExchangeDeclare(
			ex.Name,
			ex.Type.String(),
			ex.Durable,
			ex.AutoDelete,
			ex.Internal,
			ex.NoWait,
			ex.Args,
		); err != nil {
			return messaging.TopicError(ex.Name, err)
		}
	}

	// Create queues
	for _, q := range b.topology.Queues {
		args := q.ToArgs()
		if _, err := b.pubChannel.QueueDeclare(
			q.Name,
			q.Durable,
			q.AutoDelete,
			q.Exclusive,
			q.NoWait,
			args,
		); err != nil {
			return messaging.QueueError(q.Name, err)
		}
		b.metrics.RecordQueueDeclared()
	}

	// Create bindings
	for _, bind := range b.topology.Bindings {
		if err := b.pubChannel.QueueBind(
			bind.Queue,
			bind.RoutingKey,
			bind.Exchange,
			bind.NoWait,
			bind.Args,
		); err != nil {
			return messaging.QueueError(bind.Queue, err)
		}
	}

	return nil
}

// handleReconnection handles automatic reconnection.
func (b *Broker) handleReconnection() {
	for {
		select {
		case <-b.stopCh:
			return
		case amqpErr, ok := <-b.closeCh:
			if !ok {
				return
			}

			b.mu.Lock()
			b.connected = false
			b.reconnecting = true
			b.mu.Unlock()

			if amqpErr != nil {
				b.metrics.RecordConnectionFailure()
			}

			// Attempt reconnection
			attempts := 0
			for {
				select {
				case <-b.stopCh:
					return
				case <-time.After(b.config.ReconnectInterval):
					b.metrics.RecordReconnectionAttempt()
					attempts++

					if b.config.MaxReconnectAttempts > 0 && attempts > b.config.MaxReconnectAttempts {
						return
					}

					ctx, cancel := context.WithTimeout(context.Background(), b.config.ConnectTimeout)
					err := b.reconnect(ctx)
					cancel()

					if err == nil {
						b.mu.Lock()
						b.connected = true
						b.reconnecting = false
						b.mu.Unlock()

						// Re-subscribe existing subscriptions
						b.resubscribe()
						goto reconnected
					}
				}
			}
		reconnected:
		}
	}
}

// reconnect attempts to reconnect to RabbitMQ.
func (b *Broker) reconnect(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Create AMQP config
	amqpConfig := amqp.Config{
		Heartbeat: b.config.Heartbeat,
		Locale:    "en_US",
	}

	if b.config.TLS {
		amqpConfig.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: b.config.TLSInsecure,
		}
	}

	conn, err := amqp.DialConfig(b.config.URL(), amqpConfig)
	if err != nil {
		return err
	}
	b.conn = conn

	pubChannel, err := b.conn.Channel()
	if err != nil {
		b.conn.Close()
		return err
	}
	b.pubChannel = pubChannel

	if b.config.PublisherConfirm {
		if err := b.pubChannel.Confirm(false); err != nil {
			b.conn.Close()
			return err
		}
		b.confirms = b.pubChannel.NotifyPublish(make(chan amqp.Confirmation, 100))
	}

	if err := b.pubChannel.Qos(
		b.config.PrefetchCount,
		b.config.PrefetchSize,
		b.config.GlobalQos,
	); err != nil {
		b.conn.Close()
		return err
	}

	// Re-setup topology
	if err := b.setupTopology(); err != nil {
		b.conn.Close()
		return err
	}

	// Set up new close notification
	b.closeCh = make(chan *amqp.Error, 1)
	b.conn.NotifyClose(b.closeCh)

	b.metrics.RecordConnectionSuccess()
	return nil
}

// resubscribe re-subscribes all existing subscriptions.
func (b *Broker) resubscribe() {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, sub := range b.subscribers {
		if sub.active {
			go func(s *Subscription) {
				ctx := context.Background()
				if err := s.restart(ctx); err != nil {
					b.metrics.RecordError()
				}
			}(sub)
		}
	}
}

// Close closes the connection to RabbitMQ.
func (b *Broker) Close(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.connected && b.conn == nil {
		return nil
	}

	close(b.stopCh)

	// Cancel all subscriptions
	for _, sub := range b.subscribers {
		sub.cancel()
	}
	b.subscribers = make(map[string]*Subscription)

	// Close channels and connection
	var errs messaging.MultiError
	if b.pubChannel != nil {
		if err := b.pubChannel.Close(); err != nil {
			errs.Add(err)
		}
	}
	if b.conn != nil {
		if err := b.conn.Close(); err != nil {
			errs.Add(err)
		}
	}

	b.connected = false
	b.metrics.RecordDisconnection()

	return errs.ErrorOrNil()
}

// HealthCheck checks if the broker is healthy.
func (b *Broker) HealthCheck(ctx context.Context) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.connected {
		return messaging.ErrNotConnected
	}

	if b.conn == nil || b.conn.IsClosed() {
		return messaging.ErrConnectionClosed
	}

	return nil
}

// IsConnected returns true if connected to RabbitMQ.
func (b *Broker) IsConnected() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.connected && !b.reconnecting
}

// Publish publishes a message to a queue or exchange.
func (b *Broker) Publish(ctx context.Context, topic string, message *messaging.Message, opts ...messaging.PublishOption) error {
	b.mu.RLock()
	if !b.connected {
		b.mu.RUnlock()
		return messaging.ErrNotConnected
	}
	pubChannel := b.pubChannel
	confirms := b.confirms
	b.mu.RUnlock()

	options := messaging.ApplyPublishOptions(opts...)
	start := time.Now()

	// Build AMQP publishing
	pub := amqp.Publishing{
		ContentType:     options.ContentType,
		ContentEncoding: options.ContentEncoding,
		DeliveryMode:    uint8(message.DeliveryMode),
		Priority:        uint8(message.Priority),
		MessageId:       message.ID,
		Timestamp:       message.Timestamp,
		Type:            message.Type,
		Body:            message.Payload,
		Headers:         make(amqp.Table),
	}

	// Set headers
	for k, v := range message.Headers {
		pub.Headers[k] = v
	}
	if message.TraceID != "" {
		pub.Headers["trace_id"] = message.TraceID
	}
	if message.CorrelationID != "" {
		pub.CorrelationId = message.CorrelationID
	}
	if message.ReplyTo != "" {
		pub.ReplyTo = message.ReplyTo
	}
	if !message.Expiration.IsZero() {
		ttl := time.Until(message.Expiration)
		if ttl > 0 {
			pub.Expiration = string(rune(ttl.Milliseconds()))
		}
	}

	// Determine exchange and routing key
	exchange := options.Exchange
	routingKey := options.RoutingKey
	if routingKey == "" {
		routingKey = topic
	}

	// Publish with context
	pubCtx := ctx
	if options.Timeout > 0 {
		var cancel context.CancelFunc
		pubCtx, cancel = context.WithTimeout(ctx, options.Timeout)
		defer cancel()
	}

	err := pubChannel.PublishWithContext(
		pubCtx,
		exchange,
		routingKey,
		options.Mandatory,
		options.Immediate,
		pub,
	)
	if err != nil {
		b.metrics.RecordPublish(int64(len(message.Payload)), time.Since(start), false)
		return messaging.PublishError(topic, err)
	}

	// Wait for confirmation if enabled
	if b.config.PublisherConfirm && options.Confirm && confirms != nil {
		select {
		case confirm, ok := <-confirms:
			if !ok {
				b.metrics.RecordPublish(int64(len(message.Payload)), time.Since(start), false)
				return messaging.NewBrokerError(messaging.ErrCodePublishFailed, "confirm channel closed", nil)
			}
			if !confirm.Ack {
				b.metrics.RecordPublish(int64(len(message.Payload)), time.Since(start), false)
				return messaging.NewBrokerError(messaging.ErrCodePublishRejected, "message nacked by broker", nil)
			}
			b.metrics.RecordPublishConfirmation()
		case <-time.After(b.config.PublisherConfirmTimeout):
			b.metrics.RecordPublishTimeout()
			return messaging.PublishTimeoutError(topic)
		case <-pubCtx.Done():
			b.metrics.RecordPublish(int64(len(message.Payload)), time.Since(start), false)
			return pubCtx.Err()
		}
	}

	b.metrics.RecordPublish(int64(len(message.Payload)), time.Since(start), true)
	return nil
}

// PublishBatch publishes multiple messages.
func (b *Broker) PublishBatch(ctx context.Context, topic string, messages []*messaging.Message, opts ...messaging.PublishOption) error {
	for _, msg := range messages {
		if err := b.Publish(ctx, topic, msg, opts...); err != nil {
			return err
		}
	}
	return nil
}

// Subscribe creates a subscription to a queue.
func (b *Broker) Subscribe(ctx context.Context, topic string, handler messaging.MessageHandler, opts ...messaging.SubscribeOption) (messaging.Subscription, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.connected {
		return nil, messaging.ErrNotConnected
	}

	options := messaging.ApplySubscribeOptions(opts...)

	// Create a new channel for this subscription
	ch, err := b.conn.Channel()
	if err != nil {
		return nil, messaging.SubscribeError(topic, err)
	}

	// Set QoS
	if err := ch.Qos(options.Prefetch, options.PrefetchSize, false); err != nil {
		ch.Close()
		return nil, messaging.SubscribeError(topic, err)
	}

	// Declare queue if it doesn't exist
	if _, err := ch.QueueDeclare(
		topic,
		b.config.DefaultQueueDurable,
		b.config.DefaultQueueAutoDelete,
		b.config.DefaultQueueExclusive,
		false,
		nil,
	); err != nil {
		ch.Close()
		return nil, messaging.SubscribeError(topic, err)
	}

	// Start consuming
	deliveries, err := ch.Consume(
		topic,
		options.ConsumerTag,
		options.AutoAck,
		options.Exclusive,
		options.NoLocal,
		options.NoWait,
		nil,
	)
	if err != nil {
		ch.Close()
		return nil, messaging.SubscribeError(topic, err)
	}

	sub := &Subscription{
		id:         generateSubscriptionID(),
		topic:      topic,
		broker:     b,
		channel:    ch,
		deliveries: deliveries,
		handler:    handler,
		options:    options,
		active:     true,
		stopCh:     make(chan struct{}),
	}

	b.subscribers[sub.id] = sub
	b.metrics.RecordSubscription()

	// Start consuming in background
	go sub.consume(ctx)

	return sub, nil
}

// BrokerType returns the broker type.
func (b *Broker) BrokerType() messaging.BrokerType {
	return messaging.BrokerTypeRabbitMQ
}

// GetMetrics returns broker metrics.
func (b *Broker) GetMetrics() *messaging.BrokerMetrics {
	return b.metrics
}

// SetTopology sets a custom topology configuration.
func (b *Broker) SetTopology(topology *TopologyConfig) {
	b.topology = topology
}

// generateSubscriptionID generates a unique subscription ID.
func generateSubscriptionID() string {
	return "sub-" + time.Now().UTC().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random alphanumeric string.
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		time.Sleep(time.Nanosecond)
	}
	return string(b)
}
