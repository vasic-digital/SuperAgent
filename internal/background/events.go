// Package background provides background task event types and publishing
// for integration with the messaging system.
package background

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"dev.helix.agent/internal/messaging"
	"dev.helix.agent/internal/models"
	extractedbackground "digital.vasic.background"
)

// Task event topics for the messaging system.
const (
	// TopicTaskEvents is the main topic for all task lifecycle events.
	TopicTaskEvents = "helixagent.events.tasks"
	// TopicTaskCreated is for task creation events.
	TopicTaskCreated = "helixagent.events.tasks.created"
	// TopicTaskStarted is for task start events.
	TopicTaskStarted = "helixagent.events.tasks.started"
	// TopicTaskProgress is for task progress events.
	TopicTaskProgress = "helixagent.events.tasks.progress"
	// TopicTaskCompleted is for task completion events.
	TopicTaskCompleted = "helixagent.events.tasks.completed"
	// TopicTaskFailed is for task failure events.
	TopicTaskFailed = "helixagent.events.tasks.failed"
	// TopicTaskStuck is for stuck task events.
	TopicTaskStuck = "helixagent.events.tasks.stuck"
	// TopicTaskCancelled is for task cancellation events.
	TopicTaskCancelled = "helixagent.events.tasks.cancelled"
	// TopicTaskRetrying is for task retry events.
	TopicTaskRetrying = "helixagent.events.tasks.retrying"
	// TopicTaskDeadLetter is for dead letter events.
	TopicTaskDeadLetter = "helixagent.events.tasks.deadletter"
)

// TaskEventType represents the type of task event.
type TaskEventType string

const (
	TaskEventTypeCreated    TaskEventType = "task.created"
	TaskEventTypeStarted    TaskEventType = "task.started"
	TaskEventTypeProgress   TaskEventType = "task.progress"
	TaskEventTypeHeartbeat  TaskEventType = "task.heartbeat"
	TaskEventTypePaused     TaskEventType = "task.paused"
	TaskEventTypeResumed    TaskEventType = "task.resumed"
	TaskEventTypeCompleted  TaskEventType = "task.completed"
	TaskEventTypeFailed     TaskEventType = "task.failed"
	TaskEventTypeStuck      TaskEventType = "task.stuck"
	TaskEventTypeCancelled  TaskEventType = "task.cancelled"
	TaskEventTypeRetrying   TaskEventType = "task.retrying"
	TaskEventTypeDeadLetter TaskEventType = "task.deadletter"
	TaskEventTypeLog        TaskEventType = "task.log"
	TaskEventTypeResource   TaskEventType = "task.resource"
)

// String returns the string representation of TaskEventType.
func (t TaskEventType) String() string {
	return string(t)
}

// Topic returns the appropriate topic for this event type.
func (t TaskEventType) Topic() string {
	switch t {
	case TaskEventTypeCreated:
		return TopicTaskCreated
	case TaskEventTypeStarted:
		return TopicTaskStarted
	case TaskEventTypeProgress, TaskEventTypeHeartbeat:
		return TopicTaskProgress
	case TaskEventTypeCompleted:
		return TopicTaskCompleted
	case TaskEventTypeFailed:
		return TopicTaskFailed
	case TaskEventTypeStuck:
		return TopicTaskStuck
	case TaskEventTypeCancelled:
		return TopicTaskCancelled
	case TaskEventTypeRetrying:
		return TopicTaskRetrying
	case TaskEventTypeDeadLetter:
		return TopicTaskDeadLetter
	default:
		return TopicTaskEvents
	}
}

