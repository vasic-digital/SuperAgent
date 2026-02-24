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
// DebateTurnRepository Unit Tests
// These tests run without a database connection and validate struct creation,
// nil handling, JSON serialization, phase validation, and nil-pool error paths.
// =============================================================================

// -----------------------------------------------------------------------------
// Constructor Tests
// -----------------------------------------------------------------------------

func TestNewDebateTurnRepository_NilLogger(t *testing.T) {
	// Creating a repository with a nil logger should not panic;
	// the constructor creates a default logger internally.
	require.NotPanics(t, func() {
		repo := NewDebateTurnRepository(nil, nil)
		require.NotNil(t, repo, "Repository should not be nil")
	})
}

func TestNewDebateTurnRepository_ValidLogger(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	repo := NewDebateTurnRepository(nil, logger)

	require.NotNil(t, repo, "Repository should not be nil")
}

// -----------------------------------------------------------------------------
// DebateTurnEntry Field Tests
// -----------------------------------------------------------------------------

func TestDebateTurnEntry_Fields(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	confidence := 0.9234
	responseTime := 1540

	entry := &DebateTurnEntry{
		ID:             "turn-001",
		SessionID:      "session-abc",
		Round:          2,
		Phase:          "proposal",
		AgentID:        "agent-deepseek-1",
		AgentRole:      "proposer",
		Provider:       "deepseek",
		Model:          "deepseek-chat",
		Content:        "I propose the following solution...",
		Confidence:     &confidence,
		ToolCalls:      `[{"name":"search","args":{}}]`,
		TestResults:    `{"passed":5,"failed":0}`,
		Reflections:    `["Considered edge cases","Reviewed prior round"]`,
		Metadata:       `{"source":"ensemble"}`,
		CreatedAt:      now,
		ResponseTimeMs: &responseTime,
	}

	assert.Equal(t, "turn-001", entry.ID)
	assert.Equal(t, "session-abc", entry.SessionID)
	assert.Equal(t, 2, entry.Round)
	assert.Equal(t, "proposal", entry.Phase)
	assert.Equal(t, "agent-deepseek-1", entry.AgentID)
	assert.Equal(t, "proposer", entry.AgentRole)
	assert.Equal(t, "deepseek", entry.Provider)
	assert.Equal(t, "deepseek-chat", entry.Model)
	assert.Equal(t, "I propose the following solution...", entry.Content)
	require.NotNil(t, entry.Confidence)
	assert.InDelta(t, 0.9234, *entry.Confidence, 0.0001)
	assert.Equal(t, `[{"name":"search","args":{}}]`, entry.ToolCalls)
	assert.Equal(t, `{"passed":5,"failed":0}`, entry.TestResults)
	assert.Equal(t, `["Considered edge cases","Reviewed prior round"]`, entry.Reflections)
	assert.Equal(t, `{"source":"ensemble"}`, entry.Metadata)
	assert.Equal(t, now, entry.CreatedAt)
	require.NotNil(t, entry.ResponseTimeMs)
	assert.Equal(t, 1540, *entry.ResponseTimeMs)
}

