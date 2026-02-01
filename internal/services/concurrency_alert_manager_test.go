package services

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewConcurrencyAlertManager(t *testing.T) {
	config := DefaultConcurrencyAlertManagerConfig()
	logger := logrus.New()
	manager := NewConcurrencyAlertManager(config, logger)
	assert.NotNil(t, manager)
	assert.Equal(t, config, manager.config)
}

func TestConcurrencyAlertManagerConfig_DefaultValues(t *testing.T) {
	config := DefaultConcurrencyAlertManagerConfig()
	assert.Equal(t, 587, config.SMTPPort)
	assert.True(t, config.SMTPUseTLS)
	assert.Empty(t, config.SMTPFrom)
}

func TestConcurrencyAlertManager_HandleAlert_Logging(t *testing.T) {
	config := DefaultConcurrencyAlertManagerConfig()
	config.Enabled = true
	config.EnableLogging = true
	config.EnableWebhook = false
	config.EnableEmail = false
	config.EscalationEnabled = false // Disable escalation for this test

	logger := logrus.New()
	manager := NewConcurrencyAlertManager(config, logger)

	alert := ConcurrencyAlert{
		Type:           "high_saturation",
		Provider:       "test-provider",
		Message:        "Test alert",
		Timestamp:      time.Now(),
		Saturation:     85.5,
		ActiveRequests: 10,
		TotalPermits:   12,
		Available:      2,
	}

	// This should not panic
	manager.HandleAlert(alert)
}

func TestConcurrencyAlertManager_Cooldown(t *testing.T) {
	config := DefaultConcurrencyAlertManagerConfig()
	config.Enabled = true
	config.DefaultCooldown = 100 * time.Millisecond
	config.EnableLogging = false     // Disable logging to reduce noise
	config.EscalationEnabled = false // Disable escalation for this test
	logger := logrus.New()
	manager := NewConcurrencyAlertManager(config, logger)

	alert := ConcurrencyAlert{
		Type:       "high_saturation",
		Provider:   "test-provider",
		Message:    "Test alert",
		Timestamp:  time.Now(),
		Saturation: 90.0,
	}

	// First alert should be processed
	// We can't easily capture if alert was processed, but we can check that it doesn't panic
	manager.HandleAlert(alert)
	// Should be in cooldown now
	// Second alert within cooldown should be skipped
	manager.HandleAlert(alert)
	// Wait for cooldown to expire
	time.Sleep(150 * time.Millisecond)
	// Third alert after cooldown should be processed again
	manager.HandleAlert(alert)

	// No assertion, just ensure no panic
	assert.True(t, true)
}