// BackgroundTaskEvent represents a task lifecycle event.
type BackgroundTaskEvent struct {
	// EventID is the unique identifier for this event.
	EventID string `json:"event_id"`
	// EventType is the type of event.
	EventType TaskEventType `json:"event_type"`
	// TaskID is the ID of the task.
	TaskID string `json:"task_id"`
	// TaskType is the type of task.
	TaskType string `json:"task_type"`
	// TaskName is the name of the task.
	TaskName string `json:"task_name"`
	// Status is the current status of the task.
	Status models.TaskStatus `json:"status"`
	// WorkerID is the ID of the worker processing the task.
	WorkerID string `json:"worker_id,omitempty"`
	// Progress is the current progress (0-100).
	Progress float64 `json:"progress,omitempty"`
	// ProgressMessage is an optional progress message.
	ProgressMessage string `json:"progress_message,omitempty"`
	// Error contains error information if applicable.
	Error string `json:"error,omitempty"`
	// Duration is the elapsed duration.
	Duration time.Duration `json:"duration,omitempty"`
	// RetryCount is the current retry count.
	RetryCount int `json:"retry_count,omitempty"`
	// Metadata contains additional event data.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	// Timestamp is when the event occurred.
	Timestamp time.Time `json:"timestamp"`
	// CorrelationID links related events.
	CorrelationID string `json:"correlation_id,omitempty"`
	// TraceID for distributed tracing.
	TraceID string `json:"trace_id,omitempty"`
}

// logrusLoggerAdapter adapts logrus.Logger to extractedbackground.Logger interface
type logrusLoggerAdapter struct {
	logger *logrus.Logger
}

func (l *logrusLoggerAdapter) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}

func (l *logrusLoggerAdapter) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

func (l *logrusLoggerAdapter) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}

func (l *logrusLoggerAdapter) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}

// messagingHubEventPublisher adapts messaging.MessagingHub to extractedbackground.EventPublisher
type messagingHubEventPublisher struct {
	hub    *messaging.MessagingHub
	logger *logrus.Logger
}

func (m *messagingHubEventPublisher) Publish(ctx context.Context, event *extractedbackground.BackgroundTaskEvent) error {
	if event == nil || m.hub == nil {
		return nil
	}

	// Convert extracted BackgroundTaskEvent to messaging.Event
	msgEvent := m.convertToMessagingEvent(event)
	if msgEvent == nil {
		m.logger.WithField("task_id", event.TaskID).Warn("Failed to convert background task event to messaging event")
		return nil // Silently ignore conversion errors
	}

	topic := event.EventType.Topic()
	if err := m.hub.PublishEvent(ctx, topic, msgEvent); err != nil {
		m.logger.WithError(err).WithFields(logrus.Fields{
			"event_type": event.EventType,
			"task_id":    event.TaskID,
			"topic":      topic,
		}).Error("Failed to publish task event to messaging hub")
		return err
	}

	m.logger.WithFields(logrus.Fields{
		"event_type": event.EventType,
		"task_id":    event.TaskID,
		"topic":      topic,
	}).Debug("Published task event to messaging hub")

	return nil
}

func (m *messagingHubEventPublisher) convertToMessagingEvent(extracted *extractedbackground.BackgroundTaskEvent) *messaging.Event {
	if extracted == nil {
		return nil
	}

	// Marshal the event to JSON for the data field
	data, err := json.Marshal(extracted)
	if err != nil {
		return nil
	}

	// Create messaging.Event with appropriate fields
	return &messaging.Event{
		ID:            extracted.EventID,
		Type:          messaging.EventType(extracted.EventType),
		Source:        "helixagent.background",
		Subject:       extracted.TaskID,
		Data:          data,
		DataSchema:    "application/json",
		Timestamp:     extracted.Timestamp,
		CorrelationID: extracted.CorrelationID,
		TraceID:       extracted.TraceID,
	}
}

// toExtractedBackgroundTaskEvent converts internal BackgroundTaskEvent to extracted BackgroundTaskEvent
func toExtractedBackgroundTaskEvent(internal *BackgroundTaskEvent) *extractedbackground.BackgroundTaskEvent {
	if internal == nil {
		return nil
	}
	// Use JSON marshaling/unmarshaling for conversion since structs are identical
	data, err := json.Marshal(internal)
	if err != nil {
		return nil
	}
	var extracted extractedbackground.BackgroundTaskEvent
	if err := json.Unmarshal(data, &extracted); err != nil {
		return nil
	}
	return &extracted
}

