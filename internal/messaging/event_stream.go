package messaging

import (
	"context"
	"encoding/json"
	"time"
)

// EventType represents the type of event.
type EventType string

// Event types for HelixAgent.
const (
	// LLM events
	EventTypeLLMRequestStarted   EventType = "llm.request.started"
	EventTypeLLMRequestCompleted EventType = "llm.request.completed"
	EventTypeLLMRequestFailed    EventType = "llm.request.failed"
	EventTypeLLMStreamToken EventType = "llm.stream.token" // #nosec G101 - event type name, not credentials
	EventTypeLLMStreamEnd        EventType = "llm.stream.end"

	// Debate events
	EventTypeDebateStarted   EventType = "debate.started"
	EventTypeDebateRound     EventType = "debate.round"
	EventTypeDebateCompleted EventType = "debate.completed"
	EventTypeDebateFailed    EventType = "debate.failed"

	// Verification events
	EventTypeVerificationStarted   EventType = "verification.started"
	EventTypeVerificationCompleted EventType = "verification.completed"
	EventTypeVerificationFailed    EventType = "verification.failed"
	EventTypeProviderDiscovered    EventType = "provider.discovered"
	EventTypeProviderScored        EventType = "provider.scored"
	EventTypeProviderHealthCheck   EventType = "provider.health_check"
	EventTypeModelRanked           EventType = "model.ranked"
	EventTypeTeamSelected          EventType = "team.selected"

	// Task events
	EventTypeTaskCreated   EventType = "task.created"
	EventTypeTaskStarted   EventType = "task.started"
	EventTypeTaskCompleted EventType = "task.completed"
	EventTypeTaskFailed    EventType = "task.failed"
	EventTypeTaskCanceled  EventType = "task.canceled"

	// System events
	EventTypeSystemStartup  EventType = "system.startup"
	EventTypeSystemShutdown EventType = "system.shutdown"
	EventTypeSystemHealth   EventType = "system.health"
	EventTypeSystemError    EventType = "system.error"

	// Audit events
	EventTypeAuditLog EventType = "audit.log"
)

// Event represents an event in the event stream.
type Event struct {
	// ID is the unique event identifier.
	ID string `json:"id"`
	// Type is the event type.
	Type EventType `json:"type"`
	// Source is the event source (service/component).
	Source string `json:"source"`
	// Subject is the event subject (entity being acted upon).
	Subject string `json:"subject,omitempty"`
	// Data is the event payload.
	Data []byte `json:"data"`
	// DataContentType is the MIME type of the data.
	DataContentType string `json:"data_content_type,omitempty"`
	// DataSchema is the schema URL for the data.
	DataSchema string `json:"data_schema,omitempty"`
	// Timestamp is when the event occurred.
	Timestamp time.Time `json:"timestamp"`
	// Headers contains event metadata.
	Headers map[string]string `json:"headers,omitempty"`
	// TraceID is for distributed tracing.
	TraceID string `json:"trace_id,omitempty"`
	// CorrelationID links related events.
	CorrelationID string `json:"correlation_id,omitempty"`
	// Partition is the Kafka partition.
	Partition int32 `json:"partition,omitempty"`
	// Offset is the Kafka offset.
	Offset int64 `json:"offset,omitempty"`
	// Key is the partition key.
	Key []byte `json:"key,omitempty"`
}

// NewEvent creates a new event.
func NewEvent(eventType EventType, source string, data []byte) *Event {
	return &Event{
		ID:              generateEventID(),
		Type:            eventType,
		Source:          source,
		Data:            data,
		DataContentType: "application/json",
		Timestamp:       time.Now().UTC(),
		Headers:         make(map[string]string),
	}
}

// NewEventWithID creates a new event with a specific ID.
func NewEventWithID(id string, eventType EventType, source string, data []byte) *Event {
	event := NewEvent(eventType, source, data)
	event.ID = id
	return event
}

// NewEventFromJSON creates an event from JSON data.
func NewEventFromJSON(eventType EventType, source string, data interface{}) (*Event, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return NewEvent(eventType, source, payload), nil
}

// WithSubject sets the event subject.
func (e *Event) WithSubject(subject string) *Event {
	e.Subject = subject
	return e
}

// WithDataSchema sets the data schema URL.
func (e *Event) WithDataSchema(schema string) *Event {
	e.DataSchema = schema
	return e
}

// WithHeader sets a header value.
func (e *Event) WithHeader(key, value string) *Event {
	if e.Headers == nil {
		e.Headers = make(map[string]string)
	}
	e.Headers[key] = value
	return e
}

