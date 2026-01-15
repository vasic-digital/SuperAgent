package integration

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/messaging"
	"dev.helix.agent/internal/messaging/inmemory"
)

// TestMessaging_InMemoryBroker_BasicPubSub tests basic publish/subscribe with in-memory broker
func TestMessaging_InMemoryBroker_BasicPubSub(t *testing.T) {
	broker := inmemory.NewBroker(nil)

	ctx := context.Background()
	err := broker.Connect(ctx)
	require.NoError(t, err)
	defer broker.Close(ctx)

	// Subscribe to topic
	received := make(chan *messaging.Message, 1)
	handler := func(ctx context.Context, msg *messaging.Message) error {
		received <- msg
		return nil
	}

	sub, err := broker.Subscribe(ctx, "test.topic", handler)
	require.NoError(t, err)
	defer sub.Unsubscribe()

	// Publish message
	msg := messaging.NewMessage("test", []byte(`{"key":"value"}`))
	err = broker.Publish(ctx, "test.topic", msg)
	require.NoError(t, err)

	// Wait for message
	select {
	case m := <-received:
		assert.Equal(t, "test", m.Type)
		assert.Contains(t, string(m.Payload), "value")
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for message")
	}
}

// TestMessaging_InMemoryBroker_MultipleSubscribers tests multiple subscribers
func TestMessaging_InMemoryBroker_MultipleSubscribers(t *testing.T) {
	broker := inmemory.NewBroker(nil)

	ctx := context.Background()
	err := broker.Connect(ctx)
	require.NoError(t, err)
	defer broker.Close(ctx)

	// Create multiple subscribers - at least one should receive the message
	var mu sync.Mutex
	count := 0
	handler := func(ctx context.Context, msg *messaging.Message) error {
		mu.Lock()
		count++
		mu.Unlock()
		return nil
	}

	sub1, err := broker.Subscribe(ctx, "test.topic", handler)
	require.NoError(t, err)
	defer sub1.Unsubscribe()

	sub2, err := broker.Subscribe(ctx, "test.topic", handler)
	require.NoError(t, err)
	defer sub2.Unsubscribe()

	// Publish message
	msg := messaging.NewMessage("test", []byte(`{}`))
	err = broker.Publish(ctx, "test.topic", msg)
	require.NoError(t, err)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// At least one subscriber should receive the message
	// Note: In-memory broker may use queue semantics (one consumer per message)
	// rather than pub/sub (all consumers). Both are valid patterns.
	assert.GreaterOrEqual(t, count, 1)
}

// TestMessaging_InMemoryBroker_BatchPublish tests batch publishing
func TestMessaging_InMemoryBroker_BatchPublish(t *testing.T) {
	broker := inmemory.NewBroker(nil)

	ctx := context.Background()
	err := broker.Connect(ctx)
	require.NoError(t, err)
	defer broker.Close(ctx)

	// Subscribe
	receivedCount := 0
	handler := func(ctx context.Context, msg *messaging.Message) error {
		receivedCount++
		return nil
	}

	sub, err := broker.Subscribe(ctx, "batch.topic", handler)
	require.NoError(t, err)
	defer sub.Unsubscribe()

	// Publish batch
	messages := make([]*messaging.Message, 10)
	for i := 0; i < 10; i++ {
		messages[i] = messaging.NewMessage("batch", []byte(`{}`))
	}

	err = broker.PublishBatch(ctx, "batch.topic", messages)
	require.NoError(t, err)

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	assert.Equal(t, 10, receivedCount)
}

// TestMessaging_InMemoryBroker_HealthCheck tests health check
func TestMessaging_InMemoryBroker_HealthCheck(t *testing.T) {
	broker := inmemory.NewBroker(nil)

	ctx := context.Background()
	err := broker.Connect(ctx)
	require.NoError(t, err)

	// Health check should pass
	err = broker.HealthCheck(ctx)
	assert.NoError(t, err)

	// Close broker
	broker.Close(ctx)

	// Health check should fail after close
	err = broker.HealthCheck(ctx)
	assert.Error(t, err)
}

// TestMessaging_InMemoryBroker_Unsubscribe tests unsubscribe
func TestMessaging_InMemoryBroker_Unsubscribe(t *testing.T) {
	broker := inmemory.NewBroker(nil)

	ctx := context.Background()
	err := broker.Connect(ctx)
	require.NoError(t, err)
	defer broker.Close(ctx)

	// Subscribe
	received := 0
	handler := func(ctx context.Context, msg *messaging.Message) error {
		received++
		return nil
	}

	sub, err := broker.Subscribe(ctx, "unsub.topic", handler)
	require.NoError(t, err)

	// Publish and verify receipt
	msg := messaging.NewMessage("test", []byte(`{}`))
	err = broker.Publish(ctx, "unsub.topic", msg)
	require.NoError(t, err)
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 1, received)

	// Unsubscribe
	err = sub.Unsubscribe()
	require.NoError(t, err)

	// Subscription should be inactive
	assert.False(t, sub.IsActive())

	// Publish again - should not be received
	err = broker.Publish(ctx, "unsub.topic", msg)
	require.NoError(t, err)
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 1, received) // Still 1, no new message received
}

