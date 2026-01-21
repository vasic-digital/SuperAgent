package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Extended Expiration Tests - ForceExpireByTag, SetTTL, and edge cases
// ============================================================================

func TestExpirationManager_ForceExpireByTag(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
		EnableL2:  false,
	}
	tc := NewTieredCache(nil, config)
	defer tc.Close()

	expConfig := DefaultExpirationConfig()
	em := NewExpirationManager(tc, expConfig)
	require.NotNil(t, em)

	ctx := context.Background()

	// Set values with tags
	err := tc.Set(ctx, "key1", "value1", time.Minute, "tag-a")
	require.NoError(t, err)
	err = tc.Set(ctx, "key2", "value2", time.Minute, "tag-a")
	require.NoError(t, err)
	err = tc.Set(ctx, "key3", "value3", time.Minute, "tag-b")
	require.NoError(t, err)

	// ForceExpireByTag tag-a
	count, err := em.ForceExpireByTag(ctx, "tag-a")
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	// Verify metrics
	m := em.Metrics()
	assert.Equal(t, int64(2), m.ForceExpired)

	// key1 and key2 should be gone
	var result string
	found, _ := tc.Get(ctx, "key1", &result)
	assert.False(t, found)
	found, _ = tc.Get(ctx, "key2", &result)
	assert.False(t, found)

	// key3 should still exist
	found, _ = tc.Get(ctx, "key3", &result)
	assert.True(t, found)
}

func TestExpirationManager_ForceExpireByTag_NoMatches(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	defer tc.Close()

	em := NewExpirationManager(tc, DefaultExpirationConfig())

	ctx := context.Background()

	// ForceExpireByTag with non-existent tag
	count, err := em.ForceExpireByTag(ctx, "nonexistent-tag")
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestExpirationManager_SetTTL(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize:         1000,
		L1TTL:             time.Minute,
		L1CleanupInterval: 10 * time.Millisecond,
		EnableL1:          true,
	}
	tc := NewTieredCache(nil, config)
	defer tc.Close()

	em := NewExpirationManager(tc, DefaultExpirationConfig())

	ctx := context.Background()

	// Set initial value
	err := tc.Set(ctx, "ttl-key", "initial-value", time.Hour)
	require.NoError(t, err)

	// Update value with new TTL
	err = em.SetTTL(ctx, "ttl-key", "updated-value", 50*time.Millisecond)
	require.NoError(t, err)

	// Verify value was updated
	var result string
	found, err := tc.Get(ctx, "ttl-key", &result)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "updated-value", result)

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Value should be expired
	found, _ = tc.Get(ctx, "ttl-key", &result)
	assert.False(t, found)
}

func TestExpirationManager_ValidateEntry_ValidationDisabled(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	defer tc.Close()

	expConfig := &ExpirationConfig{
		CleanupInterval:    time.Minute,
		DefaultTTL:         30 * time.Minute,
		MaxAge:             time.Hour,
		EnableValidation:   false, // Validation disabled
		ValidationInterval: 5 * time.Minute,
	}
	em := NewExpirationManager(tc, expConfig)

	// With validation disabled, everything should be valid
	valid := em.ValidateEntry("any:key", "value", 100*time.Hour)
	assert.True(t, valid)
}

func TestExpirationManager_ValidateEntry_MaxAgeExceeded(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	defer tc.Close()

	expConfig := &ExpirationConfig{
		CleanupInterval:    time.Minute,
		DefaultTTL:         30 * time.Minute,
		MaxAge:             time.Hour, // 1 hour max age
		EnableValidation:   true,
		ValidationInterval: 5 * time.Minute,
	}
	em := NewExpirationManager(tc, expConfig)

	// Entry older than MaxAge should be invalid
	valid := em.ValidateEntry("key", "value", 2*time.Hour)
	assert.False(t, valid)

	// Entry younger than MaxAge should be valid
	valid = em.ValidateEntry("key", "value", 30*time.Minute)
	assert.True(t, valid)
}

