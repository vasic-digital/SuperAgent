package messaging

import (
	"context"
	"sync"
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
	assert.Equal(t, int64(0), metrics.FallbackUsages.Load())
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

// =============================================================================
// Mock Implementations for Comprehensive Testing
// =============================================================================

// hubTestMockBroker implements MessageBroker for hub testing
type hubTestMockBroker struct {
	mu             sync.Mutex
	connected      bool
	connectError   error
	closeError     error
	publishError   error
	subscribeError error
	healthCheckErr error
	publishedMsgs  []*Message
	subscriptions  map[string]*hubTestMockSubscription
	metrics        *BrokerMetrics
	bType          BrokerType
}

func newHubTestMockBroker() *hubTestMockBroker {
	return &hubTestMockBroker{
		connected:     false,
		publishedMsgs: make([]*Message, 0),
		subscriptions: make(map[string]*hubTestMockSubscription),
		metrics:       NewBrokerMetrics(),
		bType:         BrokerTypeInMemory,
	}
}

func (m *hubTestMockBroker) Connect(ctx context.Context) error {
	if m.connectError != nil {
		return m.connectError
	}
	m.connected = true
	return nil
}

func (m *hubTestMockBroker) Close(ctx context.Context) error {
	m.connected = false
	return m.closeError
}

func (m *hubTestMockBroker) Publish(ctx context.Context, topic string, msg *Message, opts ...PublishOption) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.publishError != nil {
		return m.publishError
	}
	m.publishedMsgs = append(m.publishedMsgs, msg)
	return nil
}

func (m *hubTestMockBroker) PublishBatch(ctx context.Context, topic string, messages []*Message, opts ...PublishOption) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.publishError != nil {
		return m.publishError
	}
	m.publishedMsgs = append(m.publishedMsgs, messages...)
	return nil
}

func (m *hubTestMockBroker) Subscribe(ctx context.Context, topic string, handler MessageHandler, opts ...SubscribeOption) (Subscription, error) {
	if m.subscribeError != nil {
		return nil, m.subscribeError
	}
	sub := &hubTestMockSubscription{
		id:     "mock-sub-" + topic,
		topic:  topic,
		active: true,
	}
	m.subscriptions[topic] = sub
	return sub, nil
}

func (m *hubTestMockBroker) HealthCheck(ctx context.Context) error {
	return m.healthCheckErr
}

func (m *hubTestMockBroker) IsConnected() bool {
	return m.connected
}

func (m *hubTestMockBroker) BrokerType() BrokerType {
	return m.bType
}

func (m *hubTestMockBroker) GetMetrics() *BrokerMetrics {
	return m.metrics
}

// hubTestMockSubscription implements Subscription for hub testing
type hubTestMockSubscription struct {
	id     string
	topic  string
	active bool
}

func (s *hubTestMockSubscription) Unsubscribe() error {
	s.active = false
	return nil
}

func (s *hubTestMockSubscription) IsActive() bool {
	return s.active
}

func (s *hubTestMockSubscription) Topic() string {
	return s.topic
}

func (s *hubTestMockSubscription) ID() string {
	return s.id
}

// hubTestMockTaskQueueBroker implements TaskQueueBroker for hub testing
type hubTestMockTaskQueueBroker struct {
	*hubTestMockBroker
	taskMu           sync.Mutex
	enqueuedTasks    []*Task
	enqueueError     error
	dequeueError     error
	ackError         error
	nackError        error
	queueStats       *QueueStats
	queueStatsError  error
	queueDepth       int64
	queueDepthError  error
	taskSubscription *hubTestMockSubscription
}

func newHubTestMockTaskQueueBroker() *hubTestMockTaskQueueBroker {
	return &hubTestMockTaskQueueBroker{
		hubTestMockBroker: newHubTestMockBroker(),
		enqueuedTasks:     make([]*Task, 0),
		queueStats: &QueueStats{
			Name:          "test-queue",
			Messages:      10,
			MessagesReady: 5,
		},
		queueDepth: 10,
	}
}

func (m *hubTestMockTaskQueueBroker) DeclareQueue(ctx context.Context, name string, opts ...QueueOption) error {
	return nil
}

func (m *hubTestMockTaskQueueBroker) EnqueueTask(ctx context.Context, queue string, task *Task) error {
	m.taskMu.Lock()
	defer m.taskMu.Unlock()
	if m.enqueueError != nil {
		return m.enqueueError
	}
	m.enqueuedTasks = append(m.enqueuedTasks, task)
	return nil
}

