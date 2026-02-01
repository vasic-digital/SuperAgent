package middleware

import (
	"bufio"
	"compress/gzip"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/andybalholm/brotli"
	"github.com/gin-gonic/gin"
)

// CompressionConfig holds configuration for compression middleware
type CompressionConfig struct {
	// Brotli compression level (0-11, default: 4)
	BrotliLevel int
	// Gzip compression level (1-9, default: 5)
	GzipLevel int
	// Minimum size to compress (default: 1024 bytes)
	MinSize int
	// Content types to compress
	CompressibleTypes []string
	// Skip compression for specific paths
	ExcludePaths []string
	// Enable Brotli compression
	EnableBrotli bool
	// Enable Gzip compression (fallback)
	EnableGzip bool
}

// DefaultCompressionConfig returns default compression configuration
func DefaultCompressionConfig() *CompressionConfig {
	return &CompressionConfig{
		BrotliLevel: 4,
		GzipLevel:   5,
		MinSize:     1024,
		CompressibleTypes: []string{
			"application/json",
			"application/javascript",
			"application/xml",
			"text/plain",
			"text/html",
			"text/css",
			"text/xml",
			"text/javascript",
			"application/x-javascript",
			"image/svg+xml",
		},
		ExcludePaths: []string{
			"/health",
			"/metrics",
		},
		EnableBrotli: true,
		EnableGzip:   true,
	}
}

// Pools for writers
var (
	brotliWriterPool = sync.Pool{
		New: func() interface{} {
			return brotli.NewWriter(nil)
		},
	}
	gzipWriterPool = sync.Pool{
		New: func() interface{} {
			w, _ := gzip.NewWriterLevel(nil, gzip.DefaultCompression)
			return w
		},
	}
)

// compressionWriter wraps gin.ResponseWriter with compression support
type compressionWriter struct {
	gin.ResponseWriter
	compressor io.WriteCloser
	encoding   string
	config     *CompressionConfig
	written    bool
}

// Write implements io.Writer
func (w *compressionWriter) Write(data []byte) (int, error) {
	if !w.written {
		// Check if we should compress
		if w.shouldCompress(data) {
			w.setupCompressor()
		}
		w.written = true
	}

	if w.compressor != nil {
		return w.compressor.Write(data)
	}
	return w.ResponseWriter.Write(data)
}

// WriteString implements io.StringWriter
func (w *compressionWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

// shouldCompress determines if the response should be compressed
func (w *compressionWriter) shouldCompress(data []byte) bool {
	// Check minimum size
	if len(data) < w.config.MinSize {
		return false
	}

	// Check content type
	contentType := w.Header().Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}

	for _, ct := range w.config.CompressibleTypes {
		if strings.HasPrefix(contentType, ct) {
			return true
		}
	}

	return false
}

// setupCompressor initializes the compressor based on encoding
func (w *compressionWriter) setupCompressor() {
	switch w.encoding {
	case "br":
		bw := brotliWriterPool.Get().(*brotli.Writer)
		bw.Reset(w.ResponseWriter)
		w.compressor = bw
		w.Header().Set("Content-Encoding", "br")
	case "gzip":
		gw := gzipWriterPool.Get().(*gzip.Writer)
		gw.Reset(w.ResponseWriter)
		w.compressor = gw
		w.Header().Set("Content-Encoding", "gzip")
	}

	// Remove Content-Length as it will change
	w.Header().Del("Content-Length")
	// Add Vary header
	w.Header().Add("Vary", "Accept-Encoding")
}

// Close closes the compression writer
func (w *compressionWriter) Close() error {
	if w.compressor != nil {
		err := w.compressor.Close()

		// Return writer to pool
		switch w.encoding {
		case "br":
			brotliWriterPool.Put(w.compressor)
		case "gzip":
			gzipWriterPool.Put(w.compressor)
		}

		return err
	}
	return nil
}

