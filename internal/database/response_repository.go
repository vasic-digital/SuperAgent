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

// LLMResponse represents an LLM response in the database
type LLMResponse struct {
	ID             string                 `json:"id"`
	RequestID      string                 `json:"request_id"`
	ProviderID     *string                `json:"provider_id"`
	ProviderName   string                 `json:"provider_name"`
	Content        string                 `json:"content"`
	Confidence     float64                `json:"confidence"`
	TokensUsed     int                    `json:"tokens_used"`
	ResponseTime   int64                  `json:"response_time"`
	FinishReason   string                 `json:"finish_reason"`
	Metadata       map[string]interface{} `json:"metadata"`
	Selected       bool                   `json:"selected"`
	SelectionScore float64                `json:"selection_score"`
	CreatedAt      time.Time              `json:"created_at"`
}

// ResponseRepository handles LLM response database operations
type ResponseRepository struct {
	pool *pgxpool.Pool
	log  *logrus.Logger
}

// NewResponseRepository creates a new ResponseRepository
func NewResponseRepository(pool *pgxpool.Pool, log *logrus.Logger) *ResponseRepository {
	return &ResponseRepository{
		pool: pool,
		log:  log,
	}
}

// Create creates a new LLM response in the database
func (r *ResponseRepository) Create(ctx context.Context, resp *LLMResponse) error {
	query := `
		INSERT INTO llm_responses (request_id, provider_id, provider_name, content, confidence, tokens_used, response_time, finish_reason, metadata, selected, selection_score)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at
	`

	metadataJSON, _ := json.Marshal(resp.Metadata)

	err := r.pool.QueryRow(ctx, query,
		resp.RequestID, resp.ProviderID, resp.ProviderName, resp.Content, resp.Confidence,
		resp.TokensUsed, resp.ResponseTime, resp.FinishReason, metadataJSON,
		resp.Selected, resp.SelectionScore,
	).Scan(&resp.ID, &resp.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create response: %w", err)
	}

	return nil
}

// GetByID retrieves a response by its ID
func (r *ResponseRepository) GetByID(ctx context.Context, id string) (*LLMResponse, error) {
	query := `
		SELECT id, request_id, provider_id, provider_name, content, confidence, tokens_used, response_time, finish_reason, metadata, selected, selection_score, created_at
		FROM llm_responses
		WHERE id = $1
	`

	resp := &LLMResponse{}
	var metadataJSON []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&resp.ID, &resp.RequestID, &resp.ProviderID, &resp.ProviderName, &resp.Content,
		&resp.Confidence, &resp.TokensUsed, &resp.ResponseTime, &resp.FinishReason,
		&metadataJSON, &resp.Selected, &resp.SelectionScore, &resp.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("response not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get response: %w", err)
	}

	if len(metadataJSON) > 0 {
		_ = json.Unmarshal(metadataJSON, &resp.Metadata) //nolint:errcheck
	}

	return resp, nil
}

// GetByRequestID retrieves all responses for a request
func (r *ResponseRepository) GetByRequestID(ctx context.Context, requestID string) ([]*LLMResponse, error) {
	query := `
		SELECT id, request_id, provider_id, provider_name, content, confidence, tokens_used, response_time, finish_reason, metadata, selected, selection_score, created_at
		FROM llm_responses
		WHERE request_id = $1
		ORDER BY selection_score DESC, confidence DESC
	`

	rows, err := r.pool.Query(ctx, query, requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to list responses: %w", err)
	}
	defer rows.Close()

	responses := []*LLMResponse{}
	for rows.Next() {
		resp := &LLMResponse{}
		var metadataJSON []byte

		err := rows.Scan(
			&resp.ID, &resp.RequestID, &resp.ProviderID, &resp.ProviderName, &resp.Content,
			&resp.Confidence, &resp.TokensUsed, &resp.ResponseTime, &resp.FinishReason,
			&metadataJSON, &resp.Selected, &resp.SelectionScore, &resp.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan response row: %w", err)
		}

		if len(metadataJSON) > 0 {
			_ = json.Unmarshal(metadataJSON, &resp.Metadata) //nolint:errcheck
		}

		responses = append(responses, resp)
	}

	return responses, nil
}

