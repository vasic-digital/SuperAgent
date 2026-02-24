package database

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// DebateSessionRepository Unit Tests
// These tests run without a database connection and validate struct creation,
// nil handling, JSON serialization, and nil-pool error paths.
// =============================================================================

// -----------------------------------------------------------------------------
// Constructor Tests
// -----------------------------------------------------------------------------

func TestNewDebateSessionRepository_NilLogger(t *testing.T) {
	// Creating a repository with a nil logger should not panic;
	// the constructor creates a default logger internally.
	require.NotPanics(t, func() {
		repo := NewDebateSessionRepository(nil, nil)
		require.NotNil(t, repo, "Repository should not be nil")
	})
}

func TestNewDebateSessionRepository_ValidLogger(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	repo := NewDebateSessionRepository(nil, logger)

	require.NotNil(t, repo, "Repository should not be nil")
}

// -----------------------------------------------------------------------------
// DebateSessionEntry Field Tests
// -----------------------------------------------------------------------------

func TestDebateSessionEntry_Fields(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	completedAt := now.Add(5 * time.Minute)
	consensusScore := 0.9512

	entry := &DebateSessionEntry{
		ID:                   "session-001",
		DebateID:             "debate-abc",
		Topic:                "Code review quality",
		Status:               "completed",
		TopologyType:         "mesh",
		CoordinationProtocol: "round-robin",
		Config:               `{"max_rounds":5}`,
		InitiatedBy:          "user-42",
		CreatedAt:            now,
		UpdatedAt:            now,
		CompletedAt:          &completedAt,
		TotalRounds:          3,
		FinalConsensusScore:  &consensusScore,
		Outcome:              "consensus reached",
		Metadata:             `{"tags":["code","review"]}`,
	}

	assert.Equal(t, "session-001", entry.ID)
	assert.Equal(t, "debate-abc", entry.DebateID)
	assert.Equal(t, "Code review quality", entry.Topic)
	assert.Equal(t, "completed", entry.Status)
	assert.Equal(t, "mesh", entry.TopologyType)
	assert.Equal(t, "round-robin", entry.CoordinationProtocol)
	assert.Equal(t, `{"max_rounds":5}`, entry.Config)
	assert.Equal(t, "user-42", entry.InitiatedBy)
	assert.Equal(t, now, entry.CreatedAt)
	assert.Equal(t, now, entry.UpdatedAt)
	require.NotNil(t, entry.CompletedAt)
	assert.Equal(t, completedAt, *entry.CompletedAt)
	assert.Equal(t, 3, entry.TotalRounds)
	require.NotNil(t, entry.FinalConsensusScore)
	assert.InDelta(t, 0.9512, *entry.FinalConsensusScore, 0.0001)
	assert.Equal(t, "consensus reached", entry.Outcome)
	assert.Equal(t, `{"tags":["code","review"]}`, entry.Metadata)
}

func TestDebateSessionEntry_NullableFields(t *testing.T) {
	t.Run("NilCompletedAt", func(t *testing.T) {
		entry := &DebateSessionEntry{
			ID:          "session-nil-time",
			DebateID:    "debate-1",
			Topic:       "Testing nil fields",
			Status:      "pending",
			CompletedAt: nil,
		}

		assert.Nil(t, entry.CompletedAt, "CompletedAt should be nil for pending sessions")
	})

	t.Run("NilFinalConsensusScore", func(t *testing.T) {
		entry := &DebateSessionEntry{
			ID:                  "session-nil-score",
			DebateID:            "debate-2",
			Topic:               "Testing nil consensus",
			Status:              "running",
			FinalConsensusScore: nil,
		}

		assert.Nil(t, entry.FinalConsensusScore,
			"FinalConsensusScore should be nil for in-progress sessions")
	})

	t.Run("BothNilFields", func(t *testing.T) {
		entry := &DebateSessionEntry{
			ID:                  "session-both-nil",
			DebateID:            "debate-3",
			Topic:               "Both nil",
			Status:              "pending",
			CompletedAt:         nil,
			FinalConsensusScore: nil,
		}

		assert.Nil(t, entry.CompletedAt)
		assert.Nil(t, entry.FinalConsensusScore)
	})

	t.Run("NilFieldsWithPopulatedOptionalStrings", func(t *testing.T) {
		entry := &DebateSessionEntry{
			ID:                  "session-mixed",
			DebateID:            "debate-4",
			Topic:               "Mixed nullability",
			Status:              "completed",
			TopologyType:        "",
			CoordinationProtocol: "",
			Config:              "",
			InitiatedBy:         "",
			CompletedAt:         nil,
			FinalConsensusScore: nil,
			Outcome:             "",
			Metadata:            "",
		}

		assert.Empty(t, entry.TopologyType)
		assert.Empty(t, entry.CoordinationProtocol)
		assert.Empty(t, entry.Config)
		assert.Empty(t, entry.InitiatedBy)
		assert.Nil(t, entry.CompletedAt)
		assert.Nil(t, entry.FinalConsensusScore)
		assert.Empty(t, entry.Outcome)
		assert.Empty(t, entry.Metadata)
	})
}

// -----------------------------------------------------------------------------
// JSON Serialization Tests
// -----------------------------------------------------------------------------

