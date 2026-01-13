package notifications_test

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/notifications"
)

func TestDefaultPollingConfig(t *testing.T) {
	config := notifications.DefaultPollingConfig()

	assert.NotNil(t, config)
	assert.Greater(t, config.MaxEventsPerTask, 0)
	assert.Greater(t, config.MaxGlobalEvents, 0)
	assert.Greater(t, config.EventTTL, time.Duration(0))
	assert.Greater(t, config.CleanupInterval, time.Duration(0))
}

func TestPollingStore_Creation(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	store := notifications.NewPollingStore(nil, logger)
	require.NotNil(t, store)

	store.Stop()
}

func TestPollingStore_StartAndStop(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	store := notifications.NewPollingStore(nil, logger)

	err := store.Start()
	require.NoError(t, err)

	err = store.Stop()
	assert.NoError(t, err)
}

func TestPollingStore_StoreEvent(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	store := notifications.NewPollingStore(nil, logger)
	defer store.Stop()

	notification := &notifications.TaskNotification{
		TaskID:    "task-123",
		EventType: "progress",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"percent": 50.0},
	}

	store.StoreEvent(notification)

	assert.Equal(t, 1, store.GetEventCount("task-123"))
	assert.Equal(t, 1, store.GetGlobalEventCount())
}

func TestPollingStore_GetTaskEvents(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	store := notifications.NewPollingStore(nil, logger)
	defer store.Stop()

	// Store multiple events
	for i := 0; i < 5; i++ {
		store.StoreEvent(&notifications.TaskNotification{
			TaskID:    "task-123",
			EventType: "progress",
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"percent": float64(i * 20)},
		})
	}

	events := store.GetTaskEvents("task-123", nil, 10)
	assert.Len(t, events, 5)
}

func TestPollingStore_GetTaskEvents_WithLimit(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	store := notifications.NewPollingStore(nil, logger)
	defer store.Stop()

	// Store multiple events
	for i := 0; i < 10; i++ {
		store.StoreEvent(&notifications.TaskNotification{
			TaskID:    "task-123",
			EventType: "progress",
			Timestamp: time.Now(),
		})
	}

	events := store.GetTaskEvents("task-123", nil, 5)
	assert.Len(t, events, 5)
}

func TestPollingStore_GetTaskEvents_WithSince(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	store := notifications.NewPollingStore(nil, logger)
	defer store.Stop()

	// Store events with different timestamps
	oldTime := time.Now().Add(-time.Hour)
	store.StoreEvent(&notifications.TaskNotification{
		TaskID:    "task-123",
		EventType: "progress",
		Timestamp: oldTime,
	})

	newTime := time.Now()
	store.StoreEvent(&notifications.TaskNotification{
		TaskID:    "task-123",
		EventType: "progress",
		Timestamp: newTime,
	})

	since := time.Now().Add(-30 * time.Minute)
	events := store.GetTaskEvents("task-123", &since, 10)
	assert.Len(t, events, 1)
}

func TestPollingStore_GetGlobalEvents(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	store := notifications.NewPollingStore(nil, logger)
	defer store.Stop()

	// Store events from different tasks
	store.StoreEvent(&notifications.TaskNotification{
		TaskID:    "task-1",
		EventType: "progress",
		Timestamp: time.Now(),
	})
	store.StoreEvent(&notifications.TaskNotification{
		TaskID:    "task-2",
		EventType: "progress",
		Timestamp: time.Now(),
	})

	events := store.GetGlobalEvents(nil, 10)
	assert.Len(t, events, 2)
}

func TestPollingStore_GetLatestTaskEvent(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	store := notifications.NewPollingStore(nil, logger)
	defer store.Stop()

	// Store multiple events
	for i := 0; i < 5; i++ {
		store.StoreEvent(&notifications.TaskNotification{
			TaskID:    "task-123",
			EventType: "progress",
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"index": i},
		})
	}

	latest := store.GetLatestTaskEvent("task-123")
	require.NotNil(t, latest)
	assert.Equal(t, 4, latest.Data["index"])
}

func TestPollingStore_GetLatestTaskEvent_Empty(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	store := notifications.NewPollingStore(nil, logger)
	defer store.Stop()

	latest := store.GetLatestTaskEvent("nonexistent")
	assert.Nil(t, latest)
}

func TestPollingStore_ClearTaskEvents(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	store := notifications.NewPollingStore(nil, logger)
	defer store.Stop()

	store.StoreEvent(&notifications.TaskNotification{
		TaskID:    "task-123",
		EventType: "progress",
		Timestamp: time.Now(),
	})

	assert.Equal(t, 1, store.GetEventCount("task-123"))

	store.ClearTaskEvents("task-123")

	assert.Equal(t, 0, store.GetEventCount("task-123"))
}

