package notifications

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/models"
)

// Tests for DefaultWebhookConfig
func TestDefaultWebhookConfig(t *testing.T) {
	config := DefaultWebhookConfig()

	assert.Equal(t, 5, config.MaxRetries)
	assert.Equal(t, time.Second, config.RetryBackoff)
	assert.Equal(t, 5*time.Minute, config.MaxBackoff)
	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.Equal(t, 5, config.WorkerCount)
	assert.Equal(t, 1000, config.QueueSize)
	assert.Equal(t, "X-HelixAgent-Signature", config.SignatureHeader)
}

// Tests for NewWebhookDispatcher
func TestNewWebhookDispatcher(t *testing.T) {
	logger := testLogger()

	t.Run("with default config", func(t *testing.T) {
		dispatcher := NewWebhookDispatcher(nil, logger)
		require.NotNil(t, dispatcher)

		assert.NotNil(t, dispatcher.webhooks)
		assert.NotNil(t, dispatcher.deliveryQueue)
		assert.NotNil(t, dispatcher.client)
		assert.Equal(t, logger, dispatcher.logger)

		_ = dispatcher.Stop()
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &WebhookConfig{
			MaxRetries:      3,
			RetryBackoff:    500 * time.Millisecond,
			MaxBackoff:      2 * time.Minute,
			Timeout:         15 * time.Second,
			WorkerCount:     3,
			QueueSize:       500,
			SignatureHeader: "X-Custom-Signature",
		}

		dispatcher := NewWebhookDispatcher(config, logger)
		require.NotNil(t, dispatcher)

		assert.Equal(t, 3, dispatcher.config.MaxRetries)
		assert.Equal(t, 500*time.Millisecond, dispatcher.config.RetryBackoff)

		_ = dispatcher.Stop()
	})
}

// Tests for WebhookDispatcher Start/Stop
func TestWebhookDispatcher_StartStop(t *testing.T) {
	logger := testLogger()
	dispatcher := NewWebhookDispatcher(nil, logger)

	err := dispatcher.Start()
	assert.NoError(t, err)

	err = dispatcher.Stop()
	assert.NoError(t, err)
}

// Tests for RegisterWebhook
func TestWebhookDispatcher_RegisterWebhook(t *testing.T) {
	logger := testLogger()
	dispatcher := NewWebhookDispatcher(nil, logger)
	defer func() { _ = dispatcher.Stop() }()

	t.Run("register webhook with ID", func(t *testing.T) {
		webhook := &WebhookRegistration{
			ID:      "webhook-1",
			URL:     "https://example.com/webhook",
			Events:  []string{"task.completed"},
			Enabled: true,
		}

		err := dispatcher.RegisterWebhook(webhook)
		assert.NoError(t, err)

		retrieved, found := dispatcher.GetWebhook("webhook-1")
		assert.True(t, found)
		assert.Equal(t, "https://example.com/webhook", retrieved.URL)
		assert.True(t, retrieved.Enabled)
	})

	t.Run("register webhook without ID", func(t *testing.T) {
		webhook := &WebhookRegistration{
			URL:     "https://example.com/webhook2",
			Events:  []string{"task.failed"},
			Enabled: true,
		}

		err := dispatcher.RegisterWebhook(webhook)
		assert.NoError(t, err)
		assert.NotEmpty(t, webhook.ID)
		assert.False(t, webhook.CreatedAt.IsZero())
	})

	t.Run("register webhook sets defaults", func(t *testing.T) {
		webhook := &WebhookRegistration{
			ID:  "webhook-3",
			URL: "https://example.com/webhook3",
		}

		err := dispatcher.RegisterWebhook(webhook)
		assert.NoError(t, err)
		// Note: Enabled is NOT defaulted to true - it respects the zero value (false)
		// This allows registering disabled webhooks that can be enabled later
		assert.False(t, webhook.Enabled) // Explicitly respects zero value
		assert.False(t, webhook.CreatedAt.IsZero())
	})
}

// Tests for UnregisterWebhook
func TestWebhookDispatcher_UnregisterWebhook(t *testing.T) {
	logger := testLogger()
	dispatcher := NewWebhookDispatcher(nil, logger)
	defer func() { _ = dispatcher.Stop() }()

	webhook := &WebhookRegistration{
		ID:      "webhook-1",
		URL:     "https://example.com/webhook",
		Enabled: true,
	}
	_ = dispatcher.RegisterWebhook(webhook)

	err := dispatcher.UnregisterWebhook("webhook-1")
	assert.NoError(t, err)

	_, found := dispatcher.GetWebhook("webhook-1")
	assert.False(t, found)
}