// NewBackgroundTaskEvent creates a new background task event.
func NewBackgroundTaskEvent(eventType TaskEventType, task *models.BackgroundTask) *BackgroundTaskEvent {
	event := &BackgroundTaskEvent{
		EventID:   generateEventID(),
		EventType: eventType,
		TaskID:    task.ID,
		TaskType:  task.TaskType,
		TaskName:  task.TaskName,
		Status:    task.Status,
		Progress:  task.Progress,
		Timestamp: time.Now().UTC(),
		Metadata:  make(map[string]interface{}),
	}

	if task.WorkerID != nil {
		event.WorkerID = *task.WorkerID
	}
	if task.ProgressMessage != nil {
		event.ProgressMessage = *task.ProgressMessage
	}
	if task.LastError != nil {
		event.Error = *task.LastError
	}
	if task.CorrelationID != nil {
		event.CorrelationID = *task.CorrelationID
	}
	if task.StartedAt != nil {
		event.Duration = time.Since(*task.StartedAt)
	}
	event.RetryCount = task.RetryCount

	return event
}

// ToMessagingEvent converts BackgroundTaskEvent to messaging.Event.
func (e *BackgroundTaskEvent) ToMessagingEvent() *messaging.Event {
	data, _ := json.Marshal(e) //nolint:errcheck // marshaling of simple struct should never fail
	return &messaging.Event{
		ID:            e.EventID,
		Type:          messaging.EventType(e.EventType),
		Source:        "helixagent.background",
		Subject:       e.TaskID,
		Data:          data,
		DataSchema:    "application/json",
		Timestamp:     e.Timestamp,
		CorrelationID: e.CorrelationID,
		TraceID:       e.TraceID,
	}
}

// TaskEventPublisher publishes task lifecycle events to the messaging system.
type TaskEventPublisher struct {
	hub          *messaging.MessagingHub
	logger       *logrus.Logger
	enabled      bool
	asyncPublish bool
	publishCh    chan *BackgroundTaskEvent
	stopCh       chan struct{}
	wg           sync.WaitGroup //nolint:unused
	mu           sync.RWMutex
	extracted    *extractedbackground.TaskEventPublisher
}

// TaskEventPublisherConfig holds configuration for the event publisher.
type TaskEventPublisherConfig struct {
	// Enabled enables event publishing.
	Enabled bool `json:"enabled" yaml:"enabled"`
	// AsyncPublish enables async event publishing.
	AsyncPublish bool `json:"async_publish" yaml:"async_publish"`
	// BufferSize is the async publish buffer size.
	BufferSize int `json:"buffer_size" yaml:"buffer_size"`
}

// DefaultTaskEventPublisherConfig returns default configuration.
func DefaultTaskEventPublisherConfig() *TaskEventPublisherConfig {
	return &TaskEventPublisherConfig{
		Enabled:      true,
		AsyncPublish: true,
		BufferSize:   1000,
	}
}

// NewTaskEventPublisher creates a new task event publisher.
func NewTaskEventPublisher(hub *messaging.MessagingHub, logger *logrus.Logger, config *TaskEventPublisherConfig) *TaskEventPublisher {
	if config == nil {
		config = DefaultTaskEventPublisherConfig()
	}
	if logger == nil {
		logger = logrus.New()
	}

	// Create adapters for messaging hub and logger
	eventPublisher := &messagingHubEventPublisher{
		hub:    hub,
		logger: logger,
	}
	loggerAdapter := &logrusLoggerAdapter{logger: logger}

	// Create extracted task event publisher with same configuration
	extractedConfig := &extractedbackground.TaskEventPublisherConfig{
		Enabled:      config.Enabled,
		AsyncPublish: config.AsyncPublish,
		BufferSize:   config.BufferSize,
	}
	extractedPublisher := extractedbackground.NewTaskEventPublisher(eventPublisher, loggerAdapter, extractedConfig)

	p := &TaskEventPublisher{
		hub:          hub,
		logger:       logger,
		enabled:      config.Enabled,
		asyncPublish: config.AsyncPublish,
		stopCh:       make(chan struct{}),
		extracted:    extractedPublisher,
	}

	if config.AsyncPublish && config.BufferSize > 0 {
		p.publishCh = make(chan *BackgroundTaskEvent, config.BufferSize)
	}

	// Start the extracted publisher if async publishing is enabled
	extractedPublisher.Start()

	return p
}

// Start starts the async publish goroutine if enabled.
func (p *TaskEventPublisher) Start() {
	// Delegate to extracted publisher
	if p.extracted != nil {
		p.extracted.Start()
	}
}

// Stop stops the event publisher.
func (p *TaskEventPublisher) Stop() {
	// Delegate to extracted publisher
	if p.extracted != nil {
		p.extracted.Stop()
	}
}

