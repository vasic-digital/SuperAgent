// Package adapters provides compatibility adapters between HelixAgent's internal
// interfaces and the extracted modules (eventbus, concurrency).
package adapters

import (
	"context"
	"time"

	"digital.vasic.eventbus/pkg/bus"
	"digital.vasic.eventbus/pkg/event"
	"digital.vasic.eventbus/pkg/filter"
)

// EventType provides compatibility with the old internal/events.EventType.
// It is an alias to event.Type from the eventbus module.
type EventType = event.Type

// Event provides compatibility with the old internal/events.Event.
// It is an alias to event.Event from the eventbus module.
type Event = event.Event

// Standard event types - mirroring the old internal/events constants
const (
	// Provider events
	EventProviderRegistered    EventType = "provider.registered"
	EventProviderUnregistered  EventType = "provider.unregistered"
	EventProviderHealthChanged EventType = "provider.health.changed"
	EventProviderScoreUpdated  EventType = "provider.score.updated"

	// MCP events
	EventMCPServerConnected     EventType = "mcp.server.connected"
	EventMCPServerDisconnected  EventType = "mcp.server.disconnected"
	EventMCPServerHealthChanged EventType = "mcp.server.health.changed"
	EventMCPToolExecuted        EventType = "mcp.tool.executed"
	EventMCPToolFailed          EventType = "mcp.tool.failed"

	// Debate events
	EventDebateStarted        EventType = "debate.started"
	EventDebateRoundStarted   EventType = "debate.round.started"
	EventDebateRoundCompleted EventType = "debate.round.completed"
	EventDebateCompleted      EventType = "debate.completed"
	EventDebateFailed         EventType = "debate.failed"

	// Cache events
	EventCacheHit         EventType = "cache.hit"
	EventCacheMiss        EventType = "cache.miss"
	EventCacheInvalidated EventType = "cache.invalidated"
	EventCacheExpired     EventType = "cache.expired"

	// System events
	EventSystemStartup     EventType = "system.startup"
	EventSystemShutdown    EventType = "system.shutdown"
	EventSystemHealthCheck EventType = "system.health.check"
	EventSystemError       EventType = "system.error"

	// Request events
	EventRequestReceived  EventType = "request.received"
	EventRequestCompleted EventType = "request.completed"
	EventRequestFailed    EventType = "request.failed"
)

// NewEvent creates a new event with the given type, source, and payload.
// This provides compatibility with the old internal/events.NewEvent function.
func NewEvent(eventType EventType, source string, payload interface{}) *Event {
	return event.New(eventType, source, payload)
}

// BusConfig is an alias to bus.Config for compatibility.
type BusConfig = bus.Config

// BusMetrics is an alias to bus.Metrics for compatibility.
type BusMetrics = bus.Metrics

// DefaultBusConfig returns default bus configuration.
func DefaultBusConfig() *BusConfig {
	return bus.DefaultConfig()
}

// EventBus wraps the eventbus module's EventBus to provide
// compatibility with the old internal/events.EventBus API.
type EventBus struct {
	inner *bus.EventBus
}

// NewEventBus creates a new event bus with the given configuration.
func NewEventBus(config *BusConfig) *EventBus {
	return &EventBus{
		inner: bus.New(config),
	}
}

// Publish sends an event to all matching subscribers.
func (b *EventBus) Publish(e *Event) {
	b.inner.Publish(e)
}

// PublishAsync publishes an event asynchronously.
func (b *EventBus) PublishAsync(e *Event) {
	b.inner.PublishAsync(e)
}

// Subscribe subscribes to events of a specific type.
// Returns a channel that receives events.
func (b *EventBus) Subscribe(eventType EventType) <-chan *Event {
	sub := b.inner.Subscribe(eventType)
	return sub.Channel
}

// SubscribeWithFilter subscribes to a specific event type with a filter.
func (b *EventBus) SubscribeWithFilter(eventType EventType, f func(*Event) bool) <-chan *Event {
	var filterFunc filter.Filter
	if f != nil {
		filterFunc = filter.Filter(f)
	}
	sub := b.inner.SubscribeWithFilter(eventType, filterFunc)
	return sub.Channel
}

// SubscribeMultiple subscribes to multiple event types.
func (b *EventBus) SubscribeMultiple(types ...EventType) <-chan *Event {
	sub := b.inner.SubscribeMultiple(types...)
	return sub.Channel
}

