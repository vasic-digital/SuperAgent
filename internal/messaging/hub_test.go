package messaging

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultHubConfig(t *testing.T) {
	config := DefaultHubConfig()

	require.NotNil(t, config)
	assert.True(t, config.TaskQueueEnabled)
	assert.True(t, config.EventStreamEnabled)
	assert.True(t, config.FallbackEnabled)
	assert.True(t, config.UseFallbackOnError)
	assert.Equal(t, 30*time.Second, config.HealthCheckInterval)
	assert.Equal(t, 5, config.CircuitBreakerThreshold)
	assert.Equal(t, 30*time.Second, config.CircuitBreakerTimeout)
	assert.NotNil(t, config.TaskQueueConfig)
	assert.NotNil(t, config.EventStreamConfig)
	assert.NotNil(t, config.RetryConfig)
}

func TestNewMessagingHub(t *testing.T) {
	t.Run("with default config", func(t *testing.T) {
		hub := NewMessagingHub(nil)

		require.NotNil(t, hub)
		assert.NotNil(t, hub.config)
		assert.NotNil(t, hub.router)
		assert.NotNil(t, hub.middleware)
		assert.NotNil(t, hub.metrics)
		assert.NotNil(t, hub.taskRegistry)
		assert.NotNil(t, hub.eventRegistry)
		assert.NotNil(t, hub.subscriptions)
		assert.NotNil(t, hub.stopCh)
		assert.False(t, hub.connected)
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &HubConfig{
			TaskQueueEnabled:    false,
			EventStreamEnabled:  true,
			FallbackEnabled:     true,
			HealthCheckInterval: 1 * time.Minute,
		}
		hub := NewMessagingHub(config)

		require.NotNil(t, hub)
		assert.False(t, hub.config.TaskQueueEnabled)
		assert.True(t, hub.config.EventStreamEnabled)
		assert.Equal(t, 1*time.Minute, hub.config.HealthCheckInterval)
	})
}

func TestNewHubMetrics(t *testing.T) {
	metrics := NewHubMetrics()

	require.NotNil(t, metrics)
	assert.NotNil(t, metrics.BrokerMetrics)
	assert.Nil(t, metrics.TaskQueueMetrics)
	assert.Nil(t, metrics.EventStreamMetrics)
	assert.Nil(t, metrics.FallbackMetrics)
	assert.Equal(t, int64(0), metrics.FallbackUsages)
}

func TestMessagingHub_SetBrokers(t *testing.T) {
	hub := NewMessagingHub(nil)

	// Note: We're not actually setting real brokers here
	// Just verifying the methods don't panic
	hub.SetTaskQueueBroker(nil)
	hub.SetEventStreamBroker(nil)
	hub.SetFallbackBroker(nil)

	assert.Nil(t, hub.taskQueue)
	assert.Nil(t, hub.eventStream)
	assert.Nil(t, hub.fallback)
}

func TestMessagingHub_IsConnected(t *testing.T) {
	hub := NewMessagingHub(nil)

	assert.False(t, hub.IsConnected())

	hub.connected = true
	assert.True(t, hub.IsConnected())
}

func TestMessagingHub_GetMetrics(t *testing.T) {
	hub := NewMessagingHub(nil)

	metrics := hub.GetMetrics()

	require.NotNil(t, metrics)
	assert.Same(t, hub.metrics, metrics)
}

func TestMessagingHub_Use(t *testing.T) {
	hub := NewMessagingHub(nil)

	// Add middleware
	middleware := func(next MessageHandler) MessageHandler {
		return func(ctx context.Context, msg *Message) error {
			return next(ctx, msg)
		}
	}

	hub.Use(middleware)

	// Middleware chain should have the middleware
	assert.NotNil(t, hub.middleware)
}

func TestMessagingHub_RegisterTaskHandler(t *testing.T) {
	hub := NewMessagingHub(nil)

	handler := func(ctx context.Context, task *Task) error {
		return nil
	}

	hub.RegisterTaskHandler("test_task", handler)

	// Registry should have the handler
	assert.NotNil(t, hub.taskRegistry)
}

