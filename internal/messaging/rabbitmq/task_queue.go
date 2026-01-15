// Package rabbitmq provides a RabbitMQ message broker implementation.
package rabbitmq

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"dev.helix.agent/internal/messaging"
)

// TaskQueueBroker extends Broker with task queue functionality.
type TaskQueueBroker struct {
	*Broker
	taskMetrics *TaskQueueMetrics
	consumers   map[string]*TaskConsumer
	mu          sync.RWMutex
}

// TaskQueueMetrics holds task queue specific metrics.
type TaskQueueMetrics struct {
	TasksEnqueued    int64
	TasksDequeued    int64
	TasksCompleted   int64
	TasksFailed      int64
	TasksDeadLetter  int64
	TasksRetried     int64
	mu               sync.RWMutex
}

// TaskConsumer represents a task consumer.
type TaskConsumer struct {
	id         string
	queue      string
	workerID   string
	broker     *TaskQueueBroker
	channel    *amqp.Channel
	deliveries <-chan amqp.Delivery
	handler    messaging.TaskHandler
	active     bool
	stopCh     chan struct{}
	mu         sync.RWMutex
}

// NewTaskQueueBroker creates a new task queue broker.
func NewTaskQueueBroker(config *Config) *TaskQueueBroker {
	return &TaskQueueBroker{
		Broker:      NewBroker(config),
		taskMetrics: &TaskQueueMetrics{},
		consumers:   make(map[string]*TaskConsumer),
	}
}

// DeclareQueue declares a queue for tasks.
func (b *TaskQueueBroker) DeclareQueue(ctx context.Context, name string, opts ...messaging.QueueOption) error {
	b.Broker.mu.RLock()
	if !b.Broker.connected {
		b.Broker.mu.RUnlock()
		return messaging.ErrNotConnected
	}
	pubChannel := b.Broker.pubChannel
	b.Broker.mu.RUnlock()

	options := messaging.ApplyQueueOptions(opts...)

	args := make(amqp.Table)
	if options.DeadLetterExchange != "" {
		args["x-dead-letter-exchange"] = options.DeadLetterExchange
	}
	if options.DeadLetterRoutingKey != "" {
		args["x-dead-letter-routing-key"] = options.DeadLetterRoutingKey
	}
	if options.MessageTTL > 0 {
		args["x-message-ttl"] = options.MessageTTL.Milliseconds()
	}
	if options.MaxLength > 0 {
		args["x-max-length"] = options.MaxLength
	}
	if options.MaxPriority > 0 {
		args["x-max-priority"] = options.MaxPriority
	}

	_, err := pubChannel.QueueDeclare(
		name,
		options.Durable,
		options.AutoDelete,
		options.Exclusive,
		false,
		args,
	)
	if err != nil {
		return messaging.QueueError(name, err)
	}

	b.Broker.metrics.RecordQueueDeclared()
	return nil
}

// EnqueueTask enqueues a task.
func (b *TaskQueueBroker) EnqueueTask(ctx context.Context, queue string, task *messaging.Task) error {
	if task == nil {
		return messaging.ValidationError("task is nil")
	}

	b.Broker.mu.RLock()
	if !b.Broker.connected {
		b.Broker.mu.RUnlock()
		return messaging.ErrNotConnected
	}
	pubChannel := b.Broker.pubChannel
	confirms := b.Broker.confirms
	b.Broker.mu.RUnlock()

	// Serialize task
	taskData, err := json.Marshal(task)
	if err != nil {
		return messaging.SerializationError(err)
	}

	// Build AMQP publishing
	pub := amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Priority:     uint8(task.Priority),
		MessageId:    task.ID,
		Timestamp:    task.CreatedAt,
		Type:         task.Type,
		Body:         taskData,
		Headers:      make(amqp.Table),
	}

	// Set headers
	for k, v := range task.Metadata {
		pub.Headers[k] = v
	}
	if task.TraceID != "" {
		pub.Headers["trace_id"] = task.TraceID
	}
	if task.CorrelationID != "" {
		pub.CorrelationId = task.CorrelationID
	}
	if task.ParentTaskID != "" {
		pub.Headers["parent_task_id"] = task.ParentTaskID
	}
	if !task.Deadline.IsZero() {
		ttl := time.Until(task.Deadline)
		if ttl > 0 {
			pub.Expiration = string(rune(ttl.Milliseconds()))
		}
	}

	// Update task state
	task.SetState(messaging.TaskStateQueued)

	// Publish with timeout
	pubCtx := ctx
	if b.Broker.config.PublishTimeout > 0 {
		var cancel context.CancelFunc
		pubCtx, cancel = context.WithTimeout(ctx, b.Broker.config.PublishTimeout)
		defer cancel()
	}

	err = pubChannel.PublishWithContext(
		pubCtx,
		"",    // Default exchange for direct queue publish
		queue, // Routing key = queue name
		false, // Mandatory
		false, // Immediate
		pub,
	)
	if err != nil {
		return messaging.PublishError(queue, err)
	}

	// Wait for confirmation if enabled
	if b.Broker.config.PublisherConfirm && confirms != nil {
		select {
		case confirm, ok := <-confirms:
			if !ok {
				return messaging.NewBrokerError(messaging.ErrCodePublishFailed, "confirm channel closed", nil)
			}
			if !confirm.Ack {
				return messaging.NewBrokerError(messaging.ErrCodePublishRejected, "task nacked by broker", nil)
			}
			b.Broker.metrics.RecordPublishConfirmation()
		case <-time.After(b.Broker.config.PublisherConfirmTimeout):
			return messaging.PublishTimeoutError(queue)
		case <-pubCtx.Done():
			return pubCtx.Err()
		}
	}

	b.taskMetrics.mu.Lock()
	b.taskMetrics.TasksEnqueued++
	b.taskMetrics.mu.Unlock()

	return nil
}

