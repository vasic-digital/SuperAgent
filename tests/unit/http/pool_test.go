package http

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

	// Different host creates different entry
	pool.GetClient("other.com")

	// Check client count
	assert.Equal(t, 2, pool.ClientCount())
}

func TestHTTPClientPool_WithConfig(t *testing.T) {
	config := &httppool.PoolConfig{
		MaxIdleConns:          50,
		MaxConnsPerHost:       5,
		IdleConnTimeout:       30 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
		DialTimeout:           2 * time.Second,
		RetryCount:            2,
		RetryWaitMin:          50 * time.Millisecond,
		RetryWaitMax:          1 * time.Second,
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

	// Should have exactly 4 clients (one per unique host)
	assert.Equal(t, 4, pool.ClientCount())
}

func TestHTTPClientPool_RealRequest(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	}))
	defer server.Close()

	pool := httppool.NewHTTPClientPool(nil)
	defer pool.Close()

	ctx := context.Background()
	resp, err := pool.Get(ctx, server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "Hello, World!", string(body))
}

func TestHTTPClientPool_ConnectionReuse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	pool := httppool.NewHTTPClientPool(nil)
	defer pool.Close()

	ctx := context.Background()

	// Make multiple requests
	for i := 0; i < 5; i++ {
		resp, err := pool.Get(ctx, server.URL)
		require.NoError(t, err)
		resp.Body.Close()
	}

	// Should still have only one client
	assert.Equal(t, 1, pool.ClientCount())
}

func TestHTTPClientPool_Metrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	pool := httppool.NewHTTPClientPool(nil)
	defer pool.Close()

	ctx := context.Background()

	// Make some requests
	for i := 0; i < 5; i++ {
		resp, err := pool.Get(ctx, server.URL)
		require.NoError(t, err)
		resp.Body.Close()
	}

	metrics := pool.Metrics()
	assert.Equal(t, int64(5), metrics.TotalRequests)
	assert.Equal(t, int64(5), metrics.SuccessRequests)
	assert.Equal(t, int64(5), metrics.RequestCount)
	assert.True(t, metrics.AverageLatency() > 0)
}

func TestHTTPClientPool_Close(t *testing.T) {
	pool := httppool.NewHTTPClientPool(nil)

	pool.GetClient("host1.com")
	pool.GetClient("host2.com")

	assert.Equal(t, 2, pool.ClientCount())

	err := pool.Close()
	assert.NoError(t, err)

	assert.Equal(t, 0, pool.ClientCount())
}

func TestHTTPClientPool_Timeout(t *testing.T) {
	// Create a server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := &httppool.PoolConfig{
		DialTimeout:           1 * time.Second,
		ResponseHeaderTimeout: 500 * time.Millisecond,
		RetryCount:            0, // No retries
	}
	pool := httppool.NewHTTPClientPool(config)
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err := pool.Get(ctx, server.URL)
	assert.Error(t, err)
}

func TestHTTPClientPool_RetryOnError(t *testing.T) {
	var requestCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requestCount, 1)
		if count < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := &httppool.PoolConfig{
		RetryCount:   3,
		RetryWaitMin: 10 * time.Millisecond,
		RetryWaitMax: 100 * time.Millisecond,
	}
	pool := httppool.NewHTTPClientPool(config)
	defer pool.Close()

	ctx := context.Background()
	resp, err := pool.Get(ctx, server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Should have retried twice before succeeding
	assert.Equal(t, int32(3), atomic.LoadInt32(&requestCount))

	metrics := pool.Metrics()
	assert.Equal(t, int64(2), metrics.RetryCount)
}

func TestHTTPClientPool_PostJSON(t *testing.T) {
	var receivedBody string
	var receivedContentType string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContentType = r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	pool := httppool.NewHTTPClientPool(nil)
	defer pool.Close()

	ctx := context.Background()
	body := strings.NewReader(`{"key": "value"}`)
	resp, err := pool.PostJSON(ctx, server.URL, body)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, "application/json", receivedContentType)
	assert.Equal(t, `{"key": "value"}`, receivedBody)
}

func TestHTTPClientPool_HostClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != "test-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success"))
	}))
	defer server.Close()

	pool := httppool.NewHTTPClientPool(nil)
	defer pool.Close()

	client, err := httppool.NewHostClient(pool, server.URL)
	require.NoError(t, err)

	// Set default header
	client.SetHeader("X-API-Key", "test-key")

	ctx := context.Background()
	resp, err := client.Get(ctx, "/test")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHTTPClientPool_GetClientForURL(t *testing.T) {
	pool := httppool.NewHTTPClientPool(nil)
	defer pool.Close()

	client, err := pool.GetClientForURL("https://example.com/path/to/resource")
	require.NoError(t, err)
	require.NotNil(t, client)

	// Invalid URL should fail
	_, err = pool.GetClientForURL("://invalid")
	assert.Error(t, err)
}

func TestHTTPClientPool_GlobalPool(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Initialize global pool
	httppool.InitGlobalPool(nil)

	ctx := context.Background()
	resp, err := httppool.Get(ctx, server.URL)
	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHTTPClientPool_CloseIdleConnections(t *testing.T) {
	pool := httppool.NewHTTPClientPool(nil)
	defer pool.Close()

	pool.GetClient("example.com")

	// Should not panic
	pool.CloseIdleConnections()
}

func BenchmarkHTTPClientPool_GetClient(b *testing.B) {
	pool := httppool.NewHTTPClientPool(nil)
	defer pool.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.GetClient("example.com")
	}
}

func BenchmarkHTTPClientPool_ConcurrentGetClient(b *testing.B) {
	pool := httppool.NewHTTPClientPool(nil)
	defer pool.Close()

	hosts := []string{"host1.com", "host2.com", "host3.com", "host4.com"}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			pool.GetClient(hosts[i%len(hosts)])
			i++
		}
	})
}
