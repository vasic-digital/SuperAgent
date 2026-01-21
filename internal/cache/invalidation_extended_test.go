package cache

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"dev.helix.agent/internal/events"
)

// ============================================================================
// Extended Invalidation Tests - Event Handling, ShouldInvalidate, RemoveRules
// ============================================================================

func TestEventDrivenInvalidation_ShouldInvalidate_WithRules(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 100,
		L1TTL:     time.Minute,
		EnableL1:  true,
		EnableL2:  false,
	}
	tc := NewTieredCache(nil, config)
	defer tc.Close()

	inv := NewEventDrivenInvalidation(nil, tc)

	// Create an event
	event := events.NewEvent(events.EventProviderHealthChanged, "test", map[string]interface{}{
		"name": "openai",
	})

	// ShouldInvalidate should return keys based on default rules
	keys := inv.ShouldInvalidate(event)
	assert.NotEmpty(t, keys)
	assert.Contains(t, keys, "provider:*")
}

func TestEventDrivenInvalidation_ShouldInvalidate_NoMatchingRules(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 100,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	defer tc.Close()

	inv := NewEventDrivenInvalidation(nil, tc)

	// Create event for which there's no rule
	event := events.NewEvent(events.EventSystemStartup, "test", nil)

	keys := inv.ShouldInvalidate(event)
	assert.Nil(t, keys)
}

func TestEventDrivenInvalidation_ShouldInvalidate_WithHandler(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 100,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	defer tc.Close()

	inv := NewEventDrivenInvalidation(nil, tc)

	// Test with cache invalidation event
	event := events.NewEvent(events.EventCacheInvalidated, "test", map[string]interface{}{
		"keys": []string{"key1", "key2", "key3"},
	})

	keys := inv.ShouldInvalidate(event)
	assert.Len(t, keys, 3)
	assert.Contains(t, keys, "key1")
	assert.Contains(t, keys, "key2")
	assert.Contains(t, keys, "key3")
}

func TestEventDrivenInvalidation_ShouldInvalidate_PatternPayload(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 100,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	defer tc.Close()

	inv := NewEventDrivenInvalidation(nil, tc)

	// Test with cache invalidation event using pattern
	event := events.NewEvent(events.EventCacheInvalidated, "test", map[string]interface{}{
		"pattern": "user:*",
	})

	keys := inv.ShouldInvalidate(event)
	assert.Len(t, keys, 1)
	assert.Contains(t, keys, "user:*")
}

func TestEventDrivenInvalidation_RemoveRules(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 100,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	defer tc.Close()

	inv := NewEventDrivenInvalidation(nil, tc)

	// Verify default rules exist
	event := events.NewEvent(events.EventProviderHealthChanged, "test", map[string]interface{}{
		"name": "test-provider",
	})
	keys := inv.ShouldInvalidate(event)
	assert.NotEmpty(t, keys)

	// Remove rules for this event type
	inv.RemoveRules(events.EventProviderHealthChanged)

	// Now ShouldInvalidate should return nil
	keys = inv.ShouldInvalidate(event)
	assert.Nil(t, keys)
}

func TestEventDrivenInvalidation_AddRule_Custom(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 100,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	defer tc.Close()

	inv := NewEventDrivenInvalidation(nil, tc)

	// Add custom rule
	customRule := InvalidationRule{
		EventType:  events.EventSystemError,
		KeyPattern: "error:*",
		Tags:       []string{"errors"},
		Handler: func(event *events.Event) []string {
			return []string{"custom:key1", "custom:key2"}
		},
	}
	inv.AddRule(customRule)

	// Create matching event
	event := events.NewEvent(events.EventSystemError, "system", nil)

	keys := inv.ShouldInvalidate(event)
	assert.Len(t, keys, 3) // 2 from handler + 1 from KeyPattern
	assert.Contains(t, keys, "custom:key1")
	assert.Contains(t, keys, "custom:key2")
	assert.Contains(t, keys, "error:*")
}

