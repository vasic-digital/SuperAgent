package http

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	httppool "dev.helix.agent/internal/http"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	require.NotNil(t, cfg)
	require.NotNil(t, cfg.PoolConfig)
	require.NotNil(t, cfg.QUICConfig)
	assert.False(t, cfg.EnableQUIC, "QUIC should be disabled by default")
}

func TestNewClientAdapter_DefaultConfig(t *testing.T) {
	adapter, err := NewClientAdapter(nil)
	require.NoError(t, err)
	require.NotNil(t, adapter)
	defer adapter.Close()

	assert.False(t, adapter.IsQUICEnabled())
}

func TestNewClientAdapter_WithPoolOnly(t *testing.T) {
	cfg := &Config{
		PoolConfig: httppool.DefaultPoolConfig(),
		EnableQUIC: false,
	}

	adapter, err := NewClientAdapter(cfg)
	require.NoError(t, err)
	require.NotNil(t, adapter)
	defer adapter.Close()

	assert.False(t, adapter.IsQUICEnabled())
	assert.NotNil(t, adapter.PoolMetrics())
	assert.Nil(t, adapter.QUICMetrics())
}

func TestNewClientAdapter_WithQUICEnabled(t *testing.T) {
	cfg := &Config{
		PoolConfig: httppool.DefaultPoolConfig(),
		QUICConfig: httppool.DefaultQUICConfig(),
		EnableQUIC: true,
	}

	adapter, err := NewClientAdapter(cfg)
	require.NoError(t, err)
	require.NotNil(t, adapter)
	defer adapter.Close()

	// QUIC client should be created successfully
	assert.True(t, adapter.IsQUICEnabled())
	assert.NotNil(t, adapter.QUICMetrics())
}

func TestClientAdapter_Get_PoolMode(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	adapter, err := NewClientAdapter(nil)
	require.NoError(t, err)
	defer adapter.Close()

	resp, err := adapter.Get(context.Background(), server.URL+"/test")
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, `{"status":"ok"}`, string(body))
}

func TestClientAdapter_PostJSON_PoolMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}))
	defer server.Close()

	adapter, err := NewClientAdapter(nil)
	require.NoError(t, err)
	defer adapter.Close()

	payload := `{"message":"hello"}`
	resp, err := adapter.PostJSON(
		context.Background(), server.URL+"/api",
		strings.NewReader(payload),
	)
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, payload, string(body))
}

func TestClientAdapter_Do_PoolMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom", "test-value")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	adapter, err := NewClientAdapter(nil)
	require.NoError(t, err)
	defer adapter.Close()

	req, err := http.NewRequestWithContext(
		context.Background(), http.MethodGet, server.URL+"/custom", nil,
	)
	require.NoError(t, err)

	resp, err := adapter.Do(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "test-value", resp.Header.Get("X-Custom"))
}

func TestClientAdapter_PoolMetrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	adapter, err := NewClientAdapter(nil)
	require.NoError(t, err)
	defer adapter.Close()

	// Make a request to generate metrics
	resp, err := adapter.Get(context.Background(), server.URL+"/metrics-test")
	require.NoError(t, err)
	resp.Body.Close()

	metrics := adapter.PoolMetrics()
	require.NotNil(t, metrics)
	assert.True(t, metrics.TotalRequests >= 1,
		"expected at least 1 total request, got %d", metrics.TotalRequests)
}

func TestClientAdapter_HealthCheck(t *testing.T) {
	adapter, err := NewClientAdapter(nil)
	require.NoError(t, err)

	// Check health before close
	status := adapter.HealthCheck()
	assert.True(t, status.OverallReady)
	assert.True(t, status.PoolActive)
	assert.False(t, status.QUICActive)
	assert.False(t, status.QUICEnabled)
	assert.False(t, status.Timestamp.IsZero())

	// Check health after close
	err = adapter.Close()
	require.NoError(t, err)

	status = adapter.HealthCheck()
	assert.False(t, status.OverallReady)
	assert.False(t, status.PoolActive)
}

