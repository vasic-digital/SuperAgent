package discovery

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"dev.helix.agent/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTCPDiscoverer_Discover(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Suppress logs during tests
	discoverer := &tcpDiscoverer{logger: logger}

	t.Run("successful discovery", func(t *testing.T) {
		// Start a test TCP server
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		defer listener.Close()

		endpoint := &config.ServiceEndpoint{
			Host: "127.0.0.1",
			Port: fmt.Sprintf("%d", listener.Addr().(*net.TCPAddr).Port),
		}

		ctx := context.Background()
		discovered, err := discoverer.Discover(ctx, endpoint)
		assert.NoError(t, err)
		assert.True(t, discovered)
	})

	t.Run("failed discovery - unreachable port", func(t *testing.T) {
		// Find an unused port
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		port := listener.Addr().(*net.TCPAddr).Port
		listener.Close() // Close to make port free

		endpoint := &config.ServiceEndpoint{
			Host: "127.0.0.1",
			Port: fmt.Sprintf("%d", port),
		}

		ctx := context.Background()
		discovered, err := discoverer.Discover(ctx, endpoint)
		assert.NoError(t, err)
		assert.False(t, discovered)
	})

	t.Run("timeout", func(t *testing.T) {
		endpoint := &config.ServiceEndpoint{
			Host: "192.0.2.1", // Non-routable address for timeout
			Port: "9999",
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		discovered, err := discoverer.Discover(ctx, endpoint)
		assert.NoError(t, err)
		assert.False(t, discovered)
	})
}

func TestHTTPDiscoverer_Discover(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	discoverer := &httpDiscoverer{logger: logger}

	t.Run("successful discovery with health path", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		}))
		defer server.Close()

		endpoint := &config.ServiceEndpoint{
			Host:       server.URL[7:], // strip "http://"
			HealthPath: "/health",
			HealthType: "http",
		}

		ctx := context.Background()
		discovered, err := discoverer.Discover(ctx, endpoint)
		assert.NoError(t, err)
		assert.True(t, discovered)
	})

	t.Run("failed discovery - 404", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		endpoint := &config.ServiceEndpoint{
			Host:       server.URL[7:],
			HealthPath: "/health",
			HealthType: "http",
		}

		ctx := context.Background()
		discovered, err := discoverer.Discover(ctx, endpoint)
		assert.NoError(t, err)
		assert.False(t, discovered)
	})

	t.Run("no health path uses base URL", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		endpoint := &config.ServiceEndpoint{
			Host:       server.URL[7:],
			HealthType: "http",
		}

		ctx := context.Background()
		discovered, err := discoverer.Discover(ctx, endpoint)
		assert.NoError(t, err)
		assert.True(t, discovered)
	})
}

// MockDNSResolver implements dnsResolver for testing
type mockDNSResolver struct {
	srvRecords  map[string][]*net.SRV
	hostRecords map[string][]string
	srvError    error
	hostError   error
}

func (m *mockDNSResolver) LookupSRV(ctx context.Context, service, proto, name string) (string, []*net.SRV, error) {
	if m.srvError != nil {
		return "", nil, m.srvError
	}
	key := fmt.Sprintf("%s.%s.%s", service, proto, name)
	return "", m.srvRecords[key], nil
}

func (m *mockDNSResolver) LookupHost(ctx context.Context, host string) ([]string, error) {
	if m.hostError != nil {
		return nil, m.hostError
	}
	return m.hostRecords[host], nil
}

func TestDNSDiscoverer_Discover(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	t.Run("successful SRV discovery", func(t *testing.T) {
		mockResolver := &mockDNSResolver{
			srvRecords: map[string][]*net.SRV{
				".._postgresql._tcp.localhost": {
					{Target: "db.localhost", Port: 5432, Priority: 10, Weight: 5},
				},
			},
		}
		discoverer := &dnsDiscoverer{logger: logger, resolver: mockResolver}
		endpoint := &config.ServiceEndpoint{
			ServiceName: "postgresql",
			Host:        "localhost",
		}
		ctx := context.Background()
		discovered, err := discoverer.Discover(ctx, endpoint)
		assert.NoError(t, err)
		assert.True(t, discovered)
	})

	t.Run("successful A/AAAA discovery", func(t *testing.T) {
		mockResolver := &mockDNSResolver{
			srvRecords: map[string][]*net.SRV{},
			hostRecords: map[string][]string{
				"localhost": {"127.0.0.1", "::1"},
			},
		}
		discoverer := &dnsDiscoverer{logger: logger, resolver: mockResolver}
		endpoint := &config.ServiceEndpoint{
			ServiceName: "postgresql",
			Host:        "localhost",
		}
		ctx := context.Background()
		discovered, err := discoverer.Discover(ctx, endpoint)
		assert.NoError(t, err)
		assert.True(t, discovered)
	})

	t.Run("failed DNS resolution", func(t *testing.T) {
		mockResolver := &mockDNSResolver{
			srvRecords:  map[string][]*net.SRV{},
			hostRecords: map[string][]string{},
			hostError:   fmt.Errorf("NXDOMAIN"),
		}
		discoverer := &dnsDiscoverer{logger: logger, resolver: mockResolver}
		endpoint := &config.ServiceEndpoint{
			ServiceName: "postgresql",
			Host:        "unknown.example.com",
		}
		ctx := context.Background()
		discovered, err := discoverer.Discover(ctx, endpoint)
		assert.NoError(t, err) // DNS error is not returned, just false discovery
		assert.False(t, discovered)
	})

	t.Run("IP address host skips DNS", func(t *testing.T) {
		// We'll test that TCP discoverer is called (can't mock easily, skip for now)
		t.Skip("TODO: mock TCP discoverer")
	})
}

func TestCompositeDiscoverer(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	discoverer := NewDiscoverer(logger)

	t.Run("default method is TCP", func(t *testing.T) {
		endpoint := &config.ServiceEndpoint{
			Host: "127.0.0.1",
			Port: "9999", // unreachable
		}
		ctx := context.Background()
		discovered, err := discoverer.Discover(ctx, endpoint)
		assert.NoError(t, err)
		assert.False(t, discovered)
	})

	t.Run("method selection", func(t *testing.T) {
		// This test ensures the composite discoverer delegates correctly
		// We'll just test that it doesn't panic
		endpoint := &config.ServiceEndpoint{
			Host:            "127.0.0.1",
			Port:            "9999",
			DiscoveryMethod: "http",
			HealthPath:      "/health",
			HealthType:      "http",
		}
		ctx := context.Background()
		_, err := discoverer.Discover(ctx, endpoint)
		assert.NoError(t, err)
	})
}
