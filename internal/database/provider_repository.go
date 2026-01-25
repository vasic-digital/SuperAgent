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

// LLMProvider represents an LLM provider configuration in the database
type LLMProvider struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	APIKey       string                 `json:"api_key,omitempty"`
	BaseURL      string                 `json:"base_url"`
	Model        string                 `json:"model"`
	Weight       float64                `json:"weight"`
	Enabled      bool                   `json:"enabled"`
	Config       map[string]interface{} `json:"config"`
	HealthStatus string                 `json:"health_status"`
	ResponseTime int64                  `json:"response_time"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// ProviderRepository handles LLM provider database operations
type ProviderRepository struct {
	pool *pgxpool.Pool
	log  *logrus.Logger
}

// NewProviderRepository creates a new ProviderRepository
func NewProviderRepository(pool *pgxpool.Pool, log *logrus.Logger) *ProviderRepository {
	return &ProviderRepository{
		pool: pool,
		log:  log,
	}
}

// Create creates a new LLM provider in the database
func (r *ProviderRepository) Create(ctx context.Context, provider *LLMProvider) error {
	query := `
		INSERT INTO llm_providers (name, type, api_key, base_url, model, weight, enabled, config, health_status, response_time)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`

	configJSON, err := json.Marshal(provider.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal provider config: %w", err)
	}

	err = r.pool.QueryRow(ctx, query,
		provider.Name, provider.Type, provider.APIKey, provider.BaseURL, provider.Model,
		provider.Weight, provider.Enabled, configJSON, provider.HealthStatus, provider.ResponseTime,
	).Scan(&provider.ID, &provider.CreatedAt, &provider.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	return nil
}

// GetByID retrieves a provider by its ID
func (r *ProviderRepository) GetByID(ctx context.Context, id string) (*LLMProvider, error) {
	query := `
		SELECT id, name, type, api_key, base_url, model, weight, enabled, config, health_status, response_time, created_at, updated_at
		FROM llm_providers
		WHERE id = $1
	`

	provider := &LLMProvider{}
	var configJSON []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&provider.ID, &provider.Name, &provider.Type, &provider.APIKey, &provider.BaseURL,
		&provider.Model, &provider.Weight, &provider.Enabled, &configJSON,
		&provider.HealthStatus, &provider.ResponseTime, &provider.CreatedAt, &provider.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("provider not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	if len(configJSON) > 0 {
		if err := json.Unmarshal(configJSON, &provider.Config); err != nil {
			provider.Config = make(map[string]interface{})
		}
	}

	return provider, nil
}

// GetByName retrieves a provider by its name
func (r *ProviderRepository) GetByName(ctx context.Context, name string) (*LLMProvider, error) {
	query := `
		SELECT id, name, type, api_key, base_url, model, weight, enabled, config, health_status, response_time, created_at, updated_at
		FROM llm_providers
		WHERE name = $1
	`

	provider := &LLMProvider{}
	var configJSON []byte

	err := r.pool.QueryRow(ctx, query, name).Scan(
		&provider.ID, &provider.Name, &provider.Type, &provider.APIKey, &provider.BaseURL,
		&provider.Model, &provider.Weight, &provider.Enabled, &configJSON,
		&provider.HealthStatus, &provider.ResponseTime, &provider.CreatedAt, &provider.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("provider not found: %s", name)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get provider by name: %w", err)
	}

	if len(configJSON) > 0 {
		if err := json.Unmarshal(configJSON, &provider.Config); err != nil {
			provider.Config = make(map[string]interface{})
		}
	}

	return provider, nil
}

// Update updates an existing provider
func (r *ProviderRepository) Update(ctx context.Context, provider *LLMProvider) error {
	query := `
		UPDATE llm_providers
		SET name = $2, type = $3, api_key = $4, base_url = $5, model = $6, weight = $7,
		    enabled = $8, config = $9, health_status = $10, response_time = $11, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`

	configJSON, err := json.Marshal(provider.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal provider config: %w", err)
	}

	err = r.pool.QueryRow(ctx, query,
		provider.ID, provider.Name, provider.Type, provider.APIKey, provider.BaseURL,
		provider.Model, provider.Weight, provider.Enabled, configJSON,
		provider.HealthStatus, provider.ResponseTime,
	).Scan(&provider.UpdatedAt)

	if err == pgx.ErrNoRows {
		return fmt.Errorf("provider not found: %s", provider.ID)
	}
	if err != nil {
		return fmt.Errorf("failed to update provider: %w", err)
	}

	return nil
}

// Delete deletes a provider by its ID
func (r *ProviderRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM llm_providers WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete provider: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("provider not found: %s", id)
	}

	return nil
}

// List retrieves all providers with pagination
func (r *ProviderRepository) List(ctx context.Context, limit, offset int) ([]*LLMProvider, int, error) {
	countQuery := `SELECT COUNT(*) FROM llm_providers`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count providers: %w", err)
	}

	query := `
		SELECT id, name, type, COALESCE(api_key, ''), COALESCE(base_url, ''), COALESCE(model, ''), weight, enabled, COALESCE(config, '{}'), COALESCE(health_status, ''), response_time, created_at, updated_at
		FROM llm_providers
		ORDER BY name ASC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list providers: %w", err)
	}
	defer rows.Close()

	providers := []*LLMProvider{}
	for rows.Next() {
		provider := &LLMProvider{}
		var configJSON []byte

		err := rows.Scan(
			&provider.ID, &provider.Name, &provider.Type, &provider.APIKey, &provider.BaseURL,
			&provider.Model, &provider.Weight, &provider.Enabled, &configJSON,
			&provider.HealthStatus, &provider.ResponseTime, &provider.CreatedAt, &provider.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan provider row: %w", err)
		}

		if len(configJSON) > 0 {
			if err := json.Unmarshal(configJSON, &provider.Config); err != nil {
				provider.Config = make(map[string]interface{})
			}
		}

		providers = append(providers, provider)
	}

	return providers, total, nil
}

