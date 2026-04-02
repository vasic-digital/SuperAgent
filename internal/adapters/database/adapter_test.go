package database

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	db "digital.vasic.database/pkg/database"

	"dev.helix.agent/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Mock Types
// ============================================================================

type mockDatabase struct {
	mock.Mock
}

func (m *mockDatabase) Exec(ctx context.Context, query string, args ...any) (db.Result, error) {
	callArgs := m.Called(ctx, query, args)
	return callArgs.Get(0).(db.Result), callArgs.Error(1)
}

func (m *mockDatabase) Query(ctx context.Context, query string, args ...any) (db.Rows, error) {
	callArgs := m.Called(ctx, query, args)
	if callArgs.Get(0) == nil {
		return nil, callArgs.Error(1)
	}
	return callArgs.Get(0).(db.Rows), callArgs.Error(1)
}

func (m *mockDatabase) QueryRow(ctx context.Context, query string, args ...any) db.Row {
	callArgs := m.Called(ctx, query, args)
	return callArgs.Get(0).(db.Row)
}

func (m *mockDatabase) Begin(ctx context.Context) (db.Tx, error) {
	callArgs := m.Called(ctx)
	if callArgs.Get(0) == nil {
		return nil, callArgs.Error(1)
	}
	return callArgs.Get(0).(db.Tx), callArgs.Error(1)
}

func (m *mockDatabase) Close() error {
	return m.Called().Error(0)
}

