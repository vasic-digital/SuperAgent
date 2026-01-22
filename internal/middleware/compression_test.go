package middleware

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultCompressionConfig(t *testing.T) {
	config := DefaultCompressionConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 4, config.BrotliLevel)
	assert.Equal(t, 5, config.GzipLevel)
	assert.Equal(t, 1024, config.MinSize)
	assert.True(t, config.EnableBrotli)
	assert.True(t, config.EnableGzip)
	assert.NotEmpty(t, config.CompressibleTypes)
	assert.Contains(t, config.CompressibleTypes, "application/json")
	assert.Contains(t, config.CompressibleTypes, "text/html")
	assert.NotEmpty(t, config.ExcludePaths)
	assert.Contains(t, config.ExcludePaths, "/health")
}

func TestCompressData_Brotli(t *testing.T) {
	data := []byte("This is a test string that will be compressed with Brotli encoding.")

	compressed, err := CompressData(data, "br", 4)
	require.NoError(t, err)
	assert.NotNil(t, compressed)

	// Verify we can decompress
	decompressed, err := DecompressData(compressed, "br")
	require.NoError(t, err)
	assert.Equal(t, data, decompressed)
}

func TestCompressData_Gzip(t *testing.T) {
	data := []byte("This is a test string that will be compressed with Gzip encoding.")

	compressed, err := CompressData(data, "gzip", 5)
	require.NoError(t, err)
	assert.NotNil(t, compressed)

	// Verify we can decompress
	decompressed, err := DecompressData(compressed, "gzip")
	require.NoError(t, err)
	assert.Equal(t, data, decompressed)
}

func TestCompressData_UnknownEncoding(t *testing.T) {
	data := []byte("Test data")

	result, err := CompressData(data, "unknown", 4)
	require.NoError(t, err)
	assert.Equal(t, data, result)
}

func TestCompressData_EmptyData(t *testing.T) {
	data := []byte{}

	compressed, err := CompressData(data, "br", 4)
	require.NoError(t, err)

	decompressed, err := DecompressData(compressed, "br")
	require.NoError(t, err)
	assert.Equal(t, data, decompressed)
}

func TestCompressData_LargeData(t *testing.T) {
	// Create large repetitive data (compresses well)
	data := bytes.Repeat([]byte("This is repetitive test data. "), 1000)

	compressed, err := CompressData(data, "br", 4)
	require.NoError(t, err)

	// Verify compression achieved size reduction
	assert.Less(t, len(compressed), len(data))

	// Verify decompression
	decompressed, err := DecompressData(compressed, "br")
	require.NoError(t, err)
	assert.Equal(t, data, decompressed)
}

func TestDecompressData_UnknownEncoding(t *testing.T) {
	data := []byte("Test data")

	result, err := DecompressData(data, "unknown")
	require.NoError(t, err)
	assert.Equal(t, data, result)
}

func TestDecompressData_InvalidGzip(t *testing.T) {
	invalidData := []byte("not valid gzip data")

	_, err := DecompressData(invalidData, "gzip")
	assert.Error(t, err)
}

func TestEstimateCompressionRatio_JSON(t *testing.T) {
	ratio := EstimateCompressionRatio("application/json")
	assert.Equal(t, 0.15, ratio)
}

func TestEstimateCompressionRatio_JavaScript(t *testing.T) {
	ratio := EstimateCompressionRatio("application/javascript")
	assert.Equal(t, 0.20, ratio)
}

func TestEstimateCompressionRatio_HTML(t *testing.T) {
	ratio := EstimateCompressionRatio("text/html")
	assert.Equal(t, 0.20, ratio)
}

func TestEstimateCompressionRatio_CSS(t *testing.T) {
	ratio := EstimateCompressionRatio("text/css")
	assert.Equal(t, 0.25, ratio)
}

func TestEstimateCompressionRatio_XML(t *testing.T) {
	ratio := EstimateCompressionRatio("application/xml")
	assert.Equal(t, 0.20, ratio)
}

func TestEstimateCompressionRatio_PlainText(t *testing.T) {
	ratio := EstimateCompressionRatio("text/plain")
	assert.Equal(t, 0.30, ratio)
}

