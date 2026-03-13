// Package background provides adapters bridging HelixAgent's internal background
// types to the generic digital.vasic.background module.
package background

import (
	"context"
	"encoding/json"

	"github.com/sirupsen/logrus"

	"dev.helix.agent/internal/messaging"
	extractedbackground "digital.vasic.background"
)

// MessagingHubEventPublisher adapts messaging.MessagingHub to extractedbackground.EventPublisher
type MessagingHubEventPublisher struct {
	hub    *messaging.MessagingHub
	logger *logrus.Logger
}

// NewMessagingHubEventPublisher creates a new adapter that publishes events via messaging hub
func NewMessagingHubEventPublisher(hub *messaging.MessagingHub, logger *logrus.Logger) *MessagingHubEventPublisher {
	if logger == nil {
		logger = logrus.New()
	}
	return &MessagingHubEventPublisher{
		hub:    hub,
		logger: logger,
	}
}

// Publish publishes a background task event to the messaging hub
func (m *MessagingHubEventPublisher) Publish(ctx context.Context, event *extractedbackground.BackgroundTaskEvent) error {
	if event == nil || m.hub == nil {
		return nil
	}

	// Convert extracted BackgroundTaskEvent to messaging.Event
	msgEvent := convertToMessagingEvent(event)
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

// convertToMessagingEvent converts extracted BackgroundTaskEvent to messaging.Event
func convertToMessagingEvent(extracted *extractedbackground.BackgroundTaskEvent) *messaging.Event {
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

// LoggerAdapter adapts logrus.Logger to extractedbackground.Logger interface
type LoggerAdapter struct {
	logger *logrus.Logger
}

// NewLoggerAdapter creates a new logger adapter
func NewLoggerAdapter(logger *logrus.Logger) *LoggerAdapter {
	if logger == nil {
		logger = logrus.New()
	}
	return &LoggerAdapter{logger: logger}
}

// Debugf logs debug message
func (l *LoggerAdapter) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}

// Infof logs info message
func (l *LoggerAdapter) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

// Warnf logs warning message
func (l *LoggerAdapter) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}

// Errorf logs error message
func (l *LoggerAdapter) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}