// EnqueueTaskBatch enqueues multiple tasks.
func (b *TaskQueueBroker) EnqueueTaskBatch(ctx context.Context, queue string, tasks []*messaging.Task) error {
	for _, task := range tasks {
		if err := b.EnqueueTask(ctx, queue, task); err != nil {
			return err
		}
	}
	return nil
}

// DequeueTask dequeues a task (pull-based).
func (b *TaskQueueBroker) DequeueTask(ctx context.Context, queue string, workerID string) (*messaging.Task, error) {
	b.Broker.mu.RLock()
	if !b.Broker.connected {
		b.Broker.mu.RUnlock()
		return nil, messaging.ErrNotConnected
	}
	b.Broker.mu.RUnlock()

	// Create a channel for this dequeue operation
	ch, err := b.Broker.conn.Channel()
	if err != nil {
		return nil, messaging.ConnectionError("failed to create channel", err)
	}
	defer ch.Close()

	// Get a single message
	delivery, ok, err := ch.Get(queue, false)
	if err != nil {
		return nil, messaging.QueueError(queue, err)
	}
	if !ok {
		return nil, nil // No messages available
	}

	// Deserialize task
	var task messaging.Task
	if err := json.Unmarshal(delivery.Body, &task); err != nil {
		// Reject malformed message
		delivery.Reject(false)
		return nil, messaging.DeserializationError(err)
	}

	// Update task metadata
	task.DeliveryTag = delivery.DeliveryTag
	task.SetState(messaging.TaskStateRunning)
	if task.Metadata == nil {
		task.Metadata = make(map[string]string)
	}
	task.Metadata["worker_id"] = workerID

	b.taskMetrics.mu.Lock()
	b.taskMetrics.TasksDequeued++
	b.taskMetrics.mu.Unlock()

	return &task, nil
}

// SubscribeTasks subscribes to tasks (push-based).
func (b *TaskQueueBroker) SubscribeTasks(ctx context.Context, queue string, handler messaging.TaskHandler, opts ...messaging.SubscribeOption) (messaging.Subscription, error) {
	// Wrap task handler as message handler
	msgHandler := func(ctx context.Context, msg *messaging.Message) error {
		var task messaging.Task
		if err := json.Unmarshal(msg.Payload, &task); err != nil {
			return err
		}
		task.DeliveryTag = msg.DeliveryTag
		task.SetState(messaging.TaskStateRunning)
		return handler(ctx, &task)
	}

	return b.Broker.Subscribe(ctx, queue, msgHandler, opts...)
}

// AckTask acknowledges a task.
func (b *TaskQueueBroker) AckTask(ctx context.Context, deliveryTag uint64) error {
	b.Broker.mu.RLock()
	if !b.Broker.connected {
		b.Broker.mu.RUnlock()
		return messaging.ErrNotConnected
	}
	pubChannel := b.Broker.pubChannel
	b.Broker.mu.RUnlock()

	if err := pubChannel.Ack(deliveryTag, false); err != nil {
		return messaging.NewBrokerError(messaging.ErrCodeAckFailed, "failed to ack task", err)
	}

	b.Broker.metrics.RecordAck()
	b.taskMetrics.mu.Lock()
	b.taskMetrics.TasksCompleted++
	b.taskMetrics.mu.Unlock()

	return nil
}

// NackTask negative acknowledges a task.
func (b *TaskQueueBroker) NackTask(ctx context.Context, deliveryTag uint64, requeue bool) error {
	b.Broker.mu.RLock()
	if !b.Broker.connected {
		b.Broker.mu.RUnlock()
		return messaging.ErrNotConnected
	}
	pubChannel := b.Broker.pubChannel
	b.Broker.mu.RUnlock()

	if err := pubChannel.Nack(deliveryTag, false, requeue); err != nil {
		return messaging.NewBrokerError(messaging.ErrCodeNackFailed, "failed to nack task", err)
	}

	b.Broker.metrics.RecordNack()
	if requeue {
		b.taskMetrics.mu.Lock()
		b.taskMetrics.TasksRetried++
		b.taskMetrics.mu.Unlock()
	} else {
		b.taskMetrics.mu.Lock()
		b.taskMetrics.TasksFailed++
		b.taskMetrics.mu.Unlock()
	}

	return nil
}

