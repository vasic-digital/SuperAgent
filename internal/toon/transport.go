// Package toon provides Token-Optimized Object Notation (TOON) encoding.
package toon

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

// Transport provides TOON-encoded HTTP transport.
type Transport struct {
	encoder    *Encoder
	decoder    *Decoder
	httpClient *http.Client
	baseURL    string
	headers    map[string]string
	mu         sync.RWMutex
	metrics    *TransportMetrics
}

// TransportMetrics holds metrics for the transport layer.
type TransportMetrics struct {
	RequestCount        int64
	BytesSent           int64
	BytesReceived       int64
	BytesSaved          int64
	TokensSaved         int64
	CompressionRatioSum float64
	mu                  sync.RWMutex
}

// TransportConfig configures the TOON transport.
type TransportConfig struct {
	BaseURL     string
	HTTPClient  *http.Client
	Headers     map[string]string
	Compression CompressionLevel
}

// DefaultTransportConfig returns default transport configuration.
func DefaultTransportConfig() *TransportConfig {
	return &TransportConfig{
		HTTPClient:  http.DefaultClient,
		Headers:     make(map[string]string),
		Compression: CompressionStandard,
	}
}

// NewTransport creates a new TOON transport.
func NewTransport(cfg *TransportConfig) *Transport {
	if cfg == nil {
		cfg = DefaultTransportConfig()
	}

	opts := &EncoderOptions{
		Compression: cfg.Compression,
	}

	return &Transport{
		encoder:    NewEncoder(opts),
		decoder:    NewDecoder(opts),
		httpClient: cfg.HTTPClient,
		baseURL:    cfg.BaseURL,
		headers:    cfg.Headers,
		metrics:    &TransportMetrics{},
	}
}

// Request represents a TOON-encoded request.
type Request struct {
	Method  string
	Path    string
	Body    interface{}
	Headers map[string]string
}

// Response represents a TOON-encoded response.
type Response struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

// Do executes a TOON-encoded request.
func (t *Transport) Do(ctx context.Context, req *Request) (*Response, error) {
	// Encode request body
	var body []byte
	var originalSize int
	if req.Body != nil {
		var err error
		// Get original size for metrics
		jsonBody, _ := json.Marshal(req.Body)
		originalSize = len(jsonBody)

		body, err = t.encoder.Encode(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to encode request body: %w", err)
		}
	}

	// Build HTTP request
	url := t.baseURL + req.Path
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/toon+json")
	httpReq.Header.Set("Accept", "application/toon+json")
	t.mu.RLock()
	for k, v := range t.headers {
		httpReq.Header.Set(k, v)
	}
	t.mu.RUnlock()
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	// Execute request
	resp, err := t.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Update metrics
	t.metrics.mu.Lock()
	t.metrics.RequestCount++
	t.metrics.BytesSent += int64(len(body))
	t.metrics.BytesReceived += int64(len(respBody))
	if originalSize > 0 {
		saved := int64(originalSize - len(body))
		if saved > 0 {
			t.metrics.BytesSaved += saved
			t.metrics.TokensSaved += saved / 4 // Rough token estimate
			t.metrics.CompressionRatioSum += float64(len(body)) / float64(originalSize)
		}
	}
	t.metrics.mu.Unlock()

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Headers:    resp.Header,
	}, nil
}

// Get performs a GET request.
func (t *Transport) Get(ctx context.Context, path string) (*Response, error) {
	return t.Do(ctx, &Request{
		Method: http.MethodGet,
		Path:   path,
	})
}

// Post performs a POST request with TOON-encoded body.
func (t *Transport) Post(ctx context.Context, path string, body interface{}) (*Response, error) {
	return t.Do(ctx, &Request{
		Method: http.MethodPost,
		Path:   path,
		Body:   body,
	})
}

// Put performs a PUT request with TOON-encoded body.
func (t *Transport) Put(ctx context.Context, path string, body interface{}) (*Response, error) {
	return t.Do(ctx, &Request{
		Method: http.MethodPut,
		Path:   path,
		Body:   body,
	})
}

