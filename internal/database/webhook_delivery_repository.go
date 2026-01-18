package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WebhookDeliveryStatus represents the status of a webhook delivery
type WebhookDeliveryStatus string

const (
	WebhookStatusPending   WebhookDeliveryStatus = "pending"
	WebhookStatusDelivered WebhookDeliveryStatus = "delivered"
	WebhookStatusFailed    WebhookDeliveryStatus = "failed"
	WebhookStatusRetrying  WebhookDeliveryStatus = "retrying"
)

// WebhookDelivery represents a webhook delivery record
type WebhookDelivery struct {
	ID            string                `db:"id" json:"id"`
	TaskID        *string               `db:"task_id" json:"task_id,omitempty"`
	WebhookURL    string                `db:"webhook_url" json:"webhook_url"`
	EventType     string                `db:"event_type" json:"event_type"`
	Payload       json.RawMessage       `db:"payload" json:"payload"`
	Status        WebhookDeliveryStatus `db:"status" json:"status"`
	Attempts      int                   `db:"attempts" json:"attempts"`
	LastAttemptAt *time.Time            `db:"last_attempt_at" json:"last_attempt_at,omitempty"`
	LastError     *string               `db:"last_error" json:"last_error,omitempty"`
	ResponseCode  *int                  `db:"response_code" json:"response_code,omitempty"`
	CreatedAt     time.Time             `db:"created_at" json:"created_at"`
	DeliveredAt   *time.Time            `db:"delivered_at" json:"delivered_at,omitempty"`
}

// WebhookDeliveryFilter contains filter options for listing webhook deliveries
type WebhookDeliveryFilter struct {
	TaskID    string
	Status    WebhookDeliveryStatus
	EventType string
	Limit     int
	Offset    int
}

// WebhookDeliveryRepository handles webhook delivery database operations
type WebhookDeliveryRepository struct {
	pool *pgxpool.Pool
}

// NewWebhookDeliveryRepository creates a new webhook delivery repository
func NewWebhookDeliveryRepository(pool *pgxpool.Pool) *WebhookDeliveryRepository {
	return &WebhookDeliveryRepository{pool: pool}
}

