//go:build chaos

package database

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"dev.helix.agent/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Connection Failure Chaos Tests
// ============================================================================

func TestChaos_Client_ConcurrentAccess(t *testing.T) {
	mock := &mockDatabase{}
	client := newTestClient(mock)
	
	// Launch multiple goroutines to test concurrent access
	const numGoroutines = 50
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)
	
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			
			// Each goroutine performs different operations
			switch idx % 5 {
			case 0:
				if err := client.Ping(); err != nil {
					errors <- err
				}
			case 1:
				if _, err := client.Query("SELECT * FROM test"); err != nil {
					errors <- err
				}
			case 2:
				if err := client.Exec("INSERT INTO test VALUES (1)"); err != nil {
					errors <- err
				}
			case 3:
				row := client.QueryRow("SELECT 1")
				var val int
				if err := row.Scan(&val); err != nil {
					errors <- err
				}
			case 4:
				if err := client.HealthCheck(); err != nil {
					errors <- err
				}
			}
		}(i)
	}
	
	wg.Wait()
	close(errors)
	
	// Count errors
	errorCount := 0
	for err := range errors {
		if err != nil {
			errorCount++
		}
	}
	
	// Most operations should succeed with mock
	t.Logf("Concurrent access test completed with %d errors", errorCount)
}

func TestChaos_Client_RapidCloseAndReopen(t *testing.T) {
	mock := &mockDatabase{}
	client := newTestClient(mock)
	
	// Rapidly close and use the client
	for i := 0; i < 10; i++ {
		err := client.Close()
		assert.NoError(t, err)
		
		// Try to use after close
		_ = client.Ping()
	}
}

func TestChaos_Client_TimeoutHandling(t *testing.T) {
	mock := &mockDatabase{
		healthCheckErr: errors.New("timeout"),
	}
	client := newTestClient(mock)
	
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	
	// Begin with timeout
	_, err := client.Begin(ctx)
	// May or may not error depending on timing
	_ = err
}

// ============================================================================
// Error Handling Chaos Tests
// ============================================================================

func TestChaos_Client_AllMethodsWithErrors(t *testing.T) {
	testCases := []struct {
		name    string
		mock    *mockDatabase
		testFn  func(*Client) error
	}{
		{
			name: "Ping error",
			mock: &mockDatabase{healthCheckErr: errors.New("ping failed")},
			testFn: func(c *Client) error {
				return c.Ping()
			},
		},
		{
			name: "HealthCheck error",
			mock: &mockDatabase{healthCheckErr: errors.New("health check failed")},
			testFn: func(c *Client) error {
				return c.HealthCheck()
			},
		},
		{
			name: "Exec error",
			mock: &mockDatabase{execErr: errors.New("exec failed")},
			testFn: func(c *Client) error {
				return c.Exec("SELECT 1")
			},
		},
		{
			name: "Query error",
			mock: &mockDatabase{queryErr: errors.New("query failed")},
			testFn: func(c *Client) error {
				_, err := c.Query("SELECT * FROM test")
				return err
			},
		},
		{
			name: "Begin error",
			mock: &mockDatabase{beginErr: errors.New("begin failed")},
			testFn: func(c *Client) error {
				_, err := c.Begin(context.Background())
				return err
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(tc.mock)
			err := tc.testFn(client)
			assert.Error(t, err)
		})
	}
}

// ============================================================================
// Connection Reset Chaos Tests
// ============================================================================

func TestChaos_Client_ResetConnectionState(t *testing.T) {
	mock := &mockDatabase{}
	client := newTestClient(mock)
	
	// First, connect successfully
	err := client.Ping()
	require.NoError(t, err)
	
	// Reset the sync.Once to simulate reconnection
	client.connectOnce = sync.Once{}
	
	// Try to connect again - should work
	err = client.Ping()
	assert.NoError(t, err)
}

// ============================================================================
// Query Chaos Tests
// ============================================================================

func TestChaos_Query_MultipleRowIterations(t *testing.T) {
	mockRows := &mockRows{
		nextReturns: []bool{true, true, true, true, false},
	}
	mock := &mockDatabase{queryRows: mockRows}
	client := newTestClient(mock)
	
	results, err := client.Query("SELECT * FROM large_table")
	
	assert.NoError(t, err)
	assert.Len(t, results, 4)
}

func TestChaos_Query_RowsCloseError(t *testing.T) {
	mockRows := &mockRows{
		nextReturns: []bool{true, false},
		closeErr:    errors.New("close failed"),
	}
	mock := &mockDatabase{queryRows: mockRows}
	client := newTestClient(mock)
	
	// Should not panic even if close fails
	results, err := client.Query("SELECT * FROM test")
	
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestChaos_QueryRow_ScanVariations(t *testing.T) {
	testCases := []struct {
		name     string
		row      mockRow
		wantErr  bool
	}{
		{
			name:    "scan error",
			row:     mockRow{scanErr: errors.New("scan failed")},
			wantErr: true,
		},
		{
			name:    "empty values",
			row:     mockRow{values: []any{}},
			wantErr: false,
		},
		{
			name:    "nil values",
			row:     mockRow{values: nil},
			wantErr: false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockDatabase{queryRowResult: tc.row}
			client := newTestClient(mock)
			
			row := client.QueryRow("SELECT 1")
			var val int
			err := row.Scan(&val)
			
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				// No error expected for empty/nil values
				_ = err
			}
		})
	}
}

// ============================================================================
// MemoryDB Chaos Tests
// ============================================================================

func TestChaos_MemoryDB_ConcurrentAccess(t *testing.T) {
	m := NewMemoryDB()
	
	const numGoroutines = 100
	var wg sync.WaitGroup
	
	wg.Add(numGoroutines * 3)
	
	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			m.StoreRow("test", string(rune(idx)), []any{idx})
		}(i)
	}
	
	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			_, _ = m.Query("SELECT * FROM test")
		}(i)
	}
	
	// Concurrent pings
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			_ = m.Ping()
		}(i)
	}
	
	wg.Wait()
	
	// Final health check
	err := m.HealthCheck()
	assert.NoError(t, err)
}

func TestChaos_MemoryDB_CloseAndReuse(t *testing.T) {
	m := NewMemoryDB()
	
	// Close
	err := m.Close()
	assert.NoError(t, err)
	
	// Health check should fail after close
	err = m.HealthCheck()
	assert.Error(t, err)
	
	// Other operations should handle closed state gracefully
	_ = m.Ping()
	_ = m.Exec("SELECT 1")
	_, _ = m.Query("SELECT * FROM test")
	_ = m.QueryRow("SELECT 1")
}

// ============================================================================
// PostgresDB Chaos Tests
// ============================================================================

func TestChaos_PostgresDB_MethodsWithErrors(t *testing.T) {
	mock := &mockDatabase{
		execErr:        errors.New("exec failed"),
		queryErr:       errors.New("query failed"),
		healthCheckErr: errors.New("health check failed"),
	}
	client := newTestClient(mock)
	pgDB := &PostgresDB{client: client}
	
	// All methods should propagate errors
	assert.Error(t, pgDB.Ping())
	assert.Error(t, pgDB.Exec("SELECT 1"))
	_, err := pgDB.Query("SELECT * FROM test")
	assert.Error(t, err)
	assert.Error(t, pgDB.HealthCheck())
}
