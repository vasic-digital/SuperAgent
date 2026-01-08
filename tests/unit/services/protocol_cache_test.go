package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/helixagent/helixagent/internal/services"
)

func TestProtocolCache_PatternMatching(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel) // Silence logs during tests

	cache := services.NewProtocolCache(1000, 5*time.Minute, logger)
	require.NotNil(t, cache)
	defer cache.Stop()

	ctx := context.Background()

	t.Run("prefix pattern with wildcard", func(t *testing.T) {
		testCache := services.NewProtocolCache(1000, 5*time.Minute, logger)
		defer testCache.Stop()

		testCache.Set(ctx, "mcp:server:s1", "v1", nil, 5*time.Minute)
		testCache.Set(ctx, "mcp:server:s2", "v2", nil, 5*time.Minute)
		testCache.Set(ctx, "mcp:tools:t1", "v3", nil, 5*time.Minute)
		testCache.Set(ctx, "lsp:server:l1", "v4", nil, 5*time.Minute)

		// Invalidate mcp:server:*
		err := testCache.InvalidateByPattern(ctx, "mcp:server:*")
		require.NoError(t, err)

		// Check remaining entries
		_, exists1, _ := testCache.Get(ctx, "mcp:server:s1")
		_, exists2, _ := testCache.Get(ctx, "mcp:server:s2")
		_, exists3, _ := testCache.Get(ctx, "mcp:tools:t1")
		_, exists4, _ := testCache.Get(ctx, "lsp:server:l1")

		assert.False(t, exists1, "mcp:server:s1 should be invalidated")
		assert.False(t, exists2, "mcp:server:s2 should be invalidated")
		assert.True(t, exists3, "mcp:tools:t1 should remain")
		assert.True(t, exists4, "lsp:server:l1 should remain")
	})

	t.Run("middle wildcard pattern", func(t *testing.T) {
		testCache := services.NewProtocolCache(1000, 5*time.Minute, logger)
		defer testCache.Stop()

		testCache.Set(ctx, "prefix-abc-suffix", "v1", nil, 5*time.Minute)
		testCache.Set(ctx, "prefix-xyz-suffix", "v2", nil, 5*time.Minute)
		testCache.Set(ctx, "prefix-other", "v3", nil, 5*time.Minute)

		// Invalidate prefix-*-suffix
		err := testCache.InvalidateByPattern(ctx, "prefix-*-suffix")
		require.NoError(t, err)

		_, exists1, _ := testCache.Get(ctx, "prefix-abc-suffix")
		_, exists2, _ := testCache.Get(ctx, "prefix-xyz-suffix")
		_, exists3, _ := testCache.Get(ctx, "prefix-other")

		assert.False(t, exists1, "prefix-abc-suffix should be invalidated")
		assert.False(t, exists2, "prefix-xyz-suffix should be invalidated")
		assert.True(t, exists3, "prefix-other should remain")
	})

	t.Run("question mark matches single character", func(t *testing.T) {
		testCache := services.NewProtocolCache(1000, 5*time.Minute, logger)
		defer testCache.Stop()

		testCache.Set(ctx, "test1", "v1", nil, 5*time.Minute)
		testCache.Set(ctx, "test2", "v2", nil, 5*time.Minute)
		testCache.Set(ctx, "test12", "v3", nil, 5*time.Minute)
		testCache.Set(ctx, "test", "v4", nil, 5*time.Minute)

		// Invalidate test? (test followed by exactly one char)
		err := testCache.InvalidateByPattern(ctx, "test?")
		require.NoError(t, err)

		_, exists1, _ := testCache.Get(ctx, "test1")
		_, exists2, _ := testCache.Get(ctx, "test2")
		_, exists3, _ := testCache.Get(ctx, "test12")
		_, exists4, _ := testCache.Get(ctx, "test")

		assert.False(t, exists1, "test1 should be invalidated")
		assert.False(t, exists2, "test2 should be invalidated")
		assert.True(t, exists3, "test12 should remain (too long)")
		assert.True(t, exists4, "test should remain (too short)")
	})

	t.Run("exact match pattern", func(t *testing.T) {
		testCache := services.NewProtocolCache(1000, 5*time.Minute, logger)
		defer testCache.Stop()

		testCache.Set(ctx, "exactkey", "v1", nil, 5*time.Minute)
		testCache.Set(ctx, "exactkey2", "v2", nil, 5*time.Minute)

		// Invalidate exact key
		err := testCache.InvalidateByPattern(ctx, "exactkey")
		require.NoError(t, err)

		_, exists1, _ := testCache.Get(ctx, "exactkey")
		_, exists2, _ := testCache.Get(ctx, "exactkey2")

		assert.False(t, exists1, "exactkey should be invalidated")
		assert.True(t, exists2, "exactkey2 should remain")
	})

	t.Run("empty pattern matches nothing", func(t *testing.T) {
		testCache := services.NewProtocolCache(1000, 5*time.Minute, logger)
		defer testCache.Stop()

		testCache.Set(ctx, "key1", "v1", nil, 5*time.Minute)

		err := testCache.InvalidateByPattern(ctx, "")
		require.NoError(t, err)

		_, exists, _ := testCache.Get(ctx, "key1")
		assert.True(t, exists, "key1 should remain (empty pattern)")
	})

	t.Run("complex pattern with multiple wildcards", func(t *testing.T) {
		testCache := services.NewProtocolCache(1000, 5*time.Minute, logger)
		defer testCache.Stop()

		testCache.Set(ctx, "a:b:c:d", "v1", nil, 5*time.Minute)
		testCache.Set(ctx, "a:x:c:y", "v2", nil, 5*time.Minute)
		testCache.Set(ctx, "a:bb:c:dd", "v3", nil, 5*time.Minute)
		testCache.Set(ctx, "x:b:c:d", "v4", nil, 5*time.Minute)

		// Invalidate a:*:c:*
		err := testCache.InvalidateByPattern(ctx, "a:*:c:*")
		require.NoError(t, err)

		_, exists1, _ := testCache.Get(ctx, "a:b:c:d")
		_, exists2, _ := testCache.Get(ctx, "a:x:c:y")
		_, exists3, _ := testCache.Get(ctx, "a:bb:c:dd")
		_, exists4, _ := testCache.Get(ctx, "x:b:c:d")

		assert.False(t, exists1, "a:b:c:d should be invalidated")
		assert.False(t, exists2, "a:x:c:y should be invalidated")
		assert.False(t, exists3, "a:bb:c:dd should be invalidated")
		assert.True(t, exists4, "x:b:c:d should remain")
	})

	t.Run("full wildcard matches all", func(t *testing.T) {
		testCache := services.NewProtocolCache(1000, 5*time.Minute, logger)
		defer testCache.Stop()

		testCache.Set(ctx, "key1", "v1", nil, 5*time.Minute)
		testCache.Set(ctx, "key2", "v2", nil, 5*time.Minute)
		testCache.Set(ctx, "other", "v3", nil, 5*time.Minute)

		// Invalidate all with *
		err := testCache.InvalidateByPattern(ctx, "*")
		require.NoError(t, err)

		_, exists1, _ := testCache.Get(ctx, "key1")
		_, exists2, _ := testCache.Get(ctx, "key2")
		_, exists3, _ := testCache.Get(ctx, "other")

		assert.False(t, exists1, "key1 should be invalidated")
		assert.False(t, exists2, "key2 should be invalidated")
		assert.False(t, exists3, "other should be invalidated")
	})
}