// Tests for GetWebhook
func TestWebhookDispatcher_GetWebhook(t *testing.T) {
	logger := testLogger()
	dispatcher := NewWebhookDispatcher(nil, logger)
	defer func() { _ = dispatcher.Stop() }()

	t.Run("get existing webhook", func(t *testing.T) {
		webhook := &WebhookRegistration{
			ID:      "webhook-1",
			URL:     "https://example.com/webhook",
			Secret:  "test-secret",
			Events:  []string{"task.completed"},
			Enabled: true,
		}
		_ = dispatcher.RegisterWebhook(webhook)

		retrieved, found := dispatcher.GetWebhook("webhook-1")
		assert.True(t, found)
		assert.Equal(t, "webhook-1", retrieved.ID)
		assert.Equal(t, "test-secret", retrieved.Secret)
	})

	t.Run("get nonexistent webhook", func(t *testing.T) {
		_, found := dispatcher.GetWebhook("nonexistent")
		assert.False(t, found)
	})
}

// Tests for ListWebhooks
func TestWebhookDispatcher_ListWebhooks(t *testing.T) {
	logger := testLogger()
	dispatcher := NewWebhookDispatcher(nil, logger)
	defer func() { _ = dispatcher.Stop() }()

	// Register multiple webhooks
	for i := 0; i < 3; i++ {
		webhook := &WebhookRegistration{
			ID:      "webhook-" + string(rune('a'+i)),
			URL:     "https://example.com/webhook" + string(rune('a'+i)),
			Enabled: true,
		}
		_ = dispatcher.RegisterWebhook(webhook)
	}

	webhooks := dispatcher.ListWebhooks()
	assert.Len(t, webhooks, 3)
}

// Tests for Dispatch
func TestWebhookDispatcher_Dispatch(t *testing.T) {
	logger := testLogger()

	// Create a test HTTP server
	var receivedRequests int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&receivedRequests, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := &WebhookConfig{
		MaxRetries:   1,
		RetryBackoff: 10 * time.Millisecond,
		MaxBackoff:   100 * time.Millisecond,
		Timeout:      5 * time.Second,
		WorkerCount:  2,
		QueueSize:    100,
	}

	dispatcher := NewWebhookDispatcher(config, logger)
	defer func() { _ = dispatcher.Stop() }()

	// Register a webhook pointing to test server
	webhook := &WebhookRegistration{
		ID:      "test-webhook",
		URL:     server.URL,
		Events:  []string{"*"},
		Enabled: true,
	}
	_ = dispatcher.RegisterWebhook(webhook)

	// Create a notification
	task := testTask("task-1")
	notification := &TaskNotification{
		TaskID:    "task-1",
		EventType: "task.completed",
		Timestamp: time.Now(),
		Task:      task,
	}

	// Dispatch
	dispatcher.Dispatch(notification)

	// Wait for delivery
	time.Sleep(200 * time.Millisecond)

	assert.GreaterOrEqual(t, atomic.LoadInt32(&receivedRequests), int32(1))
}

// Tests for Dispatch with event filtering
func TestWebhookDispatcher_Dispatch_EventFiltering(t *testing.T) {
	logger := testLogger()

	var receivedRequests int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&receivedRequests, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dispatcher := NewWebhookDispatcher(nil, logger)
	defer func() { _ = dispatcher.Stop() }()

	// Register webhook for specific event
	webhook := &WebhookRegistration{
		ID:      "test-webhook",
		URL:     server.URL,
		Events:  []string{"task.completed"},
		Enabled: true,
	}
	_ = dispatcher.RegisterWebhook(webhook)

	// Dispatch event that doesn't match
	notification := &TaskNotification{
		TaskID:    "task-1",
		EventType: "task.started",
		Timestamp: time.Now(),
	}
	dispatcher.Dispatch(notification)

	time.Sleep(100 * time.Millisecond)

	// Should not receive request
	assert.Equal(t, int32(0), atomic.LoadInt32(&receivedRequests))

	// Dispatch matching event
	notification2 := &TaskNotification{
		TaskID:    "task-1",
		EventType: "task.completed",
		Timestamp: time.Now(),
		Task:      testTask("task-1"),
	}
	dispatcher.Dispatch(notification2)

	time.Sleep(200 * time.Millisecond)

	// Should receive request
	assert.GreaterOrEqual(t, atomic.LoadInt32(&receivedRequests), int32(1))
}