// WithTraceID sets the trace ID.
func (e *Event) WithTraceID(traceID string) *Event {
	e.TraceID = traceID
	return e
}

// WithCorrelationID sets the correlation ID.
func (e *Event) WithCorrelationID(correlationID string) *Event {
	e.CorrelationID = correlationID
	return e
}

// WithKey sets the partition key.
func (e *Event) WithKey(key []byte) *Event {
	e.Key = key
	return e
}

// WithStringKey sets the partition key from a string.
func (e *Event) WithStringKey(key string) *Event {
	e.Key = []byte(key)
	return e
}

// GetHeader gets a header value.
func (e *Event) GetHeader(key string) string {
	if e.Headers == nil {
		return ""
	}
	return e.Headers[key]
}

// UnmarshalData unmarshals the event data into the given value.
func (e *Event) UnmarshalData(v interface{}) error {
	return json.Unmarshal(e.Data, v)
}

// Clone creates a deep copy of the event.
func (e *Event) Clone() *Event {
	clone := &Event{
		ID:              e.ID,
		Type:            e.Type,
		Source:          e.Source,
		Subject:         e.Subject,
		Data:            make([]byte, len(e.Data)),
		DataContentType: e.DataContentType,
		DataSchema:      e.DataSchema,
		Timestamp:       e.Timestamp,
		TraceID:         e.TraceID,
		CorrelationID:   e.CorrelationID,
		Partition:       e.Partition,
		Offset:          e.Offset,
		Key:             make([]byte, len(e.Key)),
	}
	copy(clone.Data, e.Data)
	copy(clone.Key, e.Key)
	if e.Headers != nil {
		clone.Headers = make(map[string]string)
		for k, v := range e.Headers {
			clone.Headers[k] = v
		}
	}
	return clone
}

// ToMessage converts the event to a Message.
func (e *Event) ToMessage() *Message {
	payload, _ := json.Marshal(e)
	msg := NewMessage(string(e.Type), payload)
	msg.ID = e.ID
	msg.TraceID = e.TraceID
	msg.CorrelationID = e.CorrelationID
	msg.Partition = e.Partition
	msg.Offset = e.Offset
	for k, v := range e.Headers {
		msg.SetHeader(k, v)
	}
	return msg
}

// EventFromMessage creates an Event from a Message.
func EventFromMessage(msg *Message) (*Event, error) {
	var event Event
	if err := json.Unmarshal(msg.Payload, &event); err != nil {
		return nil, err
	}
	event.Partition = msg.Partition
	event.Offset = msg.Offset
	return &event, nil
}

// generateEventID generates a unique event ID.
func generateEventID() string {
	return "evt-" + time.Now().UTC().Format("20060102150405.000000") + "-" + randomString(8)
}

// EventHandler is a function that processes events.
type EventHandler func(ctx context.Context, event *Event) error

// EventFilter is a function that filters events.
type EventFilter func(event *Event) bool

// EventStreamBroker defines the interface for event stream brokers (e.g., Kafka).
type EventStreamBroker interface {
	MessageBroker

	// CreateTopic creates a new topic.
	CreateTopic(ctx context.Context, name string, partitions int, replication int) error

	// DeleteTopic deletes a topic.
	DeleteTopic(ctx context.Context, name string) error

	// ListTopics lists all topics.
	ListTopics(ctx context.Context) ([]string, error)

	// GetTopicMetadata returns metadata for a topic.
	GetTopicMetadata(ctx context.Context, topic string) (*TopicMetadata, error)

	// CreateConsumerGroup creates a consumer group.
	CreateConsumerGroup(ctx context.Context, groupID string) error

	// DeleteConsumerGroup deletes a consumer group.
	DeleteConsumerGroup(ctx context.Context, groupID string) error

	// PublishEvent publishes an event to a topic.
	PublishEvent(ctx context.Context, topic string, event *Event) error

	// PublishEventBatch publishes multiple events to a topic.
	PublishEventBatch(ctx context.Context, topic string, events []*Event) error

	// SubscribeEvents subscribes to events from a topic.
	SubscribeEvents(ctx context.Context, topic string, handler EventHandler, opts ...SubscribeOption) (Subscription, error)

	// StreamMessages returns a channel of messages from a topic.
	StreamMessages(ctx context.Context, topic string, opts ...StreamOption) (<-chan *Message, error)

	// StreamEvents returns a channel of events from a topic.
	StreamEvents(ctx context.Context, topic string, opts ...StreamOption) (<-chan *Event, error)

	// CommitOffset commits the offset for a topic partition.
	CommitOffset(ctx context.Context, topic string, partition int32, offset int64) error

	// GetOffset returns the current offset for a topic partition.
	GetOffset(ctx context.Context, topic string, partition int32) (int64, error)

	// SeekToOffset seeks to a specific offset.
	SeekToOffset(ctx context.Context, topic string, partition int32, offset int64) error

	// SeekToTimestamp seeks to a specific timestamp.
	SeekToTimestamp(ctx context.Context, topic string, partition int32, ts time.Time) error

	// SeekToBeginning seeks to the beginning of a topic partition.
	SeekToBeginning(ctx context.Context, topic string, partition int32) error

	// SeekToEnd seeks to the end of a topic partition.
	SeekToEnd(ctx context.Context, topic string, partition int32) error
}