// RejectTask rejects a task without requeue.
func (b *TaskQueueBroker) RejectTask(ctx context.Context, deliveryTag uint64) error {
	return b.NackTask(ctx, deliveryTag, false)
}

// MoveToDeadLetter moves a task to the dead letter queue.
func (b *TaskQueueBroker) MoveToDeadLetter(ctx context.Context, task *messaging.Task, reason string) error {
	if task == nil {
		return messaging.ValidationError("task is nil")
	}

	// Update task state
	task.SetState(messaging.TaskStateDeadLettered)
	task.Error = reason
	if task.Metadata == nil {
		task.Metadata = make(map[string]string)
	}
	task.Metadata["dead_letter_reason"] = reason
	task.Metadata["dead_letter_time"] = time.Now().UTC().Format(time.RFC3339)

	// Enqueue to dead letter queue
	dlqName := b.Broker.config.DeadLetterQueue
	if dlqName == "" {
		dlqName = messaging.QueueDeadLetter
	}

	if err := b.EnqueueTask(ctx, dlqName, task); err != nil {
		return err
	}

	b.taskMetrics.mu.Lock()
	b.taskMetrics.TasksDeadLetter++
	b.taskMetrics.mu.Unlock()

	return nil
}

// RetryTask retries a failed task.
func (b *TaskQueueBroker) RetryTask(ctx context.Context, queue string, task *messaging.Task) error {
	if task == nil {
		return messaging.ValidationError("task is nil")
	}

	if !task.CanRetry() {
		return b.MoveToDeadLetter(ctx, task, "max retries exceeded")
	}

	task.IncrementRetry()
	task.SetState(messaging.TaskStatePending)

	// Calculate retry delay
	retryDelay := b.Broker.config.RetryDelay
	if retryDelay > 0 {
		// Use delayed queue or message TTL
		task.WithDelay(retryDelay * time.Duration(task.RetryCount))
	}

	if err := b.EnqueueTask(ctx, queue, task); err != nil {
		return err
	}

	b.taskMetrics.mu.Lock()
	b.taskMetrics.TasksRetried++
	b.taskMetrics.mu.Unlock()

	return nil
}

// GetQueueStats returns queue statistics.
func (b *TaskQueueBroker) GetQueueStats(ctx context.Context, queue string) (*messaging.QueueStats, error) {
	b.Broker.mu.RLock()
	if !b.Broker.connected {
		b.Broker.mu.RUnlock()
		return nil, messaging.ErrNotConnected
	}
	b.Broker.mu.RUnlock()

	// Create a channel for this operation
	ch, err := b.Broker.conn.Channel()
	if err != nil {
		return nil, messaging.ConnectionError("failed to create channel", err)
	}
	defer ch.Close()

	// Inspect queue
	queueInfo, err := ch.QueueInspect(queue)
	if err != nil {
		return nil, messaging.QueueError(queue, err)
	}

	return &messaging.QueueStats{
		Name:          queue,
		Messages:      int64(queueInfo.Messages),
		Consumers:     int64(queueInfo.Consumers),
		MessagesReady: int64(queueInfo.Messages),
	}, nil
}

// GetQueueDepth returns the number of messages in a queue.
func (b *TaskQueueBroker) GetQueueDepth(ctx context.Context, queue string) (int64, error) {
	stats, err := b.GetQueueStats(ctx, queue)
	if err != nil {
		return 0, err
	}
	return stats.Messages, nil
}

// PurgeQueue removes all messages from a queue.
func (b *TaskQueueBroker) PurgeQueue(ctx context.Context, queue string) error {
	b.Broker.mu.RLock()
	if !b.Broker.connected {
		b.Broker.mu.RUnlock()
		return messaging.ErrNotConnected
	}
	b.Broker.mu.RUnlock()

	ch, err := b.Broker.conn.Channel()
	if err != nil {
		return messaging.ConnectionError("failed to create channel", err)
	}
	defer ch.Close()

	_, err = ch.QueuePurge(queue, false)
	if err != nil {
		return messaging.QueueError(queue, err)
	}

	return nil
}

// DeleteQueue deletes a queue.
func (b *TaskQueueBroker) DeleteQueue(ctx context.Context, queue string) error {
	b.Broker.mu.RLock()
	if !b.Broker.connected {
		b.Broker.mu.RUnlock()
		return messaging.ErrNotConnected
	}
	b.Broker.mu.RUnlock()

	ch, err := b.Broker.conn.Channel()
	if err != nil {
		return messaging.ConnectionError("failed to create channel", err)
	}
	defer ch.Close()

	_, err = ch.QueueDelete(queue, false, false, false)
	if err != nil {
		return messaging.QueueError(queue, err)
	}

	return nil
}

// GetTaskMetrics returns task queue specific metrics.
func (b *TaskQueueBroker) GetTaskMetrics() *TaskQueueMetrics {
	return b.taskMetrics
}
