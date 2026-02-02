package toon

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultTransportConfig(t *testing.T) {
	cfg := DefaultTransportConfig()

	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.HTTPClient)
	assert.NotNil(t, cfg.Headers)
	assert.Equal(t, CompressionStandard, cfg.Compression)
}

func TestNewTransport(t *testing.T) {
	t.Run("with default config", func(t *testing.T) {
		transport := NewTransport(nil)

		assert.NotNil(t, transport)
		assert.NotNil(t, transport.encoder)
		assert.NotNil(t, transport.decoder)
		assert.NotNil(t, transport.httpClient)
		assert.NotNil(t, transport.metrics)
	})

	t.Run("with custom config", func(t *testing.T) {
		cfg := &TransportConfig{
			BaseURL:     "http://localhost:8080",
			HTTPClient:  &http.Client{},
			Headers:     map[string]string{"X-Custom": "value"},
			Compression: CompressionAggressive,
		}
		transport := NewTransport(cfg)

		assert.NotNil(t, transport)
		assert.Equal(t, "http://localhost:8080", transport.baseURL)
		assert.Equal(t, "value", transport.headers["X-Custom"])
	})
}

func TestTransport_SetHeader(t *testing.T) {
	transport := NewTransport(nil)

	transport.SetHeader("Authorization", "Bearer token")

	transport.mu.RLock()
	defer transport.mu.RUnlock()
	assert.Equal(t, "Bearer token", transport.headers["Authorization"])
}

func TestTransport_SetBaseURL(t *testing.T) {
	transport := NewTransport(nil)

	transport.SetBaseURL("http://api.example.com")

	assert.Equal(t, "http://api.example.com", transport.baseURL)
}

func TestTransport_SetCompression(t *testing.T) {
	transport := NewTransport(nil)

	transport.SetCompression(CompressionMinimal)

	assert.Equal(t, CompressionMinimal, transport.encoder.GetCompressionLevel())
}

func TestTransport_GetMetrics(t *testing.T) {
	transport := NewTransport(nil)

	// Initially all metrics should be zero
	metrics := transport.GetMetrics()

	assert.Equal(t, int64(0), metrics.RequestCount)
	assert.Equal(t, int64(0), metrics.BytesSent)
	assert.Equal(t, int64(0), metrics.BytesReceived)
}

func TestTransport_AverageCompressionRatio(t *testing.T) {
	t.Run("with no requests", func(t *testing.T) {
		transport := NewTransport(nil)
		ratio := transport.AverageCompressionRatio()
		assert.Equal(t, float64(0), ratio)
	})

	t.Run("with requests", func(t *testing.T) {
		transport := NewTransport(nil)

		// Simulate metrics
		transport.metrics.mu.Lock()
		transport.metrics.RequestCount = 2
		transport.metrics.CompressionRatioSum = 1.5
		transport.metrics.mu.Unlock()

		ratio := transport.AverageCompressionRatio()
		assert.Equal(t, 0.75, ratio)
	})
}

func TestTransport_Do(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		assert.Equal(t, "application/toon+json", r.Header.Get("Content-Type"))
		assert.Equal(t, "application/toon+json", r.Header.Get("Accept"))

		// Read and respond
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer server.Close()

	transport := NewTransport(&TransportConfig{
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
	})

	ctx := context.Background()
	req := &Request{
		Method: http.MethodPost,
		Path:   "/test",
		Body:   map[string]string{"name": "test"},
	}

	resp, err := transport.Do(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestTransport_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	transport := NewTransport(&TransportConfig{
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
	})

	resp, err := transport.Get(context.Background(), "/status")

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestTransport_Post(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"123"}`))
	}))
	defer server.Close()

	transport := NewTransport(&TransportConfig{
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
	})

	resp, err := transport.Post(context.Background(), "/create", map[string]string{"name": "test"})

	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestTransport_Put(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		_, _ = w.Write([]byte(`{"updated":true}`))
	}))
	defer server.Close()

	transport := NewTransport(&TransportConfig{
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
	})

	resp, err := transport.Put(context.Background(), "/update/123", map[string]string{"name": "updated"})

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestTransport_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	transport := NewTransport(&TransportConfig{
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
	})

	resp, err := transport.Delete(context.Background(), "/delete/123")

	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestTransport_DecodeResponse(t *testing.T) {
	transport := NewTransport(nil)

	// TOON encoded data
	encodedData := []byte(`{"i":"123","n":"test"}`)
	resp := &Response{
		StatusCode: 200,
		Body:       encodedData,
	}

	var result map[string]interface{}
	err := transport.DecodeResponse(resp, &result)

	require.NoError(t, err)
	assert.Equal(t, "123", result["id"])
	assert.Equal(t, "test", result["name"])
}

func TestTransport_WithCustomHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer token123", r.Header.Get("Authorization"))
		assert.Equal(t, "custom-value", r.Header.Get("X-Custom-Header"))
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	transport := NewTransport(&TransportConfig{
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
		Headers: map[string]string{
			"Authorization": "Bearer token123",
		},
	})

	req := &Request{
		Method: http.MethodGet,
		Path:   "/test",
		Headers: map[string]string{
			"X-Custom-Header": "custom-value",
		},
	}

	_, err := transport.Do(context.Background(), req)
	require.NoError(t, err)
}

func TestTransport_MetricsUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	transport := NewTransport(&TransportConfig{
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
	})

	// Make a request with body
	_, err := transport.Post(context.Background(), "/test", map[string]string{
		"name":        "test",
		"description": "A longer description to test compression",
	})
	require.NoError(t, err)

	metrics := transport.GetMetrics()
	assert.Equal(t, int64(1), metrics.RequestCount)
	assert.Greater(t, metrics.BytesSent, int64(0))
	assert.Greater(t, metrics.BytesReceived, int64(0))
}

func TestNewMiddleware(t *testing.T) {
	middleware := NewMiddleware()

	assert.NotNil(t, middleware)
	assert.NotNil(t, middleware.encoder)
	assert.NotNil(t, middleware.decoder)
}

func TestMiddleware_Handler_RegularRequest(t *testing.T) {
	middleware := NewMiddleware()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	wrapped := middleware.Handler(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, `{"status":"ok"}`, rr.Body.String())
}

func TestMiddleware_Handler_TOONRequest(t *testing.T) {
	middleware := NewMiddleware()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read the expanded JSON body
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		_ = json.Unmarshal(body, &data)

		// Return response
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"received_id": data["id"],
		})
	})

	wrapped := middleware.Handler(handler)

	// Create TOON encoded request body
	encoder := NewEncoder(&EncoderOptions{Compression: CompressionMinimal})
	toonBody, _ := encoder.Encode(map[string]string{"id": "123", "name": "test"})

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(toonBody))
	req.Header.Set("Content-Type", "application/toon+json")
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestMiddleware_Handler_TOONResponse(t *testing.T) {
	middleware := NewMiddleware()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"123","name":"test","status":"healthy"}`))
	})

	wrapped := middleware.Handler(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept", "application/toon+json")
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/toon+json", rr.Header().Get("Content-Type"))

	// Decode the response
	decoder := NewDecoder(nil)
	var result map[string]interface{}
	err := decoder.Decode(rr.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "123", result["id"])
}

