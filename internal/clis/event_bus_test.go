// Package clis provides CLI agent integration for HelixAgent.
package clis

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEventBus(t *testing.T) {
	eb := NewEventBus()
	require.NotNil(t, eb)
	assert.NotNil(t, eb.subscribers)
	assert.NotNil(t, eb.topics)
	assert.NotNil(t, eb.eventCh)
	assert.NotNil(t, eb.ctx)

	err := eb.Close()
	assert.NoError(t, err)
}

func TestEventBus_Subscribe(t *testing.T) {
	eb := NewEventBus()
	defer eb.Close()

	sub := eb.Subscribe(EventTypeStatus, 10)
	require.NotNil(t, sub)
	assert.NotEmpty(t, sub.ID)
	assert.Equal(t, EventTypeStatus, sub.EventType)
	assert.NotNil(t, sub.Ch)
	assert.Equal(t, 10, cap(sub.Ch))
}

func TestEventBus_SubscribeTopic(t *testing.T) {
	eb := NewEventBus()
	defer eb.Close()

	sub := eb.SubscribeTopic("test-topic", 10)
	require.NotNil(t, sub)
	assert.NotEmpty(t, sub.ID)
	assert.Equal(t, "test-topic", sub.Topic)
	assert.NotNil(t, sub.Ch)
}

func TestEventBus_SubscribeWildcard(t *testing.T) {
	eb := NewEventBus()
	defer eb.Close()

	sub := eb.SubscribeWildcard(10)
	require.NotNil(t, sub)
	assert.NotEmpty(t, sub.ID)
	assert.NotNil(t, sub.Ch)
}

func TestEventBus_SubscribeFiltered(t *testing.T) {
	eb := NewEventBus()
	defer eb.Close()

	filter := func(e *Event) bool {
		return e.Source == "test-source"
	}

	sub := eb.SubscribeFiltered(EventTypeStatus, 10, filter)
	require.NotNil(t, sub)
	assert.NotNil(t, sub.Filter)
	assert.NotNil(t, sub.Ch)

	// Test filter function
	event := &Event{Source: "test-source"}
	assert.True(t, sub.Filter(event))

	event2 := &Event{Source: "other-source"}
	assert.False(t, sub.Filter(event2))
}

func TestEventBus_Unsubscribe(t *testing.T) {
	eb := NewEventBus()
	defer eb.Close()

	sub := eb.Subscribe(EventTypeStatus, 10)

	// Verify subscription exists
	eb.mu.RLock()
	assert.Len(t, eb.subscribers[EventTypeStatus], 1)
	eb.mu.RUnlock()

	// Unsubscribe
	eb.Unsubscribe(sub)

	// Verify subscription removed
	eb.mu.RLock()
	assert.Len(t, eb.subscribers[EventTypeStatus], 0)
	eb.mu.RUnlock()
}

func TestEventBus_Publish(t *testing.T) {
	eb := NewEventBus()
	defer eb.Close()

	sub := eb.Subscribe(EventTypeStatus, 10)

	event := &Event{
		ID:        uuid.New(),
		Type:      EventTypeStatus,
		Source:    "test",
		Payload:   map[string]interface{}{"status": "active"},
		Timestamp: time.Now(),
	}

	eb.Publish(event)

	// Wait for event to be dispatched
	select {
	case received := <-sub.Ch:
		assert.Equal(t, event.ID, received.ID)
		assert.Equal(t, event.Type, received.Type)
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for event")
	}
}

func TestEventBus_PublishSync(t *testing.T) {
	eb := NewEventBus()
	defer eb.Close()

	sub := eb.Subscribe(EventTypeStatus, 10)

	event := &Event{
		ID:        uuid.New(),
		Type:      EventTypeStatus,
		Source:    "test",
		Payload:   map[string]interface{}{"data": "test-payload"},
		Timestamp: time.Now(),
	}

	eb.PublishSync(event)

	// Event should be received immediately (synchronous)
	select {
	case received := <-sub.Ch:
		assert.Equal(t, event.ID, received.ID)
	default:
		t.Fatal("Event not received synchronously")
	}
}

func TestEventBus_PublishToMultipleSubscribers(t *testing.T) {
	eb := NewEventBus()
	defer eb.Close()

	sub1 := eb.Subscribe(EventTypeStatus, 10)
	sub2 := eb.Subscribe(EventTypeStatus, 10)

	event := &Event{
		ID:        uuid.New(),
		Type:      EventTypeStatus,
		Source:    "test",
		Timestamp: time.Now(),
	}

	eb.PublishSync(event)

	// Both subscribers should receive the event
	select {
	case <-sub1.Ch:
		// Good
	case <-time.After(time.Second):
		t.Fatal("Sub1 didn't receive event")
	}

	select {
	case <-sub2.Ch:
		// Good
	case <-time.After(time.Second):
		t.Fatal("Sub2 didn't receive event")
	}
}

