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
// CodeVersionRepository Unit Tests
// These tests run without a database connection and validate struct creation,
// nil handling, JSON serialization, and nil-pool error paths.
// =============================================================================

// -----------------------------------------------------------------------------
// Constructor Tests
// -----------------------------------------------------------------------------

func TestNewCodeVersionRepository_NilLogger(t *testing.T) {
	// Creating a repository with a nil logger should not panic;
	// the constructor creates a default logger internally.
	require.NotPanics(t, func() {
		repo := NewCodeVersionRepository(nil, nil)
		require.NotNil(t, repo, "Repository should not be nil")
	})
}

func TestNewCodeVersionRepository_ValidLogger(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	repo := NewCodeVersionRepository(nil, logger)

	require.NotNil(t, repo, "Repository should not be nil")
}

// -----------------------------------------------------------------------------
// CodeVersionEntry Field Tests
// -----------------------------------------------------------------------------

func TestCodeVersionEntry_Fields(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	turnID := "turn-abc-123"
	qualityScore := 0.9150
	testPassRate := 0.9800

	entry := &CodeVersionEntry{
		ID:               "cv-001",
		SessionID:        "session-abc",
		TurnID:           &turnID,
		Language:         "go",
		Code:             "package main\n\nfunc main() {}\n",
		VersionNumber:    3,
		QualityScore:     &qualityScore,
		TestPassRate:     &testPassRate,
		Metrics:          `{"complexity":4,"lines":3}`,
		DiffFromPrevious: "--- a/main.go\n+++ b/main.go\n@@ -1 +1 @@\n",
		CreatedAt:        now,
	}

	assert.Equal(t, "cv-001", entry.ID)
	assert.Equal(t, "session-abc", entry.SessionID)
	require.NotNil(t, entry.TurnID)
	assert.Equal(t, "turn-abc-123", *entry.TurnID)
	assert.Equal(t, "go", entry.Language)
	assert.Equal(t, "package main\n\nfunc main() {}\n", entry.Code)
	assert.Equal(t, 3, entry.VersionNumber)
	require.NotNil(t, entry.QualityScore)
	assert.InDelta(t, 0.9150, *entry.QualityScore, 0.0001)
	require.NotNil(t, entry.TestPassRate)
	assert.InDelta(t, 0.9800, *entry.TestPassRate, 0.0001)
	assert.Equal(t, `{"complexity":4,"lines":3}`, entry.Metrics)
	assert.Contains(t, entry.DiffFromPrevious, "--- a/main.go")
	assert.Equal(t, now, entry.CreatedAt)
}

func TestCodeVersionEntry_NullableFields(t *testing.T) {
	t.Run("NilTurnID", func(t *testing.T) {
		entry := &CodeVersionEntry{
			ID:            "cv-nil-turn",
			SessionID:     "session-1",
			TurnID:        nil,
			Code:          "print('hello')",
			VersionNumber: 1,
		}

		assert.Nil(t, entry.TurnID,
			"TurnID should be nil when code version is not linked to a turn")
	})

	t.Run("NilQualityScore", func(t *testing.T) {
		entry := &CodeVersionEntry{
			ID:            "cv-nil-quality",
			SessionID:     "session-2",
			Code:          "def main(): pass",
			VersionNumber: 1,
			QualityScore:  nil,
		}

		assert.Nil(t, entry.QualityScore,
			"QualityScore should be nil when not yet evaluated")
	})

	t.Run("NilTestPassRate", func(t *testing.T) {
		entry := &CodeVersionEntry{
			ID:            "cv-nil-tpr",
			SessionID:     "session-3",
			Code:          "fn main() {}",
			VersionNumber: 1,
			TestPassRate:  nil,
		}

		assert.Nil(t, entry.TestPassRate,
			"TestPassRate should be nil when tests have not run")
	})

	t.Run("AllNullableFieldsNil", func(t *testing.T) {
		entry := &CodeVersionEntry{
			ID:            "cv-all-nil",
			SessionID:     "session-4",
			TurnID:        nil,
			Code:          "console.log('test')",
			VersionNumber: 1,
			QualityScore:  nil,
			TestPassRate:  nil,
		}

		assert.Nil(t, entry.TurnID)
		assert.Nil(t, entry.QualityScore)
		assert.Nil(t, entry.TestPassRate)
	})

	t.Run("NilFieldsWithEmptyOptionalStrings", func(t *testing.T) {
		entry := &CodeVersionEntry{
			ID:               "cv-empty-opt",
			SessionID:        "session-5",
			TurnID:           nil,
			Language:         "",
			Code:             "// empty",
			VersionNumber:    1,
			QualityScore:     nil,
			TestPassRate:     nil,
			Metrics:          "",
			DiffFromPrevious: "",
		}

		assert.Nil(t, entry.TurnID)
		assert.Empty(t, entry.Language)
		assert.Nil(t, entry.QualityScore)
		assert.Nil(t, entry.TestPassRate)
		assert.Empty(t, entry.Metrics)
		assert.Empty(t, entry.DiffFromPrevious)
	})
}

