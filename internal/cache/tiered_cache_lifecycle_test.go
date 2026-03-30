package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTieredCache_Close_WaitsForCleanup(t *testing.T) {
	cfg := &TieredCacheConfig{
		EnableL1:          true,
		L1MaxSize:         100,
		L1TTL:             time.Second,
		L1CleanupInterval: 50 * time.Millisecond,
	}
	tc := NewTieredCache(nil, cfg)
	err := tc.Close()
	assert.NoError(t, err)
}

func TestTieredCache_Close_Idempotent(t *testing.T) {
	cfg := &TieredCacheConfig{
		EnableL1:          true,
		L1MaxSize:         100,
		L1TTL:             time.Second,
		L1CleanupInterval: 50 * time.Millisecond,
	}
	tc := NewTieredCache(nil, cfg)
	err := tc.Close()
	assert.NoError(t, err)
	err = tc.Close()
	assert.NoError(t, err)
}

func TestTieredCache_Close_NoL1(t *testing.T) {
	cfg := &TieredCacheConfig{EnableL1: false}
	tc := NewTieredCache(nil, cfg)
	err := tc.Close()
	assert.NoError(t, err)
}
