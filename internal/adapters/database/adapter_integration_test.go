//go:build integration

package database_test

import (
	"context"
	"os"
	"testing"
	"time"

	adapter "dev.helix.agent/internal/adapters/database"
	"dev.helix.agent/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// skipIfNoPostgreSQL skips the test when PostgreSQL is not available.
func skipIfNoPostgreSQL(t *testing.T) {
	t.Helper()
	// Set test infrastructure environment variables if not already set
	if os.Getenv("DB_HOST") == "" {
		t.Setenv("DB_HOST", "localhost")
	}
	if os.Getenv("DB_PORT") == "" {
		t.Setenv("DB_PORT", "5432")
	}
	if os.Getenv("DB_USER") == "" {
		t.Setenv("DB_USER", "helixagent")
	}
	if os.Getenv("DB_PASSWORD") == "" {
		t.Setenv("DB_PASSWORD", "helixagent123")
	}
	if os.Getenv("DB_NAME") == "" {
		t.Setenv("DB_NAME", "helixagent_db")
	}
}

func TestIntegration_NewClient_WithRealPostgreSQL(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	skipIfNoPostgreSQL(t)
	// Use environment variables set by test infrastructure
	cfg := &config.Config{}
	// If env vars are set, they will be picked up by buildPostgresConfig
	client, err := adapter.NewClient(cfg)
	if err != nil {
		t.Skip("PostgreSQL not available: ", err)
	}
	defer func() {
		_ = client.Close()
	}()
	assert.NotNil(t, client)
	assert.NotNil(t, client.Pool())
	assert.NoError(t, client.Ping())
	assert.NoError(t, client.HealthCheck())
}

func TestIntegration_NewClientWithFallback_RealPostgreSQL(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	skipIfNoPostgreSQL(t)
	cfg := &config.Config{}
	client, memDB, err := adapter.NewPostgresDBWithFallback(cfg)
	require.NoError(t, err)
	// Should have real client, not memory fallback
	require.NotNil(t, client)
	assert.Nil(t, memDB)
	defer func() {
		_ = client.Close()
	}()
	assert.NoError(t, client.Ping())
}

func TestIntegration_Client_Exec(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	cfg := &config.Config{}
	client, err := adapter.NewClient(cfg)
	if err != nil {
		t.Skip("PostgreSQL not available: ", err)
	}
	defer func() { _ = client.Close() }()

	// Create a test table
	createSQL := `CREATE TABLE IF NOT EXISTS test_integration (id SERIAL PRIMARY KEY, name TEXT)`
	err = client.Exec(createSQL)
	assert.NoError(t, err)
	// Insert a row
	err = client.Exec(`INSERT INTO test_integration (name) VALUES ($1)`, "test")
	assert.NoError(t, err)
	// Query rows
	results, err := client.Query(`SELECT name FROM test_integration`)
	assert.NoError(t, err)
	assert.NotNil(t, results)
	// Cleanup
	err = client.Exec(`DROP TABLE test_integration`)
	assert.NoError(t, err)
}

func TestIntegration_Client_QueryRow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	cfg := &config.Config{}
	client, err := adapter.NewClient(cfg)
	if err != nil {
		t.Skip("PostgreSQL not available: ", err)
	}
	defer func() { _ = client.Close() }()

	err = client.Exec(`CREATE TEMP TABLE temp_test (id INT, val TEXT)`)
	assert.NoError(t, err)
	err = client.Exec(`INSERT INTO temp_test VALUES (1, 'hello')`)
	assert.NoError(t, err)

	row := client.QueryRow(`SELECT val FROM temp_test WHERE id = $1`, 1)
	var val string
	err = row.Scan(&val)
	assert.NoError(t, err)
	assert.Equal(t, "hello", val)
}

func TestIntegration_Client_Begin(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	cfg := &config.Config{}
	client, err := adapter.NewClient(cfg)
	if err != nil {
		t.Skip("PostgreSQL not available: ", err)
	}
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	tx, err := client.Begin(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, tx)
	// Rollback to avoid side effects
	_ = tx.Rollback(ctx)
}

func TestIntegration_Client_Migrate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	cfg := &config.Config{}
	client, err := adapter.NewClient(cfg)
	if err != nil {
		t.Skip("PostgreSQL not available: ", err)
	}
	defer func() { _ = client.Close() }()

	ctx := context.Background()
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS migration_test (id INT)`,
	}
	err = client.Migrate(ctx, migrations)
	assert.NoError(t, err)
	// Cleanup
	_ = client.Exec(`DROP TABLE migration_test`)
}
