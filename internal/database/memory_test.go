package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMemoryDB(t *testing.T) {
	db := NewMemoryDB()
	require.NotNil(t, db)
	assert.True(t, db.enabled)
	assert.NotNil(t, db.data)
}

func TestMemoryDB_Ping(t *testing.T) {
	db := NewMemoryDB()
	defer db.Close()

	err := db.Ping()
	assert.NoError(t, err)
}

func TestMemoryDB_Exec(t *testing.T) {
	db := NewMemoryDB()
	defer db.Close()

	err := db.Exec("INSERT INTO test VALUES ($1)", "value")
	assert.NoError(t, err)
}

func TestMemoryDB_Query(t *testing.T) {
	db := NewMemoryDB()
	defer db.Close()

	result, err := db.Query("SELECT * FROM test")
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestMemoryDB_QueryRow(t *testing.T) {
	db := NewMemoryDB()
	defer db.Close()

	row := db.QueryRow("SELECT * FROM test")
	assert.NotNil(t, row)

	var value string
	err := row.Scan(&value)
	assert.Error(t, err)
}

func TestMemoryDB_Close(t *testing.T) {
	db := NewMemoryDB()

	err := db.Close()
	assert.NoError(t, err)
	assert.False(t, db.enabled)
	assert.Nil(t, db.data)
}

func TestMemoryDB_HealthCheck(t *testing.T) {
	db := NewMemoryDB()

	err := db.HealthCheck()
	assert.NoError(t, err)

	db.Close()

	err = db.HealthCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "memory database closed")
}

func TestMemoryDB_GetPool(t *testing.T) {
	db := NewMemoryDB()
	defer db.Close()

	pool := db.GetPool()
	assert.Nil(t, pool)
}

func TestMemoryDB_IsMemoryMode(t *testing.T) {
	db := NewMemoryDB()
	defer db.Close()

	assert.True(t, db.IsMemoryMode())
}

func TestMemoryRow_Scan_StringValue(t *testing.T) {
	row := &memoryRow{
		values: []any{"test-string"},
		err:    nil,
	}

	var s string
	err := row.Scan(&s)
	assert.NoError(t, err)
	assert.Equal(t, "test-string", s)
}

func TestMemoryRow_Scan_IntValue(t *testing.T) {
	row := &memoryRow{
		values: []any{42},
		err:    nil,
	}

	var n int
	err := row.Scan(&n)
	assert.NoError(t, err)
	assert.Equal(t, 42, n)
}

func TestMemoryRow_Scan_BoolValue(t *testing.T) {
	row := &memoryRow{
		values: []any{true},
		err:    nil,
	}

	var b bool
	err := row.Scan(&b)
	assert.NoError(t, err)
	assert.True(t, b)
}

func TestMemoryRow_Scan_MultipleValues(t *testing.T) {
	row := &memoryRow{
		values: []any{"hello", 123, true},
		err:    nil,
	}

	var s string
	var n int
	var b bool
	err := row.Scan(&s, &n, &b)
	assert.NoError(t, err)
	assert.Equal(t, "hello", s)
	assert.Equal(t, 123, n)
	assert.True(t, b)
}

func TestMemoryRow_Scan_WithError(t *testing.T) {
	row := &memoryRow{
		values: nil,
		err:    assert.AnError,
	}

	var s string
	err := row.Scan(&s)
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
}

func TestMemoryRow_Scan_NoRows(t *testing.T) {
	row := &memoryRow{
		values: []any{},
		err:    nil,
	}

	var s string
	err := row.Scan(&s)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no rows")
}

func TestMemoryRow_Scan_TypeMismatch(t *testing.T) {
	row := &memoryRow{
		values: []any{"string-value"},
		err:    nil,
	}

	var n int
	// Scan won't error but won't update the value either
	err := row.Scan(&n)
	assert.NoError(t, err)
	assert.Equal(t, 0, n) // Default int value, not updated
}

func TestMemoryRow_Scan_MoreDestsThanValues(t *testing.T) {
	row := &memoryRow{
		values: []any{"value1"},
		err:    nil,
	}

	var s1, s2 string
	err := row.Scan(&s1, &s2)
	assert.NoError(t, err)
	assert.Equal(t, "value1", s1)
	assert.Empty(t, s2) // Not filled
}
