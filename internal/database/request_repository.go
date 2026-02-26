package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

// LLMRequest represents an LLM request in the database
type LLMRequest struct {
	ID             string                 `json:"id"`
	SessionID      *string                `json:"session_id"`
	UserID         *string                `json:"user_id"`
	Prompt         string                 `json:"prompt"`
	Messages       []map[string]string    `json:"messages"`
	ModelParams    map[string]interface{} `json:"model_params"`
	EnsembleConfig map[string]interface{} `json:"ensemble_config"`
	MemoryEnhanced bool                   `json:"memory_enhanced"`
	Memory         map[string]interface{} `json:"memory"`
	Status         string                 `json:"status"`
	RequestType    string                 `json:"request_type"`
	CreatedAt      time.Time              `json:"created_at"`
	StartedAt      *time.Time             `json:"started_at"`
	CompletedAt    *time.Time             `json:"completed_at"`
}

// RequestRepository handles LLM request database operations
type RequestRepository struct {
	pool *pgxpool.Pool
	log  *logrus.Logger
}

// NewRequestRepository creates a new RequestRepository
func NewRequestRepository(pool *pgxpool.Pool, log *logrus.Logger) *RequestRepository {
	return &RequestRepository{
		pool: pool,
		log:  log,
	}
}

// Create creates a new LLM request in the database
func (r *RequestRepository) Create(ctx context.Context, req *LLMRequest) error {
	query := `
		INSERT INTO llm_requests (session_id, user_id, prompt, messages, model_params, ensemble_config, memory_enhanced, memory, status, request_type)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at
	`

	messagesJSON, _ := json.Marshal(req.Messages)
	modelParamsJSON, _ := json.Marshal(req.ModelParams)
	ensembleConfigJSON, _ := json.Marshal(req.EnsembleConfig)
	memoryJSON, _ := json.Marshal(req.Memory)

	err := r.pool.QueryRow(ctx, query,
		req.SessionID, req.UserID, req.Prompt, messagesJSON, modelParamsJSON,
		ensembleConfigJSON, req.MemoryEnhanced, memoryJSON, req.Status, req.RequestType,
	).Scan(&req.ID, &req.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	return nil
}

// GetByID retrieves a request by its ID
func (r *RequestRepository) GetByID(ctx context.Context, id string) (*LLMRequest, error) {
	query := `
		SELECT id, session_id, user_id, prompt, messages, model_params, ensemble_config, memory_enhanced, memory, status, request_type, created_at, started_at, completed_at
		FROM llm_requests
		WHERE id = $1
	`

	req := &LLMRequest{}
	var messagesJSON, modelParamsJSON, ensembleConfigJSON, memoryJSON []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&req.ID, &req.SessionID, &req.UserID, &req.Prompt, &messagesJSON,
		&modelParamsJSON, &ensembleConfigJSON, &req.MemoryEnhanced, &memoryJSON,
		&req.Status, &req.RequestType, &req.CreatedAt, &req.StartedAt, &req.CompletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("request not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get request: %w", err)
	}

	_ = json.Unmarshal(messagesJSON, &req.Messages)             //nolint:errcheck
	_ = json.Unmarshal(modelParamsJSON, &req.ModelParams)       //nolint:errcheck
	_ = json.Unmarshal(ensembleConfigJSON, &req.EnsembleConfig) //nolint:errcheck
	_ = json.Unmarshal(memoryJSON, &req.Memory)                 //nolint:errcheck

	return req, nil
}

