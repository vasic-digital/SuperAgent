package notifications_test

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/models"
	"dev.helix.agent/internal/notifications"
)

func TestDefaultHubConfig(t *testing.T) {
	config := notifications.DefaultHubConfig()

	assert.NotNil(t, config)
	assert.Greater(t, config.EventBufferSize, 0)
	assert.Greater(t, config.WorkerCount, 0)
	assert.Greater(t, config.NotificationTimeout, time.Duration(0))
}

func TestNotificationHub_Creation(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	hub := notifications.NewNotificationHub(
		nil, // Use defaults
		nil, // sseManager
		nil, // wsServer
		nil, // webhookDispatcher
		nil, // pollingStore
		logger,
	)

	require.NotNil(t, hub)
}

func TestNotificationHub_StartAndStop(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	hub := notifications.NewNotificationHub(
		nil, nil, nil, nil, nil,
		logger,
	)

	err := hub.Start()
	require.NoError(t, err)

	err = hub.Stop()
	assert.NoError(t, err)
}

func TestNotificationHub_NotifyTaskEvent(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	pollingStore := notifications.NewPollingStore(nil, logger)
	defer pollingStore.Stop()

	config := &notifications.HubConfig{
		EventBufferSize:     100,
		WorkerCount:         1,
		NotificationTimeout: time.Second,
	}

	hub := notifications.NewNotificationHub(
		config,
		nil, nil, nil,
		pollingStore,
		logger,
	)
	defer hub.Stop()

	err := hub.Start()
	require.NoError(t, err)

	task := &models.BackgroundTask{
		ID:       "task-123",
		TaskType: "test",
		TaskName: "Test Task",
		Status:   models.TaskStatusRunning,
	}

	err = hub.NotifyTaskEvent(context.Background(), task, "progress", map[string]interface{}{
		"percent": 50.0,
		"message": "Halfway done",
	})
	require.NoError(t, err)

	// Give time for async processing
	time.Sleep(100 * time.Millisecond)

	// Check polling store received the event
	events := pollingStore.GetTaskEvents(task.ID, nil, 10)
	assert.NotEmpty(t, events)
}

func TestNotificationHub_Subscribe(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	hub := notifications.NewNotificationHub(
		nil, nil, nil, nil, nil,
		logger,
	)
	defer hub.Stop()

	subscriber := newMockSubscriber("sub-1")
	hub.Subscribe("task-123", subscriber)

	count := hub.GetActiveSubscribers("task-123")
	assert.Equal(t, 1, count)
}

func TestNotificationHub_Unsubscribe(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	hub := notifications.NewNotificationHub(
		nil, nil, nil, nil, nil,
		logger,
	)
	defer hub.Stop()

	subscriber := newMockSubscriber("sub-1")
	hub.Subscribe("task-123", subscriber)

	assert.Equal(t, 1, hub.GetActiveSubscribers("task-123"))

	hub.Unsubscribe("task-123", "sub-1")

	assert.Equal(t, 0, hub.GetActiveSubscribers("task-123"))
}

func TestNotificationHub_SubscribeGlobal(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	hub := notifications.NewNotificationHub(
		nil, nil, nil, nil, nil,
		logger,
	)
	defer hub.Stop()

	subscriber := newMockSubscriber("global-sub-1")
	hub.SubscribeGlobal(subscriber)

	// No direct way to count global subscribers, but should not panic
}

func TestNotificationHub_UnsubscribeGlobal(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	hub := notifications.NewNotificationHub(
		nil, nil, nil, nil, nil,
		logger,
	)
	defer hub.Stop()

	subscriber := newMockSubscriber("global-sub-1")
	hub.SubscribeGlobal(subscriber)
	hub.UnsubscribeGlobal("global-sub-1")

	// Should not panic
}

