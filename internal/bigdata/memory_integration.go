package bigdata

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"dev.helix.agent/internal/memory"
	"dev.helix.agent/pkg/messaging"
	"github.com/sirupsen/logrus"
)

// MemoryIntegration bridges the existing memory manager with distributed memory
type MemoryIntegration struct {
	memoryManager       *memory.Manager
	distributedMemory   *memory.DistributedMemoryManager
	kafkaBroker         messaging.MessageBroker
	logger              *logrus.Logger
	enableDistributed   bool
}

// NewMemoryIntegration creates a new memory integration
func NewMemoryIntegration(
	memoryManager *memory.Manager,
	distributedMemory *memory.DistributedMemoryManager,
	kafkaBroker messaging.MessageBroker,
	logger *logrus.Logger,
	enableDistributed bool,
) *MemoryIntegration {
	return &MemoryIntegration{
		memoryManager:     memoryManager,
		distributedMemory: distributedMemory,
		kafkaBroker:       kafkaBroker,
		logger:            logger,
		enableDistributed: enableDistributed,
	}
}

// AddMemory adds a memory to both local and distributed stores
func (mi *MemoryIntegration) AddMemory(ctx context.Context, mem *memory.Memory) error {
	// Add to local memory manager first
	if err := mi.memoryManager.AddMemory(ctx, mem); err != nil {
		return fmt.Errorf("local memory add failed: %w", err)
	}

	// Publish to distributed memory if enabled
	if mi.enableDistributed && mi.distributedMemory != nil {
		event := mi.createMemoryEvent("memory.created", mem)
		if err := mi.publishMemoryEvent(ctx, event); err != nil {
			mi.logger.WithError(err).Warn("Failed to publish memory event to distributed system")
			// Don't fail the operation - local memory is already saved
		}
	}

	return nil
}

// UpdateMemory updates a memory in both local and distributed stores
func (mi *MemoryIntegration) UpdateMemory(ctx context.Context, mem *memory.Memory) error {
	// Update local memory
	if err := mi.memoryManager.UpdateMemory(ctx, mem); err != nil {
		return fmt.Errorf("local memory update failed: %w", err)
	}

	// Publish update event if distributed is enabled
	if mi.enableDistributed && mi.distributedMemory != nil {
		event := mi.createMemoryEvent("memory.updated", mem)
		if err := mi.publishMemoryEvent(ctx, event); err != nil {
			mi.logger.WithError(err).Warn("Failed to publish memory update event")
		}
	}

	return nil
}

// DeleteMemory deletes a memory from both local and distributed stores
func (mi *MemoryIntegration) DeleteMemory(ctx context.Context, memoryID string) error {
	// Delete from local memory
	if err := mi.memoryManager.DeleteMemory(ctx, memoryID); err != nil {
		return fmt.Errorf("local memory delete failed: %w", err)
	}

	// Publish delete event if distributed is enabled
	if mi.enableDistributed && mi.distributedMemory != nil {
		event := &MemoryEvent{
			EventID:   generateEventID(),
			EventType: "memory.deleted",
			MemoryID:  memoryID,
			Timestamp: time.Now(),
		}
		if err := mi.publishMemoryEvent(ctx, event); err != nil {
			mi.logger.WithError(err).Warn("Failed to publish memory delete event")
		}
	}

	return nil
}

// SearchMemory searches memories (uses local store, distributed sync happens in background)
func (mi *MemoryIntegration) SearchMemory(ctx context.Context, query string, limit int) ([]*memory.Memory, error) {
	return mi.memoryManager.SearchMemory(ctx, query, limit)
}

// GetMemory retrieves a memory by ID (uses local store)
func (mi *MemoryIntegration) GetMemory(ctx context.Context, memoryID string) (*memory.Memory, error) {
	return mi.memoryManager.GetMemory(ctx, memoryID)
}

