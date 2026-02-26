package memory

import (
	"encoding/json"
	"time"
)

// MemoryEventType represents the type of memory event
type MemoryEventType string

const (
	// Memory CRUD events
	MemoryEventCreated MemoryEventType = "memory.created"
	MemoryEventUpdated MemoryEventType = "memory.updated"
	MemoryEventDeleted MemoryEventType = "memory.deleted"
	MemoryEventMerged  MemoryEventType = "memory.merged"

	// Entity events
	EntityEventCreated MemoryEventType = "entity.created"
	EntityEventUpdated MemoryEventType = "entity.updated"
	EntityEventLinked  MemoryEventType = "entity.linked"

	// Relationship events
	RelationshipEventCreated MemoryEventType = "relationship.created"
	RelationshipEventUpdated MemoryEventType = "relationship.updated"
	RelationshipEventDeleted MemoryEventType = "relationship.deleted"

	// Sync events
	MemoryEventSyncRequest MemoryEventType = "memory.sync_request"
)

// MemoryEvent represents a memory change event for distributed synchronization
type MemoryEvent struct {
	// Event metadata
	EventID   string          `json:"event_id"`
	EventType MemoryEventType `json:"event_type"`
	Timestamp time.Time       `json:"timestamp"`
	NodeID    string          `json:"node_id"` // Source node that generated the event

	// Memory data
	MemoryID  string    `json:"memory_id"`
	UserID    string    `json:"user_id"`
	SessionID string    `json:"session_id,omitempty"`
	Content   string    `json:"content,omitempty"`
	Embedding []float32 `json:"embedding,omitempty"`

	// Entity and relationship data
	Entities      []MemoryEntity       `json:"entities,omitempty"`
	Relationships []MemoryRelationship `json:"relationships,omitempty"`

	// Metadata
	Importance float64                `json:"importance,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Tags       []string               `json:"tags,omitempty"`

	// CRDT versioning
	Version     int64  `json:"version"`      // Lamport timestamp for ordering
	VectorClock string `json:"vector_clock"` // JSON-encoded vector clock

	// Merge tracking
	MergedFrom []string `json:"merged_from,omitempty"` // IDs of memories merged into this one
}

// MemoryEntity represents an entity in the event (distinct from global Entity)
type MemoryEntity struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Name       string                 `json:"name"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Confidence float64                `json:"confidence"`
	CreatedAt  time.Time              `json:"created_at"`
}

// MemoryRelationship represents a relationship in the event
type MemoryRelationship struct {
	ID         string                 `json:"id"`
	FromEntity string                 `json:"from_entity"`
	ToEntity   string                 `json:"to_entity"`
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Strength   float64                `json:"strength"`
	CreatedAt  time.Time              `json:"created_at"`
}

// EventLog stores memory events for replay and synchronization
type EventLog interface {
	// Append appends an event to the log
	Append(event *MemoryEvent) error

	// GetEvents retrieves events for a memory ID
	GetEvents(memoryID string) ([]*MemoryEvent, error)

	// GetEventsSince retrieves events after a specific timestamp
	GetEventsSince(timestamp time.Time) ([]*MemoryEvent, error)

	// GetEventsForUser retrieves events for a specific user
	GetEventsForUser(userID string) ([]*MemoryEvent, error)

	// GetEventsFromNode retrieves events from a specific node
	GetEventsFromNode(nodeID string) ([]*MemoryEvent, error)
}

// VectorClock represents a vector clock for distributed ordering
type VectorClock map[string]int64

// NewVectorClock creates a new vector clock
func NewVectorClock() VectorClock {
	return make(VectorClock)
}

// Increment increments the counter for a node
func (vc VectorClock) Increment(nodeID string) {
	vc[nodeID]++
}

// Update updates the vector clock with another clock (merge)
func (vc VectorClock) Update(other VectorClock) {
	for nodeID, count := range other {
		if count > vc[nodeID] {
			vc[nodeID] = count
		}
	}
}

// HappensBefore returns true if this clock happens before the other
func (vc VectorClock) HappensBefore(other VectorClock) bool {
	atLeastOneLess := false
	for nodeID, count := range vc {
		otherCount := other[nodeID]
		if count > otherCount {
			return false
		}
		if count < otherCount {
			atLeastOneLess = true
		}
	}
	return atLeastOneLess
}

// Concurrent returns true if the clocks are concurrent (neither happens before the other)
func (vc VectorClock) Concurrent(other VectorClock) bool {
	return !vc.HappensBefore(other) && !other.HappensBefore(vc)
}

// String converts vector clock to JSON string
func (vc VectorClock) String() string {
	data, _ := json.Marshal(vc) //nolint:errcheck
	return string(data)
}

// ParseVectorClock parses a vector clock from JSON string
func ParseVectorClock(s string) (VectorClock, error) {
	var vc VectorClock
	if err := json.Unmarshal([]byte(s), &vc); err != nil {
		return nil, err
	}
	return vc, nil
}

