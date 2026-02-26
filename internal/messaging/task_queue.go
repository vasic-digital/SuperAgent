package messaging

import (
	"context"
	"encoding/json"
	"time"
)

// TaskState represents the state of a task.
type TaskState string

const (
	// TaskStatePending indicates the task is waiting to be processed.
	TaskStatePending TaskState = "pending"
	// TaskStateQueued indicates the task is in the queue.
	TaskStateQueued TaskState = "queued"
	// TaskStateRunning indicates the task is being processed.
	TaskStateRunning TaskState = "running"
	// TaskStateCompleted indicates the task completed successfully.
	TaskStateCompleted TaskState = "completed"
	// TaskStateFailed indicates the task failed.
	TaskStateFailed TaskState = "failed"
	// TaskStateCanceled indicates the task was canceled.
	TaskStateCanceled TaskState = "canceled"
	// TaskStateDeadLettered indicates the task was moved to dead letter queue.
	TaskStateDeadLettered TaskState = "dead_lettered"
)

// TaskPriority represents task priority levels.
type TaskPriority int

const (
	// TaskPriorityLow is for background tasks.
	TaskPriorityLow TaskPriority = 1
	// TaskPriorityNormal is the default priority.
	TaskPriorityNormal TaskPriority = 5
	// TaskPriorityHigh is for time-sensitive tasks.
	TaskPriorityHigh TaskPriority = 8
	// TaskPriorityCritical is for urgent tasks.
	TaskPriorityCritical TaskPriority = 10
)

// Task represents a task in the task queue.
type Task struct {
	// ID is the unique task identifier.
	ID string `json:"id"`
	// Type is the task type for routing.
	Type string `json:"type"`
	// Payload is the task data.
	Payload []byte `json:"payload"`
	// Priority is the task priority (1-10).
	Priority TaskPriority `json:"priority"`
	// State is the current task state.
	State TaskState `json:"state"`
	// MaxRetries is the maximum retry attempts.
	MaxRetries int `json:"max_retries"`
	// RetryCount is the current retry count.
	RetryCount int `json:"retry_count"`
	// Deadline is when the task must be completed.
	Deadline time.Time `json:"deadline,omitempty"`
	// CreatedAt is when the task was created.
	CreatedAt time.Time `json:"created_at"`
	// ScheduledAt is when the task should be executed.
	ScheduledAt time.Time `json:"scheduled_at,omitempty"`
	// StartedAt is when the task started processing.
	StartedAt time.Time `json:"started_at,omitempty"`
	// CompletedAt is when the task completed.
	CompletedAt time.Time `json:"completed_at,omitempty"`
	// Metadata contains additional task metadata.
	Metadata map[string]string `json:"metadata,omitempty"`
	// Result contains the task result (if completed).
	Result []byte `json:"result,omitempty"`
	// Error contains the error message (if failed).
	Error string `json:"error,omitempty"`
	// WorkerID is the ID of the worker processing the task.
	WorkerID string `json:"worker_id,omitempty"`
	// TraceID is for distributed tracing.
	TraceID string `json:"trace_id,omitempty"`
	// CorrelationID links related tasks.
	CorrelationID string `json:"correlation_id,omitempty"`
	// ParentTaskID is the parent task ID (for subtasks).
	ParentTaskID string `json:"parent_task_id,omitempty"`
	// DeliveryTag is used for acknowledgment (broker-specific).
	DeliveryTag uint64 `json:"-"`
}

// NewTask creates a new task with default values.
func NewTask(taskType string, payload []byte) *Task {
	return &Task{
		ID:         generateTaskID(),
		Type:       taskType,
		Payload:    payload,
		Priority:   TaskPriorityNormal,
		State:      TaskStatePending,
		MaxRetries: 3,
		RetryCount: 0,
		CreatedAt:  time.Now().UTC(),
		Metadata:   make(map[string]string),
	}
}

// NewTaskWithID creates a new task with a specific ID.
func NewTaskWithID(id, taskType string, payload []byte) *Task {
	task := NewTask(taskType, payload)
	task.ID = id
	return task
}

// WithPriority sets the task priority.
func (t *Task) WithPriority(priority TaskPriority) *Task {
	t.Priority = priority
	return t
}

// WithMaxRetries sets the maximum retries.
func (t *Task) WithMaxRetries(maxRetries int) *Task {
	t.MaxRetries = maxRetries
	return t
}

// WithDeadline sets the task deadline.
func (t *Task) WithDeadline(deadline time.Time) *Task {
	t.Deadline = deadline
	return t
}

// WithTTL sets the task deadline based on TTL from now.
func (t *Task) WithTTL(ttl time.Duration) *Task {
	t.Deadline = time.Now().UTC().Add(ttl)
	return t
}