func TestPollingStore_MaxEventsPerTask(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := &notifications.PollingConfig{
		MaxEventsPerTask: 5,
		MaxGlobalEvents:  100,
		EventTTL:         time.Hour,
		CleanupInterval:  time.Hour,
	}

	store := notifications.NewPollingStore(config, logger)
	defer store.Stop()

	// Store more events than the limit
	for i := 0; i < 10; i++ {
		store.StoreEvent(&notifications.TaskNotification{
			TaskID:    "task-123",
			EventType: "progress",
			Timestamp: time.Now(),
		})
	}

	assert.Equal(t, 5, store.GetEventCount("task-123"))
}

func TestPollingStore_GetStats(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	store := notifications.NewPollingStore(nil, logger)
	defer store.Stop()

	store.StoreEvent(&notifications.TaskNotification{
		TaskID:    "task-1",
		EventType: "progress",
		Timestamp: time.Now(),
	})
	store.StoreEvent(&notifications.TaskNotification{
		TaskID:    "task-2",
		EventType: "progress",
		Timestamp: time.Now(),
	})

	stats := store.GetStats()

	assert.NotNil(t, stats)
	assert.Equal(t, 2, stats["tasks_with_events"])
	assert.Equal(t, 2, stats["task_events_total"])
	assert.Equal(t, 2, stats["global_events"])
}

func TestPollingStore_Poll(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	store := notifications.NewPollingStore(nil, logger)
	defer store.Stop()

	// Store events
	for i := 0; i < 5; i++ {
		store.StoreEvent(&notifications.TaskNotification{
			TaskID:    "task-123",
			EventType: "progress",
			Timestamp: time.Now(),
		})
	}

	req := &notifications.PollRequest{
		TaskID: "task-123",
		Limit:  10,
	}

	resp := store.Poll(req)

	assert.NotNil(t, resp)
	assert.Equal(t, 5, resp.Count)
	assert.False(t, resp.HasMore)
	assert.Len(t, resp.Events, 5)
}

func TestPollingStore_Poll_WithLimit(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	store := notifications.NewPollingStore(nil, logger)
	defer store.Stop()

	// Store events
	for i := 0; i < 10; i++ {
		store.StoreEvent(&notifications.TaskNotification{
			TaskID:    "task-123",
			EventType: "progress",
			Timestamp: time.Now(),
		})
	}

	req := &notifications.PollRequest{
		TaskID: "task-123",
		Limit:  5,
	}

	resp := store.Poll(req)

	assert.Equal(t, 5, resp.Count)
	assert.True(t, resp.HasMore)
}

func TestPollingStore_Poll_Global(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	store := notifications.NewPollingStore(nil, logger)
	defer store.Stop()

	// Store events from different tasks
	store.StoreEvent(&notifications.TaskNotification{
		TaskID:    "task-1",
		EventType: "progress",
		Timestamp: time.Now(),
	})
	store.StoreEvent(&notifications.TaskNotification{
		TaskID:    "task-2",
		EventType: "progress",
		Timestamp: time.Now(),
	})

	req := &notifications.PollRequest{
		Limit: 10,
	}

	resp := store.Poll(req)

	assert.Equal(t, 2, resp.Count)
}

func TestPollingSubscriber_Creation(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	store := notifications.NewPollingStore(nil, logger)
	defer store.Stop()

	sub := notifications.NewPollingSubscriber("sub-1", "task-123", store)

	assert.NotNil(t, sub)
	assert.Equal(t, "sub-1", sub.ID())
	assert.Equal(t, notifications.NotificationTypePolling, sub.Type())
	assert.True(t, sub.IsActive())
}

func TestPollingSubscriber_Notify(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	store := notifications.NewPollingStore(nil, logger)
	defer store.Stop()

	sub := notifications.NewPollingSubscriber("sub-1", "task-123", store)

	notification := &notifications.TaskNotification{
		TaskID:    "task-123",
		EventType: "progress",
		Timestamp: time.Now(),
	}

	err := sub.Notify(nil, notification)
	require.NoError(t, err)

	assert.Equal(t, 1, store.GetEventCount("task-123"))
}

func TestPollingSubscriber_Close(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	store := notifications.NewPollingStore(nil, logger)
	defer store.Stop()

	sub := notifications.NewPollingSubscriber("sub-1", "task-123", store)

	assert.True(t, sub.IsActive())

	err := sub.Close()
	require.NoError(t, err)

	assert.False(t, sub.IsActive())
}

func TestPollRequest_Fields(t *testing.T) {
	since := time.Now()
	req := &notifications.PollRequest{
		TaskID: "task-123",
		Since:  &since,
		Limit:  50,
	}

	assert.Equal(t, "task-123", req.TaskID)
	assert.NotNil(t, req.Since)
	assert.Equal(t, 50, req.Limit)
}

func TestPollResponse_Fields(t *testing.T) {
	resp := &notifications.PollResponse{
		Events:    []*notifications.TaskNotification{},
		Count:     5,
		Timestamp: time.Now(),
		HasMore:   true,
	}

	assert.NotNil(t, resp.Events)
	assert.Equal(t, 5, resp.Count)
	assert.False(t, resp.Timestamp.IsZero())
	assert.True(t, resp.HasMore)
}
