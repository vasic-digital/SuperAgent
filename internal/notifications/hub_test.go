package notifications

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/models"
)

// MockSubscriber implements the Subscriber interface for testing
type MockSubscriber struct {
	id           string
	notifyType   NotificationType
	active       bool
	notifications []*TaskNotification
	mu           sync.Mutex
	notifyErr    error
	closed       bool
}

func NewMockSubscriber(id string, notifyType NotificationType) *MockSubscriber {
	return &MockSubscriber{
		id:            id,
		notifyType:    notifyType,
		active:        true,
		notifications: make([]*TaskNotification, 0),
	}
}

func (m *MockSubscriber) Notify(ctx context.Context, notification *TaskNotification) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.notifyErr != nil {
		return m.notifyErr
	}
	m.notifications = append(m.notifications, notification)
	return nil
}

func (m *MockSubscriber) Type() NotificationType {
	return m.notifyType
}

func (m *MockSubscriber) ID() string {
	return m.id
}

func (m *MockSubscriber) IsActive() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.active
}

func (m *MockSubscriber) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.active = false
	m.closed = true
	return nil
}

func (m *MockSubscriber) GetNotifications() []*TaskNotification {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]*TaskNotification{}, m.notifications...)
}

func (m *MockSubscriber) SetActive(active bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.active = active
}

func (m *MockSubscriber) IsClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closed
}

// Test helper to create a test logger
func testLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Suppress logs during tests
	return logger
}

// Test helper to create a test task
func testTask(id string) *models.BackgroundTask {
	return &models.BackgroundTask{
		ID:       id,
		TaskType: "test-type",
		TaskName: "test-task",
		Status:   models.TaskStatusRunning,
		Progress: 50.0,
	}
}

// Tests for DefaultHubConfig
func TestDefaultHubConfig(t *testing.T) {
	config := DefaultHubConfig()

	assert.Equal(t, 1000, config.EventBufferSize)
	assert.Equal(t, 5, config.WorkerCount)
	assert.Equal(t, 30*time.Second, config.NotificationTimeout)
	assert.True(t, config.RetryEnabled)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, time.Second, config.RetryBackoff)
}

// Tests for NewNotificationHub
func TestNewNotificationHub(t *testing.T) {
	logger := testLogger()

	t.Run("with default config", func(t *testing.T) {
		hub := NewNotificationHub(nil, nil, nil, nil, nil, logger)
		require.NotNil(t, hub)

		assert.NotNil(t, hub.subscribers)
		assert.NotNil(t, hub.globalSubs)
		assert.NotNil(t, hub.eventChan)
		assert.Equal(t, logger, hub.logger)

		err := hub.Stop()
		assert.NoError(t, err)
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &HubConfig{
			EventBufferSize:     500,
			WorkerCount:         3,
			NotificationTimeout: 10 * time.Second,
			RetryEnabled:        false,
			MaxRetries:          1,
			RetryBackoff:        500 * time.Millisecond,
		}

		hub := NewNotificationHub(config, nil, nil, nil, nil, logger)
		require.NotNil(t, hub)

		err := hub.Stop()
		assert.NoError(t, err)
	})

	t.Run("with all components", func(t *testing.T) {
		sseManager := NewSSEManager(nil, logger)
		wsServer := NewWebSocketServer(nil, logger)
		webhookDispatcher := NewWebhookDispatcher(nil, logger)
		pollingStore := NewPollingStore(nil, logger)

		hub := NewNotificationHub(nil, sseManager, wsServer, webhookDispatcher, pollingStore, logger)
		require.NotNil(t, hub)

		assert.Equal(t, sseManager, hub.sseManager)
		assert.Equal(t, wsServer, hub.wsServer)
		assert.Equal(t, webhookDispatcher, hub.webhookDispatcher)
		assert.Equal(t, pollingStore, hub.pollingStore)

		// Cleanup
		hub.Stop()
		sseManager.Stop()
		wsServer.Stop()
		webhookDispatcher.Stop()
		pollingStore.Stop()
	})
}

