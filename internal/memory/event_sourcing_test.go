package memory

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- VectorClock ---

func TestNewVectorClock(t *testing.T) {
	vc := NewVectorClock()
	require.NotNil(t, vc)
	assert.Empty(t, vc)
}

func TestVectorClock_Increment(t *testing.T) {
	vc := NewVectorClock()

	vc.Increment("node1")
	assert.Equal(t, int64(1), vc["node1"])

	vc.Increment("node1")
	assert.Equal(t, int64(2), vc["node1"])

	vc.Increment("node2")
	assert.Equal(t, int64(1), vc["node2"])
	assert.Equal(t, int64(2), vc["node1"])
}

func TestVectorClock_Update(t *testing.T) {
	t.Run("MergeHigherValues", func(t *testing.T) {
		vc1 := VectorClock{"node1": 2, "node2": 1}
		vc2 := VectorClock{"node1": 1, "node2": 3, "node3": 1}

		vc1.Update(vc2)
		assert.Equal(t, int64(2), vc1["node1"])
		assert.Equal(t, int64(3), vc1["node2"])
		assert.Equal(t, int64(1), vc1["node3"])
	})

	t.Run("UpdateFromEmpty", func(t *testing.T) {
		vc1 := VectorClock{"node1": 1}
		vc2 := NewVectorClock()

		vc1.Update(vc2)
		assert.Equal(t, int64(1), vc1["node1"])
	})

	t.Run("UpdateEmpty", func(t *testing.T) {
		vc1 := NewVectorClock()
		vc2 := VectorClock{"node1": 5}

		vc1.Update(vc2)
		assert.Equal(t, int64(5), vc1["node1"])
	})
}

func TestVectorClock_HappensBefore(t *testing.T) {
	tests := []struct {
		name     string
		vc       VectorClock
		other    VectorClock
		expected bool
	}{
		{
			name:     "StrictlyBefore",
			vc:       VectorClock{"node1": 1, "node2": 1},
			other:    VectorClock{"node1": 2, "node2": 2},
			expected: true,
		},
		{
			name:     "PartiallyBefore",
			vc:       VectorClock{"node1": 1, "node2": 1},
			other:    VectorClock{"node1": 2, "node2": 1},
			expected: true,
		},
		{
			name:     "Equal",
			vc:       VectorClock{"node1": 1},
			other:    VectorClock{"node1": 1},
			expected: false,
		},
		{
			name:     "After",
			vc:       VectorClock{"node1": 2},
			other:    VectorClock{"node1": 1},
			expected: false,
		},
		{
			name:     "Concurrent",
			vc:       VectorClock{"node1": 2, "node2": 1},
			other:    VectorClock{"node1": 1, "node2": 2},
			expected: false,
		},
		{
			name:     "BothEmpty",
			vc:       NewVectorClock(),
			other:    NewVectorClock(),
			expected: false,
		},
		{
			name:     "EmptyBeforeNonEmpty",
			vc:       NewVectorClock(),
			other:    VectorClock{"node1": 1},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.vc.HappensBefore(tt.other))
		})
	}
}

func TestVectorClock_Concurrent(t *testing.T) {
	tests := []struct {
		name     string
		vc       VectorClock
		other    VectorClock
		expected bool
	}{
		{
			name:     "TrulyConcurrent",
			vc:       VectorClock{"node1": 2, "node2": 1},
			other:    VectorClock{"node1": 1, "node2": 2},
			expected: true,
		},
		{
			name:     "NotConcurrent_OneBefore",
			vc:       VectorClock{"node1": 1},
			other:    VectorClock{"node1": 2},
			expected: false,
		},
		{
			name:     "Equal_Concurrent",
			vc:       VectorClock{"node1": 1},
			other:    VectorClock{"node1": 1},
			expected: true,
		},
		{
			name:     "DisjointNodes",
			vc:       VectorClock{"node1": 1},
			other:    VectorClock{"node2": 1},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.vc.Concurrent(tt.other))
		})
	}
}

func TestVectorClock_String(t *testing.T) {
	vc := VectorClock{"node1": 1, "node2": 2}
	s := vc.String()

	// Parse it back to verify
	var parsed map[string]int64
	err := json.Unmarshal([]byte(s), &parsed)
	require.NoError(t, err)
	assert.Equal(t, int64(1), parsed["node1"])
	assert.Equal(t, int64(2), parsed["node2"])
}

func TestVectorClock_String_Empty(t *testing.T) {
	vc := NewVectorClock()
	assert.Equal(t, "{}", vc.String())
}