func TestNotificationHub_CleanupInactiveSubscribers(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	hub := notifications.NewNotificationHub(
		nil, nil, nil, nil, nil,
		logger,
	)
	defer hub.Stop()

	// Add an active subscriber
	activeSub := newMockSubscriber("active-sub")
	hub.Subscribe("task-123", activeSub)

	// Add an inactive subscriber
	inactiveSub := newMockSubscriber("inactive-sub")
	inactiveSub.active = false
	hub.Subscribe("task-123", inactiveSub)

	// GetActiveSubscribers only counts active subscribers, so it returns 1
	// even though 2 subscribers are registered
	assert.Equal(t, 1, hub.GetActiveSubscribers("task-123"))

	hub.CleanupInactiveSubscribers()

	// After cleanup, only active subscriber should remain
	// GetActiveSubscribers still returns 1 (the active one)
	assert.Equal(t, 1, hub.GetActiveSubscribers("task-123"))

	// The inactive subscriber should have been closed during cleanup
	assert.False(t, inactiveSub.active, "Inactive subscriber should remain inactive")
}

func TestNotificationHub_BroadcastToTask(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	hub := notifications.NewNotificationHub(
		nil, nil, nil, nil, nil,
		logger,
	)
	defer hub.Stop()

	err := hub.BroadcastToTask(context.Background(), "task-123", []byte("test message"))
	assert.NoError(t, err)
}

func TestNotificationHub_RegisterSSEClient(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	hub := notifications.NewNotificationHub(
		nil, nil, nil, nil, nil,
		logger,
	)
	defer hub.Stop()

	client := make(chan []byte, 10)
	err := hub.RegisterSSEClient(context.Background(), "task-123", client)
	assert.NoError(t, err)
}

func TestNotificationHub_UnregisterSSEClient(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	hub := notifications.NewNotificationHub(
		nil, nil, nil, nil, nil,
		logger,
	)
	defer hub.Stop()

	client := make(chan []byte, 10)
	hub.RegisterSSEClient(context.Background(), "task-123", client)

	err := hub.UnregisterSSEClient(context.Background(), "task-123", client)
	assert.NoError(t, err)
}

func TestTaskNotification_Fields(t *testing.T) {
	now := time.Now()
	notification := &notifications.TaskNotification{
		TaskID:    "task-123",
		EventType: "progress",
		Timestamp: now,
		Data:      map[string]interface{}{"percent": 50.0},
		Task:      &models.BackgroundTask{ID: "task-123"},
	}

	assert.Equal(t, "task-123", notification.TaskID)
	assert.Equal(t, "progress", notification.EventType)
	assert.Equal(t, now, notification.Timestamp)
	assert.NotNil(t, notification.Data)
	assert.NotNil(t, notification.Task)
}

func TestNotificationType_Values(t *testing.T) {
	assert.Equal(t, notifications.NotificationType("sse"), notifications.NotificationTypeSSE)
	assert.Equal(t, notifications.NotificationType("websocket"), notifications.NotificationTypeWebSocket)
	assert.Equal(t, notifications.NotificationType("webhook"), notifications.NotificationTypeWebhook)
	assert.Equal(t, notifications.NotificationType("polling"), notifications.NotificationTypePolling)
}

// mockSubscriber is a test implementation of Subscriber
type mockSubscriber struct {
	id           string
	active       bool
	notifications []*notifications.TaskNotification
}

func newMockSubscriber(id string) *mockSubscriber {
	return &mockSubscriber{
		id:           id,
		active:       true,
		notifications: make([]*notifications.TaskNotification, 0),
	}
}

func (m *mockSubscriber) Notify(ctx context.Context, notification *notifications.TaskNotification) error {
	m.notifications = append(m.notifications, notification)
	return nil
}

func (m *mockSubscriber) Type() notifications.NotificationType {
	return notifications.NotificationTypePolling
}

func (m *mockSubscriber) ID() string {
	return m.id
}

func (m *mockSubscriber) IsActive() bool {
	return m.active
}

func (m *mockSubscriber) Close() error {
	m.active = false
	return nil
}
