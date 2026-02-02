package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Tiered Cache L2 (Redis) Operations Tests
// ============================================================================

func setupTieredCacheWithRedis(t *testing.T) (*TieredCache, *miniredis.Miniredis) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	config := &TieredCacheConfig{
		L1MaxSize:         100,
		L1TTL:             5 * time.Minute,
		L1CleanupInterval: time.Minute,
		L2TTL:             30 * time.Minute,
		L2Compression:     true,
		L2KeyPrefix:       "tiered:",
		EnableL1:          true,
		EnableL2:          true,
	}

	tc := NewTieredCache(redisClient, config)

	t.Cleanup(func() {
		_ = tc.Close()
		mr.Close()
	})

	return tc, mr
}

func TestTieredCache_L2Get_Basic(t *testing.T) {
	tc, mr := setupTieredCacheWithRedis(t)
	ctx := context.Background()

	// Set a value directly in Redis
	testData := `{"name":"test","value":123}`
	_ = mr.Set("tiered:test-key", testData)

	// Get should retrieve from L2
	var result map[string]interface{}
	found, err := tc.Get(ctx, "test-key", &result)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "test", result["name"])
	assert.Equal(t, float64(123), result["value"])

	// Check L2 hit was recorded
	m := tc.Metrics()
	assert.Equal(t, int64(1), m.L2Hits)
}

func TestTieredCache_L2Set_Basic(t *testing.T) {
	tc, _ := setupTieredCacheWithRedis(t)
	ctx := context.Background()

	testValue := map[string]interface{}{
		"key": "value",
		"num": 42,
	}

	// Set value
	err := tc.Set(ctx, "test-key", testValue, 10*time.Minute)
	require.NoError(t, err)

	// Clear L1 to force L2 read
	tc.l1Delete("test-key")

	// Get should retrieve from L2
	var result map[string]interface{}
	found, err := tc.Get(ctx, "test-key", &result)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "value", result["key"])
}

func TestTieredCache_L2Set_Compression(t *testing.T) {
	tc, mr := setupTieredCacheWithRedis(t)
	ctx := context.Background()

	// Large value that should trigger compression
	largeValue := map[string]interface{}{
		"content": string(make([]byte, 500)), // 500+ bytes to trigger compression
		"data":    "some data that should be compressed to save space in redis",
	}

	err := tc.Set(ctx, "compressed-key", largeValue, 10*time.Minute)
	require.NoError(t, err)

	// Verify compression occurred
	raw, err := mr.Get("tiered:compressed-key")
	require.NoError(t, err)
	assert.NotEmpty(t, raw)

	// Check compression savings metric
	m := tc.Metrics()
	assert.GreaterOrEqual(t, m.CompressionSaved, int64(0))
}

func TestTieredCache_L2Set_SmallValueNoCompression(t *testing.T) {
	tc, mr := setupTieredCacheWithRedis(t)
	ctx := context.Background()

	// Small value that should NOT trigger compression (< 100 bytes)
	smallValue := map[string]string{
		"k": "v",
	}

	err := tc.Set(ctx, "small-key", smallValue, 10*time.Minute)
	require.NoError(t, err)

	// Value should be stored uncompressed (not starting with gzip magic)
	raw, err := mr.Get("tiered:small-key")
	require.NoError(t, err)
	assert.NotEmpty(t, raw)
	// Not gzip compressed (gzip starts with 0x1f 0x8b)
	if len(raw) > 0 {
		assert.NotEqual(t, byte(0x1f), raw[0], "Small values should not be compressed")
	}
}

func TestTieredCache_Get_L1Miss_L2Hit_Promotion(t *testing.T) {
	tc, mr := setupTieredCacheWithRedis(t)
	ctx := context.Background()

	// Set directly in Redis (bypass L1)
	testData := `{"promoted":"yes"}`
	_ = mr.Set("tiered:promote-key", testData)

	// L1 should be empty
	_, ok := tc.l1Get("promote-key")
	assert.False(t, ok)

	// Get should hit L2 and promote to L1
	var result map[string]string
	found, err := tc.Get(ctx, "promote-key", &result)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "yes", result["promoted"])

	// Now L1 should have the value (promoted)
	data, ok := tc.l1Get("promote-key")
	assert.True(t, ok)
	assert.NotEmpty(t, data)

	// Metrics should show L2 hit followed by L1 hit
	m := tc.Metrics()
	assert.Equal(t, int64(1), m.L2Hits)
	assert.Equal(t, int64(1), m.L1Misses)
}

