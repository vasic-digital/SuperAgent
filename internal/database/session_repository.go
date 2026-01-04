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

// UserSession represents a user session in the system
type UserSession struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id"`
	SessionToken string                 `json:"session_token"`
	Context      map[string]interface{} `json:"context"`
	MemoryID     *string                `json:"memory_id"`
	Status       string                 `json:"status"`
	RequestCount int                    `json:"request_count"`
	LastActivity time.Time              `json:"last_activity"`
	ExpiresAt    time.Time              `json:"expires_at"`
	CreatedAt    time.Time              `json:"created_at"`
}

// SessionRepository handles session database operations
type SessionRepository struct {
	pool *pgxpool.Pool
	log  *logrus.Logger
}

// NewSessionRepository creates a new SessionRepository
func NewSessionRepository(pool *pgxpool.Pool, log *logrus.Logger) *SessionRepository {
	return &SessionRepository{
		pool: pool,
		log:  log,
	}
}

// Create creates a new session in the database
func (r *SessionRepository) Create(ctx context.Context, session *UserSession) error {
	query := `
		INSERT INTO user_sessions (user_id, session_token, context, memory_id, status, request_count, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, last_activity, created_at
	`

	contextJSON, err := json.Marshal(session.Context)
	if err != nil {
		return fmt.Errorf("failed to marshal session context: %w", err)
	}

	err = r.pool.QueryRow(ctx, query,
		session.UserID, session.SessionToken, contextJSON, session.MemoryID,
		session.Status, session.RequestCount, session.ExpiresAt,
	).Scan(&session.ID, &session.LastActivity, &session.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// GetByID retrieves a session by its ID
func (r *SessionRepository) GetByID(ctx context.Context, id string) (*UserSession, error) {
	query := `
		SELECT id, user_id, session_token, context, memory_id, status, request_count, last_activity, expires_at, created_at
		FROM user_sessions
		WHERE id = $1
	`

	session := &UserSession{}
	var contextJSON []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&session.ID, &session.UserID, &session.SessionToken, &contextJSON,
		&session.MemoryID, &session.Status, &session.RequestCount,
		&session.LastActivity, &session.ExpiresAt, &session.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("session not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if len(contextJSON) > 0 {
		if err := json.Unmarshal(contextJSON, &session.Context); err != nil {
			session.Context = make(map[string]interface{})
		}
	}

	return session, nil
}

// GetByToken retrieves a session by its session token
func (r *SessionRepository) GetByToken(ctx context.Context, token string) (*UserSession, error) {
	query := `
		SELECT id, user_id, session_token, context, memory_id, status, request_count, last_activity, expires_at, created_at
		FROM user_sessions
		WHERE session_token = $1
	`

	session := &UserSession{}
	var contextJSON []byte

	err := r.pool.QueryRow(ctx, query, token).Scan(
		&session.ID, &session.UserID, &session.SessionToken, &contextJSON,
		&session.MemoryID, &session.Status, &session.RequestCount,
		&session.LastActivity, &session.ExpiresAt, &session.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("session not found for token")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session by token: %w", err)
	}

	if len(contextJSON) > 0 {
		if err := json.Unmarshal(contextJSON, &session.Context); err != nil {
			session.Context = make(map[string]interface{})
		}
	}

	return session, nil
}

// GetByUserID retrieves all sessions for a user
func (r *SessionRepository) GetByUserID(ctx context.Context, userID string) ([]*UserSession, error) {
	query := `
		SELECT id, user_id, session_token, context, memory_id, status, request_count, last_activity, expires_at, created_at
		FROM user_sessions
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions by user: %w", err)
	}
	defer rows.Close()

	sessions := []*UserSession{}
	for rows.Next() {
		session := &UserSession{}
		var contextJSON []byte

		err := rows.Scan(
			&session.ID, &session.UserID, &session.SessionToken, &contextJSON,
			&session.MemoryID, &session.Status, &session.RequestCount,
			&session.LastActivity, &session.ExpiresAt, &session.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session row: %w", err)
		}

		if len(contextJSON) > 0 {
			if err := json.Unmarshal(contextJSON, &session.Context); err != nil {
				session.Context = make(map[string]interface{})
			}
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

// GetActiveSessions retrieves all active (non-expired) sessions for a user
func (r *SessionRepository) GetActiveSessions(ctx context.Context, userID string) ([]*UserSession, error) {
	query := `
		SELECT id, user_id, session_token, context, memory_id, status, request_count, last_activity, expires_at, created_at
		FROM user_sessions
		WHERE user_id = $1 AND status = 'active' AND expires_at > NOW()
		ORDER BY last_activity DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active sessions: %w", err)
	}
	defer rows.Close()

	sessions := []*UserSession{}
	for rows.Next() {
		session := &UserSession{}
		var contextJSON []byte

		err := rows.Scan(
			&session.ID, &session.UserID, &session.SessionToken, &contextJSON,
			&session.MemoryID, &session.Status, &session.RequestCount,
			&session.LastActivity, &session.ExpiresAt, &session.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session row: %w", err)
		}

		if len(contextJSON) > 0 {
			if err := json.Unmarshal(contextJSON, &session.Context); err != nil {
				session.Context = make(map[string]interface{})
			}
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

// Update updates an existing session
func (r *SessionRepository) Update(ctx context.Context, session *UserSession) error {
	query := `
		UPDATE user_sessions
		SET context = $2, memory_id = $3, status = $4, request_count = $5, last_activity = NOW(), expires_at = $6
		WHERE id = $1
		RETURNING last_activity
	`

	contextJSON, err := json.Marshal(session.Context)
	if err != nil {
		return fmt.Errorf("failed to marshal session context: %w", err)
	}

	err = r.pool.QueryRow(ctx, query,
		session.ID, contextJSON, session.MemoryID, session.Status, session.RequestCount, session.ExpiresAt,
	).Scan(&session.LastActivity)

	if err == pgx.ErrNoRows {
		return fmt.Errorf("session not found: %s", session.ID)
	}
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

// UpdateActivity updates the last activity time and request count for a session
func (r *SessionRepository) UpdateActivity(ctx context.Context, id string) error {
	query := `
		UPDATE user_sessions
		SET last_activity = NOW(), request_count = request_count + 1
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to update session activity: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("session not found: %s", id)
	}

	return nil
}

// UpdateContext updates only the context for a session
func (r *SessionRepository) UpdateContext(ctx context.Context, id string, sessionContext map[string]interface{}) error {
	query := `
		UPDATE user_sessions
		SET context = $2, last_activity = NOW()
		WHERE id = $1
	`

	contextJSON, err := json.Marshal(sessionContext)
	if err != nil {
		return fmt.Errorf("failed to marshal session context: %w", err)
	}

	result, err := r.pool.Exec(ctx, query, id, contextJSON)
	if err != nil {
		return fmt.Errorf("failed to update session context: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("session not found: %s", id)
	}

	return nil
}

// Terminate marks a session as terminated
func (r *SessionRepository) Terminate(ctx context.Context, id string) error {
	query := `
		UPDATE user_sessions
		SET status = 'terminated', last_activity = NOW()
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to terminate session: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("session not found: %s", id)
	}

	return nil
}

// Delete deletes a session by its ID
func (r *SessionRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM user_sessions WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("session not found: %s", id)
	}

	return nil
}

// DeleteExpired deletes all expired sessions
func (r *SessionRepository) DeleteExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM user_sessions WHERE expires_at < NOW()`

	result, err := r.pool.Exec(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	return result.RowsAffected(), nil
}

// DeleteByUserID deletes all sessions for a user
func (r *SessionRepository) DeleteByUserID(ctx context.Context, userID string) (int64, error) {
	query := `DELETE FROM user_sessions WHERE user_id = $1`

	result, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to delete user sessions: %w", err)
	}

	return result.RowsAffected(), nil
}

// IsValid checks if a session is valid (exists, active, and not expired)
func (r *SessionRepository) IsValid(ctx context.Context, token string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM user_sessions
			WHERE session_token = $1 AND status = 'active' AND expires_at > NOW()
		)
	`

	var valid bool
	if err := r.pool.QueryRow(ctx, query, token).Scan(&valid); err != nil {
		return false, fmt.Errorf("failed to check session validity: %w", err)
	}

	return valid, nil
}

// ExtendExpiration extends the expiration time of a session
func (r *SessionRepository) ExtendExpiration(ctx context.Context, id string, newExpiry time.Time) error {
	query := `
		UPDATE user_sessions
		SET expires_at = $2, last_activity = NOW()
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, id, newExpiry)
	if err != nil {
		return fmt.Errorf("failed to extend session expiration: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("session not found: %s", id)
	}

	return nil
}