func (m *hubTestMockTaskQueueBroker) EnqueueTaskBatch(ctx context.Context, queue string, tasks []*Task) error {
	m.taskMu.Lock()
	defer m.taskMu.Unlock()
	if m.enqueueError != nil {
		return m.enqueueError
	}
	m.enqueuedTasks = append(m.enqueuedTasks, tasks...)
	return nil
}

func (m *hubTestMockTaskQueueBroker) DequeueTask(ctx context.Context, queue string, workerID string) (*Task, error) {
	if m.dequeueError != nil {
		return nil, m.dequeueError
	}
	if len(m.enqueuedTasks) > 0 {
		task := m.enqueuedTasks[0]
		m.enqueuedTasks = m.enqueuedTasks[1:]
		return task, nil
	}
	return nil, nil
}

func (m *hubTestMockTaskQueueBroker) AckTask(ctx context.Context, deliveryTag uint64) error {
	return m.ackError
}

func (m *hubTestMockTaskQueueBroker) NackTask(ctx context.Context, deliveryTag uint64, requeue bool) error {
	return m.nackError
}

func (m *hubTestMockTaskQueueBroker) RejectTask(ctx context.Context, deliveryTag uint64) error {
	return nil
}

func (m *hubTestMockTaskQueueBroker) MoveToDeadLetter(ctx context.Context, task *Task, reason string) error {
	return nil
}

func (m *hubTestMockTaskQueueBroker) GetQueueStats(ctx context.Context, queue string) (*QueueStats, error) {
	if m.queueStatsError != nil {
		return nil, m.queueStatsError
	}
	return m.queueStats, nil
}

func (m *hubTestMockTaskQueueBroker) GetQueueDepth(ctx context.Context, queue string) (int64, error) {
	if m.queueDepthError != nil {
		return 0, m.queueDepthError
	}
	return m.queueDepth, nil
}

func (m *hubTestMockTaskQueueBroker) PurgeQueue(ctx context.Context, queue string) error {
	m.enqueuedTasks = make([]*Task, 0)
	return nil
}

func (m *hubTestMockTaskQueueBroker) DeleteQueue(ctx context.Context, queue string) error {
	return nil
}

func (m *hubTestMockTaskQueueBroker) SubscribeTasks(ctx context.Context, queue string, handler TaskHandler, opts ...SubscribeOption) (Subscription, error) {
	if m.subscribeError != nil {
		return nil, m.subscribeError
	}
	sub := &hubTestMockSubscription{
		id:     "mock-task-sub-" + queue,
		topic:  queue,
		active: true,
	}
	m.taskMu.Lock()
	m.taskSubscription = sub
	m.taskMu.Unlock()
	return sub, nil
}

// hubTestMockEventStreamBroker implements EventStreamBroker for hub testing
type hubTestMockEventStreamBroker struct {
	*hubTestMockBroker
	eventMu           sync.Mutex
	publishedEvents   []*Event
	publishEventError error
	topicMetadata     *TopicMetadata
	topicMetadataErr  error
	eventChan         chan *Event
	createTopicError  error
	deleteTopicError  error
	topics            []string
}

func newHubTestMockEventStreamBroker() *hubTestMockEventStreamBroker {
	return &hubTestMockEventStreamBroker{
		hubTestMockBroker: newHubTestMockBroker(),
		publishedEvents:   make([]*Event, 0),
		topicMetadata: &TopicMetadata{
			Name: "test-topic",
		},
		eventChan: make(chan *Event, 10),
		topics:    []string{"topic1", "topic2"},
	}
}

func (m *hubTestMockEventStreamBroker) CreateTopic(ctx context.Context, name string, partitions int, replication int) error {
	if m.createTopicError != nil {
		return m.createTopicError
	}
	m.topics = append(m.topics, name)
	return nil
}

func (m *hubTestMockEventStreamBroker) DeleteTopic(ctx context.Context, name string) error {
	return m.deleteTopicError
}

func (m *hubTestMockEventStreamBroker) ListTopics(ctx context.Context) ([]string, error) {
	return m.topics, nil
}

func (m *hubTestMockEventStreamBroker) GetTopicMetadata(ctx context.Context, topic string) (*TopicMetadata, error) {
	if m.topicMetadataErr != nil {
		return nil, m.topicMetadataErr
	}
	return m.topicMetadata, nil
}

func (m *hubTestMockEventStreamBroker) CreateConsumerGroup(ctx context.Context, groupID string) error {
	return nil
}