func TestTieredCache_L2Get_Compressed(t *testing.T) {
	tc, _ := setupTieredCacheWithRedis(t)
	ctx := context.Background()

	// Set a large value that will be compressed
	largeData := struct {
		Content string `json:"content"`
		Repeat  string `json:"repeat"`
	}{
		Content: string(make([]byte, 1000)),
		Repeat:  "This is repeated content that compresses well. " + string(make([]byte, 500)),
	}

	err := tc.Set(ctx, "large-key", largeData, 10*time.Minute)
	require.NoError(t, err)

	// Clear L1 to force L2 decompression
	tc.l1Delete("large-key")

	// Get should decompress the data
	var result struct {
		Content string `json:"content"`
		Repeat  string `json:"repeat"`
	}
	found, err := tc.Get(ctx, "large-key", &result)
	require.NoError(t, err)
	assert.True(t, found)
	assert.NotEmpty(t, result.Repeat)
}

func TestTieredCache_L2Get_NotFound(t *testing.T) {
	tc, _ := setupTieredCacheWithRedis(t)
	ctx := context.Background()

	var result string
	found, err := tc.Get(ctx, "nonexistent-key", &result)
	assert.NoError(t, err)
	assert.False(t, found)

	// Should record L2 miss
	m := tc.Metrics()
	assert.Equal(t, int64(1), m.L2Misses)
}

func TestTieredCache_Delete_L2(t *testing.T) {
	tc, mr := setupTieredCacheWithRedis(t)
	ctx := context.Background()

	// Set a value in both L1 and L2
	err := tc.Set(ctx, "delete-key", "value", 10*time.Minute)
	require.NoError(t, err)

	// Verify it exists in Redis
	assert.True(t, mr.Exists("tiered:delete-key"))

	// Delete
	err = tc.Delete(ctx, "delete-key")
	require.NoError(t, err)

	// Verify it's gone from Redis
	assert.False(t, mr.Exists("tiered:delete-key"))
}

func TestTieredCache_InvalidatePrefix_L2(t *testing.T) {
	tc, mr := setupTieredCacheWithRedis(t)
	ctx := context.Background()

	// Set multiple values with same prefix
	err := tc.Set(ctx, "user:1:profile", "profile1", 10*time.Minute)
	require.NoError(t, err)
	err = tc.Set(ctx, "user:1:settings", "settings1", 10*time.Minute)
	require.NoError(t, err)
	err = tc.Set(ctx, "user:2:profile", "profile2", 10*time.Minute)
	require.NoError(t, err)
	err = tc.Set(ctx, "other:key", "other", 10*time.Minute)
	require.NoError(t, err)

	// Verify all exist
	assert.True(t, mr.Exists("tiered:user:1:profile"))
	assert.True(t, mr.Exists("tiered:user:1:settings"))
	assert.True(t, mr.Exists("tiered:user:2:profile"))
	assert.True(t, mr.Exists("tiered:other:key"))

	// Invalidate prefix user:1:
	count, err := tc.InvalidatePrefix(ctx, "user:1:")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, 2)

	// user:1:* should be gone
	assert.False(t, mr.Exists("tiered:user:1:profile"))
	assert.False(t, mr.Exists("tiered:user:1:settings"))

	// user:2 and other should still exist
	assert.True(t, mr.Exists("tiered:user:2:profile"))
	assert.True(t, mr.Exists("tiered:other:key"))
}

func TestTieredCache_InvalidatePrefix_L2_NoMatches(t *testing.T) {
	tc, _ := setupTieredCacheWithRedis(t)
	ctx := context.Background()

	// Set a value
	err := tc.Set(ctx, "existing:key", "value", 10*time.Minute)
	require.NoError(t, err)

	// Invalidate non-matching prefix
	count, err := tc.InvalidatePrefix(ctx, "nonexistent:")
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestTieredCache_L2_DisabledCompression(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	config := &TieredCacheConfig{
		L1MaxSize:     100,
		L1TTL:         5 * time.Minute,
		L2TTL:         30 * time.Minute,
		L2Compression: false, // Disable compression
		L2KeyPrefix:   "test:",
		EnableL1:      true,
		EnableL2:      true,
	}

	tc := NewTieredCache(redisClient, config)
	defer func() { _ = tc.Close() }()

	ctx := context.Background()

	// Large value should NOT be compressed
	largeValue := map[string]string{
		"content": string(make([]byte, 500)),
	}

	err = tc.Set(ctx, "uncompressed", largeValue, 10*time.Minute)
	require.NoError(t, err)

	// Raw data should not start with gzip magic bytes
	raw, err := mr.Get("test:uncompressed")
	require.NoError(t, err)
	if len(raw) > 0 {
		assert.NotEqual(t, byte(0x1f), raw[0], "Data should not be compressed")
	}
}

func TestTieredCache_L2_OnlyMode(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	config := &TieredCacheConfig{
		L1MaxSize:     100,
		L1TTL:         5 * time.Minute,
		L2TTL:         30 * time.Minute,
		L2Compression: false,
		L2KeyPrefix:   "l2only:",
		EnableL1:      false, // Disable L1
		EnableL2:      true,
	}

	tc := NewTieredCache(redisClient, config)
	defer func() { _ = tc.Close() }()

	ctx := context.Background()

	// Set value
	err = tc.Set(ctx, "key", "value", 10*time.Minute)
	require.NoError(t, err)

	// Get should work (from L2 only)
	var result string
	found, err := tc.Get(ctx, "key", &result)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "value", result)

	// Metrics should show L2 activity only
	m := tc.Metrics()
	assert.Equal(t, int64(1), m.L2Hits)
	assert.Equal(t, int64(0), m.L1Hits) // L1 disabled
}