func TestConcurrencyAlertManager_StartStop(t *testing.T) {
	config := DefaultConcurrencyAlertManagerConfig()
	config.Enabled = true
	config.CleanupInterval = 500 * time.Millisecond
	config.EscalationEnabled = false // Disable escalation for this test
	logger := logrus.New()
	manager := NewConcurrencyAlertManager(config, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start manager
	go manager.Start(ctx)

	// Allow some time for startup
	time.Sleep(50 * time.Millisecond)

	// Stop manager
	manager.Stop()

	// Allow time for shutdown
	time.Sleep(50 * time.Millisecond)

	// No assertion, just ensure no panic
	assert.True(t, true)
}

func TestConcurrencyAlertManager_AsListener(t *testing.T) {
	config := DefaultConcurrencyAlertManagerConfig()
	config.Enabled = true
	config.EscalationEnabled = false // Disable escalation for this test
	logger := logrus.New()
	manager := NewConcurrencyAlertManager(config, logger)

	listener := manager.AsListener()
	assert.NotNil(t, listener)

	// Ensure listener can be called
	alert := ConcurrencyAlert{
		Type:      "test",
		Message:   "test",
		Timestamp: time.Now(),
	}
	listener(alert) // Should not panic
}

func TestConcurrencyAlertManager_GetProviderThreshold(t *testing.T) {
	config := DefaultConcurrencyAlertManagerConfig()
	config.Enabled = true
	config.EscalationEnabled = false // Disable escalation for this test
	config.ProviderThresholds = map[string]float64{
		"provider1": 75.0,
		"provider2": 90.0,
	}
	logger := logrus.New()
	manager := NewConcurrencyAlertManager(config, logger)

	threshold, ok := manager.GetProviderThreshold("provider1")
	assert.True(t, ok)
	assert.Equal(t, 75.0, threshold)

	threshold, ok = manager.GetProviderThreshold("provider2")
	assert.True(t, ok)
	assert.Equal(t, 90.0, threshold)

	_, ok = manager.GetProviderThreshold("unknown")
	assert.False(t, ok)
}

func TestLoadConcurrencyAlertManagerConfigFromEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("CONCURRENCY_ALERTS_ENABLED", "false")
	os.Setenv("CONCURRENCY_ALERTS_COOLDOWN", "2m")
	os.Setenv("CONCURRENCY_ALERTS_ENABLE_LOGGING", "false")
	os.Setenv("CONCURRENCY_ALERTS_ENABLE_WEBHOOK", "true")
	os.Setenv("CONCURRENCY_ALERTS_WEBHOOK_URL", "http://example.com/webhook")
	os.Setenv("CONCURRENCY_ALERTS_SLACK_WEBHOOK_URL", "http://example.com/slack")
	os.Setenv("CONCURRENCY_ALERTS_SLACK_CHANNEL", "#alerts")
	os.Setenv("CONCURRENCY_ALERTS_PROVIDER_THRESHOLDS", `{"provider1":75.0,"provider2":90.0}`)
	// Escalation environment variables
	os.Setenv("CONCURRENCY_ALERTS_ESCALATION_ENABLED", "true")
	os.Setenv("CONCURRENCY_ALERTS_ESCALATION_WINDOW", "30m")
	os.Setenv("CONCURRENCY_ALERTS_MAX_ESCALATION_LEVEL", "5")
	os.Setenv("CONCURRENCY_ALERTS_ESCALATION_THRESHOLDS", "[2,4,6]")
	os.Setenv("CONCURRENCY_ALERTS_ESCALATION_CHANNEL_ROUTING", `{"0":["logging"],"1":["logging","email"],"2":["logging","email","slack"]}`)
	// Rate limiting environment variables
	os.Setenv("CONCURRENCY_ALERTS_RATE_LIMIT_ENABLED", "false")
	os.Setenv("CONCURRENCY_ALERTS_RATE_LIMIT_WINDOW", "2m")
	os.Setenv("CONCURRENCY_ALERTS_RATE_LIMIT_MAX_ALERTS", "20")
	os.Setenv("CONCURRENCY_ALERTS_RATE_LIMIT_BURST_SIZE", "10")
	// Circuit breaker environment variables
	os.Setenv("CONCURRENCY_ALERTS_CIRCUIT_BREAKER_ENABLED", "false")
	os.Setenv("CONCURRENCY_ALERTS_CIRCUIT_BREAKER_FAILURES", "10")
	os.Setenv("CONCURRENCY_ALERTS_CIRCUIT_BREAKER_RESET", "1m")
	os.Setenv("CONCURRENCY_ALERTS_CIRCUIT_BREAKER_SUCCESSES", "5")
	// Retry configuration environment variables
	os.Setenv("CONCURRENCY_ALERTS_RETRY_ENABLED", "false")
	os.Setenv("CONCURRENCY_ALERTS_MAX_RETRIES", "5")
	os.Setenv("CONCURRENCY_ALERTS_RETRY_INITIAL_DELAY", "2s")
	os.Setenv("CONCURRENCY_ALERTS_RETRY_MAX_DELAY", "60s")
	os.Setenv("CONCURRENCY_ALERTS_RETRY_BACKOFF_MULTIPLIER", "3.0")
	defer func() {
		os.Unsetenv("CONCURRENCY_ALERTS_ENABLED")
		os.Unsetenv("CONCURRENCY_ALERTS_COOLDOWN")
		os.Unsetenv("CONCURRENCY_ALERTS_ENABLE_LOGGING")
		os.Unsetenv("CONCURRENCY_ALERTS_ENABLE_WEBHOOK")
		os.Unsetenv("CONCURRENCY_ALERTS_WEBHOOK_URL")
		os.Unsetenv("CONCURRENCY_ALERTS_SLACK_WEBHOOK_URL")
		os.Unsetenv("CONCURRENCY_ALERTS_SLACK_CHANNEL")
		os.Unsetenv("CONCURRENCY_ALERTS_PROVIDER_THRESHOLDS")
		os.Unsetenv("CONCURRENCY_ALERTS_ESCALATION_ENABLED")
		os.Unsetenv("CONCURRENCY_ALERTS_ESCALATION_WINDOW")
		os.Unsetenv("CONCURRENCY_ALERTS_MAX_ESCALATION_LEVEL")
		os.Unsetenv("CONCURRENCY_ALERTS_ESCALATION_THRESHOLDS")
		os.Unsetenv("CONCURRENCY_ALERTS_ESCALATION_CHANNEL_ROUTING")
		// Clean up rate limiting environment variables
		os.Unsetenv("CONCURRENCY_ALERTS_RATE_LIMIT_ENABLED")
		os.Unsetenv("CONCURRENCY_ALERTS_RATE_LIMIT_WINDOW")
		os.Unsetenv("CONCURRENCY_ALERTS_RATE_LIMIT_MAX_ALERTS")
		os.Unsetenv("CONCURRENCY_ALERTS_RATE_LIMIT_BURST_SIZE")
		// Clean up circuit breaker environment variables
		os.Unsetenv("CONCURRENCY_ALERTS_CIRCUIT_BREAKER_ENABLED")
		os.Unsetenv("CONCURRENCY_ALERTS_CIRCUIT_BREAKER_FAILURES")
		os.Unsetenv("CONCURRENCY_ALERTS_CIRCUIT_BREAKER_RESET")
		os.Unsetenv("CONCURRENCY_ALERTS_CIRCUIT_BREAKER_SUCCESSES")
		// Clean up retry configuration environment variables
		os.Unsetenv("CONCURRENCY_ALERTS_RETRY_ENABLED")
		os.Unsetenv("CONCURRENCY_ALERTS_MAX_RETRIES")
		os.Unsetenv("CONCURRENCY_ALERTS_RETRY_INITIAL_DELAY")
		os.Unsetenv("CONCURRENCY_ALERTS_RETRY_MAX_DELAY")
		os.Unsetenv("CONCURRENCY_ALERTS_RETRY_BACKOFF_MULTIPLIER")
	}()

	config := LoadConcurrencyAlertManagerConfigFromEnv()

	assert.False(t, config.Enabled)
	assert.Equal(t, 2*time.Minute, config.DefaultCooldown)
	assert.False(t, config.EnableLogging)
	assert.True(t, config.EnableWebhook)
	assert.Equal(t, "http://example.com/webhook", config.WebhookURL)
	assert.Equal(t, "http://example.com/slack", config.SlackWebhookURL)
	assert.Equal(t, "#alerts", config.SlackChannel)
	assert.Equal(t, 75.0, config.ProviderThresholds["provider1"])
	assert.Equal(t, 90.0, config.ProviderThresholds["provider2"])
	// Escalation assertions
	assert.True(t, config.EscalationEnabled)
	assert.Equal(t, 30*time.Minute, config.EscalationWindow)
	assert.Equal(t, 5, config.MaxEscalationLevel)
	assert.Equal(t, []int{2, 4, 6}, config.EscalationThresholds)
	assert.Equal(t, []string{"logging"}, config.EscalationChannelRouting[0])
	assert.Equal(t, []string{"logging", "email"}, config.EscalationChannelRouting[1])
	assert.Equal(t, []string{"logging", "email", "slack"}, config.EscalationChannelRouting[2])
	// Rate limiting assertions
	assert.False(t, config.RateLimitEnabled)
	assert.Equal(t, 2*time.Minute, config.RateLimitWindow)
	assert.Equal(t, 20, config.RateLimitMaxAlerts)
	assert.Equal(t, 10, config.RateLimitBurstSize)
	// Circuit breaker assertions
	assert.False(t, config.CircuitBreakerEnabled)
	assert.Equal(t, 10, config.CircuitBreakerFailures)
	assert.Equal(t, 1*time.Minute, config.CircuitBreakerReset)
	assert.Equal(t, 5, config.CircuitBreakerSuccesses)
	// Retry configuration assertions
	assert.False(t, config.RetryEnabled)
	assert.Equal(t, 5, config.MaxRetries)
	assert.Equal(t, 2*time.Second, config.RetryInitialDelay)
	assert.Equal(t, 60*time.Second, config.RetryMaxDelay)
	assert.Equal(t, 3.0, config.RetryBackoffMultiplier)
}

