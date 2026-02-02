package memory

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- CRDTResolver: NewCRDTResolver ---

func TestNewCRDTResolver(t *testing.T) {
	tests := []struct {
		name     string
		strategy ConflictStrategy
	}{
		{"LastWriteWins", ConflictStrategyLastWriteWins},
		{"MergeAll", ConflictStrategyMergeAll},
		{"Importance", ConflictStrategyImportance},
		{"VectorClock", ConflictStrategyVectorClock},
		{"Custom", ConflictStrategyCustom},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := NewCRDTResolver(tt.strategy)
			require.NotNil(t, cr)
			assert.Equal(t, tt.strategy, cr.strategy)
			assert.Nil(t, cr.customResolver)
		})
	}
}

func TestCRDTResolver_WithCustomResolver(t *testing.T) {
	cr := NewCRDTResolver(ConflictStrategyCustom)
	fn := func(m *Memory, e *MemoryEvent) *Memory {
		return m
	}
	result := cr.WithCustomResolver(fn)
	assert.Same(t, cr, result)
	assert.NotNil(t, cr.customResolver)
}

// --- CRDTResolver: Merge with LastWriteWins ---

func TestCRDTResolver_Merge_LastWriteWins(t *testing.T) {
	cr := NewCRDTResolver(ConflictStrategyLastWriteWins)
	now := time.Now()

	t.Run("RemoteNewer", func(t *testing.T) {
		local := &Memory{
			ID:        "mem1",
			UserID:    "user1",
			Content:   "local content",
			UpdatedAt: now.Add(-time.Hour),
		}
		remote := &MemoryEvent{
			UserID:      "user1",
			Content:     "remote content",
			Timestamp:   now,
			VectorClock: "{}",
		}

		merged := cr.Merge(local, remote)
		assert.Equal(t, "mem1", merged.ID)
		assert.Equal(t, "remote content", merged.Content)
	})

	t.Run("LocalNewer", func(t *testing.T) {
		local := &Memory{
			ID:        "mem1",
			UserID:    "user1",
			Content:   "local content",
			UpdatedAt: now,
		}
		remote := &MemoryEvent{
			UserID:      "user1",
			Content:     "remote content",
			Timestamp:   now.Add(-time.Hour),
			VectorClock: "{}",
		}

		merged := cr.Merge(local, remote)
		assert.Equal(t, "local content", merged.Content)
	})
}

// --- CRDTResolver: Merge with MergeAll ---