func TestEventBus_PublishWithFilter(t *testing.T) {
	eb := NewEventBus()
	defer eb.Close()

	// Subscriber with filter - only accepts events from "allowed-source"
	sub := eb.SubscribeFiltered(EventTypeStatus, 10, func(e *Event) bool {
		return e.Source == "allowed-source"
	})

	eventID1 := uuid.New()
	eventID2 := uuid.New()

	// Publish matching event
	eb.PublishSync(&Event{
		ID:        eventID1,
		Type:      EventTypeStatus,
		Source:    "allowed-source",
		Timestamp: time.Now(),
	})

	// Publish non-matching event
	eb.PublishSync(&Event{
		ID:        eventID2,
		Type:      EventTypeStatus,
		Source:    "blocked-source",
		Timestamp: time.Now(),
	})

	// Should only receive the matching event
	select {
	case received := <-sub.Ch:
		assert.Equal(t, eventID1, received.ID)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Should have received matching event")
	}

	// Should not receive the blocked event
	select {
	case <-sub.Ch:
		t.Fatal("Should not have received filtered event")
	case <-time.After(100 * time.Millisecond):
		// Good - timeout means no event received
	}
}

func TestEventBus_TopicSubscription(t *testing.T) {
	eb := NewEventBus()
	defer eb.Close()

	topicSub := eb.SubscribeTopic("test-topic", 10)
	typeSub := eb.Subscribe(EventTypeStatus, 10)

	// Publish event with topic in metadata
	event := &Event{
		ID:        uuid.New(),
		Type:      EventTypeStatus,
		Source:    "test",
		Metadata:  map[string]interface{}{"topic": "test-topic"},
		Timestamp: time.Now(),
	}

	eb.PublishSync(event)

	// Topic subscriber should receive
	select {
	case <-topicSub.Ch:
		// Good
	case <-time.After(time.Second):
		t.Fatal("Topic subscriber didn't receive event")
	}

	// Type subscriber should also receive (it subscribes to all Status events)
	select {
	case <-typeSub.Ch:
		// Good
	case <-time.After(time.Second):
		t.Fatal("Type subscriber didn't receive event")
	}
}

func TestEventBus_WildcardSubscription(t *testing.T) {
	eb := NewEventBus()
	defer eb.Close()

	wildcardSub := eb.SubscribeWildcard(10)
	statusSub := eb.Subscribe(EventTypeStatus, 10)

	// Publish different event types
	eb.PublishSync(&Event{ID: uuid.New(), Type: EventTypeStatus, Timestamp: time.Now()})
	eb.PublishSync(&Event{ID: uuid.New(), Type: EventTypeProgress, Timestamp: time.Now()})
	eb.PublishSync(&Event{ID: uuid.New(), Type: EventTypeError, Timestamp: time.Now()})

	// Wildcard should receive all 3
	count := 0
	done := time.After(time.Second)
loop:
	for {
		select {
		case <-wildcardSub.Ch:
			count++
			if count == 3 {
				break loop
			}
		case <-done:
			break loop
		}
	}
	assert.Equal(t, 3, count)

	// Status subscriber should only receive 1
	select {
	case <-statusSub.Ch:
		// Good
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Status subscriber should have received 1 event")
	}

	// Should not receive more
	select {
	case <-statusSub.Ch:
		t.Fatal("Status subscriber should only receive 1 event")
	case <-time.After(100 * time.Millisecond):
		// Good
	}
}

func TestEventBus_OnceSubscription(t *testing.T) {
	eb := NewEventBus()
	defer eb.Close()

	sub := eb.Subscribe(EventTypeStatus, 10)
	sub.Once = true

	eventID1 := uuid.New()
	eventID2 := uuid.New()

	// Publish first event (async to allow unsubscribe to process)
	eb.Publish(&Event{ID: eventID1, Type: EventTypeStatus, Timestamp: time.Now()})

	// Wait for first event to be processed and unsubscribe to complete
	time.Sleep(200 * time.Millisecond)

	// Publish second event - should not be received
	eb.Publish(&Event{ID: eventID2, Type: EventTypeStatus, Timestamp: time.Now()})

	// Give time for processing
	time.Sleep(100 * time.Millisecond)

	// Should only receive first event
	select {
	case received := <-sub.Ch:
		assert.Equal(t, eventID1, received.ID)
	case <-time.After(time.Second):
		t.Fatal("Should have received first event")
	}

	// Give more time for unsubscribe to complete
	time.Sleep(300 * time.Millisecond)

	// Try to receive from channel - should not get second event
	select {
	case evt, ok := <-sub.Ch:
		if ok {
			t.Fatalf("Should not receive more events after Once, got: %v", evt)
		}
		// Channel is closed as expected
	default:
		// Channel not closed yet but no events - this is acceptable
		t.Log("Channel has no more events (Once subscription working)")
	}
}