// Delete performs a DELETE request.
func (t *Transport) Delete(ctx context.Context, path string) (*Response, error) {
	return t.Do(ctx, &Request{
		Method: http.MethodDelete,
		Path:   path,
	})
}

// DecodeResponse decodes a TOON response into the target.
func (t *Transport) DecodeResponse(resp *Response, v interface{}) error {
	return t.decoder.Decode(resp.Body, v)
}

// SetHeader sets a default header.
func (t *Transport) SetHeader(key, value string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.headers[key] = value
}

// SetBaseURL sets the base URL.
func (t *Transport) SetBaseURL(url string) {
	t.baseURL = url
}

// SetCompression sets the compression level.
func (t *Transport) SetCompression(level CompressionLevel) {
	t.encoder.SetCompression(level)
}

// GetMetrics returns transport metrics.
func (t *Transport) GetMetrics() *TransportMetrics {
	t.metrics.mu.RLock()
	defer t.metrics.mu.RUnlock()

	return &TransportMetrics{
		RequestCount:        t.metrics.RequestCount,
		BytesSent:           t.metrics.BytesSent,
		BytesReceived:       t.metrics.BytesReceived,
		BytesSaved:          t.metrics.BytesSaved,
		TokensSaved:         t.metrics.TokensSaved,
		CompressionRatioSum: t.metrics.CompressionRatioSum,
	}
}

// AverageCompressionRatio returns the average compression ratio.
func (t *Transport) AverageCompressionRatio() float64 {
	t.metrics.mu.RLock()
	defer t.metrics.mu.RUnlock()

	if t.metrics.RequestCount == 0 {
		return 0
	}
	return t.metrics.CompressionRatioSum / float64(t.metrics.RequestCount)
}

// Middleware provides TOON encoding middleware for HTTP servers.
type Middleware struct {
	encoder *Encoder
	decoder *Decoder
}

// NewMiddleware creates a new TOON middleware.
func NewMiddleware() *Middleware {
	opts := DefaultEncoderOptions()
	return &Middleware{
		encoder: NewEncoder(opts),
		decoder: NewDecoder(opts),
	}
}

// Handler wraps an HTTP handler with TOON encoding.
func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for TOON content type
		contentType := r.Header.Get("Content-Type")
		acceptTOON := r.Header.Get("Accept") == "application/toon+json"

		if contentType == "application/toon+json" {
			// Decode request body
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Failed to read body", http.StatusBadRequest)
				return
			}
			r.Body.Close()

			// Expand TOON to JSON
			expanded, err := m.expandToJSON(body)
			if err != nil {
				http.Error(w, "Failed to decode TOON", http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(expanded))
			r.Header.Set("Content-Type", "application/json")
			r.ContentLength = int64(len(expanded))
		}

		if acceptTOON {
			// Wrap response writer to capture response
			rw := &responseWrapper{
				ResponseWriter: w,
				buf:            &bytes.Buffer{},
			}
			next.ServeHTTP(rw, r)

			// Encode response as TOON
			encoded, err := m.encoder.Encode(json.RawMessage(rw.buf.Bytes()))
			if err != nil {
				http.Error(w, "Failed to encode TOON", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/toon+json")
			w.Write(encoded)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

// expandToJSON expands TOON data to standard JSON.
func (m *Middleware) expandToJSON(data []byte) ([]byte, error) {
	var obj interface{}
	if err := m.decoder.Decode(data, &obj); err != nil {
		return nil, err
	}
	return json.Marshal(obj)
}

// responseWrapper captures the response body.
type responseWrapper struct {
	http.ResponseWriter
	buf        *bytes.Buffer
	statusCode int
}

func (rw *responseWrapper) Write(b []byte) (int, error) {
	return rw.buf.Write(b)
}

func (rw *responseWrapper) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
}
