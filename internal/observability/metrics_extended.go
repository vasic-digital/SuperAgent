package observability

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type MCPMetrics struct {
	meter metric.Meter
	mu    sync.RWMutex

	ToolCallsTotal    metric.Int64Counter
	ToolDuration      metric.Float64Histogram
	ToolErrorsTotal   metric.Int64Counter
	ToolsInFlight     metric.Int64UpDownCounter
	ToolResultSize    metric.Int64Histogram
	ServerConnections metric.Int64UpDownCounter
}

func NewMCPMetrics(serviceName string) (*MCPMetrics, error) {
	meter := otel.Meter(serviceName)
	m := &MCPMetrics{meter: meter}

	var err error

	m.ToolCallsTotal, err = meter.Int64Counter("mcp_tool_calls_total",
		metric.WithDescription("Total number of MCP tool calls"))
	if err != nil {
		return nil, err
	}

	m.ToolDuration, err = meter.Float64Histogram("mcp_tool_duration_seconds",
		metric.WithDescription("MCP tool call duration in seconds"),
		metric.WithUnit("s"))
	if err != nil {
		return nil, err
	}

	m.ToolErrorsTotal, err = meter.Int64Counter("mcp_tool_errors_total",
		metric.WithDescription("Total number of MCP tool errors"))
	if err != nil {
		return nil, err
	}

	m.ToolsInFlight, err = meter.Int64UpDownCounter("mcp_tools_in_flight",
		metric.WithDescription("Number of MCP tool calls currently in progress"))
	if err != nil {
		return nil, err
	}

	m.ToolResultSize, err = meter.Int64Histogram("mcp_tool_result_size_bytes",
		metric.WithDescription("Size of MCP tool results in bytes"),
		metric.WithUnit("By"))
	if err != nil {
		return nil, err
	}

	m.ServerConnections, err = meter.Int64UpDownCounter("mcp_server_connections",
		metric.WithDescription("Number of active MCP server connections"))
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (m *MCPMetrics) RecordToolCall(ctx context.Context, tool, provider string, duration float64, resultSize int64, err error) {
	attrs := metric.WithAttributes(
		attribute.String("tool", tool),
		attribute.String("provider", provider),
	)

	m.ToolCallsTotal.Add(ctx, 1, attrs)
	m.ToolDuration.Record(ctx, duration, attrs)

	if resultSize > 0 {
		m.ToolResultSize.Record(ctx, resultSize, attrs)
	}

	if err != nil {
		errorAttrs := metric.WithAttributes(
			attribute.String("tool", tool),
			attribute.String("provider", provider),
			attribute.String("error_type", errorType(err)),
		)
		m.ToolErrorsTotal.Add(ctx, 1, errorAttrs)
	}
}

func (m *MCPMetrics) StartToolCall(ctx context.Context, tool, provider string) func(ctx context.Context, resultSize int64, err error) {
	attrs := metric.WithAttributes(
		attribute.String("tool", tool),
		attribute.String("provider", provider),
	)
	m.ToolsInFlight.Add(ctx, 1, attrs)
	return func(ctx context.Context, resultSize int64, err error) {
		m.ToolsInFlight.Add(ctx, -1, attrs)
	}
}

type EndToolCallFunc func(ctx context.Context, resultSize int64, err error)

type EmbeddingMetrics struct {
	meter metric.Meter
	mu    sync.RWMutex

	RequestsTotal    metric.Int64Counter
	LatencySeconds   metric.Float64Histogram
	TokensTotal      metric.Int64Counter
	EmbeddingSize    metric.Int64Histogram
	RequestsInFlight metric.Int64UpDownCounter
	ErrorsTotal      metric.Int64Counter
	BatchSize        metric.Int64Histogram
}

func NewEmbeddingMetrics(serviceName string) (*EmbeddingMetrics, error) {
	meter := otel.Meter(serviceName)
	m := &EmbeddingMetrics{meter: meter}

	var err error

	m.RequestsTotal, err = meter.Int64Counter("embedding_requests_total",
		metric.WithDescription("Total number of embedding requests"))
	if err != nil {
		return nil, err
	}

	m.LatencySeconds, err = meter.Float64Histogram("embedding_latency_seconds",
		metric.WithDescription("Embedding request latency in seconds"),
		metric.WithUnit("s"))
	if err != nil {
		return nil, err
	}

	m.TokensTotal, err = meter.Int64Counter("embedding_tokens_total",
		metric.WithDescription("Total tokens processed for embeddings"))
	if err != nil {
		return nil, err
	}

	m.EmbeddingSize, err = meter.Int64Histogram("embedding_size",
		metric.WithDescription("Size of generated embeddings"))
	if err != nil {
		return nil, err
	}

	m.RequestsInFlight, err = meter.Int64UpDownCounter("embedding_requests_in_flight",
		metric.WithDescription("Number of embedding requests in progress"))
	if err != nil {
		return nil, err
	}

	m.ErrorsTotal, err = meter.Int64Counter("embedding_errors_total",
		metric.WithDescription("Total embedding request errors"))
	if err != nil {
		return nil, err
	}

	m.BatchSize, err = meter.Int64Histogram("embedding_batch_size",
		metric.WithDescription("Number of texts in embedding batches"))
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (m *EmbeddingMetrics) RecordEmbedding(ctx context.Context, provider string, tokens int, embeddingSize int, latencySeconds float64, batchSize int, err error) {
	attrs := metric.WithAttributes(
		attribute.String("provider", provider),
	)

	m.RequestsTotal.Add(ctx, 1, attrs)
	m.LatencySeconds.Record(ctx, latencySeconds, attrs)
	m.TokensTotal.Add(ctx, int64(tokens), attrs)
	m.EmbeddingSize.Record(ctx, int64(embeddingSize), attrs)

	if batchSize > 0 {
		m.BatchSize.Record(ctx, int64(batchSize), attrs)
	}

	if err != nil {
		errorAttrs := metric.WithAttributes(
			attribute.String("provider", provider),
			attribute.String("error_type", errorType(err)),
		)
		m.ErrorsTotal.Add(ctx, 1, errorAttrs)
	}
}

type VectorDBMetrics struct {
	meter metric.Meter
	mu    sync.RWMutex

	OperationsTotal    metric.Int64Counter
	LatencySeconds     metric.Float64Histogram
	VectorsTotal       metric.Int64Gauge
	CollectionSize     metric.Int64Gauge
	OperationsInFlight metric.Int64UpDownCounter
	ErrorsTotal        metric.Int64Counter
	QueryResultCount   metric.Int64Histogram
}

func NewVectorDBMetrics(serviceName string) (*VectorDBMetrics, error) {
	meter := otel.Meter(serviceName)
	m := &VectorDBMetrics{meter: meter}

	var err error

	m.OperationsTotal, err = meter.Int64Counter("vectordb_operations_total",
		metric.WithDescription("Total number of vector database operations"))
	if err != nil {
		return nil, err
	}

	m.LatencySeconds, err = meter.Float64Histogram("vectordb_latency_seconds",
		metric.WithDescription("Vector database operation latency in seconds"),
		metric.WithUnit("s"))
	if err != nil {
		return nil, err
	}

	m.VectorsTotal, err = meter.Int64Gauge("vectordb_vectors_total",
		metric.WithDescription("Total number of vectors stored"))
	if err != nil {
		return nil, err
	}

	m.CollectionSize, err = meter.Int64Gauge("vectordb_collection_size_bytes",
		metric.WithDescription("Size of vector collections in bytes"),
		metric.WithUnit("By"))
	if err != nil {
		return nil, err
	}

	m.OperationsInFlight, err = meter.Int64UpDownCounter("vectordb_operations_in_flight",
		metric.WithDescription("Number of vector database operations in progress"))
	if err != nil {
		return nil, err
	}

	m.ErrorsTotal, err = meter.Int64Counter("vectordb_errors_total",
		metric.WithDescription("Total vector database errors"))
	if err != nil {
		return nil, err
	}

	m.QueryResultCount, err = meter.Int64Histogram("vectordb_query_result_count",
		metric.WithDescription("Number of results returned by queries"))
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (m *VectorDBMetrics) RecordOperation(ctx context.Context, provider, operation string, latencySeconds float64, resultCount int, err error) {
	attrs := metric.WithAttributes(
		attribute.String("provider", provider),
		attribute.String("operation", operation),
	)

	m.OperationsTotal.Add(ctx, 1, attrs)
	m.LatencySeconds.Record(ctx, latencySeconds, attrs)

	if resultCount > 0 {
		m.QueryResultCount.Record(ctx, int64(resultCount), attrs)
	}

	if err != nil {
		errorAttrs := metric.WithAttributes(
			attribute.String("provider", provider),
			attribute.String("operation", operation),
			attribute.String("error_type", errorType(err)),
		)
		m.ErrorsTotal.Add(ctx, 1, errorAttrs)
	}
}

func (m *VectorDBMetrics) RecordCollectionStats(ctx context.Context, provider, collection string, vectors int64, sizeBytes int64) {
	attrs := metric.WithAttributes(
		attribute.String("provider", provider),
		attribute.String("collection", collection),
	)

	m.VectorsTotal.Record(ctx, vectors, attrs)
	m.CollectionSize.Record(ctx, sizeBytes, attrs)
}

type MemoryMetrics struct {
	meter metric.Meter
	mu    sync.RWMutex

	OperationsTotal    metric.Int64Counter
	SearchLatency      metric.Float64Histogram
	EntriesTotal       metric.Int64Gauge
	OperationsInFlight metric.Int64UpDownCounter
	ErrorsTotal        metric.Int64Counter
	CacheHits          metric.Int64Counter
	CacheMisses        metric.Int64Counter
}

func NewMemoryMetrics(serviceName string) (*MemoryMetrics, error) {
	meter := otel.Meter(serviceName)
	m := &MemoryMetrics{meter: meter}

	var err error

	m.OperationsTotal, err = meter.Int64Counter("memory_operations_total",
		metric.WithDescription("Total number of memory operations"))
	if err != nil {
		return nil, err
	}

	m.SearchLatency, err = meter.Float64Histogram("memory_search_latency_seconds",
		metric.WithDescription("Memory search latency in seconds"),
		metric.WithUnit("s"))
	if err != nil {
		return nil, err
	}

	m.EntriesTotal, err = meter.Int64Gauge("memory_entries_total",
		metric.WithDescription("Total number of memory entries"))
	if err != nil {
		return nil, err
	}

	m.OperationsInFlight, err = meter.Int64UpDownCounter("memory_operations_in_flight",
		metric.WithDescription("Number of memory operations in progress"))
	if err != nil {
		return nil, err
	}

	m.ErrorsTotal, err = meter.Int64Counter("memory_errors_total",
		metric.WithDescription("Total memory operation errors"))
	if err != nil {
		return nil, err
	}

	m.CacheHits, err = meter.Int64Counter("memory_cache_hits_total",
		metric.WithDescription("Total memory cache hits"))
	if err != nil {
		return nil, err
	}

	m.CacheMisses, err = meter.Int64Counter("memory_cache_misses_total",
		metric.WithDescription("Total memory cache misses"))
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (m *MemoryMetrics) RecordOperation(ctx context.Context, operation, memoryType string, latencySeconds float64, err error) {
	attrs := metric.WithAttributes(
		attribute.String("operation", operation),
		attribute.String("type", memoryType),
	)

	m.OperationsTotal.Add(ctx, 1, attrs)
	m.SearchLatency.Record(ctx, latencySeconds, attrs)

	if err != nil {
		errorAttrs := metric.WithAttributes(
			attribute.String("operation", operation),
			attribute.String("type", memoryType),
			attribute.String("error_type", errorType(err)),
		)
		m.ErrorsTotal.Add(ctx, 1, errorAttrs)
	}
}

func (m *MemoryMetrics) RecordEntries(ctx context.Context, memoryType string, count int64) {
	attrs := metric.WithAttributes(
		attribute.String("type", memoryType),
	)
	m.EntriesTotal.Record(ctx, count, attrs)
}

func (m *MemoryMetrics) RecordCacheHit(ctx context.Context) {
	m.CacheHits.Add(ctx, 1)
}

func (m *MemoryMetrics) RecordCacheMiss(ctx context.Context) {
	m.CacheMisses.Add(ctx, 1)
}

type StreamingMetrics struct {
	meter metric.Meter
	mu    sync.RWMutex

	ChunksTotal     metric.Int64Counter
	ErrorsTotal     metric.Int64Counter
	Duration        metric.Float64Histogram
	StreamsInFlight metric.Int64UpDownCounter
	ChunkSize       metric.Int64Histogram
	ThroughputBytes metric.Int64Counter
}

func NewStreamingMetrics(serviceName string) (*StreamingMetrics, error) {
	meter := otel.Meter(serviceName)
	m := &StreamingMetrics{meter: meter}

	var err error

	m.ChunksTotal, err = meter.Int64Counter("stream_chunks_total",
		metric.WithDescription("Total number of stream chunks sent"))
	if err != nil {
		return nil, err
	}

	m.ErrorsTotal, err = meter.Int64Counter("stream_errors_total",
		metric.WithDescription("Total number of streaming errors"))
	if err != nil {
		return nil, err
	}

	m.Duration, err = meter.Float64Histogram("stream_duration_seconds",
		metric.WithDescription("Stream duration in seconds"),
		metric.WithUnit("s"))
	if err != nil {
		return nil, err
	}

	m.StreamsInFlight, err = meter.Int64UpDownCounter("stream_streams_in_flight",
		metric.WithDescription("Number of active streams"))
	if err != nil {
		return nil, err
	}

	m.ChunkSize, err = meter.Int64Histogram("stream_chunk_size_bytes",
		metric.WithDescription("Size of stream chunks in bytes"),
		metric.WithUnit("By"))
	if err != nil {
		return nil, err
	}

	m.ThroughputBytes, err = meter.Int64Counter("stream_throughput_bytes_total",
		metric.WithDescription("Total bytes streamed"),
		metric.WithUnit("By"))
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (m *StreamingMetrics) RecordChunk(ctx context.Context, provider string, chunkSize int64) {
	attrs := metric.WithAttributes(
		attribute.String("provider", provider),
	)

	m.ChunksTotal.Add(ctx, 1, attrs)
	m.ChunkSize.Record(ctx, chunkSize, attrs)
	m.ThroughputBytes.Add(ctx, chunkSize, attrs)
}

func (m *StreamingMetrics) RecordStream(ctx context.Context, provider string, durationSeconds float64, err error) {
	attrs := metric.WithAttributes(
		attribute.String("provider", provider),
	)

	m.Duration.Record(ctx, durationSeconds, attrs)

	if err != nil {
		errorAttrs := metric.WithAttributes(
			attribute.String("provider", provider),
			attribute.String("error_type", errorType(err)),
		)
		m.ErrorsTotal.Add(ctx, 1, errorAttrs)
	}
}

func (m *StreamingMetrics) StartStream(ctx context.Context, provider string) {
	attrs := metric.WithAttributes(
		attribute.String("provider", provider),
	)
	m.StreamsInFlight.Add(ctx, 1, attrs)
}

func (m *StreamingMetrics) EndStream(ctx context.Context, provider string) {
	attrs := metric.WithAttributes(
		attribute.String("provider", provider),
	)
	m.StreamsInFlight.Add(ctx, -1, attrs)
}

type ProtocolMetrics struct {
	meter metric.Meter
	mu    sync.RWMutex

	RequestsTotal    metric.Int64Counter
	ErrorsTotal      metric.Int64Counter
	LatencySeconds   metric.Float64Histogram
	RequestsInFlight metric.Int64UpDownCounter
}

func NewProtocolMetrics(serviceName string) (*ProtocolMetrics, error) {
	meter := otel.Meter(serviceName)
	m := &ProtocolMetrics{meter: meter}

	var err error

	m.RequestsTotal, err = meter.Int64Counter("protocol_requests_total",
		metric.WithDescription("Total number of protocol requests"))
	if err != nil {
		return nil, err
	}

	m.ErrorsTotal, err = meter.Int64Counter("protocol_errors_total",
		metric.WithDescription("Total number of protocol errors"))
	if err != nil {
		return nil, err
	}

	m.LatencySeconds, err = meter.Float64Histogram("protocol_latency_seconds",
		metric.WithDescription("Protocol request latency in seconds"),
		metric.WithUnit("s"))
	if err != nil {
		return nil, err
	}

	m.RequestsInFlight, err = meter.Int64UpDownCounter("protocol_requests_in_flight",
		metric.WithDescription("Number of protocol requests in progress"))
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (m *ProtocolMetrics) RecordRequest(ctx context.Context, protocol, operation string, latencySeconds float64, err error) {
	attrs := metric.WithAttributes(
		attribute.String("protocol", protocol),
		attribute.String("operation", operation),
	)

	m.RequestsTotal.Add(ctx, 1, attrs)
	m.LatencySeconds.Record(ctx, latencySeconds, attrs)

	if err != nil {
		errorAttrs := metric.WithAttributes(
			attribute.String("protocol", protocol),
			attribute.String("operation", operation),
			attribute.String("error_type", errorType(err)),
		)
		m.ErrorsTotal.Add(ctx, 1, errorAttrs)
	}
}

var (
	globalMCPMetrics       *MCPMetrics
	globalEmbeddingMetrics *EmbeddingMetrics
	globalVectorDBMetrics  *VectorDBMetrics
	globalMemoryMetrics    *MemoryMetrics
	globalStreamingMetrics *StreamingMetrics
	globalProtocolMetrics  *ProtocolMetrics
	extendedMetricsOnce    sync.Once
)

func InitExtendedMetrics(serviceName string) error {
	var initErr error
	extendedMetricsOnce.Do(func() {
		globalMCPMetrics, initErr = NewMCPMetrics(serviceName)
		if initErr != nil {
			return
		}
		globalEmbeddingMetrics, initErr = NewEmbeddingMetrics(serviceName)
		if initErr != nil {
			return
		}
		globalVectorDBMetrics, initErr = NewVectorDBMetrics(serviceName)
		if initErr != nil {
			return
		}
		globalMemoryMetrics, initErr = NewMemoryMetrics(serviceName)
		if initErr != nil {
			return
		}
		globalStreamingMetrics, initErr = NewStreamingMetrics(serviceName)
		if initErr != nil {
			return
		}
		globalProtocolMetrics, initErr = NewProtocolMetrics(serviceName)
	})
	return initErr
}

func GetMCPMetrics() *MCPMetrics {
	if globalMCPMetrics == nil {
		globalMCPMetrics, _ = NewMCPMetrics("helixagent")
	}
	return globalMCPMetrics
}

func GetEmbeddingMetrics() *EmbeddingMetrics {
	if globalEmbeddingMetrics == nil {
		globalEmbeddingMetrics, _ = NewEmbeddingMetrics("helixagent")
	}
	return globalEmbeddingMetrics
}

func GetVectorDBMetrics() *VectorDBMetrics {
	if globalVectorDBMetrics == nil {
		globalVectorDBMetrics, _ = NewVectorDBMetrics("helixagent")
	}
	return globalVectorDBMetrics
}

func GetMemoryMetrics() *MemoryMetrics {
	if globalMemoryMetrics == nil {
		globalMemoryMetrics, _ = NewMemoryMetrics("helixagent")
	}
	return globalMemoryMetrics
}

func GetStreamingMetrics() *StreamingMetrics {
	if globalStreamingMetrics == nil {
		globalStreamingMetrics, _ = NewStreamingMetrics("helixagent")
	}
	return globalStreamingMetrics
}

func GetProtocolMetrics() *ProtocolMetrics {
	if globalProtocolMetrics == nil {
		globalProtocolMetrics, _ = NewProtocolMetrics("helixagent")
	}
	return globalProtocolMetrics
}

func errorType(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