func TestEventBus_HighThroughput(t *testing.T) {
	eb := NewEventBus()
	defer eb.Close()

	sub := eb.Subscribe(EventTypeStatus, 1000)

	const eventCount = 1000
	var received int64

	// Start consumer
	go func() {
		for range sub.Ch {
			atomic.AddInt64(&received, 1)
		}
	}()

	// Publish events
	start := time.Now()
	for i := 0; i < eventCount; i++ {
		eb.Publish(&Event{
			ID:        uuid.New(),
			Type:      EventTypeStatus,
			Timestamp: time.Now(),
		})
	}

	// Wait for all events to be processed
	time.Sleep(500 * time.Millisecond)
	elapsed := time.Since(start)

	assert.Equal(t, int64(eventCount), atomic.LoadInt64(&received))
	t.Logf("Processed %d events in %v (%.0f events/sec)",
		eventCount, elapsed, float64(eventCount)/elapsed.Seconds())
}

func TestEventBus_ConcurrentPublishSubscribe(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	eb := NewEventBus()
	defer eb.Close()

	const numSubscribers = 10
	const numPublishers = 5
	const eventsPerPublisher = 100

	var subs []*Subscription
	var counters []int64

	// Create subscribers with larger buffers
	for i := 0; i < numSubscribers; i++ {
		sub := eb.Subscribe(EventTypeStatus, 500)
		subs = append(subs, sub)
		counters = append(counters, 0)

		// Start consumer
		go func(idx int, s *Subscription) {
			for range s.Ch {
				atomic.AddInt64(&counters[idx], 1)
			}
		}(i, sub)
	}

	// Start publishers - use PublishSync for reliable delivery
	var wg sync.WaitGroup
	for i := 0; i < numPublishers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < eventsPerPublisher; j++ {
				eb.PublishSync(&Event{
					ID:        uuid.New(),
					Type:      EventTypeStatus,
					Timestamp: time.Now(),
				})
			}
		}(i)
	}

	wg.Wait()
	time.Sleep(300 * time.Millisecond)

	// Verify all subscribers received all events
	totalEvents := int64(numPublishers * eventsPerPublisher)
	for i, counter := range counters {
		count := atomic.LoadInt64(&counter)
		assert.Equal(t, totalEvents, count, "Subscriber %d received wrong count", i)
	}
}

func TestEventBus_GetStats(t *testing.T) {
	eb := NewEventBus()
	defer eb.Close()

	// Initially empty
	stats := eb.GetStats()
	assert.Equal(t, 0, stats["total_subscriptions"])

	// Add subscribers
	eb.Subscribe(EventTypeStatus, 10)
	eb.Subscribe(EventTypeProgress, 10)
	eb.SubscribeTopic("topic1", 10)
	eb.SubscribeWildcard(10)

	stats = eb.GetStats()
	assert.Equal(t, 4, stats["total_subscriptions"])
	assert.Equal(t, 2, stats["event_types"])
	assert.Equal(t, 1, stats["topics"])
	assert.Equal(t, 1, stats["wildcards"])
}

func TestEventBus_Close(t *testing.T) {
	eb := NewEventBus()

	sub1 := eb.Subscribe(EventTypeStatus, 10)
	sub2 := eb.SubscribeTopic("test", 10)
	sub3 := eb.SubscribeWildcard(10)

	err := eb.Close()
	assert.NoError(t, err)

	// All channels should be closed
	_, ok1 := <-sub1.Ch
	_, ok2 := <-sub2.Ch
	_, ok3 := <-sub3.Ch

	assert.False(t, ok1, "Sub1 channel should be closed")
	assert.False(t, ok2, "Sub2 channel should be closed")
	assert.False(t, ok3, "Sub3 channel should be closed")
}

// Benchmarks

func BenchmarkEventBus_Publish(b *testing.B) {
	eb := NewEventBus()
	defer eb.Close()

	sub := eb.Subscribe(EventTypeStatus, b.N)

	// Drain the channel
	go func() {
		for range sub.Ch {
		}
	}()

	event := &Event{
		Type:      EventTypeStatus,
		Timestamp: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		eb.Publish(event)
	}
}

func BenchmarkEventBus_PublishSync(b *testing.B) {
	eb := NewEventBus()
	defer eb.Close()

	sub := eb.Subscribe(EventTypeStatus, b.N)

	// Drain the channel
	go func() {
		for range sub.Ch {
		}
	}()

	event := &Event{
		Type:      EventTypeStatus,
		Timestamp: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		eb.PublishSync(event)
	}
}

func BenchmarkEventBus_MultipleSubscribers(b *testing.B) {
	eb := NewEventBus()
	defer eb.Close()

	// Create 10 subscribers
	for i := 0; i < 10; i++ {
		sub := eb.Subscribe(EventTypeStatus, b.N)
		go func() {
			for range sub.Ch {
			}
		}()
	}

	event := &Event{
		Type:      EventTypeStatus,
		Timestamp: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		eb.Publish(event)
	}
}