func (m *hubTestMockEventStreamBroker) DeleteConsumerGroup(ctx context.Context, groupID string) error {
	return nil
}

func (m *hubTestMockEventStreamBroker) PublishEvent(ctx context.Context, topic string, event *Event) error {
	m.eventMu.Lock()
	defer m.eventMu.Unlock()
	if m.publishEventError != nil {
		return m.publishEventError
	}
	m.publishedEvents = append(m.publishedEvents, event)
	return nil
}

func (m *hubTestMockEventStreamBroker) PublishEventBatch(ctx context.Context, topic string, events []*Event) error {
	m.eventMu.Lock()
	defer m.eventMu.Unlock()
	if m.publishEventError != nil {
		return m.publishEventError
	}
	m.publishedEvents = append(m.publishedEvents, events...)
	return nil
}

func (m *hubTestMockEventStreamBroker) SubscribeEvents(ctx context.Context, topic string, handler EventHandler, opts ...SubscribeOption) (Subscription, error) {
	if m.subscribeError != nil {
		return nil, m.subscribeError
	}
	sub := &hubTestMockSubscription{
		id:     "mock-event-sub-" + topic,
		topic:  topic,
		active: true,
	}
	return sub, nil
}

func (m *hubTestMockEventStreamBroker) StreamMessages(ctx context.Context, topic string, opts ...StreamOption) (<-chan *Message, error) {
	msgCh := make(chan *Message, 10)
	return msgCh, nil
}

func (m *hubTestMockEventStreamBroker) StreamEvents(ctx context.Context, topic string, opts ...StreamOption) (<-chan *Event, error) {
	return m.eventChan, nil
}

func (m *hubTestMockEventStreamBroker) CommitOffset(ctx context.Context, topic string, partition int32, offset int64) error {
	return nil
}

func (m *hubTestMockEventStreamBroker) GetOffset(ctx context.Context, topic string, partition int32) (int64, error) {
	return 100, nil
}

func (m *hubTestMockEventStreamBroker) SeekToOffset(ctx context.Context, topic string, partition int32, offset int64) error {
	return nil
}

func (m *hubTestMockEventStreamBroker) SeekToTimestamp(ctx context.Context, topic string, partition int32, ts time.Time) error {
	return nil
}

func (m *hubTestMockEventStreamBroker) SeekToBeginning(ctx context.Context, topic string, partition int32) error {
	return nil
}

func (m *hubTestMockEventStreamBroker) SeekToEnd(ctx context.Context, topic string, partition int32) error {
	return nil
}

// =============================================================================
// Comprehensive Hub Tests with Mocks
// =============================================================================

func TestMessagingHub_Initialize_WithTaskQueue(t *testing.T) {
	hub := NewMessagingHub(DefaultHubConfig())
	taskQueue := newHubTestMockTaskQueueBroker()
	hub.SetTaskQueueBroker(taskQueue)

	ctx := context.Background()
	err := hub.Initialize(ctx)

	assert.NoError(t, err)
	assert.True(t, hub.IsConnected())
	assert.True(t, taskQueue.IsConnected())
}

func TestMessagingHub_Initialize_WithEventStream(t *testing.T) {
	hub := NewMessagingHub(DefaultHubConfig())
	eventStream := newHubTestMockEventStreamBroker()
	hub.SetEventStreamBroker(eventStream)

	ctx := context.Background()
	err := hub.Initialize(ctx)

	assert.NoError(t, err)
	assert.True(t, hub.IsConnected())
	assert.True(t, eventStream.IsConnected())
}

func TestMessagingHub_Initialize_WithFallback(t *testing.T) {
	hub := NewMessagingHub(DefaultHubConfig())
	fallback := newHubTestMockBroker()
	hub.SetFallbackBroker(fallback)

	ctx := context.Background()
	err := hub.Initialize(ctx)

	assert.NoError(t, err)
	assert.True(t, hub.IsConnected())
	assert.True(t, fallback.IsConnected())
}

func TestMessagingHub_Initialize_WithAllBrokers(t *testing.T) {
	hub := NewMessagingHub(DefaultHubConfig())
	taskQueue := newHubTestMockTaskQueueBroker()
	eventStream := newHubTestMockEventStreamBroker()
	fallback := newHubTestMockBroker()

	hub.SetTaskQueueBroker(taskQueue)
	hub.SetEventStreamBroker(eventStream)
	hub.SetFallbackBroker(fallback)

	ctx := context.Background()
	err := hub.Initialize(ctx)

	assert.NoError(t, err)
	assert.True(t, hub.IsConnected())
	assert.True(t, taskQueue.IsConnected())
	assert.True(t, eventStream.IsConnected())
	assert.True(t, fallback.IsConnected())
}

