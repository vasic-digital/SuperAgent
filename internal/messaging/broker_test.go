package messaging

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBrokerType_String(t *testing.T) {
	tests := []struct {
		bt       BrokerType
		expected string
	}{
		{BrokerTypeRabbitMQ, "rabbitmq"},
		{BrokerTypeKafka, "kafka"},
		{BrokerTypeInMemory, "inmemory"},
	}

	for _, tt := range tests {
		t.Run(string(tt.bt), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.bt.String())
		})
	}
}

func TestBrokerType_IsValid(t *testing.T) {
	tests := []struct {
		bt       BrokerType
		expected bool
	}{
		{BrokerTypeRabbitMQ, true},
		{BrokerTypeKafka, true},
		{BrokerTypeInMemory, true},
		{BrokerType("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.bt), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.bt.IsValid())
		})
	}
}

func TestNewMessage(t *testing.T) {
	payload := []byte(`{"key": "value"}`)
	msg := NewMessage("test.type", payload)

	assert.NotEmpty(t, msg.ID)
	assert.Equal(t, "test.type", msg.Type)
	assert.Equal(t, payload, msg.Payload)
	assert.NotNil(t, msg.Headers)
	assert.False(t, msg.Timestamp.IsZero())
	assert.Equal(t, PriorityNormal, msg.Priority)
	assert.Equal(t, 0, msg.RetryCount)
	assert.Equal(t, 3, msg.MaxRetries)
	assert.Equal(t, DeliveryModePersistent, msg.DeliveryMode)
	assert.Equal(t, MessageStatePending, msg.State)
}

func TestNewMessageWithID(t *testing.T) {
	payload := []byte(`{"key": "value"}`)
	msg := NewMessageWithID("custom-id", "test.type", payload)

	assert.Equal(t, "custom-id", msg.ID)
	assert.Equal(t, "test.type", msg.Type)
	assert.Equal(t, payload, msg.Payload)
}

func TestMessage_SetHeader(t *testing.T) {
	msg := NewMessage("test", []byte("payload"))
	msg.SetHeader("key1", "value1").SetHeader("key2", "value2")

	assert.Equal(t, "value1", msg.Headers["key1"])
	assert.Equal(t, "value2", msg.Headers["key2"])
}

func TestMessage_GetHeader(t *testing.T) {
	msg := NewMessage("test", []byte("payload"))
	msg.Headers["key"] = "value"

	assert.Equal(t, "value", msg.GetHeader("key"))
	assert.Equal(t, "", msg.GetHeader("nonexistent"))

	// Test with nil headers
	msg.Headers = nil
	assert.Equal(t, "", msg.GetHeader("key"))
}

func TestMessage_WithPriority(t *testing.T) {
	msg := NewMessage("test", []byte("payload"))
	msg.WithPriority(PriorityHigh)

	assert.Equal(t, PriorityHigh, msg.Priority)
}

func TestMessage_WithTraceID(t *testing.T) {
	msg := NewMessage("test", []byte("payload"))
	msg.WithTraceID("trace-123")

	assert.Equal(t, "trace-123", msg.TraceID)
}

func TestMessage_WithCorrelationID(t *testing.T) {
	msg := NewMessage("test", []byte("payload"))
	msg.WithCorrelationID("corr-123")

	assert.Equal(t, "corr-123", msg.CorrelationID)
}

func TestMessage_WithReplyTo(t *testing.T) {
	msg := NewMessage("test", []byte("payload"))
	msg.WithReplyTo("reply.queue")

	assert.Equal(t, "reply.queue", msg.ReplyTo)
}

func TestMessage_WithExpiration(t *testing.T) {
	msg := NewMessage("test", []byte("payload"))
	expiration := time.Now().Add(1 * time.Hour)
	msg.WithExpiration(expiration)

	assert.Equal(t, expiration, msg.Expiration)
}

func TestMessage_WithTTL(t *testing.T) {
	msg := NewMessage("test", []byte("payload"))
	before := time.Now().UTC()
	msg.WithTTL(1 * time.Hour)
	after := time.Now().UTC()

	assert.True(t, msg.Expiration.After(before.Add(59*time.Minute)))
	assert.True(t, msg.Expiration.Before(after.Add(61*time.Minute)))
}

func TestMessage_WithMaxRetries(t *testing.T) {
	msg := NewMessage("test", []byte("payload"))
	msg.WithMaxRetries(5)

	assert.Equal(t, 5, msg.MaxRetries)
}

func TestMessage_WithDeliveryMode(t *testing.T) {
	msg := NewMessage("test", []byte("payload"))
	msg.WithDeliveryMode(DeliveryModeTransient)

	assert.Equal(t, DeliveryModeTransient, msg.DeliveryMode)
}

func TestMessage_IsExpired(t *testing.T) {
	msg := NewMessage("test", []byte("payload"))

	// No expiration set
	assert.False(t, msg.IsExpired())

	// Expired
	msg.Expiration = time.Now().Add(-1 * time.Hour)
	assert.True(t, msg.IsExpired())

	// Not expired
	msg.Expiration = time.Now().Add(1 * time.Hour)
	assert.False(t, msg.IsExpired())
}

func TestMessage_CanRetry(t *testing.T) {
	msg := NewMessage("test", []byte("payload"))
	msg.MaxRetries = 3

	assert.True(t, msg.CanRetry())

	msg.RetryCount = 2
	assert.True(t, msg.CanRetry())

	msg.RetryCount = 3
	assert.False(t, msg.CanRetry())

	msg.RetryCount = 4
	assert.False(t, msg.CanRetry())
}

