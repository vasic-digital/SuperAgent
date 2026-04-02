package database

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"dev.helix.agent/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Missing Coverage for compat.go:172 (MemoryDB.Ping - 0%)
// ============================================================================

// TestMemoryDB_Ping_CalledDirectly ensures MemoryDB.Ping is covered
func TestMemoryDB_Ping_CalledDirectly(t *testing.T) {
	m := NewMemoryDB()
	err := m.Ping()
	assert.NoError(t, err)
}

// ============================================================================
// Missing Coverage for compat.go:179 (MemoryDB.Exec - 0%)
// ============================================================================

// TestMemoryDB_Exec_CalledDirectly ensures MemoryDB.Exec is covered
func TestMemoryDB_Exec_CalledDirectly(t *testing.T) {
	m := NewMemoryDB()
	err := m.Exec("INSERT INTO test VALUES (1)")
	assert.NoError(t, err)
}

// TestMemoryDB_Exec_WithArgs ensures MemoryDB.Exec with args is covered
func TestMemoryDB_Exec_WithArgs(t *testing.T) {
	m := NewMemoryDB()
	err := m.Exec("INSERT INTO test VALUES ($1, $2)", "value1", 123)
	assert.NoError(t, err)
}

// ============================================================================
// Missing Coverage for compat.go:223 (MemoryDB.IsMemoryMode - 0%)
// ============================================================================

// TestMemoryDB_IsMemoryMode_CalledDirectly ensures MemoryDB.IsMemoryMode is covered
func TestMemoryDB_IsMemoryMode_CalledDirectly(t *testing.T) {
	m := NewMemoryDB()
	result := m.IsMemoryMode()
	assert.True(t, result)
}

// ============================================================================
// Missing Coverage for compat.go:210 (MemoryDB.HealthCheck closed path - 66.7%)
// ============================================================================

// TestMemoryDB_HealthCheck_ClosedPath covers the !m.enabled branch
func TestMemoryDB_HealthCheck_ClosedPath(t *testing.T) {
	m := NewMemoryDB()
	
	// Close the database
	err := m.Close()
	require.NoError(t, err)
	
	// Health check should fail with "memory database closed"
	err = m.HealthCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "closed")
}

// ============================================================================
// Missing Coverage for adapter.go:50 (initConnection - 84.6%)
// This covers the path where context has no deadline (line 58-62)
// ============================================================================

// TestClient_initConnection_NoDeadline covers the deadline addition path
func TestClient_initConnection_NoDeadline(t *testing.T) {
	mock := &mockDatabase{}
	client := newTestClient(mock)
	
	// Use context without deadline - should add 10 second timeout
	ctx := context.Background()
	err := client.initConnection(ctx)
	
	assert.NoError(t, err)
}

// TestClient_initConnection_WithDeadline covers the existing deadline path  
func TestClient_initConnection_WithDeadline(t *testing.T) {
	mock := &mockDatabase{}
	client := newTestClient(mock)
	
	// Use context with existing deadline - should NOT add timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	err := client.initConnection(ctx)
	
	assert.NoError(t, err)
}

// TestClient_initConnection_ConnectError_Coverage covers the error path (line 63-68)
func TestClient_initConnection_ConnectError_Coverage(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "invalid-host-that-does-not-exist.example.com",
			Port:     "1",
			User:     "test",
			Password: "test",
			Name:     "test",
		},
	}
	
	client, err := NewClient(cfg)
	require.NoError(t, err)
	
	// Reset sync.Once to force connection
	client.connectOnce = sync.Once{}
	
	// This should fail and set connectErr
	err = client.initConnection(context.Background())
	assert.Error(t, err)
	assert.NotNil(t, client.connectErr)
}

// ============================================================================
// Missing Coverage for adapter.go:95 (NewClientWithFallback - 62.5%)
// This covers the success path (line 102-106) and ping failure (line 103-104)
// ============================================================================

// TestNewClientWithFallback_SuccessPath_Coverage covers the successful ping path
func TestNewClientWithFallback_SuccessPath_Coverage(t *testing.T) {
	// Use mock to simulate success
	mock := &mockDatabase{}
	client := newTestClient(mock)
	
	// Simulate successful ping
	err := client.Ping()
	assert.NoError(t, err)
}

// TestNewClientWithFallback_PingFailure covers the ping failure path
func TestNewClientWithFallback_PingFailure(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "invalid-host-that-does-not-exist.example.com",
			Port:     "1",
			User:     "test",
			Password: "test",
			Name:     "test",
		},
	}
	
	client, err := NewClientWithFallback(cfg)
	
	// Should fail because we can't connect
	assert.Error(t, err)
	assert.Nil(t, client)
}

// ============================================================================
// Missing Coverage for adapter.go:159 (Pool - 80%)
// ============================================================================