// SubscribeMultipleWithFilter subscribes to multiple event types with a filter.
func (b *EventBus) SubscribeMultipleWithFilter(f func(*Event) bool, types ...EventType) <-chan *Event {
	var filterFunc filter.Filter
	if f != nil {
		filterFunc = filter.Filter(f)
	}
	sub := b.inner.SubscribeMultipleWithFilter(filterFunc, types...)
	return sub.Channel
}

// SubscribeAll subscribes to all event types.
func (b *EventBus) SubscribeAll() <-chan *Event {
	sub := b.inner.SubscribeAll()
	return sub.Channel
}

// SubscribeAllWithFilter subscribes to all events with a filter.
func (b *EventBus) SubscribeAllWithFilter(f func(*Event) bool) <-chan *Event {
	var filterFunc filter.Filter
	if f != nil {
		filterFunc = filter.Filter(f)
	}
	sub := b.inner.SubscribeAllWithFilter(filterFunc)
	return sub.Channel
}

// Unsubscribe removes a subscriber by its channel.
func (b *EventBus) Unsubscribe(ch <-chan *Event) {
	b.inner.UnsubscribeByChannel(ch)
}

// Wait blocks until an event of the specified type is received or context is cancelled.
func (b *EventBus) Wait(ctx context.Context, eventType EventType) (*Event, error) {
	return b.inner.Wait(ctx, eventType)
}

// WaitMultiple waits for any event from the specified types.
func (b *EventBus) WaitMultiple(ctx context.Context, types ...EventType) (*Event, error) {
	return b.inner.WaitMultiple(ctx, types...)
}

// Metrics returns current bus metrics.
func (b *EventBus) Metrics() *BusMetrics {
	return b.inner.Metrics()
}

// SubscriberCount returns the number of subscribers for an event type.
func (b *EventBus) SubscriberCount(eventType EventType) int {
	return b.inner.SubscriberCount(eventType)
}

// TotalSubscribers returns the total number of active subscribers.
func (b *EventBus) TotalSubscribers() int {
	return b.inner.TotalSubscribers()
}

// Close shuts down the event bus.
func (b *EventBus) Close() error {
	return b.inner.Close()
}

// Inner returns the underlying eventbus.EventBus for advanced usage.
func (b *EventBus) Inner() *bus.EventBus {
	return b.inner
}

// GlobalBus is the default global event bus.
var GlobalBus *EventBus

// InitGlobalBus initializes the global event bus.
func InitGlobalBus(config *BusConfig) {
	if GlobalBus != nil {
		_ = GlobalBus.Close()
	}
	GlobalBus = NewEventBus(config)
}

// Emit publishes an event to the global bus.
func Emit(e *Event) {
	if GlobalBus != nil {
		GlobalBus.Publish(e)
	}
}

// EmitAsync publishes an event to the global bus asynchronously.
func EmitAsync(e *Event) {
	if GlobalBus != nil {
		GlobalBus.PublishAsync(e)
	}
}

// On subscribes to an event type on the global bus.
func On(eventType EventType) <-chan *Event {
	if GlobalBus == nil {
		ch := make(chan *Event)
		close(ch)
		return ch
	}
	return GlobalBus.Subscribe(eventType)
}

// Subscriber provides compatibility with the old internal/events.Subscriber.
type Subscriber struct {
	ID      string
	Channel chan *Event
	Filter  func(*Event) bool
	Types   []EventType
	Closed  bool
}

// InvalidationRule provides compatibility with cache invalidation rules.
type InvalidationRule struct {
	EventType  EventType
	KeyPattern string
	Tags       []string
	Handler    func(event *Event) []string
}

// LazyProviderConfig compatibility adapter for EventBus field.
type LazyProviderConfig struct {
	InitTimeout     time.Duration
	RetryAttempts   int
	RetryDelay      time.Duration
	PrewarmOnAccess bool
	EventBus        *EventBus
}

// DefaultLazyProviderConfig returns default configuration.
func DefaultLazyProviderConfig() *LazyProviderConfig {
	return &LazyProviderConfig{
		InitTimeout:     30 * time.Second,
		RetryAttempts:   3,
		RetryDelay:      1 * time.Second,
		PrewarmOnAccess: false,
	}
}
