package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

// CodeVersionEntry represents a code snapshot at a debate milestone
type CodeVersionEntry struct {
	ID               string    `json:"id" db:"id"`
	SessionID        string    `json:"session_id" db:"session_id"`
	TurnID           *string   `json:"turn_id,omitempty" db:"turn_id"`
	Language         string    `json:"language,omitempty" db:"language"`
	Code             string    `json:"code" db:"code"`
	VersionNumber    int       `json:"version_number" db:"version_number"`
	QualityScore     *float64  `json:"quality_score,omitempty" db:"quality_score"`
	TestPassRate     *float64  `json:"test_pass_rate,omitempty" db:"test_pass_rate"`
	Metrics          string    `json:"metrics,omitempty" db:"metrics"`
	DiffFromPrevious string    `json:"diff_from_previous,omitempty" db:"diff_from_previous"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

// CodeVersionRepository manages code version storage
type CodeVersionRepository struct {
	pool *pgxpool.Pool
	log  *logrus.Logger
}

// NewCodeVersionRepository creates a new code version repository
func NewCodeVersionRepository(pool *pgxpool.Pool, log *logrus.Logger) *CodeVersionRepository {
	if log == nil {
		log = logrus.New()
	}
	return &CodeVersionRepository{
		pool: pool,
		log:  log,
	}
}

// CreateTable creates the code_versions table if it doesn't exist
func (r *CodeVersionRepository) CreateTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS code_versions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			session_id UUID NOT NULL REFERENCES debate_sessions(id) ON DELETE CASCADE,
			turn_id UUID REFERENCES debate_turns(id) ON DELETE SET NULL,
			language VARCHAR(50),
			code TEXT NOT NULL,
			version_number INTEGER NOT NULL,
			quality_score DECIMAL(5,4),
			test_pass_rate DECIMAL(5,4),
			metrics JSONB DEFAULT '{}',
			diff_from_previous TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			CONSTRAINT uq_code_versions_session_version UNIQUE (session_id, version_number)
		);

		CREATE INDEX IF NOT EXISTS idx_code_versions_session_id
			ON code_versions(session_id);
		CREATE INDEX IF NOT EXISTS idx_code_versions_turn_id
			ON code_versions(turn_id);
		CREATE INDEX IF NOT EXISTS idx_code_versions_session_version
			ON code_versions(session_id, version_number);
		CREATE INDEX IF NOT EXISTS idx_code_versions_language
			ON code_versions(language);
		CREATE INDEX IF NOT EXISTS idx_code_versions_quality
			ON code_versions(quality_score)
			WHERE quality_score IS NOT NULL;
		CREATE INDEX IF NOT EXISTS idx_code_versions_test_pass_rate
			ON code_versions(test_pass_rate)
			WHERE test_pass_rate IS NOT NULL;
		CREATE INDEX IF NOT EXISTS idx_code_versions_metrics
			ON code_versions USING GIN (metrics);
	`

	_, err := r.pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create code_versions table: %w", err)
	}

	r.log.Info("Code versions table created/verified")
	return nil
}

// Insert adds a new code version entry
func (r *CodeVersionRepository) Insert(ctx context.Context, entry *CodeVersionEntry) error {
	entry.CreatedAt = time.Now()

	query := `
		INSERT INTO code_versions (
			session_id, turn_id, language, code, version_number,
			quality_score, test_pass_rate, metrics, diff_from_previous, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`

	err := r.pool.QueryRow(ctx, query,
		entry.SessionID, entry.TurnID, entry.Language, entry.Code, entry.VersionNumber,
		entry.QualityScore, entry.TestPassRate, entry.Metrics, entry.DiffFromPrevious,
		entry.CreatedAt,
	).Scan(&entry.ID)

	if err != nil {
		return fmt.Errorf("failed to insert code version: %w", err)
	}

	r.log.WithFields(logrus.Fields{
		"id":             entry.ID,
		"session_id":     entry.SessionID,
		"version_number": entry.VersionNumber,
		"language":       entry.Language,
	}).Debug("Code version entry inserted")

	return nil
}

// GetByID retrieves a code version entry by its ID
func (r *CodeVersionRepository) GetByID(ctx context.Context, id string) (*CodeVersionEntry, error) {
	query := `
		SELECT id, session_id, turn_id, language, code, version_number,
			   quality_score, test_pass_rate, metrics, diff_from_previous, created_at
		FROM code_versions
		WHERE id = $1
	`

	entry, err := r.scanRow(r.pool.QueryRow(ctx, query, id))
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("code version not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get code version: %w", err)
	}

	return entry, nil
}

// GetLatestBySession retrieves the most recent code version for a session
func (r *CodeVersionRepository) GetLatestBySession(
	ctx context.Context, sessionID string,
) (*CodeVersionEntry, error) {
	query := `
		SELECT id, session_id, turn_id, language, code, version_number,
			   quality_score, test_pass_rate, metrics, diff_from_previous, created_at
		FROM code_versions
		WHERE session_id = $1
		ORDER BY version_number DESC
		LIMIT 1
	`

	entry, err := r.scanRow(r.pool.QueryRow(ctx, query, sessionID))
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("no code versions found for session: %s", sessionID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest code version: %w", err)
	}

	return entry, nil
}