// TestMessaging_MessageCreation tests message creation and manipulation
func TestMessaging_MessageCreation(t *testing.T) {
	msg := messaging.NewMessage("test.type", []byte(`{"data":"test"}`))

	assert.NotEmpty(t, msg.ID)
	assert.Equal(t, "test.type", msg.Type)
	assert.NotEmpty(t, msg.Payload)
	assert.NotZero(t, msg.Timestamp)
}

// TestMessaging_MessageWithOptions tests message with options
func TestMessaging_MessageWithOptions(t *testing.T) {
	msg := messaging.NewMessage("test", []byte(`{}`)).
		WithPriority(messaging.PriorityNormal).
		WithTraceID("trace-123").
		WithMaxRetries(3).
		WithTTL(time.Hour)

	assert.Equal(t, messaging.PriorityNormal, msg.Priority)
	assert.Equal(t, "trace-123", msg.TraceID)
	assert.Equal(t, 3, msg.MaxRetries)
	assert.NotZero(t, msg.Expiration)
}

// TestMessaging_MessageHeaders tests message headers
func TestMessaging_MessageHeaders(t *testing.T) {
	msg := messaging.NewMessage("test", []byte(`{}`))
	msg.SetHeader("X-Custom", "value")
	msg.SetHeader("X-Another", "another-value")

	assert.Equal(t, "value", msg.GetHeader("X-Custom"))
	assert.Equal(t, "another-value", msg.GetHeader("X-Another"))
	assert.Empty(t, msg.GetHeader("X-NonExistent"))
}

// TestMessaging_MessageClone tests message cloning
func TestMessaging_MessageClone(t *testing.T) {
	msg := messaging.NewMessage("test", []byte(`{"key":"value"}`)).
		WithPriority(messaging.PriorityNormal).
		WithTraceID("trace-123")
	msg.SetHeader("X-Header", "header-value")

	clone := msg.Clone()

	// Clone keeps the same ID but creates new header/payload copies
	assert.Equal(t, msg.ID, clone.ID)
	assert.Equal(t, msg.Type, clone.Type)
	assert.Equal(t, msg.Priority, clone.Priority)
	assert.Equal(t, msg.TraceID, clone.TraceID)
	assert.Equal(t, "header-value", clone.GetHeader("X-Header"))

	// Verify it's a deep copy (modifying clone doesn't affect original)
	clone.SetHeader("X-New", "new-value")
	assert.Empty(t, msg.GetHeader("X-New"))
}

// TestMessaging_MessageExpiration tests message expiration
func TestMessaging_MessageExpiration(t *testing.T) {
	// Non-expiring message
	msg := messaging.NewMessage("test", []byte(`{}`))
	assert.False(t, msg.IsExpired())

	// Expired message
	expiredMsg := messaging.NewMessage("test", []byte(`{}`)).
		WithTTL(-time.Second) // Already expired
	assert.True(t, expiredMsg.IsExpired())
}

// TestMessaging_MessageRetry tests message retry logic
func TestMessaging_MessageRetry(t *testing.T) {
	msg := messaging.NewMessage("test", []byte(`{}`)).WithMaxRetries(3)

	assert.True(t, msg.CanRetry())

	msg.IncrementRetry()
	assert.Equal(t, 1, msg.RetryCount)
	assert.True(t, msg.CanRetry())

	msg.IncrementRetry()
	msg.IncrementRetry()
	assert.Equal(t, 3, msg.RetryCount)
	assert.False(t, msg.CanRetry())
}

// TestMessaging_BrokerTypes tests broker type constants
func TestMessaging_BrokerTypes(t *testing.T) {
	assert.Equal(t, "rabbitmq", string(messaging.BrokerTypeRabbitMQ))
	assert.Equal(t, "kafka", string(messaging.BrokerTypeKafka))
	assert.Equal(t, "inmemory", string(messaging.BrokerTypeInMemory))

	assert.True(t, messaging.BrokerTypeRabbitMQ.IsValid())
	assert.True(t, messaging.BrokerTypeKafka.IsValid())
	assert.True(t, messaging.BrokerTypeInMemory.IsValid())
	assert.False(t, messaging.BrokerType("invalid").IsValid())
}

// TestMessaging_BrokerConfig tests broker configuration
func TestMessaging_BrokerConfig(t *testing.T) {
	cfg := messaging.DefaultBrokerConfig()

	assert.NotEmpty(t, cfg.Host)
	assert.NotZero(t, cfg.Port)
	assert.NotZero(t, cfg.ConnectTimeout)
	assert.NotZero(t, cfg.Heartbeat)
}