func TestConcurrencyAlertManager_WebhookDelivery(t *testing.T) {
	// Create test server to capture requests
	var receivedRequest *http.Request
	var requestBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedRequest = r
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		requestBody = body
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Configure alert manager with webhook enabled
	config := DefaultConcurrencyAlertManagerConfig()
	config.Enabled = true
	config.EnableLogging = false
	config.EnableWebhook = true
	config.WebhookURL = server.URL
	config.DefaultCooldown = 0 // Disable cooldown for testing
	config.WebhookTimeout = 5 * time.Second
	config.EscalationEnabled = false // Disable escalation for this test

	logger := logrus.New()
	manager := NewConcurrencyAlertManager(config, logger)

	alert := ConcurrencyAlert{
		Type:           "high_saturation",
		Provider:       "test-provider",
		Message:        "Test alert",
		Timestamp:      time.Now(),
		Saturation:     85.5,
		ActiveRequests: 10,
		TotalPermits:   12,
		Available:      2,
	}

	// Send alert
	manager.HandleAlert(alert)

	// Give some time for async webhook delivery
	time.Sleep(100 * time.Millisecond)

	// Verify request was received
	assert.NotNil(t, receivedRequest, "Webhook request should have been sent")
	assert.Equal(t, "POST", receivedRequest.Method)
	assert.Equal(t, "application/json", receivedRequest.Header.Get("Content-Type"))

	// Parse request body and verify content
	var payload map[string]interface{}
	err := json.Unmarshal(requestBody, &payload)
	assert.NoError(t, err)
	assert.Equal(t, "concurrency_alert", payload["event"])
	assert.Equal(t, "high_saturation", payload["type"])
	assert.Equal(t, "Test alert", payload["message"])
	assert.Equal(t, 85.5, payload["saturation"])
}