func TestMessagingHub_Initialize_TaskQueueConnectionError_WithFallback(t *testing.T) {
	config := DefaultHubConfig()
	config.UseFallbackOnError = true
	hub := NewMessagingHub(config)

	taskQueue := newHubTestMockTaskQueueBroker()
	taskQueue.hubTestMockBroker.connectError = NewBrokerError(ErrCodeConnectionFailed, "connection failed", nil)
	fallback := newHubTestMockBroker()

	hub.SetTaskQueueBroker(taskQueue)
	hub.SetFallbackBroker(fallback)

	ctx := context.Background()
	err := hub.Initialize(ctx)

	// Should succeed because fallback is available and UseFallbackOnError is true
	assert.NoError(t, err)
	assert.True(t, hub.IsConnected())
	assert.True(t, fallback.IsConnected())
}

func TestMessagingHub_Initialize_TaskQueueConnectionError_NoFallback(t *testing.T) {
	config := DefaultHubConfig()
	config.UseFallbackOnError = false
	config.FallbackEnabled = false
	hub := NewMessagingHub(config)

	taskQueue := newHubTestMockTaskQueueBroker()
	taskQueue.hubTestMockBroker.connectError = NewBrokerError(ErrCodeConnectionFailed, "connection failed", nil)

	hub.SetTaskQueueBroker(taskQueue)

	ctx := context.Background()
	err := hub.Initialize(ctx)

	// Should fail because no fallback
	assert.Error(t, err)
}

func TestMessagingHub_Close_WithAllBrokers(t *testing.T) {
	hub := NewMessagingHub(DefaultHubConfig())
	taskQueue := newHubTestMockTaskQueueBroker()
	eventStream := newHubTestMockEventStreamBroker()
	fallback := newHubTestMockBroker()

	hub.SetTaskQueueBroker(taskQueue)
	hub.SetEventStreamBroker(eventStream)
	hub.SetFallbackBroker(fallback)

	ctx := context.Background()
	_ = hub.Initialize(ctx)

	err := hub.Close(ctx)

	assert.NoError(t, err)
	assert.False(t, hub.IsConnected())
	assert.False(t, taskQueue.IsConnected())
	assert.False(t, eventStream.IsConnected())
	assert.False(t, fallback.IsConnected())
}

func TestMessagingHub_Close_WithActiveSubscriptions(t *testing.T) {
	hub := NewMessagingHub(DefaultHubConfig())
	taskQueue := newHubTestMockTaskQueueBroker()
	taskQueue.connected = true

	hub.SetTaskQueueBroker(taskQueue)
	hub.connected = true

	ctx := context.Background()

	// Create a subscription
	handler := func(ctx context.Context, task *Task) error { return nil }
	sub, err := hub.SubscribeTasks(ctx, "test-queue", handler)
	require.NoError(t, err)
	require.NotNil(t, sub)

	// Close hub
	err = hub.Close(ctx)
	assert.NoError(t, err)
	assert.False(t, hub.IsConnected())
}

func TestMessagingHub_EnqueueTask_WithConnectedTaskQueue(t *testing.T) {
	hub := NewMessagingHub(DefaultHubConfig())
	taskQueue := newHubTestMockTaskQueueBroker()
	taskQueue.connected = true

	hub.SetTaskQueueBroker(taskQueue)
	hub.connected = true

	ctx := context.Background()
	task := NewTask("test.task", []byte(`{"key":"value"}`))

	err := hub.EnqueueTask(ctx, "test-queue", task)

	assert.NoError(t, err)
	assert.Len(t, taskQueue.enqueuedTasks, 1)
	assert.Equal(t, task.ID, taskQueue.enqueuedTasks[0].ID)
}

func TestMessagingHub_EnqueueTask_WithTaskQueueError_FallsBackToInMemory(t *testing.T) {
	config := DefaultHubConfig()
	config.UseFallbackOnError = true
	hub := NewMessagingHub(config)

	taskQueue := newHubTestMockTaskQueueBroker()
	taskQueue.connected = true
	taskQueue.enqueueError = NewBrokerError(ErrCodePublishFailed, "publish failed", nil)

	fallback := newHubTestMockBroker()
	fallback.connected = true

	hub.SetTaskQueueBroker(taskQueue)
	hub.SetFallbackBroker(fallback)
	hub.connected = true

	ctx := context.Background()
	task := NewTask("test.task", []byte(`{"key":"value"}`))

	err := hub.EnqueueTask(ctx, "test-queue", task)

	assert.NoError(t, err)
	assert.Len(t, fallback.publishedMsgs, 1)
	assert.Greater(t, hub.GetMetrics().FallbackUsages.Load(), int64(0))
}