func (m *mockDatabase) HealthCheck(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

func (m *mockDatabase) Migrate(ctx context.Context, migrations []string) error {
	return m.Called(ctx, migrations).Error(0)
}

type mockRow struct {
	mock.Mock
}

func (m *mockRow) Scan(dest ...any) error {
	return m.Called(dest).Error(0)
}

type mockRows struct {
	mock.Mock
	nextCount int
	maxNext   int
}

func (m *mockRows) Next() bool {
	if m.nextCount < m.maxNext {
		m.nextCount++
		return true
	}
	return false
}

func (m *mockRows) Scan(dest ...any) error {
	return m.Called(dest).Error(0)
}

func (m *mockRows) Close() error {
	return m.Called().Error(0)
}

func (m *mockRows) Err() error {
	return m.Called().Error(0)
}

type mockTx struct {
	mock.Mock
}

func (m *mockTx) Commit(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

func (m *mockTx) Rollback(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

func (m *mockTx) Exec(ctx context.Context, query string, args ...any) (db.Result, error) {
	callArgs := m.Called(ctx, query, args)
	return callArgs.Get(0).(db.Result), callArgs.Error(1)
}

func (m *mockTx) Query(ctx context.Context, query string, args ...any) (db.Rows, error) {
	callArgs := m.Called(ctx, query, args)
	return callArgs.Get(0).(db.Rows), callArgs.Error(1)
}

func (m *mockTx) QueryRow(ctx context.Context, query string, args ...any) db.Row {
	callArgs := m.Called(ctx, query, args)
	return callArgs.Get(0).(db.Row)
}

// ============================================================================
// ErrorRow Tests
// ============================================================================

func TestErrorRow_Scan(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{
			name:    "returns connection error",
			err:     errors.New("connection failed"),
			wantErr: true,
		},
		{
			name:    "returns timeout error",
			err:     context.DeadlineExceeded,
			wantErr: true,
		},
		{
			name:    "returns nil for nil error",
			err:     nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := &ErrorRow{err: tt.err}
			var dest string
			err := row.Scan(&dest)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// NewClient Tests
// ============================================================================

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *config.Config
		envVars     map[string]string
		expectError bool
	}{
		{
			name: "with config values",
			cfg: &config.Config{
				Database: config.DatabaseConfig{
					Host:     "localhost",
					Port:     "5432",
					User:     "testuser",
					Password: "testpass",
					Name:     "testdb",
					SSLMode:  "disable",
				},
			},
			expectError: false,
		},
		{
			name: "with empty config uses defaults",
			cfg:  &config.Config{},
			expectError: false,
		},
		{
			name:    "with environment variables",
			cfg:     &config.Config{},
			envVars: map[string]string{
				"DB_HOST":     "envhost",
				"DB_PORT":     "5433",
				"DB_USER":     "envuser",
				"DB_PASSWORD": "envpass",
				"DB_NAME":     "envdb",
			},
			expectError: false,
		},
		{
			name: "config overrides environment",
			cfg: &config.Config{
				Database: config.DatabaseConfig{
					Host: "confighost",
					Port: "5434",
				},
			},
			envVars: map[string]string{
				"DB_HOST": "envhost",
				"DB_PORT": "5433",
			},
			expectError: false,
		},
		{
			name: "invalid port in config",
			cfg: &config.Config{
				Database: config.DatabaseConfig{
					Port: "invalid",
				},
			},
			expectError: false,
		},
		{
			name: "invalid port in environment",
			cfg:  &config.Config{},
			envVars: map[string]string{
				"DB_PORT": "invalid",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set env vars
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			client, err := NewClient(tt.cfg)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

// ============================================================================
// initConnection Tests
// ============================================================================

func TestClient_initConnection_WithDeadline(t *testing.T) {
	mockDB := new(mockDatabase)
	client := &Client{
		testPG: mockDB,
	}

	// Context with deadline - should not add timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.initConnection(ctx)
	assert.NoError(t, err)
}

func TestClient_initConnection_WithoutDeadline(t *testing.T) {
	mockDB := new(mockDatabase)
	client := &Client{
		testPG: mockDB,
	}

	// Context without deadline - should add timeout
	ctx := context.Background()

	err := client.initConnection(ctx)
	assert.NoError(t, err)
}

func TestClient_initConnection_SyncOnce(t *testing.T) {
	mockDB := new(mockDatabase)
	client := &Client{
		testPG: mockDB,
	}

	// First call
	err1 := client.initConnection(context.Background())
	assert.NoError(t, err1)

	// Second call - should be idempotent due to sync.Once
	err2 := client.initConnection(context.Background())
	assert.NoError(t, err2)
}

func TestClient_initConnection_Concurrent(t *testing.T) {
	mockDB := new(mockDatabase)
	client := &Client{
		testPG: mockDB,
	}

	var wg sync.WaitGroup
	errors := make(chan error, 10)

	// Call initConnection concurrently
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

	// All should succeed
	for err := range errors {
		assert.NoError(t, err)
	}
}

// ============================================================================
// ensureConnected Tests
// ============================================================================

func TestClient_ensureConnected(t *testing.T) {
	mockDB := new(mockDatabase)
	client := &Client{
		testPG: mockDB,
	}

	err := client.ensureConnected()
	assert.NoError(t, err)
}

// ============================================================================
// Pool Tests
// ============================================================================

func TestClient_Pool_NotConnected(t *testing.T) {
	// Client without testPG - should try to connect and fail
	cfg := &config.Config{}
	client, _ := NewClient(cfg)

	pool := client.Pool()
	assert.Nil(t, pool)
}

func TestClient_Pool_WithTestDB(t *testing.T) {
	mockDB := new(mockDatabase)
	client := &Client{
		testPG: mockDB,
	}

	// Should return nil for test pool
	pool := client.Pool()
	assert.Nil(t, pool)
}

// ============================================================================
// Database Tests
// ============================================================================

func TestClient_Database_WithTestDB(t *testing.T) {
	mockDB := new(mockDatabase)
	client := &Client{
		testPG: mockDB,
	}

	db := client.Database()
	assert.Equal(t, mockDB, db)
}

func TestClient_Database_WithRealPG(t *testing.T) {
	cfg := &config.Config{}
	client, err := NewClient(cfg)
	require.NoError(t, err)

	db := client.Database()
	assert.NotNil(t, db)
}

// ============================================================================
// Close Tests
// ============================================================================

func TestClient_Close_WithTestDB(t *testing.T) {
	mockDB := new(mockDatabase)
	mockDB.On("Close").Return(nil).Once()

	client := &Client{
		testPG: mockDB,
	}

	err := client.Close()
	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
}

func TestClient_Close_WithTestDB_Error(t *testing.T) {
	mockDB := new(mockDatabase)
	mockDB.On("Close").Return(errors.New("close failed")).Once()

	client := &Client{
		testPG: mockDB,
	}

	err := client.Close()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "close failed")
}

// ============================================================================
// Ping Tests
// ============================================================================

func TestClient_Ping_NotConnected(t *testing.T) {
	cfg := &config.Config{}
	client, _ := NewClient(cfg)

	// Should fail because not connected to real DB
	err := client.Ping()
	assert.Error(t, err)
}

func TestClient_Ping_WithTestDB(t *testing.T) {
	mockDB := new(mockDatabase)
	mockDB.On("HealthCheck", mock.Anything).Return(nil).Once()

	client := &Client{
		testPG: mockDB,
	}

	err := client.Ping()
	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
}

func TestClient_Ping_WithTestDB_Error(t *testing.T) {
	mockDB := new(mockDatabase)
	mockDB.On("HealthCheck", mock.Anything).Return(errors.New("ping failed")).Once()

	client := &Client{
		testPG: mockDB,
	}

	err := client.Ping()
	assert.Error(t, err)
}

// ============================================================================
// HealthCheck Tests
// ============================================================================

func TestClient_HealthCheck_NotConnected(t *testing.T) {
	cfg := &config.Config{}
	client, _ := NewClient(cfg)

	err := client.HealthCheck()
	assert.Error(t, err)
}

func TestClient_HealthCheck_WithTestDB(t *testing.T) {
	mockDB := new(mockDatabase)
	mockDB.On("HealthCheck", mock.Anything).Return(nil).Once()

	client := &Client{
		testPG: mockDB,
	}

	err := client.HealthCheck()
	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
}

func TestClient_HealthCheck_WithTestDB_Error(t *testing.T) {
	mockDB := new(mockDatabase)
	mockDB.On("HealthCheck", mock.Anything).Return(errors.New("health check failed")).Once()

	client := &Client{
		testPG: mockDB,
	}

	err := client.HealthCheck()
	assert.Error(t, err)
}

// ============================================================================
// Exec Tests
// ============================================================================

func TestClient_Exec_NotConnected(t *testing.T) {
	cfg := &config.Config{}
	client, _ := NewClient(cfg)

	err := client.Exec("SELECT 1")
	assert.Error(t, err)
}

func TestClient_Exec_WithTestDB(t *testing.T) {
	mockDB := new(mockDatabase)
	mockDB.On("Exec", mock.Anything, "INSERT INTO users VALUES ($1)", mock.Anything).Return(nil, nil).Once()

	client := &Client{
		testPG: mockDB,
	}

	err := client.Exec("INSERT INTO users VALUES ($1)", "alice")
	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
}

func TestClient_Exec_WithTestDB_Error(t *testing.T) {
	mockDB := new(mockDatabase)
	mockDB.On("Exec", mock.Anything, "INVALID SQL", mock.Anything).Return(nil, errors.New("syntax error")).Once()

	client := &Client{
		testPG: mockDB,
	}

	err := client.Exec("INVALID SQL")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "syntax error")
}

// ============================================================================
// Query Tests
// ============================================================================

func TestClient_Query_NotConnected(t *testing.T) {
	cfg := &config.Config{}
	client, _ := NewClient(cfg)

	results, err := client.Query("SELECT * FROM users")
	assert.Error(t, err)
	assert.Nil(t, results)
}

func TestClient_Query_WithTestDB(t *testing.T) {
	mockDB := new(mockDatabase)
	mockRows := &mockRows{maxNext: 2}
	mockRows.On("Close").Return(nil).Once()
	mockRows.On("Err").Return(nil).Once()
	mockDB.On("Query", mock.Anything, "SELECT * FROM users", mock.Anything).Return(mockRows, nil).Once()

	client := &Client{
		testPG: mockDB,
	}

	results, err := client.Query("SELECT * FROM users")
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Len(t, results, 2)
	mockDB.AssertExpectations(t)
}

func TestClient_Query_WithTestDB_QueryError(t *testing.T) {
	mockDB := new(mockDatabase)
	mockDB.On("Query", mock.Anything, "INVALID", mock.Anything).Return(nil, errors.New("syntax error")).Once()

	client := &Client{
		testPG: mockDB,
	}

	results, err := client.Query("INVALID")
	assert.Error(t, err)
	assert.Nil(t, results)
}

func TestClient_Query_WithTestDB_RowsError(t *testing.T) {
	mockDB := new(mockDatabase)
	mockRows := &mockRows{maxNext: 0}
	mockRows.On("Close").Return(nil).Once()
	mockRows.On("Err").Return(errors.New("rows error")).Once()
	mockDB.On("Query", mock.Anything, "SELECT * FROM users", mock.Anything).Return(mockRows, nil).Once()

	client := &Client{
		testPG: mockDB,
	}

	results, err := client.Query("SELECT * FROM users")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rows error")
	assert.NotNil(t, results)
}

// ============================================================================
// QueryRow Tests
// ============================================================================

func TestClient_QueryRow_NotConnected(t *testing.T) {
	cfg := &config.Config{}
	client, _ := NewClient(cfg)

	row := client.QueryRow("SELECT 1")
	assert.NotNil(t, row)

	var result int
	err := row.Scan(&result)
	assert.Error(t, err)
}

func TestClient_QueryRow_WithTestDB(t *testing.T) {
	mockDB := new(mockDatabase)
	mockRow := new(mockRow)
	mockRow.On("Scan", mock.Anything).Return(nil).Once()
	mockDB.On("QueryRow", mock.Anything, "SELECT id FROM users WHERE name = $1", mock.Anything).Return(mockRow).Once()

	client := &Client{
		testPG: mockDB,
	}

	row := client.QueryRow("SELECT id FROM users WHERE name = $1", "alice")
	assert.NotNil(t, row)

	var id int
	err := row.Scan(&id)
	assert.NoError(t, err)
}

// ============================================================================
// Begin Tests
// ============================================================================

func TestClient_Begin_NotConnected(t *testing.T) {
	cfg := &config.Config{}
	client, _ := NewClient(cfg)

	tx, err := client.Begin(context.Background())
	assert.Error(t, err)
	assert.Nil(t, tx)
}

func TestClient_Begin_WithTestDB(t *testing.T) {
	mockDB := new(mockDatabase)
	mockTx := new(mockTx)
	mockDB.On("Begin", mock.Anything).Return(mockTx, nil).Once()

	client := &Client{
		testPG: mockDB,
	}

	tx, err := client.Begin(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, tx)
	mockDB.AssertExpectations(t)
}

func TestClient_Begin_WithTestDB_Error(t *testing.T) {
	mockDB := new(mockDatabase)
	mockDB.On("Begin", mock.Anything).Return(nil, errors.New("begin failed")).Once()

	client := &Client{
		testPG: mockDB,
	}

	tx, err := client.Begin(context.Background())
	assert.Error(t, err)
	assert.Nil(t, tx)
}

// ============================================================================
// Migrate Tests
// ============================================================================

func TestClient_Migrate_NotConnected(t *testing.T) {
	cfg := &config.Config{}
	client, _ := NewClient(cfg)

	migrations := []string{
		"CREATE TABLE test (id INT)",
	}
	err := client.Migrate(context.Background(), migrations)
	assert.Error(t, err)
}

func TestClient_Migrate_Success(t *testing.T) {
	// This test would need a real database connection or more complex mocking
	// For now, we test the not-connected case
}

// ============================================================================
// buildPostgresConfig Tests
// ============================================================================

func TestBuildPostgresConfig(t *testing.T) {
	tests := []struct {
		name           string
		cfg            *config.Config
		envVars        map[string]string
		expectedHost   string
		expectedPort   int
		expectedUser   string
		expectedDBName string
	}{
		{
			name: "all config values",
			cfg: &config.Config{
				Database: config.DatabaseConfig{
					Host:     "myhost",
					Port:     "5432",
					User:     "myuser",
					Password: "mypass",
					Name:     "mydb",
					SSLMode:  "require",
				},
			},
			expectedHost:   "myhost",
			expectedPort:   5432,
			expectedUser:   "myuser",
			expectedDBName: "mydb",
		},
		{
			name:           "defaults when empty",
			cfg:            &config.Config{},
			expectedHost:   "localhost",
			expectedPort:   5432,
			expectedUser:   "postgres",
			expectedDBName: "postgres",
		},
		{
			name: "environment variables",
			cfg:  &config.Config{},
			envVars: map[string]string{
				"DB_HOST": "envhost",
				"DB_PORT": "5433",
				"DB_USER": "envuser",
				"DB_NAME": "envdb",
			},
			expectedHost:   "envhost",
			expectedPort:   5433,
			expectedUser:   "envuser",
			expectedDBName: "envdb",
		},
		{
			name: "config overrides env",
			cfg: &config.Config{
				Database: config.DatabaseConfig{
					Host: "confighost",
					Port: "5434",
				},
			},
			envVars: map[string]string{
				"DB_HOST": "envhost",
				"DB_PORT": "5433",
			},
			expectedHost: "confighost",
			expectedPort: 5434,
		},
		{
			name: "invalid port falls back to default",
			cfg: &config.Config{
				Database: config.DatabaseConfig{
					Port: "invalid",
				},
			},
			expectedPort: 5432,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set env vars
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			pgCfg := buildPostgresConfig(tt.cfg)
			assert.Equal(t, tt.expectedHost, pgCfg.Host)
			assert.Equal(t, tt.expectedPort, pgCfg.Port)
			if tt.expectedUser != "" {
				assert.Equal(t, tt.expectedUser, pgCfg.User)
			}
			if tt.expectedDBName != "" {
				assert.Equal(t, tt.expectedDBName, pgCfg.DBName)
			}
			assert.Equal(t, "helixagent", pgCfg.ApplicationName)
		})
	}
}

// ============================================================================
// NewClientWithFallback Tests
// ============================================================================

func TestNewClientWithFallback_Success(t *testing.T) {
	// Use an invalid config to trigger fallback behavior
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "invalid-host-that-does-not-exist",
			Port:     "59999",
			User:     "test",
			Password: "test",
			Name:     "test",
		},
	}

	client, err := NewClientWithFallback(cfg)
	// Should fail to connect and return nil
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestNewClientWithFallback_NewClientError(t *testing.T) {
	// NewClient should not error with valid config, just connection
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host: "localhost",
			Port: "5432",
		},
	}

	client, err := NewClientWithFallback(cfg)
	// Should fail on ping since no real DB
	assert.Error(t, err)
	assert.Nil(t, client)
}

// ============================================================================
// Type Alias Tests
// ============================================================================

func TestTypeAliases(t *testing.T) {
	// Verify type aliases compile correctly
	var _ Row = (db.Row)(nil)
	var _ Rows = (db.Rows)(nil)
	var _ Tx = (db.Tx)(nil)
	var _ Result = (db.Result)(nil)
}

// ============================================================================
// Helper Functions for Tests
// ============================================================================

func setupMockClient(t *testing.T) (*Client, *mockDatabase) {
	mockDB := new(mockDatabase)
	client := &Client{
		testPG: mockDB,
	}
	return client, mockDB
}

// ============================================================================
// Result Mock
// ============================================================================

type mockResult struct {
	mock.Mock
}

func (m *mockResult) LastInsertId() (int64, error) {
	return 0, fmt.Errorf("not supported")
}

func (m *mockResult) RowsAffected() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}