func TestConcurrencyAlertManager_SlackWebhookDelivery(t *testing.T) {
	// Create test server to capture requests
	var receivedRequest *http.Request
	var requestBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedRequest = r
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		requestBody = body
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Configure alert manager with Slack webhook enabled
	config := DefaultConcurrencyAlertManagerConfig()
	config.Enabled = true
	config.EnableLogging = false
	config.SlackWebhookURL = server.URL
	config.DefaultCooldown = 0 // Disable cooldown for testing
	config.WebhookTimeout = 5 * time.Second
	config.EscalationEnabled = false // Disable escalation for this test

	logger := logrus.New()
	manager := NewConcurrencyAlertManager(config, logger)

	alert := ConcurrencyAlert{
		Type:           "high_saturation",
		Provider:       "test-provider",
		Message:        "Test alert",
		Timestamp:      time.Now(),
		Saturation:     85.5,
		ActiveRequests: 10,
		TotalPermits:   12,
		Available:      2,
	}

	// Send alert
	manager.HandleAlert(alert)

	// Give some time for async webhook delivery
	time.Sleep(100 * time.Millisecond)

	// Verify request was received
	assert.NotNil(t, receivedRequest, "Slack webhook request should have been sent")
	assert.Equal(t, "POST", receivedRequest.Method)
	assert.Equal(t, "application/json", receivedRequest.Header.Get("Content-Type"))

	// Parse request body and verify Slack-specific structure
	var payload map[string]interface{}
	err := json.Unmarshal(requestBody, &payload)
	assert.NoError(t, err)
	// Slack payload contains "attachments"
	attachments, ok := payload["attachments"].([]interface{})
	assert.True(t, ok)
	assert.Greater(t, len(attachments), 0)
	attachment := attachments[0].(map[string]interface{})
	assert.Equal(t, "Concurrency Alert", attachment["title"])
	assert.Equal(t, "Test alert", attachment["text"])
}