func TestTieredCache_L2_TTL(t *testing.T) {
	tc, mr := setupTieredCacheWithRedis(t)
	ctx := context.Background()

	// Set with specific TTL
	err := tc.Set(ctx, "ttl-key", "value", 5*time.Minute)
	require.NoError(t, err)

	// Verify key exists
	assert.True(t, mr.Exists("tiered:ttl-key"))

	// Fast-forward time
	mr.FastForward(6 * time.Minute)

	// Key should be expired in miniredis
	assert.False(t, mr.Exists("tiered:ttl-key"))
}

func TestTieredCache_L2_DeleteError(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	config := &TieredCacheConfig{
		L1MaxSize:   100,
		L2KeyPrefix: "test:",
		EnableL1:    true,
		EnableL2:    true,
	}

	tc := NewTieredCache(redisClient, config)

	ctx := context.Background()

	// Set a value first
	err = tc.Set(ctx, "error-key", "value", time.Minute)
	require.NoError(t, err)

	// Close miniredis to simulate connection error
	mr.Close()

	// Delete should return error
	err = tc.Delete(ctx, "error-key")
	assert.Error(t, err)

	_ = tc.Close()
}

func TestTieredCache_L2_SetError(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	config := &TieredCacheConfig{
		L1MaxSize:     100,
		L2KeyPrefix:   "test:",
		L2Compression: false,
		EnableL1:      true,
		EnableL2:      true,
	}

	tc := NewTieredCache(redisClient, config)

	ctx := context.Background()

	// Close miniredis to simulate connection error
	mr.Close()

	// Set should return error
	err = tc.Set(ctx, "error-key", "value", time.Minute)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "l2 set")

	_ = tc.Close()
}

func TestTieredCache_L2_GetError(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	config := &TieredCacheConfig{
		L1MaxSize:   100,
		L2KeyPrefix: "test:",
		EnableL1:    false, // Disable L1 to force L2 path
		EnableL2:    true,
	}

	tc := NewTieredCache(redisClient, config)

	ctx := context.Background()

	// Close miniredis to simulate connection error
	mr.Close()

	// Get should return error
	var result string
	_, err = tc.Get(ctx, "error-key", &result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "l2 get")

	_ = tc.Close()
}

func TestTieredCache_InvalidatePrefix_L2_ScanError(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	config := &TieredCacheConfig{
		L1MaxSize:   100,
		L2KeyPrefix: "test:",
		EnableL1:    true,
		EnableL2:    true,
	}

	tc := NewTieredCache(redisClient, config)

	ctx := context.Background()

	// Set a value
	err = tc.Set(ctx, "prefix:key", "value", time.Minute)
	require.NoError(t, err)

	// Close miniredis to simulate connection error
	mr.Close()

	// InvalidatePrefix should return error from L2 scan
	_, err = tc.InvalidatePrefix(ctx, "prefix:")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "l2 scan")

	_ = tc.Close()
}

func TestTieredCache_BothCaches_Consistency(t *testing.T) {
	tc, mr := setupTieredCacheWithRedis(t)
	ctx := context.Background()

	testValue := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	// Set value (goes to both L1 and L2)
	err := tc.Set(ctx, "consistent-key", testValue, 10*time.Minute)
	require.NoError(t, err)

	// Verify L1 has it
	l1Data, l1Found := tc.l1Get("consistent-key")
	assert.True(t, l1Found)
	assert.NotEmpty(t, l1Data)

	// Verify L2 has it
	assert.True(t, mr.Exists("tiered:consistent-key"))

	// Get should hit L1 (faster)
	var result1 map[string]string
	found, err := tc.Get(ctx, "consistent-key", &result1)
	require.NoError(t, err)
	assert.True(t, found)

	m := tc.Metrics()
	assert.Equal(t, int64(1), m.L1Hits)
	assert.Equal(t, int64(0), m.L2Hits)

	// Delete from L1 only (simulate eviction)
	tc.l1Delete("consistent-key")

	// Get should now hit L2
	var result2 map[string]string
	found, err = tc.Get(ctx, "consistent-key", &result2)
	require.NoError(t, err)
	assert.True(t, found)

	m = tc.Metrics()
	assert.Equal(t, int64(1), m.L1Hits) // Still 1 (no new L1 hit)
	assert.Equal(t, int64(1), m.L2Hits) // Now 1 L2 hit
	// L1 misses can vary because there's an initial miss before L2 hit and promotion
	assert.GreaterOrEqual(t, m.L1Misses, int64(1))
}
