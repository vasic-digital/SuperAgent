package bigdata

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"dev.helix.agent/pkg/messaging"
	"github.com/sirupsen/logrus"
)

// AnalyticsIntegration sends metrics to ClickHouse for time-series analytics
type AnalyticsIntegration struct {
	kafkaBroker messaging.MessageBroker
	logger      *logrus.Logger
	enabled     bool
}

// NewAnalyticsIntegration creates a new analytics integration
func NewAnalyticsIntegration(
	kafkaBroker messaging.MessageBroker,
	logger *logrus.Logger,
	enabled bool,
) *AnalyticsIntegration {
	return &AnalyticsIntegration{
		kafkaBroker: kafkaBroker,
		logger:      logger,
		enabled:     enabled,
	}
}

// ProviderMetrics represents metrics for a single LLM provider request
type ProviderMetrics struct {
	Provider         string    `json:"provider"`
	Model            string    `json:"model"`
	RequestID        string    `json:"request_id"`
	DebateID         string    `json:"debate_id,omitempty"`
	Round            int       `json:"round,omitempty"`
	Timestamp        time.Time `json:"timestamp"`
	ResponseTimeMs   float64   `json:"response_time_ms"`
	TokensUsed       int       `json:"tokens_used"`
	PromptTokens     int       `json:"prompt_tokens,omitempty"`
	CompletionTokens int       `json:"completion_tokens,omitempty"`
	Confidence       float64   `json:"confidence,omitempty"`
	Success          bool      `json:"success"`
	ErrorType        string    `json:"error_type,omitempty"`
	ErrorCount       int       `json:"error_count,omitempty"`
}

// DebateMetrics represents metrics for a complete debate
type DebateMetrics struct {
	DebateID         string    `json:"debate_id"`
	Topic            string    `json:"topic"`
	Timestamp        time.Time `json:"timestamp"`
	TotalRounds      int       `json:"total_rounds"`
	TotalDurationMs  float64   `json:"total_duration_ms"`
	ParticipantCount int       `json:"participant_count"`
	Winner           string    `json:"winner,omitempty"`
	WinnerProvider   string    `json:"winner_provider,omitempty"`
	WinnerModel      string    `json:"winner_model,omitempty"`
	Confidence       float64   `json:"confidence,omitempty"`
	TotalTokens      int       `json:"total_tokens"`
	Outcome          string    `json:"outcome"` // successful, abandoned, error
}

// ConversationMetrics represents metrics for a conversation
type ConversationMetrics struct {
	ConversationID   string    `json:"conversation_id"`
	UserID           string    `json:"user_id"`
	SessionID        string    `json:"session_id"`
	Timestamp        time.Time `json:"timestamp"`
	MessageCount     int       `json:"message_count"`
	EntityCount      int       `json:"entity_count"`
	TotalTokens      int64     `json:"total_tokens"`
	Compressed       bool      `json:"compressed"`
	CompressionRatio float64   `json:"compression_ratio,omitempty"`
	DebateCount      int       `json:"debate_count"`
}

// PublishProviderMetrics publishes provider performance metrics
func (ai *AnalyticsIntegration) PublishProviderMetrics(ctx context.Context, metrics *ProviderMetrics) error {
	if !ai.enabled {
		return nil
	}

	event := &AnalyticsEvent{
		EventID:   generateEventID(),
		EventType: "provider.metrics",
		Timestamp: metrics.Timestamp,
		Data:      metrics,
	}

	return ai.publishAnalyticsEvent(ctx, event, "helixagent.analytics.providers")
}

// PublishDebateMetrics publishes debate performance metrics
func (ai *AnalyticsIntegration) PublishDebateMetrics(ctx context.Context, metrics *DebateMetrics) error {
	if !ai.enabled {
		return nil
	}

	event := &AnalyticsEvent{
		EventID:   generateEventID(),
		EventType: "debate.metrics",
		Timestamp: metrics.Timestamp,
		Data:      metrics,
	}

	return ai.publishAnalyticsEvent(ctx, event, "helixagent.analytics.debates")
}

// PublishConversationMetrics publishes conversation metrics
func (ai *AnalyticsIntegration) PublishConversationMetrics(ctx context.Context, metrics *ConversationMetrics) error {
	if !ai.enabled {
		return nil
	}

	event := &AnalyticsEvent{
		EventID:   generateEventID(),
		EventType: "conversation.metrics",
		Timestamp: metrics.Timestamp,
		Data:      metrics,
	}

	return ai.publishAnalyticsEvent(ctx, event, "helixagent.analytics.conversations")
}