func TestCRDTResolver_Merge_MergeAll(t *testing.T) {
	cr := NewCRDTResolver(ConflictStrategyMergeAll)
	now := time.Now()

	t.Run("RemoteLongerContent", func(t *testing.T) {
		local := &Memory{
			ID:         "mem1",
			UserID:     "user1",
			Content:    "short",
			Importance: 0.5,
			Embedding:  []float32{0.1},
			CreatedAt:  now.Add(-time.Hour),
			UpdatedAt:  now.Add(-time.Hour),
			Metadata:   map[string]interface{}{"key1": "val1"},
		}
		remote := &MemoryEvent{
			Content:     "longer remote content here",
			Importance:  0.8,
			Embedding:   []float32{0.2, 0.3},
			Timestamp:   now,
			VectorClock: "{}",
			Metadata:    map[string]interface{}{"key2": "val2"},
		}

		merged := cr.Merge(local, remote)
		assert.Equal(t, "mem1", merged.ID)
		assert.Equal(t, "longer remote content here", merged.Content)
		assert.Equal(t, 0.8, merged.Importance)
		assert.Equal(t, []float32{0.2, 0.3}, merged.Embedding)
		assert.Equal(t, now.Add(-time.Hour), merged.CreatedAt)
		assert.Equal(t, now, merged.UpdatedAt)
		assert.Equal(t, "val1", merged.Metadata["key1"])
		assert.Equal(t, "val2", merged.Metadata["key2"])
	})

	t.Run("LocalLongerContent", func(t *testing.T) {
		local := &Memory{
			ID:        "mem1",
			Content:   "longer local content here",
			UpdatedAt: now,
		}
		remote := &MemoryEvent{
			Content:     "short",
			Timestamp:   now.Add(-time.Hour),
			VectorClock: "{}",
		}

		merged := cr.Merge(local, remote)
		assert.Equal(t, "longer local content here", merged.Content)
	})

	t.Run("EqualLengthContent_RemoteNewer", func(t *testing.T) {
		local := &Memory{
			ID:        "mem1",
			Content:   "abcde",
			UpdatedAt: now.Add(-time.Hour),
		}
		remote := &MemoryEvent{
			Content:     "fghij",
			Timestamp:   now,
			VectorClock: "{}",
		}

		merged := cr.Merge(local, remote)
		assert.Equal(t, "fghij", merged.Content)
	})

	t.Run("EqualLengthContent_LocalNewer", func(t *testing.T) {
		local := &Memory{
			ID:        "mem1",
			Content:   "abcde",
			UpdatedAt: now,
		}
		remote := &MemoryEvent{
			Content:     "fghij",
			Timestamp:   now.Add(-time.Hour),
			VectorClock: "{}",
		}

		merged := cr.Merge(local, remote)
		assert.Equal(t, "abcde", merged.Content)
	})

	t.Run("NoRemoteEmbedding_KeepsLocal", func(t *testing.T) {
		local := &Memory{
			ID:        "mem1",
			Content:   "content",
			Embedding: []float32{0.1},
			UpdatedAt: now,
		}
		remote := &MemoryEvent{
			Content:     "c",
			Embedding:   nil,
			Timestamp:   now.Add(-time.Hour),
			VectorClock: "{}",
		}

		merged := cr.Merge(local, remote)
		assert.Equal(t, []float32{0.1}, merged.Embedding)
	})

	t.Run("NilMetadata", func(t *testing.T) {
		local := &Memory{
			ID:        "mem1",
			Content:   "content",
			UpdatedAt: now,
			Metadata:  nil,
		}
		remote := &MemoryEvent{
			Content:     "c",
			Timestamp:   now.Add(-time.Hour),
			VectorClock: "{}",
			Metadata:    nil,
		}

		merged := cr.Merge(local, remote)
		assert.Nil(t, merged.Metadata)
	})

	t.Run("LocalHigherImportance", func(t *testing.T) {
		local := &Memory{
			ID:         "mem1",
			Content:    "content",
			Importance: 0.9,
			UpdatedAt:  now,
		}
		remote := &MemoryEvent{
			Content:     "c",
			Importance:  0.3,
			Timestamp:   now.Add(-time.Hour),
			VectorClock: "{}",
		}

		merged := cr.Merge(local, remote)
		assert.Equal(t, 0.9, merged.Importance)
	})
}

// --- CRDTResolver: Merge with Importance ---

func TestCRDTResolver_Merge_Importance(t *testing.T) {
	cr := NewCRDTResolver(ConflictStrategyImportance)

	t.Run("RemoteHigherImportance", func(t *testing.T) {
		local := &Memory{
			ID:         "mem1",
			Content:    "local",
			Importance: 0.3,
		}
		remote := &MemoryEvent{
			Content:     "remote",
			Importance:  0.9,
			Timestamp:   time.Now(),
			VectorClock: "{}",
		}

		merged := cr.Merge(local, remote)
		assert.Equal(t, "remote", merged.Content)
	})

	t.Run("LocalHigherImportance", func(t *testing.T) {
		local := &Memory{
			ID:         "mem1",
			Content:    "local",
			Importance: 0.9,
		}
		remote := &MemoryEvent{
			Content:     "remote",
			Importance:  0.3,
			Timestamp:   time.Now(),
			VectorClock: "{}",
		}

		merged := cr.Merge(local, remote)
		assert.Equal(t, "local", merged.Content)
	})

	t.Run("EqualImportance", func(t *testing.T) {
		local := &Memory{
			ID:         "mem1",
			Content:    "local",
			Importance: 0.5,
		}
		remote := &MemoryEvent{
			Content:     "remote",
			Importance:  0.5,
			Timestamp:   time.Now(),
			VectorClock: "{}",
		}

		merged := cr.Merge(local, remote)
		assert.Equal(t, "local", merged.Content)
	})
}

// --- CRDTResolver: Merge with VectorClock ---

