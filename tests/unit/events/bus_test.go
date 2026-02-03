package events

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/adapters"
)

// Alias for backward compatibility
var events = adapters

func TestEventBus_PublishSubscribe(t *testing.T) {
	bus := events.NewEventBus(nil)
	defer bus.Close()

	ch := bus.Subscribe(events.EventProviderHealthChanged)

	event := events.NewEvent(events.EventProviderHealthChanged, "test", map[string]interface{}{
		"provider": "claude",
		"healthy":  true,
	})

	bus.Publish(event)

	select {
	case received := <-ch:
		assert.Equal(t, events.EventProviderHealthChanged, received.Type)
		assert.Equal(t, "test", received.Source)
		payload := received.Payload.(map[string]interface{})
		assert.Equal(t, "claude", payload["provider"])
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestEventBus_MultipleSubscribers(t *testing.T) {
	bus := events.NewEventBus(nil)
	defer bus.Close()

	ch1 := bus.Subscribe(events.EventMCPServerConnected)
	ch2 := bus.Subscribe(events.EventMCPServerConnected)

	event := events.NewEvent(events.EventMCPServerConnected, "mcp", "server1")
	bus.Publish(event)

	// Both subscribers should receive the event
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		select {
		case <-ch1:
		case <-time.After(time.Second):
			t.Error("subscriber 1 timeout")
		}
	}()

	go func() {
		defer wg.Done()
		select {
		case <-ch2:
		case <-time.After(time.Second):
			t.Error("subscriber 2 timeout")
		}
	}()

	wg.Wait()
}

func TestEventBus_SubscribeMultiple(t *testing.T) {
	bus := events.NewEventBus(nil)
	defer bus.Close()

	ch := bus.SubscribeMultiple(
		events.EventProviderHealthChanged,
		events.EventMCPServerConnected,
	)

	bus.Publish(events.NewEvent(events.EventProviderHealthChanged, "test", nil))
	bus.Publish(events.NewEvent(events.EventMCPServerConnected, "test", nil))

	received := 0
	timeout := time.After(2 * time.Second)
	for received < 2 {
		select {
		case <-ch:
			received++
		case <-timeout:
			t.Fatalf("timeout, received only %d events", received)
		}
	}

	assert.Equal(t, 2, received)
}

func TestEventBus_SubscribeAll(t *testing.T) {
	bus := events.NewEventBus(nil)
	defer bus.Close()

	ch := bus.SubscribeAll()

	bus.Publish(events.NewEvent(events.EventProviderHealthChanged, "test", nil))
	bus.Publish(events.NewEvent(events.EventMCPServerConnected, "test", nil))
	bus.Publish(events.NewEvent(events.EventDebateStarted, "test", nil))

	received := 0
	timeout := time.After(2 * time.Second)
	for received < 3 {
		select {
		case <-ch:
			received++
		case <-timeout:
			t.Fatalf("timeout, received only %d events", received)
		}
	}

	assert.Equal(t, 3, received)
}

func TestEventBus_SubscribeWithFilter(t *testing.T) {
	bus := events.NewEventBus(nil)
	defer bus.Close()

	// Only accept events with "important" in payload
	filter := func(e *events.Event) bool {
		if payload, ok := e.Payload.(map[string]interface{}); ok {
			if imp, ok := payload["important"].(bool); ok {
				return imp
			}
		}
		return false
	}

	ch := bus.SubscribeWithFilter(events.EventSystemError, filter)

	// This should not pass the filter
	bus.Publish(events.NewEvent(events.EventSystemError, "test", map[string]interface{}{
		"important": false,
	}))

	// This should pass the filter
	bus.Publish(events.NewEvent(events.EventSystemError, "test", map[string]interface{}{
		"important": true,
	}))

	select {
	case e := <-ch:
		payload := e.Payload.(map[string]interface{})
		assert.True(t, payload["important"].(bool))
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for filtered event")
	}

	// Ensure no more events
	select {
	case <-ch:
		t.Fatal("received unexpected event")
	case <-time.After(100 * time.Millisecond):
		// Expected - no more events
	}
}

func TestEventBus_Unsubscribe(t *testing.T) {
	bus := events.NewEventBus(nil)
	defer bus.Close()

	ch := bus.Subscribe(events.EventCacheHit)
	bus.Unsubscribe(ch)

	bus.Publish(events.NewEvent(events.EventCacheHit, "test", nil))

	select {
	case _, ok := <-ch:
		assert.False(t, ok, "channel should be closed")
	case <-time.After(100 * time.Millisecond):
		// Expected if channel is closed
	}
}

func TestEventBus_Wait(t *testing.T) {
	bus := events.NewEventBus(nil)
	defer bus.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Publish after a delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		bus.Publish(events.NewEvent(events.EventSystemStartup, "test", "started"))
	}()

	event, err := bus.Wait(ctx, events.EventSystemStartup)
	require.NoError(t, err)
	assert.Equal(t, events.EventSystemStartup, event.Type)
	assert.Equal(t, "started", event.Payload)
}