// MemoryEvent represents a memory change event for distributed synchronization
type MemoryEvent struct {
	EventID      string                 `json:"event_id"`
	EventType    string                 `json:"event_type"` // memory.created, memory.updated, memory.deleted
	NodeID       string                 `json:"node_id"`
	Timestamp    time.Time              `json:"timestamp"`
	MemoryID     string                 `json:"memory_id"`
	UserID       string                 `json:"user_id,omitempty"`
	SessionID    string                 `json:"session_id,omitempty"`
	Content      string                 `json:"content,omitempty"`
	Embedding    []float32              `json:"embedding,omitempty"`
	Entities     []memory.Entity        `json:"entities,omitempty"`
	Relationships []memory.Relationship `json:"relationships,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// createMemoryEvent creates a memory event from a memory object
func (mi *MemoryIntegration) createMemoryEvent(eventType string, mem *memory.Memory) *MemoryEvent {
	return &MemoryEvent{
		EventID:       generateEventID(),
		EventType:     eventType,
		NodeID:        getNodeID(),
		Timestamp:     time.Now(),
		MemoryID:      mem.ID,
		UserID:        mem.UserID,
		SessionID:     mem.SessionID,
		Content:       mem.Content,
		Embedding:     mem.Embedding,
		Entities:      mem.Entities,
		Relationships: mem.Relationships,
		Metadata:      mem.Metadata,
	}
}

// publishMemoryEvent publishes a memory event to Kafka
func (mi *MemoryIntegration) publishMemoryEvent(ctx context.Context, event *MemoryEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal memory event: %w", err)
	}

	msg := &messaging.Message{
		Topic:     "helixagent.memory.events",
		Key:       event.MemoryID,
		Payload:   payload,
		Timestamp: event.Timestamp,
		Headers: map[string]string{
			"event_type": event.EventType,
			"node_id":    event.NodeID,
			"memory_id":  event.MemoryID,
		},
	}

	if err := mi.kafkaBroker.Publish(ctx, msg.Topic, msg); err != nil {
		return fmt.Errorf("kafka publish failed: %w", err)
	}

	mi.logger.WithFields(logrus.Fields{
		"event_type": event.EventType,
		"memory_id":  event.MemoryID,
	}).Debug("Published memory event to Kafka")

	return nil
}

// StartEventConsumer starts consuming memory events from other nodes
func (mi *MemoryIntegration) StartEventConsumer(ctx context.Context) error {
	if !mi.enableDistributed || mi.distributedMemory == nil {
		mi.logger.Info("Distributed memory not enabled, skipping event consumer")
		return nil
	}

	go mi.consumeMemoryEvents(ctx)
	mi.logger.Info("Started memory event consumer")
	return nil
}

// consumeMemoryEvents consumes memory events from Kafka
func (mi *MemoryIntegration) consumeMemoryEvents(ctx context.Context) {
	// Create consumer
	consumer, err := mi.kafkaBroker.Subscribe(ctx, "helixagent.memory.events", "helixagent-memory-consumer")
	if err != nil {
		mi.logger.WithError(err).Error("Failed to subscribe to memory events")
		return
	}

	nodeID := getNodeID()

	for {
		select {
		case <-ctx.Done():
			mi.logger.Info("Stopping memory event consumer")
			return
		default:
			msg, err := consumer.Consume(ctx)
			if err != nil {
				mi.logger.WithError(err).Warn("Failed to consume memory event")
				time.Sleep(1 * time.Second)
				continue
			}

			// Parse event
			var event MemoryEvent
			if err := json.Unmarshal(msg.Payload, &event); err != nil {
				mi.logger.WithError(err).Error("Failed to unmarshal memory event")
				continue
			}

			// Skip events from this node
			if event.NodeID == nodeID {
				continue
			}

			// Apply remote event
			if err := mi.applyRemoteEvent(ctx, &event); err != nil {
				mi.logger.WithError(err).
					WithField("event_id", event.EventID).
					Error("Failed to apply remote memory event")
			}
		}
	}
}

// applyRemoteEvent applies a memory event from another node
func (mi *MemoryIntegration) applyRemoteEvent(ctx context.Context, event *MemoryEvent) error {
	switch event.EventType {
	case "memory.created":
		mem := &memory.Memory{
			ID:            event.MemoryID,
			UserID:        event.UserID,
			SessionID:     event.SessionID,
			Content:       event.Content,
			Embedding:     event.Embedding,
			Entities:      event.Entities,
			Relationships: event.Relationships,
			Metadata:      event.Metadata,
		}
		return mi.memoryManager.AddMemory(ctx, mem)

	case "memory.updated":
		mem := &memory.Memory{
			ID:            event.MemoryID,
			UserID:        event.UserID,
			SessionID:     event.SessionID,
			Content:       event.Content,
			Embedding:     event.Embedding,
			Entities:      event.Entities,
			Relationships: event.Relationships,
			Metadata:      event.Metadata,
		}
		return mi.memoryManager.UpdateMemory(ctx, mem)

	case "memory.deleted":
		return mi.memoryManager.DeleteMemory(ctx, event.MemoryID)

	default:
		mi.logger.WithField("event_type", event.EventType).Warn("Unknown memory event type")
		return nil
	}
}

// Helper functions
func generateEventID() string {
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}

func getNodeID() string {
	// In production, this would be a unique node identifier
	// For now, use a simple identifier based on hostname or environment variable
	nodeID := fmt.Sprintf("node_%d", time.Now().Unix()%1000)
	return nodeID
}
