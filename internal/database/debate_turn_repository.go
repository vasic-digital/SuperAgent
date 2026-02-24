package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

// debateTurnColumns defines the column list used by all SELECT queries.
// Kept in a single constant to ensure consistent ordering across methods.
const debateTurnColumns = `id, session_id, round, phase, agent_id, agent_role,
	provider, model, content, confidence, tool_calls, test_results,
	reflections, metadata, created_at, response_time_ms`

// DebateTurnEntry represents a single agent action within a debate round and phase
type DebateTurnEntry struct {
	ID             string    `json:"id" db:"id"`
	SessionID      string    `json:"session_id" db:"session_id"`
	Round          int       `json:"round" db:"round"`
	Phase          string    `json:"phase" db:"phase"`
	AgentID        string    `json:"agent_id" db:"agent_id"`
	AgentRole      string    `json:"agent_role,omitempty" db:"agent_role"`
	Provider       string    `json:"provider,omitempty" db:"provider"`
	Model          string    `json:"model,omitempty" db:"model"`
	Content        string    `json:"content,omitempty" db:"content"`
	Confidence     *float64  `json:"confidence,omitempty" db:"confidence"`
	ToolCalls      string    `json:"tool_calls,omitempty" db:"tool_calls"`
	TestResults    string    `json:"test_results,omitempty" db:"test_results"`
	Reflections    string    `json:"reflections,omitempty" db:"reflections"`
	Metadata       string    `json:"metadata,omitempty" db:"metadata"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	ResponseTimeMs *int      `json:"response_time_ms,omitempty" db:"response_time_ms"`
}

// DebateTurnRepository manages debate turn storage
type DebateTurnRepository struct {
	pool *pgxpool.Pool
	log  *logrus.Logger
}

// NewDebateTurnRepository creates a new debate turn repository
func NewDebateTurnRepository(pool *pgxpool.Pool, log *logrus.Logger) *DebateTurnRepository {
	if log == nil {
		log = logrus.New()
	}
	return &DebateTurnRepository{
		pool: pool,
		log:  log,
	}
}

// CreateTable creates the debate_turns table if it doesn't exist
func (r *DebateTurnRepository) CreateTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS debate_turns (
			id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
			session_id       UUID         NOT NULL REFERENCES debate_sessions(id) ON DELETE CASCADE,
			round            INTEGER      NOT NULL,
			phase            VARCHAR(50)  NOT NULL,
			agent_id         VARCHAR(255) NOT NULL,
			agent_role       VARCHAR(100),
			provider         VARCHAR(100),
			model            VARCHAR(255),
			content          TEXT,
			confidence       DECIMAL(5,4),
			tool_calls       JSONB        DEFAULT '[]',
			test_results     JSONB        DEFAULT '{}',
			reflections      JSONB        DEFAULT '[]',
			metadata         JSONB        DEFAULT '{}',
			created_at       TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			response_time_ms INTEGER,
			CONSTRAINT chk_debate_turns_phase CHECK (phase IN (
				'dehallucination', 'self_evolvement', 'proposal', 'critique',
				'review', 'optimization', 'adversarial', 'convergence'
			))
		);

		CREATE INDEX IF NOT EXISTS idx_debate_turns_session_id
			ON debate_turns(session_id);
		CREATE INDEX IF NOT EXISTS idx_debate_turns_session_round
			ON debate_turns(session_id, round);
		CREATE INDEX IF NOT EXISTS idx_debate_turns_phase
			ON debate_turns(phase);
		CREATE INDEX IF NOT EXISTS idx_debate_turns_agent
			ON debate_turns(agent_id);
		CREATE INDEX IF NOT EXISTS idx_debate_turns_session_round_phase
			ON debate_turns(session_id, round, phase);
		CREATE INDEX IF NOT EXISTS idx_debate_turns_created_at
			ON debate_turns(created_at);
	`

	_, err := r.pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create debate_turns table: %w", err)
	}

	r.log.Info("Debate turns table created/verified")
	return nil
}

