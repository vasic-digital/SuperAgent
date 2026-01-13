package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpirationManager_Creation(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	expConfig := DefaultExpirationConfig()
	em := NewExpirationManager(tc, expConfig)
	require.NotNil(t, em)
}

func TestExpirationManager_RegisterValidator(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	expConfig := DefaultExpirationConfig()
	em := NewExpirationManager(tc, expConfig)
	require.NotNil(t, em)

	validator := func(key string, value interface{}, age time.Duration) bool {
		return age < time.Minute
	}

	em.RegisterValidator("test:*", validator)
	// Verify registration doesn't panic
}

func TestExpirationManager_ValidateEntry_WithoutValidator(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	expConfig := DefaultExpirationConfig()
	em := NewExpirationManager(tc, expConfig)
	require.NotNil(t, em)

	// Without validator matching, should return true (valid)
	valid := em.ValidateEntry("unknown:key", "value", time.Second)
	assert.True(t, valid)
}

func TestExpirationManager_ValidateEntry_WithValidator(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	expConfig := DefaultExpirationConfig()
	em := NewExpirationManager(tc, expConfig)
	require.NotNil(t, em)

	// Register validator that always returns false
	em.RegisterValidator("test:*", func(key string, value interface{}, age time.Duration) bool {
		return false
	})

	valid := em.ValidateEntry("test:key", "value", time.Second)
	assert.False(t, valid)
}

func TestExpirationManager_ForceExpire(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	expConfig := DefaultExpirationConfig()
	em := NewExpirationManager(tc, expConfig)
	require.NotNil(t, em)

	ctx := context.Background()

	// ForceExpire returns count of expired entries
	count, err := em.ForceExpire(ctx, "test:*")
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, count, 0)
}

func TestExpirationManager_StartAndStop(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	expConfig := &ExpirationConfig{
		CleanupInterval:    50 * time.Millisecond,
		DefaultTTL:         time.Second,
		MaxAge:             time.Hour,
		EnableValidation:   true,
		ValidationInterval: 100 * time.Millisecond,
	}
	em := NewExpirationManager(tc, expConfig)
	require.NotNil(t, em)

	// Start cleanup in background
	em.Start()

	// Let it run briefly
	time.Sleep(200 * time.Millisecond)

	// Stop should not panic
	em.Stop()
}

func TestDefaultExpirationConfig(t *testing.T) {
	config := DefaultExpirationConfig()
	assert.NotNil(t, config)
	assert.Greater(t, config.CleanupInterval, time.Duration(0))
	assert.Greater(t, config.DefaultTTL, time.Duration(0))
	assert.Greater(t, config.MaxAge, time.Duration(0))
}

func TestExpirationManager_ValidatorPattern_Matching(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	expConfig := DefaultExpirationConfig()
	em := NewExpirationManager(tc, expConfig)
	require.NotNil(t, em)

	// Register validator for provider:* pattern
	callCount := 0
	em.RegisterValidator("provider:*", func(key string, value interface{}, age time.Duration) bool {
		callCount++
		return true
	})

	// This should trigger the validator
	em.ValidateEntry("provider:openai", "data", time.Millisecond)
	// Validator should have been called
	assert.Equal(t, 1, callCount)
}

func TestExpirationManager_MultipleValidators(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	expConfig := DefaultExpirationConfig()
	em := NewExpirationManager(tc, expConfig)
	require.NotNil(t, em)

	em.RegisterValidator("cache:*", func(key string, value interface{}, age time.Duration) bool {
		return true
	})

	em.RegisterValidator("session:*", func(key string, value interface{}, age time.Duration) bool {
		return false
	})

	// Different keys should match different validators
	assert.True(t, em.ValidateEntry("cache:item", "data", time.Millisecond))
	assert.False(t, em.ValidateEntry("session:user", "data", time.Millisecond))
}

func TestExpirationManager_MaxAgeValidation(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	expConfig := &ExpirationConfig{
		CleanupInterval:  time.Minute,
		DefaultTTL:       30 * time.Minute,
		MaxAge:           time.Hour,
		EnableValidation: true,
	}
	em := NewExpirationManager(tc, expConfig)
	require.NotNil(t, em)

	// Entry younger than MaxAge should be valid
	assert.True(t, em.ValidateEntry("key", "value", 30*time.Minute))

	// Entry older than MaxAge should be invalid
	assert.False(t, em.ValidateEntry("key", "value", 2*time.Hour))
}

func TestExpirationManager_Metrics(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	expConfig := DefaultExpirationConfig()
	em := NewExpirationManager(tc, expConfig)
	require.NotNil(t, em)

	metrics := em.Metrics()
	assert.NotNil(t, metrics)
	assert.GreaterOrEqual(t, metrics.ExpiredByTTL, int64(0))
	assert.GreaterOrEqual(t, metrics.ExpiredByValidation, int64(0))
	assert.GreaterOrEqual(t, metrics.ForceExpired, int64(0))
}

func TestExpirationManager_UnregisterValidator(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	expConfig := DefaultExpirationConfig()
	em := NewExpirationManager(tc, expConfig)
	require.NotNil(t, em)

	callCount := 0
	em.RegisterValidator("test:*", func(key string, value interface{}, age time.Duration) bool {
		callCount++
		return true
	})

	// Validator should be called
	em.ValidateEntry("test:key", "value", time.Millisecond)
	assert.Equal(t, 1, callCount)

	// Unregister validator
	em.UnregisterValidator("test:*")

	// Validator should not be called anymore
	em.ValidateEntry("test:key", "value", time.Millisecond)
	assert.Equal(t, 1, callCount)
}

func TestStandardValidators(t *testing.T) {
	t.Run("ProviderHealthValidator", func(t *testing.T) {
		validator := ProviderHealthValidator(time.Minute)
		assert.True(t, validator("provider:health", "data", 30*time.Second))
		assert.False(t, validator("provider:health", "data", 2*time.Minute))
	})

	t.Run("SessionValidator", func(t *testing.T) {
		validator := SessionValidator(time.Hour)
		assert.True(t, validator("session:user", "data", 30*time.Minute))
		assert.False(t, validator("session:user", "data", 2*time.Hour))
	})

	t.Run("LLMResponseValidator", func(t *testing.T) {
		validator := LLMResponseValidator(time.Hour)
		assert.True(t, validator("llm:response", "data", 30*time.Minute))
		assert.False(t, validator("llm:response", "data", 2*time.Hour))
	})

	t.Run("NeverCacheValidator", func(t *testing.T) {
		validator := NeverCacheValidator()
		assert.False(t, validator("any:key", "data", time.Millisecond))
	})

	t.Run("AlwaysValidValidator", func(t *testing.T) {
		validator := AlwaysValidValidator()
		assert.True(t, validator("any:key", "data", time.Hour*24*365))
	})
}

func TestMCPResultValidator(t *testing.T) {
	ttls := map[string]time.Duration{
		"read_file": 5 * time.Minute,
		"get_repo":  time.Hour,
	}
	validator := MCPResultValidator(ttls, 10*time.Minute)

	// Tool with specific TTL
	assert.True(t, validator("mcp:read_file", "data", 2*time.Minute))
	assert.False(t, validator("mcp:read_file", "data", 10*time.Minute))

	// Tool with default TTL
	assert.True(t, validator("mcp:other_tool", "data", 5*time.Minute))
	assert.False(t, validator("mcp:other_tool", "data", 15*time.Minute))
}