// Tests for Hub Start/Stop
func TestNotificationHub_StartStop(t *testing.T) {
	logger := testLogger()
	hub := NewNotificationHub(nil, nil, nil, nil, nil, logger)

	err := hub.Start()
	assert.NoError(t, err)

	err = hub.Stop()
	assert.NoError(t, err)
}

// Tests for Subscribe/Unsubscribe
func TestNotificationHub_Subscribe(t *testing.T) {
	logger := testLogger()
	hub := NewNotificationHub(nil, nil, nil, nil, nil, logger)
	defer hub.Stop()

	t.Run("subscribe single subscriber", func(t *testing.T) {
		subscriber := NewMockSubscriber("sub-1", NotificationTypeSSE)
		hub.Subscribe("task-1", subscriber)

		count := hub.GetActiveSubscribers("task-1")
		assert.Equal(t, 1, count)
	})

	t.Run("subscribe multiple subscribers", func(t *testing.T) {
		subscriber2 := NewMockSubscriber("sub-2", NotificationTypeWebSocket)
		subscriber3 := NewMockSubscriber("sub-3", NotificationTypeWebhook)

		hub.Subscribe("task-1", subscriber2)
		hub.Subscribe("task-1", subscriber3)

		count := hub.GetActiveSubscribers("task-1")
		assert.Equal(t, 3, count)
	})

	t.Run("unsubscribe", func(t *testing.T) {
		hub.Unsubscribe("task-1", "sub-2")

		count := hub.GetActiveSubscribers("task-1")
		assert.Equal(t, 2, count)
	})
}

// Tests for SubscribeGlobal/UnsubscribeGlobal
func TestNotificationHub_GlobalSubscribe(t *testing.T) {
	logger := testLogger()
	hub := NewNotificationHub(nil, nil, nil, nil, nil, logger)
	defer hub.Stop()

	subscriber1 := NewMockSubscriber("global-1", NotificationTypeSSE)
	subscriber2 := NewMockSubscriber("global-2", NotificationTypeWebSocket)

	hub.SubscribeGlobal(subscriber1)
	hub.SubscribeGlobal(subscriber2)

	hub.UnsubscribeGlobal("global-1")
	assert.True(t, subscriber1.IsClosed())
}

// Tests for NotifyTaskEvent
func TestNotificationHub_NotifyTaskEvent(t *testing.T) {
	logger := testLogger()
	pollingStore := NewPollingStore(nil, logger)
	defer pollingStore.Stop()

	hub := NewNotificationHub(nil, nil, nil, nil, pollingStore, logger)
	defer hub.Stop()

	task := testTask("task-123")
	data := map[string]interface{}{"key": "value"}

	err := hub.NotifyTaskEvent(context.Background(), task, "progress", data)
	assert.NoError(t, err)

	// Give time for event processing
	time.Sleep(100 * time.Millisecond)

	// Check event was stored in polling store
	events := pollingStore.GetTaskEvents("task-123", nil, 10)
	assert.Len(t, events, 1)
	assert.Equal(t, "task-123", events[0].TaskID)
	assert.Equal(t, "progress", events[0].EventType)
}

// Tests for BroadcastToTask
func TestNotificationHub_BroadcastToTask(t *testing.T) {
	logger := testLogger()
	sseManager := NewSSEManager(nil, logger)
	defer sseManager.Stop()

	hub := NewNotificationHub(nil, sseManager, nil, nil, nil, logger)
	defer hub.Stop()

	// Register a client
	clientChan := make(chan []byte, 10)
	sseManager.RegisterClient("task-1", clientChan)

	// Broadcast message
	message := []byte(`{"type":"test"}`)
	err := hub.BroadcastToTask(context.Background(), "task-1", message)
	assert.NoError(t, err)

	// Verify client received message
	select {
	case received := <-clientChan:
		assert.Contains(t, string(received), "test")
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for message")
	}

	sseManager.UnregisterClient("task-1", clientChan)
}

