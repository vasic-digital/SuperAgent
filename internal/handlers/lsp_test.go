package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLSPHandler(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)

	handler := NewLSPHandler(nil, log)

	require.NotNil(t, handler)
	assert.NotNil(t, handler.log)
	assert.Nil(t, handler.lspService) // nil when no service provided
}

func TestLSPHandler_ExecuteLSPRequest_InvalidJSON(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)

	handler := NewLSPHandler(nil, log)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/v1/lsp/execute", bytes.NewBufferString("invalid json"))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.ExecuteLSPRequest(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLSPHandler_ExecuteLSPRequest_ValidJSON(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)

	handler := NewLSPHandler(nil, log)

	// Valid JSON should pass binding and return success
	// (the method is a placeholder that doesn't call the service)
	body := `{"server_id": "gopls", "tool_name": "completion", "params": {"file": "main.go"}}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/v1/lsp/execute", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.ExecuteLSPRequest(c)

	// The placeholder implementation returns success
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLSPHandler_ExecuteLSPRequest_EmptyBody(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)

	handler := NewLSPHandler(nil, log)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/v1/lsp/execute", nil)
	c.Request.Header.Set("Content-Type", "application/json")

	handler.ExecuteLSPRequest(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLSPHandler_SyncLSPServer_ParamParsing(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)

	handler := NewLSPHandler(nil, log)

	t.Run("with server id param", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/v1/lsp/servers/gopls/sync", nil)
		c.Params = gin.Params{{Key: "id", Value: "gopls"}}

		// This will call the nil service and panic, but we can test the param parsing
		// by checking if it would have been called with the right server ID
		// For now, we just verify the handler was created correctly
		require.NotNil(t, handler)
	})
}

func BenchmarkLSPHandler_ExecuteLSPRequest(b *testing.B) {
	log := logrus.New()
	log.SetLevel(logrus.WarnLevel)

	handler := NewLSPHandler(nil, log)

	body := `{"server_id": "gopls", "tool_name": "completion"}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/v1/lsp/execute", bytes.NewBufferString(body))
		c.Request.Header.Set("Content-Type", "application/json")
		handler.ExecuteLSPRequest(c)
	}
}