// MemorySnapshot represents a snapshot of memory state at a point in time
type MemorySnapshot struct {
	SnapshotID string    `json:"snapshot_id"`
	Timestamp  time.Time `json:"timestamp"`
	NodeID     string    `json:"node_id"`
	UserID     string    `json:"user_id"`

	// Snapshot data
	Memories      []*Memory              `json:"memories"`
	Entities      []Entity               `json:"entities"`
	Relationships []Relationship         `json:"relationships"`
	VectorClock   VectorClock            `json:"vector_clock"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// EventStream represents a stream of memory events
type EventStream struct {
	StreamID    string         `json:"stream_id"`
	UserID      string         `json:"user_id"`
	SessionID   string         `json:"session_id,omitempty"`
	StartTime   time.Time      `json:"start_time"`
	EndTime     time.Time      `json:"end_time,omitempty"`
	Events      []*MemoryEvent `json:"events"`
	EventCount  int            `json:"event_count"`
	VectorClock VectorClock    `json:"vector_clock"`
	Checkpoints []time.Time    `json:"checkpoints,omitempty"`
}

// NewMemoryEvent creates a new memory event
func NewMemoryEvent(eventType MemoryEventType, nodeID, memoryID, userID string) *MemoryEvent {
	return &MemoryEvent{
		EventID:     generateEventID(),
		EventType:   eventType,
		Timestamp:   time.Now().UTC(),
		NodeID:      nodeID,
		MemoryID:    memoryID,
		UserID:      userID,
		Version:     time.Now().UnixNano(), // Lamport timestamp
		VectorClock: NewVectorClock().String(),
	}
}

// Clone creates a deep copy of the memory event
func (e *MemoryEvent) Clone() *MemoryEvent {
	clone := &MemoryEvent{
		EventID:     e.EventID,
		EventType:   e.EventType,
		Timestamp:   e.Timestamp,
		NodeID:      e.NodeID,
		MemoryID:    e.MemoryID,
		UserID:      e.UserID,
		SessionID:   e.SessionID,
		Content:     e.Content,
		Importance:  e.Importance,
		Version:     e.Version,
		VectorClock: e.VectorClock,
	}

	// Deep copy embedding
	if len(e.Embedding) > 0 {
		clone.Embedding = make([]float32, len(e.Embedding))
		copy(clone.Embedding, e.Embedding)
	}

	// Deep copy entities
	if len(e.Entities) > 0 {
		clone.Entities = make([]MemoryEntity, len(e.Entities))
		copy(clone.Entities, e.Entities)
	}

	// Deep copy relationships
	if len(e.Relationships) > 0 {
		clone.Relationships = make([]MemoryRelationship, len(e.Relationships))
		copy(clone.Relationships, e.Relationships)
	}

	// Deep copy metadata
	if e.Metadata != nil {
		clone.Metadata = make(map[string]interface{})
		for k, v := range e.Metadata {
			clone.Metadata[k] = v
		}
	}

	// Deep copy tags
	if len(e.Tags) > 0 {
		clone.Tags = make([]string, len(e.Tags))
		copy(clone.Tags, e.Tags)
	}

	// Deep copy merged from
	if len(e.MergedFrom) > 0 {
		clone.MergedFrom = make([]string, len(e.MergedFrom))
		copy(clone.MergedFrom, e.MergedFrom)
	}

	return clone
}

// ToJSON converts the event to JSON
func (e *MemoryEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON parses an event from JSON
func FromJSON(data []byte) (*MemoryEvent, error) {
	var event MemoryEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, err
	}
	return &event, nil
}

// generateEventID generates a unique event ID
func generateEventID() string {
	// Format: evt-<timestamp>-<random>
	return "evt-" + time.Now().UTC().Format("20060102150405.000000") + "-" + randomString(8)
}

// randomString generates a random string of specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// EventStreamStats represents statistics for an event stream
type EventStreamStats struct {
	StreamID             string         `json:"stream_id"`
	TotalEvents          int            `json:"total_events"`
	EventTypes           map[string]int `json:"event_types"`
	Duration             time.Duration  `json:"duration"`
	EventsPerSecond      float64        `json:"events_per_second"`
	NodesInvolved        []string       `json:"nodes_involved"`
	MemoriesAffected     int            `json:"memories_affected"`
	EntitiesCreated      int            `json:"entities_created"`
	RelationshipsCreated int            `json:"relationships_created"`
}

// CalculateStats calculates statistics for an event stream
func (es *EventStream) CalculateStats() *EventStreamStats {
	stats := &EventStreamStats{
		StreamID:   es.StreamID,
		EventTypes: make(map[string]int),
	}

	uniqueNodes := make(map[string]bool)
	uniqueMemories := make(map[string]bool)

	for _, event := range es.Events {
		stats.TotalEvents++
		stats.EventTypes[string(event.EventType)]++
		uniqueNodes[event.NodeID] = true
		uniqueMemories[event.MemoryID] = true

		stats.EntitiesCreated += len(event.Entities)
		stats.RelationshipsCreated += len(event.Relationships)
	}

	stats.NodesInvolved = make([]string, 0, len(uniqueNodes))
	for nodeID := range uniqueNodes {
		stats.NodesInvolved = append(stats.NodesInvolved, nodeID)
	}

	stats.MemoriesAffected = len(uniqueMemories)

	if !es.EndTime.IsZero() {
		stats.Duration = es.EndTime.Sub(es.StartTime)
		if stats.Duration.Seconds() > 0 {
			stats.EventsPerSecond = float64(stats.TotalEvents) / stats.Duration.Seconds()
		}
	}

	return stats
}