// Tests for RegisterSSEClient
func TestNotificationHub_RegisterSSEClient(t *testing.T) {
	logger := testLogger()
	sseManager := NewSSEManager(nil, logger)
	defer sseManager.Stop()

	hub := NewNotificationHub(nil, sseManager, nil, nil, nil, logger)
	defer hub.Stop()

	clientChan := make(chan []byte, 10)

	err := hub.RegisterSSEClient(context.Background(), "task-1", clientChan)
	assert.NoError(t, err)

	count := sseManager.GetClientCount("task-1")
	assert.Equal(t, 1, count)

	err = hub.UnregisterSSEClient(context.Background(), "task-1", clientChan)
	assert.NoError(t, err)

	count = sseManager.GetClientCount("task-1")
	assert.Equal(t, 0, count)
}

// Tests for RegisterSSEClient with nil sseManager
func TestNotificationHub_RegisterSSEClient_NilManager(t *testing.T) {
	logger := testLogger()
	hub := NewNotificationHub(nil, nil, nil, nil, nil, logger)
	defer hub.Stop()

	clientChan := make(chan []byte, 10)

	err := hub.RegisterSSEClient(context.Background(), "task-1", clientChan)
	assert.NoError(t, err)

	err = hub.UnregisterSSEClient(context.Background(), "task-1", clientChan)
	assert.NoError(t, err)
}

// Tests for GetActiveSubscribers
func TestNotificationHub_GetActiveSubscribers(t *testing.T) {
	logger := testLogger()
	hub := NewNotificationHub(nil, nil, nil, nil, nil, logger)
	defer hub.Stop()

	// No subscribers
	count := hub.GetActiveSubscribers("nonexistent")
	assert.Equal(t, 0, count)

	// Add active subscriber
	subscriber1 := NewMockSubscriber("sub-1", NotificationTypeSSE)
	hub.Subscribe("task-1", subscriber1)

	count = hub.GetActiveSubscribers("task-1")
	assert.Equal(t, 1, count)

	// Add inactive subscriber
	subscriber2 := NewMockSubscriber("sub-2", NotificationTypeWebSocket)
	subscriber2.SetActive(false)
	hub.Subscribe("task-1", subscriber2)

	count = hub.GetActiveSubscribers("task-1")
	assert.Equal(t, 1, count) // Only active subscriber counted
}

// Tests for CleanupInactiveSubscribers
func TestNotificationHub_CleanupInactiveSubscribers(t *testing.T) {
	logger := testLogger()
	hub := NewNotificationHub(nil, nil, nil, nil, nil, logger)
	defer hub.Stop()

	// Add subscribers
	activeSub := NewMockSubscriber("active", NotificationTypeSSE)
	inactiveSub := NewMockSubscriber("inactive", NotificationTypeWebSocket)
	inactiveSub.SetActive(false)

	hub.Subscribe("task-1", activeSub)
	hub.Subscribe("task-1", inactiveSub)

	// Add global subscribers
	activeGlobal := NewMockSubscriber("global-active", NotificationTypeSSE)
	inactiveGlobal := NewMockSubscriber("global-inactive", NotificationTypeWebSocket)
	inactiveGlobal.SetActive(false)

	hub.SubscribeGlobal(activeGlobal)
	hub.SubscribeGlobal(inactiveGlobal)

	// Cleanup
	hub.CleanupInactiveSubscribers()

	// Verify only active subscribers remain
	count := hub.GetActiveSubscribers("task-1")
	assert.Equal(t, 1, count)

	// Verify inactive subscribers were closed
	assert.True(t, inactiveSub.IsClosed())
	assert.True(t, inactiveGlobal.IsClosed())
}

// Tests for TaskNotification
func TestTaskNotification(t *testing.T) {
	task := testTask("task-1")
	notification := &TaskNotification{
		TaskID:    task.ID,
		EventType: "progress",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"progress": 50},
		Task:      task,
	}

	assert.Equal(t, "task-1", notification.TaskID)
	assert.Equal(t, "progress", notification.EventType)
	assert.NotNil(t, notification.Data)
	assert.NotNil(t, notification.Task)
}

// Tests for concurrent operations
func TestNotificationHub_ConcurrentOperations(t *testing.T) {
	logger := testLogger()
	hub := NewNotificationHub(nil, nil, nil, nil, nil, logger)
	defer hub.Stop()

	var wg sync.WaitGroup
	numGoroutines := 10
	numOperations := 100

	// Concurrent subscribe/unsubscribe
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				sub := NewMockSubscriber("sub-"+string(rune(id))+"-"+string(rune(j)), NotificationTypeSSE)
				hub.Subscribe("task-1", sub)
			}
		}(i)
	}

	wg.Wait()

	// Verify no panic occurred and state is consistent
	count := hub.GetActiveSubscribers("task-1")
	assert.Greater(t, count, 0)
}