// Insert adds a new debate turn entry
func (r *DebateTurnRepository) Insert(ctx context.Context, entry *DebateTurnEntry) error {
	entry.CreatedAt = time.Now()

	query := `
		INSERT INTO debate_turns (
			session_id, round, phase, agent_id, agent_role,
			provider, model, content, confidence, tool_calls,
			test_results, reflections, metadata, created_at, response_time_ms
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id
	`

	err := r.pool.QueryRow(ctx, query,
		entry.SessionID, entry.Round, entry.Phase, entry.AgentID, entry.AgentRole,
		entry.Provider, entry.Model, entry.Content, entry.Confidence, entry.ToolCalls,
		entry.TestResults, entry.Reflections, entry.Metadata, entry.CreatedAt,
		entry.ResponseTimeMs,
	).Scan(&entry.ID)

	if err != nil {
		return fmt.Errorf("failed to insert debate turn: %w", err)
	}

	r.log.WithFields(logrus.Fields{
		"id":         entry.ID,
		"session_id": entry.SessionID,
		"round":      entry.Round,
		"phase":      entry.Phase,
		"agent_id":   entry.AgentID,
	}).Debug("Debate turn entry inserted")

	return nil
}

// GetByID retrieves a single debate turn by its ID
func (r *DebateTurnRepository) GetByID(ctx context.Context, id string) (*DebateTurnEntry, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM debate_turns
		WHERE id = $1
	`, debateTurnColumns)

	entry, err := r.scanRow(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		return nil, fmt.Errorf("failed to get debate turn by id: %w", err)
	}

	return entry, nil
}

// ListBySession retrieves all turns for a specific debate session
func (r *DebateTurnRepository) ListBySession(
	ctx context.Context, sessionID string,
) ([]*DebateTurnEntry, error) {

	query := fmt.Sprintf(`
		SELECT %s
		FROM debate_turns
		WHERE session_id = $1
		ORDER BY round ASC, created_at ASC
	`, debateTurnColumns)

	rows, err := r.pool.Query(ctx, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list debate turns by session: %w", err)
	}
	defer rows.Close()

	return r.scanRows(rows)
}

// ListBySessionAndRound retrieves turns for a specific session and round
func (r *DebateTurnRepository) ListBySessionAndRound(
	ctx context.Context, sessionID string, round int,
) ([]*DebateTurnEntry, error) {

	query := fmt.Sprintf(`
		SELECT %s
		FROM debate_turns
		WHERE session_id = $1 AND round = $2
		ORDER BY created_at ASC
	`, debateTurnColumns)

	rows, err := r.pool.Query(ctx, query, sessionID, round)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to list debate turns by session and round: %w", err,
		)
	}
	defer rows.Close()

	return r.scanRows(rows)
}

// ListBySessionAndPhase retrieves turns for a specific session and phase
func (r *DebateTurnRepository) ListBySessionAndPhase(
	ctx context.Context, sessionID string, phase string,
) ([]*DebateTurnEntry, error) {

	query := fmt.Sprintf(`
		SELECT %s
		FROM debate_turns
		WHERE session_id = $1 AND phase = $2
		ORDER BY round ASC, created_at ASC
	`, debateTurnColumns)

	rows, err := r.pool.Query(ctx, query, sessionID, phase)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to list debate turns by session and phase: %w", err,
		)
	}
	defer rows.Close()

	return r.scanRows(rows)
}

// ListByAgent retrieves all turns produced by a specific agent
func (r *DebateTurnRepository) ListByAgent(
	ctx context.Context, agentID string,
) ([]*DebateTurnEntry, error) {

	query := fmt.Sprintf(`
		SELECT %s
		FROM debate_turns
		WHERE agent_id = $1
		ORDER BY created_at ASC
	`, debateTurnColumns)

	rows, err := r.pool.Query(ctx, query, agentID)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to list debate turns by agent: %w", err,
		)
	}
	defer rows.Close()

	return r.scanRows(rows)
}

// GetReflections retrieves non-empty reflection entries for a specific agent
// within a session
func (r *DebateTurnRepository) GetReflections(
	ctx context.Context, sessionID string, agentID string,
) ([]string, error) {

	query := `
		SELECT reflections
		FROM debate_turns
		WHERE session_id = $1 AND agent_id = $2
			AND reflections IS NOT NULL AND reflections != '[]'
		ORDER BY round ASC, created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, sessionID, agentID)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get debate turn reflections: %w", err,
		)
	}
	defer rows.Close()

	var reflections []string
	for rows.Next() {
		var ref *string
		if err := rows.Scan(&ref); err != nil {
			return nil, fmt.Errorf(
				"failed to scan debate turn reflection row: %w", err,
			)
		}
		if ref != nil {
			reflections = append(reflections, *ref)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"error iterating debate turn reflection rows: %w", err,
		)
	}

	return reflections, nil
}