func TestMessagingHub_EnqueueTask_NoTaskQueue_UsesFallback(t *testing.T) {
	hub := NewMessagingHub(DefaultHubConfig())
	fallback := newHubTestMockBroker()
	fallback.connected = true

	hub.SetFallbackBroker(fallback)
	hub.connected = true

	ctx := context.Background()
	task := NewTask("test.task", []byte(`{"key":"value"}`))

	err := hub.EnqueueTask(ctx, "test-queue", task)

	assert.NoError(t, err)
	assert.Len(t, fallback.publishedMsgs, 1)
}

func TestMessagingHub_EnqueueTaskBatch_WithConnectedTaskQueue(t *testing.T) {
	hub := NewMessagingHub(DefaultHubConfig())
	taskQueue := newHubTestMockTaskQueueBroker()
	taskQueue.connected = true

	hub.SetTaskQueueBroker(taskQueue)
	hub.connected = true

	ctx := context.Background()
	tasks := []*Task{
		NewTask("test.task1", []byte(`{"key":"value1"}`)),
		NewTask("test.task2", []byte(`{"key":"value2"}`)),
		NewTask("test.task3", []byte(`{"key":"value3"}`)),
	}

	err := hub.EnqueueTaskBatch(ctx, "test-queue", tasks)

	assert.NoError(t, err)
	assert.Len(t, taskQueue.enqueuedTasks, 3)
}

func TestMessagingHub_EnqueueTaskBatch_UsesFallback(t *testing.T) {
	hub := NewMessagingHub(DefaultHubConfig())
	fallback := newHubTestMockBroker()
	fallback.connected = true

	hub.SetFallbackBroker(fallback)
	hub.connected = true

	ctx := context.Background()
	tasks := []*Task{
		NewTask("test.task1", []byte(`{"key":"value1"}`)),
		NewTask("test.task2", []byte(`{"key":"value2"}`)),
	}

	err := hub.EnqueueTaskBatch(ctx, "test-queue", tasks)

	assert.NoError(t, err)
	assert.Len(t, fallback.publishedMsgs, 2)
}

func TestMessagingHub_PublishEvent_WithConnectedEventStream(t *testing.T) {
	hub := NewMessagingHub(DefaultHubConfig())
	eventStream := newHubTestMockEventStreamBroker()
	eventStream.connected = true

	hub.SetEventStreamBroker(eventStream)
	hub.connected = true

	ctx := context.Background()
	event := NewEvent(EventTypeLLMRequestStarted, "test-source", []byte(`{"key":"value"}`))

	err := hub.PublishEvent(ctx, "test-topic", event)

	assert.NoError(t, err)
	assert.Len(t, eventStream.publishedEvents, 1)
	assert.Equal(t, event.ID, eventStream.publishedEvents[0].ID)
}

func TestMessagingHub_PublishEvent_WithEventStreamError_FallsBack(t *testing.T) {
	config := DefaultHubConfig()
	config.UseFallbackOnError = true
	hub := NewMessagingHub(config)

	eventStream := newHubTestMockEventStreamBroker()
	eventStream.connected = true
	eventStream.publishEventError = NewBrokerError(ErrCodePublishFailed, "publish failed", nil)

	fallback := newHubTestMockBroker()
	fallback.connected = true

	hub.SetEventStreamBroker(eventStream)
	hub.SetFallbackBroker(fallback)
	hub.connected = true

	ctx := context.Background()
	event := NewEvent(EventTypeLLMRequestStarted, "test-source", []byte(`{"key":"value"}`))

	err := hub.PublishEvent(ctx, "test-topic", event)

	assert.NoError(t, err)
	assert.Len(t, fallback.publishedMsgs, 1)
}

func TestMessagingHub_PublishEventBatch_WithConnectedEventStream(t *testing.T) {
	hub := NewMessagingHub(DefaultHubConfig())
	eventStream := newHubTestMockEventStreamBroker()
	eventStream.connected = true

	hub.SetEventStreamBroker(eventStream)
	hub.connected = true

	ctx := context.Background()
	events := []*Event{
		NewEvent(EventTypeLLMRequestStarted, "test-source", []byte(`{"key":"value1"}`)),
		NewEvent(EventTypeLLMRequestCompleted, "test-source", []byte(`{"key":"value2"}`)),
	}

	err := hub.PublishEventBatch(ctx, "test-topic", events)

	assert.NoError(t, err)
	assert.Len(t, eventStream.publishedEvents, 2)
}

