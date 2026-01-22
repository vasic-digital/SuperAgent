// Package notifications provides Kafka transport for notification persistence.
package notifications

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"dev.helix.agent/internal/messaging"
)

// Kafka topic constants for notifications.
const (
	TopicNotifications     = "helixagent.notifications"
	TopicTaskNotifications = "helixagent.notifications.tasks"
	TopicAuditLog          = "helixagent.events.audit"
)

// KafkaNotificationEvent represents a notification event for Kafka.
type KafkaNotificationEvent struct {
	// ID is the unique event identifier.
	ID string `json:"id"`
	// TaskID is the task this notification relates to.
	TaskID string `json:"task_id"`
	// EventType is the notification event type.
	EventType string `json:"event_type"`
	// Timestamp is when the event occurred.
	Timestamp time.Time `json:"timestamp"`
	// Data contains the notification payload.
	Data map[string]interface{} `json:"data,omitempty"`
	// Source identifies the notification source.
	Source string `json:"source"`
	// CorrelationID for request tracing.
	CorrelationID string `json:"correlation_id,omitempty"`
}

// KafkaTransportConfig holds configuration for the Kafka transport.
type KafkaTransportConfig struct {
	// Enabled enables Kafka transport.
	Enabled bool `json:"enabled" yaml:"enabled"`
	// Topic is the Kafka topic for notifications.
	Topic string `json:"topic" yaml:"topic"`
	// Async enables async publishing.
	Async bool `json:"async" yaml:"async"`
	// BufferSize is the async buffer size.
	BufferSize int `json:"buffer_size" yaml:"buffer_size"`
	// EnableAuditLog enables audit log publishing.
	EnableAuditLog bool `json:"enable_audit_log" yaml:"enable_audit_log"`
}

// DefaultKafkaTransportConfig returns default configuration.
func DefaultKafkaTransportConfig() *KafkaTransportConfig {
	return &KafkaTransportConfig{
		Enabled:        true,
		Topic:          TopicTaskNotifications,
		Async:          true,
		BufferSize:     1000,
		EnableAuditLog: true,
	}
}

// KafkaTransport publishes notifications to Kafka for persistence and replay.
type KafkaTransport struct {
	hub     *messaging.MessagingHub
	config  *KafkaTransportConfig
	logger  *logrus.Logger
	eventCh chan *KafkaNotificationEvent
	stopCh  chan struct{}
	wg      sync.WaitGroup
	mu      sync.Mutex
	started bool
}

// NewKafkaTransport creates a new Kafka transport.
func NewKafkaTransport(
	hub *messaging.MessagingHub,
	logger *logrus.Logger,
	config *KafkaTransportConfig,
) *KafkaTransport {
	if config == nil {
		config = DefaultKafkaTransportConfig()
	}
	if logger == nil {
		logger = logrus.New()
	}

	t := &KafkaTransport{
		hub:    hub,
		config: config,
		logger: logger,
		stopCh: make(chan struct{}),
	}

	if config.Async && config.BufferSize > 0 {
		t.eventCh = make(chan *KafkaNotificationEvent, config.BufferSize)
	}

	return t
}

// Start starts the async publishing goroutine if enabled.
func (t *KafkaTransport) Start() {
	t.mu.Lock()
	if t.started || !t.config.Async || t.eventCh == nil {
		t.mu.Unlock()
		return
	}
	t.started = true
	t.mu.Unlock()

	t.wg.Add(1)
	go t.asyncPublishLoop()
}

// Stop stops the transport and flushes remaining events.
func (t *KafkaTransport) Stop() {
	close(t.stopCh)
	if t.eventCh != nil {
		close(t.eventCh)
	}
	t.wg.Wait()
}

// asyncPublishLoop processes async publish events.
func (t *KafkaTransport) asyncPublishLoop() {
	defer t.wg.Done()

	for {
		select {
		case event, ok := <-t.eventCh:
			if !ok {
				return
			}
			t.doPublish(context.Background(), event)
		case <-t.stopCh:
			// Drain remaining events
			for event := range t.eventCh {
				t.doPublish(context.Background(), event)
			}
			return
		}
	}
}

// Send publishes a notification to Kafka.
func (t *KafkaTransport) Send(ctx context.Context, notification *TaskNotification) error {
	if !t.config.Enabled || t.hub == nil {
		return nil
	}

	event := &KafkaNotificationEvent{
		ID:        generateNotificationEventID(),
		TaskID:    notification.TaskID,
		EventType: notification.EventType,
		Timestamp: notification.Timestamp,
		Data:      notification.Data,
		Source:    "helixagent.notifications",
	}

	// Add task data if available
	if notification.Task != nil {
		if event.Data == nil {
			event.Data = make(map[string]interface{})
		}
		event.Data["task_status"] = notification.Task.Status
		event.Data["task_type"] = notification.Task.TaskType
		event.Data["task_name"] = notification.Task.TaskName
		event.Data["progress"] = notification.Task.Progress
	}

	return t.publish(ctx, event)
}

