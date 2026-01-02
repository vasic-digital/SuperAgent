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

func TestNewEmbeddingHandler(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)

	handler := NewEmbeddingHandler(nil, log)

	require.NotNil(t, handler)
	assert.NotNil(t, handler.log)
	assert.Nil(t, handler.embeddingManager) // nil when no manager provided
}

func TestEmbeddingHandler_GenerateEmbeddings_InvalidJSON(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)

	// Create handler with nil manager - will panic if we call the service
	// but we can test JSON parsing
	handler := NewEmbeddingHandler(nil, log)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/v1/embeddings/generate", bytes.NewBufferString("invalid json"))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.GenerateEmbeddings(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEmbeddingHandler_VectorSearch_InvalidJSON(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)

	handler := NewEmbeddingHandler(nil, log)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/v1/embeddings/search", bytes.NewBufferString("{not valid json"))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.VectorSearch(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEmbeddingHandler_IndexDocument_InvalidJSON(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)

	handler := NewEmbeddingHandler(nil, log)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/v1/embeddings/index", bytes.NewBufferString("bad"))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.IndexDocument(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEmbeddingHandler_BatchIndexDocuments_InvalidJSON(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)

	handler := NewEmbeddingHandler(nil, log)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/v1/embeddings/batch-index", bytes.NewBufferString("{invalid"))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.BatchIndexDocuments(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEmbeddingHandler_ConfigureProvider_InvalidJSON(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)

	handler := NewEmbeddingHandler(nil, log)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("PUT", "/v1/embeddings/provider", bytes.NewBufferString("not json"))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.ConfigureProvider(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEmbeddingHandler_SimilaritySearch_InvalidJSON(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)

	handler := NewEmbeddingHandler(nil, log)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/v1/embeddings/similarity", bytes.NewBufferString("{broken"))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.SimilaritySearch(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEmbeddingHandler_EmptyBody(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)

	handler := NewEmbeddingHandler(nil, log)

	t.Run("generate embeddings empty", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/v1/embeddings/generate", nil)
		c.Request.Header.Set("Content-Type", "application/json")

		handler.GenerateEmbeddings(c)

		// Empty body should still fail to bind
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("vector search empty", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/v1/embeddings/search", nil)
		c.Request.Header.Set("Content-Type", "application/json")

		handler.VectorSearch(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
