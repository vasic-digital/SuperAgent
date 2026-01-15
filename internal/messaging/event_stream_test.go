package messaging

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventType_Values(t *testing.T) {
	// LLM events
	assert.Equal(t, EventType("llm.request.started"), EventTypeLLMRequestStarted)
	assert.Equal(t, EventType("llm.request.completed"), EventTypeLLMRequestCompleted)
	assert.Equal(t, EventType("llm.request.failed"), EventTypeLLMRequestFailed)
	assert.Equal(t, EventType("llm.stream.token"), EventTypeLLMStreamToken)
	assert.Equal(t, EventType("llm.stream.end"), EventTypeLLMStreamEnd)

	// Debate events
	assert.Equal(t, EventType("debate.started"), EventTypeDebateStarted)
	assert.Equal(t, EventType("debate.round"), EventTypeDebateRound)
	assert.Equal(t, EventType("debate.completed"), EventTypeDebateCompleted)
	assert.Equal(t, EventType("debate.failed"), EventTypeDebateFailed)

	// Verification events
	assert.Equal(t, EventType("verification.started"), EventTypeVerificationStarted)
	assert.Equal(t, EventType("verification.completed"), EventTypeVerificationCompleted)
	assert.Equal(t, EventType("verification.failed"), EventTypeVerificationFailed)

	// System events
	assert.Equal(t, EventType("system.startup"), EventTypeSystemStartup)
	assert.Equal(t, EventType("system.shutdown"), EventTypeSystemShutdown)
	assert.Equal(t, EventType("system.health"), EventTypeSystemHealth)
	assert.Equal(t, EventType("system.error"), EventTypeSystemError)

	// Audit events
	assert.Equal(t, EventType("audit.log"), EventTypeAuditLog)
}

func TestNewEvent(t *testing.T) {
	data := []byte(`{"key": "value"}`)
	event := NewEvent(EventTypeLLMRequestStarted, "llm-service", data)

	assert.NotEmpty(t, event.ID)
	assert.Contains(t, event.ID, "evt-")
	assert.Equal(t, EventTypeLLMRequestStarted, event.Type)
	assert.Equal(t, "llm-service", event.Source)
	assert.Equal(t, data, event.Data)
	assert.Equal(t, "application/json", event.DataContentType)
	assert.False(t, event.Timestamp.IsZero())
	assert.NotNil(t, event.Headers)
}

func TestNewEventWithID(t *testing.T) {
	data := []byte(`{"key": "value"}`)
	event := NewEventWithID("custom-event-id", EventTypeLLMRequestStarted, "llm-service", data)

	assert.Equal(t, "custom-event-id", event.ID)
	assert.Equal(t, EventTypeLLMRequestStarted, event.Type)
	assert.Equal(t, "llm-service", event.Source)
}

func TestNewEventFromJSON(t *testing.T) {
	data := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}

	event, err := NewEventFromJSON(EventTypeLLMRequestStarted, "llm-service", data)
	require.NoError(t, err)

	assert.Equal(t, EventTypeLLMRequestStarted, event.Type)
	assert.Equal(t, "llm-service", event.Source)
	assert.NotEmpty(t, event.Data)
}

func TestEvent_WithSubject(t *testing.T) {
	event := NewEvent(EventTypeLLMRequestStarted, "source", []byte("data"))
	event.WithSubject("request-123")

	assert.Equal(t, "request-123", event.Subject)
}

func TestEvent_WithDataSchema(t *testing.T) {
	event := NewEvent(EventTypeLLMRequestStarted, "source", []byte("data"))
	event.WithDataSchema("https://schema.example.com/event/v1")

	assert.Equal(t, "https://schema.example.com/event/v1", event.DataSchema)
}

func TestEvent_WithHeader(t *testing.T) {
	event := NewEvent(EventTypeLLMRequestStarted, "source", []byte("data"))
	event.WithHeader("key1", "value1").WithHeader("key2", "value2")

	assert.Equal(t, "value1", event.Headers["key1"])
	assert.Equal(t, "value2", event.Headers["key2"])
}

func TestEvent_WithTraceID(t *testing.T) {
	event := NewEvent(EventTypeLLMRequestStarted, "source", []byte("data"))
	event.WithTraceID("trace-123")

	assert.Equal(t, "trace-123", event.TraceID)
}

func TestEvent_WithCorrelationID(t *testing.T) {
	event := NewEvent(EventTypeLLMRequestStarted, "source", []byte("data"))
	event.WithCorrelationID("corr-123")

	assert.Equal(t, "corr-123", event.CorrelationID)
}

func TestEvent_WithKey(t *testing.T) {
	event := NewEvent(EventTypeLLMRequestStarted, "source", []byte("data"))
	event.WithKey([]byte("partition-key"))

	assert.Equal(t, []byte("partition-key"), event.Key)
}

func TestEvent_WithStringKey(t *testing.T) {
	event := NewEvent(EventTypeLLMRequestStarted, "source", []byte("data"))
	event.WithStringKey("partition-key")

	assert.Equal(t, []byte("partition-key"), event.Key)
}