// Publish publishes a task event.
func (p *TaskEventPublisher) Publish(ctx context.Context, event *BackgroundTaskEvent) error {
	if p.extracted == nil {
		return nil
	}
	// Convert internal event to extracted event
	extractedEvent := toExtractedBackgroundTaskEvent(event)
	if extractedEvent == nil {
		return nil
	}
	return p.extracted.Publish(ctx, extractedEvent)
}

// PublishTaskCreated publishes a task created event.
func (p *TaskEventPublisher) PublishTaskCreated(ctx context.Context, task *models.BackgroundTask) error {
	return p.Publish(ctx, NewBackgroundTaskEvent(TaskEventTypeCreated, task))
}

// PublishTaskStarted publishes a task started event.
func (p *TaskEventPublisher) PublishTaskStarted(ctx context.Context, task *models.BackgroundTask) error {
	return p.Publish(ctx, NewBackgroundTaskEvent(TaskEventTypeStarted, task))
}

// PublishTaskProgress publishes a task progress event.
func (p *TaskEventPublisher) PublishTaskProgress(ctx context.Context, task *models.BackgroundTask) error {
	return p.Publish(ctx, NewBackgroundTaskEvent(TaskEventTypeProgress, task))
}

// PublishTaskCompleted publishes a task completed event.
func (p *TaskEventPublisher) PublishTaskCompleted(ctx context.Context, task *models.BackgroundTask, result interface{}) error {
	event := NewBackgroundTaskEvent(TaskEventTypeCompleted, task)
	if result != nil {
		event.Metadata["result"] = result
	}
	return p.Publish(ctx, event)
}

// PublishTaskFailed publishes a task failed event.
func (p *TaskEventPublisher) PublishTaskFailed(ctx context.Context, task *models.BackgroundTask, err error) error {
	event := NewBackgroundTaskEvent(TaskEventTypeFailed, task)
	if err != nil {
		event.Error = err.Error()
		event.Metadata["error_details"] = err.Error()
	}
	return p.Publish(ctx, event)
}

// PublishTaskStuck publishes a task stuck event.
func (p *TaskEventPublisher) PublishTaskStuck(ctx context.Context, task *models.BackgroundTask, reason string) error {
	event := NewBackgroundTaskEvent(TaskEventTypeStuck, task)
	event.Metadata["reason"] = reason
	return p.Publish(ctx, event)
}

// PublishTaskCancelled publishes a task cancelled event.
func (p *TaskEventPublisher) PublishTaskCancelled(ctx context.Context, task *models.BackgroundTask) error {
	return p.Publish(ctx, NewBackgroundTaskEvent(TaskEventTypeCancelled, task))
}

// PublishTaskRetrying publishes a task retrying event.
func (p *TaskEventPublisher) PublishTaskRetrying(ctx context.Context, task *models.BackgroundTask, delay time.Duration) error {
	event := NewBackgroundTaskEvent(TaskEventTypeRetrying, task)
	event.Metadata["retry_delay_ms"] = delay.Milliseconds()
	event.Metadata["next_attempt"] = time.Now().Add(delay)
	return p.Publish(ctx, event)
}

// PublishTaskDeadLetter publishes a task dead letter event.
func (p *TaskEventPublisher) PublishTaskDeadLetter(ctx context.Context, task *models.BackgroundTask, reason string) error {
	event := NewBackgroundTaskEvent(TaskEventTypeDeadLetter, task)
	event.Metadata["reason"] = reason
	return p.Publish(ctx, event)
}

// IsEnabled returns true if publishing is enabled.
func (p *TaskEventPublisher) IsEnabled() bool {
	if p.extracted != nil {
		return p.extracted.IsEnabled()
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.enabled
}

// SetEnabled enables or disables publishing.
func (p *TaskEventPublisher) SetEnabled(enabled bool) {
	if p.extracted != nil {
		p.extracted.SetEnabled(enabled)
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.enabled = enabled
}

// generateEventID generates a unique event ID.
func generateEventID() string {
	return time.Now().UTC().Format("20060102150405.000000000") + "-" + randomEventString(8)
}

// randomEventString generates a random string for event IDs.
func randomEventString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		time.Sleep(time.Nanosecond)
	}
	return string(b)
}