func TestCRDTResolver_Merge_VectorClock(t *testing.T) {
	cr := NewCRDTResolver(ConflictStrategyVectorClock)
	now := time.Now()

	t.Run("RemoteHappensAfterLocal", func(t *testing.T) {
		local := &Memory{
			ID:        "mem1",
			Content:   "local",
			UpdatedAt: now.Add(-time.Hour),
			Metadata:  map[string]interface{}{"vector_clock": `{"node1":1}`},
		}
		remote := &MemoryEvent{
			Content:     "remote",
			Timestamp:   now,
			VectorClock: `{"node1":2}`,
		}

		merged := cr.Merge(local, remote)
		assert.Equal(t, "remote", merged.Content)
	})

	t.Run("LocalHappensAfterRemote", func(t *testing.T) {
		local := &Memory{
			ID:        "mem1",
			Content:   "local",
			UpdatedAt: now,
			Metadata:  map[string]interface{}{"vector_clock": `{"node1":2}`},
		}
		remote := &MemoryEvent{
			Content:     "remote",
			Timestamp:   now.Add(-time.Hour),
			VectorClock: `{"node1":1}`,
		}

		merged := cr.Merge(local, remote)
		assert.Equal(t, "local", merged.Content)
	})

	t.Run("ConcurrentUpdates_MergesAll", func(t *testing.T) {
		local := &Memory{
			ID:         "mem1",
			UserID:     "user1",
			Content:    "short",
			Importance: 0.5,
			UpdatedAt:  now,
			CreatedAt:  now.Add(-time.Hour),
			Metadata:   map[string]interface{}{"vector_clock": `{"node1":1}`},
		}
		remote := &MemoryEvent{
			Content:     "longer concurrent content",
			Importance:  0.8,
			Timestamp:   now,
			VectorClock: `{"node2":1}`,
		}

		merged := cr.Merge(local, remote)
		// Concurrent => mergeAll; remote is longer
		assert.Equal(t, "longer concurrent content", merged.Content)
	})

	t.Run("InvalidRemoteVectorClock_FallsBackToLWW", func(t *testing.T) {
		local := &Memory{
			ID:        "mem1",
			Content:   "local",
			UpdatedAt: now.Add(-time.Hour),
			Metadata:  map[string]interface{}{"vector_clock": `{"node1":1}`},
		}
		remote := &MemoryEvent{
			Content:     "remote",
			Timestamp:   now,
			VectorClock: "not-valid-json",
		}

		merged := cr.Merge(local, remote)
		// Falls back to LWW; remote is newer
		assert.Equal(t, "remote", merged.Content)
	})

	t.Run("NoLocalVectorClock_FallsBackToLWW", func(t *testing.T) {
		local := &Memory{
			ID:        "mem1",
			Content:   "local",
			UpdatedAt: now.Add(-time.Hour),
			Metadata:  map[string]interface{}{},
		}
		remote := &MemoryEvent{
			Content:     "remote",
			Timestamp:   now,
			VectorClock: `{"node1":1}`,
		}

		merged := cr.Merge(local, remote)
		assert.Equal(t, "remote", merged.Content)
	})

	t.Run("InvalidLocalVectorClock_FallsBackToLWW", func(t *testing.T) {
		local := &Memory{
			ID:        "mem1",
			Content:   "local",
			UpdatedAt: now.Add(-time.Hour),
			Metadata:  map[string]interface{}{"vector_clock": "bad-json"},
		}
		remote := &MemoryEvent{
			Content:     "remote",
			Timestamp:   now,
			VectorClock: `{"node1":1}`,
		}

		merged := cr.Merge(local, remote)
		assert.Equal(t, "remote", merged.Content)
	})
}

// --- CRDTResolver: Merge with Custom ---

func TestCRDTResolver_Merge_Custom(t *testing.T) {
	t.Run("WithCustomFunction", func(t *testing.T) {
		cr := NewCRDTResolver(ConflictStrategyCustom)
		cr.WithCustomResolver(func(m *Memory, e *MemoryEvent) *Memory {
			return &Memory{
				ID:      m.ID,
				Content: "custom-resolved",
			}
		})

		local := &Memory{ID: "mem1", Content: "local"}
		remote := &MemoryEvent{Content: "remote", Timestamp: time.Now(), VectorClock: "{}"}

		merged := cr.Merge(local, remote)
		assert.Equal(t, "custom-resolved", merged.Content)
	})

	t.Run("NilCustomResolver_FallsBackToLWW", func(t *testing.T) {
		cr := NewCRDTResolver(ConflictStrategyCustom)
		now := time.Now()

		local := &Memory{ID: "mem1", Content: "local", UpdatedAt: now.Add(-time.Hour)}
		remote := &MemoryEvent{Content: "remote", Timestamp: now, VectorClock: "{}"}

		merged := cr.Merge(local, remote)
		assert.Equal(t, "remote", merged.Content)
	})
}