// Tests for event channel overflow
func TestNotificationHub_EventChannelOverflow(t *testing.T) {
	logger := testLogger()

	// Create hub with small buffer
	config := &HubConfig{
		EventBufferSize: 1,
		WorkerCount:     1,
		NotificationTimeout: time.Second,
	}

	pollingStore := NewPollingStore(nil, logger)
	defer pollingStore.Stop()

	hub := NewNotificationHub(config, nil, nil, nil, pollingStore, logger)

	// Pause event processing by filling the channel
	task := testTask("task-1")

	// Send many events quickly
	for i := 0; i < 100; i++ {
		hub.NotifyTaskEvent(context.Background(), task, "event", nil)
	}

	// Should not panic or block
	hub.Stop()
}

// Tests for notification dispatch to subscribers
func TestNotificationHub_DispatchToSubscribers(t *testing.T) {
	logger := testLogger()
	hub := NewNotificationHub(nil, nil, nil, nil, nil, logger)

	// Create and subscribe mock subscriber
	subscriber := NewMockSubscriber("test-sub", NotificationTypeSSE)
	hub.Subscribe("task-1", subscriber)

	task := testTask("task-1")

	// Send notification
	err := hub.NotifyTaskEvent(context.Background(), task, "progress", nil)
	assert.NoError(t, err)

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	hub.Stop()

	// Verify subscriber received notification
	notifications := subscriber.GetNotifications()
	assert.GreaterOrEqual(t, len(notifications), 1)
}

// Tests for notification dispatch to global subscribers
func TestNotificationHub_DispatchToGlobalSubscribers(t *testing.T) {
	logger := testLogger()
	hub := NewNotificationHub(nil, nil, nil, nil, nil, logger)

	// Create and subscribe global mock subscriber
	globalSub := NewMockSubscriber("global-sub", NotificationTypeSSE)
	hub.SubscribeGlobal(globalSub)

	task := testTask("task-1")

	// Send notification
	err := hub.NotifyTaskEvent(context.Background(), task, "progress", nil)
	assert.NoError(t, err)

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	hub.Stop()

	// Verify global subscriber received notification
	notifications := globalSub.GetNotifications()
	assert.GreaterOrEqual(t, len(notifications), 1)
}

// Tests for NotificationType constants
func TestNotificationType(t *testing.T) {
	assert.Equal(t, NotificationType("sse"), NotificationTypeSSE)
	assert.Equal(t, NotificationType("websocket"), NotificationTypeWebSocket)
	assert.Equal(t, NotificationType("webhook"), NotificationTypeWebhook)
	assert.Equal(t, NotificationType("polling"), NotificationTypePolling)
}

// Tests for TaskNotification JSON marshaling
func TestTaskNotification_JSON(t *testing.T) {
	task := testTask("task-1")
	notification := &TaskNotification{
		TaskID:    "task-1",
		EventType: "progress",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"progress": 50.5},
		Task:      task,
	}

	data, err := json.Marshal(notification)
	require.NoError(t, err)

	var unmarshaled TaskNotification
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, notification.TaskID, unmarshaled.TaskID)
	assert.Equal(t, notification.EventType, unmarshaled.EventType)
}

// Tests for HubConfig validation
func TestHubConfig(t *testing.T) {
	config := &HubConfig{
		EventBufferSize:     100,
		WorkerCount:         2,
		NotificationTimeout: 5 * time.Second,
		RetryEnabled:        true,
		MaxRetries:          5,
		RetryBackoff:        2 * time.Second,
	}

	assert.Equal(t, 100, config.EventBufferSize)
	assert.Equal(t, 2, config.WorkerCount)
	assert.Equal(t, 5*time.Second, config.NotificationTimeout)
	assert.True(t, config.RetryEnabled)
	assert.Equal(t, 5, config.MaxRetries)
	assert.Equal(t, 2*time.Second, config.RetryBackoff)
}