// EventStreamConfig holds configuration for event stream broker.
type EventStreamConfig struct {
	BrokerConfig

	// Brokers is the list of Kafka broker addresses.
	Brokers []string `json:"brokers" yaml:"brokers"`

	// ClientID is the client identifier.
	ClientID string `json:"client_id" yaml:"client_id"`

	// GroupID is the consumer group ID.
	GroupID string `json:"group_id" yaml:"group_id"`

	// DefaultTopic is the default topic for publishing.
	DefaultTopic string `json:"default_topic" yaml:"default_topic"`

	// AutoCreateTopics enables automatic topic creation.
	AutoCreateTopics bool `json:"auto_create_topics" yaml:"auto_create_topics"`

	// DefaultPartitions is the default number of partitions.
	DefaultPartitions int `json:"default_partitions" yaml:"default_partitions"`

	// DefaultReplication is the default replication factor.
	DefaultReplication int `json:"default_replication" yaml:"default_replication"`

	// RetentionMs is the default retention period in milliseconds.
	RetentionMs int64 `json:"retention_ms" yaml:"retention_ms"`

	// Compression is the compression algorithm.
	Compression CompressionType `json:"compression" yaml:"compression"`

	// BatchSize is the producer batch size.
	BatchSize int `json:"batch_size" yaml:"batch_size"`

	// LingerMs is the producer linger time in milliseconds.
	LingerMs int `json:"linger_ms" yaml:"linger_ms"`

	// MaxMessageSize is the maximum message size in bytes.
	MaxMessageSize int `json:"max_message_size" yaml:"max_message_size"`

	// SessionTimeout is the consumer session timeout.
	SessionTimeout time.Duration `json:"session_timeout" yaml:"session_timeout"`

	// HeartbeatInterval is the consumer heartbeat interval.
	HeartbeatInterval time.Duration `json:"heartbeat_interval" yaml:"heartbeat_interval"`

	// MaxPollRecords is the maximum records per poll.
	MaxPollRecords int `json:"max_poll_records" yaml:"max_poll_records"`

	// OffsetReset is the offset reset policy.
	OffsetReset OffsetReset `json:"offset_reset" yaml:"offset_reset"`

	// EnableIdempotence enables producer idempotence.
	EnableIdempotence bool `json:"enable_idempotence" yaml:"enable_idempotence"`

	// Acks is the acknowledgment level (0, 1, -1 for all).
	Acks int `json:"acks" yaml:"acks"`
}

// DefaultEventStreamConfig returns the default event stream configuration.
func DefaultEventStreamConfig() *EventStreamConfig {
	return &EventStreamConfig{
		BrokerConfig:       *DefaultBrokerConfig(),
		Brokers:            []string{"localhost:9092"},
		ClientID:           "helixagent",
		GroupID:            "helixagent-group",
		DefaultTopic:       "helixagent.events",
		AutoCreateTopics:   false,
		DefaultPartitions:  3,
		DefaultReplication: 1,
		RetentionMs:        7 * 24 * 60 * 60 * 1000, // 7 days
		Compression:        CompressionLZ4,
		BatchSize:          16384,
		LingerMs:           5,
		MaxMessageSize:     1024 * 1024, // 1MB
		SessionTimeout:     10 * time.Second,
		HeartbeatInterval:  3 * time.Second,
		MaxPollRecords:     500,
		OffsetReset:        OffsetResetLatest,
		EnableIdempotence:  true,
		Acks:               -1, // Wait for all replicas
	}
}