func TestExpirationManager_ValidateEntry_WithMultipleValidators(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	defer tc.Close()

	expConfig := &ExpirationConfig{
		CleanupInterval:    time.Minute,
		DefaultTTL:         30 * time.Minute,
		MaxAge:             24 * time.Hour,
		EnableValidation:   true,
		ValidationInterval: 5 * time.Minute,
	}
	em := NewExpirationManager(tc, expConfig)

	// Register multiple validators
	em.RegisterValidator("provider:*", ProviderHealthValidator(time.Minute))
	em.RegisterValidator("session:*", SessionValidator(time.Hour))
	em.RegisterValidator("llm:*", LLMResponseValidator(6 * time.Hour))

	// Provider entry older than 1 minute should be invalid
	valid := em.ValidateEntry("provider:openai", "data", 2*time.Minute)
	assert.False(t, valid)

	// Session entry younger than 1 hour should be valid
	valid = em.ValidateEntry("session:user123", "data", 30*time.Minute)
	assert.True(t, valid)

	// LLM response younger than 6 hours should be valid
	valid = em.ValidateEntry("llm:response:123", "data", 5*time.Hour)
	assert.True(t, valid)

	// LLM response older than 6 hours should be invalid
	valid = em.ValidateEntry("llm:response:456", "data", 7*time.Hour)
	assert.False(t, valid)
}

func TestExpirationManager_ValidatorRegistration_Concurrent(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	defer tc.Close()

	em := NewExpirationManager(tc, DefaultExpirationConfig())

	done := make(chan bool, 50)

	// Concurrent registration
	for i := 0; i < 25; i++ {
		go func(idx int) {
			pattern := "pattern:" + string(rune('a'+idx%26)) + ":*"
			em.RegisterValidator(pattern, AlwaysValidValidator())
			done <- true
		}(i)
	}

	// Concurrent unregistration
	for i := 0; i < 25; i++ {
		go func(idx int) {
			pattern := "pattern:" + string(rune('a'+idx%26)) + ":*"
			em.UnregisterValidator(pattern)
			done <- true
		}(i)
	}

	// Wait for all
	for i := 0; i < 50; i++ {
		<-done
	}
}

func TestExpirationManager_MetricsAccumulation(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	defer tc.Close()

	expConfig := &ExpirationConfig{
		CleanupInterval:    10 * time.Millisecond,
		DefaultTTL:         time.Minute,
		MaxAge:             time.Hour,
		EnableValidation:   true,
		ValidationInterval: 10 * time.Millisecond,
	}
	em := NewExpirationManager(tc, expConfig)

	// Register validator that always invalidates
	em.RegisterValidator("expire:*", NeverCacheValidator())

	// Start manager
	em.Start()

	// Let it run for a bit
	time.Sleep(100 * time.Millisecond)

	// Validate some entries (will trigger ExpiredByValidation)
	em.ValidateEntry("expire:key1", "value", time.Millisecond)
	em.ValidateEntry("expire:key2", "value", time.Millisecond)

	// Stop manager
	em.Stop()

	// Check metrics
	m := em.Metrics()
	assert.GreaterOrEqual(t, m.CleanupRuns, int64(1))
	assert.GreaterOrEqual(t, m.ValidationRuns, int64(1))
	assert.GreaterOrEqual(t, m.ExpiredByValidation, int64(2))
	assert.GreaterOrEqual(t, m.CleanupDuration, int64(0))
}

func TestExpirationManager_NilConfig(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	defer tc.Close()

	// Pass nil config - should use defaults
	em := NewExpirationManager(tc, nil)
	require.NotNil(t, em)
	assert.NotNil(t, em.config)
	assert.Equal(t, time.Minute, em.config.CleanupInterval)
}