// GetBySessionID retrieves all requests for a session
func (r *RequestRepository) GetBySessionID(ctx context.Context, sessionID string, limit, offset int) ([]*LLMRequest, int, error) {
	countQuery := `SELECT COUNT(*) FROM llm_requests WHERE session_id = $1`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, sessionID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count requests: %w", err)
	}

	query := `
		SELECT id, session_id, user_id, prompt, messages, model_params, ensemble_config, memory_enhanced, memory, status, request_type, created_at, started_at, completed_at
		FROM llm_requests
		WHERE session_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, sessionID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list requests: %w", err)
	}
	defer rows.Close()

	requests := []*LLMRequest{}
	for rows.Next() {
		req := &LLMRequest{}
		var messagesJSON, modelParamsJSON, ensembleConfigJSON, memoryJSON []byte

		err := rows.Scan(
			&req.ID, &req.SessionID, &req.UserID, &req.Prompt, &messagesJSON,
			&modelParamsJSON, &ensembleConfigJSON, &req.MemoryEnhanced, &memoryJSON,
			&req.Status, &req.RequestType, &req.CreatedAt, &req.StartedAt, &req.CompletedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan request row: %w", err)
		}

		_ = json.Unmarshal(messagesJSON, &req.Messages)             //nolint:errcheck
		_ = json.Unmarshal(modelParamsJSON, &req.ModelParams)       //nolint:errcheck
		_ = json.Unmarshal(ensembleConfigJSON, &req.EnsembleConfig) //nolint:errcheck
		_ = json.Unmarshal(memoryJSON, &req.Memory)                 //nolint:errcheck

		requests = append(requests, req)
	}

	return requests, total, nil
}

// GetByUserID retrieves all requests for a user
func (r *RequestRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*LLMRequest, int, error) {
	countQuery := `SELECT COUNT(*) FROM llm_requests WHERE user_id = $1`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count requests: %w", err)
	}

	query := `
		SELECT id, session_id, user_id, prompt, messages, model_params, ensemble_config, memory_enhanced, memory, status, request_type, created_at, started_at, completed_at
		FROM llm_requests
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list requests: %w", err)
	}
	defer rows.Close()

	requests := []*LLMRequest{}
	for rows.Next() {
		req := &LLMRequest{}
		var messagesJSON, modelParamsJSON, ensembleConfigJSON, memoryJSON []byte

		err := rows.Scan(
			&req.ID, &req.SessionID, &req.UserID, &req.Prompt, &messagesJSON,
			&modelParamsJSON, &ensembleConfigJSON, &req.MemoryEnhanced, &memoryJSON,
			&req.Status, &req.RequestType, &req.CreatedAt, &req.StartedAt, &req.CompletedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan request row: %w", err)
		}

		_ = json.Unmarshal(messagesJSON, &req.Messages)             //nolint:errcheck
		_ = json.Unmarshal(modelParamsJSON, &req.ModelParams)       //nolint:errcheck
		_ = json.Unmarshal(ensembleConfigJSON, &req.EnsembleConfig) //nolint:errcheck
		_ = json.Unmarshal(memoryJSON, &req.Memory)                 //nolint:errcheck

		requests = append(requests, req)
	}

	return requests, total, nil
}

// UpdateStatus updates the status and timing information for a request
func (r *RequestRepository) UpdateStatus(ctx context.Context, id, status string) error {
	var query string
	switch status {
	case "processing":
		query = `UPDATE llm_requests SET status = $2, started_at = NOW() WHERE id = $1`
	case "completed", "failed":
		query = `UPDATE llm_requests SET status = $2, completed_at = NOW() WHERE id = $1`
	default:
		query = `UPDATE llm_requests SET status = $2 WHERE id = $1`
	}

	result, err := r.pool.Exec(ctx, query, id, status)
	if err != nil {
		return fmt.Errorf("failed to update request status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("request not found: %s", id)
	}

	return nil
}

// Delete deletes a request by its ID
func (r *RequestRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM llm_requests WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete request: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("request not found: %s", id)
	}

	return nil
}

// GetPendingRequests retrieves all pending requests
func (r *RequestRepository) GetPendingRequests(ctx context.Context, limit int) ([]*LLMRequest, error) {
	query := `
		SELECT id, session_id, user_id, prompt, messages, model_params, ensemble_config, memory_enhanced, memory, status, request_type, created_at, started_at, completed_at
		FROM llm_requests
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT $1
	`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending requests: %w", err)
	}
	defer rows.Close()

	requests := []*LLMRequest{}
	for rows.Next() {
		req := &LLMRequest{}
		var messagesJSON, modelParamsJSON, ensembleConfigJSON, memoryJSON []byte

		err := rows.Scan(
			&req.ID, &req.SessionID, &req.UserID, &req.Prompt, &messagesJSON,
			&modelParamsJSON, &ensembleConfigJSON, &req.MemoryEnhanced, &memoryJSON,
			&req.Status, &req.RequestType, &req.CreatedAt, &req.StartedAt, &req.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan request row: %w", err)
		}

		_ = json.Unmarshal(messagesJSON, &req.Messages)             //nolint:errcheck
		_ = json.Unmarshal(modelParamsJSON, &req.ModelParams)       //nolint:errcheck
		_ = json.Unmarshal(ensembleConfigJSON, &req.EnsembleConfig) //nolint:errcheck
		_ = json.Unmarshal(memoryJSON, &req.Memory)                 //nolint:errcheck

		requests = append(requests, req)
	}

	return requests, nil
}

// GetRequestStats retrieves statistics about requests
func (r *RequestRepository) GetRequestStats(ctx context.Context, userID string, since time.Time) (map[string]int, error) {
	query := `
		SELECT status, COUNT(*) as count
		FROM llm_requests
		WHERE ($1 = '' OR user_id = $1) AND created_at >= $2
		GROUP BY status
	`

	rows, err := r.pool.Query(ctx, query, userID, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get request stats: %w", err)
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("failed to scan stats row: %w", err)
		}
		stats[status] = count
	}

	return stats, nil
}

// FindByUserID retrieves all requests for a user (implements LLMRequestRepository interface)
func (r *RequestRepository) FindByUserID(ctx context.Context, userID string, limit, offset int) ([]*LLMRequest, error) {
	requests, _, err := r.GetByUserID(ctx, userID, limit, offset)
	return requests, err
}