// Tests for Dispatch with disabled webhook
func TestWebhookDispatcher_Dispatch_DisabledWebhook(t *testing.T) {
	logger := testLogger()

	var receivedRequests int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&receivedRequests, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dispatcher := NewWebhookDispatcher(nil, logger)
	defer func() { _ = dispatcher.Stop() }()

	// Register disabled webhook
	webhook := &WebhookRegistration{
		ID:      "test-webhook",
		URL:     server.URL,
		Events:  []string{"*"},
		Enabled: false,
	}
	_ = dispatcher.RegisterWebhook(webhook)

	notification := &TaskNotification{
		TaskID:    "task-1",
		EventType: "task.completed",
		Timestamp: time.Now(),
	}
	dispatcher.Dispatch(notification)

	time.Sleep(100 * time.Millisecond)

	// Should not receive request
	assert.Equal(t, int32(0), atomic.LoadInt32(&receivedRequests))
}

// Tests for Dispatch with task type filtering
func TestWebhookDispatcher_Dispatch_TaskTypeFiltering(t *testing.T) {
	logger := testLogger()

	var receivedRequests int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&receivedRequests, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dispatcher := NewWebhookDispatcher(nil, logger)
	defer func() { _ = dispatcher.Stop() }()

	// Register webhook for specific task type
	webhook := &WebhookRegistration{
		ID:        "test-webhook",
		URL:       server.URL,
		TaskTypes: []string{"specific-type"},
		Enabled:   true,
	}
	_ = dispatcher.RegisterWebhook(webhook)

	// Dispatch event with different task type
	task := &models.BackgroundTask{
		ID:       "task-1",
		TaskType: "other-type",
		TaskName: "test",
	}
	notification := &TaskNotification{
		TaskID:    "task-1",
		EventType: "task.completed",
		Timestamp: time.Now(),
		Task:      task,
	}
	dispatcher.Dispatch(notification)

	time.Sleep(100 * time.Millisecond)

	// Should not receive request
	assert.Equal(t, int32(0), atomic.LoadInt32(&receivedRequests))

	// Dispatch event with matching task type
	task2 := &models.BackgroundTask{
		ID:       "task-2",
		TaskType: "specific-type",
		TaskName: "test",
	}
	notification2 := &TaskNotification{
		TaskID:    "task-2",
		EventType: "task.completed",
		Timestamp: time.Now(),
		Task:      task2,
	}
	dispatcher.Dispatch(notification2)

	time.Sleep(200 * time.Millisecond)

	// Should receive request
	assert.GreaterOrEqual(t, atomic.LoadInt32(&receivedRequests), int32(1))
}

// Tests for webhook signature
func TestWebhookDispatcher_Signature(t *testing.T) {
	logger := testLogger()

	var receivedSignature string
	var mu sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		receivedSignature = r.Header.Get("X-HelixAgent-Signature")
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dispatcher := NewWebhookDispatcher(nil, logger)
	defer func() { _ = dispatcher.Stop() }()

	// Register webhook with secret
	webhook := &WebhookRegistration{
		ID:      "test-webhook",
		URL:     server.URL,
		Secret:  "test-secret-key",
		Events:  []string{"*"},
		Enabled: true,
	}
	_ = dispatcher.RegisterWebhook(webhook)

	notification := &TaskNotification{
		TaskID:    "task-1",
		EventType: "task.completed",
		Timestamp: time.Now(),
		Task:      testTask("task-1"),
	}
	dispatcher.Dispatch(notification)

	time.Sleep(200 * time.Millisecond)

	// Signature should be present and start with sha256=
	mu.Lock()
	sig := receivedSignature
	mu.Unlock()
	assert.True(t, len(sig) > 0)
	assert.Contains(t, sig, "sha256=")
}

// Tests for webhook custom headers
func TestWebhookDispatcher_CustomHeaders(t *testing.T) {
	logger := testLogger()

	var receivedHeaders http.Header
	var mu sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		receivedHeaders = r.Header.Clone()
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dispatcher := NewWebhookDispatcher(nil, logger)
	defer func() { _ = dispatcher.Stop() }()

	// Register webhook with custom headers
	webhook := &WebhookRegistration{
		ID:  "test-webhook",
		URL: server.URL,
		Headers: map[string]string{
			"X-Custom-Header": "custom-value",
			"Authorization":   "Bearer token123",
		},
		Events:  []string{"*"},
		Enabled: true,
	}
	_ = dispatcher.RegisterWebhook(webhook)

	notification := &TaskNotification{
		TaskID:    "task-1",
		EventType: "task.completed",
		Timestamp: time.Now(),
		Task:      testTask("task-1"),
	}
	dispatcher.Dispatch(notification)

	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	headers := receivedHeaders
	mu.Unlock()
	assert.Equal(t, "custom-value", headers.Get("X-Custom-Header"))
	assert.Equal(t, "Bearer token123", headers.Get("Authorization"))
}

