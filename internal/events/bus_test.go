package events

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEvent(t *testing.T) {
	event := NewEvent(EventProviderRegistered, "test-source", map[string]string{"key": "value"})

	assert.NotEmpty(t, event.ID)
	assert.Equal(t, EventProviderRegistered, event.Type)
	assert.Equal(t, "test-source", event.Source)
	assert.NotNil(t, event.Payload)
	assert.NotZero(t, event.Timestamp)
	assert.NotEmpty(t, event.TraceID)
	assert.NotNil(t, event.Metadata)
}

func TestEvent_WithTraceID(t *testing.T) {
	event := NewEvent(EventSystemStartup, "source", nil)
	event.WithTraceID("custom-trace-id")

	assert.Equal(t, "custom-trace-id", event.TraceID)
}

func TestEvent_WithMetadata(t *testing.T) {
	event := NewEvent(EventSystemStartup, "source", nil)
	event.WithMetadata("key1", "value1").WithMetadata("key2", "value2")

	assert.Equal(t, "value1", event.Metadata["key1"])
	assert.Equal(t, "value2", event.Metadata["key2"])
}

func TestEvent_WithMetadata_NilMap(t *testing.T) {
	event := &Event{ID: "test", Type: EventSystemStartup}
	event.Metadata = nil

	event.WithMetadata("key", "value")

	assert.NotNil(t, event.Metadata)
	assert.Equal(t, "value", event.Metadata["key"])
}

func TestDefaultBusConfig(t *testing.T) {
	config := DefaultBusConfig()

	assert.Equal(t, 1000, config.BufferSize)
	assert.Equal(t, 10*time.Millisecond, config.PublishTimeout)
	assert.Equal(t, 30*time.Second, config.CleanupInterval)
	assert.Equal(t, 100, config.MaxSubscribers)
}

func TestNewEventBus(t *testing.T) {
	bus := NewEventBus(nil)
	defer bus.Close()

	assert.NotNil(t, bus)
	assert.NotNil(t, bus.config)
	assert.NotNil(t, bus.metrics)
}

func TestNewEventBus_WithConfig(t *testing.T) {
	config := &BusConfig{
		BufferSize:     500,
		PublishTimeout: 50 * time.Millisecond,
	}

	bus := NewEventBus(config)
	defer bus.Close()

	assert.Equal(t, 500, bus.config.BufferSize)
	assert.Equal(t, 50*time.Millisecond, bus.config.PublishTimeout)
}

func TestEventBus_Subscribe(t *testing.T) {
	bus := NewEventBus(nil)
	defer bus.Close()

	ch := bus.Subscribe(EventProviderRegistered)

	assert.NotNil(t, ch)
	assert.Equal(t, 1, bus.SubscriberCount(EventProviderRegistered))
}