// AnalyticsEvent represents a generic analytics event
type AnalyticsEvent struct {
	EventID   string      `json:"event_id"`
	EventType string      `json:"event_type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// publishAnalyticsEvent publishes an analytics event to Kafka
func (ai *AnalyticsIntegration) publishAnalyticsEvent(ctx context.Context, event *AnalyticsEvent, topic string) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal analytics event: %w", err)
	}

	msg := &messaging.Message{
		Topic:     topic,
		Key:       event.EventID,
		Payload:   payload,
		Timestamp: event.Timestamp,
		Headers: map[string]string{
			"event_type": event.EventType,
		},
	}

	if err := ai.kafkaBroker.Publish(ctx, msg.Topic, msg); err != nil {
		return fmt.Errorf("kafka publish failed: %w", err)
	}

	ai.logger.WithFields(logrus.Fields{
		"event_type": event.EventType,
		"topic":      topic,
	}).Debug("Published analytics event to Kafka")

	return nil
}

// RecordProviderRequest records a provider API request
func (ai *AnalyticsIntegration) RecordProviderRequest(
	ctx context.Context,
	provider, model, requestID string,
	responseTime time.Duration,
	tokensUsed int,
	success bool,
	errorType string,
) error {
	metrics := &ProviderMetrics{
		Provider:       provider,
		Model:          model,
		RequestID:      requestID,
		Timestamp:      time.Now(),
		ResponseTimeMs: float64(responseTime.Milliseconds()),
		TokensUsed:     tokensUsed,
		Success:        success,
		ErrorType:      errorType,
	}

	if !success {
		metrics.ErrorCount = 1
	}

	return ai.PublishProviderMetrics(ctx, metrics)
}

// RecordDebateRound records a single debate round
func (ai *AnalyticsIntegration) RecordDebateRound(
	ctx context.Context,
	debateID, provider, model string,
	round int,
	responseTime time.Duration,
	tokensUsed int,
	confidence float64,
) error {
	metrics := &ProviderMetrics{
		Provider:       provider,
		Model:          model,
		DebateID:       debateID,
		Round:          round,
		Timestamp:      time.Now(),
		ResponseTimeMs: float64(responseTime.Milliseconds()),
		TokensUsed:     tokensUsed,
		Confidence:     confidence,
		Success:        true,
	}

	return ai.PublishProviderMetrics(ctx, metrics)
}

// RecordDebateCompletion records a completed debate
func (ai *AnalyticsIntegration) RecordDebateCompletion(
	ctx context.Context,
	debateID, topic string,
	rounds int,
	duration time.Duration,
	participantCount int,
	winner, winnerProvider, winnerModel string,
	confidence float64,
	totalTokens int,
	outcome string,
) error {
	metrics := &DebateMetrics{
		DebateID:         debateID,
		Topic:            topic,
		Timestamp:        time.Now(),
		TotalRounds:      rounds,
		TotalDurationMs:  float64(duration.Milliseconds()),
		ParticipantCount: participantCount,
		Winner:           winner,
		WinnerProvider:   winnerProvider,
		WinnerModel:      winnerModel,
		Confidence:       confidence,
		TotalTokens:      totalTokens,
		Outcome:          outcome,
	}

	return ai.PublishDebateMetrics(ctx, metrics)
}

// RecordConversation records conversation statistics
func (ai *AnalyticsIntegration) RecordConversation(
	ctx context.Context,
	conversationID, userID, sessionID string,
	messageCount, entityCount int,
	totalTokens int64,
	compressed bool,
	compressionRatio float64,
	debateCount int,
) error {
	metrics := &ConversationMetrics{
		ConversationID:   conversationID,
		UserID:           userID,
		SessionID:        sessionID,
		Timestamp:        time.Now(),
		MessageCount:     messageCount,
		EntityCount:      entityCount,
		TotalTokens:      totalTokens,
		Compressed:       compressed,
		CompressionRatio: compressionRatio,
		DebateCount:      debateCount,
	}

	return ai.PublishConversationMetrics(ctx, metrics)
}

// BatchPublishProviderMetrics publishes multiple provider metrics in a batch
func (ai *AnalyticsIntegration) BatchPublishProviderMetrics(ctx context.Context, metricsBatch []*ProviderMetrics) error {
	if !ai.enabled {
		return nil
	}

	for _, metrics := range metricsBatch {
		if err := ai.PublishProviderMetrics(ctx, metrics); err != nil {
			ai.logger.WithError(err).
				WithField("provider", metrics.Provider).
				Warn("Failed to publish provider metrics")
			// Continue with other metrics
		}
	}

	return nil
}