func TestParseVectorClock(t *testing.T) {
	t.Run("ValidJSON", func(t *testing.T) {
		vc, err := ParseVectorClock(`{"node1":3,"node2":5}`)
		require.NoError(t, err)
		assert.Equal(t, int64(3), vc["node1"])
		assert.Equal(t, int64(5), vc["node2"])
	})

	t.Run("EmptyObject", func(t *testing.T) {
		vc, err := ParseVectorClock(`{}`)
		require.NoError(t, err)
		assert.Empty(t, vc)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		_, err := ParseVectorClock("not-json")
		require.Error(t, err)
	})

	t.Run("Roundtrip", func(t *testing.T) {
		original := VectorClock{"a": 10, "b": 20}
		s := original.String()
		parsed, err := ParseVectorClock(s)
		require.NoError(t, err)
		assert.Equal(t, original, parsed)
	})
}

// --- MemoryEvent ---

func TestNewMemoryEvent(t *testing.T) {
	event := NewMemoryEvent(MemoryEventCreated, "node1", "mem1", "user1")

	assert.NotEmpty(t, event.EventID)
	assert.Equal(t, MemoryEventCreated, event.EventType)
	assert.Equal(t, "node1", event.NodeID)
	assert.Equal(t, "mem1", event.MemoryID)
	assert.Equal(t, "user1", event.UserID)
	assert.False(t, event.Timestamp.IsZero())
	assert.NotZero(t, event.Version)
	assert.NotEmpty(t, event.VectorClock)
}

func TestMemoryEvent_Clone(t *testing.T) {
	t.Run("FullClone", func(t *testing.T) {
		original := &MemoryEvent{
			EventID:   "evt1",
			EventType: MemoryEventUpdated,
			Timestamp: time.Now(),
			NodeID:    "node1",
			MemoryID:  "mem1",
			UserID:    "user1",
			SessionID: "session1",
			Content:   "content",
			Embedding: []float32{0.1, 0.2},
			Entities:  []MemoryEntity{{ID: "e1", Name: "Entity1"}},
			Relationships: []MemoryRelationship{
				{ID: "r1", FromEntity: "e1", ToEntity: "e2"},
			},
			Metadata:    map[string]interface{}{"key": "value"},
			Tags:        []string{"tag1", "tag2"},
			Importance:  0.8,
			Version:     100,
			VectorClock: `{"node1":1}`,
			MergedFrom:  []string{"id1", "id2"},
		}

		clone := original.Clone()

		assert.Equal(t, original.EventID, clone.EventID)
		assert.Equal(t, original.EventType, clone.EventType)
		assert.Equal(t, original.Content, clone.Content)
		assert.Equal(t, original.Embedding, clone.Embedding)
		assert.Equal(t, original.Tags, clone.Tags)
		assert.Equal(t, original.MergedFrom, clone.MergedFrom)
		assert.Equal(t, original.Metadata["key"], clone.Metadata["key"])

		// Verify deep copy (modifications to clone don't affect original)
		clone.Embedding[0] = 9.9
		assert.NotEqual(t, original.Embedding[0], clone.Embedding[0])

		clone.Tags[0] = "modified"
		assert.NotEqual(t, original.Tags[0], clone.Tags[0])

		clone.Metadata["new_key"] = "new_value"
		_, exists := original.Metadata["new_key"]
		assert.False(t, exists)
	})

	t.Run("MinimalClone", func(t *testing.T) {
		original := &MemoryEvent{
			EventID:   "evt1",
			EventType: MemoryEventDeleted,
		}

		clone := original.Clone()
		assert.Equal(t, "evt1", clone.EventID)
		assert.Nil(t, clone.Embedding)
		assert.Nil(t, clone.Tags)
		assert.Nil(t, clone.Metadata)
	})
}

func TestMemoryEvent_ToJSON(t *testing.T) {
	event := &MemoryEvent{
		EventID:   "evt1",
		EventType: MemoryEventCreated,
		NodeID:    "node1",
		MemoryID:  "mem1",
		UserID:    "user1",
		Content:   "test content",
		Timestamp: time.Now().UTC(),
	}

	data, err := event.ToJSON()
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Verify it's valid JSON
	var parsed map[string]interface{}
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)
	assert.Equal(t, "evt1", parsed["event_id"])
	assert.Equal(t, "node1", parsed["node_id"])
}