// Topic names for HelixAgent event streams.
const (
	// TopicLLMResponses is for LLM response events.
	TopicLLMResponses = "helixagent.events.llm.responses"
	// TopicDebateRounds is for debate round events.
	TopicDebateRounds = "helixagent.events.debate.rounds"
	// TopicVerificationResults is for verification result events.
	TopicVerificationResults = "helixagent.events.verification.results"
	// TopicProviderHealth is for provider health events.
	TopicProviderHealth = "helixagent.events.provider.health"
	// TopicAuditLog is for audit log events.
	TopicAuditLog = "helixagent.events.audit"
	// TopicMetrics is for metrics events.
	TopicMetrics = "helixagent.events.metrics"
	// TopicErrors is for error events.
	TopicErrors = "helixagent.events.errors"
	// TopicTokenStream is for token streaming events - topic name, not credentials
	TopicTokenStream = "helixagent.stream.tokens" // #nosec G101 - topic name, not credentials
	// TopicSSEEvents is for SSE events.
	TopicSSEEvents = "helixagent.stream.sse"
	// TopicWebSocketMessages is for WebSocket messages.
	TopicWebSocketMessages = "helixagent.stream.websocket"
)

// ConsumerGroupInfo holds information about a consumer group.
type ConsumerGroupInfo struct {
	// GroupID is the consumer group ID.
	GroupID string `json:"group_id"`
	// State is the group state.
	State string `json:"state"`
	// Members are the group members.
	Members []ConsumerMemberInfo `json:"members"`
	// PartitionAssignments are the partition assignments.
	PartitionAssignments map[string][]int32 `json:"partition_assignments"`
}

// ConsumerMemberInfo holds information about a consumer group member.
type ConsumerMemberInfo struct {
	// MemberID is the member ID.
	MemberID string `json:"member_id"`
	// ClientID is the client ID.
	ClientID string `json:"client_id"`
	// Host is the member's host.
	Host string `json:"host"`
	// Partitions are the assigned partitions.
	Partitions map[string][]int32 `json:"partitions"`
}

// EventRegistry holds event type to handler mappings.
type EventRegistry struct {
	handlers map[EventType][]EventHandler
}

// NewEventRegistry creates a new event registry.
func NewEventRegistry() *EventRegistry {
	return &EventRegistry{
		handlers: make(map[EventType][]EventHandler),
	}
}

// Register registers a handler for an event type.
func (r *EventRegistry) Register(eventType EventType, handler EventHandler) {
	r.handlers[eventType] = append(r.handlers[eventType], handler)
}

// Get returns the handlers for an event type.
func (r *EventRegistry) Get(eventType EventType) []EventHandler {
	return r.handlers[eventType]
}

// Unregister removes all handlers for an event type.
func (r *EventRegistry) Unregister(eventType EventType) {
	delete(r.handlers, eventType)
}

// Types returns all registered event types.
func (r *EventRegistry) Types() []EventType {
	types := make([]EventType, 0, len(r.handlers))
	for t := range r.handlers {
		types = append(types, t)
	}
	return types
}

// Dispatch dispatches an event to all registered handlers.
func (r *EventRegistry) Dispatch(ctx context.Context, event *Event) error {
	handlers := r.handlers[event.Type]
	if len(handlers) == 0 {
		return nil
	}

	var errs MultiError
	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			errs.Add(err)
		}
	}

	return errs.ErrorOrNil()
}

// EventBuffer buffers events for batch processing.
type EventBuffer struct {
	events   []*Event
	maxSize  int
	flushFn  func(events []*Event) error
	flushCh  chan struct{}
	interval time.Duration
	stopCh   chan struct{}
}

// NewEventBuffer creates a new event buffer.
func NewEventBuffer(maxSize int, interval time.Duration, flushFn func(events []*Event) error) *EventBuffer {
	return &EventBuffer{
		events:   make([]*Event, 0, maxSize),
		maxSize:  maxSize,
		flushFn:  flushFn,
		flushCh:  make(chan struct{}, 1),
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Add adds an event to the buffer.
func (b *EventBuffer) Add(event *Event) {
	b.events = append(b.events, event)
	if len(b.events) >= b.maxSize {
		select {
		case b.flushCh <- struct{}{}:
		default:
		}
	}
}

// Flush flushes the buffer.
func (b *EventBuffer) Flush() error {
	if len(b.events) == 0 {
		return nil
	}
	events := b.events
	b.events = make([]*Event, 0, b.maxSize)
	return b.flushFn(events)
}

// Start starts the background flusher.
func (b *EventBuffer) Start() {
	go func() {
		ticker := time.NewTicker(b.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				_ = b.Flush()
			case <-b.flushCh:
				_ = b.Flush()
			case <-b.stopCh:
				_ = b.Flush()
				return
			}
		}
	}()
}

// Stop stops the background flusher.
func (b *EventBuffer) Stop() {
	close(b.stopCh)
}