func TestMessagingHub_RegisterEventHandler(t *testing.T) {
	hub := NewMessagingHub(nil)

	handler := func(ctx context.Context, event *Event) error {
		return nil
	}

	hub.RegisterEventHandler(EventTypeLLMRequestStarted, handler)

	// Registry should have the handler
	assert.NotNil(t, hub.eventRegistry)
}

func TestMessagingHub_EnqueueTask_NoQueue(t *testing.T) {
	hub := NewMessagingHub(&HubConfig{
		TaskQueueEnabled: true,
		FallbackEnabled:  false,
	})

	task := &Task{
		ID:      "test-task",
		Type:    "test",
		Payload: []byte("test payload"),
	}

	err := hub.EnqueueTask(context.Background(), "test-queue", task)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no task queue broker available")
}

func TestMessagingHub_PublishEvent_NoStream(t *testing.T) {
	hub := NewMessagingHub(&HubConfig{
		EventStreamEnabled: true,
		FallbackEnabled:    false,
	})

	event := &Event{
		ID:   "test-event",
		Type: EventTypeLLMRequestStarted,
		Data: []byte("test data"),
	}

	err := hub.PublishEvent(context.Background(), "test-topic", event)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no event stream broker available")
}

func TestMessagingHub_SubscribeTasks_NoQueue(t *testing.T) {
	hub := NewMessagingHub(&HubConfig{
		TaskQueueEnabled: true,
		FallbackEnabled:  false,
	})

	handler := func(ctx context.Context, task *Task) error {
		return nil
	}

	sub, err := hub.SubscribeTasks(context.Background(), "test-queue", handler)

	assert.Error(t, err)
	assert.Nil(t, sub)
	assert.Contains(t, err.Error(), "no task queue broker available")
}

func TestMessagingHub_SubscribeEvents_NoStream(t *testing.T) {
	hub := NewMessagingHub(&HubConfig{
		EventStreamEnabled: true,
		FallbackEnabled:    false,
	})

	handler := func(ctx context.Context, event *Event) error {
		return nil
	}

	sub, err := hub.SubscribeEvents(context.Background(), "test-topic", handler)

	assert.Error(t, err)
	assert.Nil(t, sub)
	assert.Contains(t, err.Error(), "no event stream broker available")
}

func TestMessagingHub_DeclareQueue_NoQueue(t *testing.T) {
	hub := NewMessagingHub(nil)

	// Without a task queue, this should be a no-op
	err := hub.DeclareQueue(context.Background(), "test-queue")

	assert.NoError(t, err)
}

func TestMessagingHub_CreateTopic_NoStream(t *testing.T) {
	hub := NewMessagingHub(nil)

	// Without an event stream, this should be a no-op
	err := hub.CreateTopic(context.Background(), "test-topic", 3, 1)

	assert.NoError(t, err)
}

func TestMessagingHub_GetQueueStats_NoQueue(t *testing.T) {
	hub := NewMessagingHub(nil)

	stats, err := hub.GetQueueStats(context.Background(), "test-queue")

	assert.Error(t, err)
	assert.Nil(t, stats)
}

func TestMessagingHub_GetTopicMetadata_NoStream(t *testing.T) {
	hub := NewMessagingHub(nil)

	metadata, err := hub.GetTopicMetadata(context.Background(), "test-topic")

	assert.Error(t, err)
	assert.Nil(t, metadata)
}

func TestMessagingHub_StreamEvents_NoStream(t *testing.T) {
	hub := NewMessagingHub(nil)

	ch, err := hub.StreamEvents(context.Background(), "test-topic")

	assert.Error(t, err)
	assert.Nil(t, ch)
}

func TestMessagingHub_HealthCheck_NoConnection(t *testing.T) {
	hub := NewMessagingHub(nil)

	err := hub.HealthCheck(context.Background())

	assert.NoError(t, err) // No error because no brokers connected
}