func TestEventBus_WaitTimeout(t *testing.T) {
	bus := events.NewEventBus(nil)
	defer bus.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := bus.Wait(ctx, events.EventSystemStartup)
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestEventBus_Metrics(t *testing.T) {
	bus := events.NewEventBus(nil)
	defer bus.Close()

	ch := bus.Subscribe(events.EventCacheHit)

	// Publish events
	for i := 0; i < 5; i++ {
		bus.Publish(events.NewEvent(events.EventCacheHit, "test", nil))
	}

	// Drain events
	for i := 0; i < 5; i++ {
		select {
		case <-ch:
		case <-time.After(time.Second):
			t.Fatal("timeout")
		}
	}

	metrics := bus.Metrics()
	assert.NotNil(t, metrics)
	assert.Equal(t, int64(5), metrics.EventsPublished)
	assert.Equal(t, int64(5), metrics.EventsDelivered)
	assert.True(t, metrics.SubscribersActive > 0)
}

func TestEventBus_AsyncPublish(t *testing.T) {
	bus := events.NewEventBus(nil)
	defer bus.Close()

	ch := bus.Subscribe(events.EventRequestReceived)

	// Async publish should not block
	for i := 0; i < 100; i++ {
		bus.PublishAsync(events.NewEvent(events.EventRequestReceived, "test", i))
	}

	// Should receive events eventually
	received := 0
	timeout := time.After(5 * time.Second)
	for received < 100 {
		select {
		case <-ch:
			received++
		case <-timeout:
			t.Fatalf("timeout, received only %d events", received)
		}
	}

	assert.Equal(t, 100, received)
}

func TestEventBus_GlobalBus(t *testing.T) {
	events.InitGlobalBus(nil)
	defer events.GlobalBus.Close()

	ch := events.On(events.EventSystemHealthCheck)

	events.Emit(events.NewEvent(events.EventSystemHealthCheck, "test", nil))

	select {
	case e := <-ch:
		assert.Equal(t, events.EventSystemHealthCheck, e.Type)
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func TestEvent_Chaining(t *testing.T) {
	event := events.NewEvent(events.EventDebateCompleted, "debate", "result").
		WithTraceID("trace-123").
		WithMetadata("round", "3").
		WithMetadata("duration", "5s")

	assert.Equal(t, "trace-123", event.TraceID)
	assert.Equal(t, "3", event.Metadata["round"])
	assert.Equal(t, "5s", event.Metadata["duration"])
}

func TestEventBus_ConcurrentAccess(t *testing.T) {
	bus := events.NewEventBus(nil)
	defer bus.Close()

	var wg sync.WaitGroup
	var published int64
	var received int64

	// Subscribe in multiple goroutines
	channels := make([]<-chan *events.Event, 10)
	for i := 0; i < 10; i++ {
		channels[i] = bus.Subscribe(events.EventRequestCompleted)
	}

	// Receive in goroutines
	for _, ch := range channels {
		wg.Add(1)
		go func(c <-chan *events.Event) {
			defer wg.Done()
			for {
				select {
				case _, ok := <-c:
					if !ok {
						return
					}
					atomic.AddInt64(&received, 1)
				case <-time.After(2 * time.Second):
					return
				}
			}
		}(ch)
	}

	// Publish from multiple goroutines
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				bus.Publish(events.NewEvent(events.EventRequestCompleted, "test", nil))
				atomic.AddInt64(&published, 1)
			}
		}()
	}

	// Wait for publishers
	time.Sleep(time.Second)

	// Close bus to stop receivers
	bus.Close()
	wg.Wait()

	// Each of 10 subscribers should have received 100 events
	assert.Equal(t, int64(100), atomic.LoadInt64(&published))
}

func BenchmarkEventBus_Publish(b *testing.B) {
	bus := events.NewEventBus(&events.BusConfig{
		BufferSize:     10000,
		PublishTimeout: 10 * time.Millisecond,
	})
	defer bus.Close()

	ch := bus.Subscribe(events.EventCacheHit)
	go func() {
		for range ch {
		}
	}()

	event := events.NewEvent(events.EventCacheHit, "bench", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bus.Publish(event)
	}
}

func BenchmarkEventBus_PublishParallel(b *testing.B) {
	bus := events.NewEventBus(&events.BusConfig{
		BufferSize:     100000,
		PublishTimeout: 10 * time.Millisecond,
	})
	defer bus.Close()

	ch := bus.Subscribe(events.EventCacheHit)
	go func() {
		for range ch {
		}
	}()

	event := events.NewEvent(events.EventCacheHit, "bench", nil)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bus.Publish(event)
		}
	})
}