func TestExpirationManager_ForceExpire_Error(t *testing.T) {
	// Create cache with L2 enabled but no Redis
	config := &TieredCacheConfig{
		L1MaxSize:   100,
		EnableL1:    true,
		EnableL2:    false, // L2 disabled so no error from Redis
		L2KeyPrefix: "test:",
	}
	tc := NewTieredCache(nil, config)
	defer tc.Close()

	em := NewExpirationManager(tc, DefaultExpirationConfig())

	ctx := context.Background()

	// ForceExpire should work even with no matching keys
	count, err := em.ForceExpire(ctx, "nonexistent:")
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestMatchPattern_Various(t *testing.T) {
	testCases := []struct {
		pattern  string
		input    string
		expected bool
	}{
		// Exact match
		{"key", "key", true},
		{"key", "other", false},

		// Wildcard at end
		{"provider:*", "provider:openai", true},
		{"provider:*", "provider:anthropic", true},
		{"provider:*", "other:openai", false},

		// Wildcard at start
		{"*:config", "user:config", true},
		{"*:config", "system:config", true},
		{"*:config", "config", false},

		// Wildcard in middle
		{"user:*:settings", "user:123:settings", true},
		{"user:*:settings", "user:abc:settings", true},
		{"user:*:settings", "user:settings", false},

		// Multiple wildcards
		{"*:*:*", "a:b:c", true},
		{"*:*:*", "ab:cd:ef", true},
		{"*:*:*", "a:b", false},

		// Empty cases
		{"", "", true},
		{"*", "anything", true},
		{"*", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.pattern+"_"+tc.input, func(t *testing.T) {
			result := matchPattern(tc.pattern, tc.input)
			assert.Equal(t, tc.expected, result, "Pattern: %s, Input: %s", tc.pattern, tc.input)
		})
	}
}

func TestMCPResultValidator_ToolSpecificTTLs(t *testing.T) {
	ttls := map[string]time.Duration{
		"read_file":  1 * time.Minute,
		"list_files": 2 * time.Minute,
		"get_repo":   5 * time.Minute,
	}
	defaultTTL := 10 * time.Minute

	validator := MCPResultValidator(ttls, defaultTTL)

	// Test tool with specific TTL - valid
	assert.True(t, validator("mcp:filesystem:read_file", "data", 30*time.Second))

	// Test tool with specific TTL - expired
	assert.False(t, validator("mcp:filesystem:read_file", "data", 2*time.Minute))

	// Test tool with default TTL - valid
	assert.True(t, validator("mcp:github:create_issue", "data", 5*time.Minute))

	// Test tool with default TTL - expired
	assert.False(t, validator("mcp:github:create_issue", "data", 15*time.Minute))
}

func TestStandardValidators_EdgeCases(t *testing.T) {
	t.Run("ProviderHealthValidator_BoundaryAge", func(t *testing.T) {
		validator := ProviderHealthValidator(time.Minute)

		// Exactly at boundary
		assert.False(t, validator("provider:test", "data", time.Minute))

		// Just under boundary
		assert.True(t, validator("provider:test", "data", 59*time.Second))
	})

	t.Run("SessionValidator_ZeroAge", func(t *testing.T) {
		validator := SessionValidator(time.Hour)

		// Zero age should be valid
		assert.True(t, validator("session:user", "data", 0))
	})

	t.Run("LLMResponseValidator_LargeValue", func(t *testing.T) {
		validator := LLMResponseValidator(24 * time.Hour)

		largeValue := make([]byte, 1024*1024) // 1MB
		assert.True(t, validator("llm:response", largeValue, time.Hour))
	})
}

func TestExpirationManager_CleanupLoop_ContextCancel(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	defer tc.Close()

	expConfig := &ExpirationConfig{
		CleanupInterval:    10 * time.Millisecond,
		DefaultTTL:         time.Minute,
		MaxAge:             time.Hour,
		EnableValidation:   true,
		ValidationInterval: 10 * time.Millisecond,
	}
	em := NewExpirationManager(tc, expConfig)

	em.Start()

	// Let it run briefly
	time.Sleep(50 * time.Millisecond)

	// Stop should cancel the context and exit loops
	em.Stop()

	// Wait a bit more and check no panic
	time.Sleep(50 * time.Millisecond)

	// Metrics should show some cleanup runs happened
	m := em.Metrics()
	assert.GreaterOrEqual(t, m.CleanupRuns, int64(1))
}

func TestExpirationMetrics_AllFields(t *testing.T) {
	metrics := &ExpirationMetrics{
		ExpiredByTTL:        10,
		ExpiredByValidation: 5,
		ForceExpired:        3,
		ValidationRuns:      20,
		ValidationErrors:    1,
		CleanupRuns:         15,
		CleanupDuration:     5000,
	}

	assert.Equal(t, int64(10), metrics.ExpiredByTTL)
	assert.Equal(t, int64(5), metrics.ExpiredByValidation)
	assert.Equal(t, int64(3), metrics.ForceExpired)
	assert.Equal(t, int64(20), metrics.ValidationRuns)
	assert.Equal(t, int64(1), metrics.ValidationErrors)
	assert.Equal(t, int64(15), metrics.CleanupRuns)
	assert.Equal(t, int64(5000), metrics.CleanupDuration)
}

func TestValidatorFunc_Type(t *testing.T) {
	// Test that ValidatorFunc type signature is correct
	var fn ValidatorFunc = func(key string, value interface{}, age time.Duration) bool {
		return key == "valid"
	}

	assert.True(t, fn("valid", nil, 0))
	assert.False(t, fn("invalid", nil, 0))
}
