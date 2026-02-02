package database

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// DebateLogRepository Extended Tests
// Tests for calculateExpiration, CreateTable, Insert, query methods, cleanup,
// and StartCleanupWorker using nil pool + panic recovery.
// =============================================================================

// -----------------------------------------------------------------------------
// calculateExpiration Tests
// -----------------------------------------------------------------------------

func TestDebateLogRepository_CalculateExpiration_NoExpiration(t *testing.T) {
	repo := NewDebateLogRepository(nil, nil, NoExpirationPolicy())
	result := repo.calculateExpiration()
	assert.Nil(t, result)
}

func TestDebateLogRepository_CalculateExpiration_RetentionDays(t *testing.T) {
	policy := LogRetentionPolicy{RetentionDays: 5, NoExpiration: false}
	repo := NewDebateLogRepository(nil, nil, policy)

	result := repo.calculateExpiration()
	require.NotNil(t, result)

	expected := time.Now().AddDate(0, 0, 5)
	assert.WithinDuration(t, expected, *result, 2*time.Second)
}

func TestDebateLogRepository_CalculateExpiration_RetentionTime(t *testing.T) {
	policy := LogRetentionPolicy{
		RetentionTime: 48 * time.Hour,
		NoExpiration:  false,
	}
	repo := NewDebateLogRepository(nil, nil, policy)

	result := repo.calculateExpiration()
	require.NotNil(t, result)

	expected := time.Now().Add(48 * time.Hour)
	assert.WithinDuration(t, expected, *result, 2*time.Second)
}

func TestDebateLogRepository_CalculateExpiration_RetentionTimeTakesPrecedence(t *testing.T) {
	policy := LogRetentionPolicy{
		RetentionDays: 10,
		RetentionTime: 1 * time.Hour,
		NoExpiration:  false,
	}
	repo := NewDebateLogRepository(nil, nil, policy)

	result := repo.calculateExpiration()
	require.NotNil(t, result)

	// RetentionTime is checked first, so 1 hour should win
	expected := time.Now().Add(1 * time.Hour)
	assert.WithinDuration(t, expected, *result, 2*time.Second)
}

func TestDebateLogRepository_CalculateExpiration_ZeroBoth_DefaultsFiveDay(t *testing.T) {
	policy := LogRetentionPolicy{
		RetentionDays: 0,
		RetentionTime: 0,
		NoExpiration:  false,
	}
	repo := NewDebateLogRepository(nil, nil, policy)

	result := repo.calculateExpiration()
	require.NotNil(t, result)

	expected := time.Now().AddDate(0, 0, 5)
	assert.WithinDuration(t, expected, *result, 2*time.Second)
}

// -----------------------------------------------------------------------------
// CreateTable Tests
// -----------------------------------------------------------------------------