// --- CRDTResolver: Merge with unknown strategy ---

func TestCRDTResolver_Merge_UnknownStrategy(t *testing.T) {
	cr := NewCRDTResolver(ConflictStrategy("unknown"))
	now := time.Now()

	local := &Memory{ID: "mem1", Content: "local", UpdatedAt: now.Add(-time.Hour)}
	remote := &MemoryEvent{Content: "remote", Timestamp: now, VectorClock: "{}"}

	merged := cr.Merge(local, remote)
	assert.Equal(t, "remote", merged.Content)
}

// --- CRDTResolver: memoryFromEvent ---

func TestCRDTResolver_memoryFromEvent(t *testing.T) {
	cr := NewCRDTResolver(ConflictStrategyLastWriteWins)
	now := time.Now()

	event := &MemoryEvent{
		UserID:      "user1",
		SessionID:   "session1",
		Content:     "event content",
		Embedding:   []float32{0.1, 0.2},
		Importance:  0.7,
		Timestamp:   now,
		VectorClock: `{"node1":1}`,
		Tags:        []string{"tag1", "tag2"},
		Entities: []MemoryEntity{
			{ID: "e1", Name: "Entity1"},
		},
	}

	mem := cr.memoryFromEvent("mem1", event)

	assert.Equal(t, "mem1", mem.ID)
	assert.Equal(t, "user1", mem.UserID)
	assert.Equal(t, "session1", mem.SessionID)
	assert.Equal(t, "event content", mem.Content)
	assert.Equal(t, []float32{0.1, 0.2}, mem.Embedding)
	assert.Equal(t, 0.7, mem.Importance)
	assert.Equal(t, MemoryTypeSemantic, mem.Type)
	assert.Equal(t, `{"node1":1}`, mem.Metadata["vector_clock"])
	assert.NotNil(t, mem.Metadata["tags"])
	assert.NotNil(t, mem.Metadata["entities"])
}

func TestCRDTResolver_memoryFromEvent_NoTagsOrEntities(t *testing.T) {
	cr := NewCRDTResolver(ConflictStrategyLastWriteWins)
	event := &MemoryEvent{
		UserID:      "user1",
		Content:     "content",
		Timestamp:   time.Now(),
		VectorClock: `{}`,
	}

	mem := cr.memoryFromEvent("mem1", event)
	assert.NotNil(t, mem.Metadata)
	_, hasTags := mem.Metadata["tags"]
	_, hasEntities := mem.Metadata["entities"]
	assert.False(t, hasTags)
	assert.False(t, hasEntities)
}

// --- CRDTResolver: mergeTags ---

func TestCRDTResolver_mergeTags(t *testing.T) {
	cr := NewCRDTResolver(ConflictStrategyLastWriteWins)

	tests := []struct {
		name        string
		local       []string
		remote      []string
		expectedLen int
	}{
		{
			name:        "NoOverlap",
			local:       []string{"a", "b"},
			remote:      []string{"c", "d"},
			expectedLen: 4,
		},
		{
			name:        "WithOverlap",
			local:       []string{"a", "b"},
			remote:      []string{"b", "c"},
			expectedLen: 3,
		},
		{
			name:        "BothEmpty",
			local:       []string{},
			remote:      []string{},
			expectedLen: 0,
		},
		{
			name:        "OneEmpty",
			local:       []string{"a"},
			remote:      []string{},
			expectedLen: 1,
		},
		{
			name:        "Identical",
			local:       []string{"a", "b"},
			remote:      []string{"a", "b"},
			expectedLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cr.mergeTags(tt.local, tt.remote)
			assert.Len(t, result, tt.expectedLen)
		})
	}
}

// --- CRDTResolver: mergeEntities ---