// Create creates a new webhook delivery record
func (r *WebhookDeliveryRepository) Create(ctx context.Context, delivery *WebhookDelivery) error {
	if delivery.ID == "" {
		delivery.ID = uuid.New().String()
	}
	if delivery.Status == "" {
		delivery.Status = WebhookStatusPending
	}

	query := `
		INSERT INTO webhook_deliveries (id, task_id, webhook_url, event_type, payload, status, attempts, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	now := time.Now()
	_, err := r.pool.Exec(ctx, query,
		delivery.ID, delivery.TaskID, delivery.WebhookURL, delivery.EventType,
		delivery.Payload, delivery.Status, delivery.Attempts, now)
	if err != nil {
		return fmt.Errorf("failed to create webhook delivery: %w", err)
	}

	delivery.CreatedAt = now
	return nil
}

// GetByID retrieves a webhook delivery by ID
func (r *WebhookDeliveryRepository) GetByID(ctx context.Context, id string) (*WebhookDelivery, error) {
	query := `
		SELECT id, task_id, webhook_url, event_type, payload, status, attempts,
		       last_attempt_at, last_error, response_code, created_at, delivered_at
		FROM webhook_deliveries WHERE id = $1
	`
	var delivery WebhookDelivery
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&delivery.ID, &delivery.TaskID, &delivery.WebhookURL, &delivery.EventType,
		&delivery.Payload, &delivery.Status, &delivery.Attempts,
		&delivery.LastAttemptAt, &delivery.LastError, &delivery.ResponseCode,
		&delivery.CreatedAt, &delivery.DeliveredAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook delivery: %w", err)
	}
	return &delivery, nil
}

// Update updates a webhook delivery record
func (r *WebhookDeliveryRepository) Update(ctx context.Context, delivery *WebhookDelivery) error {
	query := `
		UPDATE webhook_deliveries
		SET status = $2, attempts = $3, last_attempt_at = $4, last_error = $5,
		    response_code = $6, delivered_at = $7
		WHERE id = $1
	`
	result, err := r.pool.Exec(ctx, query,
		delivery.ID, delivery.Status, delivery.Attempts, delivery.LastAttemptAt,
		delivery.LastError, delivery.ResponseCode, delivery.DeliveredAt)
	if err != nil {
		return fmt.Errorf("failed to update webhook delivery: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("webhook delivery not found: %s", delivery.ID)
	}
	return nil
}

// Delete removes a webhook delivery by ID
func (r *WebhookDeliveryRepository) Delete(ctx context.Context, id string) error {
	result, err := r.pool.Exec(ctx, "DELETE FROM webhook_deliveries WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete webhook delivery: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("webhook delivery not found: %s", id)
	}
	return nil
}

// List retrieves webhook deliveries with optional filtering
func (r *WebhookDeliveryRepository) List(ctx context.Context, filter WebhookDeliveryFilter) ([]*WebhookDelivery, error) {
	query := `
		SELECT id, task_id, webhook_url, event_type, payload, status, attempts,
		       last_attempt_at, last_error, response_code, created_at, delivered_at
		FROM webhook_deliveries WHERE 1=1
	`
	args := []interface{}{}
	argNum := 1

	if filter.TaskID != "" {
		query += fmt.Sprintf(" AND task_id = $%d", argNum)
		args = append(args, filter.TaskID)
		argNum++
	}

	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argNum)
		args = append(args, filter.Status)
		argNum++
	}

	if filter.EventType != "" {
		query += fmt.Sprintf(" AND event_type = $%d", argNum)
		args = append(args, filter.EventType)
		argNum++
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argNum)
		args = append(args, filter.Limit)
		argNum++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argNum)
		args = append(args, filter.Offset)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list webhook deliveries: %w", err)
	}
	defer rows.Close()

	var deliveries []*WebhookDelivery
	for rows.Next() {
		var d WebhookDelivery
		if err := rows.Scan(
			&d.ID, &d.TaskID, &d.WebhookURL, &d.EventType, &d.Payload, &d.Status,
			&d.Attempts, &d.LastAttemptAt, &d.LastError, &d.ResponseCode,
			&d.CreatedAt, &d.DeliveredAt); err != nil {
			return nil, fmt.Errorf("failed to scan webhook delivery: %w", err)
		}
		deliveries = append(deliveries, &d)
	}

	return deliveries, rows.Err()
}

// GetPendingDeliveries retrieves pending webhook deliveries ready for retry
func (r *WebhookDeliveryRepository) GetPendingDeliveries(ctx context.Context, maxAttempts int, limit int) ([]*WebhookDelivery, error) {
	query := `
		SELECT id, task_id, webhook_url, event_type, payload, status, attempts,
		       last_attempt_at, last_error, response_code, created_at, delivered_at
		FROM webhook_deliveries
		WHERE status IN ('pending', 'retrying') AND attempts < $1
		ORDER BY created_at ASC
		LIMIT $2
	`
	rows, err := r.pool.Query(ctx, query, maxAttempts, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending deliveries: %w", err)
	}
	defer rows.Close()

	var deliveries []*WebhookDelivery
	for rows.Next() {
		var d WebhookDelivery
		if err := rows.Scan(
			&d.ID, &d.TaskID, &d.WebhookURL, &d.EventType, &d.Payload, &d.Status,
			&d.Attempts, &d.LastAttemptAt, &d.LastError, &d.ResponseCode,
			&d.CreatedAt, &d.DeliveredAt); err != nil {
			return nil, fmt.Errorf("failed to scan webhook delivery: %w", err)
		}
		deliveries = append(deliveries, &d)
	}

	return deliveries, rows.Err()
}

// MarkDelivered marks a webhook delivery as successfully delivered
func (r *WebhookDeliveryRepository) MarkDelivered(ctx context.Context, id string, responseCode int) error {
	now := time.Now()
	query := `
		UPDATE webhook_deliveries
		SET status = 'delivered', last_attempt_at = $2, response_code = $3, delivered_at = $4, attempts = attempts + 1
		WHERE id = $1
	`
	result, err := r.pool.Exec(ctx, query, id, now, responseCode, now)
	if err != nil {
		return fmt.Errorf("failed to mark as delivered: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("webhook delivery not found: %s", id)
	}
	return nil
}

// MarkFailed marks a webhook delivery as failed with an error message
func (r *WebhookDeliveryRepository) MarkFailed(ctx context.Context, id string, errorMsg string, responseCode *int) error {
	now := time.Now()
	query := `
		UPDATE webhook_deliveries
		SET status = 'failed', last_attempt_at = $2, last_error = $3, response_code = $4, attempts = attempts + 1
		WHERE id = $1
	`
	result, err := r.pool.Exec(ctx, query, id, now, errorMsg, responseCode)
	if err != nil {
		return fmt.Errorf("failed to mark as failed: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("webhook delivery not found: %s", id)
	}
	return nil
}

// MarkRetrying marks a webhook delivery for retry
func (r *WebhookDeliveryRepository) MarkRetrying(ctx context.Context, id string, errorMsg string, responseCode *int) error {
	now := time.Now()
	query := `
		UPDATE webhook_deliveries
		SET status = 'retrying', last_attempt_at = $2, last_error = $3, response_code = $4, attempts = attempts + 1
		WHERE id = $1
	`
	result, err := r.pool.Exec(ctx, query, id, now, errorMsg, responseCode)
	if err != nil {
		return fmt.Errorf("failed to mark for retry: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("webhook delivery not found: %s", id)
	}
	return nil
}

// Count returns the total number of webhook deliveries
func (r *WebhookDeliveryRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM webhook_deliveries").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count webhook deliveries: %w", err)
	}
	return count, nil
}

// CountByStatus returns the count of deliveries per status
func (r *WebhookDeliveryRepository) CountByStatus(ctx context.Context) (map[WebhookDeliveryStatus]int64, error) {
	query := `
		SELECT status, COUNT(*) as count
		FROM webhook_deliveries
		GROUP BY status
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to count by status: %w", err)
	}
	defer rows.Close()

	counts := make(map[WebhookDeliveryStatus]int64)
	for rows.Next() {
		var status WebhookDeliveryStatus
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("failed to scan status count: %w", err)
		}
		counts[status] = count
	}

	return counts, rows.Err()
}