// Tests for webhook payload
func TestWebhookDispatcher_Payload(t *testing.T) {
	logger := testLogger()

	var receivedPayload map[string]interface{}
	var mu sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		_ = json.Unmarshal(body, &receivedPayload)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dispatcher := NewWebhookDispatcher(nil, logger)
	defer func() { _ = dispatcher.Stop() }()

	webhook := &WebhookRegistration{
		ID:      "test-webhook",
		URL:     server.URL,
		Events:  []string{"*"},
		Enabled: true,
	}
	_ = dispatcher.RegisterWebhook(webhook)

	task := &models.BackgroundTask{
		ID:       "task-123",
		TaskType: "compute",
		TaskName: "test-task",
		Status:   models.TaskStatusRunning,
		Progress: 75.5,
	}
	notification := &TaskNotification{
		TaskID:    "task-123",
		EventType: "task.progress",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"custom": "data"},
		Task:      task,
	}
	dispatcher.Dispatch(notification)

	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	payload := receivedPayload
	mu.Unlock()

	assert.Equal(t, "task.progress", payload["event"])
	assert.Equal(t, "task-123", payload["task_id"])
	assert.NotNil(t, payload["timestamp"])

	taskData, ok := payload["task"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "task-123", taskData["id"])
	assert.Equal(t, "compute", taskData["type"])
}

// Tests for webhook retry on failure
func TestWebhookDispatcher_Retry(t *testing.T) {
	logger := testLogger()

	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requestCount, 1)
		if count < 3 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	config := &WebhookConfig{
		MaxRetries:   5,
		RetryBackoff: 10 * time.Millisecond,
		MaxBackoff:   50 * time.Millisecond,
		Timeout:      5 * time.Second,
		WorkerCount:  1,
		QueueSize:    100,
	}

	dispatcher := NewWebhookDispatcher(config, logger)
	defer func() { _ = dispatcher.Stop() }()

	webhook := &WebhookRegistration{
		ID:      "test-webhook",
		URL:     server.URL,
		Events:  []string{"*"},
		Enabled: true,
	}
	_ = dispatcher.RegisterWebhook(webhook)

	notification := &TaskNotification{
		TaskID:    "task-1",
		EventType: "task.completed",
		Timestamp: time.Now(),
		Task:      testTask("task-1"),
	}
	dispatcher.Dispatch(notification)

	// Wait for retries
	time.Sleep(500 * time.Millisecond)

	// Should have multiple requests due to retries
	assert.GreaterOrEqual(t, atomic.LoadInt32(&requestCount), int32(3))
}

// Tests for GetStats
func TestWebhookDispatcher_GetStats(t *testing.T) {
	logger := testLogger()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dispatcher := NewWebhookDispatcher(nil, logger)
	defer func() { _ = dispatcher.Stop() }()

	// Initial stats
	stats := dispatcher.GetStats()
	assert.Equal(t, 0, stats["webhooks_registered"])
	assert.Equal(t, int64(0), stats["deliveries_success"])
	assert.Equal(t, int64(0), stats["deliveries_failed"])

	// Register webhook
	webhook := &WebhookRegistration{
		ID:      "test-webhook",
		URL:     server.URL,
		Events:  []string{"*"},
		Enabled: true,
	}
	_ = dispatcher.RegisterWebhook(webhook)

	stats = dispatcher.GetStats()
	assert.Equal(t, 1, stats["webhooks_registered"])

	// Dispatch
	notification := &TaskNotification{
		TaskID:    "task-1",
		EventType: "task.completed",
		Timestamp: time.Now(),
		Task:      testTask("task-1"),
	}
	dispatcher.Dispatch(notification)

	time.Sleep(200 * time.Millisecond)

	stats = dispatcher.GetStats()
	assert.GreaterOrEqual(t, stats["deliveries_success"].(int64), int64(1))
}

