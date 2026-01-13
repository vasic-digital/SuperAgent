package events

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// EventType represents the type of event
type EventType string

// Standard event types
const (
	// Provider events
	EventProviderRegistered     EventType = "provider.registered"
	EventProviderUnregistered   EventType = "provider.unregistered"
	EventProviderHealthChanged  EventType = "provider.health.changed"
	EventProviderScoreUpdated   EventType = "provider.score.updated"

	// MCP events
	EventMCPServerConnected     EventType = "mcp.server.connected"
	EventMCPServerDisconnected  EventType = "mcp.server.disconnected"
	EventMCPServerHealthChanged EventType = "mcp.server.health.changed"
	EventMCPToolExecuted        EventType = "mcp.tool.executed"
	EventMCPToolFailed          EventType = "mcp.tool.failed"

	// Debate events
	EventDebateStarted          EventType = "debate.started"
	EventDebateRoundStarted     EventType = "debate.round.started"
	EventDebateRoundCompleted   EventType = "debate.round.completed"
	EventDebateCompleted        EventType = "debate.completed"
	EventDebateFailed           EventType = "debate.failed"

	// Cache events
	EventCacheHit               EventType = "cache.hit"
	EventCacheMiss              EventType = "cache.miss"
	EventCacheInvalidated       EventType = "cache.invalidated"
	EventCacheExpired           EventType = "cache.expired"

	// System events
	EventSystemStartup          EventType = "system.startup"
	EventSystemShutdown         EventType = "system.shutdown"
	EventSystemHealthCheck      EventType = "system.health.check"
	EventSystemError            EventType = "system.error"

	// Request events
	EventRequestReceived        EventType = "request.received"
	EventRequestCompleted       EventType = "request.completed"
	EventRequestFailed          EventType = "request.failed"
)

// Event represents a system event
type Event struct {
	ID        string
	Type      EventType
	Source    string
	Payload   interface{}
	Timestamp time.Time
	TraceID   string
	Metadata  map[string]string
}

// NewEvent creates a new event with the given type and payload
func NewEvent(eventType EventType, source string, payload interface{}) *Event {
	return &Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Source:    source,
		Payload:   payload,
		Timestamp: time.Now(),
		TraceID:   uuid.New().String(),
		Metadata:  make(map[string]string),
	}
}

// WithTraceID sets the trace ID and returns the event
func (e *Event) WithTraceID(traceID string) *Event {
	e.TraceID = traceID
	return e
}

// WithMetadata adds metadata and returns the event
func (e *Event) WithMetadata(key, value string) *Event {
	if e.Metadata == nil {
		e.Metadata = make(map[string]string)
	}
	e.Metadata[key] = value
	return e
}

// Subscriber represents an event subscriber
type Subscriber struct {
	ID       string
	Channel  chan *Event
	Filter   func(*Event) bool
	Types    []EventType
	Closed   bool
	mu       sync.RWMutex
}

// Close closes the subscriber channel
func (s *Subscriber) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.Closed {
		s.Closed = true
		close(s.Channel)
	}
}

// trySend attempts to send an event to the subscriber channel
// Returns true if sent, false if closed or would block
func (s *Subscriber) trySend(event *Event, timeout time.Duration) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.Closed {
		return false
	}

	// Non-blocking send with timeout
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case s.Channel <- event:
		return true
	case <-timer.C:
		return false
	}
}

// BusConfig holds configuration for the event bus
type BusConfig struct {
	BufferSize      int           // Buffer size for subscriber channels
	PublishTimeout  time.Duration // Timeout for publishing to subscribers
	CleanupInterval time.Duration // Interval for cleaning up dead subscribers
	MaxSubscribers  int           // Maximum number of subscribers per event type
}

// DefaultBusConfig returns default bus configuration
func DefaultBusConfig() *BusConfig {
	return &BusConfig{
		BufferSize:      1000,
		PublishTimeout:  10 * time.Millisecond,
		CleanupInterval: 30 * time.Second,
		MaxSubscribers:  100,
	}
}

// BusMetrics tracks event bus statistics
type BusMetrics struct {
	EventsPublished    int64
	EventsDelivered    int64
	EventsDropped      int64
	SubscribersActive  int64
	SubscribersTotal   int64
}

// EventBus provides pub/sub for system events
type EventBus struct {
	subscribers map[EventType][]*Subscriber
	allSubs     []*Subscriber // Subscribers to all events
	mu          sync.RWMutex
	config      *BusConfig
	metrics     *BusMetrics
	ctx         context.Context
	cancel      context.CancelFunc
	closed      bool
}