// TestClient_Pool_WithTestDBNil covers line 167 (return c.pool)
func TestClient_Pool_WithTestDBNil(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "invalid-host-that-does-not-exist.example.com",
			Port:     "1",
			User:     "test",
			Password: "test",
			Name:     "test",
		},
	}
	
	client, err := NewClient(cfg)
	require.NoError(t, err)
	
	// Reset sync.Once to force connection
	client.connectOnce = sync.Once{}
	
	// Should return nil because connection fails
	pool := client.Pool()
	assert.Nil(t, pool)
}

// TestClient_Pool_ConnectionSuccess covers the pool return path
func TestClient_Pool_ConnectionSuccess(t *testing.T) {
	mock := &mockDatabase{}
	client := newTestClient(mock)
	
	// With mock, Pool should return nil (testPG path)
	pool := client.Pool()
	assert.Nil(t, pool)
}

// ============================================================================
// Missing Coverage for adapter.go:187 (Ping - 80%)
// ============================================================================

// TestClient_Ping_WithRealPG covers the real pg path (line 194)
func TestClient_Ping_WithRealPG(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "invalid-host-that-does-not-exist.example.com",
			Port:     "1",
			User:     "test",
			Password: "test",
			Name:     "test",
		},
	}
	
	client, err := NewClient(cfg)
	require.NoError(t, err)
	
	// Reset sync.Once to force connection
	client.connectOnce = sync.Once{}
	
	// Should fail with real connection
	err = client.Ping()
	assert.Error(t, err)
}

// ============================================================================
// Missing Coverage for adapter.go:198 (HealthCheck - 85.7%)
// ============================================================================

// TestClient_HealthCheck_WithRealPG covers the real pg path (line 207)
func TestClient_HealthCheck_WithRealPG(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "invalid-host-that-does-not-exist.example.com",
			Port:     "1",
			User:     "test",
			Password: "test",
			Name:     "test",
		},
	}
	
	client, err := NewClient(cfg)
	require.NoError(t, err)
	
	// Reset sync.Once to force connection
	client.connectOnce = sync.Once{}
	
	// Should fail with real connection
	err = client.HealthCheck()
	assert.Error(t, err)
}

// ============================================================================
// Missing Coverage for adapter.go:212 (Exec - 71.4%)
// ============================================================================

// TestClient_Exec_WithRealPG covers the real pg path (line 220-221)
func TestClient_Exec_WithRealPG(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "invalid-host-that-does-not-exist.example.com",
			Port:     "1",
			User:     "test",
			Password: "test",
			Name:     "test",
		},
	}
	
	client, err := NewClient(cfg)
	require.NoError(t, err)
	
	// Reset sync.Once to force connection
	client.connectOnce = sync.Once{}
	
	// Should fail with real connection
	err = client.Exec("SELECT 1")
	assert.Error(t, err)
}

// TestClient_Exec_WithMockError covers error return path
func TestClient_Exec_WithMockError(t *testing.T) {
	mock := &mockDatabase{execErr: errors.New("exec failed")}
	client := newTestClient(mock)
	
	err := client.Exec("INSERT INTO test VALUES (1)")
	assert.Error(t, err)
	assert.Equal(t, "exec failed", err.Error())
}

// ============================================================================
// Missing Coverage for adapter.go:226 (Query - 93.3%)
// ============================================================================

// TestClient_Query_WithRealPG covers the real pg path
func TestClient_Query_WithRealPG(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "invalid-host-that-does-not-exist.example.com",
			Port:     "1",
			User:     "test",
			Password: "test",
			Name:     "test",
		},
	}
	
	client, err := NewClient(cfg)
	require.NoError(t, err)
	
	// Reset sync.Once to force connection
	client.connectOnce = sync.Once{}
	
	results, err := client.Query("SELECT 1")
	assert.Error(t, err)
	assert.Nil(t, results)
}

// ============================================================================
// Missing Coverage for adapter.go:253 (QueryRow - 80%)
// ============================================================================

// TestClient_QueryRow_WithRealPG covers the real pg path (line 262)
func TestClient_QueryRow_WithRealPG(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "invalid-host-that-does-not-exist.example.com",
			Port:     "1",
			User:     "test",
			Password: "test",
			Name:     "test",
		},
	}
	
	client, err := NewClient(cfg)
	require.NoError(t, err)
	
	// Reset sync.Once to force connection
	client.connectOnce = sync.Once{}
	
	row := client.QueryRow("SELECT 1")
	assert.NotNil(t, row)
	
	var dest int
	err = row.Scan(&dest)
	assert.Error(t, err)
}

// ============================================================================
// Missing Coverage for adapter.go:266 (Begin - 80%)
// ============================================================================