func TestEventDrivenInvalidation_HandleEvent_WithCache(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 100,
		L1TTL:     time.Minute,
		EnableL1:  true,
		EnableL2:  false,
	}
	tc := NewTieredCache(nil, config)
	defer tc.Close()

	// Pre-populate cache with some values
	ctx := tc.ctx
	tc.Set(ctx, "provider:openai", "data", time.Minute)
	tc.Set(ctx, "provider:anthropic", "data", time.Minute)
	tc.Set(ctx, "other:key", "data", time.Minute)

	// Create event bus and invalidation
	bus := events.NewEventBus(nil)
	defer bus.Close()

	inv := NewEventDrivenInvalidation(bus, tc)
	inv.Start()

	// Give time for subscription
	time.Sleep(50 * time.Millisecond)

	// Publish event
	event := events.NewEvent(events.EventProviderHealthChanged, "test", map[string]interface{}{
		"name": "openai",
	})
	bus.Publish(event)

	// Give time for event processing
	time.Sleep(100 * time.Millisecond)

	// provider:* should be invalidated (but with prefix matching, provider: prefix)
	m := inv.Metrics()
	assert.GreaterOrEqual(t, m.EventInvalidations, int64(1))

	inv.Stop()
}

func TestEventDrivenInvalidation_HandleEvent_MCPServer(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 100,
		L1TTL:     time.Minute,
		EnableL1:  true,
		EnableL2:  false,
	}
	tc := NewTieredCache(nil, config)
	defer tc.Close()

	// Pre-populate cache with MCP data
	ctx := tc.ctx
	tc.Set(ctx, "mcp:filesystem:read", "data", time.Minute)
	tc.Set(ctx, "mcp:filesystem:write", "data", time.Minute)
	tc.Set(ctx, "mcp:github:get", "data", time.Minute)

	bus := events.NewEventBus(nil)
	defer bus.Close()

	inv := NewEventDrivenInvalidation(bus, tc)
	inv.Start()

	time.Sleep(50 * time.Millisecond)

	// Publish MCP disconnect event
	event := events.NewEvent(events.EventMCPServerDisconnected, "mcp", map[string]interface{}{
		"server": "filesystem",
	})
	bus.Publish(event)

	time.Sleep(100 * time.Millisecond)

	m := inv.Metrics()
	assert.GreaterOrEqual(t, m.EventInvalidations, int64(1))

	inv.Stop()
}

func TestEventDrivenInvalidation_HandleEvent_NoRules(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 100,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	defer tc.Close()

	bus := events.NewEventBus(nil)
	defer bus.Close()

	inv := NewEventDrivenInvalidation(bus, tc)

	// Remove all rules
	inv.RemoveRules(events.EventProviderHealthChanged)
	inv.RemoveRules(events.EventMCPServerDisconnected)
	inv.RemoveRules(events.EventCacheInvalidated)

	inv.Start()
	time.Sleep(50 * time.Millisecond)

	// Publish event with no matching rules
	event := events.NewEvent(events.EventSystemStartup, "system", nil)
	bus.Publish(event)

	time.Sleep(100 * time.Millisecond)

	m := inv.Metrics()
	assert.Equal(t, int64(0), m.EventInvalidations)

	inv.Stop()
}

func TestEventDrivenInvalidation_Start_NilEventBus(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 100,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	defer tc.Close()

	inv := NewEventDrivenInvalidation(nil, tc)

	// Start with nil event bus should not panic
	inv.Start()
	time.Sleep(50 * time.Millisecond)
	inv.Stop()
}

func TestEventDrivenInvalidation_Stop_Multiple(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 100,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	defer tc.Close()

	inv := NewEventDrivenInvalidation(nil, tc)

	// Multiple stops should not panic
	inv.Stop()
	inv.Stop()
	inv.Stop()
}

func TestEventDrivenInvalidation_ConcurrentRuleAccess(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 100,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	defer tc.Close()

	inv := NewEventDrivenInvalidation(nil, tc)

	var wg sync.WaitGroup
	numGoroutines := 10

	// Concurrent AddRule
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			rule := InvalidationRule{
				EventType:  events.EventType("custom.event." + string(rune('0'+idx))),
				KeyPattern: "key:" + string(rune('0'+idx)) + ":*",
			}
			inv.AddRule(rule)
		}(i)
	}

	// Concurrent ShouldInvalidate
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			event := events.NewEvent(events.EventProviderHealthChanged, "test", map[string]interface{}{
				"name": "provider",
			})
			inv.ShouldInvalidate(event)
		}()
	}

	wg.Wait()
}