func TestEvent_GetHeader(t *testing.T) {
	event := NewEvent(EventTypeLLMRequestStarted, "source", []byte("data"))
	event.Headers["key"] = "value"

	assert.Equal(t, "value", event.GetHeader("key"))
	assert.Equal(t, "", event.GetHeader("nonexistent"))

	// Test with nil headers
	event.Headers = nil
	assert.Equal(t, "", event.GetHeader("key"))
}

func TestEvent_UnmarshalData(t *testing.T) {
	data := map[string]interface{}{
		"key1": "value1",
		"key2": float64(123),
	}
	jsonData, _ := json.Marshal(data)
	event := NewEvent(EventTypeLLMRequestStarted, "source", jsonData)

	var decoded map[string]interface{}
	err := event.UnmarshalData(&decoded)
	require.NoError(t, err)

	assert.Equal(t, "value1", decoded["key1"])
	assert.Equal(t, float64(123), decoded["key2"])
}

func TestEvent_Clone(t *testing.T) {
	event := NewEvent(EventTypeLLMRequestStarted, "source", []byte("data"))
	event.Headers["key"] = "value"
	event.TraceID = "trace-123"
	event.Key = []byte("partition-key")
	event.Partition = 5
	event.Offset = 100

	clone := event.Clone()

	assert.Equal(t, event.ID, clone.ID)
	assert.Equal(t, event.Type, clone.Type)
	assert.Equal(t, event.Source, clone.Source)
	assert.Equal(t, event.Data, clone.Data)
	assert.Equal(t, event.Headers, clone.Headers)
	assert.Equal(t, event.TraceID, clone.TraceID)
	assert.Equal(t, event.Key, clone.Key)
	assert.Equal(t, event.Partition, clone.Partition)
	assert.Equal(t, event.Offset, clone.Offset)

	// Verify deep copy
	event.Data[0] = 'X'
	assert.NotEqual(t, event.Data, clone.Data)

	event.Headers["key"] = "modified"
	assert.NotEqual(t, event.Headers["key"], clone.Headers["key"])
}

func TestEvent_ToMessage(t *testing.T) {
	event := NewEvent(EventTypeLLMRequestStarted, "source", []byte(`{"data":"test"}`))
	event.TraceID = "trace-123"
	event.CorrelationID = "corr-123"
	event.Headers["custom"] = "header"
	event.Partition = 3
	event.Offset = 100

	msg := event.ToMessage()

	assert.Equal(t, event.ID, msg.ID)
	assert.Equal(t, string(event.Type), msg.Type)
	assert.Equal(t, event.TraceID, msg.TraceID)
	assert.Equal(t, event.CorrelationID, msg.CorrelationID)
	assert.Equal(t, event.Partition, msg.Partition)
	assert.Equal(t, event.Offset, msg.Offset)
}

func TestEventFromMessage(t *testing.T) {
	originalEvent := NewEvent(EventTypeLLMRequestStarted, "source", []byte(`{"data":"test"}`))
	originalEvent.TraceID = "trace-123"

	msg := originalEvent.ToMessage()
	msg.Partition = 5
	msg.Offset = 200

	recoveredEvent, err := EventFromMessage(msg)
	require.NoError(t, err)

	assert.Equal(t, originalEvent.ID, recoveredEvent.ID)
	assert.Equal(t, originalEvent.Type, recoveredEvent.Type)
	assert.Equal(t, int32(5), recoveredEvent.Partition)
	assert.Equal(t, int64(200), recoveredEvent.Offset)
}

func TestEventFromMessage_InvalidJSON(t *testing.T) {
	msg := NewMessage("test", []byte("invalid json"))

	_, err := EventFromMessage(msg)
	assert.Error(t, err)
}

func TestDefaultEventStreamConfig(t *testing.T) {
	cfg := DefaultEventStreamConfig()

	assert.Equal(t, []string{"localhost:9092"}, cfg.Brokers)
	assert.Equal(t, "helixagent", cfg.ClientID)
	assert.Equal(t, "helixagent-group", cfg.GroupID)
	assert.Equal(t, "helixagent.events", cfg.DefaultTopic)
	assert.False(t, cfg.AutoCreateTopics)
	assert.Equal(t, 3, cfg.DefaultPartitions)
	assert.Equal(t, 1, cfg.DefaultReplication)
	assert.Equal(t, int64(7*24*60*60*1000), cfg.RetentionMs)
	assert.Equal(t, CompressionLZ4, cfg.Compression)
	assert.Equal(t, 16384, cfg.BatchSize)
	assert.Equal(t, 5, cfg.LingerMs)
	assert.Equal(t, 1024*1024, cfg.MaxMessageSize)
	assert.Equal(t, 10*time.Second, cfg.SessionTimeout)
	assert.Equal(t, 3*time.Second, cfg.HeartbeatInterval)
	assert.Equal(t, 500, cfg.MaxPollRecords)
	assert.Equal(t, OffsetResetLatest, cfg.OffsetReset)
	assert.True(t, cfg.EnableIdempotence)
	assert.Equal(t, -1, cfg.Acks)
}