func TestProtocolCache_BasicOperations(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	cache := services.NewProtocolCache(1000, 5*time.Minute, logger)
	require.NotNil(t, cache)
	defer cache.Stop()

	ctx := context.Background()

	t.Run("set and get", func(t *testing.T) {
		err := cache.Set(ctx, "testkey", "testvalue", nil, 5*time.Minute)
		require.NoError(t, err)

		value, exists, err := cache.Get(ctx, "testkey")
		assert.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, "testvalue", value)
	})

	t.Run("delete", func(t *testing.T) {
		err := cache.Set(ctx, "todelete", "value", nil, 5*time.Minute)
		require.NoError(t, err)

		err = cache.Delete(ctx, "todelete")
		require.NoError(t, err)

		_, exists, _ := cache.Get(ctx, "todelete")
		assert.False(t, exists)
	})

	t.Run("get non-existent key", func(t *testing.T) {
		_, exists, err := cache.Get(ctx, "nonexistent")
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestProtocolCache_Tags(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	cache := services.NewProtocolCache(1000, 5*time.Minute, logger)
	require.NotNil(t, cache)
	defer cache.Stop()

	ctx := context.Background()

	t.Run("invalidate by tags", func(t *testing.T) {
		err := cache.Set(ctx, "mcp1", "v1", []string{"mcp", "server"}, 5*time.Minute)
		require.NoError(t, err)
		err = cache.Set(ctx, "mcp2", "v2", []string{"mcp", "tools"}, 5*time.Minute)
		require.NoError(t, err)
		err = cache.Set(ctx, "lsp1", "v3", []string{"lsp", "server"}, 5*time.Minute)
		require.NoError(t, err)

		// Invalidate by "mcp" tag
		err = cache.InvalidateByTags(ctx, []string{"mcp"})
		require.NoError(t, err)

		_, exists1, _ := cache.Get(ctx, "mcp1")
		_, exists2, _ := cache.Get(ctx, "mcp2")
		_, exists3, _ := cache.Get(ctx, "lsp1")

		assert.False(t, exists1, "mcp1 should be invalidated")
		assert.False(t, exists2, "mcp2 should be invalidated")
		assert.True(t, exists3, "lsp1 should remain")
	})
}