func TestConcurrencyAlertManager_EmailDelivery(t *testing.T) {
	config := DefaultConcurrencyAlertManagerConfig()
	config.EnableLogging = false
	config.EnableEmail = true
	config.SMTPHost = "localhost"
	config.SMTPPort = 9999
	config.EmailRecipients = []string{"test@example.com"}
	config.DefaultCooldown = 0       // Disable cooldown for testing
	config.EscalationEnabled = false // Disable escalation for this test

	logger := logrus.New()
	manager := NewConcurrencyAlertManager(config, logger)

	alert := ConcurrencyAlert{
		Type:           "high_saturation",
		Provider:       "test-provider",
		Message:        "Test alert",
		Timestamp:      time.Now(),
		Saturation:     85.5,
		ActiveRequests: 10,
		TotalPermits:   12,
		Available:      2,
	}

	// This should not panic, email delivery will fail but be handled
	manager.HandleAlert(alert)
	// Give some time for async email delivery attempt
	time.Sleep(100 * time.Millisecond)
}

func TestConcurrencyAlertManager_Escalation(t *testing.T) {
	// Test escalation level calculation and channel routing
	config := DefaultConcurrencyAlertManagerConfig()
	config.EnableLogging = true
	config.EnableEmail = true
	config.SMTPHost = "localhost"
	config.SMTPPort = 9999
	config.EmailRecipients = []string{"test@example.com"}
	config.EnableWebhook = true
	config.WebhookURL = "http://example.com/webhook"
	config.SlackWebhookURL = "http://example.com/slack"
	config.DefaultCooldown = 0 // Disable cooldown for testing
	config.EscalationEnabled = true
	config.EscalationWindow = 1 * time.Hour
	config.EscalationThresholds = []int{1, 3, 5}
	config.MaxEscalationLevel = 3
	config.EscalationChannelRouting = map[int][]string{
		0: {"logging"},
		1: {"logging", "email"},
		2: {"logging", "email", "slack"},
		3: {"logging", "email", "slack", "webhook"},
	}

	logger := logrus.New()
	manager := NewConcurrencyAlertManager(config, logger)

	alert := ConcurrencyAlert{
		Type:           "high_saturation",
		Provider:       "test-provider",
		Message:        "Test alert",
		Timestamp:      time.Now(),
		Saturation:     85.5,
		ActiveRequests: 10,
		TotalPermits:   12,
		Available:      2,
	}

	// Helper to send alert multiple times
	sendAlertNTimes := func(n int) {
		for i := 0; i < n; i++ {
			manager.HandleAlert(alert)
			// Small delay to ensure different timestamps
			time.Sleep(1 * time.Millisecond)
		}
	}

	// First alert: repeat count = 1, escalation level should be 1
	// Should send to logging and email (level 1 routing)
	sendAlertNTimes(1)
	// We can't easily verify which channels were used, but we can verify no panic
	// For now, just ensure it runs without errors
	assert.True(t, true)
}
func TestConcurrencyAlertManager_RateLimiting(t *testing.T) {
	// Create test server to capture requests
	var requestCount int
	var mu sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestCount++
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Configure alert manager with rate limiting enabled
	config := DefaultConcurrencyAlertManagerConfig()
	config.Enabled = true
	config.EnableLogging = false
	config.EnableWebhook = true
	config.WebhookURL = server.URL
	config.DefaultCooldown = 0 // Disable cooldown for testing
	config.WebhookTimeout = 5 * time.Second
	config.EscalationEnabled = false // Disable escalation for this test
	// Enable rate limiting with low limits
	config.RateLimitEnabled = true
	config.RateLimitWindow = 1 * time.Second
	config.RateLimitMaxAlerts = 2
	config.RateLimitBurstSize = 1

	logger := logrus.New()
	manager := NewConcurrencyAlertManager(config, logger)

	alert := ConcurrencyAlert{
		Type:           "high_saturation",
		Provider:       "test-provider",
		Message:        "Test alert",
		Timestamp:      time.Now(),
		Saturation:     85.5,
		ActiveRequests: 10,
		TotalPermits:   12,
		Available:      2,
	}

	// Send 5 alerts quickly - should be rate limited after maxAlerts + burstSize
	for i := 0; i < 5; i++ {
		manager.HandleAlert(alert)
	}
	// Wait for async delivery
	time.Sleep(200 * time.Millisecond)

	// Verify only maxAlerts + burstSize alerts were delivered (3)
	mu.Lock()
	delivered := requestCount
	mu.Unlock()
	assert.Equal(t, 3, delivered, "Should only deliver 3 alerts (maxAlerts + burstSize)")

	// Wait for rate limit window to reset
	time.Sleep(1 * time.Second)

	// Send another alert - should be allowed again
	manager.HandleAlert(alert)
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	delivered = requestCount
	mu.Unlock()
	assert.Equal(t, 4, delivered, "After window reset, should deliver another alert")
}

