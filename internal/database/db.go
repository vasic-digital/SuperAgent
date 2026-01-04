package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/superagent/superagent/internal/config"
)

// Row interface for row scanning
type Row interface {
	Scan(dest ...any) error
}

// DB interface for database operations
type DB interface {
	Ping() error
	Exec(query string, args ...any) error
	Query(query string, args ...any) ([]any, error)
	QueryRow(query string, args ...any) Row
	Close() error
	HealthCheck() error
}

// PostgresDB implements DB using PostgreSQL with pgxpool
type PostgresDB struct {
	pool *pgxpool.Pool
}

// pgxRow wraps pgx.Row to implement the Row interface
type pgxRow struct {
	row pgx.Row
}

// Scan implements the Row interface
func (r *pgxRow) Scan(dest ...any) error {
	return r.row.Scan(dest...)
}

func NewPostgresDB(cfg *config.Config) (*PostgresDB, error) {
	// Use config values if provided, otherwise fall back to environment variables with defaults
	dbHost := cfg.Database.Host
	if dbHost == "" {
		dbHost = getEnv("DB_HOST", "localhost")
	}

	dbPort := cfg.Database.Port
	if dbPort == "" {
		dbPort = getEnv("DB_PORT", "5432")
	}

	dbUser := cfg.Database.User
	if dbUser == "" {
		dbUser = getEnv("DB_USER", "superagent")
	}

	dbPassword := cfg.Database.Password
	if dbPassword == "" {
		dbPassword = getEnv("DB_PASSWORD", "secret")
	}

	dbName := cfg.Database.Name
	if dbName == "" {
		dbName = getEnv("DB_NAME", "superagent_db")
	}

	sslMode := cfg.Database.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		dbUser, dbPassword, dbHost, dbPort, dbName, sslMode)

	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		log.Printf("Warning: Database connection test failed: %v", err)
	}

	log.Printf("Connected to PostgreSQL database: %s", dbName)
	return &PostgresDB{pool: pool}, nil
}

func (p *PostgresDB) Ping() error {
	return p.pool.Ping(context.Background())
}

func (p *PostgresDB) Exec(query string, args ...any) error {
	_, err := p.pool.Exec(context.Background(), query, args...)
	return err
}

func (p *PostgresDB) Query(query string, args ...any) ([]any, error) {
	rows, err := p.pool.Query(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []any
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}
		results = append(results, values)
	}
	return results, nil
}

func (p *PostgresDB) QueryRow(query string, args ...any) Row {
	row := p.pool.QueryRow(context.Background(), query, args...)
	return &pgxRow{row: row}
}

func (p *PostgresDB) Close() error {
	p.pool.Close()
	return nil
}

// GetPool returns the underlying connection pool
func (p *PostgresDB) GetPool() *pgxpool.Pool {
	return p.pool
}

// HealthCheck performs a health check on the database.
func (p *PostgresDB) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return p.pool.Ping(ctx)
}

// getEnv gets environment variable or returns default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// RunMigration executes database migrations
func RunMigration(db *PostgresDB, migrations []string) error {
	for _, migration := range migrations {
		log.Printf("Running migration: %s", migration)
		if err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", migration, err)
		}
	}

	log.Printf("All migrations completed successfully")
	return nil
}