// WithScheduledAt sets when the task should be executed.
func (t *Task) WithScheduledAt(scheduledAt time.Time) *Task {
	t.ScheduledAt = scheduledAt
	return t
}

// WithDelay sets a delay before the task is executed.
func (t *Task) WithDelay(delay time.Duration) *Task {
	t.ScheduledAt = time.Now().UTC().Add(delay)
	return t
}

// WithMetadata adds metadata to the task.
func (t *Task) WithMetadata(key, value string) *Task {
	if t.Metadata == nil {
		t.Metadata = make(map[string]string)
	}
	t.Metadata[key] = value
	return t
}

// WithTraceID sets the trace ID.
func (t *Task) WithTraceID(traceID string) *Task {
	t.TraceID = traceID
	return t
}

// WithCorrelationID sets the correlation ID.
func (t *Task) WithCorrelationID(correlationID string) *Task {
	t.CorrelationID = correlationID
	return t
}

// WithParentTaskID sets the parent task ID.
func (t *Task) WithParentTaskID(parentID string) *Task {
	t.ParentTaskID = parentID
	return t
}

// GetMetadata gets a metadata value.
func (t *Task) GetMetadata(key string) string {
	if t.Metadata == nil {
		return ""
	}
	return t.Metadata[key]
}

// IsExpired checks if the task has exceeded its deadline.
func (t *Task) IsExpired() bool {
	if t.Deadline.IsZero() {
		return false
	}
	return time.Now().UTC().After(t.Deadline)
}

// IsScheduled checks if the task is scheduled for later.
func (t *Task) IsScheduled() bool {
	if t.ScheduledAt.IsZero() {
		return false
	}
	return time.Now().UTC().Before(t.ScheduledAt)
}

// CanRetry checks if the task can be retried.
func (t *Task) CanRetry() bool {
	return t.RetryCount < t.MaxRetries
}

// IncrementRetry increments the retry count.
func (t *Task) IncrementRetry() {
	t.RetryCount++
}

// SetState sets the task state with timestamp.
func (t *Task) SetState(state TaskState) {
	t.State = state
	now := time.Now().UTC()
	switch state {
	case TaskStateRunning:
		t.StartedAt = now
	case TaskStateCompleted, TaskStateFailed, TaskStateCanceled:
		t.CompletedAt = now
	}
}

// SetResult sets the task result.
func (t *Task) SetResult(result []byte) {
	t.Result = result
	t.SetState(TaskStateCompleted)
}

// SetError sets the task error.
func (t *Task) SetError(err string) {
	t.Error = err
	t.SetState(TaskStateFailed)
}

// Duration returns the task processing duration.
func (t *Task) Duration() time.Duration {
	if t.StartedAt.IsZero() {
		return 0
	}
	if t.CompletedAt.IsZero() {
		return time.Since(t.StartedAt)
	}
	return t.CompletedAt.Sub(t.StartedAt)
}

// Clone creates a deep copy of the task.
func (t *Task) Clone() *Task {
	clone := &Task{
		ID:            t.ID,
		Type:          t.Type,
		Payload:       make([]byte, len(t.Payload)),
		Priority:      t.Priority,
		State:         t.State,
		MaxRetries:    t.MaxRetries,
		RetryCount:    t.RetryCount,
		Deadline:      t.Deadline,
		CreatedAt:     t.CreatedAt,
		ScheduledAt:   t.ScheduledAt,
		StartedAt:     t.StartedAt,
		CompletedAt:   t.CompletedAt,
		Result:        make([]byte, len(t.Result)),
		Error:         t.Error,
		WorkerID:      t.WorkerID,
		TraceID:       t.TraceID,
		CorrelationID: t.CorrelationID,
		ParentTaskID:  t.ParentTaskID,
		DeliveryTag:   t.DeliveryTag,
	}
	copy(clone.Payload, t.Payload)
	copy(clone.Result, t.Result)
	if t.Metadata != nil {
		clone.Metadata = make(map[string]string)
		for k, v := range t.Metadata {
			clone.Metadata[k] = v
		}
	}
	return clone
}

// ToMessage converts the task to a Message.
func (t *Task) ToMessage() *Message {
	payload, _ := json.Marshal(t) //nolint:errcheck
	msg := NewMessage(t.Type, payload)
	msg.ID = t.ID
	msg.Priority = MessagePriority(t.Priority)
	msg.TraceID = t.TraceID
	msg.CorrelationID = t.CorrelationID
	msg.RetryCount = t.RetryCount
	msg.MaxRetries = t.MaxRetries
	if !t.Deadline.IsZero() {
		msg.Expiration = t.Deadline
	}
	return msg
}

