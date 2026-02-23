package database_test

import (
	"testing"

	adapter "dev.helix.agent/internal/adapters/database"
	"dev.helix.agent/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// MemoryDB Tests (no database connection needed)
// ============================================================================

func TestNewMemoryDB(t *testing.T) {
	db := adapter.NewMemoryDB()
	require.NotNil(t, db)
}

func TestMemoryDB_Ping(t *testing.T) {
	db := adapter.NewMemoryDB()
	err := db.Ping()
	assert.NoError(t, err)
}

func TestMemoryDB_Exec(t *testing.T) {
	db := adapter.NewMemoryDB()
	err := db.Exec("SELECT 1")
	assert.NoError(t, err)
}

func TestMemoryDB_Query(t *testing.T) {
	db := adapter.NewMemoryDB()
	results, err := db.Query("SELECT * FROM users")
	assert.NoError(t, err)
	assert.Nil(t, results)
}

func TestMemoryDB_QueryRow(t *testing.T) {
	db := adapter.NewMemoryDB()
	row := db.QueryRow("SELECT 1")
	assert.NotNil(t, row)

	// Scan should return an error since no data is stored
	var val string
	err := row.Scan(&val)
	assert.Error(t, err)
}

func TestMemoryDB_HealthCheck(t *testing.T) {
	db := adapter.NewMemoryDB()
	err := db.HealthCheck()
	assert.NoError(t, err)
}

func TestMemoryDB_IsMemoryMode(t *testing.T) {
	db := adapter.NewMemoryDB()
	assert.True(t, db.IsMemoryMode())
}

func TestMemoryDB_GetPool(t *testing.T) {
	db := adapter.NewMemoryDB()
	pool := db.GetPool()
	// In-memory DB has no real pool
	assert.Nil(t, pool)
}

func TestMemoryDB_StoreRow(t *testing.T) {
	db := adapter.NewMemoryDB()
	// StoreRow should not panic
	db.StoreRow("users", "user-001", []any{"Alice", 30})
}

func TestMemoryDB_Close(t *testing.T) {
	db := adapter.NewMemoryDB()
	err := db.Close()
	assert.NoError(t, err)

	// After close, HealthCheck should fail
	err = db.HealthCheck()
	assert.Error(t, err)
}

// ============================================================================
// NewPostgresDBWithFallback - falls back to MemoryDB without PostgreSQL
// ============================================================================

func TestNewPostgresDBWithFallback_FallsBackToMemory(t *testing.T) {
	// Use a config pointing to a non-existent database host
	cfg := &config.Config{}
	cfg.Database.Host = "localhost"
	cfg.Database.Port = "59999" // unlikely to have anything running
	cfg.Database.User = "nouser"
	cfg.Database.Password = "nopass"
	cfg.Database.Name = "nodb"

	_, memDB, err := adapter.NewPostgresDBWithFallback(cfg)
	// Since no DB is running: err should be nil and memDB should be set
	assert.NoError(t, err)
	require.NotNil(t, memDB)
	assert.True(t, memDB.IsMemoryMode())
}

// ============================================================================
// DB interface compliance
// ============================================================================

func TestMemoryDB_ImplementsDBInterface(t *testing.T) {
	db := adapter.NewMemoryDB()
	// Compile-time check that MemoryDB implements DB interface
	var _ adapter.DB = db
	t.Log("MemoryDB implements DB interface")
}

func TestMemoryDB_ImplementsLegacyDBInterface(t *testing.T) {
	db := adapter.NewMemoryDB()
	var _ adapter.LegacyDB = db
	t.Log("MemoryDB implements LegacyDB interface")
}