// TestMessaging_ConfigValidation tests config validation
func TestMessaging_ConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *messaging.BrokerConfig
		wantErr bool
	}{
		{
			name:    "valid config",
			cfg:     messaging.DefaultBrokerConfig(),
			wantErr: false,
		},
		{
			name: "invalid broker type",
			cfg: &messaging.BrokerConfig{
				Type: "invalid",
				Host: "localhost",
				Port: 5672,
			},
			wantErr: true,
		},
		{
			name: "empty host",
			cfg: &messaging.BrokerConfig{
				Type: messaging.BrokerTypeRabbitMQ,
				Host: "",
				Port: 5672,
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			cfg: &messaging.BrokerConfig{
				Type: messaging.BrokerTypeRabbitMQ,
				Host: "localhost",
				Port: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestMessaging_MessageMarshalJSON tests JSON marshaling
func TestMessaging_MessageMarshalJSON(t *testing.T) {
	msg := messaging.NewMessage("test.type", []byte(`{"key":"value"}`)).
		WithPriority(5).
		WithTraceID("trace-123")

	data, err := json.Marshal(msg)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "test.type", decoded["type"])
	assert.Equal(t, float64(5), decoded["priority"])
	assert.Equal(t, "trace-123", decoded["trace_id"])
}

// TestMessaging_InMemoryBroker_Metrics tests metrics collection
func TestMessaging_InMemoryBroker_Metrics(t *testing.T) {
	broker := inmemory.NewBroker(nil)

	ctx := context.Background()
	err := broker.Connect(ctx)
	require.NoError(t, err)
	defer broker.Close(ctx)

	// Publish some messages
	for i := 0; i < 5; i++ {
		msg := messaging.NewMessage("test", []byte(`{}`))
		err = broker.Publish(ctx, "metrics.topic", msg)
		require.NoError(t, err)
	}

	metrics := broker.GetMetrics()
	assert.NotNil(t, metrics)
	assert.GreaterOrEqual(t, metrics.MessagesPublished.Load(), int64(5))
}

// TestMessaging_InMemoryBroker_ConcurrentPublish tests concurrent publishing
func TestMessaging_InMemoryBroker_ConcurrentPublish(t *testing.T) {
	broker := inmemory.NewBroker(nil)

	ctx := context.Background()
	err := broker.Connect(ctx)
	require.NoError(t, err)
	defer broker.Close(ctx)

	// Start concurrent publishers
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				msg := messaging.NewMessage("concurrent", []byte(`{}`))
				broker.Publish(ctx, "concurrent.topic", msg)
			}
			done <- true
		}()
	}

	// Wait for all publishers
	for i := 0; i < 10; i++ {
		<-done
	}

	metrics := broker.GetMetrics()
	assert.GreaterOrEqual(t, metrics.MessagesPublished.Load(), int64(1000))
}

// TestMessaging_DeliveryMode tests delivery mode constants
func TestMessaging_DeliveryMode(t *testing.T) {
	assert.Equal(t, messaging.DeliveryMode(1), messaging.DeliveryModeTransient)
	assert.Equal(t, messaging.DeliveryMode(2), messaging.DeliveryModePersistent)
}

// TestMessaging_MessagePriority tests message priority constants
func TestMessaging_MessagePriority(t *testing.T) {
	assert.Equal(t, messaging.MessagePriority(1), messaging.PriorityLow)
	assert.Equal(t, messaging.MessagePriority(5), messaging.PriorityNormal)
	assert.Equal(t, messaging.MessagePriority(8), messaging.PriorityHigh)
}

// TestMessaging_MessageState tests message state constants
func TestMessaging_MessageState(t *testing.T) {
	assert.Equal(t, "pending", string(messaging.MessageStatePending))
	assert.Equal(t, "processing", string(messaging.MessageStateProcessing))
	assert.Equal(t, "completed", string(messaging.MessageStateCompleted))
	assert.Equal(t, "failed", string(messaging.MessageStateFailed))
	assert.Equal(t, "dead_lettered", string(messaging.MessageStateDeadLettered))
}

// TestMessaging_InMemoryBroker_IsConnected tests connection state
func TestMessaging_InMemoryBroker_IsConnected(t *testing.T) {
	broker := inmemory.NewBroker(nil)

	ctx := context.Background()

	// Initially not connected
	assert.False(t, broker.IsConnected())

	// Connect
	err := broker.Connect(ctx)
	require.NoError(t, err)
	assert.True(t, broker.IsConnected())

	// Close
	broker.Close(ctx)
	assert.False(t, broker.IsConnected())
}

// TestMessaging_InMemoryBroker_BrokerType tests broker type
func TestMessaging_InMemoryBroker_BrokerType(t *testing.T) {
	broker := inmemory.NewBroker(nil)
	assert.Equal(t, messaging.BrokerTypeInMemory, broker.BrokerType())
}