func TestDebateTurnEntry_NullableFields(t *testing.T) {
	t.Run("NilConfidence", func(t *testing.T) {
		entry := &DebateTurnEntry{
			ID:         "turn-nil-conf",
			SessionID:  "session-1",
			Round:      1,
			Phase:      "proposal",
			AgentID:    "agent-1",
			Confidence: nil,
		}

		assert.Nil(t, entry.Confidence,
			"Confidence should be nil when not measured")
	})

	t.Run("NilResponseTimeMs", func(t *testing.T) {
		entry := &DebateTurnEntry{
			ID:             "turn-nil-rt",
			SessionID:      "session-1",
			Round:          1,
			Phase:          "critique",
			AgentID:        "agent-2",
			ResponseTimeMs: nil,
		}

		assert.Nil(t, entry.ResponseTimeMs,
			"ResponseTimeMs should be nil when not tracked")
	})

	t.Run("BothNilFields", func(t *testing.T) {
		entry := &DebateTurnEntry{
			ID:             "turn-both-nil",
			SessionID:      "session-2",
			Round:          3,
			Phase:          "review",
			AgentID:        "agent-3",
			Confidence:     nil,
			ResponseTimeMs: nil,
		}

		assert.Nil(t, entry.Confidence)
		assert.Nil(t, entry.ResponseTimeMs)
	})

	t.Run("NilFieldsWithEmptyOptionalStrings", func(t *testing.T) {
		entry := &DebateTurnEntry{
			ID:             "turn-empty-opt",
			SessionID:      "session-3",
			Round:          1,
			Phase:          "proposal",
			AgentID:        "agent-4",
			AgentRole:      "",
			Provider:       "",
			Model:          "",
			Content:        "",
			Confidence:     nil,
			ToolCalls:      "",
			TestResults:    "",
			Reflections:    "",
			Metadata:       "",
			ResponseTimeMs: nil,
		}

		assert.Empty(t, entry.AgentRole)
		assert.Empty(t, entry.Provider)
		assert.Empty(t, entry.Model)
		assert.Empty(t, entry.Content)
		assert.Nil(t, entry.Confidence)
		assert.Empty(t, entry.ToolCalls)
		assert.Empty(t, entry.TestResults)
		assert.Empty(t, entry.Reflections)
		assert.Empty(t, entry.Metadata)
		assert.Nil(t, entry.ResponseTimeMs)
	})
}

// -----------------------------------------------------------------------------
// JSON Serialization Tests
// -----------------------------------------------------------------------------

func TestDebateTurnEntry_JsonTags(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	confidence := 0.85
	responseTime := 2300

	t.Run("MarshalFullEntry", func(t *testing.T) {
		entry := &DebateTurnEntry{
			ID:             "turn-json-1",
			SessionID:      "session-json",
			Round:          1,
			Phase:          "dehallucination",
			AgentID:        "agent-claude-1",
			AgentRole:      "validator",
			Provider:       "claude",
			Model:          "claude-3-opus",
			Content:        "Validating the proposal...",
			Confidence:     &confidence,
			ToolCalls:      `[]`,
			TestResults:    `{"all":"pass"}`,
			Reflections:    `["Good coverage"]`,
			Metadata:       `{"round_context":"first"}`,
			CreatedAt:      now,
			ResponseTimeMs: &responseTime,
		}

		data, err := json.Marshal(entry)
		require.NoError(t, err)

		var decoded map[string]interface{}
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "turn-json-1", decoded["id"])
		assert.Equal(t, "session-json", decoded["session_id"])
		assert.Equal(t, float64(1), decoded["round"])
		assert.Equal(t, "dehallucination", decoded["phase"])
		assert.Equal(t, "agent-claude-1", decoded["agent_id"])
		assert.Equal(t, "validator", decoded["agent_role"])
		assert.Equal(t, "claude", decoded["provider"])
		assert.Equal(t, "claude-3-opus", decoded["model"])
		assert.Equal(t, "Validating the proposal...", decoded["content"])
		assert.NotNil(t, decoded["confidence"])
		assert.NotNil(t, decoded["response_time_ms"])
	})

	t.Run("MarshalAndUnmarshalRoundTrip", func(t *testing.T) {
		original := &DebateTurnEntry{
			ID:             "turn-roundtrip",
			SessionID:      "session-rt",
			Round:          3,
			Phase:          "convergence",
			AgentID:        "agent-gemini-2",
			AgentRole:      "synthesizer",
			Provider:       "gemini",
			Model:          "gemini-pro",
			Content:        "Final synthesis...",
			Confidence:     &confidence,
			ToolCalls:      `[]`,
			TestResults:    `{}`,
			Reflections:    `[]`,
			Metadata:       `{}`,
			CreatedAt:      now,
			ResponseTimeMs: &responseTime,
		}

		data, err := json.Marshal(original)
		require.NoError(t, err)

		var restored DebateTurnEntry
		err = json.Unmarshal(data, &restored)
		require.NoError(t, err)

		assert.Equal(t, original.ID, restored.ID)
		assert.Equal(t, original.SessionID, restored.SessionID)
		assert.Equal(t, original.Round, restored.Round)
		assert.Equal(t, original.Phase, restored.Phase)
		assert.Equal(t, original.AgentID, restored.AgentID)
		assert.Equal(t, original.AgentRole, restored.AgentRole)
		assert.Equal(t, original.Provider, restored.Provider)
		assert.Equal(t, original.Model, restored.Model)
		assert.Equal(t, original.Content, restored.Content)
		require.NotNil(t, restored.Confidence)
		assert.InDelta(t, *original.Confidence, *restored.Confidence, 0.0001)
		require.NotNil(t, restored.ResponseTimeMs)
		assert.Equal(t, *original.ResponseTimeMs, *restored.ResponseTimeMs)
	})

	t.Run("OmitemptyFields", func(t *testing.T) {
		entry := &DebateTurnEntry{
			ID:        "turn-omit",
			SessionID: "session-omit",
			Round:     1,
			Phase:     "proposal",
			AgentID:   "agent-1",
		}

		data, err := json.Marshal(entry)
		require.NoError(t, err)

		var decoded map[string]interface{}
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		// Fields with omitempty and zero/nil values should be absent
		_, hasConfidence := decoded["confidence"]
		assert.False(t, hasConfidence,
			"confidence with nil value and omitempty should be omitted")

		_, hasResponseTimeMs := decoded["response_time_ms"]
		assert.False(t, hasResponseTimeMs,
			"response_time_ms with nil value and omitempty should be omitted")
	})
}