func TestCompositeInvalidation_ShouldInvalidate_CombinesStrategies(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 100,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	defer tc.Close()

	tag := NewTagBasedInvalidation()
	event := NewEventDrivenInvalidation(nil, tc)

	// Add a custom rule to event-driven
	event.AddRule(InvalidationRule{
		EventType: events.EventSystemError,
		Handler: func(e *events.Event) []string {
			return []string{"from:event"}
		},
	})

	composite := NewCompositeInvalidation(tag, event)

	// Create event
	e := events.NewEvent(events.EventSystemError, "test", nil)

	keys := composite.ShouldInvalidate(e)
	assert.Contains(t, keys, "from:event")
}

func TestCompositeInvalidation_MultipleSameKeys(t *testing.T) {
	// Create strategies that return duplicate keys
	strategy1 := &mockStrategy{keys: []string{"key1", "key2"}}
	strategy2 := &mockStrategy{keys: []string{"key2", "key3"}}

	composite := NewCompositeInvalidation(strategy1, strategy2)

	event := events.NewEvent(events.EventSystemStartup, "test", nil)
	keys := composite.ShouldInvalidate(event)

	// Should dedupe
	assert.Len(t, keys, 3)
}

func TestTagBasedInvalidation_AddTag_Concurrency(t *testing.T) {
	inv := NewTagBasedInvalidation()

	var wg sync.WaitGroup
	numGoroutines := 50

	// Concurrent AddTag
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			inv.AddTag("key"+string(rune('0'+idx%10)), "tag"+string(rune('0'+idx%5)))
		}(i)
	}

	// Concurrent InvalidateByTag
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			inv.InvalidateByTag("tag" + string(rune('0'+idx%5)))
		}(i)
	}

	wg.Wait()
}

func TestTagBasedInvalidation_GetKeys_EmptyTag(t *testing.T) {
	inv := NewTagBasedInvalidation()

	keys := inv.GetKeys("nonexistent")
	assert.Nil(t, keys)
}

func TestTagBasedInvalidation_RemoveKey_Thorough(t *testing.T) {
	inv := NewTagBasedInvalidation()

	// Add key with multiple tags
	inv.AddTag("key1", "tag1", "tag2", "tag3")
	inv.AddTag("key2", "tag1")

	// Verify setup
	assert.Len(t, inv.GetTags("key1"), 3)
	assert.Len(t, inv.GetKeys("tag1"), 2)

	// Remove key1
	inv.RemoveKey("key1")

	// key1 should have no tags
	assert.Empty(t, inv.GetTags("key1"))

	// tag1 should only have key2
	assert.Len(t, inv.GetKeys("tag1"), 1)
	assert.Contains(t, inv.GetKeys("tag1"), "key2")

	// tag2 and tag3 should be gone (only had key1)
	assert.Nil(t, inv.GetKeys("tag2"))
	assert.Nil(t, inv.GetKeys("tag3"))
}

func TestInvalidationRule_Handler_NilPayload(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 100,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	defer tc.Close()

	inv := NewEventDrivenInvalidation(nil, tc)

	// Event with nil payload
	event := events.NewEvent(events.EventProviderHealthChanged, "test", nil)

	keys := inv.ShouldInvalidate(event)
	// Should still get KeyPattern
	assert.Contains(t, keys, "provider:*")
}

func TestInvalidationRule_Handler_WrongPayloadType(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 100,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	defer tc.Close()

	inv := NewEventDrivenInvalidation(nil, tc)

	// Event with string payload instead of map
	event := events.NewEvent(events.EventProviderHealthChanged, "test", "invalid payload")

	keys := inv.ShouldInvalidate(event)
	// Handler should handle gracefully, still get KeyPattern
	assert.Contains(t, keys, "provider:*")
}

func TestInvalidationRule_Handler_MissingField(t *testing.T) {
	config := &TieredCacheConfig{
		L1MaxSize: 100,
		L1TTL:     time.Minute,
		EnableL1:  true,
	}
	tc := NewTieredCache(nil, config)
	defer tc.Close()

	inv := NewEventDrivenInvalidation(nil, tc)

	// Event with payload missing required field
	event := events.NewEvent(events.EventProviderHealthChanged, "test", map[string]interface{}{
		"other_field": "value",
	})

	keys := inv.ShouldInvalidate(event)
	// Handler returns nil for missing name, but still get KeyPattern
	assert.Contains(t, keys, "provider:*")
}

// mockStrategy is a test double for InvalidationStrategy
type mockStrategy struct {
	keys []string
}

func (m *mockStrategy) ShouldInvalidate(event *events.Event) []string {
	return m.keys
}

func (m *mockStrategy) Name() string {
	return "mock"
}