func TestCRDTResolver_mergeEntities(t *testing.T) {
	cr := NewCRDTResolver(ConflictStrategyLastWriteWins)

	t.Run("NoOverlap", func(t *testing.T) {
		local := []MemoryEntity{{ID: "e1", Name: "A", Confidence: 0.5}}
		remote := []MemoryEntity{{ID: "e2", Name: "B", Confidence: 0.6}}

		result := cr.mergeEntities(local, remote)
		assert.Len(t, result, 2)
	})

	t.Run("OverlapRemoteHigherConfidence", func(t *testing.T) {
		local := []MemoryEntity{{ID: "e1", Name: "A", Confidence: 0.5}}
		remote := []MemoryEntity{{ID: "e1", Name: "A-Updated", Confidence: 0.9}}

		result := cr.mergeEntities(local, remote)
		assert.Len(t, result, 1)
		assert.Equal(t, "A-Updated", result[0].Name)
	})

	t.Run("OverlapLocalHigherConfidence", func(t *testing.T) {
		local := []MemoryEntity{{ID: "e1", Name: "A", Confidence: 0.9}}
		remote := []MemoryEntity{{ID: "e1", Name: "A-Updated", Confidence: 0.3}}

		result := cr.mergeEntities(local, remote)
		assert.Len(t, result, 1)
		assert.Equal(t, "A", result[0].Name)
	})

	t.Run("BothEmpty", func(t *testing.T) {
		result := cr.mergeEntities(nil, nil)
		assert.Empty(t, result)
	})
}

// --- CRDTResolver: DetectConflict ---

func TestCRDTResolver_DetectConflict(t *testing.T) {
	cr := NewCRDTResolver(ConflictStrategyLastWriteWins)
	now := time.Now()

	t.Run("NoConflict", func(t *testing.T) {
		local := &Memory{
			Content:    "same content",
			Importance: 0.5,
			UpdatedAt:  now.Add(-time.Hour),
			Metadata:   map[string]interface{}{"tags": []string{"a"}},
		}
		remote := &MemoryEvent{
			Content:    "same content",
			Importance: 0.5,
			Timestamp:  now,
			Tags:       []string{"a"},
		}

		hasConflict, desc := cr.DetectConflict(local, remote)
		assert.False(t, hasConflict)
		assert.Empty(t, desc)
	})

	t.Run("ContentConflict", func(t *testing.T) {
		local := &Memory{
			Content:    "local content",
			Importance: 0.5,
			UpdatedAt:  now,
			Metadata:   map[string]interface{}{"tags": []string{"a"}},
		}
		remote := &MemoryEvent{
			Content:    "remote content",
			Importance: 0.5,
			Timestamp:  now.Add(-time.Hour),
			Tags:       []string{"a"},
		}

		hasConflict, desc := cr.DetectConflict(local, remote)
		assert.True(t, hasConflict)
		assert.Contains(t, desc, "content")
	})

	t.Run("ImportanceConflict", func(t *testing.T) {
		local := &Memory{
			Content:    "content",
			Importance: 0.5,
			UpdatedAt:  now.Add(-time.Hour),
			Metadata:   map[string]interface{}{"tags": []string{"a"}},
		}
		remote := &MemoryEvent{
			Content:    "content",
			Importance: 0.9,
			Timestamp:  now,
			Tags:       []string{"a"},
		}

		hasConflict, desc := cr.DetectConflict(local, remote)
		assert.True(t, hasConflict)
		assert.Contains(t, desc, "importance")
	})

	t.Run("TagConflict", func(t *testing.T) {
		local := &Memory{
			Content:    "content",
			Importance: 0.5,
			UpdatedAt:  now.Add(-time.Hour),
			Metadata:   map[string]interface{}{"tags": []string{"a"}},
		}
		remote := &MemoryEvent{
			Content:    "content",
			Importance: 0.5,
			Timestamp:  now,
			Tags:       []string{"a", "b"},
		}

		hasConflict, desc := cr.DetectConflict(local, remote)
		assert.True(t, hasConflict)
		assert.Contains(t, desc, "tags")
	})

	t.Run("MultipleConflicts", func(t *testing.T) {
		local := &Memory{
			Content:    "local",
			Importance: 0.5,
			UpdatedAt:  now,
			Metadata:   map[string]interface{}{},
		}
		remote := &MemoryEvent{
			Content:    "remote",
			Importance: 0.9,
			Timestamp:  now.Add(-time.Hour),
			Tags:       []string{"tag"},
		}

		hasConflict, desc := cr.DetectConflict(local, remote)
		assert.True(t, hasConflict)
		assert.Contains(t, desc, "conflicts:")
	})

	t.Run("NilMetadataTags", func(t *testing.T) {
		local := &Memory{
			Content:    "content",
			Importance: 0.5,
			UpdatedAt:  now.Add(-time.Hour),
			Metadata:   nil,
		}
		remote := &MemoryEvent{
			Content:    "content",
			Importance: 0.5,
			Timestamp:  now,
			Tags:       nil,
		}

		hasConflict, _ := cr.DetectConflict(local, remote)
		assert.False(t, hasConflict)
	})
}