// GetByTaskID retrieves all webhook deliveries for a specific task
func (r *WebhookDeliveryRepository) GetByTaskID(ctx context.Context, taskID string) ([]*WebhookDelivery, error) {
	query := `
		SELECT id, task_id, webhook_url, event_type, payload, status, attempts,
		       last_attempt_at, last_error, response_code, created_at, delivered_at
		FROM webhook_deliveries
		WHERE task_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deliveries by task ID: %w", err)
	}
	defer rows.Close()

	var deliveries []*WebhookDelivery
	for rows.Next() {
		var d WebhookDelivery
		if err := rows.Scan(
			&d.ID, &d.TaskID, &d.WebhookURL, &d.EventType, &d.Payload, &d.Status,
			&d.Attempts, &d.LastAttemptAt, &d.LastError, &d.ResponseCode,
			&d.CreatedAt, &d.DeliveredAt); err != nil {
			return nil, fmt.Errorf("failed to scan webhook delivery: %w", err)
		}
		deliveries = append(deliveries, &d)
	}

	return deliveries, rows.Err()
}

// DeleteOldDeliveries deletes delivered webhook records older than the specified time
func (r *WebhookDeliveryRepository) DeleteOldDeliveries(ctx context.Context, cutoff time.Time) (int64, error) {
	result, err := r.pool.Exec(ctx,
		"DELETE FROM webhook_deliveries WHERE status = 'delivered' AND delivered_at < $1", cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old deliveries: %w", err)
	}
	return result.RowsAffected(), nil
}

// GetDeliveryStats returns delivery statistics
func (r *WebhookDeliveryRepository) GetDeliveryStats(ctx context.Context, since time.Time) (map[string]interface{}, error) {
	query := `
		SELECT
			COUNT(*) as total,
			COUNT(CASE WHEN status = 'delivered' THEN 1 END) as delivered,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed,
			COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending,
			COUNT(CASE WHEN status = 'retrying' THEN 1 END) as retrying,
			AVG(attempts) as avg_attempts,
			MAX(attempts) as max_attempts
		FROM webhook_deliveries
		WHERE created_at >= $1
	`
	var total, delivered, failed, pending, retrying int64
	var avgAttempts, maxAttempts *float64

	err := r.pool.QueryRow(ctx, query, since).Scan(
		&total, &delivered, &failed, &pending, &retrying,
		&avgAttempts, &maxAttempts)
	if err != nil {
		return nil, fmt.Errorf("failed to get delivery stats: %w", err)
	}

	successRate := float64(0)
	if total > 0 {
		successRate = float64(delivered) / float64(total) * 100
	}

	return map[string]interface{}{
		"total":        total,
		"delivered":    delivered,
		"failed":       failed,
		"pending":      pending,
		"retrying":     retrying,
		"avg_attempts": avgAttempts,
		"max_attempts": maxAttempts,
		"success_rate": successRate,
	}, nil
}