// Migrations for the LLM facade
var migrations = []string{
	`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`,

	`CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		username VARCHAR(255) UNIQUE NOT NULL,
		email VARCHAR(255) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		api_key VARCHAR(255) UNIQUE NOT NULL,
		role VARCHAR(50) DEFAULT 'user',
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	)`,

	`CREATE TABLE IF NOT EXISTS user_sessions (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		user_id UUID REFERENCES users(id) ON DELETE CASCADE,
		session_token VARCHAR(255) UNIQUE NOT NULL,
		context JSONB DEFAULT '{}',
		memory_id UUID,
		status VARCHAR(50) DEFAULT 'active',
		request_count INTEGER DEFAULT 0,
		last_activity TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	)`,

	`CREATE TABLE IF NOT EXISTS llm_providers (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		name VARCHAR(255) UNIQUE NOT NULL,
		type VARCHAR(100) NOT NULL,
		api_key VARCHAR(255),
		base_url VARCHAR(500),
		model VARCHAR(255),
		weight DECIMAL(5,2) DEFAULT 1.0,
		enabled BOOLEAN DEFAULT TRUE,
		config JSONB DEFAULT '{}',
		health_status VARCHAR(50) DEFAULT 'unknown',
		response_time BIGINT DEFAULT 0,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	)`,

	`CREATE TABLE IF NOT EXISTS llm_requests (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		session_id UUID REFERENCES user_sessions(id) ON DELETE CASCADE,
		user_id UUID REFERENCES users(id) ON DELETE CASCADE,
		prompt TEXT NOT NULL,
		messages JSONB NOT NULL DEFAULT '[]',
		model_params JSONB NOT NULL DEFAULT '{}',
		ensemble_config JSONB DEFAULT NULL,
		memory_enhanced BOOLEAN DEFAULT FALSE,
		memory JSONB DEFAULT '{}',
		status VARCHAR(50) DEFAULT 'pending',
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		started_at TIMESTAMP WITH TIME ZONE,
		completed_at TIMESTAMP WITH TIME ZONE,
		request_type VARCHAR(50) DEFAULT 'completion'
	)`,

	`CREATE TABLE IF NOT EXISTS llm_responses (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		request_id UUID REFERENCES llm_requests(id) ON DELETE CASCADE,
		provider_id UUID REFERENCES llm_providers(id) ON DELETE SET NULL,
		provider_name VARCHAR(100) NOT NULL,
		content TEXT NOT NULL,
		confidence DECIMAL(3,2) NOT NULL DEFAULT 0.0,
		tokens_used INTEGER DEFAULT 0,
		response_time BIGINT DEFAULT 0,
		finish_reason VARCHAR(50) DEFAULT 'stop',
		metadata JSONB DEFAULT '{}',
		selected BOOLEAN DEFAULT FALSE,
		selection_score DECIMAL(5,2) DEFAULT 0.0,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	)`,

	`CREATE TABLE IF NOT EXISTS cognee_memories (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		session_id UUID REFERENCES user_sessions(id) ON DELETE CASCADE,
		dataset_name VARCHAR(255) NOT NULL,
		content_type VARCHAR(50) DEFAULT 'text',
		content TEXT NOT NULL,
		vector_id VARCHAR(255),
		graph_nodes JSONB DEFAULT '{}',
		search_key VARCHAR(255),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	)`,

	`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
	`CREATE INDEX IF NOT EXISTS idx_users_api_key ON users(api_key)`,
	`CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id)`,
	`CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions(expires_at)`,
	`CREATE INDEX IF NOT EXISTS idx_user_sessions_session_token ON user_sessions(session_token)`,
	`CREATE INDEX IF NOT EXISTS idx_llm_providers_name ON llm_providers(name)`,
	`CREATE INDEX IF NOT EXISTS idx_llm_providers_enabled ON llm_providers(enabled)`,
	`CREATE INDEX IF NOT EXISTS idx_llm_requests_session_id ON llm_requests(session_id)`,
	`CREATE INDEX IF NOT EXISTS idx_llm_requests_user_id ON llm_requests(user_id)`,
	`CREATE INDEX IF NOT EXISTS idx_llm_requests_status ON llm_requests(status)`,
	`CREATE INDEX IF NOT EXISTS idx_llm_responses_request_id ON llm_responses(request_id)`,
	`CREATE INDEX IF NOT EXISTS idx_llm_responses_provider_id ON llm_responses(provider_id)`,
	`CREATE INDEX IF NOT EXISTS idx_llm_responses_selected ON llm_responses(selected)`,
	`CREATE INDEX IF NOT EXISTS idx_cognee_memories_session_id ON cognee_memories(session_id)`,
	`CREATE INDEX IF NOT EXISTS idx_cognee_memories_dataset_name ON cognee_memories(dataset_name)`,
	`CREATE INDEX IF NOT EXISTS idx_cognee_memories_search_key ON cognee_memories(search_key)`,

	// Migration 002: Models.dev Integration
	`ALTER TABLE llm_providers ADD COLUMN IF NOT EXISTS modelsdev_provider_id VARCHAR(255)`,
	`ALTER TABLE llm_providers ADD COLUMN IF NOT EXISTS total_models INTEGER DEFAULT 0`,
	`ALTER TABLE llm_providers ADD COLUMN IF NOT EXISTS enabled_models INTEGER DEFAULT 0`,
	`ALTER TABLE llm_providers ADD COLUMN IF NOT EXISTS last_models_sync TIMESTAMP WITH TIME ZONE`,

	`CREATE TABLE IF NOT EXISTS models_metadata (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		model_id VARCHAR(255) UNIQUE NOT NULL,
		model_name VARCHAR(255) NOT NULL,
		provider_id VARCHAR(255) NOT NULL,
		provider_name VARCHAR(255) NOT NULL,
		description TEXT,
		context_window INTEGER,
		max_tokens INTEGER,
		pricing_input DECIMAL(10, 6),
		pricing_output DECIMAL(10, 6),
		pricing_currency VARCHAR(10) DEFAULT 'USD',
		supports_vision BOOLEAN DEFAULT FALSE,
		supports_function_calling BOOLEAN DEFAULT FALSE,
		supports_streaming BOOLEAN DEFAULT FALSE,
		supports_json_mode BOOLEAN DEFAULT FALSE,
		supports_image_generation BOOLEAN DEFAULT FALSE,
		supports_audio BOOLEAN DEFAULT FALSE,
		supports_code_generation BOOLEAN DEFAULT FALSE,
		supports_reasoning BOOLEAN DEFAULT FALSE,
		benchmark_score DECIMAL(5, 2),
		popularity_score INTEGER,
		reliability_score DECIMAL(5, 2),
		model_type VARCHAR(100),
		model_family VARCHAR(100),
		version VARCHAR(50),
		tags JSONB DEFAULT '[]',
		modelsdev_url TEXT,
		modelsdev_id VARCHAR(255),
		modelsdev_api_version VARCHAR(50),
		raw_metadata JSONB DEFAULT '{}',
		last_refreshed_at TIMESTAMP WITH TIME ZONE NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	)`,

	`CREATE TABLE IF NOT EXISTS model_benchmarks (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		model_id VARCHAR(255) NOT NULL,
		benchmark_name VARCHAR(255) NOT NULL,
		benchmark_type VARCHAR(100),
		score DECIMAL(10, 4),
		rank INTEGER,
		normalized_score DECIMAL(5, 2),
		benchmark_date DATE,
		metadata JSONB DEFAULT '{}',
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	)`,

	`CREATE TABLE IF NOT EXISTS models_refresh_history (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		refresh_type VARCHAR(50) NOT NULL,
		status VARCHAR(50) NOT NULL,
		models_refreshed INTEGER DEFAULT 0,
		models_failed INTEGER DEFAULT 0,
		error_message TEXT,
		started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		completed_at TIMESTAMP WITH TIME ZONE,
		duration_seconds INTEGER,
		metadata JSONB DEFAULT '{}'
	)`,

	`CREATE INDEX IF NOT EXISTS idx_models_metadata_provider_id ON models_metadata(provider_id)`,
	`CREATE INDEX IF NOT EXISTS idx_models_metadata_model_type ON models_metadata(model_type)`,
	`CREATE INDEX IF NOT EXISTS idx_models_metadata_last_refreshed ON models_metadata(last_refreshed_at)`,
	`CREATE INDEX IF NOT EXISTS idx_benchmarks_model_id ON model_benchmarks(model_id)`,
	`CREATE INDEX IF NOT EXISTS idx_refresh_history_started ON models_refresh_history(started_at)`,

	// Migration 003: Protocol Support
	`ALTER TABLE models_metadata ADD COLUMN IF NOT EXISTS protocol_support JSONB DEFAULT '[]'`,
	`ALTER TABLE models_metadata ADD COLUMN IF NOT EXISTS mcp_server_id VARCHAR(255)`,
	`ALTER TABLE models_metadata ADD COLUMN IF NOT EXISTS lsp_server_id VARCHAR(255)`,
	`ALTER TABLE models_metadata ADD COLUMN IF NOT EXISTS acp_server_id VARCHAR(255)`,
	`ALTER TABLE models_metadata ADD COLUMN IF NOT EXISTS embedding_provider VARCHAR(50) DEFAULT 'pgvector'`,
	`ALTER TABLE models_metadata ADD COLUMN IF NOT EXISTS protocol_config JSONB DEFAULT '{}'`,
	`ALTER TABLE models_metadata ADD COLUMN IF NOT EXISTS protocol_last_sync TIMESTAMP WITH TIME ZONE DEFAULT NOW()`,

	`CREATE TABLE IF NOT EXISTS mcp_servers (
		id VARCHAR(255) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		type VARCHAR(20) NOT NULL,
		command TEXT,
		url TEXT,
		enabled BOOLEAN NOT NULL DEFAULT true,
		tools JSONB NOT NULL DEFAULT '[]',
		last_sync TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	)`,

	`CREATE TABLE IF NOT EXISTS lsp_servers (
		id VARCHAR(255) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		language VARCHAR(50) NOT NULL,
		command VARCHAR(500) NOT NULL,
		enabled BOOLEAN NOT NULL DEFAULT true,
		workspace VARCHAR(1000) DEFAULT '/workspace',
		capabilities JSONB NOT NULL DEFAULT '[]',
		last_sync TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	)`,

	`CREATE TABLE IF NOT EXISTS acp_servers (
		id VARCHAR(255) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		type VARCHAR(20) NOT NULL,
		url TEXT,
		enabled BOOLEAN NOT NULL DEFAULT true,
		tools JSONB NOT NULL DEFAULT '[]',
		last_sync TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	)`,

	`CREATE TABLE IF NOT EXISTS embedding_config (
		id SERIAL PRIMARY KEY,
		provider VARCHAR(50) NOT NULL DEFAULT 'pgvector',
		model VARCHAR(100) NOT NULL DEFAULT 'text-embedding-ada-002',
		dimension INTEGER NOT NULL DEFAULT 1536,
		api_endpoint TEXT,
		api_key TEXT,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	)`,

	`CREATE TABLE IF NOT EXISTS protocol_cache (
		cache_key VARCHAR(255) PRIMARY KEY,
		cache_data JSONB NOT NULL,
		expires_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	)`,

	`CREATE TABLE IF NOT EXISTS protocol_metrics (
		id SERIAL PRIMARY KEY,
		protocol_type VARCHAR(20) NOT NULL,
		server_id VARCHAR(255),
		operation VARCHAR(100) NOT NULL,
		status VARCHAR(20) NOT NULL,
		duration_ms INTEGER,
		error_message TEXT,
		metadata JSONB DEFAULT '{}',
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	)`,

	`CREATE INDEX IF NOT EXISTS idx_mcp_servers_enabled ON mcp_servers(enabled)`,
	`CREATE INDEX IF NOT EXISTS idx_lsp_servers_enabled ON lsp_servers(enabled)`,
	`CREATE INDEX IF NOT EXISTS idx_acp_servers_enabled ON acp_servers(enabled)`,
	`CREATE INDEX IF NOT EXISTS idx_protocol_cache_expires_at ON protocol_cache(expires_at)`,
	`CREATE INDEX IF NOT EXISTS idx_protocol_metrics_protocol_type ON protocol_metrics(protocol_type)`,
	`CREATE INDEX IF NOT EXISTS idx_protocol_metrics_created_at ON protocol_metrics(created_at)`,
}

// Legacy interface for backward compatibility
type LegacyDB interface {
	Ping() error
	Exec(query string, args ...any) error
	Query(query string, args ...any) ([]any, error)
	Close() error
}

// Connect establishes a real PostgreSQL connection via pgx.
func Connect() (LegacyDB, error) {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "superagent")
	dbPassword := getEnv("DB_PASSWORD", "secret")
	dbName := getEnv("DB_NAME", "superagent_db")

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &PostgresDB{pool: pool}, nil
}