func TestMessage_IncrementRetry(t *testing.T) {
	msg := NewMessage("test", []byte("payload"))
	assert.Equal(t, 0, msg.RetryCount)

	msg.IncrementRetry()
	assert.Equal(t, 1, msg.RetryCount)

	msg.IncrementRetry()
	assert.Equal(t, 2, msg.RetryCount)
}

func TestMessage_Clone(t *testing.T) {
	msg := NewMessage("test", []byte("payload"))
	msg.Headers["key"] = "value"
	msg.TraceID = "trace-123"
	msg.CorrelationID = "corr-123"
	msg.Priority = PriorityHigh
	msg.Partition = 5
	msg.Offset = 100

	clone := msg.Clone()

	// Verify all fields are copied
	assert.Equal(t, msg.ID, clone.ID)
	assert.Equal(t, msg.Type, clone.Type)
	assert.Equal(t, msg.Payload, clone.Payload)
	assert.Equal(t, msg.Headers, clone.Headers)
	assert.Equal(t, msg.TraceID, clone.TraceID)
	assert.Equal(t, msg.CorrelationID, clone.CorrelationID)
	assert.Equal(t, msg.Priority, clone.Priority)
	assert.Equal(t, msg.Partition, clone.Partition)
	assert.Equal(t, msg.Offset, clone.Offset)

	// Verify deep copy
	msg.Payload[0] = 'X'
	assert.NotEqual(t, msg.Payload, clone.Payload)

	msg.Headers["key"] = "modified"
	assert.NotEqual(t, msg.Headers["key"], clone.Headers["key"])
}

func TestMessage_MarshalJSON(t *testing.T) {
	msg := NewMessage("test", []byte(`{"data":"test"}`))
	msg.TraceID = "trace-123"

	data, err := json.Marshal(msg)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "test", decoded["type"])
	assert.Equal(t, "trace-123", decoded["trace_id"])
}

func TestDefaultBrokerConfig(t *testing.T) {
	config := DefaultBrokerConfig()

	assert.Equal(t, BrokerTypeInMemory, config.Type)
	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, 5672, config.Port)
	assert.Equal(t, "/", config.VirtualHost)
	assert.False(t, config.TLS)
	assert.Equal(t, 30*time.Second, config.ConnectTimeout)
	assert.Equal(t, 5*time.Second, config.ReconnectInterval)
	assert.Equal(t, 0, config.MaxReconnectAttempts)
	assert.Equal(t, 60*time.Second, config.Heartbeat)
}

func TestBrokerConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      *BrokerConfig
		expectError bool
	}{
		{
			name: "valid config",
			config: &BrokerConfig{
				Type: BrokerTypeRabbitMQ,
				Host: "localhost",
				Port: 5672,
			},
			expectError: false,
		},
		{
			name: "invalid broker type",
			config: &BrokerConfig{
				Type: BrokerType("invalid"),
				Host: "localhost",
				Port: 5672,
			},
			expectError: true,
		},
		{
			name: "empty host",
			config: &BrokerConfig{
				Type: BrokerTypeRabbitMQ,
				Host: "",
				Port: 5672,
			},
			expectError: true,
		},
		{
			name: "invalid port zero",
			config: &BrokerConfig{
				Type: BrokerTypeRabbitMQ,
				Host: "localhost",
				Port: 0,
			},
			expectError: true,
		},
		{
			name: "invalid port too high",
			config: &BrokerConfig{
				Type: BrokerTypeRabbitMQ,
				Host: "localhost",
				Port: 70000,
			},
			expectError: true,
		},
		{
			name: "TLS without cert",
			config: &BrokerConfig{
				Type:        BrokerTypeRabbitMQ,
				Host:        "localhost",
				Port:        5672,
				TLS:         true,
				TLSCertFile: "",
			},
			expectError: true,
		},
		{
			name: "TLS with cert",
			config: &BrokerConfig{
				Type:        BrokerTypeRabbitMQ,
				Host:        "localhost",
				Port:        5672,
				TLS:         true,
				TLSCertFile: "/path/to/cert.pem",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMessagePriority_Values(t *testing.T) {
	assert.Equal(t, MessagePriority(1), PriorityLow)
	assert.Equal(t, MessagePriority(5), PriorityNormal)
	assert.Equal(t, MessagePriority(8), PriorityHigh)
	assert.Equal(t, MessagePriority(10), PriorityCritical)
}

func TestDeliveryMode_Values(t *testing.T) {
	assert.Equal(t, DeliveryMode(1), DeliveryModeTransient)
	assert.Equal(t, DeliveryMode(2), DeliveryModePersistent)
}

func TestMessageState_Values(t *testing.T) {
	assert.Equal(t, MessageState("pending"), MessageStatePending)
	assert.Equal(t, MessageState("processing"), MessageStateProcessing)
	assert.Equal(t, MessageState("completed"), MessageStateCompleted)
	assert.Equal(t, MessageState("failed"), MessageStateFailed)
	assert.Equal(t, MessageState("dead_lettered"), MessageStateDeadLettered)
}

func TestGenerateMessageID(t *testing.T) {
	id1 := generateMessageID()
	time.Sleep(time.Millisecond)
	id2 := generateMessageID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
}

func TestRandomString(t *testing.T) {
	s1 := randomString(10)
	time.Sleep(time.Millisecond)
	s2 := randomString(10)

	assert.Len(t, s1, 10)
	assert.Len(t, s2, 10)
	// Note: There's a small chance they could be equal, but very unlikely
}