func TestDebateSessionEntry_JsonTags(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	completedAt := now.Add(10 * time.Minute)
	score := 0.8765

	t.Run("MarshalFullEntry", func(t *testing.T) {
		entry := &DebateSessionEntry{
			ID:                   "session-json-1",
			DebateID:             "debate-json",
			Topic:                "JSON serialization test",
			Status:               "completed",
			TopologyType:         "star",
			CoordinationProtocol: "consensus",
			Config:               `{"depth":3}`,
			InitiatedBy:          "admin",
			CreatedAt:            now,
			UpdatedAt:            now,
			CompletedAt:          &completedAt,
			TotalRounds:          5,
			FinalConsensusScore:  &score,
			Outcome:              "full agreement",
			Metadata:             `{"version":"1.0"}`,
		}

		data, err := json.Marshal(entry)
		require.NoError(t, err)

		var decoded map[string]interface{}
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "session-json-1", decoded["id"])
		assert.Equal(t, "debate-json", decoded["debate_id"])
		assert.Equal(t, "JSON serialization test", decoded["topic"])
		assert.Equal(t, "completed", decoded["status"])
		assert.Equal(t, "star", decoded["topology_type"])
		assert.Equal(t, "consensus", decoded["coordination_protocol"])
		assert.Equal(t, `{"depth":3}`, decoded["config"])
		assert.Equal(t, "admin", decoded["initiated_by"])
		assert.Equal(t, float64(5), decoded["total_rounds"])
		assert.Equal(t, "full agreement", decoded["outcome"])
		assert.Equal(t, `{"version":"1.0"}`, decoded["metadata"])
		assert.NotNil(t, decoded["completed_at"])
		assert.NotNil(t, decoded["final_consensus_score"])
	})

	t.Run("MarshalAndUnmarshalRoundTrip", func(t *testing.T) {
		original := &DebateSessionEntry{
			ID:                   "session-roundtrip",
			DebateID:             "debate-rt",
			Topic:                "Round trip test",
			Status:               "pending",
			TopologyType:         "chain",
			CoordinationProtocol: "sequential",
			Config:               `{}`,
			InitiatedBy:          "system",
			CreatedAt:            now,
			UpdatedAt:            now,
			CompletedAt:          nil,
			TotalRounds:          0,
			FinalConsensusScore:  nil,
			Outcome:              "",
			Metadata:             "",
		}

		data, err := json.Marshal(original)
		require.NoError(t, err)

		var restored DebateSessionEntry
		err = json.Unmarshal(data, &restored)
		require.NoError(t, err)

		assert.Equal(t, original.ID, restored.ID)
		assert.Equal(t, original.DebateID, restored.DebateID)
		assert.Equal(t, original.Topic, restored.Topic)
		assert.Equal(t, original.Status, restored.Status)
		assert.Equal(t, original.TopologyType, restored.TopologyType)
		assert.Equal(t, original.CoordinationProtocol, restored.CoordinationProtocol)
		assert.Equal(t, original.TotalRounds, restored.TotalRounds)
	})

	t.Run("OmitemptyFields", func(t *testing.T) {
		entry := &DebateSessionEntry{
			ID:       "session-omit",
			DebateID: "debate-omit",
			Topic:    "Omitempty test",
			Status:   "pending",
		}

		data, err := json.Marshal(entry)
		require.NoError(t, err)

		var decoded map[string]interface{}
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		// Fields with omitempty and empty values should be absent
		_, hasCompletedAt := decoded["completed_at"]
		assert.False(t, hasCompletedAt,
			"completed_at with nil value and omitempty should be omitted")

		_, hasConsensusScore := decoded["final_consensus_score"]
		assert.False(t, hasConsensusScore,
			"final_consensus_score with nil value and omitempty should be omitted")
	})
}

// -----------------------------------------------------------------------------
// Nil Pool Error Path Tests
// -----------------------------------------------------------------------------

func TestDebateSessionRepository_CreateTable_NilPool(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	repo := NewDebateSessionRepository(nil, logger)

	var err error
	panicked := false

	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		err = repo.CreateTable(context.Background())
	}()

	// With a nil pool, calling Exec will panic (nil pointer dereference)
	// or return an error. Either outcome confirms the nil pool path is exercised.
	assert.True(t, panicked || err != nil,
		"CreateTable with nil pool should panic or return an error")
}

func TestDebateSessionRepository_Insert_NilPool(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	repo := NewDebateSessionRepository(nil, logger)

	entry := &DebateSessionEntry{
		ID:       "test-insert-nil",
		DebateID: "debate-nil",
		Topic:    "Nil pool insert test",
		Status:   "pending",
	}

	var err error
	panicked := false

	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		err = repo.Insert(context.Background(), entry)
	}()

	assert.True(t, panicked || err != nil,
		"Insert with nil pool should panic or return an error")
}

func TestDebateSessionRepository_GetByID_NilPool(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	repo := NewDebateSessionRepository(nil, logger)

	var result *DebateSessionEntry
	var err error
	panicked := false

	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		result, err = repo.GetByID(context.Background(), "nonexistent-id")
	}()

	assert.True(t, panicked || err != nil,
		"GetByID with nil pool should panic or return an error")

	if !panicked {
		assert.Nil(t, result, "Result should be nil when error occurs")
	}
}