func TestDebateLogRepository_CreateTable_NilPool(t *testing.T) {
	repo := NewDebateLogRepository(nil, nil, DefaultRetentionPolicy())

	err := safeCallError(func() error {
		return repo.CreateTable(context.Background())
	})
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// Insert Tests
// -----------------------------------------------------------------------------

func TestDebateLogRepository_Insert_NilPool(t *testing.T) {
	repo := NewDebateLogRepository(nil, nil, DefaultRetentionPolicy())

	entry := &DebateLogEntry{
		DebateID:              "debate-1",
		SessionID:             "sess-1",
		ParticipantID:         "part-1",
		ParticipantIdentifier: "DeepSeek-1",
		ParticipantName:       "DeepSeek",
		Role:                  "debater",
		Provider:              "deepseek",
		Model:                 "deepseek-chat",
		Round:                 1,
		Action:                "complete",
		ResponseTimeMs:        500,
		QualityScore:          0.9,
		TokensUsed:            100,
		ContentLength:         500,
	}

	err := safeCallError(func() error {
		return repo.Insert(context.Background(), entry)
	})
	assert.Error(t, err)

	// Verify that CreatedAt and ExpiresAt were set before the pool call
	assert.False(t, entry.CreatedAt.IsZero())
	assert.NotNil(t, entry.ExpiresAt)
}

func TestDebateLogRepository_Insert_NoExpiration_NilPool(t *testing.T) {
	repo := NewDebateLogRepository(nil, nil, NoExpirationPolicy())

	entry := &DebateLogEntry{
		DebateID:              "debate-2",
		ParticipantID:         "part-2",
		ParticipantIdentifier: "Claude-1",
		Provider:              "claude",
		Action:                "start",
	}

	err := safeCallError(func() error {
		return repo.Insert(context.Background(), entry)
	})
	assert.Error(t, err)

	assert.Nil(t, entry.ExpiresAt)
}

// -----------------------------------------------------------------------------
// Query Methods Tests
// -----------------------------------------------------------------------------

func TestDebateLogRepository_GetByDebateID_NilPool(t *testing.T) {
	repo := NewDebateLogRepository(nil, nil, DefaultRetentionPolicy())

	_, err := safeCallResult(func() ([]DebateLogEntry, error) {
		return repo.GetByDebateID(context.Background(), "debate-1")
	})
	assert.Error(t, err)
}

func TestDebateLogRepository_GetByParticipantIdentifier_NilPool(t *testing.T) {
	repo := NewDebateLogRepository(nil, nil, DefaultRetentionPolicy())

	_, err := safeCallResult(func() ([]DebateLogEntry, error) {
		return repo.GetByParticipantIdentifier(context.Background(), "DeepSeek-1")
	})
	assert.Error(t, err)
}

func TestDebateLogRepository_GetExpiredLogs_NilPool(t *testing.T) {
	repo := NewDebateLogRepository(nil, nil, DefaultRetentionPolicy())

	_, err := safeCallResult(func() ([]DebateLogEntry, error) {
		return repo.GetExpiredLogs(context.Background())
	})
	assert.Error(t, err)
}

func TestDebateLogRepository_GetLogsOlderThan_NilPool(t *testing.T) {
	repo := NewDebateLogRepository(nil, nil, DefaultRetentionPolicy())

	_, err := safeCallResult(func() ([]DebateLogEntry, error) {
		return repo.GetLogsOlderThan(context.Background(), 24*time.Hour)
	})
	assert.Error(t, err)
}

func TestDebateLogRepository_GetLogCount_NilPool(t *testing.T) {
	repo := NewDebateLogRepository(nil, nil, DefaultRetentionPolicy())

	_, err := safeCallResult(func() (int64, error) {
		return repo.GetLogCount(context.Background())
	})
	assert.Error(t, err)
}

func TestDebateLogRepository_GetLogStats_NilPool(t *testing.T) {
	repo := NewDebateLogRepository(nil, nil, DefaultRetentionPolicy())

	_, err := safeCallResult(func() (*LogStats, error) {
		return repo.GetLogStats(context.Background())
	})
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// Cleanup/Delete Methods Tests
// -----------------------------------------------------------------------------

func TestDebateLogRepository_CleanupExpiredLogs_NilPool(t *testing.T) {
	repo := NewDebateLogRepository(nil, nil, DefaultRetentionPolicy())

	_, err := safeCallResult(func() (int64, error) {
		return repo.CleanupExpiredLogs(context.Background())
	})
	assert.Error(t, err)
}

func TestDebateLogRepository_DeleteLogsOlderThan_NilPool(t *testing.T) {
	repo := NewDebateLogRepository(nil, nil, DefaultRetentionPolicy())

	_, err := safeCallResult(func() (int64, error) {
		return repo.DeleteLogsOlderThan(context.Background(), 7*24*time.Hour)
	})
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// StartCleanupWorker Tests
// -----------------------------------------------------------------------------

func TestDebateLogRepository_StartCleanupWorker_ContextCancel(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel) // suppress log output
	repo := NewDebateLogRepository(nil, logger, DefaultRetentionPolicy())

	ctx, cancel := context.WithCancel(context.Background())

	// Start cleanup worker
	repo.StartCleanupWorker(ctx, 50*time.Millisecond)

	// Give it a moment to start
	time.Sleep(20 * time.Millisecond)

	// Cancel context, worker should stop
	cancel()

	// Give it time to stop
	time.Sleep(100 * time.Millisecond)
}