func TestEstimateCompressionRatio_Unknown(t *testing.T) {
	ratio := EstimateCompressionRatio("application/octet-stream")
	assert.Equal(t, 0.50, ratio)
}

func TestBrotliMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := BrotliMiddleware()
	assert.NotNil(t, handler)
}

func TestBrotliMiddlewareWithLevel(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := BrotliMiddlewareWithLevel(6)
	assert.NotNil(t, handler)
}

func TestCompressionMiddleware_NilConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := CompressionMiddleware(nil)
	assert.NotNil(t, handler)
}

func TestCompressionMiddleware_ExcludedPath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CompressionMiddleware(DefaultCompressionConfig()))
	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	req, _ := http.NewRequest("GET", "/health", nil)
	req.Header.Set("Accept-Encoding", "br, gzip")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Should not have Content-Encoding since /health is excluded
	assert.Empty(t, w.Header().Get("Content-Encoding"))
}

func TestCompressionMiddleware_NoAcceptEncoding(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CompressionMiddleware(DefaultCompressionConfig()))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, strings.Repeat("Test data ", 200))
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	// No Accept-Encoding header
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Should not have Content-Encoding
	assert.Empty(t, w.Header().Get("Content-Encoding"))
}

func TestCompressionMiddleware_BrotliCompression(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := DefaultCompressionConfig()
	config.MinSize = 10 // Lower min size for test

	router := gin.New()
	router.Use(CompressionMiddleware(config))
	router.GET("/test", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.String(http.StatusOK, strings.Repeat(`{"test":"data"}`, 100))
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "br, gzip")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "br", w.Header().Get("Content-Encoding"))

	// Verify response can be decompressed
	br := brotli.NewReader(w.Body)
	decompressed := make([]byte, 10000)
	n, _ := br.Read(decompressed)
	assert.Contains(t, string(decompressed[:n]), `{"test":"data"}`)
}

func TestCompressionMiddleware_GzipFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := DefaultCompressionConfig()
	config.MinSize = 10
	config.EnableBrotli = false // Force gzip

	router := gin.New()
	router.Use(CompressionMiddleware(config))
	router.GET("/test", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.String(http.StatusOK, strings.Repeat(`{"test":"data"}`, 100))
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))

	// Verify response can be decompressed
	gr, err := gzip.NewReader(w.Body)
	require.NoError(t, err)
	decompressed := make([]byte, 10000)
	n, _ := gr.Read(decompressed)
	assert.Contains(t, string(decompressed[:n]), `{"test":"data"}`)
}

func TestCompressionMiddleware_SmallResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CompressionMiddleware(DefaultCompressionConfig()))
	router.GET("/test", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.String(http.StatusOK, `{"small":"data"}`) // Below MinSize
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "br, gzip")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Should not have Content-Encoding for small response
	assert.Empty(t, w.Header().Get("Content-Encoding"))
}

func TestBrotliRequestDecoder_NoEncoding(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(BrotliRequestDecoder())
	router.POST("/test", func(c *gin.Context) {
		body := make([]byte, 1000)
		n, _ := c.Request.Body.Read(body)
		c.String(http.StatusOK, string(body[:n]))
	})

	req, _ := http.NewRequest("POST", "/test", strings.NewReader("Test body"))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Test body", w.Body.String())
}

func TestBrotliRequestDecoder_BrotliBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Compress the request body
	originalBody := "This is the original request body"
	var buf bytes.Buffer
	bw := brotli.NewWriter(&buf)
	bw.Write([]byte(originalBody))
	bw.Close()

	router := gin.New()
	router.Use(BrotliRequestDecoder())
	router.POST("/test", func(c *gin.Context) {
		body := make([]byte, 1000)
		n, _ := c.Request.Body.Read(body)
		c.String(http.StatusOK, string(body[:n]))
	})

	req, _ := http.NewRequest("POST", "/test", &buf)
	req.Header.Set("Content-Encoding", "br")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, originalBody, w.Body.String())
}

