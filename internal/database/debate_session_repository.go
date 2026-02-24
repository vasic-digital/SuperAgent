package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

// DebateSessionEntry represents a debate session record
type DebateSessionEntry struct {
	ID                   string     `json:"id" db:"id"`
	DebateID             string     `json:"debate_id" db:"debate_id"`
	Topic                string     `json:"topic" db:"topic"`
	Status               string     `json:"status" db:"status"`
	TopologyType         string     `json:"topology_type,omitempty" db:"topology_type"`
	CoordinationProtocol string     `json:"coordination_protocol,omitempty" db:"coordination_protocol"`
	Config               string     `json:"config,omitempty" db:"config"`
	InitiatedBy          string     `json:"initiated_by,omitempty" db:"initiated_by"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at" db:"updated_at"`
	CompletedAt          *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	TotalRounds          int        `json:"total_rounds" db:"total_rounds"`
	FinalConsensusScore  *float64   `json:"final_consensus_score,omitempty" db:"final_consensus_score"`
	Outcome              string     `json:"outcome,omitempty" db:"outcome"`
	Metadata             string     `json:"metadata,omitempty" db:"metadata"`
}

// DebateSessionRepository manages debate session storage
type DebateSessionRepository struct {
	pool *pgxpool.Pool
	log  *logrus.Logger
}

// NewDebateSessionRepository creates a new debate session repository
func NewDebateSessionRepository(pool *pgxpool.Pool, log *logrus.Logger) *DebateSessionRepository {
	if log == nil {
		log = logrus.New()
	}
	return &DebateSessionRepository{
		pool: pool,
		log:  log,
	}
}

// CreateTable creates the debate_sessions table if it doesn't exist
func (r *DebateSessionRepository) CreateTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS debate_sessions (
			id VARCHAR(255) PRIMARY KEY,
			debate_id VARCHAR(255) NOT NULL,
			topic TEXT NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			topology_type VARCHAR(100),
			coordination_protocol VARCHAR(100),
			config JSONB,
			initiated_by VARCHAR(255),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			completed_at TIMESTAMP WITH TIME ZONE,
			total_rounds INT DEFAULT 0,
			final_consensus_score DECIMAL(5,4),
			outcome TEXT,
			metadata JSONB
		);

		CREATE INDEX IF NOT EXISTS idx_debate_sessions_debate_id ON debate_sessions(debate_id);
		CREATE INDEX IF NOT EXISTS idx_debate_sessions_status ON debate_sessions(status);
		CREATE INDEX IF NOT EXISTS idx_debate_sessions_created_at ON debate_sessions(created_at);
		CREATE INDEX IF NOT EXISTS idx_debate_sessions_topology_type ON debate_sessions(topology_type);
	`

	_, err := r.pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create debate_sessions table: %w", err)
	}

	r.log.Info("Debate sessions table created/verified")
	return nil
}

// Insert adds a new debate session entry
func (r *DebateSessionRepository) Insert(ctx context.Context, entry *DebateSessionEntry) error {
	now := time.Now()
	entry.CreatedAt = now
	entry.UpdatedAt = now

	query := `
		INSERT INTO debate_sessions (
			id, debate_id, topic, status, topology_type,
			coordination_protocol, config, initiated_by,
			created_at, updated_at, completed_at, total_rounds,
			final_consensus_score, outcome, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id
	`

	err := r.pool.QueryRow(ctx, query,
		entry.ID, entry.DebateID, entry.Topic, entry.Status, entry.TopologyType,
		entry.CoordinationProtocol, entry.Config, entry.InitiatedBy,
		entry.CreatedAt, entry.UpdatedAt, entry.CompletedAt, entry.TotalRounds,
		entry.FinalConsensusScore, entry.Outcome, entry.Metadata,
	).Scan(&entry.ID)

	if err != nil {
		return fmt.Errorf("failed to insert debate session: %w", err)
	}

	r.log.WithFields(logrus.Fields{
		"id":        entry.ID,
		"debate_id": entry.DebateID,
		"topic":     entry.Topic,
		"status":    entry.Status,
	}).Debug("Debate session entry inserted")

	return nil
}

// GetByID retrieves a debate session by its ID
func (r *DebateSessionRepository) GetByID(ctx context.Context, id string) (*DebateSessionEntry, error) {
	query := `
		SELECT id, debate_id, topic, status, topology_type,
			   coordination_protocol, config, initiated_by,
			   created_at, updated_at, completed_at, total_rounds,
			   final_consensus_score, outcome, metadata
		FROM debate_sessions
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	return r.scanRow(row)
}