// GetSelectedResponse retrieves the selected response for a request
func (r *ResponseRepository) GetSelectedResponse(ctx context.Context, requestID string) (*LLMResponse, error) {
	query := `
		SELECT id, request_id, provider_id, provider_name, content, confidence, tokens_used, response_time, finish_reason, metadata, selected, selection_score, created_at
		FROM llm_responses
		WHERE request_id = $1 AND selected = true
		LIMIT 1
	`

	resp := &LLMResponse{}
	var metadataJSON []byte

	err := r.pool.QueryRow(ctx, query, requestID).Scan(
		&resp.ID, &resp.RequestID, &resp.ProviderID, &resp.ProviderName, &resp.Content,
		&resp.Confidence, &resp.TokensUsed, &resp.ResponseTime, &resp.FinishReason,
		&metadataJSON, &resp.Selected, &resp.SelectionScore, &resp.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("no selected response found for request: %s", requestID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get selected response: %w", err)
	}

	if len(metadataJSON) > 0 {
		_ = json.Unmarshal(metadataJSON, &resp.Metadata) //nolint:errcheck
	}

	return resp, nil
}

// SetSelected marks a response as selected and unselects others for the same request
func (r *ResponseRepository) SetSelected(ctx context.Context, id string, selectionScore float64) error {
	// First, get the request_id for this response
	var requestID string
	err := r.pool.QueryRow(ctx, `SELECT request_id FROM llm_responses WHERE id = $1`, id).Scan(&requestID)
	if err == pgx.ErrNoRows {
		return fmt.Errorf("response not found: %s", id)
	}
	if err != nil {
		return fmt.Errorf("failed to get request_id: %w", err)
	}

	// Start a transaction
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }() //nolint:errcheck

	// Unselect all responses for this request
	_, err = tx.Exec(ctx, `UPDATE llm_responses SET selected = false WHERE request_id = $1`, requestID)
	if err != nil {
		return fmt.Errorf("failed to unselect responses: %w", err)
	}

	// Select the specified response
	_, err = tx.Exec(ctx, `UPDATE llm_responses SET selected = true, selection_score = $2 WHERE id = $1`, id, selectionScore)
	if err != nil {
		return fmt.Errorf("failed to select response: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Delete deletes a response by its ID
func (r *ResponseRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM llm_responses WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete response: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("response not found: %s", id)
	}

	return nil
}

// DeleteByRequestID deletes all responses for a request
func (r *ResponseRepository) DeleteByRequestID(ctx context.Context, requestID string) (int64, error) {
	query := `DELETE FROM llm_responses WHERE request_id = $1`

	result, err := r.pool.Exec(ctx, query, requestID)
	if err != nil {
		return 0, fmt.Errorf("failed to delete responses: %w", err)
	}

	return result.RowsAffected(), nil
}

// GetByProviderID retrieves all responses from a specific provider
func (r *ResponseRepository) GetByProviderID(ctx context.Context, providerID string, limit, offset int) ([]*LLMResponse, int, error) {
	countQuery := `SELECT COUNT(*) FROM llm_responses WHERE provider_id = $1`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, providerID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count responses: %w", err)
	}

	query := `
		SELECT id, request_id, provider_id, provider_name, content, confidence, tokens_used, response_time, finish_reason, metadata, selected, selection_score, created_at
		FROM llm_responses
		WHERE provider_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, providerID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list responses: %w", err)
	}
	defer rows.Close()

	responses := []*LLMResponse{}
	for rows.Next() {
		resp := &LLMResponse{}
		var metadataJSON []byte

		err := rows.Scan(
			&resp.ID, &resp.RequestID, &resp.ProviderID, &resp.ProviderName, &resp.Content,
			&resp.Confidence, &resp.TokensUsed, &resp.ResponseTime, &resp.FinishReason,
			&metadataJSON, &resp.Selected, &resp.SelectionScore, &resp.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan response row: %w", err)
		}

		if len(metadataJSON) > 0 {
			_ = json.Unmarshal(metadataJSON, &resp.Metadata)
		}

		responses = append(responses, resp)
	}

	return responses, total, nil
}

// GetProviderStats retrieves statistics for providers
func (r *ResponseRepository) GetProviderStats(ctx context.Context, since time.Time) ([]map[string]interface{}, error) {
	query := `
		SELECT
			provider_name,
			COUNT(*) as total_responses,
			COUNT(CASE WHEN selected = true THEN 1 END) as selected_count,
			AVG(confidence) as avg_confidence,
			AVG(response_time) as avg_response_time,
			SUM(tokens_used) as total_tokens
		FROM llm_responses
		WHERE created_at >= $1
		GROUP BY provider_name
		ORDER BY selected_count DESC
	`

	rows, err := r.pool.Query(ctx, query, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider stats: %w", err)
	}
	defer rows.Close()

	stats := []map[string]interface{}{}
	for rows.Next() {
		var providerName string
		var totalResponses, selectedCount, totalTokens int
		var avgConfidence, avgResponseTime float64

		if err := rows.Scan(&providerName, &totalResponses, &selectedCount, &avgConfidence, &avgResponseTime, &totalTokens); err != nil {
			return nil, fmt.Errorf("failed to scan stats row: %w", err)
		}

		stats = append(stats, map[string]interface{}{
			"provider_name":     providerName,
			"total_responses":   totalResponses,
			"selected_count":    selectedCount,
			"avg_confidence":    avgConfidence,
			"avg_response_time": avgResponseTime,
			"total_tokens":      totalTokens,
		})
	}

	return stats, nil
}