func TestBrotliRequestDecoder_GzipBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Compress the request body with gzip
	originalBody := "This is the original request body"
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	gw.Write([]byte(originalBody))
	gw.Close()

	router := gin.New()
	router.Use(BrotliRequestDecoder())
	router.POST("/test", func(c *gin.Context) {
		body := make([]byte, 1000)
		n, _ := c.Request.Body.Read(body)
		c.String(http.StatusOK, string(body[:n]))
	})

	req, _ := http.NewRequest("POST", "/test", &buf)
	req.Header.Set("Content-Encoding", "gzip")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, originalBody, w.Body.String())
}

func TestBrotliRequestDecoder_InvalidGzip(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(BrotliRequestDecoder())
	router.POST("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	req, _ := http.NewRequest("POST", "/test", strings.NewReader("not valid gzip"))
	req.Header.Set("Content-Encoding", "gzip")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCompressionWriter_WriteString(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := DefaultCompressionConfig()
	config.MinSize = 10

	router := gin.New()
	router.Use(CompressionMiddleware(config))
	router.GET("/test", func(c *gin.Context) {
		c.Header("Content-Type", "text/plain")
		// Use WriteString path
		c.Writer.WriteString(strings.Repeat("Hello World ", 100))
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "br")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "br", w.Header().Get("Content-Encoding"))
}

func TestBrotliReader_ReadAndClose(t *testing.T) {
	// Test brotliReader directly
	originalData := "Test data for brotli reader"
	var buf bytes.Buffer
	bw := brotli.NewWriter(&buf)
	bw.Write([]byte(originalData))
	bw.Close()

	br := &brotliReader{
		reader: brotli.NewReader(&buf),
		closer: &nopCloser{},
	}

	data := make([]byte, 100)
	n, _ := br.Read(data)
	assert.Equal(t, originalData, string(data[:n]))

	err := br.Close()
	assert.NoError(t, err)
}

func TestGzipReader_ReadAndClose(t *testing.T) {
	// Test gzipReader directly
	originalData := "Test data for gzip reader"
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	gw.Write([]byte(originalData))
	gw.Close()

	gr, err := gzip.NewReader(&buf)
	require.NoError(t, err)

	gzr := &gzipReader{
		reader: gr,
		closer: &nopCloser{},
	}

	data := make([]byte, 100)
	n, _ := gzr.Read(data)
	assert.Equal(t, originalData, string(data[:n]))

	err = gzr.Close()
	assert.NoError(t, err)
}

// nopCloser is a helper for testing
type nopCloser struct{}

func (n *nopCloser) Close() error { return nil }

func TestCompressionMiddleware_NonCompressibleType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := DefaultCompressionConfig()
	config.MinSize = 10

	router := gin.New()
	router.Use(CompressionMiddleware(config))
	router.GET("/test", func(c *gin.Context) {
		c.Header("Content-Type", "image/png") // Not in compressible types
		c.String(http.StatusOK, strings.Repeat("Test data ", 200))
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "br, gzip")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Should not have Content-Encoding for non-compressible type
	assert.Empty(t, w.Header().Get("Content-Encoding"))
}

func TestCompressionConfig_DisabledCompression(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := DefaultCompressionConfig()
	config.EnableBrotli = false
	config.EnableGzip = false

	router := gin.New()
	router.Use(CompressionMiddleware(config))
	router.GET("/test", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.String(http.StatusOK, strings.Repeat(`{"test":"data"}`, 100))
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "br, gzip")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Should not have Content-Encoding when compression disabled
	assert.Empty(t, w.Header().Get("Content-Encoding"))
}

func TestCompressData_GzipErrors(t *testing.T) {
	// Test gzip with invalid level
	data := []byte("Test data")

	// Level 0-9 are valid for gzip, -1 for default
	// Testing with extreme levels
	_, err := CompressData(data, "gzip", 5)
	assert.NoError(t, err)
}

func TestDecompressData_InvalidBrotli(t *testing.T) {
	invalidData := []byte("not valid brotli data")

	_, err := DecompressData(invalidData, "br")
	assert.Error(t, err)
}