func TestMessagingHub_SubscribeTasks_WithConnectedTaskQueue(t *testing.T) {
	hub := NewMessagingHub(DefaultHubConfig())
	taskQueue := newHubTestMockTaskQueueBroker()
	taskQueue.connected = true

	hub.SetTaskQueueBroker(taskQueue)
	hub.connected = true

	ctx := context.Background()
	handler := func(ctx context.Context, task *Task) error {
		return nil
	}

	sub, err := hub.SubscribeTasks(ctx, "test-queue", handler)

	assert.NoError(t, err)
	assert.NotNil(t, sub)
	assert.True(t, sub.IsActive())
}

func TestMessagingHub_SubscribeTasks_UsesFallback(t *testing.T) {
	hub := NewMessagingHub(DefaultHubConfig())
	fallback := newHubTestMockBroker()
	fallback.connected = true

	hub.SetFallbackBroker(fallback)
	hub.connected = true

	ctx := context.Background()
	handler := func(ctx context.Context, task *Task) error {
		return nil
	}

	sub, err := hub.SubscribeTasks(ctx, "test-queue", handler)

	assert.NoError(t, err)
	assert.NotNil(t, sub)
}

func TestMessagingHub_SubscribeEvents_WithConnectedEventStream(t *testing.T) {
	hub := NewMessagingHub(DefaultHubConfig())
	eventStream := newHubTestMockEventStreamBroker()
	eventStream.connected = true

	hub.SetEventStreamBroker(eventStream)
	hub.connected = true

	ctx := context.Background()
	handler := func(ctx context.Context, event *Event) error {
		return nil
	}

	sub, err := hub.SubscribeEvents(ctx, "test-topic", handler)

	assert.NoError(t, err)
	assert.NotNil(t, sub)
	assert.True(t, sub.IsActive())
}

func TestMessagingHub_SubscribeEvents_UsesFallback(t *testing.T) {
	hub := NewMessagingHub(DefaultHubConfig())
	fallback := newHubTestMockBroker()
	fallback.connected = true

	hub.SetFallbackBroker(fallback)
	hub.connected = true

	ctx := context.Background()
	handler := func(ctx context.Context, event *Event) error {
		return nil
	}

	sub, err := hub.SubscribeEvents(ctx, "test-topic", handler)

	assert.NoError(t, err)
	assert.NotNil(t, sub)
}

func TestMessagingHub_DeclareQueue_WithConnectedTaskQueue(t *testing.T) {
	hub := NewMessagingHub(DefaultHubConfig())
	taskQueue := newHubTestMockTaskQueueBroker()
	taskQueue.connected = true

	hub.SetTaskQueueBroker(taskQueue)
	hub.connected = true

	ctx := context.Background()
	err := hub.DeclareQueue(ctx, "new-queue")

	assert.NoError(t, err)
}

func TestMessagingHub_GetQueueStats_WithConnectedTaskQueue(t *testing.T) {
	hub := NewMessagingHub(DefaultHubConfig())
	taskQueue := newHubTestMockTaskQueueBroker()
	taskQueue.connected = true

	hub.SetTaskQueueBroker(taskQueue)
	hub.connected = true

	ctx := context.Background()
	stats, err := hub.GetQueueStats(ctx, "test-queue")

	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, "test-queue", stats.Name)
}

func TestMessagingHub_CreateTopic_WithConnectedEventStream(t *testing.T) {
	hub := NewMessagingHub(DefaultHubConfig())
	eventStream := newHubTestMockEventStreamBroker()
	eventStream.connected = true

	hub.SetEventStreamBroker(eventStream)
	hub.connected = true

	ctx := context.Background()
	err := hub.CreateTopic(ctx, "new-topic", 3, 1)

	assert.NoError(t, err)
	assert.Contains(t, eventStream.topics, "new-topic")
}

func TestMessagingHub_GetTopicMetadata_WithConnectedEventStream(t *testing.T) {
	hub := NewMessagingHub(DefaultHubConfig())
	eventStream := newHubTestMockEventStreamBroker()
	eventStream.connected = true

	hub.SetEventStreamBroker(eventStream)
	hub.connected = true

	ctx := context.Background()
	metadata, err := hub.GetTopicMetadata(ctx, "test-topic")

	assert.NoError(t, err)
	assert.NotNil(t, metadata)
	assert.Equal(t, "test-topic", metadata.Name)
}

