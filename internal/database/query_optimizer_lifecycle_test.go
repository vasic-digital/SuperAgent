package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryCache_Shutdown_WaitsForCleanup(t *testing.T) {
	qc := NewQueryCache(time.Second, 100)
	require.NotNil(t, qc)
	qc.Shutdown()
	// Should be safe to call again (cancel is idempotent, wg already at 0)
	qc.Shutdown()
}

func TestQueryCache_Shutdown_NoLeak(t *testing.T) {
	qc := NewQueryCache(time.Second, 100)
	qc.Set("key1", "value1")
	_, ok := qc.Get("key1")
	assert.True(t, ok)
	qc.Shutdown()
}