// GetByDebateID retrieves all sessions for a specific debate
func (r *DebateSessionRepository) GetByDebateID(ctx context.Context, debateID string) ([]*DebateSessionEntry, error) {
	query := `
		SELECT id, debate_id, topic, status, topology_type,
			   coordination_protocol, config, initiated_by,
			   created_at, updated_at, completed_at, total_rounds,
			   final_consensus_score, outcome, metadata
		FROM debate_sessions
		WHERE debate_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, debateID)
	if err != nil {
		return nil, fmt.Errorf("failed to query debate sessions by debate_id: %w", err)
	}
	defer rows.Close()

	return r.scanRows(rows)
}

// UpdateStatus updates the status of a debate session
func (r *DebateSessionRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	query := `
		UPDATE debate_sessions
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	now := time.Now()
	result, err := r.pool.Exec(ctx, query, status, now, id)
	if err != nil {
		return fmt.Errorf("failed to update debate session status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("debate session not found: %s", id)
	}

	r.log.WithFields(logrus.Fields{
		"id":     id,
		"status": status,
	}).Debug("Debate session status updated")

	return nil
}

// UpdateOutcome updates the outcome fields of a completed debate session
func (r *DebateSessionRepository) UpdateOutcome(
	ctx context.Context,
	id string,
	totalRounds int,
	consensusScore float64,
	outcome string,
) error {
	query := `
		UPDATE debate_sessions
		SET total_rounds = $1,
			final_consensus_score = $2,
			outcome = $3,
			status = 'completed',
			completed_at = $4,
			updated_at = $4
		WHERE id = $5
	`

	now := time.Now()
	result, err := r.pool.Exec(ctx, query, totalRounds, consensusScore, outcome, now, id)
	if err != nil {
		return fmt.Errorf("failed to update debate session outcome: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("debate session not found: %s", id)
	}

	r.log.WithFields(logrus.Fields{
		"id":              id,
		"total_rounds":    totalRounds,
		"consensus_score": consensusScore,
		"outcome":         outcome,
	}).Debug("Debate session outcome updated")

	return nil
}

// ListByStatus retrieves all sessions with a specific status
func (r *DebateSessionRepository) ListByStatus(ctx context.Context, status string) ([]*DebateSessionEntry, error) {
	query := `
		SELECT id, debate_id, topic, status, topology_type,
			   coordination_protocol, config, initiated_by,
			   created_at, updated_at, completed_at, total_rounds,
			   final_consensus_score, outcome, metadata
		FROM debate_sessions
		WHERE status = $1
		ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to query debate sessions by status: %w", err)
	}
	defer rows.Close()

	return r.scanRows(rows)
}

// ListActive retrieves all active debate sessions
func (r *DebateSessionRepository) ListActive(ctx context.Context) ([]*DebateSessionEntry, error) {
	query := `
		SELECT id, debate_id, topic, status, topology_type,
			   coordination_protocol, config, initiated_by,
			   created_at, updated_at, completed_at, total_rounds,
			   final_consensus_score, outcome, metadata
		FROM debate_sessions
		WHERE status IN ('pending', 'running', 'paused')
		ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active debate sessions: %w", err)
	}
	defer rows.Close()

	return r.scanRows(rows)
}

// Delete removes a debate session by its ID
func (r *DebateSessionRepository) Delete(ctx context.Context, id string) error {
	query := `
		DELETE FROM debate_sessions
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete debate session: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("debate session not found: %s", id)
	}

	r.log.WithFields(logrus.Fields{
		"id": id,
	}).Debug("Debate session deleted")

	return nil
}

// scanRow scans a single database row into a DebateSessionEntry
func (r *DebateSessionRepository) scanRow(row pgx.Row) (*DebateSessionEntry, error) {
	var entry DebateSessionEntry
	var topologyType, coordinationProtocol, config, initiatedBy *string
	var completedAt *time.Time
	var finalConsensusScore *float64
	var outcome, metadata *string

	err := row.Scan(
		&entry.ID, &entry.DebateID, &entry.Topic, &entry.Status, &topologyType,
		&coordinationProtocol, &config, &initiatedBy,
		&entry.CreatedAt, &entry.UpdatedAt, &completedAt, &entry.TotalRounds,
		&finalConsensusScore, &outcome, &metadata,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan debate session row: %w", err)
	}

	if topologyType != nil {
		entry.TopologyType = *topologyType
	}
	if coordinationProtocol != nil {
		entry.CoordinationProtocol = *coordinationProtocol
	}
	if config != nil {
		entry.Config = *config
	}
	if initiatedBy != nil {
		entry.InitiatedBy = *initiatedBy
	}
	if completedAt != nil {
		entry.CompletedAt = completedAt
	}
	if finalConsensusScore != nil {
		entry.FinalConsensusScore = finalConsensusScore
	}
	if outcome != nil {
		entry.Outcome = *outcome
	}
	if metadata != nil {
		entry.Metadata = *metadata
	}

	return &entry, nil
}

// scanRows scans database rows into a DebateSessionEntry slice
func (r *DebateSessionRepository) scanRows(rows pgx.Rows) ([]*DebateSessionEntry, error) {
	var entries []*DebateSessionEntry

	for rows.Next() {
		var entry DebateSessionEntry
		var topologyType, coordinationProtocol, config, initiatedBy *string
		var completedAt *time.Time
		var finalConsensusScore *float64
		var outcome, metadata *string

		err := rows.Scan(
			&entry.ID, &entry.DebateID, &entry.Topic, &entry.Status, &topologyType,
			&coordinationProtocol, &config, &initiatedBy,
			&entry.CreatedAt, &entry.UpdatedAt, &completedAt, &entry.TotalRounds,
			&finalConsensusScore, &outcome, &metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan debate session row: %w", err)
		}

		if topologyType != nil {
			entry.TopologyType = *topologyType
		}
		if coordinationProtocol != nil {
			entry.CoordinationProtocol = *coordinationProtocol
		}
		if config != nil {
			entry.Config = *config
		}
		if initiatedBy != nil {
			entry.InitiatedBy = *initiatedBy
		}
		if completedAt != nil {
			entry.CompletedAt = completedAt
		}
		if finalConsensusScore != nil {
			entry.FinalConsensusScore = finalConsensusScore
		}
		if outcome != nil {
			entry.Outcome = *outcome
		}
		if metadata != nil {
			entry.Metadata = *metadata
		}

		entries = append(entries, &entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating debate session rows: %w", err)
	}

	return entries, nil
}
