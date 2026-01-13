package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	httppool "dev.helix.agent/internal/http"
)

func TestHTTPClientPool_BasicOperation(t *testing.T) {
	pool := httppool.NewHTTPClientPool(nil)
	defer pool.Close()

	// Get a client
	client := pool.GetClient("example.com")
	require.NotNil(t, client)

	// Same host should return same client
	client2 := pool.GetClient("example.com")
	assert.Equal(t, client, client2)

	// Different host should return different client
	client3 := pool.GetClient("other.com")
	assert.NotEqual(t, client, client3)
}

func TestHTTPClientPool_WithConfig(t *testing.T) {
	config := &httppool.PoolConfig{
		MaxIdleConns:          50,
		MaxConnsPerHost:       5,
		IdleConnTimeout:       30 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
		DialTimeout:           2 * time.Second,
		EnableHTTP2:           true,
	}

	pool := httppool.NewHTTPClientPool(config)
	defer pool.Close()

	client := pool.GetClient("example.com")
	require.NotNil(t, client)
}

func TestHTTPClientPool_ConcurrentAccess(t *testing.T) {
	pool := httppool.NewHTTPClientPool(nil)
	defer pool.Close()

	var wg sync.WaitGroup
	hosts := []string{"host1.com", "host2.com", "host3.com", "host4.com"}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			host := hosts[idx%len(hosts)]
			client := pool.GetClient(host)
			assert.NotNil(t, client)
		}(i)
	}

	wg.Wait()
}

func TestHTTPClientPool_RealRequest(t *testing.T) {
	// Create a test server
	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	pool := httppool.NewHTTPClientPool(nil)
	defer pool.Close()

	client := pool.GetClient(server.URL)
	require.NotNil(t, client)

	// Make a request
	resp, err := client.Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int32(1), atomic.LoadInt32(&requestCount))
}

func TestHTTPClientPool_ConnectionReuse(t *testing.T) {
	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	pool := httppool.NewHTTPClientPool(nil)
	defer pool.Close()

	client := pool.GetClient(server.URL)

	// Make multiple requests - should reuse connection
	for i := 0; i < 10; i++ {
		resp, err := client.Get(server.URL)
		require.NoError(t, err)
		resp.Body.Close()
	}

	assert.Equal(t, int32(10), atomic.LoadInt32(&requestCount))
}

func TestHTTPClientPool_Metrics(t *testing.T) {
	pool := httppool.NewHTTPClientPool(nil)
	defer pool.Close()

	// Create some clients
	pool.GetClient("host1.com")
	pool.GetClient("host2.com")
	pool.GetClient("host3.com")

	metrics := pool.Metrics()
	require.NotNil(t, metrics)
	assert.Equal(t, int64(3), metrics.ActiveClients)
}

func TestHTTPClientPool_Close(t *testing.T) {
	pool := httppool.NewHTTPClientPool(nil)

	pool.GetClient("host1.com")
	pool.GetClient("host2.com")

	// Close should not panic
	pool.Close()

	// Getting client after close should still work (creates new)
	client := pool.GetClient("host3.com")
	assert.NotNil(t, client)
}

func TestHTTPClientPool_Timeout(t *testing.T) {
	// Server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := &httppool.PoolConfig{
		ResponseHeaderTimeout: 100 * time.Millisecond,
	}

	pool := httppool.NewHTTPClientPool(config)
	defer pool.Close()

	client := pool.GetClient(server.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
	_, err := client.Do(req)

	// Should timeout
	assert.Error(t, err)
}

func TestRetryClient_Success(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := &httppool.RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
	}

	client := httppool.NewRetryClient(http.DefaultClient, config)

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := client.Do(req)

	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int32(1), atomic.LoadInt32(&attempts)) // No retries needed
}

func TestRetryClient_RetryOnServerError(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attempts, 1)
		if count < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := &httppool.RetryConfig{
		MaxRetries:     5,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
		RetryOn5xx:     true,
	}

	client := httppool.NewRetryClient(http.DefaultClient, config)

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := client.Do(req)

	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int32(3), atomic.LoadInt32(&attempts)) // 2 retries + 1 success
}

func TestRetryClient_MaxRetriesExceeded(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	config := &httppool.RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
		RetryOn5xx:     true,
	}

	client := httppool.NewRetryClient(http.DefaultClient, config)

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := client.Do(req)

	require.NoError(t, err)
	defer resp.Body.Close()

	// After max retries, returns the last response
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	assert.Equal(t, int32(4), atomic.LoadInt32(&attempts)) // 1 initial + 3 retries
}

func TestRetryClient_ExponentialBackoff(t *testing.T) {
	var timestamps []time.Time
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		timestamps = append(timestamps, time.Now())
		mu.Unlock()
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	config := &httppool.RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 50 * time.Millisecond,
		MaxBackoff:     500 * time.Millisecond,
		BackoffFactor:  2.0,
		RetryOn5xx:     true,
	}

	client := httppool.NewRetryClient(http.DefaultClient, config)

	req, _ := http.NewRequest("GET", server.URL, nil)
	client.Do(req)

	mu.Lock()
	defer mu.Unlock()

	// Check that backoff increases
	if len(timestamps) >= 3 {
		gap1 := timestamps[1].Sub(timestamps[0])
		gap2 := timestamps[2].Sub(timestamps[1])
		// Second gap should be larger than first (exponential backoff)
		assert.True(t, gap2 >= gap1, "Expected exponential backoff, gap1=%v, gap2=%v", gap1, gap2)
	}
}

func TestRetryClient_NoRetryOnClientError(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	config := &httppool.RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 10 * time.Millisecond,
		RetryOn5xx:     true,
		RetryOn4xx:     false,
	}

	client := httppool.NewRetryClient(http.DefaultClient, config)

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := client.Do(req)

	require.NoError(t, err)
	defer resp.Body.Close()

	// Should not retry 4xx errors
	assert.Equal(t, int32(1), atomic.LoadInt32(&attempts))
}

func TestRetryClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := &httppool.RetryConfig{
		MaxRetries:     5,
		InitialBackoff: 100 * time.Millisecond,
	}

	client := httppool.NewRetryClient(http.DefaultClient, config)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
	_, err := client.Do(req)

	// Should be cancelled
	assert.Error(t, err)
}

func BenchmarkHTTPClientPool_GetClient(b *testing.B) {
	pool := httppool.NewHTTPClientPool(nil)
	defer pool.Close()

	hosts := []string{"host1.com", "host2.com", "host3.com", "host4.com", "host5.com"}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			pool.GetClient(hosts[i%len(hosts)])
			i++
		}
	})
}

func BenchmarkRetryClient_NoRetry(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := &httppool.RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 10 * time.Millisecond,
	}

	client := httppool.NewRetryClient(http.DefaultClient, config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", server.URL, nil)
		resp, _ := client.Do(req)
		if resp != nil {
			resp.Body.Close()
		}
	}
}