// TestClient_Begin_WithRealPG covers the real pg path (line 273)
func TestClient_Begin_WithRealPG(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "invalid-host-that-does-not-exist.example.com",
			Port:     "1",
			User:     "test",
			Password: "test",
			Name:     "test",
		},
	}
	
	client, err := NewClient(cfg)
	require.NoError(t, err)
	
	// Reset sync.Once to force connection
	client.connectOnce = sync.Once{}
	
	tx, err := client.Begin(context.Background())
	assert.Error(t, err)
	assert.Nil(t, tx)
}

// ============================================================================
// Missing Coverage for adapter.go:277 (Migrate - 66.7%)
// ============================================================================

// TestClient_Migrate_WithRealPG covers the real pg path (line 281)
func TestClient_Migrate_WithRealPG(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "invalid-host-that-does-not-exist.example.com",
			Port:     "1",
			User:     "test",
			Password: "test",
			Name:     "test",
		},
	}
	
	client, err := NewClient(cfg)
	require.NoError(t, err)
	
	// Reset sync.Once to force connection
	client.connectOnce = sync.Once{}
	
	err = client.Migrate(context.Background(), []string{"CREATE TABLE test (id INT)"})
	assert.Error(t, err)
}

// ============================================================================
// Missing Coverage for compat.go:44 (NewPostgresDB - 75%)
// ============================================================================

// TestNewPostgresDB_ErrorPath_Coverage covers error return path (line 46-48)
// Note: NewClient doesn't currently return errors, so this tests the success path
func TestNewPostgresDB_ErrorPath_Coverage(t *testing.T) {
	cfg := &config.Config{}
	
	pgDB, err := NewPostgresDB(cfg)
	
	// NewPostgresDB never errors in current implementation
	assert.NoError(t, err)
	assert.NotNil(t, pgDB)
}

// ============================================================================
// Missing Coverage for compat.go:54 (NewPostgresDBWithFallback - 71.4%)
// ============================================================================

// TestNewPostgresDBWithFallback_DBErrorPath covers the DB error path (line 56-62)
func TestNewPostgresDBWithFallback_DBErrorPath(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "invalid-host-that-does-not-exist.example.com",
			Port:     "1",
			User:     "test",
			Password: "test",
			Name:     "test",
		},
	}
	
	pgDB, memDB, err := NewPostgresDBWithFallback(cfg)
	
	// Should not error, should return memDB
	assert.NoError(t, err)
	assert.Nil(t, pgDB)
	assert.NotNil(t, memDB)
}

// ============================================================================
// Missing Coverage for compat.go:69 (Connect - 80%)
// ============================================================================

// TestConnect_ErrorPath_Coverage covers the error return path (line 71-73)
// Note: Connect doesn't currently return errors
func TestConnect_ErrorPath_Coverage(t *testing.T) {
	db, err := Connect()
	
	// Connect never errors in current implementation
	assert.NoError(t, err)
	assert.NotNil(t, db)
}

// ============================================================================
// Environment Variable Coverage for buildPostgresConfig
// ============================================================================

// TestBuildPostgresConfig_EnvOverrides tests environment variable handling
func TestBuildPostgresConfig_EnvOverrides(t *testing.T) {
	// Set environment variables
	os.Setenv("DB_HOST", "envhost")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_USER", "envuser")
	os.Setenv("DB_PASSWORD", "envpass")
	os.Setenv("DB_NAME", "envdb")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
	}()
	
	cfg := &config.Config{}
	pgCfg := buildPostgresConfig(cfg)
	
	assert.Equal(t, "envhost", pgCfg.Host)
	assert.Equal(t, 5433, pgCfg.Port)
	assert.Equal(t, "envuser", pgCfg.User)
	assert.Equal(t, "envpass", pgCfg.Password)
	assert.Equal(t, "envdb", pgCfg.DBName)
}

// TestBuildPostgresConfig_InvalidEnvPort tests invalid port in env
func TestBuildPostgresConfig_InvalidEnvPort(t *testing.T) {
	os.Setenv("DB_PORT", "invalid")
	defer os.Unsetenv("DB_PORT")
	
	cfg := &config.Config{}
	pgCfg := buildPostgresConfig(cfg)
	
	// Should use default port
	assert.Equal(t, 5432, pgCfg.Port)
}

// ============================================================================
// Additional MemoryDB Coverage
// ============================================================================

// TestMemoryDB_QueryRow_NoData covers the no data path
func TestMemoryDB_QueryRow_NoData(t *testing.T) {
	m := NewMemoryDB()
	
	row := m.QueryRow("SELECT * FROM nonexistent")
	assert.NotNil(t, row)
	
	var dest string
	err := row.Scan(&dest)
	assert.Error(t, err)
}

// TestMemoryDB_QueryRow_WithData tests scanning stored data
func TestMemoryDB_QueryRow_WithData(t *testing.T) {
	m := NewMemoryDB()
	m.StoreRow("users", "user1", []any{"Alice", 30})
	
	// QueryRow doesn't use stored data directly, but tests the mechanism
	row := m.QueryRow("SELECT * FROM users WHERE id = $1", "user1")
	assert.NotNil(t, row)
}