func TestFromJSON(t *testing.T) {
	t.Run("ValidJSON", func(t *testing.T) {
		original := &MemoryEvent{
			EventID:    "evt1",
			EventType:  MemoryEventUpdated,
			NodeID:     "node1",
			MemoryID:   "mem1",
			UserID:     "user1",
			Content:    "content",
			Importance: 0.8,
			Tags:       []string{"a", "b"},
		}

		data, err := original.ToJSON()
		require.NoError(t, err)

		parsed, err := FromJSON(data)
		require.NoError(t, err)
		assert.Equal(t, original.EventID, parsed.EventID)
		assert.Equal(t, original.EventType, parsed.EventType)
		assert.Equal(t, original.Content, parsed.Content)
		assert.Equal(t, original.Importance, parsed.Importance)
		assert.Equal(t, original.Tags, parsed.Tags)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		_, err := FromJSON([]byte("not-json"))
		require.Error(t, err)
	})

	t.Run("EmptyObject", func(t *testing.T) {
		parsed, err := FromJSON([]byte("{}"))
		require.NoError(t, err)
		assert.Empty(t, parsed.EventID)
	})
}

// --- MemoryEventType constants ---

func TestMemoryEventType_Constants(t *testing.T) {
	assert.Equal(t, MemoryEventType("memory.created"), MemoryEventCreated)
	assert.Equal(t, MemoryEventType("memory.updated"), MemoryEventUpdated)
	assert.Equal(t, MemoryEventType("memory.deleted"), MemoryEventDeleted)
	assert.Equal(t, MemoryEventType("memory.merged"), MemoryEventMerged)
	assert.Equal(t, MemoryEventType("entity.created"), EntityEventCreated)
	assert.Equal(t, MemoryEventType("entity.updated"), EntityEventUpdated)
	assert.Equal(t, MemoryEventType("entity.linked"), EntityEventLinked)
	assert.Equal(t, MemoryEventType("relationship.created"), RelationshipEventCreated)
	assert.Equal(t, MemoryEventType("relationship.updated"), RelationshipEventUpdated)
	assert.Equal(t, MemoryEventType("relationship.deleted"), RelationshipEventDeleted)
	assert.Equal(t, MemoryEventType("memory.sync_request"), MemoryEventSyncRequest)
}

// --- MemorySnapshot ---

func TestMemorySnapshot_Struct(t *testing.T) {
	now := time.Now()
	snapshot := &MemorySnapshot{
		SnapshotID: "snap1",
		Timestamp:  now,
		NodeID:     "node1",
		UserID:     "user1",
		Memories: []*Memory{
			{ID: "mem1", Content: "content1"},
		},
		Entities: []Entity{
			{ID: "e1", Name: "Entity1"},
		},
		VectorClock: VectorClock{"node1": 5},
		Metadata:    map[string]interface{}{"key": "value"},
	}

	assert.Equal(t, "snap1", snapshot.SnapshotID)
	assert.Equal(t, "node1", snapshot.NodeID)
	assert.Len(t, snapshot.Memories, 1)
	assert.Len(t, snapshot.Entities, 1)
	assert.Equal(t, int64(5), snapshot.VectorClock["node1"])
}

// --- EventStream ---

func TestEventStream_CalculateStats(t *testing.T) {
	t.Run("WithEvents", func(t *testing.T) {
		now := time.Now()
		stream := &EventStream{
			StreamID:  "stream1",
			UserID:    "user1",
			StartTime: now.Add(-10 * time.Second),
			EndTime:   now,
			Events: []*MemoryEvent{
				{
					EventType: MemoryEventCreated,
					NodeID:    "node1",
					MemoryID:  "mem1",
					Entities:  []MemoryEntity{{ID: "e1"}},
					Relationships: []MemoryRelationship{
						{ID: "r1"},
					},
				},
				{
					EventType: MemoryEventUpdated,
					NodeID:    "node1",
					MemoryID:  "mem1",
				},
				{
					EventType: MemoryEventCreated,
					NodeID:    "node2",
					MemoryID:  "mem2",
					Entities:  []MemoryEntity{{ID: "e2"}, {ID: "e3"}},
				},
			},
		}

		stats := stream.CalculateStats()
		assert.Equal(t, "stream1", stats.StreamID)
		assert.Equal(t, 3, stats.TotalEvents)
		assert.Equal(t, 2, stats.EventTypes["memory.created"])
		assert.Equal(t, 1, stats.EventTypes["memory.updated"])
		assert.Equal(t, 2, stats.MemoriesAffected)
		assert.Len(t, stats.NodesInvolved, 2)
		assert.Equal(t, 3, stats.EntitiesCreated)
		assert.Equal(t, 1, stats.RelationshipsCreated)
		assert.Equal(t, 10*time.Second, stats.Duration)
		assert.InDelta(t, 0.3, stats.EventsPerSecond, 0.01)
	})

	t.Run("EmptyStream", func(t *testing.T) {
		stream := &EventStream{
			StreamID: "stream2",
			Events:   []*MemoryEvent{},
		}

		stats := stream.CalculateStats()
		assert.Equal(t, 0, stats.TotalEvents)
		assert.Empty(t, stats.NodesInvolved)
		assert.Equal(t, 0, stats.MemoriesAffected)
	})

	t.Run("NoEndTime", func(t *testing.T) {
		stream := &EventStream{
			StreamID:  "stream3",
			StartTime: time.Now(),
			Events: []*MemoryEvent{
				{EventType: MemoryEventCreated, NodeID: "n1", MemoryID: "m1"},
			},
		}

		stats := stream.CalculateStats()
		assert.Equal(t, 1, stats.TotalEvents)
		assert.Equal(t, time.Duration(0), stats.Duration)
		assert.Equal(t, float64(0), stats.EventsPerSecond)
	})
}