func TestClientAdapter_HealthCheck_WithQUIC(t *testing.T) {
	cfg := &Config{
		PoolConfig: httppool.DefaultPoolConfig(),
		QUICConfig: httppool.DefaultQUICConfig(),
		EnableQUIC: true,
	}

	adapter, err := NewClientAdapter(cfg)
	require.NoError(t, err)
	defer adapter.Close()

	status := adapter.HealthCheck()
	assert.True(t, status.OverallReady)
	assert.True(t, status.PoolActive)
	assert.True(t, status.QUICActive)
	assert.True(t, status.QUICEnabled)
}

func TestClientAdapter_Close_Idempotent(t *testing.T) {
	adapter, err := NewClientAdapter(nil)
	require.NoError(t, err)

	// First close should succeed
	err = adapter.Close()
	require.NoError(t, err)

	// Second close should also succeed (idempotent)
	err = adapter.Close()
	require.NoError(t, err)
}

func TestClientAdapter_Do_AfterClose(t *testing.T) {
	adapter, err := NewClientAdapter(nil)
	require.NoError(t, err)

	err = adapter.Close()
	require.NoError(t, err)

	// Requests after close should fail
	req, _ := http.NewRequestWithContext(
		context.Background(), http.MethodGet, "http://localhost/test", nil,
	)
	_, err = adapter.Do(req)
	assert.Error(t, err)
}

func TestClientAdapter_RoundTripper_PoolMode(t *testing.T) {
	adapter, err := NewClientAdapter(nil)
	require.NoError(t, err)
	defer adapter.Close()

	rt := adapter.RoundTripper()
	require.NotNil(t, rt)

	// Should return default transport when QUIC is disabled
	assert.Equal(t, http.DefaultTransport, rt)
}

func TestClientAdapter_RoundTripper_QUICMode(t *testing.T) {
	cfg := &Config{
		PoolConfig: httppool.DefaultPoolConfig(),
		QUICConfig: httppool.DefaultQUICConfig(),
		EnableQUIC: true,
	}

	adapter, err := NewClientAdapter(cfg)
	require.NoError(t, err)
	defer adapter.Close()

	rt := adapter.RoundTripper()
	require.NotNil(t, rt)

	// Should NOT be the default transport when QUIC is enabled
	assert.NotEqual(t, http.DefaultTransport, rt)
}

func TestClientAdapter_ConcurrentAccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	adapter, err := NewClientAdapter(nil)
	require.NoError(t, err)
	defer adapter.Close()

	const goroutines = 10
	done := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			resp, err := adapter.Get(context.Background(), server.URL+"/concurrent")
			if err != nil {
				done <- err
				return
			}
			resp.Body.Close()
			done <- nil
		}()
	}

	for i := 0; i < goroutines; i++ {
		select {
		case err := <-done:
			assert.NoError(t, err)
		case <-time.After(10 * time.Second):
			t.Fatal("timed out waiting for concurrent requests")
		}
	}

	metrics := adapter.PoolMetrics()
	require.NotNil(t, metrics)
	assert.True(t, metrics.TotalRequests >= int64(goroutines),
		"expected at least %d total requests, got %d", goroutines, metrics.TotalRequests)
}

func TestClientAdapter_QUICMetrics_Nil_WhenDisabled(t *testing.T) {
	adapter, err := NewClientAdapter(nil)
	require.NoError(t, err)
	defer adapter.Close()

	assert.Nil(t, adapter.QUICMetrics())
}

func TestHealthStatus_Fields(t *testing.T) {
	status := HealthStatus{
		Timestamp:       time.Now(),
		PoolActive:      true,
		PoolClientCount: 5,
		QUICActive:      false,
		QUICEnabled:     false,
		OverallReady:    true,
	}

	assert.True(t, status.PoolActive)
	assert.Equal(t, 5, status.PoolClientCount)
	assert.False(t, status.QUICActive)
	assert.False(t, status.QUICEnabled)
	assert.True(t, status.OverallReady)
}
