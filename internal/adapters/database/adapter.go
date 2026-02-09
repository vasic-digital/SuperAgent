// Package database provides adapters that bridge HelixAgent-specific database
// operations with the generic digital.vasic.database module.
//
// This adapter layer maintains backward compatibility with existing code while
// allowing gradual migration to the extracted database module.
package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	db "digital.vasic.database/pkg/database"
	"digital.vasic.database/pkg/postgres"

	"dev.helix.agent/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Client wraps the digital.vasic.database postgres client and provides
// backward-compatible methods for HelixAgent's existing code.
type Client struct {
	pg   *postgres.Client
	pool *pgxpool.Pool
}

// NewClient creates a new database client using the extracted database module.
// It accepts the HelixAgent config and converts it to the module's config format.
func NewClient(cfg *config.Config) (*Client, error) {
	// Build postgres config from HelixAgent config
	pgCfg := buildPostgresConfig(cfg)

	// Create the postgres client
	pg := postgres.New(pgCfg)

	// Connect with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := pg.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Printf("Connected to PostgreSQL database: %s", pgCfg.DBName)

	return &Client{
		pg:   pg,
		pool: pg.Pool(),
	}, nil
}

// NewClientWithFallback tries to create a database client, returning nil on failure.
// This is useful for standalone mode where database may not be available.
func NewClientWithFallback(cfg *config.Config) (*Client, error) {
	client, err := NewClient(cfg)
	if err != nil {
		log.Printf("Database connection failed: %v, standalone mode may be used", err)
		return nil, err
	}
	return client, nil
}

// buildPostgresConfig converts HelixAgent config to postgres module config.
func buildPostgresConfig(cfg *config.Config) *postgres.Config {
	pgCfg := postgres.DefaultConfig()

	// Use config values if provided, otherwise fall back to environment variables
	if cfg.Database.Host != "" {
		pgCfg.Host = cfg.Database.Host
	} else if host := os.Getenv("DB_HOST"); host != "" {
		pgCfg.Host = host
	}

	if cfg.Database.Port != "" {
		if port, err := strconv.Atoi(cfg.Database.Port); err == nil {
			pgCfg.Port = port
		}
	} else if port := os.Getenv("DB_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			pgCfg.Port = p
		}
	}

	if cfg.Database.User != "" {
		pgCfg.User = cfg.Database.User
	} else if user := os.Getenv("DB_USER"); user != "" {
		pgCfg.User = user
	}

	if cfg.Database.Password != "" {
		pgCfg.Password = cfg.Database.Password
	} else if pass := os.Getenv("DB_PASSWORD"); pass != "" {
		pgCfg.Password = pass
	}

	if cfg.Database.Name != "" {
		pgCfg.DBName = cfg.Database.Name
	} else if name := os.Getenv("DB_NAME"); name != "" {
		pgCfg.DBName = name
	}

	if cfg.Database.SSLMode != "" {
		pgCfg.SSLMode = cfg.Database.SSLMode
	}

	pgCfg.ApplicationName = "helixagent"

	return pgCfg
}

// Pool returns the underlying pgxpool.Pool for direct access.
// This provides backward compatibility with existing repository code.
func (c *Client) Pool() *pgxpool.Pool {
	return c.pool
}

// Database returns the underlying database.Database interface.
func (c *Client) Database() db.Database {
	return c.pg
}

// Close closes the database connection.
func (c *Client) Close() error {
	return c.pg.Close()
}

// Ping tests the database connection.
func (c *Client) Ping() error {
	return c.pg.HealthCheck(context.Background())
}

// HealthCheck performs a health check on the database.
func (c *Client) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return c.pg.HealthCheck(ctx)
}

// Exec executes a query that doesn't return rows.
// This method provides backward compatibility with the old interface.
func (c *Client) Exec(query string, args ...any) error {
	_, err := c.pg.Exec(context.Background(), query, args...)
	return err
}

// Query executes a query that returns rows.
// Returns results as a slice of any for backward compatibility.
func (c *Client) Query(query string, args ...any) ([]any, error) {
	rows, err := c.pg.Query(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var results []any
	for rows.Next() {
		// This is a simplified implementation - actual scanning would depend on query
		results = append(results, nil)
	}
	return results, rows.Err()
}

// QueryRow executes a query expected to return at most one row.
func (c *Client) QueryRow(query string, args ...any) db.Row {
	return c.pg.QueryRow(context.Background(), query, args...)
}

// Begin starts a new transaction.
func (c *Client) Begin(ctx context.Context) (db.Tx, error) {
	return c.pg.Begin(ctx)
}

// Migrate runs the provided migration statements.
func (c *Client) Migrate(ctx context.Context, migrations []string) error {
	return c.pg.Migrate(ctx, migrations)
}

// Row is an alias for the database module's Row interface.
type Row = db.Row

// Rows is an alias for the database module's Rows interface.
type Rows = db.Rows

// Tx is an alias for the database module's Tx interface.
type Tx = db.Tx

// Result is an alias for the database module's Result interface.
type Result = db.Result
