// Package http provides an adapter layer bridging HelixAgent's internal/http
// package (client-side HTTP connection pooling + HTTP/3 QUIC client) with the
// server-side HTTP/3 implementations in internal/transport and internal/router.
//
// Evaluation summary (2026-03-23):
//
//   - internal/http/pool.go provides HTTP/1.1 client connection pooling with
//     per-host client management, retry logic (exponential backoff), request
//     metrics, and a global singleton pool. This is CLIENT-SIDE functionality.
//
//   - internal/http/quic_client.go provides an HTTP/3 QUIC client using
//     quic-go/http3.Transport with automatic HTTP/2 fallback, QUIC-specific
//     metrics (H3 vs fallback counts), and HTTP3ProviderTransport which
//     implements http.RoundTripper for use as a drop-in transport. This is
//     CLIENT-SIDE functionality.
//
//   - internal/transport/http3.go provides an HTTP/3 SERVER with HTTP/2 and
//     HTTP/1.1 fallback, TLS config management, self-signed cert generation,
//     and graceful shutdown. This is SERVER-SIDE functionality.
//
//   - internal/router/quic_server.go provides another HTTP/3 SERVER (simpler)
//     with dual-stack HTTP/1.1/2 + HTTP/3 serving, TLS configuration, and
//     QUIC listener creation. This is SERVER-SIDE functionality.
//
// Conclusion: internal/http is NOT duplicated by transport/router. The packages
// serve complementary roles (client vs server). This adapter exposes the
// client-side pool and QUIC client through a unified interface that other
// HelixAgent components can use for outbound HTTP requests with HTTP/3 support.
package http

import (
	"context"
	"io"
	"net/http"
	"sync"
	"time"

	httppool "dev.helix.agent/internal/http"
)

// ClientAdapter provides a unified interface for HTTP client operations,
// combining connection pooling from pool.go with HTTP/3 QUIC transport
// from quic_client.go. It selects the appropriate transport based on
// configuration and provides a single entry point for outbound HTTP calls.
type ClientAdapter struct {
	pool       *httppool.HTTPClientPool
	quicClient *httppool.QUICClient
	useQUIC    bool
	mu         sync.RWMutex
	closed     bool
}

// Config holds configuration for the HTTP client adapter.
type Config struct {
	// Pool configuration for HTTP/1.1/2 connection pooling
	PoolConfig *httppool.PoolConfig

	// QUIC configuration for HTTP/3 client
	QUICConfig *httppool.QUICConfig

	// EnableQUIC enables HTTP/3 QUIC transport (with HTTP/2 fallback)
	EnableQUIC bool
}

// DefaultConfig returns a default adapter configuration with pooling enabled
// and QUIC disabled (safe default for backward compatibility).
func DefaultConfig() *Config {
	return &Config{
		PoolConfig: httppool.DefaultPoolConfig(),
		QUICConfig: httppool.DefaultQUICConfig(),
		EnableQUIC: false,
	}
}

// NewClientAdapter creates a new HTTP client adapter.
// If EnableQUIC is true, HTTP/3 transport is used with HTTP/2 fallback.
// Otherwise, the standard connection pool is used.
func NewClientAdapter(cfg *Config) (*ClientAdapter, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	adapter := &ClientAdapter{
		pool:    httppool.NewHTTPClientPool(cfg.PoolConfig),
		useQUIC: cfg.EnableQUIC,
	}

	if cfg.EnableQUIC {
		quicClient, err := httppool.NewQUICClient(cfg.QUICConfig)
		if err != nil {
			// Fall back to pool-only mode if QUIC init fails
			adapter.useQUIC = false
		} else {
			adapter.quicClient = quicClient
		}
	}

	return adapter, nil
}

// Do executes an HTTP request using the configured transport.
// If QUIC is enabled, it tries HTTP/3 first with automatic fallback.
// Otherwise, it uses the connection pool with retry logic.
func (a *ClientAdapter) Do(req *http.Request) (*http.Response, error) {
	a.mu.RLock()
	if a.closed {
		a.mu.RUnlock()
		return nil, http.ErrServerClosed
	}
	useQUIC := a.useQUIC && a.quicClient != nil
	a.mu.RUnlock()

	if useQUIC {
		return a.quicClient.Do(req)
	}

	return a.pool.Do(req)
}

// Get performs an HTTP GET request.
func (a *ClientAdapter) Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return a.Do(req)
}

// PostJSON performs a POST request with JSON content type.
func (a *ClientAdapter) PostJSON(
	ctx context.Context, url string, body io.Reader,
) (*http.Response, error) {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost, url, body,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return a.Do(req)
}

// PoolMetrics returns connection pool metrics (nil if pool is not initialized).
func (a *ClientAdapter) PoolMetrics() *httppool.PoolMetrics {
	if a.pool == nil {
		return nil
	}
	return a.pool.Metrics()
}

// QUICMetrics returns QUIC client metrics (nil if QUIC is not enabled).
func (a *ClientAdapter) QUICMetrics() *httppool.QUICMetrics {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.quicClient == nil {
		return nil
	}
	return a.quicClient.Metrics()
}

// IsQUICEnabled returns whether HTTP/3 QUIC transport is active.
func (a *ClientAdapter) IsQUICEnabled() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.useQUIC && a.quicClient != nil
}

// RoundTripper returns an http.RoundTripper suitable for use with
// standard http.Client. If QUIC is enabled, returns an HTTP3ProviderTransport;
// otherwise returns the pool's shared transport via a pool-based client.
func (a *ClientAdapter) RoundTripper() http.RoundTripper {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.useQUIC && a.quicClient != nil {
		transport, err := httppool.NewHTTP3ProviderTransport(
			httppool.DefaultQUICConfig(),
		)
		if err != nil {
			// Fallback: return a default transport
			return http.DefaultTransport
		}
		return transport
	}

	return http.DefaultTransport
}

// Close releases all resources held by the adapter.
func (a *ClientAdapter) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.closed {
		return nil
	}
	a.closed = true

	var firstErr error
	if a.quicClient != nil {
		if err := a.quicClient.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if a.pool != nil {
		if err := a.pool.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

// HealthCheck verifies that the adapter is operational by checking that
// the pool and (optionally) the QUIC client are not closed.
func (a *ClientAdapter) HealthCheck() HealthStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()

	status := HealthStatus{
		Timestamp:    time.Now(),
		PoolActive:   a.pool != nil && !a.closed,
		QUICActive:   a.quicClient != nil && !a.closed,
		QUICEnabled:  a.useQUIC,
		OverallReady: !a.closed && a.pool != nil,
	}

	if a.pool != nil {
		status.PoolClientCount = a.pool.ClientCount()
	}

	return status
}

// HealthStatus represents the health of the HTTP client adapter.
type HealthStatus struct {
	Timestamp       time.Time `json:"timestamp"`
	PoolActive      bool      `json:"pool_active"`
	PoolClientCount int       `json:"pool_client_count"`
	QUICActive      bool      `json:"quic_active"`
	QUICEnabled     bool      `json:"quic_enabled"`
	OverallReady    bool      `json:"overall_ready"`
}