// NewEventBus creates a new event bus
func NewEventBus(config *BusConfig) *EventBus {
	if config == nil {
		config = DefaultBusConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	bus := &EventBus{
		subscribers: make(map[EventType][]*Subscriber),
		allSubs:     make([]*Subscriber, 0),
		config:      config,
		metrics:     &BusMetrics{},
		ctx:         ctx,
		cancel:      cancel,
	}

	// Start cleanup goroutine
	go bus.cleanupLoop()

	return bus
}

// Publish sends an event to all subscribers
func (b *EventBus) Publish(event *Event) {
	if event == nil {
		return
	}

	b.mu.RLock()
	if b.closed {
		b.mu.RUnlock()
		return
	}

	// Get subscribers for this event type
	subs := b.subscribers[event.Type]
	allSubs := b.allSubs
	b.mu.RUnlock()

	atomic.AddInt64(&b.metrics.EventsPublished, 1)

	// Publish to type-specific subscribers
	for _, sub := range subs {
		b.publishToSubscriber(sub, event)
	}

	// Publish to all-event subscribers
	for _, sub := range allSubs {
		b.publishToSubscriber(sub, event)
	}
}

// publishToSubscriber sends event to a single subscriber
func (b *EventBus) publishToSubscriber(sub *Subscriber, event *Event) {
	// Apply filter if present (outside lock)
	if sub.Filter != nil && !sub.Filter(event) {
		return
	}

	// Use the subscriber's trySend method which properly holds the read lock
	if sub.trySend(event, b.config.PublishTimeout) {
		atomic.AddInt64(&b.metrics.EventsDelivered, 1)
	} else {
		atomic.AddInt64(&b.metrics.EventsDropped, 1)
	}
}

// PublishAsync publishes an event asynchronously
func (b *EventBus) PublishAsync(event *Event) {
	go b.Publish(event)
}

// Subscribe subscribes to events of a specific type
func (b *EventBus) Subscribe(eventType EventType) <-chan *Event {
	return b.SubscribeWithFilter(eventType, nil)
}

// SubscribeWithFilter subscribes with a custom filter function
func (b *EventBus) SubscribeWithFilter(eventType EventType, filter func(*Event) bool) <-chan *Event {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		ch := make(chan *Event)
		close(ch)
		return ch
	}

	sub := &Subscriber{
		ID:      uuid.New().String(),
		Channel: make(chan *Event, b.config.BufferSize),
		Filter:  filter,
		Types:   []EventType{eventType},
	}

	b.subscribers[eventType] = append(b.subscribers[eventType], sub)
	atomic.AddInt64(&b.metrics.SubscribersActive, 1)
	atomic.AddInt64(&b.metrics.SubscribersTotal, 1)

	return sub.Channel
}

// SubscribeMultiple subscribes to multiple event types
func (b *EventBus) SubscribeMultiple(types ...EventType) <-chan *Event {
	return b.SubscribeMultipleWithFilter(nil, types...)
}

// SubscribeMultipleWithFilter subscribes to multiple event types with a filter
func (b *EventBus) SubscribeMultipleWithFilter(filter func(*Event) bool, types ...EventType) <-chan *Event {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		ch := make(chan *Event)
		close(ch)
		return ch
	}

	sub := &Subscriber{
		ID:      uuid.New().String(),
		Channel: make(chan *Event, b.config.BufferSize),
		Filter:  filter,
		Types:   types,
	}

	for _, eventType := range types {
		b.subscribers[eventType] = append(b.subscribers[eventType], sub)
	}

	atomic.AddInt64(&b.metrics.SubscribersActive, 1)
	atomic.AddInt64(&b.metrics.SubscribersTotal, 1)

	return sub.Channel
}

// SubscribeAll subscribes to all event types
func (b *EventBus) SubscribeAll() <-chan *Event {
	return b.SubscribeAllWithFilter(nil)
}

// SubscribeAllWithFilter subscribes to all events with a filter
func (b *EventBus) SubscribeAllWithFilter(filter func(*Event) bool) <-chan *Event {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		ch := make(chan *Event)
		close(ch)
		return ch
	}

	sub := &Subscriber{
		ID:      uuid.New().String(),
		Channel: make(chan *Event, b.config.BufferSize),
		Filter:  filter,
	}

	b.allSubs = append(b.allSubs, sub)
	atomic.AddInt64(&b.metrics.SubscribersActive, 1)
	atomic.AddInt64(&b.metrics.SubscribersTotal, 1)

	return sub.Channel
}

