package streaming

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultKafkaStreamWriterConfig(t *testing.T) {
	config := DefaultKafkaStreamWriterConfig()

	assert.True(t, config.Enabled)
	assert.Equal(t, TopicStreamEvents, config.Topic)
	assert.True(t, config.Async)
	assert.Equal(t, 1000, config.BufferSize)
	assert.Equal(t, int64(24*60*60*1000), config.RetentionMs)
}

func TestNewKafkaStreamWriter(t *testing.T) {
	config := DefaultKafkaStreamWriterConfig()
	writer := NewKafkaStreamWriter(nil, "test-stream", nil, config)

	assert.NotNil(t, writer)
	assert.Equal(t, "test-stream", writer.StreamID())
	assert.NotNil(t, writer.eventCh)
	assert.NotNil(t, writer.stopCh)
}

func TestNewKafkaStreamWriter_NilConfig(t *testing.T) {
	writer := NewKafkaStreamWriter(nil, "test-stream", nil, nil)

	assert.NotNil(t, writer)
	assert.NotNil(t, writer.config)
	assert.True(t, writer.config.Enabled)
}

func TestKafkaStreamWriter_IsEnabled_NoHub(t *testing.T) {
	config := DefaultKafkaStreamWriterConfig()
	writer := NewKafkaStreamWriter(nil, "test-stream", nil, config)

	// Should be false when hub is nil
	assert.False(t, writer.IsEnabled())
}

func TestKafkaStreamWriter_IsEnabled_Disabled(t *testing.T) {
	config := DefaultKafkaStreamWriterConfig()
	config.Enabled = false
	writer := NewKafkaStreamWriter(nil, "test-stream", nil, config)

	assert.False(t, writer.IsEnabled())
}

func TestKafkaStreamWriter_WriteToken_NoHub(t *testing.T) {
	config := DefaultKafkaStreamWriterConfig()
	writer := NewKafkaStreamWriter(nil, "test-stream", nil, config)

	// Should not error when hub is nil (just skips publish)
	err := writer.WriteToken(context.Background(), "test token")
	assert.NoError(t, err)
}

func TestKafkaStreamWriter_WriteChunk_NoHub(t *testing.T) {
	config := DefaultKafkaStreamWriterConfig()
	writer := NewKafkaStreamWriter(nil, "test-stream", nil, config)

	chunk := &StreamChunk{
		Content:   "test content",
		Index:     1,
		Timestamp: time.Now(),
	}

	err := writer.WriteChunk(context.Background(), chunk)
	assert.NoError(t, err)
}

func TestKafkaStreamWriter_WriteDone_NoHub(t *testing.T) {
	config := DefaultKafkaStreamWriterConfig()
	writer := NewKafkaStreamWriter(nil, "test-stream", nil, config)

	err := writer.WriteDone(context.Background())
	assert.NoError(t, err)
}

func TestKafkaStreamWriter_WriteError_NoHub(t *testing.T) {
	config := DefaultKafkaStreamWriterConfig()
	writer := NewKafkaStreamWriter(nil, "test-stream", nil, config)

	err := writer.WriteError(context.Background(), errors.New("test error"))
	assert.NoError(t, err)
}

func TestKafkaStreamWriter_NextIndex(t *testing.T) {
	config := DefaultKafkaStreamWriterConfig()
	writer := NewKafkaStreamWriter(nil, "test-stream", nil, config)

	idx1 := writer.nextIndex()
	idx2 := writer.nextIndex()
	idx3 := writer.nextIndex()

	assert.Equal(t, 1, idx1)
	assert.Equal(t, 2, idx2)
	assert.Equal(t, 3, idx3)
}

func TestKafkaStreamWriter_StartStop(t *testing.T) {
	config := DefaultKafkaStreamWriterConfig()
	writer := NewKafkaStreamWriter(nil, "test-stream", nil, config)

	writer.Start()
	assert.True(t, writer.started)

	// Starting again should be a no-op
	writer.Start()
	assert.True(t, writer.started)

	writer.Stop()
}

func TestStreamEvent_Fields(t *testing.T) {
	event := &StreamEvent{
		ID:        "event-1",
		StreamID:  "stream-1",
		Type:      "token",
		Content:   "test content",
		Index:     5,
		Timestamp: time.Now().UTC(),
		Metadata:  map[string]interface{}{"key": "value"},
		Done:      false,
		Error:     "",
	}

	assert.Equal(t, "event-1", event.ID)
	assert.Equal(t, "stream-1", event.StreamID)
	assert.Equal(t, "token", event.Type)
	assert.Equal(t, "test content", event.Content)
	assert.Equal(t, 5, event.Index)
	assert.NotZero(t, event.Timestamp)
	assert.Equal(t, "value", event.Metadata["key"])
	assert.False(t, event.Done)
	assert.Empty(t, event.Error)
}