// Tests for LoadWebhooksFromTask
func TestWebhookDispatcher_LoadWebhooksFromTask(t *testing.T) {
	logger := testLogger()
	dispatcher := NewWebhookDispatcher(nil, logger)
	defer func() { _ = dispatcher.Stop() }()

	task := &models.BackgroundTask{
		ID:       "task-1",
		TaskType: "test",
		TaskName: "test-task",
		NotificationConfig: models.NotificationConfig{
			Webhooks: []models.WebhookConfig{
				{
					URL:    "https://example.com/webhook1",
					Events: []string{"task.completed"},
					Headers: map[string]string{
						"X-Custom": "value",
					},
					Secret: "secret1",
				},
				{
					URL:    "https://example.com/webhook2",
					Events: []string{"task.failed"},
				},
			},
		},
	}

	dispatcher.LoadWebhooksFromTask(task)

	webhooks := dispatcher.ListWebhooks()
	assert.Len(t, webhooks, 2)
}

// Tests for WebhookSubscriber
func TestWebhookSubscriber(t *testing.T) {
	logger := testLogger()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dispatcher := NewWebhookDispatcher(nil, logger)
	defer func() { _ = dispatcher.Stop() }()

	webhook := &WebhookRegistration{
		ID:      "test-webhook",
		URL:     server.URL,
		Events:  []string{"*"},
		Enabled: true,
	}
	_ = dispatcher.RegisterWebhook(webhook)

	t.Run("create new subscriber", func(t *testing.T) {
		subscriber := NewWebhookSubscriber("sub-1", "task-1", dispatcher, webhook)

		assert.Equal(t, "sub-1", subscriber.ID())
		assert.Equal(t, NotificationTypeWebhook, subscriber.Type())
		assert.True(t, subscriber.IsActive())
	})

	t.Run("notify subscriber dispatches webhook", func(t *testing.T) {
		subscriber := NewWebhookSubscriber("sub-1", "task-1", dispatcher, webhook)

		notification := &TaskNotification{
			TaskID:    "task-1",
			EventType: "progress",
			Timestamp: time.Now(),
			Task:      testTask("task-1"),
		}

		err := subscriber.Notify(context.Background(), notification)
		assert.NoError(t, err)
	})

	t.Run("close subscriber", func(t *testing.T) {
		subscriber := NewWebhookSubscriber("sub-1", "task-1", dispatcher, webhook)

		err := subscriber.Close()
		assert.NoError(t, err)
		assert.False(t, subscriber.IsActive())
	})
}

// Tests for concurrent operations
func TestWebhookDispatcher_ConcurrentOperations(t *testing.T) {
	logger := testLogger()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dispatcher := NewWebhookDispatcher(nil, logger)
	defer func() { _ = dispatcher.Stop() }()

	var wg sync.WaitGroup

	// Concurrent webhook registration
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			webhook := &WebhookRegistration{
				ID:      "webhook-" + string(rune('a'+id)),
				URL:     server.URL,
				Events:  []string{"*"},
				Enabled: true,
			}
			_ = dispatcher.RegisterWebhook(webhook)
		}(i)
	}

	// Concurrent dispatch
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			notification := &TaskNotification{
				TaskID:    "task-1",
				EventType: "test",
				Timestamp: time.Now(),
				Task:      testTask("task-1"),
			}
			dispatcher.Dispatch(notification)
		}()
	}

	wg.Wait()

	// Verify no panic
	webhooks := dispatcher.ListWebhooks()
	assert.Len(t, webhooks, 10)
}

// Tests for calculateBackoff
func TestWebhookDispatcher_CalculateBackoff(t *testing.T) {
	logger := testLogger()

	config := &WebhookConfig{
		RetryBackoff: time.Second,
		MaxBackoff:   5 * time.Minute,
	}

	dispatcher := NewWebhookDispatcher(config, logger)
	defer func() { _ = dispatcher.Stop() }()

	t.Run("exponential backoff", func(t *testing.T) {
		backoff1 := dispatcher.calculateBackoff(1)
		backoff2 := dispatcher.calculateBackoff(2)
		backoff3 := dispatcher.calculateBackoff(3)

		assert.Equal(t, time.Second, backoff1)
		assert.Equal(t, 2*time.Second, backoff2)
		assert.Equal(t, 4*time.Second, backoff3)
	})

	t.Run("backoff capped at max", func(t *testing.T) {
		backoff := dispatcher.calculateBackoff(20) // Would be very large

		assert.Equal(t, config.MaxBackoff, backoff)
	})
}