// PublishAuditEvent publishes an audit log event.
func (t *KafkaTransport) PublishAuditEvent(ctx context.Context, action, resource string, details map[string]interface{}) error {
	if !t.config.Enabled || !t.config.EnableAuditLog || t.hub == nil {
		return nil
	}

	event := &KafkaNotificationEvent{
		ID:        generateNotificationEventID(),
		EventType: "audit." + action,
		Timestamp: time.Now().UTC(),
		Data: map[string]interface{}{
			"action":   action,
			"resource": resource,
			"details":  details,
		},
		Source: "helixagent.audit",
	}

	return t.publishToTopic(ctx, TopicAuditLog, event)
}

// publish publishes an event (async or sync).
func (t *KafkaTransport) publish(ctx context.Context, event *KafkaNotificationEvent) error {
	if t.config.Async && t.eventCh != nil {
		select {
		case t.eventCh <- event:
			return nil
		default:
			// Buffer full, publish synchronously
			t.logger.Warn("Notification event buffer full, publishing synchronously")
			return t.doPublish(ctx, event)
		}
	}

	return t.doPublish(ctx, event)
}

// publishToTopic publishes an event to a specific topic.
func (t *KafkaTransport) publishToTopic(ctx context.Context, topic string, event *KafkaNotificationEvent) error {
	if event == nil {
		return nil
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.logger.WithError(err).Error("Failed to marshal notification event")
		return err
	}

	msgEvent := &messaging.Event{
		ID:         event.ID,
		Type:       messaging.EventType(event.EventType),
		Source:     event.Source,
		Subject:    event.TaskID,
		Data:       data,
		DataSchema: "application/json",
		Timestamp:  event.Timestamp,
	}

	if err := t.hub.PublishEvent(ctx, topic, msgEvent); err != nil {
		t.logger.WithError(err).WithFields(logrus.Fields{
			"task_id":    event.TaskID,
			"event_type": event.EventType,
			"topic":      topic,
		}).Error("Failed to publish notification event")
		return err
	}

	return nil
}

// doPublish performs the actual event publish.
func (t *KafkaTransport) doPublish(ctx context.Context, event *KafkaNotificationEvent) error {
	return t.publishToTopic(ctx, t.config.Topic, event)
}

// IsEnabled returns whether Kafka transport is enabled.
func (t *KafkaTransport) IsEnabled() bool {
	return t.config.Enabled && t.hub != nil
}

// generateNotificationEventID generates a unique event ID.
func generateNotificationEventID() string {
	return time.Now().UTC().Format("20060102150405.000000000") + "-" + randomNotificationString(8)
}

// randomNotificationString generates a random string.
func randomNotificationString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		time.Sleep(time.Nanosecond)
	}
	return string(b)
}

// NotificationHubKafkaIntegration provides Kafka integration for NotificationHub.
type NotificationHubKafkaIntegration struct {
	hub       *NotificationHub
	transport *KafkaTransport
	logger    *logrus.Logger
}

// NewNotificationHubKafkaIntegration creates a new Kafka integration for the notification hub.
func NewNotificationHubKafkaIntegration(
	notificationHub *NotificationHub,
	messagingHub *messaging.MessagingHub,
	logger *logrus.Logger,
	config *KafkaTransportConfig,
) *NotificationHubKafkaIntegration {
	if logger == nil {
		logger = logrus.New()
	}
	transport := NewKafkaTransport(messagingHub, logger, config)

	return &NotificationHubKafkaIntegration{
		hub:       notificationHub,
		transport: transport,
		logger:    logger,
	}
}

// Start starts the Kafka integration.
func (i *NotificationHubKafkaIntegration) Start() {
	i.transport.Start()
	i.logger.Info("Notification hub Kafka integration started")
}

// Stop stops the Kafka integration.
func (i *NotificationHubKafkaIntegration) Stop() {
	i.transport.Stop()
	i.logger.Info("Notification hub Kafka integration stopped")
}

// PublishNotification publishes a notification to Kafka.
func (i *NotificationHubKafkaIntegration) PublishNotification(ctx context.Context, notification *TaskNotification) error {
	return i.transport.Send(ctx, notification)
}

// PublishAudit publishes an audit event.
func (i *NotificationHubKafkaIntegration) PublishAudit(ctx context.Context, action, resource string, details map[string]interface{}) error {
	return i.transport.PublishAuditEvent(ctx, action, resource, details)
}

// Transport returns the underlying Kafka transport.
func (i *NotificationHubKafkaIntegration) Transport() *KafkaTransport {
	return i.transport
}
