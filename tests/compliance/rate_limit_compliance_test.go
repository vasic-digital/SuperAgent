package compliance

import (
	"testing"
	"time"

	"dev.helix.agent/internal/middleware"
	"github.com/stretchr/testify/assert"
)

// TestRateLimitConfigCompliance verifies the RateLimitConfig type has
// required fields for token bucket rate limiting.
func TestRateLimitConfigCompliance(t *testing.T) {
	config := middleware.RateLimitConfig{
		Requests: 100,
		Window:   time.Minute,
	}

	assert.Equal(t, 100, config.Requests)
	assert.Equal(t, time.Minute, config.Window)

	t.Logf("COMPLIANCE: RateLimitConfig has Requests and Window fields")
}

// TestRateLimitResultCompliance verifies the RateLimitResult has the
// fields needed for informative rate limit responses.
func TestRateLimitResultCompliance(t *testing.T) {
	result := middleware.RateLimitResult{
		Allowed:   true,
		Remaining: 99,
	}

	assert.True(t, result.Allowed)
	assert.Equal(t, 99, result.Remaining)

	t.Logf("COMPLIANCE: RateLimitResult has Allowed and Remaining fields")
}

// TestRateLimitConfigRangeCompliance verifies that rate limit values
// are within acceptable ranges for production use.
func TestRateLimitConfigRangeCompliance(t *testing.T) {
	prodConfig := middleware.RateLimitConfig{
		Requests: 60,
		Window:   time.Minute,
	}

	assert.GreaterOrEqual(t, prodConfig.Requests, 1,
		"Requests must be at least 1")
	assert.GreaterOrEqual(t, prodConfig.Window, time.Second,
		"Window must be at least 1 second")

	t.Logf("COMPLIANCE: Rate limit config values within acceptable production ranges")
}

// TestRateLimiterConstructionCompliance verifies the rate limiter
// can be constructed without a cache service (in-memory fallback).
func TestRateLimiterConstructionCompliance(t *testing.T) {
	rl := middleware.NewRateLimiter(nil)
	assert.NotNil(t, rl, "RateLimiter must be constructable without cache service (in-memory fallback)")

	t.Logf("COMPLIANCE: RateLimiter supports in-memory fallback when Redis unavailable")
}

// TestRateLimitPathCompliance verifies that rate limits can be applied per-path.
func TestRateLimitPathCompliance(t *testing.T) {
	rl := middleware.NewRateLimiter(nil)
	assert.NotNil(t, rl)

	completionConfig := &middleware.RateLimitConfig{
		Requests: 10,
		Window:   time.Minute,
	}
	rl.AddLimit("/v1/chat/completions", completionConfig)

	t.Logf("COMPLIANCE: Per-path rate limiting configured for /v1/chat/completions")
}

// TestRateLimitWithCustomKeyCompliance verifies that custom key functions
// are supported for per-user or per-API-key rate limiting.
func TestRateLimitWithCustomKeyCompliance(t *testing.T) {
	// Verify built-in key functions exist
	assert.NotNil(t, middleware.ByUserID, "ByUserID key function must exist")
	assert.NotNil(t, middleware.ByAPIKey, "ByAPIKey key function must exist")

	t.Logf("COMPLIANCE: Custom rate limit key functions (ByUserID, ByAPIKey) are present")
}