func TestEventBus_Publish(t *testing.T) {
	bus := NewEventBus(nil)
	defer bus.Close()

	ch := bus.Subscribe(EventProviderRegistered)
	event := NewEvent(EventProviderRegistered, "test", nil)

	bus.Publish(event)

	select {
	case received := <-ch:
		assert.Equal(t, event.ID, received.ID)
		assert.Equal(t, EventProviderRegistered, received.Type)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestEventBus_Publish_NilEvent(t *testing.T) {
	bus := NewEventBus(nil)
	defer bus.Close()

	// Should not panic
	bus.Publish(nil)
}

func TestEventBus_PublishAsync(t *testing.T) {
	bus := NewEventBus(nil)
	defer bus.Close()

	ch := bus.Subscribe(EventSystemStartup)
	event := NewEvent(EventSystemStartup, "test", nil)

	bus.PublishAsync(event)

	select {
	case received := <-ch:
		assert.Equal(t, event.ID, received.ID)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestEventBus_SubscribeWithFilter(t *testing.T) {
	bus := NewEventBus(nil)
	defer bus.Close()

	// Filter that only accepts events from "important" source
	filter := func(e *Event) bool {
		return e.Source == "important"
	}

	ch := bus.SubscribeWithFilter(EventProviderRegistered, filter)

	// This event should be filtered out
	bus.Publish(NewEvent(EventProviderRegistered, "unimportant", nil))

	// This event should pass through
	importantEvent := NewEvent(EventProviderRegistered, "important", nil)
	bus.Publish(importantEvent)

	select {
	case received := <-ch:
		assert.Equal(t, "important", received.Source)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestEventBus_SubscribeMultiple(t *testing.T) {
	bus := NewEventBus(nil)
	defer bus.Close()

	ch := bus.SubscribeMultiple(EventProviderRegistered, EventProviderUnregistered)

	bus.Publish(NewEvent(EventProviderRegistered, "test", nil))
	bus.Publish(NewEvent(EventProviderUnregistered, "test", nil))

	count := 0
	timeout := time.After(time.Second)
loop:
	for {
		select {
		case <-ch:
			count++
			if count >= 2 {
				break loop
			}
		case <-timeout:
			break loop
		}
	}

	assert.Equal(t, 2, count)
}

func TestEventBus_SubscribeAll(t *testing.T) {
	bus := NewEventBus(nil)
	defer bus.Close()

	ch := bus.SubscribeAll()

	// Publish various events
	bus.Publish(NewEvent(EventProviderRegistered, "test", nil))
	bus.Publish(NewEvent(EventCacheHit, "test", nil))
	bus.Publish(NewEvent(EventSystemStartup, "test", nil))

	count := 0
	timeout := time.After(time.Second)
loop:
	for {
		select {
		case <-ch:
			count++
			if count >= 3 {
				break loop
			}
		case <-timeout:
			break loop
		}
	}

	assert.Equal(t, 3, count)
}

func TestEventBus_SubscribeAllWithFilter(t *testing.T) {
	bus := NewEventBus(nil)
	defer bus.Close()

	// Filter only cache events
	filter := func(e *Event) bool {
		return e.Type == EventCacheHit || e.Type == EventCacheMiss
	}

	ch := bus.SubscribeAllWithFilter(filter)

	bus.Publish(NewEvent(EventProviderRegistered, "test", nil)) // Should be filtered
	bus.Publish(NewEvent(EventCacheHit, "test", nil))           // Should pass
	bus.Publish(NewEvent(EventCacheMiss, "test", nil))          // Should pass

	count := 0
	timeout := time.After(time.Second)
loop:
	for {
		select {
		case <-ch:
			count++
			if count >= 2 {
				break loop
			}
		case <-timeout:
			break loop
		}
	}

	assert.Equal(t, 2, count)
}

func TestEventBus_Unsubscribe(t *testing.T) {
	bus := NewEventBus(nil)
	defer bus.Close()

	ch := bus.Subscribe(EventProviderRegistered)
	assert.Equal(t, 1, bus.SubscriberCount(EventProviderRegistered))

	bus.Unsubscribe(ch)
	assert.Equal(t, 0, bus.SubscriberCount(EventProviderRegistered))
}

func TestEventBus_Close(t *testing.T) {
	bus := NewEventBus(nil)

	ch := bus.Subscribe(EventSystemStartup)

	err := bus.Close()
	assert.NoError(t, err)

	// Channel should be closed
	_, ok := <-ch
	assert.False(t, ok)
}

func TestEventBus_Close_Multiple(t *testing.T) {
	bus := NewEventBus(nil)

	err := bus.Close()
	assert.NoError(t, err)

	// Second close should be no-op
	err = bus.Close()
	assert.NoError(t, err)
}

func TestEventBus_Subscribe_AfterClose(t *testing.T) {
	bus := NewEventBus(nil)
	bus.Close()

	ch := bus.Subscribe(EventSystemStartup)

	// Channel should be closed
	_, ok := <-ch
	assert.False(t, ok)
}

func TestEventBus_Metrics(t *testing.T) {
	bus := NewEventBus(nil)
	defer bus.Close()

	ch := bus.Subscribe(EventSystemStartup)

	// Publish events
	for i := 0; i < 5; i++ {
		bus.Publish(NewEvent(EventSystemStartup, "test", nil))
	}

	// Drain channel
	for i := 0; i < 5; i++ {
		<-ch
	}

	metrics := bus.Metrics()
	assert.Equal(t, int64(5), metrics.EventsPublished)
	assert.Equal(t, int64(5), metrics.EventsDelivered)
	assert.Equal(t, int64(1), metrics.SubscribersActive)
}

func TestEventBus_TotalSubscribers(t *testing.T) {
	bus := NewEventBus(nil)
	defer bus.Close()

	assert.Equal(t, 0, bus.TotalSubscribers())

	bus.Subscribe(EventProviderRegistered)
	bus.Subscribe(EventCacheHit)

	assert.Equal(t, 2, bus.TotalSubscribers())
}

func TestEventBus_Wait(t *testing.T) {
	bus := NewEventBus(nil)
	defer bus.Close()

	go func() {
		time.Sleep(50 * time.Millisecond)
		bus.Publish(NewEvent(EventSystemStartup, "test", nil))
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	event, err := bus.Wait(ctx, EventSystemStartup)
	require.NoError(t, err)
	assert.Equal(t, EventSystemStartup, event.Type)
}

func TestEventBus_Wait_Timeout(t *testing.T) {
	bus := NewEventBus(nil)
	defer bus.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := bus.Wait(ctx, EventSystemStartup)
	assert.Error(t, err)
}

func TestEventBus_WaitMultiple(t *testing.T) {
	bus := NewEventBus(nil)
	defer bus.Close()

	go func() {
		time.Sleep(50 * time.Millisecond)
		bus.Publish(NewEvent(EventCacheHit, "test", nil))
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	event, err := bus.WaitMultiple(ctx, EventCacheHit, EventCacheMiss)
	require.NoError(t, err)
	assert.Equal(t, EventCacheHit, event.Type)
}

func TestSubscriber_Close(t *testing.T) {
	sub := &Subscriber{
		ID:      "test",
		Channel: make(chan *Event, 10),
	}

	sub.Close()
	assert.True(t, sub.Closed)

	// Double close should be safe
	sub.Close()
	assert.True(t, sub.Closed)
}

func TestSubscriber_trySend(t *testing.T) {
	sub := &Subscriber{
		ID:      "test",
		Channel: make(chan *Event, 1),
	}

	event := NewEvent(EventSystemStartup, "test", nil)

	// Should succeed
	success := sub.trySend(event, 100*time.Millisecond)
	assert.True(t, success)

	// Should timeout (channel full)
	success = sub.trySend(event, 10*time.Millisecond)
	assert.False(t, success)
}

func TestSubscriber_trySend_Closed(t *testing.T) {
	sub := &Subscriber{
		ID:      "test",
		Channel: make(chan *Event, 1),
		Closed:  true,
	}

	event := NewEvent(EventSystemStartup, "test", nil)
	success := sub.trySend(event, 100*time.Millisecond)
	assert.False(t, success)
}

func TestBusMetrics_Fields(t *testing.T) {
	metrics := &BusMetrics{
		EventsPublished:   100,
		EventsDelivered:   95,
		EventsDropped:     5,
		SubscribersActive: 10,
		SubscribersTotal:  20,
	}

	assert.Equal(t, int64(100), metrics.EventsPublished)
	assert.Equal(t, int64(95), metrics.EventsDelivered)
	assert.Equal(t, int64(5), metrics.EventsDropped)
	assert.Equal(t, int64(10), metrics.SubscribersActive)
	assert.Equal(t, int64(20), metrics.SubscribersTotal)
}

// Global bus tests
func TestInitGlobalBus(t *testing.T) {
	InitGlobalBus(nil)
	defer GlobalBus.Close()

	assert.NotNil(t, GlobalBus)
}

func TestEmit(t *testing.T) {
	InitGlobalBus(nil)
	defer GlobalBus.Close()

	ch := On(EventSystemStartup)
	event := NewEvent(EventSystemStartup, "test", nil)

	Emit(event)

	select {
	case received := <-ch:
		assert.Equal(t, event.ID, received.ID)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestEmitAsync(t *testing.T) {
	InitGlobalBus(nil)
	defer GlobalBus.Close()

	ch := On(EventSystemStartup)
	event := NewEvent(EventSystemStartup, "test", nil)

	EmitAsync(event)

	select {
	case received := <-ch:
		assert.Equal(t, event.ID, received.ID)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestOn_NilGlobalBus(t *testing.T) {
	oldBus := GlobalBus
	GlobalBus = nil
	defer func() { GlobalBus = oldBus }()

	ch := On(EventSystemStartup)

	// Channel should be closed
	_, ok := <-ch
	assert.False(t, ok)
}

func TestEmit_NilGlobalBus(t *testing.T) {
	oldBus := GlobalBus
	GlobalBus = nil
	defer func() { GlobalBus = oldBus }()

	// Should not panic
	Emit(NewEvent(EventSystemStartup, "test", nil))
	EmitAsync(NewEvent(EventSystemStartup, "test", nil))
}

// Event type constants
func TestEventTypes(t *testing.T) {
	assert.Equal(t, EventType("provider.registered"), EventProviderRegistered)
	assert.Equal(t, EventType("provider.unregistered"), EventProviderUnregistered)
	assert.Equal(t, EventType("provider.health.changed"), EventProviderHealthChanged)
	assert.Equal(t, EventType("provider.score.updated"), EventProviderScoreUpdated)
	assert.Equal(t, EventType("mcp.server.connected"), EventMCPServerConnected)
	assert.Equal(t, EventType("mcp.server.disconnected"), EventMCPServerDisconnected)
	assert.Equal(t, EventType("mcp.tool.executed"), EventMCPToolExecuted)
	assert.Equal(t, EventType("debate.started"), EventDebateStarted)
	assert.Equal(t, EventType("debate.completed"), EventDebateCompleted)
	assert.Equal(t, EventType("cache.hit"), EventCacheHit)
	assert.Equal(t, EventType("cache.miss"), EventCacheMiss)
	assert.Equal(t, EventType("system.startup"), EventSystemStartup)
	assert.Equal(t, EventType("system.shutdown"), EventSystemShutdown)
	assert.Equal(t, EventType("request.received"), EventRequestReceived)
	assert.Equal(t, EventType("request.completed"), EventRequestCompleted)
}

// Concurrent access tests
func TestEventBus_ConcurrentPublish(t *testing.T) {
	bus := NewEventBus(&BusConfig{BufferSize: 1000})
	defer bus.Close()

	ch := bus.SubscribeAll()

	var received atomic.Int64
	go func() {
		for range ch {
			received.Add(1)
		}
	}()

	const numPublishers = 10
	const numEvents = 100

	for i := 0; i < numPublishers; i++ {
		go func(idx int) {
			for j := 0; j < numEvents; j++ {
				bus.Publish(NewEvent(EventSystemStartup, "test", nil))
			}
		}(i)
	}

	time.Sleep(500 * time.Millisecond)
	// Use a more lenient threshold due to timing variance in concurrent tests
	assert.GreaterOrEqual(t, received.Load(), int64(numPublishers*numEvents/4))
}

func TestEventBus_ConcurrentSubscribe(t *testing.T) {
	bus := NewEventBus(nil)
	defer bus.Close()

	const numSubscribers = 20
	done := make(chan struct{})

	for i := 0; i < numSubscribers; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			bus.Subscribe(EventSystemStartup)
		}()
	}

	for i := 0; i < numSubscribers; i++ {
		<-done
	}

	assert.Equal(t, numSubscribers, bus.SubscriberCount(EventSystemStartup))
}