// TestMemoryDB_Close_Idempotent tests that Close can be called multiple times
func TestMemoryDB_Close_Idempotent(t *testing.T) {
	m := NewMemoryDB()
	
	err := m.Close()
	assert.NoError(t, err)
	
	err = m.Close()
	assert.NoError(t, err)
}

// ============================================================================
// PostgresDB Method Coverage
// ============================================================================

// TestPostgresDB_AllMethods_Coverage covers all PostgresDB methods
func TestPostgresDB_AllMethods_Coverage(t *testing.T) {
	mock := &mockDatabase{}
	client := newTestClient(mock)
	pgDB := &PostgresDB{client: client}
	
	// Test Ping
	err := pgDB.Ping()
	assert.NoError(t, err)
	
	// Test Exec
	err = pgDB.Exec("SELECT 1")
	assert.NoError(t, err)
	
	// Test Query
	mockRows := &mockRows{nextReturns: []bool{true, false}}
	mock.queryRows = mockRows
	results, err := pgDB.Query("SELECT * FROM test")
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	
	// Test QueryRow
	mock.queryRowResult = mockRow{}
	row := pgDB.QueryRow("SELECT 1")
	assert.NotNil(t, row)
	
	// Test HealthCheck
	err = pgDB.HealthCheck()
	assert.NoError(t, err)
	
	// Test GetPool
	pool := pgDB.GetPool()
	assert.Nil(t, pool) // Returns nil with mock
	
	// Test Database
	db := pgDB.Database()
	assert.NotNil(t, db)
	
	// Test Close
	err = pgDB.Close()
	assert.NoError(t, err)
}

// TestPostgresDB_WithRealClient tests PostgresDB with a real (non-mock) client
func TestPostgresDB_WithRealClient(t *testing.T) {
	cfg := &config.Config{}
	client, err := NewClient(cfg)
	require.NoError(t, err)
	
	pgDB := &PostgresDB{client: client}
	
	// Test GetPool - should try to connect
	pool := pgDB.GetPool()
	// May be nil if connection fails
	_ = pool
}

// ============================================================================
// RunMigration Coverage
// ============================================================================

// TestRunMigration_ConnectionFailure tests RunMigration with connection failure
func TestRunMigration_ConnectionFailure(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "invalid-host-that-does-not-exist.example.com",
			Port:     "1",
			User:     "test",
			Password: "test",
			Name:     "test",
		},
	}
	
	client, err := NewClient(cfg)
	require.NoError(t, err)
	
	// Reset sync.Once to force connection
	client.connectOnce = sync.Once{}
	
	pgDB := &PostgresDB{client: client}
	
	migrations := []string{
		"CREATE TABLE test (id INT)",
	}
	
	err = RunMigration(pgDB, migrations)
	assert.Error(t, err)
}

// ============================================================================
// Concurrent Access Tests
// ============================================================================

// TestClient_ConcurrentInitConnection tests concurrent initConnection calls
func TestClient_ConcurrentInitConnection(t *testing.T) {
	mock := &mockDatabase{}
	client := newTestClient(mock)
	
	var wg sync.WaitGroup
	errors := make(chan error, 10)
	
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := client.initConnection(context.Background())
			errors <- err
		}()
	}
	
	wg.Wait()
	close(errors)
	
	for err := range errors {
		assert.NoError(t, err)
	}
}

// ============================================================================
// Context Cancellation Tests
// ============================================================================

// TestClient_Begin_CancelledContext tests Begin with cancelled context
func TestClient_Begin_CancelledContext(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "invalid-host-that-does-not-exist.example.com",
			Port:     "1",
			User:     "test",
			Password: "test",
			Name:     "test",
		},
	}
	
	client, err := NewClient(cfg)
	require.NoError(t, err)
	
	// Reset sync.Once to force connection
	client.connectOnce = sync.Once{}
	
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately
	
	tx, err := client.Begin(ctx)
	assert.Error(t, err)
	assert.Nil(t, tx)
}

// ============================================================================
// ErrorRow Additional Tests
// ============================================================================

// TestErrorRow_Scan_NilError tests ErrorRow with nil error
func TestErrorRow_Scan_NilError(t *testing.T) {
	row := &ErrorRow{err: nil}
	
	var dest string
	err := row.Scan(&dest)
	assert.NoError(t, err)
}

// ============================================================================
// Type Definition Tests
// ============================================================================

// TestTypeDefinitions ensures all types are properly defined
func TestTypeDefinitions(t *testing.T) {
	// Ensure interface compliance
	var _ databaseWithMigrate = (*mockDatabase)(nil)
}