// ListBySession retrieves all code versions for a session ordered by version number
func (r *CodeVersionRepository) ListBySession(
	ctx context.Context, sessionID string,
) ([]*CodeVersionEntry, error) {
	query := `
		SELECT id, session_id, turn_id, language, code, version_number,
			   quality_score, test_pass_rate, metrics, diff_from_previous, created_at
		FROM code_versions
		WHERE session_id = $1
		ORDER BY version_number ASC
	`

	rows, err := r.pool.Query(ctx, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list code versions by session: %w", err)
	}
	defer rows.Close()

	return r.scanRows(rows)
}

// GetBySessionAndVersion retrieves a specific code version by session and version number
func (r *CodeVersionRepository) GetBySessionAndVersion(
	ctx context.Context, sessionID string, versionNumber int,
) (*CodeVersionEntry, error) {
	query := `
		SELECT id, session_id, turn_id, language, code, version_number,
			   quality_score, test_pass_rate, metrics, diff_from_previous, created_at
		FROM code_versions
		WHERE session_id = $1 AND version_number = $2
	`

	entry, err := r.scanRow(r.pool.QueryRow(ctx, query, sessionID, versionNumber))
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf(
			"code version not found for session %s version %d", sessionID, versionNumber,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get code version by session and version: %w", err)
	}

	return entry, nil
}

// GetNextVersionNumber returns the next available version number for a session
func (r *CodeVersionRepository) GetNextVersionNumber(
	ctx context.Context, sessionID string,
) (int, error) {
	query := `
		SELECT COALESCE(MAX(version_number), 0) + 1
		FROM code_versions
		WHERE session_id = $1
	`

	var nextVersion int
	err := r.pool.QueryRow(ctx, query, sessionID).Scan(&nextVersion)
	if err != nil {
		return 0, fmt.Errorf("failed to get next version number: %w", err)
	}

	return nextVersion, nil
}

// Delete removes a code version entry by its ID
func (r *CodeVersionRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM code_versions WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete code version: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("code version not found: %s", id)
	}

	r.log.WithFields(logrus.Fields{
		"id": id,
	}).Debug("Code version entry deleted")

	return nil
}

// DeleteBySession removes all code version entries for a session
func (r *CodeVersionRepository) DeleteBySession(
	ctx context.Context, sessionID string,
) (int64, error) {
	query := `DELETE FROM code_versions WHERE session_id = $1`

	result, err := r.pool.Exec(ctx, query, sessionID)
	if err != nil {
		return 0, fmt.Errorf("failed to delete code versions by session: %w", err)
	}

	deleted := result.RowsAffected()

	if deleted > 0 {
		r.log.WithFields(logrus.Fields{
			"session_id":    sessionID,
			"deleted_count": deleted,
		}).Debug("Code version entries deleted by session")
	}

	return deleted, nil
}

// scanRow scans a single database row into a CodeVersionEntry
func (r *CodeVersionRepository) scanRow(row pgx.Row) (*CodeVersionEntry, error) {
	var entry CodeVersionEntry
	var turnID *string
	var language *string
	var qualityScore *float64
	var testPassRate *float64
	var metrics *string
	var diffFromPrevious *string

	err := row.Scan(
		&entry.ID, &entry.SessionID, &turnID, &language, &entry.Code,
		&entry.VersionNumber, &qualityScore, &testPassRate, &metrics,
		&diffFromPrevious, &entry.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if turnID != nil {
		entry.TurnID = turnID
	}
	if language != nil {
		entry.Language = *language
	}
	if qualityScore != nil {
		entry.QualityScore = qualityScore
	}
	if testPassRate != nil {
		entry.TestPassRate = testPassRate
	}
	if metrics != nil {
		entry.Metrics = *metrics
	}
	if diffFromPrevious != nil {
		entry.DiffFromPrevious = *diffFromPrevious
	}

	return &entry, nil
}

// scanRows scans database rows into a CodeVersionEntry slice
func (r *CodeVersionRepository) scanRows(rows pgx.Rows) ([]*CodeVersionEntry, error) {
	var entries []*CodeVersionEntry

	for rows.Next() {
		var entry CodeVersionEntry
		var turnID *string
		var language *string
		var qualityScore *float64
		var testPassRate *float64
		var metrics *string
		var diffFromPrevious *string

		err := rows.Scan(
			&entry.ID, &entry.SessionID, &turnID, &language, &entry.Code,
			&entry.VersionNumber, &qualityScore, &testPassRate, &metrics,
			&diffFromPrevious, &entry.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan code version row: %w", err)
		}

		if turnID != nil {
			entry.TurnID = turnID
		}
		if language != nil {
			entry.Language = *language
		}
		if qualityScore != nil {
			entry.QualityScore = qualityScore
		}
		if testPassRate != nil {
			entry.TestPassRate = testPassRate
		}
		if metrics != nil {
			entry.Metrics = *metrics
		}
		if diffFromPrevious != nil {
			entry.DiffFromPrevious = *diffFromPrevious
		}

		entries = append(entries, &entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating code version rows: %w", err)
	}

	return entries, nil
}