func TestStreamEvent_WithError(t *testing.T) {
	event := &StreamEvent{
		ID:       "event-1",
		StreamID: "stream-1",
		Type:     "error",
		Error:    "something went wrong",
		Done:     true,
	}

	assert.Equal(t, "error", event.Type)
	assert.Equal(t, "something went wrong", event.Error)
	assert.True(t, event.Done)
}

func TestGenerateStreamEventID(t *testing.T) {
	id1 := generateStreamEventID()
	time.Sleep(time.Millisecond)
	id2 := generateStreamEventID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
}

func TestTopicConstants(t *testing.T) {
	assert.Equal(t, "helixagent.stream.tokens", TopicTokenStream)
	assert.Equal(t, "helixagent.stream.sse", TopicSSEEvents)
	assert.Equal(t, "helixagent.stream.websocket", TopicWebSocketMessages)
	assert.Equal(t, "helixagent.stream.events", TopicStreamEvents)
}

func TestNewHybridStreamWriter(t *testing.T) {
	kafkaWriter := NewKafkaStreamWriter(nil, "test-stream", nil, nil)
	hybrid := NewHybridStreamWriter(nil, kafkaWriter, StreamingTypeSSE, nil)

	assert.NotNil(t, hybrid)
	assert.Equal(t, StreamingTypeSSE, hybrid.StreamType())
	assert.Equal(t, kafkaWriter, hybrid.KafkaWriter())
}

func TestHybridStreamWriter_StreamType(t *testing.T) {
	tests := []struct {
		streamType StreamingType
	}{
		{StreamingTypeSSE},
		{StreamingTypeWebSocket},
		{StreamingTypeJSONL},
		{StreamingTypeEventStream},
		{StreamingTypeStdout},
	}

	for _, tt := range tests {
		t.Run(string(tt.streamType), func(t *testing.T) {
			hybrid := NewHybridStreamWriter(nil, nil, tt.streamType, nil)
			assert.Equal(t, tt.streamType, hybrid.StreamType())
		})
	}
}

func TestKafkaStreamWriter_WriteEvent_WithDefaults(t *testing.T) {
	config := DefaultKafkaStreamWriterConfig()
	writer := NewKafkaStreamWriter(nil, "test-stream", nil, config)

	event := &StreamEvent{
		Type:    "custom",
		Content: "test",
	}

	err := writer.WriteEvent(context.Background(), event)
	assert.NoError(t, err)

	// Event should have been filled with defaults
	assert.NotEmpty(t, event.ID)
	assert.Equal(t, "test-stream", event.StreamID)
	assert.NotZero(t, event.Index)
	assert.NotZero(t, event.Timestamp)
}

func TestKafkaStreamWriter_SyncMode(t *testing.T) {
	config := DefaultKafkaStreamWriterConfig()
	config.Async = false
	writer := NewKafkaStreamWriter(nil, "test-stream", nil, config)

	assert.Nil(t, writer.eventCh)

	// Start should be a no-op in sync mode
	writer.Start()
	assert.False(t, writer.started)

	// Write should still work (just skips publish since no hub)
	err := writer.WriteToken(context.Background(), "test")
	assert.NoError(t, err)
}

func TestKafkaStreamWriterConfig_Fields(t *testing.T) {
	config := &KafkaStreamWriterConfig{
		Enabled:     true,
		Topic:       "custom-topic",
		Async:       true,
		BufferSize:  500,
		RetentionMs: 3600000,
	}

	assert.True(t, config.Enabled)
	assert.Equal(t, "custom-topic", config.Topic)
	assert.True(t, config.Async)
	assert.Equal(t, 500, config.BufferSize)
	assert.Equal(t, int64(3600000), config.RetentionMs)
}

func TestHybridStreamWriter_WriteChunk_NilDirect(t *testing.T) {
	hybrid := NewHybridStreamWriter(nil, nil, StreamingTypeSSE, nil)

	chunk := &StreamChunk{
		Content: "test",
		Index:   1,
	}

	// Should not panic with nil direct writer
	err := hybrid.WriteChunk(context.Background(), chunk)
	assert.NoError(t, err)
}

func TestHybridStreamWriter_WriteDone_NilDirect(t *testing.T) {
	hybrid := NewHybridStreamWriter(nil, nil, StreamingTypeSSE, nil)

	// Should not panic with nil direct writer
	err := hybrid.WriteDone(context.Background())
	assert.NoError(t, err)
}

func TestRandomStreamString(t *testing.T) {
	str1 := randomStreamString(8)
	time.Sleep(time.Millisecond)
	str2 := randomStreamString(8)

	require.Len(t, str1, 8)
	require.Len(t, str2, 8)
	assert.NotEqual(t, str1, str2)
}