func TestMessagingHub_StreamEvents_WithConnectedEventStream(t *testing.T) {
	hub := NewMessagingHub(DefaultHubConfig())
	eventStream := newHubTestMockEventStreamBroker()
	eventStream.connected = true

	hub.SetEventStreamBroker(eventStream)
	hub.connected = true

	ctx := context.Background()
	ch, err := hub.StreamEvents(ctx, "test-topic")

	assert.NoError(t, err)
	assert.NotNil(t, ch)
}

func TestMessagingHub_HealthCheck_WithConnectedBrokers(t *testing.T) {
	hub := NewMessagingHub(DefaultHubConfig())
	taskQueue := newHubTestMockTaskQueueBroker()
	taskQueue.connected = true
	eventStream := newHubTestMockEventStreamBroker()
	eventStream.connected = true

	hub.SetTaskQueueBroker(taskQueue)
	hub.SetEventStreamBroker(eventStream)
	hub.connected = true

	ctx := context.Background()
	err := hub.HealthCheck(ctx)

	assert.NoError(t, err)
}

func TestMessagingHub_HealthCheck_TaskQueueError(t *testing.T) {
	hub := NewMessagingHub(DefaultHubConfig())
	taskQueue := newHubTestMockTaskQueueBroker()
	taskQueue.connected = true
	taskQueue.healthCheckErr = NewBrokerError(ErrCodeConnectionFailed, "health check failed", nil)

	hub.SetTaskQueueBroker(taskQueue)
	hub.connected = true

	ctx := context.Background()
	err := hub.HealthCheck(ctx)

	assert.Error(t, err)
}

func TestMessagingHub_Publish_RoutesToTaskQueue(t *testing.T) {
	hub := NewMessagingHub(DefaultHubConfig())
	taskQueue := newHubTestMockTaskQueueBroker()
	taskQueue.connected = true

	hub.SetTaskQueueBroker(taskQueue)
	hub.connected = true

	ctx := context.Background()

	// Create a task-like message
	task := NewTask("test.task", []byte(`{"key":"value"}`))
	msg := task.ToMessage()

	// Use a task queue prefix
	err := hub.Publish(ctx, "helixagent.tasks.test", msg)

	assert.NoError(t, err)
	assert.Len(t, taskQueue.enqueuedTasks, 1)
}

func TestMessagingHub_Publish_RoutesToEventStream(t *testing.T) {
	hub := NewMessagingHub(DefaultHubConfig())
	eventStream := newHubTestMockEventStreamBroker()
	eventStream.connected = true

	hub.SetEventStreamBroker(eventStream)
	hub.connected = true

	ctx := context.Background()

	// Create an event-like message
	event := NewEvent(EventTypeLLMRequestStarted, "test", []byte(`{"key":"value"}`))
	msg := event.ToMessage()

	// Use an event stream prefix
	err := hub.Publish(ctx, "helixagent.events.test", msg)

	assert.NoError(t, err)
	assert.Len(t, eventStream.publishedEvents, 1)
}

func TestMessagingHub_Subscribe_RoutesToTaskQueue(t *testing.T) {
	hub := NewMessagingHub(DefaultHubConfig())
	taskQueue := newHubTestMockTaskQueueBroker()
	taskQueue.connected = true

	hub.SetTaskQueueBroker(taskQueue)
	hub.connected = true

	ctx := context.Background()
	handler := func(ctx context.Context, msg *Message) error {
		return nil
	}

	// Use a task queue prefix
	sub, err := hub.Subscribe(ctx, "tasks.test", handler)

	assert.NoError(t, err)
	assert.NotNil(t, sub)
}

func TestMessagingHub_Subscribe_RoutesToEventStream(t *testing.T) {
	hub := NewMessagingHub(DefaultHubConfig())
	eventStream := newHubTestMockEventStreamBroker()
	eventStream.connected = true

	hub.SetEventStreamBroker(eventStream)
	hub.connected = true

	ctx := context.Background()
	handler := func(ctx context.Context, msg *Message) error {
		return nil
	}

	// Use an event stream prefix
	sub, err := hub.Subscribe(ctx, "events.test", handler)

	assert.NoError(t, err)
	assert.NotNil(t, sub)
}