func TestTopicConstants(t *testing.T) {
	assert.Equal(t, "helixagent.events.llm.responses", TopicLLMResponses)
	assert.Equal(t, "helixagent.events.debate.rounds", TopicDebateRounds)
	assert.Equal(t, "helixagent.events.verification.results", TopicVerificationResults)
	assert.Equal(t, "helixagent.events.provider.health", TopicProviderHealth)
	assert.Equal(t, "helixagent.events.audit", TopicAuditLog)
	assert.Equal(t, "helixagent.events.metrics", TopicMetrics)
	assert.Equal(t, "helixagent.events.errors", TopicErrors)
	assert.Equal(t, "helixagent.stream.tokens", TopicTokenStream)
	assert.Equal(t, "helixagent.stream.sse", TopicSSEEvents)
	assert.Equal(t, "helixagent.stream.websocket", TopicWebSocketMessages)
}

func TestEventRegistry(t *testing.T) {
	registry := NewEventRegistry()

	handler := func(ctx context.Context, event *Event) error { return nil }

	// Register
	registry.Register(EventTypeLLMRequestStarted, handler)
	registry.Register(EventTypeLLMRequestCompleted, handler)
	registry.Register(EventTypeLLMRequestStarted, handler) // Second handler for same type

	// Get
	handlers := registry.Get(EventTypeLLMRequestStarted)
	assert.Len(t, handlers, 2)

	handlers = registry.Get(EventTypeDebateStarted)
	assert.Len(t, handlers, 0)

	// Types
	types := registry.Types()
	assert.Len(t, types, 2)
	assert.Contains(t, types, EventTypeLLMRequestStarted)
	assert.Contains(t, types, EventTypeLLMRequestCompleted)

	// Unregister
	registry.Unregister(EventTypeLLMRequestStarted)
	handlers = registry.Get(EventTypeLLMRequestStarted)
	assert.Len(t, handlers, 0)
}

func TestEventRegistry_Dispatch(t *testing.T) {
	registry := NewEventRegistry()

	called := 0
	handler := func(ctx context.Context, event *Event) error {
		called++
		return nil
	}

	registry.Register(EventTypeLLMRequestStarted, handler)
	registry.Register(EventTypeLLMRequestStarted, handler)

	event := NewEvent(EventTypeLLMRequestStarted, "source", []byte("data"))
	err := registry.Dispatch(context.Background(), event)
	require.NoError(t, err)
	assert.Equal(t, 2, called)
}

func TestEventRegistry_Dispatch_NoHandlers(t *testing.T) {
	registry := NewEventRegistry()

	event := NewEvent(EventTypeLLMRequestStarted, "source", []byte("data"))
	err := registry.Dispatch(context.Background(), event)
	require.NoError(t, err)
}

func TestEventBuffer(t *testing.T) {
	flushed := make([]*Event, 0)
	flushFn := func(events []*Event) error {
		flushed = append(flushed, events...)
		return nil
	}

	buffer := NewEventBuffer(3, 100*time.Millisecond, flushFn)

	// Add events
	buffer.Add(NewEvent(EventTypeLLMRequestStarted, "source", []byte("1")))
	buffer.Add(NewEvent(EventTypeLLMRequestStarted, "source", []byte("2")))

	// Manually flush
	err := buffer.Flush()
	require.NoError(t, err)
	assert.Len(t, flushed, 2)
}

func TestEventBuffer_Flush_Empty(t *testing.T) {
	flushFn := func(events []*Event) error {
		return nil
	}

	buffer := NewEventBuffer(3, 100*time.Millisecond, flushFn)
	err := buffer.Flush()
	require.NoError(t, err)
}

func TestEvent_JSONSerialization(t *testing.T) {
	event := NewEvent(EventTypeLLMRequestStarted, "llm-service", []byte(`{"request_id":"123"}`))
	event.Subject = "request-123"
	event.TraceID = "trace-123"
	event.Headers["custom"] = "header"

	// Marshal
	data, err := json.Marshal(event)
	require.NoError(t, err)

	// Unmarshal
	var decoded Event
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, event.ID, decoded.ID)
	assert.Equal(t, event.Type, decoded.Type)
	assert.Equal(t, event.Source, decoded.Source)
	assert.Equal(t, event.Subject, decoded.Subject)
	assert.Equal(t, event.TraceID, decoded.TraceID)
}

func TestConsumerGroupInfo(t *testing.T) {
	info := ConsumerGroupInfo{
		GroupID: "test-group",
		State:   "Stable",
		Members: []ConsumerMemberInfo{
			{
				MemberID: "member-1",
				ClientID: "client-1",
				Host:     "localhost",
				Partitions: map[string][]int32{
					"topic-1": {0, 1, 2},
				},
			},
		},
		PartitionAssignments: map[string][]int32{
			"topic-1": {0, 1, 2},
		},
	}

	assert.Equal(t, "test-group", info.GroupID)
	assert.Len(t, info.Members, 1)
	assert.Equal(t, "member-1", info.Members[0].MemberID)
}