// TaskFromMessage creates a Task from a Message.
func TaskFromMessage(msg *Message) (*Task, error) {
	var task Task
	if err := json.Unmarshal(msg.Payload, &task); err != nil {
		return nil, err
	}
	task.DeliveryTag = msg.DeliveryTag
	return &task, nil
}

// generateTaskID generates a unique task ID.
func generateTaskID() string {
	return "task-" + time.Now().UTC().Format("20060102150405") + "-" + randomString(8)
}

// TaskHandler is a function that processes tasks.
type TaskHandler func(ctx context.Context, task *Task) error

// TaskFilter is a function that filters tasks.
type TaskFilter func(task *Task) bool

// TaskQueueBroker defines the interface for task queue brokers (e.g., RabbitMQ).
type TaskQueueBroker interface {
	MessageBroker

	// DeclareQueue declares a task queue.
	DeclareQueue(ctx context.Context, name string, opts ...QueueOption) error

	// EnqueueTask adds a task to the queue.
	EnqueueTask(ctx context.Context, queue string, task *Task) error

	// EnqueueTaskBatch adds multiple tasks to the queue.
	EnqueueTaskBatch(ctx context.Context, queue string, tasks []*Task) error

	// DequeueTask retrieves a task from the queue.
	DequeueTask(ctx context.Context, queue string, workerID string) (*Task, error)

	// AckTask acknowledges successful task processing.
	AckTask(ctx context.Context, deliveryTag uint64) error

	// NackTask negatively acknowledges a task (can requeue).
	NackTask(ctx context.Context, deliveryTag uint64, requeue bool) error

	// RejectTask rejects a task without requeuing.
	RejectTask(ctx context.Context, deliveryTag uint64) error

	// MoveToDeadLetter moves a failed task to dead letter queue.
	MoveToDeadLetter(ctx context.Context, task *Task, reason string) error

	// GetQueueStats returns queue statistics.
	GetQueueStats(ctx context.Context, queue string) (*QueueStats, error)

	// GetQueueDepth returns the number of messages in the queue.
	GetQueueDepth(ctx context.Context, queue string) (int64, error)

	// PurgeQueue removes all messages from a queue.
	PurgeQueue(ctx context.Context, queue string) error

	// DeleteQueue deletes a queue.
	DeleteQueue(ctx context.Context, queue string) error

	// SubscribeTasks subscribes to tasks from a queue.
	SubscribeTasks(ctx context.Context, queue string, handler TaskHandler, opts ...SubscribeOption) (Subscription, error)
}

// TaskQueueConfig holds configuration for task queue broker.
type TaskQueueConfig struct {
	BrokerConfig

	// DefaultQueue is the default queue name.
	DefaultQueue string `json:"default_queue" yaml:"default_queue"`

	// DeadLetterQueue is the dead letter queue name.
	DeadLetterQueue string `json:"dead_letter_queue" yaml:"dead_letter_queue"`

	// RetryQueue is the retry queue name.
	RetryQueue string `json:"retry_queue" yaml:"retry_queue"`

	// PrefetchCount is the prefetch count for consumers.
	PrefetchCount int `json:"prefetch_count" yaml:"prefetch_count"`

	// PublisherConfirm enables publisher confirms.
	PublisherConfirm bool `json:"publisher_confirm" yaml:"publisher_confirm"`

	// MaxPriority is the maximum priority level (0-255).
	MaxPriority int `json:"max_priority" yaml:"max_priority"`

	// TaskTTL is the default task time-to-live.
	TaskTTL time.Duration `json:"task_ttl" yaml:"task_ttl"`

	// RetryDelay is the delay before retrying failed tasks.
	RetryDelay time.Duration `json:"retry_delay" yaml:"retry_delay"`

	// MaxRetries is the default maximum retry attempts.
	MaxRetries int `json:"max_retries" yaml:"max_retries"`
}

// DefaultTaskQueueConfig returns the default task queue configuration.
func DefaultTaskQueueConfig() *TaskQueueConfig {
	return &TaskQueueConfig{
		BrokerConfig:     *DefaultBrokerConfig(),
		DefaultQueue:     "helixagent.tasks.default",
		DeadLetterQueue:  "helixagent.tasks.dlq",
		RetryQueue:       "helixagent.tasks.retry",
		PrefetchCount:    10,
		PublisherConfirm: true,
		MaxPriority:      10,
		TaskTTL:          24 * time.Hour,
		RetryDelay:       5 * time.Second,
		MaxRetries:       3,
	}
}

