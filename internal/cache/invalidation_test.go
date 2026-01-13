package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTagBasedInvalidation_Creation(t *testing.T) {
	inv := NewTagBasedInvalidation()
	require.NotNil(t, inv)
}

func TestTagBasedInvalidation_AddTag(t *testing.T) {
	inv := NewTagBasedInvalidation()
	require.NotNil(t, inv)

	inv.AddTag("key1", "tag1", "tag2")
	inv.AddTag("key2", "tag1")
	inv.AddTag("key3", "tag2")

	// Verify tags were added
	tags := inv.GetTags("key1")
	assert.Contains(t, tags, "tag1")
	assert.Contains(t, tags, "tag2")
}

func TestTagBasedInvalidation_InvalidateByTag_Single(t *testing.T) {
	inv := NewTagBasedInvalidation()
	require.NotNil(t, inv)

	inv.AddTag("key1", "tag1")
	inv.AddTag("key2", "tag1")
	inv.AddTag("key3", "tag2")

	keys := inv.InvalidateByTag("tag1")
	assert.Len(t, keys, 2)
	assert.Contains(t, keys, "key1")
	assert.Contains(t, keys, "key2")
}

func TestTagBasedInvalidation_InvalidateByTags_Multiple(t *testing.T) {
	inv := NewTagBasedInvalidation()
	require.NotNil(t, inv)

	inv.AddTag("key1", "tag1")
	inv.AddTag("key2", "tag2")
	inv.AddTag("key3", "tag3")

	keys := inv.InvalidateByTags("tag1", "tag2")
	assert.GreaterOrEqual(t, len(keys), 2)
	assert.Contains(t, keys, "key1")
	assert.Contains(t, keys, "key2")
}

func TestTagBasedInvalidation_InvalidateByTag_NonExistent(t *testing.T) {
	inv := NewTagBasedInvalidation()
	require.NotNil(t, inv)

	keys := inv.InvalidateByTag("nonexistent")
	assert.Empty(t, keys)
}

func TestTagBasedInvalidation_Concurrency(t *testing.T) {
	inv := NewTagBasedInvalidation()
	require.NotNil(t, inv)

	// Concurrent adds
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			for j := 0; j < 100; j++ {
				inv.AddTag("key"+string(rune('a'+n)), "tag"+string(rune('0'+n%10)))
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestTagBasedInvalidation_RemoveKey(t *testing.T) {
	inv := NewTagBasedInvalidation()
	require.NotNil(t, inv)

	inv.AddTag("key1", "tag1", "tag2")
	inv.AddTag("key2", "tag1")

	// Remove key1
	inv.RemoveKey("key1")

	// key1 should not be in tag1's keys anymore
	keys := inv.InvalidateByTag("tag1")
	assert.NotContains(t, keys, "key1")
	assert.Contains(t, keys, "key2")

	// key1's tags should be empty
	tags := inv.GetTags("key1")
	assert.Empty(t, tags)
}

func TestTagBasedInvalidation_GetKeys(t *testing.T) {
	inv := NewTagBasedInvalidation()
	require.NotNil(t, inv)

	inv.AddTag("key1", "tag1")
	inv.AddTag("key2", "tag1")
	inv.AddTag("key3", "tag2")

	keys := inv.GetKeys("tag1")
	assert.Len(t, keys, 2)
	assert.Contains(t, keys, "key1")
	assert.Contains(t, keys, "key2")
}

func TestTagBasedInvalidation_Metrics(t *testing.T) {
	inv := NewTagBasedInvalidation()
	require.NotNil(t, inv)

	inv.AddTag("key1", "tag1")
	inv.InvalidateByTag("tag1")

	metrics := inv.Metrics()
	assert.NotNil(t, metrics)
	assert.GreaterOrEqual(t, metrics.TagInvalidations, int64(1))
}

func TestTagBasedInvalidation_ShouldInvalidate(t *testing.T) {
	inv := NewTagBasedInvalidation()
	require.NotNil(t, inv)

	// Tag-based invalidation returns nil for event-based invalidation
	keys := inv.ShouldInvalidate(nil)
	assert.Nil(t, keys)
}

func TestTagBasedInvalidation_Name(t *testing.T) {
	inv := NewTagBasedInvalidation()
	require.NotNil(t, inv)

	assert.Equal(t, "tag-based", inv.Name())
}

func TestEventDrivenInvalidation_Creation(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	inv := NewEventDrivenInvalidation(nil, tc)
	require.NotNil(t, inv)
}

func TestEventDrivenInvalidation_StartAndStop(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)

	// Test with nil eventBus - should not panic
	inv := NewEventDrivenInvalidation(nil, tc)
	require.NotNil(t, inv)

	inv.Start()
	time.Sleep(50 * time.Millisecond)
	inv.Stop()
}

func TestEventDrivenInvalidation_Name(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	inv := NewEventDrivenInvalidation(nil, tc)

	assert.Equal(t, "event-driven", inv.Name())
}

func TestEventDrivenInvalidation_Metrics(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	inv := NewEventDrivenInvalidation(nil, tc)

	metrics := inv.Metrics()
	assert.NotNil(t, metrics)
	assert.GreaterOrEqual(t, metrics.EventInvalidations, int64(0))
}

func TestInvalidationStrategy_Interface(t *testing.T) {
	// Verify TagBasedInvalidation implements InvalidationStrategy
	var _ InvalidationStrategy = (*TagBasedInvalidation)(nil)

	// Verify EventDrivenInvalidation implements InvalidationStrategy
	var _ InvalidationStrategy = (*EventDrivenInvalidation)(nil)
}

func TestCompositeInvalidation_Creation(t *testing.T) {
	tag := NewTagBasedInvalidation()
	config := &TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	event := NewEventDrivenInvalidation(nil, tc)

	composite := NewCompositeInvalidation(tag, event)
	require.NotNil(t, composite)
}

func TestCompositeInvalidation_Name(t *testing.T) {
	composite := NewCompositeInvalidation()
	assert.Equal(t, "composite", composite.Name())
}

func TestCompositeInvalidation_AddStrategy(t *testing.T) {
	composite := NewCompositeInvalidation()
	tag := NewTagBasedInvalidation()

	composite.AddStrategy(tag)
	// Should not panic
}

func TestCompositeInvalidation_ShouldInvalidate(t *testing.T) {
	composite := NewCompositeInvalidation()
	tag := NewTagBasedInvalidation()
	composite.AddStrategy(tag)

	// Should return empty for nil event
	keys := composite.ShouldInvalidate(nil)
	assert.Empty(t, keys)
}

func TestContainsWildcard(t *testing.T) {
	assert.True(t, containsWildcard("prefix:*"))
	assert.True(t, containsWildcard("*:suffix"))
	assert.True(t, containsWildcard("pre*fix"))
	assert.False(t, containsWildcard("no_wildcard"))
	assert.False(t, containsWildcard(""))
}

func TestTrimWildcard(t *testing.T) {
	assert.Equal(t, "prefix:", trimWildcard("prefix:*"))
	assert.Equal(t, "", trimWildcard("*:suffix"))
	assert.Equal(t, "pre", trimWildcard("pre*fix"))
	assert.Equal(t, "no_wildcard", trimWildcard("no_wildcard"))
}