// Unsubscribe removes a subscriber by channel
func (b *EventBus) Unsubscribe(ch <-chan *Event) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Find and remove from type-specific subscribers
	for eventType, subs := range b.subscribers {
		for i, sub := range subs {
			if sub.Channel == ch {
				sub.Close()
				b.subscribers[eventType] = append(subs[:i], subs[i+1:]...)
				atomic.AddInt64(&b.metrics.SubscribersActive, -1)
				return
			}
		}
	}

	// Find and remove from all-event subscribers
	for i, sub := range b.allSubs {
		if sub.Channel == ch {
			sub.Close()
			b.allSubs = append(b.allSubs[:i], b.allSubs[i+1:]...)
			atomic.AddInt64(&b.metrics.SubscribersActive, -1)
			return
		}
	}
}

// cleanupLoop periodically cleans up closed subscribers
func (b *EventBus) cleanupLoop() {
	interval := b.config.CleanupInterval
	if interval <= 0 {
		interval = time.Minute // Default cleanup interval
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-b.ctx.Done():
			return
		case <-ticker.C:
			b.cleanup()
		}
	}
}

// cleanup removes closed subscribers
func (b *EventBus) cleanup() {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Clean type-specific subscribers
	for eventType, subs := range b.subscribers {
		active := make([]*Subscriber, 0, len(subs))
		for _, sub := range subs {
			sub.mu.Lock()
			if !sub.Closed {
				active = append(active, sub)
			}
			sub.mu.Unlock()
		}
		b.subscribers[eventType] = active
	}

	// Clean all-event subscribers
	active := make([]*Subscriber, 0, len(b.allSubs))
	for _, sub := range b.allSubs {
		sub.mu.Lock()
		if !sub.Closed {
			active = append(active, sub)
		}
		sub.mu.Unlock()
	}
	b.allSubs = active
}

// Metrics returns current bus metrics
func (b *EventBus) Metrics() *BusMetrics {
	return &BusMetrics{
		EventsPublished:   atomic.LoadInt64(&b.metrics.EventsPublished),
		EventsDelivered:   atomic.LoadInt64(&b.metrics.EventsDelivered),
		EventsDropped:     atomic.LoadInt64(&b.metrics.EventsDropped),
		SubscribersActive: atomic.LoadInt64(&b.metrics.SubscribersActive),
		SubscribersTotal:  atomic.LoadInt64(&b.metrics.SubscribersTotal),
	}
}

// SubscriberCount returns the number of subscribers for a given event type
func (b *EventBus) SubscriberCount(eventType EventType) int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subscribers[eventType])
}

// TotalSubscribers returns the total number of active subscribers
func (b *EventBus) TotalSubscribers() int {
	return int(atomic.LoadInt64(&b.metrics.SubscribersActive))
}

// Close shuts down the event bus
func (b *EventBus) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil
	}

	b.closed = true
	b.cancel()

	// Close all subscribers
	for _, subs := range b.subscribers {
		for _, sub := range subs {
			sub.Close()
		}
	}
	for _, sub := range b.allSubs {
		sub.Close()
	}

	return nil
}

// Wait blocks until an event of the specified type is received or context is cancelled
func (b *EventBus) Wait(ctx context.Context, eventType EventType) (*Event, error) {
	ch := b.Subscribe(eventType)
	defer b.Unsubscribe(ch)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case event, ok := <-ch:
		if !ok {
			return nil, fmt.Errorf("event bus closed")
		}
		return event, nil
	}
}

// WaitMultiple waits for any event from the specified types
func (b *EventBus) WaitMultiple(ctx context.Context, types ...EventType) (*Event, error) {
	ch := b.SubscribeMultiple(types...)
	defer b.Unsubscribe(ch)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case event, ok := <-ch:
		if !ok {
			return nil, fmt.Errorf("event bus closed")
		}
		return event, nil
	}
}

// GlobalBus is the default global event bus
var GlobalBus *EventBus

// InitGlobalBus initializes the global event bus
func InitGlobalBus(config *BusConfig) {
	if GlobalBus != nil {
		GlobalBus.Close()
	}
	GlobalBus = NewEventBus(config)
}

// Emit publishes an event to the global bus
func Emit(event *Event) {
	if GlobalBus != nil {
		GlobalBus.Publish(event)
	}
}

// EmitAsync publishes an event to the global bus asynchronously
func EmitAsync(event *Event) {
	if GlobalBus != nil {
		GlobalBus.PublishAsync(event)
	}
}

// On subscribes to an event type on the global bus
func On(eventType EventType) <-chan *Event {
	if GlobalBus == nil {
		ch := make(chan *Event)
		close(ch)
		return ch
	}
	return GlobalBus.Subscribe(eventType)
}