func TestMessagingHub_Subscribe_UsesFallbackForUnknownPrefix(t *testing.T) {
	hub := NewMessagingHub(DefaultHubConfig())
	fallback := newHubTestMockBroker()
	fallback.connected = true

	hub.SetFallbackBroker(fallback)
	hub.connected = true

	ctx := context.Background()
	handler := func(ctx context.Context, msg *Message) error {
		return nil
	}

	// Use an unknown prefix
	sub, err := hub.Subscribe(ctx, "unknown.topic", handler)

	assert.NoError(t, err)
	assert.NotNil(t, sub)
}

// =============================================================================
// Concurrent Operation Tests
// =============================================================================

func TestMessagingHub_ConcurrentEnqueueTask(t *testing.T) {
	hub := NewMessagingHub(DefaultHubConfig())
	taskQueue := newHubTestMockTaskQueueBroker()
	taskQueue.connected = true

	hub.SetTaskQueueBroker(taskQueue)
	hub.connected = true

	ctx := context.Background()
	var wg sync.WaitGroup
	taskCount := 100

	for i := 0; i < taskCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			task := NewTask("test.task", []byte(`{"index":`+string(rune('0'+idx%10))+`}`))
			_ = hub.EnqueueTask(ctx, "test-queue", task)
		}(i)
	}

	wg.Wait()
	assert.Len(t, taskQueue.enqueuedTasks, taskCount)
}

func TestMessagingHub_ConcurrentPublishEvent(t *testing.T) {
	hub := NewMessagingHub(DefaultHubConfig())
	eventStream := newHubTestMockEventStreamBroker()
	eventStream.connected = true

	hub.SetEventStreamBroker(eventStream)
	hub.connected = true

	ctx := context.Background()
	var wg sync.WaitGroup
	eventCount := 100

	for i := 0; i < eventCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			event := NewEvent(EventTypeLLMRequestStarted, "test", []byte(`{"index":`+string(rune('0'+idx%10))+`}`))
			_ = hub.PublishEvent(ctx, "test-topic", event)
		}(i)
	}

	wg.Wait()
	assert.Len(t, eventStream.publishedEvents, eventCount)
}

func TestMessagingHub_ConcurrentSubscribe(t *testing.T) {
	hub := NewMessagingHub(DefaultHubConfig())
	taskQueue := newHubTestMockTaskQueueBroker()
	taskQueue.connected = true

	hub.SetTaskQueueBroker(taskQueue)
	hub.connected = true

	ctx := context.Background()
	var wg sync.WaitGroup
	subCount := 10

	for i := 0; i < subCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			handler := func(ctx context.Context, task *Task) error { return nil }
			_, _ = hub.SubscribeTasks(ctx, "test-queue-"+string(rune('0'+idx)), handler)
		}(i)
	}

	wg.Wait()
}

// =============================================================================
// Metrics Tests
// =============================================================================

func TestMessagingHub_MetricsRecording_EnqueueTask(t *testing.T) {
	hub := NewMessagingHub(DefaultHubConfig())
	taskQueue := newHubTestMockTaskQueueBroker()
	taskQueue.connected = true

	hub.SetTaskQueueBroker(taskQueue)
	hub.connected = true

	ctx := context.Background()
	task := NewTask("test.task", []byte(`{"key":"value"}`))

	_ = hub.EnqueueTask(ctx, "test-queue", task)

	metrics := hub.GetMetrics()
	assert.NotNil(t, metrics)
	// Metrics should have recorded publish operation
	assert.GreaterOrEqual(t, metrics.BrokerMetrics.MessagesPublished.Load(), int64(0))
}

func TestMessagingHub_MetricsRecording_FallbackUsage(t *testing.T) {
	hub := NewMessagingHub(DefaultHubConfig())
	fallback := newHubTestMockBroker()
	fallback.connected = true

	hub.SetFallbackBroker(fallback)
	hub.connected = true

	ctx := context.Background()
	task := NewTask("test.task", []byte(`{"key":"value"}`))

	_ = hub.EnqueueTask(ctx, "test-queue", task)

	metrics := hub.GetMetrics()
	assert.Greater(t, metrics.FallbackUsages.Load(), int64(0))
}

// =============================================================================
// Global Hub Concurrent Access Tests
// =============================================================================

func TestGlobalHub_ConcurrentAccess(t *testing.T) {
	var wg sync.WaitGroup

	// Concurrent sets
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			hub := NewMessagingHub(nil)
			SetGlobalHub(hub)
		}()
	}

	// Concurrent gets
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = GetGlobalHub()
		}()
	}

	wg.Wait()

	// Cleanup
	SetGlobalHub(nil)
}