// Tests for generateSignature
func TestWebhookDispatcher_GenerateSignature(t *testing.T) {
	logger := testLogger()
	dispatcher := NewWebhookDispatcher(nil, logger)
	defer func() { _ = dispatcher.Stop() }()

	payload := []byte(`{"test":"data"}`)
	secret := "my-secret-key"

	signature := dispatcher.generateSignature(payload, secret)

	assert.True(t, len(signature) > 0)
	assert.Contains(t, signature, "sha256=")

	// Same input should produce same signature
	signature2 := dispatcher.generateSignature(payload, secret)
	assert.Equal(t, signature, signature2)

	// Different secret should produce different signature
	signature3 := dispatcher.generateSignature(payload, "different-secret")
	assert.NotEqual(t, signature, signature3)
}

// Tests for WebhookRegistration
func TestWebhookRegistration(t *testing.T) {
	webhook := &WebhookRegistration{
		ID:        "test-id",
		URL:       "https://example.com/webhook",
		Secret:    "secret123",
		Events:    []string{"task.completed", "task.failed"},
		TaskTypes: []string{"compute", "analysis"},
		Headers: map[string]string{
			"Authorization": "Bearer token",
		},
		Enabled:   true,
		CreatedAt: time.Now(),
	}

	assert.Equal(t, "test-id", webhook.ID)
	assert.Equal(t, "https://example.com/webhook", webhook.URL)
	assert.Equal(t, "secret123", webhook.Secret)
	assert.Len(t, webhook.Events, 2)
	assert.Len(t, webhook.TaskTypes, 2)
	assert.Len(t, webhook.Headers, 1)
	assert.True(t, webhook.Enabled)
}

// Tests for WebhookDelivery
func TestWebhookDelivery(t *testing.T) {
	now := time.Now()
	delivery := &WebhookDelivery{
		ID:          "delivery-1",
		WebhookID:   "webhook-1",
		TaskID:      "task-1",
		Event:       "task.completed",
		Payload:     map[string]interface{}{"key": "value"},
		RetryCount:  2,
		ScheduledAt: now,
	}

	assert.Equal(t, "delivery-1", delivery.ID)
	assert.Equal(t, "webhook-1", delivery.WebhookID)
	assert.Equal(t, "task.completed", delivery.Event)
	assert.Equal(t, 2, delivery.RetryCount)
}

// Tests for WebhookConfig
func TestWebhookConfig(t *testing.T) {
	config := &WebhookConfig{
		MaxRetries:      10,
		RetryBackoff:    2 * time.Second,
		MaxBackoff:      10 * time.Minute,
		Timeout:         1 * time.Minute,
		WorkerCount:     10,
		QueueSize:       5000,
		SignatureHeader: "X-Custom-Sig",
	}

	assert.Equal(t, 10, config.MaxRetries)
	assert.Equal(t, 2*time.Second, config.RetryBackoff)
	assert.Equal(t, 10*time.Minute, config.MaxBackoff)
	assert.Equal(t, 1*time.Minute, config.Timeout)
	assert.Equal(t, 10, config.WorkerCount)
	assert.Equal(t, 5000, config.QueueSize)
	assert.Equal(t, "X-Custom-Sig", config.SignatureHeader)
}

// Tests for webhook auto-disable on repeated failures
func TestWebhookDispatcher_AutoDisable(t *testing.T) {
	logger := testLogger()

	// Server that always fails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	config := &WebhookConfig{
		MaxRetries:   0, // No retries
		RetryBackoff: 1 * time.Millisecond,
		MaxBackoff:   10 * time.Millisecond,
		Timeout:      1 * time.Second,
		WorkerCount:  1,
		QueueSize:    100,
	}

	dispatcher := NewWebhookDispatcher(config, logger)
	defer func() { _ = dispatcher.Stop() }()

	webhook := &WebhookRegistration{
		ID:      "test-webhook",
		URL:     server.URL,
		Events:  []string{"*"},
		Enabled: true,
	}
	_ = dispatcher.RegisterWebhook(webhook)

	// Send many failing requests
	for i := 0; i < 15; i++ {
		notification := &TaskNotification{
			TaskID:    "task-1",
			EventType: "test",
			Timestamp: time.Now(),
			Task:      testTask("task-1"),
		}
		dispatcher.Dispatch(notification)
		time.Sleep(20 * time.Millisecond)
	}

	// Wait for processing
	time.Sleep(500 * time.Millisecond)

	// Webhook should be disabled after too many failures
	retrieved, found := dispatcher.GetWebhook("test-webhook")
	assert.True(t, found)
	assert.False(t, retrieved.Enabled)
	assert.Greater(t, retrieved.FailCount, 10)
}