// Delete removes a single debate turn by its ID
func (r *DebateTurnRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM debate_turns WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete debate turn: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("debate turn not found: %s", id)
	}

	r.log.WithFields(logrus.Fields{
		"id": id,
	}).Debug("Debate turn deleted")

	return nil
}

// DeleteBySession removes all turns for a specific session and returns the
// number of rows deleted
func (r *DebateTurnRepository) DeleteBySession(
	ctx context.Context, sessionID string,
) (int64, error) {

	query := `DELETE FROM debate_turns WHERE session_id = $1`

	result, err := r.pool.Exec(ctx, query, sessionID)
	if err != nil {
		return 0, fmt.Errorf(
			"failed to delete debate turns by session: %w", err,
		)
	}

	deleted := result.RowsAffected()

	if deleted > 0 {
		r.log.WithFields(logrus.Fields{
			"session_id":    sessionID,
			"deleted_count": deleted,
		}).Info("Debate turns deleted by session")
	}

	return deleted, nil
}

// scanRow scans a single row into a DebateTurnEntry
func (r *DebateTurnRepository) scanRow(row pgx.Row) (*DebateTurnEntry, error) {
	var entry DebateTurnEntry
	var agentRole, provider, model, content *string
	var confidence *float64
	var toolCalls, testResults, reflections, metadata *string
	var responseTimeMs *int

	err := row.Scan(
		&entry.ID, &entry.SessionID, &entry.Round, &entry.Phase,
		&entry.AgentID, &agentRole, &provider, &model, &content,
		&confidence, &toolCalls, &testResults, &reflections, &metadata,
		&entry.CreatedAt, &responseTimeMs,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan debate turn row: %w", err)
	}

	if agentRole != nil {
		entry.AgentRole = *agentRole
	}
	if provider != nil {
		entry.Provider = *provider
	}
	if model != nil {
		entry.Model = *model
	}
	if content != nil {
		entry.Content = *content
	}
	if confidence != nil {
		entry.Confidence = confidence
	}
	if toolCalls != nil {
		entry.ToolCalls = *toolCalls
	}
	if testResults != nil {
		entry.TestResults = *testResults
	}
	if reflections != nil {
		entry.Reflections = *reflections
	}
	if metadata != nil {
		entry.Metadata = *metadata
	}
	if responseTimeMs != nil {
		entry.ResponseTimeMs = responseTimeMs
	}

	return &entry, nil
}

// scanRows scans multiple database rows into a DebateTurnEntry slice
func (r *DebateTurnRepository) scanRows(
	rows pgx.Rows,
) ([]*DebateTurnEntry, error) {

	var entries []*DebateTurnEntry

	for rows.Next() {
		var entry DebateTurnEntry
		var agentRole, provider, model, content *string
		var confidence *float64
		var toolCalls, testResults, reflections, metadata *string
		var responseTimeMs *int

		err := rows.Scan(
			&entry.ID, &entry.SessionID, &entry.Round, &entry.Phase,
			&entry.AgentID, &agentRole, &provider, &model, &content,
			&confidence, &toolCalls, &testResults, &reflections,
			&metadata, &entry.CreatedAt, &responseTimeMs,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan debate turn row: %w", err)
		}

		if agentRole != nil {
			entry.AgentRole = *agentRole
		}
		if provider != nil {
			entry.Provider = *provider
		}
		if model != nil {
			entry.Model = *model
		}
		if content != nil {
			entry.Content = *content
		}
		if confidence != nil {
			entry.Confidence = confidence
		}
		if toolCalls != nil {
			entry.ToolCalls = *toolCalls
		}
		if testResults != nil {
			entry.TestResults = *testResults
		}
		if reflections != nil {
			entry.Reflections = *reflections
		}
		if metadata != nil {
			entry.Metadata = *metadata
		}
		if responseTimeMs != nil {
			entry.ResponseTimeMs = responseTimeMs
		}

		entries = append(entries, &entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating debate turn rows: %w", err)
	}

	return entries, nil
}
