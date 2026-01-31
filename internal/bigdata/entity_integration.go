package bigdata

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"dev.helix.agent/internal/memory"
	"dev.helix.agent/internal/messaging"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// EntityIntegration publishes entity updates to the knowledge graph
type EntityIntegration struct {
	kafkaBroker messaging.MessageBroker
	logger      *logrus.Logger
	enabled     bool
}

// NewEntityIntegration creates a new entity integration
func NewEntityIntegration(
	kafkaBroker messaging.MessageBroker,
	logger *logrus.Logger,
	enabled bool,
) *EntityIntegration {
	return &EntityIntegration{
		kafkaBroker: kafkaBroker,
		logger:      logger,
		enabled:     enabled,
	}
}

// PublishEntityCreated publishes an entity creation event
func (ei *EntityIntegration) PublishEntityCreated(ctx context.Context, entity *memory.Entity, conversationID string) error {
	if !ei.enabled {
		return nil
	}

	event := &EntityUpdateEvent{
		EventID:        generateEventID(),
		EventType:      "entity.created",
		Timestamp:      time.Now(),
		ConversationID: conversationID,
		Entity:         entity,
	}

	return ei.publishEntityEvent(ctx, event, "helixagent.entities.updates")
}

// PublishEntityUpdated publishes an entity update event
func (ei *EntityIntegration) PublishEntityUpdated(ctx context.Context, entity *memory.Entity, conversationID string) error {
	if !ei.enabled {
		return nil
	}

	event := &EntityUpdateEvent{
		EventID:        generateEventID(),
		EventType:      "entity.updated",
		Timestamp:      time.Now(),
		ConversationID: conversationID,
		Entity:         entity,
	}

	return ei.publishEntityEvent(ctx, event, "helixagent.entities.updates")
}

// PublishRelationshipCreated publishes a relationship creation event
func (ei *EntityIntegration) PublishRelationshipCreated(ctx context.Context, relationship *memory.Relationship, conversationID string) error {
	if !ei.enabled {
		return nil
	}

	event := &RelationshipUpdateEvent{
		EventID:        generateEventID(),
		EventType:      "relationship.created",
		Timestamp:      time.Now(),
		ConversationID: conversationID,
		Relationship:   relationship,
	}

	return ei.publishRelationshipEvent(ctx, event, "helixagent.relationships.updates")
}

// PublishEntitiesBatch publishes a batch of entity updates
func (ei *EntityIntegration) PublishEntitiesBatch(ctx context.Context, entities []memory.Entity, conversationID string) error {
	if !ei.enabled {
		return nil
	}

	for _, entity := range entities {
		e := entity // Create copy for pointer
		if err := ei.PublishEntityCreated(ctx, &e, conversationID); err != nil {
			ei.logger.WithError(err).
				WithField("entity_id", entity.ID).
				Warn("Failed to publish entity creation event")
			// Continue with other entities
		}
	}

	return nil
}

// PublishRelationshipsBatch publishes a batch of relationship updates
func (ei *EntityIntegration) PublishRelationshipsBatch(ctx context.Context, relationships []memory.Relationship, conversationID string) error {
	if !ei.enabled {
		return nil
	}

	for _, relationship := range relationships {
		r := relationship // Create copy for pointer
		if err := ei.PublishRelationshipCreated(ctx, &r, conversationID); err != nil {
			ei.logger.WithError(err).
				WithField("source_id", relationship.SourceID).
				WithField("target_id", relationship.TargetID).
				Warn("Failed to publish relationship creation event")
			// Continue with other relationships
		}
	}

	return nil
}

// EntityUpdateEvent represents an entity update for the knowledge graph
type EntityUpdateEvent struct {
	EventID        string         `json:"event_id"`
	EventType      string         `json:"event_type"` // entity.created, entity.updated, entity.merged
	Timestamp      time.Time      `json:"timestamp"`
	ConversationID string         `json:"conversation_id"`
	Entity         *memory.Entity `json:"entity"`
}

// RelationshipUpdateEvent represents a relationship update for the knowledge graph
type RelationshipUpdateEvent struct {
	EventID        string               `json:"event_id"`
	EventType      string               `json:"event_type"` // relationship.created, relationship.updated
	Timestamp      time.Time            `json:"timestamp"`
	ConversationID string               `json:"conversation_id"`
	Relationship   *memory.Relationship `json:"relationship"`
}