func TestMiddleware_Handler_InvalidTOON(t *testing.T) {
	middleware := NewMiddleware()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	wrapped := middleware.Handler(handler)

	// Invalid TOON body
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/toon+json")
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestResponseWrapper_Write(t *testing.T) {
	underlying := httptest.NewRecorder()
	wrapper := &responseWrapper{
		ResponseWriter: underlying,
		buf:            &bytes.Buffer{},
	}

	n, err := wrapper.Write([]byte("test data"))

	assert.NoError(t, err)
	assert.Equal(t, 9, n)
	assert.Equal(t, "test data", wrapper.buf.String())
}

func TestResponseWrapper_WriteHeader(t *testing.T) {
	underlying := httptest.NewRecorder()
	wrapper := &responseWrapper{
		ResponseWriter: underlying,
		buf:            &bytes.Buffer{},
	}

	wrapper.WriteHeader(http.StatusCreated)

	assert.Equal(t, http.StatusCreated, wrapper.statusCode)
}

func TestTransport_Concurrent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	transport := NewTransport(&TransportConfig{
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
	})

	// Run concurrent requests
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := transport.Get(context.Background(), "/test")
			assert.NoError(t, err)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	metrics := transport.GetMetrics()
	assert.Equal(t, int64(10), metrics.RequestCount)
}

func TestTransport_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response - this won't block because context is cancelled
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	transport := NewTransport(&TransportConfig{
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := transport.Get(ctx, "/test")
	assert.Error(t, err)
}

func TestRequest_Fields(t *testing.T) {
	req := &Request{
		Method: http.MethodPost,
		Path:   "/api/v1/test",
		Body:   map[string]string{"key": "value"},
		Headers: map[string]string{
			"X-Request-ID": "req-123",
		},
	}

	assert.Equal(t, http.MethodPost, req.Method)
	assert.Equal(t, "/api/v1/test", req.Path)
	assert.NotNil(t, req.Body)
	assert.Equal(t, "req-123", req.Headers["X-Request-ID"])
}

func TestResponse_Fields(t *testing.T) {
	resp := &Response{
		StatusCode: http.StatusOK,
		Body:       []byte(`{"result":"success"}`),
		Headers: http.Header{
			"Content-Type": []string{"application/json"},
		},
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, resp.Body)
	assert.Equal(t, "application/json", resp.Headers.Get("Content-Type"))
}

func TestTransportMetrics_Fields(t *testing.T) {
	metrics := &TransportMetrics{
		RequestCount:        100,
		BytesSent:           1024,
		BytesReceived:       2048,
		BytesSaved:          512,
		TokensSaved:         128,
		CompressionRatioSum: 0.75,
	}

	assert.Equal(t, int64(100), metrics.RequestCount)
	assert.Equal(t, int64(1024), metrics.BytesSent)
	assert.Equal(t, int64(2048), metrics.BytesReceived)
	assert.Equal(t, int64(512), metrics.BytesSaved)
	assert.Equal(t, int64(128), metrics.TokensSaved)
	assert.Equal(t, 0.75, metrics.CompressionRatioSum)
}