// -----------------------------------------------------------------------------
// JSON Serialization Tests
// -----------------------------------------------------------------------------

func TestCodeVersionEntry_JsonTags(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	turnID := "turn-json-abc"
	qualityScore := 0.8900
	testPassRate := 1.0000

	t.Run("MarshalFullEntry", func(t *testing.T) {
		entry := &CodeVersionEntry{
			ID:               "cv-json-1",
			SessionID:        "session-json",
			TurnID:           &turnID,
			Language:         "python",
			Code:             "def solve(): return 42",
			VersionNumber:    2,
			QualityScore:     &qualityScore,
			TestPassRate:     &testPassRate,
			Metrics:          `{"cyclomatic":1}`,
			DiffFromPrevious: "+def solve(): return 42",
			CreatedAt:        now,
		}

		data, err := json.Marshal(entry)
		require.NoError(t, err)

		var decoded map[string]interface{}
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "cv-json-1", decoded["id"])
		assert.Equal(t, "session-json", decoded["session_id"])
		assert.Equal(t, "turn-json-abc", decoded["turn_id"])
		assert.Equal(t, "python", decoded["language"])
		assert.Equal(t, "def solve(): return 42", decoded["code"])
		assert.Equal(t, float64(2), decoded["version_number"])
		assert.NotNil(t, decoded["quality_score"])
		assert.NotNil(t, decoded["test_pass_rate"])
		assert.Equal(t, `{"cyclomatic":1}`, decoded["metrics"])
		assert.Equal(t, "+def solve(): return 42", decoded["diff_from_previous"])
	})

	t.Run("MarshalAndUnmarshalRoundTrip", func(t *testing.T) {
		original := &CodeVersionEntry{
			ID:               "cv-roundtrip",
			SessionID:        "session-rt",
			TurnID:           &turnID,
			Language:         "rust",
			Code:             "fn main() { println!(\"hello\"); }",
			VersionNumber:    5,
			QualityScore:     &qualityScore,
			TestPassRate:     &testPassRate,
			Metrics:          `{"safety":"checked"}`,
			DiffFromPrevious: "+println!(\"hello\")",
			CreatedAt:        now,
		}

		data, err := json.Marshal(original)
		require.NoError(t, err)

		var restored CodeVersionEntry
		err = json.Unmarshal(data, &restored)
		require.NoError(t, err)

		assert.Equal(t, original.ID, restored.ID)
		assert.Equal(t, original.SessionID, restored.SessionID)
		require.NotNil(t, restored.TurnID)
		assert.Equal(t, *original.TurnID, *restored.TurnID)
		assert.Equal(t, original.Language, restored.Language)
		assert.Equal(t, original.Code, restored.Code)
		assert.Equal(t, original.VersionNumber, restored.VersionNumber)
		require.NotNil(t, restored.QualityScore)
		assert.InDelta(t, *original.QualityScore, *restored.QualityScore, 0.0001)
		require.NotNil(t, restored.TestPassRate)
		assert.InDelta(t, *original.TestPassRate, *restored.TestPassRate, 0.0001)
		assert.Equal(t, original.Metrics, restored.Metrics)
		assert.Equal(t, original.DiffFromPrevious, restored.DiffFromPrevious)
	})

	t.Run("OmitemptyFields", func(t *testing.T) {
		entry := &CodeVersionEntry{
			ID:            "cv-omit",
			SessionID:     "session-omit",
			Code:          "# minimal",
			VersionNumber: 1,
		}

		data, err := json.Marshal(entry)
		require.NoError(t, err)

		var decoded map[string]interface{}
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		// Fields with omitempty and nil/empty values should be absent
		_, hasTurnID := decoded["turn_id"]
		assert.False(t, hasTurnID,
			"turn_id with nil value and omitempty should be omitted")

		_, hasQualityScore := decoded["quality_score"]
		assert.False(t, hasQualityScore,
			"quality_score with nil value and omitempty should be omitted")

		_, hasTestPassRate := decoded["test_pass_rate"]
		assert.False(t, hasTestPassRate,
			"test_pass_rate with nil value and omitempty should be omitted")
	})
}

// -----------------------------------------------------------------------------
// Nil Pool Error Path Tests
// -----------------------------------------------------------------------------

func TestCodeVersionRepository_Insert_NilPool(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	repo := NewCodeVersionRepository(nil, logger)

	entry := &CodeVersionEntry{
		SessionID:     "session-nil-insert",
		Code:          "package main",
		VersionNumber: 1,
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
