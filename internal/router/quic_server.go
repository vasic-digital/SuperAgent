package router

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/sirupsen/logrus"
)

// QUICServer represents a dual-stack HTTP/1.1/2 and HTTP/3 server
type QUICServer struct {
	httpServer  *http.Server
	http3Server *http3.Server
	logger      *logrus.Logger
	addr        string
	certFile    string
	keyFile     string
}

// NewQUICServer creates a new dual-stack HTTP server with HTTP/3 support
func NewQUICServer(addr string, handler http.Handler, logger *logrus.Logger, certFile, keyFile string) *QUICServer {
	// Create standard HTTP/1.1/2 server
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 300 * time.Second, // 5 minutes for SSE streaming support
		IdleTimeout:  120 * time.Second,
		TLSConfig: &tls.Config{
			NextProtos: []string{"h2", "http/1.1"}, // Support HTTP/2 and HTTP/1.1
		},
	}

	// Create HTTP/3 server
	http3Server := &http3.Server{
		Addr:      addr,
		Handler:   handler,
		TLSConfig: httpServer.TLSConfig,
	}

	return &QUICServer{
		httpServer:  httpServer,
		http3Server: http3Server,
		logger:      logger,
		addr:        addr,
		certFile:    certFile,
		keyFile:     keyFile,
	}
}

// Start starts both HTTP/1.1/2 and HTTP/3 servers
func (s *QUICServer) Start() error {
	// Start HTTP/3 server in goroutine
	go func() {
		s.logger.WithFields(logrus.Fields{
			"addr":     s.addr,
			"protocol": "HTTP/3",
		}).Info("Starting HTTP/3 server")

		if err := s.http3Server.ListenAndServeTLS(s.certFile, s.keyFile); err != nil && err != http.ErrServerClosed {
			s.logger.WithError(err).Error("HTTP/3 server failed")
		}
	}()

	// Start HTTP/1.1/2 server (will also handle TLS)
	s.logger.WithFields(logrus.Fields{
		"addr":     s.addr,
		"protocol": "HTTP/1.1/2",
	}).Info("Starting HTTP/1.1/2 server")

	return s.httpServer.ListenAndServeTLS(s.certFile, s.keyFile)
}

// StartInsecure starts servers without TLS (for testing)
func (s *QUICServer) StartInsecure() error {
	// Note: HTTP/3 requires TLS, so we can't start it in insecure mode
	// We'll only start HTTP/1.1 server without TLS
	s.logger.WithFields(logrus.Fields{
		"addr":     s.addr,
		"protocol": "HTTP/1.1 (insecure)",
	}).Warn("Starting insecure HTTP server (HTTP/3 requires TLS)")

	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down both servers
func (s *QUICServer) Shutdown(ctx context.Context) error {
	var errs []error

	// Shutdown HTTP/3 server
	if err := s.http3Server.Close(); err != nil {
		errs = append(errs, fmt.Errorf("HTTP/3 shutdown error: %w", err))
	}

	// Shutdown HTTP/1.1/2 server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("HTTP/1.1/2 shutdown error: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("server shutdown errors: %v", errs)
	}

	return nil
}

// SetTLSConfig updates TLS configuration for both servers
func (s *QUICServer) SetTLSConfig(tlsConfig *tls.Config) {
	s.httpServer.TLSConfig = tlsConfig
	s.http3Server.TLSConfig = tlsConfig

	// Add HTTP/3 ALPN protocol
	if s.httpServer.TLSConfig != nil {
		s.httpServer.TLSConfig.NextProtos = append(s.httpServer.TLSConfig.NextProtos, "h3")
	}
}

// EnableHTTP3Support configures the server for optimal HTTP/3 support
func (s *QUICServer) EnableHTTP3Support() {
	if s.httpServer.TLSConfig == nil {
		s.httpServer.TLSConfig = &tls.Config{}
	}

	// Configure TLS for HTTP/3
	s.httpServer.TLSConfig.NextProtos = []string{"h3", "h2", "http/1.1"}
	s.httpServer.TLSConfig.MinVersion = tls.VersionTLS13 // HTTP/3 requires TLS 1.3

	// Update HTTP/3 server config
	s.http3Server.TLSConfig = s.httpServer.TLSConfig
}

// QUICConfig holds configuration for QUIC/HTTP3
type QUICConfig struct {
	// Addr is the address to listen on
	Addr string
	// CertFile is the path to TLS certificate file
	CertFile string
	// KeyFile is the path to TLS private key file
	KeyFile string
	// EnableHTTP3 enables HTTP/3 support
	EnableHTTP3 bool
	// QUIC specific settings
	MaxIdleTimeout        time.Duration
	KeepAlivePeriod       time.Duration
	MaxIncomingStreams    int64
	MaxIncomingUniStreams int64
}

// DefaultQUICConfig returns default QUIC configuration
func DefaultQUICConfig(addr, certFile, keyFile string) *QUICConfig {
	return &QUICConfig{
		Addr:                  addr,
		CertFile:              certFile,
		KeyFile:               keyFile,
		EnableHTTP3:           true,
		MaxIdleTimeout:        30 * time.Second,
		KeepAlivePeriod:       10 * time.Second,
		MaxIncomingStreams:    100,
		MaxIncomingUniStreams: 100,
	}
}

// CreateQUICListener creates a QUIC listener with the given configuration
func CreateQUICListener(config *QUICConfig) (*quic.Listener, error) {
	tlsCert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS certificate: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"h3"},
		MinVersion:   tls.VersionTLS13,
	}

	quicConfig := &quic.Config{
		MaxIdleTimeout:        config.MaxIdleTimeout,
		KeepAlivePeriod:       config.KeepAlivePeriod,
		MaxIncomingStreams:    config.MaxIncomingStreams,
		MaxIncomingUniStreams: config.MaxIncomingUniStreams,
	}

	return quic.ListenAddr(config.Addr, tlsConfig, quicConfig)
}

// SupportsHTTP3 checks if HTTP/3 is supported by the current configuration
func SupportsHTTP3(certFile, keyFile string) bool {
	_, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return false
	}

	// Check if quic-go is available
	// This is a compile-time check
	return true
}