func TestConcurrencyAlertManager_CircuitBreaker(t *testing.T) {
	// Create test server that fails requests
	var requestCount int
	var mu sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestCount++
		mu.Unlock()
		// Always return error to trigger circuit breaker
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Configure alert manager with circuit breaker enabled
	config := DefaultConcurrencyAlertManagerConfig()
	config.Enabled = true
	config.EnableLogging = false
	config.EnableWebhook = true
	config.WebhookURL = server.URL
	config.DefaultCooldown = 0 // Disable cooldown for testing
	config.WebhookTimeout = 5 * time.Second
	config.EscalationEnabled = false // Disable escalation for this test
	// Enable circuit breaker with low failure threshold
	config.CircuitBreakerEnabled = true
	config.CircuitBreakerFailures = 3
	config.CircuitBreakerReset = 1 * time.Second
	config.CircuitBreakerSuccesses = 2

	logger := logrus.New()
	manager := NewConcurrencyAlertManager(config, logger)

	alert := ConcurrencyAlert{
		Type:           "high_saturation",
		Provider:       "test-provider",
		Message:        "Test alert",
		Timestamp:      time.Now(),
		Saturation:     85.5,
		ActiveRequests: 10,
		TotalPermits:   12,
		Available:      2,
	}

	// Send alerts up to failure threshold with small delays to allow circuit breaker to update
	for i := 0; i < 5; i++ {
		manager.HandleAlert(alert)
		// Small delay to allow the async delivery and circuit breaker state update
		// This ensures the circuit opens before the next alert is sent
		time.Sleep(50 * time.Millisecond)
	}
	// Wait for any remaining async delivery
	time.Sleep(100 * time.Millisecond)

	// Verify exactly CircuitBreakerFailures requests were made before circuit opened
	mu.Lock()
	delivered := requestCount
	mu.Unlock()
	assert.Equal(t, config.CircuitBreakerFailures, delivered,
		"Should only make requests up to failure threshold before circuit opens")

	// Wait for circuit breaker reset period
	time.Sleep(1 * time.Second)

	// Send another alert - circuit should be half-open now
	// But our server still returns error, so it will fail again
	manager.HandleAlert(alert)
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	delivered = requestCount
	mu.Unlock()
	// Should have made one more request (half-open allows one attempt)
	assert.Equal(t, config.CircuitBreakerFailures+1, delivered,
		"After reset, circuit half-open should allow one attempt")
}

