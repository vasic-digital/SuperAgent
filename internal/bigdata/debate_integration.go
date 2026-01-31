package bigdata

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"dev.helix.agent/internal/conversation"
	"dev.helix.agent/pkg/messaging"
	"github.com/sirupsen/logrus"
)

// DebateIntegration connects the AI debate system with big data components
type DebateIntegration struct {
	infiniteContext *conversation.InfiniteContextEngine
	kafkaBroker     messaging.MessageBroker
	logger          *logrus.Logger
}

// NewDebateIntegration creates a new debate integration
func NewDebateIntegration(
	infiniteContext *conversation.InfiniteContextEngine,
	kafkaBroker messaging.MessageBroker,
	logger *logrus.Logger,
) *DebateIntegration {
	return &DebateIntegration{
		infiniteContext: infiniteContext,
		kafkaBroker:     kafkaBroker,
		logger:          logger,
	}
}

// GetConversationContext retrieves unlimited conversation context
func (di *DebateIntegration) GetConversationContext(
	ctx context.Context,
	conversationID string,
	maxTokens int,
) (*ConversationContext, error) {
	// Replay conversation from Kafka
	context, err := di.infiniteContext.ReplayConversation(ctx, conversationID)
	if err != nil {
		di.logger.WithError(err).
			WithField("conversation_id", conversationID).
			Error("Failed to replay conversation")
		return nil, fmt.Errorf("context replay failed: %w", err)
	}

	// Check if compression is needed
	if context.TotalTokens > maxTokens {
		di.logger.WithFields(logrus.Fields{
			"conversation_id": conversationID,
			"total_tokens":    context.TotalTokens,
			"max_tokens":      maxTokens,
		}).Info("Compressing conversation context")

		// Compress context
		compressed, err := di.infiniteContext.ReplayWithCompression(
			ctx,
			conversationID,
			maxTokens,
		)
		if err != nil {
			di.logger.WithError(err).Warn("Compression failed, using original")
		} else {
			context = compressed
		}
	}

	return &ConversationContext{
		ConversationID:   conversationID,
		Messages:         context.Messages,
		Entities:         context.Entities,
		TotalTokens:      context.TotalTokens,
		Compressed:       context.Compressed,
		CompressionStats: context.CompressionStats,
	}, nil
}

// PublishDebateCompletion publishes a debate completion event to Kafka
func (di *DebateIntegration) PublishDebateCompletion(
	ctx context.Context,
	completion *DebateCompletion,
) error {
	payload, err := json.Marshal(completion)
	if err != nil {
		return fmt.Errorf("failed to marshal completion: %w", err)
	}

	msg := &messaging.Message{
		Topic:     "helixagent.debates.completed",
		Key:       completion.DebateID,
		Payload:   payload,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"debate_id":       completion.DebateID,
			"conversation_id": completion.ConversationID,
			"user_id":         completion.UserID,
		},
	}

	if err := di.kafkaBroker.Publish(ctx, msg.Topic, msg); err != nil {
		di.logger.WithError(err).
			WithField("debate_id", completion.DebateID).
			Error("Failed to publish debate completion")
		return fmt.Errorf("publish failed: %w", err)
	}

	di.logger.WithFields(logrus.Fields{
		"debate_id":       completion.DebateID,
		"conversation_id": completion.ConversationID,
		"rounds":          completion.Rounds,
	}).Debug("Published debate completion")

	return nil
}

// PublishConversationEvent publishes a conversation event to Kafka
func (di *DebateIntegration) PublishConversationEvent(
	ctx context.Context,
	event *ConversationEvent,
) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := &messaging.Message{
		Topic:     "helixagent.conversations",
		Key:       event.ConversationID,
		Payload:   payload,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"conversation_id": event.ConversationID,
			"event_type":      string(event.EventType),
		},
	}

	if err := di.kafkaBroker.Publish(ctx, msg.Topic, msg); err != nil {
		di.logger.WithError(err).
			WithField("conversation_id", event.ConversationID).
			Error("Failed to publish conversation event")
		return fmt.Errorf("publish failed: %w", err)
	}

	return nil
}

// ConversationContext represents conversation context with optional compression
type ConversationContext struct {
	ConversationID   string            `json:"conversation_id"`
	Messages         []Message         `json:"messages"`
	Entities         []Entity          `json:"entities"`
	TotalTokens      int               `json:"total_tokens"`
	Compressed       bool              `json:"compressed"`
	CompressionStats *CompressionStats `json:"compression_stats,omitempty"`
}

// Message represents a conversation message
type Message struct {
	Role      string                 `json:"role"`
	Content   string                 `json:"content"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Entity represents an extracted entity
type Entity struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Importance float64                `json:"importance"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// CompressionStats holds compression statistics
type CompressionStats struct {
	Strategy           string        `json:"strategy"`
	OriginalMessages   int           `json:"original_messages"`
	CompressedMessages int           `json:"compressed_messages"`
	OriginalTokens     int           `json:"original_tokens"`
	CompressedTokens   int           `json:"compressed_tokens"`
	CompressionRatio   float64       `json:"compression_ratio"`
	QualityScore       float64       `json:"quality_score"`
	Duration           time.Duration `json:"duration"`
}

// DebateCompletion represents a completed debate
type DebateCompletion struct {
	DebateID       string                 `json:"debate_id"`
	ConversationID string                 `json:"conversation_id"`
	UserID         string                 `json:"user_id"`
	SessionID      string                 `json:"session_id"`
	Topic          string                 `json:"topic"`
	Rounds         int                    `json:"rounds"`
	Winner         string                 `json:"winner"`
	WinnerProvider string                 `json:"winner_provider"`
	WinnerModel    string                 `json:"winner_model"`
	Confidence     float64                `json:"confidence"`
	Duration       time.Duration          `json:"duration"`
	StartedAt      time.Time              `json:"started_at"`
	CompletedAt    time.Time              `json:"completed_at"`
	Participants   []DebateParticipant    `json:"participants"`
	Outcome        string                 `json:"outcome"` // successful, abandoned, error
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// DebateParticipant represents a debate participant
type DebateParticipant struct {
	Provider     string  `json:"provider"`
	Model        string  `json:"model"`
	Position     string  `json:"position"`
	ResponseTime int     `json:"response_time_ms"`
	TokensUsed   int     `json:"tokens_used"`
	Confidence   float64 `json:"confidence"`
	Won          bool    `json:"won"`
}

// ConversationEvent represents a conversation event
type ConversationEvent struct {
	EventID        string                 `json:"event_id"`
	ConversationID string                 `json:"conversation_id"`
	EventType      ConversationEventType  `json:"event_type"`
	Timestamp      time.Time              `json:"timestamp"`
	Data           map[string]interface{} `json:"data"`
}

// ConversationEventType represents the type of conversation event
type ConversationEventType string

const (
	ConversationEventMessageAdded      ConversationEventType = "message.added"
	ConversationEventEntityExtracted   ConversationEventType = "entity.extracted"
	ConversationEventDebateStarted     ConversationEventType = "debate.started"
	ConversationEventDebateCompleted   ConversationEventType = "debate.completed"
	ConversationEventContextCompressed ConversationEventType = "context.compressed"
)