// ListEnabled retrieves all enabled providers
func (r *ProviderRepository) ListEnabled(ctx context.Context) ([]*LLMProvider, error) {
	query := `
		SELECT id, name, type, COALESCE(api_key, ''), COALESCE(base_url, ''), COALESCE(model, ''), weight, enabled, COALESCE(config, '{}'), COALESCE(health_status, ''), response_time, created_at, updated_at
		FROM llm_providers
		WHERE enabled = true
		ORDER BY weight DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list enabled providers: %w", err)
	}
	defer rows.Close()

	providers := []*LLMProvider{}
	for rows.Next() {
		provider := &LLMProvider{}
		var configJSON []byte

		err := rows.Scan(
			&provider.ID, &provider.Name, &provider.Type, &provider.APIKey, &provider.BaseURL,
			&provider.Model, &provider.Weight, &provider.Enabled, &configJSON,
			&provider.HealthStatus, &provider.ResponseTime, &provider.CreatedAt, &provider.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan provider row: %w", err)
		}

		if len(configJSON) > 0 {
			if err := json.Unmarshal(configJSON, &provider.Config); err != nil {
				provider.Config = make(map[string]interface{})
			}
		}

		providers = append(providers, provider)
	}

	return providers, nil
}

// UpdateHealth updates the health status and response time for a provider
func (r *ProviderRepository) UpdateHealth(ctx context.Context, id, status string, responseTime int64) error {
	query := `
		UPDATE llm_providers
		SET health_status = $2, response_time = $3, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, id, status, responseTime)
	if err != nil {
		return fmt.Errorf("failed to update provider health: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("provider not found: %s", id)
	}

	return nil
}

// SetEnabled enables or disables a provider
func (r *ProviderRepository) SetEnabled(ctx context.Context, id string, enabled bool) error {
	query := `
		UPDATE llm_providers
		SET enabled = $2, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, id, enabled)
	if err != nil {
		return fmt.Errorf("failed to set provider enabled status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("provider not found: %s", id)
	}

	return nil
}

// UpdateWeight updates the weight for a provider
func (r *ProviderRepository) UpdateWeight(ctx context.Context, id string, weight float64) error {
	query := `
		UPDATE llm_providers
		SET weight = $2, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, id, weight)
	if err != nil {
		return fmt.Errorf("failed to update provider weight: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("provider not found: %s", id)
	}

	return nil
}

// ExistsByName checks if a provider exists with the given name
func (r *ProviderRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM llm_providers WHERE name = $1)`
	var exists bool
	if err := r.pool.QueryRow(ctx, query, name).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to check provider existence: %w", err)
	}
	return exists, nil
}

// GetHealthyProviders retrieves all providers with healthy status
func (r *ProviderRepository) GetHealthyProviders(ctx context.Context) ([]*LLMProvider, error) {
	query := `
		SELECT id, name, type, api_key, base_url, model, weight, enabled, config, health_status, response_time, created_at, updated_at
		FROM llm_providers
		WHERE enabled = true AND health_status = 'healthy'
		ORDER BY weight DESC, response_time ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list healthy providers: %w", err)
	}
	defer rows.Close()

	providers := []*LLMProvider{}
	for rows.Next() {
		provider := &LLMProvider{}
		var configJSON []byte

		err := rows.Scan(
			&provider.ID, &provider.Name, &provider.Type, &provider.APIKey, &provider.BaseURL,
			&provider.Model, &provider.Weight, &provider.Enabled, &configJSON,
			&provider.HealthStatus, &provider.ResponseTime, &provider.CreatedAt, &provider.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan provider row: %w", err)
		}

		if len(configJSON) > 0 {
			if err := json.Unmarshal(configJSON, &provider.Config); err != nil {
				provider.Config = make(map[string]interface{})
			}
		}

		providers = append(providers, provider)
	}

	return providers, nil
}