// --- MemoryEntity and MemoryRelationship structs ---

func TestMemoryEntity_Struct(t *testing.T) {
	entity := MemoryEntity{
		ID:         "e1",
		Type:       "person",
		Name:       "Alice",
		Properties: map[string]interface{}{"age": 30},
		Confidence: 0.95,
		CreatedAt:  time.Now(),
	}

	assert.Equal(t, "e1", entity.ID)
	assert.Equal(t, "person", entity.Type)
	assert.Equal(t, "Alice", entity.Name)
	assert.Equal(t, 0.95, entity.Confidence)
}

func TestMemoryRelationship_Struct(t *testing.T) {
	rel := MemoryRelationship{
		ID:         "r1",
		FromEntity: "e1",
		ToEntity:   "e2",
		Type:       "knows",
		Properties: map[string]interface{}{"since": "2024"},
		Strength:   0.8,
		CreatedAt:  time.Now(),
	}

	assert.Equal(t, "r1", rel.ID)
	assert.Equal(t, "e1", rel.FromEntity)
	assert.Equal(t, "e2", rel.ToEntity)
	assert.Equal(t, 0.8, rel.Strength)
}

// --- EventStreamStats ---

func TestEventStreamStats_Struct(t *testing.T) {
	stats := &EventStreamStats{
		StreamID:             "stream1",
		TotalEvents:          100,
		EventTypes:           map[string]int{"memory.created": 50, "memory.updated": 50},
		Duration:             10 * time.Second,
		EventsPerSecond:      10.0,
		NodesInvolved:        []string{"node1", "node2"},
		MemoriesAffected:     25,
		EntitiesCreated:      10,
		RelationshipsCreated: 5,
	}

	assert.Equal(t, "stream1", stats.StreamID)
	assert.Equal(t, 100, stats.TotalEvents)
	assert.Equal(t, 10.0, stats.EventsPerSecond)
}

// --- JSON roundtrip for MemoryEvent ---

func TestMemoryEvent_JSONRoundtrip(t *testing.T) {
	original := &MemoryEvent{
		EventID:     "evt-test",
		EventType:   MemoryEventMerged,
		Timestamp:   time.Now().UTC().Truncate(time.Millisecond),
		NodeID:      "node1",
		MemoryID:    "mem1",
		UserID:      "user1",
		SessionID:   "session1",
		Content:     "merged content",
		Embedding:   []float32{0.1, 0.2, 0.3},
		Importance:  0.75,
		Tags:        []string{"tag1"},
		MergedFrom:  []string{"mem2", "mem3"},
		Version:     42,
		VectorClock: `{"node1":5}`,
		Entities: []MemoryEntity{
			{ID: "e1", Name: "E1", Confidence: 0.9},
		},
		Relationships: []MemoryRelationship{
			{ID: "r1", FromEntity: "e1", ToEntity: "e2", Strength: 0.5},
		},
		Metadata: map[string]interface{}{"key": "val"},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored MemoryEvent
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	assert.Equal(t, original.EventID, restored.EventID)
	assert.Equal(t, original.EventType, restored.EventType)
	assert.Equal(t, original.Content, restored.Content)
	assert.Equal(t, original.Importance, restored.Importance)
	assert.Equal(t, original.Tags, restored.Tags)
	assert.Equal(t, original.MergedFrom, restored.MergedFrom)
	assert.Equal(t, original.Version, restored.Version)
	assert.Len(t, restored.Entities, 1)
	assert.Len(t, restored.Relationships, 1)
}
