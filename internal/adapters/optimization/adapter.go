// Package optimization provides adapters between HelixAgent's internal/optimization types
// and the extracted digital.vasic.optimization module.
package optimization

import (
	"context"
	"time"

	helixopt "dev.helix.agent/internal/optimization"
	modgptcache "digital.vasic.optimization/pkg/gptcache"
	modoutlines "digital.vasic.optimization/pkg/outlines"
	modstreaming "digital.vasic.optimization/pkg/streaming"
)

// CacheAdapter adapts the module's Cache interface to HelixAgent usage.
type CacheAdapter struct {
	cache modgptcache.Cache
}

// NewCacheAdapter creates a new cache adapter wrapping the module cache.
func NewCacheAdapter(cache modgptcache.Cache) *CacheAdapter {
	return &CacheAdapter{cache: cache}
}

// Get retrieves a cached response.
func (a *CacheAdapter) Get(ctx context.Context, query string) (*CacheHit, error) {
	resp, err := a.cache.Get(ctx, query)
	if err != nil {
		if err == modgptcache.ErrCacheMiss {
			return nil, nil
		}
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}

	return &CacheHit{
		Response:   resp.Response,
		Similarity: resp.Similarity,
		Metadata:   resp.Metadata,
		CachedAt:   resp.CachedAt,
	}, nil
}

// Set stores a response in the cache.
func (a *CacheAdapter) Set(ctx context.Context, query, response string) error {
	return a.cache.Set(ctx, query, response)
}

// Invalidate removes entries matching the query.
func (a *CacheAdapter) Invalidate(ctx context.Context, query string) error {
	return a.cache.Invalidate(ctx, query)
}

// CacheHit represents a cache hit result.
type CacheHit struct {
	Response   string
	Similarity float64
	Metadata   map[string]interface{}
	CachedAt   time.Time
}

// SchemaAdapter adapts the module's Schema to HelixAgent usage.
type SchemaAdapter struct{}

// NewSchemaAdapter creates a new schema adapter.
func NewSchemaAdapter() *SchemaAdapter {
	return &SchemaAdapter{}
}

// Schema represents a JSON schema for structured output.
type Schema = modoutlines.Schema

// NewSchemaBuilder creates a new schema builder.
func NewSchemaBuilder() *modoutlines.SchemaBuilder {
	return modoutlines.NewSchemaBuilder()
}

// StringSchema creates a string schema.
func StringSchema() *modoutlines.Schema {
	return modoutlines.StringSchema()
}

// IntegerSchema creates an integer schema.
func IntegerSchema() *modoutlines.Schema {
	return modoutlines.IntegerSchema()
}

// NumberSchema creates a number schema.
func NumberSchema() *modoutlines.Schema {
	return modoutlines.NumberSchema()
}

// BooleanSchema creates a boolean schema.
func BooleanSchema() *modoutlines.Schema {
	return modoutlines.BooleanSchema()
}

// ArraySchema creates an array schema.
func ArraySchema(items *modoutlines.Schema) *modoutlines.Schema {
	return modoutlines.ArraySchema(items)
}

// ObjectSchema creates an object schema.
func ObjectSchema(properties map[string]*modoutlines.Schema, required ...string) *modoutlines.Schema {
	return modoutlines.ObjectSchema(properties, required...)
}

// StreamBufferAdapter adapts the module's streaming buffer.
type StreamBufferAdapter struct {
	buffer modstreaming.Buffer
}

// NewStreamBufferAdapter creates a new stream buffer adapter.
func NewStreamBufferAdapter(strategy string, threshold int) *StreamBufferAdapter {
	var s modstreaming.FlushStrategy
	switch strategy {
	case "word":
		s = modstreaming.FlushOnWord
	case "sentence":
		s = modstreaming.FlushOnSentence
	case "line":
		s = modstreaming.FlushOnLine
	case "size":
		s = modstreaming.FlushOnSize
	default:
		s = modstreaming.FlushOnWord
	}
	return &StreamBufferAdapter{
		buffer: modstreaming.NewStreamBuffer(s, threshold),
	}
}

// Add adds text to the buffer and returns flushed content.
func (a *StreamBufferAdapter) Add(text string) []string {
	return a.buffer.Add(text)
}

// Flush returns any remaining content.
func (a *StreamBufferAdapter) Flush() string {
	return a.buffer.Flush()
}

// Reset clears the buffer.
func (a *StreamBufferAdapter) Reset() {
	a.buffer.Reset()
}

// ToHelixOptimizedRequest maps cache results to optimization results.
func ToHelixOptimizedRequest(cacheHit *CacheHit, prompt string) *helixopt.OptimizedRequest {
	if cacheHit == nil {
		return &helixopt.OptimizedRequest{
			OriginalPrompt:  prompt,
			OptimizedPrompt: prompt,
			CacheHit:        false,
		}
	}
	return &helixopt.OptimizedRequest{
		OriginalPrompt:  prompt,
		OptimizedPrompt: prompt,
		CacheHit:        true,
		CachedResponse:  cacheHit.Response,
	}
}