// --- CRDTResolver: tagsEqual ---

func TestCRDTResolver_tagsEqual(t *testing.T) {
	cr := NewCRDTResolver(ConflictStrategyLastWriteWins)

	tests := []struct {
		name     string
		a        []string
		b        []string
		expected bool
	}{
		{"BothNil", nil, nil, true},
		{"BothEmpty", []string{}, []string{}, true},
		{"Equal", []string{"a", "b"}, []string{"b", "a"}, true},
		{"DifferentLength", []string{"a"}, []string{"a", "b"}, false},
		{"DifferentContent", []string{"a", "b"}, []string{"a", "c"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, cr.tagsEqual(tt.a, tt.b))
		})
	}
}

// --- CRDTResolver: ResolveWithReport ---

func TestCRDTResolver_ResolveWithReport_NoConflict(t *testing.T) {
	cr := NewCRDTResolver(ConflictStrategyLastWriteWins)
	now := time.Now()

	local := &Memory{
		ID:         "mem1",
		Content:    "content",
		Importance: 0.5,
		UpdatedAt:  now.Add(-time.Hour),
		Metadata:   map[string]interface{}{},
	}
	remote := &MemoryEvent{
		Content:     "content",
		Importance:  0.5,
		Timestamp:   now,
		VectorClock: "{}",
	}

	report := cr.ResolveWithReport(local, remote)
	require.NotNil(t, report)
	assert.Equal(t, "mem1", report.MemoryID)
	assert.Equal(t, "none", report.ConflictType)
	assert.Equal(t, ConflictStrategyLastWriteWins, report.ResolvedWith)
	assert.Same(t, local, report.Resolution)
}

func TestCRDTResolver_ResolveWithReport_WithConflict(t *testing.T) {
	cr := NewCRDTResolver(ConflictStrategyLastWriteWins)
	now := time.Now()

	local := &Memory{
		ID:         "mem1",
		Content:    "local content",
		Importance: 0.5,
		UpdatedAt:  now.Add(-time.Hour),
		Metadata:   map[string]interface{}{},
	}
	remote := &MemoryEvent{
		Content:     "remote content",
		Importance:  0.9,
		Timestamp:   now,
		VectorClock: "{}",
	}

	report := cr.ResolveWithReport(local, remote)
	require.NotNil(t, report)
	assert.Contains(t, report.ConflictType, "importance")
	assert.NotNil(t, report.Resolution)
	assert.NotNil(t, report.Details)
	_, hasContentChanged := report.Details["content_changed"]
	assert.True(t, hasContentChanged)
	_, hasImportanceChanged := report.Details["importance_changed"]
	assert.True(t, hasImportanceChanged)
	_, hasTagsMerged := report.Details["tags_merged"]
	assert.True(t, hasTagsMerged)
}

func TestCRDTResolver_ResolveWithReport_MergeAllStrategy(t *testing.T) {
	cr := NewCRDTResolver(ConflictStrategyMergeAll)
	now := time.Now()

	local := &Memory{
		ID:         "mem1",
		Content:    "local",
		Importance: 0.3,
		UpdatedAt:  now,
		Metadata:   map[string]interface{}{"tags": []string{"a"}},
	}
	remote := &MemoryEvent{
		Content:     "remote longer content",
		Importance:  0.9,
		Timestamp:   now.Add(-time.Hour),
		Tags:        []string{"a", "b"},
		VectorClock: "{}",
	}

	report := cr.ResolveWithReport(local, remote)
	require.NotNil(t, report)
	assert.NotNil(t, report.Resolution)
	assert.Equal(t, ConflictStrategyMergeAll, report.ResolvedWith)
}