func TestMessagingHub_Publish(t *testing.T) {
	hub := NewMessagingHub(&HubConfig{
		FallbackEnabled: false,
	})

	msg := &Message{
		ID:      "test-msg",
		Payload: []byte("test"),
	}

	err := hub.Publish(context.Background(), "random-topic", msg)

	assert.Error(t, err)
}

func TestMessagingHub_Subscribe_NoQueue(t *testing.T) {
	hub := NewMessagingHub(&HubConfig{
		FallbackEnabled: false,
	})

	handler := func(ctx context.Context, msg *Message) error {
		return nil
	}

	sub, err := hub.Subscribe(context.Background(), "random-topic", handler)

	assert.Error(t, err)
	assert.Nil(t, sub)
}

func TestMessagingHub_EnqueueTaskBatch_NoQueue(t *testing.T) {
	hub := NewMessagingHub(&HubConfig{
		FallbackEnabled: false,
	})

	tasks := []*Task{
		{ID: "task-1", Payload: []byte("test1")},
		{ID: "task-2", Payload: []byte("test2")},
	}

	err := hub.EnqueueTaskBatch(context.Background(), "test-queue", tasks)

	assert.Error(t, err)
}

func TestMessagingHub_PublishEventBatch_NoStream(t *testing.T) {
	hub := NewMessagingHub(&HubConfig{
		FallbackEnabled: false,
	})

	events := []*Event{
		{ID: "event-1", Data: []byte("data1")},
		{ID: "event-2", Data: []byte("data2")},
	}

	err := hub.PublishEventBatch(context.Background(), "test-topic", events)

	assert.Error(t, err)
}

func TestNewMessageRouter(t *testing.T) {
	router := NewMessageRouter()

	require.NotNil(t, router)
	assert.NotEmpty(t, router.taskQueuePrefixes)
	assert.NotEmpty(t, router.eventStreamPrefixes)
}

func TestMessageRouter_IsTaskQueue(t *testing.T) {
	router := NewMessageRouter()

	tests := []struct {
		topic    string
		expected bool
	}{
		{"helixagent.tasks.test", true},
		{"tasks.process", true},
		{"events.notification", false},
		{"random-topic", false},
	}

	for _, tt := range tests {
		t.Run(tt.topic, func(t *testing.T) {
			assert.Equal(t, tt.expected, router.IsTaskQueue(tt.topic))
		})
	}
}

func TestMessageRouter_IsEventStream(t *testing.T) {
	router := NewMessageRouter()

	tests := []struct {
		topic    string
		expected bool
	}{
		{"helixagent.events.notification", true},
		{"helixagent.stream.data", true},
		{"events.user", true},
		{"tasks.process", false},
		{"random-topic", false},
	}

	for _, tt := range tests {
		t.Run(tt.topic, func(t *testing.T) {
			assert.Equal(t, tt.expected, router.IsEventStream(tt.topic))
		})
	}
}

func TestMessageRouter_AddTaskQueuePrefix(t *testing.T) {
	router := NewMessageRouter()
	initialLen := len(router.taskQueuePrefixes)

	router.AddTaskQueuePrefix("custom.tasks.")

	assert.Len(t, router.taskQueuePrefixes, initialLen+1)
	assert.True(t, router.IsTaskQueue("custom.tasks.test"))
}

func TestMessageRouter_AddEventStreamPrefix(t *testing.T) {
	router := NewMessageRouter()
	initialLen := len(router.eventStreamPrefixes)

	router.AddEventStreamPrefix("custom.events.")

	assert.Len(t, router.eventStreamPrefixes, initialLen+1)
	assert.True(t, router.IsEventStream("custom.events.test"))
}

func TestGlobalHub(t *testing.T) {
	// Initially nil
	assert.Nil(t, GetGlobalHub())

	// Set global hub
	hub := NewMessagingHub(nil)
	SetGlobalHub(hub)

	// Now should return the hub
	got := GetGlobalHub()
	assert.Same(t, hub, got)

	// Cleanup
	SetGlobalHub(nil)
}