// -----------------------------------------------------------------------------
// Phase Validation Tests (table-driven)
// -----------------------------------------------------------------------------

func TestDebateTurnEntry_ValidPhases(t *testing.T) {
	// These are the 8 valid phases enforced by the database CHECK constraint
	validPhases := []struct {
		phase       string
		description string
	}{
		{
			phase:       "dehallucination",
			description: "Fact-checking and hallucination detection phase",
		},
		{
			phase:       "self_evolvement",
			description: "Self-improvement and learning phase",
		},
		{
			phase:       "proposal",
			description: "Initial solution proposal phase",
		},
		{
			phase:       "critique",
			description: "Critical review and feedback phase",
		},
		{
			phase:       "review",
			description: "Peer review and validation phase",
		},
		{
			phase:       "optimization",
			description: "Performance and quality optimization phase",
		},
		{
			phase:       "adversarial",
			description: "Adversarial challenge and stress-test phase",
		},
		{
			phase:       "convergence",
			description: "Final consensus-building and convergence phase",
		},
	}

	for _, tc := range validPhases {
		t.Run(tc.phase, func(t *testing.T) {
			entry := &DebateTurnEntry{
				ID:        "turn-phase-" + tc.phase,
				SessionID: "session-phase-test",
				Round:     1,
				Phase:     tc.phase,
				AgentID:   "agent-phase-tester",
			}

			assert.Equal(t, tc.phase, entry.Phase,
				"Phase should be set correctly for: %s", tc.description)
			assert.NotEmpty(t, entry.Phase,
				"Phase should not be empty")

			// Verify the entry can be serialized with this phase
			data, err := json.Marshal(entry)
			require.NoError(t, err)
			assert.Contains(t, string(data), tc.phase)
		})
	}

	// Verify we tested exactly 8 phases
	assert.Len(t, validPhases, 8,
		"There should be exactly 8 valid debate phases")
}

// -----------------------------------------------------------------------------
// Nil Pool Error Path Tests
// -----------------------------------------------------------------------------

func TestDebateTurnRepository_Insert_NilPool(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	repo := NewDebateTurnRepository(nil, logger)

	entry := &DebateTurnEntry{
		SessionID: "session-nil-test",
		Round:     1,
		Phase:     "proposal",
		AgentID:   "agent-nil-test",
		Content:   "This should fail with nil pool",
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
