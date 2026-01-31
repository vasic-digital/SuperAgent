package bigdata

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"dev.helix.agent/internal/conversation"
	"dev.helix.agent/internal/messaging"
	"github.com/google/uuid"
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
	// Get conversation snapshot (includes messages, entities, context)
	snapshot, err := di.infiniteContext.GetConversationSnapshot(ctx, conversationID)
	if err != nil {
		di.logger.WithError(err).
			WithField("conversation_id", conversationID).
			Error("Failed to get conversation snapshot")
		return nil, fmt.Errorf("context snapshot failed: %w", err)
	}

	// Check if compression is needed
	if snapshot.Context != nil && snapshot.Context.TotalTokens > int64(maxTokens) {
		di.logger.WithFields(logrus.Fields{
			"conversation_id": conversationID,
			"total_tokens":    snapshot.Context.TotalTokens,
			"max_tokens":      maxTokens,
		}).Info("Compressing conversation context")

		// Compress context
		compressedMessages, compressionData, err := di.infiniteContext.ReplayWithCompression(
			ctx,
			conversationID,
			maxTokens,
		)
		if err != nil {
			di.logger.WithError(err).Warn("Compression failed, using original")
		} else {
			// Update snapshot messages with compressed messages
			snapshot.Messages = compressedMessages
			// Update context with compression data
			if compressionData != nil {
				snapshot.Context.CompressedCount = compressionData.CompressedMessages
				snapshot.Context.CompressionRatio = compressionData.CompressionRatio
			}
		}
	}

	// Convert conversation snapshot to ConversationContext
	result := &ConversationContext{
		ConversationID: conversationID,
		Messages:       di.convertMessages(snapshot.Messages),
		Entities:       di.convertEntities(snapshot.Entities),
		Compressed:     snapshot.Context != nil && snapshot.Context.CompressedCount > 0,
	}

	// Set total tokens
	if snapshot.Context != nil {
		result.TotalTokens = int(snapshot.Context.TotalTokens)
		// Set compression stats if compression occurred
		if snapshot.Context.CompressedCount > 0 {
			result.CompressionStats = &CompressionStats{
				Strategy:           "adaptive", // Default strategy
				OriginalMessages:   snapshot.Context.MessageCount,
				CompressedMessages: snapshot.Context.CompressedCount,
				OriginalTokens:     int(snapshot.Context.TotalTokens),
				CompressedTokens:   int(snapshot.Context.TotalTokens), // Approximate
				CompressionRatio:   snapshot.Context.CompressionRatio,
				QualityScore:       0.9, // Default quality score
				Duration:           0,   // Unknown
			}
		}
	}

	return result, nil
}

// convertMessages converts conversation.MessageData to bigdata.Message
func (di *DebateIntegration) convertMessages(messages []conversation.MessageData) []Message {
	result := make([]Message, len(messages))
	for i, msg := range messages {
		result[i] = Message{
			Role:      msg.Role,
			Content:   msg.Content,
			Timestamp: msg.CreatedAt,
			Metadata: map[string]interface{}{
				"message_id": msg.MessageID,
				"model":      msg.Model,
				"tokens":     msg.Tokens,
			},
		}
	}
	return result
}

// convertEntities converts conversation.EntityData to bigdata.Entity
func (di *DebateIntegration) convertEntities(entities []conversation.EntityData) []Entity {
	result := make([]Entity, len(entities))
	for i, entity := range entities {
		result[i] = Entity{
			ID:         entity.EntityID,
			Name:       entity.Name,
			Type:       entity.Type,
			Importance: entity.Confidence,
			Properties: entity.Properties,
		}
	}
	return result
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

	msg := messaging.NewMessage("debate.completed", payload)
	msg.ID = uuid.New().String()
	msg.Headers["debate_id"] = completion.DebateID
	msg.Headers["conversation_id"] = completion.ConversationID
	msg.Headers["user_id"] = completion.UserID

	topic := "helixagent.debates.completed"
	if err := di.kafkaBroker.Publish(ctx, topic, msg); err != nil {
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

	msg := messaging.NewMessage("conversation.event", payload)
	msg.ID = uuid.New().String()
	msg.Headers["conversation_id"] = event.ConversationID
	msg.Headers["event_type"] = string(event.EventType)

	topic := "helixagent.conversations"
	if err := di.kafkaBroker.Publish(ctx, topic, msg); err != nil {
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
