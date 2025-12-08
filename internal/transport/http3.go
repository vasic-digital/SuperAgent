package transport

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HTTP3Server represents an HTTP3/Quic server with fallback to HTTP2.
type HTTP3Server struct {
	httpServer *http.Server
	// TODO: Add Quic server implementation
}

// NewHTTP3Server creates a new HTTP3 server with HTTP2 fallback.
func NewHTTP3Server(handler *gin.Engine, addr string) *HTTP3Server {
	return &HTTP3Server{
		httpServer: &http.Server{
			Addr:    addr,
			Handler: handler,
		},
	}
}

// Start starts the HTTP3 server with HTTP2 fallback.
func (s *HTTP3Server) Start() error {
	// TODO: Implement HTTP3/Quic server
	// For now, fallback to HTTP2
	return s.httpServer.ListenAndServe()
}

// Stop stops the server gracefully.
func (s *HTTP3Server) Stop() error {
	return s.httpServer.Close()
}