// Queue names for HelixAgent task queues.
const (
	// QueueBackgroundTasks is for background command execution.
	QueueBackgroundTasks = "helixagent.tasks.background"
	// QueueLLMRequests is for LLM provider requests.
	QueueLLMRequests = "helixagent.tasks.llm"
	// QueueDebateRounds is for AI debate rounds.
	QueueDebateRounds = "helixagent.tasks.debate"
	// QueueVerification is for provider verification tasks.
	QueueVerification = "helixagent.tasks.verification"
	// QueueNotifications is for notification delivery tasks.
	QueueNotifications = "helixagent.tasks.notifications"
	// QueueDeadLetter is the dead letter queue.
	QueueDeadLetter = "helixagent.tasks.dlq"
	// QueueRetry is the retry queue.
	QueueRetry = "helixagent.tasks.retry"
)

// Exchange names for HelixAgent.
const (
	// ExchangeTasks is the main task exchange.
	ExchangeTasks = "helixagent.tasks"
	// ExchangeEvents is the event exchange.
	ExchangeEvents = "helixagent.events"
	// ExchangeNotifications is the notifications exchange.
	ExchangeNotifications = "helixagent.notifications"
	// ExchangeDeadLetter is the dead letter exchange.
	ExchangeDeadLetter = "helixagent.dlx"
)

// TaskWorkerConfig holds configuration for task workers.
type TaskWorkerConfig struct {
	// WorkerID is the unique worker identifier.
	WorkerID string `json:"worker_id" yaml:"worker_id"`
	// Queue is the queue to consume from.
	Queue string `json:"queue" yaml:"queue"`
	// Concurrency is the number of concurrent task handlers.
	Concurrency int `json:"concurrency" yaml:"concurrency"`
	// MaxRetries is the maximum retry attempts per task.
	MaxRetries int `json:"max_retries" yaml:"max_retries"`
	// RetryDelay is the delay between retries.
	RetryDelay time.Duration `json:"retry_delay" yaml:"retry_delay"`
	// Timeout is the task processing timeout.
	Timeout time.Duration `json:"timeout" yaml:"timeout"`
	// GracefulShutdown is the shutdown timeout.
	GracefulShutdown time.Duration `json:"graceful_shutdown" yaml:"graceful_shutdown"`
}

// DefaultTaskWorkerConfig returns the default task worker configuration.
func DefaultTaskWorkerConfig() *TaskWorkerConfig {
	return &TaskWorkerConfig{
		WorkerID:         "worker-" + randomString(8),
		Queue:            QueueBackgroundTasks,
		Concurrency:      10,
		MaxRetries:       3,
		RetryDelay:       5 * time.Second,
		Timeout:          5 * time.Minute,
		GracefulShutdown: 30 * time.Second,
	}
}

// TaskResult represents the result of a task execution.
type TaskResult struct {
	TaskID      string        `json:"task_id"`
	Success     bool          `json:"success"`
	Result      interface{}   `json:"result,omitempty"`
	Error       string        `json:"error,omitempty"`
	Duration    time.Duration `json:"duration"`
	CompletedAt time.Time     `json:"completed_at"`
}

// NewTaskResult creates a new task result.
func NewTaskResult(taskID string, result interface{}, err error, duration time.Duration) *TaskResult {
	r := &TaskResult{
		TaskID:      taskID,
		Success:     err == nil,
		Duration:    duration,
		CompletedAt: time.Now().UTC(),
	}
	if err != nil {
		r.Error = err.Error()
	} else {
		r.Result = result
	}
	return r
}

// TaskScheduler schedules delayed tasks.
type TaskScheduler interface {
	// Schedule schedules a task for future execution.
	Schedule(ctx context.Context, task *Task) error
	// Cancel cancels a scheduled task.
	Cancel(ctx context.Context, taskID string) error
	// GetScheduled returns scheduled tasks.
	GetScheduled(ctx context.Context) ([]*Task, error)
}

// TaskRegistry holds task type to handler mappings.
type TaskRegistry struct {
	handlers map[string]TaskHandler
}

// NewTaskRegistry creates a new task registry.
func NewTaskRegistry() *TaskRegistry {
	return &TaskRegistry{
		handlers: make(map[string]TaskHandler),
	}
}

// Register registers a handler for a task type.
func (r *TaskRegistry) Register(taskType string, handler TaskHandler) {
	r.handlers[taskType] = handler
}

// Get returns the handler for a task type.
func (r *TaskRegistry) Get(taskType string) (TaskHandler, bool) {
	handler, ok := r.handlers[taskType]
	return handler, ok
}

// Unregister removes a handler for a task type.
func (r *TaskRegistry) Unregister(taskType string) {
	delete(r.handlers, taskType)
}

// Types returns all registered task types.
func (r *TaskRegistry) Types() []string {
	types := make([]string, 0, len(r.handlers))
	for t := range r.handlers {
		types = append(types, t)
	}
	return types
}