// Hijack implements http.Hijacker
func (w *compressionWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

// Flush implements http.Flusher
func (w *compressionWriter) Flush() {
	if flusher, ok := w.compressor.(interface{ Flush() error }); ok {
		flusher.Flush()
	}
	w.ResponseWriter.(http.Flusher).Flush()
}

// BrotliMiddleware provides Brotli compression for responses
func BrotliMiddleware() gin.HandlerFunc {
	return CompressionMiddleware(DefaultCompressionConfig())
}

// BrotliMiddlewareWithLevel creates Brotli middleware with custom level
func BrotliMiddlewareWithLevel(level int) gin.HandlerFunc {
	config := DefaultCompressionConfig()
	config.BrotliLevel = level
	return CompressionMiddleware(config)
}

// CompressionMiddleware provides configurable compression middleware
func CompressionMiddleware(config *CompressionConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultCompressionConfig()
	}

	return func(c *gin.Context) {
		// Skip for excluded paths
		for _, path := range config.ExcludePaths {
			if strings.HasPrefix(c.Request.URL.Path, path) {
				c.Next()
				return
			}
		}

		// Check Accept-Encoding header
		acceptEncoding := c.GetHeader("Accept-Encoding")
		var encoding string

		// Prefer Brotli over Gzip
		if config.EnableBrotli && strings.Contains(acceptEncoding, "br") {
			encoding = "br"
		} else if config.EnableGzip && strings.Contains(acceptEncoding, "gzip") {
			encoding = "gzip"
		}

		if encoding == "" {
			c.Next()
			return
		}

		// Create compression writer
		cw := &compressionWriter{
			ResponseWriter: c.Writer,
			encoding:       encoding,
			config:         config,
		}
		c.Writer = cw

		defer func() {
			cw.Close()
		}()

		c.Next()
	}
}

// BrotliRequestDecoder provides middleware to decode Brotli-compressed request bodies
func BrotliRequestDecoder() gin.HandlerFunc {
	return func(c *gin.Context) {
		contentEncoding := c.GetHeader("Content-Encoding")

		switch contentEncoding {
		case "br":
			c.Request.Body = &brotliReader{
				reader: brotli.NewReader(c.Request.Body),
				closer: c.Request.Body,
			}
		case "gzip":
			gr, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error": "failed to decode gzip request body",
				})
				return
			}
			c.Request.Body = &gzipReader{
				reader: gr,
				closer: c.Request.Body,
			}
		}

		c.Next()
	}
}

// brotliReader wraps a brotli reader with Close support
type brotliReader struct {
	reader io.Reader
	closer io.Closer
}

func (r *brotliReader) Read(p []byte) (int, error) {
	return r.reader.Read(p)
}

func (r *brotliReader) Close() error {
	return r.closer.Close()
}

// gzipReader wraps a gzip reader with original body closer
type gzipReader struct {
	reader *gzip.Reader
	closer io.Closer
}

func (r *gzipReader) Read(p []byte) (int, error) {
	return r.reader.Read(p)
}

func (r *gzipReader) Close() error {
	r.reader.Close()
	return r.closer.Close()
}

// CompressData compresses data using the specified encoding
func CompressData(data []byte, encoding string, level int) ([]byte, error) {
	var buf strings.Builder

	switch encoding {
	case "br":
		w := brotli.NewWriterLevel(&buf, level)
		if _, err := w.Write(data); err != nil {
			return nil, err
		}
		if err := w.Close(); err != nil {
			return nil, err
		}
	case "gzip":
		w, err := gzip.NewWriterLevel(&buf, level)
		if err != nil {
			return nil, err
		}
		if _, err := w.Write(data); err != nil {
			return nil, err
		}
		if err := w.Close(); err != nil {
			return nil, err
		}
	default:
		return data, nil
	}

	return []byte(buf.String()), nil
}

// DecompressData decompresses data using the specified encoding
func DecompressData(data []byte, encoding string) ([]byte, error) {
	reader := strings.NewReader(string(data))

	switch encoding {
	case "br":
		br := brotli.NewReader(reader)
		return io.ReadAll(br)
	case "gzip":
		gr, err := gzip.NewReader(reader)
		if err != nil {
			return nil, err
		}
		defer func() { _ = gr.Close() }()
		return io.ReadAll(gr)
	default:
		return data, nil
	}
}

// EstimateCompressionRatio estimates the compression ratio for a given data type
// Returns approximate ratio (compressed/original)
func EstimateCompressionRatio(contentType string) float64 {
	switch {
	case strings.Contains(contentType, "json"):
		return 0.15 // JSON compresses very well (~85% reduction)
	case strings.Contains(contentType, "javascript"):
		return 0.20 // JS compresses well (~80% reduction)
	case strings.Contains(contentType, "html"):
		return 0.20 // HTML compresses well (~80% reduction)
	case strings.Contains(contentType, "css"):
		return 0.25 // CSS compresses well (~75% reduction)
	case strings.Contains(contentType, "xml"):
		return 0.20 // XML compresses well (~80% reduction)
	case strings.Contains(contentType, "text"):
		return 0.30 // Plain text (~70% reduction)
	default:
		return 0.50 // Unknown type, assume moderate compression
	}
}