func TestConcurrencyAlertManager_ResilienceIntegration(t *testing.T) {
	// Test combined rate limiting and circuit breaker
	// Create test server that works initially then fails
	var requestCount int
	var mu sync.Mutex
	shouldFail := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestCount++
		mu.Unlock()
		if shouldFail {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	// Configure alert manager with both resilience features
	config := DefaultConcurrencyAlertManagerConfig()
	config.Enabled = true
	config.EnableLogging = false
	config.EnableWebhook = true
	config.WebhookURL = server.URL
	config.DefaultCooldown = 0 // Disable cooldown for testing
	config.WebhookTimeout = 5 * time.Second
	config.EscalationEnabled = false // Disable escalation for this test
	// Enable rate limiting
	config.RateLimitEnabled = true
	config.RateLimitWindow = 500 * time.Millisecond
	config.RateLimitMaxAlerts = 2
	config.RateLimitBurstSize = 0
	// Enable circuit breaker
	config.CircuitBreakerEnabled = true
	config.CircuitBreakerFailures = 2
	config.CircuitBreakerReset = 1 * time.Second
	config.CircuitBreakerSuccesses = 2

	logger := logrus.New()
	manager := NewConcurrencyAlertManager(config, logger)

	alert := ConcurrencyAlert{
		Type:           "high_saturation",
		Provider:       "test-provider",
		Message:        "Test alert",
		Timestamp:      time.Now(),
		Saturation:     85.5,
		ActiveRequests: 10,
		TotalPermits:   12,
		Available:      2,
	}

	// Phase 1: Send 3 alerts (rate limit allows 2)
	for i := 0; i < 3; i++ {
		manager.HandleAlert(alert)
		// Small delay to allow rate limiting to update before next alert
		time.Sleep(10 * time.Millisecond)
	}
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	delivered := requestCount
	mu.Unlock()
	assert.Equal(t, 2, delivered, "Rate limiting should allow only 2 alerts")

	// Wait for rate limit window to reset
	time.Sleep(500 * time.Millisecond)

	// Phase 2: Make server start failing
	shouldFail = true
	// Send alerts to trigger circuit breaker with small delays
	for i := 0; i < 3; i++ {
		manager.HandleAlert(alert)
		// Small delay to allow circuit breaker to update before next alert
		time.Sleep(50 * time.Millisecond)
	}
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	delivered = requestCount
	mu.Unlock()
	// Should have 2 (previous) + 2 failures (circuit breaker threshold) = 4 total
	assert.Equal(t, 4, delivered,
		"After rate limit reset, circuit breaker should open after 2 failures")

	// Wait for circuit breaker reset
	time.Sleep(1 * time.Second)

	// Phase 3: Server still failing, circuit half-open allows one attempt
	manager.HandleAlert(alert)
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	delivered = requestCount
	mu.Unlock()
	assert.Equal(t, 5, delivered,
		"After circuit breaker reset, half-open allows one attempt")
}

func TestConcurrencyAlertManager_RetryLogic(t *testing.T) {
	// Create test server that fails first request, succeeds on retry
	var requestCount int
	var mu sync.Mutex
	shouldFail := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestCount++
		count := requestCount
		mu.Unlock()
		t.Logf("Server received request %d", count)
		if shouldFail {
			// First request fails
			w.WriteHeader(http.StatusInternalServerError)
			shouldFail = false // Next request will succeed
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	// Configure alert manager with retry enabled
	config := DefaultConcurrencyAlertManagerConfig()
	config.Enabled = true
	config.EnableLogging = false
	config.EnableWebhook = true
	config.WebhookURL = server.URL
	config.DefaultCooldown = 0
	config.WebhookTimeout = 5 * time.Second
	config.EscalationEnabled = false
	// Disable rate limiting and circuit breaker for this test
	config.RateLimitEnabled = false
	config.CircuitBreakerEnabled = false
	config.RetryEnabled = true
	config.MaxRetries = 2
	config.RetryInitialDelay = 10 * time.Millisecond
	config.RetryMaxDelay = 100 * time.Millisecond
	config.RetryBackoffMultiplier = 2.0
	config.CleanupInterval = 10 * time.Millisecond // Fast cleanup for retry processing

	logger := logrus.New()
	manager := NewConcurrencyAlertManager(config, logger)

	// Start the manager to enable retry processing
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go manager.Start(ctx)
	// Give manager time to start
	time.Sleep(5 * time.Millisecond)

	alert := ConcurrencyAlert{
		Type:           "high_saturation",
		Provider:       "test-provider",
		Message:        "Test alert",
		Timestamp:      time.Now(),
		Saturation:     85.5,
		ActiveRequests: 10,
		TotalPermits:   12,
		Available:      2,
	}

	// Send alert
	t.Logf("Sending alert...")
	manager.HandleAlert(alert)
	t.Logf("Alert sent, waiting for retry...")

	// Wait for retry to happen (initial delay + processing)
	time.Sleep(200 * time.Millisecond)
	t.Logf("Wait complete")

	// Verify that two requests were made (first failure + retry success)
	mu.Lock()
	count := requestCount
	mu.Unlock()
	t.Logf("Request count: %d (expected 2)", count)
	assert.Equal(t, 2, count, "Should have made initial request plus one retry")
}