// publishEntityEvent publishes an entity event to Kafka
func (ei *EntityIntegration) publishEntityEvent(ctx context.Context, event *EntityUpdateEvent, topic string) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal entity event: %w", err)
	}

	msg := messaging.NewMessage("entity.update", payload)
	msg.ID = uuid.New().String()
	msg.Timestamp = event.Timestamp
	msg.Headers["event_type"] = event.EventType
	msg.Headers["conversation_id"] = event.ConversationID
	msg.Headers["entity_id"] = event.Entity.ID
	msg.Headers["entity_type"] = event.Entity.Type

	if err := ei.kafkaBroker.Publish(ctx, topic, msg); err != nil {
		return fmt.Errorf("kafka publish failed: %w", err)
	}

	ei.logger.WithFields(logrus.Fields{
		"event_type":  event.EventType,
		"entity_id":   event.Entity.ID,
		"entity_type": event.Entity.Type,
	}).Debug("Published entity event to Kafka")

	return nil
}

// publishRelationshipEvent publishes a relationship event to Kafka
func (ei *EntityIntegration) publishRelationshipEvent(ctx context.Context, event *RelationshipUpdateEvent, topic string) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal relationship event: %w", err)
	}

	msg := messaging.NewMessage("relationship.update", payload)
	msg.ID = uuid.New().String()
	msg.Timestamp = event.Timestamp
	msg.Headers["event_type"] = event.EventType
	msg.Headers["conversation_id"] = event.ConversationID
	msg.Headers["source_id"] = event.Relationship.SourceID
	msg.Headers["target_id"] = event.Relationship.TargetID
	msg.Headers["relationship_type"] = event.Relationship.Type

	if err := ei.kafkaBroker.Publish(ctx, topic, msg); err != nil {
		return fmt.Errorf("kafka publish failed: %w", err)
	}

	ei.logger.WithFields(logrus.Fields{
		"event_type": event.EventType,
		"source_id":  event.Relationship.SourceID,
		"target_id":  event.Relationship.TargetID,
		"type":       event.Relationship.Type,
	}).Debug("Published relationship event to Kafka")

	return nil
}

// PublishEntityMerge publishes an entity merge event (when duplicate entities are detected)
func (ei *EntityIntegration) PublishEntityMerge(ctx context.Context, sourceEntity, targetEntity *memory.Entity, conversationID string) error {
	if !ei.enabled {
		return nil
	}

	event := &EntityMergeEvent{
		EventID:        generateEventID(),
		EventType:      "entity.merged",
		Timestamp:      time.Now(),
		ConversationID: conversationID,
		SourceEntity:   sourceEntity,
		TargetEntity:   targetEntity,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal entity merge event: %w", err)
	}

	msg := messaging.NewMessage("entity.merge", payload)
	msg.ID = uuid.New().String()
	msg.Timestamp = event.Timestamp
	msg.Headers["event_type"] = event.EventType
	msg.Headers["conversation_id"] = event.ConversationID
	msg.Headers["source_id"] = sourceEntity.ID
	msg.Headers["target_id"] = targetEntity.ID

	topic := "helixagent.entities.updates"
	if err := ei.kafkaBroker.Publish(ctx, topic, msg); err != nil {
		return fmt.Errorf("kafka publish failed: %w", err)
	}

	ei.logger.WithFields(logrus.Fields{
		"event_type": event.EventType,
		"source_id":  sourceEntity.ID,
		"target_id":  targetEntity.ID,
	}).Debug("Published entity merge event to Kafka")

	return nil
}

// EntityMergeEvent represents an entity merge event
type EntityMergeEvent struct {
	EventID        string         `json:"event_id"`
	EventType      string         `json:"event_type"` // entity.merged
	Timestamp      time.Time      `json:"timestamp"`
	ConversationID string         `json:"conversation_id"`
	SourceEntity   *memory.Entity `json:"source_entity"` // Entity being merged from
	TargetEntity   *memory.Entity `json:"target_entity"` // Entity being merged into
}
